package model

import (
	"encoding/json"
	cacheRedis "factors/cache/redis"
	U "factors/util"
	"fmt"
	"net/http"
	"reflect"

	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

const (
	QueryClassEvents      = "events"
	QueryClassInsights    = "insights"
	QueryClassFunnel      = "funnel"
	QueryClassChannel     = "channel"
	QueryClassChannelV1   = "channel_v1"
	QueryClassAttribution = "attribution"
	QueryClassWeb         = "web"

	PresentationLine   = "pl"
	PresentationBar    = "pb"
	PresentationTable  = "pt"
	PresentationCard   = "pc"
	PresentationFunnel = "pf"
)

const (
	GroupByTimestampHour  = "hour"
	GroupByTimestampDate  = "date"
	GroupByTimestampWeek  = "week"
	GroupByTimestampMonth = "month"
)

const (
	EventCondAnyGivenEvent  = "any_given_event"
	EventCondAllGivenEvent  = "all_given_event"
	EventCondEachGivenEvent = "each_given_event"
)

const (
	PropertyEntityUser  = "user"
	PropertyEntityEvent = "event"
)

const PropertyValueNone = "$none"

const (
	ErrUnsupportedGroupByEventPropertyOnUserQuery = "group by event property is not supported for user query"
	ErrMsgQueryProcessingFailure                  = "Failed processing query"
	ErrMsgMaxFunnelStepsExceeded                  = "Max funnel steps exceeded"
)

const (
	SelectDefaultEventFilter        = "events.id as event_id, events.user_id as event_user_id"
	SelectDefaultUserFilter         = "events.user_id as event_user_id"
	SelectDefaultEventFilterByAlias = "event_id, event_user_id, event_name"
	SelectDefaultUserFilterByAlias  = "coal_user_id, event_user_id, event_name"
)

const (
	GroupKeyPrefix  = "_group_key_"
	AliasEventName  = "event_name"
	AliasEventIndex = "event_index"
	AliasDateTime   = "datetime"
	AliasAggr       = "count"
	AliasError      = "error"
)

const DefaultTimezone = "UTC"

const (
	ResultsLimit    = 10000
	MaxResultsLimit = 100000
)

const (
	StepPrefix             = "step_"
	FunnelConversionPrefix = "conversion_"
)

const (
	NumericalGroupByBuckets       = 10
	NumericalGroupBySeparator     = " - "
	NumericalLowerBoundPercentile = 0.02
	NumericalUpperBoundPercentile = 0.98
)

const (
	GroupByTypeWithBuckets = "with_buckets"
	GroupByTypeRawValues   = "raw_values"
)

const (
	EqualsOpStr             = "equals"
	EqualsOp                = "="
	NotEqualOpStr           = "notEqual"
	NotEqualOp              = "!="
	GreaterThanOpStr        = "greaterThan"
	LesserThanOpStr         = "lesserThan"
	GreaterThanOrEqualOpStr = "greaterThanOrEqual"
	LesserThanOrEqualOpStr  = "lesserThanOrEqual"
	ContainsOpStr           = "contains"
	NotContainsOpStr        = "notContains"
)

// UserPropertyGroupByPresent Sent from frontend for breakdown on latest user property.
const UserPropertyGroupByPresent string = "$present"

// Query cache related contants.
const (
	QueryCacheInProgressPlaceholder string = "QUERY_CACHE_IN_PROGRESS"

	DateRangePreset2MinLabel      string = "2MIN"
	DateRangePreset30MinLabel     string = "30MIN"
	DateRangePreset2MinInSeconds  int64  = 2 * 60
	DateRangePreset30MinInSeconds int64  = 30 * 60

	QueryCachePlaceholderExpirySeconds     float64 = 2 * 60 * 60       // 2 Hours.
	QueryCacheImmutableResultExpirySeconds float64 = 30 * 24 * 60 * 60 // 30 Days.
	QueryCacheMutableResultExpirySeconds   float64 = 10 * 60           // 10 Minutes.

	QueryCacheRequestSleepHeader       string = "QuerySleepSeconds"
	QueryCacheResponseFromCacheHeader  string = "Fromcache"
	QueryCacheResponseCacheRefreshedAt string = "Refreshedat"
	QueryCacheRedisKeyPrefix           string = "query:cache"
)

const (
	QueryTypeEventsOccurrence = "events_occurrence"
	QueryTypeUniqueUsers      = "unique_users"
)

// BaseQuery Base query interface for all query classes.
type BaseQuery interface {
	GetClass() string
	GetQueryDateRange() (int64, int64)
	SetQueryDateRange(from, to int64)

	// Query cache related helper methods.
	GetQueryCacheHashString() (string, error)
	GetQueryCacheRedisKey(projectID uint64) (*cacheRedis.Key, error)
	GetQueryCacheExpiry() float64
}

type Query struct {
	Class                string                     `json:"cl"`
	Type                 string                     `json:"ty"`
	EventsCondition      string                     `json:"ec"` // all or any
	EventsWithProperties []QueryEventWithProperties `json:"ewp"`
	GroupByProperties    []QueryGroupByProperty     `json:"gbp"`
	GroupByTimestamp     interface{}                `json:"gbt"`
	Timezone             string                     `json:"tz"`
	From                 int64                      `json:"fr"`
	To                   int64                      `json:"to"`
	// Deprecated: Keeping it for old dashboard units.
	OverridePeriod    bool  `json:"ovp"`
	SessionStartEvent int64 `json:"sse"`
	SessionEndEvent   int64 `json:"see"`
}

func (q *Query) GetClass() string {
	return q.Class
}

func (q *Query) GetQueryDateRange() (from, to int64) {
	return q.From, q.To
}

func (q *Query) SetQueryDateRange(from, to int64) {
	q.From, q.To = from, to
}

func (q *Query) GetQueryCacheHashString() (string, error) {
	queryMap, err := U.EncodeStructTypeToMap(q)
	if err != nil {
		return "", err
	}
	delete(queryMap, "fr")
	delete(queryMap, "to")

	queryHash, err := U.GenerateHashStringForStruct(queryMap)
	if err != nil {
		return "", err
	}
	return queryHash, nil
}

func (q *Query) GetQueryCacheRedisKey(projectID uint64) (*cacheRedis.Key, error) {
	hashString, err := q.GetQueryCacheHashString()
	if err != nil {
		return nil, err
	}
	suffix := getQueryCacheRedisKeySuffix(hashString, q.From, q.To)
	return cacheRedis.NewKey(projectID, QueryCacheRedisKeyPrefix, suffix)
}

func (q *Query) GetQueryCacheExpiry() float64 {
	return getQueryCacheResultExpiry(q.From, q.To)
}

func (query *Query) GetGroupByTimestamp() string {
	windowInSecs := query.To - query.From
	switch query.GroupByTimestamp.(type) {
	case bool:
		// For query objects on old dashboard units,
		// with GroupByTimestamp as bool and true, to work.
		if query.GroupByTimestamp.(bool) {
			if windowInSecs <= 86400 {
				return GroupByTimestampHour
			}
			return GroupByTimestampDate
		}

		return ""
	case string:
		gbt := query.GroupByTimestamp.(string)
		if gbt != "" && windowInSecs < U.SECONDS_IN_A_DAY {
			return GroupByTimestampHour
		}
		return gbt
	default:
		return ""
	}
}

type QueryProperty struct {
	// Entity: user or event.
	Entity string `json:"en"`
	// Type: categorical or numerical
	Type      string `json:"ty"`
	Property  string `json:"pr"`
	Operator  string `json:"op"`
	Value     string `json:"va"`
	LogicalOp string `json:"lop"`
}

type QueryGroupByProperty struct {
	// Entity: user or event.
	Entity      string `json:"en"`
	Property    string `json:"pr"`
	Index       int    `json:"in"`
	Type        string `json:"pty"`  // Property type categorical / numerical.
	GroupByType string `json:"gbty"` // With buckets or raw.
	// group by specific event name.
	EventName      string `json:"ena"`
	EventNameIndex int    `json:"eni"`
	Granularity    string `json:"grn"` // currently used only for datetime - year/month/week/day/hour
}

type QueryEventWithProperties struct {
	Name       string          `json:"na"`
	Properties []QueryProperty `json:"pr"`
}

// QueryGroup - Group of query objects.
type QueryGroup struct {
	Queries []Query `json:"query_group"`
}

func (q *QueryGroup) GetClass() string {
	if len(q.Queries) > 0 {
		// all queries in query group are expected to belong to same class
		return q.Queries[0].Class
	}
	return ""
}

func (q *QueryGroup) GetQueryDateRange() (from, to int64) {
	if len(q.Queries) > 0 {
		// all queries in query group are expected to run for same time range
		return q.Queries[0].From, q.Queries[0].To
	}
	return 0, 0
}

func (q *QueryGroup) SetQueryDateRange(from, to int64) {
	for i := 0; i < len(q.Queries); i++ {
		q.Queries[i].From, q.Queries[i].To = from, to
	}
}

func (q *QueryGroup) GetQueryCacheHashString() (string, error) {
	queryMap, err := U.EncodeStructTypeToMap(q)
	if err != nil {
		return "", err
	}
	queries := queryMap["query_group"].([]interface{})
	for _, query := range queries {
		delete(query.(map[string]interface{}), "fr")
		delete(query.(map[string]interface{}), "to")
	}

	queryHash, err := U.GenerateHashStringForStruct(queryMap)
	if err != nil {
		return "", err
	}
	return queryHash, nil
}

func (q *QueryGroup) GetQueryCacheRedisKey(projectID uint64) (*cacheRedis.Key, error) {
	hashString, err := q.GetQueryCacheHashString()
	if err != nil {
		return nil, err
	}
	suffix := getQueryCacheRedisKeySuffix(hashString, q.Queries[0].From, q.Queries[0].To)
	return cacheRedis.NewKey(projectID, QueryCacheRedisKeyPrefix, suffix)
}

func (q *QueryGroup) GetQueryCacheExpiry() float64 {
	return getQueryCacheResultExpiry(q.Queries[0].From, q.Queries[0].To)
}

type DateTimePropertyValue struct {
	From           int64 `json:"fr"`
	To             int64 `json:"to"`
	OverridePeriod bool  `json:"ovp"`
}

func DecodeDateTimePropertyValue(dateTimeJson string) (*DateTimePropertyValue, error) {
	var dateTimeProperty DateTimePropertyValue
	err := json.Unmarshal([]byte(dateTimeJson), &dateTimeProperty)
	if err != nil {
		return &dateTimeProperty, err
	}

	return &dateTimeProperty, nil
}

type QueryResultMeta struct {
	Query       Query        `json:"query"`
	Currency    string       `json:"currency"` //Currency field is used for Attribution query response.
	MetaMetrics []HeaderRows `json:"metrics"`
}

type HeaderRows struct {
	Title   string          `json:"title"`
	Headers []string        `json:"headers"`
	Rows    [][]interface{} `json:"rows"`
}

func getQueryCacheRedisKeySuffix(hashString string, from, to int64) string {
	if to-from == DateRangePreset2MinInSeconds {
		return fmt.Sprintf("%s:%s", hashString, DateRangePreset2MinLabel)
	} else if to-from == DateRangePreset30MinInSeconds {
		return fmt.Sprintf("%s:%s", hashString, DateRangePreset30MinLabel)
	} else if U.IsStartOfTodaysRange(from, U.TimeZoneStringIST) {
		return fmt.Sprintf("%s:from:%d", hashString, from)
	}
	return fmt.Sprintf("%s:from:%d:to:%d", hashString, from, to)
}

func getQueryCacheResultExpiry(from, to int64) float64 {
	if to-from == DateRangePreset2MinInSeconds || to-from == DateRangePreset30MinInSeconds ||
		U.IsStartOfTodaysRange(from, U.TimeZoneStringIST) {
		return QueryCacheMutableResultExpirySeconds
	}
	return QueryCacheImmutableResultExpirySeconds
}

type QueryResult struct {
	Headers []string        `json:"headers"`
	Rows    [][]interface{} `json:"rows"`
	// Todo(Dinesh): Use Generic query result
	// for meta as interface{}.
	Meta QueryResultMeta `json:"meta"`
}

type ResultGroup struct {
	Results []QueryResult `json:"result_group"`
}

// QueryCacheResult Container to save query cache result along with timestamp.
type QueryCacheResult struct {
	Result      interface{}
	RefreshedAt int64
}

// GenericQueryResult - Common query result
// structure with meta.
type GenericQueryResult struct {
	Headers []string        `json:"headers"`
	Rows    [][]interface{} `json:"rows"`
	Meta    interface{}     `json:"meta"`
}

// NamedQueryUnit - Query structure for dashboard unit.
type NamedQueryUnit struct {
	Class     string `json:"cl"`
	Type      string `json:"type"`
	QueryName string `json:"qname"`
}

// GetQueryResultFromCache To get value from cache for a particular query payload.
// resultContainer to be passed by reference.
func GetQueryResultFromCache(projectID uint64, query BaseQuery,
	resultContainer interface{}) (QueryCacheResult, int) {

	logCtx := log.WithFields(log.Fields{
		"ProjectID": projectID,
	})

	var queryResult QueryCacheResult
	cacheKey, err := query.GetQueryCacheRedisKey(projectID)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get cache key")
		return queryResult, http.StatusInternalServerError
	}

	// Using persistent redis for this.
	value, exists, err := cacheRedis.GetIfExistsPersistent(cacheKey)
	if err != nil {
		logCtx.WithError(err).Error("Error getting value from redis")
		return queryResult, http.StatusInternalServerError
	}
	if !exists {
		return queryResult, http.StatusNotFound
	} else if value == QueryCacheInProgressPlaceholder {
		return queryResult, http.StatusAccepted
	}

	err = json.Unmarshal([]byte(value), &queryResult)
	if err != nil {
		logCtx.WithError(err).Error("Failed to unmarshal cache result to result container")
		return queryResult, http.StatusInternalServerError
	}

	err = json.Unmarshal([]byte(value), resultContainer)
	if err != nil {
		logCtx.WithError(err).Error("Failed to unmarshal cache result to result container")
		return queryResult, http.StatusInternalServerError
	}

	return queryResult, http.StatusFound
}

