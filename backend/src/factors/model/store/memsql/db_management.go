package memsql

import (
	C "factors/config"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
)

func getNonAnalyzedTablesInAnInterval(intervalInMinutes int) ([]string, int) {
	tableNames := make([]string, 0, 0)

	query := "SELECT distinct(table_name) FROM information_schema.OPTIMIZER_STATISTICS" + " " +
		"WHERE database_name='factors' AND" + " " +
		fmt.Sprintf("last_updated < NOW() - INTERVAL %d MINUTE", intervalInMinutes)

	db := C.GetServices().Db
	rows, err := db.Raw(query).Rows()
	if err != nil {
		log.WithError(err).Error("Failed to get non-analyzed tables")
		return tableNames, http.StatusInternalServerError
	}

	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			log.WithError(err).
				Error("Failed to scan row on getNonAnalyzedTablesInAnInterval.")
			continue
		}

		tableNames = append(tableNames, tableName)
	}

	return tableNames, http.StatusFound
}

func analyzeTables(tables []string) (int, []string) {
	failedTables := make([]string, 0, 0)
	for i := range tables {
		table := tables[i]
		query := fmt.Sprintf("ANALYZE TABLE %s", table)

		db := C.GetServices().Db
		err := db.Exec(query).Error
		if err != nil {
			log.WithField("table_name", table).WithError(err).
				Error("Failed to execute analyze.")

			failedTables = append(failedTables, table)
			continue
		}

		// Log to verify runs.
		log.Info(query)
	}

	if len(failedTables) > 0 {
		return http.StatusInternalServerError, failedTables
	}

	return http.StatusOK, failedTables
}

func AnalyzeTableInAnInterval(intervalInMinutes int) (int, []string) {
	if intervalInMinutes <= 0 {
		log.Error("Invalid interval")
		return http.StatusInternalServerError, []string{}
	}

	tables, status := getNonAnalyzedTablesInAnInterval(intervalInMinutes)
	if status != http.StatusFound {
		return http.StatusInternalServerError, []string{}
	}

	return analyzeTables(tables)
}
