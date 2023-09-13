package v1

import (
	"bytes"
	"encoding/json"
	"errors"
	C "factors/config"
	H "factors/handler/helpers"
	mid "factors/middleware"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm/dialects/postgres"

	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
)

type AttributionRequestPayloadV1 struct {
	Query *model.AttributionQueryV1 `json:"query"`
}

// AttributionHandlerV1 godoc
// @Summary To run attribution query.
// @Tags CoreQuery
// @Accept  json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param dashboard_id query integer false "Dashboard ID"
// @Param dashboard_unit_id query integer false "Dashboard Unit ID"
// @Param query body handler.AttributionRequestPayloadV1 true "Query payload"
// @Success 200 {string} json "{"result": model.QueryResult, "cache": false, "refreshed_at": timestamp}"
// @Router /{project_id}/attribution/query [post]
func AttributionHandlerV1(c *gin.Context) (interface{}, int, string, string, bool) {

	r := c.Request
	reqId := U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID)
	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	logCtx := log.WithFields(log.Fields{
		"reqId": reqId, "project_id": projectId,
	})
	if projectId == 0 {
		logCtx.Error("Query failed. Invalid project.")
		return nil, http.StatusUnauthorized, INVALID_PROJECT, "Query failed. Invalid project.", true
	}

	var err error
	var timezoneString U.TimeZoneString
	preset := ""
	hardRefresh := false
	refreshParam := c.Query("refresh")
	presetParam := c.Query("preset") // check preset

	if U.PresetLookup[presetParam] != "" && C.IsLastComputedWhitelisted(projectId) {
		preset = presetParam
	}
	if refreshParam != "" {
		hardRefresh, _ = strconv.ParseBool(refreshParam)
	}

	requestPayload, dashboardId, unitId, isDashboardQueryRequest, isDashboardQueryLocked, statusCode, errorCode, errMsg, isErr := getValidAttributionQueryAndDetailsFromRequestV1(r, c, logCtx, projectId)
	if statusCode != http.StatusOK {
		return nil, statusCode, errorCode, errMsg, isErr
	}

	logCtx = logCtx.WithFields(log.Fields{
		"dashboardId": dashboardId,
		"unitId":      unitId,
		"To":          requestPayload.Query.To,
		"From":        requestPayload.Query.From,
	})
	if requestPayload.Query == nil || requestPayload.Query.KPIQueries == nil || len(requestPayload.Query.KPIQueries) == 0 ||
		requestPayload.Query.KPIQueries[0].KPI.Queries == nil || len(requestPayload.Query.KPIQueries[0].KPI.Queries) == 0 {
		return nil, http.StatusBadRequest, INVALID_INPUT, "invalid query. empty query.", true
	}

	timezoneString, err = SetTimezoneForAttributionQueryV1(&requestPayload, projectId)

	if err != nil {
		return nil, http.StatusBadRequest, INVALID_INPUT, "query failed. Failed to get Timezone", true
	}

	queryRange := float64(requestPayload.Query.To-requestPayload.Query.From) / float64(model.SecsInADay)
	if queryRange > model.QueryRangeLimit {
		logCtx.Error("Query failed. Query range is out of limit.")
		return nil, http.StatusBadRequest, INVALID_INPUT, "Query failed. Query range is out of limit.", true
	}

	if requestPayload.Query.LookbackDays > model.LookBackWindowLimit {
		logCtx.Error("Query failed. LookbackDays exceeded the limit.")
		return nil, http.StatusBadRequest, INVALID_INPUT, "Query failed. LookbackDays exceeded the limit.", true
	}

	enableOptimisedFilterOnProfileQuery := c.Request.Header.Get(H.HeaderUserFilterOptForProfiles) == "true" ||
		C.EnableOptimisedFilterOnProfileQuery()

	enableOptimisedFilterOnEventUserQuery := c.Request.Header.Get(H.HeaderUserFilterOptForEventsAndUsers) == "true" ||
		C.EnableOptimisedFilterOnEventUserQuery()
	attributionQueryUnitPayload := model.AttributionQueryUnitV1{
		Class: model.QueryClassAttribution,
		Query: requestPayload.Query,
	}
	attributionQueryUnitPayload.SetTimeZone(timezoneString)
	err = attributionQueryUnitPayload.TransformDateTypeFilters()
	if err != nil {
		return nil, http.StatusBadRequest, INVALID_INPUT, err.Error(), true
	}

	// Tracking dashboard query request.
	if isDashboardQueryRequest {
		model.SetDashboardCacheAnalytics(projectId, dashboardId, unitId, requestPayload.Query.From, requestPayload.Query.To, timezoneString)
		if preset == "" && C.IsLastComputedWhitelisted(projectId) {
			preset = U.GetPresetNameByFromAndTo(requestPayload.Query.From, requestPayload.Query.To, timezoneString)
		}
	}

	if isDashboardQueryLocked {
		QueryKey, _ := attributionQueryUnitPayload.GetQueryCacheRedisKey(projectId)
		debugQueryKey := model.GetStringKeyFromCacheRedisKey(QueryKey)
		var result *model.QueryResult
		result, err = store.GetStore().ExecuteAttributionQueryV1(projectId, requestPayload.Query, debugQueryKey,
			enableOptimisedFilterOnProfileQuery, enableOptimisedFilterOnEventUserQuery, unitId)
		if err != nil {
			logCtx.Info("Failed to process query from DB - attributionv1", err.Error())
			return nil, http.StatusInternalServerError, PROCESSING_FAILED, err.Error(), true
		}
		if result == nil {
			logCtx.WithError(err).Error(" Result is nil")
			return nil, http.StatusInternalServerError, PROCESSING_FAILED, "Result is nil " + err.Error(), true
		}

		return H.DashboardQueryResponsePayload{Result: result, Cache: false, RefreshedAt: U.TimeNowIn(U.TimeZoneStringIST).Unix(), CacheMeta: nil}, http.StatusOK, "", "", false
	}

	// If refresh is passed, refresh only is Query.From is of today's beginning.

	logCtx.WithFields(log.Fields{
		"requestPayload": requestPayload,
	}).Info("Attribution query debug request payload")
	if !hardRefresh && isDashboardQueryRequest && !H.ShouldAllowHardRefresh(requestPayload.Query.From, requestPayload.Query.To, timezoneString, hardRefresh) {
		//todo satya: check if we want to use effective to and from in this flow
		// if common flow and merge is enabled for the project
		if C.IsAllowedAttributionCommonFlow(projectId) {
			logCtx.Info("Running the common DB cache flow")
			return runTheCommonDBFlow(reqId, projectId, dashboardId, unitId, requestPayload, timezoneString, enableOptimisedFilterOnProfileQuery, enableOptimisedFilterOnEventUserQuery, logCtx)
		}
	}

	// If refresh is passed, refresh only is Query.From is of today's beginning.
	if !hardRefresh && isDashboardQueryRequest && !H.ShouldAllowHardRefresh(requestPayload.Query.From, requestPayload.Query.To, timezoneString, hardRefresh) {

		effectiveFrom, effectiveTo := model.GetEffectiveTimeRangeForDashboardUnitAttributionQuery(requestPayload.Query.From, requestPayload.Query.To)
		if effectiveFrom == 0 || effectiveTo == 0 {
			return nil, http.StatusBadRequest, INVALID_INPUT, "Query time range is not valid for attribution.", true
		}
		var shouldReturn bool
		var resCode int
		var resMsg interface{}

		if C.IsAllowedAttributionDBCacheLookup(projectId) {
			logCtx.Info("Hitting the DB cache lookup")
			shouldReturn, resCode, resMsg = H.GetResponseFromDBCaching(reqId, projectId, dashboardId, unitId, effectiveFrom, effectiveTo, timezoneString)
			logCtx.WithFields(log.Fields{
				"should_return": shouldReturn,
				"res_code":      resCode,
				"res_msg":       resMsg,
			}).Info("Hitting the DB cache lookup")
			if shouldReturn {
				if resCode == http.StatusOK {
					return resMsg, resCode, "", "", false
				}
			}
		}

		if C.IsLastComputedWhitelisted(projectId) {
			logCtx.Info("Hitting GetResponseIfCachedDashboardQueryWithPreset")
			shouldReturn, resCode, resMsg = H.GetResponseIfCachedDashboardQueryWithPreset(reqId, projectId, dashboardId, unitId, preset, effectiveFrom, effectiveTo, timezoneString)
			logCtx.WithFields(log.Fields{
				"should_return": shouldReturn,
				"res_code":      resCode,
				"res_msg":       resMsg,
			}).Info("Response from GetResponseIfCachedDashboardQueryWithPreset")
		} else {
			logCtx.Info("Hitting GetResponseIfCachedDashboardQuery")
			shouldReturn, resCode, resMsg = H.GetResponseIfCachedDashboardQuery(reqId, projectId, dashboardId, unitId, effectiveFrom, effectiveTo, timezoneString)
			logCtx.WithFields(log.Fields{
				"should_return": shouldReturn,
				"res_code":      resCode,
				"res_msg":       resMsg,
			}).Info("Response from GetResponseIfCachedDashboardQuery")
		}

		if shouldReturn {
			if resCode == http.StatusOK {
				return resMsg, resCode, "", "", false
			}
		}
	}

	var cacheResult model.QueryResult
	if !hardRefresh {
		logCtx.WithFields(log.Fields{
			"query to form cache key": attributionQueryUnitPayload,
		}).Info("Hitting GetResponseIfCachedQuery")
		shouldReturn, resCode, resMsg := H.GetResponseIfCachedQuery(c, projectId, &attributionQueryUnitPayload, cacheResult, isDashboardQueryRequest, reqId, false)
		logCtx.WithFields(log.Fields{
			"should_return": shouldReturn,
			"res_code":      resCode,
			"res_msg":       resMsg,
		}).Info("Response from GetResponseIfCachedQuery")
		if shouldReturn {
			if resCode == http.StatusOK {
				return resMsg, resCode, "", "", false
			}
			logCtx.WithError(err).Error("Query failed. Error Processing/Fetching data from Query cache")
			return nil, resCode, PROCESSING_FAILED, "Error Processing/Fetching data from Query cache", true
		}
	}

	// If not found, set a placeholder for the query hash key that it has been running to avoid running again.
	model.SetQueryCachePlaceholder(projectId, &attributionQueryUnitPayload)

	H.SleepIfHeaderSet(c)

	QueryKey, _ := attributionQueryUnitPayload.GetQueryCacheRedisKey(projectId)
	debugQueryKey := model.GetStringKeyFromCacheRedisKey(QueryKey)

	var result *model.QueryResult

	logCtx.Info("Hitting ExecuteAttributionQueryV1")
	result, err = store.GetStore().ExecuteAttributionQueryV1(projectId, requestPayload.Query, debugQueryKey,
		enableOptimisedFilterOnProfileQuery, enableOptimisedFilterOnEventUserQuery, unitId)

	if err != nil {
		model.DeleteQueryCacheKey(projectId, &attributionQueryUnitPayload)
		logCtx.Info("Failed to process query from DB - attributionv1", err.Error())
		return nil, http.StatusInternalServerError, PROCESSING_FAILED, err.Error(), true
	}
	if result == nil {
		model.DeleteQueryCacheKey(projectId, &attributionQueryUnitPayload)
		logCtx.WithError(err).Error(" Result is nil")
		return nil, http.StatusInternalServerError, PROCESSING_FAILED, "Result is nil " + err.Error(), true
	}
	meta := model.CacheMeta{
		Timezone:       string(timezoneString),
		From:           requestPayload.Query.From,
		To:             requestPayload.Query.To,
		RefreshedAt:    U.TimeNowIn(U.TimeZoneStringIST).Unix(),
		LastComputedAt: U.TimeNowIn(U.TimeZoneStringIST).Unix(),
		Preset:         preset,
	}
	result.CacheMeta = meta
	model.SetQueryCacheResult(projectId, &attributionQueryUnitPayload, result)
	if isDashboardQueryRequest {
		effectiveFrom, effectiveTo := model.GetEffectiveTimeRangeForDashboardUnitAttributionQuery(requestPayload.Query.From, requestPayload.Query.To)
		if effectiveFrom == 0 || effectiveTo == 0 {
			return nil, http.StatusBadRequest, INVALID_INPUT, "Query time range is not valid for attribution.", true
		}
		if C.IsLastComputedWhitelisted(projectId) {
			model.SetCacheResultByDashboardIdAndUnitIdWithPreset(result, projectId, dashboardId, unitId, preset,
				effectiveFrom, effectiveTo, timezoneString, meta)
		} else {
			model.SetCacheResultByDashboardIdAndUnitId(result, projectId, dashboardId, unitId,
				effectiveFrom, effectiveTo, timezoneString, meta)
		}

		return H.DashboardQueryResponsePayload{Result: result, Cache: false, RefreshedAt: U.TimeNowIn(U.TimeZoneStringIST).Unix(), CacheMeta: meta}, http.StatusOK, "", "", false
	}
	result.Query = requestPayload.Query
	return result, http.StatusOK, "", "", false
}

