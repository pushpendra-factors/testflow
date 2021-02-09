package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"

	C "factors/config"
	M "factors/model"
	U "factors/util"

	"github.com/apache/beam/sdks/go/pkg/beam"
	"github.com/apache/beam/sdks/go/pkg/beam/x/beamx"

	log "github.com/sirupsen/logrus"
)

/*
Sample query to run on:
Development (Direct Runner):
go run scripts/beam_dashboard_caching.go --project_id='*'

Staging (Dataflow runner):
go run scripts/run_beam_dashboard_caching.go --project_id='2,3' --runner dataflow --project factors-staging \
	--region us-west1 --temp_location gs://factors-staging-misc/beam/tmp/ \
	--staging_location gs://factors-staging-misc/beam/binaries/ \
	--worker_harness_container_image=apache/beam_go_sdk:latest --db_host=10.12.64.2 \
	--db_user autometa_ro --db_pass='<dbPass>' --redis_host='10.0.0.24' \
	--redis_port=8379 --subnetwork='regions/us-west1/subnetworks/us-west-1-factors-staging-subnet-1' \
	--num_workers=1 --max_num_workers=1 --zone='us-west1-b'
*/

var (
	env = flag.String("env", C.DEVELOPMENT, "Environment")

	dbHost = flag.String("db_host", "localhost", "")
	dbPort = flag.Int("db_port", 5432, "")
	dbUser = flag.String("db_user", "autometa", "")
	dbName = flag.String("db_name", "autometa", "")
	dbPass = flag.String("db_pass", "@ut0me7a", "")

	redisHost = flag.String("redis_host", "localhost", "")
	redisPort = flag.Int("redis_port", 6379, "")
	sentryDSN = flag.String("sentry_dsn", "", "Sentry DSN")

	projectIDs        = flag.String("project_id", "3", "Project ids to run for. * for all.")
	excludeProjectIDs = flag.String("exclude_project_id", "", "Comma separated project ids to exclude for the run")
	onlyWebAnalytics  = flag.Bool("only_web_analytics", false, "Cache only web analytics dashboards.")
	skipWebAnalytics  = flag.Bool("skip_web_analytics", false, "Skip the web analytics and run other.")
)

// Cache response types to be used in end of the job reporting.
const (
	CacheResponseTypeDashboardUnits string = "Dashboard Units"
	CacheResponseTypeWebAnalytics   string = "Website Analytics"
)

// CacheResponse Struct to store cache response and emit.
type CacheResponse struct {
	ProjectID, DashboardID, DashboardUnitID uint64
	From, To, TimeTaken                     int64
	ErrorCode                               int
	ErrorMessage                            string
	CacheResponseType                       string
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

	C.InitConf(config.Env)
	C.SetIsBeamPipeline()
	err := C.InitDBWithMaxIdleAndMaxOpenConn(config.DBInfo, 5, 2)
	if err != nil {
		// TODO(prateek): Check how a panic here will effect the pipeline.
		log.WithError(err).Panic("Failed to initalize db.")
	}
	C.InitRedisConnection(config.RedisHost, config.RedisPort, true, 20, 5)
	C.InitSentryLogging(config.SentryDSN, config.AppName)
}

func emitIndividualProjectID(ctx context.Context, projectIDsString string, emit func(string)) {
	allProjects, projectIDsMap, _ := C.GetProjectsFromListWithAllProjectSupport(projectIDsString, "")
	projectIDs := C.ProjectIdsFromProjectIdBoolMap(projectIDsMap)
	if allProjects {
		var errCode int
		projectIDs, errCode = M.GetAllProjectIDs()
		if errCode != http.StatusFound {
			return
		}
	}
	for _, pid := range projectIDs {
		emit(fmt.Sprint(pid))
	}
}

type getDashboardUnitCachePayloadsFn struct {
	Config *C.Configuration
}

