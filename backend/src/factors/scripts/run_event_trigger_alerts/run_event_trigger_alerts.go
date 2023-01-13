package main

import (
	C "factors/config"
	T "factors/task"
	taskWrapper "factors/task/task_wrapper"
	"flag"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
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

	redisHostPersistent := flag.String("redis_host_ps", "localhost", "")
	redisPortPersistent := flag.Int("redis_port_ps", 6379, "")

	sentryDSN := flag.String("sentry_dsn", "", "Sentry DSN")

	overrideAppName := flag.String("app_name", "", "Override default app_name.")

	eventTriggerEnabled := flag.Bool("event_trigger_enabled", false, "")
	eventTriggerEnabledProjectIDs := flag.String("event_trigger_enabled_project_ids", "", "")

	flag.Parse()

	if *env != "development" &&
		*env != "staging" &&
		*env != "production" {
		err := fmt.Errorf("env [ %s ] not recognised", *env)
		panic(err)
	}

	defaultAppName := "event_trigger_alerts"
	appName := C.GetAppName(defaultAppName, *overrideAppName)

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
		PrimaryDatastore:              *primaryDatastore,
		RedisHostPersistent:           *redisHostPersistent,
		RedisPortPersistent:           *redisPortPersistent,
		SentryDSN:                     *sentryDSN,
		EventTriggerEnabled:           *eventTriggerEnabled,
		EventTriggerEnabledProjectIDs: *eventTriggerEnabledProjectIDs,
	}

	C.InitConf(config)
	C.InitSentryLogging(config.SentryDSN, config.AppName)
	C.InitRedisPersistent(config.RedisHostPersistent, config.RedisPortPersistent)

	err := C.InitDB(*config)
	if err != nil {
		log.Error("Failed to initialize DB.")
		os.Exit(1)
	}

	db := C.GetServices().Db
	defer db.Close()

	conf := make(map[string]interface{})
	status := taskWrapper.TaskFunc("EventTriggerAlertSender", 1, T.EventTriggerAlertsSender, conf)
	log.Info(status)
	if status["err"] != nil {
		log.Error("Event Trigger Alert job failing.")
	}
}
