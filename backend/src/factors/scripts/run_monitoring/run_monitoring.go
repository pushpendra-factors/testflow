package main

import (
	// "bytes"
	"encoding/json"
	"factors/company_enrichment/factors_deanon"
	C "factors/config"
	"factors/model/model"
	"factors/model/store"
	mqlStore "factors/model/store/memsql"
	"factors/util"
	"flag"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"strconv"
	"time"

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

	apiUrl := flag.String("api_url", "http://factors-dev.com:8080/health", "enter the api url")
	apiToken := flag.String("api_token", "", "enter the api token")

	queueRedisHost := flag.String("queue_redis_host", "localhost", "")
	queueRedisPort := flag.Int("queue_redis_port", 6379, "")

	duplicateQueueRedisHost := flag.String("dup_queue_redis_host", "localhost", "")
	duplicateQueueRedisPort := flag.Int("dup_queue_redis_port", 6379, "")

	gcpProjectID := flag.String("gcp_project_id", "", "Project ID on Google Cloud")
	gcpProjectLocation := flag.String("gcp_project_location", "", "Location of google cloud project cluster")

	slowQueriesThreshold := flag.Int("slow_queries_threshold", 50, "Threshold to report slow queries alert")
	sdkQueueThreshold := flag.Int("sdk_queue_threshold", 10000, "Threshold to report sdk queue size")
	integrationQueueThreshold := flag.Int("integration_queue_threshold", 1000, "Threshold to report integration queue size")
	delayedTaskThreshold := flag.Int("delayed_task_threshold", 1000, "Threshold to report delayed task size")

	overrideHealthcheckPingID := flag.String("healthcheck_ping_id", "", "Override default healthcheck ping id.")
	overrideAppName := flag.String("app_name", "", "Override default app_name.")

	enableSDKAndIntegrationRequestQueueDuplication := flag.Bool("enable_sdk_and_integration_request_queue_duplication",
		false, "Enables SDK and Integration request queue duplication monitoring.")

	enableAnalyzeTable := flag.Bool("enable_analyze_table", false, "Enables ANALYZE table if given.")
	analyzeIntervalInMins := flag.Int("analyze_tables_interval", 45,
		"Runs analyze for table, if not analyzed in given interval.")

	redisHostPersistent := flag.String("redis_host_ps", "localhost", "")
	redisPortPersistent := flag.Int("redis_port_ps", 6379, "")

	flag.Parse()

	defaultAppName := "monitoring_job"
	defaultHealthcheckPingID := C.HealthcheckMonitoringJobPingID
	if *primaryDatastore == C.DatastoreTypeMemSQL {
		defaultHealthcheckPingID = C.HealthcheckMonitoringJobMemSQLPingID
	}

	healthcheckPingID := C.GetHealthcheckPingID(defaultHealthcheckPingID, *overrideHealthcheckPingID)
	appName := C.GetAppName(defaultAppName, *overrideAppName)

	defer C.PingHealthcheckForPanic(appName, *env, healthcheckPingID)

	config := &C.Configuration{
		AppName:            appName,
		Env:                *env,
		GCPProjectID:       *gcpProjectID,
		GCPProjectLocation: *gcpProjectLocation,
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
		PrimaryDatastore:    *primaryDatastore,
		QueueRedisHost:      *queueRedisHost,
		QueueRedisPort:      *queueRedisPort,
		RedisHostPersistent: *redisHostPersistent,
		RedisPortPersistent: *redisPortPersistent,

		EnableSDKAndIntegrationRequestQueueDuplication: *enableSDKAndIntegrationRequestQueueDuplication,
		DuplicateQueueRedisHost:                        *duplicateQueueRedisHost,
		DuplicateQueueRedisPort:                        *duplicateQueueRedisPort,
	}

	C.InitConf(config)
	C.InitRedisPersistent(config.RedisHostPersistent, config.RedisPortPersistent)
	err := C.InitDB(*config)
	if err != nil {
		log.WithError(err).Fatal("Failed to initalize db.")
	}
	db := C.GetServices().Db
	defer db.Close()

	err = C.InitQueueClient(config.QueueRedisHost, config.QueueRedisPort)
	if err != nil {
		log.WithError(err).Fatal("Failed to initalize queue client.")
	}

	if C.IsQueueDuplicationEnabled() {
		err := C.InitDuplicateQueueClient(config.DuplicateQueueRedisHost, config.DuplicateQueueRedisPort)
		if err != nil {
			log.WithError(err).Fatal("Failed to initialize duplicate queue client.")
		}
	}

	C.InitMetricsExporter(config.Env, config.AppName, config.GCPProjectID, config.GCPProjectLocation)
	defer C.WaitAndFlushAllCollectors(65 * time.Second)

	apiPayload, msg, err := GetHealth(*apiToken, *apiUrl)
	if *apiToken == "" {
		log.WithError(err).Fatal("empty token")
	}
	if msg != "" {
		log.WithError(err).Error(msg)
	}

	// ANALYZE TABLE hook for updating table estimates for query planning.
	analyzeStatus := map[string]interface{}{}
	if C.UseMemSQLDatabaseStore() && *enableAnalyzeTable {
		status, failedTables := mqlStore.AnalyzeTableInAnInterval(*analyzeIntervalInMins)
		if status == http.StatusInternalServerError {
			analyzeStatus["status"] = "FAILED"
			if len(failedTables) > 0 {
				analyzeStatus["failedTables"] = failedTables
			}
		}
		analyzeStatus["status"] = "SUCCESS"
	}

	dbHealthCheckFailure := false

	dbHealthcheckPingID := C.HealthcheckDatabaseHealthPingID
	if C.UseMemSQLDatabaseStore() {
		dbHealthcheckPingID = C.HealthcheckDatabaseHealthMemSQLPingID
	}
	var factorsSlowQueries []interface{}
	var sqlAdminSlowQueries []interface{}
	if apiPayload["z_sql_admin_slow_queries"] != nil {
		sqlAdminSlowQueries = apiPayload["z_sql_admin_slow_queries"].([]interface{})
	}
	if apiPayload["z_factors_slow_queries"] != nil {
		factorsSlowQueries = apiPayload["z_factors_slow_queries"].([]interface{})
	}
	if len(factorsSlowQueries) > *slowQueriesThreshold {
		C.PingHealthcheckForFailure(dbHealthcheckPingID,
			fmt.Sprintf("Slow query count %d exceeds threshold of %d", len(factorsSlowQueries), *slowQueriesThreshold))
		dbHealthCheckFailure = true
	}

	var nodeUsageStatsWithErrors mqlStore.MemSQLNodeUsageStatsWithErrors
	if C.UseMemSQLDatabaseStore() {
		nodeUsageStatsWithErrors = mqlStore.GetStore().MonitorMemSQLDiskUsage()
		if len(nodeUsageStatsWithErrors.ErrorMessage) > 0 {
			C.PingHealthcheckForFailure(dbHealthcheckPingID, nodeUsageStatsWithErrors.ErrorMessage)
			dbHealthCheckFailure = true
		}
	}

	// ping health check success for db
	if !dbHealthCheckFailure {
		C.PingHealthcheckForSuccess(dbHealthcheckPingID, "db health check success")
	}

	var healthCheckFailurePinged bool

	var isFailure bool
	apiPayload["delayed_task_count"], apiPayload["sdk_queue_length"], apiPayload["integration_queue_length"],
		apiPayload["isQueue_duplication_enabled"], apiPayload["dup_delayed_task_count"], apiPayload["dup_sdk_queue_length"], apiPayload["dup_integration_queue_length"],
		isFailure, healthCheckFailurePinged = MonitorSDKHealth(*delayedTaskThreshold, *sdkQueueThreshold, *integrationQueueThreshold, apiPayload)
	// Should not proceed with success ping, incase of failure.
	if isFailure {
		return
	}

	// Factors Deanon enrichment count metrics monitoring
	RunFactorsDeanonEnrichmentCheck()

	monitoringPayload := map[string]interface{}{
		"factorsSlowQueries":        (factorsSlowQueries)[:util.MinInt(5, len(factorsSlowQueries))],
		"sqlAdminSlowQueries":       (sqlAdminSlowQueries)[:util.MinInt(5, len(sqlAdminSlowQueries))],
		"factorsSlowQueriesCount":   apiPayload["factors_slow_queries_count"],
		"sqlAdminSlowQueriesCount":  apiPayload["sqlAdmin_slow_queries_count"],
		"delayedTaskCount":          apiPayload["delayed_task_count"],
		"sdkQueueLength":            apiPayload["sdkQueueLength"],
		"integrationQueueLength":    apiPayload["integration_queue_length"],
		"isQueueDuplicationEnabled": apiPayload["isQueue_duplication_enabled"],
		"dupDelayedTaskCount":       apiPayload["dup_delayed_task_count"],
		"dupSDKQueueLength":         apiPayload["dup_sdk_queue_length"],
		"dupIntegrationQueueLength": apiPayload["dup_integration_queue_length"],
		"tableSizes":                apiPayload["table_sizes"],
	}
	if C.UseMemSQLDatabaseStore() {
		monitoringPayload["memsqlNodeUsageStats"] = nodeUsageStatsWithErrors
	}

	if len(analyzeStatus) > 0 {
		analyzeStatus["analyzeStatus"] = analyzeStatus
	}

	C.PingHealthcheckForSuccess(healthcheckPingID, monitoringPayload)

	if !healthCheckFailurePinged {
		C.PingHealthcheckForSuccess(C.HealthcheckSDKHealthPingID, "sdk health check success")
	}

}