func (f *getDashboardUnitCachePayloadsFn) StartBundle(ctx context.Context, emit func(M.BeamDashboardUnitCachePayload)) {
	log.Info("Initializing conf from StartBundle getDashboardUnitCachePayloadsFn")
	initConf(f.Config)
}

func (f *getDashboardUnitCachePayloadsFn) FinishBundle(ctx context.Context, emit func(M.BeamDashboardUnitCachePayload)) {
	log.Info("Closing DB Connection from FinishBundle getDashboardUnitCachePayloadsFn")
	C.GetServices().Db.Close()
	C.SafeFlushAllCollectors()
}

func (f *getDashboardUnitCachePayloadsFn) ProcessElement(ctx context.Context, projectsToRunString string,
	emit func(M.BeamDashboardUnitCachePayload)) {

	projectIDSplit := strings.Split(projectsToRunString, "|")
	stringProjectIDs := strings.TrimSpace(projectIDSplit[0])
	excludeProjectIDs := strings.TrimSpace(projectIDSplit[1])

	allProjects, projectIDsMap, excludeProjectIDsMap := C.GetProjectsFromListWithAllProjectSupport(
		stringProjectIDs, excludeProjectIDs)
	projectIDs := C.ProjectIdsFromProjectIdBoolMap(projectIDsMap)
	if allProjects {
		var errCode int
		allProjectIDs, errCode := M.GetAllProjectIDs()
		if errCode != http.StatusFound {
			return
		}
		for _, projectID := range allProjectIDs {
			if _, found := excludeProjectIDsMap[projectID]; !found {
				projectIDs = append(projectIDs, projectID)
			}
		}
	}

	for _, projectID := range projectIDs {
		dashboardUnits, errCode := M.GetDashboardUnitsForProjectID(uint64(projectID))
		if errCode != http.StatusFound {
			continue
		}

		for _, dashboardUnit := range dashboardUnits {
			queryClass, errMsg := M.GetQueryAndClassFromDashboardUnit(&dashboardUnit)
			if errMsg == "" && queryClass != M.QueryClassWeb {
				for _, rangeFunction := range U.QueryDateRangePresets {
					from, to := rangeFunction()
					cachePayload := M.BeamDashboardUnitCachePayload{
						DashboardUnit: dashboardUnit,
						QueryClass:    queryClass,
						Query:         dashboardUnit.Query,
						From:          from,
						To:            to,
					}
					emit(cachePayload)
				}
			}
		}
	}
}

type cacheDashboardUnitDoFn struct {
	Config *C.Configuration
}

func (f *cacheDashboardUnitDoFn) StartBundle(ctx context.Context, emit func(CacheResponse)) {
	log.Info("Initializing conf from StartBundle cacheDashboardUnitDoFn")
	initConf(f.Config)
}

func (f *cacheDashboardUnitDoFn) FinishBundle(ctx context.Context, emit func(CacheResponse)) {
	log.Info("Closing DB Connection from FinishBundle cacheDashboardUnitDoFn")
	C.GetServices().Db.Close()
	C.SafeFlushAllCollectors()
}

func (f *cacheDashboardUnitDoFn) ProcessElement(ctx context.Context,
	beamCachePayload M.BeamDashboardUnitCachePayload, emit func(CacheResponse)) {

	getLogContext().WithFields(log.Fields{
		"Time":     time.Now().UnixNano() / 1000,
		"TimeType": "Start",
	}).Info("ProcessElement log cacheDashboardUnitDoFn")

	baseQuery, err := M.DecodeQueryForClass(beamCachePayload.Query, beamCachePayload.QueryClass)
	if err != nil {
		return
	}
	baseQuery.SetQueryDateRange(beamCachePayload.From, beamCachePayload.To)
	cachePayload := M.DashboardUnitCachePayload{
		DashboardUnit: beamCachePayload.DashboardUnit,
		BaseQuery:     baseQuery,
	}
	startTime := U.TimeNowUnix()
	errCode, errMsg := M.CacheDashboardUnitForDateRange(cachePayload)
	timeTaken := U.TimeNowUnix() - startTime

	dashboardUnit := cachePayload.DashboardUnit
	from, to := cachePayload.BaseQuery.GetQueryDateRange()
	cacheResponse := CacheResponse{
		ProjectID:         dashboardUnit.ProjectID,
		DashboardID:       dashboardUnit.DashboardId,
		DashboardUnitID:   dashboardUnit.ID,
		From:              from,
		To:                to,
		ErrorCode:         errCode,
		ErrorMessage:      errMsg,
		TimeTaken:         timeTaken,
		CacheResponseType: CacheResponseTypeDashboardUnits,
	}
	getLogContext().WithFields(log.Fields{
		"Time":     time.Now().UnixNano() / 1000,
		"TimeType": "End",
	}).Info("ProcessElement log cacheDashboardUnitDoFn")
	emit(cacheResponse)
}

