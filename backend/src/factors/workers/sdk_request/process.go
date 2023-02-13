package main

import (
	"flag"

	C "factors/config"
	SDK "factors/sdk"
	U "factors/util"
	"factors/vendor_custom/machinery/v1"

	log "github.com/sirupsen/logrus"
)

const defaultWorkerName = "sdk_request_worker"
const duplicateWorkerName = "duplicate_sdk_request_worker"

func main() {
	env := flag.String("env", "development", "")
	gcpProjectID := flag.String("gcp_project_id", "", "Project ID on Google Cloud")
	gcpProjectLocation := flag.String("gcp_project_location", "", "Location of google cloud project cluster")

	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	memSQLPort := flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser := flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName := flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass := flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	memSQLCertificate := flag.String("memsql_cert", "", "")
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypeMemSQL, "Primary datastore type as memsql or postgres")

	redisHost := flag.String("redis_host", "localhost", "")
	redisPort := flag.Int("redis_port", 6379, "")

	queueRedisHost := flag.String("queue_redis_host", "localhost", "")
	queueRedisPort := flag.Int("queue_redis_port", 6379, "")

	duplicateQueueRedisHost := flag.String("dup_queue_redis_host", "localhost", "")
	duplicateQueueRedisPort := flag.Int("dup_queue_redis_port", 6379, "")

	geoLocFilePath := flag.String("geo_loc_path",
		"/usr/local/var/factors/geolocation_data/GeoLite2-City.mmdb", "")
	deviceDetectorPath := flag.String("device_detector_path", "/usr/local/var/factors/devicedetector_data/regexes", "")

	sentryDSN := flag.String("sentry_dsn", "", "Sentry DSN")

	workerConcurrency := flag.Int("worker_concurrency", 10, "")
	redisHostPersistent := flag.String("redis_host_ps", "localhost", "")
	redisPortPersistent := flag.Int("redis_port_ps", 6379, "")
	propertiesTypeCacheSize := flag.Int("property_details_cache_size", 0, "Cache size for in memory property detail.")
	enablePropertyTypeFromDB := flag.Bool("enable_property_type_from_db", false, "Enable property type check from db.")
	whitelistedProjectIDPropertyTypeFromDB := flag.String("whitelisted_project_ids_property_type_check_from_db", "", "Allowed project id for property type check from db.")
	blacklistedProjectIDPropertyTypeFromDB := flag.String("blacklisted_project_ids_property_type_check_from_db", "", "Blocked project id for property type check from db.")
	enableSDKAndIntegrationRequestQueueDuplication := flag.Bool("enable_sdk_and_integration_request_queue_duplication",
		false, "Enables SDK and Integration request queue duplication.")
	cacheSortedSet := flag.Bool("cache_with_sorted_set", false, "Cache with sorted set keys")
	allowSupportForUserPropertiesInIdentifyCall := flag.String("allow_support_for_user_properties_in_identify_call", "", "")
	captureSourceInUsersTable := flag.String("capture_source_in_users_table", "", "")
	enableOLTPQueriesMemSQLImprovements := flag.String("enable_OLTP_queries_memsql_improvements", "", "")
	restrictReusingUsersByCustomerUserId := flag.String("restrict_reusing_users_by_customer_user_id", "", "")
	mergeAmpIDAndSegmentIDWithUserIDByProjectID := flag.String("allow_amp_id_and_segment_id_with_user_id_by_project_id", "", "")
	clearbitEnabled := flag.Int("clearbit_enabled", 0, "To enable clearbit enrichment")
	sixSignalEnabled := flag.Int("six_signal_enabled", 0, "To enable sixSignal enrichment")
	allowIdentificationOverwriteUsingSourceByProjectID := flag.String("allow_identification_overwrite_using_source_by_project_id", "", "Allow identification overwrite based on request source.")
	IngestionTimezoneEnabledProjectIDs := flag.String("ingestion_timezone_enabled_projects", "", "List of projectIds whose ingestion timezone is enabled.")
	formFillIdentifyAllowedProjectIDs := flag.String("form_fill_identify_allowed_projects", "", "Form fill identification allowed project ids.")
	eventTriggerEnabled := flag.Bool("event_trigger_enabled", false, "")
	eventTriggerEnabledProjectIDs := flag.String("event_trigger_enabled_project_ids", "", "")
	enableSixSignalGroupByProjectID := flag.String("enable_six_signal_group_by_project_id", "", "")
	flag.Parse()

	workerName := defaultWorkerName
	if *enableSDKAndIntegrationRequestQueueDuplication {
		workerName = duplicateWorkerName
	}

	defer U.NotifyOnPanic(workerName, *env)

	config := &C.Configuration{
		AppName:             workerName,
		Env:                 *env,
		GCPProjectID:        *gcpProjectID,
		GCPProjectLocation:  *gcpProjectLocation,
		RedisHost:           *redisHost,
		RedisPort:           *redisPort,
		QueueRedisHost:      *queueRedisHost,
		QueueRedisPort:      *queueRedisPort,
		GeolocationFile:     *geoLocFilePath,
		DeviceDetectorPath:  *deviceDetectorPath,
		SentryDSN:           *sentryDSN,
		RedisHostPersistent: *redisHostPersistent,
		RedisPortPersistent: *redisPortPersistent,
		CacheSortedSet:      *cacheSortedSet,
		MemSQLInfo: C.DBConf{
			Host:        *memSQLHost,
			Port:        *memSQLPort,
			User:        *memSQLUser,
			Name:        *memSQLName,
			Password:    *memSQLPass,
			Certificate: *memSQLCertificate,
			AppName:     workerName,
		},
		PrimaryDatastore:                                   *primaryDatastore,
		DuplicateQueueRedisHost:                            *duplicateQueueRedisHost,
		DuplicateQueueRedisPort:                            *duplicateQueueRedisPort,
		EnableSDKAndIntegrationRequestQueueDuplication:     *enableSDKAndIntegrationRequestQueueDuplication,
		AllowSupportForUserPropertiesInIdentifyCall:        *allowSupportForUserPropertiesInIdentifyCall,
		CaptureSourceInUsersTable:                          *captureSourceInUsersTable,
		EnableOLTPQueriesMemSQLImprovements:                *enableOLTPQueriesMemSQLImprovements,
		RestrictReusingUsersByCustomerUserId:               *restrictReusingUsersByCustomerUserId,
		MergeAmpIDAndSegmentIDWithUserIDByProjectID:        *mergeAmpIDAndSegmentIDWithUserIDByProjectID,
		ClearbitEnabled:                                    *clearbitEnabled,
		SixSignalEnabled:                                   *sixSignalEnabled,
		AllowIdentificationOverwriteUsingSourceByProjectID: *allowIdentificationOverwriteUsingSourceByProjectID,
		IngestionTimezoneEnabledProjectIDs:                 C.GetTokensFromStringListAsString(*IngestionTimezoneEnabledProjectIDs),
		FormFillIdentificationAllowedProjects:              *formFillIdentifyAllowedProjectIDs,
		EventTriggerEnabled:                                *eventTriggerEnabled,
		EventTriggerEnabledProjectIDs:                      *eventTriggerEnabledProjectIDs,
		EnableSixSignalGroupByProjectID:                    *enableSixSignalGroupByProjectID,
	}
	C.InitConf(config)

	err := C.InitQueueWorker(config, *workerConcurrency)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize.")
		return
	}
	defer C.SafeFlushAllCollectors()

	C.InitPropertiesTypeCache(*enablePropertyTypeFromDB, *propertiesTypeCacheSize,
		*whitelistedProjectIDPropertyTypeFromDB, *blacklistedProjectIDPropertyTypeFromDB)

	var queueClient *machinery.Server
	var queueName string
	if C.IsSDKAndIntegrationRequestQueueDuplicationEnabled() {
		queueClient = C.GetServices().DuplicateQueueClient
		queueName = SDK.RequestQueueDuplicate
	} else {
		queueClient = C.GetServices().QueueClient
		queueName = SDK.RequestQueue
	}

	// Register tasks on queueClient.
	if err := queueClient.RegisterTask(SDK.ProcessRequestTask,
		SDK.ProcessQueueRequest); err != nil {

		log.WithError(err).WithField("worker", workerName).
			Fatal("Failed to register tasks on queue client in sdk_request_worker.")
	}

	worker := queueClient.NewCustomQueueWorker(workerName, *workerConcurrency, queueName)
	worker.Launch()
}
