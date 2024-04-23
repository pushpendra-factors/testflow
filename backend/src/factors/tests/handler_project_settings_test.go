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
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm/dialects/postgres"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func sendGetProjectSettingsReq(r *gin.Engine, projectId int64, agent *model.Agent) *httptest.ResponseRecorder {

	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error Creating cookieData")
	}
	rb := C.NewRequestBuilderWithPrefix(http.MethodGet, fmt.Sprintf("/projects/%d/settings", projectId)).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})

	w := httptest.NewRecorder()
	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error Creating getProjectSetting Req")
	}
	r.ServeHTTP(w, req)
	return w
}

func TestAPIGetProjectSettingHandler(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)

	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)

	// Test get project settings.
	t.Run("Success", func(t *testing.T) {
		w := sendGetProjectSettingsReq(r, project.ID, agent)
		assert.Equal(t, http.StatusOK, w.Code)
		jsonResponse, _ := ioutil.ReadAll(w.Body)
		var jsonResponseMap map[string]interface{}
		json.Unmarshal(jsonResponse, &jsonResponseMap)
		assert.NotEqual(t, 0, jsonResponseMap["id"])
		assert.NotNil(t, jsonResponseMap["auto_track"])
		assert.NotNil(t, jsonResponseMap["int_drift"])
		assert.NotNil(t, jsonResponseMap["int_clear_bit"])
		assert.NotNil(t, jsonResponseMap["timelines_config"])
	})

	// Test get project settings with bad id.
	t.Run("BadID", func(t *testing.T) {
		badProjectID := int64(0)
		w := sendGetProjectSettingsReq(r, badProjectID, agent)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		jsonResponse, _ := ioutil.ReadAll(w.Body)
		var jsonRespMap map[string]interface{}
		json.Unmarshal(jsonResponse, &jsonRespMap)
		assert.NotNil(t, jsonRespMap["error"])
		assert.Equal(t, 1, len(jsonRespMap))
	})

}

func sendUpdateProjectSettingReq(r *gin.Engine, projectId int64, agent *model.Agent, params map[string]interface{}) *httptest.ResponseRecorder {
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error Creating cookieData")
	}
	rb := C.NewRequestBuilderWithPrefix(http.MethodPut, fmt.Sprintf("/projects/%d/settings", projectId)).
		WithPostParams(params).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})

	w := httptest.NewRecorder()
	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error Creating UpdateProjectSetting Req")
	}
	r.ServeHTTP(w, req)
	return w
}

