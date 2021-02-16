package postgres

import (
	"errors"
	C "factors/config"
	"factors/model/model"
	U "factors/util"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

const (
	linkedinCampaignGroup = "campaign_group"
	linkedinCampaign      = "campaign"
	linkedinCreative      = "creative"
	linkedinStringColumn  = "linkedin"
)

var linkedinDocumentTypeAlias = map[string]int{
	"creative":                1,
	"campaign_group":          2,
	"campaign":                3,
	"creative_insights":       4,
	"campaign_group_insights": 5,
	"campaign_insights":       6,
	"ad_account":              7,
}

var objectAndPropertyToValueInLinkedinReportsMapping = map[string]string{
	"campaign_group:id":   "campaign_group_id::bigint",
	"creative:id":         "creative_id::bigint",
	"campaign:id":         "campaign_id::bigint",
	"campaign_group:name": "value->>'campaign_group_name'",
	"campaign:name":       "value->>'campaign_name'",
}

var objectToValueInLinkedinJobsMapping = map[string]string{
	"campaign_group:name": "campaign_group_name",
	"campaign:name":       "campaign_group_name",
	"campaign_group:id":   "campaign_group_id",
	"campaign:id":         "campaign_id",
	"creative:id":         "creative_id",
}

// TODO check
var linkedinMetricsToAggregatesInReportsMapping = map[string]string{
	"impressions": "SUM((value->>'impressions')::float)",
	"clicks":      "SUM((value->>'clicks')::float)",
	"spend":       "SUM((value->>'costInLocalCurrency')::float)",
	"conversions": "SUM((value->>'conversionValueInLocalCurrency')::float)",
	// "cost_per_click": "average_cost",
	// "conversion_rate": "conversion_rate"
}
var linkedinExternalRepresentationToInternalRepresentation = map[string]string{
	"name":        "name",
	"id":          "id",
	"impressions": "impressions",
	"clicks":      "clicks",
	"spend":       "spend",
	"conversion":  "conversionValueInLocalCurrency",
	"campaign":    "campaign_group",
	"ad_group":    "campaign",
	"ad":          "creative",
}

var linkedinInternalRepresentationToExternalRepresentation = map[string]string{
	"impressions":         "impressions",
	"clicks":              "clicks",
	"spend":               "spend",
	"conversions":         "conversion",
	"campaign_group:name": "campaign_name",
	"campaign:name":       "ad_group_name",
	"campaign_group:id":   "campaign_id",
	"campaign:id":         "ad_group_id",
	"creative:id":         "ad_id",
}
var linkedinInternalGroupByRepresentation = map[string]string{
	"impressions":         "impressions",
	"clicks":              "clicks",
	"spend":               "spend",
	"conversions":         "conversion",
	"campaign_group:name": "campaign_name",
	"campaign:name":       "ad_group_name",
	"campaign_group:id":   "campaign_group_id",
	"campaign:id":         "campaign_id",
	"creative:id":         "creative_id",
}
var objectToValueInLinkedinFiltersMapping = map[string]string{
	"campaign:name":       "value->>'campaign_name'",
	"campaign_group:name": "value->>'campaign_group_name'",
	"campaign:id":         "campaign_id",
	"campaign_group:id":   "campaign_group_id",
	"creative:id":         "creative_id",
}
var linkedinMetricsToOperation = map[string]string{
	"impressions": "sum",
	"clicks":      "sum",
	"spend":       "sum",
	"conversions": "sum",
}

var mapOfTypeToLinkedinJobCTEAlias = map[string]string{
	"campaign":       "campaign_cte",
	"campaign_group": "campaign_group_cte",
}

const errorDuplicateLinkedinDocument = "pq: duplicate key value violates unique constraint \"linkedin_documents_pkey\""

var errorEmptyLinkedinDocument = errors.New("empty linked document")

const linkedinFilterQueryStr = "SELECT DISTINCT(value->>?) as filter_value FROM linkedin_documents WHERE project_id = ? AND" +
	" " + "customer_ad_account_id = ? AND type = ? AND value->>? IS NOT NULL LIMIT 5000"

const fromLinkedinDocument = " FROM linkedin_documents "

const staticWhereStatementForLinkedin = "WHERE project_id = ? AND customer_ad_account_id IN ( ? ) AND type = ? AND timestamp between ? AND ? "

func getLinkedinDocumentTypeAliasByType() map[int]string {
	documentTypeMap := make(map[int]string, 0)
	for alias, typ := range linkedinDocumentTypeAlias {
		documentTypeMap[typ] = alias
	}

	return documentTypeMap
}

func (pg *Postgres) GetLinkedinLastSyncInfo(projectID uint64, CustomerAdAccountID string) ([]model.LinkedinLastSyncInfo, int) {
	db := C.GetServices().Db

	linkedinLastSyncInfos := make([]model.LinkedinLastSyncInfo, 0, 0)

	queryStr := "SELECT project_id, customer_ad_account_id, type as document_type, max(timestamp) as last_timestamp" +
		" FROM linkedin_documents WHERE project_id = ? AND customer_ad_account_id = ?" +
		" GROUP BY project_id, customer_ad_account_id, type "

	rows, err := db.Raw(queryStr, projectID, CustomerAdAccountID).Rows()
	if err != nil {
		log.WithError(err).Error("Failed to get last linkedin documents by type for sync info.")
		return linkedinLastSyncInfos, http.StatusInternalServerError
	}
	defer rows.Close()

	for rows.Next() {
		var linkedinLastSyncInfo model.LinkedinLastSyncInfo
		if err := db.ScanRows(rows, &linkedinLastSyncInfo); err != nil {
			log.WithError(err).Error("Failed to scan last linkedin documents by type for sync info.")
			return []model.LinkedinLastSyncInfo{}, http.StatusInternalServerError
		}

		linkedinLastSyncInfos = append(linkedinLastSyncInfos, linkedinLastSyncInfo)
	}
	documentTypeAliasByType := getLinkedinDocumentTypeAliasByType()

	for i := range linkedinLastSyncInfos {
		logCtx := log.WithField("project_id", linkedinLastSyncInfos[i].ProjectID)
		typeAlias, typeAliasExists := documentTypeAliasByType[linkedinLastSyncInfos[i].DocumentType]
		if !typeAliasExists {
			logCtx.WithField("document_type",
				linkedinLastSyncInfos[i].DocumentType).Error("Invalid document type given. No type alias name.")
			continue
		}

		linkedinLastSyncInfos[i].DocumentTypeAlias = typeAlias
	}
	return linkedinLastSyncInfos, http.StatusOK
}

func isDuplicateLinkedinDocumentError(err error) bool {
	return err.Error() == errorDuplicateLinkedinDocument
}

// CreatelinkedinDocument ...
func (pg *Postgres) CreateLinkedinDocument(projectID uint64, document *model.LinkedinDocument) int {
	logCtx := log.WithField("customer_acc_id", document.CustomerAdAccountID).WithField(
		"project_id", document.ProjectID)

	if document.CustomerAdAccountID == "" || document.TypeAlias == "" {
		logCtx.Error("Invalid linkedin document.")
		return http.StatusBadRequest
	}
	if document.ProjectID == 0 || document.Timestamp == 0 {
		logCtx.Error("Invalid linkedin document.")
		return http.StatusBadRequest
	}

	logCtx = logCtx.WithField("type_alias", document.TypeAlias)
	docType, docTypeExists := linkedinDocumentTypeAlias[document.TypeAlias]
	if !docTypeExists {
		logCtx.Error("Invalid type alias.")
		return http.StatusBadRequest
	}
	document.Type = docType

	campaignGroupID, campaignID, creativeID, error := getLinkedinHierarchyColumnsByType(docType, document.Value)
	if error != nil {
		logCtx.Error("Invalid docType alias.")
		return http.StatusBadRequest
	}
	document.CampaignGroupID = campaignGroupID
	document.CampaignID = campaignID
	document.CreativeID = creativeID

	db := C.GetServices().Db
	err := db.Create(&document).Error
	if err != nil {
		if isDuplicateLinkedinDocumentError(err) {
			logCtx.WithError(err).WithField("id", document.ID).Error(
				"Failed to create an linkedin doc. Duplicate.")
			return http.StatusConflict
		}
		logCtx.WithError(err).WithField("id", document.ID).Error(
			"Failed to create an linkedin doc. Continued inserting other docs.")
		return http.StatusInternalServerError
	}

	return http.StatusCreated
}

func getLinkedinHierarchyColumnsByType(docType int, valueJSON *postgres.Jsonb) (string, string, string, error) {
	if docType > len(linkedinDocumentTypeAlias) {
		return "", "", "", errors.New("invalid document type")
	}

	valueMap, err := U.DecodePostgresJsonb(valueJSON)
	if err != nil {
		return "", "", "", err
	}

	if len(*valueMap) == 0 {
		return "", "", "", errorEmptyLinkedinDocument
	}

	return U.GetStringFromMapOfInterface(*valueMap, "campaign_group_id", ""), U.GetStringFromMapOfInterface(*valueMap, "campaign_id", ""), U.GetStringFromMapOfInterface(*valueMap, "creative_id", ""), nil
}

func (pg *Postgres) ExecuteLinkedinChannelQuery(projectID uint64, query *model.ChannelQuery) (*model.ChannelQueryResult, int) {
	logCtx := log.WithField("project_id", projectID).WithField("query", query)

	if projectID == 0 || query == nil {
		logCtx.Error("Invalid project_id or query on execute linkedin channel query.")
		return nil, http.StatusInternalServerError
	}

	projectSetting, errCode := pg.GetProjectSetting(projectID)
	if errCode != http.StatusFound {
		return nil, http.StatusInternalServerError
	}

	if projectSetting.IntLinkedinAdAccount == "" {
		logCtx.Error("Execute linkedin channel query failed. No customer account id.")
		return nil, http.StatusInternalServerError
	}
	query.From = ChangeUnixTimestampToDate(query.From)
	query.To = ChangeUnixTimestampToDate(query.To)
	queryResult := &model.ChannelQueryResult{}
	result, err := getLinkedinChannelResult(projectID, projectSetting.IntLinkedinAdAccount, query)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get linked query result.")
		return queryResult, http.StatusInternalServerError
	}
	queryResult = result
	if query.Breakdown == "" {
		return queryResult, http.StatusOK
	}

	metricBreakDown, err := pg.getLinkedinMetricBreakdown(projectID, projectSetting.IntLinkedinAdAccount, query)
	queryResult.MetricsBreakdown = metricBreakDown

	impressionsIndex := 0
	for _, key := range queryResult.MetricsBreakdown.Headers {
		if key == "impressions" {
			sort.Slice(queryResult.MetricsBreakdown.Rows, func(i, j int) bool {
				return queryResult.MetricsBreakdown.Rows[i][impressionsIndex].(float64) > queryResult.MetricsBreakdown.Rows[j][impressionsIndex].(float64)
			})
			break
		}
		impressionsIndex++
	}
	return queryResult, http.StatusOK
}
func getLinkedinChannelResult(projectID uint64, customerAccountID string, query *model.ChannelQuery) (*model.ChannelQueryResult, error) {

	logCtx := log.WithField("project_id", projectID)

	sqlQuery, documentType := getLinkedinMetricsQuery(query, false)

	queryResult := &model.ChannelQueryResult{}
	db := C.GetServices().Db
	rows, err := db.Raw(sqlQuery, projectID, customerAccountID,
		query.From,
		query.To,
		documentType).Rows()
	if err != nil {
		logCtx.WithError(err).Error("Failed to build channel query result.")
		return queryResult, err
	}
	resultHeaders, resultRows, err := U.DBReadRows(rows)
	if err != nil {
		return nil, err
	}
	if len(resultRows) == 0 {
		log.Error("Aggregate query returned zero rows.")
		return nil, errors.New("no rows returned")
	}

	if len(resultRows) > 1 {
		log.Error("Aggregate query returned more than one row on get adwords metric kvs.")
	}

	metricKvs := make(map[string]interface{})
	for i, k := range resultHeaders {
		metricKvs[k] = resultRows[0][i]
	}

	queryResult.Metrics = &metricKvs
	return queryResult, nil
}
func (pg *Postgres) getLinkedinMetricBreakdown(projectID uint64, customerAccountID string, query *model.ChannelQuery) (*model.ChannelBreakdownResult, error) {
	logCtx := log.WithField("project_id", projectID).WithField("customer_account_id", customerAccountID)

	sqlQuery, documentType := getLinkedinMetricsQuery(query, true)

	db := C.GetServices().Db
	rows, err := db.Raw(sqlQuery, projectID, customerAccountID,
		query.From,
		query.To,
		documentType).Rows()
	if err != nil {
		logCtx.WithError(err).Error("Failed to build channel query result.")
		return nil, err
	}

	resultHeaders, resultRows, err := U.DBReadRows(rows)
	if err != nil {
		return nil, err
	}
	for i := range resultHeaders {
		if resultHeaders[i] == CAChannelGroupKey {
			resultHeaders[i] = query.Breakdown
		}
	}

	for ri := range resultRows {
		for ci := range resultRows[ri] {
			if ci > 0 && resultRows[ri][ci] == nil {
				resultRows[ri][ci] = 0
			}
		}
	}

	return &model.ChannelBreakdownResult{Headers: resultHeaders, Rows: resultRows}, nil
}
func getLinkedinDocumentType(query *model.ChannelQuery) int {
	var documentType int
	if query.FilterKey == "campaign_group" {
		documentType = 5
	}
	if query.FilterKey == "campaign" {
		documentType = 6
	}
	if query.FilterKey == "creative" {
		documentType = 4
	}
	return documentType
}
func getLinkedinMetricsQuery(query *model.ChannelQuery, withBreakdown bool) (string, int) {

	documentType := getLinkedinDocumentType(query)

	selectColstWithoutAlias := "SUM((value->>'impressions')::float) as %s , SUM((value->>'clicks')::float) as %s," +
		" " + "SUM((value->>'costInUsd')::float) as %s," +
		" " + "SUM((value->>'approximateUniqueImpressions')::float) as %s," +
		" " + "SUM((value->>'likes')::float) as %s, SUM((value->>'follows')::float) as %s, " +
		" " + "SUM((value->>'totalEngagements')::float) as %s," +
		" " + "AVG((value->>'conversionValueInLocalCurrency')::float) as %s"

	selectCols := fmt.Sprintf(selectColstWithoutAlias, CAColumnImpressions, CAColumnClicks,
		CAColumnTotalCost, CAColumnUniqueImpressions, CAColumnLikes,
		CAColumnFollows, CAColumnTotalEngagement, CAColumnConversionValueInLocalCurrency)

	strmntWhere := "WHERE project_id= ? AND customer_ad_account_id = ? AND timestamp>? AND timestamp<? AND type=?"

	strmntGroupBy := ""
	if withBreakdown {
		selectCols = "id, " + selectCols
		strmntGroupBy = "GROUP BY id"
	}

	sqlQuery := "SELECT" + " " + selectCols + " " + "FROM linkedin_documents" + " " + strmntWhere + " " + strmntGroupBy
	return sqlQuery, documentType
}

