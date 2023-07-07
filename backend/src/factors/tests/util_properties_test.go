package tests

import (
	"factors/model/model"
	"factors/model/store"
	SDK "factors/sdk"
	U "factors/util"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetPropertyTypeByKeyValue(t *testing.T) {
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)
	assert.Equal(t, U.PropertyTypeCategorical, store.GetStore().GetPropertyTypeByKeyValue(project.ID, "event1", "testKey", "10.24string", false))
	assert.Equal(t, U.PropertyTypeCategorical, store.GetStore().GetPropertyTypeByKeyValue(project.ID, "event1", "testKey", "10.24", false))
	assert.Equal(t, U.PropertyTypeNumerical, store.GetStore().GetPropertyTypeByKeyValue(project.ID, "event1", "testKey", 10.24, false))
	assert.Equal(t, U.PropertyTypeUnknown, store.GetStore().GetPropertyTypeByKeyValue(project.ID, "event1", "testKey", true, false))
	assert.Equal(t, U.PropertyTypeUnknown, store.GetStore().GetPropertyTypeByKeyValue(project.ID, "event1", "testKey", []string{"value1", "value2"}, false))

	// numerical property by name.
	assert.Equal(t, U.PropertyTypeNumerical, store.GetStore().GetPropertyTypeByKeyValue(project.ID, "event1", U.EP_PAGE_LOAD_TIME, "10.24", false))
	assert.Equal(t, U.PropertyTypeNumerical, store.GetStore().GetPropertyTypeByKeyValue(project.ID, "event1", U.EP_PAGE_SPENT_TIME, 1234, false))
	assert.Equal(t, U.PropertyTypeNumerical, store.GetStore().GetPropertyTypeByKeyValue(project.ID, "event1", U.EP_REVENUE, "10", false))
	// categorical property by name.
	assert.Equal(t, U.PropertyTypeNumerical, store.GetStore().GetPropertyTypeByKeyValue(project.ID, "event1", U.EP_CAMPAIGN, 10.24, false)) // This will be classified as numerical now since the default logic is remov,falseed
	assert.Equal(t, U.PropertyTypeCategorical, store.GetStore().GetPropertyTypeByKeyValue(project.ID, "event1", U.EP_CAMPAIGN_ID, "10.24", false))
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
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)
	userProperties := make(U.PropertiesMap, 0)
	userAgent := "Mozilla/5.0 (Linux; Android 6.0.1; Nexus 5X Build/MMB29P) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/41.0.2272.96 Mobile Safari/537.36 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)"
	SDK.FillUserAgentUserProperties(project.ID, &userProperties, userAgent)
	assert.NotNil(t, userProperties[U.UP_USER_AGENT])
	assert.Equal(t, userAgent, userProperties[U.UP_USER_AGENT])
	assert.NotNil(t, userProperties[U.UP_BROWSER])
	assert.Equal(t, "Bot", userProperties[U.UP_BROWSER])

	newUserProperties := make(U.PropertiesMap, 0)
	userAgent = "Mozilla/5.0 (iPhone; CPU iPhone OS 12_2 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148"
	SDK.FillUserAgentUserProperties(project.ID, &newUserProperties, userAgent)
	assert.NotNil(t, newUserProperties[U.UP_DEVICE_BRAND])
	assert.NotNil(t, newUserProperties[U.UP_DEVICE_MODEL])
	assert.NotNil(t, newUserProperties[U.UP_DEVICE_TYPE])
	assert.NotEqual(t, "Bot", newUserProperties[U.UP_BROWSER])

	UserProperties1 := make(U.PropertiesMap, 0)
	userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) browser/2020.2.2 Chrome/78.0.3904.130 Electron/7.3.2 Safari/537.36 PingdomTMS/2020.2"
	SDK.FillUserAgentUserProperties(project.ID, &UserProperties1, userAgent)
	assert.NotNil(t, UserProperties1[U.UP_USER_AGENT])
	assert.Equal(t, userAgent, UserProperties1[U.UP_USER_AGENT])
	assert.NotNil(t, UserProperties1[U.UP_BROWSER])
	assert.Equal(t, "Bot", UserProperties1[U.UP_BROWSER])
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

