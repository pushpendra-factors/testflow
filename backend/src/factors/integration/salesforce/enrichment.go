package salesforce

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	C "factors/config"
	SDK "factors/sdk"
	"factors/util"
	U "factors/util"

	"factors/model/model"
	"factors/model/store"

	log "github.com/sirupsen/logrus"
)

// Status represents current sync status for a doc type
type Status struct {
	ProjectID uint64 `json:"project_id"`
	Type      string `json:"type"`
	Status    string `json:"status"`
}

var salesforceEnrichOrderByType = [...]int{
	model.SalesforceDocumentTypeLead,
	model.SalesforceDocumentTypeContact,
	model.SalesforceDocumentTypeOpportunity,
	model.SalesforceDocumentTypeCampaign,
	model.SalesforceDocumentTypeCampaignMember,
	model.SalesforceDocumentTypeAccount,
}

// CampaignChildRelationship campaign parent to child relationship
type CampaignChildRelationship struct {
	CampaignMembers RelationshipCampaignMember `json:"CampaignMembers"`
}

// RelationshipCampaignMemberRecord  child campaignmember required field
type RelationshipCampaignMemberRecord struct {
	ID        string `json:"Id"`
	IsDeleted bool   `json:"IsDeleted"`
	LeadID    string `json:"LeadId"`
	ContactID string `json:"ContactId"`
}

// RelationshipCampaignMember campaign members of a campaign
type RelationshipCampaignMember struct {
	TotalSize int                                `json:"totalSize"`
	Done      bool                               `json:"done"`
	Records   []RelationshipCampaignMemberRecord `json:"records"`
}

var opportunityMappingOrder = []string{
	model.SalesforceDocumentTypeNameLead,
	model.SalesforceChildRelationshipNameOpportunityContactRoles,
}

var allowedCampaignFields = map[string]bool{}

func getSalesforceMappedDataTypeValue(projectID uint64, eventName, enKey string, value interface{}) (interface{}, error) {
	if value == nil || value == "" {
		return nil, nil
	}

	if !C.IsEnabledPropertyDetailFromDB() || !C.IsEnabledPropertyDetailByProjectID(projectID) {
		return value, nil
	}

	ptype := store.GetStore().GetPropertyTypeByKeyValue(projectID, eventName, enKey, value, false)

	if ptype == U.PropertyTypeDateTime {
		return model.GetSalesforceDocumentTimestamp(value)
	}

	if ptype == U.PropertyTypeNumerical {
		num, err := U.GetPropertyValueAsFloat64(value)
		if err != nil {
			return nil, errors.New("failed to get numerical property")
		}

		return num, nil
	}

	return value, nil
}

// GetSalesforceDocumentProperties return map of enriched properties
func GetSalesforceDocumentProperties(projectID uint64, document *model.SalesforceDocument) (*map[string]interface{}, *map[string]interface{}, error) {
	var enProperties map[string]interface{}
	err := json.Unmarshal(document.Value.RawMessage, &enProperties)
	if err != nil {
		return nil, nil, err
	}

	filterPropertyFieldsByProjectID(projectID, &enProperties, document.Type)

	enrichedProperties := make(map[string]interface{})
	properties := make(map[string]interface{})

	eventName := model.GetSalesforceEventNameByDocumentAndAction(document, model.SalesforceDocumentUpdated)

	for key, value := range enProperties {
		if value == nil || value == "" {
			continue
		}

		enKey := model.GetCRMEnrichPropertyKeyByType(model.SmartCRMEventSourceSalesforce, model.GetSalesforceAliasByDocType(document.Type), key)
		enValue, err := getSalesforceMappedDataTypeValue(projectID, eventName, enKey, value)
		if err != nil {
			log.WithFields(log.Fields{"project_id": projectID, "property_key": enKey}).WithError(err).Error("Failed to get property value.")
			continue
		}

		if _, exists := enrichedProperties[enKey]; !exists {
			enrichedProperties[enKey] = enValue
		}

		if _, exists := properties[key]; !exists {
			properties[key] = enValue
		}
	}

	return &enrichedProperties, &properties, nil
}

func filterPropertyFieldsByProjectID(projectID uint64, properties *map[string]interface{}, docType int) {

	if projectID == 0 {
		return
	}

	allowedfields := model.GetSalesforceAllowedfiedsByObject(projectID, model.GetSalesforceAliasByDocType(docType))
	for field, value := range *properties {
		if value == nil || value == "" || value == 0 {
			delete(*properties, field)
			continue
		}

		if allowedfields != nil {
			if _, exist := allowedfields[field]; !exist {
				delete(*properties, field)
			}
		}
	}
	delete(*properties, "attributes")                                         // delte nested meta object
	delete(*properties, model.SalesforceChildRelationshipNameCampaignMembers) // delete child relationship data
	delete(*properties, model.SalesforceChildRelationshipNameOpportunityContactRoles)
	delete(*properties, OpportunityLeadID)
}

func getSalesforceAccountID(document *model.SalesforceDocument) (string, error) {
	if document == nil {
		return "", errors.New("invalid document")
	}

	var properties map[string]interface{}
	err := json.Unmarshal(document.Value.RawMessage, &properties)
	if err != nil {
		return "", err
	}

	var accountID string
	var ok bool
	if accountID, ok = properties["Id"].(string); !ok {
		return "", errors.New("account id doest not exist")
	}

	if accountID == "" {
		return "", errors.New("empty account id")
	}

	return accountID, nil
}

func getCustomerUserIDFromProperties(projectID uint64, properties map[string]interface{}, docTypeAlias string, salesforceProjectIdentificationFieldStore *map[uint64]map[string][]string) (string, string) {

	identifiers := model.GetIdentifierPrecendenceOrderByProjectID(projectID)
	for _, indentityType := range identifiers {

		if indentityType == model.IdentificationTypePhone {
			possiblePhoneField := model.GetSalesforcePhoneFieldByProjectIDAndObjectName(projectID, docTypeAlias, salesforceProjectIdentificationFieldStore)
			for _, phoneField := range possiblePhoneField {
				if phoneNo, ok := properties[model.GetCRMEnrichPropertyKeyByType(model.SmartCRMEventSourceSalesforce, docTypeAlias, phoneField)]; ok {
					phoneStr, err := U.GetValueAsString(phoneNo)
					if err != nil || len(phoneStr) < 5 {
						continue
					}

					return store.GetStore().GetUserIdentificationPhoneNumber(projectID, phoneStr)
				}
			}
		} else if indentityType == model.IdentificationTypeEmail {
			possibleEmailField := model.GetSalesforceEmailFieldByProjectIDAndObjectName(projectID, docTypeAlias, salesforceProjectIdentificationFieldStore)
			for _, emailField := range possibleEmailField {
				if email, ok := properties[model.GetCRMEnrichPropertyKeyByType(model.SmartCRMEventSourceSalesforce, docTypeAlias, emailField)].(string); ok && email != "" && util.IsEmail(email) {
					existingEmail, errCode := store.GetStore().GetExistingCustomerUserID(projectID, []string{email})
					if errCode == http.StatusFound {
						return email, existingEmail[email]
					}

					return email, ""
				}
			}
		} else {
			log.WithFields(log.Fields{"project_id": projectID, "identity_type": indentityType, "doc_type": docTypeAlias}).Error("Invalid identifier type")
		}
	}

	return "", ""
}

