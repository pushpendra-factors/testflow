package memsql

import (
	"database/sql"
	"errors"
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	bingadsCampaign                                 = "campaigns"
	bingadsAdGroup                                  = "ad_groups"
	bingadsKeyword                                  = "keyword"
	fromIntegrationDocuments                        = " FROM integration_documents "
	staticWhereStatementForBingAds                  = "WHERE project_id = ? AND source= ? AND document_type = ? AND customer_account_id IN (?) AND timestamp between ? AND ? "
	staticWhereStatementForBingAdsWithSmartProperty = "WHERE integration_documents.project_id = ? AND integration_documents.source= ?  AND integration_documents.customer_account_id IN ( ? ) AND integration_documents.document_type = ? AND integration_documents.timestamp between ? AND ? "
	bingadsFilterQueryStr                           = "SELECT DISTINCT(LCASE(JSON_EXTRACT_STRING(value, ?))) as filter_value FROM integration_documents WHERE project_id = ? AND customer_account_id IN (?) AND document_type = ? AND JSON_EXTRACT_STRING(value, ?) IS NOT NULL AND timestamp BETWEEN ? AND ? LIMIT 5000"
)

// add other properties and vals
var objectAndPropertyToValueInBingAdsReportsMapping = map[string]string{
	"campaigns.id":                "JSON_EXTRACT_STRING(value, 'campaign_id')",
	"campaigns.status":            "JSON_EXTRACT_STRING(value, 'campaign_status')",
	"campaigns.name":              "JSON_EXTRACT_STRING(value, 'campaign_name')",
	"campaigns.type":              "JSON_EXTRACT_STRING(value, 'campaign_type')",
	"ad_groups.id":                "JSON_EXTRACT_STRING(value, 'ad_group_id')",
	"ad_groups.status":            "JSON_EXTRACT_STRING(value, 'ad_group_status')",
	"ad_groups.name":              "JSON_EXTRACT_STRING(value, 'ad_group_name')",
	"ad_groups.bid_strategy_type": "JSON_EXTRACT_STRING(value, 'ad_group_bid_strategy_type')",
	"keyword.id":                  "JSON_EXTRACT_STRING(value, 'keyword_id')",
	"keyword.name":                "JSON_EXTRACT_STRING(value, 'keyword_name')",
	"keyword.status":              "JSON_EXTRACT_STRING(value, 'keyword_status')",
	"keyword.match_type":          "JSON_EXTRACT_STRING(value, 'keyword_match_type')",
}
var BingAdsMetricsToAggregatesInReportsMapping = map[string]string{
	"impressions": "SUM(JSON_EXTRACT_STRING(value, 'impressions'))",
	"clicks":      "SUM(JSON_EXTRACT_STRING(value, 'clicks'))",
	"spend":       "SUM(JSON_EXTRACT_STRING(value, 'spend'))",
	"conversions": "SUM(JSON_EXTRACT_STRING(value, 'conversions'))",
}

const bingadsAdGroupMetadataFetchQueryStr = "WITH ad_group as (select ad_group_information.campaign_id_1 as campaign_id, ad_group_information.ad_group_id_1 as ad_group_id, ad_group_information.ad_group_name_1 as ad_group_name " +
	"from ( " +
	"select JSON_EXTRACT_STRING(value, 'campaign_id') as campaign_id_1, document_id as ad_group_id_1, JSON_EXTRACT_STRING(value, 'name') as ad_group_name_1, timestamp " +
	"from integration_documents where document_type = ? AND project_id = ? AND source= ? AND timestamp between ? AND ? AND customer_account_id IN (?) " +
	") as ad_group_information " +
	"INNER JOIN " +
	"(select document_id as ad_group_id_1, max(timestamp) as timestamp " +
	"from integration_documents where document_type = ? AND project_id = ? AND source= ? AND timestamp between ? AND ? AND customer_account_id IN (?) group by ad_group_id_1 " +
	") as ad_group_latest_timestamp_id " +
	"ON ad_group_information.ad_group_id_1 = ad_group_latest_timestamp_id.ad_group_id_1 AND ad_group_information.timestamp = ad_group_latest_timestamp_id.timestamp), " +

	" campaign as (select campaign_information.campaign_id_1 as campaign_id, campaign_information.campaign_name_1 as campaign_name " +
	"from ( " +
	"select document_id as campaign_id_1, JSON_EXTRACT_STRING(value, 'name') as campaign_name_1, timestamp " +
	"from integration_documents where document_type = ? AND project_id = ? AND source= ? AND timestamp between ? AND ? AND customer_account_id IN (?) " +
	") as campaign_information " +
	"INNER JOIN " +
	"(select document_id as campaign_id_1, max(timestamp) as timestamp " +
	"from integration_documents where document_type = ? AND project_id = ? AND source= ? AND timestamp between ? AND ? AND customer_account_id IN (?) group by campaign_id_1 " +
	") as campaign_latest_timestamp_id " +
	"ON campaign_information.campaign_id_1 = campaign_latest_timestamp_id.campaign_id_1 AND campaign_information.timestamp = campaign_latest_timestamp_id.timestamp) " +

	"select campaign.campaign_id as campaign_id, campaign.campaign_name as campaign_name, ad_group.ad_group_id as ad_group_id, ad_group.ad_group_name as ad_group_name " +
	"from campaign join ad_group on ad_group.campaign_id = campaign.campaign_id"

const bingadsCampaignMetadataFetchQueryStr = "select campaign_information.campaign_id as campaign_id, campaign_information.campaign_name as campaign_name " +
	"from ( " +
	"select document_id AS campaign_id, JSON_EXTRACT_STRING(value, 'name') as campaign_name, timestamp " +
	"from integration_documents where document_type = ? AND project_id = ? AND source= ? AND timestamp between ? AND ? AND customer_account_id IN (?) " +
	") as campaign_information " +
	"INNER JOIN " +
	"(select document_id AS campaign_id, max(timestamp) as timestamp " +
	"from integration_documents where document_type = ? AND project_id = ? AND source= ? AND timestamp between ? AND ? AND customer_account_id IN (?) group by campaign_id " +
	") as campaign_latest_timestamp_id " +
	"ON campaign_information.campaign_id = campaign_latest_timestamp_id.campaign_id AND campaign_information.timestamp = campaign_latest_timestamp_id.timestamp "

