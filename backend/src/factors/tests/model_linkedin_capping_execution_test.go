package tests

import (
	"encoding/json"
	"factors/model/model"
	"factors/model/store"
	LCE "factors/task/linkedin_company_engagements"
	LFC "factors/task/linkedin_frequency_capping"
	U "factors/util"
	"fmt"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/stretchr/testify/assert"
)

func TestFrequencyCappingRuleExecution(t *testing.T) {
	project, _, _ := SetupProjectWithAgentDAO()
	assert.NotNil(t, project)
	customerAccountID := U.RandomNumericString(10)
	_, errCode := store.GetStore().UpdateProjectSettings(project.ID, &model.ProjectSetting{
		IntLinkedinAdAccount: customerAccountID,
	})
	assert.Equal(t, http.StatusAccepted, errCode)

	timestamp := time.Now().Format("20060102")
	intTimestamp, _ := strconv.ParseInt(timestamp, 10, 64)
	errCode = createLinkedinCompanyEngagementDocsForFrequencyCapping(project.ID, customerAccountID, intTimestamp)
	assert.Equal(t, http.StatusCreated, errCode)

	projectID := fmt.Sprint(project.ID)
	projectSetting := model.LinkedinProjectSettings{
		ProjectId:            projectID,
		IntLinkedinAdAccount: customerAccountID,
	}

	errMsg, errCode := LCE.CreateGroupUserAndEventsV3(projectSetting, 1)
	assert.Equal(t, "", errMsg)
	assert.Equal(t, http.StatusOK, errCode)

	exclusionsMap := LFC.GetExistingExclusionsInDataStore()
	assert.Equal(t, 0, len(*exclusionsMap))

	currMonthStart, currentDateInt := LFC.GetStartAndTodayDateForCurrMonth()
	errMsg, errCode = LFC.ApplyRulesAndGetRuleMatchedDataset(project.ID, currMonthStart, currentDateInt)
	assert.Equal(t, "", errMsg)
	assert.Equal(t, http.StatusOK, errCode)
	errMsg, errCode = LFC.BuildAndPushExclusionToDBFromMatchedRules(project.ID, currentDateInt)
	assert.Equal(t, "", errMsg)
	assert.Equal(t, http.StatusOK, errCode)

	exclusions, errCode := store.GetStore().GetAllLinkedinCappingExclusionsForTimerange(project.ID, currMonthStart, currentDateInt)
	assert.Equal(t, http.StatusFound, errCode)
	assert.Equal(t, 3, len(exclusions))
}

