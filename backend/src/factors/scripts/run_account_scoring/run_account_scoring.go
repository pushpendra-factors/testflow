package main

import (
	"context"
	C "factors/config"
	"factors/filestore"
	"factors/merge"
	M "factors/model/model"
	"factors/model/store"
	"factors/pattern"
	serviceDisk "factors/services/disk"
	serviceEtcd "factors/services/etcd"
	serviceGCS "factors/services/gcstorage"
	T "factors/task"
	AS "factors/task/account_scoring"

	"factors/util"
	"flag"
	"fmt"
	"reflect"
	"strings"
	"time"

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
	bucketName := flag.String("bucket_name", "/usr/local/var/factors/cloud_storage", "")
	numRoutinesFlag := flag.Int("num_routines", 3, "No of routines")
	numWorkersFlag := flag.Int("num_beam_workers", 100, "Num of beam workers")
	DayTimestamp := flag.Int64("day_time_stamp", time.Now().Unix(), "time stamp for day")

	tmpBucketNameFlag := flag.String("bucket_name_tmp", "/usr/local/var/factors/cloud_storage_tmp", "--bucket_name=/usr/local/var/factors/cloud_storage_tmp pass bucket name for tmp artifacts")
	archiveBucketNameFlag := flag.String("archive_bucket_name", "/usr/local/var/factors/cloud_storage_archive", "--bucket_name=/usr/local/var/factors/cloud_storage_archive pass archive bucket name")
	sortedBucketNameFlag := flag.String("sorted_bucket_name", "/usr/local/var/factors/cloud_storage_sorted", "--bucket_name=/usr/local/var/factors/cloud_storage_sorted pass sorted bucket name")
	modelBucketNameFlag := flag.String("model_bucket_name", "/usr/local/var/factors/cloud_storage_models", "--bucket_name=/usr/local/var/factors/cloud_storage_models pass models bucket name")

	projectIdFlag := flag.String("project_ids", "",
		"Optional: Project Id. A comma separated list of project Ids and supports '*' for all projects. ex: 1,2,6,9")

	runBeam := flag.Int("run_beam", 1, "run build seq on beam ")
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
	enableFeatureGatesV2 := flag.Bool("enable_feature_gates_v2", false, "")

	lookback := flag.Int("lookback", 30, "lookback_for_delta lookup")
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
	appName := "acc_score_job"
	healthcheckPingID := C.HealthcheckAccScoringJobPingID
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
		SentryDSN:            *sentryDSN,
		PrimaryDatastore:     *primaryDatastore,
		RedisHost:            *redisHost,
		RedisPort:            *redisPort,
		RedisHostPersistent:  *redisHostPersistent,
		RedisPortPersistent:  *redisPortPersistent,
		EnableFeatureGatesV2: *enableFeatureGatesV2,
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
		"Bucket":          *bucketName,
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

	projectIdList := *projectIdFlag
	projectIdsArray := make([]int64, 0)

	if projectIdList == "*" {
		projectIdsArray, err = store.GetStore().GetAllProjectsWithFeatureEnabled(M.FEATURE_ACCOUNT_SCORING)
		if err != nil {
			errString := fmt.Errorf("failed to get feature status for all projects")
			log.WithError(err).Error(errString)
		}
	} else {
		projectIds := C.GetTokensFromStringListAsUint64(projectIdList)
		for _, projectId := range projectIds {
			available := false
			available, err = store.GetStore().GetFeatureStatusForProjectV2(projectId, M.FEATURE_ACCOUNT_SCORING)
			if err != nil {
				log.WithFields(log.Fields{"projectID": projectId}).WithError(err).Error("Failed to get feature status in account scoring job for project")
				continue
			}
			if available {
				projectIdsArray = append(projectIdsArray, projectId)
			}
		}
	}

	diskManager := serviceDisk.New(*localDiskTmpDirFlag)

	configs := make(map[string]interface{})
	configs["env"] = *envFlag
	configs["db"] = db
	configs["cloudManager"] = &cloudManager
	configs["etcdClient"] = etcdClient
	configs["diskManger"] = diskManager
	configs["bucketName"] = *bucketName
	configs["noOfPatternWorkers"] = *numRoutinesFlag
	configs["beamConfig"] = &beamConfig
	configs["tmpCloudManager"] = &cloudManager
	configs["day_timestamp"] = *DayTimestamp
	configs["lookback"] = *lookback

	useBucketV2 := true
	if useBucketV2 {
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
	} else {
		var cloudManager filestore.FileManager
		if *envFlag == "development" {
			cloudManager = serviceDisk.New(*bucketName)
		} else {
			cloudManager, err = serviceGCS.New(*bucketName)
			if err != nil {
				log.WithField("error", err).Fatal("Failed to init cloud manager.")
			}
		}
		configs["modelCloudManager"] = &cloudManager
		configs["archiveCloudManager"] = &cloudManager
		configs["sortedCloudManager"] = &cloudManager
	}

	configs["diskManager"] = diskManager
	configs["beamConfig"] = &beamConfig

	log.WithField("projects", projectIdsArray).Info("Running acc scoring for these projects")

	for _, projectId := range projectIdsArray {

		status, _ := AS.BuildAccScoringDaily(projectId, configs)
		status["project id"] = projectId
		log.Info(status)
		if status["err"] != nil {
			C.PingHealthcheckForFailure(healthcheckPingID, status)
		}
		C.PingHealthcheckForSuccess(healthcheckPingID, status)
	}
}
