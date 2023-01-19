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
	"spend":       "SUM(JSON_EXTRACT_STRING(value, 'spend') * inr_value)",
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
	limitString := ""
	if C.IsKPILimitIncreaseAllowedForProject(projectID) {
		limitString = fmt.Sprintf(" LIMIT %d", model.MaxResultsLimit)
	} else {
		limitString = fmt.Sprintf(" LIMIT %d", model.ResultsLimit)
	}
	if query.GroupByTimestamp == "" {
		sql, params, selectKeys, selectMetrics, errCode := store.GetSQLQueryAndParametersForBingAdsQueryV1(projectID,
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
			projectID, query, reqID, fetchSource, limitString, true, groupByCombinations)
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
	transformedQuery, customerAccountID, projectCurrency, err := store.transFormRequestFieldsAndFetchRequiredFieldsForBingads(
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
	dataCurrency := ""
	if(projectCurrency != ""){
		dataCurrency = store.GetDataCurrencyForBingAds(projectID)
	}
	if isSmartPropertyPresent {
		sql, params, selectKeys, selectMetrics, err = buildBingAdsQueryWithSmartProperty(transformedQuery, projectID, customerAccountID, fetchSource, limitString, isGroupByTimestamp, groupByCombinationsForGBT, dataCurrency, projectCurrency)
		if err != nil {
			return "", make([]interface{}, 0, 0), make([]string, 0, 0), make([]string, 0, 0), http.StatusInternalServerError
		}
		return sql, params, selectKeys, selectMetrics, http.StatusOK
	}
	sql, params, selectKeys, selectMetrics, err = buildBingAdsQueryV1(transformedQuery, projectID, customerAccountID, fetchSource, limitString, isGroupByTimestamp, groupByCombinationsForGBT, dataCurrency, projectCurrency)
	if err != nil {
		return "", make([]interface{}, 0, 0), make([]string, 0, 0), make([]string, 0, 0), http.StatusInternalServerError
	}
	return sql, params, selectKeys, selectMetrics, http.StatusOK
}

func (store *MemSQL) transFormRequestFieldsAndFetchRequiredFieldsForBingads(projectID int64,
	query model.ChannelQueryV1, reqID string) (*model.ChannelQueryV1, string, string, error) {
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
		return &model.ChannelQueryV1{}, "", "", err
	}
	customerAccountID := ftMapping.Accounts
	transformedQuery, err = convertFromRequestToBingAdsSpecificRepresentation(query)
	if err != nil {
		logCtx.Warn("Request failed in validation: ", err)
		return &model.ChannelQueryV1{}, "", "", err
	}
	projectSetting, _ := store.GetProjectSetting(projectID)
	return &transformedQuery, customerAccountID, projectSetting.ProjectCurrency, nil
}

func (store *MemSQL) GetDataCurrencyForBingAds(projectId int64) string{
	query := "select JSON_EXTRACT_STRING(value, 'currency_code')  from integration_documents where project_id = ? and document_type = 7 and source = 'bingads' limit 1"
	db := C.GetServices().Db
	

	params := make([]interface{},0)
	params = append(params, projectId)
	rows, err := db.Raw(query, params).Rows()
	if err != nil {
		log.WithError(err).Error("Failed to get currency code.")
	}
	defer rows.Close()

	var currency string
	for rows.Next() {
		if err := rows.Scan(&currency); err != nil {
			log.WithError(err).Error("Failed to get currency details for bingads")
		}
	}

	return currency
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
	limitString string, isGroupByTimestamp bool, groupByCombinationsForGBT map[string][]interface{}, dataCurrency string, projectCurrency string) (string, []interface{}, []string, []string, error) {
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
		fetchSource, limitString, isGroupByTimestamp, groupByCombinationsForGBT, customerAccountID, dataCurrency, projectCurrency)
	return sql, params, selectKeys, selectMetrics, nil
}

func buildBingAdsQueryWithSmartProperty(query *model.ChannelQueryV1, projectID int64, customerAccountID string, fetchSource bool,
	limitString string, isGroupByTimestamp bool, groupByCombinationsForGBT map[string][]interface{}, dataCurrency string, projectCurrency string) (string, []interface{}, []string, []string, error) {
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
		fetchSource, limitString, isGroupByTimestamp, groupByCombinationsForGBT, customerAccountID, dataCurrency, projectCurrency)
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

