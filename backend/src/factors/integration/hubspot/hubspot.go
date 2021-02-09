package hubspot

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	C "factors/config"
	"factors/model/model"
	"factors/model/store"
	SDK "factors/sdk"
	U "factors/util"
)

// Version definiton
type Version struct {
	Name      string `json:"version"`
	Value     string `json:"value"`
	Timestamp int64  `json:"timestamp"`
}

// Property definiton
type Property struct {
	Value     string    `json:"value"`
	Versions  []Version `json:"versions"`
	Timestamp int64     `json:"timestamp"`
}

// Associations struct for deal associations
type Associations struct {
	AssociatedContactIds []int64 `json:"associatedVids"`
	AssociatedCompanyIds []int64 `json:"associatedCompanyIds"`
	AssociatedDealIds    []int64 `json:"associatedDealIds"`
}

// ContactIdentity struct for contact profile
type ContactIdentity struct {
	Type      string      `json:"type"`
	Value     interface{} `json:"value"`
	IsPrimary bool        `json:"is-primary"`
}

// ContactIdentityProfile for contact
type ContactIdentityProfile struct {
	Identities []ContactIdentity `json:"identities"`
}

// Contact definition
type Contact struct {
	Vid              int64                    `json:"vid"`
	Properties       map[string]Property      `json:"properties"`
	IdentityProfiles []ContactIdentityProfile `json:"identity-profiles"`
}

// Deal definition
type Deal struct {
	DealId       int64               `json:"dealId"`
	Properties   map[string]Property `json:"properties"`
	Associations Associations        `json:"associations"`
}

// Company definition
type Company struct {
	CompanyId int64 `json:"companyId"`
	// not part of hubspot response. added to company on download.
	ContactIds []int64             `json:"contactIds"`
	Properties map[string]Property `json:"properties"`
}

var syncOrderByType = [...]int{
	model.HubspotDocumentTypeContact,
	model.HubspotDocumentTypeCompany,
	model.HubspotDocumentTypeDeal,
}

func getContactProperties(document *model.HubspotDocument) (*map[string]interface{}, *map[string]interface{}, error) {
	if document.Type != model.HubspotDocumentTypeContact {
		return nil, nil, errors.New("invalid type")
	}

	var contact Contact
	err := json.Unmarshal((document.Value).RawMessage, &contact)
	if err != nil {
		return nil, nil, err
	}

	enrichedProperties := make(map[string]interface{}, 0)
	properties := make(map[string]interface{}, 0)

	for ipi := range contact.IdentityProfiles {
		for idi := range contact.IdentityProfiles[ipi].Identities {
			key := contact.IdentityProfiles[ipi].Identities[idi].Type
			enkey := model.GetCRMEnrichPropertyKeyByType(
				model.SmartCRMEventSourceHubspot,
				model.HubspotDocumentTypeNameContact,
				key,
			)
			if _, exists := enrichedProperties[enkey]; !exists {
				enrichedProperties[enkey] = contact.IdentityProfiles[ipi].Identities[idi].Value
			}

			if _, exists := properties[key]; !exists {
				properties[key] = contact.IdentityProfiles[ipi].Identities[idi].Value
			}
		}
	}

	for pkey, pvalue := range contact.Properties {
		key := model.GetCRMEnrichPropertyKeyByType(model.SmartCRMEventSourceHubspot,
			model.HubspotDocumentTypeNameContact, pkey)

		// give precedence to identity profiles, do not
		// overwrite same key from forstore.GetStore().
		if _, exists := enrichedProperties[key]; !exists {
			enrichedProperties[key] = pvalue.Value
		}

		if _, exists := properties[pkey]; !exists {
			properties[pkey] = pvalue.Value
		}
	}

	return &enrichedProperties, &properties, nil
}

