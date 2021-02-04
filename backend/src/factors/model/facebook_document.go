package model

import (
	"errors"
	C "factors/config"
	U "factors/util"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

// FacebookDocument ...
type FacebookDocument struct {
	ProjectID           uint64          `gorm:"primary_key:true;auto_increment:false" json:"project_id"`
	CustomerAdAccountID string          `gorm:"primary_key:true;auto_increment:false" json:"customer_ad_account_id"`
	Platform            string          `gorm:"primary_key:true;auto_increment:false" json:"platform"`
	TypeAlias           string          `gorm:"-" json:"type_alias"`
	Type                int             `gorm:"primary_key:true;auto_increment:false" json:"-"`
	Timestamp           int64           `gorm:"primary_key:true;auto_increment:false" json:"timestamp"`
	ID                  string          `gorm:"primary_key:true;auto_increment:false" json:"id"`
	CampaignID          int64           `json:"-"`
	AdSetID             int64           `json:"-"`
	AdID                int64           `json:"-"`
	Value               *postgres.Jsonb `json:"value"`
	CreatedAt           time.Time       `json:"created_at"`
	UpdatedAt           time.Time       `json:"updated_at"`
}

const (
	facebookCampaign     = "campaign"
	facebookAdSet        = "ad_set"
	facebookAd           = "ad"
	facebookStringColumn = "facebook"
)

var facebookDocumentTypeAlias = map[string]int{
	"ad_account":        7,
	"campaign":          1,
	"ad":                2,
	"ad_set":            3,
	"ad_insights":       4,
	"campaign_insights": 5,
	"ad_set_insights":   6,
}

var objectAndPropertyToValueInFacebookReportsMapping = map[string]string{
	"campaign:id": "campaign_id",
	"ad_set:id":   "ad_set_id",
	"ad:id":       "id",
}

var objectToValueInFacebookJobsMapping = map[string]string{
	"campaign:name": "name",
	"ad_set:name":   "name",
	"campaign:id":   "campaign_id",
	"ad_set:id":     "ad_set_id",
	"ad:id":         "ad_id",
}

var mapOfTypeToFacebookJobCTEAlias = map[string]string{
	"ad_set":   "ad_set_cte",
	"campaign": "campaign_cte",
}

// TODO check
var facebookMetricsToAggregatesInReportsMapping = map[string]string{
	"impressions": "SUM((value->>'impressions')::float)",
	"clicks":      "SUM((value->>'clicks')::float)",
	"spend":       "SUM((value->>'spend')::float)",
	"conversions": "SUM((value->>'conversions')::float)",
	// "cost_per_click": "average_cost",
	// "conversion_rate": "conversion_rate"
}

var facebookMetricsToOperation = map[string]string{
	"impressions": "sum",
	"clicks":      "sum",
	"spend":       "sum",
	"conversions": "sum",
}

var facebookExternalRepresentationToInternalRepresentation = map[string]string{
	"name":        "name",
	"id":          "id",
	"impressions": "impressions",
	"clicks":      "clicks",
	"spend":       "spend",
	"conversion":  "conversions",
	"campaign":    "campaign",
	"ad_group":    "ad_set",
	"ad":          "ad",
}

var facebookInternalRepresentationToExternalRepresentation = map[string]string{
	"impressions": "impressions",
	"clicks":      "clicks",
	"spend":       "spend",
	"conversions": "conversion",
}

const platform = "platform"

var errorEmptyFacebookDocument = errors.New("empty facebook document")

const errorDuplicateFacebookDocument = "pq: duplicate key value violates unique constraint \"facebook_documents_pkey\""

const facebookFilterQueryStr = "SELECT DISTINCT(value->>?) as filter_value FROM facebook_documents WHERE project_id = ? AND" +
	" " + "customer_ad_account_id = ? AND type = ? AND value->>? IS NOT NULL LIMIT 5000"

const fromFacebooksDocument = " FROM facebook_documents "

const staticWhereStatementForFacebook = "WHERE project_id = ? AND customer_ad_account_id IN ( ? ) AND type = ? AND timestamp between ? AND ? "

func isDuplicateFacebookDocumentError(err error) bool {
	return err.Error() == errorDuplicateFacebookDocument
}

func getFacebookID(valueJSON *postgres.Jsonb) (string, error) {

	valueMap, err := U.DecodePostgresJsonb(valueJSON)
	if err != nil {
		return "", err
	}
	id, exists := (*valueMap)["id"]
	if !exists {
		return "", fmt.Errorf("id field %s does not exist", id)
	}

	if id == nil {
		return "", fmt.Errorf("id field %s has empty value", id)
	}

	idStr, err := U.GetValueAsString(id)
	if err != nil {
		return "", err
	}

	// ID as string always.
	return idStr, nil
}

// CreateFacebookDocument ...
func CreateFacebookDocument(projectID uint64, document *FacebookDocument) int {
	logCtx := log.WithField("customer_acc_id", document.CustomerAdAccountID).WithField(
		"project_id", document.ProjectID)

	if document.CustomerAdAccountID == "" || document.TypeAlias == "" {
		logCtx.Error("Invalid facebook document.")
		return http.StatusBadRequest
	}
	if document.ProjectID == 0 || document.Timestamp == 0 || document.Platform == "" {
		logCtx.Error("Invalid facebook document.")
		return http.StatusBadRequest
	}

	logCtx = logCtx.WithField("type_alias", document.TypeAlias)
	docType, docTypeExists := facebookDocumentTypeAlias[document.TypeAlias]
	if !docTypeExists {
		logCtx.Error("Invalid type alias.")
		return http.StatusBadRequest
	}
	document.Type = docType

	campaignIDValue, adSetID, adID, error := getFacebookHierarchyColumnsByType(docType, document.Value)
	if error != nil {
		logCtx.Error("Invalid docType alias.")
		return http.StatusBadRequest
	}
	document.CampaignID = campaignIDValue
	document.AdSetID = adSetID
	document.AdID = adID

	db := C.GetServices().Db
	err := db.Create(&document).Error
	if err != nil {
		if isDuplicateFacebookDocumentError(err) {
			logCtx.WithError(err).WithField("id", document.ID).WithField("platform", document.Platform).Error(
				"Failed to create an facebook doc. Duplicate.")
			return http.StatusConflict
		}
		logCtx.WithError(err).WithField("id", document.ID).WithField("platform", document.Platform).Error(
			"Failed to create an facebook doc. Continued inserting other docs.")
		return http.StatusInternalServerError
	}

	return http.StatusCreated
}

func getFacebookHierarchyColumnsByType(docType int, valueJSON *postgres.Jsonb) (int64, int64, int64, error) {
	if docType > len(facebookDocumentTypeAlias) {
		return 0, 0, 0, errors.New("invalid document type")
	}

	valueMap, err := U.DecodePostgresJsonb(valueJSON)
	if err != nil {
		return 0, 0, 0, err
	}

	if len(*valueMap) == 0 {
		return 0, 0, 0, errorEmptyFacebookDocument
	}
	switch docType {
	case 1:
		return U.GetInt64FromMapOfInterface(*valueMap, "id", 0), 0, 0, nil
	case 2:
		return U.GetInt64FromMapOfInterface(*valueMap, "campaign_id", 0), U.GetInt64FromMapOfInterface(*valueMap, "adset_id", 0), U.GetInt64FromMapOfInterface(*valueMap, "id", 0), nil
	case 3:
		return U.GetInt64FromMapOfInterface(*valueMap, "campaign_id", 0), U.GetInt64FromMapOfInterface(*valueMap, "id", 0), 0, nil
	case 4, 5, 6:
		return U.GetInt64FromMapOfInterface(*valueMap, "campaign_id", 0), U.GetInt64FromMapOfInterface(*valueMap, "adset_id", 0), U.GetInt64FromMapOfInterface(*valueMap, "ad_id", 0), nil
	default:
		return 0, 0, 0, nil
	}
}

// FacebookLastSyncInfo ...
type FacebookLastSyncInfo struct {
	ProjectID           uint64 `json:"project_id"`
	CustomerAdAccountID string `json:"customer_ad_acc_id"`
	Platform            string `json:"platform"`
	DocumentType        int    `json:"-"`
	DocumentTypeAlias   string `json:"type_alias"`
	LastTimestamp       int64  `json:"last_timestamp"`
}

// FacebookLastSyncInfoPayload ...
type FacebookLastSyncInfoPayload struct {
	ProjectId           string `json:"project_id"`
	CustomerAdAccountId string `json:"account_id"`
}

func getFacebookDocumentTypeAliasByType() map[int]string {
	documentTypeMap := make(map[int]string, 0)
	for alias, typ := range facebookDocumentTypeAlias {
		documentTypeMap[typ] = alias
	}

	return documentTypeMap
}

// @TODO Kark v1
func buildFbChannelConfig() *ChannelConfigResult {
	properties := buildProperties(allChannelsPropertyToRelated)
	objectsAndProperties := buildObjectsAndProperties(properties, objectsForAllChannels)

	return &ChannelConfigResult{
		SelectMetrics:        selectableMetricsForAllChannels,
		ObjectsAndProperties: objectsAndProperties,
	}
}

// GetFacebookFilterValues - @TODO Kark v1
func GetFacebookFilterValues(projectID uint64, requestFilterObject string, requestFilterProperty string, reqID string) ([]interface{}, int) {
	facebookInternalFilterProperty, docType, err := getFilterRelatedInformationForFacebook(requestFilterObject, requestFilterProperty)
	if err != http.StatusOK {
		return make([]interface{}, 0, 0), http.StatusBadRequest
	}
	filterValues, errCode := getFacebookFilterValuesByType(projectID, docType, facebookInternalFilterProperty, reqID)
	if errCode != http.StatusFound {
		return []interface{}{}, http.StatusInternalServerError
	}

	return filterValues, http.StatusFound
}

// GetFacebookSQLQueryAndParametersForFilterValues - @TODO Kark v1
func GetFacebookSQLQueryAndParametersForFilterValues(projectID uint64, requestFilterObject string, requestFilterProperty string) (string, []interface{}, int) {
	facebookInternalFilterProperty, docType, err := getFilterRelatedInformationForFacebook(requestFilterObject, requestFilterProperty)
	if err != http.StatusOK {
		return "", make([]interface{}, 0, 0), http.StatusBadRequest
	}
	projectSetting, errCode := GetProjectSetting(projectID)
	if errCode != http.StatusFound {
		return "", []interface{}{}, http.StatusInternalServerError
	}
	customerAccountID := projectSetting.IntFacebookAdAccount
	params := []interface{}{facebookInternalFilterProperty, projectID, customerAccountID, docType, facebookInternalFilterProperty}

	return "(" + facebookFilterQueryStr + ")", params, http.StatusFound
}

func getFilterRelatedInformationForFacebook(requestFilterObject string, requestFilterProperty string) (string, int, int) {
	facebookInternalFilterObject, isPresent := facebookExternalRepresentationToInternalRepresentation[requestFilterObject]
	if !isPresent {
		log.Error("Invalid facebook filter object.")
		return "", 0, http.StatusBadRequest
	}
	facebookInternalFilterProperty, isPresent := facebookExternalRepresentationToInternalRepresentation[requestFilterProperty]
	if !isPresent {
		log.Error("Invalid facebook filter property.")
		return "", 0, http.StatusBadRequest
	}
	docType := facebookDocumentTypeAlias[facebookInternalFilterObject]

	return facebookInternalFilterProperty, docType, http.StatusOK
}

// @TODO Kark v1
// TODO: Check the value of filterObject. It should be adset
func getFacebookFilterPropertyKeyByType(filterObject string, filterProperty string) (string, error) {
	property, isPropertyPresent := facebookExternalRepresentationToInternalRepresentation[filterProperty]

	if !isPropertyPresent {
		return "", errors.New("no filter key found for document type")
	}
	return property, nil
}

// @TODO Kark v1
func getFacebookFilterValuesByType(projectID uint64, docType int, property string, reqID string) ([]interface{}, int) {
	logCtx := log.WithField("req_id", reqID)
	projectSetting, errCode := GetProjectSetting(projectID)
	if errCode != http.StatusFound {
		logCtx.Error("Failed to fetch Project Setting in facebook filter values.")
		return []interface{}{}, http.StatusInternalServerError
	}
	customerAccountID := projectSetting.IntFacebookAdAccount

	logCtx = log.WithField("project_id", projectID).WithField("doc_type", docType).WithField("req_id", reqID)
	params := []interface{}{property, projectID, customerAccountID, docType, property}
	_, resultRows, _ := ExecuteSQL(facebookFilterQueryStr, params, logCtx)

	return Convert2DArrayTo1DArray(resultRows), http.StatusFound
}

// ExecuteFacebookChannelQueryV1 - @Kark TODO v1
// In this flow, Job represents the meta data associated with particular object type. Reports represent data with metrics and few filters.
// TODO - Duplicate code/flow in facebook and adwords.
func ExecuteFacebookChannelQueryV1(projectID uint64, query *ChannelQueryV1, reqID string) ([]string, [][]interface{}, error) {
	fetchSource := false
	logCtx := log.WithField("xreq_id", reqID)
	sql, params, _, err := GetSQLQueryAndParametersForFacebookQueryV1(projectID, query, reqID, fetchSource)
	if err != nil {
		return make([]string, 0, 0), make([][]interface{}, 0, 0), err
	}
	_, resultMetrics, err := ExecuteSQL(sql, params, logCtx)
	columns := buildColumns(query, fetchSource)
	return columns, resultMetrics, err
}

// GetSQLQueryAndParametersForFacebookQueryV1 ...
func GetSQLQueryAndParametersForFacebookQueryV1(projectID uint64, query *ChannelQueryV1, reqID string, fetchSource bool) (string, []interface{}, []string, error) {
	var selectMetrics []string
	var sql string
	var params []interface{}
	transformedQuery, customerAccountID, err := transFormRequestFieldsAndFetchRequiredFieldsForFacebook(projectID, *query, reqID)
	if err != nil {
		return "", make([]interface{}, 0, 0), make([]string, 0, 0), err
	}
	if hasAllIDsOnlyInGroupBy(transformedQuery) {
		sql, params, selectMetrics, err := buildFacebookSimpleQueryV1(transformedQuery, projectID, customerAccountID, reqID, fetchSource)
		if err != nil {
			return "", make([]interface{}, 0, 0), make([]string, 0, 0), err
		}
		return sql, params, selectMetrics, nil
	}
	sql, params, selectMetrics = buildFacebookComplexQueryV1(transformedQuery, projectID, customerAccountID, fetchSource)
	return sql, params, selectMetrics, nil
}

func transFormRequestFieldsAndFetchRequiredFieldsForFacebook(projectID uint64, query ChannelQueryV1, reqID string) (*ChannelQueryV1, string, error) {
	var transformedQuery ChannelQueryV1
	var err error
	logCtx := log.WithField("req_id", reqID)
	projectSetting, errCode := GetProjectSetting(projectID)
	if errCode != http.StatusFound {
		return &ChannelQueryV1{}, "", errors.New("Project setting not found")
	}
	customerAccountID := projectSetting.IntFacebookAdAccount

	transformedQuery, err = convertFromRequestToFacebookSpecificRepresentation(query)
	if err != nil {
		logCtx.Warn("Request failed in validation: ", err)
		return &ChannelQueryV1{}, "", err
	}
	return &transformedQuery, customerAccountID, nil
}

// @Kark TODO v1
// Currently, this relies on assumption of Object across different filterObjects. Change when we need robust.
func convertFromRequestToFacebookSpecificRepresentation(query ChannelQueryV1) (ChannelQueryV1, error) {
	var transformedQuery ChannelQueryV1
	var err1, err2, err3 error
	transformedQuery.SelectMetrics, err1 = getFacebookSpecificMetrics(query.SelectMetrics)
	transformedQuery.Filters, err2 = getFacebookSpecificFilters(query.Filters)
	transformedQuery.GroupBy, err3 = getFacebookSpecificGroupBy(query.GroupBy)
	if err1 != nil {
		return query, err1
	}
	if err2 != nil {
		return query, err2
	}
	if err3 != nil {
		return query, err3
	}
	transformedQuery.From = getAdwordsDateOnlyTimestampInInt64(query.From)
	transformedQuery.To = getAdwordsDateOnlyTimestampInInt64(query.To)
	transformedQuery.Timezone = query.Timezone
	transformedQuery.GroupByTimestamp = query.GroupByTimestamp

	return transformedQuery, nil
}

// @Kark TODO v1
func getFacebookSpecificMetrics(requestSelectMetrics []string) ([]string, error) {
	resultMetrics := make([]string, 0, 0)
	for _, requestMetric := range requestSelectMetrics {
		metric, isPresent := facebookExternalRepresentationToInternalRepresentation[requestMetric]
		if !isPresent {
			return make([]string, 0, 0), errors.New("Invalid metric found for document type")
		}
		resultMetrics = append(resultMetrics, metric)
	}
	return resultMetrics, nil
}

// @Kark TODO v1
func getFacebookSpecificFilters(requestFilters []FilterV1) ([]FilterV1, error) {
	resultFilters := make([]FilterV1, 0, 0)
	for _, requestFilter := range requestFilters {
		var resultFilter FilterV1
		filterObject, isPresent := facebookExternalRepresentationToInternalRepresentation[requestFilter.Object]
		if !isPresent {
			return make([]FilterV1, 0, 0), errors.New("Invalid filter key found for document type")
		}
		resultFilter = requestFilter
		resultFilter.Object = filterObject
		resultFilters = append(resultFilters, resultFilter)
	}
	return resultFilters, nil
}

// @Kark TODO v1
func getFacebookSpecificGroupBy(requestGroupBys []GroupBy) ([]GroupBy, error) {
	sortedGroupBys := make([]GroupBy, 0, 0)
	for _, groupBy := range requestGroupBys {
		if groupBy.Object == CAFilterCampaign {
			sortedGroupBys = append(sortedGroupBys, groupBy)
		}
	}

	for _, groupBy := range requestGroupBys {
		if groupBy.Object == CAFilterAdGroup {
			sortedGroupBys = append(sortedGroupBys, groupBy)
		}
	}

	for _, groupBy := range requestGroupBys {
		if groupBy.Object == CAFilterAd {
			sortedGroupBys = append(sortedGroupBys, groupBy)
		}
	}

	resultGroupBys := make([]GroupBy, 0, 0)
	for _, requestGroupBy := range sortedGroupBys {
		var resultGroupBy GroupBy
		groupByObject, isPresent := facebookExternalRepresentationToInternalRepresentation[requestGroupBy.Object]
		if !isPresent {
			return make([]GroupBy, 0, 0), errors.New("Invalid groupby key found for document type")
		}
		resultGroupBy = requestGroupBy
		resultGroupBy.Object = groupByObject
		resultGroupBys = append(resultGroupBys, resultGroupBy)
	}
	return resultGroupBys, nil
}

/*
SELECT ad_set_id, date_trunc('day', to_timestamp(timestamp::text, 'YYYYMMDD') AT TIME ZONE 'UTC') as datetime,
SUM((value->>'impressions')::float) as impressions, SUM((value->>'spend')::float) as spend, SUM((value->>'conversions')::float) as conversion
FROM facebook_documents WHERE project_id = '8' AND customer_ad_account_id IN ( 'act_2737160626567083' ) AND type = '6' AND
timestamp between '20201130' AND '20201205' GROUP BY ad_set_id, datetime ORDER BY impressions DESC, spend DESC, conversion DESC LIMIT 2500 ;
*/
// TODO Error add
func buildFacebookSimpleQueryV1(query *ChannelQueryV1, projectID uint64, customerAccountID string, reqID string, fetchSource bool) (string, []interface{}, []string, error) {
	campaignIDs, adSetIDs, err := getIDsFromFacebookJob(query, projectID, customerAccountID, reqID)
	if err != nil {
		return "", make([]interface{}, 0), make([]string, 0), err
	}
	lowestHierarchyLevel := getLowestHierarchyLevelForFacebook(query)
	lowestHierarchyReportLevel := lowestHierarchyLevel + "_insights"
	sql, params, selectMetrics := getSQLAndParamsFromFacebookReports(query, projectID, query.From, query.To, customerAccountID, facebookDocumentTypeAlias[lowestHierarchyReportLevel],
		campaignIDs, adSetIDs, reqID, fetchSource)
	return sql, params, selectMetrics, nil
}

// @TODO Kark v1
func getIDsFromFacebookJob(query *ChannelQueryV1, projectID uint64, facebookAccountIDs string, reqID string) ([]int, []int, error) {
	var err error
	campaignsFilters, adSetFilters, _ := splitFiltersByObjectTypeForFacebook(query)
	campaignIDs, err := getIDsByPropertyOnFacebookJob(projectID, query.From, query.To, facebookAccountIDs, facebookDocumentTypeAlias["campaign"], campaignsFilters, reqID)
	if err != nil {
		return make([]int, 0), make([]int, 0), err
	}
	adSetIDs, err := getIDsByPropertyOnFacebookJob(projectID, query.From, query.To, facebookAccountIDs, facebookDocumentTypeAlias["ad_set"], adSetFilters, reqID)
	if err != nil {
		return make([]int, 0), make([]int, 0), err
	}
	// adIDs := getAdIDsByPropertyOnJob(adFilters)
	return campaignIDs, adSetIDs, nil
}

// @TODO Kark v1
func getIDsByPropertyOnFacebookJob(projectID uint64, from, to int64, facebookAccountIDs string, type1 int, filters []FilterV1, reqID string) ([]int, error) {
	logCtx := log.WithField("req_id", reqID)
	db := C.GetServices().Db
	if len(filters) == 0 {
		return []int{}, nil
	}
	sqlParams := make([]interface{}, 0)
	customerAccountIDs := strings.Split(facebookAccountIDs, ",")

	selectStatement := "SELECT value->>'id'" + fromFacebooksDocument
	groupByStatement := "GROUP BY value->>'id'"

	sqlParams = append(sqlParams, projectID, customerAccountIDs, type1, from, to)
	filterPropertiesStatement, filterPropertiesParams := getFilterPropertiesForFacebookJobStatementAndParams(filters)
	completeFiltersStatement := staticWhereStatementForFacebook
	if len(filterPropertiesStatement) != 0 {
		completeFiltersStatement += "AND " + filterPropertiesStatement + " "
		sqlParams = append(sqlParams, filterPropertiesParams...)
	}

	resultSQLStatement := selectStatement + completeFiltersStatement + groupByStatement + ";"

	rows, err := db.Raw(resultSQLStatement, sqlParams...).Rows()
	if err != nil {
		logCtx.WithError(err).Error("SQL Query failed")
		return nil, err
	}
	defer rows.Close()
	ids := make([]int, 0, 0)
	for rows.Next() {
		var id int
		if err = rows.Scan(&id); err != nil {
			logCtx.WithError(err).Error("SQL Parse failed")
			continue
		}
		ids = append(ids, id)
	}

	return ids, nil
}

// @TODO Kark v1
// Complexity consideration - Having at max of 20 filters and 20 group by should be fine.
// change algo/strategy the filters and group by goes beyond 100.
func getLowestHierarchyLevelForFacebook(query *ChannelQueryV1) string {
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
		if objectName == facebookAd {
			return facebookAd
		}
	}

	for _, objectName := range objectNames {
		if objectName == facebookAdSet {
			return facebookAdSet
		}
	}

	for _, objectName := range objectNames {
		if objectName == facebookCampaign {
			return facebookCampaign
		}
	}
	return facebookCampaign
}

