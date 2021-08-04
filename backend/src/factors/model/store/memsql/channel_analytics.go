package memsql

import (
	C "factors/config"
	Const "factors/constants"
	"factors/model/model"
	U "factors/util"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	CAColumnImpressions                    = "impressions"
	CAColumnClicks                         = "clicks"
	CAColumnTotalCost                      = "total_cost"
	CAColumnConversions                    = "conversions"
	CAColumnAllConversions                 = "all_conversions"
	CAColumnCostPerClick                   = "cost_per_click"
	CAColumnConversionRate                 = "conversion_rate"
	CAColumnCostPerConversion              = "cost_per_conversion"
	CAColumnFrequency                      = "frequency"
	CAColumnReach                          = "reach"
	CAColumnInlinePostEngagement           = "inline_post_engagement"
	CAColumnUniqueClicks                   = "unique_clicks"
	CAColumnUniqueImpressions              = "approximateUniqueImpressions"
	CAColumnName                           = "name"
	CAColumnPlatform                       = "platform"
	CAFilterCampaign                       = "campaign"
	CAFilterAdGroup                        = "ad_group"
	CAFilterAd                             = "ad"
	CAFilterKeyword                        = "keyword"
	CAFilterQuery                          = "query"
	CAFilterAdset                          = "adset"
	CAChannelGoogleAds                     = "google_ads"
	CAChannelFacebookAds                   = "facebook_ads"
	CAChannelLinkedinAds                   = "linkedin_ads"
	CAChannelSearchConsole                 = "search_console"
	CAAllChannelAds                        = "all_ads"
	CAColumnValueAll                       = "all"
	CAChannelGroupKey                      = "group_key"
	innerJoinClause                        = " INNER JOIN "
	channeAnalyticsLimit                   = " LIMIT 10000 "
	source                                 = "source"
	CAColumnLikes                          = "likes"
	CAColumnFollows                        = "follows"
	CAColumnConversionValueInLocalCurrency = "conversion_value_in_local_currency"
	CAColumnTotalEngagement                = "total_engagements"
	CAFilterCampaignGroup                  = "campaign_group"
	CAFilterCreactive                      = "creative"
	CAFilterOrganicProperty                = "organic_property"
	dateTruncateString                     = "date_trunc('%s', CONVERT_TZ(TO_DATE(%s, 'YYYYMMDD'), 'UTC', '%s'))"
	dateTruncateWeekString                 = "date_trunc('WEEK', CONVERT_TZ(timestampadd(DAY, 1, TO_DATE(%s, 'YYYYMMDD')), 'UTC', '%s')) - INTERVAL 1 day"
	CAUnionFilterQuery                     = "SELECT filter_value from ( %s ) all_ads LIMIT 2500"
	CAUnionQuery1                          = "SELECT %s FROM ( %s ) all_ads ORDER BY %s %s"
	CAUnionQuery2                          = "SELECT %s FROM ( %s ) all_ads GROUP BY %s ORDER BY %s %s"
	CAUnionQuery3                          = "SELECT %s FROM ( %s ) all_ads GROUP BY %s ORDER BY %s %s"
	integrationNotAvailable                = "Document integration not available for this project."
	channelTimestamp                       = "timestamp"
)

