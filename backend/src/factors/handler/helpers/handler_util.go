package helpers

import (
	"encoding/json"
	"errors"
	cacheRedis "factors/cache/redis"
	"factors/model/model"
	"factors/model/store"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"

	C "factors/config"
	U "factors/util"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

const HeaderUserFilterOptForProfiles string = "Use-Filter-Opt-Profiles"
const HeaderUserFilterOptForEventsAndUsers string = "Use-Filter-Opt-Events-Users"

// DashboardQueryResponsePayload Query response with cache and refreshed_at.
type DashboardQueryResponsePayload struct {
	Result      interface{} `json:"result"`
	Cache       bool        `json:"cache"`
	RefreshedAt int64       `json:"refreshed_at"`
	TimeZone    string      `json:"timezone"`
	CacheMeta   interface{} `json:"cache_meta"`
	ComputeMeta interface{} `json:"compute_meta"`
}

type ComputedRangeInfo struct {
	From      int64  `json:"from"`
	To        int64  `json:"to"`
	TimeZone  string `json:"timezone"`
	FromCache bool   `json:"from_cache"`
}

func getQueryCacheResponse(c *gin.Context, cacheResult model.QueryCacheResult, forDashboard bool, skipContextVerfication bool) (bool, int, interface{}) {
	if forDashboard {
		return true, http.StatusOK, DashboardQueryResponsePayload{Result: cacheResult.Result, Cache: true, RefreshedAt: cacheResult.RefreshedAt, TimeZone: cacheResult.TimeZone, CacheMeta: cacheResult.CacheMeta}
	}
	// To Indicate if the result is served from cache without changing the response format.
	if !skipContextVerfication {
		c.Header(model.QueryCacheResponseFromCacheHeader, "true")
		c.Header(model.QueryCacheResponseCacheRefreshedAt, fmt.Sprint(cacheResult.RefreshedAt))
		c.Header(model.QueryCacheResponseCacheTimeZone, fmt.Sprint(cacheResult.TimeZone))
	}
	return true, http.StatusOK, cacheResult.Result
}

func AddPropertyLabelsToQueryCacheInterfaceArrayResponse(projectID int64, recordsInt []interface{}) (interface{}, error) {
	if len(recordsInt) == 0 {
		return recordsInt, nil
	}

	newRecordsInt := make([]interface{}, 0)
	for i := range recordsInt {
		if recordsInt[i] == nil {
			continue
		}

		record, ok := recordsInt[i].(map[string]interface{})
		if !ok {
			return recordsInt, errors.New("Failed to convert record to map on AddPropertyLabelsToQueryCacheInterfaceArrayResponse")
		}

		var err error
		record, err = store.GetStore().TransformQueryResultsColumnValuesToLabel(projectID, record)
		if err != nil {
			return recordsInt, err
		}

		newRecordsInt = append(newRecordsInt, record)
	}
	return newRecordsInt, nil
}

func AddPropertyLabelsToQueryCacheResultGroupResponse(projectID int64, record map[string]interface{}) (interface{}, error) {
	if _, exists := record["result_group"]; exists {
		var resultGroup model.ResultGroup
		err := U.DecodeInterfaceMapToStructType(record, &resultGroup)
		if err != nil {
			return record, errors.New("Failed to decode interface to ResultGroup on AddPropertyLabelsToQueryCacheResultGroupResponse")
		}
		// log.WithField("result_group_record", resultGroup).Warning("ResultGroup record in AddPropertyLabelsToQueryCacheResultGroupResponse")

		resultGroup.Results, err = store.GetStore().AddPropertyValueLabelToQueryResults(projectID, resultGroup.Results)
		if err != nil {
			return record, err
		}
		// log.WithField("result_group_record_after", resultGroup).Warning("ResultGroup after record in AddPropertyLabelsToQueryCacheResultGroupResponse")
		result, err := U.EncodeStructTypeToMap(resultGroup)
		if err != nil {
			return record, err
		}
		return result, nil
	}

	return record, nil
}

func AddPropertyLabelsToDashboardQueryResponsePayload(projectID int64, record DashboardQueryResponsePayload) (DashboardQueryResponsePayload, error) {
	var err error
	switch result := record.Result.(type) {
	case []interface{}:
		record.Result, err = AddPropertyLabelsToQueryCacheInterfaceArrayResponse(projectID, result)
		return record, err
	case map[string]interface{}:
		record.Result, err = AddPropertyLabelsToQueryCacheResultGroupResponse(projectID, result)
		return record, err
	default:
		return record, errors.New("invalid record type on AddPropertyLabelsToDashboardQueryResponsePayload")
	}
}

func TransformQueryCacheResponseColumnValuesToLabel(projectID int64, recordsInt interface{}) (interface{}, error) {
	switch records := recordsInt.(type) {
	case DashboardQueryResponsePayload:
		return AddPropertyLabelsToDashboardQueryResponsePayload(projectID, records)
	case []interface{}:
		return AddPropertyLabelsToQueryCacheInterfaceArrayResponse(projectID, records)
	case map[string]interface{}:
		return AddPropertyLabelsToQueryCacheResultGroupResponse(projectID, records)
	default:
		return nil, errors.New("invalid record type on TransformQueryCacheResponseColumnValuesToLabel")
	}
}

// ShouldAllowHardRefresh To check from query api if hard refresh should be applied or return from cache.
func ShouldAllowHardRefresh(from, to int64, timezoneString U.TimeZoneString, hardRefresh bool) bool {
	if C.DisableQueryCache() {
		// Always do hard refresh if configured.
		return true
	}
	return ((U.IsStartOfTodaysRangeIn(from, timezoneString) || U.Is30MinutesTimeRange(from, to)) && hardRefresh)
}

// SleepIfHeaderSet Sleep in request handler if header set. Currently used in testing.
func SleepIfHeaderSet(c *gin.Context) {
	if C.GetConfig().Env != C.DEVELOPMENT {
		// Sleep header only being used in development to facilitate testing.
		return
	}
	if waitTime := c.Request.Header.Get(model.QueryCacheRequestSleepHeader); waitTime != "" {
		waitTimeSeconds, err := strconv.Atoi(waitTime)
		if err == nil {
			time.Sleep(time.Duration(waitTimeSeconds) * time.Second)
		}
	}
}

// GetResponseIfCachedQuery Returns response for the query is cached.
func GetResponseIfCachedQuery(c *gin.Context, projectID int64, requestPayload model.BaseQuery,
	resultContainer interface{}, forDashboard bool, reqID string, skipContextVerfication bool) (bool, int, interface{}) {
	if C.DisableQueryCache() {
		return false, http.StatusNotFound, nil
	}

	if !skipContextVerfication {
		if c.Request.Header.Get(model.QueryCacheRequestInvalidatedCacheHeader) == "true" {
			model.DeleteQueryCacheKey(projectID, requestPayload)
			return false, http.StatusNotFound, nil
		}
	}

	cacheKey, _ := requestPayload.GetQueryCacheRedisKey(projectID)
	cacheKeyString, _ := cacheKey.Key()
	log.WithField("req_id", reqID).WithField("key", cacheKeyString).Info("Query cache key")

	cacheResult, errCode := model.GetQueryResultFromCache(projectID, requestPayload, &resultContainer)
	if errCode == http.StatusFound {
		return getQueryCacheResponse(c, cacheResult, forDashboard, skipContextVerfication)
	} else if errCode == http.StatusAccepted {
		// An instance of query is in progress. Poll for result.
		for {
			if C.GetConfig().Env == C.DEVELOPMENT {
				time.Sleep(10 * time.Millisecond)
			} else {
				time.Sleep(5 * time.Second)
			}
			cacheResult, errCode = model.GetQueryResultFromCache(projectID, requestPayload, &resultContainer)
			if errCode == http.StatusAccepted {
				continue
			} else if errCode == http.StatusFound {
				return getQueryCacheResponse(c, cacheResult, forDashboard, skipContextVerfication)
			} else {
				// If not in Accepted state, return with error.
				return true, http.StatusInternalServerError, errors.New("Query Cache: Failed to fetch from cache")
			}
		}
	}
	return false, errCode, errors.New("Query Cache: Failed to fetch from cache")
}

func UseUserFunnelV2(c *gin.Context) bool {
	if c.Request.Header.Get(model.QueryFunnelV2) == "true" {
		return true
	}

	return false
}

// InValidateSavedQueryCache Common function to invalidate cache if present.
func InValidateSavedQueryCache(query *model.Queries) int {
	failedKeys := make([]string, 0)
	units := store.GetStore().GetDashboardUnitForQueryID(query.ProjectID, query.ID)
	if units == nil {
		return http.StatusOK
	}
	for _, unit := range units {
		failedKeys = append(failedKeys, InValidateDashboardQueryCache(unit.ProjectID, unit.DashboardId, unit.ID)...)
	}

	log.WithField("key", units).Info("Query cache key")

	if len(failedKeys) != 0 {
		return http.StatusInternalServerError
	}
	return http.StatusOK
}

// This can give both pattern or keys which failed.
func InValidateDashboardQueryCache(projectID, dashboardID, unitID int64) []string {

	failedKeys := make([]string, 0)
	var cacheKeys []*cacheRedis.Key
	var err error

	pattern := fmt.Sprintf("dashboard:*:pid:%d:did:%d:duid:%d:*", projectID, dashboardID, unitID)
	cacheKey, err := cacheRedis.ScanPersistent(pattern, 50000000, 50000000)
	cacheKeys = append(cacheKeys, cacheKey...)
	if C.GetAttributionDebug() == 1 {
		log.WithFields(log.Fields{
			"projectID":   projectID,
			"dashboardID": dashboardID,
			"unitID":      unitID,
			"pattern":     pattern,
			"cacheKeys":   cacheKeys,
		}).Info("InValidateDashboardQueryCache")
	}
	if err != nil {
		log.WithError(err).Error("Failed to get cache key")
		failedKeys = append(failedKeys, pattern)
	}

	for _, cacheKey := range cacheKeys {
		err := cacheRedis.DelPersistent(cacheKey)
		key, _ := cacheKey.Key()
		if err != nil {
			failedKeys = append(failedKeys, key)
		}
	}
	return failedKeys
}

func GetResponseFromDBCaching(reqId string, projectID int64, dashboardID, unitID int64, from, to int64, timezoneString U.TimeZoneString) (bool, int, interface{}) {

	errCode, cacheResult := store.GetStore().FetchCachedResultFromDataBase(reqId, projectID, dashboardID, unitID, from, to)
	// since cacheResult.Result is []byte, need to marshall it
	resultJson := &postgres.Jsonb{RawMessage: json.RawMessage(cacheResult.Result)}
	if errCode == http.StatusFound && cacheResult.Result != nil {
		return true, http.StatusOK, DashboardQueryResponsePayload{Result: resultJson, Cache: true, RefreshedAt: cacheResult.ComputedAt, TimeZone: string(timezoneString), CacheMeta: nil}
	}
	return false, errCode, nil
}

// GetResponseIfCachedDashboardQuery Common function to fetch result from cache if present for dashboard query.
func GetResponseIfCachedDashboardQuery(reqId string, projectID int64, dashboardID, unitID int64, from, to int64,
	timezoneString U.TimeZoneString) (bool, int, interface{}) {
	cacheResult, errCode, err := model.GetCacheResultByDashboardIdAndUnitId(reqId, projectID, dashboardID, unitID, from, to, timezoneString)
	if errCode == http.StatusFound && cacheResult != nil {
		return true, http.StatusOK, DashboardQueryResponsePayload{Result: cacheResult.Result, Cache: true,
			RefreshedAt: cacheResult.RefreshedAt, TimeZone: string(timezoneString), CacheMeta: cacheResult.CacheMeta}
	}
	return false, errCode, err
}

func GetResponseIfCachedDashboardQueryWithPreset(reqId string, projectID int64, dashboardID, unitID int64, preset string,
	from, to int64, timezoneString U.TimeZoneString) (bool, int, interface{}) {
	cacheResult, errCode, err := model.GetCacheResultByDashboardIdAndUnitIdWithPreset(reqId, projectID, dashboardID,
		unitID, preset, from, to, timezoneString)
	if errCode == http.StatusFound && cacheResult != nil {
		return true, http.StatusOK, DashboardQueryResponsePayload{Result: cacheResult.Result, Cache: true,
			RefreshedAt: cacheResult.RefreshedAt, TimeZone: string(timezoneString), CacheMeta: cacheResult.CacheMeta}
	}
	return false, errCode, err
}

func GetResponseFromDBCachingObject(reqId string, projectID int64, dashboardID, unitID int64, from, to int64,
	timezoneString U.TimeZoneString) (bool, int, DashboardQueryResponsePayload) {

	errCode, cacheResult := store.GetStore().FetchCachedResultFromDataBase(reqId, projectID, dashboardID, unitID, from, to)

	// since cacheResult.Result is []byte, need to marshall it
	resultJson := &postgres.Jsonb{RawMessage: json.RawMessage(cacheResult.Result)}
	if errCode == http.StatusFound && cacheResult.Result != nil {
		return true, http.StatusOK, DashboardQueryResponsePayload{Result: resultJson, Cache: true,
			RefreshedAt: cacheResult.ComputedAt, TimeZone: string(timezoneString), CacheMeta: nil}
	}
	return false, errCode, DashboardQueryResponsePayload{}
}
