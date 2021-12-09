package postgres

import (
	"encoding/json"
	"errors"
	cacheRedis "factors/cache/redis"
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"fmt"
	"net/http"
	"strings"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

// TODO: Make index name a constant and read it
// error constants
const error_DUPLICATE_FILTER_EXPR = "pq: duplicate key value violates unique constraint \"project_filter_expr_unique_idx\""

func isDuplicateFilterExprError(err error) bool {
	return err.Error() == error_DUPLICATE_FILTER_EXPR
}

func (pg *Postgres) CreateOrGetEventName(eventName *model.EventName) (*model.EventName, int) {
	logCtx := log.WithFields(log.Fields{"event_name": &eventName})

	// Validation.
	if eventName.ProjectId == 0 || !isValidType(eventName.Type) ||
		!isValidName(eventName.Name, eventName.Type) {

		return nil, http.StatusBadRequest
	}

	eventName.Deleted = false

	db := C.GetServices().Db
	if err := db.FirstOrInit(&eventName, &eventName).Error; err != nil {
		logCtx.WithError(err).Error("Failed to create event_name.")
		return nil, http.StatusInternalServerError
	}

	// Checks new record or not.
	if !eventName.CreatedAt.IsZero() {
		return eventName, http.StatusConflict
	} else if err := db.Create(eventName).Error; err != nil {
		logCtx.WithError(err).Error("Failed to create event_name.")

		if isDuplicateFilterExprError(err) {
			return nil, http.StatusBadRequest
		}

		return nil, http.StatusInternalServerError
	}

	return eventName, http.StatusCreated
}

func (pg *Postgres) CreateOrGetUserCreatedEventName(eventName *model.EventName) (*model.EventName, int) {
	eventName.Type = model.TYPE_USER_CREATED_EVENT_NAME
	return pg.CreateOrGetEventName(eventName)
}

func (pg *Postgres) CreateOrGetAutoTrackedEventName(eventName *model.EventName) (*model.EventName, int) {
	eventName.Type = model.TYPE_AUTO_TRACKED_EVENT_NAME
	return pg.CreateOrGetEventName(eventName)
}

func (pg *Postgres) CreateOrGetFilterEventName(eventName *model.EventName) (*model.EventName, int) {
	filterExpr, valid := getValidatedFilterExpr(eventName.FilterExpr)
	if !valid {
		return nil, http.StatusBadRequest
	}

	eventName.Type = model.TYPE_FILTER_EVENT_NAME
	eventName.FilterExpr = filterExpr

	return pg.CreateOrGetEventName(eventName)
}

func (pg *Postgres) checkDuplicateSmartEventFilter(projectID uint64, inFilterExpr *model.SmartCRMEventFilter) (*model.EventName, bool) {
	eventNames, status := pg.GetSmartEventFilterEventNames(projectID, true)
	if status == http.StatusNotFound {
		return nil, false
	}

	for i := range eventNames {

		exFilterExpr, err := model.GetDecodedSmartEventFilterExp(eventNames[i].FilterExpr)
		if err != nil {
			log.WithError(err).Error("Failed to GetDecodedSmartEventFilterExp")
			continue
		}

		duplicate := model.CheckSmartEventNameDuplicateFilter(exFilterExpr, inFilterExpr)
		if duplicate {
			return &eventNames[i], true
		}
	}

	return nil, false
}

// CreateOrGetCRMSmartEventFilterEventName creates a new CRM smart event filter.
// Deleted event_name will be enabled if conflict found
func (pg *Postgres) CreateOrGetCRMSmartEventFilterEventName(projectID uint64, eventName *model.EventName,
	filterExpr *model.SmartCRMEventFilter) (*model.EventName, int) {

	logCtx := log.WithFields(log.Fields{"project_id": projectID, "event_name": *eventName, "filter_exp": *filterExpr})
	if !model.IsValidSmartEventFilterExpr(filterExpr) || filterExpr == nil || eventName.Type != "" ||
		eventName.Name == "" {
		logCtx.Error("Invalid fields.")
		return nil, http.StatusBadRequest
	}

	dupEventName, duplicate := pg.checkDuplicateSmartEventFilter(projectID, filterExpr)
	if duplicate { // re-enable the smart event name
		if dupEventName.Deleted == true {
			updateEventName := &model.EventName{
				Name:    eventName.Name, // use new event name provided by api
				Type:    getCRMSmartEventNameType(filterExpr.Source),
				Deleted: true,
			}

			filterExpr, err := model.GetDecodedSmartEventFilterExp(dupEventName.FilterExpr)
			if err != nil {
				logCtx.WithError(err).Error("Failed to GetDecodedSmartEventFilterExp.")
				return nil, http.StatusInternalServerError
			}

			updatedEventName, status := pg.updateCRMSmartEventFilter(projectID, dupEventName.ID, updateEventName.Type, updateEventName, filterExpr)
			if status != http.StatusAccepted {
				logCtx.Error("Failed to update deleted smart event filter.")
				return nil, http.StatusInternalServerError
			}

			return updatedEventName, status
		}

		return nil, http.StatusConflict
	}

	enFilterExp, err := json.Marshal(filterExpr)
	if err != nil {
		logCtx.WithError(err).Error("Failed to marshal filterExpr on CreateOrGetCRMSmartEventFilterEventName")
		return nil, http.StatusInternalServerError
	}

	eventName.FilterExpr = string(enFilterExp)

	eventName.Type = getCRMSmartEventNameType(filterExpr.Source)

	_, status := pg.CreateOrGetEventName(eventName)
	if status != http.StatusCreated {
		logCtx.Error("Failed to CreateOrGetCRMSmartEventFilterEventName.")
		return nil, http.StatusInternalServerError
	}

	return eventName, status
}

func (pg *Postgres) GetSmartEventEventName(eventName *model.EventName) (*model.EventName, int) {
	return pg.GetSmartEventEventNameByNameANDType(eventName.ProjectId, eventName.Name, eventName.Type)
}

func (pg *Postgres) GetSmartEventEventNameByNameANDType(projectID uint64, name, typ string) (*model.EventName, int) {
	if projectID == 0 || name == "" || typ == "" {
		return nil, http.StatusBadRequest
	}

	db := C.GetServices().Db

	var eventNames []model.EventName
	if err := db.Where("project_id = ? AND type = ? AND name = ? and deleted = 'false'",
		projectID, typ, name).Find(&eventNames).Error; err != nil {
		log.WithFields(log.Fields{"project_id": projectID}).WithError(err).Error("Failed getting filter_event_names")

		return nil, http.StatusInternalServerError
	}

	if len(eventNames) != 1 {
		return nil, http.StatusInternalServerError
	}

	return &eventNames[0], http.StatusFound
}

func (pg *Postgres) CreateOrGetSessionEventName(projectId uint64) (*model.EventName, int) {
	return pg.CreateOrGetEventName(&model.EventName{ProjectId: projectId, Name: U.EVENT_NAME_SESSION,
		Type: model.TYPE_INTERNAL_EVENT_NAME})
}

func (pg *Postgres) CreateOrGetOfflineTouchPointEventName(projectId uint64) (*model.EventName, int) {
	return pg.CreateOrGetEventName(&model.EventName{ProjectId: projectId, Name: U.EVENT_NAME_OFFLINE_TOUCH_POINT,
		Type: model.TYPE_INTERNAL_EVENT_NAME})
}

func (pg *Postgres) GetSessionEventName(projectId uint64) (*model.EventName, int) {
	if projectId == 0 {
		return nil, http.StatusBadRequest
	}

	logCtx := log.WithField("project_id", projectId)

	var eventNames []model.EventName

	db := C.GetServices().Db
	err := db.Limit(1).Where("project_id = ?", projectId).
		Where("name = ?", U.EVENT_NAME_SESSION).
		Find(&eventNames).Error
	if err != nil {
		logCtx.WithError(err).Error("Falied to get session event name.")
		return nil, http.StatusInternalServerError
	}

	if len(eventNames) == 0 {
		return nil, http.StatusNotFound
	}

	return &eventNames[0], http.StatusFound
}

func isValidType(nameType string) bool {
	if nameType == "" {
		return false
	}

	for _, allowedType := range model.ALLOWED_TYPES {
		if nameType == allowedType {
			return true
		}
	}
	return false
}

func isValidName(name string, typ string) bool {
	if name == "" {
		return false
	}

	if typ == model.TYPE_INTERNAL_EVENT_NAME {
		return true
	}

	for _, allowedEventName := range U.ALLOWED_INTERNAL_EVENT_NAMES {
		if name == allowedEventName {
			return true
		}
	}

	return !strings.HasPrefix(name, U.NAME_PREFIX)
}

func (pg *Postgres) GetEventName(name string, projectId uint64) (*model.EventName, int) {
	// Input Validation. (ID is to be auto generated)
	if name == "" || name == "null" || projectId == 0 {
		log.Error("GetEventName Failed. Missing name or projectId")
		return nil, http.StatusBadRequest
	}

	db := C.GetServices().Db

	var eventName model.EventName
	if err := db.Where(&model.EventName{Name: name, ProjectId: projectId}).First(&eventName).Error; err != nil {
		log.WithFields(log.Fields{"projectId": projectId, "Name": name}).WithError(err).Error(
			"Getting eventName failed on GetEventName")
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}
		return nil, http.StatusInternalServerError
	}
	return &eventName, http.StatusFound
}

