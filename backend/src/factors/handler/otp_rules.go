package handler

import (
	"encoding/json"
	v1 "factors/handler/v1"
	mid "factors/middleware"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
	"net/http"
)

//OTPRuleRequestPayload is struct for post request to create otp rule
type OTPRuleRequestPayload struct {
	RuleType          string          `json:"rule_type"`
	CRMType           string          `json:"crm_type"`
	TouchPointTimeRef string          `json:"touch_point_time_ref"`
	Filters           *postgres.Jsonb `json:"filters"`
	PropertiesMap     *postgres.Jsonb `json:"properties_map"`
}

// OTPRuleUpdatePayload is struct update
type OTPRuleUpdatePayload struct {
	RuleType          string          `json:"rule_type"`
	CRMType           string          `json:"crm_type"`
	TouchPointTimeRef string          `json:"touch_point_time_ref"`
	Filters           *postgres.Jsonb `json:"filters"`
	PropertiesMap     *postgres.Jsonb `json:"properties_map"`
}

// GetOTPRuleHandler Gets all OTP rules for a project_id
func GetOTPRuleHandler(c *gin.Context) (interface{}, int, string, string, bool) {
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	logCtx := log.WithField("project_id", projectID)

	if projectID == 0 {
		errMsg := "Get otp_rules failed. Invalid project ID."
		logCtx.Error(errMsg)
		return nil, http.StatusUnauthorized, v1.INVALID_PROJECT, errMsg, true
	}
	otpRules, errCode := store.GetStore().GetALLOTPRuleWithProjectId(projectID)
	if errCode != http.StatusFound {
		errMsg := "Get ALL OTP Rule With ProjectId failed"
		logCtx.Error(errMsg)
		return nil, http.StatusBadRequest, v1.PROCESSING_FAILED, errMsg, true
	}

	return gin.H{"result": otpRules}, http.StatusOK, "", "", false
}

// CreateOTPRuleHandler Creates OTP rule
func CreateOTPRuleHandler(c *gin.Context) (interface{}, int, string, string, bool) {
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	logCtx := log.WithField("project_id", projectID)

	if projectID == 0 {
		errMsg := "Create otp_rules failed. Invalid project ID."
		logCtx.Error(errMsg)
		return nil, http.StatusUnauthorized, v1.INVALID_PROJECT, errMsg, true
	}

	agentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)

	var requestPayload OTPRuleRequestPayload

	r := c.Request
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&requestPayload); err != nil {
		errMsg := "Get otp_rules failed. Invalid JSON."
		logCtx.WithError(err).Error(errMsg)
		return nil, http.StatusBadRequest, v1.INVALID_INPUT, errMsg, true

	}

	ruleReq := &model.OTPRule{
		ProjectID:         projectID,
		RuleType:          requestPayload.RuleType,
		CRMType:           requestPayload.CRMType,
		TouchPointTimeRef: requestPayload.TouchPointTimeRef,
		PropertiesMap:     postgres.Jsonb{RawMessage: json.RawMessage(`{}`)},
		Filters:           postgres.Jsonb{RawMessage: json.RawMessage(`{}`)},
		CreatedBy:         agentUUID,
	}

	if requestPayload.Filters != nil && !U.IsEmptyPostgresJsonb(requestPayload.Filters) {
		ruleReq.Filters = *requestPayload.Filters
	} else {
		errMsg := "invalid filters for rule, exiting"
		logCtx.Error(errMsg)
		return nil, http.StatusBadRequest, v1.INVALID_INPUT, errMsg, true
	}

	if requestPayload.PropertiesMap != nil && !U.IsEmptyPostgresJsonb(requestPayload.PropertiesMap) {
		ruleReq.PropertiesMap = *requestPayload.PropertiesMap
	} else {
		errMsg := "invalid properties map for rule, exiting"
		logCtx.Error(errMsg)
		return nil, http.StatusBadRequest, v1.INVALID_INPUT, errMsg, true
	}

	if requestPayload.TouchPointTimeRef == "" {
		errMsg := "invalidTouchPointTimeRef for rule, exiting"
		logCtx.Error(errMsg)
		return nil, http.StatusBadRequest, v1.INVALID_INPUT, errMsg, true
	}

	if requestPayload.RuleType == "" {
		errMsg := "empty RuleType exiting"
		logCtx.Error(errMsg)
		return nil, http.StatusBadRequest, v1.INVALID_INPUT, errMsg, true
	}

	if requestPayload.CRMType == "" {
		errMsg := "empty CRMType exiting"
		logCtx.Error(errMsg)
		return nil, http.StatusBadRequest, v1.INVALID_INPUT, errMsg, true
	}

	rule, errCode, errMsg1 := store.GetStore().CreateOTPRule(projectID, ruleReq)
	if errCode != http.StatusCreated {
		logCtx.Error(errMsg1)
		return nil, http.StatusBadRequest, v1.PROCESSING_FAILED, errMsg1, true
	}

	return gin.H{"result": rule}, http.StatusOK, "", "", false
}

