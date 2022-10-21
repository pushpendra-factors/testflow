package v1

import (
	"bytes"
	"encoding/json"
	C "factors/config"
	H "factors/handler/helpers"
	mid "factors/middleware"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"fmt"
	"github.com/jinzhu/gorm/dialects/postgres"
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

// AttributionHandlerV1 godoc
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
	var requestPayload AttributionRequestPayload
	var dashboardId int64
	var unitId int64
	var timezoneString U.TimeZoneString
	var statusCode int
	preset := ""
	hardRefresh := false
	dashboardIdParam := c.Query("dashboard_id")
	unitIdParam := c.Query("dashboard_unit_id")
	refreshParam := c.Query("refresh")
	presetParam := c.Query("preset") // check preset

	if U.PresetLookup[presetParam] != "" && C.IsLastComputedWhitelisted(projectId) {
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
			return nil, http.StatusBadRequest, INVALID_INPUT, "Query failed. Invalid DashboardID.", true
		}
		unitId, err = strconv.ParseInt(unitIdParam, 10, 64)
		if err != nil || unitId == 0 {
			logCtx.WithError(err).Error("Query failed. Invalid DashboardUnitID.")
			return nil, http.StatusBadRequest, INVALID_INPUT, "Query failed. Invalid DashboardUnitID.", true
		}
	}

	queryIdString := c.Query("query_id")
	if queryIdString == "" {
		var hasFailed bool
		var errMsg string
		hasFailed, errMsg, requestPayload = decodeAttributionPayload(r, logCtx)
		if hasFailed {
			logCtx.Error("Query failed. Json decode failed." + errMsg)
			return nil, http.StatusBadRequest, INVALID_INPUT, "Query failed. Json decode failed." + errMsg, true
		}
	} else {
		_, query, err := store.GetStore().GetQueryAndClassFromQueryIdString(queryIdString, projectId)
		if err != "" {
			logCtx.Error(fmt.Sprintf("Query from queryIdString failed - %v", err))
			return nil, http.StatusBadRequest, INVALID_INPUT, "Query failed. Json decode failed.", true
		}
		var requestPayloadUnit model.AttributionQueryUnit
		U.DecodePostgresJsonbToStructType(&query.Query, &requestPayloadUnit)
		requestPayload.Query = requestPayloadUnit.Query
	}

	enrichRequestUsingAttributionConfig(c, projectId, &requestPayload, logCtx)

	if requestPayload.Query.Timezone != "" {
		_, errCode := time.LoadLocation(requestPayload.Query.Timezone)
		if errCode != nil {
			return nil, http.StatusBadRequest, INVALID_INPUT, "Query failed. Invalid Timezone provided.", true
		}
		timezoneString = U.TimeZoneString(requestPayload.Query.Timezone)
	} else {
		timezoneString, statusCode = store.GetStore().GetTimezoneForProject(projectId)
		if statusCode != http.StatusFound {
			logCtx.WithError(err).Error("Query failed. Failed to get Timezone.")
			return nil, http.StatusBadRequest, INVALID_INPUT, "Query failed. Failed to get Timezone.", true
		}

		// For a KPI query, set the timezone internally for correct execution
		if requestPayload.Query.AnalyzeType == model.AnalyzeTypeHSDeals || requestPayload.Query.AnalyzeType == model.AnalyzeTypeSFOpportunities ||
			requestPayload.Query.AnalyzeType == model.AnalyzeTypeSFAccounts || requestPayload.Query.AnalyzeType == model.AnalyzeTypeHSCompanies {
			if requestPayload.Query.KPI.Queries[0].Timezone != "" {
				_, errCode := time.LoadLocation(string(requestPayload.Query.KPI.Queries[0].Timezone))
				if errCode != nil {
					return nil, http.StatusBadRequest, INVALID_INPUT, "Query failed. Invalid Timezone provided.", true
				}

				timezoneString = U.TimeZoneString(requestPayload.Query.KPI.Queries[0].Timezone)
			} else {
				timezoneString, statusCode = store.GetStore().GetTimezoneForProject(projectId)
				if statusCode != http.StatusFound {
					logCtx.Error("Query failed. Failed to get Timezone.")
					return nil, http.StatusBadRequest, INVALID_INPUT, "Query failed. Failed to get Timezone.", true
				}
				// logCtx.WithError(err).Error("Query failed. Invalid Timezone.")
			}

			requestPayload.Query.KPI.SetTimeZone(timezoneString)
			err = requestPayload.Query.KPI.TransformDateTypeFilters()
			if err != nil {
				return nil, http.StatusBadRequest, INVALID_INPUT, err.Error(), true
			}
		}
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
			return nil, http.StatusBadRequest, INVALID_INPUT, "Query time range is not valid for attribution.", true
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
		return nil, http.StatusBadRequest, INVALID_INPUT, err.Error(), true
	}
	if !hardRefresh {
		shouldReturn, resCode, resMsg := H.GetResponseIfCachedQuery(c, projectId, &attributionQueryUnitPayload, cacheResult, isDashboardQueryRequest, reqId, false)
		if shouldReturn {
			if resCode == http.StatusOK {
				return resMsg, resCode, "", "", false
			}
			logCtx.WithError(err).Error("Query failed. Error Processing/Fetching data from Query cache")
			return nil, resCode, PROCESSING_FAILED, "Error Processing/Fetching data from Query cache", true
		}
	}
	/*if isDashboardQueryRequest && C.DisableDashboardQueryDBExecution() && !isQuery {
		logCtx.WithField("request_payload", requestPayload).Warn("Skip hitting db for queries from dashboard, if not found on cache.")
		return nil, resCode, PROCESSING_FAILED, "Not found in cache. Execution suspended temporarily.", true
	}*/

	// If not found, set a placeholder for the query hash key that it has been running to avoid running again.
	model.SetQueryCachePlaceholder(projectId, &attributionQueryUnitPayload)

	enableOptimisedFilterOnProfileQuery := c.Request.Header.Get(H.HeaderUserFilterOptForProfiles) == "true" ||
		C.EnableOptimisedFilterOnProfileQuery()

	enableOptimisedFilterOnEventUserQuery := c.Request.Header.Get(H.HeaderUserFilterOptForEventsAndUsers) == "true" ||
		C.EnableOptimisedFilterOnEventUserQuery()

	H.SleepIfHeaderSet(c)
	QueryKey, _ := attributionQueryUnitPayload.GetQueryCacheRedisKey(projectId)
	debugQueryKey := model.GetStringKeyFromCacheRedisKey(QueryKey)
	var result *model.QueryResult

	result, err = store.GetStore().ExecuteAttributionQueryV1(projectId, requestPayload.Query, debugQueryKey,
		enableOptimisedFilterOnProfileQuery, enableOptimisedFilterOnEventUserQuery)

	if err != nil {
		model.DeleteQueryCacheKey(projectId, &attributionQueryUnitPayload)
		logCtx.WithError(err).Error("Failed to process query from DB")
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
		if C.IsLastComputedWhitelisted(projectId) {
			model.SetCacheResultByDashboardIdAndUnitIdWithPreset(result, projectId, dashboardId, unitId, preset,
				requestPayload.Query.From, requestPayload.Query.To, timezoneString, meta)
		} else {
			model.SetCacheResultByDashboardIdAndUnitId(result, projectId, dashboardId, unitId,
				requestPayload.Query.From, requestPayload.Query.To, timezoneString, meta)
		}

		return H.DashboardQueryResponsePayload{Result: result, Cache: false, RefreshedAt: U.TimeNowIn(U.TimeZoneStringIST).Unix(), CacheMeta: meta}, http.StatusOK, "", "", false
	}
	result.Query = requestPayload.Query
	return result, http.StatusOK, "", "", false
}

