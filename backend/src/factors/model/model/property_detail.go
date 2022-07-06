package model

import (
	"errors"
	"fmt"
	"strings"
	"time"

	U "factors/util"

	C "factors/config"

	log "github.com/sirupsen/logrus"
)

// PropertyDetail implements property_details table
type PropertyDetail struct {
	// Composite primary key with project_id, event_name_id,key .
	ProjectID   int64     `gorm:"unique_index:configured_properties_project_id_event_name_id_key_unique_idx" json:"project_id"`
	EventNameID *string   `gorm:"unique_index:configured_properties_project_id_event_name_id_key_unique_idx" json:"event_name_id"`
	Key         string    `gorm:"unique_index:configured_properties_project_id_event_name_id_key_unique_idx" json:"key"`
	Type        string    `gorm:"not null" json:"type"`
	Entity      int       `gorm:"not null" json:"entity"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// cache properties type
const (
	TypeConfiguredDatetimeProperties  = "CDP"
	TypeConfiguredNumericalProperties = "CNP"
	TypeNonConfiguredProperties       = "NCP"
	TypeMissingConfiguredProperties   = "MCP"
)

// Entity indicates user or event entity
const (
	EntityEvent = 1
	EntityUser  = 2
)

// GetEntity return user or event
func GetEntity(isUserProperty bool) int {
	if isUserProperty {
		return EntityUser
	}

	return EntityEvent
}

// GetConfiguredEventPropertiesTypeCacheKey return configured property type cache key
func GetConfiguredEventPropertiesTypeCacheKey(projectID int64, eventName, propertyName, configuredType string) string {
	if projectID == 0 || eventName == "" || propertyName == "" || configuredType == "" {
		return ""
	}

	return fmt.Sprintf("%d:%s:%s:%s", projectID, eventName, propertyName, configuredType)
}

// GetNonConfiguredEventPropertiesTypeCacheKey return non configured property type cache key
func GetNonConfiguredEventPropertiesTypeCacheKey(projectID int64, eventName, propertyName string) string {
	if projectID == 0 || propertyName == "" || eventName == "" {
		return ""
	}

	return fmt.Sprintf("%d:%s:%s:%s", projectID, TypeNonConfiguredProperties, eventName, propertyName)
}

// GetConfiguredUserPropertiesTypeCacheKey return non configured user property datetime type cache key
func GetConfiguredUserPropertiesTypeCacheKey(projectID int64, propertyName, configuredType string) string {
	if projectID == 0 || propertyName == "" || configuredType == "" {
		return ""
	}

	return fmt.Sprintf("%d:%s:%s", projectID, propertyName, configuredType)
}

// GetNonConfiguredUserPropertiesTypeCacheKey return non configured user property type cache key
func GetNonConfiguredUserPropertiesTypeCacheKey(projectID int64, propertyName string) string {
	if projectID == 0 || propertyName == "" {
		return ""
	}

	return fmt.Sprintf("%d:%s:%s", projectID, TypeNonConfiguredProperties, propertyName)
}

func getCacheConfiguredUserProperties(projectID int64, propertyName string) string {
	propertiesTypeCache := C.GetPropertiesTypeCache()
	if propertiesTypeCache == nil {
		return ""
	}

	configuredPropertykey := GetConfiguredUserPropertiesTypeCacheKey(projectID, propertyName, TypeConfiguredDatetimeProperties)
	_, found := propertiesTypeCache.Cache.Get(configuredPropertykey)
	if found {
		return TypeConfiguredDatetimeProperties
	}

	configuredPropertykey = GetConfiguredUserPropertiesTypeCacheKey(projectID, propertyName, TypeConfiguredNumericalProperties)
	_, found = propertiesTypeCache.Cache.Get(configuredPropertykey)
	if found {
		return TypeConfiguredNumericalProperties
	}

	nonConfiguredPropertyKey := GetNonConfiguredUserPropertiesTypeCacheKey(projectID, propertyName)
	_, found = propertiesTypeCache.Cache.Get(nonConfiguredPropertyKey)
	if found {
		return TypeNonConfiguredProperties
	}

	return TypeMissingConfiguredProperties
}

func getCacheConfiguredEventProperties(projectID int64, eventName, propertyName string) string {
	propertiesTypeCache := C.GetPropertiesTypeCache()
	if propertiesTypeCache == nil {
		return ""
	}

	configuredPropertykey := GetConfiguredEventPropertiesTypeCacheKey(projectID, eventName, propertyName, TypeConfiguredDatetimeProperties)
	_, found := propertiesTypeCache.Cache.Get(configuredPropertykey)
	if found {
		return TypeConfiguredDatetimeProperties
	}

	configuredPropertykey = GetConfiguredEventPropertiesTypeCacheKey(projectID, eventName, propertyName, TypeConfiguredNumericalProperties)
	_, found = propertiesTypeCache.Cache.Get(configuredPropertykey)
	if found {
		return TypeConfiguredNumericalProperties
	}

	nonConfiguredPropertyKey := GetNonConfiguredEventPropertiesTypeCacheKey(projectID, eventName, propertyName)
	_, found = propertiesTypeCache.Cache.Get(nonConfiguredPropertyKey)
	if found {
		return TypeNonConfiguredProperties
	}

	return TypeMissingConfiguredProperties
}

// GetCachePropertiesType returns property type from cache, resets the property cache once every day. event_name
// will be sent empty for user_property
func GetCachePropertiesType(projectID int64, eventName, propertyName string, isUserProperty bool) string {
	propertiesTypeCache := C.GetPropertiesTypeCache()
	if propertiesTypeCache == nil {
		return TypeMissingConfiguredProperties
	}

	// Reset cache on start of day
	currentDate := U.GetDateOnlyFromTimestampZ(U.TimeNowUnix())
	if propertiesTypeCache.LastResetDate != currentDate {
		C.ResetPropertyDetailsCacheByDate(U.TimeNowUnix())
		return TypeMissingConfiguredProperties
	}

	if isUserProperty {
		return getCacheConfiguredUserProperties(projectID, propertyName)
	}

	return getCacheConfiguredEventProperties(projectID, eventName, propertyName)
}

// SetCachePropertiesType sets key to cache by property type. If user_property then event_name will be set as empty
func SetCachePropertiesType(projectID int64, eventName, propertyName, propertyType string, isUserProperty, isConfigured bool) error {
	propertiesTypeCache := C.GetPropertiesTypeCache()
	if propertiesTypeCache == nil {
		return nil
	}

	logCtx := log.WithFields(log.Fields{"project_id": projectID, "event_name": eventName, "property_name": propertyName, "property_type": propertyType})
	if projectID == 0 || propertyName == "" || propertyType == "" {
		logCtx.Error("missing required fields.")
		return errors.New("missing parameters")
	}

	if !isUserProperty && eventName == "" {
		logCtx.Error("missing eventName.")
		return errors.New("missing eventName")
	}

	if isUserProperty {
		eventName = ""
	}

	var cacheKey string
	if !isConfigured {
		if isUserProperty {
			cacheKey = GetNonConfiguredUserPropertiesTypeCacheKey(projectID, propertyName)
		} else {
			cacheKey = GetNonConfiguredEventPropertiesTypeCacheKey(projectID, eventName, propertyName)
		}

		if cacheKey == "" {
			logCtx.Error("Failed to get non configured properties cache key.")
			return errors.New("invalid non configured properties cache key")
		}

		propertiesTypeCache.Cache.Add(cacheKey, nil)
		return nil
	}

	var configurationType string
	if propertyType == U.PropertyTypeDateTime {
		configurationType = TypeConfiguredDatetimeProperties
	}
	if propertyType == U.PropertyTypeNumerical {
		configurationType = TypeConfiguredNumericalProperties
	}

	if isUserProperty {
		cacheKey = GetConfiguredUserPropertiesTypeCacheKey(projectID, propertyName, configurationType)
	} else {
		cacheKey = GetConfiguredEventPropertiesTypeCacheKey(projectID, eventName, propertyName, configurationType)
	}

	if cacheKey == "" {
		logCtx.Error("Failed to get configured properties cache key.")
		return errors.New("invalid configured properties cache key")
	}

	propertiesTypeCache.Cache.Add(cacheKey, nil)
	return nil
}

// ErrorUsingSalesforceDatetimeTemplate property value contains salesforece datetime template. Remove this check once no warning found
var ErrorUsingSalesforceDatetimeTemplate error = errors.New("salesforce datetime template detected")

// ValidateDateTimeProperty validates value by using known key and its value logic
func ValidateDateTimeProperty(key string, value interface{}) error {
	if key == "" || value == nil || value == "" {
		return nil
	}

	_, err := U.GetPropertyValueAsFloat64(value)
	if err == nil {
		return nil
	}

	_, err = time.Parse(HubspotDateTimeLayout, U.GetPropertyValueAsString(value))
	if err == nil {
		return nil
	}

	if strings.HasPrefix(key, U.SALESFORCE_PROPERTY_PREFIX) {
		_, err := GetSalesforceDocumentTimestamp(value) // make sure timezone info is loaded to the container
		if err != nil {
			return err
		}

		return ErrorUsingSalesforceDatetimeTemplate
	}

	return errors.New("invalid value for datetime property")
}
