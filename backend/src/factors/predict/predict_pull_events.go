package predict

import (
	"bufio"
	"encoding/json"
	C "factors/config"
	"factors/filestore"
	"factors/model/store"
	P "factors/pattern"
	serviceDisk "factors/services/disk"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

type EventWithGroup struct {
	Event    P.CounterEventFormat `json:"ev"`
	Group_id string               `json:"gid"`
}

func MakeTimestamps(prj *PredictProject) error {
	// to be addded to manipulate start and end time

	return nil
}

func GetEventNameID(project_id int64, event_name string) (string, error) {
	event_details, err := store.GetStore().GetEventNameIDFromEventName(event_name, project_id)
	if err != nil {
		prLog.WithError(err).Error("SQL Query failed.")
		return "", err
	}
	return event_details.ID, nil
}

func getTime(baseTime int64, num_week_back int, buffer_time_in_hrs int64) int64 {
	var one_day_epoch int64 = 604800
	time := (baseTime - int64(num_week_back)*one_day_epoch)
	return time
}

func GetAllUsersWithGroupsByCohort(project_id int64, model_id int64, event string, start_time, end_time int64, filter_prop string,
	diskManager *serviceDisk.DiskDriver, cloudManager *filestore.FileManager) (map[string]int, error) {

	usersId := make(map[string]int)
	projectDataPath := diskManager.GetPredictProjectDataPath(project_id, model_id)
	dataFileName := "users.txt"
	usersFilePath := filepath.Join(projectDataPath, dataFileName)
	f, err := os.Create(usersFilePath)
	if err != nil {
		return nil, fmt.Errorf("unable to create events file for predict")
	}
	defer f.Close()

	if end_time < start_time {
		return nil, fmt.Errorf("start time greater than end time")
	}

	event_id, err := GetEventNameID(project_id, event)
	if err != nil {
		return nil, err
	}

	user_details_rows, err := store.GetStore().PullUserCohortDataOnEvent(project_id, start_time, end_time, event_id, filter_prop)
	if err != nil {
		prLog.Errorf("Error executing pull users for predict jobs query")
	}
	defer user_details_rows.Close()

	db := C.GetServices().Db
	lineNum := 0
	for user_details_rows.Next() {
		var user_meta UsersCohort
		if err = db.ScanRows(user_details_rows, &user_meta); err != nil {
			prLog.WithFields(log.Fields{"err": err}).Error("SQL Parse failed.")
			return nil, err
		}
		usersId[user_meta.User_id] = +1
		usersId[user_meta.User_customer_id] += 1
		userByte, err := json.Marshal(user_meta)
		if err != nil {
			return nil, fmt.Errorf("unable to marhall event :%v", err)
		}
		f.WriteString(string(userByte) + "\n")
		lineNum++
	}
	err = user_details_rows.Err()
	if err != nil {
		return nil, fmt.Errorf("error in pulling users cohort data:%v", err)
	}

	fre, err := os.OpenFile(usersFilePath, os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}
	r := bufio.NewReader(fre)
	dataPath := (*cloudManager).GetPredictProjectDataPath(project_id, model_id)
	err = (*cloudManager).Create(dataPath, dataFileName, r)
	if err != nil {
		return nil, err
	}
	defer fre.Close()
	prLog.Infof("events file written to GCP")

	return usersId, nil
}

func GetUsersWithEventAndFirstTimestamp(project_id int64, event_id string, start_time, end_time int64) (map[string]int64, error) {

	usersId, err := store.GetStore().GetUsersTimestampOnFirstEvent(project_id, event_id, start_time, end_time)
	if err != nil {
		e := fmt.Errorf("unabel to get all users by event and even's first timestamp :%v", err)
		return nil, e
	}
	return usersId, nil
}

func GetUsersEventsFirstTimestampFromHistory(project_id int64, event_id string, userIdFiltered map[string]int64) (map[string]int64, error) {

	usersId, err := store.GetStore().GetUsersEventTimeStampFromHistory(project_id, event_id, userIdFiltered)
	if err != nil {
		e := fmt.Errorf("unabel to get all users by event and even's first timestamp :%v", err)
		return nil, e
	}
	return usersId, nil
}

func GetAllBaseEvents(project_id int64, model_id int64, event_id string, start_time int64, end_time int64,
	diskManager *serviceDisk.DiskDriver, cloudManager *filestore.FileManager) (map[string]int64, map[string]int64, error) {

	prLog.Info("Getting all base events with group ids of users ")
	users_map := make(map[string]int64, 0)
	group_id_map := make(map[string]int64, 0)
	dataFileName := "base_events.txt"
	// db := C.GetServices().Db

	projectDataPath := diskManager.GetPredictProjectDataPath(project_id, model_id)
	eventsFilePath := filepath.Join(projectDataPath, dataFileName)
	f, err := os.Create(eventsFilePath)
	defer f.Close()

	if err != nil {
		return nil, nil, fmt.Errorf("unable to create events file for predict")
	}

	userIDMapRaw, err := GetUsersWithEventAndFirstTimestamp(project_id, event_id, start_time, end_time)
	if err != nil {
		return nil, nil, err
	}

	userIDMapFiltered, err := GetUsersEventsFirstTimestampFromHistory(project_id, event_id, userIDMapRaw)
	if err != nil {
		return nil, nil, err
	}

	userIDPos := make([]string, 0)
	userIDneg := make([]string, 0)

	for userID, eventTSFilt := range userIDMapFiltered {

		if ts, ok := userIDMapRaw[userID]; ok {
			if eventTSFilt < ts {
				userIDneg = append(userIDneg, userID)
			} else {
				userIDPos = append(userIDPos, userID)
			}
		}
	}
	prLog.Infof("pos user list :%v", userIDPos)
	prLog.Infof("neg user list :%v", userIDneg)

	prLog.Infof("Total number of pos users :%d", len(userIDPos))
	rows, err := store.GetStore().GetBaseEventsOnUsers(project_id, event_id, start_time, end_time, userIDPos)
	defer rows.Close()
	if err != nil {
		e := fmt.Errorf("unable to get all users by event and even's first timestamp :%v", err)
		return nil, nil, e
	}

	rowCount := 0
	nilUserProperties := 0
	lineNum := 0
	dataPull := 0
	for rows.Next() {
		var userID string
		var eventName string
		var eventTimestamp int64
		var userJoinTimestamp int64
		var eventCardinality uint
		var eventProperties *postgres.Jsonb
		var userProperties *postgres.Jsonb
		var group_id string
		var eg EventWithGroup

		if err = rows.Scan(&userID, &eventName, &eventTimestamp, &eventCardinality, &eventProperties, &userJoinTimestamp, &userProperties, &group_id); err != nil {
			prLog.WithFields(log.Fields{"err": err}).Error("SQL Parse failed.")
			return nil, nil, err
		}
		dataPull++
		// prLog.Infof("%v", tmpEvents)
		var eventPropertiesMap map[string]interface{}
		if eventProperties != nil {
			eventPropertiesBytes, err := eventProperties.Value()
			if err != nil {
				prLog.WithFields(log.Fields{"err": err}).Error("Unable to unmarshal event property.")
				return nil, nil, err
			}
			err = json.Unmarshal(eventPropertiesBytes.([]byte), &eventPropertiesMap)
			if err != nil {
				prLog.WithFields(log.Fields{"err": err}).Error("Unable to unmarshal event property.")
				return nil, nil, err
			}
		} else {
			prLog.WithFields(log.Fields{"err": err, "project_id": project_id}).Error("Nil event properties.")
		}

		var userPropertiesMap map[string]interface{}
		if userProperties != nil {
			userPropertiesBytes, err := userProperties.Value()
			if err != nil {
				prLog.WithFields(log.Fields{"err": err}).Error("Unable to unmarshal user property.")
				return nil, nil, err
			}

			err = json.Unmarshal(userPropertiesBytes.([]byte), &userPropertiesMap)
			if err != nil {
				prLog.WithFields(log.Fields{"err": err}).Error("Unable to unmarshal user property.")
				return nil, nil, err
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
		eg.Event = event
		eg.Group_id = group_id
		group_id_map[eg.Group_id] += 1
		users_map[userID] = eventTimestamp
		lineBytes, err := json.Marshal(eg)
		if err != nil {
			prLog.WithFields(log.Fields{"err": err}).Error("Unable to unmarshal event.")
			return nil, nil, err
		}
		line := string(lineBytes)
		if _, err := f.WriteString(fmt.Sprintf("%s\n", line)); err != nil {
			prLog.WithFields(log.Fields{"line": line, "err": err}).Error("Unable to write to file.")
			return nil, nil, err
		}
		lineNum++
		rowCount++
	}

	err = rows.Err()
	if err != nil {
		return nil, nil, fmt.Errorf("error fetching rows :%v", err)
	}
	prLog.Infof("number of base events written to file :%d, %d", lineNum, dataPull)

	beFile, err := os.OpenFile(eventsFilePath, os.O_RDWR, 0666)
	if err != nil {
		return nil, nil, err
	}
	r := bufio.NewReader(beFile)
	dataPath := (*cloudManager).GetPredictProjectDataPath(project_id, model_id)
	err = (*cloudManager).Create(dataPath, dataFileName, r)
	if err != nil {
		return nil, nil, err
	}
	defer beFile.Close()
	prLog.Infof("events file written to GCP")

	return users_map, group_id_map, nil

}

func GetGroupCountsOfGroupIds(project_id int64, model_id int64, group_id_map map[string]int64,
	diskManager *serviceDisk.DiskDriver, cloudManager *filestore.FileManager) error {

	group_ids := make([]string, 0)
	dataFileName := "groups.txt"

	for k, _ := range group_id_map {
		group_ids = append(group_ids, k)
	}

	prLog.Infof("Getting all group ids and counts of users :%d", len(group_ids))

	db := C.GetServices().Db

	projectDataPath := diskManager.GetPredictProjectDataPath(project_id, model_id)
	eventsFilePath := filepath.Join(projectDataPath, dataFileName)
	f, err := os.Create(eventsFilePath)
	defer f.Close()

	if err != nil {
		return fmt.Errorf("unable to create group file for predict")
	}

	rows, err := store.GetStore().GetCountOfGroupIDS(project_id, group_ids)
	if err != nil {
		e := fmt.Errorf("Unable to get all users by event and even's first timestamp :%v", err)
		return e
	}
	defer rows.Close()

	lineNum := 0
	dataPull := 0
	for rows.Next() {
		var evt GroupIDCounts
		if err = db.ScanRows(rows, &evt); err != nil {
			log.Errorf("unable to read row from DB")
			return err
		}
		dataPull += 1
		eventByte, err := json.Marshal(evt)
		if err != nil {
			return fmt.Errorf("unable to marhall event :%v", err)
		}
		f.WriteString(string(eventByte) + "\n")
		lineNum++
	}
	err = rows.Err()
	if err != nil {
		return fmt.Errorf("error fetching rows :%v", err)
	}
	prLog.Infof("number of groups written to file :%d, %d", lineNum, dataPull)

	fre, err := os.OpenFile(eventsFilePath, os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	r := bufio.NewReader(fre)
	dataPath := (*cloudManager).GetPredictProjectDataPath(project_id, model_id)
	err = (*cloudManager).Create(dataPath, dataFileName, r)
	if err != nil {
		return err
	}
	defer fre.Close()
	prLog.Infof("groups file written to GCP")

	return nil

}

func GetEventsOfUsers(project_id int64, model_id int64, usersWithTS map[string]int64, usersOntarget map[string]int,
	start_time int64, end_time int64, buffer_time int64,
	diskManager *serviceDisk.DiskDriver, cloudManager *filestore.FileManager) error {

	users := make([]string, 0)
	dataFileName := "all_events.txt"

	userOnstartEnd := make(map[string]int64)

	for k, v := range usersWithTS {
		userOnstartEnd[k] = v
	}
	prLog.Infof(" total number of users on target (to pull events):%d", len(userOnstartEnd))

	num_unique_on_target := 0
	for k, _ := range usersOntarget {
		if _, ok := userOnstartEnd[k]; !ok {
			userOnstartEnd[k] = start_time + buffer_time
			num_unique_on_target += 1
		}
	}
	prLog.Infof(" total number of uniquer users on start (to pull events):%d", num_unique_on_target)

	for k, _ := range userOnstartEnd {
		users = append(users, k)
	}
	prLog.Infof(" total number of users (to pull events):%d", len(users))
	projectDataPath := diskManager.GetPredictProjectDataPath(project_id, model_id)
	eventsFilePath := filepath.Join(projectDataPath, dataFileName)
	f, err := os.Create(eventsFilePath)
	defer f.Close()

	if err != nil {
		return fmt.Errorf("unable to create events file for predict")
	}

	// db := C.GetServices().Db

	users_pulled := make(map[string]int, 0)
	users_pull_in_time_range := make(map[string]int)
	rows, err := store.GetStore().PullEventRowsOnUsers(project_id, users, start_time, end_time)
	defer rows.Close()
	if err != nil {
		e := fmt.Errorf("unabel to get all users by event and even's first timestamp :%v", err)
		return e
	}
	rowCount := 0
	nilUserProperties := 0
	lineNum := 0
	dataPull := 0
	for rows.Next() {
		var userID string
		var eventName string
		var eventTimestamp int64
		var userJoinTimestamp int64
		var eventCardinality uint
		var eventProperties *postgres.Jsonb
		var userProperties *postgres.Jsonb

		if err = rows.Scan(&userID, &eventName, &eventTimestamp, &eventCardinality, &eventProperties, &userJoinTimestamp, &userProperties); err != nil {
			prLog.WithFields(log.Fields{"err": err}).Error("SQL Parse failed.")
			return err
		}
		dataPull++
		// prLog.Infof("%v", tmpEvents)
		var eventPropertiesMap map[string]interface{}
		if eventProperties != nil {
			eventPropertiesBytes, err := eventProperties.Value()
			if err != nil {
				prLog.WithFields(log.Fields{"err": err}).Error("Unable to unmarshal event property.")
				return err
			}
			err = json.Unmarshal(eventPropertiesBytes.([]byte), &eventPropertiesMap)
			if err != nil {
				prLog.WithFields(log.Fields{"err": err}).Error("Unable to unmarshal event property.")
				return err
			}
		} else {
			prLog.WithFields(log.Fields{"err": err, "project_id": project_id}).Error("Nil event properties.")
		}

		var userPropertiesMap map[string]interface{}
		if userProperties != nil {
			userPropertiesBytes, err := userProperties.Value()
			if err != nil {
				prLog.WithFields(log.Fields{"err": err}).Error("Unable to unmarshal user property.")
				return err
			}

			err = json.Unmarshal(userPropertiesBytes.([]byte), &userPropertiesMap)
			if err != nil {
				prLog.WithFields(log.Fields{"err": err}).Error("Unable to unmarshal user property.")
				return err
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

		if ts, ok := userOnstartEnd[userID]; ok {
			if eventTimestamp <= (ts + buffer_time) { // condition to check events less than timestamp of first_base event + buffer_time
				lineBytes, err := json.Marshal(event)
				if err != nil {
					prLog.WithFields(log.Fields{"err": err}).Error("Unable to unmarshal event.")
					return err
				}
				line := string(lineBytes)
				if _, err := f.WriteString(fmt.Sprintf("%s\n", line)); err != nil {
					prLog.WithFields(log.Fields{"line": line, "err": err}).Error("Unable to write to file.")
					return err
				}
				users_pull_in_time_range[event.UserId] += 1
				lineNum++
			}
		}
		rowCount++
		users_pulled[userID] += 1

	}

	err = rows.Err()
	if err != nil {
		return fmt.Errorf("error fetching rows :%v", err)
	}

	prLog.Infof("Total number of users (after events pull):%d , in time range : %d", len(users_pulled), len(users_pull_in_time_range))
	prLog.Infof("number of events written to file :%d, %d , %d", lineNum, dataPull, rowCount)

	fre, err := os.OpenFile(eventsFilePath, os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	r := bufio.NewReader(fre)
	dataPath := (*cloudManager).GetPredictProjectDataPath(project_id, model_id)
	err = (*cloudManager).Create(dataPath, dataFileName, r)
	if err != nil {
		return err
	}
	defer fre.Close()
	prLog.Infof("events file written to GCP")

	return nil

}
