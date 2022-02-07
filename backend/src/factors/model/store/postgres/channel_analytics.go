package postgres

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
	CAFilterCreative                       = "creative"
	CAFilterOrganicProperty                = "organic_property"
	CAFilterChannel                        = "channel"
	dateTruncateString                     = "date_trunc('%s', make_timestamptz(SUBSTRING (%s::text, 1, 4)::INTEGER, SUBSTRING (%s::text, 5, 2)::INTEGER, SUBSTRING (%s::text, 7, 2)::INTEGER, 0, 0, 0, '%s') AT TIME ZONE '%s')"
	dateTruncateWeekString                 = "date_trunc('WEEK', (make_timestamptz(SUBSTRING (%s::text, 1, 4)::INTEGER, SUBSTRING (%s::text, 5, 2)::INTEGER, SUBSTRING (%s::text, 7, 2)::INTEGER, 0, 0, 0, '%s') + INTERVAL '1 day') AT TIME ZONE '%s') - INTERVAL '1 day'"
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
	CAFilterCreative,
	CAFilterOrganicProperty,
	CAFilterChannel,
}

// TODO: Move and fetch it from respective channels - allChannels, adwords etc.. because this is error prone.
// to do change kpi analytics/all channels, linkedin when we change below.
var selectableMetricsForAllChannels = []string{"impressions", "clicks", "spend"}
var objectsForAllChannels = []string{CAFilterCampaign, CAFilterAdGroup}

var allChannelsPropertyToRelated = map[string]model.PropertiesAndRelated{
	"name": model.PropertiesAndRelated{
		TypeOfProperty: U.PropertyTypeCategorical,
	},
	"id": model.PropertiesAndRelated{
		TypeOfProperty: U.PropertyTypeCategorical,
	},
}

