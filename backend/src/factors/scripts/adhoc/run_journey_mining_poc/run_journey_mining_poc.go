package main

import (
	"flag"
	"fmt"
	"net/http"
	"strings"
	"time"

	C "factors/config"
	"factors/filestore"
	"factors/model/model"
	"factors/model/store"
	serviceDisk "factors/services/disk"
	serviceGCS "factors/services/gcstorage"
	T "factors/task"
	U "factors/util"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

func main() {
	envFlag := flag.String("env", C.DEVELOPMENT, "Environment. Could be development|staging|production")
	eventFiles := flag.String("event_files", "", "Comma separated list of event files")
	userFiles := flag.String("user_files", "", "Comma separated list of user files")
	projectIDFlag := flag.Uint64("project_id", 399, "ProjectID to run journey mining")
	startDateFlag := flag.String("start_date", "2020-09-01", "Start date in YYYY-MM-DD format. Inclusive.")
	endDateFlag := flag.String("end_date", "2020-09-23", "End date in YYYY-MM-DD format. Inclusive.")
	lookBackDaysFlag := flag.Int64("lookback_days", 0, "Number of days to lookback from the start date given")
	includeSessionFlag := flag.Bool("include_session", true, "Whether to auto include $session event in journey")
	sessionPropertyFlag := flag.String("session_property", "$campaign", "Propert of $session event shown along path")

	bucketNameFlag := flag.String("bucket_name", "/usr/local/var/factors/cloud_storage", "Bucket name for production")
	localDiskTmpDirFlag := flag.String("tmp_dir", "/usr/local/var/factors/local_disk/tmp", "Local directory path for putting tmp files.")

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

	flag.Parse()
	taskID := "Script#JourneyMiningPOC"
	defer U.NotifyOnPanic(taskID, *envFlag)
	logCtx := log.WithFields(log.Fields{"Prefix": taskID})

	if *envFlag != C.DEVELOPMENT && *envFlag != C.STAGING && *envFlag != C.PRODUCTION {
		panic(fmt.Errorf("env [ %s ] not recognised", *envFlag))
	} else if *projectIDFlag == 0 {
		panic(fmt.Errorf("invalid project id %d", *projectIDFlag))
	} else if *startDateFlag == "" || *endDateFlag == "" {
		panic(fmt.Errorf("must provide both start_date and end_date"))
	}

	logCtx = logCtx.WithFields(log.Fields{
		"Env":       *envFlag,
		"ProjectID": *projectIDFlag,
	})

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
		PrimaryDatastore: *primaryDatastore,
	}

	C.InitConf(config)
	err := C.InitDB(*config)
	if err != nil {
		logCtx.WithError(err).Fatal("Failed to initialize DB")
	}
	db := C.GetServices().Db

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

	jobDetails := make([]string, 0, 0)
	if *eventFiles == "" {
		eventFilePaths, userFilePaths, err := downloadAndGetArchivalFilesForRange(db, &cloudManager, diskManager,
			int64(*projectIDFlag), *startDateFlag, *endDateFlag, &jobDetails)
		if err != nil {
			logCtx.WithError(err).Fatal("Failed to download archive files")
		}
		*eventFiles = strings.Join(eventFilePaths, ",")
		*userFiles = strings.Join(userFilePaths, ",")
		if *eventFiles == "" {
			// User files can still be empty if there are no customer_user_id for the project.
			logCtx.Fatal("No files found for the given time range")
		}
		jobDetails = append(jobDetails, fmt.Sprintf("%d files to be processed for journey mining", len(eventFilePaths)))
	}

	var startTime, endTime time.Time
	var startTimeUnix, endTimeUnix int64
	istLocation := U.GetTimeLocationFor(U.TimeZoneStringIST)
	startTime, err = time.ParseInLocation(U.DATETIME_FORMAT_YYYYMMDD_HYPHEN, *startDateFlag, istLocation)
	if err != nil {
		logCtx.WithError(err).Fatal("Invalid start_time. Format must be YYYY-MM-DD")
	}
	endTime, err = time.ParseInLocation(U.DATETIME_FORMAT_YYYYMMDD_HYPHEN, *endDateFlag, istLocation)
	if err != nil {
		logCtx.WithError(err).Fatal("Invalid end_time. Format must be YYYY-MM-DD")
	}

	endTime = endTime.AddDate(0, 0, 1) // Add one day to make end date inclusive.
	startTimeUnix, endTimeUnix = startTime.Unix(), endTime.Unix()-1

	logCtx.WithFields(log.Fields{
		"IncludeSession":  *includeSessionFlag,
		"SessionProperty": *sessionPropertyFlag,
		"StartDate":       *startDateFlag,
		"EndDate":         *endDateFlag,
	}).Infof("Starting journey mining")
	jobStartTime := U.TimeNowUnix()
	journeyEvents, goalEvents := getChargebeeSept22Journey()
	store.GetStore().GetWeightedJourneyMatrix(int64(*projectIDFlag), journeyEvents, goalEvents, startTimeUnix,
		endTimeUnix, *lookBackDaysFlag, *eventFiles, *userFiles, *includeSessionFlag, *sessionPropertyFlag, cloudManager)

	timeTaken := U.SecondsToHMSString(U.TimeNowUnix() - jobStartTime)
	jobDetails = append(jobDetails, fmt.Sprintf("Time taken for journey mining: %s", timeTaken))
	U.NotifyThroughSNS(config.AppName, config.Env, jobDetails)

	// model.ChargebeeAnalysis(getSegmentAGoals(), getSegmentBGoals(), getSegmentCGoals(),
	// getSegmentVisitedPagesEvents(), *files, startTimeUnix, endTimeUnix, cloudManager)
}