func (store *MemSQL) GetBingadsFilterValues(projectID int64, requestFilterObject string, requestFilterProperty string, reqID string) ([]interface{}, int) {
	logFields := log.Fields{
		"project_id":              projectID,
		"request_filter_object":   requestFilterObject,
		"request_filter_property": requestFilterProperty,
		"req_id":                  reqID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)
	logCtx = log.WithField("project_id", projectID).WithField("req_id", reqID)
	ftMapping, err := store.GetActiveFiveTranMapping(projectID, model.BingAdsIntegration)
	if err != nil {
		log.WithError(err).Error("Failed to fetch connector id from db")
		return nil, http.StatusInternalServerError
	}
	customerAccountIDs := strings.Split(ftMapping.Accounts, ",")
	_, isPresent := model.SmartPropertyReservedNames[requestFilterProperty]
	if !isPresent {
		filterValues, errCode := store.getSmartPropertyFilterValues(projectID, requestFilterObject, requestFilterProperty, "bingads", reqID)
		if errCode != http.StatusFound {
			return []interface{}{}, http.StatusInternalServerError
		}
		return filterValues, http.StatusFound
	}
	requestFilterProperty = strings.TrimPrefix(requestFilterProperty, fmt.Sprintf("%v_", requestFilterObject))
	docType := model.BingadsDocumentTypeAlias[model.BingAdsObjectInternalRepresentationToExternalRepresentation[requestFilterObject]]
	filterProperty := model.BingAdsInternalRepresentationToExternalRepresentation[fmt.Sprintf("%v.%v", model.BingAdsObjectInternalRepresentationToExternalRepresentation[requestFilterObject], requestFilterProperty)]

	from, to := model.GetFromAndToDatesForFilterValues()
	params := []interface{}{filterProperty, projectID, customerAccountIDs, docType, filterProperty, from, to}
	_, resultRows, err := store.ExecuteSQL(bingadsFilterQueryStr, params, logCtx)
	if err != nil {
		logCtx.WithError(err).WithField("query", bingadsFilterQueryStr).WithField("params", params).Error(model.BingSpecificError)
		return make([]interface{}, 0), http.StatusInternalServerError
	}
	return Convert2DArrayTo1DArray(resultRows), http.StatusFound
}

func (store *MemSQL) GetBingadsFilterValuesSQLAndParams(projectID int64, requestFilterObject string, requestFilterProperty string, reqID string) (string, []interface{}, int) {
	logFields := log.Fields{
		"project_id":              projectID,
		"request_filter_object":   requestFilterObject,
		"request_filter_property": requestFilterProperty,
		"req_id":                  reqID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)
	logCtx.WithField("project_id", projectID).WithField("req_id", reqID)
	ftMapping, err := store.GetActiveFiveTranMapping(projectID, model.BingAdsIntegration)
	if err != nil {
		logCtx.WithError(err).Error("Failed to fetch connector id from db")
		return "", nil, http.StatusNotFound
	}
	customerAccountIDs := strings.Split(ftMapping.Accounts, ",")
	requestFilterProperty = strings.TrimPrefix(requestFilterProperty, fmt.Sprintf("%v_", requestFilterObject))
	docType := model.BingadsDocumentTypeAlias[model.BingAdsObjectInternalRepresentationToExternalRepresentation[requestFilterObject]]
	from, to := model.GetFromAndToDatesForFilterValues()
	filterProperty := model.BingAdsInternalRepresentationToExternalRepresentation[fmt.Sprintf("%v.%v", model.BingAdsObjectInternalRepresentationToExternalRepresentation[requestFilterObject], requestFilterProperty)]
	params := []interface{}{filterProperty, projectID, customerAccountIDs, docType, filterProperty, from, to}
	return bingadsFilterQueryStr, params, http.StatusFound
}

func (store *MemSQL) buildBingAdsChannelConfig(projectID int64) *model.ChannelConfigResult {
	bingAdsObjectsAndProperties := store.buildObjectAndPropertiesForBingAds(projectID, model.ObjectsForBingads)
	// selectable metrics for bing ads to be added ?
	selectMetrics := append(SelectableMetricsForAllChannels)
	objectsAndProperties := bingAdsObjectsAndProperties

	return &model.ChannelConfigResult{
		SelectMetrics:        selectMetrics,
		ObjectsAndProperties: objectsAndProperties,
	}
}
func (store *MemSQL) buildObjectAndPropertiesForBingAds(projectID int64, objects []string) []model.ChannelObjectAndProperties {
	objectsAndProperties := make([]model.ChannelObjectAndProperties, 0)
	for _, currentObject := range objects {
		// to do: check if normal properties present then only smart properties will be there
		propertiesAndRelated, isPresent := model.MapOfBingAdsObjectsToPropertiesAndRelated[currentObject]
		var currentProperties []model.ChannelProperty
		var currentPropertiesSmart []model.ChannelProperty
		if isPresent {
			currentProperties = buildProperties(propertiesAndRelated)
			smartProperty := store.GetSmartPropertyAndRelated(projectID, currentObject, "bingads")
			currentPropertiesSmart = buildProperties(smartProperty)
			currentProperties = append(currentProperties, currentPropertiesSmart...)
		} else {
			currentProperties = buildProperties(allChannelsPropertyToRelated)
			smartProperty := store.GetSmartPropertyAndRelated(projectID, currentObject, "bingads")
			currentPropertiesSmart = buildProperties(smartProperty)
			currentProperties = append(currentProperties, currentPropertiesSmart...)
		}
		objectsAndProperties = append(objectsAndProperties, buildObjectsAndProperties(currentProperties, []string{currentObject})...)
	}
	return objectsAndProperties
}

