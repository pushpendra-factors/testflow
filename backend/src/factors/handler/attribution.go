package handler

import (
	"encoding/json"
	mid "factors/middleware"
	M "factors/model"
	U "factors/util"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type AttributionRequestPayload struct {
	Query *M.AttributionQuery `json:"query"`
}

func AttributionHandler(c *gin.Context) {

	r := c.Request
	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Query failed. Invalid project."})
		return
	}
	logCtx := log.WithFields(log.Fields{
		"reqId": U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID), "project_id": projectId,
	})

	var err error
	var requestPayload AttributionRequestPayload
	var dashboardId uint64
	var unitId uint64

	hardRefresh := false
	dashboardIdParam := c.Query("dashboard_id")
	unitIdParam := c.Query("dashboard_unit_id")
	refreshParam := c.Query("refresh")
	if refreshParam != "" {
		hardRefresh, _ = strconv.ParseBool(refreshParam)
	}
	if dashboardIdParam != "" || unitIdParam != "" {
		dashboardId, err = strconv.ParseUint(dashboardIdParam, 10, 64)
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		unitId, err = strconv.ParseUint(unitIdParam, 10, 64)
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&requestPayload); err != nil {
		logCtx.WithError(err).Error("Query failed. Json decode failed.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Query failed. Json decode failed."})
		return
	}

	// If refresh is passed, refresh only is Query.From is of todays beginning.
	if (dashboardIdParam != "" || unitIdParam != "") && !isHardRefreshForToday(requestPayload.Query.From, hardRefresh) {

		cacheResult, errCode, errMsg := M.GetCacheResultByDashboardIdAndUnitId(projectId, dashboardId, unitId, requestPayload.Query.From, requestPayload.Query.To)
		if errCode == http.StatusFound && cacheResult != nil {
			c.JSON(http.StatusOK, gin.H{"result": cacheResult.Result, "cache": true, "refreshed_at": cacheResult.RefreshedAt})
			return
		}
		if errCode == http.StatusBadRequest {
			c.AbortWithStatusJSON(errCode, gin.H{"error": errMsg})
			return
		}
		if errCode != http.StatusNotFound {
			logCtx.WithFields(log.Fields{"project_id": projectId,
				"dashboard_id": dashboardIdParam, "dashboard_unit_id": unitIdParam,
			}).WithError(errMsg).Error("Failed to get Dashboard Cached Result for Attribution query.")
		}
	}

	result, err := M.ExecuteAttributionQuery(projectId, requestPayload.Query)
	if err != nil {
		logCtx.WithError(err).Error("Query failed. Query execution failed")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Query failed. Query execution failed"})
		return
	}

	if dashboardId != 0 && unitId != 0 {
		M.SetCacheResultByDashboardIdAndUnitId(result, projectId, dashboardId, unitId, requestPayload.Query.To, requestPayload.Query.From)
	}
	c.JSON(http.StatusOK, gin.H{"result": result, "cache": false, "refreshed_at": U.TimeNowIn(U.TimeZoneStringIST).Unix()})
}