func downloadAndGetArchivalFilesForRange(db *gorm.DB, cloudManager *filestore.FileManager, diskManager *serviceDisk.DiskDriver,
	projectID int64, startDate, endDate string, jobDetails *[]string) ([]string, []string, error) {

	startTime, err := time.Parse(U.DATETIME_FORMAT_YYYYMMDD_HYPHEN, startDate)
	if err != nil {
		log.WithError(err).Fatal("Invalid start_time. Format must be YYYY-MM-DD")
	}
	endTime, err := time.Parse(U.DATETIME_FORMAT_YYYYMMDD_HYPHEN, endDate)
	if err != nil {
		log.WithError(err).Fatal("Invalid end_time. Format must be YYYY-MM-DD")
	}

	// Add 1 day buffer to start and end time to avoid any UTC / IST overlap issue.
	startTime = startTime.AddDate(0, 0, -1)
	endTime = endTime.AddDate(0, 0, 1)

	*jobDetails = append(*jobDetails, fmt.Sprintf("Downloading archival files for range %v - %v", startTime, endTime))
	archiveJobDetails, err := T.ArchiveEventsForProject(db, cloudManager, diskManager, projectID, 0, startTime, endTime, true)
	if err != nil {
		return []string{}, []string{}, err
	}
	*jobDetails = append(*jobDetails, archiveJobDetails...)

	eventFilePaths, userFilePaths, errCode := store.GetStore().GetArchivalFileNamesForProject(
		projectID, startTime, endTime)
	if errCode != http.StatusFound {
		return eventFilePaths, userFilePaths, fmt.Errorf("error getting filepaths for project")
	}
	return eventFilePaths, userFilePaths, nil
}

func get38EventsScheduleDemoJourney() ([]model.QueryEventWithProperties, []model.QueryEventWithProperties) {
	return []model.QueryEventWithProperties{
			{Name: "www.chargebee.com"},
			{Name: "www.chargebee.com/pricing"},
			{Name: "www.chargebee.com/stripe-recurring-payments-chargebee"},
			{Name: "www.chargebee.com/subscription-management"},
			{Name: "www.chargebee.com/recurring-billing-invoicing"},
			{Name: "www.chargebee.com/recurring-payments-with-chargebee"},
			{Name: "www.chargebee.com/subscription-management-with-chargebee"},
			{Name: "www.chargebee.com/launch"},
			{Name: "www.chargebee.com/customers"},
			{Name: "www.chargebee.com/saas-pricing-models"},
			{Name: "www.chargebee.com/for-self-service-subscription-business"},
			{Name: "www.chargebee.com/integrations"},
			{Name: "www.chargebee.com/subscription-management/create-manage-plans"},
			{Name: "www.chargebee.com/trial-signup"},
			{Name: "www.chargebee.com/schedule-a-demo"},
			{Name: "www.chargebee.com/recurring-billing-with-chargebee"},
			{Name: "www.chargebee.com/recurring-payments"},
			{Name: "www.chargebee.com/saas-accounting-and-taxes"},
			{Name: "www.chargebee.com/payment-gateways"},
			{Name: "www.chargebee.com/payment-gateways/stripe"},
			{Name: "www.chargebee.com/saas-reporting"},
			{Name: "www.chargebee.com/enterprise-subscription-billing"},
			{Name: "www.chargebee.com/features"},
			{Name: "www.chargebee.com/compare-competitors/recurly"},
			{Name: "www.chargebee.com/recurring-payments/direct-debit-payments"},
			{Name: "www.chargebee.com/for-subscription-finance-operations"},
			{Name: "www.chargebee.com/customer-retention/churn-management"},
			{Name: "www.chargebee.com/subscription-management/customer-self-service-portal"},
			{Name: "www.chargebee.com/for-sales-driven-saas"},
			{Name: "www.chargebee.com/compare-competitors/zuora"},
			{Name: "www.chargebee.com/campaigns/saas-subscription-billing"},
			{Name: "www.chargebee.com/security"},
			{Name: "www.chargebee.com/saas-billing"},
			{Name: "www.chargebee.com/partners"},
			{Name: "www.chargebee.com/recurring-payments/payment-methods"},
			{Name: "www.chargebee.com/payment-gateways/braintree"},
			{Name: "www.chargebee.com/compare-competitors/chargify"},
		}, []model.QueryEventWithProperties{
			{Name: "$form_submitted", Properties: []model.QueryProperty{
				{
					Entity:    model.PropertyEntityEvent,
					LogicalOp: "AND",
					Operator:  model.EqualsOpStr,
					Property:  "$page_url",
					Type:      U.PropertyTypeCategorical,
					Value:     "www.chargebee.com/schedule-a-demo",
				},
				{
					Entity:    model.PropertyEntityEvent,
					LogicalOp: "OR",
					Operator:  model.EqualsOpStr,
					Property:  "$page_url",
					Type:      U.PropertyTypeCategorical,
					Value:     "www.chargebee.com/schedule-a-demo/",
				},
			}},
		}
}

