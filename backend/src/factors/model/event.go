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
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/lib/pq"

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
	Properties                 postgres.Jsonb `json:"properties,omitempty"`
	PropertiesUpdatedTimestamp int64          `gorm:"not null;default:0" json:"properties_updated_timestamp,omitempty"`
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
const OneDayInSeconds int64 = 24 * 60 * 60
const ThirtyMinutesInSeconds int64 = 30 * 60
const NewUserSessionInactivityInSeconds int64 = ThirtyMinutesInSeconds

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

	var additionalExpiryTime int64 = 5 * 60 // 5 mins
	cacheExpiry := NewUserSessionInactivityInSeconds + additionalExpiryTime
	err = cacheRedis.Set(key, string(cacheEventJson), float64(cacheExpiry))
	if err != nil {
		logCtx.WithError(err).Error("Failed to setCacheUserLastEvent.")
	}

	return err
}

func GetEventCountOfUserByEventName(projectId uint64, userId string, eventNameId uint64) (uint64, int) {
	var count uint64

	db := C.GetServices().Db
	if err := db.Model(&Event{}).Where("project_id = ? AND user_id = ? AND event_name_id = ?",
		projectId, userId, eventNameId).Count(&count).Error; err != nil {
		log.WithFields(log.Fields{"projectId": projectId, "userId": userId}).WithError(err).Error(
			"Failed to get count of event of user by event_name_id")
		return 0, http.StatusInternalServerError
	}

	return count, http.StatusFound
}

func CreateEvent(event *Event) (*Event, int) {
	logCtx := log.WithField("project_id", event.ProjectId)

	if event.ProjectId == 0 || event.UserId == "" || event.EventNameId == 0 {
		logCtx.Error("CreateEvent Failed. Invalid projectId or userId or eventNameId.")
		return nil, http.StatusBadRequest
	}

	if event.Timestamp == 0 {
		logCtx.WithField("timestamp", event.Timestamp).Error("CreateEvent Failed. Invalid timestamp.")
		return nil, http.StatusBadRequest
	}

	// Add id with our uuid generator, if not given.
	if event.ID == "" {
		event.ID = U.GetUUID()
	}

	// Increamenting count based on EventNameId, not by EventName.
	count, errCode := GetEventCountOfUserByEventName(event.ProjectId, event.UserId, event.EventNameId)
	if errCode == http.StatusInternalServerError {
		return nil, errCode
	}
	event.Count = count + 1

	eventPropsJSONb, err := U.FillHourAndDayEventProperty(&event.Properties, event.Timestamp)
	if err != nil {
		logCtx.WithField("eventTimestamp", event.Timestamp).WithError(err).Error(
			"Adding day of week and hour of day properties failed")
	}
	event.Properties = *eventPropsJSONb

	// Init properties updated timestamp with event timestamp.
	event.PropertiesUpdatedTimestamp = event.Timestamp

	db := C.GetServices().Db
	transTime := gorm.NowFunc()
	rows, err := db.Raw("INSERT INTO events (id, customer_event_id,project_id,user_id,user_properties_id,session_id,event_name_id,count,properties,properties_updated_timestamp,timestamp,created_at,updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) RETURNING events.id",
		event.ID, event.CustomerEventId, event.ProjectId, event.UserId, event.UserPropertiesId, event.SessionId, event.EventNameId, event.Count, event.Properties, event.PropertiesUpdatedTimestamp, event.Timestamp, transTime, transTime).Rows()
	if err != nil {
		if U.IsPostgresIntegrityViolationError(err) {
			logCtx.WithError(err).Info("CreateEvent Failed. Constraint violation.")
			return nil, http.StatusNotAcceptable
		}

		logCtx.WithFields(log.Fields{"event": &event}).WithError(err).Error("CreateEvent Failed")
		return nil, http.StatusInternalServerError
	}
	defer rows.Close()

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
		log.WithFields(log.Fields{"projectId": projectId, "userId": userId}).WithError(err).Error(
			"Getttng event failed on GetEvent")
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
			// Do not log error. Log on caller, if needed.
			return nil, http.StatusNotFound
		}

		log.WithFields(log.Fields{"project_id": projectId, "user_id": id}).WithError(err).Error(
			"Getttng event failed on GetEventbyId")
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

