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
	debugEnabled := flag.Bool("debug_enabled", false, "Enabled/Disable debug for the query.")
	numRoutinesFlag := flag.Int("num_routines", 4, "Number of dashboard units to sync in parallel. Each dashboard unit runs 4 queries")
	numRoutinesForWebAnalyticsFlag := flag.Int("num_routines_for_web_analytics", 1,
		"No.of routines to use for web analytics dashboard caching.")

	cacheForLongerExpiryProjects := flag.String("cache_for_longer_expiry_projects", "", "Comma separated project ids to run for. * to run for all")
	startTimestampForWeekMonth := flag.Int64("start_timestamp_week_month", -1,
		"Start timestamp of caching week/month")
	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	memSQLPort := flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser := flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName := flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass := flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	memSQLCertificate := flag.String("memsql_cert", "", "")
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypeMemSQL, "Primary datastore type as memsql or postgres")

	memSQLDBMaxOpenConnections := flag.Int("memsql_max_open_connections", 100, "Max no.of open connections allowed on connection pool of memsql")
	memSQLDBMaxIdleConnections := flag.Int("memsql_max_idle_connections", 50, "Max no.of idle connections allowed on connection pool of memsql")

	sentryDSN := flag.String("sentry_dsn", "", "Sentry DSN")
	onlyWebAnalytics := flag.Int("only_web_analytics", 0, "Cache only web analytics dashboards.")
	skipWebAnalytics := flag.Int("skip_web_analytics", 0, "Skip Caching for web analytics dashboards.")

	redisHost := flag.String("redis_host", "localhost", "")
	redisPort := flag.Int("redis_port", 6379, "")

	gcpProjectID := flag.String("gcp_project_id", "", "Project ID on Google Cloud")
	gcpProjectLocation := flag.String("gcp_project_location", "", "Location of google cloud project cluster")
	// better to have 0 or 1 values instead of false/true
	onlyAttribution := flag.Int("only_attribution", 0, "Cache only Attribution dashboards.")
	skipAttribution := flag.Int("skip_attribution", 0, "Skip the Attribution and run other.")
	enableUsageBasedDashboardCaching := flag.Int("enable_usage_based_caching", 0, "Usage based dashboard caching analytics for 14 days limit.")
	onlyKPICaching := flag.Int("only_kpi_caching", 0, "Cache only KPI Queries on dashboards.")
	skipKPICaching := flag.Int("skip_kpi_caching", 0, "Skip caching KPI Queries on dashboards.")

	runningForMemsql := flag.Int("running_for_memsql", 0, "Disable routines for memsql.")
	overrideAppName := flag.String("app_name", "", "Override default app_name.")
	overrideHealthcheckPingID := flag.String("healthcheck_ping_id", "", "Override default healthcheck ping id.")

	resourcePoolForAnalytics := flag.String("resource_pool_for_analytics", "",
		"Given resource_pool will be used for analytics queries.")

	enableFilterOptimisation := flag.Bool("enable_filter_optimisation", false,
		"Enables filter optimisation changes for memsql implementation.")
	filterPropertiesStartTimestamp := flag.Int64("filter_properties_start_timestamp", -1,
		"Start timestamp of data available for filtering with parquet on memsql.")
	skipEventNameStepByProjectID := flag.String("skip_event_name_step_by_project_id", "", "")
	skipUserJoinInEventQueryByProjectID := flag.String("skip_user_join_in_event_query_by_project_id", "", "")
	enableSlowDBQueryLogging := flag.Bool("log_slow_db_queries", false, "Logs queries with execution time greater than 50ms.")
	allowProfilesGroupSupport := flag.String("allow_profiles_group_support", "", "")
	enableOptimisedFilterOnProfileQuery := flag.Int("enable_optimised_filter_on_profile_query",
		0, "Enables filter optimisation logic for profiles query.")
	enableOptimisedFilterOnEventUserQuery := flag.Int("enable_optimised_filter_on_event_user_query",
		0, "Enables filter optimisation logic for events and users query.")
	customerEnabledProjectsLastComputed := flag.String("customer_enabled_projects_last_computed",
		"*", "List of projects customer enabled forLast Computed")
	IncreaseKPILimitForProjectIDs := flag.String("increase_kpi_limit_for_projectids", "", "List of projectIds where kpi limit in increased.")

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
		RedisHost:          *redisHost,
		RedisPort:          *redisPort,
		SentryDSN:          *sentryDSN,
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
		PrimaryDatastore:                      *primaryDatastore,
		EnableFilterOptimisation:              *enableFilterOptimisation,
		AllowProfilesGroupSupport:             *allowProfilesGroupSupport,
		FilterPropertiesStartTimestamp:        *filterPropertiesStartTimestamp,
		SkipAttributionDashboardCaching:       *skipAttribution,
		OnlyAttributionDashboardCaching:       *onlyAttribution,
		IsRunningForMemsql:                    *runningForMemsql,
		SkipEventNameStepByProjectID:          *skipEventNameStepByProjectID,
		SkipUserJoinInEventQueryByProjectID:   *skipUserJoinInEventQueryByProjectID,
		DebugEnabled:                          *debugEnabled,
		ResourcePoolForAnalytics:              *resourcePoolForAnalytics,
		UsageBasedDashboardCaching:            *enableUsageBasedDashboardCaching,
		OnlyKPICaching:                        *onlyKPICaching,
		SkipKPICaching:                        *skipKPICaching,
		EnableSlowDBQueryLogging:              *enableSlowDBQueryLogging,
		EnableOptimisedFilterOnProfileQuery:   *enableOptimisedFilterOnProfileQuery != 0,
		EnableOptimisedFilterOnEventUserQuery: *enableOptimisedFilterOnEventUserQuery != 0,
		CustomerEnabledProjectsLastComputed:   C.GetTokensFromStringListAsUint64(*customerEnabledProjectsLastComputed),
		IncreaseKPILimitForProjectIDs:         *IncreaseKPILimitForProjectIDs,
		StartTimestampForWeekMonth:            *startTimestampForWeekMonth,
		CacheForLongerExpiryProjects:          *cacheForLongerExpiryProjects,
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
	model.SetSmartPropertiesReservedNames()
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

	if *onlyWebAnalytics == 0 {
		if C.GetIsRunningForMemsql() == 0 {
			waitGroup.Add(1)
			go cacheDashboardUnitsForProjects(*projectIDFlag, *excludeProjectIDFlag, *numRoutinesFlag, &reportCollector, &waitGroup, *startTimestampForWeekMonth)
		} else {
			cacheDashboardUnitsForProjects(*projectIDFlag, *excludeProjectIDFlag, *numRoutinesFlag, &reportCollector, &waitGroup, *startTimestampForWeekMonth)
		}
	}
	if *skipWebAnalytics == 0 {
		if C.GetIsRunningForMemsql() == 0 {
			waitGroup.Add(1)
			go cacheWebsiteAnalyticsForProjects(*projectIDFlag, *excludeProjectIDFlag, *numRoutinesForWebAnalyticsFlag, &reportCollector, &waitGroup)
		} else {
			cacheWebsiteAnalyticsForProjects(*projectIDFlag, *excludeProjectIDFlag, *numRoutinesForWebAnalyticsFlag, &reportCollector, &waitGroup)
		}
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
	timedOutUnits := model.GetTimedOutUnitsByProject(allUnitReports)
	slowProjects := model.GetNSlowestProjects(allUnitReports, 5)
	failed, passed, notComputed := model.GetTotalFailedComputedNotComputed(allUnitReports)

	logCtx.Info("Completed dashboard caching")
	notifyMessage = fmt.Sprintf("Caching successful for %s - %s projects. Time taken: %+v. Time taken for web analytics: %+v",
		*projectIDFlag, *excludeProjectIDFlag, timeTakenString, timeTakenStringWeb)
	logCtx.Info(notifyMessage)

	status := map[string]interface{}{
		"Summary":                notifyMessage,
		"TotalFailed":            failed,
		"TotalPassed":            passed,
		"TotalNotComputed":       notComputed,
		"Top3SlowUnits":          slowUnits,
		"FailedUnitsByProject":   failedUnits,
		"TimedOutUnitsByProject": timedOutUnits,
		"Top5SlowProjects":       slowProjects,
	}
	C.PingHealthcheckForSuccess(healthcheckPingID, status)

	slowUnits = model.GetNSlowestUnits(allUnitReports, 20)
	slowProjects = model.GetNSlowestProjects(allUnitReports, 10)

	logCtx.WithFields(log.Fields{
		"Summary":                       notifyMessage,
		"TimeTakenForNormalUnits":       timeTakenString,
		"TimeTakenForWebAnalyticsUnits": timeTakenStringWeb,
		"TotalFailed":                   failed,
		"TotalPassed":                   passed,
		"TotalNotComputed":              notComputed,
		"Top20SlowUnits":                slowUnits,
		"FailedUnitsByProject":          failedUnits,
		"TimedOutUnitsByProject":        timedOutUnits,
		"Top10SlowProjects":             slowProjects}).Info("Final Caching Job Report")

}

func cacheDashboardUnitsForProjects(projectIDs, excludeProjectIDs string, numRoutines int, reportCollector *sync.Map, waitGroup *sync.WaitGroup, startTimeForCache int64) {

	if C.GetIsRunningForMemsql() == 0 {
		defer waitGroup.Done()
	}
	startTime := util.TimeNowUnix()
	store.GetStore().CacheDashboardUnitsForProjects(projectIDs, excludeProjectIDs, numRoutines, reportCollector, C.EnableOptimisedFilterOnEventUserQuery(), startTimeForCache)
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
