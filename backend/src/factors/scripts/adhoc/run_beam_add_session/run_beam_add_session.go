package main

import (
	"bytes"
	"context"
	"factors/model/model"
	"factors/model/store"
	"factors/task/session"
	"flag"
	"fmt"
	"math"
	"net"
	"net/http"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"

	C "factors/config"
	U "factors/util"

	"github.com/apache/beam/sdks/go/pkg/beam"
	"github.com/apache/beam/sdks/go/pkg/beam/x/beamx"

	log "github.com/sirupsen/logrus"
)

/*
Sample query to run on:
Development (Direct Runner):
go run scripts/run_beam_add_session/run_beam_add_session.go --project_ids='*'

Staging (Dataflow runner):

	go run scripts/run_beam_add_session/run_beam_add_session.go --project_ids='2,3' --runner dataflow --project factors-staging \
		--region us-west1 --temp_location gs://factors-staging-misc/beam/tmp/ \
		--staging_location gs://factors-staging-misc/beam/binaries/ \
		--worker_harness_container_image=apache/beam_go_sdk:latest --db_host=10.12.64.2 \
		--db_user autometa_ro --db_pass='' --redis_host='10.0.0.24' \
		--redis_port=8379 --subnetwork='regions/us-west1/subnetworks/us-west-1-factors-staging-subnet-1' \
		--num_workers=1 --max_num_workers=1 --zone='us-west1-b'
*/
const stepTrace = "StepTrace"

var (
	env = flag.String("env", "development", "")

	memSQLHost        = flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	memSQLPort        = flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser        = flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName        = flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass        = flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	memSQLCertificate = flag.String("memsql_cert", "", "")
	primaryDatastore  = flag.String("primary_datastore", C.DatastoreTypeMemSQL, "Primary datastore type as memsql or postgres")

	redisHost           = flag.String("redis_host", "localhost", "")
	redisPort           = flag.Int("redis_port", 6379, "")
	redisHostPersistent = flag.String("redis_host_ps", "localhost", "")
	redisPortPersistent = flag.Int("redis_port_ps", 6379, "")

	// Allowed list of projects to add session. Defaults to all (*), if not given.
	projectIDs                          = flag.String("project_ids", "*", "Allowed projects to create sessions offline.")
	disabledProjectIDs                  = flag.String("disabled_project_ids", "", "Disallowed projects to create sessions offline.")
	bufferTimeBeforeCreateSessionInMins = flag.Int64("buffer_time_in_mins", 30, "Buffer time to wait before processing an event for session.")

	// Limits the start_timestamp to max lookback, if exceeds.
	maxLookbackHours = flag.Int64("max_lookback_hours", 0, "Max lookback hours to look for session existence.")

	// Add session for a specific window of events.
	startTimestamp = flag.Int64("start_timestamp", 0, "Add session to specific window of events - start timestamp.")
	endTimestamp   = flag.Int64("end_timestamp", 0, "Add session to specific window of events - end timestamp.")
	logInfo        = flag.Int("log_info", 0, "Flag to enable and disable logging.")

	sentryDSN = flag.String("sentry_dsn", "", "Sentry DSN")

	gcpProjectID                      = flag.String("gcp_project_id", "", "Project ID on Google Cloud")
	gcpProjectLocation                = flag.String("gcp_project_location", "", "Location of google cloud project cluster")
	cacheSortedSet                    = flag.Bool("cache_with_sorted_set", false, "Cache with sorted set keys")
	allowChannelGroupingForProjectIDs = flag.String("allow_channel_grouping_for_projects",
		"", "List of projects to allow channel property population in session events.")

	enableOLTPQueriesMemSQLImprovements = flag.String("enable_OLTP_queries_memsql_improvements", "", "")
	captureSourceInUsersTable           = flag.String("capture_source_in_users_table", "", "")
)

