package salesforce

import (
	"encoding/json"
	"errors"
	M "factors/model"
	"fmt"
	"net/http"
	"time"

	C "factors/config"
	SDK "factors/sdk"

	log "github.com/sirupsen/logrus"
)

type Status struct {
	ProjectId uint64 `json:"project_id"`
	Type      string `json:"type"`
	Status    string `json:"status"`
}

func getUserIdFromLastestProperties(properties []M.UserProperties) string {
	latestIndex := len(properties) - 1
	return properties[latestIndex].UserId
}

func getSalesforceDocumentProperties(document *M.SalesforceDocument) (map[string]interface{}, error) {
	docType := M.GetSalesforceAliasByDocType(document.Type)
	if docType == "" {
		return nil, errors.New("invalid document type")
	}

	var properties map[string]interface{}
	err := json.Unmarshal(document.Value.RawMessage, &properties)
	if err != nil {
		return nil, err
	}

	return properties, nil
}

func sanatizeFieldsFromProperties(projectId uint64, properties map[string]interface{}, docType int) {

	allowedfields := M.GetSalesforceAllowedfiedsByObject(projectId, M.GetSalesforceAliasByDocType(docType))
	for field, value := range properties {
		if value == nil || value == "" || value == 0 {
			delete(properties, field)
			continue
		}

		if allowedfields != nil {
			if _, exist := allowedfields[field]; !exist {
				delete(properties, field)
			}
		}
	}
}

func getSalesforceAccountId(document *M.SalesforceDocument) (string, error) {
	var properties map[string]interface{}
	err := json.Unmarshal(document.Value.RawMessage, &properties)
	if err != nil {
		return "", err
	}

	var accountId string
	var ok bool
	if accountId, ok = properties["Id"].(string); !ok {
		return "", errors.New("account id doest not exist")
	}

	if accountId == "" {
		return "", errors.New("empty account id")
	}

	return accountId, nil
}

/*
TrackSalesforceEventByDocumentType tracks salesforce events by action
	for action created -> create both created and updated events with date created timestamp
	for action updated -> create on updated event with lastmodified timestamp
*/
func TrackSalesforceEventByDocumentType(projectId uint64, trackPayload *SDK.TrackPayload, document *M.SalesforceDocument) (string, string, error) {

	var eventId, userId string
	var err error
	if document.Action == M.SalesforceDocumentCreated {
		trackPayload.Name = M.GetSalesforceCreatedEventName(document.Type)
		trackPayload.Timestamp, err = M.GetSalesforceDocumentTimestampByAction(document)
		if err != nil {
			return "", "", err
		}

		status, response := SDK.Track(projectId, trackPayload, true, SDK.SourceSalesforce)
		if status != http.StatusOK && status != http.StatusFound && status != http.StatusNotModified {
			return "", "", fmt.Errorf("created event track failed for doc type %d", document.Type)
		}

		eventId = response.EventId
		userId = response.UserId
	}

	if document.Action == M.SalesforceDocumentCreated || document.Action == M.SalesforceDocumentUpdated {
		trackPayload.Name = M.GetSalesforceUpdatedEventName(document.Type)
		trackPayload.Timestamp, err = M.GetSalesforceDocumentTimestampByAction(document)
		if err != nil {
			return "", "", err
		}

		if document.Action == M.SalesforceDocumentUpdated {
			userPropertiesRecords, errCode := M.GetUserPropertiesRecordsByProperty(projectId, "Id", document.ID)
			if errCode != http.StatusFound {
				return "", "", errors.New("failed to get user with given id")
			}
			userId = getUserIdFromLastestProperties(userPropertiesRecords)
		} else {
			trackPayload.UserId = userId
		}

		status, response := SDK.Track(projectId, trackPayload, true, SDK.SourceSalesforce)
		if status != http.StatusOK && status != http.StatusFound && status != http.StatusNotModified {
			return "", "", fmt.Errorf("updated event track failed for doc type %d", document.Type)
		}

		eventId = response.EventId
	} else {
		return "", "", errors.New("invalid action on salesforce document sync.")
	}

	return eventId, userId, nil
}

