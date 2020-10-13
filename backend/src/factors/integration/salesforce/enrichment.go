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

func getSalesforceDocumentProperties(projectID uint64, document *M.SalesforceDocument) (map[string]interface{}, error) {
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

	return enrichedProperties, nil
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

			pPhoneNo := U.GetPossiblePhoneNumber(phoneStr)
			existingPhoneNo, errCode := M.GetExistingCustomerUserID(projectID, pPhoneNo)
			if errCode == http.StatusFound {
				for i := range pPhoneNo {
					if userID, exist := existingPhoneNo[pPhoneNo[i]]; exist {
						return pPhoneNo[i], userID
					}
				}
			}

			return phoneStr, ""
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

func enrichAccount(projectID uint64, document *M.SalesforceDocument) int {
	if projectID == 0 || document == nil {
		return http.StatusBadRequest
	}

	if document.Type != M.SalesforceDocumentTypeAccount {
		return http.StatusInternalServerError
	}

	logCtx := log.WithField("project_id", projectID).WithField("document_id", document.ID)

	properties, err := getSalesforceDocumentProperties(projectID, document)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get properties")
	}

	trackPayload := &SDK.TrackPayload{
		ProjectId:       projectID,
		EventProperties: properties,
		UserProperties:  properties,
	}

	eventID, userID, err := TrackSalesforceEventByDocumentType(projectID, trackPayload, document)
	if err != nil {
		logCtx.WithError(err).Error(
			"Failed to track salesforce account event.")
		return http.StatusInternalServerError
	}

	customerUserID, _ := getCustomerUserIDFromProperties(projectID, properties, M.GetSalesforceAliasByDocType(document.Type))
	if customerUserID != "" {
		status, _ := SDK.Identify(projectID, &SDK.IdentifyPayload{
			UserId: userID, CustomerUserId: customerUserID})
		if status != http.StatusOK {
			logCtx.WithField("customer_user_id", customerUserID).Error(
				"Failed to identify user on salesforce account sync.")
			return http.StatusInternalServerError
		}
	} else {
		logCtx.Error("Skipped user identification on salesforce account sync. No customer_user_id on properties.")
	}

	errCode := M.UpdateSalesforceDocumentAsSynced(projectID, document, eventID)
	if errCode != http.StatusAccepted {
		logCtx.Error("Failed to update salesforce account document as synced.")
		return http.StatusInternalServerError
	}

	return http.StatusOK
}

func enrichContact(projectID uint64, document *M.SalesforceDocument) int {
	if projectID == 0 || document == nil {
		return http.StatusBadRequest
	}

	if document.Type != M.SalesforceDocumentTypeContact {
		return http.StatusInternalServerError
	}

	logCtx := log.WithField("project_id", projectID).WithField("document_id", document.ID)
	properties, err := getSalesforceDocumentProperties(projectID, document)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get properties")
	}

	trackPayload := &SDK.TrackPayload{
		ProjectId:       projectID,
		EventProperties: properties,
		UserProperties:  properties,
	}

	eventID, userID, err := TrackSalesforceEventByDocumentType(projectID, trackPayload, document)
	if err != nil {
		logCtx.WithError(err).Error(
			"Failed to track salesforce contact event.")
		return http.StatusInternalServerError
	}

	customerUserID, _ := getCustomerUserIDFromProperties(projectID, properties, M.GetSalesforceAliasByDocType(document.Type))
	if customerUserID != "" {
		status, _ := SDK.Identify(projectID, &SDK.IdentifyPayload{
			UserId: userID, CustomerUserId: customerUserID})
		if status != http.StatusOK {
			logCtx.WithField("customer_user_id", customerUserID).Error(
				"Failed to identify user on salesforce contact sync.")
			return http.StatusInternalServerError
		}
	} else {
		logCtx.Error("Skipped user identification on salesforce contact sync. No customer_user_id on properties.")
	}

	errCode := M.UpdateSalesforceDocumentAsSynced(projectID, document, eventID)
	if errCode != http.StatusAccepted {
		logCtx.Error("Failed to update salesforce account document as synced.")
		return http.StatusInternalServerError
	}

	return http.StatusOK
}