func TestPropertyDetails(t *testing.T) {
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)
	eventName := "eventName1"
	dateTimeProperty1 := "dt_property1"
	dateTimeProperty2 := "dt_property2"
	/*
		Configured event property test
	*/

	propertyType := model.GetCachePropertiesType(project.ID, eventName, dateTimeProperty1, false)
	assert.Equal(t, model.TypeMissingConfiguredProperties, propertyType)
	propertyType = model.GetCachePropertiesType(project.ID, eventName, dateTimeProperty2, true)
	assert.Equal(t, model.TypeMissingConfiguredProperties, propertyType)

	// creating event property without registered event name
	status := store.GetStore().CreatePropertyDetails(project.ID, eventName, dateTimeProperty1, U.PropertyTypeDateTime, false, false)
	assert.Equal(t, http.StatusBadRequest, status)

	// creating event property with registered event name
	_, status = store.GetStore().CreateOrGetEventName(&model.EventName{
		ProjectId: project.ID,
		Name:      eventName,
		Type:      model.TYPE_USER_CREATED_EVENT_NAME,
	})
	assert.Equal(t, http.StatusCreated, status)
	status = store.GetStore().CreatePropertyDetails(project.ID, eventName, dateTimeProperty1, U.PropertyTypeDateTime, false, false)
	assert.Equal(t, http.StatusCreated, status)
	// duplicate configured property
	status = store.GetStore().CreatePropertyDetails(project.ID, eventName, dateTimeProperty1, U.PropertyTypeDateTime, false, false)
	assert.Equal(t, http.StatusConflict, status)

	/*
		Configured user property test
	*/

	status = store.GetStore().CreatePropertyDetails(project.ID, "", dateTimeProperty2, U.PropertyTypeDateTime, true, false)
	assert.Equal(t, http.StatusCreated, status)
	status = store.GetStore().CreatePropertyDetails(project.ID, "", dateTimeProperty2, U.PropertyTypeDateTime, true, false)
	assert.Equal(t, http.StatusConflict, status)

	// numerical property
	numericalProperty1 := "num_property1"
	numericalProperty2 := "num_property2"
	status = store.GetStore().CreatePropertyDetails(project.ID, eventName, numericalProperty1, U.PropertyTypeNumerical, false, false)
	assert.Equal(t, http.StatusCreated, status)
	status = store.GetStore().CreatePropertyDetails(project.ID, eventName, numericalProperty1, U.PropertyTypeNumerical, false, false)
	assert.Equal(t, http.StatusConflict, status)

	status = store.GetStore().CreatePropertyDetails(project.ID, "", numericalProperty2, U.PropertyTypeNumerical, true, false)
	assert.Equal(t, http.StatusCreated, status)
	status = store.GetStore().CreatePropertyDetails(project.ID, "", numericalProperty2, U.PropertyTypeNumerical, true, false)
	assert.Equal(t, http.StatusConflict, status)

	category := store.GetStore().GetPropertyTypeByKeyValue(project.ID, eventName, dateTimeProperty1, 123, false)
	assert.Equal(t, U.PropertyTypeDateTime, category)
	category = store.GetStore().GetPropertyTypeByKeyValue(project.ID, eventName, dateTimeProperty1, "123", false)
	assert.Equal(t, U.PropertyTypeDateTime, category)

	category = store.GetStore().GetPropertyTypeByKeyValue(project.ID, "", dateTimeProperty2, 123, true)
	assert.Equal(t, U.PropertyTypeDateTime, category)
	category = store.GetStore().GetPropertyTypeByKeyValue(project.ID, "", dateTimeProperty2, "123", true)
	assert.Equal(t, U.PropertyTypeDateTime, category)

	category = store.GetStore().GetPropertyTypeByKeyValue(project.ID, eventName, numericalProperty1, 123, false)
	assert.Equal(t, U.PropertyTypeNumerical, category)
	category = store.GetStore().GetPropertyTypeByKeyValue(project.ID, "", numericalProperty2, "123", true)
	assert.Equal(t, U.PropertyTypeNumerical, category)

	// missing property type but with prefix should be treadted as categorical
	category = store.GetStore().GetPropertyTypeByKeyValue(project.ID, eventName, U.SALESFORCE_PROPERTY_PREFIX+"updated_batch_no_c", nil, false)
	assert.Equal(t, U.PropertyTypeCategorical, category)

	/*
		Get from cache
	*/
	propertyType = model.GetCachePropertiesType(project.ID, eventName, dateTimeProperty1, false)
	assert.Equal(t, model.TypeConfiguredDatetimeProperties, propertyType)
	propertyType = model.GetCachePropertiesType(project.ID, "", dateTimeProperty2, true)
	assert.Equal(t, model.TypeConfiguredDatetimeProperties, propertyType)

	propertyType = model.GetCachePropertiesType(project.ID, eventName, numericalProperty1, false)
	assert.Equal(t, model.TypeConfiguredNumericalProperties, propertyType)
	propertyType = model.GetCachePropertiesType(project.ID, "", numericalProperty2, true)
	assert.Equal(t, model.TypeConfiguredNumericalProperties, propertyType)

	propertyType = model.GetCachePropertiesType(project.ID, eventName, "property2", false)
	assert.Equal(t, model.TypeMissingConfiguredProperties, propertyType)
}

