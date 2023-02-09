package main

import (
	"database/sql"
	C "factors/config"
	U "factors/util"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

type SessionFields struct {
	initialReferrer string
	source          string
	medium          string
	campaign        string
}

// NOTE: DO NOT MOVE THIS TO STORE. EXPERIMENTAION METHOD. NOT READ FOR PRODUCTION.
func getSessionEvents(
	db *gorm.DB, projectId uint64, startTime int64, endTime int64) (map[string]*SessionFields, error) {
	var sessionEvents = make(map[string]*SessionFields)
	logctx := log.WithFields(log.Fields{"projectId": projectId})

	rows, err := db.Raw("SELECT distinct(CAST (events.id AS TEXT)),"+
		"events.properties->>'$initial_referrer_domain',"+
		"events.properties->>'$source',"+
		"events.properties->>'$medium',"+
		"events.properties->>'$campaign' "+
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
		var initialReferrer, source, medium, campaign sql.NullString

		if err = rows.Scan(&sessionId, &initialReferrer, &source, &medium, &campaign); err != nil {
			logctx.Error("Error while scanning.", err)
			return sessionEvents, err
		}

		sessionEvents[sessionId] = &SessionFields{
			initialReferrer: initialReferrer.String,
			source:          source.String,
			medium:          medium.String,
			campaign:        campaign.String,
		}
	}
	log.Info(fmt.Sprintf("Total session event result %d", len(sessionEvents)))
	return sessionEvents, nil
}

// NOTE: DO NOT MOVE THIS TO STORE. EXPERIMENTAION METHOD. NOT READ FOR PRODUCTION.
func getAllEvents(db *gorm.DB, projectId uint64, projectDomain string, startTime int64, endTime int64) (
	map[string]uint, error) {
	logctx := log.WithFields(log.Fields{"projectId": projectId})
	eventsReport := make(map[string]uint)

	sessionEvents, err := getSessionEvents(db, projectId, startTime, endTime)
	if err != nil {
		log.WithError(err).Error("Error occurec in getSessionEvents")
		return eventsReport, err
	}

	rows, err := db.Raw("SELECT distinct(events.id) as event_id, events.session_id, event_names.name, "+
		"events.properties->>'$qp_utm_pageloadtype' "+
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
		var eSessionId, pageLoadType sql.NullString
		var eventName string
		var sessionId string

		if err = rows.Scan(&eventId, &eSessionId, &eventName, &pageLoadType); err != nil {
			logctx.Error("Error while scanning.", err)
			return eventsReport, err
		}

		if eventName == U.EVENT_NAME_SESSION {
			continue
		}

		if projectDomain != "" && !strings.Contains(eventName, projectDomain) {
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
			continue
		}

		// Remove trailing /
		eLen := len(eventName)
		if string(eventName[eLen-1]) == "/" {
			eventName = eventName[:eLen-1]
		}
		key := fmt.Sprintf("%s|%s|%s|%s|%s|%s",
			eventName, pageLoadType.String,
			sessionInfo.initialReferrer, sessionInfo.source,
			sessionInfo.medium, sessionInfo.campaign)
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

	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	memSQLPort := flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser := flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName := flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass := flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	memSQLCertificate := flag.String("memsql_cert", "", "")
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypeMemSQL, "Primary datastore type as memsql or postgres")

	outputFileFullPathFlag := flag.String("o_file", "pageviews.csv", "")
	projectIdFlag := flag.Uint64("project_id", 0, "Project Id.")
	projectDomainFlag := flag.String("project_domain", "", "Domain of the project")

	flag.Parse()

	if *env != "development" &&
		*env != "staging" &&
		*env != "production" {
		err := fmt.Errorf("env [ %s ] not recognised", *env)
		panic(err)
	}

	appName := "run_gen_pageview_report"
	config := &C.Configuration{
		AppName: appName,
		Env:     *env,
		MemSQLInfo: C.DBConf{
			Host:        *memSQLHost,
			Port:        *memSQLPort,
			User:        *memSQLUser,
			Name:        *memSQLName,
			Password:    *memSQLPass,
			Certificate: *memSQLCertificate,
			AppName:     appName,
		},
		PrimaryDatastore: *primaryDatastore,
	}

	C.InitConf(config)
	// Initialize configs and connections and close with defer.
	err := C.InitDB(*config)
	if err != nil {
		log.Fatal("Failed to pull events. Init failed.")
	}
	db := C.GetServices().Db
	defer db.Close()

	if *projectIdFlag <= 0 {
		log.Fatal("Failed to pull events. Invalid project_id.")
	}
	if *outputFileFullPathFlag == "" {
		log.Fatal("Invalid output file path.")
	}
	eventsReport, err := getAllEvents(
		db, *projectIdFlag, *projectDomainFlag, *startTimeFlag, *endTimeFlag)
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
		//log.Println(fmt.Sprintf("%s,%d", key, count))
		if _, err := f.WriteString(fmt.Sprintf("%s|%d\n", key, count)); err != nil {
			log.WithError(err).Fatal("Failed to write to file.")
		}
	}
}
