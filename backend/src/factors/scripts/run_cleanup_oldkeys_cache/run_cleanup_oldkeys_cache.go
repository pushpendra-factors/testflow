package main

import (
	C "factors/config"
	"flag"

	"factors/model/model"

	"github.com/gomodule/redigo/redis"
	log "github.com/sirupsen/logrus"
)

func main() {
	// Get persistent connection
	// Scan set by set (EN:pid:*) (EN:PC:pid:*) (EN:PV:pid:*) (US:PC:pid:*) (US:PV:pid:*)
	// del set by set

	env := flag.String("env", "development", "")
	redisHostPersistent := flag.String("redis_host_ps", "localhost", "")
	RedisPortPersistent := flag.Int("redis_port_ps", 6379, "")

	flag.Parse()
	taskID := "cleanup_offline"
	config := &C.Configuration{
		AppName:             taskID,
		Env:                 *env,
		RedisHostPersistent: *redisHostPersistent,
		RedisPortPersistent: *RedisPortPersistent,
	}
	C.InitConf(config)

	// Cache dependency for requests not using queue.
	C.InitRedisPersistent(config.RedisHostPersistent, config.RedisPortPersistent)

	redisConn := C.GetCacheRedisPersistentConnection()

	defer redisConn.Close()
	cursor := 0
	pattern := model.QueryCacheRedisKeyPrefix
	for {
		res, err := redis.Values(redisConn.Do("SCAN", cursor, "MATCH", pattern, "COUNT", 1000))
		if err != nil {
			log.WithError(err).Error("scan failed")
		}
		cacheKeyStrings := make([]interface{}, 0)
		redis.Scan(res, &cursor, &cacheKeyStrings)
		if len(cacheKeyStrings) > 0 {
			log.WithField("del", len(cacheKeyStrings)).Info("Del EN:")
			_, err = redisConn.Do("DEL", cacheKeyStrings...)
			if err != nil {
				log.WithError(err).Error("del failed")
			}
		}
		if cursor == 0 {
			break
		}
	}
	log.Info("Done")
}
