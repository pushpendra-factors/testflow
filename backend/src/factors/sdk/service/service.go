package main

import (
	C "factors/config"
	H "factors/handler"

	"flag"
	"strconv"

	mid "factors/middleware"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	swaggerDocs "factors/sdk/docs"
)

// @title Factors SDK Service Backend Api
// @version 1.0
// @description Factors usage doc for SDK service.
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {
	gcpProjectID := flag.String("gcp_project_id", "", "Project ID on Google Cloud")
	gcpProjectLocation := flag.String("gcp_project_location", "", "Location of google cloud project cluster")

	env := flag.String("env", C.DEVELOPMENT, "")
	port := flag.Int("port", 8085, "")
	// port := flag.Int("port", 8089, "")

	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	isPSCHost := flag.Int("memsql_is_psc_host", C.MemSQLDefaultDBParams.IsPSCHost, "")
	memSQLPort := flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser := flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName := flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass := flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	memSQLCertificate := flag.String("memsql_cert", "", "")
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypeMemSQL, "Primary datastore type as memsql or postgres")

	geoLocFilePath := flag.String("geo_loc_path",
		"/usr/local/var/factors/geolocation_data/GeoLite2-City.mmdb", "")

	deviceDetectorPath := flag.String("device_detector_path", "/usr/local/var/factors/devicedetector_data/regexes", "")

	redisHost := flag.String("redis_host", "localhost", "")
	redisPort := flag.Int("redis_port", 6379, "")

	factorsSixsignalAPIKey := flag.String("factors_sixsignal_api_key", "dummy", "")

	queueRedisHost := flag.String("queue_redis_host", "localhost", "")
	queueRedisPort := flag.Int("queue_redis_port", 6379, "")
	useQueueRedis := flag.Bool("use_queue_redis", false, "Use queue redis for caching.")

	duplicateQueueRedisHost := flag.String("dup_queue_redis_host", "localhost", "")
	duplicateQueueRedisPort := flag.Int("dup_queue_redis_port", 6379, "")

	redisHostPersistent := flag.String("redis_host_ps", "localhost", "")
	redisPortPersistent := flag.Int("redis_port_ps", 6379, "")

	sentryDSN := flag.String("sentry_dsn", "", "Sentry DSN")
	// Rollup default intentionally set to 1minute considering criticality.
	sentryRollupSyncInSecs := flag.Int("sentry_rollup_sync_in_seconds", 60, "Enables to send errors to sentry in given interval.")

	sdkRequestQueueProjectTokens := flag.String("sdk_request_queue_project_tokens", "",
		"List of project tokens allowed to use sdk request queue")
	segmentRequestQueueProjectTokens := flag.String("segment_request_queue_project_tokens", "",
		"List of project tokens allowed to use segment request queue")

	useDefaultProjectSettingForSDK := flag.Bool("use_defaul_project_setting_for_sdk",
		false, "Once set to true, it will skip db query to get project_settings, if not found on cache.")

	blockedSDKRequestProjectTokens := flag.String("blocked_sdk_request_project_tokens",
		"", "List of tokens (public and private) to block SDK requests.")

	enableSDKAndIntegrationRequestQueueDuplication := flag.Bool("enable_sdk_and_integration_request_queue_duplication",
		false, "Enables SDK and Integration request queue duplication.")
	propertiesTypeCacheSize := flag.Int("property_details_cache_size", 0, "Cache size for in memory property detail.")
	enablePropertyTypeFromDB := flag.Bool("enable_property_type_from_db", false, "Enable property type check from db.")
	whitelistedProjectIDPropertyTypeFromDB := flag.String("whitelisted_project_ids_property_type_check_from_db", "", "Allowed project id for property type check from db.")
	blacklistedProjectIDPropertyTypeFromDB := flag.String("blacklisted_project_ids_property_type_check_from_db", "", "Blocked project id for property type check from db.")
	cacheSortedSet := flag.Bool("cache_with_sorted_set", false, "Cache with sorted set keys")
	allowSupportForUserPropertiesInIdentifyCall := flag.String("allow_support_for_user_properties_in_identify_call", "", "")
	enableDebuggingForIP := flag.Bool("enable_debugging_for_ip", false, "Enables log for $ip and other properties added by $ip")

	blockedIPProjectTokens := flag.String("blocked_ip_project_tokens",
		"", "List of tokens to enable feature of IP based blocking for all sdk request types.")

	excludeBotIPV4AddressByRange := flag.String("exclude_bot_ip_by_range",
		"", "CIDR ranges for excluding bot traffic.")

	deviceServiceUrl := flag.String("device_service_url", "http://0.0.0.0:3000/device_service", "URL for the device detection service")
	enableDeviceServiceByProjectID := flag.String("enable_device_service_by_project_id", "", "")

	flag.Parse()

	appName := "sdk_server"
	config := &C.Configuration{
		GCPProjectID:       *gcpProjectID,
		GCPProjectLocation: *gcpProjectLocation,
		AppName:            appName,
		Env:                *env,
		Port:               *port,
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
		GeolocationFile:                  *geoLocFilePath,
		DeviceDetectorPath:               *deviceDetectorPath,
		RedisHost:                        *redisHost,
		RedisPort:                        *redisPort,
		QueueRedisHost:                   *queueRedisHost,
		QueueRedisPort:                   *queueRedisPort,
		FactorsSixSignalAPIKey:           *factorsSixsignalAPIKey,
		UseQueueRedis:                    *useQueueRedis,
		PrimaryDatastore:                 *primaryDatastore,
		SDKRequestQueueProjectTokens:     C.GetTokensFromStringListAsString(*sdkRequestQueueProjectTokens), // comma seperated project tokens.
		SegmentRequestQueueProjectTokens: C.GetTokensFromStringListAsString(*segmentRequestQueueProjectTokens),
		RedisHostPersistent:              *redisHostPersistent,
		RedisPortPersistent:              *redisPortPersistent,
		UseDefaultProjectSettingForSDK:   *useDefaultProjectSettingForSDK,
		// List of tokens (public and private) to block SDK requests.
		BlockedSDKRequestProjectTokens:                 C.GetTokensFromStringListAsString(*blockedSDKRequestProjectTokens),
		EnableSDKAndIntegrationRequestQueueDuplication: *enableSDKAndIntegrationRequestQueueDuplication,
		CacheSortedSet:                              *cacheSortedSet,
		DuplicateQueueRedisHost:                     *duplicateQueueRedisHost,
		DuplicateQueueRedisPort:                     *duplicateQueueRedisPort,
		SentryDSN:                                   *sentryDSN,
		SentryRollupSyncInSecs:                      *sentryRollupSyncInSecs,
		DeviceServiceURL:                            *deviceServiceUrl,
		EnableDeviceServiceByProjectID:              *enableDeviceServiceByProjectID,
		AllowSupportForUserPropertiesInIdentifyCall: *allowSupportForUserPropertiesInIdentifyCall,
		EnableDebuggingForIP:                        *enableDebuggingForIP,
		BlockedIPProjectTokens:                      *blockedIPProjectTokens,
		ExcludeBotIPV4AddressByRange:                *excludeBotIPV4AddressByRange,
	}
	C.InitConf(config)

	err := C.InitSDKService(config)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize.")
		return
	}
	defer C.SafeFlushAllCollectors()

	C.InitPropertiesTypeCache(*enablePropertyTypeFromDB, *propertiesTypeCacheSize,
		*whitelistedProjectIDPropertyTypeFromDB, *blacklistedProjectIDPropertyTypeFromDB)

	if !C.IsDevelopment() {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(mid.CustomCors())
	r.Use(mid.RequestIdGenerator())
	r.Use(mid.Logger())
	r.Use(mid.Recovery())

	// Initialize routes.
	if config.Env == C.DEVELOPMENT {
		swaggerDocs.SwaggerInfo.Host = "factors-dev.com:8085"
	} else if config.Env == C.STAGING {
		swaggerDocs.SwaggerInfo.Host = "staging-api.factors.ai"
	}
	H.InitSDKServiceRoutes(r)
	r.Run(":" + strconv.Itoa(C.GetConfig().Port))
}
