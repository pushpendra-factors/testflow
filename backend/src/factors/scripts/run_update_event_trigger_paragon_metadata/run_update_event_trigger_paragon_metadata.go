package main

import (
	C "factors/config"
	"factors/model/store"
	U "factors/util"
	"flag"
	"fmt"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"
)

const (
	ParagonMetadataColumnName string = "paragon_metadata"
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

	sentryDSN := flag.String("sentry_dsn", "", "Sentry DSN")

	overrideAppName := flag.String("app_name", "", "Override default app_name.")

	projectID := flag.Int64("project_id", 0, "project_id for which the alerts are to be updated")
	alertId := flag.String("alert_id", "", "All alert_id for which the table needs to be updated")
	eventName := flag.String("event_name", "", "Event Name to be sent in the paragon payload")

	flag.Parse()

	if *env != "development" &&
		*env != "staging" &&
		*env != "production" {
		err := fmt.Errorf("env [ %s ] not recognised", *env)
		panic(err)
	}

	defaultAppName := "update_event_trigger_paragon_metadata"
	appName := C.GetAppName(defaultAppName, *overrideAppName)

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
		PrimaryDatastore: *primaryDatastore,
		SentryDSN:        *sentryDSN,
	}

	C.InitConf(config)
	C.InitSentryLogging(config.SentryDSN, config.AppName)

	err := C.InitDB(*config)
	if err != nil {
		log.Error("Failed to initialize DB.")
		os.Exit(1)
	}

	db := C.GetServices().Db
	defer db.Close()

	alerts := C.GetTokensFromStringListAsString(*alertId)

	for _, alert := range alerts {
		metadata := map[string]interface{}{
			"EventName": *eventName,
		}

		//encode the metadata map to jsonb
		metadataJson, err := U.EncodeToPostgresJsonb(&metadata)
		if err != nil {
			log.WithFields(log.Fields{
				"project_id": *projectID,
				"alert_id":   alert,
			}).WithError(err).Error("failed to encode metadata into json")
		}

		//update column and values
		fields := map[string]interface{}{
			ParagonMetadataColumnName: metadataJson,
		}

		//store update function
		errCode, err := store.GetStore().UpdateEventTriggerAlertField(*projectID, alert, fields)
		if errCode != http.StatusAccepted || err != nil {
			log.WithFields(log.Fields{
				"project_id": *projectID,
				"alert_id":   alert,
			}).WithError(err).Error("failed to update table")
		}
	}
}