func getCustomerUserIDFromProperties(projectID uint64, properties map[string]interface{}) string {
	// identify using email if exist on properties.
	emailInt, emailExists := properties[model.GetCRMEnrichPropertyKeyByType(model.SmartCRMEventSourceHubspot,
		model.HubspotDocumentTypeNameContact, "email")]
	if emailExists || emailInt != nil {
		email, ok := emailInt.(string)
		if ok && email != "" {
			return U.GetEmailLowerCase(email)
		}
	}

	// identify using phone if exist on properties.
	phoneInt, phoneExists := properties[model.GetCRMEnrichPropertyKeyByType(model.SmartCRMEventSourceHubspot,
		model.HubspotDocumentTypeNameContact, "phone")]
	if phoneExists || phoneInt != nil {
		phone := U.GetPropertyValueAsString(phoneInt)
		if phone != "" {
			identifiedPhone, _ := store.GetStore().GetUserIdentificationPhoneNumber(projectID, phone)
			if identifiedPhone != "" {
				return identifiedPhone
			}
		}
	}

	// other possible phone keys.
	var phoneKey string
	for key := range properties {
		hasPhone := strings.Index(key, "phone")
		if hasPhone > -1 && phoneKey == "" {
			phoneKey = key
		}
	}

	if phoneKey != "" {
		phoneInt = properties[phoneKey]
		if phoneInt != nil {
			phone := U.GetPropertyValueAsString(phoneInt)
			if phone != "" {
				identifiedPhone, _ := store.GetStore().GetUserIdentificationPhoneNumber(projectID, phone)
				if identifiedPhone != "" {
					return identifiedPhone
				}
			}
		}
	}

	return ""
}

func getEventTimestamp(timestamp int64) int64 {
	if timestamp == 0 {
		return 0
	}

	return timestamp / 1000
}