func getSQLAndParamsFromFacebookReports(query *ChannelQueryV1, projectID uint64, from, to int64, facebookAccountIDs string,
	docType int, campaignIDs []int, adSetIDs []int, reqID string, fetchSource bool) (string, []interface{}, []string) {
	customerAccountIDs := strings.Split(facebookAccountIDs, ",")
	selectQuery := "SELECT "
	selectMetrics := make([]string, 0, 0)
	isGroupByTimestamp := query.GetGroupByTimestamp() != ""
	groupByStatement := ""
	groupByKeysWithoutTimestamp := make([]string, 0, 0)
	selectKeys := make([]string, 0, 0)
	responseSelectMetrics := make([]string, 0, 0)

	for _, groupBy := range query.GroupBy {
		key := groupBy.Object + ":" + groupBy.Property
		groupByKeysWithoutTimestamp = append(groupByKeysWithoutTimestamp, objectAndPropertyToValueInFacebookReportsMapping[key])
	}

	if isGroupByTimestamp {
		groupByStatement = joinWithComma(append(groupByKeysWithoutTimestamp, AliasDateTime)...)
	} else {
		groupByStatement = joinWithComma(groupByKeysWithoutTimestamp...)
	}

	if fetchSource {
		selectKeys = append(selectKeys, fmt.Sprintf("'%s' as %s", facebookStringColumn, source))
	}
	selectKeys = append(selectKeys, groupByKeysWithoutTimestamp...)
	if isGroupByTimestamp {
		selectKeys = append(selectKeys, fmt.Sprintf("%s as %s",
			getSelectTimestampByTypeForChannels(query.GetGroupByTimestamp(), query.Timezone), AliasDateTime))
	}

	for _, selectMetric := range query.SelectMetrics {
		value := fmt.Sprintf("%s as %s", facebookMetricsToAggregatesInReportsMapping[selectMetric], facebookInternalRepresentationToExternalRepresentation[selectMetric])
		selectMetrics = append(selectMetrics, value)

		value = facebookInternalRepresentationToExternalRepresentation[selectMetric]
		responseSelectMetrics = append(responseSelectMetrics, value)
	}

	selectQuery += joinWithComma(append(selectKeys, selectMetrics...)...)
	orderByQuery := "ORDER BY " + getOrderByClause(responseSelectMetrics)

	whereConditionForIDs := ""
	if len(campaignIDs) != 0 {
		campaignString := ""
		for _, campaignID := range campaignIDs {
			campaignString += strconv.Itoa(campaignID) + ","
		}
		campaignString = campaignString[:len(campaignString)-1]
		whereConditionForIDs += "AND campaign_id IN " + "(" + campaignString + ") "
	}
	if len(adSetIDs) != 0 {
		adSetString := ""
		for _, adSetID := range adSetIDs {
			adSetString += strconv.Itoa(adSetID) + ","
		}
		adSetString = adSetString[:len(adSetString)-1]
		whereConditionForIDs += "AND adset_id IN " + "(" + adSetString + ") "
	}

	resultSQLStatement := selectQuery + fromFacebooksDocument + staticWhereStatementForFacebook + whereConditionForIDs
	if len(groupByStatement) != 0 {
		resultSQLStatement += "GROUP BY " + groupByStatement
	}
	resultSQLStatement += " " + orderByQuery + channeAnalyticsLimit + ";"
	staticWhereParams := []interface{}{projectID, customerAccountIDs, docType, from, to}
	return resultSQLStatement, staticWhereParams, responseSelectMetrics
}

