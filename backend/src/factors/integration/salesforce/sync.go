package salesforce

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	C "factors/config"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"

	"github.com/jinzhu/now"
	log "github.com/sirupsen/logrus"
)

const (
	// InstanceURL field instance_url
	InstanceURL                = "instance_url"
	salesforceDataServiceRoute = "/services/data/"
	salesforceAPIVersion       = "v20.0"
)

//Salesforce API structs
type field map[string]interface{}

// Describe structure for salesforce describe API
type Describe struct {
	Custom bool    `json:"custom"`
	Fields []field `json:"fields"`
}

// QueryResponse structure for query API
type QueryResponse struct {
	TotalSize      int                      `json:"totalSize"`
	Done           bool                     `json:"done"`
	Records        []model.SalesforceRecord `json:"records"`
	NextRecordsURL string                   `json:"nextRecordsUrl"`
}

// ObjectStatus represents sync info from query to db
type ObjectStatus struct {
	ProjetID     uint64   `json:"project_id"`
	Status       string   `json:"status"`
	DocType      string   `json:"doc_type"`
	TotalRecords int      `json:"total_records"`
	Message      string   `json:"message,omitempty"`
	SyncAll      bool     `json:"syncall"`
	Failures     []string `json:"failures,omitempty"`
}

// JobStatus list all success and failed while sync from salesforce to db
type JobStatus struct {
	Status   string         `json:"status"`
	Success  []ObjectStatus `json:"success"`
	Failures []ObjectStatus `json:"failures"`
}

func buildSalesforceGETRequest(url, accessToken string) (*http.Request, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer"+" "+accessToken)
	return req, nil
}

// GETRequest performs GET request on provided url with access token
func GETRequest(url, accessToken string) (*http.Response, error) {
	req, err := buildSalesforceGETRequest(url, accessToken)
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
	return resp, nil
}

// DataServiceError impelements error interface for salesforce data api error
type DataServiceError struct {
	Message   string `json:"message"`
	ErrorCode string `json:"errorCode"`
}

func getSalesforceObjectDescription(objectName, accessToken, instanceURL string) (*Describe, error) {
	if objectName == "" || accessToken == "" || instanceURL == "" {
		return nil, errors.New("missing required fields")
	}

	url := instanceURL + salesforceDataServiceRoute + salesforceAPIVersion + "/sobjects/" + objectName + "/describe"
	resp, err := GETRequest(url, accessToken)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errBody []DataServiceError
		json.NewDecoder(resp.Body).Decode(&errBody)

		return nil, fmt.Errorf("error while getting object description  %+v", errBody)
	}

	var jsonRespone Describe
	err = json.NewDecoder(resp.Body).Decode(&jsonRespone)
	if err != nil {
		return nil, err
	}

	return &jsonRespone, nil
}

func getFieldsListFromDescription(description *Describe) ([]string, error) {
	var objectFields []string
	objectFieldDescriptions := description.Fields

	if len(description.Fields) == 0 {
		return objectFields, errors.New("invalid fileds on description")
	}

	for _, fieldDescription := range objectFieldDescriptions {
		if fieldName, ok := fieldDescription["name"].(string); ok {
			objectFields = append(objectFields, fieldName)
		}
	}

	if len(objectFields) == 0 {
		return objectFields, errors.New("empty field list")
	}

	return objectFields, nil
}

func getSalesforceDataByQuery(query, accessToken, instanceURL, dateTime string) (*QueryResponse, error) {
	if query == "" || accessToken == "" || instanceURL == "" {
		return nil, errors.New("missing required fields")
	}

	queryURL := instanceURL + salesforceDataServiceRoute + salesforceAPIVersion + "/query?q=" + query

	if dateTime != "" {
		queryURL = queryURL + "+" + "WHERE" + "+" + "LastModifiedDate" + url.QueryEscape(">"+dateTime)
	}

	resp, err := GETRequest(queryURL, accessToken)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errBody []DataServiceError
		json.NewDecoder(resp.Body).Decode(&errBody)

		return nil, fmt.Errorf("error while query data %+v", errBody)
	}

	var jsonResponse QueryResponse
	err = json.NewDecoder(resp.Body).Decode(&jsonResponse)
	if err != nil {
		return nil, errors.New("failed to decode response")
	}
	return &jsonResponse, nil
}