// v1 Api
func buildLinkedinChannelConfig() *model.ChannelConfigResult {
	properties := buildProperties(allChannelsPropertyToRelated)
	objectsAndProperties := buildObjectsAndProperties(properties, objectsForAllChannels)

	return &model.ChannelConfigResult{
		SelectMetrics:        selectableMetricsForAllChannels,
		ObjectsAndProperties: objectsAndProperties,
	}
}
func (pg *Postgres) GetLinkedinFilterValues(projectID uint64, requestFilterObject string, requestFilterProperty string, reqID string) ([]interface{}, int) {
	linkedinInternalFilterProperty, docType, err := getFilterRelatedInformationForLinkedin(requestFilterObject, requestFilterProperty)
	if err != http.StatusOK {
		return make([]interface{}, 0, 0), http.StatusBadRequest
	}
	filterValues, errCode := pg.getLinkedinFilterValuesByType(projectID, docType, linkedinInternalFilterProperty, reqID)
	if errCode != http.StatusFound {
		return []interface{}{}, http.StatusInternalServerError
	}

	return filterValues, http.StatusFound
}

func getFilterRelatedInformationForLinkedin(requestFilterObject string, requestFilterProperty string) (string, int, int) {
	linkedinInternalFilterObject, isPresent := linkedinExternalRepresentationToInternalRepresentation[requestFilterObject]
	if !isPresent {
		log.Error("Invalid linkedin filter object.")
		return "", 0, http.StatusBadRequest
	}
	linkedinInternalFilterProperty, isPresent := linkedinExternalRepresentationToInternalRepresentation[requestFilterProperty]
	if !isPresent {
		log.Error("Invalid linkedin filter property.")
		return "", 0, http.StatusBadRequest
	}
	docType := linkedinDocumentTypeAlias[linkedinInternalFilterObject]

	return linkedinInternalFilterProperty, docType, http.StatusOK
}

