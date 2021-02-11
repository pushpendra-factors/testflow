package postgres

import (
	"errors"
	C "factors/config"
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

var selectableMetricsForAdwords = []string{"conversion"}
var objectsForAdwords = []string{model.AdwordsKeyword}

var keywordsPropertyToRelated = map[string]PropertiesAndRelated{}

var mapOfObjectsToPropertiesAndRelated = map[string]map[string]PropertiesAndRelated{
	model.AdwordsKeyword: {"id": PropertiesAndRelated{typeOfProperty: U.PropertyTypeCategorical}},
}

// TODO Check if this is being used in proper places within adwords.
var objectAndPropertyToValueInsideAdwordsReportsMapping = map[string]string{
	"campaign:id":   "campaign_id",
	"ad_group:id":   "ad_group_id",
	"ad:id":         "ad_id",
	"keyword:id":    "keyword_id",
	"campaign:name": "campaign_name",
	"ad_group:name": "ad_group_name",
}

var adwordsMetricsToAggregatesInReportsMapping = map[string]string{
	"impressions": "SUM((value->>'impressions')::float)",
	"clicks":      "SUM((value->>'clicks')::float)",
	"cost":        "SUM((value->>'cost')::float)/1000000",
	// "cost_per_click": "average_cost",
	"conversions": "SUM((value->>'conversions')::float)",
	// "conversion_rate": "conversion_rate"
}

var adwordsMetricsToOperation = map[string]string{
	"impressions": "sum",
	"clicks":      "sum",
	"cost":        "sum",
	"conversions": "sum",
}

/*
TODO. Decided on using a common representation of exposed fields for the cte and simple query.
We use query.GroupBy in places where join of reports and meta happen. This should be avoided.
*/
var propertyToExposedValueFromCTE = map[string]string{
	"campaign:id":   "campaign_id",
	"campaign:name": "campaign_name",
	"ad_group:id":   "ad_group_id",
	"ad_group:name": "ad_group_name",
	"ad:id":         "ad_id",
	"keyword:id":    "keyword_id",
}

var objectAndPropertyToValueInsideAdwordsJobsMapping = map[string]map[string]string{
	"campaign": {
		"campaign:id":   "campaign_id",
		"campaign:name": "name",
	},
	"ad_group": {
		"ad_group:id":   "ad_group_id",
		"ad_group:name": "name",
		"campaign:id":   "campaign_id",
		"campaign:name": "campaign_name",
	},
	"ad": {
		"id": "ad_id",
	},
}

var adwordsExternalRepresentationToInternalRepresentation = map[string]string{
	"name":        "name",
	"id":          "id",
	"impressions": "impressions",
	"clicks":      "clicks",
	"spend":       "cost",
	"conversion":  "conversions",
	"campaign":    "campaign",
	"ad_group":    "ad_group",
	"ad":          "ad",
	"keyword":     "keyword",
}

var adwordsInternalRepresentationToExternalRepresentation = map[string]string{
	"impressions":   "impressions",
	"clicks":        "clicks",
	"cost":          "spend",
	"conversions":   "conversion",
	"campaign:id":   "campaign_id",
	"campaign:name": "campaign_name",
	"ad_group:id":   "ad_group_id",
	"ad_group:name": "ad_group_name",
	"ad:id":         "ad_id",
	"keyword:id":    "keyword_id",
}

var errorEmptyAdwordsDocument = errors.New("empty adwords document")

const (
	errorDuplicateAdwordsDocument  = "pq: duplicate key value violates unique constraint \"adwords_documents_pkey\""
	filterValueAll                 = "all"
	adwordsFilterQueryStr          = "SELECT DISTINCT(value->>?) as filter_value FROM adwords_documents WHERE project_id = ? AND" + " " + "customer_account_id = ? AND type = ? AND value->>? IS NOT NULL LIMIT 5000"
	staticWhereStatementForAdwords = "WHERE project_id = ? AND customer_account_id IN ( ? ) AND type = ? AND timestamp between ? AND ? "
	fromAdwordsDocument            = " FROM adwords_documents "
)

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
	case model.AdwordsDocumentTypeAlias["campaigns"]:
		return U.GetInt64FromMapOfInterface(*valueMap, "id", 0), 0, 0, 0
	case model.AdwordsDocumentTypeAlias["ad_groups"]:
		return U.GetInt64FromMapOfInterface(*valueMap, "campaign_id", 0), U.GetInt64FromMapOfInterface(*valueMap, "id", 0), 0, 0
	case model.AdwordsDocumentTypeAlias["click_performance_report"], model.AdwordsDocumentTypeAlias["search_performance_report"], model.AdwordsDocumentTypeAlias["ad_group_performance_report"]:
		return U.GetInt64FromMapOfInterface(*valueMap, "campaign_id", 0), U.GetInt64FromMapOfInterface(*valueMap, "ad_group_id", 0), 0, 0
	case model.AdwordsDocumentTypeAlias["campaign_performance_report"]:
		return U.GetInt64FromMapOfInterface(*valueMap, "campaign_id", 0), 0, 0, 0
	case model.AdwordsDocumentTypeAlias["ad_performance_report"]:
		return U.GetInt64FromMapOfInterface(*valueMap, "campaign_id", 0), U.GetInt64FromMapOfInterface(*valueMap, "ad_group_id", 0), U.GetInt64FromMapOfInterface(*valueMap, "id", 0), 0
	case model.AdwordsDocumentTypeAlias["keyword_performance_report"]:
		return U.GetInt64FromMapOfInterface(*valueMap, "campaign_id", 0), U.GetInt64FromMapOfInterface(*valueMap, "ad_group_id", 0), 0, U.GetInt64FromMapOfInterface(*valueMap, "id", 0)
	case model.AdwordsDocumentTypeAlias["customer_account_properties"]:
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

func getAdwordsDateOnlyTimestampInInt64(unixTimestamp int64) int64 {
	value, _ := strconv.ParseInt(time.Unix(unixTimestamp, 0).UTC().Format("20060102"), 10, 64)
	return value
}

func getAdwordsIDAndHeirarchyColumnsByType(docType int, valueJSON *postgres.Jsonb) (string, int64, int64, int64, int64, error) {
	if docType > len(model.AdwordsDocumentTypeAlias) {
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

	value1, value2, value3, value4 := getAdwordsHierarchyColumnsByType(valueMap, docType)

	// ID as string always.
	return idStr, value1, value2, value3, value4, nil
}

// CreateAdwordsDocument ...
func (pg *Postgres) CreateAdwordsDocument(adwordsDoc *model.AdwordsDocument) int {
	logCtx := log.WithField("customer_acc_id", adwordsDoc.CustomerAccountID).WithField(
		"project_id", adwordsDoc.ProjectID)

	if adwordsDoc.CustomerAccountID == "" || adwordsDoc.TypeAlias == "" {
		logCtx.Error("Invalid adwords document.")
		return http.StatusBadRequest
	}

	logCtx = logCtx.WithField("type_alias", adwordsDoc.TypeAlias)
	docType, docTypeExists := model.AdwordsDocumentTypeAlias[adwordsDoc.TypeAlias]
	if !docTypeExists {
		logCtx.Error("Invalid type alias.")
		return http.StatusBadRequest
	}
	adwordsDoc.Type = docType

	adwordsDocID, campaignIDValue, adGroupIDValue, adIDValue,
		keywordIDValue, err := getAdwordsIDAndHeirarchyColumnsByType(adwordsDoc.Type, adwordsDoc.Value)
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
	currentTime := gorm.NowFunc()

	db := C.GetServices().Db
	// TODO: Use gorm.Create method, instead of INSERT query string.
	queryStr := "INSERT INTO adwords_documents (project_id,customer_account_id,type,timestamp,id,campaign_id,ad_group_id,ad_id,keyword_id,value,created_at,updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
	rows, err := db.Raw(queryStr, adwordsDoc.ProjectID, adwordsDoc.CustomerAccountID,
		adwordsDoc.Type, adwordsDoc.Timestamp, adwordsDoc.ID, campaignIDValue, adGroupIDValue, adIDValue, keywordIDValue, adwordsDoc.Value, currentTime, currentTime).Rows()
	if err != nil {
		if isDuplicateAdwordsDocumentError(err) {
			logCtx.WithError(err).WithField("timestamp", adwordsDoc.Timestamp).WithField("id", adwordsDoc.ID).
				WithField("createdAt", currentTime).Error("Failed to create an adwords doc. Duplicate.")
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

func getDocumentTypeAliasByType() map[int]string {
	documentTypeMap := make(map[int]string, 0)
	for alias, typ := range model.AdwordsDocumentTypeAlias {
		documentTypeMap[typ] = alias
	}

	return documentTypeMap
}

// GetAllAdwordsLastSyncInfoByProjectCustomerAccountAndType - @TODO Kark v1
func (pg *Postgres) GetAllAdwordsLastSyncInfoByProjectCustomerAccountAndType() ([]model.AdwordsLastSyncInfo, int) {
	db := C.GetServices().Db

	adwordsLastSyncInfos := make([]model.AdwordsLastSyncInfo, 0, 0)

	queryStr := "SELECT project_id, customer_account_id, type as document_type, max(timestamp) as last_timestamp" +
		" " + "FROM adwords_documents GROUP BY project_id, customer_account_id, type"

	rows, err := db.Raw(queryStr).Rows()
	if err != nil {
		log.WithError(err).Error("Failed to get last adwords documents by type for sync info.")
		return adwordsLastSyncInfos, http.StatusInternalServerError
	}
	defer rows.Close()

	for rows.Next() {
		var adwordsLastSyncInfo model.AdwordsLastSyncInfo
		if err := db.ScanRows(rows, &adwordsLastSyncInfo); err != nil {
			log.WithError(err).Error("Failed to scan last adwords documents by type for sync info.")
			return []model.AdwordsLastSyncInfo{}, http.StatusInternalServerError
		}

		adwordsLastSyncInfos = append(adwordsLastSyncInfos, adwordsLastSyncInfo)
	}

	adwordsSettings, errCode := pg.GetAllIntAdwordsProjectSettings()
	if errCode != http.StatusOK {
		return []model.AdwordsLastSyncInfo{}, errCode
	}

	adwordsSettingsByProjectAndCustomerAccount := make(map[uint64]map[string]*model.AdwordsProjectSettings, 0)

	for i := range adwordsSettings {
		customerAccountIDs := strings.Split(adwordsSettings[i].CustomerAccountId, ",")
		adwordsSettingsByProjectAndCustomerAccount[adwordsSettings[i].ProjectId] = make(map[string]*model.AdwordsProjectSettings)
		for j := range customerAccountIDs {
			var setting model.AdwordsProjectSettings
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
	selectedLastSyncInfos := make([]model.AdwordsLastSyncInfo, 0, 0)

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
			for docTypeAlias := range model.AdwordsDocumentTypeAlias {
				if !accountExists || (accountExists && existingTypesForAccount[docTypeAlias] == false) {
					syncInfo := model.AdwordsLastSyncInfo{
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
func (pg *Postgres) GetGCLIDBasedCampaignInfo(projectID uint64, from, to int64, adwordsAccountIDs string) (map[string]model.CampaignInfo, error) {

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
	rows, err := pg.ExecQueryWithContext(performanceQuery, []interface{}{model.PropertyValueNone, model.PropertyValueNone, model.PropertyValueNone, model.PropertyValueNone,
		model.PropertyValueNone, model.PropertyValueNone, projectID, customerAccountIDs, AdwordsClickReportType, U.GetDateOnlyFromTimestamp(from),
		U.GetDateOnlyFromTimestamp(to)})
	if err != nil {
		logCtx.WithError(err).Error("SQL Query failed")
		return nil, err
	}
	defer rows.Close()
	gclIDBasedCampaign := make(map[string]model.CampaignInfo)
	for rows.Next() {
		var gclID string
		var adgroupName string
		var campaignName string
		var adID string
		if err = rows.Scan(&gclID, &adgroupName, &campaignName, &adID); err != nil {
			logCtx.WithError(err).Error("SQL Parse failed")
			continue
		}
		gclIDBasedCampaign[gclID] = model.CampaignInfo{
			AdgroupName:  adgroupName,
			CampaignName: campaignName,
			AdID:         adID,
		}
	}
	return gclIDBasedCampaign, nil
}

// @TODO Kark v1
func buildAdwordsChannelConfig() *model.ChannelConfigResult {
	properties := buildProperties(allChannelsPropertyToRelated)
	adwordsObjectsAndProperties := buildObjectAndPropertiesForAdwords(objectsForAdwords)
	commonObjectsAndProperties := buildObjectsAndProperties(properties, objectsForAllChannels)
	selectMetrics := append(selectableMetricsForAllChannels, selectableMetricsForAdwords...)
	objectsAndProperties := append(adwordsObjectsAndProperties, commonObjectsAndProperties...)
	return &model.ChannelConfigResult{
		SelectMetrics:        selectMetrics,
		ObjectsAndProperties: objectsAndProperties,
	}
}

func buildObjectAndPropertiesForAdwords(objects []string) []model.ChannelObjectAndProperties {
	objectsAndProperties := make([]model.ChannelObjectAndProperties, 0, 0)
	for _, currentObject := range objects {
		propertiesAndRelated, isPresent := mapOfObjectsToPropertiesAndRelated[currentObject]
		var currentProperties []model.ChannelProperty
		if isPresent {
			currentProperties = buildProperties(propertiesAndRelated)
		} else {
			currentProperties = buildProperties(allChannelsPropertyToRelated)
		}
		objectsAndProperties = append(objectsAndProperties, buildObjectsAndProperties(currentProperties, []string{currentObject})...)
	}
	return objectsAndProperties
}

// GetAdwordsFilterValues - @TODO Kark v1
func (pg *Postgres) GetAdwordsFilterValues(projectID uint64, requestFilterObject string, requestFilterProperty string, reqID string) ([]interface{}, int) {
	adwordsInternalFilterProperty, docType, err := getFilterRelatedInformationForAdwords(requestFilterObject, requestFilterProperty)
	if err != http.StatusOK {
		return make([]interface{}, 0, 0), http.StatusBadRequest
	}
	filterValues, errCode := pg.getAdwordsFilterValuesByType(projectID, docType, adwordsInternalFilterProperty, reqID)
	if errCode != http.StatusFound {
		return []interface{}{}, http.StatusInternalServerError
	}

	return filterValues, http.StatusFound
}

// GetAdwordsSQLQueryAndParametersForFilterValues - @TODO Kark v1
// Currently, properties in object dont vary with Object.
func (pg *Postgres) GetAdwordsSQLQueryAndParametersForFilterValues(projectID uint64, requestFilterObject string, requestFilterProperty string) (string, []interface{}, int) {
	adwordsInternalFilterProperty, docType, err := getFilterRelatedInformationForAdwords(requestFilterObject, requestFilterProperty)
	if err != http.StatusOK {
		return "", make([]interface{}, 0, 0), http.StatusBadRequest
	}
	projectSetting, errCode := pg.GetProjectSetting(projectID)
	if errCode != http.StatusFound {
		return "", []interface{}{}, http.StatusInternalServerError
	}
	customerAccountID := projectSetting.IntAdwordsCustomerAccountId

	params := []interface{}{adwordsInternalFilterProperty, projectID, customerAccountID, docType, adwordsInternalFilterProperty}
	return "(" + adwordsFilterQueryStr + ")", params, http.StatusFound
}

func getFilterRelatedInformationForAdwords(requestFilterObject string, requestFilterProperty string) (string, int, int) {
	adwordsInternalFilterObject, isPresent := adwordsExternalRepresentationToInternalRepresentation[requestFilterObject]
	if !isPresent {
		log.Error("Invalid adwords filter object.")
		return "", 0, http.StatusBadRequest
	}
	adwordsInternalFilterProperty, isPresent := adwordsExternalRepresentationToInternalRepresentation[requestFilterProperty]
	if !isPresent {
		log.Error("Invalid adwords filter property.")
		return "", 0, http.StatusBadRequest
	}
	docType := getAdwordsDocumentTypeForFilterKeyV1(adwordsInternalFilterObject)

	return adwordsInternalFilterProperty, docType, http.StatusOK
}

// @TODO Kark v1
func (pg *Postgres) getAdwordsFilterValuesByType(projectID uint64, docType int, property string, reqID string) ([]interface{}, int) {
	logCtx := log.WithField("req_id", reqID)
	projectSetting, errCode := pg.GetProjectSetting(projectID)
	if errCode != http.StatusFound {
		logCtx.Error("Failed to fetch Project Setting in adwords filter values.")
		return []interface{}{}, http.StatusInternalServerError
	}
	customerAccountID := projectSetting.IntAdwordsCustomerAccountId

	logCtx = log.WithField("project_id", projectID).WithField("doc_type", docType).WithField("req_id", reqID)
	params := []interface{}{property, projectID, customerAccountID, docType, property}
	_, resultRows, _ := pg.ExecuteSQL(adwordsFilterQueryStr, params, logCtx)

	return Convert2DArrayTo1DArray(resultRows), http.StatusFound
}

// @TODO Kark v1
// This method uses internal filterObject as input param and not request filterObject.
// Note: method not to be used without proper validation of request params.
func getAdwordsDocumentTypeForFilterKeyV1(filterObject string) int {
	var docType int

	switch filterObject {
	case model.AdwordsCampaign:
		docType = model.AdwordsDocumentTypeAlias[model.AdwordsCampaign+"s"]
	case model.AdwordsAdGroup:
		docType = model.AdwordsDocumentTypeAlias[model.AdwordsAdGroup+"s"]
	case model.AdwordsAd:
		docType = model.AdwordsDocumentTypeAlias[model.AdwordsAd+"s"]
	case model.AdwordsKeyword:
		docType = model.AdwordsDocumentTypeAlias[model.KeywordPerformanceReport]
	}

	return docType
}

// ExecuteAdwordsmodel.ChannelQueryV1 - @TODO Kark v1.
// Job represents the meta data associated with particular object type. Reports represent data with metrics and few filters.
// TODO - Duplicate code/flow in facebook and adwords.
func (pg *Postgres) ExecuteAdwordsChannelQueryV1(projectID uint64, query *model.ChannelQueryV1, reqID string) ([]string, [][]interface{}, error) {
	fetchSource := false
	logCtx := log.WithField("xreq_id", reqID)
	sql, params, selectKeys, selectMetrics, err := pg.GetSQLQueryAndParametersForAdwordsQueryV1(
		projectID, query, reqID, fetchSource)
	if err != nil {
		return make([]string, 0, 0), make([][]interface{}, 0, 0), err
	}
	_, resultMetrics, err := pg.ExecuteSQL(sql, params, logCtx)
	columns := append(selectKeys, selectMetrics...)
	return columns, resultMetrics, err
}

// GetSQLQueryAndParametersForAdwordsQueryV1 - @Kark TODO v1
// TODO query breakage with "!%(MISSING)" on gorm.
// TODO Understand null cases.
func (pg *Postgres) GetSQLQueryAndParametersForAdwordsQueryV1(projectID uint64, query *model.ChannelQueryV1, reqID string, fetchSource bool) (string, []interface{}, []string, []string, error) {
	var selectMetrics []string
	var selectKeys []string
	var sql string
	var params []interface{}
	transformedQuery, customerAccountID, err := pg.transFormRequestFieldsAndFetchRequiredFieldsForAdwords(projectID, *query, reqID)
	if err != nil {
		return "", make([]interface{}, 0, 0), make([]string, 0, 0), make([]string, 0, 0), err
	}
	if hasAllIDsOnlyInGroupBy(transformedQuery) {
		sql, params, selectKeys, selectMetrics, err = pg.BuildAdwordsSimpleQueryV1(transformedQuery, projectID, *customerAccountID, reqID, fetchSource)
		if err != nil {
			return "", make([]interface{}, 0, 0), make([]string, 0, 0), make([]string, 0, 0), err
		}
	} else {
		sql, params, selectKeys, selectMetrics = buildAdwordsComplexQueryV1(transformedQuery, projectID, *customerAccountID, fetchSource)
	}
	return sql, params, selectKeys, selectMetrics, nil
}

// @Kark TODO v1
func (pg *Postgres) transFormRequestFieldsAndFetchRequiredFieldsForAdwords(projectID uint64, query model.ChannelQueryV1, reqID string) (*model.ChannelQueryV1, *string, error) {
	var transformedQuery model.ChannelQueryV1
	logCtx := log.WithField("req_id", reqID)
	var err error
	projectSetting, errCode := pg.GetProjectSetting(projectID)
	if errCode != http.StatusFound {
		return &model.ChannelQueryV1{}, nil, errors.New("Project setting not found")
	}
	customerAccountID := projectSetting.IntAdwordsCustomerAccountId

	transformedQuery, err = convertFromRequestToAdwordsSpecificRepresentation(query)
	if err != nil {
		logCtx.Warn("Request failed in validation: ", err)
		return &model.ChannelQueryV1{}, nil, err
	}
	return &transformedQuery, customerAccountID, nil
}

// @Kark TODO v1
// Currently, this relies on assumption of Object across different filterObjects. Change when we need robust.
func convertFromRequestToAdwordsSpecificRepresentation(query model.ChannelQueryV1) (model.ChannelQueryV1, error) {
	var transformedQuery model.ChannelQueryV1
	var err1, err2, err3 error
	transformedQuery.SelectMetrics, err1 = getAdwordsSpecificMetrics(query.SelectMetrics)
	transformedQuery.Filters, err2 = getAdwordsSpecificFilters(query.Filters)
	transformedQuery.GroupBy, err3 = getAdwordsSpecificGroupBy(query.GroupBy)
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
func getAdwordsSpecificMetrics(requestSelectMetrics []string) ([]string, error) {
	resultMetrics := make([]string, 0, 0)
	for _, requestMetric := range requestSelectMetrics {
		metric, isPresent := adwordsExternalRepresentationToInternalRepresentation[requestMetric]
		if !isPresent {
			return make([]string, 0, 0), errors.New("Invalid metric key found for document type")
		}
		resultMetrics = append(resultMetrics, metric)
	}
	return resultMetrics, nil
}

// @Kark TODO v1
func getAdwordsSpecificFilters(requestFilters []model.FilterV1) ([]model.FilterV1, error) {
	resultFilters := make([]model.FilterV1, 0, 0)
	for _, requestFilter := range requestFilters {
		var resultFilter model.FilterV1
		filterObject, isPresent := adwordsExternalRepresentationToInternalRepresentation[requestFilter.Object]
		if !isPresent {
			return make([]model.FilterV1, 0, 0), errors.New("Invalid filter key found for document type")
		}
		resultFilter = requestFilter
		resultFilter.Object = filterObject
		resultFilters = append(resultFilters, resultFilter)
	}
	return resultFilters, nil
}

// @Kark TODO v1
func getAdwordsSpecificGroupBy(requestGroupBys []model.GroupBy) ([]model.GroupBy, error) {
	sortedGroupBys := make([]model.GroupBy, 0, 0)
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

	for _, groupBy := range requestGroupBys {
		if groupBy.Object == CAFilterKeyword {
			sortedGroupBys = append(sortedGroupBys, groupBy)
		}
	}

	resultGroupBys := make([]model.GroupBy, 0, 0)
	for _, requestGroupBy := range sortedGroupBys {
		var resultGroupBy model.GroupBy
		groupByObject, isPresent := adwordsExternalRepresentationToInternalRepresentation[requestGroupBy.Object]
		if !isPresent {
			return make([]model.GroupBy, 0, 0), errors.New("Invalid groupby key found for document type")
		}
		resultGroupBy = requestGroupBy
		resultGroupBy.Object = groupByObject
		resultGroupBys = append(resultGroupBys, resultGroupBy)
	}
	return resultGroupBys, nil
}

/*
SELECT campaign_id, date_trunc('day', to_timestamp(timestamp::text, 'YYYYMMDD') AT TIME ZONE 'UTC') as datetime,
SUM((value->>'impressions')::float) as impressions, SUM((value->>'clicks')::float) as clicks  FROM adwords_documents
 WHERE project_id = '2' AND customer_account_id IN ( '2368493227' ) AND type = '5' AND timestamp between '20200331' AND '20200401'
  GROUP BY campaign_id, datetime ORDER BY impressions DESC, clicks DESC LIMIT 2500 ;
*/
// @Kark TODO v1
func (pg *Postgres) BuildAdwordsSimpleQueryV1(query *model.ChannelQueryV1, projectID uint64, customerAccountID string, reqID string, fetchSource bool) (string, []interface{}, []string, []string, error) {
	campaignIDs, adGroupIDs, err := pg.GetIDsFromAdwordsSimpleJob(query, projectID, customerAccountID, reqID)
	if err != nil {
		return "", make([]interface{}, 0), make([]string, 0), make([]string, 0), err
	}
	lowestHierarchyLevel := getLowestHierarchyLevelForAdwords(query)
	lowestHierarchyReportLevel := lowestHierarchyLevel + "_performance_report"
	sql, params, selectKeys, selectMetrics := getSQLAndParamsFromAdwordsSimpleReports(query, projectID, query.From, query.To, customerAccountID, model.AdwordsDocumentTypeAlias[lowestHierarchyReportLevel],
		campaignIDs, adGroupIDs, reqID, fetchSource)
	return sql, params, selectKeys, selectMetrics, nil
}

// Validation issue needed. Not both ad_id , keyword_id at same time.
// @Kark TODO v1
func (pg *Postgres) GetIDsFromAdwordsSimpleJob(query *model.ChannelQueryV1, projectID uint64, adwordsAccountIDs string, reqID string) ([]int, []int, error) {
	var err error
	campaignsFilters, adGroupFilters, _ := splitFiltersByObjectTypeForAdwords(query)
	campaignIDs, err := pg.GetIDsByPropertyOnAdwordsSimpleJob(projectID, query.From, query.To, adwordsAccountIDs, model.AdwordsCampaign, campaignsFilters, reqID)
	if err != nil {
		return make([]int, 0), make([]int, 0), err
	}
	adGroupIDs, err := pg.GetIDsByPropertyOnAdwordsSimpleJob(projectID, query.From, query.To, adwordsAccountIDs, model.AdwordsAdGroup, adGroupFilters, reqID)
	if err != nil {
		return make([]int, 0), make([]int, 0), err
	}
	// adIDs := getAdIDsByPropertyOnJob(adFilters)
	return campaignIDs, adGroupIDs, nil
}

// @Kark TODO v1
func splitFiltersByObjectTypeForAdwords(query *model.ChannelQueryV1) ([]model.FilterV1, []model.FilterV1, []model.FilterV1) {
	campaignsFilters := make([]model.FilterV1, 0, 0)
	adGroupFilters := make([]model.FilterV1, 0, 0)
	adFilters := make([]model.FilterV1, 0, 0)

	for _, filter := range query.Filters {
		switch filter.Object {
		case model.AdwordsCampaign:
			campaignsFilters = append(campaignsFilters, filter)
		case model.AdwordsAdGroup:
			adGroupFilters = append(adGroupFilters, filter)
		case model.AdwordsAd:
			adFilters = append(adFilters, filter)
		}
	}
	return campaignsFilters, adGroupFilters, adFilters
}

// @TODO Kark v1
func (pg *Postgres) GetIDsByPropertyOnAdwordsSimpleJob(projectID uint64, from, to int64, adwordsAccountIDs string, typeOfJob string, filters []model.FilterV1, reqID string) ([]int, error) {
	logCtx := log.WithField("req_id", reqID)
	if len(filters) == 0 {
		return []int{}, nil
	}
	docType := model.AdwordsDocumentTypeAlias[typeOfJob+"s"]
	sqlParams := make([]interface{}, 0)
	customerAccountIDs := strings.Split(adwordsAccountIDs, ",")
	IDColumn := objectAndPropertyToValueInsideAdwordsJobsMapping[typeOfJob][typeOfJob+":id"]
	selectStatement := fmt.Sprintf("SELECT %s FROM adwords_documents", IDColumn)
	groupByStatement := fmt.Sprintf("GROUP BY %s", IDColumn)
	isNotNULLStatement := fmt.Sprintf("%s IS NOT NULL", IDColumn)

	sqlParams = append(sqlParams, projectID, customerAccountIDs, docType, from, to)
	filterPropertiesStatement, filterPropertiesParams := getFilterPropertiesForAdwordsJob(filters)
	completeFiltersStatement := fmt.Sprintf("%s AND %s ", staticWhereStatementForAdwords, isNotNULLStatement)
	if len(filterPropertiesStatement) != 0 {
		completeFiltersStatement += "AND " + filterPropertiesStatement + " "
		sqlParams = append(sqlParams, filterPropertiesParams...)
	}

	resultSQLStatement := fmt.Sprintf("%s %s %s;", selectStatement, completeFiltersStatement, groupByStatement)

	rows, err := pg.ExecQueryWithContext(resultSQLStatement, sqlParams)
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
func getLowestHierarchyLevelForAdwords(query *model.ChannelQueryV1) string {
	// Fetch the propertyNames
	return getLowestHierarchyLevelForAdwordsFiltersAndGroupBy(query.Filters, query.GroupBy)
}

// @TODO Kark v1
func getLowestHierarchyLevelForAdwordsFiltersAndGroupBy(filters []model.FilterV1, groupBys []model.GroupBy) string {
	var objectNames []string
	for _, filter := range filters {
		objectNames = append(objectNames, filter.Object)
	}

	for _, groupBy := range groupBys {
		objectNames = append(objectNames, groupBy.Object)
	}

	// Check if present
	for _, objectName := range objectNames {
		if objectName == model.AdwordsAd {
			return model.AdwordsAd
		}
	}

	for _, objectName := range objectNames {
		if objectName == model.AdwordsKeyword {
			return model.AdwordsKeyword
		}
	}

	for _, objectName := range objectNames {
		if objectName == model.AdwordsAdGroup {
			return model.AdwordsAdGroup
		}
	}

	for _, objectName := range objectNames {
		if objectName == model.AdwordsCampaign {
			return model.AdwordsCampaign
		}
	}
	return model.AdwordsCampaign
}

/*
SELECT campaign_id, date_trunc('day', to_timestamp(timestamp::text, 'YYYYMMDD') AT TIME ZONE 'UTC') as datetime,
SUM((value->>'impressions')::float) as impressions FROM adwords_documents WHERE project_id = '8' AND customer_account_id IN ( '3543296298' )
AND type = '5' AND timestamp between '20200331' AND '20200401' GROUP BY campaign_id, datetime ORDER BY impressions DESC LIMIT 2500 ;
*/
// @TODO Kark v1
func getSQLAndParamsFromAdwordsSimpleReports(query *model.ChannelQueryV1, projectID uint64, from, to int64, adwordsAccountIDs string,
	docType int, campaignIDs []int, adGroupIDs []int, reqID string, fetchSource bool) (string, []interface{}, []string, []string) {
	customerAccountIDs := strings.Split(adwordsAccountIDs, ",")
	selectQuery := "SELECT "
	selectMetrics := make([]string, 0, 0)
	isGroupByTimestamp := query.GetGroupByTimestamp() != ""
	groupByStatement := ""
	groupByKeysWithoutTimestamp := make([]string, 0, 0)
	selectKeys := make([]string, 0, 0)
	responseSelectKeys := make([]string, 0, 0)
	responseSelectMetrics := make([]string, 0, 0)
	staticWhereParams := []interface{}{projectID, customerAccountIDs, docType, from, to}
	finalWhereStatement := ""
	finalParams := make([]interface{}, 0, 0)
	// QueryBy
	for _, groupBy := range query.GroupBy {
		key := groupBy.Object + ":" + groupBy.Property
		groupByKeysWithoutTimestamp = append(groupByKeysWithoutTimestamp, adwordsInternalRepresentationToExternalRepresentation[key])
	}
	if isGroupByTimestamp {
		groupByStatement = joinWithComma(append(groupByKeysWithoutTimestamp, model.AliasDateTime)...)
	} else {
		groupByStatement = joinWithComma(groupByKeysWithoutTimestamp...)
	}

	// SelectKeys
	if fetchSource {
		selectKeys = append(selectKeys, fmt.Sprintf("'%s' as %s", model.AdwordsStringColumn, source))
		responseSelectKeys = append(responseSelectKeys, source)
	}

	for _, groupBy := range query.GroupBy {
		key := groupBy.Object + ":" + groupBy.Property
		value := fmt.Sprintf("%s as %s", objectAndPropertyToValueInsideAdwordsReportsMapping[key], adwordsInternalRepresentationToExternalRepresentation[key])
		selectKeys = append(selectKeys, value)
		responseSelectKeys = append(responseSelectKeys, adwordsInternalRepresentationToExternalRepresentation[key])
	}

	if isGroupByTimestamp {
		selectKeys = append(selectKeys, fmt.Sprintf("%s as %s",
			getSelectTimestampByTypeForChannels(query.GetGroupByTimestamp(), query.Timezone), model.AliasDateTime))
		responseSelectKeys = append(responseSelectKeys, model.AliasDateTime)
	}
	for _, selectMetric := range query.SelectMetrics {
		value := fmt.Sprintf("%s as %s", adwordsMetricsToAggregatesInReportsMapping[selectMetric], adwordsInternalRepresentationToExternalRepresentation[selectMetric])
		selectMetrics = append(selectMetrics, value)

		value = adwordsInternalRepresentationToExternalRepresentation[selectMetric]
		responseSelectMetrics = append(responseSelectMetrics, value)
	}
	selectQuery += joinWithComma(append(selectKeys, selectMetrics...)...)

	// OrderBy
	orderByQuery := "ORDER BY " + getOrderByClause(responseSelectMetrics)

	// Where
	filterIDKeys := make([]string, 0, 0)
	if len(campaignIDs) != 0 {
		campaignIDsString := ""
		for _, campaignID := range campaignIDs {
			campaignIDsString += strconv.Itoa(campaignID) + ","
		}
		campaignIDsString = campaignIDsString[:len(campaignIDsString)-1]
		campaignIdsFilter := fmt.Sprintf("campaign_id IN (%s) ", campaignIDsString)
		filterIDKeys = append(filterIDKeys, campaignIdsFilter)
	}
	if len(adGroupIDs) != 0 {
		adGroupIDsString := ""
		for _, adGroupID := range adGroupIDs {
			adGroupIDsString += strconv.Itoa(adGroupID) + ","
		}
		adGroupIDsString = adGroupIDsString[:len(adGroupIDsString)-1]
		adGroupIdsFilter := fmt.Sprintf("ad_group_id IN (%s) ", adGroupIDsString)
		filterIDKeys = append(filterIDKeys, adGroupIdsFilter)
	}
	keywordFilters := make([]model.FilterV1, 0, 0)
	for _, filter := range query.Filters {
		if strings.Contains(filter.Object, model.AdwordsKeyword) {
			keywordFilters = append(keywordFilters, filter)
		}
	}
	keywordFilterStatement, keywordFilterParams := getFilterPropertiesForAdwordsReports(keywordFilters)
	if len(keywordFilterStatement) != 0 {
		filterIDKeys = append(filterIDKeys, keywordFilterStatement)
	}
	if len(filterIDKeys) != 0 {
		filterIDsStatement := ""
		for _, filterIDKey := range filterIDKeys {
			filterIDsStatement = fmt.Sprintf("AND %s", filterIDKey)
		}
		finalWhereStatement = fmt.Sprintf(" %s %s ", staticWhereStatementForAdwords, filterIDsStatement)
	} else {
		finalWhereStatement = staticWhereStatementForAdwords
	}
	finalParams = append(finalParams, staticWhereParams...)
	finalParams = append(finalParams, keywordFilterParams...)

	// Final Query
	resultSQLStatement := fmt.Sprintf("%s %s %s ", selectQuery, fromAdwordsDocument, finalWhereStatement)
	if len(groupByStatement) != 0 {
		resultSQLStatement += "GROUP BY " + groupByStatement
	}
	resultSQLStatement += " " + orderByQuery + channeAnalyticsLimit + ";"
	return resultSQLStatement, finalParams, responseSelectKeys, responseSelectMetrics
}

// @Kark TODO v1
func buildAdwordsComplexQueryV1(query *model.ChannelQueryV1, projectID uint64, customerAccountID string, fetchSource bool) (string, []interface{}, []string, []string) {
	idBasedFilters, nonIdBasedFilters := splitFiltersBasedOnIdProperty(query.Filters)
	keywordGroupBy, nonKeywordGroupBys := splitGroupByBasedOnKeyword(query.GroupBy)
	if containsKeywords(query) {
		return buildAdwordsComplexWithKeywords(query, projectID, customerAccountID, fetchSource, idBasedFilters, nonIdBasedFilters, keywordGroupBy, nonKeywordGroupBys)
	} else {
		return buildAdwordsComplexWithoutKeywords(query, projectID, customerAccountID, fetchSource, idBasedFilters, nonIdBasedFilters)
	}
}

// @Kark TODO v1
func containsKeywords(query *model.ChannelQueryV1) bool {
	for _, filter := range query.Filters {
		if filter.Object == model.AdwordsKeyword {
			return true
		}
	}
	for _, groupBy := range query.GroupBy {
		if groupBy.Object == model.AdwordsKeyword {
			return true
		}
	}
	return false
}

/*
WITH reports_cte as (SELECT keyword_id, campaign_id, date_trunc('day', to_timestamp(timestamp::text, 'YYYYMMDD') AT TIME ZONE 'UTC') as datetime,
SUM((value->>'impressions')::float) as impressions, SUM((value->>'clicks')::float) as clicks FROM adwords_documents WHERE project_id = '2'
AND customer_account_id IN ( '2368493227' ) AND type = '8' AND timestamp between '20200331' AND '20200401'  AND 'keyword_id' = '1'
GROUP BY keyword_id, campaign_id, datetime ),

jobs_cte as (SELECT campaign_id as campaign_id, value->>'name' as campaign_name,
date_trunc('day', to_timestamp(timestamp::text, 'YYYYMMDD') AT TIME ZONE 'UTC') as datetime FROM adwords_documents WHERE
project_id = '2' AND customer_account_id IN ( '2368493227' ) AND type = '1' AND timestamp between '20200331' AND '20200401'
AND value->>'name' ILIKE '%Brand - BLR - New_Aug_Desktop_RLSA%' GROUP BY campaign_id, campaign_name, datetime)

SELECT jobs_cte.campaign_name, reports_cte.keyword_id, reports_cte.datetime, sum(reports_cte.impressions) as impressions,
sum(reports_cte.clicks) as clicks from reports_cte INNER JOIN jobs_cte ON reports_cte.campaign_id=jobs_cte.campaign_id AND
reports_cte.datetime=jobs_cte.datetime GROUP BY jobs_cte.campaign_name, reports_cte.keyword_id, reports_cte.datetime ORDER BY
impressions DESC, clicks DESC LIMIT 2500 ;
*/
// @Kark TODO v1
func buildAdwordsComplexWithKeywords(query *model.ChannelQueryV1, projectID uint64, customerAccountID string, fetchSource bool,
	idBasedFilters []model.FilterV1, nonIdBasedFilters []model.FilterV1, keywordBasedGroupBys []model.GroupBy, nonKeywordBasedGroupBys []model.GroupBy) (string, []interface{}, []string, []string) {
	isGroupByTimestamp := query.GetGroupByTimestamp() != ""
	groupByKeysWithoutTimestamp := make([]string, 0, 0)
	selectKeys := make([]string, 0, 0)
	selectMetrics := make([]string, 0, 0)
	responseSelectKeys := make([]string, 0, 0)
	responseSelectMetrics := make([]string, 0, 0)

	reportCTE, reportCTEAlias, reportSelectMetrics, reportCTEJoinFields, reportParams := getCTEAndParamsForKeywordsReportComplexStrategy(query, projectID, customerAccountID, idBasedFilters, keywordBasedGroupBys, nonKeywordBasedGroupBys)
	jobCTE, jobsCTEAlias, jobCTEJoinFields, jobsParams := getCTEAndParamsForKeywordsJobsComplexStrategy(query, projectID, customerAccountID, nonIdBasedFilters, keywordBasedGroupBys, nonKeywordBasedGroupBys)

	finalWithClause := joinWithComma(reportCTE, jobCTE)
	finalParams := make([]interface{}, 0, 0)
	finalGroupByKeys := make([]string, 0, 0)
	finalSelectStatement := ""
	finalSelectKeys := make([]string, 0, 0)
	finalInnerJoin := ""

	finalParams = append(finalParams, reportParams...)
	finalParams = append(finalParams, jobsParams...)

	// responseSelectKeys
	// model.GroupBy
	for _, groupBy := range query.GroupBy {
		key := groupBy.Object + ":" + groupBy.Property
		value := ""
		if groupBy.Object == model.AdwordsKeyword {
			value = reportCTEAlias + "." + propertyToExposedValueFromCTE[key]
		} else {
			value = jobsCTEAlias + "." + propertyToExposedValueFromCTE[key]
		}
		groupByKeysWithoutTimestamp = append(groupByKeysWithoutTimestamp, value)
	}
	if isGroupByTimestamp {
		finalGroupByKeys = append(groupByKeysWithoutTimestamp, reportCTEAlias+"."+model.AliasDateTime)
	}

	// SelectKeys
	if fetchSource {
		selectKeys = append(selectKeys, fmt.Sprintf("'%s' as %s", model.AdwordsStringColumn, source))
		responseSelectKeys = append(responseSelectKeys, source)
	}

	for _, groupBy := range query.GroupBy {
		key := groupBy.Object + ":" + groupBy.Property
		value := ""
		if groupBy.Object == model.AdwordsKeyword {
			value = fmt.Sprintf("%s.%s as %s", reportCTEAlias, propertyToExposedValueFromCTE[key], adwordsInternalRepresentationToExternalRepresentation[key])
		} else {
			value = fmt.Sprintf("%s.%s as %s", jobsCTEAlias, propertyToExposedValueFromCTE[key], adwordsInternalRepresentationToExternalRepresentation[key])
		}
		selectKeys = append(selectKeys, value)
		responseSelectKeys = append(responseSelectKeys, adwordsInternalRepresentationToExternalRepresentation[key])
	}
	if isGroupByTimestamp {
		value := fmt.Sprintf("%s.%s as %s", reportCTEAlias, model.AliasDateTime, model.AliasDateTime)
		selectKeys = append(selectKeys, value)
		responseSelectKeys = append(responseSelectKeys, model.AliasDateTime)
	}
	for _, selectMetric := range reportSelectMetrics {
		value := fmt.Sprintf("%s(%s.%s) as %s", adwordsMetricsToOperation[selectMetric], reportCTEAlias, selectMetric, adwordsInternalRepresentationToExternalRepresentation[selectMetric])
		selectMetrics = append(selectMetrics, value)

		value = adwordsInternalRepresentationToExternalRepresentation[selectMetric]
		responseSelectMetrics = append(responseSelectMetrics, value)
	}
	finalSelectKeys = append(selectKeys, selectMetrics...)
	finalSelectStatement = "SELECT " + joinWithComma(finalSelectKeys...)

	// Inner join
	finalInnerJoin = fmt.Sprintf(" from %s INNER JOIN %s ON ", reportCTEAlias, jobsCTEAlias)
	for index, jobCTEJoinField := range jobCTEJoinFields {
		finalInnerJoin += fmt.Sprintf("%s.%s=%s.%s AND ", reportCTEAlias, reportCTEJoinFields[index], jobsCTEAlias, jobCTEJoinField)
	}
	finalInnerJoin = finalInnerJoin[:len(finalInnerJoin)-4]

	// orderBy
	orderByQuery := "ORDER BY " + getOrderByClause(responseSelectMetrics)

	// forming final query
	resultSQLStatement := finalWithClause + finalSelectStatement + finalInnerJoin
	if len(finalGroupByKeys) != 0 {
		resultSQLStatement += "GROUP BY " + joinWithComma(finalGroupByKeys...)
	}
	resultSQLStatement += " " + orderByQuery + channeAnalyticsLimit + ";"
	return resultSQLStatement, finalParams, responseSelectKeys, responseSelectMetrics
}

// @Kark TODO v1
func getCTEAndParamsForKeywordsReportComplexStrategy(query *model.ChannelQueryV1, projectID uint64,
	customerAccountID string, idBasedFilters []model.FilterV1, keywordBasedGroupBys []model.GroupBy, nonKeywordBasedGroupBys []model.GroupBy) (string, string, []string, []string, []interface{}) {
	cteAlias := "reports_cte"
	lowestHierarchyLevel := getLowestHierarchyLevelForAdwords(query)
	lowestHierarchyReportLevel := lowestHierarchyLevel + "_performance_report"
	docType := model.AdwordsDocumentTypeAlias[lowestHierarchyReportLevel]
	customerAccountIDs := strings.Split(customerAccountID, ",")

	staticWhereParams := []interface{}{projectID, customerAccountIDs, docType, query.From, query.To}
	selectStatement := fmt.Sprintf("WITH %s as (SELECT ", cteAlias)
	selectKeys := make([]string, 0, 0)
	groupByStatement := ""
	var groupByKeys []string

	// Where
	finalWhereStatemnt := ""
	finalParams := make([]interface{}, 0, 0)
	idBasedFilterStatement, idBasedFilterParams := getFilterPropertiesForAdwordsReports(idBasedFilters)
	if len(idBasedFilterStatement) != 0 {
		finalWhereStatemnt += fmt.Sprintf("%s AND %s", staticWhereStatementForAdwords, idBasedFilterStatement)
	} else {
		finalWhereStatemnt = staticWhereStatementForAdwords
	}

	finalParams = append(finalParams, staticWhereParams...)
	finalParams = append(finalParams, idBasedFilterParams...)
	uniqueIDColumns := getHierarchyIdsFromGroupBysForReports(nonKeywordBasedGroupBys)

	// groupBy
	if len(keywordBasedGroupBys) != 0 {
		keywordValue := objectAndPropertyToValueInsideAdwordsReportsMapping[model.AdwordsKeyword+":id"]
		groupByKeys = append(groupByKeys, keywordValue)
	}
	groupByKeys = append(groupByKeys, uniqueIDColumns...)
	groupByKeys = append(groupByKeys, model.AliasDateTime)
	groupByStatement = joinWithComma(groupByKeys...)
	joinFields := append(uniqueIDColumns, model.AliasDateTime)

	// selectKeys
	if len(keywordBasedGroupBys) != 0 {
		selectKeys = append(selectKeys, objectAndPropertyToValueInsideAdwordsReportsMapping[model.AdwordsKeyword+":id"])
	}

	selectKeys = append(selectKeys, uniqueIDColumns...)
	selectStatement += joinWithComma(selectKeys...)
	currentSelectQuery := fmt.Sprintf("%s as %s",
		getSelectTimestampByTypeForChannels(query.GetGroupByTimestamp(), query.Timezone), model.AliasDateTime)
	selectStatement = joinWithComma(selectStatement, currentSelectQuery)
	for _, selectMetric := range query.SelectMetrics {
		currentSelectQuery := fmt.Sprintf("%s as %s", adwordsMetricsToAggregatesInReportsMapping[selectMetric], selectMetric)
		selectStatement = joinWithComma(selectStatement, currentSelectQuery)
	}

	resultSQLStatement := selectStatement + fromAdwordsDocument + finalWhereStatemnt + " GROUP BY " + groupByStatement + " )"
	return resultSQLStatement, cteAlias, query.SelectMetrics, joinFields, finalParams
}

// @Kark TODO v1
func getCTEAndParamsForKeywordsJobsComplexStrategy(query *model.ChannelQueryV1, projectID uint64,
	customerAccountID string, nonIDBasedFilters []model.FilterV1, keywordBasedGroupBys []model.GroupBy, nonKeywordBasedGroupBys []model.GroupBy) (string, string, []string, []interface{}) {
	cteAlias := "jobs_cte"
	lowestHierarchyLevel := getLowestHierarchyLevelForAdwordsFiltersAndGroupBy(nonIDBasedFilters, nonKeywordBasedGroupBys)
	lowestHierarchyJobLevel := lowestHierarchyLevel + "s"
	docType := model.AdwordsDocumentTypeAlias[lowestHierarchyJobLevel]
	customerAccountIDs := strings.Split(customerAccountID, ",")

	staticWhereParams := []interface{}{projectID, customerAccountIDs, docType, query.From, query.To}
	selectStatement := fmt.Sprintf("%s as (SELECT ", cteAlias)
	var groupByKeys []string

	finalWhereStatemnt := ""
	finalParams := make([]interface{}, 0, 0)
	finalJoinFields := make([]string, 0, 0)
	finalGroupByStatement := ""
	finalSelectStatement := ""
	resultStatement := ""

	nonIDBasedFilterStatement, nonIDBasedFilterParams := getFilterPropertiesForAdwordsJob(nonIDBasedFilters)
	if len(nonIDBasedFilterStatement) != 0 {
		finalWhereStatemnt += fmt.Sprintf("%s AND %s", staticWhereStatementForAdwords, nonIDBasedFilterStatement)
	} else {
		finalWhereStatemnt = staticWhereStatementForAdwords
	}
	finalParams = append(finalParams, staticWhereParams...)
	finalParams = append(finalParams, nonIDBasedFilterParams...)
	uniqueIDsForCTE, selectKeysWithoutDateTime, groupByKeysWithoutDateTime := getUniqueIDsForCTEAndSelectKeysAndGroupByFields(nonKeywordBasedGroupBys, lowestHierarchyLevel)
	finalJoinFields = append(uniqueIDsForCTE, model.AliasDateTime)

	groupByKeys = append(groupByKeysWithoutDateTime, model.AliasDateTime)
	if len(groupByKeys) != 0 {
		finalGroupByStatement = " GROUP BY " + joinWithComma(groupByKeys...)
	}

	finalSelectStatement = selectStatement + joinWithComma(selectKeysWithoutDateTime...)
	currentSelectQuery := fmt.Sprintf("%s as %s",
		getSelectTimestampByTypeForChannels(query.GetGroupByTimestamp(), query.Timezone), model.AliasDateTime)
	finalSelectStatement = joinWithComma(finalSelectStatement, currentSelectQuery)

	// TODO: Add filters
	resultStatement = finalSelectStatement + fromAdwordsDocument + finalWhereStatemnt + finalGroupByStatement + ")"
	return resultStatement, cteAlias, finalJoinFields, finalParams // finalGroupByStatement
}

/*
SELECT value->>'campaign_name' as campaign_name, date_trunc('day', to_timestamp(timestamp::text, 'YYYYMMDD') AT TIME ZONE 'UTC') as datetime,
SUM((value->>'impressions')::float) as impressions, SUM((value->>'clicks')::float) as clicks FROM adwords_documents WHERE project_id = '2' AND
customer_account_id IN ( '2368493227' ) AND type = '5' AND timestamp between '20200331' AND '20200401'
AND value->>'campaign_name' ILIKE '%Brand - BLR - New_Aug_Desktop_RLSA%' GROUP BY campaign_name, datetime
ORDER BY impressions DESC, clicks DESC LIMIT 2500 ;
*/
// @Kark TODO v1
func buildAdwordsComplexWithoutKeywords(query *model.ChannelQueryV1, projectID uint64, customerAccountID string, fetchSource bool, idBasedFilters []model.FilterV1, nonIDBasedFilters []model.FilterV1) (string, []interface{}, []string, []string) {
	lowestHierarchyLevel := getLowestHierarchyLevelForAdwords(query)
	lowestHierarchyReportLevel := lowestHierarchyLevel + "_performance_report"
	sql, params, selectKeys, selectMetrics := getSQLAndParamsForAdwordsComplexWithoutKeywords(query, projectID, query.From, query.To, customerAccountID, model.AdwordsDocumentTypeAlias[lowestHierarchyReportLevel],
		fetchSource)
	return sql, params, selectKeys, selectMetrics
}

// @Kark TODO v1
func getSQLAndParamsForAdwordsComplexWithoutKeywords(query *model.ChannelQueryV1, projectID uint64, from, to int64, customerAccountID string,
	docType int, fetchSource bool) (string, []interface{}, []string, []string) {
	customerAccountIDs := strings.Split(customerAccountID, ",")
	selectQuery := "SELECT "
	isGroupByTimestamp := query.GetGroupByTimestamp() != ""
	groupByKeysWithoutTimestamp := make([]string, 0, 0)
	selectKeys := make([]string, 0, 0)
	selectMetrics := make([]string, 0, 0)
	responseSelectKeys := make([]string, 0, 0)
	responseSelectMetrics := make([]string, 0, 0)
	staticWhereParams := []interface{}{projectID, customerAccountIDs, docType, from, to}

	finalParams := make([]interface{}, 0, 0)
	finalGroupByKeys := make([]string, 0, 0)
	finalSelectStatement := ""
	finalWhereStatement := ""
	finalSelectKeys := make([]string, 0, 0)

	// model.GroupBy
	for _, groupBy := range query.GroupBy {
		key := groupBy.Object + ":" + groupBy.Property
		groupByKeysWithoutTimestamp = append(groupByKeysWithoutTimestamp, propertyToExposedValueFromCTE[key])
	}
	if isGroupByTimestamp {
		finalGroupByKeys = append(groupByKeysWithoutTimestamp, model.AliasDateTime)
	} else {
		finalGroupByKeys = groupByKeysWithoutTimestamp
	}

	// SelectKeys
	if fetchSource {
		selectKeys = append(selectKeys, fmt.Sprintf("'%s' as %s", model.AdwordsStringColumn, source))
		responseSelectKeys = append(responseSelectKeys, source)
	}
	for _, groupBy := range query.GroupBy {
		key := groupBy.Object + ":" + groupBy.Property
		value := ""
		if groupBy.Property == "id" {
			value = fmt.Sprintf("%s as %s", objectAndPropertyToValueInsideAdwordsReportsMapping[key], propertyToExposedValueFromCTE[key])
		} else {
			value = fmt.Sprintf("value->>'%s' as %s", objectAndPropertyToValueInsideAdwordsReportsMapping[key], propertyToExposedValueFromCTE[key])
		}
		selectKeys = append(selectKeys, value)
		responseSelectKeys = append(responseSelectKeys, propertyToExposedValueFromCTE[key])
	}
	if isGroupByTimestamp {
		selectKeys = append(selectKeys, fmt.Sprintf("%s as %s",
			getSelectTimestampByTypeForChannels(query.GetGroupByTimestamp(), query.Timezone), model.AliasDateTime))
		responseSelectKeys = append(responseSelectKeys, model.AliasDateTime)
	}
	for _, selectMetric := range query.SelectMetrics {
		value := fmt.Sprintf("%s as %s", adwordsMetricsToAggregatesInReportsMapping[selectMetric], adwordsInternalRepresentationToExternalRepresentation[selectMetric])
		selectMetrics = append(selectMetrics, value)

		value = adwordsInternalRepresentationToExternalRepresentation[selectMetric]
		responseSelectMetrics = append(responseSelectMetrics, value)
	}
	finalSelectKeys = append(finalSelectKeys, selectKeys...)
	finalSelectKeys = append(finalSelectKeys, selectMetrics...)
	finalSelectStatement = selectQuery + joinWithComma(finalSelectKeys...)

	// Order by and where
	orderByQuery := "ORDER BY " + getOrderByClause(responseSelectMetrics)
	filterStatement, filterParams := getFilterPropertiesForAdwordsReports(query.Filters)
	finalParams = append(finalParams, staticWhereParams...)
	finalParams = append(finalParams, filterParams...)
	if len(filterStatement) != 0 {
		finalWhereStatement += fmt.Sprintf("%s AND %s", staticWhereStatementForAdwords, filterStatement)
	} else {
		finalWhereStatement += staticWhereStatementForAdwords
	}

	// final query
	resultSQLStatement := finalSelectStatement + fromAdwordsDocument + finalWhereStatement
	if len(finalGroupByKeys) != 0 {
		resultSQLStatement += " GROUP BY " + joinWithComma(finalGroupByKeys...)
	}
	resultSQLStatement += " " + orderByQuery + channeAnalyticsLimit + ";"
	return resultSQLStatement, finalParams, responseSelectKeys, responseSelectMetrics
}

// @Kark TODO v1
func getUniqueIDsForCTEAndSelectKeysAndGroupByFields(groupBys []model.GroupBy, lowestHierarchyLevel string) ([]string, []string, []string) {
	uniqueIDsForCTE := make([]string, 0, 0)
	selectKeys := make([]string, 0, 0)
	selectValue := ""
	groupByKeys := make([]string, 0, 0)

	uniqueObjects := make(map[string]struct{})
	for _, groupBy := range groupBys {
		// get the unique values for joinCTE
		isObjectPresentPreviously := false
		if _, isObjectPresentPreviously = uniqueObjects[groupBy.Object]; !isObjectPresentPreviously {
			key := groupBy.Object + ":id"
			uniqueIDColumn := objectAndPropertyToValueInsideAdwordsJobsMapping[lowestHierarchyLevel][key]
			uniqueIDsForCTE = append(uniqueIDsForCTE, uniqueIDColumn)
			uniqueObjects[groupBy.Object] = struct{}{}
		}

		if !isObjectPresentPreviously && groupBy.Property != "id" {
			key := groupBy.Object + ":id"
			groupByValue := propertyToExposedValueFromCTE[key]
			groupByKeys = append(groupByKeys, groupByValue)
			selectValue = fmt.Sprintf("%s as %s", objectAndPropertyToValueInsideAdwordsJobsMapping[lowestHierarchyLevel][key], propertyToExposedValueFromCTE[key])
			selectKeys = append(selectKeys, selectValue)
		}
		key := groupBy.Object + ":" + groupBy.Property
		groupByValue := propertyToExposedValueFromCTE[key]
		groupByKeys = append(groupByKeys, groupByValue)
		if groupBy.Property == "id" {
			selectValue = fmt.Sprintf("%s as %s", objectAndPropertyToValueInsideAdwordsJobsMapping[lowestHierarchyLevel][key], propertyToExposedValueFromCTE[key])
		} else {
			selectValue = fmt.Sprintf("value->>'%s' as %s", objectAndPropertyToValueInsideAdwordsJobsMapping[lowestHierarchyLevel][key], propertyToExposedValueFromCTE[key])
		}

		selectKeys = append(selectKeys, selectValue)
	}
	return uniqueIDsForCTE, selectKeys, groupByKeys
}

// @Kark TODO v1
func splitFiltersBasedOnIdProperty(filters []model.FilterV1) ([]model.FilterV1, []model.FilterV1) {
	idBasedFilterKeys := make([]model.FilterV1, 0, 0)
	nonIDBasedFilterKeys := make([]model.FilterV1, 0, 0)
	for _, filter := range filters {
		if strings.Contains(filter.Property, "id") || strings.Contains(filter.Property, "ID") {
			idBasedFilterKeys = append(idBasedFilterKeys, filter)
		} else {
			nonIDBasedFilterKeys = append(nonIDBasedFilterKeys, filter)
		}
	}
	return idBasedFilterKeys, nonIDBasedFilterKeys
}

func splitGroupByBasedOnKeyword(groupBys []model.GroupBy) ([]model.GroupBy, []model.GroupBy) {
	keywordBasedGroupBys := make([]model.GroupBy, 0, 0)
	nonKeywordBasedGroupBys := make([]model.GroupBy, 0, 0)
	for _, groupBy := range groupBys {
		if groupBy.Object == model.AdwordsKeyword {
			keywordBasedGroupBys = append(keywordBasedGroupBys, groupBy)
		} else {
			nonKeywordBasedGroupBys = append(nonKeywordBasedGroupBys, groupBy)
		}
	}
	return keywordBasedGroupBys, nonKeywordBasedGroupBys
}

// @Kark TODO v1
func getHierarchyIdsFromGroupBysForReports(groupBys []model.GroupBy) []string {
	uniqueIDColumns := make([]string, 0, 0)
	uniqueGroupByObjects := make(map[string]struct{})
	for _, groupBy := range groupBys {
		uniqueGroupByObjects[groupBy.Object] = struct{}{}
	}

	for key := range uniqueGroupByObjects {
		key := key + ":id"
		uniqueIDColumn := objectAndPropertyToValueInsideAdwordsReportsMapping[key]
		uniqueIDColumns = append(uniqueIDColumns, uniqueIDColumn)
	}
	return uniqueIDColumns
}

// @Kark TODO v1
func appendSelectMetricsForAdwords(selectQuery string, selectMetrics []string) ([]string, string) {
	selectKeys := make([]string, 0, 0)
	for _, key := range selectMetrics {
		value := adwordsMetricsToAggregatesInReportsMapping[key]
		selectKeys = append(selectKeys, value)
		selectQuery = joinWithComma(selectQuery, value)
	}
	return selectKeys, selectQuery
}

// @Kark TODO v1
// TODO Check if we have none operator
func getFilterPropertiesForAdwordsJob(filters []model.FilterV1) (string, []interface{}) {
	resultStatement := ""
	var filterValue string
	params := make([]interface{}, 0, 0)
	if len(filters) == 0 {
		return resultStatement, params
	}
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
		currentFilterStatement = fmt.Sprintf("value->>? %s ?", filterOperator)
		params = append(params, filter.Property, filterValue)

		if index == 0 {
			resultStatement = fmt.Sprintf("(%s", currentFilterStatement)
		} else {
			resultStatement = fmt.Sprintf("%s %s %s", resultStatement, filter.LogicalOp, currentFilterStatement)
		}
	}
	return resultStatement + ")", params
}

// @Kark TODO v1
// TODO Check if we have none operator
func getFilterPropertiesForAdwordsReports(filters []model.FilterV1) (string, []interface{}) {
	resultStatement := ""
	var filterValue string
	params := make([]interface{}, 0, 0)
	if len(filters) == 0 {
		return resultStatement, params
	}
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
		if strings.Contains(filter.Property, ("id")) {
			currentFilterStatement = fmt.Sprintf("? %s ?", filterOperator)
		} else {
			currentFilterStatement = fmt.Sprintf("value->>? %s ?", filterOperator)
		}
		key := fmt.Sprintf("%s:%s", filter.Object, filter.Property)
		params = append(params, objectAndPropertyToValueInsideAdwordsReportsMapping[key], filterValue)
		if index == 0 {
			resultStatement = fmt.Sprintf("(%s", currentFilterStatement)
		} else {
			resultStatement = fmt.Sprintf("%s %s %s", resultStatement, filter.LogicalOp, currentFilterStatement)
		}
	}
	return resultStatement + ")", params
}

// @TODO Kark v0
func (pg *Postgres) GetAdwordsChannelResultMeta(projectID uint64, customerAccountID string,
	query *model.ChannelQuery) (*model.ChannelQueryResultMeta, error) {

	customerAccountIDArray := strings.Split(customerAccountID, ",")
	stmnt := "SELECT value->>'currency_code' as currency FROM adwords_documents" +
		" " + "WHERE project_id=? AND customer_account_id IN (?) AND type=? AND timestamp BETWEEN ? AND ?" +
		" " + "ORDER BY timestamp DESC LIMIT 1"

	logCtx := log.WithField("project_id", projectID)

	rows, err := pg.ExecQueryWithContext(stmnt, []interface{}{projectID, customerAccountIDArray,
		model.AdwordsDocumentTypeAlias["customer_account_properties"],
		GetAdwordsDateOnlyTimestamp(query.From),
		GetAdwordsDateOnlyTimestamp(query.To)})
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

	return &model.ChannelQueryResultMeta{Currency: currency}, nil
}

// ExecuteAdwordsChannelQuery - @TODO Kark v0
func (pg *Postgres) ExecuteAdwordsChannelQuery(projectID uint64, query *model.ChannelQuery) (*model.ChannelQueryResult, int) {
	logCtx := log.WithField("project_id", projectID).WithField("query", query)

	if projectID == 0 || query == nil {
		logCtx.Error("Invalid project_id or query on execute adwords channel query.")
		return nil, http.StatusInternalServerError
	}

	projectSetting, errCode := pg.GetProjectSetting(projectID)
	if errCode != http.StatusFound {
		return nil, http.StatusInternalServerError
	}

	if projectSetting.IntAdwordsCustomerAccountId == nil || *projectSetting.IntAdwordsCustomerAccountId == "" {
		logCtx.Error("Execute adwords channel query failed. No customer account id.")
		return nil, http.StatusInternalServerError
	}

	queryResult := &model.ChannelQueryResult{}
	meta, err := pg.GetAdwordsChannelResultMeta(projectID,
		*projectSetting.IntAdwordsCustomerAccountId, query)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get adwords channel result meta.")
		return queryResult, http.StatusInternalServerError
	}
	queryResult.Meta = meta

	metricKvs, err := pg.getAdwordsMetrics(projectID, *projectSetting.IntAdwordsCustomerAccountId, query)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get adwords metric kvs.")
		return queryResult, http.StatusInternalServerError
	}
	queryResult.Metrics = metricKvs

	// Return, if no breakdown.
	if query.Breakdown == "" {
		return queryResult, http.StatusOK
	}

	metricBreakdown, err := pg.getAdwordsMetricsBreakdown(projectID,
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
func (pg *Postgres) GetAdwordsFilterValuesByType(projectID uint64, docType int) ([]string, int) {
	projectSetting, errCode := pg.GetProjectSetting(projectID)
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
		docType = model.AdwordsDocumentTypeAlias["campaign_performance_report"]
	case CAFilterAd:
		docType = model.AdwordsDocumentTypeAlias["ad_performance_report"]
	case CAFilterKeyword:
		docType = model.AdwordsDocumentTypeAlias["keyword_performance_report"]
	case CAFilterAdGroup:
		docType = model.AdwordsDocumentTypeAlias["ad_group_performance_report"]
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
func (pg *Postgres) getAdwordsMetricsQuery(projectID uint64, customerAccountID string, query *model.ChannelQuery,
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
func (pg *Postgres) getAdwordsMetrics(projectID uint64, customerAccountID string,
	query *model.ChannelQuery) (*map[string]interface{}, error) {

	stmnt, params, err := pg.getAdwordsMetricsQuery(projectID, customerAccountID, query, false)
	if err != nil {
		return nil, err
	}

	rows, err := pg.ExecQueryWithContext(stmnt, params)
	if err != nil {
		return nil, err
	}

	resultHeaders, resultRows, err := U.DBReadRows(rows)

	if err != nil {
		return nil, err
	}
	if len(resultRows) == 0 {
		log.Warn("Aggregate query returned zero rows.")
		return nil, errors.New("no rows returned")
	}

	if len(resultRows) > 1 {
		log.Warn("Aggregate query returned more than one row on get adwords metric kvs.")
	}

	metricKvs := make(map[string]interface{})
	for i, k := range resultHeaders {
		metricKvs[k] = resultRows[0][i]
	}

	return &metricKvs, nil
}

// @TODO Kark v0
func (pg *Postgres) getAdwordsMetricsBreakdown(projectID uint64, customerAccountID string,
	query *model.ChannelQuery) (*model.ChannelBreakdownResult, error) {

	logCtx := log.WithField("project_id", projectID).WithField("customer_account_id", customerAccountID)

	stmnt, params, err := pg.getAdwordsMetricsQuery(projectID, customerAccountID, query, true)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get adwords metrics query.")
		return nil, err
	}

	rows, err := pg.ExecQueryWithContext(stmnt, params)
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

	return &model.ChannelBreakdownResult{Headers: resultHeaders, Rows: resultRows}, nil
}
