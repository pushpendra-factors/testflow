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

	gcpProjectID := flag.String("gcp_project_id", "", "Project ID on Google Cloud")
	gcpProjectLocation := flag.String("gcp_project_location", "", "Location of google cloud project cluster")

	slowQueriesThreshold := flag.Int("slow_queries_threshold", 50, "Threshold to report slow queries alert")
	sdkQueueThreshold := flag.Int("sdk_queue_threshold", 10000, "Threshold to report sdk queue size")
	integrationQueueThreshold := flag.Int("integration_queue_threshold", 1000, "Threshold to report integration queue size")
	delayedTaskThreshold := flag.Int("delayed_task_threshold", 1000, "Threshold to report delayed task size")

	overrideHealthcheckPingID := flag.String("healthcheck_ping_id", "", "Override default healthcheck ping id.")
	overrideAppName := flag.String("app_name", "", "Override default app_name.")

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
			Host:     *memSQLHost,
			Port:     *memSQLPort,
			User:     *memSQLUser,
			Name:     *memSQLName,
			Password: *memSQLPass,
			// Todo: Remove UseSSL after enabling it by environment on all workloads.
			UseSSL:      *env == C.STAGING || *env == C.PRODUCTION,
			Certiifcate: *memSQLCertificate,
			AppName:     appName,
		},
		PrimaryDatastore: *primaryDatastore,
		QueueRedisHost:   *queueRedisHost,
		QueueRedisPort:   *queueRedisPort,
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
	C.InitMetricsExporter(config.Env, config.AppName, config.GCPProjectID, config.GCPProjectLocation)
	defer C.WaitAndFlushAllCollectors(65 * time.Second)

	sqlAdminSlowQueries, factorsSlowQueries, err := store.GetStore().MonitorSlowQueries()
	if err != nil {
		log.WithError(err).Panic("Failed to run monitoring query.")
	}
	if len(factorsSlowQueries) > *slowQueriesThreshold {
		dbHealthcheckPingID := C.HealthcheckDatabaseHealthPingID
		if C.UseMemSQLDatabaseStore() {
			dbHealthcheckPingID = C.HealthcheckDatabaseHealthMemSQLPingID
		}
		C.PingHealthcheckForFailure(dbHealthcheckPingID,
			fmt.Sprintf("Slow query count %d exceeds threshold of %d", len(factorsSlowQueries), *slowQueriesThreshold))
	}

	if C.UseMemSQLDatabaseStore() {
		mqlStore.GetStore().MonitorMemSQLDiskUsage()
	}

	delayedTaskCount, sdkQueueLength, integrationQueueLength := MonitorSDKHealth(
		*delayedTaskThreshold, *sdkQueueThreshold, *integrationQueueThreshold)

	tableSizes := store.GetStore().CollectTableSizes()

	monitoringPayload := map[string]interface{}{
		"factorsSlowQueries":       factorsSlowQueries[:util.MinInt(5, len(factorsSlowQueries))],
		"sqlAdminSlowQueries":      sqlAdminSlowQueries[:util.MinInt(5, len(sqlAdminSlowQueries))],
		"factorsSlowQueriesCount":  len(factorsSlowQueries),
		"sqlAdminSlowQueriesCount": len(sqlAdminSlowQueries),
		"delayedTaskCount":         delayedTaskCount,
		"sdkQueueLength":           sdkQueueLength,
		"integrationQueueLength":   integrationQueueLength,
		"tableSizes":               tableSizes,
	}
	C.PingHealthcheckForSuccess(healthcheckPingID, monitoringPayload)
}

func MonitorSDKHealth(delayedTaskThreshold, sdkQueueThreshold, integrationQueueThreshold int) (int, int, int) {

	queueClient := C.GetServices().QueueClient
	delayedTaskCount, err := queueClient.GetBroker().GetDelayedTasksCount()
	if err != nil {
		log.WithError(err).Panic("Failed to get delayed task count from redis")
	}
	if delayedTaskCount > delayedTaskThreshold {
		C.PingHealthcheckForFailure(C.HealthcheckSDKHealthPingID,
			fmt.Sprintf("Delayed task count %d exceeds threshold of %d", delayedTaskCount, delayedTaskThreshold))
	}

	sdkQueueLength, err := queueClient.GetBroker().GetQueueLength(sdk.RequestQueue)
	if err != nil {
		log.WithError(err).Panic("Failed to get sdk_request_queue length")
	}
	if sdkQueueLength > sdkQueueThreshold {
		C.PingHealthcheckForFailure(C.HealthcheckSDKHealthPingID,
			fmt.Sprintf("SDK queue length %d exceeds threshold of %d", sdkQueueLength, sdkQueueThreshold))
	}

	integrationQueueLength, err := queueClient.GetBroker().GetQueueLength(integration.RequestQueue)
	if err != nil {
		log.WithError(err).Panic("Failed to get integration_request_queue length")
	}
	if integrationQueueLength > integrationQueueThreshold {
		C.PingHealthcheckForFailure(C.HealthcheckSDKHealthPingID,
			fmt.Sprintf("Integration queue length %d exceeds threshold of %d", integrationQueueLength, integrationQueueThreshold))
	}

	res, err := http.Get(C.SDKAssetsURL)
	if err != nil || res.StatusCode != http.StatusOK {
		C.PingHealthcheckForFailure(C.HealthcheckSDKHealthPingID,
			fmt.Sprintf("Error '%s', Code '%d' on getting SDK from %s", err.Error(), res.StatusCode, C.SDKAssetsURL))
	}

	sdkBody, err := ioutil.ReadAll(res.Body)
	if err != nil || len(sdkBody) < 20000 || string(sdkBody[0:12]) != "var factors=" {
		// Approx file size of 20k. Error out if less than that.
		C.PingHealthcheckForFailure(C.HealthcheckSDKHealthPingID,
			fmt.Sprintf("Size '%d' of SDK file lesser than expected 20k chars. Content: '%s'", len(sdkBody), string(sdkBody)))
	}

	return delayedTaskCount, sdkQueueLength, integrationQueueLength
}
