package main

import (
	C "factors/config"
	"factors/model/model"
	"factors/model/store"
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
	env := flag.String("env", C.DEVELOPMENT, "")

	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	memSQLPort := flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser := flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName := flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass := flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypeMemSQL, "Primary datastore type as memsql or postgres")

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
		MemSQLInfo: C.DBConf{
			Host:     *memSQLHost,
			Port:     *memSQLPort,
			User:     *memSQLUser,
			Name:     *memSQLName,
			Password: *memSQLPass,
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

	projects, errCode := store.GetStore().GetProjects()
	if errCode != http.StatusFound {
		log.Fatal("Failed to get projects.")
	}

	renameStat := make(map[uint64]string)
	for _, project := range projects {
		log.Info("Running on project_id ", project.ID)
		renameDuplicateEventNames(project.ID, *dryRun, &renameStat)
	}

	if *dryRun {
		for k, v := range renameStat {
			fmt.Printf("project: %d -> duplicate event_names: %+v \n", k, v)
		}
	}

}

// NOTE: DO NOT MOVE THIS TO STORE AS THIS CANNOT BE USED AS PRODUCTION CODE. ONLY FOR MIGRATION ON POSTGRES.
func renameDuplicateEventNames(projectID uint64, dryRun bool, renameStat *map[uint64]string) {
	db := C.GetServices().Db

	queryStr := "SELECT COUNT(*), name, type FROM event_names WHERE project_id = ? and type != 'FE' GROUP BY name, type HAVING COUNT(*)>1"
	rows, err := db.Raw(queryStr, projectID).Rows()
	if err != nil {
		log.WithError(err).Fatal("Failed to get duplicate event names")
	}
	var renamedEventNames []model.EventName

	var duplicateEventNames []eventNameCount
	for rows.Next() {
		var duplicateEventName eventNameCount
		if err := db.ScanRows(rows, &duplicateEventName); err != nil {
			log.WithError(err).Fatal("Failed scanning rows on duplicate event names job")
		}
		duplicateEventNames = append(duplicateEventNames, duplicateEventName)
	}

	for _, duplicateEventName := range duplicateEventNames {
		var eventName model.EventName
		eventName.Name = duplicateEventName.Name
		eventName.ProjectId = projectID
		eventName.Type = duplicateEventName.Type
		retEventName, errCode := store.GetStore().CreateOrGetEventName(&eventName)
		if errCode != http.StatusConflict {
			log.Fatal("Failed to get event name by name, type and projectId")
		}

		queryStr = "SELECT * FROM event_names where name = ? AND project_id=? AND id != ?"
		rows, err := db.Raw(queryStr, retEventName.Name, retEventName.ProjectId, retEventName.ID).Rows()
		if err != nil {
			log.Fatal("Failed to get the event names to be renamed")
		}
		var renameEventNames []model.EventName
		var renameEventNamesID []uint64
		for rows.Next() {
			var renameEventName model.EventName
			if err := db.ScanRows(rows, &renameEventName); err != nil {
				log.WithError(err).Fatal("Failed scanning rows on duplicate event names job")
			}
			renameEventNames = append(renameEventNames, renameEventName)
			renameEventNamesID = append(renameEventNamesID, renameEventName.ID)
		}
		defer rows.Close()

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
			defer rows.Close()

			if len(eventNamesIdCount) > 0 {
				if _, exists := (*renameStat)[projectID]; !exists {
					(*renameStat)[projectID] = fmt.Sprintf("%+v", eventNamesIdCount)
				} else {
					(*renameStat)[projectID] = fmt.Sprintf("%s, %+v", (*renameStat)[projectID], eventNamesIdCount)
				}

			}
			continue
		}

		for i, renameEventName := range renameEventNames {
			renamePrefix := "$renamed_$" + strconv.Itoa(i+1) + "_"
			renameEventName.Name = renamePrefix + renameEventName.Name
			updateQueryStr := "UPDATE event_names SET name = ? WHERE project_id = ? AND id = ? RETURNING event_names.*"
			var returnEventName model.EventName
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
