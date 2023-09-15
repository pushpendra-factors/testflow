package v1

import (
	"encoding/json"
	mid "factors/middleware"
	"factors/model/model"
	"factors/model/store"
	T "factors/task"
	U "factors/util"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func CreateAlertHandler(c *gin.Context) (interface{}, int, string, string, bool) {
	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		return nil, http.StatusUnauthorized, INVALID_PROJECT, ErrorMessages[INVALID_PROJECT], true
	}

	r := c.Request
	queryIdString := c.Query("query_id")
	var queryID int64
	var err error
	if queryIdString != "" {
		queryID, err = strconv.ParseInt(queryIdString, 10, 64)
		if err != nil {
			log.Error("failed to parse queryID string")
			return nil, http.StatusInternalServerError, PROCESSING_FAILED, "failed to parse queryID string", false
		}
	}
	var alertPayload model.Alert
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&alertPayload); err != nil {
		log.WithError(err).Error("Failed to decode Json request on create alert handler.")
		return nil, http.StatusInternalServerError, PROCESSING_FAILED, "Failed to decode Json request on create alert handler.", true
	}
	alertPayload.CreatedBy = U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)
	alertPayload.QueryID = queryID
	alert, errCode, errMsg := store.GetStore().CreateAlert(projectId, alertPayload)
	if errCode != http.StatusCreated {
		log.WithFields(log.Fields{"document": alert, "err-message": errMsg}).Error("Failed to insertalert on create create alert handler.")
		return nil, errCode, PROCESSING_FAILED, errMsg, true
	}

	return alert, http.StatusCreated, "", "", false
}

func EditAlertHandler(c *gin.Context) (interface{}, int, string, string, bool) {

	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		log.Error("Failed to update alert, ProjectId parse failed.")
		return nil, http.StatusUnauthorized, INVALID_PROJECT, ErrorMessages[INVALID_PROJECT], true
	}
	alertID := c.Param("id")
	if alertID == "" {
		log.Error("Failed to update alert. failed to parse id")
		return nil, http.StatusBadRequest, INVALID_INPUT, "ID parse failed", true
	}
	var editAlertPayload model.Alert
	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&editAlertPayload); err != nil {
		log.WithError(err).Error("Failed to decode Json request on update alert handler.")
		return nil, http.StatusInternalServerError, PROCESSING_FAILED, "Failed to decode Json request on udpate alert handler.", true
	}
	editAlertPayload.CreatedBy = U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)
	alert, errCode, errMsg := store.GetStore().UpdateAlert(projectID, alertID, editAlertPayload)
	if errCode != http.StatusAccepted {
		log.WithFields(log.Fields{"document": alert, "err-message": errMsg}).Error("Failed to insert alert on create create alert handler.")
		return nil, errCode, PROCESSING_FAILED, errMsg, true
	}

	alert, errCode = store.GetStore().GetAlertById(alertID, projectID)
	if errCode != http.StatusFound {
		return nil, errCode, PROCESSING_FAILED, "Failed to get updated alert", true
	}
	return alert, http.StatusOK, "", "", false

}

func GetAlertsHandler(c *gin.Context) (interface{}, int, string, string, bool) {
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		log.Error("Get Alerts Failed, failed to parse project ID.")
		return nil, http.StatusUnauthorized, INVALID_PROJECT, ErrorMessages[INVALID_PROJECT], true
	}
	excludeSavedQueriesFlag := true
	// negation of flag to include/exlude returning saved queries i.e KPI/Events
	includeSavedQueries := c.Query("saved_queries")
	if includeSavedQueries == "true" {
		excludeSavedQueriesFlag = false
	}
	//except sheduled shared saved reports (passing true to exclude saved queries)
	alerts, errCode := store.GetStore().GetAllAlerts(projectID, excludeSavedQueriesFlag)
	if errCode != http.StatusFound {
		return nil, errCode, PROCESSING_FAILED, "Failed to get alerts", true
	}
	return alerts, http.StatusOK, "", "", false
}

func GetAlertByIDHandler(c *gin.Context) (interface{}, int, string, string, bool) {
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
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
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
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
		return nil, errCode, PROCESSING_FAILED, "Failed to delete alert ", true
	}
	return nil, http.StatusOK, "", "", false
}
func QuerySendNowHandler(c *gin.Context) {
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	loggedInAgentUUID := U.GetScopeByKeyAsString(c, mid.SCOPE_LOGGEDIN_AGENT_UUID)
	if loggedInAgentUUID == "" || projectID == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest,
			gin.H{"error": "Invalid agent id. or project id"})
		return
	}
	queryIdString := c.Query("query_id")
	overrideDateRangeString := c.Query("override_date_range")
	var shouldOverride bool
	var overRideFrom, overRideTo int64
	var err error
	if overrideDateRangeString == "true" {
		shouldOverride = true
		overRideFrom, err = strconv.ParseInt(c.Query("from_time"), 10, 64)
		if err != nil {
			log.Error("Failed to parse override date range")
			return
		}
		overRideTo, err = strconv.ParseInt(c.Query("to_time"), 10, 64)
		if err != nil {
			log.Error("Failed to parse override date range")
			return
		}
	}
	var queryID int64
	if queryIdString != "" {
		queryID, err = strconv.ParseInt(queryIdString, 10, 64)
		if err != nil {
			log.Error("failed to parse queryID string")
			c.AbortWithStatusJSON(http.StatusBadRequest,
				gin.H{"error": "Failed to parse query ID"})
			return
		}
	}
	r := c.Request
	var alert model.Alert
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&alert); err != nil {
		log.WithError(err).Error("Failed to decode Json request on send now handler.")
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			gin.H{"error": "failed to decode json request on send now handler"})
		return
	}
	alert.ProjectID = projectID
	alert.CreatedBy = loggedInAgentUUID
	alert.QueryID = queryID
	_, err = T.HandlerAlertWithQueryID(alert, nil, shouldOverride, overRideFrom, overRideTo)
	if err != nil {
		log.WithError(err).Error("failed to perform send now operation for query id ", alert.QueryID)
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			gin.H{"error": "failed to perform send now operation"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"msg": "sent report successfully"})
}
