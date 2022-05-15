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

// Test Command
// curl -i -X GET http://localhost:8080/projects/1/settings
// GetProjectSettingHandler godoc
// @Summary Retrieves the project settings for given project id.
// @Tags ProjectSettings
// @Produce json
// @Param project_id path integer true "Project ID"
// @Success 200 {object} model.ProjectSetting
// @Router /{project_id}/settings [get]
func GetProjectSettingHandler(c *gin.Context) {

	logCtx := log.WithFields(log.Fields{
		"reqId": U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID),
	})

	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		logCtx.Error("Get project_settings failed. Failed to get project_id.")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	settings, errCode := store.GetStore().GetProjectSetting(projectId)
	if errCode != http.StatusFound {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Failed to get project settings."})
	} else {
		c.JSON(http.StatusOK, settings)
	}
}

// Test Command
// curl -i -H "Content-UnitType: application/json" -X PUT http://localhost:8080/projects/1/settings -d '{"auto_track": false}'
// UpdateProjectSettingsHandler godoc
// @Summary Update the project settings for given project id.
// @Tags ProjectSettings
// @Accept json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param setting body model.ProjectSetting true "Project Setting"
// @Success 200 {object} model.ProjectSetting
// @Router /{project_id}/settings [put]
func UpdateProjectSettingsHandler(c *gin.Context) {
	r := c.Request

	logCtx := log.WithFields(log.Fields{
		"reqId": U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID),
	})

	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		logCtx.Error("Update project_settings failed. Failed to get project_id.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid project id."})
		return
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	var projectSetting model.ProjectSetting
	if err := decoder.Decode(&projectSetting); err != nil {
		logCtx.WithError(err).Error(
			"Project setting update failed. Json Decoding failed.")
		c.AbortWithStatusJSON(http.StatusBadRequest,
			gin.H{"error": "Project setting update failed. Invalid payload."})
		return
	}

	// project_id sent on json_payload should be same as project_id on uri param, if given.
	if projectSetting.ProjectId != 0 && projectId != projectSetting.ProjectId {
		c.AbortWithStatusJSON(http.StatusBadRequest,
			gin.H{"error": "Project setting update failed. Invalid payload."})
		return
	}

	updatedPSetting, errCode := store.GetStore().UpdateProjectSettings(projectId, &projectSetting)
	if errCode != http.StatusAccepted {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Failed to update project settings."})
		return
	}

	c.JSON(http.StatusOK, &updatedPSetting)
}