type getWebAnalyticsCachePayloadsFn struct {
	Config *C.Configuration
}

func (f *getWebAnalyticsCachePayloadsFn) StartBundle(ctx context.Context, emit func(M.WebAnalyticsCachePayload)) {
	log.Info("Initializing conf from StartBundle getWebAnalyticsCachePayloadsFn")
	initConf(f.Config)
}

func (f *getWebAnalyticsCachePayloadsFn) FinishBundle(ctx context.Context, emit func(M.WebAnalyticsCachePayload)) {
	log.Info("Closing DB Connection from FinishBundle getWebAnalyticsCachePayloadsFn")
	C.GetServices().Db.Close()
	C.SafeFlushAllCollectors()
}

func (f *getWebAnalyticsCachePayloadsFn) ProcessElement(ctx context.Context, projectsToRunString string,
	emit func(M.WebAnalyticsCachePayload)) {

	projectIDSplit := strings.Split(projectsToRunString, "|")
	stringProjectIDs := strings.TrimSpace(projectIDSplit[0])
	excludeProjectIDs := strings.TrimSpace(projectIDSplit[1])

	projectIDs := M.GetWebAnalyticsEnabledProjectIDsFromList(stringProjectIDs, excludeProjectIDs)
	for _, projectID := range projectIDs {
		cachePayloads, errCode, errMsg := M.GetWebAnalyticsCachePayloadsForProject(projectID)
		if errCode != http.StatusFound {
			log.Error("Error getting web analytics cache payloads ", errMsg)
			return
		}
		for index := range cachePayloads {
			emit(cachePayloads[index])
		}
	}
}

type cacheWebAnalyticsDoFn struct {
	Config *C.Configuration
}

func (f *cacheWebAnalyticsDoFn) StartBundle(ctx context.Context, emit func(CacheResponse)) {
	log.Info("Initializing conf from StartBundle cacheWebAnalyticsDoFn")
	initConf(f.Config)
}

func (f *cacheWebAnalyticsDoFn) FinishBundle(ctx context.Context, emit func(CacheResponse)) {
	log.Info("Closing DB Connection from FinishBundle cacheWebAnalyticsDoFn")
	C.GetServices().Db.Close()
	C.SafeFlushAllCollectors()
}

func (f *cacheWebAnalyticsDoFn) ProcessElement(ctx context.Context,
	cachePayload M.WebAnalyticsCachePayload, emit func(CacheResponse)) {

	getLogContext().WithFields(log.Fields{
		"Time":     time.Now().UnixNano() / 1000,
		"TimeType": "Start",
	}).Info("ProcessElement log cacheWebAnalyticsDoFn")
	startTime := U.TimeNowUnix()
	errCode := M.CacheWebsiteAnalyticsForDateRange(cachePayload)
	timeTaken := U.TimeNowUnix() - startTime

	cacheResponse := CacheResponse{
		ProjectID:         cachePayload.ProjectID,
		DashboardID:       cachePayload.DashboardID,
		From:              cachePayload.From,
		To:                cachePayload.To,
		ErrorCode:         errCode,
		TimeTaken:         timeTaken,
		CacheResponseType: CacheResponseTypeWebAnalytics,
	}
	getLogContext().WithFields(log.Fields{
		"Time":     time.Now().UnixNano() / 1000,
		"TimeType": "End",
	}).Info("ProcessElement log cacheWebAnalyticsDoFn")
	emit(cacheResponse)
}

