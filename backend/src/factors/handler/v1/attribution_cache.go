package v1

import (
	"encoding/json"
	"errors"
	C "factors/config"
	H "factors/handler/helpers"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func RunMultipleRangeAttributionQueries(projectId, dashboardId, unitId int64, requestPayload AttributionRequestPayloadV1,
	timezoneString U.TimeZoneString, reqId string, enableOptimisedFilterOnProfileQuery, enableOptimisedFilterOnEventUserQuery bool,
	rangesToRun []U.TimestampRange, logCtx *log.Entry) (bool, *model.QueryResult, []H.ComputedRangeInfo, error) { //hasFailed, Result, computeMeta, error

	var latestFoundResult *model.QueryResult
	var mergedResult *model.QueryResult
	mergedResult = nil
	var err error

	var computedMeta []H.ComputedRangeInfo

	var query *model.AttributionQueryV1
	// building a query with 10 secs range to get GetRawAttributionQueryParams faster
	if len(rangesToRun) >= 1 {
		U.DeepCopy(requestPayload.Query, &query)
		query.From = rangesToRun[0].Start
		query.To = rangesToRun[0].Start + 10
	} else {
		logCtx.Error("Can't run as date range is not correct")
		return true, mergedResult, computedMeta, errors.New("can't run as date range is not correct")
	}

	// Get the basic parameters for merging
	_, kpiAggFunctionType, errKpi := store.GetStore().GetRawAttributionQueryParams(projectId, query,
		enableOptimisedFilterOnProfileQuery, enableOptimisedFilterOnEventUserQuery)
	logCtx.WithFields(log.Fields{"kpiAggFunctionType": kpiAggFunctionType}).Info("GetRawAttributionQueryParams for the query merge")

	if errKpi != nil {
		logCtx.WithError(errKpi).Error("Error occurred during fetching merge params of attribution GetRawAttributionQueryParams")
		return true, mergedResult, computedMeta, errKpi
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
				return true, mergedResult, computedMeta, err
			}
			// this will update the last cached result as qRange ranges are in descending order
			latestFoundResult = resultForRange
			computedM := H.ComputedRangeInfo{From: qRange.Start, To: qRange.End, TimeZone: string(timezoneString), FromCache: true}
			computedMeta = append(computedMeta, computedM)
		} else {

			// Not allowing query: This is not a failure but due to some reason if the result is not cached in the DB
			// we want to avoid computing and instead throw "No Data Found" error

			logCtx.Info("Failing the query as the all parts of the query was not found in DB - attribution v1")
			return true, mergedResult, computedMeta, errors.New("no Data found")

			// compute if not found in cache
			/*
				requestPayload.Query.From = qRange.Start
				requestPayload.Query.To = qRange.End
				attributionQueryUnitPayload := model.AttributionQueryUnitV1{
					Class: model.QueryClassAttribution,
					Query: requestPayload.Query,
				}
				QueryKey, _ := attributionQueryUnitPayload.GetQueryCacheRedisKey(projectId)
				debugQueryKey := model.GetStringKeyFromCacheRedisKey(QueryKey)
				resultForRange, err = store.GetStore().ExecuteAttributionQueryV1(projectId, requestPayload.Query, debugQueryKey,
					enableOptimisedFilterOnProfileQuery, enableOptimisedFilterOnEventUserQuery, unitId)
				logCtx.WithError(err).WithFields(log.Fields{"RIndex": idx, "RStart": qRange.Start, "REnd": qRange.End}).Info("Found there ExecuteAttributionQueryV1")
				if err != nil {
					logCtx.Info("Failed to process query when not found in  DB - attribution v1")
					return true, mergedResult, computedMeta
				}
				computedM := H.ComputedRangeInfo{From: qRange.Start, To: qRange.End, TimeZone: string(timezoneString), FromCache: false}
				computedMeta = append(computedMeta, computedM)
			*/
		}
		keyIndex := model.GetLastKeyValueIndex(resultForRange.Headers)
		if requestPayload.Query.AttributionKey == model.AttributionKeyLandingPage ||
			requestPayload.Query.AttributionKey == model.AttributionKeyChannel ||
			requestPayload.Query.AttributionKey == model.AttributionKeySource ||
			requestPayload.Query.AttributionKey == model.AttributionKeyAllPageView {
			keyIndex = model.GetLastKeyValueIndexLandingPage(resultForRange.Headers)
		}

		logCtx.WithFields(log.Fields{
			"mergedResult":   mergedResult,
			"resultForRange": resultForRange,
		}).Info("before MergeTwoAttributionReportsIntoOne")

		// Now we have the result either from cache or computed
		mergedResult = model.MergeTwoAttributionReportsIntoOne(mergedResult, resultForRange,
			keyIndex, requestPayload.Query.AttributionKey,
			kpiAggFunctionType, *logCtx)

		logCtx.WithFields(log.Fields{
			"mergedResult":   mergedResult,
			"resultForRange": resultForRange,
		}).Info("after MergeTwoAttributionReportsIntoOne")

		if mergedResult == nil {
			logCtx.Info("Failed to process query from DB - attribution v1 as mergedResult is nil")
			return true, mergedResult, computedMeta, errors.New("the final result is empty, no data to found")
		}
	}
	if latestFoundResult != nil {
		mergedResult.CacheMeta = latestFoundResult.CacheMeta
	}
	return false, mergedResult, computedMeta, nil
}

func RunAttributionQuery(projectId int64, requestPayload AttributionRequestPayloadV1, debugQueryKey string,
	enableOptimisedFilterOnProfileQuery bool, enableOptimisedFilterOnEventUserQuery bool, attributionQueryUnitPayload model.AttributionQueryUnitV1,
	logCtx *log.Entry, timezoneString U.TimeZoneString, preset string, isDashboardQueryRequest bool,
	dashboardId int64, unitId int64) (interface{}, int, string, string, bool) {

	var err error
	var result *model.QueryResult
	result, err = store.GetStore().ExecuteAttributionQueryV1(projectId, requestPayload.Query, debugQueryKey,
		enableOptimisedFilterOnProfileQuery, enableOptimisedFilterOnEventUserQuery, unitId)

	if err != nil {
		model.DeleteQueryCacheKey(projectId, &attributionQueryUnitPayload)
		logCtx.Info("Failed to process query from DB - attribution v1", err.Error())
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
				requestPayload.Query.From, requestPayload.Query.To, timezoneString, meta, false)
		}

		return H.DashboardQueryResponsePayload{Result: result, Cache: false, RefreshedAt: U.TimeNowIn(U.TimeZoneStringIST).Unix(), CacheMeta: meta}, http.StatusOK, "", "", false
	}
	result.Query = requestPayload.Query
	return result, http.StatusOK, "", "", false
}
