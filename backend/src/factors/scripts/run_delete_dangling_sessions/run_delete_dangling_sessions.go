package main

import (
	"bufio"
	"encoding/json"
	C "factors/config"
	"factors/filestore"
	"factors/model/model"
	"factors/model/store"
	serviceDisk "factors/services/disk"
	serviceGCS "factors/services/gcstorage"
	"factors/util"
	"flag"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
)

func getUnusedSessionIDsFromArchiveFile(store filestore.FileManager, path, filename string) ([]string, int) {
	var unusedSessionIDs []string

	eventsArchiveFile, err := store.Get(path, filename)
	if err != nil {
		log.WithError(err).WithField("path", path).WithField("filename", filename).
			Error("Failed to get the file from cloud storage.")
		return unusedSessionIDs, http.StatusInternalServerError
	}
	defer eventsArchiveFile.Close()

	allSessionIDs := make([]string, 0, 0)

	usedSessionIDMap := make(map[string]bool, 0)
	scanner := bufio.NewScanner(eventsArchiveFile)
	for scanner.Scan() {
		eventJSON := scanner.Text()

		var event model.ArchiveEventTableFormat
		if err := json.Unmarshal([]byte(eventJSON), &event); err != nil {
			log.Error("Unable to decode the event on file.")
			return unusedSessionIDs, http.StatusInternalServerError
		}

		// session_ids associated to event.
		if event.SessionID != "" && event.EventName != util.EVENT_NAME_SESSION {
			usedSessionIDMap[event.SessionID] = true
		}

		// all session events.
		if event.EventName == util.EVENT_NAME_SESSION {
			allSessionIDs = append(allSessionIDs, event.EventID)
		}
	}
	err = scanner.Err()
	if err != nil {
		log.WithError(err).Error("Failure while scanning the events file.")
		return unusedSessionIDs, http.StatusInternalServerError
	}

	log.WithField("used_sessions_count", len(usedSessionIDMap)).
		WithField("all_sessions_count", len(allSessionIDs)).
		Info("Used sessions count.")

	unusedSessionIDMap := make(map[string]bool, 0)
	for i := range allSessionIDs {
		if _, exists := usedSessionIDMap[allSessionIDs[i]]; !exists {
			unusedSessionIDMap[allSessionIDs[i]] = true
		}
	}

	unusedSessionIDs = make([]string, 0, 0)
	for sessionID := range unusedSessionIDMap {
		if sessionID == "" {
			continue
		}

		unusedSessionIDs = append(unusedSessionIDs, sessionID)
	}

	return unusedSessionIDs, http.StatusFound
}

func main() {
	env := flag.String("env", "development", "")

	dbHost := flag.String("db_host", "localhost", "")
	dbPort := flag.Int("db_port", 5432, "")
	dbUser := flag.String("db_user", "autometa", "")
	dbName := flag.String("db_name", "autometa", "")
	dbPass := flag.String("db_pass", "@ut0me7a", "")

	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	memSQLPort := flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser := flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName := flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass := flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	memSQLCertificate := flag.String("memsql_cert", "", "")
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypeMemSQL, "Primary datastore type as memsql or postgres")

	startTimestamp := flag.Int64("start_timestamp", 0, "")
	endTimestamp := flag.Int64("end_timestamp", 0, "")

	archiveEventsFilepath := flag.String("archive_filepath", "", "")
	archiveEventsFilename := flag.String("archive_filename", "", "")

	projectID := flag.Uint64("project_id", 0, "")
	wet := flag.Bool("wet", false, "Mutate data only if it is set to true.")

	sentryDSN := flag.String("sentry_dsn", "", "Sentry DSN")

	flag.Parse()

	if *env != "development" &&
		*env != "staging" &&
		*env != "production" {
		err := fmt.Errorf("env [ %s ] not recognised", *env)
		panic(err)
	}

	appName := "delete_unused_sessions"
	config := &C.Configuration{
		Env:     *env,
		AppName: appName,
		DBInfo: C.DBConf{
			Host:     *dbHost,
			Port:     *dbPort,
			User:     *dbUser,
			Name:     *dbName,
			Password: *dbPass,
		},
		MemSQLInfo: C.DBConf{
			Host:        *memSQLHost,
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
	err := C.InitDB(*config)
	if err != nil {
		log.WithError(err).Panic("Failed to initialize db in add session.")
		return
	}
	C.InitSentryLogging(config.SentryDSN, config.AppName)
	C.GetServices().Db.LogMode(true)

	logCtx := log.WithField("project_id", *projectID).
		WithField("start_timestamp", *startTimestamp).
		WithField("end_timestamp", *endTimestamp)

	// Init storage with archival bucket.
	var cloudStorage filestore.FileManager
	if C.IsDevelopment() {
		cloudStorage = serviceDisk.New("factors-production-archival")
	} else {
		cloudStorage, err = serviceGCS.New("factors-production-archival")
		if err != nil {
			log.WithError(err).Error("Failed to init new GCS client.")
			return
		}
	}

	if *projectID == 0 || *startTimestamp == 0 || *endTimestamp == 0 {
		logCtx.Fatal("Invalid params.")
	}

	logCtx.Info("Started delete dangling sessions job.")

	startTime := util.TimeNowUnix()

	var unusedSessionIDs []string
	var errCode int
	if *archiveEventsFilepath != "" && *archiveEventsFilename != "" {
		logCtx.Info("Reading events from file.")
		unusedSessionIDs, errCode = getUnusedSessionIDsFromArchiveFile(cloudStorage, *archiveEventsFilepath, *archiveEventsFilename)
	} else {
		logCtx.Info("Reading events from DB.")
		unusedSessionIDs, errCode = store.GetStore().GetUnusedSessionIDsForJob(*projectID, *startTimestamp, *endTimestamp)
	}

	if errCode == http.StatusInternalServerError {
		logCtx.Error("Failed to get unused session ids.")
		return
	}

	logCtx = logCtx.WithField("no_of_unused_sessions", len(unusedSessionIDs)).
		WithField("download_timetaken", util.TimeNowUnix()-startTime)

	if !*wet {
		if len(unusedSessionIDs) > 100 {
			logCtx.WithField("sample_ids", unusedSessionIDs[0:100]).Info("Dry run.")
		}
		return
	}

	sessionEventName, errCode := store.GetStore().GetSessionEventName(*projectID)
	if errCode != http.StatusFound {
		logCtx.Error("Failed to get session event_name.")
		return
	}

	logCtx.Info("Started deleting dangling sessions.")
	errCode = store.GetStore().DeleteEventsByIDsInBatchForJob(*projectID, sessionEventName.ID, unusedSessionIDs, 1000)
	if errCode == http.StatusInternalServerError {
		log.Error("Failed to delete sessions in batch")
		return
	}

	logCtx.Info("Successfully delete session.")
}
