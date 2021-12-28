package memsql

import (
	"errors"
	C "factors/config"
	Const "factors/constants"
	"factors/model/model"
	U "factors/util"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
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
	"campaign_group:id":   "campaign_group_id",
	"creative:id":         "creative_id",
	"campaign:id":         "campaign_id",
	"campaign_group:name": "JSON_EXTRACT_STRING(value, 'campaign_group_name')",
	"campaign:name":       "JSON_EXTRACT_STRING(value, 'campaign_name')",
}

// TODO check
var linkedinMetricsToAggregatesInReportsMapping = map[string]string{
	"impressions": "SUM(JSON_EXTRACT_STRING(value, 'impressions'))",
	"clicks":      "SUM(JSON_EXTRACT_STRING(value, 'clicks'))",
	"spend":       "SUM(JSON_EXTRACT_STRING(value, 'costInLocalCurrency'))",
	"conversions": "SUM(JSON_EXTRACT_STRING(value, 'conversionValueInLocalCurrency'))",
	// "cost_per_click": "average_cost",
	// "conversion_rate": "conversion_rate"
}

var objectToValueInLinkedinFiltersMapping = map[string]string{
	"campaign:name":       "JSON_EXTRACT_STRING(value, 'campaign_name')",
	"campaign_group:name": "JSON_EXTRACT_STRING(value, 'campaign_group_name')",
	"campaign:id":         "campaign_id",
	"campaign_group:id":   "campaign_group_id",
	"creative:id":         "creative_id",
}
var objectToValueInLinkedinFiltersMappingWithLinkedinDocuments = map[string]string{
	"campaign:name":       "JSON_EXTRACT_STRING(linkedin_documents.value, 'campaign_name')",
	"campaign_group:name": "JSON_EXTRACT_STRING(linkedin_documents.value, 'campaign_group_name')",
	"campaign:id":         "linkedin_documents.campaign_id",
	"campaign_group:id":   "linkedin_documents.campaign_group_id",
	"creative:id":         "linkedin_documents.creative_id",
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

var errorEmptyLinkedinDocument = errors.New("empty linked document")

const linkedinFilterQueryStr = "SELECT DISTINCT(LCASE(JSON_EXTRACT_STRING(value, ?))) as filter_value FROM linkedin_documents WHERE project_id = ? AND" +
	" " + "customer_ad_account_id = ? AND type = ? AND JSON_EXTRACT_STRING(value, ?) IS NOT NULL LIMIT 5000"

const fromLinkedinDocuments = " FROM linkedin_documents "

const staticWhereStatementForLinkedin = "WHERE project_id = ? AND customer_ad_account_id IN ( ? ) AND type = ? AND timestamp between ? AND ? "
const staticWhereStatementForLinkedinWithSmartProperty = "WHERE linkedin_documents.project_id = ? AND linkedin_documents.customer_ad_account_id IN ( ? ) AND linkedin_documents.type = ? AND linkedin_documents.timestamp between ? AND ? "

const linkedinAdGroupMetadataFetchQueryStr = "WITH ad_group as (select campaign_id as ad_group_id, JSON_EXTRACT_STRING(value, 'name') as ad_group_name, campaign_group_id " +
	"from linkedin_documents where type = ? AND project_id = ? AND timestamp BETWEEN ? AND ? AND customer_ad_account_id IN (?) " +
	"AND (campaign_id, timestamp) in (select campaign_id, max(timestamp) from linkedin_documents where type = ? AND " +
	"project_id = ? AND timestamp between ? and ? AND customer_ad_account_id IN (?) group by campaign_id)), campaign as " +
	"(select campaign_group_id as campaign_id, JSON_EXTRACT_STRING(value, 'name') as campaign_name from linkedin_documents where type = ? AND " +
	"project_id = ?  AND timestamp BETWEEN ? AND ? AND customer_ad_account_id IN (?) and (campaign_group_id, timestamp) in " +
	"(select campaign_group_id, max(timestamp) from linkedin_documents where type = ? and project_id = ? and timestamp " +
	"BETWEEN ? and ?  AND customer_ad_account_id IN (?) group by campaign_group_id)) select ad_group_id, ad_group_name, " +
	"campaign.campaign_id, campaign_name from ad_group join campaign on ad_group.campaign_group_id = campaign.campaign_id"

const linkedinCampaignMetadataFetchQueryStr = "select campaign_group_id as campaign_id, JSON_EXTRACT_STRING(value, 'name') as campaign_name from linkedin_documents where " +
	"type = ? AND project_id = ? AND timestamp BETWEEN ? AND ? AND customer_ad_account_id IN (?) and " +
	"(campaign_group_id, timestamp) in (select campaign_group_id, max(timestamp) from linkedin_documents where type = ? " +
	"and project_id = ? and timestamp BETWEEN ? and ? AND customer_ad_account_id IN (?) group by campaign_group_id)"

func (store *MemSQL) satisfiesLinkedinDocumentForeignConstraints(linkedinDocument model.LinkedinDocument) int {
	_, errCode := store.GetProject(linkedinDocument.ProjectID)
	if errCode != http.StatusFound {
		return http.StatusBadRequest
	}
	return http.StatusOK
}

func (store *MemSQL) satisfiesLinkedinDocumentUniquenessConstraints(linkedinDocument *model.LinkedinDocument) int {
	errCode := store.isLinkedinDocumentExistByPrimaryKey(linkedinDocument)
	if errCode == http.StatusFound {
		return http.StatusConflict
	}
	if errCode == http.StatusNotFound {
		return http.StatusOK
	}
	return errCode
}

// Checks PRIMARY KEY (project_id, customer_ad_account_id, type, timestamp, id)
func (store *MemSQL) isLinkedinDocumentExistByPrimaryKey(document *model.LinkedinDocument) int {
	logCtx := log.WithField("document", document)

	if document.ProjectID == 0 || document.CustomerAdAccountID == "" || document.Type == 0 ||
		document.Timestamp == 0 || document.ID == "" {

		log.Error("Invalid linkedin document on primary constraint check.")
		return http.StatusBadRequest
	}

	var linkedinDocument model.LinkedinDocument

	db := C.GetServices().Db
	if err := db.Limit(1).Where("project_id = ? AND customer_ad_account_id = ? AND type = ? AND timestamp = ? AND id = ?",
		document.ProjectID, document.CustomerAdAccountID, document.Type, document.Timestamp, document.ID,
	).Select("id").Find(&linkedinDocument).Error; err != nil {

		if gorm.IsRecordNotFoundError(err) {
			return http.StatusNotFound
		}

		logCtx.WithError(err).
			Error("Failed getting to check existence linkedin document by primary keys.")
		return http.StatusInternalServerError
	}

	if linkedinDocument.ID == "" {
		logCtx.Error("Invalid id value returned on linkedin document primary key check.")
		return http.StatusInternalServerError
	}

	return http.StatusFound
}

func getLinkedinDocumentTypeAliasByType() map[int]string {
	documentTypeMap := make(map[int]string, 0)
	for alias, typ := range linkedinDocumentTypeAlias {
		documentTypeMap[typ] = alias
	}

	return documentTypeMap
}

func (store *MemSQL) GetLinkedinLastSyncInfo(projectID uint64, CustomerAdAccountID string) ([]model.LinkedinLastSyncInfo, int) {
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

// CreatelinkedinDocument ...
func (store *MemSQL) CreateLinkedinDocument(projectID uint64, document *model.LinkedinDocument) int {
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
	if errCode := store.satisfiesLinkedinDocumentForeignConstraints(*document); errCode != http.StatusOK {
		return http.StatusInternalServerError
	}

	errCode := store.satisfiesLinkedinDocumentUniquenessConstraints(document)
	if errCode != http.StatusOK {
		return errCode
	}

	db := C.GetServices().Db
	err := db.Create(&document).Error
	if err != nil {
		if IsDuplicateRecordError(err) {
			logCtx.WithError(err).WithField("id", document.ID).Error(
				"Failed to create an linkedin doc. Duplicate.")
			return http.StatusConflict
		}
		logCtx.WithError(err).WithField("id", document.ID).Error(
			"Failed to create an linkedin doc. Continued inserting other docs.")
		return http.StatusInternalServerError
	}
	UpdateCountCacheByDocumentType(projectID, &document.CreatedAt, "linkedin")
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

func (store *MemSQL) ExecuteLinkedinChannelQuery(projectID uint64, query *model.ChannelQuery) (*model.ChannelQueryResult, int) {
	defer U.NotifyOnPanicWithError(C.GetConfig().Env, C.GetConfig().AppName)

	logCtx := log.WithField("project_id", projectID).WithField("query", query)

	if projectID == 0 || query == nil {
		logCtx.Error("Invalid project_id or query on execute linkedin channel query.")
		return nil, http.StatusInternalServerError
	}

	projectSetting, errCode := store.GetProjectSetting(projectID)
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

	metricBreakDown, err := store.getLinkedinMetricBreakdown(projectID, projectSetting.IntLinkedinAdAccount, query)
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
	resultHeaders, resultRows, err := U.DBReadRows(rows, nil)
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
func (store *MemSQL) getLinkedinMetricBreakdown(projectID uint64, customerAccountID string, query *model.ChannelQuery) (*model.ChannelBreakdownResult, error) {
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

	resultHeaders, resultRows, err := U.DBReadRows(rows, nil)
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

	selectColstWithoutAlias := "SUM(JSON_EXTRACT_STRING(value, 'impressions')) as %s , SUM(JSON_EXTRACT_STRING(value, 'clicks')) as %s," +
		" " + "SUM(JSON_EXTRACT_STRING(value, 'costInUsd')) as %s," +
		" " + "SUM(JSON_EXTRACT_STRING(value, 'approximateUniqueImpressions')) as %s," +
		" " + "SUM(JSON_EXTRACT_STRING(value, 'likes')) as %s, SUM(JSON_EXTRACT_STRING(value, 'follows')) as %s, " +
		" " + "SUM(JSON_EXTRACT_STRING(value, 'totalEngagements')) as %s," +
		" " + "AVG(JSON_EXTRACT_STRING(value, 'conversionValueInLocalCurrency')) as %s"

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
func (store *MemSQL) buildLinkedinChannelConfig(projectID uint64) *model.ChannelConfigResult {
	linkedinObjectsAndProperties := store.buildObjectAndPropertiesForLinkedin(projectID, model.ObjectsForLinkedin)
	objectsAndProperties := append(linkedinObjectsAndProperties)

	return &model.ChannelConfigResult{
		SelectMetrics:        selectableMetricsForAllChannels,
		ObjectsAndProperties: objectsAndProperties,
	}
}

func (store *MemSQL) buildObjectAndPropertiesForLinkedin(projectID uint64, objects []string) []model.ChannelObjectAndProperties {
	objectsAndProperties := make([]model.ChannelObjectAndProperties, 0, 0)
	for _, currentObject := range objects {
		var currentProperties []model.ChannelProperty
		var currentPropertiesSmart []model.ChannelProperty
		currentProperties = buildProperties(allChannelsPropertyToRelated)
		smartProperty := store.GetSmartPropertyAndRelated(projectID, currentObject, "linkedin")
		currentPropertiesSmart = buildProperties(smartProperty)
		currentProperties = append(currentProperties, currentPropertiesSmart...)
		objectsAndProperties = append(objectsAndProperties, buildObjectsAndProperties(currentProperties, []string{currentObject})...)
	}
	return objectsAndProperties
}

func (store *MemSQL) GetLinkedinFilterValues(projectID uint64, requestFilterObject string, requestFilterProperty string, reqID string) ([]interface{}, int) {
	_, isPresent := Const.SmartPropertyReservedNames[requestFilterProperty]
	if !isPresent {
		filterValues, errCode := store.getSmartPropertyFilterValues(projectID, requestFilterObject, requestFilterProperty, "linkedin", reqID)
		if errCode != http.StatusFound {
			return []interface{}{}, http.StatusInternalServerError
		}
		return filterValues, http.StatusFound
	}
	linkedinInternalFilterProperty, docType, err := getFilterRelatedInformationForLinkedin(requestFilterObject, requestFilterProperty)
	if err != http.StatusOK {
		return make([]interface{}, 0, 0), http.StatusBadRequest
	}
	filterValues, errCode := store.getLinkedinFilterValuesByType(projectID, docType, linkedinInternalFilterProperty, reqID)
	if errCode != http.StatusFound {
		return []interface{}{}, http.StatusInternalServerError
	}

	return filterValues, http.StatusFound
}

func getFilterRelatedInformationForLinkedin(requestFilterObject string, requestFilterProperty string) (string, int, int) {
	linkedinInternalFilterObject, isPresent := model.LinkedinExternalRepresentationToInternalRepresentation[requestFilterObject]
	if !isPresent {
		log.Error("Invalid linkedin filter object.")
		return "", 0, http.StatusBadRequest
	}
	linkedinInternalFilterProperty, isPresent := model.LinkedinExternalRepresentationToInternalRepresentation[requestFilterProperty]
	if !isPresent {
		log.Error("Invalid linkedin filter property.")
		return "", 0, http.StatusBadRequest
	}
	docType := linkedinDocumentTypeAlias[linkedinInternalFilterObject]

	return linkedinInternalFilterProperty, docType, http.StatusOK
}

func (store *MemSQL) getLinkedinFilterValuesByType(projectID uint64, docType int, property string, reqID string) ([]interface{}, int) {
	logCtx := log.WithField("projectID", projectID).WithField("req_id", reqID)
	projectSetting, errCode := store.GetProjectSetting(projectID)
	if errCode != http.StatusFound {
		logCtx.Error("Failed to fetch Project Setting in linkedin filter values.")
		return []interface{}{}, http.StatusInternalServerError
	}
	customerAccountID := projectSetting.IntLinkedinAdAccount
	if customerAccountID == "" || len(customerAccountID) == 0 {
		logCtx.Info(integrationNotAvailable)
		return []interface{}{}, http.StatusNotFound
	}
	logCtx = log.WithField("project_id", projectID).WithField("doc_type", docType).WithField("req_id", reqID)
	params := []interface{}{property, projectID, customerAccountID, docType, property}
	_, resultRows, err := store.ExecuteSQL(linkedinFilterQueryStr, params, logCtx)
	if err != nil {
		logCtx.WithError(err).WithField("query", linkedinFilterQueryStr).WithField("params", params).Error(model.LinkedinSpecificError)
		return make([]interface{}, 0, 0), http.StatusInternalServerError
	}

	return Convert2DArrayTo1DArray(resultRows), http.StatusFound
}

func (store *MemSQL) GetLinkedinSQLQueryAndParametersForFilterValues(projectID uint64, requestFilterObject string, requestFilterProperty string, reqID string) (string, []interface{}, int) {
	logCtx := log.WithField("project_id", projectID).WithField("req_id", reqID)
	linkedinInternalFilterProperty, docType, err := getFilterRelatedInformationForLinkedin(requestFilterObject, requestFilterProperty)
	if err != http.StatusOK {
		return "", make([]interface{}, 0, 0), http.StatusBadRequest
	}
	projectSetting, errCode := store.GetProjectSetting(projectID)
	if errCode != http.StatusFound {
		logCtx.Error("failed to fetch Project Setting in linkedin filter values.")
		return "", make([]interface{}, 0, 0), http.StatusInternalServerError
	}
	customerAccountID := projectSetting.IntLinkedinAdAccount
	if customerAccountID == "" || len(customerAccountID) == 0 {
		logCtx.Info(integrationNotAvailable)
		return "", nil, http.StatusNotFound
	}
	params := []interface{}{linkedinInternalFilterProperty, projectID, customerAccountID, docType, linkedinInternalFilterProperty}

	return "(" + linkedinFilterQueryStr + ")", params, http.StatusFound
}

func (store *MemSQL) ExecuteLinkedinChannelQueryV1(projectID uint64, query *model.ChannelQueryV1, reqID string) ([]string, [][]interface{}, int) {
	defer U.NotifyOnPanicWithError(C.GetConfig().Env, C.GetConfig().AppName)
	fetchSource := false
	logCtx := log.WithField("xreq_id", reqID)
	if query.GroupByTimestamp == "" {
		sql, params, selectKeys, selectMetrics, errCode := store.GetSQLQueryAndParametersForLinkedinQueryV1(projectID,
			query, reqID, fetchSource, " LIMIT 10000", false, nil)
		if errCode != http.StatusOK {
			return make([]string, 0, 0), make([][]interface{}, 0, 0), errCode
		}
		_, resultMetrics, err := store.ExecuteSQL(sql, params, logCtx)
		columns := append(selectKeys, selectMetrics...)
		if err != nil {
			logCtx.WithError(err).WithField("query", sql).WithField("params", params).Error(model.LinkedinSpecificError)
			return make([]string, 0, 0), make([][]interface{}, 0, 0), http.StatusInternalServerError
		}
		return columns, resultMetrics, http.StatusOK
	} else {
		sql, params, selectKeys, selectMetrics, errCode := store.GetSQLQueryAndParametersForLinkedinQueryV1(
			projectID, query, reqID, fetchSource, " LIMIT 100", false, nil)
		if errCode != http.StatusOK {
			return make([]string, 0, 0), make([][]interface{}, 0, 0), errCode
		}
		_, resultMetrics, err := store.ExecuteSQL(sql, params, logCtx)
		columns := append(selectKeys, selectMetrics...)
		if err != nil {
			logCtx.WithError(err).WithField("query", sql).WithField("params", params).Error(model.LinkedinSpecificError)
			return make([]string, 0, 0), make([][]interface{}, 0, 0), http.StatusInternalServerError
		}
		groupByCombinations := model.GetGroupByCombinationsForChannelAnalytics(columns, resultMetrics)
		sql, params, selectKeys, selectMetrics, errCode = store.GetSQLQueryAndParametersForLinkedinQueryV1(
			projectID, query, reqID, fetchSource, " LIMIT 10000", true, groupByCombinations)
		if errCode != http.StatusOK {
			return make([]string, 0, 0), make([][]interface{}, 0, 0), errCode
		}
		_, resultMetrics, err = store.ExecuteSQL(sql, params, logCtx)
		columns = append(selectKeys, selectMetrics...)
		if err != nil {
			logCtx.WithError(err).WithField("query", sql).WithField("params", params).Error(model.LinkedinSpecificError)
			return make([]string, 0, 0), make([][]interface{}, 0, 0), http.StatusInternalServerError
		}
		return columns, resultMetrics, http.StatusOK
	}
}

func (store *MemSQL) GetSQLQueryAndParametersForLinkedinQueryV1(projectID uint64, query *model.ChannelQueryV1, reqID string, fetchSource bool,
	limitString string, isGroupByTimestamp bool, groupByCombinationsForGBT []map[string]interface{}) (string, []interface{}, []string, []string, int) {
	var selectMetrics []string
	var sql string
	var selectKeys []string
	var params []interface{}
	logCtx := log.WithField("project_id", projectID).WithField("req_id", reqID)
	transformedQuery, customerAccountID, err := store.transFormRequestFieldsAndFetchRequiredFieldsForLinkedin(projectID, *query, reqID)
	if err != nil && err.Error() == integrationNotAvailable {
		logCtx.WithError(err).Info(model.LinkedinSpecificError)
		return "", make([]interface{}, 0, 0), make([]string, 0, 0), make([]string, 0, 0), http.StatusNotFound
	}
	if err != nil {
		logCtx.WithError(err).Error(model.LinkedinSpecificError)
		return "", make([]interface{}, 0, 0), make([]string, 0, 0), make([]string, 0, 0), http.StatusBadRequest
	}
	isSmartPropertyPresent := checkSmartProperty(query.Filters, query.GroupBy)
	if isSmartPropertyPresent {
		sql, params, selectKeys, selectMetrics, err = buildLinkedinQueryWithSmartPropertyV1(transformedQuery, projectID, customerAccountID, fetchSource,
			limitString, isGroupByTimestamp, groupByCombinationsForGBT)
		if err != nil {
			return "", make([]interface{}, 0, 0), make([]string, 0, 0), make([]string, 0, 0), http.StatusInternalServerError
		}
		return sql, params, selectKeys, selectMetrics, http.StatusOK
	}

	sql, params, selectKeys, selectMetrics, err = buildLinkedinQueryV1(transformedQuery, projectID, customerAccountID, fetchSource,
		limitString, isGroupByTimestamp, groupByCombinationsForGBT)
	if err != nil {
		return "", make([]interface{}, 0, 0), make([]string, 0, 0), make([]string, 0, 0), http.StatusInternalServerError
	}
	return sql, params, selectKeys, selectMetrics, http.StatusOK
}

func (store *MemSQL) transFormRequestFieldsAndFetchRequiredFieldsForLinkedin(projectID uint64, query model.ChannelQueryV1, reqID string) (*model.ChannelQueryV1, string, error) {
	query.From = U.GetDateAsStringIn(query.From, U.TimeZoneString(query.Timezone))
	query.To = U.GetDateAsStringIn(query.To, U.TimeZoneString(query.Timezone))
	var err error
	logCtx := log.WithField("req_id", reqID)
	projectSetting, errCode := store.GetProjectSetting(projectID)
	if errCode != http.StatusFound {
		return &model.ChannelQueryV1{}, "", errors.New("Project setting not found")
	}
	customerAccountID := projectSetting.IntLinkedinAdAccount
	if customerAccountID == "" || len(customerAccountID) == 0 {
		return &model.ChannelQueryV1{}, "", errors.New(integrationNotAvailable)
	}

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
		metric, isPresent := model.LinkedinExternalRepresentationToInternalRepresentation[requestMetric]
		if !isPresent {
			return make([]string, 0, 0), errors.New("Invalid metric found for document type")
		}
		resultMetrics = append(resultMetrics, metric)
	}
	return resultMetrics, nil
}

// @Kark TODO v1
func getLinkedinSpecificFilters(requestFilters []model.ChannelFilterV1) ([]model.ChannelFilterV1, error) {
	filters := make([]model.ChannelFilterV1, 0)
	for _, requestFilter := range requestFilters {
		filterObject, isPresent := model.LinkedinExternalRepresentationToInternalRepresentation[requestFilter.Object]
		if !isPresent {
			return make([]model.ChannelFilterV1, 0, 0), errors.New("Invalid filter key found for document type")

		}
		filters = append(filters, model.ChannelFilterV1{Object: filterObject, Property: requestFilter.Property, Condition: requestFilter.Condition,
			Value: requestFilter.Value, LogicalOp: requestFilter.LogicalOp})
	}
	return filters, nil
}

// @Kark TODO v1
func getLinkedinSpecificGroupBy(requestGroupBys []model.ChannelGroupBy) ([]model.ChannelGroupBy, error) {
	groupBys := make([]model.ChannelGroupBy, 0)
	for _, requestGroupBy := range requestGroupBys {
		groupByObject, isPresent := model.LinkedinExternalRepresentationToInternalRepresentation[requestGroupBy.Object]
		if !isPresent {
			return make([]model.ChannelGroupBy, 0, 0), errors.New("Invalid groupby key found for document type")
		}
		groupBys = append(groupBys, model.ChannelGroupBy{Object: groupByObject, Property: requestGroupBy.Property})
	}
	return groupBys, nil
}

func buildLinkedinQueryWithSmartPropertyV1(query *model.ChannelQueryV1, projectID uint64, customerAccountID string, fetchSource bool,
	limitString string, isGroupByTimestamp bool, groupByCombinationsForGBT []map[string]interface{}) (string, []interface{}, []string, []string, error) {
	lowestHierarchyLevel := getLowestHierarchyLevelForLinkedin(query)
	lowestHierarchyReportLevel := lowestHierarchyLevel + "_insights"
	sql, params, selectKeys, selectMetrics := getSQLAndParamsFromLinkedinWithSmartPropertyReports(query, projectID, query.From, query.To, customerAccountID, linkedinDocumentTypeAlias[lowestHierarchyReportLevel],
		fetchSource, limitString, isGroupByTimestamp, groupByCombinationsForGBT)
	return sql, params, selectKeys, selectMetrics, nil
}
func buildLinkedinQueryV1(query *model.ChannelQueryV1, projectID uint64, customerAccountID string, fetchSource bool,
	limitString string, isGroupByTimestamp bool, groupByCombinationsForGBT []map[string]interface{}) (string, []interface{}, []string, []string, error) {
	lowestHierarchyLevel := getLowestHierarchyLevelForLinkedin(query)
	lowestHierarchyReportLevel := lowestHierarchyLevel + "_insights"
	sql, params, selectKeys, selectMetrics := getSQLAndParamsFromLinkedinReports(query, projectID, query.From, query.To, customerAccountID, linkedinDocumentTypeAlias[lowestHierarchyReportLevel],
		fetchSource, limitString, isGroupByTimestamp, groupByCombinationsForGBT)
	return sql, params, selectKeys, selectMetrics, nil
}
func getSQLAndParamsFromLinkedinWithSmartPropertyReports(query *model.ChannelQueryV1, projectID uint64, from, to int64, linkedinAccountIDs string, docType int,
	fetchSource bool, limitString string, isGroupByTimestamp bool, groupByCombinationsForGBT []map[string]interface{}) (string, []interface{}, []string, []string) {
	customerAccountIDs := strings.Split(linkedinAccountIDs, ",")
	selectQuery := "SELECT "
	selectMetrics := make([]string, 0, 0)
	groupByStatement := ""
	groupByKeysWithoutTimestamp := make([]string, 0, 0)
	selectKeys := make([]string, 0, 0)
	finalSelectKeys := make([]string, 0, 0)
	responseSelectKeys := make([]string, 0, 0)
	responseSelectMetrics := make([]string, 0, 0)

	smartPropertyCampaignGroupBys := make([]model.ChannelGroupBy, 0, 0)
	smartPropertyAdGroupGroupBys := make([]model.ChannelGroupBy, 0, 0)
	linkedinGroupBys := make([]model.ChannelGroupBy, 0, 0)
	// Group By
	for _, groupBy := range query.GroupBy {
		_, isPresent := Const.SmartPropertyReservedNames[groupBy.Property]
		if !isPresent {
			if groupBy.Object == "campaign_group" {
				smartPropertyCampaignGroupBys = append(smartPropertyCampaignGroupBys, groupBy)
				groupByKeysWithoutTimestamp = append(groupByKeysWithoutTimestamp, fmt.Sprintf("campaign_%s", groupBy.Property))
			} else {
				smartPropertyAdGroupGroupBys = append(smartPropertyAdGroupGroupBys, groupBy)
				groupByKeysWithoutTimestamp = append(groupByKeysWithoutTimestamp, fmt.Sprintf("ad_group_%s", groupBy.Property))
			}
		} else {
			key := groupBy.Object + ":" + groupBy.Property
			groupByKeysWithoutTimestamp = append(groupByKeysWithoutTimestamp, model.LinkedinInternalGroupByRepresentation[key])
			linkedinGroupBys = append(linkedinGroupBys, groupBy)
		}
	}
	if isGroupByTimestamp {
		groupByStatement = joinWithComma(append(groupByKeysWithoutTimestamp, model.AliasDateTime)...)
	} else {
		groupByStatement = joinWithComma(groupByKeysWithoutTimestamp...)
	}

	// SelectKeys

	for _, groupBy := range linkedinGroupBys {
		key := groupBy.Object + ":" + groupBy.Property
		value := fmt.Sprintf("%s as %s", objectAndPropertyToValueInLinkedinReportsMapping[key], model.LinkedinInternalRepresentationToExternalRepresentation[key])
		selectKeys = append(selectKeys, value)
		responseSelectKeys = append(responseSelectKeys, model.LinkedinInternalRepresentationToExternalRepresentation[key])
	}
	for _, groupBy := range smartPropertyCampaignGroupBys {
		value := fmt.Sprintf("JSON_EXTRACT_STRING(campaign.properties, '%s') as campaign_%s", groupBy.Property, groupBy.Property)
		selectKeys = append(selectKeys, value)
		responseSelectKeys = append(responseSelectKeys, fmt.Sprintf("campaign_%s", groupBy.Property))
	}
	for _, groupBy := range smartPropertyAdGroupGroupBys {
		value := fmt.Sprintf("JSON_EXTRACT_STRING(ad_group.properties,'%s') as ad_group_%s", groupBy.Property, groupBy.Property)
		selectKeys = append(selectKeys, value)
		responseSelectKeys = append(responseSelectKeys, fmt.Sprintf("ad_group_%s", groupBy.Property))
	}

	finalSelectKeys = append(finalSelectKeys, selectKeys...)
	if isGroupByTimestamp {
		finalSelectKeys = append(finalSelectKeys, fmt.Sprintf("%s as %s",
			getSelectTimestampByTypeForChannels(query.GetGroupByTimestamp(), query.Timezone), model.AliasDateTime))
		responseSelectKeys = append(responseSelectKeys, model.AliasDateTime)
	}

	for _, selectMetric := range query.SelectMetrics {
		value := fmt.Sprintf("%s as %s", linkedinMetricsToAggregatesInReportsMapping[selectMetric], model.LinkedinInternalRepresentationToExternalRepresentation[selectMetric])
		selectMetrics = append(selectMetrics, value)

		value = model.LinkedinInternalRepresentationToExternalRepresentation[selectMetric]
		responseSelectMetrics = append(responseSelectMetrics, value)
	}

	selectQuery += joinWithComma(append(finalSelectKeys, selectMetrics...)...)
	orderByQuery := "ORDER BY " + getOrderByClause(isGroupByTimestamp, responseSelectMetrics)
	whereConditionForFilters := getLinkedinFiltersWhereStatementWithSmartProperty(query.Filters, smartPropertyCampaignGroupBys, smartPropertyAdGroupGroupBys)
	filterStatementForSmartPropertyGroupBy := getNotNullFilterStatementForSmartPropertyGroupBys(smartPropertyCampaignGroupBys, smartPropertyAdGroupGroupBys)
	finalFilterStatement := joinWithWordInBetween("AND", staticWhereStatementForLinkedinWithSmartProperty, whereConditionForFilters, filterStatementForSmartPropertyGroupBy)

	fromStatement := getLinkedinFromStatementWithJoins(query.Filters, query.GroupBy)
	finalParams := make([]interface{}, 0)
	staticWhereParams := []interface{}{projectID, customerAccountIDs, docType, from, to}
	finalParams = append(finalParams, staticWhereParams...)
	if len(groupByCombinationsForGBT) != 0 {
		whereConditionForGBT, whereParams := buildWhereConditionForGBTForLinkedin(groupByCombinationsForGBT)
		finalFilterStatement += (" AND (" + whereConditionForGBT + ")")
		finalParams = append(finalParams, whereParams...)
	}
	resultSQLStatement := selectQuery + fromStatement + finalFilterStatement
	if len(groupByStatement) != 0 {
		resultSQLStatement += " GROUP BY " + groupByStatement
	}

	resultSQLStatement += " " + orderByQuery + limitString + ";"
	return resultSQLStatement, finalParams, responseSelectKeys, responseSelectMetrics
}

func getLinkedinFromStatementWithJoins(filters []model.ChannelFilterV1, groupBys []model.ChannelGroupBy) string {
	isPresentCampaignSmartProperty, isPresentAdGroupSmartProperty := checkSmartPropertyWithTypeAndSource(filters, groupBys, "linkedin")
	fromStatement := fromLinkedinDocuments
	if isPresentAdGroupSmartProperty {
		fromStatement += "inner join smart_properties ad_group on ad_group.project_id = linkedin_documents.project_id and ad_group.object_id = campaign_id "
	}
	if isPresentCampaignSmartProperty {
		fromStatement += "inner join smart_properties campaign on campaign.project_id = linkedin_documents.project_id and campaign.object_id = campaign_group_id "
	}
	return fromStatement
}

func getSQLAndParamsFromLinkedinReports(query *model.ChannelQueryV1, projectID uint64, from, to int64, linkedinAccountIDs string, docType int,
	fetchSource bool, limitString string, isGroupByTimestamp bool, groupByCombinationsForGBT []map[string]interface{}) (string, []interface{}, []string, []string) {
	customerAccountIDs := strings.Split(linkedinAccountIDs, ",")
	selectQuery := "SELECT "
	selectMetrics := make([]string, 0, 0)
	groupByStatement := ""
	groupByKeysWithoutTimestamp := make([]string, 0, 0)
	selectKeys := make([]string, 0, 0)
	finalSelectKeys := make([]string, 0, 0)
	responseSelectKeys := make([]string, 0, 0)
	responseSelectMetrics := make([]string, 0, 0)

	// Group By
	for _, groupBy := range query.GroupBy {
		key := groupBy.Object + ":" + groupBy.Property
		groupByKeysWithoutTimestamp = append(groupByKeysWithoutTimestamp, model.LinkedinInternalGroupByRepresentation[key])
	}
	if isGroupByTimestamp {
		groupByStatement = joinWithComma(append(groupByKeysWithoutTimestamp, model.AliasDateTime)...)
	} else {
		groupByStatement = joinWithComma(groupByKeysWithoutTimestamp...)
	}

	// SelectKeys

	for _, groupBy := range query.GroupBy {
		key := groupBy.Object + ":" + groupBy.Property
		value := fmt.Sprintf("%s as %s", objectAndPropertyToValueInLinkedinReportsMapping[key], model.LinkedinInternalRepresentationToExternalRepresentation[key])
		selectKeys = append(selectKeys, value)
		responseSelectKeys = append(responseSelectKeys, model.LinkedinInternalRepresentationToExternalRepresentation[key])
	}

	finalSelectKeys = append(finalSelectKeys, selectKeys...)
	if isGroupByTimestamp {
		finalSelectKeys = append(finalSelectKeys, fmt.Sprintf("%s as %s",
			getSelectTimestampByTypeForChannels(query.GetGroupByTimestamp(), query.Timezone), model.AliasDateTime))
		responseSelectKeys = append(responseSelectKeys, model.AliasDateTime)
	}

	for _, selectMetric := range query.SelectMetrics {
		value := fmt.Sprintf("%s as %s", linkedinMetricsToAggregatesInReportsMapping[selectMetric], model.LinkedinInternalRepresentationToExternalRepresentation[selectMetric])
		selectMetrics = append(selectMetrics, value)

		value = model.LinkedinInternalRepresentationToExternalRepresentation[selectMetric]
		responseSelectMetrics = append(responseSelectMetrics, value)
	}

	selectQuery += joinWithComma(append(finalSelectKeys, selectMetrics...)...)
	orderByQuery := "ORDER BY " + getOrderByClause(isGroupByTimestamp, responseSelectMetrics)
	whereConditionForFilters := getLinkedinFiltersWhereStatement(query.Filters)
	finalFilterStatement := whereConditionForFilters
	finalParams := make([]interface{}, 0)
	staticWhereParams := []interface{}{projectID, customerAccountIDs, docType, from, to}
	finalParams = append(finalParams, staticWhereParams...)
	if len(groupByCombinationsForGBT) != 0 {
		whereConditionForGBT, whereParams := buildWhereConditionForGBTForLinkedin(groupByCombinationsForGBT)
		finalFilterStatement += (" AND (" + whereConditionForGBT + ")")
		finalParams = append(finalParams, whereParams...)
	}

	resultSQLStatement := selectQuery + fromLinkedinDocuments + staticWhereStatementForLinkedin + whereConditionForFilters
	if len(groupByStatement) != 0 {
		resultSQLStatement += "GROUP BY " + groupByStatement
	}
	resultSQLStatement += " " + orderByQuery + limitString + ";"
	return resultSQLStatement, finalParams, responseSelectKeys, responseSelectMetrics
}
func buildWhereConditionForGBTForLinkedin(groupByCombinations []map[string]interface{}) (string, []interface{}) {
	whereConditionForGBT := ""
	params := make([]interface{}, 0)
	filterStringSmartPropertiesCampaign := "campaign.properties"
	filterStringSmartPropertiesAdGroup := "ad_group.properties"
	for _, groupByCombination := range groupByCombinations {
		whereConditionForEachCombination := ""
		for dimension, value := range groupByCombination {
			filterString := ""
			if strings.HasPrefix(dimension, model.CampaignPrefix) {
				key := fmt.Sprintf(`%s:%s`, "campaign_group", strings.TrimPrefix(dimension, model.CampaignPrefix))
				currentFilterKey, isPresent := objectToValueInLinkedinFiltersMappingWithLinkedinDocuments[key]
				if isPresent {
					filterString = currentFilterKey
				} else {
					filterString = fmt.Sprintf("JSON_EXTRACT_STRING(%s, '%s')", filterStringSmartPropertiesCampaign, strings.TrimPrefix(dimension, model.CampaignPrefix))
				}
			} else if strings.HasPrefix(dimension, model.AdgroupPrefix) {
				key := fmt.Sprintf(`%s:%s`, "campaign", strings.TrimPrefix(dimension, model.AdgroupPrefix))
				currentFilterKey, isPresent := objectToValueInLinkedinFiltersMappingWithLinkedinDocuments[key]
				if isPresent {
					filterString = currentFilterKey
				} else {
					filterString = fmt.Sprintf("JSON_EXTRACT_STRING(%s, '%s')", filterStringSmartPropertiesAdGroup, strings.TrimPrefix(dimension, model.AdgroupPrefix))
				}
			} else {
				key := fmt.Sprintf(`%s:%s`, "creative", strings.TrimPrefix(dimension, model.KeywordPrefix))
				currentFilterKey := objectToValueInLinkedinFiltersMappingWithLinkedinDocuments[key]
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

func getLinkedinFiltersWhereStatement(filters []model.ChannelFilterV1) string {
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
		currentFilterStatement = fmt.Sprintf("%s %s '%s' ", objectToValueInLinkedinFiltersMapping[filter.Object+":"+filter.Property], filterOperator, filterValue)
		if index == 0 {
			resultStatement = " AND " + currentFilterStatement
		} else {
			resultStatement = fmt.Sprintf("%s %s %s ", resultStatement, filter.LogicalOp, currentFilterStatement)
		}
	}
	return resultStatement
}
func getLinkedinFiltersWhereStatementWithSmartProperty(filters []model.ChannelFilterV1, smartPropertyCampaignGroupBys []model.ChannelGroupBy, smartPropertyAdGroupGroupBys []model.ChannelGroupBy) string {
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
		_, isPresent := Const.SmartPropertyReservedNames[filter.Property]
		if isPresent {
			currentFilterStatement = fmt.Sprintf("%s %s '%s' ", objectToValueInLinkedinFiltersMapping[filter.Object+":"+filter.Property], filterOperator, filterValue)
			if index == 0 {
				resultStatement = " AND " + currentFilterStatement
			} else {
				resultStatement = fmt.Sprintf("%s %s %s ", resultStatement, filter.LogicalOp, currentFilterStatement)
			}
		} else {
			currentFilterStatement = fmt.Sprintf("JSON_EXTRACT_STRING(%s.properties, '%s') %s '%s'", model.LinkedinObjectMapForSmartProperty[filter.Object], filter.Property, filterOperator, filterValue)
			if index == 0 {
				resultStatement = fmt.Sprintf("(%s", currentFilterStatement)
			} else {
				resultStatement = fmt.Sprintf("%s %s %s", resultStatement, filter.LogicalOp, currentFilterStatement)
			}
			if filter.Object == "campaign_group" {
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
		if objectName == model.LinkedinCreative {
			return model.LinkedinCreative
		}
	}

	for _, objectName := range objectNames {
		if objectName == model.LinkedinCampaign {
			return model.LinkedinCampaign
		}
	}

	for _, objectName := range objectNames {
		if objectName == model.LinkedinCampaignGroup {
			return model.LinkedinCampaignGroup
		}
	}

	return model.LinkedinCampaignGroup
}

// Since we dont have a way to store raw format, we are going with the approach of joins on query.
func (store *MemSQL) GetLatestMetaForLinkedinForGivenDays(projectID uint64, days int) ([]model.ChannelDocumentsWithFields, []model.ChannelDocumentsWithFields) {
	db := C.GetServices().Db

	channelDocumentsCampaign := make([]model.ChannelDocumentsWithFields, 0, 0)
	channelDocumentsAdGroup := make([]model.ChannelDocumentsWithFields, 0, 0)

	projectSetting, errCode := store.GetProjectSetting(projectID)
	if errCode != http.StatusFound {
		log.Error("Failed to get project settings")
		return channelDocumentsCampaign, channelDocumentsAdGroup
	}
	if projectSetting.IntLinkedinAdAccount == "" {
		log.Error("Failed to get custtomer account ids")
		return channelDocumentsCampaign, channelDocumentsAdGroup
	}
	customerAccountIDs := strings.Split(projectSetting.IntLinkedinAdAccount, ",")

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

	err = db.Raw(linkedinAdGroupMetadataFetchQueryStr, linkedinDocumentTypeAlias["campaign"], projectID, from, to, customerAccountIDs,
		linkedinDocumentTypeAlias["campaign"], projectID, from, to, customerAccountIDs, linkedinDocumentTypeAlias["campaign_group"],
		projectID, from, to, customerAccountIDs, linkedinDocumentTypeAlias["campaign_group"], projectID,
		from, to, customerAccountIDs).Find(&channelDocumentsAdGroup).Error
	if err != nil {
		errString := fmt.Sprintf("failed to get last %d ad_group meta for Linkedin", days)
		log.Error(errString)
		return channelDocumentsCampaign, channelDocumentsAdGroup
	}

	err = db.Raw(linkedinCampaignMetadataFetchQueryStr, linkedinDocumentTypeAlias["campaign_group"], projectID, from, to,
		customerAccountIDs, linkedinDocumentTypeAlias["campaign_group"], projectID, from, to,
		customerAccountIDs).Find(&channelDocumentsCampaign).Error
	if err != nil {
		errString := fmt.Sprintf("failed to get last %d campaign meta for Linkedin", days)
		log.Error(errString)
		return channelDocumentsCampaign, channelDocumentsAdGroup
	}

	return channelDocumentsCampaign, channelDocumentsAdGroup
}

func (store *MemSQL) DeleteLinkedinIntegration(projectID uint64) (int, error) {
	db := C.GetServices().Db
	updateValues := make(map[string]interface{})
	updateValues["int_linkedin_ad_account"] = nil
	updateValues["int_linkedin_access_token"] = nil
	updateValues["int_linkedin_refresh_token"] = nil
	updateValues["int_linkedin_refresh_token_expiry"] = nil
	updateValues["int_linkedin_access_token_expiry"] = nil

	err := db.Model(&model.ProjectSetting{}).Where("project_id = ?", projectID).Update(updateValues).Error
	if err != nil {
		return http.StatusInternalServerError, err
	}
	return http.StatusOK, nil
}