/*
TrackSalesforceEventByDocumentType tracks salesforce events by action
	for action created -> create both created and updated events with date created timestamp
	for action updated -> create on updated event with last modified timestamp
*/
func TrackSalesforceEventByDocumentType(projectID uint64, trackPayload *SDK.TrackPayload, document *model.SalesforceDocument, customerUserID string, objectType string) (string, string, SDK.TrackPayload, error) {

	var finalPayload SDK.TrackPayload
	if projectID == 0 {
		return "", "", finalPayload, errors.New("invalid project id")
	}

	if trackPayload == nil || document == nil {
		return "", "", finalPayload, errors.New("invalid operation")
	}

	createdTimestamp, err := model.GetSalesforceDocumentTimestampByAction(document, model.SalesforceDocumentCreated)
	if err != nil {
		return "", "", finalPayload, err
	}

	lastModifiedTimestamp, err := model.GetSalesforceDocumentTimestampByAction(document, model.SalesforceDocumentUpdated)
	if err != nil {
		return "", "", finalPayload, err
	}

	var eventID, userID string
	if document.Action == model.SalesforceDocumentCreated {
		finalPayload = *trackPayload
		if finalPayload.UserId == "" {
			newUserID, status := store.GetStore().CreateUser(&model.User{
				ProjectId:      projectID,
				CustomerUserId: customerUserID,
				JoinTimestamp:  createdTimestamp,
			})

			if status != http.StatusCreated {
				return "", "", finalPayload, fmt.Errorf("create user failed for doc type %d, status code %d", document.Type, status)
			}
			finalPayload.UserId = newUserID
		}

		finalPayload.Name = model.GetSalesforceEventNameByDocumentAndAction(document, model.SalesforceDocumentCreated)
		finalPayload.Timestamp = createdTimestamp

		status, trackResponse := SDK.Track(projectID, &finalPayload, true, SDK.SourceSalesforce, objectType)
		if status != http.StatusOK && status != http.StatusFound && status != http.StatusNotModified {
			return "", "", finalPayload, fmt.Errorf("created event track failed for doc type %d, message %s", document.Type, trackResponse.Error)
		}

		if finalPayload.UserId != "" {
			userID = finalPayload.UserId
		} else {
			userID = trackResponse.UserId
			// writing back the same userID to final payload to use this for offline touch point
			finalPayload.UserId = userID
		}

		eventID = trackResponse.EventId
	}

	if document.Action == model.SalesforceDocumentCreated || document.Action == model.SalesforceDocumentUpdated {
		finalPayload = *trackPayload
		finalPayload.Name = model.GetSalesforceEventNameByDocumentAndAction(document, model.SalesforceDocumentUpdated)

		if document.Action == model.SalesforceDocumentUpdated {

			finalPayload.Timestamp = lastModifiedTimestamp
			// TODO(maisa): Use GetSyncedSalesforceDocumentByType while updating multiple contacts in an account object
			documents, status := store.GetStore().GetSyncedSalesforceDocumentByType(projectID, []string{document.ID}, document.Type, false)
			if status != http.StatusFound {
				return "", "", finalPayload, errors.New("failed to get synced document")
			}

			event, status := store.GetStore().GetEventById(projectID, documents[0].SyncID, "")
			if status != http.StatusFound {
				return "", "", finalPayload, errors.New("failed to get event from sync id ")
			}

			if customerUserID != "" {
				status, _ = SDK.Identify(projectID, &SDK.IdentifyPayload{
					UserId:         event.UserId,
					CustomerUserId: customerUserID,
					Timestamp:      lastModifiedTimestamp,
				}, false)

				if status != http.StatusOK {
					return "", "", finalPayload, fmt.Errorf("failed indentifying user on update event track")
				}
			}
			userID = event.UserId
		} else {
			finalPayload.Timestamp = createdTimestamp
		}

		finalPayload.UserId = userID

		status, trackResponse := SDK.Track(projectID, &finalPayload, true, SDK.SourceSalesforce, objectType)
		if status != http.StatusOK && status != http.StatusFound && status != http.StatusNotModified {
			return "", "", finalPayload, fmt.Errorf("updated event track failed for doc type %d", document.Type)
		}

		eventID = trackResponse.EventId
	} else {
		return "", "", finalPayload, errors.New("invalid action on salesforce document sync")
	}

	// create additional event for created action if document is not the first version
	if document.Action == model.SalesforceDocumentCreated && createdTimestamp != lastModifiedTimestamp {
		payload := *trackPayload
		payload.Timestamp = lastModifiedTimestamp
		payload.UserId = userID
		payload.Name = model.GetSalesforceEventNameByDocumentAndAction(document, model.SalesforceDocumentUpdated)
		status, _ := SDK.Track(projectID, &payload, true, SDK.SourceSalesforce, objectType)
		if status != http.StatusOK && status != http.StatusFound && status != http.StatusNotModified {
			return "", "", finalPayload, fmt.Errorf("updated event for different timestamp track failed for doc type %d", document.Type)
		}
	}

	return eventID, userID, finalPayload, nil
}

func getAccountGroupID(enProperties *map[string]interface{}) string {

	accountName := util.GetPropertyValueAsString((*enProperties)[model.GetCRMEnrichPropertyKeyByType(model.SmartCRMEventSourceSalesforce,
		model.SalesforceDocumentTypeNameAccount, "name")])
	accountID := util.GetPropertyValueAsString((*enProperties)[model.GetCRMEnrichPropertyKeyByType(model.SmartCRMEventSourceSalesforce,
		model.SalesforceDocumentTypeNameAccount, "id")])
	accountWebsite := util.GetPropertyValueAsString((*enProperties)[model.GetCRMEnrichPropertyKeyByType(model.SmartCRMEventSourceSalesforce,
		model.SalesforceDocumentTypeNameAccount, "website")])
	if accountName != "" {
		return accountName
	}

	if accountWebsite != "" {
		return accountWebsite
	}

	return accountID
}

func enrichGroupAcccount(projectID uint64, document *model.SalesforceDocument) int {
	logCtx := log.WithField("project_id", projectID).
		WithFields(log.Fields{"doc_id": document.ID, "doc_action": document.Action, "doc_timestamp": document.Timestamp})

	if projectID == 0 || document == nil {
		return http.StatusBadRequest
	}

	if document.Type != model.SalesforceDocumentTypeAccount || document.GroupUserID != "" {
		return http.StatusInternalServerError
	}

	enProperties, _, err := GetSalesforceDocumentProperties(projectID, document)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get properties")
		return http.StatusInternalServerError
	}

	accountID := getAccountGroupID(enProperties)

	createdTimestamp, err := model.GetSalesforceDocumentTimestampByAction(document, model.SalesforceDocumentCreated)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get account created timestamp.")
		return http.StatusInternalServerError
	}

	lastModifiedTimestamp, err := model.GetSalesforceDocumentTimestampByAction(document, model.SalesforceDocumentUpdated)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get account last modified timestamp.")
		return http.StatusInternalServerError
	}

	accountUserID := ""
	var processEventNames []string
	var processEventTimestamps []int64
	if document.Action == model.SalesforceDocumentCreated {
		accountUserID, err = store.GetStore().CreateOrUpdateCompanyGroupPropertiesBySource(projectID, accountID, "", enProperties, createdTimestamp, lastModifiedTimestamp, model.SmartCRMEventSourceSalesforce)
		if err != nil {
			logCtx.WithError(err).Error("Failed to update salesforce account group.")
			return http.StatusInternalServerError
		}

		errCode := store.GetStore().UpdateSalesforceDocumentBySyncStatus(projectID, document, "", "", accountUserID, false)
		if errCode != http.StatusAccepted {
			logCtx.Error("Failed to set group_user_id in salesforce account created document.")
			return http.StatusInternalServerError
		}

		processEventNames = append(processEventNames, util.GROUP_EVENT_NAME_SALESFORCE_ACCOUNT_CREATED, util.GROUP_EVENT_NAME_SALESFORCE_ACCOUNT_UPDATED)
		processEventTimestamps = append(processEventTimestamps, createdTimestamp, createdTimestamp)

		if createdTimestamp != lastModifiedTimestamp {
			processEventNames = append(processEventNames, util.GROUP_EVENT_NAME_SALESFORCE_ACCOUNT_UPDATED)
			processEventTimestamps = append(processEventTimestamps, lastModifiedTimestamp)
		}
	}

	if document.Action == model.SalesforceDocumentUpdated {
		documents, status := store.GetStore().GetSyncedSalesforceDocumentByType(projectID, []string{document.ID}, document.Type, true)
		if status != http.StatusFound {
			return http.StatusInternalServerError
		}

		createdDocument := documents[0]

		accountUserID, err = store.GetStore().CreateOrUpdateCompanyGroupPropertiesBySource(projectID, accountID, createdDocument.GroupUserID, enProperties, createdTimestamp, lastModifiedTimestamp, model.SmartCRMEventSourceSalesforce)
		if err != nil {
			logCtx.WithError(err).Error("Failed to update salesforce account group properties.")
			return http.StatusInternalServerError
		}

		if createdDocument.GroupUserID == "" {
			errCode := store.GetStore().UpdateSalesforceDocumentBySyncStatus(projectID, &createdDocument, "", "", accountUserID, false)
			if errCode != http.StatusAccepted {
				logCtx.Error("Failed to update group_user_id in salesforce account created document.")
				return http.StatusInternalServerError
			}
		}

		errCode := store.GetStore().UpdateSalesforceDocumentBySyncStatus(projectID, document, "", "", accountUserID, false)
		if errCode != http.StatusAccepted {
			logCtx.Error("Failed to update group_user_id in salesforce account updated document.")
			return http.StatusInternalServerError
		}

		processEventNames = append(processEventNames, util.GROUP_EVENT_NAME_SALESFORCE_ACCOUNT_UPDATED)
		processEventTimestamps = append(processEventTimestamps, lastModifiedTimestamp)
	}

	for i := range processEventNames {

		trackPayload := &SDK.TrackPayload{
			Name:      processEventNames[i],
			ProjectId: projectID,
			Timestamp: processEventTimestamps[i],
			UserId:    accountUserID,
		}

		status, response := SDK.Track(projectID, trackPayload, true, SDK.SourceSalesforce, "")
		if status != http.StatusOK && status != http.StatusFound && status != http.StatusNotModified {
			logCtx.WithFields(log.Fields{"status": status, "track_response": response, "event_name": processEventNames[i],
				"event_timestamp": processEventTimestamps[i]}).Error("Failed to track salesforce account event.")
			return http.StatusInternalServerError
		}

	}

	return http.StatusOK
}

