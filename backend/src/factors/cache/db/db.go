package db

import (
	"errors"
	"factors/cache"
	"factors/config"
	"factors/util"
	"time"

	"github.com/jinzhu/gorm"
)

type CacheDBRecord struct {
	// Using k and v as column name to avoid using keywords.
	Key          string    `gorm:"k" json:"key"`
	Value        string    `gorm:"v" json:"value"`
	ProjectID    int64     `json:"project_id"`
	ExpiryInSecs float64   `json:"expiry_in_secs"`
	ExpiresAt    int64     `json:"expires_at"` // unix_timestamp for allowing sorting.
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

var tableNameCacheTable = "cache_db"
var ErrKeyNotExists = errors.New("key not exist")

func Get(key *cache.Key) (string, error) {
	if key == nil {
		return "", errors.New("invalid key")
	}

	k, err := key.Key()
	if err != nil {
		return "", errors.New("invalid key")
	}

	if key.ProjectID == 0 {
		return "", errors.New("invalid project_id")
	}

	if !config.IsCacheDBReadEnabled(key.ProjectID) {
		return "", errors.New("cache db not enabled")
	}

	var value string
	db := config.GetServices().Db
	err = db.Table(tableNameCacheTable).Select("v").Limit(1).Where("project_id = ? AND k = ?", key.ProjectID, k).Find(&value).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return "", ErrKeyNotExists
		}
		return "", err
	}

	return value, nil
}

func GetIfExists(key *cache.Key) (string, bool, error) {
	if key == nil {
		return "", false, errors.New("invalid key")
	}

	v, err := Get(key)
	if err != nil {
		if err == ErrKeyNotExists {
			return "", false, nil
		}

		// Consider key exists but error to
		// match redis implementation.
		return "", true, err
	}

	return v, true, nil
}

func getCacheRecord(key *cache.Key, value string, expiryInSecs float64) (*CacheDBRecord, error) {
	if key == nil {
		return nil, errors.New("invalid key")
	}

	if key.ProjectID == 0 {
		return nil, errors.New("invalid project_id")
	}

	if !config.IsCacheDBWriteEnabled(key.ProjectID) {
		return nil, errors.New("cache db not enabled")
	}

	if value == "" {
		return nil, errors.New("invalid value")
	}

	if expiryInSecs <= 0 {
		return nil, errors.New("invalid expiry")
	}

	k, err := key.Key()
	if err != nil {
		return nil, err
	}

	cacheRecord := &CacheDBRecord{
		Key:          k,
		Value:        value,
		ExpiryInSecs: expiryInSecs,
	}

	if expiryInSecs > 0 {
		cacheRecord.ExpiresAt = util.TimeNowUnix() + int64(expiryInSecs)
	}

	return cacheRecord, nil
}

func Set(key *cache.Key, value string, expiryInSecs float64) error {
	cacheRecord, err := getCacheRecord(key, value, expiryInSecs)
	if err != nil {
		return err
	}

	if cacheRecord == nil {
		return errors.New("invalid cache payload")
	}

	db := config.GetServices().Db
	err = db.Table(tableNameCacheTable).Create(cacheRecord).Error
	if err != nil {
		return err
	}

	return nil
}

func SetBatch(keyValue map[*cache.Key]string, expiryInSecs float64) error {
	cacheRecords := make([]*CacheDBRecord, 0)
	for k, v := range keyValue {
		cacheRecord, err := getCacheRecord(k, v, expiryInSecs)
		// Fail if one failure. Parity with redis implementation.
		if err != nil {
			return err
		}

		cacheRecords = append(cacheRecords, cacheRecord)
	}

	db := config.GetServices().Db
	err := db.Table(tableNameCacheTable).Create(cacheRecords).Error
	if err != nil {
		return err
	}

	return nil
}
