package model

import (
	"database/sql"
	"encoding/json"
	"errors"
	cacheRedis "factors/cache/redis"
	C "factors/config"
	U "factors/util"
	"fmt"
	"net/http"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

type Event struct {
	// Composite primary key with project_id and uuid.
	ID              string  `gorm:"primary_key:true;type:uuid;default:uuid_generate_v4()" json:"id"`
	CustomerEventId *string `json:"customer_event_id"`

	// Below are the foreign key constraints added in creation script.
	// project_id -> projects(id)
	// (project_id, user_id) -> users(project_id, id)
	// (project_id, event_name_id) -> event_names(project_id, id)
	ProjectId        uint64  `gorm:"primary_key:true;" json:"project_id"`
	UserId           string  `json:"user_id"`
	UserPropertiesId string  `json:"user_properties_id"`
	SessionId        *string `json:session_id`
	EventNameId      uint64  `json:"event_name_id"`
	Count            uint64  `json:"count"`
	// JsonB of postgres with gorm. https://github.com/jinzhu/gorm/issues/1183
	Properties postgres.Jsonb `json:"properties,omitempty"`
	// unix epoch timestamp in seconds.
	Timestamp int64     `json:"timestamp"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ProjectEventsInfo struct {
	ProjectId           uint64 `json:"project_id"`
	ProjectName         string `json:"project_name"`
	EventsCount         int    `json:"events_count"`
	CreatorEmail        string `json:"creator_email"`
	FirstEventTimestamp int64  `json:"-"`
	LastEventTimestamp  int64  `json:"-"`
}

type CacheEvent struct {
	ID        string `json:"id"`
	Timestamp int64  `json:"ts"`
}

const error_Duplicate_event_customerEventID = "pq: duplicate key value violates unique constraint \"project_id_customer_event_id_unique_idx\""
const eventsLimitForProperites = 50000
const NewUserSessionInactivityDuration = time.Minute * 30

const tableName = "events"
const cacheIndexUserLastEvent = "user_last_event"

func isDuplicateCustomerEventIdError(err error) bool {
	return err.Error() == error_Duplicate_event_customerEventID
}

func getUserLastEventCacheKey(projectId uint64, userId string) (*cacheRedis.Key, error) {
	suffix := fmt.Sprintf("uid:%s", userId)
	prefix := fmt.Sprintf("%s:%s", tableName, cacheIndexUserLastEvent)
	return cacheRedis.NewKey(projectId, prefix, suffix)
}

func SetCacheUserLastEvent(projectId uint64, userId string, cacheEvent *CacheEvent) error {
	logCtx := log.WithField("project_id", projectId).WithField("user_id", userId)
	if projectId == 0 || userId == "" {
		logCtx.Error("Invalid project or user id on addToCacheUserLastEventTimestamp")
		return errors.New("invalid project or user id")
	}

	if cacheEvent == nil {
		logCtx.Error("Nil cache event on setCacheUserLastEvent")
		return errors.New("nil cache event")
	}

	cacheEventJson, err := json.Marshal(cacheEvent)
	if err != nil {
		logCtx.Error("Failed cache event json marshal.")
		return err
	}

	key, err := getUserLastEventCacheKey(projectId, userId)
	if err != nil {
		return err
	}

	// Expiry: Session inactivity duration + 5 minutes.
	cacheExpiry := (NewUserSessionInactivityDuration + (time.Minute * 5)).Seconds()
	err = cacheRedis.Set(key, string(cacheEventJson), cacheExpiry)
	if err != nil {
		logCtx.WithError(err).Error("Failed to setCacheUserLastEvent.")
	}

	return err
}

func CreateEvent(event *Event) (*Event, int) {
	db := C.GetServices().Db

	log.WithFields(log.Fields{"event": &event}).Info("Creating event")

	// Input Validation. (ID is to be auto generated)
	if event.ID != "" {
		log.Error("CreateEvent Failed. Id provided.")
		return nil, http.StatusBadRequest
	}

	if event.ProjectId == 0 || event.UserId == "" {
		log.Error("CreateEvent Failed. Invalid projectId or userId.")
		return nil, http.StatusBadRequest
	}

	// Increamenting count based on EventNameId, not by EventName.
	var count uint64
	if err := db.Model(&Event{}).Where("project_id = ? AND user_id = ? AND event_name_id = ?",
		event.ProjectId, event.UserId, event.EventNameId).Count(&count).Error; err != nil {
		return nil, http.StatusInternalServerError
	}
	event.Count = count + 1

	if event.Timestamp <= 0 {
		event.Timestamp = time.Now().Unix()
	}

	eventPropsJSONb, err := U.FillHourAndDayEventProperty(&event.Properties, event.Timestamp)
	if err != nil {
		log.WithFields(log.Fields{"projectId": event.ProjectId,
			"eventTimestamp": event.Timestamp}).WithError(err).Error(
			"Adding day of week and hour of day properties failed")
	}
	event.Properties = *eventPropsJSONb

	transTime := gorm.NowFunc()
	rows, err := db.Raw("INSERT INTO events (customer_event_id,project_id,user_id,user_properties_id,session_id,event_name_id,count,properties,timestamp,created_at,updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) RETURNING events.id",
		event.CustomerEventId, event.ProjectId, event.UserId, event.UserPropertiesId, event.SessionId, event.EventNameId, event.Count, event.Properties, event.Timestamp, transTime, transTime).Rows()
	if err != nil {
		if isDuplicateCustomerEventIdError(err) {
			log.WithError(err).Info("CreateEvent Failed, duplicate customerEventId")
			return nil, http.StatusFound
		}

		log.WithFields(log.Fields{"event": &event}).WithError(err).Error("CreateEvent Failed")
		return nil, http.StatusInternalServerError
	}

	var eventId string
	for rows.Next() {
		if err = rows.Scan(&eventId); err != nil {
			log.WithError(err).Error("CreateEvent Failed. Failed to read event id.")
			return nil, http.StatusInternalServerError
		}
	}
	event.ID = eventId
	event.CreatedAt = transTime
	event.UpdatedAt = transTime

	SetCacheUserLastEvent(event.ProjectId, event.UserId,
		&CacheEvent{ID: event.ID, Timestamp: event.Timestamp})
	return event, http.StatusCreated
}

func GetEvent(projectId uint64, userId string, id string) (*Event, int) {
	db := C.GetServices().Db

	var event Event
	if err := db.Where("id = ?", id).Where("project_id = ?", projectId).Where("user_id = ?", userId).First(&event).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}
		return nil, http.StatusInternalServerError
	}
	return &event, http.StatusFound
}

func GetEventById(projectId uint64, id string) (*Event, int) {
	db := C.GetServices().Db

	var event Event
	if err := db.Where("project_id = ?", projectId).Where("id = ?", id).First(&event).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}
		return nil, http.StatusInternalServerError
	}
	return &event, http.StatusFound
}

func GetCacheUserLastEvent(projectId uint64, userId string) (*CacheEvent, error) {
	key, err := getUserLastEventCacheKey(projectId, userId)
	if err != nil {
		return nil, err
	}

	cacheEventJson, err := cacheRedis.Get(key)
	if err != nil {
		return nil, err
	}

	var cacheEvent CacheEvent
	err = json.Unmarshal([]byte(cacheEventJson), &cacheEvent)
	if err != nil {
		return nil, err
	}

	return &cacheEvent, nil
}

func GetLatestAnyEventOfUserInDurationFromCache(projectId uint64, userId string, duration time.Duration) (*Event, int) {
	logCtx := log.WithFields(log.Fields{"project_id": projectId, "user_id": userId})

	cacheEvent, err := GetCacheUserLastEvent(projectId, userId)
	if err != nil {
		if err == redis.ErrNil {
			return nil, http.StatusNotFound
		}

		logCtx.Error("Failed to get latest event of user in duration from cache.")
		return nil, http.StatusInternalServerError
	}

	// cached event older than given duration.
	if cacheEvent.Timestamp < U.UnixTimeBeforeDuration(duration) {
		return nil, http.StatusNotFound
	}

	event, errCode := GetEventById(projectId, cacheEvent.ID)
	if errCode != http.StatusFound {
		logCtx.WithField("err_code", errCode).Error("Failed to get event on using id from cache.")
		return nil, errCode
	}

	return event, http.StatusFound
}

func GetLatestAnyEventOfUserInDuration(projectId uint64, userId string, duration time.Duration) (*Event, int) {
	db := C.GetServices().Db

	if duration == 0 {
		return nil, http.StatusBadRequest
	}

	var events []Event
	if err := db.Limit(1).Order("timestamp desc").Where("project_id = ? AND user_id = ? AND timestamp > ?",
		projectId, userId, U.UnixTimeBeforeDuration(duration)).Find(&events).Error; err != nil {

		return nil, http.StatusInternalServerError
	}

	if len(events) == 0 {
		return nil, http.StatusNotFound
	}

	return &events[0], http.StatusFound
}

func GetLatestEventOfUserByEventNameId(projectId uint64, userId string, eventNameId uint64,
	startTimestamp, endTimestamp int64) (*Event, int) {

	db := C.GetServices().Db

	if startTimestamp == 0 || endTimestamp == 0 {
		return nil, http.StatusBadRequest
	}

	var events []Event
	if err := db.Limit(1).Order("timestamp desc").Where(
		"project_id = ? AND event_name_id = ? AND user_id = ? AND timestamp > ? AND timestamp <= ?",
		projectId, eventNameId, userId, startTimestamp, endTimestamp).Find(&events).Error; err != nil {

		return nil, http.StatusInternalServerError
	}

	if len(events) == 0 {
		return nil, http.StatusNotFound
	}

	return &events[0], http.StatusFound
}

func GetProjectEventsInfo() (*(map[uint64]*ProjectEventsInfo), int) {
	db := C.GetServices().Db

	queryStr := "SELECT events_info.*, agents.email FROM" +
		" " + "(SELECT projects.id, projects.name, min(events.timestamp) as first_timestamp, max(events.timestamp) as last_timestamp, count(*) as events_count FROM events" +
		" " + "LEFT JOIN projects on events.project_id = projects.id GROUP BY projects.id) as events_info" +
		" " + "LEFT JOIN project_agent_mappings ON project_agent_mappings.project_id=events_info.id AND project_agent_mappings.role=2" +
		" " + "LEFT JOIN agents ON project_agent_mappings.agent_uuid=agents.uuid ORDER BY events_info.id"

	rows, err := db.Raw(queryStr).Rows()
	if err != nil {
		log.WithError(err).Error("Failed to get events timestamp info.")
		return nil, http.StatusInternalServerError
	}
	defer rows.Close()

	projectEventsTime := make(map[uint64]*ProjectEventsInfo, 0)

	count := 0
	for rows.Next() {
		var projectId uint64
		var firstTimestamp, lastTimestamp int64
		var projectName string
		var creatorEmail sql.NullString
		var eventsCount int
		if err = rows.Scan(&projectId, &projectName, &firstTimestamp, &lastTimestamp, &eventsCount, &creatorEmail); err != nil {
			log.Error(err)
			return nil, http.StatusInternalServerError
		}

		if firstTimestamp > 0 {
			projectEventsTime[projectId] = &ProjectEventsInfo{ProjectId: projectId, FirstEventTimestamp: firstTimestamp,
				LastEventTimestamp: lastTimestamp, ProjectName: projectName, CreatorEmail: creatorEmail.String, EventsCount: eventsCount}
		}

		count++
	}

	if count == 0 {
		return nil, http.StatusNotFound
	}

	return &projectEventsTime, http.StatusFound
}

// GetRecentEventPropertyKeys - Returns unique event property
// keys from last 24 hours.
func GetRecentEventPropertyKeysWithLimits(projectId uint64, eventName string, eventsLimit int) (map[string][]string, int) {
	db := C.GetServices().Db

	eventsAfterTimestamp := U.UnixTimeBeforeDuration(24 * time.Hour)
	logCtx := log.WithFields(log.Fields{"project_id": projectId, "events_after_timestamp": eventsAfterTimestamp})

	queryStr := "SELECT distinct(properties) AS keys FROM events WHERE project_id = ?" +
		" " + "AND event_name_id IN (SELECT id FROM event_names WHERE project_id = ? AND name = ?)" +
		" " + "AND timestamp > ? AND properties != 'null' LIMIT ?"

	rows, err := db.Raw(queryStr, projectId, projectId, eventName, eventsAfterTimestamp, eventsLimit).Rows()
	if err != nil {
		logCtx.WithError(err).Error("Failed to get event properties.")
		return nil, http.StatusInternalServerError
	}
	defer rows.Close()

	propertiesMap := make(map[string]map[interface{}]bool, 0)
	for rows.Next() {
		var propertiesJson []byte
		rows.Scan(&propertiesJson)

		err := U.FillPropertyKvsFromPropertiesJson(propertiesJson, &propertiesMap, U.SamplePropertyValuesLimit)
		if err != nil {
			log.WithError(err).WithField("properties_json",
				string(propertiesJson)).Error("Failed to unmarshal json properties.")
			return nil, http.StatusInternalServerError
		}
	}

	err = rows.Err()
	if err != nil {
		logCtx.WithError(err).Error("Failed to scan recent property keys.")
		return nil, http.StatusInternalServerError
	}

	propsByType, err := U.ClassifyPropertiesType(&propertiesMap)
	if err != nil {
		logCtx.WithError(err).Error("Failed to classify properties on get recent property keys.")
		return nil, http.StatusInternalServerError
	}

	return propsByType, http.StatusFound
}

func GetRecentEventPropertyKeys(projectId uint64, eventName string) (map[string][]string, int) {
	return GetRecentEventPropertyKeysWithLimits(projectId, eventName, eventsLimitForProperites)
}

// GetRecentEventPropertyValues - Returns unique event property
// values of given property from last 24 hours.
func GetRecentEventPropertyValuesWithLimits(projectId uint64, eventName string,
	property string, eventsLimit, valuesLimit int) ([]string, int) {

	db := C.GetServices().Db

	logCtx := log.WithFields(log.Fields{"projectId": projectId, "eventName": eventName, "property": property})

	eventsAfterTimestamp := U.UnixTimeBeforeDuration(24 * time.Hour)
	values := make([]string, 0, 0)
	queryStr := "SELECT DISTINCT(value) FROM" +
		" " + "(SELECT properties->? AS value FROM events WHERE project_id = ? AND event_name_id IN" +
		" " + "(SELECT id FROM event_names WHERE project_id = ? AND name = ?) AND timestamp > ? AND properties->? IS NOT NULL LIMIT ?)" +
		" " + "AS property_values LIMIT ?"

	rows, err := db.Raw(queryStr, property, projectId, projectId, eventName,
		eventsAfterTimestamp, property, eventsLimit, valuesLimit).Rows()
	if err != nil {
		logCtx.WithError(err).Error("Failed to get recent property keys.")
		return values, http.StatusInternalServerError
	}
	defer rows.Close()

	for rows.Next() {
		var value string
		rows.Scan(&value)
		value = U.TrimQuotes(value)
		values = append(values, value)
	}

	err = rows.Err()
	if err != nil {
		logCtx.WithError(err).Error("Failed scanning property value on type classifcation.")
		return values, http.StatusInternalServerError
	}
	return values, http.StatusFound
}

func GetRecentEventPropertyValues(projectId uint64, eventName string, property string) ([]string, int) {
	return GetRecentEventPropertyValuesWithLimits(projectId, eventName, property, eventsLimitForProperites, 2000)
}

func UpdateEventProperties(projectId uint64, id string, properties *U.PropertiesMap) int {
	if projectId == 0 || id == "" {
		return http.StatusBadRequest
	}

	event, errCode := GetEventById(projectId, id)
	if errCode != http.StatusFound {
		return errCode
	}

	updatedPostgresJsonb, err := U.AddToPostgresJsonb(&event.Properties, *properties)
	if err != nil {
		return http.StatusInternalServerError
	}

	db := C.GetServices().Db
	updatedFields := map[string]interface{}{"properties": updatedPostgresJsonb}
	err = db.Model(&Event{}).Where("project_id = ? AND id = ?", projectId, id).Update(updatedFields).Error
	if err != nil {
		log.WithFields(log.Fields{"project_id": projectId, "id": id,
			"update": updatedFields}).WithError(err).Error("Failed to update event properties.")
		return http.StatusInternalServerError
	}

	return http.StatusAccepted
}

func GetUserEventsByEventNameId(projectId uint64, userId string, eventNameId uint64) ([]Event, int) {
	if projectId == 0 {
		return nil, http.StatusBadRequest
	}

	db := C.GetServices().Db

	var events []Event
	if err := db.Order("timestamp DESC").Where("project_id = ? AND user_id = ? AND event_name_id = ?",
		projectId, userId, eventNameId).Find(&events).Error; err != nil {

		return events, http.StatusInternalServerError
	}

	if len(events) == 0 {
		return events, http.StatusNotFound
	}

	return events, http.StatusFound
}

func createSessionEvent(projectId uint64, userId string, sessionEventNameId uint64,
	isFirstSession bool, requestTimestamp int64, eventProperties,
	userProperties *U.PropertiesMap, userPropertiesId string) (*Event, int) {

	sessionEventProps := U.GetSessionProperties(isFirstSession, eventProperties, userProperties)
	sessionPropsJson, err := json.Marshal(sessionEventProps)
	if err != nil {
		log.WithError(err).Error("Failed to add session event properties. JSON marshal failed.")
		return nil, http.StatusInternalServerError
	}

	newSessionEvent, errCode := CreateEvent(&Event{
		EventNameId:      sessionEventNameId,
		Timestamp:        requestTimestamp,
		Properties:       postgres.Jsonb{sessionPropsJson},
		ProjectId:        projectId,
		UserId:           userId,
		UserPropertiesId: userPropertiesId,
	})
	if errCode != http.StatusCreated {
		return nil, errCode
	}

	propertiesToInsert := make(map[string]interface{})
	(propertiesToInsert)[U.UP_SESSION_COUNT] = newSessionEvent.Count

	errCode = GetAndOverWriteUserProperties(projectId, userId, userPropertiesId, propertiesToInsert)
	if errCode != http.StatusAccepted {
		log.WithField("UserId", userId).WithField("ErrCode", errCode).Error("Failed to overwrite user Properties with session count")
	} else {
		return newSessionEvent, http.StatusCreated
	}
	return newSessionEvent, errCode
}

func CreateOrGetSessionEvent(projectId uint64, userId string, isFirstSession bool, hasDefinedMarketingProperty bool,
	requestTimestamp int64, eventProperties, userProperties *U.PropertiesMap, userPropertiesId string) (*Event, int) {

	logCtx := log.WithField("project_id", projectId).WithField("user_id", userId)

	sessionEventName, errCode := CreateOrGetSessionEventName(projectId)
	if errCode != http.StatusCreated && errCode != http.StatusConflict {
		logCtx.Error("Failed to create session event name.")
		return nil, http.StatusInternalServerError
	}

	if hasDefinedMarketingProperty {
		// If the event has a marketing property, then the user is visiting again from a marketing channel.
		// Creating a new session event irrespective of timing to keep track of multiple marketing touch points
		// from the same user.
		return createSessionEvent(projectId, userId, sessionEventName.ID, isFirstSession, requestTimestamp,
			eventProperties, userProperties, userPropertiesId)
	}

	latestUserEvent, errCode := GetLatestAnyEventOfUserInDurationFromCache(projectId, userId,
		NewUserSessionInactivityDuration)
	if errCode == http.StatusNotFound || errCode == http.StatusInternalServerError {
		// Double check user's inactivity for the duration.
		dbLatestUserEvent, errCode := GetLatestAnyEventOfUserInDuration(projectId, userId,
			NewUserSessionInactivityDuration)
		if errCode == http.StatusInternalServerError {
			logCtx.Error("Failed to get latest any event of user in duration.")
			return nil, http.StatusInternalServerError
		}

		if errCode == http.StatusNotFound {
			return createSessionEvent(projectId, userId, sessionEventName.ID, isFirstSession, requestTimestamp,
				eventProperties, userProperties, userPropertiesId)
		}

		latestUserEvent = dbLatestUserEvent
	}

	// Get latest session event of user from events between user's last event timestamp and
	// one day before user's last event timestamp.
	latestSessionEvent, errCode := GetLatestEventOfUserByEventNameId(projectId, userId, sessionEventName.ID,
		latestUserEvent.Timestamp-86400, latestUserEvent.Timestamp)

	if errCode == http.StatusInternalServerError {
		logCtx.Error("Failed to get latest session event of user.")
		return nil, http.StatusInternalServerError
	}

	logCtx = logCtx.WithField("latest_event_timestamp", latestUserEvent.Timestamp)

	if errCode == http.StatusFound {
		if latestUserEvent.SessionId != nil && *latestUserEvent.SessionId != "" &&
			latestSessionEvent.ID != *latestUserEvent.SessionId {
			logCtx.WithField("latest_session_id", latestSessionEvent.ID).WithField("user_lastest_event_session_id",
				latestUserEvent.SessionId).Error("Latest session's id and session_id on last event of user not matching.")
		}

		return latestSessionEvent, http.StatusFound
	}

	if errCode == http.StatusNotFound {
		logCtx.Error("Session length of user exceeded 1 day. Created new session.")
		return createSessionEvent(projectId, userId, sessionEventName.ID, isFirstSession, requestTimestamp,
			eventProperties, userProperties, userPropertiesId)
	}

	return latestSessionEvent, http.StatusFound
}
