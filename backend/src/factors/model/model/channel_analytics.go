package model

import (
	cacheRedis "factors/cache/redis"
	U "factors/util"
	"net/http"
	"strings"
	"time"
)

const (
	CampaignPrefix = "campaign_"
	AdgroupPrefix  = "ad_group_"
	KeywordPrefix  = "keyword_"
	Channel        = "channel"
	GoogleAds      = "Google Ads"
	FacebookAds    = "Facebook Ads"
	LinkedinAds    = "LinkedIn Ads"
	OldGoogleAds   = "google_ads"
	OldFacebookAds = "facebook_ads"
	OldLinkedinAds = "linkedin_ads"
)

type ChannelConfigResult struct {
	SelectMetrics        []string                     `json:"select_metrics"`
	ObjectsAndProperties []ChannelObjectAndProperties `json:"object_and_properties"`
}

type ChannelObjectAndProperties struct {
	Name       string            `json:"name"`
	Properties []ChannelProperty `json:"properties"`
}

type ChannelProperty struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// PropertiesAndRelated - TODO Kark v1
type PropertiesAndRelated struct {
	TypeOfProperty string // can be categorical or numerical
}

// ChannelQuery - @TODO Kark v0
type ChannelQuery struct {
	Channel     string `json:"channel"`
	FilterKey   string `json:"filter_key"`
	FilterValue string `json:"filter_value"`
	From        int64  `json:"from"` // unix timestamp
	To          int64  `json:"to"`   // unix timestamp
	Status      string `json:"status"`
	MatchType   string `json:"match_type"` // optional
	Breakdown   string `json:"breakdown"`
	Timezone    string `json:"time_zone"`
}

// ChannelQueryV1 - @TODO Kark v1
type ChannelQueryV1 struct {
	Channel          string            `json:"channel"`
	SelectMetrics    []string          `json:"select_metrics"`
	Filters          []ChannelFilterV1 `json:"filters"`
	GroupBy          []ChannelGroupBy  `json:"group_by"`
	GroupByTimestamp interface{}       `json:"gbt"`
	Timezone         string            `json:"time_zone"`
	From             int64             `json:"fr"`
	To               int64             `json:"to"`
}

// ChannelGroupBy - @TODO Kark v1
type ChannelGroupBy struct {
	Object   string `json:"name"`
	Property string `json:"property"`
}

// ChannelFilterV1 - @TODO Kark v1
type ChannelFilterV1 struct {
	Object    string `json:"name"`
	Property  string `json:"property"`
	Condition string `json:"condition"`
	Value     string `json:"value"`
	LogicalOp string `json:"logical_operator"`
}

// ChannelQueryResult - @TODO Kark v0
type ChannelQueryResult struct {
	Metrics          *map[string]interface{} `json:"metrics"`
	MetricsBreakdown *ChannelBreakdownResult `json:"metrics_breakdown"`
	Meta             *ChannelQueryResultMeta `json:"meta"`
	Query            interface{}             `json:"query"`
}

// ChannelBreakdownResult - @TODO Kark v0
type ChannelBreakdownResult struct {
	Headers []string        `json:"headers"`
	Rows    [][]interface{} `json:"rows"`
}

// ChannelQueryResultMeta - @TODO Kark v0
type ChannelQueryResultMeta struct {
	Currency string `json:"currency"`
}

// ChannelFilterValues - @TODO Kark v1
type ChannelFilterValues struct {
	FilterValues []interface{} `json:"filter_values"`
}

// ChannelResultGroupV1 - @TODO Kark v1
type ChannelResultGroupV1 struct {
	Results []ChannelQueryResultV1 `json:"result_group"`
}

// ChannelQueryResultV1 - @TODO Kark v1
type ChannelQueryResultV1 struct {
	Headers []string        `json:"headers"`
	Rows    [][]interface{} `json:"rows"`
}

// ChannelQueryUnit - @TODO Kark v0
type ChannelQueryUnit struct {
	// Json tag should match with Query's class,
	// query dispatched based on this.
	Class string                  `json:"cl"`
	Query *ChannelQuery           `json:"query"`
	Meta  *map[string]interface{} `json:"meta"`
}

func (q *ChannelQueryUnit) GetClass() string {
	return q.Class
}

func (q *ChannelQueryUnit) GetQueryDateRange() (from, to int64) {
	return q.Query.From, q.Query.To
}

func (q *ChannelQueryUnit) SetTimeZone(timezoneString U.TimeZoneString) {
	q.Query.Timezone = string(timezoneString)
}

func (q *ChannelQueryUnit) GetTimeZone() U.TimeZoneString {
	return U.TimeZoneString(q.Query.Timezone)
}

func (q *ChannelQueryUnit) SetQueryDateRange(from, to int64) {
	q.Query.From, q.Query.To = from, to
}

