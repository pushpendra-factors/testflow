package main

import (
	"context"
	"factors/filestore"
	"factors/merge"
	"factors/model/store"
	"factors/pattern"
	"factors/pull"
	serviceDisk "factors/services/disk"
	serviceGCS "factors/services/gcstorage"
	"factors/util"
	"flag"
	"fmt"
	"net/http"
	"reflect"

	C "factors/config"

	T "factors/task"

	taskWrapper "factors/task/task_wrapper"

	"github.com/apache/beam/sdks/go/pkg/beam"
	_ "github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

func main() {
	mainRunNumericalBucketing()
}

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

func mainRunNumericalBucketing() {

	projectIdFlag := flag.String("project_ids", "",
		"Optional: Project Id. A comma separated list of project Ids and supports '*' for all projects. ex: 1,2,6,9")

	envFlag := flag.String("env", "development", "")
	localDiskTmpDirFlag := flag.String("local_disk_tmp_dir", "/usr/local/var/factors/local_disk/tmp", "--local_disk_tmp_dir=/usr/local/var/factors/local_disk/tmp pass directory")
	bucketName := flag.String("bucket_name", "/usr/local/var/factors/cloud_storage", "--bucket_name=/usr/local/var/factors/cloud_storage pass bucket name")
	tmpBucketNameFlag := flag.String("bucket_name_tmp", "/usr/local/var/factors/cloud_storage_tmp", "--bucket_name=/usr/local/var/factors/cloud_storage_tmp pass bucket name for tmp artifacts")
	archiveBucketNameFlag := flag.String("archive_bucket_name", "/usr/local/var/factors/cloud_storage_archive", "--bucket_name=/usr/local/var/factors/cloud_storage_archive pass archive bucket name")
	sortedBucketNameFlag := flag.String("sorted_bucket_name", "/usr/local/var/factors/cloud_storage_sorted", "--bucket_name=/usr/local/var/factors/cloud_storage_sorted pass sorted bucket name")
	useBucketV2 := flag.Bool("use_bucket_v2", false, "Whether to use new bucketing system or not")

	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	memSQLPort := flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser := flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName := flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass := flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	memSQLCertificate := flag.String("memsql_cert", "", "")
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypeMemSQL, "Primary datastore type as memsql or postgres")

	isWeeklyEnabled := flag.Bool("weekly_enabled", false, "")
	lookback := flag.Int("lookback", 30, "lookback_for_delta lookup")

	enablePropertyTypeFromDB := flag.Bool("enable_property_type_from_db", false, "Enable property type check from db.")
	whitelistedProjectIDPropertyTypeFromDB := flag.String("whitelisted_project_ids_property_type_check_from_db", "", "Allowed project id for property type check from db.")
	blacklistedProjectIDPropertyTypeFromDB := flag.String("blacklisted_project_ids_property_type_check_from_db", "", "Blocked project id for property type check from db.")
	propertiesTypeCacheSize := flag.Int("property_details_cache_size", 0, "Cache size for in memory property detail.")

	hardPull := flag.Bool("hard_pull", false, "replace the files already present")
	runBeam := flag.Int("run_beam", 1, "run build seq on beam ")
	numWorkersFlag := flag.Int("num_beam_workers", 100, "Num of beam workers")

	redisHost := flag.String("redis_host", "localhost", "")
	redisPort := flag.Int("redis_port", 6379, "")
	redisHostPersistent := flag.String("redis_host_ps", "localhost", "")
	redisPortPersistent := flag.Int("redis_port_ps", 6379, "")

	flag.Parse()

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

	defer util.NotifyOnPanic("Task#compute_numerical_bucketing", *envFlag)

	// init Config and DB.
	config := &C.Configuration{
		AppName: "compute_numerical_bucketing",
		Env:     *envFlag,
		MemSQLInfo: C.DBConf{
			Host:        *memSQLHost,
			Port:        *memSQLPort,
			User:        *memSQLUser,
			Name:        *memSQLName,
			Password:    *memSQLPass,
			Certificate: *memSQLCertificate,
			AppName:     "compute_numerical_bucketing",
		},
		PrimaryDatastore:    *primaryDatastore,
		RedisHost:           *redisHost,
		RedisPort:           *redisPort,
		RedisHostPersistent: *redisHostPersistent,
		RedisPortPersistent: *redisPortPersistent,
	}
	C.InitConf(config)
	beamConfig.DriverConfig = config
	err := C.InitDB(*config)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize DB")
	}

	C.InitRedis(config.RedisHost, config.RedisPort)
	C.InitRedisPersistent(config.RedisHostPersistent, config.RedisPortPersistent)
	log.WithFields(log.Fields{
		"Env":             *envFlag,
		"localDiskTmpDir": *localDiskTmpDirFlag,
		"Bucket":          *bucketName,
	}).Infoln("Initialising")

	C.InitPropertiesTypeCache(*enablePropertyTypeFromDB, *propertiesTypeCacheSize, *whitelistedProjectIDPropertyTypeFromDB, *blacklistedProjectIDPropertyTypeFromDB)

	diskManager := serviceDisk.New(*localDiskTmpDirFlag)

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
	if *useBucketV2 {
		var archiveCloudManager filestore.FileManager
		var sortedCloudManager filestore.FileManager
		if *envFlag == "development" {
			archiveCloudManager = serviceDisk.New(*archiveBucketNameFlag)
			sortedCloudManager = serviceDisk.New(*sortedBucketNameFlag)
		} else {
			archiveCloudManager, err = serviceGCS.New(*archiveBucketNameFlag)
			if err != nil {
				log.WithField("error", err).Fatal("Failed to init archive cloud manager")
			}
			sortedCloudManager, err = serviceGCS.New(*sortedBucketNameFlag)
			if err != nil {
				log.WithField("error", err).Fatal("Failed to init sorted data cloud manager")
			}
		}
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
		configs["archiveCloudManager"] = &cloudManager
		configs["sortedCloudManager"] = &cloudManager
	}

	configs["diskManager"] = diskManager
	configs["hardPull"] = *hardPull
	configs["beamConfig"] = &beamConfig
	configs["useBucketV2"] = *useBucketV2

	var taskName string
	if *useBucketV2 {
		taskName = "NumericalBucketingWeeklyV2"
	} else {
		taskName = "NumericalBucketingWeekly"
	}
	if *isWeeklyEnabled {
		configs["modelType"] = T.ModelTypeWeek
		status := taskWrapper.TaskFuncWithProjectId(taskName, *lookback, projectIdsArray, T.NumericalBucketing, configs)
		log.Info(status)
	}
}
