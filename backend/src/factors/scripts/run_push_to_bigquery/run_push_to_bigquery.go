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

func main() {
	envFlag := flag.String("env", "development", "Environment. Could be development|staging|production.")
	bucketNameFlag := flag.String("bucket_name", "/usr/local/var/factors/cloud_storage", "Bucket name for production.")
	projectIDFlag := flag.Int64("project_id", 0, "Project id to be run for.")
	runForAllFlag := flag.Bool("all", false, "Whether to run for all project with bigquery enabled.")
	startDateFlag := flag.String("start_date", "", "Start date in format YYYY-MM-DD to process older files. Inclusive.")
	endDateFlag := flag.String("end_date", "", "End date in format YYYY-MM-DD to process older files. Inclusive")

	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	memSQLPort := flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser := flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName := flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass := flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	memSQLCertificate := flag.String("memsql_cert", "", "")
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypeMemSQL, "Primary datastore type as memsql or postgres")
	sentryDSN := flag.String("sentry_dsn", "", "Sentry DSN")

	gcpProjectID := flag.String("gcp_project_id", "", "Project ID on Google Cloud")
	gcpProjectLocation := flag.String("gcp_project_location", "", "Location of google cloud project cluster")
	overrideHealthcheckPingID := flag.String("healthcheck_ping_id", "", "Override default healthcheck ping id.")
	overrideAppName := flag.String("app_name", "", "Override default app_name.")

	flag.Parse()

	defaultAppName := "bigquery_upload"
	defaultHealthcheckPingID := C.HealthcheckBigqueryUploadPingID
	healthcheckPingID := C.GetHealthcheckPingID(defaultHealthcheckPingID, *overrideHealthcheckPingID)
	appName := C.GetAppName(defaultAppName, *overrideAppName)

	var pbLog = log.WithField("prefix", appName)
	defer C.PingHealthcheckForPanic(appName, *envFlag, healthcheckPingID)

	if *envFlag != "development" && *envFlag != "staging" && *envFlag != "production" {
		panic(fmt.Errorf("env [ %s ] not recognised", *envFlag))
	} else if *projectIDFlag == 0 && !*runForAllFlag {
		panic(fmt.Errorf("Invalid project id %d", *projectIDFlag))
	} else if *startDateFlag != "" && *endDateFlag == "" {
		panic(fmt.Errorf("Both start and end dates must be specified"))
	}

	pbLog.Info("Starting to initialize database.")
	config := &C.Configuration{
		AppName:            appName,
		Env:                *envFlag,
		GCPProjectID:       *gcpProjectID,
		GCPProjectLocation: *gcpProjectLocation,
		SentryDSN: *sentryDSN,
		MemSQLInfo: C.DBConf{
			Host:        *memSQLHost,
			Port:        *memSQLPort,
			User:        *memSQLUser,
			Name:        *memSQLName,
			Password:    *memSQLPass,
			Certificate: *memSQLCertificate,
			AppName:     appName,
		},
		PrimaryDatastore: *primaryDatastore,
	}
	C.InitConf(config)

	err := C.InitDB(*config)
	if err != nil {
		pbLog.WithError(err).Panic("Failed to initialize DB")
	}

	C.InitSentryLogging(config.SentryDSN, config.AppName)
	C.InitMetricsExporter(config.Env, config.AppName, config.GCPProjectID, config.GCPProjectLocation)
	defer C.WaitAndFlushAllCollectors(65 * time.Second)

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
			pbLog.WithError(err).Panic("Start date must have format YYYY-MM-DD")
		}
		endDate, err = time.Parse(util.DATETIME_FORMAT_YYYYMMDD_HYPHEN, *endDateFlag)
		if err != nil {
			pbLog.WithError(err).Panic("End date must have format YYYY-MM-DD")
		}
	}

	allJobDetails := make(map[int64][]string)
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

	if len(projectErrors) != 0 {
		for _, err = range projectErrors {
			pbLog.WithError(err).Error("Error while processing files for Bigquery")
			C.PingHealthcheckForFailure(healthcheckPingID, err)
		}
	} else {
		C.PingHealthcheckForSuccess(healthcheckPingID, allJobDetails)
	}
}
