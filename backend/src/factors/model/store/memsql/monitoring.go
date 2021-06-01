package memsql

import (
	C "factors/config"
	"factors/metrics"
	"factors/util"
	U "factors/util"
	"fmt"

	log "github.com/sirupsen/logrus"
)

type SlowQueries struct {
	ID           int64   `json:"id"`
	Time         float64 `json:"time"`
	TimeString   string  `json:"time_string"`
	Info         string  `json:"info"`
	User         string  `json:"user"`
	Command      string  `json:"command"`
	ResourcePool string  `json:"resource_pool"`
}

type DiskUsageStats struct {
	IPAddr              string `json:"ip_addr"`
	Type                string `json:"type"`
	State               string `json:"state"`
	AvailableDataDiskMB int64  `json:"available_data_disk_mb"`
}

func (store *MemSQL) MonitorSlowQueries() ([]interface{}, []interface{}, error) {
	db := C.GetServices().Db
	sqlAdminSlowQueries := make([]interface{}, 0, 0)
	factorsSlowQueries := make([]interface{}, 0, 0)

	queryStr := "SELECT id, time, info, user, command, resource_pool FROM information_schema.mv_processlist" +
		" WHERE time > 120 and db='factors' AND user != 'distributed' AND command != 'Sleep'" +
		" ORDER BY time DESC"

	rows, err := db.Raw(queryStr).Rows()
	if err != nil {
		log.WithError(err).Error("Failed to get slow queries from mv_processlist")
		return sqlAdminSlowQueries, factorsSlowQueries, err
	}

	for rows.Next() {
		var slowQuery SlowQueries
		if err := db.ScanRows(rows, &slowQuery); err != nil {
			log.WithError(err).Error("Failed to scan slow queries from db.")
			return sqlAdminSlowQueries, factorsSlowQueries, err
		}
		slowQuery.TimeString = U.SecondsToHMSString(int64(slowQuery.Time))
		slowQuery.Info = slowQuery.Info[:U.MinInt(len(slowQuery.Info), 500)]

		if slowQuery.Info != "" {
			if slowQuery.User != "factors_ro" && slowQuery.User != "factors_rw" {
				sqlAdminSlowQueries = append(sqlAdminSlowQueries, slowQuery)
			} else {
				factorsSlowQueries = append(factorsSlowQueries, slowQuery)
			}
		}
	}

	return sqlAdminSlowQueries, factorsSlowQueries, nil
}

func (store *MemSQL) MonitorMemSQLDiskUsage() {
	db := C.GetServices().Db
	queryStr := "select ip_addr, type, state, available_data_disk_mb FROM information_schema.mv_nodes"

	rows, err := db.Raw(queryStr).Rows()
	if err != nil {
		log.WithError(err).Panic("Failed to get disk usage stats")
	}

	diskUsageStats := struct {
		ErrorMessage   string
		DiskUsageStats []DiskUsageStats
	}{}
	diskUsageStats.DiskUsageStats = make([]DiskUsageStats, 0, 0)

	for rows.Next() {
		var nodeStats DiskUsageStats
		if err := db.ScanRows(rows, &nodeStats); err != nil {
			log.WithError(err).Panic("Failed to scan slow queries from db.")
		}
		if nodeStats.AvailableDataDiskMB < 20*1024 {
			// If disk available is less than 20 GB for any node, raise an alert.
			diskUsageStats.ErrorMessage = fmt.Sprintf("Disk available '%d'MB below threshold on '%s' node '%s'",
				nodeStats.AvailableDataDiskMB, nodeStats.Type, nodeStats.IPAddr)
		}
		if nodeStats.State != "online" {
			diskUsageStats.ErrorMessage = fmt.Sprintf("Node '%s' of type '%s' not online with state '%s'",
				nodeStats.IPAddr, nodeStats.Type, nodeStats.State)
		}
		diskUsageStats.DiskUsageStats = append(diskUsageStats.DiskUsageStats, nodeStats)
	}

	if diskUsageStats.ErrorMessage != "" {
		C.PingHealthcheckForFailure(C.HealthcheckMonitoringJobMemSQLPingID, diskUsageStats)
	}
}

// CollectTableSizes Captures size for major tables as metrics.
func (store *MemSQL) CollectTableSizes() map[string]string {
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

	columnarQuery := "SELECT table_name, sum(compressed_size) size" +
		" FROM information_schema.COLUMNAR_SEGMENTS WHERE database_name='factors' GROUP BY 1"
	rowstoreQuery := "SELECT table_name, sum(memory_use) size" +
		" FROM information_schema.table_statistics WHERE database_name='factors' GROUP BY 1"

	columnarTableByteSizes := getTableSizeMap(columnarQuery)
	rowstoreTableByteSizes := getTableSizeMap(rowstoreQuery)

	for tableName, tableMetric := range tableNameToMetricMap {
		columnarByteSize, isColumnar := columnarTableByteSizes[tableName]
		rowstoreByteSize, _ := rowstoreTableByteSizes[tableName]

		tableMetric = tableMetric + "_memsql"
		if isColumnar {
			// Check for columnar first. If not in columnar, then it's rowstore.
			metrics.RecordBytesSize(tableMetric, columnarByteSize)
			tableSizes[tableName] = util.BytesToReadableFormat(columnarByteSize)
		} else {
			metrics.RecordBytesSize(tableMetric, rowstoreByteSize)
			tableSizes[tableName] = util.BytesToReadableFormat(rowstoreByteSize)
		}
	}

	return tableSizes
}

func getTableSizeMap(query string) map[string]float64 {
	db := C.GetServices().Db

	tableBytesSize := make(map[string]float64)
	rows, err := db.Raw(query).Rows()
	if err != nil {
		log.WithError(err).Error("Failed to get table sizes from database")
		return tableBytesSize
	}

	for rows.Next() {
		var tableName string
		var tableSize float64
		if err := rows.Scan(&tableName, &tableSize); err != nil {
			log.WithError(err).Error("Failed to scan table size from db.")
			continue
		}

		tableBytesSize[tableName] = tableSize
	}
	return tableBytesSize
}
