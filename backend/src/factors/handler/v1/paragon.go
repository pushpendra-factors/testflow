package v1

import (
	"encoding/json"
	"factors/config"
	"factors/integration/paragon"
	mid "factors/middleware"
	"factors/model/store"
	U "factors/util"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

const ParagonIntegrationsCountLimit = 10

func GetParagonAuthenticationTokenForProject(c *gin.Context) (interface{}, int, string, string, bool) {
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)

	if projectID == 0 {
		return "", http.StatusForbidden, "", "Get request failed. Invalid project ID.", true
	}

	agentID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)

	logCtx := log.WithFields(log.Fields{"project_id": projectID})

	token, errCode, err := store.GetStore().GetParagonTokenFromProjectSetting(projectID)
	if errCode == http.StatusNotFound {
		count, errCode, err := store.GetStore().GetParagonEnabledProjectsCount(projectID)
		if err != nil || errCode != http.StatusOK {
			logCtx.WithError(err).Error("integration count can not be found")
			return "", errCode, PROCESSING_FAILED, err.Error(), true
		}

		if count > ParagonIntegrationsCountLimit {
			logCtx.WithError(err).Error("PARAGON SEATS EXHAUSTED")
			return "", http.StatusUnauthorized, PROCESSING_FAILED, "PARAGON SEATS EXHAUSTED", true
		}

		token, err = paragon.GenerateJWTTokenForProject(projectID)
		if err != nil || token == "" {
			logCtx.WithError(err).Error("failed to generate jwt token")
			return "", http.StatusUnauthorized, PROCESSING_FAILED, err.Error(), true
		}

		errCode, err = store.GetStore().AddParagonTokenAndEnablingAgentToProjectSetting(projectID, agentID, token)
		if err != nil || errCode != http.StatusAccepted {
			logCtx.WithError(err).Error("jwt token update failed")
			return "", http.StatusUnauthorized, PROCESSING_FAILED, err.Error(), true
		}

		return token, http.StatusOK, "", "", false
	}
	if errCode != http.StatusFound {
		logCtx.Error("failed to get token for project")
		return "", http.StatusInternalServerError, PROCESSING_FAILED, err.Error(), true
	}

	return token, http.StatusOK, "", "", false
}

func GetParagonUser(c *gin.Context) {
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)

	if projectID == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}

	token, errCode, err := store.GetStore().GetParagonTokenFromProjectSetting(projectID)
	if (errCode != http.StatusFound && errCode != http.StatusNotFound) || err != nil {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "failed to get paragon auth token"})
		return
	}
	if errCode == http.StatusNotFound || token == "" {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "no paragon auth token found"})
		return
	}

	paragonProjectID := config.GetParagonProjectID()
	response, err := paragon.GetParagonUserAPI(token, paragonProjectID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "fail to fetch response from paragon"})
		return
	}
	if response == nil {
		c.AbortWithStatusJSON(http.StatusExpectationFailed, gin.H{"error": "nil response from paragon"})
		return
	}

	c.JSON(http.StatusOK, response)
	return
}

func GetParagonIntegrationsMetadata(c *gin.Context) {
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)

	if projectID == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}

	token, errCode, err := store.GetStore().GetParagonTokenFromProjectSetting(projectID)
	if (errCode != http.StatusFound && errCode != http.StatusNotFound) || err != nil {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "failed to get paragon auth token"})
		return
	}
	if errCode == http.StatusNotFound || token == "" {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "no paragon auth token found"})
		return
	}

	paragonProjectID := config.GetParagonProjectID()
	response, err := paragon.GetParagonIntegrationMetadataAPI(token, paragonProjectID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "fail to fetch response from paragon"})
		return
	}
	if response == nil {
		c.AbortWithStatusJSON(http.StatusExpectationFailed, gin.H{"error": "nil response from paragon"})
		return
	}

	c.JSON(http.StatusOK, response)
	return
}

func GetParagonProjectIntegrations(c *gin.Context) {
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)

	if projectID == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}

	token, errCode, err := store.GetStore().GetParagonTokenFromProjectSetting(projectID)
	if (errCode != http.StatusFound && errCode != http.StatusNotFound) || err != nil {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "failed to get paragon auth token"})
		return
	}
	if errCode == http.StatusNotFound || token == "" {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "no paragon auth token found"})
		return
	}

	paragonProjectID := config.GetParagonProjectID()
	response, err := paragon.GetParagonProjectIntegrationsAPI(token, paragonProjectID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "fail to fetch response from paragon"})
		return
	}
	if response == nil {
		c.AbortWithStatusJSON(http.StatusExpectationFailed, gin.H{"error": "nil response from paragon"})
		return
	}

	c.JSON(http.StatusOK, response)
	return
}

func DeleteParagonProjectIntegrations(c *gin.Context) {
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)

	if projectID == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}

	token, errCode, err := store.GetStore().GetParagonTokenFromProjectSetting(projectID)
	if (errCode != http.StatusFound && errCode != http.StatusNotFound) || err != nil {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "failed to get paragon auth token"})
		return
	}
	if errCode == http.StatusNotFound || token == "" {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "no paragon auth token found"})
		return
	}

	integrationID := c.Query("integration_id")
	paragonProjectID := config.GetParagonProjectID()
	response, err := paragon.DeleteParagonProjectIntegrationAPI(token, paragonProjectID, integrationID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "fail to fetch response from paragon"})
		return
	}

	if response["status"] != "success" {
		c.AbortWithStatusJSON(http.StatusExpectationFailed, gin.H{"error": "failed to delete integration"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": "Integration deleted successfully"})
	return
}

func TriggerParagonWorkflow(c *gin.Context) {
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)

	if projectID == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}

	token, errCode, err := store.GetStore().GetParagonTokenFromProjectSetting(projectID)
	if (errCode != http.StatusFound && errCode != http.StatusNotFound) || err != nil {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "failed to get paragon auth token"})
		return
	}
	if errCode == http.StatusNotFound || token == "" {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "no paragon auth token found"})
		return
	}

	workflowID := c.Query("workflow_id")

	var payload interface{}
	r := c.Request
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "failed to decode payload"})
		return
	}

	paragonProjectID := config.GetParagonProjectID()
	response, err := paragon.TriggerParagonWorkflowAPI(token, paragonProjectID, workflowID, payload)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "fail to fetch response from paragon"})
		return
	}

	if response["status"] != "success" {
		c.AbortWithStatusJSON(http.StatusExpectationFailed, gin.H{"error": "failed to trigger workflow"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": "Integration deleted successfully"})
	return
}

func DisableParagonWorflowForUser(c *gin.Context) {
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)

	if projectID == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}

	token, errCode, err := store.GetStore().GetParagonTokenFromProjectSetting(projectID)
	if (errCode != http.StatusFound && errCode != http.StatusNotFound) || err != nil {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "failed to get paragon auth token"})
		return
	}
	if errCode == http.StatusNotFound || token == "" {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "no paragon auth token found"})
		return
	}

	workflowID := c.Query("workflow_id")
	paragonProjectID := config.GetParagonProjectID()
	response, err := paragon.DeleteParagonProjectIntegrationAPI(token, paragonProjectID, workflowID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "fail to fetch response from paragon"})
		return
	}

	if response["status"] != "success" {
		c.AbortWithStatusJSON(http.StatusExpectationFailed, gin.H{"error": "failed to delete workflow"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": "Workflow disabled successfully"})
	return
}
