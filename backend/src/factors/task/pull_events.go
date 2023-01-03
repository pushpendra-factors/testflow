package task

import (
	"database/sql"
	"encoding/json"
	"errors"
	"factors/filestore"
	M "factors/model/model"
	"factors/model/store"
	P "factors/pattern"
	serviceDisk "factors/services/disk"
	U "factors/util"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

type CounterCampaignFormat struct {
	Id         string                 `json:"id"`
	Channel    string                 `json:"source"`
	Doctype    int                    `json:"type"`
	Timestamp  int64                  `json:"timestamp"`
	Value      map[string]interface{} `json:"value"`
	SmartProps map[string]interface{} `json:"sp"`
}

type CounterUserFormat struct {
	Id            string                 `json:"id"`
	Properties    map[string]interface{} `json:"pr"`
	Is_Anonymous  bool                   `json:"ia"`
	JoinTimestamp int64                  `json:"ts"`
}

var fileType = map[string]int64{
	"events":         1,
	M.ADWORDS:        2,
	M.FACEBOOK:       3,
	M.BINGADS:        4,
	M.LINKEDIN:       5,
	M.GOOGLE_ORGANIC: 6,
	"users":          7,
}

var channelToPullMap = map[string]func(int64, int64, int64) (*sql.Rows, *sql.Tx, error){
	M.ADWORDS:        store.GetStore().PullAdwordsRows,
	M.FACEBOOK:       store.GetStore().PullFacebookRows,
	M.BINGADS:        store.GetStore().PullBingAdsRows,
	M.LINKEDIN:       store.GetStore().PullLinkedInRows,
	M.GOOGLE_ORGANIC: store.GetStore().PullGoogleOrganicRows,
}

var peLog = taskLog.WithField("prefix", "Task#PullEvents")

// pull Events (with Hubspot and Salesforce)
func pullEvents(projectID int64, startTime, endTime int64, eventsFilePath string) (int, string, error) {
	rows, tx, err := store.GetStore().PullEventRows(projectID, startTime, endTime)
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
	if nilUserProperties > 0 {
		peLog.WithFields(log.Fields{"err": err, "project_id": projectID, "count": nilUserProperties}).Error("Nil user properties.")
	}
	err = rows.Err()
	if err != nil {
		// Error from DB is captured eg: timeout error
		peLog.WithFields(log.Fields{"err": err, "project_id": projectID}).Error("Error in executing query")
		return 0, "", err
	}

	if rowCount > M.EventsPullLimit {
		// Todo(Dinesh): notify
		return rowCount, eventsFilePath, fmt.Errorf("events count has exceeded the %d limit", M.EventsPullLimit)
	}

	if rowCount > 0 {
		peLog.WithFields(log.Fields{
			"FirstEventTimestamp": firstEvent.EventTimestamp,
			"LastEventTimesamp":   lastEvent.EventTimestamp,
		}).Info("Events time information.")
	}
	return rowCount, eventsFilePath, nil
}

// pull Events (with Hubspot and Salesforce)
func pullEventsDaily(projectID int64, startTime, endTime int64, eventsFilePath string, file *os.File) (int, string, map[string]bool, error) {
	rows, tx, err := store.GetStore().PullEventRows(projectID, startTime, endTime)
	if err != nil {
		peLog.WithError(err).Error("SQL Query failed.")
		return 0, "", nil, err
	}
	defer U.CloseReadQuery(rows, tx)

	userIdMap := make(map[string]bool)
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
			return 0, "", nil, err
		}

		if _, ok := userIdMap[userID]; !ok {
			userIdMap[userID] = true
		}
		var eventPropertiesMap map[string]interface{}
		if eventProperties != nil {
			eventPropertiesBytes, err := eventProperties.Value()
			if err != nil {
				peLog.WithFields(log.Fields{"err": err}).Error("Unable to unmarshal event property.")
				return 0, "", nil, err
			}
			err = json.Unmarshal(eventPropertiesBytes.([]byte), &eventPropertiesMap)
			if err != nil {
				peLog.WithFields(log.Fields{"err": err}).Error("Unable to unmarshal event property.")
				return 0, "", nil, err
			}
		} else {
			peLog.WithFields(log.Fields{"err": err, "project_id": projectID}).Error("Nil event properties.")
		}

		var userPropertiesMap map[string]interface{}
		if userProperties != nil {
			userPropertiesBytes, err := userProperties.Value()
			if err != nil {
				peLog.WithFields(log.Fields{"err": err}).Error("Unable to unmarshal user property.")
				return 0, "", nil, err
			}

			err = json.Unmarshal(userPropertiesBytes.([]byte), &userPropertiesMap)
			if err != nil {
				peLog.WithFields(log.Fields{"err": err}).Error("Unable to unmarshal user property.")
				return 0, "", nil, err
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
			return 0, "", nil, err
		}
		line := string(lineBytes)
		if _, err := file.WriteString(fmt.Sprintf("%s\n", line)); err != nil {
			peLog.WithFields(log.Fields{"line": line, "err": err}).Error("Unable to write to file.")
			return 0, "", nil, err
		}

		lastEvent = &event
		rowCount++
	}
	if nilUserProperties > 0 {
		peLog.WithFields(log.Fields{"err": err, "project_id": projectID, "count": nilUserProperties}).Error("Nil user properties.")
	}
	err = rows.Err()
	if err != nil {
		// Error from DB is captured eg: timeout error
		peLog.WithFields(log.Fields{"err": err, "project_id": projectID}).Error("Error in executing query")
		return 0, "", nil, err
	}

	if rowCount > M.EventsPullLimit {
		// Todo(Dinesh): notify
		return rowCount, eventsFilePath, nil, fmt.Errorf("events count has exceeded the %d limit", M.EventsPullLimit)
	}

	if rowCount > 0 {
		peLog.WithFields(log.Fields{
			"FirstEventTimestamp": firstEvent.EventTimestamp,
			"LastEventTimesamp":   lastEvent.EventTimestamp,
		}).Info("Events time information.")
	}
	return rowCount, eventsFilePath, userIdMap, nil
}

