package redis

import (
	"errors"
	"fmt"

	"github.com/gomodule/redigo/redis"

	C "factors/config"
)

type Key struct {
	// any one must be set.
	ProjectID  uint64
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

func NewKey(projectId uint64, prefix string, suffix string) (*Key, error) {
	if projectId == 0 {
		return nil, ErrorInvalidProject
	}

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

func DelPersistent(key *Key) error {
	return del(key, true)
}

func Del(key *Key) error {
	return del(key, false)
}

func del(key *Key, persistent bool) error {
	if key == nil {
		return ErrorInvalidKey
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

	_, err = redisConn.Do("DEL", cKey)
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
