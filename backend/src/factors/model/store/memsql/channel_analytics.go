package memsql

import (
	C "factors/config"
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
	CAChannelBingAds                       = "bing_ads"
	CACustomAds                            = "custom_ads"
	CAChannelGoogleAds                     = "google_ads"
	CAChannelFacebookAds                   = "facebook_ads"
	CAChannelLinkedinAds                   = "linkedin_ads"
	CAChannelLinkedinCompanyEngagements    = "linkedin_company_engagements"
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
	CAFilterChannel                        = "channel"
	CAFilterCompany                        = "company"
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
	CAChannelLinkedinCompanyEngagements,
	CAAllChannelAds,
	CAChannelBingAds,
	CACustomAds,
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
	CAFilterChannel,
	CAFilterCompany,
}

// TODO: Move and fetch it from respective channels - allChannels, adwords etc.. because this is error prone.
var SelectableMetricsForAllChannels = []string{"impressions", "clicks", "spend"}
var ObjectsForAllChannels = []string{CAFilterCampaign, CAFilterAdGroup}

var allChannelsPropertyToRelated = map[string]model.PropertiesAndRelated{
	"name": model.PropertiesAndRelated{
		TypeOfProperty: U.PropertyTypeCategorical,
	},
	"id": model.PropertiesAndRelated{
		TypeOfProperty: U.PropertyTypeCategorical,
	},
}

// GetChannelConfig - @TODO Kark v1
func (store *MemSQL) GetChannelConfig(projectID int64, channel string, reqID string) (*model.ChannelConfigResult, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"channel":    channel,
		"req_id":     reqID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	sources := make([]string, 0)
	if !(isValidChannel(channel, sources)) {
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
	case CAChannelBingAds:
		result = store.buildBingAdsChannelConfig(projectID) // modify for bingads
	}
	return result, http.StatusOK
}

