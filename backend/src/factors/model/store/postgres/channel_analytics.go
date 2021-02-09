package postgres

import (
	"errors"
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"fmt"
	"net/http"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
)

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
	CAFilterCampaign             = "campaign"
	CAFilterAdGroup              = "ad_group"
	CAFilterAd                   = "ad"
	CAFilterKeyword              = "keyword"
	CAFilterQuery                = "query"
	CAFilterAdset                = "adset"
	CAChannelGoogleAds           = "google_ads"
	CAChannelFacebookAds         = "facebook_ads"
	CAAllChannelAds              = "all_ads"
	CAColumnValueAll             = "all"
	CAChannelGroupKey            = "group_key"
	innerJoinClause              = " INNER JOIN "
	channeAnalyticsLimit         = " LIMIT 2500 "
	source                       = "source"
)

var CAChannels = []string{
	CAChannelGoogleAds,
	CAChannelFacebookAds,
	CAAllChannelAds,
}

var channelMetricsToOperation = map[string]string{
	"impressions": "sum",
	"clicks":      "sum",
	"spend":       "sum",
	"conversion":  "sum",
}

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

// GetChannelConfig - @TODO Kark v1
func (pg *Postgres) GetChannelConfig(channel string, reqID string) (*model.ChannelConfigResult, int) {
	if !(isValidChannel(channel)) {
		return &model.ChannelConfigResult{}, http.StatusBadRequest
	}

	var result *model.ChannelConfigResult
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
func buildAllChannelConfig() *model.ChannelConfigResult {
	properties := buildProperties(allChannelsPropertyToRelated)
	objectsAndProperties := buildObjectsAndProperties(properties, objectsForAllChannels)

	return &model.ChannelConfigResult{
		SelectMetrics:        selectableMetricsForAllChannels,
		ObjectsAndProperties: objectsAndProperties,
	}
}

// @TODO Kark v1
func buildObjectsAndProperties(properties []model.ChannelProperty,
	filterObjectNames []string) []model.ChannelObjectAndProperties {

	var objectsAndProperties []model.ChannelObjectAndProperties
	for _, filterObjectName := range filterObjectNames {
		var objectAndProperties model.ChannelObjectAndProperties
		objectAndProperties.Name = filterObjectName
		objectAndProperties.Properties = properties
		objectsAndProperties = append(objectsAndProperties, objectAndProperties)
	}
	return objectsAndProperties
}

// @TODO Kark v1
func buildProperties(propertiesAndRelated map[string]PropertiesAndRelated) []model.ChannelProperty {
	var properties []model.ChannelProperty
	for propertyName, propertyRelated := range propertiesAndRelated {
		var property model.ChannelProperty
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
func (pg *Postgres) GetChannelFilterValuesV1(projectID uint64, channel, filterObject,
	filterProperty string, reqID string) (model.ChannelFilterValues, int) {

	var channelFilterValues model.ChannelFilterValues
	if !isValidChannel(channel) || !isValidFilterKey(filterObject) {
		return channelFilterValues, http.StatusBadRequest
	}

	var filterValues []interface{}
	var errCode int
	switch channel {
	case CAAllChannelAds:
		filterValues, errCode = pg.GetAllChannelFilterValues(projectID, filterObject, filterProperty, reqID)
	case CAChannelFacebookAds:
		filterValues, errCode = pg.GetFacebookFilterValues(projectID, filterObject, filterProperty, reqID)
	case CAChannelGoogleAds:
		filterValues, errCode = pg.GetAdwordsFilterValues(projectID, filterObject, filterProperty, reqID)
	}

	if errCode != http.StatusFound {
		return channelFilterValues, http.StatusInternalServerError
	}
	channelFilterValues.FilterValues = filterValues

	return channelFilterValues, http.StatusFound
}

// GetAllChannelFilterValues - @Kark TODO v1
func (pg *Postgres) GetAllChannelFilterValues(projectID uint64, filterObject, filterProperty string, reqID string) ([]interface{}, int) {
	logCtx := log.WithField("project_id", projectID).WithField("req_id", reqID)
	adwordsSQL, adwordsParams, adwordsErr := pg.GetAdwordsSQLQueryAndParametersForFilterValues(projectID, filterObject, filterProperty)
	facebookSQL, facebookParams, facebookErr := pg.GetFacebookSQLQueryAndParametersForFilterValues(projectID, filterObject, filterProperty)

	if adwordsErr != http.StatusFound {
		return []interface{}{}, adwordsErr
	}
	if facebookErr != http.StatusFound {
		return []interface{}{}, facebookErr
	}

	unionQuery := "SELECT filter_value from ( " + adwordsSQL + " UNION " + facebookSQL + " ) all_ads LIMIT 5000"
	unionParams := append(adwordsParams, facebookParams...)
	_, resultRows, _ := pg.ExecuteSQL(unionQuery, unionParams, logCtx)

	return Convert2DArrayTo1DArray(resultRows), http.StatusFound
}

// RunChannelGroupQuery - @TODO Kark v1
func (pg *Postgres) RunChannelGroupQuery(projectID uint64, queriesOriginal []model.ChannelQueryV1, reqID string) (model.ChannelResultGroupV1, int) {
	queries := make([]model.ChannelQueryV1, 0, 0)
	U.DeepCopy(&queriesOriginal, &queries)

	var resultGroup model.ChannelResultGroupV1
	resultGroup.Results = make([]model.ChannelQueryResultV1, len(queries))
	var waitGroup sync.WaitGroup
	count := 0
	waitGroup.Add(U.MinInt(len(queries), AllowedGoroutines))
	for index, query := range queries {
		count++
		go pg.runSingleChannelQuery(projectID, query, &resultGroup, index, &waitGroup, reqID)
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
func (pg *Postgres) runSingleChannelQuery(projectID uint64, query model.ChannelQueryV1,
	resultHolder *model.ChannelResultGroupV1, index int, waitGroup *sync.WaitGroup, reqID string) {

	environment := C.GetConfig().Env
	defer waitGroup.Done()
	defer U.NotifyOnPanicWithError(environment, "app_server")
	result, _ := pg.ExecuteChannelQueryV1(projectID, &query, reqID)
	(*resultHolder).Results[index] = *result
}

// ExecuteChannelQueryV1 - @Kark TODO v1
// TODO error handling.
func (pg *Postgres) ExecuteChannelQueryV1(projectID uint64, query *model.ChannelQueryV1,
	reqID string) (*model.ChannelQueryResultV1, int) {

	logCtx := log.WithField("req_id", reqID)
	queryResult := &model.ChannelQueryResultV1{}
	var columns []string
	var resultMetrics [][]interface{}
	status := http.StatusOK
	var err error
	if !(isValidChannel(query.Channel)) {
		return queryResult, http.StatusBadRequest
	}
	switch query.Channel {
	case CAAllChannelAds:
		columns, resultMetrics, err = pg.executeAllChannelsQueryV1(projectID, query, reqID)
	case CAChannelFacebookAds:
		columns, resultMetrics, err = pg.ExecuteFacebookChannelQueryV1(projectID, query, reqID)
	case CAChannelGoogleAds:
		columns, resultMetrics, err = pg.ExecuteAdwordsChannelQueryV1(projectID, query, reqID)
	}
	if err != nil {
		logCtx.Warn(query)
		logCtx.WithError(err).Error("Failed in channel analytics with following error: ")
		status = http.StatusBadRequest
	}
	resultMetrics = U.ConvertInternalToExternal(resultMetrics)
	queryResult.Headers = columns
	queryResult.Rows = resultMetrics

	return queryResult, status
}

// This function relies on all the columns in all tables to be in same order.
// Case 1: When there is no breakdown, there is just metrics being recalculated.
// Case 2: When there is breakdown by date, there is regrouping by date.
// Case 3: When there is breakdown by source and group.property, there is no requirement of regrouping in all channel.
func (pg *Postgres) executeAllChannelsQueryV1(projectID uint64, query *model.ChannelQueryV1,
	reqID string) ([]string, [][]interface{}, error) {

	logCtx := log.WithField("project_id", projectID).WithField("req_id", reqID)
	fetchSource := true
	var unionQuery string
	var unionParams []interface{}
	var selectMetrics, columns []string
	adwordsSQL, adwordsParams, adwordsSelectKeys, adwordsMetrics, adwordsErr := pg.GetSQLQueryAndParametersForAdwordsQueryV1(projectID, query, reqID, fetchSource)
	facebookSQL, facebookParams, _, _, facebookErr := pg.GetSQLQueryAndParametersForFacebookQueryV1(projectID, query, reqID, fetchSource)

	if adwordsErr != nil {
		return make([]string, 0, 0), [][]interface{}{}, adwordsErr
	}
	if facebookErr != nil {
		return make([]string, 0, 0), [][]interface{}{}, facebookErr
	}

	adwordsSQL = fmt.Sprintf("( %s )", adwordsSQL[:len(adwordsSQL)-2])
	facebookSQL = fmt.Sprintf("( %s )", facebookSQL[:len(facebookSQL)-2])
	if (query.GroupBy == nil || len(query.GroupBy) == 0) && (query.GroupByTimestamp == nil || len(query.GroupByTimestamp.(string)) == 0) {
		for _, metric := range adwordsMetrics {
			value := fmt.Sprintf("%s(%s) as %s", channelMetricsToOperation[metric], metric, metric)
			selectMetrics = append(selectMetrics, value)
		}
		unionQuery = fmt.Sprintf("SELECT %s FROM ( %s UNION %s ) all_ads ORDER BY %s %s", joinWithComma(selectMetrics...),
			adwordsSQL, facebookSQL, getOrderByClause(adwordsMetrics), channeAnalyticsLimit)
		unionParams = append(adwordsParams, facebookParams...)
	} else if (query.GroupBy == nil || len(query.GroupBy) == 0) && (!(query.GroupByTimestamp == nil || len(query.GroupByTimestamp.(string)) == 0)) {
		selectMetrics = append(selectMetrics, model.AliasDateTime)
		for _, metric := range adwordsMetrics {
			value := fmt.Sprintf("%s(%s) as %s", channelMetricsToOperation[metric], metric, metric)
			selectMetrics = append(selectMetrics, value)
		}
		unionQuery = fmt.Sprintf("SELECT %s FROM ( %s UNION %s ) all_ads GROUP BY %s ORDER BY %s %s", joinWithComma(selectMetrics...), adwordsSQL, facebookSQL,
			model.AliasDateTime, getOrderByClause(adwordsMetrics), channeAnalyticsLimit)
		unionParams = append(adwordsParams, facebookParams...)
	} else {
		unionQuery = fmt.Sprintf("SELECT * FROM ( %s UNION %s ) all_ads ORDER BY %s %s;", adwordsSQL, facebookSQL, getOrderByClause(adwordsMetrics), channeAnalyticsLimit)
		unionParams = append(adwordsParams, facebookParams...)
	}
	columns = append(adwordsSelectKeys, adwordsMetrics...)
	_, resultMetrics, err := pg.ExecuteSQL(unionQuery, unionParams, logCtx)
	return columns, resultMetrics, err
}

// GetChannelFilterValues - @Kark TODO v0
func (pg *Postgres) GetChannelFilterValues(projectID uint64, channel, filter string) ([]string, int) {
	if !isValidChannel(channel) || !isValidFilterKey(filter) {
		return []string{}, http.StatusBadRequest
	}

	// supports only adwords now.
	docType, err := GetAdwordsDocumentTypeForFilterKey(filter)
	if err != nil {
		return []string{}, http.StatusInternalServerError
	}

	filterValues, errCode := pg.GetAdwordsFilterValuesByType(projectID, docType)
	if errCode != http.StatusFound {
		return []string{}, http.StatusInternalServerError
	}

	return filterValues, http.StatusFound
}

// ExecuteChannelQuery - @Kark TODO v0
func (pg *Postgres) ExecuteChannelQuery(projectID uint64,
	queryOriginal *model.ChannelQuery) (*model.ChannelQueryResult, int) {

	var query *model.ChannelQuery
	U.DeepCopy(queryOriginal, &query)

	if !isValidChannel(query.Channel) || !isValidFilterKey(query.FilterKey) ||
		query.From == 0 || query.To == 0 {
		return nil, http.StatusBadRequest
	}

	if query.Channel == "google_ads" {
		result, errCode := pg.ExecuteAdwordsChannelQuery(projectID, query)
		if errCode != http.StatusOK {
			log.WithField("project_id", projectID).Error("Failed to execute adwords channel query.")
			return nil, http.StatusInternalServerError
		}
		return result, http.StatusOK
	}
	if query.Channel == "facebook_ads" {
		result, errCode := pg.ExecuteFacebookChannelQuery(projectID, query)
		if errCode != http.StatusOK {
			log.WithField("project_id", projectID).Error("Failed to execute facebook channel query.")
			return nil, http.StatusInternalServerError
		}
		return result, http.StatusOK
	}
	return nil, http.StatusBadRequest
}

// Common Methods for facebook and adwords starts here.

// Convert2DArrayTo1DArray ...
// @Kark TODO v1
func Convert2DArrayTo1DArray(inputArray [][]interface{}) []interface{} {
	result := make([]interface{}, 0, 0)
	for _, row := range inputArray {
		result = append(result, row...)
	}
	return result
}

// @Kark TODO v1
func hasAllIDsOnlyInGroupBy(query *model.ChannelQueryV1) bool {
	for _, groupBy := range query.GroupBy {
		if !(strings.Contains(groupBy.Property, "id") || strings.Contains(groupBy.Property, "ID")) {
			return false
		}
	}
	return true
}

// @Kark TODO v1
func appendSelectTimestampIfRequiredForChannels(stmnt string, groupByTimestamp string, timezone string) string {
	if groupByTimestamp == "" {
		return stmnt
	}

	return joinWithComma(stmnt, fmt.Sprintf("%s as %s",
		getSelectTimestampByTypeForChannels(groupByTimestamp, timezone), model.AliasDateTime))
}

// TO change.
// @Kark TODO v1
func getSelectTimestampByTypeForChannels(timestampType, timezone string) string {

	var selectTz string
	var selectStr string

	if timezone == "" {
		selectTz = model.DefaultTimezone
	} else {
		selectTz = timezone
	}
	if timestampType == model.GroupByTimestampHour {
		selectStr = fmt.Sprintf("date_trunc('hour', to_timestamp(timestamp::text, 'YYYYMMDD') AT TIME ZONE '%s')", selectTz)
	} else if timestampType == model.GroupByTimestampWeek {
		selectStr = fmt.Sprintf("date_trunc('week', to_timestamp(timestamp::text, 'YYYYMMDD') AT TIME ZONE '%s')", selectTz)
	} else if timestampType == model.GroupByTimestampMonth {
		selectStr = fmt.Sprintf("date_trunc('month', to_timestamp(timestamp::text, 'YYYYMMDD') AT TIME ZONE '%s')", selectTz)
	} else {
		// defaults to GroupByTimestampDate.
		selectStr = fmt.Sprintf("date_trunc('day', to_timestamp(timestamp::text, 'YYYYMMDD') AT TIME ZONE '%s')", selectTz)
	}

	return selectStr
}

// @Kark TODO v1
func getOrderByClause(selectMetrics []string) string {
	selectMetricsWithDesc := make([]string, 0, 0)
	for _, selectMetric := range selectMetrics {
		selectMetricsWithDesc = append(selectMetricsWithDesc, selectMetric+" DESC")
	}
	return joinWithComma(selectMetricsWithDesc...)
}

// ExecuteSQL - @Kark TODO v1
func (pg *Postgres) ExecuteSQL(sqlStatement string, params []interface{}, logCtx *log.Entry) ([]string, [][]interface{}, error) {
	db := C.GetServices().Db
	rows, err := db.Raw(sqlStatement, params...).Rows()
	if err != nil {
		logCtx.WithError(err).Error("SQL Query failed")
		return nil, nil, err
	}

	defer rows.Close()
	columns, resultRows, err := U.DBReadRows(rows)

	if err != nil {
		return nil, nil, err
	}
	if len(resultRows) == 0 {
		log.Warn("Aggregate query returned zero rows.")
		return nil, make([][]interface{}, 0, 0), errors.New("no rows returned")
	}
	return columns, resultRows, nil
}
