package v1

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
// @Success 200 {string} json "{"result": model.ChannelConfigResult, "refreshed_at": timestamp}"
// @Router /{project_id}/v1/channels/config [get]
func GetChannelConfigHandler(c *gin.Context) {
	reqID := U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID)
	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
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

	result, httpStatus := M.GetChannelConfig(channel, reqID)

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
// @Success 200 {string} json "{"result": model.ChannelFilterValues, "refreshed_at": timestamp}"
// @Router /{project_id}/v1/channels/filter_values [get]
func GetChannelFilterValuesHandler(c *gin.Context) {
	reqID := U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID)
	projectID := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
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

	channelFilterValues, errCode := M.GetChannelFilterValuesV1(projectID, channel, filterObject, filterProperty, reqID)
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
// @Param query body model.ChannelQueryGroupV1 true "Query payload"
// @Success 200 {string} json "{result:M.ChannelResultGroupV1"
// @Router /{project_id}/v1/channels/query [post]
func ExecuteChannelQueryHandler(c *gin.Context) {
	r := c.Request
	reqID := U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID)
	logCtx := log.WithField("req_id", reqID)

	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized,
			gin.H{"error": "Channel query failed. Invalid project."})
		return
	}
	logCtx = logCtx.WithField("project_id", projectId).WithField("req_id", reqID)
	var queryPayload M.ChannelGroupQueryV1
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&queryPayload); err != nil {
		logCtx.WithError(err).Error("Channel query failed. Json decode failed.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "Channel executeQuery failed. Json decode failed."})
		return
	}
	logCtx.Info("query:", queryPayload)

	var commonQueryFrom int64
	var commonQueryTo int64
	if len(queryPayload.Queries) == 0 {
		logCtx.Error("Query failed. Empty query group.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Query failed. Empty query group."})
		return
	} else {
		// all group queries are run for same time duration, used in dashboard unit caching
		commonQueryFrom = queryPayload.Queries[0].From
		commonQueryTo = queryPayload.Queries[0].To
	}

	var dashboardId uint64
	var unitId uint64
	var err error
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

	// If refresh is passed, refresh only is Query.From is of todays beginning.
	if (dashboardIdParam != "" || unitIdParam != "") && !isHardRefreshForToday(commonQueryFrom, hardRefresh) {
		cacheResult, errCode, errMsg := M.GetCacheResultByDashboardIdAndUnitId(projectId, dashboardId, unitId, commonQueryFrom, commonQueryTo)
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
			}).WithError(errMsg).Error("Failed to get GetCacheResultByDashboardIdAndUnitId from cache.")
		}
	}

	// Run Channel Query
	queryResult, errCode := M.RunChannelGroupQuery(projectId, queryPayload.Queries, reqID)
	if errCode != http.StatusOK {
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			gin.H{"error": "Channel query failed. Execution failure."})
		return
	}

	// if it is a dashboard query, cache it
	if dashboardId != 0 && unitId != 0 {
		M.SetCacheResultByDashboardIdAndUnitId(queryResult, projectId, dashboardId, unitId, commonQueryFrom, commonQueryTo)
		c.JSON(http.StatusOK, gin.H{"result": queryResult, "cache": false, "refreshed_at": U.TimeNowIn(U.TimeZoneStringIST).Unix()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": queryResult})
}

func isHardRefreshForToday(from int64, hardRefresh bool) bool {
	return from == U.GetBeginningOfDayTimestampZ(U.TimeNowUnix(), U.TimeZoneStringIST) && hardRefresh
}
