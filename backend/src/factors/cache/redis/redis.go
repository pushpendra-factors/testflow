package redis

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/gomodule/redigo/redis"

	C "factors/config"
)

type Key struct {
	// any one must be set.
	ProjectID  int64
	ProjectUID string
	// Prefix - Helps better grouping and searching
	// i.e table_name + index_name
	Prefix string
	// Suffix - optional
	Suffix string
}

var (
	ErrorInvalidProject  = errors.New("invalid key project")
	ErrorInvalidPrefix   = errors.New("invalid key prefix")
	ErrorInvalidKey      = errors.New("invalid redis cache key")
	ErrorInvalidValues   = errors.New("invalid values to set")
	ErrorPartialFailures = errors.New("Partial failures in Set")
)

func NewKeyWithOnlyPrefix(prefix string) (*Key, error) {

	if prefix == "" {
		return nil, ErrorInvalidPrefix
	}

	return &Key{Prefix: prefix}, nil
}

func NewKey(projectId int64, prefix string, suffix string) (*Key, error) {
	if projectId == 0 {
		return nil, ErrorInvalidProject
	}

	if prefix == "" {
		return nil, ErrorInvalidPrefix
	}

	return &Key{ProjectID: projectId, Prefix: prefix, Suffix: suffix}, nil
}

func NewKeyWithAllProjectsSupport(projectId int64, prefix string, suffix string) (*Key, error) {
	if prefix == "" {
		return nil, ErrorInvalidPrefix
	}
	return &Key{ProjectID: projectId, Prefix: prefix, Suffix: suffix}, nil
}

// NewKeyWithProjectUID - Uses projectUID as project scope on the key.
func NewKeyWithProjectUID(projectUID, prefix, suffix string) (*Key, error) {
	if projectUID == "" {
		return nil, ErrorInvalidProject
	}

	if prefix == "" {
		return nil, ErrorInvalidPrefix
	}

	return &Key{ProjectUID: projectUID, Prefix: prefix, Suffix: suffix}, nil
}

func (key *Key) Key() (string, error) {
	if key.ProjectID == 0 && key.ProjectUID == "" {
		return "", ErrorInvalidProject
	}

	if key.Prefix == "" {
		return "", ErrorInvalidPrefix
	}

	var projectScope string
	if key.ProjectID != 0 {
		projectScope = fmt.Sprintf("pid:%d", key.ProjectID)
	} else {
		projectScope = fmt.Sprintf("puid:%s", key.ProjectUID)
	}

	// key: i.e, event_names:user_last_event:pid:1:uid:1
	return fmt.Sprintf("%s:%s:%s", key.Prefix, projectScope, key.Suffix), nil
}

func (key *Key) KeyWithAllProjectsSupport() (string, error) {
	if key.Prefix == "" {
		return "", ErrorInvalidPrefix
	}

	var projectScope string
	if key.ProjectID == 0 {
		projectScope = "pid:*"
	} else {
		projectScope = fmt.Sprintf("pid:%d", key.ProjectID)
	}
	// key: i.e, event_names:user_last_event:pid:*:suffix
	return fmt.Sprintf("%s:%s:%s", key.Prefix, projectScope, key.Suffix), nil
}

func (key *Key) KeyWithOnlyPrefix() (string, error) {
	if key.Prefix == "" {
		return "", ErrorInvalidPrefix
	}

	// key: i.e, event_names:user_last_event:pid:*:suffix
	return fmt.Sprintf("%s", key.Prefix), nil
}

// KeyFromStringWithPid - Splits the cache key into prefix/suffix/projectid format
// Only for pid based cache
func KeyFromStringWithPid(key string) (*Key, error) {
	if key == "" {
		return nil, ErrorInvalidValues
	}
	cacheKey := Key{}
	keyPidSplit := strings.Split(key, ":pid:")
	if len(keyPidSplit) == 2 {
		projectIDSuffix := strings.SplitN(keyPidSplit[1], ":", 2)
		if len(projectIDSuffix) == 2 {
			cacheKey.Suffix = projectIDSuffix[1]
		}
		projectId, _ := strconv.Atoi(projectIDSuffix[0])
		cacheKey.ProjectID = int64(projectId)
		cacheKey.Prefix = keyPidSplit[0]
	}
	return &cacheKey, nil
}

func SetPersistent(key *Key, value string, expiryInSecs float64) error {
	return set(key, value, expiryInSecs, true)
}

