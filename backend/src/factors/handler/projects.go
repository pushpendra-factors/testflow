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

	logCtx := log.WithFields(log.Fields{
		"reqId": U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID),
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

	var errCode int
	_, errCode = M.CreateProjectWithDependencies(&project)
	if errCode != http.StatusCreated {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Creating project failed."})
		return
	}
	loggedInAgentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)

	// create project agent mapping
	_, errCode = M.CreateProjectAgentMapping(&M.ProjectAgentMapping{
		ProjectID: project.ID,
		AgentUUID: loggedInAgentUUID,
		Role:      M.ADMIN,
	})
	if errCode != http.StatusCreated {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Creating project failed."})
		return
	}
	c.JSON(http.StatusCreated, project)
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