// GetHealth - This method returns response of the http get request at /health
func GetHealth(apiToken string, apiUrl string) (map[string]interface{}, string, error) {
	// http get request
	apiPayload := map[string]interface{}{}
	client := &http.Client{}
	var msg string = ""
	req, err := http.NewRequest("GET", apiUrl, nil)
	req.Header.Set("Authorization", apiToken)
	resp, err := client.Do(req)
	if err != nil || (resp != nil && resp.StatusCode != http.StatusOK) {
		msg = "Failed to check health api"
		logCtx := log.WithError(err)
		if resp != nil {
			logCtx = log.WithField("status", resp.StatusCode)
		}
		logCtx.Error(msg)

		return apiPayload, msg, err
	}
	respBytes, err := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(respBytes, &apiPayload)
	if err != nil {
		msg = "Failed to unmarshall"
		return apiPayload, msg, err
	}
	return apiPayload, msg, nil
}

func MonitorSDKHealth(delayedTaskThreshold, sdkQueueThreshold, integrationQueueThreshold int,
	apiPayload map[string]interface{}) (float64, float64, float64, bool, float64, float64, float64, bool, bool) {
	var delayedTaskCount float64
	var duplicateDelayedTaskCount float64
	var sdkQueueLength float64
	var duplicateSdkQueueLength float64
	var integrationQueueLength float64
	var duplicateIntegrationQueueLength float64
	var healthCheckFailurePinged bool

	if apiPayload["delayed_task_count"] != nil {
		delayedTaskCount = apiPayload["delayed_task_count"].(float64)
	}
	if apiPayload["dup_delayed_task_count"] != nil {
		duplicateDelayedTaskCount = apiPayload["dup_delayed_task_count"].(float64)
	}
	if apiPayload["sdk_queue_length"] != nil {
		sdkQueueLength = apiPayload["sdk_queue_length"].(float64)
	}
	if apiPayload["dup_sdk_queue_length"] != nil {
		duplicateSdkQueueLength = apiPayload["dup_sdk_queue_length"].(float64)
	}
	if apiPayload["integration_queue_length"] != nil {
		integrationQueueLength = apiPayload["integration_queue_length"].(float64)
	}
	if apiPayload["dup_integration_queue_length"] != nil {
		duplicateIntegrationQueueLength = apiPayload["dup_integration_queue_length"].(float64)
	}

	if delayedTaskCount > float64(delayedTaskThreshold) {
		C.PingHealthcheckForFailure(C.HealthcheckSDKHealthPingID,
			fmt.Sprintf("Delayed task count %f exceeds threshold of %d", delayedTaskCount, delayedTaskThreshold))
		healthCheckFailurePinged = true
	}
	if C.IsQueueDuplicationEnabled() {
		if duplicateDelayedTaskCount > float64(delayedTaskThreshold) {
			C.PingHealthcheckForFailure(C.HealthcheckSDKHealthPingID,
				fmt.Sprintf("Duplicate queue delayed task count %f exceeds threshold of %d", duplicateDelayedTaskCount, delayedTaskThreshold))
			healthCheckFailurePinged = true
		}
	}

	if sdkQueueLength > float64(sdkQueueThreshold) {
		C.PingHealthcheckForFailure(C.HealthcheckSDKHealthPingID,
			fmt.Sprintf("SDK queue length %f exceeds threshold of %d", sdkQueueLength, sdkQueueThreshold))
		healthCheckFailurePinged = true
	}

	if C.IsQueueDuplicationEnabled() {
		if duplicateSdkQueueLength > float64(sdkQueueThreshold) {
			C.PingHealthcheckForFailure(C.HealthcheckSDKHealthPingID,
				fmt.Sprintf("SDK duplicate queue length %f exceeds threshold of %d", duplicateSdkQueueLength, sdkQueueThreshold))
			healthCheckFailurePinged = true
		}
	}

	if integrationQueueLength > float64(integrationQueueThreshold) {
		C.PingHealthcheckForFailure(C.HealthcheckSDKHealthPingID,
			fmt.Sprintf("Integration queue length %f exceeds threshold of %d", integrationQueueLength, integrationQueueThreshold))
		healthCheckFailurePinged = true
	}

	if C.IsQueueDuplicationEnabled() {
		if duplicateIntegrationQueueLength > float64(integrationQueueThreshold) {
			C.PingHealthcheckForFailure(C.HealthcheckSDKHealthPingID,
				fmt.Sprintf("Integration duplicate queue length %f exceeds threshold of %d", duplicateIntegrationQueueLength, integrationQueueThreshold))
			healthCheckFailurePinged = true
		}
	}

	res, err := http.Get(C.SDKAssetsURL)
	if err != nil || res.StatusCode != http.StatusOK {
		var message string
		if res == nil {
			message = fmt.Sprintf("Error '%s' and no response on getting SDK from %s", err.Error(), C.SDKAssetsURL)
		} else {
			message = fmt.Sprintf("Error '%s', Code '%d' on getting SDK from %s", err.Error(), res.StatusCode, C.SDKAssetsURL)
		}

		C.PingHealthcheckForFailure(C.HealthcheckSDKHealthPingID, message)
		healthCheckFailurePinged = true
		return delayedTaskCount, sdkQueueLength, integrationQueueLength,
			C.IsQueueDuplicationEnabled(), duplicateDelayedTaskCount, duplicateSdkQueueLength, duplicateIntegrationQueueLength,
			true, healthCheckFailurePinged
	}

	sdkBody, err := ioutil.ReadAll(res.Body)
	if err != nil || len(sdkBody) < 20000 || string(sdkBody[0:12]) != "var factors=" {
		// Approx file size of 20k. Error out if less than that.
		C.PingHealthcheckForFailure(C.HealthcheckSDKHealthPingID,
			fmt.Sprintf("Size '%d' of SDK file lesser than expected 20k chars. Content: '%s'", len(sdkBody), string(sdkBody)))
		healthCheckFailurePinged = true
		return delayedTaskCount, sdkQueueLength, integrationQueueLength,
			C.IsQueueDuplicationEnabled(), duplicateDelayedTaskCount, duplicateSdkQueueLength, duplicateIntegrationQueueLength,
			true, healthCheckFailurePinged
	}
	return delayedTaskCount, sdkQueueLength, integrationQueueLength,
		C.IsQueueDuplicationEnabled(), duplicateDelayedTaskCount, duplicateSdkQueueLength, duplicateIntegrationQueueLength, false, healthCheckFailurePinged

}