func enrichAccount(projectID uint64, document *model.SalesforceDocument, salesforceSmartEventNames []SalesforceSmartEventName) int {
	if projectID == 0 || document == nil {
		return http.StatusBadRequest
	}

	if document.Type != model.SalesforceDocumentTypeAccount {
		return http.StatusInternalServerError
	}

	logCtx := log.WithField("project_id", projectID).WithField("document_id", document.ID)

	enProperties, properties, err := GetSalesforceDocumentProperties(projectID, document)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get properties")
		return http.StatusInternalServerError
	}

	trackPayload := &SDK.TrackPayload{
		ProjectId:       projectID,
		EventProperties: *enProperties,
		UserProperties:  *enProperties,
	}

	customerUserID, _ := getCustomerUserIDFromProperties(projectID, *enProperties, model.GetSalesforceAliasByDocType(document.Type), &model.SalesforceProjectIdentificationFieldStore)
	if customerUserID == "" {
		logCtx.Warn("Skipping user identification on salesforce account sync. No customer_user_id on properties.")
	}

	eventID, userID, _, err := TrackSalesforceEventByDocumentType(projectID, trackPayload, document, customerUserID, model.SalesforceDocumentTypeNameAccount)
	if err != nil {
		logCtx.WithError(err).Error(
			"Failed to track salesforce account event.")
		return http.StatusInternalServerError
	}

	// Always use lastmodified timestamp for updated properties. Error handling already done during event creation
	lastModifiedTimestamp, _ := model.GetSalesforceDocumentTimestampByAction(document, model.SalesforceDocumentUpdated)

	var prevProperties *map[string]interface{}
	for _, smartEventName := range salesforceSmartEventNames {
		prevProperties = TrackSalesforceSmartEvent(projectID, &smartEventName, eventID, document.ID, userID, document.Type,
			properties, prevProperties, lastModifiedTimestamp)
	}

	errCode := store.GetStore().UpdateSalesforceDocumentBySyncStatus(projectID, document, eventID, userID, "", true)
	if errCode != http.StatusAccepted {
		logCtx.Error("Failed to update salesforce account document as synced.")
		return http.StatusInternalServerError
	}

	return http.StatusOK
}

// SalesforceSmartEventName struct for holding event_name and filter expression
type SalesforceSmartEventName struct {
	EventName string
	Filter    *model.SmartCRMEventFilter
	Type      string
}

func getTimestampFromField(propertyName string, properties *map[string]interface{}) (int64, error) {
	if timestamp, exists := (*properties)[propertyName]; exists {
		if unixTimestamp, ok := timestamp.(int64); ok {
			return unixTimestamp, nil
		}

		unixTimestamp, err := model.GetSalesforceDocumentTimestamp(timestamp)
		if err != nil {
			return 0, err
		}

		return unixTimestamp, nil
	}

	return 0, errors.New("field missing")
}

func enrichContact(projectID uint64, document *model.SalesforceDocument, salesforceSmartEventNames []SalesforceSmartEventName) int {
	if projectID == 0 || document == nil {
		return http.StatusBadRequest
	}

	if document.Type != model.SalesforceDocumentTypeContact {
		return http.StatusInternalServerError
	}

	logCtx := log.WithField("project_id", projectID).WithField("document_id", document.ID)
	enProperties, properties, err := GetSalesforceDocumentProperties(projectID, document)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get properties")
		return http.StatusInternalServerError
	}

	trackPayload := &SDK.TrackPayload{
		ProjectId:       projectID,
		EventProperties: *enProperties,
		UserProperties:  *enProperties,
	}

	customerUserID, _ := getCustomerUserIDFromProperties(projectID, *enProperties, model.GetSalesforceAliasByDocType(document.Type), &model.SalesforceProjectIdentificationFieldStore)
	if customerUserID == "" {
		logCtx.Warn("Skipping user identification on salesforce contact sync. No customer_user_id on properties.")
	}

	eventID, userID, _, err := TrackSalesforceEventByDocumentType(projectID, trackPayload, document, customerUserID, model.SalesforceDocumentTypeNameContact)
	if err != nil {
		logCtx.WithError(err).Error("Failed to track salesforce contact event.")
		return http.StatusInternalServerError
	}

	// Always use lastmodified timestamp for updated properties. Error handling already done during event creation
	lastModifiedTimestamp, _ := model.GetSalesforceDocumentTimestampByAction(document, model.SalesforceDocumentUpdated)

	var prevProperties *map[string]interface{}
	for _, smartEventName := range salesforceSmartEventNames {
		prevProperties = TrackSalesforceSmartEvent(projectID, &smartEventName, eventID, document.ID, userID, document.Type,
			properties, prevProperties, lastModifiedTimestamp)
	}

	errCode := store.GetStore().UpdateSalesforceDocumentBySyncStatus(projectID, document, eventID, userID, "", true)
	if errCode != http.StatusAccepted {
		logCtx.Error("Failed to update salesforce contact document as synced.")
		return http.StatusInternalServerError
	}

	if C.IsAllowedSalesforceGroupsByProjectID(projectID) {
		accountID := util.GetPropertyValueAsString((*enProperties)[model.GetCRMEnrichPropertyKeyByType(model.SmartCRMEventSourceSalesforce,
			model.GetSalesforceAliasByDocType(document.Type), "accountid")])
		if accountID != "" {
			status := updateSalesforceUserAccountGroups(projectID, accountID, userID)
			if status != http.StatusOK {
				logCtx.Error("Failed to update salesforce contact group details.")
			}
		}
	}

	return http.StatusOK
}

/*
GetSalesforceSmartEventPayload return smart event payload if the rule successfully gets passed.
WITHOUT PREVIOUS PROPERTY :- A query will be made for previous synced record which
will require userID or customerUserID and doctType
WITH PREVIOUS PROPERTY := userID, customerUserID and doctType won't be used
*/
func GetSalesforceSmartEventPayload(projectID uint64, eventName, documentID, userID string, docType int,
	currentProperties, prevProperties *map[string]interface{}, filter *model.SmartCRMEventFilter) (*model.CRMSmartEvent, *map[string]interface{}, bool) {

	var crmSmartEvent model.CRMSmartEvent
	var validProperty bool
	var newProperties map[string]interface{}

	logCtx := log.WithFields(log.Fields{"project_id": projectID, "doc_type": docType, "document_id": documentID,
		"doc_id": docType, "smart_event_rule": filter})
	if projectID == 0 || eventName == "" || filter == nil || currentProperties == nil {
		logCtx.Error("Missing required fields.")
		return nil, prevProperties, false
	}

	if prevProperties == nil && (documentID == "" || docType == 0 || userID == "") {
		logCtx.Error("Missing required fields.")
		return nil, prevProperties, false
	}

	if prevProperties != nil {
		validProperty = model.CRMFilterEvaluator(projectID, currentProperties, prevProperties, filter, model.CompareStateBoth)
	} else {
		validProperty = model.CRMFilterEvaluator(projectID, currentProperties, nil, filter, model.CompareStateCurr)
	}

	if !validProperty {
		return nil, prevProperties, false
	}

	if prevProperties == nil {
		prevDocs, status := store.GetStore().GetSyncedSalesforceDocumentByType(
			projectID, []string{documentID}, docType, false)
		if status != http.StatusFound && status != http.StatusNotFound {
			return nil, prevProperties, false
		}

		var err error
		if status == http.StatusNotFound {
			prevProperties = &map[string]interface{}{}
		} else {
			_, prevProperties, err = GetSalesforceDocumentProperties(projectID, &prevDocs[len(prevDocs)-1])
			if err != nil {
				logCtx.WithError(err).Error("Failed to GetSalesforceDocumentProperties")
				return nil, prevProperties, false
			}
		}

		if !model.CRMFilterEvaluator(projectID, currentProperties, prevProperties,
			filter, model.CompareStateBoth) {
			return nil, prevProperties, false
		}
	}

	crmSmartEvent.Name = eventName
	model.FillSmartEventCRMProperties(&newProperties, currentProperties, prevProperties, filter)
	crmSmartEvent.Properties = newProperties

	return &crmSmartEvent, prevProperties, true
}

