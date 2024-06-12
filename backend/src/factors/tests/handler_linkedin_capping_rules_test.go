package tests

import (
	"encoding/json"
	C "factors/config"
	H "factors/handler"
	"factors/handler/helpers"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
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

func TestLinkedinCappingRuleHandler(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)

	project, agent, _ := SetupProjectWithAgentDAO()
	assert.NotNil(t, project)
	customerAccountID := U.RandomNumericString(10)
	_, errCode := store.GetStore().UpdateProjectSettings(project.ID, &model.ProjectSetting{
		IntLinkedinAdAccount: customerAccountID,
	})
	assert.Equal(t, http.StatusAccepted, errCode)

	name1 := U.RandomString(8)
	name2 := U.RandomString(8)
	ruleID1 := ""
	ruleID2 := ""

	timestamp := time.Now().Format("20060102")
	intTimestamp, _ := strconv.ParseInt(timestamp, 10, 64)
	errCode = createLinkedinCompanyEngagementDocsForFrequencyCappingHandler(project.ID, customerAccountID, intTimestamp)
	assert.Equal(t, http.StatusCreated, errCode)

	t.Run("CreateLinkedinCappingRuleFailure:AccountLevelWithObjectIDS", func(t *testing.T) {
		name1 := U.RandomString(8)
		description := U.RandomString(20)
		rules := &postgres.Jsonb{RawMessage: json.RawMessage(`[]`)}
		objectIDs := &postgres.Jsonb{RawMessage: json.RawMessage(`["cg_123"]`)}
		w := sendCreateLinkedinCappingRuleReq(r, project.ID, agent, model.LINKEDIN_ACCOUNT, name1,
			description, model.LINKEDIN_STATUS_ACTIVE, objectIDs, false, model.LINKEDIN_ACCOUNT, rules)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("CreateLinkedinCappingRuleFailure:CampaignLevelWihoutObjectIDs", func(t *testing.T) {
		name1 := U.RandomString(8)
		description := U.RandomString(20)
		rules := &postgres.Jsonb{RawMessage: json.RawMessage(`[]`)}
		objectIDs := &postgres.Jsonb{RawMessage: json.RawMessage(`[]`)}
		w := sendCreateLinkedinCappingRuleReq(r, project.ID, agent, model.LINKEDIN_CAMPAIGN, name1,
			description, model.LINKEDIN_STATUS_ACTIVE, objectIDs, false, model.LINKEDIN_ACCOUNT, rules)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("CreateLinkedinCappingRuleFailure:CampaignGroupLevelWihoutObjectIDs", func(t *testing.T) {
		name1 := U.RandomString(8)
		description := U.RandomString(20)
		rules := &postgres.Jsonb{RawMessage: json.RawMessage(`[]`)}
		objectIDs := &postgres.Jsonb{RawMessage: json.RawMessage(`[]`)}
		w := sendCreateLinkedinCappingRuleReq(r, project.ID, agent, model.LINKEDIN_CAMPAIGN_GROUP, name1,
			description, model.LINKEDIN_STATUS_ACTIVE, objectIDs, false, model.LINKEDIN_ACCOUNT, rules)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("CreateLinkedinCappingRuleSuccessAccountLevel", func(t *testing.T) {
		description := U.RandomString(20)
		rules, _ := json.Marshal([]model.AdvancedRuleFilters{{
			Filters: []model.QueryProperty{
				{
					Type:      U.PropertyTypeCategorical,
					Property:  U.DP_ENGAGEMENT_LEVEL,
					Operator:  model.EqualsOp,
					Value:     model.ENGAGEMENT_LEVEL_HOT,
					LogicalOp: model.LOGICAL_OP_AND,
					Entity:    model.PropertyEntityUserGlobal,
				},
			},
			ImpressionThreshold: 1000,
			ClickThreshold:      100,
		},
		})
		w := sendCreateLinkedinCappingRuleReq(r, project.ID, agent, model.LINKEDIN_ACCOUNT, name1,
			description, model.LINKEDIN_STATUS_ACTIVE, nil, false, model.LINKEDIN_ACCOUNT, &postgres.Jsonb{RawMessage: rules})
		assert.Equal(t, http.StatusOK, w.Code)
		var result model.LinkedinCappingRule
		decoder := json.NewDecoder(w.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&result); err != nil {
			assert.NotNil(t, nil, err)
		}
		ruleID1 = result.ID
	})
	t.Run("CreateLinkedinCappingRuleFailure:RepeatedName", func(t *testing.T) {
		description := U.RandomString(20)
		rules := &postgres.Jsonb{RawMessage: json.RawMessage(`[]`)}
		objectIDs := &postgres.Jsonb{RawMessage: json.RawMessage(`[]`)}
		w := sendCreateLinkedinCappingRuleReq(r, project.ID, agent, model.LINKEDIN_ACCOUNT, name1,
			description, model.LINKEDIN_STATUS_ACTIVE, objectIDs, false, model.LINKEDIN_ACCOUNT, rules)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("GetLinkedinCappingRuleReq", func(t *testing.T) {
		w := sendGetLinkedinCappingRulesReq(r, project.ID, agent)
		var result []model.LinkedinCappingRule
		decoder := json.NewDecoder(w.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&result); err != nil {
			assert.NotNil(t, nil, err)
		}
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, 1, len(result))
		assert.Equal(t, ruleID1, result[0].ID)
	})
	t.Run("GetLinkedinCappingRuleReq", func(t *testing.T) {
		w := sendGetLinkedinCappingRuleReq(r, project.ID, ruleID1, agent)
		var result model.LinkedinCappingRule
		decoder := json.NewDecoder(w.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&result); err != nil {
			assert.NotNil(t, nil, err)
		}
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, ruleID1, result.ID)
	})

	t.Run("CreateLinkedinCappingRuleFailure:RepeatedNameWithSpecialChar", func(t *testing.T) {
		description := U.RandomString(20)
		rules := &postgres.Jsonb{RawMessage: json.RawMessage(`[]`)}
		objectIDs := &postgres.Jsonb{RawMessage: json.RawMessage(`[]`)}
		w := sendCreateLinkedinCappingRuleReq(r, project.ID, agent, model.LINKEDIN_ACCOUNT, name1+"$",
			description, model.LINKEDIN_STATUS_ACTIVE, objectIDs, false, model.LINKEDIN_ACCOUNT, rules)
		assert.Equal(t, http.StatusOK, w.Code)

		w = sendCreateLinkedinCappingRuleReq(r, project.ID, agent, model.LINKEDIN_ACCOUNT, name1+"@",
			description, model.LINKEDIN_STATUS_ACTIVE, objectIDs, false, model.LINKEDIN_ACCOUNT, rules)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("CreateLinkedinCappingRuleSuccess:CampaignLevel", func(t *testing.T) {
		description := U.RandomString(20)
		rules := &postgres.Jsonb{RawMessage: json.RawMessage(`[]`)}
		objectIDs := &postgres.Jsonb{RawMessage: json.RawMessage(`["11"]`)}
		w := sendCreateLinkedinCappingRuleReq(r, project.ID, agent, model.LINKEDIN_CAMPAIGN, name2,
			description, model.LINKEDIN_STATUS_ACTIVE, objectIDs, false, model.LINKEDIN_ACCOUNT, rules)
		assert.Equal(t, http.StatusOK, w.Code)
		var result model.LinkedinCappingRule
		decoder := json.NewDecoder(w.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&result); err != nil {
			assert.NotNil(t, nil, err)
		}
		assert.Equal(t, model.LINKEDIN_CAMPAIGN, result.ObjectType)
		ruleID2 = result.ID
	})

	t.Run("GetLinkedinCappingRuleReq", func(t *testing.T) {
		w := sendGetLinkedinCappingRuleReq(r, project.ID, ruleID2, agent)
		var result model.LinkedinCappingRule
		decoder := json.NewDecoder(w.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&result); err != nil {
			assert.NotNil(t, nil, err)
		}
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, ruleID2, result.ID)
	})

	t.Run("UpdateLinkedinCappingRule:Success", func(t *testing.T) {
		description := U.RandomString(20)
		rules := &postgres.Jsonb{RawMessage: json.RawMessage(`[]`)}
		objectIDs := &postgres.Jsonb{RawMessage: json.RawMessage(`[]`)}
		w := sendUpdateLinkedinCappingRuleReq(r, project.ID, ruleID2, agent, model.LINKEDIN_ACCOUNT, name2,
			description, model.LINKEDIN_STATUS_ACTIVE, objectIDs, false, model.LINKEDIN_ACCOUNT, rules)
		assert.Equal(t, http.StatusOK, w.Code)
		var result model.LinkedinCappingRule
		decoder := json.NewDecoder(w.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&result); err != nil {
			assert.NotNil(t, nil, err)
		}
		assert.Equal(t, (*rules), *(result.AdvancedRules))
	})

	t.Run("DeleteLinkedinCappingRule", func(t *testing.T) {
		w := sendDeleteLinkedinCappingRuleReq(r, project.ID, ruleID1, agent)
		assert.Equal(t, http.StatusOK, w.Code)
	})
	t.Run("GetDeletedRule", func(t *testing.T) {
		w := sendGetLinkedinCappingRuleReq(r, project.ID, ruleID1, agent)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func createLinkedinCompanyEngagementDocsForFrequencyCappingHandler(projectID int64, customerAccountID string, intTimestamp int64) int {
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
	return http.StatusCreated
}

func sendCreateLinkedinCappingRuleReq(r *gin.Engine, project_id int64, agent *model.Agent, objectType string,
	name string, description string, status string, objectIDs *postgres.Jsonb, isAdvRuleEN bool, advRuleType string, advRules *postgres.Jsonb) *httptest.ResponseRecorder {
	payload := map[string]interface{}{
		"project_id":               project_id,
		"object_type":              objectType,
		"display_name":             name,
		"description":              description,
		"status":                   status,
		"granularity":              "monthly",
		"impression_threshold":     1000,
		"click_threshold":          100,
		"object_ids":               objectIDs,
		"is_advanced_rule_enabled": isAdvRuleEN,
		"advanced_rule_type":       advRuleType,
		"advanced_rules":           advRules,
	}
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error Creating cookieData")
	}

	url := "/projects/" + strconv.FormatUint(uint64(project_id), 10) + "/v1/linkedin_capping/rules"
	rb := C.NewRequestBuilderWithPrefix(http.MethodPost, url).
		WithPostParams(payload).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})

	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error sending create linkedin capping rules request.")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}
func sendGetLinkedinCappingRulesReq(r *gin.Engine, projectID int64, agent *model.Agent) *httptest.ResponseRecorder {
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error Creating cookieData")
	}

	url := "/projects/" + strconv.FormatUint(uint64(projectID), 10) + "/v1/linkedin_capping/rules"
	rb := C.NewRequestBuilderWithPrefix(http.MethodGet, url).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})
	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error sending get linkedin capping rules request.")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}
func sendGetLinkedinCappingRuleReq(r *gin.Engine, projectID int64, ruleID string, agent *model.Agent) *httptest.ResponseRecorder {
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error Creating cookieData")
	}

	url := "/projects/" + strconv.FormatUint(uint64(projectID), 10) + "/v1/linkedin_capping/rules/" + ruleID
	rb := C.NewRequestBuilderWithPrefix(http.MethodGet, url).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})
	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error sending get linkedin capping rule request.")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}
func sendUpdateLinkedinCappingRuleReq(r *gin.Engine, project_id int64, ruleID string, agent *model.Agent, objectType string,
	name string, description string, status string, objectIDs *postgres.Jsonb, isAdvRuleEN bool, advRuleType string, advRules *postgres.Jsonb) *httptest.ResponseRecorder {
	payload := map[string]interface{}{
		"project_id":               project_id,
		"id":                       ruleID,
		"object_type":              objectType,
		"display_name":             name,
		"description":              description,
		"status":                   status,
		"granularity":              "monthly",
		"impression_threshold":     1000,
		"click_threshold":          100,
		"object_ids":               objectIDs,
		"is_advanced_rule_enabled": isAdvRuleEN,
		"advanced_rule_type":       advRuleType,
		"advanced_rules":           advRules,
	}
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error Creating cookieData")
	}

	url := "/projects/" + strconv.FormatUint(uint64(project_id), 10) + "/v1/linkedin_capping/rules/" + ruleID
	rb := C.NewRequestBuilderWithPrefix(http.MethodPut, url).
		WithPostParams(payload).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})

	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error sending update linkedin capping rules request.")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func sendDeleteLinkedinCappingRuleReq(r *gin.Engine, projectID int64, ruleID string, agent *model.Agent) *httptest.ResponseRecorder {
	cookieData, err := helpers.GetAuthData(agent.Email, agent.UUID, agent.Salt, 100*time.Second)
	if err != nil {
		log.WithError(err).Error("Error Creating cookieData")
	}

	url := "/projects/" + strconv.FormatUint(uint64(projectID), 10) + "/v1/linkedin_capping/rules/" + ruleID
	rb := C.NewRequestBuilderWithPrefix(http.MethodDelete, url).
		WithCookie(&http.Cookie{
			Name:   C.GetFactorsCookieName(),
			Value:  cookieData,
			MaxAge: 1000,
		})
	req, err := rb.Build()
	if err != nil {
		log.WithError(err).Error("Error sending get linkedin capping rule request.")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}
