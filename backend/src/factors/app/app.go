package main

import (
	C "factors/config"
	H "factors/handler"
	mid "factors/middleware"
	"flag"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// ./app --env=development --api_domain=localhost:8080 --app_domain=localhost:3000  --api_http_port=8080 --etcd=localhost:2379 --db_host=localhost --db_port=5432 --db_user=autometa --db_name=autometa --db_pass=@ut0me7a --geo_loc_path=/usr/local/var/factors/geolocation_data/GeoLite2-City.mmdb --aws_region=us-east-1 --aws_key=dummy --aws_secret=dummy --email_sender=support@factors.ai --error_reporting_interval=300
func main() {

	env := flag.String("env", "development", "")
	port := flag.Int("api_http_port", 8080, "")
	etcd := flag.String("etcd", "localhost:2379", "Comma separated list of etcd endpoints localhost:2379,localhost:2378")

	dbHost := flag.String("db_host", "localhost", "")
	dbPort := flag.Int("db_port", 5432, "")
	dbUser := flag.String("db_user", "autometa", "")
	dbName := flag.String("db_name", "autometa", "")
	dbPass := flag.String("db_pass", "@ut0me7a", "")

	redisHost := flag.String("redis_host", "localhost", "")
	redisPort := flag.Int("redis_port", 6379, "")

	redisHostPersistent := flag.String("redis_host_ps", "localhost", "")
	redisPortPersistent := flag.Int("redis_port_ps", 6379, "")

	geoLocFilePath := flag.String("geo_loc_path", "/usr/local/var/factors/geolocation_data/GeoLite2-City.mmdb", "")
	deviceDetectorPath := flag.String("device_detector_path", "/usr/local/var/factors/devicedetector_data/regexes", "")

	apiDomain := flag.String("api_domain", "factors-dev.com:8080", "")
	appDomain := flag.String("app_domain", "factors-dev.com:3000", "")

	awsRegion := flag.String("aws_region", "us-east-1", "")
	awsAccessKeyId := flag.String("aws_key", "dummy", "")
	awsSecretAccessKey := flag.String("aws_secret", "dummy", "")

	factorsEmailSender := flag.String("email_sender", "support-dev@factors.ai", "")

	adminLoginEmail := flag.String("admin_login_email", "", "Admin email for login")
	adminLoginToken := flag.String("admin_login_token", "", "Admin token for login")
	loginTokenMap := flag.String("login_token_map", "", "Map of token and agent email to authenticate")

	sentryDSN := flag.String("sentry_dsn", "", "Sentry DSN")

	skipTrackProjectIds := flag.String("skip_track_project_ids", "", "List or projects to skip track")
	whitelistedProjectIdsEventUserCache := flag.String("whitelisted_projects_ids_event_user_cache",
		"1", "List of project ids for which caching for events/user is enabled")
	facebookAppId := flag.String("facebook_app_id", "", "")
	facebookAppSecret := flag.String("facebook_app_secret", "", "")
	salesforceAppId := flag.String("salesforce_app_id", "", "")
	salesforceAppSecret := flag.String("salesforce_app_secret", "", "")

	skipSessionProjectIds := flag.String("skip_session_project_ids",
		"", "List or projects to skip session creation.")

	isRealTimeEventUserCachingEnabled := flag.Bool("enable_real_time_event_user_caching",
		false, "If the real time caching is enabled")
	blockedSDKRequestProjectTokens := flag.String("blocked_sdk_request_project_tokens",
		"", "List of project tokens blocked for all sdk requests.")
	cacheLookUpRangeProjects := flag.String("cache_look_up_range_projects",
		"", "List of projects and the overrided date range")
	flag.Parse()

	config := &C.Configuration{
		AppName:       "app_server",
		Env:           *env,
		Port:          *port,
		EtcdEndpoints: strings.Split(*etcd, ","),
		DBInfo: C.DBConf{
			Host:     *dbHost,
			Port:     *dbPort,
			User:     *dbUser,
			Name:     *dbName,
			Password: *dbPass,
		},
		RedisHost:                           *redisHost,
		RedisPort:                           *redisPort,
		RedisHostPersistent:                 *redisHostPersistent,
		RedisPortPersistent:                 *redisPortPersistent,
		GeolocationFile:                     *geoLocFilePath,
		DeviceDetectorPath:                  *deviceDetectorPath,
		APIDomain:                           *apiDomain,
		APPDomain:                           *appDomain,
		AWSKey:                              *awsAccessKeyId,
		AWSSecret:                           *awsSecretAccessKey,
		AWSRegion:                           *awsRegion,
		EmailSender:                         *factorsEmailSender,
		AdminLoginEmail:                     *adminLoginEmail,
		AdminLoginToken:                     *adminLoginToken,
		FacebookAppID:                       *facebookAppId,
		FacebookAppSecret:                   *facebookAppSecret,
		SalesforceAppID:                     *salesforceAppId,
		SalesforceAppSecret:                 *salesforceAppSecret,
		SentryDSN:                           *sentryDSN,
		LoginTokenMap:                       C.ParseConfigStringToMap(*loginTokenMap),                // Map of "<token>": "<agent_email>".
		SkipTrackProjectIds:                 C.GetTokensFromStringListAsUint64(*skipTrackProjectIds), // comma seperated project ids.
		SkipSessionProjectIds:               *skipSessionProjectIds,                                  // comma seperated project ids, supports "*".
		WhitelistedProjectIdsEventUserCache: *whitelistedProjectIdsEventUserCache,
		IsRealTimeEventUserCachingEnabled:   *isRealTimeEventUserCachingEnabled,
		BlockedSDKRequestProjectTokens:      C.GetTokensFromStringListAsString(*blockedSDKRequestProjectTokens),
		CacheLookUpRangeProjects:            C.ExtractProjectIdDateFromConfig(*cacheLookUpRangeProjects),
	}

	// Initialize configs and connections.
	err := C.Init(config)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize.")
		return
	}
	C.InitRedisPersistent(config.RedisHostPersistent, config.RedisPortPersistent)

	if !C.IsDevelopment() {
		gin.SetMode(gin.ReleaseMode)
	}
	defer C.SafeFlushSentryHook()

	r := gin.New()
	// Group based middlewares should be registered on corresponding init methods.
	r.Use(mid.AddSecurityHeadersForAppRoutes())
	// Root middleware for cors.
	r.Use(mid.CustomCors())
	r.Use(mid.RequestIdGenerator())
	r.Use(mid.Logger())
	r.Use(mid.Recovery())

	// Initialize routes.
	H.InitAppRoutes(r)
	H.InitIntRoutes(r)
	r.Run(":" + strconv.Itoa(C.GetConfig().Port))

	// TODO(Ankit):
	// Add graceful shutdown.
	// flush error collector before quitting the process
}
