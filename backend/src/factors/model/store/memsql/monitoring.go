package memsql

import (
	C "factors/config"
	"factors/metrics"
	"factors/model/model"
	"factors/util"
	U "factors/util"
	"fmt"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

type SlowQueries struct {
	ID           int64   `json:"id"`
	Time         float64 `json:"time"`
	ProjectName  string  `json:"project_name"`
	TimeString   string  `json:"time_string"`
	Info         string  `json:"info"`
	User         string  `json:"user"`
	Command      string  `json:"command"`
	ResourcePool string  `json:"resource_pool"`
}

func (store *MemSQL) GetProjectIdFromInfo(info string) (projectId int) {
	tempString := "project_id="
	index := strings.LastIndex(info, tempString)
	projectString := ""
	for index := index + len(tempString); index < len(info); index++ {
		value := info[index : index+1]
		_, err := strconv.Atoi(value)
		if err != nil {
			break
		} else {
			projectString += value
		}
	}
	projectId, _ = strconv.Atoi(projectString)
	return projectId
}

func (store *MemSQL) MonitorSlowQueries() ([]interface{}, []interface{}, error) {
	defer model.LogOnSlowExecutionWithParams(time.Now(), nil)
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

		// project name field intialized
		projectID := store.GetProjectIdFromInfo(slowQuery.Info)
		project, _ := store.GetProject(uint64(projectID))
		slowQuery.ProjectName = project.Name
		slowQuery.Info = slowQuery.Info[:U.MinInt(len(slowQuery.Info), 500)]

		if slowQuery.Info != "" {
			//segregated admin queries
			if slowQuery.User == "admin" {
				sqlAdminSlowQueries = append(sqlAdminSlowQueries, slowQuery)
			} else {
				factorsSlowQueries = append(factorsSlowQueries, slowQuery)
			}
		}
	}

	return sqlAdminSlowQueries, factorsSlowQueries, nil
}

// Disk and memory limit stats are computed at node level.
type NodeUsageStats struct {
	IPAddr                      string  `json:"ip_addr"`
	Type                        string  `json:"type"`
	State                       string  `json:"state"`
	AvailableDataDiskPercent    float64 `json:"available_data_disk_percent"`
	AvailableDataDiskMB         int64   `json:"available_data_disk_mb"`
	AvailableMemoryPercent      float64 `json:"available_memory_percent"`
	AvailableMemoryMB           int64   `json:"available_memory_mb"`
	AvailableTableMemoryPercent float64 `json:"available_table_memory_percent"`
	AvailableTableMemoryMB      int64   `json:"available_table_memory_mb"`
	Uptime                      int64   `json:"uptime"`
}

type MemSQLNodeUsageStatsWithErrors struct {
	ErrorMessage []string         `json:"errors"`
	UsageStats   []NodeUsageStats `json:"usage_stats"`
}

func (store *MemSQL) MonitorMemSQLDiskUsage() MemSQLNodeUsageStatsWithErrors {
	defer model.LogOnSlowExecutionWithParams(time.Now(), nil)
	db := C.GetServices().Db
	queryStr := "select ip_addr, type, state, available_data_disk_mb*100/total_data_disk_mb as available_data_disk_percent, available_data_disk_mb, " +
		"(max_memory_mb - memory_used_mb)*100/max_memory_mb as available_memory_percent, (max_memory_mb - memory_used_mb) as available_memory_mb, " +
		"(max_table_memory_mb - table_memory_used_mb)*100/max_table_memory_mb as available_table_memory_percent, (max_table_memory_mb - table_memory_used_mb) as available_table_memory_mb, " +
		"uptime " +
		"FROM information_schema.mv_nodes"

	rows, err := db.Raw(queryStr).Rows()
	if err != nil {
		log.WithError(err).Error("Failed to get disk usage stats")
		return MemSQLNodeUsageStatsWithErrors{ErrorMessage: []string{"Failed to run disk usage query.", err.Error()}}
	}

	nodeUsageStatsWithErrors := MemSQLNodeUsageStatsWithErrors{}
	nodeUsageStatsWithErrors.ErrorMessage = make([]string, 0, 0)
	nodeUsageStatsWithErrors.UsageStats = make([]NodeUsageStats, 0, 0)
	for rows.Next() {
		var nodeStats NodeUsageStats
		if err := db.ScanRows(rows, &nodeStats); err != nil {
			log.WithError(err).Error("Failed to scan slow queries from db.")
			continue
		}
		if nodeStats.AvailableDataDiskPercent < 20 {
			// If disk available is less than 20 percent for any node, raise an alert.
			nodeUsageStatsWithErrors.ErrorMessage = append(nodeUsageStatsWithErrors.ErrorMessage, fmt.Sprintf("Disk available '%d'MB '%f' percentage below threshold on '%s' node '%s'",
				nodeStats.AvailableDataDiskMB, nodeStats.AvailableDataDiskPercent, nodeStats.Type, nodeStats.IPAddr))
		}
		if nodeStats.AvailableMemoryPercent < 10 {
			// If memory available is less than 10 percent for any node, raise an alert.
			nodeUsageStatsWithErrors.ErrorMessage = append(nodeUsageStatsWithErrors.ErrorMessage, fmt.Sprintf("Memory available '%d'MB '%f' percentage below threshold on '%s' node '%s'",
				nodeStats.AvailableMemoryMB, nodeStats.AvailableMemoryPercent, nodeStats.Type, nodeStats.IPAddr))
		}
		if nodeStats.AvailableTableMemoryPercent < 10 {
			// If memory available for table is less than 10 percent for any node, raise an alert.
			nodeUsageStatsWithErrors.ErrorMessage = append(nodeUsageStatsWithErrors.ErrorMessage, fmt.Sprintf("Memory available '%d'MB '%f' percentage below threshold on '%s' node '%s'",
				nodeStats.AvailableTableMemoryMB, nodeStats.AvailableTableMemoryPercent, nodeStats.Type, nodeStats.IPAddr))
		}
		if nodeStats.State != "online" {
			nodeUsageStatsWithErrors.ErrorMessage = append(nodeUsageStatsWithErrors.ErrorMessage, fmt.Sprintf("Node '%s' of type '%s' not online with state '%s'",
				nodeStats.IPAddr, nodeStats.Type, nodeStats.State))
		}
		if nodeStats.Uptime <= 15*60 {
			// If uptime is les than 15 minutes which is equal to current monitoring run time, raise an alert.
			nodeUsageStatsWithErrors.ErrorMessage = append(nodeUsageStatsWithErrors.ErrorMessage, fmt.Sprintf("Node '%s' of type '%s' has been restarted before '%d'",
				nodeStats.IPAddr, nodeStats.Type, nodeStats.Uptime))
		}

		nodeUsageStatsWithErrors.UsageStats = append(nodeUsageStatsWithErrors.UsageStats, nodeStats)
	}

	return nodeUsageStatsWithErrors
}

// CollectTableSizes Captures size for major tables as metrics.
func (store *MemSQL) CollectTableSizes() map[string]string {
	defer model.LogOnSlowExecutionWithParams(time.Now(), nil)
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
	logFields := log.Fields{
		"query": query,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
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
