package main

import (
	C "factors/config"
	"factors/integration/slack"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"flag"
	"fmt"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"
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

	overrideHealthcheckPingID := flag.String("healthcheck_ping_id", "", "Override default healthcheck ping id.")
	sentryDSN := flag.String("sentry_dsn", "", "Sentry DSN")

	overrideAppName := flag.String("app_name", "", "Override default app_name.")
	enableFeatureGatesV2 := flag.Bool("enable_feature_gates_v2", false, "")

	flag.Parse()

	if *env != "development" &&
		*env != "staging" &&
		*env != "production" {
		err := fmt.Errorf("env [ %s ] not recognised", *env)
		panic(err)
	}

	defaultAppName := "slack_users_sync"
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
		PrimaryDatastore:     *primaryDatastore,
		SentryDSN:            *sentryDSN,
		EnableFeatureGatesV2: *enableFeatureGatesV2,
	}
	defaultHealthcheckPingID := C.HealthcheckEventTriggerAlertPingID
	healthcheckPingID := C.GetHealthcheckPingID(defaultHealthcheckPingID, *overrideHealthcheckPingID)
	C.InitConf(config)
	C.InitSentryLogging(config.SentryDSN, config.AppName)

	err := C.InitDB(*config)
	if err != nil {
		log.Error("Failed to initialize DB.")
		os.Exit(1)
	}

	db := C.GetServices().Db
	defer db.Close()

	finalStatus := make(map[string]interface{})
	success := true
	projectIDs, _ := store.GetStore().GetAllProjectIDs()

	for _, projectID := range projectIDs {
		available := true
		if *enableFeatureGatesV2 {
			available, err = store.GetStore().GetFeatureStatusForProjectV2(projectID, model.FEATURE_SLACK, false)
			if err != nil {
				log.WithError(err).Error("Failed to get feature status in slack job for project ID ", projectID)
				finalStatus[fmt.Sprintf("Failure-Feature-Status %v", projectID)] = true
			}
		}

		if !available {
			log.Error("Feature Not Available... Skipping slack users sync job for project ID ", projectID)
			return
		}

		err := SlackUserListSync(projectID)
		if err != nil {
			success = false
			finalStatus[fmt.Sprintf("SlackUsersSync-Failed-%d", projectID)] = true
			log.WithField("project_id", projectID).Error("slack users sync job failed for project")
		}

	}
	if !success {
		C.PingHealthcheckForFailure(healthcheckPingID, finalStatus)
	} else {
		C.PingHealthcheckForSuccess(healthcheckPingID, finalStatus)
	}

}

func SlackUserListSync(projectID int64) error {
	projectAgentMappings, errCode := store.GetStore().GetProjectAgentMappingsByProjectId(projectID)
	if errCode != http.StatusFound {
		log.WithField("project_id", projectID).Error("no project agent mapping found")
	}

	slackUsersSuccess := false
	for _, agent := range projectAgentMappings {
		//slack api to get slack users
		slackUsers, errCode, err := slack.GetSlackUsersList(projectID, agent.AgentUUID)
		if err != nil || errCode != http.StatusFound {
			if errCode == http.StatusExpectationFailed {
				//if the agent has not re-integrated their slack, we will get this error so trying for a different agent
				continue
			}
			log.WithFields(log.Fields{"project_id": projectID, "agent_id": agent.AgentUUID}).
				Error("can not get slack users for the project-agent combination")
			continue
		}
		if slackUsers == nil {
			log.WithFields(log.Fields{"project_id": projectID, "agent_id": agent.AgentUUID}).
				Warn("no slack users found")
			continue
		}

		slackUsersSuccess = true
		jsonSlackUsers, err := U.EncodeStructTypeToPostgresJsonb(slackUsers)
		if err != nil {
			log.WithFields(log.Fields{"project_id": projectID, "agent_id": agent.AgentUUID}).
				Error("unable to encode slackMember struct to json")
			return err
		}
		
		//update slack_users_sync table
		fields := make(map[string]interface{})
		fields["agent_id"] = agent.AgentUUID
		fields["users_list"] = jsonSlackUsers
		errCode, err = store.GetStore().UpdateSlackUsersListForProject(projectID, fields)
		if err != nil || errCode != http.StatusOK {
			log.WithFields(log.Fields{"project_id": projectID, "agent_id": agent.AgentUUID}).
				Error("failed to update in db")
			return err
		}
		// if atleast one update is successful, then move to other project
		break
	}
	if !slackUsersSuccess {
		log.WithField("project_id", projectID).Error("no slack user for the project")
		return fmt.Errorf("no slack users for the project")
	}

	return nil
}
