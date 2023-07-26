package model

import (
	"encoding/json"
	cacheRedis "factors/cache/redis"

	U "factors/util"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gomodule/redigo/redis"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

type SixSignalQueryGroup struct {
	Queries []SixSignalQuery `json:"six_signal_query_group"`
}

func (q *SixSignalQueryGroup) SetTimeZone(timezoneString U.TimeZoneString) {
	for i := 0; i < len(q.Queries); i++ {
		q.Queries[i].Timezone = timezoneString
	}
}

type SixSignalQuery struct {
	Timezone U.TimeZoneString `json:"tz"`
	From     int64            `json:"fr"`
	To       int64            `json:"to"`
	IsSaved  bool             `json:"isSaved"`
	PageView []string         `json:"pageView"`
}

type SixSignalResultGroup struct {
	Results   []SixSignalQueryResult `json:"result_group"`
	Query     interface{}            `json:"query"`
	CacheMeta interface{}            `json:"cache_meta"`
	//is_shareable reflects if the results are allowed to share through a public-url
	IsShareable bool `json:"is_shareable"`
}

type SixSignalQueryResult struct {
	Headers []string        `json:"headers"`
	Rows    [][]interface{} `json:"rows"`
	Query   interface{}     `json:"query"`
}

type SixSignalShareableURLParams struct {
	Query           *postgres.Jsonb `json:"six_signal_query"`
	EntityType      int             `json:"entity_type"`
	ShareType       int             `json:"share_type"`
	IsExpirationSet bool            `json:"is_expiration_set"`
	ExpirationTime  int64           `json:"expiration_time"`
}

type SixSignalPublicURLResponse struct {
	RouteVersion string `json:"route_version"`
	QueryID      string `json:"query_id"`
}

type SixSignalEmailAndMessage struct {
	EmailIDs []string         `json:"email_ids"`
	Url      string           `json:"url"`
	Domain   string           `json:"domain"`
	From     int64            `json:"fr"`
	To       int64            `json:"to"`
	Timezone U.TimeZoneString `json:"tz"`
}

// SetSixSignalCacheResult Sets the cache result key in redis.
func SetSixSignalCacheResult(projectID int64, userId string, userIP string) {
	logCtx := log.WithFields(log.Fields{
		"project_id": projectID,
		"user_id":    userId,
		"user_ip":    userIP,
	})

	cacheKey, err := GetSixSignalCacheRedisKey(projectID, userId, userIP)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get cache key")
		return
	}

	resultString, err := json.Marshal(true)
	if err != nil {
		return
	}

	var ipEnrichmentCacheInvalidationDuration float64 = 24 * 60 * 60 // 24 hrs
	err = cacheRedis.SetPersistent(cacheKey, string(resultString), ipEnrichmentCacheInvalidationDuration)
	if err != nil {
		logCtx.WithError(err).Error("Failed to set cache for ip enrichment")
		return
	}

}
func GetSixSignalCacheRedisKey(projectID int64, userId string, userIP string) (*cacheRedis.Key, error) {
	prefix := "ip:enrichment:sixsignal"
	suffix := fmt.Sprintf("userId:%s:userIP:%s", userId, userIP)
	return cacheRedis.NewKey(projectID, prefix, suffix)
}

func GetSixSignalCacheResult(projectID int64, userId string, userIP string) (bool, int) {
	var cacheResult bool
	logCtx := log.WithFields(log.Fields{
		"project_id": projectID,
		"user_id":    userId,
		"user_ip":    userIP,
	})
	cacheKey, err := GetSixSignalCacheRedisKey(projectID, userId, userIP)
	if err != nil {
		logCtx.WithError(err).Error("Error getting cache key")
		return cacheResult, http.StatusInternalServerError
	}

	result, err := cacheRedis.GetPersistent(cacheKey)
	if err == redis.ErrNil {
		return cacheResult, http.StatusNotFound
	} else if err != nil {
		logCtx.WithError(err).Error("Error getting key from redis")
		return cacheResult, http.StatusInternalServerError
	}
	err = json.Unmarshal([]byte(result), &cacheResult)
	if err != nil {
		logCtx.WithError(err).Errorf("Error decoding redis result %v", result)
		return cacheResult, http.StatusInternalServerError
	}
	return cacheResult, http.StatusFound
}

// GetSixSignalAPICountCacheRedisKey returns the redis key when given projectID and timeZone
func GetSixSignalAPICountCacheRedisKey(projectID int64, date uint64) (*cacheRedis.Key, error) {
	prefix := "ip:enrichment:sixsignal"
	suffix := fmt.Sprintf("%d", date)
	return cacheRedis.NewKey(projectID, prefix, suffix) //Sample Key: "ip:enrichment:sixsignal:pid:399:20221130"
}

// SetSixSignalAPICountCacheResult fetches the count of number of times API has been hit and increases it by 1.
func SetSixSignalAPICountCacheResult(projectID int64, timeZone U.TimeZoneString) {
	logCtx := log.WithFields(log.Fields{
		"project_id": projectID,
	})

	date := U.DateAsYYYYMMDDFormat(U.TimeNowIn(timeZone))

	cacheKey, err := GetSixSignalAPICountCacheRedisKey(projectID, date)
	if err != nil {
		logCtx.Warn("Failed to get cache key")
		return
	}
	count := GetSixSignalAPICountCacheResult(projectID, date)
	if count <= 0 {
		count = 0
	}
	count = count + 1
	var apiCountCacheInvalidationDuration int64 = 60 * U.SECONDS_IN_A_DAY
	err = cacheRedis.SetPersistent(cacheKey, strconv.Itoa(count), float64(apiCountCacheInvalidationDuration))
	if err != nil {
		logCtx.Warn("Failed to set cache for API Count")
		return
	}

}

