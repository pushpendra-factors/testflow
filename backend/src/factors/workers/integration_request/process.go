package main

import (
	"flag"
	"net/http"

	log "github.com/sirupsen/logrus"

	C "factors/config"
	Int "factors/integration"
	IntSegment "factors/integration/segment"
	U "factors/util"
	"factors/vendor_custom/machinery/v1"
)

const defaultWorkerName = "integration_request_worker"
const duplicateWorkerName = "duplicate_integration_request_worker"

func ProcessRequest(token, reqType, reqPayload string) (float64, string, error) {
	switch reqType {
	case Int.TypeSegment, Int.TypeRudderstack:
		return IntSegment.ProcessQueueEvent(token, reqPayload)
	case Int.TypeShopify:
		// Todo: Add shopify request process method.
	}

	log.WithField("req_type", reqType).Error(
		"Unknown request type on process integration request.")
	return http.StatusBadRequest, "", nil
}

func main() {
	env := flag.String("env", "development", "")
	gcpProjectID := flag.String("gcp_project_id", "", "Project ID on Google Cloud")
	gcpProjectLocation := flag.String("gcp_project_location", "", "Location of google cloud project cluster")

	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	isPSCHost := flag.Int("memsql_is_psc_host", C.MemSQLDefaultDBParams.IsPSCHost, "")
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

	deviceDetectorPath := flag.String("device_detector_path",
		"/usr/local/var/factors/devicedetector_data/regexes", "")

	sentryDSN := flag.String("sentry_dsn", "", "Sentry DSN")
	sentryRollupSyncInSecs := flag.Int("sentry_rollup_sync_in_seconds", 300, "Enables to send errors to sentry in given interval.")

	workerConcurrency := flag.Int("worker_concurrency", 10, "")
	redisHostPersistent := flag.String("redis_host_ps", "localhost", "")
	redisPortPersistent := flag.Int("redis_port_ps", 6379, "")
	propertiesTypeCacheSize := flag.Int("property_details_cache_size", 0, "Cache size for in memory property detail.")
	enablePropertyTypeFromDB := flag.Bool("enable_property_type_from_db", false, "Enable property type check from db.")
	whitelistedProjectIDPropertyTypeFromDB := flag.String("whitelisted_project_ids_property_type_check_from_db", "", "Allowed project id for property type check from db.")
	blacklistedProjectIDPropertyTypeFromDB := flag.String("blacklisted_project_ids_property_type_check_from_db", "", "Blocked project id for property type check from db.")
	cacheSortedSet := flag.Bool("cache_with_sorted_set", false, "Cache with sorted set keys")
	enableSDKAndIntegrationRequestQueueDuplication := flag.Bool("enable_sdk_and_integration_request_queue_duplication",
		false, "Enables SDK and Integration request queue duplication.")
	segmentExcludedCustomerUserIDByProject := flag.String("segment_excluded_customer_user_ids", "",
		"Map of project_id and customer_user_id to exclude identification on segment.")
	allowSupportForUserPropertiesInIdentifyCall := flag.String("allow_support_for_user_properties_in_identify_call", "", "")
	captureSourceInUsersTable := flag.String("capture_source_in_users_table", "", "")
	enableOLTPQueriesMemSQLImprovements := flag.String("enable_OLTP_queries_memsql_improvements", "", "")
	restrictReusingUsersByCustomerUserId := flag.String("restrict_reusing_users_by_customer_user_id", "", "")
	mergeAmpIDAndSegmentIDWithUserIDByProjectID := flag.String("allow_amp_id_and_segment_id_with_user_id_by_project_id", "", "")
	clearbitEnabled := flag.Int("clearbit_enabled", 0, "To enable clearbit enrichment")
	sixSignalEnabled := flag.Int("six_signal_enabled", 0, "To enable sixSignal enrichment")
	allowIdentificationOverwriteUsingSourceByProjectID := flag.String("allow_identification_overwrite_using_source_by_project_id", "", "Allow identification overwrite based on request source.")
	IngestionTimezoneEnabledProjectIDs := flag.String("ingestion_timezone_enabled_projects", "", "List of projectIds whose ingestion timezone is enabled.")
	enableSixSignalGroupByProjectID := flag.String("enable_six_signal_group_by_project_id", "", "")
	enableDomainsGroupByProjectID := flag.String("enable_domains_group_by_project_id", "", "")
	enableUserDomainsGroupByProjectID := flag.String("enable_user_domains_group_by_project_id", "", "Allow domains group for users")
	allowEmailDomainsByProjectID := flag.String("allow_email_domain_by_project_id", "", "Allow email domains for domain group")

	flag.Parse()

	workerName := defaultWorkerName
	if *enableSDKAndIntegrationRequestQueueDuplication {
		workerName = duplicateWorkerName
	}

	defer U.NotifyOnPanic(workerName, *env)

	config := &C.Configuration{
		AppName:                 workerName,
		Env:                     *env,
		GCPProjectID:            *gcpProjectID,
		GCPProjectLocation:      *gcpProjectLocation,
		RedisHost:               *redisHost,
		RedisPort:               *redisPort,
		QueueRedisHost:          *queueRedisHost,
		QueueRedisPort:          *queueRedisPort,
		GeolocationFile:         *geoLocFilePath,
		DeviceDetectorPath:      *deviceDetectorPath,
		SentryDSN:               *sentryDSN,
		SentryRollupSyncInSecs:  *sentryRollupSyncInSecs,
		RedisHostPersistent:     *redisHostPersistent,
		RedisPortPersistent:     *redisPortPersistent,
		DuplicateQueueRedisHost: *duplicateQueueRedisHost,
		DuplicateQueueRedisPort: *duplicateQueueRedisPort,
		EnableSDKAndIntegrationRequestQueueDuplication: *enableSDKAndIntegrationRequestQueueDuplication,
		MemSQLInfo: C.DBConf{
			Host:        *memSQLHost,
			IsPSCHost:   *isPSCHost,
			Port:        *memSQLPort,
			User:        *memSQLUser,
			Name:        *memSQLName,
			Password:    *memSQLPass,
			Certificate: *memSQLCertificate,
			AppName:     workerName,
		},
		CacheSortedSet:   *cacheSortedSet,
		PrimaryDatastore: *primaryDatastore,
		SegmentExcludedCustomerIDByProject: C.ParseProjectIDToStringMapFromConfig(
			*segmentExcludedCustomerUserIDByProject, "SegmentExcludedCustomerIDByProject"),
		AllowSupportForUserPropertiesInIdentifyCall:        *allowSupportForUserPropertiesInIdentifyCall,
		CaptureSourceInUsersTable:                          *captureSourceInUsersTable,
		EnableOLTPQueriesMemSQLImprovements:                *enableOLTPQueriesMemSQLImprovements,
		RestrictReusingUsersByCustomerUserId:               *restrictReusingUsersByCustomerUserId,
		MergeAmpIDAndSegmentIDWithUserIDByProjectID:        *mergeAmpIDAndSegmentIDWithUserIDByProjectID,
		ClearbitEnabled:                                    *clearbitEnabled,
		SixSignalEnabled:                                   *sixSignalEnabled,
		AllowIdentificationOverwriteUsingSourceByProjectID: *allowIdentificationOverwriteUsingSourceByProjectID,
		IngestionTimezoneEnabledProjectIDs:                 C.GetTokensFromStringListAsString(*IngestionTimezoneEnabledProjectIDs),
		EnableSixSignalGroupByProjectID:                    *enableSixSignalGroupByProjectID,
		EnableDomainsGroupByProjectID:                      *enableDomainsGroupByProjectID,
		EnableUserDomainsGroupByProjectID:                  *enableUserDomainsGroupByProjectID,
		AllowEmailDomainsByProjectID:                       *allowEmailDomainsByProjectID,
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
		queueName = Int.RequestQueueDuplicate
	} else {
		queueClient = C.GetServices().QueueClient
		queueName = Int.RequestQueue
	}

	// Register tasks on queueClient.
	err = queueClient.RegisterTask(Int.ProcessRequestTask, ProcessRequest)
	if err != nil {
		log.WithError(err).WithField("worker", workerName).
			Fatal("Failed to register tasks on queue client in integration request worker.")
	}

	worker := queueClient.NewCustomQueueWorker(workerName, *workerConcurrency, queueName)
	worker.Launch()
}
