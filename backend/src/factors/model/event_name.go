package model

import (
	"bufio"
	"encoding/json"
	"errors"
	cacheRedis "factors/cache/redis"
	C "factors/config"
	U "factors/util"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

type EventName struct {
	// Composite primary key with projectId.
	ID   uint64 `gorm:"primary_key:true;" json:"id"`
	Name string `json:"name"`
	Type string `gorm:"not null;type:varchar(2)" json:"type"`
	// Below are the foreign key constraints added in creation script.
	// project_id -> projects(id)
	ProjectId uint64 `gorm:"primary_key:true;" json:"project_id"`
	// if default is not set as NULL empty string will be installed.
	FilterExpr string    `gorm:"type:varchar(500);default:null" json:"filter_expr"`
	Deleted    bool      `gorm:"not null;default:false" json:"deleted"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type EventNameWithAggregation struct {
	// Composite primary key with projectId.
	ID   uint64 `gorm:"primary_key:true;" json:"id"`
	Name string `json:"name"`
	Type string `gorm:"not null;type:varchar(2)" json:"type"`
	// Below are the foreign key constraints added in creation script.
	// project_id -> projects(id)
	ProjectId uint64 `gorm:"primary_key:true;" json:"project_id"`
	// if default is not set as NULL empty string will be installed.
	FilterExpr string    `gorm:"type:varchar(500);default:null" json:"filter_expr"`
	Deleted    bool      `gorm:"not null;default:false" json:"deleted"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	LastSeen   uint64    `json:"last_seen"`
	Count      int64     `json:"count"`
}

type CacheEventNames struct {
	EventNames []EventName
	Timestamp  int64
}

type FilterInfo struct {
	// filter_expr split by URI_SLASH
	tokenizedFilter []string
	eventName       *EventName
}

type CacheEventNamesWithTimestamp struct {
	EventNames map[string]U.CountTimestampTuple `json:"en"`
}

const TYPE_USER_CREATED_EVENT_NAME = "UC"
const TYPE_AUTO_TRACKED_EVENT_NAME = "AT"
const TYPE_FILTER_EVENT_NAME = "FE"
const TYPE_INTERNAL_EVENT_NAME = "IE"
const TYPE_CRM_SALESFORCE = "CS"
const TYPE_CRM_HUBSPOT = "CH"
const EVENT_NAME_REQUEST_TYPE_APPROX = "approx"
const EVENT_NAME_REQUEST_TYPE_EXACT = "exact"
const EVENT_NAME_TYPE_SMART_EVENT = "SE"

var ALLOWED_TYPES = [...]string{
	TYPE_USER_CREATED_EVENT_NAME,
	TYPE_AUTO_TRACKED_EVENT_NAME,
	TYPE_FILTER_EVENT_NAME,
	TYPE_INTERNAL_EVENT_NAME,
	TYPE_CRM_SALESFORCE,
	TYPE_CRM_HUBSPOT,
}

const URI_PROPERTY_PREFIX = ":"
const PROPERTY_VALUE_ANY = "value_any"
const EVENT_NAMES_LIMIT = 5000

// PropertyState holds string representing state of the property
type PropertyState string

// PropertyState represents the current or prevous state of property
const (
	CurrentState  PropertyState = "curr"
	PreviousState PropertyState = "last"
)

// CRMFilterRule struct for filter rule
type CRMFilterRule struct {
	Operator      string        `json:"op" enums:"EQUAL,NOT EQUAL,GREATER THAN,LESS THAN,CONTAINS,NOT CONTAINS"`
	PropertyState PropertyState `json:"gen" enums:"curr,last"` // previous or current
	Value         interface{}   `json:"value"`                 // value_any or property value
}

// PropertyFilter struct hold name of the property and logical operations on rules provided
type PropertyFilter struct {
	Name      string          `json:"property_name"`
	Rules     []CRMFilterRule `josn:"rules"`
	LogicalOp string          `json:"logical_op" enums:"AND"`
}

// SmartCRMEventFilter struct is base for CRM smart event filter
type SmartCRMEventFilter struct {
	Source                  string           `json:"source" enums:"salesforce,hubspot"`
	ObjectType              string           `json:"object_type" enums:"salesforce[account,contact,lead],hubspot[contact]"`
	Description             string           `json:"description"`
	FilterEvaluationType    string           `json:"property_evaluation_type" enums:"specific,any"` //any change or specific
	Filters                 []PropertyFilter `json:"filters"`
	TimestampReferenceField string           `json:"timestamp_reference_field" enums:"timestamp_in_track, <any property name>"`
	LogicalOp               string           `json:"logical_op" enums:"AND"`
}

// list of comparision operators for CRM filter
const (
	COMPARE_EQUAL        = "EQUAL"
	COMPARE_NOT_EQUAL    = "NOT EQUAL"
	COMPARE_GREATER_THAN = "GREATER THAN"
	COMPARE_LESS_THAN    = "LESS THAN"
	COMPARE_CONTAINS     = "CONTAINS"
	COMPARE_NOT_CONTAINS = "NOT CONTAINS"
)

// list of logical operators for CRM filter
const (
	LOGICAL_OP_OR  = "OR"
	LOGICAL_OP_AND = "AND"
)

// comparisonOp is map of comparision operator  and its function
var comparisonOp = map[string]func(interface{}, interface{}) bool{
	COMPARE_EQUAL: func(rValue, pValue interface{}) bool {
		if rValue == PROPERTY_VALUE_ANY { // should not be blank
			return pValue != ""
		}

		return rValue == pValue
	},
	COMPARE_NOT_EQUAL: func(rValue, pValue interface{}) bool {
		if rValue == PROPERTY_VALUE_ANY { // value not equal to anything
			return pValue == ""
		}

		return rValue != pValue
	},
	COMPARE_GREATER_THAN: func(rValue, pValue interface{}) bool {
		intRValue, _ := U.GetPropertyValueAsFloat64(rValue)
		intpValue, _ := U.GetPropertyValueAsFloat64(pValue)
		return intpValue > intRValue
	},
	COMPARE_LESS_THAN: func(rValue, pValue interface{}) bool {
		intRValue, _ := U.GetPropertyValueAsFloat64(rValue)
		intpValue, _ := U.GetPropertyValueAsFloat64(pValue)
		return intpValue < intRValue
	},
	COMPARE_CONTAINS:     func(rValue, pValue interface{}) bool { return strings.Contains(pValue.(string), rValue.(string)) },
	COMPARE_NOT_CONTAINS: func(rValue, pValue interface{}) bool { return !strings.Contains(pValue.(string), rValue.(string)) },
}

// FilterEvaluationTypeSpecific for specific change in property or any change property
const (
	FilterEvaluationTypeSpecific = "specific"
	FilterEvaluationTypeAny      = "any"
)

// Support source for CRM smart event filter
const (
	SmartCRMEventSourceSalesforce = "salesforce"
	SmartCRMEventSourceHubspot    = "hubspot"
)

// TimestampReferenceTypeTrack is the field to be used for smart event time
const TimestampReferenceTypeTrack = "timestamp_in_track"

// TODO: Make index name a constant and read it
// error constants
const error_DUPLICATE_FILTER_EXPR = "pq: duplicate key value violates unique constraint \"project_filter_expr_unique_idx\""

func isDuplicateFilterExprError(err error) bool {
	return err.Error() == error_DUPLICATE_FILTER_EXPR
}

func CreateOrGetEventName(eventName *EventName) (*EventName, int) {
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

func CreateOrGetUserCreatedEventName(eventName *EventName) (*EventName, int) {
	eventName.Type = TYPE_USER_CREATED_EVENT_NAME
	return CreateOrGetEventName(eventName)
}

func CreateOrGetAutoTrackedEventName(eventName *EventName) (*EventName, int) {
	eventName.Type = TYPE_AUTO_TRACKED_EVENT_NAME
	return CreateOrGetEventName(eventName)
}

func CreateOrGetFilterEventName(eventName *EventName) (*EventName, int) {
	filterExpr, valid := getValidatedFilterExpr(eventName.FilterExpr)
	if !valid {
		return nil, http.StatusBadRequest
	}

	eventName.Type = TYPE_FILTER_EVENT_NAME
	eventName.FilterExpr = filterExpr

	return CreateOrGetEventName(eventName)
}

// CreateOrGetCRMSmartEventFilterEventName creates a new CRM smart event filter. Deleted event_name will be enabled if conflict found
func CreateOrGetCRMSmartEventFilterEventName(projectID uint64, eventName *EventName, filterExpr *SmartCRMEventFilter) (*EventName, int) {
	if !isValidSmartEventFilterExpr(filterExpr) || filterExpr == nil || eventName.Type != "" {
		return nil, http.StatusBadRequest
	}

	enFilterExp, err := json.Marshal(filterExpr)
	if err != nil {
		log.WithError(err).Error("Failed to marshal filterExpr on CreateOrGetCRMSmartEventFilterEventName")
		return nil, http.StatusInternalServerError
	}

	eventName.FilterExpr = string(enFilterExp)

	eventName.Type = getCRMSmartEventNameType(filterExpr.Source)

	_, status := CreateOrGetEventName(eventName)
	if status == http.StatusConflict && eventName.Deleted == true {
		eventName.ProjectId = 0
		_, status = updateCRMSmartEventFilter(projectID, eventName.ID, eventName.Type, eventName, nil)
	}

	return eventName, status
}

func GetSmartEventEventName(eventName *EventName) (*EventName, int) {
	return GetSmartEventEventNameByNameANDType(eventName.ProjectId, eventName.Name, eventName.Type)
}

func GetSmartEventEventNameByNameANDType(projectID uint64, name, typ string) (*EventName, int) {
	if projectID == 0 || name == "" || typ == "" {
		return nil, http.StatusBadRequest
	}

	db := C.GetServices().Db

	var eventNames []EventName
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

func CreateOrGetSessionEventName(projectId uint64) (*EventName, int) {
	return CreateOrGetEventName(&EventName{ProjectId: projectId, Name: U.EVENT_NAME_SESSION,
		Type: TYPE_INTERNAL_EVENT_NAME})
}

func GetSessionEventName(projectId uint64) (*EventName, int) {
	if projectId == 0 {
		return nil, http.StatusBadRequest
	}

	logCtx := log.WithField("project_id", projectId)

	var eventNames []EventName

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

	for _, allowedType := range ALLOWED_TYPES {
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

	if typ == TYPE_INTERNAL_EVENT_NAME {
		return true
	}

	for _, allowedEventName := range U.ALLOWED_INTERNAL_EVENT_NAMES {
		if name == allowedEventName {
			return true
		}
	}

	return !strings.HasPrefix(name, U.NAME_PREFIX)
}

func GetEventName(name string, projectId uint64) (*EventName, int) {
	// Input Validation. (ID is to be auto generated)
	if name == "" || projectId == 0 {
		log.Error("GetEventName Failed. Missing name or projectId")
		return nil, http.StatusBadRequest
	}

	db := C.GetServices().Db

	var eventName EventName
	if err := db.Where(&EventName{Name: name, ProjectId: projectId}).First(&eventName).Error; err != nil {
		log.WithFields(log.Fields{"projectId": projectId, "Name": name}).WithError(err).Error(
			"Getting eventName failed on GetEventName")
		if gorm.IsRecordNotFoundError(err) {
			return nil, http.StatusNotFound
		}
		return nil, http.StatusInternalServerError
	}
	return &eventName, http.StatusFound
}

func GetEventNames(projectId uint64) ([]EventName, int) {
	if projectId == 0 {
		log.Error("GetEventNames Failed. Missing projectId")
		return nil, http.StatusBadRequest
	}

	db := C.GetServices().Db

	var eventNames []EventName
	if err := db.Order("created_at ASC").Where("project_id = ?", projectId).Limit(2000).Find(&eventNames).Error; err != nil {
		return nil, http.StatusInternalServerError
	}
	if len(eventNames) == 0 {
		return eventNames, http.StatusNotFound
	}
	return eventNames, http.StatusFound
}

// GetOrderedEventNamesFromDb - Get 'limit' events from DB sort by occurence for a given time period
func GetOrderedEventNamesFromDb(
	projectID uint64, startTimestamp int64, endTimestamp int64, limit int) ([]EventNameWithAggregation, error) {
	db := C.GetServices().Db
	hasLimit := limit > 0
	eventNames := make([]EventNameWithAggregation, 0)

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
		var eventName EventNameWithAggregation
		if err := db.ScanRows(rows, &eventName); err != nil {
			logCtx.WithError(err).Error("Failed scanning rows on get event names ordered by occurrence.")
			return eventNames, err
		}
		eventNames = append(eventNames, eventName)
	}
	return eventNames, nil
}

// Today's keys
func GetPropertiesByEventCategoryCacheKey(projectId uint64, event_name string, property string, category string, date string) (*cacheRedis.Key, error) {
	prefix := "EN:PC"
	return cacheRedis.NewKey(projectId, fmt.Sprintf("%s:%s", prefix, event_name), fmt.Sprintf("%s:%s:%s", date, category, property))
}
func GetEventNamesOrderByOccurrenceAndRecencyCacheKey(projectId uint64, event_name string, date string) (*cacheRedis.Key, error) {
	prefix := "EN"
	return cacheRedis.NewKey(projectId, prefix, fmt.Sprintf("%s:%s", date, event_name))
}

func GetSmartEventNamesOrderByOccurrenceAndRecencyCacheKey(projectId uint64, event_name string, date string) (*cacheRedis.Key, error) {
	prefix := "EN:SE"
	return cacheRedis.NewKey(projectId, prefix, fmt.Sprintf("%s:%s", date, event_name))
}

func GetValuesByEventPropertyCacheKey(projectId uint64, event_name string, property_name string, value string, date string) (*cacheRedis.Key, error) {
	prefix := "EN:PV"
	return cacheRedis.NewKey(projectId, fmt.Sprintf("%s:%s:%s", prefix, event_name, property_name), fmt.Sprintf("%s:%s", date, value))
}

// Rollup keys
func GetPropertiesByEventCategoryRollUpCacheKey(projectId uint64, event_name string, date string) (*cacheRedis.Key, error) {
	prefix := "RollUp:EN:PC"
	return cacheRedis.NewKey(projectId, fmt.Sprintf("%s:%s", prefix, event_name), date)
}
func GetEventNamesOrderByOccurrenceAndRecencyRollUpCacheKey(projectId uint64, date string) (*cacheRedis.Key, error) {
	prefix := "RollUp:EN"
	return cacheRedis.NewKey(projectId, prefix, date)
}

func GetValuesByEventPropertyRollUpCacheKey(projectId uint64, event_name string, property_name string, date string) (*cacheRedis.Key, error) {
	prefix := "RollUp:EN:PV"
	return cacheRedis.NewKey(projectId, fmt.Sprintf("%s:%s:%s", prefix, event_name, property_name), date)
}

// Today's keys count per project used for clean up
func GetPropertiesByEventCategoryCountCacheKey(projectId uint64, dateKey string) (*cacheRedis.Key, error) {
	prefix := "C:EN:PC"
	return cacheRedis.NewKeyWithAllProjectsSupport(projectId, prefix, dateKey)
}
func GetEventNamesOrderByOccurrenceAndRecencyCountCacheKey(projectId uint64, dateKey string) (*cacheRedis.Key, error) {
	prefix := "C:EN"
	return cacheRedis.NewKeyWithAllProjectsSupport(projectId, prefix, dateKey)
}

func GetValuesByEventPropertyCountCacheKey(projectId uint64, dateKey string) (*cacheRedis.Key, error) {
	prefix := "C:EN:PV"
	return cacheRedis.NewKeyWithAllProjectsSupport(projectId, prefix, dateKey)

}

//GetPropertyValuesByEventProperty This method iterates for last n days to get all the top 'limit' property values for the given property/event
// Picks all last 24 hours values and sorts the remaining by occurence and returns top 'limit' values
func GetPropertyValuesByEventProperty(projectID uint64, eventName string, propertyName string, limit int, lastNDays int) ([]string, error) {
	if projectID == 0 {
		return []string{}, errors.New("invalid project on GetPropertyValuesByEventProperty")
	}

	if eventName == "" {
		return []string{}, errors.New("invalid event_name on GetPropertyValuesByEventProperty")
	}

	if propertyName == "" {
		return []string{}, errors.New("invalid property_name on GetPropertyValuesByEventProperty")
	}
	currentDate := OverrideCacheDateRangeForProjects(projectID)
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

	eventPropertyValuesKey, err := GetValuesByEventPropertyRollUpCacheKey(projectID, eventName, propertyName, dateKey)
	if err != nil {
		return U.CachePropertyValueWithTimestamp{}, err
	}
	values, _, err := cacheRedis.GetIfExistsPersistent(eventPropertyValuesKey)
	if values == "" {
		eventNameWithSlash := fmt.Sprintf("%s/", eventName)
		eventPropertyValuesKeyWithSlash, err := GetValuesByEventPropertyRollUpCacheKey(projectID, eventNameWithSlash, propertyName, dateKey)
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

//GetPropertiesByEvent This method iterates for last n days to get all the top 'limit' properties for the given event
// Picks all last 24 hours properties and sorts the remaining by occurence and returns top 'limit' properties
func GetPropertiesByEvent(projectID uint64, eventName string, limit int, lastNDays int) (map[string][]string, error) {
	properties := make(map[string][]string)
	if projectID == 0 {
		return properties, errors.New("invalid project on GetPropertiesByEvent")
	}
	currentDate := OverrideCacheDateRangeForProjects(projectID)
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

	for _, v := range eventPropertiesSorted {
		if properties[v.Category] == nil {
			properties[v.Category] = make([]string, 0)
		}
		properties[v.Category] = append(properties[v.Category], v.Name)
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

	eventPropertiesKey, err := GetPropertiesByEventCategoryRollUpCacheKey(projectID, eventName, dateKey)
	if err != nil {
		return U.CachePropertyWithTimestamp{}, err
	}
	eventProperties, _, err := cacheRedis.GetIfExistsPersistent(eventPropertiesKey)
	if eventProperties == "" {
		eventNameWithSlash := fmt.Sprintf("%s/", eventName)
		eventPropertiesKeyWithSlash, err := GetPropertiesByEventCategoryRollUpCacheKey(projectID, eventNameWithSlash, dateKey)
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

func extractCategoryProperty(categoryProperty string) (string, string, string) {
	catPr := strings.SplitN(categoryProperty, ":", 3)
	return catPr[0], catPr[1], catPr[2]
}

func aggregateEventsAcrossDate(events []CacheEventNamesWithTimestamp) []U.NameCountTimestampCategory {
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

			if eventDetails.Type == EVENT_NAME_TYPE_SMART_EVENT {
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

//GetEventNamesOrderedByOccurenceAndRecency This method iterates for last n days to get all the top 'limit' events for the given project
// Picks all last 24 hours events and sorts the remaining by occurence and returns top 'limit' events
func GetEventNamesOrderedByOccurenceAndRecency(projectID uint64, limit int, lastNDays int) (map[string][]string, error) {
	if projectID == 0 {
		return nil, errors.New("invalid project on get event names ordered by occurence and recency")
	}
	currentDate := OverrideCacheDateRangeForProjects(projectID)
	events := make([]CacheEventNamesWithTimestamp, 0)
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

	if limit > 0 {
		sliceLength := len(eventsSorted)
		if sliceLength > limit {
			eventsSorted = eventsSorted[0:limit]
		}
	}

	eventStringWithGroups := make(map[string][]string)
	for _, event := range eventsSorted {
		eventStringWithGroups[event.GroupName] = append(eventStringWithGroups[event.GroupName], event.Name)
	}

	return eventStringWithGroups, nil
}

func getEventNamesOrderedByOccurenceAndRecencyFromCache(projectID uint64, dateKey string) (CacheEventNamesWithTimestamp, error) {
	logCtx := log.WithFields(log.Fields{
		"project_id": projectID,
	})
	if projectID == 0 {
		return CacheEventNamesWithTimestamp{}, errors.New("invalid project on get event names ordered by occurence and recency from cache")
	}

	eventNamesKey, err := GetEventNamesOrderByOccurrenceAndRecencyRollUpCacheKey(projectID, dateKey)
	if err != nil {
		return CacheEventNamesWithTimestamp{}, err
	}
	eventNames, _, err := cacheRedis.GetIfExistsPersistent(eventNamesKey)
	if eventNames == "" {
		logCtx.WithField("date_key", dateKey).Info("MISSING ROLLUP EN")
		return CacheEventNamesWithTimestamp{}, nil
	}
	var cacheEventNames CacheEventNamesWithTimestamp
	err = json.Unmarshal([]byte(eventNames), &cacheEventNames)
	if err != nil {
		return CacheEventNamesWithTimestamp{}, err
	}
	return cacheEventNames, nil
}

func isCachePrefixTypeSmartEvent(prefix string) bool {
	prefixes := strings.SplitN(prefix, ":", 2)
	if len(prefixes) == 2 && prefixes[1] == EVENT_NAME_TYPE_SMART_EVENT {
		return true
	}
	return false
}

func GetCacheEventObject(events []*cacheRedis.Key, eventCounts []string) CacheEventNamesWithTimestamp {
	eventNames := make(map[string]U.CountTimestampTuple)
	for index, eventCount := range eventCounts {
		key, value := ExtractKeyDateCountFromCacheKey(eventCount, events[index].Suffix)
		if isCachePrefixTypeSmartEvent(events[index].Prefix) {
			value.Type = EVENT_NAME_TYPE_SMART_EVENT
		}

		eventNames[key] = value
	}
	cacheEventNames := CacheEventNamesWithTimestamp{
		EventNames: eventNames}
	return cacheEventNames
}

func GetCachePropertyValueObject(values []*cacheRedis.Key, valueCounts []string) U.CachePropertyValueWithTimestamp {
	propertyValues := make(map[string]U.CountTimestampTuple)
	for index, valuesCount := range valueCounts {
		key, value := ExtractKeyDateCountFromCacheKey(valuesCount, values[index].Suffix)
		propertyValues[key] = value
	}
	cachePropertyValues := U.CachePropertyValueWithTimestamp{
		PropertyValue: propertyValues}
	return cachePropertyValues
}

func GetCachePropertyObject(properties []*cacheRedis.Key, propertyCounts []string) U.CachePropertyWithTimestamp {
	var dateKeyInTime time.Time
	eventProperties := make(map[string]U.PropertyWithTimestamp)
	propertyCategory := make(map[string]map[string]int64)
	for index, propertiesCount := range propertyCounts {
		dateKey, cat, pr := extractCategoryProperty(properties[index].Suffix)
		dateKeyInTime, _ = time.Parse(U.DATETIME_FORMAT_YYYYMMDD, dateKey)
		if propertyCategory[pr] == nil {
			propertyCategory[pr] = make(map[string]int64)
		}
		catCount, _ := strconv.Atoi(propertiesCount)
		propertyCategory[pr][cat] = int64(catCount)
	}
	for pr, catCount := range propertyCategory {
		cwc := make(map[string]int64)
		totalCount := int64(0)
		for cat, catCount := range catCount {
			cwc[cat] = catCount
			totalCount += catCount
		}
		prWithTs := U.PropertyWithTimestamp{CategorywiseCount: cwc,
			CountTime: U.CountTimestampTuple{Count: totalCount, LastSeenTimestamp: dateKeyInTime.Unix()}}
		eventProperties[pr] = prWithTs
	}
	cacheProperties := U.CachePropertyWithTimestamp{
		Property: eventProperties}
	return cacheProperties
}
func ExtractKeyDateCountFromCacheKey(keyCount string, cacheKey string) (string, U.CountTimestampTuple) {
	dateKey := strings.SplitN(cacheKey, ":", 2)

	keyDate, _ := time.Parse(U.DATETIME_FORMAT_YYYYMMDD, dateKey[0])
	KeyCountNum, _ := strconv.Atoi(keyCount)
	return dateKey[1], U.CountTimestampTuple{
		LastSeenTimestamp: keyDate.Unix(),
		Count:             int64(KeyCountNum),
	}
}

func GetFilterEventNames(projectId uint64) ([]EventName, int) {
	db := C.GetServices().Db

	var eventNames []EventName
	if err := db.Where("project_id = ? AND type = ? AND deleted = 'false'",
		projectId, TYPE_FILTER_EVENT_NAME).Find(&eventNames).Error; err != nil {
		log.WithFields(log.Fields{"project_id": projectId}).WithError(err).Error("Failed getting filter_event_names")

		return nil, http.StatusInternalServerError
	}

	if len(eventNames) == 0 {
		return eventNames, http.StatusNotFound
	}

	return eventNames, http.StatusFound
}

// GetSmartEventFilterEventNames returns a list of all smart events
func GetSmartEventFilterEventNames(projectID uint64) ([]EventName, int) {
	db := C.GetServices().Db

	var eventNames []EventName
	if err := db.Where("project_id = ? AND type IN(?) AND deleted = 'false'",
		projectID, []string{TYPE_CRM_SALESFORCE, TYPE_CRM_HUBSPOT}).Find(&eventNames).Error; err != nil {
		log.WithFields(log.Fields{"project_id": projectID}).WithError(err).Error("Failed getting filter_event_names")

		return nil, http.StatusInternalServerError
	}

	if len(eventNames) == 0 {
		return eventNames, http.StatusNotFound
	}

	return eventNames, http.StatusFound
}

// GetSmartEventFilterEventNameByID returns the smart event by event_name id
func GetSmartEventFilterEventNameByID(projectID, id uint64) (*EventName, int) {
	if id == 0 || projectID == 0 {
		return nil, http.StatusBadRequest
	}

	db := C.GetServices().Db

	var eventName EventName
	if err := db.Where("project_id = ? AND type IN(?) AND deleted = 'false' AND id = ?",
		projectID, []string{TYPE_CRM_SALESFORCE, TYPE_CRM_HUBSPOT}, id).First(&eventName).Error; err != nil {
		log.WithFields(log.Fields{"project_id": projectID}).WithError(err).Error("Failed getting smart event filter_event_name")

		return nil, http.StatusInternalServerError
	}

	return &eventName, http.StatusFound
}

// GetDecodedSmartEventFilterExp unmarhsal encoded CRM smart event filter exp to SmartCRMEventFilter struct
func GetDecodedSmartEventFilterExp(enFilterExp string) (*SmartCRMEventFilter, error) {
	if enFilterExp == "" {
		return nil, errors.New("empty string")
	}

	var smartCRMEventFilter SmartCRMEventFilter
	err := json.Unmarshal([]byte(enFilterExp), &smartCRMEventFilter)
	if err != nil {
		return nil, err
	}

	return &smartCRMEventFilter, nil
}

// isRuleApplicable compare property based on rule provided
func isRuleApplicable(properties *map[string]interface{}, propertyName string, rule *CRMFilterRule) bool {
	if propertyValue, exists := (*properties)[propertyName]; exists {
		if comparisonOp[rule.Operator](rule.Value, propertyValue) {
			return true
		}
	} else {
		if comparisonOp[rule.Operator](rule.Value, "") {
			return true
		}
	}

	return false
}

// CRMSmartEvent holds payload for creating smart event
type CRMSmartEvent struct {
	Name       string
	Properties map[string]interface{}
	Timestamp  uint64
}

func getPrevPropertyName(name string) string {
	if name == "" {
		return ""
	}
	return fmt.Sprintf("$prev_%s", strings.TrimPrefix(name, U.NAME_PREFIX))
}

func getCurrPropertyName(name string) string {
	if name == "" {
		return ""
	}
	return fmt.Sprintf("$curr_%s", strings.TrimPrefix(name, U.NAME_PREFIX))
}

// FillSmartEventCRMProperties fills all properties from CRM smart filter to new properties
func FillSmartEventCRMProperties(newProperties, current, prev *map[string]interface{}, filter *SmartCRMEventFilter) {
	if *newProperties == nil {
		*newProperties = make(map[string]interface{})
	}

	for i := range filter.Filters {
		if value, exists := (*current)[filter.Filters[i].Name]; exists {
			(*newProperties)[getCurrPropertyName(filter.Filters[i].Name)] = value
		}
		if value, exists := (*prev)[filter.Filters[i].Name]; exists {
			(*newProperties)[getPrevPropertyName(filter.Filters[i].Name)] = value
		}
	}
}

func validateMatch(anyCurrMatch, anyPrevMatch bool, compareMode string, ruleSkipable bool) bool {
	switch compareMode {
	case CompareStateBoth:
		return (anyCurrMatch && anyPrevMatch) || (ruleSkipable && (anyCurrMatch || anyPrevMatch))
	case CompareStateCurr:
		return anyCurrMatch
	case CompareStatePrev:
		return anyPrevMatch
	default:
		return false
	}
}

// compare modes for validating properties
const (
	CompareStateCurr = "curr"
	CompareStatePrev = "prev"
	CompareStateBoth = "both"
)

func IsEventNameTypeSmartEvent(eventType string) bool {
	return eventType == TYPE_CRM_HUBSPOT || eventType == TYPE_CRM_SALESFORCE
}

// CRMFilterEvaluator evaluates a CRM filter on the properties provided. Can work in current properties or current and previous property mode
func CRMFilterEvaluator(projectID uint64, currProperty, prevProperty *map[string]interface{}, filter *SmartCRMEventFilter, compareState string) bool {
	if filter == nil {
		return false
	}

	if compareState == "" ||
		(compareState == CompareStateCurr && currProperty == nil) ||
		(compareState == CompareStatePrev && prevProperty == nil) ||
		(compareState == CompareStateBoth && (currProperty == nil || prevProperty == nil)) {
		return false
	}

	filterSkipable := filter.LogicalOp == LOGICAL_OP_OR

	anyfilterTrue := false
	for _, filterProperty := range filter.Filters { // a successfull completion of this loop implies a vaild AND or failed OR operation
		ruleSkipable := filterProperty.LogicalOp == LOGICAL_OP_OR
		var anyPrevMatch bool
		var anyCurrMatch bool

		// avoid same value in previous and current properties
		if compareState == CompareStateBoth {
			diffPropertyValue := (*currProperty)[filterProperty.Name] != (*prevProperty)[filterProperty.Name]
			if !diffPropertyValue {
				if !filterSkipable {
					return false
				}
				continue
			}

			if filter.FilterEvaluationType == FilterEvaluationTypeAny {
				if diffPropertyValue {
					anyfilterTrue = true
				} else {
					if !filterSkipable {
						return false
					}
				}
				continue
			}
		}

		// cannot compare if only one is provided, return true and switch to both mode
		if (compareState == CompareStateCurr || compareState == CompareStatePrev) && filter.FilterEvaluationType == FilterEvaluationTypeAny {
			return true
		}

		for _, rule := range filterProperty.Rules { // a successfull completion of this loop implies a vaild AND or failed OR operation

			if (compareState == CompareStateCurr || compareState == CompareStateBoth) && rule.PropertyState == CurrentState {
				if !isRuleApplicable(currProperty, filterProperty.Name, &rule) {
					if !ruleSkipable && !filterSkipable {
						return false
					}
				} else {
					anyCurrMatch = true
				}
			}

			if (compareState == CompareStatePrev || compareState == CompareStateBoth) && rule.PropertyState == PreviousState {
				if !isRuleApplicable(prevProperty, filterProperty.Name, &rule) {
					if !ruleSkipable && !filterSkipable {
						return false
					}
				} else {
					anyPrevMatch = true
				}
			}
		}

		if !filterSkipable {

			// is it an OR operation ? either previous or current should have a match
			if !validateMatch(anyCurrMatch, anyPrevMatch, compareState, ruleSkipable) {
				return false
			}

		} else if validateMatch(anyCurrMatch, anyPrevMatch, compareState, ruleSkipable) {
			return true
		}
	}

	if !filterSkipable {
		return true
	} else if anyfilterTrue {
		return true
	}

	return false
}

// returns list of EventNames objects for given names
func GetEventNamesByNames(projectId uint64, names []string) ([]EventName, int) {

	db := C.GetServices().Db
	var eventNames []EventName
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

func GetFilterEventNamesByExprPrefix(projectId uint64, prefix string) ([]EventName, int) {
	db := C.GetServices().Db

	var eventNames []EventName
	if err := db.Where("project_id = ? AND type = ? AND filter_expr LIKE ? AND deleted = 'false'",
		projectId, TYPE_FILTER_EVENT_NAME, fmt.Sprintf("%s%%", prefix)).Find(&eventNames).Error; err != nil {
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

func UpdateEventName(projectId uint64, id uint64,
	nameType string, eventName *EventName) (*EventName, int) {
	db := C.GetServices().Db

	// update not allowed for internal event names.
	if nameType == TYPE_INTERNAL_EVENT_NAME {
		return nil, http.StatusBadRequest
	}

	// Validation
	if projectId == 0 || eventName.ProjectId != 0 || !isValidType(nameType) ||
		!isValidName(eventName.Name, eventName.Type) {

		return nil, http.StatusBadRequest
	}

	var updatedEventName EventName
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

func updateCRMSmartEventFilter(projectID uint64, id uint64, nameType string, eventName *EventName, filterExpr *SmartCRMEventFilter) (*EventName, int) {

	// Validation
	if id == 0 || projectID == 0 || eventName.ProjectId != 0 ||
		!isValidName(eventName.Name, eventName.Type) {

		return nil, http.StatusBadRequest
	}

	// update not allowed for non CRM based smart event.
	if nameType != TYPE_CRM_SALESFORCE && nameType != TYPE_CRM_HUBSPOT {
		return nil, http.StatusBadRequest
	}

	if eventName.FilterExpr != "" {
		return nil, http.StatusBadRequest
	}

	if filterExpr != nil && !isValidSmartEventFilterExpr(filterExpr) {
		return nil, http.StatusBadRequest
	}

	db := C.GetServices().Db
	updateFields := map[string]interface{}{}
	updateFields["name"] = eventName.Name

	if eventName.Deleted == true {
		updateFields["deleted"] = false
	}

	if filterExpr != nil {
		prevEventName, status := GetSmartEventFilterEventNameByID(projectID, id)
		if status != http.StatusFound {
			return nil, http.StatusBadRequest
		}

		var existingFilterExp *SmartCRMEventFilter
		var err error

		existingFilterExp, err = GetDecodedSmartEventFilterExp(prevEventName.FilterExpr)
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

	var updatedEventName EventName
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
	if source == SmartCRMEventSourceSalesforce {
		return TYPE_CRM_SALESFORCE
	}

	if source == SmartCRMEventSourceHubspot {
		return TYPE_CRM_HUBSPOT
	}
	return ""
}

func UpdateCRMSmartEventFilter(projectID uint64, id uint64, eventName *EventName, filterExpr *SmartCRMEventFilter) (*EventName, int) {
	eventName.Type = getCRMSmartEventNameType(filterExpr.Source)

	return updateCRMSmartEventFilter(projectID, id, eventName.Type, eventName, filterExpr)
}

func UpdateFilterEventName(projectId uint64, id uint64, eventName *EventName) (*EventName, int) {
	return UpdateEventName(projectId, id, TYPE_FILTER_EVENT_NAME, eventName)
}

func DeleteEventName(projectId uint64, id uint64,
	nameType string) int {
	db := C.GetServices().Db

	// Validation
	if projectId == 0 {
		return http.StatusBadRequest
	}

	var updatedEventName EventName
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

func DeleteFilterEventName(projectId uint64, id uint64) int {
	return DeleteEventName(projectId, id, TYPE_FILTER_EVENT_NAME)
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

func isValidSmartCRMFilterObjectType(smartCRMFilter *SmartCRMEventFilter) bool {
	if smartCRMFilter.Source == SmartCRMEventSourceSalesforce {
		typeInt := GetSalesforceDocTypeByAlias(smartCRMFilter.ObjectType)
		if typeInt != 0 {
			return true
		}
	}

	if smartCRMFilter.Source == SmartCRMEventSourceHubspot {
		if smartCRMFilter.ObjectType == HubspotDocumentTypeNameContact || smartCRMFilter.ObjectType == HubspotDocumentTypeNameDeal {
			return true
		}
	}

	return false
}

func isValidSmartCRMFilterOperator(operator string) bool {
	if _, exists := comparisonOp[operator]; !exists {
		return false
	}
	return true
}

func isValidSmartCRMFilterLogicalOp(logicalOp string) bool {
	if logicalOp != LOGICAL_OP_AND && logicalOp != LOGICAL_OP_OR {
		return false
	}
	return true
}

// Validates smart event filter expression
func isValidSmartEventFilterExpr(smartCRMFilter *SmartCRMEventFilter) bool {
	if smartCRMFilter == nil {
		return false
	}

	if smartCRMFilter.TimestampReferenceField == "" || smartCRMFilter.FilterEvaluationType != FilterEvaluationTypeSpecific && smartCRMFilter.FilterEvaluationType != FilterEvaluationTypeAny {
		return false
	}

	if !isValidSmartCRMFilterObjectType(smartCRMFilter) {
		return false
	}

	if len(smartCRMFilter.Filters) < 1 {
		return false
	}

	for i := range smartCRMFilter.Filters {
		if smartCRMFilter.Filters[i].Name == "" {
			return false
		}

		if smartCRMFilter.FilterEvaluationType == FilterEvaluationTypeAny {
			if len(smartCRMFilter.Filters[i].Rules) > 0 { // for any change, rules not required
				return false
			}
			continue
		}

		if !isValidSmartCRMFilterLogicalOp(smartCRMFilter.Filters[i].LogicalOp) {
			return false
		}

		if len(smartCRMFilter.Filters[i].Rules) < 2 { // avoid single rule filter, require prev and curr property rule
			return false
		}

		var anyCurr bool
		var anyPrev bool
		for _, rule := range smartCRMFilter.Filters[i].Rules {
			if !isValidSmartCRMFilterOperator(rule.Operator) {
				return false
			}

			if rule.PropertyState == CurrentState {
				anyCurr = true
			}

			if rule.PropertyState == PreviousState {
				anyPrev = true
			}

			if rule.Value == "" {
				return false
			}

			if rule.Value == PROPERTY_VALUE_ANY && rule.Operator != COMPARE_EQUAL && rule.Operator != COMPARE_NOT_EQUAL {
				return false
			}
		}

		if anyCurr == false || anyPrev == false {
			return false
		}
	}

	return true
}

// IsFilterMatch checks for exact match of filter and uri passed.
// skips uri_token, if filter_token prefixed with semicolon (URI_PROPERTY_PREFIX).
func IsFilterMatch(tokenizedFilter []string, tokenizedMatchURI []string) bool {
	if len(tokenizedMatchURI) != len(tokenizedFilter) {
		return false
	}

	lastIndexTF := len(tokenizedFilter) - 1
	for i, ftoken := range tokenizedFilter {
		if !strings.HasPrefix(ftoken, URI_PROPERTY_PREFIX) {
			// filter_token is not property, should be == uri_token.
			if ftoken != tokenizedMatchURI[i] {
				return false
			}
		} else {
			// last index of filter_token as property with uri_token as "". edge case.
			if i == lastIndexTF && tokenizedMatchURI[0] == "" {
				return false
			}
		}
	}

	return true
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
		if strings.HasPrefix(token, URI_PROPERTY_PREFIX) {
			score = score - 1
		} else {
			score = score + 1
		}
	}

	return score
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

// AddSmartEventReferenceMeta adds reference_id and meta for debuging purpose
func AddSmartEventReferenceMeta(properties *map[string]interface{}, eventID string) {
	if eventID != "" {
		(*properties)[U.EP_CRM_REFERENCE_EVENT_ID] = eventID
	}
}

// matchEventNameWithFilters match uri by passing through filters.
func matchEventURIWithFilters(filters *[]FilterInfo,
	tokenizedEventURI []string) (*FilterInfo, bool) {
	if len(tokenizedEventURI) == 0 {
		return nil, false
	}

	matches := make([]*FilterInfo, 0, 0)
	for i, finfo := range *filters {
		if IsFilterMatch(finfo.tokenizedFilter, tokenizedEventURI) {
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

func makeFilterInfos(eventNames []EventName) (*[]FilterInfo, error) {
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
func FilterEventNameByEventURL(projectId uint64, eventURL string) (*EventName, int) {
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
	eventNames, errCode := GetFilterEventNamesByExprPrefix(projectId,
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

// FillEventPropertiesByFilterExpr - Parses and fills event properties
// from tokenized_event_uri using tokenized_filter_expr.
func FillEventPropertiesByFilterExpr(eventProperties *U.PropertiesMap,
	filterExpr string, eventURL string) error {

	parsedEventURL, err := U.ParseURLStable(eventURL)
	if err != nil {
		return err
	}
	tokenizedEventURI := U.TokenizeURI(U.GetURLPathWithHash(parsedEventURL))

	parsedFilterExpr, err := U.ParseURLWithoutProtocol(filterExpr)
	if err != nil {
		return err
	}
	tokenizedFilterExpr := U.TokenizeURI(U.GetURLPathWithHash(parsedFilterExpr))

	for pos := 0; pos < len(tokenizedFilterExpr); pos++ {
		if strings.HasPrefix(tokenizedFilterExpr[pos], URI_PROPERTY_PREFIX) {
			// Adding semicolon removed filter_expr_token as key and event_uri_token as value.
			(*eventProperties)[strings.TrimPrefix(tokenizedFilterExpr[pos],
				URI_PROPERTY_PREFIX)] = tokenizedEventURI[pos]
		}
	}

	return nil
}

func GetEventNameFromEventNameId(eventNameId uint64, projectId uint64) (*EventName, error) {
	db := C.GetServices().Db
	var eventName EventName
	queryStr := "SELECT * FROM event_names WHERE id = ? AND project_id = ?"
	err := db.Raw(queryStr, eventNameId, projectId).Scan(&eventName).Error
	if err != nil {
		log.Error("Failed to get event_name from event_name_id")
		return nil, err
	}
	return &eventName, nil
}

func convert(eventNamesWithAggregation []EventNameWithAggregation) []EventName {
	eventNames := make([]EventName, 0)
	for _, event := range eventNamesWithAggregation {
		eventNames = append(eventNames, EventName{
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

// GetEventNamesFromFile read unique eventNames from Event file
func GetEventNamesFromFile(scanner *bufio.Scanner, projectId uint64) ([]string, error) {
	logCtx := log.WithField("project_id", projectId)
	scanner.Split(bufio.ScanLines)
	var txtline string
	eventNames := make([]string, 0)
	var dat map[string]interface{}
	s := map[string]bool{}

	for scanner.Scan() {
		txtline = scanner.Text()
		if err := json.Unmarshal([]byte(txtline), &dat); err != nil {
			logCtx.Error("Unable to decode line")
		}
		eventNameString := dat["en"].(string)
		_, ok := s[eventNameString]
		if ok != true {
			eventNames = append(eventNames, eventNameString)
			s[eventNameString] = true
		}

	}
	err := scanner.Err()
	logCtx.Info("Extraced Unique EventNames from file")

	if err != nil {
		return []string{}, err
	}

	return eventNames, nil

}

func GetEventTypeFromDb(
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
