package pull

import (
	"database/sql"
	"encoding/json"
	"factors/filestore"
	M "factors/model/model"
	"factors/model/store"
	P "factors/pattern"
	U "factors/util"
	"fmt"
	"io"
	"math"
	"net/http"
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
	M.USERS:          7,
}

// check if add session job is completed, then pull events(with Hubspot and Salesforce) data into cloud files with proper logging
func PullDataForEvents(projectId int64, cloudManager *filestore.FileManager, startTimestamp, endTimestamp, startTimestampInProjectTimezone, endTimestampInProjectTimezone int64, splitRangeProjectIds []int64, noOfSplits int, status map[string]interface{}, logCtx *log.Entry) (error, bool) {

	if yes, err := CheckIfAddSessionCompleted(projectId, endTimestampInProjectTimezone); !yes {
		if err != nil {
			logCtx.WithError(err).Error("checkIfAddSessionCompleted failed")
			status["events-error"] = err.Error()
			return err, false
		}
		logCtx.Error("add session job not completed")
		status["events-error"] = "add session job not completed"
		return fmt.Errorf("add session job not completed"), false
	}

	var splitRange bool = false
	if U.ContainsInt64InArray(splitRangeProjectIds, projectId) {
		splitRange = true
	}

	logCtx.Info("Pulling events.")
	_, cName := (*cloudManager).GetDailyEventArchiveFilePathAndName(projectId, 0, startTimestamp, endTimestamp)
	startAt := time.Now().UnixNano()
	eventsCount, err := pullEventsData(projectId, startTimestampInProjectTimezone, endTimestampInProjectTimezone, cName, cloudManager, startTimestamp, endTimestamp, splitRange, noOfSplits)
	if err != nil {
		logCtx.WithError(err).Error("Pull events failed. Pull and write events failed.")
		status["events-error"] = err.Error()
		return err, false
	}
	timeTakenToPullEvents := (time.Now().UnixNano() - startAt) / 1000000

	if eventsCount == 0 {
		logCtx.Info("No events found.")
		status["events-info"] = "No events found."
	} else {
		logCtx.WithFields(log.Fields{
			"EventsCount":           eventsCount,
			"TimeTakenToPullEvents": timeTakenToPullEvents,
		}).Info("Successfully pulled events and written to file.")
		status["EventsCount"] = eventsCount
	}

	return nil, true
}

