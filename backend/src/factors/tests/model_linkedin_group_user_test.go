package tests

import (
	"encoding/json"
	"factors/model/model"
	"factors/model/store"
	"factors/task"
	U "factors/util"
	"fmt"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestLinkedinCompanyEnagagementEnrichmentV2(t *testing.T) {
	project, _, _ := SetupProjectWithAgentDAO()
	assert.NotNil(t, project)
	customerAccountID := U.RandomNumericString(10)
	_, errCode := store.GetStore().UpdateProjectSettings(project.ID, &model.ProjectSetting{
		IntLinkedinAdAccount: customerAccountID,
	})
	assert.Equal(t, http.StatusAccepted, errCode)

	timestamp := time.Now().Format("20060102")
	intTimestamp, _ := strconv.ParseInt(timestamp, 10, 64)
	currTimeUnix := time.Now().Unix()
	timestamp2daysAgoUnix := currTimeUnix - (2 * 86400)
	errCode = createLinkedinCompanyEngagementDocs(project, customerAccountID, intTimestamp)
	assert.Equal(t, http.StatusCreated, errCode)

	projectID := fmt.Sprint(project.ID)
	projectSetting := model.LinkedinProjectSettings{
		ProjectId:            projectID,
		IntLinkedinAdAccount: customerAccountID,
	}

	errMsg, errCode := task.CreateGroupUserAndEventsV2(projectSetting, 5)
	assert.Equal(t, "", errMsg)
	assert.Equal(t, http.StatusOK, errCode)

	imprEventName, err := store.GetStore().GetEventNameIDFromEventName(U.GROUP_EVENT_NAME_LINKEDIN_VIEWED_AD, project.ID)
	assert.Nil(t, err)
	clickEventName, err := store.GetStore().GetEventNameIDFromEventName(U.GROUP_EVENT_NAME_LINKEDIN_CLICKED_AD, project.ID)
	assert.Nil(t, err)

	imprEvents, errCode := store.GetStore().GetEventsByEventNameIDANDTimeRange(project.ID, imprEventName.ID, timestamp2daysAgoUnix, currTimeUnix)
	assert.Equal(t, http.StatusFound, errCode)
	clickEvents, errCode := store.GetStore().GetEventsByEventNameIDANDTimeRange(project.ID, clickEventName.ID, timestamp2daysAgoUnix, currTimeUnix)
	assert.Equal(t, http.StatusFound, errCode)
	for _, event := range imprEvents {
		propertiesMap, _ := U.DecodePostgresJsonbAsPropertiesMap(&event.Properties)
		userPropertiesMap, _ := U.DecodePostgresJsonbAsPropertiesMap(event.UserProperties)
		key := (*userPropertiesMap)[U.LI_DOMAIN].(string) + (*propertiesMap)[U.EP_CAMPAIGN_ID].(string)
		if expectedEventsMetricsBeforeUpdate[key]["impressions"] != (*propertiesMap)[U.LI_AD_VIEW_COUNT].(float64) {
			log.Fatal(key, propertiesMap, userPropertiesMap)
		}
		assert.Equal(t, expectedEventsMetricsBeforeUpdate[key]["impressions"], (*propertiesMap)[U.LI_AD_VIEW_COUNT].(float64))
	}
	for _, event := range clickEvents {
		propertiesMap, _ := U.DecodePostgresJsonbAsPropertiesMap(&event.Properties)
		userPropertiesMap, _ := U.DecodePostgresJsonbAsPropertiesMap(event.UserProperties)
		key := (*userPropertiesMap)[U.LI_DOMAIN].(string) + (*propertiesMap)[U.EP_CAMPAIGN_ID].(string)
		assert.Equal(t, expectedEventsMetricsBeforeUpdate[key]["clicks"], (*propertiesMap)[U.LI_AD_CLICK_COUNT].(float64))
	}

	users, errCode := store.GetStore().GetGroupUsersGroupIdsByGroupName(project.ID, U.GROUP_NAME_LINKEDIN_COMPANY)
	assert.Equal(t, http.StatusFound, errCode)
	for _, user := range users {
		properties, errCode := store.GetStore().GetUserPropertiesByUserID(project.ID, user.ID)
		assert.Equal(t, errCode, http.StatusFound)

		propertiesMap, _ := U.DecodePostgresJsonbAsPropertiesMap(properties)
		assert.Equal(t, expectedUsersMetricsBeforeUpdate[(*propertiesMap)[U.LI_DOMAIN].(string)]["impressions"], (*propertiesMap)[U.LI_TOTAL_AD_VIEW_COUNT].(float64))
		assert.Equal(t, expectedUsersMetricsBeforeUpdate[(*propertiesMap)[U.LI_DOMAIN].(string)]["clicks"], (*propertiesMap)[U.LI_TOTAL_AD_CLICK_COUNT].(float64))
	}

	domainDataSet, errCode := store.GetStore().GetCompanyDataFromLinkedinDocsForTimestamp(projectSetting.ProjectId, intTimestamp)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, 0, len(domainDataSet))

	errCode = deleteAndUpdateLinkedinCompanyEngagementDocs(project, customerAccountID, intTimestamp)
	assert.Equal(t, http.StatusCreated, errCode)

	errMsg, errCode = task.CreateGroupUserAndEventsV2(projectSetting, 5)
	assert.Equal(t, "", errMsg)
	assert.Equal(t, http.StatusOK, errCode)

	imprEvents, errCode = store.GetStore().GetEventsByEventNameIDANDTimeRange(project.ID, imprEventName.ID, timestamp2daysAgoUnix, currTimeUnix)
	assert.Equal(t, http.StatusFound, errCode)
	clickEvents, errCode = store.GetStore().GetEventsByEventNameIDANDTimeRange(project.ID, clickEventName.ID, timestamp2daysAgoUnix, currTimeUnix)
	assert.Equal(t, http.StatusFound, errCode)
	for _, event := range imprEvents {
		propertiesMap, _ := U.DecodePostgresJsonbAsPropertiesMap(&event.Properties)
		userPropertiesMap, _ := U.DecodePostgresJsonbAsPropertiesMap(event.UserProperties)
		key := (*userPropertiesMap)[U.LI_DOMAIN].(string) + (*propertiesMap)[U.EP_CAMPAIGN_ID].(string)
		assert.Equal(t, expectedEventsMetricsAfterUpdate[key]["impressions"], (*propertiesMap)[U.LI_AD_VIEW_COUNT].(float64))
	}
	for _, event := range clickEvents {
		propertiesMap, _ := U.DecodePostgresJsonbAsPropertiesMap(&event.Properties)
		userPropertiesMap, _ := U.DecodePostgresJsonbAsPropertiesMap(event.UserProperties)
		key := (*userPropertiesMap)[U.LI_DOMAIN].(string) + (*propertiesMap)[U.EP_CAMPAIGN_ID].(string)
		assert.Equal(t, expectedEventsMetricsAfterUpdate[key]["clicks"], (*propertiesMap)[U.LI_AD_CLICK_COUNT].(float64))
	}

	users, errCode = store.GetStore().GetGroupUsersGroupIdsByGroupName(project.ID, U.GROUP_NAME_LINKEDIN_COMPANY)
	assert.Equal(t, http.StatusFound, errCode)
	for _, user := range users {
		properties, errCode := store.GetStore().GetUserPropertiesByUserID(project.ID, user.ID)
		assert.Equal(t, errCode, http.StatusFound)

		propertiesMap, _ := U.DecodePostgresJsonbAsPropertiesMap(properties)

		assert.Equal(t, expectedUsersMetricsAfterUpdate[(*propertiesMap)[U.LI_DOMAIN].(string)]["impressions"], (*propertiesMap)[U.LI_TOTAL_AD_VIEW_COUNT].(float64))
		assert.Equal(t, expectedUsersMetricsAfterUpdate[(*propertiesMap)[U.LI_DOMAIN].(string)]["clicks"], (*propertiesMap)[U.LI_TOTAL_AD_CLICK_COUNT].(float64))
	}

	domainDataSet, errCode = store.GetStore().GetCompanyDataFromLinkedinDocsForTimestamp(projectSetting.ProjectId, intTimestamp)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, 0, len(domainDataSet))
}

