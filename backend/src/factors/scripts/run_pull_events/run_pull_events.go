package main

// Pull events that needs to be processed and write to file.
// Sample usage in terminal.
// export GOPATH=/Users/aravindmurthy/code/factors/backend/
// go run run_pull_events.go --project_id=1 --output_dir="" --end_time=""

import (
	"context"
	C "factors/config"
	"factors/filestore"
	"factors/model/store"
	serviceDisk "factors/services/disk"
	serviceGCS "factors/services/gcstorage"
	T "factors/task"
	taskWrapper "factors/task/task_wrapper"
	"factors/util"
	"flag"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/apache/beam/sdks/go/pkg/beam"
	log "github.com/sirupsen/logrus"
)

const (
	DayInSecs   = 24 * 60 * 60
	MonthInSecs = 31 * DayInSecs
	WeekInSecs  = 7 * DayInSecs
)

func registerStructs() {
	log.Info("Registering structs for beam")
	beam.RegisterType(reflect.TypeOf((*T.CounterCampaignFormat)(nil)).Elem())
	beam.RegisterType(reflect.TypeOf((*T.CounterUserFormat)(nil)).Elem())
	beam.RegisterType(reflect.TypeOf((*T.RunBeamConfig)(nil)).Elem())
	beam.RegisterType(reflect.TypeOf((*T.CUserIdsBeam)(nil)).Elem())
	beam.RegisterType(reflect.TypeOf((*T.UidMap)(nil)).Elem())

	// do fn
	beam.RegisterType(reflect.TypeOf((*T.SortUsDoFn)(nil)).Elem())
}

