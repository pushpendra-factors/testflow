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

func GetEventTriggerAlertsByProjectHandler(c *gin.Context) (interface{}, int, string, string, bool) {
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)

	if projectID == 0 {
		return nil, http.StatusForbidden, "", "Get request failed. Invalid project ID.", true
	}
	trigger, errCode := store.GetStore().GetAllEventTriggerAlertsByProject(projectID)
	if errCode != http.StatusFound {
		return nil, errCode, "", "Get Saved Queries failed.", true
	}

	return trigger, http.StatusOK, "", "", false

}

func CreateEventTriggerAlertHandler(c *gin.Context) (interface{}, int, string, string, bool) {

	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	userID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)

	if projectID == 0 {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Create failed. Invalid project id."})
		return nil, http.StatusUnauthorized, INVALID_PROJECT, ErrorMessages[INVALID_PROJECT], true
	}
	log.Info("Create function handler triggered.")

	var alert model.EventTriggerAlertConfig
	r := c.Request
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&alert); err != nil {
		errMsg := "Create TriggerAlert failed. Invalid JSON."
		log.WithFields(log.Fields{"project_id": projectID}).WithError(err).Error(errMsg)
		return nil, http.StatusBadRequest, errMsg, "", true
	}

	obj, errCode, errMsg := store.GetStore().CreateEventTriggerAlert(userID, projectID, &alert)
	if errCode != http.StatusCreated {
		log.WithFields(log.Fields{"document": alert, "err-message": errMsg}).Error("Failed to create alert in handler.")
		return nil, errCode, PROCESSING_FAILED, errMsg, true
	}

	return obj.EventTriggerAlert, http.StatusCreated, "", "", false
}

func DeleteEventTriggerAlertHandler(c *gin.Context) {
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Delete TriggerAlert failed. Invalid project."})
		return
	}

	id := c.Param("id")
	if id == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Delete failed. Invalid id provided."})
		return
	}

	errCode, errMsg := store.GetStore().DeleteEventTriggerAlert(projectID, id)
	if errCode != http.StatusAccepted {
		c.AbortWithStatusJSON(errCode, errMsg)
		return
	}

	c.JSON(errCode, gin.H{"Status": "OK"})
}

func EditEventTriggerAlertHandler(c *gin.Context) (interface{}, int, string, string, bool) {
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Edit TriggerAlert failed. Invalid project."})
		return nil, http.StatusBadRequest, INVALID_PROJECT, ErrorMessages[INVALID_PROJECT], true
	}

	id := c.Param("id")
	if id == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Edit failed. Invalid id provided."})
		return nil, http.StatusBadRequest, PROCESSING_FAILED, ErrorMessages[PROCESSING_FAILED], true
	}
	userID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)

	errCode, errMsg := store.GetStore().DeleteEventTriggerAlert(projectID, id)
	if errCode != http.StatusAccepted || errMsg != "" {
		log.WithFields(log.Fields{"project_id": projectID}).Error("Cannot find any alert to update")
		return nil, http.StatusBadRequest, "Cannot find any alert to update", "", true
	}

	var alert model.EventTriggerAlertConfig

	r := c.Request
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&alert); err != nil {
		errMsg := "Edit TriggerAlert failed. Invalid JSON."
		log.WithFields(log.Fields{"project_id": projectID}).WithError(err).Error(errMsg)
		return nil, http.StatusBadRequest, errMsg, "", true
	}

	eta, errCode, errMsg := store.GetStore().CreateEventTriggerAlert(userID, projectID, &alert)
	if errMsg != "" || errCode != http.StatusCreated || eta == nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Edit TriggerAlert failed while updating db"})
		return nil, http.StatusInternalServerError, PROCESSING_FAILED, ErrorMessages[PROCESSING_FAILED], true
	}

	return alert, http.StatusAccepted, "", "", false
}
