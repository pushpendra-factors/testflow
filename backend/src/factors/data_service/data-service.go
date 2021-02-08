package main

import (
	C "factors/config"
	"flag"
	"strconv"

	H "factors/handler"
	mid "factors/middleware"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func main() {
	gcpProjectID := flag.String("gcp_project_id", "", "Project ID on Google Cloud")
	gcpProjectLocation := flag.String("gcp_project_location", "", "Location of google cloud project cluster")

	env := flag.String("env", "development", "")
	port := flag.Int("port", 8089, "")

	dbHost := flag.String("db_host", "localhost", "")
	dbPort := flag.Int("db_port", 5432, "")
	dbUser := flag.String("db_user", "autometa", "")
	dbName := flag.String("db_name", "autometa", "")
	dbPass := flag.String("db_pass", "@ut0me7a", "")

	redisHost := flag.String("redis_host", "localhost", "")
	redisPort := flag.Int("redis_port", 6379, "")

	sentryDSN := flag.String("sentry_dsn", "", "Sentry DSN")

	flag.Parse()

	config := &C.Configuration{
		AppName:            "data_server",
		GCPProjectID:       *gcpProjectID,
		GCPProjectLocation: *gcpProjectLocation,
		Env:                *env,
		Port:               *port,
		DBInfo: C.DBConf{
			Host:     *dbHost,
			Port:     *dbPort,
			User:     *dbUser,
			Name:     *dbName,
			Password: *dbPass,
			AppName:  "data_server",
		},
		RedisHost: *redisHost,
		RedisPort: *redisPort,
		SentryDSN: *sentryDSN,
	}

	err := C.InitDataService(config)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize.")
		return
	}
	defer C.SafeFlushAllCollectors()

	if !C.IsDevelopment() {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(mid.RequestIdGenerator())
	r.Use(mid.Logger())
	r.Use(mid.Recovery())

	// Initialize routes.
	H.InitDataServiceRoutes(r)
	r.Run(":" + strconv.Itoa(C.GetConfig().Port))
}
