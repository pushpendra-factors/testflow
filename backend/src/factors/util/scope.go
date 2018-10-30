package util

import "github.com/gin-gonic/gin"

// SetScope sets scope to the context with a key/value.
func SetScope(c *gin.Context, key string, value interface{}) {
	scopeValue, exists := c.Get("scopes")
	if !exists {
		// Initializes scope with the key and value.
		c.Set("scopes", map[string]interface{}{key: value})
		return
	}

	scopeValue.(map[string]interface{})[key] = value
}

// GetScopeByKey gets specific scope by key from scopes.
func GetScopeByKey(c *gin.Context, key string) interface{} {
	scopeValue, exists := c.Get("scopes")
	if exists {
		return scopeValue.(map[string]interface{})[key]
	}
	return nil
}
