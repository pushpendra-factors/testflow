package handler

import (
	"encoding/json"
	H "factors/handler/helpers"
	mid "factors/middleware"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// ChannelQueryHandler godoc
// @Summary To run a channel query.
// @Tags CoreQuery
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param dashboard_id query integer false "Dashboard ID"
// @Param dashboard_unit_id query integer false "Dashboard Unit ID"
// @Param query body model.ChannelQuery true "Query payload"
// @Success 200 {string} json "{"result": model.ChannelQueryResult, "cache": false, "refreshed_at": timestamp}"
// @Router /{project_id}/channels/query [post]
func ChannelQueryHandler(c *gin.Context) {
	r := c.Request

	logCtx := log.WithFields(log.Fields{
		"reqId": U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID),
	})

	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized,
			gin.H{"error": "Channel query failed. Invalid project."})
		return
	}

	logCtx = log.WithField("project_id", projectId)

	var err error
	var queryPayload model.ChannelQuery
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

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&queryPayload); err != nil {
		logCtx.WithError(err).Error("Channel query failed. Json decode failed.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "Channel query failed. Json decode failed."})
		return
	}

	// If refresh is passed, refresh only is Query.From is of todays beginning.
	if isDashboardQueryRequest && !H.IsHardRefreshForToday(queryPayload.From, hardRefresh) {
		shouldReturn, resCode, resMsg := H.GetResponseIfCachedDashboardQuery(projectId, dashboardId, unitId, queryPayload.From, queryPayload.To)
		if shouldReturn {
			c.AbortWithStatusJSON(resCode, resMsg)
			return
		}
	}

	var cacheResult model.ChannelQueryResult
	channelQueryUnitPayload := model.ChannelQueryUnit{
		Class: model.QueryClassChannel,
		Query: &queryPayload,
	}
	shouldReturn, resCode, resMsg := H.GetResponseIfCachedQuery(c, projectId, &channelQueryUnitPayload, cacheResult, isDashboardQueryRequest)
	if shouldReturn {
		c.AbortWithStatusJSON(resCode, resMsg)
		return
	}

	// If not found, set a placeholder for the query hash key that it has been running to avoid running again.
	model.SetQueryCachePlaceholder(projectId, &channelQueryUnitPayload)
	H.SleepIfHeaderSet(c)

	queryResult, errCode := store.GetStore().ExecuteChannelQuery(projectId, &queryPayload)
	if errCode != http.StatusOK {
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			gin.H{"error": "Channel query failed. Execution failure."})
		model.DeleteQueryCacheKey(projectId, &channelQueryUnitPayload)
		return
	}
	model.SetQueryCacheResult(projectId, &channelQueryUnitPayload, queryResult)

	if isDashboardQueryRequest {
		model.SetCacheResultByDashboardIdAndUnitId(queryResult, projectId, dashboardId, unitId, queryPayload.From, queryPayload.To)
		c.JSON(http.StatusOK, H.DashboardQueryResponsePayload{
			Result: queryResult, Cache: false, RefreshedAt: U.TimeNowIn(U.TimeZoneStringIST).Unix()})
		return
	}

	c.JSON(http.StatusOK, queryResult)
}

// GetChannelFilterValuesHandler godoc
// @Summary To filter on values for channel query.
// @Tags CoreQuery
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param channel query string true "Channel"
// @Param filter query string true "Filter"
// @Success 302 {string} json "{"filter_values": []string}"
// @Router /{project_id}/channels/filter_values [get]
func GetChannelFilterValuesHandler(c *gin.Context) {
	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized,
			gin.H{"error": "Channel query failed. Invalid project."})
		return
	}

	channel := c.Query("channel")
	filter := c.Query("filter")
	if channel == "" || filter == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Missing params channel and filter."})
		return
	}

	filterValues, errCode := store.GetStore().GetChannelFilterValues(projectId, channel, filter)
	if errCode != http.StatusFound {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Failed to get filter values for channel."})
		return
	}

	c.JSON(http.StatusFound, gin.H{"filter_values": filterValues})
}
