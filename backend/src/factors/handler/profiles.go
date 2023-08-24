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

func ProfilesQueryHandler(c *gin.Context) (interface{}, int, string, string, bool) {
	if C.GetConfig().PrimaryDatastore != C.DatastoreTypeMemSQL {
		return nil, http.StatusUnauthorized, V1.PROCESSING_FAILED, "Query failed. Query only allowed for memSQl.", true
	}
	var timezoneString U.TimeZoneString
	var err error
	hardRefresh := false
	refreshParam := c.Query("refresh")
	var dashboardId int64
	var unitId int64

	reqID := U.GetScopeByKeyAsString(c, mid.SCOPE_REQ_ID)
	logCtx := log.WithFields(log.Fields{
		"reqId": reqID,
	})

	if refreshParam != "" {
		hardRefresh, _ = strconv.ParseBool(refreshParam)
	}
	/*isQuery := false
	isQueryParam := c.Query("is_query")
	if isQueryParam != "" {
		isQuery, _ = strconv.ParseBool(isQueryParam)
	}*/

	r := c.Request

	projectID := U.GetScopeByKeyAsInt64(c, mid.SCOPE_PROJECT_ID)
	if projectID == 0 {
		logCtx.Error("Query failed. Invalid project.")
		return nil, http.StatusUnauthorized, V1.INVALID_PROJECT, "Query failed. Invalid project.", true
	}

	logCtx = log.WithFields(log.Fields{
		"reqId":     reqID,
		"projectID": projectID,
	})

	profileQueryGroup, dashboardId, unitId, isDashboardQueryRequest, statusCode, errorCode, errMsg, isErr := getValidProfilesQueryAndDetailsFromRequest(r, c, logCtx, projectID)
	if statusCode != http.StatusOK {
		return nil, statusCode, errorCode, errMsg, isErr
	}

	profileQueryGroup, statusCode, errorCode, errMsg, isErr = setTimezoneForProfilesRequest(logCtx, profileQueryGroup, projectID)
	if statusCode != http.StatusOK {
		return nil, statusCode, errorCode, errMsg, isErr
	}
	timezoneString = profileQueryGroup.GetTimeZone()

	if profileQueryGroup.From == 0 || profileQueryGroup.To == 0 {
		logCtx.WithError(err).Error("Query failed. Invalid date range provided.")
		return nil, http.StatusBadRequest, V1.INVALID_INPUT, "Query failed. Invalid date range provided.", true
	}

	allowSupportForSourceColumnInUsers := C.IsProfileQuerySourceSupported(projectID)
	allowProfilesGroupSupport := C.IsProfileGroupSupportEnabled(projectID)

	// copying global filters and groupby into sparate queries and datetime transformations
	for index, _ := range profileQueryGroup.Queries {
		profileQueryGroup.Queries[index].Filters = append(profileQueryGroup.Queries[index].Filters, profileQueryGroup.GlobalFilters...)
		profileQueryGroup.Queries[index].GroupBys = append(profileQueryGroup.Queries[index].GroupBys, profileQueryGroup.GlobalGroupBys...)

		// passing date range
		profileQueryGroup.Queries[index].From = profileQueryGroup.From
		profileQueryGroup.Queries[index].To = profileQueryGroup.To

		if allowSupportForSourceColumnInUsers && !model.IsValidUserSource(profileQueryGroup.Queries[index].Type) {
			logCtx.WithError(err).Error("Query failed. Invalid user source.")
			message := fmt.Sprintf("Query failed. Invalid user source provided : %s", profileQueryGroup.Queries[index].Type)
			return nil, http.StatusBadRequest, V1.INVALID_INPUT, message, true
		}

		if allowProfilesGroupSupport {
			if !model.IsValidProfileQueryGroupName(profileQueryGroup.GroupAnalysis) {
				logCtx.WithError(err).Error("Query failed. Invalid group name.")
				message := fmt.Sprintf("Query failed. Invalid group name provided : %s", profileQueryGroup.GroupAnalysis)
				return nil, http.StatusBadRequest, V1.INVALID_INPUT, message, true
			} else {
				profileQueryGroup.Queries[index].GroupAnalysis = profileQueryGroup.GroupAnalysis
			}
		}

		// setting up the timezone for individual queries from the global value
		profileQueryGroup.Queries[index].SetTimeZone(timezoneString)
		err = profileQueryGroup.Queries[index].TransformDateTypeFilters()
		if err != nil {
			return nil, http.StatusBadRequest, V1.INVALID_INPUT, err.Error(), true
		}
		// setting granularity for datetime filters
		for indexGroupBy := range profileQueryGroup.Queries[index].GroupBys {
			profileQueryGroup.Queries[index].GroupBys[indexGroupBy].Index = indexGroupBy
			if profileQueryGroup.Queries[index].GroupBys[indexGroupBy].Type == U.PropertyTypeDateTime &&
				profileQueryGroup.Queries[index].GroupBys[indexGroupBy].Granularity == "" {
				profileQueryGroup.Queries[index].GroupBys[indexGroupBy].Granularity = U.DateTimeBreakdownDailyGranularity
			}
		}
	}

	// Tracking dashboard query request.
	if isDashboardQueryRequest {
		model.SetDashboardCacheAnalytics(projectID, dashboardId, unitId, profileQueryGroup.From, profileQueryGroup.To, timezoneString)
	}

	allowSyncReferenceFields := C.AllowSyncReferenceFields(projectID)

	if isDashboardQueryRequest && !H.ShouldAllowHardRefresh(0, 0, timezoneString, hardRefresh) {
		shouldReturn, resCode, resMsg := H.GetResponseIfCachedDashboardQuery(
			reqID, projectID, dashboardId, unitId, 0, 0, timezoneString)
		if shouldReturn {
			if resCode == http.StatusOK {
				if allowSyncReferenceFields && resMsg != nil {
					resMsg, err = H.TransformQueryCacheResponseColumnValuesToLabel(projectID, resMsg)
					if err != nil {
						logCtx.WithError(err).Error("Failed to set property value label.")
					}
				}
				return resMsg, resCode, "", "", false
			}
		}
	}

	var cacheResult model.ResultGroup
	shouldReturn, resCode, resMsg := H.GetResponseIfCachedQuery(c, projectID, &profileQueryGroup, cacheResult, false, reqID, false)
	if shouldReturn {
		if resCode == http.StatusOK {
			if allowSyncReferenceFields && resMsg != nil {
				resMsg, err = H.TransformQueryCacheResponseColumnValuesToLabel(projectID, resMsg)
				if err != nil {
					logCtx.WithError(err).Error("Failed to set property value label.")
				}
			}
			return gin.H{"result": resMsg}, resCode, "", "", false
		}
		logCtx.Error("Query failed. Error Processing/Fetching data from Query cache")
		return nil, resCode, V1.PROCESSING_FAILED, "Error Processing/Fetching data from Query cache", true
	}

	/*if isDashboardQueryRequest && C.DisableDashboardQueryDBExecution() && !isQuery {
		logCtx.WithField("request_payload", profileQueryGroup).Warn("Skip hitting db for queries from dashboard, if not found on cache.")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "Query failed. Not found in cache. Suspended db execution."})
	}*/

	// Use optimised filter for profiles query if enabled using header or configuration.
	enableOptimisedFilter := c.Request.Header.Get(H.HeaderUserFilterOptForProfiles) == "true" ||
		C.EnableOptimisedFilterOnProfileQuery()

	model.SetQueryCachePlaceholder(projectID, &profileQueryGroup)
	H.SleepIfHeaderSet(c)
	resultGroup, errCode := store.GetStore().RunProfilesGroupQuery(profileQueryGroup.Queries, projectID, enableOptimisedFilter)
	if errCode != http.StatusOK {
		model.DeleteQueryCacheKey(projectID, &profileQueryGroup)
		logCtx.Error("Profile Query failed. Failed to process query from DB")
		if errCode == http.StatusPartialContent {
			return resultGroup, errCode, V1.PROCESSING_FAILED, "Failed to process query from DB", true
		}
		return nil, errCode, V1.PROCESSING_FAILED, "Failed to process query from DB", true
	}
	model.SetQueryCacheResult(projectID, &profileQueryGroup, resultGroup)
	resultGroup.Query = profileQueryGroup

	if allowSyncReferenceFields {
		resultGroup.Results, err = store.GetStore().AddPropertyValueLabelToQueryResults(projectID, resultGroup.Results)
		if err != nil {
			logCtx.WithError(err).Error("Failed to set property value label.")
		}
	}

	return resultGroup, http.StatusOK, "", "", false
}

