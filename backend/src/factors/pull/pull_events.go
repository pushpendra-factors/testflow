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
	"strconv"
	"strings"
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
	nilEventProperties := 0

	var perQueryRange = endTimestamp - startTimestamp
	startSplit, endSplit := getTimeStampArraysAfterSplit(splitRange, noOfSplits, startTimeTimezone, endTimeTimezone, &perQueryRange)
	for i := range startSplit {
		splitRowCount := 0
		splitNum := i + 1
		pullStartTime := startSplit[i]
		pullEndTime := endSplit[i]

		//pull data from db in the form of rows
		log.WithFields(log.Fields{"start": pullStartTime, "end": pullEndTime}).Infof("Executing split %d out of %d", splitNum, len(startSplit))
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
			var is_group_user_null sql.NullBool
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
				&is_group_user_null, &group_1_user_id_null, &group_2_user_id_null, &group_3_user_id_null, &group_4_user_id_null,
				&group_5_user_id_null, &group_6_user_id_null, &group_7_user_id_null, &group_8_user_id_null,
				&group_1_id_null, &group_2_id_null, &group_3_id_null, &group_4_id_null,
				&group_5_id_null, &group_6_id_null, &group_7_id_null, &group_8_id_null); err != nil {
				log.WithFields(log.Fields{"err": err}).Error("SQL Parse failed.")
				return 0, err
			}

			//get file timestamp w.r.t event timestamp
			daysFromStart := int64(math.Floor(float64(eventTimestamp-startTimeTimezone) / float64(U.Per_day_epoch)))
			fileTimestamp := startTimestamp + daysFromStart*U.Per_day_epoch

			//get apt writer w.r.t file timestamp
			writer, err := getAptWriterFromMap(projectID, writerMap, cloudManager, fileName, fileTimestamp, U.DataTypeEvent, "")
			if err != nil {
				log.WithError(err).Error("error getting apt writer from file")
				return 0, err
			}

			//event props
			eventPropertiesMap, err := getMapFromPostgresJson(eventProperties)
			if err != nil {
				log.WithError(err).WithFields(log.Fields{"project_id": projectID}).Error("error getting event properties")
				return 0, err
			}
			if len(eventPropertiesMap) == 0 {
				nilEventProperties++
			}

			//user props
			userPropertiesMap, err := getMapFromPostgresJson(userProperties)
			if err != nil {
				log.WithError(err).WithFields(log.Fields{"project_id": projectID}).Error("error getting user properties")
				return 0, err
			}
			if len(userPropertiesMap) == 0 {
				nilUserProperties++
			}

			//group variables
			is_group_user := U.IfThenElse(is_group_user_null.Valid, is_group_user_null.Bool, false).(bool)

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

			//initialise event with apt values
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

			//write event
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

			//info variables
			if event.EventTimestamp < firstEventTimestamp {
				firstEventTimestamp = event.EventTimestamp
			}
			if event.EventTimestamp > lastEventTimestamp {
				lastEventTimestamp = event.EventTimestamp
			}
			splitRowCount++
		}
		// Error from DB is captured eg: timeout error
		err = rows.Err()
		if err != nil {
			log.WithFields(log.Fields{"err": err, "project_id": projectID}).Error("Error in executing query")
			return 0, err
		}

		U.CloseReadQuery(rows, tx)
		rowCount += splitRowCount
		log.WithFields(log.Fields{"start": pullStartTime, "end": pullEndTime, "rowCount": splitRowCount}).Infof("Completed split %d out of %d", splitNum, len(startSplit))
	}
	noOfFiles := len(writerMap)
	log.WithFields(log.Fields{"rowCount": rowCount, "writersCount": noOfFiles, "start": startTimeTimezone, "end": endTimeTimezone}).Info("Allotted rows to writers")

	if rowCount > M.EventsPullLimit {
		// Todo(Dinesh): notify
		return rowCount, fmt.Errorf("events count has exceeded the %d limit", M.EventsPullLimit)
	}

	//Closing writers (writing files)
	log.WithFields(log.Fields{"writers": noOfFiles, "start": startTimeTimezone, "end": endTimeTimezone}).Info("Closing writers")
	for _, writer := range writerMap {
		err := (*writer).Close()
		if err != nil {
			log.WithError(err).Error("Error closing writer")
			return 0, err
		}
	}
	log.WithFields(log.Fields{"writers": noOfFiles, "start": startTimeTimezone, "end": endTimeTimezone}).Info("All files successfully created")

	if rowCount > 0 {
		log.WithFields(log.Fields{
			"FirstEventTimestamp": firstEventTimestamp,
			"LastEventTimesamp":   lastEventTimestamp,
			"FilesCreated":        noOfFiles,
			"NumberOfRows":        rowCount,
			"NilEventProperties":  nilEventProperties,
			"NilUserProperties":   nilUserProperties,
		}).Info("Events pulled information.")
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

// check if add session is completed till given timestamp
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

// check if events file has been generated already for given start and end timestamp, in dataTimestamp's day and
// its previous day's folder
func CheckEventFileExists(cloudManager *filestore.FileManager, projectId int64, dataTimestamp, startTimestamp, endTimestamp int64) (bool, error) {
	path, _ := (*cloudManager).GetDailyEventArchiveFilePathAndName(projectId, dataTimestamp, startTimestamp, endTimestamp)
	filesPathList := (*cloudManager).ListFiles(path)
	path, _ = (*cloudManager).GetDailyEventArchiveFilePathAndName(projectId, dataTimestamp-U.Per_day_epoch, startTimestamp, endTimestamp)
	filesPathList = append(filesPathList, (*cloudManager).ListFiles(path)...)
	return checkStartAndEndtimestampsExist(filesPathList, U.EVENTS_FILENAME_PREFIX, startTimestamp, endTimestamp)
}

// check if start and end times exist in file names in filesPathList
func checkStartAndEndtimestampsExist(filesPathList []string, fileNamePrefix string, startTimestamp int64, endTimestamp int64) (bool, error) {
	var startExists, endExists bool
	for _, fpath := range filesPathList {
		partFNamelist := strings.Split(fpath, "/")
		partFileName := partFNamelist[len(partFNamelist)-1]
		if !strings.HasPrefix(partFileName, fileNamePrefix) {
			continue
		}
		start, end, err := getStartAndEndTimestampFromFileName(partFileName)
		if err != nil {
			log.WithError(err).Error("error getStartAndEndTimestampFromFileName")
			return false, err
		}
		if start == startTimestamp {
			startExists = true
		}
		if end == endTimestamp {
			endExists = true
		}
		if startExists && endExists {
			return true, nil
		}
	}
	return false, nil
}

// get start and end timestamp from archive file name
func getStartAndEndTimestampFromFileName(fileName string) (int64, int64, error) {
	nameArr := strings.Split(fileName, "_")
	startEndTxtStr := nameArr[len(nameArr)-1]
	startEndStr := strings.Replace(startEndTxtStr, ".txt", "", 1)
	startEndArr := strings.Split(startEndStr, "-")
	start, err := strconv.ParseInt(startEndArr[0], 10, 64)
	if err != nil {
		log.WithError(err).Error("error parsing startTimestamp to int64")
		return 0, 0, err
	}
	end, err := strconv.ParseInt(startEndArr[1], 10, 64)
	if err != nil {
		log.WithError(err).Error("error parsing endTimestamp to int64")
		return 0, 0, err
	}
	return start, end, nil
}

// get 2 equal length arrays with start and end timestamps after split
func getTimeStampArraysAfterSplit(splitRange bool, noOfSplits int, startTimestamp, endTimestamp int64, perQueryRange *int64) ([]int64, []int64) {
	var startSplit = make([]int64, 0)
	var endSplit = make([]int64, 0)
	if splitRange && noOfSplits > 1 {
		var tmpStart, tmpEnd int64
		*perQueryRange = (endTimestamp - startTimestamp) / int64(noOfSplits)
		for i := 0; i < noOfSplits; i++ {
			if i == 0 {
				tmpStart = startTimestamp
			} else {
				tmpStart = tmpEnd + 1
			}
			if i == noOfSplits-1 {
				tmpEnd = endTimestamp
			} else {
				tmpEnd = tmpStart + *perQueryRange - 1
			}
			startSplit = append(startSplit, tmpStart)
			endSplit = append(endSplit, tmpEnd)
		}
	} else {
		startSplit = []int64{startTimestamp}
		endSplit = []int64{endTimestamp}
	}
	return startSplit, endSplit
}

// get writer (initialise or get from map)
func getAptWriterFromMap(projectID int64, writerMap map[int64]*io.WriteCloser, cloudManager *filestore.FileManager, cName string, dataTimestamp int64, dataType, channelOrDatefield string) (*io.WriteCloser, error) {
	writer, ok := writerMap[dataTimestamp]
	if !ok {
		cPath, _ := GetDailyArchiveFilePathAndName(cloudManager, dataType, channelOrDatefield, projectID, dataTimestamp, 0, 0)
		cloudWriter, err := (*cloudManager).GetWriter(cPath, cName)
		if err != nil {
			log.WithFields(log.Fields{"file": cName, "err": err}).Error("Unable to get cloud file writer")
			return writer, err
		}
		writerMap[dataTimestamp] = &cloudWriter
		writer = &cloudWriter
	}
	return writer, nil
}

// get map from json
func getMapFromPostgresJson(data *postgres.Jsonb) (map[string]interface{}, error) {
	var dataMap map[string]interface{}
	if data != nil {
		dataBytes, err := data.Value()
		if err != nil {
			log.WithFields(log.Fields{"err": err}).Error("Unable to get value of json.")
			return nil, err
		}
		err = json.Unmarshal(dataBytes.([]byte), &dataMap)
		if err != nil {
			log.WithFields(log.Fields{"err": err}).Error("Unable to unmarshal property.")
			return nil, err
		}
	}
	return dataMap, nil
}

// get data file (daily) path and name given fileManager and dataType
//
// channelOrDatefield - channel name for ad_reports dataType and dateField for users dataType,
// sortOnGroup - (0:uid, i:group_i_id) for events dataType
func GetDailyArchiveFilePathAndName(fileManager *filestore.FileManager, dataType, channelOrDatefield string, projectId int64, dataTimestamp, startTime, endTime int64) (string, string) {
	if dataType == U.DataTypeEvent {
		return (*fileManager).GetDailyEventArchiveFilePathAndName(projectId, dataTimestamp, startTime, endTime)
	} else if dataType == U.DataTypeAdReport {
		channel := channelOrDatefield
		return (*fileManager).GetDailyChannelArchiveFilePathAndName(channel, projectId, dataTimestamp, startTime, endTime)
	} else if dataType == U.DataTypeUser {
		dateField := channelOrDatefield
		return (*fileManager).GetDailyUsersArchiveFilePathAndName(dateField, projectId, dataTimestamp, startTime, endTime)
	} else {
		log.Errorf("wrong dataType: %s", dataType)
	}
	return "", ""
}
