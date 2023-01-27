package salesforce

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	C "factors/config"
	SDK "factors/sdk"
	"factors/util"
	U "factors/util"

	"factors/model/model"
	"factors/model/store"

	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

// Status represents current sync status for a doc type
type Status struct {
	ProjectID int64  `json:"project_id"`
	Type      string `json:"type"`
	Status    string `json:"status"`
	Message   string `json:"message,omitempty"`
}

var salesforceEnrichOrderByType = [...]int{
	model.SalesforceDocumentTypeLead,
	model.SalesforceDocumentTypeContact,
	model.SalesforceDocumentTypeCampaignMember,
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

func getSalesforceMappedDataTypeValue(projectID int64, eventName, enKey string, value interface{}, typeAlias string, dateProperties *map[string]bool, timeZoneStr U.TimeZoneString) (interface{}, error) {
	if value == nil || value == "" {
		return "", nil
	}

	if !C.IsEnabledPropertyDetailFromDB() || !C.IsEnabledPropertyDetailByProjectID(projectID) {
		return value, nil
	}

	if dateProperties != nil {
		for key := range *dateProperties {
			dateEnKey := model.GetCRMEnrichPropertyKeyByType(model.SmartCRMEventSourceSalesforce, typeAlias, key)
			if enKey != dateEnKey {
				continue
			}

			enValue, err := model.GetDateAsMidnightTimestampByTimeZone(value, timeZoneStr)
			if err != nil {
				return nil, err
			}
			return enValue, nil

		}
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
func GetSalesforceDocumentProperties(projectID int64, document *model.SalesforceDocument) (*map[string]interface{}, *map[string]interface{}, error) {

	var enProperties map[string]interface{}
	err := json.Unmarshal(document.Value.RawMessage, &enProperties)
	if err != nil {
		return nil, nil, err
	}

	filterPropertyFieldsByProjectID(projectID, &enProperties, document.Type)

	enrichedProperties := make(map[string]interface{})
	properties := make(map[string]interface{})

	eventName := model.GetSalesforceEventNameByDocumentAndAction(document, model.SalesforceDocumentUpdated)

	dateProperties := document.GetDateProperties()
	for key, value := range enProperties {

		typeAlias := model.GetSalesforceAliasByDocType(document.Type)
		enKey := model.GetCRMEnrichPropertyKeyByType(model.SmartCRMEventSourceSalesforce, typeAlias, key)

		enValue, err := getSalesforceMappedDataTypeValue(projectID, eventName, enKey, value, typeAlias, dateProperties, document.GetDocumentTimeZone())
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

func filterPropertyFieldsByProjectID(projectID int64, properties *map[string]interface{}, docType int) {

	if projectID == 0 {
		return
	}

	allowedfields := model.GetSalesforceAllowedfiedsByObject(projectID, model.GetSalesforceAliasByDocType(docType))
	for field := range *properties {

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
	delete(*properties, OpportunityMultipleLeadID)
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

func getCustomerUserIDFromProperties(projectID int64, properties map[string]interface{}, docTypeAlias string, salesforceProjectIdentificationFieldStore *map[int64]map[string][]string) (string, string) {

	identifiers := model.GetIdentifierPrecendenceOrderByProjectID(projectID)
	for _, indentityType := range identifiers {

		if indentityType == model.IdentificationTypePhone {
			possiblePhoneField := model.GetSalesforcePhoneFieldByProjectIDAndObjectName(projectID, docTypeAlias, salesforceProjectIdentificationFieldStore)
			for _, phoneField := range possiblePhoneField {
				if phoneNo, ok := properties[model.GetCRMEnrichPropertyKeyByType(model.SmartCRMEventSourceSalesforce, docTypeAlias, phoneField)]; ok {
					phoneStr, err := U.GetValueAsString(phoneNo)
					if err != nil || !U.IsValidPhone(phoneStr) {
						continue
					}

					return store.GetStore().GetUserIdentificationPhoneNumber(projectID, phoneStr)
				}
			}
		} else if indentityType == model.IdentificationTypeEmail {
			possibleEmailField := model.GetSalesforceEmailFieldByProjectIDAndObjectName(projectID, docTypeAlias, salesforceProjectIdentificationFieldStore)
			for _, emailField := range possibleEmailField {
				if email, ok := properties[model.GetCRMEnrichPropertyKeyByType(model.SmartCRMEventSourceSalesforce, docTypeAlias, emailField)].(string); ok && email != "" && util.IsEmail(email) {
					existingEmails, errCode := store.GetStore().GetExistingUserByCustomerUserID(projectID, []string{email})
					if errCode == http.StatusFound {
						for userID, existingEmail := range existingEmails {
							if existingEmail == email {
								return email, userID
							}
						}
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
func TrackSalesforceEventByDocumentType(projectID int64, trackPayload *SDK.TrackPayload, document *model.SalesforceDocument, customerUserID string, objectType string) (string, string, SDK.TrackPayload, error) {

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
				Source:         model.GetRequestSourcePointer(model.UserSourceSalesforce),
			})

			if status != http.StatusCreated {
				return "", "", finalPayload, fmt.Errorf("create user failed for doc type %d, status code %d", document.Type, status)
			}
			finalPayload.UserId = newUserID
		}

		finalPayload.Name = model.GetSalesforceEventNameByDocumentAndAction(document, model.SalesforceDocumentCreated)
		finalPayload.Timestamp = createdTimestamp
		finalPayload.RequestSource = model.UserSourceSalesforce

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

	var isNewCampaignMember bool
	if document.Action == model.SalesforceDocumentCreated && document.Type == model.SalesforceDocumentTypeCampaignMember {
		isNewCampaignMember = true
	}

	if (document.Action == model.SalesforceDocumentCreated || document.Action == model.SalesforceDocumentUpdated) && !isNewCampaignMember {
		finalPayload = *trackPayload
		finalPayload.Name = model.GetSalesforceEventNameByDocumentAndAction(document, model.SalesforceDocumentUpdated)

		if document.Action == model.SalesforceDocumentUpdated {
			finalPayload.Timestamp = lastModifiedTimestamp
			// TODO(maisa): Use GetSyncedSalesforceDocumentByType while updating multiple contacts in an account object
			documents, status := store.GetStore().GetSyncedSalesforceDocumentByType(projectID, []string{document.ID}, document.Type, false)
			if status != http.StatusFound {
				return "", "", finalPayload, errors.New("failed to get synced document")
			}

			_, userID, status := store.GetStore().GetUserIdFromEventId(projectID, documents[0].SyncID, "")
			if status != http.StatusFound {
				return "", "", finalPayload, errors.New("failed to get event from sync id ")
			}

			if customerUserID != "" {
				status, _ = SDK.Identify(projectID, &SDK.IdentifyPayload{
					UserId:         userID,
					CustomerUserId: customerUserID,
					Timestamp:      lastModifiedTimestamp,
					RequestSource:  model.UserSourceSalesforce,
				}, false)

				if status != http.StatusOK {
					return "", "", finalPayload, fmt.Errorf("failed indentifying user on update event track")
				}
			}
		} else {
			finalPayload.Timestamp = createdTimestamp
		}

		finalPayload.UserId = userID
		finalPayload.RequestSource = model.UserSourceSalesforce

		status, trackResponse := SDK.Track(projectID, &finalPayload, true, SDK.SourceSalesforce, objectType)
		if status != http.StatusOK && status != http.StatusFound && status != http.StatusNotModified {
			return "", "", finalPayload, fmt.Errorf("updated event track failed for doc type %d", document.Type)
		}

		eventID = trackResponse.EventId
	} else if !isNewCampaignMember {
		return "", "", finalPayload, errors.New("invalid action on salesforce document sync")
	}

	// create additional event for created action if document is not the first version
	if document.Action == model.SalesforceDocumentCreated && createdTimestamp != lastModifiedTimestamp {
		payload := *trackPayload
		payload.Timestamp = lastModifiedTimestamp
		payload.UserId = userID
		payload.Name = model.GetSalesforceEventNameByDocumentAndAction(document, model.SalesforceDocumentUpdated)
		payload.RequestSource = model.UserSourceSalesforce
		status, _ := SDK.Track(projectID, &payload, true, SDK.SourceSalesforce, objectType)
		if status != http.StatusOK && status != http.StatusFound && status != http.StatusNotModified {
			return "", "", finalPayload, fmt.Errorf("updated event for different timestamp track failed for doc type %d", document.Type)
		}
	}

	return eventID, userID, finalPayload, nil
}

func getAccountGroupID(enProperties *map[string]interface{}, document *model.SalesforceDocument) string {

	if document.ID != "" {
		return document.ID
	}

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

func enrichGroupAccount(projectID int64, document *model.SalesforceDocument, salesforceSmartEventNames []SalesforceSmartEventName) int {
	logCtx := log.WithField("project_id", projectID).
		WithFields(log.Fields{"doc_id": document.ID, "doc_action": document.Action, "doc_timestamp": document.Timestamp})

	if projectID == 0 || document == nil {
		logCtx.Error("Invalid parameters for enrichGroupAccount")
		return http.StatusBadRequest
	}

	if (document.Type != model.SalesforceDocumentTypeAccount && document.Type != model.SalesforceDocumentTypeGroupAccount) || document.GroupUserID != "" {
		logCtx.WithField("doc_type", document.Type).Error("Invalid document type for enrichGroupAccount")
		return http.StatusInternalServerError
	}

	enProperties, properties, err := GetSalesforceDocumentProperties(projectID, document)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get properties")
		return http.StatusInternalServerError
	}

	accountID := getAccountGroupID(enProperties, document)

	groupAccountUserID, status, eventID := createOrUpdateSalesforceGroupsProperties(projectID, document, model.GROUP_NAME_SALESFORCE_ACCOUNT, accountID)
	if status != http.StatusOK {
		logCtx.Error("Failed to create or update salesforce groups properties.")
		return status
	}

	document.GroupUserID = groupAccountUserID

	// Always use lastmodified timestamp for updated properties. Error handling already done during event creation
	lastModifiedTimestamp, _ := model.GetSalesforceDocumentTimestampByAction(document, model.SalesforceDocumentUpdated)

	var prevProperties *map[string]interface{}
	for _, smartEventName := range salesforceSmartEventNames {
		prevProperties = TrackSalesforceSmartEvent(projectID, &smartEventName, eventID, document.ID, groupAccountUserID, document.Type,
			properties, prevProperties, lastModifiedTimestamp, false)
	}

	errCode := store.GetStore().UpdateSalesforceDocumentBySyncStatus(projectID, document, eventID, "", groupAccountUserID, true)
	if errCode != http.StatusAccepted {
		logCtx.Error("Failed to update salesforce account document as synced.")
		return http.StatusInternalServerError
	}

	return http.StatusOK
}

// SalesforceSmartEventName struct for holding event_name and filter expression
type SalesforceSmartEventName struct {
	EventName   string
	EventNameID string
	Filter      *model.SmartCRMEventFilter
	Type        string
}

func GetTimestampFromField(propertyName string, properties *map[string]interface{}) (int64, error) {
	return getTimestampFromField(propertyName, properties)
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

func enrichContact(projectID int64, document *model.SalesforceDocument, salesforceSmartEventNames []SalesforceSmartEventName, pendingOpportunityAssociatedRecords map[string]string) int {
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
		RequestSource:   model.UserSourceSalesforce,
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
			properties, prevProperties, lastModifiedTimestamp, false)
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

		if groupUserID, exist := pendingOpportunityAssociatedRecords[document.ID]; exist {
			_, status := store.GetStore().UpdateUserGroup(projectID, userID, model.GROUP_NAME_SALESFORCE_OPPORTUNITY, "", groupUserID)
			if status != http.StatusAccepted && status != http.StatusNotModified {
				logCtx.Error("Failed associating contact user with group opportunity.")
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
func GetSalesforceSmartEventPayload(projectID int64, eventName, documentID, userID string, docType int,
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
func TrackSalesforceSmartEvent(projectID int64, salesforceSmartEventName *SalesforceSmartEventName, eventID, documentID, userID string, docType int,
	currentProperties, prevProperties *map[string]interface{}, lastModifiedTimestamp int64, isPast bool) *map[string]interface{} {
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
		RequestSource:   model.UserSourceSalesforce,
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

	if isPast {
		createSmartEventWithEmptyUserProperties(projectID, userID, salesforceSmartEventName.EventNameID, smartEventPayload.Properties, smartEventTrackPayload.Timestamp, eventID)
		return prevProperties
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

func createSmartEventWithEmptyUserProperties(projectID int64, userID, eventNameID string, eventProperties map[string]interface{}, eventTimestamp int64, referenceEvenID string) error {
	logCtx := log.WithFields(log.Fields{"project_id": projectID, "user_id": userID, "event_name_id": eventNameID, "smart_event_properties": eventProperties, "event_timestamp": eventTimestamp, "reference_event_id": referenceEvenID})

	if projectID <= 0 || userID == "" || eventNameID == "" || len(eventProperties) <= 0 || eventTimestamp <= 0 || referenceEvenID == "" {
		logCtx.Error("Invalid paramters on createSmartEventWithEmptyUserProperties.")
		return errors.New("invalid parameter on createSmartEventWithEmptyUserProperties")
	}

	logCtx.Info("Running in past mode.")

	exist, err := store.GetStore().IsSmartEventAlreadyExist(projectID, userID, eventNameID, referenceEvenID, eventTimestamp)
	if err != nil {
		logCtx.WithError(err).Error("Failed to validate existing smart event.")
		return err
	}
	if exist {
		logCtx.Info("Event already exist")
		return nil
	}

	// set skip session = true
	eventProperties[util.EP_SKIP_SESSION] = util.PROPERTY_VALUE_TRUE
	propertiesJsonb, err := util.EncodeToPostgresJsonb(&eventProperties)
	if err != nil {
		logCtx.WithError(err).Error("Failed to encode properties to jsonb.")
		return err
	}

	// Use create event to avoid user properties getting set
	event := &model.Event{
		EventNameId:    eventNameID,
		ProjectId:      projectID,
		UserId:         userID,
		Properties:     *propertiesJsonb,
		Timestamp:      eventTimestamp,
		UserProperties: &postgres.Jsonb{[]byte(`{}`)},
	}

	logCtx.Info("Salesforce smart event.")

	if !C.IsDryRunCRMSmartEvent() {

		createdEvent, status := store.GetStore().CreateEvent(event)

		if status != http.StatusOK && status != http.StatusCreated {
			logCtx.WithFields(log.Fields{"err_code": status}).Error("Failed to create salesforce smart event")
		} else {
			logCtx.WithFields(log.Fields{"created_event": createdEvent}).Info("Created salesforce smart event.")
		}
	}

	return nil
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
	OpportunityContactRole    RelationshipOpportunityContactRole `json:"OpportunityContactRoles"`
	OppLeadID                 string                             `json:"opportunity_to_lead"`
	OpportunityMultipleLeadID map[string]bool                    `json:"opportunity_to_multiple_lead"`
}

type RelationshipActivityWho struct {
	ID         string                 `json:"Id"`
	Type       string                 `json:"Type"`
	Attributes map[string]interface{} `json:"Attributes"`
}

type RelationshipActivityRecord struct {
	Who   RelationshipActivityWho `json:"Who"`
	WhoId string                  `json:"WhoId"`
}

var errMissingOpportunityLeadAndContact = errors.New("missing lead and contact record for opportunity link")

func getOpportuntityLeadAndContactID(document *model.SalesforceDocument) ([]string, []string, string, string, error) {
	logCtx := log.WithFields(log.Fields{"project_id": document.ProjectID, "doc_id": document.ID, "doc_type": document.Type})
	var opportunityChildRelationship OpportunityChildRelationship
	err := json.Unmarshal(document.Value.RawMessage, &opportunityChildRelationship)
	if err != nil {
		return nil, nil, "", "", err
	}

	allowedObjects := model.GetSalesforceDocumentTypeAlias(document.ProjectID)
	oppLeadID := ""
	var oppLeadIDs []string
	var oppContactIDs []string
	oppContactID := ""

	if _, exist := allowedObjects[model.SalesforceDocumentTypeNameContact]; exist {
		records := opportunityChildRelationship.OpportunityContactRole.Records

		for i := range records {
			oppContactIDs = append(oppContactIDs, records[i].ContactID)
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
			oppLeadIDs = append(oppLeadIDs, oppLeadID)
		}
	}

	for id := range opportunityChildRelationship.OpportunityMultipleLeadID {
		if id != opportunityChildRelationship.OppLeadID {
			oppLeadIDs = append(oppLeadIDs, id)
		}
	}

	return oppLeadIDs, oppContactIDs, oppLeadID, oppContactID, nil
}

func getOpportunityLinkedLeadOrContactDocument(projectID int64, document *model.SalesforceDocument) (*model.SalesforceDocument, error) {

	_, _, oppLeadID, oppContactID, err := getOpportuntityLeadAndContactID(document)
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

func getGroupEventName(docType int) (string, string) {
	if docType == model.SalesforceDocumentTypeAccount {
		return util.GROUP_EVENT_NAME_SALESFORCE_ACCOUNT_CREATED, util.GROUP_EVENT_NAME_SALESFORCE_ACCOUNT_UPDATED
	}
	if docType == model.SalesforceDocumentTypeOpportunity {
		return util.GROUP_EVENT_NAME_SALESFORCE_OPPORTUNITY_CREATED, util.GROUP_EVENT_NAME_SALESFORCE_OPPORTUNITY_UPDATED
	}
	return "", ""
}

func isValidGroupName(documentType int, groupName string) bool {
	if documentType == model.SalesforceDocumentTypeAccount && groupName == model.GROUP_NAME_SALESFORCE_ACCOUNT {
		return true
	}

	if documentType == model.SalesforceDocumentTypeOpportunity && groupName == model.GROUP_NAME_SALESFORCE_OPPORTUNITY {
		return true
	}
	return false
}

func createOrUpdateSalesforceGroupsProperties(projectID int64, document *model.SalesforceDocument, groupName, groupID string) (string, int, string) {
	logCtx := log.WithFields(log.Fields{"project_id": projectID, "doc_type": document.Type, "document": document, "group_name": groupName,
		"group_id": groupID})

	if projectID == 0 || document == nil {
		logCtx.Error("Invalid parameters")
		return "", http.StatusBadRequest, ""
	}

	if !isValidGroupName(document.Type, groupName) {
		logCtx.Error("Invalid group name")
		return "", http.StatusBadRequest, ""
	}

	enProperties, _, err := GetSalesforceDocumentProperties(projectID, document)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get properties")
		return "", http.StatusInternalServerError, ""
	}

	createdTimestamp, err := model.GetSalesforceDocumentTimestampByAction(document, model.SalesforceDocumentCreated)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get created timestamp.")
		return "", http.StatusInternalServerError, ""
	}

	lastModifiedTimestamp, err := model.GetSalesforceDocumentTimestampByAction(document, model.SalesforceDocumentUpdated)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get last modified timestamp.")
		return "", http.StatusInternalServerError, ""
	}

	createdEventName, updatedEventName := getGroupEventName(document.Type)

	var processEventNames []string
	var processEventTimestamps []int64
	var groupUserID string
	if document.Action == model.SalesforceDocumentCreated {
		groupUserID, err = store.GetStore().CreateOrUpdateGroupPropertiesBySource(projectID, groupName, groupID, "",
			enProperties, createdTimestamp, lastModifiedTimestamp, model.SmartCRMEventSourceSalesforce)
		if err != nil {
			logCtx.WithError(err).Error("Failed to update salesforce group.")
			return "", http.StatusInternalServerError, ""
		}

		errCode := store.GetStore().UpdateSalesforceDocumentBySyncStatus(projectID, document, "", "", groupUserID, false)
		if errCode != http.StatusAccepted {
			logCtx.Error("Failed to set group_user_id in salesforce created document.")
			return "", http.StatusInternalServerError, ""
		}

		processEventNames = append(processEventNames, createdEventName, updatedEventName)
		processEventTimestamps = append(processEventTimestamps, createdTimestamp, createdTimestamp)

		if createdTimestamp != lastModifiedTimestamp {
			processEventNames = append(processEventNames, updatedEventName)
			processEventTimestamps = append(processEventTimestamps, lastModifiedTimestamp)
		}
	}

	if document.Action == model.SalesforceDocumentUpdated {
		documents, status := store.GetStore().GetSyncedSalesforceDocumentByType(projectID, []string{document.ID}, document.Type, true)
		if status != http.StatusFound {
			return "", http.StatusInternalServerError, ""
		}

		createdDocument := documents[0]

		groupUserID, err = store.GetStore().CreateOrUpdateGroupPropertiesBySource(projectID, groupName, groupID, createdDocument.GroupUserID,
			enProperties, createdTimestamp, lastModifiedTimestamp, model.SmartCRMEventSourceSalesforce)
		if err != nil {
			logCtx.WithError(err).Error("Failed to update salesforce group properties.")
			return "", http.StatusInternalServerError, ""
		}

		if createdDocument.GroupUserID == "" {
			errCode := store.GetStore().UpdateSalesforceDocumentBySyncStatus(projectID, &createdDocument, "", "", groupUserID, false)
			if errCode != http.StatusAccepted {
				logCtx.Error("Failed to update group_user_id in salesforce created document.")
				return "", http.StatusInternalServerError, ""
			}
		}

		errCode := store.GetStore().UpdateSalesforceDocumentBySyncStatus(projectID, document, "", "", groupUserID, false)
		if errCode != http.StatusAccepted {
			logCtx.Error("Failed to update group_user_id in salesforce updated document.")
			return "", http.StatusInternalServerError, ""
		}

		processEventNames = append(processEventNames, updatedEventName)
		processEventTimestamps = append(processEventTimestamps, lastModifiedTimestamp)
	}

	var eventID string
	for i := range processEventNames {
		trackPayload := &SDK.TrackPayload{
			Name:          processEventNames[i],
			ProjectId:     projectID,
			Timestamp:     processEventTimestamps[i],
			UserId:        groupUserID,
			RequestSource: model.UserSourceSalesforce,
		}

		status, response := SDK.Track(projectID, trackPayload, true, SDK.SourceSalesforce, "")
		if status != http.StatusOK && status != http.StatusFound && status != http.StatusNotModified {
			logCtx.WithFields(log.Fields{"status": status, "track_response": response, "event_name": processEventNames[i],
				"event_timestamp": processEventTimestamps[i]}).Error("Failed to track salesforce group event.")
			return "", http.StatusInternalServerError, ""
		}

		eventID = response.EventId
	}

	return groupUserID, http.StatusOK, eventID
}

func enrichGroupOpportunity(projectID int64, document *model.SalesforceDocument, salesforceSmartEventNames []SalesforceSmartEventName) (map[string]map[string]string, int) {
	logCtx := log.WithFields(log.Fields{"project_id": projectID, "document": document})
	if projectID == 0 || document == nil {
		logCtx.Error("Invalid project_id or document_id")
		return nil, http.StatusBadRequest
	}

	if document.Type != model.SalesforceDocumentTypeOpportunity {
		logCtx.Error("invalid document in group opportunities.")
		return nil, http.StatusInternalServerError
	}

	if document.GroupUserID != "" {
		// we skip opportunity processing if associated lead or contact record not processed.
		// Groups would have already processed this record
		logCtx.Error("Opportuntiy already processed for groups. Skipping record.")
		return nil, http.StatusOK
	}

	groupUserID, status, eventID := createOrUpdateSalesforceGroupsProperties(projectID, document, model.GROUP_NAME_SALESFORCE_OPPORTUNITY, document.ID)
	if status != http.StatusOK {
		logCtx.Error("Failed to create or update salesforce groups properties.")
		return nil, status
	}

	enProperties, properties, err := GetSalesforceDocumentProperties(projectID, document)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get properties")
		return nil, http.StatusInternalServerError
	}

	accountID := util.GetPropertyValueAsString((*enProperties)[model.GetCRMEnrichPropertyKeyByType(model.SmartCRMEventSourceSalesforce,
		model.GetSalesforceAliasByDocType(document.Type), "accountid")])
	if accountID != "" {
		// CreateSalesforceGroupRelationship will create two relationship
		err = CreateSalesforceGroupRelationship(projectID, model.GROUP_NAME_SALESFORCE_OPPORTUNITY, groupUserID,
			model.GROUP_NAME_SALESFORCE_ACCOUNT, accountID, model.SalesforceDocumentTypeAccount)
		if err != nil {
			logCtx.WithError(err).Error("Failed to create salesforce group relationship on CreateSalesforceGroupRelationship.")
		}
	} else {
		logCtx.Warning("No account id found for group relationship.")
	}

	oppLeadIds, oppContactIds, _, _, err := getOpportuntityLeadAndContactID(document)
	if err != nil {
		logCtx.WithError(err).Error("Failed to getOpportuntityLeadAndContactID on enrich group opportunity")
	}

	pendingSyncRecords := associateGroupUserOpportunitytoUser(projectID, oppLeadIds, oppContactIds, groupUserID)

	document.GroupUserID = groupUserID

	// Always use lastmodified timestamp for updated properties. Error handling already done during event creation
	lastModifiedTimestamp, _ := model.GetSalesforceDocumentTimestampByAction(document, model.SalesforceDocumentUpdated)

	var prevProperties *map[string]interface{}
	for _, smartEventName := range salesforceSmartEventNames {
		prevProperties = TrackSalesforceSmartEvent(projectID, &smartEventName, eventID, document.ID, groupUserID,
			document.Type, properties, prevProperties, lastModifiedTimestamp, false)
	}

	errCode := store.GetStore().UpdateSalesforceDocumentBySyncStatus(projectID, document, eventID, "", groupUserID, true)
	if errCode != http.StatusAccepted {
		logCtx.Error("Failed to update salesforce opportunity document as synced.")
		return nil, http.StatusInternalServerError
	}

	return pendingSyncRecords, http.StatusOK
}

// CreateSalesforceGroupRelationship create two way group relationship in group relationship table
func CreateSalesforceGroupRelationship(projectID int64, leftGroupName, leftgroupUserID, rightGroupName, rightDocID string, rightDocTyp int) error {
	logCtx := log.WithFields(log.Fields{"project_id": projectID, "left_group_user_id": leftgroupUserID, "right_document_id": rightDocID})
	if projectID == 0 || leftgroupUserID == "" || rightDocID == "" {
		logCtx.Error("Missing required fields")
		return errors.New("missing required fields")
	}

	documents, status := store.GetStore().GetSyncedSalesforceDocumentByType(projectID, []string{rightDocID}, rightDocTyp, true)
	if status != http.StatusFound || len(documents) < 1 {
		logCtx.Error("Failed get document on CreateSalesforceGroupRelationship")
		return errors.New("failed to get document on CreateSalesforceGroupRelationship")
	}

	if documents[0].GroupUserID == "" {
		logCtx.Error("Group user id not exist in document")
		return errors.New("group user id not exist in document")
	}

	_, status = store.GetStore().CreateGroupRelationship(projectID, leftGroupName, leftgroupUserID, rightGroupName, documents[0].GroupUserID)
	if status != http.StatusCreated && status != http.StatusConflict {
		logCtx.Error("failed to create salesforce group relationship")
		return errors.New("failed to create salesforce group relationship")
	}

	// create two way relationship
	_, status = store.GetStore().CreateGroupRelationship(projectID, rightGroupName, documents[0].GroupUserID, leftGroupName, leftgroupUserID)
	if status != http.StatusCreated && status != http.StatusConflict {
		logCtx.Error("failed to create salesforce group relationship")
		return errors.New("failed to create salesforce group relationship")
	}

	return nil
}

func enrichOpportunityContactRoles(projectID int64, document *model.SalesforceDocument) int {
	logCtx := log.WithFields(log.Fields{"project_id": projectID, "document": document})
	if projectID == 0 || document == nil {
		return http.StatusBadRequest
	}

	if document.Type != model.SalesforceDocumentTypeOpportunityContactRole {
		return http.StatusInternalServerError
	}

	var properties map[string]interface{}
	err := json.Unmarshal(document.Value.RawMessage, &properties)
	if err != nil {
		logCtx.WithError(err).Error("Failed to unmarshal opportunity contact roles.")
		return http.StatusInternalServerError
	}

	oppID, contactID := U.GetPropertyValueAsString(properties["OpportunityId"]), U.GetPropertyValueAsString(properties["ContactId"])
	if oppID == "" || contactID == "" {
		logCtx.Error("Failed to get opportunity id or contact id.")
		return http.StatusInternalServerError
	}

	associatedUserID := ""
	groupUserID := ""
	documents, status := store.GetStore().GetSyncedSalesforceDocumentByType(projectID, []string{oppID}, model.SalesforceDocumentTypeOpportunity, true)
	if status != http.StatusFound && status != http.StatusNotFound {
		return http.StatusInternalServerError
	}

	if status == http.StatusFound {
		if documents[0].Synced == false {
			return http.StatusOK
		}

		groupUserID = documents[0].GroupUserID
		if groupUserID == "" {
			return http.StatusOK
		}

		documents, status = store.GetStore().GetSyncedSalesforceDocumentByType(projectID, []string{contactID}, model.SalesforceDocumentTypeContact, true)
		if status != http.StatusFound && status != http.StatusNotFound {
			return http.StatusInternalServerError
		}

		if status == http.StatusFound {
			if documents[0].Synced == false { // record not processed should be picked later for association
				return http.StatusOK
			}

			contactUserID := documents[0].UserID
			_, status = store.GetStore().UpdateUserGroup(projectID, contactUserID, model.GROUP_NAME_SALESFORCE_OPPORTUNITY, "", groupUserID)
			if status != http.StatusAccepted && status != http.StatusNotModified {
				log.WithFields(log.Fields{"project_id": projectID, "user_id": contactUserID, "group_user_id": groupUserID}).
					Error("Failed to update salesforce user group id for opportunity contact roles.")
				return http.StatusInternalServerError
			}

			if status == http.StatusAccepted {
				associatedUserID = contactUserID
			}
		}
	}

	status = store.GetStore().UpdateSalesforceDocumentBySyncStatus(projectID, document, "", associatedUserID, groupUserID, true)
	if status != http.StatusAccepted {
		logCtx.Error("Failed to update salesforce opportunity document as synced.")
		return http.StatusInternalServerError
	}

	return http.StatusOK
}

func updateSalesforceUserAccountGroups(projectID int64, accountID, userID string) int {
	documents, status := store.GetStore().GetSyncedSalesforceDocumentByType(projectID, []string{accountID},
		model.SalesforceDocumentTypeAccount, true)
	if status != http.StatusFound {
		if status == http.StatusNotFound {
			return http.StatusOK // return ok if account was never capture
		}
		return http.StatusInternalServerError
	}

	groupUserID := documents[0].GroupUserID
	if groupUserID == "" {
		// update old record group status
		status := enrichGroupAccount(projectID, &documents[0], nil)
		if status != http.StatusOK || documents[0].GroupUserID == "" {
			log.WithFields(log.Fields{"project_id": projectID, "account_id": accountID}).
				Error("Failed to create group user on existing sync record.")
			return status
		}

		groupUserID = documents[0].GroupUserID
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

func enrichLeads(projectID int64, document *model.SalesforceDocument, salesforceSmartEventNames []SalesforceSmartEventName, pendingOpportunityAssociatedRecords map[string]string) int {
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
		RequestSource:   model.UserSourceSalesforce,
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
			properties, prevProperties, lastModifiedTimestamp, false)
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

		if groupUserID, exist := pendingOpportunityAssociatedRecords[document.ID]; exist {
			_, status := store.GetStore().UpdateUserGroup(projectID, userID, model.GROUP_NAME_SALESFORCE_OPPORTUNITY, "", groupUserID)
			if status != http.StatusAccepted && status != http.StatusNotModified {
				logCtx.Error("Failed associating contact user with group opportunity.")
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

func enrichCampaignToAllCampaignMembers(project *model.Project, otpRules *[]model.OTPRule, document *model.SalesforceDocument, endTimestamp int64) int {
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
				return http.StatusOK
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
			RequestSource:   model.UserSourceSalesforce,
		}

		eventID, userID, finalTrackPayload, err := TrackSalesforceEventByDocumentType(project.ID, trackPayload, &referenceDocument, "", "")
		if err != nil {
			logCtx.WithField("member_id", referenceDocument.ID).WithError(err).Error(
				"Failed to track salesforce campaign member update on campaign update.")
			return http.StatusInternalServerError
		}

		if memberDocuments[i].Synced == false {
			err = ApplySFOfflineTouchPointRule(project, otpRules, &finalTrackPayload, &memberDocuments[i], endTimestamp)
			if err != nil {
				// log and continue
				logCtx.WithField("EventID", eventID).WithField("userID", eventID).WithField("userID", eventID).Warn("failed creating SF offline touch point")
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
func getExistingCampaignMemberUserIDFromProperties(projectID int64, properties *map[string]interface{}) string {
	existingContactMemberID := (*properties)[model.GetCRMEnrichPropertyKeyByType(model.SmartCRMEventSourceSalesforce, model.SalesforceDocumentTypeNameCampaignMember, "ContactId")]
	existingLeadMemberID := (*properties)[model.GetCRMEnrichPropertyKeyByType(model.SmartCRMEventSourceSalesforce, model.SalesforceDocumentTypeNameCampaignMember, "LeadId")]

	logCtx := log.WithFields(log.Fields{"project_id": projectID, "contact_id": existingContactMemberID, "lead_id": existingLeadMemberID, "properties": properties})
	// use contact Id associated user id. Once user converts from lead to contact, salesforce prioritize contact based identification
	if existingContactMemberID != "" {
		// pull unsynced records along with synced for silently skip processing if contact not processed
		existingMember, status := store.GetStore().GetSyncedSalesforceDocumentByType(projectID, []string{util.GetPropertyValueAsString(existingContactMemberID)}, model.SalesforceDocumentTypeContact, true)
		if status != http.StatusFound && status != http.StatusNotFound {
			logCtx.WithFields(log.Fields{"err_code": status}).Error("Failed to get salesforce contact for campaign member enrichment")
			return ""
		}

		if status == http.StatusFound {
			if existingMember[0].Synced == false || existingMember[0].UserID == "" {
				logCtx.Warning("Contact record not processed yet for campaign member.")
				return ""
			}

			return existingMember[0].UserID
		}
	}

	existingMember, status := store.GetStore().GetSyncedSalesforceDocumentByType(projectID, []string{util.GetPropertyValueAsString(existingLeadMemberID)}, model.SalesforceDocumentTypeLead, true)
	if status != http.StatusFound && status != http.StatusNotFound {
		logCtx.WithFields(log.Fields{"err_code": status}).Error("Failed to get salesforce lead for campaign member enrichment")
		return ""
	}

	if status == http.StatusFound {
		if existingMember[0].Synced == false || existingMember[0].UserID == "" {
			logCtx.Warning("Lead record not processed yet for campaign member.")
			return ""
		}
		return existingMember[0].UserID
	}

	logCtx.Warning("Failed to get campaign member associated lead and contact record.")

	return ""
}

func enrichCampaignMemberResponded(project *model.Project, document *model.SalesforceDocument, userID string, campaignMemberProperties map[string]interface{}) int {
	logCtx := log.WithFields(log.Fields{"project_id": project.ID, "document_id": document.ID})
	if document.Type != model.SalesforceDocumentTypeCampaignMember {
		return http.StatusBadRequest
	}

	currentCampaignMemberDocumentProperties, _, err := GetSalesforceDocumentProperties(project.ID, document)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get properties for current campaign member.")
		return http.StatusInternalServerError
	}

	var currentCampaignMemberResponded bool
	if currentCampaignMemberDocumentProperties == nil || (*currentCampaignMemberDocumentProperties)[model.EP_SFCampaignMemberResponded] == nil {
		logCtx.Info("FirstResponsedDate not set in current campaign member.")
		return http.StatusOK
	}

	currentCampaignMemberResponded = (*currentCampaignMemberDocumentProperties)[model.EP_SFCampaignMemberResponded].(bool)

	var currentCampaignMemberFirstRespondedDateAsTimestamp int64
	if currentCampaignMemberResponded {
		currentCampaignMemberFirstRespondedDateAsTimestamp = (*currentCampaignMemberDocumentProperties)[model.EP_SFCampaignMemberFirstRespondedDate].(int64)
	}

	previousCampaignMemberDocuments, errCode := store.GetStore().GetLatestSalesforceDocumentByID(project.ID, []string{document.ID}, model.SalesforceDocumentTypeCampaignMember, document.Timestamp-1)
	if errCode != http.StatusFound && errCode != http.StatusNotFound {
		logCtx.Error("Failed to get previous campaign members.")
		return http.StatusInternalServerError
	}

	var previousCampaignMemberDocumentProperties *map[string]interface{}
	if previousCampaignMemberDocuments != nil {
		previousCampaignMemberDocumentProperties, _, err = GetSalesforceDocumentProperties(document.ProjectID,
			&previousCampaignMemberDocuments[len(previousCampaignMemberDocuments)-1])
		if err != nil {
			logCtx.WithError(err).Error("Failed to get properties for previous campaign member.")
			return http.StatusInternalServerError
		}
	}

	var previousCampaignMemberResponded bool
	if previousCampaignMemberDocumentProperties == nil || (*previousCampaignMemberDocumentProperties)[model.EP_SFCampaignMemberResponded] == nil {
		logCtx.Info("FirstResponsedDate not set in previous campaign member.")
		previousCampaignMemberResponded = false
	} else {
		previousCampaignMemberResponded = (*previousCampaignMemberDocumentProperties)[model.EP_SFCampaignMemberResponded].(bool)
	}

	var previousCampaignMemberFirstRespondedDateAsTimestamp int64
	if previousCampaignMemberResponded {
		previousCampaignMemberFirstRespondedDateAsTimestamp = (*previousCampaignMemberDocumentProperties)[model.EP_SFCampaignMemberFirstRespondedDate].(int64)
	}

	if currentCampaignMemberFirstRespondedDateAsTimestamp > 0 && previousCampaignMemberFirstRespondedDateAsTimestamp == 0 {
		timestamp := U.GetEndOfDayTimestampIn(currentCampaignMemberFirstRespondedDateAsTimestamp, document.GetDocumentTimeZone())

		trackPayload := &SDK.TrackPayload{
			ProjectId:       project.ID,
			Name:            U.EVENT_NAME_SALESFORCE_CAMPAIGNMEMBER_RESPONDED_TO_CAMPAIGN,
			Timestamp:       timestamp,
			RequestSource:   model.UserSourceSalesforce,
			UserId:          userID,
			EventProperties: campaignMemberProperties,
		}

		status, response := SDK.Track(project.ID, trackPayload, true, SDK.SourceSalesforce, "")
		if status != http.StatusOK {
			logCtx.WithFields(log.Fields{"status": status, "track_response": response, "event_name": U.EVENT_NAME_SALESFORCE_CAMPAIGNMEMBER_RESPONDED_TO_CAMPAIGN,
				"event_timestamp": timestamp}).Error("Failed to track salesforce responded to campaign event.")
			return http.StatusInternalServerError
		}
	}

	return http.StatusOK
}

func enrichCampaignMember(project *model.Project, otpRules *[]model.OTPRule, document *model.SalesforceDocument, endTimestamp int64) int {
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
	campaignDocuments, status := store.GetStore().GetSyncedSalesforceDocumentByType(project.ID, []string{util.GetPropertyValueAsString(campaignID)}, model.SalesforceDocumentTypeCampaign, true)
	if status != http.StatusFound && len(campaignDocuments) < 1 {
		logCtx.Warn("Failed to get campaign document for campaign member.")
		return http.StatusInternalServerError
	}

	var associateCampaignUpdate *model.SalesforceDocument
	for i := range campaignDocuments { // pick the campaign update which is before the campaign member update
		if campaignDocuments[i].Timestamp <= document.Timestamp {
			associateCampaignUpdate = &campaignDocuments[i]
		} else {
			break
		}
	}

	if associateCampaignUpdate == nil { // use first document if non available
		associateCampaignUpdate = &campaignDocuments[0]
	}

	enCampaignProperties, _, err := GetSalesforceDocumentProperties(project.ID, associateCampaignUpdate)
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
		if existingUserID == "" {
			return http.StatusOK
		}
	}

	trackPayload := &SDK.TrackPayload{
		ProjectId:       project.ID,
		EventProperties: *enCampaignMemberProperties,
		UserId:          existingUserID,
		RequestSource:   model.UserSourceSalesforce,
	}

	eventID, userID, finalTrackPayload, err := TrackSalesforceEventByDocumentType(project.ID, trackPayload, document, "", "")
	if err != nil {
		logCtx.WithError(err).Error("Failed to track salesforce lead event.")
		return http.StatusInternalServerError
	}

	err = ApplySFOfflineTouchPointRule(project, otpRules, &finalTrackPayload, document, endTimestamp)
	if err != nil {
		// log and continue
		logCtx.WithField("EventID", eventID).WithField("userID", userID).WithField("error", err).Warn("Create SF offline touch point")
	}

	errCode := enrichCampaignMemberResponded(project, document, userID, *enCampaignMemberProperties)
	if errCode != http.StatusOK {
		logCtx.Error("Failed to enrich Responded to Campaign event.")
	}

	errCode = store.GetStore().UpdateSalesforceDocumentBySyncStatus(project.ID, document, eventID, userID, "", true)
	if errCode != http.StatusAccepted {
		logCtx.Error("Failed to update salesforce lead document as synced.")
		return http.StatusInternalServerError
	}
	return http.StatusOK
}

var activtiesErrNoWhoIdAssociated = errors.New("No WhoId associated with activities document.")

func GetActivitiesUserID(document *model.SalesforceDocument) (string, error) {
	logCtx := log.WithFields(log.Fields{"project_id": document.ProjectID, "doc_id": document.ID})

	if document.Type != model.SalesforceDocumentTypeTask && document.Type != model.SalesforceDocumentTypeEvent {
		return "", errors.New("Invalid docType for salesforce activities.")
	}
	logCtx.WithField("doc_type", document.Type)

	var relationshipActivityRecord RelationshipActivityRecord
	err := json.Unmarshal(document.Value.RawMessage, &relationshipActivityRecord)
	if err != nil {
		return "", err
	}

	if relationshipActivityRecord.Who.ID == "" {
		logCtx.Warning("No WhoId associated with activities document.")
		return "", activtiesErrNoWhoIdAssociated
	}

	var relationshipActivityType int
	if relationshipActivityRecord.Who.Type == U.CapitalizeFirstLetter(model.SalesforceDocumentTypeNameLead) {
		relationshipActivityType = model.SalesforceDocumentTypeLead
	} else if relationshipActivityRecord.Who.Type == U.CapitalizeFirstLetter(model.SalesforceDocumentTypeNameContact) {
		relationshipActivityType = model.SalesforceDocumentTypeContact
	} else {
		logCtx.WithFields(log.Fields{"associated_doc_id": relationshipActivityRecord.Who.ID, "associated_doc_type": relationshipActivityRecord.Who.Type}).Warning("Invalid docType associated with activities document.")
		return "", errors.New("Invalid docType associated with activities document.")
	}

	doc, status := store.GetStore().GetSalesforceDocumentByTypeAndAction(document.ProjectID, relationshipActivityRecord.Who.ID, relationshipActivityType, model.SalesforceDocumentCreated)
	if status != http.StatusFound {
		return "", errors.New("Failed to get salesforce lead documents on GetActivitiesLeadOrContactId.")
	}

	return doc.UserID, nil
}

func canProcessSalesforceActivity(document *model.SalesforceDocument) (bool, error) {
	if document.Type != model.SalesforceDocumentTypeTask && document.Type != model.SalesforceDocumentTypeEvent {
		return false, errors.New("Invalid doc type in canSkipProcessingSalesforceActivity")
	}

	if document.Action != model.SalesforceDocumentCreated && document.Action != model.SalesforceDocumentUpdated {
		return false, errors.New("Invalid doc action in canSkipProcessingSalesforceActivity")
	}

	if document.Action == model.SalesforceDocumentCreated {
		return true, nil
	}

	var properties map[string]interface{}
	err := json.Unmarshal(document.Value.RawMessage, &properties)
	if err != nil {
		return false, err
	}

	var updatedDocHasWhoID bool
	if U.GetPropertyValueAsString(properties["WhoId"]) != "" {
		updatedDocHasWhoID = true
	}

	docs, errCode := store.GetStore().GetSyncedSalesforceDocumentByType(document.ProjectID, []string{document.ID}, document.Type, true)
	if errCode != http.StatusFound {
		return false, errors.New("Failed to get synced docs on canProcessSalesforceActivity")
	}

	var createdDocProperties map[string]interface{}
	err = json.Unmarshal(docs[0].Value.RawMessage, &createdDocProperties)
	if err != nil {
		return false, err
	}

	var createdDocHasWhoID bool
	if U.GetPropertyValueAsString(createdDocProperties["WhoId"]) != "" {
		createdDocHasWhoID = true
	}

	// Skip Processing if created_doc doesn't contain WhoId and updated_doc contains WhoId
	if !createdDocHasWhoID && updatedDocHasWhoID {
		return false, nil
	}

	return true, nil
}

func enrichTask(project *model.Project, otpRules *[]model.OTPRule, uniqueOTPEventKeys *[]string, document *model.SalesforceDocument) int {
	logCtx := log.WithField("project_id", project.ID).WithField("document_id", document.ID)

	if project.ID == 0 || document == nil {
		logCtx.Error("Invalid parameters in enrich Task")
		return http.StatusBadRequest
	}

	if document.Type != model.SalesforceDocumentTypeTask {
		logCtx.Error("Invalid docType in enrich Task")
		return http.StatusInternalServerError
	}

	if document.Action == model.SalesforceDocumentUpdated {
		canProcess, err := canProcessSalesforceActivity(document)
		if err != nil {
			logCtx.WithError(err).Error("Failed in canProcessSalesforceActivity on enrich Task")
			return http.StatusInternalServerError
		}

		if !canProcess {
			return http.StatusOK
		}
	}

	enProperties, _, err := GetSalesforceDocumentProperties(project.ID, document)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get properties")
		return http.StatusInternalServerError
	}

	activityUserID, err := GetActivitiesUserID(document)
	if err != nil {
		if err == activtiesErrNoWhoIdAssociated {
			logCtx.WithError(err).Info("Skipping processing for task record.")
			return http.StatusOK
		}
		logCtx.WithError(err).Error("Failed to GetActivitiesUserID on enrich task")
		return http.StatusInternalServerError
	}

	if activityUserID == "" {
		logCtx.Error("Lead or contact associated is not synced for processing task document.")
		return http.StatusOK
	}

	trackPayload := &SDK.TrackPayload{
		ProjectId:       project.ID,
		UserId:          activityUserID,
		EventProperties: *enProperties,
		RequestSource:   model.UserSourceSalesforce,
	}

	eventID, userID, finalPayload, err := TrackSalesforceEventByDocumentType(project.ID, trackPayload, document, "", model.SalesforceDocumentTypeNameTask)
	if err != nil {
		logCtx.WithError(err).Error("Failed to track salesforce task event.")
		return http.StatusInternalServerError
	}

	err = ApplySFOfflineTouchPointRuleForTasks(project, otpRules, uniqueOTPEventKeys, &finalPayload, document)
	if err != nil {
		// log and continue
		logCtx.WithField("EventID", eventID).WithField("userID", userID).WithField("error", err).Warn("Failed creating offline touch point event for SF Tasks")
	}

	errCode := store.GetStore().UpdateSalesforceDocumentBySyncStatus(project.ID, document, eventID, userID, "", true)
	if errCode != http.StatusAccepted {
		logCtx.Error("Failed to update salesforce task document as synced.")
		return http.StatusInternalServerError
	}

	return http.StatusOK
}

func enrichEvent(project *model.Project, otpRules *[]model.OTPRule, uniqueOTPEventKeys *[]string, document *model.SalesforceDocument) int {
	logCtx := log.WithField("project_id", project.ID).WithField("document_id", document.ID)

	if project.ID == 0 || document == nil {
		logCtx.Error("Invalid parameters in enrich Event")
		return http.StatusBadRequest
	}

	if document.Type != model.SalesforceDocumentTypeEvent {
		logCtx.Error("Invalid docType in enrich Event")
		return http.StatusInternalServerError
	}

	if document.Action == model.SalesforceDocumentUpdated {
		canProcess, err := canProcessSalesforceActivity(document)
		if err != nil {
			logCtx.WithError(err).Error("Failed in canProcessSalesforceActivity on enrich Event")
			return http.StatusInternalServerError
		}

		if !canProcess {
			return http.StatusOK
		}
	}

	enProperties, _, err := GetSalesforceDocumentProperties(project.ID, document)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get properties")
		return http.StatusInternalServerError
	}

	activityUserID, err := GetActivitiesUserID(document)
	if err != nil {
		if err == activtiesErrNoWhoIdAssociated {
			logCtx.WithError(err).Warning("Skipping processing for event record.")
			return http.StatusOK
		}
		logCtx.WithError(err).Error("Failed to GetActivitiesUserID on enrich event")
		return http.StatusInternalServerError
	}

	if activityUserID == "" {
		logCtx.Error("Lead or contact associated is not synced for processing event document.")
		return http.StatusOK
	}

	trackPayload := &SDK.TrackPayload{
		ProjectId:       project.ID,
		UserId:          activityUserID,
		EventProperties: *enProperties,
		RequestSource:   model.UserSourceSalesforce,
	}

	eventID, userID, finalPayload, err := TrackSalesforceEventByDocumentType(project.ID, trackPayload, document, "", model.SalesforceDocumentTypeNameEvent)
	if err != nil {
		logCtx.WithError(err).Error("Failed to track salesforce event event.")
		return http.StatusInternalServerError
	}

	err = ApplySFOfflineTouchPointRuleForEvents(project, otpRules, uniqueOTPEventKeys, &finalPayload, document)
	if err != nil {
		// log and continue
		logCtx.WithField("EventID", eventID).WithField("userID", userID).WithField("error", err).Warn("Failed creating offline touch point event for SF Events")
	}

	errCode := store.GetStore().UpdateSalesforceDocumentBySyncStatus(project.ID, document, eventID, userID, "", true)
	if errCode != http.StatusAccepted {
		logCtx.Error("Failed to update salesforce event document as synced.")
		return http.StatusInternalServerError
	}

	return http.StatusOK
}

//Check if the condition are satisfied for creating OTP events for each rule for SF Tasks Updated.
func ApplySFOfflineTouchPointRuleForTasks(project *model.Project, otpRules *[]model.OTPRule, uniqueOTPEventKeys *[]string, trackPayload *SDK.TrackPayload, document *model.SalesforceDocument) error {

	logCtx := log.WithFields(log.Fields{"project_id": project.ID, "method": "ApplySFOfflineTouchPointRuleForTasks",
		"document_id": document.ID, "document_action": document.Action, "document": document})

	if otpRules == nil || project == nil || trackPayload == nil || document == nil {
		return nil
	}
	for _, rule := range *otpRules {

		otpUniqueKey, err := createOTPUniqueKeyForTasks(rule, trackPayload, logCtx)
		if err != http.StatusCreated {
			logCtx.Error("Failed to create otp_unique_key")
			continue
		}

		//Check if rule type is sf_tasks
		if rule.RuleType != model.TouchPointRuleTypeTasks {
			continue
		}

		// check if rule is applicable w.r.t filters
		if !filterCheck(rule, trackPayload, logCtx) {
			continue
		}

		//Checks if the otpUniqueKey is already present in other OTP Event Properties
		if !isSalesforceOTPKeyUnique(otpUniqueKey, uniqueOTPEventKeys) {
			continue
		}

		_, err1 := CreateTouchPointEventForTasksAndEvents(project, trackPayload, document, rule, otpUniqueKey)
		if err1 != nil {
			logCtx.WithError(err1).Error("failed to create touch point for salesforce tasks.")
			continue
		}

	}
	return nil
}

//Check if the condition are satisfied for creating OTP events for each rule for SF Event Updated.
func ApplySFOfflineTouchPointRuleForEvents(project *model.Project, otpRules *[]model.OTPRule, uniqueOTPEventKeys *[]string, trackPayload *SDK.TrackPayload, document *model.SalesforceDocument) error {

	logCtx := log.WithFields(log.Fields{"project_id": project.ID, "method": "ApplySFOfflineTouchPointRuleForEvents",
		"document_id": document.ID, "document_action": document.Action, "document": document})

	if otpRules == nil || project == nil || trackPayload == nil || document == nil {
		return nil
	}
	for _, rule := range *otpRules {

		otpUniqueKey, err := createOTPUniqueKeyForEvents(rule, trackPayload, logCtx)
		if err != http.StatusCreated {
			logCtx.Error("Failed to create otp_unique_key")
			continue
		}

		//Check if rule type is sf_events
		if rule.RuleType != model.TouchPointRuleTypeEvents {
			continue
		}

		// check if rule is applicable w.r.t filters
		if !filterCheck(rule, trackPayload, logCtx) {
			continue
		}

		//Checks if the otpUniqueKey is already present in other OTP Event Properties
		if !isSalesforceOTPKeyUnique(otpUniqueKey, uniqueOTPEventKeys) {
			continue
		}

		_, err1 := CreateTouchPointEventForTasksAndEvents(project, trackPayload, document, rule, otpUniqueKey)
		if err1 != nil {
			logCtx.WithError(err1).Error("failed to create touch point for salesforce events.")
			continue
		}

	}
	return nil
}

//Check if the condition are satisfied for creating OTP events for each rule for SF Campaign.
func ApplySFOfflineTouchPointRule(project *model.Project, otpRules *[]model.OTPRule, trackPayload *SDK.TrackPayload, document *model.SalesforceDocument, endTimestamp int64) error {

	logCtx := log.WithFields(log.Fields{"project_id": project.ID, "method": "ApplySFOfflineTouchPointRule",
		"document_id": document.ID, "document_action": document.Action, "document": document})

	if otpRules == nil || project == nil || trackPayload == nil || document == nil {
		return nil
	}

	var campaignMemberDocuments []model.SalesforceDocument
	var status int
	fistRespondedRuleApplicable := true
	// Checking if the EP_SFCampaignMemberResponded has already been set as true for same customer id
	if document.Action == model.SalesforceDocumentUpdated {

		campaignMemberDocuments, status = store.GetStore().GetSyncedSalesforceDocumentByType(project.ID,
			[]string{util.GetPropertyValueAsString(document.ID)}, model.SalesforceDocumentTypeCampaignMember, false)
		if status != http.StatusFound {
			campaignMemberDocuments = nil
		}
		// Ignore responded event for "Created event" based rule.
		if campaignMemberDocuments != nil || len(campaignMemberDocuments) != 0 {

			logCtx.WithField("Total_Documents", len(campaignMemberDocuments)).
				WithField("Document[size-1]", campaignMemberDocuments[len(campaignMemberDocuments)-1]).
				Info("Found existing campaign member document")

			// len(campaignMemberDocuments) > 0 && timestamp sorted desc
			enCampaignMemberProperties, _, err := GetSalesforceDocumentProperties(project.ID, &campaignMemberDocuments[len(campaignMemberDocuments)-1])

			if err == nil {
				// ignore to create a new touch point if last updated doc has EP_SFCampaignMemberResponded=true
				if val, exists := (*enCampaignMemberProperties)[model.EP_SFCampaignMemberResponded]; exists {
					if val != nil && val.(bool) == true {
						logCtx.Info("found EP_SFCampaignMemberResponded=true for the document, skipping creating OTP.")
						fistRespondedRuleApplicable = false
					}
				}
			} else {
				logCtx.WithError(err).Error("Failed to get properties for salesforce campaign member, continuing")
			}

		}
	}

	for _, rule := range *otpRules {

		// check if rule is applicable
		if !filterCheck(rule, trackPayload, logCtx) {
			continue
		}

		// Run for create document rule
		if rule.TouchPointTimeRef == model.SFCampaignMemberCreated && document.Action == model.SalesforceDocumentCreated {
			_, err := CreateTouchPointEvent(project, trackPayload, document, rule)
			if err != nil {
				logCtx.WithError(err).Error("failed to create touch point for salesforce campaign member document. trying for responded rule")
			}
		}

		// Run for only first responded rules & documents where first responded is not set.
		if rule.TouchPointTimeRef == model.SFCampaignMemberResponded && fistRespondedRuleApplicable {

			logCtx.Info("Found existing salesforce campaign member document")
			if val, exists := trackPayload.EventProperties[model.EP_SFCampaignMemberResponded]; exists {
				if val.(bool) == true {
					_, err := CreateTouchPointEvent(project, trackPayload, document, rule)
					if err != nil {
						logCtx.WithError(err).Error("failed to create touch point for salesforce campaign member document.")
					}
				}
			}
		}
	}
	return nil
}

// CreateTouchPointEventForTasksAndEvents - Creates offline touchpoint for SF update events with given rule for SF Tasks/Events
func CreateTouchPointEventForTasksAndEvents(project *model.Project, trackPayload *SDK.TrackPayload, document *model.SalesforceDocument,
	rule model.OTPRule, otpUniqueKey string) (*SDK.TrackResponse, error) {

	logCtx := log.WithFields(log.Fields{"project_id": project.ID, "method": "CreateTouchPointEvent", "document_id": document.ID, "document_action": document.Action})
	logCtx.WithField("document", document).WithField("trackPayload", trackPayload).Info("CreateTouchPointEventForTasksAndEvents: creating salesforce OFFLINE TOUCH POINT document")
	var trackResponse *SDK.TrackResponse
	var err error
	eventProperties := make(U.PropertiesMap, 0)
	payload := &SDK.TrackPayload{
		ProjectId:       project.ID,
		EventProperties: eventProperties,
		UserId:          trackPayload.UserId,
		Name:            U.EVENT_NAME_OFFLINE_TOUCH_POINT,
		RequestSource:   trackPayload.RequestSource,
	}

	var timestamp int64
	timeValue, exists := (trackPayload.EventProperties)[rule.TouchPointTimeRef]
	if !exists {
		logCtx.WithField("TouchPointTimeRef", rule.TouchPointTimeRef).
			Error("couldn't get the timestamp on hubspot contact properties using given rule.TouchPointTimeRef")
		return nil, errors.New(fmt.Sprintf("couldn't get the timestamp on hubspot contact properties "+
			"using given rule.TouchPointTimeRef - %s", rule.TouchPointTimeRef))
	}
	val, ok := timeValue.(int64)
	if !ok {
		logCtx.WithField("TouchPointTimeRef", rule.TouchPointTimeRef).WithField("TimeValue", timeValue).
			Error("couldn't convert the timestamp on hubspot contact properties. using trackPayload timestamp instead, val")
		timestamp = trackPayload.Timestamp
	} else {
		timestamp = val
	}

	// Adding mandatory properties
	payload.EventProperties[U.EP_OTP_RULE_ID] = rule.ID
	payload.EventProperties[U.EP_OTP_UNIQUE_KEY] = otpUniqueKey
	payload.Timestamp = timestamp

	// Mapping touch point properties:
	var rulePropertiesMap map[string]model.TouchPointPropertyValue
	err = U.DecodePostgresJsonbToStructType(&rule.PropertiesMap, &rulePropertiesMap)
	if err != nil {
		logCtx.WithField("Document", trackPayload).WithError(err).Error("Failed to decode/fetch offline touch point rule PROPERTIES for salesforce document.")
		return trackResponse, errors.New(fmt.Sprintf("create salesforce touchpoint event track failed in method CreateTouchPointEventForTasksAndEvents for doc type %d, message %s", document.Type, trackResponse.Error))
	}

	for key, value := range rulePropertiesMap {

		if value.Type == model.TouchPointPropertyValueAsConstant {
			payload.EventProperties[key] = value.Value
		} else {
			if _, exists := trackPayload.EventProperties[value.Value]; exists {
				payload.EventProperties[key] = trackPayload.EventProperties[value.Value]
			} else {
				// Property value is not found, hence keeping it as $none
				payload.EventProperties[key] = model.PropertyValueNone
			}
		}
	}

	status, trackResponse := SDK.Track(project.ID, payload, true, SDK.SourceSalesforce, "")
	if status != http.StatusOK && status != http.StatusFound && status != http.StatusNotModified {
		logCtx.WithField("Document", trackPayload).WithError(err).Error(fmt.Errorf("create salesforce OTP event track failed for doc type %d, message %s", document.Type, trackResponse.Error))
		return trackResponse, errors.New(fmt.Sprintf("create salesforce touchpoint event track failed in method CreateTouchPointEventForTasksAndEvents for doc type %d, message %s", document.Type, trackResponse.Error))
	}
	logCtx.WithField("statusCode", status).WithField("trackResponsePayload", trackResponse).Info("Successfully: created salesforce offline touch point")
	return trackResponse, nil

}

//CreateTouchPointEvent- Creates offline touch point event for SF Campaign
func CreateTouchPointEvent(project *model.Project, trackPayload *SDK.TrackPayload, document *model.SalesforceDocument, rule model.OTPRule) (*SDK.TrackResponse, error) {

	logCtx := log.WithFields(log.Fields{"project_id": project.ID, "method": "CreateTouchPointEvent", "document_id": document.ID, "document_action": document.Action})
	logCtx.WithField("document", document).WithField("trackPayload", trackPayload).Info("CreateTouchPointEvent: creating salesforce document")
	var trackResponse *SDK.TrackResponse
	var err error
	eventProperties := make(U.PropertiesMap, 0)
	payload := &SDK.TrackPayload{
		ProjectId:       project.ID,
		EventProperties: eventProperties,
		UserId:          trackPayload.UserId,
		Name:            U.EVENT_NAME_OFFLINE_TOUCH_POINT,
		RequestSource:   trackPayload.RequestSource,
	}

	var timestamp int64
	timestamp, err = model.GetSalesforceDocumentTimestampByAction(document, model.SalesforceDocumentUpdated)
	if err != nil {
		logCtx.WithError(err).Error("failed to timestamp for SF for offline touch point.")
		return trackResponse, err
	}
	payload.Timestamp = timestamp

	if rule.TouchPointTimeRef == model.SFCampaignMemberResponded {
		if val, exists := trackPayload.EventProperties[model.EP_SFCampaignMemberFirstRespondedDate]; exists {
			if tt, ok := val.(int64); ok {
				payload.Timestamp = tt
			} else {
				logCtx.WithError(err).Error("failed to set timestamp for SF for offline touch point - First responded time.")
			}
		}
	}

	// Mapping touch point properties:
	var rulePropertiesMap map[string]model.TouchPointPropertyValue
	err = U.DecodePostgresJsonbToStructType(&rule.PropertiesMap, &rulePropertiesMap)
	if err != nil {
		logCtx.WithField("Document", trackPayload).WithError(err).Error("Failed to decode/fetch offline touch point rule PROPERTIES for salesforce document.")
		return trackResponse, errors.New(fmt.Sprintf("create salesforce touchpoint event track failed for doc type %d, message %s", document.Type, trackResponse.Error))
	}

	for key, value := range rulePropertiesMap {

		if value.Type == model.TouchPointPropertyValueAsConstant {
			payload.EventProperties[key] = value.Value
		} else {
			if _, exists := trackPayload.EventProperties[value.Value]; exists {
				payload.EventProperties[key] = trackPayload.EventProperties[value.Value]
			} else {
				// Property value is not found, hence keeping it as $none
				payload.EventProperties[key] = model.PropertyValueNone
			}
		}
	}
	// Adding mandatory properties
	payload.EventProperties[U.EP_OTP_RULE_ID] = rule.ID

	status, trackResponse := SDK.Track(project.ID, payload, true, SDK.SourceSalesforce, "")
	if status != http.StatusOK && status != http.StatusFound && status != http.StatusNotModified {
		logCtx.WithField("Document", trackPayload).WithError(err).Error(fmt.Errorf("create salesforce touchpoint event track failed for doc type %d, message %s", document.Type, trackResponse.Error))
		return trackResponse, errors.New(fmt.Sprintf("create salesforce touchpoint event track failed for doc type %d, message %s", document.Type, trackResponse.Error))
	}
	logCtx.WithField("statusCode", status).WithField("trackResponsePayload", trackResponse).Info("Successfully: created salesforce offline touch point")
	return trackResponse, nil
}

// Returns true or false if the otpKey (userID+ruleID+keyID) is not present in uniqueOTPEventKeys i.e. Unique OTP key.
func isSalesforceOTPKeyUnique(otpUniqueKey string, uniqueOTPEventKeys *[]string) bool {
	return U.StringValueIn(otpUniqueKey, *uniqueOTPEventKeys)
}

//Creates a unique key using ruleID, userID and salesforce task activity ID  as keyID for Salesforce Tasks.
func createOTPUniqueKeyForTasks(rule model.OTPRule, trackPayload *SDK.TrackPayload, logCtx *log.Entry) (string, int) {

	ruleID := rule.ID
	userID := trackPayload.UserId
	var keyID string
	var uniqueKey string

	if _, exists := trackPayload.EventProperties[U.EP_SF_TASK_ID]; exists {
		keyID = fmt.Sprintf("%v", trackPayload.EventProperties[U.EP_SF_TASK_ID])
	} else {
		logCtx.Error("Event Property $salesforce_task_id does not exist.")
		return uniqueKey, http.StatusNotFound
	}

	uniqueKey = userID + ruleID + keyID
	return uniqueKey, http.StatusCreated
}

//Creates a unique key using ruleID, userID and salesforce Event activity ID  as keyID for Salesforce Tasks.
func createOTPUniqueKeyForEvents(rule model.OTPRule, trackPayload *SDK.TrackPayload, logCtx *log.Entry) (string, int) {

	ruleID := rule.ID
	userID := trackPayload.UserId
	var keyID string
	var uniqueKey string

	if _, exists := trackPayload.EventProperties[U.EP_SF_EVENT_ID]; exists {
		keyID = fmt.Sprintf("%v", trackPayload.EventProperties[U.EP_SF_EVENT_ID])
	} else {
		logCtx.Error("Event Property $salesforce_event_id does not exist.")
		return uniqueKey, http.StatusNotFound
	}

	uniqueKey = userID + ruleID + keyID
	return uniqueKey, http.StatusCreated
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

//filterCheck- Returns true if all the filters applied are passed.
func filterCheck(rule model.OTPRule, trackPayload *SDK.TrackPayload, logCtx *log.Entry) bool {

	var ruleFilters []model.TouchPointFilter
	err := U.DecodePostgresJsonbToStructType(&rule.Filters, &ruleFilters)
	if err != nil {
		logCtx.WithField("Document", trackPayload).WithError(err).Error("Failed to decode/fetch offline touch point rule FILTERS for salesforce document.")
		return false
	}

	filtersPassed := 0
	for _, filter := range ruleFilters {
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
			logCtx.WithField("Rule", rule).WithField("TrackPayload", trackPayload).Error("No matching operator found for offline touch point rules for salesforce document.")
			continue
		}
	}
	return filtersPassed != 0 && filtersPassed == len(ruleFilters)
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
func enrichCampaign(project *model.Project, otpRules *[]model.OTPRule, document *model.SalesforceDocument, endTimestamp int64) int {
	if project.ID == 0 || document == nil {
		return http.StatusBadRequest
	}

	if document.Type == model.SalesforceDocumentTypeCampaign {
		return enrichCampaignToAllCampaignMembers(project, otpRules, document, endTimestamp)
	}

	if document.Type == model.SalesforceDocumentTypeCampaignMember {
		return enrichCampaignMember(project, otpRules, document, endTimestamp)
	}

	return http.StatusBadRequest
}

func enrichAll(project *model.Project, otpRules *[]model.OTPRule, uniqueOTPEventKeys *[]string, documents []model.SalesforceDocument, salesforceSmartEventNames []SalesforceSmartEventName,
	pendingOpportunityGroupAssociations map[string]map[string]string, endTimestamp int64) int {
	if project.ID == 0 {
		return http.StatusBadRequest
	}
	logCtx := log.WithField("project_id", project.ID)

	var seenFailures bool
	var errCode int
	for i := range documents {
		startTime := time.Now().Unix()
		switch documents[i].Type {
		case model.SalesforceDocumentTypeContact:
			errCode = enrichContact(project.ID, &documents[i], salesforceSmartEventNames, pendingOpportunityGroupAssociations[model.SalesforceDocumentTypeNameContact])
		case model.SalesforceDocumentTypeLead:
			errCode = enrichLeads(project.ID, &documents[i], salesforceSmartEventNames, pendingOpportunityGroupAssociations[model.SalesforceDocumentTypeNameLead])
		case model.SalesforceDocumentTypeCampaign, model.SalesforceDocumentTypeCampaignMember:
			errCode = enrichCampaign(project, otpRules, &documents[i], endTimestamp)
		case model.SalesforceDocumentTypeOpportunityContactRole:
			errCode = enrichOpportunityContactRoles(project.ID, &documents[i])
		case model.SalesforceDocumentTypeTask:
			errCode = enrichTask(project, otpRules, uniqueOTPEventKeys, &documents[i])
		case model.SalesforceDocumentTypeEvent:
			errCode = enrichEvent(project, otpRules, uniqueOTPEventKeys, &documents[i])
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
func GetSalesforceSmartEventNames(projectID int64) *map[string][]SalesforceSmartEventName {

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
		salesforceSmartEventName.EventNameID = eventNames[i].ID

		if _, exists := salesforceSmartEventNames[decFilterExp.ObjectType]; !exists {
			salesforceSmartEventNames[decFilterExp.ObjectType] = []SalesforceSmartEventName{}
		}

		salesforceSmartEventNames[decFilterExp.ObjectType] = append(salesforceSmartEventNames[decFilterExp.ObjectType], salesforceSmartEventName)
	}

	return &salesforceSmartEventNames
}

type enrichGroupWorkerStatus struct {
	HasFailure                bool
	OverAllPendingSyncRecords map[string]map[string]string
	Lock                      sync.Mutex
}

func updateGroupWorkerStatus(errCode int, pendingSyncRecords map[string]map[string]string, status *enrichGroupWorkerStatus) {
	status.Lock.Lock()
	defer status.Lock.Unlock()
	if errCode != http.StatusOK {
		status.HasFailure = true
	}

	if status.OverAllPendingSyncRecords == nil {
		status.OverAllPendingSyncRecords = make(map[string]map[string]string)
	}

	// capture pending record not synced for first time
	for docTypeAlias := range pendingSyncRecords {
		for docID := range pendingSyncRecords[docTypeAlias] {
			if _, exist := status.OverAllPendingSyncRecords[docTypeAlias]; !exist {
				status.OverAllPendingSyncRecords[docTypeAlias] = make(map[string]string)
			}

			if _, exist := status.OverAllPendingSyncRecords[docTypeAlias][docID]; !exist {
				status.OverAllPendingSyncRecords[docTypeAlias][docID] = pendingSyncRecords[docTypeAlias][docID]
			}
		}
	}
}

func enrichAllGroup(projectID int64, wg *sync.WaitGroup, docType int, smartEventNames []SalesforceSmartEventName, documents []model.SalesforceDocument, status *enrichGroupWorkerStatus) {
	defer wg.Done()
	for i := range documents {
		startTime := time.Now().Unix()

		var errCode int
		var pendingSyncRecords map[string]map[string]string
		switch documents[i].Type {
		case model.SalesforceDocumentTypeAccount, model.SalesforceDocumentTypeGroupAccount:
			errCode = enrichGroupAccount(projectID, &documents[i], smartEventNames)
		case model.SalesforceDocumentTypeOpportunity:
			pendingSyncRecords, errCode = enrichGroupOpportunity(projectID, &documents[i], smartEventNames)
		}

		updateGroupWorkerStatus(errCode, pendingSyncRecords, status)

		log.WithFields(log.Fields{"project_id": projectID, "doc_type": docType, "time_taken_in_secs": time.Now().Unix() - startTime, "doc_id": documents[i].ID}).
			Debug("Completed group document sync.")
	}
}

func enrichGroup(projectID int64, workerPerProject int, salesforceSmartEventNames *map[string][]SalesforceSmartEventName) (map[string]bool, map[string]map[string]string, int) {
	logCtx := log.WithFields(log.Fields{"project_id": projectID})

	if projectID == 0 {
		logCtx.Error("Invalid project id.")
		return nil, nil, http.StatusBadRequest
	}

	overAllSyncStatus := make(map[string]bool)
	overAllPendingSyncRecords := make(map[string]map[string]string)
	for _, docTypeAlias := range []string{model.SalesforceDocumentTypeNameAccount, model.SalesforceDocumentTypeNameOpportunity} {

		docType := model.GetSalesforceDocTypeByAlias(docTypeAlias)
		documents, errCode := store.GetStore().GetSalesforceDocumentsByTypeForSync(projectID, docType, 0, 0)
		if errCode != http.StatusFound {
			if errCode == http.StatusNotFound {
				continue
			}

			logCtx.Error("Failed to get salesforce account documents for groups.")
			return nil, nil, http.StatusInternalServerError
		}

		overAllSyncStatus[docTypeAlias] = false

		batches := GetBatchedOrderedDocumentsByID(documents, workerPerProject)
		var status enrichGroupWorkerStatus
		workerIndex := 0
		for i := range batches {
			batch := batches[i]
			var wg sync.WaitGroup
			for docID := range batch {
				logCtx.WithFields(log.Fields{"worker": workerIndex, "doc_id": docID, "type": docTypeAlias, "is_group": true}).Info("Processing Batch by doc_id")
				wg.Add(1)
				go enrichAllGroup(projectID, &wg, docType, (*salesforceSmartEventNames)[docTypeAlias], batch[docID], &status)
				workerIndex++
			}
			wg.Wait()
		}

		overAllSyncStatus[docTypeAlias] = status.HasFailure
		overAllPendingSyncRecords = status.OverAllPendingSyncRecords
	}

	return overAllSyncStatus, overAllPendingSyncRecords, http.StatusOK
}

func associateGroupUserOpportunitytoUser(projectID int64, oppLeadIds, oppContactIds []string, OpportunityGroupUserID string) map[string]map[string]string {
	pendingSyncRecords := make(map[string]map[string]string)
	for docTypeAlias, docIDs := range map[string][]string{model.SalesforceDocumentTypeNameLead: oppLeadIds, model.SalesforceDocumentTypeNameContact: oppContactIds} {
		docType := model.GetSalesforceDocTypeByAlias(docTypeAlias)
		for i := range docIDs {
			logCtx := log.WithFields(log.Fields{"project_id": projectID, "doc_type": docType, "doc_id": docIDs[i]})
			documets, status := store.GetStore().GetSyncedSalesforceDocumentByType(projectID, []string{docIDs[i]}, docType, true)
			if status != http.StatusFound {
				logCtx.Error("Failed to get synced salesforce document for opportunity group association.")
				continue
			}

			createdDocument := documets[0]
			if createdDocument.Synced == false {
				// only the first occurrence of associations
				if _, exist := pendingSyncRecords[docTypeAlias]; !exist {
					pendingSyncRecords[docTypeAlias] = make(map[string]string)
				}

				if _, exist := pendingSyncRecords[docTypeAlias][createdDocument.ID]; !exist {
					pendingSyncRecords[docTypeAlias][createdDocument.ID] = OpportunityGroupUserID
				}
				continue
			}

			_, status = store.GetStore().UpdateUserGroup(projectID, createdDocument.UserID, model.GROUP_NAME_SALESFORCE_OPPORTUNITY, "", OpportunityGroupUserID)
			if status != http.StatusAccepted && status != http.StatusNotModified {
				logCtx.Error("Failed associating user with group opportunity.")
			}
		}
	}

	return pendingSyncRecords
}

// GetBatchedOrderedDocumentsByID return list of document in batches. Order is maintained on document id.
func GetBatchedOrderedDocumentsByID(documents []model.SalesforceDocument, batchSize int) []map[string][]model.SalesforceDocument {

	if len(documents) < 0 {
		return nil
	}

	documentsMap := make(map[string][]model.SalesforceDocument)
	for i := range documents {
		if _, exist := documentsMap[documents[i].ID]; !exist {
			documentsMap[documents[i].ID] = make([]model.SalesforceDocument, 0)
		}
		documentsMap[documents[i].ID] = append(documentsMap[documents[i].ID], documents[i])
	}

	batchedDocumentsByID := make([]map[string][]model.SalesforceDocument, 1)
	isBatched := make(map[string]bool)
	batchLen := 0
	batchedDocumentsByID[batchLen] = make(map[string][]model.SalesforceDocument)
	for i := range documents {
		if isBatched[documents[i].ID] {
			continue
		}

		if len(batchedDocumentsByID[batchLen]) >= batchSize {
			batchedDocumentsByID = append(batchedDocumentsByID, make(map[string][]model.SalesforceDocument))
			batchLen++
		}

		batchedDocumentsByID[batchLen][documents[i].ID] = documentsMap[documents[i].ID]
		isBatched[documents[i].ID] = true
	}

	return batchedDocumentsByID
}

type enrichWorkerStatus struct {
	HasFailure bool
	Lock       sync.Mutex
}

func enrichAllWorker(project *model.Project, wg *sync.WaitGroup, enrichStatus *enrichWorkerStatus, otpRules *[]model.OTPRule,
	uniqueOTPEventKeys *[]string, documents []model.SalesforceDocument, smartEventNames []SalesforceSmartEventName,
	pendingOpportunityGroupAssociations map[string]map[string]string, timeRange int64) {
	defer wg.Done()
	errCode := enrichAll(project, otpRules, uniqueOTPEventKeys, documents, smartEventNames, pendingOpportunityGroupAssociations, timeRange)

	enrichStatus.Lock.Lock()
	defer enrichStatus.Lock.Unlock()
	if errCode != http.StatusOK {
		enrichStatus.HasFailure = true
	}
}

func fillTimeZoneAndDateProperty(documents []model.SalesforceDocument, TimeZoneStr U.TimeZoneString, dateProperties *map[string]bool) error {
	if dateProperties == nil || TimeZoneStr == "" {
		return errors.New("missing date properties")
	}

	for i := range documents {
		documents[i].SetDocumentTimeZone(TimeZoneStr)
		documents[i].SetDateProperties(dateProperties)
	}
	return nil
}

// Enrich sync salesforce documents to events
func Enrich(projectID int64, workerPerProject int, dataPropertiesByType map[int]*map[string]bool) ([]Status, bool) {

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

	otpRules, errCode := store.GetStore().GetALLOTPRuleWithProjectId(projectID)
	if errCode != http.StatusFound && errCode != http.StatusNotFound {
		logCtx.WithField("err_code", errCode).Error("Failed to get otp Rules for Project")
		statusByProjectAndType = append(statusByProjectAndType, Status{ProjectID: projectID,
			Status: "Failed to get OTP rules"})
		return statusByProjectAndType, true
	}

	uniqueOTPEventKeys, errCode := store.GetStore().GetUniqueKeyPropertyForOTPEventForLast3Months(projectID)
	if errCode != http.StatusFound && errCode != http.StatusNotFound {
		logCtx.WithField("err_code", errCode).Error("Failed to get OTP Unique Keys for Project")
		statusByProjectAndType = append(statusByProjectAndType, Status{ProjectID: projectID,
			Status: "Failed to get OTP Unique Keys"})
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

	pendingOpportunityGroupAssociations := make(map[string]map[string]string)
	enrichOrderByType := salesforceEnrichOrderByType[:]
	if C.IsAllowedSalesforceActivityTasksByProjectID(projectID) {
		enrichOrderByType = append(enrichOrderByType, model.SalesforceDocumentTypeTask)
	}
	if C.IsAllowedSalesforceActivityEventsByProjectID(projectID) {
		enrichOrderByType = append(enrichOrderByType, model.SalesforceDocumentTypeEvent)
	}

	if C.IsAllowedSalesforceGroupsByProjectID(projectID) {
		var syncStatus map[string]bool
		syncStatus, pendingOpportunityGroupAssociations, status = enrichGroup(projectID, workerPerProject, salesforceSmartEventNames)
		if status != http.StatusOK {
			overAllSyncStatus["groups"] = true
		}

		for docType := range syncStatus {
			overAllSyncStatus[fmt.Sprintf("groups_%s", docType)] = syncStatus[docType]
		}

		enrichOrderByType = append(enrichOrderByType, model.SalesforceDocumentTypeOpportunityContactRole)
	}

	timeZoneStr, status := store.GetStore().GetTimezoneForProject(projectID)
	if status != http.StatusFound {
		logCtx.Error("Failed to get timezone for project.")
		return statusByProjectAndType, true
	}

	for _, timeRange := range orderedTimeSeries {

		for _, docType := range enrichOrderByType {

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
				if errCode != http.StatusNotFound {
					logCtx.Error("Failed to get salesforce document by type for sync.")
				}
				continue
			}

			fillTimeZoneAndDateProperty(documents, timeZoneStr, dataPropertiesByType[docType])

			batches := GetBatchedOrderedDocumentsByID(documents, workerPerProject)

			workerIndex := 0
			var enrichStatus enrichWorkerStatus
			for i := range batches {
				batch := batches[i]
				var wg sync.WaitGroup
				for docID := range batch {
					logCtx.WithFields(log.Fields{"worker": workerIndex, "doc_id": docID, "type": docTypeAlias}).Info("Processing Batch by doc_id")
					wg.Add(1)
					go enrichAllWorker(project, &wg, &enrichStatus, &otpRules, &uniqueOTPEventKeys, batch[docID], (*salesforceSmartEventNames)[docTypeAlias], pendingOpportunityGroupAssociations, timeRange[1])
					workerIndex++
				}
				wg.Wait()
			}

			if _, exist := overAllSyncStatus[docTypeAlias]; !exist {
				overAllSyncStatus[docTypeAlias] = false
			}

			if enrichStatus.HasFailure {
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