// RunFactorsDeanonEnrichmentCheck checks the last run and on the basis of that monitor the metrics related to factors deanon enrichment.
func RunFactorsDeanonEnrichmentCheck() {

	startTime := time.Now().Unix()
	var factorsDeanonAlertMap map[int64]map[string]float64
	var err error
	factorsDeanonAlertLastRun, err := model.GetFactorsDeanonAlertRedisResult()
	if err != nil {
		log.WithError(err).Error("failed to get factors deanon last run")
		return
	}
	if startTime-factorsDeanonAlertLastRun > 24*60*60 {
		factorsDeanonAlertMap, err = MonitorFactorsDeanonDailyEnrichment()
		if err != nil {
			log.Error("Failed to run factors deanon enrichment check ", time.Now())
			return
		}
		if len(factorsDeanonAlertMap) > 0 {
			C.PingHealthcheckForFailure(C.HealthcheckFactorsDeanonAlertPingID, factorsDeanonAlertMap)
			model.SetFactorsDeanonAlertRedisResult(startTime)
			return
		}
		model.SetFactorsDeanonAlertRedisResult(startTime)
		C.PingHealthcheckForSuccess(C.HealthcheckFactorsDeanonAlertPingID, factorsDeanonAlertMap)
	}

}

// MonitorFactorsDeanonDailyEnrichment fetches the project ids of the projects on paid plan and get the required metrics for each projects.
func MonitorFactorsDeanonDailyEnrichment() (map[int64]map[string]float64, error) {

	projectIds, errCode, errMsg, err := store.GetStore().GetAllProjectIdsUsingPaidPlan()
	if errCode != http.StatusFound {
		log.WithError(err).Error(errMsg)
		C.PingHealthcheckForFailure(C.HealthcheckFactorsDeanonAlertPingID, "Failed to fetch project ids for monitoring factors deanonymisation.")
		return nil, err
	}

	alertMap := make(map[int64]map[string]float64)
	for _, projectId := range projectIds {

		projectSettings, errCode := store.GetStore().GetProjectSetting(projectId)
		if errCode != http.StatusFound {
			log.WithField("project_id", projectId).Error("Failed to fetch project details.")
			continue
		}

		isEligible, err := IsProjectEligibleForFactorsDeanonAlerts(projectSettings)
		if err != nil {
			log.WithField("project_id", projectId).Error("Failed to check eligibilty.")
			continue
		}
		if !isEligible {
			log.WithField("project_id", projectId).Info("Project not eligible for alert.")
			continue
		}

		createdAt := projectSettings.CreatedAt.Unix()
		currentTime := time.Now().Unix()
		diffTime := currentTime - createdAt

		if diffTime > 8*24*60*60 && diffTime <= 15*24*60*60 {
			totalCountDiff, successfulCountDiff, err := CheckFactorsDeanonymisationAlertForRecentProjects(projectId)
			if err != nil {
				log.Error("Failed to get redis values")
				return alertMap, err
			}
			if math.Abs(totalCountDiff) >= 0.3 || math.Abs(successfulCountDiff) >= 0.3 {
				alertMap[projectId] = map[string]float64{"total_count": totalCountDiff, "successful_count": successfulCountDiff}
			}
		} else if diffTime > 15*24*60*60 {
			totalCountDiff, successfulCountDiff, err := CheckFactorsDeanonymisationAlertForOlderProjects(projectId)
			if err != nil {
				log.Error("Failed to get redis values")
				return alertMap, err
			}
			if math.Abs(totalCountDiff) >= 0.3 || math.Abs(successfulCountDiff) >= 0.3 {
				alertMap[projectId] = map[string]float64{"total_count": totalCountDiff, "successful_count": successfulCountDiff}
			}
		}

	}

	return alertMap, nil
}

