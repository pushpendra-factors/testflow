package main

import (
	C "factors/config"
	Const "factors/constants"
	H "factors/handler"
	mid "factors/middleware"
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
	memSQLPort := flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser := flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName := flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass := flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	memSQLCertificate := flag.String("memsql_cert", "", "")

	memSQLDBMaxOpenConnections := flag.Int("memsql_max_open_connections", 100, "Max no.of open connections allowed on connection pool of memsql")
	memSQLDBMaxIdleConnections := flag.Int("memsql_max_idle_connections", 50, "Max no.of idle connections allowed on connection pool of memsql")

	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypePostgres, "Primary datastore type as memsql or postgres")
	disableDBWrites := flag.Bool("disable_db_writes", false, "To disable DB writes.")
	disableRedisWrites := flag.Bool("disable_redis_writes", false, "To disable redis writes.")
	disableQueryCache := flag.Bool("disable_query_cache", false, "To disable dashboard and query analytics cache.")

	redisHost := flag.String("redis_host", "localhost", "")
	redisPort := flag.Int("redis_port", 6379, "")

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

	skipTrackProjectIds := flag.String("skip_track_project_ids", "", "List or projects to skip track")
	lookbackWindowForEventUserCache := flag.Int("lookback_window_event_user_cache",
		30, "look back window in cache for event/user cache")
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
	customerEnabledProjectsWeeklyInsights := flag.String("customer_enabled_projects_wi",
		"", "List of projects customer enabled for Weekly Insights")
	demoProjectIds := flag.String("demo_projects", "", "List of projectIds those are used for demo")
	attributionDebug := flag.Int("attribution_debug", 0, "Enables debug logging for attribution queries")
	enableMQLAPI := flag.Bool("enable_mql_api", false, "Enable MQL API routes.")
	enableDemoReadAccess := flag.Bool("enable_demo_read_access", false, "Enable Admin API restrictions.")
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
	resourcePoolForAnalytics := flag.String("resource_pool_for_analytics", "",
		"Given resource_pool will be used for analytics queries.")
	hubspotAPIOnboardingHAPIKey := flag.String("hubspot_API_onboarding_HAPI_key", "", "")
	allowProfilesGroupSupport := flag.String("allow_profiles_group_support", "", "")

	fivetranGroupId := flag.String("fivetran_group_id", "", "")
	fivetranLicenseKey := flag.String("fivetran_license_key", "", "")
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
		PrimaryDatastore:                        *primaryDatastore,
		RedisHost:                               *redisHost,
		RedisPort:                               *redisPort,
		RedisHostPersistent:                     *redisHostPersistent,
		RedisPortPersistent:                     *redisPortPersistent,
		GeolocationFile:                         *geoLocFilePath,
		DeviceDetectorPath:                      *deviceDetectorPath,
		APIDomain:                               *apiDomain,
		APPDomain:                               *appDomain,
		APPOldDomain:                            *appOldDomain,
		AWSKey:                                  *awsAccessKeyId,
		AWSSecret:                               *awsSecretAccessKey,
		AWSRegion:                               *awsRegion,
		EmailSender:                             *factorsEmailSender,
		AdminLoginEmail:                         *adminLoginEmail,
		AdminLoginToken:                         *adminLoginToken,
		FacebookAppID:                           *facebookAppId,
		FacebookAppSecret:                       *facebookAppSecret,
		LinkedinClientID:                        *linkedinClientID,
		LinkedinClientSecret:                    *linkedinClientSecret,
		SalesforceAppID:                         *salesforceAppId,
		SalesforceAppSecret:                     *salesforceAppSecret,
		SentryDSN:                               *sentryDSN,
		LoginTokenMap:                           C.ParseConfigStringToMap(*loginTokenMap),                // Map of "<token>": "<agent_email>".
		SkipTrackProjectIds:                     C.GetTokensFromStringListAsUint64(*skipTrackProjectIds), // comma seperated project ids.
		BlockedSDKRequestProjectTokens:          C.GetTokensFromStringListAsString(*blockedSDKRequestProjectTokens),
		CacheLookUpRangeProjects:                C.ExtractProjectIdDateFromConfig(*cacheLookUpRangeProjects),
		LookbackWindowForEventUserCache:         *lookbackWindowForEventUserCache,
		ActiveFactorsGoalsLimit:                 *factorsActiveGoalsLimit,
		ActiveFactorsTrackedEventsLimit:         *factorsActiveTrackedEventsLimit,
		ActiveFactorsTrackedUserPropertiesLimit: *factorsActiveTrackedUserPropertiesLimit,
		AllowSmartEventRuleCreation:             *allowSmartEventRuleCreation,
		ProjectAnalyticsWhitelistedUUIds:        C.GetUUIdsFromStringListAsString(*projectAnalyticsWhitelistedUUIds),
		CustomerEnabledProjectsWeeklyInsights:   C.GetTokensFromStringListAsUint64(*customerEnabledProjectsWeeklyInsights),
		DemoProjectIds:                          C.GetTokensFromStringListAsUint64(*demoProjectIds),
		EnableDemoReadAccess:                    enableDemoReadAccess,
		EnableMQLAPI:                            *enableMQLAPI,
		DisableDBWrites:                         disableDBWrites,
		DisableRedisWrites:                      disableRedisWrites,
		DisableQueryCache:                       disableQueryCache,
		AttributionDebug:                        *attributionDebug,
		DisableDashboardQueryDBExecution:        *disableDashboardQueryDBExecution,
		EnableFilterOptimisation:                *enableFilterOptimisation,
		FilterPropertiesStartTimestamp:          *filterPropertiesStartTimestamp,
		DevBox:                                  *devBox,
		SkipEventNameStepByProjectID:            *skipEventNameStepByProjectID,
		SkipUserJoinInEventQueryByProjectID:     *skipUserJoinInEventQueryByProjectID,
		EnableEventLevelEventProperties:         *enableEventLevelEventProperties,
		AllowSupportForSourceColumnInUsers:      *allowSupportForSourceColumnInUsers,
		ResourcePoolForAnalytics:                *resourcePoolForAnalytics,
		HubspotAPIOnboardingHAPIKey:             *hubspotAPIOnboardingHAPIKey,
		AllowProfilesGroupSupport:               *allowProfilesGroupSupport,
		FivetranGroupId:                         *fivetranGroupId,
		FivetranLicenseKey:                      *fivetranLicenseKey,
	}
	C.InitConf(config)

	// Initialize configs and connections.
	err := C.InitAppServer(config)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize.")
		return
	}
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
	H.InitAppRoutes(r)
	H.InitIntRoutes(r)
	Const.SetSmartPropertiesReservedNames()

	C.KillDBQueriesOnExit()
	r.Run(":" + strconv.Itoa(C.GetConfig().Port))

	// TODO(Ankit):
	// Add graceful shutdown.
	// flush error collector before quitting the process
}
