package salesforce

import (
	"encoding/json"
	"errors"
	M "factors/model"
	"fmt"
	"net/http"
	"strings"
	"time"

	C "factors/config"
	SDK "factors/sdk"
	U "factors/util"

	log "github.com/sirupsen/logrus"
)

// Status represents current sync status for a doc type
type Status struct {
	ProjectID uint64 `json:"project_id"`
	Type      string `json:"type"`
	Status    string `json:"status"`
}

var possiblePhoneField = []string{
	"mobilephone",
	"mobilephone__c",
	"phone",
	"phone__c",
	"mobile__c",
	"personmobilephone",
}

var salesforceSyncOrderByType = [...]int{
	M.SalesforceDocumentTypeContact,
	M.SalesforceDocumentTypeAccount,
	M.SalesforceDocumentTypeLead,
	M.SalesforceDocumentTypeOpportunity,
}

func getUserIDFromLastestProperties(properties []M.UserProperties) string {
	latestIndex := len(properties) - 1
	return properties[latestIndex].UserId
}

// GetSalesforceDocumentProperties return map of enriched properties
func GetSalesforceDocumentProperties(projectID uint64, document *M.SalesforceDocument) (*map[string]interface{}, error) {
	var properties map[string]interface{}
	err := json.Unmarshal(document.Value.RawMessage, &properties)
	if err != nil {
		return nil, err
	}

	filterPropertyFieldsByProjectID(projectID, &properties, document.Type)

	enrichedProperties := make(map[string]interface{})

	for key, value := range properties {
		enKey := getPropertyKeyByType(M.GetSalesforceAliasByDocType(document.Type), key)
		if _, exists := properties[enKey]; !exists {
			enrichedProperties[enKey] = value
		}
	}

	return &enrichedProperties, nil
}

func filterPropertyFieldsByProjectID(projectID uint64, properties *map[string]interface{}, docType int) {

	if projectID == 0 {
		return
	}

	allowedfields := M.GetSalesforceAllowedfiedsByObject(projectID, M.GetSalesforceAliasByDocType(docType))
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
	delete(*properties, "attributes") // delte nested meta object
}