func emitProjectKeyCacheResponse(ctx context.Context, cacheResponse CacheResponse, emit func(uint64, CacheResponse)) {
	emit(cacheResponse.ProjectID, cacheResponse)
}

func emitCommonKeyCacheResponse(ctx context.Context, cacheResponse CacheResponse, emit func(uint64, CacheResponse)) {
	emit(0, cacheResponse)
}

func reportProjectLevelSummary(projectID uint64, values func(*CacheResponse) bool) string {
	var timeTakenForProject int64
	var successCount, failedCount int64

	var cacheResponse CacheResponse
	var cacheResponseType string
	for values(&cacheResponse) {
		cacheResponseType = cacheResponse.CacheResponseType
		timeTakenForProject += cacheResponse.TimeTaken
		if cacheResponse.ErrorCode == http.StatusOK {
			successCount++
		} else {
			failedCount++
			log.WithFields(log.Fields{
				"Method":          "reportProjectLevelSummary",
				"ProjectID":       projectID,
				"DashboardID":     cacheResponse.DashboardID,
				"DashboardUnitID": cacheResponse.DashboardUnitID,
				"Type":            cacheResponse.CacheResponseType,
				"ErrorMessage":    cacheResponse.ErrorMessage,
			}).Info("Error in caching.")
		}
	}

	timeTakenString := U.SecondsToHMSString(timeTakenForProject)
	log.WithFields(log.Fields{
		"Method":          "reportProjectLevelSummary",
		"ProjectID":       projectID,
		"TimeTaken":       timeTakenForProject,
		"TimeTakenString": timeTakenString,
		"Success":         successCount,
		"Failed":          failedCount,
	}).Infof("%s: Completed caching for project.", cacheResponseType)
	return fmt.Sprintf("ProjectID: %d, TimeTaken: %s, Success: %d, Failed: %d",
		projectID, timeTakenString, successCount, failedCount)
}

func reportOverallJobSummary(commonKey uint64, values func(*CacheResponse) bool) string {
	overallStats := make(map[string]map[string]int64)
	overallStats[CacheResponseTypeDashboardUnits] = make(map[string]int64)
	overallStats[CacheResponseTypeWebAnalytics] = make(map[string]int64)

	var cacheResponse CacheResponse
	for values(&cacheResponse) {
		overallStats[cacheResponse.CacheResponseType]["TimeTaken"] += cacheResponse.TimeTaken
		if cacheResponse.ErrorCode == http.StatusOK {
			overallStats[cacheResponse.CacheResponseType]["SuccessCount"]++
		} else {
			overallStats[cacheResponse.CacheResponseType]["FailedCount"]++
		}
	}
	unitsTimeTakenString := U.SecondsToHMSString(overallStats[CacheResponseTypeDashboardUnits]["TimeTaken"])
	webTimeTakenString := U.SecondsToHMSString(overallStats[CacheResponseTypeWebAnalytics]["TimeTaken"])

	jobStartTime, _ := strconv.ParseInt(beam.PipelineOptions.Get("StartTime"), 10, 64)
	jobEndTime := U.TimeNowUnix()
	overallTimeTaken := U.SecondsToHMSString(jobEndTime - jobStartTime)

	unitsSummary := fmt.Sprintf("Time taken for dashboard units: %s, Success: %d, Failed: %d", unitsTimeTakenString,
		overallStats[CacheResponseTypeDashboardUnits]["SuccessCount"], overallStats[CacheResponseTypeDashboardUnits]["FailedCount"])
	webSummary := fmt.Sprintf("Time taken for web analytics: %s, Success: %d, Failed: %d", webTimeTakenString,
		overallStats[CacheResponseTypeWebAnalytics]["SuccessCount"], overallStats[CacheResponseTypeWebAnalytics]["FailedCount"])
	message := "Overall time taken: " + overallTimeTaken + " | " + unitsSummary + " | " + webSummary

	log.WithFields(log.Fields{
		"Method": "reportOverallJobSummary",
	}).Info(message)
	if overallStats[CacheResponseTypeDashboardUnits]["FailedCount"] > 0 || overallStats[CacheResponseTypeWebAnalytics]["FailedCount"] > 0 {
		C.PingHealthcheckForFailure(beam.PipelineOptions.Get("HealthchecksPingID"), message)
	} else {
		C.PingHealthcheckForSuccess(beam.PipelineOptions.Get("HealthchecksPingID"), message)
	}
	return message
}

