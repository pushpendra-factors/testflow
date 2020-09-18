package tests

import (
	SDK "factors/sdk"
	U "factors/util"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetPropertyTypeByKeyValue(t *testing.T) {
	assert.Equal(t, U.PropertyTypeCategorical, U.GetPropertyTypeByKeyValue("testKey", "10.24string"))
	assert.Equal(t, U.PropertyTypeCategorical, U.GetPropertyTypeByKeyValue("testKey", "10.24"))
	assert.Equal(t, U.PropertyTypeNumerical, U.GetPropertyTypeByKeyValue("testKey", 10.24))
	assert.Equal(t, U.PropertyTypeUnknown, U.GetPropertyTypeByKeyValue("testKey", true))
	assert.Equal(t, U.PropertyTypeUnknown, U.GetPropertyTypeByKeyValue("testKey", []string{"value1", "value2"}))

	// numerical property by name.
	assert.Equal(t, U.PropertyTypeNumerical, U.GetPropertyTypeByKeyValue(U.EP_PAGE_LOAD_TIME, "10.24"))
	assert.Equal(t, U.PropertyTypeNumerical, U.GetPropertyTypeByKeyValue(U.EP_PAGE_SPENT_TIME, 1234))
	assert.Equal(t, U.PropertyTypeNumerical, U.GetPropertyTypeByKeyValue(U.EP_REVENUE, "10"))
	// categorical property by name.
	assert.Equal(t, U.PropertyTypeNumerical, U.GetPropertyTypeByKeyValue(U.EP_CAMPAIGN, 10.24)) // This will be classified as numerical now since the default logic is removed
	assert.Equal(t, U.PropertyTypeCategorical, U.GetPropertyTypeByKeyValue(U.EP_CAMPAIGN_ID, "10.24"))
}

func TestGetCleanPropertyValue(t *testing.T) {
	assert.Equal(t, "value with?reserved+characters$are(properties)",
		U.GetUnEscapedPropertyValue("value%20with%3Freserved%2Bcharacters%24are%28properties%29"))
}

func TestGetPropertyValueAsString(t *testing.T) {
	var value interface{}
	value = float64(6444173670)
	assert.Equal(t, "6444173670", U.GetPropertyValueAsString(value))

	value = int64(6444173670)
	assert.Equal(t, "6444173670", U.GetPropertyValueAsString(value))

	value = "google"
	assert.Equal(t, "google", U.GetPropertyValueAsString(value))

	value = true
	assert.Equal(t, "true", U.GetPropertyValueAsString(value))
}

func TestFillUserAgentUserProperties(t *testing.T) {
	userProperties := make(U.PropertiesMap, 0)
	userAgent := "Mozilla/5.0 (Linux; Android 6.0.1; Nexus 5X Build/MMB29P) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/41.0.2272.96 Mobile Safari/537.36 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)"
	SDK.FillUserAgentUserProperties(&userProperties, userAgent)
	assert.NotNil(t, userProperties[U.UP_USER_AGENT])
	assert.Equal(t, userAgent, userProperties[U.UP_USER_AGENT])
	assert.NotNil(t, userProperties[U.UP_BROWSER])
	assert.Equal(t, "Bot", userProperties[U.UP_BROWSER])

	newUserProperties := make(U.PropertiesMap, 0)
	userAgent = "Mozilla/5.0 (iPhone; CPU iPhone OS 12_2 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148"
	SDK.FillUserAgentUserProperties(&newUserProperties, userAgent)
	assert.NotNil(t, newUserProperties[U.UP_DEVICE_BRAND])
	assert.NotNil(t, newUserProperties[U.UP_DEVICE_MODEL])
	assert.NotNil(t, newUserProperties[U.UP_DEVICE_TYPE])
	assert.NotEqual(t, "Bot", newUserProperties[U.UP_BROWSER])
}

func TestFillFirstEventUserPropertiesIfNotExist(t *testing.T) {
	existingUserProperties := make(map[string]interface{}, 0)
	newUserProperties := make(U.PropertiesMap, 0)
	eventTimestamp := time.Now().Unix()
	_ = U.FillFirstEventUserPropertiesIfNotExist(&existingUserProperties,
		&newUserProperties, eventTimestamp)
	assert.NotNil(t, (newUserProperties)[U.UP_DAY_OF_FIRST_EVENT])
	assert.Equal(t, (newUserProperties)[U.UP_DAY_OF_FIRST_EVENT],
		time.Unix(eventTimestamp, 0).Weekday().String())
	hourOfFirstEvent, _, _ := time.Unix(eventTimestamp, 0).Clock()
	assert.NotNil(t, (newUserProperties)[U.UP_HOUR_OF_FIRST_EVENT])
	assert.Equal(t, (newUserProperties)[U.UP_HOUR_OF_FIRST_EVENT], hourOfFirstEvent)
}
