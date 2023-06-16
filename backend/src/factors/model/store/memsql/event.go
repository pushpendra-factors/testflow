package memsql

import (
	"database/sql"
	"encoding/json"
	"errors"
	cacheRedis "factors/cache/redis"
	C "factors/config"
	"factors/metrics"
	"factors/util"
	U "factors/util"
	"fmt"
	"net/http"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"

	"factors/model/model"
)

const eventsLimitForProperites = 50000
const OneDayInSeconds int64 = 24 * 60 * 60

func (store *MemSQL) GetHubspotFormEvents(projectID int64, userId string, timestamps []interface{}) ([]model.Event, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"user_id":    userId,
		"timestamps": timestamps,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	if projectID < 1 || userId == "" {
		log.Error("GetHubspotFormEvents Failed. Invalid projectId or userId")
		return nil, http.StatusBadRequest
	}

	if len(timestamps) == 0 {
		log.WithField("timestamps", timestamps).Error("GetHubspotFormEvents Failed. no available timestamp.")
		return nil, http.StatusBadRequest
	}

	db := C.GetServices().Db

	eventName, status := store.GetEventName(U.EVENT_NAME_HUBSPOT_CONTACT_FORM_SUBMISSION, projectID)
	if status != http.StatusOK && status != http.StatusFound && status != http.StatusNotModified {
		log.WithField("err_code", status).Error("Failed to get event name")
		return nil, http.StatusInternalServerError
	}

	var keyValues []interface{}
	keyValues = append(keyValues, projectID, userId, timestamps, eventName.ID)
	stmnt := "project_id = ? AND user_id = ? AND timestamp in ( ? ) AND event_name_id = ?"

	var events []model.Event
	err := db.Model(&model.Event{}).Where(stmnt, keyValues...).Find(&events).Error
	if err != nil {
		log.WithFields(logFields).WithError(err).Error("Failed to get the rows from events table")
		return nil, http.StatusInternalServerError
	}

	if len(events) == 0 {
		return nil, http.StatusNotFound
	}
	return events, http.StatusFound
}

func satisfiesEventConstraints(event model.Event) int {
	logFields := log.Fields{
		"event": event,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)

	// Unique (project_id, customer_event_id)
	if event.CustomerEventId != nil && *event.CustomerEventId != "" {
		if exists := existsEventByCustomerEventID(event.ProjectId, event.UserId, *event.CustomerEventId); exists {
			logCtx.WithField("customer_event_id", event.CustomerEventId).Warn("Event exists with customer event id")
			return http.StatusNotAcceptable
		}
	}

	// Deduplicate by event_id if not done by customer_event_id.
	if event.CustomerEventId == nil || (event.CustomerEventId != nil && *event.CustomerEventId == "") {
		if existsIDForProject(event.ProjectId, event.UserId, event.ID) {
			logCtx.Warn("Event exists with user_id and project_id")
			return http.StatusNotAcceptable
		}
	}

	if !U.IsValidUUID(event.ID) {
		logCtx.Error("Invalid value for event ID")
		// Internal server error error same as returned from Postgres on uuid voilation from DB side.
		return http.StatusInternalServerError
	}
	return http.StatusOK
}

func existsIDForProject(projectID int64, userID, eventID string) bool {
	logFields := log.Fields{
		"project_id": projectID,
		"user_id":    userID,
		"event_id":   eventID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	db := C.GetServices().Db

	var event model.Event
	err := db.Limit(1).Where("project_id = ? AND user_id = ? AND id = ?", projectID, userID, eventID).Select("id").Find(&event).Error
	if err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			log.WithFields(log.Fields{
				"project_id": projectID, "event_id": eventID, "user_id": userID}).Error("Failed to check event exists")
		}
		return false
	}

	if event.ID != "" {
		return true
	}
	return false
}

