package salesforce

import (
	"encoding/json"
	"errors"
	C "factors/config"
	M "factors/model"
	SDK "factors/sdk"
	"fmt"
	"net/http"
	"reflect"
	"strings"

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

func getAccountProperties(document *M.SalesforceDocument) (map[string]interface{}, error) {
	if document.Type != M.SalesforceDocumentTypeAccount {
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

func syncAccount(projectId uint64, document *M.SalesforceDocument) int {
	if document.Type != M.SalesforceDocumentTypeAccount {
		return http.StatusInternalServerError
	}

	logCtx := log.WithField("project_id", projectId).WithField("document_id", document.ID)

	properties, err := getAccountProperties(document)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get properties")
	}

	trackPayload := &SDK.TrackPayload{
		ProjectId:       projectId,
		EventProperties: properties,
		UserProperties:  properties,
		Timestamp:       document.Timestamp,
	}

	var eventId string
	var userId string
	if document.Action == M.SalesforceDocumentCreated {
		trackPayload.Name = "Sf_account_created"
		status, response := SDK.Track(projectId, trackPayload, true, "Salesforce")
		if status != http.StatusOK && status != http.StatusFound && status != http.StatusNotModified {
			logCtx.WithField("status", status).Error("Failed to track salesfore account created event.")
			return http.StatusInternalServerError
		}

		eventId = response.EventId
		userId = response.UserId
	} else if document.Action == M.SalesforceDocumentUpdated {
		trackPayload.Name = "Sf_account_updated"
		userPropertiesRecords, errCode := M.GetUserPropertiesRecordsByProperty(
			projectId, "Id", document.ID)
		if errCode != http.StatusFound {
			logCtx.WithField("err_code", errCode).Error(
				"Failed to get user with given id. Failed to track salesforce account updated event.")
			return http.StatusInternalServerError
		}

		userId = getUserIdFromLastestProperties(userPropertiesRecords)
		trackPayload.UserId = userId
		status, response := SDK.Track(projectId, trackPayload, true, "Salesforce")
		if status != http.StatusOK && status != http.StatusFound && status != http.StatusNotModified {
			logCtx.WithField("status", status).Error("Failed to track salesforce account updated event.")
			return http.StatusInternalServerError
		}

		eventId = response.EventId
	} else {
		logCtx.Error("Invalid action on salesforce account sync.")
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
	var status []int
	for i := range documents {
		status = append(status, syncAccount(projectId, &documents[i]))
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

func Sync(projectId uint64) []Status {
	logCtx := log.WithField("project_id", projectId)

	statusByProjectAndType := make([]Status, 0, 0)

	documents, errCode := GetSalesforceDocumentsByTypeForSync(projectId, M.SalesforceDocumentTypeAccount)
	if errCode != http.StatusFound {
		logCtx.Error("Failed to get salesforce document by type for sync.")
		return statusByProjectAndType
	}
	status := &Status{
		ProjectId: projectId,
		Type:      M.SalesforceDocumentTypeNameAccount,
	}

	errCode = syncAll(projectId, documents)
	if errCode == http.StatusOK {
		status.Status = "success"
	} else {
		status.Status = "failures_seen"
	}

	return statusByProjectAndType
}
