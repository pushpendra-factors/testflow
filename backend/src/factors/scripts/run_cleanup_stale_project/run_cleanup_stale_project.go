package main

import (
	C "factors/config"
	M "factors/model/model"
	"flag"
	"fmt"
	"strings"
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
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypeMemSQL, "Primary datastore type as memsql or postgres")

	projectId := flag.Int64("project_id", 0, "Please enter a project id")
	startTimestamp := flag.Int64("start_timestamp", 0, "Please enter a start timestamp")
	endTimestamp := flag.Int64("end_timestamp", 0, "Please enter a end timestamp")
	excludeTables := flag.String("exclude_tables", "", "")
	batch := flag.Int64("batch", 100, "Please enter a batch")
	maxRows := flag.Int64("max_rows", 0, "Max no.of rows can be deleted.")

	flag.Parse()

	config := &C.Configuration{
		Env: *env,
		DBInfo: C.DBConf{
			Host:     *dbHost,
			Port:     *dbPort,
			User:     *dbUser,
			Name:     *dbName,
			Password: *dbPass,
		},
		MemSQLInfo: C.DBConf{
			Host:        *memSQLHost,
			Port:        *memSQLPort,
			User:        *memSQLUser,
			Name:        *memSQLName,
			Password:    *memSQLPass,
			Certificate: *memSQLCertificate,
		},
		PrimaryDatastore: *primaryDatastore,
	}

	C.InitConf(config)
	err := C.InitDB(*config)
	if err != nil {
		log.Fatal("Failed to pull events. Init failed.")
	}
	if *projectId == 0 {
		log.Error("Invalid project_id given")
		return
	}

	var count int64
	db := C.GetServices().Db
	err = db.Model(&M.Project{}).Where("id = ?", projectId).Count(&count).Error
	if err != nil {
		log.Error("Failed to count the rows in project table")
		return
	}
	defer db.Close()

	if count == 0 {
		log.Error("No project found.")
		return
	}

	excludeTablesArray := strings.Split(*excludeTables, ",")
	tablesArray := [3]string{"events", "event_names", "users"}
	for i := 0; i < len(tablesArray); i++ {
		var isPresent bool = false
		for j := 0; j < len(excludeTablesArray); j++ {
			if tablesArray[i] == excludeTablesArray[j] {
				isPresent = true
				break
			}
		}
		if !isPresent {
			deleteRowFromTable(tablesArray[i], *projectId, *batch, *startTimestamp, *endTimestamp, *maxRows)
		}
	}
}

func deleteRowFromTable(tableName string, projectId, batch, startTimestamp, endTimestamp, maxRows int64) {
	logFields := log.Fields{
		"project_id":      projectId,
		"table_name":      tableName,
		"batch":           batch,
		"start_timestamp": startTimestamp,
		"end_timestamp":   endTimestamp,
	}
	logCtx := log.WithFields(logFields)

	var isRowsAffectedCountZero bool = false
	var totalCount int64

	// Adjust batch based on max rows.
	if maxRows > 0 && batch > maxRows {
		batch = maxRows
	}

	for !isRowsAffectedCountZero {
		db := C.GetServices().Db
		db.LogMode(true)

		params := make([]interface{}, 0, 0)
		stmnt := "DELETE FROM" + " " + tableName

		stmnt = stmnt + " " + "WHERE project_id = ?"
		params = append(params, projectId)

		if startTimestamp > 0 {
			startTime := time.Unix(startTimestamp, 0)
			stmnt = stmnt + " " + "AND created_at >= ?"
			params = append(params, startTime)
		}
		if endTimestamp > 0 {
			endTime := time.Unix(endTimestamp, 0)
			stmnt = stmnt + " " + "AND created_at >= ?"
			params = append(params, endTime)
		}

		stmnt = stmnt + " " + fmt.Sprintf("LIMIT %d", batch)

		rows := db.Exec(stmnt, params...)
		if rows.Error != nil {
			logCtx.WithError(rows.Error).Error("Failed to delete the row from " + tableName + " table.")
			return
		}

		isRowsAffectedCountZero = rows.RowsAffected == 0
		totalCount = totalCount + rows.RowsAffected
		logCtx.WithField("deleted_in_batch", rows.RowsAffected).
			WithField("total_deleted", totalCount).
			Info("Deleted a batch of rows")

		if totalCount > 0 && totalCount >= maxRows {
			break
		}
	}
}