func registerStructs() {
	beam.RegisterType(reflect.TypeOf((*M.DashboardUnit)(nil)).Elem())
	beam.RegisterType(reflect.TypeOf((*M.BeamDashboardUnitCachePayload)(nil)).Elem())

	beam.RegisterType(reflect.TypeOf((*getDashboardUnitCachePayloadsFn)(nil)).Elem())
	beam.RegisterType(reflect.TypeOf((*cacheDashboardUnitDoFn)(nil)).Elem())

	beam.RegisterType(reflect.TypeOf((*M.WebAnalyticsCachePayload)(nil)).Elem())
	beam.RegisterType(reflect.TypeOf((*getWebAnalyticsCachePayloadsFn)(nil)).Elem())
	beam.RegisterType(reflect.TypeOf((*cacheWebAnalyticsDoFn)(nil)).Elem())
}

type cachingJobSummary struct {
	TimeTaken                 int64
	SuccessCount, FailedCount int64
}

type jobStatsAccumulator struct {
	WebAnalyticsSummary   map[uint64]cachingJobSummary
	DashboardUnitsSummary map[uint64]cachingJobSummary
}

type reportJobSummaryCombineFn struct{}

func (f *reportJobSummaryCombineFn) CreateAccumulator() jobStatsAccumulator {
	return jobStatsAccumulator{
		DashboardUnitsSummary: make(map[uint64]cachingJobSummary),
		WebAnalyticsSummary:   make(map[uint64]cachingJobSummary),
	}
}

func (f *reportJobSummaryCombineFn) AddInput(a jobStatsAccumulator, value interface{}) jobStatsAccumulator {
	_, ok := value.(CacheResponse)
	if ok {
		castedValue := value.(CacheResponse)
		var successCount, failedCount int64
		if castedValue.ErrorCode == http.StatusOK {
			successCount++
		} else {
			failedCount++
		}
		if _, found := a.DashboardUnitsSummary[castedValue.ProjectID]; found {
			a.DashboardUnitsSummary[castedValue.ProjectID] = cachingJobSummary{
				TimeTaken:    a.DashboardUnitsSummary[castedValue.ProjectID].TimeTaken + castedValue.TimeTaken,
				SuccessCount: a.DashboardUnitsSummary[castedValue.ProjectID].SuccessCount + successCount,
				FailedCount:  a.DashboardUnitsSummary[castedValue.ProjectID].FailedCount + failedCount,
			}
		} else {
			a.DashboardUnitsSummary[castedValue.ProjectID] = cachingJobSummary{
				TimeTaken:    castedValue.TimeTaken,
				SuccessCount: successCount,
				FailedCount:  failedCount,
			}
		}
	} else {
		castedValue := value.(CacheResponse)
		var successCount, failedCount int64
		if castedValue.ErrorCode == http.StatusOK {
			successCount++
		} else {
			failedCount++
		}
		if _, found := a.WebAnalyticsSummary[castedValue.ProjectID]; found {
			a.WebAnalyticsSummary[castedValue.ProjectID] = cachingJobSummary{
				TimeTaken:    a.WebAnalyticsSummary[castedValue.ProjectID].TimeTaken + castedValue.TimeTaken,
				SuccessCount: a.WebAnalyticsSummary[castedValue.ProjectID].SuccessCount + successCount,
				FailedCount:  a.WebAnalyticsSummary[castedValue.ProjectID].FailedCount + failedCount,
			}
		} else {
			a.WebAnalyticsSummary[castedValue.ProjectID] = cachingJobSummary{
				TimeTaken:    castedValue.TimeTaken,
				SuccessCount: successCount,
				FailedCount:  failedCount,
			}
		}
	}
	return a
}

