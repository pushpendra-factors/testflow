package memsql

import (
	"database/sql"
	C "factors/config"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

	"factors/model/model"
)

type UsersByEvent struct {
	User_id    string `json:"uid"`
	Event_name string `json:"en"`
	Timestamp  int64  `json:"ts"`
}

func (store *MemSQL) GetEventNameFromEventNameId(eventNameId string, projectId int64) (*model.EventName, error) {
	logFields := log.Fields{
		"project_id":    projectId,
		"event_name_id": eventNameId,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db
	var eventName model.EventName
	queryStr := "SELECT * FROM event_names WHERE id = ? AND project_id = ?"
	err := db.Raw(queryStr, eventNameId, projectId).Scan(&eventName).Error
	if err != nil {
		log.WithError(err).Error("Failed to get event_name from event_name_id")
		return nil, err
	}
	return &eventName, nil
}

// PullUserCohortDataOnEvent job - Function to pull users.
func (store *MemSQL) PullUserCohortDataOnEvent(projectID int64, startTime, endTime int64, event_id string, filter_property string) (*sql.Rows, error) {
	logFields := log.Fields{
		"project_id": projectID,
		"start_time": startTime,
		"end_time":   endTime,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db

	rawQuery := "SELECT usid as user_id,ucustid as user_customer_id,ugrpId as group_id,events.project_id, events.event_name_id, event_names.name as event_name,gcpp as group_count, CASE WHEN JSON_EXTRACT_STRING(events.user_properties, ? )=\"\" THEN -1 ELSE JSON_EXTRACT_STRING(events.user_properties, ? ) END as amount, events.timestamp from (SELECT users.id as usid , users.customer_user_id as ucustid, users.group_2_user_id as ugrpId, users.project_id , gc.pp as gcpp from users JOIN (SELECT group_2_user_id , count(group_2_user_id) as pp, project_id from users where project_id=? group by group_2_user_id) as gc on gc.group_2_user_id=users.group_2_user_id) as gcp JOIN events ON events.user_id=gcp.usid JOIN event_names ON event_names.id = events.event_name_id where events.project_id=? AND events.event_name_id = ? and events.timestamp between  ? AND ? ORDER BY events.timestamp"
	rows, err := db.Raw(rawQuery, filter_property, filter_property, projectID, projectID, event_id, startTime, endTime).Rows()
	if err != nil {
		return nil, err
	}

	return rows, nil
}

// GetGroupsOnEvent job - Function to pull events for archival.
func (store *MemSQL) GetGroupsOnEvent(projectID int64, event_name string) (*sql.Rows, *sql.Tx, error) {
	logFields := log.Fields{
		"project_id": projectID,
		"event_name": event_name,
	}
	limit := 1
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	rawQuery := fmt.Sprintf("select * from groups where project_id = %d  and name=%s limit %d", projectID, event_name, limit)

	rows, tx, err, _ := store.ExecQueryWithContext(rawQuery, []interface{}{})
	return rows, tx, err
}

// GetGroupsOnEvent job - Function to pull events for archival.
func (store *MemSQL) GetUsersTimestampOnFirstEvent(projectID int64, event_name_id string, start_time int64, end_time int64) (map[string]int64, error) {
	usersId := make(map[string]int64)
	logFields := log.Fields{
		"project_id": projectID,
		"event_name": event_name_id,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	db := C.GetServices().Db
	queryStmnt := "SELECT  user_id,event_name_id,min(timestamp) FROM events where project_id= ? AND event_name_id=? AND timestamp BETWEEN ? AND ? GROUP BY user_id ORDER BY timestamp"
	rows, err := db.Raw(queryStmnt, projectID, event_name_id, start_time, end_time).Rows()
	if err != nil {
		log.WithError(err).Error("Failed to get users along with first event.")
		return nil, fmt.Errorf("Failed to get users along with first event")
	}
	defer rows.Close()

	for rows.Next() {
		var evt UsersByEvent
		if err = db.ScanRows(rows, &evt); err != nil {
			log.Errorf("unable to read row from DB")
			return nil, err
		}
		usersId[evt.User_id] = evt.Timestamp
	}

	return usersId, err
}

// GetGroupsOnEvent job - Function to pull events for archival.
func (store *MemSQL) GetUsersEventTimeStampFromHistory(projectID int64, event_name_id string, userIdFiltered map[string]int64) (map[string]int64, error) {
	usersIdList := make([]string, 0)

	usersIdMap := make(map[string]int64, 0)
	logFields := log.Fields{
		"project_id": projectID,
		"event_name": event_name_id,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	for k, _ := range userIdFiltered {
		usersIdList = append(usersIdList, k)
	}

	db := C.GetServices().Db
	queryStmnt := "SELECT  user_id,event_name_id,min(timestamp) FROM events where project_id= ? AND event_name_id=? AND user_id IN (?) GROUP BY user_id ORDER BY timestamp"
	rows, err := db.Raw(queryStmnt, projectID, event_name_id, usersIdList).Rows()
	if err != nil {
		log.WithError(err).Error("Failed to get users along with first event.")
		return nil, fmt.Errorf("Failed to get users along with first event")
	}
	defer rows.Close()

	for rows.Next() {
		var evt UsersByEvent
		if err = db.ScanRows(rows, &evt); err != nil {
			log.Errorf("unable to read row from DB")
			return nil, err
		}
		usersIdMap[evt.User_id] = evt.Timestamp
	}

	return usersIdMap, err
}

// GetGroupsOnEvent job - Function to pull events.
func (store *MemSQL) GetAllEventsWithUsers(projectID int64, event_name string, start_time int64, end_time int64) (*sql.Rows, error) {
	logFields := log.Fields{
		"project_id": projectID,
		"event_name": event_name,
		"start_time": start_time,
		"end_time":   end_time,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	db := C.GetServices().Db
	queryStmnt := "select user_id,event_name_id, event_names.name as event_name,timestamp,users.is_group_user as is_group_user,users.group_2_user_id as group_id from events JOIN users ON events.user_id=users.id JOIN event_names ON event_names.id = events.event_name_id where events.project_id = ? and events.event_name_id= ? AND events.timestamp BETWEEN ? AND ? GROUP BY events.user_id ORDER BY events.timestamp"
	rows, err := db.Raw(queryStmnt, projectID, event_name, start_time, end_time).Rows()
	if err != nil {
		log.WithError(err).Error("Failed to get users on event.")
		return nil, fmt.Errorf("Failed to get users on event")
	}
	return rows, err
}

// GetGroupsOnEvent job - Function to pull events.
func (store *MemSQL) GetBaseEventsOnUsers(projectID int64, event_name_id string, start_time int64, end_time int64, users []string) (*sql.Rows, error) {
	logFields := log.Fields{
		"project_id": projectID,
		"event_name": event_name_id,
		"start_time": start_time,
		"end_time":   end_time,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	log.Infof("Total number of users to query:%d", len(users))
	db := C.GetServices().Db

	queryStmnt := "SELECT COALESCE(users.customer_user_id, users.id), event_names.name , events.timestamp, events.count , events.properties , users.join_timestamp, events.user_properties, COALESCE(users.group_2_user_id,\"\") as group_id FROM events LEFT JOIN event_names ON events.event_name_id = event_names.id LEFT JOIN users ON events.user_id = users.id AND users.project_id = ? WHERE events.project_id = ? AND events.event_name_id = ? AND events.user_id IN ( ? ) AND events.timestamp BETWEEN ? AND ? ORDER BY COALESCE(users.customer_user_id, users.id), events.timestamp"
	rows, err := db.Raw(queryStmnt, projectID, projectID, event_name_id, users, start_time, end_time).Rows()

	if err != nil {
		log.WithError(err).Error("Failed to get users on event.")
		return nil, fmt.Errorf("Failed to get users on event")
	}
	return rows, err
}

// GetGroupsOnEvent job - Function to pull events for archival.
func (store *MemSQL) GetAllEventsOnUsersBetweenTime(projectID int64, users []string, start_time int64, end_time int64) (*sql.Rows, *sql.Tx, error) {
	var rawQuery string
	logFields := log.Fields{
		"project_id": projectID,
		"start_time": start_time,
		"end_time":   end_time,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	rawQuery = fmt.Sprintf("select user_id,event_name_id,timestamp from events where project_id = %d  and  AND events.timestamp BETWEEN %d AND %d", projectID, start_time, end_time)

	rows, tx, err, _ := store.ExecQueryWithContext(rawQuery, []interface{}{})
	return rows, tx, err
}

func (store *MemSQL) GetAllEventsOnUsers(projectId int64, arrayCustomerUserID []string) (*sql.Rows, error) {
	logFields := log.Fields{
		"project_id":             projectId,
		"array_customer_user_id": arrayCustomerUserID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	if len(arrayCustomerUserID) == 0 {
		return nil, fmt.Errorf("num of customer_id is null")
	}

	db := C.GetServices().Db
	queryStmnt := "SELECT events.user_id,events.event_name_id,event_names.name as event_name,events.timestamp FROM events JOIN event_names ON " +
		"event_names.project_id=events.project_id AND event_names.id=events.event_name_id AND events.project_id = ? AND " +
		"event_names.project_id = ? AND events.user_id IN ( ? )"
	rows, err := db.Raw(queryStmnt, projectId, projectId, arrayCustomerUserID).Rows()
	if err != nil {
		log.WithError(err).Error("Failed to get customer_user_id.")
		return nil, fmt.Errorf("Failed to get customer_user_id")
	}
	return rows, nil
}

func (store *MemSQL) GetCountOfGroupIDS(projectId int64, arrayGroupID []string) (*sql.Rows, error) {
	logFields := log.Fields{
		"project_id": projectId,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	if len(arrayGroupID) == 0 {
		return nil, fmt.Errorf("num of customer_id is 0")
	}

	db := C.GetServices().Db
	queryStmnt := "SELECT project_id,group_2_user_id as group_id,COUNT(group_2_user_id) as group_count FROM users WHERE project_id = ? AND users.group_2_user_id IN ( ? ) GROUP BY group_2_user_id"
	rows, err := db.Raw(queryStmnt, projectId, arrayGroupID).Rows()
	if err != nil {
		log.WithError(err).Error("Failed group_id with count.")
		return nil, err
	}

	return rows, nil
}

func (store *MemSQL) PullEventRowsOnUsers(projectID int64, users []string, start_time, end_time int64) (*sql.Rows, error) {
	logFields := log.Fields{
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	db := C.GetServices().Db

	queryStmnt := "SELECT COALESCE(users.customer_user_id, users.id), event_names.name , events.timestamp, events.count , events.properties , users.join_timestamp, events.user_properties FROM events LEFT JOIN event_names ON events.event_name_id = event_names.id LEFT JOIN users ON events.user_id = users.id AND users.project_id = ? WHERE events.project_id = ? AND COALESCE(users.customer_user_id, users.id) IN ( ? )  AND events.timestamp <= ?  ORDER BY COALESCE(users.customer_user_id, users.id), events.timestamp"

	rows, err := db.Raw(queryStmnt, projectID, projectID, users, end_time).Rows()
	if err != nil {
		log.WithError(err).Error("Failed to events of all users.")
		return nil, fmt.Errorf("Failed to get all events of users")
	}

	return rows, err
}
