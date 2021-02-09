package main

import (
	C "factors/config"
	"factors/filestore"
	"factors/model/store"
	serviceDisk "factors/services/disk"
	serviceEtcd "factors/services/etcd"
	serviceGCS "factors/services/gcstorage"
	T "factors/task"
	"factors/util"
	"flag"
	"fmt"
	"net/http"
	"strings"

	_ "github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

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
	modelType := flag.String("model_type", "weekly", "Optional: Model Type can take 3 values : {all, weekly, monthly}")
	lookBackPeriodInDays := flag.Int64("look_back_days", 30,
		"Optional: Build projects which were build in last N days. Provide N here.")
	noOfDaysToBuild := flag.Int64("no_of_days", 0, "Optional: No.of days to build for. Defaults to current_timestamp.")
	shouldCountOccurence := flag.Bool("count_occurence", false, "")
	numActiveFactorsGoalsLimit := flag.Int("goals_limit", 50, "Max number of goals model")
	numActiveFactorsTrackedEventsLimit := flag.Int("max_tracked_events", 50, "Max number of Tracked Events")
	numActiveFactorsTrackedUserPropertiesLimit := flag.Int("max_user_properties", 50, "Max numbr of Tracked user properties")
	numCampaignsLimit := flag.Int("max_campaigns_limit", -1, "Max number of campaigns")

	dbHost := flag.String("db_host", "localhost", "")
	dbPort := flag.Int("db_port", 5432, "")
	dbUser := flag.String("db_user", "autometa", "")
	dbName := flag.String("db_name", "autometa", "")
	dbPass := flag.String("db_pass", "@ut0me7a", "")

	redisHost := flag.String("redis_host", "localhost", "")
	redisPort := flag.Int("redis_port", 6379, "")
	redisHostPersistent := flag.String("redis_host_ps", "localhost", "")
	redisPortPersistent := flag.Int("redis_port_ps", 6379, "")

	flag.Parse()

	defer util.NotifyOnPanic("Task#BuildSeq", *envFlag)

	if *envFlag != "development" &&
		*envFlag != "staging" &&
		*envFlag != "production" {
		err := fmt.Errorf("env [ %s ] not recognised", *envFlag)
		panic(err)
	}

	if *modelType != T.ModelTypeAll &&
		*modelType != T.ModelTypeWeekly &&
		*modelType != T.ModelTypeMonthly {
		err := fmt.Errorf("modelType [ %s ] not recognised", *modelType)
		panic(err)
	}

	// init DB, etcd
	config := &C.Configuration{
		AppName:       "build_seq_job",
		Env:           *envFlag,
		EtcdEndpoints: strings.Split(*etcd, ","),
		DBInfo: C.DBConf{
			Host:     *dbHost,
			Port:     *dbPort,
			User:     *dbUser,
			Name:     *dbName,
			Password: *dbPass,
		},
		RedisHost:           *redisHost,
		RedisPort:           *redisPort,
		RedisHostPersistent: *redisHostPersistent,
		RedisPortPersistent: *redisPortPersistent,
	}

	C.InitConf(config.Env)

	// db is used by M.GetEventNames to build eventInfo.
	err := C.InitDB(config.DBInfo)
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

	C.GetConfig().ActiveFactorsGoalsLimit = *numActiveFactorsGoalsLimit
	C.GetConfig().ActiveFactorsTrackedEventsLimit = *numActiveFactorsTrackedEventsLimit
	C.GetConfig().ActiveFactorsTrackedUserPropertiesLimit = *numActiveFactorsTrackedUserPropertiesLimit
	log.Info("config :", config)
	diskManager := serviceDisk.New(*localDiskTmpDirFlag)

	_ = T.BuildSequential(*envFlag, db, &cloudManager, etcdClient, diskManager,
		*bucketName, *numRoutinesFlag, projectIdsToRun, projectIdsToSkip, *maxModelSizeFlag,
		*modelType, *lookBackPeriodInDays, *noOfDaysToBuild, *shouldCountOccurence, *numCampaignsLimit)
}
