package tests

import (
	"encoding/json"
	C "factors/config"
	H "factors/handler"
	"factors/handler/helpers"
	"factors/model/model"
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

/*
project_id bigint NOT NULL,
type text NOT NULL,
name text NOT NULL,
display_name text NOT NULL,
description text NOT NULL,
status text NOT NULL,
object_ids json,
impression_threshold bigint NOT NULL,
click_threshold bigint NOT NULL,
is_advance_rule_enabled bool DEFAULT FALSE,
adv_rule_type text,
adv_rules json,
*/
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
		"advanced_rule":            advRules,
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
		"advanced_rule":            advRules,
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

func TestLinkedinCappingRuleHandler(t *testing.T) {
	r := gin.Default()
	H.InitAppRoutes(r)

	project, agent, _ := SetupProjectWithAgentDAO()
	assert.NotNil(t, project)

	name1 := U.RandomString(8)
	name2 := U.RandomString(8)
	ruleID1 := ""
	ruleID2 := ""

	t.Run("CreateLinkedinCappingRuleFailure:AccountLevelWithObjectIDS", func(t *testing.T) {
		name1 := U.RandomString(8)
		description := U.RandomString(20)
		rules := &postgres.Jsonb{json.RawMessage(`[]`)}
		objectIDs := &postgres.Jsonb{json.RawMessage(`["cg_123"]`)}
		w := sendCreateLinkedinCappingRuleReq(r, project.ID, agent, model.LINKEDIN_ACCOUNT, name1,
			description, model.LINKEDIN_STATUS_ACTIVE, objectIDs, false, model.LINKEDIN_ACCOUNT, rules)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("CreateLinkedinCappingRuleFailure:CampaignLevelWihoutObjectIDs", func(t *testing.T) {
		name1 := U.RandomString(8)
		description := U.RandomString(20)
		rules := &postgres.Jsonb{json.RawMessage(`[]`)}
		objectIDs := &postgres.Jsonb{json.RawMessage(`[]`)}
		w := sendCreateLinkedinCappingRuleReq(r, project.ID, agent, model.LINKEDIN_CAMPAIGN, name1,
			description, model.LINKEDIN_STATUS_ACTIVE, objectIDs, false, model.LINKEDIN_ACCOUNT, rules)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("CreateLinkedinCappingRuleFailure:CampaignGroupLevelWihoutObjectIDs", func(t *testing.T) {
		name1 := U.RandomString(8)
		description := U.RandomString(20)
		rules := &postgres.Jsonb{json.RawMessage(`[]`)}
		objectIDs := &postgres.Jsonb{json.RawMessage(`[]`)}
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
			description, model.LINKEDIN_STATUS_ACTIVE, nil, false, model.LINKEDIN_ACCOUNT, &postgres.Jsonb{rules})
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
		rules := &postgres.Jsonb{json.RawMessage(`[]`)}
		objectIDs := &postgres.Jsonb{json.RawMessage(`[]`)}
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
		assert.Equal(t, len(result), 1)
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
		rules := &postgres.Jsonb{json.RawMessage(`[]`)}
		objectIDs := &postgres.Jsonb{json.RawMessage(`[]`)}
		w := sendCreateLinkedinCappingRuleReq(r, project.ID, agent, model.LINKEDIN_ACCOUNT, name1+"$",
			description, model.LINKEDIN_STATUS_ACTIVE, objectIDs, false, model.LINKEDIN_ACCOUNT, rules)
		assert.Equal(t, http.StatusOK, w.Code)

		w = sendCreateLinkedinCappingRuleReq(r, project.ID, agent, model.LINKEDIN_ACCOUNT, name1+"@",
			description, model.LINKEDIN_STATUS_ACTIVE, objectIDs, false, model.LINKEDIN_ACCOUNT, rules)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("CreateLinkedinCappingRuleSuccess:CampaignLevel", func(t *testing.T) {
		description := U.RandomString(20)
		rules := &postgres.Jsonb{json.RawMessage(`[]`)}
		objectIDs := &postgres.Jsonb{json.RawMessage(`["c123"]`)}
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

	t.Run("UpdateLinkedinCappingRuleFailure:RepeatedName", func(t *testing.T) {
		description := U.RandomString(20)
		rules := &postgres.Jsonb{json.RawMessage(`[]`)}
		objectIDs := &postgres.Jsonb{json.RawMessage(`[]`)}
		w := sendUpdateLinkedinCappingRuleReq(r, project.ID, ruleID2, agent, model.LINKEDIN_ACCOUNT, name1,
			description, model.LINKEDIN_STATUS_ACTIVE, objectIDs, false, model.LINKEDIN_ACCOUNT, rules)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
	t.Run("UpdateLinkedinCappingRule:Success", func(t *testing.T) {
		description := U.RandomString(20)
		rules := &postgres.Jsonb{json.RawMessage(`[]`)}
		objectIDs := &postgres.Jsonb{json.RawMessage(`[]`)}
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
