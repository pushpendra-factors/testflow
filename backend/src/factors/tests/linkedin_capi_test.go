package tests

import (
	"encoding/json"
	H "factors/handler"
	"factors/integration/linkedin_capi"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/stretchr/testify/assert"
)

func TestLinkedinCapi(t *testing.T) {

	r := gin.Default()
	H.InitAppRoutes(r)

	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)

	_, status := store.GetStore().UpdateProjectSettings(project.ID, &model.ProjectSetting{
		IntLinkedinAdAccount:   "1234",
		IntLinkedinAccessToken: "12345566",
	})
	assert.Equal(t, http.StatusAccepted, status)

	t.Run("TestLinkedinAPIs", func(t *testing.T) {

		_, status := store.GetStore().UpdateProjectSettings(project.ID, &model.ProjectSetting{
			IntLinkedinAdAccount:   "508934217",
			IntLinkedinAccessToken: "12345",
		})
		assert.Equal(t, http.StatusAccepted, status)

		aa := `{"action_performed":"action_event","addtional_configuration":[{"account":"urn:li:sponsoredAccount:508934217","enabled":true,"id":17819097,"name":"MQL Conversions Alpha - Factors"}],"alert_limit":5,"breakdown_properties":[],"cool_down_time":1800,"description":"fe-testcapi-email-known","event":"$session","event_level":"user","filters":[{"en":"user","grpn":"user","lop":"AND","op":"notEqual","pr":"$email","ty":"categorical","va":"$none"}],"message_properties":{},"notifications":false,"repeat_alerts":true,"template_description":"","template_id":4000005,"template_title":"","title":"fe-testcapi-email-known"}
	`
		linkedInCapiAlertbodyJsonString := `{
			"action_performed": "action_event",
			"alert_limit": 5,
			"breakdown_properties": [],
			"cool_down_time": 1800,
			"event": "$session",
			"event_level": "user",
			"filters": [],
			"notifications": false,
			"repeat_alerts": true,
			"title": "15-feb-linkedincapitest-2",
			"description": "15-feb-linkedincapitest",
			"template_id": 4000005,
			"message_properties": {},
			"addtional_configuration": {
				"conversions": {
					"elements": [
						{
							"account": "urn:li:sponsoredAccount:508934217",
							"enabled": true,
							"id": 17819097,
							"name": "MQL Conversions Alpha - Factors"
						}
					]
				}
			}
		}`

		var workflow model.WorkflowAlertBody

		err := U.DecodePostgresJsonbToStructType(&postgres.Jsonb{RawMessage: json.RawMessage(linkedInCapiAlertbodyJsonString)}, &workflow)
		assert.Nil(t, err)

		err = U.DecodePostgresJsonbToStructType(&postgres.Jsonb{RawMessage: json.RawMessage(aa)}, &workflow)
		assert.Nil(t, err)

		wf, _, err := store.GetStore().CreateWorkflow(project.ID, agent.UUID, "", workflow)
		assert.Nil(t, err)

		config := model.LinkedinCAPIConfig{}
		alertBody := model.WorkflowAlertBody{}
		err = U.DecodePostgresJsonbToStructType(wf.AlertBody, &alertBody)
		assert.Nil(t, err)
		assert.NotNil(t, alertBody)

		err = U.DecodePostgresJsonbToStructType(alertBody.AdditonalConfigurations, &config)
		assert.Nil(t, err)
		assert.NotNil(t, config)
		assert.Equal(t, config.IsLinkedInCAPI, true)
		assert.NotNil(t, config.LinkedInAccessToken)
		assert.NotNil(t, config.LinkedInAdAccounts)

		jsonData := `{
		"elements":
			[
				{
					"conversion": "urn:lla:llaPartnerConversion:17050482",
					"conversionHappenedAt": 1718396407000,
					 "user": {
					  "userIds": [ {
							 "idType": "SHA256_EMAIL",
							 "idValue": "bad8677b6c86f5d308ee82786c183482a5995f066694246c58c4df37b0cc41f1"
							  }
					  ]
	  
					  }
			  },
			  {
					"conversion": "urn:lla:llaPartnerConversion:17050482",
					"conversionHappenedAt": 1718396404000,
					 "user": {
					  "userIds": [ {
							 "idType": "SHA256_EMAIL",
							 "idValue": "bad8677b6c86f5d308ee82786c183482a5995f066694246c58c4df37b0cc41f1"
							  }
					  ]
	  
					  }
			  }
	  
	  
			  
		  ]
	  }`
		var body model.BatchLinkedinCAPIRequestPayload

		err = json.Unmarshal([]byte(jsonData), &body)
		assert.Nil(t, err)

		reponse := map[string]interface{}{
			"status": "succsess",
			"elements": []map[string]interface{}{
				{
					"status": 201,
				},
				{
					"status": 201,
				},
			},
		}

		linkedinCapiSendEventsMock := linkedin_capi.LinkedInCapiMock{}
		linkedinCapiSendEventsMock.SendEventsToLinkedCAPIData = reponse

		res, err := linkedinCapiSendEventsMock.SendEventsToLinkedCAPI(config, body)
		assert.NotNil(t, res)
		assert.Nil(t, err)

		// res, err = linkedin_capi.GetLinkedInCapi().SendEventsToLinkedCAPI(config, body)
		// assert.NotNil(t, res)
		// assert.Nil(t, err)

		linkedinCapiGetConversionListMock := linkedin_capi.LinkedInCapiMock{}
		linkedinCapiGetConversionListMock.GetConversionFromLinkedCAPIData = model.BatchLinkedInCAPIConversionsResponse{
			LinkedInCAPIConversionsResponseList: []model.SingleLinkedInCAPIConversionsResponse{
				{ConversionsId: 1234, ConversoinsDisplayName: "test", IsEnabled: true, AdAccount: "12345"},
				{ConversionsId: 12344, ConversoinsDisplayName: "test2", IsEnabled: true, AdAccount: "12345"},
			},
		}

		res1, err := linkedinCapiGetConversionListMock.GetConversionFromLinkedCAPI(config)
		assert.NotNil(t, res1)
		assert.Nil(t, err)

		// res1, err = linkedin_capi.GetLinkedInCapi().GetConversionFromLinkedCAPI(config)
		// assert.NotNil(t, res1)
		// assert.Nil(t, err)
	})

}

