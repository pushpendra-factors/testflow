package model

import (
	"errors"
	"factors/util"
	"fmt"
	"strings"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
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
	UserID    string           `gorm:"default:null" json:"user_id"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
}

type SalesforceAction int

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

// SalesforceRecord is map for fields and their values
type SalesforceRecord map[string]interface{}

func GetCRMEnrichPropertyKeyByType(source, typ, key string) string {
	return util.NAME_PREFIX + getCRMPropertyKeyByType(source, typ, key)
}

func getCRMPropertyKeyByType(source, objectType, key string) string {
	return fmt.Sprintf("%s_%s_%s", source, objectType, strings.ToLower(key))
}

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

func GetSalesforceLastModifiedTimestamp(document *SalesforceDocument) (int64, error) {
	if document.Type == 0 {
		return 0, errors.New("invalid document type")
	}

	value, err := util.DecodePostgresJsonb(document.Value)
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
func GetSalesforceDocumentTimestampByAction(document *SalesforceDocument,
	action SalesforceAction) (int64, error) {

	if document.Type == 0 {
		return 0, errors.New("invalid document type")
	}
	if action == 0 {
		return 0, errors.New("invalid action")
	}

	value, err := util.DecodePostgresJsonb(document.Value)
	if err != nil {
		return 0, err
	}

	if action == SalesforceDocumentUpdated {
		return GetSalesforceLastModifiedTimestamp(document)
	}

	dateKey := "CreatedDate"

	date, exists := (*value)[dateKey]
	if !exists || date == nil {
		return 0, errors.New("failed to get date")
	}

	return GetSalesforceDocumentTimestamp(date)
}

// GetSalesforceDocumentTimestamp return unix timestamp for salesforce formated timestamp
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
