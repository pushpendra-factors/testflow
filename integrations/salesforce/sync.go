package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"factors/util"
	"flag"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	SALESFORCE_DATA_SERVICE_ROUTE = "/services/data/"
	SALESFORCE_API_VERSION        = "v20.0"
	REDIRECT_URL                  = "https://api.factors.com/salesforce/auth"
	REFRESH_TOKEN_URL             = "https://login.salesforce.com/services/oauth2/token"
	BATCH_SIZE                    = 50
)

type NextSyncInfo struct {
	ProjectId    uint64
	RefreshToken string
	DocType      map[string]int64
	SyncAll      bool
	InstanceURL  string
}

// SalesforceSync client wrapper for salesforce api
type SalesforceSync struct {
	Client          *http.Client
	QueryURL        string
	AccessToken     string
	RedirectURL     string
	RefreshTokenURL string
	ClientSecret    string
	ClientId        string
	SyncInfo        []NextSyncInfo
}
type SalesforceProjectSettings struct {
	RefreshToken string `json:"refresh_token"`
	InstanceURL  string `json:"instance_url"`
}

type SalesforceSyncInfo struct {
	ProjectSettings map[uint64]*SalesforceProjectSettings `json:"project_settings"`
	LastSyncInfo    map[uint64]map[string]int64           `json:"last_sync_info"`
}

type SalesforceClient struct {
	Client          http.Client
	ClientSecret    string
	ClientId        string
	RefreshTokenURL string
	RedirectURL     string
	DataServiceHost string
	DryRun          bool
}

var sfClient *SalesforceClient

func NewSalesforceClient(clientId, clientSecret, redirectURL, env, dataServiceHost string) error {
	salesforceClient := &SalesforceClient{
		ClientId:        clientId,
		ClientSecret:    clientSecret,
		RedirectURL:     redirectURL,
		RefreshTokenURL: REFRESH_TOKEN_URL,
		DryRun:          env == "development",
		DataServiceHost: dataServiceHost,
	}

	salesforceClient.Client = http.Client{
		Timeout: 10 * time.Minute,
	}

	sfClient = salesforceClient
	return nil
}

