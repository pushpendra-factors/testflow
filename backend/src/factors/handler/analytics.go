package handler

import (
	"encoding/json"
	mid "factors/middleware"
	M "factors/model"
	U "factors/util"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strconv"
)

type QueryRequestPayload struct {
	Query M.Query `json:"query"`
}

/*
Test Command

Unique User:
curl -i -H 'cookie: factors-sid=<COOKIE>' -H "Content-Type: application/json" -i -X POST http://factors-dev.com:8080/projects/2/query -d '{"query":{"type":"unique_users","eventsCondition":"all","from":1393632004,"to":1396310325,"eventsWithProperties":[{"name":"View Project","properties":[{"entity":"event","property":"category","operator":"equals","type":"categorical","value":"Sports"},{"entity":"user","property":"gender","operator":"equals","type":"categorical","value":"M"}]},{"name":"Fund Project","properties":[{"entity":"event","property":"category","operator":"equals","type":"categorical","value":"Sports"},{"entity":"user","property":"gender","operator":"equals","type":"categorical","value":"M"}]}],"groupByProperties":[{"property":"$region","entity":"user","index":1}]}}'

Events Occurence:
curl -i -H 'cookie: factors-sid=<COOKIE>' -H "Content-Type: application/json" -i -X POST http://factors-dev.com:8080/projects/2/query -d '{"query":{"type":"events_occurrence","eventsCondition":"any","from":1393632004,"to":1396310325,"eventsWithProperties":[{"name":"View Project","properties":[{"entity":"user","property":"gender","operator":"equals","type":"categorical","value":"M"}]},{"name":"Fund Project","properties":[{"entity":"user","property":"gender","operator":"equals","type":"categorical","value":"M"}]}],"groupByProperties":[{"property":"$region","entity":"user","index":0},{"property":"category","entity":"event","index":1}]}}'
*/

func EventsQueryHandler(c *gin.Context) {

	logCtx := log.WithFields(log.Fields{
		"reqId": U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID),
	})

	r := c.Request

	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Query failed. Invalid project."})
		return
	}

	var requestPayload M.QueryGroup
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&requestPayload); err != nil {
		logCtx.WithError(err).Error("Query failed. Json decode failed.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Query failed. Json decode failed."})
		return
	}

	var commonQueryFrom int64
	var commonQueryTo int64
	if len(requestPayload.Queries) == 0 {
		logCtx.Error("Query failed. Empty query group.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Query failed. Empty query group."})
		return
	} else {
		// all group queries are run for same time duration, used in dashboard unit caching
		commonQueryFrom = requestPayload.Queries[0].From
		commonQueryTo = requestPayload.Queries[0].To
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

	resultGroup, errCode := M.RunEventsGroupQuery(requestPayload.Queries, projectId)
	if errCode != http.StatusOK {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "group query failed to run"})
		return
	}

	// if it is a dashboard query, cache it
	if dashboardId != 0 && unitId != 0 {
		M.SetCacheResultByDashboardIdAndUnitId(resultGroup, projectId, dashboardId, unitId, commonQueryFrom, commonQueryTo)
		c.JSON(http.StatusOK, gin.H{"result": resultGroup, "cache": false, "refreshed_at": U.TimeNowIn(U.TimeZoneStringIST).Unix()})
		return
	}
	c.JSON(http.StatusOK, resultGroup)
}

func QueryHandler(c *gin.Context) {

	logCtx := log.WithFields(log.Fields{
		"reqId": U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID),
	})

	r := c.Request

	projectId := U.GetScopeByKeyAsUint64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Query failed. Invalid project."})
		return
	}

	var requestPayload QueryRequestPayload
	var dashboardId uint64
	var unitId uint64
	var err error
	var result *M.QueryResult
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
			}).WithError(errMsg).Error("Failed to get GetCacheResultByDashboardIdAndUnitId from cache.")
		}
	}

	for index, _ := range requestPayload.Query.GroupByProperties {
		if requestPayload.Query.GroupByProperties[index].Type == U.PropertyTypeDateTime &&
			requestPayload.Query.GroupByProperties[index].Granularity == "" {
			requestPayload.Query.GroupByProperties[index].Granularity = U.DateTimeBreakdownDailyGranularity
		}
	}
	result, errCode, errMsg := M.Analyze(projectId, requestPayload.Query)
	if errCode != http.StatusOK {
		c.AbortWithStatusJSON(errCode, gin.H{"error": errMsg})
		return
	}

	if dashboardId != 0 && unitId != 0 {
		M.SetCacheResultByDashboardIdAndUnitId(result, projectId, dashboardId, unitId, requestPayload.Query.From, requestPayload.Query.To)
		c.JSON(http.StatusOK, gin.H{"result": result, "cache": false, "refreshed_at": U.TimeNowIn(U.TimeZoneStringIST).Unix()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func isHardRefreshForToday(from int64, hardRefresh bool) bool {
	return from == U.GetBeginningOfDayTimestampZ(U.TimeNowUnix(), U.TimeZoneStringIST) && hardRefresh
}