func getValidAttributionQueryAndDetailsFromRequestV1(r *http.Request, c *gin.Context, logCtx *log.Entry, projectId int64) (AttributionRequestPayloadV1, int64, int64, bool, bool, int, string, string, bool) {
	var dashboardId, unitId int64
	var err error
	queryPayload, requestPayload, isDashboardQueryLocked := AttributionRequestPayloadV1{}, AttributionRequestPayloadV1{}, false
	var dbQuery model.AttributionQueryUnitV1

	dashboardIdParam := c.Query("dashboard_id")
	unitIdParam := c.Query("dashboard_unit_id")
	queryIdString := c.Query("query_id")
	isDashboardQueryRequest := dashboardIdParam != "" && unitIdParam != ""
	logCtx.WithFields(log.Fields{
		"dashboardIdParam": dashboardIdParam,
		"unitIdParam":      unitIdParam,
		"queryIdString":    queryIdString,
	}).Info("query requestPayload")
	if queryIdString == "" {
		var hasFailed bool
		var errMsg string

		hasFailed, errMsg, requestPayload = decodeAttributionPayload(r, logCtx)
		if hasFailed {
			logCtx.WithField("errMsg", errMsg).Error("Query failed. Json decode failed.")
			return queryPayload, 0, 0, false, isDashboardQueryLocked, http.StatusBadRequest, INVALID_INPUT, "Query failed. Json decode failed.", true
		}
	}

	if isDashboardQueryRequest {

		dashboardId, err = strconv.ParseInt(dashboardIdParam, 10, 64)
		if err != nil || dashboardId == 0 {
			logCtx.WithError(err).Error("Query failed. Invalid DashboardID.")
			return queryPayload, dashboardId, 0, true, isDashboardQueryLocked, http.StatusBadRequest, INVALID_INPUT, "Query failed. Invalid DashboardID.", true
		}
		unitId, err = strconv.ParseInt(unitIdParam, 10, 64)
		if err != nil || unitId == 0 {
			logCtx.WithError(err).Error("Query failed. Invalid DashboardUnitID.")
			return queryPayload, dashboardId, unitId, true, isDashboardQueryLocked, http.StatusBadRequest, INVALID_INPUT, "Query failed. Invalid DashboardUnitID.", true
		}
		_, query, err := store.GetStore().GetQueryFromUnitID(projectId, unitId)
		if err != "" {
			logCtx.Error(fmt.Sprintf("Query from queryIdString failed - %v", err))
			return queryPayload, dashboardId, unitId, true, isDashboardQueryLocked, http.StatusBadRequest, INVALID_INPUT, "Query failed. Json decode failed.", true
		}
		isDashboardQueryLocked = query.LockedForCacheInvalidation
		U.DecodePostgresJsonbToStructType(&query.Query, &dbQuery)
		queryPayload.Query = dbQuery.Query
	} else if queryIdString != "" {
		_, query, err := store.GetStore().GetQueryAndClassFromQueryIdString(queryIdString, projectId)
		if err != "" {
			logCtx.Error(fmt.Sprintf("Query from queryIdString failed - %v", err))
			return queryPayload, 0, 0, false, isDashboardQueryLocked, http.StatusBadRequest, INVALID_INPUT, "Query failed. Json decode failed.", true
		}
		U.DecodePostgresJsonbToStructType(&query.Query, &dbQuery)
		queryPayload.Query = dbQuery.Query
	} else {
		queryPayload = requestPayload
	}

	if queryIdString == "" {
		queryUnitPayload := model.AttributionQueryUnitV1{Query: queryPayload.Query}
		queryUnitPayload.SetQueryDateRange(requestPayload.Query.From, requestPayload.Query.To)
		if requestPayload.Query.Timezone != "" {
			queryUnitPayload.SetTimeZone(U.TimeZoneString(requestPayload.Query.Timezone))
		}

		if requestPayload.Query.KPIQueries[0].KPI.Class != "" {
			var inputGroupByTimestamp string

			for groupIndex := range requestPayload.Query.KPIQueries {
				for _, query := range requestPayload.Query.KPIQueries[groupIndex].KPI.Queries {
					if query.GroupByTimestamp != "" {
						inputGroupByTimestamp = query.GroupByTimestamp
					}
				}
			}

			for groupIndex := range requestPayload.Query.KPIQueries {
				for index, query := range requestPayload.Query.KPIQueries[groupIndex].KPI.Queries {
					if query.GroupByTimestamp != "" {
						requestPayload.Query.KPIQueries[groupIndex].KPI.Queries[index].GroupByTimestamp = inputGroupByTimestamp
					}
				}
			}

		}
		queryPayload.Query = queryUnitPayload.Query
	}

	return queryPayload, dashboardId, unitId, isDashboardQueryRequest, isDashboardQueryLocked, http.StatusOK, "", "", false
}

