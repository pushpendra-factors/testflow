package v1

import (
	mid "factors/middleware"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type ProjectSettings struct {
	Settings     model.ProjectSetting `json:"project_settings"`
	IntCompleted bool                 `json:"int_completed"`
	IntSlack     bool                 `json:"int_slack"`
	IntTeams     bool                 `json:"int_teams"`
}

func GetProjectSettingHandler(c *gin.Context) {

	logCtx := log.WithFields(log.Fields{
		"reqId": U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID),
	})

	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	agentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)
	if projectId == 0 {
		logCtx.Error("Get project_settings failed. Failed to get project_id.")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	isSlackIntegrated, errCode := store.GetStore().IsSlackIntegratedForProject(projectId, agentUUID)
	if errCode != http.StatusOK {
		logCtx.Error("Get slack integration status failed.")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	isTeamsIntegrated, errCode := store.GetStore().IsTeamsIntegratedForProject(projectId, agentUUID)
	if errCode != http.StatusOK {
		logCtx.Error("Get teams integration status failed.")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	settings, errCode := store.GetStore().GetProjectSetting(projectId)
	if errCode != http.StatusFound {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Failed to get project settings."})
		return
	}
	isExist, errCode := store.GetStore().IsEventExistsWithType(projectId, model.TYPE_AUTO_TRACKED_EVENT_NAME)
	int_completed := isExist

	projectSettings := ProjectSettings{
		Settings:     *settings,
		IntCompleted: int_completed,
		IntSlack:     isSlackIntegrated,
		IntTeams:     isTeamsIntegrated,
	}
	if errCode != http.StatusFound && errCode != http.StatusNotFound {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Failed to get project settings."})
	} else {
		// Returning random url sets everytime.
		// Todo: Persist url at project level and return the same.
		assetURL, apiURL := model.GetProjectSDKAPIAndAssetURL(projectId)
		projectSettings.Settings.SDKAPIURL = apiURL
		projectSettings.Settings.SDKAssetURL = assetURL

		c.JSON(http.StatusOK, projectSettings)
	}
}
