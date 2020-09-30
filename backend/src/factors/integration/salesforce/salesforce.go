package salesforce

import (
	"encoding/json"
	"errors"
	C "factors/config"
	M "factors/model"
	SDK "factors/sdk"
	U "factors/util"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

const (
	SALESFORCE_TOKEN_URL          = "login.salesforce.com/services/oauth2/token"
	SALESFORCE_AUTH_URL           = "login.salesforce.com/services/oauth2/authorize"
	REFRESH_TOKEN_URL             = "https://login.salesforce.com/services/oauth2/token"
	SALESFORCE_APP_SETTINGS_URL   = "/#/settings/salesforce"
	SALESFORCE_REFRESH_TOKEN      = "refresh_token"
	SALESFORCE_INSTANCE_URL       = "instance_url"
	SALESFORCE_DATA_SERVICE_ROUTE = "/services/data/"
	SALESFORCE_API_VERSION        = "v20.0"
)

type OAuthState struct {
	ProjectId uint64  `json:"pid"`
	AgentUUID *string `json:"aid"`
}

// SalesforceAuthParams common struct throughout auth
type SalesforceAuthParams struct {
	GrantType    string `token_param:"grant_type"`
	AccessCode   string `token_param:"code"`
	ClientSecret string `token_param:"client_secret"`
	ClientId     string `token_param:"client_id" auth_param:"client_id" `
	RedirectURL  string `token_param:"redirect_uri" auth_param:"redirect_uri"`
	ResponseType string `auth_param:"response_type"`
	State        string `auth_param:"state"`
}

type Status struct {
	ProjectId uint64 `json:"project_id"`
	Type      string `json:"type"`
	Status    string `json:"status"`
}

//Salesforce API structs
type field map[string]interface{}

type record map[string]interface{}

type Describe struct {
	Custom bool    `json:"custom"`
	Fields []field `json:"fields"`
}

type QueryRespone struct {
	TotalSize      int      `json:"totalSize"`
	Done           bool     `json:"done"`
	Records        []record `json:"records"`
	NextRecordsUrl string   `json:"nextRecordsUrl"`
}

type SalesforceObjectStatus struct {
	ProjetId     uint64   `json:"project_id"`
	Status       string   `json:"status"`
	DocType      string   `json:"doc_type"`
	TotalRecords int      `json:"total_records"`
	Message      string   `json:"message,omitempty"`
	SyncAll      bool     `json:"syncall"`
	Failures     []string `json:"failures,omitempty"`
}

type SalesforceJobStatus struct {
	Status   string                   `json:"status"`
	Success  []SalesforceObjectStatus `json:"success"`
	Failures []SalesforceObjectStatus `json:"failures"`
}

func GetSalesforceUserToken(salesforceTokenParams *SalesforceAuthParams) (map[string]interface{}, error) {
	var credentials map[string]interface{}
	urlParamsStr, err := buildQueryParamsByTagName(*salesforceTokenParams, "token_param")
	if err != nil {
		return credentials, errors.New("failed to build query parameter")
	}

	tokenUrl := fmt.Sprintf("https://%s?%s", SALESFORCE_TOKEN_URL, urlParamsStr)
	resp, err := http.Post(tokenUrl, "application/json", strings.NewReader(""))
	if err != nil {
		return credentials, errors.New("failed to build request to salesforce tokenUrl")
	}

	if resp.StatusCode != http.StatusOK {
		return credentials, errors.New("fetching salesforce user credentials failed")
	}

	err = json.NewDecoder(resp.Body).Decode(&credentials)
	if err != nil {
		return credentials, errors.New("failed to decode salesforce token response")
	}
	return credentials, nil
}

func GetSalesforceAuthorizationUrl(clientId, redirectUrl, responseType, state string) string {
	baseUrl := "https://" + SALESFORCE_AUTH_URL
	urlParams := SalesforceAuthParams{
		ClientId:     clientId,
		RedirectURL:  redirectUrl,
		ResponseType: responseType,
		State:        state,
	}

	urlParamsStr, err := buildQueryParamsByTagName(urlParams, "auth_param")
	if err != nil {
		return ""
	}

	return fmt.Sprintf("%s?%s", baseUrl, urlParamsStr)
}

// buildQueryParamsByTagName generates url parameters by struct tags
func buildQueryParamsByTagName(params interface{}, tag string) (string, error) {
	rParams := reflect.ValueOf(params)
	if rParams.Kind() != reflect.Struct {
		return "", errors.New("params must be struct type")
	}

	var urlParams string
	paramsTyp := rParams.Type()
	for i := 0; i < rParams.NumField(); i++ {
		paramField := paramsTyp.Field(i)
		if tagName := paramField.Tag.Get(tag); tagName != "" {
			if urlParams == "" {
				urlParams = tagName + "=" + rParams.Field(i).Interface().(string)
			} else {
				urlParams = urlParams + "&" + tagName + "=" + rParams.Field(i).Interface().(string)
			}
		}

	}
	return urlParams, nil
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

func getUserIdFromLastestProperties(properties []M.UserProperties) string {
	latestIndex := len(properties) - 1
	return properties[latestIndex].UserId
}

func getCustomerUserIdFromProperties(projectId uint64, properties map[string]interface{}) string {

	if phoneNo, ok := properties["MobilePhone"].(string); ok && phoneNo != "" {
		for pPhoneNo := range U.GetPossiblePhoneNumber(phoneNo) {
			_, errCode := M.GetUserLatestByCustomerUserId(projectId, pPhoneNo)
			if errCode == http.StatusFound {
				return pPhoneNo
			}
		}
		return phoneNo
	}

	if phoneNo, ok := properties["Phone"].(string); ok && phoneNo != "" {
		for pPhoneNo := range U.GetPossiblePhoneNumber(phoneNo) {
			_, errCode := M.GetUserLatestByCustomerUserId(projectId, pPhoneNo)
			if errCode == http.StatusFound {
				return pPhoneNo
			}
		}
		return phoneNo
	}

	return ""
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

func buildSalesforceGETRequest(url, accessToken string) (*http.Request, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer"+" "+accessToken)
	return req, nil
}

// SalesforceGetRequest performs GET request on provided url with access token
func SalesforceGetRequest(url, accessToken string) (*http.Response, error) {
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

func getSalesforceObjectDescription(objectName, accessToken, instanceURL string) (*Describe, error) {
	if objectName == "" || accessToken == "" || instanceURL == "" {
		return nil, errors.New("missing required fields")
	}

	url := instanceURL + SALESFORCE_DATA_SERVICE_ROUTE + SALESFORCE_API_VERSION + "/sobjects/" + objectName + "/describe"
	resp, err := SalesforceGetRequest(url, accessToken)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error while getting object description with respone code %d ", resp.StatusCode)
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

func getSalesforceDataByQuery(query, accessToken, instanceURL, dateTime string) (*QueryRespone, error) {
	if query == "" || accessToken == "" || instanceURL == "" {
		return nil, errors.New("missing required fields")
	}

	var whereStmnt string
	if dateTime != "" {
		whereStmnt = "WHERE" + "+" + "LastModifiedDate" + url.QueryEscape(">"+dateTime)
	}

	queryURL := instanceURL + SALESFORCE_DATA_SERVICE_ROUTE + SALESFORCE_API_VERSION + "/query?q=" + query + "+" + whereStmnt
	resp, err := SalesforceGetRequest(queryURL, accessToken)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error while query data with respone code %d ", resp.StatusCode)
	}

	var jsonRespone QueryRespone
	err = json.NewDecoder(resp.Body).Decode(&jsonRespone)
	if err != nil {
		return nil, errors.New("failed to decode response")
	}
	return &jsonRespone, nil
}

func buildAndUpsertDocument(projectId uint64, objectName string, value record) error {
	if projectId == 0 {
		return errors.New("invalid project id")
	}
	if objectName == "" || value == nil {
		return errors.New("invalid oject name or value")
	}

	var document M.SalesforceDocument
	document.ProjectId = projectId
	document.TypeAlias = objectName
	enValue, err := json.Marshal(value)
	if err != nil {
		return err
	}

	document.Value = &postgres.Jsonb{RawMessage: json.RawMessage(enValue)}
	status := M.CreateSalesforceDocument(projectId, &document)
	if status != http.StatusCreated && status != http.StatusConflict {
		return fmt.Errorf("error while creating document Status %d", status)
	}

	return nil
}

func syncByType(ps *M.SalesforceProjectSettings, accessToken, objectName, dateTime string) (SalesforceObjectStatus, error) {
	var salesforceObjectStatus SalesforceObjectStatus
	salesforceObjectStatus.ProjetId = ps.ProjectId
	salesforceObjectStatus.DocType = objectName

	logCtx := log.WithFields(log.Fields{"project_id": ps.ProjectId, "doc_type": objectName})

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
	queryRespone, err := getSalesforceDataByQuery(queryStmnt, accessToken, ps.InstanceURL, dateTime)
	if err != nil {
		logCtx.WithError(err).Error("Failed to getSalesforceDataByQuery.")
		return salesforceObjectStatus, err
	}
	salesforceObjectStatus.TotalRecords = queryRespone.TotalSize
	records := queryRespone.Records

	hasMore := true
	nextBatchRoute := ""
	for hasMore {
		if nextBatchRoute != "" {
			queryRespone, err = getSalesforceNextBatch(nextBatchRoute, ps.InstanceURL, accessToken)
			if err != nil {
				logCtx.WithError(err).Error("Failed to getSalesforceNextBatch.")
				return salesforceObjectStatus, err
			}
			records = queryRespone.Records
		}

		var failures []string
		for i := range records {
			err = buildAndUpsertDocument(ps.ProjectId, objectName, records[i])
			if err != nil {
				logCtx.WithError(err).Error("Failed to buildAndUpsertDocument.")
				failures = append(failures, err.Error())
			}
		}

		salesforceObjectStatus.Failures = append(salesforceObjectStatus.Failures, failures...)
		hasMore = !queryRespone.Done
		nextBatchRoute = queryRespone.NextRecordsUrl
	}

	return salesforceObjectStatus, nil
}

func getSalesforceNextBatch(nextBatchRoute, InstanceURL string, accessToken string) (*QueryRespone, error) {
	if nextBatchRoute == "" || InstanceURL == "" || accessToken == "" {
		return nil, errors.New("missing required fields")
	}
	url := InstanceURL + nextBatchRoute
	resp, err := SalesforceGetRequest(url, accessToken)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error while query next batch with respone code %d ", resp.StatusCode)
	}

	var jsonRespone QueryRespone
	err = json.NewDecoder(resp.Body).Decode(&jsonRespone)
	if err != nil {
		return nil, errors.New("failed to decode response")
	}

	return &jsonRespone, nil
}

// GetAccessToken gets new salesforce access token by refresh token
func GetAccessToken(ps *M.SalesforceProjectSettings, redirectUrl string) (string, error) {
	queryParams := fmt.Sprintf("grant_type=%s&refresh_token=%s&client_id=%s&client_secret=%s&redirect_uri=%s",
		"refresh_token", ps.RefreshToken, C.GetSalesforceAppId(), C.GetSalesforceAppSecret(), redirectUrl)
	url := REFRESH_TOKEN_URL + "?" + queryParams

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
		return "", fmt.Errorf("error while query data with respone code %d ", resp.StatusCode)
	}

	var jsonRespone map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&jsonRespone)
	if err != nil {
		return "", err
	}

	access_token, exists := jsonRespone["access_token"].(string)
	if !exists && access_token == "" {
		return "", errors.New("failed to get access token by refresh token")
	}

	return access_token, nil
}

// SyncDocuments syncs from salesforce to database by doc type
func SyncDocuments(ps *M.SalesforceProjectSettings, lastSyncInfo map[string]int64, accessToken string) []SalesforceObjectStatus {
	var allObjectStatus []SalesforceObjectStatus

	for docType, timestamp := range lastSyncInfo {
		var sfFormatedTime string
		if timestamp != 0 {
			t := time.Unix(timestamp, 0)
			sfFormatedTime = t.UTC().Format(M.SalesforceDocumentTimeLayout)
		}

		objectStatus, err := syncByType(ps, accessToken, docType, sfFormatedTime)
		if err != nil || len(objectStatus.Failures) != 0 {
			log.WithFields(log.Fields{
				"project_id": ps.ProjectId,
				"doctype":    docType,
			}).WithError(err).Errorf("Failed to sync documents")

			if err != nil {
				objectStatus.Message = err.Error()
			}

			objectStatus.Status = "Has failures"
		} else {
			objectStatus.Status = "Success"
		}

		objectStatus.SyncAll = timestamp == 0
		allObjectStatus = append(allObjectStatus, objectStatus)
	}

	return allObjectStatus
}
