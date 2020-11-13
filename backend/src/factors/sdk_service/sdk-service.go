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
	gcpProjectID := flag.String("gcp_project_id", "", "Project ID on Google Cloud")
	gcpProjectLocation := flag.String("gcp_project_location", "", "Location of google cloud project cluster")

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

	redisHostPersistent := flag.String("redis_host_ps", "localhost", "")
	redisPortPersistent := flag.Int("redis_port_ps", 6379, "")

	sentryDSN := flag.String("sentry_dsn", "", "Sentry DSN")

	sdkRequestQueueProjectTokens := flag.String("sdk_request_queue_project_tokens", "",
		"List of project tokens allowed to use sdk request queue")
	segmentRequestQueueProjectTokens := flag.String("segment_request_queue_project_tokens", "",
		"List of project tokens allowed to use segment request queue")

	useDefaultProjectSettingForSDK := flag.Bool("use_defaul_project_setting_for_sdk",
		false, "Once set to true, it will skip db query to get project_settings, if not found on cache.")

	blockedSDKRequestProjectTokens := flag.String("blocked_sdk_request_project_tokens",
		"", "List of tokens (public and private) to block SDK requests.")
	flag.Parse()

	config := &C.Configuration{
		GCPProjectID:       *gcpProjectID,
		GCPProjectLocation: *gcpProjectLocation,
		AppName:            "sdk_server",
		Env:                *env,
		Port:               *port,
		DBInfo: C.DBConf{
			Host:     *dbHost,
			Port:     *dbPort,
			User:     *dbUser,
			Name:     *dbName,
			Password: *dbPass,
		},
		GeolocationFile:                  *geoLocFilePath,
		DeviceDetectorPath:               *deviceDetectorPath,
		RedisHost:                        *redisHost,
		RedisPort:                        *redisPort,
		QueueRedisHost:                   *queueRedisHost,
		QueueRedisPort:                   *queueRedisPort,
		SentryDSN:                        *sentryDSN,
		SDKRequestQueueProjectTokens:     C.GetTokensFromStringListAsString(*sdkRequestQueueProjectTokens), // comma seperated project tokens.
		SegmentRequestQueueProjectTokens: C.GetTokensFromStringListAsString(*segmentRequestQueueProjectTokens),
		RedisHostPersistent:              *redisHostPersistent,
		RedisPortPersistent:              *redisPortPersistent,
		UseDefaultProjectSettingForSDK:   *useDefaultProjectSettingForSDK,
		// List of tokens (public and private) to block SDK requests.
		BlockedSDKRequestProjectTokens: C.GetTokensFromStringListAsString(*blockedSDKRequestProjectTokens),
	}

	err := C.InitSDKService(config)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize.")
		return
	}
	defer C.SafeFlushAllCollectors()

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
