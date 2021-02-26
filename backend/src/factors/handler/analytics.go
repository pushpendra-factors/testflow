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

type QueryRequestPayload struct {
	Query model.Query `json:"query"`
}

/*
Test Command

Unique User:
curl -i -H 'cookie: factors-sid=<COOKIE>' -H "Content-Type: application/json" -i -X POST http://factors-dev.com:8080/projects/2/query -d '{"query":{"type":"unique_users","eventsCondition":"all","from":1393632004,"to":1396310325,"eventsWithProperties":[{"name":"View Project","properties":[{"entity":"event","property":"category","operator":"equals","type":"categorical","value":"Sports"},{"entity":"user","property":"gender","operator":"equals","type":"categorical","value":"M"}]},{"name":"Fund Project","properties":[{"entity":"event","property":"category","operator":"equals","type":"categorical","value":"Sports"},{"entity":"user","property":"gender","operator":"equals","type":"categorical","value":"M"}]}],"groupByProperties":[{"property":"$region","entity":"user","index":1}]}}'

Events Occurence:
curl -i -H 'cookie: factors-sid=<COOKIE>' -H "Content-Type: application/json" -i -X POST http://factors-dev.com:8080/projects/2/query -d '{"query":{"type":"events_occurrence","eventsCondition":"any","from":1393632004,"to":1396310325,"eventsWithProperties":[{"name":"View Project","properties":[{"entity":"user","property":"gender","operator":"equals","type":"categorical","value":"M"}]},{"name":"Fund Project","properties":[{"entity":"user","property":"gender","operator":"equals","type":"categorical","value":"M"}]}],"groupByProperties":[{"property":"$region","entity":"user","index":0},{"property":"category","entity":"event","index":1}]}}'
*/

// EventsQueryHandler godoc
// @Summary To run events core query as a query group for user and event count.
// @Tags V1Api
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param dashboard_id query integer false "Dashboard ID"
// @Param dashboard_unit_id query integer false "Dashboard Unit ID"
// @Param query_group body model.QueryGroup true "Query group"
// @Success 200 {string} json "{"result": model.QueryResult, "cache": false, "refreshed_at": timestamp}"
// @Router /{project_id}/v1/query [post]
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

	var requestPayload model.QueryGroup
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

	// If refresh is passed, refresh only is Query.From is of todays beginning.
	if isDashboardQueryRequest && !H.ShouldAllowHardRefresh(commonQueryFrom, commonQueryTo, hardRefresh) {
		shouldReturn, resCode, resMsg := H.GetResponseIfCachedDashboardQuery(projectId, dashboardId, unitId, commonQueryFrom, commonQueryTo)
		if shouldReturn {
			c.AbortWithStatusJSON(resCode, resMsg)
			return
		}
	}

	var cacheResult model.ResultGroup
	shouldReturn, resCode, resMsg := H.GetResponseIfCachedQuery(c, projectId, &requestPayload, cacheResult, isDashboardQueryRequest)
	if shouldReturn {
		c.AbortWithStatusJSON(resCode, resMsg)
		return
	}

	// If not found, set a placeholder for the query hash key that it has been running to avoid running again.
	model.SetQueryCachePlaceholder(projectId, &requestPayload)
	H.SleepIfHeaderSet(c)

	resultGroup, errCode := store.GetStore().RunEventsGroupQuery(requestPayload.Queries, projectId)
	if errCode != http.StatusOK {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "group query failed to run"})
		model.DeleteQueryCacheKey(projectId, &requestPayload)
		return
	}
	model.SetQueryCacheResult(projectId, &requestPayload, resultGroup)

	// if it is a dashboard query, cache it
	if isDashboardQueryRequest {
		model.SetCacheResultByDashboardIdAndUnitId(resultGroup, projectId, dashboardId, unitId, commonQueryFrom, commonQueryTo)
		c.JSON(http.StatusOK, H.DashboardQueryResponsePayload{
			Result: resultGroup, Cache: false, RefreshedAt: U.TimeNowIn(U.TimeZoneStringIST).Unix()})
		return
	}
	c.JSON(http.StatusOK, resultGroup)
}

// QueryHandler godoc
// @Summary To run a particular query from core query or dashboards.
// @Tags CoreQuery
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param dashboard_id query integer false "Dashboard ID"
// @Param dashboard_unit_id query integer false "Dashboard Unit ID"
// @Param query body handler.QueryRequestPayload true "Query payload"
// @Success 200 {string} json "{"result": model.QueryResult, "cache": false, "refreshed_at": timestamp}"
// @Router /{project_id}/query [post]
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
	var result *model.QueryResult
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
	if err := decoder.Decode(&requestPayload); err != nil {
		logCtx.WithError(err).Error("Query failed. Json decode failed.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Query failed. Json decode failed."})
		return
	}

	// If refresh is passed, refresh only is Query.From is of todays beginning.
	if isDashboardQueryRequest && !H.ShouldAllowHardRefresh(requestPayload.Query.From, requestPayload.Query.To, hardRefresh) {
		shouldReturn, resCode, resMsg := H.GetResponseIfCachedDashboardQuery(
			projectId, dashboardId, unitId, requestPayload.Query.From, requestPayload.Query.To)
		if shouldReturn {
			c.AbortWithStatusJSON(resCode, resMsg)
			return
		}
	}

	for index := range requestPayload.Query.GroupByProperties {
		if requestPayload.Query.GroupByProperties[index].Type == U.PropertyTypeDateTime &&
			requestPayload.Query.GroupByProperties[index].Granularity == "" {
			requestPayload.Query.GroupByProperties[index].Granularity = U.DateTimeBreakdownDailyGranularity
		}
	}

	var cacheResult model.QueryResult
	shouldReturn, resCode, resMsg := H.GetResponseIfCachedQuery(c, projectId, &requestPayload.Query, cacheResult, isDashboardQueryRequest)
	if shouldReturn {
		c.AbortWithStatusJSON(resCode, resMsg)
		return
	}

	// If not found, set a placeholder for the query hash key that it has been running to avoid running again.
	model.SetQueryCachePlaceholder(projectId, &requestPayload.Query)
	H.SleepIfHeaderSet(c)

	result, errCode, errMsg := store.GetStore().Analyze(projectId, requestPayload.Query)
	if errCode != http.StatusOK {
		c.AbortWithStatusJSON(errCode, gin.H{"error": errMsg})
		model.DeleteQueryCacheKey(projectId, &requestPayload.Query)
		return
	}
	model.SetQueryCacheResult(projectId, &requestPayload.Query, result)

	if isDashboardQueryRequest {
		model.SetCacheResultByDashboardIdAndUnitId(result, projectId, dashboardId, unitId, requestPayload.Query.From, requestPayload.Query.To)
		c.JSON(http.StatusOK, H.DashboardQueryResponsePayload{Result: result, Cache: false, RefreshedAt: U.TimeNowIn(U.TimeZoneStringIST).Unix()})
		return
	}
	c.JSON(http.StatusOK, result)
}