// GetBucketRangeForStartAndEnd Converts 2 - 2 range types to 2.
func GetBucketRangeForStartAndEnd(rangeStart, rangeEnd interface{}) string {
	if rangeStart == rangeEnd {
		return fmt.Sprintf("%v", rangeStart)
	}
	return fmt.Sprintf("%v%s%v", rangeStart, NumericalGroupBySeparator, rangeEnd)
}

func DecodeQueryForClass(queryJSON postgres.Jsonb, queryClass string) (BaseQuery, error) {
	var baseQuery BaseQuery
	var err error
	switch queryClass {
	case QueryClassFunnel, QueryClassInsights:
		var query Query
		err = U.DecodePostgresJsonbToStructType(&queryJSON, &query)
		baseQuery = &query
	case QueryClassAttribution:
		var query AttributionQueryUnit
		err = U.DecodePostgresJsonbToStructType(&queryJSON, &query)
		baseQuery = &query
	case QueryClassChannel:
		var query ChannelQueryUnit
		err = U.DecodePostgresJsonbToStructType(&queryJSON, &query)
		baseQuery = &query
	case QueryClassChannelV1:
		var query ChannelGroupQueryV1
		err = U.DecodePostgresJsonbToStructType(&queryJSON, &query)
		baseQuery = &query
	case QueryClassEvents:
		var query QueryGroup
		err = U.DecodePostgresJsonbToStructType(&queryJSON, &query)
		baseQuery = &query
	case QueryClassWeb:
		var query DashboardUnitsWebAnalyticsQuery
		err = U.DecodePostgresJsonbToStructType(&queryJSON, &query)
		baseQuery = &query
	default:
		return baseQuery, fmt.Errorf("query class %s not supported", queryClass)
	}

	return baseQuery, err
}

