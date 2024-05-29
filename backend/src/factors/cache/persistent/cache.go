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
	redisErr := redis.SetPersistentBatch(keyValue, expiryInSecs)
	logCtx.Info("Writing to cache db.")

	dbErr := db.SetBatch(keyValue, expiryInSecs)
	if dbErr != nil {
		logCtx.WithError(dbErr).Warn("Failed to write to db cache.")
	}

	// Using redis error till we fix Long Data and other issues related to batch writing.
	return redisErr
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

func Del(keys []*cache.Key, useDB bool) error {
	logCtx := log.WithField("tag", "cache_db").WithField("keys", len(keys)).WithField("use_db", useDB)

	if len(keys) == 0 {
		return nil
	}

	useCacheDB := config.IsCacheDBWriteEnabled(keys[0].ProjectID) && useDB
	if !useCacheDB {
		return redis.Del(keys...)
	}

	// TODO: Writing to both. Remove redis once migration is completed.
	redis.DelPersistent(keys...)

	logCtx.Info("Writing to cache db.")
	err := db.Del(keys...)
	if err != nil {
		logCtx.WithError(err).Warn("Failed to delete from db cache.")
	}
	return err
}
