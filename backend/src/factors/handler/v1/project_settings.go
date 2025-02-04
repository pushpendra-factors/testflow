package v1

import (
	"encoding/json"
	slack "factors/integration/slack"
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

func IntegrationsStatusHandler(c *gin.Context) {
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

	settings, errCode := store.GetStore().GetProjectSetting(projectId)
	if errCode != http.StatusFound {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Failed to get project settings."})
		return
	}

	var intStatusMap map[string]string
	if settings.IntegrationStatus != nil {
		err := json.Unmarshal(settings.IntegrationStatus.RawMessage, &intStatusMap)
		if err != nil {
			logCtx.WithField("integration status", settings.IntegrationStatus).WithError(err).Error("Failed to decode integration status ")
		}
	}

	result := map[string]model.IntegrationState{}
	for _, integrationName := range model.IntegrationNameList {
		if statusString, ok := intStatusMap[integrationName]; ok && statusString != model.SUCCESS {

			result[integrationName] = model.IntegrationState{
				State:   statusString,
				Message: model.ErrorStateToErrorMessageMap[statusString],
			}

		} else {
			state, errCode := store.GetStore().GetIntegrationState(projectId, integrationName)
			if errCode != http.StatusOK {
				logCtx.Error("Failed to get Integration state")
			}
			result[integrationName] = state
		}
	}

	result[model.FEATURE_SLACK] = slack.GetSlackIntegrationState(projectId, agentUUID)

	for _, integrationName := range []string{model.FEATURE_FACTORS_DEANONYMISATION, model.FEATURE_SIX_SIGNAL, model.FEATURE_CLEARBIT} {
		if statusString, ok := intStatusMap[integrationName]; ok {
			result[integrationName] = model.IntegrationState{
				State:   statusString,
				Message: model.ErrorStateToErrorMessageMap[statusString],
			}
		}
	}

	connected, disconnected := store.GetStore().GetIntegrationStatusesCount(*settings, projectId, agentUUID)

	for _, integrationDisplayName := range disconnected["disconnected"].([]string) {
		result[model.FeatureDisplayNameMap[integrationDisplayName]] = model.IntegrationState{State: model.DISCONNECTED}
	}

	for _, integrationDisplayName := range connected["connected"].([]string) {

		if _, ok := result[model.FeatureDisplayNameMap[integrationDisplayName]]; !ok {
			result[model.FeatureDisplayNameMap[integrationDisplayName]] = model.IntegrationState{State: model.CONNECTED}
		}

	}

	c.JSON(http.StatusOK, result)

}
