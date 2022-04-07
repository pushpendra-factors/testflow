package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"

	C "factors/config"
	"factors/model/store"
	"factors/util"
)

func main() {
	projectIds := flag.String("project_ids",
		"", "List or projects ids to add dashboard units")

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

	flag.Parse()

	if *env != C.DEVELOPMENT &&
		*env != C.STAGING &&
		*env != C.PRODUCTION {
		err := fmt.Errorf("env [ %s ] not recognised", *env)
		panic(err)
	}

	taskID := "Task#CreateWebAnalyticsDashboard"
	defer util.NotifyOnPanic(taskID, *env)

	config := &C.Configuration{
		AppName: "add_web_analytics_dashboard",
		Env:     *env,
		DBInfo: C.DBConf{
			Host:     *dbHost,
			Port:     *dbPort,
			User:     *dbUser,
			Name:     *dbName,
			Password: *dbPass,
		},
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

	err := C.InitDB(*config)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize db.")
	}

	_, allowedProjects, _ := C.GetProjectsFromListWithAllProjectSupport(*projectIds, "")
	addProjectIds, errCode := store.GetStore().GetProjectsWithoutWebAnalyticsDashboard(allowedProjects)

	if errCode != http.StatusFound {
		if errCode == http.StatusNotFound {
			log.Error("No projects without web analytics dashbaord.")
		} else {
			log.Error("Failed to get projects without web analytics dashbaord.")
		}

		os.Exit(0)
	}

	for _, projectId := range addProjectIds {
		logCtx := log.WithField("project_id", projectId)

		agentUuid, errCode := store.GetStore().GetPrimaryAgentOfProject(projectId)
		if errCode != http.StatusFound {
			logCtx.Error("Failed to get primary agent uuid for creating dashboard.")
			continue
		}

		errCode = store.GetStore().CreateWebAnalyticsDefaultDashboardWithUnits(projectId, agentUuid)
		if errCode != http.StatusCreated {
			logCtx.WithField("err_code", errCode).Error("Failed to create web analytics dashboard.")
			continue
		}
	}

	log.Info("Successfully added created web analytics dashboard.")
}