func TestLinkedinCompanyEnagagementEnrichmentV1(t *testing.T) {
	project, _, _ := SetupProjectWithAgentDAO()
	assert.NotNil(t, project)
	customerAccountID := U.RandomNumericString(10)
	_, errCode := store.GetStore().UpdateProjectSettings(project.ID, &model.ProjectSetting{
		IntLinkedinAdAccount: customerAccountID,
	})
	assert.Equal(t, http.StatusAccepted, errCode)

	timestamp := time.Now().Format("20060102")
	intTimestamp, _ := strconv.ParseInt(timestamp, 10, 64)
	errCode = createLinkedinCompanyEngagementDocs(project, customerAccountID, intTimestamp)
	assert.Equal(t, http.StatusCreated, errCode)

	projectID := fmt.Sprint(project.ID)
	projectSetting := model.LinkedinProjectSettings{
		ProjectId:            projectID,
		IntLinkedinAdAccount: customerAccountID,
	}

	errMsg, errCode := task.CreateGroupUserAndEventsV1(projectSetting, 5)
	assert.Equal(t, "", errMsg)
	assert.Equal(t, http.StatusOK, errCode)

	domainDataSet, errCode := store.GetStore().GetCompanyDataFromLinkedinDocsForTimestamp(projectSetting.ProjectId, intTimestamp)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, 0, len(domainDataSet))

}

