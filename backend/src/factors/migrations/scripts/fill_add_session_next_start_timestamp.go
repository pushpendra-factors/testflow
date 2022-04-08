package main

import (
	"flag"
	"fmt"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"

	C "factors/config"
	"factors/model/model"
	"factors/model/store"
	"factors/model/store/postgres"
	U "factors/util"
)

func getNextSessionStartTimestamp(projectID uint64, maxLookbackTimestamp int64) (int64, int) {
	logCtx := log.WithField("project_id", projectID)

	eventName, errCode := store.GetStore().GetEventName(U.EVENT_NAME_SESSION, projectID)
	if errCode == http.StatusInternalServerError || errCode == http.StatusBadRequest {
		logCtx.Error("Failed to get session event name.")
		return 0, http.StatusInternalServerError
	}

	var project *model.Project
	if errCode == http.StatusNotFound {
		project, errCode = store.GetStore().GetProject(projectID)
		if errCode != http.StatusFound {
			logCtx.Error("Failed to get project by id.")
			return 0, http.StatusInternalServerError
		}

		return project.CreatedAt.Unix(), http.StatusFound
	}

	// Using previous initial query to build next session_info to initialize the project level metadata.
	oldUsersStartTimestamp, errCode := postgres.GetStore().GetNextSessionEventInfoFromDB(
		projectID, true, eventName.ID, maxLookbackTimestamp)
	if errCode == http.StatusInternalServerError {
		logCtx.Error("Failed to get next session start for users with session already.")
		return 0, http.StatusInternalServerError
	}
	startTimestamp := oldUsersStartTimestamp

	newUsersStartTimestamp, errCode := postgres.GetStore().GetNextSessionEventInfoFromDB(
		projectID, false, eventName.ID, maxLookbackTimestamp)
	if errCode == http.StatusInternalServerError {
		logCtx.Error("Failed to get next session start new users.")
		return 0, http.StatusInternalServerError
	}
	if newUsersStartTimestamp < oldUsersStartTimestamp {
		startTimestamp = newUsersStartTimestamp
	}

	if startTimestamp > 0 {
		return startTimestamp, http.StatusFound
	}

	// Use last session event timestamp, if no event on last 6 hours.
	lastEventTimestamp, errCode := postgres.GetStore().GetLastSessionEventTimestamp(projectID, eventName.ID)
	if errCode == http.StatusInternalServerError {
		logCtx.Error("Failed to last session event timestamp")
		return 0, http.StatusInternalServerError
	}

	if lastEventTimestamp > 0 {
		return lastEventTimestamp, http.StatusFound
	}

	// Use the project creation timestamp, if no events found.
	if project == nil {
		project, errCode = store.GetStore().GetProject(projectID)
		if errCode != http.StatusFound {
			logCtx.Error("Failed to project by id")
			return 0, http.StatusInternalServerError
		}
	}

	return project.CreatedAt.Unix(), http.StatusFound
}

func getAndUpdateNextSessionStartTimestamp(projectID uint64, maxLookbackTimestamp int64) int {
	if projectID == 0 {
		log.WithField("project_id", projectID).Error("Invalid project_id")
		return http.StatusInternalServerError
	}

	project, errCode := store.GetStore().GetProject(projectID)
	if errCode != http.StatusFound {
		return http.StatusInternalServerError
	}

	// Skip the projects which has jobs_metadata
	// as this is the first field being added.
	if project.JobsMetadata != nil {
		return http.StatusOK
	}

	startTimestamp, errCode := getNextSessionStartTimestamp(projectID, maxLookbackTimestamp)
	if errCode == http.StatusInternalServerError {
		return http.StatusInternalServerError
	}

	errCode = postgres.GetStore().FillNextSessionStartTimestampForProject(
		projectID, startTimestamp)
	if errCode != http.StatusAccepted {
		return http.StatusInternalServerError
	}

	return http.StatusOK
}

func main() {
	env := flag.String("env", C.DEVELOPMENT, "")
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
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypeMemSQL, "Primary datastore type as memsql or postgres")
	projectID := flag.Uint64("project_id", 0, "")

	maxLookbackHours := flag.Int64("max_lookback_hours", 24, "")

	flag.Parse()

	if *env != "development" && *env != "staging" && *env != "production" {
		err := fmt.Errorf("env [ %s ] not recognised", *env)
		panic(err)
	}

	appName := "fill_session_next_start_timestamp"
	config := &C.Configuration{
		AppName: appName,
		Env:     *env,
		DBInfo: C.DBConf{
			Host:     *dbHost,
			Port:     *dbPort,
			User:     *dbUser,
			Name:     *dbName,
			Password: *dbPass,
			AppName:  appName,
		},
		MemSQLInfo: C.DBConf{
			Host:     *memSQLHost,
			Port:     *memSQLPort,
			User:     *memSQLUser,
			Name:     *memSQLName,
			Password: *memSQLPass,
			AppName:  appName,
		},
		PrimaryDatastore: *primaryDatastore,
	}

	C.InitConf(config)
	err := C.InitDB(*config)
	if err != nil {
		log.WithError(err).Panic("Failed to initialize db in add session.")
	}

	var maxLookbackTimestamp int64
	if *maxLookbackHours > 0 {
		maxLookbackTimestamp = U.UnixTimeBeforeDuration(time.Hour * time.Duration(*maxLookbackHours))
	}

	if *projectID > 0 {
		logCtx := log.WithField("project_id", *projectID)
		errCode := getAndUpdateNextSessionStartTimestamp(*projectID, maxLookbackTimestamp)
		if errCode == http.StatusInternalServerError {
			logCtx.Error("Failed to update next session start timestamp.")
			return
		}

		logCtx.Info("Successfully updated next session start timestamp for project.")
		return
	} else {
		projectIDs, errCode := store.GetStore().GetAllProjectIDs()
		if errCode != http.StatusFound {
			log.Fatal("No projects found.")
		}

		for i := range projectIDs {
			logCtx := log.WithField("total_projects", len(projectIDs)).
				WithField("in-progress", i+1).
				WithField("project_id", projectIDs[i])

			errCode := getAndUpdateNextSessionStartTimestamp(projectIDs[i], maxLookbackTimestamp)
			if errCode == http.StatusInternalServerError {
				logCtx.Error("Failed to update next session start timestamp.")
			} else {
				logCtx.Info("Updated next session start timestamp.")
			}
		}
	}

	log.Info("Successfully updated next session start timestamp.")
}
