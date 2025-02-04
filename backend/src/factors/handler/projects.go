package handler

import (
	"encoding/json"
	mid "factors/middleware"
	"factors/model/model"
	"factors/model/store"
	PC "factors/pattern_client"
	U "factors/util"
	"net/http"
	"strings"

	V1 "factors/handler/v1"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// Test command.
// curl -H "Content-UnitType: application/json" -i -X POST http://localhost:8080/projects -d '{ "name": "project_name"}'
func CreateProjectHandler(c *gin.Context) {
	r := c.Request

	loggedInAgentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)
	logCtx := log.WithFields(log.Fields{
		"reqId":      U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID),
		"agent_uuid": loggedInAgentUUID,
	})
	createDefaultDashBoard := c.Query("create_dashboard")
	var createDashboard bool = true // by default
	if createDefaultDashBoard == "false" {
		createDashboard = false
	}
	var project model.Project
	err := json.NewDecoder(r.Body).Decode(&project)
	if err != nil {
		logCtx.WithError(err).Error("CreateProject Failed. Json Decoding failed.")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "json decoding : " + err.Error(),
			"status": http.StatusBadRequest,
		})
		return
	}

	if project.Name == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	if !U.IsUserOrProjectNameValid(project.Name) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid character"})
		return
	}
	if project.ProfilePicture != "" && !(strings.HasPrefix(project.ProfilePicture, "data:image/png") || strings.HasPrefix(project.ProfilePicture, "data:image/jpeg")) {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	billingAcc, errCode := store.GetStore().GetBillingAccountByAgentUUID(loggedInAgentUUID)
	if errCode != http.StatusFound {
		logCtx.WithField("err_code", errCode).Error("CreateProject Failed, billing account error")
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Creating project failed"})
		return
	}

	updatedProject, errCode := store.GetStore().CreateProjectWithDependencies(&project, loggedInAgentUUID, model.ADMIN, billingAcc.ID, createDashboard)
	if errCode != http.StatusCreated {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Creating project failed."})
		return
	}

	_, errCode = store.GetStore().CreateWidgetGroups(updatedProject.ID)
	if errCode != http.StatusCreated {
		logCtx.WithField("err_code", errCode).Error("CreateProject Failed, Create widget groups failed.")
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Creating project failed."})
		return
	}

	c.JSON(http.StatusCreated, V1.MapProjectToString(*updatedProject))
	return
}

// Test command.
// curl -H "Content-UnitType: application/json" -i -X PUT http://localhost:8080/projects/1 -d '{ "name": "project_name"}'
// EditProjectHandler godoc
// @Summary To edit the allowed fields of an existing project.
// @Tags Projects
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param unit body model.Project true "Edit project"
// @Success 201 {object} model.Project
// @Router /{project_id} [put]
func EditProjectHandler(c *gin.Context) {
	r := c.Request

	loggedInAgentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)

	logCtx := log.WithFields(log.Fields{
		"reqId":      U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID),
		"agent_uuid": loggedInAgentUUID,
		"project_id": projectID,
	})

	if projectID == 0 {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	loggedInAgentPAM, errCode := store.GetStore().GetProjectAgentMapping(projectID, loggedInAgentUUID)
	if errCode != http.StatusFound {
		c.AbortWithStatus(http.StatusInternalServerError)
		logCtx.Errorln("Failed to fetch loggedInAgentPAM")
		return
	}

	if loggedInAgentPAM.Role != model.ADMIN {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "operation denied for non-admins"})
		return
	}

	if errCode == http.StatusInternalServerError {
		c.AbortWithStatus(errCode)
		return
	}
	var projectEditDetails model.Project

	err := json.NewDecoder(r.Body).Decode(&projectEditDetails)
	if err != nil {
		logCtx.WithError(err).Error("EditProject Failed. Json Decoding failed.")
	}
	if !U.IsUserOrProjectNameValid(projectEditDetails.Name) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid character"})
		return
	}
	if projectEditDetails.ProfilePicture != "" && !(strings.HasPrefix(projectEditDetails.ProfilePicture, "data:image/png") || strings.HasPrefix(projectEditDetails.ProfilePicture, "data:image/jpeg")) {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	errCode = store.GetStore().UpdateProject(projectID, &projectEditDetails)
	if errCode == http.StatusInternalServerError {
		c.AbortWithStatus(errCode)
		return
	}
	projectIdsToGet := []int64{}
	projectIdsToGet = append(projectIdsToGet, projectID)
	projectDetailsAfterEdit, errCode := store.GetStore().GetProjectsByIDs(projectIdsToGet)
	c.JSON(http.StatusCreated, V1.MapProjectToString(projectDetailsAfterEdit[0]))
	return
}

// Test command.
// curl -i -X GET http://localhost:8080/projects
// GetProjectsHandler godoc
// @Summary To fetch the list of authorized projects for the user.
// @Tags Projects
// @Accept  json
// @Produce json
// @Success 200 {string} json "{"projects": []Project}"
// @Router / [get]
func GetProjectsHandler(c *gin.Context) {
	authorizedProjects := U.GetScopeByKey(c, mid.SCOPE_AUTHORIZED_PROJECTS)

	projects, errCode := store.GetStore().GetProjectsByIDs(authorizedProjects.([]int64))
	if errCode == http.StatusInternalServerError {
		c.AbortWithStatus(errCode)
		return
	} else if errCode == http.StatusNoContent || errCode == http.StatusBadRequest {
		resp := make(map[string]interface{})
		resp["projects"] = []model.ProjectString{}
		c.JSON(http.StatusNotFound, resp)
		return
	}

	resp := make(map[string][]model.ProjectString)
	for _, project := range projects {
		resp["projects"] = append(resp["projects"], V1.MapProjectToString(project))
	}
	c.JSON(http.StatusOK, resp)
	return
}

// curl -i -X GET http://localhost:8080/projects/1/models
// GetProjectModelsHandler godoc
// @Summary To get model infos for the given project id.
// @Tags Factors
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Success 200 {array} string "{"mid": uint64, "mt": string, "st": timestamp, "et": timestamp}"
// @Router /{project_id}/models [get]
func GetProjectModelsHandler(c *gin.Context) {
	reqId := U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID)
	logCtx := log.WithFields(log.Fields{
		"reqId": reqId,
	})

	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	modelIntervals := make([]PC.ModelInfo, 0)
	modelMetadata, errCode, msg := store.GetStore().GetProjectModelMetadata(projectId)
	if errCode != http.StatusFound {
		logCtx.Error(msg)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	for _, metadata := range modelMetadata {
		modelIntervals = append(modelIntervals, PC.ModelInfo{ModelId: metadata.ModelId,
			ModelType:      metadata.ModelType,
			StartTimestamp: metadata.StartTime,
			EndTimestamp:   metadata.EndTime})
	}
	c.JSON(http.StatusOK, modelIntervals)
}

func GetProjectHandler(c *gin.Context) {
	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid project id."})
		return
	}

	project, errCode := store.GetStore().GetProject(projectId)
	if errCode != http.StatusFound {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Failed to get project."})
		return
	}

	c.JSON(http.StatusOK, V1.MapProjectToString(*project))
	return
}
