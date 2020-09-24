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

const SALESFORCE_TOKEN_URL = "login.salesforce.com/services/oauth2/token"
const SALESFORCE_AUTH_URL = "login.salesforce.com/services/oauth2/authorize"
const SALESFORCE_APP_SETTINGS_URL = "/#/settings/salesforce"
const SALESFORCE_REFRESH_TOKEN = "refresh_token"
const SALESFORCE_INSTANCE_URL = "instance_url"

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

type Account struct {
	AccountId  string                 `json:"Id"`
	Properties map[string]interface{} `json:"properties"`
}

type Status struct {
	ProjectId uint64 `json:"project_id"`
	Type      string `json:"type"`
	Status    string `json:"status"`
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
			userLatest, errCode := M.GetUserLatestByCustomerUserId(projectId, pPhoneNo)
			if errCode == http.StatusFound {
				return userLatest.ID
			}
		}
	}

	return ""
}

func removeEmptyFieldsFromProperties(properties map[string]interface{}) error {
	for field, value := range properties {
		if value == nil || value == "" {
			delete(properties, field)
		}
	}
	return nil
}

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
			return "", "", fmt.Errorf("created event track failed to doc type %d", document.Type)
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
			return "", "", fmt.Errorf("updated event track failed to doc type %d", document.Type)
		}

		eventId = response.EventId
	} else {
		return "", "", errors.New("invalid action on salesforce document sync.")
	}

	return eventId, userId, nil
}
func removeSkipableFieldsFromProperties(properties map[string]interface{}) error {
	for _, field := range M.SalesforceSkippablefields {
		if _, exist := properties[field]; exist {
			delete(properties, field)
		}
	}
	return nil
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

	removeEmptyFieldsFromProperties(properties)
	removeSkipableFieldsFromProperties(properties)
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

func syncContact(projectId uint64, document *M.SalesforceDocument) int {
	if document.Type != M.SalesforceDocumentTypeContact {
		return http.StatusInternalServerError
	}

	logCtx := log.WithField("project_id", projectId).WithField("document_id", document.ID)
	properties, err := getSalesforceDocumentProperties(document)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get properties")
	}
	removeEmptyFieldsFromProperties(properties)
	removeSkipableFieldsFromProperties(properties)
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
	removeEmptyFieldsFromProperties(properties)
	removeSkipableFieldsFromProperties(properties)
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

func SyncEnrichment(projectId uint64) []Status {
	logCtx := log.WithField("project_id", projectId)

	statusByProjectAndType := make([]Status, 0, 0)

	for _, docType := range M.SalesforceSupportedDocumentType {
		logCtx = logCtx.WithField("type", docType)
		documents, errCode := GetSalesforceDocumentsByTypeForSync(projectId, docType)
		if errCode != http.StatusFound {
			logCtx.Error("Failed to get salesforce document by type for sync.")
			return statusByProjectAndType
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
	}

	return statusByProjectAndType
}

type field map[string]interface{}

type record map[string]interface{}

type Describe struct {
	Custom bool    `json:"custom"`
	Fields []field `json:"fields"`
}

const (
	SALESFORCE_DATA_SERVICE_ROUTE = "/services/data/"
	SALESFORCE_API_VERSION        = "v20.0"
)

func getSalesforceObjectDescription(objectName, accessToken, instanceURL string) (*Describe, error) {
	url := instanceURL + SALESFORCE_DATA_SERVICE_ROUTE + SALESFORCE_API_VERSION + "/sobjects/" + objectName + "/describe"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer"+" "+accessToken)
	req.Header.Add("Accept", "application/json")

	client := &http.Client{
		Timeout: 10 * time.Minute,
	}

	resp, err := client.Do(req)
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
		return nil, errors.New("failed to decode response")
	}
	return &jsonRespone, nil

}

func getFieldsListFromDescription(description *Describe) ([]string, error) {
	var objectFields []string
	objectFieldDescriptions := description.Fields

	for _, fieldDescription := range objectFieldDescriptions {
		if fieldName, ok := fieldDescription["name"].(string); ok {
			objectFields = append(objectFields, fieldName)
		}
	}

	return objectFields, nil

}

type QueryRespone struct {
	TotalSize      int      `json:"totalSize"`
	Done           bool     `json:"done"`
	Records        []record `json:"records"`
	NextRecordsUrl string   `json:"nextRecordsUrl"`
}

func getSalesforceDataByQuery(query, accessToken, instanceURL, dateTime string) (*QueryRespone, error) {
	var whereStmnt string
	if dateTime != "" {
		whereStmnt = "WHERE" + "+" + "LastModifiedDate" + url.QueryEscape(">"+dateTime)
	}

	queryURL := instanceURL + SALESFORCE_DATA_SERVICE_ROUTE + SALESFORCE_API_VERSION + "/query?q=" + query + "+" + whereStmnt
	req, err := http.NewRequest("GET", queryURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer"+" "+accessToken)
	req.Header.Add("Accept", "application/json")
	client := &http.Client{
		Timeout: 10 * time.Minute,
	}
	resp, err := client.Do(req)
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

func syncByType(ps *M.SalesforceProjectSettings, accessToken, objectName, dateTime string) (bool, error) {
	logCtx := log.WithField("project_id", ps.ProjectId)
	description, err := getSalesforceObjectDescription(objectName, accessToken, ps.InstanceURL)
	if err != nil {
		return false, err
	}

	fields, err := getFieldsListFromDescription(description)
	if err != nil {
		return false, err
	}

	selectStmnt := strings.Join(fields, ",")
	queryStmnt := fmt.Sprintf("SELECT+%s+FROM+%s", selectStmnt, objectName)
	queryRespone, err := getSalesforceDataByQuery(queryStmnt, accessToken, ps.InstanceURL, dateTime)
	if err != nil {
		return false, err
	}
	records := queryRespone.Records

	hasFailures := false
	for i := range records {
		var document M.SalesforceDocument
		document.ProjectId = ps.ProjectId
		document.TypeAlias = objectName
		enValue, err := json.Marshal(records[i])
		if err != nil {
			hasFailures = true
			logCtx.WithError(err).Error("Error while encoding record for upserting")
			continue
		}
		document.Value = &postgres.Jsonb{RawMessage: json.RawMessage(enValue)}
		status := M.CreateSalesforceDocument(ps.ProjectId, &document)
		if status != http.StatusCreated && status != http.StatusConflict {
			hasFailures = true
			logCtx.Errorf("Error while create salesforce record status %d", status)
			continue
		}
	}

	hasMore := !queryRespone.Done
	for hasMore {
		nextBatchRoute := queryRespone.NextRecordsUrl
		queryRespone, _ = getSalesforceNextBatch(nextBatchRoute, ps.InstanceURL, accessToken)
		records = queryRespone.Records
		for i := range records {
			var document M.SalesforceDocument
			document.ProjectId = ps.ProjectId
			document.TypeAlias = objectName
			enValue, err := json.Marshal(records[i])
			if err != nil {
				hasFailures = true
				logCtx.WithError(err).Error("Error while encoding record for upserting")
				continue
			}
			document.Value = &postgres.Jsonb{RawMessage: json.RawMessage(enValue)}
			status := M.CreateSalesforceDocument(ps.ProjectId, &document)
			if status != http.StatusCreated && status != http.StatusConflict {
				hasFailures = true
				logCtx.Errorf("Error while create salesforce record status %d", status)
				continue
			}
		}
		hasMore = queryRespone.Done
	}

	return hasFailures, nil
}
func getSalesforceNextBatch(nextBatchRoute, InstanceURL string, accessToken string) (*QueryRespone, error) {
	url := InstanceURL + nextBatchRoute
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer"+" "+accessToken)
	req.Header.Add("Accept", "application/json")
	client := http.Client{
		Timeout: 10 * time.Minute,
	}
	resp, err := client.Do(req)
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

const REFRESH_TOKEN_URL = "https://login.salesforce.com/services/oauth2/token"

func GetAccessToken(ps *M.SalesforceProjectSettings, redirectUrl string) (string, error) {
	queryParams := fmt.Sprintf("grant_type=%s&refresh_token=%s&client_id=%s&client_secret=%s&redirect_uri=%s",
		"refresh_token", ps.RefreshToken, C.GetSalesforceAppId(), C.GetSalesforceAppSecret(), redirectUrl)
	url := REFRESH_TOKEN_URL + "?" + queryParams
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("Accept", "application/json")

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
		return "", errors.New("failed to decode response")
	}

	access_token, exists := jsonRespone["access_token"].(string)
	if !exists && access_token == "" {
		return "", errors.New("failed to get access token by refresh token")
	}

	return access_token, nil
}

func SyncDocuments(ps *M.SalesforceProjectSettings, lastSyncInfo map[string]int64, accessToken string) {
	// logCtx := log.WithField("project_id", ps.ProjectId)
	for docType, timeStamp := range lastSyncInfo {
		var sfFormatedTime string
		if timeStamp != 0 {
			t := time.Unix(timeStamp, 0)
			sfFormatedTime = t.UTC().Format(M.SalesforceDocumentTimeLayout)
		}

		syncByType(ps, accessToken, docType, sfFormatedTime)
	}
}