func (pg *Postgres) GetEventNameSQL(name string, projectID uint64) (string, []interface{}) {
	var params []interface{}
	params = append(params, projectID)
	params = append(params, name)
	return "select id from event_names where project_id = ? AND name = ?", params
}

func (pg *Postgres) GetEventNames(projectId uint64) ([]model.EventName, int) {
	if projectId == 0 {
		log.Error("GetEventNames Failed. Missing projectId")
		return nil, http.StatusBadRequest
	}

	db := C.GetServices().Db

	var eventNames []model.EventName
	if err := db.Order("created_at ASC").Where("project_id = ?", projectId).Limit(2000).Find(&eventNames).Error; err != nil {
		return nil, http.StatusInternalServerError
	}
	if len(eventNames) == 0 {
		return eventNames, http.StatusNotFound
	}
	return eventNames, http.StatusFound
}

// GetOrderedEventNamesFromDb - Get 'limit' events from DB sort by occurence for a given time period
func (pg *Postgres) GetOrderedEventNamesFromDb(
	projectID uint64, startTimestamp int64, endTimestamp int64, limit int) ([]model.EventNameWithAggregation, error) {
	db := C.GetServices().Db
	hasLimit := limit > 0
	eventNames := make([]model.EventNameWithAggregation, 0)

	logCtx := log.WithFields(log.Fields{"projectId": projectID,
		"startTimestamp": startTimestamp, "endTimestamp": endTimestamp})

	// Gets occurrence count of event from events table for a
	// limited time window and upto 100k and order by count
	// then join with event names.
	queryStr := "SELECT * FROM (SELECT event_name_id, COUNT(*) as count, MAX(timestamp) AS last_seen  FROM" +
		" " + "(SELECT event_name_id, timestamp FROM events WHERE project_id=? AND timestamp > ?" +
		" " + "AND timestamp <= ? ORDER BY timestamp DESC LIMIT ?) AS sample_events" +
		" " + "GROUP BY event_name_id ORDER BY count DESC) AS event_occurrence" +
		" " + "LEFT JOIN event_names ON event_occurrence.event_name_id=event_names.id "

	if hasLimit {
		queryStr = queryStr + " " + "LIMIT ?"
	}

	const noOfEventsToLoadForOccurrenceSort = 100000

	params := make([]interface{}, 0)
	params = append(params, projectID, startTimestamp,
		endTimestamp, noOfEventsToLoadForOccurrenceSort)
	if hasLimit {
		params = append(params, limit)
	}

	rows, err := db.Raw(queryStr, params...).Rows()
	if err != nil {
		logCtx.WithError(err).Error(
			"Failed to execute query of get event names ordered by occurrence.")
		return eventNames, err
	}

	for rows.Next() {
		var eventName model.EventNameWithAggregation
		if err := db.ScanRows(rows, &eventName); err != nil {
			logCtx.WithError(err).Error("Failed scanning rows on get event names ordered by occurrence.")
			return eventNames, err
		}
		eventNames = append(eventNames, eventName)
	}
	return eventNames, nil
}

