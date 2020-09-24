package main

import (
	"flag"
	"fmt"
	"net/http"
	"strings"
	"time"

	C "factors/config"
	"factors/filestore"
	M "factors/model"
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

	dbHost := flag.String("db_host", "localhost", "")
	dbPort := flag.Int("db_port", 5432, "")
	dbUser := flag.String("db_user", "autometa", "")
	dbName := flag.String("db_name", "autometa", "")
	dbPass := flag.String("db_pass", "@ut0me7a", "")

	flag.Parse()
	taskID := "Script#JourneyMiningPOC"
	defer U.NotifyOnPanic(taskID, *envFlag)
	logCtx := log.WithFields(log.Fields{"Prefix": taskID})

	if *envFlag != C.DEVELOPMENT && *envFlag != C.STAGING && *envFlag != C.PRODUCTION {
		panic(fmt.Errorf("env [ %s ] not recognised", *envFlag))
	} else if *projectIDFlag == 0 {
		panic(fmt.Errorf("invalid project id %s", *projectIDFlag))
	} else if *startDateFlag == "" || *endDateFlag == "" {
		panic(fmt.Errorf("Must provide both start_date and end_date"))
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
		},
	}

	C.InitConf(config.Env)
	err := C.InitDB(config.DBInfo)
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
			*projectIDFlag, *startDateFlag, *endDateFlag, &jobDetails)
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
	M.GetWeightedJourneyMatrix(*projectIDFlag, journeyEvents, goalEvents, startTimeUnix,
		endTimeUnix, *lookBackDaysFlag, *eventFiles, *userFiles, *includeSessionFlag, *sessionPropertyFlag, cloudManager)

	timeTaken := U.SecondsToHMSString(U.TimeNowUnix() - jobStartTime)
	jobDetails = append(jobDetails, fmt.Sprintf("Time taken for journey mining: %s", timeTaken))
	U.NotifyThroughSNS(config.AppName, config.Env, jobDetails)

	// M.ChargebeeAnalysis(getSegmentAGoals(), getSegmentBGoals(), getSegmentCGoals(),
	// getSegmentVisitedPagesEvents(), *files, startTimeUnix, endTimeUnix, cloudManager)
}

func downloadAndGetArchivalFilesForRange(db *gorm.DB, cloudManager *filestore.FileManager, diskManager *serviceDisk.DiskDriver,
	projectID uint64, startDate, endDate string, jobDetails *[]string) ([]string, []string, error) {

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

	eventFilePaths, userFilePaths, errCode := M.GetArchivalFileNamesForProject(projectID, startTime, endTime)
	if errCode != http.StatusFound {
		return eventFilePaths, userFilePaths, fmt.Errorf("Error getting filepaths for project")
	}
	return eventFilePaths, userFilePaths, nil
}

func get38EventsScheduleDemoJourney() ([]M.QueryEventWithProperties, []M.QueryEventWithProperties) {
	return []M.QueryEventWithProperties{
			M.QueryEventWithProperties{Name: "www.chargebee.com"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/pricing"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/stripe-recurring-payments-chargebee"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/subscription-management"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/recurring-billing-invoicing"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/recurring-payments-with-chargebee"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/subscription-management-with-chargebee"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/launch"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/customers"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/saas-pricing-models"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/for-self-service-subscription-business"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/integrations"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/subscription-management/create-manage-plans"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/trial-signup"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/schedule-a-demo"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/recurring-billing-with-chargebee"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/recurring-payments"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/saas-accounting-and-taxes"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/payment-gateways"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/payment-gateways/stripe"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/saas-reporting"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/enterprise-subscription-billing"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/features"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/compare-competitors/recurly"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/recurring-payments/direct-debit-payments"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/for-subscription-finance-operations"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/customer-retention/churn-management"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/subscription-management/customer-self-service-portal"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/for-sales-driven-saas"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/compare-competitors/zuora"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/campaigns/saas-subscription-billing"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/security"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/saas-billing"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/partners"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/recurring-payments/payment-methods"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/payment-gateways/braintree"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/compare-competitors/chargify"},
		}, []M.QueryEventWithProperties{
			M.QueryEventWithProperties{Name: "$form_submitted", Properties: []M.QueryProperty{
				M.QueryProperty{
					Entity:    M.PropertyEntityEvent,
					LogicalOp: "AND",
					Operator:  M.EqualsOpStr,
					Property:  "$page_url",
					Type:      U.PropertyTypeCategorical,
					Value:     "www.chargebee.com/schedule-a-demo",
				},
				M.QueryProperty{
					Entity:    M.PropertyEntityEvent,
					LogicalOp: "OR",
					Operator:  M.EqualsOpStr,
					Property:  "$page_url",
					Type:      U.PropertyTypeCategorical,
					Value:     "www.chargebee.com/schedule-a-demo/",
				},
			}},
		}
}