/*
 WITH reports_cte as (SELECT ad_set_id, date_trunc('day', to_timestamp(timestamp::text, 'YYYYMMDD') AT TIME ZONE 'UTC') as datetime,
 SUM((value->>'impressions')::float) as impressions, SUM((value->>'spend')::float) as spend, SUM((value->>'conversions')::float) as conversions
 FROM facebook_documents WHERE project_id = '8' AND customer_ad_account_id IN ( 'act_2737160626567083' ) AND type = '6'
 AND timestamp between '20201130' AND '20201205'  GROUP BY ad_set_id, datetime ),
 ad_set_cte as ( SELECT distinct_id.ad_set_id as ad_set_id, value->>'name' as name FROM (SELECT ad_set_id , max(timestamp) as timestamp
 FROM facebook_documents WHERE project_id = '8' AND customer_ad_account_id IN ( 'act_2737160626567083' ) AND type = '3' AND
 timestamp between '20201130' AND '20201205' GROUP BY ad_set_id) as distinct_id INNER JOIN
 (SELECT * FROM facebook_documents WHERE project_id = '8' AND customer_ad_account_id IN ( 'act_2737160626567083' ) AND type = '3'
 AND timestamp between '20201130' AND '20201205' ) as JobRecords
 ON distinct_id.ad_set_id = JobRecords.ad_set_id AND distinct_id.timestamp = JobRecords.timestamp)
 SELECT ad_set_cte.name, reports_cte.datetime, sum(reports_cte.impressions) as impressions, sum(reports_cte.spend) as spend,
 sum(reports_cte.conversions) as conversion from reports_cte  INNER JOIN ad_set_cte ON reports_cte.ad_set_id = ad_set_cte.ad_set_id
 GROUP BY ad_set_cte.name, datetime ORDER BY impressions DESC, spend DESC, conversion DESC LIMIT 2500 ;
*/
func buildFacebookComplexQueryV1(query *ChannelQueryV1, projectID uint64, customerAccountID string, fetchSource bool) (string, []interface{}, []string) {
	selectQuery := "SELECT "
	selectMetrics := make([]string, 0, 0)
	isGroupByTimestamp := query.GetGroupByTimestamp() != ""
	groupByStatement := ""
	groupByKeysWithoutTimestamp := make([]string, 0, 0)
	lowestHierarchyLevel := getLowestHierarchyLevelForFacebook(query)
	lowestHierarchyReportLevel := lowestHierarchyLevel + "_insights"
	selectKeys := make([]string, 0, 0)
	responseSelectMetrics := make([]string, 0, 0)

	reportCTE, reportCTEAlias, reportSelectMetrics, reportCTEJoinFields, reportParams := getCTEAndParamsForFacebookReportComplexStrategy(query, projectID, customerAccountID, facebookDocumentTypeAlias[lowestHierarchyReportLevel])
	jobsCTE, jobsCTEAliases, jobCTEJoinFields, jobsParams := getCTEAndParamsForFacebookJobsComplexStrategy(query, projectID, customerAccountID)

	completeWithClause := reportCTE
	params := make([]interface{}, 0, 0)
	params = append(params, reportParams...)

	params = append(params, jobsParams...)
	for _, jobCTE := range jobsCTE {
		completeWithClause += jobCTE
	}
	completeWithClause = completeWithClause[:len(completeWithClause)-2]

	for _, groupBy := range query.GroupBy {
		key := groupBy.Object + ":" + groupBy.Property
		value := mapOfTypeToFacebookJobCTEAlias[groupBy.Object] + "." + objectToValueInFacebookJobsMapping[key]
		groupByKeysWithoutTimestamp = append(groupByKeysWithoutTimestamp, value)
	}

	if isGroupByTimestamp {
		groupByStatement = joinWithComma(append(groupByKeysWithoutTimestamp, AliasDateTime)...)
	} else {
		groupByStatement = joinWithComma(groupByKeysWithoutTimestamp...)
	}

	if fetchSource {
		selectKeys = append(selectKeys, fmt.Sprintf("'%s' as %s", facebookStringColumn, source))
	}
	selectKeys = append(selectKeys, groupByKeysWithoutTimestamp...)
	if isGroupByTimestamp {
		selectKeys = append(selectKeys, reportCTEAlias+"."+AliasDateTime)
	}

	for _, selectMetric := range reportSelectMetrics {
		value := fmt.Sprintf("%s(%s.%s) as %s", facebookMetricsToOperation[selectMetric], reportCTEAlias, selectMetric, facebookInternalRepresentationToExternalRepresentation[selectMetric])
		selectMetrics = append(selectMetrics, value)

		value = facebookInternalRepresentationToExternalRepresentation[selectMetric]
		responseSelectMetrics = append(responseSelectMetrics, value)
	}
	selectQuery += joinWithComma(append(selectKeys, selectMetrics...)...)
	orderByQuery := "ORDER BY " + getOrderByClause(responseSelectMetrics)

	completeInnerJoin := " from " + reportCTEAlias + " "
	for index, jobsCTEAlias := range jobsCTEAliases {
		completeInnerJoin += innerJoinClause + jobsCTEAlias + " ON " + reportCTEAlias + "." + reportCTEJoinFields[index] + " = " + jobsCTEAlias + "." + jobCTEJoinFields[index] + " "
	}
	completeInnerJoin = completeInnerJoin + " "

	resultSQLStatement := completeWithClause + selectQuery + completeInnerJoin

	if len(groupByStatement) != 0 {
		resultSQLStatement += "GROUP BY " + groupByStatement
	}
	resultSQLStatement += " " + orderByQuery + channeAnalyticsLimit + ";"
	return resultSQLStatement, params, responseSelectMetrics
}

