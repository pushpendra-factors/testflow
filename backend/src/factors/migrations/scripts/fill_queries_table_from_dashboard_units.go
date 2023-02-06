package main

import (
	"factors/model/model"
	"factors/model/store"
	"factors/util"
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
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypeMemSQL, "Primary datastore type as memsql or postgres")

	flag.Parse()

	if *env != C.DEVELOPMENT &&
		*env != C.STAGING &&
		*env != C.PRODUCTION {
		err := fmt.Errorf("env [ %s ] not recognised", *env)
		panic(err)
	} else if *projectIDFlag == "" {
		panic(fmt.Errorf("Invalid project id %s", *projectIDFlag))
	}

	taskID := "Script#FillQueriesFromDashboardUnit"
	defer util.NotifyOnPanic(taskID, *env)
	logCtx := log.WithFields(log.Fields{"Prefix": taskID})

	config := &C.Configuration{
		Env: *env,
		MemSQLInfo: C.DBConf{
			Host:     *memSQLHost,
			Port:     *memSQLPort,
			User:     *memSQLUser,
			Name:     *memSQLName,
			Password: *memSQLPass,
		},
		PrimaryDatastore: *primaryDatastore,
	}

	C.InitConf(config)
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
		dashboardUnits, errCode := store.GetStore().GetDashboardUnitsForProjectID(projectID)
		if errCode != http.StatusFound {
			logCtx.WithError(err).Error("Queries table migration failed. Failed to fetch data from dahboard_units table.")
			return
		}
		logCtx.Infof("Migrating %d units for project_id %d", len(dashboardUnits), projectID)
		for _, dashboardUnit := range dashboardUnits {
			if dashboardUnit.QueryId != 0 {
				logCtx.Infof("  Skipping already migrated unit %d", dashboardUnit.ID)
				continue
			}
			query := model.Queries{
				ProjectID: dashboardUnit.ProjectID,
				Title:     dashboardUnit.Title,
				Query:     dashboardUnit.Query,
				Type:      model.QueryTypeDashboardQuery,
				CreatedAt: dashboardUnit.CreatedAt,
				UpdatedAt: dashboardUnit.UpdatedAt,
			}
			err = db.Create(&query).Error
			if err != nil {
				logCtx.WithError(err).Error("Migration failed. Failed to add data to queries table.")
				return
			}
			dashboardUnit.QueryId = query.ID
			err = db.Save(&dashboardUnit).Error
			if err != nil {
				logCtx.WithError(err).Error("Migration failed. Failed to add query_id reference from queries table to dashboardUnits table.")
				return
			}
		}
	}

	return
}
