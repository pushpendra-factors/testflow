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
	}

	C.InitConf(config.Env)

	err := C.InitDB(config.DBInfo)
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
