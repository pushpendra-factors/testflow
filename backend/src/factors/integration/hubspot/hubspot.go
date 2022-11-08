package hubspot

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	C "factors/config"
	"factors/model/model"
	"factors/model/store"
	"factors/sdk"
	SDK "factors/sdk"
	"factors/util"
	U "factors/util"
)

// Version definition
type Version struct {
	Name      string `json:"version"`
	Value     string `json:"value"`
	Timestamp int64  `json:"timestamp"`
}

// Property definition
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
	Timestamp int64       `json:"timestamp"`
}

// ContactIdentityProfile for contact
type ContactIdentityProfile struct {
	Identities []ContactIdentity `json:"identities"`
}
type Engagements struct {
	Engagement   map[string]interface{}   `json:"engagement"`
	Associations map[string][]interface{} `json:"associations"`
	Metadata     map[string]interface{}   `json:"metadata"`
}

// Contact definition
type Contact struct {
	Vid              int64                    `json:"vid"`
	Properties       map[string]Property      `json:"properties"`
	IdentityProfiles []ContactIdentityProfile `json:"identity-profiles"`
	FormSubmissions  []map[string]interface{} `json:"form-submissions"`
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

type ContactListMembership struct {
	StaticListId   int64 `json:"static-list-id"`
	InternalListId int64 `json:"internal-list-id"`
	Timestamp      int64 `json:"timestamp"`
	Vid            int64 `json:"vid"`
	IsMember       bool  `json:"is-member"`
}

// ContactList definition
type NewContactList struct {
	ListId          int64                            `json:"listId"`
	ListName        string                           `json:"name"`
	ListType        string                           `json:"listType"`
	ListCreatedAt   int64                            `json:"createdAt"`
	ContactIds      []int64                          `json:"contactIds"`
	ListMemberships map[string]ContactListMembership `json:"listMemberships"`
}

type OldContactList struct {
	ListId          int64                              `json:"listId"`
	ListName        string                             `json:"name"`
	ListType        string                             `json:"listType"`
	ListCreatedAt   int64                              `json:"createdAt"`
	ContactIds      []int64                            `json:"contactIds"`
	ListMemberships map[string][]ContactListMembership `json:"listMemberships"`
}

// PropertyDetail definition for hubspot properties api
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
	model.HubspotDocumentTypeEngagement,
	model.HubspotDocumentTypeContactList,
}

func GetHubspotObjectTypeForSync() []int {
	return syncOrderByType[:]
}

func GetDecodedValue(encodedValue string, limit int) string {
	prevValue := encodedValue
	for i := 0; i <= limit; i++ {
		curr_value, err := url.QueryUnescape(prevValue)
		if err != nil || curr_value == prevValue {
			if err != nil {
				log.WithField("encodedValue", encodedValue).Error("error while decoding")
			}
			return prevValue
		}
		if i == limit && prevValue != curr_value {
			log.WithField("encodedValue", encodedValue).Error("limit exceeded on decoding")
			return prevValue
		}
		prevValue = curr_value
	}

	return prevValue
}

func GetURLParameterAsMap(pageUrl string) map[string]interface{} {
	u, err := url.Parse(pageUrl)
	if err != nil {
		log.Error(err)
		return nil
	}
	queries := u.Query()

	urlParameters := make(map[string]interface{})
	for key, value := range queries {
		if _, exists := urlParameters[key]; !exists {
			for _, v := range value {
				urlParameters[key] = GetDecodedValue(v, 2)
			}
		}
	}
	return urlParameters
}

func extractingFormSubmissionDetails(projectId int64, contact Contact, properties map[string]interface{}) []map[string]interface{} {
	form := make([]map[string]interface{}, 0)
	keyArr := []string{"conversion-id", "form-id", "form-type", "page-title", "page-url", "portal-id", "timestamp", "title"}

	for userFormNo := range contact.FormSubmissions {
		form = append(form, make(map[string]interface{}))

		for idx := range keyArr {
			if contact.FormSubmissions[userFormNo][keyArr[idx]] == nil {
				continue
			}

			if keyArr[idx] == "timestamp" {
				if _, exists := form[userFormNo][keyArr[idx]]; !exists {
					val := contact.FormSubmissions[userFormNo][keyArr[idx]]
					vfloat, _ := util.GetPropertyValueAsFloat64(val)

					form[userFormNo][keyArr[idx]] = (int64)(vfloat / 1000)
				}
			} else if keyArr[idx] == "page-url" {
				val := contact.FormSubmissions[userFormNo][keyArr[idx]]
				form[userFormNo][keyArr[idx]] = val
			} else {
				if _, exists := form[userFormNo][keyArr[idx]]; !exists {
					val := contact.FormSubmissions[userFormNo][keyArr[idx]]
					form[userFormNo][keyArr[idx]] = val
				}
			}
		}
		key, value := getCustomerUserIDFromProperties(projectId, properties)
		if _, exists := form[userFormNo][key]; !exists {
			form[userFormNo][key] = value
		}
	}
	return form
}

func syncContactFormSubmissions(project *model.Project, userId string, document *model.HubspotDocument) {
	logFields := log.Fields{
		"project":  project,
		"user_id":  userId,
		"document": document,
	}

	logCtx := log.WithFields(logFields)
	if userId == "" {
		log.Error("syncContactFormSubmissions Failed. Invalid userId")
		return
	}

	var contact Contact
	err := json.Unmarshal((document.Value).RawMessage, &contact)
	if err != nil {
		logCtx.Error("Error occured during unmarshal of hubspot document")
		return
	}

	enProperties, _, _, _, err := GetContactProperties(project.ID, document)
	if err != nil {
		return
	}

	form := extractingFormSubmissionDetails(project.ID, contact, *enProperties)

	if len(form) == 0 {
		return
	}

	var timestamps []interface{}
	for i := range form {
		timestamps = append(timestamps, form[i]["timestamp"])
	}

	events, status := store.GetStore().GetHubspotFormEvents(project.ID, userId, timestamps)
	if status == http.StatusInternalServerError {
		logCtx.Error("Internal server error")
		return
	}

	for idx := range form {
		encodeProperties := make(map[string]interface{}, 0)
		formID := form[idx]["form-id"]
		conversionID := form[idx]["conversion-id"]
		eventTimestamp := form[idx]["timestamp"].(int64)

		eventExists := false
		for i := range events {
			if events[i].Timestamp == eventTimestamp {
				propertiesMap := make(map[string]interface{})
				err := json.Unmarshal(events[i].Properties.RawMessage, &propertiesMap)
				if err != nil {
					log.Error("Error occured during unmarshal of hubspot document")
					return
				}

				encodeFormId := model.GetCRMEnrichPropertyKeyByType(model.SmartCRMEventSourceHubspot, model.HubspotDocumentTypeNameFormSubmission, "form-id")
				encodeConversionId := model.GetCRMEnrichPropertyKeyByType(model.SmartCRMEventSourceHubspot, model.HubspotDocumentTypeNameFormSubmission, "conversion-id")

				if propertiesMap[encodeFormId] == formID && propertiesMap[encodeConversionId] == conversionID {
					eventExists = true
					break
				}
			}
		}
		if eventExists {
			continue
		}

		for key, val := range form[idx] {
			if key == "page-url" {
				urlParameters := GetURLParameterAsMap(util.GetPropertyValueAsString(val))
				for k, v := range urlParameters {
					encodeProperties[k] = v
				}

				url, err := util.ParseURLStable(util.GetPropertyValueAsString(val))
				if err != nil {
					log.WithField("val", val).Error("Error occured while ParseURLStable.")
					continue
				}
				enkey := model.GetCRMEnrichPropertyKeyByType(model.SmartCRMEventSourceHubspot, model.HubspotDocumentTypeNameFormSubmission, "page-url-no-qp")
				encodeProperties[enkey] = util.GetURLHostAndPath(url)
			}
			enkey := model.GetCRMEnrichPropertyKeyByType(model.SmartCRMEventSourceHubspot, model.HubspotDocumentTypeNameFormSubmission, key)
			encodeProperties[enkey] = val
		}

		payload := &SDK.TrackPayload{
			ProjectId:       project.ID,
			Name:            U.EVENT_NAME_HUBSPOT_CONTACT_FORM_SUBMISSION,
			EventProperties: encodeProperties,
			UserId:          userId,
			Timestamp:       eventTimestamp,
		}

		status, _ := sdk.Track(project.ID, payload, true, SDK.SourceHubspot, "")
		if status != http.StatusOK && status != http.StatusFound && status != http.StatusNotModified {
			logCtx.Error("Failed to create hubspot form-submission event")
			return
		}

	}
}

func fillDatePropertiesAndTimeZone(documents []model.HubspotDocument, dateProperties *map[string]bool, timeZone U.TimeZoneString) {
	for i := range documents {
		documents[i].SetDateProperties(dateProperties)
		documents[i].SetTimeZone(timeZone)
	}
}

func GetHubspotPropertiesByDataType(projectId int64, docTypeAPIObjects *map[string]string, apiKey, refreshToken, dataType string) (map[int]*map[string]bool, error) {
	propertiesByObjectType := make(map[int]*map[string]bool)
	for typeAlias, apiObjectName := range *docTypeAPIObjects {
		propertiesMeta, err := GetHubspotPropertiesMeta(apiObjectName, apiKey, refreshToken)
		if err != nil {
			log.WithFields(log.Fields{"project_id": projectId, "api_object_name": apiObjectName, "doc_Type": typeAlias}).
				WithError(err).Error("Failed to get hubspot properties meta.")
			return nil, err
		}

		dataTypeProperties := make(map[string]bool)
		for _, property := range propertiesMeta {
			if property.Type == dataType {
				dataTypeProperties[property.Name] = true
			}
		}
		docType, err := model.GetHubspotTypeByAlias(typeAlias)
		if err != nil {
			return nil, err
		}

		propertiesByObjectType[docType] = &dataTypeProperties
	}

	return propertiesByObjectType, nil
}

func GetContactProperties(projectID int64, document *model.HubspotDocument) (*map[string]interface{}, *map[string]interface{}, []string, string, error) {
	if document.Type != model.HubspotDocumentTypeContact {
		return nil, nil, nil, "", errors.New("invalid type")
	}

	var contact Contact
	err := json.Unmarshal((document.Value).RawMessage, &contact)
	if err != nil {
		return nil, nil, nil, "", err
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

			if !C.AllowIdentificationOverwriteUsingSource(projectID) {
				if _, exists := enrichedProperties[enkey]; !exists {
					enrichedProperties[enkey] = contact.IdentityProfiles[ipi].Identities[idi].Value
				}

				if _, exists := properties[key]; !exists {
					properties[key] = contact.IdentityProfiles[ipi].Identities[idi].Value
				}
				continue
			}

			if key != "EMAIL" {
				if _, exists := enrichedProperties[enkey]; !exists {
					enrichedProperties[enkey] = contact.IdentityProfiles[ipi].Identities[idi].Value
				}

				if _, exists := properties[key]; !exists {
					properties[key] = contact.IdentityProfiles[ipi].Identities[idi].Value
				}
				continue
			}

			// store primary email in contact properties
			if _, exists := enrichedProperties[enkey]; !exists || contact.IdentityProfiles[ipi].Identities[idi].IsPrimary {
				enrichedProperties[enkey] = contact.IdentityProfiles[ipi].Identities[idi].Value
				properties[key] = contact.IdentityProfiles[ipi].Identities[idi].Value
			}

		}
	}

	for pkey, pvalue := range contact.Properties {
		enKey := model.GetCRMEnrichPropertyKeyByType(model.SmartCRMEventSourceHubspot,
			model.HubspotDocumentTypeNameContact, pkey)
		value, err := getHubspotMappedDataTypeValue(projectID, U.EVENT_NAME_HUBSPOT_CONTACT_UPDATED, enKey, pvalue.Value, model.HubspotDocumentTypeContact, document.GetDateProperties(), string(document.GetTimeZone()))
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

	if C.AllowIdentificationOverwriteUsingSource(projectID) {
		allIdentities := make([]ContactIdentity, 0)
		for ipi := range contact.IdentityProfiles {
			allIdentities = append(allIdentities, contact.IdentityProfiles[ipi].Identities...)
		}

		sort.Slice(allIdentities, func(i, j int) bool { return allIdentities[i].Timestamp > allIdentities[j].Timestamp })

		secondaryEmails := make([]string, 0)
		primaryEmail := ""
		for i := range allIdentities {
			if allIdentities[i].Type != "EMAIL" {
				continue
			}

			if allIdentities[i].IsPrimary {
				primaryEmail = U.GetPropertyValueAsString(allIdentities[i].Value)
				continue
			}

			secondaryEmails = append(secondaryEmails, U.GetPropertyValueAsString(allIdentities[i].Value))
		}

		emailKey := model.GetCRMEnrichPropertyKeyByType(model.SmartCRMEventSourceHubspot,
			model.HubspotDocumentTypeNameContact, "email")
		if primaryEmail != "" {
			enrichedProperties[emailKey] = primaryEmail
			properties["EMAIL"] = primaryEmail
		} else if len(secondaryEmails) > 0 {
			enrichedProperties[emailKey] = secondaryEmails[0]
			properties["EMAIL"] = secondaryEmails[0]
		}

		return &enrichedProperties, &properties, secondaryEmails, primaryEmail, nil
	}

	return &enrichedProperties, &properties, nil, "", nil
}

