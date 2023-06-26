package tests

import (
	"encoding/json"
	C "factors/config"
	H "factors/handler"
	"factors/handler/helpers"
	"factors/model/model"
	U "factors/util"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

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
