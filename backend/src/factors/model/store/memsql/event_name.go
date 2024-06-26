package memsql

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"factors/cache"
	pCache "factors/cache/persistent"
	cacheRedis "factors/cache/redis"
	C "factors/config"
	U "factors/util"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"

	"factors/model/model"
)

var BLACKLISTED_EVENTS_FOR_EVENT_PROPERTIES = map[string]string{
	"$hubspot_contact_created": "$hubspot_",
	"$hubspot_contact_updated": "$hubspot_",
	"$hubspot_company_":        "$hubspot_",
	"$hubspot_deal_":           "$hubspot_",
	"$sf_contact_":             "$salesforce_",
	"$sf_lead_":                "$salesforce_",
	"$sf_account_":             "$salesforce_",
	"$sf_opportunity_":         "$salesforce_",
}

func satisfiesEventNameConstraints(eventName model.EventName) int {
	logFields := log.Fields{
		"event_name": eventName,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	// Unique (project_id, name, type) WHERE type != 'FE'.
	_, errCode := getNonFilterEventsByName(eventName.ProjectId, eventName.Name)
	if errCode == http.StatusFound {
		return http.StatusConflict
	}

	// Unique (project_id, type, filter_expr)
	if eventName.FilterExpr != "" {
		errCode := isEventNameExistByTypeAndFitlerExpr(eventName.ProjectId,
			eventName.Type, eventName.FilterExpr)
		if errCode == http.StatusFound {
			return http.StatusConflict
		}
	}

	return http.StatusOK
}

func (store *MemSQL) satisfiesEventNameForeignConstraints(eventName model.EventName) int {
	logFields := log.Fields{
		"event_name": eventName,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	_, errCode := store.GetProject(eventName.ProjectId)
	if errCode != http.StatusFound {
		return http.StatusBadRequest
	}
	return http.StatusOK
}

func (store *MemSQL) CreateOrGetEventName(eventName *model.EventName) (*model.EventName, int) {

	logFields := log.Fields{
		"event_name": eventName,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)

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
	} else if errCode := satisfiesEventNameConstraints(*eventName); errCode != http.StatusOK {
		return nil, errCode
	} else if errCode := store.satisfiesEventNameForeignConstraints(*eventName); errCode != http.StatusOK {
		return nil, http.StatusInternalServerError
	}

	if eventName.ID == "" {
		eventName.ID = U.GetUUID()
	}
	if err := db.Create(eventName).Error; err != nil {
		logCtx.WithError(err).Error("Failed to create event_name.")
		return nil, http.StatusInternalServerError
	}

	return eventName, http.StatusCreated
}

func (store *MemSQL) CreateOrGetUserCreatedEventName(eventName *model.EventName) (*model.EventName, int) {
	logFields := log.Fields{
		"event_name": eventName,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	eventName.Type = model.TYPE_USER_CREATED_EVENT_NAME
	return store.CreateOrGetEventName(eventName)
}

func (store *MemSQL) GetEventNameIdsWithGivenNames(projectID int64, eventNameIDsList map[string]bool) (map[string]string, int) {
	params := []interface{}{projectID}
	queryStr := "SELECT name, id from event_names where project_id = ? AND ("
	eventNameIDsMap := make(map[string]string)
	for name := range eventNameIDsList {
		params = append(params, name)
		if len(params) == 2 {
			queryStr = queryStr + "name = ?"
			continue
		}
		queryStr = queryStr + " OR name = ?"
	}

	queryStr = queryStr + ")"

	db := C.GetServices().Db

	var eventNames []model.EventName
	if err := db.Raw(queryStr, params...).Scan(&eventNames).Error; err != nil {
		log.WithFields(log.Fields{"project_id": projectID}).WithError(err).Error("Failed getting event_names")

		return map[string]string{}, http.StatusInternalServerError
	}

	if len(eventNames) < 1 {
		return map[string]string{}, http.StatusInternalServerError
	}

	for _, eventName := range eventNames {
		eventNameIDsMap[eventName.Name] = eventName.ID
	}

	return eventNameIDsMap, http.StatusFound

}

func (store *MemSQL) CreateOrGetAutoTrackedEventName(eventName *model.EventName) (*model.EventName, int) {
	logFields := log.Fields{
		"event_name": eventName,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	eventName.Type = model.TYPE_AUTO_TRACKED_EVENT_NAME
	return store.CreateOrGetEventName(eventName)
}

func (store *MemSQL) CreateOrGetFilterEventName(eventName *model.EventName) (*model.EventName, int) {
	logFields := log.Fields{
		"event_name": eventName,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	filterExpr, valid := getValidatedFilterExpr(eventName.FilterExpr)
	if !valid {
		return nil, http.StatusBadRequest
	}

	eventName.Type = model.TYPE_FILTER_EVENT_NAME
	eventName.FilterExpr = filterExpr

	return store.CreateOrGetEventName(eventName)
}

func (store *MemSQL) checkDuplicateSmartEventFilter(projectID int64, inFilterExpr *model.SmartCRMEventFilter) (*model.EventName, bool) {
	logFields := log.Fields{
		"project_id":     projectID,
		"in_filter_expr": inFilterExpr,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	eventNames, status := store.GetSmartEventFilterEventNames(projectID, true)
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
func (store *MemSQL) CreateOrGetCRMSmartEventFilterEventName(projectID int64, eventName *model.EventName,
	filterExpr *model.SmartCRMEventFilter) (*model.EventName, int) {
	logFields := log.Fields{
		"project_id":  projectID,
		"event_name":  eventName,
		"filter_expr": filterExpr,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)
	if !model.IsValidSmartEventFilterExpr(filterExpr) || filterExpr == nil || eventName.Type != "" ||
		eventName.Name == "" {
		logCtx.Error("Invalid fields.")
		return nil, http.StatusBadRequest
	}

	dupEventName, duplicate := store.checkDuplicateSmartEventFilter(projectID, filterExpr)
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

			updatedEventName, status := store.updateCRMSmartEventFilter(projectID, dupEventName.ID, updateEventName.Type, updateEventName, filterExpr)
			if status != http.StatusAccepted {
				logCtx.WithField("err_code", status).Error("Failed to update deleted smart event filter.")
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

	_, status := store.CreateOrGetEventName(eventName)
	if status != http.StatusCreated {
		logCtx.WithField("err_code", status).Error("Failed to CreateOrGetCRMSmartEventFilterEventName.")
		return nil, http.StatusInternalServerError
	}

	return eventName, status
}

func (store *MemSQL) GetSmartEventEventName(eventName *model.EventName) (*model.EventName, int) {
	logFields := log.Fields{
		"event_name": eventName,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	return store.GetSmartEventEventNameByNameANDType(eventName.ProjectId, eventName.Name, eventName.Type)
}

func (store *MemSQL) GetSmartEventEventNameByNameANDType(projectID int64, name, typ string) (*model.EventName, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"name":       name,
		"typ":        typ,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	if projectID == 0 || name == "" || typ == "" {
		return nil, http.StatusBadRequest
	}

	db := C.GetServices().Db

	var eventNames []model.EventName
	if err := db.Where("project_id = ? AND type = ? AND name = ? and deleted = 0",
		projectID, typ, name).Find(&eventNames).Error; err != nil {
		log.WithFields(log.Fields{"project_id": projectID}).WithError(err).Error("Failed getting filter_event_names")

		return nil, http.StatusInternalServerError
	}

	if len(eventNames) != 1 {
		return nil, http.StatusInternalServerError
	}

	return &eventNames[0], http.StatusFound
}

func (store *MemSQL) CreateOrGetSessionEventName(projectId int64) (*model.EventName, int) {
	logFields := log.Fields{
		"project_id": projectId,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	return store.CreateOrGetEventName(&model.EventName{ProjectId: projectId, Name: U.EVENT_NAME_SESSION,
		Type: model.TYPE_INTERNAL_EVENT_NAME})
}

func (store *MemSQL) CreateOrGetOfflineTouchPointEventName(projectId int64) (*model.EventName, int) {
	logFields := log.Fields{
		"project_id": projectId,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	return store.CreateOrGetEventName(&model.EventName{ProjectId: projectId, Name: U.EVENT_NAME_OFFLINE_TOUCH_POINT,
		Type: model.TYPE_INTERNAL_EVENT_NAME})
}

func (store *MemSQL) GetSessionEventName(projectId int64) (*model.EventName, int) {
	logFields := log.Fields{
		"project_id": projectId,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	if projectId == 0 {
		return nil, http.StatusBadRequest
	}

	logCtx := log.WithFields(logFields)
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
	logFields := log.Fields{
		"name_type": nameType,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
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
	logFields := log.Fields{
		"name": name,
		"typ":  typ,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
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

func (store *MemSQL) GetEventName(name string, projectId int64) (*model.EventName, int) {
	logFields := log.Fields{
		"project_id": projectId,
		"name":       name,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	// Input Validation. (ID is to be auto generated)
	if name == "" || projectId == 0 {
		log.Error("GetEventName Failed. Missing name or projectId")
		return nil, http.StatusBadRequest
	}

	// invalid event names from frontend state issues.
	if name == "null" || name == "undefined" {
		return nil, http.StatusNotFound
	}

	// Get event name from cache
	eventNameCache, err := GetEventNameFromCache(projectId, name)
	if err == nil {
		return &eventNameCache, http.StatusFound
	}

	// Get event name from Query
	var eventName model.EventName
	db := C.GetServices().Db
	if err := db.Limit(1).
		Where("project_id = ? AND name = ?", projectId, name).
		Find(&eventName).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}

		log.WithFields(logFields).WithError(err).Error("Failed to get event_name.")
		return nil, http.StatusInternalServerError
	}

	// Cache the retrieved event name
	SetEventNameCache(projectId, name, eventName)

	return &eventName, http.StatusFound
}

func (store *MemSQL) GetEventNames(projectId int64) ([]model.EventName, int) {
	logFields := log.Fields{
		"project_id": projectId,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
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
func (store *MemSQL) IsEventExistsWithType(projectId int64, eventType string) (bool, int) {
	logFields := log.Fields{
		"project_id": projectId,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	if projectId == 0 {
		log.Error("GetEventExists With Type Failed. Missing projectId")
		return false, http.StatusBadRequest
	}

	db := C.GetServices().Db

	var eventName model.EventName
	if err := db.Order("created_at ASC").Where("project_id = ? AND type = ?", projectId, eventType).Limit(1).Find(&eventName).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return false, http.StatusNotFound
		}
		return false, http.StatusInternalServerError
	}
	if eventName.Type != eventType {
		return false, http.StatusNotFound
	}
	return true, http.StatusFound
}

func (store *MemSQL) GetDomainNamesByProjectID(projectId int64) ([]string, int) {
	logFields := log.Fields{
		"project_id": projectId,
	}
	domainNames := make([]string, 0)
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	if projectId == 0 {
		log.Error("GetEventExists With Type Failed. Missing projectId")
		return domainNames, http.StatusBadRequest
	}

	db := C.GetServices().Db

	query := " select DISTINCT(LEFT(JSON_EXTRACT_STRING(properties, '$page_url'), POSITION('/' in  JSON_EXTRACT_STRING(properties, '$page_url') ) - 1)) as domain from events where project_id = ? and timestamp > (SELECT MAX(timestamp) - 604800 FROM events WHERE project_id= ? ) and JSON_EXTRACT_STRING(properties, '$is_page_view') = 'true' limit 25000 "

	rows, err := db.Raw(query, projectId, projectId).Rows()
	if err != nil {
		log.WithError(err).Error("Failed to get GetEventNamesByType")
		return domainNames, http.StatusInternalServerError
	}

	for rows.Next() {
		var domainName string
		if err := rows.Scan(&domainName); err != nil {
			log.WithError(err).
				Error("Failed to scan row on GetEventNamesByType.")
			continue
		}
		if domainName != "" {
			domainNames = append(domainNames, domainName)
		}
	}

	return domainNames, http.StatusFound
}

// GetOrderedEventNamesFromDb - Get 'limit' events from DB sort by occurence for a given time period
func (store *MemSQL) GetOrderedEventNamesFromDb(
	projectID int64, startTimestamp int64, endTimestamp int64, limit int) ([]model.EventNameWithAggregation, error) {
	logFields := log.Fields{
		"project_id":      projectID,
		"start_timestamp": startTimestamp,
		"end_timestamp":   endTimestamp,
		"limit":           limit,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db
	hasLimit := limit > 0
	eventNames := make([]model.EventNameWithAggregation, 0)

	logCtx := log.WithFields(logFields)

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

func (store *MemSQL) GetDisplayNamesForEventProperties(projectId int64, properties map[string][]string, eventName string) map[string]string {
	logCtx := log.WithFields(log.Fields{
		"projectId": projectId,
	})

	_, displayNames := store.GetDisplayNamesForAllEventProperties(projectId, eventName)
	standardPropertiesAllEvent := U.STANDARD_EVENT_PROPERTIES_DISPLAY_NAMES
	displayNamesOp := make(map[string]string)
	for property, displayName := range standardPropertiesAllEvent {
		displayNamesOp[property] = strings.Title(displayName)
	}
	if eventName == U.EVENT_NAME_SESSION {
		standardPropertiesSession := U.STANDARD_SESSION_PROPERTIES_DISPLAY_NAMES
		for property, displayName := range standardPropertiesSession {
			displayNamesOp[property] = strings.Title(displayName)
		}
	}
	for property, displayName := range displayNames {
		displayNamesOp[property] = strings.Title(displayName)
	}

	_, displayNames = store.GetDisplayNamesForObjectEntities(projectId)
	for property, displayName := range displayNames {
		displayNamesOp[property] = strings.Title(displayName)
	}
	for _, props := range properties {
		for _, prop := range props {
			displayName := U.CreateVirtualDisplayName(prop)
			_, exist := displayNamesOp[prop]
			if !exist {
				displayNamesOp[prop] = displayName
			}
		}
	}
	dupCheck := make(map[string]bool)
	for _, name := range displayNamesOp {
		_, exists := dupCheck[name]
		if exists {
			logCtx.Warning(fmt.Sprintf("Duplicate display name %s", name))
		}
		dupCheck[name] = true
	}
	return displayNamesOp
}

// GetPropertyValuesByEventProperty (Part of event_name and properties caching) This method iterates for
// last n days to get all the top 'limit' property values for the given property/event
// Picks all last 24 hours values and sorts the remaining by occurence and returns top 'limit' values
func (store *MemSQL) GetPropertyValuesByEventProperty(projectID int64, eventName string,
	propertyName string, limit int, lastNDays int) ([]string, error) {
	logFields := log.Fields{
		"project_id":    projectID,
		"event_name":    eventName,
		"property_name": propertyName,
		"limit":         limit,
		"last_N_days":   lastNDays,
	}
	logCtx := log.WithFields(logFields)
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

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

	isAggregateCacheUnavailable := false
	var valuesAggregated []U.NameCountTimestampCategory
	if C.IsAggrEventPropertyValuesCacheEnabled(projectID) {
		eventPropertyValuesAggCacheKey, err := model.GetValuesByEventPropertyRollUpAggregateCacheKey(
			projectID, eventName, propertyName)
		if err != nil {
			logCtx.WithError(err).Error("Failed to get event property rollup agg cache key.")
		}

		var existingAggregate U.CacheEventPropertyValuesAggregate
		existingAggCache, isExists, err := pCache.GetIfExists(eventPropertyValuesAggCacheKey, true)
		if err != nil {
			logCtx.WithError(err).Error("Failed to get cache aggregate on API.")
			return []string{}, err
		}

		if isExists {
			// Merge aggregated rollup,  current date rollup and create new aggregate.
			err = json.Unmarshal([]byte(existingAggCache), &existingAggregate)
			if err != nil {
				logCtx.WithError(err).Error("Failed to unmarshal values cache aggregate on API.")
				return []string{}, err
			}

			logCtx.
				WithField("len_agg_keys", len(existingAggregate.NameCountTimestampCategoryList)).
				Info("Values are from aggregate cache.")

			currentDateOnlyFormat := U.TimeNowZ().Format(U.DATETIME_FORMAT_YYYYMMDD)
			currentDateValues, err := getPropertyValuesByEventPropertyFromCache(projectID, eventName, propertyName, currentDateOnlyFormat)
			if err != nil {
				logCtx.WithError(err).Error("Failed to current date rollup values to add to aggregate.")
			}

			valuesListFromAggMap := make(map[string]U.CountTimestampTuple)
			for i := range existingAggregate.NameCountTimestampCategoryList {
				exa := existingAggregate.NameCountTimestampCategoryList[i]
				valuesListFromAggMap[exa.Name] = U.CountTimestampTuple{LastSeenTimestamp: exa.Timestamp, Count: exa.Count}
			}

			valuesListFromAgg := U.CachePropertyValueWithTimestamp{PropertyValue: valuesListFromAggMap}
			valuesList := make([]U.CachePropertyValueWithTimestamp, 0)
			valuesList = append(valuesList, currentDateValues)
			valuesList = append(valuesList, valuesListFromAgg)

			valuesAggregated = U.AggregatePropertyValuesAcrossDate(valuesList, false, 0)
		} else {
			isAggregateCacheUnavailable = true
		}
	}

	if isAggregateCacheUnavailable || !C.IsAggrEventPropertyValuesCacheEnabled(projectID) {
		values := make([]U.CachePropertyValueWithTimestamp, 0)
		for i := 0; i < lastNDays; i++ {
			currentDateOnlyFormat := currentDate.AddDate(0, 0, -i).Format(U.DATETIME_FORMAT_YYYYMMDD)
			value, err := getPropertyValuesByEventPropertyFromCache(projectID, eventName, propertyName, currentDateOnlyFormat)
			if err != nil {
				return []string{}, err
			}
			values = append(values, value)
		}

		valuesAggregated = U.AggregatePropertyValuesAcrossDate(values, false, 0)
	}

	valueStrings := make([]string, 0)
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

func getPropertyValuesByEventPropertyFromCache(projectID int64, eventName string, propertyName string, dateKey string) (U.CachePropertyValueWithTimestamp, error) {
	logFields := log.Fields{
		"project_id":    projectID,
		"event_name":    eventName,
		"property_name": propertyName,
		"date_key":      dateKey,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)
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
	values, _, err := pCache.GetIfExists(eventPropertyValuesKey, true)
	if values == "" {
		eventNameWithSlash := fmt.Sprintf("%s/", eventName)
		eventPropertyValuesKeyWithSlash, err := model.GetValuesByEventPropertyRollUpCacheKey(projectID, eventNameWithSlash, propertyName, dateKey)
		if err != nil {
			return U.CachePropertyValueWithTimestamp{}, err
		}
		valuesWithSlash, _, err := pCache.GetIfExists(eventPropertyValuesKeyWithSlash, true)
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

// Fetches properties from Cache and filter the required properties based on the eventName provided.
func (store *MemSQL) GetEventPropertiesAndModifyResultsForNonExplain(projectId int64,
	eventName string) (map[string][]string, int) {

	logCtx := log.WithFields(log.Fields{
		"projectId":  projectId,
		"event_name": eventName,
	})

	properties := make(map[string][]string)
	propertiesFromCache, err := store.GetPropertiesByEvent(projectId, eventName, 2500,
		C.GetLookbackWindowForEventUserCache())
	toBeFiltered := false
	propertyPrefixToRemove := ""

	// By default, hubspot or salesforce prefixed properties are not be added projects. If a project is given in input,
	// EnableEventLevelEventProperties is true, hubspot or salesforce properties are to be added.
	enableEventLevelEventProperties := C.EnableEventLevelEventProperties(projectId)
	for eventPrefix, propertyPrefix := range BLACKLISTED_EVENTS_FOR_EVENT_PROPERTIES {
		if strings.HasPrefix(eventName, eventPrefix) && !enableEventLevelEventProperties {
			propertyPrefixToRemove = propertyPrefix
			toBeFiltered = true
			break
		}
	}
	if toBeFiltered == true {
		for category, props := range propertiesFromCache {
			if properties[category] == nil {
				properties[category] = make([]string, 0)
			}
			for _, property := range props {
				if !strings.HasPrefix(property, propertyPrefixToRemove) {
					properties[category] = append(properties[category], property)
				}
			}
		}
	} else {
		properties = propertiesFromCache
	}
	if err != nil {
		logCtx.WithError(err).Error("Failed to get properties by event")
		return make(map[string][]string, 0), http.StatusInternalServerError
	}

	if len(properties) == 0 {
		logCtx.WithError(err).Warn("No event properties returned")
	}

	return properties, http.StatusOK
}

// GetPropertiesByEvent (Part of event_name and properties caching) This method iterates for last n days to get all the
// top 'limit' properties for the given event. Picks all last 24 hours properties and sorts the remaining by occurence
// and returns top 'limit' properties
func (store *MemSQL) GetPropertiesByEvent(projectID int64, eventName string, limit int, lastNDays int) (map[string][]string, error) {
	logFields := log.Fields{
		"project_id":  projectID,
		"event_name":  eventName,
		"limit":       limit,
		"last_N_days": lastNDays,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
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

	propertyDetails, propertyDetailsStatus := store.GetAllPropertyDetailsByProjectID(projectID, eventName, false)
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

func getPropertiesByEventFromCache(projectID int64, eventName string, dateKey string) (U.CachePropertyWithTimestamp, error) {
	logFields := log.Fields{
		"project_id": projectID,
		"event_name": eventName,
		"date_key":   dateKey,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)
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
	eventProperties, _, err := pCache.GetIfExists(eventPropertiesKey, true)
	if eventProperties == "" {
		eventNameWithSlash := fmt.Sprintf("%s/", eventName)
		eventPropertiesKeyWithSlash, err := model.GetPropertiesByEventCategoryRollUpCacheKey(projectID, eventNameWithSlash, dateKey)
		if err != nil {
			return U.CachePropertyWithTimestamp{}, err
		}
		eventPropertiesWithSlash, _, err := pCache.GetIfExists(eventPropertiesKeyWithSlash, true)
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
	logFields := log.Fields{
		"events": events,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
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

			if eventDetails.Type == model.EVENT_NAME_TYPE_PAGE_VIEW_EVENT {
				eventsAggregatedInt.Type = U.PageViewEvent
			}

			eventsAggregated[eventNameSuffixTrim] = eventsAggregatedInt
		}
	}
	eventsAggregatedSlice := make([]U.NameCountTimestampCategory, 0)
	for k, v := range eventsAggregated {
		eventsAggregatedSlice = append(eventsAggregatedSlice, U.NameCountTimestampCategory{
			Name: k, Count: v.Count, Timestamp: v.LastSeenTimestamp, Category: v.Type, GroupName: ""})
	}
	return eventsAggregatedSlice
}

// Hacked solution - to fetch a type of EventNames.
func (store *MemSQL) GetMostFrequentlyEventNamesByType(projectID int64, limit int, lastNDays int, typeOfEvent string) ([]string, error) {
	logFields := log.Fields{
		"project_id":    projectID,
		"type_of_event": typeOfEvent,
		"limit":         limit,
		"last_N_days":   lastNDays,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	mostFrequentEventNames := make([]string, 0)
	var eventNameTypes []string
	var exists bool
	var finalEventNames []model.EventName
	var result []string
	db := C.GetServices().Db

	if eventNameTypes, exists = model.EventTypeToEnameType[typeOfEvent]; !exists {
		return nil, errors.New("invalid type is provided.")
	}
	eventsSorted, err := getEventNamesAggregatedAndSortedAcrossDate(projectID, limit, lastNDays)
	if err != nil {
		return nil, err
	}
	if typeOfEvent == model.PageViewsDisplayCategory {
		for _, event := range eventsSorted {
			if event.Category == U.PageViewEvent {
				mostFrequentEventNames = append(mostFrequentEventNames, event.Name)
			}
		}
	} else {
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
	}

	if limit > 0 {
		sliceLength := len(mostFrequentEventNames)
		if sliceLength > limit*2 {
			mostFrequentEventNames = mostFrequentEventNames[0 : limit*2]
		}
	}

	if typeOfEvent == model.PageViewsDisplayCategory {

		result = mostFrequentEventNames

	} else {

		if dbResult := db.Where("type IN (?) AND name IN (?)", eventNameTypes, mostFrequentEventNames).Select("name").Limit(limit).Find(&finalEventNames); dbResult.Error != nil {
			return nil, dbResult.Error
		}

		hashMapOfFinalEventNames := make(map[string]int)
		for _, eventName := range finalEventNames {
			hashMapOfFinalEventNames[eventName.Name] = 1
		}

		// Commenting I dont see this as reachable code.
		// if typeOfEvent == model.PageViewsDisplayCategory {
		// 	if _, ok := hashMapOfFinalEventNames[U.EVENT_NAME_FORM_SUBMITTED]; ok {
		// 		delete(hashMapOfFinalEventNames, U.EVENT_NAME_FORM_SUBMITTED)
		// 	}
		// }

		for _, event := range mostFrequentEventNames {
			if _, ok := hashMapOfFinalEventNames[event]; ok {
				result = append(result, event)
			}
		}
	}

	return result, nil
}

// GetEventNamesOrderedByOccurenceAndRecency (Part of event_name and properties caching) This method iterates for last n days to
// get all the top 'limit' events for the given project. Picks all last 24 hours events and sorts the remaining by
// occurence and returns top 'limit' events
func (store *MemSQL) GetEventNamesOrderedByOccurenceAndRecency(projectID int64, limit int, lastNDays int) (map[string][]string, error) {
	logFields := log.Fields{
		"project_id":  projectID,
		"limit":       limit,
		"last_N_days": lastNDays,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
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

	trackedEvents, _ := store.GetAllFactorsTrackedEventsByProject(projectID)
	eventsMap := make(map[string]bool)

	eventStringWithGroups := make(map[string][]string)
	for _, event := range eventsFiltered {
		eventsMap[event.Name] = true
		eventStringWithGroups[event.GroupName] = append(eventStringWithGroups[event.GroupName], event.Name)
	}
	for _, trackedEvent := range trackedEvents {
		if eventsMap[trackedEvent.Name] != true {
			eventsMap[trackedEvent.Name] = true
			eventStringWithGroups[U.FrequentlySeen] = append(eventStringWithGroups[U.FrequentlySeen], trackedEvent.Name)
		}
	}

	allStandardEvents, _ := store.GetEventNamesByNames(projectID, U.STANDARD_EVENTS_IN_DROPDOWN)
	for _, standardEvent := range allStandardEvents {
		if eventsMap[standardEvent.Name] != true {
			eventsMap[standardEvent.Name] = true
			eventStringWithGroups[U.FrequentlySeen] = append(eventStringWithGroups[U.FrequentlySeen], standardEvent.Name)
		}
	}
	return eventStringWithGroups, nil
}

func getEventNamesAggregatedAndSortedAcrossDate(projectID int64, limit int, lastNDays int) ([]U.NameCountTimestampCategory, error) {
	logFields := log.Fields{
		"project_id":  projectID,
		"limit":       limit,
		"last_N_days": lastNDays,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
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

func getEventNamesOrderedByOccurenceAndRecencyFromCache(projectID int64, dateKey string) (model.CacheEventNamesWithTimestamp, error) {
	logFields := log.Fields{
		"project_id": projectID,
		"date_key":   dateKey,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)
	if projectID == 0 {
		return model.CacheEventNamesWithTimestamp{},
			errors.New("invalid project on get event names ordered by occurence and recency from cache")
	}

	eventNamesKey, err := model.GetEventNamesOrderByOccurrenceAndRecencyRollUpCacheKey(projectID, dateKey)
	if err != nil {
		return model.CacheEventNamesWithTimestamp{}, err
	}
	eventNames, _, err := pCache.GetIfExists(eventNamesKey, true)
	if eventNames == "" {
		logCtx.WithError(err).WithField("date_key", dateKey).Info("MISSING ROLLUP EN")
		return model.CacheEventNamesWithTimestamp{}, nil
	}
	var cacheEventNames model.CacheEventNamesWithTimestamp
	err = json.Unmarshal([]byte(eventNames), &cacheEventNames)
	if err != nil {
		keyString, _ := eventNamesKey.Key()
		log.WithError(err).
			WithField("event_names", eventNames).
			WithField("key", keyString).
			Warn("JSON unmarshal error")
		return model.CacheEventNamesWithTimestamp{}, err
	}
	return cacheEventNames, nil
}

func getNonFilterEventsByName(projectID int64, eventName string) ([]model.EventName, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"event_name": eventName,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db

	var eventNames []model.EventName
	if err := db.Where("project_id = ? AND type != ? AND name = ? AND deleted = 0",
		projectID, model.TYPE_FILTER_EVENT_NAME, eventName).Find(&eventNames).Error; err != nil {
		log.WithFields(log.Fields{"project_id": projectID}).WithError(err).Error("Failed getting filter_events_by_name")

		return nil, http.StatusInternalServerError
	}

	if len(eventNames) == 0 {
		return eventNames, http.StatusNotFound
	}

	return eventNames, http.StatusFound
}

func isEventNameExistByTypeAndFitlerExpr(projectID int64, typ string, filterExpr string) int {
	logFields := log.Fields{
		"project_id":  projectID,
		"typ":         typ,
		"filter_expr": filterExpr,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db

	var eventNames model.EventName
	if err := db.Limit(1).Where("project_id = ? AND type = ? AND filter_expr = ? AND deleted = 0",
		projectID, typ, filterExpr).Select("id").Find(&eventNames).Error; err != nil {

		if gorm.IsRecordNotFoundError(err) {
			return http.StatusNotFound
		}

		log.WithFields(log.Fields{"project_id": projectID}).WithError(err).
			Error("Failed getting to check existence of event_name by type and filter_expr")

		return http.StatusInternalServerError
	}

	return http.StatusFound
}

func (store *MemSQL) GetFilterEventNames(projectId int64) ([]model.EventName, int) {
	logFields := log.Fields{
		"project_id": projectId,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db

	var eventNames []model.EventName
	if err := db.Where("project_id = ? AND type = ? AND deleted = 0",
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
func (store *MemSQL) GetSmartEventFilterEventNames(projectID int64, includeDeleted bool) ([]model.EventName, int) {
	logFields := log.Fields{
		"project_id":       projectID,
		"included_deleted": includeDeleted,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db

	whereStmnt := "project_id = ? AND type IN(?)"
	if !includeDeleted {
		whereStmnt = whereStmnt + " AND deleted = 0 "
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
func (store *MemSQL) GetSmartEventFilterEventNameByID(projectID int64, id string, isDeleted bool) (*model.EventName, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"id":         id,
		"is_deleted": isDeleted,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	if id == "" || projectID == 0 {
		return nil, http.StatusBadRequest
	}

	whereStmnt := "project_id = ? AND type IN(?) AND id =? "
	if !isDeleted {
		whereStmnt = whereStmnt + " AND deleted = 0 "
	}

	db := C.GetServices().Db

	var eventName model.EventName
	if err := db.Limit(1).Where(whereStmnt,
		projectID, []string{model.TYPE_CRM_SALESFORCE, model.TYPE_CRM_HUBSPOT}, id).Find(&eventName).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}
		log.WithFields(log.Fields{"project_id": projectID}).WithError(err).Error("Failed getting smart event filter_event_name")

		return nil, http.StatusInternalServerError
	}

	return &eventName, http.StatusFound
}

// GetEventNamesByNames returns list of EventNames objects for given names
func (store *MemSQL) GetEventNamesByNames(projectId int64, names []string) ([]model.EventName, int) {
	logFields := log.Fields{
		"project_id": projectId,
		"names":      names,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
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

	if len(eventNames) < 1 {
		return nil, http.StatusNotFound
	}

	return eventNames, http.StatusFound
}

func (store *MemSQL) GetFilterEventNamesByExprPrefix(projectId int64, prefix string) ([]model.EventName, int) {
	logFields := log.Fields{
		"project_id": projectId,
		"prefix":     prefix,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	var eventNames []model.EventName

	db := C.GetServices().Db
	if err := db.Where("project_id = ? AND type = ? AND filter_expr LIKE ? AND deleted = 0",
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

func (store *MemSQL) UpdateEventName(projectId int64, id string,
	nameType string, eventName *model.EventName) (*model.EventName, int) {
	logFields := log.Fields{
		"project_id": projectId,
		"id":         id,
		"name_type":  nameType,
		"event_name": eventName,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
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

	if err := DeleteEventNameFromCache(projectId, eventName.Name); err != nil {
		log.WithFields(logFields).Error("Failed to invalidate cache on UpdateEventName")
	}

	return &updatedEventName, http.StatusAccepted
}

func (store *MemSQL) updateCRMSmartEventFilter(projectID int64, id string, nameType string,
	eventName *model.EventName, filterExpr *model.SmartCRMEventFilter) (*model.EventName, int) {
	logFields := log.Fields{
		"project_id":  projectID,
		"event_name":  eventName,
		"id":          id,
		"name_type":   nameType,
		"filter_expr": filterExpr,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields) // Validation
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
		prevEventName, status := store.GetSmartEventFilterEventNameByID(projectID, id, eventName.Deleted)
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

	if err := DeleteEventNameFromCache(projectID, eventName.Name); err != nil {
		log.WithFields(logFields).Error("Failed to invalidate cache on updateCRMSmartEventFilter")
	}

	return &updatedEventName, http.StatusAccepted
}

func getCRMSmartEventNameType(source string) string {
	logFields := log.Fields{
		"source": source,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	if source == model.SmartCRMEventSourceSalesforce {
		return model.TYPE_CRM_SALESFORCE
	}

	if source == model.SmartCRMEventSourceHubspot {
		return model.TYPE_CRM_HUBSPOT
	}
	return ""
}

func (store *MemSQL) UpdateCRMSmartEventFilter(projectID int64, id string, eventName *model.EventName,
	filterExpr *model.SmartCRMEventFilter) (*model.EventName, int) {
	logFields := log.Fields{
		"project_id":  projectID,
		"event_name":  eventName,
		"id":          id,
		"filter_expr": filterExpr,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	_, duplicate := store.checkDuplicateSmartEventFilter(projectID, filterExpr)
	if duplicate {
		return nil, http.StatusConflict
	}

	eventName.Type = getCRMSmartEventNameType(filterExpr.Source)

	return store.updateCRMSmartEventFilter(projectID, id, eventName.Type, eventName, filterExpr)
}

// DeleteSmartEventFilter soft delete smart event name with filter expression
func (store *MemSQL) DeleteSmartEventFilter(projectID int64, id string) (*model.EventName, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"id":         id,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	eventName, status := store.GetSmartEventFilterEventNameByID(projectID, id, false)
	if status != http.StatusFound {
		return nil, http.StatusBadRequest
	}

	status = store.DeleteEventName(projectID, eventName.ID, eventName.Type)
	if status != http.StatusAccepted {
		return nil, http.StatusInternalServerError
	}

	return eventName, status
}

func (store *MemSQL) UpdateFilterEventName(projectId int64, id string, eventName *model.EventName) (*model.EventName, int) {
	logFields := log.Fields{
		"project_id": projectId,
		"event_name": eventName,
		"id":         id,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	return store.UpdateEventName(projectId, id, model.TYPE_FILTER_EVENT_NAME, eventName)
}

func (store *MemSQL) DeleteEventName(projectId int64, id string,
	nameType string) int {
	logFields := log.Fields{
		"project_id": projectId,
		"id":         id,
		"name_type":  nameType,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
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

	// Cache invalidation
	eventName, _, err := store.GetEventNameByID(projectId, id)
	if err != nil {
		return http.StatusBadRequest
	}

	if err := DeleteEventNameFromCache(projectId, eventName.Name); err != nil {
		log.WithFields(logFields).Error("Failed to invalidate cache on UpdateEventName")
	}

	return http.StatusAccepted
}

func (store *MemSQL) DeleteFilterEventName(projectId int64, id string) int {
	logFields := log.Fields{
		"project_id": projectId,
		"id":         id,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	return store.DeleteEventName(projectId, id, model.TYPE_FILTER_EVENT_NAME)
}

// Returns sanitized filter expression and valid or not bool.
func getValidatedFilterExpr(filterExpr string) (string, bool) {
	logFields := log.Fields{
		"filter_expr": filterExpr,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
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
	logFields := log.Fields{
		"tokenized_filter": tokenizedFilter,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
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
	logFields := log.Fields{
		"filters": filters,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
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
	logFields := log.Fields{
		"tokenized_event_uri": tokenizedEventURI,
		"filters":             filters,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
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
	logFields := log.Fields{
		"filters":   filters,
		"event_uri": eventURI,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

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
	logFields := log.Fields{
		"event_names": eventNames,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
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
func (store *MemSQL) FilterEventNameByEventURL(projectId int64, eventURL string) (*model.EventName, int) {
	logFields := log.Fields{
		"project_id": projectId,
		"event_url":  eventURL,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
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
	eventNames, errCode := store.GetFilterEventNamesByExprPrefix(projectId,
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

func (store *MemSQL) GetEventNameIDFromEventName(eventName string, projectId int64) (*model.EventName, error) {
	logFields := log.Fields{
		"project_id": projectId,
		"event_name": eventName,
	}

	eventNameModel, errCode := store.GetEventName(eventName, projectId)
	if errCode != http.StatusFound {
		log.WithFields(logFields).Error("Failed to get event_id from event_name")
		return nil, errors.New("Failed to get event_id from event_name")
	}
	return eventNameModel, nil
}

func convert(eventNamesWithAggregation []model.EventNameWithAggregation) []model.EventName {
	logFields := log.Fields{
		"event_names_with_aggregation": eventNamesWithAggregation,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
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

func (store *MemSQL) GetEventTypeFromDb(
	projectID int64, eventNames []string, limit int64) (map[string]string, error) {
	logFields := log.Fields{
		"project_id":  projectID,
		"event_names": eventNames,
		"limit":       limit,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	// var err string
	db := C.GetServices().Db
	hasLimit := limit > 0

	logCtx := log.WithFields(logFields)
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
	queryStr := fmt.Sprintf("SELECT name,type FROM event_names WHERE project_id = %d and name IN (?)", projectID)

	if hasLimit {
		queryStr = queryStr + fmt.Sprintf(" LIMIT %d", limit)
	}

	params := make([]interface{}, 0)
	params = append(params, tmpEventNames)
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

func (store *MemSQL) IsGroupEventName(projectID int64, eventName, eventNameID string) (string, int) {
	logFields := log.Fields{
		"project_id":    projectID,
		"event_name":    eventName,
		"event_name_id": eventNameID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)
	if projectID == 0 {
		logCtx.Error("Invaid project id.")
		return "", http.StatusBadRequest
	}

	if eventName == "" && eventNameID == "" {
		logCtx.Error("Invalid event name or event name id.")
		return "", http.StatusBadRequest
	}

	groupName := U.GetGroupNameFromGroupEventName(eventName)
	if groupName != "" {
		return groupName, http.StatusFound
	}

	if eventNameID == "" {
		return "", http.StatusNotFound
	}

	smartEventName, status := store.GetSmartEventFilterEventNameByID(projectID, eventNameID, false)
	if status != http.StatusFound {
		if status != http.StatusNotFound {
			logCtx.Error("Failed to check for IsGroupEventName")
		}

		return "", status
	}
	groupName, exist := model.IsGroupSmartEventName(projectID, smartEventName)
	if !exist {
		return "", http.StatusNotFound
	}

	return groupName, http.StatusFound
}

func (store *MemSQL) GetEventNameByID(projectID int64, id string) (*model.EventName, int, error) {
	logFields := log.Fields{
		"project_id": projectID,
		"prefix":     id,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	var eventName model.EventName

	db := C.GetServices().Db

	if err := db.Where("project_id = ? AND id = ?", projectID, id).
		Find(&eventName).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound, err
		}
		log.WithFields(logFields).WithError(err).Error("event_name not found")
		return nil, http.StatusInternalServerError, err
	}

	return &eventName, http.StatusFound, nil
}

func SetEventNameCache(projectID int64, eventName string, eventNameModel model.EventName) {
	logCtx := log.WithFields(log.Fields{
		"project_id": projectID,
		"event_name": eventName,
	})

	cacheKey, err := GetEventNameCacheKey(projectID, eventName)
	if err != nil {
		logCtx.WithError(err).Error("Error getting SetEventNameCache key")
		return
	}

	eventNameBytes, err := json.Marshal(eventNameModel)
	if err != nil {
		logCtx.WithError(err).Error("Failed to marshal value on SetEventNameCache")
		return
	}

	const cacheInvalidationDuration = 24 * 60 * 60 // 24 hours in seconds
	err = cacheRedis.Set(cacheKey, string(eventNameBytes), cacheInvalidationDuration)
	if err != nil {
		logCtx.WithError(err).Error("Failed to set cache on SetEventNameCache")
		return
	}
}

func GetEventNameFromCache(projectID int64, eventName string) (model.EventName, error) {
	var eventNameModel model.EventName
	logCtx := log.WithFields(log.Fields{
		"project_id": projectID,
		"event_name": eventName,
	})

	cacheKey, err := GetEventNameCacheKey(projectID, eventName)
	if err != nil {
		logCtx.WithError(err).Error("error getting GetEventNameFromCache key")
		return model.EventName{}, err
	}

	result, err := cacheRedis.Get(cacheKey)
	if err != nil {
		if err == redis.ErrNil {
			return model.EventName{}, err
		}
		logCtx.WithError(err).Error("error getting cache result on GetEventNameFromCache")
		return model.EventName{}, err
	}

	if err := json.Unmarshal([]byte(result), &eventNameModel); err != nil {
		logCtx.WithError(err).Error("error decoding cache on GetEventNameFromCache")
		return model.EventName{}, err
	}

	return eventNameModel, nil
}

func DeleteEventNameFromCache(projectID int64, eventName string) error {
	if cacheKey, err := GetEventNameCacheKey(projectID, eventName); err != nil {
		return err
	} else if err := cacheRedis.Del(cacheKey); err != nil {
		return err
	}
	return nil
}

func GetEventNameCacheKey(projectID int64, eventName string) (*cache.Key, error) {
	return cache.NewKey(projectID, "EN", base64.StdEncoding.EncodeToString([]byte(eventName)))
}
