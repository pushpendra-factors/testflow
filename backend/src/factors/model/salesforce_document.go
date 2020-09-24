package model

import (
	"errors"
	U "factors/util"
	"net/http"
	"time"

	C "factors/config"

	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

type SalesforceDocument struct {
	ProjectId uint64           `gorm:"primary_key:true;auto_increment:false" json:"project_id"`
	ID        string           `gorm:"primary_key:true;auto_increment:false" json:"id"`
	Type      int              `gorm:"primary_key:true;auto_increment:false" json:"-"`
	Action    SalesforceAction `gorm:"auto_increment:false;not null" json:"action"`
	Timestamp int64            `gorm:"primary_key:true;auto_increment:false" json:"timestamp"`
	TypeAlias string           `gorm:"-" json:"type_alias"`
	Value     *postgres.Jsonb  `json:"value"`
	Synced    bool             `gorm:"default:false;not null" json:"synced"`
	SyncId    string           `gorm:"default:null" json:"sync_id"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
}

type SalesforceAction int

const (
	SalesforceDocumentTypeContact = 1
	SalesforceDocumentTypeLead    = 2
	SalesforceDocumentTypeAccount = 3

	SalesforceDocumentTypeNameContact = "contact"
	SalesforceDocumentTypeNameLead    = "lead"
	SalesforceDocumentTypeNameAccount = "account"

	SalesforceDocumentCreated SalesforceAction = 1
	SalesforceDocumentUpdated SalesforceAction = 2

	SalesforceDocumentTimeLayout = "2006-01-02T15:04:05.000-0700"
)

var SalesforceDocumentTypeAlias = map[string]int{
	SalesforceDocumentTypeNameContact: SalesforceDocumentTypeContact,
	SalesforceDocumentTypeNameLead:    SalesforceDocumentTypeLead,
	SalesforceDocumentTypeNameAccount: SalesforceDocumentTypeAccount,
}

var SalesforceSupportedDocumentType = []int{
	SalesforceDocumentTypeAccount,
	SalesforceDocumentTypeContact,
	SalesforceDocumentTypeLead,
}

var SalesforceSkippablefields = []string{
	"LastModifiedDate",
	"CreatedDate",
	"CreatedById",
	"LastModifiedById",
	"SystemModstamp",
	"attributes",
}

var errorDuplicateRecord = errors.New("duplicate record")

func GetSalesforceAliasByDocType(typ int) string {
	for a, t := range SalesforceDocumentTypeAlias {
		if typ == t {
			return a
		}
	}

	return ""
}
func getSalesforceDocTypeByAlias(alias string) (int, error) {
	if alias == "" {
		return 0, errors.New("empty document type alias")
	}

	typ, typExists := SalesforceDocumentTypeAlias[alias]
	if !typExists {
		return 0, errors.New("invalid document type alias")
	}

	return typ, nil
}

func GetSalesforceCreatedEventName(docType int) string {
	switch docType {
	case SalesforceDocumentTypeAccount:
		return U.EVENT_NAME_SALESFORCE_ACCOUNT_CREATED
	case SalesforceDocumentTypeContact:
		return U.EVENT_NAME_SALESFORCE_CONTACT_CREATED
	case SalesforceDocumentTypeLead:
		return U.EVENT_NAME_SALESFORCE_LEAD_CREATED
	}
	return ""
}

func GetSalesforceUpdatedEventName(docType int) string {
	switch docType {
	case SalesforceDocumentTypeAccount:
		return U.EVENT_NAME_SALESFORCE_ACCOUNT_UPDATED
	case SalesforceDocumentTypeContact:
		return U.EVENT_NAME_SALESFORCE_CONTACT_UPDATED
	case SalesforceDocumentTypeLead:
		return U.EVENT_NAME_SALESFORCE_LEAD_UPDATED
	}
	return ""
}

type SalesforceLastSyncInfo struct {
	ProjectId uint64 `json:"-"`
	Type      int    `json:"type"`
	Timestamp int64  `json:"timestamp"`
}

type SalesforceSyncInfo struct {
	ProjectSettings map[uint64]*SalesforceProjectSettings `json:"project_settings"`
	// project_id: { type: last_sync_info }
	LastSyncInfo map[uint64]map[string]int64 `json:"last_sync_info"`
}

func GetSalesforceSyncInfo() (*SalesforceSyncInfo, int) {
	var lastSyncInfo []SalesforceLastSyncInfo

	db := C.GetServices().Db
	err := db.Table("salesforce_documents").Select(
		"project_id, type, MAX(timestamp) as timestamp").Group(
		"project_id, type").Find(&lastSyncInfo).Error
	if err != nil {
		return nil, http.StatusInternalServerError
	}

	lastSyncInfoByProject := make(map[uint64]map[string]int64, 0)
	for _, syncInfo := range lastSyncInfo {
		if _, projectExists := lastSyncInfoByProject[syncInfo.ProjectId]; !projectExists {
			lastSyncInfoByProject[syncInfo.ProjectId] = make(map[string]int64)
		}

		lastSyncInfoByProject[syncInfo.ProjectId][GetSalesforceAliasByDocType(syncInfo.Type)] = syncInfo.Timestamp
	}

	enabledProjectLastSync := make(map[uint64]map[string]int64, 0)

	projectSettings, errCode := GetAllSalesforceProjectSettings()
	if errCode != http.StatusFound {
		return nil, http.StatusInternalServerError
	}

	settingsByProject := make(map[uint64]*SalesforceProjectSettings, 0)
	for i, ps := range projectSettings {
		_, pExists := lastSyncInfoByProject[ps.ProjectId]

		if !pExists {
			// add projects not synced before.
			enabledProjectLastSync[ps.ProjectId] = make(map[string]int64, 0)
		} else {
			// add sync info if avaliable.
			enabledProjectLastSync[ps.ProjectId] = lastSyncInfoByProject[ps.ProjectId]
		}

		// add types not synced before.
		for typ := range SalesforceDocumentTypeAlias {
			_, typExists := enabledProjectLastSync[ps.ProjectId][typ]
			if !typExists {
				// last sync timestamp as zero as type not synced before.
				enabledProjectLastSync[ps.ProjectId][typ] = 0
			}
		}

		settingsByProject[projectSettings[i].ProjectId] = &projectSettings[i]
	}

	var syncInfo SalesforceSyncInfo
	syncInfo.LastSyncInfo = enabledProjectLastSync
	syncInfo.ProjectSettings = settingsByProject

	return &syncInfo, http.StatusOK
}

func getSalesforceDocumentId(document *SalesforceDocument) (string, error) {
	documentMap, err := U.DecodePostgresJsonb(document.Value)
	if err != nil {
		return "", err
	}

	id, idExists := (*documentMap)["Id"]
	if !idExists {
		return "", errors.New("id key not exist on salesforce document")
	}

	idAsString := U.GetPropertyValueAsString(id)
	if idAsString == "" {
		return "", errors.New("invalid id on salesforce document")
	}
	return idAsString, nil
}

func getSalesforceDocumentByIdAndType(projectId uint64, id string, docType int) ([]SalesforceDocument, int) {
	logCtx := log.WithFields(log.Fields{"project_id": projectId, "id": id, "type": docType})

	var documents []SalesforceDocument
	if projectId == 0 || id == "" || docType == 0 {
		logCtx.Error("Failed to get salesforce document by id and type. Invalid project_id or id or type.")
		return documents, http.StatusBadRequest
	}

	db := C.GetServices().Db
	err := db.Where("project_id = ? AND id = ? AND type = ?", projectId, id,
		docType).Find(&documents).Error
	if err != nil {
		logCtx.WithError(err).Error("Failed to get salesforce documents.")
		return documents, http.StatusInternalServerError
	}

	if len(documents) == 0 {
		return documents, http.StatusNotFound
	}

	return documents, http.StatusFound
}

func CreateSalesforceDocument(projectId uint64, document *SalesforceDocument) int {
	logCtx := log.WithField("project_id", document.ProjectId)
	if projectId == 0 {
		logCtx.Error("Invalid project_id on create salesforce document.")
		return http.StatusBadRequest
	}
	document.ProjectId = projectId

	documentType, err := getSalesforceDocTypeByAlias(document.TypeAlias)
	if err != nil {
		logCtx.WithError(err).Error("Invalid type on create salesforce document.")
		return http.StatusBadRequest
	}
	document.Type = documentType

	if U.IsEmptyPostgresJsonb(document.Value) {
		logCtx.Error("Empty document value on create salesforce document.")
		return http.StatusBadRequest
	}

	documentId, err := getSalesforceDocumentId(document)
	if err != nil {
		logCtx.WithError(err).Error(
			"Failed to get id for salesforce document on create.")
		return http.StatusInternalServerError
	}
	document.ID = documentId

	logCtx = logCtx.WithField("type", document.Type).WithField("value", document.Value)

	_, errCode := getSalesforceDocumentByIdAndType(document.ProjectId,
		document.ID, document.Type)
	if errCode == http.StatusInternalServerError || errCode == http.StatusBadRequest {
		return errCode
	}

	isNew := errCode == http.StatusNotFound
	if isNew {
		err = CreateSalesforceDocumentByAction(projectId, document, SalesforceDocumentCreated)
		if err != nil {
			logCtx.WithError(err).Error("Failed to create salesforce document.")
			return http.StatusInternalServerError
		}
	}

	err = CreateSalesforceDocumentByAction(projectId, document, SalesforceDocumentUpdated)
	if err != nil {
		if err == errorDuplicateRecord {
			return http.StatusConflict
		}

		logCtx.WithError(err).Error("Failed to create salesforce document.")
		return http.StatusInternalServerError
	}

	return http.StatusCreated
}

func CreateSalesforceDocumentByAction(projectId uint64, document *SalesforceDocument, action SalesforceAction) error {
	document.Action = action
	timestamp, err := GetSalesforceDocumentTimestampByAction(document)
	if err != nil {
		return err
	}
	document.Timestamp = timestamp

	db := C.GetServices().Db
	err = db.Create(document).Error
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code.Name() == "unique_violation" {
			return errorDuplicateRecord
		}

		return err
	}

	return nil
}

func GetSalesforceDocumentTimestampByAction(document *SalesforceDocument) (int64, error) {
	if document.Type == 0 {
		return 0, errors.New("invalid document type")
	}

	value, err := U.DecodePostgresJsonb(document.Value)
	if err != nil {
		return 0, err
	}

	var dateKey string
	if document.Action == SalesforceDocumentCreated {
		dateKey = "CreatedDate"
	}
	if document.Action == SalesforceDocumentUpdated {
		dateKey = "LastModifiedDate"
	}

	date, exists := (*value)[dateKey]
	if !exists || date == nil {
		return 0, errors.New("failed to get date")
	}

	timestamp, err := readSalesforceDocumentTimestamp(date)
	if err != nil {
		return 0, err
	}
	return timestamp, nil
}

func readSalesforceDocumentTimestamp(timestamp interface{}) (int64, error) {
	timestampStr, ok := timestamp.(string)
	if !ok || timestampStr == "" {
		return 0, errors.New("invalid timestamp")
	}

	t, err := time.Parse(SalesforceDocumentTimeLayout, timestampStr)
	if err != nil {
		return 0, err
	}

	return t.Unix(), nil
}

func UpdateSalesforceDocumentAsSynced(projectId uint64, id string, syncId string) int {
	logCtx := log.WithField("project_id", projectId).WithField("id", id)

	updates := make(map[string]interface{}, 0)
	updates["synced"] = true
	if syncId != "" {
		updates["sync_id"] = syncId
	}

	db := C.GetServices().Db
	err := db.Model(&SalesforceDocument{}).Where("project_id = ? AND id = ?",
		projectId, id).Updates(updates).Error
	if err != nil {
		logCtx.WithError(err).Error("Failed to update salesforce document as synced.")
		return http.StatusInternalServerError
	}

	return http.StatusAccepted
}
