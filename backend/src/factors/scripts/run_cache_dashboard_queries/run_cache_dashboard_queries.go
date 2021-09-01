package main

import (
	"flag"
	"fmt"
	"sync"
	"time"

	C "factors/config"
	"factors/model/store"
	"factors/util"

	log "github.com/sirupsen/logrus"
)

func main() {
	envFlag := flag.String("env", C.DEVELOPMENT, "Environment. Could be development|staging|production")
	projectIDFlag := flag.String("project_id", "", "Comma separated project ids to run for. * to run for all")
	excludeProjectIDFlag := flag.String("exclude_project_id", "", "Comma separated project ids to exclude for the run")
	numRoutinesFlag := flag.Int("num_routines", 4, "Number of dashboard units to sync in parallel. Each dashboard unit runs 4 queries")
	numRoutinesForWebAnalyticsFlag := flag.Int("num_routines_for_web_analytics", 1,
		"No.of routines to use for web analytics dashboard caching.")

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
	memSQLResourcePool := flag.String("memsql_resource_pool", "", "If provided, all the queries will run under the given resource pool")
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypePostgres, "Primary datastore type as memsql or postgres")

	sentryDSN := flag.String("sentry_dsn", "", "Sentry DSN")
	onlyWebAnalytics := flag.Bool("only_web_analytics", false, "Cache only web analytics dashboards.")

	redisHost := flag.String("redis_host", "localhost", "")
	redisPort := flag.Int("redis_port", 6379, "")

	gcpProjectID := flag.String("gcp_project_id", "", "Project ID on Google Cloud")
	gcpProjectLocation := flag.String("gcp_project_location", "", "Location of google cloud project cluster")
	enableMemSQLRedisWrites := flag.Bool("enable_mql_redis_writes", false, "To enable redis writes when using MemSQL")

	flag.Parse()
	disableMemSQLRedisWrites := !(*enableMemSQLRedisWrites)
	taskID := "dashboard_caching"
	healthcheckPingID := C.HealthcheckDashboardCachingPingID
	defer C.PingHealthcheckForPanic(taskID, *envFlag, healthcheckPingID)
	logCtx := log.WithFields(log.Fields{"Prefix": taskID})

	if *envFlag != C.DEVELOPMENT && *envFlag != C.STAGING && *envFlag != C.PRODUCTION {
		panic(fmt.Errorf("env [ %s ] not recognised", *envFlag))
	} else if *projectIDFlag == "" {
		panic(fmt.Errorf("Invalid project id %s", *projectIDFlag))
	} else if *numRoutinesFlag == 0 {
		panic(fmt.Errorf("Num routines must at least be 1"))
	}

	logCtx.Info("Starting to initialize database.")
	config := &C.Configuration{
		AppName:            taskID,
		Env:                *envFlag,
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
		RedisHost: *redisHost,
		RedisPort: *redisPort,
		SentryDSN: *sentryDSN,
		MemSQLInfo: C.DBConf{
			Host:         *memSQLHost,
			Port:         *memSQLPort,
			User:         *memSQLUser,
			Name:         *memSQLName,
			Password:     *memSQLPass,
			Certificate:  *memSQLCertificate,
			ResourcePool: *memSQLResourcePool,
			AppName:      taskID,
		},
		PrimaryDatastore:         *primaryDatastore,
		DisableMemSQLRedisWrites: &disableMemSQLRedisWrites,
	}

	C.InitConf(config)
	err := C.InitDB(*config)
	if err != nil {
		logCtx.WithError(err).Panic("Failed to initialize DB")
	}
	C.KillDBQueriesOnExit()
	C.InitRedisPersistent(config.RedisHost, config.RedisPort)

	C.InitSentryLogging(config.SentryDSN, config.AppName)
	C.InitMetricsExporter(config.Env, config.AppName, config.GCPProjectID, config.GCPProjectLocation)
	defer C.WaitAndFlushAllCollectors(65 * time.Second)

	logCtx = logCtx.WithFields(log.Fields{
		"Env":              *envFlag,
		"ProjectID":        *projectIDFlag,
		"ExcludeProjectID": *excludeProjectIDFlag,
		"NumRoutines":      *numRoutinesFlag,
	})

	var notifyMessage string
	var waitGroup sync.WaitGroup
	var timeTaken sync.Map

	if !*onlyWebAnalytics {
		waitGroup.Add(1)
		go cacheDashboardUnitsForProjects(*projectIDFlag, *excludeProjectIDFlag, *numRoutinesFlag, &timeTaken, &waitGroup)
	}

	waitGroup.Add(1)
	go cacheWebsiteAnalyticsForProjects(*projectIDFlag, *excludeProjectIDFlag, *numRoutinesForWebAnalyticsFlag, &timeTaken, &waitGroup)

	waitGroup.Wait()
	timeTakenString, _ := timeTaken.Load("all")
	timeTakenStringWeb, _ := timeTaken.Load("web")
	notifyMessage = fmt.Sprintf("Caching successful for %s - %s projects. Time taken: %+v. Time taken for web analytics: %+v",
		*projectIDFlag, *excludeProjectIDFlag, timeTakenString, timeTakenStringWeb)
	C.PingHealthcheckForSuccess(healthcheckPingID, notifyMessage)
}

func cacheDashboardUnitsForProjects(projectIDs, excludeProjectIDs string, numRoutines int, timeTaken *sync.Map, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()
	startTime := util.TimeNowUnix()
	store.GetStore().CacheDashboardUnitsForProjects(projectIDs, excludeProjectIDs, numRoutines)
	timeTakenString := util.SecondsToHMSString(util.TimeNowUnix() - startTime)
	timeTaken.Store("all", timeTakenString)
}

func cacheWebsiteAnalyticsForProjects(projectIDs, excludeProjectIDs string, numRoutines int, timeTaken *sync.Map, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()
	startTime := util.TimeNowUnix()
	store.GetStore().CacheWebsiteAnalyticsForProjects(projectIDs, excludeProjectIDs, numRoutines)
	timeTakenStringWeb := util.SecondsToHMSString(util.TimeNowUnix() - startTime)
	timeTaken.Store("web", timeTakenStringWeb)
}
