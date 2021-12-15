package main

import (
	C "factors/config"
	"factors/integration"
	"factors/model/store"
	mqlStore "factors/model/store/memsql"
	"factors/sdk"
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
	memSQLCertificate := flag.String("memsql_cert", "", "")
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypePostgres, "Primary datastore type as memsql or postgres")

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
		DBInfo: C.DBConf{
			Host:     *dbHost,
			Port:     *dbPort,
			User:     *dbUser,
			Name:     *dbName,
			Password: *dbPass,
			AppName:  appName,
		},
		MemSQLInfo: C.DBConf{
			Host:        *memSQLHost,
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
		log.WithError(err).Panic("Failed to initalize db.")
	}
	db := C.GetServices().Db
	defer db.Close()

	err = C.InitQueueClient(config.QueueRedisHost, config.QueueRedisPort)
	if err != nil {
		log.WithError(err).Panic("Failed to initalize queue client.")
	}

	if C.IsQueueDuplicationEnabled() {
		err := C.InitDuplicateQueueClient(config.DuplicateQueueRedisHost, config.DuplicateQueueRedisPort)
		if err != nil {
			log.WithError(err).Fatal("Failed to initialize duplicate queue client.")
		}
	}

	C.InitMetricsExporter(config.Env, config.AppName, config.GCPProjectID, config.GCPProjectLocation)
	defer C.WaitAndFlushAllCollectors(65 * time.Second)

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

	sqlAdminSlowQueries, factorsSlowQueries, err := store.GetStore().MonitorSlowQueries()
	if err != nil {
		log.WithError(err).Panic("Failed to run monitoring query.")
	}

	dbHealthcheckPingID := C.HealthcheckDatabaseHealthPingID
	if C.UseMemSQLDatabaseStore() {
		dbHealthcheckPingID = C.HealthcheckDatabaseHealthMemSQLPingID
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

	delayedTaskCount, sdkQueueLength, integrationQueueLength,
		isQueueDuplicationEnabled, dupDelayedTaskCount, dupSDKQueueLength, dupIntegrationQueueLength,
		isFailure := MonitorSDKHealth(*delayedTaskThreshold, *sdkQueueThreshold, *integrationQueueThreshold)
	// Should not proceed with success ping, incase of failure.
	if isFailure {
		return
	}

	tableSizes := store.GetStore().CollectTableSizes()

	monitoringPayload := map[string]interface{}{
		"factorsSlowQueries":        factorsSlowQueries[:util.MinInt(5, len(factorsSlowQueries))],
		"sqlAdminSlowQueries":       sqlAdminSlowQueries[:util.MinInt(5, len(sqlAdminSlowQueries))],
		"factorsSlowQueriesCount":   len(factorsSlowQueries),
		"sqlAdminSlowQueriesCount":  len(sqlAdminSlowQueries),
		"delayedTaskCount":          delayedTaskCount,
		"sdkQueueLength":            sdkQueueLength,
		"integrationQueueLength":    integrationQueueLength,
		"isQueueDuplicationEnabled": isQueueDuplicationEnabled,
		"dupDelayedTaskCount":       dupDelayedTaskCount,
		"dupSDKQueueLength":         dupSDKQueueLength,
		"dupIntegrationQueueLength": dupIntegrationQueueLength,
		"tableSizes":                tableSizes,
	}
	if C.UseMemSQLDatabaseStore() {
		monitoringPayload["memsqlNodeUsageStats"] = nodeUsageStatsWithErrors
	}

	if len(analyzeStatus) > 0 {
		analyzeStatus["analyzeStatus"] = analyzeStatus
	}

	C.PingHealthcheckForSuccess(healthcheckPingID, monitoringPayload)
}

