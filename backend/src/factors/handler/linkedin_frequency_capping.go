package handler

import (
	"encoding/json"
	V1 "factors/handler/v1"
	mid "factors/middleware"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

const (
	timestampKey   = "timestamp"
	startTimestamp = "startTimestamp"
	endTimestamp   = "endTimestamp"
)

// GetLinkedinCappingConfigHandler godoc
// @Summary Get the configs for creating a linkedin freq capping rule.
// @Tag v1ApiLinkedinFreqCapping
// @Accept  None
// @Produce json
// @Param project_id path integer true "Project ID"
// @Success 200 [{object}] []model.LinkedinCappingConfig
// @Router /{project_id}/v1/linkedin_capping/rules/config [get]
func GetLinkedinCappingConfigHandler(c *gin.Context) (interface{}, int, string, string, bool) {
	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		return nil, http.StatusUnauthorized, INVALID_PROJECT, V1.ErrorMessages[INVALID_PROJECT], true
	}

	result, httpStatus := store.GetStore().GetLinkedinCappingConfig(projectId)
	if httpStatus != http.StatusOK {
		return nil, http.StatusUnauthorized, INVALID_INPUT, "Invalid object type", true
	}

	return result, http.StatusOK, "", "", false
}

// CreateLinkedinCappingRulesHandler godoc
// @Summary To create a new linkedin capping rule.
// @Tags V1ApiLinkedinCapping
// @Accept  json
// @Produce int
// @Param project_id path integer true "Project ID"
// @Param Rule body model.LinkedinCappingRules true "Create LinkedinCappingRules"
// @Success 200 {string, CreatedLinkedinCappingRule}
// @Router /{project_id}/v1/linkedin_capping/rules  [post]
func CreateLinkedinCappingRulesHandler(c *gin.Context) (interface{}, int, string, string, bool) {
	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		return nil, http.StatusUnauthorized, INVALID_PROJECT, V1.ErrorMessages[INVALID_PROJECT], true
	}

	r := c.Request

	var linkedinCappingRulesReqPayload model.LinkedinCappingRule
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&linkedinCappingRulesReqPayload); err != nil {
		log.WithError(err).Error("Failed to decode Json request on create linkedin capping handler.")
		return nil, http.StatusInternalServerError, V1.PROCESSING_FAILED, "Failed to decode Json request on create linkedin capping handler.", true
	}

	errCode := store.GetStore().CreateLinkedinCappingRule(projectId, &linkedinCappingRulesReqPayload)
	if errCode != http.StatusCreated {
		log.WithField("document", linkedinCappingRulesReqPayload).Error("Failed to insert linkedin capping on create linkedin capping handler.")
		return nil, errCode, V1.PROCESSING_FAILED, "Failed to insert linkedin capping on create linkedin capping handler.", true
	}

	return nil, http.StatusOK, "", "", false
}

// UpdateLinkedinCappingRulesHandler godoc
// @Summary To update an existing linkedin capping rule.
// @Tags V1ApiLinkedinCapping
// @Accept  json
// @Produce Int
// @Param rule_id query integer true "Rule ID"
// @Param project_id path integer true "Project ID"
// @Param Rule body model.LinkedinCappingRules true "Update LinkedinCappingRules"
// @Success 200 {string, LinekdinCappingRule}
// @Router /{project_id}/v1/linkedin_capping/rules/{rule_id}  [put]
func UpdateLinkedinCappingRulesHandler(c *gin.Context) (interface{}, int, string, string, bool) {
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		log.Error("UpdateLinkedinCapping Failed. ProjectId parse failed.")
		return nil, http.StatusUnauthorized, INVALID_PROJECT, V1.ErrorMessages[INVALID_PROJECT], true
	}
	ruleID := c.Params.ByName(ruleIDKey)
	if ruleID == "" {
		log.Error("UpdateLinkedinCapping Failed. RuleID parse failed.")
		return nil, http.StatusBadRequest, INVALID_INPUT, "RuleID parse failed", true
	}

	r := c.Request
	var linkedinCapping model.LinkedinCappingRule
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&linkedinCapping); err != nil {
		log.WithError(err).Error("Failed to decode Json request on update linkedin capping handler.")
		return nil, http.StatusInternalServerError, V1.PROCESSING_FAILED, "Failed to decode Json request on create linkedin capping handler.", true
	}

	errCode := store.GetStore().UpdateLinkedinCappingRule(projectID, ruleID)
	if errCode != http.StatusAccepted {
		errMsg := "Failed to update linkedin capping rule"
		return nil, errCode, V1.PROCESSING_FAILED, errMsg, true
	}
	linkedinCappingRule, errCode := store.GetStore().GetLinkedinCappingRule(projectID, ruleID)
	if errCode != http.StatusFound {
		return nil, errCode, V1.PROCESSING_FAILED, "Failed to get updated rule", true
	}
	return linkedinCappingRule, http.StatusOK, "", "", false
}

