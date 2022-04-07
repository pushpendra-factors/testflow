package cookie

import (
	C "factors/config"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

func (c *Cookie) InitSessionStore(r *gin.Engine) error {
	store := cookie.NewStore([]byte(C.GetSessionStoreSecret()))
	// Set maxage to one minute
	store.Options(sessions.Options{
		MaxAge: 60,
	})
	r.Use(sessions.Sessions("session", store))
	return nil
}

func (cookie *Cookie) GetValueAsString(c *gin.Context, key string) string {
	session := sessions.Default(c)
	v := session.Get(key)
	if v == nil {
		return ""
	}
	return v.(string)
}

func (cookie *Cookie) SetValue(c *gin.Context, key string, value string) error {
	session := sessions.Default(c)
	session.Set(key, value)
	return session.Save()
}

func (cookie *Cookie) DeleteValue(c *gin.Context, key string) error {
	session := sessions.Default(c)
	session.Delete(key)
	return session.Save()
}
