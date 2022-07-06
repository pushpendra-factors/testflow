package v1

import (
	"encoding/json"
	H "factors/handler/helpers"
	mid "factors/middleware"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

const ChannelQueryParamKey = "channel"
const ChannelFilterObjectParamKey = "filter_object"
const ChannelFilterPropertyParamKey = "filter_property"

// GetChannelConfigHandler godoc
// @Summary To get config for the required channel.
// @Tags ChannelQuery,CoreQuery
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param channel query string true "Channel"
// @Success 200 {string} json "{"result": model.ChannelConfigResult"
// @Router /{project_id}/v1/channels/config [get]
func GetChannelConfigHandler(c *gin.Context) {
	reqID := U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID)
	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized,
			gin.H{"error": "Channel query failed. Invalid project."})
		return
	}
	channel := c.Query(ChannelQueryParamKey)
	if channel == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Missing params channel."})
		return
	}

	result, httpStatus := store.GetStore().GetChannelConfig(projectId, channel, reqID)

	c.JSON(httpStatus, gin.H{"result": result})
}

// GetChannelFilterValuesHandler godoc
// @Summary To filter on values for channel query.
// @Tags ChannelQuery,CoreQuery
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param channel query string true "Channel"
// @Param filter_object query string true "campaign"
// @Param filter_property query string true "name"
// @Success 200 {string} json "{"result": model.ChannelFilterValues}"
// @Router /{project_id}/v1/channels/filter_values [get]
func GetChannelFilterValuesHandler(c *gin.Context) {
	reqID := U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID)
	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized,
			gin.H{"error": "Channel query failed. Invalid project."})
		return
	}

	channel := c.Query(ChannelQueryParamKey)
	filterObject := c.Query(ChannelFilterObjectParamKey)
	filterProperty := c.Query(ChannelFilterPropertyParamKey)
	if channel == "" || filterObject == "" || filterProperty == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Missing params channel and filter."})
		return
	}

	channelFilterValues, errCode := store.GetStore().GetChannelFilterValuesV1(projectID, channel, filterObject, filterProperty, reqID)
	if errCode != http.StatusFound {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Failed to get filter values for channel."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": channelFilterValues})
}

