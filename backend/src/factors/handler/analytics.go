package handler

import (
	"encoding/json"
	C "factors/config"
	H "factors/handler/helpers"
	V1 "factors/handler/v1"
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

type QueryRequestPayload struct {
	Query model.Query `json:"query"`
}

/*
Test Command

Unique User:
curl -i -H 'cookie: factors-sid=<COOKIE>' -H "Content-UnitType: application/json" -i -X POST http://factors-dev.com:8080/projects/2/query -d '{"query":{"type":"unique_users","eventsCondition":"all","from":1393632004,"to":1396310325,"eventsWithProperties":[{"name":"View Project","properties":[{"entity":"event","property":"category","operator":"equals","type":"categorical","value":"Sports"},{"entity":"user","property":"gender","operator":"equals","type":"categorical","value":"M"}]},{"name":"Fund Project","properties":[{"entity":"event","property":"category","operator":"equals","type":"categorical","value":"Sports"},{"entity":"user","property":"gender","operator":"equals","type":"categorical","value":"M"}]}],"groupByProperties":[{"property":"$region","entity":"user","index":1}]}}'

Events Occurence:
curl -i -H 'cookie: factors-sid=<COOKIE>' -H "Content-UnitType: application/json" -i -X POST http://factors-dev.com:8080/projects/2/query -d '{"query":{"type":"events_occurrence","eventsCondition":"any","from":1393632004,"to":1396310325,"eventsWithProperties":[{"name":"View Project","properties":[{"entity":"user","property":"gender","operator":"equals","type":"categorical","value":"M"}]},{"name":"Fund Project","properties":[{"entity":"user","property":"gender","operator":"equals","type":"categorical","value":"M"}]}],"groupByProperties":[{"property":"$region","entity":"user","index":0},{"property":"category","entity":"event","index":1}]}}'
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
func EventsQueryHandler(c *gin.Context) (interface{}, int, string, string, bool) {

	reqId := U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID)
	logCtx := log.WithFields(log.Fields{
		"reqId": reqId,
	})

	r := c.Request

	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		logCtx.Error("Query failed. Invalid project.")
		return nil, http.StatusUnauthorized, V1.INVALID_PROJECT, "Query failed. Invalid project.", true
	}

	logCtx = log.WithFields(log.Fields{
		"reqId":     reqId,
		"projectId": projectId,
	})
	queryIdString := c.Query("query_id")
	var requestPayload model.QueryGroup
	if queryIdString == "" {
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&requestPayload); err != nil {
			logCtx.WithError(err).Error("Query failed. Json decode failed.")
			return nil, http.StatusBadRequest, V1.INVALID_INPUT, "Query failed. Json decode failed.", true
		}
	} else {
		_, query, err := store.GetStore().GetQueryAndClassFromQueryIdString(queryIdString, projectId)
		if err != "" {
			logCtx.Error(fmt.Sprintf("Query from queryIdString failed - %v", err))
			return nil, http.StatusBadRequest, V1.INVALID_INPUT, "Query failed. Json decode failed.", true
		}
		U.DecodePostgresJsonbToStructType(&query.Query, &requestPayload)
	}

	var commonQueryFrom int64
	var commonQueryTo int64
	if len(requestPayload.Queries) == 0 {
		logCtx.Error("Query failed. Empty query group.")
		return nil, http.StatusBadRequest, V1.INVALID_INPUT, "Query failed. Empty query group.", true
	} else {
		// all group queries are run for same time duration, used in dashboard unit caching
		commonQueryFrom = requestPayload.Queries[0].From
		commonQueryTo = requestPayload.Queries[0].To
	}

	var dashboardId int64
	var unitId int64
	var err error
	hardRefresh := false
	preset := ""
	dashboardIdParam := c.Query("dashboard_id")
	unitIdParam := c.Query("dashboard_unit_id")
	refreshParam := c.Query("refresh")
	presetParam := c.Query("preset")
	if U.PresetLookup[presetParam] != "" {
		preset = presetParam
	}
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
			return nil, http.StatusBadRequest, V1.INVALID_INPUT, "Query failed. Invalid DashboardID.", true
		}
		unitId, err = strconv.ParseInt(unitIdParam, 10, 64)
		if err != nil || unitId == 0 {
			logCtx.WithError(err).Error("Query failed. Invalid DashboardUnitID.")
			return nil, http.StatusBadRequest, V1.INVALID_INPUT, "Query failed. Invalid DashboardUnitID.", true
		}
	}

	if requestPayload.Queries[0].Timezone != "" {
		_, errCode := time.LoadLocation(string(requestPayload.Queries[0].Timezone))
		if errCode != nil {
			return nil, http.StatusBadRequest, V1.INVALID_INPUT, "Query failed. Invalid Timezone provided.", true
		}
		timezoneString = U.TimeZoneString(requestPayload.Queries[0].Timezone)
	} else {
		timezoneString, statusCode = store.GetStore().GetTimezoneForProject(projectId)
		if statusCode != http.StatusFound {
			logCtx.WithError(err).Error("Query failed. Failed to get Timezone.")
			return nil, http.StatusBadRequest, V1.INVALID_INPUT, "Query failed. Failed to get Timezone.", true
		}
		// logCtx.WithError(err).Error("Query failed. Invalid Timezone.")
	}
	requestPayload.SetTimeZone(timezoneString)

	// Tracking dashboard query request.
	if isDashboardQueryRequest {
		if preset == "" {
			preset = U.GetPresetNameByFromAndTo(commonQueryFrom, commonQueryTo, timezoneString)
		}
		model.SetDashboardCacheAnalytics(projectId, dashboardId, unitId, commonQueryFrom, commonQueryTo, timezoneString)
	}
	// If refresh is passed, refresh only is Query.From is of todays beginning.
	if !hardRefresh && isDashboardQueryRequest && !H.ShouldAllowHardRefresh(commonQueryFrom, commonQueryTo, timezoneString, hardRefresh) {
		shouldReturn, resCode, resMsg := H.GetResponseIfCachedDashboardQuery(reqId, projectId, dashboardId, unitId, preset, commonQueryFrom, commonQueryTo, timezoneString)

		if shouldReturn {
			if resCode == http.StatusOK {
				return resMsg, resCode, "", "", false
			}
		}
	}

	err = requestPayload.TransformDateTypeFilters()
	if err != nil {
		return nil, http.StatusBadRequest, V1.INVALID_INPUT, err.Error(), true
	}
	var cacheResult model.ResultGroup
	if !hardRefresh {
		shouldReturn, resCode, resMsg := H.GetResponseIfCachedQuery(c, projectId, &requestPayload, cacheResult, isDashboardQueryRequest, reqId, false)
		if shouldReturn {
			if resCode == http.StatusOK {
				return resMsg, resCode, "", "", false
			}
			logCtx.Error("Query failed. Error Processing/Fetching data from Query cache")
			return nil, resCode, V1.PROCESSING_FAILED, "Error Processing/Fetching data from Query cache", true
		}
	}

	// If not found, set a placeholder for the query hash key that it has been running to avoid running again.
	model.SetQueryCachePlaceholder(projectId, &requestPayload)
	H.SleepIfHeaderSet(c)

	enableOptimisedFilterOnEventUserQuery := c.Request.Header.Get(H.HeaderUserFilterOptForEventsAndUsers) == "true" ||
		C.EnableOptimisedFilterOnEventUserQuery()

	resultGroup, errCode := store.GetStore().RunEventsGroupQuery(requestPayload.Queries, projectId, enableOptimisedFilterOnEventUserQuery)
	if errCode != http.StatusOK {
		model.DeleteQueryCacheKey(projectId, &requestPayload)
		logCtx.Error("Query failed. Failed to process query from DB")
		if errCode == http.StatusPartialContent {
			return resultGroup, errCode, V1.PROCESSING_FAILED, "Failed to process query from DB", true
		}
		return nil, errCode, V1.PROCESSING_FAILED, "Failed to process query from DB", true
	}

	meta := H.CacheMeta{
		Timezone:       string(timezoneString),
		From:           commonQueryFrom,
		To:             commonQueryFrom,
		RefreshedAt:    U.TimeNowIn(U.TimeZoneStringIST).Unix(),
		LastComputedAt: U.TimeNowIn(U.TimeZoneStringIST).Unix(),
		Preset:         preset,
	}
	for i, _ := range resultGroup.Results {
		resultGroup.Results[i].CacheMeta = meta
	}
	resultGroup.CacheMeta = meta
	model.SetQueryCacheResult(projectId, &requestPayload, resultGroup)
	// if it is a dashboard query, cache it
	if isDashboardQueryRequest {

		model.SetCacheResultByDashboardIdAndUnitId(resultGroup, projectId, dashboardId, unitId, preset, commonQueryFrom, commonQueryFrom, timezoneString, meta)
		return H.DashboardQueryResponsePayload{
			Result: resultGroup, Cache: false, RefreshedAt: U.TimeNowIn(U.TimeZoneStringIST).Unix(), TimeZone: string(timezoneString), CacheMeta: meta}, http.StatusOK, "", "", false
	}
	resultGroup.Query = requestPayload
	resultGroup.IsShareable = isQueryShareable(requestPayload)
	return resultGroup, http.StatusOK, "", "", false
}
func isQueryShareable(queryGroup model.QueryGroup) bool {
	for _, query := range queryGroup.Queries {
		if query.GroupByProperties != nil && len(query.GroupByProperties) > 0 {
			return false
		}
	}
	return true
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
func QueryHandler(c *gin.Context) (interface{}, int, string, string, bool) {

	reqId := U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID)
	logCtx := log.WithFields(log.Fields{
		"reqId": reqId,
	})

	r := c.Request

	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectId == 0 {
		logCtx.Error("Query failed. Invalid project.")
		return nil, http.StatusUnauthorized, V1.INVALID_PROJECT, "Query failed. Invalid project.", true
	}

	var requestPayload QueryRequestPayload
	var dashboardId int64
	var unitId int64
	var err error
	var result *model.QueryResult
	hardRefresh := false
	preset := ""
	dashboardIdParam := c.Query("dashboard_id")
	unitIdParam := c.Query("dashboard_unit_id")
	refreshParam := c.Query("refresh")
	presetParam := c.Query("preset")
	var timezoneString U.TimeZoneString
	var statusCode int
	if U.PresetLookup[presetParam] != "" {
		preset = presetParam
	}
	if refreshParam != "" {
		hardRefresh, _ = strconv.ParseBool(refreshParam)
	}
	/*isQuery := false
	isQueryParam := c.Query("is_query")
	if isQueryParam != "" {
		isQuery, _ = strconv.ParseBool(isQueryParam)
	}*/

	isDashboardQueryRequest := dashboardIdParam != "" && unitIdParam != ""
	if isDashboardQueryRequest {
		dashboardId, err = strconv.ParseInt(dashboardIdParam, 10, 64)
		if err != nil || dashboardId == 0 {
			logCtx.WithError(err).Error("Query failed. Invalid DashboardID.")
			return nil, http.StatusBadRequest, V1.INVALID_INPUT, "Query failed. Invalid DashboardID.", true
		}
		unitId, err = strconv.ParseInt(unitIdParam, 10, 64)
		if err != nil || unitId == 0 {
			logCtx.WithError(err).Error("Query failed. Invalid DashboardUnitID.")
			return nil, http.StatusBadRequest, V1.INVALID_INPUT, "Query failed. Invalid DashboardUnitID.", true
		}
	}

	queryIdString := c.Query("query_id")
	if queryIdString == "" {
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&requestPayload); err != nil {
			logCtx.WithError(err).Error("Query failed. Json decode failed.")
			return nil, http.StatusBadRequest, V1.INVALID_INPUT, "Query failed. Json decode failed.", true
		}
	} else {
		_, query, err := store.GetStore().GetQueryAndClassFromQueryIdString(queryIdString, projectId)
		if err != "" {
			logCtx.Error(fmt.Sprintf("Query from queryIdString failed - %v", err))
			return nil, http.StatusBadRequest, V1.INVALID_INPUT, "Query failed. Json decode failed.", true
		}
		U.DecodePostgresJsonbToStructType(&query.Query, &requestPayload.Query)
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
	requestPayload.Query.SetTimeZone(timezoneString)

	for index := range requestPayload.Query.GroupByProperties {
		if requestPayload.Query.GroupByProperties[index].Type == U.PropertyTypeDateTime &&
			requestPayload.Query.GroupByProperties[index].Granularity == "" {
			requestPayload.Query.GroupByProperties[index].Granularity = U.DateTimeBreakdownDailyGranularity
		}
	}

	// Tracking dashboard query request.
	if isDashboardQueryRequest {
		if preset == "" {
			preset = U.GetPresetNameByFromAndTo(requestPayload.Query.From, requestPayload.Query.To, timezoneString)
		}
		model.SetDashboardCacheAnalytics(projectId, dashboardId, unitId, requestPayload.Query.From, requestPayload.Query.To, timezoneString)
	}

	// If refresh is passed, refresh only is Query.From is of today's beginning.
	if isDashboardQueryRequest && !H.ShouldAllowHardRefresh(requestPayload.Query.From, requestPayload.Query.To, timezoneString, hardRefresh) {
		shouldReturn, resCode, resMsg := H.GetResponseIfCachedDashboardQuery(
			reqId, projectId, dashboardId, unitId, preset, requestPayload.Query.From, requestPayload.Query.To, timezoneString)
		if shouldReturn && !hardRefresh {
			if resCode == http.StatusOK {
				return resMsg, resCode, "", "", false
			}
		}
	}

	err = requestPayload.Query.TransformDateTypeFilters()
	if err != nil {
		return nil, http.StatusBadRequest, V1.INVALID_INPUT, err.Error(), true
	}

	var cacheResult model.QueryResult
	shouldReturn, resCode, resMsg := H.GetResponseIfCachedQuery(c, projectId, &requestPayload.Query, cacheResult, isDashboardQueryRequest, reqId, false)

	if shouldReturn && !hardRefresh {
		if resCode == http.StatusOK {
			return resMsg, resCode, "", "", false
		}
		logCtx.WithError(err).Error("Query failed. Error Processing/Fetching data from Query cache")
		return nil, resCode, V1.PROCESSING_FAILED, "Error Processing/Fetching data from Query cache", true
	}

	/*if isDashboardQueryRequest && C.DisableDashboardQueryDBExecution() && !isQuery {
		logCtx.WithField("request_payload", requestPayload).Warn("Skip hitting db for queries from dashboard, if not found on cache.")
		return nil, resCode, V1.PROCESSING_FAILED, "Not found in cache. Execution suspended temporarily.", true
	}*/

	// If not found, set a placeholder for the query hash key that it has been running to avoid running again.
	model.SetQueryCachePlaceholder(projectId, &requestPayload.Query)
	H.SleepIfHeaderSet(c)

	enableOptimisedFilterOnEventUserQuery := c.Request.Header.Get(H.HeaderUserFilterOptForEventsAndUsers) == "true" ||
		C.EnableOptimisedFilterOnEventUserQuery()

	result, errCode, errMsg := store.GetStore().Analyze(projectId, requestPayload.Query, enableOptimisedFilterOnEventUserQuery)
	if errCode != http.StatusOK {
		model.DeleteQueryCacheKey(projectId, &requestPayload.Query)
		logCtx.Error("Failed to process query from DB - " + errMsg)
		return nil, errCode, V1.PROCESSING_FAILED, "Failed to process query from DB - " + errMsg, true
	}
	if result == nil {
		model.DeleteQueryCacheKey(projectId, &requestPayload.Query)
		logCtx.Error(" Result is nil - " + errMsg)
		return nil, errCode, V1.PROCESSING_FAILED, "Result is nil - " + errMsg, true
	}
	meta := H.CacheMeta{
		Timezone:       string(timezoneString),
		From:           requestPayload.Query.From,
		To:             requestPayload.Query.To,
		RefreshedAt:    U.TimeNowIn(U.TimeZoneStringIST).Unix(),
		LastComputedAt: U.TimeNowIn(U.TimeZoneStringIST).Unix(),
		Preset:         preset,
	}
	result.CacheMeta = meta
	model.SetQueryCacheResult(projectId, &requestPayload.Query, result)

	if isDashboardQueryRequest {
		model.SetCacheResultByDashboardIdAndUnitId(result, projectId, dashboardId, unitId, preset, requestPayload.Query.From, requestPayload.Query.To, timezoneString, meta)
		return H.DashboardQueryResponsePayload{Result: result, Cache: false, RefreshedAt: U.TimeNowIn(U.TimeZoneStringIST).Unix(), TimeZone: string(timezoneString), CacheMeta: meta}, http.StatusOK, "", "", false
	}
	result.Query = requestPayload.Query
	return result, http.StatusOK, "", "", false
}
