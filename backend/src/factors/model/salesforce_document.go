package model

import (
	"errors"
	U "factors/util"
	"net/http"
	"time"

	C "factors/config"

	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

type SalesforceDocument struct {
	ProjectId uint64           `gorm:"primary_key:true;auto_increment:false" json:"project_id"`
	ID        string           `gorm:"primary_key:true;auto_increment:false" json:"id"`
	Type      int              `gorm:"primary_key:true;auto_increment:false" json:"-"`
	Action    SalesforceAction `gorm:"primary_key:true;auto_increment:false" json:"action"`
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

	var timestamp int64
	if isNew {
		document.Action = SalesforceDocumentCreated // created
		timestamp, err = getSalesforceDocumentTimestampByAction(document, SalesforceDocumentCreated)
	} else {
		document.Action = SalesforceDocumentUpdated // updated
		timestamp, err = getSalesforceDocumentTimestampByAction(document, SalesforceDocumentUpdated)
	}

	if err != nil {
		logCtx.WithField("action", document.Action).WithError(err).Error(
			"Failed to get timestamp from salesforce document on create.")
		return http.StatusInternalServerError
	}

	db := C.GetServices().Db
	err = db.Create(document).Error
	if err != nil {
		//check duplicate recode error

		logCtx.WithError(err).Error("Failed to create salesforce document.")
		return http.StatusInternalServerError
	}

	return http.StatusCreated
}

func getSalesforceDocumentTimestampByAction(document *SalesforceDocument, action SalesforceAction) (int64, error) {
	if document.Type == 0 {
		return 0, errors.New("invalid document type")
	}

	value, err := U.DecodePostgresJsonb(document.Value)
	if err != nil {
		return 0, err
	}

	var dateKey string
	if action == SalesforceDocumentCreated {
		dateKey = "CreatedDate"
	}
	if action == SalesforceDocumentUpdated {
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
	if timestampStr, ok := timestamp.(string); ok {
		if t, err := time.Parse(SalesforceDocumentTimeLayout, timestampStr); err != nil {
			return t.Unix(), nil
		}
	}

	return 0, errors.New("invalid timestamp")
}
