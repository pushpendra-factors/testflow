package model

import (
	"errors"
	C "factors/config"
	U "factors/util"
	"fmt"
	"net/http"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

type AdwordsDocument struct {
	ProjectId         uint64          `gorm:"primary_key:true" json:"project_id"`
	CustomerAccountId string          `gorm:"primary_key:true" json:"customer_acc_id"`
	TypeAlias         string          `gorm:"-" json:"type_alias"`
	Type              int             `gorm:"primary_key:true" json:"-"`
	Timestamp         int64           `gorm:"primary_key:true" json:"timestamp"`
	ID                string          `gorm:"primary_key:true" json:"id"`
	Value             *postgres.Jsonb `json:"value"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
}

var documentTypeByAlias = map[string]int{
	"campaigns":                   1,
	"ads":                         2,
	"ad_groups":                   3,
	"click_performance_report":    4,
	"campaign_performance_report": 5,
	"ad_performance_report":       6,
	"search_performance_report":   7,
	"keyword_performance_report":  8,
}

const error_DuplicateAdwordsDocument = "pq: duplicate key value violates unique constraint \"adwords_documents_pkey\""
const filterValueAll = "all"

func isDuplicateAdwordsDocumentError(err error) bool {
	return err.Error() == error_DuplicateAdwordsDocument
}

func getAdwordsIdFieldNameByType(docType int) string {
	switch docType {
	case 4: // click_performance_report
		return "gcl_id"
	case 5: // campaign_performance_report
		return "campaign_id"
	case 7: // search_performance_report
		return "query"
	default: // others
		return "id"
	}
}

func getAdwordsIdByType(docType int, valueJson *postgres.Jsonb) (string, error) {
	if docType > len(documentTypeByAlias) {
		return "", errors.New("invalid document type")
	}

	valueMap, err := U.DecodePostgresJsonb(valueJson)
	if err != nil {
		return "", err
	}

	idFieldName := getAdwordsIdFieldNameByType(docType)
	id, exists := (*valueMap)[idFieldName]
	if !exists {
		return "", fmt.Errorf("id field %s does not exist on doc of type %s", idFieldName, docType)
	}

	if id == nil {
		return "", fmt.Errorf("id field %s has empty value on doc of type %s", idFieldName, docType)
	}

	idStr, err := U.GetValueAsString(id)
	if err != nil {
		return "", err
	}

	// Id as string always.
	return idStr, nil
}

func CreateAdwordsDocument(adwordsDoc *AdwordsDocument) int {
	logCtx := log.WithField("customer_acc_id", adwordsDoc.CustomerAccountId).WithField(
		"project_id", adwordsDoc.ProjectId)

	if adwordsDoc.CustomerAccountId == "" || adwordsDoc.TypeAlias == "" {
		logCtx.Error("Invalid adwords document.")
		return http.StatusBadRequest
	}

	logCtx = logCtx.WithField("type_alias", adwordsDoc.TypeAlias)
	docType, docTypeExists := documentTypeByAlias[adwordsDoc.TypeAlias]
	if !docTypeExists {
		logCtx.Error("Invalid type alias.")
		return http.StatusBadRequest
	}
	adwordsDoc.Type = docType

	adwordsDocId, err := getAdwordsIdByType(adwordsDoc.Type, adwordsDoc.Value)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get id by adowords doc type.")
		return http.StatusInternalServerError
	}
	adwordsDoc.ID = adwordsDocId

	db := C.GetServices().Db
	// Todo: db.Create(newAdwordsDocs[i]) causes unaddressable value error. Find why?
	queryStr := "INSERT INTO adwords_documents (project_id,customer_account_id,type,timestamp,id,value) VALUES (?, ?, ?, ?, ?, ?)"
	rows, err := db.Raw(queryStr, adwordsDoc.ProjectId, adwordsDoc.CustomerAccountId,
		adwordsDoc.Type, adwordsDoc.Timestamp, adwordsDoc.ID, adwordsDoc.Value).Rows()
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
	for alias, typ := range documentTypeByAlias {
		documentTypeMap[typ] = alias
	}

	return documentTypeMap
}

func GetAllAdwordsLastSyncInfoByProjectAndType() ([]AdwordsLastSyncInfo, int) {
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

	adwordsSettingsByProject := make(map[uint64]*AdwordsProjectSettings, 0)
	for _, settings := range adwordsSettings {
		adwordsSettingsByProject[settings.ProjectId] = &settings
	}

	documentTypeAliasByType := getDocumentTypeAliasByType()

	// add settings for project_id existing on adwords documents.
	existingProjects := make(map[uint64]bool, 0)
	selectedLastSyncInfos := make([]AdwordsLastSyncInfo, 0, 0)
	for i := range adwordsLastSyncInfos {
		logCtx := log.WithField("project_id", adwordsLastSyncInfos[i].ProjectId)

		settings, exists := adwordsSettingsByProject[adwordsLastSyncInfos[i].ProjectId]
		if !exists {
			logCtx.Error("Adwords project settings not found for project adwords synced earlier.")
		}

		// Do not select sync info, if customer_account_id mismatch, as user
		// would have changed customer_account mapped to project.
		if adwordsLastSyncInfos[i].CustomerAccountId != settings.CustomerAccountId {
			logCtx.Warn("customer_account_id mapped to project has been changed.")
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
		existingProjects[adwordsLastSyncInfos[i].ProjectId] = true
	}

	// add new projects.
	for _, settings := range adwordsSettings {
		if _, exists := existingProjects[settings.ProjectId]; exists {
			continue
		}

		// add sync info for each document type.
		for docType, _ := range documentTypeByAlias {
			syncInfo := AdwordsLastSyncInfo{
				ProjectId:         settings.ProjectId,
				RefreshToken:      settings.RefreshToken,
				CustomerAccountId: settings.CustomerAccountId,
				LastTimestamp:     0, // no sync yet.
				DocumentTypeAlias: docType,
			}

			selectedLastSyncInfos = append(selectedLastSyncInfos, syncInfo)
		}
	}

	return selectedLastSyncInfos, http.StatusOK
}

func getAdwordsFilterKeyByType(docType int) (string, error) {
	filterKeyByType := map[int]string{
		5: "campaign_name",
		8: "criteria",
		6: "id",
	}

	filterKey, filterKeyExists := filterKeyByType[docType]
	if !filterKeyExists {
		return "", errors.New("no filter key found for document type")
	}

	return filterKey, nil
}

func GetAdwordsFilterValuesByType(projectId uint64, docType int) ([]string, int) {
	projectSetting, errCode := GetProjectSetting(projectId)
	if errCode != http.StatusFound {
		return []string{}, http.StatusInternalServerError
	}
	customerAccountId := projectSetting.IntAdwordsCustomerAccountId

	db := C.GetServices().Db
	logCtx := log.WithField("project_id", projectId).WithField("doc_type", docType)

	filterValueKey, err := getAdwordsFilterKeyByType(docType)
	if err != nil {
		logCtx.WithError(err).Error("Unknown doc type for get adwords filter key.")
		return []string{}, http.StatusBadRequest
	}

	queryStr := "SELECT DISTINCT(value->>?) as filter_value FROM adwords_documents WHERE project_id = ? AND" +
		" " + "customer_account_id = ? AND type = ? LIMIT 5000"
	rows, err := db.Raw(queryStr, filterValueKey, projectId, customerAccountId, docType).Rows()
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

func GetAdwordsDocumentTypeForFilterKey(filter string) (int, error) {
	var docType int

	switch filter {
	case CAFilterCampaign:
		docType = documentTypeByAlias["campaign_performance_report"]
	case CAFilterAd:
		docType = documentTypeByAlias["ad_performance_report"]
	case CAFilterKeyword:
		docType = documentTypeByAlias["keyword_performance_report"]
	}

	if docType == 0 {
		return docType, errors.New("no adwords document type for filter")
	}

	return docType, nil
}

// select value->>'impressions' as impressions, value->>'clicks' as clicks,
// value->>'average_cost' as cost_per_click, value->>'cost' as total_cost,
// value->>'conversions' as conversions, value->>'all_conversions' as all_conversions
// from adwords_documents where type=8 and timestamp between 20191120 and 20191120;
func GetAdwordsMetricKvs(projectId uint64, query *ChannelQuery) (*map[string]interface{}, error) {
	stmntWithoutAlias := "SELECT SUM((value->>'impressions')::float) as %s, SUM((value->>'clicks')::float) as %s," +
		" " + "SUM((value->>'cost')::float) as %s, SUM((value->>'conversions')::float) as %s," +
		" " + "SUM((value->>'all_conversions')::float) as %s FROM adwords_documents"

	stmnt := fmt.Sprintf(stmntWithoutAlias, CAColumnImpressions, CAColumnClicks,
		CAColumnTotalCost, CAColumnAllConversions, CAColumnAllConversions)

	stmntWhere := "WHERE type=? AND timestamp BETWEEN ? and ?"

	isWhereByFilterRequired := query.FilterValue != filterValueAll
	if isWhereByFilterRequired {
		stmntWhere = stmntWhere + " " + "and" + " " + "value->>?=?"
	}

	docType, err := GetAdwordsDocumentTypeForFilterKey(query.FilterKey)
	if err != nil {
		return nil, err
	}

	params := make([]interface{}, 0, 0)
	params = append(params, docType, query.DateFrom, query.DateTo)

	if isWhereByFilterRequired {
		filterKey, err := getAdwordsFilterKeyByType(docType)
		if err != nil {
			return nil, err
		}

		params = append(params, filterKey, query.FilterValue)
	}

	// append where to stmnt.
	stmnt = stmnt + " " + stmntWhere

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
