package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	log "github.com/sirupsen/logrus"

	C "factors/config"
	"factors/util"

	"factors/task/session"
)

func main() {
	env := flag.String("env", C.DEVELOPMENT, "")

	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	isPSCHost := flag.Int("memsql_is_psc_host", C.MemSQLDefaultDBParams.IsPSCHost, "")
	memSQLPort := flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser := flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName := flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass := flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	memSQLCertificate := flag.String("memsql_cert", "", "")
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypeMemSQL, "Primary datastore type as memsql or postgres")

	memSQLUseExactConnectionsConfig := flag.Bool("memsql_use_exact_connection_config", false, "Use exact connection for open and idle as given.")
	memSQLDBMaxOpenConnections := flag.Int("memsql_max_open_connections", 100, "Max no.of open connections allowed on connection pool of memsql")
	memSQLDBMaxIdleConnections := flag.Int("memsql_max_idle_connections", 50, "Max no.of idle connections allowed on connection pool of memsql")

	redisHost := flag.String("redis_host", "localhost", "")
	redisPort := flag.Int("redis_port", 6379, "")
	redisHostPersistent := flag.String("redis_host_ps", "localhost", "")
	redisPortPersistent := flag.Int("redis_port_ps", 6379, "")

	// Allowed list of projects to add session. Defaults to all (*), if not given.
	projectIds := flag.String("project_ids", "*", "Allowed projects to create sessions offline.")
	disabledProjectIds := flag.String("disabled_project_ids", "", "Disallowed projects to create sessions offline.")
	numProjectRoutines := flag.Int("num_project_routines", 1, "Number of project level routines to use.")
	numUserRoutines := flag.Int("num_user_routines", 1, "Number of user level routines to use.")
	bufferTimeBeforeCreateSessionInMins := flag.Int64("buffer_time_in_mins", 30, "Buffer time to wait before processing an event for session.")

	// Limits the start_timestamp to max lookback, if exceeds.
	maxLookbackHours := flag.Int64("max_lookback_hours", 0, "Max lookback hours to look for session existence.")

	// Add session for a specific window of events.
	startTimestamp := flag.Int64("start_timestamp", 0, "Add session to specific window of events - start timestamp.")
	endTimestamp := flag.Int64("end_timestamp", 0, "Add session to specific window of events - end timestamp.")

	sentryDSN := flag.String("sentry_dsn", "", "Sentry DSN")
	sentryRollupSyncInSecs := flag.Int("sentry_rollup_sync_in_seconds", 300, "Enables to send errors to sentry in given interval.")

	gcpProjectID := flag.String("gcp_project_id", "", "Project ID on Google Cloud")
	gcpProjectLocation := flag.String("gcp_project_location", "", "Location of google cloud project cluster")
	cacheSortedSet := flag.Bool("cache_with_sorted_set", false, "Cache with sorted set keys")

	overrideHealthcheckPingID := flag.String("healthcheck_ping_id", "", "Override default healthcheck ping id.")
	overrideAppName := flag.String("app_name", "", "Override default app_name.")
	allowChannelGroupingForProjectIDs := flag.String("allow_channel_grouping_for_projects",
		"", "List of projects to allow channel property population in sesion events.")
	removeDisabledEventUserPropertiesByProjectId := flag.String("remove_disabled_event_user_properties",
		"", "List of projects to disable event user property population in events.")

	enableOLTPQueriesMemSQLImprovements := flag.String("enable_OLTP_queries_memsql_improvements", "", "")
	captureSourceInUsersTable := flag.String("capture_source_in_users_table", "", "")
	sessionBatchTransactionBatchSize := flag.Int("session_batch_transaction_batch_size", 0, "")
	IngestionTimezoneEnabledProjectIDs := flag.String("ingestion_timezone_enabled_projects", "", "List of projectIds whose ingestion timezone is enabled.")

	enableUserLevelEventPullForAddSessionByProjectID := flag.String("enable_user_level_pull", "", "List of projectIds where user level event pull is enabled for AddSession job")
	eventsPullMaxLimit := flag.Int("max_limit_for_events_pull", 50000, "Maximum limit for pulling events in V2") // Default is 50000
	batchRangeInSeconds := flag.Int64("batch_range_in_seconds", 0, "Batch size for Add Session job")

	disableUpdateNextSessionTimestamp := flag.Int("disable_update_next_session_timestamp", 0, "Disable the update next session timestamp. Used for historical runs.")
	userPropertyUpdateOptProjects := flag.String("user_property_update_opt_projects", "", "")
	enableTotalSessionPropertiesV2ByProjectID := flag.String("enable_total_session_properties_v2", "", "")
	enableDomainWebsitePropertiesByProjectID := flag.String("enable_domain_website_properties_by_project_id", "", "")

	flag.Parse()

	if *env != "development" &&
		*env != "staging" &&
		*env != "production" {
		err := fmt.Errorf("env [ %s ] not recognised", *env)
		panic(err)
	}

	defaultAppName := "add_session"
	defaultHealthcheckPingID := C.HealthcheckAddSessionPingID
	healthcheckPingID := C.GetHealthcheckPingID(defaultHealthcheckPingID, *overrideHealthcheckPingID)
	appName := C.GetAppName(defaultAppName, *overrideAppName)
	defer C.PingHealthcheckForPanic(appName, *env, healthcheckPingID)

	config := &C.Configuration{
		AppName:            appName,
		Env:                *env,
		GCPProjectID:       *gcpProjectID,
		GCPProjectLocation: *gcpProjectLocation,
		MemSQLInfo: C.DBConf{
			Host:        *memSQLHost,
			IsPSCHost:   *isPSCHost,
			Port:        *memSQLPort,
			User:        *memSQLUser,
			Name:        *memSQLName,
			Password:    *memSQLPass,
			Certificate: *memSQLCertificate,
			AppName:     appName,

			MaxOpenConnections:     *memSQLDBMaxOpenConnections,
			MaxIdleConnections:     *memSQLDBMaxIdleConnections,
			UseExactConnFromConfig: *memSQLUseExactConnectionsConfig,
		},
		PrimaryDatastore:                    *primaryDatastore,
		RedisHost:                           *redisHost,
		RedisPort:                           *redisPort,
		RedisHostPersistent:                 *redisHostPersistent,
		RedisPortPersistent:                 *redisPortPersistent,
		SentryDSN:                           *sentryDSN,
		SentryRollupSyncInSecs:              *sentryRollupSyncInSecs,
		CacheSortedSet:                      *cacheSortedSet,
		AllowChannelGroupingForProjectIDs:   *allowChannelGroupingForProjectIDs,
		EnableOLTPQueriesMemSQLImprovements: *enableOLTPQueriesMemSQLImprovements,
		CaptureSourceInUsersTable:           *captureSourceInUsersTable,
		SessionBatchTransactionBatchSize:    *sessionBatchTransactionBatchSize,
		IngestionTimezoneEnabledProjectIDs:  C.GetTokensFromStringListAsString(*IngestionTimezoneEnabledProjectIDs),
		EnableUserLevelEventPullForAddSessionByProjectID: *enableUserLevelEventPullForAddSessionByProjectID,
		EventsPullMaxLimit:                           *eventsPullMaxLimit,
		DisableUpdateNextSessionTimestamp:            *disableUpdateNextSessionTimestamp,
		RemoveDisabledEventUserPropertiesByProjectID: *removeDisabledEventUserPropertiesByProjectId,
		UserPropertyUpdateOptProjects:                *userPropertyUpdateOptProjects,
		EnableTotalSessionPropertiesV2ByProjectID:    *enableTotalSessionPropertiesV2ByProjectID,
		EnableDomainWebsitePropertiesByProjectID:     *enableDomainWebsitePropertiesByProjectID,
	}

	C.InitConf(config)
	C.InitSortedSetCache(config.CacheSortedSet)
	// Will allow all 50/50 connection to be idle on the pool.
	// As we allow num_routines (per project) as per no.of db connections
	// and will be used continiously.
	err := C.InitDBWithMaxIdleAndMaxOpenConn(*config, 50, 50)
	if err != nil {
		log.WithError(err).Panic("Failed to initialize db in add session.")
	}

	// Cache dependency for requests not using queue.
	C.InitRedis(config.RedisHost, config.RedisPort)
	C.InitRedisPersistent(config.RedisHostPersistent, config.RedisPortPersistent)
	C.InitSentryLogging(config.SentryDSN, config.AppName)
	C.InitMetricsExporter(config.Env, config.AppName, config.GCPProjectID, config.GCPProjectLocation)
	defer C.WaitAndFlushAllCollectors(65 * time.Second)

	allowedProjectIds, errCode := session.GetAddSessionAllowedProjects(*projectIds, *disabledProjectIds)
	if errCode != http.StatusFound {
		log.WithField("err_code", errCode).Error("Failed to get add session allowed project ids.")
		os.Exit(0)
	}

	logCtx := log.WithField("start_timestamp", *startTimestamp).WithField("end_timestamp", *endTimestamp)
	if *endTimestamp > 0 && *startTimestamp == 0 {
		logCtx.Fatal("start_timestamp cannot be zero when start_timestamp is provided.")
	}
	if *startTimestamp > 0 && *endTimestamp == 0 {
		logCtx.Fatal("end_timestamp cannot be zero when start_timestamp is provided.")
	}
	if *startTimestamp > 0 && *endTimestamp <= *startTimestamp {
		logCtx.Fatal("end_timestamp cannot be lower than or equal to start_timestamp.")
	}

	var maxLookbackTimestamp int64
	if *maxLookbackHours > 0 {
		maxLookbackTimestamp = util.UnixTimeBeforeDuration(time.Hour * time.Duration(*maxLookbackHours))
	}

	overAllStatusMap := make(map[int64]session.Status, 0)
	var overAllError error

	if *startTimestamp > 0 && *endTimestamp > 0 && *batchRangeInSeconds > 0 { // New logic
		batchedTimestamp := util.GetBatchRangeFromStartAndEndTimestamp(*startTimestamp, *endTimestamp, *batchRangeInSeconds)
		for _, timeRange := range batchedTimestamp {
			statusMap, err := session.AddSession(allowedProjectIds, maxLookbackTimestamp,
				timeRange[0], timeRange[1], *bufferTimeBeforeCreateSessionInMins, *numProjectRoutines, *numUserRoutines)

			if err != nil {
				overAllError = err
			}

			for pid, status := range statusMap {
				if _, exists := overAllStatusMap[pid]; !exists {
					overAllStatusMap[pid] = status
					continue
				}

				existingStatus := overAllStatusMap[pid]
				if existingStatus.SeenFailure {
					existingStatus.Status = session.StatusFailed
					continue
				}

				if status.Status != session.StatusNotModified {
					existingStatus.Status = status.Status
				}

				existingStatus.NoOfSessionsCreated += status.NoOfSessionsCreated
				existingStatus.NoOfEventsProcessed += status.NoOfEventsProcessed
				existingStatus.NoOfUserPropertiesUpdates += status.NoOfUserPropertiesUpdates
				overAllStatusMap[pid] = existingStatus
			}
		}
	} else { // Old logic
		overAllStatusMap, overAllError = session.AddSession(allowedProjectIds, maxLookbackTimestamp,
			*startTimestamp, *endTimestamp, *bufferTimeBeforeCreateSessionInMins,
			*numProjectRoutines, *numUserRoutines)
	}

	modifiedStatusMap := make(map[int64]session.Status, 0)
	for pid, status := range overAllStatusMap {
		if status.Status == session.StatusNotModified {
			continue
		}
		modifiedStatusMap[pid] = status
	}

	status := map[string]interface{}{
		"new_session_status": modifiedStatusMap,
	}

	if overAllError != nil {
		C.PingHealthcheckForFailure(healthcheckPingID, overAllError)
		log.WithError(overAllError).WithField("status", overAllStatusMap).Error("Seen failures while adding sessions.")
	} else {
		C.PingHealthcheckForSuccess(healthcheckPingID, status)
		log.WithField("no_of_projects", len(allowedProjectIds)).Info("Successfully added sessions.")
	}
}