// TrackSalesforceSmartEvent valids current properties with CRM smart filter and creates a event
func TrackSalesforceSmartEvent(projectID uint64, salesforceSmartEventName *SalesforceSmartEventName, eventID, documentID, userID string, docType int,
	currentProperties, prevProperties *map[string]interface{}, lastModifiedTimestamp int64) *map[string]interface{} {
	var valid bool
	var smartEventPayload *model.CRMSmartEvent

	logCtx := log.WithFields(log.Fields{"project_id": projectID, "doc_type": docType,
		"user_id": userID, "document_id": documentID, "smart_event_rule": salesforceSmartEventName})
	if projectID == 0 || currentProperties == nil || docType == 0 || userID == "" || lastModifiedTimestamp == 0 {
		logCtx.Error("Missing required fields.")
		return prevProperties
	}

	if salesforceSmartEventName.EventName == "" || salesforceSmartEventName.Type == "" || salesforceSmartEventName.Filter == nil {
		logCtx.Error("Missing smart event fileds.")
		return prevProperties
	}

	smartEventPayload, prevProperties, valid = GetSalesforceSmartEventPayload(projectID, salesforceSmartEventName.EventName, documentID,
		userID, docType, currentProperties, prevProperties, salesforceSmartEventName.Filter)
	if !valid {
		return prevProperties
	}

	model.AddSmartEventReferenceMeta(&smartEventPayload.Properties, eventID)

	smartEventTrackPayload := &SDK.TrackPayload{
		ProjectId:       projectID,
		EventProperties: smartEventPayload.Properties,
		Name:            smartEventPayload.Name,
		SmartEventType:  salesforceSmartEventName.Type,
		UserId:          userID,
	}

	timestampReferenceField := salesforceSmartEventName.Filter.TimestampReferenceField
	if timestampReferenceField == model.TimestampReferenceTypeDocument {
		smartEventTrackPayload.Timestamp = lastModifiedTimestamp + 1

	} else {
		fieldTimestamp, err := getTimestampFromField(timestampReferenceField, currentProperties)
		if err == nil {
			smartEventTrackPayload.Timestamp = fieldTimestamp + 1
		} else {
			logCtx.WithField("timestamp_reference_field", timestampReferenceField).
				WithError(err).Error("Failed to get timestamp from reference field")
			smartEventTrackPayload.Timestamp = lastModifiedTimestamp + 1

		}
	}

	if !C.IsDryRunCRMSmartEvent() {
		status, _ := SDK.Track(projectID, smartEventTrackPayload, true, SDK.SourceSalesforce, "")
		if status != http.StatusOK && status != http.StatusFound && status != http.StatusNotModified {
			logCtx.Error("Failed to create salesforce smart event")
		}
	} else {
		logCtx.WithFields(log.Fields{"properties": smartEventPayload.Properties, "event_name": smartEventPayload.Name,
			"filter_exp":            *salesforceSmartEventName.Filter,
			"smart_event_timestamp": smartEventTrackPayload.Timestamp}).Info("Dry run smart event creation.")
	}

	return prevProperties
}

type OpportunityContactRoleRecord struct {
	ID            string `json:"Id"` // only required field
	IsPrimary     bool   `json:"IsPrimary"`
	ContactID     string `json:"ContactId"`
	Role          string `json:"Role"`
	OpportunityID string `json:"OpportunityId"`
}

type RelationshipOpportunityContactRole struct {
	Records []OpportunityContactRoleRecord `json:"records"`
}

type OpportunityChildRelationship struct {
	OpportunityContactRole RelationshipOpportunityContactRole `json:"OpportunityContactRoles"`
	OppLeadID              string                             `json:"opportunity_to_lead"`
}

var errMissingOpportunityLeadAndContact = errors.New("missing lead and contact record for opportunity link")

func getOpportuntityLeadAndContactID(document *model.SalesforceDocument) (string, string, error) {
	logCtx := log.WithFields(log.Fields{"project_id": document.ProjectID, "doc_id": document.ID, "doc_type": document.Type})
	var opportunityChildRelationship OpportunityChildRelationship
	err := json.Unmarshal(document.Value.RawMessage, &opportunityChildRelationship)
	if err != nil {
		return "", "", err
	}

	allowedObjects := model.GetSalesforceDocumentTypeAlias(document.ProjectID)
	oppLeadID := ""
	oppContactID := ""

	if _, exist := allowedObjects[model.SalesforceDocumentTypeNameContact]; exist {
		records := opportunityChildRelationship.OpportunityContactRole.Records

		for i := range records {
			if records[i].IsPrimary {
				if records[i].ContactID == "" {
					logCtx.Error("Missing primary contact id.")
					break
				}

				oppContactID = records[i].ContactID
			}
		}
	}

	if _, exist := allowedObjects[model.SalesforceDocumentTypeNameLead]; exist {
		if opportunityChildRelationship.OppLeadID != "" {
			oppLeadID = opportunityChildRelationship.OppLeadID
		}
	}

	return oppLeadID, oppContactID, nil
}

func getOpportunityLinkedLeadOrContactDocument(projectID uint64, document *model.SalesforceDocument) (*model.SalesforceDocument, error) {

	oppLeadID, oppContactID, err := getOpportuntityLeadAndContactID(document)
	if err != nil {
		return nil, err
	}

	for i := range opportunityMappingOrder {
		if opportunityMappingOrder[i] == model.SalesforceChildRelationshipNameOpportunityContactRoles && oppContactID != "" {
			linkedObject, status := store.GetStore().GetSyncedSalesforceDocumentByType(projectID, []string{oppContactID}, model.SalesforceDocumentTypeContact, false)
			if status == http.StatusFound {
				return &linkedObject[0], nil // get the first document
			}

		}

		if opportunityMappingOrder[i] == model.SalesforceDocumentTypeNameLead && oppLeadID != "" {
			linkedObject, status := store.GetStore().GetSyncedSalesforceDocumentByType(projectID, []string{oppLeadID}, model.SalesforceDocumentTypeLead, false)
			if status == http.StatusFound {
				return &linkedObject[0], nil
			}

		}
	}

	if oppLeadID != "" || oppContactID != "" {
		return nil, errMissingOpportunityLeadAndContact
	}

	return nil, errors.New("no object associated with opportunity")
}