func MonitorSDKHealth(delayedTaskThreshold, sdkQueueThreshold, integrationQueueThreshold int) (int, int, int, bool, int, int, int, bool) {

	queueClient := C.GetServices().QueueClient
	duplicateQueueClient := C.GetServices().DuplicateQueueClient

	delayedTaskCount, err := queueClient.GetBroker().GetDelayedTasksCount()
	if err != nil {
		log.WithError(err).Panic("Failed to get delayed task count from redis")
	}
	if delayedTaskCount > delayedTaskThreshold {
		C.PingHealthcheckForFailure(C.HealthcheckSDKHealthPingID,
			fmt.Sprintf("Delayed task count %d exceeds threshold of %d", delayedTaskCount, delayedTaskThreshold))
	}

	var dupDelayedTaskCount int
	if C.IsQueueDuplicationEnabled() {
		dupDelayedTaskCount, err = duplicateQueueClient.GetBroker().GetDelayedTasksCount()
		if err != nil {
			log.WithError(err).Panic("Failed to get delayed task count from duplicate queue redis.")
		}
		if dupDelayedTaskCount > delayedTaskThreshold {
			C.PingHealthcheckForFailure(C.HealthcheckSDKHealthPingID,
				fmt.Sprintf("Duplicate queue delayed task count %d exceeds threshold of %d", dupDelayedTaskCount, delayedTaskThreshold))
		}
	}

	sdkQueueLength, err := queueClient.GetBroker().GetQueueLength(sdk.RequestQueue)
	if err != nil {
		log.WithError(err).Panic("Failed to get sdk_request_queue length")
	}
	if sdkQueueLength > sdkQueueThreshold {
		C.PingHealthcheckForFailure(C.HealthcheckSDKHealthPingID,
			fmt.Sprintf("SDK queue length %d exceeds threshold of %d", sdkQueueLength, sdkQueueThreshold))
	}

	var dupSdkQueueLength int
	if C.IsQueueDuplicationEnabled() {
		dupSdkQueueLength, err = duplicateQueueClient.GetBroker().GetQueueLength(sdk.RequestQueueDuplicate)
		if err != nil {
			log.WithError(err).Panic("Failed to get duplicate sdk_request_queue length")
		}
		if dupSdkQueueLength > sdkQueueThreshold {
			C.PingHealthcheckForFailure(C.HealthcheckSDKHealthPingID,
				fmt.Sprintf("SDK duplicate queue length %d exceeds threshold of %d", dupSdkQueueLength, sdkQueueThreshold))
		}
	}

	integrationQueueLength, err := queueClient.GetBroker().GetQueueLength(integration.RequestQueue)
	if err != nil {
		log.WithError(err).Panic("Failed to get integration_request_queue length")
	}
	if integrationQueueLength > integrationQueueThreshold {
		C.PingHealthcheckForFailure(C.HealthcheckSDKHealthPingID,
			fmt.Sprintf("Integration queue length %d exceeds threshold of %d", integrationQueueLength, integrationQueueThreshold))
	}

	var dupIntegrationQueueLength int
	if C.IsQueueDuplicationEnabled() {
		dupIntegrationQueueLength, err = queueClient.GetBroker().GetQueueLength(integration.RequestQueueDuplicate)
		if err != nil {
			log.WithError(err).Panic("Failed to get duplicate integration_request_queue length")
		}
		if dupIntegrationQueueLength > integrationQueueThreshold {
			C.PingHealthcheckForFailure(C.HealthcheckSDKHealthPingID,
				fmt.Sprintf("Integration duplicate queue length %d exceeds threshold of %d", dupIntegrationQueueLength, integrationQueueThreshold))
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
			C.IsQueueDuplicationEnabled(), dupDelayedTaskCount, dupSdkQueueLength, dupIntegrationQueueLength,
			true
	}

	sdkBody, err := ioutil.ReadAll(res.Body)
	if err != nil || len(sdkBody) < 20000 || string(sdkBody[0:12]) != "var factors=" {
		// Approx file size of 20k. Error out if less than that.
		C.PingHealthcheckForFailure(C.HealthcheckSDKHealthPingID,
			fmt.Sprintf("Size '%d' of SDK file lesser than expected 20k chars. Content: '%s'", len(sdkBody), string(sdkBody)))
		return delayedTaskCount, sdkQueueLength, integrationQueueLength,
			C.IsQueueDuplicationEnabled(), dupDelayedTaskCount, dupSdkQueueLength, dupIntegrationQueueLength,
			true
	}

	return delayedTaskCount, sdkQueueLength, integrationQueueLength,
		C.IsQueueDuplicationEnabled(), dupDelayedTaskCount, dupSdkQueueLength, dupIntegrationQueueLength,
		false
}