func TestAPIUpdateProjectSettingsHandler(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)

	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)

	t.Run("UpdateAutoTrack", func(t *testing.T) {
		w := sendUpdateProjectSettingReq(r, project.ID, agent, map[string]interface{}{
			"auto_track": true,
		})
		assert.Equal(t, http.StatusOK, w.Code)
		jsonResponse, _ := ioutil.ReadAll(w.Body)
		var jsonResponseMap map[string]interface{}
		json.Unmarshal(jsonResponse, &jsonResponseMap)
		assert.NotNil(t, jsonResponseMap["auto_track"])

	})

	t.Run("UpdateAttribution_config", func(t *testing.T) {
		w := sendUpdateProjectSettingReq(r, project.ID, agent, map[string]interface{}{
			"attribution_config": model.AttributionConfig{AttributionWindow: 6},
		})
		assert.Equal(t, http.StatusOK, w.Code)
		jsonResponse, _ := ioutil.ReadAll(w.Body)
		var projectSettings model.ProjectSetting
		json.Unmarshal(jsonResponse, &projectSettings)
		//assert.Equal(t, int64(6), projectSettings.AttributionConfig)

	})

	t.Run("UpdateTimelines_config", func(t *testing.T) {
		timelinesConfig := model.TimelinesConfig{
			DisabledEvents: []string{"$hubspot_contact_updated", "$sf_contact_updated"},
			UserConfig: model.UserConfig{
				Milestones:    []string{},
				TableProps:    []string{"$country"},
				LeftpaneProps: []string{"$email", "$user_id"},
			},
			AccountConfig: model.AccountConfig{
				Milestones:    []string{},
				TableProps:    []string{"$country"},
				LeftpaneProps: []string{"$hubspot_company_industry", "$hubspot_company_country"},
				UserProp:      "$hubspot_contact_jobtitle",
			},
		}

		w := sendUpdateProjectSettingReq(r, project.ID, agent, map[string]interface{}{"timelines_config": timelinesConfig})
		assert.Equal(t, http.StatusOK, w.Code)
		jsonResponse, _ := ioutil.ReadAll(w.Body)
		var projectSettings model.ProjectSetting
		json.Unmarshal(jsonResponse, &projectSettings)
		rawConfigFromProject := projectSettings.TimelinesConfig.RawMessage
		tlConfigDecoded := model.TimelinesConfig{}
		err = json.Unmarshal(rawConfigFromProject, &tlConfigDecoded)
		assert.Nil(t, err)
		assert.NotNil(t, tlConfigDecoded)
		assert.Equal(t, timelinesConfig.DisabledEvents, tlConfigDecoded.DisabledEvents)
		assert.Equal(t, timelinesConfig.UserConfig.LeftpaneProps, tlConfigDecoded.UserConfig.LeftpaneProps)
		assert.Equal(t, timelinesConfig.AccountConfig.LeftpaneProps, tlConfigDecoded.AccountConfig.LeftpaneProps)
		assert.Equal(t, timelinesConfig.AccountConfig.UserProp, tlConfigDecoded.AccountConfig.UserProp)
		assert.Equal(t, timelinesConfig.AccountConfig.TableProps, tlConfigDecoded.AccountConfig.TableProps)
	})

	t.Run("UpdateIntDrift", func(t *testing.T) {
		w := sendUpdateProjectSettingReq(r, project.ID, agent, map[string]interface{}{
			"int_drift": true,
		})
		assert.Equal(t, http.StatusOK, w.Code)
		jsonResponse, _ := ioutil.ReadAll(w.Body)
		var jsonResponseMap map[string]interface{}
		json.Unmarshal(jsonResponse, &jsonResponseMap)
		assert.NotNil(t, jsonResponseMap["int_drift"])
	})

	t.Run("UpdateIntClearBit", func(t *testing.T) {
		w := sendUpdateProjectSettingReq(r, project.ID, agent, map[string]interface{}{
			"int_clear_bit": true,
		})
		assert.Equal(t, http.StatusOK, w.Code)
		jsonResponse, _ := ioutil.ReadAll(w.Body)
		var jsonResponseMap map[string]interface{}
		json.Unmarshal(jsonResponse, &jsonResponseMap)
		assert.NotNil(t, jsonResponseMap["int_clear_bit"])
	})

	t.Run("UpdateExcludeBot", func(t *testing.T) {
		w := sendUpdateProjectSettingReq(r, project.ID, agent, map[string]interface{}{
			"exclude_bot": false,
		})
		assert.Equal(t, http.StatusOK, w.Code)
		jsonResponse, _ := ioutil.ReadAll(w.Body)
		var jsonResponseMap map[string]interface{}
		json.Unmarshal(jsonResponse, &jsonResponseMap)
		assert.NotNil(t, jsonResponseMap["exclude_bot"])
	})

	// Test updating project id.
	t.Run("BadParamsTryUpdatingProjectId", func(t *testing.T) {
		randomProjectId := int64(999999999)
		params := map[string]interface{}{
			"auto_track": true,
			"project_id": randomProjectId,
		}
		w := sendUpdateProjectSettingReq(r, project.ID, agent, params)
		// project_id becomes unknown field as omitted on json.
		assert.Equal(t, http.StatusBadRequest, w.Code)
		jsonResponse, _ := ioutil.ReadAll(w.Body)
		var jsonRespMap map[string]interface{}
		json.Unmarshal(jsonResponse, &jsonRespMap)
		assert.NotNil(t, jsonRespMap["error"])
	})

	// Test update project settings with bad project id.
	t.Run("BadParamsInvalidProjectId", func(t *testing.T) {

		w := sendUpdateProjectSettingReq(r, 0, agent, map[string]interface{}{
			"auto_track": true,
		})
		assert.Equal(t, http.StatusBadRequest, w.Code)

		jsonResponse, _ := ioutil.ReadAll(w.Body)
		var jsonRespMap map[string]interface{}
		json.Unmarshal(jsonResponse, &jsonRespMap)
		assert.NotNil(t, jsonRespMap["error"])
		assert.Equal(t, 1, len(jsonRespMap))
	})
	// Test updating autotrack_spa_page_view
	t.Run("UpdateAutoTrackSpaPageView", func(t *testing.T) {
		w := sendUpdateProjectSettingReq(r, project.ID, agent, map[string]interface{}{
			"auto_track_spa_page_view": false,
		})
		assert.Equal(t, http.StatusOK, w.Code)

		jsonResponse, _ := ioutil.ReadAll(w.Body)
		var jsonRespMap map[string]interface{}
		json.Unmarshal(jsonResponse, &jsonRespMap)
		assert.NotNil(t, jsonRespMap["auto_track_spa_page_view"])
	})

	//Test updating filter_ips
	t.Run("UpdateFilterIps", func(t *testing.T) {
		// updating with invalid ip
		filterIps := model.FilterIps{
			BlockIps: []string{"192.168.000.354", "10.40.210.253", "2001:0db8:85a3:0000:0000:8a2e:0370:7334"},
		}
		filtersIpsEncoded, err := U.EncodeStructTypeToPostgresJsonb(filterIps)
		assert.Nil(t, err)
		w := sendUpdateProjectSettingReq(r, project.ID, agent, map[string]interface{}{"filter_ips": filtersIpsEncoded})
		assert.Equal(t, http.StatusBadRequest, w.Code)

		// updating with valid ip
		filterIps = model.FilterIps{
			BlockIps: []string{"192.158.1.38", "10.40.210.253", "2001:0db8:85a3:0000:0000:8a2e:0370:7334"},
		}
		filtersIpsEncoded1, err := U.EncodeStructTypeToPostgresJsonb(filterIps)
		assert.Nil(t, err)
		w = sendUpdateProjectSettingReq(r, project.ID, agent, map[string]interface{}{"filter_ips": filtersIpsEncoded1})
		assert.Equal(t, http.StatusOK, w.Code)

		w = sendGetProjectSettingsReq(r, project.ID, agent)
		assert.Equal(t, http.StatusOK, w.Code)
		jsonResponse, _ := ioutil.ReadAll(w.Body)
		var jsonResponseMap map[string]interface{}
		json.Unmarshal(jsonResponse, &jsonResponseMap)
		assert.NotEmpty(t, jsonResponseMap["filter_ips"])
	})
}

