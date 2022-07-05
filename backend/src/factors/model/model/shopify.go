package model

import (
	"fmt"
	"net/http"

	cacheRedis "factors/cache/redis"

	log "github.com/sirupsen/logrus"
)

func getShopifyCartTokenCacheKey(projectId int64, cartToken string) (*cacheRedis.Key, error) {
	prefix := "shp:ctkn"
	suffix := cartToken
	return cacheRedis.NewKey(projectId, prefix, suffix)
}

func SetCacheShopifyCartTokenToUserId(projectId int64, cartToken string, userId string) int {
	logCtx := log.WithField("project_id", projectId)
	if projectId == 0 || cartToken == "" || userId == "" {
		logCtx.Error(fmt.Sprintf(
			"Invalid shopify cart token input. projectId: %d, cartToken:%s, userId:%s",
			projectId, cartToken, userId))
		return http.StatusBadRequest
	}

	cartTokenKey, err := getShopifyCartTokenCacheKey(projectId, cartToken)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get ShopifyCartTokenCacheKey.")
		return http.StatusInternalServerError
	}

	err = cacheRedis.Set(cartTokenKey, userId, 7*24*60*60)
	if err != nil {
		logCtx.WithError(err).Error("Failed to setCacheShopifyCartTokenToUserId.")
		return http.StatusInternalServerError
	}
	return http.StatusOK
}

func GetCacheUserIdForShopifyCartToken(projectId int64, cartToken string) (string, int) {
	userId := ""
	logCtx := log.WithField("project_id", projectId)
	if projectId == 0 {
		logCtx.Error("GetCacheUserIdForShopifyCartToken: Invalid project id")
		return userId, http.StatusBadRequest
	}

	cartTokenKey, err := getShopifyCartTokenCacheKey(projectId, cartToken)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get ShopifyCartTokenCacheKey.")
		return userId, http.StatusInternalServerError
	}

	userId, err = cacheRedis.Get(cartTokenKey)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get value for ShopifyCartTokenCacheKey.")
		return userId, http.StatusInternalServerError
	}

	return userId, http.StatusOK
}
