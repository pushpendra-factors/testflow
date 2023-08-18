package model

import (
	"encoding/base64"
	"encoding/json"
	cacheRedis "factors/cache/redis"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

type OsInfo struct {
	Name      string `json:"name"`
	ShortName string `json:"short_name"`
	Version   string `json:"version"`
	Platform  string `json:"platform"`
	Family    string `json:"family"`
}
type ClientInfo struct {
	Type          string `json:"type"`
	Name          string `json:"name"`
	ShortName     string `json:"short_name"`
	Version       string `json:"version"`
	Engine        string `json:"engine"`
	EngineVersion string `json:"engine_version"`
	Family        string `json:"family"`
}

type DeviceInfo struct {
	IsBot       bool       `json:"is_bot"`
	ClientInfo  ClientInfo `json:"client_info"`
	OsInfo      OsInfo     `json:"os_info"`
	DeviceType  string     `json:"device_type"`
	DeviceBrand string     `json:"device_brand"`
	DeviceModel string     `json:"device_model"`
}

func GetDeviceInfoCacheRedisKey(userAgent string) (*cacheRedis.Key, error) {
	projectUid := base64.StdEncoding.EncodeToString([]byte(userAgent))
	prefix := "dd"
	return cacheRedis.NewKeyWithProjectUID(projectUid, prefix, "")
}

func GetCacheResultByUserAgent(userAgent string) (*DeviceInfo, int, error) {
	start := time.Now()
	var cacheResult *DeviceInfo
	logCtx := log.WithFields(log.Fields{
		"userAgent": userAgent,
		"Method":    "GetCacheResultByUserAgent",
	})

	if userAgent == "" {
		logCtx.Error("invalid user-agent")
		return cacheResult, http.StatusBadRequest, nil

	}

	cacheKey, err := GetDeviceInfoCacheRedisKey(userAgent)
	if err != nil {
		logCtx.WithError(err).Error("Failed to fetch cache key")
		return cacheResult, http.StatusInternalServerError, err
	}
	result, status, err := cacheRedis.GetIfExistsPersistent(cacheKey)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get data from cache")
		return cacheResult, http.StatusInternalServerError, err
	}
	if !status {
		logCtx.WithField("status", status).Warning("Not found in cache.")
		return cacheResult, http.StatusNotFound, err
	}

	err = json.Unmarshal([]byte(result), &cacheResult)
	if err != nil {
		logCtx.WithError(err).Error("Failed to unmarshal cache response")
		return cacheResult, http.StatusInternalServerError, err
	}

	// check for time elapsed for getting cache from redis
	elapsed := time.Since(start).Milliseconds()
	if elapsed > 100 {
		log.WithFields(log.Fields{"total_time": elapsed, "user_agent": userAgent}).Warning("GetCacheResultByUserAgent took more than 100 ms")
	}

	return cacheResult, http.StatusFound, nil

}

func SetCacheResultByUserAgent(userAgent string, result *DeviceInfo) {
	start := time.Now()

	logCtx := log.WithFields(log.Fields{
		"userAgent": userAgent,
		"Method":    "GetCacheResultByUserAgent",
	})

	if userAgent == "" {
		logCtx.Error("empty user agent")
		return
	}

	cacheKey, err := GetDeviceInfoCacheRedisKey(userAgent)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get cache key")
		return
	}
	deviceInfoCacheResult := result

	deviceInfoCacheResultJSON, err := json.Marshal(&deviceInfoCacheResult)
	if err != nil {
		logCtx.WithError(err).Error("Failed to encode deviceInfoCacheResult.")
		return
	}

	err = cacheRedis.SetPersistent(cacheKey, string(deviceInfoCacheResultJSON), 0)
	if err != nil {
		logCtx.WithError(err).Error("Failed to set cache for Device Service")
		return
	}

	// check for time elapsed for setting cache in redis
	elapsed := time.Since(start).Milliseconds()
	if elapsed > 100 {
		log.WithFields(log.Fields{"total_time": elapsed, "user_agent": userAgent}).Warning("SetCacheResultByUserAgent took more than 100 ms")
	}

}
