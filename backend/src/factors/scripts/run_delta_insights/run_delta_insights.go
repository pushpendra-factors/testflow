package main

// Sample usage in terminal.
// export GOPATH=~/factors/backend/
// go run run_delta_insights.go --project_id=<projectId> --model_id1=<modelId1> --model_id2=<modelId2> --env=development --local_disk_tmp_dir=/usr/local/var/factors/local_disk/tmp --bucket_name=<bucketName>

import (
	"context"
	C "factors/config"
	D "factors/delta"
	"factors/filestore"
	"factors/merge"
	"factors/model/model"
	"factors/model/store"
	"factors/pattern"
	"factors/pull"
	serviceDisk "factors/services/disk"
	serviceGCS "factors/services/gcstorage"
	taskWrapper "factors/task/task_wrapper"
	"factors/util"
	"flag"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/apache/beam/sdks/go/pkg/beam"
	_ "github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

func main() {
	mainRunDeltaInsights()
}

func registerStructs() {
	log.Info("Registering structs for beam")
	beam.RegisterType(reflect.TypeOf((*pattern.CounterEventFormat)(nil)).Elem())
	beam.RegisterType(reflect.TypeOf((*C.Configuration)(nil)).Elem())
	beam.RegisterType(reflect.TypeOf((*pull.CounterCampaignFormat)(nil)).Elem())

	beam.RegisterType(reflect.TypeOf((*merge.RunBeamConfig)(nil)).Elem())
	beam.RegisterType(reflect.TypeOf((*merge.CUserIdsBeam)(nil)).Elem())
	beam.RegisterType(reflect.TypeOf((*merge.UidMap)(nil)).Elem())

	// do fn
	beam.RegisterType(reflect.TypeOf((*merge.SortUsDoFn)(nil)).Elem())
	beam.RegisterType(reflect.TypeOf((*merge.SortAdDoFn)(nil)).Elem())
}

func mainRunDeltaInsights() {

	projectIdFlag := flag.String("project_ids", "",
		"Optional: Project Id. A comma separated list of project Ids and supports '*' for all projects. ex: 1,2,6,9")

	envFlag := flag.String("env", "development", "")
	localDiskTmpDirFlag := flag.String("local_disk_tmp_dir", "/usr/local/var/factors/local_disk/tmp", "--local_disk_tmp_dir=/usr/local/var/factors/local_disk/tmp pass directory")
	tmpBucketNameFlag := flag.String("bucket_name_tmp", "/usr/local/var/factors/cloud_storage_tmp", "--bucket_name=/usr/local/var/factors/cloud_storage_tmp pass bucket name for tmp artifacts")
	archiveBucketNameFlag := flag.String("archive_bucket_name", "/usr/local/var/factors/cloud_storage_archive", "--bucket_name=/usr/local/var/factors/cloud_storage_archive pass archive bucket name")
	sortedBucketNameFlag := flag.String("sorted_bucket_name", "/usr/local/var/factors/cloud_storage_sorted", "--bucket_name=/usr/local/var/factors/cloud_storage_sorted pass sorted data bucket name")
	modelBucketNameFlag := flag.String("model_bucket_name", "/usr/local/var/factors/cloud_storage_models", "--bucket_name=/usr/local/var/factors/cloud_storage_models pass model bucket name")

	hardPull := flag.Bool("hard_pull", false, "replace the files already present")
	runBeam := flag.Int("run_beam", 1, "run build seq on beam ")
	numWorkersFlag := flag.Int("num_beam_workers", 100, "Num of beam workers")
	useSortedFilesMerge := flag.Bool("use_sorted_merge", false, "whether to use sorted files (if possible) or achive files")

	fileTypesFlag := flag.String("file_types", "*",
		"Optional: file type. A comma separated list of file types and supports '*' for all files. ex: 1,2,6,9") //refer to pull.FileType map

	isWeeklyEnabled := flag.Bool("weekly_enabled", false, "")

	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	isPSCHost := flag.Int("memsql_is_psc_host", C.MemSQLDefaultDBParams.IsPSCHost, "")
	memSQLPort := flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser := flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName := flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass := flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	memSQLCertificate := flag.String("memsql_cert", "", "")
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypeMemSQL, "Primary datastore type as memsql or postgres")

	redisHost := flag.String("redis_host", "localhost", "")
	redisPort := flag.Int("redis_port", 6379, "")
	redisHostPersistent := flag.String("redis_host_ps", "localhost", "")
	redisPortPersistent := flag.Int("redis_port_ps", 6379, "")

	lookback := flag.Int("lookback", 30, "lookback_for_delta lookup")

	lookbackWindowForEventUserCache := flag.Int("lookback_window_event_user_cache",
		30, "look back window in cache for event/user cache(to get props for custom kpis")
	kValue := flag.Int("k", -1, "--k=10")
	whitelistedDashboardIds := flag.String("whitelisted_dashboard_ids", "*", "")
	skipWpi := flag.Bool("skip_wpi", false, "")
	skipWpi2 := flag.Bool("skip_wpi2", false, "")
	runKpi := flag.Bool("run_kpi", false, "")
	projectsFromDB := flag.Bool("projects_from_db", false, "")
	isMailerRun := flag.Bool("is_mailer_run", false, "")

	flag.Parse()

	if *envFlag != "development" &&
		*envFlag != "staging" &&
		*envFlag != "production" {
		err := fmt.Errorf("env [ %s ] not recognised", *envFlag)
		panic(err)
	}

	defer util.NotifyOnPanic("Task#DeltaInsights", *envFlag)

	//init beam
	var beamConfig merge.RunBeamConfig
	if *runBeam == 1 {
		log.Info("Initializing all beam constructs")
		registerStructs()
		beam.Init()
		beamConfig.RunOnBeam = true
		beamConfig.Env = *envFlag
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

	// init Config and DB.
	var appName = "compute_delta_insights"
	healthcheckPingID := C.HealthCheckWeeklyInsightsPingID
	defer C.PingHealthcheckForPanic(appName, *envFlag, healthcheckPingID)

	config := &C.Configuration{
		AppName: appName,
		Env:     *envFlag,
		MemSQLInfo: C.DBConf{
			Host:        *memSQLHost,
			IsPSCHost:   *isPSCHost,
			Port:        *memSQLPort,
			User:        *memSQLUser,
			Name:        *memSQLName,
			Password:    *memSQLPass,
			Certificate: *memSQLCertificate,
			AppName:     appName,
		},
		PrimaryDatastore:                *primaryDatastore,
		RedisHost:                       *redisHost,
		RedisPort:                       *redisPort,
		RedisHostPersistent:             *redisHostPersistent,
		RedisPortPersistent:             *redisPortPersistent,
		LookbackWindowForEventUserCache: *lookbackWindowForEventUserCache,
	}

	C.InitConf(config)
	beamConfig.DriverConfig = config
	C.InitSentryLogging(config.SentryDSN, config.AppName)
	err := C.InitDB(*config)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize DB")
	}
	// db := C.GetServices().Db
	// defer db.Close()

	C.InitRedis(config.RedisHost, config.RedisPort)
	C.InitRedisPersistent(config.RedisHostPersistent, config.RedisPortPersistent)
	log.WithFields(log.Fields{
		"Env":             *envFlag,
		"localDiskTmpDir": *localDiskTmpDirFlag,
		"ModelBucket":     *modelBucketNameFlag,
	}).Infoln("Initialising")

	diskManager := serviceDisk.New(*localDiskTmpDirFlag)

	projectIdsToRun := make(map[int64]bool, 0)
	if *projectsFromDB {
		//wi_projects, _ := store.GetStore().GetAllWeeklyInsightsEnabledProjects()
		wi_projects, _ := store.GetStore().GetAllProjectsWithFeatureEnabled(model.FEATURE_WEEKLY_INSIGHTS, false)
		for _, id := range wi_projects {
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
	for projectId, _ := range projectIdsToRun {
		projectIdsArray = append(projectIdsArray, projectId)
	}

	log.Info("config :", config)
	configs := make(map[string]interface{})

	var cloudManagerTmp filestore.FileManager
	if *envFlag == "development" {
		cloudManagerTmp = serviceDisk.New(*tmpBucketNameFlag)
	} else {
		cloudManagerTmp, err = serviceGCS.New(*tmpBucketNameFlag)
		if err != nil {
			log.WithField("error", err).Fatal("Failed to init cloud manager for tmp.")
		}
	}
	configs["tmpCloudManager"] = &cloudManagerTmp
	var archiveCloudManager filestore.FileManager
	var sortedCloudManager filestore.FileManager
	var modelCloudManager filestore.FileManager
	if *envFlag == "development" {
		modelCloudManager = serviceDisk.New(*modelBucketNameFlag)
		archiveCloudManager = serviceDisk.New(*archiveBucketNameFlag)
		sortedCloudManager = serviceDisk.New(*sortedBucketNameFlag)
	} else {
		modelCloudManager, err = serviceGCS.New(*modelBucketNameFlag)
		if err != nil {
			log.WithField("error", err).Fatal("Failed to init model cloud manager.")
		}
		archiveCloudManager, err = serviceGCS.New(*archiveBucketNameFlag)
		if err != nil {
			log.WithField("error", err).Fatal("Failed to init archive cloud manager")
		}
		sortedCloudManager, err = serviceGCS.New(*sortedBucketNameFlag)
		if err != nil {
			log.WithField("error", err).Fatal("Failed to init sorted data cloud manager")
		}
	}
	configs["modelCloudManager"] = &modelCloudManager
	configs["archiveCloudManager"] = &archiveCloudManager
	configs["sortedCloudManager"] = &sortedCloudManager

	configs["hardPull"] = *hardPull
	configs["diskManager"] = diskManager
	configs["beamConfig"] = &beamConfig
	configs["useSortedFilesMerge"] = *useSortedFilesMerge

	allDashboard, allDashboards, _ := C.GetProjectsFromListWithAllProjectSupport(*whitelistedDashboardIds, "")
	whitelistedIds := make(map[string]bool)
	if allDashboard {
		whitelistedIds["*"] = true
	} else {
		for id, _ := range allDashboards {
			whitelistedIds[fmt.Sprintf("%v", id)] = true
		}
	}
	configs["whitelistedDashboardUnits"] = whitelistedIds
	var k int = *kValue // Selecting all top features if k = -1.
	configs["k"] = k
	configs["skipWpi"] = (*skipWpi)
	configs["skipWpi2"] = (*skipWpi2)
	configs["runKpi"] = (*runKpi)

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
	configs["fileTypes"] = fileTypesMap

	// This job has dependency on pull_data
	if *isWeeklyEnabled && !(*isMailerRun) {
		taskName := "WIWeeklyV2"
		C.PingHealthcheckForStart(healthcheckPingID)
		status := taskWrapper.TaskFuncWithProjectId(taskName, *lookback, projectIdsArray, D.ComputeDeltaInsights, configs)
		log.Info(status)
		C.PingHealthCheckBasedOnStatus(status, healthcheckPingID)
	}

	if *isWeeklyEnabled && *isMailerRun {
		healthcheckPingID = C.HealthcheckMailWIPingID
		taskName := "WIWeeklyMailerV2"
		configs["run_type"] = "mailer"
		C.PingHealthcheckForStart(healthcheckPingID)
		status := taskWrapper.TaskFuncWithProjectId(taskName, *lookback, projectIdsArray, D.ComputeDeltaInsights, configs)
		log.Info(status)
		C.PingHealthCheckBasedOnStatus(status, healthcheckPingID)
	}
}