// GetChannelConfig - @TODO Kark v1
func (pg *Postgres) GetChannelConfig(projectID uint64, channel string, reqID string) (*model.ChannelConfigResult, int) {
	if !(isValidChannel(channel)) {
		return &model.ChannelConfigResult{}, http.StatusBadRequest
	}

	var result *model.ChannelConfigResult
	switch channel {
	case CAAllChannelAds:
		result = pg.buildAllChannelConfig(projectID)
	case CAChannelFacebookAds:
		result = pg.buildFbChannelConfig(projectID)
	case CAChannelGoogleAds:
		result = pg.buildAdwordsChannelConfig(projectID)
	case CAChannelLinkedinAds:
		result = pg.buildLinkedinChannelConfig(projectID)
	case CAChannelSearchConsole:
		result = pg.buildGoogleOrganicChannelConfig()
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
func (pg *Postgres) buildAllChannelConfig(projectID uint64) *model.ChannelConfigResult {
	objectsAndProperties := pg.buildObjectAndPropertiesForAllChannel(projectID, objectsForAllChannels)

	return &model.ChannelConfigResult{
		SelectMetrics:        selectableMetricsForAllChannels,
		ObjectsAndProperties: objectsAndProperties,
	}
}
func (pg *Postgres) buildObjectAndPropertiesForAllChannel(projectID uint64, objects []string) []model.ChannelObjectAndProperties {
	objectsAndProperties := make([]model.ChannelObjectAndProperties, 0, 0)
	for _, currentObject := range objects {
		currentProperties := buildProperties(allChannelsPropertyToRelated)
		smartProperty := pg.GetSmartPropertyAndRelated(projectID, currentObject, "all")
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
func buildProperties(PropertiesAndRelated map[string]model.PropertiesAndRelated) []model.ChannelProperty {
	var properties []model.ChannelProperty
	for propertyName, propertyRelated := range PropertiesAndRelated {
		var property model.ChannelProperty
		property.Name = propertyName
		property.Type = propertyRelated.TypeOfProperty
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
	case CAChannelLinkedinAds:
		filterValues, errCode = pg.GetLinkedinFilterValues(projectID, filterObject, filterProperty, reqID)
	case CAChannelSearchConsole:
		filterValues, errCode = pg.GetGoogleOrganicFilterValues(projectID, filterObject, filterProperty, reqID)
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
	_, isPresent := Const.SmartPropertyReservedNames[filterProperty]
	if !isPresent {
		filterValues, errCode := pg.getSmartPropertyFilterValues(projectID, filterObject, filterProperty, "all", reqID)
		if errCode != http.StatusFound {
			return []interface{}{}, http.StatusInternalServerError
		}
		return filterValues, http.StatusFound
	}
	if filterObject == CAFilterChannel && filterProperty == "name" {
		return []interface{}{"adwords", "facebook", "linkedin"}, http.StatusFound
	}
	adwordsSQL, adwordsParams, adwordsErr := pg.GetAdwordsSQLQueryAndParametersForFilterValues(projectID, filterObject, filterProperty, reqID)
	facebookSQL, facebookParams, facebookErr := pg.GetFacebookSQLQueryAndParametersForFilterValues(projectID, filterObject, filterProperty, reqID)
	linkedinSQL, linkedinParams, linkedinErr := pg.GetLinkedinSQLQueryAndParametersForFilterValues(projectID, filterObject, filterProperty, reqID)

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
	_, resultRows, err := pg.ExecuteSQL(finalQuery, finalParams, logCtx)
	logCtx.Warn(err)

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
	actualRoutineLimit := U.MinInt(len(queries), AllowedGoroutines)
	waitGroup.Add(actualRoutineLimit)
	for index, query := range queries {
		count++
		go pg.runSingleChannelQuery(projectID, query, &resultGroup.Results[index], &waitGroup, reqID)
		if count%actualRoutineLimit == 0 {
			waitGroup.Wait()
			waitGroup.Add(U.MinInt(len(queries)-count, actualRoutineLimit))
		}
	}
	waitGroup.Wait()
	// log.Warn(resultGroup.Results)
	for _, result := range resultGroup.Results {
		if result.Headers[0] == model.AliasError {
			// log.Warn(resultGroup)
			return resultGroup, http.StatusPartialContent
		}
	}
	return resultGroup, http.StatusOK
}

// @Kark TODO v1
// TODO Handling errorcase.
func (pg *Postgres) runSingleChannelQuery(projectID uint64, query model.ChannelQueryV1,
	resultHolder *model.ChannelQueryResultV1, waitGroup *sync.WaitGroup, reqID string) {
	defer waitGroup.Done()
	tempResultHolder, _ := pg.ExecuteChannelQueryV1(projectID, &query, reqID)
	resultHolder.Headers = tempResultHolder.Headers
	resultHolder.Rows = tempResultHolder.Rows
}

// ExecuteChannelQueryV1 - @Kark TODO v1
// TODO error handling.
func (pg *Postgres) ExecuteChannelQueryV1(projectID uint64, query *model.ChannelQueryV1,
	reqID string) (*model.ChannelQueryResultV1, int) {
	defer U.NotifyOnPanicWithError(C.GetConfig().Env, C.GetConfig().AppName)

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
		columns, resultMetrics, err = pg.executeAllChannelsQueryV1(projectID, query, reqID)
	case CAChannelFacebookAds:
		columns, resultMetrics, err = pg.ExecuteFacebookChannelQueryV1(projectID, query, reqID)
	case CAChannelGoogleAds:
		columns, resultMetrics, err = pg.ExecuteAdwordsChannelQueryV1(projectID, query, reqID)
	case CAChannelLinkedinAds:
		columns, resultMetrics, err = pg.ExecuteLinkedinChannelQueryV1(projectID, query, reqID)
	case CAChannelSearchConsole:
		columns, resultMetrics, err = pg.ExecuteGoogleOrganicChannelQueryV1(projectID, query, reqID)
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
func (pg *Postgres) executeAllChannelsQueryV1(projectID uint64, query *model.ChannelQueryV1,
	reqID string) ([]string, [][]interface{}, int) {
	defer U.NotifyOnPanicWithError(C.GetConfig().Env, C.GetConfig().AppName)

	logCtx := log.WithField("project_id", projectID).WithField("req_id", reqID)
	var finalQuery string
	var finalParams []interface{}
	var selectMetrics, columns []string
	isGroupByTimestamp := query.GetGroupByTimestamp() != ""

	projectSetting, errCode := pg.GetProjectSetting(projectID)
	if errCode != http.StatusFound {
		headers := model.GetHeadersFromQuery(*query)
		return headers, make([][]interface{}, 0, 0), http.StatusNotFound
	} else if (projectSetting.IntAdwordsCustomerAccountId == nil || *projectSetting.IntAdwordsCustomerAccountId == "") &&
		(projectSetting.IntFacebookAdAccount == "") && (projectSetting.IntLinkedinAdAccount == "") {
		log.Warn("Integration not present for channels.")
		headers := model.GetHeadersFromQuery(*query)
		return headers, make([][]interface{}, 0, 0), http.StatusOK
	}

	if (query.GroupBy == nil || len(query.GroupBy) == 0) && (query.GroupByTimestamp == nil || len(query.GroupByTimestamp.(string)) == 0) {
		adwordsSQL, adwordsParams, commonKeys, commonMetrics, facebookSQL, facebookParams, linkedinSQL, linkedinParams, err := pg.getIndividualChannelsSQLAndParametersV1(projectID, query, reqID, false)
		if errCode == http.StatusNotFound {
			headers := model.GetHeadersFromQuery(*query)
			return headers, make([][]interface{}, 0, 0), http.StatusOK
		}
		if err != http.StatusOK {
			headers := model.GetHeadersFromQuery(*query)
			return headers, make([][]interface{}, 0, 0), err
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
		adwordsSQL, adwordsParams, commonKeys, commonMetrics, facebookSQL, facebookParams, linkedinSQL, linkedinParams, err := pg.getIndividualChannelsSQLAndParametersV1(projectID, query, reqID, false)
		if errCode == http.StatusNotFound {
			headers := model.GetHeadersFromQuery(*query)
			return headers, make([][]interface{}, 0, 0), http.StatusOK
		}
		if err != http.StatusOK {
			headers := model.GetHeadersFromQuery(*query)
			return headers, make([][]interface{}, 0, 0), err
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
		adwordsSQL, adwordsParams, commonKeys, commonMetrics, facebookSQL, facebookParams, linkedinSQL, linkedinParams, err := pg.getIndividualChannelsSQLAndParametersV1(projectID, query, reqID, true)
		if errCode == http.StatusNotFound {
			headers := model.GetHeadersFromQuery(*query)
			return headers, make([][]interface{}, 0, 0), http.StatusOK
		}
		if err != http.StatusOK {
			headers := model.GetHeadersFromQuery(*query)
			return headers, make([][]interface{}, 0, 0), err
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

	_, resultMetrics, err := pg.ExecuteSQL(finalQuery, finalParams, logCtx)
	if err != nil {
		logCtx.WithError(err).Error("Failed in channel analytics with following error.")
		return columns, make([][]interface{}, 0, 0), http.StatusInternalServerError
	}
	return columns, resultMetrics, http.StatusOK
}

func (pg *Postgres) getIndividualChannelsSQLAndParametersV1(projectID uint64, query *model.ChannelQueryV1, reqID string, fetchSource bool) (string, []interface{}, []string, []string, string, []interface{}, string, []interface{}, int) {
	isGroupBytimestamp := query.GetGroupByTimestamp() != ""
	genericFilters, channelBreakdownFilters := model.GetDecoupledFiltersForChannelBreakdownFilters(query.Filters)
	query.Filters = genericFilters
	isAdwordsReq, isFacebookReq, isLinkedinReq, errCode := model.GetRequiredChannels(channelBreakdownFilters)
	if errCode != http.StatusOK {
		return "", []interface{}{}, make([]string, 0, 0), make([]string, 0, 0), "", []interface{}{}, "", []interface{}{}, errCode
	}

	finAdwordsSQL, finAdwordsParams, finFacebookSQL, finFacebookParams, finLinkedinSQL, finLinkedinParams := "", make([]interface{}, 0), "", make([]interface{}, 0), "", make([]interface{}, 0)

	finalKeys := make([]string, 0, 0)
	finalMetrics := make([]string, 0, 0)
	if isAdwordsReq {
		adwordsSQL, adwordsParams, adwordsSelectKeys, adwordsMetrics, adwordsErr := pg.GetSQLQueryAndParametersForAdwordsQueryV1(projectID, query, reqID, fetchSource, " LIMIT 10000", isGroupBytimestamp, nil)
		if adwordsErr != http.StatusOK && adwordsErr != http.StatusNotFound {
			return "", []interface{}{}, make([]string, 0, 0), make([]string, 0, 0), "", []interface{}{}, "", []interface{}{}, adwordsErr
		}
		if len(adwordsSQL) > 0 {
			finalKeys = adwordsSelectKeys
			finalMetrics = adwordsMetrics
			adwordsSQL = fmt.Sprintf("( %s )", adwordsSQL[:len(adwordsSQL)-2])
		}
		finAdwordsSQL, finAdwordsParams = adwordsSQL, adwordsParams
	}
	if isFacebookReq {
		facebookSQL, facebookParams, facebookSelectKeys, facebookMetrics, facebookErr := pg.GetSQLQueryAndParametersForFacebookQueryV1(projectID, query, reqID, fetchSource, " LIMIT 10000", isGroupBytimestamp, nil)
		if facebookErr != http.StatusOK && facebookErr != http.StatusNotFound {
			return "", []interface{}{}, make([]string, 0, 0), make([]string, 0, 0), "", []interface{}{}, "", []interface{}{}, facebookErr
		}
		if len(facebookSQL) > 0 {
			finalKeys = facebookSelectKeys
			finalMetrics = facebookMetrics
			facebookSQL = fmt.Sprintf("( %s )", facebookSQL[:len(facebookSQL)-2])
		}
		finFacebookSQL, finFacebookParams = facebookSQL, facebookParams
	}
	if isLinkedinReq {
		linkedinSQL, linkedinParams, linkedinSelectKeys, linkedinMetrics, linkedinErr := pg.GetSQLQueryAndParametersForLinkedinQueryV1(projectID, query, reqID, fetchSource, " LIMIT 10000", isGroupBytimestamp, nil)
		if linkedinErr != http.StatusOK && linkedinErr != http.StatusNotFound {
			return "", []interface{}{}, make([]string, 0, 0), make([]string, 0, 0), "", []interface{}{}, "", []interface{}{}, linkedinErr
		}
		if len(linkedinSQL) > 0 {
			finalKeys = linkedinSelectKeys
			finalMetrics = linkedinMetrics
			linkedinSQL = fmt.Sprintf("( %s )", linkedinSQL[:len(linkedinSQL)-2])
		}
		finLinkedinSQL, finLinkedinParams = linkedinSQL, linkedinParams
	}
	return finAdwordsSQL, finAdwordsParams, finalKeys, finalMetrics, finFacebookSQL, finFacebookParams, finLinkedinSQL, finLinkedinParams, http.StatusOK
}

// GetChannelFilterValues - @Kark TODO v0
func (pg *Postgres) GetChannelFilterValues(projectID uint64, channel, filter string) ([]string, int) {
	if !isValidChannel(channel) || !isValidFilterKey(filter) {
		return []string{}, http.StatusBadRequest
	}

	// supports only adwords now.
	docType, err := GetAdwordsDocumentTypeForFilterKey(filter)
	if err != nil {
		log.WithError(err).Error("Failed in channel filters with following error.")
		return []string{}, http.StatusInternalServerError
	}

	filterValues, errCode := pg.GetAdwordsFilterValuesByType(projectID, docType)
	if errCode != http.StatusFound {
		return []string{}, http.StatusInternalServerError
	}

	return filterValues, http.StatusBadRequest
}

// ExecuteChannelQuery - @Kark TODO v0
func (pg *Postgres) ExecuteChannelQuery(projectID uint64,
	queryOriginal *model.ChannelQuery) (*model.ChannelQueryResult, int) {
	defer U.NotifyOnPanicWithError(C.GetConfig().Env, C.GetConfig().AppName)
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
	if query.Channel == "linkedin_ads" {
		result, errCode := pg.ExecuteLinkedinChannelQuery(projectID, query)
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
	selectTz = model.DefaultTimezone

	if timestampType == model.GroupByTimestampHour {
		selectStr = fmt.Sprintf(dateTruncateString, "hour", channelTimestamp, channelTimestamp, channelTimestamp, selectTz, selectTz)
	} else if timestampType == model.GroupByTimestampWeek {
		selectStr = fmt.Sprintf(dateTruncateWeekString, channelTimestamp, channelTimestamp, channelTimestamp, selectTz, selectTz)
	} else if timestampType == model.GroupByTimestampMonth {
		selectStr = fmt.Sprintf(dateTruncateString, "month", channelTimestamp, channelTimestamp, channelTimestamp, selectTz, selectTz)
	} else if timestampType == model.GroupByTimestampQuarter {
		selectStr = fmt.Sprintf(dateTruncateString, "quarter", channelTimestamp, channelTimestamp, channelTimestamp, selectTz, selectTz)
	} else {
		// defaults to GroupByTimestampDate.
		selectStr = fmt.Sprintf(dateTruncateString, "day", channelTimestamp, channelTimestamp, channelTimestamp, selectTz, selectTz)
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
func getOrderByClauseForSearchConsole(isGroupByTimestamp bool, selectMetrics []string) string {
	selectMetricsWithDesc := make([]string, 0)
	if isGroupByTimestamp {
		selectMetricsWithDesc = append(selectMetricsWithDesc, model.AliasDateTime+" ASC")
	} else {
		for _, selectMetric := range selectMetrics {
			if model.AscendingOrderByMetricsForGoogleOrganic[selectMetric] {
				selectMetricsWithDesc = append(selectMetricsWithDesc, selectMetric+" ASC")
			} else {
				selectMetricsWithDesc = append(selectMetricsWithDesc, selectMetric+" DESC")
			}
		}
	}
	return joinWithComma(selectMetricsWithDesc...)
}

// ExecuteSQL - @Kark TODO v1
func (pg *Postgres) ExecuteSQL(sqlStatement string, params []interface{}, logCtx *log.Entry) ([]string, [][]interface{}, error) {
	rows, tx, err := pg.ExecQueryWithContext(sqlStatement, params)
	if err != nil {
		logCtx.WithError(err).WithField("query", sqlStatement).WithField("params", params).Error("SQL Query failed.")
		return nil, nil, err
	}
	defer U.CloseReadQuery(rows, tx)

	columns, resultRows, err := U.DBReadRows(rows, tx)
	if err != nil {
		return nil, nil, err
	}
	if len(resultRows) == 0 {
		logCtx.Warn("Aggregate query returned zero rows: ", sqlStatement, params)
		return nil, make([][]interface{}, 0, 0), nil
	}
	return columns, resultRows, nil
}

func (pg *Postgres) GetSmartPropertyAndRelated(projectID uint64, object string, source string) map[string]model.PropertiesAndRelated {
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
	smartPropertyFilterConfig := make(map[string]model.PropertiesAndRelated)
	for _, smartPropertyRule := range smartPropertyRules {
		var rules []model.Rule
		err := U.DecodePostgresJsonbToStructType(smartPropertyRule.Rules, &rules)
		if err != nil {
			continue
		}
		for _, rule := range rules {
			if rule.Source == "all" || rule.Source == source {
				smartPropertyFilterConfig[smartPropertyRule.Name] = model.PropertiesAndRelated{TypeOfProperty: U.PropertyTypeCategorical}
				break
			}
		}
	}

	return smartPropertyFilterConfig
}