func get38EventsTrialSignupJourney() ([]model.QueryEventWithProperties, []model.QueryEventWithProperties) {
	return []model.QueryEventWithProperties{
			{Name: "www.chargebee.com"},
			{Name: "www.chargebee.com/pricing"},
			{Name: "www.chargebee.com/stripe-recurring-payments-chargebee"},
			{Name: "www.chargebee.com/subscription-management"},
			{Name: "www.chargebee.com/recurring-billing-invoicing"},
			{Name: "www.chargebee.com/recurring-payments-with-chargebee"},
			{Name: "www.chargebee.com/subscription-management-with-chargebee"},
			{Name: "www.chargebee.com/launch"},
			{Name: "www.chargebee.com/customers"},
			{Name: "www.chargebee.com/saas-pricing-models"},
			{Name: "www.chargebee.com/for-self-service-subscription-business"},
			{Name: "www.chargebee.com/integrations"},
			{Name: "www.chargebee.com/subscription-management/create-manage-plans"},
			{Name: "www.chargebee.com/trial-signup"},
			{Name: "www.chargebee.com/schedule-a-demo"},
			{Name: "www.chargebee.com/recurring-billing-with-chargebee"},
			{Name: "www.chargebee.com/recurring-payments"},
			{Name: "www.chargebee.com/saas-accounting-and-taxes"},
			{Name: "www.chargebee.com/payment-gateways"},
			{Name: "www.chargebee.com/payment-gateways/stripe"},
			{Name: "www.chargebee.com/saas-reporting"},
			{Name: "www.chargebee.com/enterprise-subscription-billing"},
			{Name: "www.chargebee.com/features"},
			{Name: "www.chargebee.com/compare-competitors/recurly"},
			{Name: "www.chargebee.com/recurring-payments/direct-debit-payments"},
			{Name: "www.chargebee.com/for-subscription-finance-operations"},
			{Name: "www.chargebee.com/customer-retention/churn-management"},
			{Name: "www.chargebee.com/subscription-management/customer-self-service-portal"},
			{Name: "www.chargebee.com/for-sales-driven-saas"},
			{Name: "www.chargebee.com/compare-competitors/zuora"},
			{Name: "www.chargebee.com/campaigns/saas-subscription-billing"},
			{Name: "www.chargebee.com/security"},
			{Name: "www.chargebee.com/saas-billing"},
			{Name: "www.chargebee.com/partners"},
			{Name: "www.chargebee.com/recurring-payments/payment-methods"},
			{Name: "www.chargebee.com/payment-gateways/braintree"},
			{Name: "www.chargebee.com/compare-competitors/chargify"},
		}, []model.QueryEventWithProperties{
			{Name: "$form_submitted", Properties: []model.QueryProperty{
				{
					Entity:    model.PropertyEntityEvent,
					LogicalOp: "AND",
					Operator:  model.EqualsOpStr,
					Property:  "$page_url",
					Type:      U.PropertyTypeCategorical,
					Value:     "www.chargebee.com/trial-signup",
				},
				{
					Entity:    model.PropertyEntityEvent,
					LogicalOp: "OR",
					Operator:  model.EqualsOpStr,
					Property:  "$page_url",
					Type:      U.PropertyTypeCategorical,
					Value:     "www.chargebee.com/trial-signup/",
				},
			}},
		}
}

func get14EventsJourney() ([]model.QueryEventWithProperties, []model.QueryEventWithProperties) {
	return []model.QueryEventWithProperties{
			{Name: "www.chargebee.com"},
			{Name: "www.chargebee.com/pricing"},
			{Name: "www.chargebee.com/stripe-recurring-payments-chargebee"},
			{Name: "www.chargebee.com/subscription-management"},
			{Name: "www.chargebee.com/recurring-billing-invoicing"},
			{Name: "www.chargebee.com/recurring-payments-with-chargebee"},
			{Name: "www.chargebee.com/subscription-management-with-chargebee"},
			{Name: "www.chargebee.com/launch"},
			{Name: "www.chargebee.com/customers"},
			{Name: "www.chargebee.com/saas-pricing-models"},
			{Name: "www.chargebee.com/for-self-service-subscription-business"},
			{Name: "www.chargebee.com/integrations"},
			{Name: "www.chargebee.com/subscription-management/create-manage-plans"},
		}, []model.QueryEventWithProperties{
			{Name: "$form_submitted", Properties: []model.QueryProperty{
				{
					Entity:    model.PropertyEntityEvent,
					LogicalOp: "AND",
					Operator:  model.EqualsOpStr,
					Property:  "$page_url",
					Type:      U.PropertyTypeCategorical,
					Value:     "www.chargebee.com/schedule-a-demo",
				},
				{
					Entity:    model.PropertyEntityEvent,
					LogicalOp: "OR",
					Operator:  model.EqualsOpStr,
					Property:  "$page_url",
					Type:      U.PropertyTypeCategorical,
					Value:     "www.chargebee.com/schedule-a-demo/",
				},
			}},
		}
}

