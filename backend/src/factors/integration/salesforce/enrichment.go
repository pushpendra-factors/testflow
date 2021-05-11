package salesforce

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
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

var salesforceSyncOrderByType = [...]int{
	model.SalesforceDocumentTypeContact,
	model.SalesforceDocumentTypeLead,
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
	model.SalesforceChildRelationshipNameOpportunityContactRoles,
	model.SalesforceDocumentTypeNameLead,
}

var allowedCampaignFields = map[string]bool{}

func getUserIDFromLastestProperties(properties []model.UserProperties) string {
	latestIndex := len(properties) - 1
	return properties[latestIndex].UserId
}

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
	for action updated -> create on updated event with lastmodified timestamp
*/
func TrackSalesforceEventByDocumentType(projectID uint64, trackPayload *SDK.TrackPayload, document *model.SalesforceDocument, customerUserID string) (string, string, error) {
	if projectID == 0 {
		return "", "", errors.New("invalid project id")
	}

	if trackPayload == nil || document == nil {
		return "", "", errors.New("invalid operation")
	}

	createdTimestamp, err := model.GetSalesforceDocumentTimestampByAction(document, model.SalesforceDocumentCreated)
	if err != nil {
		return "", "", err
	}

	lastModifiedTimestamp, err := model.GetSalesforceDocumentTimestampByAction(document, model.SalesforceDocumentUpdated)
	if err != nil {
		return "", "", err
	}

	var eventID, userID string
	if document.Action == model.SalesforceDocumentCreated {
		payload := *trackPayload
		if customerUserID != "" {
			user, status := store.GetStore().CreateUser(&model.User{
				ProjectId:      projectID,
				CustomerUserId: customerUserID,
				JoinTimestamp:  createdTimestamp,
			})

			if status != http.StatusCreated {
				return "", "", fmt.Errorf("create user failed for doc type %d, status code %d", document.Type, status)
			}
			payload.UserId = user.ID
		}

		payload.Name = model.GetSalesforceEventNameByDocumentAndAction(document, model.SalesforceDocumentCreated)
		payload.Timestamp = createdTimestamp

		status, response := SDK.Track(projectID, &payload, true, SDK.SourceSalesforce)
		if status != http.StatusOK && status != http.StatusFound && status != http.StatusNotModified {
			return "", "", fmt.Errorf("created event track failed for doc type %d, message %s", document.Type, response.Error)
		}

		if payload.UserId != "" {
			userID = payload.UserId
		} else {
			userID = response.UserId
		}

		eventID = response.EventId
	}

	if document.Action == model.SalesforceDocumentCreated || document.Action == model.SalesforceDocumentUpdated {
		payload := *trackPayload
		payload.Name = model.GetSalesforceEventNameByDocumentAndAction(document, model.SalesforceDocumentUpdated)

		if document.Action == model.SalesforceDocumentUpdated {

			payload.Timestamp = lastModifiedTimestamp
			// TODO(maisa): Use GetSyncedSalesforceDocumentByType while updating multiple contacts in an account object
			documents, status := store.GetStore().GetSyncedSalesforceDocumentByType(projectID, []string{document.ID}, document.Type)
			if status != http.StatusFound {
				return "", "", errors.New("failed to get synced document")
			}

			event, status := store.GetStore().GetEventById(projectID, documents[0].SyncID)
			if status != http.StatusFound {
				return "", "", errors.New("failed to get event from sync id ")
			}

			if customerUserID != "" {
				status, _ = SDK.Identify(projectID, &SDK.IdentifyPayload{
					UserId:         event.UserId,
					CustomerUserId: customerUserID,
					Timestamp:      lastModifiedTimestamp,
				}, false)

				if status != http.StatusOK {
					return "", "", fmt.Errorf("failed indentifying user on update event track")
				}
			}

			userID = event.UserId
		} else {
			payload.Timestamp = createdTimestamp
		}

		payload.UserId = userID

		status, response := SDK.Track(projectID, &payload, true, SDK.SourceSalesforce)
		if status != http.StatusOK && status != http.StatusFound && status != http.StatusNotModified {
			return "", "", fmt.Errorf("updated event track failed for doc type %d", document.Type)
		}

		eventID = response.EventId
	} else {
		return "", "", errors.New("invalid action on salesforce document sync")
	}

	// create additional event for created action if document is not the first version
	if document.Action == model.SalesforceDocumentCreated && createdTimestamp != lastModifiedTimestamp {
		payload := *trackPayload
		payload.Timestamp = lastModifiedTimestamp
		payload.UserId = userID
		payload.Name = model.GetSalesforceEventNameByDocumentAndAction(document, model.SalesforceDocumentUpdated)
		status, _ := SDK.Track(projectID, &payload, true, SDK.SourceSalesforce)
		if status != http.StatusOK && status != http.StatusFound && status != http.StatusNotModified {
			return "", "", fmt.Errorf("updated event for different timestamp track failed for doc type %d", document.Type)
		}
	}

	return eventID, userID, nil
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
		logCtx.Error("Skipping user identification on salesforce account sync. No customer_user_id on properties.")
	}

	eventID, userID, err := TrackSalesforceEventByDocumentType(projectID, trackPayload, document, customerUserID)
	if err != nil {
		logCtx.WithError(err).Error(
			"Failed to track salesforce account event.")
		return http.StatusInternalServerError
	}

	// ALways us lastmodified timestamp for updated properties. Error handling already done during event creation
	lastModifiedTimestamp, _ := model.GetSalesforceDocumentTimestampByAction(document, model.SalesforceDocumentUpdated)

	var prevProperties *map[string]interface{}
	for _, smartEventName := range salesforceSmartEventNames {
		prevProperties = TrackSalesforceSmartEvent(projectID, &smartEventName, eventID, customerUserID, userID, document.Type, properties, prevProperties, lastModifiedTimestamp)
	}

	errCode := store.GetStore().UpdateSalesforceDocumentAsSynced(projectID, document, eventID, userID)
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
		logCtx.Error("Skipping user identification on salesforce contact sync. No customer_user_id on properties.")
	}

	eventID, userID, err := TrackSalesforceEventByDocumentType(projectID, trackPayload, document, customerUserID)
	if err != nil {
		logCtx.WithError(err).Error(
			"Failed to track salesforce contact event.")
		return http.StatusInternalServerError
	}

	// ALways us lastmodified timestamp for updated properties. Error handling already done during event creation
	lastModifiedTimestamp, _ := model.GetSalesforceDocumentTimestampByAction(document, model.SalesforceDocumentUpdated)

	var prevProperties *map[string]interface{}
	for _, smartEventName := range salesforceSmartEventNames {
		prevProperties = TrackSalesforceSmartEvent(projectID, &smartEventName, eventID, customerUserID, userID, document.Type, properties, prevProperties, lastModifiedTimestamp)
	}

	errCode := store.GetStore().UpdateSalesforceDocumentAsSynced(projectID, document, eventID, userID)
	if errCode != http.StatusAccepted {
		logCtx.Error("Failed to update salesforce contact document as synced.")
		return http.StatusInternalServerError
	}

	return http.StatusOK
}