// GetPropertyValuesByEventProperty (Part of event_name and properties caching) This method iterates for
// last n days to get all the top 'limit' property values for the given property/event
// Picks all last 24 hours values and sorts the remaining by occurence and returns top 'limit' values
func (pg *Postgres) GetPropertyValuesByEventProperty(projectID uint64, eventName string,
	propertyName string, limit int, lastNDays int) ([]string, error) {

	if projectID == 0 {
		return []string{}, errors.New("invalid project on GetPropertyValuesByEventProperty")
	}

	if eventName == "" {
		return []string{}, errors.New("invalid event_name on GetPropertyValuesByEventProperty")
	}

	if propertyName == "" {
		return []string{}, errors.New("invalid property_name on GetPropertyValuesByEventProperty")
	}
	currentDate := model.OverrideCacheDateRangeForProjects(projectID)
	values := make([]U.CachePropertyValueWithTimestamp, 0)
	for i := 0; i < lastNDays; i++ {
		currentDateOnlyFormat := currentDate.AddDate(0, 0, -i).Format(U.DATETIME_FORMAT_YYYYMMDD)
		value, err := getPropertyValuesByEventPropertyFromCache(projectID, eventName, propertyName, currentDateOnlyFormat)
		if err != nil {
			return []string{}, err
		}
		values = append(values, value)
	}

	valueStrings := make([]string, 0)
	valuesAggregated := U.AggregatePropertyValuesAcrossDate(values)
	valuesSorted := U.SortByTimestampAndCount(valuesAggregated)

	for _, v := range valuesSorted {
		valueStrings = append(valueStrings, v.Name)
	}
	if limit > 0 {
		sliceLength := len(valueStrings)
		if sliceLength > limit {
			return valueStrings[0:limit], nil
		}
	}
	return valueStrings, nil
}

func getPropertyValuesByEventPropertyFromCache(projectID uint64, eventName string, propertyName string, dateKey string) (U.CachePropertyValueWithTimestamp, error) {
	logCtx := log.WithFields(log.Fields{
		"project_id": projectID,
	})
	if projectID == 0 {
		return U.CachePropertyValueWithTimestamp{}, errors.New("invalid project on GetPropertyValuesByEventPropertyFromCache")
	}

	if eventName == "" {
		return U.CachePropertyValueWithTimestamp{}, errors.New("invalid event_name on GetPropertyValuesByEventPropertyFromCache")
	}

	if propertyName == "" {
		return U.CachePropertyValueWithTimestamp{}, errors.New("invalid property_name on GetPropertyValuesByEventPropertyFromCache")
	}

	eventPropertyValuesKey, err := model.GetValuesByEventPropertyRollUpCacheKey(projectID, eventName, propertyName, dateKey)
	if err != nil {
		return U.CachePropertyValueWithTimestamp{}, err
	}
	values, _, err := cacheRedis.GetIfExistsPersistent(eventPropertyValuesKey)
	if values == "" {
		eventNameWithSlash := fmt.Sprintf("%s/", eventName)
		eventPropertyValuesKeyWithSlash, err := model.GetValuesByEventPropertyRollUpCacheKey(projectID, eventNameWithSlash, propertyName, dateKey)
		if err != nil {
			return U.CachePropertyValueWithTimestamp{}, err
		}
		valuesWithSlash, _, err := cacheRedis.GetIfExistsPersistent(eventPropertyValuesKeyWithSlash)
		if valuesWithSlash == "" {
			logCtx.WithField("date_key", dateKey).Info("MISSING ROLLUP EPV")
			return U.CachePropertyValueWithTimestamp{}, nil
		}
		var cacheValueWithSlash U.CachePropertyValueWithTimestamp
		err = json.Unmarshal([]byte(valuesWithSlash), &cacheValueWithSlash)
		if err != nil {
			return U.CachePropertyValueWithTimestamp{}, err
		}
		return cacheValueWithSlash, nil
	}
	var cacheValue U.CachePropertyValueWithTimestamp
	err = json.Unmarshal([]byte(values), &cacheValue)
	if err != nil {
		return U.CachePropertyValueWithTimestamp{}, err
	}
	return cacheValue, nil
}

