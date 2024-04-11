package redis

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/gomodule/redigo/redis"
	log "github.com/sirupsen/logrus"

	C "factors/config"
)

type Key struct {
	// any one must be set.
	ProjectID  int64
	ProjectUID string
	AgentUUID  string
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

func NewKeyWithAgentUID(agentUUID, prefix, suffix string) (*Key, error) {
	if agentUUID == "" {
		return nil, ErrorInvalidProject
	}

	if prefix == "" {
		return nil, ErrorInvalidPrefix
	}

	return &Key{AgentUUID: agentUUID, Prefix: prefix, Suffix: suffix}, nil
}

func (key *Key) Key() (string, error) {
	if key.ProjectID == 0 && key.ProjectUID == "" && key.AgentUUID == "" {
		return "", ErrorInvalidProject
	}

	if key.Prefix == "" {
		return "", ErrorInvalidPrefix
	}

	var projectScope string
	if key.ProjectID != 0 {
		projectScope = fmt.Sprintf("pid:%d", key.ProjectID)
	} else if key.ProjectUID != "" {
		projectScope = fmt.Sprintf("puid:%s", key.ProjectUID)
	} else {
		projectScope = fmt.Sprintf("auuid:%s", key.AgentUUID)
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
	err := set(key, value, expiryInSecs, true, false)

	// Log for measuring dashboard and query cache usage.
	dashboardCache := strings.HasPrefix(key.Prefix, "dashboard:")
	queryCache := strings.HasPrefix(key.Prefix, "query:")
	if dashboardCache || queryCache {
		log.WithField("key", key).
			WithField("expiry_in_secs", expiryInSecs).
			WithField("is_dashboard_cache", dashboardCache).
			WithField("is_query_cache", queryCache).
			WithField("error", err).
			Info("Write dashboard/query cache.")
	}

	return err
}

func Set(key *Key, value string, expiryInSecs float64) error {
	return set(key, value, expiryInSecs, false, false)
}

func SetQueueRedis(key *Key, value string, expiryInSecs float64) error {
	return set(key, value, expiryInSecs, false, true)
}

func set(key *Key, value string, expiryInSecs float64, persistent bool, queue bool) error {
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
	if queue {
		redisConn = C.GetCacheQueueRedisConnection()
	} else if persistent {
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

func ZAddPersistent(key *Key, value string, expiryInSecs float64) error {
	return zadd(key, value, expiryInSecs, true, false)
}

func ZAdd(key *Key, value string, expiryInSecs float64) error {
	return zadd(key, value, expiryInSecs, false, false)
}

func ZAddQueueRedis(key *Key, value string, expiryInSecs float64) error {
	return zadd(key, value, expiryInSecs, false, true)
}

func zadd(key *Key, value string, expiryInSecs float64, persistent bool, queue bool) error {
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
	if queue {
		redisConn = C.GetCacheQueueRedisConnection()
	} else if persistent {
		redisConn = C.GetCacheRedisPersistentConnection()
	} else {
		redisConn = C.GetCacheRedisConnection()
	}
	defer redisConn.Close()

	_, err = redisConn.Do("ZADD", cKey, 1, value)
	if expiryInSecs != 0 {
		redisConn.Do("EXPIRE", cKey, int64(expiryInSecs))
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
	response, err := get(key, true, false)

	// Log for measuring dashboard and query cache usage.
	dashboardCache := strings.HasPrefix(key.Prefix, "dashboard:")
	queryCache := strings.HasPrefix(key.Prefix, "query:")
	if dashboardCache || queryCache {
		log.WithField("key", key).
			WithField("is_dashboard_cache", dashboardCache).
			WithField("is_query_cache", queryCache).
			WithField("is_cache_hit", err == nil).
			WithField("error", err).
			Info("Read dashboard/query cache.")
	}

	return response, err
}

func Get(key *Key) (string, error) {
	return get(key, false, false)
}

func GetQueueRedis(key *Key) (string, error) {
	return get(key, false, true)
}

func get(key *Key, persistent bool, queue bool) (string, error) {
	if key == nil {
		return "", ErrorInvalidKey
	}

	cKey, err := key.Key()
	if err != nil {
		return "", err
	}

	var redisConn redis.Conn
	if queue {
		redisConn = C.GetCacheQueueRedisConnection()
	} else if persistent {
		redisConn = C.GetCacheRedisPersistentConnection()
	} else {
		redisConn = C.GetCacheRedisConnection()
	}
	defer redisConn.Close()

	return redis.String(redisConn.Do("GET", cKey))
}

func MGetPersistent(keys ...*Key) ([]string, error) {
	return mGet(true, false, keys...)
}

func MGet(keys ...*Key) ([]string, error) {
	return mGet(false, false, keys...)
}

func MGetQueueRedis(keys ...*Key) ([]string, error) {
	return mGet(false, true, keys...)
}

// MGet Function to get multiple keys from redis. Returns slice of result strings.
func mGet(persistent bool, queue bool, keys ...*Key) ([]string, error) {

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
	if queue {
		redisConn = C.GetCacheQueueRedisConnection()
	} else if persistent {
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
	return del(true, false, keys...)
}

func Del(keys ...*Key) error {
	return del(false, false, keys...)
}

func DelQueueRedis(keys ...*Key) error {
	return del(false, true, keys...)
}

func del(persistent bool, queue bool, keys ...*Key) error {
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
	if queue {
		redisConn = C.GetCacheQueueRedisConnection()
	} else if persistent {
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
	return exists(key, true, false)
}

func Exists(key *Key) (bool, error) {
	return exists(key, false, false)
}

func ExistsQueueRedis(key *Key) (bool, error) {
	return exists(key, false, true)
}

// Exists Checks if a key exists in Redis.
func exists(key *Key, persistent bool, queue bool) (bool, error) {
	if key == nil {
		return false, ErrorInvalidKey
	}

	cKey, err := key.Key()
	if err != nil {
		return false, err
	}

	var redisConn redis.Conn
	if queue {
		redisConn = C.GetCacheQueueRedisConnection()
	} else if persistent {
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
	return incrBatch(false, false, keys)
}
func IncrPersistentBatch(keys ...*Key) ([]int64, error) {
	return incrBatch(true, false, keys)
}
func IncrQueueRedisBatch(keys ...*Key) ([]int64, error) {
	return incrBatch(false, true, keys)
}
func incrBatch(persistent bool, queue bool, keys []*Key) ([]int64, error) {
	if len(keys) == 0 {
		return nil, ErrorInvalidValues
	}
	var redisConn redis.Conn
	if queue {
		redisConn = C.GetCacheQueueRedisConnection()
	} else if persistent {
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
	return zincrBatch(false, false, keys, OnlyPrefixKey)
}
func ZincrPersistentBatch(OnlyPrefixKey bool, keys ...SortedSetKeyValueTuple) ([]int64, error) {
	return zincrBatch(true, false, keys, OnlyPrefixKey)
}
func ZincrQueueRedisBatch(OnlyPrefixKey bool, keys ...SortedSetKeyValueTuple) ([]int64, error) {
	return zincrBatch(false, true, keys, OnlyPrefixKey)
}

func zincrBatch(persistent bool, queue bool, keys []SortedSetKeyValueTuple, OnlyPrefixKey bool) ([]int64, error) {
	if len(keys) == 0 {
		return nil, ErrorInvalidValues
	}
	var redisConn redis.Conn
	if queue {
		redisConn = C.GetCacheQueueRedisConnection()
	} else if persistent {
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
	return setBatch(values, expiryInSecs, false, false)
}

func SetPersistentBatch(values map[*Key]string, expiryInSecs float64) error {
	return setBatch(values, expiryInSecs, true, false)
}

func SetBatchQueueRedis(values map[*Key]string, expiryInSecs float64) error {
	return setBatch(values, expiryInSecs, false, true)
}

func setBatch(values map[*Key]string, expiryInSecs float64, persistent bool, queue bool) error {
	if len(values) == 0 {
		return ErrorInvalidValues
	}
	var redisConn redis.Conn
	if queue {
		redisConn = C.GetCacheQueueRedisConnection()
	} else if persistent {
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
	return getKeys(pattern, true, false)
}

func GetKeys(pattern string) ([]*Key, error) {
	return getKeys(pattern, false, false)
}

func GetKeysQueueRedis(pattern string) ([]*Key, error) {
	return getKeys(pattern, false, true)
}

func getKeys(pattern string, persistent bool, queue bool) ([]*Key, error) {
	if pattern == "" {
		return nil, ErrorInvalidKey
	}

	var redisConn redis.Conn
	if queue {
		redisConn = C.GetCacheQueueRedisConnection()
	} else if persistent {
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
	return pfAdd(cacheKey, value, expiryInSeconds, true, false)
}

func PFAdd(cacheKey *Key, value string, expiryInSeconds float64) (bool, error) {
	return pfAdd(cacheKey, value, expiryInSeconds, false, false)
}

func PFAddQueueRedis(cacheKey *Key, value string, expiryInSeconds float64) (bool, error) {
	return pfAdd(cacheKey, value, expiryInSeconds, false, true)
}

func pfAdd(cacheKey *Key, value string, expiryInSeconds float64, persistent bool, queue bool) (bool, error) {
	if cacheKey == nil {
		return false, ErrorInvalidKey
	}
	cKey, err := cacheKey.Key()
	if err != nil {
		return false, err
	}
	var redisConn redis.Conn
	if queue {
		redisConn = C.GetCacheQueueRedisConnection()
	} else if persistent {
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

// PFCountPersistent - Returns the cardinalilty of the data present in HyperlogLog redis data structure
func PFCountPersistent(key *Key) (interface{}, error) {
	return pfCount(key, true, false)
}

func pfCount(key *Key, persistent bool, queue bool) (interface{}, error) {
	if key == nil {
		return -1, ErrorInvalidKey
	}

	cKey, err := key.Key()
	if err != nil {
		return -1, err
	}

	var redisConn redis.Conn
	if queue {
		redisConn = C.GetCacheQueueRedisConnection()
	} else if persistent {
		redisConn = C.GetCacheRedisPersistentConnection()
	} else {
		redisConn = C.GetCacheRedisConnection()
	}
	defer redisConn.Close()

	ifExists, err := redisConn.Do("EXISTS", cKey)
	if err != nil {
		return 0, nil
	}

	if ifExists.(int64) != 1 {
		return 0, nil
	}

	count, err := redisConn.Do("PFCOUNT", cKey)
	if err != nil {
		return -1, err
	}

	return count, nil
}

func Scan(pattern string, perScanCount int64, limit int64) ([]*Key, error) {
	return scan(pattern, perScanCount, limit, false, false)
}

func ScanPersistent(pattern string, perScanCount int64, limit int64) ([]*Key, error) {
	return scan(pattern, perScanCount, limit, true, false)
}

func ScanQueueRedis(pattern string, perScanCount int64, limit int64) ([]*Key, error) {
	return scan(pattern, perScanCount, limit, false, true)
}

func scan(pattern string, perScanCount int64, limit int64, persistent bool, queue bool) ([]*Key, error) {
	var redisConn redis.Conn
	if queue {
		redisConn = C.GetCacheQueueRedisConnection()
	} else if persistent {
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
	return incrByBatch(keys, false, false)
}

func IncrByBatchPersistent(keys []KeyCountTuple) ([]int64, error) {
	return incrByBatch(keys, true, false)
}

func IncrByBatchQueueRedis(keys []KeyCountTuple) ([]int64, error) {
	return incrByBatch(keys, false, true)
}

func incrByBatch(keys []KeyCountTuple, persistent bool, queue bool) ([]int64, error) {
	if len(keys) == 0 {
		return nil, ErrorInvalidValues
	}
	var redisConn redis.Conn
	if queue {
		redisConn = C.GetCacheQueueRedisConnection()
	} else if persistent {
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
	return decrByBatch(keys, false, false)
}

func DecrByBatchPersistent(keys map[*Key]int64) error {
	return decrByBatch(keys, true, false)
}

func DecrByBatchQueueRedis(keys map[*Key]int64) error {
	return decrByBatch(keys, false, true)
}

func decrByBatch(keys map[*Key]int64, persistent bool, queue bool) error {
	if len(keys) == 0 {
		return ErrorInvalidValues
	}
	var redisConn redis.Conn
	if queue {
		redisConn = C.GetCacheQueueRedisConnection()
	} else if persistent {
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
	return zcard(key, false, false)
}
func ZcardPersistent(key *Key) (int64, error) {
	return zcard(key, true, false)
}
func ZcardQueueRedis(key *Key) (int64, error) {
	return zcard(key, false, true)
}
func zcard(key *Key, persistent bool, queue bool) (int64, error) {
	if key == nil {
		return 0, ErrorInvalidKey
	}

	cKey, err := key.Key()
	if err != nil {
		return 0, err
	}

	var redisConn redis.Conn
	if queue {
		redisConn = C.GetCacheQueueRedisConnection()
	} else if persistent {
		redisConn = C.GetCacheRedisPersistentConnection()
	} else {
		redisConn = C.GetCacheRedisConnection()
	}
	defer redisConn.Close()

	return redis.Int64(redisConn.Do("ZCARD", cKey))
}

func ZScore(key *Key, member string) (int64, error) {
	return zscore(key, member, false, false)
}
func ZScorePersistent(key *Key, member string) (int64, error) {
	return zscore(key, member, true, false)
}
func ZScoreQueueRedis(key *Key, member string) (int64, error) {
	return zscore(key, member, false, true)
}
func zscore(key *Key, member string, persistent bool, queue bool) (int64, error) {
	if key == nil {
		return 0, ErrorInvalidKey
	}

	cKey, err := key.Key()
	if err != nil {
		return 0, err
	}

	var redisConn redis.Conn
	if queue {
		redisConn = C.GetCacheQueueRedisConnection()
	} else if persistent {
		redisConn = C.GetCacheRedisPersistentConnection()
	} else {
		redisConn = C.GetCacheRedisConnection()
	}
	defer redisConn.Close()

	result, err := redisConn.Do("ZSCORE", cKey, member)
	return redis.Int64(result, err)
}

func ZrangeWithScores(OnlyPrefixKey bool, key *Key) (map[string]string, error) {
	return zrangeWithScores(OnlyPrefixKey, key, false, false)
}
func ZrangeWithScoresPersistent(OnlyPrefixKey bool, key *Key) (map[string]string, error) {
	return zrangeWithScores(OnlyPrefixKey, key, true, false)
}
func ZrangeWithScoresQueueRedis(OnlyPrefixKey bool, key *Key) (map[string]string, error) {
	return zrangeWithScores(OnlyPrefixKey, key, false, true)
}
func zrangeWithScores(OnlyPrefixKey bool, key *Key, persistent bool, queue bool) (map[string]string, error) {
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
	if queue {
		redisConn = C.GetCacheQueueRedisConnection()
	} else if persistent {
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
	return zRemRange(key, startIndex, endIndex, false, false)
}
func ZRemRangePersistent(key *Key, startIndex int, endIndex int) (int64, error) {
	return zRemRange(key, startIndex, endIndex, true, false)
}
func ZRemRangeQueueRedis(key *Key, startIndex int, endIndex int) (int64, error) {
	return zRemRange(key, startIndex, endIndex, false, true)
}
func zRemRange(key *Key, startIndex int, endIndex int, persistent bool, queue bool) (int64, error) {
	if key == nil {
		return 0, ErrorInvalidKey
	}

	cKey, err := key.Key()
	if err != nil {
		return 0, err
	}

	var redisConn redis.Conn
	if queue {
		redisConn = C.GetCacheQueueRedisConnection()
	} else if persistent {
		redisConn = C.GetCacheRedisPersistentConnection()
	} else {
		redisConn = C.GetCacheRedisConnection()
	}
	defer redisConn.Close()

	return redis.Int64(redisConn.Do("ZREMRANGEBYRANK", cKey, startIndex, endIndex))
}

func ZRem(key *Key, OnlyPrefixKey bool, members ...string) (int64, error) {
	return zRem(key, OnlyPrefixKey, false, false, members...)
}
func ZRemPersistent(key *Key, OnlyPrefixKey bool, members ...string) (int64, error) {
	return zRem(key, OnlyPrefixKey, true, false, members...)
}
func ZRemQueueRedis(key *Key, OnlyPrefixKey bool, members ...string) (int64, error) {
	return zRem(key, OnlyPrefixKey, false, true, members...)
}
func zRem(key *Key, OnlyPrefixKey bool, persistent bool, queue bool, members ...string) (int64, error) {
	if key == nil {
		return 0, ErrorInvalidKey
	}

	var cKey string
	var err error
	if OnlyPrefixKey {
		cKey, err = key.KeyWithOnlyPrefix()
		if err != nil {
			return 0, err
		}
	} else {
		cKey, err = key.Key()
		if err != nil {
			return 0, err
		}
	}

	var redisConn redis.Conn
	if queue {
		redisConn = C.GetCacheQueueRedisConnection()
	} else if persistent {
		redisConn = C.GetCacheRedisPersistentConnection()
	} else {
		redisConn = C.GetCacheRedisConnection()
	}
	defer redisConn.Close()

	return redis.Int64(redisConn.Do("ZREM", redis.Args{}.Add(cKey).AddFlat(members)...))
}

func SetExpiry(key *Key, expiryInSeconds int) (int64, error) {
	return setExpiry(key, false, false, expiryInSeconds)
}
func SetExpiryPersistent(key *Key, expiryInSeconds int) (int64, error) {
	return setExpiry(key, true, false, expiryInSeconds)
}
func SetExpiryQueueRedis(key *Key, expiryInSeconds int) (int64, error) {
	return setExpiry(key, false, true, expiryInSeconds)
}
func setExpiry(key *Key, persistent bool, queue bool, expiryInSeconds int) (int64, error) {
	if key == nil {
		return 0, ErrorInvalidKey
	}

	cKey, err := key.Key()
	if err != nil {
		return 0, err
	}

	var redisConn redis.Conn
	if queue {
		redisConn = C.GetCacheQueueRedisConnection()
	} else if persistent {
		redisConn = C.GetCacheRedisPersistentConnection()
	} else {
		redisConn = C.GetCacheRedisConnection()
	}
	defer redisConn.Close()

	return redis.Int64(redisConn.Do("EXPIRE", cKey, expiryInSeconds))
}