// Pull Channel Data
func pullChannelData(channel string, projectID int64, startTime, endTime int64, eventsFilePath string) (int, string, error) {

	rows, tx, err := channelToPullMap[channel](projectID, startTime, endTime)
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

	rowCount := 0
	for rows.Next() {
		var id string
		var value *postgres.Jsonb
		var timestamp int64
		var docType int
		var smartProps *postgres.Jsonb
		if err = rows.Scan(&id, &value, &timestamp, &docType, &smartProps); err != nil {
			peLog.WithFields(log.Fields{"err": err}).Error("SQL Parse failed.")
			return 0, "", err
		}

		var valueMap map[string]interface{}
		if value != nil {
			valueBytes, err := value.Value()
			if err != nil {
				peLog.WithFields(log.Fields{"err": err}).Error("Unable to unmarshal value")
				return 0, "", err
			}
			err = json.Unmarshal(valueBytes.([]byte), &valueMap)
			if err != nil {
				peLog.WithFields(log.Fields{"err": err}).Error("Unable to unmarshal value")
				return 0, "", err
			}
		} else {
			peLog.WithFields(log.Fields{"err": err, "project_id": projectID}).Error("Nil value")
		}

		var smartPropsMap map[string]interface{}
		if smartProps != nil {
			smartPropsBytes, err := smartProps.Value()
			if err != nil {
				peLog.WithFields(log.Fields{"err": err}).Error("Unable to unmarshal smart props")
				return 0, "", err
			}
			err = json.Unmarshal(smartPropsBytes.([]byte), &smartPropsMap)
			if err != nil {
				peLog.WithFields(log.Fields{"err": err}).Error("Unable to unmarshal smart props")
				return 0, "", err
			}
		}

		doc := CounterCampaignFormat{
			Id:         id,
			Channel:    channel,
			Value:      valueMap,
			Timestamp:  timestamp,
			Doctype:    docType,
			SmartProps: smartPropsMap,
		}

		lineBytes, err := json.Marshal(doc)
		if err != nil {
			peLog.WithFields(log.Fields{"err": err}).Error("Unable to marshal document.")
			return 0, "", err
		}
		line := string(lineBytes)
		if _, err := file.WriteString(fmt.Sprintf("%s\n", line)); err != nil {
			peLog.WithFields(log.Fields{"line": line, "err": err}).Error("Unable to write to file.")
			return 0, "", err
		}
		rowCount++
	}

	if rowCount > M.EventsPullLimit {
		// Todo(Dinesh): notify
		return rowCount, eventsFilePath, fmt.Errorf("events count has exceeded the %d limit", M.EventsPullLimit)
	}

	return rowCount, eventsFilePath, nil
}