// Added case when statement for NULL value and empty value for group bys
func getSQLAndParamsFromBingAdsReports(query *model.ChannelQueryV1, projectID int64, from, to int64, docType int,
	fetchSource bool, limitString string, isGroupByTimestamp bool, groupByCombinationsForGBT map[string][]interface{}, customerAccountID string, dataCurrency string, projectCurrency string) (string, []interface{}, []string, []string) {
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
			value := fmt.Sprintf("'Bing Ads' as %s", model.BingAdsInternalRepresentationToExternalRepresentationForReports[key])
			selectKeys = append(selectKeys, value)
			responseSelectKeys = append(responseSelectKeys, model.BingAdsInternalRepresentationToExternalRepresentationForReports[key])
		} else {
			value := fmt.Sprintf("CASE WHEN %s IS NULL THEN '$none' WHEN %s = '' THEN '$none' ELSE %s END as %s", objectAndPropertyToValueInBingAdsReportsMapping[key], objectAndPropertyToValueInBingAdsReportsMapping[key], objectAndPropertyToValueInBingAdsReportsMapping[key], model.BingAdsInternalRepresentationToExternalRepresentationForReports[key])
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
	whereConditionForFilters, filterParams, err := getFilterPropertiesForBingReportsNew(query.Filters)
	if err != nil {
		return "", nil, nil, nil
	}
	if whereConditionForFilters != "" {
		whereConditionForFilters = " AND " + whereConditionForFilters
	}
	customerAccountIDs := strings.Split(customerAccountID, ",")
	finalParams := make([]interface{}, 0)
	if (dataCurrency != "" && projectCurrency != "") && U.ContainsStringInArray(query.SelectMetrics, "spend"){
		finalParams = append(finalParams, projectCurrency, dataCurrency)
	}
	staticWhereParams := []interface{}{projectID, model.BingAdsIntegration, docType, customerAccountIDs, from, to}
	finalParams = append(finalParams, staticWhereParams...)
	finalParams = append(finalParams, filterParams...)
	if len(groupByCombinationsForGBT) != 0 {
		whereConditionForGBT, whereParams := buildWhereConditionForGBTForBingAds(groupByCombinationsForGBT)
		whereConditionForFilters += " AND (" + whereConditionForGBT + ") "
		finalParams = append(finalParams, whereParams...)
	}

	resultSQLStatement := ""
	if (dataCurrency != "" && projectCurrency != "") && U.ContainsStringInArray(query.SelectMetrics, "spend"){
		resultSQLStatement = selectQuery + fromIntegrationDocuments + currencyQuery + staticWhereStatementForBingAds + whereConditionForFilters
	} else {
		selectQuery = strings.Replace(selectQuery, "* inr_value", "", -1)
		resultSQLStatement = selectQuery + fromIntegrationDocuments + staticWhereStatementForBingAds + whereConditionForFilters
	}
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
		fromStatement += "left join smart_properties ad_group on ad_group.project_id = integration_documents.project_id and ad_group.object_id = document_id "
	}
	if isPresentCampaignSmartProperty {
		fromStatement += "left join smart_properties campaign on campaign.project_id = integration_documents.project_id and campaign.object_id = document_id "
	}
	return fromStatement
}

