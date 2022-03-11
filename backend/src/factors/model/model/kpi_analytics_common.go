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

type KPIQueryGroup struct {
	Class         string       `json:"cl"`
	Queries       []KPIQuery   `json:"qG"`
	GlobalFilters []KPIFilter  `json:"gFil"`
	GlobalGroupBy []KPIGroupBy `json:"gGBy"`
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

func (q *KPIQueryGroup) GetQueryCacheRedisKey(projectID uint64) (*cacheRedis.Key, error) {
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

type KPIFilter struct {
	ObjectType       string `json:"objTy"`
	PropertyName     string `json:"prNa"`
	PropertyDataType string `json:"prDaTy"`
	Entity           string `json:"en"`
	Condition        string `json:"co"`
	Value            string `json:"va"`
	LogicalOp        string `json:"lOp"`
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

var MapOfMetricsToData = map[string]map[string]map[string]string{
	WebsiteSessionDisplayCategory: {
		TotalSessions:          {"display_name": "Total Sessions", "object_type": U.EVENT_NAME_SESSION},
		UniqueUsers:            {"display_name": "Unique Users", "object_type": U.EVENT_NAME_SESSION},
		NewUsers:               {"display_name": "New Users", "object_type": U.EVENT_NAME_SESSION},
		RepeatUsers:            {"display_name": "Repeat Users", "object_type": U.EVENT_NAME_SESSION},
		SessionsPerUser:        {"display_name": "Session Per User", "object_type": U.EVENT_NAME_SESSION},
		EngagedSessions:        {"display_name": "Engaged Sessions", "object_type": U.EVENT_NAME_SESSION},
		EngagedUsers:           {"display_name": "Engaged Users", "object_type": U.EVENT_NAME_SESSION},
		EngagedSessionsPerUser: {"display_name": "Engaged Sessions per user", "object_type": U.EVENT_NAME_SESSION},
		TotalTimeOnSite:        {"display_name": "Total time on site", "object_type": U.EVENT_NAME_SESSION},
		AvgSessionDuration:     {"display_name": "Avg session duration", "object_type": U.EVENT_NAME_SESSION},
		AvgPageViewsPerSession: {"display_name": "Avg page views per session", "object_type": U.EVENT_NAME_SESSION},
		AvgInitialPageLoadTime: {"display_name": "Avg initial page load time", "object_type": U.EVENT_NAME_SESSION},
		BounceRate:             {"display_name": "Bounce rate", "object_type": U.EVENT_NAME_SESSION},
		EngagementRate:         {"display_name": "Engagement rate", "object_type": U.EVENT_NAME_SESSION},
	},
	PageViewsDisplayCategory: {
		Entrances:                {"display_name": "Entrances", "object_type": U.EVENT_NAME_SESSION},
		Exits:                    {"display_name": "Exits", "object_type": U.EVENT_NAME_SESSION},
		PageViews:                {"display_name": "Page Views"},
		UniqueUsers:              {"display_name": "Unique users"},
		PageviewsPerUser:         {"display_name": "Page views per user"},
		AvgPageLoadTime:          {"display_name": "Avg page load time"},
		AvgVerticalScrollPercent: {"display_name": "Avg vertical scroll percent"},
		AvgTimeOnPage:            {"display_name": "Avg time on page"},
		EngagedPageViews:         {"display_name": "Engaged page views"},
		EngagedUsers:             {"display_name": "Engaged Users"},
		EngagementRate:           {"display_name": "Engagement rate"},
	},
	FormSubmissionsDisplayCategory: {
		Count:        {"display_name": "Count", "object_type": U.EVENT_NAME_FORM_SUBMITTED},
		UniqueUsers:  {"display_name": "Unique users", "object_type": U.EVENT_NAME_FORM_SUBMITTED},
		CountPerUser: {"display_name": "Count per user", "object_type": U.EVENT_NAME_FORM_SUBMITTED},
	},
	AllChannelsDisplayCategory: {
		"impressions": {"display_name": "Impressions"},
		"clicks":      {"display_name": "Clicks"},
		"spend":       {"display_name": "Spend"},
	},
	GoogleAdsDisplayCategory: {
		Conversion:                                 {"display_name": "Conversion"},
		ClickThroughRate:                           {"display_name": "Click through rate"},
		ConversionRate:                             {"display_name": "Conversion rate"},
		CostPerClick:                               {"display_name": "Cost per click"},
		CostPerConversion:                          {"display_name": "Cost per conversion"},
		SearchImpressionShare:                      {"display_name": "Search Impr. share"},
		SearchClickShare:                           {"display_name": "Search click share"},
		SearchTopImpressionShare:                   {"display_name": "Search top Impr. share"},
		SearchAbsoluteTopImpressionShare:           {"display_name": "Search abs. top Impr. share"},
		SearchBudgetLostAbsoluteTopImpressionShare: {"display_name": "Search budget lost abs top impr. share"},
		SearchBudgetLostImpressionShare:            {"display_name": "Search budget lost Impr. share"},
		SearchBudgetLostTopImpressionShare:         {"display_name": "Search budget lost top Impr. share"},
		SearchRankLostAbsoluteTopImpressionShare:   {"display_name": "Search rank lost abs. top Impr. share"},
		SearchRankLostImpressionShare:              {"display_name": "Search rank lost Impr. share"},
		SearchRankLostTopImpressionShare:           {"display_name": "Search rank lost top Impr. share"},
	},
	FacebookDisplayCategory: {
		"video_p50_watched_actions":     {"display_name": "Video p50 watched actions"},
		"video_p25_watched_actions":     {"display_name": "Video p25 watched actions"},
		"video_30_sec_watched_actions":  {"display_name": "Video 30 sec watched actions"},
		"video_p100_watched_actions":    {"display_name": "Video p100 watched actions"},
		"video_p75_watched_actions":     {"display_name": "Video p75 watched actions"},
		"cost_per_click":                {"display_name": "Cost per click"},
		"cost_per_link_click":           {"display_name": "Cost per link click"},
		"cost_per_thousand_impressions": {"display_name": "Cost per thousand impressions"},
		"click_through_rate":            {"display_name": "Click through rate"},
		"link_click_through_rate":       {"display_name": "Link click through rate"},
		"link_clicks":                   {"display_name": "Link clicks"},
		"frequency":                     {"display_name": "frequency"},
		"reach":                         {"display_name": "reach"},
	},
	BingAdsDisplayCategory: {
		Conversions: {"display_name": "Conversions"},
	},
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
func GetMetricsForDisplayCategory(category string) []map[string]string {
	resultMetrics := []map[string]string{}
	mapOfMetricsToData := MapOfMetricsToData[category]
	for metricName, data := range mapOfMetricsToData {
		currentMetrics := map[string]string{}
		currentMetrics["name"] = metricName
		currentMetrics["display_name"] = data["display_name"]
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
					displayName = propertyName
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

// Below function relies on fact that each query has only one metric.
func GetNonGBTResultsFromGBTResults(queryResults []QueryResult, query KPIQuery) []QueryResult {
	finalResultantQueryResults := make([]QueryResult, 0, 0)

	for _, queryResult := range queryResults {
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

	return finalResultantQueryResults
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

func MergeQueryResults(queryResults []QueryResult, queries []KPIQuery, timezoneString string, finalStatusCode int, isTimezoneEnabled bool) QueryResult {
	if finalStatusCode != http.StatusOK || len(queryResults) == 0 {
		queryResult := QueryResult{}
		return queryResult
	}

	queryResult := QueryResult{}
	queryResult.Headers = TransformColumnResultGroup(queryResults, queries, timezoneString)
	queryResult.Rows = TransformRowsResultGroup(queryResults, timezoneString, isTimezoneEnabled)
	return queryResult
}

// NOTE: Basing on single metric being sent per query.
func TransformColumnResultGroup(queryResults []QueryResult, queries []KPIQuery, timezoneString string) []string {
	finalResultantColumns := make([]string, 0)
	for index, queryResult := range queryResults {
		if index == 0 {
			finalResultantColumns = append(queryResult.Headers[:len(queryResult.Headers)-1], queries[index].Metrics...)
		} else {
			finalResultantColumns = append(finalResultantColumns, queries[index].Metrics...)
		}
	}
	return finalResultantColumns
}

// Form Map with key as combination of columns and values.
// Steps involved are as follows.
// 1. Make an empty hashMap with key and value as array of 0's as prefixed values.
// 2. Add the values to hashMap. Here keys are contextual to kpi and will not be duplicate.
// 3. Convert Map to 2d Array and then sort.
func TransformRowsResultGroup(queryResults []QueryResult, timezoneString string, isTimezoneEnabled bool) [][]interface{} {
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

	// Step 3
	currentResultantRows := make([][]interface{}, 0, 0)
	for key, value := range resultAsMap {
		currentRow := make([]interface{}, 0)
		columns := strings.Split(key, ":;")
		for _, column := range columns[:len(columns)-1] {
			if strings.HasPrefix(column, "dat$") {
				unixValue, _ := strconv.ParseInt(strings.TrimPrefix(column, "dat$"), 10, 64)
				columnValue, _ := U.GetTimeFromUnixTimestampWithZone(unixValue, timezoneString, isTimezoneEnabled)
				currentRow = append(currentRow, columnValue)
			} else {
				currentRow = append(currentRow, column)
			}
		}
		currentRow = append(currentRow, value...)

		currentResultantRows = append(currentResultantRows, currentRow)
	}
	currentResultantRows = U.GetSorted2DArrays(currentResultantRows)
	return currentResultantRows
}
