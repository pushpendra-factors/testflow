package main

import (
	// "bytes"
	"encoding/json"
	C "factors/config"
	mqlStore "factors/model/store/memsql"
	"factors/util"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
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
		PrimaryDatastore: *primaryDatastore,
		QueueRedisHost:   *queueRedisHost,
		QueueRedisPort:   *queueRedisPort,

		EnableSDKAndIntegrationRequestQueueDuplication: *enableSDKAndIntegrationRequestQueueDuplication,
		DuplicateQueueRedisHost:                        *duplicateQueueRedisHost,
		DuplicateQueueRedisPort:                        *duplicateQueueRedisPort,
	}

	C.InitConf(config)
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
	}

	var nodeUsageStatsWithErrors mqlStore.MemSQLNodeUsageStatsWithErrors
	if C.UseMemSQLDatabaseStore() {
		nodeUsageStatsWithErrors = mqlStore.GetStore().MonitorMemSQLDiskUsage()
		if len(nodeUsageStatsWithErrors.ErrorMessage) > 0 {
			C.PingHealthcheckForFailure(dbHealthcheckPingID, nodeUsageStatsWithErrors.ErrorMessage)
		}
	}

	var isFailure bool
	apiPayload["delayed_task_count"], apiPayload["sdk_queue_length"], apiPayload["integration_queue_length"],
		apiPayload["isQueue_duplication_enabled"], apiPayload["dup_delayed_task_count"], apiPayload["dup_sdk_queue_length"], apiPayload["dup_integration_queue_length"],
		isFailure = MonitorSDKHealth(*delayedTaskThreshold, *sdkQueueThreshold, *integrationQueueThreshold, apiPayload)
	// Should not proceed with success ping, incase of failure.
	if isFailure {
		return
	}

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
	apiPayload map[string]interface{}) (float64, float64, float64, bool, float64, float64, float64, bool) {
	var delayedTaskCount float64
	var duplicateDelayedTaskCount float64
	var sdkQueueLength float64
	var duplicateSdkQueueLength float64
	var integrationQueueLength float64
	var duplicateIntegrationQueueLength float64

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
	}
	if C.IsQueueDuplicationEnabled() {
		if duplicateDelayedTaskCount > float64(delayedTaskThreshold) {
			C.PingHealthcheckForFailure(C.HealthcheckSDKHealthPingID,
				fmt.Sprintf("Duplicate queue delayed task count %f exceeds threshold of %d", duplicateDelayedTaskCount, delayedTaskThreshold))
		}
	}

	if sdkQueueLength > float64(sdkQueueThreshold) {
		C.PingHealthcheckForFailure(C.HealthcheckSDKHealthPingID,
			fmt.Sprintf("SDK queue length %f exceeds threshold of %d", sdkQueueLength, sdkQueueThreshold))
	}

	if C.IsQueueDuplicationEnabled() {
		if duplicateSdkQueueLength > float64(sdkQueueThreshold) {
			C.PingHealthcheckForFailure(C.HealthcheckSDKHealthPingID,
				fmt.Sprintf("SDK duplicate queue length %f exceeds threshold of %d", duplicateSdkQueueLength, sdkQueueThreshold))
		}
	}

	if integrationQueueLength > float64(integrationQueueThreshold) {
		C.PingHealthcheckForFailure(C.HealthcheckSDKHealthPingID,
			fmt.Sprintf("Integration queue length %f exceeds threshold of %d", integrationQueueLength, integrationQueueThreshold))
	}

	if C.IsQueueDuplicationEnabled() {
		if duplicateIntegrationQueueLength > float64(integrationQueueThreshold) {
			C.PingHealthcheckForFailure(C.HealthcheckSDKHealthPingID,
				fmt.Sprintf("Integration duplicate queue length %f exceeds threshold of %d", duplicateIntegrationQueueLength, integrationQueueThreshold))
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
		return delayedTaskCount, sdkQueueLength, integrationQueueLength,
			C.IsQueueDuplicationEnabled(), duplicateDelayedTaskCount, duplicateSdkQueueLength, duplicateIntegrationQueueLength,
			true
	}

	sdkBody, err := ioutil.ReadAll(res.Body)
	if err != nil || len(sdkBody) < 20000 || string(sdkBody[0:12]) != "var factors=" {
		// Approx file size of 20k. Error out if less than that.
		C.PingHealthcheckForFailure(C.HealthcheckSDKHealthPingID,
			fmt.Sprintf("Size '%d' of SDK file lesser than expected 20k chars. Content: '%s'", len(sdkBody), string(sdkBody)))
		return delayedTaskCount, sdkQueueLength, integrationQueueLength,
			C.IsQueueDuplicationEnabled(), duplicateDelayedTaskCount, duplicateSdkQueueLength, duplicateIntegrationQueueLength,
			true
	}
	return delayedTaskCount, sdkQueueLength, integrationQueueLength,
		C.IsQueueDuplicationEnabled(), duplicateDelayedTaskCount, duplicateSdkQueueLength, duplicateIntegrationQueueLength, false

}
