package task

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"factors/filestore"
	"factors/model/model"
	"factors/model/store"
	P "factors/pattern"
	serviceDisk "factors/services/disk"

	U "factors/util"

	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

var peLog = taskLog.WithField("prefix", "Task#PullEvents")

func pullEventsForBuildSeq(projectID int64, startTime, endTime int64, eventsFilePath string) (int, string, error) {
	rows, tx, err := store.GetStore().PullEventRowsForBuildSequenceJob(projectID, startTime, endTime)
	if err != nil {
		peLog.WithError(err).Error("SQL Query failed.")
		return 0, "", err
	}
	defer U.CloseReadQuery(rows, tx)

	file, err := os.Create(eventsFilePath)
	if err != nil {
		peLog.WithFields(log.Fields{"file": eventsFilePath, "err": err}).Error("Unable to create file.")
		return 0, "", err
	}
	defer file.Close()

	var firstEvent, lastEvent *P.CounterEventFormat

	rowCount := 0
	nilUserProperties := 0
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
			nilUserProperties++
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
	peLog.WithFields(log.Fields{"err": err, "project_id": projectID, "count": nilUserProperties}).Error("Nil user properties.")

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

func PullEventsForArchive(projectID int64, eventsFilePath, usersFilePath string,
	startTime, endTime int64) (int, string, string, error) {

	rows, tx, err := store.GetStore().PullEventRowsForArchivalJob(projectID, startTime, endTime)
	if err != nil {
		peLog.WithFields(log.Fields{"err": err}).Error("SQL Query failed.")
		return 0, "", "", err
	}
	defer U.CloseReadQuery(rows, tx)

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

		var eventPropertiesString, userPropertiesString []byte
		if userProperties != nil {
			var userPropertiesMap *map[string]interface{}
			userPropertiesMap, err = U.DecodePostgresJsonb(userProperties)
			if err != nil {
				peLog.WithFields(log.Fields{"err": err}).Error("Unable to unmarshal user property.")
				return 0, "", "", err
			}
			userPropertiesString, _ = json.Marshal(model.SanitizeUserProperties(*userPropertiesMap))
		} else {
			peLog.WithFields(log.Fields{"err": err, "project_id": projectID}).Error("Nil user properties.")
			userPropertiesString = []byte("")
		}

		eventPropertiesString, _ = json.Marshal(model.SanitizeEventProperties(*eventPropertiesMap))
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

func PullEvents(projectId int64, configs map[string]interface{}) (map[string]interface{}, bool) {

	modelType := configs["modelType"].(string)
	startTimestamp := configs["startTimestamp"].(int64)
	endTimestamp := configs["endTimestamp"].(int64)
	diskManager := configs["diskManager"].(*serviceDisk.DiskDriver)
	cloudManager := configs["cloudManager"].(*filestore.FileManager)

	status := make(map[string]interface{})
	if projectId == 0 {
		status["error"] = "invalid project_id"
		return status, false
	}
	if startTimestamp == 0 {
		status["error"] = "invalid start timestamp"
		return status, false
	}
	if endTimestamp == 0 || endTimestamp > U.TimeNowUnix() {
		status["error"] = "invalid end timestamp"
		return status, false
	}

	projectDetails, _ := store.GetStore().GetProject(projectId)
	startTimestampInProjectTimezone := startTimestamp
	endTimestampInProjectTimezone := endTimestamp
	if projectDetails.TimeZone != "" {
		// Input time is in UTC. We need the same time in the other timezone
		// if 2021-08-30 00:00:00 is UTC then we need the epoch equivalent in 2021-08-30 00:00:00 IST(project time zone)
		offset := U.FindOffsetInUTC(U.TimeZoneString(projectDetails.TimeZone))
		startTimestampInProjectTimezone = startTimestamp - int64(offset)
		endTimestampInProjectTimezone = endTimestamp - int64(offset)
	}

	logCtx := peLog.WithFields(log.Fields{"ProjectId": projectId,
		"StartTime": startTimestampInProjectTimezone, "EndTime": endTimestampInProjectTimezone})

	logCtx.Info("Pulling events.")
	// Writing events to tmp file before upload.
	fPath, fName := diskManager.GetModelEventsFilePathAndName(projectId, startTimestamp, modelType)
	serviceDisk.MkdirAll(fPath) // create dir if not exist.
	tmpEventsFile := fPath + fName
	startAt := time.Now().UnixNano()

	eventsCount, eventsFilePath, err := pullEventsForBuildSeq(projectId, startTimestampInProjectTimezone, endTimestampInProjectTimezone, tmpEventsFile)
	if err != nil {
		logCtx.WithField("error", err).Error("Pull events failed. Pull and write events failed.")
		status["error"] = err.Error()
		return status, false
	}
	timeTakenToPullEvents := (time.Now().UnixNano() - startAt) / 1000000
	logCtx = logCtx.WithField("TimeTakenToPullEvents", timeTakenToPullEvents)

	// Zero events. Returns eventCount as 0.
	if eventsCount == 0 {
		logCtx.Info("No events found.")
		status["error"] = "No events found."
		return status, true
	}

	tmpOutputFile, err := os.Open(tmpEventsFile)
	if err != nil {
		logCtx.WithField("error", err).Error("Failed to pull events. Write to tmp failed.")
		status["error"] = "Failed to pull events. Write to tmp failed."
		return status, false
	}

	cDir, cName := (*cloudManager).GetModelEventsFilePathAndName(projectId, startTimestamp, modelType)
	err = (*cloudManager).Create(cDir, cName, tmpOutputFile)
	if err != nil {
		logCtx.WithField("error", err).Error("Failed to pull events. Upload failed.")
		status["error"] = "Failed to pull events. Upload failed."
		return status, false
	}

	logCtx.WithFields(log.Fields{
		"EventsCount":           eventsCount,
		"EventsFilePath":        eventsFilePath,
		"TimeTakenToPullEvents": timeTakenToPullEvents,
	}).Info("Successfully pulled events and written to file.")

	status["EventsCount"] = eventsCount
	status["EventsFilePath"] = eventsFilePath
	status["TimeTakenToPullEvents"] = timeTakenToPullEvents
	return status, true
}