func (q *ChannelQueryUnit) GetQueryCacheHashString() (string, error) {
	queryMap, err := U.EncodeStructTypeToMap(q)
	if err != nil {
		return "", err
	}
	delete(queryMap, "meta")
	delete(queryMap["query"].(map[string]interface{}), "from")
	delete(queryMap["query"].(map[string]interface{}), "to")

	queryHash, err := U.GenerateHashStringForStruct(queryMap)
	if err != nil {
		return "", err
	}
	return queryHash, nil
}

func (q *ChannelQueryUnit) GetQueryCacheRedisKey(projectID uint64) (*cacheRedis.Key, error) {
	hashString, err := q.GetQueryCacheHashString()
	if err != nil {
		return nil, err
	}
	suffix := getQueryCacheRedisKeySuffix(hashString, q.Query.From, q.Query.To, U.TimeZoneString(q.Query.Timezone))
	return cacheRedis.NewKey(projectID, QueryCacheRedisKeyPrefix, suffix)
}

func (q *ChannelQueryUnit) GetQueryCacheExpiry() float64 {
	return getQueryCacheResultExpiry(q.Query.From, q.Query.To, q.Query.Timezone)
}

func (q *ChannelQueryUnit) TransformDateTypeFilters() error {
	return nil
}

func (q *ChannelQueryUnit) ConvertAllDatesFromTimezone1ToTimezone2(currentTimezone, nextTimezone string) error {
	return nil
}

// ChannelGroupQueryV1 - @TODO Kark v1
type ChannelGroupQueryV1 struct {
	Class   string           `json:"cl"`
	Queries []ChannelQueryV1 `json:"query_group"`
}

func (q *ChannelGroupQueryV1) GetClass() string {
	if len(q.Queries) > 0 {
		// all queries in query group are expected to belong to same class
		return q.Class
	}
	return ""
}

func (q *ChannelGroupQueryV1) GetQueryDateRange() (from, to int64) {
	if len(q.Queries) > 0 {
		// all queries in query group are expected to run for same time range
		return q.Queries[0].From, q.Queries[0].To
	}
	return 0, 0
}

func (q *ChannelGroupQueryV1) SetTimeZone(timezoneString U.TimeZoneString) {
	for i := 0; i < len(q.Queries); i++ {
		q.Queries[i].Timezone = string(timezoneString)
	}
}

func (q *ChannelGroupQueryV1) GetTimeZone() U.TimeZoneString {
	return U.TimeZoneString(q.Queries[0].Timezone)
}

func (q *ChannelGroupQueryV1) SetQueryDateRange(from, to int64) {
	for i := 0; i < len(q.Queries); i++ {
		q.Queries[i].From, q.Queries[i].To = from, to
	}
}