func syncAccount(projectId uint64, document *M.SalesforceDocument) int {
	if document.Type != M.SalesforceDocumentTypeAccount {
		return http.StatusInternalServerError
	}

	logCtx := log.WithField("project_id", projectId).WithField("document_id", document.ID)

	properties, err := getSalesforceDocumentProperties(document)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get properties")
	}
	sanatizeFieldsFromProperties(projectId, properties, document.Type)
	trackPayload := &SDK.TrackPayload{
		ProjectId:       projectId,
		EventProperties: properties,
		UserProperties:  properties,
	}

	eventId, userId, err := TrackSalesforceEventByDocumentType(projectId, trackPayload, document)
	if err != nil {
		logCtx.WithError(err).Error(
			"Failed to track salesforce account event.")
		return http.StatusInternalServerError
	}

	accountId, err := getSalesforceAccountId(document)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get salesforce account id")
	}

	if accountId != "" {
		status, _ := SDK.Identify(projectId, &SDK.IdentifyPayload{
			UserId: userId, CustomerUserId: accountId,
		})
		if status != http.StatusOK {
			logCtx.WithField("customer_user_id", accountId).Error(
				"Failed to identify user on salesforce account sync.")
			return http.StatusInternalServerError
		}
	}

	errCode := M.UpdateSalesforceDocumentAsSynced(projectId, document.ID, eventId)
	if errCode != http.StatusAccepted {
		logCtx.Error("Failed to update salesforce account document as synced.")
		return http.StatusInternalServerError
	}

	return http.StatusOK
}

func syncContact(projectId uint64, document *M.SalesforceDocument) int {
	if document.Type != M.SalesforceDocumentTypeContact {
		return http.StatusInternalServerError
	}

	logCtx := log.WithField("project_id", projectId).WithField("document_id", document.ID)
	properties, err := getSalesforceDocumentProperties(document)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get properties")
	}
	sanatizeFieldsFromProperties(projectId, properties, document.Type)
	trackPayload := &SDK.TrackPayload{
		ProjectId:       projectId,
		EventProperties: properties,
		UserProperties:  properties,
	}

	eventId, _, err := TrackSalesforceEventByDocumentType(projectId, trackPayload, document)
	if err != nil {
		logCtx.WithError(err).Error(
			"Failed to track salesforce contact event.")
		return http.StatusInternalServerError
	}

	errCode := M.UpdateSalesforceDocumentAsSynced(projectId, document.ID, eventId)
	if errCode != http.StatusAccepted {
		logCtx.Error("Failed to update salesforce account document as synced.")
		return http.StatusInternalServerError
	}

	return http.StatusOK
}

func syncOpportunities(projectId uint64, document *M.SalesforceDocument) int {
	if document.Type != M.SalesforceDocumentTypeOpportunity {
		return http.StatusInternalServerError
	}

	logCtx := log.WithField("project_id", projectId).WithField("document_id", document.ID)
	properties, err := getSalesforceDocumentProperties(document)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get properties")
	}
	sanatizeFieldsFromProperties(projectId, properties, document.Type)
	trackPayload := &SDK.TrackPayload{
		ProjectId:       projectId,
		EventProperties: properties,
		UserProperties:  properties,
	}

	eventId, _, err := TrackSalesforceEventByDocumentType(projectId, trackPayload, document)
	if err != nil {
		logCtx.WithError(err).Error(
			"Failed to track salesforce opportunity event.")
		return http.StatusInternalServerError
	}

	errCode := M.UpdateSalesforceDocumentAsSynced(projectId, document.ID, eventId)
	if errCode != http.StatusAccepted {
		logCtx.Error("Failed to update salesforce opportunity document as synced.")
		return http.StatusInternalServerError
	}

	return http.StatusOK
}

