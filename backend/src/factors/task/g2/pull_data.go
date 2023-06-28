package g2

import (
	"encoding/json"
	"errors"
	"factors/model/model"
	"factors/model/store"
	SDK "factors/sdk"
	U "factors/util"
	"fmt"
	"strings"
	"time"

	"net/http"

	log "github.com/sirupsen/logrus"
)

var defaultSyncInfo = map[string]int64{
	"event_stream": 0,
}

const NO_DATA_ERROR = "No data found"
const CATEGORIES_TAG_PREFIX = "categories/"

var tagEnum = map[string]string{
	"products.competitors":                        U.GROUP_EVENT_NAME_G2_ALTERNATIVE,
	"ad.category_competitor.product_left_sidebar": U.GROUP_EVENT_NAME_G2_SPONSORED,
	"comparisons.show":                            U.GROUP_EVENT_NAME_G2_COMPARISON,
	"categories.show":                             U.GROUP_EVENT_NAME_G2_CATEGORY,
	"categories.learn":                            U.GROUP_EVENT_NAME_G2_CATEGORY,
	"products.reviews":                            U.GROUP_EVENT_NAME_G2_PRODUCT_PROFILE,
	"products.features":                           U.GROUP_EVENT_NAME_G2_PRODUCT_PROFILE,
	"products.details":                            U.GROUP_EVENT_NAME_G2_PRODUCT_PROFILE,
	"products.pricing":                            U.GROUP_EVENT_NAME_G2_PRICING,
	"products.discussions":                        U.GROUP_EVENT_NAME_G2_PRODUCT_PROFILE,
	"products.discuss":                            U.GROUP_EVENT_NAME_G2_PRODUCT_PROFILE,
	"reports.show":                                U.GROUP_EVENT_NAME_G2_REPORT,
	"reports.preview":                             U.GROUP_EVENT_NAME_G2_REPORT,
	"reviewers.take_survey":                       U.GROUP_EVENT_NAME_G2_PRODUCT_PROFILE,
}

type EventStreamResponseStruct struct {
	Data  []EventStreamResponseDataFormat `json:"data"`
	Links map[string]string               `json:"links"`
}
type EventStreamResponseDataFormat struct {
	Attributes    map[string]interface{}                  `json:"attributes"`
	ID            string                                  `json:"id"`
	Type          string                                  `json:"type"`
	Relationships map[string]map[string]map[string]string `json:"relationships"`
}

func PerformETLForProject(projectSetting model.G2ProjectSettings) string {
	lastSyncInfo, errCode := store.GetStore().GetG2LastSyncInfo(projectSetting.ProjectID)
	if errCode != http.StatusOK {
		return "failed to get last sync info"
	}
	mapTypeAliasToLastSyncTimestamp := buildMapTypeAliasToLastSyncTimestamp(lastSyncInfo)
	for typeAlias, lastSync := range mapTypeAliasToLastSyncTimestamp {
		data, err := extractData(lastSync, projectSetting.IntG2APIKey)
		if err != nil {
			return err.Error()
		}
		if len(data) == 0 {
			return NO_DATA_ERROR
		}
		transformedData, err := transformData(data)
		if err != nil {
			return err.Error()
		}
		err = buildDocumentAndInsertData(projectSetting.ProjectID, typeAlias, transformedData)
		if err != nil {
			return err.Error()
		}
	}
	return ""
}

// Todo: Add retries
func extractData(lastSync int64, apiKey string) ([]EventStreamResponseDataFormat, error) {
	data := make([]EventStreamResponseDataFormat, 0)
	from, to := getFromAndToTimestampInReqFormat(lastSync)
	url := fmt.Sprintf("https://data.g2.com/api/v1/ahoy/remote-event-streams?page[size]=25&&filter[start_time]=%s&&filter[end_time]=%s", from, to)
	response, err := fetchEventStreamDataFromAPI(url, apiKey)
	if err != nil {
		return make([]EventStreamResponseDataFormat, 0), err
	} else {
		data = append(data, response.Data...)
	}
	fetchNextBatch := true
	if len(response.Data) == 0 {
		fetchNextBatch = false
	}
	for fetchNextBatch {
		response, err = fetchEventStreamDataFromAPI(response.Links["next"], apiKey)
		if err != nil {
			return make([]EventStreamResponseDataFormat, 0), err
		} else {
			data = append(data, response.Data...)
		}
		if len(response.Data) == 0 {
			fetchNextBatch = false
		}
	}
	return data, nil
}

func transformData(data []EventStreamResponseDataFormat) ([]map[string]interface{}, error) {
	transformedData := make([]map[string]interface{}, 0)
	for _, value := range data {
		newRow := value.Attributes
		newRow["id"] = value.ID
		newRow["type"] = value.Type
		newRow["company_url"] = value.Relationships["company"]["links"]["related"]
		stringTime := fmt.Sprintf("%v", newRow["time"])
		timestamp, err := time.Parse(time.RFC3339, stringTime)
		if err != nil {
			return make([]map[string]interface{}, 0), err
		}
		newRow["timestamp"] = timestamp.Unix()
		transformedData = append(transformedData, newRow)
	}
	return transformedData, nil
}