func registerStructs() {

	beam.RegisterType(reflect.TypeOf((*StatusBeam)(nil)).Elem())
	beam.RegisterType(reflect.TypeOf((*EventsByProjectResponse)(nil)).Elem())
	beam.RegisterType(reflect.TypeOf((*model.EventName)(nil)).Elem())
	beam.RegisterType(reflect.TypeOf((*model.Event)(nil)).Elem())

	beam.RegisterType(reflect.TypeOf((*getAddSessionAllowedProjectsFn)(nil)).Elem())
	beam.RegisterType(reflect.TypeOf((*pullEventsByProjectIdFn)(nil)).Elem())

	beam.RegisterType(reflect.TypeOf((*UserIDSessionCreationResponse)(nil)).Elem())
	beam.RegisterType(reflect.TypeOf((*addSessionsByUserIDProjectIDFn)(nil)).Elem())

	beam.RegisterType(reflect.TypeOf((*ProjectStatusResponse)(nil)).Elem())
	beam.RegisterType(reflect.TypeOf((*reportProjectLevelSummary)(nil)).Elem())

	beam.RegisterType(reflect.TypeOf((*AddSessionJobProps)(nil)).Elem())
}

type StatusBeam struct {
	Status                       string `json:"status"`
	EventsDownloadIntervalInMins int64  `json:"events_download_interval_in_mins"`
	NoOfEvents                   int    `json:"no_of_events_downloaded"`
	// count of after filter events. actual no.of events processed for session.
	NoOfEventsProcessed       int `json:"no_of_events_processed"`
	NoOfUsers                 int `json:"no_of_users"`
	NoOfSessionsContinued     int `json:"no_of_sessions_continued"`
	NoOfSessionsCreated       int `json:"no_of_sessions_created"`
	NoOfUserPropertiesUpdates int `json:"no_of_user_properties_updates"`
	// Only projectLevel summary
	SeenFailure bool `json:"-"`
	// Used only for project's user count
	NoOfUserSessionFailures int `json:"-"`
	NoOfUserSessionSuccess  int `json:"-"`
	// Used on user event worker.
	ErrorDetail string `json:"error_detail"`
}

func (s *StatusBeam) Set(noOfProcessedEvents, noOfCreated int, isContinuedFirst bool,
	noOfUserPropUpdates int, seenFailure bool, status *StatusBeam) {

	status.SeenFailure = seenFailure

	if isContinuedFirst {
		status.NoOfSessionsContinued++
	}
	status.NoOfSessionsCreated += noOfCreated
	status.NoOfEventsProcessed += noOfProcessedEvents
	status.NoOfUserPropertiesUpdates += noOfUserPropUpdates
}

func getGoroutineID() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, _ := strconv.ParseUint(string(b), 10, 64)
	return n
}

func getMacAddr() []string {
	ifas, err := net.Interfaces()
	if err != nil {
		return nil
	}
	var as []string
	for _, ifa := range ifas {
		a := ifa.HardwareAddr.String()
		if a != "" {
			as = append(as, a)
		}
	}
	return as
}

func getLogContext() *log.Entry {
	return log.WithFields(log.Fields{
		"Name":    os.Args[0],
		"PID":     os.Getpid(),
		"PPID":    os.Getppid(),
		"Routine": getGoroutineID(),
		"MACAddr": getMacAddr(),
	})
}

func initConf(config *C.Configuration) {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetReportCaller(true)

	C.InitConf(config)
	C.SetIsBeamPipeline()
	err := C.InitDBWithMaxIdleAndMaxOpenConn(*config, 50, 50)
	if err != nil {
		// TODO(prateek): Check how a panic here will effect the pipeline.
		log.WithError(err).Panic("Failed to initalize db.")
	}
	// Cache dependency for requests not using queue.
	C.InitRedis(config.RedisHost, config.RedisPort)
	C.InitRedisPersistent(config.RedisHostPersistent, config.RedisPortPersistent)
	C.InitSentryLogging(config.SentryDSN, config.AppName)
	C.InitMetricsExporter(config.Env, config.AppName, config.GCPProjectID, config.GCPProjectLocation)

	C.KillDBQueriesOnExit()
	defer C.WaitAndFlushAllCollectors(65 * time.Second)
}