func TestUpdateHubspotProjectSettings(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)

	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)

	w := sendUpdateProjectSettingReq(r, project.ID, agent, map[string]interface{}{
		"int_hubspot_api_key": "1234",
	})
	assert.Equal(t, http.StatusOK, w.Code)
	var jsonResponseMap map[string]interface{}
	jsonResponse, _ := ioutil.ReadAll(w.Body)
	json.Unmarshal(jsonResponse, &jsonResponseMap)
	assert.NotNil(t, jsonResponseMap["int_hubspot_api_key"])

	w = sendUpdateProjectSettingReq(r, project.ID, agent, map[string]interface{}{
		"int_hubspot_portal_id": 1234,
	})

	jsonResponse, _ = ioutil.ReadAll(w.Body)
	json.Unmarshal(jsonResponse, &jsonResponseMap)
	assert.Equal(t, float64(1234), jsonResponseMap["int_hubspot_portal_id"])
}

func sendIntegrationsStatusReq(r *gin.Engine, projectId int64, agent *model.Agent) *httptest.ResponseRecorder {
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error Creating cookieData")
	}
	rb := C.NewRequestBuilderWithPrefix(http.MethodGet, fmt.Sprintf("/projects/%d/integrations_status", projectId)).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})

	w := httptest.NewRecorder()
	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error Creating UpdateProjectSetting Req")
	}
	r.ServeHTTP(w, req)
	return w
}