func buildDocumentAndInsertData(projectID int64, typeAlias string, data []map[string]interface{}) error {
	documents := make([]model.G2Document, 0)
	for _, value := range data {
		valueJsonb, err := U.EncodeStructTypeToPostgresJsonb(value)
		if err != nil {
			return err
		}
		id := fmt.Sprintf("%v", value["id"])
		timestamp, ok := value["timestamp"].(int64)
		if !ok {
			return errors.New("Failed to convert timestamp interface")
		}
		document := model.G2Document{
			ProjectID: projectID,
			ID:        id,
			TypeAlias: typeAlias,
			Timestamp: timestamp,
			Value:     valueJsonb,
			Synced:    false,
		}
		documents = append(documents, document)
	}
	errCode := store.GetStore().CreateMultipleG2Document(documents)
	if errCode != http.StatusCreated {
		return errors.New("Failed to insert documents to db")
	}
	return nil
}

func fetchEventStreamDataFromAPI(url string, apiKey string) (EventStreamResponseStruct, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return EventStreamResponseStruct{}, err
	}
	authToken := fmt.Sprintf("Token token=%s", apiKey)
	req.Header.Add("Authorization", authToken)
	resp, err := client.Do(req)
	if err != nil {
		return EventStreamResponseStruct{}, err
	}

	var jsonResponse EventStreamResponseStruct
	err = json.NewDecoder(resp.Body).Decode(&jsonResponse)
	if err != nil {
		return EventStreamResponseStruct{}, err
	}
	return jsonResponse, nil
}

func getFromAndToTimestampInReqFormat(lastSync int64) (string, string) {
	var startTime string
	if lastSync == 0 {
		startTime = time.Now().Add(time.Hour * -24).Format(time.RFC3339)
	} else {
		startTime = time.Unix(lastSync, 0).Format(time.RFC3339)
	}
	endTime := time.Now().Add(time.Hour * -1).Format(time.RFC3339)
	return startTime, endTime
}

func buildMapTypeAliasToLastSyncTimestamp(lastSyncInfo []model.G2LastSyncInfo) map[string]int64 {
	mapTypeAliasToLastSyncTimestamp := defaultSyncInfo

	for _, info := range lastSyncInfo {
		mapTypeAliasToLastSyncTimestamp[info.TypeAlias] = info.Timestamp
	}
	return mapTypeAliasToLastSyncTimestamp
}

type CompanyAPIResponseStruct struct {
	Data CompanyAPIResponseFormat `json:"data"`
}

type CompanyAPIResponseFormat struct {
	ID         string                  `json:"id"`
	Attributes CompanyAttributesStruct `json:"attributes"`
}

type CompanyAttributesStruct struct {
	Name           string `json:"name"`
	LegalName      string `json:"legal_name"`
	Employees      int64  `json:"employees"`
	EmployeesRange string `json:"employees_range"`
	Country        string `json:"country"`
	Domain         string `json:"domain"`
}

var defaultPropertiesForEvent = map[string]interface{}{
	U.EP_SKIP_SESSION: U.PROPERTY_VALUE_TRUE,
}

