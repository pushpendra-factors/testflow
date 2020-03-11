package main

import (
	"flag"

	C "factors/config"
	H "factors/handler"
	U "factors/util"

	log "github.com/sirupsen/logrus"
)

const workerName = "sdk_request_worker"

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

	awsRegion := flag.String("aws_region", "us-east-1", "")
	awsAccessKeyId := flag.String("aws_key", "dummy", "")
	awsSecretAccessKey := flag.String("aws_secret", "dummy", "")

	factorsEmailSender := flag.String("email_sender", "support-dev@factors.ai", "")
	errorReportingInterval := flag.Int("error_reporting_interval", 300, "")

	workerConcurrency := flag.Int("worker_concurrency", 10, "")

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
		RedisHost:              *redisHost,
		RedisPort:              *redisPort,
		QueueRedisHost:         *queueRedisHost,
		QueueRedisPort:         *queueRedisPort,
		GeolocationFile:        *geoLocFilePath,
		AWSKey:                 *awsAccessKeyId,
		AWSSecret:              *awsSecretAccessKey,
		AWSRegion:              *awsRegion,
		EmailSender:            *factorsEmailSender,
		ErrorReportingInterval: *errorReportingInterval,
	}

	err := C.InitQueueWorker(config)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize.")
		return
	}

	// Register tasks on queueClient.
	queueClient := C.GetServices().SDKQueueClient
	if err := queueClient.RegisterTask(H.SDKProcessRequestTask,
		H.SDKProcessQueueRequest); err != nil {

		log.WithError(err).Fatal(
			"Failed to register tasks on SDK queue client.")
	}

	// Todo(Dinesh): Add pod_id to worker name.
	worker := queueClient.NewCustomQueueWorker(
		workerName, *workerConcurrency, H.SDKRequestQueue)
	worker.Launch()
}
