package pull

import (
	"encoding/json"
	"errors"
	"factors/filestore"
	M "factors/model/model"
	"factors/model/store"
	serviceDisk "factors/services/disk"
	U "factors/util"
	"fmt"
	"math"
	"net/http"
	"os"
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
func PullUsersDataForCustomMetrics(projectId int64, cloudManager *filestore.FileManager, diskManager *serviceDisk.DiskDriver, startTimestamp, endTimestamp, startTimestampInProjectTimezone, endTimestampInProjectTimezone int64, hardPull *bool, status map[string]interface{}, logCtx *log.Entry) (error, bool) {

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
	uniqueDateFields := make(map[string]string)
	{
		customMetrics, errStr, getStatus := store.GetStore().GetCustomMetricByProjectIdAndQueryType(projectId, M.ProfileQueryType)
		if getStatus != http.StatusFound {
			logCtx.WithField("error", errStr).Error("Pull users failed. get custom metrics failed.")
			status["users-error"] = errStr
			return fmt.Errorf("%s", errStr), false
		}
		for _, customMetric := range customMetrics {
			var customMetricTransformation M.CustomMetricTransformation
			err := U.DecodePostgresJsonbToStructType(customMetric.Transformations, &customMetricTransformation)
			if err != nil {
				status["users-error"] = "Error during decode of custom metrics transformations"
				return err, false
			}
			if _, ok := uniqueDateFields[customMetricTransformation.DateField]; !ok && customMetric.TypeOfQuery == 1 {
				uniqueDateFields[customMetricTransformation.DateField] = customMetric.ObjectType
			}
		}
	}

	for dateField, objectType := range uniqueDateFields {
		if !*hardPull {
			if ok, _ := checkUsersFileExists(dateField, cloudManager, projectId, startTimestamp, startTimestamp, endTimestamp); ok {
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
		if err, ok := pullDataForUsers(projectId, cloudManager, diskManager, startTimestamp, endTimestamp, startTimestampInProjectTimezone, endTimestampInProjectTimezone, dateField, source, group, status, logCtx); !ok {
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

// pull users data for given datefield into files, and upload each local file to its proper cloud location
func pullDataForUsers(projectId int64, cloudManager *filestore.FileManager, diskManager *serviceDisk.DiskDriver, startTimestamp, endTimestamp, startTimestampInProjectTimezone, endTimestampInProjectTimezone int64,
	dateField string, source int, group int, status map[string]interface{}, logCtx *log.Entry) (error, bool) {

	logCtx.Infof("Pulling users for %s", dateField)

	// Writing users data to tmp file before upload.
	_, fName := diskManager.GetDailyUsersArchiveFilePathAndName(dateField, projectId, 0, startTimestamp, endTimestamp)
	startAt := time.Now().UnixNano()

	count, filePaths, err := pullUsersData(dateField, source, group, projectId, startTimestampInProjectTimezone, endTimestampInProjectTimezone, fName, diskManager, startTimestamp, endTimestamp)
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

	_, cName := (*cloudManager).GetDailyUsersArchiveFilePathAndName(dateField, projectId, 0, startTimestamp, endTimestamp)
	for timestamp, path := range filePaths {
		tmpOutputFile, err := os.Open(path)
		if err != nil {
			logCtx.WithField("error", err).Error("Failed to pull " + dateField + ". Write to tmp failed.")
			status["users-"+dateField+"-error"] = "Failed to pull " + dateField + ". Write to tmp failed."
			return err, false
		}

		cDir, _ := (*cloudManager).GetDailyUsersArchiveFilePathAndName(dateField, projectId, timestamp, 0, 0)

		err = (*cloudManager).Create(cDir, cName, tmpOutputFile)
		if err != nil {
			logCtx.WithField("error", err).Errorf("Failed to pull %s. Upload failed.", dateField)
			status["users-"+dateField+"-error"] = "Failed to pull " + dateField + ". Upload failed."
			return err, false
		}

		err = os.Remove(path)
		if err != nil {
			logCtx.Errorf("unable to delete file")
			return err, false
		}
	}

	return nil, true
}

// pull user rows from db and generate local files w.r.t timestamp and return map with (key,value) as (timestamp,path)
func pullUsersData(dateField string, source int, group int, projectID int64, startTimeTimezone, endTimeTimezone int64, fileName string, diskManager *serviceDisk.DiskDriver, startTimestamp, endTimestamp int64) (int, map[int64]string, error) {

	rows, tx, err := store.GetStore().PullUsersRowsForWIV2(projectID, startTimeTimezone, endTimeTimezone, dateField, source, group)
	if err != nil {
		log.WithError(err).Error("SQL Query failed.")
		return 0, nil, err
	}
	defer U.CloseReadQuery(rows, tx)

	var fileMap = make(map[int64]*os.File)
	var pathMap = make(map[int64]string)
	rowCount := 0
	for rows.Next() {
		var id string
		var properties *postgres.Jsonb
		var is_anonymous bool
		var join_timestamp int64
		var timestamp int
		if err = rows.Scan(&id, &properties, &is_anonymous, &join_timestamp, &timestamp); err != nil {
			log.WithFields(log.Fields{"err": err}).Error("SQL Parse failed.")
			return 0, nil, err
		}

		daysFromStart := int64(math.Floor(float64(int64(timestamp)-startTimeTimezone) / float64(U.Per_day_epoch)))
		fileTimestamp := startTimestamp + daysFromStart*U.Per_day_epoch
		file, ok := fileMap[fileTimestamp]
		if !ok {
			fPath, _ := diskManager.GetDailyUsersArchiveFilePathAndName(dateField, projectID, fileTimestamp, 0, 0)
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

		var propsMap map[string]interface{}
		if properties != nil {
			propsBytes, err := properties.Value()
			if err != nil {
				log.WithFields(log.Fields{"err": err}).Error("Unable to unmarshal properties")
				return 0, nil, err
			}
			err = json.Unmarshal(propsBytes.([]byte), &propsMap)
			if err != nil {
				log.WithFields(log.Fields{"err": err}).Error("Unable to unmarshal properties")
				return 0, nil, err
			}
		} else {
			log.WithFields(log.Fields{"err": err, "project_id": projectID}).Error("Nil properties")
		}

		user := CounterUserFormat{
			Id:            id,
			Properties:    propsMap,
			Is_Anonymous:  is_anonymous,
			JoinTimestamp: join_timestamp,
		}

		lineBytes, err := json.Marshal(user)
		if err != nil {
			log.WithFields(log.Fields{"err": err}).Error("Unable to marshal user.")
			return 0, nil, err
		}
		line := string(lineBytes)
		if _, err := file.WriteString(fmt.Sprintf("%s\n", line)); err != nil {
			log.WithFields(log.Fields{"line": line, "err": err}).Error("Unable to write to file.")
			return 0, nil, err
		}
		rowCount++
	}

	if rowCount > M.EventsPullLimit {
		// Todo(Dinesh): notify
		return rowCount, pathMap, fmt.Errorf("events count has exceeded the %d limit", M.EventsPullLimit)
	}

	return rowCount, pathMap, nil
}

// check if users file for given dateField has been generated already for given start and end timestamp, in dataTimestamp's day and
// its previous day's folder
func checkUsersFileExists(dateField string, cloudManager *filestore.FileManager, projectId int64, dataTimestamp, startTimestamp, endTimestamp int64) (bool, error) {
	path, name := (*cloudManager).GetDailyUsersArchiveFilePathAndName(dateField, projectId, dataTimestamp, startTimestamp, endTimestamp)
	if yes, _ := U.CheckFileExists(cloudManager, path, name); yes {
		return true, nil
	}
	path, _ = (*cloudManager).GetDailyUsersArchiveFilePathAndName(dateField, projectId, dataTimestamp-U.Per_day_epoch, startTimestamp, endTimestamp)
	if yes, _ := U.CheckFileExists(cloudManager, path, name); yes {
		return true, nil
	}
	return false, fmt.Errorf("file not found in cloud: %s", name)
}
