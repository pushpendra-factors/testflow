package pull

import (
	"database/sql"
	"encoding/json"
	"factors/filestore"
	M "factors/model/model"
	"factors/model/store"
	P "factors/pattern"
	serviceDisk "factors/services/disk"
	U "factors/util"
	"fmt"
	"math"
	"os"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

var FileType = map[string]int64{
	"events":         1,
	M.ADWORDS:        2,
	M.FACEBOOK:       3,
	M.BINGADS:        4,
	M.LINKEDIN:       5,
	M.GOOGLE_ORGANIC: 6,
	"users":          7,
}

// pull events(with Hubspot and Salesforce) data into files, and upload each local file to its proper cloud location
func PullEventsData(projectId int64, cloudManager *filestore.FileManager, diskManager *serviceDisk.DiskDriver, startTimestamp, endTimestamp, startTimestampInProjectTimezone, endTimestampInProjectTimezone int64, status map[string]interface{}, logCtx *log.Entry) (error, bool) {

	logCtx.Info("Pulling events.")
	// Writing events to tmp file before upload.
	_, fName := diskManager.GetDailyEventArchiveFilePathAndName(projectId, 0, startTimestamp, endTimestamp)

	startAt := time.Now().UnixNano()
	eventsCount, filePaths, err := pullEvents(projectId, startTimestampInProjectTimezone, endTimestampInProjectTimezone, fName, diskManager, startTimestamp, endTimestamp)
	if err != nil {
		logCtx.WithField("error", err).Error("Pull events failed. Pull and write events failed.")
		status["events-error"] = err.Error()
		return err, false
	}
	timeTakenToPullEvents := (time.Now().UnixNano() - startAt) / 1000000
	logCtx = logCtx.WithField("TimeTakenToPullEvents", timeTakenToPullEvents)

	// Zero events. Returns eventCount as 0.
	if eventsCount == 0 {
		logCtx.Info("No events found.")
		status["events-info"] = "No events found."
		return nil, true
	}

	_, cName := (*cloudManager).GetDailyEventArchiveFilePathAndName(projectId, 0, startTimestamp, endTimestamp)
	for timestamp, path := range filePaths {
		tmpOutputFile, err := os.Open(path)
		if err != nil {
			logCtx.WithField("error", err).Error("Failed to pull events. Write to tmp failed.")
			status["events-error"] = "Failed to pull events. Write to tmp failed."
			return err, false
		}

		cDir, _ := (*cloudManager).GetDailyEventArchiveFilePathAndName(projectId, timestamp, 0, 0)
		err = (*cloudManager).Create(cDir, cName, tmpOutputFile)
		if err != nil {
			logCtx.WithField("error", err).Error("Failed to pull events. Upload failed.")
			status["events-error"] = "Failed to pull events. Upload failed."
			return err, false
		}

		err = os.Remove(path)
		if err != nil {
			logCtx.Errorf("unable to delete file")
			return err, false
		}
		logCtx.Infof("deleted events file created:%s", path)
	}

	logCtx.WithFields(log.Fields{
		"EventsCount":           eventsCount,
		"TimeTakenToPullEvents": timeTakenToPullEvents,
	}).Info("Successfully pulled events and written to file.")

	status["EventsCount"] = eventsCount

	return nil, true
}

// pull event rows from db and generate local files w.r.t timestamp and return map with (key,value) as (timestamp,path)
func pullEvents(projectID int64, startTimeTimezone, endTimeTimezone int64, fileName string, diskManager *serviceDisk.DiskDriver, startTimestamp, endTimestamp int64) (int, map[int64]string, error) {
	rows, tx, err := store.GetStore().PullEventRowsV2(projectID, startTimeTimezone, endTimeTimezone)
	if err != nil {
		log.WithError(err).Error("SQL Query failed.")
		return 0, nil, err
	}
	defer U.CloseReadQuery(rows, tx)

	var firstEvent, lastEvent *P.CounterEventFormat

	var fileMap = make(map[int64]*os.File)
	var pathMap = make(map[int64]string)
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
		var group_1_user_id_null sql.NullString
		var group_2_user_id_null sql.NullString
		var group_3_user_id_null sql.NullString
		var group_4_user_id_null sql.NullString
		if err = rows.Scan(&userID, &eventName, &eventTimestamp, &eventCardinality, &eventProperties, &userJoinTimestamp, &userProperties,
			&group_1_user_id_null, &group_2_user_id_null, &group_3_user_id_null, &group_4_user_id_null); err != nil {
			log.WithFields(log.Fields{"err": err}).Error("SQL Parse failed.")
			return 0, nil, err
		}
		daysFromStart := int64(math.Floor(float64(eventTimestamp-startTimeTimezone) / float64(U.Per_day_epoch)))
		fileTimestamp := startTimestamp + daysFromStart*U.Per_day_epoch
		file, ok := fileMap[fileTimestamp]
		if !ok {
			fPath, _ := diskManager.GetDailyEventArchiveFilePathAndName(projectID, fileTimestamp, 0, 0)
			serviceDisk.MkdirAll(fPath) // create dir if not exist.
			tmpEventsFile := fPath + fileName
			fileNew, err := os.Create(tmpEventsFile)
			if err != nil {
				log.WithFields(log.Fields{"file": fileName, "err": err}).Error("Unable to create file.")
				return 0, nil, err
			}
			defer file.Close()
			pathMap[fileTimestamp] = tmpEventsFile
			fileMap[fileTimestamp] = fileNew
			file = fileNew
		}

		var eventPropertiesMap map[string]interface{}
		if eventProperties != nil {
			eventPropertiesBytes, err := eventProperties.Value()
			if err != nil {
				log.WithFields(log.Fields{"err": err}).Error("Unable to unmarshal event property.")
				return 0, nil, err
			}
			err = json.Unmarshal(eventPropertiesBytes.([]byte), &eventPropertiesMap)
			if err != nil {
				log.WithFields(log.Fields{"err": err}).Error("Unable to unmarshal event property.")
				return 0, nil, err
			}
		} else {
			log.WithFields(log.Fields{"err": err, "project_id": projectID}).Error("Nil event properties.")
		}

		var userPropertiesMap map[string]interface{}
		if userProperties != nil {
			userPropertiesBytes, err := userProperties.Value()
			if err != nil {
				log.WithFields(log.Fields{"err": err}).Error("Unable to unmarshal user property.")
				return 0, nil, err
			}

			err = json.Unmarshal(userPropertiesBytes.([]byte), &userPropertiesMap)
			if err != nil {
				log.WithFields(log.Fields{"err": err}).Error("Unable to unmarshal user property.")
				return 0, nil, err
			}
		} else {
			nilUserProperties++
		}

		group_1_user_id := U.IfThenElse(group_1_user_id_null.Valid, group_1_user_id_null.String, "").(string)
		group_2_user_id := U.IfThenElse(group_2_user_id_null.Valid, group_2_user_id_null.String, "").(string)
		group_3_user_id := U.IfThenElse(group_3_user_id_null.Valid, group_3_user_id_null.String, "").(string)
		group_4_user_id := U.IfThenElse(group_4_user_id_null.Valid, group_4_user_id_null.String, "").(string)

		event := P.CounterEventFormat{
			UserId:            userID,
			UserJoinTimestamp: userJoinTimestamp,
			EventName:         eventName,
			EventTimestamp:    eventTimestamp,
			EventCardinality:  eventCardinality,
			EventProperties:   eventPropertiesMap,
			UserProperties:    userPropertiesMap,
			Group1UserId:      group_1_user_id,
			Group2UserId:      group_2_user_id,
			Group3UserId:      group_3_user_id,
			Group4UserId:      group_4_user_id,
		}

		if rowCount == 0 {
			firstEvent = &event
		}

		lineBytes, err := json.Marshal(event)
		if err != nil {
			log.WithFields(log.Fields{"err": err}).Error("Unable to unmarshal event.")
			return 0, nil, err
		}
		line := string(lineBytes)
		if _, err := file.WriteString(fmt.Sprintf("%s\n", line)); err != nil {
			log.WithFields(log.Fields{"line": line, "err": err}).Error("Unable to write to file.")
			return 0, nil, err
		}

		lastEvent = &event
		rowCount++
	}
	if nilUserProperties > 0 {
		log.WithFields(log.Fields{"err": err, "project_id": projectID, "count": nilUserProperties}).Error("Nil user properties.")
	}
	err = rows.Err()
	if err != nil {
		// Error from DB is captured eg: timeout error
		log.WithFields(log.Fields{"err": err, "project_id": projectID}).Error("Error in executing query")
		return 0, nil, err
	}

	if rowCount > M.EventsPullLimit {
		// Todo(Dinesh): notify
		return rowCount, pathMap, fmt.Errorf("events count has exceeded the %d limit", M.EventsPullLimit)
	}

	if rowCount > 0 {
		log.WithFields(log.Fields{
			"FirstEventTimestamp": firstEvent.EventTimestamp,
			"LastEventTimesamp":   lastEvent.EventTimestamp,
		}).Info("Events time information.")
	}
	return rowCount, pathMap, nil
}

