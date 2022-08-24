package helpers

import (
	"errors"
	"factors/model/model"
	"fmt"
	"net/http"
	"strconv"
	"time"

	C "factors/config"
	U "factors/util"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

const HeaderUserFilterOptForProfiles string = "Use-Filter-Opt-Profiles"
const HeaderUserFilterOptForEventsAndUsers string = "Use-Filter-Opt-Events-Users"

// DashboardQueryResponsePayload Query query response with cache and refreshed_at.
type DashboardQueryResponsePayload struct {
	Result      interface{} `json:"result"`
	Cache       bool        `json:"cache"`
	RefreshedAt int64       `json:"refreshed_at"`
	TimeZone    string      `json:"timezone"`
	CacheMeta   interface{} `json:"cache_meta"`
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

// GetResponseIfCachedDashboardQuery Common function to fetch result from cache if present for dashboard query.
func GetResponseIfCachedDashboardQuery(reqId string, projectID int64, dashboardID, unitID int64, from, to int64, timezoneString U.TimeZoneString) (bool, int, interface{}) {
	cacheResult, errCode, err := model.GetCacheResultByDashboardIdAndUnitId(reqId, projectID, dashboardID, unitID, from, to, timezoneString)
	if errCode == http.StatusFound && cacheResult != nil {
		return true, http.StatusOK, DashboardQueryResponsePayload{Result: cacheResult.Result, Cache: true, RefreshedAt: cacheResult.RefreshedAt, TimeZone: string(timezoneString), CacheMeta: cacheResult.CacheMeta}
	}
	return false, errCode, err
}

func GetResponseIfCachedDashboardQueryWithPreset(reqId string, projectID int64, dashboardID, unitID int64, preset string, from, to int64, timezoneString U.TimeZoneString) (bool, int, interface{}) {
	cacheResult, errCode, err := model.GetCacheResultByDashboardIdAndUnitIdWithPreset(reqId, projectID, dashboardID, unitID, preset, from, to, timezoneString)
	if errCode == http.StatusFound && cacheResult != nil {
		return true, http.StatusOK, DashboardQueryResponsePayload{Result: cacheResult.Result, Cache: true, RefreshedAt: cacheResult.RefreshedAt, TimeZone: string(timezoneString), CacheMeta: cacheResult.CacheMeta}
	}
	return false, errCode, err
}
