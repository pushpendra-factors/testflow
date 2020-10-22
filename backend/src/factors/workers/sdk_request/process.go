package main

import (
	"flag"

	C "factors/config"
	SDK "factors/sdk"
	U "factors/util"

	log "github.com/sirupsen/logrus"
)

const workerName = "sdk_request_worker"

func main() {
	env := flag.String("env", "development", "")
	gcpProjectID := flag.String("gcp_project_id", "", "Project ID on Google Cloud")
	gcpProjectLocation := flag.String("gcp_project_location", "", "Location of google cloud project cluster")

	dbHost := flag.String("db_host", "localhost", "")
	dbPort := flag.Int("db_port", 5432, "")
	dbUser := flag.String("db_user", "autometa", "")
	dbName := flag.String("db_name", "autometa", "")
	dbPass := flag.String("db_pass", "@ut0me7a", "")

	redisHost := flag.String("redis_host", "localhost", "")
	redisPort := flag.Int("redis_port", 6379, "")

	queueRedisHost := flag.String("queue_redis_host", "localhost", "")
	queueRedisPort := flag.Int("queue_redis_port", 6379, "")

	geoLocFilePath := flag.String("geo_loc_path",
		"/usr/local/var/factors/geolocation_data/GeoLite2-City.mmdb", "")
	deviceDetectorPath := flag.String("device_detector_path", "/usr/local/var/factors/devicedetector_data/regexes", "")

	sentryDSN := flag.String("sentry_dsn", "", "Sentry DSN")

	skipSessionProjectIds := flag.String("skip_session_project_ids",
		"", "List or projects to create session offline.")

	workerConcurrency := flag.Int("worker_concurrency", 10, "")
	redisHostPersistent := flag.String("redis_host_ps", "localhost", "")
	redisPortPersistent := flag.Int("redis_port_ps", 6379, "")
	isRealTimeEventUserCachingEnabled := flag.Bool("enable_real_time_event_user_caching", false, "If the real time caching is enabled")
	realTimeEventUserCachingProjectIds := flag.String("real_time_event_user_caching_project_ids", "", "If the real time caching is enabled and the whitelisted projectids")

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
		},
		RedisHost:                          *redisHost,
		RedisPort:                          *redisPort,
		QueueRedisHost:                     *queueRedisHost,
		QueueRedisPort:                     *queueRedisPort,
		GeolocationFile:                    *geoLocFilePath,
		DeviceDetectorPath:                 *deviceDetectorPath,
		SentryDSN:                          *sentryDSN,
		SkipSessionProjectIds:              *skipSessionProjectIds, // comma seperated project ids, supports "*".
		RedisHostPersistent:                *redisHostPersistent,
		RedisPortPersistent:                *redisPortPersistent,
		IsRealTimeEventUserCachingEnabled:  *isRealTimeEventUserCachingEnabled,
		RealTimeEventUserCachingProjectIds: *realTimeEventUserCachingProjectIds,
	}

	err := C.InitQueueWorker(config)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize.")
		return
	}
	defer C.SafeFlushAllCollectors()

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
