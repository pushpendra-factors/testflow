package cookie

type Cookie struct{}

func GetSessionStore() *Cookie {
	return &Cookie{}
}
