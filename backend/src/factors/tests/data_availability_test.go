package tests

import (
	"encoding/json"
	C "factors/config"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/stretchr/testify/assert"
)

func TestDataAvailability(t *testing.T) {
	C.GetConfig().DataAvailabilityExpiry = 15
	project, agent, _ := SetupProjectWithAgentDAO()
	result, _ := store.GetStore().GetLatestDataStatus([]string{"*"}, project.ID, false)
	// Default has session
	assert.Equal(t, len(result), 1)
	intHubspot := true
	// Adding hubspot
	_, _ = store.GetStore().UpdateProjectSettings(project.ID, &model.ProjectSetting{
		IntHubspot: &intHubspot, IntHubspotApiKey: "1234",
	})
	result, _ = store.GetStore().GetLatestDataStatus([]string{"*"}, project.ID, false)
	assert.Equal(t, len(result), 2)
	// Adding salesforce
	refreshToken := U.RandomLowerAphaNumString(5)
	instanceURL := U.RandomLowerAphaNumString(5)
	_ = store.GetStore().UpdateAgentIntSalesforce(agent.UUID,
		refreshToken,
		instanceURL,
	)
	_, _ = store.GetStore().UpdateProjectSettings(project.ID, &model.ProjectSetting{
		IntSalesforceEnabledAgentUUID: &agent.UUID,
	})
	result, _ = store.GetStore().GetLatestDataStatus([]string{"*"}, project.ID, false)
	assert.Equal(t, len(result), 3)
	// adwords
	accountId := U.RandomLowerAphaNumString(6)
	_, _ = store.GetStore().UpdateProjectSettings(project.ID, &model.ProjectSetting{
		IntAdwordsCustomerAccountId: &accountId, IntAdwordsEnabledAgentUUID: &(agent.UUID)})
	result, _ = store.GetStore().GetLatestDataStatus([]string{"*"}, project.ID, false)
	assert.Equal(t, len(result), 4)
	// linkedin
	_, _ = store.GetStore().UpdateProjectSettings(project.ID, &model.ProjectSetting{
		IntLinkedinAdAccount: "123", IntLinkedinAccessToken: "123",
		IntLinkedinAgentUUID: &(agent.UUID), IntLinkedinAccessTokenExpiry: time.Now().Unix() + 10000,
	})
	result, _ = store.GetStore().GetLatestDataStatus([]string{"*"}, project.ID, false)
	assert.Equal(t, len(result), 5)
	// google-organic
	addEnableAgentUUIDSetting := model.ProjectSetting{IntGoogleOrganicEnabledAgentUUID: &(agent.UUID)}
	_, _ = store.GetStore().UpdateProjectSettings(project.ID, &addEnableAgentUUIDSetting)
	urlPrefix := U.RandomNumericString(10)
	_, _ = store.GetStore().UpdateProjectSettings(project.ID, &model.ProjectSetting{
		IntGoogleOrganicURLPrefixes: &urlPrefix,
	})
	result, _ = store.GetStore().GetLatestDataStatus([]string{"*"}, project.ID, false)
	assert.Equal(t, len(result), 6)
	// facebook
	_, _ = store.GetStore().UpdateProjectSettings(project.ID, &model.ProjectSetting{
		IntFacebookEmail: "123", IntFacebookAccessToken: "123",
		IntFacebookAgentUUID: &(agent.UUID), IntFacebookUserID: "123",
		IntFacebookAdAccount: "1235w35", IntFacebookTokenExpiry: time.Now().Unix() + 10000})
	result, _ = store.GetStore().GetLatestDataStatus([]string{"*"}, project.ID, false)
	assert.Equal(t, len(result), 7)
	// bingads
	store.GetStore().PostFiveTranMapping(project.ID, "bingads", "123", "23", "124")
	store.GetStore().EnableFiveTranMapping(project.ID, "bingads", "123", "124")
	result, _ = store.GetStore().GetLatestDataStatus([]string{"*"}, project.ID, false)
	assert.Equal(t, len(result), 8)
	// bingads
	store.GetStore().PostFiveTranMapping(project.ID, "marketo", "123", "23", "124")
	store.GetStore().EnableFiveTranMapping(project.ID, "marketo", "123", "124")
	result, _ = store.GetStore().GetLatestDataStatus([]string{"*"}, project.ID, false)
	assert.Equal(t, len(result), 9)

	// insert dummy marketo data
	user1Properties := postgres.Jsonb{json.RawMessage(`{"name":"abc","city":"xyz"}`)}
	user1 := &model.CRMUser{
		ProjectID:  project.ID,
		Source:     U.CRM_SOURCE_MARKETO,
		Type:       1,
		ID:         "123",
		Properties: &user1Properties,
		Timestamp:  time.Now().Unix(),
	}
	_, _ = store.GetStore().CreateCRMUser(user1)
	store.GetStore().UpdateCRMUserAsSynced(project.ID, U.CRM_SOURCE_MARKETO, user1, "123", "123")

	// insert dummy bingads data
	intDocument := model.IntegrationDocument{
		DocumentId:        "123",
		ProjectID:         project.ID,
		CustomerAccountID: "123",
		Source:            model.BingAdsIntegration,
		DocumentType:      1,
		Timestamp:         20220601,
		Value:             &user1Properties,
	}
	_ = store.GetStore().UpsertIntegrationDocument(intDocument)

	// insert dummy adwords data
	adwords_value := map[string]interface{}{"cost": "100", "clicks": "50", "campaign_id": "1", "impressions": "1000", "campaign_name": "Campaign_1"}
	adwords_valueJSON, _ := U.EncodeToPostgresJsonb(&adwords_value)
	adwordsdoc := model.AdwordsDocument{
		ProjectID:         project.ID,
		CustomerAccountID: "123",
		TypeAlias:         "campaign_performance_report",
		Type:              1,
		Timestamp:         20220601,
		ID:                "1",
		Value:             adwords_valueJSON,
	}
	store.GetStore().CreateMultipleAdwordsDocument([]model.AdwordsDocument{adwordsdoc})

	// insert dummy facebook data
	fbCustomerAccountId := U.RandomLowerAphaNumString(5)
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
	_ = store.GetStore().CreateFacebookDocument(project.ID, documentFB)

	// insert dummy linkedin data
	linkedinCustomerAccountId := U.RandomLowerAphaNumString(5)
	value = []byte(`{"costInLocalCurrency": "1","clicks": "1","campaign_group_id":"10000","impressions": "1", "campaign_group_name": "Campaign_Linkedin_10000"}`)
	documentLinkedin := &model.LinkedinDocument{
		ProjectID:           project.ID,
		ID:                  "10000",
		CustomerAdAccountID: linkedinCustomerAccountId,
		TypeAlias:           "campaign_group_insights",
		Timestamp:           20200510,
		Value:               &postgres.Jsonb{RawMessage: value},
		CampaignGroupID:     "10000",
	}
	_ = store.GetStore().CreateLinkedinDocument(project.ID, documentLinkedin)

	// insert dummy hubspot data
	userID1 := U.RandomLowerAphaNumString(5)
	userID2 := U.RandomLowerAphaNumString(5)
	userID3 := U.RandomLowerAphaNumString(5)
	cuid := U.RandomLowerAphaNumString(5)
	_, status := store.GetStore().CreateUser(&model.User{ProjectId: project.ID, ID: userID1, CustomerUserId: cuid, Source: model.GetRequestSourcePointer(model.UserSourceHubspot)})
	assert.Equal(t, http.StatusCreated, status)
	_, status = store.GetStore().CreateUser(&model.User{ProjectId: project.ID, ID: userID2, CustomerUserId: cuid, Source: model.GetRequestSourcePointer(model.UserSourceHubspot)})
	assert.Equal(t, http.StatusCreated, status)
	_, status = store.GetStore().CreateUser(&model.User{ProjectId: project.ID, ID: userID3, CustomerUserId: cuid, Source: model.GetRequestSourcePointer(model.UserSourceHubspot)})
	assert.Equal(t, http.StatusCreated, status)

	createdAt := time.Now().AddDate(0, 0, -11)
	updatedDate := createdAt.AddDate(0, 0, 1)

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

	jsonContact := fmt.Sprintf(jsonContactModel, 1, updatedDate.Unix(), createdAt.UnixNano()/int64(time.Millisecond), updatedDate.UnixNano()/int64(time.Millisecond), "lead", cuid, "123-45")
	contactPJson := postgres.Jsonb{json.RawMessage(jsonContact)}

	hubspotDocument := model.HubspotDocument{
		TypeAlias: model.HubspotDocumentTypeNameContact,
		Value:     &contactPJson,
	}

	store.GetStore().CreateHubspotDocument(project.ID, &hubspotDocument)
	store.GetStore().UpdateHubspotDocumentAsSynced(project.ID, hubspotDocument.ID, model.HubspotDocumentTypeContact, "", hubspotDocument.Timestamp, hubspotDocument.Action, userID1, "")

	// insert dummy salesforce data
	contactID := U.RandomLowerAphaNumString(5)
	userID1 = U.RandomLowerAphaNumString(5)
	userID2 = U.RandomLowerAphaNumString(5)
	userID3 = U.RandomLowerAphaNumString(5)
	cuid = U.RandomLowerAphaNumString(5)
	_, status = store.GetStore().CreateUser(&model.User{ProjectId: project.ID, ID: userID1, CustomerUserId: cuid, Source: model.GetRequestSourcePointer(model.UserSourceSalesforce)})
	assert.Equal(t, http.StatusCreated, status)
	_, status = store.GetStore().CreateUser(&model.User{ProjectId: project.ID, ID: userID2, CustomerUserId: cuid, Source: model.GetRequestSourcePointer(model.UserSourceSalesforce)})
	assert.Equal(t, http.StatusCreated, status)
	_, status = store.GetStore().CreateUser(&model.User{ProjectId: project.ID, ID: userID3, CustomerUserId: cuid, Source: model.GetRequestSourcePointer(model.UserSourceSalesforce)})
	assert.Equal(t, http.StatusCreated, status)

	createdAt = time.Now().AddDate(0, 0, -11)
	updatedDate = createdAt.AddDate(0, 0, -11)
	propertyDay := "Sunday"
	jsonData := fmt.Sprintf(`{"Id":"%s", "day":"%s","CreatedDate":"%s", "LastModifiedDate":"%s"}`, contactID, propertyDay, createdAt.UTC().Format(model.SalesforceDocumentDateTimeLayout), updatedDate.Format(model.SalesforceDocumentDateTimeLayout))
	salesforceDocument := &model.SalesforceDocument{
		ProjectID: project.ID,
		TypeAlias: model.SalesforceDocumentTypeNameContact,
		Value:     &postgres.Jsonb{RawMessage: json.RawMessage([]byte(jsonData))},
	}
	_ = store.GetStore().CreateSalesforceDocument(project.ID, salesforceDocument)
	_ = store.GetStore().UpdateSalesforceDocumentBySyncStatus(project.ID, salesforceDocument, "", userID3, "", true)

	urlPrefix = U.RandomNumericString(10)
	_, _ = store.GetStore().UpdateProjectSettings(project.ID, &model.ProjectSetting{
		IntGoogleOrganicURLPrefixes: &urlPrefix,
	})

	id1 := U.RandomNumericString(8)
	value1, _ := json.Marshal(map[string]interface{}{"query": "factors ai", "clicks": "50", "id": id1, "impressions": "1000"})
	id2 := U.RandomNumericString(8)
	value2, _ := json.Marshal(map[string]interface{}{"query": "factors.ai", "clicks": "100", "id": id2, "impressions": "2000"})

	googleOrganicDocuments := []model.GoogleOrganicDocument{
		{ID: id1, ProjectID: project.ID, URLPrefix: urlPrefix, Timestamp: 20220601,
			Value: &postgres.Jsonb{value1}},

		{ID: id2, ProjectID: project.ID, URLPrefix: urlPrefix, Timestamp: 20220601,
			Value: &postgres.Jsonb{value2}},
	}

	store.GetStore().CreateMultipleGoogleOrganicDocument(googleOrganicDocuments)
	result, _ = store.GetStore().GetLatestDataStatus([]string{"*"}, project.ID, false)
	assert.Equal(t, len(result), 9)
	for _, value := range result {
		assert.NotEqual(t, value.LatestData, 0)
	}

	_, errs := store.GetStore().GetLatestDataStatus([]string{"xyz"}, project.ID, false)
	assert.NotEqual(t, errs, nil)

	isAvailable := store.GetStore().IsDataAvailable(project.ID, model.HUBSPOT, 1654144936)
	assert.Equal(t, isAvailable, false)
	isAvailable = store.GetStore().IsDataAvailable(project.ID, model.HUBSPOT, 1653088333)
	assert.Equal(t, isAvailable, true)

}
