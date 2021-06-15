package model

import (
	"encoding/json"
	"errors"
	"factors/util"
	U "factors/util"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

type HubspotDocument struct {
	ProjectId uint64 `gorm:"primary_key:true;auto_increment:false" json:"project_id"`
	ID        string `gorm:"primary_key:true;auto_increment:false" json:"id"`
	Type      int    `gorm:"primary_key:true;auto_increment:false" json:"type"`
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

// HubspotLastSyncInfo doc type last sync info
type HubspotLastSyncInfo struct {
	ProjectID uint64 `json:"-"`
	Type      int    `json:"type"`
	TypeAlias string `json:"type_alias"`
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

// Hubspot errors
var (
	ErrorHubspotUsingFallbackKey                 = errors.New("using fallback key from document")
	ErrorHubspotInvalidHubspotDocumentType       = errors.New("invalid document type")
	errorFailedToGetCreatedAtFromHubspotDocument = errors.New("failed to get created_at from document")
	errorFailedToGetUpdatedAtFromHubspotDocument = errors.New("failed to get updated_at from document")
)

// HubspotProperty only holds the value for hubspot document properties
type HubspotProperty struct {
	Value string `json:"value"`
}

// HubspotDocumentProperties only holds the properties object of the document
type HubspotDocumentProperties struct {
	Properties map[string]HubspotProperty `json:"properties"`
}

// HubspotProjectSyncStatus hubspot project sync status
type HubspotProjectSyncStatus struct {
	ProjectID uint64 `json:"project_id"`
	DocType   string `json:"doc_type"`
	Status    string `json:"status"`
	SyncAll   bool   `json:"sync_all"`
	Timestamp int64  `json:"timestamp"`
}

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

// HubspotDocumentTypeAlias hubspot document alias and their type
var HubspotDocumentTypeAlias = map[string]int{
	HubspotDocumentTypeNameCompany:        HubspotDocumentTypeCompany,
	HubspotDocumentTypeNameContact:        HubspotDocumentTypeContact,
	HubspotDocumentTypeNameDeal:           HubspotDocumentTypeDeal,
	HubspotDocumentTypeNameForm:           HubspotDocumentTypeForm,
	HubspotDocumentTypeNameFormSubmission: HubspotDocumentTypeFormSubmission,
}

// GetHubspotTypeByAlias gets document type by document alias
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

// GetHubspotTypeAliasByType get hubspot document type name by document type
func GetHubspotTypeAliasByType(typ int) string {
	for a, t := range HubspotDocumentTypeAlias {
		if typ == t {
			return a
		}
	}

	return ""
}

// GetHubspotAllowedObjects returns hubspot objects for api
func GetHubspotAllowedObjects(projectID uint64) *map[string]string {
	if projectID == 0 {
		return nil
	}

	return &hubspotObjectType
}

// GetHubspotObjectTypeByDocumentType get hubspot matching queriable object by document type
func GetHubspotObjectTypeByDocumentType(docType string) string {
	if docType == "" {
		return ""
	}

	if objectType, exist := hubspotObjectType[docType]; exist {
		return objectType
	}

	return ""
}

func getTimestampFromPropertiesByKey(propertiesMap map[string]interface{}, key string) (int64, error) {
	propertyValue, exists := propertiesMap[key]
	if !exists || propertyValue == nil {
		return 0, errors.New("failed to get timestamp from property key")
	}

	propertyValueMap := propertyValue.(map[string]interface{})
	timestampValue, exists := propertyValueMap["timestamp"]
	if !exists || timestampValue == nil {
		return 0, errors.New("timestamp key not exist on property map")
	}

	timestamp, err := ReadHubspotTimestamp(timestampValue)
	if err != nil || timestamp == 0 {
		return 0, errors.New("failed to read hubspot timestamp value")
	}

	return timestamp, nil
}

// GetHubspotDocumentUpdatedTimestamp get last updated timestamp from hubspot document
func GetHubspotDocumentUpdatedTimestamp(document *HubspotDocument) (int64, error) {
	if document.Type == 0 {
		return 0, ErrorHubspotInvalidHubspotDocumentType
	}

	value, err := U.DecodePostgresJsonb(document.Value)
	if err != nil {
		return 0, err
	}

	// property nested value.
	var propertyUpdateAtKey string
	if document.Type == HubspotDocumentTypeCompany ||
		document.Type == HubspotDocumentTypeDeal {

		propertyUpdateAtKey = "hs_lastmodifieddate"
	} else if document.Type == HubspotDocumentTypeContact {
		propertyUpdateAtKey = "lastmodifieddate"
	}
	if propertyUpdateAtKey != "" {
		properties, exists := (*value)["properties"]
		if !exists || properties == nil {
			return 0, errorFailedToGetUpdatedAtFromHubspotDocument
		}
		propertiesMap := properties.(map[string]interface{})

		propertyUpdateAt, exists := propertiesMap[propertyUpdateAtKey]
		if !exists || propertyUpdateAt == nil {
			return 0, errorFailedToGetUpdatedAtFromHubspotDocument
		}

		propertyUpdateAtMap := propertyUpdateAt.(map[string]interface{})
		value, exists := propertyUpdateAtMap["value"]
		if !exists || value == nil {
			return 0, errorFailedToGetUpdatedAtFromHubspotDocument
		}

		updatedAt, err := ReadHubspotTimestamp(value)
		if err != nil || updatedAt == 0 {
			return 0, errorFailedToGetUpdatedAtFromHubspotDocument
		}

		return updatedAt, nil
	}

	// direct values.
	var updatedAtKey string
	if document.Type == HubspotDocumentTypeForm {
		updatedAtKey = "updatedAt"
	} else if document.Type == HubspotDocumentTypeFormSubmission {
		updatedAtKey = "submittedAt"
	}
	if updatedAtKey != "" {
		updatedAtInt, exists := (*value)[updatedAtKey]
		if !exists || updatedAtInt == nil {
			return 0, errorFailedToGetUpdatedAtFromHubspotDocument
		}

		updatedAt, err := ReadHubspotTimestamp(updatedAtInt)
		if err != nil || updatedAt == 0 {
			return 0, errorFailedToGetUpdatedAtFromHubspotDocument
		}

		return updatedAt, nil
	}

	return 0, errorFailedToGetUpdatedAtFromHubspotDocument
}

// GetHubspotDocumentCreatedTimestamp get "created" timestamp of a hubspot document
func GetHubspotDocumentCreatedTimestamp(document *HubspotDocument) (int64, error) {
	if document.Type == 0 {
		return 0, ErrorHubspotInvalidHubspotDocumentType
	}

	value, err := U.DecodePostgresJsonb(document.Value)
	if err != nil {
		return 0, err
	}

	if document.Type == HubspotDocumentTypeCompany {
		properties, exists := (*value)["properties"]
		if !exists || properties == nil {
			return 0, errorFailedToGetCreatedAtFromHubspotDocument
		}
		propertiesMap := properties.(map[string]interface{})

		createdAt, err := getTimestampFromPropertiesByKey(propertiesMap, "name")
		if err != nil {
			createdAt, err = getTimestampFromPropertiesByKey(propertiesMap, "createdate")
			if err != nil {
				return 0, err
			}
		}

		return createdAt, nil
	}

	if document.Type == HubspotDocumentTypeDeal {
		properties, exists := (*value)["properties"]
		if !exists || properties == nil {
			return 0, errorFailedToGetCreatedAtFromHubspotDocument
		}
		propertiesMap := properties.(map[string]interface{})

		hsCreateDate, exists := propertiesMap["hs_createdate"]
		if exists {
			hsCreateDateMap := hsCreateDate.(map[string]interface{})
			createdAtValue, exists := hsCreateDateMap["value"]
			if exists && createdAtValue != nil {
				createdAt, err := ReadHubspotTimestamp(createdAtValue)
				if err != nil || createdAt == 0 {
					return 0, errorFailedToGetCreatedAtFromHubspotDocument
				}

				return createdAt, nil
			}
		}

		createdAt, err := getTimestampFromPropertiesByKey(propertiesMap, "dealname")
		if err != nil {
			return 0, err
		}

		return createdAt, nil
	}

	var createdAtKey string
	if document.Type == HubspotDocumentTypeContact {
		createdAtKey = "createdate"
	} else if document.Type == HubspotDocumentTypeForm {
		createdAtKey = "createdAt"
	} else if document.Type == HubspotDocumentTypeFormSubmission {
		createdAtKey = "submittedAt"
	} else {
		return 0, errorFailedToGetCreatedAtFromHubspotDocument
	}

	if document.Type == HubspotDocumentTypeContact {

		var hubspotDocumentProperties HubspotDocumentProperties
		err := json.Unmarshal(document.Value.RawMessage, &hubspotDocumentProperties)
		if err != nil {
			return 0, errors.New("Failed to unmarshal document properties")
		}

		createdAtStr, exist := hubspotDocumentProperties.Properties[createdAtKey]
		if exist {
			createdAt, err := ReadHubspotTimestamp(createdAtStr.Value)
			if err != nil {
				return 0, err
			}
			return createdAt, nil
		}

		createdAtKey = "addedAt"
		createdAtInt, exists := (*value)[createdAtKey]
		if !exists {
			return 0, errorFailedToGetCreatedAtFromHubspotDocument
		}

		createdAt, err := ReadHubspotTimestamp(createdAtInt)
		if err != nil || createdAt == 0 {
			log.WithFields(log.Fields{"document_id": document.ID}).Warn("Failed to read addedAt from contact document.")
			return 0, errorFailedToGetCreatedAtFromHubspotDocument
		}

		return createdAt, ErrorHubspotUsingFallbackKey
	}

	createdAtInt, exists := (*value)[createdAtKey]
	if !exists || createdAtInt == nil {
		return 0, errorFailedToGetCreatedAtFromHubspotDocument
	}

	createdAt, err := ReadHubspotTimestamp(createdAtInt)
	if err != nil || createdAt == 0 {
		return 0, errorFailedToGetCreatedAtFromHubspotDocument
	}

	return createdAt, nil
}

// HubspotIntegrationAccount account specific data for the hubspot api key
type HubspotIntegrationAccount struct {
	PortalID  int    `json:"portalId"`
	TimeZone  string `json:"timeZone"`
	Currency  string `json:"currency"`
	UtcOffset string `json:"utcOffset"`
}

// GetHubspotIntegrationAccount gets hubspot integration account using access token
func GetHubspotIntegrationAccount(apiKey string) (*HubspotIntegrationAccount, error) {
	var hubspotIntegrationAccount HubspotIntegrationAccount

	if apiKey == "" {
		return &hubspotIntegrationAccount, errors.New("missing hubspot api key")
	}

	url := "https://api.hubapi.com/integrations/v1/me?hapikey=" + apiKey

	resp, err := http.Get(url)
	if err != nil {
		return &hubspotIntegrationAccount, err
	}

	if resp.StatusCode != http.StatusOK {
		return &hubspotIntegrationAccount, fmt.Errorf("error getting integration account using hubspot access token")
	}

	err = json.NewDecoder(resp.Body).Decode(&hubspotIntegrationAccount)
	if err != nil {
		return &hubspotIntegrationAccount, err
	}
	return &hubspotIntegrationAccount, nil
}

// GetHubspotDecodedSyncInfo decode sync info from project settings to map
func GetHubspotDecodedSyncInfo(syncInfo *postgres.Jsonb) (*map[string]int64, error) {
	syncInfoMap := make(map[string]int64)
	if syncInfo == nil {
		return &syncInfoMap, nil
	}

	err := json.Unmarshal(syncInfo.RawMessage, &syncInfoMap)
	if err != nil {
		return nil, err
	}

	return &syncInfoMap, nil
}

// GetHubspotProjectOverAllStatus return  list of success projects and last successfull timestamp per document type
func GetHubspotProjectOverAllStatus(success []HubspotProjectSyncStatus,
	failure []HubspotProjectSyncStatus) (map[uint64]map[string]int64, map[uint64]bool) {

	status := make(map[uint64]bool)
	syncStatus := make(map[uint64]map[string]int64)
	for i := range success {
		status[success[i].ProjectID] = true

		if _, exist := syncStatus[success[i].ProjectID]; !exist {
			syncStatus[success[i].ProjectID] = make(map[string]int64)
		}
		syncStatus[success[i].ProjectID][success[i].DocType] = success[i].Timestamp
	}

	for i := range failure {
		status[success[i].ProjectID] = false
		log.WithFields(log.Fields{"project_id": failure[i].ProjectID, "doc_type": failure[i].DocType}).
			Error("Failed to complete hubspot first time sync.")
	}

	return syncStatus, status
}

// GetHubspotSyncUpdatedInfo return merged sync info
func GetHubspotSyncUpdatedInfo(incomingSyncInfo, existingSyncInfo *map[string]int64) *map[string]int64 {
	mergedSyncInfo := make(map[string]int64)
	for docType, timestamp := range *incomingSyncInfo {
		var hubspotLastSyncInfo HubspotLastSyncInfo
		hubspotLastSyncInfo.TypeAlias = docType
		if existingSyncInfo != nil {
			if _, exist := (*existingSyncInfo)[docType]; !exist {
				mergedSyncInfo[docType] = 0
			}

			if timestamp < (*existingSyncInfo)[docType] {
				mergedSyncInfo[docType] = (*existingSyncInfo)[docType]
			} else {
				mergedSyncInfo[docType] = timestamp
			}
		}
	}

	return &mergedSyncInfo
}
