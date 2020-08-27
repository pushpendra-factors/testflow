package main

import (
	"flag"
	"net/http"

	log "github.com/sirupsen/logrus"

	C "factors/config"
	Int "factors/integration"
	IntSegment "factors/integration/segment"
	U "factors/util"
)

const workerName = "integration_request_worker"

func ProcessRequest(token, reqType, reqPayload string) (float64, string, error) {
	switch reqType {
	case Int.TypeSegment:
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

	deviceDetectorPath := flag.String("device_detector_path",
		"/usr/local/var/factors/devicedetector_data/regexes", "")

	awsRegion := flag.String("aws_region", "us-east-1", "")
	awsAccessKeyId := flag.String("aws_key", "dummy", "")
	awsSecretAccessKey := flag.String("aws_secret", "dummy", "")

	sentryDSN := flag.String("sentry_dsn", "", "Sentry DSN")

	factorsEmailSender := flag.String("email_sender", "support-dev@factors.ai", "")
	errorReportingInterval := flag.Int("error_reporting_interval", 300, "")

	skipSessionProjectIds := flag.String("skip_session_project_ids",
		"", "List or projects to create session offline.")
	mergeUserPropertiesProjectIDS := flag.String("merge_usp_project_ids", "",
		"Comma separated list of project IDs for which user properties merge is enabled. '*' for all.")

	workerConcurrency := flag.Int("worker_concurrency", 10, "")
	redisHostPersistent := flag.String("redis_host_ps", "localhost", "")
	redisPortPersistent := flag.Int("redis_port_ps", 6379, "")
	isRealTimeCachingEnabled := flag.Bool("is_real_time_caching_enabled", false, "If the real time caching is enabled")
	flag.Parse()

	defer U.NotifyOnPanic(workerName, *env)

	config := &C.Configuration{
		AppName: workerName,
		Env:     *env,
		DBInfo: C.DBConf{
			Host:     *dbHost,
			Port:     *dbPort,
			User:     *dbUser,
			Name:     *dbName,
			Password: *dbPass,
		},
		RedisHost:                *redisHost,
		RedisPort:                *redisPort,
		QueueRedisHost:           *queueRedisHost,
		QueueRedisPort:           *queueRedisPort,
		GeolocationFile:          *geoLocFilePath,
		DeviceDetectorPath:       *deviceDetectorPath,
		AWSKey:                   *awsAccessKeyId,
		AWSSecret:                *awsSecretAccessKey,
		AWSRegion:                *awsRegion,
		SentryDSN:                *sentryDSN,
		EmailSender:              *factorsEmailSender,
		ErrorReportingInterval:   *errorReportingInterval,
		SkipSessionProjectIds:    *skipSessionProjectIds, // comma seperated project ids, supports "*".
		MergeUspProjectIds:       *mergeUserPropertiesProjectIDS,
		RedisHostPersistent:      *redisHostPersistent,
		RedisPortPersistent:      *redisPortPersistent,
		IsRealTimeCachingEnabled: *isRealTimeCachingEnabled,
	}

	err := C.InitQueueWorker(config)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize.")
		return
	}

	// Register tasks on queueClient.
	queueClient := C.GetServices().QueueClient
	err = queueClient.RegisterTask(Int.ProcessRequestTask, ProcessRequest)
	if err != nil {
		log.WithError(err).Fatal(
			"Failed to register tasks on queue client in integration request worker.")
	}

	// Todo(Dinesh): Add pod_id to worker name.
	worker := queueClient.NewCustomQueueWorker(workerName,
		*workerConcurrency, Int.RequestQueue)
	worker.Launch()
}
