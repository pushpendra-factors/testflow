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
	Runtime float64 `json:"runtime"`
	Query   string  `json:"query"`
	Pid     int64   `json:"pid"`
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

	flag.Parse()
	taskID := "monitoring_job"

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
		},
		QueueRedisHost: *queueRedisHost,
		QueueRedisPort: *queueRedisPort,
	}

	C.InitConf(config.Env)
	err := C.InitDB(config.DBInfo)
	if err != nil {
		log.WithError(err).Fatal("Failed to initalize db.")
	}

	err = C.InitQueueClient(config.QueueRedisHost, config.QueueRedisPort)
	if err != nil {
		log.WithError(err).Fatal("Failed to initalize queue client.")
	}
	C.InitMetricsExporter(config.Env, config.AppName, config.GCPProjectID, config.GCPProjectLocation)
	defer C.WaitAndFlushAllCollectors(65 * time.Second)

	slowQueries := make([]SlowQueries, 0, 0)

	db := C.GetServices().Db
	defer db.Close()

	queryStr := `SELECT EXTRACT(epoch from (now() - query_start)) as runtime,query, pid FROM  pg_stat_activity` +
		` WHERE EXTRACT(epoch from (now() - query_start)) > 120 AND state = 'active' AND query NOT ILIKE '%pg_stat_activity%'` +
		` ORDER BY runtime DESC LIMIT 10`
	rows, err := db.Raw(queryStr).Rows()
	if err != nil {
		log.WithError(err).Error("Failed to get slow queries from pg_stat_activity")
	}

	for rows.Next() {
		var slowQuery SlowQueries
		if err := db.ScanRows(rows, &slowQuery); err != nil {
			log.WithError(err).Error("Failed to scan slow queries from db.")
		}
		if slowQuery.Query != "" {
			slowQueries = append(slowQueries, slowQuery)
		}
	}

	queueClient := C.GetServices().QueueClient
	delayedTaskCount, err := queueClient.GetBroker().GetDelayedTasksCount()
	if err != nil {
		log.WithError(err).Error("Failed to get delayed task count from redis")
	}

	sdkQueueLength, err := queueClient.GetBroker().GetQueueLength(sdk.RequestQueue)
	if err != nil {
		log.WithError(err).Error("Failed to get sdk_request_queue length")
	}

	integrationQueueLength, err := queueClient.GetBroker().GetQueueLength(integration.RequestQueue)
	if err != nil {
		log.WithError(err).Error("Failed to get integration_request_queue length")
	}

	tableSizes := collectTableSizes()

	slowQueriesStatus := map[string]interface{}{
		"slowQueries":            slowQueries,
		"delayedTaskCount":       delayedTaskCount,
		"sdkQueueLength":         sdkQueueLength,
		"integrationQueueLength": integrationQueueLength,
		"tableSizes":             tableSizes,
	}

	if *env == "development" {
		log.Info(slowQueriesStatus)
	} else {
		if len(slowQueries) > 0 || delayedTaskCount > 1000 ||
			sdkQueueLength > 1000 || integrationQueueLength > 1000 {
			if err := util.NotifyThroughSNS(taskID, *env, slowQueriesStatus); err != nil {
				log.WithError(err).Error("Failed to notify slow queries status.")
			} else {
				log.Info("Notified slow queries status.")
			}
		}
	}
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
		var tableSize int64
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