func (pg *Postgres) getLinkedinFilterValuesByType(projectID uint64, docType int, property string, reqID string) ([]interface{}, int) {
	logCtx := log.WithField("req_id", reqID)
	projectSetting, errCode := pg.GetProjectSetting(projectID)
	if errCode != http.StatusFound {
		logCtx.Error("Failed to fetch Project Setting in linkedin filter values.")
		return []interface{}{}, http.StatusInternalServerError
	}
	customerAccountID := projectSetting.IntLinkedinAdAccount

	logCtx = log.WithField("project_id", projectID).WithField("doc_type", docType).WithField("req_id", reqID)
	params := []interface{}{property, projectID, customerAccountID, docType, property}
	_, resultRows, _ := pg.ExecuteSQL(linkedinFilterQueryStr, params, logCtx)

	return Convert2DArrayTo1DArray(resultRows), http.StatusFound
}
func (pg *Postgres) GetLinkedinSQLQueryAndParametersForFilterValues(projectID uint64, requestFilterObject string, requestFilterProperty string, reqID string) (string, []interface{}, int) {
	logCtx := log.WithField("project_id", projectID).WithField("req_id", reqID)
	linkedinInternalFilterProperty, docType, err := getFilterRelatedInformationForLinkedin(requestFilterObject, requestFilterProperty)
	if err != http.StatusOK {
		return "", make([]interface{}, 0, 0), http.StatusBadRequest
	}
	projectSetting, errCode := pg.GetProjectSetting(projectID)
	if errCode != http.StatusFound {
		logCtx.Error("failed to fetch Project Setting in linkedin filter values.")
		return "", []interface{}{}, http.StatusInternalServerError
	}
	customerAccountID := projectSetting.IntLinkedinAdAccount
	params := []interface{}{linkedinInternalFilterProperty, projectID, customerAccountID, docType, linkedinInternalFilterProperty}

	return "(" + linkedinFilterQueryStr + ")", params, http.StatusFound
}

