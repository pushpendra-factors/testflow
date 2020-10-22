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

var taskID = "Script#PushToBigquery"
var pbLog = log.WithField("prefix", taskID)

func main() {
	envFlag := flag.String("env", "development", "Environment. Could be development|staging|production.")
	bucketNameFlag := flag.String("bucket_name", "/usr/local/var/factors/cloud_storage", "Bucket name for production.")
	projectIDFlag := flag.Uint64("project_id", 0, "Project id to be run for.")
	runForAllFlag := flag.Bool("all", false, "Whether to run for all project with bigquery enabled.")
	startDateFlag := flag.String("start_date", "", "Start date in format YYYY-MM-DD to process older files. Inclusive.")
	endDateFlag := flag.String("end_date", "", "End date in format YYYY-MM-DD to process older files. Inclusive")

	dbHost := flag.String("db_host", "localhost", "")
	dbPort := flag.Int("db_port", 5432, "")
	dbUser := flag.String("db_user", "autometa", "")
	dbName := flag.String("db_name", "autometa", "")
	dbPass := flag.String("db_pass", "@ut0me7a", "")
	sentryDSN := flag.String("sentry_dsn", "", "Sentry DSN")

	flag.Parse()
	defer util.NotifyOnPanic(taskID, *envFlag)

	if *envFlag != "development" && *envFlag != "staging" && *envFlag != "production" {
		panic(fmt.Errorf("env [ %s ] not recognised", *envFlag))
	} else if *projectIDFlag == 0 && !*runForAllFlag {
		panic(fmt.Errorf("Invalid project id %d", *projectIDFlag))
	} else if *startDateFlag != "" && *endDateFlag == "" {
		panic(fmt.Errorf("Both start and end dates must be specified"))
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
		SentryDSN: *sentryDSN,
	}
	C.InitConf(config.Env)

	err := C.InitDB(config.DBInfo)
	if err != nil {
		pbLog.WithError(err).Fatal("Failed to initialize DB")
	}

	C.InitSentryLogging(config.SentryDSN, config.AppName)
	defer C.SafeFlushAllCollectors()

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

	var startDate, endDate time.Time
	if *startDateFlag != "" {
		startDate, err = time.Parse(util.DATETIME_FORMAT_YYYYMMDD_HYPHEN, *startDateFlag)
		if err != nil {
			pbLog.WithError(err).Fatal("Start date must have format YYYY-MM-DD")
		}
		endDate, err = time.Parse(util.DATETIME_FORMAT_YYYYMMDD_HYPHEN, *endDateFlag)
		if err != nil {
			pbLog.WithError(err).Fatal("End date must have format YYYY-MM-DD")
		}
	}

	allJobDetails := make(map[uint64][]string)
	var projectErrors []error
	if *runForAllFlag {
		allJobDetails, projectErrors = T.PushToBigquery(&cloudManager, startDate, endDate)
	} else {
		jobDetails, err := T.PushToBigqueryForProject(&cloudManager, *projectIDFlag, startDate, endDate)
		if err != nil {
			projectErrors = append(projectErrors, err)
		}
		allJobDetails[*projectIDFlag] = jobDetails
	}
	err = util.NotifyThroughSNS("bigquery_upload", *envFlag, allJobDetails)
	if err != nil {
		pbLog.WithError(err).Error("SNS notification failed", allJobDetails)
	}

	if len(projectErrors) != 0 {
		for _, err = range projectErrors {
			pbLog.WithError(err).Error("Error while processing files for Bigquery")
		}
	}
}