// Added case when statement for NULL value and empty value for group bys
// Added case when statement for NULL value for smart properties. Didn't add for empty values as such case will not be present
func getSQLAndParamsFromBingAdsReportsWithSmartProperty(query *model.ChannelQueryV1, projectID int64, from, to int64, docType int,
	fetchSource bool, limitString string, isGroupByTimestamp bool, groupByCombinationsForGBT map[string][]interface{}, customerAccountID string, dataCurrency string, projectCurrency string) (string, []interface{}, []string, []string) {
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

				value := fmt.Sprintf("Case when JSON_EXTRACT_STRING(campaign.properties, '%s') is null then '$none' else JSON_EXTRACT_STRING(campaign.properties, '%s') END as campaign_%s", groupBy.Property, groupBy.Property, groupBy.Property)
				selectKeys = append(selectKeys, value)
				responseSelectKeys = append(responseSelectKeys, fmt.Sprintf("campaign_%s", groupBy.Property))

				groupByKeysWithoutTimestamp = append(groupByKeysWithoutTimestamp, fmt.Sprintf("campaign_%s", groupBy.Property))
			} else {
				value := fmt.Sprintf("Case when JSON_EXTRACT_STRING(ad_group.properties,'%s') is null then '$none' else JSON_EXTRACT_STRING(ad_group.properties,'%s') END as ad_group_%s", groupBy.Property, groupBy.Property, groupBy.Property)
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
				value := fmt.Sprintf("CASE WHEN %s IS NULL THEN '$none' WHEN %s = '' THEN '$none' ELSE %s END as %s", objectAndPropertyToValueInBingAdsReportsMapping[key], objectAndPropertyToValueInBingAdsReportsMapping[key], objectAndPropertyToValueInBingAdsReportsMapping[key], model.BingAdsInternalRepresentationToExternalRepresentationForReports[key])
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
	whereConditionForFilters, filterParams, err := getFilterPropertiesForBingReportsNew(query.Filters)
	if err != nil {
		return "", nil, nil, nil
	}
	filterStatementForSmartPropertyGroupBy := getNotNullFilterStatementForSmartPropertyGroupBys(query.GroupBy)
	finalFilterStatement := joinWithWordInBetween("AND", staticWhereStatementForBingAdsWithSmartProperty, whereConditionForFilters, filterStatementForSmartPropertyGroupBy)
	customerAccountIDs := strings.Split(customerAccountID, ",")
	finalParams := make([]interface{}, 0)
	if (dataCurrency != "" && projectCurrency != "") && U.ContainsStringInArray(query.SelectMetrics, "spend"){
		finalParams = append(finalParams, projectCurrency, dataCurrency)
	}
	staticWhereParams := []interface{}{projectID, model.BingAdsIntegration, customerAccountIDs, docType, from, to}
	finalParams = append(finalParams, staticWhereParams...)
	finalParams = append(finalParams, filterParams...)
	if len(groupByCombinationsForGBT) != 0 {
		whereConditionForGBT, whereParams := buildWhereConditionForGBTForBingAds(groupByCombinationsForGBT)
		whereConditionForFilters += " AND (" + whereConditionForGBT + ") "
		finalParams = append(finalParams, whereParams...)
	}

	fromStatement := getBingAdsFromStatementWithJoins(query.Filters, query.GroupBy)
	resultSQLStatement := ""
	if (dataCurrency != "" && projectCurrency != "") && U.ContainsStringInArray(query.SelectMetrics, "spend"){
		resultSQLStatement = selectQuery + fromStatement + currencyQuery +  finalFilterStatement
	} else {
		selectQuery = strings.Replace(selectQuery, "* inr_value", "", -1)
		resultSQLStatement = selectQuery + fromStatement + finalFilterStatement
	}
	if len(groupByStatement) != 0 {
		resultSQLStatement += "GROUP BY " + groupByStatement
	}
	resultSQLStatement += " " + orderByQuery + limitString + ";"
	return resultSQLStatement, finalParams, responseSelectKeys, responseSelectMetrics
}

