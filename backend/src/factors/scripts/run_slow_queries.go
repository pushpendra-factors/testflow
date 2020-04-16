package main

import (
	C "factors/config"
	"factors/util"
	"flag"
	"fmt"

	log "github.com/sirupsen/logrus"
)

type SlowQueries struct {
	Runtime []uint8 `json:"runtime"`
	Query   string  `json:"query"`
	Pid     int64   `json:"pid"`
}

const taskID = "Task#SlowQueries"

func GetSlowQueries(env string) (map[string]interface{}, error) {
	slowQueries := make([]SlowQueries, 0, 0)
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

	flag.Parse()

	config := &C.Configuration{
		AppName: "slow_queries",
		Env:     env,
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
		return nil, err
	}

	C.InitConf(config.Env)
	// Initialize configs and connections and close with defer.
	err = C.InitDB(config.DBInfo)
	if err != nil {
		log.Fatal("Failed to run slow queries. Init failed.")
	}

	db := C.GetServices().Db
	defer db.Close()

	queryStr := `SELECT (now() - query_start) as runtime,query, pid FROM  pg_stat_activity` +
		` WHERE (now() - query_start) > '2 minutes'::interval ORDER BY runtime DESC LIMIT 10`
	rows, err := db.Raw(queryStr).Rows()
	if err != nil {
		log.WithError(err).Error("Failed to get slow queries from pg_stat_activity")
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var slowQuery SlowQueries
		if err := db.ScanRows(rows, &slowQuery); err != nil {
			log.WithError(err).Error("Failed to scan slow queries from db.")
			return nil, err
		}

		slowQueries = append(slowQueries, slowQuery)
	}

	queueClient := C.GetServices().QueueClient
	delayedTaskCount, err := queueClient.GetBroker().GetDelayedTasksCount()
	if err != nil {
		log.WithError(err).Error("Failed to get delayed task count from redis")
		return nil, err
	}

	sdkQueueLength, err := queueClient.GetBroker().GetQueueLength("sdk_request_queue")
	if err != nil {
		log.WithError(err).Error("Failed to get sdk_request_queue length")
		return nil, err
	}

	integrationQueueLength, err := queueClient.GetBroker().GetQueueLength("integration_request_queue")
	if err != nil {
		log.WithError(err).Error("Failed to get integration_request_queue length")
		return nil, err
	}

	slowQueriesStatus := map[string]interface{}{
		"slowQueries":            slowQueries,
		"delayedTaskCount":       delayedTaskCount,
		"sdkQueueLength":         sdkQueueLength,
		"integrationQueueLength": integrationQueueLength,
	}
	return slowQueriesStatus, nil
}

func main() {
	envFlag := flag.String("env", "development", "")

	flag.Parse()
	defer util.NotifyOnPanic("Task#SlowQueries", *envFlag)

	if *envFlag != "development" &&
		*envFlag != "staging" &&
		*envFlag != "production" {
		err := fmt.Errorf("env [ %s ] not recognised", *envFlag)
		panic(err)
	}

	//err already logged in function, so suppressed it here
	slowQueriesStatus, _ := GetSlowQueries(*envFlag)

	if *envFlag == "development" {
		log.Info(slowQueriesStatus)
	} else {
		if err := util.NotifyThroughSNS(taskID, *envFlag, slowQueriesStatus); err != nil {
			log.WithError(err).Error("Failed to notify slow queries status.")
		} else {
			log.Info("Notified slow queries status.")
		}
	}
}