// pull events(rows) from db into cloud files
func pullEventsData(projectID int64, startTimeTimezone, endTimeTimezone int64, fileName string, cloudManager *filestore.FileManager, startTimestamp, endTimestamp int64, splitRange bool, noOfSplits int) (int, error) {

	var firstEventTimestamp, lastEventTimestamp int64
	var writerMap = make(map[int64]*io.WriteCloser)
	rowCount := 0
	nilUserProperties := 0

	var startSplit = make([]int64, 0)
	var endSplit = make([]int64, 0)
	if splitRange && noOfSplits > 1 {
		var tmpStart, tmpEnd int64
		perQueryRange := U.Per_day_epoch / int64(noOfSplits)
		for i := 0; i < noOfSplits; i++ {
			if i == 0 {
				tmpStart = startTimeTimezone
			} else {
				tmpStart = tmpEnd + 1
			}
			if i == noOfSplits-1 {
				tmpEnd = endTimeTimezone
			} else {
				tmpEnd = tmpStart + perQueryRange - 1
			}
			startSplit = append(startSplit, tmpStart)
			endSplit = append(endSplit, tmpEnd)
		}
	} else {
		startSplit = []int64{startTimeTimezone}
		endSplit = []int64{endTimeTimezone}
	}
	for i := range startSplit {
		pullStartTime := startSplit[i]
		pullEndTime := endSplit[i]
		log.WithFields(log.Fields{"start": pullStartTime, "end": pullEndTime}).Infof("Executing split %d out of %d", i+1, len(startSplit))
		rows, tx, err := store.GetStore().PullEventRowsV2(projectID, pullStartTime, pullEndTime)
		if err != nil {
			log.WithError(err).Error("SQL Query failed.")
			return 0, err
		}

		for rows.Next() {
			var userID string
			var eventName string
			var eventTimestamp int64
			var userJoinTimestamp int64
			var eventCardinality uint
			var eventProperties *postgres.Jsonb
			var userProperties *postgres.Jsonb
			var is_group_user bool
			var group_1_user_id_null sql.NullString
			var group_2_user_id_null sql.NullString
			var group_3_user_id_null sql.NullString
			var group_4_user_id_null sql.NullString
			var group_5_user_id_null sql.NullString
			var group_6_user_id_null sql.NullString
			var group_7_user_id_null sql.NullString
			var group_8_user_id_null sql.NullString
			var group_1_id_null sql.NullString
			var group_2_id_null sql.NullString
			var group_3_id_null sql.NullString
			var group_4_id_null sql.NullString
			var group_5_id_null sql.NullString
			var group_6_id_null sql.NullString
			var group_7_id_null sql.NullString
			var group_8_id_null sql.NullString
			if err = rows.Scan(&userID, &eventName, &eventTimestamp, &eventCardinality, &eventProperties, &userJoinTimestamp, &userProperties,
				&is_group_user, &group_1_user_id_null, &group_2_user_id_null, &group_3_user_id_null, &group_4_user_id_null,
				&group_5_user_id_null, &group_6_user_id_null, &group_7_user_id_null, &group_8_user_id_null,
				&group_1_id_null, &group_2_id_null, &group_3_id_null, &group_4_id_null,
				&group_5_id_null, &group_6_id_null, &group_7_id_null, &group_8_id_null); err != nil {
				log.WithFields(log.Fields{"err": err}).Error("SQL Parse failed.")
				return 0, err
			}
			daysFromStart := int64(math.Floor(float64(eventTimestamp-startTimeTimezone) / float64(U.Per_day_epoch)))
			fileTimestamp := startTimestamp + daysFromStart*U.Per_day_epoch
			writer, ok := writerMap[fileTimestamp]
			if !ok {
				cPath, _ := (*cloudManager).GetDailyEventArchiveFilePathAndName(projectID, fileTimestamp, 0, 0)
				cloudWriter, err := (*cloudManager).GetWriter(cPath, fileName)
				if err != nil {
					log.WithFields(log.Fields{"file": fileName, "err": err}).Error("Unable to get cloud file writer")
					return 0, err
				}
				writerMap[fileTimestamp] = &cloudWriter
				writer = &cloudWriter
			}

			var eventPropertiesMap map[string]interface{}
			if eventProperties != nil {
				eventPropertiesBytes, err := eventProperties.Value()
				if err != nil {
					log.WithFields(log.Fields{"err": err}).Error("Unable to unmarshal event property.")
					return 0, err
				}
				err = json.Unmarshal(eventPropertiesBytes.([]byte), &eventPropertiesMap)
				if err != nil {
					log.WithFields(log.Fields{"err": err}).Error("Unable to unmarshal event property.")
					return 0, err
				}
			} else {
				log.WithFields(log.Fields{"err": err, "project_id": projectID}).Error("Nil event properties.")
			}

			var userPropertiesMap map[string]interface{}
			if userProperties != nil {
				userPropertiesBytes, err := userProperties.Value()
				if err != nil {
					log.WithFields(log.Fields{"err": err}).Error("Unable to unmarshal user property.")
					return 0, err
				}

				err = json.Unmarshal(userPropertiesBytes.([]byte), &userPropertiesMap)
				if err != nil {
					log.WithFields(log.Fields{"err": err}).Error("Unable to unmarshal user property.")
					return 0, err
				}
			} else {
				nilUserProperties++
			}

			group_1_user_id := U.IfThenElse(group_1_user_id_null.Valid, group_1_user_id_null.String, "").(string)
			group_2_user_id := U.IfThenElse(group_2_user_id_null.Valid, group_2_user_id_null.String, "").(string)
			group_3_user_id := U.IfThenElse(group_3_user_id_null.Valid, group_3_user_id_null.String, "").(string)
			group_4_user_id := U.IfThenElse(group_4_user_id_null.Valid, group_4_user_id_null.String, "").(string)
			group_5_user_id := U.IfThenElse(group_5_user_id_null.Valid, group_5_user_id_null.String, "").(string)
			group_6_user_id := U.IfThenElse(group_6_user_id_null.Valid, group_6_user_id_null.String, "").(string)
			group_7_user_id := U.IfThenElse(group_7_user_id_null.Valid, group_7_user_id_null.String, "").(string)
			group_8_user_id := U.IfThenElse(group_8_user_id_null.Valid, group_8_user_id_null.String, "").(string)

			group_1_id := U.IfThenElse(group_1_id_null.Valid, group_1_id_null.String, "").(string)
			group_2_id := U.IfThenElse(group_2_id_null.Valid, group_2_id_null.String, "").(string)
			group_3_id := U.IfThenElse(group_3_id_null.Valid, group_3_id_null.String, "").(string)
			group_4_id := U.IfThenElse(group_4_id_null.Valid, group_4_id_null.String, "").(string)
			group_5_id := U.IfThenElse(group_5_id_null.Valid, group_5_id_null.String, "").(string)
			group_6_id := U.IfThenElse(group_6_id_null.Valid, group_6_id_null.String, "").(string)
			group_7_id := U.IfThenElse(group_7_id_null.Valid, group_7_id_null.String, "").(string)
			group_8_id := U.IfThenElse(group_8_id_null.Valid, group_8_id_null.String, "").(string)

			event := P.CounterEventFormat{
				UserId:            userID,
				UserJoinTimestamp: userJoinTimestamp,
				EventName:         eventName,
				EventTimestamp:    eventTimestamp,
				EventCardinality:  eventCardinality,
				EventProperties:   eventPropertiesMap,
				UserProperties:    userPropertiesMap,
				IsGroupUser:       is_group_user,
				Group1UserId:      group_1_user_id,
				Group2UserId:      group_2_user_id,
				Group3UserId:      group_3_user_id,
				Group4UserId:      group_4_user_id,
				Group5UserId:      group_5_user_id,
				Group6UserId:      group_6_user_id,
				Group7UserId:      group_7_user_id,
				Group8UserId:      group_8_user_id,
				Group1Id:          group_1_id,
				Group2Id:          group_2_id,
				Group3Id:          group_3_id,
				Group4Id:          group_4_id,
				Group5Id:          group_5_id,
				Group6Id:          group_6_id,
				Group7Id:          group_7_id,
				Group8Id:          group_8_id,
			}

			if rowCount == 0 && i == 0 {
				firstEventTimestamp = event.EventTimestamp
			}

			lineBytes, err := json.Marshal(event)
			if err != nil {
				log.WithFields(log.Fields{"err": err}).Error("Unable to unmarshal event.")
				return 0, err
			}
			line := string(lineBytes)
			if _, err := io.WriteString(*writer, fmt.Sprintf("%s\n", line)); err != nil {
				log.WithFields(log.Fields{"line": line, "err": err}).Error("Unable to write to file.")
				return 0, err
			}

			lastEventTimestamp = event.EventTimestamp
			rowCount++
		}
		err = rows.Err()
		if err != nil {
			// Error from DB is captured eg: timeout error
			log.WithFields(log.Fields{"err": err, "project_id": projectID}).Error("Error in executing query")
			return 0, err
		}
		U.CloseReadQuery(rows, tx)
	}

	if nilUserProperties > 0 {
		log.WithFields(log.Fields{"project_id": projectID, "count": nilUserProperties}).Error("Nil user properties.")
	}

	if rowCount > M.EventsPullLimit {
		// Todo(Dinesh): notify
		return rowCount, fmt.Errorf("events count has exceeded the %d limit", M.EventsPullLimit)
	}

	for _, writer := range writerMap {
		err := (*writer).Close()
		if err != nil {
			log.WithError(err).Error("Error closing writer")
			return 0, err
		}
	}

	if rowCount > 0 {
		log.WithFields(log.Fields{
			"FirstEventTimestamp": firstEventTimestamp,
			"LastEventTimesamp":   lastEventTimestamp,
		}).Info("Events time information.")
	}
	return rowCount, nil
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

func CheckIfAddSessionCompleted(projectID int64, endTimestamp int64) (bool, error) {
	var eventsDownloadStartTimestamp int64
	{
		var errCode int
		eventsDownloadStartTimestamp, errCode = store.GetStore().GetNextSessionStartTimestampForProject(projectID)
		if errCode != http.StatusFound {
			msg := fmt.Sprintf("failed to get last min session timestamp of user for project, errCode: %v", errCode)
			log.Error(msg)
			return false, fmt.Errorf("GetNextSessionStartTimestampForProject failed, errCode: %d", errCode)
		}
	}
	if endTimestamp > eventsDownloadStartTimestamp {
		log.Errorf("jobEndTime: %d, nextSessionTimestamp: %d, add session job not completed for given range", endTimestamp, eventsDownloadStartTimestamp)
		return false, nil
	}
	return true, nil
}