func enrichOpportunities(projectID uint64, document *model.SalesforceDocument, salesforceSmartEventNames []SalesforceSmartEventName) int {
	if projectID == 0 || document == nil {
		return http.StatusBadRequest
	}

	if document.Type != model.SalesforceDocumentTypeOpportunity {
		return http.StatusInternalServerError
	}

	logCtx := log.WithField("project_id", projectID).WithField("document_id", document.ID)
	enProperties, properties, err := GetSalesforceDocumentProperties(projectID, document)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get properties")
		return http.StatusInternalServerError
	}

	trackPayload := &SDK.TrackPayload{
		ProjectId:       projectID,
		EventProperties: *enProperties,
		UserProperties:  *enProperties,
	}

	var eventID string
	var eventUserID string
	var customerUserID, userID string
	var assocationPresent bool
	if C.UseOpportunityAssociationByProjectID(projectID) {
		linkedDocument, err := getOpportunityLinkedLeadOrContactDocument(projectID, document)
		if err != nil {
			if err == errMissingOpportunityLeadAndContact {
				// record may not be processed. Should be made success on next call
				logCtx.WithError(err).Error("Failed to get linked document for opportunity.")
				return http.StatusOK
			}
		} else {
			assocationPresent = true
			if linkedDocument.Synced == true && (linkedDocument.UserID != "" || linkedDocument.SyncID != "") {

				linkedDocumentUserID := ""
				if linkedDocument.UserID != "" {
					linkedDocumentUserID = linkedDocument.UserID
				} else {
					event, status := store.GetStore().GetEventById(projectID, linkedDocument.SyncID, "")
					if status != http.StatusFound {
						logCtx.WithFields(log.Fields{"linked_document_id": linkedDocument.ID}).WithError(err).
							Error("Failed to get user by linked document event for opportunity.")
						return http.StatusInternalServerError
					}

					// update the user_id column for later reference
					errCode := store.GetStore().UpdateSalesforceDocumentBySyncStatus(projectID, linkedDocument,
						event.ID, event.UserId, "", true)
					if errCode != http.StatusAccepted {
						logCtx.WithFields(log.Fields{"linked_document_id": linkedDocument.ID}).
							Error("Failed to update user id in linked document.")
					}

					linkedDocumentUserID = event.UserId
				}

				user, status := store.GetStore().GetUser(projectID, linkedDocumentUserID)
				if status != http.StatusFound {
					logCtx.WithError(err).Error("Failed to get opportunity associated document user.")
					return http.StatusInternalServerError
				}
				customerUserID = user.CustomerUserId
				userID = user.ID

			} else {
				/*
					Document associated is not processed yet.
					Skip processing or opportunities event user properties won't have the lead data.
				*/
				logCtx.WithError(err).Error("Failed to process linked document for opportunity.")
				return http.StatusOK
			}

		}
	}

	if userID == "" && customerUserID == "" && !assocationPresent {
		customerUserID, userID = getCustomerUserIDFromProperties(projectID, *enProperties, model.GetSalesforceAliasByDocType(document.Type), &model.SalesforceProjectIdentificationFieldStore)
	}

	if customerUserID != "" || userID != "" {
		if userID != "" {
			trackPayload.UserId = userID // will also handle opportunity updated event which is not stiched with other object
			eventID, eventUserID, _, err = TrackSalesforceEventByDocumentType(projectID, trackPayload, document, "", model.SalesforceDocumentTypeNameOpportunity)
		} else {
			eventID, eventUserID, _, err = TrackSalesforceEventByDocumentType(projectID, trackPayload, document, customerUserID, model.SalesforceDocumentTypeNameOpportunity)
		}

		if err != nil {
			logCtx.WithError(err).Error(
				"Failed to track salesforce opportunity event.")
			return http.StatusInternalServerError
		}
	} else {
		eventID, eventUserID, _, err = TrackSalesforceEventByDocumentType(projectID, trackPayload, document, "", model.SalesforceDocumentTypeNameOpportunity)
		if err != nil {
			logCtx.WithError(err).Error(
				"Failed to track salesforce opportunity event.")
			return http.StatusInternalServerError
		}

		logCtx.Error("Skipped user identification on salesforce opportunity sync. No customer_user_id on properties.")
	}

	if eventUserID != "" && userID != eventUserID {
		userID = eventUserID
	}

	// Always use lastmodified timestamp for updated properties. Error handling already done during event creation
	lastModifiedTimestamp, _ := model.GetSalesforceDocumentTimestampByAction(document, model.SalesforceDocumentUpdated)

	var prevProperties *map[string]interface{}
	for _, smartEventName := range salesforceSmartEventNames {
		prevProperties = TrackSalesforceSmartEvent(projectID, &smartEventName, eventID, document.ID, userID,
			document.Type, properties, prevProperties, lastModifiedTimestamp)
	}

	errCode := store.GetStore().UpdateSalesforceDocumentBySyncStatus(projectID, document, eventID, userID, "", true)
	if errCode != http.StatusAccepted {
		logCtx.Error("Failed to update salesforce opportunity document as synced.")
		return http.StatusInternalServerError
	}

	if C.IsAllowedSalesforceGroupsByProjectID(projectID) {
		accountID := util.GetPropertyValueAsString((*enProperties)[model.GetCRMEnrichPropertyKeyByType(model.SmartCRMEventSourceSalesforce,
			model.GetSalesforceAliasByDocType(document.Type), "accountid")])
		if accountID != "" {
			status := updateSalesforceUserAccountGroups(projectID, accountID, userID)
			if status != http.StatusOK {
				logCtx.Error("Failed to update salesforce opportunity group details.")
			}
		}
	}

	return http.StatusOK
}

func updateSalesforceUserAccountGroups(projectID uint64, accountID, userID string) int {
	documents, status := store.GetStore().GetSyncedSalesforceDocumentByType(projectID, []string{accountID}, model.SalesforceDocumentTypeAccount, true)
	if status != http.StatusFound {
		if status == http.StatusNotFound {
			return http.StatusOK // return ok if account was never capture
		}
		return http.StatusInternalServerError
	}

	groupUserID := documents[0].GroupUserID
	if groupUserID == "" {
		return http.StatusInternalServerError
	}

	groupUser, status := store.GetStore().GetUser(projectID, groupUserID)
	if status != http.StatusFound {
		return http.StatusInternalServerError
	}

	groupID, err := model.GetUserGroupID(groupUser)
	if err != nil {
		log.WithError(err).Error("Failed to get group user group id.")
		return http.StatusInternalServerError
	}

	_, status = store.GetStore().UpdateUserGroup(projectID, userID, model.GROUP_NAME_SALESFORCE_ACCOUNT, groupID, groupUserID)
	if status != http.StatusAccepted && status != http.StatusNotModified {
		log.WithFields(log.Fields{"project_id": projectID, "user_id": userID, "group_user_id": groupUserID, "group_id": groupID}).
			Error("Failed to update salesforce user group id.")
		return http.StatusInternalServerError
	}

	return http.StatusOK
}

func enrichLeads(projectID uint64, document *model.SalesforceDocument, salesforceSmartEventNames []SalesforceSmartEventName) int {
	if projectID == 0 || document == nil {
		return http.StatusBadRequest
	}

	if document.Type != model.SalesforceDocumentTypeLead {
		return http.StatusInternalServerError
	}

	logCtx := log.WithField("project_id", projectID).WithField("document_id", document.ID)
	enProperties, properties, err := GetSalesforceDocumentProperties(projectID, document)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get properties")
		return http.StatusInternalServerError
	}

	trackPayload := &SDK.TrackPayload{
		ProjectId:       projectID,
		EventProperties: *enProperties,
		UserProperties:  *enProperties,
	}

	customerUserID, _ := getCustomerUserIDFromProperties(projectID, *enProperties, model.GetSalesforceAliasByDocType(document.Type), &model.SalesforceProjectIdentificationFieldStore)
	if customerUserID == "" {
		logCtx.Warn("Skipped user identification on salesforce lead sync. No customer_user_id on properties.")
	}

	eventID, userID, _, err := TrackSalesforceEventByDocumentType(projectID, trackPayload, document, customerUserID, model.SalesforceDocumentTypeNameLead)
	if err != nil {
		logCtx.WithError(err).Error("Failed to track salesforce lead event.")
		return http.StatusInternalServerError
	}

	// ALways us lastmodified timestamp for updated properties, error handling already done during event creation
	lastModifiedTimestamp, _ := model.GetSalesforceDocumentTimestampByAction(document, model.SalesforceDocumentUpdated)

	var prevProperties *map[string]interface{}
	for _, smartEventName := range salesforceSmartEventNames {
		prevProperties = TrackSalesforceSmartEvent(projectID, &smartEventName, eventID, document.ID, userID, document.Type,
			properties, prevProperties, lastModifiedTimestamp)
	}

	errCode := store.GetStore().UpdateSalesforceDocumentBySyncStatus(projectID, document, eventID, userID, "", true)
	if errCode != http.StatusAccepted {
		logCtx.Error("Failed to update salesforce lead document as synced.")
		return http.StatusInternalServerError
	}

	if C.IsAllowedSalesforceGroupsByProjectID(projectID) {
		accountID := util.GetPropertyValueAsString((*enProperties)[model.GetCRMEnrichPropertyKeyByType(model.SmartCRMEventSourceSalesforce,
			model.GetSalesforceAliasByDocType(document.Type), "convertedaccountid")])
		if accountID != "" {
			status := updateSalesforceUserAccountGroups(projectID, accountID, userID)
			if status != http.StatusOK {
				logCtx.Error("Failed to update salesforce lead group details.")
			}
		}
	}

	return http.StatusOK
}

func getCampaignMemberIDsFromCampaign(document *model.SalesforceDocument) ([]string, error) {

	var campaignChildRelationship CampaignChildRelationship
	err := json.Unmarshal(document.Value.RawMessage, &campaignChildRelationship)
	if err != nil {
		return nil, err
	}

	records := campaignChildRelationship.CampaignMembers.Records
	campaignMemberIDs := make([]string, len(records))
	for i := range records {
		campaignMemberIDs[i] = records[i].ID
	}

	return campaignMemberIDs, nil
}