func enrichOpportunities(projectID uint64, document *M.SalesforceDocument) int {
	if projectID == 0 || document == nil {
		return http.StatusBadRequest
	}

	if document.Type != M.SalesforceDocumentTypeOpportunity {
		return http.StatusInternalServerError
	}

	logCtx := log.WithField("project_id", projectID).WithField("document_id", document.ID)
	properties, err := getSalesforceDocumentProperties(projectID, document)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get properties")
	}

	trackPayload := &SDK.TrackPayload{
		ProjectId:       projectID,
		EventProperties: properties,
		UserProperties:  properties,
	}

	var eventID string
	customerUserID, userID := getCustomerUserIDFromProperties(projectID, properties, M.GetSalesforceAliasByDocType(document.Type))
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

	errCode := M.UpdateSalesforceDocumentAsSynced(projectID, document, eventID)
	if errCode != http.StatusAccepted {
		logCtx.Error("Failed to update salesforce opportunity document as synced.")
		return http.StatusInternalServerError
	}

	return http.StatusOK
}

func enrichLeads(projectID uint64, document *M.SalesforceDocument) int {
	if projectID == 0 || document == nil {
		return http.StatusBadRequest
	}

	if document.Type != M.SalesforceDocumentTypeLead {
		return http.StatusInternalServerError
	}

	logCtx := log.WithField("project_id", projectID).WithField("document_id", document.ID)

	properties, err := getSalesforceDocumentProperties(projectID, document)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get properties")
	}

	trackPayload := &SDK.TrackPayload{
		ProjectId:       projectID,
		EventProperties: properties,
		UserProperties:  properties,
	}

	eventID, userID, err := TrackSalesforceEventByDocumentType(projectID, trackPayload, document)
	if err != nil {
		logCtx.WithError(err).Error(
			"Failed to track salesforce lead event.")
		return http.StatusInternalServerError
	}

	customerUserID, _ := getCustomerUserIDFromProperties(projectID, properties, M.GetSalesforceAliasByDocType(document.Type))
	if customerUserID != "" {
		status, _ := SDK.Identify(projectID, &SDK.IdentifyPayload{
			UserId: userID, CustomerUserId: customerUserID})
		if status != http.StatusOK {
			logCtx.WithField("customer_user_id", customerUserID).Error(
				"Failed to identify user on salesforce lead sync.")
			return http.StatusInternalServerError
		}
	} else {
		logCtx.Error("Skipped user identification on salesforce lead sync. No customer_user_id on properties.")
	}

	errCode := M.UpdateSalesforceDocumentAsSynced(projectID, document, eventID)
	if errCode != http.StatusAccepted {
		logCtx.Error("Failed to update salesforce lead document as synced.")
		return http.StatusInternalServerError
	}

	return http.StatusOK
}

func enrichAll(projectID uint64, documents []M.SalesforceDocument) int {
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
			errCode = enrichAccount(projectID, &documents[i])
		case M.SalesforceDocumentTypeContact:
			errCode = enrichContact(projectID, &documents[i])
		case M.SalesforceDocumentTypeLead:
			errCode = enrichLeads(projectID, &documents[i])
		case M.SalesforceDocumentTypeOpportunity:
			errCode = enrichOpportunities(projectID, &documents[i])
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
	for _, docType := range salesforceSyncOrderByType {
		if _, exist := allowedDocTypes[M.GetSalesforceAliasByDocType(docType)]; !exist {
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
			Type:      M.GetSalesforceAliasByDocType(docType),
		}

		errCode = enrichAll(projectID, documents)
		if errCode == http.StatusOK {
			status.Status = "success"
		} else {
			status.Status = "failures_seen"
		}
		statusByProjectAndType = append(statusByProjectAndType, status)
	}

	return statusByProjectAndType
}