func TestLinkedinCapiForWorkflow(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)

	project, agent, err := SetupProjectWithAgentDAO()
	assert.Nil(t, err)
	assert.NotNil(t, project)

	_, status := store.GetStore().UpdateProjectSettings(project.ID, &model.ProjectSetting{
		IntLinkedinAdAccount:   "1234",
		IntLinkedinAccessToken: "12345566",
	})
	assert.Equal(t, http.StatusAccepted, status)

	t.Run("TestRun", func(t *testing.T) {

		_, status := store.GetStore().UpdateProjectSettings(project.ID, &model.ProjectSetting{
			IntLinkedinAdAccount:   "508934217",
			IntLinkedinAccessToken: "AQVCGk2KGIsrJU2M5nOlCQQXCTn1GzZyrzQ1dEKdMMRv-iNCG3-MQo1zhfYplv_yyzmX5Uzj4aUSGeBBt2EpYyniNckulTHBvZiVJHDlcvYuj1lWU7i2nWmqiuYAOK-Cz25Wq-3R6QE7mRh0vwTUG6CbbS1dUVHsrR_BCmJ9TW0iArOJNkK8yVkNGBM9MOBM6rTbRcKgp1I309W3lmJarL4jnPjIp0lZhbYC9g306gVRXRQMxizOmmMwN2M8jMz2WKrPL_EYkoeHWSHYrMcmfTX4FFaS101PeAeokJ1oTCBxVLNb29sa3lzl3HN_dKl7dbXEbfu4U6hDOdn3B3DmQ41exDawlZVMCEEcoVYFOrXvzP_f2JMAlbJDDfLIaqIgt8Jsy6PwrsgNf_HOQ9w",
		})
		assert.Equal(t, http.StatusAccepted, status)

		aa := `{"action_performed":"action_event","addtional_configuration":[{"account":"urn:li:sponsoredAccount:508934217","enabled":true,"id":17819097,"name":"MQL Conversions Alpha - Factors"}],"alert_limit":5,"breakdown_properties":[],"cool_down_time":1800,"description":"fe-testcapi-email-known","event":"$session","event_level":"user","filters":[{"en":"user","grpn":"user","lop":"AND","op":"notEqual","pr":"$email","ty":"categorical","va":"$none"}],"message_properties":{},"notifications":false,"repeat_alerts":true,"template_description":"","template_id":4000005,"template_title":"","title":"fe-testcapi-email-known"}
	`
		linkedInCapiAlertbodyJsonString := `{
			"action_performed": "action_event",
			"alert_limit": 5,
			"breakdown_properties": [],
			"cool_down_time": 1800,
			"event": "$session",
			"event_level": "user",
			"filters": [],
			"notifications": false,
			"repeat_alerts": true,
			"title": "15-feb-linkedincapitest-2",
			"description": "15-feb-linkedincapitest",
			"template_id": 4000005,
			"message_properties": {},
			"addtional_configuration": {
				"conversions": {
					"elements": [
						{
							"account": "urn:li:sponsoredAccount:508934217",
							"enabled": true,
							"id": 17819097,
							"name": "MQL Conversions Alpha - Factors"
						}
					]
				}
			}
		}`

		var workflow model.WorkflowAlertBody

		err := U.DecodePostgresJsonbToStructType(&postgres.Jsonb{RawMessage: json.RawMessage(linkedInCapiAlertbodyJsonString)}, &workflow)
		assert.Nil(t, err)

		err = U.DecodePostgresJsonbToStructType(&postgres.Jsonb{RawMessage: json.RawMessage(aa)}, &workflow)
		assert.Nil(t, err)

		wf, _, err := store.GetStore().CreateWorkflow(project.ID, agent.UUID, "", workflow)
		assert.Nil(t, err)

		config := model.LinkedinCAPIConfig{}
		alertBody := model.WorkflowAlertBody{}
		err = U.DecodePostgresJsonbToStructType(wf.AlertBody, &alertBody)
		assert.Nil(t, err)
		assert.NotNil(t, alertBody)

		err = U.DecodePostgresJsonbToStructType(alertBody.AdditonalConfigurations, &config)
		assert.Nil(t, err)
		assert.NotNil(t, config)
		assert.Equal(t, config.IsLinkedInCAPI, true)
		assert.NotNil(t, config.LinkedInAccessToken)
		assert.NotNil(t, config.LinkedInAdAccounts)

		jsonData := `{
		"elements":
			[
				{
					"conversion": "urn:lla:llaPartnerConversion:17050482",
					"conversionHappenedAt": 1718396407000,
					 "user": {
					  "userIds": [ {
							 "idType": "SHA256_EMAIL",
							 "idValue": "bad8677b6c86f5d308ee82786c183482a5995f066694246c58c4df37b0cc41f1"
							  }
					  ]
	  
					  }
			  },
			  {
					"conversion": "urn:lla:llaPartnerConversion:17050482",
					"conversionHappenedAt": 1718396404000,
					 "user": {
					  "userIds": [ {
							 "idType": "SHA256_EMAIL",
							 "idValue": "bad8677b6c86f5d308ee82786c183482a5995f066694246c58c4df37b0cc41f1"
							  }
					  ]
	  
					  }
			  }
	  
	  
			  
		  ]
	  }`
		var body model.BatchLinkedinCAPIRequestPayload

		err = json.Unmarshal([]byte(jsonData), &body)
		assert.Nil(t, err)

		reponse := map[string]interface{}{
			"status": "succsess",
			"elements": []map[string]interface{}{
				{
					"status": 201,
				},
				{
					"status": 201,
				},
			},
		}

		linkedinCapiSendEventsMock := linkedin_capi.LinkedInCapiMock{}
		linkedinCapiSendEventsMock.SendEventsToLinkedCAPIData = reponse

		res, err := linkedinCapiSendEventsMock.SendEventsToLinkedCAPI(config, body)
		assert.NotNil(t, res)
		assert.Nil(t, err)

	})

}