func enrichCampaignToAllCampaignMembers(project *model.Project, document *model.SalesforceDocument, endTimestamp int64) int {
	logCtx := log.WithFields(log.Fields{"project_id": project.ID, "document_id": document.ID, "end_timestamp": endTimestamp})
	if document.Type != model.SalesforceDocumentTypeCampaign {
		return http.StatusBadRequest
	}

	enProperties, _, err := GetSalesforceDocumentProperties(project.ID, document)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get properties for campaign.")
		return http.StatusInternalServerError
	}

	campaignMemberIDs, err := getCampaignMemberIDsFromCampaign(document)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get campaign members ids.")
		return http.StatusInternalServerError
	}

	if len(campaignMemberIDs) < 1 {
		status := store.GetStore().UpdateSalesforceDocumentBySyncStatus(project.ID, document, "", "", "", true)
		if status != http.StatusAccepted {
			logCtx.Error("Failed to mark campaign as synced.")
			return http.StatusInternalServerError
		}

		return http.StatusOK
	}

	var memberDocuments []model.SalesforceDocument
	var status int

	/*
		NOTE: IF member document is not available for this time range, mark it as synced.
		This can only happened on the first time of this campaign pull where campaign is created on day 1 and member on day 2

		When CAMPAIGN MEMBER is picked up for processing then, it will refer this document as for last campaign update.
		Refer enrichCampaignMember function for opposite case
	*/
	memberDocuments, status = store.GetStore().GetLatestSalesforceDocumentByID(project.ID, campaignMemberIDs, model.SalesforceDocumentTypeCampaignMember, endTimestamp)
	if status != http.StatusFound {
		logCtx.Warn("Failed to get campaign members.")
		status = store.GetStore().UpdateSalesforceDocumentBySyncStatus(project.ID, document, "", "", "", true)
		if status != http.StatusAccepted {
			logCtx.Error("Failed to mark campaign as synced.")
			return http.StatusInternalServerError
		}

		return http.StatusOK
	}

	for i := range memberDocuments {
		enMemberProperties, _, err := GetSalesforceDocumentProperties(project.ID, &memberDocuments[i])
		if err != nil {
			logCtx.WithError(err).Error("Failed to get campaign member properties.")
			return http.StatusInternalServerError
		}

		for pName := range *enProperties {
			(*enMemberProperties)[pName] = (*enProperties)[pName]
		}

		referenceDocument := memberDocuments[i]
		existingUserID := ""
		if referenceDocument.Action == model.SalesforceDocumentCreated && memberDocuments[i].Synced == false {
			existingUserID = getExistingCampaignMemberUserIDFromProperties(project.ID, enMemberProperties)
			if existingUserID == "" {
				logCtx.WithField("member_id", referenceDocument.ID).Error("Missing lead or contact record for a campaign.")
			}
		} else {
			referenceDocument.Action = model.SalesforceDocumentUpdated
		}

		// use latest timestamp
		if referenceDocument.Timestamp < document.Timestamp {
			referenceDocument.Value = document.Value
			referenceDocument.Timestamp = document.Timestamp
		}

		trackPayload := &SDK.TrackPayload{
			ProjectId:       project.ID,
			EventProperties: *enMemberProperties, // no user properties for campaign members
			UserId:          existingUserID,
		}

		eventID, userID, finalTrackPayload, err := TrackSalesforceEventByDocumentType(project.ID, trackPayload, &referenceDocument, "", "")
		if err != nil {
			logCtx.WithField("member_id", referenceDocument.ID).WithError(err).Error(
				"Failed to track salesforce campaign member update on campaign update.")
			return http.StatusInternalServerError
		}

		if memberDocuments[i].Synced == false {
			err = ApplySFOfflineTouchPointRule(project, &finalTrackPayload, &memberDocuments[i], endTimestamp)
			if err != nil {
				// log and continue
				logCtx.WithField("EventID", eventID).WithField("userID", eventID).WithField("userID", eventID).Info("failed creating SF offline touch point")
			}
		}

		if memberDocuments[i].Synced == false {
			status = store.GetStore().UpdateSalesforceDocumentBySyncStatus(project.ID, &memberDocuments[i], eventID, userID, "", true)
			if status != http.StatusAccepted {
				logCtx.WithField("member_id", referenceDocument.ID).Error("Failed to mark campaign member as synced.")
				return http.StatusInternalServerError
			}
		}
	}

	status = store.GetStore().UpdateSalesforceDocumentBySyncStatus(project.ID, document, "", "", "", true)
	if status != http.StatusAccepted {
		logCtx.Error("Failed to mark campaign as synced.")
		return http.StatusInternalServerError
	}

	return http.StatusOK
}

// Get existing lead or contact user ID from campaign members data
func getExistingCampaignMemberUserIDFromProperties(projectID uint64, properties *map[string]interface{}) string {
	existingUserID := ""
	existingContactMemberID := (*properties)[model.GetCRMEnrichPropertyKeyByType(model.SmartCRMEventSourceSalesforce, model.SalesforceDocumentTypeNameCampaignMember, "ContactId")]
	existingLeadMemberID := (*properties)[model.GetCRMEnrichPropertyKeyByType(model.SmartCRMEventSourceSalesforce, model.SalesforceDocumentTypeNameCampaignMember, "LeadId")]

	// use contact Id associated user id. Once user converts from lead to contact, salesforce prioritize contact based identification
	if existingContactMemberID != "" {
		existingMember, status := store.GetStore().GetSyncedSalesforceDocumentByType(projectID, []string{util.GetPropertyValueAsString(existingContactMemberID)}, model.SalesforceDocumentTypeContact, false)
		if status == http.StatusFound {
			if existingMember[0].UserID != "" {
				existingUserID = existingMember[0].UserID
			}
		}
	}

	if existingUserID == "" { // use lead Id if available
		existingMember, status := store.GetStore().GetSyncedSalesforceDocumentByType(projectID, []string{util.GetPropertyValueAsString(existingLeadMemberID)}, model.SalesforceDocumentTypeLead, false)
		if status == http.StatusFound {
			if existingMember[0].UserID != "" {
				existingUserID = existingMember[0].UserID
			}
		}
	}

	return existingUserID
}

func enrichCampaignMember(project *model.Project, document *model.SalesforceDocument, endTimestamp int64) int {
	logCtx := log.WithFields(log.Fields{"project_id": project.ID, "document_id": document.ID})
	if document.Type != model.SalesforceDocumentTypeCampaignMember {
		return http.StatusBadRequest
	}

	enCampaignMemberProperties, _, err := GetSalesforceDocumentProperties(project.ID, document)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get properties for campaign member.")
		return http.StatusInternalServerError
	}

	campaignID, exist := (*enCampaignMemberProperties)[model.GetCRMEnrichPropertyKeyByType(model.SmartCRMEventSourceSalesforce, model.SalesforceDocumentTypeNameCampaignMember, "CampaignId")]
	if !exist {
		logCtx.Error("Missing campaign_id in campaign member")
		return http.StatusInternalServerError
	}

	/*
		NOTE: IF campaign document is not available for this time range don't mark it as synced and continue.

		When CAMPAIGN is picked up for processing then, it will refer this document as for last campaign member.
		Refer enrichCampaignToAllCampaignMembers function for opposite case
	*/
	campaignDocuments, status := store.GetStore().GetLatestSalesforceDocumentByID(project.ID, []string{util.GetPropertyValueAsString(campaignID)}, model.SalesforceDocumentTypeCampaign, endTimestamp)
	if status != http.StatusFound { // log warning and don't mark it as synced. It will be processed when campaign is found
		logCtx.Warn("Failed to get campaign document for campaign member.")
		return http.StatusOK
	}

	enCampaignProperties, _, err := GetSalesforceDocumentProperties(project.ID, &campaignDocuments[len(campaignDocuments)-1])
	if err != nil {
		logCtx.WithError(err).Error("Failed to get properties for campaign member.")
		return http.StatusInternalServerError
	}

	for pName := range *enCampaignProperties {
		(*enCampaignMemberProperties)[pName] = (*enCampaignProperties)[pName]
	}

	existingUserID := ""
	// use user_id from lead or contact id
	if document.Action == model.SalesforceDocumentCreated {
		existingUserID = getExistingCampaignMemberUserIDFromProperties(project.ID, enCampaignMemberProperties)
	}

	trackPayload := &SDK.TrackPayload{
		ProjectId:       project.ID,
		EventProperties: *enCampaignMemberProperties,
		UserId:          existingUserID,
	}

	eventID, userID, finalTrackPayload, err := TrackSalesforceEventByDocumentType(project.ID, trackPayload, document, "", "")
	if err != nil {
		logCtx.WithError(err).Error("Failed to track salesforce lead event.")
		return http.StatusInternalServerError
	}

	err = ApplySFOfflineTouchPointRule(project, &finalTrackPayload, document, endTimestamp)
	if err != nil {
		// log and continue
		logCtx.WithField("EventID", eventID).WithField("userID", eventID).WithField("userID", eventID).Info("Create SF offline touch point")
	}

	errCode := store.GetStore().UpdateSalesforceDocumentBySyncStatus(project.ID, document, eventID, userID, "", true)
	if errCode != http.StatusAccepted {
		logCtx.Error("Failed to update salesforce lead document as synced.")
		return http.StatusInternalServerError
	}
	return http.StatusOK
}

