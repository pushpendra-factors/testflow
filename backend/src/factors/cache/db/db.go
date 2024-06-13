package db

import (
	"errors"
	"factors/cache"
	"factors/config"
	"factors/util"
	"strings"
	"sync"
	"time"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
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

// Runs each batch of queries with same transaction and wait for all batches to complete.
// Routines are used for concurrency per batch.
// Lower the size higher the concurrency as no.of queries per transaction is lower.
const setBatchBatchSize = 5

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

	var record CacheDBRecord
	db := config.GetServices().Db
	err = db.Table(tableNameCacheTable).Select("v").Limit(1).Where("project_id = ? AND k = ?", key.ProjectID, k).Find(&record).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return "", ErrKeyNotExists
		}
		return "", err
	}

	return record.Value, nil
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

	k, err := key.Key()
	if err != nil {
		return nil, err
	}

	// Any expiry set less than equal to 0 is set to 1 year.
	if expiryInSecs <= 0 {
		expiryInSecs = 31560000
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

			// return nil after update on duplicate.
			return nil
		}
		return err
	}

	return nil
}

func setBatchWithBatchRoutines(keyValue map[*cache.Key]string, expiryInSecs float64) error {
	lenM := len(keyValue)

	if lenM <= setBatchBatchSize {
		return setBatch(keyValue, expiryInSecs)
	}

	// Create batches before routines for simplicity.
	i := 0
	totalI := 0
	kvMap := map[*cache.Key]string{}
	batchMap := make([]map[*cache.Key]string, 0)
	for k, v := range keyValue {
		kvMap[k] = v
		i++
		totalI++

		if i == setBatchBatchSize || totalI == lenM {
			batchMap = append(batchMap, kvMap)
			kvMap = map[*cache.Key]string{}
			i = 0
		}
	}

	var wg sync.WaitGroup
	for i := range batchMap {
		wg.Add(1)
		go setBatchWithWg(batchMap[i], expiryInSecs, &wg)
	}
	wg.Wait()

	return nil

}

func SetBatch(keyValue map[*cache.Key]string, expiryInSecs float64) error {
	var wg sync.WaitGroup

	// Using as many go routines as length of the map for simplicity.
	// Temporarily controlling db concurrency.
	for k, v := range keyValue {
		wg.Add(1)
		go setWithWg(k, v, expiryInSecs, &wg)
	}
	wg.Wait()

	return nil
}

func setBatchWithWg(keyValue map[*cache.Key]string, expiryInSecs float64, wg *sync.WaitGroup) error {
	defer wg.Done()
	return setBatch(keyValue, expiryInSecs)
}

func setWithWg(key *cache.Key, value string, expiryInSecs float64, wg *sync.WaitGroup) error {
	defer wg.Done()
	return Set(key, value, expiryInSecs)
}

func setBatch(keyValue map[*cache.Key]string, expiryInSecs float64) error {
	db := config.GetServices().Db
	tx, err := db.DB().Begin()
	defer util.CloseTx(tx)
	if err != nil {
		return err
	}

	for k, v := range keyValue {
		cacheRecord, err := getCacheRecord(k, v, expiryInSecs)
		// Fail if one failure. Parity with redis implementation.
		if err != nil {
			return err
		}

		stmnt := "INSERT INTO cache_db (`project_id`, `k`, `v`, `expiry_in_secs`, `expires_at`, `created_at`, `updated_at`) VALUES (?, ?, ?, ?, ?, ?, ?);"
		_, err = tx.Exec(stmnt, cacheRecord.ProjectID, cacheRecord.Key, cacheRecord.Value,
			cacheRecord.ExpiryInSecs, cacheRecord.ExpiresAt, util.TimeNowZ(), util.TimeNowZ())
		if err != nil {
			if IsDuplicateRecordError(err) {
				// Update the record incase of conflict.
				updateStmnt := "UPDATE cache_db SET v = ? WHERE project_id = ? AND k = ?;"
				_, err = tx.Exec(updateStmnt, cacheRecord.Value, cacheRecord.ProjectID, cacheRecord.Key)
				if err != nil {
					log.WithField("k", k).Info("Updated key on db_cache after conflict during insert.")
				}
				continue
			}

			return err
		}
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