// UpdateOTPRuleHandler updates a given rule
func UpdateOTPRuleHandler(c *gin.Context) (interface{}, int, string, string, bool) {
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	logCtx := log.WithField("project_id", projectID)

	if projectID == 0 {
		errMsg := "Update otp_rules failed. Invalid project ID."
		logCtx.Error(errMsg)
		return nil, http.StatusUnauthorized, v1.INVALID_PROJECT, errMsg, true
	}
	agentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)

	var requestPayload OTPRuleUpdatePayload

	r := c.Request
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&requestPayload); err != nil {
		errMsg := "Get otp_rules failed. Invalid JSON."
		logCtx.WithError(err).Error(errMsg)
		return nil, http.StatusBadRequest, v1.INVALID_INPUT, errMsg, true
	}

	ruleId := c.Params.ByName("rule_id")
	if ruleId == "" {
		errMsg := "Invalid rule id."
		logCtx.Error(errMsg)
		return nil, http.StatusBadRequest, v1.INVALID_INPUT, errMsg, true
	}

	rule := &model.OTPRule{
		RuleType:          requestPayload.RuleType,
		CRMType:           requestPayload.CRMType,
		TouchPointTimeRef: requestPayload.TouchPointTimeRef,
		PropertiesMap:     postgres.Jsonb{RawMessage: json.RawMessage(`{}`)},
		Filters:           postgres.Jsonb{RawMessage: json.RawMessage(`{}`)},
		CreatedBy:         agentUUID,
	}

	if requestPayload.Filters != nil && !U.IsEmptyPostgresJsonb(requestPayload.Filters) {
		rule.Filters = *requestPayload.Filters
	} else {
		errMsg := "invalid filters for rule, exiting"
		logCtx.Error(errMsg)
		return nil, http.StatusBadRequest, v1.INVALID_INPUT, errMsg, true
	}

	if requestPayload.PropertiesMap != nil && !U.IsEmptyPostgresJsonb(requestPayload.PropertiesMap) {
		rule.PropertiesMap = *requestPayload.PropertiesMap
	} else {
		errMsg := "invalid properties map for rule, exiting"
		logCtx.Error(errMsg)
		return nil, http.StatusBadRequest, v1.INVALID_INPUT, errMsg, true
	}

	if requestPayload.TouchPointTimeRef == "" {
		errMsg := ""
		logCtx.Error(errMsg)
		return nil, http.StatusBadRequest, v1.INVALID_INPUT, errMsg, true
	}

	if requestPayload.RuleType == "" {
		errMsg := "empty RuleType exiting"
		logCtx.Error(errMsg)
		return nil, http.StatusBadRequest, v1.INVALID_INPUT, errMsg, true
	}

	if requestPayload.CRMType == "" {
		errMsg := "empty CRMType exiting"
		logCtx.Error(errMsg)
		return nil, http.StatusBadRequest, v1.INVALID_INPUT, errMsg, true
	}

	updatedRule, errCode := store.GetStore().UpdateOTPRule(projectID, ruleId, rule)
	if errCode != http.StatusAccepted {
		errMsg := "Failed to update Saved OTPRule."
		logCtx.Error(errMsg)
		return nil, http.StatusBadRequest, v1.PROCESSING_FAILED, errMsg, true
	}

	return gin.H{"result": updatedRule}, http.StatusOK, "", "", false
}

// DeleteOTPRuleHandler soft deletes a rule
func DeleteOTPRuleHandler(c *gin.Context) (interface{}, int, string, string, bool) {
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	logCtx := log.WithField("project_id", projectID)

	if projectID == 0 {
		errMsg := "Delete otp_rules failed. Invalid ProjectId."
		logCtx.Error(errMsg)
		return nil, http.StatusUnauthorized, v1.INVALID_PROJECT, errMsg, true
	}

	ruleID := c.Params.ByName("rule_id")
	if ruleID == "" {
		errMsg := "Invalid rule id."
		logCtx.Error(errMsg)
		return nil, http.StatusBadRequest, v1.INVALID_INPUT, errMsg, true
	}

	errCode, errMsg1 := store.GetStore().DeleteOTPRule(projectID, ruleID)
	if errCode != http.StatusAccepted {
		logCtx.Error(errMsg1)
		return nil, http.StatusBadRequest, v1.PROCESSING_FAILED, errMsg1, true
	}

	return gin.H{"result": "Successfully deleted."}, http.StatusOK, "", "", false
}

// SearchOTPRuleHandler Search a rule
func SearchOTPRuleHandler(c *gin.Context) (interface{}, int, string, string, bool) {

	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	logCtx := log.WithField("project_id", projectID)
	if projectID == 0 {
		errMsg := "Delete otp_rules failed. Invalid ProjectId."
		logCtx.Error(errMsg)
		return nil, http.StatusUnauthorized, v1.INVALID_PROJECT, errMsg, true
	}
	ruleId, ok := c.GetQuery("rule_id")
	if !ok || ruleId == "" {
		errMsg := "Invalid rule id."
		logCtx.Error(errMsg)
		return nil, http.StatusBadRequest, v1.INVALID_INPUT, errMsg, true
	}

	otpRules, errCode := store.GetStore().GetOTPRuleWithRuleId(projectID, ruleId)
	if errCode != http.StatusFound {
		errMsg := "Search OTPRule failed. No rule found."
		logCtx.Error(errMsg)
		return nil, http.StatusUnauthorized, v1.PROCESSING_FAILED, errMsg, true
	}

	return gin.H{"result": otpRules}, errCode, "", "", false
}
