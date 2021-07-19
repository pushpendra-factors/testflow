package hubspot

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
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

// PropertyDetail defination for hubspot properties api
type PropertyDetail struct {
	Name      string `json:"name"`
	Label     string `json:"label"`
	Type      string `json:"type"`
	FieldType string `json:"fieldType"`
}

var syncOrderByType = [...]int{
	model.HubspotDocumentTypeContact,
	model.HubspotDocumentTypeCompany,
	model.HubspotDocumentTypeDeal,
}

var allowedEventNames = []string{
	U.EVENT_NAME_HUBSPOT_CONTACT_CREATED,
	U.EVENT_NAME_HUBSPOT_CONTACT_UPDATED,
	U.EVENT_NAME_HUBSPOT_DEAL_STATE_CHANGED,
}

func getContactProperties(projectID uint64, document *model.HubspotDocument) (*map[string]interface{}, *map[string]interface{}, error) {
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
		enKey := model.GetCRMEnrichPropertyKeyByType(model.SmartCRMEventSourceHubspot,
			model.HubspotDocumentTypeNameContact, pkey)
		value, err := getHubspotMappedDataTypeValue(projectID, U.EVENT_NAME_HUBSPOT_CONTACT_UPDATED, enKey, pvalue.Value)
		if err != nil {
			log.WithFields(log.Fields{"project_id": projectID, "property_key": enKey}).WithError(err).Error("Failed to get property value.")
			continue
		}

		// give precedence to identity profiles, do not
		// overwrite same key from forstore.GetStore().
		if _, exists := enrichedProperties[enKey]; !exists {
			enrichedProperties[enKey] = value
		}

		if _, exists := properties[pkey]; !exists {
			properties[pkey] = value
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
	phoneInt := properties[model.GetCRMEnrichPropertyKeyByType(model.SmartCRMEventSourceHubspot,
		model.HubspotDocumentTypeNameContact, "phone")]
	if phoneInt != nil {
		phone := U.GetPropertyValueAsString(phoneInt)
		identifiedPhone, _ := store.GetStore().GetUserIdentificationPhoneNumber(projectID, phone)
		if identifiedPhone != "" {
			return identifiedPhone
		}

	}

	// other possible phone keys.
	for key := range properties {
		hasPhone := strings.Index(key, "phone")
		if hasPhone > -1 {
			phone := U.GetPropertyValueAsString(properties[key])
			identifiedPhone, _ := store.GetStore().GetUserIdentificationPhoneNumber(projectID, phone)
			if identifiedPhone != "" {
				return identifiedPhone
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
will require docID and doctType
WITH PREVIOUS PROPERTY := docID and doctType won't be used
*/
func GetHubspotSmartEventPayload(projectID uint64, eventName, docID string,
	docType int, currentProperties, prevProperties *map[string]interface{},
	filter *model.SmartCRMEventFilter) (*model.CRMSmartEvent, *map[string]interface{}, bool) {

	logCtx := log.WithFields(log.Fields{"project_id": projectID, "doc_id": docID, "doc_type": docType, "filter": filter})
	var crmSmartEvent model.CRMSmartEvent
	var validProperty bool
	var newProperties map[string]interface{}

	if projectID == 0 || eventName == "" || filter == nil || currentProperties == nil ||
		(prevProperties == nil && (docType == 0 || docID == "")) {
		logCtx.Error("Missing required fields")
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
		prevDoc, status := store.GetStore().GetLastSyncedHubspotDocumentByID(projectID, docID, docType)
		if status != http.StatusFound && status != http.StatusNotFound {
			return nil, prevProperties, false
		}

		var err error
		if status == http.StatusNotFound { // use empty properties if no previous record exist
			prevProperties = &map[string]interface{}{}
		} else {

			if docType == model.HubspotDocumentTypeContact {
				_, prevProperties, err = getContactProperties(projectID, prevDoc)
			}
			if docType == model.HubspotDocumentTypeDeal {
				_, prevProperties, err = getDealProperties(projectID, prevDoc)
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

func getTimestampFromField(projectID uint64, propertyName string, properties *map[string]interface{}) (int64, error) {
	if timestampInt, exists := (*properties)[propertyName]; exists {

		if C.IsEnabledPropertyDetailFromDB() && C.IsEnabledPropertyDetailByProjectID(projectID) {
			timestampStr := U.GetPropertyValueAsString(timestampInt)

			if len(timestampStr) == 13 {
				log.WithFields(log.Fields{"property_name": propertyName, "property_value": timestampStr}).Error("Timestamp not in seconds.")
				timestamp, err := model.ReadHubspotTimestamp(timestampInt)
				if timestamp > 0 {
					return timestamp / 1000, err
				}

				return 0, err
			}

			return model.ReadHubspotTimestamp(timestampStr)
		}

		timestamp, err := model.ReadHubspotTimestamp(timestampInt)
		return getEventTimestamp(timestamp), err
	}

	return 0, errors.New("field doest not exist")
}

// TrackHubspotSmartEvent validates hubspot current properties with CRM smart filter and creates a event
func TrackHubspotSmartEvent(projectID uint64, hubspotSmartEventName *HubspotSmartEventName, eventID, docID, userID string, docType int,
	currentProperties, prevProperties *map[string]interface{}, defaultTimestamp int64, usingFallbackUserID bool) *map[string]interface{} {
	var valid bool
	var smartEventPayload *model.CRMSmartEvent

	logCtx := log.WithFields(log.Fields{"project_id": projectID, "doc_type": docType})

	if projectID == 0 || userID == "" || docType == 0 || currentProperties == nil || defaultTimestamp == 0 {
		logCtx.Error("Missing required fields.")
		return prevProperties
	}

	if hubspotSmartEventName.EventName == "" || hubspotSmartEventName.Filter == nil || hubspotSmartEventName.Type == "" {
		logCtx.Error("Missing smart event fileds.")
		return prevProperties
	}

	smartEventPayload, prevProperties, valid = GetHubspotSmartEventPayload(projectID, hubspotSmartEventName.EventName, docID,
		docType, currentProperties, prevProperties, hubspotSmartEventName.Filter)
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
	if timestampReferenceField == model.TimestampReferenceTypeDocument {
		smartEventTrackPayload.Timestamp = getEventTimestamp(defaultTimestamp) + 1
	} else {
		fieldTimestamp, err := getTimestampFromField(projectID, timestampReferenceField, currentProperties)
		if err != nil {
			logCtx.WithField("timestamp_refrence_field", timestampReferenceField).
				WithError(err).Errorf("Failed to get timestamp from reference field")
			smartEventTrackPayload.Timestamp = getEventTimestamp(defaultTimestamp) + 1 // use record timestamp if custom timestamp not available
		} else {
			if fieldTimestamp <= 0 {
				logCtx.WithField("timestamp_refrence_field", timestampReferenceField).
					WithError(err).Error("O timestamp from timestamp refrence field.")
				smartEventTrackPayload.Timestamp = getEventTimestamp(defaultTimestamp) + 1
			} else {
				smartEventTrackPayload.Timestamp = fieldTimestamp // make sure timestamp in seconds
			}

		}
	}

	if !C.IsDryRunCRMSmartEvent() {
		if usingFallbackUserID {
			logCtx.WithFields(log.Fields{"properties": smartEventPayload.Properties, "event_name": smartEventPayload.Name,
				"filter_exp":            *hubspotSmartEventName.Filter,
				"smart_event_timestamp": smartEventTrackPayload.Timestamp}).Warning("Smart event using fallback user id detected.")

		} else {
			status, _ := SDK.Track(projectID, smartEventTrackPayload, true, SDK.SourceHubspot)
			if status != http.StatusOK && status != http.StatusFound && status != http.StatusNotModified {
				logCtx.Error("Failed to create hubspot smart event")
			}
		}

	} else {
		logCtx.WithFields(log.Fields{"properties": smartEventPayload.Properties, "event_name": smartEventPayload.Name,
			"filter_exp":            *hubspotSmartEventName.Filter,
			"smart_event_timestamp": smartEventTrackPayload.Timestamp}).Info("Dry run smart event creation.")
	}

	return prevProperties
}

func GetHubspotPropertiesMeta(objectType string, apiKey string) ([]PropertyDetail, error) {
	if objectType == "" || apiKey == "" {
		return nil, errors.New("invalid parameters")
	}

	url := "https://" + "api.hubapi.com" + "/properties/v1/" + objectType + "/properties?hapikey=" + apiKey

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Timeout: 10 * time.Minute,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		var body interface{}
		json.NewDecoder(resp.Body).Decode(&body)
		return nil, fmt.Errorf("error while query data %s ", body)
	}

	var propertyDetails []PropertyDetail
	err = json.NewDecoder(resp.Body).Decode(&propertyDetails)

	if err != nil {
		return nil, err
	}

	return propertyDetails, nil
}

// CreateOrGetHubspotEventName makes sure event name exist
func CreateOrGetHubspotEventName(projectID uint64) int {
	logCtx := log.WithFields(log.Fields{"project_id": projectID})

	for i := range allowedEventNames {
		_, status := store.GetStore().CreateOrGetEventName(&model.EventName{
			ProjectId: projectID,
			Name:      allowedEventNames[i],
			Type:      model.TYPE_USER_CREATED_EVENT_NAME,
		})

		if status != http.StatusFound && status != http.StatusConflict && status != http.StatusCreated {
			logCtx.Error("Failed to create event name on SyncDatetimeAndNumericalProperties.")
			return http.StatusInternalServerError
		}

	}

	return http.StatusOK
}

func syncHubspotPropertyByType(projectID uint64, doctTypeAlias string, fieldName, fieldType string) error {

	logCtx := log.WithFields(log.Fields{"project_id": projectID, "doct_type_alias": doctTypeAlias, "field_name": fieldName, "field_type": fieldType})

	if projectID == 0 || doctTypeAlias == "" || fieldName == "" || fieldType == "" {
		logCtx.Error("Missing required fields.")
		return errors.New("missing required fields")
	}

	pType := model.GetHubspotMappedDataType(fieldType)

	enKey := model.GetCRMEnrichPropertyKeyByType(
		model.SmartCRMEventSourceHubspot,
		doctTypeAlias,
		U.GetPropertyValueAsString(fieldName),
	)

	if doctTypeAlias == model.HubspotDocumentTypeNameContact || doctTypeAlias == model.HubspotDocumentTypeNameCompany {
		eventName := U.EVENT_NAME_HUBSPOT_CONTACT_CREATED
		err := store.GetStore().CreateOrDeletePropertyDetails(projectID, eventName, enKey, pType, false, true)
		if err != nil {
			logCtx.WithFields(log.Fields{"enriched_property_key": enKey}).WithError(err).Error("Failed to create event property details.")
			return errors.New("failed to create created event property details")
		}

		eventName = U.EVENT_NAME_HUBSPOT_CONTACT_UPDATED
		err = store.GetStore().CreateOrDeletePropertyDetails(projectID, eventName, enKey, pType, false, true)
		if err != nil {
			logCtx.WithFields(log.Fields{"enriched_property_key": enKey}).Error("Failed to create updated event property details.")
			return errors.New("failed to create updated event property details")
		}

	} else if doctTypeAlias == model.HubspotDocumentTypeNameDeal {
		eventName := U.EVENT_NAME_HUBSPOT_DEAL_STATE_CHANGED
		err := store.GetStore().CreateOrDeletePropertyDetails(projectID, eventName, enKey, pType, false, true)
		if err != nil {
			logCtx.WithFields(log.Fields{"enriched_property_key": enKey}).WithError(err).Error("Failed to create event property details.")
			return errors.New("failed to create deal event property details")
		}
	}

	err := store.GetStore().CreateOrDeletePropertyDetails(projectID, "", enKey, pType, true, true)
	if err != nil {
		logCtx.WithFields(log.Fields{"enriched_property_key": enKey}).WithError(err).Error("Failed to create user property details.")
		return errors.New("failed to user property details")
	}

	return nil
}

// SyncDatetimeAndNumericalProperties sync datetime and numerical properties to the property_details table
func SyncDatetimeAndNumericalProperties(projectID uint64, apiKey string) (bool, []Status) {
	logCtx := log.WithFields(log.Fields{"project_id": projectID})

	if projectID == 0 || apiKey == "" {
		logCtx.Error("Missing required field.")
		return false, nil
	}

	status := CreateOrGetHubspotEventName(projectID)
	if status != http.StatusOK {
		logCtx.Error("Failed to CreateOrGetHubspotEventName.")
		return true, nil
	}

	var allStatus []Status
	anyFailures := false
	for docType, objectType := range *model.GetHubspotAllowedObjects(projectID) {
		propertiesMeta, err := GetHubspotPropertiesMeta(objectType, apiKey)
		if err != nil {
			logCtx.WithFields(log.Fields{"object_type": objectType}).WithError(err).Error("Failed to sync datetime and numerical properties.")
			continue
		}

		var status Status
		status.ProjectId = projectID
		status.Type = docType
		docTypeFailure := false
		for i := range propertiesMeta {
			fieldType := U.GetPropertyValueAsString(propertiesMeta[i].Type)
			if fieldType == "" {
				logCtx.Error("Failed to get property type field.")
				docTypeFailure = true
				continue
			}

			fieldName := U.GetPropertyValueAsString(propertiesMeta[i].Name)
			if fieldName == "" {
				logCtx.Error("Failed to get property name field.")
				docTypeFailure = true
				continue
			}

			label := U.GetPropertyValueAsString(propertiesMeta[i].Label)
			if label == "" {
				logCtx.Error("Failed to get property label")
			} else {
				err := store.GetStore().CreateOrUpdateDisplayNameByObjectType(projectID, model.GetCRMEnrichPropertyKeyByType(
					model.SmartCRMEventSourceHubspot,
					docType,
					fieldName,
				), docType, label, model.SmartCRMEventSourceHubspot)
				if err != http.StatusCreated {
					logCtx.Error("Failed to create or update display name")
				}
			}

			if failure := syncHubspotPropertyByType(projectID, docType, fieldName, fieldType); failure != nil {
				docTypeFailure = true
			}

		}

		if docTypeFailure {
			status.Status = U.CRM_SYNC_STATUS_FAILURES
			anyFailures = true
		} else {
			status.Status = U.CRM_SYNC_STATUS_SUCCESS
		}

		allStatus = append(allStatus, status)
	}

	return anyFailures, allStatus
}

func syncContact(projectID uint64, document *model.HubspotDocument, hubspotSmartEventNames []HubspotSmartEventName) int {
	logCtx := log.WithField("project_id",
		projectID).WithField("document_id", document.ID)

	enProperties, properties, err := getContactProperties(projectID, document)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get properites from hubspot contact.")
		return http.StatusInternalServerError
	}

	leadGUID, exists := (*enProperties)[model.UserPropertyHubspotContactLeadGUID]
	if !exists {
		logCtx.Error("Missing lead_guid on hubspot contact properties. Sync failed.")
		errCode := store.GetStore().UpdateHubspotDocumentAsSynced(
			projectID, document.ID, model.HubspotDocumentTypeContact, "", document.Timestamp, document.Action, "")
		if errCode != http.StatusAccepted {
			logCtx.Error("Failed to update hubspot contact document as synced.")
			return http.StatusInternalServerError
		}

		return http.StatusOK
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

		createdUserID, status := store.GetStore().CreateUser(&model.User{
			ProjectId:      projectID,
			JoinTimestamp:  getEventTimestamp(document.Timestamp),
			CustomerUserId: customerUserID})
		if status != http.StatusCreated {
			logCtx.WithField("status", status).Error("Failed to create user for hubspot contact created event.")
			return http.StatusInternalServerError
		}

		trackPayload.Name = U.EVENT_NAME_HUBSPOT_CONTACT_CREATED
		trackPayload.UserId = createdUserID

		status, response := SDK.Track(projectID, trackPayload, true, SDK.SourceHubspot)
		if status != http.StatusOK && status != http.StatusFound && status != http.StatusNotModified {
			logCtx.WithFields(log.Fields{"status": status, "track_response": response}).Error("Failed to track hubspot contact created event.")
			return http.StatusInternalServerError
		}

		userID = createdUserID
		eventID = response.EventId
	} else if document.Action == model.HubspotDocumentActionUpdated {
		trackPayload.Name = U.EVENT_NAME_HUBSPOT_CONTACT_UPDATED
		createdDocuments, status := store.GetStore().GetHubspotContactCreatedSyncIDAndUserID(projectID, document.ID)
		if status != http.StatusFound {
			if status != http.StatusMultipleChoices {
				logCtx.WithField("error_code", status).Error("Failed to get user from contact created document.")
				return http.StatusInternalServerError
			}

			previousUserID := ""
			for i := range createdDocuments {

				if previousUserID != "" && createdDocuments[i].UserId != "" &&
					createdDocuments[i].UserId != previousUserID {
					logCtx.Error("Multiple user id for contact created document found.")
					return http.StatusInternalServerError
				}
				previousUserID = createdDocuments[i].UserId
			}
		}

		if createdDocuments[0].UserId != "" {
			userID = createdDocuments[0].UserId
		} else {
			event, errCode := store.GetStore().GetEventById(projectID, createdDocuments[0].SyncId, "")
			if errCode != http.StatusFound {
				logCtx.WithField("event_id", createdDocuments[0].SyncId).Error(
					"Failed to get contact created event for getting user id.")
				return http.StatusInternalServerError
			}

			errCode = store.GetStore().UpdateHubspotDocumentAsSynced(
				projectID, document.ID, model.HubspotDocumentTypeContact, event.ID, createdDocuments[0].Timestamp, model.HubspotDocumentActionCreated, event.UserId)
			if errCode != http.StatusAccepted {
				logCtx.Error("Failed to update hubspot contact created document user id.")
			}

			userID = event.UserId
		}

		if customerUserID != "" {
			status, _ := SDK.Identify(projectID, &SDK.IdentifyPayload{
				UserId: userID, CustomerUserId: customerUserID}, false)
			if status != http.StatusOK {
				logCtx.WithField("customer_user_id", customerUserID).Error(
					"Failed to identify user on hubspot contact sync.")
				return http.StatusInternalServerError
			}
		} else {
			logCtx.Warning("Skipped user identification on hubspot contact sync. No customer_user_id on properties.")
		}

		trackPayload.UserId = userID
		status, response := SDK.Track(projectID, trackPayload, true, SDK.SourceHubspot)
		if status != http.StatusOK && status != http.StatusFound && status != http.StatusNotModified {
			logCtx.WithFields(log.Fields{"status": status, "track_response": response}).Error("Failed to track hubspot contact updated event.")
			return http.StatusInternalServerError
		}
		eventID = response.EventId

	} else {
		logCtx.Error("Invalid action on hubspot contact sync.")
		return http.StatusInternalServerError
	}

	var defaultSmartEventTimestamp int64
	if timestamp, err := model.GetHubspotDocumentUpdatedTimestamp(document); err != nil {
		logCtx.WithError(err).Warn("Failed to get last modified timestamp for smart event. Using document timestamp")
		defaultSmartEventTimestamp = document.Timestamp
	} else {
		defaultSmartEventTimestamp = timestamp
	}

	user, status := store.GetStore().GetUser(projectID, userID)
	if status != http.StatusFound {
		logCtx.WithField("error_code", status).Error("Failed to get user on sync contact.")
	}

	existingCustomerUserID := user.CustomerUserId

	if existingCustomerUserID != customerUserID {
		logCtx.WithFields(log.Fields{"existing_customer_user_id": existingCustomerUserID, "new_customer_user_id": customerUserID}).
			Warn("Different customer user id seen on sync contact")
	}

	var prevProperties *map[string]interface{}
	for i := range hubspotSmartEventNames {
		prevProperties = TrackHubspotSmartEvent(projectID, &hubspotSmartEventNames[i], eventID, document.ID, userID, document.Type,
			properties, prevProperties, defaultSmartEventTimestamp, false)
	}

	errCode := store.GetStore().UpdateHubspotDocumentAsSynced(
		projectID, document.ID, model.HubspotDocumentTypeContact, eventID, document.Timestamp, document.Action, userID)
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

	event, errCode := store.GetStore().GetEventById(projectID, contactDocument.SyncId, "")
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

	eventNames, errCode := store.GetStore().GetSmartEventFilterEventNames(projectID, false)
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
		logCtx.Warning("Skipped company sync. No contacts associated to company.")
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
					projectID, contactDocument.SyncId, "")
				if errCode == http.StatusFound {

					contactUser, status := store.GetStore().GetUser(projectID, contactSyncEvent.UserId)
					if status != http.StatusFound {
						logCtx.WithField("user_id", contactSyncEvent.UserId).Error(
							"Failed to get user by contact event user update user properites with company properties.")
						isContactsUpdateFailed = true
						continue
					}

					_, errCode := store.GetStore().UpdateUserProperties(projectID,
						contactUser.ID, userPropertiesJsonb, contactUser.PropertiesUpdatedTimestamp+1)
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
	errCode := store.GetStore().UpdateHubspotDocumentAsSynced(projectID, document.ID, model.HubspotDocumentTypeCompany, "", document.Timestamp, document.Action, "")
	if errCode != http.StatusAccepted {
		logCtx.Error("Failed to update hubspot deal document as synced.")
		return http.StatusInternalServerError
	}

	return http.StatusOK
}

func getHubspotMappedDataTypeValue(projectID uint64, eventName, enKey string, value interface{}) (interface{}, error) {
	if value == nil || value == "" {
		return nil, nil
	}

	if !C.IsEnabledPropertyDetailFromDB() || !C.IsEnabledPropertyDetailByProjectID(projectID) {
		return value, nil
	}

	ptype := store.GetStore().GetPropertyTypeByKeyValue(projectID, eventName, enKey, value, false)

	if ptype == U.PropertyTypeDateTime {
		datetime, err := U.GetPropertyValueAsFloat64(value)
		if err != nil {
			return nil, errors.New("failed to get datetime property")
		}

		return getEventTimestamp(int64(datetime)), nil

	}

	if ptype == U.PropertyTypeNumerical {
		num, err := U.GetPropertyValueAsFloat64(value)
		if err != nil {

			// try removing comma separated number
			cleanedValue := strings.ReplaceAll(U.GetPropertyValueAsString(value), ",", "")
			num, err := U.GetPropertyValueAsFloat64(cleanedValue)
			if err != nil {
				return nil, err
			}

			return num, nil
		}

		return num, nil
	}

	return value, nil
}

func getDealProperties(projectID uint64, document *model.HubspotDocument) (*map[string]interface{}, *map[string]interface{}, error) {

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
		enKey := model.GetCRMEnrichPropertyKeyByType(model.SmartCRMEventSourceHubspot,
			model.HubspotDocumentTypeNameDeal, k)
		value, err := getHubspotMappedDataTypeValue(projectID, U.EVENT_NAME_HUBSPOT_DEAL_STATE_CHANGED, enKey, v.Value)
		if err != nil {
			log.WithFields(log.Fields{"project_id": projectID, "property_key": enKey}).WithError(err).Error("Failed to get property value.")
			continue
		}

		enProperties[enKey] = value
		properties[k] = value

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

	enProperties, properties, err := getDealProperties(projectID, document)
	if err != nil {
		logCtx.Error("Failed to get hubspot deal document properties")
		return http.StatusInternalServerError
	}

	dealStage, dealstageExists := (*enProperties)[U.CRM_HUBSPOT_DEALSTAGE]

	userID := getDealUserID(projectID, &deal)

	eventID := ""
	if userID == "" {
		logCtx.Error("Skipped deal sync. No user associated to hubspot deal.")
	} else if !dealstageExists || dealStage == nil {
		logCtx.Error("No deal stage property found on hubspot deal.")
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

		var defaultSmartEventTimestamp int64
		if timestamp, err := model.GetHubspotDocumentUpdatedTimestamp(document); err != nil {
			logCtx.WithError(err).Warn("Failed to get last modified timestamp for smart event. Using document timestamp")
			defaultSmartEventTimestamp = document.Timestamp
		} else {
			defaultSmartEventTimestamp = timestamp
		}

		eventID = response.EventId
		var prevProperties *map[string]interface{}
		for i := range hubspotSmartEventNames {
			prevProperties = TrackHubspotSmartEvent(projectID, &hubspotSmartEventNames[i], response.EventId, document.ID, userID, document.Type,
				properties, prevProperties, defaultSmartEventTimestamp, false)
		}
	}

	errCode := store.GetStore().UpdateHubspotDocumentAsSynced(projectID,
		document.ID, model.HubspotDocumentTypeDeal, eventID, document.Timestamp, document.Action, userID)
	if errCode != http.StatusAccepted {
		logCtx.Error("Failed to update hubspot deal document as synced.")
		return http.StatusInternalServerError
	}

	return http.StatusOK
}

// GetBatchedOrderedDocumentsByID return list of document in batches. Order is maintained on document id.
func GetBatchedOrderedDocumentsByID(documents []model.HubspotDocument, batchSize int) []map[string][]model.HubspotDocument {

	if len(documents) < 0 {
		return nil
	}

	documentsMap := make(map[string][]model.HubspotDocument)
	for i := range documents {
		if _, exist := documentsMap[documents[i].ID]; !exist {
			documentsMap[documents[i].ID] = make([]model.HubspotDocument, 0)
		}
		documentsMap[documents[i].ID] = append(documentsMap[documents[i].ID], documents[i])
	}

	batchedDocumentsByID := make([]map[string][]model.HubspotDocument, 1)
	isBatched := make(map[string]bool)
	batchLen := 0
	batchedDocumentsByID[batchLen] = make(map[string][]model.HubspotDocument)
	for i := range documents {
		if isBatched[documents[i].ID] {
			continue
		}

		if len(batchedDocumentsByID[batchLen]) >= batchSize {
			batchedDocumentsByID = append(batchedDocumentsByID, make(map[string][]model.HubspotDocument))
			batchLen++
		}

		batchedDocumentsByID[batchLen][documents[i].ID] = documentsMap[documents[i].ID]
		isBatched[documents[i].ID] = true
	}

	return batchedDocumentsByID
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

// GetHubspotTimeSeriesByStartTimestamp returns time series for batch processing -> {Day1,Day2}, {Day2,Day3},{Day3,Day4} upto current day
func GetHubspotTimeSeriesByStartTimestamp(projectID uint64, from int64) [][]int64 {
	logCtx := log.WithFields(log.Fields{"project_id": projectID, "from": from})
	if from < 1 {
		logCtx.Error("Invalid timestamp from batch processing by day.")
		return nil
	}

	timeSeries := [][]int64{}
	startTime := time.Unix(from/1000, 0)
	startDate := time.Date(startTime.Year(), startTime.Month(), startTime.Day(), 0, 0, 0, 0, time.UTC)
	currentTime := time.Now()
	for ; startDate.Unix() < currentTime.Unix(); startDate = startDate.AddDate(0, 0, 1) {
		timeSeries = append(timeSeries, []int64{startTime.Unix() * 1000, startDate.AddDate(0, 0, 1).Unix() * 1000})
		startTime = startDate.AddDate(0, 0, 1)
	}

	return timeSeries
}

type syncWorkerStatus struct {
	HasFailure bool
	Lock       sync.Mutex
}

// syncAllWorker is a wrapper over syncAll function for providing concurrency
func syncAllWorker(projectID uint64, wg *sync.WaitGroup, syncStatus *syncWorkerStatus, documents []model.HubspotDocument, hubspotSmartEventNames []HubspotSmartEventName) {
	defer wg.Done()

	errCode := syncAll(projectID, documents, hubspotSmartEventNames)

	syncStatus.Lock.Lock()
	defer syncStatus.Lock.Unlock()
	if errCode != http.StatusOK {
		syncStatus.HasFailure = true
	}
}

// Sync - Syncs hubspot documents in an order of type.
func Sync(projectID uint64, workersPerProject int) ([]Status, bool) {
	logCtx := log.WithField("project_id", projectID)

	statusByProjectAndType := make([]Status, 0, 0)
	hubspotSmartEventNames := GetHubspotSmartEventNames(projectID)
	status := CreateOrGetHubspotEventName(projectID)
	if status != http.StatusOK {
		statusByProjectAndType = append(statusByProjectAndType, Status{ProjectId: projectID,
			Status: "Failed to create event names"})
		return statusByProjectAndType, true
	}

	var orderedTimeSeries [][]int64
	minTimestamp, errCode := store.GetStore().GetHubspotDocumentBeginingTimestampByDocumentTypeForSync(projectID)
	if errCode != http.StatusFound {
		if errCode == http.StatusNotFound {
			statusByProjectAndType = append(statusByProjectAndType, Status{ProjectId: projectID,
				Status: U.CRM_SYNC_STATUS_SUCCESS})
			return statusByProjectAndType, false
		}

		logCtx.WithField("err_code", errCode).Error("Failed to get time series.")
		statusByProjectAndType = append(statusByProjectAndType, Status{ProjectId: projectID,
			Status: "Failed to get time series."})
		return statusByProjectAndType, true
	}

	if workersPerProject > 1 {
		orderedTimeSeries = GetHubspotTimeSeriesByStartTimestamp(projectID, minTimestamp)
	} else {
		// generate single time series
		orderedTimeSeries = append(orderedTimeSeries, []int64{minTimestamp, time.Now().Unix() * 1000})
	}

	anyFailure := false
	overAllSyncStatus := make(map[string]bool)
	for _, timeRange := range orderedTimeSeries {

		for i := range syncOrderByType {

			logCtx = logCtx.WithFields(log.Fields{"type": syncOrderByType[i], "time_range": timeRange})

			logCtx.Info("Processing started for given time range")
			var documents []model.HubspotDocument
			var errCode int
			if workersPerProject > 1 {
				documents, errCode = store.GetStore().GetHubspotDocumentsByTypeANDRangeForSync(projectID, syncOrderByType[i], timeRange[0], timeRange[1])
			} else {
				documents, errCode = store.GetStore().
					GetHubspotDocumentsByTypeForSync(projectID, syncOrderByType[i])
			}

			if errCode != http.StatusFound {
				logCtx.WithFields(log.Fields{"time_range": timeRange, "doc_type": syncOrderByType[i]}).Error("Failed to get hubspot document by type for sync.")
				continue
			}

			docTypeAlias := model.GetHubspotTypeAliasByType(syncOrderByType[i])

			batches := GetBatchedOrderedDocumentsByID(documents, workersPerProject)

			var syncStatus syncWorkerStatus
			var workerIndex int
			for bi := range batches {
				batch := batches[bi]
				var wg sync.WaitGroup
				for docID := range batch {
					logCtx.WithFields(log.Fields{"worker": workerIndex, "doc_id": docID, "type": syncOrderByType[i]}).Info("Processing Batch by doc_id")
					workerIndex++
					wg.Add(1)
					go syncAllWorker(projectID, &wg, &syncStatus, batch[docID], (*hubspotSmartEventNames)[docTypeAlias])
				}
				wg.Wait()
			}

			if _, exist := overAllSyncStatus[docTypeAlias]; !exist {
				overAllSyncStatus[docTypeAlias] = false
			}

			if syncStatus.HasFailure {
				overAllSyncStatus[docTypeAlias] = true
			}

			logCtx.Info("Processing completed for given time range")
		}
	}

	for docTypeAlias, failure := range overAllSyncStatus {
		status := Status{ProjectId: projectID,
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