func main() {
	env := flag.String("env", C.DEVELOPMENT, "")
	bucketNameFlag := flag.String("bucket_name", "/usr/local/var/factors/cloud_storage", "--bucket_name=/usr/local/var/factors/cloud_storage pass bucket name")
	localDiskTmpDirFlag := flag.String("local_disk_tmp_dir", "/usr/local/var/factors/local_disk/tmp", "--local_disk_tmp_dir=/usr/local/var/factors/local_disk/tmp pass directory.")

	dbHost := flag.String("db_host", C.PostgresDefaultDBParams.Host, "")
	dbPort := flag.Int("db_port", C.PostgresDefaultDBParams.Port, "")
	dbUser := flag.String("db_user", C.PostgresDefaultDBParams.User, "")
	dbName := flag.String("db_name", C.PostgresDefaultDBParams.Name, "")
	dbPass := flag.String("db_pass", C.PostgresDefaultDBParams.Password, "")

	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	memSQLPort := flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser := flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName := flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass := flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	memSQLCertificate := flag.String("memsql_cert", "", "")
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypeMemSQL, "Primary datastore type as memsql or postgres")
	sentryDSN := flag.String("sentry_dsn", "", "Sentry DSN")

	isWeeklyEnabled := flag.Bool("weekly_enabled", false, "")
	isMonthlyEnabled := flag.Bool("monthly_enabled", false, "")
	isQuarterlyEnabled := flag.Bool("quarterly_enabled", false, "")

	hardPull := flag.Bool("hard_pull", false, "replace the files already present")

	fileTypesFlag := flag.String("file_types", "*",
		"Optional: file type. A comma separated list of file types and supports '*' for all files. ex: 1,2,6,9") //refer to T.fileType map
	projectIdFlag := flag.String("project_ids", "",
		"Optional: Project Id. A comma separated list of project Ids and supports '*' for all projects. ex: 1,2,6,9")
	lookback := flag.Int("lookback", 30, "lookback_for_delta lookup")
	projectsFromDB := flag.Bool("projects_from_db", false, "")
	redisHost := flag.String("redis_host", "localhost", "")
	redisPort := flag.Int("redis_port", 6379, "")
	redisHostPersistent := flag.String("redis_host_ps", "localhost", "")
	redisPortPersistent := flag.Int("redis_port_ps", 6379, "")

	runBeam := flag.Int("run_beam", 1, "run build seq on beam ")
	numWorkersFlag := flag.Int("num_beam_workers", 100, "Num of beam workers")

	flag.Parse()

	if *env != "development" &&
		*env != "staging" &&
		*env != "production" {
		err := fmt.Errorf("env [ %s ] not recognised", *env)
		panic(err)
	}

	//init beam
	var beamConfig T.RunBeamConfig
	if *runBeam == 1 {
		log.Info("Initializing all beam constructs")
		registerStructs()
		beam.Init()
		beamConfig.RunOnBeam = true
		beamConfig.Env = *env
		beamConfig.Ctx = context.Background()
		beamConfig.Pipe = beam.NewPipeline()
		beamConfig.Scp = beamConfig.Pipe.Root()
		beamConfig.NumWorker = *numWorkersFlag
		if beam.Initialized() {
			log.Info("Initalized all Beam Inits")
		} else {
			log.Fatal("unable to initialize runners")

		}
	} else {
		beamConfig.RunOnBeam = false
	}

	defer util.NotifyOnPanic("Task#PullEvents", *env)

	appName := "pull_events_job"
	healthcheckPingID := C.HealthcheckPullEventsPingID
	defer C.PingHealthcheckForPanic(appName, *env, healthcheckPingID)

	config := &C.Configuration{
		AppName: appName,
		Env:     *env,
		DBInfo: C.DBConf{
			Host:     *dbHost,
			Port:     *dbPort,
			User:     *dbUser,
			Name:     *dbName,
			Password: *dbPass,
		},
		MemSQLInfo: C.DBConf{
			Host:        *memSQLHost,
			Port:        *memSQLPort,
			User:        *memSQLUser,
			Name:        *memSQLName,
			Password:    *memSQLPass,
			Certificate: *memSQLCertificate,
			AppName:     appName,
		},
		PrimaryDatastore:    *primaryDatastore,
		SentryDSN:           *sentryDSN,
		RedisHost:           *redisHost,
		RedisPort:           *redisPort,
		RedisHostPersistent: *redisHostPersistent,
		RedisPortPersistent: *redisPortPersistent,
	}

	C.InitConf(config)
	beamConfig.DriverConfig = config
	C.InitSentryLogging(config.SentryDSN, config.AppName)

	// Initialize configs and connections and close with defer.
	err := C.InitDB(*config)
	if err != nil {
		log.Fatal("Failed to pull events. Init failed.")
	}
	db := C.GetServices().Db
	defer db.Close()

	fileTypesList := strings.TrimSpace(*fileTypesFlag)
	var fileTypes []int64
	if fileTypesList == "*" {
		fileTypes = []int64{1, 2, 3, 4, 5, 6, 7}
	} else {
		fileTypes = C.GetTokensFromStringListAsUint64(fileTypesList)
	}
	fileTypesMap := make(map[int64]bool)
	for _, i := range fileTypes {
		fileTypesMap[i] = true
	}

	projectIdsToRun := make(map[int64]bool, 0)
	if *projectsFromDB {
		wi_projects, _ := store.GetStore().GetAllWeeklyInsightsEnabledProjects()
		explain_projects, _ := store.GetStore().GetAllExplainEnabledProjects()
		for _, id := range wi_projects {
			projectIdsToRun[id] = true
		}
		for _, id := range explain_projects {
			projectIdsToRun[id] = true
		}

	} else {
		var allProjects bool
		allProjects, projectIdsToRun, _ = C.GetProjectsFromListWithAllProjectSupport(*projectIdFlag, "")
		if allProjects {
			projectIDs, errCode := store.GetStore().GetAllProjectIDs()
			if errCode != http.StatusFound {
				log.Fatal("Failed to get all projects and project_ids set to '*'.")
			}
			for _, projectID := range projectIDs {
				projectIdsToRun[projectID] = true
			}
		}
	}

	projectIdsArray := make([]int64, 0)
	for projectId := range projectIdsToRun {
		projectIdsArray = append(projectIdsArray, projectId)
	}
	// Init cloud manager.
	var cloudManager filestore.FileManager
	if *env == "development" {
		cloudManager = serviceDisk.New(*bucketNameFlag)
	} else {
		cloudManager, err = serviceGCS.New(*bucketNameFlag)
		if err != nil {
			log.WithField("error", err).Fatal("Failed to init cloud manager.")
		}
	}

	diskManager := serviceDisk.New(*localDiskTmpDirFlag)

	configs := make(map[string]interface{})
	configs["diskManager"] = diskManager
	configs["cloudManager"] = &cloudManager
	configs["hardPull"] = hardPull
	configs["beamConfig"] = &beamConfig

	fileTypesMapOnlyEvents := make(map[int64]bool)
	fileTypesMapOnlyEvents[1] = true
	C.PingHealthcheckForStart(healthcheckPingID)
	if *isWeeklyEnabled {
		configs["fileTypes"] = fileTypesMapOnlyEvents
		configs["modelType"] = T.ModelTypeWeek
		status := taskWrapper.TaskFuncWithProjectId("PullEventsWeeklyOnlyEvents", *lookback, projectIdsArray, T.PullAllData, configs)
		log.Info(status)
		var isSuccess bool = true
		for reason, message := range status {
			if message == false {
				C.PingHealthcheckForFailure(healthcheckPingID, reason+": pull events weekly failure")
				isSuccess = false
				break
			}
		}
		if isSuccess {
			C.PingHealthcheckForSuccess(healthcheckPingID, "Pull Events Weekly run success.")
		}
	}

	if *isWeeklyEnabled {
		configs["fileTypes"] = fileTypesMap
		configs["modelType"] = T.ModelTypeWeek
		status := taskWrapper.TaskFuncWithProjectId("PullEventsWeekly", *lookback, projectIdsArray, T.PullAllData, configs)
		log.Info(status)
		var isSuccess bool = true
		for reason, message := range status {
			if message == false {
				C.PingHealthcheckForFailure(healthcheckPingID, reason+": pull events weekly run failure")
				isSuccess = false
				break
			}
		}
		if isSuccess {
			C.PingHealthcheckForSuccess(healthcheckPingID, "Pull Events Weekly run success.")
		}
	}

	if *isMonthlyEnabled {
		configs["fileTypes"] = fileTypesMapOnlyEvents
		configs["modelType"] = T.ModelTypeMonth
		status := taskWrapper.TaskFuncWithProjectId("PullEventsMonthly", *lookback, projectIdsArray, T.PullAllData, configs)
		log.Info(status)
		var isSuccess bool = true
		for reason, message := range status {
			if message == false {
				C.PingHealthcheckForFailure(healthcheckPingID, reason+": pull events monthly run failure")
				isSuccess = false
				break
			}
		}
		if isSuccess {
			C.PingHealthcheckForSuccess(healthcheckPingID, "Pull Events Monthly run success.")
		}
	}

	if *isQuarterlyEnabled {
		configs["fileTypes"] = fileTypesMapOnlyEvents
		configs["modelType"] = T.ModelTypeQuarter
		status := taskWrapper.TaskFuncWithProjectId("PullEventsQuarterly", *lookback, projectIdsArray, T.PullAllData, configs)
		log.Info(status)
		var isSuccess bool = true
		for reason, message := range status {
			if message == false {
				C.PingHealthcheckForFailure(healthcheckPingID, reason+": pull events quarterly run failure")
				isSuccess = false
				break
			}
		}
		if isSuccess {
			C.PingHealthcheckForSuccess(healthcheckPingID, "Pull Events quarterly run success.")
		}
	}
}