// GetPropertiesByEvent (Part of event_name and properties caching) This method iterates for last n days to get all the
// top 'limit' properties for the given event. Picks all last 24 hours properties and sorts the remaining by occurence
// and returns top 'limit' properties
func (pg *Postgres) GetPropertiesByEvent(projectID uint64, eventName string, limit int, lastNDays int) (map[string][]string, error) {
	properties := make(map[string][]string)
	if projectID == 0 {
		return properties, errors.New("invalid project on GetPropertiesByEvent")
	}
	currentDate := model.OverrideCacheDateRangeForProjects(projectID)
	if eventName == "" {
		return properties, errors.New("invalid event_name on GetPropertiesByEvent")
	}
	eventProperties := make([]U.CachePropertyWithTimestamp, 0)
	for i := 0; i < lastNDays; i++ {
		currentDateOnlyFormat := currentDate.AddDate(0, 0, -i).Format(U.DATETIME_FORMAT_YYYYMMDD)
		eventProperty, err := getPropertiesByEventFromCache(projectID, eventName, currentDateOnlyFormat)
		if err != nil {
			return nil, err
		}
		eventProperties = append(eventProperties, eventProperty)
	}

	eventPropertiesAggregated := U.AggregatePropertyAcrossDate(eventProperties)
	eventPropertiesSorted := U.SortByTimestampAndCount(eventPropertiesAggregated)

	if limit > 0 {
		sliceLength := len(eventPropertiesSorted)
		if sliceLength > limit {
			eventPropertiesSorted = eventPropertiesSorted[0:limit]
		}
	}

	propertyDetails, propertyDetailsStatus := pg.GetAllPropertyDetailsByProjectID(projectID, eventName, false)
	for _, v := range eventPropertiesSorted {
		category := v.Category
		if propertyDetailsStatus == http.StatusFound {
			pName := model.GetPropertyNameByTrimmedSmartEventPropertyPrefix(v.Name)
			if pType, exist := (*propertyDetails)[pName]; exist {
				category = pType
			}
		}

		if properties[category] == nil {
			properties[category] = make([]string, 0)
		}
		properties[category] = append(properties[category], v.Name)
	}

	return properties, nil
}

func getPropertiesByEventFromCache(projectID uint64, eventName string, dateKey string) (U.CachePropertyWithTimestamp, error) {
	logCtx := log.WithFields(log.Fields{
		"project_id": projectID,
	})
	if projectID == 0 {
		return U.CachePropertyWithTimestamp{}, errors.New("invalid project on GetPropertiesByEventFromCache")
	}

	if eventName == "" {
		return U.CachePropertyWithTimestamp{}, errors.New("invalid event_name on GetPropertiesByEventFromCache")
	}

	eventPropertiesKey, err := model.GetPropertiesByEventCategoryRollUpCacheKey(projectID, eventName, dateKey)
	if err != nil {
		return U.CachePropertyWithTimestamp{}, err
	}
	eventProperties, _, err := cacheRedis.GetIfExistsPersistent(eventPropertiesKey)
	if eventProperties == "" {
		eventNameWithSlash := fmt.Sprintf("%s/", eventName)
		eventPropertiesKeyWithSlash, err := model.GetPropertiesByEventCategoryRollUpCacheKey(projectID, eventNameWithSlash, dateKey)
		if err != nil {
			return U.CachePropertyWithTimestamp{}, err
		}
		eventPropertiesWithSlash, _, err := cacheRedis.GetIfExistsPersistent(eventPropertiesKeyWithSlash)
		if eventPropertiesWithSlash == "" {
			logCtx.WithField("date_key", dateKey).Info("MISSING ROLLUP EP")
			return U.CachePropertyWithTimestamp{}, nil
		}
		var cacheValueWithSlash U.CachePropertyWithTimestamp
		err = json.Unmarshal([]byte(eventPropertiesWithSlash), &cacheValueWithSlash)
		if err != nil {
			return U.CachePropertyWithTimestamp{}, err
		}
		return cacheValueWithSlash, nil
	}
	var cacheValue U.CachePropertyWithTimestamp
	err = json.Unmarshal([]byte(eventProperties), &cacheValue)
	if err != nil {
		return U.CachePropertyWithTimestamp{}, err
	}
	return cacheValue, nil
}

func aggregateEventsAcrossDate(events []model.CacheEventNamesWithTimestamp) []U.NameCountTimestampCategory {
	eventsAggregated := make(map[string]U.CountTimestampTuple)
	// Sort Event Properties by timestamp, count and return top n
	for _, event := range events {
		for eventName, eventDetails := range event.EventNames {
			eventNameSuffixTrim := strings.TrimSuffix(eventName, "/")
			eventsAggregatedInt := eventsAggregated[eventNameSuffixTrim]
			eventsAggregatedInt.Count += eventDetails.Count
			if eventsAggregatedInt.LastSeenTimestamp < eventDetails.LastSeenTimestamp {
				eventsAggregatedInt.LastSeenTimestamp = eventDetails.LastSeenTimestamp
			}

			if eventDetails.Type == model.EVENT_NAME_TYPE_SMART_EVENT {
				eventsAggregatedInt.Type = U.SmartEvent
			}

			eventsAggregated[eventNameSuffixTrim] = eventsAggregatedInt
		}
	}
	eventsAggregatedSlice := make([]U.NameCountTimestampCategory, 0)
	for k, v := range eventsAggregated {
		eventsAggregatedSlice = append(eventsAggregatedSlice, U.NameCountTimestampCategory{
			k, v.Count, v.LastSeenTimestamp, v.Type, ""})
	}
	return eventsAggregatedSlice
}

