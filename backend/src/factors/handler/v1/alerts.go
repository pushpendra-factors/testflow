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

func CreateAlertHandler(c *gin.Context) (interface{}, int, string, string, bool) {
	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		return nil, http.StatusUnauthorized, INVALID_PROJECT, ErrorMessages[INVALID_PROJECT], true
	}

	r := c.Request

	var alertPayload model.Alert
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&alertPayload); err != nil {
		log.WithError(err).Error("Failed to decode Json request on create alert handler.")
		return nil, http.StatusInternalServerError, PROCESSING_FAILED, "Failed to decode Json request on create alert handler.", true
	}
	alertPayload.CreatedBy = U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)
	alert, errCode, errMsg := store.GetStore().CreateAlert(projectId, alertPayload)
	if errCode != http.StatusCreated {
		log.WithFields(log.Fields{"document": alert, "err-message": errMsg}).Error("Failed to insertalert on create create alert handler.")
		return nil, errCode, PROCESSING_FAILED, errMsg, true
	}

	return alert, http.StatusCreated, "", "", false
}

func GetAlertsHandler(c *gin.Context) (interface{}, int, string, string, bool) {
	projectID := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		log.Error("Get Alerts Failed, failed to parse project ID.")
		return nil, http.StatusUnauthorized, INVALID_PROJECT, ErrorMessages[INVALID_PROJECT], true
	}
	alerts, errCode := store.GetStore().GetAllAlerts(projectID)
	if errCode != http.StatusFound {
		return nil, errCode, PROCESSING_FAILED, "Failed to get alerts", true
	}
	return alerts, http.StatusOK, "", "", false
}

func GetAlertByIDHandler(c *gin.Context) (interface{}, int, string, string, bool) {
	projectID := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		log.Error("Failed to get alert, failed to parse Project")
		return nil, http.StatusUnauthorized, INVALID_PROJECT, ErrorMessages[INVALID_PROJECT], true
	}
	alertID := c.Param("id")
	if alertID == "" {
		log.Error("Get Alert Failed. no id found")
		return nil, http.StatusBadRequest, INVALID_INPUT, "failed to get id", true
	}

	alert, errCode := store.GetStore().GetAlertById(alertID, projectID)
	if errCode != http.StatusFound {
		return nil, errCode, PROCESSING_FAILED, "Failed to get alert", true
	}
	return alert, http.StatusOK, "", "", false
}

func DeleteAlertHandler(c *gin.Context) (interface{}, int, string, string, bool) {
	projectID := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		log.Error("Failed to delete alert, ProjectId parse failed.")
		return nil, http.StatusUnauthorized, INVALID_PROJECT, ErrorMessages[INVALID_PROJECT], true
	}

	alertID := c.Param("id")
	if alertID == "" {
		log.Error("Failed to delete alert. failed to parse id")
		return nil, http.StatusBadRequest, INVALID_INPUT, "ID parse failed", true
	}

	errCode, errMsg := store.GetStore().DeleteAlert(alertID, projectID)
	if errCode != http.StatusAccepted {
		log.Error("failed to delete alert" + errMsg)
		return nil, errCode, PROCESSING_FAILED, "Failed to delete alert " , true
	}
	return nil, http.StatusOK, "", "", false
}
