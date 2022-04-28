package postgres

import (
	C "factors/config"
	"factors/metrics"
	"factors/util"
	U "factors/util"
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
)

type SlowQueries struct {
	Runtime                  float64 `json:"runtime"`
	RuntimeString            string  `json:"runtime_string"`
	Query                    string  `json:"query"`
	Pid                      int64   `json:"pid"`
	Usename                  string  `json:"usename"`
	ApplicationName          string  `json:"application_name"`
	VacuumProgressPercentage float64 `json:"vacuum_progress_percentage,omitempty"`
	VacuumPhase              string  `json:"vacuum_phase,omitempty"`
}

func (pg *Postgres) MonitorSlowQueries() ([]interface{}, []interface{}, error) {
	sqlAdminSlowQueries := make([]interface{}, 0, 0)
	factorsSlowQueries := make([]interface{}, 0, 0)

	queryStr := "SELECT EXTRACT(epoch from (now() - query_start)) as runtime,query, pg_stat_activity.pid, usename, application_name," + " " +
		"CASE WHEN heap_blks_vacuumed > 0 THEN (heap_blks_vacuumed::FLOAT/heap_blks_total::FLOAT) * 100 ELSE 0 END vacuum_progress_percentage, phase as vacuum_phase" + " " +
		"FROM  pg_stat_activity LEFT JOIN pg_stat_progress_vacuum ON pg_stat_progress_vacuum.pid=pg_stat_activity.pid" + " " +
		"WHERE EXTRACT(epoch from (now() - query_start)) > 120 AND state = 'active' AND query NOT ILIKE '%pg_stat_activity%' ORDER BY runtime DESC;"

	db := C.GetServices().Db
	rows, err := db.Raw(queryStr).Rows()
	if err != nil {
		log.WithError(err).Error("Failed to get slow queries from pg_stat_activity")
		return sqlAdminSlowQueries, factorsSlowQueries, err
	}

	for rows.Next() {
		var slowQuery SlowQueries
		if err := db.ScanRows(rows, &slowQuery); err != nil {
			log.WithError(err).Error("Failed to scan slow queries from db.")
			return sqlAdminSlowQueries, factorsSlowQueries, err
		}
		slowQuery.RuntimeString = U.SecondsToHMSString(int64(slowQuery.Runtime))
		if slowQuery.Query != "" {
			if slowQuery.Usename == "cloudsqladmin" {
				sqlAdminSlowQueries = append(sqlAdminSlowQueries, slowQuery)
			} else {
				factorsSlowQueries = append(factorsSlowQueries, slowQuery)
			}
		}
	}

	return sqlAdminSlowQueries, factorsSlowQueries, nil
}

// collectTableSizes Captures size for major tables as metrics.
func (pg *Postgres) CollectTableSizes() map[string]string {
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