// TODO handle duplicates of groupBy - edge case
// @Kark TODO v1
func getCTEAndParamsForFacebookReportComplexStrategy(query *ChannelQueryV1, projectID uint64,
	customerAccountID string, docType int) (string, string, []string, []string, []interface{}) {
	cteAlias := "reports_cte"
	customerAccountIDs := strings.Split(customerAccountID, ",")
	staticWhereParams := []interface{}{projectID, customerAccountIDs, docType, query.From, query.To}
	selectQuery := "WITH " + cteAlias + " as (SELECT "
	cteJoinFields := []string{}
	isGroupByTimestamp := query.GetGroupByTimestamp() != ""
	groupByKeysWithoutTimestamp := make([]string, 0, 0)
	groupByStatement := ""
	uniqueGroupByObjects := make(map[string]struct{})
	for _, groupBy := range query.GroupBy {
		uniqueGroupByObjects[groupBy.Object] = struct{}{}
	}

	for key := range uniqueGroupByObjects {
		key := key + ":id"
		value := objectAndPropertyToValueInFacebookReportsMapping[key]
		groupByKeysWithoutTimestamp = append(groupByKeysWithoutTimestamp, value)
		cteJoinFields = append(cteJoinFields, value)
	}

	if isGroupByTimestamp {
		groupByStatement = joinWithComma(append(groupByKeysWithoutTimestamp, AliasDateTime)...)
	} else {
		groupByStatement = joinWithComma(groupByKeysWithoutTimestamp...)
	}

	selectQuery += joinWithComma(groupByKeysWithoutTimestamp...)
	selectQuery = appendSelectTimestampIfRequiredForChannels(selectQuery, query.GetGroupByTimestamp(), query.Timezone)

	for _, selectMetric := range query.SelectMetrics {
		currentSelectQuery := facebookMetricsToAggregatesInReportsMapping[selectMetric] + " as " + selectMetric
		selectQuery = joinWithComma(selectQuery, currentSelectQuery)
	}

	resultSQLStatement := selectQuery + fromFacebooksDocument + staticWhereStatementForFacebook + " GROUP BY " + groupByStatement + " ), "

	return resultSQLStatement, cteAlias, query.SelectMetrics, cteJoinFields, staticWhereParams
}

