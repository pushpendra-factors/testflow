package main

import (
	C "factors/config"
	"flag"
	"fmt"
	"time"

	cleanup "factors/task/event_user_cache"

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

	eventsLimit := flag.Int("events_limit", 10000, "")
	propertiesLimit := flag.Int("properties_limit", 100000, "")
	valuesLimit := flag.Int("values_limit", 100000, "")
	// This is in days
	rollupLookback := flag.Int("rollup_lookback", 1, "")

	sentryDSN := flag.String("sentry_dsn", "", "Sentry DSN")
	gcpProjectID := flag.String("gcp_project_id", "", "Project ID on Google Cloud")
	gcpProjectLocation := flag.String("gcp_project_location", "", "Location of google cloud project cluster")

	flag.Parse()
	if *env != "development" &&
		*env != "staging" &&
		*env != "production" {
		err := fmt.Errorf("env [ %s ] not recognised", *env)
		panic(err)
	}

	taskID := "cleanup_sortedset_cache"
	healthcheckPingID := C.HealthcheckCleanupEventUserCachePingID
	defer C.PingHealthcheckForPanic(taskID, *env, healthcheckPingID)

	config := &C.Configuration{
		AppName:            taskID,
		Env:                *env,
		GCPProjectID:       *gcpProjectID,
		GCPProjectLocation: *gcpProjectLocation,
		DBInfo: C.DBConf{
			Host:     *dbHost,
			Port:     *dbPort,
			User:     *dbUser,
			Name:     *dbName,
			Password: *dbPass,
			AppName:  taskID,
		},
		RedisHostPersistent: *redisHostPersistent,
		RedisPortPersistent: *RedisPortPersistent,
		SentryDSN:           *sentryDSN,
	}

	C.InitConf(config.Env)

	// Cache dependency for requests not using queue.
	C.InitRedisPersistent(config.RedisHostPersistent, config.RedisPortPersistent)

	C.InitSentryLogging(config.SentryDSN, config.AppName)
	C.InitMetricsExporter(config.Env, config.AppName, config.GCPProjectID, config.GCPProjectLocation)
	defer C.WaitAndFlushAllCollectors(65 * time.Second)

	status := cleanup.DoCleanUpSortedSet(eventsLimit, propertiesLimit, valuesLimit, rollupLookback)

	log.Info("Done!!!")
	C.PingHealthcheckForSuccess(healthcheckPingID, status)
}