func syncByType(ps *model.SalesforceProjectSettings, accessToken, objectName, dateTime string) (ObjectStatus, error) {
	var salesforceObjectStatus ObjectStatus
	salesforceObjectStatus.ProjetID = ps.ProjectID
	salesforceObjectStatus.DocType = objectName

	logCtx := log.WithFields(log.Fields{"project_id": ps.ProjectID, "doc_type": objectName})

	description, err := getSalesforceObjectDescription(objectName, accessToken, ps.InstanceURL)
	if err != nil {
		logCtx.WithError(err).Error("Failed to getSalesforceObjectDescription.")
		return salesforceObjectStatus, err
	}

	fields, err := getFieldsListFromDescription(description)
	if err != nil {
		logCtx.WithError(err).Error("Failed to getFieldsListFromDescription.")
		return salesforceObjectStatus, err
	}

	selectStmnt := strings.Join(fields, ",")
	queryStmnt := fmt.Sprintf("SELECT+%s+FROM+%s", selectStmnt, objectName)
	queryResponse, err := getSalesforceDataByQuery(queryStmnt, accessToken, ps.InstanceURL, dateTime)
	if err != nil {
		logCtx.WithError(err).Error("Failed to getSalesforceDataByQuery.")
		return salesforceObjectStatus, err
	}
	salesforceObjectStatus.TotalRecords = queryResponse.TotalSize
	records := queryResponse.Records

	hasMore := true
	nextBatchRoute := ""
	for hasMore {
		if nextBatchRoute != "" {
			queryResponse, err = getSalesforceNextBatch(nextBatchRoute, ps.InstanceURL, accessToken)
			if err != nil {
				logCtx.WithError(err).Error("Failed to getSalesforceNextBatch.")
				return salesforceObjectStatus, err
			}
			records = queryResponse.Records
		}

		var failures []string
		for i := range records {
			err = store.GetStore().BuildAndUpsertDocument(ps.ProjectID, objectName, records[i])
			if err != nil {
				logCtx.WithError(err).Error("Failed to BuildAndUpsertDocument.")
				failures = append(failures, err.Error())
			}
		}

		salesforceObjectStatus.Failures = append(salesforceObjectStatus.Failures, failures...)
		hasMore = !queryResponse.Done
		nextBatchRoute = queryResponse.NextRecordsURL
		records = make([]model.SalesforceRecord, 0)
	}

	return salesforceObjectStatus, nil
}

func getSalesforceNextBatch(nextBatchRoute, InstanceURL string, accessToken string) (*QueryResponse, error) {
	if nextBatchRoute == "" || InstanceURL == "" || accessToken == "" {
		return nil, errors.New("missing required fields")
	}
	url := InstanceURL + nextBatchRoute
	resp, err := GETRequest(url, accessToken)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		var errBody []DataServiceError
		json.NewDecoder(resp.Body).Decode(&errBody)
		return nil, fmt.Errorf("error while query next batch %+v ", errBody)
	}

	var jsonRespone QueryResponse
	err = json.NewDecoder(resp.Body).Decode(&jsonRespone)
	if err != nil {
		return nil, errors.New("failed to decode response")
	}

	return &jsonRespone, nil
}

// TokenError implements error interface for token api error
type TokenError struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

