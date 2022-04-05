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
	memSQLCertificate := flag.String("memsql_cert", "", "")
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypeMemSQL, "Primary datastore type as memsql or postgres")

	redisHostPersistent := flag.String("redis_host_ps", "localhost", "")
	RedisPortPersistent := flag.Int("redis_port_ps", 6379, "")

	eventsLimit := flag.Int("events_limit", 10000, "")
	propertiesLimit := flag.Int("properties_limit", 100000, "")
	valuesLimit := flag.Int("values_limit", 100000, "")

	sentryDSN := flag.String("sentry_dsn", "", "Sentry DSN")
	gcpProjectID := flag.String("gcp_project_id", "", "Project ID on Google Cloud")
	gcpProjectLocation := flag.String("gcp_project_location", "", "Location of google cloud project cluster")

	overrideHealthcheckPingID := flag.String("healthcheck_ping_id", "", "Override default healthcheck ping id.")
	overrideAppName := flag.String("app_name", "", "Override default app_name.")

	flag.Parse()
	if *env != "development" &&
		*env != "staging" &&
		*env != "production" {
		err := fmt.Errorf("env [ %s ] not recognised", *env)
		panic(err)
	}

	defaultAppName := "cleanup_sortedset_cache"
	defaultHealthcheckPingID := C.HealthcheckCleanupEventUserCachePingID
	healthcheckPingID := C.GetHealthcheckPingID(defaultHealthcheckPingID, *overrideHealthcheckPingID)
	appName := C.GetAppName(defaultAppName, *overrideAppName)

	defer C.PingHealthcheckForPanic(appName, *env, healthcheckPingID)

	config := &C.Configuration{
		AppName:            appName,
		Env:                *env,
		GCPProjectID:       *gcpProjectID,
		GCPProjectLocation: *gcpProjectLocation,
		DBInfo: C.DBConf{
			Host:     *dbHost,
			Port:     *dbPort,
			User:     *dbUser,
			Name:     *dbName,
			Password: *dbPass,
			AppName:  appName,
		},
		RedisHostPersistent: *redisHostPersistent,
		RedisPortPersistent: *RedisPortPersistent,
		SentryDSN:           *sentryDSN,
		MemSQLInfo: C.DBConf{
			Host:        *memSQLHost,
			Port:        *memSQLPort,
			User:        *memSQLUser,
			Name:        *memSQLName,
			Password:    *memSQLPass,
			Certificate: *memSQLCertificate,
			AppName:     appName,
		},
		PrimaryDatastore: *primaryDatastore,
	}

	C.InitConf(config)
	err := C.InitDB(*config)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{"env": *env,
			"host": *dbHost, "port": *dbPort}).Panic("Failed to initialize DB.")
		os.Exit(0)
	}
	// Cache dependency for requests not using queue.
	C.InitRedisPersistent(config.RedisHostPersistent, config.RedisPortPersistent)

	C.InitSentryLogging(config.SentryDSN, config.AppName)
	C.InitMetricsExporter(config.Env, config.AppName, config.GCPProjectID, config.GCPProjectLocation)
	defer C.WaitAndFlushAllCollectors(65 * time.Second)

	configs := make(map[string]interface{})
	configs["eventsLimit"] = *eventsLimit
	configs["propertiesLimit"] = *propertiesLimit
	configs["valuesLimit"] = *valuesLimit

	status := taskWrapper.TaskFunc("CleanUpSortedSet", 1, cleanup.DoCleanUpSortedSet, configs)

	log.Info("Done!!!")
	C.PingHealthcheckForSuccess(healthcheckPingID, status)
}
