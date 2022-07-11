package handler

import (
	"encoding/json"
	V1 "factors/handler/v1"
	mid "factors/middleware"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

const smartPropertyTypeKey = "object_type"
const ruleIDKey = "rule_id"

// GetSmartPropertyRulesConfigHandler godoc
// @Summary Get the configs for creating a smart properties rule.
// @Tag v1ApiSmartProperty
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param object_type path string true "Object UnitType (campaign/ad_group)"
// @Success 200 {object} model.SmartPropertyRulesConfig
// @Router /{project_id}/v1/smart_properties/config/{object_type} [get]
func GetSmartPropertyRulesConfigHandler(c *gin.Context) (interface{}, int, string, string, bool) {
	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		return nil, http.StatusUnauthorized, V1.INVALID_PROJECT, V1.ErrorMessages[V1.INVALID_PROJECT], true
	}

	objectType := c.Params.ByName(smartPropertyTypeKey)
	result, httpStatus := store.GetStore().GetSmartPropertyRulesConfig(projectId, objectType)
	if httpStatus != http.StatusOK {
		return nil, http.StatusUnauthorized, V1.INVALID_INPUT, "Invalid object type", true
	}

	return result, http.StatusOK, "", "", false
}

// CreateSmartPropertyRulesHandler godoc
// @Summary To create a new smart properties rule.
// @Tags V1ApiSmartProperty
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param Rule body model.SmartPropertyRules true "Create SmartPropertyRules"
// @Success 200 {string, CreatedSmartPropertyRule}
// @Router /{project_id}/v1/smart_properties  [post]
func CreateSmartPropertyRulesHandler(c *gin.Context) (interface{}, int, string, string, bool) {
	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		return nil, http.StatusUnauthorized, V1.INVALID_PROJECT, V1.ErrorMessages[V1.INVALID_PROJECT], true
	}

	r := c.Request

	var smartPropertyRulesReqPayload model.SmartPropertyRules
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&smartPropertyRulesReqPayload); err != nil {
		log.WithError(err).Error("Failed to decode Json request on create smart properties handler.")
		return nil, http.StatusInternalServerError, V1.PROCESSING_FAILED, "Failed to decode Json request on create smart properties handler.", true
	}

	smartPropertyRule, errMsg, errCode := store.GetStore().CreateSmartPropertyRules(projectId, &smartPropertyRulesReqPayload)
	if errCode != http.StatusCreated {
		log.WithField("document", smartPropertyRulesReqPayload).Error("Failed to insert smart properties on create smart properties handler.")
		return nil, errCode, V1.PROCESSING_FAILED, errMsg, true
	}

	return smartPropertyRule, http.StatusOK, "", "", false
}

// UpdateSmartPropertyRulesHandler godoc
// @Summary To update an existing smart properties rule.
// @Tags V1ApiSmartProperty
// @Accept  json
// @Produce json
// @Param rule_id query integer true "Rule ID"
// @Param project_id path integer true "Project ID"
// @Param Rule body model.SmartPropertyRules true "Update SmartPropertyRules"
// @Success 200 {string, updatedSmartPropertyRule}
// @Router /{project_id}/v1/smart_properties/rules/{rule_id}  [put]
func UpdateSmartPropertyRulesHandler(c *gin.Context) (interface{}, int, string, string, bool) {
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		log.Error("UpdateSmartProperty Failed. ProjectId parse failed.")
		return nil, http.StatusUnauthorized, V1.INVALID_PROJECT, V1.ErrorMessages[V1.INVALID_PROJECT], true
	}
	ruleID := c.Params.ByName(ruleIDKey)
	if ruleID == "" {
		log.Error("UpdateSmartProperty Failed. RuleID parse failed.")
		return nil, http.StatusBadRequest, V1.INVALID_INPUT, "RuleID parse failed", true
	}

	r := c.Request
	var smartProperty model.SmartPropertyRules
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&smartProperty); err != nil {
		log.WithError(err).Error("Failed to decode Json request on update smart properties handler.")
		return nil, http.StatusInternalServerError, V1.PROCESSING_FAILED, "Failed to decode Json request on create smart properties handler.", true
	}

	_, errMsg, errCode := store.GetStore().UpdateSmartPropertyRules(projectID, ruleID, smartProperty)
	if errCode != http.StatusAccepted {
		return nil, errCode, V1.PROCESSING_FAILED, errMsg, true
	}
	smartPropertyRule, errCode := store.GetStore().GetSmartPropertyRule(projectID, ruleID)
	if errCode != http.StatusFound {
		return nil, errCode, V1.PROCESSING_FAILED, "Failed to get updated rule", true
	}
	return smartPropertyRule, http.StatusOK, "", "", false
}

