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
	M "factors/model"
	U "factors/util"

	log "github.com/sirupsen/logrus"
)

const (
	SALESFORCE_INSTANCE_URL       = "instance_url"
	SALESFORCE_DATA_SERVICE_ROUTE = "/services/data/"
	SALESFORCE_API_VERSION        = "v20.0"
)

//Salesforce API structs
type field map[string]interface{}

type Describe struct {
	Custom bool    `json:"custom"`
	Fields []field `json:"fields"`
}

type QueryRespone struct {
	TotalSize      int                  `json:"totalSize"`
	Done           bool                 `json:"done"`
	Records        []M.SalesforceRecord `json:"records"`
	NextRecordsUrl string               `json:"nextRecordsUrl"`
}

type SalesforceObjectStatus struct {
	ProjetID     uint64   `json:"project_id"`
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

func getCustomerUserIDFromProperties(projectID uint64, properties map[string]interface{}) string {

	if phoneNo, ok := properties["MobilePhone"].(string); ok && phoneNo != "" {
		for pPhoneNo := range U.GetPossiblePhoneNumber(phoneNo) {
			_, errCode := M.GetUserLatestByCustomerUserId(projectID, pPhoneNo)
			if errCode == http.StatusFound {
				return pPhoneNo
			}
		}
		return phoneNo
	}

	if phoneNo, ok := properties["Phone"].(string); ok && phoneNo != "" {
		for pPhoneNo := range U.GetPossiblePhoneNumber(phoneNo) {
			_, errCode := M.GetUserLatestByCustomerUserId(projectID, pPhoneNo)
			if errCode == http.StatusFound {
				return pPhoneNo
			}
		}
		return phoneNo
	}

	return ""
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

func syncByType(ps *M.SalesforceProjectSettings, accessToken, objectName, dateTime string) (SalesforceObjectStatus, error) {
	var salesforceObjectStatus SalesforceObjectStatus
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
			err = M.BuildAndUpsertDocument(ps.ProjectID, objectName, records[i])
			if err != nil {
				logCtx.WithError(err).Error("Failed to BuildAndUpsertDocument.")
				failures = append(failures, err.Error())
			}
		}

		salesforceObjectStatus.Failures = append(salesforceObjectStatus.Failures, failures...)
		hasMore = !queryResponse.Done
		nextBatchRoute = queryResponse.NextRecordsUrl
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
				"project_id": ps.ProjectID,
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