func Set(key *Key, value string, expiryInSecs float64) error {
	return set(key, value, expiryInSecs, false)
}

func set(key *Key, value string, expiryInSecs float64, persistent bool) error {
	if key == nil {
		return ErrorInvalidKey
	}

	if value == "" {
		return errors.New("empty cache key value")
	}

	cKey, err := key.Key()
	if err != nil {
		return err
	}

	var redisConn redis.Conn
	if persistent {
		redisConn = C.GetCacheRedisPersistentConnection()
	} else {
		redisConn = C.GetCacheRedisConnection()
	}
	defer redisConn.Close()

	if expiryInSecs == 0 {
		_, err = redisConn.Do("SET", cKey, value)
	} else {
		_, err = redisConn.Do("SET", cKey, value, "EX", int64(expiryInSecs))
	}

	return err
}

// GetIfExistsPersistent - Check if the cache key exists, if not return null
// Get value if the cache key exists
func GetIfExistsPersistent(key *Key) (string, bool, error) {
	ifExists, err := ExistsPersistent(key)
	if err != nil {
		return "", false, err
	}
	if !ifExists {
		return "", false, err
	}
	value, err := GetPersistent(key)
	if err != nil {
		return "", true, err
	}
	return value, true, nil
}

func GetPersistent(key *Key) (string, error) {
	return get(key, true)
}

func Get(key *Key) (string, error) {
	return get(key, false)
}

func get(key *Key, persistent bool) (string, error) {
	if key == nil {
		return "", ErrorInvalidKey
	}

	cKey, err := key.Key()
	if err != nil {
		return "", err
	}

	var redisConn redis.Conn
	if persistent {
		redisConn = C.GetCacheRedisPersistentConnection()
	} else {
		redisConn = C.GetCacheRedisConnection()
	}
	defer redisConn.Close()

	return redis.String(redisConn.Do("GET", cKey))
}

func MGetPersistent(keys ...*Key) ([]string, error) {
	return mGet(true, keys...)
}

func MGet(keys ...*Key) ([]string, error) {
	return mGet(false, keys...)
}

// MGet Function to get multiple keys from redis. Returns slice of result strings.
func mGet(persistent bool, keys ...*Key) ([]string, error) {

	var cKeys []interface{}
	var cValues []string
	for _, key := range keys {
		if key == nil {
			return cValues, ErrorInvalidKey
		}
		cKey, err := key.Key()
		if err != nil {
			return cValues, err
		}
		cKeys = append(cKeys, cKey)
	}

	var redisConn redis.Conn
	if persistent {
		redisConn = C.GetCacheRedisPersistentConnection()
	} else {
		redisConn = C.GetCacheRedisConnection()
	}
	defer redisConn.Close()

	values, err := redis.Values(redisConn.Do("MGET", cKeys...))
	if err != nil {
		return cValues, err
	}

	if err := redis.ScanSlice(values, &cValues); err != nil {
		return cValues, err
	}
	return cValues, nil
}

func DelPersistent(keys ...*Key) error {
	return del(true, keys...)
}

func Del(keys ...*Key) error {
	return del(false, keys...)
}

func del(persistent bool, keys ...*Key) error {
	var cKeys []interface{}

	for _, key := range keys {
		if key == nil {
			return ErrorInvalidKey
		}
		cKey, err := key.Key()
		if err != nil {
			return err
		}
		cKeys = append(cKeys, cKey)
	}

	var redisConn redis.Conn
	if persistent {
		redisConn = C.GetCacheRedisPersistentConnection()
	} else {
		redisConn = C.GetCacheRedisConnection()
	}
	defer redisConn.Close()

	_, err := redisConn.Do("DEL", cKeys...)
	if err != nil {
		return err
	}
	return err
}

func ExistsPersistent(key *Key) (bool, error) {
	return exists(key, true)
}

func Exists(key *Key) (bool, error) {
	return exists(key, false)
}

// Exists Checks if a key exists in Redis.
func exists(key *Key, persistent bool) (bool, error) {
	if key == nil {
		return false, ErrorInvalidKey
	}

	cKey, err := key.Key()
	if err != nil {
		return false, err
	}

	var redisConn redis.Conn
	if persistent {
		redisConn = C.GetCacheRedisPersistentConnection()
	} else {
		redisConn = C.GetCacheRedisConnection()
	}
	defer redisConn.Close()

	count, err := redisConn.Do("EXISTS", cKey)
	if err != nil {
		return false, err
	}
	return count.(int64) == 1, nil
}