// GetAccessToken gets new salesforce access token by refresh token
func GetAccessToken(ps *model.SalesforceProjectSettings, redirectURL string) (string, error) {
	if ps == nil || redirectURL == "" {
		return "", errors.New("invalid project setting or redirect url")
	}

	queryParams := fmt.Sprintf("grant_type=%s&refresh_token=%s&client_id=%s&client_secret=%s&redirect_uri=%s",
		"refresh_token", ps.RefreshToken, C.GetSalesforceAppId(), C.GetSalesforceAppSecret(), redirectURL)
	url := RefreshTokenURL + "?" + queryParams

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return "", err
	}

	client := &http.Client{
		Timeout: 10 * time.Minute,
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		var errBody TokenError
		json.NewDecoder(resp.Body).Decode(&errBody)
		return "", fmt.Errorf("error while query data %s : %s", errBody.Error, errBody.ErrorDescription)
	}

	var jsonResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&jsonResponse)
	if err != nil {
		return "", err
	}

	accessToken, exists := jsonResponse["access_token"].(string)
	if !exists && accessToken == "" {
		return "", errors.New("failed to get access token by refresh token")
	}

	return accessToken, nil
}

// CreateOrGetSalesforceEventName makes sure salesforce event name exists
func CreateOrGetSalesforceEventName(projectID uint64) int {
	logCtx := log.WithFields(log.Fields{"project_id": projectID})

	for _, doctype := range model.GetSalesforceAllowedObjects(projectID) {
		typAlias := model.GetSalesforceAliasByDocType(doctype)
		eventName := model.GetSalesforceEventNameByAction(typAlias, model.SalesforceDocumentCreated)
		_, status := store.GetStore().CreateOrGetEventName(&model.EventName{
			ProjectId: projectID,
			Name:      eventName,
			Type:      model.TYPE_USER_CREATED_EVENT_NAME,
		})

		if status != http.StatusFound && status != http.StatusConflict && status != http.StatusCreated {
			logCtx.Error("Failed to create event name on SyncDatetimeAndNumericalProperties.")
			return http.StatusInternalServerError
		}

		eventName = model.GetSalesforceEventNameByAction(typAlias, model.SalesforceDocumentUpdated)
		_, status = store.GetStore().CreateOrGetEventName(&model.EventName{
			ProjectId: projectID,
			Name:      eventName,
			Type:      model.TYPE_USER_CREATED_EVENT_NAME,
		})

		if status != http.StatusFound && status != http.StatusConflict && status != http.StatusCreated {
			logCtx.Error("Failed to create updated event name on SyncDatetimeAndNumericalProperties.")
			return http.StatusInternalServerError
		}
	}

	return http.StatusOK
}

func syncSalesforcePropertyByType(projectID uint64, doctTypeAlias string, fieldName, fieldType string) error {
	logCtx := log.WithFields(log.Fields{"project_id": projectID, "field_name": fieldName, "field_type": fieldType, "doc_type_alias": doctTypeAlias})

	if fieldName == "" || fieldType == "" || projectID == 0 || doctTypeAlias == "" {
		logCtx.Error("Missing required field.")
		return errors.New("missing required field")
	}

	pType := model.GetSalesforceMappedDataType(U.GetPropertyValueAsString(fieldType))

	enKey := model.GetCRMEnrichPropertyKeyByType(
		model.SmartCRMEventSourceSalesforce,
		doctTypeAlias,
		fieldName,
	)

	eventName := model.GetSalesforceEventNameByAction(doctTypeAlias, model.SalesforceDocumentCreated)
	err := store.GetStore().CreateOrDeletePropertyDetails(projectID, eventName, enKey, pType, false, true)
	if err != nil {
		logCtx.WithFields(log.Fields{"enriched_property_key": enKey}).WithError(err).
			Error("Failed to create created event property details.")
		return err
	}

	eventName = model.GetSalesforceEventNameByAction(doctTypeAlias, model.SalesforceDocumentUpdated)
	err = store.GetStore().CreateOrDeletePropertyDetails(projectID, eventName, enKey, pType, false, true)
	if err != nil {
		logCtx.WithFields(log.Fields{"enriched_property_key": enKey}).WithError(err).
			Error("Failed to create updated event property details.")
		return err
	}

	err = store.GetStore().CreateOrDeletePropertyDetails(projectID, "", enKey, pType, true, true)
	if err != nil {
		logCtx.WithFields(log.Fields{"enriched_property_key": enKey}).WithError(err).
			Error("Failed to create user property details.")
		return err
	}

	return nil
}

