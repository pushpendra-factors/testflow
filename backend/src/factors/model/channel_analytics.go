package model

import (
	U "factors/util"
	"net/http"
	"sync"

	log "github.com/sirupsen/logrus"
)

// ChannelFilterValues - @TODO Kark v1
type ChannelFilterValues struct {
	FilterValues []interface{} `json:"filter_values"`
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
}

// ChannelQueryV1 - @TODO Kark v1
type ChannelQueryV1 struct {
	Channel          string      `json:"channel"`
	SelectMetrics    []string    `json:"select_metrics"`
	Filters          []FilterV1  `json:"filters"`
	GroupBy          []GroupBy   `json:"group_by"`
	GroupByTimestamp interface{} `json:"gbt"`
	Timezone         string      `json:"time_zone"`
	From             int64       `json:"fr"`
	To               int64       `json:"to"`
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

// GroupBy - @TODO Kark v1
type GroupBy struct {
	Object   string `json:"name"`
	Property string `json:"property"`
}

// FilterV1 - @TODO Kark v1
type FilterV1 struct {
	Object    string `json:"name"`
	Property  string `json:"property"`
	Condition string `json:"condition"`
	Value     string `json:"value"`
	LogicalOp string `json:"logical_operator"`
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

// ChannelQueryResult - @TODO Kark v0
type ChannelQueryResult struct {
	Metrics          *map[string]interface{} `json:"metrics"`
	MetricsBreakdown *ChannelBreakdownResult `json:"metrics_breakdown"`
	Meta             *ChannelQueryResultMeta `json:"meta"`
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

func (q *ChannelQueryUnit) SetQueryDateRange(from, to int64) {
	q.Query.From, q.Query.To = from, to
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

func (q *ChannelGroupQueryV1) SetQueryDateRange(from, to int64) {
	for i := 0; i < len(q.Queries); i++ {
		q.Queries[i].From, q.Queries[i].To = from, to
	}
}

func (query *ChannelQueryV1) GetGroupByTimestamp() string {
	if query.GroupByTimestamp == nil {
		return ""
	}
	return query.GroupByTimestamp.(string)
}

const CAChannelGoogleAds = "google_ads"
const CAChannelFacebookAds = "facebook_ads"
const CAAllChannelAds = "all_ads"
const CAChannelGroupKey = "group_key"

var CAChannels = []string{
	CAChannelGoogleAds,
	CAChannelFacebookAds,
	CAAllChannelAds,
}

const CAColumnValueAll = "all"

const (
	CAColumnImpressions          = "impressions"
	CAColumnClicks               = "clicks"
	CAColumnTotalCost            = "total_cost"
	CAColumnConversions          = "conversions"
	CAColumnAllConversions       = "all_conversions"
	CAColumnCostPerClick         = "cost_per_click"
	CAColumnConversionRate       = "conversion_rate"
	CAColumnCostPerConversion    = "cost_per_conversion"
	CAColumnFrequency            = "frequency"
	CAColumnReach                = "reach"
	CAColumnInlinePostEngagement = "inline_post_engagement"
	CAColumnUniqueClicks         = "unique_clicks"
	CAColumnName                 = "name"
	CAColumnPlatform             = "platform"
)

const (
	CAFilterCampaign = "campaign"
	CAFilterAdGroup  = "ad_group"
	CAFilterAd       = "ad"
	CAFilterKeyword  = "keyword"
	CAFilterQuery    = "query"
	CAFilterAdset    = "adset"
)

// CAFilters ...
var CAFilters = []string{
	CAFilterCampaign,
	CAFilterAdGroup,
	CAFilterAd,
	CAFilterKeyword,
	CAFilterQuery,
	CAFilterAdset,
}

// TODO: Move and fetch it from respective channels - allChannels, adwords etc.. because this is error prone.
var selectableMetricsForAllChannels = []string{"impressions", "clicks", "spend"}
var objectsForAllChannels = []string{CAFilterCampaign, CAFilterAdGroup}

// PropertiesAndRelated - TODO Kark v1
type PropertiesAndRelated struct {
	typeOfProperty string // can be categorical or numerical
}

var allChannelsPropertyToRelated = map[string]PropertiesAndRelated{
	"name": PropertiesAndRelated{
		typeOfProperty: U.PropertyTypeCategorical,
	},
	"id": PropertiesAndRelated{
		typeOfProperty: U.PropertyTypeCategorical,
	},
}

// ChannelConfigResult - TODO Kark v1
type ChannelConfigResult struct {
	SelectMetrics        []string              `json:"select_metrics"`
	ObjectsAndProperties []ObjectAndProperties `json:"object_and_properties"`
}

// ObjectAndProperties - TODO Kark v1
type ObjectAndProperties struct {
	Name       string     `json:"name"`
	Properties []Property `json:"properties"`
}

// Property - TODO Kark v1
type Property struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// GetChannelConfig - @TODO Kark v1
func GetChannelConfig(channel string, reqID string) (*ChannelConfigResult, int) {
	if !(isValidChannel(channel)) {
		return &ChannelConfigResult{}, http.StatusBadRequest
	}

	var result *ChannelConfigResult
	switch channel {
	case CAAllChannelAds:
		result = buildAllChannelConfig()
	case CAChannelFacebookAds:
		result = buildFbChannelConfig()
	case CAChannelGoogleAds:
		result = buildAdwordsChannelConfig()
	}
	return result, http.StatusOK
}

// @TODO Kark v0, v1
func isValidFilterKey(filter string) bool {
	for _, f := range CAFilters {
		if filter == f {
			return true
		}
	}

	return false
}

// @TODO Kark v1
func isValidChannel(channel string) bool {
	for _, c := range CAChannels {
		if channel == c {
			return true
		}
	}

	return false
}

// @TODO Kark v1
func buildAllChannelConfig() *ChannelConfigResult {
	properties := buildProperties(allChannelsPropertyToRelated)
	objectsAndProperties := buildObjectsAndProperties(properties, objectsForAllChannels)

	return &ChannelConfigResult{
		SelectMetrics:        selectableMetricsForAllChannels,
		ObjectsAndProperties: objectsAndProperties,
	}
}

// @TODO Kark v1
func buildObjectsAndProperties(properties []Property, filterObjectNames []string) []ObjectAndProperties {
	var objectsAndProperties []ObjectAndProperties
	for _, filterObjectName := range filterObjectNames {
		var objectAndProperties ObjectAndProperties
		objectAndProperties.Name = filterObjectName
		objectAndProperties.Properties = properties
		objectsAndProperties = append(objectsAndProperties, objectAndProperties)
	}
	return objectsAndProperties
}

// @TODO Kark v1
func buildProperties(propertiesAndRelated map[string]PropertiesAndRelated) []Property {
	var properties []Property
	for propertyName, propertyRelated := range propertiesAndRelated {
		var property Property
		property.Name = propertyName
		property.Type = propertyRelated.typeOfProperty
		properties = append(properties, property)
	}
	return properties
}

// GetChannelFilterValuesV1 - TODO: Define the role of classes and encapsulation correctly.
// Should request params to correct types be converted here - QueryAggregator responsibility?
// Adwords - Keywords will fail currently.
// @TODO Kark v1
func GetChannelFilterValuesV1(projectID uint64, channel, filterObject, filterProperty string, reqID string) (ChannelFilterValues, int) {
	var channelFilterValues ChannelFilterValues
	if !isValidChannel(channel) || !isValidFilterKey(filterObject) {
		return channelFilterValues, http.StatusBadRequest
	}

	var filterValues []interface{}
	var errCode int
	switch channel {
	case CAAllChannelAds:
		filterValues, errCode = GetAllChannelFilterValues(projectID, filterObject, filterProperty, reqID)
	case CAChannelFacebookAds:
		filterValues, errCode = GetFacebookFilterValues(projectID, filterObject, filterProperty, reqID)
	case CAChannelGoogleAds:
		filterValues, errCode = GetAdwordsFilterValues(projectID, filterObject, filterProperty, reqID)
	}

	if errCode != http.StatusFound {
		return channelFilterValues, http.StatusInternalServerError
	}
	channelFilterValues.FilterValues = filterValues

	return channelFilterValues, http.StatusFound
}

// GetAllChannelFilterValues - @Kark TODO v1
func GetAllChannelFilterValues(projectID uint64, filterObject, filterProperty string, reqID string) ([]interface{}, int) {
	logCtx := log.WithField("project_id", projectID).WithField("req_id", reqID)
	adwordsSQL, adwordsParams, adwordsErr := GetAdwordsSQLQueryAndParametersForFilterValues(projectID, filterObject, filterProperty)
	facebookSQL, facebookParams, facebookErr := GetFacebookSQLQueryAndParametersForFilterValues(projectID, filterObject, filterProperty)

	if adwordsErr != http.StatusFound {
		return []interface{}{}, adwordsErr
	}
	if facebookErr != http.StatusFound {
		return []interface{}{}, facebookErr
	}

	unionQuery := "SELECT filter_value from ( " + adwordsSQL + " UNION " + facebookSQL + " ) all_ads LIMIT 5000"
	unionParams := append(adwordsParams, facebookParams...)
	_, resultRows, _ := ExecuteSQL(unionQuery, unionParams, logCtx)

	return Convert2DArrayTo1DArray(resultRows), http.StatusFound
}

// RunChannelGroupQuery - @TODO Kark v1
func RunChannelGroupQuery(projectID uint64, queries []ChannelQueryV1, reqID string) (ChannelResultGroupV1, int) {

	var resultGroup ChannelResultGroupV1
	resultGroup.Results = make([]ChannelQueryResultV1, len(queries))
	var waitGroup sync.WaitGroup
	count := 0
	waitGroup.Add(U.MinInt(len(queries), AllowedGoroutines))
	for index, query := range queries {
		count++
		go runSingleChannelQuery(projectID, query, &resultGroup, index, &waitGroup, reqID)
		if count%AllowedGoroutines == 0 {
			waitGroup.Wait()
			waitGroup.Add(U.MinInt(len(queries)-count, AllowedGoroutines))
		}
	}
	waitGroup.Wait()
	return resultGroup, http.StatusOK
}

// @Kark TODO v1
// TODO Handling errorcase.
func runSingleChannelQuery(projectID uint64, query ChannelQueryV1, resultHolder *ChannelResultGroupV1, index int, waitGroup *sync.WaitGroup, reqID string) {
	logCtx := log.WithField("xreq_id", reqID)
	logCtx.Info(query)
	defer waitGroup.Done()
	result, _ := ExecuteChannelQueryV1(projectID, &query, reqID)
	(*resultHolder).Results[index] = *result
}

// ExecuteChannelQueryV1 - @Kark TODO v1
// TODO error handling.
func ExecuteChannelQueryV1(projectID uint64, query *ChannelQueryV1, reqID string) (*ChannelQueryResultV1, int) {
	queryResult := &ChannelQueryResultV1{}
	status := http.StatusOK
	if !(isValidChannel(query.Channel)) {
		return queryResult, http.StatusBadRequest
	}

	switch query.Channel {
	// case CAAllChannelAds:
	// 	result = ExecuteAllChannelsQueryV1()
	// case CAChannelFacebookAds:
	// 	result = ExecuteFacebookChannelQueryV1()
	case CAChannelGoogleAds:
		columns := buildColumns(query)
		_, resultMetrics, err := ExecuteAdwordsChannelQueryV1(projectID, query, reqID)
		queryResult.Headers = columns
		queryResult.Rows = resultMetrics
		if err != nil {
			status = http.StatusBadRequest
		}
	}
	return queryResult, status
}

// GetChannelFilterValues - @Kark TODO v0
func GetChannelFilterValues(projectID uint64, channel, filter string) ([]string, int) {
	if !isValidChannel(channel) || !isValidFilterKey(filter) {
		return []string{}, http.StatusBadRequest
	}

	// supports only adwords now.
	docType, err := GetAdwordsDocumentTypeForFilterKey(filter)
	if err != nil {
		return []string{}, http.StatusInternalServerError
	}

	filterValues, errCode := GetAdwordsFilterValuesByType(projectID, docType)
	if errCode != http.StatusFound {
		return []string{}, http.StatusInternalServerError
	}

	return filterValues, http.StatusFound
}

// ExecuteChannelQuery - @Kark TODO v0
func ExecuteChannelQuery(projectID uint64, query *ChannelQuery) (*ChannelQueryResult, int) {
	if !isValidChannel(query.Channel) || !isValidFilterKey(query.FilterKey) ||
		query.From == 0 || query.To == 0 {
		return nil, http.StatusBadRequest
	}

	if query.Channel == "google_ads" {
		result, errCode := ExecuteAdwordsChannelQuery(projectID, query)
		if errCode != http.StatusOK {
			log.WithField("project_id", projectID).Error("Failed to execute adwords channel query.")
			return nil, http.StatusInternalServerError
		}
		return result, http.StatusOK
	}
	if query.Channel == "facebook_ads" {
		result, errCode := ExecuteFacebookChannelQuery(projectID, query)
		if errCode != http.StatusOK {
			log.WithField("project_id", projectID).Error("Failed to execute facebook channel query.")
			return nil, http.StatusInternalServerError
		}
		return result, http.StatusOK
	}
	return nil, http.StatusBadRequest
}

// Convert2DArrayTo1DArray ...
func Convert2DArrayTo1DArray(inputArray [][]interface{}) []interface{} {
	result := make([]interface{}, 0, 0)
	for _, row := range inputArray {
		result = append(result, row...)
	}
	return result
}

func buildColumns(query *ChannelQueryV1) []string {
	result := make([]string, 0, 0)
	for _, groupBy := range query.GroupBy {
		result = append(result, groupBy.Object+"_"+groupBy.Property)
	}

	groupByTimeStamp := query.GetGroupByTimestamp()
	if len(groupByTimeStamp) != 0 {
		result = append(result, "datetime")
	}
	for _, selectMetrics := range query.SelectMetrics {
		result = append(result, selectMetrics)
	}
	return result
}
