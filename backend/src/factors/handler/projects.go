package handler

import (
	"encoding/json"
	mid "factors/middleware"
	M "factors/model"
	PC "factors/pattern_client"
	U "factors/util"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// Test command.
// curl -H "Content-Type: application/json" -i -X POST http://localhost:8080/projects -d '{ "name": "project_name"}'
func CreateProjectHandler(c *gin.Context) {
	r := c.Request

	loggedInAgentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)
	logCtx := log.WithFields(log.Fields{
		"reqId":      U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID),
		"agent_uuid": loggedInAgentUUID,
	})

	var project M.Project
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

	billingAcc, errCode := M.GetBillingAccountByAgentUUID(loggedInAgentUUID)
	if errCode != http.StatusFound {
		logCtx.WithField("err_code", errCode).Error("CreateProject Failed, billing account error")
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Creating project failed"})
		return
	}

	_, errCode = M.CreateProjectWithDependencies(&project, loggedInAgentUUID, M.ADMIN, billingAcc.ID)
	if errCode != http.StatusCreated {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Creating project failed."})
		return
	}

	c.JSON(http.StatusCreated, project)
	return
}

// Test command.
// curl -H "Content-Type: application/json" -i -X PUT http://localhost:8080/projects/1 -d '{ "name": "project_name"}'
func EditProjectHandler(c *gin.Context) {
	r := c.Request

	loggedInAgentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)
	projectID := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)

	logCtx := log.WithFields(log.Fields{
		"reqId":      U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID),
		"agent_uuid": loggedInAgentUUID,
		"project_id": projectID,
	})

	if projectID == 0 {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	loggedInAgentPAM, errCode := M.GetProjectAgentMapping(projectID, loggedInAgentUUID)
	if errCode != http.StatusFound {
		c.AbortWithStatus(http.StatusInternalServerError)
		logCtx.Errorln("Failed to fetch loggedInAgentPAM")
		return
	}

	if loggedInAgentPAM.Role != M.ADMIN {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "operation denied for non-admins"})
		return
	}

	if errCode == http.StatusInternalServerError {
		c.AbortWithStatus(errCode)
		return
	}
	var projectEditDetails M.Project

	err := json.NewDecoder(r.Body).Decode(&projectEditDetails)
	if err != nil {
		logCtx.WithError(err).Error("EditProject Failed. Json Decoding failed.")
	}

	errCode = M.UpdateProject(projectID, &projectEditDetails)
	if errCode == http.StatusInternalServerError {
		c.AbortWithStatus(errCode)
		return
	}
	projectIdsToGet := []uint64{}
	projectIdsToGet = append(projectIdsToGet, projectID)
	projectDetailsAfterEdit, errCode := M.GetProjectsByIDs(projectIdsToGet)
	c.JSON(http.StatusCreated, projectDetailsAfterEdit[0])
	return
}

// Test command.
// curl -i -X GET http://localhost:8080/projects
func GetProjectsHandler(c *gin.Context) {
	authorizedProjects := U.GetScopeByKey(c, "authorizedProjects")

	projects, errCode := M.GetProjectsByIDs(authorizedProjects.([]uint64))
	if errCode == http.StatusInternalServerError {
		c.AbortWithStatus(errCode)
		return
	} else if errCode == http.StatusNoContent || errCode == http.StatusBadRequest {
		resp := make(map[string]interface{})
		resp["projects"] = []M.Project{}
		c.JSON(http.StatusNotFound, resp)
		return
	}
	resp := make(map[string]interface{})
	resp["projects"] = projects
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

	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	modelIntervals, err := PC.GetProjectModelIntervals(reqId, projectId)
	if err != nil {
		logCtx.WithError(err).Error("falied to get projectModelIntervals")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, modelIntervals)
}