func get15EventsScheduleDemoJourney() ([]model.QueryEventWithProperties, []model.QueryEventWithProperties) {
	return []model.QueryEventWithProperties{
			{Name: "www.chargebee.com"},
			{Name: "www.chargebee.com/stripe-recurring-payments-chargebee"},
			{Name: "www.chargebee.com/subscription-management-with-chargebee"},
			{Name: "www.chargebee.com/pricing"},
			{Name: "www.chargebee.com/trial-signup"},
			{Name: "www.chargebee.com/recurring-billing-invoicing"},
			{Name: "www.chargebee.com/schedule-a-demo"},
			{Name: "www.chargebee.com/launch"},
			{Name: "www.chargebee.com/subscription-management"},
			{Name: "www.chargebee.com/recurring-payments"},
			{Name: "www.chargebee.com/payment-gateways"},
			{Name: "www.chargebee.com/customers"},
			{Name: "www.chargebee.com/saas-pricing-models"},
			{Name: "www.chargebee.com/subscription-management/create-manage-plans"},
		}, []model.QueryEventWithProperties{
			{Name: "$form_submitted", Properties: []model.QueryProperty{
				{
					Entity:    model.PropertyEntityEvent,
					LogicalOp: "AND",
					Operator:  model.EqualsOpStr,
					Property:  "$page_url",
					Type:      U.PropertyTypeCategorical,
					Value:     "www.chargebee.com/schedule-a-demo",
				},
				{
					Entity:    model.PropertyEntityEvent,
					LogicalOp: "OR",
					Operator:  model.EqualsOpStr,
					Property:  "$page_url",
					Type:      U.PropertyTypeCategorical,
					Value:     "www.chargebee.com/schedule-a-demo/",
				},
			}},
		}
}

func get15EventsTrialSignupJourney() ([]model.QueryEventWithProperties, []model.QueryEventWithProperties) {
	return []model.QueryEventWithProperties{
			{Name: "www.chargebee.com"},
			{Name: "www.chargebee.com/stripe-recurring-payments-chargebee"},
			{Name: "www.chargebee.com/subscription-management-with-chargebee"},
			{Name: "www.chargebee.com/pricing"},
			{Name: "www.chargebee.com/trial-signup"},
			{Name: "www.chargebee.com/recurring-billing-invoicing"},
			{Name: "www.chargebee.com/schedule-a-demo"},
			{Name: "www.chargebee.com/launch"},
			{Name: "www.chargebee.com/subscription-management"},
			{Name: "www.chargebee.com/recurring-payments"},
			{Name: "www.chargebee.com/payment-gateways"},
			{Name: "www.chargebee.com/customers"},
			{Name: "www.chargebee.com/saas-pricing-models"},
			{Name: "www.chargebee.com/subscription-management/create-manage-plans"},
		}, []model.QueryEventWithProperties{
			{Name: "$form_submitted", Properties: []model.QueryProperty{
				{
					Entity:    model.PropertyEntityEvent,
					LogicalOp: "AND",
					Operator:  model.EqualsOpStr,
					Property:  "$page_url",
					Type:      U.PropertyTypeCategorical,
					Value:     "www.chargebee.com/trial-signup",
				},
				{
					Entity:    model.PropertyEntityEvent,
					LogicalOp: "OR",
					Operator:  model.EqualsOpStr,
					Property:  "$page_url",
					Type:      U.PropertyTypeCategorical,
					Value:     "www.chargebee.com/trial-signup/",
				},
			}},
		}
}

func get5EventScheduledDemoJourney() ([]model.QueryEventWithProperties, []model.QueryEventWithProperties) {
	return []model.QueryEventWithProperties{
			{Name: "www.chargebee.com/trial-signup"},
			{Name: "www.chargebee.com/schedule-a-demo"},
			{Name: "$form_submitted", Properties: []model.QueryProperty{
				{
					Entity:    model.PropertyEntityEvent,
					LogicalOp: "AND",
					Operator:  model.EqualsOpStr,
					Property:  "$page_url",
					Type:      U.PropertyTypeCategorical,
					Value:     "www.chargebee.com/trial-signup",
				},
				{
					Entity:    model.PropertyEntityEvent,
					LogicalOp: "OR",
					Operator:  model.EqualsOpStr,
					Property:  "$page_url",
					Type:      U.PropertyTypeCategorical,
					Value:     "www.chargebee.com/trial-signup/",
				},
			}},
		}, []model.QueryEventWithProperties{
			{Name: "$form_submitted", Properties: []model.QueryProperty{
				{
					Entity:    model.PropertyEntityEvent,
					LogicalOp: "AND",
					Operator:  model.EqualsOpStr,
					Property:  "$page_url",
					Type:      U.PropertyTypeCategorical,
					Value:     "www.chargebee.com/schedule-a-demo",
				},
				{
					Entity:    model.PropertyEntityEvent,
					LogicalOp: "OR",
					Operator:  model.EqualsOpStr,
					Property:  "$page_url",
					Type:      U.PropertyTypeCategorical,
					Value:     "www.chargebee.com/schedule-a-demo/",
				},
			}},
		}
}

