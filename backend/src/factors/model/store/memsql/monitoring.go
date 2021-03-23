package memsql

import (
	"factors/metrics"
)

type SlowQueries struct {
	Runtime                  float64 `json:"runtime"`
	Query                    string  `json:"query"`
	Pid                      int64   `json:"pid"`
	Usename                  string  `json:"usename"`
	ApplicationName          string  `json:"application_name"`
	VacuumProgressPercentage float64 `json:"vacuum_progress_percentage,omitempty"`
	VacuumPhase              string  `json:"vacuum_phase,omitempty"`
}

func (store *MemSQL) RunMonitoringQuery() ([]SlowQueries, []SlowQueries, error) {
	sqlAdminSlowQueries := make([]SlowQueries, 0, 0)
	factorsSlowQueries := make([]SlowQueries, 0, 0)

	// TODO(prateek): Check on monitoring queries. Below can be a starting point.
	// https://stackoverflow.com/questions/46948271/how-to-view-all-queries-in-memsql-or-mysql - show full processlist;
	// https://docs.singlestore.com/v7.0/concepts/workload-profiling/workload-profiling/

	return sqlAdminSlowQueries, factorsSlowQueries, nil
}

// collectTableSizes Captures size for major tables as metrics.
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
	// TODO(prateek): Check for MemSQL.
	// https://www.singlestore.com/forum/t/are-there-any-methods-to-calculate-the-size-of-a-specific-table-in-memsql/1292/3
	return tableSizes
}
