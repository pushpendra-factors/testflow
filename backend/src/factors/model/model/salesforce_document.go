package model

import (
	"errors"
	"factors/util"
	"fmt"
	"strings"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

// SalesforceDocument is an interface for salesforce_documents table
type SalesforceDocument struct {
	ProjectID   int64            `gorm:"primary_key:true;auto_increment:false" json:"project_id"`
	ID          string           `gorm:"primary_key:true;auto_increment:false" json:"id"`
	Type        int              `gorm:"primary_key:true;auto_increment:false" json:"type"`
	Action      SalesforceAction `gorm:"auto_increment:false;not null" json:"action"`
	Timestamp   int64            `gorm:"primary_key:true;auto_increment:false" json:"timestamp"`
	TypeAlias   string           `gorm:"-" json:"type_alias"`
	Value       *postgres.Jsonb  `json:"value"`
	Synced      bool             `gorm:"default:false;not null" json:"synced"`
	SyncID      string           `gorm:"default:null" json:"sync_id"`
	UserID      string           `gorm:"default:null" json:"user_id"`
	GroupUserID string           `gorm:"default:null" json:"group_user_id"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
	SyncTries   int              `gorm:"default:0" json:"sync_tries"`

	// fields for internal use
	dateTimeZone   util.TimeZoneString `gorm:"-" json:"-"`
	dateProperties *map[string]bool    `gorm:"-" json:"-"`
}

type SalesforceAction int

// SalesforceLastSyncInfo contains information about the latest timestamp and type of document for a project
type SalesforceLastSyncInfo struct {
	ProjectID int64 `json:"-"`
	Type      int   `json:"type"`
	Timestamp int64 `json:"timestamp"`
}

// SalesforceSyncInfo lists project_id and their last sync info per doc type
type SalesforceSyncInfo struct {
	ProjectSettings map[int64]*SalesforceProjectSettings `json:"project_settings"`
	// project_id: { type: last_sync_info }
	LastSyncInfo              map[int64]map[string]int64 `json:"last_sync_info"`
	DeletedRecordLastSyncInfo map[int64]map[string]int64 `json:"deleted_record_last_sync_info"`
}

// SalesforceRecord is map for fields and their values
type SalesforceRecord map[string]interface{}

const (
	SalesforceDataTypeDate     = "date"
	SalesforceDataTypeDateTime = "datetime"
)

var (
	salesforceDataTypeDatetime = map[string]bool{
		SalesforceDataTypeDateTime: true,
		SalesforceDataTypeDate:     true,
	}

	salesforceDataTypeNumerical = map[string]bool{
		"double":   true,
		"int":      true,
		"long":     true,
		"currency": true,
	}
)

var SalesforceAllowedObjectsByPlan = map[string][]string{
	FEATURE_SALESFORCE_BASIC: {SalesforceDocumentTypeNameAccount, SalesforceDocumentTypeNameOpportunity, SalesforceDocumentTypeNameUser},
	FEATURE_SALESFORCE: {SalesforceDocumentTypeNameContact, SalesforceDocumentTypeNameLead, SalesforceDocumentTypeNameAccount,
		SalesforceDocumentTypeNameOpportunity, SalesforceDocumentTypeNameCampaign, SalesforceDocumentTypeNameCampaignMember,
		SalesforceDocumentTypeNameOpportunityContactRole, SalesforceDocumentTypeNameTask,
		SalesforceDocumentTypeNameEvent, SalesforceDocumentTypeNameUser},
}

var SalesforceEventNametoDocTypeMapping = map[string]string{
	util.GROUP_EVENT_NAME_SALESFORCE_OPPORTUNITY_CREATED: SalesforceDocumentTypeNameOpportunity,
	util.GROUP_EVENT_NAME_SALESFORCE_OPPORTUNITY_UPDATED: SalesforceDocumentTypeNameOpportunity,
	util.GROUP_EVENT_NAME_SALESFORCE_ACCOUNT_CREATED:     SalesforceDocumentTypeNameAccount,
	util.GROUP_EVENT_NAME_SALESFORCE_ACCOUNT_UPDATED:     SalesforceDocumentTypeNameAccount,
}

// GetSalesforceMappedDataType returns mapped factors data type
func GetSalesforceMappedDataType(dataType string) string {
	if dataType == "" {
		return ""
	}

	if _, exists := salesforceDataTypeDatetime[dataType]; exists {
		return util.PropertyTypeDateTime
	}

	if _, exists := salesforceDataTypeNumerical[dataType]; exists {
		return util.PropertyTypeNumerical
	}

	return util.PropertyTypeUnknown
}

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
	SalesforceDocumentTypeContact                = 1
	SalesforceDocumentTypeLead                   = 2
	SalesforceDocumentTypeAccount                = 3
	SalesforceDocumentTypeOpportunity            = 4
	SalesforceDocumentTypeCampaign               = 5
	SalesforceDocumentTypeCampaignMember         = 6
	SalesforceDocumentTypeGroupAccount           = 7
	SalesforceDocumentTypeOpportunityContactRole = 8
	SalesforceDocumentTypeTask                   = 9
	SalesforceDocumentTypeEvent                  = 10
	SalesforceDocumentTypeUser                   = 11

	SalesforceDocumentTypeNameContact                = "contact"
	SalesforceDocumentTypeNameLead                   = "lead"
	SalesforceDocumentTypeNameAccount                = "account"
	SalesforceDocumentTypeNameOpportunity            = "opportunity"
	SalesforceDocumentTypeNameCampaign               = "campaign"
	SalesforceDocumentTypeNameCampaignMember         = "campaignmember"
	SalesforceDocumentTypeNameGroupAccount           = "group_account"
	SalesforceDocumentTypeNameOpportunityContactRole = "opportunityContactRole"
	SalesforceDocumentTypeNameTask                   = "task"
	SalesforceDocumentTypeNameEvent                  = "event"
	SalesforceDocumentTypeNameUser                   = "user"

	SFCampaignMemberResponded                              = "campaign_member_first_responded_date"
	SFCampaignMemberCreated                                = "campaign_member_created_date"
	EP_SFCampaignMemberResponded                           = "$salesforce_campaignmember_hasresponded"
	EP_SFCampaignMemberFirstRespondedDate                  = "$salesforce_campaignmember_firstrespondeddate"
	EP_SFCampaignMemberStatus                              = "$salesforce_campaignmember_status"
	EP_SFCampaignMemberUpdated                             = "$salesforce_campaignmember_lastmodifieddate"
	EP_SFCampaignMemberCreated                             = "$salesforce_campaignmember_createddate"
	SalesforceDocumentCreated             SalesforceAction = 1
	SalesforceDocumentUpdated             SalesforceAction = 2
	SalesforceDocumentDeleted             SalesforceAction = 3

	// Standard template for salesforce date time
	SalesforceDocumentDateTimeLayout = "2006-01-02T15:04:05.000-0700"
	SalesforceDocumentDateLayout     = "2006-01-02"
)

// Parent to child relationship for query related data, use plural form of names
const (
	SalesforceChildRelationshipNameCampaignMembers         = "CampaignMembers"
	SalesforceChildRelationshipNameOpportunityContactRoles = "OpportunityContactRoles"
)

// SalesforceDocumentTypeAlias maps document type to alias
var SalesforceDocumentTypeAlias = map[string]int{
	SalesforceDocumentTypeNameContact:                SalesforceDocumentTypeContact,
	SalesforceDocumentTypeNameLead:                   SalesforceDocumentTypeLead,
	SalesforceDocumentTypeNameAccount:                SalesforceDocumentTypeAccount,
	SalesforceDocumentTypeNameOpportunity:            SalesforceDocumentTypeOpportunity,
	SalesforceDocumentTypeNameCampaign:               SalesforceDocumentTypeCampaign,
	SalesforceDocumentTypeNameCampaignMember:         SalesforceDocumentTypeCampaignMember,
	SalesforceDocumentTypeNameGroupAccount:           SalesforceDocumentTypeGroupAccount,
	SalesforceDocumentTypeNameOpportunityContactRole: SalesforceDocumentTypeOpportunityContactRole,
	SalesforceDocumentTypeNameTask:                   SalesforceDocumentTypeTask,
	SalesforceDocumentTypeNameEvent:                  SalesforceDocumentTypeEvent,
	SalesforceDocumentTypeNameUser:                   SalesforceDocumentTypeUser,
}

// SalesforceStandardDocumentType will be pulled if no custom list is provided
var SalesforceStandardDocumentType = []int{
	SalesforceDocumentTypeAccount,
	SalesforceDocumentTypeContact,
	SalesforceDocumentTypeLead,
	SalesforceDocumentTypeOpportunity,
}

// SalesforceDocumentActionAlias Salesforce document alias and their action
var SalesforceDocumentActionAlias = map[int]string{
	int(SalesforceDocumentCreated): "SalesforceDocumentActionCreated",
	int(SalesforceDocumentUpdated): "SalesforceDocumentActionUpdated",
}

// SalesforceCampaignDocuments campaign related documents
var SalesforceCampaignDocuments = []int{
	SalesforceDocumentTypeCampaign,
	SalesforceDocumentTypeCampaignMember,
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
func GetSalesforceDocumentTypeAlias(projectID int64) map[string]int {
	docTypes := make(map[string]int)
	for _, doctype := range GetSalesforceAllowedObjects(projectID) {
		docTypes[GetSalesforceAliasByDocType(doctype)] = doctype
	}

	return docTypes
}

// GetSalesforceEventNameByDocumentAndAction creates event name by SalesforceAction and doc type name
func GetSalesforceEventNameByDocumentAndAction(doc *SalesforceDocument, action SalesforceAction) string {
	typAlias := GetSalesforceAliasByDocType(doc.Type)

	return GetSalesforceEventNameByAction(typAlias, action)
}

func GetSalesforceCustomEventNameByType(typAlias string) string {
	if typAlias == "" {
		return ""
	}

	if typAlias == SalesforceDocumentTypeNameCampaignMember {
		return "$sf_campaign_member_responded_to_campaign"
	}
	return ""
}

func GetSalesforceEventNameByAction(typAlias string, action SalesforceAction) string {
	if typAlias == "" || action == 0 {
		return ""
	}

	if typAlias == SalesforceDocumentTypeNameCampaignMember || typAlias == SalesforceDocumentTypeNameCampaign {
		typAlias = "campaign_member"
	}

	if action == SalesforceDocumentCreated {
		return fmt.Sprintf("$sf_%s_created", typAlias)
	}
	if action == SalesforceDocumentUpdated {
		return fmt.Sprintf("$sf_%s_updated", typAlias)
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

const SALESFORCE_OBJECT_DELETED_KEY = "IsDeleted"

func isDeletedDocument(document *SalesforceDocument) (bool, error) {
	if document.Type == 0 {
		return false, errors.New("invalid document type")
	}

	value, err := util.DecodePostgresJsonb(document.Value)
	if err != nil {
		return false, err
	}

	deletedKey := SALESFORCE_OBJECT_DELETED_KEY
	deleted, exists := (*value)[deletedKey]
	if !exists {
		return false, nil
	}

	return deleted == true, nil
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

func GetDateAsMidnightTimestampByTimeZone(date interface{}, timeZoneStr util.TimeZoneString) (interface{}, error) {
	if date == nil || date == "" {
		return "", nil
	}

	loc, err := time.LoadLocation(string(timeZoneStr))
	if err != nil {
		log.WithField("time_zone", timeZoneStr).WithError(err).Error("Failed to parse time location.")
		return nil, err
	}

	t, err := time.ParseInLocation(SalesforceDocumentDateLayout, util.GetPropertyValueAsString(date), loc)
	if err != nil {
		log.WithField("date", date).WithError(err).Error("Failed to parse date in timezone.")
		return nil, err
	}

	return t.Unix(), nil
}

// GetSalesforceDocumentTimestamp return unix timestamp for salesforce formated timestamp
func GetSalesforceDocumentTimestamp(timestamp interface{}) (int64, error) {
	timestampStr, ok := timestamp.(string)
	if !ok || timestampStr == "" {
		return 0, errors.New("invalid timestamp")
	}

	t, err := time.Parse(SalesforceDocumentDateTimeLayout, timestampStr)
	if err != nil {
		loc, err := time.LoadLocation(string(util.TimeZoneStringIST))
		if err != nil {
			return 0, err
		}

		t, err := time.ParseInLocation(SalesforceDocumentDateLayout, timestampStr, loc)
		if err != nil {
			return 0, err
		}

		return t.Unix(), nil
	}

	return t.Unix(), nil
}

func GetSalesforceDocumentsWithActionAndTimestamp(documents []*SalesforceDocument, existDocuments map[string]bool) []*SalesforceDocument {
	documentsWithAction := make([]*SalesforceDocument, 0)

	documentsIDs := make(map[string]bool)
	for i := range documents {

		if _, exist := documentsIDs[documents[i].ID]; exist {
			log.WithFields(log.Fields{"document": documents[i]}).Error("Found duplicate salesforce document on same batch.")
		}

		documentsIDs[documents[i].ID] = true

		timestamp, err := GetSalesforceLastModifiedTimestamp(documents[i])
		if err != nil {
			log.WithError(err).Error("Failed to get last modified timestamp on salesforce document. Skipping.")
			continue
		}

		documents[i].Timestamp = timestamp

		deleted, err := isDeletedDocument(documents[i])
		if err == nil && deleted {
			if existDocuments[documents[i].ID] {
				documents[i].Action = SalesforceDocumentDeleted
				documents[i].Timestamp++
				documentsWithAction = append(documentsWithAction, documents[i])
			}
			continue
		}

		if !existDocuments[documents[i].ID] {
			documents[i].Action = SalesforceDocumentCreated
			documentsWithAction = append(documentsWithAction, documents[i])
		} else {
			documents[i].Action = SalesforceDocumentUpdated
			documentsWithAction = append(documentsWithAction, documents[i])
		}
	}

	return documentsWithAction
}

func GetSalesforceDocumentsAsBatch(list []*SalesforceDocument, batchSize int) [][]*SalesforceDocument {
	batchList := make([][]*SalesforceDocument, 0, 0)
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

func (document *SalesforceDocument) SetDocumentTimeZone(timeZone util.TimeZoneString) {
	document.dateTimeZone = timeZone
}

func (document *SalesforceDocument) SetDateProperties(dateProperties *map[string]bool) {
	document.dateProperties = dateProperties
}

func (document *SalesforceDocument) GetDocumentTimeZone() util.TimeZoneString {
	return document.dateTimeZone
}

func (document *SalesforceDocument) GetDateProperties() *map[string]bool {
	return document.dateProperties
}

func GetSalesforceAllowedObjectsByPlan(plan string) (map[string]bool, error) {
	allowedObjectsMap := map[string]bool{}
	if _, exist := SalesforceAllowedObjectsByPlan[plan]; !exist {
		return nil, errors.New("invalid salesforce mapped plan")
	}

	for _, obj := range SalesforceAllowedObjectsByPlan[plan] {
		allowedObjectsMap[obj] = true
	}

	return allowedObjectsMap, nil
}