// Hacked solution - to fetch a type of EventNames.
func (pg *Postgres) GetMostFrequentlyEventNamesByType(projectID uint64, limit int, lastNDays int, typeOfEvent string) ([]string, error) {
	mostFrequentEventNames := make([]string, 0)
	var eventNameType string
	var exists bool
	var finalEventNames []model.EventName
	var result []string
	db := C.GetServices().Db

	if eventNameType, exists = model.EventTypeToEnameType[typeOfEvent]; !exists {
		return nil, errors.New("invalid type is provided.")
	}
	eventsSorted, err := getEventNamesAggregatedAndSortedAcrossDate(projectID, limit, lastNDays)
	if err != nil {
		return nil, err
	}

	for _, event := range eventsSorted {
		if event.GroupName == U.MostRecent {
			mostFrequentEventNames = append(mostFrequentEventNames, event.Name)
		}
	}
	for _, event := range eventsSorted {
		if event.GroupName == U.FrequentlySeen {
			mostFrequentEventNames = append(mostFrequentEventNames, event.Name)
		}
	}
	if limit > 0 {
		sliceLength := len(mostFrequentEventNames)
		if sliceLength > limit*2 {
			mostFrequentEventNames = mostFrequentEventNames[0 : limit*2]
		}
	}
	if dbResult := db.Where("type = ? AND name IN (?)", eventNameType, mostFrequentEventNames).Select("name").Limit(limit).Find(&finalEventNames); dbResult.Error != nil {
		return nil, dbResult.Error
	}
	hashMapOfFinalEventNames := make(map[string]int)
	for _, eventName := range finalEventNames {
		hashMapOfFinalEventNames[eventName.Name] = 1
	}
	for _, event := range mostFrequentEventNames {
		if _, ok := hashMapOfFinalEventNames[event]; ok {
			result = append(result, event)
		}
	}
	return result, nil
}

// GetEventNamesOrderedByOccurenceAndRecency (Part of event_name and properties caching) This method iterates for last n days to
// get all the top 'limit' events for the given project. Picks all last 24 hours events and sorts the remaining by
// occurence and returns top 'limit' events
func (pg *Postgres) GetEventNamesOrderedByOccurenceAndRecency(projectID uint64, limit int, lastNDays int) (map[string][]string, error) {
	eventsSorted, err := getEventNamesAggregatedAndSortedAcrossDate(projectID, limit, lastNDays)
	if err != nil {
		return nil, err
	}
	eventsFiltered := make([]U.NameCountTimestampCategory, 0)
	eventsFiltered = eventsSorted
	if limit > 0 {
		sliceLength := len(eventsSorted)
		if sliceLength > limit {
			eventsFiltered = eventsSorted[0:limit]
			for _, event := range eventsSorted[limit:sliceLength] {
				_, ok := U.STANDARD_EVENTS_GROUP_NAMES[event.Name]
				if ok {
					eventsFiltered = append(eventsFiltered, event)
				}
			}
		}
	}

	eventStringWithGroups := make(map[string][]string)
	for _, event := range eventsFiltered {
		eventStringWithGroups[event.GroupName] = append(eventStringWithGroups[event.GroupName], event.Name)
	}

	return eventStringWithGroups, nil
}

func getEventNamesAggregatedAndSortedAcrossDate(projectID uint64, limit int, lastNDays int) ([]U.NameCountTimestampCategory, error) {
	if projectID == 0 {
		return nil, errors.New("invalid project on get event names ordered by occurence and recency")
	}
	currentDate := model.OverrideCacheDateRangeForProjects(projectID)
	events := make([]model.CacheEventNamesWithTimestamp, 0)
	for i := 0; i < lastNDays; i++ {
		currentDateOnlyFormat := currentDate.AddDate(0, 0, -i).Format(U.DATETIME_FORMAT_YYYYMMDD)
		event, err := getEventNamesOrderedByOccurenceAndRecencyFromCache(projectID, currentDateOnlyFormat)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	eventsAggregated := aggregateEventsAcrossDate(events)

	eventsSorted := U.SortByTimestampAndCount(eventsAggregated)
	return eventsSorted, nil
}

func getEventNamesOrderedByOccurenceAndRecencyFromCache(projectID uint64, dateKey string) (model.CacheEventNamesWithTimestamp, error) {
	logCtx := log.WithFields(log.Fields{
		"project_id": projectID,
	})
	if projectID == 0 {
		return model.CacheEventNamesWithTimestamp{},
			errors.New("invalid project on get event names ordered by occurence and recency from cache")
	}

	eventNamesKey, err := model.GetEventNamesOrderByOccurrenceAndRecencyRollUpCacheKey(projectID, dateKey)
	if err != nil {
		return model.CacheEventNamesWithTimestamp{}, err
	}
	eventNames, _, err := cacheRedis.GetIfExistsPersistent(eventNamesKey)
	if eventNames == "" {
		logCtx.WithField("date_key", dateKey).Info("MISSING ROLLUP EN")
		return model.CacheEventNamesWithTimestamp{}, nil
	}
	var cacheEventNames model.CacheEventNamesWithTimestamp
	err = json.Unmarshal([]byte(eventNames), &cacheEventNames)
	if err != nil {
		return model.CacheEventNamesWithTimestamp{}, err
	}
	return cacheEventNames, nil
}

func (pg *Postgres) GetFilterEventNames(projectId uint64) ([]model.EventName, int) {
	db := C.GetServices().Db

	var eventNames []model.EventName
	if err := db.Where("project_id = ? AND type = ? AND deleted = 'false'",
		projectId, model.TYPE_FILTER_EVENT_NAME).Find(&eventNames).Error; err != nil {
		log.WithFields(log.Fields{"project_id": projectId}).WithError(err).Error("Failed getting filter_event_names")

		return nil, http.StatusInternalServerError
	}

	if len(eventNames) == 0 {
		return eventNames, http.StatusNotFound
	}

	return eventNames, http.StatusFound
}

// GetSmartEventFilterEventNames returns a list of all smart events
func (pg *Postgres) GetSmartEventFilterEventNames(projectID uint64, includeDeleted bool) ([]model.EventName, int) {
	db := C.GetServices().Db

	whereStmnt := "project_id = ? AND type IN(?)"
	if !includeDeleted {
		whereStmnt = whereStmnt + " AND deleted = 'false' "
	}

	var eventNames []model.EventName
	if err := db.Where(whereStmnt,
		projectID, []string{model.TYPE_CRM_SALESFORCE, model.TYPE_CRM_HUBSPOT}).Find(&eventNames).Error; err != nil {
		log.WithFields(log.Fields{"project_id": projectID}).WithError(err).Error("Failed getting filter_event_names")

		return nil, http.StatusInternalServerError
	}

	if len(eventNames) == 0 {
		return eventNames, http.StatusNotFound
	}

	return eventNames, http.StatusFound
}

// GetSmartEventFilterEventNameByID returns the smart event by event_name id
func (pg *Postgres) GetSmartEventFilterEventNameByID(projectID uint64, id string, isDeleted bool) (*model.EventName, int) {
	if id == "" || projectID == 0 {
		return nil, http.StatusBadRequest
	}

	whereStmnt := "project_id = ? AND type IN(?) AND id =? "
	if !isDeleted {
		whereStmnt = whereStmnt + " AND deleted = 'false' "
	}

	db := C.GetServices().Db

	var eventName model.EventName
	if err := db.Where(whereStmnt,
		projectID, []string{model.TYPE_CRM_SALESFORCE, model.TYPE_CRM_HUBSPOT}, id).First(&eventName).Error; err != nil {
		log.WithFields(log.Fields{"project_id": projectID}).WithError(err).Error("Failed getting smart event filter_event_name")

		return nil, http.StatusInternalServerError
	}

	return &eventName, http.StatusFound
}

// GetEventNamesByNames returns list of EventNames objects for given names
func (pg *Postgres) GetEventNamesByNames(projectId uint64, names []string) ([]model.EventName, int) {
	var eventNames []model.EventName

	db := C.GetServices().Db
	if err := db.Where("project_id = ? AND name IN (?)",
		projectId, names).Find(&eventNames).Error; err != nil {

		log.WithFields(log.Fields{"ProjectId": projectId}).WithError(err).Error(
			"failed to get event names")
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}
		return nil, http.StatusInternalServerError
	}

	return eventNames, http.StatusFound
}

