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

	env := flag.String("env", C.DEVELOPMENT, "")
	port := flag.Int("port", 8089, "")

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
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypePostgres, "Primary datastore type as memsql or postgres")

	memSQLDBMaxOpenConnections := flag.Int("memsql_max_open_connections", 100, "Max no.of open connections allowed on connection pool of memsql")
	memSQLDBMaxIdleConnections := flag.Int("memsql_max_idle_connections", 50, "Max no.of idle connections allowed on connection pool of memsql")

	redisHost := flag.String("redis_host", "localhost", "")
	redisPort := flag.Int("redis_port", 6379, "")

	redisHostPersistent := flag.String("redis_host_ps", "localhost", "")
	redisPortPersistent := flag.Int("redis_port_ps", 6379, "")

	sentryDSN := flag.String("sentry_dsn", "", "Sentry DSN")
	overrideAppName := flag.String("app_name", "", "Override default app_name.")
	disableCRMUniquenessConstraintsCheckByProjectID := flag.String("disable_crm_unique_constraint_check_by_project_id", "", "")

	flag.Parse()

	defaultAppName := "data_server"
	appName := C.GetAppName(defaultAppName, *overrideAppName)

	config := &C.Configuration{
		AppName:            appName,
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

			// Pooling.
			MaxOpenConnections:     *memSQLDBMaxOpenConnections,
			MaxIdleConnections:     *memSQLDBMaxIdleConnections,
			UseExactConnFromConfig: true,
		},
		PrimaryDatastore:    *primaryDatastore,
		RedisHost:           *redisHost,
		RedisPort:           *redisPort,
		RedisHostPersistent: *redisHostPersistent,
		RedisPortPersistent: *redisPortPersistent,
		SentryDSN:           *sentryDSN,
		DisableCRMUniquenessConstraintsCheckByProjectID: *disableCRMUniquenessConstraintsCheckByProjectID,
	}
	C.InitConf(config)
	C.InitRedisPersistent(config.RedisHostPersistent, config.RedisPortPersistent)
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
