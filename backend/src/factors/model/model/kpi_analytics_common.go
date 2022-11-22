package model

import (
	"encoding/json"
	cacheRedis "factors/cache/redis"
	U "factors/util"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	MetricsDateType       = "date_type"
	MetricsPercentageType = "percentage_type"
	alpha                 = "abcdefghijklmnopqrstuvwxyz"
	KpiStaticQueryType    = "static"
	KpiCustomQueryType    = "custom"
	KpiDerivedQueryType   = "derived"
)

type KPIQueryGroup struct {
	Class         string       `json:"cl"`
	Queries       []KPIQuery   `json:"qG"`
	GlobalFilters []KPIFilter  `json:"gFil"`
	GlobalGroupBy []KPIGroupBy `json:"gGBy"`
	Formula       string       `json:"for"`
}

func (q *KPIQueryGroup) GetClass() string {
	return q.Class
}

func (q *KPIQueryGroup) GetQueryDateRange() (from, to int64) {
	if len(q.Queries) > 0 {
		// all queries in query group are expected to run for same time range
		return q.Queries[0].From, q.Queries[0].To
	}
	return 0, 0
}

func (query *KPIQueryGroup) CheckIfNameIsPresent(nameOfQuery string) bool {
	for _, query := range query.Queries {
		if query.CheckIfNameIsPresent(nameOfQuery) {
			return true
		}
	}
	return false
}

func (q *KPIQueryGroup) SetQueryDateRange(from, to int64) {
	for index, _ := range q.Queries {
		q.Queries[index].From, q.Queries[index].To = from, to
	}
}

func (q *KPIQueryGroup) SetTimeZone(timezoneString U.TimeZoneString) {
	for index, _ := range q.Queries {
		q.Queries[index].Timezone = string(timezoneString)
	}
}

func (q *KPIQueryGroup) GetTimeZone() U.TimeZoneString {
	return U.TimeZoneString(q.Queries[0].Timezone)
}