func (f *reportJobSummaryCombineFn) MergeAccumulators(a, b jobStatsAccumulator) jobStatsAccumulator {
	mergedAccumulator := jobStatsAccumulator{
		DashboardUnitsSummary: make(map[uint64]cachingJobSummary),
		WebAnalyticsSummary:   make(map[uint64]cachingJobSummary),
	}

	for _, accum := range append([]jobStatsAccumulator{a}, b) {
		for projectID, jobSummary := range accum.DashboardUnitsSummary {
			mergedAccumulator.DashboardUnitsSummary[projectID] = jobSummary
		}

		for projectID, jobSummary := range accum.WebAnalyticsSummary {
			mergedAccumulator.WebAnalyticsSummary[projectID] = jobSummary
		}
	}
	return mergedAccumulator
}

func (f *reportJobSummaryCombineFn) ExtractOutput(a jobStatsAccumulator) {
	var overallTimeTaken, overallSuccessCount, overallFailedCount int64
	for projectID, jobSummary := range a.DashboardUnitsSummary {
		log.WithFields(log.Fields{
			"ProjectID":       projectID,
			"TimeTaken":       jobSummary.TimeTaken,
			"TimeTakenString": U.SecondsToHMSString(jobSummary.TimeTaken),
		}).Info("Completed dashboard unit caching for project %d. Success: %d, Failed: %d",
			projectID, jobSummary.SuccessCount, jobSummary.FailedCount)

		overallTimeTaken += jobSummary.TimeTaken
		overallSuccessCount += jobSummary.SuccessCount
		overallFailedCount += jobSummary.FailedCount
	}
	overallSummary := fmt.Sprintf("Time taken for dashboard units: %s (Success: %d, Failed: %d).",
		U.SecondsToHMSString(overallTimeTaken), overallSuccessCount, overallFailedCount)

	var webOverallTimeTaken, webOverallSuccessCount, webOverallFailedCount int64
	for projectID, jobSummary := range a.WebAnalyticsSummary {
		log.WithFields(log.Fields{
			"ProjectID":       projectID,
			"TimeTaken":       jobSummary.TimeTaken,
			"TimeTakenString": U.SecondsToHMSString(jobSummary.TimeTaken),
		}).Info("Completed web analytics caching for project %d. Success: %d, Failed: %d",
			projectID, jobSummary.SuccessCount, jobSummary.FailedCount)

		webOverallTimeTaken += jobSummary.TimeTaken
		webOverallSuccessCount += jobSummary.SuccessCount
		webOverallFailedCount += jobSummary.FailedCount
	}
	webOverallSummary := fmt.Sprintf("Time taken for web analytics: %s (Success: %d, Failed: %d).",
		U.SecondsToHMSString(webOverallTimeTaken), webOverallSuccessCount, webOverallFailedCount)

	message := overallSummary + " " + webOverallSummary
	if overallFailedCount > 0 || webOverallFailedCount > 0 {
		C.PingHealthcheckForFailure(beam.PipelineOptions.Get("HealthchecksPingID"), message)
	} else {
		C.PingHealthcheckForSuccess(beam.PipelineOptions.Get("HealthchecksPingID"), message)
	}
}