// SyncDatetimeAndNumericalProperties sync datetime and numerical properties to the property_details table
func SyncDatetimeAndNumericalProperties(projectID uint64, accessToken, instanceURL string) (bool, []Status) {
	if projectID == 0 || accessToken == "" || instanceURL == "" {
		return false, nil
	}

	logCtx := log.WithFields(log.Fields{"project_id": projectID})
	status := CreateOrGetSalesforceEventName(projectID)
	if status != http.StatusOK {
		logCtx.Errorf("Failed to CreateOrGetSalesforceEventName status %d", status)
		return true, nil
	}

	var allStatus []Status
	anyFailures := false
	for _, doctype := range model.GetSalesforceAllowedObjects(projectID) {
		var status Status
		typAlias := model.GetSalesforceAliasByDocType(doctype)
		status.Type = typAlias
		status.ProjectID = projectID

		docTypeFailure := false
		describe, err := getSalesforceObjectDescription(typAlias, accessToken, instanceURL)
		if err != nil {
			logCtx.WithError(err).Error("Failed to sync datetime and numerical properties.")
			anyFailures = true
			continue
		}

		for i := range describe.Fields {
			fieldType, exist := describe.Fields[i]["type"]
			if !exist {
				logCtx.WithFields(log.Fields{"property_type": fieldType}).Error("Failed to get property type field.")
				docTypeFailure = true
				continue
			}

			fieldName, exist := describe.Fields[i]["name"]
			if !exist {
				logCtx.Error("Failed to get property name field.")
				docTypeFailure = true
				continue
			}

			if failure := syncSalesforcePropertyByType(projectID, typAlias, U.GetPropertyValueAsString(fieldName), U.GetPropertyValueAsString(fieldType)); failure != nil {
				docTypeFailure = true
			}

			label, exist := describe.Fields[i]["label"]
			if !exist {
				logCtx.Error("Failed to get property label.")
			} else {
				logCtx.Info("Inserting display names")
				err := store.GetStore().CreateOrUpdateDisplayNameByObjectType(projectID, model.GetCRMEnrichPropertyKeyByType(
					model.SmartCRMEventSourceSalesforce,
					typAlias,
					U.GetPropertyValueAsString(fieldName),
				), typAlias, U.GetPropertyValueAsString(label), model.SmartCRMEventSourceSalesforce)
				if(err != http.StatusCreated){
					logCtx.Error("Failed to create or update display name")
				}
			}
		}

		if docTypeFailure {
			status.Status = U.CRM_SYNC_STATUS_FAILURES
			anyFailures = true
		} else {
			status.Status = U.CRM_SYNC_STATUS_FAILURES
		}

		allStatus = append(allStatus, status)
	}

	return anyFailures, allStatus
}

// SyncDocuments syncs from salesforce to database by doc type
func SyncDocuments(ps *model.SalesforceProjectSettings, lastSyncInfo map[string]int64, accessToken string) []ObjectStatus {
	var allObjectStatus []ObjectStatus

	for docType, timestamp := range lastSyncInfo {
		var sfFormatedTime string
		var syncAll bool
		if timestamp == 0 {
			currentTime := time.Now().AddDate(0, 0, -30).UTC()
			timestamp = now.New(currentTime).BeginningOfDay().Unix() // get from last 30 days
			syncAll = true
		}

		t := time.Unix(timestamp, 0)
		sfFormatedTime = t.UTC().Format(model.SalesforceDocumentDateTimeLayout)

		objectStatus, err := syncByType(ps, accessToken, docType, sfFormatedTime)
		if err != nil || len(objectStatus.Failures) != 0 {
			log.WithFields(log.Fields{
				"project_id": ps.ProjectID,
				"doctype":    docType,
			}).WithError(err).Errorf("Failed to sync documents")

			if err != nil {
				objectStatus.Message = err.Error()
			}

			objectStatus.Status = U.CRM_SYNC_STATUS_FAILURES
		} else {
			objectStatus.Status = U.CRM_SYNC_STATUS_SUCCESS
		}

		objectStatus.SyncAll = syncAll
		allObjectStatus = append(allObjectStatus, objectStatus)
	}

	return allObjectStatus
}
