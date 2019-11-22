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
	ProjectId         uint64           `gorm:"primary_key:true" json:"project_id"`
	CustomerAccountId string           `gorm:"primary_key:true" json:"customer_acc_id"`
	TypeAlias         string           `gorm:"-" json:"type_alias"`
	Type              int              `gorm:"primary_key:true" json:"-"`
	Timestamp         int64            `gorm:"primary_key:true" json:"timestamp"`
	ID                string           `gorm:"primary_key:true" json:"id"`
	Values            []postgres.Jsonb `gorm: "-" json:"values"` // List of value.
	Value             *postgres.Jsonb  `json:"-"`
	CreatedAt         time.Time        `json:"created_at"`
	UpdatedAt         time.Time        `json:"updated_at"`
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

func getIdFieldNameByType(docType int) string {
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

func getIdByType(docType int, valueJson *postgres.Jsonb) (string, error) {
	if docType > len(documentTypeByAlias) {
		return "", errors.New("invalid document type")
	}

	valueMap, err := U.DecodePostgresJsonb(valueJson)
	if err != nil {
		return "", err
	}

	idFieldName := getIdFieldNameByType(docType)
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

// Builds a new adword document with each value on values list.
func getNewAdwordsDocumentFromValuesList(adwordsDocumentWithValues *AdwordsDocument) []AdwordsDocument {
	adwordsDocuments := make([]AdwordsDocument, 0, 0)

	for i := range adwordsDocumentWithValues.Values {
		// Skip insert, if not able to get an id.
		id, err := getIdByType(adwordsDocumentWithValues.Type, &adwordsDocumentWithValues.Values[i])
		if err != nil {
			log.WithFields(log.Fields{"project_id": adwordsDocumentWithValues.ProjectId,
				"timestamp": adwordsDocumentWithValues.Timestamp,
				"value":     adwordsDocumentWithValues.Values[i]}).WithError(err).Error(
				"Failed to add adwords document.")
			continue
		}

		newAdwordsDoc := AdwordsDocument{
			ID:                id,
			ProjectId:         adwordsDocumentWithValues.ProjectId,
			CustomerAccountId: adwordsDocumentWithValues.CustomerAccountId,
			Type:              adwordsDocumentWithValues.Type,
			Timestamp:         adwordsDocumentWithValues.Timestamp,
			Value:             &adwordsDocumentWithValues.Values[i],
		}

		adwordsDocuments = append(adwordsDocuments, newAdwordsDoc)
	}

	return adwordsDocuments
}

func CreateAdwordsDocument(adwordsDoc *AdwordsDocument) int {
	logCtx := log.WithField("customer_acc_id", adwordsDoc.CustomerAccountId)

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

	newAdwordsDocs := getNewAdwordsDocumentFromValuesList(adwordsDoc)

	// Todo: Make it bulk insert.
	failure := false
	db := C.GetServices().Db
	for i := range newAdwordsDocs {
		// Todo: db.Create(newAdwordsDocs[i]) causes unaddressable value error. Find why?
		_, err := db.Raw("INSERT INTO adwords_documents (project_id,customer_account_id,type,timestamp,id,value) VALUES (?, ?, ?, ?, ?, ?)",
			newAdwordsDocs[i].ProjectId, newAdwordsDocs[i].CustomerAccountId, newAdwordsDocs[i].Type,
			newAdwordsDocs[i].Timestamp, newAdwordsDocs[i].ID, newAdwordsDocs[i].Value).Rows()
		if err != nil {
			logCtx.WithError(err).Error(
				"Failed to create an adwords doc. Continued inserting other docs.")
			failure = true
		}
	}

	if failure {
		return http.StatusInternalServerError
	}

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
