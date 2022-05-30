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
	Timestamp   int64           `gorm:"primary_key:true;auto_increment:false" json:"timestamp"`
	TypeAlias   string          `gorm:"-" json:"type_alias"`
	Value       *postgres.Jsonb `json:"value"`
	Synced      bool            `gorm:"default:false;not null" json:"synced"`
	SyncId      string          `gorm:"default:null" json:"sync_id"`
	UserId      string          `gorm:"default:null" json:"user_id"`
	GroupUserId string          `gorm:"default:null" json:"group_user_id"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	// for internal use only
	timeZone       U.TimeZoneString `gorm:"-" json:"-"`
	dateProperties *map[string]bool `gorm:"-" json:"-"`
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

	HubspotDateTimeLayout   = "2006-01-02T15:04:05.000Z"
	HubspotDataTypeDate     = "date"
	HubspotDataTypeDatetime = "datetime"
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

// HubspotProjectSyncStatus hubspot project sync status
type HubspotProjectSyncStatus struct {
	ProjectID uint64 `json:"project_id"`
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
	ProjectID uint64
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
func GetCRMTimeSeriesByStartTimestamp(projectID uint64, from int64, CRMEventSource string) [][]int64 {
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
func GetHubspotAllowedObjects(projectID uint64) *map[string]string {
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

	if document.Action == HubspotDocumentActionDeleted {
		return GetHubspotDocumentLastModifiedDate(document)
	}

	value, err := U.DecodePostgresJsonb(document.Value)
	if err != nil {
		return 0, err
	}

	if document.Type == HubspotDocumentTypeEngagement {
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

	if document.Action == HubspotDocumentActionDeleted {
		return time.Now().UnixNano() / int64(time.Millisecond), nil
	}

	value, err := U.DecodePostgresJsonb(document.Value)
	if err != nil {
		return 0, err
	}

	if document.Type == HubspotDocumentTypeEngagement {
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
	return user.Group4ID, "group_4_id"
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
