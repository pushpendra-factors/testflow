package handler

import (
	"encoding/json"
	mid "factors/middleware"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

const smartPropertiesTypeKey = "object_type"
const ruleIDKey = "rule_id"

// GetSmartPropertiesRulesConfigHandler godoc
// @Summary Get the configs for creating a smart properties rule.
// @Tag v1ApiSmartProperties
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param object_type path string true "Object Type (campaign/ad_group)"
// @Success 200 {object} model.SmartPropertiesRulesConfig
// @Router /{project_id}/v1/smart_properties/config/{object_type} [get]
func GetSmartPropertiesRulesConfigHandler(c *gin.Context) {
	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatusJSON(http.StatusForbidden,
			gin.H{"error": "Get dashboard units failed. Invalid project."})
		return
	}

	objectType := c.Params.ByName(smartPropertiesTypeKey)
	result, httpStatus := store.GetStore().GetSmartPropertiesRulesConfig(projectId, objectType)
	if httpStatus != http.StatusOK {
		c.AbortWithStatusJSON(http.StatusForbidden,
			gin.H{"error": "invalid object type"})
		return
	}

	c.JSON(httpStatus, gin.H{"result": result})
}

// CreateSmartPropertiesRulesHandler godoc
// @Summary To create a new smart properties rule.
// @Tags V1ApiSmartProperties
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param Rule body model.SmartPropertiesRules true "Create SmartPropertiesRules"
// @Success 201 {string}
// @Router /{project_id}/v1/smart_properties  [post]
func CreateSmartPropertiesRulesHandler(c *gin.Context) {
	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatusJSON(http.StatusForbidden,
			gin.H{"error": "Create smart properties failed. Invalid project."})
		return
	}
	loggedInAgentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)

	projectAgentMapping, errCode := store.GetStore().GetProjectAgentMapping(projectId, loggedInAgentUUID)
	if errCode != http.StatusFound {
		c.AbortWithStatusJSON(http.StatusForbidden,
			gin.H{"error": "Unauthorized Access."})
		return
	}
	if projectAgentMapping.Role != model.ADMIN {
		c.AbortWithStatusJSON(http.StatusForbidden,
			gin.H{"error": "Only admins can create smart properties rules."})
		return
	}

	r := c.Request

	var smartProperties model.SmartPropertiesRules
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&smartProperties); err != nil {
		log.WithError(err).Error("Failed to decode Json request on create smart properties handler.")
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			gin.H{"error": "Invalid request json."})
		return
	}

	errMsg, errCode := store.GetStore().CreateSmartPropertiesRules(projectId, &smartProperties)
	if errCode != http.StatusCreated {
		log.WithField("document", smartProperties).Error("Failed to insert smart properties on create smart properties handler.")
		c.AbortWithStatusJSON(errCode,
			gin.H{"error": errMsg})
		return
	}

	c.JSON(errCode, gin.H{"message": "Successfully created smart properties."})
}

// UpdateSmartPropertiesRulesHandler godoc
// @Summary To update an existing smart properties rule.
// @Tags V1ApiSmartProperties
// @Accept  json
// @Produce json
// @Param rule_id query integer true "Rule ID"
// @Param project_id path integer true "Project ID"
// @Param Rule body model.SmartPropertiesRules true "Update SmartPropertiesRules"
// @Success 200 {string}
// @Router /{project_id}/v1/smart_properties/rules/{rule_id}  [put]
func UpdateSmartPropertiesRulesHandler(c *gin.Context) {
	projectID := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		log.Error("UpdateSmartProperties Failed. ProjectId parse failed.")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	ruleID := c.Params.ByName(ruleIDKey)
	if ruleID == "" {
		log.Error("UpdateSmartProperties Failed. RuleID parse failed.")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	loggedInAgentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)

	projectAgentMapping, errCode := store.GetStore().GetProjectAgentMapping(projectID, loggedInAgentUUID)
	if errCode != http.StatusFound {
		c.AbortWithStatusJSON(http.StatusForbidden,
			gin.H{"error": "Unauthorized Access."})
		return
	}
	if projectAgentMapping.Role != model.ADMIN {
		c.AbortWithStatusJSON(http.StatusForbidden,
			gin.H{"error": "Only admins can update smart properties rules."})
		return
	}

	r := c.Request
	var smartProperties model.SmartPropertiesRules
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&smartProperties); err != nil {
		log.WithError(err).Error("Failed to decode Json request on update smart properties handler.")
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			gin.H{"error": "Invalid request json."})
		return
	}

	errMsg, errCode := store.GetStore().UpdateSmartPropertiesRules(projectID, ruleID, smartProperties)
	if errCode != http.StatusAccepted {
		c.AbortWithStatusJSON(errCode, gin.H{"error": errMsg})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Successfully updated smart property."})
}