func get5EventTrialSignupJourney() ([]model.QueryEventWithProperties, []model.QueryEventWithProperties) {
	return []model.QueryEventWithProperties{
			{Name: "www.chargebee.com/trial-signup"},
			{Name: "www.chargebee.com/schedule-a-demo"},
			{Name: "$form_submitted", Properties: []model.QueryProperty{
				{
					Entity:    model.PropertyEntityEvent,
					LogicalOp: "AND",
					Operator:  model.EqualsOpStr,
					Property:  "$page_url",
					Type:      U.PropertyTypeCategorical,
					Value:     "www.chargebee.com/schedule-a-demo",
				},
				{
					Entity:    model.PropertyEntityEvent,
					LogicalOp: "OR",
					Operator:  model.EqualsOpStr,
					Property:  "$page_url",
					Type:      U.PropertyTypeCategorical,
					Value:     "www.chargebee.com/schedule-a-demo/",
				},
			}},
		}, []model.QueryEventWithProperties{
			{Name: "$form_submitted", Properties: []model.QueryProperty{
				{
					Entity:    model.PropertyEntityEvent,
					LogicalOp: "AND",
					Operator:  model.EqualsOpStr,
					Property:  "$page_url",
					Type:      U.PropertyTypeCategorical,
					Value:     "www.chargebee.com/trial-signup",
				},
				{
					Entity:    model.PropertyEntityEvent,
					LogicalOp: "OR",
					Operator:  model.EqualsOpStr,
					Property:  "$page_url",
					Type:      U.PropertyTypeCategorical,
					Value:     "www.chargebee.com/trial-signup/",
				},
			}},
		}
}

func get5EventNotDoneScheduledDemoJourney() ([]model.QueryEventWithProperties, []model.QueryEventWithProperties) {
	return []model.QueryEventWithProperties{
			{Name: "www.chargebee.com/trial-signup"},
			{Name: "www.chargebee.com/schedule-a-demo"},
		}, []model.QueryEventWithProperties{
			{Name: "$form_submitted", Properties: []model.QueryProperty{
				{
					Entity:    model.PropertyEntityEvent,
					LogicalOp: "AND",
					Operator:  model.EqualsOpStr,
					Property:  "$page_url",
					Type:      U.PropertyTypeCategorical,
					Value:     "www.chargebee.com/trial-signup",
				},
				{
					Entity:    model.PropertyEntityEvent,
					LogicalOp: "OR",
					Operator:  model.EqualsOpStr,
					Property:  "$page_url",
					Type:      U.PropertyTypeCategorical,
					Value:     "www.chargebee.com/trial-signup/",
				},
			}},
			{Name: "$form_submitted", Properties: []model.QueryProperty{
				{
					Entity:    model.PropertyEntityEvent,
					LogicalOp: "AND",
					Operator:  model.EqualsOpStr,
					Property:  "$page_url",
					Type:      U.PropertyTypeCategorical,
					Value:     "www.chargebee.com/schedule-a-demo",
				},
				{
					Entity:    model.PropertyEntityEvent,
					LogicalOp: "OR",
					Operator:  model.EqualsOpStr,
					Property:  "$page_url",
					Type:      U.PropertyTypeCategorical,
					Value:     "www.chargebee.com/schedule-a-demo/",
				},
			}},
		}
}

func get10EventHippoVideoJourney() ([]model.QueryEventWithProperties, []model.QueryEventWithProperties) {
	return []model.QueryEventWithProperties{
			{Name: "www.hippovideo.io/hippo-video-editor.html"},
			{Name: "www.hippovideo.io/video-for-sales-and-prospecting.html"},
			{Name: "www.hippovideo.io/edit-videos-online.html"},
			{Name: "www.hippovideo.io/pricing.html"},
			{Name: "www.hippovideo.io/create-videos-online.html"},
			{Name: "www.hippovideo.io/video-editing-software.html"},
			{Name: "www.hippovideo.io/video-for-sales.html"},
			{Name: "www.hippovideo.io/video-personalization.html"},
		}, []model.QueryEventWithProperties{
			{Name: "$form_submitted", Properties: []model.QueryProperty{
				{
					Entity:    model.PropertyEntityEvent,
					LogicalOp: "AND",
					Operator:  model.EqualsOpStr,
					Property:  "$page_url",
					Type:      U.PropertyTypeCategorical,
					Value:     "www.hippovideo.io/users/sign_up",
				},
				{
					Entity:    model.PropertyEntityEvent,
					LogicalOp: "OR",
					Operator:  model.EqualsOpStr,
					Property:  "$page_url",
					Type:      U.PropertyTypeCategorical,
					Value:     "www.hippovideo.io/users/sign_up/",
				},
			}},
		}
}