func TestLinkedinCompanyEnagagementEnrichment(t *testing.T) {
	project, _, _ := SetupProjectWithAgentDAO()
	assert.NotNil(t, project)
	customerAccountID := U.RandomNumericString(10)
	_, errCode := store.GetStore().UpdateProjectSettings(project.ID, &model.ProjectSetting{
		IntLinkedinAdAccount: customerAccountID,
	})
	assert.Equal(t, http.StatusAccepted, errCode)

	timestamp := time.Now().Format("20060102")
	intTimestamp, _ := strconv.ParseInt(timestamp, 10, 64)
	errCode = createLinkedinCompanyEngagementDocs(project, customerAccountID, intTimestamp)
	assert.Equal(t, http.StatusCreated, errCode)

	projectID := fmt.Sprint(project.ID)
	projectSetting := model.LinkedinProjectSettings{
		ProjectId:            projectID,
		IntLinkedinAdAccount: customerAccountID,
	}

	errMsg, errCode := task.CreateGroupUserAndEvents(projectSetting)
	assert.Equal(t, "", errMsg)
	assert.Equal(t, http.StatusOK, errCode)

	domainDataSet, errCode := store.GetStore().GetCompanyDataFromLinkedinDocsForTimestamp(projectSetting.ProjectId, intTimestamp)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, 0, len(domainDataSet))

}

var expectedEventsMetricsBeforeUpdate = map[string]map[string]float64{
	"google.com123":   {"impressions": float64(1000), "clicks": float64(50)},
	"google.com345":   {"impressions": float64(1000), "clicks": float64(50)},
	"linkedin.com123": {"impressions": float64(2000), "clicks": float64(100)},
	"linkedin.com345": {"impressions": float64(2000), "clicks": float64(100)},
	"pqr.com123":      {"impressions": float64(1000), "clicks": float64(50)},
}
var expectedEventsMetricsAfterUpdate = map[string]map[string]float64{
	"google.com123":   {"impressions": float64(1000), "clicks": float64(500)},
	"google.com345":   {"impressions": float64(1000), "clicks": float64(500)},
	"linkedin.com123": {"impressions": float64(2000), "clicks": float64(1000)},
	"linkedin.com345": {"impressions": float64(2000), "clicks": float64(1000)},
	"pqr.com123":      {"impressions": float64(1000), "clicks": float64(500)},
}
var expectedUsersMetricsBeforeUpdate = map[string]map[string]float64{
	"google.com":   {"impressions": float64(2000), "clicks": float64(100)},
	"linkedin.com": {"impressions": float64(4000), "clicks": float64(200)},
	"pqr.com":      {"impressions": float64(1000), "clicks": float64(50)},
}
var expectedUsersMetricsAfterUpdate = map[string]map[string]float64{
	"google.com":   {"impressions": float64(2000), "clicks": float64(1000)},
	"linkedin.com": {"impressions": float64(4000), "clicks": float64(2000)},
	"pqr.com":      {"impressions": float64(1000), "clicks": float64(500)},
}