func runTheCommonDBFlow(reqId string, projectId int64, dashboardId int64, unitId int64, requestPayload AttributionRequestPayloadV1,
	timezoneString U.TimeZoneString, enableOptimisedFilterOnProfileQuery bool, enableOptimisedFilterOnEventUserQuery bool,
	logCtx *log.Entry) (interface{}, int, string, string, bool) {

	logCtx.Info("Hitting the DB cache lookup runTheCommonDBFlow")
	shouldReturn, resCode, resMsg := H.GetResponseFromDBCaching(reqId, projectId, dashboardId, unitId,
		requestPayload.Query.From, requestPayload.Query.To, timezoneString)
	logCtx.WithFields(log.Fields{
		"should_return": shouldReturn,
		"res_code":      resCode,
		"dashboard_Id":  dashboardId,
		"unit_id":       unitId,
	}).Info("Hitting the DB cache lookup")
	if shouldReturn {
		logCtx.Info("Found the result in DB runTheCommonDBFlow")
		if resCode == http.StatusOK {
			return resMsg, resCode, "", "", false
		}
	}

	/*
		errCode, cacheResult := store.GetStore().FetchCachedResultFromDataBase(reqId, projectId, dashboardId, unitId,
			requestPayload.Query.From, requestPayload.Query.To)

		if errCode == http.StatusFound {
			return H.DashboardQueryResponsePayload{Result: cacheResult.Result, Cache: true, RefreshedAt: cacheResult.ComputedAt,
				CacheMeta: cacheResult.CreatedAt}, http.StatusOK, "", "", false
		}*/

	// Reached here, it means that the exact date range is not there in the DB result storage
	// Check for monthly range query, we assume that the range has continuous date range inputs
	isMonthsQuery, last12Months := U.IsAMonthlyRangeQuery(timezoneString, requestPayload.Query.From, requestPayload.Query.To)
	logCtx.WithFields(log.Fields{
		"timezoneString": timezoneString,
		"from":           requestPayload.Query.From,
		"to":             requestPayload.Query.To,
		"last12Months":   last12Months,
		"isMonthsQuery":  isMonthsQuery,
	}).Info("debug last12Months")
	if isMonthsQuery {

		monthsToRun := U.GetAllValidRangesInBetween(requestPayload.Query.From, requestPayload.Query.To, last12Months)
		logCtx.WithFields(log.Fields{
			"monthsToRun": monthsToRun,
		}).Info("Figured it a Month range query, running")
		hasFailed, mergedResult, computeMeta := RunMultipleRangeAttributionQueries(projectId, dashboardId, unitId, requestPayload,
			timezoneString, reqId, enableOptimisedFilterOnProfileQuery, enableOptimisedFilterOnEventUserQuery,
			monthsToRun, logCtx)
		if hasFailed {
			logCtx.Error("Month range query failed to run")
			return nil, http.StatusInternalServerError, PROCESSING_FAILED, "Month range query failed to run", true
		}

		return H.DashboardQueryResponsePayload{Result: mergedResult, Cache: false, RefreshedAt: U.TimeNowIn(U.TimeZoneStringIST).Unix(),
			CacheMeta: mergedResult.CacheMeta, ComputeMeta: computeMeta}, http.StatusOK, "", "", false
	}

	// Check for weekly range query, we assume that the range has continuous date range inputs
	isWeeksQuery, last48Weeks := U.IsAWeeklyRangeQuery(timezoneString, requestPayload.Query.From, requestPayload.Query.To)
	logCtx.WithFields(log.Fields{
		"timezoneString": timezoneString,
		"from":           requestPayload.Query.From,
		"to":             requestPayload.Query.To,
		"last48Weeks":    last48Weeks,
		"isWeeksQuery":   isWeeksQuery,
	}).Info("debug last48Weeks")
	if isWeeksQuery {

		weeksToRun := U.GetAllValidRangesInBetween(requestPayload.Query.From, requestPayload.Query.To, last48Weeks)
		logCtx.WithFields(log.Fields{
			"weeksToRun": weeksToRun,
		}).Info("Figured it a Week range query, running")
		hasFailed, mergedResult, computeMeta := RunMultipleRangeAttributionQueries(projectId, dashboardId, unitId, requestPayload,
			timezoneString, reqId, enableOptimisedFilterOnProfileQuery, enableOptimisedFilterOnEventUserQuery,
			weeksToRun, logCtx)

		if hasFailed {
			logCtx.Error("Week range query failed to run")
			return nil, http.StatusInternalServerError, PROCESSING_FAILED, "Week range query failed to run", true
		}
		logCtx.WithFields(log.Fields{
			"mergedResult": mergedResult,
			"computeMeta":  computeMeta,
		}).Info("post RunMultipleRangeAttributionQueries merge - mergedRows")
		return H.DashboardQueryResponsePayload{Result: mergedResult, Cache: false, RefreshedAt: U.TimeNowIn(U.TimeZoneStringIST).Unix(),
			CacheMeta: mergedResult.CacheMeta, ComputeMeta: computeMeta}, http.StatusOK, "", "", false
	}

	// Nothing worked, fail the query finally
	logCtx.Error("Query failed. The query range is not a standard range, aborting ")
	return nil, http.StatusBadRequest, INVALID_INPUT, "Query failed. The query range is not a standard range, aborting ", true
}

