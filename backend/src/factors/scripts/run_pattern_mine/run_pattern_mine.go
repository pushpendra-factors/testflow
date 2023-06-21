package main

import (
	"context"
	C "factors/config"
	"factors/filestore"
	"factors/merge"
	"factors/model/store"
	"factors/pattern"
	"factors/pull"
	serviceDisk "factors/services/disk"
	serviceEtcd "factors/services/etcd"
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
	_ "github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

func registerStructs() {
	log.Info("Registering structs for beam")
	beam.RegisterType(reflect.TypeOf((*pattern.UserAndEventsInfo)(nil)).Elem())
	beam.RegisterType(reflect.TypeOf((*pattern.Pattern)(nil)).Elem())
	beam.RegisterType(reflect.TypeOf((*pattern.PropertiesInfo)(nil)).Elem())
	beam.RegisterType(reflect.TypeOf((*pattern.CounterEventFormat)(nil)).Elem())
	beam.RegisterType(reflect.TypeOf((*pattern.PropertiesCount)(nil)).Elem())
	beam.RegisterType(reflect.TypeOf((*C.Configuration)(nil)).Elem())
	beam.RegisterType(reflect.TypeOf((*pull.CounterCampaignFormat)(nil)).Elem())

	beam.RegisterType(reflect.TypeOf((*merge.RunBeamConfig)(nil)).Elem())
	beam.RegisterType(reflect.TypeOf((*merge.CUserIdsBeam)(nil)).Elem())
	beam.RegisterType(reflect.TypeOf((*merge.UidMap)(nil)).Elem())

	beam.RegisterType(reflect.TypeOf((*T.CpThreadDoFn)(nil)).Elem())
	beam.RegisterType(reflect.TypeOf((*T.UpThreadDoFn)(nil)).Elem())
	beam.RegisterType(reflect.TypeOf((*T.CPatternsBeam)(nil)).Elem())

	// do fn
	beam.RegisterType(reflect.TypeOf((*merge.SortUsDoFn)(nil)).Elem())
	beam.RegisterType(reflect.TypeOf((*merge.SortAdDoFn)(nil)).Elem())
}

func main() {

	envFlag := flag.String("env", "development", "")
	etcd := flag.String("etcd", "localhost:2379", "Comma separated list of etcd endpoints localhost:2379,localhost:2378")
	localDiskTmpDirFlag := flag.String("local_disk_tmp_dir", "/usr/local/var/factors/local_disk/tmp",
		"--local_disk_tmp_dir=/usr/local/var/factors/local_disk/tmp pass directory")
	tmpBucketNameFlag := flag.String("bucket_name_tmp", "/usr/local/var/factors/cloud_storage_tmp", "--bucket_name=/usr/local/var/factors/cloud_storage_tmp pass bucket name for tmp artifacts")
	archiveBucketNameFlag := flag.String("archive_bucket_name", "/usr/local/var/factors/cloud_storage_archive", "--bucket_name=/usr/local/var/factors/cloud_storage_archive pass archive bucket name")
	sortedBucketNameFlag := flag.String("sorted_bucket_name", "/usr/local/var/factors/cloud_storage_sorted", "--bucket_name=/usr/local/var/factors/cloud_storage_sorted pass sorted data bucket name")
	modelBucketNameFlag := flag.String("model_bucket_name", "/usr/local/var/factors/cloud_storage_models", "--bucket_name=/usr/local/var/factors/cloud_storage_models pass model bucket name")

	numRoutinesFlag := flag.Int("num_routines", 3, "No of routines")
	numWorkersFlag := flag.Int("num_beam_workers", 100, "Num of beam workers")

	projectIdFlag := flag.String("project_ids", "",
		"Optional: Project Id. A comma separated list of project Ids and supports '*' for all projects. ex: 1,2,6,9")
	projectIdsToSkipFlag := flag.String("project_ids_to_skip", "", "Optional: Comma separated values of projects to skip")
	maxModelSizeFlag := flag.Int64("max_size", 10000000000, "Max size of the model")
	shouldCountOccurence := flag.Bool("count_occurence", false, "")
	isWeeklyEnabled := flag.Bool("weekly_enabled", false, "")
	isMonthlyEnabled := flag.Bool("monthly_enabled", false, "")
	isQuarterlyEnabled := flag.Bool("quarterly_enabled", false, "")
	numActiveFactorsGoalsLimit := flag.Int("goals_limit", 50, "Max number of goals model")
	numActiveFactorsTrackedEventsLimit := flag.Int("max_tracked_events", 50, "Max number of Tracked Events")
	numActiveFactorsTrackedUserPropertiesLimit := flag.Int("max_user_properties", 50, "Max numbr of Tracked user properties")
	numCampaignsLimit := flag.Int("max_campaigns_limit", -1, "Max number of campaigns")
	runBeam := flag.Int("run_beam", 1, "run build seq on beam ")
	countsVersion := flag.Int("count_version", 1, "run fp tree code")
	hmineSupport := flag.Float64("hmine_support", 0.010, "value for hmine support")
	hmine_persist := flag.Int("hmine_persist", 0, "persist properties file while counting")
	start_event_v2 := flag.String("start_event", "", "start event for explain v2 job")
	end_event_v2 := flag.String("end_event", "", "end event for explain v2 job")
	include_events_v2 := flag.String("included_events", "", "Optional comma seperated values")

	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	isPSCHost := flag.Int("memsql_is_psc_host", C.MemSQLDefaultDBParams.IsPSCHost, "")
	memSQLPort := flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser := flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName := flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass := flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	memSQLCertificate := flag.String("memsql_cert", "", "")
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypeMemSQL, "Primary datastore type as memsql or postgres")

	sentryDSN := flag.String("sentry_dsn", "", "Sentry DSN")
	redisHost := flag.String("redis_host", "localhost", "")
	redisPort := flag.Int("redis_port", 6379, "")
	redisHostPersistent := flag.String("redis_host_ps", "localhost", "")
	redisPortPersistent := flag.Int("redis_port_ps", 6379, "")

	hardPull := flag.Bool("hard_pull", false, "replace the files already present")
	useSortedFilesMerge := flag.Bool("use_sorted_merge", false, "whether to use sorted files(if possible) or archive files")
	lookback := flag.Int("lookback", 30, "lookback_for_delta lookup")
	createMetadata := flag.Bool("create_metadata", false, "")
	flag.Parse()

	defer util.NotifyOnPanic("Task#PatternMine", *envFlag)

	if *envFlag != "development" &&
		*envFlag != "staging" &&
		*envFlag != "production" {
		err := fmt.Errorf("env [ %s ] not recognised", *envFlag)
		panic(err)
	}

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

	// init DB, etcd
	appName := "pattern_mine_job"
	healthcheckPingID := C.HealthcheckPatternMinePingID
	defer C.PingHealthcheckForPanic(appName, *envFlag, healthcheckPingID)

	config := &C.Configuration{
		AppName:       appName,
		Env:           *envFlag,
		EtcdEndpoints: strings.Split(*etcd, ","),
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
		SentryDSN:           *sentryDSN,
		PrimaryDatastore:    *primaryDatastore,
		RedisHost:           *redisHost,
		RedisPort:           *redisPort,
		RedisHostPersistent: *redisHostPersistent,
		RedisPortPersistent: *redisPortPersistent,
	}

	C.InitConf(config)
	beamConfig.DriverConfig = config
	C.InitSentryLogging(config.SentryDSN, config.AppName)

	// db is used by M.GetEventNames to build eventInfo.
	err := C.InitDB(*config)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize DB")
	}
	db := C.GetServices().Db
	defer db.Close()

	err = C.InitEtcd(config.EtcdEndpoints)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize etcd")
	}
	etcdClient, err := serviceEtcd.New([]string{*etcd})
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Fatal("Failed to init etcd client")
	}

	C.InitRedis(config.RedisHost, config.RedisPort)
	C.InitRedisPersistent(config.RedisHostPersistent, config.RedisPortPersistent)
	log.WithFields(log.Fields{
		"Env":             *envFlag,
		"EtcdEndpoints":   *etcd,
		"localDiskTmpDir": *localDiskTmpDirFlag,
		"Bucket":          *modelBucketNameFlag,
		"NumRoutines":     *numRoutinesFlag,
	}).Infoln("Initialising")

	if *numRoutinesFlag < 1 {
		log.Fatal("num_routines is less than one.")
	}

	projectIdsToSkip := util.GetIntBoolMapFromStringList(projectIdsToSkipFlag)
	allProjects, projectIdsToRun, _ := C.GetProjectsFromListWithAllProjectSupport(*projectIdFlag, "")
	if allProjects {
		projectIDs, errCode := store.GetStore().GetAllProjectIDs()
		if errCode != http.StatusFound {
			log.Fatal("Failed to get all projects and project_ids set to '*'.")
		}

		projectIdsToRun = make(map[int64]bool, 0)
		for _, projectID := range projectIDs {
			projectIdsToRun[projectID] = true
		}
	}

	projectIdsArray := make([]int64, 0)
	for projectId, _ := range projectIdsToRun {
		projectIdsArray = append(projectIdsArray, projectId)
	}

	C.GetConfig().ActiveFactorsGoalsLimit = *numActiveFactorsGoalsLimit
	C.GetConfig().ActiveFactorsTrackedEventsLimit = *numActiveFactorsTrackedEventsLimit
	C.GetConfig().ActiveFactorsTrackedUserPropertiesLimit = *numActiveFactorsTrackedUserPropertiesLimit
	log.Info("config :", config)
	diskManager := serviceDisk.New(*localDiskTmpDirFlag)

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
			log.WithField("error", err).Fatal("Failed to init cloud manager.")
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

	configs["env"] = *envFlag
	configs["db"] = db
	configs["etcdClient"] = etcdClient
	configs["diskManager"] = diskManager
	configs["noOfPatternWorkers"] = *numRoutinesFlag
	configs["projectIdsToSkip"] = projectIdsToSkip
	configs["maxModelSize"] = *maxModelSizeFlag
	configs["countOccurence"] = *shouldCountOccurence
	configs["numCampaignsLimit"] = *numCampaignsLimit
	configs["beamConfig"] = &beamConfig
	configs["countsVersion"] = *countsVersion
	configs["create_metadata"] = *createMetadata
	configs["hmineSupport"] = float32(*hmineSupport)
	configs["hminePersist"] = *hmine_persist
	configs["hardPull"] = *hardPull
	configs["start_event"] = *start_event_v2
	configs["end_event"] = *end_event_v2
	configs["included_events"] = *include_events_v2
	configs["useSortedFilesMerge"] = *useSortedFilesMerge

	log.Infof("configs :%v", configs)
	// profiling

	// This job has dependency on pull_events
	if *isWeeklyEnabled {
		var taskName string = "PatternMineWeeklyV2"
		C.PingHealthcheckForStart(healthcheckPingID)
		configs["modelType"] = T.ModelTypeWeek
		status := taskWrapper.TaskFuncWithProjectId(taskName, *lookback, projectIdsArray, T.BuildSequential, configs)
		log.Info(status)
		var isSuccess bool = true
		for reason, message := range status {
			if message == false {
				C.PingHealthcheckForFailure(healthcheckPingID, reason+": pattern mine run failure")
				isSuccess = false
				break
			}
		}
		if isSuccess {
			C.PingHealthcheckForSuccess(healthcheckPingID, "Pattern Mine Weekly run success.")
		}
	}

	if *isMonthlyEnabled {
		var taskName string = "PatternMineMonthlyV2"
		C.PingHealthcheckForStart(healthcheckPingID)
		configs["modelType"] = T.ModelTypeMonth
		status := taskWrapper.TaskFuncWithProjectId(taskName, *lookback, projectIdsArray, T.BuildSequential, configs)
		log.Info(status)
		var isSuccess bool = true
		for reason, message := range status {
			if message == false {
				C.PingHealthcheckForFailure(healthcheckPingID, reason+": pattern mine run failure")
				isSuccess = false
				break
			}
		}
		if isSuccess {
			C.PingHealthcheckForSuccess(healthcheckPingID, "Pattern Mine Monthly run success.")
		}
	}

	if *isQuarterlyEnabled {
		var taskName string = "PatternMineQuarterlyV2"
		C.PingHealthcheckForStart(healthcheckPingID)
		configs["modelType"] = T.ModelTypeQuarter
		status := taskWrapper.TaskFuncWithProjectId(taskName, *lookback, projectIdsArray, T.BuildSequential, configs)
		log.Info(status)
		var isSuccess bool = true
		for reason, message := range status {
			if message == false {
				C.PingHealthcheckForFailure(healthcheckPingID, reason+": pattern mine run failure")
				isSuccess = false
				break
			}
		}
		if isSuccess {
			C.PingHealthcheckForSuccess(healthcheckPingID, "Pattern Mine Quarterly run success.")
		}
	}
}
