package model

import (
	"errors"
	"factors/util"
	"strconv"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
)

type HubspotDocument struct {
	ProjectId uint64 `gorm:"primary_key:true;auto_increment:false" json:"project_id"`
	ID        string `gorm:"primary_key:true;auto_increment:false" json:"id"`
	Type      int    `gorm:"primary_key:true;auto_increment:false" json:"-"`
	Action    int    `gorm:"primary_key:true;auto_increment:false" json:"action"`
	// created or updated timestamp from hubspot.
	Timestamp int64           `gorm:"primary_key:true;auto_increment:false" json:"timestamp"`
	TypeAlias string          `gorm:"-" json:"type_alias"`
	Value     *postgres.Jsonb `json:"value"`
	Synced    bool            `gorm:"default:false;not null" json:"synced"`
	SyncId    string          `gorm:"default:null" json:"sync_id"`
	UserId    string          `gorm:"default:null" json:"user_id"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

type HubspotLastSyncInfo struct {
	ProjectId uint64 `json:"-"`
	Type      int    `json:"type"`
	Timestamp int64  `json:"timestamp"`
}

type HubspotSyncInfo struct {
	ProjectSettings map[uint64]*HubspotProjectSettings `json:"project_settings"`
	// project_id: { type: last_sync_info }
	LastSyncInfo map[uint64]map[string]int64 `json:"last_sync_info"`
}

const (
	HubspotDocumentActionCreated = 1
	HubspotDocumentActionUpdated = 2
)

const (
	HubspotDocumentTypeCompany            = 1
	HubspotDocumentTypeNameCompany        = "company"
	HubspotDocumentTypeContact            = 2
	HubspotDocumentTypeNameContact        = "contact"
	HubspotDocumentTypeDeal               = 3
	HubspotDocumentTypeNameDeal           = "deal"
	HubspotDocumentTypeForm               = 4
	HubspotDocumentTypeNameForm           = "form"
	HubspotDocumentTypeFormSubmission     = 5
	HubspotDocumentTypeNameFormSubmission = "form_submission"
)

var (
	hubspotDataTypeDatetime = map[string]bool{
		"datetime": true,
		"date":     true,
	}

	hubspotDataTypeNumerical = map[string]bool{
		"number": true,
	}

	hubspotObjectType = map[string]string{
		HubspotDocumentTypeNameCompany: "companies",
		HubspotDocumentTypeNameContact: "contacts",
		HubspotDocumentTypeNameDeal:    "deals",
	}
)

// GetHubspotMappedDataType returns mapped factors data type
func GetHubspotMappedDataType(dataType string) string {
	if dataType == "" {
		return ""
	}

	if _, exists := hubspotDataTypeDatetime[dataType]; exists {
		return util.PropertyTypeDateTime
	}

	if _, exists := hubspotDataTypeNumerical[dataType]; exists {
		return util.PropertyTypeNumerical
	}

	return util.PropertyTypeUnknown
}

// ReadHubspotTimestamp returns timestamp in int64 format. Warning - documents use milliseconds
func ReadHubspotTimestamp(value interface{}) (int64, error) {
	switch value.(type) {
	case float64:
		return int64(uint64(value.(float64))), nil
	case string:
		timestamp, err := strconv.ParseInt(value.(string), 10, 64)
		if err != nil {
			return 0, err
		}
		return timestamp, nil
	}

	return 0, errors.New("unsupported hubspot timestamp type")
}

var HubspotDocumentTypeAlias = map[string]int{
	HubspotDocumentTypeNameCompany:        HubspotDocumentTypeCompany,
	HubspotDocumentTypeNameContact:        HubspotDocumentTypeContact,
	HubspotDocumentTypeNameDeal:           HubspotDocumentTypeDeal,
	HubspotDocumentTypeNameForm:           HubspotDocumentTypeForm,
	HubspotDocumentTypeNameFormSubmission: HubspotDocumentTypeFormSubmission,
}

func GetHubspotTypeByAlias(alias string) (int, error) {
	if alias == "" {
		return 0, errors.New("empty document type alias")
	}

	typ, typExists := HubspotDocumentTypeAlias[alias]
	if !typExists {
		return 0, errors.New("invalid document type alias")
	}

	return typ, nil
}

func GetHubspotTypeAliasByType(typ int) string {
	for a, t := range HubspotDocumentTypeAlias {
		if typ == t {
			return a
		}
	}

	return ""
}

func GetHubspotAllowedObjects(projectID uint64) *map[string]string {
	if projectID == 0 {
		return nil
	}

	return &hubspotObjectType
}

func GetHubspotObjectTypeByDocumentType(docType string) string {
	if docType == "" {
		return ""
	}

	if objectType, exist := hubspotObjectType[docType]; exist {
		return objectType
	}

	return ""
}