func (pg *Postgres) ExecuteLinkedinChannelQueryV1(projectID uint64, query *model.ChannelQueryV1, reqID string) ([]string, [][]interface{}, error) {
	fetchSource := false
	logCtx := log.WithField("xreq_id", reqID)
	sql, params, selectKeys, selectMetrics, err := pg.GetSQLQueryAndParametersForLinkedinQueryV1(projectID, query, reqID, fetchSource)
	if err != nil {
		return make([]string, 0, 0), make([][]interface{}, 0, 0), err
	}
	_, resultMetrics, err := pg.ExecuteSQL(sql, params, logCtx)
	columns := append(selectKeys, selectMetrics...)
	return columns, resultMetrics, err
}

func (pg *Postgres) GetSQLQueryAndParametersForLinkedinQueryV1(projectID uint64, query *model.ChannelQueryV1, reqID string, fetchSource bool) (string, []interface{}, []string, []string, error) {
	var selectMetrics []string
	var sql string
	var selectKeys []string
	var params []interface{}
	transformedQuery, customerAccountID, err := pg.transFormRequestFieldsAndFetchRequiredFieldsForLinkedin(projectID, *query, reqID)
	if err != nil {
		return "", make([]interface{}, 0, 0), make([]string, 0, 0), make([]string, 0, 0), err
	}
	sql, params, selectKeys, selectMetrics, err = buildLinkedinQueryV1(transformedQuery, projectID, customerAccountID, fetchSource)
	if err != nil {
		return "", make([]interface{}, 0, 0), make([]string, 0, 0), make([]string, 0, 0), err
	}
	return sql, params, selectKeys, selectMetrics, nil
}

