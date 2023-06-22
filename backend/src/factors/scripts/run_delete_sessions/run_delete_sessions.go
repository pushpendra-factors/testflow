package main

import (
	"flag"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"

	C "factors/config"
	"factors/model/store"
)

func main() {
	env := flag.String("env", C.DEVELOPMENT, "")

	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	isPSCHost := flag.Int("memsql_is_psc_host", C.MemSQLDefaultDBParams.IsPSCHost, "")
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

	sentryDSN := flag.String("sentry_dsn", "", "Sentry DSN")

	startTimestamp := flag.Int64("start_timestamp", 0, "Start timestamp")
	endTimestamp := flag.Int64("end_timestamp", 0, "End timestamp")
	wetRun := flag.Bool("wet", false, "Wet run")

	projectIDs := flag.String("project_ids", "", "Project ids")
	flag.Parse()

	appName := "run_delete_sessions"

	if *startTimestamp == 0 || *endTimestamp == 0 {
		log.Panic("Missing start timestamp or end timestamp.")
	}

	config := &C.Configuration{
		AppName: appName,
		Env:     *env,
		MemSQLInfo: C.DBConf{
			Host:        *memSQLHost,
			IsPSCHost:   *isPSCHost,
			Port:        *memSQLPort,
			User:        *memSQLUser,
			Name:        *memSQLName,
			Password:    *memSQLPass,
			Certificate: *memSQLCertificate,
			AppName:     appName,
		},
		PrimaryDatastore:    *primaryDatastore,
		RedisHost:           *redisHost,
		RedisPort:           *redisPort,
		RedisHostPersistent: *redisHostPersistent,
		RedisPortPersistent: *redisPortPersistent,
		SentryDSN:           *sentryDSN,
	}

	C.InitConf(config)
	C.InitSentryLogging(config.SentryDSN, config.AppName)

	err := C.InitDB(*config)
	if err != nil {
		log.Error("Failed to initialize DB.")
		os.Exit(1)
	}
	// Log queries.
	C.GetServices().Db.LogMode(true)

	if !*wetRun {
		log.Info("Running in dry run")
	} else {
		log.Info("Running in wet run")
	}

	log.WithFields(log.Fields{"start_timestamp": *startTimestamp, "end_timestamp": *endTimestamp}).Info("Running for following time range.")

	projectIDList := store.GetStore().GetProjectsToRunForIncludeExcludeString(*projectIDs, "")

	for i := range projectIDList {
		log.WithFields(log.Fields{"project_id": projectIDList[i]}).Info("Running for project.")
		RunDeleteSession(projectIDList[i], *startTimestamp, *endTimestamp, *wetRun)
	}
}

func RunDeleteSession(projectID int64, startTimestamp, endTimestamp int64, wetRun bool) {
	logCtx := log.WithFields(log.Fields{"project_id": projectID, "start_timestamp": startTimestamp, "end_timestamp": endTimestamp, "wet_run": wetRun})
	if projectID == 0 || startTimestamp == 0 || endTimestamp == 0 {
		logCtx.Error("Invalid parameters.")
		return
	}

	sessionsDeleted, associationsRemoved, status := store.GetStore().DeleteSessionsAndAssociationForTimerange(projectID, startTimestamp, endTimestamp)
	logCtx = logCtx.WithFields(log.Fields{"sessions_deleted": sessionsDeleted, "associations_removed": associationsRemoved, "status": status})

	if status != http.StatusAccepted {
		logCtx.Error("Failure on deleting sessions.")
		return
	}

	logCtx.Error("Deleted sessions.")
}
