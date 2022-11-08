package clear_bit

import (
	"encoding/json"
	cacheRedis "factors/cache/redis"
	"fmt"
	"net/http"

	"github.com/gomodule/redigo/redis"
	log "github.com/sirupsen/logrus"
)

//SetClearBitCacheResult Sets the cache result key in redis.
func SetClearBitCacheResult(projectID int64, userId string, userIP string) {
	logCtx := log.WithFields(log.Fields{
		"project_id": projectID,
		"user_id":    userId,
		"user_ip":    userIP,
	})

	cacheKey, err := GetClearbitCacheRedisKey(projectID, userId, userIP)
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
func GetClearbitCacheRedisKey(projectID int64, userId string, userIP string) (*cacheRedis.Key, error) {
	prefix := "ip:enrichment:clearbit"
	suffix := fmt.Sprintf("userId:%d:userIP:%v", userId, userIP)
	return cacheRedis.NewKey(projectID, prefix, suffix)
}

func GetClearbitCacheResult(projectID int64, userId string, userIP string) (bool, int) {
	var cacheResult bool
	logCtx := log.WithFields(log.Fields{
		"project_id": projectID,
		"user_id":    userId,
		"user_ip":    userIP,
	})
	cacheKey, err := GetClearbitCacheRedisKey(projectID, userId, userIP)
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