// pull Events for Archive
func PullEventsForArchive(projectID int64, eventsFilePath, usersFilePath string,
	startTime, endTime int64) (int, string, string, error) {

	rows, tx, err := store.GetStore().PullEventRowsForArchivalJob(projectID, startTime, endTime)
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error("SQL Query failed.")
		return 0, "", "", err
	}
	defer U.CloseReadQuery(rows, tx)

	eventsFile, err := os.Create(eventsFilePath)
	if err != nil {
		log.WithFields(log.Fields{"file": eventsFilePath, "err": err}).Error("Unable to create file.")
		return 0, "", "", err
	}
	defer eventsFile.Close()

	usersFile, err := os.Create(usersFilePath)
	if err != nil {
		log.WithFields(log.Fields{"file": usersFilePath, "err": err}).Error("Unable to create file.")
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
			log.WithFields(log.Fields{"err": err}).Error("SQL Parse failed.")
			return 0, "", "", err
		}

		var eventPropertiesMap *map[string]interface{}
		if eventProperties != nil {
			eventPropertiesMap, err = U.DecodePostgresJsonb(eventProperties)
			if err != nil {
				log.WithFields(log.Fields{"err": err}).Error("Unable to unmarshal event property.")
				return 0, "", "", err
			}
		} else {
			log.WithFields(log.Fields{"err": err, "project_id": projectID}).Error("Nil event properties.")
		}

		var eventPropertiesString, userPropertiesString []byte
		if userProperties != nil {
			var userPropertiesMap *map[string]interface{}
			userPropertiesMap, err = U.DecodePostgresJsonb(userProperties)
			if err != nil {
				log.WithFields(log.Fields{"err": err}).Error("Unable to unmarshal user property.")
				return 0, "", "", err
			}
			userPropertiesString, _ = json.Marshal(M.SanitizeUserProperties(*userPropertiesMap))
		} else {
			log.WithFields(log.Fields{"err": err, "project_id": projectID}).Error("Nil user properties.")
			userPropertiesString = []byte("")
		}

		eventPropertiesString, _ = json.Marshal(M.SanitizeEventProperties(*eventPropertiesMap))
		event := M.ArchiveEventTableFormat{
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
			log.WithFields(log.Fields{"err": err}).Error("Unable to unmarshal event.")
			return 0, "", "", err
		}
		line := string(lineBytes)
		if _, err := eventsFile.WriteString(fmt.Sprintf("%s\n", line)); err != nil {
			log.WithFields(log.Fields{"line": line, "err": err}).Error("Unable to write to file.")
			return 0, "", "", err
		}

		if _, found := userIDMap[userID]; !found && customerUserID.String != "" {
			userIDMap[userID] = customerUserID.String
		}

		rowCount++
	}

	for userID, customerUserID := range userIDMap {
		user := M.ArchiveUsersTableFormat{
			UserID:         userID,
			CustomerUserID: customerUserID,
			IngestionDate:  time.Unix(startTime, 0).UTC(),
		}
		lineBytes, err := json.Marshal(user)
		if err != nil {
			log.WithFields(log.Fields{"err": err}).Error("Unable to unmarshal user.")
			return 0, "", "", err
		}
		line := string(lineBytes)
		if _, err := usersFile.WriteString(fmt.Sprintf("%s\n", line)); err != nil {
			log.WithFields(log.Fields{"line": line, "err": err}).Error("Unable to write to file.")
			return 0, "", "", err
		}
	}

	return rowCount, eventsFilePath, usersFilePath, nil
}

// check if events file has been generated already for given start and end timestamp, in dataTimestamp's day and
// its previous day's folder
func CheckEventFileExists(cloudManager *filestore.FileManager, projectId int64, dataTimestamp, startTimestamp, endTimestamp int64) (bool, error) {
	path, name := (*cloudManager).GetDailyEventArchiveFilePathAndName(projectId, dataTimestamp, startTimestamp, endTimestamp)
	if yes, _ := U.CheckFileExists(cloudManager, path, name); yes {
		return true, nil
	}
	path, _ = (*cloudManager).GetDailyEventArchiveFilePathAndName(projectId, dataTimestamp-U.Per_day_epoch, startTimestamp, endTimestamp)
	if yes, _ := U.CheckFileExists(cloudManager, path, name); yes {
		return true, nil
	}
	return false, fmt.Errorf("file not found in cloud: %s", name)
}
