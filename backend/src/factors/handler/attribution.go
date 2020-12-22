package handler

import (
	"bytes"
	"encoding/json"
	mid "factors/middleware"
	M "factors/model"
	U "factors/util"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type AttributionRequestPayload struct {
	Query *M.AttributionQuery `json:"query"`
}

// AttributionHandler godoc
// @Summary To run attribution query.
// @Tags CoreQuery
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param dashboard_id query integer false "Dashboard ID"
// @Param dashboard_unit_id query integer false "Dashboard Unit ID"
// @Param query body handler.AttributionRequestPayload true "Query payload"
// @Success 200 {string} json "{"result": model.QueryResult, "cache": false, "refreshed_at": timestamp}"
// @Router /{project_id}/attribution/query [post]
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

	hasFailed, errMsg, requestPayload := decodeAttributionPayload(r, logCtx)
	if hasFailed {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"Error": errMsg})
		return
	}

	// If refresh is passed, refresh only is Query.From is of today's beginning.
	if (dashboardIdParam != "" || unitIdParam != "") &&
		!isHardRefreshForToday(requestPayload.Query.From, hardRefresh) {

		cacheResult, errCode, errMsg := M.GetCacheResultByDashboardIdAndUnitId(projectId, dashboardId,
			unitId, requestPayload.Query.From, requestPayload.Query.To)
		if errCode == http.StatusFound && cacheResult != nil {
			c.JSON(http.StatusOK, gin.H{"result": cacheResult.Result, "cache": true,
				"refreshed_at": cacheResult.RefreshedAt})
			return
		}
		if errCode == http.StatusBadRequest {
			c.AbortWithStatusJSON(errCode, gin.H{"error": errMsg})
			return
		}
		if errCode != http.StatusNotFound {
			logCtx.WithFields(log.Fields{"project_id": projectId,
				"dashboard_id": dashboardIdParam, "dashboard_unit_id": unitIdParam,
			}).WithError(errMsg).Error("failed to get Dashboard Cached Result for Attribution query")
		}
	}

	result, err := M.ExecuteAttributionQuery(projectId, requestPayload.Query)
	if err != nil {
		logCtx.WithError(err).Error("query execution failed")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Query execution failed"})
		return
	}

	if dashboardId != 0 && unitId != 0 {
		M.SetCacheResultByDashboardIdAndUnitId(result, projectId, dashboardId, unitId,
			requestPayload.Query.From, requestPayload.Query.To)
	}
	c.JSON(http.StatusOK, gin.H{"result": result, "cache": false, "refreshed_at": U.TimeNowIn(U.TimeZoneStringIST).Unix()})
}

// decodeAttributionPayload decodes attribution requestPayload for 2 json formats to support old and new
// request formats
func decodeAttributionPayload(r *http.Request, logCtx *log.Entry) (bool, string, AttributionRequestPayload) {

	var err error
	var requestPayload AttributionRequestPayload
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logCtx.WithError(err).Error("query failed due to Error while reading r.Body")
		return true, "Error while reading r.Body", requestPayload
	}
	decoder1 := json.NewDecoder(bytes.NewReader(data))
	decoder1.DisallowUnknownFields()
	if err = decoder1.Decode(&requestPayload); err == nil {
		return false, "", requestPayload
	}

	decoder2 := json.NewDecoder(bytes.NewReader(data))
	decoder2.DisallowUnknownFields()
	var requestPayloadUnit M.AttributionQueryUnit
	if err = decoder2.Decode(&requestPayloadUnit); err == nil {
		requestPayload.Query = requestPayloadUnit.Query
		return false, "", requestPayload
	}
	logCtx.WithError(err).Error("query failed as JSON decode failed")
	return true, "Query failed. Json decode failed", requestPayload
}