func (sf *SalesforceClient) getNewAccessToken(refreshToken string) (string, error) {
	queryParams := fmt.Sprintf("grant_type=%s&refresh_token=%s&client_id=%s&client_secret=%s&redirect_uri=%s",
		"refresh_token", refreshToken, sf.ClientId, sf.ClientSecret, sf.RedirectURL)
	url := sf.RefreshTokenURL + "?" + queryParams
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("Accept", "application/json")

	resp, err := sf.Client.Do(req)
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

type QueryRespone struct {
	TotalSize      int      `json:"totalSize"`
	Done           bool     `json:"done"`
	Records        []record `json:"records"`
	NextRecordsUrl string   `json:"nextRecordsUrl"`
}

type Describe struct {
	Custom bool   `json:"custom"`
	Fields fields `json:"fields"`
}

func (sf *SalesforceClient) describe(object, accessToken, instanceURL string) (*Describe, error) {
	url := instanceURL + SALESFORCE_DATA_SERVICE_ROUTE + SALESFORCE_API_VERSION + "/sobjects/" + object + "/describe"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer"+" "+accessToken)
	req.Header.Add("Accept", "application/json")
	resp, err := sf.Client.Do(req)
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

func (sf *SalesforceClient) query(query, accessToken, instanceURL string) (*QueryRespone, error) {
	queryURL := instanceURL + SALESFORCE_DATA_SERVICE_ROUTE + SALESFORCE_API_VERSION + "/query?q=" + query
	fmt.Println("queryURL ", queryURL)
	req, err := http.NewRequest("GET", queryURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer"+" "+accessToken)
	req.Header.Add("Accept", "application/json")
	resp, err := sf.Client.Do(req)
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

func (sf *SalesforceClient) getSyncInfo() ([]NextSyncInfo, error) {
	var syncinfo SalesforceSyncInfo
	route := "/data_service/salesforce/documents/last_sync_info"
	url := sf.DataServiceHost + route
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := sfClient.Client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Error while requesting data")
	}

	err = json.NewDecoder(resp.Body).Decode(&syncinfo)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response")
	}

	var syncInfo []NextSyncInfo
	for pid, ps := range syncinfo.ProjectSettings {
		var nextSyncInfo NextSyncInfo
		nextSyncInfo.DocType = syncinfo.LastSyncInfo[pid]
		nextSyncInfo.ProjectId = pid
		nextSyncInfo.RefreshToken = ps.RefreshToken
		nextSyncInfo.InstanceURL = ps.InstanceURL
		syncInfo = append(syncInfo, nextSyncInfo)
	}

	return syncInfo, nil
}

func NewSalesforceSync(salesforceAppId, salesforceAppSecret string) *SalesforceSync {
	return &SalesforceSync{
		ClientId:     salesforceAppId,
		ClientSecret: salesforceAppSecret,
		Client:       &http.Client{},
	}
}

type fields []map[string]interface{}

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

type record map[string]interface{}

type Document struct {
	ProjectId uint64 `json:"project_id"`
	TypeAlias string `json:"type_alias"`
	Value     record `json:"value"`
}

func (sf *SalesforceClient) queryNextBatch(nextBatchRoute, InstanceURL string, accessToken string) (*QueryRespone, error) {
	url := InstanceURL + nextBatchRoute
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer"+" "+accessToken)
	req.Header.Add("Accept", "application/json")
	resp, err := sf.Client.Do(req)
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
func (sf *SalesforceClient) createDocument(projectId uint64, docType string, doc record) {
	route := "/data_service/salesforce/documents/add"
	url := sf.DataServiceHost + route
	fmt.Println("url ", url)
	document := &Document{
		ProjectId: projectId,
		TypeAlias: docType,
		Value:     doc,
	}

	body, err := json.Marshal(document)
	if err != nil {
		fmt.Printf("cannot marshal document: %w", err)
		return
	}

	fmt.Println("err3")
	if sf.DryRun {
		fmt.Println("Dry run, skip upsert")
		fmt.Println("Request body ", body)
		return
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		fmt.Printf("cannot create request: %w", err)
		return
	}
	res, err := sf.Client.Do(req)
	if res.StatusCode != http.StatusCreated {
		fmt.Println("err5 ", res.StatusCode)
		return
	}
	fmt.Println("Success creating document typ ", docType)
}

func syncByType(projectSyncInfo *NextSyncInfo, accessToken, object string) {
	description, err := sfClient.describe(object, accessToken, projectSyncInfo.InstanceURL)
	if err != nil {
		panic(err)
	}
	fields, err := getFieldsListFromDescription(description)
	if err != nil {
		panic(err)
	}
	selectStmnt := strings.Join(fields, ",")
	queryStmnt := fmt.Sprintf("SELECT+%s+FROM+%s", selectStmnt, object)
	queryRespone, err := sfClient.query(queryStmnt, accessToken, projectSyncInfo.InstanceURL)
	if err != nil {
		panic(err)
	}
	records := queryRespone.Records

	for i := range records {
		sfClient.createDocument(projectSyncInfo.ProjectId, object, records[i])
	}

	hasMore := !queryRespone.Done
	for hasMore {
		nextBatchRoute := queryRespone.NextRecordsUrl
		queryRespone, _ = sfClient.queryNextBatch(nextBatchRoute, projectSyncInfo.InstanceURL, accessToken)
		records = queryRespone.Records
		for i := range records {
			sfClient.createDocument(projectSyncInfo.ProjectId, object, records[i])
		}
		hasMore = queryRespone.Done
	}
}

func sync(projectSyncInfo NextSyncInfo) {
	accessToken, _ := sfClient.getNewAccessToken(projectSyncInfo.RefreshToken)
	for docType := range projectSyncInfo.DocType {
		syncByType(&projectSyncInfo, accessToken, docType)
	}
}

func main() {
	env := flag.String("env", "development", "")
	salesforceAppId := flag.String("salesforce_app_id", "", "")
	salesforceAppSecret := flag.String("salesforce_app_secret", "", "")
	dataServiceHost := flag.String("data_service_host", "http://localhost:8089", "")
	flag.Parse()

	if *salesforceAppId == "" || *salesforceAppSecret == "" {
		panic(fmt.Errorf("salesforce_app_secret or salesforce_app_secret not recognised"))
	}

	taskID := "Task#SalesforceSync"
	defer util.NotifyOnPanic(taskID, *env)

	err := NewSalesforceClient(*salesforceAppId, *salesforceAppSecret, "https://factors-dev.com:8080/integrations/salesforce/auth/callback", *env, *dataServiceHost)
	if err != nil {
		panic(fmt.Errorf("failed to create salesforce client "))
	}

	var syncInfo []NextSyncInfo
	syncInfo, err = sfClient.getSyncInfo()
	if err != nil {
		fmt.Println("Error while getting sync info ", err)
		return
	}
	for _, projectSyncInfo := range syncInfo {
		sync(projectSyncInfo)
	}
}