// @Kark TODO v1
func getCTEAndParamsForFacebookJobsComplexStrategy(query *ChannelQueryV1, projectID uint64, customerAccountID string) ([]string, []string, []string, []interface{}) {
	customerAccountIDs := strings.Split(customerAccountID, ",")

	campaignsFilters, adSetFilters, _ := splitFiltersByObjectTypeForFacebook(query)
	campaignsGroupBy, adSetsGroupBy, _ := splitGroupByByObjectTypeForFacebook(query)
	campaignJobCTE, campaignCTEAliasName, campaignJoinField, campaignParams := getCTEAndParamsForFacebookJob(query, projectID, customerAccountIDs, "campaign", campaignsFilters, campaignsGroupBy)
	adSetJobCTE, adSetCTEAliasName, adSetJoinField, adSetParams := getCTEAndParamsForFacebookJob(query, projectID, customerAccountIDs, "ad_set", adSetFilters, adSetsGroupBy)
	resultParams := append(make([]interface{}, 0, 0), campaignParams...)
	resultParams = append(resultParams, adSetParams...)
	return U.AppendNonNullValues(campaignJobCTE, adSetJobCTE), U.AppendNonNullValues(campaignCTEAliasName, adSetCTEAliasName), U.AppendNonNullValues(campaignJoinField, adSetJoinField), resultParams
}

