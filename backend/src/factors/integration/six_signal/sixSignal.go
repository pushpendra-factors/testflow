package six_signal

import (
	"encoding/json"
	cacheRedis "factors/cache/redis"
	U "factors/util"
	"fmt"
	"github.com/gomodule/redigo/redis"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strconv"
)

//SetSixSignalCacheResult Sets the cache result key in redis.
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

	var ipEnrichmentCacheInvalidationDuration float64 = 30 * 60 // 30 minutes.
	err = cacheRedis.SetPersistent(cacheKey, string(resultString), ipEnrichmentCacheInvalidationDuration)
	if err != nil {
		logCtx.WithError(err).Error("Failed to set cache for ip enrichment")
		return
	}

}
func GetSixSignalCacheRedisKey(projectID int64, userId string, userIP string) (*cacheRedis.Key, error) {
	prefix := "ip:enrichment:sixsignal"
	suffix := fmt.Sprintf("userId:%d:userIP:%s", userId, userIP)
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

//GetSixSignalAPICountCacheRedisKey returns the redis key when given projectID and timeZone
func GetSixSignalAPICountCacheRedisKey(projectID int64, date uint64) (*cacheRedis.Key, error) {
	prefix := "ip:enrichment:sixsignal"
	suffix := fmt.Sprintf("%d", date)
	return cacheRedis.NewKey(projectID, prefix, suffix) //Sample Key: "ip:enrichment:sixsignal:pid:399:20221130"
}

//GetSixSignalAPICountCacheResult returns the count of number of times 6Signal API has been called when projectID and timeZone is given
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

//SetSixSignalAPICountCacheResult fetches the count of number of times API has been hit and increases it by 1.
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
