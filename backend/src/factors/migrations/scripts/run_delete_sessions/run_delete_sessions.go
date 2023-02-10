package main

import (
	"flag"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"

	C "factors/config"
	"factors/model/model"
	"factors/model/store"
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

	if !*wetRun {
		log.Info("Running in dry run")
	} else {
		log.Info("Running in wet run")
	}

	log.WithFields(log.Fields{"start_timestamp": *startTimestamp, "end_timestamp": *endTimestamp}).Info("Running for following time range.")

	projectIDList := store.GetStore().GetProjectsToRunForIncludeExcludeString(*projectIDs, "")

	for i := range projectIDList {
		log.WithFields(log.Fields{"project_id": projectIDList[i]}).Info("Running for project.")

		status := RunDissociateSession(projectIDList[i], *startTimestamp, *endTimestamp, *wetRun)
		if status != http.StatusOK {
			log.WithFields(log.Fields{"project_id": projectIDList[i], "status": status}).Error("Failed to RunDissociateSession. Exiting.")
			continue
		}
	}
}

func RunDissociateSession(projectID int64, startTimestamp, endTimestamp int64, wetRun bool) int {
	logCtx := log.WithFields(log.Fields{"project_id": projectID, "start_timestamp": startTimestamp, "end_timestamp": endTimestamp, "wet_run": wetRun})
	if projectID == 0 || startTimestamp == 0 || endTimestamp == 0 {
		logCtx.Error("Invalid parameters.")
		return http.StatusBadRequest
	}

	sessionEventName, status := store.GetStore().GetSessionEventName(projectID)
	if status != http.StatusFound {
		logCtx.Error("Failed to get event name.")
		return http.StatusInternalServerError
	}

	sessionEvents, status := store.GetStore().GetEventsByEventNameIDANDTimeRange(projectID, sessionEventName.ID, startTimestamp, endTimestamp)
	if status != http.StatusFound {
		logCtx.Error("Failed to get session events for given range.")
		return status
	}

	sessionAssociatedEvents, totalEvents := getSessionAssociatedEvents(projectID, sessionEvents)

	logCtx.WithFields(log.Fields{"total_session_events": len(sessionEvents), "total_events_associated_with_session": totalEvents}).Info("Total session events.")

	for sessionID, events := range sessionAssociatedEvents {
		logCtx := logCtx.WithFields(log.Fields{"session_id": sessionID, "total_events": len(events)})
		if !wetRun {
			for i := range events {
				logCtx.WithFields(log.Fields{"event_id": events[i].ID}).Info("Event session id will be removed.")
			}
			continue
		}

		status := store.GetStore().DissociateEventsFromSession(projectID, events, sessionID)
		if status != http.StatusAccepted {
			logCtx.Error("Failed to dissociate session from events.")
			return http.StatusInternalServerError
		}

		// delete session event
		status = store.GetStore().DeleteEventByIDs(projectID, sessionEventName.ID, []string{sessionID})
		if status != http.StatusAccepted {
			logCtx.Error("Failed to delete session event.")
			continue
		}

		logCtx.Info("Removed session association from events.")
	}

	return http.StatusOK
}

func getSessionAssociatedEvents(projectID int64, sessionEvents []model.Event) (map[string][]model.Event, int) {
	sessionAssociatedEvents := make(map[string][]model.Event)
	totalEvents := 0
	for i := range sessionEvents {
		events, status := store.GetStore().GetEventsBySessionEvent(projectID, sessionEvents[i].ID, sessionEvents[i].UserId)
		if status != http.StatusFound {
			log.WithFields(log.Fields{"project_id": projectID, "session_event_id": sessionEvents[i].ID, "session_event_user_id": sessionEvents[i].UserId}).
				Error("Failed to get events from given session id. Skip processing current session id.")
			continue
		}
		sessionAssociatedEvents[sessionEvents[i].ID] = events
		totalEvents += len(events)
	}

	return sessionAssociatedEvents, totalEvents
}
