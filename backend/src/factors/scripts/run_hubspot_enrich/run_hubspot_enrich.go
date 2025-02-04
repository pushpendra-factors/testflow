package main

import (
	C "factors/config"
	T "factors/task/hubspot_enrich"
	"flag"
	"fmt"
	"time"

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

	memSQLDBMaxOpenConnections := flag.Int("memsql_max_open_connections", 50, "Max no.of open connections allowed on connection pool of memsql")
	memSQLDBMaxIdleConnections := flag.Int("memsql_max_idle_connections", 50, "Max no.of idle connections allowed on connection pool of memsql")

	redisHost := flag.String("redis_host", "localhost", "")
	redisPort := flag.Int("redis_port", 6379, "")
	redisHostPersistent := flag.String("redis_host_ps", "localhost", "")
	redisPortPersistent := flag.Int("redis_port_ps", 6379, "")

	factorsSixsignalAPIKey := flag.String("factors_sixsignal_api_key", "dummy", "")

	sentryDSN := flag.String("sentry_dsn", "", "Sentry DSN")
	sentryRollupSyncInSecs := flag.Int("sentry_rollup_sync_in_seconds", 300, "Enables to send errors to sentry in given interval.")

	gcpProjectID := flag.String("gcp_project_id", "", "Project ID on Google Cloud")
	gcpProjectLocation := flag.String("gcp_project_location", "", "Location of google cloud project cluster")

	dryRunSmartEvent := flag.Bool("dry_run_smart_event", false, "Dry run mode for smart event creation")
	propertiesTypeCacheSize := flag.Int("property_details_cache_size", 0, "Cache size for in memory property detail.")
	enablePropertyTypeFromDB := flag.Bool("enable_property_type_from_db", false, "Enable property type check from db.")
	whitelistedProjectIDPropertyTypeFromDB := flag.String("whitelisted_project_ids_property_type_check_from_db", "", "Allowed project id for property type check from db.")
	blacklistedProjectIDPropertyTypeFromDB := flag.String("blacklisted_project_ids_property_type_check_from_db", "", "Blocked project id for property type check from db.")
	cacheSortedSet := flag.Bool("cache_with_sorted_set", false, "Cache with sorted set keys")

	projectIDList := flag.String("project_ids", "*", "List of project_id to run for.")
	skippedOtpProjectIDs := flag.String("skipped_otp_project_ids", "", "List of project_id to be skip for otp job.")
	disabledProjectIDList := flag.String("disabled_project_ids", "", "List of project_ids to exclude.")
	numProjectRoutines := flag.Int("num_project_routines", 1, "Number of project level go routines to run in parallel.")

	numDocRoutines := flag.Int("num_unique_doc_routines", 1, "Number of unique document go routines per project")

	overrideHealthcheckPingID := flag.String("healthcheck_ping_id", "", "Override default healthcheck ping id.")
	overrideHubspotProjectDistributerHealthcheckPingID := flag.String("project_distributer_healthcheck_ping_id", "", "Override default project distributer healthcheck ping id.")
	overrideAppName := flag.String("app_name", "", "Override default app_name.")
	projectDistributerAppName := flag.String("project_distributer_app_name", "hubspot_project_distributer", "Override default app_name for project distributer.")
	taskManagementLookback := flag.Int("task_management_lookback", 1, "")
	enableHubspotGroupsByProjectID := flag.String("enable_hubspot_groups_by_project_id", "", "Enable hubspot groups for projects.")
	useSourcePropertyOverwriteByProjectID := flag.String("use_source_property_overwrite_by_project_id", "", "")
	captureSourceInUsersTable := flag.String("capture_source_in_users_table", "", "")
	restrictReusingUsersByCustomerUserId := flag.String("restrict_reusing_users_by_customer_user_id", "", "")
	enableHubspotFormEventsByProjectID := flag.String("enable_hubspot_form_events_by_project_id", "", "")
	hubspotMaxCreatedAt := flag.Int64("huspot_max_created_at", time.Now().Unix(), "max created_at for records to process.")
	lightProjectsCountThreshold := flag.Int("light_projects_count_threshold", 50000, "Threshold on count for distribution across jobs")
	enrichHeavy := flag.Bool("enrich_heavy", false, "Run heavy projects")
	recordProcessLimit := flag.Int("record_process_limit", 50000, "Number of records to process per project.")
	disableNonMarketingContactByProjectID := flag.String("disable_non_marketing_contact_by_project_id", "", "Disable hubspot non marketing contacts from processing")
	hubspotAppID := flag.String("hubspot_app_id", "", "Hubspot app id for oauth integration")
	hubspotAppSecret := flag.String("hubspot_app_secret", "", "Hubspot app secret for oauth integration")
	allowIdentificationOverwriteUsingSourceByProjectID := flag.String("allow_identification_overwrite_using_source_by_project_id", "", "Allow identification overwrite based on request source.")
	allowHubspotPastEventsEnrichmentByProjectID := flag.String("allow_hubspot_past_events_enrichment_by_project_id", "", "")
	allowHubspotContactListInsertByProjectID := flag.String("allow_hubspot_contact_list_insert_by_project_id", "", "")
	IngestionTimezoneEnabledProjectIDs := flag.String("ingestion_timezone_enabled_projects", "", "List of projectIds whose ingestion timezone is enabled.")
	enableDomainsGroupByProjectID := flag.String("enable_domains_group_by_project_id", "", "")
	enableSyncReferenceFieldsByProjectID := flag.String("enable_sync_reference_fields_by_project_id", "", "")
	enableUserDomainsGroupByProjectID := flag.String("enable_user_domains_group_by_project_id", "", "Allow domains group for users")
	useHubspotCompaniesv3APIByProjectID := flag.String("use_hubspot_companies_v3_by_project_id", "", "")
	allowEmailDomainsByProjectID := flag.String("allow_email_domain_by_project_id", "", "Allow email domains for domain group")
	useHubspotEngagementsv3APIByProjectID := flag.String("use_hubspot_engagements_v3_by_project_id", "", "")
	useHubspotDealsv3APIByProjectID := flag.String("use_hubspot_deals_v3_by_project_id", "", "")
	removeDisabledEventUserPropertiesByProjectId := flag.String("remove_disabled_event_user_properties",
		"", "List of projects to disable event user property population in events.")
	useHashIDForCRMGroupUserByProject := flag.String("use_hash_id_for_crm_group_user_by_project_id", "", "")
	moveHubspotCompanyAssocationFlowToContactByPojectID := flag.String("move_hubspot_company_association_flow_to_contact_by_project_id", "", "")
	enrichPullLimit := flag.Int("enrich_pull_limit", 0, "Total records to be pulled in single db call")
	userPropertyUpdateOptProjects := flag.String("user_property_update_opt_projects", "", "")
	associateDealToDomainByProjectID := flag.String("associate_deal_to_domain_by_project_id", "", "")
	enableSyncTries := flag.Bool("enable_sync_tries", false, "Filter using un-sync document using sync-tries")
	addCRMObjectURLByProjectID := flag.String("add_crm_object_url_by_project_id", "", "")
	firstTimeEnrich := flag.Bool("first_time_enrich", false, "")
	enableTotalSessionPropertiesV2ByProjectID := flag.String("enable_total_session_properties_v2", "", "")
	enableDomainWebsitePropertiesByProjectID := flag.String("enable_domain_website_properties_by_project_id", "", "")
	skipContactUpdatesByProjectID := flag.String("skip_contact_updates_by_project_id", "", "")
	hubspotEnrichBackfillLimit := flag.Int("hubspot_enrich_backfill_limit", 0, "")

	flag.Parse()
	if *env != "development" && *env != "staging" && *env != "production" {
		panic(fmt.Errorf("env [ %s ] not recognised", *env))
	}
	defaultAppName := "hubspot_enrich_job"
	defaultHealthcheckPingID := C.HealthcheckHubspotEnrichPingID
	if *firstTimeEnrich {
		defaultHealthcheckPingID = C.HealthcheckHubspotFirstTimeEnrichPingID
	}

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
			UseExactConnFromConfig: true,
		},
		PrimaryDatastore:                              *primaryDatastore,
		RedisHost:                                     *redisHost,
		RedisPort:                                     *redisPort,
		RedisHostPersistent:                           *redisHostPersistent,
		RedisPortPersistent:                           *redisPortPersistent,
		FactorsSixSignalAPIKey:                        *factorsSixsignalAPIKey,
		SentryDSN:                                     *sentryDSN,
		SentryRollupSyncInSecs:                        *sentryRollupSyncInSecs,
		DryRunCRMSmartEvent:                           *dryRunSmartEvent,
		CacheSortedSet:                                *cacheSortedSet,
		AllowedHubspotGroupsByProjectIDs:              *enableHubspotGroupsByProjectID,
		SkippedOtpProjectIDs:                          *skippedOtpProjectIDs,
		UseSourcePropertyOverwriteByProjectIDs:        *useSourcePropertyOverwriteByProjectID,
		CaptureSourceInUsersTable:                     *captureSourceInUsersTable,
		RestrictReusingUsersByCustomerUserId:          *restrictReusingUsersByCustomerUserId,
		EnableHubspotFormsEventsByProjectID:           *enableHubspotFormEventsByProjectID,
		DisableHubspotNonMarketingContactsByProjectID: *disableNonMarketingContactByProjectID,
		HubspotAppID:                                  *hubspotAppID,
		HubspotAppSecret:                              *hubspotAppSecret,
		AllowIdentificationOverwriteUsingSourceByProjectID:  *allowIdentificationOverwriteUsingSourceByProjectID,
		AllowHubspotPastEventsEnrichmentByProjectID:         *allowHubspotPastEventsEnrichmentByProjectID,
		AllowHubspotContactListInsertByProjectID:            *allowHubspotContactListInsertByProjectID,
		IngestionTimezoneEnabledProjectIDs:                  C.GetTokensFromStringListAsString(*IngestionTimezoneEnabledProjectIDs),
		EnableDomainsGroupByProjectID:                       *enableDomainsGroupByProjectID,
		EnableSyncReferenceFieldsByProjectID:                *enableSyncReferenceFieldsByProjectID,
		EnableUserDomainsGroupByProjectID:                   *enableUserDomainsGroupByProjectID,
		UseHubspotCompaniesV3APIByProjectID:                 *useHubspotCompaniesv3APIByProjectID,
		AllowEmailDomainsByProjectID:                        *allowEmailDomainsByProjectID,
		UseHubspotEngagementsV3APIByProjectID:               *useHubspotEngagementsv3APIByProjectID,
		UseHubspotDealsV3APIByProjectID:                     *useHubspotDealsv3APIByProjectID,
		RemoveDisabledEventUserPropertiesByProjectID:        *removeDisabledEventUserPropertiesByProjectId,
		UseHashIDForCRMGroupUserByProject:                   *useHashIDForCRMGroupUserByProject,
		MoveHubspotCompanyAssocationFlowToContactByPojectID: *moveHubspotCompanyAssocationFlowToContactByPojectID,
		UserPropertyUpdateOptProjects:                       *userPropertyUpdateOptProjects,
		AssociateDealToDomainByProjectID:                    *associateDealToDomainByProjectID,
		EnableSyncTriesFlag:                                 *enableSyncTries,
		AddCRMObjectURLPropertyByProjectID:                  *addCRMObjectURLByProjectID,
		EnableTotalSessionPropertiesV2ByProjectID:           *enableTotalSessionPropertiesV2ByProjectID,
		EnableDomainWebsitePropertiesByProjectID:            *enableDomainWebsitePropertiesByProjectID,
		HubspotEnrichSkipContactUpdatesByProjectID:          *skipContactUpdatesByProjectID,
	}

	C.InitConf(config)
	C.InitSortedSetCache(config.CacheSortedSet)

	err := C.InitDBWithMaxIdleAndMaxOpenConn(*config, 200, 100)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{"env": *env,
			"host": *memSQLHost, "port": *memSQLPort}).Panic("Failed to initialize DB.")
	}
	db := C.GetServices().Db
	defer db.Close()

	C.InitRedis(config.RedisHost, config.RedisPort)
	C.InitRedisPersistent(config.RedisHostPersistent, config.RedisPortPersistent)
	C.InitSentryLogging(config.SentryDSN, config.AppName)
	C.InitMetricsExporter(config.Env, config.AppName, config.GCPProjectID, config.GCPProjectLocation)
	C.InitSmartEventMode(config.DryRunCRMSmartEvent)
	C.InitPropertiesTypeCache(*enablePropertyTypeFromDB, *propertiesTypeCacheSize, *whitelistedProjectIDPropertyTypeFromDB, *blacklistedProjectIDPropertyTypeFromDB)
	defer C.WaitAndFlushAllCollectors(65 * time.Second)

	configsEnrich := make(map[string]interface{})
	configsEnrich["project_ids"] = *projectIDList
	configsEnrich["disabled_project_ids"] = *disabledProjectIDList
	configsEnrich["num_unique_doc_routines"] = *numDocRoutines
	configsEnrich["health_check_ping_id"] = defaultHealthcheckPingID
	configsEnrich["override_healthcheck_ping_id"] = *overrideHealthcheckPingID
	configsEnrich["num_project_routines"] = *numProjectRoutines
	configsEnrich["max_record_created_at"] = *hubspotMaxCreatedAt
	configsEnrich["enrich_heavy"] = *enrichHeavy
	configsEnrich["record_process_limit_per_project"] = *recordProcessLimit
	configsEnrich["enrich_pull_limit"] = *enrichPullLimit
	configsEnrich["first_time_enrich"] = *firstTimeEnrich
	configsEnrich["hubspot_enrich_backfill_limit"] = *hubspotEnrichBackfillLimit

	configsDistributer := make(map[string]interface{})
	configsDistributer["health_check_ping_id"] = ""
	configsDistributer["max_record_created_at"] = *hubspotMaxCreatedAt
	configsDistributer["override_healthcheck_ping_id"] = *overrideHubspotProjectDistributerHealthcheckPingID
	configsDistributer["light_projects_count_threshold"] = *lightProjectsCountThreshold

	// distributer should only run on light job
	if !(*enrichHeavy) && !*firstTimeEnrich {
		taskWrapper.TaskFunc(*projectDistributerAppName, *taskManagementLookback, T.RunHubspotProjectDistributer, configsDistributer)
	}

	taskWrapper.TaskFunc(appName, *taskManagementLookback, T.RunHubspotEnrich, configsEnrich)
}
