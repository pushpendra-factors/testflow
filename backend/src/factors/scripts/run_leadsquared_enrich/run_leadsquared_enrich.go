package main

import (
	"flag"
	"time"

	C "factors/config"
	"factors/model/model"
	"factors/model/store"
	"factors/util"

	leadSquaredEnrich "factors/task/lead_squared_enrich"
	taskWrapper "factors/task/task_wrapper"

	U "factors/util"

	log "github.com/sirupsen/logrus"
)

func main() {

	envFlag := flag.String("env", C.DEVELOPMENT, "Environment. Could be development|staging|production.")
	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	isPSCHost := flag.Int("memsql_is_psc_host", C.MemSQLDefaultDBParams.IsPSCHost, "")
	memSQLPort := flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser := flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName := flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass := flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	memSQLCertificate := flag.String("memsql_cert", "", "")
	sentryRollupSyncInSecs := flag.Int("sentry_rollup_sync_in_seconds", 300, "Enables to send errors to sentry in given interval.")
	sentryDSN := flag.String("sentry_dsn", "", "Sentry DSN")
	redisHost := flag.String("redis_host", "localhost", "")
	redisPort := flag.Int("redis_port", 6379, "")
	redisHostPersistent := flag.String("redis_host_ps", "localhost", "")
	redisPortPersistent := flag.Int("redis_port_ps", 6379, "")
	cacheSortedSet := flag.Bool("cache_with_sorted_set", false, "Cache with sorted set keys")
	useSourcePropertyOverwriteByProjectID := flag.String("use_source_property_overwrite_by_project_id", "", "")
	captureSourceInUsersTable := flag.String("capture_source_in_users_table", "", "")
	restrictReusingUsersByCustomerUserId := flag.String("restrict_reusing_users_by_customer_user_id", "", "")
	propertiesTypeCacheSize := flag.Int("property_details_cache_size", 0, "Cache size for in memory property detail.")
	enablePropertyTypeFromDB := flag.Bool("enable_property_type_from_db", false, "Enable property type check from db.")
	whitelistedProjectIDPropertyTypeFromDB := flag.String("whitelisted_project_ids_property_type_check_from_db", "", "Allowed project id for property type check from db.")
	blacklistedProjectIDPropertyTypeFromDB := flag.String("blacklisted_project_ids_property_type_check_from_db", "", "Blocked project id for property type check from db.")
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypeMemSQL, "Primary datastore type as memsql or postgres")
	lookback := flag.Int("lookback", 1, "lookback for job")
	overrideHealthcheckPingID := flag.String("healthcheck_ping_id", "", "Override default healthcheck ping id.")
	numDocRoutines := flag.Int("num_unique_doc_routines", 1, "Number of unique document go routines per project")
	minSyncTimestamp := flag.Int64("min_sync_timestamp", 0, "Min timstamp from where to process records")
	projectIDList := flag.String("project_ids", "*", "List of project_id to run for.")
	disabledProjectIDList := flag.String("disabled_project_ids", "", "List of project_ids to exclude.")
	IngestionTimezoneEnabledProjectIDs := flag.String("ingestion_timezone_enabled_projects", "", "List of projectIds whose ingestion timezone is enabled.")
	enableDomainsGroupByProjectID := flag.String("enable_domains_group_by_project_id", "", "")
	enableUserDomainsGroupByProjectID := flag.String("enable_user_domains_group_by_project_id", "", "Allow domains group for users")
	allowEmailDomainsByProjectID := flag.String("allow_email_domain_by_project_id", "", "Allow email domains for domain group")
	removeDisabledEventUserPropertiesByProjectId := flag.String("remove_disabled_event_user_properties",
		"", "List of projects to disable event user property population in events.")

	recordProcessLimit := flag.Int("record_process_limit", 0, "Adding limit for processing records") // By default, pull all records.
	userPropertyUpdateOptProjects := flag.String("user_property_update_opt_projects", "", "")
	enableTotalSessionPropertiesV2ByProjectID := flag.String("enable_total_session_properties_v2", "", "")
	enableDomainWebsitePropertiesByProjectID := flag.String("enable_domain_website_properties_by_project_id", "", "")

	flag.Parse()

	appName := "lead_squared_enrich"
	defaultHealthcheckPingID := C.HealthcheckLeadSquaredEnrichPingID
	healthcheckPingID := C.GetHealthcheckPingID(defaultHealthcheckPingID, *overrideHealthcheckPingID)
	defer C.PingHealthcheckForPanic(appName, *envFlag, healthcheckPingID)

	config := &C.Configuration{
		AppName: appName,
		Env:     *envFlag,
		MemSQLInfo: C.DBConf{
			Host:        *memSQLHost,
			IsPSCHost:   *isPSCHost,
			Port:        *memSQLPort,
			User:        *memSQLUser,
			Name:        *memSQLName,
			Password:    *memSQLPass,
			Certificate: *memSQLCertificate,
			AppName:     appName,
		},
		PrimaryDatastore:                             *primaryDatastore,
		RedisHost:                                    *redisHost,
		RedisPort:                                    *redisPort,
		RedisHostPersistent:                          *redisHostPersistent,
		RedisPortPersistent:                          *redisPortPersistent,
		SentryDSN:                                    *sentryDSN,
		SentryRollupSyncInSecs:                       *sentryRollupSyncInSecs,
		CacheSortedSet:                               *cacheSortedSet,
		UseSourcePropertyOverwriteByProjectIDs:       *useSourcePropertyOverwriteByProjectID,
		CaptureSourceInUsersTable:                    *captureSourceInUsersTable,
		RestrictReusingUsersByCustomerUserId:         *restrictReusingUsersByCustomerUserId,
		IngestionTimezoneEnabledProjectIDs:           C.GetTokensFromStringListAsString(*IngestionTimezoneEnabledProjectIDs),
		EnableDomainsGroupByProjectID:                *enableDomainsGroupByProjectID,
		EnableUserDomainsGroupByProjectID:            *enableUserDomainsGroupByProjectID,
		AllowEmailDomainsByProjectID:                 *allowEmailDomainsByProjectID,
		RemoveDisabledEventUserPropertiesByProjectID: *removeDisabledEventUserPropertiesByProjectId,
		UserPropertyUpdateOptProjects:                *userPropertyUpdateOptProjects,
		EnableTotalSessionPropertiesV2ByProjectID:    *enableTotalSessionPropertiesV2ByProjectID,
		EnableDomainWebsitePropertiesByProjectID:     *enableDomainWebsitePropertiesByProjectID,
	}
	C.InitConf(config)

	err := C.InitDB(*config)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize DB")
	}

	err = C.InitDBWithMaxIdleAndMaxOpenConn(*config, 200, 100)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{"env": *envFlag,
			"host": *memSQLHost, "port": *memSQLPort}).Panic("Failed to initialize DB.")
	}
	db := C.GetServices().Db
	defer db.Close()

	C.InitRedis(config.RedisHost, config.RedisPort)
	C.InitRedisPersistent(config.RedisHostPersistent, config.RedisPortPersistent)
	C.InitSentryLogging(config.SentryDSN, config.AppName)
	C.InitMetricsExporter(config.Env, config.AppName, config.GCPProjectID, config.GCPProjectLocation)
	defer C.WaitAndFlushAllCollectors(65 * time.Second)
	C.InitPropertiesTypeCache(*enablePropertyTypeFromDB, *propertiesTypeCacheSize, *whitelistedProjectIDPropertyTypeFromDB, *blacklistedProjectIDPropertyTypeFromDB)

	projectIdsArray := make([]int64, 0)
	mappings, err := store.GetStore().GetAllLeadSquaredEnabledProjects()
	if err != nil {
		C.PingHealthcheckForFailure(healthcheckPingID, "Failed to get LeadSquared Projects")
		return
	}

	featureProjectIDs, err := store.GetStore().GetAllProjectsWithFeatureEnabled(model.FEATURE_LEADSQUARED, false)
	if err != nil {
		log.WithError(err).Error("Failed to get leadsquared feature enabled projects.")
		return
	}

	featureEnabledIntegrations := map[int64]model.LeadSquaredConfig{}
	for pid := range mappings {
		if util.ContainsInt64InArray(featureProjectIDs, pid) {
			featureEnabledIntegrations[pid] = mappings[pid]
		}
	}

	mappings = featureEnabledIntegrations

	allProjects, allowedProjects, disabledProjects := C.GetProjectsFromListWithAllProjectSupport(
		*projectIDList, *disabledProjectIDList)
	if !allProjects {
		log.WithField("projects", allowedProjects).Info("Running only for the given list of projects.")
	}

	for projectID := range mappings {
		if exists := disabledProjects[projectID]; exists {
			continue
		}

		if !allProjects {
			if _, exists := allowedProjects[projectID]; !exists {
				continue
			}
		}

		projectIdsArray = append(projectIdsArray, projectID)
	}

	configs := make(map[string]interface{})
	configs["document_routines"] = *numDocRoutines
	configs["min_sync_timestamp"] = *minSyncTimestamp
	configs["record_process_limit"] = *recordProcessLimit
	status := taskWrapper.TaskFuncWithProjectId(appName, *lookback, projectIdsArray, leadSquaredEnrich.RunLeadSquaredEnrich, configs)
	if len(status) == 0 { // skip ping to healthcheck if no task ran
		return
	}

	anyFailure := false
	for _, valInt := range status {
		var projectEnrichStatus map[string]leadSquaredEnrich.EnrichStatus
		var ok bool
		if projectEnrichStatus, ok = valInt.(map[string]leadSquaredEnrich.EnrichStatus); !ok {
			continue
		}

		for commonStatus := range projectEnrichStatus {
			if commonStatus == U.CRM_SYNC_STATUS_FAILURES {
				anyFailure = true
			}
		}
	}

	log.Info(status)
	if anyFailure {
		C.PingHealthcheckForFailure(healthcheckPingID, status)
		return
	}
	C.PingHealthcheckForSuccess(healthcheckPingID, status)

}
