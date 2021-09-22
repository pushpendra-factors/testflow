package handler

import (
	"bytes"
	"encoding/json"
	C "factors/config"
	H "factors/handler/helpers"
	V1 "factors/handler/v1"
	mid "factors/middleware"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type AttributionRequestPayload struct {
	Query *model.AttributionQuery `json:"query"`
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
func AttributionHandler(c *gin.Context) (interface{}, int, string, string, bool) {

	r := c.Request
	reqId := U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID)
	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	logCtx := log.WithFields(log.Fields{
		"reqId": reqId, "project_id": projectId,
	})
	if projectId == 0 {
		logCtx.Error("Query failed. Invalid project.")
		return nil, http.StatusUnauthorized, V1.INVALID_PROJECT, "Query failed. Invalid project.", true
	}

	var err error
	var requestPayload AttributionRequestPayload
	var dashboardId uint64
	var unitId uint64
	var timezoneString U.TimeZoneString
	var statusCode int

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
		if err != nil || dashboardId == 0 {
			logCtx.WithError(err).Error("Query failed. Invalid DashboardID.")
			return nil, http.StatusBadRequest, V1.INVALID_INPUT, "Query failed. Invalid DashboardID.", true
		}
		unitId, err = strconv.ParseUint(unitIdParam, 10, 64)
		if err != nil || unitId == 0 {
			logCtx.WithError(err).Error("Query failed. Invalid DashboardUnitID.")
			return nil, http.StatusBadRequest, V1.INVALID_INPUT, "Query failed. Invalid DashboardUnitID.", true
		}
	}

	hasFailed, errMsg, requestPayload := decodeAttributionPayload(r, logCtx)
	if hasFailed {
		logCtx.Error("Query failed. Json decode failed." + errMsg)
		return nil, http.StatusBadRequest, V1.INVALID_INPUT, "Query failed. Json decode failed." + errMsg, true
	}

	if requestPayload.Query.Timezone != "" {
		_, errCode := time.LoadLocation(string(requestPayload.Query.Timezone))
		if errCode != nil {
			return nil, http.StatusBadRequest, V1.INVALID_INPUT, "Query failed. Invalid Timezone provided.", true
		}
		timezoneString = U.TimeZoneString(requestPayload.Query.Timezone)
	} else {
		timezoneString, statusCode = store.GetStore().GetTimezoneForProject(projectId)
		if statusCode != http.StatusFound {
			logCtx.WithError(err).Error("Query failed. Failed to get Timezone.")
			return nil, http.StatusBadRequest, V1.INVALID_INPUT, "Query failed. Failed to get Timezone.", true
		}
		// logCtx.WithError(err).Error("Query failed. Invalid Timezone.")
	}

	// If refresh is passed, refresh only is Query.From is of today's beginning.
	if isDashboardQueryRequest && !H.ShouldAllowHardRefresh(requestPayload.Query.From, requestPayload.Query.To, timezoneString, hardRefresh) {

		effectiveFrom, effectiveTo := model.GetEffectiveTimeRangeForDashboardUnitAttributionQuery(requestPayload.Query.From, requestPayload.Query.To)
		if effectiveFrom == 0 || effectiveTo == 0 {
			return nil, http.StatusBadRequest, V1.INVALID_INPUT, "Query time range is not valid for attribution.", true
		}
		shouldReturn, resCode, resMsg := H.GetResponseIfCachedDashboardQuery(reqId, projectId, dashboardId, unitId, effectiveFrom, effectiveTo, timezoneString)
		if shouldReturn {
			if resCode == http.StatusOK {
				return resMsg, resCode, "", "", false
			}
		}
	}

	var cacheResult model.QueryResult
	attributionQueryUnitPayload := model.AttributionQueryUnit{
		Class: model.QueryClassAttribution,
		Query: requestPayload.Query,
	}
	attributionQueryUnitPayload.SetTimeZone(timezoneString)
	err = attributionQueryUnitPayload.TransformDateTypeFilters()
	if err != nil {
		return nil, http.StatusBadRequest, V1.INVALID_INPUT, err.Error(), true
	}

	shouldReturn, resCode, resMsg := H.GetResponseIfCachedQuery(c, projectId, &attributionQueryUnitPayload, cacheResult, isDashboardQueryRequest, reqId)
	if shouldReturn {
		if resCode == http.StatusOK {
			return resMsg, resCode, "", "", false
		}
		logCtx.WithError(err).Error("Query failed. Error Processing/Fetching data from Query cache")
		return nil, resCode, V1.PROCESSING_FAILED, "Error Processing/Fetching data from Query cache", true
	}

	if isDashboardQueryRequest && C.DisableDashboardQueryDBExecution() {
		logCtx.WithField("request_payload", requestPayload).Warn("Skip hitting db for queries from dashboard, if not found on cache.")
		return nil, resCode, V1.PROCESSING_FAILED, "Not found in cache. Execution suspended temporarily.", true
	}

	// If not found, set a placeholder for the query hash key that it has been running to avoid running again.
	model.SetQueryCachePlaceholder(projectId, &attributionQueryUnitPayload)
	H.SleepIfHeaderSet(c)

	result, err := store.GetStore().ExecuteAttributionQuery(projectId, requestPayload.Query)
	if err != nil {
		model.DeleteQueryCacheKey(projectId, &attributionQueryUnitPayload)
		logCtx.WithError(err).Error("Failed to process query from DB")
		return nil, http.StatusInternalServerError, V1.PROCESSING_FAILED, err.Error(), true
	}
	model.SetQueryCacheResult(projectId, &attributionQueryUnitPayload, result)

	if isDashboardQueryRequest {
		model.SetCacheResultByDashboardIdAndUnitId(result, projectId, dashboardId, unitId,
			requestPayload.Query.From, requestPayload.Query.To, timezoneString)
		return H.DashboardQueryResponsePayload{Result: result, Cache: false, RefreshedAt: U.TimeNowIn(U.TimeZoneStringIST).Unix()}, http.StatusOK, "", "", false
	}
	return result, http.StatusOK, "", "", false
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
	var requestPayloadUnit model.AttributionQueryUnit
	if err = decoder2.Decode(&requestPayloadUnit); err == nil {
		requestPayload.Query = requestPayloadUnit.Query
		return false, "", requestPayload
	}
	logCtx.WithError(err).Error("query failed as JSON decode failed")
	return true, "Query failed. Json decode failed", requestPayload
}