func getCustomerUserIDFromProperties(projectID int64, properties map[string]interface{}) (string, string) {
	// identify using email if exist on properties.
	emailInt, emailExists := properties[model.GetCRMEnrichPropertyKeyByType(model.SmartCRMEventSourceHubspot,
		model.HubspotDocumentTypeNameContact, "email")]
	if emailExists || emailInt != nil {
		email, ok := emailInt.(string)
		if ok && email != "" {
			return "email", U.GetEmailLowerCase(email)
		}
	}

	// identify using phone if exist on properties.
	phoneInt := properties[model.GetCRMEnrichPropertyKeyByType(model.SmartCRMEventSourceHubspot,
		model.HubspotDocumentTypeNameContact, "phone")]
	if phoneInt != nil {
		phone := U.GetPropertyValueAsString(phoneInt)
		if U.IsValidPhone(phone) {
			identifiedPhone, _ := store.GetStore().GetUserIdentificationPhoneNumber(projectID, phone)
			if identifiedPhone != "" {
				return "phone", identifiedPhone
			}
		}
	}

	// other possible phone keys.
	for key := range properties {
		hasPhone := strings.Index(key, "phone")
		if hasPhone > -1 {
			phone := U.GetPropertyValueAsString(properties[key])
			if U.IsValidPhone(phone) {
				identifiedPhone, _ := store.GetStore().GetUserIdentificationPhoneNumber(projectID, phone)
				if identifiedPhone != "" {
					return "phone", identifiedPhone
				}
			}
		}
	}

	return "", ""
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
func GetHubspotSmartEventPayload(projectID int64, eventName, docID string,
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
				_, prevProperties, _, _, err = GetContactProperties(projectID, prevDoc)
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

func getTimestampFromField(projectID int64, propertyName string, properties *map[string]interface{}) (int64, error) {
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
func TrackHubspotSmartEvent(projectID int64, hubspotSmartEventName *HubspotSmartEventName, eventID, docID, userID string, docType int,
	currentProperties, prevProperties *map[string]interface{}, defaultTimestamp int64, usingFallbackUserID bool) *map[string]interface{} {
	var valid bool
	var smartEventPayload *model.CRMSmartEvent

	logCtx := log.WithFields(log.Fields{"project_id": projectID, "doc_type": docType})

	if projectID == 0 || userID == "" || docType == 0 || currentProperties == nil || defaultTimestamp == 0 {
		logCtx.Error("Missing required fields.")
		return prevProperties
	}

	if hubspotSmartEventName.EventName == "" || hubspotSmartEventName.Filter == nil || hubspotSmartEventName.Type == "" {
		logCtx.Error("Missing smart event fields.")
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
		RequestSource:   model.UserSourceHubspot,
	}

	timestampReferenceField := hubspotSmartEventName.Filter.TimestampReferenceField
	if timestampReferenceField == model.TimestampReferenceTypeDocument {
		smartEventTrackPayload.Timestamp = getEventTimestamp(defaultTimestamp) + 1
	} else {
		fieldTimestamp, err := getTimestampFromField(projectID, timestampReferenceField, currentProperties)
		if err != nil {
			logCtx.WithField("timestamp_reference_field", timestampReferenceField).
				WithError(err).Errorf("Failed to get timestamp from reference field")
			smartEventTrackPayload.Timestamp = getEventTimestamp(defaultTimestamp) + 1 // use record timestamp if custom timestamp not available
		} else {
			if fieldTimestamp <= 0 {
				logCtx.WithField("timestamp_reference_field", timestampReferenceField).
					WithError(err).Error("O timestamp from timestamp reference field.")
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
			status, _ := SDK.Track(projectID, smartEventTrackPayload, true, SDK.SourceHubspot, "")
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

func GetHubspotPropertiesMeta(objectType string, apiKey, refreshToken string) ([]PropertyDetail, error) {
	if objectType == "" {
		return nil, errors.New("invalid parameters")
	}

	if apiKey == "" && refreshToken == "" {
		return nil, errors.New("missing api key and refresh token")
	}

	var accessToken string
	var err error
	if refreshToken != "" {
		accessToken, err = model.GetHubspotAccessToken(refreshToken, C.GetHubspotAppID(), C.GetHubspotAppSecret())
		if err != nil {
			return nil, err
		}
		apiKey = ""
	}

	url := "https://" + "api.hubapi.com" + "/properties/v1/" + objectType + "/properties?"
	resp, err := model.ActionHubspotRequestHandler("GET", url, apiKey, accessToken, "", nil)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		var body interface{}
		json.NewDecoder(resp.Body).Decode(&body)
		log.WithFields(log.Fields{"resp_body": body}).Error("Failed to GetHubspotPropertiesMeta.")
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
func CreateOrGetHubspotEventName(projectID int64) int {
	logCtx := log.WithFields(log.Fields{"project_id": projectID})

	for i := range model.AllowedEventNamesForHubspot {
		_, status := store.GetStore().CreateOrGetEventName(&model.EventName{
			ProjectId: projectID,
			Name:      model.AllowedEventNamesForHubspot[i],
			Type:      model.TYPE_USER_CREATED_EVENT_NAME,
		})

		if status != http.StatusFound && status != http.StatusConflict && status != http.StatusCreated {
			logCtx.Error("Failed to create event name on SyncDatetimeAndNumericalProperties.")
			return http.StatusInternalServerError
		}

	}

	if C.IsAllowedHubspotGroupsByProjectID(projectID) {
		_, status := store.GetStore().CreateGroup(projectID, model.GROUP_NAME_HUBSPOT_COMPANY, model.AllowedGroupNames)
		if status != http.StatusCreated && status != http.StatusConflict {
			return http.StatusInternalServerError
		}

		_, status = store.GetStore().CreateGroup(projectID, model.GROUP_NAME_HUBSPOT_DEAL, model.AllowedGroupNames)
		if status != http.StatusCreated && status != http.StatusConflict {
			return http.StatusInternalServerError
		}
	}

	return http.StatusOK
}

func syncHubspotPropertyByType(projectID int64, doctTypeAlias string, fieldName, fieldType string) error {

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
func SyncDatetimeAndNumericalProperties(projectID int64, apiKey, refreshToken string) (bool, []Status) {
	logCtx := log.WithFields(log.Fields{"project_id": projectID})

	if projectID == 0 {
		logCtx.Error("Missing project_id.")
		return true, nil
	}

	if apiKey == "" && refreshToken == "" {
		logCtx.Error("Missing api key and refresh token.")
		return true, nil
	}

	status := CreateOrGetHubspotEventName(projectID)
	if status != http.StatusOK {
		logCtx.Error("Failed to CreateOrGetHubspotEventName.")
		return true, nil
	}

	var allStatus []Status
	anyFailures := false
	for docType, objectType := range *model.GetHubspotAllowedObjects(projectID) {
		propertiesMeta, err := GetHubspotPropertiesMeta(objectType, apiKey, refreshToken)
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
				if err != http.StatusCreated && err != http.StatusConflict {
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

func syncContact(project *model.Project, document *model.HubspotDocument, hubspotSmartEventNames []HubspotSmartEventName) int {
	logCtx := log.WithField("project_id", project.ID).WithField("document_id", document.ID)

	if document.Type != model.HubspotDocumentTypeContact {
		logCtx.Error("Invalid contact document.")
		return http.StatusInternalServerError
	}

	if document.Action == model.HubspotDocumentActionDeleted {
		contactDocuments, status := store.GetStore().GetHubspotDocumentByTypeAndActions(project.ID, []string{document.ID}, model.HubspotDocumentTypeContact, []int{model.HubspotDocumentActionCreated})
		if status != http.StatusFound {
			logCtx.Error(
				"Failed to get hubspot documents by type and action on sync contact, action delete.")
			return http.StatusInternalServerError
		}
		userProperties := make(map[string]interface{})
		keyDelete := model.GetCRMEnrichPropertyKeyByType(model.SmartCRMEventSourceHubspot,
			model.HubspotDocumentTypeNameContact, "deleted")
		userProperties[keyDelete] = true
		userPropertiesJsonb, err := U.EncodeToPostgresJsonb(&userProperties)
		if err != nil {
			logCtx.WithError(err).Error("Failed to marshal company properties to Jsonb, in sync contact, action delete.")
			return http.StatusInternalServerError
		}

		deleteContactUserID := contactDocuments[0].UserId
		if deleteContactUserID == "" {
			event, errCode := store.GetStore().GetEventById(project.ID, contactDocuments[0].SyncId, "")
			if errCode != http.StatusFound {
				logCtx.WithField("delete_contact", contactDocuments[0].ID).Error(
					"Failed to get merged contact created event for getting user id.")
				return http.StatusInternalServerError
			}
			deleteContactUserID = event.UserId
		}

		_, errCode := store.GetStore().UpdateUserProperties(project.ID, deleteContactUserID, userPropertiesJsonb,
			document.Timestamp)
		if errCode != http.StatusAccepted && errCode != http.StatusNotModified {
			logCtx.WithField("UserID", contactDocuments[0].UserId).WithField("userPropertiesJsonb", userPropertiesJsonb).Error("Failed to update user properties for contact delete action")
			return http.StatusInternalServerError
		}
		errCode = store.GetStore().UpdateHubspotDocumentAsSynced(
			project.ID, document.ID, model.HubspotDocumentTypeContact, "", document.Timestamp, document.Action, contactDocuments[0].UserId, "")
		if errCode != http.StatusAccepted {
			logCtx.Error("Failed to update hubspot contact document as synced, contact deleted document.")
			return http.StatusInternalServerError
		}
		return http.StatusOK
	}

	// process merged contact only in updated record
	if document.Action == model.HubspotDocumentActionUpdated {
		value, err := U.DecodePostgresJsonb(document.Value)
		if err != nil {
			logCtx.WithField("document.Value", document.Value).Error("Failed to decode hubspot Json document-Value.")
			return http.StatusInternalServerError
		}

		_, exists := (*value)["merged-vids"]
		if exists {
			var mergedVIDs []string
			for _, vInt := range (*value)["merged-vids"].([]interface{}) {
				vfloat, err := util.GetPropertyValueAsFloat64(vInt)
				if err != nil {
					logCtx.WithError(err).Error("Failed to convert contact id to float.")
					continue
				}
				mergedVIDs = append(mergedVIDs, fmt.Sprintf("%v", int64(vfloat)))
			}

			if len(mergedVIDs) != 0 {
				mergeContactDocuments, status := store.GetStore().GetHubspotDocumentByTypeAndActions(project.ID, mergedVIDs, model.HubspotDocumentTypeContact, []int{model.HubspotDocumentActionCreated})
				if status != http.StatusFound {
					logCtx.Error("Failed to get hubspot documents by type and action on sync contact.")
					return http.StatusInternalServerError
				}
				for _, mergedContact := range mergeContactDocuments {
					if mergedContact.ID != fmt.Sprintf("%v", (*value)["canonical-vid"]) {
						mergeUserProperties := make(map[string]interface{})
						keyMerge := model.GetCRMEnrichPropertyKeyByType(model.SmartCRMEventSourceHubspot,
							model.HubspotDocumentTypeNameContact, "merged")
						mergeUserProperties[keyMerge] = true
						keyPrimaryContact := model.GetCRMEnrichPropertyKeyByType(model.SmartCRMEventSourceHubspot,
							model.HubspotDocumentTypeNameContact, "primary_contact")
						mergeUserProperties[keyPrimaryContact] = (*value)["canonical-vid"]
						mergeUserPropertiesJsonb, err := U.EncodeToPostgresJsonb(&mergeUserProperties)
						if err != nil {
							logCtx.WithError(err).Error("Failed to marshal merged contact properties to Jsonb, in sync contact.")
							return http.StatusInternalServerError
						}

						mergedContactUserID := mergedContact.UserId
						if mergedContactUserID == "" {
							event, errCode := store.GetStore().GetEventById(project.ID, mergedContact.SyncId, "")
							if errCode != http.StatusFound {
								logCtx.WithField("merged_contact", mergedContact.ID).Error(
									"Failed to get merged contact created event for getting user id.")
								continue
							}
							mergedContactUserID = event.UserId
						}

						_, errCode := store.GetStore().UpdateUserProperties(project.ID, mergedContactUserID,
							mergeUserPropertiesJsonb, document.Timestamp)
						if errCode != http.StatusAccepted && errCode != http.StatusNotModified {
							logCtx.WithField("UserID", mergedContact.UserId).WithField("userPropertiesJsonb", mergeUserPropertiesJsonb).Error("Failed to update user properties")
							return http.StatusInternalServerError
						}
					}
				}
			}
		}
	}

	enProperties, properties, secondaryEmails, primaryEmail, err := GetContactProperties(project.ID, document)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get properites from hubspot contact.")
		return http.StatusInternalServerError
	}

	if primaryEmail == "" && len(secondaryEmails) > 0 {
		primaryEmail = secondaryEmails[0]
	}

	leadGUID, exists := (*enProperties)[model.UserPropertyHubspotContactLeadGUID]
	if !exists {
		logCtx.Error("Missing lead_guid on hubspot contact properties. Sync failed.")
		errCode := store.GetStore().UpdateHubspotDocumentAsSynced(
			project.ID, document.ID, model.HubspotDocumentTypeContact, "", document.Timestamp, document.Action, "", "")
		if errCode != http.StatusAccepted {
			logCtx.Error("Failed to update hubspot contact document as synced.")
			return http.StatusInternalServerError
		}

		return http.StatusOK
	}

	trackPayload := &SDK.TrackPayload{
		ProjectId:       project.ID,
		EventProperties: *enProperties,
		UserProperties:  *enProperties,
		Timestamp:       getEventTimestamp(document.Timestamp),
		RequestSource:   model.UserSourceHubspot,
	}

	logCtx = logCtx.WithField("action", document.Action).WithField(
		model.UserPropertyHubspotContactLeadGUID, leadGUID)

	_, customerUserID := getCustomerUserIDFromProperties(project.ID, *enProperties)

	emails := []string{}
	userByCustomerUserID := make(map[string]string)
	if C.AllowIdentificationOverwriteUsingSource(project.ID) {
		emails = append([]string{primaryEmail}, secondaryEmails...)

		usersCustomerUserID, status := store.GetStore().GetExistingUserByCustomerUserID(project.ID, emails, model.UserSourceWeb)
		if status != http.StatusNotFound && status != http.StatusFound {
			logCtx.Error("Failed to get users by customer user id.")
			return http.StatusInternalServerError
		}

		if status == http.StatusFound {
			userByCustomerUserID = usersCustomerUserID
		}
	}

	var eventID, userID string
	if document.Action == model.HubspotDocumentActionCreated {

		if C.AllowIdentificationOverwriteUsingSource(project.ID) {
			for i := range emails {
				for _, existingEmail := range userByCustomerUserID {
					if existingEmail == emails[i] {
						customerUserID = existingEmail
						break
					}
				}

				if i == len(emails)-1 {
					customerUserID = emails[0]
				}
			}

		}

		createdUserID, status := store.GetStore().CreateUser(&model.User{
			ProjectId:      project.ID,
			JoinTimestamp:  getEventTimestamp(document.Timestamp),
			CustomerUserId: customerUserID,
			Source:         model.GetRequestSourcePointer(model.UserSourceHubspot)})
		if status != http.StatusCreated {
			logCtx.WithField("status", status).Error("Failed to create user for hubspot contact created event.")
			return http.StatusInternalServerError
		}

		trackPayload.Name = U.EVENT_NAME_HUBSPOT_CONTACT_CREATED
		trackPayload.UserId = createdUserID

		if C.AllowIdentificationOverwriteUsingSource(project.ID) {
			// add this user for re-identification, in case new user was created with secondary email
			userByCustomerUserID[userID] = customerUserID
		}

		status, response := SDK.Track(project.ID, trackPayload, true, SDK.SourceHubspot, model.HubspotDocumentTypeNameContact)
		if status != http.StatusOK && status != http.StatusFound && status != http.StatusNotModified {
			logCtx.WithFields(log.Fields{"status": status, "track_response": response}).Error("Failed to track hubspot contact created event.")
			return http.StatusInternalServerError
		}

		userID = createdUserID
		eventID = response.EventId
	} else if document.Action == model.HubspotDocumentActionUpdated {
		trackPayload.Name = U.EVENT_NAME_HUBSPOT_CONTACT_UPDATED
		createdDocuments, status := store.GetStore().GetHubspotContactCreatedSyncIDAndUserID(project.ID, document.ID)
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
			event, errCode := store.GetStore().GetEventById(project.ID, createdDocuments[0].SyncId, "")
			if errCode != http.StatusFound {
				logCtx.WithField("event_id", createdDocuments[0].SyncId).Error(
					"Failed to get contact created event for getting user id.")
				return http.StatusInternalServerError
			}

			errCode = store.GetStore().UpdateHubspotDocumentAsSynced(
				project.ID, document.ID, model.HubspotDocumentTypeContact, event.ID, createdDocuments[0].Timestamp, model.HubspotDocumentActionCreated, event.UserId, "")
			if errCode != http.StatusAccepted {
				logCtx.Error("Failed to update hubspot contact created document user id.")
			}

			userID = event.UserId
		}

		if !C.AllowIdentificationOverwriteUsingSource(project.ID) {
			if customerUserID != "" {
				status, _ := SDK.Identify(project.ID, &SDK.IdentifyPayload{
					UserId: userID, CustomerUserId: customerUserID, RequestSource: model.UserSourceHubspot}, false)
				if status != http.StatusOK {
					logCtx.WithField("customer_user_id", customerUserID).Error(
						"Failed to identify user on hubspot contact sync.")
					return http.StatusInternalServerError
				}
			} else {
				logCtx.Warning("Skipped user identification on hubspot contact sync. No customer_user_id on properties.")
			}
		}

		if C.AllowIdentificationOverwriteUsingSource(project.ID) {
			// add this user for re-identification, in case new user was created with secondary email
			userByCustomerUserID[userID] = customerUserID
		}

		trackPayload.UserId = userID
		status, response := SDK.Track(project.ID, trackPayload, true, SDK.SourceHubspot, model.HubspotDocumentTypeNameContact)
		if status != http.StatusOK && status != http.StatusFound && status != http.StatusNotModified {
			logCtx.WithFields(log.Fields{"status": status, "track_response": response}).Error("Failed to track hubspot contact updated event.")
			return http.StatusInternalServerError
		}
		eventID = response.EventId

	} else {
		logCtx.Error("Invalid action on hubspot contact sync.")
		return http.StatusInternalServerError
	}

	// re-identify all users from web and hubspot with primary email
	if C.AllowIdentificationOverwriteUsingSource(project.ID) {

		for userID := range userByCustomerUserID {
			user, status := store.GetStore().GetUserWithoutProperties(project.ID, userID)
			if status != http.StatusFound {
				logCtx.WithFields(log.Fields{"err_code": status, "user_id": userID}).Error("Failed to get user from hubspot re-identification.")
				continue
			}

			if user.Source != nil && *user.Source != model.UserSourceHubspot && *user.Source != model.UserSourceWeb {
				continue
			}

			if user.CustomerUserId == primaryEmail {
				continue
			}

			status, _ = SDK.Identify(project.ID, &SDK.IdentifyPayload{
				UserId: userID, CustomerUserId: primaryEmail, RequestSource: model.UserSourceHubspot}, true)
			if status != http.StatusOK {
				logCtx.WithFields(log.Fields{"primary_email": primaryEmail, "user_id": userID}).Error(
					"Failed to identify user with primary email.")
			}
		}
	}

	var defaultSmartEventTimestamp int64
	if timestamp, err := model.GetHubspotDocumentUpdatedTimestamp(document); err != nil {
		logCtx.WithError(err).Warn("Failed to get last modified timestamp for smart event. Using document timestamp")
		defaultSmartEventTimestamp = document.Timestamp
	} else {
		defaultSmartEventTimestamp = timestamp
	}

	user, status := store.GetStore().GetUser(project.ID, userID)
	if status != http.StatusFound {
		logCtx.WithField("error_code", status).Error("Failed to get user on sync contact.")
		return http.StatusInternalServerError
	}

	existingCustomerUserID := user.CustomerUserId

	if existingCustomerUserID != customerUserID {
		logCtx.WithFields(log.Fields{"existing_customer_user_id": existingCustomerUserID, "new_customer_user_id": customerUserID}).
			Warn("Different customer user id seen on sync contact")
	}

	if document.Action == model.HubspotDocumentActionUpdated {
		err = ApplyHSOfflineTouchPointRule(project, trackPayload, document, defaultSmartEventTimestamp)
		if err != nil {
			// log and continue
			logCtx.WithField("EventID", eventID).WithField("userID", eventID).WithField("userID", eventID).Info("failed creating hubspot offline touch point")
		}
	}

	var prevProperties *map[string]interface{}
	for i := range hubspotSmartEventNames {
		prevProperties = TrackHubspotSmartEvent(project.ID, &hubspotSmartEventNames[i], eventID, document.ID, userID, document.Type,
			properties, prevProperties, defaultSmartEventTimestamp, false)
	}

	if C.EnableHubspotFormsEventsByProjectID(project.ID) {
		syncContactFormSubmissions(project, userID, document)
	}

	errCode := store.GetStore().UpdateHubspotDocumentAsSynced(
		project.ID, document.ID, model.HubspotDocumentTypeContact, eventID, document.Timestamp, document.Action, userID, "")
	if errCode != http.StatusAccepted {
		logCtx.Error("Failed to update hubspot contact document as synced.")
		return http.StatusInternalServerError
	}

	return http.StatusOK
}

func ApplyHSOfflineTouchPointRule(project *model.Project, trackPayload *SDK.TrackPayload, document *model.HubspotDocument, lastModifiedTimeStamp int64) error {

	logCtx := log.WithFields(log.Fields{"project_id": project.ID, "method": "ApplyHSOfflineTouchPointRule",
		"document_id": document.ID, "document_action": document.Action, "document": document})

	lastModifiedTimeStamp = U.CheckAndGetStandardTimestamp(lastModifiedTimeStamp)

	if &project.HubspotTouchPoints != nil && !U.IsEmptyPostgresJsonb(&project.HubspotTouchPoints) {

		var touchPointRules map[string][]model.HSTouchPointRule
		err := U.DecodePostgresJsonbToStructType(&project.HubspotTouchPoints, &touchPointRules)
		if err != nil {
			// logging and continuing.
			logCtx.WithField("Document", trackPayload).WithError(err).Error("Failed to fetch " +
				"offline touch point rules for hubspot document.")
			return err
		}

		// Get the last sync doc for the current update doc.
		prevDoc, status := store.GetStore().GetLastSyncedHubspotUpdateDocumentByID(document.ProjectId, document.ID, document.Type)
		if status != http.StatusFound {
			// In case no prev properties
			prevDoc = nil
		}

		rules := touchPointRules["hs_touch_point_rules"]
		for _, rule := range rules {

			// Check if rule is applicable & the record has changed property w.r.t filters
			if !canCreateHSTouchPoint(document.Action) || !filterCheck(rule, trackPayload, document, prevDoc, logCtx) {
				continue
			}

			_, err = CreateTouchPointEvent(project, trackPayload, document, rule, lastModifiedTimeStamp)
			if err != nil {
				logCtx.WithError(err).Error("failed to create touch point for hubspot contact updated document.")
				continue
			}

		}
	}
	return nil
}

func ApplyHSOfflineTouchPointRuleForEngagement(project *model.Project, trackPayload *SDK.TrackPayload,
	document *model.HubspotDocument, engagementType string) error {

	logCtx := log.WithFields(log.Fields{"project_id": project.ID, "method": "ApplyHSOfflineTouchPointRuleForEngagement",
		"document_id": document.ID, "document_action": document.Action, "document": document})

	if &project.HubspotTouchPoints != nil && !U.IsEmptyPostgresJsonb(&project.HubspotTouchPoints) {

		var touchPointRules map[string][]model.HSTouchPointRule
		err := U.DecodePostgresJsonbToStructType(&project.HubspotTouchPoints, &touchPointRules)
		if err != nil {
			// logging and continuing.
			logCtx.WithField("Document", trackPayload).WithError(err).Error("Failed to fetch " +
				"offline touch point rules for hubspot document.")
			return err
		}

		rules := touchPointRules["hs_touch_point_rules"]
		for _, rule := range rules {

			// Check if rule is applicable & the record has changed property w.r.t filters
			if !canCreateHSEngagementTouchPoint(engagementType, rule.RuleType) || !filterCheckEngagement(rule, trackPayload, logCtx) {
				continue
			}

			_, err = CreateTouchPointEventForEngagement(project, trackPayload, document, rule, engagementType)
			if err != nil {
				logCtx.WithError(err).Error("failed to create touch point for hubspot contact updated document.")
				continue
			}

		}
	}
	return nil
}

// CreateTouchPointEvent - Creates offline touchpoint for HS create/update events with given rule
func CreateTouchPointEvent(project *model.Project, trackPayload *SDK.TrackPayload, document *model.HubspotDocument,
	rule model.HSTouchPointRule, lastModifiedTimeStamp int64) (*SDK.TrackResponse, error) {

	logCtx := log.WithFields(log.Fields{"project_id": project.ID, "method": "CreateTouchPointEvent",
		"document_id": document.ID, "document_action": document.Action})
	logCtx.WithField("document", document).WithField("trackPayload", trackPayload).
		Info("CreateTouchPointEvent: creating hubspot offline touch point document")
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
	if rule.TouchPointTimeRef == model.LastModifiedTimeRef {
		timestamp = lastModifiedTimeStamp
	} else {
		timeValue, exists := (trackPayload.EventProperties)[rule.TouchPointTimeRef]
		if !exists {
			logCtx.Error("couldn't get the timestamp on hubspot contact properties using "+
				"given rule.TouchPointTimeRef-", rule.TouchPointTimeRef)
			return nil, errors.New(fmt.Sprintf("couldn't get the timestamp on hubspot "+
				"contact properties using given rule.TouchPointTimeRef - %s", rule.TouchPointTimeRef))
		}
		val, ok := timeValue.(int64)
		if !ok {
			logCtx.Error("couldn't convert the timestamp on hubspot contact properties. "+
				"using lastModifiedTimeStamp instead, val", rule.TouchPointTimeRef, timeValue)
			timestamp = lastModifiedTimeStamp
		} else {
			timestamp = val
		}
	}

	payload.Timestamp = timestamp

	// Mapping touch point properties:
	for key, value := range rule.PropertiesMap {

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

	status, trackResponse := SDK.Track(project.ID, payload, true, sdk.SourceHubspot, "")
	if status != http.StatusOK && status != http.StatusFound && status != http.StatusNotModified {
		logCtx.WithField("Document", trackPayload).WithError(err).Error(fmt.Errorf("create "+
			"hubspot touchpoint event track failed for doc type %d, message %s", document.Type, trackResponse.Error))
		return trackResponse, errors.New(fmt.Sprintf("create hubspot touchpoint event "+
			"track failed for doc type %d, message %s", document.Type, trackResponse.Error))
	}
	logCtx.WithField("statusCode", status).WithField("trackResponsePayload", trackResponse).Info("Successfully: created hubspot offline touch point")
	return trackResponse, nil
}

// CreateTouchPointEventForEngagement - Creates offline touchpoint for HS engagements (calls, meetings, forms, emails) with give rule
func CreateTouchPointEventForEngagement(project *model.Project, trackPayload *SDK.TrackPayload, document *model.HubspotDocument,
	rule model.HSTouchPointRule, engagementType string) (*SDK.TrackResponse, error) {

	logCtx := log.WithFields(log.Fields{"project_id": project.ID, "method": "CreateTouchPointEvent",
		"document_id": document.ID, "document_action": document.Action})
	logCtx.WithField("document", document).WithField("trackPayload", trackPayload).
		Info("CreateTouchPointEvent: creating hubspot offline touch point document")
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

	switch engagementType {
	case EngagementTypeEmail:
	case EngagementTypeIncomingEmail:
		{
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
		}

	default:
		logCtx.Error("engagement type not supported yet for rule creation")

	}

	payload.Timestamp = timestamp

	// Mapping touch point properties:
	for key, value := range rule.PropertiesMap {

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

	status, trackResponse := SDK.Track(project.ID, payload, true, sdk.SourceHubspot, "")
	if status != http.StatusOK && status != http.StatusFound && status != http.StatusNotModified {
		logCtx.WithField("Document", trackPayload).WithError(err).Error(fmt.Errorf("create "+
			"hubspot engagement touchpoint event track failed for doc type %d, message %s", document.Type, trackResponse.Error))
		return trackResponse, errors.New(fmt.Sprintf("create hubspot engagement touchpoint event "+
			"track failed for doc type %d, message %s", document.Type, trackResponse.Error))
	}
	logCtx.WithField("statusCode", status).WithField("trackResponsePayload", trackResponse).Info("Successfully: created hubspot engagement offline touch point")
	return trackResponse, nil
}

func canCreateHSEngagementTouchPoint(engagementType string, ruleType string) bool {

	switch engagementType {
	case EngagementTypeEmail:
	case EngagementTypeIncomingEmail:
		if ruleType == model.TouchPointRuleTypeEmails {
			return true
		}
	default:
		return false
	}
	return false
}

func canCreateHSTouchPoint(documentActionType int) bool {
	// Ignore doc types other than HubspotDocumentActionUpdated
	if documentActionType != model.HubspotDocumentActionUpdated {
		return false
	}
	return true
}

func filterCheck(rule model.HSTouchPointRule, trackPayload *SDK.TrackPayload, document *model.HubspotDocument, prevDoc *model.HubspotDocument, logCtx *log.Entry) bool {

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
				if filter.Property != "" {
					val, ok := trackPayload.EventProperties[filter.Property].(string)
					if ok && strings.Contains(val, filter.Value) {
						filtersPassed++
					}
				}
			}
		default:
			logCtx.WithField("Rule", rule).WithField("TrackPayload", trackPayload).
				Error("No matching operator found for offline touch point rules for hubspot document.")
			continue
		}
	}

	// Once filters passed, now check for the existing properties
	if filtersPassed != 0 && filtersPassed == len(rule.Filters) {
		if prevDoc == nil {
			// In case no prev properties exist continue creating OTP
			return true
		}

		if prevDoc.Action == model.HubspotDocumentActionCreated {
			// In case the only last sync doc was a CreateDocument, create an OTP for this one.
			return true
		}

		var err error
		var prevProperties *map[string]interface{}

		if document.Type == model.HubspotDocumentTypeContact {
			prevProperties, _, _, _, err = GetContactProperties(document.ProjectId, prevDoc)
		}

		if err != nil {
			logCtx.WithField("Rule", rule).WithField("TrackPayload", trackPayload).WithError(err).
				Error("Failed to GetHubspotDocumentProperties - Offline touch point. Continuing.")
			// In case of err with previous properties, log error but continue creating OTP
			return true
		}

		samePropertyMatchingScore := 0
		for _, filter := range rule.Filters {
			if val1, exists1 := trackPayload.EventProperties[filter.Property]; exists1 {
				if val2, exists2 := (*prevProperties)[filter.Property]; exists2 {
					if val1 == val2 {
						samePropertyMatchingScore++
					}
				}
			}
		}
		// If all filter properties matches with that of the previous found properties, skip and fail
		if samePropertyMatchingScore == len(rule.Filters) {
			return false
		} else {
			return true
		}
	}
	// When neither filters matched nor (filters matched but values are same)
	return false
}

func filterCheckEngagement(rule model.HSTouchPointRule, trackPayload *SDK.TrackPayload, logCtx *log.Entry) bool {

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
				if filter.Property != "" {
					val, ok := trackPayload.EventProperties[filter.Property].(string)
					if ok && strings.Contains(val, filter.Value) {
						filtersPassed++
					}
				}
			}
		default:
			logCtx.WithField("Rule", rule).WithField("TrackPayload", trackPayload).
				Error("No matching operator found for offline touch point rules for hubspot engagement document.")
			continue
		}
	}

	// return true if all the filters passed
	if filtersPassed != 0 && filtersPassed == len(rule.Filters) {
		return true
	}
	// When neither filters matched nor (filters matched but values are same)
	return false
}

func getDealUserID(projectID int64, deal *Deal) string {
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

	if C.DisableHubspotNonMarketingContactsByProjectID(projectID) && len(contactDocuments) == 0 {
		logCtx.Warning("No marketing contacts found for hubspot deal.")
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
func GetHubspotSmartEventNames(projectID int64) *map[string][]HubspotSmartEventName {

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
			if err == model.ErrorSmartEventFiterEmptyString {
				logCtx.WithError(err).Warn("Empty string on smart event filter.")
			} else {
				logCtx.WithError(err).Error("Failed to decode smart event filter expression")
			}
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

func getCompanyNameAndDomainName(document *model.HubspotDocument) (string, string, error) {
	if document.Type != model.HubspotDocumentTypeCompany {
		return "", "", errors.New("invalid document type")
	}
	var company Company
	err := json.Unmarshal(document.Value.RawMessage, &company)
	if err != nil {
		return "", "", err
	}

	companyName := company.Properties["name"].Value
	domainName := company.Properties["domain"].Value

	return companyName, domainName, nil
}

func getCompanyGroupID(document *model.HubspotDocument, companyName, domainName string) string {
	if document.ID != "" {
		return document.ID
	}
	if companyName != "" {
		return companyName
	}
	return domainName
}

func getCompanyProperties(projectID int64, document *model.HubspotDocument) (map[string]interface{}, error) {
	if projectID < 1 || document == nil {
		return nil, errors.New("invalid parameters")
	}

	if document.Type != model.HubspotDocumentTypeCompany {
		return nil, errors.New("invalid document type")
	}

	var company Company
	err := json.Unmarshal((document.Value).RawMessage, &company)
	if err != nil {
		return nil, err
	}

	// build user properties from properties.
	// make sure company name exist.
	userProperties := make(map[string]interface{}, 0)
	for key, value := range company.Properties {
		// add company name to user default property.
		if key == "name" {
			userProperties[U.UP_COMPANY] = value.Value
		}

		propertyKey := model.GetCRMEnrichPropertyKeyByType(model.SmartCRMEventSourceHubspot,
			model.HubspotDocumentTypeNameCompany, key)
		value, err := getHubspotMappedDataTypeValue(projectID, U.EVENT_NAME_HUBSPOT_CONTACT_UPDATED, propertyKey,
			value.Value, model.HubspotDocumentTypeCompany, document.GetDateProperties(), string(document.GetTimeZone()))
		if err != nil {
			log.WithFields(log.Fields{"project_id": projectID, "property_key": propertyKey}).WithError(err).Error("Failed to get property value.")
			continue
		}

		userProperties[propertyKey] = value
	}

	return userProperties, nil
}

func syncCompany(projectID int64, document *model.HubspotDocument) int {
	if document.Type != model.HubspotDocumentTypeCompany {
		return http.StatusInternalServerError
	}

	logCtx := log.WithFields(log.Fields{"project_id": projectID, "document_id": document.ID,
		"doc_timestamp": document.Timestamp})

	var company Company
	err := json.Unmarshal((document.Value).RawMessage, &company)
	if err != nil {
		logCtx.WithError(err).Error("Falied to unmarshal hubspot company document.")
		return http.StatusInternalServerError
	}

	contactIds := make([]string, 0, 0)
	for i := range company.ContactIds {
		contactIds = append(contactIds,
			strconv.FormatInt(company.ContactIds[i], 10))
	}

	userProperties, err := getCompanyProperties(projectID, document)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get company properties")
		return http.StatusInternalServerError
	}

	var companyUserID string
	var companyGroupID string
	if C.IsAllowedHubspotGroupsByProjectID(projectID) && document.GroupUserId == "" {
		companyUserID, companyGroupID, err = syncGroupCompany(projectID, document, &userProperties)
		if err != nil {
			logCtx.WithError(err).Error("Failed to update company group properties")
		}
	}

	if len(company.ContactIds) == 0 {
		logCtx.Warning("Skipped company sync. No contacts associated to company.")
		// No sync_id as no event or user or one user property created.
		errCode := store.GetStore().UpdateHubspotDocumentAsSynced(projectID, document.ID, model.HubspotDocumentTypeCompany, "", document.Timestamp, document.Action, "", companyUserID)
		if errCode != http.StatusAccepted {
			logCtx.Error("Failed to update hubspot deal document as synced.")
			return http.StatusInternalServerError
		}
		return http.StatusOK
	}

	var contactDocuments []model.HubspotDocument
	var errCode int
	if len(contactIds) > 0 {
		contactDocuments, errCode = store.GetStore().GetHubspotDocumentByTypeAndActions(projectID,
			contactIds, model.HubspotDocumentTypeContact, []int{model.HubspotDocumentActionCreated})
		if errCode == http.StatusInternalServerError {
			logCtx.Error("Failed to get hubspot documents by type and action on sync company.")
			return errCode
		}
	}

	if C.DisableHubspotNonMarketingContactsByProjectID(projectID) && len(contactDocuments) == 0 {
		logCtx.Warning("No marketing contacts found for hubspot company.")
	}

	// update $hubspot_company_name and other company
	// properties on each associated contact user.
	isContactsUpdateFailed := false
	contactUpdateCount := 0
	for _, contactDocument := range contactDocuments {
		if contactDocument.SyncId != "" {
			contactSyncEvent, errCode := store.GetStore().GetEventById(
				projectID, contactDocument.SyncId, "")
			if errCode == http.StatusFound {

				contactUser, status := store.GetStore().GetUser(projectID, contactSyncEvent.UserId)
				if status != http.StatusFound {
					logCtx.WithField("user_id", contactSyncEvent.UserId).Error(
						"Failed to get user by contact event user update user properties with company properties.")
					isContactsUpdateFailed = true
					continue
				}

				if C.IsAllowedHubspotGroupsByProjectID(projectID) {
					logCtx.Info("Updating user company group user id.")
					_, status = store.GetStore().UpdateUserGroup(projectID, contactUser.ID, model.GROUP_NAME_HUBSPOT_COMPANY, companyGroupID, companyUserID)
					if status != http.StatusAccepted && status != http.StatusNotModified {
						logCtx.Error("Failed to update user group id.")
					}
				}

				if contactUpdateCount > 100 {
					continue
				}
			}
		}
	}

	if isContactsUpdateFailed {
		logCtx.Error("Failed to update some hubspot company properties on user properties.")
		return http.StatusInternalServerError
	}

	// No sync_id as no event or user or one user property created.
	errCode = store.GetStore().UpdateHubspotDocumentAsSynced(projectID, document.ID, model.HubspotDocumentTypeCompany, "", document.Timestamp, document.Action, "", companyUserID)
	if errCode != http.StatusAccepted {
		logCtx.Error("Failed to update hubspot deal document as synced.")
		return http.StatusInternalServerError
	}

	return http.StatusOK
}

func getHubspotDateTimestampAsMidnightTimeZoneTimestamp(dateUTCMS interface{}, timeZone string) (int64, error) {
	timestamp, err := model.ReadHubspotTimestamp(dateUTCMS)
	if err != nil {
		return 0, err
	}

	loc, err := time.LoadLocation(timeZone)
	if err != nil {
		return 0, err
	}

	t := time.Unix(getEventTimestamp(timestamp), 0).UTC()
	timeInLoc := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, loc)
	return timeInLoc.Unix(), nil
}

func getHubspotMappedDataTypeValue(projectID int64, eventName, enKey string, value interface{}, typ int, dateProperties *map[string]bool, timeZone string) (interface{}, error) {
	if value == nil || value == "" {
		return "", nil
	}

	if !C.IsEnabledPropertyDetailFromDB() || !C.IsEnabledPropertyDetailByProjectID(projectID) {
		return value, nil
	}

	if dateProperties != nil {
		for key := range *dateProperties {
			typeAlias := model.GetHubspotTypeAliasByType(typ)
			enDateKey := model.GetCRMEnrichPropertyKeyByType(model.SmartCRMEventSourceHubspot,
				typeAlias, key)

			if enDateKey != enKey {
				continue
			}
			return getHubspotDateTimestampAsMidnightTimeZoneTimestamp(value, timeZone)
		}
	}
	ptype := store.GetStore().GetPropertyTypeByKeyValue(projectID, eventName, enKey, value, false)

	if ptype == U.PropertyTypeDateTime {
		datetime, err := U.GetPropertyValueAsFloat64(value)
		if err != nil {
			formatedTime, err := time.Parse(model.HubspotDateTimeLayout, U.GetPropertyValueAsString(value))
			if err == nil {
				return formatedTime.Unix(), nil
			}

			log.WithError(err).Error("Failed convert datetime property.")
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

func getDealProperties(projectID int64, document *model.HubspotDocument) (*map[string]interface{}, *map[string]interface{}, error) {

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
		value, err := getHubspotMappedDataTypeValue(projectID, U.EVENT_NAME_HUBSPOT_DEAL_STATE_CHANGED, enKey,
			v.Value, model.HubspotDocumentTypeDeal, document.GetDateProperties(), string(document.GetTimeZone()))
		if err != nil {
			log.WithFields(log.Fields{"project_id": projectID, "property_key": enKey, "value": value}).
				WithError(err).Error("Failed to get property value.")
			continue
		}

		enProperties[enKey] = value
		properties[k] = value

	}

	return &enProperties, &properties, nil
}

func isValidGroupName(documentType int, groupName string) bool {
	if documentType == model.HubspotDocumentTypeCompany && groupName == model.GROUP_NAME_HUBSPOT_COMPANY {
		return true
	}

	if documentType == model.HubspotDocumentTypeDeal && groupName == model.GROUP_NAME_HUBSPOT_DEAL {
		return true
	}

	return false
}

func getGroupEventName(docType int) (string, string) {
	if docType == model.HubspotDocumentTypeCompany {
		return util.GROUP_EVENT_NAME_HUBSPOT_COMPANY_CREATED, util.GROUP_EVENT_NAME_HUBSPOT_COMPANY_UPDATED
	}

	if docType == model.HubspotDocumentTypeDeal {
		return util.GROUP_EVENT_NAME_HUBSPOT_DEAL_CREATED, util.GROUP_EVENT_NAME_HUBSPOT_DEAL_UPDATED
	}

	return "", ""
}

func updateCreatedDocument(createdDocument *model.HubspotDocument) bool {
	if createdDocument.Type == model.HubspotDocumentTypeCompany {
		if createdDocument.GroupUserId == "" && createdDocument.UserId == "" {
			return true
		}
	}

	if createdDocument.Type == model.HubspotDocumentTypeDeal {
		if createdDocument.GroupUserId == "" {
			return true
		}
	}

	return false
}

func getGroupUserID(createdDocument *model.HubspotDocument) string {
	if createdDocument.GroupUserId != "" {
		return createdDocument.GroupUserId
	}

	return ""
}

func createOrUpdateHubspotGroupsProperties(projectID int64, document *model.HubspotDocument,
	enProperties *map[string]interface{}, groupName, groupID string) (string, string, int) {
	logCtx := log.WithFields(log.Fields{"project_id": projectID, "doc_type": document.Type, "document": document,
		"group_name": groupName, "group_id": groupID})

	if projectID == 0 || document == nil || enProperties == nil {
		logCtx.Error("Invalid parameters")
		return "", "", http.StatusBadRequest
	}

	if document.GroupUserId != "" {
		logCtx.Error("Document already processed for groups. Using existing group user id.")
		return document.GroupUserId, "", http.StatusOK
	}

	if !isValidGroupName(document.Type, groupName) {
		logCtx.Error("Invalid group name")
		return "", "", http.StatusBadRequest
	}

	groupUserID := ""
	var processEventNames []string
	var processEventTimestamps []int64
	var err error

	createdEventName, updatedEventName := getGroupEventName(document.Type)
	if document.Action == model.HubspotDocumentActionCreated {
		groupUserID, err = store.GetStore().CreateOrUpdateGroupPropertiesBySource(projectID, groupName, groupID, "",
			enProperties, getEventTimestamp(document.Timestamp), getEventTimestamp(document.Timestamp), model.SmartCRMEventSourceHubspot)

		if err != nil {
			logCtx.WithError(err).Error("Failed to update hubspot created group properties.")
			return "", "", http.StatusInternalServerError
		}

		processEventNames = append(processEventNames, createdEventName)
		processEventTimestamps = append(processEventTimestamps, document.Timestamp)
	}

	updateCreatedRecord := false
	if document.Action == model.HubspotDocumentActionUpdated {
		createdDocument, status := store.GetStore().GetSyncedHubspotDocumentByFilter(projectID,
			document.ID, document.Type, model.HubspotDocumentActionCreated)
		if status != http.StatusFound {
			logCtx.Error("Failed to get hubspot company created document for groups.")
			return "", "", http.StatusInternalServerError
		}

		if updateCreatedDocument(createdDocument) {
			processEventNames = append(processEventNames, createdEventName)
			processEventTimestamps = append(processEventTimestamps, createdDocument.Timestamp)
			updateCreatedRecord = true
		}

		groupUser := getGroupUserID(createdDocument)
		groupUserID, err = store.GetStore().CreateOrUpdateGroupPropertiesBySource(projectID, groupName, groupID,
			groupUser, enProperties, getEventTimestamp(createdDocument.Timestamp), getEventTimestamp(document.Timestamp),
			model.SmartCRMEventSourceHubspot)
		if err != nil {
			logCtx.WithError(err).Error("Failed to update hubspot updated group properties.")
			return "", "", http.StatusInternalServerError
		}

		processEventNames = append(processEventNames, updatedEventName)
		processEventTimestamps = append(processEventTimestamps, document.Timestamp)

	}

	if document.Action == model.HubspotDocumentActionAssociationsUpdated {
		createdDocument, status := store.GetStore().GetSyncedHubspotDocumentByFilter(projectID,
			document.ID, document.Type, model.HubspotDocumentActionCreated)
		if status != http.StatusFound {
			logCtx.WithFields(log.Fields{"err_code": status}).Error("Failed to get hubspot company created document for deals association update.")
			return "", "", http.StatusInternalServerError
		}

		return createdDocument.GroupUserId, "", http.StatusOK
	}

	if groupUserID == "" {
		logCtx.Error("Invalid group user id state.")
		return "", "", http.StatusInternalServerError
	}

	var eventId string
	for i := range processEventNames {

		trackPayload := &SDK.TrackPayload{
			Name:          processEventNames[i],
			ProjectId:     projectID,
			Timestamp:     getEventTimestamp(processEventTimestamps[i]),
			UserId:        groupUserID,
			RequestSource: model.UserSourceHubspot,
		}
		docTypeAlias := model.GetHubspotTypeAliasByType(document.Type)

		status, response := SDK.Track(projectID, trackPayload, true, SDK.SourceHubspot, docTypeAlias)
		if status != http.StatusOK && status != http.StatusFound && status != http.StatusNotModified {
			logCtx.WithFields(log.Fields{"status": status, "track_response": response, "doc_type": docTypeAlias,
				"event_name": processEventNames[i], "event_timestamp": processEventTimestamps[i]}).
				Error("Failed to track hubspot group event.")
			return "", "", http.StatusInternalServerError
		}
		eventId = response.EventId

		if processEventNames[i] == createdEventName && updateCreatedRecord {
			errCode := store.GetStore().UpdateHubspotDocumentAsSynced(projectID, document.ID, document.Type, "",
				processEventTimestamps[i], model.HubspotDocumentActionCreated, "", groupUserID) // marking user_id as empty won't update the column
			if errCode != http.StatusAccepted {
				logCtx.Error("Failed to update group user_id in hubspot created document as synced.")
				return "", "", http.StatusInternalServerError
			}
		}
	}

	return groupUserID, eventId, http.StatusOK
}

func getDealAssociatedIDs(projectID int64, document *model.HubspotDocument) ([]string, []string, error) {
	if document.Type != model.HubspotDocumentTypeDeal {
		return nil, nil, errors.New("invalid document type")
	}

	var deal Deal
	err := json.Unmarshal((document.Value).RawMessage, &deal)
	if err != nil {
		return nil, nil, err
	}

	var contactIDs []string
	var companyIDs []string
	associatedContactIDs := deal.Associations.AssociatedContactIds
	for i := range associatedContactIDs {
		contactID := strconv.FormatInt(associatedContactIDs[i], 10)
		contactIDs = append(contactIDs, contactID)
	}

	associatedCompanyIDs := deal.Associations.AssociatedCompanyIds
	for i := range associatedCompanyIDs {
		companyID := strconv.FormatInt(associatedCompanyIDs[i], 10)
		companyIDs = append(companyIDs, companyID)
	}

	return contactIDs, companyIDs, nil
}

func syncGroupCompany(projectID int64, document *model.HubspotDocument, enProperties *map[string]interface{}) (string, string, error) {
	companyName, domainName, err := getCompanyNameAndDomainName(document)
	if err != nil {
		return "", "", err
	}

	companyGroupID := getCompanyGroupID(document, companyName, domainName)
	companyUserID, _, status := createOrUpdateHubspotGroupsProperties(projectID, document, enProperties, model.GROUP_NAME_HUBSPOT_COMPANY, companyGroupID)
	if status != http.StatusOK {
		return "", "", errors.New("failed to update company group properties")
	}

	return companyUserID, companyGroupID, nil
}

func syncGroupDeal(projectID int64, enProperties *map[string]interface{}, document *model.HubspotDocument) (string, string, int) {
	logCtx := log.WithFields(log.Fields{"project_id": projectID, "document": document.ID, "doc_type": document.Type})
	if document.Type != model.HubspotDocumentTypeDeal {
		logCtx.Error("Invalid document type on syncGroupDeal.")
		return "", "", http.StatusBadRequest
	}
	if document.GroupUserId != "" {
		logCtx.Error("Deal already processed for groups.")
		return document.GroupUserId, "", http.StatusOK
	}

	dealGroupUserID, eventId, status := createOrUpdateHubspotGroupsProperties(projectID, document, enProperties, model.GROUP_NAME_HUBSPOT_DEAL, document.ID)
	if status != http.StatusOK {
		logCtx.Error("Failed to update deal group properties.")
		return "", "", http.StatusInternalServerError
	}

	contactIDList, companyIDList, err := getDealAssociatedIDs(projectID, document)
	if err != nil {
		logCtx.WithFields(log.Fields{"contact_ids": contactIDList, "company_ids": companyIDList}).
			WithError(err).Error("Failed to getDealAssociatedIDs.")
		return dealGroupUserID, eventId, http.StatusOK
	}

	if len(contactIDList) > 0 {
		documents, status := store.GetStore().GetHubspotDocumentByTypeAndActions(projectID, contactIDList, model.HubspotDocumentTypeContact, []int{model.HubspotDocumentActionCreated})
		if status != http.StatusFound && status != http.StatusNotFound {
			logCtx.WithFields(log.Fields{"contact_ids": contactIDList, "err_code": status}).
				Error("Failed to get contact created documents for syncGroupDeal.")
		}

		if C.DisableHubspotNonMarketingContactsByProjectID(projectID) && len(documents) == 0 {
			logCtx.Warning("No marketing contacts found for hubspot deal group..")
		}

		for i := range documents {
			userID := documents[i].UserId
			if userID == "" {
				logCtx.WithField("contact_id", documents[i].ID).Error("No user id found on contact create document")
				continue
			}

			_, status := store.GetStore().UpdateUserGroup(projectID, userID, model.GROUP_NAME_HUBSPOT_DEAL, "", dealGroupUserID)
			if status != http.StatusAccepted && status != http.StatusNotModified {
				logCtx.WithFields(log.Fields{"contact_id": documents[i].ID, "deal_group_user_id": dealGroupUserID, "err_code": status}).
					Error("Failed to update contact user group for hubspot deal.")
			}

		}
	}

	if len(companyIDList) > 0 {
		documents, status := store.GetStore().GetHubspotDocumentByTypeAndActions(projectID, companyIDList,
			model.HubspotDocumentTypeCompany, []int{model.HubspotDocumentActionCreated})
		if status != http.StatusFound {
			logCtx.WithFields(log.Fields{"company_ids": companyIDList}).
				Error("Failed to get company created documents for syncGroupDeal.")
		}

		for i := range documents {
			groupUserID := getGroupUserID(&documents[i])
			if groupUserID == "" {
				userProperties, err := getCompanyProperties(projectID, &documents[i])
				if err != nil {
					logCtx.WithFields(log.Fields{"document": documents[i]}).Error("Failed to get company properties in sync deal groups.")
					continue
				}

				groupUserID, _, err = syncGroupCompany(projectID, &documents[i], &userProperties)
				if err != nil {
					logCtx.WithFields(log.Fields{"document": documents[i]}).WithError(err).Error("Missing group user id in company record in sync deal groups.")
					continue
				}

				// update group_user_id  details on created record
				errCode := store.GetStore().UpdateHubspotDocumentAsSynced(projectID, documents[i].ID, documents[i].Type, "",
					documents[i].Timestamp, model.HubspotDocumentActionCreated, "", groupUserID)
				if errCode != http.StatusAccepted {
					logCtx.Error("Failed to update group user_id in hubspot created document as synced in sync deal company.")
					continue
				}
			}

			_, status = store.GetStore().CreateGroupRelationship(projectID, model.GROUP_NAME_HUBSPOT_DEAL, dealGroupUserID,
				model.GROUP_NAME_HUBSPOT_COMPANY, groupUserID)
			if status != http.StatusCreated && status != http.StatusConflict {
				logCtx.WithFields(log.Fields{"company_id": documents[i].ID,
					"left_group_name":     model.GROUP_NAME_HUBSPOT_DEAL,
					"right_group_name":    model.GROUP_NAME_HUBSPOT_COMPANY,
					"left_group_user_id":  dealGroupUserID,
					"right_group_user_id": groupUserID}).
					Error("Failed to update hubspot deal group relationships.")
			}

			_, status = store.GetStore().CreateGroupRelationship(projectID, model.GROUP_NAME_HUBSPOT_COMPANY, groupUserID,
				model.GROUP_NAME_HUBSPOT_DEAL, dealGroupUserID)
			if status != http.StatusCreated && status != http.StatusConflict {
				logCtx.WithFields(log.Fields{"company_id": documents[i].ID,
					"right_group_name":    model.GROUP_NAME_HUBSPOT_DEAL,
					"left_group_name":     model.GROUP_NAME_HUBSPOT_COMPANY,
					"right_group_user_id": dealGroupUserID,
					"left_group_user_id":  groupUserID}).
					Error("Failed to update hubspot deal group relationships.")
			}
		}
	}

	return dealGroupUserID, eventId, http.StatusOK
}

func syncDeal(projectID int64, document *model.HubspotDocument, hubspotSmartEventNames []HubspotSmartEventName) int {
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

	var groupUserID string
	var status int
	var eventId string
	if C.IsAllowedHubspotGroupsByProjectID(projectID) {
		groupUserID, eventId, status = syncGroupDeal(projectID, enProperties, document)
		if status != http.StatusOK {
			logCtx.Error("Failed to syncGroupDeal.")
			return http.StatusInternalServerError
		}
	}

	var defaultSmartEventTimestamp int64
	if timestamp, err := model.GetHubspotDocumentUpdatedTimestamp(document); err != nil {
		logCtx.WithError(err).Warn("Failed to get last modified timestamp for smart event. Using document timestamp")
		defaultSmartEventTimestamp = document.Timestamp
	} else {
		defaultSmartEventTimestamp = timestamp
	}

	var prevProperties *map[string]interface{}
	for i := range hubspotSmartEventNames {
		prevProperties = TrackHubspotSmartEvent(projectID, &hubspotSmartEventNames[i], eventId, document.ID, groupUserID, document.Type,
			properties, prevProperties, defaultSmartEventTimestamp, false)
	}

	errCode := store.GetStore().UpdateHubspotDocumentAsSynced(projectID,
		document.ID, model.HubspotDocumentTypeDeal, eventId, document.Timestamp, document.Action, "", groupUserID)
	if errCode != http.StatusAccepted {
		logCtx.Error("Failed to update hubspot deal document as synced.")
		return http.StatusInternalServerError
	}

	return http.StatusOK
}

var keyArrEngagementMeeting = []string{"id", "timestamp", "type", "source", "active"}
var keyArrMetaMeeting = []string{"startTime", "endTime", "title", "meetingOutcome"}
var keyArrEngagementCall = []string{"id", "timestamp", "type", "source", "activityType"}
var keyArrMetaCall = []string{"durationMilliseconds", "disposition", "status", "title"}
var keyArrEngagementEmail = []string{"id", "createdAt", "lastUpdated", "type", "teamId", "ownerId", "active", "timestamp", "source"}
var keyArrMetaEmail = []string{"from", "to", "subject", "sentVia"}

const (
	EngagementTypeCall          = "CALL"
	EngagementTypeEmail         = "EMAIL"
	EngagementTypeIncomingEmail = "INCOMING_EMAIL"
	EngagementTypeMeeting       = "MEETING"

	HSEngagementTimestampProperty = "$hubspot_engagement_timestamp"
)

func extractionOfPropertiesWithOutEmailOrContact(engagement Engagements, engagementType string) map[string]interface{} {
	logCtx := log.WithField("engagement_type", engagementType).WithField("engagement", engagement)
	properties := make(map[string]interface{})
	var engagementArray []string
	var metaDataArray []string
	if engagementType == EngagementTypeMeeting {
		engagementArray = keyArrEngagementMeeting
		metaDataArray = keyArrMetaMeeting
	} else if engagementType == EngagementTypeCall {
		engagementArray = keyArrEngagementCall
		metaDataArray = keyArrMetaCall
	} else if engagementType == EngagementTypeIncomingEmail || engagementType == EngagementTypeEmail {
		engagementArray = keyArrEngagementEmail
		metaDataArray = keyArrMetaEmail
	}
	for _, key := range engagementArray {
		if key == "timestamp" {
			vfloat, _ := util.GetPropertyValueAsFloat64(engagement.Engagement[key])
			properties[key] = (int64)(vfloat / 1000)
		} else if key == "id" {
			properties[key] = util.GetPropertyValueAsString(engagement.Engagement[key])
		} else {
			properties[key] = engagement.Engagement[key]
		}
	}
	for _, key := range metaDataArray {
		if key == "startTime" || key == "endTime" {
			properties[key] = (int64)((engagement.Metadata[key]).(float64) / 1000)
		} else if key == "to" {
			toInterface := engagement.Metadata[key]
			interfaceArray, isConvert := toInterface.([]interface{})
			if !isConvert {
				logCtx.Error("cannot convert interface to interface array")
				continue
			}

			if len(interfaceArray) == 0 {
				logCtx.Error("length of interface array is zero")
				continue
			}

			toMap, isConvert := interfaceArray[0].(map[string]interface{})
			if !isConvert {
				logCtx.Error("cannot convert interface to map")
				continue
			}
			properties[key] = toMap["email"]
		} else if key == "from" {
			toInterface := engagement.Metadata[key]
			fromMap, isConvert := toInterface.(map[string]interface{})
			if !isConvert {
				logCtx.Error("cannot convert interface to fromMap")
				continue
			}
			properties[key] = fromMap["email"]
		} else {
			properties[key] = engagement.Metadata[key]
		}
	}
	return properties
}

func getEngagementContactIds(engagementTypeStr string, engagement Engagements) ([]string, int) {
	logCtx := log.WithField("engagement_type_str", engagementTypeStr).WithField("engagement", engagement)
	contactIds := make([]string, 0, 0)
	if engagementTypeStr == EngagementTypeCall || engagementTypeStr == EngagementTypeMeeting {
		contactIdArr := engagement.Associations["contactIds"]
		for i := range contactIdArr {
			contactId, err := U.GetPropertyValueAsFloat64(contactIdArr[i])
			if err != nil {
				logCtx.WithError(err).Error("cannot convert interface to float64")
				return contactIds, http.StatusInternalServerError
			}
			contactIds = append(contactIds,
				strconv.FormatInt((int64)(contactId), 10))
		}
	} else if engagementTypeStr == EngagementTypeIncomingEmail || engagementTypeStr == EngagementTypeEmail {
		var contactId float64
		var err error
		var contact interface{}
		if engagementTypeStr == EngagementTypeIncomingEmail {
			contactIdInterface := engagement.Metadata["from"]

			fromMap, isConvert := contactIdInterface.(map[string]interface{})
			if !isConvert {
				logCtx.Error("cannot convert interface to fromMap")
				return contactIds, http.StatusInternalServerError
			}
			contact = fromMap["contactId"]
			if contact == "" || contact == nil {
				logCtx.Error("No contact id for INCOMING_EMAIL engamement. Will be marked as synced")
				return nil, http.StatusOK
			}

		} else {
			contactIdInterface := engagement.Metadata["to"]
			interfaceArray, isConvert := contactIdInterface.([]interface{})
			if !isConvert {
				logCtx.Error("cannot convert interface to interface array")
				return contactIds, http.StatusInternalServerError
			}

			if len(interfaceArray) == 0 {
				return contactIds, http.StatusOK
			}

			toMap, isConvert := interfaceArray[0].(map[string]interface{})
			if !isConvert {
				logCtx.Error("cannot convert interface to map")
				return contactIds, http.StatusInternalServerError
			}

			if len(toMap) == 0 {
				return contactIds, http.StatusOK
			}
			contact = toMap["contactId"]
			if contact == "" || contact == nil {
				logCtx.Error("No contact id for EMAIL engamement. Will be marked as synced")
				return nil, http.StatusOK
			}
		}

		contactId, err = U.GetPropertyValueAsFloat64(contact)
		if err != nil {
			logCtx.WithError(err).Error("cannot convert interface to float64")
			return contactIds, http.StatusInternalServerError
		}
		contactIds = append(contactIds,
			strconv.FormatInt((int64)(contactId), 10))
	}

	return contactIds, http.StatusOK
}

func syncEngagements(project *model.Project, document *model.HubspotDocument) int {
	logCtx := log.WithField("project_id", project.ID).WithField("document_id", document.ID)
	if document.Type != model.HubspotDocumentTypeEngagement {
		logCtx.Error("It is not a type of engagement")
		return http.StatusInternalServerError
	}

	var engagement Engagements
	err := json.Unmarshal((document.Value).RawMessage, &engagement)
	if err != nil {
		logCtx.WithError(err).Error("Failed to unmarshal hubspot engagement document.")
		return http.StatusInternalServerError
	}

	engagementType, isPresent := engagement.Engagement["type"]
	if !isPresent {
		logCtx.Error("Failed to find type as a key.")
		return http.StatusInternalServerError
	}

	engagementTypeStr := fmt.Sprintf("%v", engagementType)

	if engagementTypeStr != EngagementTypeCall && engagementTypeStr != EngagementTypeMeeting && engagementTypeStr != EngagementTypeIncomingEmail && engagementTypeStr != EngagementTypeEmail {
		logCtx.Error("Invalid engagement type")
		return http.StatusInternalServerError
	}

	if (engagementTypeStr == EngagementTypeIncomingEmail || engagementTypeStr == EngagementTypeEmail) && document.Action == model.HubspotDocumentActionUpdated {
		errCode := store.GetStore().UpdateHubspotDocumentAsSynced(project.ID, document.ID, model.HubspotDocumentTypeEngagement, "", document.Timestamp, document.Action, "", "")
		if errCode != http.StatusAccepted {
			logCtx.Error("Failed to update hubspot engagement document as synced.")
			return http.StatusInternalServerError
		}
		return http.StatusOK
	}

	contactIds, error := getEngagementContactIds(engagementTypeStr, engagement)
	if error != http.StatusOK {
		logCtx.Error("failed to get the contact id")
		return error
	}
	if len(contactIds) == 0 {
		errCode := store.GetStore().UpdateHubspotDocumentAsSynced(project.ID, document.ID, model.HubspotDocumentTypeEngagement, "", document.Timestamp, document.Action, "", "")
		if errCode != http.StatusAccepted {
			logCtx.Error("Failed to update hubspot engagement document as synced.")
			return http.StatusInternalServerError
		}
		return http.StatusOK
	}

	properties := extractionOfPropertiesWithOutEmailOrContact(engagement, engagementTypeStr)
	contactEngagementProperties := make(map[string]map[string]interface{})

	var contactDocuments []model.HubspotDocument
	var status int

	if len(contactIds) > 0 {
		contactDocuments, status = store.GetStore().GetHubspotDocumentByTypeAndActions(project.ID, contactIds, model.HubspotDocumentTypeContact, []int{model.HubspotDocumentActionCreated, model.HubspotDocumentActionUpdated})
		if status != http.StatusFound {
			if status != http.StatusNotFound {
				logCtx.Error(
					"Failed to get hubspot documents by type and action on sync engagement.")
				return http.StatusInternalServerError
			}
			logCtx.Warning("Missing engagement associated contact record.")
			// Avoid returning error if associated record is not present.
			return http.StatusOK
		}
	}

	if C.DisableHubspotNonMarketingContactsByProjectID(project.ID) && len(contactDocuments) == 0 {
		logCtx.Warning("No marketing contacts found for hubspot engagement.")
	}

	for i := range contactIds {
		var latestContactDocument *model.HubspotDocument
		for j := range contactDocuments {
			if contactIds[i] != contactDocuments[j].ID {
				continue
			}

			// pick the latest contact documet before the engagment timestamp or the first contact document.
			if latestContactDocument == nil || latestContactDocument.Timestamp < contactDocuments[j].Timestamp {
				latestContactDocument = &contactDocuments[j]
			}
		}

		if latestContactDocument == nil {
			logCtx.WithFields(log.Fields{"contact_id": contactIds[i]}).Warning("Missing contact record for activity.")
			continue
		}

		propertiesWithEmailOrContact := make(map[string]interface{})
		enProperties, _, _, _, err := GetContactProperties(project.ID, latestContactDocument)
		if err != nil {
			logCtx.WithError(err).Error("can't get contact properties")
			return http.StatusInternalServerError
		}
		key, value := getCustomerUserIDFromProperties(project.ID, *enProperties)
		enkey := model.GetCRMEnrichPropertyKeyByType(model.SmartCRMEventSourceHubspot, model.HubspotDocumentTypeNameEngagement, key)
		if _, exists := propertiesWithEmailOrContact[enkey]; !exists {
			propertiesWithEmailOrContact[enkey] = value
		}
		for key, value := range properties {
			enkey := model.GetCRMEnrichPropertyKeyByType(model.SmartCRMEventSourceHubspot, model.HubspotDocumentTypeNameEngagement, key)
			propertiesWithEmailOrContact[enkey] = value
		}

		contactEngagementProperties[contactIds[i]] = propertiesWithEmailOrContact
	}

	if len(contactEngagementProperties) < 1 {
		logCtx.Warn("No contacts for processing engagement.")
		return http.StatusInternalServerError
	}

	eventName := getEventNameByDocumentTypeAndAction(engagementTypeStr, document.Action)
	for i := range contactIds {
		var userId string
		for j := range contactDocuments {
			if contactIds[i] == contactDocuments[j].ID && contactDocuments[j].Action == 1 && contactDocuments[j].Synced {
				userId = contactDocuments[j].UserId
			}
		}

		if _, exist := contactEngagementProperties[contactIds[i]]; !exist || userId == "" {
			continue
		}

		payload := &SDK.TrackPayload{
			ProjectId:       project.ID,
			Name:            eventName,
			EventProperties: contactEngagementProperties[contactIds[i]],
			UserId:          userId,
			Timestamp:       getEventTimestamp(document.Timestamp),
		}
		status, _ = sdk.Track(project.ID, payload, true, SDK.SourceHubspot, "")
		if status != http.StatusOK && status != http.StatusFound && status != http.StatusNotModified {
			logCtx.Error("Failed to create hubspot engagement event")
			return http.StatusInternalServerError
		}

		err = ApplyHSOfflineTouchPointRuleForEngagement(project, payload, document, engagementTypeStr)
		if err != nil {
			// log and continue
			logCtx.WithField("TrackPayload", payload).WithField("userID", userId).Info("failed " +
				"creating engagement hubspot offline touch point")
		}

	}
	errCode := store.GetStore().UpdateHubspotDocumentAsSynced(project.ID, document.ID, model.HubspotDocumentTypeEngagement, "", document.Timestamp, document.Action, "", "")
	if errCode != http.StatusAccepted {
		logCtx.Error("Failed to update hubspot engagement document as synced.")
		return http.StatusInternalServerError
	}
	return http.StatusOK
}

func getEventNameByDocumentTypeAndAction(Type string, action int) string {
	if Type == EngagementTypeIncomingEmail || Type == EngagementTypeEmail {
		return U.EVENT_NAME_HUBSPOT_ENGAGEMENT_EMAIL
	}

	if model.HubspotDocumentActionCreated == action {
		if Type == EngagementTypeMeeting {
			return U.EVENT_NAME_HUBSPOT_ENGAGEMENT_MEETING_CREATED
		}
		return U.EVENT_NAME_HUBSPOT_ENGAGEMENT_CALL_CREATED
	}
	if Type == EngagementTypeMeeting {
		return U.EVENT_NAME_HUBSPOT_ENGAGEMENT_MEETING_UPDATED
	}
	return U.EVENT_NAME_HUBSPOT_ENGAGEMENT_CALL_UPDATED
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

func syncContactList(projectID int64, document *model.HubspotDocument, minTimestamp int64) int {
	if document.Type != model.HubspotDocumentTypeContactList {
		return http.StatusInternalServerError
	}

	pastEnrichmentEnabled := false
	if C.PastEventEnrichmentEnabled(projectID) {
		pastEnrichmentEnabled = true
	}

	logCtx := log.WithFields(log.Fields{"project_id": projectID, "document_id": document.ID,
		"doc_timestamp": document.Timestamp, "min_timestamp": minTimestamp})

	if !C.ContactListInsertEnabled(projectID) {
		logCtx.Warning("Skipped hubspot contact_list sync. contact_list sync not enabled.")
		return http.StatusOK
	}

	if document.Action == model.HubspotDocumentActionCreated {
		errCode := store.GetStore().UpdateHubspotDocumentAsSynced(
			projectID, document.ID, model.HubspotDocumentTypeContactList, "", document.Timestamp, document.Action, "", "")
		if errCode != http.StatusAccepted {
			logCtx.Error("Failed to update hubspot contact_list document as synced.")
			return http.StatusInternalServerError
		}

		return http.StatusOK
	}

	var newContactList NewContactList
	err := json.Unmarshal((document.Value).RawMessage, &newContactList)
	if err != nil {
		logCtx.WithError(err).Error("Failed to unmarshal new hubspot contact_list document.")
		return http.StatusInternalServerError
	}

	oldContactListDocument, status := store.GetStore().GetHubspotDocumentByTypeAndActions(projectID, []string{document.ID}, model.HubspotDocumentTypeContactList, []int{model.HubspotDocumentActionCreated, model.HubspotDocumentActionUpdated})
	if status == http.StatusBadRequest || status == http.StatusInternalServerError {
		logCtx.Error("Failed to get old synced hubspot contact_list document.")
		return http.StatusInternalServerError
	}

	var prevContactListDocument *model.HubspotDocument
	for i := range oldContactListDocument {
		if oldContactListDocument[i].Timestamp < document.Timestamp {
			prevContactListDocument = &oldContactListDocument[i]
		}
	}

	var errOldContactList error
	var errNewContactList error

	var prevOldContactList OldContactList
	var prevNewContactList NewContactList
	if prevContactListDocument != nil {
		errOldContactList := json.Unmarshal(((*prevContactListDocument).Value).RawMessage, &prevOldContactList)
		errNewContactList := json.Unmarshal(((*prevContactListDocument).Value).RawMessage, &prevNewContactList)

		if errOldContactList != nil && errNewContactList != nil {
			logCtx.Error("Failed to unmarshal old hubspot contact_list document to oldContactList and newContactList.")
			return http.StatusInternalServerError
		}

		if errOldContactList == nil && errNewContactList == nil {
			logCtx.Error("Unmarshaled old hubspot contact_list document to both oldContactList and newContactList. Not possible.")
			return http.StatusInternalServerError
		}
	}

	contactIds := make([]string, 0, 0)

	if errOldContactList == nil {
		for i := range newContactList.ContactIds {
			if !U.ContainsInt64InArray(prevOldContactList.ContactIds, newContactList.ContactIds[i]) {
				contactIds = append(contactIds, strconv.FormatInt(newContactList.ContactIds[i], 10))
			}
		}
	}

	if errNewContactList == nil {
		for i := range newContactList.ContactIds {
			if !U.ContainsInt64InArray(prevNewContactList.ContactIds, newContactList.ContactIds[i]) {
				contactIds = append(contactIds, strconv.FormatInt(newContactList.ContactIds[i], 10))
			}
		}
	}

	if len(contactIds) == 0 {
		logCtx.Warning("Skipped contact_list sync. No contacts in contact_list.")
		// No sync_id as no event or user or one user property created.
		errCode := store.GetStore().UpdateHubspotDocumentAsSynced(projectID, document.ID, model.HubspotDocumentTypeContactList, "", document.Timestamp, document.Action, "", "")
		if errCode != http.StatusAccepted {
			logCtx.Error("Failed to update hubspot contact_list document as synced.")
			return http.StatusInternalServerError
		}
		return http.StatusOK
	}

	var contactDocuments []model.HubspotDocument
	var errCode int

	contactDocuments, errCode = store.GetStore().GetHubspotDocumentByTypeAndActions(projectID,
		contactIds, model.HubspotDocumentTypeContact, []int{model.HubspotDocumentActionCreated})
	if errCode == http.StatusInternalServerError {
		logCtx.Error("Failed to get hubspot contact documents by type and action on sync contact_list.")
		return errCode
	}

	for _, doc := range contactDocuments {
		if !doc.Synced {
			logCtx.Error("Hubspot contact document not synced. Skipping processing for now.")
			return http.StatusOK
		}
	}

	contactIdDocumentMap := make(map[string]model.HubspotDocument)
	for _, doc := range contactDocuments {
		contactIdDocumentMap[doc.ID] = doc
	}

	for contactId, listMembership := range newContactList.ListMemberships {
		doc, ok := contactIdDocumentMap[contactId]
		if !ok {
			logCtx.Info("Hubspot contact document not present in contact list.")
			continue
		}

		if listMembership.StaticListId != newContactList.ListId {
			continue
		}

		isPast := false
		if pastEnrichmentEnabled {
			isPast = listMembership.Timestamp < minTimestamp
		}

		_, properties, _, _, _ := GetContactProperties(projectID, &doc)
		var emailId string
		if (*properties)["EMAIL"] != nil {
			emailId, err = U.GetValueAsString((*properties)["EMAIL"])
			if err != nil {
				logCtx.Error("Failed to get emailId from contact properties.")
				emailId = ""
			}
		} else {
			logCtx.Error("No emailId in contact properties.")
			emailId = ""
		}

		eventProperties := map[string]interface{}{
			model.GetCRMEnrichPropertyKeyByType(model.SmartCRMEventSourceHubspot, model.HubspotDocumentTypeNameContactList,
				"list_name"): newContactList.ListName,
			model.GetCRMEnrichPropertyKeyByType(model.SmartCRMEventSourceHubspot, model.HubspotDocumentTypeNameContactList,
				"list_id"): newContactList.ListId,
			model.GetCRMEnrichPropertyKeyByType(model.SmartCRMEventSourceHubspot, model.HubspotDocumentTypeNameContactList,
				"list_create_timestamp"): getEventTimestamp(newContactList.ListCreatedAt),
			model.GetCRMEnrichPropertyKeyByType(model.SmartCRMEventSourceHubspot, model.HubspotDocumentTypeNameContactList,
				"timestamp"): getEventTimestamp(listMembership.Timestamp),
			model.GetCRMEnrichPropertyKeyByType(model.SmartCRMEventSourceHubspot, model.HubspotDocumentTypeNameContactList,
				"type"): newContactList.ListType,
			model.GetCRMEnrichPropertyKeyByType(model.SmartCRMEventSourceHubspot, model.HubspotDocumentTypeNameContactList,
				"contact_id"): listMembership.Vid,
			model.GetCRMEnrichPropertyKeyByType(model.SmartCRMEventSourceHubspot, model.HubspotDocumentTypeNameContactList,
				"contact_email"): emailId,
		}

		request := &SDK.TrackPayload{
			ProjectId:       projectID,
			Timestamp:       getEventTimestamp(listMembership.Timestamp),
			EventProperties: eventProperties,
			RequestSource:   model.UserSourceHubspot,
			Name:            U.EVENT_NAME_HUBSPOT_CONTACT_LIST,
			UserId:          contactIdDocumentMap[contactId].UserId,
			IsPast:          isPast,
		}

		status, response := SDK.Track(projectID, request, true, SDK.SourceHubspot, model.HubspotDocumentTypeNameContactList)
		if status != http.StatusOK && status != http.StatusFound && status != http.StatusNotModified {
			logCtx.WithFields(log.Fields{"status": status, "track_response": response}).Error("Failed to track hubspot added to a list event.")
			return http.StatusInternalServerError
		}
	}

	errCode = store.GetStore().UpdateHubspotDocumentAsSynced(
		projectID, document.ID, model.HubspotDocumentTypeContactList, "", document.Timestamp, document.Action, "", "")
	if errCode != http.StatusAccepted {
		logCtx.Error("Failed to update hubspot contact_list document as synced.")
		return http.StatusInternalServerError
	}

	return http.StatusOK
}

func syncAll(project *model.Project, documents []model.HubspotDocument, hubspotSmartEventNames []HubspotSmartEventName, minTimestamp int64) int {
	logCtx := log.WithField("project_id", project.ID)
	var seenFailures bool
	for i := range documents {
		logCtx = logCtx.WithFields(log.Fields{"document_id": documents[i].ID, "doc_type": documents[i].Type, "document_timestamp": documents[i].Timestamp})
		startTime := time.Now().Unix()

		switch documents[i].Type {

		case model.HubspotDocumentTypeContact:
			errCode := syncContact(project, &documents[i], hubspotSmartEventNames)
			if errCode != http.StatusOK {
				seenFailures = true
			}
		case model.HubspotDocumentTypeCompany:
			errCode := syncCompany(project.ID, &documents[i])
			if errCode != http.StatusOK {
				seenFailures = true
			}
		case model.HubspotDocumentTypeDeal:
			errCode := syncDeal(project.ID, &documents[i], hubspotSmartEventNames)
			if errCode != http.StatusOK {
				seenFailures = true
			}
		case model.HubspotDocumentTypeEngagement:
			errCode := syncEngagements(project, &documents[i])
			if errCode != http.StatusOK {
				seenFailures = true
			}
		case model.HubspotDocumentTypeContactList:
			errCode := syncContactList(project.ID, &documents[i], minTimestamp)
			if errCode != http.StatusOK {
				seenFailures = true
			}
		}

		logCtx.WithField("time_taken_in_secs", time.Now().Unix()-startTime).Info(
			"Sync completed.")
	}

	if seenFailures {
		return http.StatusInternalServerError
	}

	return http.StatusOK
}

// Status definition
type Status struct {
	ProjectId              int64  `json:"project_id"`
	Type                   string `json:"type"`
	Status                 string `json:"status"`
	Count                  int    `json:"count"`
	TotalTime              string `json:"total_time`
	Message                string `json:"message,omiempty"`
	IsProcessLimitExceeded bool   `json:"process_limit_exceeded"`
}

type syncWorkerStatus struct {
	HasFailure bool
	Lock       sync.Mutex
}

// syncAllWorker is a wrapper over syncAll function for providing concurrency
func syncAllWorker(project *model.Project, wg *sync.WaitGroup, syncStatus *syncWorkerStatus, documents []model.HubspotDocument, hubspotSmartEventNames []HubspotSmartEventName, minTimestamp int64) {
	defer wg.Done()

	errCode := syncAll(project, documents, hubspotSmartEventNames, minTimestamp)

	syncStatus.Lock.Lock()
	defer syncStatus.Lock.Unlock()
	if errCode != http.StatusOK {
		syncStatus.HasFailure = true
	}
}

func syncByOrderedTimeSeries(project *model.Project, orderedTimeSeries [][]int64, workersPerProject int, recordsMaxCreatedAtSec int64, datePropertiesByObjectType map[int]*map[string]bool, timeZone U.TimeZoneString, recordsProcessLimit int,
	hubspotSmartEventNames *map[string][]HubspotSmartEventName) (map[string]bool, map[string]int64, map[string]int, bool) {
	logCtx := log.WithFields(log.Fields{"project_id": project.ID, "worker_per_project": workersPerProject,
		"record_max_created_at": recordsMaxCreatedAtSec, "record_process_limit": recordsProcessLimit})
	if project == nil || len(orderedTimeSeries) == 0 {
		logCtx.Error("Invalid parameters.")
		return nil, nil, nil, false
	}

	minTimestamps := make(map[int]int64)
	for i := range syncOrderByType {
		if syncOrderByType[i] == model.HubspotDocumentTypeContactList && !C.ContactListInsertEnabled(project.ID) {
			continue
		}

		// for contact-list set last 48 hrs as begenning for recent events
		if syncOrderByType[i] == model.HubspotDocumentTypeContactList {
			minTimestamps[syncOrderByType[i]] = U.TimeNowZ().Add(-48*time.Hour).Unix() * 1000
			continue
		}

		minTimestamp, err := store.GetStore().GetMinTimestampByFirstSync(project.ID, syncOrderByType[i])
		if err != http.StatusFound && err != http.StatusNotFound {
			logCtx.WithFields(log.Fields{"project_id": project.ID, "doc_type": syncOrderByType[i]}).Error("Failed to get timestamp by first sync in hubspot document.")
			return nil, nil, nil, false
		}

		minTimestamps[syncOrderByType[i]] = minTimestamp
	}

	processedCount := 0
	overAllSyncStatus := make(map[string]bool)
	overallExecutionTime := make(map[string]int64)
	overallProcessedCount := make(map[string]int)
	for _, timeRange := range orderedTimeSeries {

		for i := range syncOrderByType {
			if syncOrderByType[i] == model.HubspotDocumentTypeContactList && !C.ContactListInsertEnabled(project.ID) {
				continue
			}

			startTime := time.Now()
			logCtx = logCtx.WithFields(log.Fields{"type": syncOrderByType[i], "time_range": timeRange})

			logCtx.Info("Processing started for given time range")
			var documents []model.HubspotDocument
			var errCode int
			if workersPerProject > 1 {
				documents, errCode = store.GetStore().GetHubspotDocumentsByTypeANDRangeForSync(project.ID, syncOrderByType[i], timeRange[0], timeRange[1], recordsMaxCreatedAtSec)
			} else {
				documents, errCode = store.GetStore().
					GetHubspotDocumentsByTypeForSync(project.ID, syncOrderByType[i], recordsMaxCreatedAtSec)
			}

			if errCode != http.StatusFound {
				logCtx.WithFields(log.Fields{"time_range": timeRange, "doc_type": syncOrderByType[i]}).Error("Failed to get hubspot document by type for sync.")
				continue
			}

			fillDatePropertiesAndTimeZone(documents, datePropertiesByObjectType[syncOrderByType[i]], timeZone)
			docTypeAlias := model.GetHubspotTypeAliasByType(syncOrderByType[i])

			batches := GetBatchedOrderedDocumentsByID(documents, workersPerProject)

			var syncStatus syncWorkerStatus
			var workerIndex int
			isProcessLimitExceeded := false
			for bi := range batches {
				batch := batches[bi]
				var wg sync.WaitGroup
				for docID := range batch {
					processedCount += len(batch[docID])
					logCtx.WithFields(log.Fields{"worker": workerIndex, "doc_id": docID, "type": syncOrderByType[i]}).Info("Processing Batch by doc_id")
					workerIndex++
					wg.Add(1)
					go syncAllWorker(project, &wg, &syncStatus, batch[docID], (*hubspotSmartEventNames)[docTypeAlias], minTimestamps[syncOrderByType[i]])
				}
				wg.Wait()
				if processedCount > recordsProcessLimit {
					isProcessLimitExceeded = true
					break
				}
			}

			if _, exist := overAllSyncStatus[docTypeAlias]; !exist {
				overAllSyncStatus[docTypeAlias] = false
			}

			if syncStatus.HasFailure {
				overAllSyncStatus[docTypeAlias] = true
			}

			overallExecutionTime[docTypeAlias] += time.Since(startTime).Milliseconds()
			overallProcessedCount[docTypeAlias] += len(documents)
			if isProcessLimitExceeded {
				return overAllSyncStatus, overallExecutionTime, overallProcessedCount, true
			}

			logCtx.Info("Processing completed for given time range")
		}
	}

	return overAllSyncStatus, overallExecutionTime, overallProcessedCount, false
}

// Sync - Syncs hubspot documents in an order of type.
func Sync(projectID int64, workersPerProject int, recordsMaxCreatedAtSec int64, datePropertiesByObjectType map[int]*map[string]bool, timeZone U.TimeZoneString, recordsProcessLimit int) ([]Status, bool) {
	logCtx := log.WithFields(log.Fields{"project_id": projectID, "workers_per_project": workersPerProject, "record_max_created_at": recordsMaxCreatedAtSec})
	logCtx.Info("Running sync for project.")

	statusByProjectAndType := make([]Status, 0, 0)
	hubspotSmartEventNames := GetHubspotSmartEventNames(projectID)
	status := CreateOrGetHubspotEventName(projectID)
	if status != http.StatusOK {
		statusByProjectAndType = append(statusByProjectAndType, Status{ProjectId: projectID,
			Status: "Failed to create event names"})
		return statusByProjectAndType, true
	}

	var orderedTimeSeries [][]int64
	minTimestamp, errCode := store.GetStore().GetHubspotDocumentBeginingTimestampByDocumentTypeForSync(projectID, syncOrderByType[:])
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
		orderedTimeSeries = model.GetCRMTimeSeriesByStartTimestamp(projectID, minTimestamp, model.SmartCRMEventSourceHubspot)
	} else {
		// generate single time series
		orderedTimeSeries = append(orderedTimeSeries, []int64{minTimestamp, time.Now().Unix() * 1000})
	}

	// Get/Create SF touch point event name
	_, status = store.GetStore().CreateOrGetOfflineTouchPointEventName(projectID)
	if status != http.StatusFound && status != http.StatusConflict && status != http.StatusCreated {
		logCtx.Error("failed to create event name on SF for offline touch point")
		return statusByProjectAndType, true
	}

	project, errCode := store.GetStore().GetProject(projectID)
	if errCode != http.StatusFound {
		log.Error("Failed to get project")
		return statusByProjectAndType, true
	}

	anyFailure := false
	overAllSyncStatus, overallExecutionTime, overallProcessedCount, isProcessLimitExceeded := syncByOrderedTimeSeries(project, orderedTimeSeries, workersPerProject,
		recordsMaxCreatedAtSec, datePropertiesByObjectType, timeZone, recordsProcessLimit, hubspotSmartEventNames)

	for docTypeAlias, failure := range overAllSyncStatus {
		status := Status{ProjectId: projectID,
			Type:                   docTypeAlias,
			IsProcessLimitExceeded: isProcessLimitExceeded}
		if failure {
			status.Status = U.CRM_SYNC_STATUS_FAILURES
			anyFailure = true
		} else {
			status.Status = U.CRM_SYNC_STATUS_SUCCESS
		}
		status.Count = overallProcessedCount[docTypeAlias]
		status.TotalTime = time.Duration(overallExecutionTime[docTypeAlias] * int64(time.Millisecond)).String()
		statusByProjectAndType = append(statusByProjectAndType, status)
	}

	return statusByProjectAndType, anyFailure
}
