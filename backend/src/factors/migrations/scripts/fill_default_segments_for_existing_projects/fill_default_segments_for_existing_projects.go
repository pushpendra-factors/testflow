package main

import (
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"flag"
	"fmt"
	"net/http"

	C "factors/config"

	log "github.com/sirupsen/logrus"
)

func main() {
	env := flag.String("env", C.DEVELOPMENT, "")
	projectIDFlag := flag.String("project_id", "", "Comma separated project ids to run for. * to run for all")

	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	memSQLPort := flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser := flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName := flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass := flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	memSQLCertificate := flag.String("memsql_cert", "", "")
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypeMemSQL, "Primary datastore type as memsql or postgres")
	sentryDSN := flag.String("sentry_dsn", "", "Sentry DSN")

	flag.Parse()

	if *env != C.DEVELOPMENT &&
		*env != C.STAGING &&
		*env != C.PRODUCTION {
		err := fmt.Errorf("env [ %s ] not recognised", *env)
		panic(err)
	} else if *projectIDFlag == "" {
		panic(fmt.Errorf("invalid project id %s", *projectIDFlag))
	}

	taskID := "Script#AddDefaultSegmentsForExistingProjects"
	defer U.NotifyOnPanic(taskID, *env)
	logCtx := log.WithFields(log.Fields{"Prefix": taskID})

	config := &C.Configuration{
		Env: *env,
		MemSQLInfo: C.DBConf{
			Host:        *memSQLHost,
			Port:        *memSQLPort,
			User:        *memSQLUser,
			Name:        *memSQLName,
			Password:    *memSQLPass,
			Certificate: *memSQLCertificate,
		},
		PrimaryDatastore: *primaryDatastore,
		SentryDSN:        *sentryDSN,
	}

	C.InitConf(config)
	C.InitSentryLogging(config.SentryDSN, config.AppName)
	// Initialize configs and connections and close with defer.
	err := C.InitDB(*config)
	if err != nil {
		logCtx.WithError(err).Fatal("Failed to run migration. Init failed.")
	}
	db := C.GetServices().Db
	defer db.Close()
	logCtx = logCtx.WithFields(log.Fields{
		"Env":       *env,
		"ProjectID": *projectIDFlag,
	})

	allProjects, projectIDsMap, _ := C.GetProjectsFromListWithAllProjectSupport(*projectIDFlag, "")
	projectIDs := C.ProjectIdsFromProjectIdBoolMap(projectIDsMap)
	if allProjects {
		var errCode int
		projectIDs, errCode = store.GetStore().GetAllProjectIDs()
		if errCode != http.StatusFound {
			return
		}
	}

	for _, projectID := range projectIDs {
		groups, errCode := store.GetStore().GetGroups(projectID)
		if errCode != http.StatusFound {
			logCtx.WithError(err).Error("Default Segments addition failed. Failed to fetch groups.")
			return
		}

		// List of all segments
		allSegmentsMap, statusCode := store.GetStore().GetAllSegments(projectID)
		if statusCode != http.StatusFound {
			log.WithField("project_id", projectID).Error("No segment found for this project")
		}

		for _, group := range groups {
			groupSegmentExists := false
			newSegmentName := U.ALL_ACCOUNT_DEFAULT_PROPERTIES_DISPLAY_NAMES[U.GROUP_TO_DEFAULT_SEGMENT_MAP[group.Name]]

			for _, segment := range allSegmentsMap[U.GROUP_NAME_DOMAINS] {
				if segment.Name == newSegmentName {
					groupSegmentExists = true
					break
				}
			}

			if !groupSegmentExists && model.AccountGroupNames[group.Name] {
				status, err := store.GetStore().CreateDefaultSegment(projectID, group.Name, true)
				if status != http.StatusCreated {
					log.WithError(err).Error("Failed to create default segment.")
				}
			}
		}

		userSegmentExists := false
		for _, segment := range allSegmentsMap[U.GROUP_NAME_DOMAINS] {
			if segment.Name == U.ALL_ACCOUNT_DEFAULT_PROPERTIES_DISPLAY_NAMES[U.VISITED_WEBSITE] {
				userSegmentExists = true
			}
		}

		if !userSegmentExists {
			status, err := store.GetStore().CreateDefaultSegment(projectID, model.PropertyEntityUser, false)
			if status != http.StatusCreated {
				log.WithError(err).Error("Failed to create default segment.")
			}
		}

	}

}