func (pg *Postgres) transFormRequestFieldsAndFetchRequiredFieldsForLinkedin(projectID uint64, query model.ChannelQueryV1, reqID string) (*model.ChannelQueryV1, string, error) {
	query.From = getAdwordsDateOnlyTimestampInInt64(query.From)
	query.To = getAdwordsDateOnlyTimestampInInt64(query.To)
	var err error
	logCtx := log.WithField("req_id", reqID)
	projectSetting, errCode := pg.GetProjectSetting(projectID)
	if errCode != http.StatusFound {
		return &model.ChannelQueryV1{}, "", errors.New("Project setting not found")
	}
	customerAccountID := projectSetting.IntLinkedinAdAccount

	query, err = convertFromRequestToLinkedinSpecificRepresentation(query)
	if err != nil {
		logCtx.Warn("Request failed in validation: ", err)
		return &model.ChannelQueryV1{}, "", err
	}
	return &query, customerAccountID, nil
}

func convertFromRequestToLinkedinSpecificRepresentation(query model.ChannelQueryV1) (model.ChannelQueryV1, error) {
	var err1, err2, err3 error
	query.SelectMetrics, err1 = getLinkedinSpecificMetrics(query.SelectMetrics)
	query.Filters, err2 = getLinkedinSpecificFilters(query.Filters)
	query.GroupBy, err3 = getLinkedinSpecificGroupBy(query.GroupBy)
	if err1 != nil {
		return query, err1
	}
	if err2 != nil {
		return query, err2
	}
	if err3 != nil {
		return query, err3
	}
	return query, nil
}