type getAddSessionAllowedProjectsFn struct {
	Config *C.Configuration
}

func (f *getAddSessionAllowedProjectsFn) StartBundle(ctx context.Context, emit func(uint64)) {
	log.Info("Initializing conf from StartBundle getAddSessionAllowedProjectsFn")
	initConf(f.Config)
}

func (f *getAddSessionAllowedProjectsFn) FinishBundle(ctx context.Context, emit func(uint64)) {
	log.Info("Closing DB Connection from FinishBundle getAddSessionAllowedProjectsFn")
	C.GetServices().Db.Close()
	C.SafeFlushAllCollectors()
}

func (f *getAddSessionAllowedProjectsFn) ProcessElement(ctx context.Context, projectsToRunString string,
	emit func(int64)) {

	logCtx := log.WithField("log_type", stepTrace).WithFields(log.Fields{
		"Method": "getAddSessionAllowedProjectsFn",
	})

	projectIDSplit := strings.Split(projectsToRunString, "|")
	stringProjectIDs := strings.TrimSpace(projectIDSplit[0])
	excludeProjectIDs := strings.TrimSpace(projectIDSplit[1])

	var allowedProjectIds []int64
	allowedProjectIds, errCode := session.GetAddSessionAllowedProjects(stringProjectIDs, excludeProjectIDs)
	if errCode != http.StatusFound {
		logCtx.WithField("err_code", errCode).Error("Failed to get add session allowed project ids.")
		return
	}
	for _, projectID := range allowedProjectIds {
		logCtx.WithField("project_id", projectID).Info("Emitting projectID from getAddSessionAllowedProjectsFn.")
		emit(projectID)
	}
}

// EventsByProjectResponse Struct to store events pull by projectId response and emit.
type EventsByProjectResponse struct {
	ProjectID        int64
	UserID           string
	TimeTaken        int64
	Events           []model.Event
	SessionEventName *model.EventName
	Status           *StatusBeam
	ErrorCode        int
}

type pullEventsByProjectIdFn struct {
	Config   *C.Configuration
	JobProps *AddSessionJobProps
}

func (f *pullEventsByProjectIdFn) StartBundle(ctx context.Context, emit func(EventsByProjectResponse)) {
	log.Info("Initializing conf from StartBundle pullEventsByProjectIdFn")
	initConf(f.Config)
}

func (f *pullEventsByProjectIdFn) FinishBundle(ctx context.Context, emit func(EventsByProjectResponse)) {
	log.Info("Closing DB Connection from FinishBundle pullEventsByProjectIdFn")
	C.GetServices().Db.Close()
	C.SafeFlushAllCollectors()
}

