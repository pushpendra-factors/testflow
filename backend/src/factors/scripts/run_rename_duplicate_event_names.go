package main

import (
	C "factors/config"
	M "factors/model"
	"flag"
	"fmt"
	"net/http"
	"strconv"

	_ "github.com/jinzhu/gorm"
	"github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

type eventNameCount struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Count int64  `'json:"count"`
}

type EventNameIDCount struct {
	EventNameId uint64 `json:"event_name_id"`
	Count       int64  `json:"count"`
}
type ProjectID struct {
	ID uint64 `json:"id"`
}

func main() {
	env := flag.String("env", "development", "")
	dbHost := flag.String("db_host", "localhost", "")
	dbPort := flag.Int("db_port", 5432, "")
	dbUser := flag.String("db_user", "autometa", "")
	dbName := flag.String("db_name", "autometa", "")
	dbPass := flag.String("db_pass", "@ut0me7a", "")
	dryRun := flag.Bool("dry_run", true, "")

	flag.Parse()
	if *env != "development" &&
		*env != "staging" &&
		*env != "production" {
		err := fmt.Errorf("env [ %s ] not recognised", *env)
		panic(err)
	}

	config := &C.Configuration{
		AppName: "rename_duplicate_event_names_job",
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
	defer db.Close()
	var projectIDs []ProjectID
	err = db.Raw("SELECT DISTINCT(id) FROM projects").Find(&projectIDs).Error
	if err != nil {
		log.Fatal("Failed to get project ids from db")
	}

	for _, projectID := range projectIDs {
		log.Info("ProjectID: ", projectID)
		renameDuplicateEventNames(projectID.ID, *dryRun)
	}

}
func renameDuplicateEventNames(projectID uint64, dryRun bool) {
	db := C.GetServices().Db

	queryStr := "SELECT COUNT(*), name, type FROM event_names WHERE project_id= ? GROUP BY name, type HAVING COUNT(*)>1"
	rows, err := db.Raw(queryStr, projectID).Rows()
	if err != nil {
		log.WithError(err).Fatal("Failed to get duplicate event names")
	}
	var renamedEventNames []M.EventName

	var duplicateEventNames []eventNameCount
	for rows.Next() {
		var duplicateEventName eventNameCount
		if err := db.ScanRows(rows, &duplicateEventName); err != nil {
			log.WithError(err).Fatal("Failed scanning rows on duplicate event names job")
		}
		duplicateEventNames = append(duplicateEventNames, duplicateEventName)
	}
	for _, duplicateEventName := range duplicateEventNames {
		var eventName M.EventName
		eventName.Name = duplicateEventName.Name
		eventName.ProjectId = projectID
		eventName.Type = duplicateEventName.Type
		retEventName, errCode := M.CreateOrGetEventName(&eventName)
		if errCode != http.StatusConflict {
			log.Fatal("Failed to get event name by name, type and projectId")
		}
		queryStr = "SELECT * FROM event_names where name = ? AND project_id=? AND id != ?"
		rows, err := db.Raw(queryStr, retEventName.Name, retEventName.ProjectId, retEventName.ID).Rows()
		if err != nil {
			log.Fatal("Failed to get the event names to be renamed")
		}
		var renameEventNames []M.EventName
		var renameEventNamesID []uint64
		for rows.Next() {
			var renameEventName M.EventName
			if err := db.ScanRows(rows, &renameEventName); err != nil {
				log.WithError(err).Fatal("Failed scanning rows on duplicate event names job")
			}
			renameEventNames = append(renameEventNames, renameEventName)
			renameEventNamesID = append(renameEventNamesID, renameEventName.ID)
		}
		if dryRun == true {
			eventsQueryStr := "SELECT events.event_name_id, count(*) FROM events WHERE project_id=?" +
				" AND events.event_name_id = ANY(?) GROUP BY events.event_name_id"
			rows, err := db.Raw(eventsQueryStr, projectID, pq.Array(renameEventNamesID)).Rows()
			if err != nil {
				log.Fatal("Failed to get event_name_id data from events table")
			}
			var eventNamesIdCount []EventNameIDCount
			for rows.Next() {
				var eventNameIdCount EventNameIDCount
				if err := db.ScanRows(rows, &eventNameIdCount); err != nil {
					log.WithError(err).Fatal("Failed scanning event.event_name_id count on duplicate event names job")
				}
				eventNamesIdCount = append(eventNamesIdCount, eventNameIdCount)
			}
			log.Info("Event Name IDs to be renamed along with count: ", eventNamesIdCount)
		} else {
			for i, renameEventName := range renameEventNames {
				renamePrefix := "$renamed_$" + strconv.Itoa(i+1) + "_"
				renameEventName.Name = renamePrefix + renameEventName.Name
				updateQueryStr := "UPDATE event_names SET name = ? WHERE project_id = ? AND id = ? RETURNING event_names.*"
				var returnEventName M.EventName
				err := db.Raw(updateQueryStr, renameEventName.Name,
					renameEventName.ProjectId, renameEventName.ID).Scan(&returnEventName).Error
				if err != nil {
					log.Fatal("Failed to rename event name")
				}
				renamedEventNames = append(renamedEventNames, renameEventName)
			}
			log.Info("renamed event names : ", renamedEventNames)
			log.Info("Count of event names renamed: ", len(renamedEventNames))
		}
	}
}