func get38EventsTrialSignupJourney() ([]M.QueryEventWithProperties, []M.QueryEventWithProperties) {
	return []M.QueryEventWithProperties{
			M.QueryEventWithProperties{Name: "www.chargebee.com"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/pricing"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/stripe-recurring-payments-chargebee"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/subscription-management"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/recurring-billing-invoicing"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/recurring-payments-with-chargebee"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/subscription-management-with-chargebee"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/launch"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/customers"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/saas-pricing-models"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/for-self-service-subscription-business"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/integrations"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/subscription-management/create-manage-plans"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/trial-signup"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/schedule-a-demo"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/recurring-billing-with-chargebee"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/recurring-payments"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/saas-accounting-and-taxes"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/payment-gateways"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/payment-gateways/stripe"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/saas-reporting"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/enterprise-subscription-billing"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/features"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/compare-competitors/recurly"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/recurring-payments/direct-debit-payments"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/for-subscription-finance-operations"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/customer-retention/churn-management"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/subscription-management/customer-self-service-portal"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/for-sales-driven-saas"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/compare-competitors/zuora"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/campaigns/saas-subscription-billing"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/security"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/saas-billing"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/partners"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/recurring-payments/payment-methods"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/payment-gateways/braintree"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/compare-competitors/chargify"},
		}, []M.QueryEventWithProperties{
			M.QueryEventWithProperties{Name: "$form_submitted", Properties: []M.QueryProperty{
				M.QueryProperty{
					Entity:    M.PropertyEntityEvent,
					LogicalOp: "AND",
					Operator:  M.EqualsOpStr,
					Property:  "$page_url",
					Type:      U.PropertyTypeCategorical,
					Value:     "www.chargebee.com/trial-signup",
				},
				M.QueryProperty{
					Entity:    M.PropertyEntityEvent,
					LogicalOp: "OR",
					Operator:  M.EqualsOpStr,
					Property:  "$page_url",
					Type:      U.PropertyTypeCategorical,
					Value:     "www.chargebee.com/trial-signup/",
				},
			}},
		}
}

func get14EventsJourney() ([]M.QueryEventWithProperties, []M.QueryEventWithProperties) {
	return []M.QueryEventWithProperties{
			M.QueryEventWithProperties{Name: "www.chargebee.com"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/pricing"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/stripe-recurring-payments-chargebee"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/subscription-management"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/recurring-billing-invoicing"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/recurring-payments-with-chargebee"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/subscription-management-with-chargebee"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/launch"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/customers"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/saas-pricing-models"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/for-self-service-subscription-business"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/integrations"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/subscription-management/create-manage-plans"},
		}, []M.QueryEventWithProperties{
			M.QueryEventWithProperties{Name: "$form_submitted", Properties: []M.QueryProperty{
				M.QueryProperty{
					Entity:    M.PropertyEntityEvent,
					LogicalOp: "AND",
					Operator:  M.EqualsOpStr,
					Property:  "$page_url",
					Type:      U.PropertyTypeCategorical,
					Value:     "www.chargebee.com/schedule-a-demo",
				},
				M.QueryProperty{
					Entity:    M.PropertyEntityEvent,
					LogicalOp: "OR",
					Operator:  M.EqualsOpStr,
					Property:  "$page_url",
					Type:      U.PropertyTypeCategorical,
					Value:     "www.chargebee.com/schedule-a-demo/",
				},
			}},
		}
}

