package main

import (
	"flag"
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/jinzhu/gorm/dialects/postgres"

	U "factors/util"
)

var memSQLDB *gorm.DB

// Script for testing the bulk download performance of MemSQL.
func main() {
	env := flag.String("env", "development", "")

	memSQLDSN := flag.String(
		"memsql_dsn",
		"admin:N4hQ2gPolt@tcp(svc-bc37ae65-0bf2-4b7b-add4-401132647f90-ddl.gcp-virginia-1.db.memsql.com:3306)/factors?charset=utf8mb4&parseTime=True&loc=Local",
		"",
	)
	projectID := flag.Uint64("project_id", 0, "")
	startTimestamp := flag.Int64("start_timestamp", 0, "")
	endTimestamp := flag.Int64("end_timestamp", 0, "")
	flag.Parse()

	if *projectID == 0 || *startTimestamp == 0 || *endTimestamp == 0 {
		log.WithFields(log.Fields{
			"start_timestap": *startTimestamp,
			"end_timestamp":  *endTimestamp,
			"project_id":     *projectID,
		}).Fatal("Invalid flags.")
	}

	initMemSQLDB(*env, *memSQLDSN)
	pullEventsFlat(memSQLDB, *projectID, *startTimestamp, *endTimestamp)
}

func initMemSQLDB(env, dsn string) {
	var err error
	// dsn sample admin:LpAHQyAMyI@tcp(svc-2b9e36ee-d5d0-4082-9779-2027e39fcbab-ddl.gcp-virginia-1.db.memsql.com:3306)/factors?charset=utf8mb4&parseTime=True&loc=Local
	memSQLDB, err = gorm.Open("mysql", dsn)
	if err != nil {
		log.WithError(err).Fatal("Failed connecting to memsql.")
	}

	if env == "development" {
		memSQLDB.LogMode(true)
	} else {
		memSQLDB.LogMode(false)
	}
}

func pullEventsFlat(db *gorm.DB, projectID uint64, startTime, endTime int64) {

	queryExecStartTime := U.TimeNowUnix()
	rawQuery := fmt.Sprintf("SELECT COALESCE(users.customer_user_id, users.id), event_names.name, events.timestamp, events.count,"+
		" events.properties, users.join_timestamp, events.user_properties FROM events "+
		"LEFT JOIN event_names ON events.event_name_id=event_names.id LEFT JOIN users ON events.user_id = users.id "+
		"WHERE events.project_id = %d AND events.timestamp BETWEEN  %d AND %d "+
		"ORDER BY COALESCE(users.customer_user_id, users.id), events.timestamp", projectID, startTime, endTime)

	rows, err := db.Raw(rawQuery).Rows()
	if err != nil {
		log.WithField("error", err).Fatal("Failed to read rows.")
	}
	defer rows.Close()
	queryExecStopTime := U.TimeNowUnix()

	rowCount := 0
	downloadStartTime := U.TimeNowUnix()
	for rows.Next() {
		var userID string
		var eventName string
		var eventTimestamp int64
		var userJoinTimestamp int64
		var eventCardinality uint
		var eventProperties *postgres.Jsonb
		var userProperties *postgres.Jsonb
		if err = rows.Scan(&userID, &eventName, &eventTimestamp, &eventCardinality,
			&eventProperties, &userJoinTimestamp, &userProperties); err != nil {
			log.WithFields(log.Fields{"err": err}).Error("SQL Parse failed.")
			return
		}
		rowCount++
	}

	if rows.Err() != nil {
		log.WithField("error", rows.Err()).Fatal("Failed on rows scanner after read.")
	}

	log.WithField("rows", rowCount).
		WithField("exec_time_taken_in_secs", queryExecStopTime-queryExecStartTime).
		WithField("read_time_taken_in_secs", U.TimeNowUnix()-downloadStartTime).
		Info("Successfully downloaded.")
}
