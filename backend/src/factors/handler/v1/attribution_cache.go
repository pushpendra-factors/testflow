package v1

import (
	C "factors/config"
	H "factors/handler/helpers"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	log "github.com/sirupsen/logrus"
	"net/http"
)

/*
func CheckForMonthlyCacheComputation(requestPayload AttributionRequestPayloadV1, timezoneString U.TimeZoneString,
	effectiveFrom int64, effectiveTo int64, shouldReturn bool, resCode int, reqId string, projectId int64,
	dashboardId int64, unitId int64, enableOptimisedFilterOnProfileQuery, enableOptimisedFilterOnEventUserQuery bool,
	logCtx *log.Entry) (bool, int, interface{}, int, string, string, bool, bool) {

	// 1. Check if query from and to is month start, end respectively
	last12Months := U.GenerateLast12MonthsTimestamps(string(timezoneString))
	var queryRangeIsInMonths bool
	queryRangeIsInMonths = isAMonthlyRangeQuery(last12Months, effectiveFrom, effectiveTo)
	if !queryRangeIsInMonths {
		// todo
		return
	}

	monthsToRun := U.GetAllMonthsInBetween(effectiveFrom, effectiveTo, last12Months)
	var resultMerged *model.QueryResult
	var forCacheMetaData model.DashQueryResult
	resultMerged = nil

	// fetch or compute result for each month
	for idx, month := range monthsToRun {

		var resultForMonth *model.QueryResult
		// check if result for this month is cached
		errCode, cacheResult := store.GetStore().FetchCachedResultFromDataBase(reqId, projectId, dashboardId, unitId,
			month.Start, month.End)
		if errCode == http.StatusFound && cacheResult.Result != nil {
			// Unmarshal the byte result into the QueryResult struct
			err := json.Unmarshal(cacheResult.Result, &resultForMonth)
			if err != nil {
				logCtx.Error("Error occurred during unmarshal of attribution cached report")
				// todo return nil
			}
			// this will update the last cached result as month ranges are in decending order
			forCacheMetaData = cacheResult
		} else {
			// compute if not found in cache
			// query, err := store.GetStore().GetQueryWithDashboardUnitIdString(projectId, unitId)
			requestPayload.Query.From = month.Start
			requestPayload.Query.To = month.End
			attributionQueryUnitPayload := model.AttributionQueryUnitV1{
				Class: model.QueryClassAttribution,
				Query: requestPayload.Query,
			}
			QueryKey, _ := attributionQueryUnitPayload.GetQueryCacheRedisKey(projectId)
			debugQueryKey := model.GetStringKeyFromCacheRedisKey(QueryKey)
			resultForMonth, errCode1 := store.GetStore().ExecuteAttributionQueryV1(projectId, requestPayload.Query, debugQueryKey,
			enableOptimisedFilterOnProfileQuery, enableOptimisedFilterOnEventUserQuery)
			if errCode1 != nil {
				logCtx.Info("Failed to process query from DB - attribution v1", errCode1.Error())
				// todo return
			}
		}

		// Now we have the result either from cache or computed
		if idx == 0 {
			resultMerged = resultForMonth
		} else {
			// resultMerged = model.MergeQueryResults(resultMerged, resultForMonth)
		}
	}
	return shouldReturn, resCode, nil, 0, "", "", false, false
}
*/

func isAMonthlyRangeQuery(last12Months []U.TimestampRange, effectiveFrom int64, effectiveTo int64) bool {
	stMatch := 0
	enMatch := 0
	for _, rng := range last12Months {
		if rng.Start == effectiveFrom {
			stMatch = 1
		}
		if rng.End == effectiveTo {
			enMatch = 1
		}
	}
	// both start and end timestamp belong to a month
	if stMatch == 1 && enMatch == 1 {
		return true
	}
	return false
}

func RunAttributionQuery(projectId int64, requestPayload AttributionRequestPayloadV1, debugQueryKey string,
	enableOptimisedFilterOnProfileQuery bool, enableOptimisedFilterOnEventUserQuery bool, attributionQueryUnitPayload model.AttributionQueryUnitV1,
	logCtx *log.Entry, timezoneString U.TimeZoneString, preset string, isDashboardQueryRequest bool,
	dashboardId int64, unitId int64) (interface{}, int, string, string, bool) {

	var err error
	var result *model.QueryResult
	result, err = store.GetStore().ExecuteAttributionQueryV1(projectId, requestPayload.Query, debugQueryKey,
		enableOptimisedFilterOnProfileQuery, enableOptimisedFilterOnEventUserQuery)

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