var CAChannels = []string{
	CAChannelGoogleAds,
	CAChannelFacebookAds,
	CAChannelLinkedinAds,
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
	CAFilterCampaignGroup,
	CAFilterCreactive,
	CAFilterOrganicProperty,
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
func (store *MemSQL) GetChannelConfig(projectID uint64, channel string, reqID string) (*model.ChannelConfigResult, int) {
	if !(isValidChannel(channel)) {
		return &model.ChannelConfigResult{}, http.StatusBadRequest
	}

	var result *model.ChannelConfigResult
	switch channel {
	case CAAllChannelAds:
		result = store.buildAllChannelConfig(projectID)
	case CAChannelFacebookAds:
		result = store.buildFbChannelConfig(projectID)
	case CAChannelGoogleAds:
		result = store.buildAdwordsChannelConfig(projectID)
	case CAChannelLinkedinAds:
		result = store.buildLinkedinChannelConfig(projectID)
	case CAChannelSearchConsole:
		result = store.buildGoogleOrganicChannelConfig()
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

	return channel == CAChannelSearchConsole
}

// @TODO Kark v1
func (store *MemSQL) buildAllChannelConfig(projectID uint64) *model.ChannelConfigResult {
	objectsAndProperties := store.buildObjectAndPropertiesForAllChannel(projectID, objectsForAllChannels)

	return &model.ChannelConfigResult{
		SelectMetrics:        selectableMetricsForAllChannels,
		ObjectsAndProperties: objectsAndProperties,
	}
}
func (store *MemSQL) buildObjectAndPropertiesForAllChannel(projectID uint64, objects []string) []model.ChannelObjectAndProperties {
	objectsAndProperties := make([]model.ChannelObjectAndProperties, 0, 0)
	for _, currentObject := range objects {
		currentProperties := buildProperties(allChannelsPropertyToRelated)
		smartProperty := store.GetSmartPropertyAndRelated(projectID, currentObject, "all")
		currentPropertiesSmart := buildProperties(smartProperty)
		currentProperties = append(currentProperties, currentPropertiesSmart...)
		objectsAndProperties = append(objectsAndProperties, buildObjectsAndProperties(currentProperties, []string{currentObject})...)
	}
	return objectsAndProperties
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
func (store *MemSQL) GetChannelFilterValuesV1(projectID uint64, channel, filterObject,
	filterProperty string, reqID string) (model.ChannelFilterValues, int) {

	var channelFilterValues model.ChannelFilterValues
	if !isValidChannel(channel) || !isValidFilterKey(filterObject) {
		return channelFilterValues, http.StatusBadRequest
	}

	var filterValues []interface{}
	var errCode int
	switch channel {
	case CAAllChannelAds:
		filterValues, errCode = store.GetAllChannelFilterValues(projectID, filterObject, filterProperty, reqID)
	case CAChannelFacebookAds:
		filterValues, errCode = store.GetFacebookFilterValues(projectID, filterObject, filterProperty, reqID)
	case CAChannelGoogleAds:
		filterValues, errCode = store.GetAdwordsFilterValues(projectID, filterObject, filterProperty, reqID)
	case CAChannelLinkedinAds:
		filterValues, errCode = store.GetLinkedinFilterValues(projectID, filterObject, filterProperty, reqID)
	case CAChannelSearchConsole:
		filterValues, errCode = store.GetGoogleOrganicFilterValues(projectID, filterObject, filterProperty, reqID)
	}

	if errCode != http.StatusFound {
		return channelFilterValues, http.StatusInternalServerError
	}
	channelFilterValues.FilterValues = filterValues

	return channelFilterValues, http.StatusFound
}

// GetAllChannelFilterValues - @Kark TODO v1
func (store *MemSQL) GetAllChannelFilterValues(projectID uint64, filterObject, filterProperty string, reqID string) ([]interface{}, int) {
	logCtx := log.WithField("project_id", projectID).WithField("req_id", reqID)
	_, isPresent := Const.SmartPropertyReservedNames[filterProperty]
	if !isPresent {
		filterValues, errCode := store.getSmartPropertyFilterValues(projectID, filterObject, filterProperty, "all", reqID)
		if errCode != http.StatusFound {
			return []interface{}{}, http.StatusInternalServerError
		}
		return filterValues, http.StatusFound
	}
	adwordsSQL, adwordsParams, adwordsErr := store.GetAdwordsSQLQueryAndParametersForFilterValues(projectID, filterObject, filterProperty, reqID)
	facebookSQL, facebookParams, facebookErr := store.GetFacebookSQLQueryAndParametersForFilterValues(projectID, filterObject, filterProperty, reqID)
	linkedinSQL, linkedinParams, linkedinErr := store.GetLinkedinSQLQueryAndParametersForFilterValues(projectID, filterObject, filterProperty, reqID)

	if adwordsErr != http.StatusFound && adwordsErr != http.StatusNotFound {
		return []interface{}{}, adwordsErr
	}
	if facebookErr != http.StatusFound && facebookErr != http.StatusNotFound {
		return []interface{}{}, facebookErr
	}
	if linkedinErr != http.StatusFound && linkedinErr != http.StatusNotFound {
		return []interface{}{}, linkedinErr
	}

	finalSQLs := U.AppendNonNullValues(adwordsSQL, facebookSQL, linkedinSQL)
	finalParams := append(adwordsParams, facebookParams...)
	finalParams = append(finalParams, linkedinParams...)

	finalQuery := fmt.Sprintf(CAUnionFilterQuery, joinWithWordInBetween("UNION", finalSQLs...))
	_, resultRows, _ := store.ExecuteSQL(finalQuery, finalParams, logCtx)

	return Convert2DArrayTo1DArray(resultRows), http.StatusFound
}

// RunChannelGroupQuery - @TODO Kark v1
func (store *MemSQL) RunChannelGroupQuery(projectID uint64, queriesOriginal []model.ChannelQueryV1, reqID string) (model.ChannelResultGroupV1, int) {
	queries := make([]model.ChannelQueryV1, 0, 0)
	U.DeepCopy(&queriesOriginal, &queries)

	var resultGroup model.ChannelResultGroupV1
	resultGroup.Results = make([]model.ChannelQueryResultV1, len(queries))
	var waitGroup sync.WaitGroup
	count := 0
	waitGroup.Add(U.MinInt(len(queries), AllowedGoroutines))
	for index, query := range queries {
		count++
		go store.runSingleChannelQuery(projectID, query, &resultGroup.Results[index], &waitGroup, reqID)
		if count%AllowedGoroutines == 0 {
			waitGroup.Wait()
			waitGroup.Add(U.MinInt(len(queries)-count, AllowedGoroutines))
		}
	}
	waitGroup.Wait()
	for _, result := range resultGroup.Results {
		if result.Headers[0] == model.AliasError {
			return resultGroup, http.StatusPartialContent
		}
	}
	return resultGroup, http.StatusOK
}

// @Kark TODO v1
// TODO Handling errorcase.
func (store *MemSQL) runSingleChannelQuery(projectID uint64, query model.ChannelQueryV1,
	resultHolder *model.ChannelQueryResultV1, waitGroup *sync.WaitGroup, reqID string) {

	environment := C.GetConfig().Env
	defer waitGroup.Done()
	defer U.NotifyOnPanicWithError(environment, "app_server")
	tempResultHolder, _ := store.ExecuteChannelQueryV1(projectID, &query, reqID)
	resultHolder.Headers = tempResultHolder.Headers
	resultHolder.Rows = tempResultHolder.Rows
}

// ExecuteChannelQueryV1 - @Kark TODO v1
// TODO error handling.
func (store *MemSQL) ExecuteChannelQueryV1(projectID uint64, query *model.ChannelQueryV1,
	reqID string) (*model.ChannelQueryResultV1, int) {

	logCtx := log.WithField("req_id", reqID)
	queryResult := &model.ChannelQueryResultV1{}
	var columns []string
	var resultMetrics [][]interface{}
	status := http.StatusOK
	var err int
	if !(isValidChannel(query.Channel)) {
		return queryResult, http.StatusBadRequest
	}
	switch query.Channel {
	case CAAllChannelAds:
		columns, resultMetrics, err = store.executeAllChannelsQueryV1(projectID, query, reqID)
	case CAChannelFacebookAds:
		columns, resultMetrics, err = store.ExecuteFacebookChannelQueryV1(projectID, query, reqID)
	case CAChannelGoogleAds:
		columns, resultMetrics, err = store.ExecuteAdwordsChannelQueryV1(projectID, query, reqID)
	case CAChannelLinkedinAds:
		columns, resultMetrics, err = store.ExecuteLinkedinChannelQueryV1(projectID, query, reqID)
	case CAChannelSearchConsole:
		columns, resultMetrics, err = store.ExecuteGoogleOrganicChannelQueryV1(projectID, query, reqID)
	}
	if err != http.StatusOK {
		logCtx.Warn(query)
		errorResult := model.BuildErrorResultForChannelsV1("")
		return errorResult, http.StatusBadRequest
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

// removed source as we want aggregated results for all channels
func (store *MemSQL) executeAllChannelsQueryV1(projectID uint64, query *model.ChannelQueryV1,
	reqID string) ([]string, [][]interface{}, int) {

	logCtx := log.WithField("project_id", projectID).WithField("req_id", reqID)
	var finalQuery string
	var finalParams []interface{}
	var selectMetrics, columns []string
	isGroupByTimestamp := query.GetGroupByTimestamp() != ""

	projectSetting, errCode := store.GetProjectSetting(projectID)
	if errCode != http.StatusFound {
		return make([]string, 0, 0), [][]interface{}{}, http.StatusNotFound
	} else if (projectSetting.IntAdwordsCustomerAccountId == nil || *projectSetting.IntAdwordsCustomerAccountId == "") &&
		(projectSetting.IntFacebookAdAccount == "") && (projectSetting.IntLinkedinAdAccount == "") {
		log.Warn("Integration not present for channels.")
		return make([]string, 0, 0), [][]interface{}{}, http.StatusNotFound
	}

	if (query.GroupBy == nil || len(query.GroupBy) == 0) && (query.GroupByTimestamp == nil || len(query.GroupByTimestamp.(string)) == 0) {
		adwordsSQL, adwordsParams, commonKeys, commonMetrics, facebookSQL, facebookParams, linkedinSQL, linkedinParams, err := store.getIndividualChannelsSQLAndParametersV1(projectID, query, reqID, false)
		if err != http.StatusOK {
			return make([]string, 0, 0), [][]interface{}{}, err
		}
		finalSQLs := U.AppendNonNullValues(adwordsSQL, facebookSQL, linkedinSQL)
		finalParams = append(adwordsParams, facebookParams...)
		finalParams = append(finalParams, linkedinParams...)

		for _, metric := range commonMetrics {
			value := fmt.Sprintf("%s(%s) as %s", channelMetricsToOperation[metric], metric, metric)
			selectMetrics = append(selectMetrics, value)
		}
		finalQuery = fmt.Sprintf(CAUnionQuery1, joinWithComma(selectMetrics...), joinWithWordInBetween("UNION", finalSQLs...),
			getOrderByClause(isGroupByTimestamp, commonMetrics), channeAnalyticsLimit)
		columns = append(commonKeys, commonMetrics...)
	} else if (query.GroupBy == nil || len(query.GroupBy) == 0) && (!(query.GroupByTimestamp == nil || len(query.GroupByTimestamp.(string)) == 0)) {
		adwordsSQL, adwordsParams, commonKeys, commonMetrics, facebookSQL, facebookParams, linkedinSQL, linkedinParams, err := store.getIndividualChannelsSQLAndParametersV1(projectID, query, reqID, false)
		if err != http.StatusOK {
			return make([]string, 0, 0), [][]interface{}{}, err
		}
		finalSQLs := U.AppendNonNullValues(adwordsSQL, facebookSQL, linkedinSQL)
		finalParams = append(adwordsParams, facebookParams...)
		finalParams = append(finalParams, linkedinParams...)

		selectMetrics = append(selectMetrics, model.AliasDateTime)
		for _, metric := range commonMetrics {
			value := fmt.Sprintf("%s(%s) as %s", channelMetricsToOperation[metric], metric, metric)
			selectMetrics = append(selectMetrics, value)
		}
		finalQuery = fmt.Sprintf(CAUnionQuery2, joinWithComma(selectMetrics...), joinWithWordInBetween("UNION", finalSQLs...),
			model.AliasDateTime, getOrderByClause(isGroupByTimestamp, commonMetrics), channeAnalyticsLimit)
		columns = append(commonKeys, commonMetrics...)
	} else {
		adwordsSQL, adwordsParams, commonKeys, commonMetrics, facebookSQL, facebookParams, linkedinSQL, linkedinParams, err := store.getIndividualChannelsSQLAndParametersV1(projectID, query, reqID, true)
		if err != http.StatusOK {
			return make([]string, 0, 0), [][]interface{}{}, err
		}
		finalSQLs := U.AppendNonNullValues(adwordsSQL, facebookSQL, linkedinSQL)
		finalParams = append(adwordsParams, facebookParams...)
		finalParams = append(finalParams, linkedinParams...)
		selectMetrics = append(selectMetrics, commonKeys...)
		for _, metric := range commonMetrics {
			value := fmt.Sprintf("%s(%s) as %s", channelMetricsToOperation[metric], metric, metric)
			selectMetrics = append(selectMetrics, value)
		}

		finalQuery = fmt.Sprintf(CAUnionQuery3, joinWithComma(selectMetrics...), joinWithWordInBetween("UNION", finalSQLs...), joinWithComma(commonKeys...), getOrderByClause(isGroupByTimestamp, commonMetrics), channeAnalyticsLimit)
		columns = append(commonKeys, commonMetrics...)
	}
	_, resultMetrics, err := store.ExecuteSQL(finalQuery, finalParams, logCtx)
	if err != nil {
		logCtx.WithError(err).Error("Failed in channel analytics with following error.")
		return make([]string, 0, 0), make([][]interface{}, 0, 0), http.StatusInternalServerError
	}
	return columns, resultMetrics, http.StatusOK
}

func (store *MemSQL) getIndividualChannelsSQLAndParametersV1(projectID uint64, query *model.ChannelQueryV1, reqID string, fetchSource bool) (string, []interface{}, []string, []string, string, []interface{}, string, []interface{}, int) {
	adwordsSQL, adwordsParams, adwordsSelectKeys, adwordsMetrics, adwordsErr := store.GetSQLQueryAndParametersForAdwordsQueryV1(projectID, query, reqID, fetchSource)
	facebookSQL, facebookParams, facebookSelectKeys, facebookMetrics, facebookErr := store.GetSQLQueryAndParametersForFacebookQueryV1(projectID, query, reqID, fetchSource)
	linkedinSQL, linkedinParams, linkedinSelectKeys, linkedinMetrics, linkedinErr := store.GetSQLQueryAndParametersForLinkedinQueryV1(projectID, query, reqID, fetchSource)
	finalKeys := make([]string, 0, 0)
	finalMetrics := make([]string, 0, 0)

	if adwordsErr != http.StatusOK && adwordsErr != http.StatusNotFound {
		return "", []interface{}{}, make([]string, 0, 0), make([]string, 0, 0), "", []interface{}{}, "", []interface{}{}, adwordsErr
	}
	if facebookErr != http.StatusOK && facebookErr != http.StatusNotFound {
		return "", []interface{}{}, make([]string, 0, 0), make([]string, 0, 0), "", []interface{}{}, "", []interface{}{}, facebookErr
	}
	if linkedinErr != http.StatusOK && linkedinErr != http.StatusNotFound {
		return "", []interface{}{}, make([]string, 0, 0), make([]string, 0, 0), "", []interface{}{}, "", []interface{}{}, linkedinErr
	}
	if len(adwordsSQL) > 0 {
		finalKeys = adwordsSelectKeys
		finalMetrics = adwordsMetrics
		adwordsSQL = fmt.Sprintf("( %s )", adwordsSQL[:len(adwordsSQL)-2])
	}
	if len(facebookSQL) > 0 {
		finalKeys = facebookSelectKeys
		finalMetrics = facebookMetrics
		facebookSQL = fmt.Sprintf("( %s )", facebookSQL[:len(facebookSQL)-2])
	}
	if len(linkedinSQL) > 0 {
		finalKeys = linkedinSelectKeys
		finalMetrics = linkedinMetrics
		linkedinSQL = fmt.Sprintf("( %s )", linkedinSQL[:len(linkedinSQL)-2])
	}
	return adwordsSQL, adwordsParams, finalKeys, finalMetrics, facebookSQL, facebookParams, linkedinSQL, linkedinParams, http.StatusOK
}

// GetChannelFilterValues - @Kark TODO v0
func (store *MemSQL) GetChannelFilterValues(projectID uint64, channel, filter string) ([]string, int) {
	if !isValidChannel(channel) || !isValidFilterKey(filter) {
		return []string{}, http.StatusBadRequest
	}

	// supports only adwords now.
	docType, err := GetAdwordsDocumentTypeForFilterKey(filter)
	if err != nil {
		log.WithError(err).Error("Failed in channel filters with following error.")
		return []string{}, http.StatusInternalServerError
	}

	filterValues, errCode := store.GetAdwordsFilterValuesByType(projectID, docType)
	if errCode != http.StatusFound {
		return []string{}, http.StatusInternalServerError
	}

	return filterValues, http.StatusBadRequest
}

// ExecuteChannelQuery - @Kark TODO v0
func (store *MemSQL) ExecuteChannelQuery(projectID uint64,
	queryOriginal *model.ChannelQuery) (*model.ChannelQueryResult, int) {

	var query *model.ChannelQuery
	U.DeepCopy(queryOriginal, &query)

	if !isValidChannel(query.Channel) || !isValidFilterKey(query.FilterKey) ||
		query.From == 0 || query.To == 0 {
		return nil, http.StatusBadRequest
	}

	if query.Channel == "google_ads" {
		result, errCode := store.ExecuteAdwordsChannelQuery(projectID, query)
		if errCode != http.StatusOK {
			log.WithField("project_id", projectID).Error("Failed to execute adwords channel query.")
			return nil, http.StatusInternalServerError
		}
		return result, http.StatusOK
	}
	if query.Channel == "facebook_ads" {
		result, errCode := store.ExecuteFacebookChannelQuery(projectID, query)
		if errCode != http.StatusOK {
			log.WithField("project_id", projectID).Error("Failed to execute facebook channel query.")
			return nil, http.StatusInternalServerError
		}
		return result, http.StatusOK
	}
	if query.Channel == "linkedin_ads" {
		result, errCode := store.ExecuteLinkedinChannelQuery(projectID, query)
		if errCode != http.StatusOK {
			log.WithField("project_id", projectID).Error("Failed to execute linkedin channel query.")
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

// format yyyymmdd
func ChangeUnixTimestampToDate(timestamp int64) int64 {
	time := time.Unix(timestamp, 0)
	date, _ := strconv.ParseInt(time.Format("20060102"), 10, 64)
	return date
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
		selectStr = fmt.Sprintf(dateTruncateString, "hour", channelTimestamp, selectTz)
	} else if timestampType == model.GroupByTimestampWeek {
		selectStr = fmt.Sprintf(dateTruncateWeekString, channelTimestamp, selectTz)
	} else if timestampType == model.GroupByTimestampMonth {
		selectStr = fmt.Sprintf(dateTruncateString, "month", channelTimestamp, selectTz)
	} else if timestampType == model.GroupByTimestampQuarter {
		selectStr = fmt.Sprintf(dateTruncateString, "quarter", channelTimestamp, selectTz)
	} else {
		// defaults to GroupByTimestampDate.
		selectStr = fmt.Sprintf(dateTruncateString, "day", channelTimestamp, selectTz)
	}

	return selectStr
}

// @Kark TODO v1
func getOrderByClause(isGroupByTimestamp bool, selectMetrics []string) string {
	selectMetricsWithDesc := make([]string, 0, 0)
	if isGroupByTimestamp {
		selectMetricsWithDesc = append(selectMetricsWithDesc, model.AliasDateTime+" ASC")
	} else {
		for _, selectMetric := range selectMetrics {
			selectMetricsWithDesc = append(selectMetricsWithDesc, selectMetric+" DESC")
		}
	}
	return joinWithComma(selectMetricsWithDesc...)
}

// ExecuteSQL - @Kark TODO v1
func (store *MemSQL) ExecuteSQL(sqlStatement string, params []interface{}, logCtx *log.Entry) ([]string, [][]interface{}, error) {
	rows, err := store.ExecQueryWithContext(sqlStatement, params)
	if err != nil {
		logCtx.WithError(err).WithField("query", sqlStatement).WithField("params", params).Error("SQL Query failed.")
		return nil, nil, err
	}

	defer rows.Close()
	columns, resultRows, err := U.DBReadRows(rows)

	if err != nil {
		return nil, nil, err
	}
	if len(resultRows) == 0 {
		logCtx.Warn("Aggregate query returned zero rows: ", sqlStatement, params)
		return nil, make([][]interface{}, 0, 0), nil
	}
	return columns, resultRows, nil
}

func (store *MemSQL) GetSmartPropertyAndRelated(projectID uint64, object string, source string) map[string]PropertiesAndRelated {
	db := C.GetServices().Db
	var smartPropertyRules []model.SmartPropertyRules
	object_type, isPresent := model.SmartPropertyRulesTypeAliasToType[object]
	if !isPresent {
		return nil
	}
	err := db.Table("smart_property_rules").Where("project_id = ? AND type = ? and is_deleted = ?", projectID, object_type, false).Find(&smartPropertyRules).Error
	if err != nil {
		log.Error("Failed to get smart property filters from DB")
	}

	if len(smartPropertyRules) == 0 {
		return nil
	}
	smartPropertyFilterConfig := make(map[string]PropertiesAndRelated)
	for _, smartPropertyRule := range smartPropertyRules {
		var rules []model.Rule
		err := U.DecodePostgresJsonbToStructType(smartPropertyRule.Rules, &rules)
		if err != nil {
			continue
		}
		for _, rule := range rules {
			if rule.Source == "all" || rule.Source == source {
				smartPropertyFilterConfig[smartPropertyRule.Name] = PropertiesAndRelated{typeOfProperty: U.PropertyTypeCategorical}
				break
			}
		}
	}

	return smartPropertyFilterConfig
}