// modify
func (store *MemSQL) ExecuteBingAdsChannelQueryV1(projectID int64, query *model.ChannelQueryV1, reqID string) ([]string, [][]interface{}, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"query":      query,
		"req_id":     reqID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	fetchSource := false
	defer U.NotifyOnPanicWithError(C.GetConfig().Env, C.GetConfig().AppName)
	logCtx := log.WithField("xreq_id", reqID)

	if query.GroupByTimestamp == "" {
		sql, params, selectKeys, selectMetrics, errCode := store.GetSQLQueryAndParametersForBingAdsQueryV1(projectID,
			query, reqID, fetchSource, " LIMIT 10000", false, nil)
		if errCode == http.StatusNotFound {
			headers := model.GetHeadersFromQuery(*query)
			return headers, make([][]interface{}, 0, 0), http.StatusOK
		}
		if errCode != http.StatusOK {
			headers := model.GetHeadersFromQuery(*query)
			return headers, make([][]interface{}, 0, 0), errCode
		}
		_, resultMetrics, err := store.ExecuteSQL(sql, params, logCtx)
		columns := append(selectKeys, selectMetrics...)
		if err != nil {
			logCtx.WithError(err).WithField("query", sql).WithField("params", params).Error("failed in bingads with error")
			return columns, make([][]interface{}, 0, 0), http.StatusInternalServerError
		}
		return columns, resultMetrics, http.StatusOK
	} else {
		sql, params, selectKeys, selectMetrics, errCode := store.GetSQLQueryAndParametersForBingAdsQueryV1(
			projectID, query, reqID, fetchSource, " LIMIT 100", false, nil)
		if errCode == http.StatusNotFound {
			headers := model.GetHeadersFromQuery(*query)
			return headers, make([][]interface{}, 0, 0), http.StatusOK
		}
		if errCode != http.StatusOK {
			headers := model.GetHeadersFromQuery(*query)
			return headers, make([][]interface{}, 0, 0), errCode
		}
		_, resultMetrics, err := store.ExecuteSQL(sql, params, logCtx)
		columns := append(selectKeys, selectMetrics...)
		if err != nil {
			logCtx.WithError(err).WithField("query", sql).WithField("params", params).Error("failed in bingads with error")
			return columns, make([][]interface{}, 0, 0), http.StatusInternalServerError
		}
		groupByCombinations := model.GetGroupByCombinationsForChannelAnalytics(columns, resultMetrics)
		sql, params, selectKeys, selectMetrics, errCode = store.GetSQLQueryAndParametersForBingAdsQueryV1(
			projectID, query, reqID, fetchSource, " LIMIT 10000", true, groupByCombinations)
		if errCode != http.StatusOK {
			headers := model.GetHeadersFromQuery(*query)
			return headers, make([][]interface{}, 0, 0), errCode
		}
		_, resultMetrics, err = store.ExecuteSQL(sql, params, logCtx)
		columns = append(selectKeys, selectMetrics...)
		if err != nil {
			logCtx.WithError(err).WithField("query", sql).WithField("params", params).Error("Failed in bingads with the error.")
			return columns, make([][]interface{}, 0, 0), http.StatusInternalServerError
		}
		return columns, resultMetrics, http.StatusOK
	}
}