// SetQueryCachePlaceholder To set a placeholder temporarily to indicate that query is already running.
func SetQueryCachePlaceholder(projectID uint64, query BaseQuery) {
	cacheKey, err := query.GetQueryCacheRedisKey(projectID)
	if err != nil {
		return
	}

	cacheRedis.SetPersistent(cacheKey, QueryCacheInProgressPlaceholder, QueryCachePlaceholderExpirySeconds)
}

// SetQueryCacheResult Sets the query cache result key in redis.
func SetQueryCacheResult(projectID uint64, query BaseQuery, queryResult interface{}) {
	cacheKey, err := query.GetQueryCacheRedisKey(projectID)
	if err != nil {
		return
	}

	queryCache := QueryCacheResult{
		Result:      queryResult,
		RefreshedAt: U.TimeNowIn(U.TimeZoneStringIST).Unix(),
	}

	queryResultString, err := json.Marshal(queryCache)
	if err != nil {
		return
	}
	cacheRedis.SetPersistent(cacheKey, string(queryResultString), query.GetQueryCacheExpiry())
}

// DeleteQueryCacheKey Delete a query cache key on error.
func DeleteQueryCacheKey(projectID uint64, query BaseQuery) {
	cacheKey, err := query.GetQueryCacheRedisKey(projectID)
	if err != nil {
		return
	}

	cacheRedis.DelPersistent(cacheKey)
}

