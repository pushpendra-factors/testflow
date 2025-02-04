package model

import (
	"encoding/json"
	"errors"
	"factors/cache"
	pCache "factors/cache/persistent"
	C "factors/config"
	U "factors/util"
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

const (
	QueryClassEvents        = "events"
	QueryClassInsights      = "insights"
	QueryClassFunnel        = "funnel"
	QueryClassChannel       = "channel"
	QueryClassChannelV1     = "channel_v1"
	QueryClassAttribution   = "attribution"
	QueryClassAttributionV1 = "attribution_v1"
	QueryClassWeb           = "web"
	QueryClassKPI           = "kpi"
	QueryClassProfiles      = "profiles"
	QueryClassAccounts      = "accounts" // This is not used with BaseQuery interface.

	PresentationScatterPlot   = "sp"
	PresentationLine          = "pl"
	PresentationBar           = "pb"
	PresentationTable         = "pt"
	PresentationCard          = "pc"
	PresentationFunnel        = "pf"
	PresentationStack         = "ps"
	PresentationArea          = "pa"
	PresentationHorizontalBar = "ph"
	HorizontalBar             = "hb"
	FunnelChart               = "fc"
	MetricChart               = "mc"
)

const (
	GroupByTimestampSecond  = "second"
	GroupByTimestampHour    = "hour"
	GroupByTimestampDate    = "date"
	GroupByTimestampWeek    = "week"
	GroupByTimestampMonth   = "month"
	GroupByTimestampQuarter = "quarter"
)

const (
	EventCondAnyGivenEvent       = "any_given_event"
	EventCondAllGivenEvent       = "all_given_event"
	EventCondEachGivenEvent      = "each_given_event"
	EventCondFunnelAnyGivenEvent = "funnel_any_given_event"
)

const (
	PropertyEntityUser        = "user"
	PropertyEntityEvent       = "event"
	PropertyEntityUserGlobal  = "user_g"
	PropertyEntityGroup       = "group"
	PropertyEntityUserGroup   = "user_group"
	PropertyEntityDomainGroup = "domain_group"
	MaxEventsLimitInSQL       = 25000
	EventsLimit               = 2500
)

const PropertyValueNone = "$none"
const PropertyValueZero = 0
const ChannelGroupValueDirect = "Direct"

// FilterOptLimit - Limit used for preloading with non-json filters as part of optimisation.
const FilterOptLimit = 10000000000

const (
	ErrUnsupportedGroupByEventPropertyOnUserQuery = "group by event property is not supported for user query"
	ErrMsgQueryProcessingFailure                  = "Failed processing query"
	ErrMsgMaxFunnelStepsExceeded                  = "Max funnel steps exceeded"
	ErrMsgFunnelQueryV2Failure                    = "Invalid funnel query v2"
)

const (
	SelectDefaultEventFilter        = "events.id as event_id, events.user_id as event_user_id"
	SelectDefaultGroupEventFilter   = "events.id as event_id, users.users_user_id as event_user_id"
	SelectDefaultUserFilter         = "events.user_id as event_user_id"
	SelectDefaultEventFilterByAlias = "event_id, event_user_id, event_name"
	SelectDefaultUserFilterByAlias  = "coal_user_id, event_user_id, event_name"
)

const (
	GroupKeyPrefix  = "_group_key_"
	AliasEventName  = "event_name"
	AliasEventIndex = "event_index"
	AliasDateTime   = "datetime"
	AliasAggr       = "aggregate"
	DefaultAggrFunc = "count"
	AliasError      = "error"
)

const DefaultTimezone = "UTC"

const (
	ResultsLimit                  = 10000
	FilterValuesOrEventNamesLimit = 2500
	MaxResultsLimit               = 100000
)

const (
	StepPrefix             = "step_"
	FunnelConversionPrefix = "conversion_"
	FunnelTimeSuffix       = "_time"
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
	ILikeOp                 = "ILIKE"
	NotiLikeOp              = "NOT ILIKE"
	RLikeOp                 = "RLIKE"
	NotRLikeOp              = "NOT RLIKE"
	NotEqualOpStr           = "notEqual"
	NotEqualOp              = "!="
	GreaterThanOpStr        = "greaterThan"
	LesserThanOpStr         = "lesserThan"
	GreaterThanOrEqualOpStr = "greaterThanOrEqual"
	LesserThanOrEqualOpStr  = "lesserThanOrEqual"
	ContainsOpStr           = "contains"
	NotContainsOpStr        = "notContains"
	BetweenStr              = "between"
	NotInBetweenStr         = "notInBetween"
	BeforeStr               = "before"
	SinceStr                = "since"
	InLastStr               = "inLast"
	NotInLastStr            = "notInLast"
	InCurrent               = "inCurrent"
	NotInCurrent            = "notInCurrent"
	InPrevious              = "inPrevious"
	NotInPrevious           = "notInPrevious"
	StartsWith              = "startsWith"
	EndsWith                = "endsWith"
	InList                  = "inList"
	NotInList               = "notInList"
	IsKnown                 = "isKnown"
	IsUnKnown               = "isUnKnown"
)

// UserPropertyGroupByPresent Sent from frontend for breakdown on latest user property.
const UserPropertyGroupByPresent string = "$present"

// NumericalValuePostgresRegex Used to remove non numerical values in numerical bucketing.
const NumericalValuePostgresRegex string = "\\$none|^-?[0-9]+\\.?[0-9]*$"

// Query cache related constants.
const (
	QueryCacheInProgressPlaceholder string = "QUERY_CACHE_IN_PROGRESS"

	DateRangePreset2MinLabel      string = "2MIN"
	DateRangePreset30MinLabel     string = "30MIN"
	DateRangePreset2MinInSeconds  int64  = 2 * 60
	DateRangePreset30MinInSeconds int64  = 30 * 60

	QueryCachePlaceholderExpirySeconds   float64 = 2 * 60 * 60       // 2 Hours.
	QueryCacheMutableResultExpirySeconds float64 = 10 * 60           // 10 Minutes.
	QueryCacheMutableResultMonth         float64 = 24 * 60 * 60 * 30 // 1 Month (30 days).

	QueryCacheRequestInvalidatedCacheHeader string = "Invalidate-Cache"
	QueryFunnelV2                           string = "Funnel-V2"
	QueryCacheRequestSleepHeader            string = "QuerySleepSeconds"
	QueryCacheResponseFromCacheHeader       string = "Fromcache"
	QueryCacheResponseCacheRefreshedAt      string = "Refreshedat"
	QueryCacheResponseCacheTimeZone         string = "TimeZone"
	QueryCacheRedisKeyPrefix                string = "query:cache"
	QueryCacheKeyForSegmentAnalytics        string = "segment:cache"

	WidgetGroupMarker string = "widget_marker"
)

const (
	QueryTypeEventsOccurrence = "events_occurrence"
	QueryTypeUniqueUsers      = "unique_users"
)

var GroupByTimestampTypes = []string{
	GroupByTimestampDate,
	GroupByTimestampHour,
	GroupByTimestampWeek,
	GroupByTimestampMonth,
	GroupByTimestampQuarter,
}

// BaseQuery Base query interface for all query classes.
type BaseQuery interface {
	GetClass() string
	GetQueryDateRange() (int64, int64)
	SetQueryDateRange(from, to int64)
	SetTimeZone(timezoneString U.TimeZoneString)
	GetTimeZone() U.TimeZoneString

	// Query cache related helper methods.
	GetQueryCacheHashString() (string, error)
	GetQueryCacheRedisKey(projectID int64) (*cache.Key, error)
	GetQueryCacheExpiry(projectID int64) float64
	TransformDateTypeFilters() error
	ConvertAllDatesFromTimezone1ToTimezone2(currentTimezone, nextTimezone string) error
	CheckIfNameIsPresent(nameOfQuery string) bool
	SetDefaultGroupByTimestamp()
	GetGroupByTimestamps() []string
}

type Query struct {
	Class                string                     `json:"cl"`
	Type                 string                     `json:"ty"`
	EventsCondition      string                     `json:"ec"` // all or any
	EventsWithProperties []QueryEventWithProperties `json:"ewp"`
	GroupByProperties    []QueryGroupByProperty     `json:"gbp"`
	GlobalUserProperties []QueryProperty            `json:"gup"`
	GroupByTimestamp     interface{}                `json:"gbt"`
	Timezone             string                     `json:"tz"`
	From                 int64                      `json:"fr"`
	To                   int64                      `json:"to"`
	GroupAnalysis        string                     `json:"grpa"`
	ConversionTime       string                     `json:"cnvtm"`

	// For EventsWithProperties in segments
	Caller                     string   `gorm:"default:null" json:"caller"`
	Source                     string   `json:"source"`
	TableProps                 []string `json:"table_props"`
	DownloadAccountsLimitGiven bool     `gorm:"default:false" json:"dalg"`
	DownloadAccountsLimit      int64    `json:"dal"`

	// Deprecated: Keeping it for old dashboard units.
	OverridePeriod    bool  `json:"ovp"`
	SessionStartEvent int64 `json:"sse"`
	SessionEndEvent   int64 `json:"see"`

	// For specific case of KPI - single eventType
	AggregateFunction     string `json:"agFn"`
	AggregateProperty     string `json:"agPr"`
	AggregateEntity       string `json:"agEn"`
	AggregatePropertyType string `json:"agTy"`
	IsLimitNotApplicable  bool
}

var IN_PROPERTIES_DEFAULT_QUERY_MAP = map[string]QueryProperty{
	U.IN_HUBSPOT: {
		Entity:    "user_g",
		GroupName: GROUP_NAME_HUBSPOT_COMPANY,
		Type:      U.PropertyTypeNumerical,
		Property:  HSCompanyIDProperty,
		Operator:  GreaterThanOpStr,
		Value:     "0",
	},

	U.IN_G2: {
		Entity:    "user_g",
		GroupName: GROUP_NAME_G2,
		Type:      U.PropertyTypeCategorical,
		Property:  U.G2_DOMAIN,
		Operator:  NotEqualOpStr,
		Value:     "$none",
	},

	U.VISITED_WEBSITE: {
		Entity:    "user_group",
		GroupName: "user",
		Type:      U.PropertyTypeNumerical,
		Property:  U.SP_PAGE_COUNT,
		Operator:  GreaterThanOpStr,
		Value:     "0",
	},

	U.IN_SALESFORCE: {
		Entity:    "user_g",
		GroupName: GROUP_NAME_SALESFORCE_ACCOUNT,
		Type:      U.PropertyTypeCategorical,
		Property:  SFAccountIDProperty,
		Operator:  NotEqualOpStr,
		Value:     "$none",
	},

	U.IN_LINKEDIN: {
		Entity:    "user_g",
		GroupName: GROUP_NAME_LINKEDIN_COMPANY,
		Type:      U.PropertyTypeCategorical,
		Property:  U.LI_DOMAIN,
		Operator:  NotEqualOpStr,
		Value:     "$none",
	},
}

var PROPERTY_TO_TABLE_COLUMN_MAP = map[string]string{
	U.IDENTIFIED_USER_ID: U.CUSTOMER_USER_ID,
}

var USER_PROPERTY_TO_COLUMN_MAP = map[string]string{
	U.IDENTIFIED_USER_ID: "$user_id",
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

func (q *Query) SetTimeZone(timezoneString U.TimeZoneString) {
	q.Timezone = string(timezoneString)
}

func (q *Query) GetTimeZone() U.TimeZoneString {
	return U.TimeZoneString(q.Timezone)
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

func (q *Query) GetQueryCacheRedisKey(projectID int64) (*cache.Key, error) {
	hashString, err := q.GetQueryCacheHashString()
	if err != nil {
		return nil, err
	}
	suffix := getQueryCacheRedisKeySuffix(hashString, q.From, q.To, U.TimeZoneString(q.Timezone))
	return cache.NewKey(projectID, QueryCacheRedisKeyPrefix, suffix)
}

func (q *Query) GetQueryCacheExpiry(projectID int64) float64 {
	return getQueryCacheResultExpiry(projectID, q.From, q.To, q.Timezone)
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
		//we have gbt as second in attribution only
		if gbt == GroupByTimestampSecond {
			return GroupByTimestampSecond
		}
		if gbt != "" && windowInSecs < U.SECONDS_IN_A_DAY {
			return GroupByTimestampHour
		}
		return gbt
	default:
		return ""
	}
}

func (query *Query) GetGroupByTimestamps() []string {
	return []string{query.GetGroupByTimestamp()}
}

func (query *Query) GetAggregateFunction() string {
	if query.AggregateFunction == "" {
		return strings.ToUpper(DefaultAggrFunc)
	} else {
		return query.AggregateFunction
	}
}

func (query *Query) TransformDateTypeFilters() error {
	for _, ewp := range query.EventsWithProperties {
		err := ewp.TransformDateTypeFilters(query.GetTimeZone())
		if err != nil {
			return err
		}
	}
	for i := range query.GlobalUserProperties {
		err := query.GlobalUserProperties[i].TransformDateTypeFilters(query.GetTimeZone())
		if err != nil {
			return err
		}
	}
	return nil
}

func (query *Query) ConvertAllDatesFromTimezone1ToTimezone2(currentTimezone, nextTimezone string) error {
	for _, ewp := range query.EventsWithProperties {
		ewp.ConvertAllDatesFromTimezone1ToTimzone2(currentTimezone, nextTimezone)
	}
	for i := range query.GlobalUserProperties {
		query.GlobalUserProperties[i].ConvertAllDatesFromTimezone1ToTimzone2(currentTimezone, nextTimezone)
	}
	return nil
}

func (query *Query) CheckIfNameIsPresent(nameOfQuery string) bool {
	return false
}

func (query *Query) SetDefaultGroupByTimestamp() {
	defaultGroupByTimestamp := ""
	if (query.Class == QueryClassEvents || query.Class == QueryClassInsights) && query.Type == QueryTypeUniqueUsers &&
		(query.EventsCondition == EventCondAllGivenEvent || query.EventsCondition == EventCondAnyGivenEvent) {
		query.GroupByTimestamp = ""
	} else {
		defaultGroupByTimestamp = GetDefaultGroupByTimestampForQueries(query.From, query.To, query.GetGroupByTimestamp())
		if defaultGroupByTimestamp != "" {
			query.GroupByTimestamp = defaultGroupByTimestamp
		}
	}
}

type QueryProperty struct {
	// Entity: user or event.
	Entity string `json:"en"`
	// Type: categorical or numerical
	Type string `json:"ty"`

	// currenly used only in all accounts to get the group name for the property.
	GroupName   string `json:"grpn"`
	GroupNameID int    `json:"-"` // internal property holds the group id from the GroupName

	Property  string `json:"pr"`
	Operator  string `json:"op"`
	Value     string `json:"va"`
	LogicalOp string `json:"lop"`
}

// Duplicate code present between QueryProperty and KPIFilter
func (qp *QueryProperty) TransformDateTypeFilters(timezoneString U.TimeZoneString) error {
	var dateTimeValue *DateTimePropertyValue
	var err error
	if qp.Type == U.PropertyTypeDateTime {
		dateTimeValue, err = DecodeDateTimePropertyValue(qp.Value)
		if err != nil {
			log.WithError(err).Error("Failed reading dateTimeValue.")
			return err
		}
		transformedFrom, err := getEpochInSecondsFromMilliseconds(dateTimeValue.From)
		if err != nil {
			return err
		}
		transformedTo, err := getEpochInSecondsFromMilliseconds(dateTimeValue.To)
		if err != nil {
			return err
		}
		dateTimeValue.From = transformedFrom
		dateTimeValue.To = transformedTo
		if qp.Operator == InCurrent || qp.Operator == NotInCurrent {
			startTime, _, err := U.GetDynamicRangesForCurrentBasedOnGranularity(dateTimeValue.Granularity, timezoneString)
			if err != nil {
				return err
			}
			dateTimeValue.From = startTime
		}
		if qp.Operator == InPrevious || qp.Operator == NotInPrevious || qp.Operator == InLastStr || qp.Operator == NotInLastStr {
			startTime, endTime, err := U.GetDynamicPreviousRanges(dateTimeValue.Granularity, dateTimeValue.Number, timezoneString)
			if err != nil {
				return err
			}
			dateTimeValue.From = startTime
			dateTimeValue.To = endTime
		}

		transformedValue, _ := json.Marshal(dateTimeValue)
		qp.Value = string(transformedValue)
	}
	return nil
}

func (qp *QueryProperty) ConvertAllDatesFromTimezone1ToTimzone2(currentTimezone, nextTimezone string) error {
	var dateTimeValue *DateTimePropertyValue
	var err error
	if qp.Type == U.PropertyTypeDateTime {
		dateTimeValue, err = DecodeDateTimePropertyValue(qp.Value)
		if err != nil {
			log.WithError(err).Error("Failed reading dateTimeValue.")
			return err
		}
		transformedFrom, err := getEpochInSecondsFromMilliseconds(dateTimeValue.From)
		if err != nil {
			return err
		}
		transformedTo, err := getEpochInSecondsFromMilliseconds(dateTimeValue.To)
		if err != nil {
			return err
		}
		if qp.Operator == BetweenStr || qp.Operator == NotInBetweenStr {
			transformedFrom = U.GetStartOfDateEpochInOtherTimezone(transformedFrom, currentTimezone, nextTimezone)
			transformedTo = U.GetEndOfDateEpochInOtherTimezone(transformedTo, currentTimezone, nextTimezone)
		} else if qp.Operator == BeforeStr {
			transformedTo = U.GetStartOfDateEpochInOtherTimezone(transformedTo, currentTimezone, nextTimezone)
		} else if qp.Operator == SinceStr {
			transformedFrom = U.GetStartOfDateEpochInOtherTimezone(transformedFrom, currentTimezone, nextTimezone)
		}
		dateTimeValue.From = transformedFrom
		dateTimeValue.To = transformedTo
		transformedValue, _ := json.Marshal(dateTimeValue)
		qp.Value = string(transformedValue)
	}
	return nil
}

func getEpochInSecondsFromMilliseconds(epoch int64) (int64, error) {
	if epoch == 0 {
		return epoch, nil
	}
	countOfDigits := U.GetNumberOfDigits(epoch)
	if countOfDigits == 10 {
		return epoch, nil
	} else if countOfDigits == 13 {
		return epoch / 1000, nil
	} else {
		return epoch, errors.New("Wrong date type filter range is given.")
	}
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

	// currenly used only in all accounts to get the group name for the property.
	GroupName   string `json:"grpn"`
	GroupNameID int    `json:"-"` // internal property holds the group id from the GroupName
}

type QueryEventWithProperties struct {
	Name              string          `json:"na"`
	AliasName         string          `json:"an"`
	GroupAnalysis     string          `json:"grpa"`
	Properties        []QueryProperty `json:"pr"`
	EventNameIDs      []interface{}   `json:"-"`
	Range             int64           `json:"range"` // number of days
	FrequencyOperator string          `json:"freqOp"`
	Frequency         string          `json:"freq"`
	IsEventPerformed  bool            `json:"iep"`
}

func (ewp *QueryEventWithProperties) TransformDateTypeFilters(timezoneString U.TimeZoneString) error {
	for i := range ewp.Properties {
		err := ewp.Properties[i].TransformDateTypeFilters(timezoneString)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ewp *QueryEventWithProperties) ConvertAllDatesFromTimezone1ToTimzone2(currentTimezone, nextTimezone string) error {
	for i := range ewp.Properties {
		err := ewp.Properties[i].ConvertAllDatesFromTimezone1ToTimzone2(currentTimezone, nextTimezone)
		if err != nil {
			return err
		}
	}
	return nil
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

func (q *QueryGroup) SetTimeZone(timezoneString U.TimeZoneString) {
	for i := 0; i < len(q.Queries); i++ {
		q.Queries[i].Timezone = string(timezoneString)
	}
}

func (q *QueryGroup) GetTimeZone() U.TimeZoneString {
	return U.TimeZoneString(q.Queries[0].Timezone)
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

func (q *QueryGroup) GetQueryCacheRedisKey(projectID int64) (*cache.Key, error) {
	hashString, err := q.GetQueryCacheHashString()
	if err != nil {
		return nil, err
	}
	suffix := getQueryCacheRedisKeySuffix(hashString, q.Queries[0].From, q.Queries[0].To, U.TimeZoneString(q.Queries[0].Timezone))
	return cache.NewKey(projectID, QueryCacheRedisKeyPrefix, suffix)
}

func (q *QueryGroup) GetQueryCacheExpiry(projectID int64) float64 {
	return getQueryCacheResultExpiry(projectID, q.Queries[0].From, q.Queries[0].To, q.Queries[0].Timezone)
}

func (q *QueryGroup) TransformDateTypeFilters() error {
	for _, query := range q.Queries {
		err := query.TransformDateTypeFilters()
		if err != nil {
			return err
		}
	}
	return nil
}

func (q *QueryGroup) CheckIfNameIsPresent(nameOfQuery string) bool {
	return false
}

func (q *QueryGroup) ConvertAllDatesFromTimezone1ToTimezone2(currentTimezone, nextTimezone string) error {
	for i := range q.Queries {
		q.Queries[i].ConvertAllDatesFromTimezone1ToTimezone2(currentTimezone, nextTimezone)
	}
	return nil
}

func (query *QueryGroup) SetDefaultGroupByTimestamp() {
	for index, _ := range query.Queries {
		defaultGroupByTimestamp := ""
		if (query.Queries[index].Class == QueryClassEvents || query.Queries[index].Class == QueryClassInsights) && query.Queries[index].Type == QueryTypeUniqueUsers &&
			(query.Queries[index].EventsCondition == EventCondAllGivenEvent || query.Queries[index].EventsCondition == EventCondAnyGivenEvent) {
			query.Queries[index].GroupByTimestamp = ""
		} else {
			defaultGroupByTimestamp = GetDefaultGroupByTimestampForQueries(query.Queries[index].From, query.Queries[index].To, query.Queries[index].GetGroupByTimestamp())
			if defaultGroupByTimestamp != "" {
				query.Queries[index].GroupByTimestamp = defaultGroupByTimestamp
			}
		}
	}
}

func (query *QueryGroup) GetGroupByTimestamps() []string {
	queryResultString := make([]string, 0)
	for _, intQuery := range query.Queries {
		queryResultString = append(queryResultString, intQuery.GetGroupByTimestamp())
	}
	return queryResultString
}

type DateTimePropertyValue struct {
	From           int64  `json:"fr"`
	To             int64  `json:"to"`
	OverridePeriod bool   `json:"ovp"`
	Number         int64  `json:"num"`
	Granularity    string `json:"gran"`
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

func getQueryCacheRedisKeySuffix(hashString string, from, to int64, timezoneString U.TimeZoneString) string {
	if to-from == DateRangePreset2MinInSeconds {
		return fmt.Sprintf("%s:%s", hashString, DateRangePreset2MinLabel)
	} else if to-from == DateRangePreset30MinInSeconds {
		return fmt.Sprintf("%s:%s", hashString, DateRangePreset30MinLabel)
	} else if U.IsStartOfTodaysRangeIn(from, timezoneString) {
		return fmt.Sprintf("%s:from:%d", hashString, from)
	}
	return fmt.Sprintf("%s:from:%d:to:%d", hashString, from, to)
}

func getQueryCacheResultExpiry(projectID, from, to int64, timezone string) float64 {
	var timezoneString U.TimeZoneString
	timezoneString = U.TimeZoneString(timezone)
	if to-from == DateRangePreset2MinInSeconds || to-from == DateRangePreset30MinInSeconds {
		return QueryCacheMutableResultExpirySeconds
	}
	if C.IsProjectAllowedForLongerExpiry(projectID) {
		// Approx 1 months for any query less than 3 months
		if to-from < (15 * U.SECONDS_IN_A_DAY) {
			return float64(31 * U.SECONDS_IN_A_DAY)
		}
		// Approx 3 months for any query more than a month
		if to-from > (27 * U.SECONDS_IN_A_DAY) {
			return float64(93 * U.SECONDS_IN_A_DAY)
		}
		// for anything between 15 to 1 month - 31 days
		return float64(U.CacheExpiryDefaultInSeconds)
	}
	return U.GetQueryCacheResultExpiryInSeconds(from, to, timezoneString)
}

type QueryResult struct {
	Headers []string        `json:"headers"`
	Rows    [][]interface{} `json:"rows"`
	// Todo(Dinesh): Use Generic query result
	// for meta as interface{}.
	Meta      QueryResultMeta `json:"meta"`
	Query     interface{}     `json:"query"`
	CacheMeta interface{}     `json:"cache_meta"`
}
type ResultGroup struct {
	Results     []QueryResult `json:"result_group"`
	Query       interface{}   `json:"query"`
	IsShareable bool          `json:"is_shareable"`
	CacheMeta   interface{}   `json:"cache_meta"`
}

// QueryCacheResult Container to save query cache result along with timestamp.
type QueryCacheResult struct {
	Result      interface{}
	RefreshedAt int64
	TimeZone    string
	CacheMeta   interface{} `json:"cache_meta"`
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
// NOTE: Duplicate code in GetSegmentResultFromCache. Slightly different but verify.
func GetQueryResultFromCache(projectID int64, query BaseQuery,
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
	value, exists, err := pCache.GetIfExists(cacheKey, true)
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
	case QueryClassAttributionV1:
		var query AttributionQueryUnitV1
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
	case QueryClassKPI:
		var query KPIQueryGroup
		err = U.DecodePostgresJsonbToStructType(&queryJSON, &query)
		baseQuery = &query
	case QueryClassWeb:
		var query DashboardUnitsWebAnalyticsQuery
		err = U.DecodePostgresJsonbToStructType(&queryJSON, &query)
		baseQuery = &query
	case QueryClassProfiles:
		var query ProfileQueryGroup
		err = U.DecodePostgresJsonbToStructType(&queryJSON, &query)
		baseQuery = &query
	default:
		return baseQuery, fmt.Errorf("query class %s not supported", queryClass)
	}

	return baseQuery, err
}

// SetQueryCachePlaceholder To set a placeholder temporarily to indicate that query is already running.
func SetQueryCachePlaceholder(projectID int64, query BaseQuery) {
	cacheKey, err := query.GetQueryCacheRedisKey(projectID)
	if err != nil {
		return
	}

	pCache.Set(cacheKey, QueryCacheInProgressPlaceholder, QueryCachePlaceholderExpirySeconds, true)
}

type CacheMeta struct {
	From           int64  `json:"from"`
	To             int64  `json:"to"`
	RefreshedAt    int64  `json:"refreshed_at"`
	Preset         string `json:"preset"`
	LastComputedAt int64  `json:"last_computed_at"`
	Timezone       string `json:"timezone"`
}

// SetQueryCacheResult Sets the query cache result key in redis.
func SetQueryCacheResult(projectID int64, query BaseQuery, queryResult interface{}) {
	cacheKey, err := query.GetQueryCacheRedisKey(projectID)
	if err != nil {
		return
	}
	from, to := query.GetQueryDateRange()
	meta := CacheMeta{
		From:           from,
		To:             to,
		RefreshedAt:    U.TimeNowIn(U.TimeZoneStringIST).Unix(),
		Timezone:       string(query.GetTimeZone()),
		LastComputedAt: U.TimeNowIn(U.TimeZoneStringIST).Unix(),
	}
	queryCache := QueryCacheResult{
		Result:      queryResult,
		RefreshedAt: U.TimeNowIn(U.TimeZoneStringIST).Unix(),
		TimeZone:    string(query.GetTimeZone()),
		CacheMeta:   meta,
	}

	queryResultString, err := json.Marshal(queryCache)
	if err != nil {
		return
	}
	pCache.Set(cacheKey, string(queryResultString), query.GetQueryCacheExpiry(projectID), true)
}

// DeleteQueryCacheKey Delete a query cache key on error.
func DeleteQueryCacheKey(projectID int64, query BaseQuery) {
	cacheKey, err := query.GetQueryCacheRedisKey(projectID)
	if err != nil {
		return
	}

	pCache.Del([]*cache.Key{cacheKey}, true)
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
	defer U.NotifyOnPanicWithError(C.GetConfig().Env, C.GetConfig().AppName)

	placeholdersCount := 0
	for _, c := range stmnt {
		if c == '?' {
			placeholdersCount += 1
		}
	}
	if len(params) != placeholdersCount {
		log.WithField("stmnt", stmnt).WithField("params", params).Error("Parameters mismatch: ")
	}
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

func SanitizeStringSumToNumeric(result *QueryResult) {
	stepIndices := make(map[string]int)
	for i, header := range result.Headers {
		if strings.HasPrefix(header, StepPrefix) {
			stepIndices[header] = i
		}
	}

	for i := range result.Rows {
		for _, j := range stepIndices {
			result.Rows[i][j] = U.SafeConvertToFloat64(result.Rows[i][j])
		}
	}
}

func CheckIfMapHasNoneFilter(propertyToHasNoneFilter map[string]bool, p QueryProperty) bool {
	propertyKey := p.Entity + "." + p.Property
	hasNoneFilter := false
	if exists := propertyToHasNoneFilter[propertyKey]; exists {
		hasNoneFilter = true
	}
	return hasNoneFilter
}

func GetPropertyToHasNoneFilter(properties []QueryProperty) map[string]bool {
	propertyToHasNoneFilter := make(map[string]bool)
	for _, p := range properties {
		if p.Value == PropertyValueNone {
			propertyKey := p.Entity + "." + p.Property
			propertyToHasNoneFilter[propertyKey] = true
		}
	}
	return propertyToHasNoneFilter
}

func GetPropertyToHasNegativeFilter(properties []QueryProperty) []QueryProperty {

	negativeFilters, _ := GetPropertyGroupedNegativeAndPostiveFilter(properties)
	return negativeFilters
}

func GetPropertyToHasPositiveFilter(properties []QueryProperty) []QueryProperty {

	_, positiveFilters := GetPropertyGroupedNegativeAndPostiveFilter(properties)
	return positiveFilters
}

func GetPropertyGroupedNegativeAndPostiveFilter(properties []QueryProperty) ([]QueryProperty, []QueryProperty) {
	negativeFilters := make([]QueryProperty, 0)
	positiveFilters := make([]QueryProperty, 0)
	for _, filter := range properties {
		if IsNegativeFilter(filter) {
			negativeFilters = append(negativeFilters, filter)
			continue
		}
		positiveFilters = append(positiveFilters, filter)
	}
	return negativeFilters, positiveFilters
}

func IsNegativeFilter(filter QueryProperty) bool {
	isNegativeFilter := false

	if (filter.Operator == NotContainsOpStr && filter.Value != PropertyValueNone) ||
		(filter.Operator == ContainsOpStr && filter.Value == PropertyValueNone) ||
		(filter.Operator == NotEqualOpStr && filter.Value != PropertyValueNone) ||
		(filter.Operator == EqualsOpStr && filter.Value == PropertyValueNone) ||
		(filter.Operator == NotInList && filter.Value != PropertyValueNone) {
		isNegativeFilter = true
	}

	return isNegativeFilter
}

func GetNegativeFilterNegated(negativeFilters []QueryProperty) []QueryProperty {
	filters := make([]QueryProperty, 0)
	for _, filter := range negativeFilters {

		if filter.Operator == NotContainsOpStr {
			filter.Operator = ContainsOpStr
		} else if filter.Operator == NotEqualOpStr {
			filter.Operator = EqualsOpStr
		} else if filter.Operator == ContainsOpStr {
			filter.Operator = NotContainsOpStr
		} else if filter.Operator == EqualsOpStr {
			filter.Operator = NotEqualOpStr
		} else if filter.Operator == NotInList {
			filter.Operator = InList
		} else if filter.Operator == InList {
			filter.Operator = NotInList
		}

		filter.LogicalOp = "OR"
		filters = append(filters, filter)
	}
	return filters
}

func GetFilterPropertyGroupIDs(properties [][]QueryProperty) map[int]bool {
	groupIDsMap := make(map[int]bool)
	for i := range properties {
		for _, p := range properties[i] {
			groupIDsMap[p.GroupNameID] = true

		}
	}
	return groupIDsMap
}

// If UI presents filters in "(a or b) AND (c or D)" order, Request has it as "a or b AND c or D"
// Using AND as a separation between lines and execution order to achieve the same as above.
func GetPropertiesGrouped(properties []QueryProperty) [][]QueryProperty {
	groupedProperties := make([][]QueryProperty, 0)
	currentGroupedProperties := make([]QueryProperty, 0)
	for index, p := range properties {
		if index == 0 || p.LogicalOp != "AND" {
			currentGroupedProperties = append(currentGroupedProperties, p)
		} else {
			groupedProperties = append(groupedProperties, currentGroupedProperties)

			currentGroupedProperties = make([]QueryProperty, 0)
			currentGroupedProperties = append(currentGroupedProperties, p)
		}
	}
	if len(currentGroupedProperties) != 0 {
		groupedProperties = append(groupedProperties, currentGroupedProperties)
	}
	return groupedProperties
}

// GetPropertiesGroupedByGroupName groups properties by their group and tells if
// it is only OR or only AND properties
func GetPropertiesGroupedByGroupName(properties []QueryProperty) ([][]QueryProperty, bool, bool) {

	if len(properties) == 0 {
		return nil, false, false
	}

	isOnlyOR := true
	isOnlyAND := true
	if len(properties) == 1 {
		isOnlyOR = false
		isOnlyAND = true
	}

	groupPropertyVisited := make(map[string]map[string]bool)
	groupPropertyVisited[properties[0].GroupName] = map[string]bool{
		properties[0].Property: true,
	}
	for _, property := range properties[1:] {

		if _, exist := groupPropertyVisited[property.GroupName]; !exist {
			groupPropertyVisited[property.GroupName] = make(map[string]bool)
		}

		visited := groupPropertyVisited[property.GroupName][property.Property]

		if !visited && property.LogicalOp == "AND" {
			isOnlyOR = false
		}

		if !visited && property.LogicalOp == "OR" {
			isOnlyAND = false
		}

		groupPropertyVisited[property.GroupName][property.Property] = true
	}

	groupedPropertiesMap := make(map[string][]QueryProperty, 0)
	for i := range properties {
		property := properties[i]
		if _, exist := groupedPropertiesMap[property.GroupName]; !exist {
			groupedPropertiesMap[property.GroupName] = make([]QueryProperty, 0)
		}
		groupedPropertiesMap[property.GroupName] = append(groupedPropertiesMap[property.GroupName], property)
	}

	groupedProperties := make([][]QueryProperty, 0)
	for groupName := range groupedPropertiesMap {
		groupedProperties = append(groupedProperties, groupedPropertiesMap[groupName])
	}

	return groupedProperties, isOnlyOR, isOnlyAND
}

// IsEventLevelGroupBy Checks if the groupBy is for a particular event in query.ewp.
func IsEventLevelGroupBy(groupBy QueryGroupByProperty) bool {

	return groupBy.EventName != "" && groupBy.EventNameIndex != 0
}

func FilterGroupPropsByType(gp []QueryGroupByProperty, entity string) []QueryGroupByProperty {
	groupProps := make([]QueryGroupByProperty, 0)

	for _, v := range gp {
		if v.Entity == entity {
			groupProps = append(groupProps, v)
		}
	}
	return groupProps
}

func removeEventSpecificUserGroupBys(groupBys []QueryGroupByProperty) []QueryGroupByProperty {
	filteredProps := make([]QueryGroupByProperty, 0)
	for _, prop := range groupBys {
		if IsEventLevelGroupBy(prop) {
			// For $present, event name index is not set and is default 0.
			continue
		}
		filteredProps = append(filteredProps, prop)
	}
	return filteredProps
}

func GetGlobalGroupByUserProperties(properties []QueryGroupByProperty) []QueryGroupByProperty {
	userGroupProps := FilterGroupPropsByType(properties, PropertyEntityUser)
	return removeEventSpecificUserGroupBys(userGroupProps)
}

// CheckIfHasNoneFilter Returns if set of filters has $none as a value
func CheckIfHasNoneFilter(properties []QueryProperty) bool {

	for _, p := range properties {
		if p.Value == PropertyValueNone {
			return true
		}
	}
	return false
}

// CheckIfHasGlobalUserFilter Returns if set of filters entity has global filter
func CheckIfHasGlobalUserFilter(properties []QueryProperty) bool {

	for _, p := range properties {
		if p.Entity == PropertyEntityUserGlobal {
			return true
		}
	}
	return false
}

// CheckIfHasDomainFilter Returns if set of filters entity has domain filter
func CheckIfHasDomainFilter(properties []QueryProperty) bool {

	for _, p := range properties {
		if isDomainFilter(p.Entity, p.GroupName) {
			return true
		}
	}
	return false
}

func isDomainFilter(entity, groupName string) bool {
	return entity == PropertyEntityDomainGroup && groupName == GROUP_NAME_DOMAINS
}

// CheckIfHasDomainGroupBy Returns checks if has domain breakdown
func CheckIfHasDomainGroupBy(properties []QueryGroupByProperty) bool {

	for _, p := range properties {
		if isDomainFilter(p.Entity, p.GroupName) {
			return true
		}
	}

	return false
}

// getGroupIDsFromGlobalUserFilter returns group names for the global properties
func getGroupIDsFromGlobalUserGroupByANDFilters(queryProperties []QueryProperty, groupByProperties []QueryGroupByProperty) map[string]int {
	groupNameIDs := make(map[string]int)
	for _, p := range queryProperties {
		groupNameIDs[p.GroupName] = p.GroupNameID
	}

	for _, p := range groupByProperties {
		groupNameIDs[p.GroupName] = p.GroupNameID
	}
	return groupNameIDs
}

// GetDomainsAsscocaitedGroupSourceANDColumnIDs gives the group_x_id and source for properties and breakdowns for domain association
func GetDomainsAsscocaitedGroupSourceANDColumnIDs(globalUserProperties []QueryProperty, userGroupProps []QueryGroupByProperty) (string, string) {
	groupNameIDs := getGroupIDsFromGlobalUserGroupByANDFilters(globalUserProperties, userGroupProps)
	globalGroupIDColumns := ""
	globalGroupSource := ""
	for groupName, groupID := range groupNameIDs {
		if globalGroupIDColumns != "" {
			globalGroupIDColumns += " OR "
		}
		if globalGroupSource != "" {
			globalGroupSource += ","
		}

		globalGroupIDColumns += fmt.Sprintf("group_users.group_%d_id IS NOT NULL", groupID)
		globalGroupSource += fmt.Sprintf("%d", GroupUserSource[groupName])
	}

	return globalGroupIDColumns, globalGroupSource
}

func GetPropertyEntityFieldForFilter(entityName string) string {
	switch entityName {
	case PropertyEntityUser:
		return "events.user_properties"
	case PropertyEntityEvent:
		return "events.properties"
	case PropertyEntityUserGlobal:
		return "users.properties"
	case PropertyEntityUserGroup:
		return "user_global_user_properties"
	case PropertyEntityDomainGroup:
		return "domain_properties"
	}

	return ""
}

func AddMissingEventNamesInResult(result *QueryResult, query *Query) {
	eventNameIndex := getEventNameIndex(result)
	if eventNameIndex == -1 || len(result.Rows) == 0 {
		return
	}

	mapOfEventNamesPresentInResult := make(map[string]bool)
	for _, row := range result.Rows {
		mapOfEventNamesPresentInResult[row[eventNameIndex].(string)] = true
	}

	for index, eventWithProperties := range query.EventsWithProperties {
		eventWithPrefix := fmt.Sprintf("%d_%s", index, eventWithProperties.Name)
		if _, exists := mapOfEventNamesPresentInResult[eventWithPrefix]; !exists {
			defaultRow := make([]interface{}, 0)
			defaultRow = append(defaultRow, result.Rows[0]...)

			defaultRow[len(defaultRow)-1] = 0

			mapOfEventNamesPresentInResult[eventWithPrefix] = true
			defaultRow[eventNameIndex] = eventWithPrefix
			result.Rows = append(result.Rows, defaultRow)
		}
	}
}

// AddAliasNameOnEventCondEachGivenEventQueryResult replaces EventName in the result's header with the AliasName
func AddAliasNameOnEventCondEachGivenEventQueryResult(result *QueryResult, query Query) {
	// Identify the index for the event_name
	eventNameIndex := getEventNameIndex(result)

	// If eventNameIndex == -1, the AliasEventName is not found in the header. Hence skip!
	if eventNameIndex == -1 {
		return
	}

	i := 0
	for i < len(result.Rows) {
		// Fetching the index from the indexed event_name in result.rows, and utilizing it to establish mapping
		// with the corresponding alias_name in the query
		eventName, validConversion := result.Rows[i][eventNameIndex].(string)
		if !validConversion {
			i += 1
			continue
		}
		splitPos := strings.Index(eventName, "_")
		j := eventName[0:splitPos]
		index, err := strconv.Atoi(j)
		if err == nil {
			// Replace the event_name only if corresponding alias_name is provided
			// Replacing the event_name with "{index}_{alias_name}". This is becasue this name will get
			// replaced with pure {alias_name} inside method updateEventNameInHeaderAndAddMeta, which gets
			// invoked subsequently
			if query.EventsWithProperties[index].AliasName != "" {
				result.Rows[i][eventNameIndex] = j + "_" + query.EventsWithProperties[index].AliasName
			}
		} else {
			log.WithError(err).WithField("query : ", query).WithField("result : ", result).Error("Failed to get index of event_name for replacing with alias_name.")
		}
		i += 1
	}
}

func getEventNameIndex(result *QueryResult) int {
	eventNameIndex := -1
	for k, key := range result.Headers {
		if key == AliasEventName {
			eventNameIndex = k
		}
	}
	return eventNameIndex
}

func HasGroupByDateTypeProperties(groupProps []QueryGroupByProperty) bool {
	for _, groupByProp := range groupProps {
		if groupByProp.Type == U.PropertyTypeDateTime {
			return true
		}
	}
	return false
}

// Adds timezone offset to dateType row value for dateType row.
func SanitizeDateTypeRows(result *QueryResult, query *Query) {
	headerIndexMap := make(map[string][]int)
	for index, header := range result.Headers {
		// If same group by is added twice, it will appear twice in headers.
		// Keep as a list to sanitize both indexes.
		headerIndexMap[header] = append(headerIndexMap[header], index)
	}

	alreadySanitizedProperties := make(map[string]bool)
	for _, gbp := range query.GroupByProperties {
		if gbp.Type == U.PropertyTypeDateTime {
			if _, sanitizedAlready := alreadySanitizedProperties[gbp.Property]; sanitizedAlready {
				continue
			}
			indexesToSanitize := headerIndexMap[gbp.Property]
			for _, indexToSanitize := range indexesToSanitize {
				sanitizeDateTypeForSpecificIndex(query, result.Rows, indexToSanitize)
			}
			alreadySanitizedProperties[gbp.Property] = true
		}
	}
}

func sanitizeDateTypeForSpecificIndex(query *Query, rows [][]interface{}, indexToSanitize int) {

	for index, row := range rows {
		if (query.Class == QueryClassFunnel && index == 0) || row[indexToSanitize] == nil || row[indexToSanitize].(string) == "" ||
			row[indexToSanitize].(string) == PropertyValueNone {
			// For funnel queries, first row is $no_group query. Skip sanitization.
			continue
		}
		currentValueInTimeFormat, _ := time.Parse(U.DATETIME_FORMAT_DB, row[indexToSanitize].(string))
		row[indexToSanitize] = U.GetTimestampAsStrWithTimezone(currentValueInTimeFormat, query.Timezone)
	}
}

// Logic which uses from, to and current query Timestamp to set the default GroupByTimestamp.
func GetDefaultGroupByTimestampForQueries(from, to int64, currentGroupByTimestamp string) string {
	if currentGroupByTimestamp != "" && U.IsLessThanTimeRange(from, to, U.SECONDS_IN_A_DAY) {
		return GroupByTimestampHour
	} else if currentGroupByTimestamp == GroupByTimestampHour && U.IsGreaterThanEqualTimeRange(from, to, U.SECONDS_IN_A_DAY) {
		return GroupByTimestampDate
	}

	return ""
}

func IsQueryGroupByLatestUserProperty(queryGroupByProperty []QueryGroupByProperty) bool {
	for i := range queryGroupByProperty {
		if queryGroupByProperty[i].EventName == UserPropertyGroupByPresent {
			return true
		}
	}
	return false
}

var funnelQueryGroupUserID = regexp.MustCompile("user_groups\\.group_\\d_user_id")
var funnelQueryUserGroupUserID = regexp.MustCompile("users\\.group_\\d_user_id")
var queryDomainGroupID = regexp.MustCompile("domains\\.group_\\d_id")

func GetQueryGroupUserID(stmnt string) string {

	return funnelQueryGroupUserID.FindString(stmnt)
}

func GetQueryUserGroupUserID(stmnt string) string {

	return funnelQueryUserGroupUserID.FindString(stmnt)
}

func GetQueryDomainGroupID(stmnt string) string {

	return queryDomainGroupID.FindString(stmnt)
}

func IsValidFunnelQueryGroupName(group string) bool {
	_, exists := AllowedGroupNames[group]
	if exists || IsFunnelQueryGroupNameUser(group) || IsQueryGroupNameAllAccounts(group) {
		return true
	}

	return false
}

func IsFunnelQueryGroupNameUser(group string) bool {
	return group == USERS
}

func IsQueryGroupNameAllAccounts(group string) bool {
	return group == GROUP_NAME_DOMAINS
}

func IsFiltersContainGlobalUserPropertyForDomains(filters []QueryProperty) bool {
	for i := range filters {
		if IsFilterGlobalUserPropertiesByDefaultQueryMap(filters[i].Entity) {
			return true
		}
	}
	return false
}

func IsFilterGlobalUserPropertiesByDefaultQueryMap(entity string) bool {
	return entity == PropertyEntityUserGroup
}

func FilterGlobalGroupPropertiesFilterForDomains(filters []QueryProperty) []QueryProperty {
	filteredGlobalGroupProperties := make([]QueryProperty, 0)
	for i := range filters {
		if IsFilterGlobalUserPropertiesByDefaultQueryMap(filters[i].Entity) {
			continue
		}

		if isDomainFilter(filters[i].Entity, filters[i].GroupName) {
			continue
		}

		filteredGlobalGroupProperties = append(filteredGlobalGroupProperties, filters[i])
	}
	return filteredGlobalGroupProperties
}

func RemoveDomainGroupByProperties(groupByProps []QueryGroupByProperty) []QueryGroupByProperty {
	groupBys := make([]QueryGroupByProperty, 0)
	for i := range groupByProps {
		if isDomainFilter(groupByProps[i].Entity, groupByProps[i].GroupName) {
			continue
		}
		groupBys = append(groupBys, groupByProps[i])
	}

	return groupBys
}

func GetFilteredDomainGroupProperties(groupByProps []QueryProperty) []QueryProperty {
	groupBys := make([]QueryProperty, 0)
	for i := range groupByProps {
		if isDomainFilter(groupByProps[i].Entity, groupByProps[i].GroupName) {
			continue
		}
		groupBys = append(groupBys, groupByProps[i])
	}

	return groupBys
}