// modify acc to sample data
func (store *MemSQL) GetSQLQueryAndParametersForBingAdsQueryV1(projectID int64, query *model.ChannelQueryV1, reqID string, fetchSource bool,
	limitString string, isGroupByTimestamp bool, groupByCombinationsForGBT []map[string]interface{}) (string, []interface{}, []string, []string, int) {
	logFields := log.Fields{
		"project_id":                    projectID,
		"query":                         query,
		"fetch_source":                  fetchSource,
		"req_id":                        reqID,
		"limit_string":                  limitString,
		"is_group_by_timestamp":         isGroupByTimestamp,
		"group_by_combinations_for_gbt": groupByCombinationsForGBT,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	var selectMetrics []string
	var sql string
	var selectKeys []string
	var params []interface{}
	logCtx := log.WithFields(logFields)
	transformedQuery, customerAccountID, err := store.transFormRequestFieldsAndFetchRequiredFieldsForBingads(
		projectID, *query, reqID)
	if err != nil && err.Error() == "record not found" {
		logCtx.WithError(err).Info(model.BingSpecificError)
		return "", nil, make([]string, 0, 0), make([]string, 0, 0), http.StatusNotFound
	}
	if err != nil {
		logCtx.WithError(err).Error(model.BingSpecificError)
		return "", make([]interface{}, 0, 0), make([]string, 0, 0), make([]string, 0, 0), http.StatusBadRequest
	}
	// smart properties check
	isSmartPropertyPresent := checkSmartProperty(query.Filters, query.GroupBy)
	if isSmartPropertyPresent {
		sql, params, selectKeys, selectMetrics, err = buildBingAdsQueryWithSmartProperty(transformedQuery, projectID, customerAccountID, fetchSource, limitString, isGroupByTimestamp, groupByCombinationsForGBT)
		if err != nil {
			return "", make([]interface{}, 0, 0), make([]string, 0, 0), make([]string, 0, 0), http.StatusInternalServerError
		}
		return sql, params, selectKeys, selectMetrics, http.StatusOK
	}
	sql, params, selectKeys, selectMetrics, err = buildBingAdsQueryV1(transformedQuery, projectID, customerAccountID, fetchSource, limitString, isGroupByTimestamp, groupByCombinationsForGBT)
	if err != nil {
		return "", make([]interface{}, 0, 0), make([]string, 0, 0), make([]string, 0, 0), http.StatusInternalServerError
	}
	return sql, params, selectKeys, selectMetrics, http.StatusOK
}

func (store *MemSQL) transFormRequestFieldsAndFetchRequiredFieldsForBingads(projectID int64,
	query model.ChannelQueryV1, reqID string) (*model.ChannelQueryV1, string, error) {
	logFields := log.Fields{
		"project_id": projectID,
		"query":      query,
		"req_id":     reqID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	var transformedQuery model.ChannelQueryV1
	var err error
	logCtx := log.WithFields(logFields)
	ftMapping, err := store.GetActiveFiveTranMapping(projectID, model.BingAdsIntegration)
	if err != nil {
		log.WithError(err).Error("Failed to fetch connector id from db")
		return &model.ChannelQueryV1{}, "", err
	}
	customerAccountID := ftMapping.Accounts
	transformedQuery, err = convertFromRequestToBingAdsSpecificRepresentation(query)
	if err != nil {
		logCtx.Warn("Request failed in validation: ", err)
		return &model.ChannelQueryV1{}, "", err
	}
	return &transformedQuery, customerAccountID, nil
}

func convertFromRequestToBingAdsSpecificRepresentation(query model.ChannelQueryV1) (model.ChannelQueryV1, error) {
	logFields := log.Fields{
		"query": query,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	var transformedQuery model.ChannelQueryV1
	var err1, err2, err3 error
	transformedQuery.SelectMetrics, err1 = getBingAdsSpecificMetrics(query.SelectMetrics)
	transformedQuery.Filters, err2 = getBingAdsSpecificFilters(query.Filters)
	transformedQuery.GroupBy, err3 = getBingAdsSpecificGroupBy(query.GroupBy)
	if err1 != nil {
		return query, err1
	}
	if err2 != nil {
		return query, err2
	}
	if err3 != nil {
		return query, err3
	}
	transformedQuery.From = U.GetDateAsStringIn(query.From, U.TimeZoneString(query.Timezone))
	transformedQuery.To = U.GetDateAsStringIn(query.To, U.TimeZoneString(query.Timezone))
	transformedQuery.Timezone = query.Timezone
	transformedQuery.GroupByTimestamp = query.GroupByTimestamp

	return transformedQuery, nil
}

func getBingAdsSpecificMetrics(requestSelectMetrics []string) ([]string, error) {
	logFields := log.Fields{
		"request_select_metrics": requestSelectMetrics,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	resultMetrics := make([]string, 0, 0)
	for _, requestMetric := range requestSelectMetrics {
		metric, isPresent := model.BingAdsInternalRepresentationToExternalRepresentationForReports[requestMetric]
		if !isPresent {
			return make([]string, 0, 0), errors.New("Invalid metric found for document type")
		}
		resultMetrics = append(resultMetrics, metric)
	}
	return resultMetrics, nil
}

func getBingAdsSpecificFilters(requestFilters []model.ChannelFilterV1) ([]model.ChannelFilterV1, error) {
	logFields := log.Fields{
		"request_filters": requestFilters,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	resultFilters := make([]model.ChannelFilterV1, 0, 0)
	for _, requestFilter := range requestFilters {
		var resultFilter model.ChannelFilterV1
		filterObject, isPresent := model.BingAdsObjectInternalRepresentationToExternalRepresentation[requestFilter.Object]
		if !isPresent {
			return make([]model.ChannelFilterV1, 0, 0), errors.New("Invalid filter key found for document type")
		}
		resultFilter = requestFilter
		resultFilter.Object = filterObject
		resultFilters = append(resultFilters, resultFilter)
	}
	return resultFilters, nil
}

func getBingAdsSpecificGroupBy(requestGroupBys []model.ChannelGroupBy) ([]model.ChannelGroupBy, error) {

	resultGroupBys := make([]model.ChannelGroupBy, 0, 0)
	for _, requestGroupBy := range requestGroupBys {
		var resultGroupBy model.ChannelGroupBy
		groupByObject, isPresent := model.BingAdsObjectInternalRepresentationToExternalRepresentation[requestGroupBy.Object]
		if !isPresent {
			return make([]model.ChannelGroupBy, 0, 0), errors.New("Invalid groupby key found for document type")
		}
		resultGroupBy = requestGroupBy
		resultGroupBy.Object = groupByObject
		resultGroupBys = append(resultGroupBys, resultGroupBy)
	}
	return resultGroupBys, nil
}

func buildBingAdsQueryV1(query *model.ChannelQueryV1, projectID int64, customerAccountID string, fetchSource bool,
	limitString string, isGroupByTimestamp bool, groupByCombinationsForGBT []map[string]interface{}) (string, []interface{}, []string, []string, error) {
	logFields := log.Fields{
		"project_id":                    projectID,
		"query":                         query,
		"fetch_source":                  fetchSource,
		"limit_string":                  limitString,
		"is_group_by_timestamp":         isGroupByTimestamp,
		"group_by_combinations_for_gbt": groupByCombinationsForGBT,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	lowestHierarchyLevel := getLowestHierarchyLevelForBingAds(query)
	lowestHierarchyReportLevel := model.BingAdsObjectToPerfomanceReportRepresentation[lowestHierarchyLevel] // suffix tbd
	sql, params, selectKeys, selectMetrics := getSQLAndParamsFromBingAdsReports(query, projectID, query.From, query.To, model.BingadsDocumentTypeAlias[lowestHierarchyReportLevel],
		fetchSource, limitString, isGroupByTimestamp, groupByCombinationsForGBT, customerAccountID)
	return sql, params, selectKeys, selectMetrics, nil
}

func buildBingAdsQueryWithSmartProperty(query *model.ChannelQueryV1, projectID int64, customerAccountID string, fetchSource bool,
	limitString string, isGroupByTimestamp bool, groupByCombinationsForGBT []map[string]interface{}) (string, []interface{}, []string, []string, error) {
	logFields := log.Fields{
		"project_id":                    projectID,
		"query":                         query,
		"fetch_source":                  fetchSource,
		"limit_string":                  limitString,
		"is_group_by_timestamp":         isGroupByTimestamp,
		"group_by_combinations_for_gbt": groupByCombinationsForGBT,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	lowestHierarchyLevel := getLowestHierarchyLevelForBingAds(query)
	lowestHierarchyReportLevel := model.BingAdsObjectToPerfomanceReportRepresentation[lowestHierarchyLevel] // suffix tbd
	sql, params, selectKeys, selectMetrics := getSQLAndParamsFromBingAdsReportsWithSmartProperty(query, projectID, query.From, query.To, model.BingadsDocumentTypeAlias[lowestHierarchyReportLevel],
		fetchSource, limitString, isGroupByTimestamp, groupByCombinationsForGBT, customerAccountID)
	return sql, params, selectKeys, selectMetrics, nil
}

func getLowestHierarchyLevelForBingAds(query *model.ChannelQueryV1) string {
	logFields := log.Fields{
		"query": query,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	// Fetch the propertyNames
	var objectNames []string
	for _, filter := range query.Filters {
		objectNames = append(objectNames, filter.Object)
	}

	for _, groupBy := range query.GroupBy {
		objectNames = append(objectNames, groupBy.Object)
	}

	// Check if present
	for _, objectName := range objectNames {
		if objectName == bingadsKeyword {
			return bingadsKeyword
		}
	}

	for _, objectName := range objectNames {
		if objectName == bingadsAdGroup {
			return bingadsAdGroup
		}
	}

	for _, objectName := range objectNames {
		if objectName == bingadsCampaign {
			return bingadsCampaign
		}
	}
	return bingadsCampaign
}
func getSQLAndParamsFromBingAdsReports(query *model.ChannelQueryV1, projectID int64, from, to int64, docType int,
	fetchSource bool, limitString string, isGroupByTimestamp bool, groupByCombinationsForGBT []map[string]interface{}, customerAccountID string) (string, []interface{}, []string, []string) {
	logFields := log.Fields{
		"project_id":                    projectID,
		"query":                         query,
		"from":                          from,
		"to":                            to,
		"document_type":                 docType,
		"fetch_source":                  fetchSource,
		"limit_string":                  limitString,
		"is_group_by_timestamp":         isGroupByTimestamp,
		"group_by_combinations_for_gbt": groupByCombinationsForGBT,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	selectQuery := "SELECT "
	selectMetrics := make([]string, 0, 0)
	groupByStatement := ""
	groupByKeysWithoutTimestamp := make([]string, 0, 0)
	selectKeys := make([]string, 0, 0)
	finalSelectKeys := make([]string, 0, 0)
	responseSelectKeys := make([]string, 0, 0)
	responseSelectMetrics := make([]string, 0, 0)

	for _, groupBy := range query.GroupBy {
		key := groupBy.Object + "." + groupBy.Property
		groupByKeysWithoutTimestamp = append(groupByKeysWithoutTimestamp, model.BingAdsInternalRepresentationToExternalRepresentationForReports[key])
	}
	if isGroupByTimestamp {
		groupByStatement = joinWithComma(append(groupByKeysWithoutTimestamp, model.AliasDateTime)...)
	} else {
		groupByStatement = joinWithComma(groupByKeysWithoutTimestamp...)
	}
	// SelectKeys

	for _, groupBy := range query.GroupBy {
		key := groupBy.Object + "." + groupBy.Property
		if groupBy.Object == CAFilterChannel {
			value := fmt.Sprintf("'bingads' as %s", model.BingAdsInternalRepresentationToExternalRepresentationForReports[key])
			selectKeys = append(selectKeys, value)
			responseSelectKeys = append(responseSelectKeys, model.BingAdsInternalRepresentationToExternalRepresentationForReports[key])
		} else {
			value := fmt.Sprintf("%s as %s", objectAndPropertyToValueInBingAdsReportsMapping[key], model.BingAdsInternalRepresentationToExternalRepresentationForReports[key])
			selectKeys = append(selectKeys, value)
			responseSelectKeys = append(responseSelectKeys, model.BingAdsInternalRepresentationToExternalRepresentationForReports[key])
		}
	}

	finalSelectKeys = append(finalSelectKeys, selectKeys...)
	if isGroupByTimestamp {
		finalSelectKeys = append(finalSelectKeys, fmt.Sprintf("%s as %s",
			getSelectTimestampByTypeForChannels(query.GetGroupByTimestamp(), query.Timezone), model.AliasDateTime))
		responseSelectKeys = append(responseSelectKeys, model.AliasDateTime)
	}

	for _, selectMetric := range query.SelectMetrics {
		value := fmt.Sprintf("%s as %s", BingAdsMetricsToAggregatesInReportsMapping[selectMetric], model.BingAdsInternalRepresentationToExternalRepresentationForReports[selectMetric])
		selectMetrics = append(selectMetrics, value)

		value = model.BingAdsInternalRepresentationToExternalRepresentationForReports[selectMetric]
		responseSelectMetrics = append(responseSelectMetrics, value)
	}

	selectQuery += joinWithComma(append(finalSelectKeys, selectMetrics...)...)
	orderByQuery := "ORDER BY " + getOrderByClause(isGroupByTimestamp, responseSelectMetrics)
	whereConditionForFilters := getBingAdsFiltersWhereStatement(query.Filters)
	customerAccountIDs := strings.Split(customerAccountID, ",")
	finalParams := make([]interface{}, 0)
	staticWhereParams := []interface{}{projectID, model.BingAdsIntegration, docType, customerAccountIDs, from, to}
	finalParams = append(finalParams, staticWhereParams...)
	if len(groupByCombinationsForGBT) != 0 {
		whereConditionForGBT, whereParams := buildWhereConditionForGBTForBingAds(groupByCombinationsForGBT)
		whereConditionForFilters += (" AND (" + whereConditionForGBT + ")")
		finalParams = append(finalParams, whereParams...)
	}

	resultSQLStatement := selectQuery + fromIntegrationDocuments + staticWhereStatementForBingAds + whereConditionForFilters
	if len(groupByStatement) != 0 {
		resultSQLStatement += "GROUP BY " + groupByStatement
	}
	resultSQLStatement += " " + orderByQuery + limitString + ";"
	return resultSQLStatement, finalParams, responseSelectKeys, responseSelectMetrics
}

func getBingAdsFromStatementWithJoins(filters []model.ChannelFilterV1, groupBys []model.ChannelGroupBy) string {
	logFields := log.Fields{
		"filters":   filters,
		"group_bys": groupBys,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	isPresentCampaignSmartProperty, isPresentAdGroupSmartProperty := checkSmartPropertyWithTypeAndSource(filters, groupBys, "bingads")
	fromStatement := fromIntegrationDocuments
	if isPresentAdGroupSmartProperty {
		fromStatement += "inner join smart_properties ad_group on ad_group.project_id = integration_documents.project_id and ad_group.object_id = document_id "
	}
	if isPresentCampaignSmartProperty {
		fromStatement += "inner join smart_properties campaign on campaign.project_id = integration_documents.project_id and campaign.object_id = document_id "
	}
	return fromStatement
}

func getSQLAndParamsFromBingAdsReportsWithSmartProperty(query *model.ChannelQueryV1, projectID int64, from, to int64, docType int,
	fetchSource bool, limitString string, isGroupByTimestamp bool, groupByCombinationsForGBT []map[string]interface{}, customerAccountID string) (string, []interface{}, []string, []string) {
	logFields := log.Fields{
		"project_id":                    projectID,
		"query":                         query,
		"from":                          from,
		"to":                            to,
		"document_type":                 docType,
		"fetch_source":                  fetchSource,
		"limit_string":                  limitString,
		"is_group_by_timestamp":         isGroupByTimestamp,
		"group_by_combinations_for_gbt": groupByCombinationsForGBT,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	selectQuery := "SELECT "
	selectMetrics := make([]string, 0, 0)
	groupByStatement := ""
	groupByKeysWithoutTimestamp := make([]string, 0, 0)
	selectKeys := make([]string, 0, 0)
	finalSelectKeys := make([]string, 0, 0)
	responseSelectKeys := make([]string, 0, 0)
	responseSelectMetrics := make([]string, 0, 0)

	// SelectKeys

	for _, groupBy := range query.GroupBy {
		_, isPresent := model.SmartPropertyReservedNames[groupBy.Property]
		isSmartProperty := !isPresent
		if isSmartProperty {
			if groupBy.Object == bingadsCampaign {

				value := fmt.Sprintf("JSON_EXTRACT_STRING(campaign.properties, '%s') as campaign_%s", groupBy.Property, groupBy.Property)
				selectKeys = append(selectKeys, value)
				responseSelectKeys = append(responseSelectKeys, fmt.Sprintf("campaign_%s", groupBy.Property))

				groupByKeysWithoutTimestamp = append(groupByKeysWithoutTimestamp, fmt.Sprintf("campaign_%s", groupBy.Property))
			} else {
				value := fmt.Sprintf("JSON_EXTRACT_STRING(ad_group.properties,'%s') as ad_group_%s", groupBy.Property, groupBy.Property)
				selectKeys = append(selectKeys, value)
				responseSelectKeys = append(responseSelectKeys, fmt.Sprintf("ad_group_%s", groupBy.Property))

				groupByKeysWithoutTimestamp = append(groupByKeysWithoutTimestamp, fmt.Sprintf("ad_group_%s", groupBy.Property))
			}
		} else {
			key := groupBy.Object + "." + groupBy.Property
			if groupBy.Object == CAFilterChannel {
				value := fmt.Sprintf("'bingads' as %s", model.BingAdsInternalRepresentationToExternalRepresentationForReports[key])
				selectKeys = append(selectKeys, value)
				responseSelectKeys = append(responseSelectKeys, model.BingAdsInternalRepresentationToExternalRepresentationForReports[key])
			} else {
				value := fmt.Sprintf("%s as %s", objectAndPropertyToValueInBingAdsReportsMapping[key], model.BingAdsInternalRepresentationToExternalRepresentationForReports[key])
				selectKeys = append(selectKeys, value)
				responseSelectKeys = append(responseSelectKeys, model.BingAdsInternalRepresentationToExternalRepresentationForReports[key])
			}
			groupByKeysWithoutTimestamp = append(groupByKeysWithoutTimestamp, model.BingAdsInternalRepresentationToExternalRepresentationForReports[key])
		}
	}

	if isGroupByTimestamp {
		groupByStatement = joinWithComma(append(groupByKeysWithoutTimestamp, model.AliasDateTime)...)
	} else {
		groupByStatement = joinWithComma(groupByKeysWithoutTimestamp...)
	}

	finalSelectKeys = append(finalSelectKeys, selectKeys...)
	if isGroupByTimestamp {
		finalSelectKeys = append(finalSelectKeys, fmt.Sprintf("%s as %s",
			getSelectTimestampByTypeForChannels(query.GetGroupByTimestamp(), query.Timezone), model.AliasDateTime))
		responseSelectKeys = append(responseSelectKeys, model.AliasDateTime)
	}

	for _, selectMetric := range query.SelectMetrics {
		value := fmt.Sprintf("%s as %s", BingAdsMetricsToAggregatesInReportsMapping[selectMetric], model.BingAdsInternalRepresentationToExternalRepresentationForReports[selectMetric])
		selectMetrics = append(selectMetrics, value)

		value = model.BingAdsInternalRepresentationToExternalRepresentationForReports[selectMetric]
		responseSelectMetrics = append(responseSelectMetrics, value)
	}

	selectQuery += joinWithComma(append(finalSelectKeys, selectMetrics...)...)
	orderByQuery := "ORDER BY " + getOrderByClause(isGroupByTimestamp, responseSelectMetrics)
	whereConditionForFilters := getBingAdsFiltersWhereStatementWithSmartProperty(query.Filters)
	filterStatementForSmartPropertyGroupBy := getNotNullFilterStatementForSmartPropertyGroupBys(query.GroupBy)
	finalFilterStatement := joinWithWordInBetween("AND", staticWhereStatementForBingAdsWithSmartProperty, whereConditionForFilters, filterStatementForSmartPropertyGroupBy)
	customerAccountIDs := strings.Split(customerAccountID, ",")
	finalParams := make([]interface{}, 0)
	staticWhereParams := []interface{}{projectID, model.BingAdsIntegration, customerAccountIDs, docType, from, to}
	finalParams = append(finalParams, staticWhereParams...)
	if len(groupByCombinationsForGBT) != 0 {
		whereConditionForGBT, whereParams := buildWhereConditionForGBTForBingAds(groupByCombinationsForGBT)
		whereConditionForFilters += (" AND (" + whereConditionForGBT + ")")
		finalParams = append(finalParams, whereParams...)
	}

	fromStatement := getBingAdsFromStatementWithJoins(query.Filters, query.GroupBy)
	resultSQLStatement := selectQuery + fromStatement + finalFilterStatement
	if len(groupByStatement) != 0 {
		resultSQLStatement += "GROUP BY " + groupByStatement
	}
	resultSQLStatement += " " + orderByQuery + limitString + ";"
	return resultSQLStatement, finalParams, responseSelectKeys, responseSelectMetrics
}

func getBingAdsFiltersWhereStatement(filters []model.ChannelFilterV1) string {
	logFields := log.Fields{
		"filters": filters,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	resultStatement := ""
	var filterValue string
	for index, filter := range filters {
		currentFilterStatement := ""
		if filter.LogicalOp == "" {
			filter.LogicalOp = "AND"
		}
		filterOperator := getOp(filter.Condition, "categorical")
		if filter.Condition == model.ContainsOpStr || filter.Condition == model.NotContainsOpStr {
			filterValue = fmt.Sprintf("%s", filter.Value)
		} else {
			filterValue = filter.Value
		}

		currentFilterStatement = fmt.Sprintf("%s %s '%s' ", objectAndPropertyToValueInBingAdsReportsMapping[filter.Object+"."+filter.Property], filterOperator, filterValue)
		if index == 0 {
			resultStatement = " AND " + currentFilterStatement
		} else {
			resultStatement = fmt.Sprintf("%s %s %s ", resultStatement, filter.LogicalOp, currentFilterStatement)
		}
	}
	return resultStatement
}

func getBingAdsFiltersWhereStatementWithSmartProperty(filters []model.ChannelFilterV1) string {
	logFields := log.Fields{
		"filters": filters,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	resultStatement := ""
	var filterValue string
	campaignFilter := ""
	adGroupFilter := ""
	for index, filter := range filters {
		currentFilterStatement := ""
		if filter.LogicalOp == "" {
			filter.LogicalOp = "AND"
		}
		filterOperator := getOp(filter.Condition, "categorical")
		if filter.Condition == model.ContainsOpStr || filter.Condition == model.NotContainsOpStr {
			filterValue = fmt.Sprintf("%s", filter.Value)
		} else {
			filterValue = filter.Value
		}
		_, isPresent := model.SmartPropertyReservedNames[filter.Property]
		if isPresent {
			currentFilterStatement = fmt.Sprintf("%s %s '%s' ", objectAndPropertyToValueInBingAdsReportsMapping[filter.Object+"."+filter.Property], filterOperator, filterValue)
			if index == 0 {
				resultStatement = " AND " + currentFilterStatement
			} else {
				resultStatement = fmt.Sprintf("%s %s %s ", resultStatement, filter.LogicalOp, currentFilterStatement)
			}
		} else {
			currentFilterStatement = fmt.Sprintf("JSON_EXTRACT_STRING(%s.properties, '%s') %s '%s'", model.BingAdsObjectMapForSmartProperty[filter.Object], filter.Property, filterOperator, filterValue)
			if index == 0 {
				resultStatement = fmt.Sprintf("(%s", currentFilterStatement)
			} else {
				resultStatement = fmt.Sprintf("%s %s %s", resultStatement, filter.LogicalOp, currentFilterStatement)
			}
			if filter.Object == bingadsCampaign {
				campaignFilter = smartPropertyCampaignStaticFilter
			} else {
				adGroupFilter = smartPropertyAdGroupStaticFilter
			}
		}
	}
	if campaignFilter != "" {
		resultStatement += (" AND " + campaignFilter)
	}
	if adGroupFilter != "" {
		resultStatement += (" AND " + adGroupFilter)
	}
	if resultStatement == "" {
		return resultStatement
	}
	return resultStatement + ")"
	return resultStatement
}

func buildWhereConditionForGBTForBingAds(groupByCombinations []map[string]interface{}) (string, []interface{}) {
	logFields := log.Fields{
		"group_by_combinations": groupByCombinations,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	whereConditionForGBT := ""
	params := make([]interface{}, 0)
	filterStringSmartPropertiesCampaign := "campaign.properties"
	filterStringSmartPropertiesAdGroup := "ad_group.properties"
	for _, groupByCombination := range groupByCombinations {
		whereConditionForEachCombination := ""
		for dimension, value := range groupByCombination {
			filterString := ""
			if strings.HasPrefix(dimension, model.CampaignPrefix) {
				key := fmt.Sprintf(`%s.%s`, "campaigns", strings.TrimPrefix(dimension, model.CampaignPrefix))
				currentFilterKey, isPresent := objectAndPropertyToValueInBingAdsReportsMapping[key]
				if isPresent {
					filterString = currentFilterKey
				} else {
					filterString = fmt.Sprintf("JSON_EXTRACT_STRING(%s,'%s')", filterStringSmartPropertiesCampaign, strings.TrimPrefix(dimension, model.CampaignPrefix))
				}

			} else if strings.HasPrefix(dimension, model.AdgroupPrefix) {
				key := fmt.Sprintf(`%s.%s`, "ad_groups", strings.TrimPrefix(dimension, model.AdgroupPrefix))
				currentFilterKey, isPresent := objectAndPropertyToValueInBingAdsReportsMapping[key]
				if isPresent {
					filterString = currentFilterKey
				} else {
					filterString = fmt.Sprintf("JSON_EXTRACT_STRING(%s,'%s')", filterStringSmartPropertiesAdGroup, strings.TrimPrefix(dimension, model.AdgroupPrefix))
				}
			} else {
				key := fmt.Sprintf(`%s.%s`, "keyword", strings.TrimPrefix(dimension, model.KeywordPrefix))
				currentFilterKey := objectAndPropertyToValueInBingAdsReportsMapping[key]
				filterString = currentFilterKey
			}
			if whereConditionForEachCombination == "" {
				if value != nil {
					whereConditionForEachCombination = fmt.Sprintf("%s = ? ", filterString)
					params = append(params, value)
				} else {
					whereConditionForEachCombination = fmt.Sprintf("%s is null ", filterString)
				}
			} else {
				if value != nil {
					whereConditionForEachCombination += fmt.Sprintf(" AND %s = ? ", filterString)
					params = append(params, value)
				} else {
					whereConditionForEachCombination += fmt.Sprintf(" AND %s is null ", filterString)
				}
			}
		}
		if whereConditionForGBT == "" {
			if whereConditionForEachCombination != "" {
				whereConditionForGBT = "(" + whereConditionForEachCombination + ")"
			}
		} else {
			if whereConditionForEachCombination != "" {
				whereConditionForGBT += (" OR (" + whereConditionForEachCombination + ")")
			}
		}
	}

	return whereConditionForGBT, params
}

func (store *MemSQL) GetLatestMetaForBingAdsForGivenDays(projectID int64, days int) ([]model.ChannelDocumentsWithFields, []model.ChannelDocumentsWithFields) {
	logFields := log.Fields{
		"project_id": projectID,
		"days":       days,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	channelDocumentsCampaign := make([]model.ChannelDocumentsWithFields, 0)
	channelDocumentsAdGroup := make([]model.ChannelDocumentsWithFields, 0)

	ftMapping, err := store.GetActiveFiveTranMapping(projectID, model.BingAdsIntegration)
	if err != nil {
		log.WithError(err).Error("Failed to fetch connector id from db")
		return channelDocumentsCampaign, channelDocumentsAdGroup
	}
	customerAccountIDs := strings.Split(ftMapping.Accounts, ",")

	to, err := strconv.ParseUint(time.Now().Format("20060102"), 10, 64)
	if err != nil {
		log.Error("Failed to parse to timestamp")
		return channelDocumentsCampaign, channelDocumentsAdGroup
	}

	from, err := strconv.ParseUint(time.Now().AddDate(0, 0, -days).Format("20060102"), 10, 64)
	if err != nil {
		log.Error("Failed to parse from timestamp")
		return channelDocumentsCampaign, channelDocumentsAdGroup
	}
	query := bingadsAdGroupMetadataFetchQueryStr
	params := []interface{}{model.BingadsDocumentTypeAlias["ad_groups"], projectID, model.BingAdsIntegration, from, to,
		customerAccountIDs, model.BingadsDocumentTypeAlias["ad_groups"], projectID, model.BingAdsIntegration, from, to, customerAccountIDs,
		model.BingadsDocumentTypeAlias["campaigns"], projectID, model.BingAdsIntegration, from, to, customerAccountIDs,
		model.BingadsDocumentTypeAlias["campaigns"], projectID, model.BingAdsIntegration, from, to, customerAccountIDs}

	startExecTime1 := time.Now()
	rows1, tx1, err, queryID1 := store.ExecQueryWithContext(query, params)
	U.LogExecutionTimeWithQueryRequestID(startExecTime1, queryID1, &logFields)
	if err != nil {
		errString := fmt.Sprintf("failed to get last %d ad_group meta for bingads", days)
		log.WithField("error string", err).Error(errString)
		U.CloseReadQuery(rows1, tx1)
		return channelDocumentsCampaign, channelDocumentsAdGroup
	}

	startReadTime1 := time.Now()
	for rows1.Next() {
		currentRecord := model.ChannelDocumentsWithFields{}
		rows1.Scan(&currentRecord.CampaignID, &currentRecord.CampaignName, &currentRecord.AdGroupID, &currentRecord.AdGroupName)
		channelDocumentsAdGroup = append(channelDocumentsAdGroup, currentRecord)
	}
	U.LogReadTimeWithQueryRequestID(startReadTime1, queryID1, &logFields)
	U.CloseReadQuery(rows1, tx1)

	query = bingadsCampaignMetadataFetchQueryStr
	params = []interface{}{model.BingadsDocumentTypeAlias["campaigns"], projectID, model.BingAdsIntegration, from, to,
		customerAccountIDs, model.BingadsDocumentTypeAlias["campaigns"], projectID, model.BingAdsIntegration, from, to, customerAccountIDs}

	startExecTime2 := time.Now()
	rows2, tx2, err, queryID2 := store.ExecQueryWithContext(query, params)
	U.LogExecutionTimeWithQueryRequestID(startExecTime2, queryID2, &logFields)
	if err != nil {
		errString := fmt.Sprintf("failed to get last %d campaign meta for bingads", days)
		log.WithField("error string", err).Error(errString)
		U.CloseReadQuery(rows2, tx2)
		return channelDocumentsCampaign, channelDocumentsAdGroup
	}

	startReadTime2 := time.Now()
	for rows2.Next() {
		currentRecord := model.ChannelDocumentsWithFields{}
		rows2.Scan(&currentRecord.CampaignID, &currentRecord.CampaignName)
		channelDocumentsCampaign = append(channelDocumentsCampaign, currentRecord)
	}
	U.LogReadTimeWithQueryRequestID(startReadTime2, queryID2, &logFields)
	U.CloseReadQuery(rows2, tx2)

	return channelDocumentsCampaign, channelDocumentsAdGroup
}

// PullBingAdsRows - Function to pull all bing integration documents
// Selecting VALUE, TIMESTAMP, TYPE from integration_documents and PROPERTIES, OBJECT_TYPE from smart_properties
// Left join smart_properties filtered by project_id and source=bingads
// where integration_documents.value["campaign_id"] = smart_properties.object_id (when smart_properties.object_type = 1)
//	 or integration_documents.value["ad_group_id"] = smart_properties.object_id (when smart_properties.object_type = 2)
// [make sure there aren't multiple smart_properties rows for a particular object,
// or weekly insights for bing would show incorrect data.]
func (store *MemSQL) PullBingAdsRows(projectID int64, startTime, endTime int64) (*sql.Rows, *sql.Tx, error) {
	logFields := log.Fields{
		"project_id": projectID,
		"start_time": startTime,
		"end_time":   endTime,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	year, month, date := time.Unix(startTime, 0).Date()
	start := year*10000 + int(month)*100 + date + 1

	year, month, date = time.Unix(endTime, 0).Date()
	end := year*10000 + int(month)*100 + date

	rawQuery := fmt.Sprintf("SELECT bing.document_id, bing.value, bing.timestamp, bing.type, sp.properties FROM integration_documents bing "+
		"LEFT JOIN smart_properties sp ON sp.project_id = %d AND sp.source = '%s' AND "+
		"((COALESCE(sp.object_type,1) = 1 AND (sp.object_id = JSON_EXTRACT_STRING(bing.value, 'campaign_id') OR sp.object_id = JSON_EXTRACT_STRING(bing.value, 'base_campaign_id'))) OR "+
		"(COALESCE(sp.object_type,2) = 2 AND (sp.object_id = JSON_EXTRACT_STRING(bing.value, 'ad_group_id') OR sp.object_id = JSON_EXTRACT_STRING(bing.value, 'base_ad_group_id')))) "+
		"WHERE bing.project_id = %d AND bing.timestamp BETWEEN %d AND %d "+
		"ORDER BY bing.type, bing.timestamp LIMIT %d",
		projectID, model.ChannelBingAds, projectID, start, end, model.BingPullLimit+1)

	rows, tx, err, _ := store.ExecQueryWithContext(rawQuery, []interface{}{})
	return rows, tx, err
}