// @Kark TODO v1
func getLinkedinSpecificMetrics(requestSelectMetrics []string) ([]string, error) {
	resultMetrics := make([]string, 0, 0)
	for _, requestMetric := range requestSelectMetrics {
		metric, isPresent := linkedinExternalRepresentationToInternalRepresentation[requestMetric]
		if !isPresent {
			return make([]string, 0, 0), errors.New("Invalid metric found for document type")
		}
		resultMetrics = append(resultMetrics, metric)
	}
	return resultMetrics, nil
}

// @Kark TODO v1
func getLinkedinSpecificFilters(requestFilters []model.ChannelFilterV1) ([]model.ChannelFilterV1, error) {
	for index, requestFilter := range requestFilters {
		filterObject, isPresent := linkedinExternalRepresentationToInternalRepresentation[requestFilter.Object]
		if !isPresent {
			return make([]model.ChannelFilterV1, 0, 0), errors.New("Invalid filter key found for document type")
		}
		(&requestFilters[index]).Object = filterObject
	}
	return requestFilters, nil
}

// @Kark TODO v1
func getLinkedinSpecificGroupBy(requestGroupBys []model.ChannelGroupBy) ([]model.ChannelGroupBy, error) {
	for index, requestGroupBy := range requestGroupBys {
		groupByObject, isPresent := linkedinExternalRepresentationToInternalRepresentation[requestGroupBy.Object]
		if !isPresent {
			return make([]model.ChannelGroupBy, 0, 0), errors.New("Invalid groupby key found for document type")
		}
		(&requestGroupBys[index]).Object = groupByObject
	}
	return requestGroupBys, nil
}

