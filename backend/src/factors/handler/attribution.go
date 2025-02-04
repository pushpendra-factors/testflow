package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	C "factors/config"
	H "factors/handler/helpers"
	V1 "factors/handler/v1"
	mid "factors/middleware"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"fmt"
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
	projectId := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	logCtx := log.WithFields(log.Fields{
		"reqId": reqId, "project_id": projectId,
	})
	if projectId == 0 {
		logCtx.Error("Query failed. Invalid project.")
		return nil, http.StatusUnauthorized, V1.INVALID_PROJECT, "Query failed. Invalid project.", true
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

	requestPayload, dashboardId, unitId, isDashboardQueryRequest, statusCode, errorCode, errMsg, isErr := getValidAttributionQueryAndDetailsFromRequest(r, c, logCtx, projectId)
	if statusCode != http.StatusOK {
		return nil, statusCode, errorCode, errMsg, isErr
	}

	if requestPayload.Query == nil {
		return nil, http.StatusBadRequest, V1.INVALID_INPUT, "invalid query. empty query.", true
	}
	timezoneString, err = SetTimezoneForAttributionQuery(&requestPayload, projectId)

	if err != nil {
		return nil, http.StatusBadRequest, V1.INVALID_INPUT, "query failed. Failed to get Timezone", true
	}

	queryRange := float64(requestPayload.Query.To-requestPayload.Query.From) / float64(model.SecsInADay)
	if queryRange > model.QueryRangeLimit {
		logCtx.Error("Query failed. Query range is out of limit.")
		return nil, http.StatusBadRequest, V1.INVALID_INPUT, "Query failed. Query range is out of limit.", true
	}

	if requestPayload.Query.LookbackDays > model.LookBackWindowLimit {
		logCtx.Error("Query failed. LookbackDays exceeded the limit.")
		return nil, http.StatusBadRequest, V1.INVALID_INPUT, "Query failed. LookbackDays exceeded the limit.", true
	}

	// Tracking dashboard query request.
	if isDashboardQueryRequest {
		model.SetDashboardCacheAnalytics(projectId, dashboardId, unitId, requestPayload.Query.From, requestPayload.Query.To, timezoneString)
		if preset == "" && C.IsLastComputedWhitelisted(projectId) {
			preset = U.GetPresetNameByFromAndTo(requestPayload.Query.From, requestPayload.Query.To, timezoneString)
		}
	}

	// If refresh is passed, refresh only is Query.From is of today's beginning.
	if !hardRefresh && isDashboardQueryRequest && !H.ShouldAllowHardRefresh(requestPayload.Query.From, requestPayload.Query.To, timezoneString, hardRefresh) {

		effectiveFrom, effectiveTo := model.GetEffectiveTimeRangeForDashboardUnitAttributionQuery(requestPayload.Query.From, requestPayload.Query.To)
		if effectiveFrom == 0 || effectiveTo == 0 {
			return nil, http.StatusBadRequest, V1.INVALID_INPUT, "Query time range is not valid for attribution.", true
		}
		var shouldReturn bool
		var resCode int
		var resMsg interface{}

		if C.IsLastComputedWhitelisted(projectId) {

			shouldReturn, resCode, resMsg = H.GetResponseIfCachedDashboardQueryWithPreset(reqId, projectId, dashboardId, unitId, preset, effectiveFrom, effectiveTo, timezoneString)

		} else {

			shouldReturn, resCode, resMsg = H.GetResponseIfCachedDashboardQuery(reqId, projectId, dashboardId, unitId, effectiveFrom, effectiveTo, timezoneString)

		}
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
	if !hardRefresh {
		shouldReturn, resCode, resMsg := H.GetResponseIfCachedQuery(c, projectId, &attributionQueryUnitPayload, cacheResult, isDashboardQueryRequest, reqId, false)
		if shouldReturn {
			if resCode == http.StatusOK {
				return resMsg, resCode, "", "", false
			}
			logCtx.WithError(err).Error("Query failed. Error Processing/Fetching data from Query cache")
			return nil, resCode, V1.PROCESSING_FAILED, "Error Processing/Fetching data from Query cache", true
		}
	}

	// If not found, set a placeholder for the query hash key that it has been running to avoid running again.
	model.SetQueryCachePlaceholder(projectId, &attributionQueryUnitPayload)

	enableOptimisedFilterOnProfileQuery := c.Request.Header.Get(H.HeaderUserFilterOptForProfiles) == "true" ||
		C.EnableOptimisedFilterOnProfileQuery()

	enableOptimisedFilterOnEventUserQuery := c.Request.Header.Get(H.HeaderUserFilterOptForEventsAndUsers) == "true" ||
		C.EnableOptimisedFilterOnEventUserQuery()

	H.SleepIfHeaderSet(c)
	QueryKey, _ := attributionQueryUnitPayload.GetQueryCacheRedisKey(projectId)
	log.WithFields(log.Fields{"projectId": projectId,
		"QueryKey": QueryKey}).Info("AttributionHandler project_id")
	debugQueryKey := model.GetStringKeyFromCacheRedisKey(QueryKey)
	var result *model.QueryResult
	result, err = store.GetStore().ExecuteAttributionQueryV0(projectId, requestPayload.Query, debugQueryKey,
		enableOptimisedFilterOnProfileQuery, enableOptimisedFilterOnEventUserQuery)

	if err != nil {
		model.DeleteQueryCacheKey(projectId, &attributionQueryUnitPayload)
		logCtx.WithError(err).Error("Failed to process query from DB")
		return nil, http.StatusInternalServerError, V1.PROCESSING_FAILED, err.Error(), true
	}
	if result == nil {
		model.DeleteQueryCacheKey(projectId, &attributionQueryUnitPayload)
		logCtx.WithError(err).Error(" Result is nil")
		return nil, http.StatusInternalServerError, V1.PROCESSING_FAILED, "Result is nil " + err.Error(), true
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
		if C.IsLastComputedWhitelisted(projectId) {
			model.SetCacheResultByDashboardIdAndUnitIdWithPreset(result, projectId, dashboardId, unitId, preset,
				requestPayload.Query.From, requestPayload.Query.To, timezoneString, meta)
		} else {
			model.SetCacheResultByDashboardIdAndUnitId(result, projectId, dashboardId, unitId,
				requestPayload.Query.From, requestPayload.Query.To, timezoneString, meta, false)
		}

		return H.DashboardQueryResponsePayload{Result: result, Cache: false, RefreshedAt: U.TimeNowIn(U.TimeZoneStringIST).Unix(), CacheMeta: meta}, http.StatusOK, "", "", false
	}
	result.Query = requestPayload.Query
	return result, http.StatusOK, "", "", false
}

func getValidAttributionQueryAndDetailsFromRequest(r *http.Request, c *gin.Context, logCtx *log.Entry, projectId int64) (AttributionRequestPayload, int64, int64, bool, int, string, string, bool) {
	var dashboardId, unitId int64
	var err error
	queryPayload, requestPayload := AttributionRequestPayload{}, AttributionRequestPayload{}
	var dbQuery model.AttributionQueryUnit

	dashboardIdParam := c.Query("dashboard_id")
	unitIdParam := c.Query("dashboard_unit_id")
	queryIdString := c.Query("query_id")

	if queryIdString == "" {
		var hasFailed bool
		var errMsg string
		hasFailed, errMsg, requestPayload = decodeAttributionPayload(r, logCtx)
		if hasFailed {
			logCtx.WithField("errMsg", errMsg).Error("Query failed. Json decode failed.")
			return queryPayload, 0, 0, false, http.StatusBadRequest, V1.INVALID_INPUT, "Query failed. Json decode failed.", true
		}
	}

	isDashboardQueryRequest := dashboardIdParam != "" && unitIdParam != ""
	if isDashboardQueryRequest {

		dashboardId, err = strconv.ParseInt(dashboardIdParam, 10, 64)
		if err != nil || dashboardId == 0 {
			logCtx.WithError(err).Error("Query failed. Invalid DashboardID.")
			return queryPayload, dashboardId, 0, true, http.StatusBadRequest, V1.INVALID_INPUT, "Query failed. Invalid DashboardID.", true
		}
		unitId, err = strconv.ParseInt(unitIdParam, 10, 64)
		if err != nil || unitId == 0 {
			logCtx.WithError(err).Error("Query failed. Invalid DashboardUnitID.")
			return queryPayload, dashboardId, unitId, true, http.StatusBadRequest, V1.INVALID_INPUT, "Query failed. Invalid DashboardUnitID.", true
		}
		_, query, err := store.GetStore().GetQueryFromUnitID(projectId, unitId)
		if err != "" {
			logCtx.Error(fmt.Sprintf("Query from queryIdString failed - %v", err))
			return queryPayload, dashboardId, unitId, true, http.StatusBadRequest, V1.INVALID_INPUT, "Query failed. Json decode failed.", true
		}
		if query.LockedForCacheInvalidation {
			return queryPayload, dashboardId, unitId, true, http.StatusConflict, V1.PROCESSING_FAILED, "Query is not processed due to saved query updated", false
		}
		errDecode := U.DecodePostgresJsonbToStructType(&query.Query, &dbQuery)
		if errDecode != nil {
			log.WithError(errDecode).WithField("project_id", projectId).Error("Failed while decoding query: ", errDecode)
			return queryPayload, dashboardId, unitId, true, http.StatusBadRequest, V1.INVALID_INPUT, "Query failed. Json decode failed.", true
		}
		queryPayload.Query = dbQuery.Query
	} else if queryIdString != "" {
		_, query, err := store.GetStore().GetQueryAndClassFromQueryIdString(queryIdString, projectId)
		if err != "" {
			logCtx.Error(fmt.Sprintf("Query from queryIdString failed - %v", err))
			return queryPayload, 0, 0, false, http.StatusBadRequest, V1.INVALID_INPUT, "Query failed. Json decode failed.", true
		}
		if query.LockedForCacheInvalidation {
			return queryPayload, 0, 0, false, http.StatusConflict, V1.PROCESSING_FAILED, "Query is not processed due to saved query updated", false
		}
		errDecode := U.DecodePostgresJsonbToStructType(&query.Query, &dbQuery)
		if errDecode != nil {
			log.WithError(errDecode).WithField("project_id", projectId).Error("Failed while decoding query: ", errDecode)
			return queryPayload, dashboardId, unitId, true, http.StatusBadRequest, V1.INVALID_INPUT, "Query failed. Json decode failed.", true
		}
		queryPayload.Query = dbQuery.Query
	} else {
		queryPayload = requestPayload
	}

	if queryIdString == "" {
		queryUnitPayload := model.AttributionQueryUnit{Query: queryPayload.Query}

		queryUnitPayload.SetQueryDateRange(requestPayload.Query.From, requestPayload.Query.To)
		if requestPayload.Query.Timezone != "" {
			queryUnitPayload.SetTimeZone(U.TimeZoneString(requestPayload.Query.Timezone))
		}

		if requestPayload.Query.KPI.Class != "" {
			var inputGroupByTimestamp string
			for _, query := range requestPayload.Query.KPI.Queries {
				if query.GroupByTimestamp != "" {
					inputGroupByTimestamp = query.GroupByTimestamp
				}
			}
			for index, query := range queryUnitPayload.Query.KPI.Queries {
				if query.GroupByTimestamp != "" {
					queryUnitPayload.Query.KPI.Queries[index].GroupByTimestamp = inputGroupByTimestamp
				}
			}
		}

		queryPayload.Query = queryUnitPayload.Query
	}

	return queryPayload, dashboardId, unitId, isDashboardQueryRequest, http.StatusOK, "", "", false
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
	// commenting out this for KPI queries for attribution
	// decoder2.DisallowUnknownFields()
	var requestPayloadUnit model.AttributionQueryUnit
	if err = decoder2.Decode(&requestPayloadUnit); err == nil {
		requestPayload.Query = requestPayloadUnit.Query
		return false, "", requestPayload
	}
	logCtx.WithError(err).Error("query failed as JSON decode failed")
	return true, "Query failed. Json decode failed", requestPayload
}

// SetTimezoneForAttributionQuery sets timezone for the attribution query
func SetTimezoneForAttributionQuery(requestPayload *AttributionRequestPayload, projectId int64) (U.TimeZoneString, error) {
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
		if requestPayload.Query.AnalyzeType == model.AnalyzeTypeHSDeals || requestPayload.Query.AnalyzeType == model.AnalyzeTypeSFOpportunities {
			if requestPayload.Query.KPI.Queries[0].Timezone != "" {
				_, err := time.LoadLocation(requestPayload.Query.KPI.Queries[0].Timezone)
				if err != nil {
					logCtx.WithError(err).Error("Query failed. Invalid Timezone")
					return "", err
				}

				timezoneString = U.TimeZoneString(requestPayload.Query.KPI.Queries[0].Timezone)
			} else {
				timezoneString, statusCode = store.GetStore().GetTimezoneForProject(projectId)
				if statusCode != http.StatusFound {
					logCtx.Error("query failed. Failed to get Timezone")
					return "", errors.New("query failed. Failed to get Timezone")
				}
			}

			requestPayload.Query.KPI.SetTimeZone(timezoneString)
			err := requestPayload.Query.KPI.TransformDateTypeFilters()
			if err != nil {
				return "", err
			}
		}
	}
	return timezoneString, nil
}