func (pg *Postgres) GetFilterEventNamesByExprPrefix(projectId uint64, prefix string) ([]model.EventName, int) {
	var eventNames []model.EventName

	db := C.GetServices().Db
	if err := db.Where("project_id = ? AND type = ? AND filter_expr LIKE ? AND deleted = 'false'",
		projectId, model.TYPE_FILTER_EVENT_NAME, fmt.Sprintf("%s%%", prefix)).Find(&eventNames).Error; err != nil {
		log.WithFields(log.Fields{"projectId": projectId, "prefix": prefix}).WithError(err).Error(
			"filtering eventName failed on GetFilterEventNamesByExprPrefix")
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}

		return nil, http.StatusInternalServerError
	}

	if len(eventNames) == 0 {
		return nil, http.StatusNotFound
	}

	return eventNames, http.StatusFound
}

func (pg *Postgres) UpdateEventName(projectId uint64, id string,
	nameType string, eventName *model.EventName) (*model.EventName, int) {
	db := C.GetServices().Db

	// update not allowed for internal event names.
	if nameType == model.TYPE_INTERNAL_EVENT_NAME {
		return nil, http.StatusBadRequest
	}

	// Validation
	if projectId == 0 || eventName.ProjectId != 0 || !isValidType(nameType) ||
		!isValidName(eventName.Name, eventName.Type) {

		return nil, http.StatusBadRequest
	}

	var updatedEventName model.EventName
	updateFields := map[string]interface{}{}
	updateFields["name"] = eventName.Name

	query := db.Model(&updatedEventName).Where(
		"project_id = ? AND id = ? AND type = ?",
		projectId, id, nameType).Updates(updateFields)

	if err := query.Error; err != nil {
		log.WithFields(log.Fields{"event_name": eventName,
			"update_fields": updateFields,
		}).WithError(err).Error("Failed updating filter.")

		return nil, http.StatusInternalServerError
	}

	if query.RowsAffected == 0 {
		return nil, http.StatusBadRequest
	}

	return &updatedEventName, http.StatusAccepted
}

func (pg *Postgres) updateCRMSmartEventFilter(projectID uint64, id string, nameType string,
	eventName *model.EventName, filterExpr *model.SmartCRMEventFilter) (*model.EventName, int) {

	logCtx := log.WithFields(log.Fields{"project_id": projectID, "event_name_id": id, "event_name_type": nameType})
	// Validation
	if id == "" || projectID == 0 || eventName.ProjectId != 0 ||
		!isValidName(eventName.Name, eventName.Type) {
		logCtx.Error("Missing required Fields")
		return nil, http.StatusBadRequest
	}

	// update not allowed for non CRM based smart event.
	if !model.IsEventNameTypeSmartEvent(nameType) {
		return nil, http.StatusBadRequest
	}

	if eventName.FilterExpr != "" {
		return nil, http.StatusBadRequest
	}

	if filterExpr != nil && !model.IsValidSmartEventFilterExpr(filterExpr) {
		logCtx.WithField("filter_exp", *filterExpr).Error("Invalid smart event filter expression.")
		return nil, http.StatusBadRequest
	}

	db := C.GetServices().Db
	updateFields := map[string]interface{}{}
	updateFields["name"] = eventName.Name

	if eventName.Deleted == true {
		updateFields["deleted"] = false
	}

	if filterExpr != nil {
		prevEventName, status := pg.GetSmartEventFilterEventNameByID(projectID, id, eventName.Deleted)
		if status != http.StatusFound {
			return nil, http.StatusBadRequest
		}

		var existingFilterExp *model.SmartCRMEventFilter
		var err error

		existingFilterExp, err = model.GetDecodedSmartEventFilterExp(prevEventName.FilterExpr)
		if err != nil {
			log.WithError(err).Error("Failed to decode existing smart event Filter")
			return nil, http.StatusInternalServerError
		}

		if existingFilterExp.Source != filterExpr.Source {
			return nil, http.StatusBadRequest
		}

		enFilterExp, err := json.Marshal(filterExpr)
		if err != nil {
			log.WithError(err).Error("Failed to marshal new smart event Filter")
			return nil, http.StatusInternalServerError
		}

		if string(enFilterExp) != prevEventName.FilterExpr {
			updateFields["filter_expr"] = enFilterExp
		}

	}

	var updatedEventName model.EventName
	query := db.Model(&updatedEventName).Where(
		"project_id = ? AND id = ? AND type = ?",
		projectID, id, nameType).Updates(updateFields)

	if err := query.Error; err != nil {
		log.WithFields(log.Fields{"event_name": eventName,
			"update_fields": updateFields,
		}).WithError(err).Error("Failed updating smart event filter.")

		return nil, http.StatusInternalServerError
	}

	if query.RowsAffected == 0 {
		return nil, http.StatusBadRequest
	}

	return &updatedEventName, http.StatusAccepted
}