func getValidProfilesQueryAndDetailsFromRequest(r *http.Request, c *gin.Context, logCtx *log.Entry, projectId int64) (model.ProfileQueryGroup, int64, int64, bool, int, string, string, bool) {
	var dashboardId, unitId int64
	var err error
	requestPayload, queryPayload := model.ProfileQueryGroup{}, model.ProfileQueryGroup{}

	dashboardIdParam := c.Query("dashboard_id")
	unitIdParam := c.Query("dashboard_unit_id")
	queryIdString := c.Query("query_id")

	if queryIdString == "" {
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&requestPayload); err != nil {
			logCtx.WithError(err).Error("Query failed. Json decode failed.")
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
		U.DecodePostgresJsonbToStructType(&query.Query, &queryPayload)
	} else if queryIdString != "" {
		_, query, err := store.GetStore().GetQueryAndClassFromQueryIdString(queryIdString, projectId)
		if err != "" {
			logCtx.Error(fmt.Sprintf("Query from queryIdString failed - %v", err))
			return queryPayload, 0, 0, false, http.StatusBadRequest, V1.INVALID_INPUT, "Query failed. Json decode failed.", true
		}
		if query.LockedForCacheInvalidation {
			return queryPayload, 0, 0, false, http.StatusConflict, V1.PROCESSING_FAILED, "Query is not processed due to saved query updated", false
		}
		U.DecodePostgresJsonbToStructType(&query.Query, &queryPayload)
	} else {
		queryPayload = requestPayload
	}

	if queryIdString == "" {
		queryPayload.From = requestPayload.From
		queryPayload.To = requestPayload.To
		if requestPayload.Timezone != "" {
			queryPayload.SetTimeZone(U.TimeZoneString(requestPayload.Timezone))
		}
	}

	if len(requestPayload.Queries) == 0 {
		logCtx.Error("Query failed. Empty query group.")
		return queryPayload, dashboardId, unitId, isDashboardQueryRequest, http.StatusBadRequest, V1.INVALID_INPUT, "Query failed. Empty query group.", true
	}
	return queryPayload, dashboardId, unitId, isDashboardQueryRequest, http.StatusOK, "", "", false
}

func setTimezoneForProfilesRequest(logCtx *log.Entry, requestPayload model.ProfileQueryGroup, projectId int64) (model.ProfileQueryGroup, int, string, string, bool) {
	var timezoneString U.TimeZoneString
	if requestPayload.Queries[0].Timezone != "" {
		_, errCode := time.LoadLocation(string(requestPayload.Queries[0].Timezone))
		if errCode != nil {
			return requestPayload, http.StatusBadRequest, V1.INVALID_INPUT, "Query failed. Invalid Timezone provided.", true
		}
		timezoneString = U.TimeZoneString(requestPayload.Queries[0].Timezone)
	} else {
		var statusCode int
		timezoneString, statusCode = store.GetStore().GetTimezoneForProject(projectId)
		if statusCode != http.StatusFound {
			logCtx.Error("Query failed. Failed to get Timezone.")
			return requestPayload, http.StatusBadRequest, V1.INVALID_INPUT, "Query failed. Failed to get Timezone.", true
		}
	}
	requestPayload.SetTimeZone(timezoneString)
	return requestPayload, http.StatusOK, "", "", false
}
