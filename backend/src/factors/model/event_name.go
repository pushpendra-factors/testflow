package model

import (
	"errors"
	cacheRedis "factors/cache/redis"
	C "factors/config"
	U "factors/util"
	"fmt"
	"net/http"
	"strings"
	"time"

	"encoding/json"

	"github.com/gomodule/redigo/redis"
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

type CacheEventNames struct {
	EventNames []EventName
	Timestamp  int64
}
type FilterInfo struct {
	// filter_expr split by URI_SLASH
	tokenizedFilter []string
	eventName       *EventName
}

const TYPE_USER_CREATED_EVENT_NAME = "UC"
const TYPE_AUTO_TRACKED_EVENT_NAME = "AT"
const TYPE_FILTER_EVENT_NAME = "FE"
const TYPE_INTERNAL_EVENT_NAME = "IE"
const EVENT_NAME_REQUEST_TYPE_APPROX = "approx"
const EVENT_NAME_REQUEST_TYPE_EXACT = "exact"

var ALLOWED_TYPES = [...]string{
	TYPE_USER_CREATED_EVENT_NAME,
	TYPE_AUTO_TRACKED_EVENT_NAME,
	TYPE_FILTER_EVENT_NAME,
	TYPE_INTERNAL_EVENT_NAME,
}

const URI_PROPERTY_PREFIX = ":"
const EVENT_NAMES_LIMIT = 2500

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

func GetOrderedEventNamesFromDb(
	projectId uint64, startTimestamp int64, endTimestamp int64, limit int) ([]EventName, error) {
	db := C.GetServices().Db
	hasLimit := limit > 0
	eventNames := make([]EventName, 0)

	logCtx := log.WithFields(log.Fields{"projectId": projectId,
		"startTimestamp": startTimestamp, "endTimestamp": endTimestamp})

	// Gets occurrence count of event from events table for a
	// limited time window and upto 100k and order by count
	// then join with event names.
	queryStr := "SELECT * FROM (SELECT event_name_id, COUNT(*) FROM" +
		" " + "(SELECT event_name_id FROM events WHERE project_id=? AND timestamp > ?" +
		" " + "AND timestamp <= ? ORDER BY timestamp DESC LIMIT ?) AS sample_events" +
		" " + "GROUP BY event_name_id ORDER BY count DESC) AS event_occurrence" +
		" " + "LEFT JOIN event_names ON event_occurrence.event_name_id=event_names.id "

	if hasLimit {
		queryStr = queryStr + " " + "LIMIT ?"
	}

	const noOfEventsToLoadForOccurrenceSort = 100000

	params := make([]interface{}, 0)
	params = append(params, projectId, startTimestamp,
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
		var eventName EventName
		if err := db.ScanRows(rows, &eventName); err != nil {
			logCtx.WithError(err).Error("Failed scanning rows on get event names ordered by occurrence.")
			return eventNames, err
		}
		eventNames = append(eventNames, eventName)
	}
	return eventNames, nil
}

func getOccurredEventNamesOrderedByOccurrenceWithLimit(projectId uint64, requestType string, limit int) ([]EventName, int64, error) {
	eventNames, timestamp, err := GetCacheEventNamesOrderedByOccurrence(projectId)
	if err == nil || requestType == EVENT_NAME_REQUEST_TYPE_APPROX {
		return eventNames, timestamp, nil
	}

	startTimestamp := U.UnixTimeBeforeAWeek()
	logCtx := log.WithFields(log.Fields{"projectId": projectId, "eventsAfterTimestamp": startTimestamp})
	if err != redis.ErrNil {
		logCtx.WithError(err).Error("Failed to get EventNamesOrderedByOccurrence from cache.")
	}

	endTimestamp := time.Now().UTC().Unix()

	eventNames, err = GetOrderedEventNamesFromDb(
		projectId, startTimestamp, endTimestamp, limit)
	if err != nil {
		return eventNames, 0, err
	}

	timestamp, err = setCacheEventNamesOrderedByOccurrence(projectId, eventNames)
	if err != nil {
		logCtx.WithError(err).Error(
			"Failed to setCacheEventNamesOrderedByOccurrence on getEventNamesOrderedByOccurredWithLimit.")
	}

	return eventNames, timestamp, nil
}

// GetEventNamesOrderedByOccurrenceWithLimit - Returns event names ordered by occurrence
// and back fills event names which haven't occurred ordered by created_at.
// limit = 0, for no limit.
func GetEventNamesOrderedByOccurrenceWithLimit(projectId uint64, requestType string, limit int) ([]EventName, bool, int) {
	eventNames := make([]EventName, 0)
	hasLimit := limit > 0
	// Get event names only occurred on the sample window ordered by occurrence.
	occurredEventNames, timestamp, err := getOccurredEventNamesOrderedByOccurrenceWithLimit(projectId, requestType, limit)
	if err != nil {
		log.WithError(err).Error("Failed to get occured events")
	}

	// Add all event names not occurred in the sample window with the limit.
	addedNamesLookup := make(map[uint64]bool, 0)

	for _, eventName := range occurredEventNames {
		if hasLimit && len(eventNames) == limit {
			break
		}

		eventNames = append(eventNames, eventName)
		addedNamesLookup[eventName.ID] = true
	}
	isToday := U.IsTimestampToday(timestamp)
	// return, if limit reached already.
	if hasLimit && len(eventNames) == limit {
		return eventNames, isToday, http.StatusFound
	}

	allEventNames, errCode := GetEventNames(projectId)
	if errCode == http.StatusInternalServerError {
		return eventNames, false, errCode
	}

	if errCode == http.StatusNotFound {
		return eventNames, isToday, errCode
	}

	// fill event names not on occurred list.
	for _, eventName := range allEventNames {
		if hasLimit && len(eventNames) == limit {
			break
		}

		if _, exists := addedNamesLookup[eventName.ID]; !exists {
			eventNames = append(eventNames, eventName)
		}
	}

	if len(eventNames) == 0 {
		return eventNames, isToday, http.StatusNotFound
	}

	return eventNames, isToday, http.StatusFound
}

func getEventNamesOrderByOccurrenceCacheKey(projectId uint64) (*cacheRedis.Key, error) {
	prefix := "event_names:ordered_by_occurrence"
	return cacheRedis.NewKey(projectId, prefix, "")
}

func setCacheEventNamesOrderedByOccurrence(projectId uint64, eventNames []EventName) (int64, error) {
	var cacheEventNames CacheEventNames
	logCtx := log.WithField("project_id", projectId)
	if projectId == 0 {
		logCtx.Error("Invalid project id on setCacheEventNamesOrderedByOccurrence")
		return 0, errors.New("invalid project id")
	}

	if eventNames == nil || len(eventNames) == 0 {
		return 0, nil
	}

	currentTimeStamp := time.Now().Unix()
	eventNamesKey, err := getEventNamesOrderByOccurrenceCacheKey(projectId)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get EventNamesOrderByOccurrenceCacheKey.")
		return 0, err
	}

	cacheEventNames.EventNames = eventNames
	cacheEventNames.Timestamp = currentTimeStamp
	enEventCache, err := json.Marshal(cacheEventNames)
	if err != nil {
		logCtx.Error("Failed event names json marshal.")
		return 0, err
	}

	err = cacheRedis.Set(eventNamesKey, string(enEventCache), 10*24*60*60)
	if err != nil {
		logCtx.WithError(err).Error("Failed to setCacheEventNamesOrderedByOccurrence.")
		return 0, err
	}
	return currentTimeStamp, nil
}

