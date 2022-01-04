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
	numRoutinesFlag := flag.Int("num_routines", 12, "Number of dashboard units to sync in parallel. Each dashboard unit runs 4 queries")
	numRoutinesForWebAnalyticsFlag := flag.Int("num_routines_for_web_analytics", 2,
		"No.of routines to use for web analytics dashboard caching.")
	numMonthsFlag := flag.Int("num_months", 12, "Number of previous months to be backfilled")

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
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypePostgres, "Primary datastore type as memsql or postgres")

	runningForMemsql := flag.Int("running_for_memsql", 0, "Disable routines for memsql.")

	sentryDSN := flag.String("sentry_dsn", "", "Sentry DSN")

	redisHost := flag.String("redis_host", "localhost", "")
	redisPort := flag.Int("redis_port", 6379, "")

	enableFilterOptimisation := flag.Bool("enable_filter_optimisation", false,
		"Enables filter optimisation changes for memsql implementation.")
	filterPropertiesStartTimestamp := flag.Int64("filter_properties_start_timestamp", -1,
		"Start timestamp of data available for filtering with parquet on memsql.")
	skipEventNameStepByProjectID := flag.String("skip_event_name_step_by_project_id", "", "")
	skipUserJoinInEventQueryByProjectID := flag.String("skip_user_join_in_event_query_by_project_id", "", "")

	flag.Parse()

	taskID := "monthly_dashboard_caching"
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
		AppName: taskID,
		Env:     *envFlag,
		DBInfo: C.DBConf{
			Host:     *dbHost,
			Port:     *dbPort,
			User:     *dbUser,
			Name:     *dbName,
			Password: *dbPass,
			AppName:  taskID,
		},
		MemSQLInfo: C.DBConf{
			Host:        *memSQLHost,
			Port:        *memSQLPort,
			User:        *memSQLUser,
			Name:        *memSQLName,
			Password:    *memSQLPass,
			Certificate: *memSQLCertificate,
			AppName:     taskID,
		},
		PrimaryDatastore:                    *primaryDatastore,
		RedisHost:                           *redisHost,
		RedisPort:                           *redisPort,
		SentryDSN:                           *sentryDSN,
		EnableFilterOptimisation:            *enableFilterOptimisation,
		FilterPropertiesStartTimestamp:      *filterPropertiesStartTimestamp,
		IsRunningForMemsql:                  *runningForMemsql,
		SkipEventNameStepByProjectID:        *skipEventNameStepByProjectID,
		SkipUserJoinInEventQueryByProjectID: *skipUserJoinInEventQueryByProjectID,
	}
	C.InitConf(config)

	err := C.InitDB(*config)
	if err != nil {
		logCtx.WithError(err).Panic("Failed to initialize DB")
	}
	C.KillDBQueriesOnExit()
	C.InitRedisPersistent(config.RedisHost, config.RedisPort)

	C.InitSentryLogging(config.SentryDSN, config.AppName)
	defer C.WaitAndFlushAllCollectors(65 * time.Second)

	logCtx = logCtx.WithFields(log.Fields{
		"Env":              *envFlag,
		"ProjectID":        *projectIDFlag,
		"ExcludeProjectID": *excludeProjectIDFlag,
		"NumRoutines":      *numRoutinesFlag,
	})

	var notifyMessage string
	var waitGroup sync.WaitGroup
	var reportCollector sync.Map

	waitGroup.Add(1)
	go cacheMonthlyDashboardUnitsForProjects(
		*projectIDFlag, *excludeProjectIDFlag, *numMonthsFlag, *numRoutinesFlag, &reportCollector, &reportCollector, &waitGroup)

	waitGroup.Add(1)
	go cacheMonthlyWebsiteAnalyticsForProjects(
		*projectIDFlag, *excludeProjectIDFlag, *numMonthsFlag, *numRoutinesForWebAnalyticsFlag, &reportCollector, &reportCollector, &waitGroup)

	waitGroup.Wait()
	timeTakenString, _ := reportCollector.Load("all")
	timeTakenStringWeb, _ := reportCollector.Load("web")
	notifyMessage = fmt.Sprintf("Caching successful for %s - %s projects. Time taken: %+v. Time taken for web analytics: %+v",
		*projectIDFlag, *excludeProjectIDFlag, timeTakenString, timeTakenStringWeb)
	C.PingHealthcheckForSuccess(healthcheckPingID, notifyMessage)
}

func cacheMonthlyDashboardUnitsForProjects(projectIDs, excludeProjectIDs string, numMonths, numRoutines int, reportCollector *sync.Map,
	timeTaken *sync.Map, waitGroup *sync.WaitGroup) {
	if C.GetIsRunningForMemsql() == 0 {
		defer waitGroup.Done()
	}
	startTime := util.TimeNowUnix()
	store.GetStore().CacheDashboardsForMonthlyRange(projectIDs, excludeProjectIDs, numMonths, numRoutines, reportCollector)
	timeTakenString := util.SecondsToHMSString(util.TimeNowUnix() - startTime)
	timeTaken.Store("all", timeTakenString)
}

func cacheMonthlyWebsiteAnalyticsForProjects(projectIDs, excludeProjectIDs string, numMonths, numRoutines int, reportCollector *sync.Map,
	timeTaken *sync.Map, waitGroup *sync.WaitGroup) {
	if C.GetIsRunningForMemsql() == 0 {
		defer waitGroup.Done()
	}
	startTime := util.TimeNowUnix()
	store.GetStore().CacheWebsiteAnalyticsForMonthlyRange(projectIDs, excludeProjectIDs, numMonths, numRoutines, reportCollector)
	timeTakenStringWeb := util.SecondsToHMSString(util.TimeNowUnix() - startTime)
	timeTaken.Store("web", timeTakenStringWeb)
}
