package v1

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

func GetAllWorkflowTemplatesHandler(c *gin.Context) {
	workflowTemplates, errCode := store.GetStore().GetAllWorkflowTemplates()
	if errCode != http.StatusOK {
		c.AbortWithStatus(errCode)
	}

	c.JSON(errCode, workflowTemplates)
}

func GetAllSavedWorkflowsHandler(c *gin.Context) (interface{}, int, string, string, bool) {
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)

	if projectID == 0 {
		return nil, http.StatusBadRequest, INVALID_PROJECT, "Invalid project ID.", true
	}

	workflows, errCode, err := store.GetStore().GetAllWorklfowsByProject(projectID)
	if err != nil {
		return nil, errCode, PROCESSING_FAILED, "Failed to get saved workflows.", true
	}

	return workflows, http.StatusOK, "", "", false
}

func CreateWorkflowHandler(c *gin.Context) (interface{}, int, string, string, bool) {
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		return nil, http.StatusBadRequest, INVALID_INPUT, ErrorMessages[INVALID_INPUT], true
	}

	agentID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)
	if agentID == "" {
		return nil, http.StatusBadRequest, INVALID_INPUT, ErrorMessages[INVALID_INPUT], true
	}

	var workflow model.WorkflowAlertBody
	r := c.Request
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&workflow); err != nil {
		return nil, http.StatusBadRequest, INVALID_INPUT, "Unable to decode workflow body.", true
	}

	obj, errCode, err := store.GetStore().CreateWorkflow(projectID, agentID, "", workflow)
	if err != nil {
		return nil, errCode, PROCESSING_FAILED, err.Error(), true
	}

	return obj.AlertBody, http.StatusCreated, "", "", false
}

func DeleteWorkflowHandler(c *gin.Context) {
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Delete workflow failed. Invalid project."})
		return
	}

	id := c.Param("id")
	if id == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Delete failed. Invalid id provided."})
		return
	}

	agentID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)
	if agentID == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid agent."})
		return
	}

	errCode, err := store.GetStore().DeleteWorkflow(projectID, id, agentID)
	if errCode != http.StatusAccepted {
		c.AbortWithStatusJSON(errCode, err)
		return
	}

	c.JSON(errCode, gin.H{"Status": "OK"})
}

func EditWorkflowHandler(c *gin.Context) (interface{}, int, string, string, bool) {
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		return nil, http.StatusBadRequest, INVALID_PROJECT, ErrorMessages[INVALID_PROJECT], true
	}

	id := c.Param("id")
	if id == "" {
		return nil, http.StatusBadRequest, PROCESSING_FAILED, ErrorMessages[INVALID_INPUT], true
	}

	agentID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)
	if agentID == "" {
		return nil, http.StatusBadRequest, PROCESSING_FAILED, ErrorMessages[INVALID_INPUT], true
	}

	var workflow model.WorkflowAlertBody
	r := c.Request
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&workflow); err != nil {
		return nil, http.StatusBadRequest, PROCESSING_FAILED, ErrorMessages[PROCESSING_FAILED], true
	}

	newWorkflow, errCode, err := store.GetStore().CreateWorkflow(projectID, agentID, id, workflow)
	if err != nil {
		log.WithError(err).Error("Failed to edit workflow.")
		return nil, errCode, PROCESSING_FAILED, "Failed to edit workflow.", true
	}

	errCode, err = store.GetStore().DeleteWorkflow(projectID, id, agentID)
	if err != nil{
		return nil, errCode, PROCESSING_FAILED, "Failed to edit workflow.", true
	}

	return newWorkflow, http.StatusAccepted, "", "", false
}