package memsql

import (
	"database/sql"
	"errors"
	cacheRedis "factors/cache/redis"
	"factors/config"
	C "factors/config"
	"factors/metrics"
	"factors/model/model"
	"factors/util"
	U "factors/util"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/lib/pq"

	log "github.com/sirupsen/logrus"
)

const error_Duplicate_event_customerEventID = "pq: duplicate key value violates unique constraint \"project_id_customer_event_id_unique_idx\""
const eventsLimitForProperites = 50000
const OneDayInSeconds int64 = 24 * 60 * 60

func isDuplicateCustomerEventIdError(err error) bool {
	return err.Error() == error_Duplicate_event_customerEventID
}

func (store *MemSQL) GetEventCountOfUserByEventName(projectId uint64, userId string, eventNameId uint64) (uint64, int) {
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
func (store *MemSQL) GetEventCountOfUsersByEventName(projectID uint64, userIDs []string, eventNameID uint64) (uint64, int) {
	var count uint64

	db := C.GetServices().Db
	if err := db.Model(&model.Event{}).Where("project_id = ? AND user_id = ANY(?) AND event_name_id = ?",
		projectID, pq.Array(userIDs), eventNameID).Count(&count).Error; err != nil {
		log.WithFields(log.Fields{"projectId": projectID, "userId": userIDs}).WithError(err).Error(
			"Failed to get count of event for users by event_name_id")
		return 0, http.StatusInternalServerError
	}

	return count, http.StatusFound
}

func (store *MemSQL) addEventDetailsToCache(projectID uint64, event *model.Event, isUpdateEventProperty bool) {
	// TODO: Remove this check after enabling caching realtime.
	blackListedForUpdate := make(map[string]bool)
	blackListedForUpdate[U.EP_PAGE_SPENT_TIME] = true
	blackListedForUpdate[U.EP_PAGE_SCROLL_PERCENT] = true

	eventsToIncrSortedSet := make([]cacheRedis.SortedSetKeyValueTuple, 0)
	propertiesToIncrSortedSet := make([]cacheRedis.SortedSetKeyValueTuple, 0)
	valuesToIncrSortedSet := make([]cacheRedis.SortedSetKeyValueTuple, 0)
	logCtx := log.WithField("project_id", projectID)

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

	currentTime := U.TimeNow()
	currentTimeDatePart := currentTime.Format(U.DATETIME_FORMAT_YYYYMMDD)

	var eventNamesKeySortedSet *cacheRedis.Key
	if IsEventNameTypeSmartEvent(eventNameDetails.Type) {
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
	begin := U.TimeNow()
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
	end := U.TimeNow()
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

	// Use current properties of user, if user_properties is not provided.
	if event.UserProperties == nil {
		if C.IsOnTableUserPropertiesWriteAllowed(event.ProjectId) {
			properties, errCode := store.GetUserPropertiesByUserID(event.ProjectId, event.UserId)
			if errCode != http.StatusFound {
				logCtx.WithField("err_code", errCode).Error("Failed to get properties of user for event creation.")
			}
			event.UserProperties = properties
		}
	}

	if !C.IsOnTableUserPropertiesWriteAllowed(event.ProjectId) {
		// Reset user_properties, if not allowed on config,
		// for controlled rollout.
		event.UserProperties = nil
	}

	// Incrementing count based on EventNameId, not by EventName.
	count, errCode := store.GetEventCountOfUserByEventName(event.ProjectId, event.UserId, event.EventNameId)
	if errCode == http.StatusInternalServerError {
		return nil, errCode
	}
	event.Count = count + 1

	eventPropsJSONb, err := U.FillHourDayAndTimestampEventProperty(&event.Properties, event.Timestamp)
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
	if event.UserPropertiesId != "" {
		columnsInOrder = columnsInOrder + "," + "user_properties_id"
		paramsInOrder = append(paramsInOrder, event.UserPropertiesId)
	}
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

	db := C.GetServices().Db
	statement := fmt.Sprintf("INSERT INTO events (%s) VALUES (%s) RETURNING events.id", columnsInOrder, columnsPlaceholder)
	rows, err := db.Raw(statement, paramsInOrder...).Rows()
	if err != nil {
		if U.IsPostgresIntegrityViolationError(err) {
			logCtx.WithError(err).Info("CreateEvent Failed. Constraint violation.")
			return nil, http.StatusNotAcceptable
		}

		logCtx.WithField("event", event).WithError(err).Error("CreateEvent Failed")
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

	model.SetCacheUserLastEvent(event.ProjectId, event.UserId,
		&model.CacheEvent{ID: event.ID, Timestamp: event.Timestamp})
	return event, http.StatusCreated
}

func (store *MemSQL) GetEvent(projectId uint64, userId string, id string) (*model.Event, int) {
	db := C.GetServices().Db

	var event model.Event
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

func (store *MemSQL) GetEventById(projectId uint64, id string) (*model.Event, int) {
	db := C.GetServices().Db

	var event model.Event
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

func (store *MemSQL) GetLatestEventOfUserByEventNameId(projectId uint64, userId string, eventNameId uint64,
	startTimestamp int64, endTimestamp int64) (*model.Event, int) {

	db := C.GetServices().Db

	if startTimestamp == 0 || endTimestamp == 0 {
		return nil, http.StatusBadRequest
	}

	var events []model.Event
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

// GetRecentEventPropertyKeysWithLimits This method gets all the recent 'limit' property keys from DB for a given project/event
func (store *MemSQL) GetRecentEventPropertyKeysWithLimits(projectID uint64, eventName string, starttime int64, endtime int64, eventsLimit int) ([]U.Property, error) {

	logCtx := log.WithFields(log.Fields{"project_id": projectID, "eventName": eventName, "starttime": starttime, "endtime": endtime, "eventsLimit": eventsLimit})
	properties := make([]U.Property, 0)

	queryStr := "SELECT distinct(json_object_keys(properties::json)) AS key, " +
		" " + "COUNT(*) AS count,  " +
		" " + "MAX(timestamp) as last_seen" +
		" " + "FROM events  " +
		" " + "WHERE project_id = ? AND event_name_id IN ( " +
		" " + "	SELECT id FROM event_names WHERE project_id = ? AND name = ? " +
		" " + ") " +
		" " + "AND timestamp > ? AND timestamp <= ? AND properties != 'null' " +
		" " + "GROUP BY key ORDER BY count DESC LIMIT ?"

	db := C.GetServices().Db
	rows, err := db.Raw(queryStr, projectID, projectID, eventName, starttime, endtime, eventsLimit).Rows()
	if err != nil {
		logCtx.WithError(err).Error("Failed to get event properties.")
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var property U.Property
		if err := db.ScanRows(rows, &property); err != nil {
			logCtx.WithError(err).Error("Failed scanning rows on GetRecentEventPropertyKeysWithLimits")
			return properties, err
		}
		properties = append(properties, property)
	}
	return properties, nil
}

// GetRecentEventPropertyValuesWithLimits This method gets all the recent 'limit' property values from DB for a given project/event/property
func (store *MemSQL) GetRecentEventPropertyValuesWithLimits(projectID uint64, eventName string,
	property string, valuesLimit int, rowsLimit int, starttime int64, endtime int64) ([]U.PropertyValue, string, error) {
	db := C.GetServices().Db
	logCtx := log.WithFields(log.Fields{"projectId": projectID, "eventName": eventName, "property": property,
		"valuesLimit": valuesLimit, "rowsLimit": rowsLimit, "starttime": starttime, "endtime": endtime})

	values := make([]U.PropertyValue, 0)
	queryStr := "SELECT value, COUNT(*) AS count, MAX(timestamp) AS last_seen, json_typeof(value::json) AS value_type FROM" +
		" " + "(SELECT properties->? AS value, timestamp FROM events WHERE project_id = ? AND event_name_id IN" +
		" " + "(SELECT id FROM event_names WHERE project_id = ? AND name = ?) AND timestamp > ? AND timestamp <= ? AND properties->? IS NOT NULL LIMIT ?)" +
		" " + "AS property_values GROUP BY value ORDER BY count DESC LIMIT ?"

	rows, err := db.Raw(queryStr, property, projectID, projectID, eventName,
		starttime, endtime, property, rowsLimit, valuesLimit).Rows()
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

func (store *MemSQL) UpdateEventProperties(projectId uint64, id string,
	properties *U.PropertiesMap, updateTimestamp int64) int {

	if projectId == 0 || id == "" {
		return http.StatusBadRequest
	}

	event, errCode := store.GetEventById(projectId, id)
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

	db := C.GetServices().Db
	updatedFields := map[string]interface{}{
		"properties":                   updatedPostgresJsonb,
		"properties_updated_timestamp": propertiesLastUpdatedAt,
	}
	err = db.Model(&model.Event{}).Where("project_id = ? AND id = ?", projectId, id).Update(updatedFields).Error
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
	return http.StatusAccepted
}

func (store *MemSQL) GetUserEventsByEventNameId(projectId uint64, userId string, eventNameId uint64) ([]model.Event, int) {
	if projectId == 0 {
		return nil, http.StatusBadRequest
	}

	db := C.GetServices().Db

	var events []model.Event
	if err := db.Order("timestamp DESC").Where("project_id = ? AND user_id = ? AND event_name_id = ?",
		projectId, userId, eventNameId).Find(&events).Error; err != nil {

		return events, http.StatusInternalServerError
	}

	if len(events) == 0 {
		return events, http.StatusNotFound
	}

	return events, http.StatusFound
}

func getPageCountAndTimeSpentFromEventsList(events []*model.Event, sessionEvent *model.Event) (float64, float64) {
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

func getPageCountAndTimeSpentForContinuedSession(projectId uint64, userId string,
	continuedSessionEvent *model.Event, events []*model.Event) (float64, float64, float64, float64, int) {

	logCtx := log.WithFields(log.Fields{"project_id": projectId, "user_id": userId})

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

func (store *MemSQL) OverwriteEventProperties(projectId uint64, userId string, eventId string,
	newEventProperties *postgres.Jsonb) int {

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

	return http.StatusAccepted
}

func (store *MemSQL) OverwriteEventPropertiesByID(projectId uint64, id string,
	newEventProperties *postgres.Jsonb) int {
	logCtx := log.WithFields(log.Fields{"projectId": projectId, "id": id})

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

	return http.StatusAccepted
}

func doesEventIsPageViewAndHasMarketingProperty(event *model.Event) (bool, error) {
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
	err := db.Model(&model.Event{}).Where("project_id = ? AND id IN (?)",
		projectId, eventIds).Update(updateFields).Error
	if err != nil {
		logCtx.WithError(err).Error("Failed to associate session to events.")
		return http.StatusInternalServerError
	}

	return http.StatusAccepted
}

func associateSessionToEventsInBatch(projectId uint64, events []*model.Event,
	sessionId string, batchSize int) int {

	eventIds := make([]string, 0, len(events))
	for i := range events {
		event := *events[i]
		eventIds = append(eventIds, event.ID)
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

// AddSessionForUser - Wrapper for addSessionForUser to handle creating
// new session for last event when new session conditions met.
func (store *MemSQL) AddSessionForUser(projectId uint64, userId string, userEvents []model.Event,
	bufferTimeBeforeSessionCreateInSecs int64,
	sessionEventNameId uint64) (int, int, bool, int, int) {

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
func (pg *MemSQL) addSessionForUser(projectId uint64, userId string, userEvents []model.Event,
	bufferTimeBeforeSessionCreateInSecs int64,
	sessionEventNameId uint64) (int, int, bool, int, bool, int) {

	logCtx := log.WithFields(log.Fields{"project_id": projectId, "user_id": userId})

	if len(userEvents) == 0 {
		return 0, 0, false, 0, false, http.StatusNotModified
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

	noOfFilteredEvents := len(events)

	sessionStartIndex := 0
	sessionEndIndex := 0

	noOfSessionsCreated := 0
	sessionContinuedFlag := false
	isLastEventToBeProcessed := false

	// user_properties_id would be the key till the user_properties table
	// is permanantly deprecated and all the event level user_properties
	// moved to event itself.
	// map[id or user_properties_id of events] = session_user_properties
	sessionUserPropertiesRecordMap := make(map[string]model.SessionUserProperties, 0)

	// Use 2 moving cursor current, next. if diff(current, previous) > in-activity
	// period or has marketing property, use current_event - 1 as session end
	// and update. Update current_event as session start and do the same till the end.
	for i := 0; i < len(events); {
		hasMarketingProperty, err := doesEventIsPageViewAndHasMarketingProperty(events[i])
		if err != nil {
			logCtx.WithError(err).
				Error("Failed to check marketing prperty on event properties.")
			return noOfFilteredEvents, noOfSessionsCreated, sessionContinuedFlag, 0,
				isLastEventToBeProcessed, http.StatusInternalServerError
		}

		isNewSessionRequired := (i == 0 && len(events) == 1) ||
			(i > 0 && ((events[i].Timestamp - events[i-1].Timestamp) > model.NewUserSessionInactivityInSeconds))
		// Balance events on the list after creating session for the previous set.
		isLastSetOfEvents := i == len(events)-1

		isStartingWthMarketingProperty := i == 0 && len(events) > 1 && hasMarketingProperty
		if isStartingWthMarketingProperty {
			i++
			continue
		}

		if hasMarketingProperty || isNewSessionRequired || isLastSetOfEvents {
			var sessionEvent *model.Event
			var isSessionContinued bool

			if i > 0 {
				sessionEndIndex = i - 1
			}

			// Skip the associating previous session to last event, If it satisfies
			// new session condition. Instead of manipulating cursor, setting the
			// isLastEventToBeProcessed as true, to process it seperately.
			if isLastSetOfEvents {
				if !(hasMarketingProperty || isNewSessionRequired) {
					sessionEndIndex = i
				} else {
					if len(events) > 1 {
						isLastEventToBeProcessed = true
					}
				}
			}

			// Continue with the last session_id, if available. This will be true as
			// first event will have max_timestamp (used as start_timestamp) where
			// session_id is not null.
			var existingSessionEvent *model.Event
			if events[sessionStartIndex].SessionId != nil {
				var errCode int
				existingSessionEvent, errCode = pg.GetEventById(projectId, *events[sessionStartIndex].SessionId)
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

				logCtx = logCtx.WithField("event_id", firstEvent.ID)

				var userPropertiesMap U.PropertiesMap
				if C.ShouldUseUserPropertiesTableForRead(projectId) {
					// TODO(Dinesh): Remove the code block after permenant
					// deprecation of user_properties table.
					if firstEvent.UserPropertiesId != "" {
						userProperties, errCode := pg.GetUserProperties(projectId, userId, firstEvent.UserPropertiesId)
						if errCode != http.StatusFound {
							logCtx.WithField("err_code", errCode).
								WithField("user_properties_id", firstEvent.UserPropertiesId).
								Error("Failed to get user properties of first event on session.")
							return noOfFilteredEvents, noOfSessionsCreated, sessionContinuedFlag, 0,
								isLastEventToBeProcessed, http.StatusInternalServerError
						}

						userPropertiesDecoded, err := U.DecodePostgresJsonb(userProperties)
						if err != nil {
							logCtx.WithField("user_properties_id", firstEvent.UserPropertiesId).
								Error("Failed to decode user properties of first event on session.")
							return noOfFilteredEvents, noOfSessionsCreated, sessionContinuedFlag, 0,
								isLastEventToBeProcessed, http.StatusInternalServerError
						}

						userPropertiesMap = U.PropertiesMap(*userPropertiesDecoded)
					} else {
						logCtx.Error("Empty first event user_properties_id.")
					}
				} else {
					isEmptyUserProperties := firstEvent.UserProperties == nil ||
						U.IsEmptyPostgresJsonb(firstEvent.UserProperties)
					if !isEmptyUserProperties {
						userPropertiesDecoded, err := U.DecodePostgresJsonb(firstEvent.UserProperties)
						if err != nil {
							logCtx.WithField("user_properties", firstEvent.UserProperties).
								Error("Failed to decode user properties of first event on session.")
							return noOfFilteredEvents, noOfSessionsCreated, sessionContinuedFlag, 0,
								isLastEventToBeProcessed, http.StatusInternalServerError
						}

						userPropertiesMap = U.PropertiesMap(*userPropertiesDecoded)
					} else {
						logCtx.Error("Empty first event user_properties.")
					}
				}

				firstEventPropertiesDecoded, err := U.DecodePostgresJsonb(&firstEvent.Properties)
				if err != nil {
					logCtx.Error("Failed to decode event properties of first event on session.")
					return noOfFilteredEvents, noOfSessionsCreated, sessionContinuedFlag, 0,
						isLastEventToBeProcessed, http.StatusInternalServerError
				}
				firstEventPropertiesMap := U.PropertiesMap(*firstEventPropertiesDecoded)

				sessionEventCount, errCode := pg.GetEventCountOfUserByEventName(projectId, userId, sessionEventNameId)
				if errCode == http.StatusInternalServerError {
					logCtx.Error("Failed to get session event count for user.")
					return noOfFilteredEvents, noOfSessionsCreated, sessionContinuedFlag,
						0, isLastEventToBeProcessed, errCode
				}
				isFirstSession := sessionEventCount == 0
				sessionPropertiesMap := U.GetSessionProperties(isFirstSession,
					&firstEventPropertiesMap, &userPropertiesMap)
				sessionPropertiesEncoded := map[string]interface{}(*sessionPropertiesMap)

				sessionPropertiesJsonb, err := U.EncodeToPostgresJsonb(&sessionPropertiesEncoded)
				if err != nil {
					logCtx.WithError(err).Error("Failed to encode session properties as postgres jsonb.")
					return noOfFilteredEvents, noOfSessionsCreated, sessionContinuedFlag, 0,
						isLastEventToBeProcessed, http.StatusInternalServerError
				}

				// session event properties, to be updated
				newSessionEvent, errCode := pg.CreateEvent(&model.Event{
					EventNameId: sessionEventNameId,
					// Timestamp - 1sec before the first event of session.
					Timestamp: firstEvent.Timestamp - 1,
					ProjectId: projectId,
					UserId:    userId,
					// TODO(Dinesh): Remove UserPropertiesId after user_properties
					// table is permanatly deprecated. The value will be empty
					// when it is deprecated using flag.
					UserPropertiesId: firstEvent.UserPropertiesId,
					UserProperties:   firstEvent.UserProperties,
					Properties:       *sessionPropertiesJsonb,
				})

				if errCode != http.StatusCreated {
					logCtx.Error("Failed to create session event.")
					return noOfFilteredEvents, noOfSessionsCreated, sessionContinuedFlag,
						0, isLastEventToBeProcessed, errCode
				}

				sessionEvent = newSessionEvent
				noOfSessionsCreated++
			}

			eventsOfSession := events[sessionStartIndex : sessionEndIndex+1]

			// Update the session_id to all events between start index and end index + 1.
			errCode := associateSessionToEventsInBatch(projectId, eventsOfSession, sessionEvent.ID, 100)
			if errCode == http.StatusInternalServerError {
				logCtx.Error("Failed to associate session to events.")
				return noOfFilteredEvents, noOfSessionsCreated, sessionContinuedFlag,
					0, isLastEventToBeProcessed, errCode
			}

			lastEventProperties, err := U.DecodePostgresJsonb(&events[sessionEndIndex].Properties)
			if err != nil {
				logCtx.Error("Failed to decode properties of last event of session.")
				return noOfFilteredEvents, noOfSessionsCreated, sessionContinuedFlag, 0,
					isLastEventToBeProcessed, http.StatusInternalServerError
			}

			sessionPropertiesMap := U.PropertiesMap{}
			if _, exists := (*lastEventProperties)[U.EP_PAGE_RAW_URL]; exists {
				sessionPropertiesMap[U.SP_LATEST_PAGE_RAW_URL] = (*lastEventProperties)[U.EP_PAGE_RAW_URL]
			}
			if _, exists := (*lastEventProperties)[U.EP_PAGE_URL]; exists {
				sessionPropertiesMap[U.SP_LATEST_PAGE_URL] = (*lastEventProperties)[U.EP_PAGE_URL]
			}

			// Using existing method to get count and page spent time.
			var sessionPageCount, sessionPageSpentTime, onlyThisSessionPageCount, onlyThisSessionPageSpentTime float64

			if isSessionContinued {
				// Using db query, since previous session continued, we don't have all the events of the session.
				sessionPageCount, sessionPageSpentTime, onlyThisSessionPageCount, onlyThisSessionPageSpentTime, errCode = getPageCountAndTimeSpentForContinuedSession(
					projectId, userId, sessionEvent, events[sessionStartIndex:sessionEndIndex+1])
				if errCode == http.StatusInternalServerError {
					logCtx.Error("Failed to get page count and spent time of session on add session.")
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

			// Update session event properties.
			errCode = pg.UpdateEventProperties(projectId, sessionEvent.ID,
				&sessionPropertiesMap, sessionEvent.Timestamp+1)
			if errCode == http.StatusInternalServerError {
				logCtx.Error("Failed updating session event properties on add session.")
				return noOfFilteredEvents, noOfSessionsCreated, sessionContinuedFlag,
					0, isLastEventToBeProcessed, errCode
			}

			// associate user_properties state using session of the event.
			for i := range eventsOfSession {
				var userPropertiesRefID string
				if C.ShouldUseUserPropertiesTableForRead(projectId) {
					userPropertiesRefID = eventsOfSession[i].UserPropertiesId
				} else {
					userPropertiesRefID = eventsOfSession[i].ID
				}

				sessionUserPropertiesRecordMap[userPropertiesRefID] = model.SessionUserProperties{
					UserID:                userId,
					SessionEventTimestamp: sessionEvent.Timestamp,

					SessionCount:         sessionEvent.Count,
					SessionPageCount:     onlyThisSessionPageCount,
					SessionPageSpentTime: onlyThisSessionPageSpentTime,

					EventUserProperties: eventsOfSession[i].UserProperties,
				}
			}

			sessionStartIndex = i
		}

		i++
	}

	// Todo: The property values being updated are not accurate. Fix it.
	// Issue - https://github.com/Slashbit-Technologies/factors/issues/445
	errCode := pg.UpdateUserPropertiesForSession(projectId, &sessionUserPropertiesRecordMap)
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

// GetDatesForNextEventsArchivalBatch Get dates for events since startTime, excluding today's date.
func (store *MemSQL) GetDatesForNextEventsArchivalBatch(projectID uint64, startTime int64) (map[string]int64, int) {
	db := C.GetServices().Db
	countByDates := make(map[string]int64)

	rows, err := db.Model(&model.Event{}).
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

func (store *MemSQL) GetNextSessionEventInfoFromDB(projectID uint64, withSession bool,
	sessionEventNameId uint64, maxLookbackTimestamp int64) (int64, int) {

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

func (store *MemSQL) GetLastSessionEventTimestamp(projectID uint64, sessionEventNameID uint64) (int64, int) {
	logCtx := log.WithField("project_id", projectID)

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
func (store *MemSQL) GetAllEventsForSessionCreationAsUserEventsMap(projectId, sessionEventNameId uint64,
	startTimestamp, endTimestamp int64) (*map[string][]model.Event, int, int) {

	logCtx := log.WithFields(log.Fields{"project_id": projectId,
		"start_timestamp": startTimestamp, "end_timestamp": endTimestamp})

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
	excludeSkipSessionCondition := fmt.Sprintf("(JSON_EXTRACT_STRING(properties, '%s') IS NULL OR JSON_EXTRACT_STRING(properties, '%s') = ?)",
		U.EP_SKIP_SESSION, U.EP_SKIP_SESSION)
	if err := db.Order("timestamp, created_at ASC").
		Where("project_id = ? AND event_name_id != ? AND timestamp BETWEEN ? AND ?"+" AND "+excludeSkipSessionCondition,
			projectId, sessionEventNameId, startTimestamp, endTimestamp, false).
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

// GetEventsWithoutPropertiesAndWithPropertiesByName - Use for getting properties with and without values
// and use it for updating the events which doesn't have the values. User for fixing data for YourStory.
func (store *MemSQL) GetEventsWithoutPropertiesAndWithPropertiesByNameForYourStory(projectID uint64, from,
	to int64, mandatoryProperties []string) ([]model.EventWithProperties, *map[string]U.PropertiesMap, int) {
	logCtx := log.WithField("project_id", projectID).
		WithField("from", from).
		WithField("to", to)

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
			if !exists && propAuthorName != nil {
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

func (store *MemSQL) GetUnusedSessionIDsForJob(projectID uint64, startTimestamp, endTimestamp int64) ([]string, int) {
	logCtx := log.WithField("project_id", projectID).
		WithField("start_timestamp", startTimestamp).
		WithField("end_timestamp", endTimestamp)

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
		logCtx.Error("Failed to get session event_name.")
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

func (store *MemSQL) DeleteEventsByIDsInBatchForJob(projectID, eventNameID uint64, ids []string, batchSize int) int {
	logCtx := log.WithField("project_id", projectID).WithField("batch_size", batchSize)
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

func (store *MemSQL) DeleteEventByIDs(projectID, eventNameID uint64, ids []string) int {
	logCtx := log.WithField("project_id", projectID)

	db := C.GetServices().Db
	exec := db.Where("project_id = ? AND id = ANY(?)", projectID, pq.Array(ids)).Delete(&model.Event{})
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

func (store *MemSQL) OverwriteEventUserPropertiesByID(projectID uint64,
	id string, userProperties *postgres.Jsonb) int {

	logCtx := log.WithField("project_id", projectID).WithField("id", id)

	if projectID == 0 || id == "" {
		logCtx.Error("Invalid values for arguments.")
		return http.StatusBadRequest
	}

	if userProperties == nil || U.IsEmptyPostgresJsonb(userProperties) {
		logCtx.Error("Failed to overwrite user_properties. Empty or nil properties.")
		return http.StatusBadRequest
	}

	// Not updating the event_user_properties
	db := C.GetServices().Db
	err := db.Model(&model.Event{}).Where("project_id = ? AND id = ?", projectID, id).
		Update("user_properties", userProperties).Error
	if err != nil {
		logCtx.WithError(err).Error("Failed to overwrite user properteis.")
		return http.StatusInternalServerError
	}

	return http.StatusAccepted
}

// PullEventRowsForBuildSequenceJob - Function to pull events for factors model building sequentially.
func (store *MemSQL) PullEventRowsForBuildSequenceJob(projectID uint64, startTime, endTime int64) (*sql.Rows, error) {
	rawQuery := fmt.Sprintf("SELECT COALESCE(users.customer_user_id, users.id), event_names.name, events.timestamp, events.count,"+
		" events.properties, users.join_timestamp, events.user_properties FROM events "+
		"LEFT JOIN event_names ON events.event_name_id = event_names.id LEFT JOIN users ON events.user_id = users.id "+
		"WHERE events.project_id = %d AND events.timestamp BETWEEN  %d AND %d "+
		"ORDER BY COALESCE(users.customer_user_id, users.id), events.timestamp LIMIT %d", projectID, startTime, endTime, model.EventsPullLimit+1)

	if config.ShouldUseUserPropertiesTableForRead(projectID) {
		rawQuery = fmt.Sprintf("SELECT COALESCE(users.customer_user_id, users.id), event_names.name, events.timestamp, events.count,"+
			" events.properties, users.join_timestamp, user_properties.properties FROM events "+
			"LEFT JOIN event_names ON events.event_name_id = event_names.id LEFT JOIN users ON events.user_id = users.id "+
			"LEFT JOIN user_properties ON events.user_properties_id = user_properties.id "+
			"WHERE events.project_id = %d AND events.timestamp BETWEEN  %d AND %d "+
			"ORDER BY COALESCE(users.customer_user_id, users.id), events.timestamp LIMIT %d", projectID, startTime, endTime, model.EventsPullLimit+1)
	}

	db := C.GetServices().Db
	return db.Raw(rawQuery).Rows()
}

// PullEventsForArchivalJob - Function to pull events for archival.
func (store *MemSQL) PullEventRowsForArchivalJob(projectID uint64, startTime, endTime int64) (*sql.Rows, error) {

	rawQuery := fmt.Sprintf("SELECT events.id, users.id, users.customer_user_id, "+
		"event_names.name, events.timestamp, events.session_id, events.properties, users.join_timestamp, events.user_properties FROM events "+
		"LEFT JOIN event_names ON events.event_name_id = event_names.id LEFT JOIN users ON events.user_id = users.id "+
		"WHERE events.project_id = %d AND events.timestamp BETWEEN %d AND %d", projectID, startTime, endTime)

	if config.ShouldUseUserPropertiesTableForRead(projectID) {
		rawQuery = fmt.Sprintf("SELECT events.id, users.id, users.customer_user_id, "+
			"event_names.name, events.timestamp, events.session_id, events.properties, users.join_timestamp, events.user_properties FROM events "+
			"LEFT JOIN event_names ON events.event_name_id = event_names.id LEFT JOIN users ON events.user_id = users.id "+
			"LEFT JOIN user_properties ON events.user_properties_id = user_properties.id "+
			"WHERE events.project_id = %d AND events.timestamp BETWEEN %d AND %d", projectID, startTime, endTime)
	}

	db := C.GetServices().Db
	return db.Raw(rawQuery).Rows()
}
