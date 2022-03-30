package main

import (
	"flag"
	"fmt"
	"sync"
	"time"

	"github.com/jinzhu/gorm"

	C "factors/config"
	"factors/filestore"
	serviceDisk "factors/services/disk"
	serviceGCS "factors/services/gcstorage"
	T "factors/task"
	U "factors/util"

	log "github.com/sirupsen/logrus"
)

func main() {
	envFlag := flag.String("env", C.DEVELOPMENT, "Environment. Could be development|staging|production.")
	bucketNameFlag := flag.String("bucket_name", "/usr/local/var/factors/cloud_storage", "Bucket name for production.")
	localDiskTmpDirFlag := flag.String("tmp_dir", "/usr/local/var/factors/local_disk/tmp", "Local directory path for putting tmp files.")
	projectIDFlag := flag.String("project_id", "", "Comma separated list of project ids to run")
	numRoutinesFlag := flag.Int("num_routines", 2, "Number of projects to run in parallel")
	startDateFlag := flag.String("start_date", "", "Start date in YYYY-MM-DD format to run for specific period. Inclusive.")
	endDateFlag := flag.String("end_date", "", "End date in YYYY-MM-DD format to run for specific period. Inclusive.")

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
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypeMemSQL, "Primary datastore type as memsql or postgres")

	sentryDSN := flag.String("sentry_dsn", "", "Sentry DSN")

	flag.Parse()

	taskID := "Script#AdhocArchiveEvents"
	defer U.NotifyOnPanic(taskID, *envFlag)

	logCtx := log.WithField("prefix", taskID)
	if *envFlag != C.DEVELOPMENT && *envFlag != C.STAGING && *envFlag != C.PRODUCTION {
		panic(fmt.Errorf("env [ %s ] not recognised", *envFlag))
	} else if *projectIDFlag == "" {
		panic(fmt.Errorf("Invalid project id %s", *projectIDFlag))
	} else if *startDateFlag == "" || *endDateFlag == "" {
		panic(fmt.Errorf("Must provide both start_date and end_date"))
	} else if *numRoutinesFlag == 0 {
		*numRoutinesFlag = 1
	}

	logCtx.Info("Starting to initialize database.")
	config := &C.Configuration{
		AppName: taskID,
		Env:     *envFlag,
		DBInfo: C.DBConf{
			Host:     *dbHost,
			Port:     *dbPort,
			User:     *dbUser,
			Name:     *dbName,
			Password: *dbPass,
			AppName:  taskID,
		},
		MemSQLInfo: C.DBConf{
			Host:        *memSQLHost,
			Port:        *memSQLPort,
			User:        *memSQLUser,
			Name:        *memSQLName,
			Password:    *memSQLPass,
			Certificate: *memSQLCertificate,
			AppName:     taskID,
		},
		SentryDSN:        *sentryDSN,
		PrimaryDatastore: *primaryDatastore,
	}
	C.InitConf(config)

	err := C.InitDB(*config)
	if err != nil {
		logCtx.WithError(err).Fatal("Failed to initialize DB")
	}
	db := C.GetServices().Db
	defer db.Close()

	C.InitSentryLogging(config.SentryDSN, config.AppName)
	defer C.SafeFlushAllCollectors()

	var cloudManager filestore.FileManager
	if *envFlag == "development" {
		cloudManager = serviceDisk.New(*bucketNameFlag)
	} else {
		cloudManager, err = serviceGCS.New(*bucketNameFlag)
		if err != nil {
			logCtx.WithError(err).Errorln("Failed to init New GCS Client")
			panic(err)
		}
	}
	diskManager := serviceDisk.New(*localDiskTmpDirFlag)

	startTime, err := time.Parse(U.DATETIME_FORMAT_YYYYMMDD_HYPHEN, *startDateFlag)
	if err != nil {
		logCtx.WithError(err).Fatal("Invalid start_time. Format must be YYYY-MM-DD")
	}
	endTime, err := time.Parse(U.DATETIME_FORMAT_YYYYMMDD_HYPHEN, *endDateFlag)
	if err != nil {
		logCtx.WithError(err).Fatal("Invalid end_time. Format must be YYYY-MM-DD")
	}

	allProjects, projectIDMap, _ := C.GetProjectsFromListWithAllProjectSupport(*projectIDFlag, "")
	if allProjects {
		logCtx.Fatal("Running for all projects not supported")
	}

	routinesCount := 0
	var allJobDetails, allJobErrors sync.Map
	var waitGroup sync.WaitGroup

	for projectID := range projectIDMap {
		logCtx.Infof("Running for project id: %d", projectID)
		waitGroup.Add(1)
		routinesCount++
		go runArchiveAsRoutineForProjectID(db, &cloudManager, diskManager, projectID,
			startTime, endTime, &waitGroup, &allJobDetails, &allJobErrors)

		if routinesCount%*numRoutinesFlag == 0 {
			waitGroup.Wait()
		}
	}
	waitGroup.Wait()

	allJobDetailsMap := make(map[uint64]interface{})
	allJobDetails.Range(func(key, value interface{}) bool {
		allJobDetailsMap[key.(uint64)] = value
		return true
	})
	allJobErrors.Range(func(key, value interface{}) bool {
		allJobDetailsMap[key.(uint64)] = value
		return true
	})

	err = U.NotifyThroughSNS(taskID, *envFlag, allJobDetailsMap)
	if err != nil {
		logCtx.WithError(err).Error("SNS notification failed", allJobDetailsMap)
	}
}

func runArchiveAsRoutineForProjectID(db *gorm.DB, cloudManager *filestore.FileManager, diskManager *serviceDisk.DiskDriver,
	projectID uint64, startTime, endTime time.Time, waitGroup *sync.WaitGroup, allJobDetails, allJobErrors *sync.Map) {
	defer waitGroup.Done()

	jobDetails, err := T.ArchiveEventsForProject(db, cloudManager, diskManager, projectID, 0, startTime, endTime, true)
	if err != nil {
		allJobErrors.Store(projectID, err)
	}
	allJobDetails.Store(projectID, jobDetails)
}