// GetSmartPropertyRulesHandler godoc
// @Summary Get the list of existing smart properties rules.
// @Tags v1ApiSmartProperty
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Success 200 {array} model.SmartPropertyRules
// @Router /{project_id}/v1/smart_properties [get]
func GetSmartPropertyRulesHandler(c *gin.Context) (interface{}, int, string, string, bool) {
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		log.Error("GetSmartProperty Failed. ProjectId parse failed.")
		return nil, http.StatusUnauthorized, V1.INVALID_PROJECT, V1.ErrorMessages[V1.INVALID_PROJECT], true
	}
	smartPropertyRules, errCode := store.GetStore().GetSmartPropertyRules(projectID)
	if errCode != http.StatusFound {
		return nil, errCode, V1.PROCESSING_FAILED, "Failed to get smart property rules", true
	}
	return smartPropertyRules, http.StatusOK, "", "", false
}

// GetSmartPropertyRuleByRuleIDHandler godoc
// @Summary Get one of existing smart properties rules using rule id.
// @Tags v1ApiSmartProperty
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param rule_id path integer true "Rule ID"
// @Success 200 {object} model.SmartPropertyRules
// @Router /{project_id}/v1/smart_properties/rules/{rule_id} [get]
func GetSmartPropertyRuleByRuleIDHandler(c *gin.Context) (interface{}, int, string, string, bool) {
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		log.Error("GetSmartProperty Failed. ProjectId parse failed.")
		return nil, http.StatusUnauthorized, V1.INVALID_PROJECT, V1.ErrorMessages[V1.INVALID_PROJECT], true
	}
	ruleID := c.Params.ByName(ruleIDKey)
	if ruleID == "" {
		log.Error("GetSmartProperty Failed. RuleID parse failed.")
		return nil, http.StatusBadRequest, V1.INVALID_INPUT, "RuleID parse failed", true
	}

	smartPropertyRule, errCode := store.GetStore().GetSmartPropertyRule(projectID, ruleID)
	if errCode != http.StatusFound {
		return nil, errCode, V1.PROCESSING_FAILED, "Failed to get smart property rule", true
	}
	return smartPropertyRule, http.StatusOK, "", "", false
}

// DeleteSmartPropertyRulesHandler godoc
// @Summary To delete an existing smart properties rule.
// @Tags V1ApiSmartProperty
// @Accept  json
// @Produce json
// @Param rue_id query integer true "Rule ID"
// @Param project_id path integer true "Project ID"
// @Success 200 {string}
// @Router /{project_id}/v1/smart_properties/rules/{rule_id} [delete]
func DeleteSmartPropertyRulesHandler(c *gin.Context) (interface{}, int, string, string, bool) {
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		log.Error("DeleteSmartProperty Failed. ProjectId parse failed.")
		return nil, http.StatusUnauthorized, V1.INVALID_PROJECT, V1.ErrorMessages[V1.INVALID_PROJECT], true
	}

	ruleID := c.Params.ByName(ruleIDKey)
	if ruleID == "" {
		log.Error("DeleteSmartProperty Failed. RuleID parse failed.")
		return nil, http.StatusBadRequest, V1.INVALID_INPUT, "RuleID parse failed", true
	}

	errCode := store.GetStore().DeleteSmartPropertyRules(projectID, ruleID)
	if errCode != http.StatusAccepted {
		return nil, errCode, V1.PROCESSING_FAILED, "Failed to delete smart property rule", true
	}
	return nil, http.StatusOK, "", "", false
}
