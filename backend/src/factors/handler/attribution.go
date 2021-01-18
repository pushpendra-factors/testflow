package handler

import (
	"bytes"
	"encoding/json"
	H "factors/handler/helpers"
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

	isDashboardQueryRequest := dashboardIdParam != "" && unitIdParam != ""
	if isDashboardQueryRequest {
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
	if isDashboardQueryRequest && !H.IsHardRefreshForToday(requestPayload.Query.From, hardRefresh) {
		shouldReturn, resCode, resMsg := H.GetResponseIfCachedDashboardQuery(
			projectId, dashboardId, unitId, requestPayload.Query.From, requestPayload.Query.To)
		if shouldReturn {
			c.AbortWithStatusJSON(resCode, resMsg)
			return
		}
	}

	var cacheResult M.QueryResult
	attributionQueryUnitPayload := M.AttributionQueryUnit{
		Class: M.QueryClassAttribution,
		Query: requestPayload.Query,
	}
	shouldReturn, resCode, resMsg := H.GetResponseIfCachedQuery(c, projectId, &attributionQueryUnitPayload, cacheResult, isDashboardQueryRequest)
	if shouldReturn {
		c.AbortWithStatusJSON(resCode, resMsg)
		return
	}

	// If not found, set a placeholder for the query hash key that it has been running to avoid running again.
	M.SetQueryCachePlaceholder(projectId, &attributionQueryUnitPayload)
	H.SleepIfHeaderSet(c)

	result, err := M.ExecuteAttributionQuery(projectId, requestPayload.Query)
	if err != nil {
		logCtx.WithError(err).Error("query execution failed")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Query execution failed"})
		M.DeleteQueryCacheKey(projectId, &attributionQueryUnitPayload)
		return
	}
	M.SetQueryCacheResult(projectId, &attributionQueryUnitPayload, result)

	if isDashboardQueryRequest {
		M.SetCacheResultByDashboardIdAndUnitId(result, projectId, dashboardId, unitId,
			requestPayload.Query.From, requestPayload.Query.To)
	}
	c.JSON(http.StatusOK, H.DashboardQueryResponsePayload{Result: result, Cache: false, RefreshedAt: U.TimeNowIn(U.TimeZoneStringIST).Unix()})
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