// GetSmartPropertiesRulesHandler godoc
// @Summary Get the list of existing smart properties rules.
// @Tags v1ApiSmartProperties
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Success 200 {array} model.SmartPropertiesRules
// @Router /{project_id}/v1/smart_properties [get]
func GetSmartPropertiesRulesHandler(c *gin.Context) {
	projectID := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		log.Error("GetSmartProperties Failed. ProjectId parse failed.")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	smartProperties, errCode := store.GetStore().GetSmartPropertiesRules(projectID)
	if errCode != http.StatusFound {
		c.AbortWithStatus(errCode)
		return
	}
	c.JSON(http.StatusOK, smartProperties)
}

// GetSmartPropertiesRuleByRuleIDHandler godoc
// @Summary Get one of existing smart properties rules using rule id.
// @Tags v1ApiSmartProperties
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param rule_id path integer true "Rule ID"
// @Success 200 {object} model.SmartPropertiesRules
// @Router /{project_id}/v1/smart_properties/rules/{rule_id} [get]
func GetSmartPropertiesRuleByRuleIDHandler(c *gin.Context) {
	projectID := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		log.Error("GetSmartProperties Failed. ProjectId parse failed.")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	ruleID := c.Params.ByName(ruleIDKey)
	if ruleID == "" {
		log.Error("GetSmartProperties Failed. RuleID parse failed.")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	smartProperty, errCode := store.GetStore().GetSmartPropertiesRule(projectID, ruleID)
	if errCode != http.StatusFound {
		c.AbortWithStatus(errCode)
		return
	}
	c.JSON(http.StatusOK, smartProperty)
}

// DeleteSmartPropertiesRulesHandler godoc
// @Summary To delete an existing smart properties rule.
// @Tags V1ApiSmartProperties
// @Accept  json
// @Produce json
// @Param rue_id query integer true "Rule ID"
// @Param project_id path integer true "Project ID"
// @Success 200 {string}
// @Router /{project_id}/v1/smart_properties/rules/{rule_id} [delete]
func DeleteSmartPropertiesRulesHandler(c *gin.Context) {
	projectID := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		log.Error("DeleteSmartProperties Failed. ProjectId parse failed.")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	ruleID := c.Params.ByName(ruleIDKey)
	if ruleID == "" {
		log.Error("DeleteSmartProperties Failed. RuleID parse failed.")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	loggedInAgentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)

	projectAgentMapping, errCode := store.GetStore().GetProjectAgentMapping(projectID, loggedInAgentUUID)
	if errCode != http.StatusFound {
		c.AbortWithStatusJSON(http.StatusForbidden,
			gin.H{"error": "Unauthorized Access."})
		return
	}
	if projectAgentMapping.Role != model.ADMIN {
		c.AbortWithStatusJSON(http.StatusForbidden,
			gin.H{"error": "Only admins can delete smart properties rules."})
		return
	}

	errCode = store.GetStore().DeleteSmartPropertiesRules(projectID, ruleID)
	if errCode != http.StatusAccepted {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Failed to delete smart properties rule."})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Successfully delete smart property."})
}
