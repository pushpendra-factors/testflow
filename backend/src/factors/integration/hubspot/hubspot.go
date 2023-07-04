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

// Engagement definition
type Engagements struct {
	Engagement   map[string]interface{}   `json:"engagement"`
	Associations map[string][]interface{} `json:"associations"`
	Metadata     map[string]interface{}   `json:"metadata"`
}

// EngagementV3 definition
type EngagementsV3 struct {
	Id           string                   `json:"id"`
	Properties   map[string]interface{}   `json:"properties"`
	Associations map[string][]interface{} `json:"associations"`
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
	DealId     int64               `json:"dealId"`
	Properties map[string]Property `json:"properties"`
}

// DealV3 definition
type DealV3 struct {
	DealId     string                 `json:"id"`
	Properties map[string]interface{} `json:"properties"`
}

// Deal Association definition
type DealAssociations struct {
	Associations Associations `json:"associations"`
}

// Company definition
type Company struct {
	CompanyId int64 `json:"companyId"`
	// not part of hubspot response. added to company on download.
	ContactIds []int64             `json:"contactIds"`
	Properties map[string]Property `json:"properties"`
}

// CompanyV3 definition
type CompanyV3 struct {
	CompanyId string `json:"id"`
	// not part of hubspot response. added to company on download.
	ContactIds []int64           `json:"contactIds"`
	Properties map[string]string `json:"properties"`
}

// Option definition
type Option struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

// PropertyDetail definition for hubspot properties api
type PropertyDetail struct {
	Name                         string   `json:"name"`
	Label                        string   `json:"label"`
	Type                         string   `json:"type"`
	FieldType                    string   `json:"fieldType"`
	ExternalOptionsReferenceType string   `json:"externalOptionsReferenceType"`
	Options                      []Option `json:"options"`
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

func syncContactFormSubmissions(project *model.Project, otpRules *[]model.OTPRule, uniqueOTPEventKeys *[]string, userId string, document *model.HubspotDocument) {
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
	logCtx.WithFields(log.Fields{"ProjectID": project.ID}).Info("Inside method syncContactFormSubmissions")
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

		status, trackResponse := sdk.Track(project.ID, payload, true, SDK.SourceHubspot, "")
		if status != http.StatusOK && status != http.StatusFound && status != http.StatusNotModified {
			logCtx.Error("Failed to create hubspot form-submission event")
			return
		}
		logCtx.WithFields(log.Fields{"ProjectID": project.ID, "Payload": payload}).Info("Invoking method ApplyHSOfflineTouchPointRuleForForms")

		if !C.IsProjectIDSkippedForOtp(project.ID) {
			err = ApplyHSOfflineTouchPointRuleForForms(project, otpRules, uniqueOTPEventKeys, payload, document, eventTimestamp)
			if err != nil {
				// log and continue
				logCtx.WithField("EventID", trackResponse.EventId).WithField("userID", trackResponse.UserId).Info("failed creating hubspot offline touch point for form submission")
			}
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
			log.WithFields(log.Fields{"project_id": projectID, "property_key": enKey}).
				WithError(err).Error("Failed to get property value.")
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

func getCustomIdentification(projectID int64, document *model.HubspotDocument) (string, bool) {
	logCtx := log.WithFields(log.Fields{"project_id": projectID, "document_id": document.ID,
		"timestamp": document.Timestamp})

	field := model.GetHubspotCustomIdentificationFieldByProjectID(projectID)
	if field == "" {
		return "", false
	}

	_, properties, _, _, err := GetContactProperties(projectID, document)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get contact properties" +
			"on getCustomIdentificationProperty.")
		return "", true
	}

	return U.GetPropertyValueAsString((*properties)[field]), true
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

/*
Map defined as follows:
eventName -> property -> Property type
*/
var engagementDatetimePropertiesMap = map[string]map[string]string{
	U.EVENT_NAME_HUBSPOT_ENGAGEMENT_MEETING_CREATED: {
		"starttime": U.PropertyTypeDateTime,
		"endtime":   U.PropertyTypeDateTime,
		"timestamp": U.PropertyTypeDateTime,
	},
	U.EVENT_NAME_HUBSPOT_ENGAGEMENT_MEETING_UPDATED: {
		"starttime": U.PropertyTypeDateTime,
		"endtime":   U.PropertyTypeDateTime,
		"timestamp": U.PropertyTypeDateTime,
	},
	U.EVENT_NAME_HUBSPOT_ENGAGEMENT_EMAIL: {
		"createdat":   U.PropertyTypeDateTime,
		"lastupdated": U.PropertyTypeDateTime,
		"timestamp":   U.PropertyTypeDateTime,
	},
	U.EVENT_NAME_HUBSPOT_ENGAGEMENT_CALL_CREATED: {
		"timestamp": U.PropertyTypeDateTime,
	},
	U.EVENT_NAME_HUBSPOT_ENGAGEMENT_CALL_UPDATED: {
		"timestamp": U.PropertyTypeDateTime,
	},
}

func SyncEngagementDatetimeProperties(projectID int64) (bool, []Status) {
	logCtx := log.WithFields(log.Fields{"project_id": projectID})

	if projectID == 0 {
		logCtx.Error("Missing project_id.")
		return true, nil
	}

	var allStatus []Status
	anyFailures := false

	for eventName := range engagementDatetimePropertiesMap {
		var engagementMeetingStatus Status
		engagementMeetingStatus.ProjectId = projectID
		engagementMeetingStatus.Type = model.HubspotDocumentTypeNameEngagement

		var engagementMeetingFailure bool

		for fieldName, pType := range engagementDatetimePropertiesMap[eventName] {
			enKey := model.GetCRMEnrichPropertyKeyByType(
				model.SmartCRMEventSourceHubspot,
				model.HubspotDocumentTypeNameEngagement,
				fieldName,
			)

			err := store.GetStore().CreateOrDeletePropertyDetails(projectID, eventName, enKey, pType, false, true)
			if err != nil {
				logCtx.WithFields(log.Fields{"enriched_property_key": enKey}).WithError(err).Error("Failed to create event property details.")
				engagementMeetingFailure = true
			} else {
				engagementMeetingFailure = false
			}
		}

		if engagementMeetingFailure {
			engagementMeetingStatus.Status = U.CRM_SYNC_STATUS_FAILURES
			anyFailures = true
		} else {
			engagementMeetingStatus.Status = U.CRM_SYNC_STATUS_SUCCESS
		}

		allStatus = append(allStatus, engagementMeetingStatus)
	}

	return anyFailures, allStatus
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

	failure, engagementsStatus := SyncEngagementDatetimeProperties(projectID)
	if failure {
		anyFailures = true
	}

	allStatus = append(allStatus, engagementsStatus...)

	return anyFailures, allStatus
}

func GetHubspotPropertyDetailsByDataType(apiKey, refreshToken string, docTypes []string) (map[string][]PropertyDetail, bool) {
	propertyDetailsMap := make(map[string][]PropertyDetail)
	failures := false

	for i := range docTypes {
		logCtx := log.WithFields(log.Fields{"doc_type": docTypes[i]})

		apiObjectName := model.GetHubspotObjectTypeByDocumentType(docTypes[i])
		if apiObjectName == "" {
			logCtx.Error("Invalid doc_type for GetHubspotPropertyDetailsByDataType.")
			failures = true
			continue
		}

		propertiesMeta, err := GetHubspotPropertiesMeta(apiObjectName, apiKey, refreshToken)
		if err != nil {
			logCtx.WithField("api_object_name", apiObjectName).WithError(err).Error("Failed to get hubspot properties meta.")
			failures = true
			continue
		}

		propertyDetailsMap[docTypes[i]] = propertiesMeta
	}
	return propertyDetailsMap, failures
}

func SyncPropertiesOptions(projectID int64, propertiesMetaMap map[string][]PropertyDetail) bool {
	logCtx := log.WithFields(log.Fields{"project_id": projectID})

	failures := false
	for docType, propertiesMeta := range propertiesMetaMap {
		logCtx.WithFields(log.Fields{"doc_type": docType})

		for _, property := range propertiesMeta {
			for i := range property.Options {
				if property.Options[i].Value == "" || property.Options[i].Label == "" {
					continue
				}

				propertyKey := model.GetCRMEnrichPropertyKeyByType(model.SmartCRMEventSourceHubspot, docType, property.Name)
				status := store.GetStore().CreateOrUpdateDisplayNameLabel(projectID, model.SmartCRMEventSourceHubspot, propertyKey, property.Options[i].Value, property.Options[i].Label)
				if status == http.StatusBadRequest || status == http.StatusInternalServerError {
					logCtx.WithFields(log.Fields{"key": propertyKey, "value": property.Options[i].Value, "label": property.Options[i].Label}).
						Error("Failed to create or update display name label from reference field")
					failures = true
					continue
				}
			}
		}
	}

	return failures
}

func SyncOwnerReferenceFields(projectID int64, propertiesMetaMap map[string][]PropertyDetail, recordsMaxCreatedAtSec int64) bool {
	logCtx := log.WithFields(log.Fields{"project_id": projectID})

	failures := false

	ownerRecords, status := store.GetStore().GetHubspotDocumentsByTypeForSync(projectID, model.HubspotDocumentTypeOwner, recordsMaxCreatedAtSec)
	if status != http.StatusNotFound && status != http.StatusFound {
		logCtx.WithFields(log.Fields{"doc_type": model.HubspotDocumentTypeOwner,
			"recordsMaxCreatedAtSec": recordsMaxCreatedAtSec}).Error("Failed to get hubspot owner document for sync.")
		return true
	} else if status == http.StatusNotFound {
		logCtx.WithFields(log.Fields{"doc_type": model.HubspotDocumentTypeOwner,
			"recordsMaxCreatedAtSec": recordsMaxCreatedAtSec}).Warning("No hubspot owner document available for sync.")
		return false
	}

	for _, document := range ownerRecords {
		for docType, propertiesMeta := range propertiesMetaMap {
			for _, property := range propertiesMeta {
				value, err := U.DecodePostgresJsonb(document.Value)
				if err != nil {
					logCtx.WithFields(log.Fields{"doc_type": docType, "owner_doc_id": document.ID, "timestamp": document.Timestamp}).
						WithError(err).Error("Error occured during unmarshal of hubspot owner document")
					failures = true
					continue
				}

				propertyKey := model.GetCRMEnrichPropertyKeyByType(model.SmartCRMEventSourceHubspot, docType, property.Name)
				ownerId := U.GetPropertyValueAsString((*value)["ownerId"])

				firstName := U.GetPropertyValueAsString((*value)["firstName"])
				lastName := U.GetPropertyValueAsString((*value)["lastName"])
				label := strings.TrimSpace(firstName + " " + lastName)
				if label == "" {
					continue
				}

				status = store.GetStore().CreateOrUpdateDisplayNameLabel(projectID, U.CRM_SOURCE_NAME_HUBSPOT, propertyKey, ownerId, label)
				if status != http.StatusCreated && status != http.StatusConflict && status != http.StatusAccepted {
					logCtx.WithFields(log.Fields{"key": propertyKey, "value": ownerId, "label": label}).
						Error("Failed to create or update display name label from reference field")
					failures = true
					continue
				}
			}
		}

		status = store.GetStore().UpdateHubspotDocumentAsSynced(projectID, document.ID, model.HubspotDocumentTypeOwner, "", document.Timestamp, document.Action, "", "")
		if status != http.StatusAccepted {
			logCtx.WithFields(log.Fields{"owner_doc_id": document.ID, "timestamp": document.Timestamp}).Error("Failed to update hubspot owner document as synced.")
			failures = true
			continue
		}
	}

	return failures
}

type Stage struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

type DealPipeline struct {
	ID     string  `json:"id"`
	Label  string  `json:"label"`
	Stages []Stage `json:"stages"`
}

type DealPipelineResults struct {
	Results []DealPipeline `json:"results"`
}

func GetHubspotDealStagesAndPipelines(apiKey, refreshToken string) ([]DealPipeline, error) {
	if apiKey == "" && refreshToken == "" {
		return nil, errors.New("missing api key and refresh token on GetHubspotDealStagesAndPipelines")
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

	url := "https://api.hubapi.com/crm/v3/pipelines/deals?"
	resp, err := model.ActionHubspotRequestHandler("GET", url, apiKey, accessToken, "", nil)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		var body interface{}
		json.NewDecoder(resp.Body).Decode(&body)
		log.WithFields(log.Fields{"resp_body": body}).Error("Failed to GetHubspotDealStagesAndPipelines")
		return nil, fmt.Errorf("error while query data %s ", body)
	}

	var results DealPipelineResults
	err = json.NewDecoder(resp.Body).Decode(&results)
	if err != nil {
		return nil, err
	}

	return results.Results, nil
}

func syncHubspotDealStageAndPipeline(projectID int64, apiKey, refreshToken string) bool {
	logCtx := log.WithFields(log.Fields{"project_id": projectID})
	failures := false

	if projectID == 0 || (apiKey == "" && refreshToken == "") {
		logCtx.Error("Invalid parameters on syncHubspotDealStageAndPipeline")
		return true
	}

	dealPipelines, err := GetHubspotDealStagesAndPipelines(apiKey, refreshToken)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get deal stage and pipeline.")
		return true
	}

	for i := range dealPipelines {
		if dealPipelines[i].ID != "" && dealPipelines[i].Label != "" {
			propertyKey := model.GetCRMEnrichPropertyKeyByType(model.SmartCRMEventSourceHubspot, model.HubspotDocumentTypeNameDeal, "pipeline")
			status := store.GetStore().CreateOrUpdateDisplayNameLabel(projectID, model.SmartCRMEventSourceHubspot, propertyKey, dealPipelines[i].ID, dealPipelines[i].Label)
			if status == http.StatusBadRequest || status == http.StatusInternalServerError {
				logCtx.WithFields(log.Fields{"key": propertyKey, "value": dealPipelines[i].ID, "label": dealPipelines[i].Label}).
					Error("Failed to create or update display name label from deal pipeline")
				failures = true
				continue
			}
		}

		for _, dealStage := range dealPipelines[i].Stages {
			if dealStage.ID == "" || dealStage.Label == "" {
				continue
			}

			propertyKey := model.GetCRMEnrichPropertyKeyByType(model.SmartCRMEventSourceHubspot, model.HubspotDocumentTypeNameDeal, "dealstage")
			status := store.GetStore().CreateOrUpdateDisplayNameLabel(projectID, model.SmartCRMEventSourceHubspot, propertyKey, dealStage.ID, dealStage.Label)
			if status == http.StatusBadRequest || status == http.StatusInternalServerError {
				logCtx.WithFields(log.Fields{"key": propertyKey, "value": dealStage.ID, "label": dealStage.Label}).
					Error("Failed to create or update display name label from deal stage")
				failures = true
				continue
			}
		}
	}

	return failures
}

func SyncReferenceField(projectID int64, apiKey, refreshToken string, hubspotAllowedDocTypes []string, recordsMaxCreatedAtSec int64) bool {
	logCtx := log.WithFields(log.Fields{"project_id": projectID})

	propertiesMetaMap, failures := GetHubspotPropertyDetailsByDataType(apiKey, refreshToken, hubspotAllowedDocTypes)

	externalOptionsReferencedRecords := make(map[string][]PropertyDetail, 0)
	propertiesOptions := make(map[string][]PropertyDetail)
	for docType, propertiesMeta := range propertiesMetaMap {
		logCtx.WithFields(log.Fields{"doc_type": docType})

		for _, property := range propertiesMeta {
			if property.ExternalOptionsReferenceType != "OWNER" && len(property.Options) == 0 {
				logCtx.WithField("key", property.Name).Warning("No external options reference type in syncReferenceField")
				continue
			} else if property.ExternalOptionsReferenceType == "OWNER" {
				if _, exists := externalOptionsReferencedRecords[docType]; !exists {
					externalOptionsReferencedRecords[docType] = make([]PropertyDetail, 0)
				}
				externalOptionsReferencedRecords[docType] = append(externalOptionsReferencedRecords[docType], property)
				continue
			} else {
				if _, exists := propertiesOptions[docType]; !exists {
					propertiesOptions[docType] = make([]PropertyDetail, 0)
				}
				propertiesOptions[docType] = append(propertiesOptions[docType], property)
				continue
			}
		}
	}

	if len(propertiesOptions) > 0 {
		propertyOptionsFailures := SyncPropertiesOptions(projectID, propertiesOptions)
		if propertyOptionsFailures {
			failures = true
		}
	}

	if len(externalOptionsReferencedRecords) > 0 {
		ownerReferenceFieldsFailures := SyncOwnerReferenceFields(projectID, externalOptionsReferencedRecords, recordsMaxCreatedAtSec)
		if ownerReferenceFieldsFailures {
			failures = true
		}
	}

	log.Info(fmt.Sprintf("Starting sync engagement call disposition for project %d", projectID))
	callDispositionFailures := SyncHubspotEngagementCallDispositions(projectID, apiKey, refreshToken)
	if callDispositionFailures {
		failures = true
	}
	log.Info(fmt.Sprintf("Synced engagement call disposition for project %d", projectID))

	log.Info(fmt.Sprintf("Starting sync deal stage and pipeline for project %d", projectID))
	dealStageAndPipelineFailures := syncHubspotDealStageAndPipeline(projectID, apiKey, refreshToken)
	if dealStageAndPipelineFailures {
		failures = true
	}
	log.Info(fmt.Sprintf("Synced deal stage and pipeline for project %d", projectID))

	return failures
}

type EngagementCallDisposition struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

func GetHubspotEngagementCallDispositions(apiKey, refreshToken string) ([]EngagementCallDisposition, error) {
	if apiKey == "" && refreshToken == "" {
		return nil, errors.New("missing api key and refresh token on GetHubspotEngagementDispositions")
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

	url := "https://api.hubapi.com/calling/v1/dispositions?"
	resp, err := model.ActionHubspotRequestHandler("GET", url, apiKey, accessToken, "", nil)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		var body interface{}
		json.NewDecoder(resp.Body).Decode(&body)
		log.WithFields(log.Fields{"resp_body": body}).Error("Failed to GetHubspotEngagementDispositions")
		return nil, fmt.Errorf("error while query data %s ", body)
	}

	var callDispositions []EngagementCallDisposition
	err = json.NewDecoder(resp.Body).Decode(&callDispositions)
	if err != nil {
		return nil, err
	}

	return callDispositions, nil
}

func SyncHubspotEngagementCallDispositions(projectID int64, apiKey, refreshToken string) bool {
	logCtx := log.WithFields(log.Fields{"project_id": projectID})
	failures := false

	if projectID == 0 || (apiKey == "" && refreshToken == "") {
		logCtx.Error("Invalid parameters on syncDispositionLabels")
		return true
	}

	callDispositions, err := GetHubspotEngagementCallDispositions(apiKey, refreshToken)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get engagement call dispositions.")
		return true
	}

	for i := range callDispositions {
		if callDispositions[i].ID == "" || callDispositions[i].Label == "" {
			continue
		}

		propertyKey := model.GetCRMEnrichPropertyKeyByType(model.SmartCRMEventSourceHubspot, model.HubspotDocumentTypeNameEngagement, "disposition")
		status := store.GetStore().CreateOrUpdateDisplayNameLabel(projectID, model.SmartCRMEventSourceHubspot, propertyKey, callDispositions[i].ID, callDispositions[i].Label)
		if status == http.StatusBadRequest || status == http.StatusInternalServerError {
			logCtx.WithFields(log.Fields{"key": propertyKey, "value": callDispositions[i].ID, "label": callDispositions[i].Label}).
				Error("Failed to create or update display name label from engagement call disposition")
			failures = true
			continue
		}
	}

	return failures
}