func syncLeads(projectId uint64, document *M.SalesforceDocument) int {
	if document.Type != M.SalesforceDocumentTypeLead {
		return http.StatusInternalServerError
	}

	logCtx := log.WithField("project_id", projectId).WithField("document_id", document.ID)

	properties, err := getSalesforceDocumentProperties(document)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get properties")
	}

	sanatizeFieldsFromProperties(projectId, properties, document.Type)
	trackPayload := &SDK.TrackPayload{
		ProjectId:       projectId,
		EventProperties: properties,
		UserProperties:  properties,
	}

	eventId, userId, err := TrackSalesforceEventByDocumentType(projectId, trackPayload, document)
	if err != nil {
		logCtx.WithError(err).Error(
			"Failed to track salesforce lead event.")
		return http.StatusInternalServerError
	}

	customerUserId := getCustomerUserIdFromProperties(projectId, properties)
	if customerUserId != "" {
		status, _ := SDK.Identify(projectId, &SDK.IdentifyPayload{
			UserId: userId, CustomerUserId: customerUserId})
		if status != http.StatusOK {
			logCtx.WithField("customer_user_id", customerUserId).Error(
				"Failed to identify user on salesforce lead sync.")
			return http.StatusInternalServerError
		}
	} else {
		logCtx.Error("Skipped user identification on salesforce lead sync. No customer_user_id on properties.")
	}

	errCode := M.UpdateSalesforceDocumentAsSynced(projectId, document.ID, eventId)
	if errCode != http.StatusAccepted {
		logCtx.Error("Failed to update salesforce lead document as synced.")
		return http.StatusInternalServerError
	}

	return http.StatusOK
}

func syncAll(projectId uint64, documents []M.SalesforceDocument) int {
	logCtx := log.WithField("project_id", projectId)

	var seenFailures bool
	var errCode int
	for i := range documents {
		startTime := time.Now().Unix()

		switch documents[i].Type {
		case M.SalesforceDocumentTypeAccount:
			errCode = syncAccount(projectId, &documents[i])
		case M.SalesforceDocumentTypeContact:
			errCode = syncContact(projectId, &documents[i])
		case M.SalesforceDocumentTypeLead:
			errCode = syncLeads(projectId, &documents[i])
		case M.SalesforceDocumentTypeOpportunity:
			errCode = syncOpportunities(projectId, &documents[i])
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
func GetSalesforceDocumentsByTypeForSync(projectId uint64, typ int) ([]M.SalesforceDocument, int) {
	logCtx := log.WithFields(log.Fields{"project_id": projectId, "type": typ})

	if projectId == 0 || typ == 0 {
		logCtx.Error("Invalid project_id or type on get salesforce documents by type.")
		return nil, http.StatusBadRequest
	}

	var documents []M.SalesforceDocument

	db := C.GetServices().Db
	err := db.Order("timestamp, created_at ASC").Where("project_id=? AND type=? AND synced=false",
		projectId, typ).Find(&documents).Error
	if err != nil {
		logCtx.WithError(err).Error("Failed to get salesforce documents by type.")
		return nil, http.StatusInternalServerError
	}

	return documents, http.StatusFound
}

// SyncEnrichment sync salesforce documents to events
func SyncEnrichment(projectId uint64) []Status {
	logCtx := log.WithField("project_id", projectId)

	statusByProjectAndType := make([]Status, 0, 0)

	for _, docType := range M.GetSalesforceAllowedObjects(projectId) {
		logCtx = logCtx.WithFields(log.Fields{
			"doc_type":   docType,
			"project_id": projectId,
		})

		documents, errCode := GetSalesforceDocumentsByTypeForSync(projectId, docType)
		if errCode != http.StatusFound {
			logCtx.Error("Failed to get salesforce document by type for sync.")
			continue
		}

		status := Status{
			ProjectId: projectId,
			Type:      M.GetSalesforceAliasByDocType(docType),
		}

		errCode = syncAll(projectId, documents)
		if errCode == http.StatusOK {
			status.Status = "success"
		} else {
			status.Status = "failures_seen"
		}
		statusByProjectAndType = append(statusByProjectAndType, status)
	}

	return statusByProjectAndType
}
