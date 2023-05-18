package main

import (
	"context"
	C "factors/config"
	"factors/filestore"
	"factors/merge"
	"factors/model/store"
	"factors/pattern"
	serviceDisk "factors/services/disk"
	serviceEtcd "factors/services/etcd"
	serviceGCS "factors/services/gcstorage"
	T "factors/task"
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
	numRoutinesFlag := flag.Int("num_routines", 3, "No of routines")
	numWorkersFlag := flag.Int("num_beam_workers", 100, "Num of beam workers")

	tmpBucketNameFlag := flag.String("bucket_name_tmp", "/usr/local/var/factors/cloud_storage_tmp", "--bucket_name=/usr/local/var/factors/cloud_storage_tmp pass bucket name for tmp artifacts")
	archiveBucketNameFlag := flag.String("archive_bucket_name", "/usr/local/var/factors/cloud_storage_archive", "--bucket_name=/usr/local/var/factors/cloud_storage_archive pass archive bucket name")
	sortedBucketNameFlag := flag.String("sorted_bucket_name", "/usr/local/var/factors/cloud_storage_sorted", "--bucket_name=/usr/local/var/factors/cloud_storage_sorted pass sorted bucket name")
	modelBucketNameFlag := flag.String("model_bucket_name", "/usr/local/var/factors/cloud_storage_models", "--bucket_name=/usr/local/var/factors/cloud_storage_models pass models bucket name")
	hardPull := flag.Bool("hard_pull", false, "replace the files already present")
	useSortedFilesMerge := flag.Bool("use_sorted_merge", false, "whether to use sorted files(if possible) or archive files")

	projectIdFlag := flag.String("project_ids", "",
		"Optional: Project Id. A comma separated list of project Ids and supports '*' for all projects. ex: 1,2,6,9")
	projectIdsToSkipFlag := flag.String("project_ids_to_skip", "", "Optional: Comma separated values of projects to skip")
	maxModelSizeFlag := flag.Int64("max_size", 10000000000, "Max size of the model")
	shouldCountOccurence := flag.Bool("count_occurence", false, "")
	numActiveFactorsGoalsLimit := flag.Int("goals_limit", 50, "Max number of goals model")
	numActiveFactorsTrackedEventsLimit := flag.Int("max_tracked_events", 50, "Max number of Tracked Events")
	numActiveFactorsTrackedUserPropertiesLimit := flag.Int("max_user_properties", 50, "Max numbr of Tracked user properties")
	numCampaignsLimit := flag.Int("max_campaigns_limit", -1, "Max number of campaigns")
	runBeam := flag.Int("run_beam", 1, "run build seq on beam ")
	countsVersion := flag.Int("count_version", 1, "run fp tree code")
	hmineSupport := flag.Float64("hmine_support", 0.010, "value for hmine support")
	hmine_persist := flag.Int("hmine_persist", 0, "persist properties file while counting")
	// isWeeklyEnabled := flag.Bool("weekly_enabled", false, "")
	// isMonthlyEnabled := flag.Bool("monthly_enabled", false, "")
	// isQuarterlyEnabled := flag.Bool("quarterly_enabled", false, "")

	// dbHost := flag.String("db_host", C.PostgresDefaultDBParams.Host, "")
	// dbPort := flag.Int("db_port", C.PostgresDefaultDBParams.Port, "")
	// dbUser := flag.String("db_user", C.PostgresDefaultDBParams.User, "")
	// dbName := flag.String("db_name", C.PostgresDefaultDBParams.Name, "")
	// dbPass := flag.String("db_pass", C.PostgresDefaultDBParams.Password, "")

	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
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
	appName := "explain_v2_job"
	healthcheckPingID := C.HealthcheckPatternMinePingID
	defer C.PingHealthcheckForPanic(appName, *envFlag, healthcheckPingID)

	config := &C.Configuration{
		AppName:       appName,
		Env:           *envFlag,
		EtcdEndpoints: strings.Split(*etcd, ","),
		MemSQLInfo: C.DBConf{
			Host:        *memSQLHost,
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

	var cloudManager filestore.FileManager
	if *envFlag == "development" {
		cloudManager = serviceDisk.New(*tmpBucketNameFlag)
	} else {
		cloudManager, err = serviceGCS.New(*tmpBucketNameFlag)
		if err != nil {
			log.WithError(err).Errorln("Failed to init New GCS Client")
			panic(err)
		}
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
	configs["env"] = *envFlag
	configs["db"] = db
	configs["cloudManager"] = &cloudManager
	configs["etcdClient"] = etcdClient
	configs["diskManger"] = diskManager
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
	configs["modelType"] = "w"
	configs["tmpCloudManager"] = &cloudManager

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

	configs["diskManager"] = diskManager
	configs["beamConfig"] = &beamConfig
	configs["hardPull"] = *hardPull
	configs["useSortedFilesMerge"] = *useSortedFilesMerge

	log.Infof("configs :%v", configs)

	// // This job has dependency on pull_events
	// C.PingHealthcheckForStart(healthcheckPingID)
	// configs["modelType"] = T.ModelTypeWeek
	// status := taskWrapper.TaskFuncWithProjectId("ExplainV2Job", *lookback, projectIdsArray, T.BuildSequentialV2, configs)
	// log.Info(status)
	// var isSuccess bool = true
	// for reason, message := range status {
	// 	if message == false {
	// 		C.PingHealthcheckForFailure(healthcheckPingID, reason+": pattern mine run failure")
	// 		isSuccess = false
	// 		break
	// 	}
	// }
	// if isSuccess {
	// 	C.PingHealthcheckForSuccess(healthcheckPingID, "Pattern Mine  success.")
	// }

	var result bool
	finalStatus := make(map[string]interface{})
	for _, projectId := range projectIdsArray {
		status := make(map[string]interface{})
		status, result := T.BuildSequentialV2(projectId, configs)
		if result == false {
			finalStatus["err"+fmt.Sprintf("%v", projectId)] = status
			break
		}
		finalStatus[fmt.Sprintf("%v", projectId)] = status
	}
	if result == false {
		C.PingHealthcheckForFailure(healthcheckPingID, finalStatus)
	} else {
		C.PingHealthcheckForSuccess(healthcheckPingID, finalStatus)
	}
}