// GetLinkedinCappingRulesHandler godoc
// @Summary Get the list of existing linkedin capping rules.
// @Tags v1ApiLinkedinCapping
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Success 200 {array} model.LinkedinCappingRules
// @Router /{project_id}/v1/linkedin_capping/rules [get]
func GetLinkedinCappingRulesHandler(c *gin.Context) (interface{}, int, string, string, bool) {
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		log.Error("GetLinkedinCapping Failed. ProjectId parse failed.")
		return nil, http.StatusUnauthorized, INVALID_PROJECT, V1.ErrorMessages[INVALID_PROJECT], true
	}
	linkedinCappingRules, errCode := store.GetStore().GetAllLinkedinCappingRules(projectID)
	if errCode != http.StatusFound {
		return nil, errCode, V1.PROCESSING_FAILED, "Failed to get smart property rules", true
	}
	return linkedinCappingRules, http.StatusOK, "", "", false
}

// GetLinkedinCappingRuleByRuleIDHandler godoc
// @Summary Get one of existing linkedin capping rules using rule id.
// @Tags v1ApiLinkedinCapping
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param rule_id path integer true "Rule ID"
// @Success 200 {object} model.LinkedinCappingRules
// @Router /{project_id}/v1/linkedin_capping/rules/{rule_id} [get]
func GetLinkedinCappingRuleByRuleIDHandler(c *gin.Context) (interface{}, int, string, string, bool) {
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		log.Error("GetLinkedinCapping Failed. ProjectId parse failed.")
		return nil, http.StatusUnauthorized, INVALID_PROJECT, V1.ErrorMessages[INVALID_PROJECT], true
	}
	ruleID := c.Params.ByName(ruleIDKey)
	if ruleID == "" {
		log.Error("GetLinkedinCapping Failed. RuleID parse failed.")
		return nil, http.StatusBadRequest, INVALID_INPUT, "RuleID parse failed", true
	}

	linkedinCappingRule, errCode := store.GetStore().GetLinkedinCappingRule(projectID, ruleID)
	if errCode != http.StatusFound {
		return nil, errCode, V1.PROCESSING_FAILED, "Failed to get smart property rule", true
	}
	return linkedinCappingRule, http.StatusOK, "", "", false
}

// DeleteLinkedinCappingRulesHandler godoc
// @Summary To delete an existing linkedin capping rule.
// @Tags V1ApiLinkedinCapping
// @Accept  json
// @Produce json
// @Param rue_id query integer true "Rule ID"
// @Param project_id path integer true "Project ID"
// @Success 200 {string}
// @Router /{project_id}/v1/linkedin_capping/rules/{rule_id} [delete]
func DeleteLinkedinCappingRulesHandler(c *gin.Context) (interface{}, int, string, string, bool) {
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		log.Error("DeleteLinkedinCapping Failed. ProjectId parse failed.")
		return nil, http.StatusUnauthorized, INVALID_PROJECT, V1.ErrorMessages[INVALID_PROJECT], true
	}

	ruleID := c.Params.ByName(ruleIDKey)
	if ruleID == "" {
		log.Error("DeleteLinkedinCapping Failed. RuleID parse failed.")
		return nil, http.StatusBadRequest, INVALID_INPUT, "RuleID parse failed", true
	}

	errCode := store.GetStore().DeleteLinkedinCappingRule(projectID, ruleID)
	if errCode != http.StatusAccepted {
		return nil, errCode, V1.PROCESSING_FAILED, "Failed to delete smart property rule", true
	}
	return nil, http.StatusOK, "", "", false
}

