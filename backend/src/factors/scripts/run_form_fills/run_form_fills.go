package main

import (
	"flag"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"

	C "factors/config"

	T "factors/task"
)

func main() {
	env := flag.String("env", C.DEVELOPMENT, "")

	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	memSQLPort := flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser := flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName := flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass := flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	memSQLCertificate := flag.String("memsql_cert", "", "")
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypeMemSQL, "Primary datastore type as memsql or postgres")

	redisHost := flag.String("redis_host", "localhost", "")
	redisPort := flag.Int("redis_port", 6379, "")
	redisHostPersistent := flag.String("redis_host_ps", "localhost", "")
	redisPortPersistent := flag.Int("redis_port_ps", 6379, "")

	queueRedisHost := flag.String("queue_redis_host", "localhost", "")
	queueRedisPort := flag.Int("queue_redis_port", 6379, "")

	sentryDSN := flag.String("sentry_dsn", "", "Sentry DSN")

	overrideHealthcheckPingID := flag.String("healthcheck_ping_id", "", "Override default healthcheck ping id.")
	overrideAppName := flag.String("app_name", "", "Override default app_name.")

	sdkRequestQueueProjectTokens := flag.String("sdk_request_queue_project_tokens", "*",
		"List of project tokens allowed to use sdk request queue.")
	formFillIdentifyAllowedProjectIDs := flag.String("form_fill_identify_allowed_projects", "*",
		"Form fill identification allowed project ids.")

	flag.Parse()

	if *env != "development" &&
		*env != "staging" &&
		*env != "production" {
		err := fmt.Errorf("env [ %s ] not recognised", *env)
		panic(err)
	}

	const form_fill_ping_id = "0c4216eb-f6be-4aaa-9dae-6876f9d7f3b9"
	defaultAppName := "form_fill"
	defaultHealthcheckPingID := form_fill_ping_id
	healthcheckPingID := C.GetHealthcheckPingID(defaultHealthcheckPingID, *overrideHealthcheckPingID)
	appName := C.GetAppName(defaultAppName, *overrideAppName)
	defer C.PingHealthcheckForPanic(appName, *env, healthcheckPingID)

	config := &C.Configuration{
		AppName: appName,
		Env:     *env,
		MemSQLInfo: C.DBConf{
			Host:        *memSQLHost,
			Port:        *memSQLPort,
			User:        *memSQLUser,
			Name:        *memSQLName,
			Password:    *memSQLPass,
			Certificate: *memSQLCertificate,
			AppName:     appName,
		},
		PrimaryDatastore:                      *primaryDatastore,
		RedisHost:                             *redisHost,
		RedisPort:                             *redisPort,
		RedisHostPersistent:                   *redisHostPersistent,
		RedisPortPersistent:                   *redisPortPersistent,
		QueueRedisHost:                        *queueRedisHost,
		QueueRedisPort:                        *queueRedisPort,
		SentryDSN:                             *sentryDSN,
		FormFillIdentificationAllowedProjects: *formFillIdentifyAllowedProjectIDs,
		SDKRequestQueueProjectTokens:          C.GetTokensFromStringListAsString(*sdkRequestQueueProjectTokens), // comma seperated project tokens.
	}

	C.InitConf(config)
	C.InitSentryLogging(config.SentryDSN, config.AppName)

	// Mandatory requirement.
	err := C.InitDB(*config)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize DB.")
	}

	// Mandatory requirement.
	err = C.InitQueueClient(config.QueueRedisHost, config.QueueRedisPort)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize queue client on form fills.")
	}

	C.InitRedis(config.RedisHost, config.RedisPort)
	C.InitRedisPersistent(config.RedisHostPersistent, config.RedisPortPersistent)

	errCode := T.FormFillProcessing()
	if errCode != http.StatusOK {
		C.PingHealthcheckForFailure(healthcheckPingID, "Form processing failed.")
		return
	}
	C.PingHealthcheckForSuccess(healthcheckPingID, "Form processing successful.")
}