func (f *pullEventsByProjectIdFn) ProcessElement(ctx context.Context,
	projectID int64, emit func(EventsByProjectResponse)) {

	logCtx := log.WithField("log_type", stepTrace).WithFields(log.Fields{
		"method":     "pullEventsByProjectIdFn",
		"project_id": projectID,
	})
	startTime := time.Now().UnixNano() / 1000
	getLogContext().WithField("log_type", stepTrace).WithFields(log.Fields{
		"project_id": projectID,
		"time_micro": startTime,
		"time_type":  "Start",
	}).Info("ProcessElement log start pullEventsByProjectIdFn")

	logCtx.WithField("log_type", stepTrace).
		WithField("project_id", projectID).
		WithField("max_lookback", f.JobProps.MaxLookbackTimestampInSec).
		WithField("start_timestamp", f.JobProps.StartTimestamp).
		WithField("end_timestamp", f.JobProps.EndTimestamp).
		WithField("buffer_time_in_mins", f.JobProps.BufferTimeBeforeSessionCreateInSecs).
		Info("Pulling events for the project.")

	beamStatus := &StatusBeam{
		SeenFailure:               false,
		NoOfEvents:                0,
		NoOfEventsProcessed:       0,
		NoOfUsers:                 0,
		NoOfSessionsContinued:     0,
		NoOfSessionsCreated:       0,
		NoOfUserPropertiesUpdates: 0,
		ErrorDetail:               "",
	}

	shouldReturn, sessionEventName, errCode, userEventsMap,
		noOfEventsDownloaded, errorDetail := session.GetNextSessionAndPullEvent(projectID, f.JobProps.MaxLookbackTimestampInSec, f.JobProps.StartTimestamp,
		f.JobProps.EndTimestamp, f.JobProps.BufferTimeBeforeSessionCreateInSecs, logCtx)

	logCtx.Info(fmt.Sprintf("For ProjectID: %v, sessionEventName: %v and noOfEventsDownloaded: %v", projectID, sessionEventName, noOfEventsDownloaded))

	// return if getNextSessionAndPullEvent failed or NOT_MODIFIED
	if shouldReturn {
		logCtx.WithField("log_type", stepTrace).WithFields(log.Fields{"project_id": projectID, "err_code": errCode, "error_detail": errorDetail}).Info("No events found for the project. Exiting.")
		return
	}
	usersCount := 0
	if noOfEventsDownloaded > 0 {
		beamStatus.NoOfEvents = noOfEventsDownloaded
		beamStatus.NoOfUsers = len(*userEventsMap)
		for userId, events := range *userEventsMap {
			usersCount++
			eventsResponse := EventsByProjectResponse{
				ProjectID:        projectID,
				UserID:           userId,
				Events:           events,
				SessionEventName: sessionEventName,
				Status:           beamStatus,
				ErrorCode:        http.StatusFound,
			}
			emit(eventsResponse)
		}
	}
	getLogContext().WithField("log_type", stepTrace).WithFields(log.Fields{
		"project_id":   projectID,
		"time_taken":   time.Now().UnixNano()/1000 - startTime,
		"time_micro":   time.Now().UnixNano() / 1000,
		"time_type":    "End",
		"no_of_users":  usersCount,
		"no_of_events": noOfEventsDownloaded,
	}).WithField("error_code", errCode).Info("ProcessElement log end pullEventsByProjectIdFn")
}

// UserIDSessionCreationResponse Struct to store userID session creation response and emit.
type UserIDSessionCreationResponse struct {
	ProjectID          int64
	TimeTaken          int64
	EventCount         int
	LastEventTimestamp int64
	Status             *StatusBeam
	ErrorCode          int
}

type addSessionsByUserIDProjectIDFn struct {
	Config   *C.Configuration
	JobProps *AddSessionJobProps
}

func (f *addSessionsByUserIDProjectIDFn) StartBundle(ctx context.Context, emit func(UserIDSessionCreationResponse)) {
	log.Info("Initializing conf from StartBundle addSessionsByUserIDProjectIDFn")
	initConf(f.Config)
}

func (f *addSessionsByUserIDProjectIDFn) FinishBundle(ctx context.Context, emit func(UserIDSessionCreationResponse)) {
	log.Info("Closing DB Connection from FinishBundle addSessionsByUserIDProjectIDFn")
	C.GetServices().Db.Close()
	C.SafeFlushAllCollectors()
}