func getCRMSmartEventNameType(source string) string {
	if source == model.SmartCRMEventSourceSalesforce {
		return model.TYPE_CRM_SALESFORCE
	}

	if source == model.SmartCRMEventSourceHubspot {
		return model.TYPE_CRM_HUBSPOT
	}
	return ""
}

func (pg *Postgres) UpdateCRMSmartEventFilter(projectID uint64, id string, eventName *model.EventName,
	filterExpr *model.SmartCRMEventFilter) (*model.EventName, int) {

	_, duplicate := pg.checkDuplicateSmartEventFilter(projectID, filterExpr)
	if duplicate {
		return nil, http.StatusConflict
	}

	eventName.Type = getCRMSmartEventNameType(filterExpr.Source)

	return pg.updateCRMSmartEventFilter(projectID, id, eventName.Type, eventName, filterExpr)
}

// DeleteSmartEventFilter soft delete smart event name with filter expression
func (pg *Postgres) DeleteSmartEventFilter(projectID uint64, id string) (*model.EventName, int) {
	eventName, status := pg.GetSmartEventFilterEventNameByID(projectID, id, false)
	if status != http.StatusFound {
		return nil, http.StatusBadRequest
	}

	status = DeleteEventName(projectID, eventName.ID, eventName.Type)
	if status != http.StatusAccepted {
		return nil, http.StatusInternalServerError
	}

	return eventName, status
}

func (pg *Postgres) UpdateFilterEventName(projectId uint64, id string, eventName *model.EventName) (*model.EventName, int) {
	return pg.UpdateEventName(projectId, id, model.TYPE_FILTER_EVENT_NAME, eventName)
}

func DeleteEventName(projectId uint64, id string, nameType string) int {
	db := C.GetServices().Db

	// Validation
	if projectId == 0 {
		return http.StatusBadRequest
	}

	var updatedEventName model.EventName
	updateFields := map[string]interface{}{"deleted": true}

	query := db.Model(&updatedEventName).Where("project_id = ? AND id = ? AND type = ?",
		projectId, id, nameType).Updates(updateFields)

	if err := query.Error; err != nil {
		log.WithError(err).Error("Failed deleting filter.")

		return http.StatusInternalServerError
	}

	if query.RowsAffected == 0 {
		return http.StatusBadRequest
	}

	return http.StatusAccepted
}

func (pg *Postgres) DeleteFilterEventName(projectId uint64, id string) int {
	return DeleteEventName(projectId, id, model.TYPE_FILTER_EVENT_NAME)
}

// Returns sanitized filter expression and valid or not bool.
func getValidatedFilterExpr(filterExpr string) (string, bool) {
	if filterExpr == "" {
		return "", false
	}

	parsedURL, err := U.ParseURLStable(filterExpr)
	if err != nil {
		return "", false
	}

	if parsedURL.Host == "" {
		return "", false
	}

	noPath := parsedURL.Path == "" || parsedURL.Path == U.URI_SLASH
	noHashPath := parsedURL.Fragment == ""

	var path string
	if noPath && noHashPath {
		path = U.URI_SLASH
	} else {
		path = U.GetURLPathWithHash(parsedURL)
	}

	return fmt.Sprintf("%s%s", parsedURL.Host, path), true
}

// calculateDefinitionScore -  calculates score based on number of property_token
// and number of defined static_token.
// algo: increament on static_token (u1), decreament on property_token (v1).
// ["u1", "u2", "u3"] -> 3
// ["u1", "u2", ":v1"] -> 1
// ["u1", "u2", ":v1", ":v2"] -> 0
func calculateDefinitionScore(tokenizedFilter []string) int16 {
	if len(tokenizedFilter) == 0 {
		return -9999
	}

	var score int16 = 0
	for _, token := range tokenizedFilter {
		if strings.HasPrefix(token, model.URI_PROPERTY_PREFIX) {
			score = score - 1
		} else {
			score = score + 1
		}
	}

	return score
}

type FilterInfo struct {
	// filter_expr split by URI_SLASH
	tokenizedFilter []string
	eventName       *model.EventName
}

// getHighDefinitionFilter - Returns filter with high definition score.
func getHighDefinitionFilter(filters []*FilterInfo) *FilterInfo {
	if filtersLen := len(filters); filtersLen == 0 {
		return nil
	} else if filtersLen == 1 {
		return filters[0]
	}

	var highScoredFilter *FilterInfo
	var score int16 = -9999

	for _, f := range filters {
		// calculating definity score for filter_expr everytime,
		// can be avoided by memoizing it on db column during expr insert?
		defScore := calculateDefinitionScore(f.tokenizedFilter)
		if defScore > score {
			score = defScore
			highScoredFilter = f
		}
	}

	return highScoredFilter
}

