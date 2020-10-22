package main

import (
	C "factors/config"
	H "factors/handler"
	IntSalesforce "factors/integration/salesforce"
	"factors/metrics"
	M "factors/model"
	"factors/util"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

type salesforceSyncStatus struct {
	Success  []IntSalesforce.ObjectStatus `json:"success"`
	Failures []IntSalesforce.ObjectStatus `json:"failures,omitempty"`
}

type salesforceJobStatus struct {
	SyncStatus   salesforceSyncStatus   `json:"sync_status"`
	EnrichStatus []IntSalesforce.Status `json:"enrich_status"`
}

func main() {
	env := flag.String("env", "development", "")
	dbHost := flag.String("db_host", "localhost", "")
	dbPort := flag.Int("db_port", 5432, "")
	dbUser := flag.String("db_user", "autometa", "")
	dbName := flag.String("db_name", "autometa", "")
	dbPass := flag.String("db_pass", "@ut0me7a", "")
	salesforceAppID := flag.String("salesforce_app_id", "", "")
	salesforceAppSecret := flag.String("salesforce_app_secret", "", "")
	apiDomain := flag.String("api_domain", "factors-dev.com:8080", "")
	sentryDSN := flag.String("sentry_dsn", "", "Sentry DSN")
	redisHost := flag.String("redis_host", "localhost", "")
	redisPort := flag.Int("redis_port", 6379, "")
	redisHostPersistent := flag.String("redis_host_ps", "localhost", "")
	redisPortPersistent := flag.Int("redis_port_ps", 6379, "")
	isRealTimeEventUserCachingEnabled := flag.Bool("enable_real_time_event_user_caching", false, "If the real time caching is enabled")
	realTimeEventUserCachingProjectIds := flag.String("real_time_event_user_caching_project_ids", "",
		"If the real time caching is enabled and the whitelisted projectids")

	gcpProjectID := flag.String("gcp_project_id", "", "Project ID on Google Cloud")
	gcpProjectLocation := flag.String("gcp_project_location", "", "Location of google cloud project cluster")

	flag.Parse()

	if *env != "development" && *env != "staging" && *env != "production" {
		panic(fmt.Errorf("env [ %s ] not recognised", *env))
	}

	if *salesforceAppID == "" || *salesforceAppSecret == "" {
		panic(fmt.Errorf("salesforce_app_id or salesforce_app_secret not recognised"))
	}

	config := &C.Configuration{
		AppName:            "salesforce_enrich",
		Env:                *env,
		GCPProjectID:       *gcpProjectID,
		GCPProjectLocation: *gcpProjectLocation,
		DBInfo: C.DBConf{
			Host:     *dbHost,
			Port:     *dbPort,
			User:     *dbUser,
			Name:     *dbName,
			Password: *dbPass,
		},
		APIDomain:                          *apiDomain,
		SentryDSN:                          *sentryDSN,
		SalesforceAppID:                    *salesforceAppID,
		SalesforceAppSecret:                *salesforceAppSecret,
		RedisHost:                          *redisHost,
		RedisPort:                          *redisPort,
		RedisHostPersistent:                *redisHostPersistent,
		RedisPortPersistent:                *redisPortPersistent,
		IsRealTimeEventUserCachingEnabled:  *isRealTimeEventUserCachingEnabled,
		RealTimeEventUserCachingProjectIds: *realTimeEventUserCachingProjectIds,
	}

	C.InitConf(config.Env)
	C.InitSalesforceConfig(config.SalesforceAppID, config.SalesforceAppSecret)
	C.InitEventUserRealTimeCachingConfig(config.IsRealTimeEventUserCachingEnabled, config.RealTimeEventUserCachingProjectIds)
	C.InitRedis(config.RedisHost, config.RedisPort)
	C.InitRedisPersistent(config.RedisHostPersistent, config.RedisPortPersistent)
	C.InitSentryLogging(config.SentryDSN, config.AppName)
	C.InitMetricsExporter(config.Env, config.AppName, config.GCPProjectID, config.GCPProjectLocation)
	defer C.WaitAndFlushAllCollectors(65 * time.Second)

	err := C.InitDB(config.DBInfo)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{"env": *env,
			"host": *dbHost, "port": *dbPort}).Fatal("Failed to initialize DB.")
		os.Exit(0)
	}

	db := C.GetServices().Db
	defer db.Close()

	taskID := "Task#SalesforceEnrich"
	defer util.NotifyOnPanic(taskID, *env)

	syncInfo, status := M.GetSalesforceSyncInfo()
	if status != http.StatusFound {
		log.Errorf("Failed to get salesforce syncinfo: %d", status)
	}

	var syncStatus salesforceSyncStatus
	for pid, projectSettings := range syncInfo.ProjectSettings {
		accessToken, err := IntSalesforce.GetAccessToken(projectSettings, H.GetSalesforceRedirectURL())
		if err != nil {
			log.WithField("project_id", pid).Errorf("Failed to get salesforce access token: %s", err)
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

	// salesforce enrich
	salesforceEnabledProjects, status := M.GetAllSalesforceProjectSettings()
	if status != http.StatusFound {
		log.Fatal("No projects enabled salesforce integration.")
	}

	statusList := make([]IntSalesforce.Status, 0, 0)
	for _, settings := range salesforceEnabledProjects {
		status := IntSalesforce.Enrich(settings.ProjectID)
		statusList = append(statusList, status...)
	}

	var jobStatus salesforceJobStatus
	jobStatus.SyncStatus = syncStatus
	jobStatus.EnrichStatus = statusList
	err = util.NotifyThroughSNS("salesforce_enrich", *env, jobStatus)
	if err != nil {
		log.WithError(err).Fatal("Failed to notify through SNS on salesforce enrich.")
	}
	metrics.Increment(metrics.IncrCronSalesforceEnrichSuccess)
}
