package redis

import (
	"errors"
	"fmt"

	"github.com/gomodule/redigo/redis"

	C "factors/config"
)

type Key struct {
	ProjectID uint64
	Prefix    string
	Suffix    string
}

var (
	ErrorInvalidProject = errors.New("invalid key project_id")
	ErrorInvalidPrefix  = errors.New("invalid key prefix")
	ErrorInvalidSuffix  = errors.New("invalid key suffix")
	ErrorInvalidKey     = errors.New("invalid redis cache key")
)

func NewKey(projectId uint64, prefix string, suffix string) (*Key, error) {
	if projectId == 0 {
		return nil, ErrorInvalidProject
	}

	if prefix == "" {
		return nil, ErrorInvalidPrefix
	}

	if suffix == "" {
		return nil, ErrorInvalidSuffix
	}

	return &Key{ProjectID: projectId, Prefix: prefix, Suffix: suffix}, nil
}

func (key *Key) Key() (string, error) {
	if key.ProjectID == 0 {
		return "", ErrorInvalidProject
	}

	if key.Prefix == "" {
		return "", ErrorInvalidPrefix
	}

	if key.Suffix == "" {
		return "", ErrorInvalidSuffix
	}

	// key: i.e, pid:1:uid:1:user_last_event
	return fmt.Sprintf("%s:pid:%d:%s", key.Prefix, key.ProjectID, key.Suffix), nil
}

func Set(key *Key, value string, expiryInSecs float64) error {
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

	redisConn := C.GetRedisConn()
	defer redisConn.Close()

	if expiryInSecs == 0 {
		_, err = redisConn.Do("SET", cKey, value)
	} else {
		_, err = redisConn.Do("SET", cKey, value, "EX", expiryInSecs)
	}

	return err
}

func Get(key *Key) (string, error) {
	if key == nil {
		return "", ErrorInvalidKey
	}

	cKey, err := key.Key()
	if err != nil {
		return "", err
	}

	redisConn := C.GetRedisConn()
	defer redisConn.Close()

	return redis.String(redisConn.Do("GET", cKey))
}