func AttributionCommonHandlerV1() {

}

func runAttributionQuery() {

}

func cacheAttributionResultInRedis() {

}

func persistAttributionResultInDB() {

}

func enrichRequestUsingAttributionConfig(c *gin.Context, projectID int64, requestPayload *AttributionRequestPayloadV1, logCtx *log.Entry) {

	// pulling project setting to build attribution query
	settings, errCode := store.GetStore().GetProjectSetting(projectID)
	if errCode != http.StatusFound {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Failed to get project settings during attribution call."})
	}

	attributionConfig, err1 := decodeAttributionConfig(settings.AttributionConfig)
	if err1 != nil {
		c.AbortWithStatusJSON(errCode, gin.H{"error": "Failed to decode attribution config from project settings."})
	}

	//Todo (Anil) Add enrichment of attribution Window, handle case of 'Entire User Journey'
	for groupQueryIndex := range requestPayload.Query.KPIQueries {
		switch requestPayload.Query.KPIQueries[groupQueryIndex].AnalyzeType {

		case model.AnalyzeTypeUsers:
			requestPayload.Query.KPIQueries[groupQueryIndex].RunType = model.RunTypeUser
		case model.AnalyzeTypeUserKPI:
			requestPayload.Query.KPIQueries[groupQueryIndex].RunType = model.RunTypeUserKPI
		case model.AnalyzeTypeHSDeals:
			if &attributionConfig != nil && attributionConfig.AnalyzeTypeHSCompaniesEnabled == true {
				requestPayload.Query.KPIQueries[groupQueryIndex].RunType = model.RunTypeHSCompanies
			} else if &attributionConfig != nil && attributionConfig.AnalyzeTypeHSDealsEnabled == true {
				requestPayload.Query.KPIQueries[groupQueryIndex].RunType = model.RunTypeHSDeals
			} else {
				logCtx.WithFields(log.Fields{"Query": requestPayload.Query, "AttributionConfig": attributionConfig}).Error("Failed to set analyze type")
				c.AbortWithStatusJSON(errCode, gin.H{"error": "Invalid config/query. Failed to set analyze type from attribution config & project settings."})
			}
		case model.AnalyzeTypeSFOpportunities:
			if &attributionConfig != nil && attributionConfig.AnalyzeTypeSFAccountsEnabled == true {
				requestPayload.Query.KPIQueries[groupQueryIndex].RunType = model.RunTypeSFAccounts
			} else if &attributionConfig != nil && attributionConfig.AnalyzeTypeSFOpportunitiesEnabled == true {
				requestPayload.Query.KPIQueries[groupQueryIndex].RunType = model.RunTypeSFOpportunities
			} else {
				logCtx.WithFields(log.Fields{"Query": requestPayload.Query, "AttributionConfig": attributionConfig}).Error("Failed to set analyze type")
				c.AbortWithStatusJSON(errCode, gin.H{"error": "Invalid config/query. Failed to set analyze type from attribution config & project settings."})
			}
		default:
			logCtx.WithFields(log.Fields{"Query": requestPayload.Query, "AttributionConfig": attributionConfig}).Error("Failed to set analyze type")
			c.AbortWithStatusJSON(errCode, gin.H{"error": "Invalid config/query. Failed to set analyze type from attribution config & project settings."})
		}
	}
}

