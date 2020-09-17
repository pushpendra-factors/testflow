package main

import (
	C "factors/config"
	H "factors/handler"

	"flag"
	"strconv"

	mid "factors/middleware"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func main() {
	env := flag.String("env", "development", "")
	port := flag.Int("port", 8085, "")

	dbHost := flag.String("db_host", "localhost", "")
	dbPort := flag.Int("db_port", 5432, "")
	dbUser := flag.String("db_user", "autometa", "")
	dbName := flag.String("db_name", "autometa", "")
	dbPass := flag.String("db_pass", "@ut0me7a", "")

	geoLocFilePath := flag.String("geo_loc_path",
		"/usr/local/var/factors/geolocation_data/GeoLite2-City.mmdb", "")

	deviceDetectorPath := flag.String("device_detector_path", "/usr/local/var/factors/devicedetector_data/regexes", "")

	redisHost := flag.String("redis_host", "localhost", "")
	redisPort := flag.Int("redis_port", 6379, "")

	queueRedisHost := flag.String("queue_redis_host", "localhost", "")
	queueRedisPort := flag.Int("queue_redis_port", 6379, "")

	sentryDSN := flag.String("sentry_dsn", "", "Sentry DSN")

	sdkRequestQueueProjectTokens := flag.String("sdk_request_queue_project_tokens", "",
		"List of project tokens allowed to use sdk request queue")
	segmentRequestQueueProjectTokens := flag.String("segment_request_queue_project_tokens", "",
		"List of project tokens allowed to use segment request queue")

	mergeUserPropertiesProjectIDS := flag.String("merge_usp_project_ids", "",
		"Comma separated list of project IDs for which user properties merge is enabled. '*' for all.")
	skipSessionProjectIds := flag.String("skip_session_project_ids",
		"", "List or projects to create session offline.")
	redisHostPersistent := flag.String("redis_host_ps", "localhost", "")
	redisPortPersistent := flag.Int("redis_port_ps", 6379, "")
	isRealTimeEventUserCachingEnabled := flag.Bool("enable_real_time_event_user_caching",
		true, "If the real time caching is enabled")
	realTimeEventUserCachingProjectIds := flag.String("real_time_event_user_caching_project_ids", "1",
		"If the real time caching is enabled and the whitelisted projectids")
	blockedSDKRequestProjectTokens := flag.String("blocked_sdk_request_project_tokens",
		"", "List of tokens (public and private) to block SDK requests.")
	flag.Parse()

	config := &C.Configuration{
		AppName: "sdk_server",
		Env:     *env,
		Port:    *port,
		DBInfo: C.DBConf{
			Host:     *dbHost,
			Port:     *dbPort,
			User:     *dbUser,
			Name:     *dbName,
			Password: *dbPass,
		},
		GeolocationFile:                    *geoLocFilePath,
		DeviceDetectorPath:                 *deviceDetectorPath,
		RedisHost:                          *redisHost,
		RedisPort:                          *redisPort,
		QueueRedisHost:                     *queueRedisHost,
		QueueRedisPort:                     *queueRedisPort,
		SentryDSN:                          *sentryDSN,
		SDKRequestQueueProjectTokens:       C.GetTokensFromStringListAsString(*sdkRequestQueueProjectTokens), // comma seperated project tokens.
		SegmentRequestQueueProjectTokens:   C.GetTokensFromStringListAsString(*segmentRequestQueueProjectTokens),
		MergeUspProjectIds:                 *mergeUserPropertiesProjectIDS,
		SkipSessionProjectIds:              *skipSessionProjectIds, // comma seperated project ids, supports "*".
		RedisHostPersistent:                *redisHostPersistent,
		RedisPortPersistent:                *redisPortPersistent,
		IsRealTimeEventUserCachingEnabled:  *isRealTimeEventUserCachingEnabled,
		RealTimeEventUserCachingProjectIds: *realTimeEventUserCachingProjectIds,
		// List of tokens (public and private) to block SDK requests.
		BlockedSDKRequestProjectTokens: C.GetTokensFromStringListAsString(*blockedSDKRequestProjectTokens),
	}

	err := C.InitSDKService(config)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize.")
		return
	}
	defer C.SafeFlushSentryHook()

	if !C.IsDevelopment() {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(mid.CustomCors())
	r.Use(mid.RequestIdGenerator())
	r.Use(mid.Logger())
	r.Use(mid.Recovery())

	H.InitSDKServiceRoutes(r)
	r.Run(":" + strconv.Itoa(C.GetConfig().Port))
}
