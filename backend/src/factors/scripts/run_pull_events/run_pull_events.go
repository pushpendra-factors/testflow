package main

// Pull events that needs to be processed and write to file.
// Sample usage in terminal.
// export GOPATH=/Users/aravindmurthy/code/factors/backend/
// go run run_pull_events.go --project_id=1 --output_dir="" --end_time=""

import (
	"context"
	C "factors/config"
	"factors/filestore"
	"factors/merge"
	"factors/model/store"
	"factors/pattern"
	serviceDisk "factors/services/disk"
	serviceGCS "factors/services/gcstorage"
	T "factors/task"
	taskWrapper "factors/task/task_wrapper"
	U "factors/util"
	"flag"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/apache/beam/sdks/go/pkg/beam"
	log "github.com/sirupsen/logrus"
)

func registerStructs() {
	log.Info("Registering structs for beam")
	beam.RegisterType(reflect.TypeOf((*pattern.CounterEventFormat)(nil)).Elem())

	beam.RegisterType(reflect.TypeOf((*merge.RunBeamConfig)(nil)).Elem())
	beam.RegisterType(reflect.TypeOf((*merge.CUserIdsBeam)(nil)).Elem())
	beam.RegisterType(reflect.TypeOf((*merge.UidMap)(nil)).Elem())

	// do fn
	beam.RegisterType(reflect.TypeOf((*T.SortUsDoFn)(nil)).Elem())
}

func main() {
	env := flag.String("env", C.DEVELOPMENT, "")
	bucketNameFlag := flag.String("bucket_name", "/usr/local/var/factors/cloud_storage", "--bucket_name=/usr/local/var/factors/cloud_storage pass bucket name")
	tmpBucketNameFlag := flag.String("bucket_name_tmp", "/usr/local/var/factors/cloud_storage_tmp", "--bucket_name=/usr/local/var/factors/cloud_storage_tmp pass bucket name for tmp artifacts")
	archiveBucketNameFlag := flag.String("archive_bucket_name", "/usr/local/var/factors/cloud_storage_archive", "--bucket_name=/usr/local/var/factors/cloud_storage_archive pass archive bucket name")
	useBucketV2 := flag.Bool("use_bucket_v2", false, "Whether to use new bucketing system or not")

	localDiskTmpDirFlag := flag.String("local_disk_tmp_dir", "/usr/local/var/factors/local_disk/tmp", "--local_disk_tmp_dir=/usr/local/var/factors/local_disk/tmp pass directory.")

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
		"Optional: file type. A comma separated list of file types and supports '*' for all files. ex: 1,2,6,9") //refer to pull.FileType map
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
	var beamConfig merge.RunBeamConfig
	if !*useBucketV2 && *runBeam == 1 {
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

	defer U.NotifyOnPanic("Task#PullEvents", *env)

	appName := "pull_events_job"
	healthcheckPingID := C.HealthcheckPullEventsPingID
	defer C.PingHealthcheckForPanic(appName, *env, healthcheckPingID)

	config := &C.Configuration{
		AppName: appName,
		Env:     *env,
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
		log.Fatal("Failed to pull data. Init failed.")
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

	projectIdsToRun := make(map[int64]bool)
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

	configs := make(map[string]interface{})
	// Init cloud manager.
	if *useBucketV2 {
		var archiveCloudManager filestore.FileManager
		if *env == "development" {
			archiveCloudManager = serviceDisk.New(*archiveBucketNameFlag)
		} else {
			archiveCloudManager, err = serviceGCS.New(*archiveBucketNameFlag)
			if err != nil {
				log.WithField("error", err).Fatal("Failed to init archive cloud manager")
			}
		}
		configs["cloudManager"] = &archiveCloudManager
	} else {
		var cloudManagerTmp filestore.FileManager
		if *env == "development" {
			cloudManagerTmp = serviceDisk.New(*tmpBucketNameFlag)
		} else {
			cloudManagerTmp, err = serviceGCS.New(*tmpBucketNameFlag)
			if err != nil {
				log.WithField("error", err).Fatal("Failed to init cloud manager for tmp.")
			}
		}
		configs["cloudManagertmp"] = &cloudManagerTmp
		var cloudManager filestore.FileManager
		if *env == "development" {
			cloudManager = serviceDisk.New(*bucketNameFlag)
		} else {
			cloudManager, err = serviceGCS.New(*bucketNameFlag)
			if err != nil {
				log.WithField("error", err).Fatal("Failed to init cloud manager.")
			}
		}
		configs["cloudManager"] = &cloudManager
	}

	diskManager := serviceDisk.New(*localDiskTmpDirFlag)

	configs["diskManager"] = diskManager
	configs["hardPull"] = hardPull

	C.PingHealthcheckForStart(healthcheckPingID)

	if *useBucketV2 {
		if U.ContainsInt64InArray(fileTypes, 1) {
			fileTypesMapOnlyEvents := make(map[int64]bool)
			fileTypesMapOnlyEvents[1] = true
			configs["fileTypes"] = fileTypesMapOnlyEvents
			status := taskWrapper.TaskFuncWithProjectId("PullEventsDaily", *lookback, projectIdsArray, T.PullAllDataV2, configs)
			log.Info(status)
			var isSuccess bool = true
			for reason, message := range status {
				if message == false {
					C.PingHealthcheckForFailure(healthcheckPingID, reason+": pull events daily failure")
					isSuccess = false
					break
				}
			}
			if isSuccess {
				C.PingHealthcheckForSuccess(healthcheckPingID, "Pull Events Daily run success.")
			}
		}

		configs["fileTypes"] = fileTypesMap
		status := taskWrapper.TaskFuncWithProjectId("PullDataDaily", *lookback, projectIdsArray, T.PullAllDataV2, configs)
		log.Info(status)
		isSuccess := true
		for reason, message := range status {
			if message == false {
				C.PingHealthcheckForFailure(healthcheckPingID, reason+": pull data daily run failure")
				isSuccess = false
				break
			}
		}
		if isSuccess {
			C.PingHealthcheckForSuccess(healthcheckPingID, "Pull Data Daily run success.")
		}
	} else {
		configs["beamConfig"] = &beamConfig
		fileTypesMapOnlyEvents := make(map[int64]bool)
		fileTypesMapOnlyEvents[1] = true
		if *isWeeklyEnabled {
			configs["fileTypes"] = fileTypesMapOnlyEvents
			status := taskWrapper.TaskFuncWithProjectId("PullEventsWeeklyOnlyEvents", *lookback, projectIdsArray, T.PullAllDataV1, configs)
			log.Info(status)
			var isSuccess bool = true
			for reason, message := range status {
				if message == false {
					C.PingHealthcheckForFailure(healthcheckPingID, reason+": pull events weekly only events failure")
					isSuccess = false
					break
				}
			}
			if isSuccess {
				C.PingHealthcheckForSuccess(healthcheckPingID, "Pull Events Weekly only events run success.")
			}
		}

		if *isWeeklyEnabled {
			configs["fileTypes"] = fileTypesMap
			status := taskWrapper.TaskFuncWithProjectId("PullEventsWeekly", *lookback, projectIdsArray, T.PullAllDataV1, configs)
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
			status := taskWrapper.TaskFuncWithProjectId("PullEventsMonthly", *lookback, projectIdsArray, T.PullAllDataV1, configs)
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
			status := taskWrapper.TaskFuncWithProjectId("PullEventsQuarterly", *lookback, projectIdsArray, T.PullAllDataV1, configs)
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

}