func (q *ChannelGroupQueryV1) GetQueryCacheHashString() (string, error) {
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

func (q *ChannelGroupQueryV1) GetQueryCacheRedisKey(projectID uint64) (*cacheRedis.Key, error) {
	hashString, err := q.GetQueryCacheHashString()
	if err != nil {
		return nil, err
	}
	suffix := getQueryCacheRedisKeySuffix(hashString, q.Queries[0].From, q.Queries[0].To, U.TimeZoneString(q.Queries[0].Timezone))
	return cacheRedis.NewKey(projectID, QueryCacheRedisKeyPrefix, suffix)
}

func (q *ChannelGroupQueryV1) GetQueryCacheExpiry() float64 {
	return getQueryCacheResultExpiry(q.Queries[0].From, q.Queries[0].To, q.Queries[0].Timezone)
}

func (q *ChannelGroupQueryV1) TransformDateTypeFilters() error {
	return nil
}

func (q *ChannelGroupQueryV1) ConvertAllDatesFromTimezone1ToTimezone2(currentTimezone, nextTimezone string) error {
	return nil
}

func (query *ChannelQueryV1) GetGroupByTimestamp() string {
	if query.GroupByTimestamp == nil {
		return ""
	}
	return query.GroupByTimestamp.(string)
}

// buildErrorResult takes the failure msg and wraps it into a model.QueryResult object
func BuildErrorResultForChannelsV1(errMsg string) *ChannelQueryResultV1 {
	errMsg = "Query failed:" + " - " + errMsg
	headers := []string{AliasError}
	rows := make([][]interface{}, 0, 0)
	row := make([]interface{}, 0, 0)
	row = append(row, errMsg)
	rows = append(rows, row)
	errorResult := &ChannelQueryResultV1{Headers: headers, Rows: rows}
	return errorResult
}

// sample format :
// {campaign_property: property_value, ad_group_property: property_value}
func GetGroupByCombinationsForChannelAnalytics(columns []string, resultMetrics [][]interface{}) []map[string]interface{} {
	groupByCombinations := make([]map[string]interface{}, 0)
	for _, resultRow := range resultMetrics {
		groupByCombination := make(map[string]interface{})
		for index, column := range columns {
			if strings.HasPrefix(column, CampaignPrefix) || strings.HasPrefix(column, AdgroupPrefix) || strings.HasPrefix(column, KeywordPrefix) {
				groupByCombination[column] = resultRow[index]
			}
		}
		if len(groupByCombination) != 0 {
			groupByCombinations = append(groupByCombinations, groupByCombination)
		}
	}
	return groupByCombinations
}

func TransformDateTypeValueForChannels(headers []string, rows [][]interface{}, groupByTimestampPresent bool, hasAnyGroupBy bool, timezoneString string) [][]interface{} {
	indexForDateTime := -1
	if headers[0] == AliasError {
		return rows
	}
	if !groupByTimestampPresent {
		return rows
	}
	for index, header := range headers {
		if header == "datetime" {
			indexForDateTime = index
			break
		}
	}

	for _, row := range rows {
		currentValueInTimeFormat, _ := row[indexForDateTime].(time.Time)
		row[indexForDateTime] = U.GetTimestampAsStrWithTimezone(currentValueInTimeFormat, timezoneString)
	}

	if hasAnyGroupBy && groupByTimestampPresent {
		for index, row := range rows {
			size := len(row)
			resultantRow := make([]interface{}, 0)
			resultantRow = append(resultantRow, row[size-2])
			resultantRow = append(resultantRow, row[:size-2]...)
			resultantRow = append(resultantRow, row[size-1])
			rows[index] = resultantRow
		}
	}

	return rows
}

func GetHeadersFromQuery(query ChannelQueryV1) []string {
	headers := make([]string, 0, 0)
	for _, currentGroupBy := range query.GroupBy {
		headers = append(headers, currentGroupBy.Object+"_"+currentGroupBy.Property)
	}

	if query.GroupByTimestamp == "" {
		headers = append(headers, "timestamp")
	}
	for _, metric := range query.SelectMetrics {
		headers = append(headers, metric)
	}
	return headers
}

func GetDecoupledFiltersForChannelBreakdownFilters(filters []ChannelFilterV1) ([]ChannelFilterV1, []ChannelFilterV1) {
	channelBreakdownFilters := make([]ChannelFilterV1, 0)
	genericFilters := make([]ChannelFilterV1, 0)
	for _, filter := range filters {
		if filter.Object == Channel {
			channelBreakdownFilters = append(channelBreakdownFilters, filter)
		} else {
			genericFilters = append(genericFilters, filter)
		}
	}
	return genericFilters, channelBreakdownFilters
}

func evaluateFilter(channelName string, filter ChannelFilterV1) bool {
	isChannelRequired := false
	if filter.Condition == EqualsOpStr || filter.Condition == ContainsOpStr {
		isChannelRequired = strings.Contains(strings.ToLower(channelName), strings.ToLower(filter.Value))
	} else if filter.Condition == NotEqualOpStr || filter.Condition == NotContainsOpStr {
		isChannelRequired = !(strings.Contains(strings.ToLower(channelName), strings.ToLower(filter.Value)))
	} else {
		return false
	}
	return isChannelRequired
}
func checkIfChannelReq(channelName string, filters []ChannelFilterV1) bool {
	isChannelReq := false
	for i, filter := range filters {
		if i == 0 {
			isChannelReq = evaluateFilter(channelName, filter)
		} else {
			if filter.LogicalOp == LOGICAL_OP_AND {
				isChannelReq = isChannelReq && evaluateFilter(channelName, filter)
				if !isChannelReq {
					return isChannelReq
				}
			} else {
				isChannelReq = isChannelReq || evaluateFilter(channelName, filter)
			}
		}
	}
	return isChannelReq
}

func GetRequiredChannels(filters []ChannelFilterV1) (bool, bool, bool, bool, int) {
	isAdwordsReq, isFacebookReq, isLinkedinReq, isBingAdsReq := false, false, false, false
	if len(filters) == 0 {
		return true, true, true, true, http.StatusOK
	}
	isAdwordsReq = checkIfChannelReq(GoogleAds, filters) || checkIfChannelReq(OldGoogleAds, filters)
	isFacebookReq = checkIfChannelReq(FacebookAds, filters) || checkIfChannelReq(OldFacebookAds, filters)
	isLinkedinReq = checkIfChannelReq(LinkedinAds, filters) || checkIfChannelReq(OldLinkedinAds, filters)
	isBingAdsReq = checkIfChannelReq(ChannelBingAds, filters)
	return isAdwordsReq, isFacebookReq, isLinkedinReq, isBingAdsReq, http.StatusOK
}