/*
GetHubspotSmartEventPayload return smart event payload if the rule successfully gets passed.
WITHOUT PREVIOUS PROPERTY :- A query will be made for previous synced record which
will require userID or customerUserID and doctType
WITH PREVIOUS PROPERTY := userID, customerUserID and doctType won't be used
*/
func GetHubspotSmartEventPayload(projectID uint64, eventName, customerUserID, userID string,
	docType int, currentProperties, prevProperties *map[string]interface{},
	filter *model.SmartCRMEventFilter) (*model.CRMSmartEvent, *map[string]interface{}, bool) {

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
		prevDoc, status := store.GetStore().GetLastSyncedHubspotDocumentByCustomerUserIDORUserID(projectID, customerUserID, userID, docType)
		if status != http.StatusFound && status != http.StatusNotFound {
			return nil, prevProperties, false
		}

		var err error
		if status == http.StatusNotFound { // use empty properties if no previous record exist
			prevProperties = &map[string]interface{}{}
		} else {

			if docType == model.HubspotDocumentTypeContact {
				_, prevProperties, err = getContactProperties(prevDoc)
			}
			if docType == model.HubspotDocumentTypeDeal {
				_, prevProperties, err = getDealProperties(prevDoc)
			}

			if err != nil {
				logCtx.WithError(err).Error("Failed to GetHubspotDocumentProperties")
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

func getTimestampFromField(propertyName string, properties *map[string]interface{}) (int64, error) {
	if timestamp, exists := (*properties)[propertyName]; exists {
		return model.ReadHubspotTimestamp(timestamp)
	}

	return 0, errors.New("field doest not exist")
}

// TrackHubspotSmartEvent valids hubspot current properties with CRM smart filter and creates a event
func TrackHubspotSmartEvent(projectID uint64, hubspotSmartEventName *HubspotSmartEventName, eventID, customerUserID, userID string, docType int, currentProperties, prevProperties *map[string]interface{}, recordTimestamp int64) *map[string]interface{} {
	var valid bool
	var smartEventPayload *model.CRMSmartEvent
	if hubspotSmartEventName.EventName == "" || projectID == 0 || hubspotSmartEventName.Type == "" {
		return prevProperties
	}

	if userID == "" || docType == 0 || currentProperties == nil || hubspotSmartEventName.Filter == nil {
		return prevProperties
	}

	logCtx := log.WithFields(log.Fields{"project_id": projectID, "doc_type": docType})
	smartEventPayload, prevProperties, valid = GetHubspotSmartEventPayload(projectID, hubspotSmartEventName.EventName, customerUserID,
		userID, docType, currentProperties, prevProperties, hubspotSmartEventName.Filter)
	if !valid {
		return prevProperties
	}

	model.AddSmartEventReferenceMeta(&smartEventPayload.Properties, eventID)

	smartEventTrackPayload := &SDK.TrackPayload{
		ProjectId:       projectID,
		EventProperties: smartEventPayload.Properties,
		Name:            smartEventPayload.Name,
		SmartEventType:  hubspotSmartEventName.Type,
		UserId:          userID,
	}

	timestampReferenceField := hubspotSmartEventName.Filter.TimestampReferenceField
	if timestampReferenceField == model.TimestampReferenceTypeTrack {
		smartEventTrackPayload.Timestamp = getEventTimestamp(recordTimestamp)
	} else {
		fieldTimestamp, err := getTimestampFromField(timestampReferenceField, currentProperties)
		if err != nil {
			logCtx.WithField("timestamp_refrence_field", timestampReferenceField).
				WithError(err).Errorf("Failed to get timestamp from reference field")
			smartEventTrackPayload.Timestamp = getEventTimestamp(recordTimestamp) // use record timestamp if custom timestamp not available
		} else {
			smartEventTrackPayload.Timestamp = getEventTimestamp(fieldTimestamp)
		}
	}

	if !C.IsDryRunCRMSmartEvent() {
		status, _ := SDK.Track(projectID, smartEventTrackPayload, true, SDK.SourceHubspot)
		if status != http.StatusOK && status != http.StatusFound && status != http.StatusNotModified {
			logCtx.Error("Failed to create hubspot smart event")
		}
	} else {
		logCtx.WithFields(log.Fields{"properties": smartEventPayload.Properties, "event_name": smartEventPayload.Name,
			"filter_exp":            *hubspotSmartEventName.Filter,
			"smart_event_timestamp": smartEventTrackPayload.Timestamp}).Info("Dry run smart event creation.")
	}

	return prevProperties
}

func syncContact(projectID uint64, document *model.HubspotDocument, hubspotSmartEventNames []HubspotSmartEventName) int {
	logCtx := log.WithField("project_id",
		projectID).WithField("document_id", document.ID)

	enProperties, properties, err := getContactProperties(document)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get properites from hubspot contact.")
		return http.StatusInternalServerError
	}

	leadGUID, exists := (*enProperties)[model.UserPropertyHubspotContactLeadGUID]
	if !exists {
		logCtx.Error("Missing lead_guid on hubspot contact properties. Sync failed.")
		return http.StatusInternalServerError
	}

	trackPayload := &SDK.TrackPayload{
		ProjectId:       projectID,
		EventProperties: *enProperties,
		UserProperties:  *enProperties,
		Timestamp:       getEventTimestamp(document.Timestamp),
	}

	logCtx = logCtx.WithField("action", document.Action).WithField(
		model.UserPropertyHubspotContactLeadGUID, leadGUID)

	customerUserID := getCustomerUserIDFromProperties(projectID, *enProperties)
	var eventID, userID string
	if document.Action == model.HubspotDocumentActionCreated {

		user, status := store.GetStore().CreateUser(&model.User{
			ProjectId:      projectID,
			JoinTimestamp:  getEventTimestamp(document.Timestamp),
			CustomerUserId: customerUserID})
		if status != http.StatusCreated {
			logCtx.WithField("status", status).Error("Failed to create user for hubspot contact created event.")
			return http.StatusInternalServerError
		}

		trackPayload.Name = U.EVENT_NAME_HUBSPOT_CONTACT_CREATED
		trackPayload.UserId = user.ID

		status, response := SDK.Track(projectID, trackPayload, true, SDK.SourceHubspot)
		if status != http.StatusOK && status != http.StatusFound && status != http.StatusNotModified {
			logCtx.WithField("status", status).Error("Failed to track hubspot contact created event.")
			return http.StatusInternalServerError
		}

		userID = user.ID
		eventID = response.EventId
	} else if document.Action == model.HubspotDocumentActionUpdated {
		trackPayload.Name = U.EVENT_NAME_HUBSPOT_CONTACT_UPDATED

		userPropertiesRecords, errCode := store.GetStore().GetUserPropertiesRecordsByProperty(
			projectID, model.UserPropertyHubspotContactLeadGUID, leadGUID)
		if errCode != http.StatusFound {
			logCtx.WithField("err_code", errCode).Error(
				"Failed to get user with given lead_guid. Failed to track hubspot contact updated event.")
			return http.StatusInternalServerError
		}

		// use the user_id of same lead_guid done
		// contact created event.
		userID = userPropertiesRecords[0].UserId
		trackPayload.UserId = userID

		if customerUserID != "" {
			status, _ := SDK.Identify(projectID, &SDK.IdentifyPayload{
				UserId: userID, CustomerUserId: customerUserID}, false)
			if status != http.StatusOK {
				logCtx.WithField("customer_user_id", customerUserID).Error(
					"Failed to identify user on hubspot contact sync.")
				return http.StatusInternalServerError
			}
		} else {
			logCtx.Error("Skipped user identification on hubspot contact sync. No customer_user_id on properties.")
		}

		status, response := SDK.Track(projectID, trackPayload, true, SDK.SourceHubspot)
		if status != http.StatusOK && status != http.StatusFound && status != http.StatusNotModified {
			logCtx.WithField("status", status).Error("Failed to track hubspot contact updated event.")
			return http.StatusInternalServerError
		}
		eventID = response.EventId

	} else {
		logCtx.Error("Invalid action on hubspot contact sync.")
		return http.StatusInternalServerError
	}

	var prevProperties *map[string]interface{}
	for i := range hubspotSmartEventNames {
		prevProperties = TrackHubspotSmartEvent(projectID, &hubspotSmartEventNames[i], eventID, customerUserID, userID, document.Type, properties, prevProperties, document.Timestamp)
	}

	// Mark as synced, if customer_user_id not present or present and identified.
	errCode := store.GetStore().UpdateHubspotDocumentAsSynced(
		projectID, document.ID, eventID, document.Timestamp, document.Action, userID)
	if errCode != http.StatusAccepted {
		logCtx.Error("Failed to update hubspot contact document as synced.")
		return http.StatusInternalServerError
	}

	return http.StatusOK
}

func getDealUserID(projectID uint64, deal *Deal) string {
	logCtx := log.WithField("project_id", projectID)

	contactIds := make([]string, 0, 0)
	// Get directly associated contacts.
	if len(deal.Associations.AssociatedContactIds) > 0 {
		// Considering first contact as primary contact.
		for i := range deal.Associations.AssociatedContactIds {
			contactIds = append(contactIds,
				strconv.FormatInt(deal.Associations.AssociatedContactIds[i], 10))
		}
	}

	// If no directly associated contacts available, get
	// contacts of companies directly associated.
	if len(contactIds) == 0 && len(deal.Associations.AssociatedCompanyIds) > 0 {
		// Considering first company as primary company.
		companyID := strconv.FormatInt(deal.Associations.AssociatedCompanyIds[0], 10)
		companyDocuments, errCode := store.GetStore().GetHubspotDocumentByTypeAndActions(projectID,
			[]string{companyID}, model.HubspotDocumentTypeCompany,
			[]int{model.HubspotDocumentActionCreated, model.HubspotDocumentActionUpdated})

		if errCode == http.StatusInternalServerError {
			logCtx.Error(
				"Failed to get deal user. Failed to get synced hubspot company documents.")
			return ""
		}

		companyContactIds := make(map[int64]bool, 0)
		for _, companyDocument := range companyDocuments {
			var company Company
			err := json.Unmarshal((companyDocument.Value).RawMessage, &company)
			if err != nil {
				logCtx.WithError(err).Error("Failed to unmarshal company document on get deal user")
			}

			for i := range company.ContactIds {
				companyContactIds[company.ContactIds[i]] = true
			}
		}

		for id := range companyContactIds {
			if id > 0 {
				contactIds = append(contactIds, strconv.FormatInt(id, 10))
			}
		}
	}

	if len(contactIds) == 0 {
		return ""
	}

	contactDocuments, errCode := store.GetStore().GetHubspotDocumentByTypeAndActions(projectID,
		contactIds, model.HubspotDocumentTypeContact, []int{model.HubspotDocumentActionCreated})
	if errCode == http.StatusInternalServerError {
		logCtx.Error(
			"Failed to get deal user. Failed to get synced hubspot contact documents.")
		return ""
	}

	// No synced contact document.
	if errCode == http.StatusNotFound || len(contactDocuments) == 0 {
		return ""
	}

	// Use first contact as primary contact.
	contactDocument := contactDocuments[0]
	if contactDocument.SyncId == "" {
		logCtx.Error("No sync_id on synced hubspot contact document.")
		return ""
	}

	event, errCode := store.GetStore().GetEventById(projectID, contactDocument.SyncId)
	if errCode != http.StatusFound {
		logCtx.WithField("event_id", contactDocument.SyncId).Error(
			"Failed to get deal user. Failed to get hubspot contact created event using sync_id.")
		return ""
	}

	return event.UserId
}

// HubspotSmartEventName holds event_name and filter expression
type HubspotSmartEventName struct {
	EventName string
	Filter    *model.SmartCRMEventFilter
	Type      string
}

// GetHubspotSmartEventNames returns all the smart_event for hubspot by object_type
func GetHubspotSmartEventNames(projectID uint64) *map[string][]HubspotSmartEventName {

	logCtx := log.WithFields(log.Fields{"project_id": projectID})

	eventNames, errCode := store.GetStore().GetSmartEventFilterEventNames(projectID)
	if errCode == http.StatusInternalServerError {
		logCtx.Error("Error while GetSmartEventFilterEventNames")
	}

	hubspotSmartEventNames := make(map[string][]HubspotSmartEventName)

	if len(eventNames) == 0 {
		return &hubspotSmartEventNames
	}

	for i := range eventNames {
		if eventNames[i].Type != model.TYPE_CRM_HUBSPOT {
			continue
		}

		var hubspotSmartEventName HubspotSmartEventName
		decFilterExp, err := model.GetDecodedSmartEventFilterExp(eventNames[i].FilterExpr)
		if err != nil {
			logCtx.WithError(err).Error("Failed to decode smart event filter expression")
			continue
		}

		hubspotSmartEventName.EventName = eventNames[i].Name
		hubspotSmartEventName.Filter = decFilterExp
		hubspotSmartEventName.Type = model.TYPE_CRM_HUBSPOT

		if _, exists := hubspotSmartEventNames[decFilterExp.ObjectType]; !exists {
			hubspotSmartEventNames[decFilterExp.ObjectType] = []HubspotSmartEventName{}
		}

		hubspotSmartEventNames[decFilterExp.ObjectType] = append(hubspotSmartEventNames[decFilterExp.ObjectType], hubspotSmartEventName)
	}

	return &hubspotSmartEventNames
}

func syncCompany(projectID uint64, document *model.HubspotDocument) int {
	if document.Type != model.HubspotDocumentTypeCompany {
		return http.StatusInternalServerError
	}

	logCtx := log.WithField("project_id",
		projectID).WithField("document_id", document.ID)

	var company Company
	err := json.Unmarshal((document.Value).RawMessage, &company)
	if err != nil {
		logCtx.WithError(err).Error("Falied to unmarshal hubspot company document.")
		return http.StatusInternalServerError
	}

	if len(company.ContactIds) == 0 {
		logCtx.Error("Skipped company sync. No contacts associated to company.")
	} else {
		contactIds := make([]string, 0, 0)
		for i := range company.ContactIds {
			contactIds = append(contactIds,
				strconv.FormatInt(company.ContactIds[i], 10))
		}

		contactDocuments, errCode := store.GetStore().GetHubspotDocumentByTypeAndActions(projectID,
			contactIds, model.HubspotDocumentTypeContact, []int{model.HubspotDocumentActionCreated})
		if errCode == http.StatusInternalServerError {
			logCtx.Error("Failed to get hubspot documents by type and action on sync company.")
			return errCode
		}

		// build user properties from properties.
		// make sure company name exist.
		userProperties := make(map[string]interface{}, 0)
		for key, value := range company.Properties {
			// add company name to user default property.
			if key == "name" {
				userProperties[U.UP_COMPANY] = value.Value
			}

			propertyKey := model.GetCRMEnrichPropertyKeyByType(model.SmartCRMEventSourceHubspot, model.HubspotDocumentTypeNameCompany, key)
			userProperties[propertyKey] = value.Value
		}

		userPropertiesJsonb, err := U.EncodeToPostgresJsonb(&userProperties)
		if err != nil {
			logCtx.WithError(err).Error("Failed to marshal company properties to Jsonb.")
			return http.StatusInternalServerError
		}

		// update $hubspot_company_name and other company
		// properties on each associated contact user.
		isContactsUpdateFailed := false
		for _, contactDocument := range contactDocuments {
			if contactDocument.SyncId != "" {
				contactSyncEvent, errCode := store.GetStore().GetEventById(
					projectID, contactDocument.SyncId)
				if errCode == http.StatusFound {
					_, errCode := store.GetStore().UpdateUserProperties(projectID,
						contactSyncEvent.UserId, userPropertiesJsonb, time.Now().Unix())
					if errCode != http.StatusAccepted && errCode != http.StatusNotModified {
						logCtx.WithField("user_id", contactSyncEvent.UserId).Error(
							"Failed to update user properites with company properties.")
						isContactsUpdateFailed = true
					}
				}
			}
		}

		if isContactsUpdateFailed {
			logCtx.Error("Failed to update some hubspot company properties on user properties.")
			return http.StatusInternalServerError
		}
	}

	// No sync_id as no event or user or one user property created.
	errCode := store.GetStore().UpdateHubspotDocumentAsSynced(projectID, document.ID, "", document.Timestamp, document.Action, "")
	if errCode != http.StatusAccepted {
		logCtx.Error("Failed to update hubspot deal document as synced.")
		return http.StatusInternalServerError
	}

	return http.StatusOK
}

func getDealProperties(document *model.HubspotDocument) (*map[string]interface{}, *map[string]interface{}, error) {

	if document.Type != model.HubspotDocumentTypeDeal {
		return nil, nil, errors.New("invalid type")
	}

	var deal Deal
	err := json.Unmarshal((document.Value).RawMessage, &deal)
	if err != nil {
		return nil, nil, err
	}

	enProperties := make(map[string]interface{}, 0)
	properties := make(map[string]interface{})
	for k, v := range deal.Properties {
		key := model.GetCRMEnrichPropertyKeyByType(model.SmartCRMEventSourceHubspot,
			model.HubspotDocumentTypeNameDeal, k)
		enProperties[key] = v.Value
		properties[k] = v.Value
	}

	return &enProperties, &properties, nil
}

func syncDeal(projectID uint64, document *model.HubspotDocument, hubspotSmartEventNames []HubspotSmartEventName) int {
	if document.Type != model.HubspotDocumentTypeDeal {
		return http.StatusInternalServerError
	}

	logCtx := log.WithField("project_id",
		projectID).WithField("document_id", document.ID)

	var deal Deal
	err := json.Unmarshal((document.Value).RawMessage, &deal)
	if err != nil {
		logCtx.Error("Failed to unmarshal hubspot document deal.")
		return http.StatusInternalServerError
	}

	enProperties, properties, err := getDealProperties(document)
	if err != nil {
		logCtx.Error("Failed to get hubspot deal document properties")
		return http.StatusInternalServerError
	}

	dealStage, exists := (*enProperties)[U.CRM_HUBSPOT_DEALSTAGE]
	if !exists || dealStage == nil {
		logCtx.Error("No deal stage property found on hubspot deal.")
		return http.StatusInternalServerError
	}

	userID := getDealUserID(projectID, &deal)

	eventID := ""
	if userID == "" {
		logCtx.Error("Skipped deal sync. No user associated to hubspot deal.")
	} else {
		trackPayload := &SDK.TrackPayload{
			Name:            U.EVENT_NAME_HUBSPOT_DEAL_STATE_CHANGED,
			ProjectId:       projectID,
			UserId:          userID,
			EventProperties: *enProperties,
			UserProperties:  *enProperties,
			Timestamp:       getEventTimestamp(document.Timestamp),
		}

		// Track deal stage change only if, deal with same id and
		// same stage, not synced before.
		dealID := strconv.FormatInt(deal.DealId, 10)
		if dealID == "" {
			logCtx.Error("Invalid deal_id on conversion. Failed to sync deal.")
			return http.StatusInternalServerError
		}

		_, errCode := store.GetStore().GetSyncedHubspotDealDocumentByIdAndStage(projectID,
			dealID, dealStage.(string))
		if errCode == http.StatusInternalServerError || errCode == http.StatusBadRequest {
			logCtx.Error("Failed to get synced deal document by stage on sync_deal")
			return http.StatusInternalServerError
		}

		// skip sync as deal stage is synced already.
		if errCode == http.StatusFound {
			return http.StatusOK
		}

		status, response := SDK.Track(projectID, trackPayload, true, SDK.SourceHubspot)
		if status != http.StatusOK && status != http.StatusFound &&
			status != http.StatusNotModified {

			logCtx.WithField("status", status).Error(
				"Failed to track hubspot contact deal stage change event.")
			return http.StatusInternalServerError
		}

		eventID = response.EventId
		var prevProperties *map[string]interface{}
		for i := range hubspotSmartEventNames {
			prevProperties = TrackHubspotSmartEvent(projectID, &hubspotSmartEventNames[i], response.EventId, "", userID, document.Type, properties, prevProperties, document.Timestamp)
		}
	}

	errCode := store.GetStore().UpdateHubspotDocumentAsSynced(projectID,
		document.ID, eventID, document.Timestamp, document.Action, userID)
	if errCode != http.StatusAccepted {
		logCtx.Error("Failed to update hubspot deal document as synced.")
		return http.StatusInternalServerError
	}

	return http.StatusOK
}

func syncAll(projectID uint64, documents []model.HubspotDocument, hubspotSmartEventNames []HubspotSmartEventName) int {
	logCtx := log.WithField("project_id", projectID)

	var seenFailures bool
	for i := range documents {
		logCtx = logCtx.WithField("document", documents[i])
		startTime := time.Now().Unix()

		switch documents[i].Type {
		case model.HubspotDocumentTypeContact:
			errCode := syncContact(projectID, &documents[i], hubspotSmartEventNames)
			if errCode != http.StatusOK {
				seenFailures = true
			}
		case model.HubspotDocumentTypeCompany:
			errCode := syncCompany(projectID, &documents[i])
			if errCode != http.StatusOK {
				seenFailures = true
			}
		case model.HubspotDocumentTypeDeal:
			errCode := syncDeal(projectID, &documents[i], hubspotSmartEventNames)
			if errCode != http.StatusOK {
				seenFailures = true
			}
		}

		logCtx.WithField("time_taken_in_secs", time.Now().Unix()-startTime).Debugf(
			"Sync %s completed.", documents[i].TypeAlias)
	}

	if seenFailures {
		return http.StatusInternalServerError
	}

	return http.StatusOK
}

// Status definition
type Status struct {
	ProjectId uint64 `json:"project_id"`
	Type      string `json:"type"`
	Status    string `json:"status"`
}

// Sync - Syncs hubspot documents in an order of type.
func Sync(projectID uint64) []Status {
	logCtx := log.WithField("project_id", projectID)

	statusByProjectAndType := make([]Status, 0, 0)
	hubspotSmartEventNames := GetHubspotSmartEventNames(projectID)

	for i := range syncOrderByType {
		logCtx = logCtx.WithField("type", syncOrderByType[i])

		documents, errCode := store.GetStore().
			GetHubspotDocumentsByTypeForSync(projectID, syncOrderByType[i])
		if errCode != http.StatusFound {
			logCtx.Error("Failed to get hubspot document by type for sync.")
			continue
		}

		docTypeAlias := model.GetHubspotTypeAliasByType(syncOrderByType[i])
		status := Status{ProjectId: projectID,
			Type: docTypeAlias}

		errCode = syncAll(projectID, documents, (*hubspotSmartEventNames)[docTypeAlias])
		if errCode == http.StatusOK {
			status.Status = U.CRM_SYNC_STATUS_SUCCESS
		} else {
			status.Status = U.CRM_SYNC_STATUS_FAILURES
		}
		statusByProjectAndType = append(statusByProjectAndType, status)
	}

	return statusByProjectAndType
}
