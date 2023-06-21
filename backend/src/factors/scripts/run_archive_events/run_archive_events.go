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
	localDiskTmpDirFlag := flag.String("tmp_dir", "/usr/local/var/factors/local_disk/tmp", "Local directory path for putting tmp files.")
	projectIDFlag := flag.Int64("project_id", 0, "Project id to be run for.")
	maxLookbackDaysFlag := flag.Int("max_lookback_days", 365, "Maximum number of lookback days for events.")
	startDateFlag := flag.String("start_date", "", "Start date in YYYY-MM-DD format to run for specific period. Inclusive.")
	endDateFlag := flag.String("end_date", "", "End date in YYYY-MM-DD format to run for specific period. Inclusive.")
	runForAllFlag := flag.Bool("all", false, "Whether to run for all archive enabled projects.")

	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	isPSCHost := flag.Int("memsql_is_psc_host", C.MemSQLDefaultDBParams.IsPSCHost, "")
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

	defaultAppName := "archive_events"
	defaultHealthcheckPingID := C.HealthcheckArchiveEventsPingID
	healthcheckPingID := C.GetHealthcheckPingID(defaultHealthcheckPingID, *overrideHealthcheckPingID)
	appName := C.GetAppName(defaultAppName, *overrideAppName)

	var pbLog = log.WithField("prefix", appName)
	defer C.PingHealthcheckForPanic(appName, *envFlag, healthcheckPingID)

	if *envFlag != "development" && *envFlag != "staging" && *envFlag != "production" {
		panic(fmt.Errorf("env [ %s ] not recognised", *envFlag))
	} else if *projectIDFlag == 0 && !*runForAllFlag {
		panic(fmt.Errorf("Invalid project id %d", *projectIDFlag))
	} else if *startDateFlag != "" && *endDateFlag == "" {
		panic(fmt.Errorf("Must provide both start_date and end_date"))
	}

	pbLog.Info("Starting to initialize database.")
	config := &C.Configuration{
		AppName:            appName,
		Env:                *envFlag,
		GCPProjectID:       *gcpProjectID,
		GCPProjectLocation: *gcpProjectLocation,
		SentryDSN:          *sentryDSN,
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
		PrimaryDatastore: *primaryDatastore,
	}
	C.InitConf(config)

	err := C.InitDB(*config)
	if err != nil {
		pbLog.WithError(err).Panic("Failed to initialize DB")
	}
	db := C.GetServices().Db
	defer db.Close()

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

	var startTime, endTime time.Time
	if *startDateFlag != "" {
		startTime, err = time.Parse(util.DATETIME_FORMAT_YYYYMMDD_HYPHEN, *startDateFlag)
		if err != nil {
			pbLog.WithError(err).Panic("Invalid start_time. Format must be YYYY-MM-DD")
		}
		endTime, err = time.Parse(util.DATETIME_FORMAT_YYYYMMDD_HYPHEN, *endDateFlag)
		if err != nil {
			pbLog.WithError(err).Panic("Invalid end_time. Format must be YYYY-MM-DD")
		}
	}

	diskManager := serviceDisk.New(*localDiskTmpDirFlag)

	allJobDetails := make(map[int64][]string)
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

	if len(projectErrors) != 0 {
		for _, err = range projectErrors {
			pbLog.WithError(err).Error("Error while archiving events")
			C.PingHealthcheckForFailure(healthcheckPingID, err)
		}
	} else {
		C.PingHealthcheckForSuccess(healthcheckPingID, allJobDetails)
	}
}