func (f *addSessionsByUserIDProjectIDFn) ProcessElement(ctx context.Context, eventsInput EventsByProjectResponse,
	emit func(UserIDSessionCreationResponse)) {

	logCtx := log.WithField("log_type", stepTrace).WithFields(log.Fields{
		"method": "addSessionsByUserIDProjectIDFn",
	})

	startTime := time.Now().UnixNano() / 1000
	if f.JobProps.LogInfo == 1 {
		getLogContext().WithField("log_type", stepTrace).WithFields(log.Fields{
			"project_id": eventsInput.ProjectID,
			"user_id":    eventsInput.UserID,
			"time_micro": startTime,
			"time_type":  "Start",
		}).Info("ProcessElement log start addSessionsByUserIDProjectIDFn")
	}

	userAddSessionResponse := UserIDSessionCreationResponse{
		ProjectID:  eventsInput.ProjectID,
		EventCount: len(eventsInput.Events),
		Status:     eventsInput.Status,
	}
	if eventsInput.Status.SeenFailure {
		// Logged this error already, exiting with failed status
		return
	}

	if len(eventsInput.Events) == 0 {
		logCtx.WithField("log_type", stepTrace).WithFields(log.Fields{"project_id": eventsInput.ProjectID, "user_id": eventsInput.UserID}).Info("No events found for the project. Exiting.")
		return
	}
	var errCode int

	// Min event time for given user.
	lastEventTimestamp := eventsInput.Events[len(eventsInput.Events)-1].Timestamp
	userAddSessionResponse.LastEventTimestamp = lastEventTimestamp

	noOfProcessedEvents, noOfCreated, isContinuedFirst, noOfUserPropUpdates,
		errCode := store.GetStore().AddSessionForUser(eventsInput.ProjectID, eventsInput.UserID, eventsInput.Events,
		f.JobProps.BufferTimeBeforeSessionCreateInSecs, eventsInput.SessionEventName.ID)

	if errCode == http.StatusInternalServerError || errCode == http.StatusBadRequest {
		msg := fmt.Sprintf("failed to get user events map on add session for project, errCode: %v", errCode)
		logCtx.WithField("log_type", stepTrace).
			WithField("error_code", errCode).
			WithField("events_input_UserID", eventsInput.UserID).
			WithField("events_input_events_count", len(eventsInput.Events)).Error(msg)
		return
	}

	userAddSessionResponse.Status.Status = session.StatusSuccess
	userAddSessionResponse.Status.Set(noOfProcessedEvents, noOfCreated, isContinuedFirst,
		noOfUserPropUpdates, false, userAddSessionResponse.Status)
	emit(userAddSessionResponse)

	if f.JobProps.LogInfo == 1 {
		getLogContext().WithField("log_type", stepTrace).WithFields(log.Fields{
			"project_id": eventsInput.ProjectID,
			"time_taken": time.Now().UnixNano()/1000 - startTime,
			"time_micro": time.Now().UnixNano() / 1000,
			"time_type":  "End",
		}).WithField("error_code", errCode).Info("Completed ProcessElement log end addSessionsByUserIDProjectIDFn")
	}
}

func emitProjectKeyUserSessionResponse(ctx context.Context, sessionResponse UserIDSessionCreationResponse,
	emit func(int64, UserIDSessionCreationResponse)) {
	emit(sessionResponse.ProjectID, sessionResponse)
}

func (f *reportProjectLevelSummary) StartBundle(ctx context.Context, emit func(ProjectStatusResponse)) {
	log.Info("Initializing conf from StartBundle reportProjectLevelSummary")
	initConf(f.Config)
}

func (f *reportProjectLevelSummary) FinishBundle(ctx context.Context, emit func(ProjectStatusResponse)) {
	log.Info("Closing DB Connection from FinishBundle reportProjectLevelSummary")
	C.GetServices().Db.Close()
	C.SafeFlushAllCollectors()
}

