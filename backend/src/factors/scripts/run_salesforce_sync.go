package main

import (
	C "factors/config"
	H "factors/handler"
	IntSalesforce "factors/integration/salesforce"
	M "factors/model"
	"factors/util"
	"flag"
	"fmt"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"
)

type SyncStatus struct {
	Success  []IntSalesforce.SalesforceObjectStatus `json:"success"`
	Failures []IntSalesforce.SalesforceObjectStatus `json:"failures,omitempty"`
}

func main() {
	env := flag.String("env", "development", "")
	dbHost := flag.String("db_host", "localhost", "")
	dbPort := flag.Int("db_port", 5432, "")
	dbUser := flag.String("db_user", "autometa", "")
	dbName := flag.String("db_name", "autometa", "")
	dbPass := flag.String("db_pass", "@ut0me7a", "")
	salesforceAppId := flag.String("salesforce_app_id", "", "")
	salesforceAppSecret := flag.String("salesforce_app_secret", "", "")
	sentryDSN := flag.String("sentry_dsn", "", "Sentry DSN")

	flag.Parse()

	if *salesforceAppId == "" || *salesforceAppSecret == "" {
		panic(fmt.Errorf("salesforce_app_secret or salesforce_app_secret not recognised"))
	}

	config := &C.Configuration{
		AppName: "salesforce_sync",
		Env:     *env,
		DBInfo: C.DBConf{
			Host:     *dbHost,
			Port:     *dbPort,
			User:     *dbUser,
			Name:     *dbName,
			Password: *dbPass,
		},
		SentryDSN:           *sentryDSN,
		SalesforceAppID:     *salesforceAppId,
		SalesforceAppSecret: *salesforceAppSecret,
	}

	C.InitConf(config.Env)
	C.InitSalesforceConfig(config.SalesforceAppID, config.SalesforceAppSecret)
	C.InitSentryLogging(config.SentryDSN, config.AppName)
	defer C.SafeFlushSentryHook()

	err := C.InitDB(config.DBInfo)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{"env": *env,
			"host": *dbHost, "port": *dbPort}).Fatal("Failed to initialize DB.")
		os.Exit(0)
	}

	taskID := "Task#SalesforceSync"
	defer util.NotifyOnPanic(taskID, *env)

	syncInfo, status := M.GetSalesforceSyncInfo()
	if status != http.StatusFound {
		log.Errorf("Failed to get salesforce syncinfo: %d", status)
	}

	var syncStatus SyncStatus
	for pid, projectSettings := range syncInfo.ProjectSettings {
		accessToken, err := IntSalesforce.GetAccessToken(projectSettings, H.GetSalesforceRedirectURL())
		if err != nil {
			log.WithField("project_id", pid).Errorf("Failed to get salesforce access token: %d", status)
			continue
		}

		objectStatus := IntSalesforce.SyncDocuments(projectSettings, syncInfo.LastSyncInfo[pid], accessToken)
		for i := range objectStatus {
			if objectStatus[i].Status != "Success" {
				syncStatus.Failures = append(syncStatus.Failures, objectStatus[i])
			} else {
				syncStatus.Success = append(syncStatus.Success, objectStatus[i])
			}
		}
	}
	err = util.NotifyThroughSNS("salesforce_sync", *env, syncStatus)
	if err != nil {
		log.WithError(err).Fatal("Failed to notify through SNS on salesforce sync.")
	}
}
