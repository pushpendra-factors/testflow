package cache

import (
	"errors"
	"factors/cache"
	"factors/cache/db"
	"factors/cache/redis"
	"factors/config"

	log "github.com/sirupsen/logrus"
)

func Set(key *cache.Key, value string, expiryInSecs float64, useDB bool) error {
	logCtx := log.WithField("tag", "cache_db").WithField("key", key).
		WithField("expirty_in_secs", expiryInSecs).
		WithField("use_db", useDB)

	useCacheDB := config.IsCacheDBWriteEnabled(key.ProjectID) && useDB
	if !useCacheDB {
		return redis.SetPersistent(key, value, expiryInSecs)
	}

	// TODO: Writing to both. Remove redis once migration is completed.
	redis.SetPersistent(key, value, expiryInSecs)

	logCtx.Info("Writing to cache db.")
	err := db.Set(key, value, expiryInSecs)
	if err != nil {
		logCtx.WithError(err).Warn("Failed to write to db cache.")
	}
	return err
}

func SetBatch(keyValue map[*cache.Key]string, expiryInSecs float64, useDB bool) error {
	logCtx := log.WithField("tag", "cache_db").WithField("keys", len(keyValue)).
		WithField("expirty_in_secs", expiryInSecs).
		WithField("use_db", useDB)

	if len(keyValue) == 0 {
		return errors.New("invalid key values")
	}

	var projectID int64
	for k := range keyValue {
		projectID = k.ProjectID
	}

	useCacheDB := config.IsCacheDBWriteEnabled(projectID) && useDB
	if !useCacheDB {
		return redis.SetPersistentBatch(keyValue, expiryInSecs)
	}

	// TODO: Writing to both. Remove redis once migration is completed.
	redis.SetPersistentBatch(keyValue, expiryInSecs)

	logCtx.Info("Writing to cache db.")
	err := db.SetBatch(keyValue, expiryInSecs)
	if err != nil {
		logCtx.WithError(err).Warn("Failed to write to db cache.")
	}
	return err
}

func Get(key *cache.Key, useDB bool) (string, error) {
	logCtx := log.WithField("tag", "cache_db").WithField("key", key).WithField("use_db", useDB)

	useCacheDB := config.IsCacheDBReadEnabled(key.ProjectID) && useDB
	if !useCacheDB {
		return redis.GetPersistent(key)
	}

	logCtx.Info("Getting from cache db.")
	v, err := db.Get(key)
	if err != nil {
		logCtx.WithError(err).Warn("Failed to get from cache.")
	}
	return v, err
}

func GetIfExists(key *cache.Key, useDB bool) (string, bool, error) {
	logCtx := log.WithField("tag", "cache_db").WithField("key", key).WithField("use_db", useDB)

	useCacheDB := config.IsCacheDBReadEnabled(key.ProjectID) && useDB
	if !useCacheDB {
		return redis.GetIfExistsPersistent(key)
	}

	logCtx.Info("Getting from cache db.")
	v, b, err := db.GetIfExists(key)
	if err != nil {
		logCtx.WithError(err).Warn("Failed to get from cache.")
	}
	return v, b, err
}