// decodeAttributionConfig decode attribution config from project settings to map
func decodeAttributionConfig(config *postgres.Jsonb) (model.AttributionConfig, error) {
	attributionConfig := model.AttributionConfig{}
	if config == nil {
		return attributionConfig, nil
	}

	err := json.Unmarshal(config.RawMessage, &attributionConfig)
	if err != nil {
		return attributionConfig, err
	}

	return attributionConfig, nil
}

// decodeAttributionPayload decodes attribution requestPayload for 2 json formats to support old and new
// request formats
func decodeAttributionPayload(r *http.Request, logCtx *log.Entry) (bool, string, AttributionRequestPayloadV1) {

	var err error
	var requestPayload AttributionRequestPayloadV1
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
	// commenting out this for KPI queries for attribution
	// decoder2.DisallowUnknownFields()
	var requestPayloadUnit model.AttributionQueryUnitV1
	if err = decoder2.Decode(&requestPayloadUnit); err == nil {
		requestPayload.Query = requestPayloadUnit.Query
		return false, "", requestPayload
	}
	logCtx.WithError(err).Error("query failed as JSON decode failed")
	return true, "Query failed. Json decode failed", requestPayload
}

// SetTimezoneForAttributionQueryV1 sets timezone for the attribution query
func SetTimezoneForAttributionQueryV1(requestPayload *AttributionRequestPayloadV1, projectId int64) (U.TimeZoneString, error) {
	var timezoneString U.TimeZoneString
	logCtx := log.WithField("project_id", projectId)
	if requestPayload.Query.Timezone != "" {
		_, err := time.LoadLocation(requestPayload.Query.Timezone)
		if err != nil {
			logCtx.WithError(err).Error("Query failed. Invalid Timezone")
			return "", err
		}
		timezoneString = U.TimeZoneString(requestPayload.Query.Timezone)

	} else {
		var statusCode int
		timezoneString, statusCode = store.GetStore().GetTimezoneForProject(projectId)
		if statusCode != http.StatusFound {
			logCtx.Error("query failed. Failed to get Timezone from project")
			return "", errors.New("query failed. Failed to get Timezone from project")
		}

		// For a KPI query, set the timezone internally for correct execution
		if requestPayload.Query.KPIQueries[0].KPI.Queries[0].Timezone != "" {
			_, err := time.LoadLocation(string(requestPayload.Query.KPIQueries[0].KPI.Queries[0].Timezone))
			if err != nil {
				logCtx.WithError(err).Error("Query failed. Invalid Timezone")
				return "", err
			}

			timezoneString = U.TimeZoneString(requestPayload.Query.KPIQueries[0].KPI.Queries[0].Timezone)
		} else {
			timezoneString, statusCode = store.GetStore().GetTimezoneForProject(projectId)
			if statusCode != http.StatusFound {
				logCtx.Error("query failed. Failed to get Timezone")
				return "", errors.New("query failed. Failed to get Timezone")
			}
		}
		for index := range requestPayload.Query.KPIQueries {
			requestPayload.Query.KPIQueries[index].KPI.SetTimeZone(timezoneString)
			err := requestPayload.Query.KPIQueries[index].KPI.TransformDateTypeFilters()
			if err != nil {
				return "", err
			}
		}
	}
	return timezoneString, nil
}
