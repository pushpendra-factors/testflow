package main

import (
	"database/sql"
	C "factors/config"
	U "factors/util"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

type SessionFields struct {
	initialReferrer string
	//source       string
	//medium        string
	//campaign        string
}

func getSessionEvents(
	db *gorm.DB, projectId uint64, startTime int64, endTime int64) (map[string]*SessionFields, error) {
	var sessionEvents = make(map[string]*SessionFields)
	logctx := log.WithFields(log.Fields{"projectId": projectId})

	rows, err := db.Raw("SELECT distinct(CAST (events.id AS TEXT)) as session_id,"+
		"events.properties->>'$initial_referrer_domain' "+
		//"events.properties->>'$source',"+
		//"events.properties->>'$medium',"+
		//"events.properties->>'$campaign' "+
		"FROM events "+
		"WHERE events.project_id = ?  AND events.timestamp >= ? AND events.timestamp <= ? "+
		"AND events.event_name_id IN (SELECT id FROM event_names WHERE project_id = ? AND name = '$session')",
		projectId, startTime, endTime, projectId).Rows()
	defer rows.Close()
	if err != nil {
		logctx.WithFields(log.Fields{"err": err}).Error("SQL Query failed.")
		return sessionEvents, err
	}

	for rows.Next() {
		var sessionId string
		var initialReferrer sql.NullString
		//var source, medium, campaign sql.NullString

		//if err = rows.Scan(&sessionId, &initialReferrer, &source, &medium, &campaign); err != nil {
		if err = rows.Scan(&sessionId, &initialReferrer); err != nil {
			logctx.Error("Error while scanning.", err)
			return sessionEvents, err
		}

		sessionEvents[sessionId] = &SessionFields{
			initialReferrer: initialReferrer.String,
			//source:       source.String,
			//medium:        medium.String,
			//campaign:      campaign.String,
		}
	}
	log.Info(fmt.Sprintf("Total session event result %d", len(sessionEvents)))
	return sessionEvents, nil
}

func getAllEvents(db *gorm.DB, projectId uint64, startTime int64, endTime int64) (
	map[string]uint, error) {
	logctx := log.WithFields(log.Fields{"projectId": projectId})
	eventsReport := make(map[string]uint)

	sessionEvents, err := getSessionEvents(db, projectId, startTime, endTime)
	if err != nil {
		log.WithError(err).Error("Error occurec in getSessionEvents")
		return eventsReport, err
	}

	rows, err := db.Raw("SELECT distinct(events.id) as event_id, events.session_id, event_names.name "+
		"FROM events "+
		"LEFT JOIN event_names ON events.event_name_id = event_names.id "+
		"WHERE events.project_id = ? AND events.timestamp >= ? AND events.timestamp <= ?",
		projectId, startTime, endTime).Rows()
	defer rows.Close()
	if err != nil {
		logctx.WithFields(log.Fields{"err": err}).Error("SQL Query failed.")
		return eventsReport, err
	}

	for rows.Next() {
		var eventId string
		var eSessionId sql.NullString
		var eventName string
		var sessionId string

		if err = rows.Scan(&eventId, &eSessionId, &eventName); err != nil {
			logctx.Error("Error while scanning.", err)
			return eventsReport, err
		}

		if eventName == U.EVENT_NAME_SESSION {
			continue
		}

		if !eSessionId.Valid {
			log.Error(fmt.Sprintf(
				"Null sessionId %s encountered for event:%s eventId:%s",
				eSessionId.String, eventName, eventId))
			continue
		}

		sessionId = eSessionId.String
		var sessionInfo *SessionFields
		var ok bool
		if sessionInfo, ok = sessionEvents[sessionId]; !ok {
			log.Error(fmt.Sprintf("Missing sessionId %s", sessionId))
		}

		// Remove trailing /
		eLen := len(eventName)
		if string(eventName[eLen-1]) == "/" {
			eventName = eventName[:eLen-1]
		}
		//key := fmt.Sprintf("%s,%s,%s,%s,%s",
		//	eventName, sessionInfo.initialReferrer, sessionInfo.source,
		// sessionInfo.medium, sessionInfo.campaign)
		key := fmt.Sprintf("%s,%s",
			eventName, sessionInfo.initialReferrer)
		if ec, ok := eventsReport[key]; !ok {
			eventsReport[key] = 1
		} else {
			eventsReport[key] = ec + 1
		}
	}
	return eventsReport, err
}

// go run_gen_pageview_report.go --project_id=<id> --start_time=<> --end_time=<>
func main() {
	env := flag.String("env", "development", "")
	startTimeFlag := flag.Int64(
		"start_time", 0, "Pull events, interval start timestamp. Format is unix timestamp.")
	endTimeFlag := flag.Int64(
		"end_time", time.Now().Unix(), "Pull events, interval end timestamp. defaults to current timestamp. Format is unix timestamp.")
	dbHost := flag.String("db_host", "localhost", "")
	dbPort := flag.Int("db_port", 5432, "")
	dbUser := flag.String("db_user", "autometa", "")
	dbName := flag.String("db_name", "autometa", "")
	dbPass := flag.String("db_pass", "@ut0me7a", "")
	outputFileFullPathFlag := flag.String("o_file", "pageviews.csv", "")
	projectIdFlag := flag.Uint64("project_id", 0, "Project Id.")

	flag.Parse()

	if *env != "development" &&
		*env != "staging" &&
		*env != "production" {
		err := fmt.Errorf("env [ %s ] not recognised", *env)
		panic(err)
	}

	config := &C.Configuration{
		AppName: "run_gen_pageview_report",
		Env:     *env,
		DBInfo: C.DBConf{
			Host:     *dbHost,
			Port:     *dbPort,
			User:     *dbUser,
			Name:     *dbName,
			Password: *dbPass,
		},
	}

	C.InitConf(config.Env)
	// Initialize configs and connections and close with defer.
	err := C.InitDB(config.DBInfo)
	if err != nil {
		log.Fatal("Failed to pull events. Init failed.")
	}
	db := C.GetServices().Db

	if *projectIdFlag <= 0 {
		log.Fatal("Failed to pull events. Invalid project_id.")
	}
	if *outputFileFullPathFlag == "" {
		log.Fatal("Invalid output file path.")
	}
	eventsReport, err := getAllEvents(db, *projectIdFlag, *startTimeFlag, *endTimeFlag)
	if err != nil {
		log.WithError(err).Error("failed to getAllEvents.")
		return
	}

	f, err := os.Create(*outputFileFullPathFlag)
	if err != nil {
		log.WithError(err).Fatal("failed to create file.")
		return
	}
	defer f.Close()

	for key, count := range eventsReport {
		log.Println(fmt.Sprintf("%s,%d", key, count))
		if _, err := f.WriteString(fmt.Sprintf("%s,%d\n", key, count)); err != nil {
			log.WithError(err).Fatal("Failed to write to file.")
		}
	}
}
