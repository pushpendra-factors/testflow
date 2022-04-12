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
}

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
		return
	}
	events, errCode := store.GetStore().GetEventNames(projectId)
	var int_completed bool = false
	if len(events) > 1 {
		int_completed = true
	}
	projectSettings := ProjectSettings{
		Settings: *settings,
		IntCompleted: int_completed,
	}
	if errCode != http.StatusFound && errCode != http.StatusNotFound{
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Failed to get project settings."})
	} else {
		c.JSON(http.StatusOK, projectSettings)
	}
}
