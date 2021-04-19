package tests

import (
	"bytes"
	"encoding/json"
	C "factors/config"
	Const "factors/constants"
	H "factors/handler"
	"factors/handler/helpers"
	"factors/model/model"
	M "factors/model/model"
	"factors/model/store"
	U "factors/util"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func sendCreateFacebookDocumentReq(r *gin.Engine, project_id uint64, customerAccountID string, valueJSON *postgres.Jsonb, id string, timestamp int64, type_alias string) *httptest.ResponseRecorder {
	payload := map[string]interface{}{
		"project_id":             project_id,
		"customer_ad_account_id": customerAccountID,
		"type_alias":             type_alias,
		"id":                     id,
		"value":                  valueJSON,
		"timestamp":              timestamp,
		"platform":               "facebook",
	}

	rb := U.NewRequestBuilder(http.MethodPost, "http://localhost:8089/data_service/facebook/documents/add").
		WithPostParams(payload)

	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error sending facebook document add requests to data server.")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func sendChannelAnalyticsQueryReq(r *gin.Engine, project_id uint64, agent *M.Agent, channelQueryJSON map[string]interface{}) *httptest.ResponseRecorder {
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error Creating cookieData")
	}

	url := fmt.Sprintf("/projects/%d/v1/channels/query", project_id)
	rb := C.NewRequestBuilderWithPrefix(http.MethodPost, url).
		WithPostParams(channelQueryJSON).WithCookie(&http.Cookie{
		Name:   C.GetFactorsCookieName(),
		Value:  cookieData,
		MaxAge: 1000,
	})

	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error sending channel query request to app server.")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

type resultStruct struct {
	Result M.ChannelResultGroupV1 `json:"result"`
}

func TestExecuteChannelQueryHandlerForFacebook(t *testing.T) {
	r := gin.Default()
	H.InitDataServiceRoutes(r)

	a := gin.Default()
	H.InitAppRoutes(a)
	Const.SetSmartPropertiesReservedNames()

	//inserting sample data in facebook, also testing data service endpoint facebook/documents/add
	project, agent, _ := SetupProjectWithAgentDAO()
	assert.NotNil(t, project)
	customerAccountID1 := U.RandomNumericString(10)
	customerAccountID2 := U.RandomNumericString(10)
	customerAccountID3 := U.RandomNumericString(10)
	_, errCode := store.GetStore().UpdateProjectSettings(project.ID, &M.ProjectSetting{
		IntFacebookAdAccount: customerAccountID1 + "," + customerAccountID2,
	})
	assert.Equal(t, http.StatusAccepted, errCode)

	campaignID1 := U.RandomNumericStringNonZeroStart(8)
	campaignID1Float, _ := strconv.ParseFloat(campaignID1, 64)
	value := map[string]interface{}{"spend": "100", "clicks": "50", "campaign_id": campaignID1, "impressions": "1000", "campaign_name": "Campaign_1", "account_currency": "USD"}
	valueJSON, err := U.EncodeToPostgresJsonb(&value)
	assert.Nil(t, err)

	w := sendCreateFacebookDocumentReq(r, project.ID, customerAccountID1, valueJSON, campaignID1, 20210205, "campaign_insights")
	assert.Equal(t, http.StatusCreated, w.Code)

	campaignID2 := U.RandomNumericString(8)
	// campaignID2Float, _ := strconv.ParseFloat(campaignID2, 64)
	value = map[string]interface{}{"spend": "200", "clicks": "100", "campaign_id": campaignID2, "impressions": "2000", "campaign_name": "Campaign_2"}
	valueJSON, err = U.EncodeToPostgresJsonb(&value)
	assert.Nil(t, err)
	w = sendCreateFacebookDocumentReq(r, project.ID, customerAccountID1, valueJSON, campaignID2, 20210206, "campaign_insights")
	assert.Equal(t, http.StatusCreated, w.Code)

	adgroupID1_1 := U.RandomNumericString(8)
	// adgroupID1_1Float, _ := strconv.ParseFloat(adgroupID1_1, 64)
	value = map[string]interface{}{"spend": "30", "clicks": "30", "adset_id": adgroupID1_1, "adset_name": "Adgroup_1_1", "campaign_id": campaignID1, "impressions": "600", "campaign_name": "Campaign_1", "account_currency": "USD"}
	valueJSON, err = U.EncodeToPostgresJsonb(&value)
	assert.Nil(t, err)
	w = sendCreateFacebookDocumentReq(r, project.ID, customerAccountID2, valueJSON, adgroupID1_1, 20210205, "ad_set_insights")
	assert.Equal(t, http.StatusCreated, w.Code)

	adgroupID1_2 := U.RandomNumericString(8)
	// adgroupID1_2Float, _ := strconv.ParseFloat(adgroupID1_2, 64)
	value = map[string]interface{}{"spend": "70", "clicks": "20", "adset_id": adgroupID1_2, "adset_name": "Adgroup_1_2", "campaign_id": campaignID1, "impressions": "400", "campaign_name": "Campaign_1", "account_currency": "USD"}
	valueJSON, err = U.EncodeToPostgresJsonb(&value)
	assert.Nil(t, err)
	w = sendCreateFacebookDocumentReq(r, project.ID, customerAccountID2, valueJSON, adgroupID1_2, 20210205, "ad_set_insights")
	assert.Equal(t, http.StatusCreated, w.Code)

	adgroupID2_1 := U.RandomNumericString(8)
	// adgroupID2_1Float, _ := strconv.ParseFloat(adgroupID2_1, 64)
	value = map[string]interface{}{"spend": "120", "clicks": "25", "adset_id": adgroupID2_1, "adset_name": "Adgroup_2_1", "campaign_id": campaignID2, "impressions": "1500", "campaign_name": "Campaign_2", "account_currency": "USD"}
	valueJSON, err = U.EncodeToPostgresJsonb(&value)
	assert.Nil(t, err)
	w = sendCreateFacebookDocumentReq(r, project.ID, customerAccountID3, valueJSON, adgroupID2_1, 20210206, "ad_set_insights")
	assert.Equal(t, http.StatusCreated, w.Code)

	adgroupID2_2 := U.RandomNumericString(8)
	// adgroupID2_2Float, _ := strconv.ParseFloat(adgroupID2_2, 64)
	value = map[string]interface{}{"spend": "80", "clicks": "75", "adset_id": adgroupID1_2, "adset_name": "Adgroup_2_2", "campaign_id": campaignID2, "impressions": "500", "campaign_name": "Campaign_2", "account_currency": "USD"}
	valueJSON, err = U.EncodeToPostgresJsonb(&value)
	assert.Nil(t, err)
	w = sendCreateFacebookDocumentReq(r, project.ID, customerAccountID2, valueJSON, adgroupID2_2, 20210206, "ad_set_insights")
	assert.Equal(t, http.StatusCreated, w.Code)

	//channel query test
	// filters: campaignName contains '1' & adGroupName contains '1_1', groupBy: campaignID, adGroupName
	channelQuery := map[string]interface{}{"query_group": [1]map[string]interface{}{{"channel": "facebook_ads", "select_metrics": [3]string{"clicks", "impressions", "spend"},
		"group_by": [2]map[string]interface{}{{"name": "campaign", "property": "id"}, {"name": "ad_group", "property": "name"}},
		"filters":  [2]map[string]interface{}{{"name": "campaign", "property": "name", "condition": "contains", "logical_operator": "AND", "value": "1"}, {"name": "ad_group", "property": "name", "condition": "contains", "logical_operator": "AND", "value": "1_1"}},
		"gbt":      "", "fr": 1612314000, "to": 1612746000}}, "cl": "channel_v1"}

	w = sendChannelAnalyticsQueryReq(a, project.ID, agent, channelQuery)
	assert.Equal(t, http.StatusOK, w.Code)

	var result resultStruct
	decoder := json.NewDecoder(w.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&result); err != nil {
		assert.NotNil(t, nil, err)
	}
	assert.Equal(t, len(result.Result.Results[0].Headers), 5)
	assert.Equal(t, len(result.Result.Results[0].Rows), 1)
	assert.Equal(t, len(result.Result.Results[0].Rows[0]), 5)
	if C.UseMemSQLDatabaseStore() {
		assert.Equal(t, result.Result.Results[0].Rows[0][0], campaignID1)
	} else {
		assert.Equal(t, result.Result.Results[0].Rows[0][0], campaignID1Float)
	}
	assert.Equal(t, result.Result.Results[0].Rows[0][1], "Adgroup_1_1")
	assert.Equal(t, result.Result.Results[0].Rows[0][2], float64(30))
	assert.Equal(t, result.Result.Results[0].Rows[0][3], float64(600))
	assert.Equal(t, result.Result.Results[0].Rows[0][4], float64(30))

	// filters : campaignID equals campaignID1, adGroupName contains 1_1
	channelQuery = map[string]interface{}{"query_group": [1]map[string]interface{}{{"channel": "facebook_ads", "select_metrics": [3]string{"clicks", "impressions", "spend"},
		"filters": [2]map[string]interface{}{{"name": "campaign", "property": "id", "condition": "equals", "logical_operator": "AND", "value": campaignID1}, {"name": "ad_group", "property": "name", "condition": "contains", "logical_operator": "AND", "value": "1_1"}},
		"gbt":     "", "fr": 1612314000, "to": 1612746000}}, "cl": "channel_v1"}

	w = sendChannelAnalyticsQueryReq(a, project.ID, agent, channelQuery)
	assert.Equal(t, http.StatusOK, w.Code)

	var result1 resultStruct
	decoder = json.NewDecoder(w.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&result1); err != nil {
		assert.NotNil(t, nil, err)
	}
	assert.Equal(t, len(result1.Result.Results[0].Headers), 3)
	assert.Equal(t, len(result1.Result.Results[0].Rows), 1)
	assert.Equal(t, len(result1.Result.Results[0].Rows[0]), 3)
	assert.Equal(t, result1.Result.Results[0].Rows[0][0], float64(30))
	assert.Equal(t, result1.Result.Results[0].Rows[0][1], float64(600))
	assert.Equal(t, result1.Result.Results[0].Rows[0][2], float64(30))

	// filters : campaignID equals campaignID1, adGroupName contains 2_1, result should be 0 in result rows
	channelQuery = map[string]interface{}{"query_group": [1]map[string]interface{}{{"channel": "facebook_ads", "select_metrics": [3]string{"clicks", "impressions", "spend"},
		"filters": [2]map[string]interface{}{{"name": "campaign", "property": "id", "condition": "equals", "logical_operator": "AND", "value": campaignID1}, {"name": "ad_group", "property": "name", "condition": "contains", "logical_operator": "AND", "value": "2_1"}},
		"gbt":     "", "fr": 1612314000, "to": 1612746000}}, "cl": "channel_v1"}

	w = sendChannelAnalyticsQueryReq(a, project.ID, agent, channelQuery)
	assert.Equal(t, http.StatusOK, w.Code)

	var result2 resultStruct
	decoder = json.NewDecoder(w.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&result2); err != nil {
		assert.NotNil(t, nil, err)
	}
	assert.Equal(t, len(result2.Result.Results[0].Headers), 3)
	assert.Equal(t, len(result2.Result.Results[0].Rows), 1)
	assert.Equal(t, len(result2.Result.Results[0].Rows[0]), 3)
	assert.Equal(t, result2.Result.Results[0].Rows[0][0], float64(0))
	assert.Equal(t, result2.Result.Results[0].Rows[0][1], float64(0))
	assert.Equal(t, result2.Result.Results[0].Rows[0][2], float64(0))

	//groupBy: campaignName, adGroupID
	channelQuery = map[string]interface{}{"query_group": [1]map[string]interface{}{{"channel": "facebook_ads", "select_metrics": [3]string{"clicks", "impressions", "spend"},
		"group_by": [2]map[string]interface{}{{"name": "campaign", "property": "name"}, {"name": "ad_group", "property": "id"}},
		"gbt":      "", "fr": 1612314000, "to": 1612746000}}, "cl": "channel_v1"}

	w = sendChannelAnalyticsQueryReq(a, project.ID, agent, channelQuery)
	assert.Equal(t, http.StatusOK, w.Code)

	var result3 resultStruct
	decoder = json.NewDecoder(w.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&result3); err != nil {
		assert.NotNil(t, nil, err)
	}
	assert.Equal(t, len(result3.Result.Results[0].Headers), 5)
	assert.Equal(t, len(result3.Result.Results[0].Rows), 3)
	assert.Equal(t, len(result3.Result.Results[0].Rows[0]), 5)
	assert.Equal(t, result3.Result.Results[0].Rows[0][2], float64(75))
	assert.Equal(t, result3.Result.Results[0].Rows[0][3], float64(500))
	assert.Equal(t, result3.Result.Results[0].Rows[0][4], float64(80))

	assert.Equal(t, result3.Result.Results[0].Rows[1][2], float64(30))
	assert.Equal(t, result3.Result.Results[0].Rows[1][3], float64(600))
	assert.Equal(t, result3.Result.Results[0].Rows[1][4], float64(30))

	assert.Equal(t, result3.Result.Results[0].Rows[2][2], float64(20))
	assert.Equal(t, result3.Result.Results[0].Rows[2][3], float64(400))
	assert.Equal(t, result3.Result.Results[0].Rows[2][4], float64(70))
}

type MyStruct struct {
	id     int64           `json:"id"`
	column string          `json:"column"`
	value  *postgres.Jsonb `json:"value"`
}

func TestChannelQueryHandlerForAdwords(t *testing.T) {
	r := gin.Default()
	H.InitDataServiceRoutes(r)
	Const.SetSmartPropertiesReservedNames()

	a := gin.Default()
	H.InitAppRoutes(a)

	project, customerAccountID, agent, statusCode := createProjectAndAddAdwordsDocument(t, r)
	if statusCode != http.StatusAccepted {
		assert.Equal(t, false, true)
		return
	}

	adwordsDocuments := []M.AdwordsDocument{
		{ID: "1", Timestamp: 20210201, ProjectID: project.ID, CustomerAccountID: customerAccountID, TypeAlias: "campaign_performance_report",
			Value: &postgres.Jsonb{json.RawMessage(`{"cost": "11","clicks": "101","campaign_id":"1","impressions": "1001", "campaign_name": "test1"}`)}},
		{ID: "1", Timestamp: 20210202, ProjectID: project.ID, CustomerAccountID: customerAccountID, TypeAlias: "campaign_performance_report",
			Value: &postgres.Jsonb{json.RawMessage(`{"cost": "11","clicks": "101","campaign_id":"1","impressions": "1001", "campaign_name": "test1"}`)}},
		{ID: "2", Timestamp: 20210201, ProjectID: project.ID, CustomerAccountID: customerAccountID, TypeAlias: "campaign_performance_report",
			Value: &postgres.Jsonb{json.RawMessage(`{"cost": "12","clicks": "102","campaign_id":"2","impressions": "1002", "campaign_name": "test2"}`)}},
		{ID: "2", Timestamp: 20210202, ProjectID: project.ID, CustomerAccountID: customerAccountID, TypeAlias: "campaign_performance_report",
			Value: &postgres.Jsonb{json.RawMessage(`{"cost": "12","clicks": "102","campaign_id":"2","impressions": "1002", "campaign_name": "test2"}`)}},

		{ID: "11", Timestamp: 20210201, ProjectID: project.ID, CustomerAccountID: customerAccountID, TypeAlias: "ad_group_performance_report",
			Value: &postgres.Jsonb{json.RawMessage(`{"cost": "11","clicks": "101","campaign_id":"1","impressions": "1001", "campaign_name":"test1","ad_group_id":"11","ad_group_name":"agtest1", "search_click_share":"10%", "total_search_click":"10010"}`)}},
		{ID: "11", Timestamp: 20210202, ProjectID: project.ID, CustomerAccountID: customerAccountID, TypeAlias: "ad_group_performance_report",
			Value: &postgres.Jsonb{json.RawMessage(`{"cost": "11","clicks": "101","campaign_id":"1","impressions": "1001", "campaign_name": "test1","ad_group_id":"11","ad_group_name":"agtest1", "search_click_share":"10%", "total_search_click":"10010"}`)}},
		{ID: "12", Timestamp: 20210201, ProjectID: project.ID, CustomerAccountID: customerAccountID, TypeAlias: "ad_group_performance_report",
			Value: &postgres.Jsonb{json.RawMessage(`{"cost": "12","clicks": "102","campaign_id":"1","impressions": "1002", "campaign_name": "test1","ad_group_id":"12","ad_group_name":"agtest2","status":"paused"}`)}},
		{ID: "12", Timestamp: 20210202, ProjectID: project.ID, CustomerAccountID: customerAccountID, TypeAlias: "ad_group_performance_report",
			Value: &postgres.Jsonb{json.RawMessage(`{"cost": "12","clicks": "102","campaign_id":"1","impressions": "1002", "campaign_name": "test1","ad_group_id":"12","ad_group_name":"agtest2","status":"enabled"}`)}},
		{ID: "21", Timestamp: 20210201, ProjectID: project.ID, CustomerAccountID: customerAccountID, TypeAlias: "ad_group_performance_report",
			Value: &postgres.Jsonb{json.RawMessage(`{"cost": "13","clicks": "103","campaign_id":"2","impressions": "1002", "campaign_name": "test2","ad_group_id":"21","ad_group_name":"agtest3","status":"enabled"}`)}},
		{ID: "21", Timestamp: 20210202, ProjectID: project.ID, CustomerAccountID: customerAccountID, TypeAlias: "ad_group_performance_report",
			Value: &postgres.Jsonb{json.RawMessage(`{"cost": "13","clicks": "103","campaign_id":"2","impressions": "1002", "campaign_name": "test2","ad_group_id":"21","ad_group_name":"agtest3"}`)}},

		{ID: "111", Timestamp: 20210201, ProjectID: project.ID, CustomerAccountID: customerAccountID, TypeAlias: "keyword_performance_report",
			Value: &postgres.Jsonb{json.RawMessage(`{"cost": "11","clicks": "101","campaign_id":"1","impressions": "1001", "campaign_name": "test1", "ad_group_id":"11","ad_group_name":"agtest1","id":"111","quality_score":"0.1"}`)}},
		{ID: "111", Timestamp: 20210202, ProjectID: project.ID, CustomerAccountID: customerAccountID, TypeAlias: "keyword_performance_report",
			Value: &postgres.Jsonb{json.RawMessage(`{"cost": "11","clicks": "101","campaign_id":"1","impressions": "1001", "campaign_name": "test1", "ad_group_id":"11","ad_group_name":"agtest1","id":"111"}`)}},
		{ID: "121", Timestamp: 20210201, ProjectID: project.ID, CustomerAccountID: customerAccountID, TypeAlias: "keyword_performance_report",
			Value: &postgres.Jsonb{json.RawMessage(`{"cost": "12","clicks": "102","campaign_id":"1","impressions": "1002", "campaign_name": "test1", "ad_group_id":"12","ad_group_name":"agtest2","id":"121", "quality_score":"0.2"}`)}},
		{ID: "121", Timestamp: 20210202, ProjectID: project.ID, CustomerAccountID: customerAccountID, TypeAlias: "keyword_performance_report",
			Value: &postgres.Jsonb{json.RawMessage(`{"cost": "12","clicks": "102","campaign_id":"1","impressions": "1002", "campaign_name": "test1", "ad_group_id":"12","ad_group_name":"agtest2","id":"121","quality_score":"0.2"}`)}},
		{ID: "211", Timestamp: 20210201, ProjectID: project.ID, CustomerAccountID: customerAccountID, TypeAlias: "keyword_performance_report",
			Value: &postgres.Jsonb{json.RawMessage(`{"cost": "13","clicks": "103","campaign_id":"2","impressions": "1002", "campaign_name": "test2", "ad_group_id":"21","ad_group_name":"agtest3","id":"211"}`)}},
		{ID: "211", Timestamp: 20210202, ProjectID: project.ID, CustomerAccountID: customerAccountID, TypeAlias: "keyword_performance_report",
			Value: &postgres.Jsonb{json.RawMessage(`{"cost": "13","clicks": "103","campaign_id":"2","impressions": "1002", "campaign_name": "test2", "ad_group_id":"21","ad_group_name":"agtest3","id":"211"}`)}},
	}
	for _, adwordsDocument := range adwordsDocuments {
		status := store.GetStore().CreateAdwordsDocument(&adwordsDocument)
		assert.Equal(t, http.StatusCreated, status)
	}

	/*
		Set 1: Simple.
		Set 2: FilterBy.
		Set 3: GroupBy.
		Set 4: SelectMetrics.
		Set 5: GroupBy and FilterBy.
		Set 6: Fields of GroupBy or FilterBy which are not common. To merge later into proper set.
	*/
	successChannelQueries := []map[string]interface{}{
		{"query_group": [1]map[string]interface{}{{"channel": "google_ads", "select_metrics": [3]string{"clicks", "impressions", "spend"},
			"gbt": "", "fr": 1611964800, "to": 1612310400}}, "cl": "channel_v1"},

		{"query_group": [1]map[string]interface{}{{"channel": "google_ads", "select_metrics": [3]string{"clicks", "impressions", "spend"},
			"filters": [2]map[string]interface{}{{"name": "campaign", "property": "name", "condition": "contains", "logical_operator": "AND", "value": "1"}, {"name": "ad_group", "property": "name", "condition": "contains", "logical_operator": "AND", "value": "1"}},
			"gbt":     "", "fr": 1611964800, "to": 1612310400}}, "cl": "channel_v1"},
		{"query_group": [1]map[string]interface{}{{"channel": "google_ads", "select_metrics": [3]string{"clicks", "impressions", "spend"},
			"filters": [2]map[string]interface{}{{"name": "campaign", "property": "name", "condition": "contains", "logical_operator": "AND", "value": "2"}, {"name": "campaign", "property": "id", "condition": "equals", "logical_operator": "AND", "value": "2"}},
			"gbt":     "", "fr": 1611964800, "to": 1612310400}}, "cl": "channel_v1"},
		{"query_group": [1]map[string]interface{}{{"channel": "google_ads", "select_metrics": [3]string{"clicks", "impressions", "spend"},
			"filters": [2]map[string]interface{}{{"name": "campaign", "property": "name", "condition": "contains", "logical_operator": "AND", "value": "2"}, {"name": "campaign", "property": "id", "condition": "equals", "logical_operator": "AND", "value": "2"}},
			"gbt":     "", "fr": 1611964800, "to": 1612310400}}, "cl": "channel_v1"},
		{"query_group": [1]map[string]interface{}{{"channel": "google_ads", "select_metrics": [3]string{"clicks", "impressions", "spend"},
			"filters": [2]map[string]interface{}{{"name": "ad_group", "property": "name", "condition": "contains", "logical_operator": "AND", "value": "2"}, {"name": "campaign", "property": "status", "condition": "equals", "logical_operator": "AND", "value": "enabled"}},
			"gbt":     "", "fr": 1611964800, "to": 1612310400}}, "cl": "channel_v1"},

		{"query_group": [1]map[string]interface{}{{"channel": "google_ads", "select_metrics": [3]string{"clicks", "impressions", "spend"},
			"group_by": [2]map[string]interface{}{{"name": "campaign", "property": "id"}, {"name": "campaign", "property": "name"}},
			"gbt":      "", "fr": 1611964800, "to": 1612310400}}, "cl": "channel_v1"},
		{"query_group": [1]map[string]interface{}{{"channel": "google_ads", "select_metrics": [3]string{"clicks", "impressions", "spend"},
			"group_by": [2]map[string]interface{}{{"name": "campaign", "property": "id"}, {"name": "ad_group", "property": "name"}},
			"gbt":      "", "fr": 1611964800, "to": 1612310400}}, "cl": "channel_v1"},
		{"query_group": [1]map[string]interface{}{{"channel": "google_ads", "select_metrics": [3]string{"clicks", "impressions", "spend"},
			"group_by": [2]map[string]interface{}{{"name": "ad_group", "property": "name"}, {"name": "ad_group", "property": "id"}},
			"gbt":      "", "fr": 1611964800, "to": 1612310400}}, "cl": "channel_v1"},

		{"query_group": [1]map[string]interface{}{{"channel": "google_ads", "select_metrics": []string{"clicks", "impressions", "spend", "search_top_impression_share"},
			"gbt": "", "fr": 1611964800, "to": 1612310400}}, "cl": "channel_v1"},

		{"query_group": [1]map[string]interface{}{{"channel": "google_ads", "select_metrics": [3]string{"clicks", "impressions", "spend"},
			"group_by": [2]map[string]interface{}{{"name": "campaign", "property": "id"}},
			"filters":  [1]map[string]interface{}{{"name": "campaign", "property": "name", "condition": "contains", "logical_operator": "AND", "value": "1"}},
			"gbt":      "", "fr": 1611964800, "to": 1612310400}}, "cl": "channel_v1"},
		{"query_group": [1]map[string]interface{}{{"channel": "google_ads", "select_metrics": [3]string{"clicks", "impressions", "spend"},
			"group_by": [1]map[string]interface{}{{"name": "campaign", "property": "id"}},
			"filters":  [1]map[string]interface{}{{"name": "ad_group", "property": "name", "condition": "contains", "logical_operator": "AND", "value": "1"}},
			"gbt":      "", "fr": 1611964800, "to": 1612310400}}, "cl": "channel_v1"},
		{"query_group": [1]map[string]interface{}{{"channel": "google_ads", "select_metrics": [3]string{"clicks", "impressions", "spend"},
			"group_by": [1]map[string]interface{}{{"name": "ad_group", "property": "name"}},
			"filters":  [1]map[string]interface{}{{"name": "campaign", "property": "id", "condition": "equals", "logical_operator": "AND", "value": "1"}},
			"gbt":      "", "fr": 1611964800, "to": 1612310400}}, "cl": "channel_v1"},
		{"query_group": [1]map[string]interface{}{{"channel": "google_ads", "select_metrics": [3]string{"clicks", "impressions", "spend"},
			"group_by": [2]map[string]interface{}{{"name": "campaign", "property": "id"}, {"name": "ad_group", "property": "name"}},
			"filters":  [2]map[string]interface{}{{"name": "campaign", "property": "name", "condition": "contains", "logical_operator": "AND", "value": "1"}, {"name": "ad_group", "property": "name", "condition": "contains", "logical_operator": "AND", "value": "1"}},
			"gbt":      "", "fr": 1611964800, "to": 1612310400}}, "cl": "channel_v1"},

		{"query_group": [1]map[string]interface{}{{"channel": "google_ads", "select_metrics": []string{"impressions", "search_click_share"},
			"gbt": "", "fr": 1611964800, "to": 1612310400}}, "cl": "channel_v1"},
		{"query_group": [1]map[string]interface{}{{"channel": "google_ads", "select_metrics": []string{"search_click_share"},
			"gbt": "", "fr": 1611964800, "to": 1612310400}}, "cl": "channel_v1"},
		{"query_group": [1]map[string]interface{}{{"channel": "google_ads", "select_metrics": []string{"search_click_share"},
			"group_by": [1]map[string]interface{}{{"name": "ad_group", "property": "name"}},
			"gbt":      "", "fr": 1611964800, "to": 1612310400}}, "cl": "channel_v1"},
		{"query_group": [1]map[string]interface{}{{"channel": "google_ads", "select_metrics": []string{"clicks"},
			"group_by": [1]map[string]interface{}{{"name": "keyword", "property": "quality_score"}},
			"gbt":      "", "fr": 1611964800, "to": 1612310400}}, "cl": "channel_v1"},
	}

	successChannelResponse := [][]byte{
		[]byte(`{"result":{"result_group":[{"headers":["clicks","impressions","spend"],"rows":[[406,4006,0]]}]}}`),

		[]byte(`{"result":{"result_group":[{"headers":["clicks","impressions","spend"],"rows":[[202,2002,0]]}]}}`),
		[]byte(`{"result":{"result_group":[{"headers":["clicks","impressions","spend"],"rows":[[204,2004,0]]}]}}`),
		[]byte(`{"result_group":[{"headers":["clicks","impressions","spend"],"rows":[[204,2004,0]]}]}`),
		[]byte(`{"result":{"result_group":[{"headers":["clicks","impressions","spend"],"rows":[[0,0,0]]}]}}`),

		[]byte(`{"result":{"result_group":[{"headers":["campaign_id","campaign_name","clicks","impressions","spend"],"rows":[[2,"test2",204,2004,0],[1,"test1",202,2002,0]]}]}}`),
		[]byte(`{"result":{"result_group":[{"headers":["campaign_id","ad_group_name","clicks","impressions","spend"],"rows":[[2,"agtest3",206,2004,0],[1,"agtest2",204,2004,0],[1,"agtest1",202,2002,0]]}]}}`),
		[]byte(`{"result":{"result_group":[{"headers":["ad_group_name","ad_group_id","clicks","impressions","spend"],"rows":[["agtest3",21,206,2004,0],["agtest2",12,204,2004,0],["agtest1",11,202,2002,0]]}]}}`),

		[]byte(`{"result":{"result_group":[{"headers":["clicks","impressions","spend","search_top_impression_share"],"rows":[[406,4006,0,0]]}]}}`),

		[]byte(`{"result":{"result_group":[{"headers":["campaign_id","clicks","impressions","spend"],"rows":[[1,202,2002,0]]}]}}`),
		[]byte(`{"result":{"result_group":[{"headers":["campaign_id","clicks","impressions","spend"],"rows":[[1,202,2002,0]]}]}}`),
		[]byte(`{"result":{"result_group":[{"headers":["ad_group_name","clicks","impressions","spend"],"rows":[["agtest2",204,2004,0],["agtest1",202,2002,0]]}]}}`),
		[]byte(`{"result":{"result_group":[{"headers":["campaign_id","ad_group_name","clicks","impressions","spend"],"rows":[[1,"agtest1",202,2002,0]]}]}}`),

		[]byte(`{"result":{"result_group":[{"headers":["impressions","search_click_share"],"rows":[[4006,0]]}]}}`),
		[]byte(`{"result":{"result_group":[{"headers":["search_click_share"],"rows":[[0]]}]}}`),
		[]byte(`{"result":{"result_group":[{"headers":["ad_group_name","search_click_share"],"rows":[["agtest2",0],["agtest3",0],["agtest1",0.1]]}]}}`),
		[]byte(`{"result":{"result_group":[{"headers":["keyword_quality_score","clicks"],"rows":[[0,307],["0.2",204],["0.1",101]]}]}}`),
	}

	if C.UseMemSQLDatabaseStore() {
		// On memsql, seems to be coming last in the order as compared to postgres where null comes first.
		successChannelResponse[15] = []byte(`{"result":{"result_group":[{"headers":["ad_group_name","search_click_share"],"rows":[["agtest1",0.1],["agtest2",0],["agtest3",0]]}]}}`)
	}

	for index, channelQuery := range successChannelQueries {
		w := sendChannelAnalyticsQueryReq(a, project.ID, agent, channelQuery)
		assert.Equal(t, http.StatusOK, w.Code)
		assertIfResponseIsEqualToExpected(t, w.Body, successChannelResponse[index], index)
	}
}

func assertIfResponseIsEqualToExpected(t *testing.T, responseBody *bytes.Buffer, expectedResult []byte, index int) {
	var current interface{}
	var expected interface{}
	readBuf, _ := ioutil.ReadAll(responseBody)
	json.Unmarshal(readBuf, &current)
	err := json.Unmarshal([]byte(expectedResult), &expected)
	// Used for debugging.
	if err != nil {
		log.Warn("o1", current)
		log.Warn("o2", expected)
		log.WithError(err).Error("Error unmarshalling responseBody.", index)
	}
	// Used for debugging.
	if reflect.DeepEqual(current, expected) != true {
		log.Warn("o1", current)
		log.Warn("o2", expected)
		log.Error("Response and expected are not equal", index)
	}
	assert.Equal(t, reflect.DeepEqual(current, expected), true)
}

func TestExecuteChannelQueryHandlerForLinkedin(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)
	Const.SetSmartPropertiesReservedNames()

	//inserting sample data in linkedin, also testing data service endpoint linkedin/documents/add
	project, agent, _ := SetupProjectWithAgentDAO()
	assert.NotNil(t, project)
	customerAccountID := U.RandomNumericString(10)
	_, errCode := store.GetStore().UpdateProjectSettings(project.ID, &model.ProjectSetting{
		IntLinkedinAdAccount: customerAccountID,
	})
	assert.Equal(t, http.StatusAccepted, errCode)

	campaignID1 := U.RandomNumericString(8)
	campaignID1Float, _ := strconv.ParseFloat(campaignID1, 64)
	campaign1Value, _ := json.Marshal(map[string]interface{}{"costInLocalCurrency": "100", "clicks": "50", "campaign_group_id": campaignID1, "impressions": "1000", "campaign_group_name": "campaign_group_1"})
	campaignID2 := U.RandomNumericString(8)
	campaign2Value, _ := json.Marshal(map[string]interface{}{"costInLocalCurrency": "200", "clicks": "100", "campaign_group_id": campaignID2, "impressions": "2000", "campaign_group_name": "campaign_group_2"})
	adgroupID1_1 := U.RandomNumericString(8)
	adgroupID1_1Float, _ := strconv.ParseFloat(adgroupID1_1, 64)
	adgroupID1_1Value, _ := json.Marshal(map[string]interface{}{"costInLocalCurrency": "30", "clicks": "30", "campaign_id": adgroupID1_1, "campaign_name": "Adgroup_1_1", "campaign_group_id": campaignID1, "impressions": "600", "campaign_group_name": "campaign_group_1"})
	adgroupID1_2 := U.RandomNumericString(8)
	adgroupID1_2Float, _ := strconv.ParseFloat(adgroupID1_2, 64)
	adgroup1_2Value, _ := json.Marshal(map[string]interface{}{"costInLocalCurrency": "70", "clicks": "20", "campaign_id": adgroupID1_2, "campaign_name": "Adgroup_1_2", "campaign_group_id": campaignID1, "impressions": "400", "campaign_group_name": "campaign_group_1"})
	adgroupID2_1 := U.RandomNumericString(8)
	adgroupID2_1Float, _ := strconv.ParseFloat(adgroupID2_1, 64)
	adgroup2_1Value, _ := json.Marshal(map[string]interface{}{"costInLocalCurrency": "120", "clicks": "25", "campaign_id": adgroupID2_1, "campaign_name": "Adgroup_2_1", "campaign_group_id": campaignID2, "impressions": "1500", "campaign_group_name": "campaign_group_2"})
	adgroupID2_2 := U.RandomNumericString(8)
	adgroupID2_2Float, _ := strconv.ParseFloat(adgroupID2_2, 64)
	adgroup2_2Value, _ := json.Marshal(map[string]interface{}{"costInLocalCurrency": "80", "clicks": "75", "campaign_id": adgroupID2_2, "campaign_name": "Adgroup_2_2", "campaign_group_id": campaignID2, "impressions": "500", "campaign_group_name": "campaign_group_2"})
	linkedinDocuments := []model.LinkedinDocument{
		{ID: campaignID1, ProjectID: project.ID, CustomerAdAccountID: customerAccountID, TypeAlias: "campaign_group_insights", Timestamp: 20210205,
			Value: &postgres.Jsonb{campaign1Value}},

		{ID: campaignID2, ProjectID: project.ID, CustomerAdAccountID: customerAccountID, TypeAlias: "campaign_group_insights", Timestamp: 20210206,
			Value: &postgres.Jsonb{campaign2Value}},

		{ID: adgroupID1_1, ProjectID: project.ID, CustomerAdAccountID: customerAccountID, TypeAlias: "campaign_insights", Timestamp: 20210205,
			Value: &postgres.Jsonb{adgroupID1_1Value}},

		{ID: adgroupID1_2, ProjectID: project.ID, CustomerAdAccountID: customerAccountID, TypeAlias: "campaign_insights", Timestamp: 20210205,
			Value: &postgres.Jsonb{adgroup1_2Value}},

		{ID: adgroupID2_1, ProjectID: project.ID, CustomerAdAccountID: customerAccountID, TypeAlias: "campaign_insights", Timestamp: 20210206,
			Value: &postgres.Jsonb{adgroup2_1Value}},

		{ID: adgroupID2_2, ProjectID: project.ID, CustomerAdAccountID: customerAccountID, TypeAlias: "campaign_insights", Timestamp: 20210206,
			Value: &postgres.Jsonb{adgroup2_2Value}},
	}

	for _, linkedinDocument := range linkedinDocuments {
		log.Warn("tx", linkedinDocument)
		status := store.GetStore().CreateLinkedinDocument(project.ID, &linkedinDocument)
		assert.Equal(t, http.StatusCreated, status)
	}
	channelQuery := map[string]interface{}{"query_group": []map[string]interface{}{{"channel": "linkedin_ads", "select_metrics": []string{"clicks", "impressions", "spend"},
		"group_by": []map[string]interface{}{{"name": "campaign", "property": "id"}, {"name": "ad_group", "property": "name"}},
		"filters":  []map[string]interface{}{{"name": "campaign", "property": "name", "condition": "contains", "logical_operator": "AND", "value": "1"}, {"name": "ad_group", "property": "name", "condition": "contains", "logical_operator": "AND", "value": "1_1"}},
		"gbt":      "", "fr": 1612314000, "to": 1612746000}}, "cl": "channel_v1"}
	w := sendChannelAnalyticsQueryReq(r, project.ID, agent, channelQuery)
	assert.Equal(t, http.StatusOK, w.Code)

	var result resultStruct
	decoder := json.NewDecoder(w.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&result); err != nil {
		assert.NotNil(t, nil, err)
	}
	var expectedResult resultStruct
	expectedResult = resultStruct{
		Result: model.ChannelResultGroupV1{
			Results: []model.ChannelQueryResultV1{
				{
					Headers: []string{"campaign_id", "ad_group_name", "clicks", "impressions", "spend"},
					Rows:    [][]interface{}{{campaignID1Float, "Adgroup_1_1", float64(30), float64(600), float64(30)}},
				},
			},
		},
	}
	if C.UseMemSQLDatabaseStore() {
		expectedResult = resultStruct{
			Result: model.ChannelResultGroupV1{
				Results: []model.ChannelQueryResultV1{
					{
						Headers: []string{"campaign_id", "ad_group_name", "clicks", "impressions", "spend"},
						Rows:    [][]interface{}{{campaignID1, "Adgroup_1_1", float64(30), float64(600), float64(30)}},
					},
				},
			},
		}
	}
	assert.Equal(t, result, expectedResult)

	// filters : campaignID equals campaignID1, adGroupName contains 1_1
	channelQuery = map[string]interface{}{"query_group": [1]map[string]interface{}{{"channel": "linkedin_ads", "select_metrics": [3]string{"clicks", "impressions", "spend"},
		"filters": [2]map[string]interface{}{{"name": "campaign", "property": "id", "condition": "equals", "logical_operator": "AND", "value": campaignID1}, {"name": "ad_group", "property": "name", "condition": "contains", "logical_operator": "AND", "value": "1_1"}},
		"gbt":     "", "fr": 1612314000, "to": 1612746000}}, "cl": "channel_v1"}

	w = sendChannelAnalyticsQueryReq(r, project.ID, agent, channelQuery)
	assert.Equal(t, http.StatusOK, w.Code)

	var result1 resultStruct
	decoder = json.NewDecoder(w.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&result1); err != nil {
		assert.NotNil(t, nil, err)
	}
	expectedResult = resultStruct{
		Result: model.ChannelResultGroupV1{
			Results: []model.ChannelQueryResultV1{
				{
					Headers: []string{"clicks", "impressions", "spend"},
					Rows:    [][]interface{}{{float64(30), float64(600), float64(30)}},
				},
			},
		},
	}
	assert.Equal(t, result1, expectedResult)

	// filters : campaignID equals campaignID1, adGroupName contains 2_1, result should be 0 in result rows
	channelQuery = map[string]interface{}{"query_group": [1]map[string]interface{}{{"channel": "linkedin_ads", "select_metrics": [3]string{"clicks", "impressions", "spend"},
		"filters": [2]map[string]interface{}{{"name": "campaign", "property": "id", "condition": "equals", "logical_operator": "AND", "value": campaignID1}, {"name": "ad_group", "property": "name", "condition": "contains", "logical_operator": "AND", "value": "2_1"}},
		"gbt":     "", "fr": 1612314000, "to": 1612746000}}, "cl": "channel_v1"}

	w = sendChannelAnalyticsQueryReq(r, project.ID, agent, channelQuery)
	assert.Equal(t, http.StatusOK, w.Code)

	var result2 resultStruct
	decoder = json.NewDecoder(w.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&result2); err != nil {
		assert.NotNil(t, nil, err)
	}
	expectedResult = resultStruct{
		Result: model.ChannelResultGroupV1{
			Results: []model.ChannelQueryResultV1{
				{
					Headers: []string{"clicks", "impressions", "spend"},
					Rows:    [][]interface{}{{float64(0), float64(0), float64(0)}},
				},
			},
		},
	}
	assert.Equal(t, result2, expectedResult)

	//groupBy: campaignName, adGroupID
	channelQuery = map[string]interface{}{"query_group": [1]map[string]interface{}{{"channel": "linkedin_ads", "select_metrics": [3]string{"clicks", "impressions", "spend"},
		"group_by": [2]map[string]interface{}{{"name": "campaign", "property": "name"}, {"name": "ad_group", "property": "id"}},
		"gbt":      "", "fr": 1612314000, "to": 1612746000}}, "cl": "channel_v1"}

	w = sendChannelAnalyticsQueryReq(r, project.ID, agent, channelQuery)
	assert.Equal(t, http.StatusOK, w.Code)

	var result3 resultStruct
	decoder = json.NewDecoder(w.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&result3); err != nil {
		assert.NotNil(t, nil, err)
	}
	expectedResult = resultStruct{
		Result: model.ChannelResultGroupV1{
			Results: []model.ChannelQueryResultV1{
				{
					Headers: []string{"campaign_name", "ad_group_id", "clicks", "impressions", "spend"},
					Rows: [][]interface{}{
						{"campaign_group_2", adgroupID2_2Float, float64(75), float64(500), float64(80)},
						{"campaign_group_1", adgroupID1_1Float, float64(30), float64(600), float64(30)},
						{"campaign_group_2", adgroupID2_1Float, float64(25), float64(1500), float64(120)},
						{"campaign_group_1", adgroupID1_2Float, float64(20), float64(400), float64(70)}},
				},
			},
		},
	}
	if C.UseMemSQLDatabaseStore() {
		expectedResult = resultStruct{
			Result: model.ChannelResultGroupV1{
				Results: []model.ChannelQueryResultV1{
					{
						Headers: []string{"campaign_name", "ad_group_id", "clicks", "impressions", "spend"},
						Rows: [][]interface{}{
							{"campaign_group_2", adgroupID2_2, float64(75), float64(500), float64(80)},
							{"campaign_group_1", adgroupID1_1, float64(30), float64(600), float64(30)},
							{"campaign_group_2", adgroupID2_1, float64(25), float64(1500), float64(120)},
							{"campaign_group_1", adgroupID1_2, float64(20), float64(400), float64(70)}},
					},
				},
			},
		}
	}

	assert.Equal(t, result3, expectedResult)
}
