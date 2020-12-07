package main

// Mine TOP_K Frequent patterns for every event combination (segment) at every iteration.

// Sample usage in terminal.
// export GOPATH=/Users/aravindmurthy/code/factors/backend/
// go run run_pattern_mine.go --env=development --etcd=localhost:2379 --disk_dir=/usr/local/var/factors/local_disk --s3_region=us-east-1 --s3=/usr/local/var/factors/cloud_storage --num_routines=3 --project_id=<projectId> --model_id=<modelId>
// or
// go run run_pattern_mine.go --project_id=<projectId> --model_id=<modelId>
// default of count occurence is False
// go run run_pattern_mine.go --project_id=<projectId> --model_id=<modelId> --count_occurence=true/false

import (
	C "factors/config"
	"factors/filestore"
	serviceDisk "factors/services/disk"
	serviceEtcd "factors/services/etcd"
	serviceGCS "factors/services/gcstorage"
	T "factors/task"
	"factors/util"
	"flag"
	"fmt"
	"strings"

	_ "github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

func main() {
	projectIdFlag := flag.Uint64("project_id", 0, "Project Id.")
	modelIdFlag := flag.Uint64("model_id", 0, "Model Id")

	envFlag := flag.String("env", "development", "")
	etcd := flag.String("etcd", "localhost:2379",
		"Comma separated list of etcd endpoints localhost:2379,localhost:2378")
	localDiskTmpDirFlag := flag.String("local_disk_tmp_dir", "/usr/local/var/factors/local_disk/tmp", "--local_disk_tmp_dir=/usr/local/var/factors/local_disk/tmp pass directory")
	bucketName := flag.String("bucket_name", "/usr/local/var/factors/cloud_storage", "")
	numRoutinesFlag := flag.Int("num_routines", 3, "No of routines")
	maxModelSizeFlag := flag.Int64("max_size", 10000000000, "Max size of the model")
	shouldCountOccurence := flag.Bool("count_occurence", false, "")
	numActiveFactorsGoalsLimit := flag.Int("goals_limit", 50, "Max number of goals model")
	numActiveFactorsTrackedEventsLimit := flag.Int("max_tracked_events", 50, "Max number of Tracked Events")
	numActiveFactorsTrackedUserPropertiesLimit := flag.Int("max_user_properties", 50, "Max numbr of Tracked user properties")

	dbHost := flag.String("db_host", "localhost", "")
	dbPort := flag.Int("db_port", 5432, "")
	dbUser := flag.String("db_user", "autometa", "")
	dbName := flag.String("db_name", "autometa", "")
	dbPass := flag.String("db_pass", "@ut0me7a", "")

	flag.Parse()

	if *envFlag != "development" &&
		*envFlag != "staging" &&
		*envFlag != "production" {
		err := fmt.Errorf("env [ %s ] not recognised", *envFlag)
		panic(err)
	}

	defer util.NotifyOnPanic("Task#PatternMine", *envFlag)

	// init DB, etcd
	config := &C.Configuration{
		AppName:       "pattern_mine_job",
		Env:           *envFlag,
		EtcdEndpoints: strings.Split(*etcd, ","),
		DBInfo: C.DBConf{
			Host:     *dbHost,
			Port:     *dbPort,
			User:     *dbUser,
			Name:     *dbName,
			Password: *dbPass,
		},
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

	log.WithFields(log.Fields{
		"Env":             *envFlag,
		"EtcdEndpoints":   *etcd,
		"localDiskTmpDir": *localDiskTmpDirFlag,
		"ProjectId":       *projectIdFlag,
		"ModelId":         *modelIdFlag,
		"Bucket":          *bucketName,
		"NumRoutines":     *numRoutinesFlag,
	}).Infoln("Initialising")

	if *projectIdFlag <= 0 || *modelIdFlag <= 0 {
		log.Fatal("project_id and model_id are required.")
	}

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

	diskManager := serviceDisk.New(*localDiskTmpDirFlag)
	C.GetConfig().ActiveFactorsGoalsLimit = *numActiveFactorsGoalsLimit
	C.GetConfig().ActiveFactorsTrackedEventsLimit = *numActiveFactorsTrackedEventsLimit
	C.GetConfig().ActiveFactorsTrackedUserPropertiesLimit = *numActiveFactorsTrackedUserPropertiesLimit

	// modelType, startTime, endTime is part of update meta.
	// kept null on run script.
	_, _, err = T.PatternMine(db, etcdClient, &cloudManager, diskManager,
		*bucketName, *numRoutinesFlag, *projectIdFlag, *modelIdFlag, "", 0, 0, *maxModelSizeFlag, *shouldCountOccurence)
	if err != nil {
		log.WithError(err).Fatal("Pattern mining failed")
	}
}