// CheckFactorsDeanonymisationAlertForRecentProjects checks alert data for projects that are 8 to 15 days older.
func CheckFactorsDeanonymisationAlertForRecentProjects(projectId int64) (float64, float64, error) {

	yesterdayDate := time.Now().AddDate(0, 0, -1).Format(util.DATETIME_FORMAT_YYYYMMDD)
	eightDaysAgoDate := time.Now().AddDate(0, 0, -8).Format(util.DATETIME_FORMAT_YYYYMMDD)

	yesterdayUint64, _ := strconv.ParseUint(yesterdayDate, 10, 64)
	eightDaysAgoUint64, _ := strconv.ParseUint(eightDaysAgoDate, 10, 64)

	// Fetching total api count for n-1 and n-8
	yesterdayTotalApiCount, err := model.GetSixSignalAPITotalHitCountCacheResult(projectId, yesterdayUint64)
	if err != nil {
		return 0, 0, err
	}
	eightDaysAgoTotalApiCount, err := model.GetSixSignalAPITotalHitCountCacheResult(projectId, eightDaysAgoUint64)
	if err != nil {
		return 0, 0, err
	}

	// Fetching successful(domain is present) api count for n-1 and n-8
	yesterdaySuccessfulApiCount, err := model.GetSixSignalAPICountCacheResult(projectId, yesterdayUint64)
	if err != nil {
		return 0, 0, err
	}
	eightDaysAgoSuccessfulApiCount, err := model.GetSixSignalAPICountCacheResult(projectId, eightDaysAgoUint64)
	if err != nil {
		return 0, 0, err
	}

	totalCountDiff := float64(yesterdayTotalApiCount-eightDaysAgoTotalApiCount) / float64(eightDaysAgoTotalApiCount)
	successfulCountDiff := float64(yesterdaySuccessfulApiCount-eightDaysAgoSuccessfulApiCount) / float64(eightDaysAgoSuccessfulApiCount)

	if eightDaysAgoTotalApiCount <= 0 {
		return 0, 0, nil
	}

	return totalCountDiff, successfulCountDiff, nil

}

