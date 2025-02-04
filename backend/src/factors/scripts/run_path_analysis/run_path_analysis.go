package main

import (
	"context"
	C "factors/config"
	"factors/merge"
	"factors/pattern"
	"factors/pull"
	U "factors/util"
	"flag"
	"fmt"
	"reflect"

	M "factors/model/model"
	T "factors/task"

	"factors/filestore"
	serviceDisk "factors/services/disk"
	serviceGCS "factors/services/gcstorage"

	"factors/model/store"

	"github.com/apache/beam/sdks/go/pkg/beam"
	log "github.com/sirupsen/logrus"
)

func registerStructs() {
	log.Info("Registering structs for beam")
	beam.RegisterType(reflect.TypeOf((*C.Configuration)(nil)).Elem())

	beam.RegisterType(reflect.TypeOf((*pattern.CounterEventFormat)(nil)).Elem())
	beam.RegisterType(reflect.TypeOf((*pull.CounterCampaignFormat)(nil)).Elem())

	beam.RegisterType(reflect.TypeOf((*merge.RunBeamConfig)(nil)).Elem())
	beam.RegisterType(reflect.TypeOf((*merge.CUserIdsBeam)(nil)).Elem())
	beam.RegisterType(reflect.TypeOf((*merge.UidMap)(nil)).Elem())

	// do fn
	beam.RegisterType(reflect.TypeOf((*merge.SortUsDoFn)(nil)).Elem())
	beam.RegisterType(reflect.TypeOf((*merge.SortAdDoFn)(nil)).Elem())
}

func main() {
	env := flag.String("env", C.DEVELOPMENT, "")
	tmpBucketNameFlag := flag.String("bucket_name_tmp", "/usr/local/var/factors/cloud_storage_tmp", "--bucket_name=/usr/local/var/factors/cloud_storage_tmp pass bucket name for tmp artifacts")
	archiveBucketNameFlag := flag.String("archive_bucket_name", "/usr/local/var/factors/cloud_storage_archive", "--bucket_name=/usr/local/var/factors/cloud_storage_archive pass archive bucket name")
	sortedBucketNameFlag := flag.String("sorted_bucket_name", "/usr/local/var/factors/cloud_storage_sorted", "--bucket_name=/usr/local/var/factors/cloud_storage_sorted pass sorted bucket name")
	modelBucketNameFlag := flag.String("model_bucket_name", "/usr/local/var/factors/cloud_storage_models", "--bucket_name=/usr/local/var/factors/cloud_storage_models pass models bucket name")
	hardPull := flag.Bool("hard_pull", false, "replace the files already present")
	useSortedFilesMerge := flag.Bool("use_sorted_merge", false, "whether to use sorted files(if possible) or archive files")
	localDiskTmpDirFlag := flag.String("local_disk_tmp_dir", "/usr/local/var/factors/local_disk/tmp", "--local_disk_tmp_dir=/usr/local/var/factors/local_disk/tmp pass directory.")

	numWorkersFlag := flag.Int("num_beam_workers", 100, "Num of beam workers")
	runBeam := flag.Int("run_beam", 1, "run merge sort on beam ")
	sortOnGroup := flag.Int("sort_group", 0, "sort events based on group (0 for uid)")

	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	isPSCHost := flag.Int("memsql_is_psc_host", C.MemSQLDefaultDBParams.IsPSCHost, "")
	memSQLPort := flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser := flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName := flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass := flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	memSQLCertificate := flag.String("memsql_cert", "", "")
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypeMemSQL, "Primary datastore type as memsql or postgres")

	overrideHealthcheckPingID := flag.String("healthcheck_ping_id", "", "Override default healthcheck ping id.")

	redisHost := flag.String("redis_host", "localhost", "")
	redisPort := flag.Int("redis_port", 6379, "")
	redisHostPersistent := flag.String("redis_host_ps", "localhost", "")
	redisPortPersistent := flag.Int("redis_port_ps", 6379, "")
	blacklistedQueries := flag.String("blacklisted_queries", "", "")

	flag.Parse()
	if *env != "development" &&
		*env != "staging" &&
		*env != "production" {
		err := fmt.Errorf("env [ %s ] not recognised", *env)
		panic(err)
	}

	//init beam
	var beamConfig merge.RunBeamConfig
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

	defer U.NotifyOnPanic("Script#run_wi_alerts", *env)
	appName := "run_wi_mailer"
	config := &C.Configuration{
		AppName: appName,
		Env:     *env,
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
		PrimaryDatastore:    *primaryDatastore,
		RedisHost:           *redisHost,
		RedisPort:           *redisPort,
		RedisHostPersistent: *redisHostPersistent,
		RedisPortPersistent: *redisPortPersistent,
	}
	defaultHealthcheckPingID := C.HealthcheckPathAnalysisPingID
	healthcheckPingID := C.GetHealthcheckPingID(defaultHealthcheckPingID, *overrideHealthcheckPingID)
	C.InitConf(config)
	beamConfig.DriverConfig = config
	err := C.InitDB(*config)
	if err != nil {
		log.Fatal("Init failed.")
	}
	C.InitRedis(config.RedisHost, config.RedisPort)
	C.InitRedisPersistent(config.RedisHostPersistent, config.RedisPortPersistent)
	db := C.GetServices().Db
	defer db.Close()
	//Initialized configs

	projectIDs, err := store.GetStore().GetAllProjectsWithFeatureEnabled(M.FEATURE_PATH_ANALYSIS, false)
	if err != nil {
		log.WithError(err).Fatal("failed to get project ids using feature gates")
	}

	// Get All the Projects for which the path analysis has pending items
	configs := make(map[string]interface{})
	// Init cloud manager.
	var cloudManagerTmp filestore.FileManager
	if *env == "development" {
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
	if *env == "development" {
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

	diskManager := serviceDisk.New(*localDiskTmpDirFlag)

	configs["diskManager"] = diskManager
	configs["beamConfig"] = &beamConfig
	configs["hardPull"] = *hardPull
	configs["sortOnGroup"] = *sortOnGroup
	configs["useSortedFilesMerge"] = *useSortedFilesMerge

	configs["blacklistedQueries"] = C.GetTokensFromStringListAsString(*blacklistedQueries)

	var result bool
	finalStatus := make(map[int64]interface{})
	for _, projectId := range projectIDs {
		status := make(map[string]interface{})
		status, result = T.PathAnalysis(projectId, configs)
		if len(status) > 0 {
			finalStatus[projectId] = status
		}
		if !result {
			break
		}
	}
	if !result {
		C.PingHealthcheckForFailure(healthcheckPingID, finalStatus)
	} else {
		C.PingHealthcheckForSuccess(healthcheckPingID, finalStatus)
	}

}