/*
GetSalesforceSmartEventPayload return smart event payload if the rule successfully gets passed.
WITHOUT PREVIOUS PROPERTY :- A query will be made for previous synced record which
will require userID or customerUserID and doctType
WITH PREVIOUS PROPERTY := userID, customerUserID and doctType won't be used
*/
func GetSalesforceSmartEventPayload(projectID uint64, eventName, customerUserID, userID string, docType int,
	currentProperties, prevProperties *map[string]interface{}, filter *model.SmartCRMEventFilter) (*model.CRMSmartEvent, *map[string]interface{}, bool) {

	var crmSmartEvent model.CRMSmartEvent
	var validProperty bool
	var newProperties map[string]interface{}

	if projectID == 0 || eventName == "" || filter == nil || currentProperties == nil {
		return nil, prevProperties, false
	}

	if prevProperties == nil && (docType == 0 || userID == "") {
		return nil, prevProperties, false
	}

	if prevProperties != nil {
		validProperty = model.CRMFilterEvaluator(projectID, currentProperties, prevProperties, filter, model.CompareStateBoth)
	} else {
		validProperty = model.CRMFilterEvaluator(projectID, currentProperties, nil, filter, model.CompareStateCurr)
	}

	logCtx := log.WithFields(log.Fields{"project_id": projectID, "doc_type": docType})

	if !validProperty {
		return nil, prevProperties, false
	}

	if prevProperties == nil {
		prevDoc, status := store.GetStore().GetLastSyncedSalesforceDocumentByCustomerUserIDORUserID(
			projectID, customerUserID, userID, docType)
		if status != http.StatusFound && status != http.StatusNotFound {
			return nil, prevProperties, false
		}

		var err error
		if status == http.StatusNotFound {
			prevProperties = &map[string]interface{}{}
		} else {
			_, prevProperties, err = GetSalesforceDocumentProperties(projectID, prevDoc)
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
func TrackSalesforceSmartEvent(projectID uint64, salesforceSmartEventName *SalesforceSmartEventName, eventID, customerUserID, userID string, docType int, currentProperties, prevProperties *map[string]interface{}, lastModifiedTimestamp int64) *map[string]interface{} {
	var valid bool
	var smartEventPayload *model.CRMSmartEvent

	logCtx := log.WithFields(log.Fields{"project_id": projectID, "doc_type": docType, "user_id": userID, "customer_user_id": customerUserID, "smart_event_rule": salesforceSmartEventName})
	if projectID == 0 || currentProperties == nil || docType == 0 || userID == "" || lastModifiedTimestamp == 0 {
		logCtx.Error("Missing required fields.")
		return prevProperties
	}

	if salesforceSmartEventName.EventName == "" || salesforceSmartEventName.Type == "" || salesforceSmartEventName.Filter == nil {
		logCtx.Error("Missing smart event fileds.")
		return prevProperties
	}

	smartEventPayload, prevProperties, valid = GetSalesforceSmartEventPayload(projectID, salesforceSmartEventName.EventName, customerUserID,
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
		status, _ := SDK.Track(projectID, smartEventTrackPayload, true, SDK.SourceSalesforce)
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

var errMissingOpportunityLeadAndContact = errors.New("missing lead and contact id")

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
			linkedObject, status := store.GetStore().GetSyncedSalesforceDocumentByType(projectID, []string{oppContactID}, model.SalesforceDocumentTypeContact)
			if status == http.StatusFound {
				return &linkedObject[0], nil
			}

		}

		if opportunityMappingOrder[i] == model.SalesforceDocumentTypeNameLead && oppLeadID != "" {
			linkedObject, status := store.GetStore().GetSyncedSalesforceDocumentByType(projectID, []string{oppLeadID}, model.SalesforceDocumentTypeLead)
			if status == http.StatusFound {
				return &linkedObject[0], nil
			}

		}
	}

	return nil, errors.New("missing lead and contact record for opportunity link")
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
	if C.UseOpportunityAssociationByProjectID(projectID) {
		linkedDocument, err := getOpportunityLinkedLeadOrContactDocument(projectID, document)
		if err != nil {
			if err != errMissingOpportunityLeadAndContact {
				logCtx.WithError(err).Error("Failed to get linked document for opportunity.")
			}

		} else {
			linkedDocEnProperties, _, err := GetSalesforceDocumentProperties(projectID, linkedDocument)
			if err == nil {
				customerUserID, userID = getCustomerUserIDFromProperties(projectID, *linkedDocEnProperties, model.GetSalesforceAliasByDocType(linkedDocument.Type), &model.SalesforceProjectIdentificationFieldStore)
				if customerUserID == "" && userID == "" {
					if linkedDocument.UserID != "" {
						userID = linkedDocument.UserID
					}
				}
			} else {
				logCtx.WithError(err).Error("Failed to get properties on opportunities associations")
			}

		}
	}

	if userID == "" {
		customerUserID, userID = getCustomerUserIDFromProperties(projectID, *enProperties, model.GetSalesforceAliasByDocType(document.Type), &model.SalesforceProjectIdentificationFieldStore)
	}

	if customerUserID != "" {
		if userID != "" {
			trackPayload.UserId = userID // will also handle opportunity updated event which is not stiched with other object
			eventID, eventUserID, err = TrackSalesforceEventByDocumentType(projectID, trackPayload, document, "")
		} else {
			eventID, eventUserID, err = TrackSalesforceEventByDocumentType(projectID, trackPayload, document, customerUserID)
		}

		if err != nil {
			logCtx.WithError(err).Error(
				"Failed to track salesforce opportunity event.")
			return http.StatusInternalServerError
		}
	} else {
		trackPayload.UserId = userID
		eventID, eventUserID, err = TrackSalesforceEventByDocumentType(projectID, trackPayload, document, "")
		if err != nil {
			logCtx.WithError(err).Error(
				"Failed to track salesforce opportunity event.")
			return http.StatusInternalServerError
		}

		logCtx.Error("Skipped user identification on salesforce opportunity sync. No customer_user_id on properties.")
	}

	if userID == "" {
		userID = eventUserID
	}

	// ALways us lastmodified timestamp for updated properties. Error handling already done during event creation
	lastModifiedTimestamp, _ := model.GetSalesforceDocumentTimestampByAction(document, model.SalesforceDocumentUpdated)

	var prevProperties *map[string]interface{}
	for _, smartEventName := range salesforceSmartEventNames {
		prevProperties = TrackSalesforceSmartEvent(projectID, &smartEventName, eventID, customerUserID, userID,
			document.Type, properties, prevProperties, lastModifiedTimestamp)
	}

	errCode := store.GetStore().UpdateSalesforceDocumentAsSynced(projectID, document, eventID, userID)
	if errCode != http.StatusAccepted {
		logCtx.Error("Failed to update salesforce opportunity document as synced.")
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
		logCtx.Error("Skipped user identification on salesforce lead sync. No customer_user_id on properties.")
	}

	eventID, userID, err := TrackSalesforceEventByDocumentType(projectID, trackPayload, document, customerUserID)
	if err != nil {
		logCtx.WithError(err).Error(
			"Failed to track salesforce lead event.")
		return http.StatusInternalServerError
	}

	// ALways us lastmodified timestamp for updated properties, error handling already done during event creation
	lastModifiedTimestamp, _ := model.GetSalesforceDocumentTimestampByAction(document, model.SalesforceDocumentUpdated)

	var prevProperties *map[string]interface{}
	for _, smartEventName := range salesforceSmartEventNames {
		prevProperties = TrackSalesforceSmartEvent(projectID, &smartEventName, eventID, customerUserID, userID, document.Type, properties, prevProperties, lastModifiedTimestamp)
	}

	errCode := store.GetStore().UpdateSalesforceDocumentAsSynced(projectID, document, eventID, userID)
	if errCode != http.StatusAccepted {
		logCtx.Error("Failed to update salesforce lead document as synced.")
		return http.StatusInternalServerError
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

func enrichCampaignToAllCampaignMembers(projectID uint64, document *model.SalesforceDocument) int {
	logCtx := log.WithFields(log.Fields{"project_id": projectID, "document_id": document.ID})
	if document.Type != model.SalesforceDocumentTypeCampaign {
		return http.StatusBadRequest
	}

	enProperties, _, err := GetSalesforceDocumentProperties(projectID, document)
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
		status := store.GetStore().UpdateSalesforceDocumentAsSynced(projectID, document, "", "")
		if status != http.StatusAccepted {
			logCtx.Error("Failed to mark campaign as synced.")
			return http.StatusInternalServerError
		}

		return http.StatusOK
	}

	memberDocuments, status := store.GetStore().GetLatestSalesforceDocumentByID(projectID, campaignMemberIDs, model.SalesforceDocumentTypeCampaignMember)
	if status != http.StatusFound {
		logCtx.WithError(err).Error("Failed to get campaign members.")
		return http.StatusInternalServerError
	}

	for i := range memberDocuments {
		enMemberProperties, _, err := GetSalesforceDocumentProperties(projectID, &memberDocuments[i])
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
			existingUserID = getExistingCampaignMemberUserIDFromProperties(projectID, enMemberProperties)
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
			ProjectId:       projectID,
			EventProperties: *enMemberProperties, // no user properties for campaign members
			UserId:          existingUserID,
		}

		eventID, userID, err := TrackSalesforceEventByDocumentType(projectID, trackPayload, &referenceDocument, "")
		if err != nil {
			logCtx.WithField("member_id", referenceDocument.ID).WithError(err).Error(
				"Failed to track salesforce campaign member update on campaign update.")
			return http.StatusInternalServerError
		}

		if memberDocuments[i].Synced == false {
			status = store.GetStore().UpdateSalesforceDocumentAsSynced(projectID, &memberDocuments[i], eventID, userID)
			if status != http.StatusAccepted {
				logCtx.WithField("member_id", referenceDocument.ID).Error("Failed to mark campaign member as synced.")
				return http.StatusInternalServerError
			}
		}
	}

	status = store.GetStore().UpdateSalesforceDocumentAsSynced(projectID, document, "", "")
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
		existingMember, status := store.GetStore().GetSyncedSalesforceDocumentByType(projectID, []string{util.GetPropertyValueAsString(existingContactMemberID)}, model.SalesforceDocumentTypeContact)
		if status == http.StatusFound {
			if existingMember[0].UserID != "" {
				existingUserID = existingMember[0].UserID
			}
		}
	}

	if existingUserID == "" { // use lead Id if available
		existingMember, status := store.GetStore().GetSyncedSalesforceDocumentByType(projectID, []string{util.GetPropertyValueAsString(existingLeadMemberID)}, model.SalesforceDocumentTypeLead)
		if status == http.StatusFound {
			if existingMember[0].UserID != "" {
				existingUserID = existingMember[0].UserID
			}
		}
	}

	return existingUserID
}

func enrichCampaignMember(projectID uint64, document *model.SalesforceDocument) int {
	logCtx := log.WithFields(log.Fields{"project_id": projectID, "document_id": document.ID})
	if document.Type != model.SalesforceDocumentTypeCampaignMember {
		return http.StatusBadRequest
	}

	enCampaingMemberProperties, _, err := GetSalesforceDocumentProperties(projectID, document)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get properties for campaign member.")
		return http.StatusInternalServerError
	}

	campaignID, exist := (*enCampaingMemberProperties)[model.GetCRMEnrichPropertyKeyByType(model.SmartCRMEventSourceSalesforce, model.SalesforceDocumentTypeNameCampaignMember, "CampaignId")]
	if !exist {
		logCtx.Error("Missing campaign_id in campaign member")
		return http.StatusInternalServerError
	}

	campaignDocuments, status := store.GetStore().GetSyncedSalesforceDocumentByType(projectID, []string{util.GetPropertyValueAsString(campaignID)}, model.SalesforceDocumentTypeCampaign)
	if status != http.StatusFound {
		logCtx.Error("Failed to get campaign document for campaign member.")
		return http.StatusInternalServerError
	}

	enCampaignProperties, _, err := GetSalesforceDocumentProperties(projectID, &campaignDocuments[len(campaignDocuments)-1])
	if err != nil {
		logCtx.WithError(err).Error("Failed to get properties for campaign member.")
		return http.StatusInternalServerError
	}

	for pName := range *enCampaignProperties {
		(*enCampaingMemberProperties)[pName] = (*enCampaignProperties)[pName]
	}

	existingUserID := ""
	// use user_id from lead or contact id
	if document.Action == model.SalesforceDocumentCreated {
		existingUserID = getExistingCampaignMemberUserIDFromProperties(projectID, enCampaingMemberProperties)
	}

	trackPayload := &SDK.TrackPayload{
		ProjectId:       projectID,
		EventProperties: *enCampaingMemberProperties,
		UserId:          existingUserID,
	}

	eventID, userID, err := TrackSalesforceEventByDocumentType(projectID, trackPayload, document, "")
	if err != nil {
		logCtx.WithError(err).Error(
			"Failed to track salesforce lead event.")
		return http.StatusInternalServerError
	}

	errCode := store.GetStore().UpdateSalesforceDocumentAsSynced(projectID, document, eventID, userID)
	if errCode != http.StatusAccepted {
		logCtx.Error("Failed to update salesforce lead document as synced.")
		return http.StatusInternalServerError
	}

	return http.StatusOK

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
func enrichCampaign(projectID uint64, document *model.SalesforceDocument) int {
	if projectID == 0 || document == nil {
		return http.StatusBadRequest
	}

	if document.Type == model.SalesforceDocumentTypeCampaign {
		return enrichCampaignToAllCampaignMembers(projectID, document)
	}

	if document.Type == model.SalesforceDocumentTypeCampaignMember {
		return enrichCampaignMember(projectID, document)
	}

	return http.StatusBadRequest
}

func enrichAll(projectID uint64, documents []model.SalesforceDocument, salesforceSmartEventNames []SalesforceSmartEventName) int {
	if projectID == 0 {
		return http.StatusBadRequest
	}
	logCtx := log.WithField("project_id", projectID)

	var seenFailures bool
	var errCode int
	for i := range documents {
		startTime := time.Now().Unix()

		switch documents[i].Type {
		case model.SalesforceDocumentTypeAccount:
			errCode = enrichAccount(projectID, &documents[i], salesforceSmartEventNames)
		case model.SalesforceDocumentTypeContact:
			errCode = enrichContact(projectID, &documents[i], salesforceSmartEventNames)
		case model.SalesforceDocumentTypeLead:
			errCode = enrichLeads(projectID, &documents[i], salesforceSmartEventNames)
		case model.SalesforceDocumentTypeOpportunity:
			errCode = enrichOpportunities(projectID, &documents[i], salesforceSmartEventNames)
		case model.SalesforceDocumentTypeCampaign, model.SalesforceDocumentTypeCampaignMember:
			errCode = enrichCampaign(projectID, &documents[i])
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
			logCtx.WithError(err).Error("Failed to decode smart event filter expression")
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

// Enrich sync salesforce documents to events
func Enrich(projectID uint64) ([]Status, bool) {

	logCtx := log.WithField("project_id", projectID)

	statusByProjectAndType := make([]Status, 0, 0)
	if projectID == 0 {
		return statusByProjectAndType, true
	}

	allowedDocTypes := model.GetSalesforceDocumentTypeAlias(projectID)

	salesforceSmartEventNames := GetSalesforceSmartEventNames(projectID)

	anyFailure := false
	for _, docType := range salesforceSyncOrderByType {
		docTypeAlias := model.GetSalesforceAliasByDocType(docType)
		if _, exist := allowedDocTypes[docTypeAlias]; !exist {
			continue
		}

		logCtx = logCtx.WithFields(log.Fields{
			"doc_type":   docType,
			"project_id": projectID,
		})

		documents, errCode := store.GetStore().GetSalesforceDocumentsByTypeForSync(projectID, docType)
		if errCode != http.StatusFound {
			logCtx.Error("Failed to get salesforce document by type for sync.")
			continue
		}

		status := Status{
			ProjectID: projectID,
			Type:      docTypeAlias,
		}

		errCode = enrichAll(projectID, documents, (*salesforceSmartEventNames)[docTypeAlias])
		if errCode == http.StatusOK {
			status.Status = U.CRM_SYNC_STATUS_SUCCESS
		} else {
			status.Status = U.CRM_SYNC_STATUS_FAILURES
			anyFailure = true
		}

		statusByProjectAndType = append(statusByProjectAndType, status)
	}

	return statusByProjectAndType, anyFailure
}