// Pull Users Data
func pullUsersData(dateField string, source int, group int, projectID int64, startTime, endTime int64, eventsFilePath string) (int, string, error) {

	rows, tx, err := store.GetStore().PullUsersRowsForWI(projectID, startTime, endTime, dateField, source, group)
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

	rowCount := 0
	for rows.Next() {
		var id string
		var properties *postgres.Jsonb
		var is_anonymous bool
		var join_timestamp int64
		if err = rows.Scan(&id, &properties, &is_anonymous, &join_timestamp); err != nil {
			peLog.WithFields(log.Fields{"err": err}).Error("SQL Parse failed.")
			return 0, "", err
		}

		var propsMap map[string]interface{}
		if properties != nil {
			propsBytes, err := properties.Value()
			if err != nil {
				peLog.WithFields(log.Fields{"err": err}).Error("Unable to unmarshal properties")
				return 0, "", err
			}
			err = json.Unmarshal(propsBytes.([]byte), &propsMap)
			if err != nil {
				peLog.WithFields(log.Fields{"err": err}).Error("Unable to unmarshal properties")
				return 0, "", err
			}
		} else {
			peLog.WithFields(log.Fields{"err": err, "project_id": projectID}).Error("Nil properties")
		}

		user := CounterUserFormat{
			Id:            id,
			Properties:    propsMap,
			Is_Anonymous:  is_anonymous,
			JoinTimestamp: join_timestamp,
		}

		lineBytes, err := json.Marshal(user)
		if err != nil {
			peLog.WithFields(log.Fields{"err": err}).Error("Unable to marshal user.")
			return 0, "", err
		}
		line := string(lineBytes)
		if _, err := file.WriteString(fmt.Sprintf("%s\n", line)); err != nil {
			peLog.WithFields(log.Fields{"line": line, "err": err}).Error("Unable to write to file.")
			return 0, "", err
		}
		rowCount++
	}

	if rowCount > M.EventsPullLimit {
		// Todo(Dinesh): notify
		return rowCount, eventsFilePath, fmt.Errorf("events count has exceeded the %d limit", M.EventsPullLimit)
	}

	return rowCount, eventsFilePath, nil
}

