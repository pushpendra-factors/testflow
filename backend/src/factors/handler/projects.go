package handler

import (
	"encoding/json"
	M "factors/model"
	U "factors/util"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// Test command.
// curl -H "Content-Type: application/json" -i -X POST http://localhost:8080/projects -d '{ "name": "project_name"}'
func CreateProjectHandler(c *gin.Context) {
	r := c.Request

	var project M.Project
	err := json.NewDecoder(r.Body).Decode(&project)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("CreateProject Failed. Json Decoding failed.")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "json decoding : " + err.Error(),
			"status": http.StatusBadRequest,
		})
		return
	}

	var errCode int
	_, errCode = M.CreateProjectWithDependencies(&project)
	if errCode != http.StatusCreated {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Creating project failed."})
	} else {
		c.JSON(http.StatusCreated, project)
	}
}

// Test command.
// curl -i -X GET http://localhost:8080/projects
func GetProjectsHandler(c *gin.Context) {
	authorizedProjects := U.GetScopeByKey(c, "authorizedProjects")

	projects, errCode := M.GetProjectsByIDs(authorizedProjects.([]uint64))
	if errCode != http.StatusFound {
		c.AbortWithStatus(errCode)
	} else {
		c.JSON(http.StatusOK, projects)
	}
}
