package main

import (
	C "factors/config"
	"factors/integration"
	"factors/metrics"
	"factors/sdk"
	"factors/util"
	"flag"
	"fmt"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

type SlowQueries struct {
	Runtime         float64 `json:"runtime"`
	Query           string  `json:"query"`
	Pid             int64   `json:"pid"`
	Usename         string  `json:"usename"`
	ApplicationName string  `json:"application_name"`
}

func main() {
	env := flag.String("env", "development", "")
	dbHost := flag.String("db_host", "localhost", "")
	dbPort := flag.Int("db_port", 5432, "")
	dbUser := flag.String("db_user", "autometa", "")
	dbName := flag.String("db_name", "autometa", "")
	dbPass := flag.String("db_pass", "@ut0me7a", "")

	queueRedisHost := flag.String("queue_redis_host", "localhost", "")
	queueRedisPort := flag.Int("queue_redis_port", 6379, "")

	gcpProjectID := flag.String("gcp_project_id", "", "Project ID on Google Cloud")
	gcpProjectLocation := flag.String("gcp_project_location", "", "Location of google cloud project cluster")

	slowQueriesThreshold := flag.Int("slow_queries_threshold", 50, "Threshold to report slow queries alert")
	sdkQueueThreshold := flag.Int("sdk_queue_threshold", 10000, "Threshold to report sdk queue size")
	integrationQueueThreshold := flag.Int("integration_queue_threshold", 1000, "Threshold to report integration queue size")
	delayedTaskThreshold := flag.Int("delayed_task_threshold", 1000, "Threshold to report delayed task size")

	killSlowQueries := flag.Bool("kill_slow_queries", false, "Kill slow queries. TO BE USED WITH CAUTION")

	flag.Parse()
	taskID := "monitoring_job"
	healthcheckPingID := C.HealthcheckMonitoringJobPingID
	defer C.PingHealthcheckForPanic(taskID, *env, healthcheckPingID)

	config := &C.Configuration{
		AppName:            taskID,
		Env:                *env,
		GCPProjectID:       *gcpProjectID,
		GCPProjectLocation: *gcpProjectLocation,
		DBInfo: C.DBConf{
			Host:     *dbHost,
			Port:     *dbPort,
			User:     *dbUser,
			Name:     *dbName,
			Password: *dbPass,
			AppName:  taskID,
		},
		QueueRedisHost: *queueRedisHost,
		QueueRedisPort: *queueRedisPort,
	}

	C.InitConf(config.Env)
	err := C.InitDB(config.DBInfo)
	if err != nil {
		log.WithError(err).Panic("Failed to initalize db.")
	}

	err = C.InitQueueClient(config.QueueRedisHost, config.QueueRedisPort)
	if err != nil {
		log.WithError(err).Panic("Failed to initalize queue client.")
	}
	C.InitMetricsExporter(config.Env, config.AppName, config.GCPProjectID, config.GCPProjectLocation)
	defer C.WaitAndFlushAllCollectors(65 * time.Second)

	sqlAdminSlowQueries := make([]SlowQueries, 0, 0)
	factorsSlowQueries := make([]SlowQueries, 0, 0)

	db := C.GetServices().Db
	defer db.Close()

	queryStr := `SELECT EXTRACT(epoch from (now() - query_start)) as runtime,query, pid, usename, application_name FROM  pg_stat_activity` +
		` WHERE EXTRACT(epoch from (now() - query_start)) > 120 AND state = 'active' AND query NOT ILIKE '%pg_stat_activity%'` +
		` ORDER BY runtime DESC`
	rows, err := db.Raw(queryStr).Rows()
	if err != nil {
		log.WithError(err).Panic("Failed to get slow queries from pg_stat_activity")
	}

	for rows.Next() {
		var slowQuery SlowQueries
		if err := db.ScanRows(rows, &slowQuery); err != nil {
			log.WithError(err).Panic("Failed to scan slow queries from db.")
		}
		if slowQuery.Query != "" {
			if slowQuery.Usename == "cloudsqladmin" {
				sqlAdminSlowQueries = append(sqlAdminSlowQueries, slowQuery)
			} else {
				factorsSlowQueries = append(factorsSlowQueries, slowQuery)
			}
		}
	}

	if *killSlowQueries && len(factorsSlowQueries) > 0 {
		var killQuery string
		for _, slowQuery := range factorsSlowQueries {
			killQuery = killQuery + fmt.Sprintf("SELECT pg_cancel_backend(%d);", slowQuery.Pid)
		}
		_, err = db.Raw(killQuery).Rows()
		if err != nil {
			log.WithError(err).Panic("Failed to kill slow queries")
		}
		util.NotifyThroughSNS(taskID, *env, fmt.Sprintf("Killed %d slow queries. %v", len(factorsSlowQueries), factorsSlowQueries))
		return
	}

	if len(factorsSlowQueries) > *slowQueriesThreshold {
		C.PingHealthcheckForFailure(C.HealthcheckDatabaseHealthPingID,
			fmt.Sprintf("Slow query count %d exceeds threshold of %d", len(factorsSlowQueries), *slowQueriesThreshold))
	}

	queueClient := C.GetServices().QueueClient
	delayedTaskCount, err := queueClient.GetBroker().GetDelayedTasksCount()
	if err != nil {
		log.WithError(err).Panic("Failed to get delayed task count from redis")
	}
	if delayedTaskCount > *delayedTaskThreshold {
		C.PingHealthcheckForFailure(C.HealthcheckSDKHealthPingID,
			fmt.Sprintf("Delayed task count %d exceeds threshold of %d", delayedTaskCount, *delayedTaskThreshold))
	}

	sdkQueueLength, err := queueClient.GetBroker().GetQueueLength(sdk.RequestQueue)
	if err != nil {
		log.WithError(err).Panic("Failed to get sdk_request_queue length")
	}
	if sdkQueueLength > *sdkQueueThreshold {
		C.PingHealthcheckForFailure(C.HealthcheckSDKHealthPingID,
			fmt.Sprintf("SDK queue length %d exceeds threshold of %d", sdkQueueLength, *sdkQueueThreshold))
	}

	integrationQueueLength, err := queueClient.GetBroker().GetQueueLength(integration.RequestQueue)
	if err != nil {
		log.WithError(err).Panic("Failed to get integration_request_queue length")
	}
	if integrationQueueLength > *integrationQueueThreshold {
		C.PingHealthcheckForFailure(C.HealthcheckSDKHealthPingID,
			fmt.Sprintf("Integration queue length %d exceeds threshold of %d", integrationQueueLength, *integrationQueueThreshold))
	}

	tableSizes := collectTableSizes()

	slowQueriesStatus := map[string]interface{}{
		"factorsSlowQueries":       factorsSlowQueries[:util.MinInt(5, len(factorsSlowQueries))],
		"sqlAdminSlowQueries":      sqlAdminSlowQueries[:util.MinInt(5, len(sqlAdminSlowQueries))],
		"factorsSlowQueriesCount":  len(factorsSlowQueries),
		"sqlAdminSlowQueriesCount": len(sqlAdminSlowQueries),
		"delayedTaskCount":         delayedTaskCount,
		"sdkQueueLength":           sdkQueueLength,
		"integrationQueueLength":   integrationQueueLength,
		"tableSizes":               tableSizes,
	}
	C.PingHealthcheckForSuccess(healthcheckPingID, slowQueriesStatus)
}

// collectTableSizes Captures size for major tables as metrics.
func collectTableSizes() map[string]string {
	// Tables with size in GBs. Not including all tables to avoid cluttering in the chart.
	tableNameToMetricMap := map[string]string{
		"adwords_documents": metrics.BytesTableSizeAdwordsDocuments,
		"events":            metrics.BytesTableSizeEvents,
		"hubspot_documents": metrics.BytesTableSizeHubspotDocuments,
		"user_properties":   metrics.BytesTableSizeUserProperties,
		"users":             metrics.BytesTableSizeUsers,
	}

	tableSizes := make(map[string]string)
	tablesToMonitor := make([]string, 0, 0)
	for tableName := range tableNameToMetricMap {
		tablesToMonitor = append(tablesToMonitor, tableName)
	}
	query := fmt.Sprintf("SELECT relname, pg_total_relation_size(relname::text) FROM pg_stat_user_tables "+
		"WHERE relname in ('%s')", strings.Join(tablesToMonitor, "','"))

	db := C.GetServices().Db
	rows, err := db.Raw(query).Rows()
	if err != nil {
		log.WithError(err).Error("Failed to get table sizes from database")
		return tableSizes
	}

	for rows.Next() {
		var tableName string
		var tableSize float64
		if err := rows.Scan(&tableName, &tableSize); err != nil {
			log.WithError(err).Error("Failed to scan table size from db.")
		}
		tableMetric := tableNameToMetricMap[tableName]
		metrics.RecordBytesSize(tableMetric, tableSize)

		tableSizes[tableName] = util.BytesToReadableFormat(tableSize)
	}

	err = rows.Err()
	if err != nil {
		log.WithError(err).Error("Error while scanning table sizes")
	}
	return tableSizes
}
