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
	"github.com/stretchr/testify/assert"
)

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

	domainDataSet, errCode := store.GetStore().GetCompanyDataFromLinkedinForTimestamp(projectSetting.ProjectId, intTimestamp)
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

	domainDataSet, errCode := store.GetStore().GetCompanyDataFromLinkedinForTimestamp(projectSetting.ProjectId, intTimestamp)
	assert.Equal(t, http.StatusOK, errCode)
	assert.Equal(t, 0, len(domainDataSet))

}

func createLinkedinCompanyEngagementDocs(project *model.Project, customerAccountID string, intTimestamp int64) int {

	campaignID1 := U.RandomNumericString(8)
	campaignID2 := U.RandomNumericString(8)
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