// @TODO Kark v0, v1
func isValidFilterKey(filter string) bool {
	logFields := log.Fields{
		"filter": filter,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	for _, f := range CAFilters {
		if filter == f {
			return true
		}
	}

	return false
}

// @TODO Kark v1
func isValidChannel(channel string, sources []string) bool {
	logFields := log.Fields{
		"channel": channel,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	for _, c := range CAChannels {
		if channel == c {
			return true
		}
	}
	for _, source := range sources {
		if source == channel {
			return true
		}
	}
	return channel == CAChannelSearchConsole
}

// @TODO Kark v1
func (store *MemSQL) buildAllChannelConfig(projectID int64) *model.ChannelConfigResult {
	logFields := log.Fields{
		"project_id": projectID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	objectsAndProperties := store.buildObjectAndPropertiesForAllChannel(projectID, ObjectsForAllChannels)

	return &model.ChannelConfigResult{
		SelectMetrics:        SelectableMetricsForAllChannels,
		ObjectsAndProperties: objectsAndProperties,
	}
}
func (store *MemSQL) buildObjectAndPropertiesForAllChannel(projectID int64, objects []string) []model.ChannelObjectAndProperties {
	logFields := log.Fields{
		"project_id": projectID,
		"objects":    objects,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	objectsAndProperties := make([]model.ChannelObjectAndProperties, 0, 0)

	objectsAndProperties = append(objectsAndProperties, model.ChannelNameProperty)
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
	logFields := log.Fields{
		"properties":          properties,
		"filter_object_names": filterObjectNames,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

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
	logFields := log.Fields{
		"properties_and_related": PropertiesAndRelated,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
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
func (store *MemSQL) GetChannelFilterValuesV1(projectID int64, channel, filterObject,
	filterProperty string, reqID string) (model.ChannelFilterValues, int) {

	logFields := log.Fields{
		"project_id":      projectID,
		"channel":         channel,
		"filter_property": filterObject,
		"req_id":          reqID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	var channelFilterValues model.ChannelFilterValues
	sources := make([]string, 0)
	if !isValidChannel(channel, sources) || !isValidFilterKey(filterObject) {
		return channelFilterValues, http.StatusBadRequest
	}

	var filterValues []interface{}
	var errCode int
	switch channel {
	case CAAllChannelAds:
		filterValues, errCode = store.GetAllChannelFilterValues(projectID, filterObject, filterProperty, channel, reqID)
	case CAChannelFacebookAds:
		filterValues, errCode = store.GetFacebookFilterValues(projectID, filterObject, filterProperty, reqID)
	case CAChannelGoogleAds:
		filterValues, errCode = store.GetAdwordsFilterValues(projectID, filterObject, filterProperty, reqID)
	case CAChannelLinkedinAds:
		filterValues, errCode = store.GetLinkedinFilterValues(projectID, filterObject, filterProperty, reqID)
	case CAChannelLinkedinCompanyEngagements:
		filterValues, errCode = store.GetLinkedinCompanyEngagementsFilterValues(projectID, filterObject, filterProperty, reqID)
	case CAChannelSearchConsole:
		filterValues, errCode = store.GetGoogleOrganicFilterValues(projectID, filterObject, filterProperty, reqID)
	case CAChannelBingAds:
		filterValues, errCode = store.GetBingadsFilterValues(projectID, filterObject, filterProperty, reqID)
	}

	if errCode != http.StatusFound {
		return channelFilterValues, http.StatusInternalServerError
	}
	channelFilterValues.FilterValues = filterValues

	return channelFilterValues, http.StatusFound
}

func (store *MemSQL) GetCustomChannelFilterValuesV1(projectID int64, source, channel, filterObject,
	filterProperty string, reqID string) (model.ChannelFilterValues, int) {

	logFields := log.Fields{
		"project_id":      projectID,
		"channel":         channel,
		"filter_property": filterObject,
		"req_id":          reqID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	var channelFilterValues model.ChannelFilterValues
	sources := make([]string, 0)
	if !isValidChannel(channel, sources) || !isValidFilterKey(filterObject) {
		return channelFilterValues, http.StatusBadRequest
	}

	var filterValues []interface{}
	var errCode int
	switch channel {
	case CACustomAds:
		filterValues, errCode = store.GetCustomAdsFilterValues(projectID, filterObject, filterProperty, source, reqID)
	}

	if errCode != http.StatusFound {
		return channelFilterValues, http.StatusInternalServerError
	}
	channelFilterValues.FilterValues = filterValues

	return channelFilterValues, http.StatusFound
}

// GetAllChannelFilterValues - @Kark TODO v1
func (store *MemSQL) GetAllChannelFilterValues(projectID int64, filterObject, filterProperty string, source string, reqID string) ([]interface{}, int) {
	logFields := log.Fields{
		"project_id":      projectID,
		"filter_object":   filterObject,
		"filter_property": filterProperty,
		"req_id":          reqID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)
	_, isPresent := model.SmartPropertyReservedNames[filterProperty]
	if !isPresent {
		filterValues, errCode := store.getSmartPropertyFilterValues(projectID, filterObject, filterProperty, "all", reqID)
		if errCode != http.StatusFound {
			return []interface{}{}, http.StatusInternalServerError
		}
		return filterValues, http.StatusFound
	}
	isBingAdsAvailable := store.IsBingIntegrationAvailable(projectID)
	IsCustomAdsAvailable := store.IsCustomAdsAvailable(projectID)
	if filterObject == CAFilterChannel && filterProperty == "name" {
		integrations := []interface{}{model.GoogleAds, model.FacebookAds, model.LinkedinAds}
		if isBingAdsAvailable {
			integrations = append(integrations, model.NewBingAds)
		}
		if IsCustomAdsAvailable {
			sources, _ := store.GetCustomAdsSourcesByProject(projectID)
			for _, source := range sources {
				integrations = append(integrations, source)
			}
		}
		return integrations, http.StatusFound
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
	if isBingAdsAvailable {
		bingAdsinSQL, bingAdsParams, bingAdsErr := store.GetBingadsFilterValuesSQLAndParams(projectID, filterObject, filterProperty, reqID)

		if bingAdsErr != http.StatusFound && bingAdsErr != http.StatusNotFound {
			return []interface{}{}, bingAdsErr
		}
		finalSQLs = U.AppendNonNullValues(adwordsSQL, facebookSQL, linkedinSQL, bingAdsinSQL)
		finalParams = append(finalParams, bingAdsParams...)
	}
	if IsCustomAdsAvailable {
		customAdsinSQL, customAdsParams, customAdsErr := store.GetCustomadsFilterValuesSQLAndParams(projectID, filterObject, filterProperty, source, reqID)

		if customAdsErr != http.StatusFound && customAdsErr != http.StatusNotFound {
			return []interface{}{}, customAdsErr
		}

		finalSQLs = U.AppendNonNullValuesList(finalSQLs, customAdsinSQL)
		finalParams = append(finalParams, customAdsParams...)
	}
	finalQuery := fmt.Sprintf(CAUnionFilterQuery, joinWithWordInBetween("UNION", finalSQLs...))
	_, resultRows, _ := store.ExecuteSQL(finalQuery, finalParams, logCtx)

	return Convert2DArrayTo1DArray(resultRows), http.StatusFound
}

// RunChannelGroupQuery - @TODO Kark v1
func (store *MemSQL) RunChannelGroupQuery(projectID int64, queriesOriginal []model.ChannelQueryV1, reqID string) (model.ChannelResultGroupV1, int) {
	logFields := log.Fields{
		"project_id":       projectID,
		"queries_original": queriesOriginal,
		"req_id":           reqID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
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
		go store.runSingleChannelQuery(projectID, query, &resultGroup.Results[index], &waitGroup, reqID)
		if count%actualRoutineLimit == 0 {
			waitGroup.Wait()
			waitGroup.Add(U.MinInt(len(queries)-count, actualRoutineLimit))
		}
	}
	waitGroup.Wait()
	for _, result := range resultGroup.Results {
		if result.Headers == nil {
			return resultGroup, http.StatusInternalServerError
		}
		if result.Headers[0] == model.AliasError {
			return resultGroup, http.StatusPartialContent
		}
	}
	return resultGroup, http.StatusOK
}

// @Kark TODO v1
// TODO Handling errorcase.
func (store *MemSQL) runSingleChannelQuery(projectID int64, query model.ChannelQueryV1,
	resultHolder *model.ChannelQueryResultV1, waitGroup *sync.WaitGroup, reqID string) {
	logFields := log.Fields{
		"project_id":    projectID,
		"query":         query,
		"result_holder": resultHolder,
		"wait_group":    waitGroup,
		"req_id":        reqID,
	}
	defer U.NotifyOnPanicWithError(C.GetConfig().Env, C.GetConfig().AppName)
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	defer waitGroup.Done()
	tempResultHolder, _ := store.ExecuteChannelQueryV1(projectID, &query, reqID)
	resultHolder.Headers = tempResultHolder.Headers
	resultHolder.Rows = tempResultHolder.Rows
}

// ExecuteChannelQueryV1 - @Kark TODO v1
// TODO error handling.
func (store *MemSQL) ExecuteChannelQueryV1(projectID int64, query *model.ChannelQueryV1,
	reqID string) (*model.ChannelQueryResultV1, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"query":      query,
		"req_id":     reqID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	logCtx := log.WithFields(logFields)
	queryResult := &model.ChannelQueryResultV1{}
	var columns []string
	var resultMetrics [][]interface{}
	status := http.StatusOK
	var err int
	sources, _ := store.GetCustomAdsSourcesByProject(projectID)
	if !(isValidChannel(query.Channel, sources)) {
		return queryResult, http.StatusBadRequest
	}

	switch query.Channel {
	case CAAllChannelAds:
		columns, resultMetrics, err = store.executeAllChannelsQueryV1(projectID, query, reqID)
	case CAChannelBingAds:
		columns, resultMetrics, err = store.ExecuteBingAdsChannelQueryV1(projectID, query, reqID)
	case CAChannelFacebookAds:
		columns, resultMetrics, err = store.ExecuteFacebookChannelQueryV1(projectID, query, reqID)
	case CAChannelGoogleAds:
		columns, resultMetrics, err = store.ExecuteAdwordsChannelQueryV1(projectID, query, reqID)
	case CAChannelLinkedinAds:
		columns, resultMetrics, err = store.ExecuteLinkedinChannelQueryV1(projectID, query, reqID)
	case CAChannelLinkedinCompanyEngagements:
		columns, resultMetrics, err = store.ExecuteLinkedinCompanyEngagementsChannelQueryV1(projectID, query, reqID)
	case CAChannelSearchConsole:
		columns, resultMetrics, err = store.ExecuteGoogleOrganicChannelQueryV1(projectID, query, reqID)
	}
	for _, source := range sources {
		if source == query.Channel {
			columns, resultMetrics, err = store.ExecuteCustomAdsChannelQueryV1(projectID, query, reqID)
		}
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
func (store *MemSQL) executeAllChannelsQueryV1(projectID int64, query *model.ChannelQueryV1,
	reqID string) ([]string, [][]interface{}, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"query":      query,
		"req_id":     reqID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	defer U.NotifyOnPanicWithError(C.GetConfig().Env, C.GetConfig().AppName)

	logCtx := log.WithFields(logFields)
	var finalQuery string
	var finalParams []interface{}
	var selectMetrics, columns []string
	isGroupByTimestamp := query.GetGroupByTimestamp() != ""

	ftMapping, err := store.GetActiveFiveTranMapping(projectID, model.BingAdsIntegration)
	customerAccountID := ""
	if err == nil {
		customerAccountID = ftMapping.Accounts
	}
	projectSetting, errCode := store.GetProjectSetting(projectID)
	customAdsInt := store.IsCustomAdsAvailable(projectID)
	customAdsAccounts := make([]string, 0)
	sources := make([]string, 0)
	if customAdsInt {
		sources, _ = store.GetCustomAdsSourcesByProject(projectID)
		customAdsAccounts, _ = store.GetCustomAdsAccountsByProject(projectID, sources)
	}
	if errCode != http.StatusFound {
		headers := model.GetHeadersFromQuery(*query)
		return headers, make([][]interface{}, 0, 0), http.StatusNotFound
	} else if (projectSetting.IntAdwordsCustomerAccountId == nil || *projectSetting.IntAdwordsCustomerAccountId == "") &&
		(projectSetting.IntFacebookAdAccount == "") && (projectSetting.IntLinkedinAdAccount == "") && customerAccountID == "" && len(customAdsAccounts) < 0 {
		log.Warn("Integration not present for channels.")
		headers := model.GetHeadersFromQuery(*query)
		return headers, make([][]interface{}, 0, 0), http.StatusOK
	}
	if (query.GroupBy == nil || len(query.GroupBy) == 0) && (query.GroupByTimestamp == nil || len(query.GroupByTimestamp.(string)) == 0) {
		adwordsSQL, adwordsParams, commonKeys, commonMetrics, facebookSQL, facebookParams, linkedinSQL, linkedinParams, bingAdsSQL, bingAdsParams, customAdsSQL, customAdsParams, err := store.getIndividualChannelsSQLAndParametersV1(projectID, query, reqID, false, sources)
		if err == http.StatusNotFound {
			headers := model.GetHeadersFromQuery(*query)
			return headers, make([][]interface{}, 0, 0), http.StatusOK
		}
		if err != http.StatusOK {
			headers := model.GetHeadersFromQuery(*query)
			return headers, make([][]interface{}, 0, 0), err
		}
		finalSQLs := U.AppendNonNullValues(adwordsSQL, facebookSQL, linkedinSQL, bingAdsSQL)
		for _, adsSql := range customAdsSQL {
			finalSQLs = append(finalSQLs, adsSql)
		}
		finalParams = append(adwordsParams, facebookParams...)
		finalParams = append(finalParams, linkedinParams...)
		finalParams = append(finalParams, bingAdsParams...)
		for _, customAdsParam := range customAdsParams {
			finalParams = append(finalParams, customAdsParam...)
		}

		for _, metric := range commonMetrics {
			value := fmt.Sprintf("%s(%s) as %s", channelMetricsToOperation[metric], metric, metric)
			selectMetrics = append(selectMetrics, value)
		}
		finalQuery = fmt.Sprintf(CAUnionQuery1, joinWithComma(selectMetrics...), joinWithWordInBetween("UNION", finalSQLs...),
			getOrderByClause(isGroupByTimestamp, commonMetrics), channeAnalyticsLimit)
		columns = append(commonKeys, commonMetrics...)
	} else if (query.GroupBy == nil || len(query.GroupBy) == 0) && (!(query.GroupByTimestamp == nil || len(query.GroupByTimestamp.(string)) == 0)) {
		adwordsSQL, adwordsParams, commonKeys, commonMetrics, facebookSQL, facebookParams, linkedinSQL, linkedinParams, bingAdsSQL, bingAdsParams, customAdsSQL, customAdsParams, err := store.getIndividualChannelsSQLAndParametersV1(projectID, query, reqID, false, sources)
		if err == http.StatusNotFound {
			headers := model.GetHeadersFromQuery(*query)
			return headers, make([][]interface{}, 0, 0), http.StatusOK
		}
		if err != http.StatusOK {
			headers := model.GetHeadersFromQuery(*query)
			return headers, make([][]interface{}, 0, 0), err
		}
		finalSQLs := U.AppendNonNullValues(adwordsSQL, facebookSQL, linkedinSQL, bingAdsSQL)
		for _, adsSql := range customAdsSQL {
			finalSQLs = append(finalSQLs, adsSql)
		}
		finalParams = append(adwordsParams, facebookParams...)
		finalParams = append(finalParams, linkedinParams...)
		finalParams = append(finalParams, bingAdsParams...)
		for _, customAdsParam := range customAdsParams {
			finalParams = append(finalParams, customAdsParam...)
		}

		selectMetrics = append(selectMetrics, model.AliasDateTime)
		for _, metric := range commonMetrics {
			value := fmt.Sprintf("%s(%s) as %s", channelMetricsToOperation[metric], metric, metric)
			selectMetrics = append(selectMetrics, value)
		}
		finalQuery = fmt.Sprintf(CAUnionQuery2, joinWithComma(selectMetrics...), joinWithWordInBetween("UNION", finalSQLs...),
			model.AliasDateTime, getOrderByClause(isGroupByTimestamp, commonMetrics), channeAnalyticsLimit)
		columns = append(commonKeys, commonMetrics...)
	} else {
		adwordsSQL, adwordsParams, commonKeys, commonMetrics, facebookSQL, facebookParams, linkedinSQL, linkedinParams, bingAdsSQL, bingAdsParams, customAdsSQL, customAdsParams, err := store.getIndividualChannelsSQLAndParametersV1(projectID, query, reqID, true, sources)
		if err == http.StatusNotFound {
			headers := model.GetHeadersFromQuery(*query)
			return headers, make([][]interface{}, 0, 0), http.StatusOK
		}
		if err != http.StatusOK {
			headers := model.GetHeadersFromQuery(*query)
			return headers, make([][]interface{}, 0, 0), err
		}
		finalSQLs := U.AppendNonNullValues(adwordsSQL, facebookSQL, linkedinSQL, bingAdsSQL)
		for _, adsSql := range customAdsSQL {
			finalSQLs = append(finalSQLs, adsSql)
		}
		finalParams = append(adwordsParams, facebookParams...)
		finalParams = append(finalParams, linkedinParams...)
		finalParams = append(finalParams, bingAdsParams...)
		for _, customAdsParam := range customAdsParams {
			finalParams = append(finalParams, customAdsParam...)
		}

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
		headers := model.GetHeadersFromQuery(*query)
		return headers, make([][]interface{}, 0, 0), http.StatusInternalServerError
	}
	return columns, resultMetrics, http.StatusOK
}

func (store *MemSQL) getIndividualChannelsSQLAndParametersV1(projectID int64, query *model.ChannelQueryV1, reqID string, fetchSource bool, customSources []string) (string, []interface{}, []string, []string, string, []interface{}, string, []interface{}, string, []interface{}, []string, [][]interface{}, int) {
	logFields := log.Fields{
		"project_id":   projectID,
		"query":        query,
		"req_id":       reqID,
		"fetch_source": fetchSource,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	isGroupBytimestamp := query.GetGroupByTimestamp() != ""
	genericFilters, channelBreakdownFilters := model.GetDecoupledFiltersForChannelBreakdownFilters(query.Filters)
	query.Filters = genericFilters
	isAdwordsReq, isFacebookReq, isLinkedinReq, isBingAdsReq, isCustomAdsReq, errCode := model.GetRequiredChannels(channelBreakdownFilters, customSources)
	projectSetting, _ := store.GetProjectSetting(projectID)
	bingAdsInt := store.IsBingIntegrationAvailable(projectID)
	customAdsInt := store.IsCustomAdsAvailable(projectID)
	isAdwordsReq = isAdwordsReq && (projectSetting.IntAdwordsCustomerAccountId != nil && *projectSetting.IntAdwordsCustomerAccountId != "")
	isFacebookReq = isFacebookReq && (projectSetting.IntFacebookAdAccount != "")
	isLinkedinReq = (isLinkedinReq && projectSetting.IntLinkedinAdAccount != "")
	isBingAdsReq = isBingAdsReq && (bingAdsInt == true)
	isCustomAdsReq = isCustomAdsReq && (customAdsInt == true)
	if errCode != http.StatusOK {
		return "", []interface{}{}, make([]string, 0, 0), make([]string, 0, 0), "", []interface{}{}, "", []interface{}{}, "", []interface{}{}, make([]string, 0), make([][]interface{}, 0), errCode
	}

	finAdwordsSQL, finAdwordsParams, finFacebookSQL, finFacebookParams, finLinkedinSQL, finLinkedinParams, finBingAdsSQL, finBingAdsParams, finCustomAdsSQL, finCustomAdsParams := "", make([]interface{}, 0), "", make([]interface{}, 0), "", make([]interface{}, 0), "", make([]interface{}, 0), make([]string, 0), make([][]interface{}, 0)

	finalKeys := make([]string, 0, 0)
	finalMetrics := make([]string, 0, 0)
	if isAdwordsReq {
		adwordsSQL, adwordsParams, adwordsSelectKeys, adwordsMetrics, adwordsErr := store.GetSQLQueryAndParametersForAdwordsQueryV1(projectID, query, reqID, fetchSource, " LIMIT 10000", isGroupBytimestamp, nil)
		if adwordsErr != http.StatusOK && adwordsErr != http.StatusNotFound {
			return "", []interface{}{}, make([]string, 0, 0), make([]string, 0, 0), "", []interface{}{}, "", []interface{}{}, "", []interface{}{}, make([]string, 0), make([][]interface{}, 0), adwordsErr
		}
		if len(adwordsSQL) > 0 {
			finalKeys = adwordsSelectKeys
			finalMetrics = adwordsMetrics
			adwordsSQL = fmt.Sprintf("( %s )", adwordsSQL[:len(adwordsSQL)-2])
		}
		finAdwordsSQL, finAdwordsParams = adwordsSQL, adwordsParams
	}
	if isFacebookReq {
		facebookSQL, facebookParams, facebookSelectKeys, facebookMetrics, facebookErr := store.GetSQLQueryAndParametersForFacebookQueryV1(projectID, query, reqID, fetchSource, " LIMIT 10000", isGroupBytimestamp, nil)
		if facebookErr != http.StatusOK && facebookErr != http.StatusNotFound {
			return "", []interface{}{}, make([]string, 0, 0), make([]string, 0, 0), "", []interface{}{}, "", []interface{}{}, "", []interface{}{}, make([]string, 0), make([][]interface{}, 0), facebookErr
		}
		if len(facebookSQL) > 0 {
			finalKeys = facebookSelectKeys
			finalMetrics = facebookMetrics
			facebookSQL = fmt.Sprintf("( %s )", facebookSQL[:len(facebookSQL)-2])
		}
		finFacebookSQL, finFacebookParams = facebookSQL, facebookParams
	}
	if isLinkedinReq {
		linkedinSQL, linkedinParams, linkedinSelectKeys, linkedinMetrics, linkedinErr := store.GetSQLQueryAndParametersForLinkedinQueryV1(projectID, query, reqID, fetchSource, " LIMIT 10000", isGroupBytimestamp, nil)
		if linkedinErr != http.StatusOK && linkedinErr != http.StatusNotFound {
			return "", []interface{}{}, make([]string, 0, 0), make([]string, 0, 0), "", []interface{}{}, "", []interface{}{}, "", []interface{}{}, make([]string, 0), make([][]interface{}, 0), linkedinErr
		}
		if len(linkedinSQL) > 0 {
			finalKeys = linkedinSelectKeys
			finalMetrics = linkedinMetrics
			linkedinSQL = fmt.Sprintf("( %s )", linkedinSQL[:len(linkedinSQL)-2])
		}
		finLinkedinSQL, finLinkedinParams = linkedinSQL, linkedinParams
	}
	if isBingAdsReq {
		bingAdsSQL, bingAdsParams, bingAdsSelectKeys, bingAdsMetrics, bingAdsErr := store.GetSQLQueryAndParametersForBingAdsQueryV1(projectID, query, reqID, fetchSource, " LIMIT 10000", isGroupBytimestamp, nil)
		if bingAdsErr != http.StatusOK && bingAdsErr != http.StatusNotFound {
			return "", []interface{}{}, make([]string, 0, 0), make([]string, 0, 0), "", []interface{}{}, "", []interface{}{}, "", []interface{}{}, make([]string, 0), make([][]interface{}, 0), bingAdsErr
		}
		if len(bingAdsSQL) > 0 {
			finalKeys = bingAdsSelectKeys
			finalMetrics = bingAdsMetrics
			bingAdsSQL = fmt.Sprintf("( %s )", bingAdsSQL[:len(bingAdsSQL)-2])
		}
		finBingAdsSQL, finBingAdsParams = bingAdsSQL, bingAdsParams

	}
	if isCustomAdsReq {
		sources, _ := store.GetCustomAdsSourcesByProject(projectID)
		for _, source := range sources {
			var queryCopy *model.ChannelQueryV1
			U.DeepCopy(query, &queryCopy)
			queryCopy.Channel = source
			customAdsSQL, customAdsParams, customAdsSelectKeys, customAdsMetrics, customAdsErr := store.GetSQLQueryAndParametersForCustomAdsQueryV1(projectID, queryCopy, reqID, fetchSource, " LIMIT 10000", isGroupBytimestamp, nil)
			if customAdsErr != http.StatusOK && customAdsErr != http.StatusNotFound {
				return "", []interface{}{}, make([]string, 0, 0), make([]string, 0, 0), "", []interface{}{}, "", []interface{}{}, "", []interface{}{}, make([]string, 0), make([][]interface{}, 0), customAdsErr
			}
			if len(customAdsSQL) > 0 {
				finalKeys = customAdsSelectKeys
				finalMetrics = customAdsMetrics
				customAdsSQL = fmt.Sprintf("( %s )", customAdsSQL[:len(customAdsSQL)-2])
			}
			finCustomAdsSQL = append(finCustomAdsSQL, customAdsSQL)
			finCustomAdsParams = append(finCustomAdsParams, customAdsParams)
		}

	}
	if !isAdwordsReq && !isFacebookReq && !isLinkedinReq && !isBingAdsReq && !isCustomAdsReq {
		return "", []interface{}{}, make([]string, 0, 0), make([]string, 0, 0), "", []interface{}{}, "", []interface{}{}, "", []interface{}{}, make([]string, 0), make([][]interface{}, 0), http.StatusNotFound
	}
	return finAdwordsSQL, finAdwordsParams, finalKeys, finalMetrics, finFacebookSQL, finFacebookParams, finLinkedinSQL, finLinkedinParams, finBingAdsSQL, finBingAdsParams, finCustomAdsSQL, finCustomAdsParams, http.StatusOK
}

// Common Methods for facebook and adwords starts here.

// Convert2DArrayTo1DArray ...
// @Kark TODO v1
func Convert2DArrayTo1DArray(inputArray [][]interface{}) []interface{} {
	logFields := log.Fields{
		"input_array": inputArray,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	result := make([]interface{}, 0, 0)
	for _, row := range inputArray {
		result = append(result, row...)
	}
	return result
}

// format yyyymmdd
func ChangeUnixTimestampToDate(timestamp int64) int64 {
	logFields := log.Fields{
		"timestamp": timestamp,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	time := time.Unix(timestamp, 0)
	date, _ := strconv.ParseInt(time.Format("20060102"), 10, 64)
	return date
}

// @Kark TODO v1
func hasAllIDsOnlyInGroupBy(query *model.ChannelQueryV1) bool {
	logFields := log.Fields{
		"query": query,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	for _, groupBy := range query.GroupBy {
		if !(strings.Contains(groupBy.Property, "id") || strings.Contains(groupBy.Property, "ID")) {
			return false
		}
	}
	return true
}

// @Kark TODO v1
func appendSelectTimestampIfRequiredForChannels(stmnt string, groupByTimestamp string, timezone string) string {
	logFields := log.Fields{
		"stmnt":              stmnt,
		"group_by_timestamp": groupByTimestamp,
		"timezone":           timezone,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	if groupByTimestamp == "" {
		return stmnt
	}

	return joinWithComma(stmnt, fmt.Sprintf("%s as %s",
		getSelectTimestampByTypeForChannels(groupByTimestamp, timezone), model.AliasDateTime))
}

// @Kark TODO v1
func getSelectTimestampByTypeForChannels(timestampType, timezone string) string {
	logFields := log.Fields{
		"timestamp_type": timestampType,
		"timezone":       timezone,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	var selectTz string
	var selectStr string
	selectTz = model.DefaultTimezone

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
	logFields := log.Fields{
		"is_group_by_timestamp": isGroupByTimestamp,
		"select_metrics":        selectMetrics,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
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
	logFields := log.Fields{
		"is_group_by_timestamp": isGroupByTimestamp,
		"select_metrics":        selectMetrics,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
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
func (store *MemSQL) ExecuteSQL(sqlStatement string, params []interface{}, logCtx *log.Entry) ([]string, [][]interface{}, error) {
	logFields := log.Fields{
		"sql_statement": sqlStatement,
		"params":        params,
		"log_ctx":       logCtx,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	rows, tx, err, reqID := store.ExecQueryWithContext(sqlStatement, params)
	if err != nil {
		logCtx.WithError(err).WithField("query", sqlStatement).WithField("params", params).Error("SQL Query failed.")
		return nil, nil, err
	}

	columns, resultRows, err := U.DBReadRows(rows, tx, reqID)
	if err != nil {
		return nil, nil, err
	}
	if len(resultRows) == 0 {
		logCtx.Warn("Aggregate query returned zero rows: ", sqlStatement, params)
		return nil, make([][]interface{}, 0, 0), nil
	}
	return columns, resultRows, nil
}

func (store *MemSQL) GetSmartPropertyAndRelated(projectID int64, object string, source string) map[string]model.PropertiesAndRelated {
	logFields := log.Fields{
		"project_id": projectID,
		"object":     object,
		"source":     source,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	db := C.GetServices().Db
	var smartPropertyRules []model.SmartPropertyRules
	object_type, isPresent := model.SmartPropertyRulesTypeAliasToType[object]
	if !isPresent {
		return nil
	}
	err := db.Table("smart_property_rules").Where("project_id = ? AND type = ? and is_deleted = ?", projectID, object_type, false).Find(&smartPropertyRules).Error
	if err != nil {
		log.WithError(err).Error("Failed to get smart property filters from DB")
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