func getFilterPropertiesForBingReportsNew(filters []model.ChannelFilterV1) (rStmnt string, rParams []interface{}, err error) {
	logFields := log.Fields{
		"filters": filters,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	campaignFilter := ""
	adGroupFilter := ""
	filtersLen := len(filters)
	if filtersLen == 0 {
		return rStmnt, rParams, nil
	}

	rParams = make([]interface{}, 0)
	groupedProperties := model.GetChannelFiltersGrouped(filters)

	for indexOfGroup, currentGroupedProperties := range groupedProperties {
		var currentGroupStmnt, pStmnt string
		for indexOfProperty, p := range currentGroupedProperties {

			if p.LogicalOp == "" {
				p.LogicalOp = "AND"
			}

			if !isValidLogicalOp(p.LogicalOp) {
				return rStmnt, rParams, errors.New("invalid logical op on where condition")
			}
			pStmnt = ""
			propertyOp := getOp(p.Condition, "categorical")
			// categorical property type.
			pValue := ""
			if p.Condition == model.ContainsOpStr || p.Condition == model.NotContainsOpStr {
				pValue = fmt.Sprintf("%s", p.Value)
			} else {
				pValue = p.Value
			}
			_, isPresent := model.SmartPropertyReservedNames[p.Property]
			if isPresent {
				key := fmt.Sprintf("%s.%s", p.Object, p.Property)
				pFilter := objectAndPropertyToValueInBingAdsReportsMapping[key]

				if p.Value != model.PropertyValueNone {
					pStmnt = fmt.Sprintf("%s %s '%s' ", pFilter, propertyOp, pValue)
				} else {
					// where condition for $none value.
					if propertyOp == model.EqualsOp || propertyOp == model.RLikeOp {
						pStmnt = fmt.Sprintf("(%s IS NULL OR %s = '')", pFilter, pFilter)
					} else if propertyOp == model.NotEqualOp || propertyOp == model.NotRLikeOp {
						pStmnt = fmt.Sprintf("(%s IS NOT NULL OR %s != '')", pFilter, pFilter)
					} else {
						return "", nil, fmt.Errorf("unsupported opertator %s for property value none", propertyOp)
					}
				}
			} else {
				if p.Value != model.PropertyValueNone {
					pStmnt = fmt.Sprintf("JSON_EXTRACT_STRING(%s.properties, '%s') %s ?", model.BingAdsObjectMapForSmartProperty[p.Object], p.Property, propertyOp)
					rParams = append(rParams, pValue)
				} else {
					if propertyOp == model.EqualsOp || propertyOp == model.RLikeOp {
						pStmnt = fmt.Sprintf("(JSON_EXTRACT_STRING(%s.properties, '%s') IS NULL OR JSON_EXTRACT_STRING(%s.properties, '%s') = '')", model.BingAdsObjectMapForSmartProperty[p.Object], p.Property, model.BingAdsObjectMapForSmartProperty[p.Object], p.Property)
					} else if propertyOp == model.NotEqualOp || propertyOp == model.NotRLikeOp {
						pStmnt = fmt.Sprintf("(JSON_EXTRACT_STRING(%s.properties, '%s') IS NOT NULL OR JSON_EXTRACT_STRING(%s.properties, '%s') != '')", model.BingAdsObjectMapForSmartProperty[p.Object], p.Property, model.BingAdsObjectMapForSmartProperty[p.Object], p.Property)
					} else {
						return "", nil, fmt.Errorf("unsupported opertator %s for property value none", propertyOp)
					}
				}
				if p.Object == bingadsCampaign {
					campaignFilter = smartPropertyCampaignStaticFilter
				} else {
					adGroupFilter = smartPropertyAdGroupStaticFilter
				}
			}
			if indexOfProperty == 0 {
				currentGroupStmnt = pStmnt
			} else {
				currentGroupStmnt = fmt.Sprintf("%s %s %s", currentGroupStmnt, p.LogicalOp, pStmnt)
			}
		}
		if indexOfGroup == 0 {
			rStmnt = fmt.Sprintf("(%s)", currentGroupStmnt)
		} else {
			rStmnt = fmt.Sprintf("%s AND (%s)", rStmnt, currentGroupStmnt)
		}

	}
	if campaignFilter != "" {
		rStmnt += (" AND " + campaignFilter)
	}
	if adGroupFilter != "" {
		rStmnt += (" AND " + adGroupFilter)
	}
	return rStmnt, rParams, nil
}

func buildWhereConditionForGBTForBingAds(groupByCombinations map[string][]interface{}) (string, []interface{}) {
	logFields := log.Fields{
		"group_by_combinations": groupByCombinations,
	}

	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
	resultantWhereCondition := ""
	resultantInClauses := make([]string, 0)
	params := make([]interface{}, 0)

	for dimension, values := range groupByCombinations {
		currentInClause := ""

		jsonExtractExpression := GetFilterObjectExpressionForChannelBingAds(dimension)

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

// request has dimension - campaign_name.
// response has string with JSON_EXTRACT(adwords_documents.value, 'campaign_name')
func GetFilterObjectExpressionForChannelBingAds(dimension string) string {
	filterObjectForSmartPropertiesCampaign := "campaign.properties"
	filterObjectForSmartPropertiesAdGroup := "ad_group.properties"

	filterExpression := ""
	isNotSmartProperty := false
	if strings.HasPrefix(dimension, model.CampaignPrefix) {
		filterExpression, isNotSmartProperty = GetFilterExpressionIfPresentForBingAds("campaigns", dimension, model.CampaignPrefix)
		if !isNotSmartProperty {
			filterExpression = fmt.Sprintf("JSON_EXTRACT_STRING(%s,'%s')", filterObjectForSmartPropertiesCampaign, strings.TrimPrefix(dimension, model.CampaignPrefix))
		}
	} else if strings.HasPrefix(dimension, model.AdgroupPrefix) {
		filterExpression, isNotSmartProperty = GetFilterExpressionIfPresentForBingAds("ad_groups", dimension, model.AdgroupPrefix)
		if !isNotSmartProperty {
			filterExpression = fmt.Sprintf("JSON_EXTRACT_STRING(%s,'%s')", filterObjectForSmartPropertiesAdGroup, strings.TrimPrefix(dimension, model.AdgroupPrefix))
		}
	} else {
		filterExpression, _ = GetFilterExpressionIfPresentForBingAds("keyword", dimension, model.KeywordPrefix)
	}
	return filterExpression
}

// Input: objectType - campaign, dimension - , prefix - . TODO
func GetFilterExpressionIfPresentForBingAds(objectType, dimension, prefix string) (string, bool) {
	key := fmt.Sprintf(`%s.%s`, objectType, strings.TrimPrefix(dimension, prefix))
	reportProperty, isPresent := objectAndPropertyToValueInBingAdsReportsMapping[key]
	return reportProperty, isPresent
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
//
//	or integration_documents.value["ad_group_id"] = smart_properties.object_id (when smart_properties.object_type = 2)
//
// [make sure there aren't multiple smart_properties rows for a particular object,
// or weekly insights for bing would show incorrect data.]
func (store *MemSQL) PullBingAdsRowsV2(projectID int64, startTime, endTime int64) (*sql.Rows, *sql.Tx, error) {
	logFields := log.Fields{
		"project_id": projectID,
		"start_time": startTime,
		"end_time":   endTime,
	}
	defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)

	rawQuery := fmt.Sprintf("SELECT bing.document_id, bing.value, bing.timestamp, bing.document_type, sp.properties FROM integration_documents bing "+
		"LEFT JOIN smart_properties sp ON sp.project_id = %d AND sp.source = '%s' AND "+
		"((COALESCE(sp.object_type,1) = 1 AND (sp.object_id = JSON_EXTRACT_STRING(bing.value, 'campaign_id') OR sp.object_id = JSON_EXTRACT_STRING(bing.value, 'base_campaign_id'))) OR "+
		"(COALESCE(sp.object_type,2) = 2 AND (sp.object_id = JSON_EXTRACT_STRING(bing.value, 'ad_group_id') OR sp.object_id = JSON_EXTRACT_STRING(bing.value, 'base_ad_group_id')))) "+
		"WHERE bing.project_id = %d AND UNIX_TIMESTAMP(bing.created_at) BETWEEN %d AND %d "+
		"LIMIT %d",
		projectID, model.ChannelBingAds, projectID, startTime, endTime, model.BingPullLimit+1)

	rows, tx, err, _ := store.ExecQueryWithContext(rawQuery, []interface{}{})
	return rows, tx, err
}

// PullBingAdsRows - Function to pull all bing integration documents
// Selecting VALUE, TIMESTAMP, TYPE from integration_documents and PROPERTIES, OBJECT_TYPE from smart_properties
// Left join smart_properties filtered by project_id and source=bingads
// where integration_documents.value["campaign_id"] = smart_properties.object_id (when smart_properties.object_type = 1)
//
//	or integration_documents.value["ad_group_id"] = smart_properties.object_id (when smart_properties.object_type = 2)
//
// [make sure there aren't multiple smart_properties rows for a particular object,
// or weekly insights for bing would show incorrect data.]
func (store *MemSQL) PullBingAdsRowsV1(projectID int64, startTime, endTime int64) (*sql.Rows, *sql.Tx, error) {
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

	rawQuery := fmt.Sprintf("SELECT bing.document_id, bing.value, bing.timestamp, bing.document_type, sp.properties FROM integration_documents bing "+
		"LEFT JOIN smart_properties sp ON sp.project_id = %d AND sp.source = '%s' AND "+
		"((COALESCE(sp.object_type,1) = 1 AND (sp.object_id = JSON_EXTRACT_STRING(bing.value, 'campaign_id') OR sp.object_id = JSON_EXTRACT_STRING(bing.value, 'base_campaign_id'))) OR "+
		"(COALESCE(sp.object_type,2) = 2 AND (sp.object_id = JSON_EXTRACT_STRING(bing.value, 'ad_group_id') OR sp.object_id = JSON_EXTRACT_STRING(bing.value, 'base_ad_group_id')))) "+
		"WHERE bing.project_id = %d AND bing.timestamp BETWEEN %d AND %d "+
		"ORDER BY bing.document_type, bing.timestamp LIMIT %d",
		projectID, model.ChannelBingAds, projectID, start, end, model.BingPullLimit+1)

	rows, tx, err, _ := store.ExecQueryWithContext(rawQuery, []interface{}{})
	return rows, tx, err
}