func get15EventsScheduleDemoJourney() ([]M.QueryEventWithProperties, []M.QueryEventWithProperties) {
	return []M.QueryEventWithProperties{
			M.QueryEventWithProperties{Name: "www.chargebee.com"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/stripe-recurring-payments-chargebee"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/subscription-management-with-chargebee"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/pricing"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/trial-signup"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/recurring-billing-invoicing"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/schedule-a-demo"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/launch"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/subscription-management"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/recurring-payments"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/payment-gateways"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/customers"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/saas-pricing-models"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/subscription-management/create-manage-plans"},
		}, []M.QueryEventWithProperties{
			M.QueryEventWithProperties{Name: "$form_submitted", Properties: []M.QueryProperty{
				M.QueryProperty{
					Entity:    M.PropertyEntityEvent,
					LogicalOp: "AND",
					Operator:  M.EqualsOpStr,
					Property:  "$page_url",
					Type:      U.PropertyTypeCategorical,
					Value:     "www.chargebee.com/schedule-a-demo",
				},
				M.QueryProperty{
					Entity:    M.PropertyEntityEvent,
					LogicalOp: "OR",
					Operator:  M.EqualsOpStr,
					Property:  "$page_url",
					Type:      U.PropertyTypeCategorical,
					Value:     "www.chargebee.com/schedule-a-demo/",
				},
			}},
		}
}

func get15EventsTrialSignupJourney() ([]M.QueryEventWithProperties, []M.QueryEventWithProperties) {
	return []M.QueryEventWithProperties{
			M.QueryEventWithProperties{Name: "www.chargebee.com"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/stripe-recurring-payments-chargebee"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/subscription-management-with-chargebee"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/pricing"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/trial-signup"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/recurring-billing-invoicing"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/schedule-a-demo"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/launch"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/subscription-management"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/recurring-payments"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/payment-gateways"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/customers"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/saas-pricing-models"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/subscription-management/create-manage-plans"},
		}, []M.QueryEventWithProperties{
			M.QueryEventWithProperties{Name: "$form_submitted", Properties: []M.QueryProperty{
				M.QueryProperty{
					Entity:    M.PropertyEntityEvent,
					LogicalOp: "AND",
					Operator:  M.EqualsOpStr,
					Property:  "$page_url",
					Type:      U.PropertyTypeCategorical,
					Value:     "www.chargebee.com/trial-signup",
				},
				M.QueryProperty{
					Entity:    M.PropertyEntityEvent,
					LogicalOp: "OR",
					Operator:  M.EqualsOpStr,
					Property:  "$page_url",
					Type:      U.PropertyTypeCategorical,
					Value:     "www.chargebee.com/trial-signup/",
				},
			}},
		}
}

func get5EventScheduledDemoJourney() ([]M.QueryEventWithProperties, []M.QueryEventWithProperties) {
	return []M.QueryEventWithProperties{
			M.QueryEventWithProperties{Name: "www.chargebee.com/trial-signup"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/schedule-a-demo"},
			M.QueryEventWithProperties{Name: "$form_submitted", Properties: []M.QueryProperty{
				M.QueryProperty{
					Entity:    M.PropertyEntityEvent,
					LogicalOp: "AND",
					Operator:  M.EqualsOpStr,
					Property:  "$page_url",
					Type:      U.PropertyTypeCategorical,
					Value:     "www.chargebee.com/trial-signup",
				},
				M.QueryProperty{
					Entity:    M.PropertyEntityEvent,
					LogicalOp: "OR",
					Operator:  M.EqualsOpStr,
					Property:  "$page_url",
					Type:      U.PropertyTypeCategorical,
					Value:     "www.chargebee.com/trial-signup/",
				},
			}},
		}, []M.QueryEventWithProperties{
			M.QueryEventWithProperties{Name: "$form_submitted", Properties: []M.QueryProperty{
				M.QueryProperty{
					Entity:    M.PropertyEntityEvent,
					LogicalOp: "AND",
					Operator:  M.EqualsOpStr,
					Property:  "$page_url",
					Type:      U.PropertyTypeCategorical,
					Value:     "www.chargebee.com/schedule-a-demo",
				},
				M.QueryProperty{
					Entity:    M.PropertyEntityEvent,
					LogicalOp: "OR",
					Operator:  M.EqualsOpStr,
					Property:  "$page_url",
					Type:      U.PropertyTypeCategorical,
					Value:     "www.chargebee.com/schedule-a-demo/",
				},
			}},
		}
}

func get5EventTrialSignupJourney() ([]M.QueryEventWithProperties, []M.QueryEventWithProperties) {
	return []M.QueryEventWithProperties{
			M.QueryEventWithProperties{Name: "www.chargebee.com/trial-signup"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/schedule-a-demo"},
			M.QueryEventWithProperties{Name: "$form_submitted", Properties: []M.QueryProperty{
				M.QueryProperty{
					Entity:    M.PropertyEntityEvent,
					LogicalOp: "AND",
					Operator:  M.EqualsOpStr,
					Property:  "$page_url",
					Type:      U.PropertyTypeCategorical,
					Value:     "www.chargebee.com/schedule-a-demo",
				},
				M.QueryProperty{
					Entity:    M.PropertyEntityEvent,
					LogicalOp: "OR",
					Operator:  M.EqualsOpStr,
					Property:  "$page_url",
					Type:      U.PropertyTypeCategorical,
					Value:     "www.chargebee.com/schedule-a-demo/",
				},
			}},
		}, []M.QueryEventWithProperties{
			M.QueryEventWithProperties{Name: "$form_submitted", Properties: []M.QueryProperty{
				M.QueryProperty{
					Entity:    M.PropertyEntityEvent,
					LogicalOp: "AND",
					Operator:  M.EqualsOpStr,
					Property:  "$page_url",
					Type:      U.PropertyTypeCategorical,
					Value:     "www.chargebee.com/trial-signup",
				},
				M.QueryProperty{
					Entity:    M.PropertyEntityEvent,
					LogicalOp: "OR",
					Operator:  M.EqualsOpStr,
					Property:  "$page_url",
					Type:      U.PropertyTypeCategorical,
					Value:     "www.chargebee.com/trial-signup/",
				},
			}},
		}
}

func get5EventNotDoneScheduledDemoJourney() ([]M.QueryEventWithProperties, []M.QueryEventWithProperties) {
	return []M.QueryEventWithProperties{
			M.QueryEventWithProperties{Name: "www.chargebee.com/trial-signup"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/schedule-a-demo"},
		}, []M.QueryEventWithProperties{
			M.QueryEventWithProperties{Name: "$form_submitted", Properties: []M.QueryProperty{
				M.QueryProperty{
					Entity:    M.PropertyEntityEvent,
					LogicalOp: "AND",
					Operator:  M.EqualsOpStr,
					Property:  "$page_url",
					Type:      U.PropertyTypeCategorical,
					Value:     "www.chargebee.com/trial-signup",
				},
				M.QueryProperty{
					Entity:    M.PropertyEntityEvent,
					LogicalOp: "OR",
					Operator:  M.EqualsOpStr,
					Property:  "$page_url",
					Type:      U.PropertyTypeCategorical,
					Value:     "www.chargebee.com/trial-signup/",
				},
			}},
			M.QueryEventWithProperties{Name: "$form_submitted", Properties: []M.QueryProperty{
				M.QueryProperty{
					Entity:    M.PropertyEntityEvent,
					LogicalOp: "AND",
					Operator:  M.EqualsOpStr,
					Property:  "$page_url",
					Type:      U.PropertyTypeCategorical,
					Value:     "www.chargebee.com/schedule-a-demo",
				},
				M.QueryProperty{
					Entity:    M.PropertyEntityEvent,
					LogicalOp: "OR",
					Operator:  M.EqualsOpStr,
					Property:  "$page_url",
					Type:      U.PropertyTypeCategorical,
					Value:     "www.chargebee.com/schedule-a-demo/",
				},
			}},
		}
}

func get10EventHippoVideoJourney() ([]M.QueryEventWithProperties, []M.QueryEventWithProperties) {
	return []M.QueryEventWithProperties{
			M.QueryEventWithProperties{Name: "www.hippovideo.io/hippo-video-editor.html"},
			M.QueryEventWithProperties{Name: "www.hippovideo.io/video-for-sales-and-prospecting.html"},
			M.QueryEventWithProperties{Name: "www.hippovideo.io/edit-videos-online.html"},
			M.QueryEventWithProperties{Name: "www.hippovideo.io/pricing.html"},
			M.QueryEventWithProperties{Name: "www.hippovideo.io/create-videos-online.html"},
			M.QueryEventWithProperties{Name: "www.hippovideo.io/video-editing-software.html"},
			M.QueryEventWithProperties{Name: "www.hippovideo.io/video-for-sales.html"},
			M.QueryEventWithProperties{Name: "www.hippovideo.io/video-personalization.html"},
		}, []M.QueryEventWithProperties{
			M.QueryEventWithProperties{Name: "$form_submitted", Properties: []M.QueryProperty{
				M.QueryProperty{
					Entity:    M.PropertyEntityEvent,
					LogicalOp: "AND",
					Operator:  M.EqualsOpStr,
					Property:  "$page_url",
					Type:      U.PropertyTypeCategorical,
					Value:     "www.hippovideo.io/users/sign_up",
				},
				M.QueryProperty{
					Entity:    M.PropertyEntityEvent,
					LogicalOp: "OR",
					Operator:  M.EqualsOpStr,
					Property:  "$page_url",
					Type:      U.PropertyTypeCategorical,
					Value:     "www.hippovideo.io/users/sign_up/",
				},
			}},
		}
}

func get6EventHippoVideoJourney() ([]M.QueryEventWithProperties, []M.QueryEventWithProperties) {
	return []M.QueryEventWithProperties{
			M.QueryEventWithProperties{Name: "www.hippovideo.io"},
			M.QueryEventWithProperties{Name: "www.hippovideo.io/video-for-sales-and-prospecting.html"},
			M.QueryEventWithProperties{Name: "www.hippovideo.io/pricing.html"},
			M.QueryEventWithProperties{Name: "www.hippovideo.io/video-for-sales.html"},
			M.QueryEventWithProperties{Name: "www.hippovideo.io/video-personalization.html"},
		}, []M.QueryEventWithProperties{
			M.QueryEventWithProperties{Name: "$form_submitted", Properties: []M.QueryProperty{
				M.QueryProperty{
					Entity:    M.PropertyEntityEvent,
					LogicalOp: "AND",
					Operator:  M.NotEqualOpStr,
					Property:  "$page_url",
					Type:      U.PropertyTypeCategorical,
					Value:     "www.hippovideo.io/users/sign_in",
				},
				M.QueryProperty{
					Entity:    M.PropertyEntityEvent,
					LogicalOp: "AND",
					Operator:  M.NotEqualOpStr,
					Property:  "$page_url",
					Type:      U.PropertyTypeCategorical,
					Value:     "www.hippovideo.io/users/sign_in/",
				},
			}},
		}
}

func get3EventLivspaceJourney() ([]M.QueryEventWithProperties, []M.QueryEventWithProperties) {
	return []M.QueryEventWithProperties{
			M.QueryEventWithProperties{Name: "stage2.livspace.com/in/lead-quiz"},
		}, []M.QueryEventWithProperties{
			M.QueryEventWithProperties{Name: "stage2_lead-quiz:language_next-button-click"},
		}
}

func getSegmentAGoals() []M.QueryEventWithProperties {
	return []M.QueryEventWithProperties{
		M.QueryEventWithProperties{Name: "$form_submitted", Properties: []M.QueryProperty{
			M.QueryProperty{
				Entity:    M.PropertyEntityEvent,
				LogicalOp: "AND",
				Operator:  M.EqualsOpStr,
				Property:  "$page_url",
				Type:      U.PropertyTypeCategorical,
				Value:     "www.chargebee.com/trial-signup",
			},
			M.QueryProperty{
				Entity:    M.PropertyEntityEvent,
				LogicalOp: "OR",
				Operator:  M.EqualsOpStr,
				Property:  "$page_url",
				Type:      U.PropertyTypeCategorical,
				Value:     "www.chargebee.com/trial-signup/",
			},
		}},
	}
}

func getSegmentBGoals() []M.QueryEventWithProperties {
	return []M.QueryEventWithProperties{
		M.QueryEventWithProperties{Name: "$form_submitted", Properties: []M.QueryProperty{
			M.QueryProperty{
				Entity:    M.PropertyEntityEvent,
				LogicalOp: "AND",
				Operator:  "contains",
				Property:  "$page_url",
				Type:      U.PropertyTypeCategorical,
				Value:     "www.chargebee.com/schedule-a-demo",
			},
			M.QueryProperty{
				Entity:    M.PropertyEntityEvent,
				LogicalOp: "OR",
				Operator:  "contains",
				Property:  "$page_url",
				Type:      U.PropertyTypeCategorical,
				Value:     "www.chargebee.com/subscription-management-with-chargebee",
			},
			M.QueryProperty{
				Entity:    M.PropertyEntityEvent,
				LogicalOp: "OR",
				Operator:  "contains",
				Property:  "$page_url",
				Type:      U.PropertyTypeCategorical,
				Value:     "www.chargebee.com/recurring-billing-with-chargebee",
			},
			M.QueryProperty{
				Entity:    M.PropertyEntityEvent,
				LogicalOp: "OR",
				Operator:  "contains",
				Property:  "$page_url",
				Type:      U.PropertyTypeCategorical,
				Value:     "www.chargebee.com/recurring-payments-with-chargebee",
			},
			M.QueryProperty{
				Entity:    M.PropertyEntityEvent,
				LogicalOp: "OR",
				Operator:  "contains",
				Property:  "$page_url",
				Type:      U.PropertyTypeCategorical,
				Value:     "www.chargebee.com/stripe-recurring-payments-chargebee/#take-a-tour",
			},
		}},
	}
}

func getSegmentCGoals() []M.QueryEventWithProperties {
	return []M.QueryEventWithProperties{
		M.QueryEventWithProperties{Name: "$session", Properties: []M.QueryProperty{
			M.QueryProperty{
				Entity:    M.PropertyEntityUser,
				LogicalOp: "AND",
				Operator:  M.EqualsOpStr,
				Property:  "$hubspot_company_contact_hs_band",
				Type:      U.PropertyTypeCategorical,
				Value:     "$none",
			},
		}},
	}
}

func getSegmentVisitedPagesEvents() []M.QueryEventWithProperties {
	return []M.QueryEventWithProperties{
		M.QueryEventWithProperties{Name: "www.chargebee.com/pricing"},
		M.QueryEventWithProperties{Name: "www.chargebee.com/for-education"},
		M.QueryEventWithProperties{Name: "www.chargebee.com/for-self-service-subscription-business"},
		M.QueryEventWithProperties{Name: "www.chargebee.com/subscription-management"},
		M.QueryEventWithProperties{Name: "www.chargebee.com/recurring-billing-invoicing"},
		M.QueryEventWithProperties{Name: "www.chargebee.com/recurring-payments"},
		M.QueryEventWithProperties{Name: "www.chargebee.com/payment-gateways"},
		M.QueryEventWithProperties{Name: "www.chargebee.com/customers"},
		M.QueryEventWithProperties{Name: "www.chargebee.com/saas-pricing-models"},
		M.QueryEventWithProperties{Name: "www.chargebee.com/subscription-management/create-manage-plans"},
		M.QueryEventWithProperties{Name: "www.chargebee.com/trial-signup"},
		M.QueryEventWithProperties{Name: "www.chargebee.com/schedule-a-demo"},
	}
}

func getChargebeeHubspotContactUpdatedJourney() ([]M.QueryEventWithProperties, []M.QueryEventWithProperties) {
	return []M.QueryEventWithProperties{
			M.QueryEventWithProperties{Name: "$hubspot_contact_updated"},
		}, []M.QueryEventWithProperties{
			M.QueryEventWithProperties{Name: "$hubspot_contact_updated", Properties: []M.QueryProperty{
				M.QueryProperty{
					Entity:    M.PropertyEntityEvent,
					LogicalOp: "AND",
					Operator:  M.EqualsOpStr,
					Property:  "$hubspot_contact_lifecyclestage",
					Type:      U.PropertyTypeCategorical,
					Value:     "customer",
				},
			}},
		}
}

func getChargebeeSept22Journey() ([]M.QueryEventWithProperties, []M.QueryEventWithProperties) {
	return []M.QueryEventWithProperties{
			M.QueryEventWithProperties{Name: "$hubspot_contact_updated", Properties: []M.QueryProperty{
				M.QueryProperty{
					Entity:    M.PropertyEntityEvent,
					LogicalOp: "AND",
					Operator:  M.EqualsOpStr,
					Property:  "$hubspot_contact_marketing_conversions_mode",
					Type:      U.PropertyTypeCategorical,
					Value:     "SignUp",
				},
				M.QueryProperty{
					Entity:    M.PropertyEntityEvent,
					LogicalOp: "OR",
					Operator:  "contains",
					Property:  "$hubspot_contact_marketing_conversions_mode",
					Type:      U.PropertyTypeCategorical,
					Value:     "Drift",
				},
				M.QueryProperty{
					Entity:    M.PropertyEntityEvent,
					LogicalOp: "OR",
					Operator:  "contains",
					Property:  "$hubspot_contact_marketing_conversions_mode",
					Type:      U.PropertyTypeCategorical,
					Value:     "Calendly",
				},
			}},
			M.QueryEventWithProperties{Name: "www.chargebee.com/pricing"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/trial-signup"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/schedule-a-demo"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/subscription-management"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/recurring-billing-invoicing"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/recurring-payments"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/saas-accounting-and-taxes"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/saas-reporting"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/integrations"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/payment-gateways"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/for-education"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/for-self-service-subscription-busines"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/for-sales-driven-saas"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/for-subscription-finance-operations"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/enterprise-subscription-billing"},
			M.QueryEventWithProperties{Name: "www.chargebee.com/customers"},
		}, []M.QueryEventWithProperties{
			M.QueryEventWithProperties{Name: "$hubspot_contact_updated", Properties: []M.QueryProperty{
				M.QueryProperty{
					Entity:    M.PropertyEntityEvent,
					LogicalOp: "AND",
					Operator:  M.NotEqualOpStr,
					Property:  "$hubspot_contact_demo_booked_on",
					Type:      U.PropertyTypeCategorical,
					Value:     "$none",
				},
			}},
		}
}

func getObserveAISept24Journey() ([]M.QueryEventWithProperties, []M.QueryEventWithProperties) {
	return []M.QueryEventWithProperties{
			M.QueryEventWithProperties{Name: "$hubspot_contact_updated", Properties: []M.QueryProperty{
				M.QueryProperty{
					Entity:    M.PropertyEntityEvent,
					LogicalOp: "AND",
					Operator:  M.EqualsOpStr,
					Property:  "$hubspot_contact_lifecyclestage",
					Type:      U.PropertyTypeCategorical,
					Value:     "lead",
				},
			}},
			M.QueryEventWithProperties{Name: "$hubspot_contact_updated", Properties: []M.QueryProperty{
				M.QueryProperty{
					Entity:    M.PropertyEntityEvent,
					LogicalOp: "AND",
					Operator:  M.EqualsOpStr,
					Property:  "$hubspot_contact_lifecyclestage",
					Type:      U.PropertyTypeCategorical,
					Value:     "marketingqualifiedlead",
				},
			}},
			M.QueryEventWithProperties{Name: "$hubspot_contact_updated", Properties: []M.QueryProperty{
				M.QueryProperty{
					Entity:    M.PropertyEntityEvent,
					LogicalOp: "AND",
					Operator:  M.EqualsOpStr,
					Property:  "$hubspot_contact_lifecyclestage",
					Type:      U.PropertyTypeCategorical,
					Value:     "salesqualifiedlead",
				},
			}},
			M.QueryEventWithProperties{Name: "$hubspot_contact_updated", Properties: []M.QueryProperty{
				M.QueryProperty{
					Entity:    M.PropertyEntityEvent,
					LogicalOp: "AND",
					Operator:  M.EqualsOpStr,
					Property:  "$hubspot_contact_lifecyclestage",
					Type:      U.PropertyTypeCategorical,
					Value:     "customer",
				},
			}},
			M.QueryEventWithProperties{Name: "$hubspot_contact_updated", Properties: []M.QueryProperty{
				M.QueryProperty{
					Entity:    M.PropertyEntityEvent,
					LogicalOp: "AND",
					Operator:  M.EqualsOpStr,
					Property:  "$hubspot_contact_lifecyclestage",
					Type:      U.PropertyTypeCategorical,
					Value:     "other",
				},
			}},
			M.QueryEventWithProperties{Name: "$hubspot_contact_updated", Properties: []M.QueryProperty{
				M.QueryProperty{
					Entity:    M.PropertyEntityEvent,
					LogicalOp: "AND",
					Operator:  M.EqualsOpStr,
					Property:  "$hubspot_contact_lifecyclestage",
					Type:      U.PropertyTypeCategorical,
					Value:     "$none",
				},
			}},
			M.QueryEventWithProperties{Name: "$hubspot_contact_updated", Properties: []M.QueryProperty{
				M.QueryProperty{
					Entity:    M.PropertyEntityEvent,
					LogicalOp: "AND",
					Operator:  M.EqualsOpStr,
					Property:  "$hubspot_contact_salesforceopportunitystage",
					Type:      U.PropertyTypeCategorical,
					Value:     "Qualification ( SS-3)",
				},
				M.QueryProperty{
					Entity:    M.PropertyEntityEvent,
					LogicalOp: "OR",
					Operator:  M.EqualsOpStr,
					Property:  "$hubspot_contact_salesforceopportunitystage",
					Type:      U.PropertyTypeCategorical,
					Value:     "SS-3 Qualification",
				},
			}},
			M.QueryEventWithProperties{Name: "$hubspot_contact_updated", Properties: []M.QueryProperty{
				M.QueryProperty{
					Entity:    M.PropertyEntityEvent,
					LogicalOp: "AND",
					Operator:  M.EqualsOpStr,
					Property:  "$hubspot_contact_salesforceopportunitystage",
					Type:      U.PropertyTypeCategorical,
					Value:     "SS-4 Opportunity Validation",
				},
			}},
			M.QueryEventWithProperties{Name: "$hubspot_contact_updated", Properties: []M.QueryProperty{
				M.QueryProperty{
					Entity:    M.PropertyEntityEvent,
					LogicalOp: "AND",
					Operator:  M.EqualsOpStr,
					Property:  "$hubspot_contact_salesforceopportunitystage",
					Type:      U.PropertyTypeCategorical,
					Value:     "SS-5 Need Confirmed",
				},
			}},
			M.QueryEventWithProperties{Name: "$hubspot_contact_updated", Properties: []M.QueryProperty{
				M.QueryProperty{
					Entity:    M.PropertyEntityEvent,
					LogicalOp: "AND",
					Operator:  M.EqualsOpStr,
					Property:  "$hubspot_contact_salesforceopportunitystage",
					Type:      U.PropertyTypeCategorical,
					Value:     "SS-6 Alignment & Poc",
				},
			}},
			M.QueryEventWithProperties{Name: "$hubspot_contact_updated", Properties: []M.QueryProperty{
				M.QueryProperty{
					Entity:    M.PropertyEntityEvent,
					LogicalOp: "AND",
					Operator:  M.EqualsOpStr,
					Property:  "$hubspot_contact_salesforceopportunitystage",
					Type:      U.PropertyTypeCategorical,
					Value:     "SS-7 Pov Complete & Shortlist",
				},
			}},
			M.QueryEventWithProperties{Name: "$hubspot_contact_updated", Properties: []M.QueryProperty{
				M.QueryProperty{
					Entity:    M.PropertyEntityEvent,
					LogicalOp: "AND",
					Operator:  M.EqualsOpStr,
					Property:  "$hubspot_contact_salesforceopportunitystage",
					Type:      U.PropertyTypeCategorical,
					Value:     "SS-8 Verbal & Terms",
				},
			}},
			M.QueryEventWithProperties{Name: "$hubspot_contact_updated", Properties: []M.QueryProperty{
				M.QueryProperty{
					Entity:    M.PropertyEntityEvent,
					LogicalOp: "AND",
					Operator:  M.EqualsOpStr,
					Property:  "$hubspot_contact_salesforceopportunitystage",
					Type:      U.PropertyTypeCategorical,
					Value:     "SS-12 Closed Lost",
				},
			}},
		}, []M.QueryEventWithProperties{
			M.QueryEventWithProperties{Name: "$hubspot_contact_updated", Properties: []M.QueryProperty{
				M.QueryProperty{
					Entity:    M.PropertyEntityEvent,
					LogicalOp: "AND",
					Operator:  M.EqualsOpStr,
					Property:  "$hubspot_contact_salesforceopportunitystage",
					Type:      U.PropertyTypeCategorical,
					Value:     "SS-10 Closed Won",
				},
				M.QueryProperty{
					Entity:    M.PropertyEntityEvent,
					LogicalOp: "OR",
					Operator:  M.EqualsOpStr,
					Property:  "$hubspot_contact_salesforceopportunitystage",
					Type:      U.PropertyTypeCategorical,
					Value:     "SS-11 Closed Won",
				},
			}},
		}
}