func createLinkedinCompanyEngagementDocsForFrequencyCapping(projectID int64, customerAccountID string, intTimestamp int64) int {
	adAccountID := "01"
	adGroupID1 := "11"
	adGroupID2 := "12"
	adGroupID3 := "13"
	adGroupID4 := "14"
	adGroup1Name := "ag1"
	adGroup2Name := "ag2"
	adGroup3Name := "ag3"
	adGroup4Name := "ag4"
	campaignID1 := "1"
	campaignID2 := "2"
	campaign1Name := "cg1"
	campaign2Name := "cg2"
	orgID1 := "111"
	orgID2 := "112"
	org1Name := "o1"
	org2Name := "o2"
	org1Domain := "factors1.com"
	org2Domain := "factors2.com"

	ag1, _ := json.Marshal(map[string]interface{}{"campaign_group_id": campaignID1, "campaign_group_name": campaign1Name, "campaign_id": adGroupID1, "campaign_name": adGroup1Name})
	ag2, _ := json.Marshal(map[string]interface{}{"campaign_group_id": campaignID1, "campaign_group_name": campaign1Name, "campaign_id": adGroupID2, "campaign_name": adGroup2Name})
	ag3, _ := json.Marshal(map[string]interface{}{"campaign_group_id": campaignID2, "campaign_group_name": campaign2Name, "campaign_id": adGroupID3, "campaign_name": adGroup3Name})
	ag4, _ := json.Marshal(map[string]interface{}{"campaign_group_id": campaignID2, "campaign_group_name": campaign2Name, "campaign_id": adGroupID4, "campaign_name": adGroup4Name})

	c1, _ := json.Marshal(map[string]interface{}{"campaign_group_id": campaignID1, "campaign_group_name": campaign1Name})
	c2, _ := json.Marshal(map[string]interface{}{"campaign_group_id": campaignID2, "campaign_group_name": campaign2Name})

	row1, _ := json.Marshal(map[string]interface{}{"clicks": "50", "impressions": "1000", "companyHeadquarters": "US", "preferredCountry": "US", "campaign_group_id": campaignID1, "campaign_group_name": campaign1Name, "campaign_id": adGroupID1, "campaign_name": adGroup1Name, "vanityName": org1Name, "localizedWebsite": org1Domain, "localizedName": org1Name})
	row2, _ := json.Marshal(map[string]interface{}{"clicks": "100", "impressions": "2000", "companyHeadquarters": "IN", "preferredCountry": "IN", "campaign_group_id": campaignID1, "campaign_group_name": campaign1Name, "campaign_id": adGroupID1, "campaign_name": adGroup1Name, "vanityName": org2Name, "localizedWebsite": org2Domain, "localizedName": org2Name})

	row3, _ := json.Marshal(map[string]interface{}{"clicks": "50", "impressions": "1000", "companyHeadquarters": "US", "preferredCountry": "US", "campaign_group_id": campaignID1, "campaign_group_name": campaign1Name, "campaign_id": adGroupID2, "campaign_name": adGroup2Name, "vanityName": org1Name, "localizedWebsite": org1Domain, "localizedName": org1Name})
	row4, _ := json.Marshal(map[string]interface{}{"clicks": "100", "impressions": "2000", "companyHeadquarters": "IN", "preferredCountry": "IN", "campaign_group_id": campaignID1, "campaign_group_name": campaign1Name, "campaign_id": adGroupID2, "campaign_name": adGroup2Name, "vanityName": org2Name, "localizedWebsite": org2Domain, "localizedName": org2Name})

	row5, _ := json.Marshal(map[string]interface{}{"clicks": "50", "impressions": "1000", "companyHeadquarters": "US", "preferredCountry": "US", "campaign_group_id": campaignID2, "campaign_group_name": campaign2Name, "campaign_id": adGroupID3, "campaign_name": adGroup3Name, "vanityName": org1Name, "localizedWebsite": org1Domain, "localizedName": org1Name})
	row6, _ := json.Marshal(map[string]interface{}{"clicks": "100", "impressions": "2000", "companyHeadquarters": "IN", "preferredCountry": "IN", "campaign_group_id": campaignID2, "campaign_group_name": campaign2Name, "campaign_id": adGroupID3, "campaign_name": adGroup3Name, "vanityName": org2Name, "localizedWebsite": org2Domain, "localizedName": org2Name})

	row7, _ := json.Marshal(map[string]interface{}{"clicks": "5", "impressions": "100", "companyHeadquarters": "US", "preferredCountry": "US", "campaign_group_id": campaignID2, "campaign_group_name": campaign2Name, "campaign_id": adGroupID4, "campaign_name": adGroup4Name, "vanityName": org1Name, "localizedWebsite": org1Domain, "localizedName": org1Name})
	row8, _ := json.Marshal(map[string]interface{}{"clicks": "10", "impressions": "200", "companyHeadquarters": "IN", "preferredCountry": "IN", "campaign_group_id": campaignID2, "campaign_group_name": campaign2Name, "campaign_id": adGroupID4, "campaign_name": adGroup4Name, "vanityName": org2Name, "localizedWebsite": org2Domain, "localizedName": org2Name})

	metadataDocuments := []model.LinkedinDocument{
		{ID: campaignID1, CampaignGroupID: campaignID1, ProjectID: projectID, CustomerAdAccountID: customerAccountID, TypeAlias: "campaign_group", Timestamp: intTimestamp,
			Value: &postgres.Jsonb{RawMessage: c1}},
		{ID: campaignID2, CampaignGroupID: campaignID2, ProjectID: projectID, CustomerAdAccountID: customerAccountID, TypeAlias: "campaign_group", Timestamp: intTimestamp,
			Value: &postgres.Jsonb{RawMessage: c2}},
		{ID: adGroupID1, CampaignID: adGroupID1, CampaignGroupID: campaignID1, ProjectID: projectID, CustomerAdAccountID: customerAccountID, TypeAlias: "campaign", Timestamp: intTimestamp,
			Value: &postgres.Jsonb{RawMessage: ag1}},
		{ID: adGroupID2, CampaignID: adGroupID2, CampaignGroupID: campaignID1, ProjectID: projectID, CustomerAdAccountID: customerAccountID, TypeAlias: "campaign", Timestamp: intTimestamp,
			Value: &postgres.Jsonb{RawMessage: ag2}},
		{ID: adGroupID3, CampaignID: adGroupID3, CampaignGroupID: campaignID2, ProjectID: projectID, CustomerAdAccountID: customerAccountID, TypeAlias: "campaign", Timestamp: intTimestamp,
			Value: &postgres.Jsonb{RawMessage: ag3}},
		{ID: adGroupID4, CampaignID: adGroupID4, CampaignGroupID: campaignID1, ProjectID: projectID, CustomerAdAccountID: customerAccountID, TypeAlias: "campaign", Timestamp: intTimestamp,
			Value: &postgres.Jsonb{RawMessage: ag4}},
	}
	orgDocuments := []model.LinkedinDocument{
		{ID: orgID1, CampaignID: adGroupID1, CampaignGroupID: campaignID1, ProjectID: projectID, CustomerAdAccountID: customerAccountID, TypeAlias: "member_company_insights", Timestamp: intTimestamp,
			Value: &postgres.Jsonb{RawMessage: row1}},
		{ID: orgID2, CampaignID: adGroupID1, CampaignGroupID: campaignID1, ProjectID: projectID, CustomerAdAccountID: customerAccountID, TypeAlias: "member_company_insights", Timestamp: intTimestamp,
			Value: &postgres.Jsonb{RawMessage: row2}},
		{ID: orgID1, CampaignID: adGroupID2, CampaignGroupID: campaignID1, ProjectID: projectID, CustomerAdAccountID: customerAccountID, TypeAlias: "member_company_insights", Timestamp: intTimestamp,
			Value: &postgres.Jsonb{RawMessage: row3}},
		{ID: orgID2, CampaignID: adGroupID2, CampaignGroupID: campaignID1, ProjectID: projectID, CustomerAdAccountID: customerAccountID, TypeAlias: "member_company_insights", Timestamp: intTimestamp,
			Value: &postgres.Jsonb{RawMessage: row4}},
		{ID: orgID1, CampaignID: adGroupID3, CampaignGroupID: campaignID2, ProjectID: projectID, CustomerAdAccountID: customerAccountID, TypeAlias: "member_company_insights", Timestamp: intTimestamp,
			Value: &postgres.Jsonb{RawMessage: row5}},
		{ID: orgID2, CampaignID: adGroupID3, CampaignGroupID: campaignID2, ProjectID: projectID, CustomerAdAccountID: customerAccountID, TypeAlias: "member_company_insights", Timestamp: intTimestamp,
			Value: &postgres.Jsonb{RawMessage: row6}},
		{ID: orgID1, CampaignID: adGroupID4, CampaignGroupID: campaignID2, ProjectID: projectID, CustomerAdAccountID: customerAccountID, TypeAlias: "member_company_insights", Timestamp: intTimestamp,
			Value: &postgres.Jsonb{RawMessage: row7}},
		{ID: orgID2, CampaignID: adGroupID4, CampaignGroupID: campaignID2, ProjectID: projectID, CustomerAdAccountID: customerAccountID, TypeAlias: "member_company_insights", Timestamp: intTimestamp,
			Value: &postgres.Jsonb{RawMessage: row8}},
	}
	linkedinDocuments := append(metadataDocuments, orgDocuments...)

	errCode := store.GetStore().CreateMultipleLinkedinDocument(linkedinDocuments)
	if errCode != http.StatusCreated {
		return errCode
	}

	campaignObjectIDs, _ := json.Marshal([]string{campaignID1})
	adGroupObjectIDs, _ := json.Marshal([]string{adGroupID1, adGroupID2, adGroupID3, adGroupID4})
	advrule, _ := json.Marshal([]model.AdvancedRuleFilters{{
		Filters: []model.QueryProperty{
			{
				Type:      U.PropertyTypeCategorical,
				Property:  U.LI_DOMAIN,
				Operator:  model.EqualsOpStr,
				Value:     org1Domain,
				LogicalOp: model.LOGICAL_OP_AND,
				Entity:    model.PropertyEntityUserGlobal,
			},
		},
		ImpressionThreshold: 100,
		ClickThreshold:      10,
	},
	})
	rules := []*model.LinkedinCappingRule{
		{
			ProjectID:             projectID,
			ObjectType:            model.LINKEDIN_ACCOUNT,
			DisplayName:           "Default criteria rule account level",
			Description:           "rule_1",
			Status:                model.LINKEDIN_STATUS_ACTIVE,
			Granularity:           "monthly",
			ImpressionThreshold:   5000,
			ClickThreshold:        5000,
			IsAdvancedRuleEnabled: false,
		},
		{
			ProjectID:             projectID,
			ObjectType:            model.LINKEDIN_CAMPAIGN_GROUP,
			ObjectIDs:             &postgres.Jsonb{RawMessage: campaignObjectIDs},
			DisplayName:           "Default criteria rule campaign group level",
			Description:           "rule_2",
			Status:                model.LINKEDIN_STATUS_ACTIVE,
			Granularity:           "monthly",
			ImpressionThreshold:   2000,
			ClickThreshold:        3000,
			IsAdvancedRuleEnabled: false,
		},

		{
			ProjectID:             projectID,
			ObjectType:            model.LINKEDIN_CAMPAIGN,
			ObjectIDs:             &postgres.Jsonb{RawMessage: adGroupObjectIDs},
			DisplayName:           "Advanced criteria rule campaign level",
			Description:           "rule_3",
			Status:                model.LINKEDIN_STATUS_ACTIVE,
			Granularity:           "monthly",
			ImpressionThreshold:   10000,
			ClickThreshold:        1000,
			IsAdvancedRuleEnabled: true,
			AdvancedRuleType:      model.LINKEDIN_ACCOUNT,
			AdvancedRules:         &postgres.Jsonb{RawMessage: advrule},
		},
	}

	for _, rule := range rules {
		_, _, errCode = store.GetStore().CreateLinkedinCappingRule(projectID, rule)
		if errCode != http.StatusCreated {
			return errCode
		}
	}

	campaigns := map[string]model.CampaignNameTargetingCriteria{
		adGroupID1: model.CampaignNameTargetingCriteria{
			CampaignName:      adGroup1Name,
			AdAccountID:       adAccountID,
			CampaignGroupID:   campaignID1,
			TargetingCriteria: make(map[string]bool),
		},
		adGroupID2: model.CampaignNameTargetingCriteria{
			CampaignName:      adGroup2Name,
			AdAccountID:       adAccountID,
			CampaignGroupID:   campaignID1,
			TargetingCriteria: make(map[string]bool),
		},
		adGroupID3: model.CampaignNameTargetingCriteria{
			CampaignName:      adGroup3Name,
			AdAccountID:       adAccountID,
			CampaignGroupID:   campaignID2,
			TargetingCriteria: make(map[string]bool),
		},
		adGroupID4: model.CampaignNameTargetingCriteria{
			CampaignName:      adGroup4Name,
			AdAccountID:       adAccountID,
			CampaignGroupID:   campaignID2,
			TargetingCriteria: make(map[string]bool),
		},
	}
	LFC.SetCampaignTargetingCriteriaMapInDataStore(campaigns)
	return http.StatusCreated
}