func TestOverwritePropertyDetails(t *testing.T) {
	project, err := SetupProjectReturnDAO()
	assert.Nil(t, err)
	eventName := "eventName1"
	dateTimeProperty1 := "dt_property1"

	// creating event property with registered event name
	_, status := store.GetStore().CreateOrGetEventName(&model.EventName{
		ProjectId: project.ID,
		Name:      eventName,
		Type:      model.TYPE_USER_CREATED_EVENT_NAME,
	})

	assert.Equal(t, http.StatusCreated, status)

	// create property details with overwrite
	status = store.GetStore().CreatePropertyDetails(project.ID, eventName, dateTimeProperty1, U.PropertyTypeDateTime, false, true)
	assert.Equal(t, http.StatusCreated, status)
	_, propertyDetails := store.GetStore().GetPropertyTypeFromDB(project.ID, eventName, dateTimeProperty1, false)
	assert.Equal(t, U.PropertyTypeDateTime, propertyDetails.Type)
	// overwrite configured property
	status = store.GetStore().CreatePropertyDetails(project.ID, eventName, dateTimeProperty1, U.PropertyTypeNumerical, false, true)
	assert.Equal(t, http.StatusAccepted, status)
	_, propertyDetails = store.GetStore().GetPropertyTypeFromDB(project.ID, eventName, dateTimeProperty1, false)
	assert.Equal(t, U.PropertyTypeNumerical, propertyDetails.Type)

	// delete property details if type is unknown
	err = store.GetStore().CreateOrDeletePropertyDetails(project.ID, eventName, dateTimeProperty1, U.PropertyTypeUnknown, false, true)
	assert.Nil(t, err)
	status, propertyDetails = store.GetStore().GetPropertyTypeFromDB(project.ID, eventName, dateTimeProperty1, false)
	assert.Equal(t, http.StatusNotFound, status)

	// use CreateOrDeletePropertyDetails for creating new property details and overwrite test
	err = store.GetStore().CreateOrDeletePropertyDetails(project.ID, eventName, dateTimeProperty1, U.PropertyTypeDateTime, false, true)
	assert.Nil(t, err)
	status, propertyDetails = store.GetStore().GetPropertyTypeFromDB(project.ID, eventName, dateTimeProperty1, false)
	assert.Equal(t, http.StatusFound, status)
	assert.Equal(t, U.PropertyTypeDateTime, propertyDetails.Type)

	err = store.GetStore().CreateOrDeletePropertyDetails(project.ID, eventName, dateTimeProperty1, U.PropertyTypeNumerical, false, true)
	assert.Nil(t, err)
	status, propertyDetails = store.GetStore().GetPropertyTypeFromDB(project.ID, eventName, dateTimeProperty1, false)
	assert.Equal(t, http.StatusFound, status)
	assert.Equal(t, U.PropertyTypeNumerical, propertyDetails.Type)

	/*
		Test conflict
	*/
	// event property conflict without overwrite
	status = store.GetStore().CreatePropertyDetails(project.ID, eventName, dateTimeProperty1, U.PropertyTypeNumerical, false, false)
	assert.Equal(t, http.StatusConflict, status)
	// event property conflict witt overwrite
	status = store.GetStore().CreatePropertyDetails(project.ID, eventName, dateTimeProperty1, U.PropertyTypeNumerical, false, true)
	assert.Equal(t, http.StatusConflict, status)
	// user property conflict witt overwrite
	status = store.GetStore().CreatePropertyDetails(project.ID, "", dateTimeProperty1, U.PropertyTypeDateTime, true, true)
	assert.Equal(t, http.StatusCreated, status)
	status = store.GetStore().CreatePropertyDetails(project.ID, "", dateTimeProperty1, U.PropertyTypeDateTime, true, true)
	assert.Equal(t, http.StatusConflict, status)

	/*
		Test delete property details
	*/
	dateTimeProperty2 := "dt_property2"
	err = store.GetStore().CreateOrDeletePropertyDetails(project.ID, eventName, dateTimeProperty2, U.PropertyTypeDateTime, false, true)
	assert.Nil(t, err)
	status, propertyDetails = store.GetStore().GetPropertyTypeFromDB(project.ID, eventName, dateTimeProperty2, false)
	assert.Equal(t, http.StatusFound, status)
	status, propertyDetails = store.GetStore().GetPropertyTypeFromDB(project.ID, eventName, dateTimeProperty2, true)
	assert.Equal(t, http.StatusNotFound, status)

	err = store.GetStore().CreateOrDeletePropertyDetails(project.ID, eventName, dateTimeProperty2, U.PropertyTypeDateTime, true, true)
	assert.Nil(t, err)
	status, propertyDetails = store.GetStore().GetPropertyTypeFromDB(project.ID, eventName, dateTimeProperty2, false)
	assert.Equal(t, http.StatusFound, status)
	status, propertyDetails = store.GetStore().GetPropertyTypeFromDB(project.ID, eventName, dateTimeProperty2, true)
	assert.Equal(t, http.StatusFound, status)

	// delete event property details for dt_property2
	err = store.GetStore().CreateOrDeletePropertyDetails(project.ID, eventName, dateTimeProperty2, U.PropertyTypeUnknown, false, true)
	assert.Nil(t, err)
	// event property deleted
	status, propertyDetails = store.GetStore().GetPropertyTypeFromDB(project.ID, eventName, dateTimeProperty2, false)
	assert.Equal(t, http.StatusNotFound, status)
	// user property should exist
	status, propertyDetails = store.GetStore().GetPropertyTypeFromDB(project.ID, eventName, dateTimeProperty2, true)
	assert.Equal(t, http.StatusFound, status)
	// event property for dt_property1 shouldn't be deleted
	status, propertyDetails = store.GetStore().GetPropertyTypeFromDB(project.ID, eventName, dateTimeProperty1, false)
	assert.Equal(t, http.StatusFound, status)

	// delete user property details for dt_property2
	err = store.GetStore().CreateOrDeletePropertyDetails(project.ID, eventName, dateTimeProperty2, U.PropertyTypeUnknown, false, true)
	assert.Nil(t, err)
	// event property deleted
	status, propertyDetails = store.GetStore().GetPropertyTypeFromDB(project.ID, eventName, dateTimeProperty2, false)
	assert.Equal(t, http.StatusNotFound, status)
	// event property for dt_property1
	status, propertyDetails = store.GetStore().GetPropertyTypeFromDB(project.ID, eventName, dateTimeProperty1, false)
	assert.Equal(t, http.StatusFound, status)

}
