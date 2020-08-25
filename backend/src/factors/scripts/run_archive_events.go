package main

import (
	"flag"
	"fmt"
	"time"

	C "factors/config"
	"factors/filestore"
	serviceDisk "factors/services/disk"
	serviceGCS "factors/services/gcstorage"
	T "factors/task"
	"factors/util"

	log "github.com/sirupsen/logrus"
)

var taskID = "Script#EventsArchival"
var pbLog = log.WithField("prefix", taskID)

func main() {
	envFlag := flag.String("env", "development", "Environment. Could be development|staging|production.")
	bucketNameFlag := flag.String("bucket_name", "/usr/local/var/factors/cloud_storage", "Bucket name for production.")
	localDiskTmpDirFlag := flag.String("tmp_dir", "/usr/local/var/factors/local_disk/tmp", "Local directory path for putting tmp files.")
	projectIDFlag := flag.Uint64("project_id", 0, "Project id to be run for.")
	maxLookbackDaysFlag := flag.Int("max_lookback_days", 365, "Maximum number of lookback days for events.")
	startDateFlag := flag.String("start_date", "", "Start date in YYYY-MM-DD format to run for specific period. Inclusive.")
	endDateFlag := flag.String("end_date", "", "End date in YYYY-MM-DD format to run for specific period. Inclusive.")
	runForAllFlag := flag.Bool("all", false, "Whether to run for all archive enabled projects.")

	dbHost := flag.String("db_host", "localhost", "")
	dbPort := flag.Int("db_port", 5432, "")
	dbUser := flag.String("db_user", "autometa", "")
	dbName := flag.String("db_name", "autometa", "")
	dbPass := flag.String("db_pass", "@ut0me7a", "")

	flag.Parse()
	defer util.NotifyOnPanic(taskID, *envFlag)

	if *envFlag != "development" && *envFlag != "staging" && *envFlag != "production" {
		panic(fmt.Errorf("env [ %s ] not recognised", *envFlag))
	} else if *projectIDFlag == 0 && !*runForAllFlag {
		panic(fmt.Errorf("Invalid project id %d", *projectIDFlag))
	} else if *startDateFlag != "" && *endDateFlag == "" {
		panic(fmt.Errorf("Must provide both start_date and end_date"))
	}

	pbLog.Info("Starting to initialize database.")
	config := &C.Configuration{
		AppName: "script_push_to_bigquery",
		Env:     *envFlag,
		DBInfo: C.DBConf{
			Host:     *dbHost,
			Port:     *dbPort,
			User:     *dbUser,
			Name:     *dbName,
			Password: *dbPass,
		},
	}
	C.InitConf(config.Env)

	err := C.InitDB(config.DBInfo)
	if err != nil {
		pbLog.WithError(err).Fatal("Failed to initialize DB")
	}
	db := C.GetServices().Db
	defer db.Close()

	var cloudManager filestore.FileManager
	if *envFlag == "development" {
		cloudManager = serviceDisk.New(*bucketNameFlag)
	} else {
		cloudManager, err = serviceGCS.New(*bucketNameFlag)
		if err != nil {
			pbLog.WithError(err).Errorln("Failed to init New GCS Client")
			panic(err)
		}
	}

	var startTime, endTime time.Time
	if *startDateFlag != "" {
		startTime, err = time.Parse(util.DATETIME_FORMAT_YYYYMMDD_HYPHEN, *startDateFlag)
		if err != nil {
			pbLog.WithError(err).Fatal("Invalid start_time. Format must be YYYY-MM-DD")
		}
		endTime, err = time.Parse(util.DATETIME_FORMAT_YYYYMMDD_HYPHEN, *endDateFlag)
		if err != nil {
			pbLog.WithError(err).Fatal("Invalid end_time. Format must be YYYY-MM-DD")
		}
	}

	diskManager := serviceDisk.New(*localDiskTmpDirFlag)

	allJobDetails := make(map[uint64][]string)
	var projectErrors []error
	if *runForAllFlag {
		allJobDetails, projectErrors = T.ArchiveEvents(db, &cloudManager, diskManager, *maxLookbackDaysFlag, startTime, endTime)
	} else {
		jobDetails, err := T.ArchiveEventsForProject(db, &cloudManager, diskManager, *projectIDFlag, *maxLookbackDaysFlag, startTime, endTime, false)
		if err != nil {
			projectErrors = append(projectErrors, err)
		}
		allJobDetails[*projectIDFlag] = jobDetails
	}

	err = util.NotifyThroughSNS("archive_events", *envFlag, allJobDetails)
	if err != nil {
		pbLog.WithError(err).Error("SNS notification failed", allJobDetails)
	}
	if len(projectErrors) != 0 {
		for _, err = range projectErrors {
			pbLog.WithError(err).Error("Error while archiving events")
		}
		panic(projectErrors)
	}
}