// @Kark TODO v1
func getCTEAndParamsForFacebookJob(query *ChannelQueryV1, projectID uint64, customerAccountIDs []string, objectType string, filters []FilterV1, groupBy []GroupBy) (string, string, string, []interface{}) {
	if len(groupBy) < 1 {
		return "", "", "", make([]interface{}, 0, 0)
	}
	docType := facebookDocumentTypeAlias[objectType]
	staticWhereParams := []interface{}{projectID, customerAccountIDs, docType, query.From, query.To}
	aliasName := mapOfTypeToFacebookJobCTEAlias[objectType]
	withClause := aliasName + " as ("

	objectID := objectToValueInFacebookJobsMapping[objectType+":"+"id"]

	table1SQL, table1Alias, table1ColumnName, table1Params := getIDAndMaxTimeSQLAndParamsForFacebook(query, staticWhereParams, objectType, filters)
	table2SQL, table2Alias, table2ColumnName, table2Params := getCompleteRowSQLAndParamsForFacebookJob(query, staticWhereParams, objectType, filters)
	cteJoinField := objectID
	groupByQuery := getSelectPropertiesExceptIDsForFacebookJob(groupBy)
	selectQuery := "SELECT " + table1Alias + "." + objectID + " as " + objectID + ", " + groupByQuery

	resultStatement := fmt.Sprintf("%s %s FROM %s INNER JOIN %s ON %s.%s = %s.%s AND %s.%s = %s.%s), ", withClause, selectQuery, table1SQL, table2SQL,
		table1Alias, table1ColumnName, table2Alias, table2ColumnName, table1Alias, "timestamp", table2Alias, "timestamp")
	resultParams := append(make([]interface{}, 0, 0), table1Params...)
	resultParams = append(resultParams, table2Params...)
	return resultStatement, aliasName, cteJoinField, resultParams
}

