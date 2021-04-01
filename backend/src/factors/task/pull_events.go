package task

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"time"

	"factors/filestore"
	"factors/model/model"
	"factors/model/store"
	P "factors/pattern"
	serviceDisk "factors/services/disk"

	U "factors/util"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

var peLog = taskLog.WithField("prefix", "Task#PullEvents")

func pullEventsForBuildSeq(projectID uint64, startTime, endTime int64, eventsFilePath string) (int, string, error) {
	rows, err := store.GetStore().PullEventRowsForBuildSequenceJob(projectID, startTime, endTime)
	defer rows.Close()
	if err != nil {
		peLog.WithError(err).Error("SQL Query failed.")
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
		var userID string
		var eventName string
		var eventTimestamp int64
		var userJoinTimestamp int64
		var eventCardinality uint
		var eventProperties *postgres.Jsonb
		var userProperties *postgres.Jsonb
		if err = rows.Scan(&userID, &eventName, &eventTimestamp,
			&eventCardinality, &eventProperties, &userJoinTimestamp, &userProperties); err != nil {
			peLog.WithFields(log.Fields{"err": err}).Error("SQL Parse failed.")
			return 0, "", err
		}

		var eventPropertiesMap map[string]interface{}
		if eventProperties != nil {
			eventPropertiesBytes, err := eventProperties.Value()
			if err != nil {
				peLog.WithFields(log.Fields{"err": err}).Error("Unable to unmarshal event property.")
				return 0, "", err
			}
			err = json.Unmarshal(eventPropertiesBytes.([]byte), &eventPropertiesMap)
			if err != nil {
				peLog.WithFields(log.Fields{"err": err}).Error("Unable to unmarshal event property.")
				return 0, "", err
			}
		} else {
			peLog.WithFields(log.Fields{"err": err, "project_id": projectID}).Error("Nil event properties.")
		}

		var userPropertiesMap map[string]interface{}
		if userProperties != nil {
			userPropertiesBytes, err := userProperties.Value()
			if err != nil {
				peLog.WithFields(log.Fields{"err": err}).Error("Unable to unmarshal user property.")
				return 0, "", err
			}

			err = json.Unmarshal(userPropertiesBytes.([]byte), &userPropertiesMap)
			if err != nil {
				peLog.WithFields(log.Fields{"err": err}).Error("Unable to unmarshal user property.")
				return 0, "", err
			}
		} else {
			peLog.WithFields(log.Fields{"err": err, "project_id": projectID}).Error("Nil user properties.")
		}

		event := P.CounterEventFormat{
			UserId:            userID,
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

	if rowCount > model.EventsPullLimit {
		// Todo(Dinesh): notify
		return rowCount, eventsFilePath, fmt.Errorf("events count has exceeded the %d limit", model.EventsPullLimit)
	}

	if rowCount > 0 {
		peLog.WithFields(log.Fields{
			"FirstEventTimestamp": firstEvent.EventTimestamp,
			"LastEventTimesamp":   lastEvent.EventTimestamp,
		}).Info("Events time information.")
	}
	return rowCount, eventsFilePath, nil
}

func PullEventsForArchive(projectID uint64, eventsFilePath, usersFilePath string,
	startTime, endTime int64) (int, string, string, error) {

	rows, err := store.GetStore().PullEventRowsForArchivalJob(projectID, startTime, endTime)
	defer rows.Close()
	if err != nil {
		peLog.WithFields(log.Fields{"err": err}).Error("SQL Query failed.")
		return 0, "", "", err
	}

	eventsFile, err := os.Create(eventsFilePath)
	if err != nil {
		peLog.WithFields(log.Fields{"file": eventsFilePath, "err": err}).Error("Unable to create file.")
		return 0, "", "", err
	}
	defer eventsFile.Close()

	usersFile, err := os.Create(usersFilePath)
	if err != nil {
		peLog.WithFields(log.Fields{"file": usersFilePath, "err": err}).Error("Unable to create file.")
		return 0, "", "", err
	}
	defer usersFile.Close()
	userIDMap := make(map[string]string) // userID to identifiedUserID map.

	rowCount := 0
	for rows.Next() {
		var eventID string
		var userID string
		var customerUserID sql.NullString
		var eventName string
		var eventTimestamp int64
		var userJoinTimestamp int64
		var sessionID sql.NullString
		var eventProperties *postgres.Jsonb
		var userProperties *postgres.Jsonb
		if err = rows.Scan(&eventID, &userID, &customerUserID, &eventName, &eventTimestamp,
			&sessionID, &eventProperties, &userJoinTimestamp, &userProperties); err != nil {
			peLog.WithFields(log.Fields{"err": err}).Error("SQL Parse failed.")
			return 0, "", "", err
		}

		var eventPropertiesMap *map[string]interface{}
		if eventProperties != nil {
			eventPropertiesMap, err = U.DecodePostgresJsonb(eventProperties)
			if err != nil {
				peLog.WithFields(log.Fields{"err": err}).Error("Unable to unmarshal event property.")
				return 0, "", "", err
			}
		} else {
			peLog.WithFields(log.Fields{"err": err, "project_id": projectID}).Error("Nil event properties.")
		}

		var userPropertiesMap *map[string]interface{}
		if userProperties != nil {
			userPropertiesMap, err = U.DecodePostgresJsonb(userProperties)
			if err != nil {
				peLog.WithFields(log.Fields{"err": err}).Error("Unable to unmarshal user property.")
				return 0, "", "", err
			}
		} else {
			peLog.WithFields(log.Fields{"err": err, "project_id": projectID}).Error("Nil user properties.")
		}

		eventPropertiesString, _ := json.Marshal(model.SanitizeEventProperties(*eventPropertiesMap))
		userPropertiesString, _ := json.Marshal(model.SanitizeUserProperties(*userPropertiesMap))
		event := model.ArchiveEventTableFormat{
			EventID:           eventID,
			UserID:            userID,
			UserJoinTimestamp: userJoinTimestamp,
			EventName:         eventName,
			EventTimestamp:    time.Unix(eventTimestamp, 0).UTC(),
			SessionID:         sessionID.String,
			EventProperties:   string(eventPropertiesString[:]),
			UserProperties:    string(userPropertiesString[:]),
		}

		lineBytes, err := json.Marshal(event)
		if err != nil {
			peLog.WithFields(log.Fields{"err": err}).Error("Unable to unmarshal event.")
			return 0, "", "", err
		}
		line := string(lineBytes)
		if _, err := eventsFile.WriteString(fmt.Sprintf("%s\n", line)); err != nil {
			peLog.WithFields(log.Fields{"line": line, "err": err}).Error("Unable to write to file.")
			return 0, "", "", err
		}

		if _, found := userIDMap[userID]; !found && customerUserID.String != "" {
			userIDMap[userID] = customerUserID.String
		}

		rowCount++
	}

	for userID, customerUserID := range userIDMap {
		user := model.ArchiveUsersTableFormat{
			UserID:         userID,
			CustomerUserID: customerUserID,
			IngestionDate:  time.Unix(startTime, 0).UTC(),
		}
		lineBytes, err := json.Marshal(user)
		if err != nil {
			peLog.WithFields(log.Fields{"err": err}).Error("Unable to unmarshal user.")
			return 0, "", "", err
		}
		line := string(lineBytes)
		if _, err := usersFile.WriteString(fmt.Sprintf("%s\n", line)); err != nil {
			peLog.WithFields(log.Fields{"line": line, "err": err}).Error("Unable to write to file.")
			return 0, "", "", err
		}
	}

	return rowCount, eventsFilePath, usersFilePath, nil
}

func PullEvents(db *gorm.DB, cloudManager *filestore.FileManager,
	diskManager *serviceDisk.DiskDriver, projectId uint64, startTimestamp int64,
	endTimestamp int64) (uint64, int, error) {

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
	curTimeInMilliSecs := time.Now().UnixNano() / 1000000
	// modelId = time in millisecs + random number upto 3 digits.
	modelId := uint64(curTimeInMilliSecs + rand.Int63n(999))

	logCtx.Info("Pulling events.")
	// Writing events to tmp file before upload.
	fPath, fName := diskManager.GetModelEventsFilePathAndName(projectId, modelId)
	serviceDisk.MkdirAll(fPath) // create dir if not exist.
	tmpEventsFile := fPath + fName
	eventsCount, eventsFilePath, err := pullEventsForBuildSeq(projectId, startTimestamp, endTimestamp, tmpEventsFile)
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
