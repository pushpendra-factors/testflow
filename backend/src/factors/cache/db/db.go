package db

import (
	"errors"
	"factors/cache"
	"factors/config"
	"factors/util"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
)

type CacheDBRecord struct {
	// Using k and v as column name to avoid using keywords.
	Key          string    `gorm:"column:k" json:"key"`
	Value        string    `gorm:"column:v" json:"value"`
	ProjectID    int64     `json:"project_id"`
	ExpiryInSecs float64   `json:"expiry_in_secs"`
	ExpiresAt    int64     `json:"expires_at"` // unix_timestamp for allowing sorting.
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (CacheDBRecord) TableName() string {
	return "cache_db"
}

const (
	// Error 1062: Leaf Error (127.0.0.1:3307): Duplicate entry '5000247-FactorsGoal1' for key 'unique_project_id_name_idx'
	// Error 1062: Leaf Error (127.0.0.1:3307): Duplicate entry '6000762-f6e8c235-7aa0-42fe-b987-137866bcdd8f' for key 'PRIMARY'
	MEMSQL_ERROR_CODE_DUPLICATE_ENTRY = "Error 1062"
)

func IsDuplicateRecordError(err error) bool {
	return strings.HasPrefix(err.Error(), MEMSQL_ERROR_CODE_DUPLICATE_ENTRY)
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
		ProjectID:    key.ProjectID,
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

	cacheRecord.ProjectID = key.ProjectID

	db := config.GetServices().Db
	err = db.Table(tableNameCacheTable).Create(cacheRecord).Error
	if err != nil {
		if IsDuplicateRecordError(err) {
			k, _ := key.Key()
			err := db.Table(tableNameCacheTable).
				Where("project_id = ? AND k = ?", key.ProjectID, k).
				Updates(map[string]interface{}{"v": value}).Error
			if err != nil {
				return err
			}
		}
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

func Del(keys ...*cache.Key) error {
	var projectID int64

	keyNames := make([]string, 0)
	for i := range keys {
		k, err := keys[i].Key()
		if err != nil {
			continue
		}

		if i == 0 {
			projectID = keys[i].ProjectID
		}
		if projectID != keys[i].ProjectID {
			return errors.New("different projects not allowed")
		}

		keyNames = append(keyNames, k)
	}

	if projectID == 0 {
		return errors.New("invalid project_id")
	}

	if !config.IsCacheDBWriteEnabled(projectID) {
		return errors.New("cache db not enabled")
	}

	db := config.GetServices().Db
	err := db.Table(tableNameCacheTable).
		Where("project_id = ? AND k IN (?)", projectID, keyNames).
		Delete(&CacheDBRecord{}).Error
	if err != nil {
		return err
	}

	return nil
}
