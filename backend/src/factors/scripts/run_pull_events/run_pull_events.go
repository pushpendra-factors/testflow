package main

// Pull events that needs to be processed and write to file.
// Sample usage in terminal.
// export GOPATH=/Users/aravindmurthy/code/factors/backend/
// go run run_pull_events.go --project_id=1 --output_dir="" --end_time=""

import (
	C "factors/config"
	"factors/filestore"
	serviceDisk "factors/services/disk"
	serviceGCS "factors/services/gcstorage"
	T "factors/task"
	"factors/util"
	"flag"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	DayInSecs   = 24 * 60 * 60
	MonthInSecs = 31 * DayInSecs
	WeekInSecs  = 7 * DayInSecs
)

func main() {
	env := flag.String("env", C.DEVELOPMENT, "")
	bucketNameFlag := flag.String("bucket_name", "/usr/local/var/factors/cloud_storage", "--bucket_name=/usr/local/var/factors/cloud_storage pass bucket name")
	localDiskTmpDirFlag := flag.String("local_disk_tmp_dir", "/usr/local/var/factors/local_disk/tmp", "--local_disk_tmp_dir=/usr/local/var/factors/local_disk/tmp pass directory.")
	endTimeFlag := flag.Int64("end_time", time.Now().Unix(), "Pull events, interval end timestamp. defaults to current timestamp. Format is unix timestamp.")
	modelTypeFlag := flag.String("model_type", "monthly", "Type of model for which to pull events, can be weekly or monthly. defaults to monthly.")
	startTimeFlag := flag.Int64("start_time", 0, "Pull events, interval start timestamp. Format is unix timestamp.")

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
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypePostgres, "Primary datastore type as memsql or postgres")

	projectIdFlag := flag.Uint64("project_id", 0, "Project Id.")

	flag.Parse()

	if *env != "development" &&
		*env != "staging" &&
		*env != "production" {
		err := fmt.Errorf("env [ %s ] not recognised", *env)
		panic(err)
	}

	defer util.NotifyOnPanic("Task#PullEvents", *env)

	appName := "pull_events_job"
	config := &C.Configuration{
		AppName: appName,
		Env:     *env,
		DBInfo: C.DBConf{
			Host:     *dbHost,
			Port:     *dbPort,
			User:     *dbUser,
			Name:     *dbName,
			Password: *dbPass,
		},
		MemSQLInfo: C.DBConf{
			Host:     *memSQLHost,
			Port:     *memSQLPort,
			User:     *memSQLUser,
			Name:     *memSQLName,
			Password: *memSQLPass,
			AppName:  appName,
		},
		PrimaryDatastore: *primaryDatastore,
	}

	C.InitConf(config)

	// Initialize configs and connections and close with defer.
	err := C.InitDB(*config)
	if err != nil {
		log.Fatal("Failed to pull events. Init failed.")
	}
	db := C.GetServices().Db
	defer db.Close()

	if *projectIdFlag <= 0 {
		log.Fatal("Failed to pull events. Invalid project_id.")
	}

	var startTime int64
	if *startTimeFlag > 0 {
		// Give precedence to given start time.
		startTime = *startTimeFlag
	} else {
		// Calculate start time based on model type, if start time not given.
		if *modelTypeFlag == "weekly" {
			startTime = *endTimeFlag - WeekInSecs
		} else if *modelTypeFlag == "monthly" {
			startTime = *endTimeFlag - MonthInSecs
		} else {
			log.Fatal("Invalid model_type. Use weekly or monthly.")
		}
	}

	// Init cloud manager.
	var cloudManager filestore.FileManager
	if *env == "development" {
		cloudManager = serviceDisk.New(*bucketNameFlag)
	} else {
		cloudManager, err = serviceGCS.New(*bucketNameFlag)
		if err != nil {
			log.WithField("error", err).Fatal("Failed to init cloud manager.")
		}
	}

	diskManager := serviceDisk.New(*localDiskTmpDirFlag)

	_, _, err = T.PullEvents(db, &cloudManager, diskManager, *projectIdFlag, startTime, *endTimeFlag)
	if err != nil {
		log.WithError(err).Fatal("Failed to pull events.")
	}
}
