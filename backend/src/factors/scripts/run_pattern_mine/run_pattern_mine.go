package main

import (
	"context"
	C "factors/config"
	"factors/filestore"
	"factors/model/store"
	"factors/pattern"
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

	beam.RegisterType(reflect.TypeOf((*T.CpThreadDoFn)(nil)).Elem())
	beam.RegisterType(reflect.TypeOf((*T.RunBeamConfig)(nil)).Elem())
	beam.RegisterType(reflect.TypeOf((*T.CPatternsBeam)(nil)).Elem())

}

func main() {

	envFlag := flag.String("env", "development", "")
	etcd := flag.String("etcd", "localhost:2379", "Comma separated list of etcd endpoints localhost:2379,localhost:2378")
	localDiskTmpDirFlag := flag.String("local_disk_tmp_dir", "/usr/local/var/factors/local_disk/tmp",
		"--local_disk_tmp_dir=/usr/local/var/factors/local_disk/tmp pass directory")
	bucketName := flag.String("bucket_name", "/usr/local/var/factors/cloud_storage", "")
	numRoutinesFlag := flag.Int("num_routines", 3, "No of routines")
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
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypePostgres, "Primary datastore type as memsql or postgres")

	redisHost := flag.String("redis_host", "localhost", "")
	redisPort := flag.Int("redis_port", 6379, "")
	redisHostPersistent := flag.String("redis_host_ps", "localhost", "")
	redisPortPersistent := flag.Int("redis_port_ps", 6379, "")

	lookback := flag.Int("lookback", 30, "lookback_for_delta lookup")
	createMetadata := flag.Bool("create_metadata",false,"")
	flag.Parse()

	defer util.NotifyOnPanic("Task#PatternMine", *envFlag)

	if *envFlag != "development" &&
		*envFlag != "staging" &&
		*envFlag != "production" {
		err := fmt.Errorf("env [ %s ] not recognised", *envFlag)
		panic(err)
	}

	//init beam
	var beamConfig T.RunBeamConfig
	if *runBeam == 1 {
		log.Info("Initializing all beam constructs")
		registerStructs()
		beam.Init()
		beamConfig.RunOnBeam = true
		beamConfig.Env = *envFlag
		beamConfig.Ctx = context.Background()
		beamConfig.Pipe = beam.NewPipeline()
		beamConfig.Scp = beamConfig.Pipe.Root()
		if beam.Initialized() == true {
			log.Info("Initalized all Beam Inits")
		} else {
			log.Fatal("unable to initialize runners")

		}
	} else {
		beamConfig.RunOnBeam = false
	}

	// init DB, etcd
	appName := "pattern_mine_job"
	config := &C.Configuration{
		AppName:       appName,
		Env:           *envFlag,
		EtcdEndpoints: strings.Split(*etcd, ","),
		DBInfo: C.DBConf{
			Host:     *dbHost,
			Port:     *dbPort,
			User:     *dbUser,
			Name:     *dbName,
			Password: *dbPass,
			AppName:  appName,
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
		RedisHost:           *redisHost,
		RedisPort:           *redisPort,
		RedisHostPersistent: *redisHostPersistent,
		RedisPortPersistent: *redisPortPersistent,
	}

	C.InitConf(config)
	beamConfig.DriverConfig = config

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
		"Bucket":          *bucketName,
		"NumRoutines":     *numRoutinesFlag,
	}).Infoln("Initialising")

	if *numRoutinesFlag < 1 {
		log.Fatal("num_routines is less than one.")
	}

	var cloudManager filestore.FileManager
	if *envFlag == "development" {
		cloudManager = serviceDisk.New(*bucketName)
	} else {
		cloudManager, err = serviceGCS.New(*bucketName)
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

		projectIdsToRun = make(map[uint64]bool, 0)
		for _, projectID := range projectIDs {
			projectIdsToRun[projectID] = true
		}
	}

	projectIdsArray := make([]uint64, 0)
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
	configs["bucketName"] = *bucketName
	configs["noOfPatternWorkers"] = *numRoutinesFlag
	configs["projectIdsToSkip"] = projectIdsToSkip
	configs["maxModelSize"] = *maxModelSizeFlag
	configs["countOccurence"] = *shouldCountOccurence
	configs["numCampaignsLimit"] = *numCampaignsLimit
	configs["beamConfig"] = &beamConfig
	configs["countsVersion"] = *countsVersion
	configs["create_metadata"] = *createMetadata

	// profiling

	// This job has dependency on pull_events
	if *isWeeklyEnabled {
		configs["modelType"] = T.ModelTypeWeek
		status := taskWrapper.TaskFuncWithProjectId("PatternMineWeekly", *lookback, projectIdsArray, T.BuildSequential, configs)
		log.Info(status)
	}

	if *isMonthlyEnabled {
		configs["modelType"] = T.ModelTypeMonth
		status := taskWrapper.TaskFuncWithProjectId("PatternMineMonthly", *lookback, projectIdsArray, T.BuildSequential, configs)
		log.Info(status)
	}

	if *isQuarterlyEnabled {
		configs["modelType"] = T.ModelTypeQuarter
		status := taskWrapper.TaskFuncWithProjectId("PatternMineQuarterly", *lookback, projectIdsArray, T.BuildSequential, configs)
		log.Info(status)
	}

}