func (f *reportProjectLevelSummary) ProcessElement(ctx context.Context, projectID uint64, values func(*UserIDSessionCreationResponse) bool,
	emit func(ProjectStatusResponse)) {

	startTime := time.Now().UnixNano() / 1000
	logCtx := log.WithField("log_type", stepTrace).WithFields(log.Fields{
		"method":     "reportProjectLevelSummary",
		"time_micro": startTime,
		"project_id": projectID,
	})
	logCtx.Info("ProcessElement log reportProjectLevelSummary")
	var avgSuccessTimeTakenForUser float64
	totalSuccessTimeForUser := float64(0)

	successCount := 0
	failedCount := 0
	noOfUsers := 0
	noOfEvents := 0
	noOfEventsProcessed := 0
	noOfSessionsContinued := 0
	noOfSessionsCreated := 0
	noOfUserPropertiesUpdates := 0

	var userSessionResp UserIDSessionCreationResponse
	var minOfSessionAddedLastEventTimestamp int64
	minOfSessionAddedLastEventTimestamp = math.MaxInt64
	for values(&userSessionResp) {

		noOfUsers += 1
		if userSessionResp.Status.SeenFailure == false || userSessionResp.ErrorCode == http.StatusOK {
			// Successful execution for user
			successCount++
			noOfEvents += userSessionResp.EventCount
			noOfEventsProcessed += userSessionResp.Status.NoOfEventsProcessed
			noOfSessionsContinued += userSessionResp.Status.NoOfSessionsContinued
			noOfSessionsCreated += userSessionResp.Status.NoOfSessionsCreated
			noOfUserPropertiesUpdates += userSessionResp.Status.NoOfUserPropertiesUpdates
			if userSessionResp.LastEventTimestamp < minOfSessionAddedLastEventTimestamp {
				minOfSessionAddedLastEventTimestamp = userSessionResp.LastEventTimestamp
			}
			totalSuccessTimeForUser = totalSuccessTimeForUser + float64(userSessionResp.TimeTaken)
		} else {
			// Failed execution for user
			failedCount++
			logCtx.WithField("project_id", projectID).WithField("error_detail",
				userSessionResp.Status.ErrorDetail).Error("Error in add session beam job.")
			continue
		}
	}

	if successCount != 0 {
		avgSuccessTimeTakenForUser = totalSuccessTimeForUser / float64(successCount)
	}
	projectStatus := &StatusBeam{
		SeenFailure:               false,
		NoOfEvents:                noOfEvents,
		NoOfEventsProcessed:       noOfEventsProcessed,
		NoOfUsers:                 noOfUsers,
		NoOfSessionsContinued:     noOfSessionsContinued,
		NoOfSessionsCreated:       noOfSessionsCreated,
		NoOfUserPropertiesUpdates: noOfUserPropertiesUpdates,
		ErrorDetail:               fmt.Sprintf("Project had total users session successCount %d, failedCount %d", successCount, failedCount),
		Status:                    session.StatusSuccess,
		NoOfUserSessionFailures:   failedCount,
		NoOfUserSessionSuccess:    successCount,
	}
	if failedCount != 0 {
		projectStatus.SeenFailure = true
	}
	projectReport := ProjectStatusResponse{
		ProjectID:                  projectID,
		Status:                     projectStatus,
		AvgSuccessTimeTakenForUser: avgSuccessTimeTakenForUser,
		HasUpdatedJobLastTimestamp: true,
	}

	// Update next sessions start timestamp with min of last session
	// added event timestamp across all users for the project.
	errCode := store.GetStore().UpdateNextSessionStartTimestampForProject(
		userSessionResp.ProjectID, minOfSessionAddedLastEventTimestamp)
	if errCode != http.StatusAccepted {
		msg := "failed to UpdateNextSessionStartTimestamp For Project."
		logCtx.WithField("project_id", projectID).Error(msg)
		projectReport.HasUpdatedJobLastTimestamp = false
	}
	logCtx.WithField("project_id", projectID).
		WithField("error_code", errCode).
		WithField("status", projectReport.Status.Status).
		WithField("time_taken", time.Now().UnixNano()/1000-startTime).
		WithField("time", time.Now().UnixNano()/1000).
		WithField("seen_failure", projectReport.Status.SeenFailure).
		WithField("avg_success_time_taken_for_user", projectReport.AvgSuccessTimeTakenForUser).
		Info("project summary before emit.")

	emit(projectReport)
}

func emitDefaultKeyResponse(ctx context.Context, projectStatusResponse ProjectStatusResponse, emit func(uint64, ProjectStatusResponse)) {
	// Having 0 as key for collecting it on single node
	emit(0, projectStatusResponse)
}