func GetLatestAnyEventOfUserForSessionFromCache(projectId uint64,
	userId string, newEventTimestamp int64) (*Event, int) {

	logCtx := log.WithFields(log.Fields{"project_id": projectId, "user_id": userId})

	cachedLastEvent, err := GetCacheUserLastEvent(projectId, userId)
	if err != nil {
		if err == redis.ErrNil {
			return nil, http.StatusNotFound
		}

		logCtx.WithError(err).Error(
			"Failed to get latest event of user in duration from cache.")
		return nil, http.StatusInternalServerError
	}

	isActive := cachedLastEvent.Timestamp > (newEventTimestamp - NewUserSessionInactivityInSeconds)
	if !isActive {
		return nil, http.StatusNotFound
	}

	event, errCode := GetEventById(projectId, cachedLastEvent.ID)
	if errCode != http.StatusFound {
		logCtx.WithField("err_code", errCode).Error(
			"Failed to get event on using id from cache.")
		return nil, errCode
	}

	return event, http.StatusFound
}

func GetLatestAnyEventOfUserForSession(projectId uint64, userId string,
	newEventTimestamp int64) (*Event, int) {

	if newEventTimestamp == 0 {
		return nil, http.StatusBadRequest
	}

	var events []Event
	// Is there any event in last 30 mins (inactivity) from current event timestamp?
	db := C.GetServices().Db
	err := db.Limit(1).Order("timestamp desc").Where("project_id = ? AND user_id = ? AND timestamp > ?",
		projectId, userId, newEventTimestamp-NewUserSessionInactivityInSeconds).Find(&events).Error
	if err != nil {
		return nil, http.StatusInternalServerError
	}

	if len(events) == 0 {
		return nil, http.StatusNotFound
	}

	return &events[0], http.StatusFound
}

