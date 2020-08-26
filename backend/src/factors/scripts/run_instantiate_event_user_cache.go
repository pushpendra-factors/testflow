package main

import (
	C "factors/config"
	M "factors/model"
	S "factors/sdk"
	"factors/util"
	U "factors/util"
	"flag"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

func main() {

	env := flag.String("env", "development", "")
	dbHost := flag.String("db_host", "localhost", "")
	dbPort := flag.Int("db_port", 5432, "")
	dbUser := flag.String("db_user", "autometa", "")
	dbName := flag.String("db_name", "autometa", "")
	dbPass := flag.String("db_pass", "@ut0me7a", "")

	redisHostPersistent := flag.String("redis_host_ps", "localhost", "")
	RedisPortPersistent := flag.Int("redis_port_ps", 6379, "")

	awsRegion := flag.String("aws_region", "us-east-1", "")
	awsAccessKeyId := flag.String("aws_key", "dummy", "")
	awsSecretAccessKey := flag.String("aws_secret", "dummy", "")
	factorsEmailSender := flag.String("email_sender", "support-dev@factors.ai", "")
	errorReportingInterval := flag.Int("error_reporting_interval", 300, "")

	projectIds := flag.String("project_ids", "", "Projects for which the cache is to be refreshed")
	eventRecordsLimit := flag.Int("event_records_limit", 100000, "")
	usersProcessedLimit := flag.Int("users_processed_limit", 10000, "")
	eventsLimit := flag.Int("events_limit", 10000, "")
	propertiesLimit := flag.Int("properties_limit", 5000, "")
	valuesLimit := flag.Int("values_limit", 2000, "")
	lookBackDays := flag.Int("look_back_days", 30, "")

	sentryDSN := flag.String("sentry_dsn", "", "Sentry DSN")

	flag.Parse()
	if *env != "development" &&
		*env != "staging" &&
		*env != "production" {
		err := fmt.Errorf("env [ %s ] not recognised", *env)
		panic(err)
	}

	taskID := "Task#InstantiateEventUserCache"
	defer util.NotifyOnPanic(taskID, *env)

	config := &C.Configuration{
		AppName: "instantiate_event_user_cacheß",
		Env:     *env,
		DBInfo: C.DBConf{
			Host:     *dbHost,
			Port:     *dbPort,
			User:     *dbUser,
			Name:     *dbName,
			Password: *dbPass,
		},
		RedisHostPersistent:    *redisHostPersistent,
		RedisPortPersistent:    *RedisPortPersistent,
		AWSKey:                 *awsAccessKeyId,
		AWSSecret:              *awsSecretAccessKey,
		AWSRegion:              *awsRegion,
		EmailSender:            *factorsEmailSender,
		ErrorReportingInterval: *errorReportingInterval,
		SentryDSN:              *sentryDSN,
	}

	C.InitConf(config.Env)

	// Will allow all 50/50 connection to be idle on the pool.
	// As we allow num_routines (per project) as per no.of db connections
	// and will be used continiously.
	err := C.InitDBWithMaxIdleAndMaxOpenConn(config.DBInfo, 50, 50)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize db in add session.")
	}

	// Cache dependency for requests not using queue.
	C.InitRedisPersistent(config.RedisHostPersistent, config.RedisPortPersistent)

	C.InitLogClient(config.Env, config.AppName, config.EmailSender, config.AWSKey,
		config.AWSSecret, config.AWSRegion, config.ErrorReportingInterval, config.SentryDSN)
	currentTime := U.TimeNow()
	startOfCurrentDay := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), 0, 0, 0, 0, time.UTC)

	projectIdMap := util.GetIntBoolMapFromStringList(projectIds)

	for projectId, _ := range projectIdMap {
		S.RefreshCacheFromDb(projectId, startOfCurrentDay, *lookBackDays, *eventsLimit, *propertiesLimit, *valuesLimit, *eventRecordsLimit)
		M.RefreshCacheForUserProperties(projectId, startOfCurrentDay, *usersProcessedLimit, *propertiesLimit, *valuesLimit)
	}

}