func enrichRequestUsingAttributionConfig(c *gin.Context, projectID int64, requestPayload *AttributionRequestPayload, logCtx *log.Entry) {

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

	if requestPayload.Query.AnalyzeType == model.AnalyzeTypeHSDeals {

		if &attributionConfig != nil && attributionConfig.AnalyzeTypeHSCompaniesEnabled == true {
			requestPayload.Query.AnalyzeType = model.AnalyzeTypeHSCompanies
		} else if &attributionConfig != nil && attributionConfig.AnalyzeTypeHSDealsEnabled == true {
			requestPayload.Query.AnalyzeType = model.AnalyzeTypeHSDeals
		} else {
			logCtx.WithFields(log.Fields{"Query": requestPayload.Query, "AttributionConfig": attributionConfig}).Error("Failed to set analyze type")
			c.AbortWithStatusJSON(errCode, gin.H{"error": "Invalid config/query. Failed to set analyze type from attribution config & project settings."})
		}
	}
	if requestPayload.Query.AnalyzeType == model.AnalyzeTypeSFOpportunities {

		if &attributionConfig != nil && attributionConfig.AnalyzeTypeSFAccountsEnabled == true {
			requestPayload.Query.AnalyzeType = model.AnalyzeTypeSFAccounts
		} else if &attributionConfig != nil && attributionConfig.AnalyzeTypeSFOpportunitiesEnabled == true {
			requestPayload.Query.AnalyzeType = model.AnalyzeTypeSFOpportunities
		} else {
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
