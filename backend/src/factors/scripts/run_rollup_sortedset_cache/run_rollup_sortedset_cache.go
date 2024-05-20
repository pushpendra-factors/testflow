package main

import (
	C "factors/config"
	"flag"
	"fmt"
	"os"
	"time"

	cleanup "factors/task/event_user_cache"
	taskWrapper "factors/task/task_wrapper"

	log "github.com/sirupsen/logrus"
)

func main() {

	env := flag.String("env", "development", "")

	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	isPSCHost := flag.Int("memsql_is_psc_host", C.MemSQLDefaultDBParams.IsPSCHost, "")
	memSQLPort := flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser := flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName := flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass := flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	memSQLCertificate := flag.String("memsql_cert", "", "")
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypeMemSQL, "Primary datastore type as memsql or postgres")

	redisHostPersistent := flag.String("redis_host_ps", "localhost", "")
	RedisPortPersistent := flag.Int("redis_port_ps", 6379, "")
	// This is in days
	rollupLookback := flag.Int("rollup_lookback", 1, "")

	// same from app-server
	lookbackWindowForEventUserCache := flag.Int("lookback_window_event_user_cache",
		20, "look back window in cache for event/user cache")
	deleteRollupAfterAddingToAggregate := flag.Int("del_rollup_after_aggregate", 0, "")

	sentryDSN := flag.String("sentry_dsn", "", "Sentry DSN")
	gcpProjectID := flag.String("gcp_project_id", "", "Project ID on Google Cloud")
	gcpProjectLocation := flag.String("gcp_project_location", "", "Location of google cloud project cluster")
	aggrEventPropertyValuesCacheByProjectID := flag.String("aggr_event_property_values_project_ids", "", "")

	enableCacheDBWriteProjects := flag.String("cache_db_write_projects", "", "")
	enableCacheDBReadProjects := flag.String("cache_db_read_projects", "", "")

	flag.Parse()
	if *env != "development" &&
		*env != "staging" &&
		*env != "production" {
		err := fmt.Errorf("env [ %s ] not recognised", *env)
		panic(err)
	}

	taskID := "rollup_sortedset_cache"
	healthcheckPingID := C.HealthcheckCleanupEventUserCachePingID
	defer C.PingHealthcheckForPanic(taskID, *env, healthcheckPingID)

	config := &C.Configuration{
		AppName:             taskID,
		Env:                 *env,
		GCPProjectID:        *gcpProjectID,
		GCPProjectLocation:  *gcpProjectLocation,
		RedisHostPersistent: *redisHostPersistent,
		RedisPortPersistent: *RedisPortPersistent,
		SentryDSN:           *sentryDSN,
		MemSQLInfo: C.DBConf{
			Host:        *memSQLHost,
			IsPSCHost:   *isPSCHost,
			Port:        *memSQLPort,
			User:        *memSQLUser,
			Name:        *memSQLName,
			Password:    *memSQLPass,
			Certificate: *memSQLCertificate,
			AppName:     taskID,
		},
		PrimaryDatastore:                        *primaryDatastore,
		AggrEventPropertyValuesCacheByProjectID: *aggrEventPropertyValuesCacheByProjectID,
		LookbackWindowForEventUserCache:         *lookbackWindowForEventUserCache,
		EnableCacheDBWriteProjects:              *enableCacheDBWriteProjects,
		EnableCacheDBReadProjects:               *enableCacheDBReadProjects,
	}

	C.InitConf(config)
	err := C.InitDB(*config)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{"env": *env,
			"host": *memSQLHost, "port": *memSQLPort}).Panic("Failed to initialize DB.")
		os.Exit(0)
	}
	// Cache dependency for requests not using queue.
	C.InitRedisPersistent(config.RedisHostPersistent, config.RedisPortPersistent)

	C.InitSentryLogging(config.SentryDSN, config.AppName)
	C.InitMetricsExporter(config.Env, config.AppName, config.GCPProjectID, config.GCPProjectLocation)
	defer C.WaitAndFlushAllCollectors(65 * time.Second)

	configs := make(map[string]interface{})
	configs["rollupLookback"] = *rollupLookback

	configs["lookbackWindowForEventUserCache"] = *lookbackWindowForEventUserCache
	configs["deleteRollupAfterAddingToAggregate"] = *deleteRollupAfterAddingToAggregate

	status := taskWrapper.TaskFunc("RollUpSortedSet", 1, cleanup.DoRollUpSortedSet, configs)

	log.Info("Done!!!")
	C.PingHealthcheckForSuccess(healthcheckPingID, status)
}
