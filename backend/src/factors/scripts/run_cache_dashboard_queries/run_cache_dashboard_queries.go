package main

import (
	"factors/model/model"
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
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypePostgres, "Primary datastore type as memsql or postgres")

	memSQLDBMaxOpenConnections := flag.Int("memsql_max_open_connections", 100, "Max no.of open connections allowed on connection pool of memsql")
	memSQLDBMaxIdleConnections := flag.Int("memsql_max_idle_connections", 50, "Max no.of idle connections allowed on connection pool of memsql")

	sentryDSN := flag.String("sentry_dsn", "", "Sentry DSN")
	onlyWebAnalytics := flag.Bool("only_web_analytics", false, "Cache only web analytics dashboards.")

	redisHost := flag.String("redis_host", "localhost", "")
	redisPort := flag.Int("redis_port", 6379, "")

	gcpProjectID := flag.String("gcp_project_id", "", "Project ID on Google Cloud")
	gcpProjectLocation := flag.String("gcp_project_location", "", "Location of google cloud project cluster")
	disableRedisWrites := flag.Bool("disable_redis_writes", false, "To disable redis writes.")
	// better to have 0 or 1 values instead of false/true
	onlyAttribution := flag.Int("only_attribution", 0, "Cache only Attribution dashboards.")
	skipAttribution := flag.Int("skip_attribution", 0, "Skip the Attribution and run other.")

	runningForMemsql := flag.Int("running_for_memsql", 0, "Disable routines for memsql.")
	overrideAppName := flag.String("app_name", "", "Override default app_name.")
	overrideHealthcheckPingID := flag.String("healthcheck_ping_id", "", "Override default healthcheck ping id.")

	enableFilterOptimisation := flag.Bool("enable_filter_optimisation", false,
		"Enables filter optimisation changes for memsql implementation.")
	filterPropertiesStartTimestamp := flag.Int64("filter_properties_start_timestamp", -1,
		"Start timestamp of data available for filtering with parquet on memsql.")
	skipEventNameStepByProjectID := flag.String("skip_event_name_step_by_project_id", "", "")
	skipUserJoinInEventQueryByProjectID := flag.String("skip_user_join_in_event_query_by_project_id", "", "")

	flag.Parse()

	taskID := "dashboard_caching"
	if *overrideAppName != "" {
		taskID = *overrideAppName
	}
	defaultHealthcheckPingID := C.HealthcheckDashboardCachingPingID
	healthcheckPingID := C.GetHealthcheckPingID(defaultHealthcheckPingID, *overrideHealthcheckPingID)
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
			Host:        *memSQLHost,
			Port:        *memSQLPort,
			User:        *memSQLUser,
			Name:        *memSQLName,
			Password:    *memSQLPass,
			Certificate: *memSQLCertificate,
			AppName:     taskID,

			MaxOpenConnections:     *memSQLDBMaxOpenConnections,
			MaxIdleConnections:     *memSQLDBMaxIdleConnections,
			UseExactConnFromConfig: true,
		},
		PrimaryDatastore:                    *primaryDatastore,
		DisableRedisWrites:                  disableRedisWrites,
		EnableFilterOptimisation:            *enableFilterOptimisation,
		FilterPropertiesStartTimestamp:      *filterPropertiesStartTimestamp,
		SkipAttributionDashboardCaching:     *skipAttribution,
		OnlyAttributionDashboardCaching:     *onlyAttribution,
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
	var reportCollector sync.Map

	if !*onlyWebAnalytics {
		if C.GetIsRunningForMemsql() == 0 {
			waitGroup.Add(1)
			go cacheDashboardUnitsForProjects(*projectIDFlag, *excludeProjectIDFlag, *numRoutinesFlag, &reportCollector, &waitGroup)
		} else {
			cacheDashboardUnitsForProjects(*projectIDFlag, *excludeProjectIDFlag, *numRoutinesFlag, &reportCollector, &waitGroup)
		}
	}
	if C.GetIsRunningForMemsql() == 0 {
		waitGroup.Add(1)
		go cacheWebsiteAnalyticsForProjects(*projectIDFlag, *excludeProjectIDFlag, *numRoutinesForWebAnalyticsFlag, &reportCollector, &waitGroup)
	} else {
		cacheWebsiteAnalyticsForProjects(*projectIDFlag, *excludeProjectIDFlag, *numRoutinesForWebAnalyticsFlag, &reportCollector, &waitGroup)
	}

	if C.GetIsRunningForMemsql() == 0 {
		waitGroup.Wait()
	}
	timeTakenString, _ := reportCollector.Load("all")
	timeTakenStringWeb, _ := reportCollector.Load("web")
	// Collect all the reports in an array
	var allUnitReports []model.CachingUnitReport
	reportCollector.Range(func(key, value interface{}) bool {
		if key.(string) != "web" && key.(string) != "all" {
			allUnitReports = append(allUnitReports, value.(model.CachingUnitReport))
		}
		return true
	})

	slowUnits := model.GetNSlowestUnits(allUnitReports, 3)
	failedUnits := model.GetFailedUnitsByProject(allUnitReports)
	slowProjects := model.GetNSlowestProjects(allUnitReports, 5)
	failed, passed, notComputed := model.GetTotalFailedComputedNotComputed(allUnitReports)

	logCtx.Info("Completed dashboard caching")
	logCtx.WithFields(log.Fields{"slowUnits": slowUnits, "failedUnits": failedUnits, "slowProjects": slowProjects}).Info("Final Caching Job Report")
	notifyMessage = fmt.Sprintf("Caching successful for %s - %s projects. Time taken: %+v. Time taken for web analytics: %+v",
		*projectIDFlag, *excludeProjectIDFlag, timeTakenString, timeTakenStringWeb)
	logCtx.Info(notifyMessage)

	status := map[string]interface{}{
		"Summary":              notifyMessage,
		"TotalFailed":          failed,
		"TotalPassed":          passed,
		"TotalNotComputed":     notComputed,
		"Top3SlowUnits":        slowUnits,
		"FailedUnitsByProject": failedUnits,
		"Top5SlowProjects":     slowProjects,
	}
	C.PingHealthcheckForSuccess(healthcheckPingID, status)
}

func cacheDashboardUnitsForProjects(projectIDs, excludeProjectIDs string, numRoutines int, reportCollector *sync.Map, waitGroup *sync.WaitGroup) {

	if C.GetIsRunningForMemsql() == 0 {
		defer waitGroup.Done()
	}
	startTime := util.TimeNowUnix()
	store.GetStore().CacheDashboardUnitsForProjects(projectIDs, excludeProjectIDs, numRoutines, reportCollector)
	timeTakenString := util.SecondsToHMSString(util.TimeNowUnix() - startTime)
	reportCollector.Store("all", timeTakenString)
}

func cacheWebsiteAnalyticsForProjects(projectIDs, excludeProjectIDs string, numRoutines int, reportCollector *sync.Map, waitGroup *sync.WaitGroup) {
	if C.GetIsRunningForMemsql() == 0 {
		defer waitGroup.Done()
	}
	startTime := util.TimeNowUnix()
	store.GetStore().CacheWebsiteAnalyticsForProjects(projectIDs, excludeProjectIDs, numRoutines, reportCollector)
	timeTakenStringWeb := util.SecondsToHMSString(util.TimeNowUnix() - startTime)
	reportCollector.Store("web", timeTakenStringWeb)
}