// GetLinkedinCappingExclusionsHandler godoc
// @Summary Get the list of existing linkedin capping rules.
// @Tags v1ApiLinkedinCapping
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Success 200 {array} model.LinkedinCappingExclusions
// @Router /{project_id}/v1/linkedin_capping/exclusions/?startTimestamp=&endTimestamp [get]
func GetLinkedinCappingExclusionsHandler(c *gin.Context) (interface{}, int, string, string, bool) {
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		log.Error("GetLinkedinCapping exclusions for month Failed. ProjectId parse failed.")
		return nil, http.StatusUnauthorized, INVALID_PROJECT, V1.ErrorMessages[INVALID_PROJECT], true
	}

	timestamp := c.Params.ByName(timestampKey)
	if timestamp == "" {
		log.Error("GetLinkedinCapping exclusions for month Failed. timestamp parse failed.")
		return nil, http.StatusBadRequest, INVALID_INPUT, "timestamp parse failed", true
	}
	startTimestamp, err := strconv.ParseInt(c.Query("startTimestamp"), 10, 64)
	if err != nil {
		log.Error("GetLinkedinCapping exclusions for month Failed. timestamp conversion failed.")
		return nil, http.StatusBadRequest, INVALID_INPUT, "timestamp conversion failed", true
	}
	endTimestamp, err := strconv.ParseInt(c.Query("endTimestamp"), 10, 64)
	if err != nil {
		log.Error("GetLinkedinCapping exclusions for month Failed. timestamp conversion failed.")
		return nil, http.StatusBadRequest, INVALID_INPUT, "timestamp conversion failed", true
	}
	linkedinCappingExclusions, errCode := store.GetStore().GetAllLinkedinCappingExclusionsForTimerange(projectID, startTimestamp, endTimestamp)
	if errCode != http.StatusFound {
		return nil, errCode, V1.PROCESSING_FAILED, "Failed to get linkedin capping rules", true
	}
	return linkedinCappingExclusions, http.StatusOK, "", "", false
}

// GetLinkedinCappingxclusionsByRuleIDHandler godoc
// @Summary Get one of existing linkedin capping rules using rule id.
// @Tags v1ApiLinkedinCapping
// @Accept  none
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param rule_id path integer true "Rule ID"
// @Success 200 {object} model.LinkedinCappingRules
// @Router /{project_id}/v1/linkedin_capping/exclusions/{rule_id} [get]
func GetLinkedinCappingExclusionsByRuleIDHandler(c *gin.Context) (interface{}, int, string, string, bool) {
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		log.Error("GetLinkedinCapping Failed. ProjectId parse failed.")
		return nil, http.StatusUnauthorized, INVALID_PROJECT, V1.ErrorMessages[INVALID_PROJECT], true
	}
	ruleID := c.Params.ByName(ruleIDKey)
	if ruleID == "" {
		log.Error("GetLinkedinCapping Failed. RuleID parse failed.")
		return nil, http.StatusBadRequest, INVALID_INPUT, "RuleID parse failed", true
	}

	linkedinCappingExclusions, errCode := store.GetStore().GetLinkedinCappingExclusionsForRule(projectID, ruleID)
	if errCode != http.StatusFound {
		return nil, errCode, V1.PROCESSING_FAILED, "Failed to get smart property rule", true
	}
	return linkedinCappingExclusions, http.StatusOK, "", "", false
}

func GetLinkedinCappingExclusionsDashboardHandler(c *gin.Context) (interface{}, int, string, string, bool) {
	return model.SampleExclusionDashboard, http.StatusOK, "", "", false
}