func IncrBatch(keys ...*Key) ([]int64, error) {
	return incrBatch(false, keys)
}
func IncrPersistentBatch(keys ...*Key) ([]int64, error) {
	return incrBatch(true, keys)
}
func incrBatch(persistent bool, keys []*Key) ([]int64, error) {
	if len(keys) == 0 {
		return nil, ErrorInvalidValues
	}
	var redisConn redis.Conn
	if persistent {
		redisConn = C.GetCacheRedisPersistentConnection()
	} else {
		redisConn = C.GetCacheRedisConnection()
	}
	defer redisConn.Close()

	err := redisConn.Send("MULTI")
	if err != nil {
		return nil, err

	}
	for _, key := range keys {
		cKey, err := key.Key()
		if err != nil {
			return nil, err
		}
		err = redisConn.Send("INCR", cKey)
		if err != nil {
			return nil, err
		}
	}
	res, err := redis.Values(redisConn.Do("EXEC"))
	if err != nil {
		return nil, err
	}
	counts := make([]int64, 0)
	if err := redis.ScanSlice(res, &counts); err != nil {
		return nil, err
	}

	return counts, nil
}

type SortedSetKeyValueTuple struct {
	Key   *Key
	Value string
}

func ZincrBatch(OnlyPrefixKey bool, keys ...SortedSetKeyValueTuple) ([]int64, error) {
	return zincrBatch(false, keys, OnlyPrefixKey)
}
func ZincrPersistentBatch(OnlyPrefixKey bool, keys ...SortedSetKeyValueTuple) ([]int64, error) {
	return zincrBatch(true, keys, OnlyPrefixKey)
}

func zincrBatch(persistent bool, keys []SortedSetKeyValueTuple, OnlyPrefixKey bool) ([]int64, error) {
	if len(keys) == 0 {
		return nil, ErrorInvalidValues
	}
	var redisConn redis.Conn
	if persistent {
		redisConn = C.GetCacheRedisPersistentConnection()
	} else {
		redisConn = C.GetCacheRedisConnection()
	}
	defer redisConn.Close()

	err := redisConn.Send("MULTI")
	if err != nil {
		return nil, err

	}
	for _, key := range keys {
		var err error
		var cKey string
		if OnlyPrefixKey == true {
			cKey, err = key.Key.KeyWithOnlyPrefix()
		} else {
			cKey, err = key.Key.Key()
		}
		if err != nil {
			return nil, err
		}
		err = redisConn.Send("ZINCRBY", cKey, 1, key.Value)
		if err != nil {
			return nil, err
		}
	}
	res, err := redis.Values(redisConn.Do("EXEC"))
	if err != nil {
		return nil, err
	}
	counts := make([]int64, 0)
	if err := redis.ScanSlice(res, &counts); err != nil {
		return nil, err
	}

	return counts, nil
}

func SetBatch(values map[*Key]string, expiryInSecs float64) error {
	return setBatch(values, expiryInSecs, false)
}

func SetPersistentBatch(values map[*Key]string, expiryInSecs float64) error {
	return setBatch(values, expiryInSecs, true)
}

func setBatch(values map[*Key]string, expiryInSecs float64, persistent bool) error {
	if len(values) == 0 {
		return ErrorInvalidValues
	}
	var redisConn redis.Conn
	if persistent {
		redisConn = C.GetCacheRedisPersistentConnection()
	} else {
		redisConn = C.GetCacheRedisConnection()
	}
	defer redisConn.Close()

	err := redisConn.Send("MULTI")
	if err != nil {
		return err
	}
	for key, value := range values {
		cKey, err := key.Key()
		if err != nil {
			return err
		}
		if expiryInSecs == 0 {
			err = redisConn.Send("SET", cKey, value)
		} else {
			err = redisConn.Send("SET", cKey, value, "EX", int64(expiryInSecs))
		}
		if err != nil {
			return err
		}
	}
	res, err := redis.Values(redisConn.Do("EXEC"))
	if err != nil {
		return err
	}
	if len(res) != len(values) {
		return ErrorPartialFailures
	}
	return nil
}

func GetKeysPersistent(pattern string) ([]*Key, error) {
	return getKeys(pattern, true)
}

func GetKeys(pattern string) ([]*Key, error) {
	return getKeys(pattern, false)
}