// CheckFactorsDeanonymisationAlertForOlderProjects checks alert data for projects that are older than 15 days.
func CheckFactorsDeanonymisationAlertForOlderProjects(projectId int64) (float64, float64, error) {

	currentDate := time.Now()
	var last14DaysTotalApiCount []int
	var last14DaysSuccessfulApiCount []int
	for i := 1; i <= 14; i++ {
		// Subtract i days from the current date
		date, _ := strconv.ParseUint(currentDate.AddDate(0, 0, -i).Format(util.DATETIME_FORMAT_YYYYMMDD), 10, 64)

		totalApiCount, err := model.GetSixSignalAPITotalHitCountCacheResult(projectId, date)
		if err != nil {
			return 0, 0, err
		}
		successfulApiCount, err := model.GetSixSignalAPICountCacheResult(projectId, date)
		if err != nil {
			return 0, 0, err
		}

		last14DaysTotalApiCount = append(last14DaysTotalApiCount, totalApiCount)
		last14DaysSuccessfulApiCount = append(last14DaysSuccessfulApiCount, successfulApiCount)
	}

	totalCountThisWeekSum := util.Sum(last14DaysTotalApiCount[:7])
	totalCountPastWeekSum := util.Sum(last14DaysTotalApiCount[7:])
	totalCountDiff := float64(totalCountThisWeekSum-totalCountPastWeekSum) / float64(totalCountPastWeekSum)

	successfulCountThisWeekSum := util.Sum(last14DaysSuccessfulApiCount[:7])
	successfulCountPastWeekSum := util.Sum(last14DaysSuccessfulApiCount[7:])
	successfulCountDiff := float64(successfulCountThisWeekSum-successfulCountPastWeekSum) / float64(successfulCountPastWeekSum)

	if totalCountPastWeekSum <= 0 {
		return 0, 0, nil
	}

	return totalCountDiff, successfulCountDiff, nil
}

// IsProjectEligibleForFactorsDeanonAlerts checks if the project passes the creteria for factors deanon monitoring and alerts.
func IsProjectEligibleForFactorsDeanonAlerts(projectSettings *model.ProjectSetting) (bool, error) {

	projectId := projectSettings.ProjectId
	logCtx := log.WithField("project_id", projectId)

	featureFlag, err := store.GetStore().GetFeatureStatusForProjectV2(projectId, model.FEATURE_FACTORS_DEANONYMISATION, false)
	if err != nil {
		logCtx.Error("Failed to fetch feature flag")
		return false, err
	}
	if !featureFlag {
		return false, nil
	}

	isDeanonQuotaAvailable, err := factors_deanon.CheckingFactorsDeanonQuotaLimit(projectId)
	if err != nil {
		logCtx.Error("Error in checking deanon quota exhausted.")
		return false, err
	}
	if !isDeanonQuotaAvailable {
		return false, nil
	}

	intFactorsDeanon := projectSettings.IntFactorsSixSignalKey

	eligible := featureFlag && isDeanonQuotaAvailable && *intFactorsDeanon

	return eligible, nil
}