func reportOverallJobSummary(commonKey uint64, values func(*ProjectStatusResponse) bool) string {

	logCtx := log.WithField("log_type", stepTrace).WithFields(log.Fields{
		"method": "reportOverallJobSummary",
	})
	startTime := time.Now().UnixNano() / 1000
	logCtx.WithField("start_time", startTime).Info("Processing reportOverallJobSummary")
	var projectStatusResponse ProjectStatusResponse
	projectsWithError := 0
	projectsWithSuccess := 0
	projectsNotModified := 0

	for values(&projectStatusResponse) {

		if projectStatusResponse.HasUpdatedJobLastTimestamp == false && projectStatusResponse.Status.NoOfEvents > 0 {
			projectsWithError++
			message := fmt.Sprintf("Had error(s) for project ID: %d", projectStatusResponse.ProjectID)
			C.PingHealthcheckForFailure(beam.PipelineOptions.Get("HealthchecksPingID"), message)
		} else if projectStatusResponse.Status.NoOfSessionsCreated == 0 {
			projectsNotModified++
		} else {
			projectsWithSuccess++
		}

		logCtx.WithField("project_id", projectStatusResponse.ProjectID).
			WithField("has_updated_next_session_job_time", projectStatusResponse.HasUpdatedJobLastTimestamp).
			WithField("no_of_events", projectStatusResponse.Status.NoOfEvents).
			WithField("no_of_users", projectStatusResponse.Status.NoOfUsers).
			WithField("no_of_sessionsContinued", projectStatusResponse.Status.NoOfSessionsContinued).
			WithField("no_of_sessionsCreated", projectStatusResponse.Status.NoOfSessionsCreated).
			WithField("no_of_userPropertiesUpdates", projectStatusResponse.Status.NoOfUserPropertiesUpdates).
			WithField("no_of_user_session_success", projectStatusResponse.Status.NoOfUserSessionSuccess).
			WithField("no_of_user_session_errors", projectStatusResponse.Status.NoOfUserSessionFailures).
			WithField("project_error_detail", projectStatusResponse.Status.ErrorDetail).
			Info("final project summary received")
	}

	summary := fmt.Sprintf("Total projects: passed = %d, failed = %d, notModified = %d",
		projectsWithSuccess, projectsWithError, projectsNotModified)
	logCtx.Info(summary)

	if projectsWithError != 0 {
		C.PingHealthcheckForFailure(beam.PipelineOptions.Get("HealthchecksPingID"), summary)
	} else {
		C.PingHealthcheckForSuccess(beam.PipelineOptions.Get("HealthchecksPingID"), summary)
	}
	logCtx.WithField("time_taken", time.Now().UnixNano()/1000-startTime).Info("Processing done reportOverallJobSummary")
	return summary
}

// ProjectStatusResponse Struct to share project's overall execution status and report.
type ProjectStatusResponse struct {
	ProjectID                  uint64
	Status                     *StatusBeam
	AvgSuccessTimeTakenForUser float64
	HasUpdatedJobLastTimestamp bool
}

type reportProjectLevelSummary struct {
	Config *C.Configuration
}

// AddSessionJobProps Job level properties  to store events pull by projectId response and emit.
type AddSessionJobProps struct {
	BufferTimeBeforeSessionCreateInSecs int64
	MaxLookbackTimestampInSec           int64
	StartTimestamp                      int64
	EndTimestamp                        int64
	LogInfo                             int
}