// @Kark TODO v1
func appendSelectMetricsForFacebook(selectQuery string, selectMetrics []string) ([]string, string) {
	selectKeys := make([]string, 0, 0)
	for _, key := range selectMetrics {
		value := facebookMetricsToAggregatesInReportsMapping[key]
		selectKeys = append(selectKeys, value)
		selectQuery = joinWithComma(selectQuery, value)
	}
	return selectKeys, selectQuery
}

// @Kark TODO v1
func getSelectPropertiesExceptIDsForFacebookJob(groupBys []GroupBy) string {
	groupByQuery := ""
	for _, groupBy := range groupBys {
		key := groupBy.Object + ":" + groupBy.Property
		if groupBy.Property != "id" {
			groupByQuery += "value->>'" + objectToValueInFacebookJobsMapping[key] + "' as " + objectToValueInFacebookJobsMapping[key] + ", "
		}
	}
	return groupByQuery[:len(groupByQuery)-2]
}

// @Kark TODO v1
func getIDAndMaxTimeSQLAndParamsForFacebook(query *ChannelQueryV1, staticWhereParams []interface{}, objectType string, filters []FilterV1) (string, string, string, []interface{}) {
	aliasName := "distinct_id"
	idColumnName := objectType + "_id"
	selectStatement := "(SELECT " + idColumnName + " , max(timestamp) as timestamp" + fromFacebooksDocument
	groupByStatement := "GROUP BY " + idColumnName + ") "
	sqlParams := staticWhereParams
	filterPropertiesStatement, filterParams := getFilterPropertiesForFacebookJobStatementAndParams(filters)
	completeFiltersStatement := staticWhereStatementForFacebook
	if len(filterPropertiesStatement) != 0 {
		completeFiltersStatement += "AND " + filterPropertiesStatement + " "
		sqlParams = append(sqlParams, filterParams...)
	}
	resultStatement := selectStatement + completeFiltersStatement + groupByStatement + "as " + aliasName
	return resultStatement, aliasName, idColumnName, sqlParams
}

// @Kark TODO v1
func getCompleteRowSQLAndParamsForFacebookJob(query *ChannelQueryV1, staticWhereParams []interface{}, objectType string, filters []FilterV1) (string, string, string, []interface{}) {
	aliasName := "JobRecords"
	idColumnName := objectType + "_id"
	selectStatement := "(SELECT * FROM facebook_documents "
	resultStatement := selectStatement + staticWhereStatementForFacebook + ") as " + aliasName
	return resultStatement, aliasName, idColumnName, staticWhereParams
}

// @Kark TODO v1
func splitFiltersByObjectTypeForFacebook(query *ChannelQueryV1) ([]FilterV1, []FilterV1, []FilterV1) {
	campaignsFilters := make([]FilterV1, 0, 0)
	adSetFilters := make([]FilterV1, 0, 0)
	adFilters := make([]FilterV1, 0, 0)

	for _, filter := range query.Filters {
		switch filter.Object {
		case facebookCampaign:
			campaignsFilters = append(campaignsFilters, filter)
		case facebookAdSet:
			adSetFilters = append(adSetFilters, filter)
		case facebookAd:
			adFilters = append(adFilters, filter)
		}
	}
	return campaignsFilters, adSetFilters, adFilters
}

// @Kark TODO v1
func splitGroupByByObjectTypeForFacebook(query *ChannelQueryV1) ([]GroupBy, []GroupBy, []GroupBy) {
	campaignsGroupBys := make([]GroupBy, 0, 0)
	adSetGroupBys := make([]GroupBy, 0, 0)
	adSetBys := make([]GroupBy, 0, 0)

	for _, groupBy := range query.GroupBy {
		switch groupBy.Object {
		case facebookCampaign:
			campaignsGroupBys = append(campaignsGroupBys, groupBy)
		case facebookAdSet:
			adSetGroupBys = append(adSetGroupBys, groupBy)
		case facebookAd:
			adSetBys = append(adSetBys, groupBy)
		}
	}
	return campaignsGroupBys, adSetGroupBys, adSetBys
}

// @Kark TODO v1
// TODO Check if we have none operator
// TODO add support for indexed columns campaign_ids, ad_set_ids ... Hence not moving to common place
func getFilterPropertiesForFacebookJobStatementAndParams(filters []FilterV1) (string, []interface{}) {
	resultStatement := ""
	params := make([]interface{}, 0, 0)
	var filterValue string
	for index, filter := range filters {
		currentFilterStatement := ""
		if filter.LogicalOp == "" {
			filter.LogicalOp = "AND"
		}
		filterOperator := getOp(filter.Condition)
		if filter.Condition == ContainsOpStr || filter.Condition == NotContainsOpStr {
			filterValue = fmt.Sprintf("%%%s%%", filter.Value)
		} else {
			filterValue = filter.Value
		}
		currentFilterStatement = fmt.Sprintf("value->>? %s ?", filterOperator)
		params = append(params, filter.Property, filterValue)

		if index == 0 {
			resultStatement = currentFilterStatement
		} else {
			resultStatement = fmt.Sprintf("%s %s %s", resultStatement, filter.LogicalOp, currentFilterStatement)
		}
	}
	return resultStatement, params
}

// GetFacebookLastSyncInfo ...
func GetFacebookLastSyncInfo(projectID uint64, CustomerAdAccountID string) ([]FacebookLastSyncInfo, int) {
	db := C.GetServices().Db

	facebookLastSyncInfos := make([]FacebookLastSyncInfo, 0, 0)

	queryStr := "SELECT project_id, customer_ad_account_id, platform, type as document_type, max(timestamp) as last_timestamp" +
		" FROM facebook_documents WHERE project_id = ? AND customer_ad_account_id = ?" +
		" GROUP BY project_id, customer_ad_account_id, platform, type "

	rows, err := db.Raw(queryStr, projectID, CustomerAdAccountID).Rows()
	if err != nil {
		log.WithError(err).Error("Failed to get last facebook documents by type for sync info.")
		return facebookLastSyncInfos, http.StatusInternalServerError
	}
	defer rows.Close()

	for rows.Next() {
		var facebookLastSyncInfo FacebookLastSyncInfo
		if err := db.ScanRows(rows, &facebookLastSyncInfo); err != nil {
			log.WithError(err).Error("Failed to scan last facebook documents by type for sync info.")
			return []FacebookLastSyncInfo{}, http.StatusInternalServerError
		}

		facebookLastSyncInfos = append(facebookLastSyncInfos, facebookLastSyncInfo)
	}
	documentTypeAliasByType := getFacebookDocumentTypeAliasByType()

	for i := range facebookLastSyncInfos {
		logCtx := log.WithField("project_id", facebookLastSyncInfos[i].ProjectID)
		typeAlias, typeAliasExists := documentTypeAliasByType[facebookLastSyncInfos[i].DocumentType]
		if !typeAliasExists {
			logCtx.WithField("document_type",
				facebookLastSyncInfos[i].DocumentType).Error("Invalid document type given. No type alias name.")
			continue
		}

		facebookLastSyncInfos[i].DocumentTypeAlias = typeAlias
	}
	return facebookLastSyncInfos, http.StatusOK
}

