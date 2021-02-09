package helpers

import (
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

// DashboardQueryResponsePayload Query query response with cache and refreshed_at.
type DashboardQueryResponsePayload struct {
	Result      interface{} `json:"result"`
	Cache       bool        `json:"cache"`
	RefreshedAt int64       `json:"refreshed_at"`
}

func getQueryCacheResponse(c *gin.Context, cacheResult model.QueryCacheResult, forDashboard bool) (bool, int, interface{}) {
	if forDashboard {
		return true, http.StatusOK, DashboardQueryResponsePayload{Result: cacheResult.Result, Cache: true, RefreshedAt: cacheResult.RefreshedAt}
	}
	// To Indicate if the result is served from cache without changing the response format.
	c.Header(model.QueryCacheResponseFromCacheHeader, "true")
	c.Header(model.QueryCacheResponseCacheRefreshedAt, fmt.Sprint(cacheResult.RefreshedAt))
	return true, http.StatusOK, cacheResult.Result
}

// IsHardRefreshForToday To check from query api if hard refresh should be applied or return from cache.
func IsHardRefreshForToday(from int64, hardRefresh bool) bool {
	return U.IsStartOfTodaysRange(from, U.TimeZoneStringIST) && hardRefresh
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
func GetResponseIfCachedQuery(c *gin.Context, projectID uint64, requestPayload model.BaseQuery,
	resultContainer interface{}, forDashboard bool) (bool, int, interface{}) {

	cacheResult, errCode := model.GetQueryResultFromCache(projectID, requestPayload, &resultContainer)
	if errCode == http.StatusFound {
		return getQueryCacheResponse(c, cacheResult, forDashboard)
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
				return getQueryCacheResponse(c, cacheResult, forDashboard)
			} else {
				// If not in Accepted state, return with error.
				return true, http.StatusInternalServerError, gin.H{"error": "Failed executing query"}
			}
		}
	}
	return false, http.StatusNotFound, nil
}

// GetResponseIfCachedDashboardQuery Common function to fetch result from cache if present for dashboard query.
func GetResponseIfCachedDashboardQuery(projectID, dashboardID, unitID uint64, from, to int64) (bool, int, interface{}) {
	cacheResult, errCode, errMsg := model.GetCacheResultByDashboardIdAndUnitId(projectID, dashboardID, unitID, from, to)
	if errCode == http.StatusFound && cacheResult != nil {
		return true, http.StatusOK, DashboardQueryResponsePayload{Result: cacheResult.Result, Cache: true, RefreshedAt: cacheResult.RefreshedAt}
	} else if errCode == http.StatusBadRequest {
		return true, errCode, gin.H{"error": errMsg}
	}

	if errCode != http.StatusNotFound {
		log.WithFields(log.Fields{"project_id": projectID,
			"dashboard_id": dashboardID, "dashboard_unit_id": unitID,
		}).WithError(errMsg).Error("Failed to get GetCacheChannelResultByDashboardIdAndUnitId from cache.")
	}
	return false, http.StatusNotFound, nil
}