func getKeys(pattern string, persistent bool) ([]*Key, error) {
	if pattern == "" {
		return nil, ErrorInvalidKey
	}

	var redisConn redis.Conn
	if persistent {
		redisConn = C.GetCacheRedisPersistentConnection()
	} else {
		redisConn = C.GetCacheRedisConnection()
	}
	defer redisConn.Close()

	keys, err := redis.Values(redisConn.Do("KEYS", pattern))
	if err != nil {
		return nil, err
	}
	cacheKeyStrings := make([]string, 0)
	_ = redis.ScanSlice(keys, &cacheKeyStrings)

	cacheKeys := make([]*Key, 0)
	for _, key := range cacheKeyStrings {
		cacheKey, _ := KeyFromStringWithPid(key)
		cacheKeys = append(cacheKeys, cacheKey)
	}
	return cacheKeys, nil
}

func PFAddPersistent(cacheKey *Key, value string, expiryInSeconds float64) (bool, error) {
	return pfAdd(cacheKey, value, expiryInSeconds, true)
}

func PFAdd(cacheKey *Key, value string, expiryInSeconds float64) (bool, error) {
	return pfAdd(cacheKey, value, expiryInSeconds, false)
}

func pfAdd(cacheKey *Key, value string, expiryInSeconds float64, persistent bool) (bool, error) {
	if cacheKey == nil {
		return false, ErrorInvalidKey
	}
	cKey, err := cacheKey.Key()
	if err != nil {
		return false, err
	}
	var redisConn redis.Conn
	if persistent {
		redisConn = C.GetCacheRedisPersistentConnection()
	} else {
		redisConn = C.GetCacheRedisConnection()
	}
	defer redisConn.Close()

	res, err := redisConn.Do("PFADD", cKey, value)
	if err != nil {
		return false, err
	}
	if expiryInSeconds != 0 {
		_, err := redisConn.Do("EXPIRE", cKey, int64(expiryInSeconds))
		if err != nil {
			return false, err
		}
	}
	if res.(int64) == 1 {
		return true, nil
	}
	return false, nil
}

func Scan(pattern string, perScanCount int64, limit int64) ([]*Key, error) {
	return scan(pattern, perScanCount, limit, false)
}

func ScanPersistent(pattern string, perScanCount int64, limit int64) ([]*Key, error) {
	return scan(pattern, perScanCount, limit, true)
}

func scan(pattern string, perScanCount int64, limit int64, persistent bool) ([]*Key, error) {
	var redisConn redis.Conn
	if persistent {
		redisConn = C.GetCacheRedisPersistentConnection()
	} else {
		redisConn = C.GetCacheRedisConnection()
	}
	defer redisConn.Close()

	cacheKeys := make([]*Key, 0)
	cursor := 0
	for {
		res, err := redis.Values(redisConn.Do("SCAN", cursor, "MATCH", pattern, "COUNT", perScanCount))
		if err != nil {
			return nil, err
		}
		cacheKeyStrings := make([]string, 0)
		redis.Scan(res, &cursor, &cacheKeyStrings)
		if err != nil {
			return nil, err
		}
		for _, key := range cacheKeyStrings {
			cacheKey, _ := KeyFromStringWithPid(key)
			cacheKeys = append(cacheKeys, cacheKey)
		}
		if cursor == 0 || (limit != -1 && int64(len(cacheKeys)) >= limit) {
			break
		}
	}
	return cacheKeys, nil
}

type KeyCountTuple struct {
	Key   *Key
	Count int64
}

func IncrByBatch(keys []KeyCountTuple) ([]int64, error) {
	return incrByBatch(keys, false)
}

func IncrByBatchPersistent(keys []KeyCountTuple) ([]int64, error) {
	return incrByBatch(keys, true)
}

func incrByBatch(keys []KeyCountTuple, persistent bool) ([]int64, error) {
	if len(keys) == 0 {
		return nil, ErrorInvalidValues
	}
	var redisConn redis.Conn
	if persistent {
		redisConn = C.GetCacheRedisPersistentConnection()
	} else {
		redisConn = C.GetCacheRedisConnection()
	}
	defer redisConn.Close()

	err := redisConn.Send("MULTI")
	if err != nil {
		return nil, err
	}
	for _, tuple := range keys {
		cKey, err := tuple.Key.Key()
		if err != nil {
			return nil, err
		}
		err = redisConn.Send("INCRBY", cKey, tuple.Count)
		if err != nil {
			return nil, err
		}
	}
	res, err := redis.Values(redisConn.Do("EXEC"))
	if err != nil {
		return nil, err
	}
	counts := make([]int64, 0)
	if err := redis.ScanSlice(res, &counts); err != nil {
		return nil, err
	}
	// TODO: Check for partial failures
	return counts, nil
}

