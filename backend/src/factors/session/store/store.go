package store

import (
	"factors/session"
	"factors/config"
	cookieStore "factors/session/store/cookie"
	redisStore "factors/session/store/redis"
)

func GetSessionStore() session.Session {
	var store session.Session
	if st := config.GetSessionStore(); st == "cookie" {
		store = &cookieStore.Cookie{}
	} else if st == "redis" {
		store = &redisStore.Redis{}
	}
	return store
}