func createLinkedinCompanyEngagementDocs(project *model.Project, customerAccountID string, intTimestamp int64) int {

	campaignID1 := "123"
	campaignID2 := "345"
	row1, _ := json.Marshal(map[string]interface{}{"costInLocalCurrency": "100", "clicks": "50", "campaign_group_id": campaignID1, "impressions": "1000", "vanityName": "Org1", "preferredCountry": "US", "localizedWebsite": "google.com", "localizedName": "Org_1", "companyHeadquarters": "US", "campaign_group_name": "CG1"})
	row2, _ := json.Marshal(map[string]interface{}{"costInLocalCurrency": "200", "clicks": "100", "campaign_group_id": campaignID1, "impressions": "2000", "vanityName": "Org2", "preferredCountry": "US", "localizedWebsite": "linkedin.com", "localizedName": "Org_2", "companyHeadquarters": "US", "campaign_group_name": "CG1"})
	row3, _ := json.Marshal(map[string]interface{}{"costInLocalCurrency": "100", "clicks": "50", "campaign_group_id": campaignID1, "impressions": "1000", "vanityName": "Org3", "preferredCountry": "US", "localizedWebsite": "pqr.com", "localizedName": "Org_3", "companyHeadquarters": "US", "campaign_group_name": "CG1"})
	row4, _ := json.Marshal(map[string]interface{}{"costInLocalCurrency": "100", "clicks": "50", "campaign_group_id": campaignID2, "impressions": "1000", "vanityName": "Org1", "preferredCountry": "US", "localizedWebsite": "google.com", "localizedName": "Org_1", "companyHeadquarters": "US", "campaign_group_name": "CG2"})
	row5, _ := json.Marshal(map[string]interface{}{"costInLocalCurrency": "200", "clicks": "100", "campaign_group_id": campaignID2, "impressions": "2000", "vanityName": "Org2", "preferredCountry": "US", "localizedWebsite": "linkedin.com", "localizedName": "Org_2", "companyHeadquarters": "US", "campaign_group_name": "CG2"})

	linkedinDocuments := []model.LinkedinDocument{
		{ID: "1", CampaignGroupID: campaignID1, ProjectID: project.ID, CustomerAdAccountID: customerAccountID, TypeAlias: "member_company_insights", Timestamp: intTimestamp,
			Value: &postgres.Jsonb{RawMessage: row1}},
		{ID: "2", CampaignGroupID: campaignID1, ProjectID: project.ID, CustomerAdAccountID: customerAccountID, TypeAlias: "member_company_insights", Timestamp: intTimestamp,
			Value: &postgres.Jsonb{RawMessage: row2}},
		{ID: "3", CampaignGroupID: campaignID1, ProjectID: project.ID, CustomerAdAccountID: customerAccountID, TypeAlias: "member_company_insights", Timestamp: intTimestamp,
			Value: &postgres.Jsonb{RawMessage: row3}},
		{ID: "1", CampaignGroupID: campaignID2, ProjectID: project.ID, CustomerAdAccountID: customerAccountID, TypeAlias: "member_company_insights", Timestamp: intTimestamp,
			Value: &postgres.Jsonb{RawMessage: row4}},
		{ID: "2", CampaignGroupID: campaignID2, ProjectID: project.ID, CustomerAdAccountID: customerAccountID, TypeAlias: "member_company_insights", Timestamp: intTimestamp,
			Value: &postgres.Jsonb{RawMessage: row5}},
	}

	return store.GetStore().CreateMultipleLinkedinDocument(linkedinDocuments)
}