func DecrByBatch(keys map[*Key]int64) error {
	return decrByBatch(keys, false)
}

func DecrByBatchPersistent(keys map[*Key]int64) error {
	return decrByBatch(keys, true)
}

func decrByBatch(keys map[*Key]int64, persistent bool) error {
	if len(keys) == 0 {
		return ErrorInvalidValues
	}
	var redisConn redis.Conn
	if persistent {
		redisConn = C.GetCacheRedisPersistentConnection()
	} else {
		redisConn = C.GetCacheRedisConnection()
	}
	defer redisConn.Close()

	err := redisConn.Send("MULTI")
	if err != nil {
		return err
	}
	for key, value := range keys {
		cKey, err := key.Key()
		if err != nil {
			return err
		}
		err = redisConn.Send("DECRBY", cKey, value)
		if err != nil {
			return err
		}
	}
	_, err = redis.Values(redisConn.Do("EXEC"))
	if err != nil {
		return err
	}
	// TODO: Check for partial failures
	return nil
}

func Zcard(key *Key) (int64, error) {
	return zcard(key, false)
}
func ZcardPersistent(key *Key) (int64, error) {
	return zcard(key, true)
}
func zcard(key *Key, persistent bool) (int64, error) {
	if key == nil {
		return 0, ErrorInvalidKey
	}

	cKey, err := key.Key()
	if err != nil {
		return 0, err
	}

	var redisConn redis.Conn
	if persistent {
		redisConn = C.GetCacheRedisPersistentConnection()
	} else {
		redisConn = C.GetCacheRedisConnection()
	}
	defer redisConn.Close()

	return redis.Int64(redisConn.Do("ZCARD", cKey))
}

func ZrangeWithScores(OnlyPrefixKey bool, key *Key) (map[string]string, error) {
	return zrangeWithScores(OnlyPrefixKey, key, false)
}
func ZrangeWithScoresPersistent(OnlyPrefixKey bool, key *Key) (map[string]string, error) {
	return zrangeWithScores(OnlyPrefixKey, key, true)
}
func zrangeWithScores(OnlyPrefixKey bool, key *Key, persistent bool) (map[string]string, error) {
	if key == nil {
		return nil, ErrorInvalidKey
	}
	var err error
	var cKey string
	if OnlyPrefixKey == true {
		cKey, err = key.KeyWithOnlyPrefix()
	} else {
		cKey, err = key.Key()
	}
	if err != nil {
		return nil, err
	}

	var redisConn redis.Conn
	if persistent {
		redisConn = C.GetCacheRedisPersistentConnection()
	} else {
		redisConn = C.GetCacheRedisConnection()
	}
	defer redisConn.Close()

	values, err := redis.Values(redisConn.Do("ZRANGE", cKey, 0, -1, "WITHSCORES"))
	if err != nil {
		return nil, err
	}
	sortedSetValues := make([]string, 0)
	_ = redis.ScanSlice(values, &sortedSetValues)

	sortedSetMap := make(map[string]string)
	for i := 0; i < len(sortedSetValues); {
		sortedSetMap[sortedSetValues[i]] = sortedSetValues[i+1]
		i = i + 2
	}
	return sortedSetMap, nil
}

func ZRemRange(key *Key, startIndex int, endIndex int) (int64, error) {
	return zRemRange(key, startIndex, endIndex, false)
}
func ZRemRangePersistent(key *Key, startIndex int, endIndex int) (int64, error) {
	return zRemRange(key, startIndex, endIndex, true)
}
func zRemRange(key *Key, startIndex int, endIndex int, persistent bool) (int64, error) {
	if key == nil {
		return 0, ErrorInvalidKey
	}

	cKey, err := key.Key()
	if err != nil {
		return 0, err
	}

	var redisConn redis.Conn
	if persistent {
		redisConn = C.GetCacheRedisPersistentConnection()
	} else {
		redisConn = C.GetCacheRedisConnection()
	}
	defer redisConn.Close()

	return redis.Int64(redisConn.Do("ZREMRANGEBYRANK", cKey, startIndex, endIndex))
}
