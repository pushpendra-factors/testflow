package model

import (
	"errors"
	C "factors/config"
	U "factors/util"
	"fmt"
	"net/http"
	"sort"
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

var facebookDocumentTypeAlias = map[string]int{
	"ad_account":        7,
	"campaign":          1,
	"ad":                2,
	"ad_set":            3,
	"ad_insights":       4,
	"campaign_insights": 5,
	"ad_set_insights":   6,
}

const platform = "platform"

const errorDuplicateFacebookDocument = "pq: duplicate key value violates unique constraint \"facebook_documents_pkey\""

const facebookFilterQueryStr = "SELECT DISTINCT(value->>?) as filter_value FROM facebook_documents WHERE project_id = ? AND" +
	" " + "customer_ad_account_id = ? AND type = ? LIMIT 5000"

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
	if docType > len(AdwordsDocumentTypeAlias) {
		return 0, 0, 0, errors.New("invalid document type")
	}

	valueMap, err := U.DecodePostgresJsonb(valueJSON)
	if err != nil {
		return 0, 0, 0, err
	}

	if len(*valueMap) == 0 {
		return 0, 0, 0, errorEmptyAdwordsDocument
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

// GetFacebookFilterValues - @TODO Kark v1
func GetFacebookFilterValues(projectID uint64, filterObject string, filterProperty string, reqID string) ([]interface{}, int) {
	docType, property, errCode := getFacebookDocumentTypeAndPropertyKeyForFilter(filterObject, filterProperty)
	if errCode != http.StatusFound {
		return []interface{}{}, errCode
	}

	filterValues, errCode := getFacebookFilterValuesByType(projectID, docType, property, reqID)
	if errCode != http.StatusFound {
		return []interface{}{}, http.StatusInternalServerError
	}

	return filterValues, http.StatusFound
}

// GetFacebookSQLQueryAndParametersForFilterValues - @TODO Kark v1
func GetFacebookSQLQueryAndParametersForFilterValues(projectID uint64, filterObject string, filterProperty string) (string, []interface{}, int) {
	docType, property, errCode := getFacebookDocumentTypeAndPropertyKeyForFilter(filterObject, filterProperty)
	if errCode != http.StatusFound {
		return "", []interface{}{}, errCode
	}

	projectSetting, errCode := GetProjectSetting(projectID)
	if errCode != http.StatusFound {
		return "", []interface{}{}, http.StatusInternalServerError
	}
	customerAccountID := projectSetting.IntFacebookAdAccount
	params := []interface{}{property, projectID, customerAccountID, docType}

	return "(" + facebookFilterQueryStr + ")", params, http.StatusFound
}

// @TODO Kark v1
func getFacebookDocumentTypeAndPropertyKeyForFilter(filterObject string, filterProperty string) (int, string, int) {
	docType, err := getFacebookDocumentTypeForFilterKey(filterObject)

	if err != nil {
		return 0, "", http.StatusInternalServerError
	}

	property, err := getFacebookFilterPropertyKeyByType(filterObject, filterProperty)
	if err != nil {
		return 0, "", http.StatusBadRequest
	}

	return docType, property, http.StatusFound
}

// @TODO Kark v1
// TODO IMP (impl as part of other PR) - Mapping of adGroup to adSet.
func getFacebookDocumentTypeForFilterKey(filterObject string) (int, error) {
	var docType int = -1

	switch filterObject {
	case CAFilterCampaign:
		docType = facebookDocumentTypeAlias["campaign"]
	case CAFilterAdGroup:
		docType = facebookDocumentTypeAlias["ad_set"]
	case CAFilterAd:
		docType = facebookDocumentTypeAlias["ad"]
	}

	if docType == -1 {
		return docType, errors.New("no adwords document type for filter")
	}

	return docType, nil
}

var facebookRequestPropertiesToSQLProperty = map[string]string{
	"name": "name",
	"id":   "id",
}

// @TODO Kark v1
// TODO: Check the value of filterObject. It should be adset
func getFacebookFilterPropertyKeyByType(filterObject string, filterProperty string) (string, error) {
	property, isPropertyPresent := facebookRequestPropertiesToSQLProperty[filterProperty]

	if !isPropertyPresent {
		return "", errors.New("no filter key found for document type")
	}
	return property, nil
}

// @TODO Kark v1
func getFacebookFilterValuesByType(projectID uint64, docType int, property string, reqID string) ([]interface{}, int) {
	projectSetting, errCode := GetProjectSetting(projectID)
	if errCode != http.StatusFound {
		return []interface{}{}, http.StatusInternalServerError
	}
	customerAccountID := projectSetting.IntFacebookAdAccount

	logCtx := log.WithField("project_id", projectID).WithField("doc_type", docType).WithField("req_id", reqID)
	params := []interface{}{property, projectID, customerAccountID, docType}
	_, resultRows, _ := ExecuteSQL(adwordsFilterQueryStr, params, logCtx)

	return Convert2DArrayTo1DArray(resultRows), http.StatusFound
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

	strmntWhere := "WHERE project_id= ? AND customer_ad_account_id = ? AND timestamp>? AND timestamp<? AND type=? and platform!='facebook_all'"

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
		log.Error("Aggregate query returned more than one row on get adwords metric kvs.")
	}

	metricKvs := make(map[string]interface{})
	for i, k := range resultHeaders {
		metricKvs[k] = resultRows[0][i]
	}

	queryResult.Metrics = &metricKvs
	return queryResult, nil
}