func deleteAndUpdateLinkedinCompanyEngagementDocs(project *model.Project, customerAccountID string, intTimestamp int64) int {
	errCode := store.GetStore().DeleteLinkedinDocuments(model.LinkedinDeleteDocumentsPayload{
		ProjectID:           project.ID,
		CustomerAdAccountID: customerAccountID,
		Timestamp:           intTimestamp,
		TypeAlias:           "member_company_insights",
	})
	if errCode != http.StatusAccepted {
		return errCode
	}

	campaignID1 := "123"
	campaignID2 := "345"
	row1, _ := json.Marshal(map[string]interface{}{"costInLocalCurrency": "1000", "clicks": "500", "campaign_group_id": campaignID1, "impressions": "1000", "vanityName": "Org1", "preferredCountry": "US", "localizedWebsite": "google.com", "localizedName": "Org_1", "companyHeadquarters": "US", "campaign_group_name": "CG1"})
	row2, _ := json.Marshal(map[string]interface{}{"costInLocalCurrency": "2000", "clicks": "1000", "campaign_group_id": campaignID1, "impressions": "2000", "vanityName": "Org2", "preferredCountry": "US", "localizedWebsite": "linkedin.com", "localizedName": "Org_2", "companyHeadquarters": "US", "campaign_group_name": "CG1"})
	row3, _ := json.Marshal(map[string]interface{}{"costInLocalCurrency": "1000", "clicks": "500", "campaign_group_id": campaignID1, "impressions": "1000", "vanityName": "Org3", "preferredCountry": "US", "localizedWebsite": "pqr.com", "localizedName": "Org_3", "companyHeadquarters": "US", "campaign_group_name": "CG1"})
	row4, _ := json.Marshal(map[string]interface{}{"costInLocalCurrency": "1000", "clicks": "500", "campaign_group_id": campaignID2, "impressions": "1000", "vanityName": "Org1", "preferredCountry": "US", "localizedWebsite": "google.com", "localizedName": "Org_1", "companyHeadquarters": "US", "campaign_group_name": "CG2"})
	row5, _ := json.Marshal(map[string]interface{}{"costInLocalCurrency": "2000", "clicks": "1000", "campaign_group_id": campaignID2, "impressions": "2000", "vanityName": "Org2", "preferredCountry": "US", "localizedWebsite": "linkedin.com", "localizedName": "Org_2", "companyHeadquarters": "US", "campaign_group_name": "CG2"})

	linkedinDocuments := []model.LinkedinDocument{
		{ID: "1", CampaignGroupID: campaignID1, ProjectID: project.ID, CustomerAdAccountID: customerAccountID, TypeAlias: "member_company_insights", Timestamp: intTimestamp,
			Value: &postgres.Jsonb{RawMessage: row1}},
		{ID: "2", CampaignGroupID: campaignID1, ProjectID: project.ID, CustomerAdAccountID: customerAccountID, TypeAlias: "member_company_insights", Timestamp: intTimestamp,
			Value: &postgres.Jsonb{RawMessage: row2}},
		{ID: "3", CampaignGroupID: campaignID1, ProjectID: project.ID, CustomerAdAccountID: customerAccountID, TypeAlias: "member_company_insights", Timestamp: intTimestamp,
			Value: &postgres.Jsonb{RawMessage: row3}},
		{ID: "1", CampaignGroupID: campaignID2, ProjectID: project.ID, CustomerAdAccountID: customerAccountID, TypeAlias: "member_company_insights", Timestamp: intTimestamp,
			Value: &postgres.Jsonb{RawMessage: row4}},
		{ID: "2", CampaignGroupID: campaignID2, ProjectID: project.ID, CustomerAdAccountID: customerAccountID, TypeAlias: "member_company_insights", Timestamp: intTimestamp,
			Value: &postgres.Jsonb{RawMessage: row5}},
	}

	return store.GetStore().CreateMultipleLinkedinDocument(linkedinDocuments)
}