func GetCacheEventNamesOrderedByOccurrence(projectId uint64) ([]EventName, int64, error) {
	var cacheEventNames CacheEventNames
	if projectId == 0 {
		return []EventName{}, 0, errors.New("invalid project on GetCacheEventNamesOrderedByOccurrence")
	}

	eventNamesKey, err := getEventNamesOrderByOccurrenceCacheKey(projectId)
	if err != nil {
		return []EventName{}, 0, err
	}
	enEventNames, err := cacheRedis.Get(eventNamesKey)
	if err != nil {
		return []EventName{}, 0, err
	}

	err = json.Unmarshal([]byte(enEventNames), &cacheEventNames)
	if err != nil {
		return []EventName{}, 0, err
	}

	if len(cacheEventNames.EventNames) == 0 {
		return []EventName{}, 0, errors.New("Empty cache event names")
	}
	return cacheEventNames.EventNames, cacheEventNames.Timestamp, nil
}

func GetEventNamesOrderedByOccurrence(projectId uint64, requestType string) ([]EventName, bool, int) {
	return GetEventNamesOrderedByOccurrenceWithLimit(projectId, requestType, EVENT_NAMES_LIMIT)
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

func GetFilterEventNamesByExprPrefix(projectId uint64, prefix string) ([]EventName, int) {
	db := C.GetServices().Db

	var eventNames []EventName
	if err := db.Where("project_id = ? AND type = ? AND filter_expr LIKE ? AND deleted = 'false'",
		projectId, TYPE_FILTER_EVENT_NAME, fmt.Sprintf("%s%%", prefix)).Find(&eventNames).Error; err != nil {
		log.WithFields(log.Fields{"projectId": projectId, "prefix": prefix}).WithError(err).Error(
			"Filtering eventName failed on GetFilterEventNamesByExprPrefix")
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
		log.WithFields(log.Fields{"event_url": eventURL}).WithError(err).Error(
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