func get6EventHippoVideoJourney() ([]model.QueryEventWithProperties, []model.QueryEventWithProperties) {
	return []model.QueryEventWithProperties{
			{Name: "www.hippovideo.io"},
			{Name: "www.hippovideo.io/video-for-sales-and-prospecting.html"},
			{Name: "www.hippovideo.io/pricing.html"},
			{Name: "www.hippovideo.io/video-for-sales.html"},
			{Name: "www.hippovideo.io/video-personalization.html"},
		}, []model.QueryEventWithProperties{
			{Name: "$form_submitted", Properties: []model.QueryProperty{
				{
					Entity:    model.PropertyEntityEvent,
					LogicalOp: "AND",
					Operator:  model.NotEqualOpStr,
					Property:  "$page_url",
					Type:      U.PropertyTypeCategorical,
					Value:     "www.hippovideo.io/users/sign_in",
				},
				{
					Entity:    model.PropertyEntityEvent,
					LogicalOp: "AND",
					Operator:  model.NotEqualOpStr,
					Property:  "$page_url",
					Type:      U.PropertyTypeCategorical,
					Value:     "www.hippovideo.io/users/sign_in/",
				},
			}},
		}
}

func get3EventLivspaceJourney() ([]model.QueryEventWithProperties, []model.QueryEventWithProperties) {
	return []model.QueryEventWithProperties{
			{Name: "stage2.livspace.com/in/lead-quiz"},
		}, []model.QueryEventWithProperties{
			{Name: "stage2_lead-quiz:language_next-button-click"},
		}
}

func getSegmentAGoals() []model.QueryEventWithProperties {
	return []model.QueryEventWithProperties{
		{Name: "$form_submitted", Properties: []model.QueryProperty{
			{
				Entity:    model.PropertyEntityEvent,
				LogicalOp: "AND",
				Operator:  model.EqualsOpStr,
				Property:  "$page_url",
				Type:      U.PropertyTypeCategorical,
				Value:     "www.chargebee.com/trial-signup",
			},
			{
				Entity:    model.PropertyEntityEvent,
				LogicalOp: "OR",
				Operator:  model.EqualsOpStr,
				Property:  "$page_url",
				Type:      U.PropertyTypeCategorical,
				Value:     "www.chargebee.com/trial-signup/",
			},
		}},
	}
}

func getSegmentBGoals() []model.QueryEventWithProperties {
	return []model.QueryEventWithProperties{
		{Name: "$form_submitted", Properties: []model.QueryProperty{
			{
				Entity:    model.PropertyEntityEvent,
				LogicalOp: "AND",
				Operator:  "contains",
				Property:  "$page_url",
				Type:      U.PropertyTypeCategorical,
				Value:     "www.chargebee.com/schedule-a-demo",
			},
			{
				Entity:    model.PropertyEntityEvent,
				LogicalOp: "OR",
				Operator:  "contains",
				Property:  "$page_url",
				Type:      U.PropertyTypeCategorical,
				Value:     "www.chargebee.com/subscription-management-with-chargebee",
			},
			{
				Entity:    model.PropertyEntityEvent,
				LogicalOp: "OR",
				Operator:  "contains",
				Property:  "$page_url",
				Type:      U.PropertyTypeCategorical,
				Value:     "www.chargebee.com/recurring-billing-with-chargebee",
			},
			{
				Entity:    model.PropertyEntityEvent,
				LogicalOp: "OR",
				Operator:  "contains",
				Property:  "$page_url",
				Type:      U.PropertyTypeCategorical,
				Value:     "www.chargebee.com/recurring-payments-with-chargebee",
			},
			{
				Entity:    model.PropertyEntityEvent,
				LogicalOp: "OR",
				Operator:  "contains",
				Property:  "$page_url",
				Type:      U.PropertyTypeCategorical,
				Value:     "www.chargebee.com/stripe-recurring-payments-chargebee/#take-a-tour",
			},
		}},
	}
}

func getSegmentCGoals() []model.QueryEventWithProperties {
	return []model.QueryEventWithProperties{
		{Name: "$session", Properties: []model.QueryProperty{
			{
				Entity:    model.PropertyEntityUser,
				LogicalOp: "AND",
				Operator:  model.EqualsOpStr,
				Property:  "$hubspot_company_contact_hs_band",
				Type:      U.PropertyTypeCategorical,
				Value:     "$none",
			},
		}},
	}
}

