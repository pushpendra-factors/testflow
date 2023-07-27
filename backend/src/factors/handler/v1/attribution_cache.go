package v1

import (
	"encoding/json"
	C "factors/config"
	H "factors/handler/helpers"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

func RunMultipleRangeAttributionQueries(projectId, dashboardId, unitId int64, requestPayload AttributionRequestPayloadV1,
	timezoneString U.TimeZoneString, reqId string, enableOptimisedFilterOnProfileQuery, enableOptimisedFilterOnEventUserQuery bool,
	rangesToRun []U.TimestampRange, logCtx *log.Entry) (bool, *model.QueryResult) { //hasFailed, Result

	var latestFoundResult *model.QueryResult
	var mergedResult *model.QueryResult
	mergedResult = nil
	var err error

	// Get the basic parameters for merging
	_, kpiAggFunctionType, errKpi := store.GetStore().GetRawAttributionQueryParams(projectId, requestPayload.Query,
		enableOptimisedFilterOnProfileQuery, enableOptimisedFilterOnEventUserQuery)
	logCtx.WithFields(log.Fields{"kpiAggFunctionType": kpiAggFunctionType}).Info("GetRawAttributionQueryParams for the query merge")

	if errKpi != nil {
		logCtx.WithError(errKpi).Error("Error occurred during fetching merge params of attribution GetRawAttributionQueryParams")
		return true, mergedResult
	}

	logCtx.WithFields(log.Fields{"Ranges": rangesToRun}).Info("Ranges to run the query")
	// fetch or compute result for each qRange
	for idx, qRange := range rangesToRun {

		var resultForRange *model.QueryResult
		// check if result for this qRange is cached
		errCode, cacheResult := store.GetStore().FetchCachedResultFromDataBase(reqId, projectId, dashboardId, unitId,
			qRange.Start, qRange.End)
		if errCode == http.StatusFound && cacheResult.Result != nil {
			logCtx.WithFields(log.Fields{"RIndex": idx, "RStart": qRange.Start, "REnd": qRange.End}).Info("Found there FetchCachedResultFromDataBase")
			// Unmarshal the byte result into the QueryResult struct
			err = json.Unmarshal(cacheResult.Result, &resultForRange)
			if err != nil {
				logCtx.WithError(err).Error("Error occurred during unmarshal of attribution cached report")
				return true, mergedResult
			}
			// this will update the last cached result as qRange ranges are in descending order
			latestFoundResult = resultForRange
		} else {
			// compute if not found in cache
			requestPayload.Query.From = qRange.Start
			requestPayload.Query.To = qRange.End
			attributionQueryUnitPayload := model.AttributionQueryUnitV1{
				Class: model.QueryClassAttribution,
				Query: requestPayload.Query,
			}
			QueryKey, _ := attributionQueryUnitPayload.GetQueryCacheRedisKey(projectId)
			debugQueryKey := model.GetStringKeyFromCacheRedisKey(QueryKey)
			resultForRange, err = store.GetStore().ExecuteAttributionQueryV1(projectId, requestPayload.Query, debugQueryKey,
				enableOptimisedFilterOnProfileQuery, enableOptimisedFilterOnEventUserQuery)
			logCtx.WithError(err).WithFields(log.Fields{"RIndex": idx, "RStart": qRange.Start, "REnd": qRange.End}).Info("Found there ExecuteAttributionQueryV1")
			if err != nil {
				logCtx.Info("Failed to process query when not found in  DB - attribution v1")
				return true, mergedResult
			}
		}
		keyIndex := model.GetLastKeyValueIndex(mergedResult.Headers)
		if requestPayload.Query.AttributionKey == model.AttributionKeyLandingPage ||
			requestPayload.Query.AttributionKey == model.AttributionKeyChannel ||
			requestPayload.Query.AttributionKey == model.AttributionKeySource ||
			requestPayload.Query.AttributionKey == model.AttributionKeyAllPageView {
			keyIndex = model.GetLastKeyValueIndexLandingPage(mergedResult.Headers)
		}

		// Now we have the result either from cache or computed
		mergedResult = model.MergeTwoAttributionReportsIntoOne(mergedResult, resultForRange,
			keyIndex, requestPayload.Query.AttributionKey,
			kpiAggFunctionType, *logCtx)
		if err != nil {
			logCtx.Info("Failed to process query from DB - attribution v1", err.Error())
			return true, mergedResult
		}
	}
	mergedResult.CacheMeta = latestFoundResult.CacheMeta
	return false, mergedResult
}

func IsAMonthlyRangeQuery(timezoneString U.TimeZoneString, effectiveFrom, effectiveTo int64) (bool, []U.TimestampRange) {

	last12Months := U.GenerateLast12MonthsTimestamps(string(timezoneString))
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
		return true, last12Months
	}
	return false, last12Months
}

func IsAWeeklyRangeQuery(timezoneString U.TimeZoneString, effectiveFrom, effectiveTo int64) (bool, []U.TimestampRange) {

	last48Weeks := U.GenerateLast12MonthsTimestamps(string(timezoneString))
	stMatch := 0
	enMatch := 0
	for _, rng := range last48Weeks {
		if rng.Start == effectiveFrom {
			stMatch = 1
		}
		if rng.End == effectiveTo {
			enMatch = 1
		}
	}
	// both start and end timestamp belong to a month
	if stMatch == 1 && enMatch == 1 {
		return true, last48Weeks
	}
	return false, last48Weeks
}

// GenerateLast48WeeksTimestamps returns start-end of last 48 weeks in descending order (last week first) from now
func GenerateLast48WeeksTimestamps(timezone string) []U.TimestampRange {
	timestamps := make([]U.TimestampRange, 0)

	loc, err := time.LoadLocation(timezone)
	if err != nil {
		fmt.Println("Invalid timezone:", timezone)
		return timestamps
	}

	now := time.Now().In(loc)

	// Set the weekday to Sunday (0) and adjust the time to 00:00:00
	weekStart := now.AddDate(0, 0, -(int(now.Weekday())))
	weekStart = time.Date(weekStart.Year(), weekStart.Month(), weekStart.Day(), 0, 0, 0, 0, loc)

	for i := 0; i < 48; i++ {
		// Get the end of the week by adding 6 days, 23 hours, 59 minutes, and 59 seconds
		weekEnd := weekStart.AddDate(0, 0, 6).Add(time.Hour*23 + time.Minute*59 + time.Second*59)

		// add all the valid ranges which are smaller than today's time
		if weekStart.Unix() < now.Unix() && weekEnd.Unix() < now.Unix() {
			timestamps = append(timestamps, U.TimestampRange{
				Start: weekStart.Unix(),
				End:   weekEnd.Unix(),
			})
		}
		// Move to the previous week's start
		weekStart = weekStart.AddDate(0, 0, -7)
	}

	return timestamps
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
