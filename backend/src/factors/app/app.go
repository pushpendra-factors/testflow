package main

import (
	C "factors/config"
	DD "factors/default_data"
	H "factors/handler"
	mid "factors/middleware"
	"factors/model/model"
	session "factors/session/store"
	U "factors/util"
	"flag"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	swaggerDocs "factors/docs"
)

// ./app --env=development --api_domain=localhost:8080 --app_domain=localhost:3000  --api_http_port=8080 --etcd=localhost:2379 --db_host=localhost --db_port=5432 --db_user=autometa --db_name=autometa --db_pass=@ut0me7a --geo_loc_path=/usr/local/var/factors/geolocation_data/GeoLite2-City.mmdb --aws_region=us-east-1 --aws_key=dummy --aws_secret=dummy --email_sender=support@factors.ai --error_reporting_interval=300
// @title Factors Backend Api
// @version 1.0
// @description Factors usage doc for golang api server.
// @BasePath /projects
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {
	gcpProjectID := flag.String("gcp_project_id", "", "Project ID on Google Cloud")
	gcpProjectLocation := flag.String("gcp_project_location", "", "Location of google cloud project cluster")

	env := flag.String("env", "development", "")
	port := flag.Int("api_http_port", 8080, "")
	etcd := flag.String("etcd", "localhost:2379", "Comma separated list of etcd endpoints localhost:2379,localhost:2378")

	dbHost := flag.String("db_host", C.PostgresDefaultDBParams.Host, "")
	dbPort := flag.Int("db_port", C.PostgresDefaultDBParams.Port, "")
	dbUser := flag.String("db_user", C.PostgresDefaultDBParams.User, "")
	dbName := flag.String("db_name", C.PostgresDefaultDBParams.Name, "")
	dbPass := flag.String("db_pass", C.PostgresDefaultDBParams.Password, "")

	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	isPSCHost := flag.Int("memsql_is_psc_host", C.MemSQLDefaultDBParams.IsPSCHost, "")
	memSQLPort := flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser := flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName := flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass := flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	memSQLCertificate := flag.String("memsql_cert", "", "")

	memSQLDBMaxOpenConnections := flag.Int("memsql_max_open_connections", 100, "Max no.of open connections allowed on connection pool of memsql")
	memSQLDBMaxIdleConnections := flag.Int("memsql_max_idle_connections", 50, "Max no.of idle connections allowed on connection pool of memsql")

	enableDBConnectionPool2 := flag.Bool("enable_db_conn_pool2", false, "")
	memSQLHost2 := flag.String("memsql_host_2", C.MemSQLDefaultDBParams.Host, "")
	isPSCHost2 := flag.Int("memsql_is_psc_host_2", C.MemSQLDefaultDBParams.IsPSCHost, "")
	memSQLPort2 := flag.Int("memsql_port_2", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser2 := flag.String("memsql_user_2", C.MemSQLDefaultDBParams.User, "")
	memSQLName2 := flag.String("memsql_name_2", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass2 := flag.String("memsql_pass_2", C.MemSQLDefaultDBParams.Password, "")
	memSQLCertificate2 := flag.String("memsql_cert_2", "", "")

	memSQLDBMaxOpenConnections2 := flag.Int("memsql_max_open_connections_2", 100, "Max no.of open connections allowed on connection pool of memsql")
	memSQLDBMaxIdleConnections2 := flag.Int("memsql_max_idle_connections_2", 50, "Max no.of idle connections allowed on connection pool of memsql")

	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypeMemSQL, "Primary datastore type as memsql or postgres")
	disableDBWrites := flag.Bool("disable_db_writes", false, "To disable DB writes.")
	disableQueryCache := flag.Bool("disable_query_cache", false, "To disable dashboard and query analytics cache.")

	redisHost := flag.String("redis_host", "localhost", "")
	redisPort := flag.Int("redis_port", 6379, "")

	queueRedisHost := flag.String("queue_redis_host", "localhost", "")
	queueRedisPort := flag.Int("queue_redis_port", 6379, "")
	useQueueRedis := flag.Bool("use_queue_redis", false, "Use queue redis for sdk related caching.")

	duplicateQueueRedisHost := flag.String("dup_queue_redis_host", "localhost", "")
	duplicateQueueRedisPort := flag.Int("dup_queue_redis_port", 6379, "")
	sdkQueueThreshold := flag.Int("sdk_queue_threshold", 10000, "Threshold to report sdk queue size")
	integrationQueueThreshold := flag.Int("integration_queue_threshold", 1000, "Threshold to report integration queue size")
	delayedTaskThreshold := flag.Int("delayed_task_threshold", 1000, "Threshold to report delayed task size")
	enableSDKAndIntegrationRequestQueueDuplication := flag.Bool("enable_sdk_and_integration_request_queue_duplication",
		false, "Enables SDK and Integration request queue duplication monitoring.")
	monitoringAPIToken := flag.String("monitoring_api_token", "", "enter  monitoring api token")

	redisHostPersistent := flag.String("redis_host_ps", "localhost", "")
	redisPortPersistent := flag.Int("redis_port_ps", 6379, "")

	geoLocFilePath := flag.String("geo_loc_path", "/usr/local/var/factors/geolocation_data/GeoLite2-City.mmdb", "")
	deviceDetectorPath := flag.String("device_detector_path", "/usr/local/var/factors/devicedetector_data/regexes", "")

	apiDomain := flag.String("api_domain", "factors-dev.com:8080", "")
	appDomain := flag.String("app_domain", "factors-dev.com:3000", "")
	appOldDomain := flag.String("app_old_domain", "factors-dev.com:3000", "")

	awsRegion := flag.String("aws_region", "us-east-1", "")
	awsAccessKeyId := flag.String("aws_key", "dummy", "")
	awsSecretAccessKey := flag.String("aws_secret", "dummy", "")

	factorsEmailSender := flag.String("email_sender", "support-dev@factors.ai", "")

	adminLoginEmail := flag.String("admin_login_email", "", "Admin email for login")
	adminLoginToken := flag.String("admin_login_token", "", "Admin token for login")
	loginTokenMap := flag.String("login_token_map", "", "Map of token and agent email to authenticate")

	sentryDSN := flag.String("sentry_dsn", "", "Sentry DSN")
	sentryRollupSyncInSecs := flag.Int("sentry_rollup_sync_in_seconds", 300, "Enables to send errors to sentry in given interval.")

	skipTrackProjectIds := flag.String("skip_track_project_ids", "", "List or projects to skip track")
	lookbackWindowForEventUserCache := flag.Int("lookback_window_event_user_cache",
		20, "look back window in cache for event/user cache")
	facebookAppId := flag.String("facebook_app_id", "", "")
	facebookAppSecret := flag.String("facebook_app_secret", "", "")
	linkedinClientID := flag.String("linkedin_client_id", "", "")
	linkedinClientSecret := flag.String("linkedin_client_secret", "", "")
	salesforceAppId := flag.String("salesforce_app_id", "", "")
	salesforceAppSecret := flag.String("salesforce_app_secret", "", "")

	blockedSDKRequestProjectTokens := flag.String("blocked_sdk_request_project_tokens",
		"", "List of project tokens blocked for all sdk requests.")
	cacheLookUpRangeProjects := flag.String("cache_look_up_range_projects",
		"", "List of projects and the overrided date range")
	factorsActiveGoalsLimit := flag.Int("active_goals_limit",
		50, "Active Goals limit per project")
	factorsActiveTrackedEventsLimit := flag.Int("active_tracked_events_limit",
		50, "Active Tracked events limit per project")
	factorsActiveTrackedUserPropertiesLimit := flag.Int("active_tracked_user_properties_limit",
		50, "Active Tracked user properties limit per project")
	allowSmartEventRuleCreation := flag.Bool("allow_smart_event_rule_creation", false, "Should allow smart event rule creation")
	projectAnalyticsWhitelistedUUIds := flag.String("project_analytics_whitelisted_uuids",
		"", "List of UUIDs whitelisted for project analytics API")
	customerEnabledProjectsLastComputed := flag.String("customer_enabled_projects_last_computed",
		"*", "List of projects customer enabled forLast Computed")
	attributionDebug := flag.Int("attribution_debug", 0, "Enables debug logging for attribution queries")
	attributionDBCacheLookup := flag.String("attribution_db_cache_lookup", "", "For given projects, Lookup for cache results in DB for dashboard queries")
	attributionCommonFlow := flag.String("attribution_common_flow", "", "For given projects, run attribution queries with common flow for "+
		"dashboard and normal query. Both flow will check DB, cache based on week, months and so on..")
	attributionDebugKPI := flag.String("attribution_debug_kpi", "ignore", "Attribution Debug KPI ID.")
	enableMQLAPI := flag.Bool("enable_mql_api", false, "Enable MQL API routes.")
	overrideAppName := flag.String("app_name", "", "Override default app_name.")

	disableDashboardQueryDBExecution := flag.Bool("disable_dashboard_query_db_execution", false,
		"Disable direct execution of query from dashboard, if not available on cache.")

	bucketName := flag.String("bucket_name", "/usr/local/var/factors/cloud_storage", "")

	enableFilterOptimisation := flag.Bool("enable_filter_optimisation", false,
		"Enables filter optimisation changes for memsql implementation.")
	filterPropertiesStartTimestamp := flag.Int64("filter_properties_start_timestamp", -1,
		"Start timestamp of data available for filtering with parquet on memsql.")
	devBox := flag.Bool("dev_box", false, "Is this going to be deployed on one box")
	skipEventNameStepByProjectID := flag.String("skip_event_name_step_by_project_id", "", "")
	skipUserJoinInEventQueryByProjectID := flag.String("skip_user_join_in_event_query_by_project_id", "", "")
	enableEventLevelEventProperties := flag.String("enable_event_level_event_properties", "", "")
	allowSupportForSourceColumnInUsers := flag.String("allow_support_for_source_column_in_users", "", "")
	allowSupportForV1AvgKPIComputation := flag.String("allow_support_for_v1_avg_kpi_computation", "", "")

	resourcePoolForAnalytics := flag.String("resource_pool_for_analytics", "",
		"Given resource_pool will be used for analytics queries.")
	hubspotAPIOnboardingHAPIKey := flag.String("hubspot_API_onboarding_HAPI_key", "", "")
	hubspotAPIOnboardingPrivateAccessToken := flag.String("hubspot_API_onboarding_private_access_token", "", "")
	mailmodoOnboardingAPIKey := flag.String("mailmodo_onboarding_API_key", "TJ5JF61-44NMRN5-GAEA2WH-8Z99P4H", "")
	mailmodoOnboardingURL1 := flag.String("mailmodo_onboarding_URL1", "https://api.mailmodo.com/hooks/start/1df3694b-8651-441f-a9ce-2f64d5e6b6ff", "")
	mailmodoOnboardingURL2 := flag.String("mailmodo_onboarding_URL2", "https://api.mailmodo.com/hooks/start/ef8af6d0-e925-47e2-8c03-2b010c9a59f5", "")
	slackOnboardingWebhookURL := flag.String("slack_onboarding_webhook_url", "https://hooks.slack.com/services/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", "")
	allowProfilesGroupSupport := flag.String("allow_profiles_group_support", "", "")

	auth0ClientID := flag.String("auth0_client_id", "", "")
	auth0ClientSecret := flag.String("auth0_client_secret", "", "")
	auth0Domain := flag.String("auth0_domain", "", "")
	auth0CallbackURL := flag.String("callback_url", "", "")

	sessionStore := flag.String("session_store", "cookie", "")
	sessionStoreSecret := flag.String("session_store_secret", "", "")

	fivetranGroupId := flag.String("fivetran_group_id", "", "")
	fivetranLicenseKey := flag.String("fivetran_license_key", "", "")
	allowEventsFunnelsGroupSupport := flag.String("allow_events_funnels_group_support", "", "")

	enableBingAdsAttribution := flag.Bool("enable_bing_ads_attribution", false, "")
	salesforcePropertyLookBackTimeHr := flag.Int("salesforce_property_lookback_time_hr", 0, "")
	hubspotPropertyLookbackLimit := flag.Int("hubspot_property_lookback_limit", 1000, "")
	enableSlowDBQueryLogging := flag.Bool("log_slow_db_queries", false, "Logs queries with execution time greater than 50ms.")

	slackAppClientID := flag.String("slack_app_client_id", "", "")
	slackAppClientSecret := flag.String("slack_app_client_secret", "", "")

	dataAvailabilityExpiry := flag.Int("data_availability_expiry", 30, "")

	enableOptimisedFilterOnProfileQuery := flag.Int("enable_optimised_filter_on_profile_query",
		0, "Enables filter optimisation logic for profiles query.")
	hubspotAppID := flag.String("hubspot_app_id", "", "Hubspot app id for oauth integration")
	hubspotAppSecret := flag.String("hubspot_app_secret", "", "Hubspot app secret for oauth integration")
	enableOptimisedFilterOnEventUserQuery := flag.Int("enable_optimised_filter_on_event_user_query",
		0, "Enables filter optimisation logic for events and users query.")
	enableEmailBlocking := flag.Bool("enable_email_blocking", true, "Blocks signup from emails in the blocked_email_list")
	enableIPBlocking := flag.Bool("enable_IP_blocking", true, "Blocks access from IPs in the blocked_IP_list")
	blockedEmailList := flag.String("blocked_email_list", "", "List containing all the blocked emails")
	blockedIPList := flag.String("blocked_IP_list", "", "List containing all the blocked IP address")
	blockedEmailDomainList := flag.String("blocked_email_domain_list", "", "List containing all blocked email domains")
	allAccountsProjectId := flag.String("all_accounts_project_id", "", "List of projectIds to enable domain.")
	timelinesTablePropsQueryOpt := flag.String("timelines_table_props_query_opt", "", "List of projectIds to enable timelines table props optimised query.")
	markerPreviewAllAccountsProjectId := flag.String("marker_preview_all_accounts_project_id", "", "List of projectIds to enable preview using marker.")
	batchSizePreviewDomain := flag.Int("batch_size_preview_domain", 100, "Batch size for goroutines to process domains for preview using marker.")
	accountsToProcessForPreview := flag.Int("accounts_to_process_for_preview", 5000, "No of domains to process domains for preview using marker per run.")
	numberOfRunsForPreview := flag.Int("number_of_runs_for_preview", 10, "No of runs to process domains for preview using marker per run.")
	accountLimitPreviewListing := flag.Int("account_limit_preview_listing", 100, "No of accounts to show for preview using marker per run.")
	useMarkerByProjectID := flag.String("use_marker_by_project_id", "", "List of projectIds to enable segment marker.")
	useOptimisedEventsQueryProjectIDs := flag.String("use_optimised_events_query_project_ids", "",
		"Project Id to enable optimised query for event based filters check For Marker. A comma separated list of project Ids and supports '*' for all projects. ex: 1,2,6,9")
	enableNewAllAccountsByProjectID := flag.String("enable_new_all_accounts_by_project_id", "", "List of projectIds to enable domain.")
	IngestionTimezoneEnabledProjectIDs := flag.String("ingestion_timezone_enabled_projects", "", "List of projectIds whose ingestion timezone is enabled.")
	enableEventFiltersInSegments := flag.Bool("enable_event_filters_in_segments", false, "Enables adding event filters in segment query")
	enableFeatureGates := flag.Bool("enable_feature_gates", false, "Enable Feature Gates")
	websiteAggregationTestEnabledProjects := flag.String("website_aggregation_test_enabled_projects", "", "Flag - website aggregation test enabled projects")

	teamsAppTenantID := flag.String("teams_app_tenant_id", "", "")
	teamsAppClientID := flag.String("teams_app_client_id", "", "")
	teamsAppClientSecret := flag.String("teams_app_client_secret", "", "")
	teamsApplicationID := flag.String("teams_application_id", "", "")
	enableSyncReferenceFieldsByProjectID := flag.String("enable_sync_reference_fields_by_project_id", "", "")
	allowEventAnalyticsGroupsByProjectID := flag.String("allow_event_analytics_groups_by_project_id", "", "")
	enableFeatureGatesV2 := flag.Bool("enable_feature_gates_v2", false, "")
	enableScoreByProjectID := flag.String("enable_score_by_project_id", "", "List of projectIds with scoring enabled.")
	explainV3Query := flag.Bool("explain_v3_query", false, "whether to implement new query payload")
	chargebeeApiKey := flag.String("chargebee_api_key", "dummy", "Chargebee api key")
	chargebeeSiteName := flag.String("chargebee_site_name", "dummy", "Chargebee site name")
	aggrEventPropertyValuesCacheByProjectID := flag.String("aggr_event_property_values_project_ids", "", "")
	paragonSigningKey := flag.String("paragon_signing_key", "", "")
	paragonProjectID := flag.String("paragon_project_id", "", "")
	clearbitAccProvisionKey := flag.String("cb_acc_provision_key", "dummy", "")
	emailUTMParameterAllowedProjects := flag.String("email_utm_parameter_allowed_projects", "", "")
	enableCacheDBWriteProjects := flag.String("cache_db_write_projects", "", "")
	enableCacheDBReadProjects := flag.String("cache_db_read_projects", "", "")
	chatDebug := flag.Int("chat_debug", 0, "")
	flag.Parse()

	defaultAppName := "app_server"
	appName := C.GetAppName(defaultAppName, *overrideAppName)

	config := &C.Configuration{
		GCPProjectID:       *gcpProjectID,
		GCPProjectLocation: *gcpProjectLocation,
		AppName:            appName,
		Env:                *env,
		Port:               *port,
		EtcdEndpoints:      strings.Split(*etcd, ","),
		DBInfo: C.DBConf{
			Host:     *dbHost,
			Port:     *dbPort,
			User:     *dbUser,
			Name:     *dbName,
			Password: *dbPass,
			AppName:  appName,
		},
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
		MemSQL2Info: C.DBConf{
			Host:        *memSQLHost2,
			IsPSCHost:   *isPSCHost2,
			Port:        *memSQLPort2,
			User:        *memSQLUser2,
			Name:        *memSQLName2,
			Password:    *memSQLPass2,
			Certificate: *memSQLCertificate2,
			AppName:     appName,

			MaxOpenConnections:     *memSQLDBMaxOpenConnections2,
			MaxIdleConnections:     *memSQLDBMaxIdleConnections2,
			UseExactConnFromConfig: true,
		},
		Auth0Info: C.Auth0Conf{
			Domain:       *auth0Domain,
			ClientId:     *auth0ClientID,
			ClientSecret: *auth0ClientSecret,
			CallbackUrl:  *auth0CallbackURL,
		},
		SessionStore:                                   *sessionStore,
		SessionStoreSecret:                             *sessionStoreSecret,
		PrimaryDatastore:                               *primaryDatastore,
		RedisHost:                                      *redisHost,
		RedisPort:                                      *redisPort,
		RedisHostPersistent:                            *redisHostPersistent,
		RedisPortPersistent:                            *redisPortPersistent,
		GeolocationFile:                                *geoLocFilePath,
		DeviceDetectorPath:                             *deviceDetectorPath,
		APIDomain:                                      *apiDomain,
		APPDomain:                                      *appDomain,
		APPOldDomain:                                   *appOldDomain,
		AWSKey:                                         *awsAccessKeyId,
		AWSSecret:                                      *awsSecretAccessKey,
		AWSRegion:                                      *awsRegion,
		EmailSender:                                    *factorsEmailSender,
		AdminLoginEmail:                                *adminLoginEmail,
		AdminLoginToken:                                *adminLoginToken,
		FacebookAppID:                                  *facebookAppId,
		FacebookAppSecret:                              *facebookAppSecret,
		LinkedinClientID:                               *linkedinClientID,
		LinkedinClientSecret:                           *linkedinClientSecret,
		SalesforceAppID:                                *salesforceAppId,
		SalesforceAppSecret:                            *salesforceAppSecret,
		SentryDSN:                                      *sentryDSN,
		LoginTokenMap:                                  C.ParseConfigStringToMap(*loginTokenMap),                // Map of "<token>": "<agent_email>".
		SkipTrackProjectIds:                            C.GetTokensFromStringListAsUint64(*skipTrackProjectIds), // comma seperated project ids.
		BlockedSDKRequestProjectTokens:                 C.GetTokensFromStringListAsString(*blockedSDKRequestProjectTokens),
		CacheLookUpRangeProjects:                       C.ExtractProjectIdDateFromConfig(*cacheLookUpRangeProjects),
		LookbackWindowForEventUserCache:                *lookbackWindowForEventUserCache,
		ActiveFactorsGoalsLimit:                        *factorsActiveGoalsLimit,
		ActiveFactorsTrackedEventsLimit:                *factorsActiveTrackedEventsLimit,
		ActiveFactorsTrackedUserPropertiesLimit:        *factorsActiveTrackedUserPropertiesLimit,
		AllowSmartEventRuleCreation:                    *allowSmartEventRuleCreation,
		ProjectAnalyticsWhitelistedUUIds:               C.GetUUIdsFromStringListAsString(*projectAnalyticsWhitelistedUUIds),
		CustomerEnabledProjectsLastComputed:            C.GetTokensFromStringListAsUint64(*customerEnabledProjectsLastComputed),
		EnableMQLAPI:                                   *enableMQLAPI,
		DisableDBWrites:                                disableDBWrites,
		DisableQueryCache:                              disableQueryCache,
		AttributionDebug:                               *attributionDebug,
		AttributionCommonFlow:                          *attributionCommonFlow,
		AttributionDBCacheLookup:                       *attributionDBCacheLookup,
		AttributionDebugKPI:                            *attributionDebugKPI,
		DisableDashboardQueryDBExecution:               *disableDashboardQueryDBExecution,
		EnableFilterOptimisation:                       *enableFilterOptimisation,
		FilterPropertiesStartTimestamp:                 *filterPropertiesStartTimestamp,
		DevBox:                                         *devBox,
		SkipEventNameStepByProjectID:                   *skipEventNameStepByProjectID,
		SkipUserJoinInEventQueryByProjectID:            *skipUserJoinInEventQueryByProjectID,
		EnableEventLevelEventProperties:                *enableEventLevelEventProperties,
		AllowSupportForSourceColumnInUsers:             *allowSupportForSourceColumnInUsers,
		AllowSupportForV1AvgKPIComputation:             *allowSupportForV1AvgKPIComputation,
		ResourcePoolForAnalytics:                       *resourcePoolForAnalytics,
		HubspotAPIOnboardingHAPIKey:                    *hubspotAPIOnboardingHAPIKey,
		HubspotAPIOnboardingPrivateAccessToken:         *hubspotAPIOnboardingPrivateAccessToken,
		MailModoOnboardingAPIKey:                       *mailmodoOnboardingAPIKey,
		MailModoOnboardingURL1:                         *mailmodoOnboardingURL1,
		MailModoOnboardingURL2:                         *mailmodoOnboardingURL2,
		SlackOnboardingWebhookURL:                      *slackOnboardingWebhookURL,
		AllowProfilesGroupSupport:                      *allowProfilesGroupSupport,
		FivetranGroupId:                                *fivetranGroupId,
		FivetranLicenseKey:                             *fivetranLicenseKey,
		AllowEventsFunnelsGroupSupport:                 *allowEventsFunnelsGroupSupport,
		QueueRedisHost:                                 *queueRedisHost,
		QueueRedisPort:                                 *queueRedisPort,
		EnableSDKAndIntegrationRequestQueueDuplication: *enableSDKAndIntegrationRequestQueueDuplication,
		DuplicateQueueRedisHost:                        *duplicateQueueRedisHost,
		DuplicateQueueRedisPort:                        *duplicateQueueRedisPort,
		DelayedTaskThreshold:                           *delayedTaskThreshold,
		SdkQueueThreshold:                              *sdkQueueThreshold,
		IntegrationQueueThreshold:                      *integrationQueueThreshold,
		EnableBingAdsAttribution:                       *enableBingAdsAttribution,
		MonitoringAPIToken:                             *monitoringAPIToken,
		UseQueueRedis:                                  *useQueueRedis,
		SalesforcePropertyLookBackTimeHr:               *salesforcePropertyLookBackTimeHr,
		HubspotPropertyLookBackLimit:                   *hubspotPropertyLookbackLimit,
		EnableSlowDBQueryLogging:                       *enableSlowDBQueryLogging,
		SlackAppClientID:                               *slackAppClientID,
		SlackAppClientSecret:                           *slackAppClientSecret,
		DataAvailabilityExpiry:                         *dataAvailabilityExpiry,
		EnableOptimisedFilterOnProfileQuery:            *enableOptimisedFilterOnProfileQuery != 0,
		HubspotAppID:                                   *hubspotAppID,
		HubspotAppSecret:                               *hubspotAppSecret,
		EnableOptimisedFilterOnEventUserQuery:          *enableOptimisedFilterOnEventUserQuery != 0,
		EnableEmailBlockingFlag:                        *enableEmailBlocking,
		EnableIPBlockingFlag:                           *enableIPBlocking,
		BlockedEmailList:                               C.GetBlockedEmailFromStringListAsString(*blockedEmailList),
		BlockedIPList:                                  C.GetBlockedIPFromStringListAsString(*blockedIPList),
		BlockedEmailDomainList:                         C.GetBlockedEmailDomainFromStringListAsString(*blockedEmailDomainList),
		AllAccountsProjectId:                           *allAccountsProjectId,
		TimelinesTablePropsQueryOpt:                    *timelinesTablePropsQueryOpt,
		MarkerPreviewAllAccountsProjectId:              *markerPreviewAllAccountsProjectId,
		BatchSizePreviewDomain:                         *batchSizePreviewDomain,
		AccountsToProcessForPreview:                    *accountsToProcessForPreview,
		NumberOfRunsForPreview:                         *numberOfRunsForPreview,
		AccountLimitPreviewListing:                     *accountLimitPreviewListing,
		UseMarkerByProjectID:                           *useMarkerByProjectID,
		UseOptimisedEventsQueryProjectIDs:              *useOptimisedEventsQueryProjectIDs,
		IngestionTimezoneEnabledProjectIDs:             C.GetTokensFromStringListAsString(*IngestionTimezoneEnabledProjectIDs),
		EnableEventFiltersInSegments:                   *enableEventFiltersInSegments,
		EnableFeatureGates:                             *enableFeatureGates,
		EnableDBConnectionPool2:                        *enableDBConnectionPool2,
		SentryRollupSyncInSecs:                         *sentryRollupSyncInSecs,
		TeamsAppTenantID:                               *teamsAppTenantID,
		TeamsAppClientID:                               *teamsAppClientID,
		TeamsAppClientSecret:                           *teamsAppClientSecret,
		TeamsApplicationID:                             *teamsApplicationID,
		EnableSyncReferenceFieldsByProjectID:           *enableSyncReferenceFieldsByProjectID,
		AllowEventAnalyticsGroupsByProjectID:           *allowEventAnalyticsGroupsByProjectID,
		EnableFeatureGatesV2:                           *enableFeatureGatesV2,
		EnableScoringByProjectID:                       *enableScoreByProjectID,
		ExplainV3QueryBuilder:                          *explainV3Query,
		EnableNewAllAccountsByProjectID:                *enableNewAllAccountsByProjectID,
		ChargebeeApiKey:                                *chargebeeApiKey,
		ChargebeeSiteName:                              *chargebeeSiteName,
		AggrEventPropertyValuesCacheByProjectID:        *aggrEventPropertyValuesCacheByProjectID,
		ParagonTokenSigningKey:                         *paragonSigningKey,
		ParagonProjectID:                               *paragonProjectID,
		ClearbitProvisionAccountAPIKey:                 *clearbitAccProvisionKey,
		EmailUTMParameterAllowedProjects:               *emailUTMParameterAllowedProjects,
		EnableCacheDBWriteProjects:                     *enableCacheDBWriteProjects,
		EnableCacheDBReadProjects:                      *enableCacheDBReadProjects,
		WebsiteAggregationTestEnabledProjects:          *websiteAggregationTestEnabledProjects,
		ChatDebug:                                      *chatDebug,
	}
	C.InitConf(config)

	// Initialize configs and connections.
	err := C.InitAppServer(config)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize.")
		return
	}

	CheckIfDefaultDatasAreCorrect()
	C.InitMonitoringAPIServices(config)
	C.InitRedisPersistent(config.RedisHostPersistent, config.RedisPortPersistent)
	C.InitFilemanager(*bucketName, *env, config)

	if !C.IsDevelopment() {
		gin.SetMode(gin.ReleaseMode)
	}
	defer C.SafeFlushAllCollectors()
	defer U.NotifyOnPanicWithError(*env, appName)

	r := gin.New()
	// Group based middlewares should be registered on corresponding init methods.
	r.Use(mid.AddSecurityHeadersForAppRoutes())
	// Root middleware for cors.
	r.Use(mid.CustomCors())
	r.Use(mid.RequestIdGenerator())
	r.Use(mid.Logger())
	r.Use(mid.Recovery())

	// Initialize routes.
	if config.Env == C.DEVELOPMENT {
		swaggerDocs.SwaggerInfo.Host = "factors-dev.com:8080"
	} else if config.Env == C.STAGING {
		swaggerDocs.SwaggerInfo.Host = "staging-api.factors.ai"
	}
	err = session.GetSessionStore().InitSessionStore(r)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize session store.")
		return
	}
	H.InitAppRoutes(r)
	H.InitIntRoutes(r)

	if *auth0ClientID != "" && *auth0ClientSecret != "" && *auth0Domain != "" && *auth0CallbackURL != "" {
		authenticator, err := H.NewAuth()
		if err != nil {
			log.WithError(err).Fatal("Failed to initialize auth.")
			return
		}
		H.InitExternalAuth(r, authenticator)
	}

	model.SetSmartPropertiesReservedNames()

	C.KillDBQueriesOnExit()
	r.Run(":" + strconv.Itoa(C.GetConfig().Port))
}

func CheckIfDefaultDatasAreCorrect() {
	if DD.CheckIfDefaultKPIDatasAreCorrect() {
		return
	}
	log.Warn("Failed because defaultDatas and transformations are of incorrect length.")
}