func getSegmentVisitedPagesEvents() []model.QueryEventWithProperties {
	return []model.QueryEventWithProperties{
		{Name: "www.chargebee.com/pricing"},
		{Name: "www.chargebee.com/for-education"},
		{Name: "www.chargebee.com/for-self-service-subscription-business"},
		{Name: "www.chargebee.com/subscription-management"},
		{Name: "www.chargebee.com/recurring-billing-invoicing"},
		{Name: "www.chargebee.com/recurring-payments"},
		{Name: "www.chargebee.com/payment-gateways"},
		{Name: "www.chargebee.com/customers"},
		{Name: "www.chargebee.com/saas-pricing-models"},
		{Name: "www.chargebee.com/subscription-management/create-manage-plans"},
		{Name: "www.chargebee.com/trial-signup"},
		{Name: "www.chargebee.com/schedule-a-demo"},
	}
}

func getChargebeeHubspotContactUpdatedJourney() ([]model.QueryEventWithProperties, []model.QueryEventWithProperties) {
	return []model.QueryEventWithProperties{
			{Name: "$hubspot_contact_updated"},
		}, []model.QueryEventWithProperties{
			{Name: "$hubspot_contact_updated", Properties: []model.QueryProperty{
				{
					Entity:    model.PropertyEntityEvent,
					LogicalOp: "AND",
					Operator:  model.EqualsOpStr,
					Property:  "$hubspot_contact_lifecyclestage",
					Type:      U.PropertyTypeCategorical,
					Value:     "customer",
				},
			}},
		}
}

func getChargebeeSept22Journey() ([]model.QueryEventWithProperties, []model.QueryEventWithProperties) {
	return []model.QueryEventWithProperties{
			{Name: "$hubspot_contact_updated", Properties: []model.QueryProperty{
				{
					Entity:    model.PropertyEntityEvent,
					LogicalOp: "AND",
					Operator:  model.EqualsOpStr,
					Property:  "$hubspot_contact_marketing_conversions_mode",
					Type:      U.PropertyTypeCategorical,
					Value:     "SignUp",
				},
				{
					Entity:    model.PropertyEntityEvent,
					LogicalOp: "OR",
					Operator:  "contains",
					Property:  "$hubspot_contact_marketing_conversions_mode",
					Type:      U.PropertyTypeCategorical,
					Value:     "Drift",
				},
				{
					Entity:    model.PropertyEntityEvent,
					LogicalOp: "OR",
					Operator:  "contains",
					Property:  "$hubspot_contact_marketing_conversions_mode",
					Type:      U.PropertyTypeCategorical,
					Value:     "Calendly",
				},
			}},
			{Name: "www.chargebee.com/pricing"},
			{Name: "www.chargebee.com/trial-signup"},
			{Name: "www.chargebee.com/schedule-a-demo"},
			{Name: "www.chargebee.com/subscription-management"},
			{Name: "www.chargebee.com/recurring-billing-invoicing"},
			{Name: "www.chargebee.com/recurring-payments"},
			{Name: "www.chargebee.com/saas-accounting-and-taxes"},
			{Name: "www.chargebee.com/saas-reporting"},
			{Name: "www.chargebee.com/integrations"},
			{Name: "www.chargebee.com/payment-gateways"},
			{Name: "www.chargebee.com/for-education"},
			{Name: "www.chargebee.com/for-self-service-subscription-busines"},
			{Name: "www.chargebee.com/for-sales-driven-saas"},
			{Name: "www.chargebee.com/for-subscription-finance-operations"},
			{Name: "www.chargebee.com/enterprise-subscription-billing"},
			{Name: "www.chargebee.com/customers"},
		}, []model.QueryEventWithProperties{
			{Name: "$hubspot_contact_updated", Properties: []model.QueryProperty{
				{
					Entity:    model.PropertyEntityEvent,
					LogicalOp: "AND",
					Operator:  model.NotEqualOpStr,
					Property:  "$hubspot_contact_demo_booked_on",
					Type:      U.PropertyTypeCategorical,
					Value:     "$none",
				},
			}},
		}
}

