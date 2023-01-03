package util

import (
	"errors"
	"strings"

	"github.com/mssola/user_agent"

	"github.com/gin-gonic/gin"
)

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

// GetScopeByKeyAsUint64 gets scope by key and value of type uint64.
func GetScopeByKeyAsUint64(c *gin.Context, key string) uint64 {
	intrfce := GetScopeByKey(c, key)
	if intrfce == nil {
		return 0
	}
	return intrfce.(uint64)
}

func GetScopeByKeyAsInt64(c *gin.Context, key string) int64 {
	intrfce := GetScopeByKey(c, key)
	if intrfce == nil {
		return 0
	}
	return intrfce.(int64)
}

func GetScopeByKeyAsString(c *gin.Context, key string) string {
	iface := GetScopeByKey(c, key)
	if iface == nil {
		return ""
	}
	return iface.(string)
}

func GetScopeByKeyAsBool(c *gin.Context, key string) bool {
	iface := GetScopeByKey(c, key)
	if iface == nil {
		return false
	}
	return iface.(bool)
}

// GetRequestSubdomain returns sample on sample.factors.ai
func GetRequestSubdomain(host string) (string, error) {
	splitHost := strings.Split(host, ".")

	if len(splitHost) != 3 {
		return "", errors.New("invalid subdomain on the request host")
	}

	return splitHost[0], nil
}

// IsRequestFromLocalhost - Check localhost on dev environment.
func IsRequestFromLocalhost(host string) bool {
	// ServeHTTP on tests doesn't set host.
	if host == "" {
		return true
	}

	splitHost := strings.Split(host, ":")

	if len(splitHost) > 2 || splitHost[0] != "localhost" {
		return false
	}

	return true
}

// IsPinggdomBot - Check whether it is pingdom bot or not
func IsPingdomBot(userAgent string) bool {
	return strings.Contains(strings.ToLower(userAgent), "pingdom")
}

// IsLighthouse - Check whether it is lighthouse useragent or not.
func IsLighthouse(userAgent string) bool {
	return strings.Contains(strings.ToLower(userAgent), "lighthouse")
}

// IsBotUserAgent - Check request user agent is bot or not.
func IsBotUserAgent(userAgent string) bool {
	if userAgent == "" {
		return false
	}

	if IsPingdomBot(userAgent) || IsLighthouse(userAgent) {
		return true
	}

	return user_agent.New(userAgent).Bot()
}

// IsBotEventByPrefix - Checks for event_name with selected bot prefix.
// gtm-msr.appspot.com/render2 is a pageview recorded in multiple
// projects (Including factors) upto Dec 29, 2022.
func IsBotEventByPrefix(eventName string) bool {
	return strings.HasPrefix(eventName, "gtm-msr.appspot.com")
}
