package cache

import (
	"errors"
	"factors/cache"
	"factors/cache/db"
	"factors/cache/redis"
	"factors/config"
)

func Set(key *cache.Key, value string, expiryInSecs float64, useDB bool) error {
	useCacheDB := config.IsCacheDBWriteEnabled(key.ProjectID) && useDB
	if !useCacheDB {
		return redis.SetPersistent(key, value, expiryInSecs)
	}

	// TODO: Writing to both. Remove redis once migration is completed.
	redis.SetPersistent(key, value, expiryInSecs)
	return db.Set(key, value, expiryInSecs)
}

func SetBatch(keyValue map[*cache.Key]string, expiryInSecs float64, useDB bool) error {
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
	return db.SetBatch(keyValue, expiryInSecs)
}

func Get(key *cache.Key, useDB bool) (string, error) {
	useCacheDB := config.IsCacheDBReadEnabled(key.ProjectID) && useDB
	if !useCacheDB {
		return redis.GetPersistent(key)
	}

	return db.Get(key)
}

func GetIfExists(key *cache.Key, useDB bool) (string, bool, error) {
	useCacheDB := config.IsCacheDBReadEnabled(key.ProjectID) && useDB
	if !useCacheDB {
		return redis.GetIfExistsPersistent(key)
	}

	return db.GetIfExists(key)
}
