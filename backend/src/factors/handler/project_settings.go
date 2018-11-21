package handler

import (
	"encoding/json"
	M "factors/model"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// Test Command
// curl -i -X GET http://localhost:8080/projects/1/settings
func GetProjectSettingHandler(c *gin.Context) {
	projectId, err := strconv.ParseUint(c.Params.ByName("project_id"), 10, 64)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("GetEvent Failed. ProjectId parse failed.")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	settings, errCode := M.GetProjectSetting(projectId)
	if errCode != M.DB_SUCCESS {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Failed to get project settings."})
	} else {
		c.JSON(http.StatusOK, settings)
	}
}

// Test Command
// curl -i -H "Content-Type: application/json" -X PUT http://localhost:8080/projects/1/settings -d '{"auto_track": false}'
func UpdateProjectSettingsHandler(c *gin.Context) {
	r := c.Request

	projectId, err := strconv.ParseUint(c.Params.ByName("project_id"), 10, 64)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("GetEvent Failed. ProjectId parse failed.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid project id."})
		return
	}

	var projectSetting M.ProjectSetting
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&projectSetting); err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Project setting update failed. Json Decoding failed.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Project setting update failed. Invalid payload."})
		return
	}

	if projectSetting.ProjectId != 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Project setting failed. Tried updating disallowed field."})
		return
	}

	updatedPSetting, errCode := M.UpdateProjectSettings(projectId, &projectSetting)
	if errCode != M.DB_SUCCESS {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Failed to update project settings."})
		return
	}

	c.JSON(http.StatusOK, &updatedPSetting)
}