func PerformCompanyEnrichmentAndUserAndEventCreationForProject(projectSetting model.G2ProjectSettings) (string, int) {
	g2Documents, errCode := store.GetStore().GetG2DocumentsForGroupUserCreation(projectSetting.ProjectID)
	if errCode != http.StatusOK {
		return "Failed to get documents for company enrichment", errCode
	}
	projectID := projectSetting.ProjectID
	eventNameG2All, errCode := store.GetStore().CreateOrGetUserCreatedEventName(&model.EventName{
		ProjectId: projectID,
		Name:      U.GROUP_EVENT_NAME_G2_ALL,
	})
	if errCode != http.StatusCreated && errCode != http.StatusConflict {
		return "Failed in creating all pageview event name", errCode
	}

	for _, g2Document := range g2Documents {
		logFields := log.Fields{
			"project_id": projectID,
			"doument":    g2Document,
		}
		defer model.LogOnSlowExecutionWithParams(time.Now(), &logFields)
		logCtx := log.WithFields(logFields)
		valueMap := make(map[string]interface{})
		err := U.DecodePostgresJsonbToStructType(g2Document.Value, &valueMap)
		if err != nil {
			return err.Error(), http.StatusInternalServerError
		}

		companyURL := fmt.Sprintf("%v", valueMap["company_url"])

		client := &http.Client{}
		req, err := http.NewRequest("GET", companyURL, nil)
		if err != nil {
			return err.Error(), req.Response.StatusCode
		}
		authToken := fmt.Sprintf("Token token=%s", projectSetting.IntG2APIKey)
		req.Header.Add("Authorization", authToken)
		resp, err := client.Do(req)
		if err != nil {
			return err.Error(), resp.StatusCode
		}

		var jsonResponse CompanyAPIResponseStruct
		err = json.NewDecoder(resp.Body).Decode(&jsonResponse)
		if err != nil {
			return err.Error(), http.StatusInternalServerError
		}
		if jsonResponse.Data.Attributes.Domain == "" {
			logCtx.Error("Ashhar - no domain given in the API") // to be removed after testing
			err = store.GetStore().UpdateG2GroupUserCreationDetails(g2Document)
			if err != nil {
				logCtx.WithError(err).Error("Failed in updating user creation details")
				return "Failed in updating user creation details", http.StatusInternalServerError
			}
			continue
		}
		userPropertiesMap := U.PropertiesMap{
			U.G2_COMPANY_ID:      jsonResponse.Data.ID,
			U.G2_COUNTRY:         jsonResponse.Data.Attributes.Country,
			U.G2_DOMAIN:          jsonResponse.Data.Attributes.Domain,
			U.G2_LEGAL_NAME:      jsonResponse.Data.Attributes.LegalName,
			U.G2_NAME:            jsonResponse.Data.Attributes.Name,
			U.G2_EMPLOYEES:       jsonResponse.Data.Attributes.Employees,
			U.G2_EMPLOYEES_RANGE: jsonResponse.Data.Attributes.EmployeesRange,
		}
		userID, errCode := SDK.TrackGroupWithDomain(projectID, U.GROUP_NAME_G2, jsonResponse.Data.Attributes.Domain, userPropertiesMap, g2Document.Timestamp)
		if errCode != http.StatusOK {
			logCtx.Error("Failed in TrackGroupWithDomain")
			return "Failed in TrackGroupWithDomain", errCode
		}
		eventName, errMsg, errCode := getEventNameFromTag(valueMap)
		if errMsg != "" {
			return errMsg, errCode
		}

		eventNameG2, errCode := store.GetStore().CreateOrGetUserCreatedEventName(&model.EventName{
			ProjectId: projectID,
			Name:      eventName,
		})
		if errCode != http.StatusCreated && errCode != http.StatusConflict {
			return "Failed in creating pageview event name", errCode
		}
		properties := getEventPropertiesFromValueMap(valueMap)
		propertiesJSONB, err := U.EncodeStructTypeToPostgresJsonb(&properties)
		if err != nil {
			logCtx.WithError(err).Error("Failed in encoding properties to JSONb")
			return "Failed in encoding properties to JSONb", http.StatusInternalServerError
		}
		allPageviewEvent := model.Event{
			EventNameId: eventNameG2All.ID,
			Timestamp:   g2Document.Timestamp,
			ProjectId:   projectID,
			UserId:      userID,
			Properties:  *propertiesJSONB,
		}
		_, errCode = store.GetStore().CreateEvent(&allPageviewEvent)
		if errCode != http.StatusCreated {
			logCtx.Error("Failed in creating all pageview event")
			return "Failed in creating all pageview event", errCode
		}

		g2PageviewEvent := model.Event{
			EventNameId: eventNameG2.ID,
			Timestamp:   g2Document.Timestamp,
			ProjectId:   projectID,
			UserId:      userID,
			Properties:  *propertiesJSONB,
		}
		_, errCode = store.GetStore().CreateEvent(&g2PageviewEvent)
		if errCode != http.StatusCreated {
			logCtx.Error("Failed in creating g2 pageview event")
			return "Failed in creating g2 pageview event", errCode
		}

		err = store.GetStore().UpdateG2GroupUserCreationDetails(g2Document)
		if err != nil {
			logCtx.WithError(err).Error("Failed in updating user creation details")
			return "Failed in updating user creation details", http.StatusInternalServerError
		}

	}
	return "", http.StatusOK
}

func getEventNameFromTag(valueMap map[string]interface{}) (string, string, int) {
	if _, exists := valueMap["tag"]; !exists {
		return "", "Failed in tag not exists for eventname creation", http.StatusBadRequest
	}
	tag := fmt.Sprintf("%v", valueMap["tag"])
	eventName, exists := tagEnum[tag]
	if !exists {
		if strings.HasPrefix(tag, CATEGORIES_TAG_PREFIX) {
			eventName = U.GROUP_EVENT_NAME_G2_CATEGORY
		} else {
			errMsg := "Tag " + tag + " is not present is enum"
			return "", errMsg, http.StatusBadRequest
		}
	}
	return eventName, "", http.StatusOK
}

func getEventPropertiesFromValueMap(valueMap map[string]interface{}) map[string]interface{} {
	properties := defaultPropertiesForEvent
	tag := fmt.Sprintf("%v", valueMap["tag"])
	title := fmt.Sprintf("%v", valueMap["title"])
	page_url := fmt.Sprintf("%v", valueMap["url"])
	properties[U.EP_PAGE_TITLE] = title
	properties[U.EP_PAGE_URL] = page_url
	properties[U.EP_G2_TAG] = tag
	return properties
}
