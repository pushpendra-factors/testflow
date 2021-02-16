package tests

import (
	"encoding/json"
	C "factors/config"
	H "factors/handler"
	"factors/handler/helpers"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func sendChannelAnalyticsQueryReq(r *gin.Engine, project_id uint64, agent *model.Agent, channelQueryJSON map[string]interface{}) *httptest.ResponseRecorder {
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error Creating cookieData")
	}
	url := fmt.Sprintf("http://localhost:8080/projects/%d/v1/channels/query", project_id)
	rb := U.NewRequestBuilder(http.MethodPost, url).
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
	Result model.ChannelResultGroupV1 `json:"result"`
}

func TestExecuteChannelQueryHandlerForLinkedin(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)

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
	expectedResult := resultStruct{
		Result: model.ChannelResultGroupV1{
			Results: []model.ChannelQueryResultV1{
				{
					Headers: []string{"campaign_id", "ad_group_name", "clicks", "impressions", "spend"},
					Rows:    [][]interface{}{{campaignID1Float, "Adgroup_1_1", float64(30), float64(600), float64(30)}},
				},
			},
		},
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
	assert.Equal(t, result3, expectedResult)
}