// TransformQueryPlaceholdersForContext Converts ? in queries to $1, $2 format.
func TransformQueryPlaceholdersForContext(stmnt string) string {
	var newStmnt string
	placeholderCount := 1
	for _, c := range stmnt {
		if c == '?' {
			newStmnt += fmt.Sprintf("$%d", placeholderCount)
			placeholderCount++
		} else {
			newStmnt += string(c)
		}
	}
	return newStmnt
}

// ExpandArrayWithIndividualValues Converts query string ...value IN (?) with array param to ...value IN (?, ?).
// Expands array param to the params values. To support array param in sql.DB.Query.
func ExpandArrayWithIndividualValues(stmnt string, params []interface{}) (string, []interface{}) {
	var newStmnt string
	var newParams []interface{}
	placeholderIndex := 0
	for _, c := range stmnt {
		if c == '?' {
			param := params[placeholderIndex]
			if reflect.TypeOf(param).Kind() == reflect.Slice || reflect.TypeOf(param).Kind() == reflect.Array {
				arrayParam := reflect.ValueOf(param)
				for j := 0; j < arrayParam.Len(); j++ {
					if j == 0 {
						newStmnt += "?"
					} else {
						newStmnt += ", ?"
					}
					newParams = append(newParams, arrayParam.Index(j).Interface())
				}
			} else {
				newStmnt += string(c)
				newParams = append(newParams, params[placeholderIndex])
			}
			placeholderIndex++
		} else {
			newStmnt += string(c)
		}
	}
	return newStmnt, newParams
}
