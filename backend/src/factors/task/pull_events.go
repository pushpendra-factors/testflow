package task

import (
	"encoding/json"
	"errors"
	"factors/filestore"
	P "factors/pattern"
	serviceDisk "factors/services/disk"
	"fmt"
	"os"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

const max_EVENTS = 30000000 // 30 million. (million a day)
var peLog = taskLog.WithField("prefix", "Task#PullEvents")

func pullAndWriteEventsToFile(db *gorm.DB, projectId uint64, startTime int64,
	endTime int64, eventsFilePath string) (int, string, error) {

	rows, err := db.Raw("SELECT COALESCE(users.customer_user_id, users.id), event_names.name, events.timestamp, events.count,"+
		" events.properties, users.join_timestamp, user_properties.properties FROM events "+
		"LEFT JOIN event_names ON events.event_name_id = event_names.id LEFT JOIN users ON events.user_id = users.id "+
		"LEFT JOIN user_properties ON events.user_properties_id = user_properties.id "+
		"WHERE events.project_id = ? AND events.timestamp BETWEEN  ? AND ? "+
		"ORDER BY events.user_id, events.timestamp LIMIT ?", projectId, startTime, endTime, max_EVENTS+1).Rows()
	defer rows.Close()
	if err != nil {
		peLog.WithFields(log.Fields{"err": err}).Error("SQL Query failed.")
		return 0, "", err
	}

	file, err := os.Create(eventsFilePath)
	if err != nil {
		peLog.WithFields(log.Fields{"file": eventsFilePath, "err": err}).Error("Unable to create file.")
		return 0, "", err
	}
	defer file.Close()

	var firstEvent, lastEvent *P.CounterEventFormat

	rowCount := 0
	for rows.Next() {
		var userId string
		var eventName string
		var eventTimestamp int64
		var userJoinTimestamp int64
		var eventCardinality uint
		var eventProperties postgres.Jsonb
		var userProperties postgres.Jsonb
		if err = rows.Scan(&userId, &eventName, &eventTimestamp,
			&eventCardinality, &eventProperties, &userJoinTimestamp, &userProperties); err != nil {
			peLog.WithFields(log.Fields{"err": err}).Error("SQL Parse failed.")
			return 0, "", err
		}
		eventPropertiesBytes, err := eventProperties.Value()
		if err != nil {
			peLog.WithFields(log.Fields{"err": err}).Error("Unable to unmarshal event property.")
			return 0, "", err
		}
		var eventPropertiesMap map[string]interface{}
		err = json.Unmarshal(eventPropertiesBytes.([]byte), &eventPropertiesMap)
		if err != nil {
			peLog.WithFields(log.Fields{"err": err}).Error("Unable to unmarshal event property.")
			return 0, "", err
		}
		userPropertiesBytes, err := userProperties.Value()
		if err != nil {
			peLog.WithFields(log.Fields{"err": err}).Error("Unable to unmarshal user property.")
			return 0, "", err
		}
		var userPropertiesMap map[string]interface{}
		err = json.Unmarshal(userPropertiesBytes.([]byte), &userPropertiesMap)
		if err != nil {
			peLog.WithFields(log.Fields{"err": err}).Error("Unable to unmarshal user property.")
			return 0, "", err
		}
		event := P.CounterEventFormat{
			UserId:            userId,
			UserJoinTimestamp: userJoinTimestamp,
			EventName:         eventName,
			EventTimestamp:    eventTimestamp,
			EventCardinality:  eventCardinality,
			EventProperties:   eventPropertiesMap,
			UserProperties:    userPropertiesMap,
		}

		if rowCount == 0 {
			firstEvent = &event
		}

		lineBytes, err := json.Marshal(event)
		if err != nil {
			peLog.WithFields(log.Fields{"err": err}).Error("Unable to unmarshal event.")
			return 0, "", err
		}
		line := string(lineBytes)
		if _, err := file.WriteString(fmt.Sprintf("%s\n", line)); err != nil {
			peLog.WithFields(log.Fields{"line": line, "err": err}).Error("Unable to write to file.")
			return 0, "", err
		}

		lastEvent = &event
		rowCount++
	}

	if rowCount > max_EVENTS {
		// Todo(Dinesh): notify
		return rowCount, eventsFilePath, fmt.Errorf("events count has exceeded the %d limit", max_EVENTS)
	}

	if rowCount > 0 {
		peLog.WithFields(log.Fields{
			"FirstEventTimestamp": firstEvent.EventTimestamp,
			"LastEventTimesamp":   lastEvent.EventTimestamp,
		}).Info("Events time information.")
	}
	return rowCount, eventsFilePath, nil
}

func PullEvents(db *gorm.DB, cloudManager *filestore.FileManager, diskManager *serviceDisk.DiskDriver,
	projectId uint64, startTimestamp int64, endTimestamp int64) (uint64, int, error) {

	var err error

	if projectId == 0 {
		return 0, 0, errors.New("invalid project_id")
	}
	if startTimestamp == 0 {
		return 0, 0, errors.New("invalid start timestamp")
	}
	if endTimestamp == 0 {
		return 0, 0, errors.New("invalid end timestamp")
	}

	logCtx := peLog.WithFields(log.Fields{"ProjectId": projectId,
		"StartTime": startTimestamp, "EndTime": endTimestamp})

	// Todo(Dinesh): Move modelId assignment to build task.
	// Prefix timestamp with randomAlphanumeric(5).
	modelId := uint64(time.Now().Unix())

	logCtx.Info("Pulling events.")
	// Writing events to tmp file before upload.
	fPath, fName := diskManager.GetModelEventsFilePathAndName(projectId, modelId)
	serviceDisk.MkdirAll(fPath) // create dir if not exist.
	tmpEventsFile := fPath + fName
	eventsCount, eventsFilePath, err := pullAndWriteEventsToFile(db, projectId,
		startTimestamp, endTimestamp, tmpEventsFile)
	if err != nil {
		logCtx.WithField("error", err).Error("Pull events failed. Pull and write events failed.")
		return 0, 0, err
	}

	// Zero events. Returns eventCount as 0.
	if eventsCount == 0 {
		logCtx.Info("No events found.")
		return 0, 0, nil
	}

	tmpOutputFile, err := os.Open(tmpEventsFile)
	if err != nil {
		logCtx.WithField("error", err).Error("Failed to pull events. Write to tmp failed.")
		return 0, 0, err
	}

	cDir, cName := (*cloudManager).GetModelEventsFilePathAndName(projectId, modelId)
	err = (*cloudManager).Create(cDir, cName, tmpOutputFile)
	if err != nil {
		logCtx.WithField("error", err).Error("Failed to pull events. Upload failed.")
		return 0, 0, err
	}

	logCtx.WithFields(log.Fields{
		"ModelId":        modelId,
		"EventsCount":    eventsCount,
		"EventsFilePath": eventsFilePath,
	}).Info("Successfully pulled events and written to file.")
	return modelId, eventsCount, nil
}
