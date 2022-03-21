package session

import "github.com/gin-gonic/gin"

// "factors/model/model"

type Session interface {
	InitSessionStore(r *gin.Engine) error
	GetValueAsString(c *gin.Context, key string) string
	SetValue(c *gin.Context, key string, value string) error
	DeleteValue(c *gin.Context, key string) error
}
