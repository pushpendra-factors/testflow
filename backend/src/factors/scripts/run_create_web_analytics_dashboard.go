package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"

	C "factors/config"
	"factors/model/model"
	"factors/model/store"
	"factors/util"
)

func getProjectIdsWithoutWebAnalyticsDashboard(onlyProjectsMap map[uint64]bool) (projectIds []uint64, errCode int) {

	logCtx := log.WithField("projects", onlyProjectsMap)

	onlyProjectIds := make([]uint64, 0, len(onlyProjectsMap))
	for k := range onlyProjectsMap {
		onlyProjectIds = append(onlyProjectIds, k)
	}

	projectIds = make([]uint64, 0, 0)

	db := C.GetServices().Db
	queryStmnt := "SELECT id FROM projects WHERE id not in (SELECT distinct(project_id) FROM dashboards WHERE dashboards.name = '" + model.DefaultDashboardWebsiteAnalytics + "')"

	//TODO(Maisa): create util function for joining []uint64
	inProjectIds := ""
	for i, opid := range onlyProjectIds {
		inProjectIds = inProjectIds + fmt.Sprintf("%d", opid)

		if i < len(onlyProjectIds)-1 {
			inProjectIds = inProjectIds + ","
		}
	}

	if len(onlyProjectIds) > 0 {
		queryStmnt = queryStmnt + " " + fmt.Sprintf("AND id IN (%s)", inProjectIds)
	}

	rows, err := db.Raw(queryStmnt).Rows()
	if err != nil {
		logCtx.WithError(err).
			Error("Failed to get projectIds on getProjectIdsWithoutWebAnalyticsDashboard.")
		return projectIds, http.StatusInternalServerError
	}
	defer rows.Close()

	for rows.Next() {
		var projectId uint64

		if err = rows.Scan(&projectId); err != nil {
			logCtx.WithError(err).
				Error("Failed to scan rows on getProjectIdsWithoutWebAnalyticsDashboard.")
			return projectIds, http.StatusInternalServerError
		}

		projectIds = append(projectIds, projectId)
	}

	return projectIds, http.StatusFound
}

func getPrimaryAgentOfProject(projectId uint64) (uuid string, errCode int) {
	db := C.GetServices().Db

	var projectAgentMappings []model.ProjectAgentMapping
	err := db.Limit(1).Order("created_at ASC").
		Where("project_id = ?", projectId).Find(&projectAgentMappings).Error
	if err != nil {
		log.WithError(err).Error("Failed to get primary agent of project.")
		return "", http.StatusInternalServerError
	}

	if len(projectAgentMappings) == 0 {
		return "", http.StatusNotFound
	}

	return projectAgentMappings[0].AgentUUID, http.StatusFound
}

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
	addProjectIds, errCode := getProjectIdsWithoutWebAnalyticsDashboard(allowedProjects)

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

		agentUuid, errCode := getPrimaryAgentOfProject(projectId)
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