func ApplySFOfflineTouchPointRule(project *model.Project, trackPayload *SDK.TrackPayload, document *model.SalesforceDocument, endTimestamp int64) error {

	logCtx := log.WithFields(log.Fields{"project_id": project.ID, "method": "ApplySFOfflineTouchPointRule", "document_id": document.ID, "document_action": document.Action})

	if &project.SalesforceTouchPoints != nil && !U.IsEmptyPostgresJsonb(&project.SalesforceTouchPoints) {

		var touchPointRules map[string][]model.SFTouchPointRule
		err := U.DecodePostgresJsonbToStructType(&project.SalesforceTouchPoints, &touchPointRules)
		if err != nil {
			// logging and continuing.
			logCtx.WithField("Document", trackPayload).WithError(err).Error("Failed to fetch offline touch point rules for salesforce document.")
			return err
		}

		rules := touchPointRules["sf_touch_point_rules"]

		for _, rule := range rules {

			// check if rule is applicable
			if !canCreateSFTouchPoint(rule.TouchPointTimeRef, document.Action) || !filterCheck(rule, trackPayload, logCtx) {
				continue
			}

			switch document.Action {

			case model.SalesforceDocumentCreated:

				_, err = CreateTouchPointEvent(project, trackPayload, document, rule)
				if err != nil {
					logCtx.WithError(err).Error("failed to create touch point for salesforce campaign member document.")
					continue
				}
			case model.SalesforceDocumentUpdated:

				campaignMemberDocuments, status := store.GetStore().GetLatestSalesforceDocumentByID(project.ID, []string{util.GetPropertyValueAsString(document.ID)}, model.SalesforceDocumentTypeCampaignMember, endTimestamp)
				if status != http.StatusFound {
					logCtx.Warn("failed to get campaign member salesforce document for campaign member.")
					continue
				}
				logCtx.WithField("Total_Documents", len(campaignMemberDocuments)).WithField("Document[0]", campaignMemberDocuments[0]).Info("Found existing campaign member document")

				// len(campaignMemberDocuments) > 0 && timestamp sorted desc
				enCampaignMemberProperties, _, err := GetSalesforceDocumentProperties(project.ID, &campaignMemberDocuments[0])
				if err != nil {
					logCtx.WithError(err).Error("Failed to get properties for salesforce campaign member.")
					continue
				}
				// ignore to create a new touch point if last updated doc has EP_SFCampaignMemberResponded=true
				if val, exists := (*enCampaignMemberProperties)[model.EP_SFCampaignMemberResponded]; exists {
					if val.(bool) == true {
						continue
					}
				}
				logCtx.Info("Found existing salesforce campaign member document")
				if val, exists := trackPayload.EventProperties[model.EP_SFCampaignMemberResponded]; exists {
					if val.(bool) == true {
						_, err = CreateTouchPointEvent(project, trackPayload, document, rule)
						if err != nil {
							logCtx.WithError(err).Error("failed to create touch point for salesforce campaign member document.")
							continue
						}
					}
				}
			}
		}
	}
	return nil
}

func CreateTouchPointEvent(project *model.Project, trackPayload *SDK.TrackPayload, document *model.SalesforceDocument, rule model.SFTouchPointRule) (*SDK.TrackResponse, error) {

	logCtx := log.WithFields(log.Fields{"project_id": project.ID, "method": "CreateTouchPointEvent", "document_id": document.ID, "document_action": document.Action})
	logCtx.WithField("document", document).WithField("trackPayload", trackPayload).Info("CreateTouchPointEvent: creating salesforce document")
	var trackResponse *SDK.TrackResponse
	var err error
	payload := &SDK.TrackPayload{
		ProjectId:       project.ID,
		EventProperties: trackPayload.EventProperties,
		UserId:          trackPayload.UserId,
		Name:            U.EVENT_NAME_OFFLINE_TOUCH_POINT,
	}

	var timestamp int64
	if document.Action == model.SalesforceDocumentCreated {
		timestamp, err = model.GetSalesforceDocumentTimestampByAction(document, model.SalesforceDocumentCreated)
		if err != nil {
			logCtx.Error("failed to timestamp for SF for offline touch point.")
			return trackResponse, err
		}
	} else if document.Action == model.SalesforceDocumentUpdated {
		timestamp, err = model.GetSalesforceDocumentTimestampByAction(document, model.SalesforceDocumentUpdated)
		if err != nil {
			logCtx.Error("failed to timestamp for SF for offline touch point.")
			return trackResponse, err
		}
	}
	payload.Timestamp = timestamp

	// Mapping touch point properties:
	for key, value := range rule.PropertiesMap {
		if _, exists := trackPayload.EventProperties[value]; exists {
			payload.EventProperties[key] = trackPayload.EventProperties[value]
		} else {
			payload.EventProperties[key] = model.PropertyValueNone
		}
	}

	status, trackResponse := SDK.Track(project.ID, payload, true, "", "")
	if status != http.StatusOK && status != http.StatusFound && status != http.StatusNotModified {
		logCtx.WithField("Document", trackPayload).WithError(err).Error(fmt.Errorf("create salesforce touchpoint event track failed for doc type %d, message %s", document.Type, trackResponse.Error))
		return trackResponse, errors.New(fmt.Sprintf("create salesforce touchpoint event track failed for doc type %d, message %s", document.Type, trackResponse.Error))
	}
	logCtx.WithField("statusCode", status).WithField("trackResponsePayload", trackResponse).Info("Successfully: created salesforce offline touch point")
	return trackResponse, nil
}

func canCreateSFTouchPoint(touchPointTimeRef string, documentActionType model.SalesforceAction) bool {
	// Ignore created event for "first responded" based rule.
	if touchPointTimeRef == model.SFCampaignMemberResponded && documentActionType == model.SalesforceDocumentCreated {
		return false
	}

	// Ignore responded event for "Created event" based rule.
	if touchPointTimeRef == model.SFCampaignMemberCreated && documentActionType == model.SalesforceDocumentUpdated {
		return false
	}
	return true
}

func filterCheck(rule model.SFTouchPointRule, trackPayload *SDK.TrackPayload, logCtx *log.Entry) bool {

	filtersPassed := 0
	for _, filter := range rule.Filters {
		switch filter.Operator {
		case model.EqualsOpStr:
			if _, exists := trackPayload.EventProperties[filter.Property]; exists {
				if filter.Value != "" && trackPayload.EventProperties[filter.Property] == filter.Value {
					filtersPassed++
				}
			}
		case model.NotEqualOpStr:
			if _, exists := trackPayload.EventProperties[filter.Property]; exists {
				if filter.Value != "" && trackPayload.EventProperties[filter.Property] != filter.Value {
					filtersPassed++
				}
			}
		case model.ContainsOpStr:
			if _, exists := trackPayload.EventProperties[filter.Property]; exists {
				if filter.Value != "" && strings.Contains(trackPayload.EventProperties[filter.Property].(string), filter.Value) {
					filtersPassed++
				}
			}
		default:
			logCtx.WithField("Document", trackPayload).Error("No matching operator found for offline touch point rules for salesforce document.")
			continue
		}
	}
	return filtersPassed != 0 && filtersPassed == len(rule.Filters)
}

/*
	Campaign{
		ID:
		Name:
		CampaignMembers:{
			Records:[{
				ID:
			},
			{
				ID:
			}
			]
		}
	}

	CampaignMember{
		ID:
		CampaignID:
		LeadID:
		ContactID:
	}
*/
func enrichCampaign(project *model.Project, document *model.SalesforceDocument, endTimestamp int64) int {
	if project.ID == 0 || document == nil {
		return http.StatusBadRequest
	}

	if document.Type == model.SalesforceDocumentTypeCampaign {
		return enrichCampaignToAllCampaignMembers(project, document, endTimestamp)
	}

	if document.Type == model.SalesforceDocumentTypeCampaignMember {
		return enrichCampaignMember(project, document, endTimestamp)
	}

	return http.StatusBadRequest
}