func (store *MemSQL) GetEventCountOfUserByEventName(projectId int64, userId string, eventNameId string) (uint64, int) {
	logFields := log.Fields{
		"project_id":    projectId,
		"user_id":       userId,
		"event_name_id": eventNameId,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	var count uint64

	db := C.GetServices().Db
	if err := db.Model(&model.Event{}).Where("project_id = ? AND user_id = ? AND event_name_id = ?",
		projectId, userId, eventNameId).Count(&count).Error; err != nil {
		log.WithFields(log.Fields{"projectId": projectId, "userId": userId}).WithError(err).Error(
			"Failed to get count of event of user by event_name_id")
		return 0, http.StatusInternalServerError
	}

	return count, http.StatusFound
}

// GetEventCountOfUsersByEventName Get count of events for event_name_id for multiple users.
func (store *MemSQL) GetEventCountOfUsersByEventName(projectID int64, userIDs []string, eventNameID string) (uint64, int) {
	logFields := log.Fields{
		"project_id":    projectID,
		"user_ids":      userIDs,
		"event_name_id": eventNameID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	var count uint64

	db := C.GetServices().Db
	if err := db.Model(&model.Event{}).Where("project_id = ? AND user_id IN (?) AND event_name_id = ?",
		projectID, userIDs, eventNameID).Count(&count).Error; err != nil {
		log.WithFields(log.Fields{"projectId": projectID, "userId": userIDs}).WithError(err).Error(
			"Failed to get count of event for users by event_name_id")
		return 0, http.StatusInternalServerError
	}

	return count, http.StatusFound
}

func (store *MemSQL) addEventDetailsToCache(projectID int64, event *model.Event, isUpdateEventProperty bool) {
	logFields := log.Fields{
		"project_id":               projectID,
		"event":                    event,
		"is_update_event_property": isUpdateEventProperty,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	// TODO: Remove this check after enabling caching realtime.
	blackListedForUpdate := make(map[string]bool)
	blackListedForUpdate[U.EP_PAGE_SPENT_TIME] = true
	blackListedForUpdate[U.EP_PAGE_SCROLL_PERCENT] = true

	eventsToIncrSortedSet := make([]cacheRedis.SortedSetKeyValueTuple, 0)
	propertiesToIncrSortedSet := make([]cacheRedis.SortedSetKeyValueTuple, 0)
	valuesToIncrSortedSet := make([]cacheRedis.SortedSetKeyValueTuple, 0)
	logCtx := log.WithFields(logFields)
	eventNameDetails, err := store.GetEventNameFromEventNameId(event.EventNameId, projectID)
	if err != nil {
		logCtx.WithError(err).Info("Failed to get event name from id")
		return
	}
	eventName := eventNameDetails.Name

	propertyMap, err := U.DecodePostgresJsonb(&event.Properties)
	if err != nil {
		logCtx.WithError(err).Info("Failed to decode json blob properties")
		return
	}
	eventProperties := *propertyMap

	currentTime := U.TimeNowZ()
	currentTimeDatePart := currentTime.Format(U.DATETIME_FORMAT_YYYYMMDD)

	var eventNamesKeySortedSet *cacheRedis.Key
	if model.IsEventNameTypeSmartEvent(eventNameDetails.Type) {
		eventNamesKeySortedSet, err = model.GetSmartEventNamesOrderByOccurrenceAndRecencyCacheKeySortedSet(projectID,
			currentTimeDatePart)
	} else {
		eventNamesKeySortedSet, err = model.GetEventNamesOrderByOccurrenceAndRecencyCacheKeySortedSet(projectID,
			currentTimeDatePart)
	}

	if err != nil {
		logCtx.WithError(err).Error("Failed to get cache key - events")
		return
	}
	eventsToIncrSortedSet = append(eventsToIncrSortedSet, cacheRedis.SortedSetKeyValueTuple{
		Key:   eventNamesKeySortedSet,
		Value: eventName,
	})

	for property, value := range eventProperties {
		if value == nil {
			continue
		}
		if !blackListedForUpdate[property] || !isUpdateEventProperty {
			category := store.GetPropertyTypeByKeyValue(projectID, eventName, property, value, false)
			var propertyValue string
			if category == U.PropertyTypeUnknown && reflect.TypeOf(value).Kind() == reflect.Bool {
				category = U.PropertyTypeCategorical
				propertyValue = fmt.Sprintf("%v", value)
			}
			if reflect.TypeOf(value).Kind() == reflect.String {
				propertyValue = value.(string)
			}
			propertyCategoryKeySortedSet, err := model.GetPropertiesByEventCategoryCacheKeySortedSet(projectID, currentTimeDatePart)
			if err != nil {
				logCtx.WithError(err).Error("Failed to get cache key - property category")
				return
			}
			propertiesToIncrSortedSet = append(propertiesToIncrSortedSet, cacheRedis.SortedSetKeyValueTuple{
				Key:   propertyCategoryKeySortedSet,
				Value: fmt.Sprintf("%s:SS-EN-PC:%s:%s", eventName, category, property),
			})
			if category == U.PropertyTypeCategorical {
				if propertyValue != "" {
					valueKeySortedSet, err := model.GetValuesByEventPropertyCacheKeySortedSet(projectID, currentTimeDatePart)
					if err != nil {
						logCtx.WithError(err).Error("Failed to get cache key - values")
						return
					}
					valuesToIncrSortedSet = append(valuesToIncrSortedSet, cacheRedis.SortedSetKeyValueTuple{
						Key:   valueKeySortedSet,
						Value: fmt.Sprintf("%s:SS-EN-PC:%s:SS-EN-PV:%s", eventName, property, propertyValue),
					})
				}
			}
		}
	}
	begin := U.TimeNowZ()
	keysToIncrSortedSet := make([]cacheRedis.SortedSetKeyValueTuple, 0)
	if !isUpdateEventProperty {
		keysToIncrSortedSet = append(keysToIncrSortedSet, eventsToIncrSortedSet...)
	}
	keysToIncrSortedSet = append(keysToIncrSortedSet, propertiesToIncrSortedSet...)
	keysToIncrSortedSet = append(keysToIncrSortedSet, valuesToIncrSortedSet...)
	if len(keysToIncrSortedSet) <= 0 {
		return
	}
	counts, _ := cacheRedis.ZincrPersistentBatch(false, keysToIncrSortedSet...)
	end := U.TimeNowZ()
	metrics.Increment(metrics.IncrEventCacheCounter)
	metrics.RecordLatency(metrics.LatencyEventCache, float64(end.Sub(begin).Milliseconds()))
	if err != nil {
		logCtx.WithError(err).Error("Failed to increment keys")
		return
	}

	newEventCount := int64(0)
	index := 0
	if len(counts) > 0 {
		if counts[index] == 1 && !isUpdateEventProperty {
			newEventCount++
			index++
		}
	}
	analyticsKeysInCache := make([]cacheRedis.SortedSetKeyValueTuple, 0)
	if newEventCount > 0 {
		uniqueEventsCountKey, err := model.UniqueEventNamesAnalyticsCacheKey(currentTimeDatePart)
		if err != nil {
			logCtx.WithError(err).Error("Failed to get cache key - uniqueEventsCountKey")
			return
		}
		analyticsKeysInCache = append(analyticsKeysInCache, cacheRedis.SortedSetKeyValueTuple{
			Key:   uniqueEventsCountKey,
			Value: fmt.Sprintf("%v", projectID)})
	}
	if !isUpdateEventProperty {
		totalEventsCountKey, err := model.EventsCountAnalyticsCacheKey(currentTimeDatePart)
		if err != nil {
			logCtx.WithError(err).Error("Failed to get cache key - totalEventsCountKey")
			return
		}
		analyticsKeysInCache = append(analyticsKeysInCache, cacheRedis.SortedSetKeyValueTuple{
			Key:   totalEventsCountKey,
			Value: fmt.Sprintf("%v", projectID)})
	}
	if len(analyticsKeysInCache) > 0 {
		_, err = cacheRedis.ZincrPersistentBatch(true, analyticsKeysInCache...)
		if err != nil {
			logCtx.WithError(err).Error("Failed to increment keys")
			return
		}
	}
}

func (store *MemSQL) CreateEvent(event *model.Event) (*model.Event, int) {
	logFields := log.Fields{
		"event": event,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)
	if event.ProjectId == 0 || event.UserId == "" || event.EventNameId == "" {
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

	var timezoneString U.TimeZoneString
	if C.IsIngestionTimezoneEnabled(event.ProjectId) {
		var statusCode int
		timezoneString, statusCode = store.GetTimezoneByIDWithCache(event.ProjectId)
		if statusCode != http.StatusFound {
			log.Errorf("Failed to get project Timezone for project: %d and event: %v ", event.ProjectId, event)
			return nil, http.StatusInternalServerError
		}
	} else {
		timezoneString = U.TimeZoneStringUTC
	}

	if event.IsFromPast {
		eventProperties, err := U.AddToPostgresJsonb(&event.Properties, map[string]interface{}{"$is_from_past": event.IsFromPast}, true)
		if err != nil {
			logCtx.WithError(err).Error("Failed to add IsPast in event properties during event creation.")
		} else {
			event.Properties = *eventProperties
		}
	}

	// Use current properties of user, if user_properties is not provided and if it is not a past event.
	if event.UserProperties == nil && !event.IsFromPast {
		properties, errCode := store.GetUserPropertiesByUserID(event.ProjectId, event.UserId)

		newUserProperties := RemoveDisabledEventUserProperties(event.ProjectId, properties)

		if errCode != http.StatusFound {
			logCtx.WithField("err_code", errCode).Error("Failed to get properties of user for event creation.")
		}
		event.UserProperties = newUserProperties

	}

	// Incrementing count based on EventNameId, not by EventName.
	count, errCode := store.GetEventCountOfUserByEventName(event.ProjectId, event.UserId, event.EventNameId)
	if errCode == http.StatusInternalServerError {
		return nil, errCode
	}
	event.Count = count + 1

	// Sanitize Unicode Properties.
	U.SantizePostgresJsonbForUnicode(&event.Properties)
	U.SantizePostgresJsonbForUnicode(event.UserProperties)

	eventPropsJSONb, err := U.FillHourDayAndTimestampEventProperty(&event.Properties, event.Timestamp, timezoneString)
	if err != nil {
		logCtx.WithField("eventTimestamp", event.Timestamp).WithError(err).Error(
			"Adding day of week, hour of day and timestamp properties failed")
	}
	eventPropsJSONb = U.SanitizePropertiesJsonb(eventPropsJSONb)
	event.Properties = *eventPropsJSONb
	// Init properties updated timestamp with event timestamp.
	event.PropertiesUpdatedTimestamp = event.Timestamp

	// Adding the data to cache. Even if it fails, continue silent
	store.addEventDetailsToCache(event.ProjectId, event, false)

	transTime := gorm.NowFunc()
	columnsInOrder := "id, customer_event_id, project_id, user_id, session_id, event_name_id," + " " +
		"count, properties, properties_updated_timestamp, timestamp, created_at, updated_at"
	paramsInOrder := []interface{}{event.ID, event.CustomerEventId, event.ProjectId, event.UserId,
		event.SessionId, event.EventNameId, event.Count, event.Properties, event.PropertiesUpdatedTimestamp,
		event.Timestamp, transTime, transTime}

	// Conditinal columns added to the end of column list and params.
	if event.UserProperties != nil {
		columnsInOrder = columnsInOrder + "," + "user_properties"
		paramsInOrder = append(paramsInOrder, event.UserProperties)
	}

	var columnsPlaceholder string
	for i := range paramsInOrder {
		columnsPlaceholder = columnsPlaceholder + "?"
		if i < len(paramsInOrder)-1 {
			columnsPlaceholder = columnsPlaceholder + ", "
		}
	}

	errCode = satisfiesEventConstraints(*event)
	if errCode != http.StatusOK {
		return nil, errCode
	}

	db := C.GetServices().Db
	statement := fmt.Sprintf("INSERT INTO events (%s) VALUES (%s)", columnsInOrder, columnsPlaceholder)
	rows, err := db.Raw(statement, paramsInOrder...).Rows()
	if err != nil {
		logCtx.WithField("event", event).WithError(err).Error("CreateEvent Failed")
		return nil, http.StatusInternalServerError
	}
	defer rows.Close()

	// log for analysis.
	log.WithField("project_id", event.ProjectId).
		WithField("event_name_id", event.EventNameId).
		WithField("tag", "create_event").
		Info("Created Event.")

	event.CreatedAt = transTime
	event.UpdatedAt = transTime

	model.SetCacheUserLastEvent(event.ProjectId, event.UserId,
		&model.CacheEvent{ID: event.ID, Timestamp: event.Timestamp})

	t1 := time.Now()
	alerts, eventName, ErrCode := store.MatchEventTriggerAlertWithTrackPayload(event.ProjectId, event.EventNameId, &event.Properties, event.UserProperties, nil, false)
	if ErrCode == http.StatusFound && alerts != nil {
		// log.WithFields(log.Fields{"project_id": event.ProjectId,
		// 	"event_trigger_alerts": *alerts}).Info("EventTriggerAlert found. Caching Alert.")

		for _, alert := range *alerts {
			success := store.CacheEventTriggerAlert(&alert, event, eventName)
			if !success {
				log.WithFields(log.Fields{"project_id": event.ProjectId,
					"event_trigger_alert": alert}).Error("Caching alert failure for ", alert)
			}
		}
	}
	log.Info("Control past EventTrigger block: ", time.Since(t1))

	return event, http.StatusCreated
}

// existsEventByCustomerEventID Get events by projectID and customerEventID.
func existsEventByCustomerEventID(projectID int64, userID, customerEventID string) bool {
	logFields := log.Fields{
		"project_id":         projectID,
		"user_id":            userID,
		"ecustomer_event_id": customerEventID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	var event model.Event

	db := C.GetServices().Db
	if err := db.Limit(1).Where("project_id = ?", projectID).
		Where("user_id = ?", userID).
		Where("customer_event_id = ?", customerEventID).
		Select("id").Find(&event).Error; err != nil {
		return false
	}

	if event.ID != "" {
		return true
	}

	return false
}

func (store *MemSQL) GetEvent(projectId int64, userId string, id string) (*model.Event, int) {
	logFields := log.Fields{
		"project_id": projectId,
		"user_id":    userId,
		"id":         id,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	if !U.IsValidUUID(id) {
		return nil, http.StatusInternalServerError
	}
	var event model.Event

	db := C.GetServices().Db
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

func (store *MemSQL) GetEventById(projectId int64, id, userID string) (*model.Event, int) {
	logFields := log.Fields{
		"project_id": projectId,
		"user_id":    userID,
		"id":         id,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	var event model.Event

	db := C.GetServices().Db
	dbx := db.Limit(1).Where("project_id = ?", projectId).Where("id = ?", id)
	// TODO: Make userID mandatory once support is added to all queries.
	if userID != "" {
		dbx = dbx.Where("user_id = ?", userID)
	}

	if err := dbx.Find(&event).Error; err != nil {
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

func (store *MemSQL) GetLatestEventOfUserByEventNameId(projectId int64, userId string, eventNameId string,
	startTimestamp int64, endTimestamp int64) (*model.Event, int) {
	logFields := log.Fields{
		"project_id":      projectId,
		"user_id":         userId,
		"event_name_id":   eventNameId,
		"start_timestamp": startTimestamp,
		"end_timestamp":   endTimestamp,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	if startTimestamp == 0 || endTimestamp == 0 {
		return nil, http.StatusBadRequest
	}

	var events []model.Event

	db := C.GetServices().Db
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

// GetRecentEventPropertyKeysWithLimits This method gets all the recent 'limit' property keys
// from DB for a given project/event
func (store *MemSQL) GetRecentEventPropertyKeysWithLimits(projectID int64, eventName string,
	starttime int64, endtime int64, eventsLimit int) ([]U.Property, error) {
	logFields := log.Fields{
		"project_id":   projectID,
		"events_limit": eventsLimit,
		"event_name":   eventName,
		"starttime":    starttime,
		"endtime":      endtime,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)
	properties := make([]U.Property, 0)

	queryStr := "SELECT properties, " +
		" " + "timestamp as last_seen" +
		" " + "FROM events  " +
		" " + "WHERE project_id = ? AND event_name_id IN ( " +
		" " + "	SELECT id FROM event_names WHERE project_id = ? AND name = ? " +
		" " + ") " +
		" " + "AND timestamp > ? AND timestamp <= ? AND properties != 'null' AND properties IS NOT NULL"

	db := C.GetServices().Db
	rows, err := db.Raw(queryStr, projectID, projectID, eventName, starttime, endtime).Rows()
	if err != nil {
		logCtx.WithError(err).Error("Failed to get event properties.")
		return nil, err
	}
	defer rows.Close()

	propertiesCounts := make(map[string]map[string]int64)
	for rows.Next() {
		var lastSeen int64
		var eventProperties postgres.Jsonb
		if err := rows.Scan(&eventProperties, &lastSeen); err != nil {
			logCtx.WithError(err).Error("Failed scanning rows on GetRecentEventPropertyKeysWithLimits")
			return properties, err
		}
		propertiesMap, err := U.DecodePostgresJsonbAsPropertiesMap(&eventProperties)
		if err != nil {
			logCtx.WithError(err).Error("Failed to decode properties on GetRecentEventPropertyKeysWithLimits")
			return properties, err
		}

		for key := range *propertiesMap {
			if _, found := propertiesCounts[key]; found {
				propertiesCounts[key]["count"]++
				propertiesCounts[key]["last_seen"] = U.Max(propertiesCounts[key]["last_seen"], lastSeen)
			} else {
				propertiesCounts[key] = map[string]int64{
					"count":     1,
					"last_seen": lastSeen,
				}
			}
		}
	}

	for propertyKey := range propertiesCounts {
		properties = append(properties, U.Property{
			Key:      propertyKey,
			LastSeen: uint64(propertiesCounts[propertyKey]["last_seen"]),
			Count:    propertiesCounts[propertyKey]["count"]})
	}

	sort.Slice(properties, func(i, j int) bool {
		return properties[i].Count > properties[j].Count
	})
	return properties[:U.MinInt(eventsLimit, len(properties))], nil
}

// GetRecentEventPropertyValuesWithLimits This method gets all the recent 'limit' property values from DB for a given project/event/property
func (store *MemSQL) GetRecentEventPropertyValuesWithLimits(projectID int64, eventName string,
	property string, valuesLimit int, rowsLimit int, starttime int64,
	endtime int64) ([]U.PropertyValue, string, error) {
	logFields := log.Fields{
		"project_id":   projectID,
		"property":     property,
		"event_name":   eventName,
		"starttime":    starttime,
		"endtime":      endtime,
		"rows_limit":   rowsLimit,
		"values_limit": valuesLimit,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	db := C.GetServices().Db
	logCtx := log.WithFields(logFields)

	values := make([]U.PropertyValue, 0)
	// TODO: Use additional table for property value fetching, if this is slow.
	queryStr := fmt.Sprintf("SELECT value, COUNT(*) AS count, MAX(timestamp) AS last_seen, JSON_GET_TYPE(value) AS value_type FROM"+
		" "+"(SELECT JSON_EXTRACT_STRING(properties, ?) AS value, timestamp FROM events WHERE project_id = ? AND event_name_id IN"+
		" "+"(SELECT id FROM event_names WHERE project_id = ? AND name = ?) AND timestamp > ? AND timestamp <= ? AND JSON_EXTRACT_STRING(properties, ?) IS NOT NULL LIMIT %d)"+
		" "+"AS property_values GROUP BY value ORDER BY count DESC LIMIT %d", rowsLimit, valuesLimit)

	rows, err := db.Raw(queryStr, property, projectID, projectID, eventName,
		starttime, endtime, property).Rows()
	if err != nil {
		logCtx.WithError(err).Error("Failed to get recent property values.")
		return nil, "", err
	}
	defer rows.Close()

	for rows.Next() {
		var value U.PropertyValue
		if err := db.ScanRows(rows, &value); err != nil {
			logCtx.WithError(err).Error("Failed scanning rows on GetRecentEventPropertyValuesWithLimits")
			return nil, "", err
		}
		value.Value = U.TrimQuotes(value.Value)
		values = append(values, value)
	}

	err = rows.Err()
	if err != nil {
		logCtx.WithError(err).Error("Failed scanning property value on type classifcation.")
		return nil, "", err
	}
	return values, U.GetCategoryType(property, values), nil
}

func (store *MemSQL) UpdateEventPropertiesInBatch(projectID int64,
	batchedUpdateEventPropertiesParams []model.UpdateEventPropertiesParams) bool {

	logFields := log.Fields{"project_id": projectID}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)

	db := C.GetServices().Db
	dbTx := db.Begin()
	if dbTx.Error != nil {
		logCtx.WithError(dbTx.Error).Error("Failed to begin transaction in batch user properties update.")
		return true
	}
	logCtx.Info("Using batch transaction in UpdateEventPropertiesInBatch.")

	hasFailure := false
	for i := range batchedUpdateEventPropertiesParams {
		projectID := batchedUpdateEventPropertiesParams[i].ProjectID
		userID := batchedUpdateEventPropertiesParams[i].UserID
		eventID := batchedUpdateEventPropertiesParams[i].EventID
		updateTimestamp := batchedUpdateEventPropertiesParams[i].SessionEventTimestamp
		properties := batchedUpdateEventPropertiesParams[i].SessionProperties

		optionalEventUserProperties := batchedUpdateEventPropertiesParams[i].NewSessionEventUserProperties
		newUserProperties := RemoveDisabledEventUserProperties(projectID, optionalEventUserProperties)

		status := store.updateEventPropertiesWithTransaction(projectID, eventID, userID, properties,
			updateTimestamp, newUserProperties, dbTx)
		if status != http.StatusAccepted {
			logCtx.WithFields(log.Fields{"update_event_properties_params": batchedUpdateEventPropertiesParams[i]}).
				Error("Failed to overwrite event user properties in batch.")
			hasFailure = true
		}
	}

	err := dbTx.Commit().Error
	if err != nil {
		logCtx.WithError(err).Error("Failure in batch event user properties update.")
		hasFailure = true
	}

	return hasFailure
}

func (store *MemSQL) UpdateEventProperties(projectId int64, id, userID string,
	properties *U.PropertiesMap, updateTimestamp int64,
	optionalEventUserProperties *postgres.Jsonb) int {
	db := C.GetServices().Db
	return store.updateEventPropertiesWithTransaction(projectId, id, userID, properties, updateTimestamp, optionalEventUserProperties, db)
}

func (store *MemSQL) updateEventPropertiesWithTransaction(projectId int64, id, userID string,
	properties *U.PropertiesMap, updateTimestamp int64,
	optionalEventUserProperties *postgres.Jsonb, dbTx *gorm.DB) int {

	logFields := log.Fields{
		"project_id":                     projectId,
		"id":                             id,
		"user_id":                        userID,
		"properties":                     properties,
		"updateTimestamp":                updateTimestamp,
		"optional_event_user_properties": optionalEventUserProperties,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	if projectId == 0 || id == "" {
		return http.StatusBadRequest
	}

	if dbTx == nil {
		log.WithFields(log.Fields{"project_id": projectId, "id": id}).Error("Missing DB method in updateEventProperties.")
		return http.StatusBadRequest
	}

	if value := U.SafeConvertToFloat64((*properties)[U.EP_PAGE_SCROLL_PERCENT]); value > float64(100.00) {
		log.WithFields(logFields).Info("Scroll Percent greater than 100. Changing value to 100")
		(*properties)[U.EP_PAGE_SCROLL_PERCENT] = 100.00
	}

	event, errCode := store.GetEventById(projectId, id, userID)
	if errCode != http.StatusFound {
		return errCode
	}
	U.SanitizeProperties(properties)

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

	updatedFields := map[string]interface{}{
		"properties":                   updatedPostgresJsonb,
		"properties_updated_timestamp": propertiesLastUpdatedAt,
	}

	// Optional event user_properties update with
	// event properties update.
	if optionalEventUserProperties != nil {

		newUserProperties := RemoveDisabledEventUserProperties(projectId, optionalEventUserProperties)
		updatedFields["user_properties"] = newUserProperties
	}

	EnableOLTPQueriesMemSQLImprovements := C.EnableOLTPQueriesMemSQLImprovements(projectId)

	var dbx *gorm.DB
	if EnableOLTPQueriesMemSQLImprovements {
		dbx = dbTx.Model(&model.Event{}).Where("project_id = ? AND id = ? AND timestamp = ? AND event_name_id = ?",
			projectId, id, event.Timestamp, event.EventNameId)
	} else {
		dbx = dbTx.Model(&model.Event{}).Where("project_id = ? AND id = ?", projectId, id)
	}
	if userID != "" {
		dbx = dbx.Where("user_id = ?", userID)
	}

	err = dbx.Update(updatedFields).Error
	if err != nil {
		log.WithFields(log.Fields{"project_id": projectId, "id": id,
			"update": updatedFields}).WithError(err).Error("Failed to update event properties.")
		return http.StatusInternalServerError
	}
	updatedProperties := make(map[string]interface{}, 0)
	updatedProperties = *properties
	updatedPropertiesOnlyJsonBlob, err := U.EncodeToPostgresJsonb(&updatedProperties)
	if err == nil {
		store.addEventDetailsToCache(projectId, &model.Event{EventNameId: event.EventNameId, Properties: *updatedPropertiesOnlyJsonBlob}, true)
	}

	//log.Info("EventTriggerAlerts match function trigger point.")
	alerts, eventName, ErrCode := store.MatchEventTriggerAlertWithTrackPayload(event.ProjectId, event.EventNameId, updatedPostgresJsonb, event.UserProperties, updatedPropertiesOnlyJsonBlob, true)
	if ErrCode == http.StatusFound && alerts != nil {
		// log.WithFields(log.Fields{"project_id": event.ProjectId,
		// 	"event_trigger_alerts": *alerts}).Info("EventTriggerAlert found. Caching Alert.")

		updatedEvent := model.Event{}
		updatedEvent = *event
		updatedEvent.Properties = *updatedPostgresJsonb
		for _, alert := range *alerts {
			success := store.CacheEventTriggerAlert(&alert, &updatedEvent, eventName)
			if !success {
				log.WithFields(log.Fields{"project_id": event.ProjectId,
					"event_trigger_alert": alert}).Error("Caching alert failure for ", alert)
			}
		}
	}

	// Log for analysis.
	log.WithField("project_id", projectId).WithField("tag", "update_event").Info("Updated event.")

	return http.StatusAccepted
}

func (store *MemSQL) GetUserEventsByEventNameId(projectId int64, userId string, eventNameId string) ([]model.Event, int) {
	logFields := log.Fields{
		"project_id":    projectId,
		"user_id":       userId,
		"event_name_id": eventNameId,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	if projectId == 0 {
		return nil, http.StatusBadRequest
	}

	var events []model.Event

	db := C.GetServices().Db
	if err := db.Where("project_id = ? AND user_id = ? AND event_name_id = ?",
		projectId, userId, eventNameId).Find(&events).Error; err != nil {

		return events, http.StatusInternalServerError
	}

	if len(events) == 0 {
		return events, http.StatusNotFound
	}

	// sort by timestamp DESC
	sort.Slice(events, func(i, j int) bool {
		return events[i].Timestamp > events[j].Timestamp
	})

	return events, http.StatusFound
}

func getPageCountAndTimeSpentFromEventsList(events []*model.Event, sessionEvent *model.Event) (float64, float64) {
	logFields := log.Fields{
		"events":        events,
		"session_event": sessionEvent,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	if len(events) == 0 {
		return 0, 0
	}
	timeSpent := float64(0)
	pageCount := float64(0)
	for _, event := range events {
		if event.ID != sessionEvent.ID && event.SessionId == nil {
			properties, _ := U.DecodePostgresJsonb(&event.Properties)
			pageSpentTime, _ := U.GetPropertyValueAsFloat64((*properties)[U.EP_PAGE_SPENT_TIME])
			timeSpent += pageSpentTime
			pageCount += 1
		}
	}

	return pageCount, timeSpent
}

func getPageCountAndTimeSpentForContinuedSession(projectId int64, userId string,
	continuedSessionEvent *model.Event, events []*model.Event) (float64, float64, float64, float64, int) {
	logFields := log.Fields{
		"project_id":              projectId,
		"user_id":                 userId,
		"continued_session_event": continuedSessionEvent,
		"events":                  events,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)
	existingPropertiesMap, err := U.DecodePostgresJsonb(&continuedSessionEvent.Properties)
	if err != nil {
		return 0, 0, 0, 0, http.StatusInternalServerError
	}

	var existingPageCount float64
	if existingPageCountValue, exists := (*existingPropertiesMap)[U.SP_PAGE_COUNT]; exists {
		existingPageCount, err = U.GetPropertyValueAsFloat64(existingPageCountValue)
		if err != nil {
			logCtx.WithError(err).Error("Failed to get page count property value as float64.")
		}
	}

	var existingSpentTime float64
	if existingSpentTimeValue, exists := (*existingPropertiesMap)[U.SP_SPENT_TIME]; exists {
		existingSpentTime, err = U.GetPropertyValueAsFloat64(existingSpentTimeValue)
		if err != nil {
			logCtx.WithError(err).Error("Failed to get spent count property value as float64.")
		}
	}

	currentPageCount, currentSpentTime := getPageCountAndTimeSpentFromEventsList(events, continuedSessionEvent)
	// Decrement by 1 to remove the last event of session pulled for
	// existing session which is duplicate on count.
	pageCount := existingPageCount + currentPageCount
	spentTime := existingSpentTime + currentSpentTime
	return pageCount, spentTime, currentPageCount, currentSpentTime, http.StatusFound
}

func (store *MemSQL) OverwriteEventProperties(projectId int64, userId string, eventId string,
	newEventProperties *postgres.Jsonb) int {
	logFields := log.Fields{
		"project_id":           projectId,
		"user_id":              userId,
		"event_id":             eventId,
		"new_event_properties": newEventProperties,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	if newEventProperties == nil {
		return http.StatusBadRequest
	}
	newEventProperties = U.SanitizePropertiesJsonb(newEventProperties)

	db := C.GetServices().Db
	if err := db.Model(&model.Event{}).Where("project_id = ? AND user_id = ? AND id = ?",
		projectId, userId, eventId).Update(
		"properties", newEventProperties).Error; err != nil {
		log.WithFields(log.Fields{"projectId": projectId, "userId": userId}).WithError(err).Error(
			"Updating event properties failed in OverwriteEventProperties")
		return http.StatusInternalServerError
	}

	// Log for analysis.
	log.WithField("project_id", projectId).WithField("tag", "update_event").Info("Updated event.")

	return http.StatusAccepted
}

func (store *MemSQL) OverwriteEventPropertiesByID(projectId int64, id string,
	newEventProperties *postgres.Jsonb) int {
	logFields := log.Fields{
		"project_id":           projectId,
		"id":                   id,
		"new_event_properties": newEventProperties,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)
	if newEventProperties == nil {
		return http.StatusBadRequest
	}
	newEventProperties = U.SanitizePropertiesJsonb(newEventProperties)

	db := C.GetServices().Db
	err := db.Model(&model.Event{}).
		Where("project_id = ? AND id = ?", projectId, id).
		Update("properties", newEventProperties).Error
	if err != nil {
		logCtx.WithError(err).Error("Updating event properties failed in OverwriteEventPropertiesByID.")
		return http.StatusInternalServerError
	}

	// Log for analysis.
	log.WithField("project_id", projectId).WithField("tag", "update_event").Info("Updated event.")

	return http.StatusAccepted
}

func doesEventIsPageViewAndHasMarketingProperty(event *model.Event) (bool, error) {
	logFields := log.Fields{
		"event": event,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	if event == nil {
		return false, errors.New("nil event")
	}

	eventPropertiesDecoded, err := U.DecodePostgresJsonb(&((*event).Properties))
	if err != nil {
		return false, err
	}
	eventPropertiesMap := U.PropertiesMap(*eventPropertiesDecoded)

	isPageAndHasMarketingProperty := U.IsPageViewEvent(&eventPropertiesMap) &&
		U.HasDefinedMarketingProperty(&eventPropertiesMap)
	return isPageAndHasMarketingProperty, nil
}

func filterEventsForSession(events []model.Event, endTimestamp int64) []*model.Event {
	logFields := log.Fields{
		"events":        events,
		"end_timestamp": endTimestamp,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	filteredEvents := make([]*model.Event, 0, 0)
	// Filter events by user specific end_timestamp.
	for i := range events {
		if events[i].Timestamp <= endTimestamp {
			// Using address as append doesn't use ref by default.
			filteredEvents = append(filteredEvents, &events[i])
		}
	}

	// Remove the session continuation event (first event with session_id)
	// when the first event to add session have marketing property,
	// to avoid continuing session.
	if len(filteredEvents) > 1 && filteredEvents[0].SessionId != nil {
		hasMarketingProperty, err := doesEventIsPageViewAndHasMarketingProperty(filteredEvents[1])
		if err != nil {
			log.WithError(err).Error("Failed to decode properties Jsonb.")
			return filteredEvents
		}

		if hasMarketingProperty {
			return filteredEvents[1:len(filteredEvents)]
		}
	}

	return filteredEvents
}

func (store *MemSQL) AssociateSessionByEventIdsBatchV2(projectID int64, userID string,
	batchedEvents [][]*model.Event, sessionID string, sessionEventNameID string) bool {
	logFields := log.Fields{"project_id": projectID, "session_id": sessionID,
		"session_event_name_id": sessionEventNameID, "user_id": userID}

	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)

	db := C.GetServices().Db
	dbTx := db.Begin()
	if dbTx.Error != nil {
		logCtx.WithError(dbTx.Error).Error("Failed to begin transaction in AssociateSessionByEventIdsInBatchV2.")
		return true
	}
	logCtx.Info("Using batch transaction in AssociateSessionByEventIdsInBatchV2.")

	hasFailure := false
	for i := range batchedEvents {
		events := batchedEvents[i]
		status := store.associateSessionByEventIdsWithTransaction(projectID, userID, events, sessionID, sessionEventNameID, dbTx)
		if status != http.StatusAccepted {
			logCtx.WithFields(log.Fields{"events": events, "session_id": sessionID,
				"session_event_name_id": sessionEventNameID}).
				Error("Failed to associate session id in event using AssociateSessionByEventIdsInBatch.")
			hasFailure = true
		}
	}

	err := dbTx.Commit().Error
	if err != nil {
		logCtx.WithError(err).Error("Failed to commit in associate session id in event using AssociateSessionByEventIdsInBatch.")
		hasFailure = true
	}

	return hasFailure
}

func (store *MemSQL) AssociateSessionByEventIds(projectID int64,
	userID string, events []*model.Event, sessionID string, sessionEventNameID string) int {
	db := C.GetServices().Db

	return store.associateSessionByEventIdsWithTransaction(projectID, userID, events, sessionID, sessionEventNameID, db)
}

func (store *MemSQL) associateSessionByEventIdsWithTransaction(projectId int64,
	userID string, events []*model.Event, sessionId string, sessionEventNameId string, dbTx *gorm.DB) int {
	logFields := log.Fields{
		"project_id":            projectId,
		"user_id":               userID,
		"events":                events,
		"session_id":            sessionId,
		"session_event_name_id": sessionEventNameId,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	fromTimestamp, toTimestamp, eventIds, eventNameIds := model.GetEventsMinMaxTimestampsAndEventnameIds(events)

	logCtx := log.WithFields(logFields)

	if projectId == 0 || len(eventIds) == 0 || sessionId == "" || userID == "" {
		logCtx.Error("Invalid args on associateSessionToEvents.")
		return http.StatusBadRequest
	}

	// Updates session_id to all events between given timestamp.
	updateFields := map[string]interface{}{"session_id": sessionId}
	EnableOLTPQueriesMemSQLImprovements := C.EnableOLTPQueriesMemSQLImprovements(projectId)
	var err error
	if EnableOLTPQueriesMemSQLImprovements {
		err = dbTx.Model(&model.Event{}).
			Where("project_id = ? AND user_id = ? AND id IN (?) AND timestamp >= ? AND timestamp <= ? AND event_name_id != ? AND event_name_id IN (?)",
				projectId, userID, eventIds, fromTimestamp, toTimestamp, sessionEventNameId, eventNameIds).Update(updateFields).Error
	} else {
		err = dbTx.Model(&model.Event{}).
			Where("project_id = ? AND user_id = ? AND id IN (?)", projectId, userID, eventIds).
			Update(updateFields).Error
	}
	if err != nil {
		logCtx.WithError(err).Error("Failed to associate session to events.")
		return http.StatusInternalServerError
	}

	// Log for analysis.
	log.WithField("project_id", projectId).WithField("tag", "update_event").Info("Updated event.")

	return http.StatusAccepted
}

func (store *MemSQL) associateSessionToEventsInBatch(projectId int64, userID string, events []*model.Event,
	sessionId string, batchSize int, sessionEventNameId string) int {
	logFields := log.Fields{
		"project_id":            projectId,
		"user_id":               userID,
		"events":                events,
		"session_id":            sessionId,
		"session_event_name_id": sessionEventNameId,
		"batch_size":            batchSize,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	if C.GetSessionBatchTransactionBatchSize() > 0 {
		batchEvents := model.GetEventListAsBatch(events, C.GetSessionBatchTransactionBatchSize())
		anyFailure := store.AssociateSessionByEventIdsBatchV2(projectId, userID, batchEvents, sessionId, sessionEventNameId)
		if anyFailure {
			log.WithFields(logFields).WithFields(log.Fields{"user_id": userID,
				"session_id": sessionId, "session_event_name_id": sessionEventNameId}).
				Error("Failed to AssociateSessionByEventIdsBatchV2.")
			return http.StatusInternalServerError
		}

		return http.StatusAccepted
	}

	batchEvents := model.GetEventListAsBatch(events, batchSize)
	for i := range batchEvents {

		errCode := store.AssociateSessionByEventIds(projectId, userID, batchEvents[i], sessionId, sessionEventNameId)
		if errCode != http.StatusAccepted {
			return errCode
		}
	}

	return http.StatusAccepted
}

func (store *MemSQL) DissociateEventsFromSession(projectID int64, events []model.Event, sessionID string) int {
	logFields := log.Fields{
		"project_id":   projectID,
		"total_events": len(events),
		"session_id":   sessionID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)
	if projectID == 0 || len(events) == 0 || sessionID == "" {
		logCtx.Error("Invalid parameter.")
		return http.StatusBadRequest
	}

	startTimestamp, endTimestamp := events[0].Timestamp, events[0].Timestamp
	eventIDs := make([]string, 0)
	for i := range events {
		startTimestamp = U.Min(startTimestamp, events[i].Timestamp)
		endTimestamp = U.Max(endTimestamp, events[i].Timestamp)
		eventIDs = append(eventIDs, events[i].ID)
	}

	updateFields := map[string]interface{}{"session_id": nil}
	db := C.GetServices().Db
	if err := db.Model(&model.Event{}).Where("project_id = ? AND id in ( ? ) and session_id = ? AND timestamp BETWEEN ? AND ? ",
		projectID, eventIDs, sessionID, startTimestamp, endTimestamp).Update(updateFields).Error; err != nil {
		logCtx.Error("Failed to DissociateEventFromSession.")
		return http.StatusInternalServerError
	}

	return http.StatusAccepted

}

// AddSessionForUser - Wrapper for addSessionForUser to handle creating
// new session for last event when new session conditions met.
func (store *MemSQL) AddSessionForUser(projectId int64, userId string, userEvents []model.Event,
	bufferTimeBeforeSessionCreateInSecs int64, sessionEventNameId string) (int, int, bool, int, int) {
	logFields := log.Fields{
		"project_id":  projectId,
		"user_id":     userId,
		"user_events": userEvents,
		"buffer_time_before_session_create_in_secs": bufferTimeBeforeSessionCreateInSecs,
		"session_event_name_id":                     sessionEventNameId,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	noOfFilteredEvents, noOfSessionsCreated, sessionContinuedFlag,
		noOfUserPropertiesUpdated, isLastEventToBeProcessed,
		errCode := store.addSessionForUser(projectId, userId, userEvents,
		bufferTimeBeforeSessionCreateInSecs, sessionEventNameId)

	if errCode == http.StatusInternalServerError || errCode == http.StatusBadRequest {
		return noOfFilteredEvents, noOfSessionsCreated, sessionContinuedFlag,
			noOfUserPropertiesUpdated, errCode
	}

	// Fix for last event not being processed when the last but previous meets
	// new session creation condition. Calling the add_session for user
	// with only last event, for simplicity.
	if isLastEventToBeProcessed {
		lastUserEventAsList := userEvents[len(userEvents)-1:]
		_, _, _, _, _, errCode = store.addSessionForUser(projectId, userId, lastUserEventAsList,
			bufferTimeBeforeSessionCreateInSecs, sessionEventNameId)

		noOfSessionsCreated++
	}

	return noOfFilteredEvents, noOfSessionsCreated, sessionContinuedFlag,
		noOfUserPropertiesUpdated, errCode
}

/*
addSessionForUser - Will add session event based on conditions and associate session to each event.
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
func (store *MemSQL) addSessionForUser(projectId int64, userId string, userEvents []model.Event,
	bufferTimeBeforeSessionCreateInSecs int64, sessionEventNameId string) (int, int, bool, int, bool, int) {
	logFields := log.Fields{
		"project_id":  projectId,
		"user_id":     userId,
		"user_events": userEvents,
		"buffer_time_before_session_create_in_secs": bufferTimeBeforeSessionCreateInSecs,
		"session_event_name_id":                     sessionEventNameId,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)
	if len(userEvents) == 0 {
		return 0, 0, false, 0, false, http.StatusNotModified
	}

	isV2Enabled := C.EnableUserLevelEventPullForAddSessionByProjectID(projectId)
	if isV2Enabled {
		eventWithSession, status := store.GetLastEventWithSessionByUser(projectId, userEvents[0].UserId, userEvents[0].Timestamp)
		if status == http.StatusFound && eventWithSession != nil {
			userEvents = model.PrependEvent(*eventWithSession, userEvents)
		}
	}

	startTimestamp := userEvents[0].Timestamp

	latestUserEvent := &userEvents[len(userEvents)-1]
	// User level buffer time. Mainly added for segment.
	endTimestamp := latestUserEvent.Timestamp - bufferTimeBeforeSessionCreateInSecs
	// Max buffer time should be current timestamp - configured buffer time.
	maxEndTimestamp := U.TimeNowUnix() - bufferTimeBeforeSessionCreateInSecs

	if endTimestamp < maxEndTimestamp || endTimestamp <= startTimestamp {
		endTimestamp = latestUserEvent.Timestamp
	}

	events := filterEventsForSession(userEvents, endTimestamp)
	if len(events) == 0 {
		return 0, 0, false, 0, false, http.StatusNotModified
	}

	project, errCode := store.GetProject(projectId)
	if errCode != http.StatusFound {
		logCtx.WithField("err_code", errCode).Error("Failed to get project on addSessionForUser")
		return 0, 0, false, 0, false, http.StatusNotModified
	}

	noOfFilteredEvents := len(events)

	sessionStartIndex := 0
	sessionEndIndex := 0

	isMatchingMktPropsOn := false
	noOfSessionsCreated := 0
	sessionContinuedFlag := false
	isLastEventToBeProcessed := false

	// user_properties_id would be the key till the user_properties table
	// is permanently deprecated and all the event level user_properties
	// moved to event itself.
	// map[id or user_properties_id of events] = session_user_properties
	sessionUserPropertiesRecordMap := make(map[string]model.SessionUserProperties, 0)

	// Use 2 moving cursor current, next. if diff(current, previous) > in-activity
	// period or has marketing property, use current_event - 1 as session end
	// and update. Update current_event as session start and do the same till the end.
	var currentSessionCandidateEvent model.Event
	isFirstEvent := true
	updateEventSessionUserPropertiesRecordMap := make(map[string]model.SessionUserProperties, 0)
	updateEventPropertiesParams := make([]model.UpdateEventPropertiesParams, 0)
	for i := 0; i < len(events); {

		if isFirstEvent {
			currentSessionCandidateEvent = *events[i]
			isFirstEvent = false
		}
		hasMarketingProperty, err := doesEventIsPageViewAndHasMarketingProperty(events[i])
		if err != nil {
			logCtx.WithError(err).
				Error("Failed to check marketing property on event properties.")
			return noOfFilteredEvents, noOfSessionsCreated, sessionContinuedFlag, 0,
				isLastEventToBeProcessed, http.StatusInternalServerError
		}

		isNewSessionRequired := (i == 0 && len(events) == 1) ||
			(i+1 < len(events) && ((events[i+1].Timestamp - events[i].Timestamp) > model.NewUserSessionInactivityInSeconds))

		// Balance events on the list after creating session for the previous set.
		isLastSetOfEvents := i == len(events)-1

		// This checks if the first event is having session and can we continue with that or not
		// It will do i++; continue; if session can be continued
		isStartingWthMarketingProperty := i == 0 && len(events) > 1 && hasMarketingProperty && (*events[i]).SessionId != nil && !isNewSessionRequired
		if isStartingWthMarketingProperty {
			i++
			continue
		}
		backMatch := false
		forwardMatch := false
		// Skip or Continue adding this event to the session
		if !isNewSessionRequired && (*events[i]).SessionId == nil {
			// Backward properties matching case
			if i-1 >= 0 && model.AreMarketingPropertiesMatching(*events[i-1], *events[i]) {
				hasMarketingProperty = false
				isMatchingMktPropsOn = true
				backMatch = true
			} else {
				backMatch = false
			}
			// Default case for first element
			if i == 0 {
				backMatch = true
			}
			// Forward properties matching case
			if i+1 < len(events) && model.AreMarketingPropertiesMatching(currentSessionCandidateEvent, *events[i+1]) {
				// continue with the next event in case next one has the same matching properties
				forwardMatch = true
				i++
				continue
			} else {
				forwardMatch = false
			}
		}
		if (hasMarketingProperty || isNewSessionRequired || isLastSetOfEvents || (backMatch && !forwardMatch)) && !(!backMatch && forwardMatch && i < len(events)-1) {
			isFirstEvent = true
			var sessionEvent *model.Event
			var isSessionContinued bool

			sessionEndIndex = i

			// Skip the associating previous session to last event, If it satisfies
			// new session condition. Instead of manipulating cursor, setting the
			// isLastEventToBeProcessed as true, to process it separately.
			if isLastSetOfEvents {
				if !(hasMarketingProperty || isNewSessionRequired) {
					sessionEndIndex = i
				} else {
					if len(events) > 1 && (sessionStartIndex != sessionEndIndex) {
						isLastEventToBeProcessed = true
					}
				}
			}

			// End condition for same marketing prop events.
			if i == len(events)-1 && isMatchingMktPropsOn {
				sessionEndIndex = i
				isLastEventToBeProcessed = false
			}

			// Continue with the last session_id, if available. This will be true as
			// first event will have max_timestamp (used as start_timestamp) where
			// session_id is not null.
			var existingSessionEvent *model.Event
			if events[sessionStartIndex].SessionId != nil {
				var errCode int
				existingSessionEvent, errCode = store.GetEventById(projectId,
					*events[sessionStartIndex].SessionId, *&events[sessionStartIndex].UserId)
				if errCode == http.StatusNotFound {
					// Log and continue with new session, if the session event is not found.
					logCtx.WithField("session_id", events[sessionStartIndex].SessionId).
						WithField("err_code", errCode).
						Error("Failed to find the session event associated.")

				} else if errCode != http.StatusFound {
					logCtx.WithField("err_code", errCode).Error(
						"Failed to get existing session using session id on add session.")
					return noOfFilteredEvents, noOfSessionsCreated, sessionContinuedFlag, 0,
						isLastEventToBeProcessed, http.StatusInternalServerError
				}
			}

			if existingSessionEvent != nil {
				sessionEvent = existingSessionEvent
				isSessionContinued = true
				sessionContinuedFlag = true
			} else {
				firstEvent := events[sessionStartIndex]

				logCtx := log.WithFields(logFields)
				var userPropertiesMap U.PropertiesMap
				isEmptyUserProperties := firstEvent.UserProperties == nil ||
					U.IsEmptyPostgresJsonb(firstEvent.UserProperties)
				if !isEmptyUserProperties {
					userPropertiesDecoded, err := U.DecodePostgresJsonb(firstEvent.UserProperties)
					if err != nil {
						logCtx.WithError(err).WithField("user_properties", firstEvent.UserProperties).
							Error("Failed to decode user properties of first event on session.")
						return noOfFilteredEvents, noOfSessionsCreated, sessionContinuedFlag, 0,
							isLastEventToBeProcessed, http.StatusInternalServerError
					}

					userPropertiesMap = U.PropertiesMap(*userPropertiesDecoded)
				} else {
					logCtx.Error("Empty first event user_properties.")
				}

				firstEventPropertiesDecoded, err := U.DecodePostgresJsonb(&firstEvent.Properties)
				if err != nil {
					logCtx.WithError(err).Error("Failed to decode event properties of first event on session.")
					return noOfFilteredEvents, noOfSessionsCreated, sessionContinuedFlag, 0,
						isLastEventToBeProcessed, http.StatusInternalServerError
				}
				firstEventPropertiesMap := U.PropertiesMap(*firstEventPropertiesDecoded)

				sessionEventCount, errCode := store.GetEventCountOfUserByEventName(projectId, userId, sessionEventNameId)
				if errCode == http.StatusInternalServerError {
					logCtx.WithField("err_code", errCode).Error("Failed to get session event count for user.")
					return noOfFilteredEvents, noOfSessionsCreated, sessionContinuedFlag,
						0, isLastEventToBeProcessed, errCode
				}
				isFirstSession := sessionEventCount == 0
				sessionPropertiesMap := U.GetSessionProperties(isFirstSession,
					&firstEventPropertiesMap, &userPropertiesMap)

				initialPageUrl, exists := (*sessionPropertiesMap)[U.SP_INITIAL_PAGE_URL]
				if exists {
					contentGroups := store.CheckURLContentGroupValue(initialPageUrl.(string), projectId)
					for key, value := range contentGroups {
						(*sessionPropertiesMap)[key] = value
					}
				}
				sessionPropertiesEncoded := map[string]interface{}(*sessionPropertiesMap)

				sessionPropertiesJsonb, err := U.EncodeToPostgresJsonb(&sessionPropertiesEncoded)
				if err != nil {
					logCtx.WithError(err).Error("Failed to encode session properties as postgres jsonb.")
					return noOfFilteredEvents, noOfSessionsCreated, sessionContinuedFlag, 0,
						isLastEventToBeProcessed, http.StatusInternalServerError
				}

				// session event properties, to be updated
				newSessionEvent, errCode := store.CreateEvent(&model.Event{
					EventNameId: sessionEventNameId,
					// Timestamp - 1sec before the first event of session.
					Timestamp:      firstEvent.Timestamp - 1,
					ProjectId:      projectId,
					UserId:         userId,
					UserProperties: firstEvent.UserProperties,
					Properties:     *sessionPropertiesJsonb,
				})

				if errCode != http.StatusCreated {
					logCtx.WithField("err_code", errCode).Error("Failed to create session event.")
					return noOfFilteredEvents, noOfSessionsCreated, sessionContinuedFlag,
						0, isLastEventToBeProcessed, errCode
				}
				isMatchingMktPropsOn = false
				sessionEvent = newSessionEvent
				noOfSessionsCreated++
			}

			eventsOfSession := events[sessionStartIndex : sessionEndIndex+1]
			channelOfSessionEvent := make(map[string]string)

			// Update the session_id to all events between start index and end index + 1.
			errCode := store.associateSessionToEventsInBatch(projectId, userId,
				eventsOfSession, sessionEvent.ID, 100, sessionEventNameId)
			if errCode == http.StatusInternalServerError {
				logCtx.WithField("err_code", errCode).Error("Failed to associate session to events.")
				return noOfFilteredEvents, noOfSessionsCreated, sessionContinuedFlag,
					0, isLastEventToBeProcessed, errCode
			}

			lastEventProperties, err := U.DecodePostgresJsonb(&events[sessionEndIndex].Properties)
			if err != nil {
				logCtx.WithError(err).Error("Failed to decode properties of last event of session.")
				return noOfFilteredEvents, noOfSessionsCreated, sessionContinuedFlag, 0,
					isLastEventToBeProcessed, http.StatusInternalServerError
			}

			sessionPropertiesMap := U.PropertiesMap{}
			if _, exists := (*lastEventProperties)[U.EP_PAGE_RAW_URL]; exists {
				sessionPropertiesMap[U.SP_LATEST_PAGE_RAW_URL] = (*lastEventProperties)[U.EP_PAGE_RAW_URL]
			} else {
				logCtx.WithField("EventId", sessionEvent.ID).Info("Missing SP_LATEST_PAGE_RAW_URL")
			}
			if _, exists := (*lastEventProperties)[U.EP_PAGE_URL]; exists {
				sessionPropertiesMap[U.SP_LATEST_PAGE_URL] = (*lastEventProperties)[U.EP_PAGE_URL]
			} else {
				logCtx.WithField("EventId", sessionEvent.ID).Info("Missing SP_LATEST_PAGE_URL")
			}

			// Using existing method to get count and page spent time.
			var sessionPageCount, sessionPageSpentTime, onlyThisSessionPageCount, onlyThisSessionPageSpentTime float64

			if isSessionContinued {
				// Using db query, since previous session continued, we don't have all the events of the session.
				sessionPageCount, sessionPageSpentTime, onlyThisSessionPageCount,
					onlyThisSessionPageSpentTime, errCode = getPageCountAndTimeSpentForContinuedSession(
					projectId, userId, sessionEvent, events[sessionStartIndex:sessionEndIndex+1])
				if errCode == http.StatusInternalServerError {
					logCtx.WithField("err_code", errCode).Error("Failed to get page count and spent time of session on add session.")
				}
			} else {
				// events from sessionStartIndex till i.
				sessionPageCount, sessionPageSpentTime =
					getPageCountAndTimeSpentFromEventsList(events[sessionStartIndex:sessionEndIndex+1], sessionEvent)
				onlyThisSessionPageCount, onlyThisSessionPageSpentTime = sessionPageCount, sessionPageSpentTime
			}

			if sessionPageCount > 0 {
				sessionPropertiesMap[U.SP_PAGE_COUNT] = sessionPageCount
			}
			if sessionPageSpentTime > 0 {
				sessionPropertiesMap[U.SP_SPENT_TIME] = sessionPageSpentTime
			}
			sessionEventProps, err := U.DecodePostgresJsonb(&sessionEvent.Properties)
			if err != nil {
				logCtx.WithError(err).Error("Failed to decode session event properties for adding channel property on add session")
			} else {
				channel, errString := model.GetChannelGroup(*project, *sessionEventProps)
				if errString != "" {
					logCtx.Error(errString)
				} else {
					sessionPropertiesMap[U.EP_CHANNEL] = channel

					// Note: All events of the session are from same channel, hence we update only session event and user properties
					if !isSessionContinued {
						channelOfSessionEvent[sessionEvent.ID] = channel
					}
				}
			}

			sessionPropertiesMap[U.EP_SESSION_COUNT] = sessionEvent.Count

			if C.GetSessionBatchTransactionBatchSize() > 0 {
				updateEventPropertiesParams = append(updateEventPropertiesParams,
					model.UpdateEventPropertiesParams{
						ProjectID:                     projectId,
						EventID:                       sessionEvent.ID,
						UserID:                        sessionEvent.UserId,
						SessionProperties:             &sessionPropertiesMap,
						SessionEventTimestamp:         sessionEvent.Timestamp,
						NewSessionEventUserProperties: sessionEvent.UserProperties,
						EventsOfSession:               eventsOfSession,
					})
			} else {
				// Update session event properties.
				errCode = store.UpdateEventProperties(projectId, sessionEvent.ID,
					sessionEvent.UserId, &sessionPropertiesMap, sessionEvent.Timestamp+1,
					sessionEvent.UserProperties)
				if errCode == http.StatusInternalServerError {
					logCtx.WithField("err_code", errCode).Error("Failed updating session event properties on add session.")
					return noOfFilteredEvents, noOfSessionsCreated, sessionContinuedFlag,
						0, isLastEventToBeProcessed, errCode
				}
			}

			// associate user_properties state using session of the event.
			for i := range eventsOfSession {

				userPropertiesRefID := eventsOfSession[i].ID
				sessionUserProperties := model.SessionUserProperties{
					UserID:                userId,
					SessionEventTimestamp: sessionEvent.Timestamp,
					SessionPageCount:      onlyThisSessionPageCount,
					SessionPageSpentTime:  onlyThisSessionPageSpentTime,
					SessionChannel:        channelOfSessionEvent[userPropertiesRefID],
					IsSessionEvent:        false,
					EventUserProperties:   eventsOfSession[i].UserProperties,
				}

				if C.GetSessionBatchTransactionBatchSize() > 0 {
					updateEventSessionUserPropertiesRecordMap[userPropertiesRefID] = sessionUserProperties
					continue
				}

				sessionUserPropertiesRecordMap[userPropertiesRefID] = sessionUserProperties
			}

			// doing this only for non-continued session because initial and latest channel property would already be set and we don't need to change that
			if !isSessionContinued {

				sessionUserPropertiesRecordMap[sessionEvent.ID] = model.SessionUserProperties{
					UserID:                userId,
					SessionChannel:        channelOfSessionEvent[sessionEvent.ID],
					EventUserProperties:   sessionEvent.UserProperties,
					IsSessionEvent:        true,
					SessionEventTimestamp: sessionEvent.Timestamp,
				}
			}
			sessionStartIndex = i + 1
		}
		i++
	}

	if C.GetSessionBatchTransactionBatchSize() > 0 {
		batchedUpdatedEventPropertiesParams := model.GetUpdateEventPropertiesParamsAsBatch(updateEventPropertiesParams,
			C.GetSessionBatchTransactionBatchSize())
		for batchIndex := range batchedUpdatedEventPropertiesParams {
			failure := store.UpdateEventPropertiesInBatch(projectId, batchedUpdatedEventPropertiesParams[batchIndex])
			if failure {
				logCtx.Error("Failed updating session event properties on add session in batch mode.")
				return noOfFilteredEvents, noOfSessionsCreated, sessionContinuedFlag,
					0, isLastEventToBeProcessed, errCode
			}

			for i := range batchedUpdatedEventPropertiesParams[batchIndex] {
				eventsOfSession := batchedUpdatedEventPropertiesParams[batchIndex][i].EventsOfSession
				for _, event := range eventsOfSession {
					userPropertiesRefID := event.ID
					sessionUserPropertiesRecordMap[userPropertiesRefID] = updateEventSessionUserPropertiesRecordMap[userPropertiesRefID]
				}
			}
		}
	}

	// Todo: The property values being updated are not accurate. Fix it.
	// Issue - https://github.com/Slashbit-Technologies/factors/issues/445
	errCode = store.UpdateUserPropertiesForSession(projectId, &sessionUserPropertiesRecordMap)
	if errCode != http.StatusAccepted {
		logCtx.WithField("err_code", errCode).
			Error("Failed to update user properties record for session.")
		return noOfFilteredEvents, noOfSessionsCreated, sessionContinuedFlag,
			0, isLastEventToBeProcessed, http.StatusInternalServerError
	}

	noOfUserPropertiesUpdated := len(sessionUserPropertiesRecordMap)

	return noOfFilteredEvents, noOfSessionsCreated, sessionContinuedFlag,
		noOfUserPropertiesUpdated, isLastEventToBeProcessed, http.StatusOK
}

// RemoveDisabledEventUserProperties remove disabled event level user properties
func RemoveDisabledEventUserProperties(ProjectId int64, userProperties *postgres.Jsonb) *postgres.Jsonb {

	newSessionEventUserPropertiesJsonb, err := U.RemoveFromJsonb(userProperties, U.DISABLED_EVENT_USER_LEVEL_PROPERTIES)
	if err != nil {
		// continue with event properties update skipping event_user_properties.
		newSessionEventUserPropertiesJsonb = nil
	}
	if C.DisableEventUserPropertiesByProjectID(ProjectId) {
		return newSessionEventUserPropertiesJsonb
	}
	return userProperties

}

// GetDatesForNextEventsArchivalBatch Get dates for events since startTime, excluding today's date.
func (store *MemSQL) GetDatesForNextEventsArchivalBatch(projectID int64, startTime int64) (map[string]int64, int) {
	defer model.LogOnSlowExecutionWithParams(time.Now(), nil)
	logFields := log.Fields{
		"project_id": projectID,
		"start_time": startTime,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	countByDates := make(map[string]int64)

	db := C.GetServices().Db
	rows, err := db.Model(&model.Event{}).
		Where("project_id = ? AND timestamp BETWEEN ? AND (UNIX_TIMESTAMP(CURRENT_DATE()) - 1)", projectID, startTime).
		Group("date(FROM_UNIXTIME(timestamp))").
		Select("date(FROM_UNIXTIME(timestamp)), count(*)").Rows()
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

func (store *MemSQL) GetNextSessionEventInfoFromDB(projectID int64, withSession bool,
	sessionEventNameId uint64, maxLookbackTimestamp int64) (int64, int) {
	logFields := log.Fields{
		"project_id":             projectID,
		"with_session":           withSession,
		"max_lookback_timestamp": maxLookbackTimestamp,
		"session_event_name_id":  sessionEventNameId,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	sessionExistStr := "NOT NULL"
	startTimestampAggrFunc := "max"
	if !withSession {
		sessionExistStr = "NULL"
		startTimestampAggrFunc = "min"
	}
	selectStmnt := fmt.Sprintf("%s(timestamp) as start_timestamp",
		startTimestampAggrFunc)

	db := C.GetServices().Db
	query := db.Table("events").
		Where("project_id = ? AND event_name_id != ?", projectID, sessionEventNameId).
		Where(fmt.Sprintf("session_id IS %s AND JSON_EXTRACT_STRING(properties, '%s') IS NULL",
			sessionExistStr, U.EP_SKIP_SESSION)).
		Select(selectStmnt)

	if maxLookbackTimestamp > 0 {
		query = query.Where("timestamp > ?", maxLookbackTimestamp)
	}

	rows, err := query.Rows()
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return 0, http.StatusNotFound
		}

		log.WithField("project_id", projectID).WithError(err).
			Error("Failed to get next session start timestamp for project.")
		return 0, http.StatusInternalServerError
	}
	defer rows.Close()

	var startTimestamp *int64
	for rows.Next() {
		err = rows.Scan(&startTimestamp)
		if err != nil {
			log.WithError(err).Error("Failed to read next session start timestamp.")
			return 0, http.StatusInternalServerError
		}
	}

	if startTimestamp == nil {
		return 0, http.StatusNotFound
	}

	return *startTimestamp, http.StatusFound
}

func (store *MemSQL) GetLastSessionEventTimestamp(projectID int64, sessionEventNameID uint64) (int64, int) {
	logFields := log.Fields{
		"project_id":            projectID,
		"session_event_name_id": sessionEventNameID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)
	// This is a faster query.
	// ORDER BY project_id, event_name_id, timestamp DESC is used to instead of
	// MIN to avoid ordering and use the ordered index on that column.
	db := C.GetServices().Db
	query := db.Raw("SELECT timestamp FROM events WHERE project_id = ? AND event_name_id = ? ORDER BY project_id, event_name_id, timestamp DESC LIMIT 1",
		projectID, sessionEventNameID)
	rows, err := query.Rows()
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return 0, http.StatusNotFound
		}

		logCtx.WithError(err).Error("SQL Query failed")
		return 0, http.StatusInternalServerError
	}
	defer rows.Close()

	var startTimestamp *int64
	for rows.Next() {
		err = rows.Scan(&startTimestamp)
		if err != nil {
			log.WithError(err).Error("Failed to read on get last event timestamp.")
			return 0, http.StatusInternalServerError
		}
	}

	if startTimestamp == nil {
		return 0, http.StatusNotFound
	}

	return *startTimestamp, http.StatusFound
}

// GetAllEventsForSessionCreationAsUserEventsMap - Returns a map of user:[events...] withing given period,
// excluding session event and event with session_id.
func (store *MemSQL) GetAllEventsForSessionCreationAsUserEventsMap(projectId int64, sessionEventNameId string,
	startTimestamp, endTimestamp int64) (*map[string][]model.Event, int, int) {
	logFields := log.Fields{
		"project_id":            projectId,
		"session_event_name_id": sessionEventNameId,
		"start_timestamp":       startTimestamp,
		"end_timestamp":         endTimestamp,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	defer model.LogOnSlowExecutionWithParams(time.Now(), nil)

	logCtx := log.WithFields(logFields)

	var userEventsMap map[string][]model.Event
	var events []model.Event
	if startTimestamp == 0 || endTimestamp == 0 {
		logCtx.Error("Invalid start_timestamp or end_timestamp.")
		return &userEventsMap, 0, http.StatusInternalServerError
	}

	queryStartTime := U.TimeNowUnix()
	db := C.GetServices().Db
	// Ordered by timestamp, created_at to fix the order for events with same
	// timestamp, as event timestamp is in seconds. This fixes the invalid first
	// event used for enrichment.
	excludeSkipSessionCondition := fmt.Sprintf("(JSON_EXTRACT_STRING(properties, '%s') IS NULL OR JSON_EXTRACT_STRING(properties, '%s') = 'f')",
		U.EP_SKIP_SESSION, U.EP_SKIP_SESSION)
	if err := db.Where("project_id = ? AND event_name_id != ? AND timestamp BETWEEN ? AND ?"+" AND "+excludeSkipSessionCondition,
		projectId, sessionEventNameId, startTimestamp, endTimestamp).
		Find(&events).Error; err != nil {

		logCtx.WithError(err).Error("Failed to get all events of project.")
		return &userEventsMap, 0, http.StatusInternalServerError
	}

	if len(events) == 0 {
		return &userEventsMap, 0, http.StatusNotFound
	}

	queryTimeInSecs := U.TimeNowUnix() - queryStartTime
	logCtx = logCtx.WithField("no_of_events", len(events)).
		WithField("time_taken_in_secs", queryTimeInSecs)

	if queryTimeInSecs >= (2 * 60) {
		logCtx.Error("Too much time taken to download events on get_all_events_as_user_map.")
	}

	// sort by timestamp, created_at ASC
	sort.Slice(events, func(i, j int) bool {
		if events[i].Timestamp != events[j].Timestamp {
			return events[i].Timestamp < events[j].Timestamp
		}

		return events[i].CreatedAt.Before(events[j].CreatedAt)
	})

	userEventsMap = make(map[string][]model.Event)
	for i := range events {
		if _, exists := userEventsMap[events[i].UserId]; !exists {
			userEventsMap[events[i].UserId] = make([]model.Event, 0, 0)
		} else {
			// Event with session should be added as first event, if available.
			// To support continuation of the session.
			currentUserEventHasSession := events[i].SessionId != nil
			firstUserEventHasSession := userEventsMap[events[i].UserId][0].SessionId != nil
			userHasNoSessionEvent := firstUserEventHasSession && len(userEventsMap[events[i].UserId]) > 1
			if !userHasNoSessionEvent && firstUserEventHasSession && currentUserEventHasSession {
				// Add current event as first event.
				userEventsMap[events[i].UserId][0] = events[i]
				continue
			}
		}

		userEventsMap[events[i].UserId] = append(userEventsMap[events[i].UserId], events[i])
	}

	logCtx.WithField("no_of_users", len(userEventsMap)).
		Info("Got all events on get_all_events_as_user_map.")

	return &userEventsMap, len(events), http.StatusFound
}

func doesPropertiesMapHaveKeys(propertiesMap U.PropertiesMap, keys []string) (bool, bool, U.PropertiesMap) {
	logFields := log.Fields{
		"properties_map": propertiesMap,
		"keys":           keys,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	filteredPropertiesMap := U.PropertiesMap{}

	if propertiesMap == nil {
		return false, false, filteredPropertiesMap
	}

	for i := range keys {
		value, exists := propertiesMap[keys[i]]
		if exists && value != nil && value != "" {
			filteredPropertiesMap[keys[i]] = value
		}
	}

	hasAll := len(filteredPropertiesMap) == len(keys)
	hasSome := len(filteredPropertiesMap) > 0 && len(filteredPropertiesMap) < len(keys)

	return hasAll, hasSome, filteredPropertiesMap
}

func getPropertiesByNameAndMaxOccurrence(
	propertiesByNameAndOccurence *map[string]map[string]*model.EventPropertiesWithCount,
) *map[string]U.PropertiesMap {
	logFields := log.Fields{
		"proiperties_by_name_ans_occurence": propertiesByNameAndOccurence,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	propertiesWithCount := make(map[string]model.EventPropertiesWithCount, 0)
	for name, propertiesByAuthor := range *propertiesByNameAndOccurence {
		for _, pwc := range propertiesByAuthor {
			// Select the poroeprties with max occurrence count.
			if (*pwc).Count > propertiesWithCount[name].Count &&
				// Consider only max no.of properties available.
				len((*pwc).Properties) >= len(propertiesWithCount[name].Properties) {

				propertiesWithCount[name] = *pwc
			}
		}
	}

	propertiesByName := make(map[string]U.PropertiesMap)
	for name, pwc := range propertiesWithCount {
		if pwc.Count > 0 && len(pwc.Properties) > 0 {
			propertiesByName[name] = pwc.Properties
		}
	}

	return &propertiesByName
}

// GetEventsByEventNameId return all events in given time frame for the event name.
// CAUTION: Predict no. of events this can pull for given range and then use this method.
func (store *MemSQL) GetEventsByEventNameId(projectId int64, eventNameId string,
	startTimestamp int64, endTimestamp int64) ([]model.Event, int) {
	logFields := log.Fields{
		"project_id":      projectId,
		"event_name_id":   eventNameId,
		"start_timestamp": startTimestamp,
		"end_timestamp":   endTimestamp,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	if startTimestamp == 0 || endTimestamp == 0 {
		return nil, http.StatusBadRequest
	}

	var events []model.Event

	db := C.GetServices().Db
	if err := db.Limit(1).Order("timestamp desc").Where(
		"project_id = ? AND event_name_id = ? AND timestamp > ? AND timestamp <= ?",
		projectId, eventNameId, startTimestamp, endTimestamp).Find(&events).Error; err != nil {

		return nil, http.StatusInternalServerError
	}

	if len(events) == 0 {
		return nil, http.StatusNotFound
	}

	return events, http.StatusFound
}

func (store *MemSQL) GetEventsByEventNameIDANDTimeRange(projectID int64, eventNameID string,
	startTimestamp int64, endTimestamp int64) ([]model.Event, int) {
	logFields := log.Fields{
		"project_id":      projectID,
		"event_name_id":   eventNameID,
		"start_timestamp": startTimestamp,
		"end_timestamp":   endTimestamp,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)
	if projectID == 0 || eventNameID == "" || startTimestamp == 0 || endTimestamp == 0 {
		logCtx.Error("Invalid parameters.")
		return nil, http.StatusBadRequest
	}

	var events []model.Event

	db := C.GetServices().Db
	if err := db.Order("timestamp").Where(
		"project_id = ? AND event_name_id = ? AND timestamp BETWEEN ? AND ? ",
		projectID, eventNameID, startTimestamp, endTimestamp).Find(&events).Error; err != nil {

		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}

		logCtx.WithError(err).Error("Failed to get events by event name id in time range.")
		return nil, http.StatusInternalServerError
	}

	if len(events) == 0 {
		return nil, http.StatusNotFound
	}

	return events, http.StatusFound
}

// GetEventsWithoutPropertiesAndWithPropertiesByName - Use for getting properties with and without values
// and use it for updating the events which doesn't have the values. User for fixing data for YourStory.
func (store *MemSQL) GetEventsWithoutPropertiesAndWithPropertiesByNameForYourStory(projectID int64, from,
	to int64, mandatoryProperties []string) ([]model.EventWithProperties, *map[string]U.PropertiesMap, int) {
	logFields := log.Fields{
		"project_id":           projectID,
		"from":                 from,
		"to":                   to,
		"mandatory_properties": mandatoryProperties,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)

	eventsWithoutProperties := make([]model.EventWithProperties, 0, 0)
	// map[event_name]map[authorName]*PropertiesWithCount
	propertiesByNameAndOccurence := make(map[string]map[string]*model.EventPropertiesWithCount, 0)

	queryStartTimestamp := U.TimeNowUnix()
	// LIKE '%.%' is for excluding custom event_names which are not urls.
	queryStmnt := "SELECT events.id, name, properties FROM events" + " " +
		"LEFT JOIN event_names ON events.event_name_id = event_names.id" + " " +
		"WHERE events.project_id = ? AND event_names.name != '$session'" + " " +
		"AND event_names.name LIKE '%.%' AND timestamp BETWEEN ? AND ?"

	db := C.GetServices().Db
	rows, err := db.Raw(queryStmnt, projectID, from, to).Rows()
	if err != nil {
		logCtx.WithError(err).
			Error("Failed to execute raw query on getEventsWithoutPropertiesAndWithPropertiesByName.")
		return eventsWithoutProperties, nil, http.StatusInternalServerError
	}
	defer rows.Close()
	logCtx = logCtx.WithField("query_exec_time_in_secs", U.TimeNowUnix()-queryStartTimestamp)

	var rowCount int
	for rows.Next() {
		var id string
		var name string
		var properties postgres.Jsonb

		err = rows.Scan(&id, &name, &properties)
		if err != nil {
			logCtx.WithError(err).Error("Failed to scan row.")
			continue
		}

		propertiesMap, err := U.DecodePostgresJsonbAsPropertiesMap(&properties)
		if err != nil {
			logCtx.WithError(err).Error("Failed to decode properties.")
			continue
		}

		if _, exists := propertiesByNameAndOccurence[name]; !exists {
			propertiesByNameAndOccurence[name] = make(map[string]*model.EventPropertiesWithCount, 0)
		}

		hasAll, hasSome, filteredPropertiesMap := doesPropertiesMapHaveKeys(*propertiesMap, mandatoryProperties)
		if hasAll {
			authorName, asserted := filteredPropertiesMap["authorName"].(string)
			if !asserted {
				log.WithField("author", authorName).Warn("Failed to assert author name as string.")
				continue
			}

			if _, exists := propertiesByNameAndOccurence[name][authorName]; !exists {
				propertiesByNameAndOccurence[name][authorName] = &model.EventPropertiesWithCount{
					Properties: filteredPropertiesMap,
					Count:      1,
				}
			} else {
				// Always overwrite, to keep adding hasAll state.
				(*propertiesByNameAndOccurence[name][authorName]).Properties = filteredPropertiesMap
				(*propertiesByNameAndOccurence[name][authorName]).Count++
			}
		}

		if hasSome {
			propAuthorName, exists := filteredPropertiesMap["authorName"]
			if !exists && propAuthorName == nil {
				continue
			}
			authorName := propAuthorName.(string)

			if propertiesWithCount, authorExists := propertiesByNameAndOccurence[name][authorName]; !authorExists {
				propertiesByNameAndOccurence[name][authorName] = &model.EventPropertiesWithCount{
					Properties: filteredPropertiesMap,
					Count:      1,
				}
			} else {
				// Do no overwrite, hasAll state with hasSome state.
				if allKeysExist, _, _ := doesPropertiesMapHaveKeys((*propertiesWithCount).Properties,
					mandatoryProperties); allKeysExist {
					continue
				}

				// Add properties if more properties available this time.
				if len(filteredPropertiesMap) > len((*propertiesWithCount).Properties) {
					(*propertiesByNameAndOccurence[name][authorName]).Properties = filteredPropertiesMap
				}
				(*propertiesByNameAndOccurence[name][authorName]).Count++
			}
		}

		// Adds all events for update, to support update with most occurrence.
		eventsWithoutProperties = append(
			eventsWithoutProperties,
			model.EventWithProperties{
				ID:            id,
				Name:          name,
				PropertiesMap: *propertiesMap,
			},
		)

		rowCount++
	}

	propertiesByName := getPropertiesByNameAndMaxOccurrence(&propertiesByNameAndOccurence)

	logCtx.WithField("rows", rowCount).Info("Scanned all rows.")
	return eventsWithoutProperties, propertiesByName, http.StatusFound
}

func (store *MemSQL) GetUnusedSessionIDsForJob(projectID int64, startTimestamp, endTimestamp int64) ([]string, int) {
	logFields := log.Fields{
		"project_id":      projectID,
		"start_timestamp": startTimestamp,
		"end_timesstamp":  endTimestamp,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)

	var unusedSessions []string
	if projectID == 0 || startTimestamp == 0 || endTimestamp == 0 {
		logCtx.Error("Invalid params.")
		return unusedSessions, http.StatusInternalServerError
	}

	if startTimestamp >= endTimestamp {
		logCtx.Error("Start timestamp should not be greater or equal to end timestamp")
		return unusedSessions, http.StatusInternalServerError
	}

	sessionEventName, errCode := store.GetSessionEventName(projectID)
	if errCode != http.StatusFound {
		logCtx.WithField("err_code", errCode).Error("Failed to get session event_name.")
		return unusedSessions, http.StatusInternalServerError
	}

	db := C.GetServices().Db
	queryStmnt := "SELECT id, session_id, event_name_id FROM events WHERE project_id = ? AND timestamp BETWEEN ? AND ?"
	rows, err := db.Raw(queryStmnt, projectID, startTimestamp, endTimestamp).Rows()
	if err != nil {
		logCtx.WithError(err).Error("Failed to get events.")
		return unusedSessions, http.StatusInternalServerError
	}
	defer rows.Close()

	usedSessionIDMap := make(map[string]bool, 0)
	allSessionIDs := make([]string, 0, 0)
	for rows.Next() {
		var event model.Event
		if err := db.ScanRows(rows, &event); err != nil {
			logCtx.WithError(err).Error("Failed scanning event rows.")
		}

		// session_ids associated to event.
		if event.SessionId != nil && *event.SessionId != "" &&
			event.EventNameId != sessionEventName.ID {

			usedSessionIDMap[*event.SessionId] = true
		}

		// all session events.
		if event.EventNameId == sessionEventName.ID {
			allSessionIDs = append(allSessionIDs, event.ID)
		}
	}

	unusedSessionIDMap := make(map[string]bool, 0)
	for i := range allSessionIDs {
		if _, exists := usedSessionIDMap[allSessionIDs[i]]; !exists {
			unusedSessionIDMap[allSessionIDs[i]] = true
		}
	}

	unusedSessions = make([]string, len(unusedSessionIDMap), len(unusedSessionIDMap))
	for sessionID := range unusedSessionIDMap {
		unusedSessions = append(unusedSessions, sessionID)
	}

	return unusedSessions, http.StatusFound
}

func (store *MemSQL) DeleteEventsByIDsInBatchForJob(projectID int64, eventNameID string, ids []string, batchSize int) int {
	logFields := log.Fields{
		"project_id":    projectID,
		"event_name_id": eventNameID,
		"ids":           ids,
		"batch_size":    batchSize,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)
	if projectID == 0 || batchSize == 0 {
		logCtx.Error("Invalid params.")
		return http.StatusInternalServerError
	}

	batches := util.GetStringListAsBatch(ids, batchSize)
	for i := range batches {
		errCode := store.DeleteEventByIDs(projectID, eventNameID, batches[i])
		if errCode != http.StatusAccepted {
			return errCode
		}

		// Logging for analysis, as this method used only on jobs.
		logCtx.WithField("batch_count", i+1).Info("Deleted batch.")
	}

	return http.StatusAccepted
}

func (store *MemSQL) DeleteEventByIDs(projectID int64, eventNameID string, ids []string) int {
	logFields := log.Fields{
		"project_id":    projectID,
		"event_name_id": eventNameID,
		"ids":           ids,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)
	db := C.GetServices().Db
	exec := db.Where("project_id = ? AND id IN (?)", projectID, ids).Delete(&model.Event{})
	if err := exec.Error; err != nil {
		logCtx.WithError(err).Error("Failed to delete session events.")
		return http.StatusInternalServerError
	}

	logCtx.WithField("no_of_ids", len(ids)).
		WithField("ids", ids).
		WithField("rows_affected", exec.RowsAffected).
		Info("Deleted events by id.")

	return http.StatusAccepted
}

func (store *MemSQL) OverwriteEventUserPropertiesByID(projectID int64, userID,
	id string, userProperties *postgres.Jsonb) int {
	logFields := log.Fields{
		"project_id":      projectID,
		"user_id":         userID,
		"id":              id,
		"user_properties": userProperties,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	newUserProperties := RemoveDisabledEventUserProperties(projectID, userProperties)

	logCtx := log.WithFields(logFields)
	if projectID == 0 || id == "" {
		logCtx.Error("Invalid values for arguments.")
		return http.StatusBadRequest
	}

	if newUserProperties == nil || U.IsEmptyPostgresJsonb(newUserProperties) {
		logCtx.Error("Failed to overwrite user_properties. Empty or nil properties.")
		return http.StatusBadRequest
	}

	// Not updating the event_user_properties
	db := C.GetServices().Db
	dbx := db.Model(&model.Event{}).Where("project_id = ? AND id = ?", projectID, id)
	if userID != "" {
		dbx = dbx.Where("user_id = ?", userID)
	}

	err := dbx.Update("user_properties", newUserProperties).Error
	if err != nil {
		logCtx.WithError(err).Error("Failed to overwrite user properteis.")
		return http.StatusInternalServerError
	}

	// Log for analysis.
	log.WithField("project_id", projectID).WithField("tag", "update_event").Info("Updated event.")

	return http.StatusAccepted
}

// PullEventRows - Function to pull events for factors model building sequentially.
func (store *MemSQL) PullEventRowsV2(projectID int64, startTime, endTime int64) (*sql.Rows, *sql.Tx, error) {
	logFields := log.Fields{
		"project_id": projectID,
		"start_time": startTime,
		"end_time":   endTime,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	rawQuery := fmt.Sprintf("SELECT COALESCE(users.customer_user_id, users.id), event_names.name, events.timestamp, events.count,"+
		" events.properties, users.join_timestamp, events.user_properties, users.is_group_user , users.group_1_user_id, users.group_2_user_id, users.group_3_user_id, users.group_4_user_id,"+
		" users.group_5_user_id, users.group_6_user_id, users.group_7_user_id, users.group_8_user_id,users.group_1_id, users.group_2_id, users.group_3_id, users.group_4_id,users.group_5_id,"+
		" users.group_6_id, users.group_7_id, users.group_8_id FROM events "+
		"LEFT JOIN event_names ON events.event_name_id = event_names.id "+
		"LEFT JOIN users ON events.user_id = users.id AND users.project_id = %d "+
		"WHERE events.project_id = %d AND UNIX_TIMESTAMP(events.created_at) BETWEEN %d AND %d "+
		"LIMIT %d",
		projectID, projectID, startTime, endTime, model.EventsPullLimit+1)

	rows, tx, err, _ := store.ExecQueryWithContext(rawQuery, []interface{}{})
	return rows, tx, err
}

// PullEventRows - Function to pull events for factors model building sequentially.
func (store *MemSQL) PullEventRowsV1(projectID int64, startTime, endTime int64) (*sql.Rows, *sql.Tx, error) {
	logFields := log.Fields{
		"project_id": projectID,
		"start_time": startTime,
		"end_time":   endTime,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	rawQuery := fmt.Sprintf("SELECT COALESCE(users.customer_user_id, users.id), event_names.name, events.timestamp, events.count,"+
		" events.properties, users.join_timestamp, events.user_properties FROM events "+
		"LEFT JOIN event_names ON events.event_name_id = event_names.id "+
		"LEFT JOIN users ON events.user_id = users.id AND users.project_id = %d "+
		"WHERE events.project_id = %d AND events.timestamp BETWEEN  %d AND %d "+
		"ORDER BY COALESCE(users.customer_user_id, users.id), events.timestamp LIMIT %d",
		projectID, projectID, startTime, endTime, model.EventsPullLimit+1)

	rows, tx, err, _ := store.ExecQueryWithContext(rawQuery, []interface{}{})
	return rows, tx, err
}

// PullEventsForArchivalJob - Function to pull events for archival.
func (store *MemSQL) PullEventRowsForArchivalJob(projectID int64, startTime, endTime int64) (*sql.Rows, *sql.Tx, error) {
	logFields := log.Fields{
		"project_id": projectID,
		"start_time": startTime,
		"end_time":   endTime,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	rawQuery := fmt.Sprintf("SELECT events.id, users.id, users.customer_user_id, "+
		"event_names.name, events.timestamp, events.session_id, events.properties, users.join_timestamp, events.user_properties FROM events "+
		"LEFT JOIN event_names ON events.event_name_id = event_names.id "+
		"LEFT JOIN users ON events.user_id = users.id AND users.project_id = %d "+
		"WHERE events.project_id = %d AND events.timestamp BETWEEN %d AND %d", projectID, projectID, startTime, endTime)

	rows, tx, err, _ := store.ExecQueryWithContext(rawQuery, []interface{}{})
	return rows, tx, err
}

// IsSmartEventAlreadyExist verify for exisitng smart event by same reference event id and timestamp for the user id
func (store *MemSQL) IsSmartEventAlreadyExist(projectID int64, userID, eventNameID, referenceEventID string,
	eventTimestamp int64) (bool, error) {
	db := C.GetServices().Db

	var event model.Event
	err := db.Where("project_id = ? AND user_id = ? and event_name_id = ? "+
		" and timestamp = ? AND JSON_EXTRACT_STRING(properties,?) = ?",
		projectID, userID, eventNameID, eventTimestamp, util.EP_CRM_REFERENCE_EVENT_ID, referenceEventID).
		Select("id").Limit(1).Find(&event).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return false, nil
		}
		return false, err
	}
	if event.ID == "" {
		return false, nil
	}

	return true, nil
}

/*
SELECT * FROM events WHERE project_id = ? AND user_id = ? AND session_id IS NOT NULL AND timestamp < ?
ORDER BY timestamp, created_at DESC LIMIT 1;
*/
func (store *MemSQL) GetLastEventWithSessionByUser(projectId int64, userId string, firstEventTimestamp int64) (*model.Event, int) {
	logFields := log.Fields{
		"project_id":            projectId,
		"user_id":               userId,
		"first_event_timestamp": firstEventTimestamp,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	if projectId == 0 || userId == "" || firstEventTimestamp == 0 {
		return nil, http.StatusBadRequest
	}

	var event model.Event

	// Max window size by the inactivity period allowed for session continuation.
	// We don't consider last event with session, which is older than the inactivity period.
	startTimestamp := (firstEventTimestamp - model.NewUserSessionInactivityInSeconds) + 2

	db := C.GetServices().Db
	if err := db.Limit(1).Order("timestamp, created_at DESC").
		Where("project_id = ? AND user_id = ? AND timestamp > ? AND timestamp < ? AND session_id IS NOT NULL",
			projectId, userId, startTimestamp, firstEventTimestamp).
		Find(&event).Error; err != nil {

		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}

		log.WithFields(logFields).WithError(err).Error("Getting event failed on GetLatestEventWithSessionForUser")
		return nil, http.StatusInternalServerError
	}

	return &event, http.StatusFound
}

// GetAllEventsForSessionCreationAsUserEventsMapV2 - Returns a map of user:[events...] within a given period,
// excluding session event and event with session_id.
/*
	SELECT * FROM events WHERE project_id = ? AND event_name_id != ? AND session_id IS NULL AND
	timestamp BETWEEN ? AND ? AND JSON_EXTRACT_STRING(properties, '$skip_session') IS NULL
	ORDER BY timestamp, created_at ASC LIMIT 50000;
*/
func (store *MemSQL) GetAllEventsForSessionCreationAsUserEventsMapV2(projectId int64, sessionEventNameId string, startTimestamp int64,
	endTimestamp int64) (*map[string][]model.Event, int, int) {
	logFields := log.Fields{
		"project_id":            projectId,
		"session_event_name_id": sessionEventNameId,
		"start_timestamp":       startTimestamp,
		"end_timestamp":         endTimestamp,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	defer model.LogOnSlowExecutionWithParams(time.Now(), nil)

	logCtx := log.WithFields(logFields)

	var userEventsMap map[string][]model.Event
	var events []model.Event

	if projectId == 0 || sessionEventNameId == "" {
		return &userEventsMap, 0, http.StatusBadRequest
	}

	eventsPullLimit := C.GetConfig().EventsPullMaxLimit

	queryStartTime := U.TimeNowUnix()
	db := C.GetServices().Db
	// Ordered by timestamp, created_at to fix the order for events with same timestamp, as event timestamp is in seconds.
	// This fixes the invalid first event used for enrichment.
	excludeSkipSessionCondition := fmt.Sprintf("JSON_EXTRACT_STRING(properties, '%s') IS NULL", U.EP_SKIP_SESSION)
	if err := db.Limit(eventsPullLimit).Order("timestamp, created_at ASC").
		Where("project_id = ? AND event_name_id != ? AND session_id IS NULL AND timestamp BETWEEN ? AND ?"+" AND "+excludeSkipSessionCondition,
			projectId, sessionEventNameId, startTimestamp, endTimestamp).Find(&events).Error; err != nil {

		logCtx.WithError(err).Error("Failed to get all events of project.")
		return &userEventsMap, 0, http.StatusInternalServerError
	}

	if len(events) == eventsPullLimit {
		logCtx.Error("Project event pull reached threshold")
	}

	if len(events) == 0 {
		return &userEventsMap, 0, http.StatusNotFound
	}

	queryTimeInSecs := U.TimeNowUnix() - queryStartTime
	logCtx = logCtx.WithField("no_of_events", len(events)).WithField("time_taken_in_secs", queryTimeInSecs)

	if queryTimeInSecs >= (2 * 60) {
		logCtx.Error("Too much time taken to download events on get_all_events_as_user_map_v2.")
	}

	userEventsMap = make(map[string][]model.Event)
	for i := range events {
		if _, exists := userEventsMap[events[i].UserId]; !exists {
			userEventsMap[events[i].UserId] = make([]model.Event, 0, 0)
		}

		userEventsMap[events[i].UserId] = append(userEventsMap[events[i].UserId], events[i])
	}

	logCtx.WithFields(log.Fields{"no_of_users": len(userEventsMap)}).Info("Got all events on get_all_events_as_user_map_v2.")

	return &userEventsMap, len(events), http.StatusFound
}

func (store *MemSQL) GetUserIdFromEventId(projectID int64, id string, userId string) (string, string, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"id":         id,
		"user_id":    userId,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)
	if projectID == 0 || id == "" {
		logCtx.Error("Invalid parameters.")
		return "", "", http.StatusBadRequest
	}

	db := C.GetServices().Db
	dbx := db.Where("project_id = ? AND id = ?", projectID, id)
	if userId != "" {
		dbx = dbx.Where("user_id = ?", userId)
	}

	var event model.Event
	if err := dbx.Limit(1).Select("id, user_id").Find(&event).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return "", "", http.StatusNotFound
		}

		logCtx.WithError(err).Error("Getting user_id failed on GetUserIdFromEventById")
		return "", "", http.StatusInternalServerError
	}
	return event.ID, event.UserId, http.StatusFound
}

func (store *MemSQL) GetEventsBySessionEvent(projectID int64, sessionEventID, userID string) ([]model.Event, int) {
	logFields := log.Fields{
		"project_id":       projectID,
		"session_event_id": sessionEventID,
		"user_id":          userID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)
	if projectID == 0 || sessionEventID == "" || userID == "" {
		logCtx.Error("Invalid parameters.")
		return nil, http.StatusBadRequest
	}

	db := C.GetServices().Db

	var events []model.Event
	if err := db.Where("project_id = ? AND session_id = ? AND user_id = ? ", projectID, sessionEventID, userID).Find(&events).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}

		logCtx.WithError(err).Error("Getting user_id failed on GetUserIdFromEventById")
		return nil, http.StatusInternalServerError
	}

	if len(events) == 0 {
		return nil, http.StatusNotFound
	}
	return events, http.StatusFound
}

func (store *MemSQL) DeleteSessionsAndAssociationForTimerange(projectID, startTimestamp, endTimestamp int64) (int64, int64, int) {

	logCtx := log.WithField("project_id", projectID).WithField("start_timestamp", startTimestamp).WithField("end_timestamp", endTimestamp)

	db := C.GetServices().Db
	delSessionsSQL := "DELETE FROM events WHERE project_id=? AND event_name_id=(SELECT id FROM event_names WHERE project_id=? AND name=? LIMIT 1) AND timestamp BETWEEN ? AND ?"
	delSessionExec := db.Exec(delSessionsSQL, projectID, projectID, U.EVENT_NAME_SESSION, startTimestamp, endTimestamp)
	if delSessionExec.Error != nil {
		logCtx.WithError(delSessionExec.Error).Error("Failed to delete session events.")
		return 0, 0, http.StatusInternalServerError
	}

	removeAssociationsSQL := "UPDATE events SET session_id=NULL WHERE project_id=? AND timestamp between ? AND ? AND session_id IS NOT NULL"
	removeAssociationsExec := db.Raw(removeAssociationsSQL, projectID, startTimestamp, endTimestamp)
	if removeAssociationsExec.Error != nil {
		logCtx.WithError(removeAssociationsExec.Error).Error("Failed to delete session associations.")
		return delSessionExec.RowsAffected, 0, http.StatusInternalServerError
	}

	return delSessionExec.RowsAffected, removeAssociationsExec.RowsAffected, http.StatusAccepted
}

func (store *MemSQL) PullEventIdsWithEventNameId(projectId int64, startTimestamp int64, endTimestamp int64, eventNameId string) ([]string, map[string]model.EventIdToProperties, error) {
	db := C.GetServices().Db

	events := make(map[string]model.EventIdToProperties, 0)
	eventsIds := make([]string, 0)

	rows, _ := db.Raw("SELECT events.id, events.user_id , events.timestamp, events.properties FROM events "+
		"WHERE events.project_id = ? AND events.event_name_id = ? AND events.timestamp >= ? AND events.timestamp <= ?"+
		" ORDER BY events.timestamp, events.created_at ASC", projectId, eventNameId, startTimestamp, endTimestamp).Rows()

	rowNum := 0
	for rows.Next() {
		var id string
		var userId string
		var timestamp int64
		var properties *postgres.Jsonb

		if err := rows.Scan(&id, &userId, &timestamp, &properties); err != nil {
			log.WithError(err).Error("Failed to scan rows")
			return nil, nil, err
		}

		var eventPropertiesBytes interface{}
		var err error
		if properties != nil {
			eventPropertiesBytes, err = properties.Value()
			if err != nil {
				log.WithError(err).Error("Failed to read event properties")
				return nil, nil, err
			}
		} else {
			eventPropertiesBytes = []byte(model.EmptyJsonStr)
		}

		var eventPropertiesMap map[string]interface{}
		err = json.Unmarshal(eventPropertiesBytes.([]byte), &eventPropertiesMap)
		if err != nil {
			log.WithError(err).Error("Failed to marshal event properties")
			return nil, nil, err
		}

		eventsIds = append(eventsIds, id)

		events[id] = model.EventIdToProperties{
			ID:              id,
			UserId:          userId,
			ProjectId:       projectId,
			EventProperties: eventPropertiesMap,
			Timestamp:       timestamp,
		}

		rowNum++
	}

	return eventsIds, events, nil
}

func (store *MemSQL) GetLatestEventTimeStampByEventNameId(projectId int64, eventNameId string,
	startTimestamp int64, endTimestamp int64) (int64, int) {

	if endTimestamp == 0 {
		return 0, http.StatusInternalServerError
	}

	db := C.GetServices().Db

	var timestamp int64

	rows, _ := db.Raw("SELECT MAX(timestamp ) FROM events "+
		"WHERE events.project_id = ? AND events.event_name_id = ? "+
		"AND events.timestamp >= ? AND events.timestamp <= ?", projectId, eventNameId, startTimestamp, endTimestamp).Rows()

	rowNum := 0
	for rows.Next() {

		if err := rows.Scan(&timestamp); err != nil {
			log.WithError(err).Error("Failed to scan rows")
			return 0, http.StatusNotFound
		}

		rowNum++
	}

	return timestamp, http.StatusFound
}
