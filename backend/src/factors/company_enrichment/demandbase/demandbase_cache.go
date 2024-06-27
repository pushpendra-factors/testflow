package demandbase

import (
	"encoding/json"
	"factors/cache"
	cacheRedis "factors/cache/redis"
	"fmt"
	"net/http"

	"github.com/gomodule/redigo/redis"
	log "github.com/sirupsen/logrus"
)

// SetDemandbaseCacheResult sets the value of redis key as true with 24 hours expiry.
func SetDemandbaseCacheResult(projectID int64, userId string, userIP string) {
	logCtx := log.WithFields(log.Fields{
		"project_id": projectID,
		"user_id":    userId,
		"user_ip":    userIP,
	})

	cacheKey, err := GetDemandbaseRedisKey(projectID, userId, userIP)
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

// GetDemandbaseRedisKey returns the redis key which is a combination of userIP, userID and projectID.
func GetDemandbaseRedisKey(projectID int64, userId string, userIP string) (*cache.Key, error) {
	prefix := "ip:enrichment:demandbase"
	suffix := fmt.Sprintf("userId:%v:userIP:%v", userId, userIP)
	return cache.NewKey(projectID, prefix, suffix)
}

// GetDemandbaseRedisCacheResult fetches the value of demandbase redis key from redis.
func GetDemandbaseRedisCacheResult(projectID int64, userId string, userIP string) (bool, int) {
	var cacheResult bool
	logCtx := log.WithFields(log.Fields{
		"project_id": projectID,
		"user_id":    userId,
		"user_ip":    userIP,
	})
	cacheKey, err := GetDemandbaseRedisKey(projectID, userId, userIP)
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