func enrichAll(project *model.Project, documents []model.SalesforceDocument, salesforceSmartEventNames []SalesforceSmartEventName, endTimestamp int64) int {
	if project.ID == 0 {
		return http.StatusBadRequest
	}
	logCtx := log.WithField("project_id", project.ID)

	var seenFailures bool
	var errCode int
	for i := range documents {
		startTime := time.Now().Unix()

		switch documents[i].Type {
		case model.SalesforceDocumentTypeAccount:
			errCode = enrichAccount(project.ID, &documents[i], salesforceSmartEventNames)
		case model.SalesforceDocumentTypeContact:
			errCode = enrichContact(project.ID, &documents[i], salesforceSmartEventNames)
		case model.SalesforceDocumentTypeLead:
			errCode = enrichLeads(project.ID, &documents[i], salesforceSmartEventNames)
		case model.SalesforceDocumentTypeOpportunity:
			errCode = enrichOpportunities(project.ID, &documents[i], salesforceSmartEventNames)
		case model.SalesforceDocumentTypeCampaign, model.SalesforceDocumentTypeCampaignMember:
			errCode = enrichCampaign(project, &documents[i], endTimestamp)
		default:
			log.Errorf("invalid salesforce document type found %d", documents[i].Type)
			continue
		}

		if errCode != http.StatusOK {
			seenFailures = true
		}

		logCtx.WithField("time_taken_in_secs", time.Now().Unix()-startTime).Debugf(
			"Sync %s completed.", documents[i].TypeAlias)
	}

	if seenFailures {
		return http.StatusInternalServerError
	}

	return http.StatusOK
}

// GetSalesforceSmartEventNames returns all the smart_event for salesforce by object_type
func GetSalesforceSmartEventNames(projectID uint64) *map[string][]SalesforceSmartEventName {

	logCtx := log.WithFields(log.Fields{"project_id": projectID})

	eventNames, errCode := store.GetStore().GetSmartEventFilterEventNames(projectID, false)
	if errCode == http.StatusInternalServerError {
		logCtx.Error("Error while GetSmartEventFilterEventNames")
	}

	salesforceSmartEventNames := make(map[string][]SalesforceSmartEventName)

	if len(eventNames) == 0 {
		return &salesforceSmartEventNames
	}

	for i := range eventNames {
		if eventNames[i].Type != model.TYPE_CRM_SALESFORCE {
			continue
		}

		var salesforceSmartEventName SalesforceSmartEventName
		decFilterExp, err := model.GetDecodedSmartEventFilterExp(eventNames[i].FilterExpr)
		if err != nil {
			if err == model.ErrorSmartEventFiterEmptyString {
				logCtx.WithError(err).Warn("Empty string on smart event filter.")
			} else {
				logCtx.WithError(err).Error("Failed to decode smart event filter expression")
			}
			continue
		}

		salesforceSmartEventName.EventName = eventNames[i].Name
		salesforceSmartEventName.Filter = decFilterExp
		salesforceSmartEventName.Type = model.TYPE_CRM_SALESFORCE

		if _, exists := salesforceSmartEventNames[decFilterExp.ObjectType]; !exists {
			salesforceSmartEventNames[decFilterExp.ObjectType] = []SalesforceSmartEventName{}
		}

		salesforceSmartEventNames[decFilterExp.ObjectType] = append(salesforceSmartEventNames[decFilterExp.ObjectType], salesforceSmartEventName)
	}

	return &salesforceSmartEventNames
}

func enrichGroup(projectID uint64) int {
	logCtx := log.WithFields(log.Fields{"project_id": projectID})

	if projectID == 0 {
		logCtx.Error("Invalid project id.")
		return http.StatusBadRequest
	}

	documents, errCode := store.GetStore().GetSalesforceDocumentsByTypeForSync(projectID, model.SalesforceDocumentTypeAccount, 0, 0)
	if errCode != http.StatusFound {
		logCtx.Error("Failed to get salesforce account documents for groups.")
		return http.StatusInternalServerError
	}

	var seenFailures bool
	for i := range documents {
		startTime := time.Now().Unix()

		errCode = enrichGroupAcccount(projectID, &documents[i])
		if errCode != http.StatusOK {
			seenFailures = true
		}
		logCtx.WithFields(log.Fields{"time_taken_in_secs": time.Now().Unix() - startTime, "doc_id": documents[i].ID}).
			Debug("Completed group document sync.")
	}

	if seenFailures {
		return http.StatusInternalServerError
	}

	return http.StatusOK
}

// Enrich sync salesforce documents to events
func Enrich(projectID uint64) ([]Status, bool) {

	logCtx := log.WithField("project_id", projectID)

	statusByProjectAndType := make([]Status, 0, 0)
	if projectID == 0 {
		return statusByProjectAndType, true
	}

	status := CreateOrGetSalesforceEventName(projectID)
	if status != http.StatusOK {
		statusByProjectAndType = append(statusByProjectAndType, Status{ProjectID: projectID,
			Status: "Failed to create event names"})
		return statusByProjectAndType, true
	}

	// Get/Create SF touch point event name
	_, status = store.GetStore().CreateOrGetOfflineTouchPointEventName(projectID)
	if status != http.StatusFound && status != http.StatusConflict && status != http.StatusCreated {
		logCtx.Error("failed to create event name on SF for offline touch point")
		return statusByProjectAndType, true
	}

	allowedDocTypes := model.GetSalesforceDocumentTypeAlias(projectID)

	salesforceSmartEventNames := GetSalesforceSmartEventNames(projectID)

	docMinTimestamp, minTimestamp, errCode := store.GetStore().GetSalesforceDocumentBeginingTimestampByDocumentTypeForSync(projectID)
	if errCode != http.StatusFound {
		if errCode == http.StatusNotFound {
			statusByProjectAndType = append(statusByProjectAndType, Status{ProjectID: projectID,
				Status: U.CRM_SYNC_STATUS_SUCCESS})
			return statusByProjectAndType, false
		}

		logCtx.WithField("err_code", errCode).Error("Failed to get time series.")
		statusByProjectAndType = append(statusByProjectAndType, Status{ProjectID: projectID,
			Status: "Failed to get time series."})
		return statusByProjectAndType, true
	}

	orderedTimeSeries := model.GetCRMTimeSeriesByStartTimestamp(projectID, minTimestamp, model.SmartCRMEventSourceSalesforce)

	project, errCode := store.GetStore().GetProject(projectID)
	if errCode != http.StatusFound {
		log.Error("Failed to get project")
		return statusByProjectAndType, true
	}

	anyFailure := false
	overAllSyncStatus := make(map[string]bool)

	if C.IsAllowedSalesforceGroupsByProjectID(projectID) {
		errCode = enrichGroup(projectID)
		if errCode != http.StatusOK {
			overAllSyncStatus[model.SalesforceDocumentTypeNameGroupAccount] = true
		} else {
			overAllSyncStatus[model.SalesforceDocumentTypeNameGroupAccount] = false
		}
	}

	for _, timeRange := range orderedTimeSeries {

		for _, docType := range salesforceEnrichOrderByType {

			if docMinTimestamp[docType] <= 0 || timeRange[1] < docMinTimestamp[docType] {
				continue
			}

			docTypeAlias := model.GetSalesforceAliasByDocType(docType)
			if _, exist := allowedDocTypes[docTypeAlias]; !exist {
				continue
			}

			logCtx = logCtx.WithFields(log.Fields{"type": docTypeAlias, "time_range": timeRange, "project_id": projectID})
			logCtx.Info("Processing started for given time range")

			var documents []model.SalesforceDocument
			documents, errCode = store.GetStore().GetSalesforceDocumentsByTypeForSync(projectID, docType, timeRange[0], timeRange[1])

			if errCode != http.StatusFound {
				logCtx.Error("Failed to get salesforce document by type for sync.")
				continue
			}

			errCode = enrichAll(project, documents, (*salesforceSmartEventNames)[docTypeAlias], timeRange[1])
			if errCode == http.StatusOK {
				if _, exist := overAllSyncStatus[docTypeAlias]; !exist {
					overAllSyncStatus[docTypeAlias] = false
				}
			} else {
				overAllSyncStatus[docTypeAlias] = true
			}

		}
	}

	for docTypeAlias, failure := range overAllSyncStatus {
		status := Status{ProjectID: projectID,
			Type: docTypeAlias}
		if failure {
			status.Status = U.CRM_SYNC_STATUS_FAILURES
			anyFailure = true
		} else {
			status.Status = U.CRM_SYNC_STATUS_SUCCESS
		}
		statusByProjectAndType = append(statusByProjectAndType, status)
	}

	return statusByProjectAndType, anyFailure
}
