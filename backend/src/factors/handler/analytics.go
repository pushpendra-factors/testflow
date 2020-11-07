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

type QueryRequestPayload struct {
	Query M.Query `json:"query"`
}

type QueryGroup struct {
	Queries []M.Query `json:"query_group"`
}

/*
Test Command

Unique User:
curl -i -H 'cookie: factors-sid=<COOKIE>' -H "Content-Type: application/json" -i -X POST http://factors-dev.com:8080/projects/2/query -d '{"query":{"type":"unique_users","eventsCondition":"all","from":1393632004,"to":1396310325,"eventsWithProperties":[{"name":"View Project","properties":[{"entity":"event","property":"category","operator":"equals","type":"categorical","value":"Sports"},{"entity":"user","property":"gender","operator":"equals","type":"categorical","value":"M"}]},{"name":"Fund Project","properties":[{"entity":"event","property":"category","operator":"equals","type":"categorical","value":"Sports"},{"entity":"user","property":"gender","operator":"equals","type":"categorical","value":"M"}]}],"groupByProperties":[{"property":"$region","entity":"user","index":1}]}}'

Events Occurence:
curl -i -H 'cookie: factors-sid=<COOKIE>' -H "Content-Type: application/json" -i -X POST http://factors-dev.com:8080/projects/2/query -d '{"query":{"type":"events_occurrence","eventsCondition":"any","from":1393632004,"to":1396310325,"eventsWithProperties":[{"name":"View Project","properties":[{"entity":"user","property":"gender","operator":"equals","type":"categorical","value":"M"}]},{"name":"Fund Project","properties":[{"entity":"user","property":"gender","operator":"equals","type":"categorical","value":"M"}]}],"groupByProperties":[{"property":"$region","entity":"user","index":0},{"property":"category","entity":"event","index":1}]}}'
*/

// EventsQueryHandler godoc
// @Summary To run a particular query group from core query or dashboards.
// @Tags V1Api
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param dashboard_id query integer false "Dashboard ID"
// @Param dashboard_unit_id query integer false "Dashboard Unit ID"
// @Param query_group body handler.QueryGroup true "Query group"
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

	var requestPayload QueryGroup
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&requestPayload); err != nil {
		logCtx.WithError(err).Error("Query failed. Json decode failed.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Query failed. Json decode failed."})
		return
	}
	// TODO (Anil) Add dashboard caching layer by query/query_group?
	resultGroup := M.RunEventsGroupQuery(requestPayload.Queries, projectId)
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