func getObserveAISept24Journey() ([]model.QueryEventWithProperties, []model.QueryEventWithProperties) {
	return []model.QueryEventWithProperties{
			{Name: "$hubspot_contact_updated", Properties: []model.QueryProperty{
				{
					Entity:    model.PropertyEntityEvent,
					LogicalOp: "AND",
					Operator:  model.EqualsOpStr,
					Property:  "$hubspot_contact_lifecyclestage",
					Type:      U.PropertyTypeCategorical,
					Value:     "lead",
				},
			}},
			{Name: "$hubspot_contact_updated", Properties: []model.QueryProperty{
				{
					Entity:    model.PropertyEntityEvent,
					LogicalOp: "AND",
					Operator:  model.EqualsOpStr,
					Property:  "$hubspot_contact_lifecyclestage",
					Type:      U.PropertyTypeCategorical,
					Value:     "marketingqualifiedlead",
				},
			}},
			{Name: "$hubspot_contact_updated", Properties: []model.QueryProperty{
				{
					Entity:    model.PropertyEntityEvent,
					LogicalOp: "AND",
					Operator:  model.EqualsOpStr,
					Property:  "$hubspot_contact_lifecyclestage",
					Type:      U.PropertyTypeCategorical,
					Value:     "salesqualifiedlead",
				},
			}},
			{Name: "$hubspot_contact_updated", Properties: []model.QueryProperty{
				{
					Entity:    model.PropertyEntityEvent,
					LogicalOp: "AND",
					Operator:  model.EqualsOpStr,
					Property:  "$hubspot_contact_lifecyclestage",
					Type:      U.PropertyTypeCategorical,
					Value:     "customer",
				},
			}},
			{Name: "$hubspot_contact_updated", Properties: []model.QueryProperty{
				{
					Entity:    model.PropertyEntityEvent,
					LogicalOp: "AND",
					Operator:  model.EqualsOpStr,
					Property:  "$hubspot_contact_lifecyclestage",
					Type:      U.PropertyTypeCategorical,
					Value:     "other",
				},
			}},
			{Name: "$hubspot_contact_updated", Properties: []model.QueryProperty{
				{
					Entity:    model.PropertyEntityEvent,
					LogicalOp: "AND",
					Operator:  model.EqualsOpStr,
					Property:  "$hubspot_contact_lifecyclestage",
					Type:      U.PropertyTypeCategorical,
					Value:     "$none",
				},
			}},
			{Name: "$hubspot_contact_updated", Properties: []model.QueryProperty{
				{
					Entity:    model.PropertyEntityEvent,
					LogicalOp: "AND",
					Operator:  model.EqualsOpStr,
					Property:  "$hubspot_contact_salesforceopportunitystage",
					Type:      U.PropertyTypeCategorical,
					Value:     "Qualification ( SS-3)",
				},
				{
					Entity:    model.PropertyEntityEvent,
					LogicalOp: "OR",
					Operator:  model.EqualsOpStr,
					Property:  "$hubspot_contact_salesforceopportunitystage",
					Type:      U.PropertyTypeCategorical,
					Value:     "SS-3 Qualification",
				},
			}},
			{Name: "$hubspot_contact_updated", Properties: []model.QueryProperty{
				{
					Entity:    model.PropertyEntityEvent,
					LogicalOp: "AND",
					Operator:  model.EqualsOpStr,
					Property:  "$hubspot_contact_salesforceopportunitystage",
					Type:      U.PropertyTypeCategorical,
					Value:     "SS-4 Opportunity Validation",
				},
			}},
			{Name: "$hubspot_contact_updated", Properties: []model.QueryProperty{
				{
					Entity:    model.PropertyEntityEvent,
					LogicalOp: "AND",
					Operator:  model.EqualsOpStr,
					Property:  "$hubspot_contact_salesforceopportunitystage",
					Type:      U.PropertyTypeCategorical,
					Value:     "SS-5 Need Confirmed",
				},
			}},
			{Name: "$hubspot_contact_updated", Properties: []model.QueryProperty{
				{
					Entity:    model.PropertyEntityEvent,
					LogicalOp: "AND",
					Operator:  model.EqualsOpStr,
					Property:  "$hubspot_contact_salesforceopportunitystage",
					Type:      U.PropertyTypeCategorical,
					Value:     "SS-6 Alignment & Poc",
				},
			}},
			{Name: "$hubspot_contact_updated", Properties: []model.QueryProperty{
				{
					Entity:    model.PropertyEntityEvent,
					LogicalOp: "AND",
					Operator:  model.EqualsOpStr,
					Property:  "$hubspot_contact_salesforceopportunitystage",
					Type:      U.PropertyTypeCategorical,
					Value:     "SS-7 Pov Complete & Shortlist",
				},
			}},
			{Name: "$hubspot_contact_updated", Properties: []model.QueryProperty{
				{
					Entity:    model.PropertyEntityEvent,
					LogicalOp: "AND",
					Operator:  model.EqualsOpStr,
					Property:  "$hubspot_contact_salesforceopportunitystage",
					Type:      U.PropertyTypeCategorical,
					Value:     "SS-8 Verbal & Terms",
				},
			}},
			{Name: "$hubspot_contact_updated", Properties: []model.QueryProperty{
				{
					Entity:    model.PropertyEntityEvent,
					LogicalOp: "AND",
					Operator:  model.EqualsOpStr,
					Property:  "$hubspot_contact_salesforceopportunitystage",
					Type:      U.PropertyTypeCategorical,
					Value:     "SS-12 Closed Lost",
				},
			}},
		}, []model.QueryEventWithProperties{
			{Name: "$hubspot_contact_updated", Properties: []model.QueryProperty{
				{
					Entity:    model.PropertyEntityEvent,
					LogicalOp: "AND",
					Operator:  model.EqualsOpStr,
					Property:  "$hubspot_contact_salesforceopportunitystage",
					Type:      U.PropertyTypeCategorical,
					Value:     "SS-10 Closed Won",
				},
				{
					Entity:    model.PropertyEntityEvent,
					LogicalOp: "OR",
					Operator:  model.EqualsOpStr,
					Property:  "$hubspot_contact_salesforceopportunitystage",
					Type:      U.PropertyTypeCategorical,
					Value:     "SS-11 Closed Won",
				},
			}},
		}
}
