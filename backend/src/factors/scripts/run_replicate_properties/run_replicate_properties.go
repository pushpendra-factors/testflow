package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	C "factors/config"
	U "factors/util"

	log "github.com/sirupsen/logrus"
)

func main() {
	env := flag.String("env", "development", "")

	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	memSQLPort := flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser := flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName := flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass := flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	memSQLCertificate := flag.String("memsql_cert", "", "")
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypePostgres,
		"Primary datastore type as memsql or postgres")

	healthcheckPingID := flag.String("healthcheck_ping_id", "", "Healthcheck ping id.")
	sentryDSN := flag.String("sentry_dsn", "", "Sentry DSN")
	appName := flag.String("app_name", "replicate_properties_json", "Default app_name.")
	enableUserPropertiesReplication := flag.Bool("enable_user_properties_replication", false,
		"Enables user_properties replication.")
	enableEventPropertiesReplication := flag.Bool("enable_event_properties_replication", false,
		"Enables event_properties replication.")

	flag.Parse()

	if *env != "development" &&
		*env != "staging" &&
		*env != "production" {
		err := fmt.Errorf("env [ %s ] not recognised", *env)
		panic(err)
	}

	defer C.PingHealthcheckForPanic(*appName, *env, *healthcheckPingID)

	config := &C.Configuration{
		AppName:   *appName,
		Env:       *env,
		SentryDSN: *sentryDSN,
		MemSQLInfo: C.DBConf{
			Host:        *memSQLHost,
			Port:        *memSQLPort,
			User:        *memSQLUser,
			Name:        *memSQLName,
			Password:    *memSQLPass,
			Certificate: *memSQLCertificate,
			AppName:     *appName,
		},
		PrimaryDatastore: *primaryDatastore,
	}

	C.InitConf(config)
	err := C.InitDB(*config)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{"env": *env,
			"host": *memSQLHost, "port": *memSQLPort}).Panic("Failed to initialize DB.")
		os.Exit(0)
	}

	if *enableUserPropertiesReplication {
		replicateUserProperties()
	}

	if *enableEventPropertiesReplication {
		replicateEventProperties()
	}
}

func getUserPropertiesJsonMaxUpdatedAt() (string, int) {
	db := C.GetServices().Db
	query := db.Raw("SELECT max(updated_at) FROM user_properties_json")
	rows, err := query.Rows()
	if err != nil {
		log.WithError(err).Error("Failed to get max updated_at of user_properties_json.")
		return "", http.StatusInternalServerError
	}
	defer rows.Close()

	var maxUpdatedAt string
	for rows.Next() {
		err = rows.Scan(&maxUpdatedAt)
		if err != nil {
			log.WithError(err).Error("Failed to read max_updated_at from result.")
			return "", http.StatusInternalServerError
		}
	}

	return maxUpdatedAt, http.StatusOK
}

func deleteUserPropertiesJsonByUsersMaxUpdatedAt(maxUpdatedAt string) int {
	logCtx := log.WithField("max_updated_at", maxUpdatedAt)
	if maxUpdatedAt == "" {
		logCtx.Error("Empty max_updated_at on deleteUserPropertiesJsonByUsersMaxUpdatedAt.")
		return http.StatusBadRequest
	}

	db := C.GetServices().Db
	query := db.Raw("DELETE FROM user_properties_json WHERE id IN (SELECT id FROM users WHERE updated_at > ?)", maxUpdatedAt)
	rows, err := query.Rows()
	if err != nil {
		logCtx.WithError(err).Error("Failed to delete by max_updated_at of users.")
		return http.StatusInternalServerError
	}
	defer rows.Close()

	return http.StatusOK
}

func copyPropertiesFromUsers(maxUpdatedAt string) int {
	logCtx := log.WithField("max_updated_at", maxUpdatedAt)

	if maxUpdatedAt == "" {
		logCtx.Error("Empty max_updated_at on copyPropertiesFromUsers.")
		return http.StatusBadRequest
	}

	statement := "INSERT INTO user_properties_json (id, project_id, properties_json, created_at, updated_at)" + " " +
		"SELECT id, project_id, properties, created_at, updated_at FROM users WHERE updated_at > ?"

	db := C.GetServices().Db
	query := db.Raw(statement, maxUpdatedAt)
	rows, err := query.Rows()
	if err != nil {
		logCtx.WithError(err).Error("Failed to delete by max_updated_at of users.")
		return http.StatusInternalServerError
	}
	defer rows.Close()

	return http.StatusOK
}

