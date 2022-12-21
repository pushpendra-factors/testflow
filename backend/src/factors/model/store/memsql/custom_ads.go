package memsql

import (
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
	customadsCampaign                                 = "campaigns"
	customadsAdGroup                                  = "ad_groups"
	customadsKeyword                                  = "keyword"
	staticWhereStatementForCustomAds                  = "WHERE project_id = ? AND source= ? AND document_type = ? AND customer_account_id IN (?) AND timestamp between ? AND ? "
	staticWhereStatementForCustomAdsWithSmartProperty = "WHERE integration_documents.project_id = ? AND integration_documents.source= ?  AND integration_documents.customer_account_id IN ( ? ) AND integration_documents.document_type = ? AND integration_documents.timestamp between ? AND ? "
	customadsFilterQueryStr                           = "SELECT DISTINCT(LCASE(JSON_EXTRACT_STRING(value, ?))) as filter_value FROM integration_documents WHERE project_id = ? AND customer_account_id IN (?) AND document_type = ? AND JSON_EXTRACT_STRING(value, ?) IS NOT NULL AND timestamp BETWEEN ? AND ? AND source IN (?) LIMIT 5000"
)

var objectAndPropertyToValueInCustomAdsReportsMapping = map[string]string{
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

var CustomAdsMetricsToAggregatesInReportsMapping = map[string]string{
	"impressions": "SUM(JSON_EXTRACT_STRING(value, 'impressions'))",
	"clicks":      "SUM(JSON_EXTRACT_STRING(value, 'clicks'))",
	"spend":       "SUM(JSON_EXTRACT_STRING(value, 'spend'))",
	"conversions": "SUM(JSON_EXTRACT_STRING(value, 'conversions'))",
}

const customadsAdGroupMetadataFetchQueryStr = "WITH ad_group as (select ad_group_information.campaign_id_1 as campaign_id, ad_group_information.ad_group_id_1 as ad_group_id, ad_group_information.ad_group_name_1 as ad_group_name " +
	"from ( " +
	"select JSON_EXTRACT_STRING(value, 'campaign_id') as campaign_id_1, document_id as ad_group_id_1, JSON_EXTRACT_STRING(value, 'name') as ad_group_name_1, timestamp " +
	"from integration_documents where document_type = ? AND project_id = ? AND source IN (?) AND timestamp between ? AND ? AND customer_account_id IN (?) " +
	") as ad_group_information " +
	"INNER JOIN " +
	"(select document_id as ad_group_id_1, max(timestamp) as timestamp " +
	"from integration_documents where document_type = ? AND project_id = ? AND source IN (?) AND timestamp between ? AND ? AND customer_account_id IN (?) group by ad_group_id_1 " +
	") as ad_group_latest_timestamp_id " +
	"ON ad_group_information.ad_group_id_1 = ad_group_latest_timestamp_id.ad_group_id_1 AND ad_group_information.timestamp = ad_group_latest_timestamp_id.timestamp), " +

	" campaign as (select campaign_information.campaign_id_1 as campaign_id, campaign_information.campaign_name_1 as campaign_name " +
	"from ( " +
	"select document_id as campaign_id_1, JSON_EXTRACT_STRING(value, 'name') as campaign_name_1, timestamp " +
	"from integration_documents where document_type = ? AND project_id = ? AND source IN (?) AND timestamp between ? AND ? AND customer_account_id IN (?) " +
	") as campaign_information " +
	"INNER JOIN " +
	"(select document_id as campaign_id_1, max(timestamp) as timestamp " +
	"from integration_documents where document_type = ? AND project_id = ? AND source IN (?) AND timestamp between ? AND ? AND customer_account_id IN (?) group by campaign_id_1 " +
	") as campaign_latest_timestamp_id " +
	"ON campaign_information.campaign_id_1 = campaign_latest_timestamp_id.campaign_id_1 AND campaign_information.timestamp = campaign_latest_timestamp_id.timestamp) " +

	"select campaign.campaign_id as campaign_id, campaign.campaign_name as campaign_name, ad_group.ad_group_id as ad_group_id, ad_group.ad_group_name as ad_group_name " +
	"from campaign join ad_group on ad_group.campaign_id = campaign.campaign_id"

const customadsCampaignMetadataFetchQueryStr = "select campaign_information.campaign_id as campaign_id, campaign_information.campaign_name as campaign_name " +
	"from ( " +
	"select document_id AS campaign_id, JSON_EXTRACT_STRING(value, 'name') as campaign_name, timestamp " +
	"from integration_documents where document_type = ? AND project_id = ? AND source IN (?) AND timestamp between ? AND ? AND customer_account_id IN (?) " +
	") as campaign_information " +
	"INNER JOIN " +
	"(select document_id AS campaign_id, max(timestamp) as timestamp " +
	"from integration_documents where document_type = ? AND project_id = ? AND source IN (?) AND timestamp between ? AND ? AND customer_account_id IN (?) group by campaign_id " +
	") as campaign_latest_timestamp_id " +
	"ON campaign_information.campaign_id = campaign_latest_timestamp_id.campaign_id AND campaign_information.timestamp = campaign_latest_timestamp_id.timestamp "

func (store *MemSQL) buildObjectAndPropertiesForCustomAds(projectID int64, source string, objects []string) []model.ChannelObjectAndProperties {
	objectsAndProperties := make([]model.ChannelObjectAndProperties, 0)
	for _, currentObject := range objects {
		// to do: check if normal properties present then only smart properties will be there
		propertiesAndRelated, isPresent := model.MapOfCustomAdsObjectsToPropertiesAndRelated[currentObject]
		var currentProperties []model.ChannelProperty
		var currentPropertiesSmart []model.ChannelProperty
		if isPresent {
			currentProperties = buildProperties(propertiesAndRelated)
			smartProperty := store.GetSmartPropertyAndRelated(projectID, currentObject, source)
			currentPropertiesSmart = buildProperties(smartProperty)
			currentProperties = append(currentProperties, currentPropertiesSmart...)
		} else {
			currentProperties = buildProperties(allChannelsPropertyToRelated)
			smartProperty := store.GetSmartPropertyAndRelated(projectID, currentObject, source)
			currentPropertiesSmart = buildProperties(smartProperty)
			currentProperties = append(currentProperties, currentPropertiesSmart...)
		}
		objectsAndProperties = append(objectsAndProperties, buildObjectsAndProperties(currentProperties, []string{currentObject})...)
	}
	return objectsAndProperties
}

func (store *MemSQL) GetCustomAdsFilterValues(projectID int64, requestFilterObject string, requestFilterProperty string, source string, reqID string) ([]interface{}, int) {
	logFields := log.Fields{
		"project_id":              projectID,
		"request_filter_object":   requestFilterObject,
		"request_filter_property": requestFilterProperty,
		"req_id":                  reqID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)
	logCtx = log.WithField("project_id", projectID).WithField("req_id", reqID)
	exist := store.IsCustomAdsAvailable(projectID)
	if !exist {
		return []interface{}{}, http.StatusOK
	}
	sources := make([]string, 0)
	if source == CAAllChannelAds {
		sources, _ = store.GetCustomAdsSourcesByProject(projectID)
	} else {
		sources = []string{source}
	}
	accounts, _ := store.GetCustomAdsAccountsByProject(projectID, sources)
	_, isPresent := model.SmartPropertyReservedNames[requestFilterProperty]
	if !isPresent {
		filterValues, errCode := store.getSmartPropertyFilterValues(projectID, requestFilterObject, requestFilterProperty, source, reqID)
		if errCode != http.StatusFound {
			return []interface{}{}, http.StatusInternalServerError
		}
		return filterValues, http.StatusFound
	}
	requestFilterProperty = strings.TrimPrefix(requestFilterProperty, fmt.Sprintf("%v_", requestFilterObject))
	docType := model.CustomadsDocumentTypeAlias[model.CustomAdsObjectInternalRepresentationToExternalRepresentation[requestFilterObject]]
	filterProperty := model.CustomAdsInternalRepresentationToExternalRepresentation[fmt.Sprintf("%v.%v", model.CustomAdsObjectInternalRepresentationToExternalRepresentation[requestFilterObject], requestFilterProperty)]

	from, to := model.GetFromAndToDatesForFilterValues()
	params := []interface{}{filterProperty, projectID, accounts, docType, filterProperty, from, to, sources}
	_, resultRows, err := store.ExecuteSQL(customadsFilterQueryStr, params, logCtx)
	if err != nil {
		logCtx.WithError(err).WithField("query", customadsFilterQueryStr).WithField("params", params).Error(model.CustomAdsSpecificError)
		return make([]interface{}, 0), http.StatusInternalServerError
	}
	return Convert2DArrayTo1DArray(resultRows), http.StatusFound
}

func (store *MemSQL) GetCustomadsFilterValuesSQLAndParams(projectID int64, requestFilterObject string, requestFilterProperty string, source string, reqID string) (string, []interface{}, int) {
	logFields := log.Fields{
		"project_id":              projectID,
		"request_filter_object":   requestFilterObject,
		"request_filter_property": requestFilterProperty,
		"req_id":                  reqID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	logCtx := log.WithFields(logFields)
	logCtx.WithField("project_id", projectID).WithField("req_id", reqID)
	exist := store.IsCustomAdsAvailable(projectID)
	if !exist {
		return "", nil, http.StatusNotFound
	}
	sources := make([]string, 0)
	if source == CAAllChannelAds {
		sources, _ = store.GetCustomAdsSourcesByProject(projectID)
	} else {
		sources = []string{source}
	}
	accounts, _ := store.GetCustomAdsAccountsByProject(projectID, sources)
	requestFilterProperty = strings.TrimPrefix(requestFilterProperty, fmt.Sprintf("%v_", requestFilterObject))
	docType := model.CustomadsDocumentTypeAlias[model.CustomAdsObjectInternalRepresentationToExternalRepresentation[requestFilterObject]]
	from, to := model.GetFromAndToDatesForFilterValues()
	filterProperty := model.CustomAdsInternalRepresentationToExternalRepresentation[fmt.Sprintf("%v.%v", model.CustomAdsObjectInternalRepresentationToExternalRepresentation[requestFilterObject], requestFilterProperty)]
	params := []interface{}{filterProperty, projectID, accounts, docType, filterProperty, from, to, sources}
	return customadsFilterQueryStr, params, http.StatusFound
}

func (store *MemSQL) ExecuteCustomAdsChannelQueryV1(projectID int64, query *model.ChannelQueryV1, reqID string) ([]string, [][]interface{}, int) {
	logFields := log.Fields{
		"project_id": projectID,
		"query":      query,
		"req_id":     reqID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	fetchSource := false
	defer U.NotifyOnPanicWithError(C.GetConfig().Env, C.GetConfig().AppName)
	logCtx := log.WithField("xreq_id", reqID)
	limitString := ""
	if C.IsKPILimitIncreaseAllowedForProject(projectID) {
		limitString = fmt.Sprintf(" LIMIT %d", model.MaxResultsLimit)
	} else {
		limitString = fmt.Sprintf(" LIMIT %d", model.ResultsLimit)
	}
	if query.GroupByTimestamp == "" {
		sql, params, selectKeys, selectMetrics, errCode := store.GetSQLQueryAndParametersForCustomAdsQueryV1(projectID,
			query, reqID, fetchSource, limitString, false, nil)
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
			logCtx.WithError(err).WithField("query", sql).WithField("params", params).Error("failed in custom ads with error")
			return columns, make([][]interface{}, 0, 0), http.StatusInternalServerError
		}
		return columns, resultMetrics, http.StatusOK
	} else {
		sql, params, selectKeys, selectMetrics, errCode := store.GetSQLQueryAndParametersForCustomAdsQueryV1(
			projectID, query, reqID, fetchSource, " LIMIT 1000", false, nil)
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
			logCtx.WithError(err).WithField("query", sql).WithField("params", params).Error("failed in custom ads with error")
			return columns, make([][]interface{}, 0, 0), http.StatusInternalServerError
		}
		groupByCombinations := model.GetGroupByCombinationsForChannelAnalytics(columns, resultMetrics)
		sql, params, selectKeys, selectMetrics, errCode = store.GetSQLQueryAndParametersForCustomAdsQueryV1(
			projectID, query, reqID, fetchSource, limitString, true, groupByCombinations)
		if errCode != http.StatusOK {
			headers := model.GetHeadersFromQuery(*query)
			return headers, make([][]interface{}, 0, 0), errCode
		}
		_, resultMetrics, err = store.ExecuteSQL(sql, params, logCtx)
		columns = append(selectKeys, selectMetrics...)
		if err != nil {
			logCtx.WithError(err).WithField("query", sql).WithField("params", params).Error("Failed in custom ads with the error.")
			return columns, make([][]interface{}, 0, 0), http.StatusInternalServerError
		}
		return columns, resultMetrics, http.StatusOK
	}
}

func (store *MemSQL) GetSQLQueryAndParametersForCustomAdsQueryV1(projectID int64, query *model.ChannelQueryV1, reqID string, fetchSource bool,
	limitString string, isGroupByTimestamp bool, groupByCombinationsForGBT map[string][]interface{}) (string, []interface{}, []string, []string, int) {
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
	transformedQuery, customerAccountID, err := store.transFormRequestFieldsAndFetchRequiredFieldsForCustomads(
		projectID, *query, reqID)
	if err != nil && err.Error() == "record not found" {
		logCtx.WithError(err).Info(model.CustomAdsSpecificError)
		return "", nil, make([]string, 0, 0), make([]string, 0, 0), http.StatusNotFound
	}
	if err != nil {
		logCtx.WithError(err).Error(model.CustomAdsSpecificError)
		return "", make([]interface{}, 0, 0), make([]string, 0, 0), make([]string, 0, 0), http.StatusBadRequest
	}
	// smart properties check
	isSmartPropertyPresent := checkSmartProperty(query.Filters, query.GroupBy)
	if isSmartPropertyPresent {
		sql, params, selectKeys, selectMetrics, err = buildCustomAdsQueryWithSmartProperty(transformedQuery, projectID, customerAccountID, fetchSource, limitString, isGroupByTimestamp, groupByCombinationsForGBT)
		if err != nil {
			return "", make([]interface{}, 0, 0), make([]string, 0, 0), make([]string, 0, 0), http.StatusInternalServerError
		}
		return sql, params, selectKeys, selectMetrics, http.StatusOK
	}
	sql, params, selectKeys, selectMetrics, err = buildCustomAdsQueryV1(transformedQuery, projectID, customerAccountID, fetchSource, limitString, isGroupByTimestamp, groupByCombinationsForGBT)
	if err != nil {
		return "", make([]interface{}, 0, 0), make([]string, 0, 0), make([]string, 0, 0), http.StatusInternalServerError
	}
	return sql, params, selectKeys, selectMetrics, http.StatusOK
}

func (store *MemSQL) transFormRequestFieldsAndFetchRequiredFieldsForCustomads(projectID int64,
	query model.ChannelQueryV1, reqID string) (*model.ChannelQueryV1, []string, error) {
	logFields := log.Fields{
		"project_id": projectID,
		"query":      query,
		"req_id":     reqID,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	var transformedQuery model.ChannelQueryV1
	var err error
	logCtx := log.WithFields(logFields)
	accounts, _ := store.GetCustomAdsAccountsByProject(projectID, []string{query.Channel})
	transformedQuery, err = convertFromRequestToCustomAdsSpecificRepresentation(query)
	if err != nil {
		logCtx.Warn("Request failed in validation: ", err)
		return &model.ChannelQueryV1{}, nil, err
	}
	return &transformedQuery, accounts, nil
}

func convertFromRequestToCustomAdsSpecificRepresentation(query model.ChannelQueryV1) (model.ChannelQueryV1, error) {
	logFields := log.Fields{
		"query": query,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	var transformedQuery model.ChannelQueryV1
	var err1, err2, err3 error
	transformedQuery.SelectMetrics, err1 = getCustomAdsSpecificMetrics(query.SelectMetrics)
	transformedQuery.Filters, err2 = getCustomAdsSpecificFilters(query.Filters)
	transformedQuery.GroupBy, err3 = getCustomAdsSpecificGroupBy(query.GroupBy)
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
	transformedQuery.Channel = query.Channel
	return transformedQuery, nil
}

func buildCustomAdsQueryWithSmartProperty(query *model.ChannelQueryV1, projectID int64, customerAccountID []string, fetchSource bool,
	limitString string, isGroupByTimestamp bool, groupByCombinationsForGBT map[string][]interface{}) (string, []interface{}, []string, []string, error) {
	logFields := log.Fields{
		"project_id":                    projectID,
		"query":                         query,
		"fetch_source":                  fetchSource,
		"limit_string":                  limitString,
		"is_group_by_timestamp":         isGroupByTimestamp,
		"group_by_combinations_for_gbt": groupByCombinationsForGBT,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	lowestHierarchyLevel := getLowestHierarchyLevelForCustomAds(query)
	lowestHierarchyReportLevel := model.CustomAdsObjectToPerfomanceReportRepresentation[lowestHierarchyLevel] // suffix tbd
	sql, params, selectKeys, selectMetrics := getSQLAndParamsFromCustomAdsReportsWithSmartProperty(query, projectID, query.From, query.To, model.CustomadsDocumentTypeAlias[lowestHierarchyReportLevel],
		fetchSource, limitString, isGroupByTimestamp, groupByCombinationsForGBT, customerAccountID)
	return sql, params, selectKeys, selectMetrics, nil
}

func buildCustomAdsQueryV1(query *model.ChannelQueryV1, projectID int64, customerAccountID []string, fetchSource bool,
	limitString string, isGroupByTimestamp bool, groupByCombinationsForGBT map[string][]interface{}) (string, []interface{}, []string, []string, error) {
	logFields := log.Fields{
		"project_id":                    projectID,
		"query":                         query,
		"fetch_source":                  fetchSource,
		"limit_string":                  limitString,
		"is_group_by_timestamp":         isGroupByTimestamp,
		"group_by_combinations_for_gbt": groupByCombinationsForGBT,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	lowestHierarchyLevel := getLowestHierarchyLevelForCustomAds(query)
	lowestHierarchyReportLevel := model.CustomAdsObjectToPerfomanceReportRepresentation[lowestHierarchyLevel] // suffix tbd
	sql, params, selectKeys, selectMetrics := getSQLAndParamsFromCustomAdsReports(query, projectID, query.From, query.To, model.CustomadsDocumentTypeAlias[lowestHierarchyReportLevel],
		fetchSource, limitString, isGroupByTimestamp, groupByCombinationsForGBT, customerAccountID)
	return sql, params, selectKeys, selectMetrics, nil
}

func getCustomAdsSpecificMetrics(requestSelectMetrics []string) ([]string, error) {
	logFields := log.Fields{
		"request_select_metrics": requestSelectMetrics,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	resultMetrics := make([]string, 0, 0)
	for _, requestMetric := range requestSelectMetrics {
		metric, isPresent := model.CustomAdsInternalRepresentationToExternalRepresentationForReports[requestMetric]
		if !isPresent {
			return make([]string, 0, 0), errors.New("Invalid metric found for document type")
		}
		resultMetrics = append(resultMetrics, metric)
	}
	return resultMetrics, nil
}

func getCustomAdsSpecificFilters(requestFilters []model.ChannelFilterV1) ([]model.ChannelFilterV1, error) {
	logFields := log.Fields{
		"request_filters": requestFilters,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	resultFilters := make([]model.ChannelFilterV1, 0, 0)
	for _, requestFilter := range requestFilters {
		var resultFilter model.ChannelFilterV1
		filterObject, isPresent := model.CustomAdsObjectInternalRepresentationToExternalRepresentation[requestFilter.Object]
		if !isPresent {
			return make([]model.ChannelFilterV1, 0, 0), errors.New("Invalid filter key found for document type")
		}
		resultFilter = requestFilter
		resultFilter.Object = filterObject
		resultFilters = append(resultFilters, resultFilter)
	}
	return resultFilters, nil
}

func getCustomAdsSpecificGroupBy(requestGroupBys []model.ChannelGroupBy) ([]model.ChannelGroupBy, error) {

	resultGroupBys := make([]model.ChannelGroupBy, 0, 0)
	for _, requestGroupBy := range requestGroupBys {
		var resultGroupBy model.ChannelGroupBy
		groupByObject, isPresent := model.CustomAdsObjectInternalRepresentationToExternalRepresentation[requestGroupBy.Object]
		if !isPresent {
			return make([]model.ChannelGroupBy, 0, 0), errors.New("Invalid groupby key found for document type")
		}
		resultGroupBy = requestGroupBy
		resultGroupBy.Object = groupByObject
		resultGroupBys = append(resultGroupBys, resultGroupBy)
	}
	return resultGroupBys, nil
}

func getLowestHierarchyLevelForCustomAds(query *model.ChannelQueryV1) string {
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
		if objectName == customadsKeyword {
			return customadsKeyword
		}
	}

	for _, objectName := range objectNames {
		if objectName == customadsAdGroup {
			return customadsAdGroup
		}
	}

	for _, objectName := range objectNames {
		if objectName == customadsCampaign {
			return customadsCampaign
		}
	}
	return customadsCampaign
}

func getSQLAndParamsFromCustomAdsReportsWithSmartProperty(query *model.ChannelQueryV1, projectID int64, from, to int64, docType int,
	fetchSource bool, limitString string, isGroupByTimestamp bool, groupByCombinationsForGBT map[string][]interface{}, customerAccountID []string) (string, []interface{}, []string, []string) {
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
			if groupBy.Object == customadsCampaign {

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
				value := fmt.Sprintf("'%s' as %s", query.Channel, model.CustomAdsInternalRepresentationToExternalRepresentationForReports[key])
				selectKeys = append(selectKeys, value)
				responseSelectKeys = append(responseSelectKeys, model.CustomAdsInternalRepresentationToExternalRepresentationForReports[key])
			} else {
				value := fmt.Sprintf("%s as %s", objectAndPropertyToValueInCustomAdsReportsMapping[key], model.CustomAdsInternalRepresentationToExternalRepresentationForReports[key])
				selectKeys = append(selectKeys, value)
				responseSelectKeys = append(responseSelectKeys, model.CustomAdsInternalRepresentationToExternalRepresentationForReports[key])
			}
			groupByKeysWithoutTimestamp = append(groupByKeysWithoutTimestamp, model.CustomAdsInternalRepresentationToExternalRepresentationForReports[key])
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
		value := fmt.Sprintf("%s as %s", CustomAdsMetricsToAggregatesInReportsMapping[selectMetric], model.CustomAdsInternalRepresentationToExternalRepresentationForReports[selectMetric])
		selectMetrics = append(selectMetrics, value)

		value = model.CustomAdsInternalRepresentationToExternalRepresentationForReports[selectMetric]
		responseSelectMetrics = append(responseSelectMetrics, value)
	}

	selectQuery += joinWithComma(append(finalSelectKeys, selectMetrics...)...)
	orderByQuery := "ORDER BY " + getOrderByClause(isGroupByTimestamp, responseSelectMetrics)
	whereConditionForFilters := getCustomAdsFiltersWhereStatementWithSmartProperty(query.Filters)
	filterStatementForSmartPropertyGroupBy := getNotNullFilterStatementForSmartPropertyGroupBys(query.GroupBy)
	finalFilterStatement := joinWithWordInBetween("AND", staticWhereStatementForCustomAdsWithSmartProperty, whereConditionForFilters, filterStatementForSmartPropertyGroupBy)
	finalParams := make([]interface{}, 0)
	staticWhereParams := []interface{}{projectID, query.Channel, customerAccountID, docType, from, to}
	finalParams = append(finalParams, staticWhereParams...)
	if len(groupByCombinationsForGBT) != 0 {
		whereConditionForGBT, whereParams := buildWhereConditionForGBTForCustomAds(groupByCombinationsForGBT)
		whereConditionForFilters += (" AND (" + whereConditionForGBT + ")")
		finalParams = append(finalParams, whereParams...)
	}

	fromStatement := getCustomAdsFromStatementWithJoins(query.Filters, query.GroupBy)
	resultSQLStatement := selectQuery + fromStatement + finalFilterStatement
	if len(groupByStatement) != 0 {
		resultSQLStatement += "GROUP BY " + groupByStatement
	}
	resultSQLStatement += " " + orderByQuery + limitString + ";"
	return resultSQLStatement, finalParams, responseSelectKeys, responseSelectMetrics
}

func getSQLAndParamsFromCustomAdsReports(query *model.ChannelQueryV1, projectID int64, from, to int64, docType int,
	fetchSource bool, limitString string, isGroupByTimestamp bool, groupByCombinationsForGBT map[string][]interface{}, customerAccountID []string) (string, []interface{}, []string, []string) {
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
		groupByKeysWithoutTimestamp = append(groupByKeysWithoutTimestamp, model.CustomAdsInternalRepresentationToExternalRepresentationForReports[key])
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
			value := fmt.Sprintf("'%s' as %s", query.Channel, model.CustomAdsInternalRepresentationToExternalRepresentationForReports[key])
			selectKeys = append(selectKeys, value)
			responseSelectKeys = append(responseSelectKeys, model.CustomAdsInternalRepresentationToExternalRepresentationForReports[key])
		} else {
			value := fmt.Sprintf("%s as %s", objectAndPropertyToValueInCustomAdsReportsMapping[key], model.CustomAdsInternalRepresentationToExternalRepresentationForReports[key])
			selectKeys = append(selectKeys, value)
			responseSelectKeys = append(responseSelectKeys, model.CustomAdsInternalRepresentationToExternalRepresentationForReports[key])
		}
	}

	finalSelectKeys = append(finalSelectKeys, selectKeys...)
	if isGroupByTimestamp {
		finalSelectKeys = append(finalSelectKeys, fmt.Sprintf("%s as %s",
			getSelectTimestampByTypeForChannels(query.GetGroupByTimestamp(), query.Timezone), model.AliasDateTime))
		responseSelectKeys = append(responseSelectKeys, model.AliasDateTime)
	}

	for _, selectMetric := range query.SelectMetrics {
		value := fmt.Sprintf("%s as %s", CustomAdsMetricsToAggregatesInReportsMapping[selectMetric], model.CustomAdsInternalRepresentationToExternalRepresentationForReports[selectMetric])
		selectMetrics = append(selectMetrics, value)

		value = model.CustomAdsInternalRepresentationToExternalRepresentationForReports[selectMetric]
		responseSelectMetrics = append(responseSelectMetrics, value)
	}

	selectQuery += joinWithComma(append(finalSelectKeys, selectMetrics...)...)
	orderByQuery := "ORDER BY " + getOrderByClause(isGroupByTimestamp, responseSelectMetrics)
	whereConditionForFilters := getCustomAdsFiltersWhereStatement(query.Filters)
	finalParams := make([]interface{}, 0)
	staticWhereParams := []interface{}{projectID, query.Channel, docType, customerAccountID, from, to}
	finalParams = append(finalParams, staticWhereParams...)
	if len(groupByCombinationsForGBT) != 0 {
		whereConditionForGBT, whereParams := buildWhereConditionForGBTForCustomAds(groupByCombinationsForGBT)
		whereConditionForFilters += (" AND (" + whereConditionForGBT + ")")
		finalParams = append(finalParams, whereParams...)
	}

	resultSQLStatement := selectQuery + fromIntegrationDocuments + staticWhereStatementForCustomAds + whereConditionForFilters
	if len(groupByStatement) != 0 {
		resultSQLStatement += "GROUP BY " + groupByStatement
	}
	resultSQLStatement += " " + orderByQuery + limitString + ";"
	return resultSQLStatement, finalParams, responseSelectKeys, responseSelectMetrics
}

func getCustomAdsFiltersWhereStatementWithSmartProperty(filters []model.ChannelFilterV1) string {
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
			currentFilterStatement = fmt.Sprintf("%s %s '%s' ", objectAndPropertyToValueInCustomAdsReportsMapping[filter.Object+"."+filter.Property], filterOperator, filterValue)
			if index == 0 {
				resultStatement = " AND " + currentFilterStatement
			} else {
				resultStatement = fmt.Sprintf("%s %s %s ", resultStatement, filter.LogicalOp, currentFilterStatement)
			}
		} else {
			currentFilterStatement = fmt.Sprintf("JSON_EXTRACT_STRING(%s.properties, '%s') %s '%s'", model.CustomAdsObjectMapForSmartProperty[filter.Object], filter.Property, filterOperator, filterValue)
			if index == 0 {
				resultStatement = fmt.Sprintf("(%s", currentFilterStatement)
			} else {
				resultStatement = fmt.Sprintf("%s %s %s", resultStatement, filter.LogicalOp, currentFilterStatement)
			}
			if filter.Object == customadsCampaign {
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

func buildWhereConditionForGBTForCustomAds(groupByCombinations map[string][]interface{}) (string, []interface{}) {
	logFields := log.Fields{
		"group_by_combinations": groupByCombinations,
	}

	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	resultantWhereCondition := ""
	resultantInClauses := make([]string, 0)
	params := make([]interface{}, 0)

	for dimension, values := range groupByCombinations {
		currentInClause := ""

		jsonExtractExpression := GetFilterObjectExpressionForChannelCustomAds(dimension)

		valuesInString := make([]string, 0)
		for _, value := range values {
			valuesInString = append(valuesInString, "?")
			params = append(params, value)
		}
		currentInClause = joinWithComma(valuesInString...)

		resultantInClauses = append(resultantInClauses, jsonExtractExpression+" IN ("+currentInClause+") ")
	}
	resultantWhereCondition = joinWithWordInBetween("AND", resultantInClauses...)

	return resultantWhereCondition, params

}

func getCustomAdsFromStatementWithJoins(filters []model.ChannelFilterV1, groupBys []model.ChannelGroupBy) string {
	logFields := log.Fields{
		"filters":   filters,
		"group_bys": groupBys,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	isPresentCampaignSmartProperty, isPresentAdGroupSmartProperty := checkSmartPropertyWithTypeAndSource(filters, groupBys, "customads")
	fromStatement := fromIntegrationDocuments
	if isPresentAdGroupSmartProperty {
		fromStatement += "inner join smart_properties ad_group on ad_group.project_id = integration_documents.project_id and ad_group.object_id = document_id "
	}
	if isPresentCampaignSmartProperty {
		fromStatement += "inner join smart_properties campaign on campaign.project_id = integration_documents.project_id and campaign.object_id = document_id "
	}
	return fromStatement
}

func getCustomAdsFiltersWhereStatement(filters []model.ChannelFilterV1) string {
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

		currentFilterStatement = fmt.Sprintf("%s %s '%s' ", objectAndPropertyToValueInCustomAdsReportsMapping[filter.Object+"."+filter.Property], filterOperator, filterValue)
		if index == 0 {
			resultStatement = " AND " + currentFilterStatement
		} else {
			resultStatement = fmt.Sprintf("%s %s %s ", resultStatement, filter.LogicalOp, currentFilterStatement)
		}
	}
	return resultStatement
}

func GetFilterObjectExpressionForChannelCustomAds(dimension string) string {
	filterObjectForSmartPropertiesCampaign := "campaign.properties"
	filterObjectForSmartPropertiesAdGroup := "ad_group.properties"

	filterExpression := ""
	isNotSmartProperty := false
	if strings.HasPrefix(dimension, model.CampaignPrefix) {
		filterExpression, isNotSmartProperty = GetFilterExpressionIfPresentForCustomAds("campaigns", dimension, model.CampaignPrefix)
		if !isNotSmartProperty {
			filterExpression = fmt.Sprintf("JSON_EXTRACT_STRING(%s,'%s')", filterObjectForSmartPropertiesCampaign, strings.TrimPrefix(dimension, model.CampaignPrefix))
		}
	} else if strings.HasPrefix(dimension, model.AdgroupPrefix) {
		filterExpression, isNotSmartProperty = GetFilterExpressionIfPresentForCustomAds("ad_groups", dimension, model.AdgroupPrefix)
		if !isNotSmartProperty {
			filterExpression = fmt.Sprintf("JSON_EXTRACT_STRING(%s,'%s')", filterObjectForSmartPropertiesAdGroup, strings.TrimPrefix(dimension, model.AdgroupPrefix))
		}
	} else {
		filterExpression, _ = GetFilterExpressionIfPresentForCustomAds("keyword", dimension, model.KeywordPrefix)
	}
	return filterExpression
}

func GetFilterExpressionIfPresentForCustomAds(objectType, dimension, prefix string) (string, bool) {
	key := fmt.Sprintf(`%s.%s`, objectType, strings.TrimPrefix(dimension, prefix))
	reportProperty, isPresent := objectAndPropertyToValueInCustomAdsReportsMapping[key]
	return reportProperty, isPresent
}

func (store *MemSQL) GetLatestMetaForCustomAdsForGivenDays(projectID int64, source string, days int) ([]model.ChannelDocumentsWithFields, []model.ChannelDocumentsWithFields) {
	logFields := log.Fields{
		"project_id": projectID,
		"days":       days,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	channelDocumentsCampaign := make([]model.ChannelDocumentsWithFields, 0)
	channelDocumentsAdGroup := make([]model.ChannelDocumentsWithFields, 0)

	customerAccountIDs, _ := store.GetCustomAdsAccountsByProject(projectID, []string{source})

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
	query := customadsAdGroupMetadataFetchQueryStr
	params := []interface{}{model.CustomadsDocumentTypeAlias["ad_groups"], projectID, source, from, to,
		customerAccountIDs, model.CustomadsDocumentTypeAlias["ad_groups"], projectID, source, from, to, customerAccountIDs,
		model.CustomadsDocumentTypeAlias["campaigns"], projectID, source, from, to, customerAccountIDs,
		model.CustomadsDocumentTypeAlias["campaigns"], projectID, source, from, to, customerAccountIDs}

	startExecTime1 := time.Now()
	rows1, tx1, err, queryID1 := store.ExecQueryWithContext(query, params)
	U.LogExecutionTimeWithQueryRequestID(startExecTime1, queryID1, &logFields)
	if err != nil {
		errString := fmt.Sprintf("failed to get last %d ad_group meta for custom ads", days)
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

	query = customadsCampaignMetadataFetchQueryStr
	params = []interface{}{model.CustomadsDocumentTypeAlias["campaigns"], projectID, source, from, to,
		customerAccountIDs, model.CustomadsDocumentTypeAlias["campaigns"], projectID, source, from, to, customerAccountIDs}

	startExecTime2 := time.Now()
	rows2, tx2, err, queryID2 := store.ExecQueryWithContext(query, params)
	U.LogExecutionTimeWithQueryRequestID(startExecTime2, queryID2, &logFields)
	if err != nil {
		errString := fmt.Sprintf("failed to get last %d campaign meta for customads", days)
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
