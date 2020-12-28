package model

import (
	"encoding/json"
	"errors"
	U "factors/util"
	"fmt"
	"net/http"
	"time"

	C "factors/config"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

// SalesforceDocument is an interface for salesforce_documents table
type SalesforceDocument struct {
	ProjectID uint64           `gorm:"primary_key:true;auto_increment:false" json:"project_id"`
	ID        string           `gorm:"primary_key:true;auto_increment:false" json:"id"`
	Type      int              `gorm:"primary_key:true;auto_increment:false" json:"-"`
	Action    SalesforceAction `gorm:"auto_increment:false;not null" json:"action"`
	Timestamp int64            `gorm:"primary_key:true;auto_increment:false" json:"timestamp"`
	TypeAlias string           `gorm:"-" json:"type_alias"`
	Value     *postgres.Jsonb  `json:"value"`
	Synced    bool             `gorm:"default:false;not null" json:"synced"`
	SyncID    string           `gorm:"default:null" json:"sync_id"`
	UserID    string           `gorm:"default:null", json:"-"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
}

// SalesforceAction defines the existance of record in salesforce_documents
type SalesforceAction int

/*
 Salesforce supported document types and their alias
*/
const (
	SalesforceDocumentTypeContact     = 1
	SalesforceDocumentTypeLead        = 2
	SalesforceDocumentTypeAccount     = 3
	SalesforceDocumentTypeOpportunity = 4

	SalesforceDocumentTypeNameContact     = "contact"
	SalesforceDocumentTypeNameLead        = "lead"
	SalesforceDocumentTypeNameAccount     = "account"
	SalesforceDocumentTypeNameOpportunity = "opportunity"

	SalesforceDocumentCreated SalesforceAction = 1
	SalesforceDocumentUpdated SalesforceAction = 2

	// Standard template for salesforce date time
	SalesforceDocumentTimeLayout = "2006-01-02T15:04:05.000-0700"
)

// SalesforceDocumentTypeAlias maps document type to alias
var SalesforceDocumentTypeAlias = map[string]int{
	SalesforceDocumentTypeNameContact:     SalesforceDocumentTypeContact,
	SalesforceDocumentTypeNameLead:        SalesforceDocumentTypeLead,
	SalesforceDocumentTypeNameAccount:     SalesforceDocumentTypeAccount,
	SalesforceDocumentTypeNameOpportunity: SalesforceDocumentTypeOpportunity,
}

// SalesforceStandardDocumentType will be pulled if no custom list is provided
var SalesforceStandardDocumentType = []int{
	SalesforceDocumentTypeAccount,
	SalesforceDocumentTypeContact,
	SalesforceDocumentTypeLead,
}

var errorDuplicateRecord = errors.New("duplicate record")

// GetSalesforceAliasByDocType return name for the doc type
func GetSalesforceAliasByDocType(typ int) string {
	for a, t := range SalesforceDocumentTypeAlias {
		if typ == t {
			return a
		}
	}

	return ""
}

// GetSalesforceDocTypeByAlias return number representing the doc type name
func GetSalesforceDocTypeByAlias(alias string) int {
	if alias == "" {
		return 0
	}

	typ, typExists := SalesforceDocumentTypeAlias[alias]
	if !typExists {
		return 0
	}

	return typ
}

// GetSalesforceDocumentTypeAlias returns a configured map of doc type name and its corresponding number
func GetSalesforceDocumentTypeAlias(projectID uint64) map[string]int {
	docTypes := make(map[string]int)
	for _, doctype := range GetSalesforceAllowedObjects(projectID) {
		docTypes[GetSalesforceAliasByDocType(doctype)] = doctype
	}
	return docTypes
}

// GetSalesforceEventNameByAction creates event name by SalesforceAction and doc type name
func GetSalesforceEventNameByAction(doc *SalesforceDocument, action SalesforceAction) string {
	typAlias := GetSalesforceAliasByDocType(doc.Type)

	if typAlias != "" {
		if action == SalesforceDocumentCreated {
			return fmt.Sprintf("$sf_%s_created", typAlias)
		}
		if action == SalesforceDocumentUpdated {
			return fmt.Sprintf("$sf_%s_updated", typAlias)
		}
	}

	return ""
}

// SalesforceLastSyncInfo contains information about the latest timestamp and type of document for a project
type SalesforceLastSyncInfo struct {
	ProjectID uint64 `json:"-"`
	Type      int    `json:"type"`
	Timestamp int64  `json:"timestamp"`
}

// SalesforceSyncInfo lists project_id and their last sync info per doc type
type SalesforceSyncInfo struct {
	ProjectSettings map[uint64]*SalesforceProjectSettings `json:"project_settings"`
	// project_id: { type: last_sync_info }
	LastSyncInfo map[uint64]map[string]int64 `json:"last_sync_info"`
}

// GetSalesforceSyncInfo returns list of projects and their corresponding sync status
func GetSalesforceSyncInfo() (SalesforceSyncInfo, int) {
	var lastSyncInfo []SalesforceLastSyncInfo
	var syncInfo SalesforceSyncInfo

	db := C.GetServices().Db
	err := db.Table("salesforce_documents").Select(
		"project_id, type, MAX(timestamp) as timestamp").Group(
		"project_id, type").Find(&lastSyncInfo).Error
	if err != nil {
		return syncInfo, http.StatusInternalServerError
	}

	lastSyncInfoByProject := make(map[uint64]map[string]int64, 0)
	for _, syncInfo := range lastSyncInfo {
		if _, projectExists := lastSyncInfoByProject[syncInfo.ProjectID]; !projectExists {
			lastSyncInfoByProject[syncInfo.ProjectID] = make(map[string]int64)
		}

		lastSyncInfoByProject[syncInfo.ProjectID][GetSalesforceAliasByDocType(syncInfo.Type)] = syncInfo.Timestamp
	}

	enabledProjectLastSync := make(map[uint64]map[string]int64, 0)

	projectSettings, errCode := GetAllSalesforceProjectSettings()
	if errCode != http.StatusFound {
		return syncInfo, http.StatusInternalServerError
	}

	settingsByProject := make(map[uint64]*SalesforceProjectSettings, 0)
	for i, ps := range projectSettings {
		_, pExists := lastSyncInfoByProject[ps.ProjectID]

		if !pExists {
			// add projects not synced before.
			enabledProjectLastSync[ps.ProjectID] = make(map[string]int64, 0)
		} else {
			// add sync info if avaliable.
			enabledProjectLastSync[ps.ProjectID] = lastSyncInfoByProject[ps.ProjectID]
		}

		// add types not synced before.
		for typ := range GetSalesforceDocumentTypeAlias(ps.ProjectID) {
			_, typExists := enabledProjectLastSync[ps.ProjectID][typ]
			if !typExists {
				// last sync timestamp as zero as type not synced before.
				enabledProjectLastSync[ps.ProjectID][typ] = 0
			}
		}

		settingsByProject[projectSettings[i].ProjectID] = &projectSettings[i]
	}

	syncInfo.LastSyncInfo = enabledProjectLastSync
	syncInfo.ProjectSettings = settingsByProject

	return syncInfo, http.StatusFound
}

func getSalesforceDocumentID(document *SalesforceDocument) (string, error) {
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

// GetSyncedSalesforceDocumentByType return salesforce_documents by doc type which are synced
func GetSyncedSalesforceDocumentByType(projectID uint64, ids []string,
	docType int) ([]SalesforceDocument, int) {

	logCtx := log.WithFields(log.Fields{"project_id": projectID, "ids": ids,
		"type": docType})

	var documents []SalesforceDocument
	if projectID == 0 || len(ids) == 0 || docType == 0 {
		logCtx.Error("Failed to get salesforce document by id and type. Invalid project_id or id or type.")
		return nil, http.StatusBadRequest
	}

	db := C.GetServices().Db
	err := db.Order("timestamp").Where(
		"project_id = ? AND id IN (?) AND type = ? AND synced = true",
		projectID, ids, docType).Find(&documents).Error
	if err != nil {
		logCtx.WithError(err).Error("Failed to get salesforce documents.")
		return nil, http.StatusInternalServerError
	}

	if len(documents) == 0 {
		return nil, http.StatusNotFound
	}

	return documents, http.StatusFound
}

func getSalesforceDocumentByIDAndType(projectID uint64, id string, docType int) ([]SalesforceDocument, int) {
	logCtx := log.WithFields(log.Fields{"project_id": projectID, "id": id, "type": docType})

	var documents []SalesforceDocument
	if projectID == 0 || id == "" || docType == 0 {
		logCtx.Error("Failed to get salesforce document by id and type. Invalid project_id or id or type.")
		return documents, http.StatusBadRequest
	}

	db := C.GetServices().Db
	err := db.Where("project_id = ? AND id = ? AND type = ?", projectID, id,
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

// CreateSalesforceDocument fills required fields before inserting into salesforce_document table
func CreateSalesforceDocument(projectID uint64, document *SalesforceDocument) int {
	logCtx := log.WithField("project_id", document.ProjectID)
	if projectID == 0 {
		logCtx.Error("Invalid project_id on create salesforce document.")
		return http.StatusBadRequest
	}
	document.ProjectID = projectID

	document.Type = GetSalesforceDocTypeByAlias(document.TypeAlias)

	if U.IsEmptyPostgresJsonb(document.Value) {
		logCtx.Error("Empty document value on create salesforce document.")
		return http.StatusBadRequest
	}

	documentID, err := getSalesforceDocumentID(document)
	if err != nil {
		logCtx.WithError(err).Error(
			"Failed to get id for salesforce document on create.")
		return http.StatusInternalServerError
	}
	document.ID = documentID

	logCtx = logCtx.WithField("type", document.Type).WithField("value", document.Value)

	_, errCode := getSalesforceDocumentByIDAndType(document.ProjectID,
		document.ID, document.Type)
	if errCode == http.StatusInternalServerError || errCode == http.StatusBadRequest {
		return errCode
	}

	isNew := errCode == http.StatusNotFound
	if isNew {
		status := CreateSalesforceDocumentByAction(projectID, document, SalesforceDocumentCreated)
		if status != http.StatusOK {
			if status != http.StatusConflict {
				logCtx.Error("Failed to create salesforce document.")
			}

			return status
		}

		return http.StatusCreated
	}

	status := CreateSalesforceDocumentByAction(projectID, document, SalesforceDocumentUpdated)
	if status != http.StatusOK {
		if status != http.StatusConflict {
			logCtx.Error("Failed to create salesforce document.")
		}

		return status
	}

	return http.StatusCreated
}

// CreateSalesforceDocumentByAction inserts salesforce_document to table by SalesforceAction
func CreateSalesforceDocumentByAction(projectID uint64, document *SalesforceDocument, action SalesforceAction) int {
	if projectID == 0 {
		return http.StatusBadRequest
	}

	if action == 0 {
		return http.StatusBadRequest
	}

	document.Action = action
	timestamp, err := getSalesforceLastModifiedTimestamp(document)
	if err != nil {
		log.WithError(err).Error("Failed to get last modified timestamp")
		return http.StatusBadRequest
	}
	document.Timestamp = timestamp

	db := C.GetServices().Db
	err = db.Create(document).Error
	if err != nil {
		if U.IsPostgresUniqueIndexViolationError("salesforce_documents_pkey", err) {
			return http.StatusConflict
		}

		return http.StatusInternalServerError
	}

	return http.StatusOK
}

func getSalesforceLastModifiedTimestamp(document *SalesforceDocument) (int64, error) {
	if document.Type == 0 {
		return 0, errors.New("invalid document type")
	}

	value, err := U.DecodePostgresJsonb(document.Value)
	if err != nil {
		return 0, err
	}

	dateKey := "LastModifiedDate"
	date, exists := (*value)[dateKey]
	if !exists || date == nil {
		return 0, errors.New("failed to get date")
	}

	return GetSalesforceDocumentTimestamp(date)
}

// GetSalesforceDocumentTimestampByAction returns created or last modified timestamp by SalesforceAction
func GetSalesforceDocumentTimestampByAction(document *SalesforceDocument, action SalesforceAction) (int64, error) {
	if document.Type == 0 {
		return 0, errors.New("invalid document type")
	}
	if action == 0 {
		return 0, errors.New("invalid action")
	}

	value, err := U.DecodePostgresJsonb(document.Value)
	if err != nil {
		return 0, err
	}

	if action == SalesforceDocumentUpdated {
		return getSalesforceLastModifiedTimestamp(document)
	}

	dateKey := "CreatedDate"

	date, exists := (*value)[dateKey]
	if !exists || date == nil {
		return 0, errors.New("failed to get date")
	}

	return GetSalesforceDocumentTimestamp(date)
}

// GetLastSyncedSalesforceDocumentByCustomerUserIDORUserID returns latest synced record by customer_user_id or user_id.
func GetLastSyncedSalesforceDocumentByCustomerUserIDORUserID(projectID uint64, customerUserID, userID string, docType int) (*SalesforceDocument, int) {
	if projectID == 0 {
		return nil, http.StatusBadRequest
	}

	if userID == "" || docType == 0 {
		return nil, http.StatusBadRequest
	}

	logCtx := log.WithFields(log.Fields{"project_id": projectID, "user_id": userID, "customer_user_id": customerUserID, "doc_type": docType})

	db := C.GetServices().Db

	var whereStmn string
	var whereParams []interface{}

	if customerUserID != "" {
		userIDs, status := GetAllUserIDByCustomerUserID(projectID, customerUserID)
		if status == http.StatusFound {
			whereStmn = "type = ? AND project_id=? AND user_id IN(?) AND synced = true"
			whereParams = []interface{}{docType, projectID, userIDs}
		} else {
			logCtx.Error("Failed to GetAllUserIDByCustomerUserID.")
		}
	}

	if customerUserID == "" || whereStmn == "" {
		whereStmn = "type = ? AND synced = true AND project_id=? AND user_id = ? "
		whereParams = []interface{}{docType, projectID, userID}
	}

	var document []SalesforceDocument

	if err := db.Where(whereStmn, whereParams...).Order("timestamp DESC").First(&document).Error; err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			logCtx.WithError(err).Error("Failed to get latest salesforce document by userID.")
			return nil, http.StatusInternalServerError
		}
		return nil, http.StatusNotFound
	}
	if len(document) != 1 {
		return nil, http.StatusNotFound
	}

	return &document[0], http.StatusFound
}

func GetSalesforceDocumentTimestamp(timestamp interface{}) (int64, error) {
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

// UpdateSalesforceDocumentAsSynced inserts syncID and updates the status of the document as synced
func UpdateSalesforceDocumentAsSynced(projectID uint64, document *SalesforceDocument, syncID, userID string) int {
	logCtx := log.WithField("project_id", projectID).WithField("id", document.ID)

	updates := make(map[string]interface{}, 0)
	updates["synced"] = true
	if syncID != "" {
		updates["sync_id"] = syncID
	}

	if userID != "" {
		updates["user_id"] = userID
	}

	db := C.GetServices().Db
	err := db.Model(&SalesforceDocument{}).Where("project_id = ? AND id = ? AND timestamp = ? AND type = ? AND action = ?",
		projectID, document.ID, document.Timestamp, document.Type, document.Action).Updates(updates).Error
	if err != nil {
		logCtx.WithError(err).Error("Failed to update salesforce document as synced.")
		return http.StatusInternalServerError
	}

	return http.StatusAccepted
}

// SalesforceRecord is map for fields and their values
type SalesforceRecord map[string]interface{}

// BuildAndUpsertDocument creates new salesforce_document for insertion
func BuildAndUpsertDocument(projectID uint64, objectName string, value SalesforceRecord) error {
	if projectID == 0 {
		return errors.New("invalid project id")
	}
	if objectName == "" || value == nil {
		return errors.New("invalid oject name or value")
	}

	if len(value) == 0 {
		return errors.New("empty value")
	}

	var document SalesforceDocument
	document.ProjectID = projectID
	document.TypeAlias = objectName
	enValue, err := json.Marshal(value)
	if err != nil {
		return err
	}

	document.Value = &postgres.Jsonb{RawMessage: json.RawMessage(enValue)}
	status := CreateSalesforceDocument(projectID, &document)
	if status != http.StatusCreated && status != http.StatusConflict {
		return fmt.Errorf("error while creating document Status %d", status)
	}

	return nil
}