func main() {

	flag.Parse()
	if *env != "development" &&
		*env != "staging" &&
		*env != "production" {
		err := fmt.Errorf("env [ %s ] not recognised", *env)
		panic(err)
	}

	logCtx := log.WithField("log_type", stepTrace).WithField("start_timestamp", *startTimestamp).WithField("end_timestamp", *endTimestamp)
	if *endTimestamp > 0 && *startTimestamp == 0 {
		logCtx.Fatal("start_timestamp cannot be zero when start_timestamp is provided.")
	}
	if *startTimestamp > 0 && *endTimestamp == 0 {
		logCtx.Fatal("end_timestamp cannot be zero when start_timestamp is provided.")
	}
	if *startTimestamp > 0 && *endTimestamp <= *startTimestamp {
		logCtx.Fatal("end_timestamp cannot be lower than or equal to start_timestamp.")
	}

	var maxLookbackTimestamp int64
	if *maxLookbackHours > 0 {
		maxLookbackTimestamp = U.UnixTimeBeforeDuration(time.Hour * time.Duration(*maxLookbackHours))
	}

	jobProps := &AddSessionJobProps{
		BufferTimeBeforeSessionCreateInSecs: *bufferTimeBeforeCreateSessionInMins * 60,
		MaxLookbackTimestampInSec:           maxLookbackTimestamp,
		StartTimestamp:                      *startTimestamp,
		EndTimestamp:                        *endTimestamp,
		LogInfo:                             *logInfo,
	}

	registerStructs()
	beam.Init()

	// Creating a pipeline.
	p, s := beam.NewPipelineWithRoot()

	appName := "add_session_beam"
	healthcheckPingID := C.HealthcheckAddSessionPingID
	defer C.PingHealthcheckForPanic(appName, *env, healthcheckPingID)

	config := &C.Configuration{
		AppName:            appName,
		Env:                *env,
		GCPProjectID:       *gcpProjectID,
		GCPProjectLocation: *gcpProjectLocation,
		MemSQLInfo: C.DBConf{
			Host:        *memSQLHost,
			Port:        *memSQLPort,
			User:        *memSQLUser,
			Name:        *memSQLName,
			Password:    *memSQLPass,
			Certificate: *memSQLCertificate,
			AppName:     appName,
		},
		PrimaryDatastore:                    *primaryDatastore,
		RedisHost:                           *redisHost,
		RedisPort:                           *redisPort,
		RedisHostPersistent:                 *redisHostPersistent,
		RedisPortPersistent:                 *redisPortPersistent,
		SentryDSN:                           *sentryDSN,
		CacheSortedSet:                      *cacheSortedSet,
		AllowChannelGroupingForProjectIDs:   *allowChannelGroupingForProjectIDs,
		EnableOLTPQueriesMemSQLImprovements: *enableOLTPQueriesMemSQLImprovements,
		CaptureSourceInUsersTable:           *captureSourceInUsersTable,
	}

	beam.PipelineOptions.Set("HealthchecksPingID", "0e224a5e-01bd-454c-b361-651d303562c6")
	beam.PipelineOptions.Set("StartTime", fmt.Sprint(U.TimeNowUnix()))

	// Create initial PCollection for the projectIDs string passed to be processed.
	projectIDString := beam.Create(s, fmt.Sprintf("%s|%s", *projectIDs, *disabledProjectIDs))

	// Fetches allowed projects for all projectIDs and disabledProjectIDs
	allowedProjectId := beam.ParDo(s, &getAddSessionAllowedProjectsFn{Config: config}, projectIDString)

	reShuffledProjectIDs := beam.Reshuffle(s, allowedProjectId)

	// Pull events for each project by project ID.
	eventByProjectIdResponses := beam.ParDo(s, &pullEventsByProjectIdFn{Config: config, JobProps: jobProps}, reShuffledProjectIDs)

	reShuffledEventsByProjectIDs := beam.Reshuffle(s, eventByProjectIdResponses)

	// Add sessions by userID.
	userAddSessionResponses := beam.ParDo(s, &addSessionsByUserIDProjectIDFn{Config: config, JobProps: jobProps}, reShuffledEventsByProjectIDs)

	// Emit (project, user session response) for project level summary.
	userSessionProjectKeyResponses := beam.ParDo(s, emitProjectKeyUserSessionResponse, userAddSessionResponses)
	userSessionProjectGroupedResponses := beam.GroupByKey(s, userSessionProjectKeyResponses)
	projectReports := beam.ParDo(s, &reportProjectLevelSummary{Config: config}, userSessionProjectGroupedResponses)

	allProjectReport := beam.ParDo(s, emitDefaultKeyResponse, projectReports)
	allProjectResponsesGrouped := beam.GroupByKey(s, allProjectReport)
	_ = beam.ParDo(s, reportOverallJobSummary, allProjectResponsesGrouped)

	if err := beamx.Run(context.Background(), p); err != nil {
		log.Fatalf("Failed to execute job: %v", err)
	}

}
