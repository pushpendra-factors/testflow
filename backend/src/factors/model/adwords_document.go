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

// AdwordsDocument ...
type AdwordsDocument struct {
	ProjectID         uint64          `gorm:"primary_key:true;auto_increment:false" json:"project_id"`
	CustomerAccountID string          `gorm:"primary_key:true;auto_increment:false" json:"customer_acc_id"`
	TypeAlias         string          `gorm:"-" json:"type_alias"`
	Type              int             `gorm:"primary_key:true;auto_increment:false" json:"-"`
	Timestamp         int64           `gorm:"primary_key:true;auto_increment:false" json:"timestamp"`
	ID                string          `gorm:"primary_key:true;auto_increment:false" json:"id"`
	Value             *postgres.Jsonb `json:"value"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
}

const campaignPerformanceReport = "campaign_performance_report"
const adGroupPerformanceReport = "ad_group_performance_report"
const adPerformanceReport = "ad_performance_report"
const keywordPerformanceReport = "keyword_performance_report"

// AdwordsDocumentTypeAlias ...
var AdwordsDocumentTypeAlias = map[string]int{
	"campaigns":                   1,
	"ads":                         2,
	"ad_groups":                   3,
	"click_performance_report":    4,
	campaignPerformanceReport:     5,
	adPerformanceReport:           6,
	"search_performance_report":   7,
	keywordPerformanceReport:      8,
	"customer_account_properties": 9,
	adGroupPerformanceReport:      10,
}

var objectAndPropertyToValueInReportsMapping = map[string]string{
	"campaign:id": "campaign_id",
	"ad_group:id": "ad_group_id",
	"ad:id":       "id",
}

var objectToValueInJobsMapping = map[string]string{
	"campaign:name": "name",
	"ad_group:name": "name",
	"campaign:id":   "campaign_id",
	"ad_group:id":   "ad_group_id",
	"ad:id":         "ad_id",
}

var mapOfTypeToJobCTEAlias = map[string]string{
	"ad_group": "ad_group_cte",
	"campaign": "campaign_cte",
}

var adwordsMetricsToAggregatesInReportsMapping = map[string]string{
	"impressions": "SUM((value->>'impressions')::float)",
	"clicks":      "SUM((value->>'clicks')::float)",
	"spend":       "SUM((value->>'cost')::float)",
	// "cost_per_click": "average_cost",
	"conversion": "SUM((value->>'conversions')::float",
	// "conversion_rate": "conversion_rate"
}

var adwordsMetricsToOperation = map[string]string{
	"impressions": "sum",
	"clicks":      "sum",
	"cost":        "sum",
	"conversions": "sum",
}

var adwordsRequestPropertiesToSQLproperty = map[string]string{
	"name":        "name",
	"id":          "id",
	"impressions": "impressions",
	"clicks":      "clicks",
	"spend":       "cost",
	"conversion":  "conversions",
}

const errorDuplicateAdwordsDocument = "pq: duplicate key value violates unique constraint \"adwords_documents_pkey\""
const filterValueAll = "all"

var errorEmptyAdwordsDocument = errors.New("empty adwords document")

const adwordsFilterQueryStr = "SELECT DISTINCT(value->>?) as filter_value FROM adwords_documents WHERE project_id = ? AND" + " " + "customer_account_id = ? AND type = ? LIMIT 5000"

const staticWhereStatement = "WHERE project_id = ? AND customer_account_id IN ( ? ) AND type = ? AND timestamp between ? AND ? "

const innerJoinClause = " INNER JOIN "

const fromAdwordsDocument = " FROM adwords_documents "

const channeAnalyticsLimit = " LIMIT 2500 "

func isDuplicateAdwordsDocumentError(err error) bool {
	return err.Error() == errorDuplicateAdwordsDocument
}

func getAdwordsIDFieldNameByType(docType int) string {
	switch docType {
	case 4: // click_performance_report
		return "gcl_id"
	case 5: // campaign_performance_report
		return "campaign_id"
	case 7: // search_performance_report
		return "query"
	case 9: // customer_account_properties
		return "customer_id"
	case 10: // ad_group_performance_report
		return "ad_group_id"
	default: // others
		return "id"
	}
}

// Returns campaign_id, ad_group_id, ad_id, keyword_id
func getAdwordsHierarchyColumnsByType(valueMap *map[string]interface{}, docType int) (int64, int64, int64, int64) {
	switch docType {
	case AdwordsDocumentTypeAlias["campaigns"]:
		return U.GetInt64FromMapOfInterface(*valueMap, "id", 0), 0, 0, 0
	case AdwordsDocumentTypeAlias["ad_groups"]:
		return U.GetInt64FromMapOfInterface(*valueMap, "campaign_id", 0), U.GetInt64FromMapOfInterface(*valueMap, "id", 0), 0, 0
	case AdwordsDocumentTypeAlias["click_performance_report"], AdwordsDocumentTypeAlias["search_performance_report"], AdwordsDocumentTypeAlias["ad_group_performance_report"]:
		return U.GetInt64FromMapOfInterface(*valueMap, "campaign_id", 0), U.GetInt64FromMapOfInterface(*valueMap, "ad_group_id", 0), 0, 0
	case AdwordsDocumentTypeAlias["campaign_performance_report"]:
		return U.GetInt64FromMapOfInterface(*valueMap, "campaign_id", 0), 0, 0, 0
	case AdwordsDocumentTypeAlias["ad_performance_report"]:
		return U.GetInt64FromMapOfInterface(*valueMap, "campaign_id", 0), U.GetInt64FromMapOfInterface(*valueMap, "ad_group_id", 0), U.GetInt64FromMapOfInterface(*valueMap, "id", 0), 0
	case AdwordsDocumentTypeAlias["keyword_performance_report"]:
		return U.GetInt64FromMapOfInterface(*valueMap, "campaign_id", 0), U.GetInt64FromMapOfInterface(*valueMap, "ad_group_id", 0), 0, U.GetInt64FromMapOfInterface(*valueMap, "id", 0)
	case AdwordsDocumentTypeAlias["customer_account_properties"]:
		return 0, 0, 0, 0
	default:
		return 0, 0, 0, 0
	}
}

// GetAdwordsDateOnlyTimestamp - Date only timestamp to query adwords documents.
func GetAdwordsDateOnlyTimestamp(unixTimestamp int64) string {
	// Todo: Add timezone support using util.getTimeFromUnixTimestampWithZone.
	return time.Unix(unixTimestamp, 0).UTC().Format("20060102")
}

func getAdwordsIDAndHeirarchyColumnsByType(docType int, valueJSON *postgres.Jsonb) (string, int64, int64, int64, int64, error) {
	if docType > len(AdwordsDocumentTypeAlias) {
		return "", 0, 0, 0, 0, errors.New("invalid document type")
	}

	valueMap, err := U.DecodePostgresJsonb(valueJSON)
	if err != nil {
		return "", 0, 0, 0, 0, err
	}

	if len(*valueMap) == 0 {
		return "", 0, 0, 0, 0, errorEmptyAdwordsDocument
	}

	idFieldName := getAdwordsIDFieldNameByType(docType)
	value1, value2, value3, value4 := getAdwordsHierarchyColumnsByType(valueMap, docType)
	id, exists := (*valueMap)[idFieldName]
	if !exists {
		return "", 0, 0, 0, 0, fmt.Errorf("id field %s does not exist on doc of type %v", idFieldName, docType)
	}

	if id == nil {
		return "", 0, 0, 0, 0, fmt.Errorf("id field %s has empty value on doc of type %v", idFieldName, docType)
	}

	idStr, err := U.GetValueAsString(id)
	if err != nil {
		return "", 0, 0, 0, 0, err
	}

	// ID as string always.
	return idStr, value1, value2, value3, value4, nil
}

// CreateAdwordsDocument ...
func CreateAdwordsDocument(adwordsDoc *AdwordsDocument) int {
	logCtx := log.WithField("customer_acc_id", adwordsDoc.CustomerAccountID).WithField(
		"project_id", adwordsDoc.ProjectID)

	if adwordsDoc.CustomerAccountID == "" || adwordsDoc.TypeAlias == "" {
		logCtx.Error("Invalid adwords document.")
		return http.StatusBadRequest
	}

	logCtx = logCtx.WithField("type_alias", adwordsDoc.TypeAlias)
	docType, docTypeExists := AdwordsDocumentTypeAlias[adwordsDoc.TypeAlias]
	if !docTypeExists {
		logCtx.Error("Invalid type alias.")
		return http.StatusBadRequest
	}
	adwordsDoc.Type = docType

	adwordsDocID, campaignIDValue, adGroupIDValue, adIDValue, keywordIDValue, err := getAdwordsIDAndHeirarchyColumnsByType(adwordsDoc.Type, adwordsDoc.Value)
	if err != nil {
		if err == errorEmptyAdwordsDocument {
			// Using UUID to allow storing empty response.
			// To avoid downloading reports again for the same timerange.
			adwordsDocID = U.GetUUID()
		} else {
			logCtx.WithError(err).Error("Failed to get id by adowords doc type.")
			return http.StatusInternalServerError
		}
	}

	adwordsDoc.ID = adwordsDocID

	db := C.GetServices().Db
	queryStr := "INSERT INTO adwords_documents (project_id,customer_account_id,type,timestamp,id,campaign_id,ad_group_id,ad_id,keyword_id,value,created_at,updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
	rows, err := db.Raw(queryStr, adwordsDoc.ProjectID, adwordsDoc.CustomerAccountID,
		adwordsDoc.Type, adwordsDoc.Timestamp, adwordsDoc.ID, campaignIDValue, adGroupIDValue, adIDValue, keywordIDValue, adwordsDoc.Value, time.Now().UTC(), time.Now().UTC()).Rows()
	if err != nil {
		if isDuplicateAdwordsDocumentError(err) {
			logCtx.WithError(err).WithField("id", adwordsDoc.ID).Error(
				"Failed to create an adwords doc. Duplicate.")
			return http.StatusConflict
		} else {
			logCtx.WithError(err).WithField("id", adwordsDoc.ID).Error(
				"Failed to create an adwords doc. Continued inserting other docs.")
			return http.StatusInternalServerError
		}
	}
	defer rows.Close()

	return http.StatusCreated
}

// AdwordsLastSyncInfo ...
type AdwordsLastSyncInfo struct {
	ProjectId         uint64 `json:"project_id"`
	CustomerAccountId string `json:"customer_acc_id"`
	RefreshToken      string `json:"refresh_token"`
	DocumentType      int    `json:"-"`
	DocumentTypeAlias string `json:"doc_type_alias"`
	LastTimestamp     int64  `json:"last_timestamp"`
}

func getDocumentTypeAliasByType() map[int]string {
	documentTypeMap := make(map[int]string, 0)
	for alias, typ := range AdwordsDocumentTypeAlias {
		documentTypeMap[typ] = alias
	}

	return documentTypeMap
}

// GetAllAdwordsLastSyncInfoByProjectCustomerAccountAndType - @TODO Kark v1
func GetAllAdwordsLastSyncInfoByProjectCustomerAccountAndType() ([]AdwordsLastSyncInfo, int) {
	db := C.GetServices().Db

	adwordsLastSyncInfos := make([]AdwordsLastSyncInfo, 0, 0)

	queryStr := "SELECT project_id, customer_account_id, type as document_type, max(timestamp) as last_timestamp" +
		" " + "FROM adwords_documents GROUP BY project_id, customer_account_id, type"

	rows, err := db.Raw(queryStr).Rows()
	if err != nil {
		log.WithError(err).Error("Failed to get last adwords documents by type for sync info.")
		return adwordsLastSyncInfos, http.StatusInternalServerError
	}
	defer rows.Close()

	for rows.Next() {
		var adwordsLastSyncInfo AdwordsLastSyncInfo
		if err := db.ScanRows(rows, &adwordsLastSyncInfo); err != nil {
			log.WithError(err).Error("Failed to scan last adwords documents by type for sync info.")
			return []AdwordsLastSyncInfo{}, http.StatusInternalServerError
		}

		adwordsLastSyncInfos = append(adwordsLastSyncInfos, adwordsLastSyncInfo)
	}

	adwordsSettings, errCode := GetAllIntAdwordsProjectSettings()
	if errCode != http.StatusOK {
		return []AdwordsLastSyncInfo{}, errCode
	}

	adwordsSettingsByProjectAndCustomerAccount := make(map[uint64]map[string]*AdwordsProjectSettings, 0)

	for i := range adwordsSettings {
		customerAccountIDs := strings.Split(adwordsSettings[i].CustomerAccountId, ",")
		adwordsSettingsByProjectAndCustomerAccount[adwordsSettings[i].ProjectId] = make(map[string]*AdwordsProjectSettings)
		for j := range customerAccountIDs {
			var setting AdwordsProjectSettings
			setting.ProjectId = adwordsSettings[i].ProjectId
			setting.AgentUUID = adwordsSettings[i].AgentUUID
			setting.RefreshToken = adwordsSettings[i].RefreshToken
			setting.CustomerAccountId = customerAccountIDs[j]
			adwordsSettingsByProjectAndCustomerAccount[adwordsSettings[i].ProjectId][customerAccountIDs[j]] = &setting
		}
	}
	documentTypeAliasByType := getDocumentTypeAliasByType()

	// add settings for project_id existing on adwords documents.
	existingProjectAndCustomerAccountWithTypes := make(map[uint64]map[string]map[string]bool, 0)
	selectedLastSyncInfos := make([]AdwordsLastSyncInfo, 0, 0)

	for i := range adwordsLastSyncInfos {
		logCtx := log.WithFields(
			log.Fields{"project_id": adwordsLastSyncInfos[i].ProjectId,
				"customer_account_id": adwordsLastSyncInfos[i].CustomerAccountId})

		settings, exists := adwordsSettingsByProjectAndCustomerAccount[adwordsLastSyncInfos[i].ProjectId][adwordsLastSyncInfos[i].CustomerAccountId]
		if !exists {
			logCtx.Error("Adwords project settings not found for customer account adwords synced earlier.")
		}

		if settings == nil {
			logCtx.Info("Adwords disabled for project.")
			continue
		}

		typeAlias, typeAliasExists := documentTypeAliasByType[adwordsLastSyncInfos[i].DocumentType]
		if !typeAliasExists {
			logCtx.WithField("document_type",
				adwordsLastSyncInfos[i].DocumentType).Error("Invalid document type given. No type alias name.")
			continue
		}

		adwordsLastSyncInfos[i].DocumentTypeAlias = typeAlias // map the type to type alias name.
		adwordsLastSyncInfos[i].RefreshToken = settings.RefreshToken

		selectedLastSyncInfos = append(selectedLastSyncInfos, adwordsLastSyncInfos[i])

		if _, projectWithCustomerAccountExists := existingProjectAndCustomerAccountWithTypes[adwordsLastSyncInfos[i].ProjectId][adwordsLastSyncInfos[i].CustomerAccountId]; !projectWithCustomerAccountExists {
			if _, projectExists := existingProjectAndCustomerAccountWithTypes[adwordsLastSyncInfos[i].ProjectId]; !projectExists {
				existingProjectAndCustomerAccountWithTypes[adwordsLastSyncInfos[i].ProjectId] = make(map[string]map[string]bool, 0)
			}
			existingProjectAndCustomerAccountWithTypes[adwordsLastSyncInfos[i].ProjectId][adwordsLastSyncInfos[i].CustomerAccountId] = make(map[string]bool, 0)
		}

		existingProjectAndCustomerAccountWithTypes[adwordsLastSyncInfos[i].ProjectId][adwordsLastSyncInfos[i].CustomerAccountId][adwordsLastSyncInfos[i].DocumentTypeAlias] = true
	}

	// add all types for missing projects and
	// add missing types for existing projects.
	for i := range adwordsSettings {
		customerAccountIDs := strings.Split(adwordsSettings[i].CustomerAccountId, ",")
		for _, accountID := range customerAccountIDs {
			existingTypesForAccount, accountExists := existingProjectAndCustomerAccountWithTypes[adwordsSettings[i].ProjectId][accountID]
			for docTypeAlias := range AdwordsDocumentTypeAlias {
				if !accountExists || (accountExists && existingTypesForAccount[docTypeAlias] == false) {
					syncInfo := AdwordsLastSyncInfo{
						ProjectId:         adwordsSettings[i].ProjectId,
						RefreshToken:      adwordsSettings[i].RefreshToken,
						CustomerAccountId: accountID,
						LastTimestamp:     0, // no sync yet.
						DocumentTypeAlias: docTypeAlias,
					}

					selectedLastSyncInfos = append(selectedLastSyncInfos, syncInfo)
				}
			}
		}

	}

	return selectedLastSyncInfos, http.StatusOK
}

// GetGCLIDBasedCampaignInfo - It returns GCLID based campaign info ( Adgroup, Campaign and Ad) for given time range and adwords account
func GetGCLIDBasedCampaignInfo(projectID uint64, from, to int64, adwordsAccountIDs string) (map[string]CampaignInfo, error) {

	db := C.GetServices().Db
	logCtx := log.WithFields(log.Fields{"ProjectID": projectID, "Range": fmt.Sprintf("%d - %d", from, to)})
	adGroupNameCase := "CASE WHEN value->>'ad_group_name' IS NULL THEN ? " +
		" WHEN value->>'ad_group_name' = '' THEN ? ELSE value->>'ad_group_name' END AS ad_group_name"
	campaignNameCase := "CASE WHEN value->>'campaign_name' IS NULL THEN ? " +
		" WHEN value->>'campaign_name' = '' THEN ? ELSE value->>'campaign_name' END AS campaign_name"
	adIDCase := "CASE WHEN value->>'creative_id' IS NULL THEN ? " +
		" WHEN value->>'creative_id' = '' THEN ? ELSE value->>'creative_id' END AS creative_id"

	performanceQuery := "SELECT id, " + adGroupNameCase + ", " + campaignNameCase + ", " + adIDCase +
		" FROM adwords_documents where project_id = ? AND customer_account_id IN (?) AND type = ? AND timestamp between ? AND ? "
	customerAccountIDs := strings.Split(adwordsAccountIDs, ",")
	rows, err := db.Raw(performanceQuery, PropertyValueNone, PropertyValueNone, PropertyValueNone, PropertyValueNone,
		PropertyValueNone, PropertyValueNone, projectID, customerAccountIDs, AdwordsClickReportType, U.GetDateOnlyFromTimestamp(from),
		U.GetDateOnlyFromTimestamp(to)).Rows()
	if err != nil {
		logCtx.WithError(err).Error("SQL Query failed")
		return nil, err
	}
	defer rows.Close()
	gclIDBasedCampaign := make(map[string]CampaignInfo)
	for rows.Next() {
		var gclID string
		var adgroupName string
		var campaignName string
		var adID string
		if err = rows.Scan(&gclID, &adgroupName, &campaignName, &adID); err != nil {
			logCtx.WithError(err).Error("SQL Parse failed")
			continue
		}
		gclIDBasedCampaign[gclID] = CampaignInfo{
			AdgroupName:  adgroupName,
			CampaignName: campaignName,
			AdID:         adID,
		}
	}
	return gclIDBasedCampaign, nil
}

// GetAdwordsFilterValues - @TODO Kark v1
func GetAdwordsFilterValues(projectID uint64, filterObject string, filterProperty string, reqID string) ([]interface{}, int) {
	docType, property, errCode := getAdwordsDocumentTpeAndPropertyKeyForFilter(filterObject, filterProperty)

	if errCode != http.StatusFound {
		return []interface{}{}, errCode
	}

	filterValues, errCode := getAdwordsFilterValuesByType(projectID, docType, property, reqID)
	if errCode != http.StatusFound {
		return []interface{}{}, http.StatusInternalServerError
	}

	return filterValues, http.StatusFound
}

// GetAdwordsSQLQueryAndParametersForFilterValues - @TODO Kark v1
func GetAdwordsSQLQueryAndParametersForFilterValues(projectID uint64, filterObject string, filterProperty string) (string, []interface{}, int) {
	docType, property, errCode := getAdwordsDocumentTpeAndPropertyKeyForFilter(filterObject, filterProperty)
	if errCode != http.StatusFound {
		return "", []interface{}{}, errCode
	}

	projectSetting, errCode := GetProjectSetting(projectID)
	if errCode != http.StatusFound {
		return "", []interface{}{}, http.StatusInternalServerError
	}
	customerAccountID := projectSetting.IntAdwordsCustomerAccountId

	params := []interface{}{property, projectID, customerAccountID, docType}
	return "(" + adwordsFilterQueryStr + ")", params, http.StatusFound
}

// @TODO Kark v1
func getAdwordsDocumentTpeAndPropertyKeyForFilter(filterObject string, filterProperty string) (int, string, int) {
	docType, err := getAdwordsDocumentTypeForFilterKeyV1(filterObject)

	if err != nil {
		return 0, "", http.StatusInternalServerError
	}

	property, err := getAdwordsFilterPropertyKeyByTypeV1(filterObject, filterProperty)
	if err != nil {
		return 0, "", http.StatusBadRequest
	}

	return docType, property, http.StatusFound
}

// @TODO Kark v1
func getAdwordsFilterPropertyKeyByTypeV1(filterObject string, filterProperty string) (string, error) {
	property, isPropertyPresent := adwordsRequestPropertiesToSQLproperty[filterProperty]

	if !isPropertyPresent {
		return "", errors.New("no filter key found for document type")
	}
	return property, nil
}

// @TODO Kark v1
func getAdwordsFilterValuesByType(projectID uint64, docType int, property string, reqID string) ([]interface{}, int) {
	projectSetting, errCode := GetProjectSetting(projectID)
	if errCode != http.StatusFound {
		return []interface{}{}, http.StatusInternalServerError
	}
	customerAccountID := projectSetting.IntAdwordsCustomerAccountId

	logCtx := log.WithField("project_id", projectID).WithField("doc_type", docType).WithField("req_id", reqID)
	params := []interface{}{property, projectID, customerAccountID, docType}
	_, resultRows, _ := ExecuteSQL(adwordsFilterQueryStr, params, logCtx)

	return Convert2DArrayTo1DArray(resultRows), http.StatusFound
}

// @TODO Kark v1
func getAdwordsDocumentTypeForFilterKeyV1(filterObject string) (int, error) {
	var docType int

	switch filterObject {
	case CAFilterCampaign:
		docType = AdwordsDocumentTypeAlias[CAFilterCampaign+"s"]
	case CAFilterAdGroup:
		docType = AdwordsDocumentTypeAlias[CAFilterAdGroup+"s"]
	case CAFilterAd:
		docType = AdwordsDocumentTypeAlias[CAFilterAd+"s"]
		// case CAFilterKeyword:
		// 	docType = AdwordsDocumentTypeAlias[keywordPerformanceReport]
	}

	if docType == 0 {
		return docType, errors.New("no adwords document type for filter")
	}

	return docType, nil
}

// ExecuteAdwordsChannelQueryV1 - @TODO Kark v1.
// Job represents the meta data associated with particular object type. Reports represent data with metrics and few filters.
func ExecuteAdwordsChannelQueryV1(projectID uint64, query *ChannelQueryV1, reqID string) ([]string, [][]interface{}, error) {
	logCtx := log.WithField("req_id", reqID)
	logCtx.Info("query", query)
	projectSetting, errCode := GetProjectSetting(projectID)
	if errCode != http.StatusFound {
		return nil, nil, nil
	}
	customerAccountID := projectSetting.IntAdwordsCustomerAccountId

	if hasAllIDsOnlyInGroupBy(query) {
		return buildAndExecuteAdwordsSimpleQueryV1(query, projectID, *customerAccountID, reqID)
	}
	return buildAndExecuteAdwordsComplexQueryV1(query, projectID, *customerAccountID, reqID)
}

// @Kark TODO v1
// TODO Error add
func buildAndExecuteAdwordsSimpleQueryV1(query *ChannelQueryV1, projectID uint64, customerAccountID string, reqID string) ([]string, [][]interface{}, error) {
	campaignIDs, adGroupIDs := getIDsFromJob(query, projectID, customerAccountID, reqID)
	lowestHierarchyLevel, _ := getLowestHierarchyLevel(query)
	lowestHierarchyReportLevel := lowestHierarchyLevel + "_performance_report"

	return getResultFromReports(query, projectID, query.From, query.To, customerAccountID, AdwordsDocumentTypeAlias[lowestHierarchyReportLevel],
		campaignIDs, adGroupIDs, reqID)
}

// @TODO Kark v1
func getIDsFromJob(query *ChannelQueryV1, projectID uint64, adwordsAccountIDs string, reqID string) ([]int, []int) {
	campaignsFilters, adGroupFilters, _ := splitFiltersByObjectType(query)

	campaignIDs, _ := getIDsByPropertyOnJob(projectID, query.From, query.To, adwordsAccountIDs, AdwordsDocumentTypeAlias["campaigns"], campaignsFilters, reqID)
	adGroupIDs, _ := getIDsByPropertyOnJob(projectID, query.From, query.To, adwordsAccountIDs, AdwordsDocumentTypeAlias["ad_groups"], adGroupFilters, reqID)
	// adIDs := getAdIDsByPropertyOnJob(adFilters)
	return campaignIDs, adGroupIDs
}

// @TODO Kark v1
func getIDsByPropertyOnJob(projectID uint64, from, to int64, adwordsAccountIDs string, type1 int, filters []FilterV1, reqID string) ([]int, error) {
	logCtx := log.WithField("req_id", reqID)
	db := C.GetServices().Db
	if len(filters) == 0 {
		return []int{}, nil
	}
	sqlParams := make([]interface{}, 0)
	customerAccountIDs := strings.Split(adwordsAccountIDs, ",")

	selectStatement := "SELECT value->'id'" + fromAdwordsDocument
	groupByStatement := "GROUP BY value->'id'"

	sqlParams = append(sqlParams, projectID, customerAccountIDs, type1, from, to)
	filterPropertiesStatement, filterPropertiesParams := getFilterPropertiesForJobStatementAndParams(filters)
	completeFiltersStatement := staticWhereStatement
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
func getLowestHierarchyLevel(query *ChannelQueryV1) (string, error) {
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
		if objectName == "ad" {
			return "ad", nil
		}
	}

	for _, objectName := range objectNames {
		if objectName == "ad_group" {
			return "ad_group", nil
		}
	}

	for _, objectName := range objectNames {
		if objectName == "campaign" {
			return "campaign", nil
		}
	}
	return "campaign", nil
}

// @TODO Kark v1
// TODO Map of request params to proper fields in documents.
func getResultFromReports(query *ChannelQueryV1, projectID uint64, from, to int64, adwordsAccountIDs string,
	docType int, campaignIDs []int, adGroupIDs []int, reqID string) ([]string, [][]interface{}, error) {

	logCtx := log.WithField("req_id", reqID)
	customerAccountIDs := strings.Split(adwordsAccountIDs, ",")
	selectQuery := ""
	isGroupByTimestamp := query.GetGroupByTimestamp() != ""
	groupByStatement := ""
	groupByKeysWithoutTimestamp := make([]string, 0, 0)

	for _, groupBy := range query.GroupBy {
		key := groupBy.Object + ":" + groupBy.Property
		groupByKeysWithoutTimestamp = append(groupByKeysWithoutTimestamp, objectAndPropertyToValueInReportsMapping[key])
	}

	if isGroupByTimestamp {
		groupByStatement = joinWithComma(append(groupByKeysWithoutTimestamp, AliasDateTime)...)
	} else {
		groupByStatement = joinWithComma(groupByKeysWithoutTimestamp...)
	}

	selectQuery += joinWithComma(groupByKeysWithoutTimestamp...)
	selectQuery = appendSelectTimestampIfRequired(selectQuery, query.GetGroupByTimestamp(), query.Timezone)
	selectMetrics, selectQuery := appendSelectMetrics(selectQuery, query.SelectMetrics)
	selectQuery = "SELECT " + selectQuery

	orderByQuery := "ORDER BY " + getOrderByClause(selectMetrics)

	whereConditionForIDs := ""
	if len(campaignIDs) != 0 {
		campaignString := ""
		for _, campaignID := range campaignIDs {
			campaignString += strconv.Itoa(campaignID) + ","
		}
		campaignString = campaignString[:len(campaignString)-1]
		whereConditionForIDs += "AND campaign_id IN " + "(" + campaignString + ") "
	}
	if len(adGroupIDs) != 0 {
		adGroupstring := ""
		for _, adGroupID := range adGroupIDs {
			adGroupstring += strconv.Itoa(adGroupID) + ","
		}
		adGroupstring = adGroupstring[:len(adGroupstring)-1]
		whereConditionForIDs += "AND ad_group_id IN " + "(" + adGroupstring + ") "
	}

	resultSQLStatement := selectQuery + fromAdwordsDocument + staticWhereStatement + whereConditionForIDs
	if len(groupByStatement) != 0 {
		resultSQLStatement += "GROUP BY " + groupByStatement
	}
	resultSQLStatement += " " + orderByQuery + channeAnalyticsLimit + ";"
	staticWhereParams := []interface{}{projectID, customerAccountIDs, docType, from, to}
	return ExecuteSQL(resultSQLStatement, staticWhereParams, logCtx)
}

/*
With reportsCTE as (SELECT campaign_id, SUM((value->>'impressions')::float) as impressions FROM adwords_documents WHERE project_id = '8' AND customer_account_id IN ( '3543296298' )
AND type = '5' AND timestamp between '20200401' AND '20200402'  GROUP BY campaign_id ),

CampaignCTE as (SELECT DistinctID.campaign_id as campaign_id, value->>'name' as name from
(SELECT campaign_id , max(timestamp) FROM adwords_documents WHERE project_id = '8' AND customer_account_id IN ( '3543296298' ) AND type = '1' AND
timestamp between '20200401' AND '20200402' AND value->>'name' = 'Brand - NOIDA - New_Aug_Desktop_RLSA' OR value->>'name' = 'LS_Display_SDC_BLR' GROUP BY campaign_id) as DistinctID
INNER JOIN (SELECT * FROM adwords_documents WHERE project_id = '8' AND customer_account_id IN ( '3543296298' ) AND type = '1' AND timestamp between '20200401' AND '20200402' ) as JobRecords
ON DistinctID.campaign_id = JobRecords.campaign_id)

SELECT CampaignCTE.name, sum(reportsCTE.impressions) from reportsCTE
INNER JOIN CampaignCTE ON reportsCTE.campaign_id = CampaignCTE.campaign_id  GROUP BY CampaignCTE.name;
*/
// @Kark TODO v1
func buildAndExecuteAdwordsComplexQueryV1(query *ChannelQueryV1, projectID uint64, customerAccountID string, reqID string) ([]string, [][]interface{}, error) {
	logCtx := log.WithField("xreq_id", reqID)
	selectQuery := "SELECT "
	selectMetrics := make([]string, 0, 0)
	isGroupByTimestamp := query.GetGroupByTimestamp() != ""
	groupByStatement := ""
	groupByKeysWithoutTimestamp := make([]string, 0, 0)
	lowestHierarchyLevel, _ := getLowestHierarchyLevel(query)
	lowestHierarchyReportLevel := lowestHierarchyLevel + "_performance_report"

	reportCTE, reportCTEAlias, reportSelectMetrics, reportCTEJoinFields, reportParams := getCTEAndParamsForReportComplexStrategy(query, projectID, customerAccountID, AdwordsDocumentTypeAlias[lowestHierarchyReportLevel])
	jobsCTE, jobsCTEAliases, jobCTEJoinFields, jobsParams := getCTEAndParamsForJobsComplexStrategy(query, projectID, customerAccountID)

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
		value := mapOfTypeToJobCTEAlias[groupBy.Object] + "." + objectToValueInJobsMapping[key]
		groupByKeysWithoutTimestamp = append(groupByKeysWithoutTimestamp, value)
	}

	if isGroupByTimestamp {
		groupByStatement = joinWithComma(append(groupByKeysWithoutTimestamp, AliasDateTime)...)
	} else {
		groupByStatement = joinWithComma(groupByKeysWithoutTimestamp...)
	}

	selectKeys := groupByKeysWithoutTimestamp
	if isGroupByTimestamp {
		selectKeys = append(selectKeys, reportCTEAlias+"."+AliasDateTime)
	}

	for _, selectMetric := range reportSelectMetrics {
		value := fmt.Sprintf("%s(%s.%s)", adwordsMetricsToOperation[selectMetric], reportCTEAlias, selectMetric)
		selectMetrics = append(selectMetrics, value)
	}
	selectQuery += joinWithComma(append(selectKeys, selectMetrics...)...)
	orderByQuery := "ORDER BY " + getOrderByClause(selectMetrics)

	completeInnerJoin := " from " + reportCTEAlias + " "
	for index, jobsCTEAlias := range jobsCTEAliases {
		completeInnerJoin += innerJoinClause + jobsCTEAlias + " ON " + reportCTEAlias + "." + reportCTEJoinFields[index] + " = " + jobsCTEAlias + "." + jobCTEJoinFields[index] + " AND "
	}
	completeInnerJoin = completeInnerJoin[:len(completeInnerJoin)-4] + " "

	resultSQLStatement := completeWithClause + selectQuery + completeInnerJoin

	if len(groupByStatement) != 0 {
		resultSQLStatement += "GROUP BY " + groupByStatement
	}
	resultSQLStatement += " " + orderByQuery + channeAnalyticsLimit + ";"
	return ExecuteSQL(resultSQLStatement, params, logCtx)
}

// TODO handle duplicates of groupBy - edge case
// @Kark TODO v1
func getCTEAndParamsForReportComplexStrategy(query *ChannelQueryV1, projectID uint64,
	customerAccountID string, docType int) (string, string, []string, []string, []interface{}) {
	cteAlias := "reports_cte"
	customerAccountIDs := strings.Split(customerAccountID, ",")
	staticWhereParams := []interface{}{projectID, customerAccountIDs, docType, query.From, query.To}
	selectQuery := "WITH " + cteAlias + " as (SELECT "
	cteJoinFields := []string{}
	isGroupByTimestamp := query.GetGroupByTimestamp() != ""
	groupByKeysWithoutTimestamp := make([]string, 0, 0)
	groupByStatement := ""
	selectMetrics := make([]string, 0, 0)
	uniqueGroupByObjects := make(map[string]struct{})
	for _, groupBy := range query.GroupBy {
		uniqueGroupByObjects[groupBy.Object] = struct{}{}
	}

	for key := range uniqueGroupByObjects {
		key := key + ":id"
		value := objectAndPropertyToValueInReportsMapping[key]
		groupByKeysWithoutTimestamp = append(groupByKeysWithoutTimestamp, value)
		cteJoinFields = append(cteJoinFields, value)
	}

	if isGroupByTimestamp {
		groupByStatement = joinWithComma(append(groupByKeysWithoutTimestamp, AliasDateTime)...)
	} else {
		groupByStatement = joinWithComma(groupByKeysWithoutTimestamp...)
	}

	selectQuery += joinWithComma(groupByKeysWithoutTimestamp...)
	selectQuery = appendSelectTimestampIfRequired(selectQuery, query.GetGroupByTimestamp(), query.Timezone)

	for _, selectMetric := range query.SelectMetrics {
		selectMetrics = append(selectMetrics, adwordsRequestPropertiesToSQLproperty[selectMetric])
		currentSelectQuery := adwordsMetricsToAggregatesInReportsMapping[selectMetric] + " as " + adwordsRequestPropertiesToSQLproperty[selectMetric]
		selectQuery = joinWithComma(selectQuery, currentSelectQuery)
	}

	resultSQLStatement := selectQuery + fromAdwordsDocument + staticWhereStatement + " GROUP BY " + groupByStatement + " ), "

	return resultSQLStatement, cteAlias, selectMetrics, cteJoinFields, staticWhereParams
}

// @Kark TODO v1
func getCTEAndParamsForJobsComplexStrategy(query *ChannelQueryV1, projectID uint64, customerAccountID string) ([]string, []string, []string, []interface{}) {
	customerAccountIDs := strings.Split(customerAccountID, ",")

	campaignsFilters, adGroupFilters, _ := splitFiltersByObjectType(query)
	campaignsGroupBy, adGroupsGroupBy, _ := splitGroupByByObjectType(query)
	campaignJobCTE, campaignCTEAliasName, campaignJoinField, campaignParams := getCTEAndParamsForJob(query, projectID, customerAccountIDs, "campaign", campaignsFilters, campaignsGroupBy)
	adGroupJobCTE, adGroupCTEAliasName, adGroupJoinField, adGroupParams := getCTEAndParamsForJob(query, projectID, customerAccountIDs, "ad_group", adGroupFilters, adGroupsGroupBy)
	resultParams := append(make([]interface{}, 0, 0), campaignParams...)
	resultParams = append(resultParams, adGroupParams...)
	return U.AppendNonNullValues(campaignJobCTE, adGroupJobCTE), U.AppendNonNullValues(campaignCTEAliasName, adGroupCTEAliasName), U.AppendNonNullValues(campaignJoinField, adGroupJoinField), resultParams
}

// @Kark TODO v1
func getCTEAndParamsForJob(query *ChannelQueryV1, projectID uint64, customerAccountIDs []string, objectType string, filters []FilterV1, groupBy []GroupBy) (string, string, string, []interface{}) {
	if len(groupBy) < 1 {
		return "", "", "", make([]interface{}, 0, 0)
	}
	docType := AdwordsDocumentTypeAlias[objectType+"s"]
	staticWhereParams := []interface{}{projectID, customerAccountIDs, docType, query.From, query.To}
	aliasName := mapOfTypeToJobCTEAlias[objectType]
	withClause := aliasName + " as ("

	objectID := objectToValueInJobsMapping[objectType+":"+"id"]

	table1SQL, table1Alias, table1ColumnName, table1Params := getIDAndMaxTimeSQLAndParams(query, staticWhereParams, objectType, filters)
	table2SQL, table2Alias, table2ColumnName, table2Params := getCompleteRowSQLAndParamsForJob(query, staticWhereParams, objectType, filters)
	cteJoinField := objectID
	groupByQuery := getSelectPropertiesExceptIDsForJob(groupBy)
	selectQuery := "SELECT " + table1Alias + "." + objectID + " as " + objectID + ", " + groupByQuery

	resultStatement := withClause + selectQuery + " from " + table1SQL + innerJoinClause + table2SQL + " ON " + table1Alias + "." + table1ColumnName + " = " + table2Alias + "." + table2ColumnName + " ), "
	resultParams := append(make([]interface{}, 0, 0), table1Params...)
	resultParams = append(resultParams, table2Params...)
	return resultStatement, aliasName, cteJoinField, resultParams
}

// @Kark TODO v1
func appendSelectMetrics(selectQuery string, selectMetrics []string) ([]string, string) {
	selectKeys := make([]string, 0, 0)
	for _, key := range selectMetrics {
		value := adwordsMetricsToAggregatesInReportsMapping[key]
		selectMetrics = append(selectKeys, value)
		selectQuery = joinWithComma(selectQuery, value)
	}
	return selectMetrics, selectQuery
}

// @Kark TODO v1
func getOrderByClause(selectMetrics []string) string {
	selectMetricsWithDesc := make([]string, 0, 0)
	for _, selectMetric := range selectMetrics {
		selectMetricsWithDesc = append(selectMetricsWithDesc, selectMetric+" DESC")
	}
	return joinWithComma(selectMetricsWithDesc...)
}

// @Kark TODO v1
func getSelectPropertiesExceptIDsForJob(groupBys []GroupBy) string {
	groupByQuery := ""
	for _, groupBy := range groupBys {
		key := groupBy.Object + ":" + groupBy.Property
		if groupBy.Property != "id" {
			groupByQuery += "value->>'" + objectToValueInJobsMapping[key] + "' as " + objectToValueInJobsMapping[key] + ", "
		}
	}
	return groupByQuery[:len(groupByQuery)-2]
}

// @Kark TODO v1
func getIDAndMaxTimeSQLAndParams(query *ChannelQueryV1, staticWhereParams []interface{}, objectType string, filters []FilterV1) (string, string, string, []interface{}) {
	aliasName := "distinct_id"
	idColumnName := objectType + "_id"
	selectStatement := "(SELECT " + idColumnName + " , max(timestamp)" + fromAdwordsDocument
	groupByStatement := "GROUP BY " + idColumnName + ") "
	sqlParams := staticWhereParams
	filterPropertiesStatement, filterParams := getFilterPropertiesForJobStatementAndParams(filters)
	completeFiltersStatement := staticWhereStatement
	if len(filterPropertiesStatement) != 0 {
		completeFiltersStatement += "AND " + filterPropertiesStatement + " "
		sqlParams = append(sqlParams, filterParams...)
	}
	resultStatement := selectStatement + completeFiltersStatement + groupByStatement + "as " + aliasName
	return resultStatement, aliasName, idColumnName, sqlParams
}

// @Kark TODO v1
func getCompleteRowSQLAndParamsForJob(query *ChannelQueryV1, staticWhereParams []interface{}, objectType string, filters []FilterV1) (string, string, string, []interface{}) {
	aliasName := "JobRecords"
	idColumnName := objectType + "_id"
	selectStatement := "(SELECT * FROM adwords_documents "
	resultStatement := selectStatement + staticWhereStatement + ") as " + aliasName
	return resultStatement, aliasName, idColumnName, staticWhereParams
}

// @Kark TODO v1
func splitFiltersByObjectType(query *ChannelQueryV1) ([]FilterV1, []FilterV1, []FilterV1) {
	campaignsFilters := make([]FilterV1, 0, 0)
	adGroupFilters := make([]FilterV1, 0, 0)
	adFilters := make([]FilterV1, 0, 0)

	for _, filter := range query.Filters {
		switch filter.Object {
		case "campaign":
			campaignsFilters = append(campaignsFilters, filter)
		case "ad_groups":
			adGroupFilters = append(adGroupFilters, filter)
		case "ad":
			adFilters = append(adFilters, filter)
		}
	}
	return campaignsFilters, adGroupFilters, adFilters
}

// @Kark TODO v1
func splitGroupByByObjectType(query *ChannelQueryV1) ([]GroupBy, []GroupBy, []GroupBy) {
	campaignsGroupBys := make([]GroupBy, 0, 0)
	adGroupGroupBys := make([]GroupBy, 0, 0)
	adGroupBys := make([]GroupBy, 0, 0)

	for _, filter := range query.GroupBy {
		switch filter.Object {
		case "campaign":
			campaignsGroupBys = append(campaignsGroupBys, filter)
		case "ad_groups":
			adGroupGroupBys = append(adGroupGroupBys, filter)
		case "ad":
			adGroupBys = append(adGroupBys, filter)
		}
	}
	return campaignsGroupBys, adGroupGroupBys, adGroupBys
}

// @Kark TODO v1
// TODO Check if we have none operator
func getFilterPropertiesForJobStatementAndParams(filters []FilterV1) (string, []interface{}) {
	resultStatement := ""
	params := make([]interface{}, 0, 0)
	for index, filter := range filters {
		currentFilterStatement := ""
		if filter.LogicalOp == "" {
			filter.LogicalOp = "AND"
		}
		filter.Condition = getOp(filter.Condition)
		currentFilterStatement = fmt.Sprintf("value->>? %s ?", filter.Condition)
		params = append(params, filter.Property, filter.Value)

		if index == 0 {
			resultStatement = currentFilterStatement
		} else {
			resultStatement = fmt.Sprintf("%s %s %s", resultStatement, filter.LogicalOp, currentFilterStatement)
		}
	}
	return resultStatement, params
}

// @Kark TODO v1
func hasAllIDsOnlyInGroupBy(query *ChannelQueryV1) bool {
	for _, groupBy := range query.GroupBy {
		if !(strings.Contains(groupBy.Property, "id") || strings.Contains(groupBy.Property, "ID")) {
			return false
		}
	}
	return true
}

// ExecuteSQL - @Kark TODO v1
func ExecuteSQL(sqlStatement string, params []interface{}, logCtx *log.Entry) ([]string, [][]interface{}, error) {
	db := C.GetServices().Db
	rows, err := db.Raw(sqlStatement, params...).Rows()
	if err != nil {
		logCtx.WithError(err).Error("SQL Query failed")
		return nil, nil, err
	}

	defer rows.Close()
	columns, resultRows, err := U.DBReadRows(rows)

	if err != nil {
		return nil, nil, err
	}
	if len(resultRows) == 0 {
		log.Error("Aggregate query returned zero rows.")
		return nil, nil, errors.New("no rows returned")
	}
	return columns, resultRows, nil
}

// @TODO Kark v0
func getAdwordsChannelResultMeta(projectID uint64, customerAccountID string,
	query *ChannelQuery) (*ChannelQueryResultMeta, error) {

	customerAccountIDArray := strings.Split(customerAccountID, ",")
	stmnt := "SELECT value->>'currency_code' as currency FROM adwords_documents" +
		" " + "WHERE project_id=? AND customer_account_id IN (?) AND type=? AND timestamp BETWEEN ? AND ?" +
		" " + "ORDER BY timestamp DESC LIMIT 1"

	logCtx := log.WithField("project_id", projectID)

	db := C.GetServices().Db
	rows, err := db.Raw(stmnt, projectID, customerAccountIDArray,
		AdwordsDocumentTypeAlias["customer_account_properties"],
		GetAdwordsDateOnlyTimestamp(query.From),
		GetAdwordsDateOnlyTimestamp(query.To)).Rows()
	if err != nil {
		logCtx.WithError(err).Error("Failed to build meta for channel query result.")
		return nil, err
	}
	defer rows.Close()

	var currency string
	for rows.Next() {
		rows.Scan(&currency)
	}

	err = rows.Err()
	if err != nil {
		logCtx.WithError(err).Error("Failed to build meta for channel query result.")
		return nil, err
	}

	return &ChannelQueryResultMeta{Currency: currency}, nil
}

// ExecuteAdwordsChannelQuery - @TODO Kark v0
func ExecuteAdwordsChannelQuery(projectID uint64, query *ChannelQuery) (*ChannelQueryResult, int) {
	logCtx := log.WithField("project_id", projectID).WithField("query", query)

	if projectID == 0 || query == nil {
		logCtx.Error("Invalid project_id or query on execute adwords channel query.")
		return nil, http.StatusInternalServerError
	}

	projectSetting, errCode := GetProjectSetting(projectID)
	if errCode != http.StatusFound {
		return nil, http.StatusInternalServerError
	}

	if projectSetting.IntAdwordsCustomerAccountId == nil || *projectSetting.IntAdwordsCustomerAccountId == "" {
		logCtx.Error("Execute adwords channel query failed. No customer account id.")
		return nil, http.StatusInternalServerError
	}

	queryResult := &ChannelQueryResult{}
	meta, err := getAdwordsChannelResultMeta(projectID,
		*projectSetting.IntAdwordsCustomerAccountId, query)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get adwords channel result meta.")
		return queryResult, http.StatusInternalServerError
	}
	queryResult.Meta = meta

	metricKvs, err := getAdwordsMetrics(projectID, *projectSetting.IntAdwordsCustomerAccountId, query)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get adwords metric kvs.")
		return queryResult, http.StatusInternalServerError
	}
	queryResult.Metrics = metricKvs

	// Return, if no breakdown.
	if query.Breakdown == "" {
		return queryResult, http.StatusOK
	}

	metricBreakdown, err := getAdwordsMetricsBreakdown(projectID,
		*projectSetting.IntAdwordsCustomerAccountId, query)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get adwords metric breakdown.")
		return queryResult, http.StatusInternalServerError
	}
	queryResult.MetricsBreakdown = metricBreakdown

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

// GetAdwordsFilterPropertyKeyByType - @TODO Kark v0
func GetAdwordsFilterPropertyKeyByType(docType int) (string, error) {
	filterKeyByType := map[int]string{
		5:  "campaign_name",
		10: "ad_group_name",
		8:  "criteria",
		6:  "id",
	}

	filterKey, filterKeyExists := filterKeyByType[docType]
	if !filterKeyExists {
		return "", errors.New("no filter key found for document type")
	}

	return filterKey, nil
}

// GetAdwordsFilterValuesByType - @TODO Kark v0
func GetAdwordsFilterValuesByType(projectID uint64, docType int) ([]string, int) {
	projectSetting, errCode := GetProjectSetting(projectID)
	if errCode != http.StatusFound {
		return []string{}, http.StatusInternalServerError
	}
	customerAccountID := projectSetting.IntAdwordsCustomerAccountId

	db := C.GetServices().Db
	logCtx := log.WithField("project_id", projectID).WithField("doc_type", docType)

	filterValueKey, err := GetAdwordsFilterPropertyKeyByType(docType)
	if err != nil {
		logCtx.WithError(err).Error("Unknown doc type for get adwords filter key.")
		return []string{}, http.StatusBadRequest
	}

	queryStr := "SELECT DISTINCT(value->>?) as filter_value FROM adwords_documents WHERE project_id = ? AND" +
		" " + "customer_account_id = ? AND type = ? LIMIT 5000"
	rows, err := db.Raw(queryStr, filterValueKey, projectID, customerAccountID, docType).Rows()
	if err != nil {
		logCtx.WithError(err).Error("Failed to distinct filter values by type from adwords documents.")
		return []string{}, http.StatusInternalServerError
	}
	defer rows.Close()

	filterValues := make([]string, 0, 0)
	for rows.Next() {
		var filterValue string
		if err := rows.Scan(&filterValue); err != nil {
			logCtx.WithError(err).Error("Failed to distinct filter values by type from adwords documents.")
			return filterValues, http.StatusInternalServerError
		}

		filterValues = append(filterValues, filterValue)
	}

	return filterValues, http.StatusFound
}

// GetAdwordsDocumentTypeForFilterKey - @TODO Kark v0
func GetAdwordsDocumentTypeForFilterKey(filter string) (int, error) {
	var docType int

	switch filter {
	case CAFilterCampaign:
		docType = AdwordsDocumentTypeAlias["campaign_performance_report"]
	case CAFilterAd:
		docType = AdwordsDocumentTypeAlias["ad_performance_report"]
	case CAFilterKeyword:
		docType = AdwordsDocumentTypeAlias["keyword_performance_report"]
	case CAFilterAdGroup:
		docType = AdwordsDocumentTypeAlias["ad_group_performance_report"]
	}

	if docType == 0 {
		return docType, errors.New("no adwords document type for filter")
	}

	return docType, nil
}

/*
GetAdwordsMetricsQuery
SELECT value->>'criteria', SUM((value->>'impressions')::float) as impressions, SUM((value->>'clicks')::float) as clicks,
SUM((value->>'cost')::float) as total_cost, SUM((value->>'conversions')::float) as all_conversions,
SUM((value->>'all_conversions')::float) as all_conversions FROM adwords_documents
WHERE type='5' AND timestamp BETWEEN '20191122' and '20191129' AND value->>'campaign_name'='Desktop Only'
GROUP BY value->>'criteria';
*/
func getAdwordsMetricsQuery(projectID uint64, customerAccountID string, query *ChannelQuery,
	withBreakdown bool) (string, []interface{}, error) {

	customerAccountIDArray := strings.Split(customerAccountID, ",")
	// select handling.
	selectColstWithoutAlias := "SUM((value->>'impressions')::float) as %s, SUM((value->>'clicks')::float) as %s," +
		" " + "SUM((value->>'cost')::float)/1000000 as %s, SUM((value->>'conversions')::float) as %s," +
		" " + "SUM((value->>'all_conversions')::float) as %s," +
		" " + "SUM((value->>'cost')::float)/NULLIF(SUM((value->>'clicks')::float), 0)/1000000 as %s," +
		" " + "SUM((value->>'clicks')::float * regexp_replace(value->>'conversion_rate', ?, '')::float)/NULLIF(SUM((value->>'clicks')::float), 0) as %s," +
		" " + "SUM((value->>'cost')::float)/NULLIF(SUM((value->>'conversions')::float), 0)/1000000 as %s"
	selectCols := fmt.Sprintf(selectColstWithoutAlias, CAColumnImpressions, CAColumnClicks,
		CAColumnTotalCost, CAColumnConversions, CAColumnAllConversions,
		CAColumnCostPerClick, CAColumnConversionRate, CAColumnCostPerConversion)

	paramsSelect := make([]interface{}, 0, 0)

	// Where handling.
	stmntWhere := "WHERE project_id=? AND customer_account_id IN (?) AND type=? AND timestamp BETWEEN ? AND ?"
	paramsWhere := make([]interface{}, 0, 0)

	docType, err := GetAdwordsDocumentTypeForFilterKey(query.FilterKey)
	if err != nil {
		return "", []interface{}{}, err
	}

	paramsWhere = append(paramsWhere, projectID, customerAccountIDArray, docType,
		GetAdwordsDateOnlyTimestamp(query.From), GetAdwordsDateOnlyTimestamp(query.To))

	isWhereByFilterRequired := query.FilterValue != filterValueAll
	if isWhereByFilterRequired {
		stmntWhere = stmntWhere + " " + "AND" + " " + "value->>?=?"

		filterKey, err := GetAdwordsFilterPropertyKeyByType(docType)
		if err != nil {
			return "", []interface{}{}, err
		}

		paramsWhere = append(paramsWhere, filterKey, query.FilterValue)
	}

	// group by handling.
	var stmntGroupBy string
	paramsGroupBy := make([]interface{}, 0, 0)
	if withBreakdown {
		// Todo: Use a seperate or a generic method for getting property key to group by
		// for a specific key type. Now using method which does same for filterKey
		// for breakdownKey. Say campaigns, group by campaign_name.
		docType, err := GetAdwordsDocumentTypeForFilterKey(query.Breakdown)
		if err != nil {
			log.WithError(err).Error("Failed to get adwords doc type by filter key.")
			return "", []interface{}{}, err
		}
		propertyKey, err := GetAdwordsFilterPropertyKeyByType(docType)
		if err != nil {
			log.WithError(err).Error("Failed to get filter propery key by type.")
			return "", []interface{}{}, err
		}

		// prepend group by col on select.
		selectCols = "value->>? as %s" + ", " + selectCols
		selectCols = fmt.Sprintf(selectCols, CAChannelGroupKey)
		paramsSelect = append(paramsSelect, propertyKey)

		stmntGroupBy = "GROUP BY" + " " + "%s"
		stmntGroupBy = fmt.Sprintf(stmntGroupBy, CAChannelGroupKey)
	}

	// Using prepared statement to replace '%', to avoid
	// query breakage with "!%(MISSING)" on gorm.
	paramsSelect = append(paramsSelect, "%")

	params := make([]interface{}, 0, 0)

	stmnt := "SELECT" + " " + selectCols + " " + "FROM adwords_documents" + " " + stmntWhere + " " + stmntGroupBy
	params = append(params, paramsSelect...)
	params = append(params, paramsWhere...)
	params = append(params, paramsGroupBy...)

	return stmnt, params, nil
}

// @TODO Kark v0
func getAdwordsMetrics(projectID uint64, customerAccountID string,
	query *ChannelQuery) (*map[string]interface{}, error) {

	stmnt, params, err := getAdwordsMetricsQuery(projectID, customerAccountID, query, false)
	if err != nil {
		return nil, err
	}

	db := C.GetServices().Db
	rows, err := db.Raw(stmnt, params...).Rows()
	if err != nil {
		return nil, err
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

	return &metricKvs, nil
}

// @TODO Kark v0
func getAdwordsMetricsBreakdown(projectID uint64, customerAccountID string,
	query *ChannelQuery) (*ChannelBreakdownResult, error) {

	logCtx := log.WithField("project_id", projectID).WithField("customer_account_id", customerAccountID)

	stmnt, params, err := getAdwordsMetricsQuery(projectID, customerAccountID, query, true)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get adwords metrics query.")
		return nil, err
	}

	db := C.GetServices().Db
	rows, err := db.Raw(stmnt, params...).Rows()
	if err != nil {
		return nil, err
	}

	resultHeaders, resultRows, err := U.DBReadRows(rows)
	if err != nil {
		return nil, err
	}

	// Translate group key.
	for i := range resultHeaders {
		if resultHeaders[i] == CAChannelGroupKey {
			resultHeaders[i] = query.Breakdown
		}
	}

	// Fill null with zero for aggr.
	// Do I need to show this as NA?
	for ri := range resultRows {
		for ci := range resultRows[ri] {
			// if not group key and nil: zero.
			if ci > 0 && resultRows[ri][ci] == nil {
				resultRows[ri][ci] = float64(0)
			}
		}
	}

	return &ChannelBreakdownResult{Headers: resultHeaders, Rows: resultRows}, nil
}