func (q *KPIQueryGroup) GetQueryCacheHashString() (string, error) {
	queryMap, err := U.EncodeStructTypeToMap(q)
	if err != nil {
		return "", err
	}
	queries := queryMap["qG"].([]interface{})
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

func (q *KPIQueryGroup) GetQueryCacheRedisKey(projectID int64) (*cacheRedis.Key, error) {
	hashString, err := q.GetQueryCacheHashString()
	if err != nil {
		return nil, err
	}
	suffix := getQueryCacheRedisKeySuffix(hashString, q.Queries[0].From, q.Queries[0].To, U.TimeZoneString(q.Queries[0].Timezone))
	return cacheRedis.NewKey(projectID, QueryCacheRedisKeyPrefix, suffix)
}

func (q *KPIQueryGroup) GetQueryCacheExpiry() float64 {
	return getQueryCacheResultExpiry(q.Queries[0].From, q.Queries[0].To, q.Queries[0].Timezone)
}

func (q *KPIQueryGroup) GetGroupByTimestamp() string {
	if q.Queries[0].GroupByTimestamp == "" {
		return ""
	}
	return q.Queries[0].GroupByTimestamp
}

func (q *KPIQueryGroup) TransformDateTypeFilters() error {
	timezoneString := q.GetTimeZone()
	err := transformDateTypeFiltersForKPIFilters(q.GlobalFilters, timezoneString)
	if err != nil {
		return err
	}
	for _, query := range q.Queries {
		err := transformDateTypeFiltersForKPIFilters(query.Filters, timezoneString)
		if err != nil {
			return err
		}
	}
	return nil
}

func (q *KPIQueryGroup) ConvertAllDatesFromTimezone1ToTimezone2(currentTimezone, nextTimezone string) error {
	for i := range q.GlobalFilters {
		q.GlobalFilters[i].ConvertAllDatesFromTimezone1ToTimzone2(currentTimezone, nextTimezone)
	}
	for i := range q.Queries {
		for j := range q.Queries[i].Filters {
			q.Queries[i].Filters[j].ConvertAllDatesFromTimezone1ToTimzone2(currentTimezone, nextTimezone)
		}
	}
	return nil
}

func (q *KPIQueryGroup) IsValid() bool {
	for _, query := range q.Queries {
		if !query.IsValid() {
			return false
		}
	}
	return true
}

func (q *KPIQueryGroup) IsValidDerivedKPI() (bool, string) {
	isValidFormula := U.ValidateArithmeticFormula(q.Formula)
	if !isValidFormula {
		return false, "Invalid arithmetic formula"
	}
	isValid, errMsg := validateQueryAndFormulaVariable(q.Formula, q.Queries)
	if !isValid {
		return false, errMsg
	}
	for _, query := range q.Queries {
		if !query.IsValid() {
			return false, "Invalid query in query builder"
		}
		if len(query.GroupBy) != 0 || query.GroupByTimestamp != "" {
			return false, "Group by not allowed in derived kpi"
		}
	}
	return true, ""
}

func validateQueryAndFormulaVariable(expression string, queries []KPIQuery) (bool, string) {
	mapOfFormulaVars := make(map[string]bool)
	for _, c := range expression {
		ch := strings.ToLower(string(c))
		if strings.Contains(U.Alpha, strings.ToLower(ch)) {
			mapOfFormulaVars[ch] = true
		}
	}
	mapOfQueryVars := make(map[string]bool)
	for _, query := range queries {
		mapOfQueryVars[query.Name] = true
	}
	if len(mapOfFormulaVars) != len(mapOfQueryVars) {
		return false, "No of formula variables and queries don't match"
	}

	for key := range mapOfFormulaVars {
		if _, exists := mapOfQueryVars[key]; !exists {
			return false, "Please use formula starting from A"
		}
	}
	for key := range mapOfQueryVars {
		if _, exists := mapOfFormulaVars[key]; !exists {
			return false, "Please use formula variables for all queries selected"
		}
	}

	return true, ""
}
func (query *KPIQueryGroup) SetDefaultGroupByTimestamp() {
	for index, _ := range query.Queries {
		defaultGroupByTimestamp := GetDefaultGroupByTimestampForQueries(query.Queries[index].From, query.Queries[index].To, query.Queries[index].GroupByTimestamp)
		if defaultGroupByTimestamp != "" {
			query.Queries[index].GroupByTimestamp = defaultGroupByTimestamp
		}
	}
}

func (query *KPIQueryGroup) GetGroupByTimestamps() []string {
	queryResultString := make([]string, 0)
	for _, intQuery := range query.Queries {
		queryResultString = append(queryResultString, intQuery.GroupByTimestamp)
	}
	return queryResultString
}

func (customKPI *KPIQueryGroup) ContainsNameInInternalTransformation(input string) bool {
	for _, query := range customKPI.Queries {
		for _, metric := range query.Metrics {
			if input == metric {
				return true
			}
		}
	}
	return false
}

func (customKPI *KPIQueryGroup) ValidateFilterAndGroupBy() bool {
	if len(customKPI.GlobalFilters) != 0 || len(customKPI.GlobalGroupBy) != 0 {
		return false
	}
	return true
}

func (customKPI *KPIQueryGroup) ValidateQueries() bool {
	for _, query := range customKPI.Queries {
		if !U.ContainsStringInArray(KpiCategories, query.Category) {
			return false
		}

		// if _, exists := KPIDisplayCategories[query.DisplayCategory]; !exists {
		// 	return false
		// }

	}
	allMetrics := make([]string, 0)

	for _, query := range customKPI.Queries {
		allMetrics = append(allMetrics, query.Metrics...)
	}
	if U.ContainsDuplicate(allMetrics) {
		return false
	}

	return true
}

func transformDateTypeFiltersForKPIFilters(filters []KPIFilter, timezoneString U.TimeZoneString) error {
	for i := range filters {
		err := filters[i].TransformDateTypeFilters(timezoneString)
		if err != nil {
			return err
		}
	}
	return nil
}

type KPIQuery struct {
	Category         string       `json:"ca"`
	DisplayCategory  string       `json:"dc"`
	PageUrl          string       `json:"pgUrl"`
	Metrics          []string     `json:"me"`
	Filters          []KPIFilter  `json:"fil"`
	GroupBy          []KPIGroupBy `json:"gBy"`
	GroupByTimestamp string       `json:"gbt"`
	Timezone         string       `json:"tz"`
	From             int64        `json:"fr"`
	To               int64        `json:"to"`
	Operator         string       `json:"op"`
	QueryType        string       `json:"qt"`
	Name             string       `json:"na"`
}

func (q KPIQuery) GetHashCodeForKPI() (string, error) {
	hashCode := ""
	var err error
	if q.Category == ProfileCategory && q.GroupByTimestamp != "" {
		q.GroupByTimestamp = ""
		hashCode, err = q.GetQueryCacheHashString()
	}
	return hashCode, err
}

func (q *KPIQuery) GetQueryCacheHashString() (string, error) {
	queryMap, err := U.EncodeStructTypeToMap(q)
	if err != nil {
		return "", err
	}

	queryHash, err := U.GenerateHashStringForStruct(queryMap)
	if err != nil {
		return "", err
	}
	return queryHash, nil
}

func (query *KPIQuery) CheckIfNameIsPresent(nameOfQuery string) bool {
	for _, metric := range query.Metrics {
		if metric == nameOfQuery {
			return true
		}
	}
	return false
}

func (query *KPIQuery) IsValid() bool {
	for _, filter := range query.Filters {
		if !filter.IsValid() {
			return false
		}
	}

	for _, groupBy := range query.GroupBy {
		if !groupBy.IsValid() {
			return false
		}
	}

	return true
}

type KPIFilter struct {
	ObjectType       string `json:"objTy"`
	PropertyName     string `json:"prNa"`
	PropertyDataType string `json:"prDaTy"`
	Entity           string `json:"en"`
	Condition        string `json:"co"`
	Value            string `json:"va"`
	LogicalOp        string `json:"lOp"`
}

// Basic type validation.
func (qFilter *KPIFilter) IsValid() bool {
	return !(strings.Contains(qFilter.Entity, " ") || strings.Contains(qFilter.ObjectType, " ") || strings.Contains(qFilter.PropertyName, " "))
}

func (qFilter *KPIFilter) ConvertAllDatesFromTimezone1ToTimzone2(currentTimezone, nextTimezone string) error {
	var dateTimeValue *DateTimePropertyValue
	var err error
	if qFilter.PropertyDataType == U.PropertyTypeDateTime {
		dateTimeValue, err = DecodeDateTimePropertyValue(qFilter.Value)
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
		if qFilter.Condition == BetweenStr || qFilter.Condition == NotInBetweenStr {
			transformedFrom = U.GetStartOfDateEpochInOtherTimezone(transformedFrom, currentTimezone, nextTimezone)
			transformedTo = U.GetEndOfDateEpochInOtherTimezone(transformedTo, currentTimezone, nextTimezone)
		} else if qFilter.Condition == BeforeStr {
			transformedTo = U.GetStartOfDateEpochInOtherTimezone(transformedTo, currentTimezone, nextTimezone)
		} else if qFilter.Condition == SinceStr {
			transformedFrom = U.GetStartOfDateEpochInOtherTimezone(transformedFrom, currentTimezone, nextTimezone)
		}
		dateTimeValue.From = transformedFrom
		dateTimeValue.To = transformedTo
		transformedValue, _ := json.Marshal(dateTimeValue)
		qFilter.Value = string(transformedValue)
	}
	return nil
}

// Duplicate code present between QueryProperty and KPIFilter
func (qp *KPIFilter) TransformDateTypeFilters(timezoneString U.TimeZoneString) error {
	if qp.PropertyDataType == U.PropertyTypeDateTime && (qp.Condition == InLastStr || qp.Condition == NotInLastStr) {
		dateTimeValue, err := DecodeDateTimePropertyValue(qp.Value)
		if err != nil {
			log.WithError(err).Error("Failed reading timestamp on user join query.")
			return err
		}
		lastXthDay := U.GetDateBeforeXPeriod(dateTimeValue.Number, dateTimeValue.Granularity, timezoneString)
		dateTimeValue.From = lastXthDay
		transformedValue, _ := json.Marshal(dateTimeValue)
		qp.Value = string(transformedValue)
	}
	return nil
}

type KPIGroupBy struct {
	ObjectType       string `json:"objTy"`
	PropertyName     string `json:"prNa"`
	PropertyDataType string `json:"prDaTy"`
	GroupByType      string `json:"gbty"`
	Entity           string `json:"en"`
	Granularity      string `json:"grn"`
}

func (kpiGroupBy *KPIGroupBy) IsValid() bool {
	return !(strings.Contains(kpiGroupBy.Entity, " ") || strings.Contains(kpiGroupBy.ObjectType, " ") || strings.Contains(kpiGroupBy.PropertyName, " "))
}

// TODO add check later.
var KPIDisplayCategories = map[string]struct{}{
	WebsiteSessionDisplayCategory:  {},
	PageViewsDisplayCategory:       {},
	FormSubmissionsDisplayCategory: {},

	AllChannelsDisplayCategory:   {},
	GoogleAdsDisplayCategory:     {},
	FacebookDisplayCategory:      {},
	LinkedinDisplayCategory:      {},
	BingAdsDisplayCategory:       {},
	GoogleOrganicDisplayCategory: {},

	HubspotContactsDisplayCategory:  {},
	HubspotCompaniesDisplayCategory: {},
	HubspotDealsDisplayCategory:     {},

	SalesforceUsersDisplayCategory:         {},
	SalesforceAccountsDisplayCategory:      {},
	SalesforceOpportunitiesDisplayCategory: {},

	MarketoLeadsDisplayCategory: {},
}

var MapOfMetricsToData = map[string]map[string]map[string]string{
	WebsiteSessionDisplayCategory: {
		TotalSessions:          {"display_name": "Total Sessions", "object_type": U.EVENT_NAME_SESSION, "type": ""},
		UniqueUsers:            {"display_name": "Unique Users", "object_type": U.EVENT_NAME_SESSION, "type": ""},
		NewUsers:               {"display_name": "New Users", "object_type": U.EVENT_NAME_SESSION, "type": ""},
		RepeatUsers:            {"display_name": "Repeat Users", "object_type": U.EVENT_NAME_SESSION, "type": ""},
		SessionsPerUser:        {"display_name": "Session Per User", "object_type": U.EVENT_NAME_SESSION, "type": ""},
		EngagedSessions:        {"display_name": "Engaged Sessions", "object_type": U.EVENT_NAME_SESSION, "type": ""},
		EngagedUsers:           {"display_name": "Engaged Users", "object_type": U.EVENT_NAME_SESSION, "type": ""},
		EngagedSessionsPerUser: {"display_name": "Engaged Sessions per user", "object_type": U.EVENT_NAME_SESSION, "type": ""},
		TotalTimeOnSite:        {"display_name": "Total time on site", "object_type": U.EVENT_NAME_SESSION, "type": MetricsDateType},
		AvgSessionDuration:     {"display_name": "Avg session duration", "object_type": U.EVENT_NAME_SESSION, "type": MetricsDateType},
		AvgPageViewsPerSession: {"display_name": "Avg page views per session", "object_type": U.EVENT_NAME_SESSION, "type": ""},
		AvgInitialPageLoadTime: {"display_name": "Avg initial page load time", "object_type": U.EVENT_NAME_SESSION, "type": MetricsDateType},
		BounceRate:             {"display_name": "Bounce rate", "object_type": U.EVENT_NAME_SESSION, "type": MetricsPercentageType},
		EngagementRate:         {"display_name": "Engagement rate", "object_type": U.EVENT_NAME_SESSION, "type": MetricsPercentageType},
	},
	PageViewsDisplayCategory: {
		Entrances:                {"display_name": "Entrances", "object_type": U.EVENT_NAME_SESSION, "type": ""},
		Exits:                    {"display_name": "Exits", "object_type": U.EVENT_NAME_SESSION, "type": ""},
		PageViews:                {"display_name": "Page Views", "type": ""},
		UniqueUsers:              {"display_name": "Unique users", "type": ""},
		PageviewsPerUser:         {"display_name": "Page views per user", "type": ""},
		AvgPageLoadTime:          {"display_name": "Avg page load time", "type": MetricsDateType},
		AvgVerticalScrollPercent: {"display_name": "Avg vertical scroll percent", "type": MetricsPercentageType},
		AvgTimeOnPage:            {"display_name": "Avg time on page", "type": MetricsDateType},
		EngagedPageViews:         {"display_name": "Engaged page views", "type": ""},
		EngagedUsers:             {"display_name": "Engaged Users", "type": ""},
		EngagementRate:           {"display_name": "Engagement rate", "type": MetricsPercentageType},
	},
	FormSubmissionsDisplayCategory: {
		Count:        {"display_name": "Count", "object_type": U.EVENT_NAME_FORM_SUBMITTED, "type": ""},
		UniqueUsers:  {"display_name": "Unique users", "object_type": U.EVENT_NAME_FORM_SUBMITTED, "type": ""},
		CountPerUser: {"display_name": "Count per user", "object_type": U.EVENT_NAME_FORM_SUBMITTED, "type": ""},
	},
	AllChannelsDisplayCategory: {
		"impressions": {"display_name": "Impressions", "type": ""},
		"clicks":      {"display_name": "Clicks", "type": ""},
		"spend":       {"display_name": "Spend", "type": ""},
	},
	GoogleAdsDisplayCategory: {
		Conversion:                                 {"display_name": "Conversion", "type": ""},
		ClickThroughRate:                           {"display_name": "Click through rate", "type": ""},
		ConversionRate:                             {"display_name": "Conversion rate", "type": ""},
		CostPerClick:                               {"display_name": "Cost per click", "type": ""},
		CostPerConversion:                          {"display_name": "Cost per conversion", "type": ""},
		SearchImpressionShare:                      {"display_name": "Search Impr. share", "type": ""},
		SearchClickShare:                           {"display_name": "Search click share", "type": ""},
		SearchTopImpressionShare:                   {"display_name": "Search top Impr. share", "type": ""},
		SearchAbsoluteTopImpressionShare:           {"display_name": "Search abs. top Impr. share", "type": ""},
		SearchBudgetLostAbsoluteTopImpressionShare: {"display_name": "Search budget lost abs top impr. share", "type": ""},
		SearchBudgetLostImpressionShare:            {"display_name": "Search budget lost Impr. share", "type": ""},
		SearchBudgetLostTopImpressionShare:         {"display_name": "Search budget lost top Impr. share", "type": ""},
		SearchRankLostAbsoluteTopImpressionShare:   {"display_name": "Search rank lost abs. top Impr. share", "type": ""},
		SearchRankLostImpressionShare:              {"display_name": "Search rank lost Impr. share", "type": ""},
		SearchRankLostTopImpressionShare:           {"display_name": "Search rank lost top Impr. share", "type": ""},
		ConversionValue:                            {"display_name": "Conversion Value", "type": ""},
	},
	FacebookDisplayCategory: {
		"video_p50_watched_actions":              {"display_name": "Video p50 watched actions", "type": ""},
		"video_p25_watched_actions":              {"display_name": "Video p25 watched actions", "type": ""},
		"video_30_sec_watched_actions":           {"display_name": "Video 30 sec watched actions", "type": ""},
		"video_p100_watched_actions":             {"display_name": "Video p100 watched actions", "type": ""},
		"video_p75_watched_actions":              {"display_name": "Video p75 watched actions", "type": ""},
		"cost_per_click":                         {"display_name": "Cost per click", "type": ""},
		"cost_per_link_click":                    {"display_name": "Cost per link click", "type": ""},
		"cost_per_thousand_impressions":          {"display_name": "Cost per thousand impressions", "type": ""},
		"click_through_rate":                     {"display_name": "Click through rate", "type": ""},
		"link_click_through_rate":                {"display_name": "Link click through rate", "type": ""},
		"link_clicks":                            {"display_name": "Link clicks", "type": ""},
		"frequency":                              {"display_name": "frequency", "type": ""},
		"reach":                                  {"display_name": "reach", "type": ""},
		"fb_pixel_purchase_count":                {"display_name": "Offsite Purchase (Count)", "type": ""},
		"fb_pixel_purchase_revenue":              {"display_name": "Offsite Purchase (Revenue)", "type": ""},
		"fb_pixel_purchase_cost_per_action_type": {"display_name": "Cost Per Offsite Purchase", "type": ""},
		"fb_pixel_purchase_roas":                 {"display_name": "Offsite Purchase (ROAS)", "type": ""},
	},
	BingAdsDisplayCategory: {
		Conversions: {"display_name": "Conversions", "type": ""},
	},
	GoogleOrganicDisplayCategory: {
		Impressions:                        {"display_name": "Impressions", "type": ""},
		Clicks:                             {"display_name": "Clicks", "type": ""},
		ClickThroughRate:                   {"display_name": "Click through rate", "type": ""},
		"position_avg":                     {"display_name": "Position Avg", "type": ""},
		"position_impression_weighted_avg": {"display_name": "Position Impression weighted Avg", "type": ""},
	},
	CustomAdsDisplayCategory: {},
}

type TransformQueryi struct {
	Metrics KpiToEventMetricRepr
	Filters []QueryProperty
}

type KpiToEventMetricRepr struct {
	Aggregation string
	Entity      string
	Property    string
	Operator    string
	GroupByType string
}

// Util/Common Methods.
func GetStaticallyDefinedMetricsForDisplayCategory(category string) []map[string]string {
	resultMetrics := []map[string]string{}
	mapOfMetricsToData := MapOfMetricsToData[category]
	for metricName, data := range mapOfMetricsToData {
		currentMetrics := map[string]string{}
		currentMetrics["name"] = metricName
		currentMetrics["display_name"] = data["display_name"]
		currentMetrics["type"] = data["type"]
		currentMetrics["kpi_query_type"] = KpiStaticQueryType
		resultMetrics = append(resultMetrics, currentMetrics)
	}
	return resultMetrics
}

func AddObjectTypeToProperties(kpiConfig map[string]interface{}, value string) map[string]interface{} {
	properties := kpiConfig["properties"].([]map[string]string)
	for index := range properties {
		properties[index]["object_type"] = value
	}
	kpiConfig["properties"] = properties
	return kpiConfig
}

func TransformCRMPropertiesToKPIConfigProperties(properties map[string][]string, propertiesToDisplayNames map[string]string, prefix string) []map[string]string {
	var resultantKPIConfigProperties []map[string]string
	var tempKPIConfigProperty map[string]string
	for dataType, propertyNames := range properties {
		for _, propertyName := range propertyNames {
			if strings.HasPrefix(propertyName, prefix) {
				var displayName string
				displayName, exists := propertiesToDisplayNames[propertyName]
				if !exists {
					displayName = U.CreateVirtualDisplayName(propertyName)
				}
				tempKPIConfigProperty = map[string]string{
					"name":         propertyName,
					"display_name": displayName,
					"data_type":    dataType,
					"entity":       UserEntity,
				}
				resultantKPIConfigProperties = append(resultantKPIConfigProperties, tempKPIConfigProperty)
			}
		}
	}
	if resultantKPIConfigProperties == nil {
		return make([]map[string]string, 0)
	}
	return resultantKPIConfigProperties
}

func ValidateKPIQueryMetricsForAnyEventType(kpiQueryMetrics []string, mapOfMetrics map[string]map[string]string) bool {
	for _, metric := range kpiQueryMetrics {
		if _, exists := mapOfMetrics[metric]; !exists {
			return false
		}
	}
	return true
}

// Duplicated function present in postgres and memsql/kpi_analytics_website_session.
func ValidateKPIQueryFiltersForAnyEventType(kpiQueryFilters []KPIFilter, configPropertiesData []map[string]string) bool {
	mapOfPropertyName := make(map[string]struct{})
	for _, propertyData := range configPropertiesData {
		mapOfPropertyName[propertyData["name"]] = struct{}{}
	}
	for _, filter := range kpiQueryFilters {
		if _, exists := mapOfPropertyName[filter.PropertyName]; !exists {
			return false
		}
	}
	return true
}

// Duplicated function present in postgres and memsql/kpi_analytics_website_session.
func ValidateKPIQueryGroupByForAnyEventType(kpiQueryGroupBys []KPIGroupBy, configPropertiesData []map[string]string) bool {
	mapOfPropertyName := make(map[string]struct{})
	for _, propertyData := range configPropertiesData {
		mapOfPropertyName[propertyData["name"]] = struct{}{}
	}
	for _, groupBy := range kpiQueryGroupBys {
		if _, exists := mapOfPropertyName[groupBy.PropertyName]; !exists {
			return false
		}
	}
	return true
}

func GetNonGBTResultsFromGBTResultsAndMaps(reqID string, kpiQueryGroup KPIQueryGroup, mapOfNonGBTDerivedKPIToInternalKPIToResults map[string]map[string][]QueryResult,
	mapOfGBTDerivedKPIToInternalKPIToResults map[string]map[string][]QueryResult, mapOfNonGBTKPINormalQueryToResults map[string][]QueryResult,
	mapOfGBTKPINormalQueryToResults map[string][]QueryResult, externalQueryToInternalQueries map[string]KPIQueryGroup) (
	int, map[string]map[string][]QueryResult, map[string][]QueryResult) {

	finalStatusCode := http.StatusOK
	logEntry := log.WithField("reqID", reqID).
		WithField("kpiQueryGroup", kpiQueryGroup).
		WithField("mapOfNonGBTDerivedKPIToInternalKPIToResults", mapOfNonGBTDerivedKPIToInternalKPIToResults).
		WithField("mapOfGBTDerivedKPIToInternalKPIToResults", mapOfGBTDerivedKPIToInternalKPIToResults).
		WithField("mapOfNonGBTKPINormalQueryToResults", mapOfNonGBTKPINormalQueryToResults).
		WithField("mapOfGBTKPINormalQueryToResults", mapOfGBTKPINormalQueryToResults).
		WithField("externalQueryToInternalQueries", externalQueryToInternalQueries)
	for _, externalQuery := range kpiQueryGroup.Queries {
		if externalQuery.QueryType == Derived {
			if externalQuery.GroupByTimestamp == "" {
				externalQueryHashCode, err := externalQuery.GetQueryCacheHashString()
				logEntry = logEntry.WithField("externalQuery", externalQuery).WithField("externalQueryHashCode", externalQueryHashCode)
				if err != nil {
					logEntry.Warn("Hash string not found 1:" + err.Error())
					finalStatusCode = http.StatusInternalServerError
					break
				}

				internalKPIQuery, exists := externalQueryToInternalQueries[externalQueryHashCode]
				if !exists {
					logEntry.Warn("Hash code not found in hash map 1")
					finalStatusCode = http.StatusInternalServerError
					break
				}
				for _, internalQuery := range internalKPIQuery.Queries {
					logEntry = logEntry.WithField("internalQuery", internalQuery)
					if internalQuery.Category == ProfileCategory {
						mapOfInternalKPIToResults, exists := mapOfGBTDerivedKPIToInternalKPIToResults[externalQueryHashCode]
						logEntry = logEntry.WithField("mapOfInternalKPIToResults", mapOfInternalKPIToResults)
						if !exists {
							logEntry.Warn("Hash code not found in hash map 2")
							finalStatusCode = http.StatusInternalServerError
							break
						}
						nonGBTResults, err := GetNonGBTResultsFromGBTResults(mapOfInternalKPIToResults, internalQuery)
						logEntry = logEntry.WithField("nonGBTResults", nonGBTResults)
						if err != "" {
							logEntry.Warn("Error in getting non GBT from GBT 1: " + err)
							finalStatusCode = http.StatusInternalServerError
							break
						}

						internalQueryHashCode, _ := internalQuery.GetQueryCacheHashString()
						logEntry = logEntry.WithField("internalQueryHashCode", internalQueryHashCode)

						mapOfNonGBTDerivedKPIToInternalKPIToResults[externalQueryHashCode][internalQueryHashCode] = nonGBTResults
					}
				}
			}
		} else {
			if externalQuery.GroupByTimestamp == "" {
				queryHashCode, _ := externalQuery.GetQueryCacheHashString()
				logEntry = logEntry.WithField("externalQuery", externalQuery).WithField("queryHashCode", queryHashCode)
				if externalQuery.Category == ProfileCategory {
					nonGBTResults, err := GetNonGBTResultsFromGBTResults(mapOfGBTKPINormalQueryToResults, externalQuery)
					logEntry = logEntry.WithField("nonGBTResults", nonGBTResults)
					if err != "" {
						logEntry.Warn("Error in getting non GBT from GBT 2: " + err)
						finalStatusCode = http.StatusInternalServerError
						break
					}

					mapOfNonGBTKPINormalQueryToResults[queryHashCode] = nonGBTResults
				}
			}
		}
	}
	return finalStatusCode, mapOfNonGBTDerivedKPIToInternalKPIToResults, mapOfNonGBTKPINormalQueryToResults
}

// Below function relies on fact that each query has only one metric.
func GetNonGBTResultsFromGBTResults(hashMapOfQueryToResult map[string][]QueryResult, query KPIQuery) ([]QueryResult, string) {
	finalResultantQueryResults := make([]QueryResult, 0, 0)

	hashCode, err := query.GetQueryCacheHashString()
	if err != nil {
		return []QueryResult{{}, {}}, "Failed while generating hashString for kpi 2."
	}

	if resultsWithGbt, exists := hashMapOfQueryToResult[hashCode]; exists {
		for _, queryResult := range resultsWithGbt {
			resultAsMap := make(map[string][]interface{})
			currentResultantRows := make([][]interface{}, 0, 0)
			currentQueryResult := QueryResult{}

			for _, row := range queryResult.Rows {
				currentRow := getRowAfterDeletionOfDateTime(row, queryResult.Headers)
				key := getKeyWithoutDateTime(currentRow)

				if val, ok := resultAsMap[key]; ok {
					totalValue, err := U.SafeAddition(val[len(currentRow)-1], currentRow[len(currentRow)-1])
					if err != nil {
						resultAsMap = make(map[string][]interface{})
						break
					}
					currentRow[len(currentRow)-1] = totalValue
					resultAsMap[key] = currentRow
				} else {
					resultAsMap[key] = currentRow
				}
			}

			for _, val := range resultAsMap {
				currentResultantRows = append(currentResultantRows, val)
			}
			currentQueryResult.Rows = currentResultantRows
			currentQueryResult.Headers = U.RemoveElementFromArray(queryResult.Headers, AliasDateTime)
			finalResultantQueryResults = append(finalResultantQueryResults, currentQueryResult)
		}
	} else {
		return []QueryResult{{}, {}}, "Query group doesnt contain all the gbt and non gbt pair of query."
	}
	return finalResultantQueryResults, ""
}

func GetFinalResultantResultsForKPI(reqID string, kpiQueryGroup KPIQueryGroup, mapOfNonGBTDerivedKPIToInternalKPIToResults map[string]map[string][]QueryResult,
	mapOfGBTDerivedKPIToInternalKPIToResults map[string]map[string][]QueryResult, mapOfNonGBTKPINormalQueryToResults map[string][]QueryResult,
	mapOfGBTKPINormalQueryToResults map[string][]QueryResult, externalQueryToInternalQueries map[string]KPIQueryGroup) ([]QueryResult, int) {

	finalResultantResults := make([]QueryResult, 0)
	finalStatusCode := http.StatusOK
	logEntry := log.WithField("reqID", reqID).
		WithField("kpiQueryGroup", kpiQueryGroup).
		WithField("mapOfNonGBTDerivedKPIToInternalKPIToResults", mapOfNonGBTDerivedKPIToInternalKPIToResults).
		WithField("mapOfGBTDerivedKPIToInternalKPIToResults", mapOfGBTDerivedKPIToInternalKPIToResults).
		WithField("mapOfNonGBTKPINormalQueryToResults", mapOfNonGBTKPINormalQueryToResults).
		WithField("mapOfGBTKPINormalQueryToResults", mapOfGBTKPINormalQueryToResults).
		WithField("externalQueryToInternalQueries", externalQueryToInternalQueries)

	for _, externalQuery := range kpiQueryGroup.Queries {
		results := make([]QueryResult, 0)
		groupByTimestamp := externalQuery.GroupByTimestamp
		externalQuery.GroupByTimestamp = ""
		externalQueryHashCode, err := externalQuery.GetQueryCacheHashString()
		logEntry = logEntry.WithField("externalQuery", externalQuery).WithField("externalQueryHashCode", externalQueryHashCode)
		if err != nil {
			logEntry.Warn("Hash string not found 3: " + err.Error())
			finalStatusCode = http.StatusInternalServerError
			break
		}
		if externalQuery.QueryType == Derived {

			mapOfFormulaVariableToQueryResult := make(map[string]QueryResult)
			internalKPIQuery, exists := externalQueryToInternalQueries[externalQueryHashCode]
			if !exists {
				logEntry.Warn("Hash code not found in hash map 3")
				finalStatusCode = http.StatusInternalServerError
				break
			}

			for _, internalQuery := range internalKPIQuery.Queries {
				var queryResult QueryResult
				if groupByTimestamp == "" {

					internalQueryHashCode, err := internalQuery.GetQueryCacheHashString()
					logEntry = logEntry.WithField("internalQueryHashCode", internalQueryHashCode)
					if err != nil {
						logEntry.Warn("Hash string found in hash map 4: " + err.Error())
						finalStatusCode = http.StatusInternalServerError
						break
					}
					nonGBTResults, exists := mapOfNonGBTDerivedKPIToInternalKPIToResults[externalQueryHashCode][internalQueryHashCode]
					logEntry = logEntry.WithField("nonGBTResults", nonGBTResults)
					if !exists {
						logEntry.Warn("Hash code not found in hash map 4")
						finalStatusCode = http.StatusInternalServerError
						break
					}
					queryResult = nonGBTResults[0]

				} else {

					internalQuery.GroupByTimestamp = ""
					internalQueryHashCode, err := internalQuery.GetQueryCacheHashString()
					logEntry = logEntry.WithField("internalQueryHashCode", internalQueryHashCode)
					if err != nil {
						logEntry.Warn("Hash string found in hash map 5: " + err.Error())
						finalStatusCode = http.StatusInternalServerError
						break
					}
					gbTResults, exists := mapOfGBTDerivedKPIToInternalKPIToResults[externalQueryHashCode][internalQueryHashCode]
					if !exists {
						logEntry.Warn("Hash code found in hash map 5: " + err.Error())
						finalStatusCode = http.StatusInternalServerError
						break
					}
					queryResult = gbTResults[0]
				}

				mapOfFormulaVariableToQueryResult[internalQuery.Name] = queryResult
			}

			finalResultantResults = append(finalResultantResults, EvaluateKPIExpressionWithBraces(mapOfFormulaVariableToQueryResult, externalQuery.Timezone, strings.ToLower(internalKPIQuery.Formula)))

		} else {
			if groupByTimestamp == "" {
				var exists bool
				results, exists = mapOfNonGBTKPINormalQueryToResults[externalQueryHashCode]
				if !exists {
					logEntry.Warn("Hash code not found in hash map 6")
					finalStatusCode = http.StatusInternalServerError
					break
				}
			} else {
				var exists bool
				results, exists = mapOfGBTKPINormalQueryToResults[externalQueryHashCode]
				if !exists {
					logEntry.Warn("Hash code not found in hash map 7")
					finalStatusCode = http.StatusInternalServerError
					break
				}
			}
			finalResultantResults = append(finalResultantResults, results...)
		}
	}

	return finalResultantResults, finalStatusCode
}

func EvaluateKPIExpressionWithBraces(mapOfFormulaVariableToQueryResult map[string]QueryResult, timezone string, formula string) QueryResult {
	valueStack := make([]QueryResult, 0)
	operatorStack := make([]string, 0)

	for _, currentVariable := range formula {
		currentFormulaVariable := string(currentVariable)
		if currentFormulaVariable == "(" {
			operatorStack = append(operatorStack, currentFormulaVariable)
		} else if strings.Contains(U.Alpha, strings.ToLower(currentFormulaVariable)) {
			valueStack = append(valueStack, mapOfFormulaVariableToQueryResult[currentFormulaVariable])
		} else if currentFormulaVariable == ")" {
			for len(operatorStack) != 0 && operatorStack[len(operatorStack)-1] != "(" {
				v1 := valueStack[len(valueStack)-1]
				valueStack = valueStack[:len(valueStack)-1]
				v2 := valueStack[len(valueStack)-1]
				valueStack = valueStack[:len(valueStack)-1]
				results := make([]*QueryResult, 0)
				results = append(results, &v2)
				results = append(results, &v1)
				op := operatorStack[len(operatorStack)-1]
				ops := make([]string, 0)
				ops = append(ops, op)
				operatorStack = operatorStack[:len(operatorStack)-1]
				valueStack = append(valueStack, HandlingEventResultsByApplyingOperations(results, ops, timezone)) // apply operations and return result
			}
			if len(operatorStack) != 0 {
				operatorStack = operatorStack[:len(operatorStack)-1]
			}
		} else {
			for len(operatorStack) != 0 && U.Precedence(operatorStack[len(operatorStack)-1]) >= U.Precedence(currentFormulaVariable) {
				v1 := valueStack[len(valueStack)-1]
				valueStack = valueStack[:len(valueStack)-1]
				v2 := valueStack[len(valueStack)-1]
				results := make([]*QueryResult, 0)
				results = append(results, &v2)
				results = append(results, &v1)
				valueStack = valueStack[:len(valueStack)-1]
				op := operatorStack[len(operatorStack)-1]
				ops := make([]string, 0)
				ops = append(ops, op)
				operatorStack = operatorStack[:len(operatorStack)-1]
				valueStack = append(valueStack, HandlingEventResultsByApplyingOperations(results, ops, timezone)) // apply operations and return result
			}
			operatorStack = append(operatorStack, currentFormulaVariable)
		}
	}

	for len(operatorStack) != 0 {
		v1 := valueStack[len(valueStack)-1]
		valueStack = valueStack[:len(valueStack)-1]
		v2 := valueStack[len(valueStack)-1]
		results := make([]*QueryResult, 0)
		results = append(results, &v2)
		results = append(results, &v1)
		valueStack = valueStack[:len(valueStack)-1]
		op := operatorStack[len(operatorStack)-1]
		ops := make([]string, 0)
		ops = append(ops, op)
		operatorStack = operatorStack[:len(operatorStack)-1]
		valueStack = append(valueStack, HandlingEventResultsByApplyingOperations(results, ops, timezone))
	}
	return valueStack[len(valueStack)-1]
}

func getRowAfterDeletionOfDateTime(row []interface{}, headers []string) []interface{} {
	dateIndex := -1
	for index, header := range headers {
		if header == AliasDateTime {
			dateIndex = index
		}
	}
	if dateIndex == -1 {
		return row
	}

	finalResultantRow := make([]interface{}, 0)
	for index, element := range row {
		if index != dateIndex {
			finalResultantRow = append(finalResultantRow, element)
		}
	}
	return finalResultantRow
}

func getKeyWithoutDateTime(row []interface{}) string {
	if len(row) <= 1 {
		return "1"
	}
	var key string
	for _, value := range row[:len(row)-1] {
		if _, ok := (value.(time.Time)); !ok {
			key = key + fmt.Sprintf("%v", value) + ":;"
		}
	}

	return key
}

func SplitQueryResultsIntoGBTAndNonGBT(queryResults []QueryResult, kpiQueryGroup KPIQueryGroup, finalStatusCode int) ([]QueryResult, []QueryResult, []KPIQuery, []KPIQuery) {
	gbtRelatedQueryResults := make([]QueryResult, 0)
	gbtRelatedQueries := make([]KPIQuery, 0)
	nonGbtRelatedQueryResults := make([]QueryResult, 0)
	nonGbtRelatedQueries := make([]KPIQuery, 0)
	for index, kpiQuery := range kpiQueryGroup.Queries {
		if kpiQuery.GroupByTimestamp != "" {
			gbtRelatedQueryResults = append(gbtRelatedQueryResults, queryResults[index])
			gbtRelatedQueries = append(gbtRelatedQueries, kpiQuery)
		} else {
			nonGbtRelatedQueryResults = append(nonGbtRelatedQueryResults, queryResults[index])
			nonGbtRelatedQueries = append(nonGbtRelatedQueries, kpiQuery)
		}
	}
	if len(nonGbtRelatedQueries) == 0 {
		nonGbtRelatedQueryResults = nil
	}
	if len(gbtRelatedQueries) == 0 {
		gbtRelatedQueryResults = nil
	}
	return gbtRelatedQueryResults, nonGbtRelatedQueryResults, gbtRelatedQueries, nonGbtRelatedQueries
}

func MergeQueryResults(queryResults []QueryResult, queries []KPIQuery, timezoneString string, finalStatusCode int) QueryResult {
	if finalStatusCode != http.StatusOK || len(queryResults) == 0 {
		queryResult := QueryResult{}
		return queryResult
	}

	queryResult := QueryResult{}
	queryResult.Headers = TransformColumnResultGroup(queryResults, queries, timezoneString)
	queryResult.Rows = TransformRowsResultGroup(queryResults, timezoneString)
	return queryResult
}

// NOTE: Basing on single metric being sent per query.
func TransformColumnResultGroup(queryResults []QueryResult, queries []KPIQuery, timezoneString string) []string {
	finalResultantColumns := make([]string, 0)
	for index, queryResult := range queryResults {
		resultantMetrics := make([]string, 0)
		if queries[index].Category == ChannelCategory {
			for _, metric := range queries[index].Metrics {
				resultantMetrics = append(resultantMetrics, queries[index].DisplayCategory+"_"+metric)
			}
		} else {
			resultantMetrics = queries[index].Metrics
		}
		if index == 0 {
			finalResultantColumns = append(queryResult.Headers[:len(queryResult.Headers)-1], resultantMetrics...)
		} else {
			finalResultantColumns = append(finalResultantColumns, resultantMetrics...)
		}
	}
	return finalResultantColumns
}

// Form Map with key as combination of columns and values.
// Steps involved are as follows.
// 1. Make an empty hashMap with key and value as array of 0's as prefixed values.
// 2. Add the values to hashMap. Here keys are contextual to kpi and will not be duplicate.
// 3. Convert Map to 2d Array and then sort.
func TransformRowsResultGroup(queryResults []QueryResult, timezoneString string) [][]interface{} {
	resultAsMap := GetResultAsMap(queryResults)

	currentResultantRows := make([][]interface{}, 0, 0)
	for key, value := range resultAsMap {
		currentRow := SplitKeysAndGetRow(key, timezoneString)
		currentRow = append(currentRow, value...)
		currentResultantRows = append(currentResultantRows, currentRow)
	}
	currentResultantRows = U.GetSorted2DArrays(currentResultantRows)
	return currentResultantRows
}

func GetResultAsMap(queryResults []QueryResult) map[string][]interface{} {
	resultAsMap := make(map[string][]interface{})
	numberOfQueryResults := len(queryResults)

	// Step 1
	for _, queryResult := range queryResults {
		for _, row := range queryResult.Rows {
			key := U.GetkeyFromRow(row)
			emptyValues := make([]interface{}, numberOfQueryResults)
			for index := range emptyValues {
				emptyValues[index] = 0
			}
			resultAsMap[key] = emptyValues
		}
	}

	// Step 2
	for queryIndex, queryResult := range queryResults {
		for _, row := range queryResult.Rows {
			key := U.GetkeyFromRow(row)
			val := row[len(row)-1]
			resultAsMap[key][queryIndex] = val
		}
	}
	return resultAsMap
}

func SplitKeysAndGetRow(key string, timezoneString string) []interface{} {
	currentRow := make([]interface{}, 0)
	columns := strings.Split(key, ":;")
	for _, column := range columns[:len(columns)-1] {
		if strings.HasPrefix(column, "dat$") {
			unixValue, _ := strconv.ParseInt(strings.TrimPrefix(column, "dat$"), 10, 64)
			columnValue, _ := U.GetTimeFromUnixTimestampWithZone(unixValue, timezoneString)
			currentRow = append(currentRow, columnValue)
		} else {
			currentRow = append(currentRow, column)
		}
	}
	return currentRow
}