// format yyyymmdd
func changeUnixTimestampToDate(timestamp int64) int64 {
	time := time.Unix(timestamp, 0)
	date, _ := strconv.ParseInt(time.Format("20060102"), 10, 64)
	return date
}

// ExecuteFacebookChannelQuery - @TODO Kark v0
func ExecuteFacebookChannelQuery(projectID uint64, query *ChannelQuery) (*ChannelQueryResult, int) {
	logCtx := log.WithField("project_id", projectID).WithField("query", query)

	if projectID == 0 || query == nil {
		logCtx.Error("Invalid project_id or query on execute facebook channel query.")
		return nil, http.StatusInternalServerError
	}

	projectSetting, errCode := GetProjectSetting(projectID)
	if errCode != http.StatusFound {
		return nil, http.StatusInternalServerError
	}

	if projectSetting.IntFacebookAdAccount == "" {
		logCtx.Error("Execute facebook channel query failed. No customer account id.")
		return nil, http.StatusInternalServerError
	}

	query.From = changeUnixTimestampToDate(query.From)
	query.To = changeUnixTimestampToDate(query.To)
	queryResult := &ChannelQueryResult{}
	result, err := getFacebookChannelResult(projectID, projectSetting.IntFacebookAdAccount, query)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get facebook query result.")
		return queryResult, http.StatusInternalServerError
	}
	queryResult = result
	if query.Breakdown == "" {
		return queryResult, http.StatusOK
	}

	metricBreakDown, err := getFacebookMetricBreakdown(projectID, projectSetting.IntFacebookAdAccount, query)
	queryResult.MetricsBreakdown = metricBreakDown

	// sort only if the impression is there as column
	impressionsIndex := 0
	for _, key := range queryResult.MetricsBreakdown.Headers {
		if key == "impressions" {
			// sort the rows by impressions count in descending order
			sort.Slice(queryResult.MetricsBreakdown.Rows, func(i, j int) bool {
				return queryResult.MetricsBreakdown.Rows[i][impressionsIndex].(float64) > queryResult.MetricsBreakdown.Rows[j][impressionsIndex].(float64)
			})
			break
		}
		impressionsIndex++
	}
	return queryResult, http.StatusOK
}

// @TODO Kark v0
func getFacebookMetricBreakdown(projectID uint64, customerAccountID string, query *ChannelQuery) (*ChannelBreakdownResult, error) {
	logCtx := log.WithField("project_id", projectID).WithField("customer_account_id", customerAccountID)

	sqlQuery, documentType := getFacebookMetricsQuery(query, true)

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

	return &ChannelBreakdownResult{Headers: resultHeaders, Rows: resultRows}, nil
}

// @TODO Kark v0
func getFacebookDocumentType(query *ChannelQuery) int {
	var documentType int
	if query.FilterKey == "ad" {
		documentType = 4
	}
	if query.FilterKey == "campaign" {
		documentType = 5
	}
	if query.FilterKey == "adset" {
		documentType = 6
	}
	return documentType
}

// @TODO Kark v0
func getFacebookMetricsQuery(query *ChannelQuery, withBreakdown bool) (string, int) {

	documentType := getFacebookDocumentType(query)

	selectColstWithoutAlias := "SUM((value->>'impressions')::float) as %s , SUM((value->>'clicks')::float) as %s," +
		" " + "SUM((value->>'spend')::float) as %s," +
		" " + "SUM((value->>'unique_clicks')::float) as %s," +
		" " + "SUM((value->>'reach')::float) as %s, AVG((value->>'frequency')::float) as %s, " +
		" " + "SUM((value->>'inline_post_engagement')::float) as %s," +
		" " + "AVG((value->>'cpc')::float) as %s"

	selectCols := fmt.Sprintf(selectColstWithoutAlias, CAColumnImpressions, CAColumnClicks,
		CAColumnTotalCost, CAColumnUniqueClicks, CAColumnReach,
		CAColumnFrequency, CAColumnInlinePostEngagement, CAColumnCostPerClick)

	strmntWhere := "WHERE project_id= ? AND customer_ad_account_id = ? AND timestamp BETWEEN ? AND ? AND type=? and platform!='facebook_all'"

	strmntGroupBy := ""
	if withBreakdown {
		if query.Breakdown == platform {
			selectCols = platform + ", " + selectCols
			strmntGroupBy = "GROUP BY " + platform
		} else {
			firstValue := "(value->>'%s_name') as name, "
			firstValue = fmt.Sprintf(firstValue, query.Breakdown)
			selectCols = firstValue + selectCols
			strmntGroupBy = "GROUP BY id, (value->>'%s_name')"
			strmntGroupBy = fmt.Sprintf(strmntGroupBy, query.Breakdown)
		}
	}

	sqlQuery := "SELECT" + " " + selectCols + " " + "FROM facebook_documents" + " " + strmntWhere + " " + strmntGroupBy
	return sqlQuery, documentType
}

// @TODO Kark v0
func getFacebookChannelResult(projectID uint64, customerAccountID string, query *ChannelQuery) (*ChannelQueryResult, error) {

	logCtx := log.WithField("project_id", projectID)

	sqlQuery, documentType := getFacebookMetricsQuery(query, false)

	queryResult := &ChannelQueryResult{}
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
		log.Error("Aggregate query returned more than one row on get facebook metric kvs.")
	}

	metricKvs := make(map[string]interface{})
	for i, k := range resultHeaders {
		metricKvs[k] = resultRows[0][i]
	}

	queryResult.Metrics = &metricKvs
	return queryResult, nil
}