// ChannelQueryHandler godoc
// @Summary To run a channel query.
// @Tags ChannelQuery,CoreQuery
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param dashboard_id query integer false "Dashboard ID"
// @Param dashboard_unit_id query integer false "Dashboard Unit ID"
// @Param query body model.ChannelGroupQueryV1 true "Query payload"
// @Success 200 {string} json "{result:store.GetStore().ChannelResultGroupV1}"
// @Router /{project_id}/v1/channels/query [post]
func ExecuteChannelQueryHandler(c *gin.Context) (interface{}, int, string, string, bool) {
	r := c.Request
	reqID := U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID)
	logCtx := log.WithField("reqId", reqID)

	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		logCtx.Error("Query failed. Invalid project.")
		return nil, http.StatusUnauthorized, INVALID_PROJECT, "Query failed. Invalid project.", true
	}
	logCtx = logCtx.WithField("project_id", projectId).WithField("reqId", reqID)

	var queryPayload model.ChannelGroupQueryV1
	queryIdString := c.Query("query_id")
	if queryIdString == "" {
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()

		if err := decoder.Decode(&queryPayload); err != nil {
			logCtx.WithError(err).Error("Query failed. Json decode failed.")
			return nil, http.StatusBadRequest, INVALID_INPUT, "Query failed. Json decode failed.", true
		}
	} else {
		_, query, err := store.GetStore().GetQueryAndClassFromQueryIdString(queryIdString, projectId)
		if err != "" {
			logCtx.Error(fmt.Sprintf("Query from queryIdString failed - %v", err))
			return nil, http.StatusBadRequest, INVALID_INPUT, "Query failed. Json decode failed.", true
		}
		U.DecodePostgresJsonbToStructType(&query.Query, &queryPayload)
	}

	var commonQueryFrom int64
	var commonQueryTo int64
	if len(queryPayload.Queries) == 0 {
		logCtx.Error("Query failed. Empty query group.")
		return nil, http.StatusBadRequest, INVALID_INPUT, "Query failed. Empty query group.", true
	} else {
		// all group queries are run for same time duration, used in dashboard unit caching
		commonQueryFrom = queryPayload.Queries[0].From
		commonQueryTo = queryPayload.Queries[0].To
	}

	var dashboardId int64
	var unitId int64
	var err error
	hardRefresh := false
	dashboardIdParam := c.Query("dashboard_id")
	unitIdParam := c.Query("dashboard_unit_id")
	refreshParam := c.Query("refresh")
	var timezoneString U.TimeZoneString
	var statusCode int

	if refreshParam != "" {
		hardRefresh, _ = strconv.ParseBool(refreshParam)
	}

	isDashboardQueryRequest := dashboardIdParam != "" && unitIdParam != ""
	if isDashboardQueryRequest {
		dashboardId, err = strconv.ParseInt(dashboardIdParam, 10, 64)
		if err != nil || dashboardId == 0 {
			logCtx.WithError(err).Error("Query failed. Invalid DashboardID.")
			return nil, http.StatusBadRequest, INVALID_INPUT, "Query failed. Invalid DashboardID.", true
		}
		unitId, err = strconv.ParseInt(unitIdParam, 10, 64)
		if err != nil || unitId == 0 {
			logCtx.WithError(err).Error("Query failed. Invalid DashboardUnitID.")
			return nil, http.StatusBadRequest, INVALID_INPUT, "Query failed. Invalid DashboardUnitID.", true
		}
	}

	if queryPayload.Queries[0].Timezone != "" {
		_, errCode := time.LoadLocation(string(queryPayload.Queries[0].Timezone))
		if errCode != nil {
			return nil, http.StatusBadRequest, INVALID_INPUT, "Query failed. Invalid Timezone provided.", true
		}

		timezoneString = U.TimeZoneString(queryPayload.Queries[0].Timezone)
	} else {
		timezoneString, statusCode = store.GetStore().GetTimezoneForProject(projectId)
		if statusCode != http.StatusFound {
			logCtx.Error("Query failed. Failed to get Timezone.")
			return nil, http.StatusBadRequest, INVALID_INPUT, "Query failed. Failed to get Timezone.", true
		}
		// logCtx.WithError(err).Error("Query failed. Invalid Timezone.")
	}

	// Tracking dashboard query request.
	if isDashboardQueryRequest {
		model.SetDashboardCacheAnalytics(projectId, dashboardId, unitId, commonQueryFrom, commonQueryTo, timezoneString)
	}

	// If refresh is passed, refresh only is Query.From is of todays beginning.
	if isDashboardQueryRequest && !H.ShouldAllowHardRefresh(commonQueryFrom, commonQueryTo, timezoneString, hardRefresh) {
		shouldReturn, resCode, resMsg := H.GetResponseIfCachedDashboardQuery(reqID, projectId, dashboardId, unitId, commonQueryFrom, commonQueryTo, timezoneString)
		if shouldReturn {
			if resCode == http.StatusOK {
				return resMsg, resCode, "", "", false
			}
		}
	}
	queryPayload.SetTimeZone(timezoneString)

	var cacheResult model.ChannelResultGroupV1
	shouldReturn, resCode, resMsg := H.GetResponseIfCachedQuery(c, projectId, &queryPayload, cacheResult, isDashboardQueryRequest, reqID)
	if shouldReturn {
		if resCode == http.StatusOK {
			return resMsg, resCode, "", "", false
		}
		logCtx.Error("Query failed. Error Processing/Fetching data from Query cache")
		return nil, resCode, PROCESSING_FAILED, "Error Processing/Fetching data from Query cache", true
	}

	// If not found, set a placeholder for the query hash key that it has been running to avoid running again.
	model.SetQueryCachePlaceholder(projectId, &queryPayload)
	H.SleepIfHeaderSet(c)

	// Run Channel Query
	queryResult, errCode := store.GetStore().RunChannelGroupQuery(projectId, queryPayload.Queries, reqID)
	if errCode != http.StatusOK {
		model.DeleteQueryCacheKey(projectId, &queryPayload)
		logCtx.Error("Failed to process query from DB")
		if errCode == http.StatusPartialContent {
			return queryResult, errCode, PROCESSING_FAILED, "Failed to process query from DB", true
		}
		return nil, errCode, PROCESSING_FAILED, "Failed to process query from DB", true
	}
	model.SetQueryCacheResult(projectId, &queryPayload, queryResult)

	// if it is a dashboard query, cache it
	if isDashboardQueryRequest {
		model.SetCacheResultByDashboardIdAndUnitId(queryResult, projectId, dashboardId, unitId, commonQueryFrom, commonQueryTo, timezoneString)
		return H.DashboardQueryResponsePayload{Result: queryResult, Cache: false, RefreshedAt: U.TimeNowIn(U.TimeZoneStringIST).Unix(), TimeZone: string(timezoneString)}, http.StatusOK, "", "", false
	}
	return gin.H{"result": queryResult, "query": queryPayload}, http.StatusOK, "", "", false
}
