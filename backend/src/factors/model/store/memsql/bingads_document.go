package memsql

import (
	"errors"
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"fmt"
	"net/http"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	bingadsCampaign                = "campaigns"
	bingadsAdGroup                 = "ad_groups"
	bingadsKeyword                 = "keyword"
	fromIntegrationDocuments       = " FROM integration_documents "
	staticWhereStatementForBingAds = "WHERE project_id = ? AND source= ? AND document_type = ? AND customer_account_id IN (?) AND timestamp between ? AND ? "
	bingadsFilterQueryStr          = "SELECT DISTINCT(LCASE(JSON_EXTRACT_STRING(value, ?))) as filter_value FROM integration_documents WHERE project_id = ? AND customer_account_id IN (?) AND document_type = ? AND JSON_EXTRACT_STRING(value, ?) IS NOT NULL LIMIT 5000"
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

func (store *MemSQL) GetBingadsFilterValues(projectID uint64, requestFilterObject string, requestFilterProperty string, reqID string) ([]interface{}, int) {
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
	requestFilterProperty = strings.TrimPrefix(requestFilterProperty, fmt.Sprintf("%v_", requestFilterObject))
	docType := model.BingadsDocumentTypeAlias[model.BingAdsObjectInternalRepresentationToExternalRepresentation[requestFilterObject]]
	filterProperty := model.BingAdsInternalRepresentationToExternalRepresentation[fmt.Sprintf("%v.%v", model.BingAdsObjectInternalRepresentationToExternalRepresentation[requestFilterObject], requestFilterProperty)]
	params := []interface{}{filterProperty, projectID, customerAccountIDs, docType, filterProperty}
	_, resultRows, err := store.ExecuteSQL(bingadsFilterQueryStr, params, logCtx)
	if err != nil {
		logCtx.WithError(err).WithField("query", bingadsFilterQueryStr).WithField("params", params).Error(model.BingSpecificError)
		return make([]interface{}, 0), http.StatusInternalServerError
	}
	return Convert2DArrayTo1DArray(resultRows), http.StatusFound
}

func (store *MemSQL) GetBingadsFilterValuesSQLAndParams(projectID uint64, requestFilterObject string, requestFilterProperty string, reqID string) (string, []interface{}, int) {
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
	filterProperty := model.BingAdsInternalRepresentationToExternalRepresentation[fmt.Sprintf("%v.%v", model.BingAdsObjectInternalRepresentationToExternalRepresentation[requestFilterObject], requestFilterProperty)]
	params := []interface{}{filterProperty, projectID, customerAccountIDs, docType, filterProperty}
	return bingadsFilterQueryStr, params, http.StatusFound
}

func (store *MemSQL) buildBingAdsChannelConfig(projectID uint64) *model.ChannelConfigResult {
	bingAdsObjectsAndProperties := store.buildObjectAndPropertiesForBingAds(projectID, model.ObjectsForBingads)
	// selectable metrics for bing ads to be added ?
	selectMetrics := append(SelectableMetricsForAllChannels)
	objectsAndProperties := bingAdsObjectsAndProperties

	return &model.ChannelConfigResult{
		SelectMetrics:        selectMetrics,
		ObjectsAndProperties: objectsAndProperties,
	}
}
func (store *MemSQL) buildObjectAndPropertiesForBingAds(projectID uint64, objects []string) []model.ChannelObjectAndProperties {
	objectsAndProperties := make([]model.ChannelObjectAndProperties, 0)
	for _, currentObject := range objects {
		// to do: check if normal properties present then only smart properties will be there
		propertiesAndRelated, isPresent := model.MapOfBingAdsObjectsToPropertiesAndRelated[currentObject]
		var currentProperties []model.ChannelProperty
		var currentPropertiesSmart []model.ChannelProperty
		if isPresent {
			currentProperties = buildProperties(propertiesAndRelated)
			//	smartProperty := GetSmartPropertyAndRelated(projectID, currentObject, "bing_ads")
			//	currentPropertiesSmart = buildProperties(smartProperty)
			currentProperties = append(currentProperties, currentPropertiesSmart...)
		} else {
			currentProperties = buildProperties(allChannelsPropertyToRelated)
			//	smartProperty := GetSmartPropertyAndRelated(projectID, currentObject, "bing_ads")
			//	currentPropertiesSmart = buildProperties(smartProperty)
			currentProperties = append(currentProperties, currentPropertiesSmart...)
		}
		objectsAndProperties = append(objectsAndProperties, buildObjectsAndProperties(currentProperties, []string{currentObject})...)
	}
	return objectsAndProperties
}

// modify
func (store *MemSQL) ExecuteBingAdsChannelQueryV1(projectID uint64, query *model.ChannelQueryV1, reqID string) ([]string, [][]interface{}, int) {
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
func (store *MemSQL) GetSQLQueryAndParametersForBingAdsQueryV1(projectID uint64, query *model.ChannelQueryV1, reqID string, fetchSource bool,
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
	sql, params, selectKeys, selectMetrics, err = buildBingAdsQueryV1(transformedQuery, projectID, customerAccountID, fetchSource, limitString, isGroupByTimestamp, groupByCombinationsForGBT)
	if err != nil {
		return "", make([]interface{}, 0, 0), make([]string, 0, 0), make([]string, 0, 0), http.StatusInternalServerError
	}
	return sql, params, selectKeys, selectMetrics, http.StatusOK
}

func (store *MemSQL) transFormRequestFieldsAndFetchRequiredFieldsForBingads(projectID uint64,
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

func buildBingAdsQueryV1(query *model.ChannelQueryV1, projectID uint64, customerAccountID string, fetchSource bool,
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
func getSQLAndParamsFromBingAdsReports(query *model.ChannelQueryV1, projectID uint64, from, to int64, docType int,
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

func buildWhereConditionForGBTForBingAds(groupByCombinations []map[string]interface{}) (string, []interface{}) {
	logFields := log.Fields{
		"group_by_combinations": groupByCombinations,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	whereConditionForGBT := ""
	params := make([]interface{}, 0)
	for _, groupByCombination := range groupByCombinations {
		whereConditionForEachCombination := ""
		for dimension, value := range groupByCombination {
			filterString := ""
			if strings.HasPrefix(dimension, model.CampaignPrefix) {
				key := fmt.Sprintf(`%s.%s`, "campaigns", strings.TrimPrefix(dimension, model.CampaignPrefix))
				currentFilterKey, _ := objectAndPropertyToValueInBingAdsReportsMapping[key]
				filterString = currentFilterKey

			} else if strings.HasPrefix(dimension, model.AdgroupPrefix) {
				key := fmt.Sprintf(`%s.%s`, "ad_groups", strings.TrimPrefix(dimension, model.AdgroupPrefix))
				currentFilterKey, _ := objectAndPropertyToValueInBingAdsReportsMapping[key]
				filterString = currentFilterKey
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
