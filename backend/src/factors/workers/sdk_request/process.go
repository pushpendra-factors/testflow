package main

import (
	"flag"

	C "factors/config"
	SDK "factors/sdk"
	U "factors/util"

	log "github.com/sirupsen/logrus"
)

const workerName = "sdk_request_worker"
const duplicateWorkerName = "duplicate_sdk_request_worker"

func main() {
	env := flag.String("env", "development", "")
	gcpProjectID := flag.String("gcp_project_id", "", "Project ID on Google Cloud")
	gcpProjectLocation := flag.String("gcp_project_location", "", "Location of google cloud project cluster")

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
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypePostgres, "Primary datastore type as memsql or postgres")

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
	ontableUserPropertiesWriteAllowedProjectIDs := flag.String("ontable_user_properties_allowed_projects",
		"", "List of projects to enable writing to on-table user_properties column.")
	deprecateUserPropertiesTableWriteProjectIDs := flag.String("deprecate_user_properties_table_write_projects",
		"", "List of projects to stop writing to user_properties table.")
	deprecateUserPropertiesTableReadProjectIDs := flag.String("deprecate_user_properties_table_read_projects",
		"", "List of projects for which user_properties table read to be deprecated.")
	enableSDKAndIntegrationRequestQueueDuplication := flag.Bool("enable_sdk_and_integration_request_queue_duplication",
		false, "Enables SDK and Integration request queue duplication.")
	cacheSortedSet := flag.Bool("cache_with_sorted_set", false, "Cache with sorted set keys")
	flag.Parse()

	defer U.NotifyOnPanic(workerName, *env)

	config := &C.Configuration{
		AppName:            workerName,
		Env:                *env,
		GCPProjectID:       *gcpProjectID,
		GCPProjectLocation: *gcpProjectLocation,
		DBInfo: C.DBConf{
			Host:     *dbHost,
			Port:     *dbPort,
			User:     *dbUser,
			Name:     *dbName,
			Password: *dbPass,
			AppName:  workerName,
		},
		RedisHost:           *redisHost,
		RedisPort:           *redisPort,
		QueueRedisHost:      *queueRedisHost,
		QueueRedisPort:      *queueRedisPort,
		GeolocationFile:     *geoLocFilePath,
		DeviceDetectorPath:  *deviceDetectorPath,
		SentryDSN:           *sentryDSN,
		RedisHostPersistent: *redisHostPersistent,
		RedisPortPersistent: *redisPortPersistent,
		// List of project to enable on-table user_properties write on events and users table.
		OnTableUserPropertiesWriteAllowedProjects: *ontableUserPropertiesWriteAllowedProjectIDs,
		// List of projects to stop writing to user_properties table.
		DeprecateUserPropertiesTableWriteProjects: *deprecateUserPropertiesTableWriteProjectIDs,
		// List of projects to use on-table user_properties for read.
		DeprecateUserPropertiesTableReadProjects: *deprecateUserPropertiesTableReadProjectIDs,
		CacheSortedSet:                           *cacheSortedSet,
		MemSQLInfo: C.DBConf{
			Host:     *memSQLHost,
			Port:     *memSQLPort,
			User:     *memSQLUser,
			Name:     *memSQLName,
			Password: *memSQLPass,
			AppName:  workerName,
		},
		PrimaryDatastore:                               *primaryDatastore,
		DuplicateQueueRedisHost:                        *duplicateQueueRedisHost,
		DuplicateQueueRedisPort:                        *duplicateQueueRedisPort,
		EnableSDKAndIntegrationRequestQueueDuplication: *enableSDKAndIntegrationRequestQueueDuplication,
	}

	err := C.InitQueueWorker(config)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize.")
		return
	}
	defer C.SafeFlushAllCollectors()

	C.InitPropertiesTypeCache(*enablePropertyTypeFromDB, *propertiesTypeCacheSize,
		*whitelistedProjectIDPropertyTypeFromDB, *blacklistedProjectIDPropertyTypeFromDB)

	// Start worker with duplicate queue client, if enabled.
	if C.IsSDKAndIntegrationRequestQueueDuplicationEnabled() {
		// Register tasks on queueClient.
		duplicateQueueClient := C.GetServices().DuplicateQueueClient
		if err := duplicateQueueClient.RegisterTask(SDK.ProcessRequestTask,
			SDK.ProcessQueueRequest); err != nil {

			log.WithError(err).Fatal(
				"Failed to register tasks on duplicate queue client in sdk_request_worker.")
		}

		duplicateQueueWorker := duplicateQueueClient.NewCustomQueueWorker(
			duplicateWorkerName, *workerConcurrency, SDK.RequestQueueDuplicate)
		duplicateQueueWorker.Launch()
	}

	// Register tasks on queueClient.
	queueClient := C.GetServices().QueueClient
	if err := queueClient.RegisterTask(SDK.ProcessRequestTask,
		SDK.ProcessQueueRequest); err != nil {

		log.WithError(err).Fatal(
			"Failed to register tasks on queue client in sdk_request_worker.")
	}

	// Todo(Dinesh): Add pod_id to worker name.
	worker := queueClient.NewCustomQueueWorker(
		workerName, *workerConcurrency, SDK.RequestQueue)
	worker.Launch()
}
