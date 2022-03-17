package redis

type Redis struct{}

func GetSessionStore() *Redis {
	return &Redis{}
}