func GetLatestEventOfUserByEventNameId(projectId uint64, userId string, eventNameId uint64,
	startTimestamp int64, endTimestamp int64) (*Event, int) {

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

func UpdateEventProperties(projectId uint64, id string,
	properties *U.PropertiesMap, updateTimestamp int64) int {

	if projectId == 0 || id == "" {
		return http.StatusBadRequest
	}

	event, errCode := GetEventById(projectId, id)
	if errCode != http.StatusFound {
		return errCode
	}

	overwriteExistingProperties := false
	propertiesLastUpdatedAt := event.PropertiesUpdatedTimestamp

	if updateTimestamp >= event.PropertiesUpdatedTimestamp {
		// Overwrite existing properties only, if the
		// state of update is future compared to existing.
		overwriteExistingProperties = true
		propertiesLastUpdatedAt = updateTimestamp
	}

	updatedPostgresJsonb, err := U.AddToPostgresJsonb(&event.Properties,
		*properties, overwriteExistingProperties)
	if err != nil {
		return http.StatusInternalServerError
	}

	db := C.GetServices().Db
	updatedFields := map[string]interface{}{
		"properties":                   updatedPostgresJsonb,
		"properties_updated_timestamp": propertiesLastUpdatedAt,
	}
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

func getPageCountAndTimeSpentFromEventsList(events []Event, sessionEvent *Event) (int64, float64) {
	if len(events) == 0 {
		return 0, 0
	}

	timeSpent := float64(events[len(events)-1].Timestamp - sessionEvent.Timestamp)
	pageCount := int64(len(events))

	return pageCount, timeSpent
}

func getPageCountAndTimeSpentFromSessionEvent(projectId uint64, userId string,
	sessionEvent *Event) (int64, float64, int) {

	logCtx := log.WithFields(log.Fields{"project_id": projectId, "user_id": userId})

	var count, timestamp int64
	db := C.GetServices().Db
	queryStr := "SELECT COUNT(*), MAX(timestamp) FROM events WHERE project_id = ? AND user_id = ? AND session_id = ?"
	execQuery := db.Raw(queryStr, projectId, userId, sessionEvent.ID)
	if err := execQuery.Error; err != nil {
		logCtx.WithError(err).Error("Failed to get session page count and spent time.")
		return 0, 0, http.StatusInternalServerError
	}

	row := execQuery.Row()
	if err := row.Scan(&count, &timestamp); err != nil {
		logCtx.WithError(err).Error("Failed to scan rows while reading session page count and spent time.")
		return 0, 0, http.StatusInternalServerError
	}

	var timeSpent float64
	if timestamp > 0 {
		timeSpent = float64(timestamp - sessionEvent.Timestamp)
	}

	return count, timeSpent, http.StatusFound
}

/* COMMENTED TEMPORARILY BECAUSE OF SLOW QUERY.
func enrichPreviousSessionEventProperties(projectId uint64, userId string,
	previousSessionEvent *Event) (float64, float64, int) {

	previousEventProperties, err := U.DecodePostgresJsonb(&previousSessionEvent.Properties)
	if err != nil {
		log.WithField("projectId", projectId).WithError(err).Error(
			"Failed to decode previous session event properties on createSessionEvent.")
		return 0, 0, http.StatusInternalServerError
	}

	latestEvent, errCode := GetLatestAnyEventOfUserForSession(projectId,
		userId, previousSessionEvent.Timestamp)
	if errCode != http.StatusFound {
		log.WithField("projectId", projectId).WithField("sessionId",
			previousSessionEvent.ID).Error("Failed to get latest event of session")
		return 0, 0, http.StatusInternalServerError
	}

	latestEventProperties, err := U.DecodePostgresJsonb(&latestEvent.Properties)
	if err != nil {
		log.WithField("projectId", projectId).WithError(err).Error(
			"Failed to decode latest event's properties for session.")
		return 0, 0, http.StatusInternalServerError
	}

	(*previousEventProperties)[U.SP_LATEST_PAGE_RAW_URL] = (*latestEventProperties)[U.EP_PAGE_RAW_URL]
	(*previousEventProperties)[U.SP_LATEST_PAGE_URL] = (*latestEventProperties)[U.EP_PAGE_URL]

	pageCount, timeSpent, errCode := getPageCountAndTimeSpentFromSessionEvent(
		projectId, userId, previousSessionEvent)
	if errCode == http.StatusInternalServerError {
		return 0, 0, errCode
	}
	(*previousEventProperties)[U.SP_PAGE_COUNT] = pageCount
	(*previousEventProperties)[U.SP_SPENT_TIME] = timeSpent

	previousEventPropertiesJSONb, err := U.EncodeToPostgresJsonb(previousEventProperties)
	if err != nil {
		log.WithField("projectId", projectId).WithError(err).Error(
			"Failed to encode previous session event on createSessionEvent.")
		return 0, 0, http.StatusInternalServerError
	}

	errCode = OverwriteEventProperties(projectId, userId, previousSessionEvent.ID, previousEventPropertiesJSONb)
	if errCode != http.StatusAccepted {
		log.WithField("projectId", projectId).Error(
			"Failed to overWrite previous session event properties on createSessionEvent.")
		return 0, 0, errCode
	}

	return float64(pageCount), timeSpent, http.StatusAccepted
}
*/

func createSessionEvent(projectId uint64, userId string, sessionEventNameId uint64,
	isFirstSession bool, requestTimestamp int64, eventProperties,
	userProperties *U.PropertiesMap, userPropertiesId string) (*Event, int) {

	var timeSpent float64
	var pageCount float64
	var err error

	/* COMMENTED TEMPORARILY
	previousSessionEvent, errCode := GetLatestEventOfUserByEventNameId(
		projectId, userId, sessionEventNameId, requestTimestamp-OneDayInSeconds, requestTimestamp)

	get page count and page spent time
	if errCode == http.StatusFound {
		pageCount, timeSpent, errCode = enrichPreviousSessionEventProperties(projectId,
			userId, previousSessionEvent)
		if errCode != http.StatusAccepted {
			log.WithField("project_id", projectId).Error(
				"Failed to enrich previous session event properties on createSesseionEvent")
		}
	}
	*/

	logCtx := log.WithField("project_id", projectId).WithField("user_id", userId)

	sessionEventProps := U.GetSessionProperties(isFirstSession, eventProperties, userProperties)
	sessionPropsJson, err := json.Marshal(sessionEventProps)
	if err != nil {
		logCtx.WithError(err).Error(
			"Failed to add session event properties. JSON marshal failed.")
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
		logCtx.Error("Failed to create session event.")
		return nil, errCode
	}

	newUserPropertiesFromSession := map[string]interface{}{
		// Properties from current session.
		U.UP_SESSION_COUNT: newSessionEvent.Count,
		// Properties from previous session.
		U.UP_PAGE_COUNT:       pageCount,
		U.UP_TOTAL_SPENT_TIME: timeSpent,
	}

	errCode = EnrichUserPropertiesWithSessionProperties(projectId, userId,
		userPropertiesId, newUserPropertiesFromSession, isFirstSession)
	if errCode != http.StatusAccepted {
		log.WithField("UserId", userId).WithField("err_code", errCode).Error(
			"Failed to overwrite user Properties with session count")
	} else {
		return newSessionEvent, http.StatusCreated
	}
	return newSessionEvent, errCode
}

func CreateOrGetSessionEvent(projectId uint64, userId string, isNewUser bool,
	hasDefinedMarketingProperty bool, newEventTimestamp int64, eventProperties,
	userProperties *U.PropertiesMap, userPropertiesId string) (*Event, int) {

	logCtx := log.WithField("project_id", projectId).WithField("user_id", userId)

	sessionEventName, errCode := CreateOrGetSessionEventName(projectId)
	if errCode != http.StatusCreated && errCode != http.StatusConflict {
		logCtx.Error("Failed to create session event name.")
		return nil, http.StatusInternalServerError
	}

	if isNewUser || hasDefinedMarketingProperty {
		// If user is new, it is unneccessary to check for users inactivity
		// before session creation as no events would be available.

		// If the event has a marketing property, then the user is visiting again
		// from a marketing channel. Creating a new session event irrespective of
		// timing to keep track of multiple marketing touch points
		// from the same user.

		return createSessionEvent(projectId, userId, sessionEventName.ID, isNewUser,
			newEventTimestamp, eventProperties, userProperties, userPropertiesId)
	}

	latestUserEvent, errCode := GetLatestAnyEventOfUserForSessionFromCache(projectId,
		userId, newEventTimestamp)
	if errCode == http.StatusNotFound || errCode == http.StatusInternalServerError {
		// Double check user's inactivity for the duration.
		dbLatestUserEvent, errCode := GetLatestAnyEventOfUserForSession(
			projectId, userId, newEventTimestamp)
		if errCode == http.StatusInternalServerError {
			logCtx.Error("Failed to get latest any event of user in duration.")
			return nil, http.StatusInternalServerError
		}

		if errCode == http.StatusNotFound {
			return createSessionEvent(projectId, userId, sessionEventName.ID,
				isNewUser, newEventTimestamp, eventProperties, userProperties,
				userPropertiesId)
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
			logCtx.WithField("latest_session_id", latestSessionEvent.ID).WithField(
				"user_lastest_event_session_id", latestUserEvent.SessionId).Error(
				"Latest session's id and session_id on last event of user not matching.")
		}

		return latestSessionEvent, http.StatusFound
	}

	if errCode == http.StatusNotFound {
		logCtx.Error("Session length of user exceeded 1 day. Created new session.")
		return createSessionEvent(projectId, userId, sessionEventName.ID, isNewUser,
			newEventTimestamp, eventProperties, userProperties, userPropertiesId)
	}

	return latestSessionEvent, http.StatusFound
}

func OverwriteEventProperties(projectId uint64, userId string, eventId string,
	newEventProperties *postgres.Jsonb) int {

	if newEventProperties == nil {
		return http.StatusBadRequest
	}

	db := C.GetServices().Db
	if err := db.Model(&Event{}).Where("project_id = ? AND user_id = ? AND id = ?",
		projectId, userId, eventId).Update(
		"properties", newEventProperties).Error; err != nil {
		log.WithFields(log.Fields{"projectId": projectId, "userId": userId}).WithError(err).Error(
			"Updating event properties failed in OverwriteEventProperties")
		return http.StatusInternalServerError
	}
	return http.StatusAccepted
}

func EnrichUserPropertiesWithSessionProperties(projectId uint64, userId string,
	userPropertiesId string, sessionProperties map[string]interface{}, isFirstSession bool) int {

	if len(sessionProperties) == 0 {
		return http.StatusBadRequest
	}

	userProperties, errCode := GetUserProperties(projectId, userId, userPropertiesId)
	if errCode != http.StatusFound {
		return errCode
	}

	userPropertiesMap, err := U.DecodePostgresJsonb(userProperties)
	if err != nil {
		return http.StatusInternalServerError
	}

	// Increment user properties by previous session properties.
	for key, value := range sessionProperties {
		// checking for the 2 properties that are to be updated only for new users
		if key == U.UP_PAGE_COUNT || key == U.UP_TOTAL_SPENT_TIME {
			// checking for firstSession => new user
			if isFirstSession {
				(*userPropertiesMap)[key] = 0
			} else {
				// checking if property is already initialized. * property will be initialized only in case of new user
				if (*userPropertiesMap)[key] != nil {
					(*userPropertiesMap)[key] = sessionProperties[key].(float64) +
						(*userPropertiesMap)[key].(float64)
				}
			}
		} else {
			(*userPropertiesMap)[key] = value
		}
	}

	userPropertiesJsonb, err := U.EncodeToPostgresJsonb(userPropertiesMap)
	if err != nil {
		return http.StatusInternalServerError
	}

	return OverwriteUserProperties(projectId, userId, userPropertiesId, userPropertiesJsonb)
}

func filterEventsBetweenTimestamp(events []Event, startTimestamp,
	endTimestamp int64) []Event {

	filteredEvents := make([]Event, 0, 0)
	for i := range events {
		if events[i].Timestamp >= startTimestamp && events[i].Timestamp <= endTimestamp {
			filteredEvents = append(filteredEvents, events[i])
		}
	}

	return filteredEvents
}

func associateSessionByEventIds(projectId uint64, eventIds []string, sessionId string) int {
	logCtx := log.WithFields(log.Fields{"project_id": projectId,
		"event_ids": eventIds, "session_id": sessionId})

	if projectId == 0 || len(eventIds) == 0 {
		logCtx.Error("Invalid args on associateSessionToEvents.")
		return http.StatusBadRequest
	}

	// Updates session_id to all events between given timestamp.
	updateFields := map[string]interface{}{"session_id": sessionId}
	db := C.GetServices().Db
	err := db.Model(&Event{}).Where("project_id = ? AND id = ANY(?)",
		projectId, pq.Array(eventIds)).Update(updateFields).Error
	if err != nil {
		logCtx.WithError(err).Error("Failed to associate session to events.")
		return http.StatusInternalServerError
	}

	return http.StatusAccepted
}

func associateSessionToEventsInBatch(projectId uint64, events []Event,
	sessionId string, batchSize int) int {

	eventIds := make([]string, 0, len(events))
	for i := range events {
		eventIds = append(eventIds, events[i].ID)
	}

	batchEventIds := U.GetStringListAsBatch(eventIds, batchSize)
	for i := range batchEventIds {
		errCode := associateSessionByEventIds(projectId, batchEventIds[i], sessionId)
		if errCode != http.StatusAccepted {
			return errCode
		}
	}

	return http.StatusAccepted
}

/*
AddSessionForUser - Will add session event based on conditions and associate session to each event.
The list of events being processed, would be like any of the given 2 cases.
* For users with session already (within max_lookback, if given). The first event would be the last event with session.
event_id - timestamp - session_id
e1 - t1 - s1
e2 - t2
e3 - t3
* For users without session already (within max_lookback, if given).
event_id - timestamp - session_id
e1 - t1
e2 - t2
e3 - t3
*/
func AddSessionForUser(projectId uint64, userId string, userEvents []Event,
	bufferTimeBeforeSessionCreateInSecs, startTimestamp int64,
	sessionEventNameId uint64) (int, bool, int) {

	logCtx := log.WithFields(log.Fields{"project_id": projectId, "user_id": userId})

	if len(userEvents) == 0 {
		return 0, false, http.StatusNotModified
	}

	latestUserEvent := &userEvents[len(userEvents)-1]
	endTimestamp := latestUserEvent.Timestamp - bufferTimeBeforeSessionCreateInSecs
	if endTimestamp <= startTimestamp {
		endTimestamp = latestUserEvent.Timestamp
	}

	events := filterEventsBetweenTimestamp(userEvents, startTimestamp, endTimestamp)
	if len(events) == 0 {
		return 0, false, http.StatusNotModified
	}

	// Use 2 moving cursor current, next. if diff(current, next) > in-activity
	// period, use current event as session_end_timestamp and update.
	// Update next event as session start and do the same till the end.
	sessionStartIndex := 0
	noOfSessionsCreated := 0
	sessionContinuedFlag := false
	for i := 0; i < len(events); {
		eventPropertiesDecoded, err := U.DecodePostgresJsonb(&events[i].Properties)
		if err != nil {
			logCtx.Error("Failed to decode event properties of first event on session.")
			return noOfSessionsCreated, sessionContinuedFlag, http.StatusInternalServerError
		}
		eventPropertiesMap := U.PropertiesMap(*eventPropertiesDecoded)

		var isNewSessionRequired bool
		if i < len(events)-1 {
			isNewSessionRequired = ((events[i+1].Timestamp - events[i].Timestamp) > NewUserSessionInactivityInSeconds)
		}
		_, hasMarketingProperty := U.MapEventPropertiesToDefinedProperties(&eventPropertiesMap)
		isLastSetOfEvents := i == len(events)-1

		if hasMarketingProperty || isNewSessionRequired || isLastSetOfEvents {
			var sessionEvent *Event
			var isSessionContinued bool

			// Continue with the last session_id, if available. This will be true as
			// first event will have max_timestamp (used as start_timestamp) where
			// session_id is not null.
			if events[sessionStartIndex].SessionId != nil {
				existingSessionEvent, errCode := GetEventById(projectId,
					*events[sessionStartIndex].SessionId)
				if errCode != http.StatusFound {
					logCtx.WithField("err_code", errCode).Error(
						"Failed to get existing session using session id on add session.")
					return noOfSessionsCreated, sessionContinuedFlag, http.StatusInternalServerError
				}
				sessionEvent = existingSessionEvent
				isSessionContinued = true
				sessionContinuedFlag = true
			} else {
				firstEvent := events[sessionStartIndex]

				userProperties, errCode := GetUserProperties(projectId, userId, firstEvent.UserPropertiesId)
				if errCode != http.StatusFound {
					logCtx.WithField("err_code", errCode).
						WithField("event_id", firstEvent.ID).
						WithField("user_properties_id", firstEvent.UserPropertiesId).
						Error("Failed to get user properties of first event on session.")
					return noOfSessionsCreated, sessionContinuedFlag, http.StatusInternalServerError
				}

				userPropertiesDecoded, err := U.DecodePostgresJsonb(userProperties)
				if err != nil {
					logCtx.WithField("user_properties_id", firstEvent.UserPropertiesId).
						Error("Failed to decode user properties of first event on session.")
					return noOfSessionsCreated, sessionContinuedFlag, http.StatusInternalServerError
				}
				userPropertiesMap := U.PropertiesMap(*userPropertiesDecoded)

				// Todo: Reuse event properties decoded for all events.
				firstEventPropertiesDecoded, err := U.DecodePostgresJsonb(&firstEvent.Properties)
				if err != nil {
					logCtx.Error("Failed to decode event properties of first event on session.")
					return noOfSessionsCreated, sessionContinuedFlag, http.StatusInternalServerError
				}
				firstEventPropertiesMap := U.PropertiesMap(*firstEventPropertiesDecoded)

				sessionEventCount, errCode := GetEventCountOfUserByEventName(projectId, userId, sessionEventNameId)
				if errCode == http.StatusInternalServerError {
					logCtx.Error("Failed to get session event count for user.")
					return noOfSessionsCreated, sessionContinuedFlag, errCode
				}
				isFirstSession := sessionEventCount == 0
				sessionPropertiesMap := U.GetSessionProperties(isFirstSession,
					&firstEventPropertiesMap, &userPropertiesMap)
				sessionPropertiesEncoded := map[string]interface{}(*sessionPropertiesMap)

				sessionPropertiesJsonb, err := U.EncodeToPostgresJsonb(&sessionPropertiesEncoded)
				if err != nil {
					logCtx.WithError(err).Error("Failed to encode session properties as postgres jsonb.")
					return noOfSessionsCreated, sessionContinuedFlag, http.StatusInternalServerError
				}

				// session event properties, to be updated
				newSessionEvent, errCode := CreateEvent(&Event{
					EventNameId: sessionEventNameId,
					// Timestamp - 1sec before the first event of session.
					Timestamp: firstEvent.Timestamp - 1,
					ProjectId: projectId,
					UserId:    userId,
					// UserPropertiesId - properties state at the time of first event of session.
					UserPropertiesId: firstEvent.UserPropertiesId,
					Properties:       *sessionPropertiesJsonb,
				})

				if errCode != http.StatusCreated {
					logCtx.Error("Failed to create session event.")
					return noOfSessionsCreated, sessionContinuedFlag, errCode
				}

				sessionEvent = newSessionEvent
				noOfSessionsCreated++
			}

			// Update the session_id to all events between start index and i.
			errCode := associateSessionToEventsInBatch(projectId,
				events[sessionStartIndex:i+1], sessionEvent.ID, 100)
			if errCode == http.StatusInternalServerError {
				logCtx.Error("Failed to associate session to events.")
				return noOfSessionsCreated, sessionContinuedFlag, errCode
			}

			lastEventProperties, err := U.DecodePostgresJsonb(&events[i].Properties)
			if err != nil {
				logCtx.Error("Failed to decode properties of last event of session.")
				return noOfSessionsCreated, sessionContinuedFlag, http.StatusInternalServerError
			}

			sessionPropertiesMap := U.PropertiesMap{}
			if _, exists := (*lastEventProperties)[U.EP_PAGE_RAW_URL]; exists {
				sessionPropertiesMap[U.SP_LATEST_PAGE_RAW_URL] = (*lastEventProperties)[U.EP_PAGE_RAW_URL]
			}
			if _, exists := (*lastEventProperties)[U.EP_PAGE_URL]; exists {
				sessionPropertiesMap[U.SP_LATEST_PAGE_URL] = (*lastEventProperties)[U.EP_PAGE_URL]
			}

			// Using existing method to get count and page spent time.
			var sessionPageCount int64
			var sessionPageSpentTime float64

			if isSessionContinued {
				// Using db query, since previous session continued, we don't have all the events of the session.
				sessionPageCount, sessionPageSpentTime, errCode = getPageCountAndTimeSpentFromSessionEvent(
					projectId, userId, sessionEvent)
				if errCode == http.StatusInternalServerError {
					logCtx.Error("Failed to get page count and spent time of session on add session.")
				}
			} else {
				// events from sessionStartIndex till i.
				sessionPageCount, sessionPageSpentTime =
					getPageCountAndTimeSpentFromEventsList(events[sessionStartIndex:i+1], sessionEvent)
			}

			if sessionPageCount > 0 {
				sessionPropertiesMap[U.SP_PAGE_COUNT] = sessionPageCount
			}
			if sessionPageSpentTime > 0 {
				sessionPropertiesMap[U.SP_SPENT_TIME] = sessionPageSpentTime
			}

			// Update session event properties.
			errCode = UpdateEventProperties(projectId, sessionEvent.ID,
				&sessionPropertiesMap, sessionEvent.Timestamp+1)
			if errCode == http.StatusInternalServerError {
				logCtx.Error("Failed updating session event properties on add session.")
				return noOfSessionsCreated, sessionContinuedFlag, errCode
			}

			sessionStartIndex = i + 1
		}

		i++
	}

	return noOfSessionsCreated, sessionContinuedFlag, http.StatusOK
}

func GetLatestUserEventByPageURLFromDB(projectID uint64, userID string, pageURL string) (*Event, int) {
	logCtx := log.WithField("project_id", projectID)

	db := C.GetServices().Db
	queryStr := "SELECT * FROM events WHERE project_id =? AND user_id = ? AND properties->>'$page_url' = ? ORDER BY timestamp DESC LIMIT 1"
	var event Event
	err := db.Raw(queryStr, projectID, userID, pageURL).Scan(&event).Error
	if err != nil {
		logCtx.WithError(err).Error("Failed to get event_id from project_id, user_id and page_url")
		return nil, http.StatusNotFound
	}
	return &event, http.StatusFound
}

// GetDatesForNextEventsArchivalBatch Get dates for events since startTime, excluding today's date.
func GetDatesForNextEventsArchivalBatch(projectID uint64, startTime int64) (map[string]int64, int) {
	db := C.GetServices().Db
	countByDates := make(map[string]int64)

	rows, err := db.Model(&Event{}).
		Where("project_id = ? AND timestamp BETWEEN ? AND (extract(epoch from current_date::timestamp at time zone 'UTC') - 1)", projectID, startTime).
		Group("date(to_timestamp(timestamp) at time zone 'UTC')").
		Select("date(to_timestamp(timestamp) at time zone 'UTC'), count(*)").Rows()
	if err != nil {
		log.WithError(err).Error("Failed to get dates for next event batches")
		return countByDates, http.StatusInternalServerError
	}
	defer rows.Close()

	for rows.Next() {
		var eventDate string
		var eventCount int64
		err = rows.Scan(&eventDate, &eventCount)
		if err != nil {
			log.WithError(err).Error("Failed to parse records")
			continue
		} else {
			countByDates[strings.Split(eventDate, "T")[0]] = eventCount
		}
	}

	return countByDates, http.StatusFound
}