func getSalesforceAccountID(document *M.SalesforceDocument) (string, error) {
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

func getCustomerUserIDFromProperties(projectID uint64, properties map[string]interface{}, docTypeAlias string) (string, string) {

	for _, phoneField := range possiblePhoneField {
		if phoneNo, ok := properties[getPropertyKeyByType(docTypeAlias, phoneField)]; ok {
			phoneStr, err := U.GetValueAsString(phoneNo)
			if err != nil || phoneStr == "" {
				continue
			}

			return M.GetUserIdentificationPhoneNumber(projectID, phoneStr)
		}
	}

	possibleEmailField := []string{
		"Email",
		"Email__c",
		"PersonEmail",
	}

	for _, emailField := range possibleEmailField {
		if email, ok := properties[getPropertyKeyByType(docTypeAlias, emailField)].(string); ok && email != "" {
			existingEmail, errCode := M.GetExistingCustomerUserID(projectID, []string{email})
			if errCode == http.StatusFound {
				return email, existingEmail[email]
			}

			return email, ""
		}
	}

	return "", ""
}

func getPropertyKeyByType(typ, key string) string {
	return fmt.Sprintf("$%s_%s_%s", SDK.SourceSalesforce, typ, strings.ToLower(key))
}

/*
TrackSalesforceEventByDocumentType tracks salesforce events by action
	for action created -> create both created and updated events with date created timestamp
	for action updated -> create on updated event with lastmodified timestamp
*/
func TrackSalesforceEventByDocumentType(projectID uint64, trackPayload *SDK.TrackPayload, document *M.SalesforceDocument) (string, string, error) {
	if projectID == 0 {
		return "", "", errors.New("invalid project id")
	}

	if trackPayload == nil || document == nil {
		return "", "", errors.New("invalid operation")
	}

	createdTimestamp, err := M.GetSalesforceDocumentTimestampByAction(document, M.SalesforceDocumentCreated)
	if err != nil {
		return "", "", err
	}

	lastModifiedTimestamp, err := M.GetSalesforceDocumentTimestampByAction(document, M.SalesforceDocumentUpdated)
	if err != nil {
		return "", "", err
	}

	var eventID, userID string
	if document.Action == M.SalesforceDocumentCreated {
		payload := *trackPayload
		payload.Name = M.GetSalesforceEventNameByAction(document, M.SalesforceDocumentCreated)
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

	if document.Action == M.SalesforceDocumentCreated || document.Action == M.SalesforceDocumentUpdated {
		payload := *trackPayload
		payload.Name = M.GetSalesforceEventNameByAction(document, M.SalesforceDocumentUpdated)

		if document.Action == M.SalesforceDocumentUpdated {
			payload.Timestamp = lastModifiedTimestamp
			// TODO(maisa): Use GetSyncedSalesforceDocumentByType while updating multiple contacts in an account object
			documents, status := M.GetSyncedSalesforceDocumentByType(projectID, []string{document.ID}, document.Type)
			if status != http.StatusFound {
				return "", "", errors.New("failed to get synced document")
			}

			event, status := M.GetEventById(projectID, documents[0].SyncID)
			if status != http.StatusFound {
				return "", "", errors.New("failed to get event from sync id ")
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
	if document.Action == M.SalesforceDocumentCreated && createdTimestamp != lastModifiedTimestamp {
		payload := *trackPayload
		payload.Timestamp = lastModifiedTimestamp
		payload.UserId = userID
		payload.Name = M.GetSalesforceEventNameByAction(document, M.SalesforceDocumentUpdated)
		status, _ := SDK.Track(projectID, &payload, true, SDK.SourceSalesforce)
		if status != http.StatusOK && status != http.StatusFound && status != http.StatusNotModified {
			return "", "", fmt.Errorf("updated event for different timestamp track failed for doc type %d", document.Type)
		}
	}

	return eventID, userID, nil
}

func enrichAccount(projectID uint64, document *M.SalesforceDocument, salesforceSmartEventNames []SalesforceSmartEventName) int {
	if projectID == 0 || document == nil {
		return http.StatusBadRequest
	}

	if document.Type != M.SalesforceDocumentTypeAccount {
		return http.StatusInternalServerError
	}

	logCtx := log.WithField("project_id", projectID).WithField("document_id", document.ID)

	properties, err := GetSalesforceDocumentProperties(projectID, document)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get properties")
	}

	trackPayload := &SDK.TrackPayload{
		ProjectId:       projectID,
		EventProperties: *properties,
		UserProperties:  *properties,
	}

	eventID, userID, err := TrackSalesforceEventByDocumentType(projectID, trackPayload, document)
	if err != nil {
		logCtx.WithError(err).Error(
			"Failed to track salesforce account event.")
		return http.StatusInternalServerError
	}

	customerUserID, _ := getCustomerUserIDFromProperties(projectID, *properties, M.GetSalesforceAliasByDocType(document.Type))
	if customerUserID != "" {
		status, _ := SDK.Identify(projectID, &SDK.IdentifyPayload{
			UserId: userID, CustomerUserId: customerUserID}, false)
		if status != http.StatusOK {
			logCtx.WithField("customer_user_id", customerUserID).Error(
				"Failed to identify user on salesforce account sync.")
			return http.StatusInternalServerError
		}
	} else {
		logCtx.Error("Skipped user identification on salesforce account sync. No customer_user_id on properties.")
	}

	var prevProperties *map[string]interface{}
	for _, smartEventName := range salesforceSmartEventNames {
		prevProperties = TrackSalesforceSmartEvent(projectID, &smartEventName, eventID, customerUserID, userID, document.Type, properties, prevProperties)
	}

	errCode := M.UpdateSalesforceDocumentAsSynced(projectID, document, eventID, userID)
	if errCode != http.StatusAccepted {
		logCtx.Error("Failed to update salesforce account document as synced.")
		return http.StatusInternalServerError
	}

	return http.StatusOK
}

// SalesforceSmartEventName struct for holding event_name and filter expression
type SalesforceSmartEventName struct {
	EventName string
	Filter    *M.SmartCRMEventFilter
	Type      string
}

func getTimestampFromField(propertyName string, properties *map[string]interface{}) int64 {
	if timestamp, exists := (*properties)[propertyName]; exists {
		if unixTimestamp, ok := timestamp.(int64); ok {
			return unixTimestamp
		}

		unixTimestamp, err := M.GetSalesforceDocumentTimestamp(timestamp)
		if err != nil {
			return 0
		}

		return unixTimestamp
	}

	return 0
}

func enrichContact(projectID uint64, document *M.SalesforceDocument, salesforceSmartEventNames []SalesforceSmartEventName) int {
	if projectID == 0 || document == nil {
		return http.StatusBadRequest
	}

	if document.Type != M.SalesforceDocumentTypeContact {
		return http.StatusInternalServerError
	}

	logCtx := log.WithField("project_id", projectID).WithField("document_id", document.ID)
	properties, err := GetSalesforceDocumentProperties(projectID, document)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get properties")
	}

	trackPayload := &SDK.TrackPayload{
		ProjectId:       projectID,
		EventProperties: *properties,
		UserProperties:  *properties,
	}

	eventID, userID, err := TrackSalesforceEventByDocumentType(projectID, trackPayload, document)
	if err != nil {
		logCtx.WithError(err).Error(
			"Failed to track salesforce contact event.")
		return http.StatusInternalServerError
	}

	customerUserID, _ := getCustomerUserIDFromProperties(projectID, *properties, M.GetSalesforceAliasByDocType(document.Type))
	if customerUserID != "" {
		status, _ := SDK.Identify(projectID, &SDK.IdentifyPayload{
			UserId: userID, CustomerUserId: customerUserID}, false)
		if status != http.StatusOK {
			logCtx.WithField("customer_user_id", customerUserID).Error(
				"Failed to identify user on salesforce contact sync.")
			return http.StatusInternalServerError
		}
	} else {
		logCtx.Error("Skipped user identification on salesforce contact sync. No customer_user_id on properties.")
	}

	var prevProperties *map[string]interface{}
	for _, smartEventName := range salesforceSmartEventNames {
		prevProperties = TrackSalesforceSmartEvent(projectID, &smartEventName, eventID, customerUserID, userID, document.Type, properties, prevProperties)
	}

	errCode := M.UpdateSalesforceDocumentAsSynced(projectID, document, eventID, userID)
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
	currentProperties, prevProperties *map[string]interface{}, filter *M.SmartCRMEventFilter) (*M.CRMSmartEvent, *map[string]interface{}, bool) {

	var crmSmartEvent M.CRMSmartEvent
	var validProperty bool
	var newProperties map[string]interface{}

	if projectID == 0 || eventName == "" || filter == nil || currentProperties == nil {
		return nil, prevProperties, false
	}

	if prevProperties == nil && (docType == 0 || userID == "") {
		return nil, prevProperties, false
	}

	if prevProperties != nil {
		validProperty = M.CRMFilterEvaluator(projectID, currentProperties, prevProperties, filter, M.CompareStateBoth)
	} else {
		validProperty = M.CRMFilterEvaluator(projectID, currentProperties, nil, filter, M.CompareStateCurr)
	}

	logCtx := log.WithFields(log.Fields{"project_id": projectID, "doc_type": docType})

	if !validProperty {
		return nil, prevProperties, false
	}

	if prevProperties == nil {
		prevDoc, status := M.GetLastSyncedSalesforceDocumentByCustomerUserIDORUserID(projectID, customerUserID, userID, docType)
		if status != http.StatusFound {
			return nil, prevProperties, false
		}

		var err error
		prevProperties, err = GetSalesforceDocumentProperties(projectID, prevDoc)
		if err != nil {
			logCtx.WithError(err).Error("Failed to GetSalesforceDocumentProperties")
			return nil, prevProperties, false
		}

		if !M.CRMFilterEvaluator(projectID, currentProperties, prevProperties,
			filter, M.CompareStateBoth) {
			return nil, prevProperties, false
		}
	}

	crmSmartEvent.Name = eventName
	M.FillSmartEventCRMProperties(&newProperties, currentProperties, prevProperties, filter)
	crmSmartEvent.Properties = newProperties

	return &crmSmartEvent, prevProperties, true
}

// TrackSalesforceSmartEvent valids current properties with CRM smart filter and creates a event
func TrackSalesforceSmartEvent(projectID uint64, salesforceSmartEventName *SalesforceSmartEventName, eventID, customerUserID, userID string, docType int, currentProperties, prevProperties *map[string]interface{}) *map[string]interface{} {
	var valid bool
	var smartEventPayload *M.CRMSmartEvent
	if salesforceSmartEventName.EventName == "" || projectID == 0 || salesforceSmartEventName.Type == "" {
		return prevProperties
	}

	if userID == "" || docType == 0 || currentProperties == nil || salesforceSmartEventName.Filter == nil {
		return prevProperties
	}

	logCtx := log.WithFields(log.Fields{"project_id": projectID, "doc_type": docType})
	smartEventPayload, prevProperties, valid = GetSalesforceSmartEventPayload(projectID, salesforceSmartEventName.EventName, customerUserID,
		userID, docType, currentProperties, prevProperties, salesforceSmartEventName.Filter)
	if !valid {
		return prevProperties
	}

	M.AddSmartEventReferenceMeta(&smartEventPayload.Properties, eventID)

	smartEventTrackPayload := &SDK.TrackPayload{
		ProjectId:       projectID,
		EventProperties: smartEventPayload.Properties,
		Name:            smartEventPayload.Name,
		SmartEventType:  salesforceSmartEventName.Type,
	}

	timestampReferenceField := salesforceSmartEventName.Filter.TimestampReferenceField
	if timestampReferenceField != M.TimestampReferenceTypeTrack {
		fieldTimestamp := getTimestampFromField(timestampReferenceField, currentProperties)
		if fieldTimestamp == 0 {
			logCtx.Errorf("Failed to get timestamp from reference field %s", timestampReferenceField)
			return prevProperties
		}
		smartEventTrackPayload.Timestamp = fieldTimestamp
	}

	if !C.IsDryRunCRMSmartEvent() {
		status, _ := SDK.Track(projectID, smartEventTrackPayload, true, SDK.SourceSalesforce)
		if status != http.StatusOK && status != http.StatusFound && status != http.StatusNotModified {
			logCtx.Error("Failed to create salesforce smart event")
		}
	} else {
		logCtx.WithFields(log.Fields{"properties": smartEventPayload.Properties, "event_name": smartEventPayload.Name,
			"filter_exp": *salesforceSmartEventName.Filter,
			"timestamp":  smartEventTrackPayload.Timestamp}).Info("Dry run smart event creation.")
	}

	return prevProperties
}

func enrichOpportunities(projectID uint64, document *M.SalesforceDocument, salesforceSmartEventNames []SalesforceSmartEventName) int {
	if projectID == 0 || document == nil {
		return http.StatusBadRequest
	}

	if document.Type != M.SalesforceDocumentTypeOpportunity {
		return http.StatusInternalServerError
	}

	logCtx := log.WithField("project_id", projectID).WithField("document_id", document.ID)
	properties, err := GetSalesforceDocumentProperties(projectID, document)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get properties")
	}

	trackPayload := &SDK.TrackPayload{
		ProjectId:       projectID,
		EventProperties: *properties,
		UserProperties:  *properties,
	}

	var eventID string
	customerUserID, userID := getCustomerUserIDFromProperties(projectID, *properties, M.GetSalesforceAliasByDocType(document.Type))
	if customerUserID != "" {
		trackPayload.UserId = userID
		eventID, _, err = TrackSalesforceEventByDocumentType(projectID, trackPayload, document)
		if err != nil {
			logCtx.WithError(err).Error(
				"Failed to track salesforce opportunity event.")
			return http.StatusInternalServerError
		}
	} else {
		eventID, _, err = TrackSalesforceEventByDocumentType(projectID, trackPayload, document)
		if err != nil {
			logCtx.WithError(err).Error(
				"Failed to track salesforce opportunity event.")
			return http.StatusInternalServerError
		}

		logCtx.Error("Skipped user identification on salesforce opportunity sync. No customer_user_id on properties.")
	}

	var prevProperties *map[string]interface{}
	for _, smartEventName := range salesforceSmartEventNames {
		prevProperties = TrackSalesforceSmartEvent(projectID, &smartEventName, eventID, customerUserID, userID,
			document.Type, properties, prevProperties)
	}

	errCode := M.UpdateSalesforceDocumentAsSynced(projectID, document, eventID, userID)
	if errCode != http.StatusAccepted {
		logCtx.Error("Failed to update salesforce opportunity document as synced.")
		return http.StatusInternalServerError
	}

	return http.StatusOK
}

func enrichLeads(projectID uint64, document *M.SalesforceDocument, salesforceSmartEventNames []SalesforceSmartEventName) int {
	if projectID == 0 || document == nil {
		return http.StatusBadRequest
	}

	if document.Type != M.SalesforceDocumentTypeLead {
		return http.StatusInternalServerError
	}

	logCtx := log.WithField("project_id", projectID).WithField("document_id", document.ID)

	properties, err := GetSalesforceDocumentProperties(projectID, document)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get properties")
	}

	trackPayload := &SDK.TrackPayload{
		ProjectId:       projectID,
		EventProperties: *properties,
		UserProperties:  *properties,
	}

	eventID, userID, err := TrackSalesforceEventByDocumentType(projectID, trackPayload, document)
	if err != nil {
		logCtx.WithError(err).Error(
			"Failed to track salesforce lead event.")
		return http.StatusInternalServerError
	}

	customerUserID, _ := getCustomerUserIDFromProperties(projectID, *properties, M.GetSalesforceAliasByDocType(document.Type))
	if customerUserID != "" {
		status, _ := SDK.Identify(projectID, &SDK.IdentifyPayload{
			UserId: userID, CustomerUserId: customerUserID}, false)
		if status != http.StatusOK {
			logCtx.WithField("customer_user_id", customerUserID).Error(
				"Failed to identify user on salesforce lead sync.")
			return http.StatusInternalServerError
		}
	} else {
		logCtx.Error("Skipped user identification on salesforce lead sync. No customer_user_id on properties.")
	}

	var prevProperties *map[string]interface{}
	for _, smartEventName := range salesforceSmartEventNames {
		prevProperties = TrackSalesforceSmartEvent(projectID, &smartEventName, eventID, customerUserID, userID, document.Type, properties, prevProperties)
	}

	errCode := M.UpdateSalesforceDocumentAsSynced(projectID, document, eventID, userID)
	if errCode != http.StatusAccepted {
		logCtx.Error("Failed to update salesforce lead document as synced.")
		return http.StatusInternalServerError
	}

	return http.StatusOK
}

func enrichAll(projectID uint64, documents []M.SalesforceDocument, salesforceSmartEventNames []SalesforceSmartEventName) int {
	if projectID == 0 {
		return http.StatusBadRequest
	}
	logCtx := log.WithField("project_id", projectID)

	var seenFailures bool
	var errCode int
	for i := range documents {
		startTime := time.Now().Unix()

		switch documents[i].Type {
		case M.SalesforceDocumentTypeAccount:
			errCode = enrichAccount(projectID, &documents[i], salesforceSmartEventNames)
		case M.SalesforceDocumentTypeContact:
			errCode = enrichContact(projectID, &documents[i], salesforceSmartEventNames)
		case M.SalesforceDocumentTypeLead:
			errCode = enrichLeads(projectID, &documents[i], salesforceSmartEventNames)
		case M.SalesforceDocumentTypeOpportunity:
			errCode = enrichOpportunities(projectID, &documents[i], salesforceSmartEventNames)
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

	eventNames, errCode := M.GetSmartEventFilterEventNames(projectID)
	if errCode == http.StatusInternalServerError {
		logCtx.Error("Error while GetSmartEventFilterEventNames")
	}

	salesforceSmartEventNames := make(map[string][]SalesforceSmartEventName)

	if len(eventNames) == 0 {
		return &salesforceSmartEventNames
	}

	for i := range eventNames {
		if eventNames[i].Type != M.TYPE_CRM_SALESFORCE {
			continue
		}

		var salesforceSmartEventName SalesforceSmartEventName
		decFilterExp, err := M.GetDecodedSmartEventFilterExp(eventNames[i].FilterExpr)
		if err != nil {
			logCtx.WithError(err).Error("Failed to decode smart event filter expression")
			continue
		}

		salesforceSmartEventName.EventName = eventNames[i].Name
		salesforceSmartEventName.Filter = decFilterExp
		salesforceSmartEventName.Type = M.TYPE_CRM_SALESFORCE

		if _, exists := salesforceSmartEventNames[decFilterExp.ObjectType]; !exists {
			salesforceSmartEventNames[decFilterExp.ObjectType] = []SalesforceSmartEventName{}
		}

		salesforceSmartEventNames[decFilterExp.ObjectType] = append(salesforceSmartEventNames[decFilterExp.ObjectType], salesforceSmartEventName)
	}

	return &salesforceSmartEventNames
}

// GetSalesforceDocumentsByTypeForSync pulls salesforce documents which are not synced
func GetSalesforceDocumentsByTypeForSync(projectID uint64, typ int) ([]M.SalesforceDocument, int) {
	logCtx := log.WithFields(log.Fields{"project_id": projectID, "type": typ})

	if projectID == 0 || typ == 0 {
		logCtx.Error("Invalid project_id or type on get salesforce documents by type.")
		return nil, http.StatusBadRequest
	}

	var documents []M.SalesforceDocument

	db := C.GetServices().Db
	err := db.Order("timestamp, created_at ASC").Where("project_id=? AND type=? AND synced=false",
		projectID, typ).Find(&documents).Error
	if err != nil {
		logCtx.WithError(err).Error("Failed to get salesforce documents by type.")
		return nil, http.StatusInternalServerError
	}

	return documents, http.StatusFound
}

// Enrich sync salesforce documents to events
func Enrich(projectID uint64) []Status {

	logCtx := log.WithField("project_id", projectID)

	statusByProjectAndType := make([]Status, 0, 0)
	if projectID == 0 {
		return statusByProjectAndType
	}

	allowedDocTypes := M.GetSalesforceDocumentTypeAlias(projectID)

	salesforceSmartEventNames := GetSalesforceSmartEventNames(projectID)

	for _, docType := range salesforceSyncOrderByType {
		docTypeAlias := M.GetSalesforceAliasByDocType(docType)
		if _, exist := allowedDocTypes[docTypeAlias]; !exist {
			continue
		}

		logCtx = logCtx.WithFields(log.Fields{
			"doc_type":   docType,
			"project_id": projectID,
		})

		documents, errCode := GetSalesforceDocumentsByTypeForSync(projectID, docType)
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
		}
		statusByProjectAndType = append(statusByProjectAndType, status)
	}

	return statusByProjectAndType
}