func buildLinkedinQueryV1(query *model.ChannelQueryV1, projectID uint64, customerAccountID string, fetchSource bool) (string, []interface{}, []string, []string, error) {
	lowestHierarchyLevel := getLowestHierarchyLevelForLinkedin(query)
	lowestHierarchyReportLevel := lowestHierarchyLevel + "_insights"
	sql, params, selectKeys, selectMetrics := getSQLAndParamsFromLinkedinReports(query, projectID, query.From, query.To, customerAccountID, linkedinDocumentTypeAlias[lowestHierarchyReportLevel],
		fetchSource)
	return sql, params, selectKeys, selectMetrics, nil
}
func getSQLAndParamsFromLinkedinReports(query *model.ChannelQueryV1, projectID uint64, from, to int64, linkedinAccountIDs string,
	docType int, fetchSource bool) (string, []interface{}, []string, []string) {
	customerAccountIDs := strings.Split(linkedinAccountIDs, ",")
	selectQuery := "SELECT "
	selectMetrics := make([]string, 0, 0)
	isGroupByTimestamp := query.GetGroupByTimestamp() != ""
	groupByStatement := ""
	groupByKeysWithoutTimestamp := make([]string, 0, 0)
	selectKeys := make([]string, 0, 0)
	finalSelectKeys := make([]string, 0, 0)
	responseSelectKeys := make([]string, 0, 0)
	responseSelectMetrics := make([]string, 0, 0)

	// Group By
	for _, groupBy := range query.GroupBy {
		key := groupBy.Object + ":" + groupBy.Property
		groupByKeysWithoutTimestamp = append(groupByKeysWithoutTimestamp, linkedinInternalGroupByRepresentation[key])
	}
	if isGroupByTimestamp {
		groupByStatement = joinWithComma(append(groupByKeysWithoutTimestamp, model.AliasDateTime)...)
	} else {
		groupByStatement = joinWithComma(groupByKeysWithoutTimestamp...)
	}

	// SelectKeys
	if fetchSource {
		finalSelectKeys = append(finalSelectKeys, fmt.Sprintf("'%s' as %s", linkedinStringColumn, source))
		responseSelectKeys = append(responseSelectKeys, source)
	}

	for _, groupBy := range query.GroupBy {
		key := groupBy.Object + ":" + groupBy.Property
		value := fmt.Sprintf("%s as %s", objectAndPropertyToValueInLinkedinReportsMapping[key], linkedinInternalRepresentationToExternalRepresentation[key])
		selectKeys = append(selectKeys, value)
		responseSelectKeys = append(responseSelectKeys, linkedinInternalRepresentationToExternalRepresentation[key])
	}

	finalSelectKeys = append(finalSelectKeys, selectKeys...)
	if isGroupByTimestamp {
		finalSelectKeys = append(finalSelectKeys, fmt.Sprintf("%s as %s",
			getSelectTimestampByTypeForChannels(query.GetGroupByTimestamp(), query.Timezone), model.AliasDateTime))
		responseSelectKeys = append(responseSelectKeys, model.AliasDateTime)
	}

	for _, selectMetric := range query.SelectMetrics {
		value := fmt.Sprintf("%s as %s", linkedinMetricsToAggregatesInReportsMapping[selectMetric], linkedinInternalRepresentationToExternalRepresentation[selectMetric])
		selectMetrics = append(selectMetrics, value)

		value = linkedinInternalRepresentationToExternalRepresentation[selectMetric]
		responseSelectMetrics = append(responseSelectMetrics, value)
	}

	selectQuery += joinWithComma(append(finalSelectKeys, selectMetrics...)...)
	orderByQuery := "ORDER BY " + getOrderByClause(responseSelectMetrics)
	whereConditionForFilters := getLinkedinFiltersWhereStatement(query.Filters)

	resultSQLStatement := selectQuery + fromLinkedinDocument + staticWhereStatementForLinkedin + whereConditionForFilters
	if len(groupByStatement) != 0 {
		resultSQLStatement += "GROUP BY " + groupByStatement
	}
	resultSQLStatement += " " + orderByQuery + channeAnalyticsLimit + ";"
	staticWhereParams := []interface{}{projectID, customerAccountIDs, docType, from, to}
	return resultSQLStatement, staticWhereParams, responseSelectKeys, responseSelectMetrics
}

func getLinkedinFiltersWhereStatement(filters []model.ChannelFilterV1) string {
	resultStatement := ""
	var filterValue string
	for index, filter := range filters {
		currentFilterStatement := ""
		if filter.LogicalOp == "" {
			filter.LogicalOp = "AND"
		}
		filterOperator := getOp(filter.Condition)
		if filter.Condition == model.ContainsOpStr || filter.Condition == model.NotContainsOpStr {
			filterValue = fmt.Sprintf("%%%s%%", filter.Value)
		} else {
			filterValue = filter.Value
		}
		currentFilterStatement = fmt.Sprintf("%s %s '%s' ", objectToValueInLinkedinFiltersMapping[filter.Object+":"+filter.Property], filterOperator, filterValue)
		if index == 0 {
			resultStatement = " AND " + currentFilterStatement
		} else {
			resultStatement = fmt.Sprintf("%s %s %s ", resultStatement, filter.LogicalOp, currentFilterStatement)
		}
	}
	return resultStatement
}

func getLowestHierarchyLevelForLinkedin(query *model.ChannelQueryV1) string {
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
		if objectName == linkedinCreative {
			return linkedinCreative
		}
	}

	for _, objectName := range objectNames {
		if objectName == linkedinCampaign {
			return linkedinCampaign
		}
	}

	for _, objectName := range objectNames {
		if objectName == linkedinCampaignGroup {
			return linkedinCampaignGroup
		}
	}

	return linkedinCampaignGroup
}