// TODO(prateek): Check a way to add handling for panic and worker errors.
func main() {
	flag.Parse()
	registerStructs()
	beam.Init()

	if *skipWebAnalytics && *onlyWebAnalytics {
		log.Fatal("Both skip and only web analytics can not be set")
	}

	// Creating a pipeline.
	p, s := beam.NewPipelineWithRoot()

	appName := "beam_dashboard_caching"
	config := &C.Configuration{
		AppName: appName,
		Env:     *env,
		DBInfo: C.DBConf{
			Host:     *dbHost,
			Port:     *dbPort,
			User:     *dbUser,
			Name:     *dbName,
			Password: *dbPass,
			AppName:  appName,
		},
		RedisHost: *redisHost,
		RedisPort: *redisPort,
		SentryDSN: *sentryDSN,
	}
	beam.PipelineOptions.Set("HealthchecksPingID", "ecb259b9-4ff8-4825-b989-81d47bd34d93")
	beam.PipelineOptions.Set("StartTime", fmt.Sprint(U.TimeNowUnix()))

	// Create initial PCollection for the projectIDs string passed to be processed.
	projectIDString := beam.Create(s, fmt.Sprintf("%s|%s", *projectIDs, *excludeProjectIDs))

	var cacheResponses, webAnalyticsCacheResponses beam.PCollection
	if !*onlyWebAnalytics {
		// Fetches dashboard_units for all project_id and emits a unit of work along with query and date range.
		cachePayloads := beam.ParDo(s, &getDashboardUnitCachePayloadsFn{Config: config}, projectIDString)
		reShuffledCachePayloads := beam.Reshuffle(s, cachePayloads)
		// Processes each DashboardUnitCachePayload and emits a cache response which includes errCode, errMsg, time etc.
		cacheResponses = beam.ParDo(s, &cacheDashboardUnitDoFn{Config: config}, reShuffledCachePayloads)

		projectKeyCacheResponses := beam.ParDo(s, emitProjectKeyCacheResponse, cacheResponses)
		projectGroupedCacheResponses := beam.GroupByKey(s, projectKeyCacheResponses)
		_ = beam.ParDo(s, reportProjectLevelSummary, projectGroupedCacheResponses)
	}

	if !*skipWebAnalytics {
		webAnalyticsCachePayloads := beam.ParDo(s, &getWebAnalyticsCachePayloadsFn{Config: config}, projectIDString)
		reShuffledWebAnalyticsCachePayloads := beam.Reshuffle(s, webAnalyticsCachePayloads)

		webAnalyticsCacheResponses = beam.ParDo(s, &cacheWebAnalyticsDoFn{Config: config}, reShuffledWebAnalyticsCachePayloads)

		// Emit (project, cache response) for project level summary.
		webAnalyticsProjectKeyCacheResponses := beam.ParDo(s, emitProjectKeyCacheResponse, webAnalyticsCacheResponses)
		webAnalyticsProjectGroupedCacheResponses := beam.GroupByKey(s, webAnalyticsProjectKeyCacheResponses)
		_ = beam.ParDo(s, reportProjectLevelSummary, webAnalyticsProjectGroupedCacheResponses)
	}

	// To log an overall job summary.
	var allCacheResponses beam.PCollection
	if *onlyWebAnalytics {
		allCacheResponses = webAnalyticsCacheResponses
	} else if *skipWebAnalytics {
		allCacheResponses = cacheResponses
	} else {
		allCacheResponses = beam.Flatten(s, cacheResponses, webAnalyticsCacheResponses)
	}
	allCommonKeyCacheResponses := beam.ParDo(s, emitCommonKeyCacheResponse, allCacheResponses)
	allCacheResponsesGrouped := beam.GroupByKey(s, allCommonKeyCacheResponses)
	_ = beam.ParDo(s, reportOverallJobSummary, allCacheResponsesGrouped)

	if err := beamx.Run(context.Background(), p); err != nil {
		log.Fatalf("Failed to execute job: %v", err)
	}
}