// pull Events for Archive
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
			userPropertiesString, _ = json.Marshal(M.SanitizeUserProperties(*userPropertiesMap))
		} else {
			peLog.WithFields(log.Fields{"err": err, "project_id": projectID}).Error("Nil user properties.")
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
		user := M.ArchiveUsersTableFormat{
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

func PullAllData(projectId int64, configs map[string]interface{}) (map[string]interface{}, bool) {

	modelType := configs["modelType"].(string)
	startTimestamp := configs["startTimestamp"].(int64)
	endTimestamp := configs["endTimestamp"].(int64)
	diskManager := configs["diskManager"].(*serviceDisk.DiskDriver)
	cloudManager := configs["cloudManager"].(*filestore.FileManager)
	cloudManagerTmp := configs["cloudManagertmp"].(*filestore.FileManager)

	hardPull := configs["hardPull"].(*bool)
	fileTypes := configs["fileTypes"].(map[int64]bool)
	beamConfig := configs["beamConfig"].(*RunBeamConfig)

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

	integrations := store.GetStore().IsIntegrationAvailable(projectId)

	success := true

	// EVENTS
	if fileTypes[fileType["events"]] {
		exists := false
		if !*hardPull {
			if ok, _ := checkEventFileExists(cloudManager, projectId, startTimestamp, modelType); ok {
				status["events-info"] = "File already exists"
				exists = true
			}
		}

		if !exists {
			pull := true
			var errMsg string = "data not available for "
			for _, channel := range []string{M.HUBSPOT, M.SALESFORCE} {
				if !integrations[channel] {
					status[channel+"-info"] = "Not Integrated"
				} else {
					if !store.GetStore().IsDataAvailable(projectId, channel, uint64(endTimestampInProjectTimezone)) {
						errMsg += channel + " "
						pull = false
					}
				}
			}

			if pull {
				if beamConfig.RunOnBeam == true {
					if _, ok := PullEventsDataDaily(projectId, cloudManager, cloudManagerTmp, diskManager, startTimestamp, startTimestampInProjectTimezone, endTimestampInProjectTimezone, modelType, beamConfig, status, logCtx); !ok {
						return status, false
					}

				} else {
					if _, ok := PullEventsData(projectId, cloudManager, diskManager, startTimestamp, startTimestampInProjectTimezone, endTimestampInProjectTimezone, modelType, status, logCtx); !ok {
						return status, false
					}
				}
			} else {
				success = false
				status["events-error"] = errMsg
			}
		}

	}

	// CAMPAIGN REPORTS
	for _, channel := range []string{M.ADWORDS, M.BINGADS, M.FACEBOOK, M.GOOGLE_ORGANIC, M.LINKEDIN} {
		if fileTypes[fileType[channel]] {
			if !*hardPull {
				if ok, _ := checkChannelFileExists(channel, cloudManager, projectId, startTimestamp, modelType); ok {
					status[channel+"-info"] = "File already exists"
					continue
				}
			}
			if !integrations[channel] {
				status[channel+"-info"] = "Not Integrated"
			} else {
				if store.GetStore().IsDataAvailable(projectId, channel, uint64(endTimestampInProjectTimezone)) {
					if _, ok := PullDataForChannel(channel, projectId, cloudManager, diskManager, startTimestamp, startTimestampInProjectTimezone, endTimestampInProjectTimezone, modelType, status, logCtx); !ok {
						return status, false
					}
				} else {
					success = false
					status[channel+"-error"] = "Data not available"
				}
			}
		}
	}

	//USERS
	if fileTypes[fileType["users"]] {
		if _, ok := PullUsersDataForCustomMetrics(projectId, cloudManager, diskManager, startTimestamp, startTimestampInProjectTimezone, endTimestampInProjectTimezone, modelType, hardPull, status, logCtx); !ok {
			return status, false
		}
	}

	return status, success
}

func PullEventsData(projectId int64, cloudManager *filestore.FileManager, diskManager *serviceDisk.DiskDriver, startTimestamp, startTimestampInProjectTimezone, endTimestampInProjectTimezone int64,
	modelType string, status map[string]interface{}, logCtx *log.Entry) (error, bool) {

	logCtx.Info("Pulling events.")
	// Writing events to tmp file before upload.
	fPath, fName := diskManager.GetModelEventsFilePathAndName(projectId, startTimestamp, modelType)
	serviceDisk.MkdirAll(fPath) // create dir if not exist.
	tmpEventsFile := fPath + fName
	startAt := time.Now().UnixNano()

	eventsCount, eventsFilePath, err := pullEvents(projectId, startTimestampInProjectTimezone, endTimestampInProjectTimezone, tmpEventsFile)
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

	tmpOutputFile, err := os.Open(tmpEventsFile)
	if err != nil {
		logCtx.WithField("error", err).Error("Failed to pull events. Write to tmp failed.")
		status["events-error"] = "Failed to pull events. Write to tmp failed."
		return err, false
	}

	cDir, cName := (*cloudManager).GetModelEventsFilePathAndName(projectId, startTimestamp, modelType)
	err = (*cloudManager).Create(cDir, cName, tmpOutputFile)
	if err != nil {
		logCtx.WithField("error", err).Error("Failed to pull events. Upload failed.")
		status["events-error"] = "Failed to pull events. Upload failed."
		return err, false
	}

	logCtx.WithFields(log.Fields{
		"EventsCount":           eventsCount,
		"EventsFilePath":        eventsFilePath,
		"TimeTakenToPullEvents": timeTakenToPullEvents,
	}).Info("Successfully pulled events and written to file.")

	status["EventsCount"] = eventsCount
	return nil, true
}

func PullDataForChannel(channel string, projectId int64, cloudManager *filestore.FileManager, diskManager *serviceDisk.DiskDriver, startTimestamp, startTimestampInProjectTimezone, endTimestampInProjectTimezone int64,
	modelType string, status map[string]interface{}, logCtx *log.Entry) (error, bool) {

	logCtx.Infof("Pulling " + channel)
	// Writing adwords data to tmp file before upload.
	fPath, fName := diskManager.GetModelChannelFilePathAndName(channel, projectId, startTimestamp, modelType)
	serviceDisk.MkdirAll(fPath) // create dir if not exist.
	tmpEventsFile := fPath + fName
	startAt := time.Now().UnixNano()

	count, filePath, err := pullChannelData(channel, projectId, startTimestampInProjectTimezone, endTimestampInProjectTimezone, tmpEventsFile)
	if err != nil {
		logCtx.WithField("error", err).Error("Pull " + channel + " failed. Pull and write failed.")
		status[channel+"-error"] = err.Error()
		return err, false
	}
	timeTaken := (time.Now().UnixNano() - startAt) / 1000000

	// Zero events. Returns eventCount as 0.
	if count == 0 {
		logCtx.Info("No " + channel + " data found.")
		status[channel+"-info"] = "No " + channel + " data found."
		return err, true
	}

	tmpOutputFile, err := os.Open(tmpEventsFile)
	if err != nil {
		logCtx.WithField("error", err).Error("Failed to pull " + channel + ". Write to tmp failed.")
		status[channel+"-error"] = "Failed to pull " + channel + ". Write to tmp failed."
		return err, false
	}

	cDir, cName := (*cloudManager).GetModelChannelFilePathAndName(channel, projectId, startTimestamp, modelType)
	err = (*cloudManager).Create(cDir, cName, tmpOutputFile)
	if err != nil {
		logCtx.WithField("error", err).Errorf("Failed to pull "+channel+". Upload failed.", channel)
		status[channel+"-error"] = "Failed to pull " + channel + ". Upload failed."
		return err, false
	}

	logCtx.WithFields(log.Fields{
		channel + "-RowsCount":       count,
		channel + "-FilePath":        filePath,
		channel + "-TimeTakenToPull": timeTaken,
	}).Info("Successfully pulled " + channel + " and written to file.")

	status["DataFolderPath"] = fPath
	status[channel+"-RowsCount"] = count
	return nil, true
}

func PullUsersDataForCustomMetrics(projectId int64, cloudManager *filestore.FileManager, diskManager *serviceDisk.DiskDriver, startTimestamp, startTimestampInProjectTimezone, endTimestampInProjectTimezone int64,
	modelType string, hardPull *bool, status map[string]interface{}, logCtx *log.Entry) (error, bool) {

	filesCreated := 0
	totalRowsCount := 0
	var totalTimeTaken int64 = 0

	logCtx.Infof("Pulling users")

	groupsMap := make(map[string]int)
	{
		groups, errCode := store.GetStore().GetGroups(projectId)
		if errCode != http.StatusFound {
			status["users-error"] = "failed to get groups for project"
			return errors.New("failed to get groups for project"), false
		}

		for _, group := range groups {
			groupsMap[group.Name] = group.ID
		}
		if _, ok := groupsMap[M.USERS]; !ok {
			groupsMap[M.USERS] = 0
		}
	}

	// get unique datefields
	uniqueDateFileds := make(map[string]string)
	{
		customMetrics, errStr, getStatus := store.GetStore().GetCustomMetricsByProjectId(projectId)
		if getStatus != http.StatusFound {
			logCtx.WithField("error", errStr).Error("Pull users failed. get custom metrics failed.")
			status["users-error"] = errStr
			return fmt.Errorf("%s", errStr), false
		}
		for _, customMetric := range customMetrics {
			if customMetric.TypeOfQuery == M.ProfileQueryType && customMetric.ObjectType != "others" {
				var customMetricTransformation M.CustomMetricTransformation
				err := U.DecodePostgresJsonbToStructType(customMetric.Transformations, &customMetricTransformation)
				if err != nil {
					status["users-error"] = "Error during decode of custom metrics transformations"
					return err, false
				}
				if _, ok := uniqueDateFileds[customMetricTransformation.DateField]; !ok {
					uniqueDateFileds[customMetricTransformation.DateField] = customMetric.ObjectType
				}
			}
		}
	}

	for dateField, objectType := range uniqueDateFileds {
		if !*hardPull {
			if ok, _ := checkUsersFileExists(dateField, cloudManager, projectId, startTimestamp, modelType); ok {
				status["users-"+dateField+"-info"] = "File already exists"
				continue
			}
		}
		var group int
		if groupStr, ok := M.MapOfKPICategoryToProfileGroupAnalysis[objectType]; ok {
			if _, ok := groupsMap[groupStr]; ok {
				group = groupsMap[groupStr]
			}
		} else {
			status["users-"+dateField+"-error"] = "Unknown object type"
			continue
		}

		var source int
		if sourceStr, ok := M.MapOfKPIToProfileType[objectType]; ok {
			source = M.UserSourceMap[sourceStr]
		} else {
			status["users-"+dateField+"-error"] = "Unknown object type"
			continue
		}
		if err, ok := PullDataForUsers(projectId, cloudManager, diskManager, startTimestamp, startTimestampInProjectTimezone, endTimestampInProjectTimezone, modelType, dateField, source, group, status, logCtx); !ok {
			return err, false
		} else {
			totalRowsCount += status["users-RowsCount"].(int)
			totalTimeTaken += status["users-TimeTakenToPull"].(int64)
			filesCreated += 1
			delete(status, "users-RowsCount")
			delete(status, "users-TimeTakenToPull")
		}
	}

	status["users-NumberOfDatefieldsProcessed"] = filesCreated
	status["users-TotalRowsCount"] = totalRowsCount
	status["users-TotalTimeTakenToPull"] = totalTimeTaken
	logCtx.WithFields(log.Fields{
		"users-NumberOfDatefieldsProcessed": filesCreated,
		"users-TotalRowsCount":              totalRowsCount,
		"users-TotalTimeTakenToPull":        totalTimeTaken,
	}).Info("Successfully pulled users data and written to files.")
	return nil, true
}

func PullDataForUsers(projectId int64, cloudManager *filestore.FileManager, diskManager *serviceDisk.DiskDriver, startTimestamp, startTimestampInProjectTimezone, endTimestampInProjectTimezone int64,
	modelType string, dateField string, source int, group int, status map[string]interface{}, logCtx *log.Entry) (error, bool) {

	logCtx.Infof("Pulling users for %s", dateField)
	cDir, cName := (*cloudManager).GetModelUsersFilePathAndName(dateField, projectId, startTimestamp, modelType)

	// Writing users data to tmp file before upload.
	fPath, fName := diskManager.GetModelUsersFilePathAndName(dateField, projectId, startTimestamp, modelType)
	serviceDisk.MkdirAll(fPath) // create dir if not exist.
	tmpEventsFile := fPath + fName
	startAt := time.Now().UnixNano()

	count, _, err := pullUsersData(dateField, source, group, projectId, startTimestampInProjectTimezone, endTimestampInProjectTimezone, tmpEventsFile)
	if err != nil {
		logCtx.WithField("error", err).Error("Pull users failed for" + dateField + ". Pull and write failed.")
		status["users-"+dateField+"-error"] = err.Error()
		return err, false
	}
	timeTaken := (time.Now().UnixNano() - startAt) / 1000000

	status["users-RowsCount"] = count
	status["users-TimeTakenToPull"] = timeTaken
	// Zero events. Returns eventCount as 0.
	if count == 0 {
		logCtx.Info("No users data found for " + dateField)
		status["users-"+dateField+"-info"] = "No users data found."
		return err, true
	}

	tmpOutputFile, err := os.Open(tmpEventsFile)
	if err != nil {
		logCtx.WithField("error", err).Error("Failed to pull users. Write to tmp failed.")
		status["users-"+dateField+"-error"] = "Failed to pull users. Write to tmp failed."
		return err, false
	}

	err = (*cloudManager).Create(cDir, cName, tmpOutputFile)
	if err != nil {
		logCtx.WithField("error", err).Error("Failed to pull users. Upload failed.")
		status["users-"+dateField+"-error"] = "Failed to pull users. Upload failed."
		return err, false
	}

	return nil, true
}

func checkEventFileExists(cloudManager *filestore.FileManager, projectId int64, startTimestamp int64, modelType string) (bool, error) {
	path, name := (*cloudManager).GetModelEventsFilePathAndName(projectId, startTimestamp, modelType)
	return checkFileExists(cloudManager, path, name)
}

func checkChannelFileExists(channel string, cloudManager *filestore.FileManager, projectId int64, startTimestamp int64, modelType string) (bool, error) {
	path, name := (*cloudManager).GetModelChannelFilePathAndName(channel, projectId, startTimestamp, modelType)
	return checkFileExists(cloudManager, path, name)
}

func checkUsersFileExists(dateField string, cloudManager *filestore.FileManager, projectId int64, startTimestamp int64, modelType string) (bool, error) {
	path, name := (*cloudManager).GetModelUsersFilePathAndName(dateField, projectId, startTimestamp, modelType)
	return checkFileExists(cloudManager, path, name)
}

func checkFileExists(cloudManager *filestore.FileManager, path, name string) (bool, error) {
	if _, err := (*cloudManager).Get(path, name); err != nil {
		log.WithFields(log.Fields{"err": err, "filePath": path,
			"fileName": name}).Error("Failed to fetch from cloud path")
		return false, err
	} else {
		return true, nil
	}
}