func TestIntegrationsStatusHandler(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)

	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)

	secondDocumentID := rand.Intn(100) + 200
	cuid := getRandomEmail()
	leadguid := U.RandomString(5)
	lastmodifieddate := time.Now().UTC().Unix() * 1000
	createdDate := time.Now().AddDate(0, 0, -1).Unix() * 1000
	// Hubspot Campaign performance report
	jsonContactModel := `{
		"vid": %d,
		"addedAt": %d,
		"properties": {
		  "createdate": { "value": "%d" },
		  "lastmodifieddate": { "value": "%d" },
		  "lifecyclestage": { "value": "%s" }
		},
		"identity-profiles": [
		  {
			"vid": 1,
			"identities": [
			  {
				"type": "EMAIL",
				"value": "%s"
			  },
			  {
				"type": "LEAD_GUID",
				"value": "%s"
			  }
			]
		  }
		]
	  }`

	jsonContact := fmt.Sprintf(jsonContactModel, secondDocumentID, createdDate, createdDate, lastmodifieddate, "lead", cuid, leadguid)
	contactPJson := postgres.Jsonb{json.RawMessage(jsonContact)}

	hubspotDocument := model.HubspotDocument{
		TypeAlias: model.HubspotDocumentTypeNameContact,
		Action:    model.HubspotDocumentActionCreated,
		Value:     &contactPJson,
	}

	status := store.GetStore().CreateHubspotDocument(project.ID, &hubspotDocument)
	assert.Equal(t, http.StatusCreated, status)

	adwordsCustomerAccountId := U.RandomLowerAphaNumString(5)
	fbCustomerAccountId := U.RandomLowerAphaNumString(5)

	assert.Nil(t, err)

	// Facebook Campaign performance report
	value := []byte(`{"spend": "1","clicks": "1","campaign_id":"1000","impressions": "1", "campaign_name": "Campaign_Facebook_1000", "platform":"facebook"}`)
	documentFB := &model.FacebookDocument{
		ProjectID:           project.ID,
		ID:                  "1000",
		CustomerAdAccountID: fbCustomerAccountId,
		TypeAlias:           "campaign_insights",
		Timestamp:           20200510,
		Value:               &postgres.Jsonb{RawMessage: value},
		CampaignID:          "1000",
		Platform:            "facebook",
	}
	status = store.GetStore().CreateFacebookDocument(project.ID, documentFB)
	assert.Equal(t, http.StatusCreated, status)

	// Adwords Campaign performance report
	value = []byte(`{"cost": "1","clicks": "1","campaign_id":"100","impressions": "1", "campaign_name": "Campaign_Adwords_100"}`)
	document := &model.AdwordsDocument{
		ProjectID:         project.ID,
		ID:                "100",
		CustomerAccountID: adwordsCustomerAccountId,
		TypeAlias:         model.CampaignPerformanceReport,
		Timestamp:         20200510,
		Value:             &postgres.Jsonb{RawMessage: value},
		CampaignID:        100,
	}
	status = store.GetStore().CreateAdwordsDocument(document)
	assert.Equal(t, http.StatusCreated, status)

	user1Properties := postgres.Jsonb{json.RawMessage(`{"name":"abc","city":"xyz"}`)}

	user1 := &model.CRMUser{
		ProjectID:  project.ID,
		Source:     U.CRM_SOURCE_LEADSQUARED,
		Type:       1,
		ID:         "123",
		Properties: &user1Properties,
		Timestamp:  time.Now().Unix(),
	}
	status, err = store.GetStore().CreateCRMUser(user1)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusCreated, status)
	assert.Equal(t, user1.Action, model.CRMActionCreated)

	user2Properties := postgres.Jsonb{json.RawMessage(`{"name":"abcd","city":"wxyz"}`)}

	user2 := &model.CRMUser{
		ProjectID:  project.ID,
		Source:     U.CRM_SOURCE_MARKETO,
		Type:       1,
		ID:         "1234",
		Properties: &user2Properties,
		Timestamp:  time.Now().Unix(),
		CreatedAt:  time.Now().AddDate(0, 0, -3),
	}
	status, err = store.GetStore().CreateCRMUser(user2)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusCreated, status)
	assert.Equal(t, user1.Action, model.CRMActionCreated)

	// creating linkedin document
	customerAccountID := U.RandomNumericString(10)
	_, errCode := store.GetStore().UpdateProjectSettings(project.ID, &model.ProjectSetting{
		IntLinkedinAdAccount: customerAccountID,
	})
	assert.Equal(t, http.StatusAccepted, errCode)

	campaignID1 := U.RandomNumericString(8)
	campaign1Value, _ := json.Marshal(map[string]interface{}{"costInLocalCurrency": "100", "clicks": "50", "campaign_group_id": campaignID1, "impressions": "1000", "campaign_group_name": "campaign_group_1"})

	linkedinDocument := model.LinkedinDocument{
		ID:                  campaignID1,
		ProjectID:           project.ID,
		CustomerAdAccountID: customerAccountID,
		TypeAlias:           "campaign_group_insights",
		Timestamp:           20210205,
		Value:               &postgres.Jsonb{campaign1Value},
	}

	errCode = store.GetStore().CreateLinkedinDocument(linkedinDocument.ProjectID, &linkedinDocument)
	assert.Equal(t, http.StatusCreated, errCode)
	currentTime := time.Now().AddDate(0, 0, -4)
	_, errCode = store.GetStore().CreateUser(&model.User{ProjectId: project.ID, Source: model.GetRequestSourcePointer(model.UserSourceWeb), LastEventAt: &currentTime})
	assert.Equal(t, http.StatusCreated, errCode)

	w := sendIntegrationsStatusReq(r, project.ID, agent)
	assert.Equal(t, http.StatusOK, w.Code)
	var jsonResponseMap map[string]model.IntegrationState
	jsonResponse, _ := ioutil.ReadAll(w.Body)
	json.Unmarshal(jsonResponse, &jsonResponseMap)

	assert.Equal(t, jsonResponseMap["adwords"].State, model.PULL_DELAYED)
	assert.NotEqual(t, jsonResponseMap["adwords"].LastSyncedAt, int64(0))

	assert.Equal(t, jsonResponseMap["google_organic"].State, model.SYNCED)
	assert.Equal(t, jsonResponseMap["google_organic"].LastSyncedAt, int64(0))

	assert.Equal(t, jsonResponseMap["facebook"].State, model.PULL_DELAYED)
	assert.NotEqual(t, jsonResponseMap["facebook"].LastSyncedAt, int64(0))

	assert.Equal(t, jsonResponseMap["linkedin"].State, model.SYNCED)
	assert.NotEqual(t, jsonResponseMap["linkedin"].LastSyncedAt, int64(0))

	assert.Equal(t, jsonResponseMap["salesforce"].State, model.SYNCED)
	assert.Equal(t, jsonResponseMap["salesforce"].LastSyncedAt, int64(0))

	assert.Equal(t, jsonResponseMap["hubspot"].State, model.SYNC_PENDING)
	assert.NotEqual(t, jsonResponseMap["hubspot"].LastSyncedAt, int64(0))

	assert.Equal(t, jsonResponseMap["leadsquared"].State, model.SYNC_PENDING)
	assert.NotEqual(t, jsonResponseMap["leadsquared"].LastSyncedAt, int64(0))

	assert.Equal(t, jsonResponseMap["marketo"].State, model.SYNCED)
	assert.NotEqual(t, jsonResponseMap["marketo"].LastSyncedAt, int64(0))

	assert.Equal(t, jsonResponseMap["sdk"].State, model.SYNCED)
	assert.NotEqual(t, jsonResponseMap["sdk"].LastSyncedAt, int64(0))

	assert.Equal(t, jsonResponseMap["segment"].State, model.SYNCED)
	assert.NotEqual(t, jsonResponseMap["segment"].LastSyncedAt, int64(0))

	assert.Equal(t, jsonResponseMap["rudderstack"].State, model.SYNCED)
	assert.NotEqual(t, jsonResponseMap["rudderstack"].LastSyncedAt, int64(0))

}