func syncContact(project *model.Project, otpRules *[]model.OTPRule, uniqueOTPEventKeys *[]string, document *model.HubspotDocument, hubspotSmartEventNames []HubspotSmartEventName) int {
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
			_, deleteContactUserID, status = store.GetStore().GetUserIdFromEventId(project.ID, contactDocuments[0].SyncId, "")
			if status != http.StatusFound {
				logCtx.WithField("delete_contact", contactDocuments[0].ID).Error(
					"Failed to get merged contact created event for getting user id.")
				return http.StatusInternalServerError
			}
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
							_, mergedContactUserID, status = store.GetStore().GetUserIdFromEventId(project.ID, mergedContact.SyncId, "")
							if status != http.StatusFound {
								logCtx.WithField("merged_contact", mergedContact.ID).Error(
									"Failed to get merged contact created event for getting user id.")
								continue
							}
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

	customIdentification, isCustomIdentificationEnabled := getCustomIdentification(project.ID, document)

	primaryIdentification := primaryEmail

	if isCustomIdentificationEnabled {
		logCtx.WithFields(log.Fields{"custom_identification": customIdentification}).Info("Using custom identification")
		primaryIdentification = customIdentification
		// set primaryEmail and secondaryEmails as empty for not re identifying those users with primaryIdentification
		primaryEmail = ""
		secondaryEmails = []string{}
	} else if primaryEmail == "" && len(secondaryEmails) > 0 {
		primaryIdentification = secondaryEmails[0]
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

	customerUserID := ""
	if !isCustomIdentificationEnabled {
		_, customerUserID = getCustomerUserIDFromProperties(project.ID, *enProperties)
	}

	emails := []string{}
	userByCustomerUserID := make(map[string]string)
	if C.AllowIdentificationOverwriteUsingSource(project.ID) {
		emails = append([]string{primaryIdentification}, secondaryEmails...)

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
			userByCustomerUserID[createdUserID] = customerUserID
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
			_, userID, status = store.GetStore().GetUserIdFromEventId(project.ID, createdDocuments[0].SyncId, "")
			if status != http.StatusFound {
				logCtx.WithField("event_id", createdDocuments[0].SyncId).Error(
					"Failed to get contact created event for getting user id.")
				return http.StatusInternalServerError
			}

			status = store.GetStore().UpdateHubspotDocumentAsSynced(
				project.ID, document.ID, model.HubspotDocumentTypeContact, createdDocuments[0].SyncId, createdDocuments[0].Timestamp, model.HubspotDocumentActionCreated, userID, "")
			if status != http.StatusAccepted {
				logCtx.Error("Failed to update hubspot contact created document user id.")
			}
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
	if C.AllowIdentificationOverwriteUsingSource(project.ID) && primaryIdentification != "" {

		for userID := range userByCustomerUserID {
			user, status := store.GetStore().GetUserWithoutProperties(project.ID, userID)
			if status != http.StatusFound {
				logCtx.WithFields(log.Fields{"err_code": status, "user_id": userID}).Error("Failed to get user from hubspot re-identification.")
				continue
			}

			if user.Source != nil && *user.Source != model.UserSourceHubspot && *user.Source != model.UserSourceWeb {
				continue
			}

			if user.CustomerUserId == primaryIdentification {
				continue
			}

			status, _ = SDK.Identify(project.ID, &SDK.IdentifyPayload{
				UserId: userID, CustomerUserId: primaryIdentification, RequestSource: model.UserSourceHubspot, Source: SDK.SourceHubspot}, true)
			if status != http.StatusOK {
				logCtx.WithFields(log.Fields{"primary_identification": primaryIdentification, "user_id": userID}).Error(
					"Failed to identify user with primary identification.")
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

	existingCustomerUserID, status := store.GetStore().GetCustomerUserIdFromUserId(project.ID, userID)
	if status != http.StatusFound {
		logCtx.WithField("error_code", status).Error("Failed to get user on sync contact.")
		return http.StatusInternalServerError
	}

	if existingCustomerUserID != customerUserID {
		logCtx.WithFields(log.Fields{"existing_customer_user_id": existingCustomerUserID, "new_customer_user_id": customerUserID}).
			Warn("Different customer user id seen on sync contact")
	}

	if document.Action == model.HubspotDocumentActionUpdated && !C.IsProjectIDSkippedForOtp(project.ID) {
		err = ApplyHSOfflineTouchPointRule(project, otpRules, uniqueOTPEventKeys, trackPayload, document, defaultSmartEventTimestamp)
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
		logCtx.WithFields(log.Fields{"ProjectID": project.ID}).Info("Invoking method syncContactFormSubmission")
		syncContactFormSubmissions(project, otpRules, uniqueOTPEventKeys, userID, document)
	}

	errCode := store.GetStore().UpdateHubspotDocumentAsSynced(
		project.ID, document.ID, model.HubspotDocumentTypeContact, eventID, document.Timestamp, document.Action, userID, "")
	if errCode != http.StatusAccepted {
		logCtx.Error("Failed to update hubspot contact document as synced.")
		return http.StatusInternalServerError
	}

	return http.StatusOK
}

// Check if the condition are satisfied for creating OTP events for each rule for HS Contact
func ApplyHSOfflineTouchPointRule(project *model.Project, otpRules *[]model.OTPRule, uniqueOTPEventKeys *[]string, trackPayload *SDK.TrackPayload, document *model.HubspotDocument, lastModifiedTimeStamp int64) error {

	logCtx := log.WithFields(log.Fields{"project_id": project.ID, "method": "ApplyHSOfflineTouchPointRule",
		"document_id": document.ID, "document_action": document.Action, "document": document})

	if otpRules == nil || project == nil || trackPayload == nil || document == nil {
		return nil
	}

	lastModifiedTimeStamp = U.CheckAndGetStandardTimestamp(lastModifiedTimeStamp)

	// Get the last sync doc for the current update doc.
	prevDoc, status := store.GetStore().GetLastSyncedHubspotUpdateDocumentByID(document.ProjectId, document.ID, document.Type)
	if status != http.StatusFound {
		// In case no prev properties
		prevDoc = nil
	}

	for _, rule := range *otpRules {

		otpUniqueKey, err := createOTPUniqueKeyForFormsAndContacts(rule, trackPayload)
		if err != http.StatusCreated {
			logCtx.Warn("Failed to create otp_unique_key")
			continue
		}

		// Check if rule is applicable & the record has changed property w.r.t filters
		if !canCreateHSTouchPoint(document.Action) {
			continue
		}
		if !filterCheck(rule, trackPayload, document, prevDoc, logCtx) {
			continue
		}
		//Checks if the otpUniqueKey is already present in other OTP Event Properties
		if !IsOTPKeyUnique(otpUniqueKey, uniqueOTPEventKeys, logCtx) {
			continue
		}

		_, err1 := CreateTouchPointEventForFormsAndContacts(project, trackPayload, document, rule, lastModifiedTimeStamp, otpUniqueKey)
		if err1 != nil {
			logCtx.WithError(err1).Error("failed to create touch point for hubspot contact updated document.")
			continue
		}

	}
	return nil
}

// Check if the condition are satisfied for creating OTP events for each rule for HS Forms Submission
func ApplyHSOfflineTouchPointRuleForForms(project *model.Project, otpRules *[]model.OTPRule, uniqueOTPEventKeys *[]string, trackPayload *SDK.TrackPayload, document *model.HubspotDocument, formTimestamp int64) error {

	logCtx := log.WithFields(log.Fields{"project_id": project.ID, "method": "ApplyHSOfflineTouchPointRuleForForms",
		"document_id": document.ID, "document_action": document.Action, "document": document})

	if otpRules == nil || project == nil || trackPayload == nil || document == nil {
		logCtx.Error("something is empty")
		return nil
	}
	logCtx.WithFields(log.Fields{"ProjectID": project.ID}).Info("Inside method ApplyHSOfflineTouchPointRuleForForms")
	formTimestamp = U.CheckAndGetStandardTimestamp(formTimestamp)

	for _, rule := range *otpRules {

		otpUniqueKey, err := createOTPUniqueKeyForFormsAndContacts(rule, trackPayload)
		if err != http.StatusCreated {
			logCtx.Warn("Failed to create otp_unique_key")
			continue
		}

		// Check if rule is applicable & the record has changed property w.r.t filters
		if rule.RuleType != model.TouchPointRuleTypeForms {
			logCtx.Info("Rule Type is failing the OTP event creation.")
			continue
		}
		if !filterCheckGeneral(rule, trackPayload, logCtx) {
			continue
		}
		//Checks if the otpUniqueKey is already present in other OTP Event Properties
		if !IsOTPKeyUnique(otpUniqueKey, uniqueOTPEventKeys, logCtx) {
			continue
		}

		logCtx.WithFields(log.Fields{"ProjectID": project.ID, "OTPrules": rule}).Info("Invoking method CreateTouchPointEvent")
		_, err1 := CreateTouchPointEventForFormsAndContacts(project, trackPayload, document, rule, formTimestamp, otpUniqueKey)
		if err1 != nil {
			logCtx.WithError(err1).Error("failed to create touch point for hubspot contact updated document.")
			continue
		}

	}
	return nil
}

// Check if the condition are satisfied for creating OTP events for each rule for HS Engagements - Meetings/Calls/Emails
func ApplyHSOfflineTouchPointRuleForEngagement(project *model.Project, otpRules *[]model.OTPRule, uniqueOTPEventKeys *[]string, trackPayload *SDK.TrackPayload,
	document *model.HubspotDocument, threadID string, engagementType string) error {

	logCtx := log.WithFields(log.Fields{"project_id": project.ID, "method": "ApplyHSOfflineTouchPointRuleForEngagement",
		"document_id": document.ID, "document_action": document.Action, "document": document})

	if otpRules == nil || project == nil || trackPayload == nil || document == nil {
		logCtx.Error("something is empty")
		return nil
	}
	for _, rule := range *otpRules {

		otpUniqueKey, err := createOTPUniqueKeyForEngagements(rule, trackPayload, engagementType, logCtx)
		if err != http.StatusCreated {
			logCtx.Warn("Failed to create otp_unique_key")
			continue
		}
		// Check if rule is applicable & the record has changed property w.r.t filters
		if !canCreateHSEngagementTouchPoint(engagementType, rule.RuleType) {
			logCtx.Info("Rule Type is failing the OTP event creation.")
			continue
		}
		if !filterCheckGeneral(rule, trackPayload, logCtx) {
			continue
		}
		//Checks if the otpUniqueKey is already present in other OTP Event Properties
		if !IsOTPKeyUnique(otpUniqueKey, uniqueOTPEventKeys, logCtx) {
			continue
		}

		_, err1 := CreateTouchPointEventForEngagement(project, trackPayload, document, rule, threadID, engagementType, otpUniqueKey)
		if err1 != nil {
			logCtx.WithError(err1).Error("failed to create touch point for hubspot contact updated document.")
			continue

		}
	}
	return nil
}

// Check if the condition are satisfied for creating OTP events for each rule for HS Contact list
func ApplyHSOfflineTouchPointRuleForContactList(project *model.Project, otpRules *[]model.OTPRule, uniqueOTPEventKeys *[]string, trackPayload *SDK.TrackPayload, document *model.HubspotDocument) error {

	logCtx := log.WithFields(log.Fields{"project_id": project.ID, "method": "ApplyHSOfflineTouchPointRuleForContactList",
		"document_id": document.ID, "document_action": document.Action, "document": document})

	if otpRules == nil || project == nil || trackPayload == nil || document == nil {
		logCtx.Error("something is empty")
		return nil
	}
	for _, rule := range *otpRules {

		otpUniqueKey, err := createOTPUniqueKeyForContactList(rule, trackPayload, logCtx)
		if err != http.StatusCreated {
			logCtx.Warn("Failed to create otp_unique_key")
			continue
		}

		// Check if rule is applicable & the record has changed property w.r.t filters
		if rule.RuleType != model.TouchPointRuleTypeContactList {
			logCtx.Info("Rule Type is failing the OTP event creation.")
			continue
		}
		if !filterCheckGeneral(rule, trackPayload, logCtx) {
			continue
		}
		//Checks if the otpUniqueKey is already present in other OTP Event Properties
		if !IsOTPKeyUnique(otpUniqueKey, uniqueOTPEventKeys, logCtx) {
			continue
		}
		_, err1 := CreateTouchPointEventForLists(project, trackPayload, document, rule, otpUniqueKey)
		if err1 != nil {
			logCtx.WithError(err1).Error("failed to create touch point for hubspot lists.")
			continue
		}

	}
	return nil
}

// CreateTouchPointEventForFormsAndContacts - Creates offline touchpoint for HS create/update events with given rule for HS Contacts and Forms
func CreateTouchPointEventForFormsAndContacts(project *model.Project, trackPayload *SDK.TrackPayload, document *model.HubspotDocument,
	rule model.OTPRule, lastModifiedTimeStamp int64, otpUniqueKey string) (*SDK.TrackResponse, error) {

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

	// Adding mandatory properties
	payload.EventProperties[U.EP_OTP_RULE_ID] = rule.ID
	payload.EventProperties[U.EP_OTP_UNIQUE_KEY] = otpUniqueKey
	payload.Timestamp = timestamp

	// Mapping touch point properties:
	var rulePropertiesMap map[string]model.TouchPointPropertyValue
	err = U.DecodePostgresJsonbToStructType(&rule.PropertiesMap, &rulePropertiesMap)
	if err != nil {
		logCtx.WithField("Document", trackPayload).WithError(err).Error("Failed to decode/fetch offline touch point rule PROPERTIES for hubspot document.")
		return trackResponse, errors.New(fmt.Sprintf("create hubspot touchpoint event track failed for doc type %d, message %s", document.Type, trackResponse.Error))
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

// CreateTouchPointEventForLists - Creates OTP for HS lists
func CreateTouchPointEventForLists(project *model.Project, trackPayload *SDK.TrackPayload, document *model.HubspotDocument,
	rule model.OTPRule, otpUniqueKey string) (*SDK.TrackResponse, error) {

	logCtx := log.WithFields(log.Fields{"project_id": project.ID, "method": "CreateTouchPointEventForLists",
		"document_id": document.ID, "document_action": document.Action})

	logCtx.WithField("document", document).WithField("trackPayload", trackPayload).
		Info("CreateTouchPointEventForLists: creating hubspot offline touch point document")

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
		logCtx.WithField("Document", trackPayload).WithError(err).Error("Failed to decode/fetch offline touch point rule PROPERTIES for hubspot document.")
		return trackResponse, errors.New(fmt.Sprintf("create hubspot list touchpoint event track failed for doc type %d, message %s", document.Type, trackResponse.Error))
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

	status, trackResponse := SDK.Track(project.ID, payload, true, sdk.SourceHubspot, "")
	if status != http.StatusOK && status != http.StatusFound && status != http.StatusNotModified {
		logCtx.WithField("Document", trackPayload).WithError(err).Error(fmt.Errorf("create "+
			"hubspot list touchpoint event track failed for doc type %d, message %s", document.Type, trackResponse.Error))
		return trackResponse, errors.New(fmt.Sprintf("create hubspot touchpoint event "+
			"track failed for doc type %d, message %s", document.Type, trackResponse.Error))
	}

	logCtx.WithField("statusCode", status).WithField("trackResponsePayload", trackResponse).Info("Successfully: created hubspot lists offline touch point")
	return trackResponse, nil
}

// isEmailEngagementAlreadyTracked- Checks if the Email (a type of Engagement) is already tracked for creating OTP event.
func isEmailEngagementAlreadyTracked(projectID int64, ruleID string, threadID string, logCtx *log.Entry) (bool, error) {

	en, status := store.GetStore().CreateOrGetOfflineTouchPointEventName(projectID)
	if status != http.StatusFound && status != http.StatusConflict && status != http.StatusCreated {
		logCtx.Error("failed to create event name on SF for offline touch point")
		return false, errors.New(fmt.Sprintf("failed to create event name on SF for offline touch point"))
	}

	last30DaysOTPEvents, err := store.GetStore().GetEventsByEventNameId(projectID, en.ID, time.Now().Unix()-U.MonthInSecs, time.Now().Unix())
	if err != http.StatusFound && err != http.StatusNotFound {
		logCtx.Info("no events found for engagement, continuing")
	} else {

		for _, event := range last30DaysOTPEvents {
			propertiesMap := make(map[string]interface{})
			err := json.Unmarshal(event.Properties.RawMessage, &propertiesMap)
			if err != nil {
				log.Error("Error occurred during unmarshal of otp event properties, continuing")
				continue
			}
			threadIDTracked, threadExists := propertiesMap[U.EP_HUBSPOT_ENGAGEMENT_THREAD_ID]
			if threadExists && threadID == threadIDTracked {
				ruleIDTracked, ruleExists := propertiesMap[U.EP_OTP_RULE_ID]
				if ruleExists && ruleIDTracked == ruleID {
					logCtx.Info("OTP already created for the rule and email engagement, skipping this thread")
					return true, nil
				}
			}
		}
	}

	return false, nil
}

func getThreadIDFromEngagementV3(engagement EngagementsV3, engagementType string) (string, error) {
	threadID, isPresent := engagement.Properties["hs_email_thread_id"]
	if !isPresent {
		return "", errors.New("couldn't get the threadID on hubspot email engagement_v3, logging and continuing")
	}

	threadIDStr := U.GetPropertyValueAsString(threadID)

	return threadIDStr, nil
}

func getThreadIDFromOldEngagement(engagement Engagements, engagementType string) (string, error) {
	threadID, isPresent := engagement.Metadata["threadId"]
	if !isPresent {
		return "", errors.New("couldn't get the threadID on hubspot email engagement, logging and continuing")
	}

	threadIDStr := U.GetPropertyValueAsString(threadID)

	return threadIDStr, nil
}

func getThreadIDFromEngagement(engagement interface{}, engagementType string) (string, error) {
	if engagementType != EngagementTypeEmail && engagementType != EngagementTypeIncomingEmail {
		return "", nil
	}

	switch record := engagement.(type) {
	case EngagementsV3:
		return getThreadIDFromEngagementV3(record, engagementType)
	case Engagements:
		return getThreadIDFromOldEngagement(record, engagementType)
	default:
		return "", errors.New("failed to get type of engagement record")
	}
}

// CreateTouchPointEventForEngagement - Creates offline touchpoint for HS engagements (calls, meetings, forms, emails) with give rule
func CreateTouchPointEventForEngagement(project *model.Project, trackPayload *SDK.TrackPayload, document *model.HubspotDocument,
	rule model.OTPRule, threadID string, engagementType string, otpUniqueKey string) (*SDK.TrackResponse, error) {

	logCtx := log.WithFields(log.Fields{"project_id": project.ID, "method": "CreateTouchPointEvent",
		"document_id": document.ID, "document_action": document.Action, "threadID": threadID})

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
	case EngagementTypeEmail, EngagementTypeIncomingEmail, EngagementTypeMeeting, EngagementTypeCall:
		{
			if threadID != "" {
				found, errT := isEmailEngagementAlreadyTracked(project.ID, rule.ID, threadID, logCtx)
				if found || errT != nil {
					return trackResponse, errT
				}
			}

			payload.EventProperties[U.EP_HUBSPOT_ENGAGEMENT_THREAD_ID] = threadID

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

	// Adding mandatory properties
	payload.EventProperties[U.EP_OTP_RULE_ID] = rule.ID
	payload.EventProperties[U.EP_OTP_UNIQUE_KEY] = otpUniqueKey
	payload.Timestamp = timestamp

	// Mapping touch point properties:
	var rulePropertiesMap map[string]model.TouchPointPropertyValue
	err = U.DecodePostgresJsonbToStructType(&rule.PropertiesMap, &rulePropertiesMap)
	if err != nil {
		logCtx.WithField("Document", trackPayload).WithError(err).Error("Failed to decode/fetch offline touch point rule PROPERTIES for hubspot document.")
		return trackResponse, errors.New(fmt.Sprintf("create hubspot touchpoint event track failed for doc type %d, message %s", document.Type, trackResponse.Error))
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

// IsOTPKeyUnique Returns true or false if the otpKey (userID+ruleID+keyID) is not present in uniqueOTPEventKeys i.e. Unique OTP key.
func IsOTPKeyUnique(otpUniqueKey string, uniqueOTPEventKeys *[]string, logCtx *log.Entry) bool {
	isUnique := !U.StringValueIn(otpUniqueKey, *uniqueOTPEventKeys)
	if !isUnique {
		log.WithField("uniqueOTPEventKeys", uniqueOTPEventKeys).WithField("otpUniqueKey", otpUniqueKey).Warn("The OTP Key is not unique.")
	}
	return isUnique
}

// Creates a unique key using ruleID, userID and engagementID as keyID for Engagements {Emails, Calls and Meetings}
func createOTPUniqueKeyForEngagements(rule model.OTPRule, trackPayload *SDK.TrackPayload, engagementType string, logCtx *log.Entry) (string, int) {

	ruleID := rule.ID
	userID := trackPayload.UserId
	var keyID string
	var uniqueKey string

	switch engagementType {

	case EngagementTypeEmail, EngagementTypeIncomingEmail, EngagementTypeMeeting, EngagementTypeCall:
		if _, exists := trackPayload.EventProperties[U.EP_HUBSPOT_ENGAGEMENT_ID]; exists {
			keyID = fmt.Sprintf("%v", trackPayload.EventProperties[U.EP_HUBSPOT_ENGAGEMENT_ID])
		} else {
			logCtx.Error("Event Property $hubspot_engagement_id does not exist.")
			return uniqueKey, http.StatusNotFound
		}

	default:
		logCtx.Error("engagement type not supported yet for otp_unique_key creation")
		return uniqueKey, http.StatusNotFound
	}

	uniqueKey = userID + ruleID + keyID
	return uniqueKey, http.StatusCreated
}

// Creates a unique key using ruleID, userID and eventID as keyID for Forms and contacts
func createOTPUniqueKeyForFormsAndContacts(rule model.OTPRule, trackPayload *SDK.TrackPayload) (string, int) {

	ruleID := rule.ID
	userID := trackPayload.UserId
	keyID := trackPayload.EventId

	uniqueKey := userID + ruleID + keyID

	return uniqueKey, http.StatusCreated

}

// Creates a unique key using ruleID, userID and contact lists list ID  as keyID for Contact Lists
func createOTPUniqueKeyForContactList(rule model.OTPRule, trackPayload *SDK.TrackPayload, logCtx *log.Entry) (string, int) {
	ruleID := rule.ID
	userID := trackPayload.UserId
	var keyID string
	var uniqueKey string

	if _, exists := trackPayload.EventProperties[U.EP_HUBSPOT_CONTACT_LIST_LIST_ID]; exists {
		keyID = fmt.Sprintf("%v", trackPayload.EventProperties[U.EP_HUBSPOT_CONTACT_LIST_LIST_ID])
	} else {
		logCtx.Error("Event Property $hubspot_contact_list_list_id does not exist.")
		return uniqueKey, http.StatusNotFound
	}

	uniqueKey = userID + ruleID + keyID
	return uniqueKey, http.StatusCreated

}

// canCreateHSEngagementTouchPoint- Checks if the rule type of OTP rule is in accordance with the engagement type.
func canCreateHSEngagementTouchPoint(engagementType string, ruleType string) bool {

	switch engagementType {

	case EngagementTypeEmail, EngagementTypeIncomingEmail:
		if ruleType == model.TouchPointRuleTypeEmails {
			return true
		}
	case EngagementTypeMeeting:
		if ruleType == model.TouchPointRuleTypeMeetings {
			return true
		}
	case EngagementTypeCall:
		if ruleType == model.TouchPointRuleTypeCalls {
			return true
		}
	default:
		return false
	}
	return false
}

// canCreateHSTouchPoint- Returns true if the document action type is Updated for HS Contacts.
func canCreateHSTouchPoint(documentActionType int) bool {
	// Ignore doc types other than HubspotDocumentActionUpdated
	if documentActionType != model.HubspotDocumentActionUpdated {
		return false
	}
	return true
}

// filterCheck- Checks if all the filters applied are passed and checks HS documents for HS Contacts
func filterCheck(rule model.OTPRule, trackPayload *SDK.TrackPayload, document *model.HubspotDocument, prevDoc *model.HubspotDocument, logCtx *log.Entry) bool {

	var ruleFilters []model.TouchPointFilter
	err := U.DecodePostgresJsonbToStructType(&rule.Filters, &ruleFilters)
	if err != nil {
		logCtx.WithField("Document", trackPayload).WithError(err).Error("Failed to decode/fetch offline touch point rule FILTERS for Hubspot document.")
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
	if filtersPassed != 0 && filtersPassed == len(ruleFilters) {
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
		for _, filter := range ruleFilters {
			if val1, exists1 := trackPayload.EventProperties[filter.Property]; exists1 {
				if val2, exists2 := (*prevProperties)[filter.Property]; exists2 {
					if val1 == val2 {
						samePropertyMatchingScore++
					}
				}
			}
		}
		// If all filter properties matches with that of the previous found properties, skip and fail
		if samePropertyMatchingScore == len(ruleFilters) {
			return false
		} else {
			return true
		}
	}
	// When neither filters matched nor (filters matched but values are same)
	return false
}

// filterCheckGeneral- Returns true if all the filters applied are passed.
func filterCheckGeneral(rule model.OTPRule, trackPayload *SDK.TrackPayload, logCtx *log.Entry) bool {

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
	if filtersPassed != 0 && filtersPassed == len(ruleFilters) {
		return true
	}

	// When neither filters matched nor (filters matched but values are same)
	logCtx.Warn("Filter check general is failing for offline touch point rule")
	return false
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

	isCompanyV3, err := checkIfCompanyV3(document)
	if err != nil {
		return "", "", err
	}

	if isCompanyV3 {
		var company CompanyV3
		err = json.Unmarshal(document.Value.RawMessage, &company)
		if err != nil {
			return "", "", err
		}

		companyName := company.Properties["name"]
		domainName := company.Properties["domain"]

		return companyName, domainName, nil
	}

	var company Company
	err = json.Unmarshal(document.Value.RawMessage, &company)
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

	isCompanyV3, err := checkIfCompanyV3(document)
	if err != nil {
		return nil, err
	}

	if isCompanyV3 {
		return getCompanyPropertiesV3(projectID, document)
	}

	var company Company
	err = json.Unmarshal((document.Value).RawMessage, &company)
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

func checkIfCompanyV3(document *model.HubspotDocument) (bool, error) {
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

func syncCompany(projectID int64, document *model.HubspotDocument) int {
	var value interface{}
	err := U.DecodePostgresJsonbToStructType(document.Value, &value)
	if err != nil {
		log.WithFields(log.Fields{"project_id": projectID})
		return http.StatusInternalServerError
	}

	isCompanyV3, err := checkIfCompanyV3(document)
	if err != nil {
		log.WithFields(log.Fields{"project_id": projectID}).WithError(err).Error("failed to check type of company record")
	}

	if isCompanyV3 {
		return syncCompanyV3(projectID, document)
	}

	return syncCompanyV2(projectID, document)
}

func syncCompanyV2(projectID int64, document *model.HubspotDocument) int {
	if document.Type != model.HubspotDocumentTypeCompany {
		return http.StatusInternalServerError
	}

	logCtx := log.WithFields(log.Fields{"project_id": projectID, "document_id": document.ID,
		"doc_timestamp": document.Timestamp})

	var company Company
	err := json.Unmarshal((document.Value).RawMessage, &company)
	if err != nil {
		logCtx.WithError(err).Error("Failed to unmarshal hubspot company document.")
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
			logCtx.Error("Failed to update hubspot company document as synced.")
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
			logCtx.Error("Failed to get hubspot contact documents by type and action on sync company.")
			return errCode
		}
	}

	if C.DisableHubspotNonMarketingContactsByProjectID(projectID) && len(contactDocuments) == 0 {
		logCtx.Warning("No marketing contacts found for hubspot company.")
	}

	// update $hubspot_company_name and other company
	// properties on each associated contact user.
	for _, contactDocument := range contactDocuments {
		if contactDocument.UserId != "" {
			if C.IsAllowedHubspotGroupsByProjectID(projectID) {
				logCtx.Info("Updating user company group user id.")
				_, status := store.GetStore().UpdateUserGroup(projectID, contactDocument.UserId, model.GROUP_NAME_HUBSPOT_COMPANY, companyGroupID, companyUserID, false)
				if status != http.StatusAccepted && status != http.StatusNotModified {
					logCtx.Error("Failed to update user group id.")
				}
			}

			if C.EnableUserDomainsGroupByProjectID(projectID) {
				status := store.GetStore().AssociateUserDomainsGroup(projectID, contactDocument.UserId, model.GROUP_NAME_HUBSPOT_COMPANY, companyUserID)
				if status != http.StatusOK && status != http.StatusNotModified {
					logCtx.WithFields(log.Fields{"err_code": status}).Error("Failed to AssociateUserDomainsGroup on hubspot sync company.")
				}
			}
		}
	}

	// No sync_id as no event or user or one user property created.
	errCode = store.GetStore().UpdateHubspotDocumentAsSynced(projectID, document.ID, model.HubspotDocumentTypeCompany, "", document.Timestamp, document.Action, "", companyUserID)
	if errCode != http.StatusAccepted {
		logCtx.Error("Failed to update hubspot company document as synced.")
		return http.StatusInternalServerError
	}

	return http.StatusOK
}

func getCompanyPropertiesV3(projectID int64, document *model.HubspotDocument) (map[string]interface{}, error) {
	if projectID < 1 || document == nil {
		return nil, errors.New("invalid parameters")
	}

	if document.Type != model.HubspotDocumentTypeCompany {
		return nil, errors.New("invalid document type")
	}

	var company CompanyV3
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
			userProperties[U.UP_COMPANY] = value
		}

		propertyKey := model.GetCRMEnrichPropertyKeyByType(model.SmartCRMEventSourceHubspot,
			model.HubspotDocumentTypeNameCompany, key)
		value, err := getHubspotMappedDataTypeValue(projectID, U.EVENT_NAME_HUBSPOT_CONTACT_UPDATED, propertyKey,
			value, model.HubspotDocumentTypeCompany, document.GetDateProperties(), string(document.GetTimeZone()))
		if err != nil {
			log.WithFields(log.Fields{"project_id": projectID, "property_key": propertyKey}).WithError(err).Error("Failed to get property value.")
			continue
		}

		userProperties[propertyKey] = value
	}

	return userProperties, nil
}

func syncCompanyV3(projectID int64, document *model.HubspotDocument) int {
	if document.Type != model.HubspotDocumentTypeCompany {
		return http.StatusInternalServerError
	}

	logCtx := log.WithFields(log.Fields{"project_id": projectID, "document_id": document.ID,
		"doc_timestamp": document.Timestamp})

	var company CompanyV3
	err := json.Unmarshal((document.Value).RawMessage, &company)
	if err != nil {
		logCtx.WithError(err).Error("Failed to unmarshal hubspot company document.")
		return http.StatusInternalServerError
	}

	contactIds := make([]string, 0, 0)
	for i := range company.ContactIds {
		contactIds = append(contactIds, strconv.FormatInt(company.ContactIds[i], 10))
	}

	userProperties, err := getCompanyPropertiesV3(projectID, document)
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
			logCtx.Error("Failed to update hubspot company_v3 document as synced.")
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
			logCtx.Error("Failed to get hubspot contact documents by type and action on sync company.")
			return errCode
		}
	}

	if C.DisableHubspotNonMarketingContactsByProjectID(projectID) && len(contactDocuments) == 0 {
		logCtx.Warning("No marketing contacts found for hubspot company.")
	}

	// update $hubspot_company_name and other company
	// properties on each associated contact user.
	for _, contactDocument := range contactDocuments {
		if contactDocument.UserId != "" {
			if C.IsAllowedHubspotGroupsByProjectID(projectID) {
				logCtx.Info("Updating user company group user id.")
				_, status := store.GetStore().UpdateUserGroup(projectID, contactDocument.UserId, model.GROUP_NAME_HUBSPOT_COMPANY, companyGroupID, companyUserID, false)
				if status != http.StatusAccepted && status != http.StatusNotModified {
					logCtx.Error("Failed to update user group id.")
				}
			}

			if C.EnableUserDomainsGroupByProjectID(projectID) {
				status := store.GetStore().AssociateUserDomainsGroup(projectID, contactDocument.UserId, model.GROUP_NAME_HUBSPOT_COMPANY, companyUserID)
				if status != http.StatusOK && status != http.StatusNotModified {
					logCtx.WithFields(log.Fields{"err_code": status}).Error("Failed to AssociateUserDomainsGroup on hubspot sync company.")
				}
			}
		}
	}

	// No sync_id as no event or user or one user property created.
	errCode = store.GetStore().UpdateHubspotDocumentAsSynced(projectID, document.ID, model.HubspotDocumentTypeCompany, "", document.Timestamp, document.Action, "", companyUserID)
	if errCode != http.StatusAccepted {
		logCtx.Error("Failed to update hubspot company_v3 document as synced.")
		return http.StatusInternalServerError
	}

	return http.StatusOK
}

func getHubspotDateTimestampAsMidnightTimeZoneTimestamp(dateUTCMS interface{}, timeZone string) (int64, error) {
	timestamp, err := model.GetTimestampForV3Records(dateUTCMS)
	if err != nil {
		timestamp, err = model.ReadHubspotTimestamp(dateUTCMS)
		if err != nil {
			return 0, err
		}
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
			formatedTime, err := model.GetTimestampForV3Records(value)
			if err == nil {
				return getEventTimestamp(formatedTime), nil
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

	isDealV3, err := checkIfDealV3(document)
	if err != nil {
		return nil, nil, err
	}

	if isDealV3 {
		return getDealPropertiesV3(projectID, document)
	}

	var deal Deal
	err = json.Unmarshal((document.Value).RawMessage, &deal)
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
			enProperties, getEventTimestamp(document.Timestamp), getEventTimestamp(document.Timestamp), model.UserSourceHubspotString)

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
			model.UserSourceHubspotString)
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

	var dealAssociations DealAssociations
	err := json.Unmarshal((document.Value).RawMessage, &dealAssociations)
	if err != nil {
		return nil, nil, err
	}

	var contactIDs []string
	var companyIDs []string
	associatedContactIDs := dealAssociations.Associations.AssociatedContactIds
	for i := range associatedContactIDs {
		contactID := strconv.FormatInt(associatedContactIDs[i], 10)
		contactIDs = append(contactIDs, contactID)
	}

	associatedCompanyIDs := dealAssociations.Associations.AssociatedCompanyIds
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

	if C.IsAllowedDomainsGroupByProjectID(projectID) {
		domainName := ""
		if accountDomain := util.GetPropertyValueAsString((*enProperties)[model.GetCRMEnrichPropertyKeyByType(model.SmartCRMEventSourceHubspot,
			model.HubspotDocumentTypeNameCompany, "domain")]); accountDomain != "" {
			domainName = accountDomain
		} else {
			domainName = util.GetPropertyValueAsString((*enProperties)[model.GetCRMEnrichPropertyKeyByType(model.SmartCRMEventSourceHubspot,
				model.HubspotDocumentTypeNameCompany, "website")])
		}

		if domainName != "" {
			status := sdk.TrackDomainsGroup(projectID, companyUserID, model.GROUP_NAME_HUBSPOT_COMPANY, domainName, nil, document.Timestamp)
			if status != http.StatusOK {
				log.WithFields(log.Fields{"project_id": projectID, "document_id": document.ID, "timestamp": document.Timestamp}).
					Error("Failed to TrackDomainsGroup in hubspot company enrichment.")
			}
		}
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

			_, status := store.GetStore().UpdateUserGroup(projectID, userID, model.GROUP_NAME_HUBSPOT_DEAL, "", dealGroupUserID, false)
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
				Warning("Failed to get company created documents for syncGroupDeal.")
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

func checkIfDealV3(document *model.HubspotDocument) (bool, error) {
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

func syncDeal(projectID int64, document *model.HubspotDocument, hubspotSmartEventNames []HubspotSmartEventName) int {
	isDealV3, err := checkIfDealV3(document)
	if err != nil {
		log.WithFields(log.Fields{"project_id": projectID}).WithError(err).Error("failed to check type of deal record")
		return http.StatusInternalServerError
	}

	if isDealV3 {
		return syncDealV3(projectID, document, hubspotSmartEventNames)
	}

	return syncDealV2(projectID, document, hubspotSmartEventNames)
}

func syncDealV2(projectID int64, document *model.HubspotDocument, hubspotSmartEventNames []HubspotSmartEventName) int {
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

func getDealPropertiesV3(projectID int64, document *model.HubspotDocument) (*map[string]interface{}, *map[string]interface{}, error) {
	if document.Type != model.HubspotDocumentTypeDeal {
		return nil, nil, errors.New("invalid type")
	}

	var deal DealV3
	err := json.Unmarshal((document.Value).RawMessage, &deal)
	if err != nil {
		return nil, nil, err
	}

	enProperties := make(map[string]interface{}, 0)
	properties := make(map[string]interface{})
	for k, v := range deal.Properties {
		enKey := model.GetCRMEnrichPropertyKeyByType(model.SmartCRMEventSourceHubspot, model.HubspotDocumentTypeNameDeal, k)
		value, err := getHubspotMappedDataTypeValue(projectID, U.EVENT_NAME_HUBSPOT_DEAL_STATE_CHANGED, enKey,
			v, model.HubspotDocumentTypeDeal, document.GetDateProperties(), string(document.GetTimeZone()))
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

func syncDealV3(projectID int64, document *model.HubspotDocument, hubspotSmartEventNames []HubspotSmartEventName) int {
	if document.Type != model.HubspotDocumentTypeDeal {
		return http.StatusInternalServerError
	}

	logCtx := log.WithFields(log.Fields{"project_id": projectID, "document_id": document.ID})

	var deal DealV3
	err := json.Unmarshal((document.Value).RawMessage, &deal)
	if err != nil {
		logCtx.Error("Failed to unmarshal hubspot document deal_v3.")
		return http.StatusInternalServerError
	}

	enProperties, properties, err := getDealPropertiesV3(projectID, document)
	if err != nil {
		logCtx.Error("Failed to get hubspot deal_v3 document properties")
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
		logCtx.Error("Failed to update hubspot deal_v3 document as synced.")
		return http.StatusInternalServerError
	}

	return http.StatusOK
}

var keyArrEngagementMeeting = []string{"id", "timestamp", "type", "source", "active"}
var keyArrMetaMeeting = []string{"startTime", "endTime", "title", "meetingOutcome"}
var keyArrEngagementCall = []string{"id", "timestamp", "type", "source", "activityType"}
var keyArrMetaCall = []string{"durationMilliseconds", "disposition", "status", "title", "disposition_label"}
var keyArrEngagementEmail = []string{"id", "createdAt", "lastUpdated", "type", "teamId", "ownerId", "active", "timestamp", "source"}
var keyArrMetaEmail = []string{"from", "to", "subject", "sentVia"}

const (
	EngagementTypeCall           = "CALL"
	EngagementTypeEmail          = "EMAIL"
	EngagementTypeIncomingEmail  = "INCOMING_EMAIL"
	EngagementTypeForwardedEmail = "FORWARDED_EMAIL"
	EngagementTypeMeeting        = "MEETING"

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
			vfloat, _ := util.GetPropertyValueAsFloat64(engagement.Metadata[key])
			properties[key] = (int64)(vfloat / 1000)
		} else if key == "to" {
			toInterface := engagement.Metadata[key]
			if toInterface == nil {
				logCtx.Error("No to in engagement metadata")
				continue
			}

			interfaceArray, isConvert := toInterface.([]interface{})
			if !isConvert {
				logCtx.Error("cannot convert interface to interface array")
				continue
			}

			if len(interfaceArray) == 0 {
				logCtx.Warn("Length of interface array is zero")
				continue
			}

			toMap, isConvert := interfaceArray[0].(map[string]interface{})
			if !isConvert {
				logCtx.Error("cannot convert interface to toMap")
				continue
			}
			properties[key] = toMap["email"]
		} else if key == "from" {
			fromInterface := engagement.Metadata[key]
			if fromInterface == nil {
				logCtx.Error("No from in engagement metadata")
				continue
			}

			fromMap, isConvert := fromInterface.(map[string]interface{})
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
			if contactIdInterface == nil {
				logCtx.Warning("No from for INCOMING_EMAIL engagement. Will be marked as synced.")
				return nil, http.StatusOK
			}

			fromMap, isConvert := contactIdInterface.(map[string]interface{})
			if !isConvert {
				logCtx.Error("cannot convert interface to fromMap")
				return contactIds, http.StatusInternalServerError
			}
			contact = fromMap["contactId"]
			if contact == "" || contact == nil {
				logCtx.Warning("No contact id for INCOMING_EMAIL engamement. Will be marked as synced")
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

func checkIfEngagementV3(document *model.HubspotDocument) (bool, error) {
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

func syncEngagements(project *model.Project, otpRules *[]model.OTPRule, uniqueOTPEventKeys *[]string, document *model.HubspotDocument) int {
	isEngagementV3, err := checkIfEngagementV3(document)
	if err != nil {
		log.WithFields(log.Fields{"project_id": project.ID}).WithError(err).Error("failed to check type of engagement record")
		return http.StatusInternalServerError
	}

	if isEngagementV3 {
		return syncEngagementsV3(project, otpRules, uniqueOTPEventKeys, document)
	}

	return syncEngagementsV2(project, otpRules, uniqueOTPEventKeys, document)
}

func syncEngagementsV2(project *model.Project, otpRules *[]model.OTPRule, uniqueOTPEventKeys *[]string, document *model.HubspotDocument) int {
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

		if !C.IsProjectIDSkippedForOtp(project.ID) {
			threadID, err := getThreadIDFromEngagement(engagement, engagementTypeStr)
			if err != nil {
				logCtx.Warn("couldn't get the threadID on hubspot email engagement, logging and continuing")
			}
			err = ApplyHSOfflineTouchPointRuleForEngagement(project, otpRules, uniqueOTPEventKeys, payload, document, threadID, engagementTypeStr)
			if err != nil {
				// log and continue
				logCtx.WithField("TrackPayload", payload).WithField("userID", userId).Info("failed " +
					"creating engagement hubspot offline touch point")
			}
		}

	}
	errCode := store.GetStore().UpdateHubspotDocumentAsSynced(project.ID, document.ID, model.HubspotDocumentTypeEngagement, "", document.Timestamp, document.Action, "", "")
	if errCode != http.StatusAccepted {
		logCtx.Error("Failed to update hubspot engagement document as synced.")
		return http.StatusInternalServerError
	}
	return http.StatusOK
}

var engagementMeetingV3PropertiesMap = map[string]string{
	"id":                    "id",
	"hs_timestamp":          "timestamp",
	"type":                  "type",
	"hs_meeting_source":     "source",
	"hs_meeting_active":     "active",
	"hs_meeting_start_time": "startTime",
	"hs_meeting_end_time":   "endTime",
	"hs_meeting_title":      "title",
	"hs_meeting_outcome":    "meetingOutcome",
}
var engagementCallV3PropertiesMap = map[string]string{
	"id":                        "id",
	"hs_timestamp":              "timestamp",
	"type":                      "type",
	"hs_call_source":            "source",
	"hs_activity_type":          "activityType",
	"hs_call_duration":          "durationMilliseconds",
	"hs_call_disposition":       "disposition",
	"hs_call_status":            "status",
	"hs_call_title":             "title",
	"hs_call_disposition_label": "disposition_label",
}
var engagementEmailV3PropertiesMap = map[string]string{
	"id":                  "id",
	"hs_createdate":       "createdAt",
	"hs_lastmodifieddate": "lastUpdated",
	"type":                "type",
	"hs_email_team_id":    "teamId",
	"hubspot_owner_id":    "ownerId",
	"hs_email_active":     "active",
	"hs_timestamp":        "timestamp",
	"hs_email_source":     "source",
	"hs_email_subject":    "subject",
	"hs_email_sent_via":   "sentVia",
}
var engagementEmailV3Headers = []string{"from", "to"}

func extractionOfPropertiesWithOutEmailOrContactV3(engagement EngagementsV3, engagementType string) map[string]interface{} {
	logCtx := log.WithField("engagement_type", engagementType).WithField("engagement", engagement)
	properties := make(map[string]interface{})
	var engagementArray map[string]string
	emailHeadersArray := make([]string, 0)

	if engagementType == EngagementTypeMeeting {
		engagementArray = engagementMeetingV3PropertiesMap
	} else if engagementType == EngagementTypeCall {
		engagementArray = engagementCallV3PropertiesMap
	} else if engagementType == EngagementTypeIncomingEmail || engagementType == EngagementTypeEmail {
		engagementArray = engagementEmailV3PropertiesMap
		emailHeadersArray = engagementEmailV3Headers
	}

	for key, oldKey := range engagementArray {
		if key == "id" {
			properties[oldKey] = engagement.Id
		} else if key == "hs_timestamp" || key == "hs_meeting_start_time" || key == "hs_meeting_end_time" {
			timestamp, err := model.GetTimestampForV3Records(engagement.Properties[key])
			if err != nil {
				logCtx.WithField(key, engagement.Properties[key]).Error("failed to get timestamp in engagementArray on extractionOfPropertiesWithOutEmailOrContactV3")
				continue
			}
			properties[oldKey] = (int64)(timestamp / 1000)
		} else if key == "hs_createdate" || key == "hs_lastmodifieddate" {
			timestamp, err := model.GetTimestampForV3Records(engagement.Properties[key])
			if err != nil {
				logCtx.WithField(key, engagement.Properties[key]).Error("failed to get timestamp in engagementArray on extractionOfPropertiesWithOutEmailOrContactV3")
				continue
			}
			properties[oldKey] = timestamp
		} else if key == "hs_call_duration" {
			durationInFloat, err := U.GetPropertyValueAsFloat64(engagement.Properties[key])
			if err != nil {
				logCtx.WithField("hs_call_duration", engagement.Properties[key]).Error("failed to get call duration in engagementArray on extractionOfPropertiesWithOutEmailOrContactV3")
				continue
			}
			properties[oldKey] = int64(durationInFloat)
		} else {
			properties[oldKey] = engagement.Properties[key]
		}
	}

	for _, key := range emailHeadersArray {
		if key == "to" {
			emailHeaders, ok := engagement.Properties["hs_email_headers"].(map[string]interface{})
			if !ok {
				logCtx.Error("failed to convert hs_email_header from interface to map")
				continue
			}

			toInterface := emailHeaders[key]
			if toInterface == nil {
				logCtx.Error("No to in engagement properties hs_email_header")
				continue
			}

			interfaceArray, isConvert := toInterface.([]interface{})
			if !isConvert {
				logCtx.Error("cannot convert interface to interface array")
				continue
			}

			if len(interfaceArray) == 0 {
				logCtx.Warn("Length of interface array is zero")
				continue
			}

			toMap, isConvert := interfaceArray[0].(map[string]interface{})
			if !isConvert {
				logCtx.Error("cannot convert interface to toMap")
				continue
			}
			properties[key] = toMap["email"]
		} else if key == "from" {
			emailHeaders, ok := engagement.Properties["hs_email_headers"].(map[string]interface{})
			if !ok {
				logCtx.Error("failed to convert hs_email_header from interface to map")
				continue
			}

			fromInterface := emailHeaders[key]
			if fromInterface == nil {
				logCtx.Error("No from in engagement properties hs_email_header")
				continue
			}

			fromMap, isConvert := fromInterface.(map[string]interface{})
			if !isConvert {
				logCtx.Error("cannot convert interface to fromMap")
				continue
			}
			properties[key] = fromMap["email"]
		} else {
			properties[key] = engagement.Properties[key]
		}
	}
	return properties
}

func getEngagementContactIdsV3(engagementTypeStr string, engagement EngagementsV3) ([]string, int) {
	logCtx := log.WithField("engagement_type_str", engagementTypeStr).WithField("engagement", engagement)
	contactIds := make([]string, 0)
	if engagementTypeStr == EngagementTypeCall || engagementTypeStr == EngagementTypeMeeting {
		contactIdArr := engagement.Associations["contactIds"]
		for i := range contactIdArr {
			contactId, err := U.GetPropertyValueAsFloat64(contactIdArr[i])
			if err != nil {
				logCtx.WithError(err).Error("cannot convert interface to float64")
				return contactIds, http.StatusInternalServerError
			}
			contactIds = append(contactIds, strconv.FormatInt((int64)(contactId), 10))
		}
	} else if engagementTypeStr == EngagementTypeIncomingEmail || engagementTypeStr == EngagementTypeEmail {
		var contactId float64
		var err error
		var contact interface{}

		if engagement.Properties["hs_email_headers"] == nil {
			return nil, http.StatusOK
		}

		emailHeaders, ok := engagement.Properties["hs_email_headers"].(map[string]interface{})
		if !ok {
			logCtx.Error("Failed to convert interface to map for email_headers")
			return nil, http.StatusInternalServerError
		}

		if engagementTypeStr == EngagementTypeIncomingEmail {
			contactIdInterface := emailHeaders["from"]
			if contactIdInterface == nil {
				logCtx.Warning("No from for INCOMING_EMAIL engagement. Will be marked as synced.")
				return nil, http.StatusOK
			}

			fromMap, isConvert := contactIdInterface.(map[string]interface{})
			if !isConvert {
				logCtx.Error("cannot convert interface to fromMap")
				return contactIds, http.StatusInternalServerError
			}
			contact = fromMap["contactId"]
			if contact == "" || contact == nil {
				logCtx.Warning("No contact id for INCOMING_EMAIL engamement. Will be marked as synced")
				return nil, http.StatusOK
			}
		} else {
			contactIdInterface := emailHeaders["to"]
			if contactIdInterface == nil {
				logCtx.Warning("No to for EMAIL engagement. Will be marked as synced.")
				return nil, http.StatusOK
			}

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
				logCtx.Warning("No contact id for EMAIL engamement_v3. Will be marked as synced")
				return nil, http.StatusOK
			}
		}

		contactId, err = U.GetPropertyValueAsFloat64(contact)
		if err != nil {
			logCtx.WithError(err).Error("cannot convert interface to float64 in getEngagementContactIdsV3")
			return contactIds, http.StatusInternalServerError
		}
		contactIds = append(contactIds, strconv.FormatInt((int64)(contactId), 10))
	}

	return contactIds, http.StatusOK
}

func syncEngagementsV3(project *model.Project, otpRules *[]model.OTPRule, uniqueOTPEventKeys *[]string, document *model.HubspotDocument) int {
	logCtx := log.WithField("project_id", project.ID).WithField("document_id", document.ID)
	if document.Type != model.HubspotDocumentTypeEngagement {
		logCtx.Error("It is not a type of engagement_v3")
		return http.StatusInternalServerError
	}

	var engagement EngagementsV3
	err := json.Unmarshal((document.Value).RawMessage, &engagement)
	if err != nil {
		logCtx.WithError(err).Error("Failed to unmarshal hubspot engagement_v3 document.")
		return http.StatusInternalServerError
	}

	engagementTypeInt, isPresent := engagement.Properties["type"]
	if !isPresent {
		logCtx.Error("Failed to find type as a key in engagement_v3.")
		return http.StatusInternalServerError
	}
	engagementType := U.GetPropertyValueAsString(engagementTypeInt)

	if engagementType != EngagementTypeCall && engagementType != EngagementTypeMeeting && engagementType != EngagementTypeIncomingEmail && engagementType != EngagementTypeEmail && engagementType != EngagementTypeForwardedEmail {
		logCtx.WithField("engagement_type", engagementTypeInt).Error("Invalid engagement_v3 type")
		return http.StatusInternalServerError
	}

	if engagementType == EngagementTypeForwardedEmail {
		errCode := store.GetStore().UpdateHubspotDocumentAsSynced(project.ID, document.ID, model.HubspotDocumentTypeEngagement, "", document.Timestamp, document.Action, "", "")
		if errCode != http.StatusAccepted {
			logCtx.Error("Failed to update hubspot engagement_v3 document as synced.")
			return http.StatusInternalServerError
		}
		return http.StatusOK
	}

	if (engagementType == EngagementTypeIncomingEmail || engagementType == EngagementTypeEmail) && document.Action == model.HubspotDocumentActionUpdated {
		errCode := store.GetStore().UpdateHubspotDocumentAsSynced(project.ID, document.ID, model.HubspotDocumentTypeEngagement, "", document.Timestamp, document.Action, "", "")
		if errCode != http.StatusAccepted {
			logCtx.Error("Failed to update hubspot engagement_v3 document as synced.")
			return http.StatusInternalServerError
		}
		return http.StatusOK
	}

	contactIds, error := getEngagementContactIdsV3(engagementType, engagement)
	if error != http.StatusOK {
		logCtx.Error("failed to get the contact id in engagement_v3")
		return error
	}

	if len(contactIds) == 0 {
		errCode := store.GetStore().UpdateHubspotDocumentAsSynced(project.ID, document.ID, model.HubspotDocumentTypeEngagement, "", document.Timestamp, document.Action, "", "")
		if errCode != http.StatusAccepted {
			logCtx.Error("Failed to update hubspot engagement_v3 document as synced.")
			return http.StatusInternalServerError
		}
		return http.StatusOK
	}

	properties := extractionOfPropertiesWithOutEmailOrContactV3(engagement, engagementType)
	contactEngagementProperties := make(map[string]map[string]interface{})

	var contactDocuments []model.HubspotDocument
	var status int

	if len(contactIds) > 0 {
		contactDocuments, status = store.GetStore().GetHubspotDocumentByTypeAndActions(project.ID, contactIds, model.HubspotDocumentTypeContact, []int{model.HubspotDocumentActionCreated, model.HubspotDocumentActionUpdated})
		if status != http.StatusFound {
			if status != http.StatusNotFound {
				logCtx.Error(
					"Failed to get hubspot documents by type and action on sync engagement_v3.")
				return http.StatusInternalServerError
			}
			logCtx.Warning("Missing engagement_v3 associated contact record.")
			// Avoid returning error if associated record is not present.
			return http.StatusOK
		}
	}

	if C.DisableHubspotNonMarketingContactsByProjectID(project.ID) && len(contactDocuments) == 0 {
		logCtx.Warning("No marketing contacts found for hubspot engagement_v3.")
	}

	for i := range contactIds {
		var latestContactDocument *model.HubspotDocument
		for j := range contactDocuments {
			if contactIds[i] != contactDocuments[j].ID {
				continue
			}

			// pick the latest contact document before the engagment timestamp or the first contact document.
			if latestContactDocument == nil || latestContactDocument.Timestamp < contactDocuments[j].Timestamp {
				latestContactDocument = &contactDocuments[j]
			}
		}

		if latestContactDocument == nil {
			logCtx.WithFields(log.Fields{"contact_id": contactIds[i]}).Warning("Missing contact record for activity in engagement_v3.")
			continue
		}

		propertiesWithEmailOrContact := make(map[string]interface{})
		enProperties, _, _, _, err := GetContactProperties(project.ID, latestContactDocument)
		if err != nil {
			logCtx.WithError(err).Error("can't get contact properties in engagement_v3")
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
		logCtx.Warn("No contacts for processing in engagement_v3.")
		return http.StatusInternalServerError
	}

	eventName := getEventNameByDocumentTypeAndAction(engagementType, document.Action)
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
			logCtx.Error("Failed to create hubspot engagement_v3 event")
			return http.StatusInternalServerError
		}

		if !C.IsProjectIDSkippedForOtp(project.ID) {
			threadID, err := getThreadIDFromEngagement(engagement, engagementType)
			if err != nil {
				logCtx.Error("couldn't get the threadID on hubspot email engagement_v3, logging and continuing")
			}
			err = ApplyHSOfflineTouchPointRuleForEngagement(project, otpRules, uniqueOTPEventKeys, payload, document, threadID, engagementType)
			if err != nil {
				// log and continue
				logCtx.WithField("TrackPayload", payload).WithField("userID", userId).Info("failed " +
					"creating engagement_v3 hubspot offline touch point")
			}
		}

	}
	errCode := store.GetStore().UpdateHubspotDocumentAsSynced(project.ID, document.ID, model.HubspotDocumentTypeEngagement, "", document.Timestamp, document.Action, "", "")
	if errCode != http.StatusAccepted {
		logCtx.Error("Failed to update hubspot engagement_v3 document as synced.")
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

func syncContactListV2(project *model.Project, otpRules *[]model.OTPRule, uniqueOTPEventKeys *[]string, document *model.HubspotDocument, minTimestamp int64) int {
	logCtx := log.WithFields(log.Fields{"project_id": project.ID, "document_id": document.ID,
		"doc_timestamp": document.Timestamp, "min_timestamp": minTimestamp})

	if document.Type != model.HubspotDocumentTypeContactList {
		logCtx.Error("Invalid contact_list document.")
		return http.StatusInternalServerError
	}

	pastEnrichmentEnabled := false
	if C.PastEventEnrichmentEnabled(project.ID) {
		pastEnrichmentEnabled = true
	}

	if !C.ContactListInsertEnabled(project.ID) {
		logCtx.Warning("Skipped hubspot contact_list sync. contact_list sync not enabled.")
		return http.StatusOK
	}

	if document.Action == model.HubspotDocumentActionUpdated {
		errCode := store.GetStore().UpdateHubspotDocumentAsSynced(
			project.ID, document.ID, model.HubspotDocumentTypeContactList, "", document.Timestamp, document.Action, document.UserId, "")
		if errCode != http.StatusAccepted {
			logCtx.Error("Failed to update hubspot contact_list document as synced.")
			return http.StatusInternalServerError
		}

		return http.StatusOK
	}

	propertiesMap, err := U.DecodePostgresJsonb(document.Value)
	if err != nil {
		logCtx.WithError(err).Error("Failed to decode hubspot Json document-value into PropertiesMap in syncContactListV2.")
		return http.StatusInternalServerError
	}

	isPast := false
	if pastEnrichmentEnabled {
		isPast = document.Timestamp < minTimestamp
	}

	contactID := U.GetPropertyValueAsString((*propertiesMap)["contact_id"])
	contact_document, errCode := store.GetStore().GetSyncedHubspotDocumentByFilter(project.ID, contactID, model.HubspotDocumentTypeContact, model.HubspotDocumentActionCreated)
	if errCode != http.StatusFound {
		logCtx.Error("Failed to get contact document in syncContactListV2")
		if errCode == http.StatusNotFound {
			return http.StatusOK
		}
		return errCode
	}

	_, properties, _, _, _ := GetContactProperties(project.ID, contact_document)
	var emailId string
	if (*properties)["EMAIL"] != nil {
		emailId, err = U.GetValueAsString((*properties)["EMAIL"])
		if err != nil {
			logCtx.Error("Failed to get emailId from contact properties.")
		}
	} else {
		logCtx.Error("No emailId in contact properties.")
	}

	propertyToValueMap := map[string]interface{}{
		"list_name":             (*propertiesMap)["name"],
		"list_id":               (*propertiesMap)["listId"],
		"list_type":             (*propertiesMap)["listType"],
		"contact_id":            contact_document.ID,
		"contact_email":         emailId,
		"list_create_timestamp": getEventTimestamp(U.GetInt64FromMapOfInterface(*propertiesMap, "createdAt", 0)),
		"timestamp":             getEventTimestamp(document.Timestamp),
	}

	eventProperties := make(map[string]interface{})

	for property, value := range propertyToValueMap {
		key := model.GetCRMEnrichPropertyKeyByType(model.SmartCRMEventSourceHubspot, model.HubspotDocumentTypeNameContactList, property)
		eventProperties[key] = value
	}

	request := &SDK.TrackPayload{
		ProjectId:       project.ID,
		Timestamp:       getEventTimestamp(document.Timestamp),
		EventProperties: eventProperties,
		RequestSource:   model.UserSourceHubspot,
		Name:            U.EVENT_NAME_HUBSPOT_CONTACT_LIST,
		UserId:          contact_document.UserId,
		IsPast:          isPast,
	}

	status, response := SDK.Track(project.ID, request, true, SDK.SourceHubspot, model.HubspotDocumentTypeNameContactList)
	if status != http.StatusOK && status != http.StatusFound && status != http.StatusNotModified {
		logCtx.WithFields(log.Fields{"status": status, "track_response": response}).Error("Failed to track hubspot added to a list event.")
		return http.StatusInternalServerError
	}

	if !C.IsProjectIDSkippedForOtp(project.ID) {
		err = ApplyHSOfflineTouchPointRuleForContactList(project, otpRules, uniqueOTPEventKeys, request, document)
		if err != nil {
			// log and continue
			logCtx.WithField("EventID", response.EventId).WithField("userID", response.UserId).Info("failed creating hubspot offline touch point for contact list")
		}
	}
	errCode = store.GetStore().UpdateHubspotDocumentAsSynced(
		project.ID, document.ID, model.HubspotDocumentTypeContactList, "", document.Timestamp, document.Action, contact_document.UserId, "")
	if errCode != http.StatusAccepted {
		logCtx.Error("Failed to update hubspot contact_list document as synced.")
		return http.StatusInternalServerError
	}

	return http.StatusOK
}

func syncAll(project *model.Project, otpRules *[]model.OTPRule, uniqueOTPEventKeys *[]string, documents []model.HubspotDocument, hubspotSmartEventNames []HubspotSmartEventName, minTimestamp int64) int {
	logCtx := log.WithField("project_id", project.ID)
	var seenFailures bool
	for i := range documents {
		logCtx = logCtx.WithFields(log.Fields{"document_id": documents[i].ID, "doc_type": documents[i].Type, "document_timestamp": documents[i].Timestamp})
		startTime := time.Now().Unix()

		switch documents[i].Type {

		case model.HubspotDocumentTypeContact:
			errCode := syncContact(project, otpRules, uniqueOTPEventKeys, &documents[i], hubspotSmartEventNames)
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
			errCode := syncEngagements(project, otpRules, uniqueOTPEventKeys, &documents[i])
			if errCode != http.StatusOK {
				seenFailures = true
			}
		case model.HubspotDocumentTypeContactList:
			errCode := syncContactListV2(project, otpRules, uniqueOTPEventKeys, &documents[i], minTimestamp)
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
func syncAllWorker(project *model.Project, wg *sync.WaitGroup, syncStatus *syncWorkerStatus, otpRules *[]model.OTPRule, uniqueOTPEventKeys *[]string, documents []model.HubspotDocument, hubspotSmartEventNames []HubspotSmartEventName, minTimestamp int64) {
	defer wg.Done()

	errCode := syncAll(project, otpRules, uniqueOTPEventKeys, documents, hubspotSmartEventNames, minTimestamp)

	syncStatus.Lock.Lock()
	defer syncStatus.Lock.Unlock()
	if errCode != http.StatusOK {
		syncStatus.HasFailure = true
	}
}

func syncByOrderedTimeSeries(project *model.Project, otpRules *[]model.OTPRule, uniqueOTPEventKeys *[]string, orderedTimeSeries [][]int64, workersPerProject int, recordsMaxCreatedAtSec int64, datePropertiesByObjectType map[int]*map[string]bool, timeZone U.TimeZoneString, recordsProcessLimit int,
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

			if errCode != http.StatusFound && errCode != http.StatusNotFound {
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
					go syncAllWorker(project, &wg, &syncStatus, otpRules, uniqueOTPEventKeys, batch[docID], (*hubspotSmartEventNames)[docTypeAlias], minTimestamps[syncOrderByType[i]])
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

	otpRules, errCode := store.GetStore().GetALLOTPRuleWithProjectId(projectID)
	if errCode != http.StatusFound && errCode != http.StatusNotFound {
		logCtx.WithField("err_code", errCode).Error("Failed to get otp Rules for Project")
		statusByProjectAndType = append(statusByProjectAndType, Status{ProjectId: projectID,
			Status: "Failed to get OTP rules"})
		return statusByProjectAndType, true
	}

	uniqueOTPEventKeys, errCode := store.GetStore().GetUniqueKeyPropertyForOTPEventForLast3Months(projectID)
	if errCode != http.StatusFound && errCode != http.StatusNotFound {
		logCtx.WithField("err_code", errCode).Error("Failed to get OTP Unique Keys for Project")
		statusByProjectAndType = append(statusByProjectAndType, Status{ProjectId: projectID,
			Status: "Failed to get OTP Unique Keys"})
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
	overAllSyncStatus, overallExecutionTime, overallProcessedCount, isProcessLimitExceeded := syncByOrderedTimeSeries(project, &otpRules, &uniqueOTPEventKeys, orderedTimeSeries, workersPerProject,
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
