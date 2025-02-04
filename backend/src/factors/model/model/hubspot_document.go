package model

import (
	"bytes"
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
	ProjectId int64  `gorm:"primary_key:true;auto_increment:false" json:"project_id"`
	ID        string `gorm:"primary_key:true;auto_increment:false" json:"id"`
	Type      int    `gorm:"primary_key:true;auto_increment:false" json:"type"`
	Action    int    `gorm:"primary_key:true;auto_increment:false" json:"action"`
	// created or updated timestamp from hubspot.
	Timestamp   int64           `gorm:"primary_key:true;auto_increment:false" json:"timestamp"`
	TypeAlias   string          `gorm:"-" json:"type_alias"`
	Value       *postgres.Jsonb `json:"value"`
	Synced      bool            `gorm:"default:false;not null" json:"synced"`
	SyncId      string          `gorm:"default:null" json:"sync_id"`
	UserId      string          `gorm:"default:null" json:"user_id"`
	GroupUserId string          `gorm:"default:null" json:"group_user_id"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	SyncTries   int             `gorm:"default:0" json:"sync_tries"`

	// for internal use only
	timeZone       U.TimeZoneString `gorm:"-" json:"-"`
	dateProperties *map[string]bool `gorm:"-" json:"-"`
}

// HubspotLastSyncInfo doc type last sync info
type HubspotLastSyncInfo struct {
	ProjectID int64  `json:"-"`
	Type      int    `json:"type"`
	TypeAlias string `json:"type_alias"`
	Timestamp int64  `json:"timestamp"`
}

type HubspotSyncInfo struct {
	ProjectSettings map[int64]*HubspotProjectSettings `json:"project_settings"`
	// project_id: { type: last_sync_info }
	LastSyncInfo map[int64]map[string]int64 `json:"last_sync_info"`
}

const (
	HubspotDocumentActionCreated             = 1
	HubspotDocumentActionUpdated             = 2
	HubspotDocumentActionDeleted             = 3
	HubspotDocumentActionAssociationsUpdated = 4
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
	HubspotDocumentTypeEngagement         = 6
	HubspotDocumentTypeNameEngagement     = "engagement"
	HubspotDocumentTypeContactList        = 7
	HubspotDocumentTypeNameContactList    = "contact_list"
	HubspotDocumentTypeOwner              = 8
	HubspotDocumentTypeNameOwner          = "owner"

	HubspotDateTimeLayout                    = "2006-01-02T15:04:05.000Z"
	HubspotDateTimeWithoutMilliSecondsLayout = "2006-01-02T15:04:05Z"
	HubspotDateLayout                        = "2006-01-02"
	HubspotDataTypeDate                      = "date"
	HubspotDataTypeDatetime                  = "datetime"
)

var (
	hubspotDataTypeDatetime = map[string]bool{
		HubspotDataTypeDatetime: true,
		HubspotDataTypeDate:     true,
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

var HubspotAlowedObjectsByPlan = map[string][]string{
	FEATURE_HUBSPOT_BASIC: {HubspotDocumentTypeNameCompany, HubspotDocumentTypeNameDeal, HubspotDocumentTypeNameOwner},
	FEATURE_HUBSPOT: {HubspotDocumentTypeNameCompany, HubspotDocumentTypeNameDeal, HubspotDocumentTypeNameContact, HubspotDocumentTypeNameEngagement, HubspotDocumentTypeNameForm,
		HubspotDocumentTypeNameFormSubmission, HubspotDocumentTypeNameOwner},
}

var HubspotEventNameToDocTypeMapping = map[string]string{
	U.EVENT_NAME_HUBSPOT_CONTACT_CREATED:            HubspotDocumentTypeNameContact,
	U.EVENT_NAME_HUBSPOT_CONTACT_UPDATED:            HubspotDocumentTypeNameContact,
	U.EVENT_NAME_HUBSPOT_CONTACT_FORM_SUBMISSION:    HubspotDocumentTypeNameContact,
	U.GROUP_EVENT_NAME_HUBSPOT_COMPANY_CREATED:      HubspotDocumentTypeNameCompany,
	U.GROUP_EVENT_NAME_HUBSPOT_COMPANY_UPDATED:      HubspotDocumentTypeNameCompany,
	U.GROUP_EVENT_NAME_HUBSPOT_DEAL_CREATED:         HubspotDocumentTypeNameDeal,
	U.GROUP_EVENT_NAME_HUBSPOT_DEAL_UPDATED:         HubspotDocumentTypeNameDeal,
	U.EVENT_NAME_HUBSPOT_DEAL_STATE_CHANGED:         HubspotDocumentTypeNameDeal,
	U.EVENT_NAME_HUBSPOT_ENGAGEMENT_MEETING_CREATED: HubspotDocumentTypeNameEngagement,
	U.EVENT_NAME_HUBSPOT_ENGAGEMENT_MEETING_UPDATED: HubspotDocumentTypeNameEngagement,
	U.EVENT_NAME_HUBSPOT_ENGAGEMENT_EMAIL:           HubspotDocumentTypeNameEngagement,
	U.EVENT_NAME_HUBSPOT_ENGAGEMENT_CALL_CREATED:    HubspotDocumentTypeNameEngagement,
	U.EVENT_NAME_HUBSPOT_ENGAGEMENT_CALL_UPDATED:    HubspotDocumentTypeNameEngagement,
}

var HubspotGroupNameToDocTypeMapping = map[string]string{
	U.GROUP_NAME_HUBSPOT_COMPANY: HubspotDocumentTypeNameCompany,
	U.GROUP_NAME_HUBSPOT_DEAL:    HubspotDocumentTypeNameDeal,
}

// Hubspot errors
var (
	ErrorHubspotUsingFallbackKey                        = errors.New("using fallback key from document")
	ErrorHubspotInvalidHubspotDocumentType              = errors.New("invalid document type")
	errorFailedToGetCreatedAtFromHubspotDocument        = errors.New("failed to get created_at from document")
	errorFailedToGetUpdatedAtFromHubspotDocument        = errors.New("failed to get updated_at from document")
	errorFailedToGetPropertiesFromHubspotDocument       = errors.New("failed to get properties from document")
	errorFailedToGetLastModifiedDateFromHubspotDocument = errors.New("failed to get results from document")
)

// HubspotProperty only holds the value for hubspot document properties
type HubspotProperty struct {
	Value string `json:"value"`
}

// HubspotDocumentProperties only holds the properties object of the document
type HubspotDocumentProperties struct {
	Properties map[string]HubspotProperty `json:"properties"`
}

// HubspotDocumentProperties only holds the properties object of the document
type HubspotDocumentPropertiesV3 struct {
	Properties map[string]interface{} `json:"properties"`
}

// HubspotProjectSyncStatus hubspot project sync status
type HubspotProjectSyncStatus struct {
	ProjectID int64  `json:"project_id"`
	DocType   string `json:"doc_type"`
	Status    string `json:"status"`
	SyncAll   bool   `json:"sync_all"`
	Timestamp int64  `json:"timestamp"`
}

// Associations struct for deal associations
type Associations struct {
	AssociatedCompanyIds []int64 `json:"associatedCompanyIds"`
}

// Deal definition
type Deal struct {
	Associations Associations `json:"associations"`
}

type HubspotDocumentCount struct {
	ProjectID int64
	Count     int
}

func (document *HubspotDocument) SetTimeZone(timeZone U.TimeZoneString) {
	document.timeZone = timeZone
}

func (document *HubspotDocument) GetTimeZone() U.TimeZoneString {
	return document.timeZone
}

func (document *HubspotDocument) SetDateProperties(dateProperties *map[string]bool) {
	document.dateProperties = dateProperties
}

func (document *HubspotDocument) GetDateProperties() *map[string]bool {
	return document.dateProperties
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

func GetHubspotEngagementId(documentMap map[string]interface{}, idKey string) (string, error) {
	if _, exists := documentMap["id"]; exists { // Engagement V3
		return U.GetPropertyValueAsString(documentMap["id"]), nil
	}

	engagementInterface, engagementExists := documentMap["engagement"]
	if !engagementExists {
		return "", errors.New("engagement not found on results document type")
	}
	engagementMap, isConverted := engagementInterface.(map[string]interface{})
	if !isConverted {
		log.Error("interface has not converted to map")
	}
	engagementMapvalueInFloat, ok := U.GetPropertyValueAsFloat64(engagementMap[idKey])
	if ok != nil {
		return "", errors.New("failed to convert interface into float64")
	}
	return fmt.Sprintf("%v", (int64)(engagementMapvalueInFloat)), nil
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
	HubspotDocumentTypeNameEngagement:     HubspotDocumentTypeEngagement,
	HubspotDocumentTypeNameContactList:    HubspotDocumentTypeContactList,
	HubspotDocumentTypeNameOwner:          HubspotDocumentTypeOwner,
}

// HubspotDocumentActionAlias hubspot document alias and their action
var HubspotDocumentActionAlias = map[int]string{
	HubspotDocumentActionCreated:             "HubspotDocumentActionCreated",
	HubspotDocumentActionUpdated:             "HubspotDocumentActionUpdated",
	HubspotDocumentActionDeleted:             "HubspotDocumentActionDeleted",
	HubspotDocumentActionAssociationsUpdated: "HubspotDocumentActionAssociationsUpdated",
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

// GetCRMTimeSeriesByStartTimestamp returns time series for batch processing -> {Day1,Day2}, {Day2,Day3},{Day3,Day4} upto current day
func GetCRMTimeSeriesByStartTimestamp(projectID int64, from int64, CRMEventSource string) [][]int64 {
	logCtx := log.WithFields(log.Fields{"project_id": projectID, "from": from, "crm_source": CRMEventSource})
	if from < 1 {
		logCtx.Error("Invalid timestamp from batch processing by day.")
		return nil
	}

	if CRMEventSource != SmartCRMEventSourceSalesforce && CRMEventSource != SmartCRMEventSourceHubspot &&
		CRMEventSource != U.CRM_SOURCE_NAME_MARKETO {
		logCtx.Error("Invalid source.")
		return nil
	}

	multiplier := int64(1)
	if CRMEventSource == SmartCRMEventSourceHubspot {
		multiplier = 1000
	}

	timeSeries := [][]int64{}
	startTime := time.Unix(from/multiplier, 0)
	startDate := time.Date(startTime.UTC().Year(), startTime.UTC().Month(), startTime.UTC().Day(), 0, 0, 0, 0, time.UTC)
	currentTime := time.Now()
	for ; startDate.Unix() < currentTime.Unix(); startDate = startDate.AddDate(0, 0, 1) {
		timeSeries = append(timeSeries, []int64{startTime.Unix() * multiplier, startDate.AddDate(0, 0, 1).Unix() * multiplier})
		startTime = startDate.AddDate(0, 0, 1)
	}

	return timeSeries
}

// GetHubspotAllowedObjects returns hubspot objects for api
func GetHubspotAllowedObjects(projectID int64) *map[string]string {
	if projectID == 0 {
		return nil
	}

	return &hubspotObjectType
}

// GetHubspotObjectTypeByDocumentType get hubspot matching queryable object by document type
func GetHubspotObjectTypeByDocumentType(docType string) string {
	if docType == "" {
		return ""
	}

	if objectType, exist := hubspotObjectType[docType]; exist {
		return objectType
	}

	return ""
}

func GetTimestampForV3Records(propertyValue interface{}) (int64, error) {
	tm, err := time.Parse(HubspotDateTimeLayout, U.GetPropertyValueAsString(propertyValue))
	if err == nil {
		return tm.UnixNano() / int64(time.Millisecond), nil
	}

	tm, err = time.Parse(HubspotDateTimeWithoutMilliSecondsLayout, U.GetPropertyValueAsString(propertyValue))
	if err == nil {
		return tm.UnixNano() / int64(time.Millisecond), nil
	}

	tm, err = time.Parse(HubspotDateLayout, U.GetPropertyValueAsString(propertyValue))
	if err == nil {
		return tm.UnixNano() / int64(time.Millisecond), nil
	}
	return 0, errors.New("failed to convert timestamp inside getTimestampFromPropertiesByKeyV3")
}

func getTimestampFromPropertiesByKey(propertiesMap map[string]interface{}, key string) (int64, error) {
	propertyValue, exists := propertiesMap[key]
	if !exists || propertyValue == nil {
		return 0, errors.New("failed to get timestamp from property key")
	}

	propertyValueMap, ok := propertyValue.(map[string]interface{})
	if !ok {
		return GetTimestampForV3Records(propertyValue)
	}

	timestamp, err := ReadHubspotTimestamp(propertyValueMap["value"])
	if err == nil {
		return timestamp, nil
	}

	timestampValue, exists := propertyValueMap["timestamp"]
	if !exists || timestampValue == nil {
		return 0, errors.New("timestamp key not exist on property map")
	}

	timestamp, err = ReadHubspotTimestamp(timestampValue)
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

	if document.Action == HubspotDocumentActionDeleted {
		return GetHubspotDocumentLastModifiedDate(document)
	}

	value, err := U.DecodePostgresJsonb(document.Value)
	if err != nil {
		return 0, err
	}

	if document.Type == HubspotDocumentTypeEngagement {
		if engagementV3Interface, engagementV3Exists := (*value)["properties"]; engagementV3Exists {
			engagementV3Map, isConverted := engagementV3Interface.(map[string]interface{})
			if !isConverted {
				log.Error("interface has not converted to map")
			}

			value, exists := engagementV3Map["hs_lastmodifieddate"]
			if !exists || value == nil {
				return 0, errorFailedToGetUpdatedAtFromHubspotDocument
			}

			valueInInt64, ok := GetTimestampForV3Records(value)
			if ok != nil {
				return 0, errors.New("failed to convert interface into float64")
			}
			return valueInInt64, nil
		}

		engagementInterface, engagementExists := (*value)["engagement"]
		if !engagementExists {
			return 0, errors.New("engagement not found on results document type")
		}
		engagementMap, isConverted := engagementInterface.(map[string]interface{})
		if !isConverted {
			log.Error("interface has not converted to map")
		}
		value, exists := engagementMap["lastUpdated"]
		if !exists || value == nil {
			return 0, errorFailedToGetUpdatedAtFromHubspotDocument
		}
		valueInFloat, ok := U.GetPropertyValueAsFloat64(value)
		if ok != nil {
			return 0, errors.New("failed to convert interface into float64")
		}
		return int64(valueInFloat), nil
	}

	if document.Type == HubspotDocumentTypeContactList {
		timestampInt, exists := (*value)["contact_timestamp"]
		if !exists {
			return 0, errors.New("Failed to get contact_timestamp in GetHubspotDocumentUpdatedTimestamp")
		}

		timestamp, err := ReadHubspotTimestamp(timestampInt)
		if err != nil || timestamp == 0 {
			return 0, errors.New("Failed to read hubspot timestamp value in GetHubspotDocumentUpdatedTimestamp")
		}

		return timestamp, nil
	}

	// property nested value.
	var propertyUpdateAtKey string
	if document.Type == HubspotDocumentTypeCompany ||
		document.Type == HubspotDocumentTypeDeal {

		propertyUpdateAtKey = U.PROPERTY_KEY_LAST_MODIFIED_DATE_HS
	} else if document.Type == HubspotDocumentTypeContact {
		propertyUpdateAtKey = U.PROPERTY_KEY_LAST_MODIFIED_DATE
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

		propertyUpdateAtMap, ok := propertyUpdateAt.(map[string]interface{})
		if !ok {
			return GetTimestampForV3Records(propertyUpdateAt)
		}
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
	if document.Type == HubspotDocumentTypeForm || document.Type == HubspotDocumentTypeOwner {
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

	if document.Action == HubspotDocumentActionDeleted {
		return time.Now().UnixNano() / int64(time.Millisecond), nil
	}

	value, err := U.DecodePostgresJsonb(document.Value)
	if err != nil {
		return 0, err
	}

	if document.Type == HubspotDocumentTypeEngagement {
		if engagementV3Interface, engagementV3Exists := (*value)["properties"]; engagementV3Exists {
			engagementV3Map, isConverted := engagementV3Interface.(map[string]interface{})
			if !isConverted {
				log.Error("interface has not converted to map")
			}

			value, exists := engagementV3Map["hs_createdate"]
			if !exists || value == nil {
				return 0, errorFailedToGetCreatedAtFromHubspotDocument
			}

			valueInInt64, ok := GetTimestampForV3Records(value)
			if ok != nil {
				return 0, errors.New("failed to convert interface into float64 for engagement_V3")
			}
			return valueInInt64, nil
		}

		engagementInterface, engagementExists := (*value)["engagement"]
		if !engagementExists {
			return 0, errors.New("engagement not found on results document type")
		}

		engagementMap, isConverted := engagementInterface.(map[string]interface{})
		if !isConverted {
			log.Error("interface has not converted to map")
		}

		value, exists := engagementMap["createdAt"]
		if !exists || value == nil {
			return 0, errorFailedToGetCreatedAtFromHubspotDocument
		}
		valueInFloat, ok := U.GetPropertyValueAsFloat64(value)
		if ok != nil {
			return 0, errors.New("failed to convert interface into float64")
		}
		return int64(valueInFloat), nil
	}

	if document.Type == HubspotDocumentTypeCompany {
		properties, exists := (*value)["properties"]
		if !exists || properties == nil {
			return 0, errorFailedToGetCreatedAtFromHubspotDocument
		}
		propertiesMap := properties.(map[string]interface{})

		createdAt, err := getTimestampFromPropertiesByKey(propertiesMap, "createdate")
		if err != nil {
			createdAt, err = getTimestampFromPropertiesByKey(propertiesMap, "name")
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

		if _, isDealV3Record := (*value)["id"]; isDealV3Record {
			createDate, exists := propertiesMap["createdate"]
			if exists && createDate != nil && createDate != "" {
				createTimestamp, err := GetTimestampForV3Records(createDate)
				if err != nil || createTimestamp == 0 {
					return 0, err
				}
				return createTimestamp, nil
			}

			hsCreateDate, exists := propertiesMap["hs_createdate"]
			if exists && hsCreateDate != nil && hsCreateDate != "" {
				createTimestamp, err := GetTimestampForV3Records(hsCreateDate)
				if err != nil || createTimestamp == 0 {
					return 0, err
				}
				return createTimestamp, nil
			}

			log.WithField("document", *value).Error("failed to get created_date and hs_createdate for deal_v3")
			return 0, errorFailedToGetCreatedAtFromHubspotDocument
		}

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

	if document.Type == HubspotDocumentTypeContactList {
		timestampInt, exists := (*value)["contact_timestamp"]
		if !exists {
			return 0, errors.New("Failed to get contact_timestamp in GetHubspotDocumentCreatedTimestamp")
		}

		timestamp, err := ReadHubspotTimestamp(timestampInt)
		if err != nil || timestamp == 0 {
			return 0, errors.New("Failed to read hubspot timestamp value in GetHubspotDocumentCreatedTimestamp")
		}

		return timestamp, nil
	}

	var createdAtKey string
	if document.Type == HubspotDocumentTypeContact {
		createdAtKey = "createdate"
	} else if document.Type == HubspotDocumentTypeForm || document.Type == HubspotDocumentTypeOwner {
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
	PortalID              int              `json:"portalId"`
	TimeZone              U.TimeZoneString `json:"timeZone"`
	Currency              string           `json:"currency"`
	UTCOffsetMilliseconds int              `json:"utcOffsetMilliseconds"`
	UtcOffset             string           `json:"utcOffset"`
}

type HubspotOAuthUserCredentials struct {
	RefreshToken string `json:"refresh_token"`
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

func GetHubspotAccessToken(refreshToken, appID, appSecret string) (string, error) {
	url := "https://api.hubapi.com/oauth/v1/token?"
	parameters := fmt.Sprintf("grant_type=%s&client_id=%s&client_secret=%s&refresh_token=%s", "refresh_token", appID, appSecret, refreshToken)
	url = url + parameters

	resp, err := ActionHubspotRequestHandler("POST", url, "", "", "application/x-www-form-urlencoded;charset=utf-8", nil)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		var body interface{}
		json.NewDecoder(resp.Body).Decode(&body)
		log.WithFields(log.Fields{"response_body": body}).Error("Failed to get hubspot access token")
		return "", fmt.Errorf("failed to get hubspot access token")
	}

	var userCredentials HubspotOAuthUserCredentials
	err = json.NewDecoder(resp.Body).Decode(&userCredentials)
	if err != nil {
		return "", err
	}

	return userCredentials.AccessToken, nil
}

func ActionHubspotRequestHandler(method, url, apiKey, accessToken, contentType string, payload []byte) (*http.Response, error) {

	if apiKey != "" {
		url = url + "hapikey=" + apiKey
	}

	var req *http.Request
	var err error
	if payload != nil {
		body := bytes.NewBuffer(payload)
		req, err = http.NewRequest(method, url, body)
		if err != nil {
			return nil, err
		}
	} else {
		req, err = http.NewRequest(method, url, nil)
		if err != nil {
			return nil, err
		}
	}

	if accessToken != "" {
		req.Header["Authorization"] = []string{"Bearer " + accessToken}
	}

	if contentType != "" {
		req.Header["Content-Type"] = []string{contentType}
	}

	client := &http.Client{
		Timeout: 10 * time.Minute,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// GetHubspotIntegrationAccount gets hubspot integration account using api key or access token. appID and appSecret is used only in refresh token
func GetHubspotIntegrationAccount(projectID int64, apiKey, refreshToken, appID, appSecret string) (*HubspotIntegrationAccount, error) {
	logCtx := log.WithFields(log.Fields{"project_id": projectID})
	var hubspotIntegrationAccount HubspotIntegrationAccount

	if apiKey == "" && refreshToken == "" {
		logCtx.Error("Failed to get hubspot api key and refresh token")
		return &hubspotIntegrationAccount, errors.New("missing api key and refresh token")
	}

	url := "https://api.hubapi.com/integrations/v1/me?"

	var resp *http.Response
	var err error
	if refreshToken != "" {
		accessToken, err := GetHubspotAccessToken(refreshToken, appID, appSecret)
		if err != nil {
			logCtx.WithError(err).Error("Failed to get hubspot access token.")
			return &hubspotIntegrationAccount, err
		}

		resp, err = ActionHubspotRequestHandler("GET", url, "", accessToken, "", nil)
		if err != nil {
			logCtx.WithError(err).Error("Failed to request for hubspot account info using access token.")
			return &hubspotIntegrationAccount, err
		}

	} else {
		resp, err = ActionHubspotRequestHandler("GET", url, apiKey, "", "", nil)
		if err != nil {
			logCtx.WithError(err).Error("Failed to request for hubspot account info using api key.")
			return &hubspotIntegrationAccount, err
		}
	}

	if resp.StatusCode != http.StatusOK {
		var body interface{}
		json.NewDecoder(resp.Body).Decode(&body)
		logCtx.WithFields(log.Fields{"respone_body": body}).WithError(err).Error("Failed to get hubspot account info ")
		return &hubspotIntegrationAccount, fmt.Errorf("error getting integration account")
	}

	err = json.NewDecoder(resp.Body).Decode(&hubspotIntegrationAccount)
	if err != nil {
		logCtx.WithError(err).Error("Failed to decode hubspot account info.")
		return &hubspotIntegrationAccount, err
	}

	return &hubspotIntegrationAccount, nil
}

func GetHubspotAccountTimezoneAndPortalID(projectID int64, apiKey, refreshToken, appID, appSecret string) (U.TimeZoneString, string, error) {
	logCtx := log.WithFields(log.Fields{"project_id": projectID})
	account, err := GetHubspotIntegrationAccount(projectID, apiKey, refreshToken, appID, appSecret)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get hubspot account info for timezone.")
		return "", "", err
	}

	_, err = time.LoadLocation(string(account.TimeZone))
	if err != nil {
		logCtx.WithError(err).Error("Failed to load timezone from account.")
		return "", "", err
	}

	return account.TimeZone, U.GetPropertyValueAsString(account.PortalID), nil
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
	failure []HubspotProjectSyncStatus) (map[int64]map[string]int64, map[int64]bool) {

	status := make(map[int64]bool)
	syncStatus := make(map[int64]map[string]int64)
	for i := range success {
		status[success[i].ProjectID] = true

		if _, exist := syncStatus[success[i].ProjectID]; !exist {
			syncStatus[success[i].ProjectID] = make(map[string]int64)
		}
		syncStatus[success[i].ProjectID][success[i].DocType] = success[i].Timestamp
	}

	for i := range failure {
		status[failure[i].ProjectID] = false
		log.WithFields(log.Fields{"project_id": failure[i].ProjectID, "doc_type": failure[i].DocType}).
			Error("Failed to complete hubspot first time sync.")
	}

	return syncStatus, status
}

// GetHubspotSyncUpdatedInfo return merged sync info
func GetHubspotSyncUpdatedInfo(incomingSyncInfo, existingSyncInfo *map[string]int64) *map[string]int64 {
	mergedSyncInfo := make(map[string]int64)
	for docType, timestamp := range *existingSyncInfo {
		mergedSyncInfo[docType] = timestamp
	}

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

func GetHubspotDocumentLastModifiedDate(document *HubspotDocument) (int64, error) {
	logCtx := log.WithFields(log.Fields{"project_id": document.ProjectId})
	if document.Type == 0 {
		return 0, ErrorHubspotInvalidHubspotDocumentType
	}

	value, err := U.DecodePostgresJsonb(document.Value)
	if err != nil {
		return 0, err
	}

	properties, exists := (*value)["properties"]
	if !exists || properties == nil {
		return 0, errorFailedToGetPropertiesFromHubspotDocument
	}
	propertiesMap := properties.(map[string]interface{})
	lastmodifieddate, exists := (propertiesMap)[U.PROPERTY_KEY_LAST_MODIFIED_DATE]
	if !exists || lastmodifieddate == nil {
		return 0, errorFailedToGetLastModifiedDateFromHubspotDocument
	}
	tm, err := time.Parse(HubspotDateTimeLayout, U.GetPropertyValueAsString(lastmodifieddate))
	if err != nil {
		logCtx.WithField("action", document.Action).WithError(err).Error("Failed to convert timestamp inside GetHubspotDocumentLastModifiedDate")
		return time.Now().UnixNano() / int64(time.Millisecond), nil
	}
	return tm.UnixNano() / int64(time.Millisecond), nil
}

// IsDealUpdatedRequired checks if any update on associated company in deal.
func IsDealUpdatedRequired(incoming, existing *HubspotDocument) (bool, error) {
	if incoming.Type != HubspotDocumentTypeDeal || existing.Type != HubspotDocumentTypeDeal {
		return false, errors.New("invalid document type")
	}

	var incomingDeal Deal
	var existingMap Deal
	err := json.Unmarshal(incoming.Value.RawMessage, &incomingDeal)
	if err != nil {
		return false, err
	}

	err = json.Unmarshal(existing.Value.RawMessage, &existingMap)
	if err != nil {
		return false, err
	}

	existingCompanyIDs := make(map[interface{}]bool)
	for i := range existingMap.Associations.AssociatedCompanyIds {
		existingCompanyIDs[existingMap.Associations.AssociatedCompanyIds[i]] = true
	}

	for i := range incomingDeal.Associations.AssociatedCompanyIds {
		if !existingCompanyIDs[incomingDeal.Associations.AssociatedCompanyIds[i]] {
			return true, nil
		}
	}

	return false, nil
}

func GetCurrentGroupIdAndColumnName(user *User) (string, string) {
	if user.Group1ID != "" {
		return user.Group1ID, "group_1_id"
	}
	if user.Group2ID != "" {
		return user.Group2ID, "group_2_id"
	}
	if user.Group3ID != "" {
		return user.Group3ID, "group_3_id"
	}
	if user.Group4ID != "" {
		return user.Group4ID, "group_4_id"
	}
	if user.Group5ID != "" {
		return user.Group5ID, "group_5_id"
	}
	if user.Group6ID != "" {
		return user.Group6ID, "group_6_id"
	}
	if user.Group7ID != "" {
		return user.Group7ID, "group_7_id"
	}

	return user.Group8ID, "group_8_id"
}

func GetHubspotDocumentsListAsBatch(list []*HubspotDocument, batchSize int) [][]*HubspotDocument {
	batchList := make([][]*HubspotDocument, 0, 0)
	listLen := len(list)
	for i := 0; i < listLen; {
		next := i + batchSize
		if next > listLen {
			next = listLen
		}

		batchList = append(batchList, list[i:next])
		i = next
	}

	return batchList
}

func GetHubspotDocumentsListAsBatchById(list []*HubspotDocument, batchSize int) [][]*HubspotDocument {
	uniqueDocList := make(map[string][]*HubspotDocument)
	for i := range list {
		if _, exist := uniqueDocList[list[i].ID]; !exist {
			uniqueDocList[list[i].ID] = make([]*HubspotDocument, 0)
		}
		uniqueDocList[list[i].ID] = append(uniqueDocList[list[i].ID], list[i])
	}

	batchList := make([][]*HubspotDocument, 0)
	visitedID := make(map[string]bool)
	for i := range list {
		if visitedID[list[i].ID] {
			continue
		}

		listLen := len(batchList)
		docs := uniqueDocList[list[i].ID]
		if listLen == 0 ||
			len(batchList[listLen-1])+len(docs) > batchSize {
			batchList = append(batchList, make([]*HubspotDocument, 0))
			listLen++
		}

		batchList[listLen-1] = append(batchList[listLen-1], docs...)
		visitedID[list[i].ID] = true
	}

	return batchList
}

func CheckIfEngagementV3(document *HubspotDocument) (bool, error) {
	value, err := U.DecodePostgresJsonb(document.Value)
	if err != nil {
		return false, err
	}

	if _, ok := (*value)["properties"]; ok { // Engagement V3 (New payload)
		return true, nil
	}

	if _, ok := (*value)["engagement"]; ok { // Engagement V2 (Old payload)
		return false, nil
	}

	return false, errors.New("invalid engagement document")
}

func CheckIfDealV3(document *HubspotDocument) (bool, error) {
	value, err := U.DecodePostgresJsonb(document.Value)
	if err != nil {
		return false, err
	}

	if _, ok := (*value)["id"]; ok { // Deal V3 (New payload)
		return true, nil
	}

	if _, ok := (*value)["dealId"]; ok { // Deal V2 (Old payload)
		return false, nil
	}

	return false, errors.New("invalid deal document")
}

func CheckIfCompanyV3(document *HubspotDocument) (bool, error) {
	value, err := U.DecodePostgresJsonb(document.Value)
	if err != nil {
		return false, err
	}

	if _, ok := (*value)["id"]; ok { // Company V3 (New payload)
		return true, nil
	}

	if _, ok := (*value)["companyId"]; ok { // Company V2 (Old payload)
		return false, nil
	}

	return false, errors.New("invalid company document")
}

func GetCRMObjectURLKey(projectID int64, source, objectTyp string) string {
	return GetCRMEnrichPropertyKeyByType(source, objectTyp, "$object_url")
}

func GetHubspotAllowedObjectsByPlan(plan string) (map[string]bool, error) {
	allowedObjectsMap := make(map[string]bool)
	if _, exist := HubspotAlowedObjectsByPlan[plan]; !exist {
		return nil, errors.New("invalid hubspot mapped plan")
	}

	for _, obj := range HubspotAlowedObjectsByPlan[plan] {
		allowedObjectsMap[obj] = true
	}
	return allowedObjectsMap, nil
}
