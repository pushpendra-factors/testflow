package main

import (
	C "factors/config"
	U "factors/util"
	"flag"
	"fmt"

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

	flag.Parse()
	if *env != "development" &&
		*env != "staging" &&
		*env != "production" {
		err := fmt.Errorf("env [ %s ] not recognised", *env)
		panic(err)
	}

	taskID := "Task#CleanUpEventUserCache"
	defer U.NotifyOnPanic(taskID, *env)

	config := &C.Configuration{
		AppName: "CleanUpEventUserCache",
		Env:     *env,
		DBInfo: C.DBConf{
			Host:     *dbHost,
			Port:     *dbPort,
			User:     *dbUser,
			Name:     *dbName,
			Password: *dbPass,
		},
		RedisHostPersistent: *redisHostPersistent,
		RedisPortPersistent: *RedisPortPersistent,
		SentryDSN:           *sentryDSN,
	}

	C.InitConf(config.Env)

	// Cache dependency for requests not using queue.
	C.InitRedisPersistent(config.RedisHostPersistent, config.RedisPortPersistent)

	defer C.SafeFlushAllCollectors()

	status := cleanup.DoRollUpAndCleanUp(eventsLimit, propertiesLimit, valuesLimit, rollupLookback)

	if err := U.NotifyThroughSNS(taskID, *env, status); err != nil {
		log.Fatalf("Failed to notify status %+v", status)
	}
	log.Info("Done!!!")

}
