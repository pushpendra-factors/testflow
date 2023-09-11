package pull

import (
	"encoding/json"
	"errors"
	"factors/filestore"
	M "factors/model/model"
	"factors/model/store"
	U "factors/util"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

type CounterUserFormat struct {
	Id            string                 `json:"id"`
	Properties    map[string]interface{} `json:"pr"`
	Is_Anonymous  bool                   `json:"ia"`
	JoinTimestamp int64                  `json:"ts"`
}

// get all profile type custom metrics and get their datefields, pull users data for each datefield
func PullUsersDataForCustomMetrics(projectId int64, cloudManager *filestore.FileManager, startTimestamp, endTimestamp, startTimestampInProjectTimezone, endTimestampInProjectTimezone int64, splitRangeProjectIds []int64, noOfSplits int, hardPull *bool, status map[string]interface{}, logCtx *log.Entry) (error, bool) {

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

	// get unique datefields mapped to object type
	uniqueDateFieldsAndObjectTypes, err := getUniqueDateFieldsAndObjectTypes(projectId, M.ProfileQueryType, logCtx)
	if err != nil {
		logCtx.WithField("error", err).Error("Pull users failed")
		status["users-error"] = err.Error()
		return err, false
	}

	for dateField, objectType := range uniqueDateFieldsAndObjectTypes {
		if !*hardPull {
			if ok, _ := checkUsersFileExists(dateField, cloudManager, projectId, startTimestamp, startTimestamp, endTimestamp); ok {
				status["users-"+dateField+"-info"] = "File already exists"
				continue
			}
		}

		//get group from objectType
		var group int
		if groupStr, ok := M.MapOfKPICategoryToProfileGroupAnalysis[objectType]; ok {
			if _, ok := groupsMap[groupStr]; ok {
				group = groupsMap[groupStr]
			}
		} else {
			status["users-"+dateField+"-error"] = "Unknown object type"
			continue
		}

		//get source from objectType
		var source int
		if sourceStr, ok := M.MapOfKPIToProfileType[objectType]; ok {
			source = M.UserSourceMap[sourceStr]
		} else {
			status["users-"+dateField+"-error"] = "Unknown object type"
			continue
		}

		if err, ok := pullDataForUsers(projectId, cloudManager, startTimestamp, endTimestamp, startTimestampInProjectTimezone, endTimestampInProjectTimezone, splitRangeProjectIds, noOfSplits, dateField, source, group, status, logCtx); !ok {
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

// pull users data for a datefield into cloud files with proper logging
func pullDataForUsers(projectId int64, cloudManager *filestore.FileManager, startTimestamp, endTimestamp, startTimestampInProjectTimezone, endTimestampInProjectTimezone int64, splitRangeProjectIds []int64, noOfSplits int,
	dateField string, source int, group int, status map[string]interface{}, logCtx *log.Entry) (error, bool) {

	var splitRange bool = false
	if U.ContainsInt64InArray(splitRangeProjectIds, projectId) {
		splitRange = true
	}

	logCtx.Infof("Pulling users for %s", dateField)

	_, cName := (*cloudManager).GetDailyUsersArchiveFilePathAndName(dateField, projectId, 0, startTimestamp, endTimestamp)
	startAt := time.Now().UnixNano()

	count, err := pullUsersDataFromDB(dateField, source, group, projectId, startTimestampInProjectTimezone, endTimestampInProjectTimezone, cName, cloudManager, startTimestamp, endTimestamp, splitRange, noOfSplits)
	if err != nil {
		logCtx.WithField("error", err).Error("Pull users failed for" + dateField + ". Pull and write failed.")
		status["users-"+dateField+"-error"] = err.Error()
		return err, false
	}
	timeTaken := (time.Now().UnixNano() - startAt) / 1000000

	status["users-RowsCount"] = count
	status["users-TimeTakenToPull"] = timeTaken
	if count == 0 {
		logCtx.Info("No users data found for " + dateField)
		status["users-"+dateField+"-info"] = "No users data found."
		return err, true
	}

	return nil, true
}

// pull users(rows) for a datefield from db into cloud files
func pullUsersDataFromDB(dateField string, source int, group int, projectID int64, startTimeTimezone, endTimeTimezone int64, fileName string, cloudManager *filestore.FileManager, startTimestamp, endTimestamp int64, splitRange bool, noOfSplits int) (int, error) {

	var writerMap = make(map[int64]*io.WriteCloser)
	rowCount := 0
	nilProperties := 0

	var perQueryRange = endTimestamp - startTimestamp
	startSplit, endSplit := getTimeStampArraysAfterSplit(splitRange, noOfSplits, startTimeTimezone, endTimeTimezone, &perQueryRange)
	for i := range startSplit {
		splitRowCount := 0
		splitNum := i + 1
		pullStartTime := startSplit[i]
		pullEndTime := endSplit[i]

		log.WithFields(log.Fields{"start": pullStartTime, "end": pullEndTime}).Infof("Executing split %d out of %d", splitNum, len(startSplit))
		rows, tx, err := store.GetStore().PullUsersRowsForWIV2(projectID, pullStartTime, pullEndTime, dateField, source, group)
		if err != nil {
			log.WithError(err).Error("SQL Query failed.")
			return 0, err
		}

		for rows.Next() {
			var id string
			var properties *postgres.Jsonb
			var is_anonymous bool
			var join_timestamp int64
			var timestamp int
			if err = rows.Scan(&id, &properties, &is_anonymous, &join_timestamp, &timestamp); err != nil {
				log.WithFields(log.Fields{"err": err}).Error("SQL Parse failed.")
				return 0, err
			}

			//get file timestamp w.r.t event timestamp
			timestamp = U.GetTimestampInSecs(timestamp)
			daysFromStart := int64(math.Floor(float64(int64(timestamp)-startTimeTimezone) / float64(U.Per_day_epoch)))
			fileTimestamp := startTimestamp + daysFromStart*U.Per_day_epoch

			//get apt writer w.r.t file timestamp
			writer, err := getAptWriterFromMap(projectID, writerMap, cloudManager, fileName, fileTimestamp, U.DataTypeUser, dateField)
			if err != nil {
				log.WithError(err).Error("error getting apt writer from file")
				return 0, err
			}

			//properties
			propertiesMap, err := getMapFromPostgresJson(properties)
			if err != nil {
				log.WithError(err).WithFields(log.Fields{"project_id": projectID}).Error("error getting properties")
				return 0, err
			}
			if len(propertiesMap) == 0 {
				nilProperties++
			}

			//initialise user
			user := CounterUserFormat{
				Id:            id,
				Properties:    propertiesMap,
				Is_Anonymous:  is_anonymous,
				JoinTimestamp: join_timestamp,
			}

			//write user to file
			userBytes, err := json.Marshal(user)
			if err != nil {
				log.WithFields(log.Fields{"err": err}).Error("Unable to marshal user.")
				return 0, err
			}
			line := string(userBytes)
			if _, err := io.WriteString(*writer, fmt.Sprintf("%s\n", line)); err != nil {
				log.WithFields(log.Fields{"line": line, "err": err}).Error("Unable to write to file.")
				return 0, err
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

	if rowCount > M.UsersPullLimit {
		// Todo(Dinesh): notify
		return rowCount, fmt.Errorf("users count has exceeded the %d limit", M.UsersPullLimit)
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
			"FilesCreated":  noOfFiles,
			"NumberOfRows":  rowCount,
			"NilProperties": nilProperties,
		}).Info("Events pulled information.")
	}

	return rowCount, nil
}

// check if users file for given dateField has been generated already for given start and end timestamp, in dataTimestamp's day and
// its previous day's folder
func checkUsersFileExists(dateField string, cloudManager *filestore.FileManager, projectId int64, dataTimestamp, startTimestamp, endTimestamp int64) (bool, error) {
	path, _ := (*cloudManager).GetDailyUsersArchiveFilePathAndName(dateField, projectId, dataTimestamp, startTimestamp, endTimestamp)
	filesPathList := (*cloudManager).ListFiles(path)
	path, _ = (*cloudManager).GetDailyUsersArchiveFilePathAndName(dateField, projectId, dataTimestamp-U.Per_day_epoch, startTimestamp, endTimestamp)
	filesPathList = append(filesPathList, (*cloudManager).ListFiles(path)...)
	return checkStartAndEndtimestampsExist(filesPathList, dateField, startTimestamp, endTimestamp)
}

// get all datefields and corresponding object type
func getUniqueDateFieldsAndObjectTypes(projectId int64, queryType int, logCtx *log.Entry) (map[string]string, error) {
	var uniqueDateFieldsToObjectType = make(map[string]string)
	customMetrics, errStr, getStatus := store.GetStore().GetCustomMetricByProjectIdAndQueryType(projectId, queryType)
	if getStatus != http.StatusFound {
		logCtx.WithField("error", errStr).Error("Get custom metrics failed.")
		return uniqueDateFieldsToObjectType, fmt.Errorf("%s", errStr)
	}
	for _, customMetric := range customMetrics {
		var customMetricTransformation M.CustomMetricTransformation
		err := U.DecodePostgresJsonbToStructType(customMetric.Transformations, &customMetricTransformation)
		if err != nil {
			logCtx.WithField("error", err).Error("Error during decode of custom metrics transformations")
			return uniqueDateFieldsToObjectType, err
		}
		if _, ok := uniqueDateFieldsToObjectType[customMetricTransformation.DateField]; !ok && customMetric.TypeOfQuery == 1 {
			uniqueDateFieldsToObjectType[customMetricTransformation.DateField] = customMetric.ObjectType
		}
	}
	return uniqueDateFieldsToObjectType, nil
}