// matchEventNameWithFilters match uri by passing through filters.
func matchEventURIWithFilters(filters *[]FilterInfo,
	tokenizedEventURI []string) (*FilterInfo, bool) {
	if len(tokenizedEventURI) == 0 {
		return nil, false
	}

	matches := make([]*FilterInfo, 0, 0)
	for i, finfo := range *filters {
		if model.IsFilterMatch(finfo.tokenizedFilter, tokenizedEventURI) {
			matches = append(matches, &(*filters)[i])
		}
	}

	// Get one high definition filter from matches.
	matchedFilter := getHighDefinitionFilter(matches)
	if matchedFilter == nil {
		return nil, false
	}

	return matchedFilter, true
}

// popAndMatchEventURIWithFilters - Pops event_uri by slash and
// compare after_popped_uri with filter.
func popAndMatchEventURIWithFilters(filters *[]FilterInfo,
	eventURI string) (*FilterInfo, bool) {

	for afterPopURI := eventURI; afterPopURI != ""; afterPopURI, _ = U.PopURIBySlash(afterPopURI) {
		tokenizedEventURI := U.TokenizeURI(afterPopURI)

		if finfo, matched := matchEventURIWithFilters(
			filters, tokenizedEventURI); matched {
			return finfo, true
		}
	}

	return nil, false
}

func makeFilterInfos(eventNames []model.EventName) (*[]FilterInfo, error) {
	// Selected list of filters to use after pruning.
	filters := make([]FilterInfo, len(eventNames))
	for i := 0; i < len(eventNames); i++ {
		// Todo(Dinesh): Can be removed if we store domain seperately.
		parsedFilterExpr, err := U.ParseURLWithoutProtocol(eventNames[i].FilterExpr)
		if err != nil {
			log.WithFields(log.Fields{
				"filter_expr": eventNames[i].FilterExpr,
			}).WithError(err).Error(
				"Failed parsing filter_expr. Insert validator might be failing.")
			return nil, err
		}

		tokenizedFilter := U.TokenizeURI(U.GetURLPathWithHash(parsedFilterExpr))

		filters[i] = FilterInfo{
			tokenizedFilter: tokenizedFilter,
			eventName:       &eventNames[i]}
	}

	return &filters, nil
}

// FilterEventNameByEventURL - Filter and return an event_name by event_url.
func (pg *Postgres) FilterEventNameByEventURL(projectId uint64, eventURL string) (*model.EventName, int) {
	if projectId == 0 && eventURL == "" {
		return nil, http.StatusBadRequest
	}

	parsedEventURL, err := U.ParseURLStable(eventURL)
	if err != nil {
		log.WithFields(log.Fields{"event_url": eventURL}).WithError(err).Warn(
			"Failed parsing event_url.")
		return nil, http.StatusBadRequest
	}

	// Get expressions with same domain prefix.
	eventNames, errCode := pg.GetFilterEventNamesByExprPrefix(projectId,
		parsedEventURL.Host)
	if errCode != http.StatusFound {
		return nil, errCode
	}

	filters, err := makeFilterInfos(eventNames)
	if err != nil {
		return nil, http.StatusInternalServerError
	}

	if len(*filters) == 0 {
		return nil, http.StatusNotFound
	}

	filterInfo, matched := popAndMatchEventURIWithFilters(filters,
		U.GetURLPathWithHash(parsedEventURL))
	if !matched {
		return nil, http.StatusNotFound
	}

	return filterInfo.eventName, http.StatusFound
}

func (pg *Postgres) GetEventNameFromEventNameId(eventNameId string, projectId uint64) (*model.EventName, error) {
	db := C.GetServices().Db
	var eventName model.EventName
	queryStr := "SELECT * FROM event_names WHERE id = ? AND project_id = ?"
	err := db.Raw(queryStr, eventNameId, projectId).Scan(&eventName).Error
	if err != nil {
		log.Error("Failed to get event_name from event_name_id")
		return nil, err
	}
	return &eventName, nil
}

func convert(eventNamesWithAggregation []model.EventNameWithAggregation) []model.EventName {
	eventNames := make([]model.EventName, 0)
	for _, event := range eventNamesWithAggregation {
		eventNames = append(eventNames, model.EventName{
			ID:         event.ID,
			Name:       event.Name,
			CreatedAt:  event.CreatedAt,
			Deleted:    event.Deleted,
			FilterExpr: event.FilterExpr,
			ProjectId:  event.ProjectId,
			Type:       event.Type,
			UpdatedAt:  event.UpdatedAt,
		})
	}
	return eventNames
}

func (pg *Postgres) GetEventTypeFromDb(
	projectID uint64, eventNames []string, limit int64) (map[string]string, error) {
	// var err string
	db := C.GetServices().Db
	hasLimit := limit > 0

	logCtx := log.WithFields(log.Fields{"projectId": projectID})

	type allEventNameAndType struct {
		Name string
		Type string
	}
	EventNameType := make(map[string]string)
	var tmpEventNames []string
	for _, b := range eventNames {
		//to remove empty strings which might break sql query
		if len(b) > 0 {
			tmpEventNames = append(tmpEventNames, b)
		}
	}

	queryStr := "SELECT name,type FROM event_names WHERE project_id=? and name IN (?)"

	if hasLimit {
		queryStr = queryStr + " " + "LIMIT ?"
	}

	params := make([]interface{}, 0)
	params = append(params, projectID)
	params = append(params, tmpEventNames)
	if hasLimit {
		params = append(params, limit)
	}

	rows, err := db.Raw(queryStr, params...).Rows()
	if err != nil {
		logCtx.WithError(err).Error(
			"Failed scanning rows on get event names and Types.")
		return EventNameType, err
	}

	for rows.Next() {
		var eventNameAndType allEventNameAndType
		if err := db.ScanRows(rows, &eventNameAndType); err != nil {
			logCtx.WithError(err).Error("Failed scanning rows on get event names and Types.")
			return EventNameType, err
		}
		EventNameType[eventNameAndType.Name] = eventNameAndType.Type
	}
	return EventNameType, nil
}