func replicateUserProperties() {
	maxUpdatedAt, status := getUserPropertiesJsonMaxUpdatedAt()
	if status != http.StatusOK {
		return
	}
	logCtx := log.WithField("max_updated_at", maxUpdatedAt)
	logCtx.Info("Fetched max updated_at of user_properties_json")

	status = deleteUserPropertiesJsonByUsersMaxUpdatedAt(maxUpdatedAt)
	if status != http.StatusOK {
		return
	}
	logCtx.Info("Deleted user_properties_json by max_updated_at")

	status = copyPropertiesFromUsers(maxUpdatedAt)
	if status != http.StatusOK {
		return
	}
	logCtx.Info("Replicated properties from users by max_updated_at")
}

func getEventPropertiesJsonMaxUpdatedAt() (string, int) {
	db := C.GetServices().Db
	query := db.Raw("SELECT max(updated_at) FROM event_properties_json")
	rows, err := query.Rows()
	if err != nil {
		log.WithError(err).Error("Failed to get max updated_at of event_properties_json.")
		return "", http.StatusInternalServerError
	}
	defer rows.Close()

	var maxUpdatedAt string
	for rows.Next() {
		err = rows.Scan(&maxUpdatedAt)
		if err != nil {
			log.WithError(err).Error("Failed to read max_updated_at from event_properties_json.")
			return "", http.StatusInternalServerError
		}
	}

	return maxUpdatedAt, http.StatusOK
}

func getEventPropertiesStartTimestampFromMaxUpdatedAt(maxUpdatedAt string) (int64, error) {
	parsedMaxUpdatedAt, err := time.Parse(time.RFC3339, maxUpdatedAt)
	if err != nil {
		parsedMaxUpdatedAt, err = time.Parse(U.DATETIME_FORMAT_DB, maxUpdatedAt)
		if err != nil {
			return 0, err
		}
	}

	// Uses timestamp to improve performance of the query.
	// Filter with timestamp considering that event updated after latest
	// updated_at on event_properties_json, would not got created before an hour ago.
	startTimestamp := parsedMaxUpdatedAt.Unix() - int64((1 * time.Hour).Seconds())

	return startTimestamp, nil
}

func deletEventPropertiesJsonByMaxUpdatedAt(maxUpdatedAt string) int {
	logCtx := log.WithField("max_updated_at", maxUpdatedAt)

	if maxUpdatedAt == "" {
		logCtx.Error("Empty max_updated_at on deletEventPropertiesJsonByMaxUpdatedAt.")
		return http.StatusBadRequest
	}

	startTimestamp, err := getEventPropertiesStartTimestampFromMaxUpdatedAt(maxUpdatedAt)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get start timestamp on deletEventPropertiesJsonByMaxUpdatedAt.")
		return http.StatusInternalServerError
	}

	db := C.GetServices().Db
	query := db.Raw("DELETE FROM event_properties_json WHERE id IN (SELECT id FROM events WHERE timestamp > ? AND updated_at > ?)",
		startTimestamp, maxUpdatedAt)
	rows, err := query.Rows()
	if err != nil {
		logCtx.WithError(err).Error("Failed to delete by max_updated_at of event_properties_json.")
		return http.StatusInternalServerError
	}
	defer rows.Close()

	return http.StatusOK
}

func copyPropertiesFromEvents(maxUpdatedAt string) int {
	logCtx := log.WithField("max_updated_at", maxUpdatedAt)

	if maxUpdatedAt == "" {
		logCtx.Error("Empty max_updated_at on copyPropertiesFromEvents.")
		return http.StatusBadRequest
	}

	startTimestamp, err := getEventPropertiesStartTimestampFromMaxUpdatedAt(maxUpdatedAt)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get start timestamp on copyPropertiesFromEvents.")
		return http.StatusInternalServerError
	}

	statement := "INSERT INTO event_properties_json (id, project_id, user_id, properties_json, user_properties_json, created_at, updated_at)" + " " +
		"SELECT id, project_id, user_id, properties, user_properties, created_at, updated_at FROM events where timestamp > ? and updated_at > ?"

	db := C.GetServices().Db
	query := db.Raw(statement, startTimestamp, maxUpdatedAt)
	rows, err := query.Rows()
	if err != nil {
		logCtx.WithError(err).Error("Failed to delete by max_updated_at of events.")
		return http.StatusInternalServerError
	}
	defer rows.Close()

	return http.StatusOK
}

func replicateEventProperties() {
	maxUpdatedAt, status := getEventPropertiesJsonMaxUpdatedAt()
	if status != http.StatusOK {
		return
	}
	logCtx := log.WithField("max_updated_at", maxUpdatedAt)
	logCtx.Info("Fetched max updated_at of event_properties_json")

	status = deletEventPropertiesJsonByMaxUpdatedAt(maxUpdatedAt)
	if status != http.StatusOK {
		return
	}
	logCtx.Info("Deleted event_properties_json by max_updated_at")

	status = copyPropertiesFromEvents(maxUpdatedAt)
	if status != http.StatusOK {
		return
	}
	logCtx.Info("Replicated properties from events by max_updated_at")
}
