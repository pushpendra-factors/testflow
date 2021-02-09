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
	env := flag.String("env", "development", "")
	projectIDFlag := flag.String("project_id", "", "Comma separated project ids to run for. * to run for all")

	dbHost := flag.String("db_host", "localhost", "")
	dbPort := flag.Int("db_port", 5432, "")
	dbUser := flag.String("db_user", "autometa", "")
	dbName := flag.String("db_name", "autometa", "")
	dbPass := flag.String("db_pass", "@ut0me7a", "")

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
		DBInfo: C.DBConf{
			Host:     *dbHost,
			Port:     *dbPort,
			User:     *dbUser,
			Name:     *dbName,
			Password: *dbPass,
		},
	}

	C.InitConf(config.Env)
	// Initialize configs and connections and close with defer.
	err := C.InitDB(config.DBInfo)
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
		dashboardUnits, err := getAllDashboardUnitsRowsForProject(projectID)
		if err != nil {
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

func getAllDashboardUnitsRowsForProject(projectID uint64) ([]model.DashboardUnit, error) {
	db := C.GetServices().Db
	dashboardUnits := make([]model.DashboardUnit, 0, 0)
	err := db.Table("dashboard_units").Where("project_id = ?", projectID).Find(&dashboardUnits).Error
	if err != nil {
		log.WithError(err).Error("Failed to fetch rows from dashboardUnits table")
		return dashboardUnits, err
	}
	return dashboardUnits, nil
}