// GetSixSignalAPITotalHitCountCacheRedisKey returns the redis key when given projectID and timeZone
func GetSixSignalAPITotalHitCountCacheRedisKey(projectID int64, date uint64) (*cacheRedis.Key, error) {
	prefix := "ip:enrichment:total:sixsignal"
	suffix := fmt.Sprintf("%d", date)
	return cacheRedis.NewKey(projectID, prefix, suffix) //Sample Key: "ip:enrichment:total:sixsignal:pid:399:20221130"
}

// GetSixSignalAPITotalHitCountCacheResult returns the total count of number of times 6Signal API has been called when projectID and timeZone is given
func GetSixSignalAPITotalHitCountCacheResult(projectID int64, date uint64) int {
	cacheResult := 0
	logCtx := log.WithFields(log.Fields{
		"project_id": projectID,
	})
	cacheKey, err := GetSixSignalAPITotalHitCountCacheRedisKey(projectID, date)
	if err != nil {
		logCtx.WithError(err).Error("Error getting cache key")
		return cacheResult
	}

	result, err := cacheRedis.GetPersistent(cacheKey)
	if err == redis.ErrNil {
		return cacheResult
	} else if err != nil {
		logCtx.WithError(err).Error("Error getting key from redis")
		return cacheResult
	}

	err = json.Unmarshal([]byte(result), &cacheResult)
	if err != nil {
		logCtx.Warn("Error decoding redis result %v", result)
		return cacheResult
	}
	return cacheResult

}

// SetSixSignalAPITotalHitCountCacheResult fetches the total count of number of times API has been hit and increases it by 1.
func SetSixSignalAPITotalHitCountCacheResult(projectID int64, timeZone U.TimeZoneString) {
	logCtx := log.WithFields(log.Fields{
		"project_id": projectID,
	})

	date := U.DateAsYYYYMMDDFormat(U.TimeNowIn(timeZone))

	cacheKey, err := GetSixSignalAPITotalHitCountCacheRedisKey(projectID, date)
	if err != nil {
		logCtx.Warn("Failed to get cache key total api hit count")
		return
	}
	count := GetSixSignalAPITotalHitCountCacheResult(projectID, date)
	if count <= 0 {
		count = 0
	}
	count = count + 1
	var apiCountCacheInvalidationDuration int64 = 60 * U.SECONDS_IN_A_DAY
	err = cacheRedis.SetPersistent(cacheKey, strconv.Itoa(count), float64(apiCountCacheInvalidationDuration))
	if err != nil {
		logCtx.Warn("Failed to set cache for API Total Count")
		return
	}
}

// GetSixSignalAPICountCacheResult returns the count of number of times 6Signal API has been called when projectID and timeZone is given
func GetSixSignalAPICountCacheResult(projectID int64, date uint64) int {
	cacheResult := 0
	logCtx := log.WithFields(log.Fields{
		"project_id": projectID,
	})
	cacheKey, err := GetSixSignalAPICountCacheRedisKey(projectID, date)
	if err != nil {
		logCtx.WithError(err).Error("Error getting cache key")
		return cacheResult
	}

	result, err := cacheRedis.GetPersistent(cacheKey)
	if err == redis.ErrNil {
		return cacheResult
	} else if err != nil {
		logCtx.WithError(err).Error("Error getting key from redis")
		return cacheResult
	}

	err = json.Unmarshal([]byte(result), &cacheResult)
	if err != nil {
		logCtx.Warn("Error decoding redis result %v", result)
		return cacheResult
	}
	return cacheResult

}

func GetSixSignalMonthlyUniqueEnrichmentKey(projectId int64, monthYear string) (*cacheRedis.Key, error) {
	prefix := "unique:enrichment:monthly:sixsignal"
	suffix := monthYear
	return cacheRedis.NewKey(projectId, prefix, suffix) //Sample Key: "unique:enrichment:monthly:sixsignal:pid:399:May2023"
}

func GetSixSignalMonthlyUniqueEnrichmentCount(projectId int64, monthYear string) (int64, error) {
	key, err := GetSixSignalMonthlyUniqueEnrichmentKey(projectId, monthYear)
	if err != nil {
		return -1, err
	}

	count, err := cacheRedis.PFCountPersistent(key)
	if err != nil {
		return -1, err
	}

	intCount := convertInterfaceByType(count)

	//Decreasing 1% of the count to handle the 0.81% error rate of redis hyperloop PFCOUNT.
	finalCount := int64(0.99 * float64(intCount))

	return finalCount, nil
}

func SetSixSignalMonthlyUniqueEnrichmentCount(projectId int64, value string, timeZone U.TimeZoneString) error {

	monthYear := U.GetCurrentMonthYear(timeZone)
	key, err := GetSixSignalMonthlyUniqueEnrichmentKey(projectId, monthYear)
	if err != nil {
		return err
	}

	_, err = cacheRedis.PFAddPersistent(key, value, 0)
	return err
}

func convertInterfaceByType(obj interface{}) int64 {

	var intCount int64
	switch intCount := obj.(type) {

	case int64:
		intCount = obj.(int64)
		return intCount
	case int:
		intCount = obj.(int)
		return int64(intCount)
	}

	return intCount
}
