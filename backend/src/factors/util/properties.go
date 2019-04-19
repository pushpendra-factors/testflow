package util

import (
	"encoding/json"
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
)

// Common properties type.
type PropertiesMap map[string]interface{}

// Special Event Names used when building patterns and for querying.
const SEN_ALL_ACTIVE_USERS = "$AllActiveUsers"
const SEN_ALL_ACTIVE_USERS_DISPLAY_STRING = "All Active Users"

/* Properties Constants */

// Event Properties.
var EP_OCCURRENCE_COUNT string = "$occurrenceCount"

// Default Event Properites
var EP_INTERNAL_IP string = "$ip"
var EP_LOCATION_LATITUDE string = "$locationLat"
var EP_LOCATION_LONGITUDE string = "$locationLng"
var EP_REFERRER string = "$referrer"
var EP_PAGE_TITLE string = "$pageTitle"
var EP_RAW_URL string = "$rawURL"

// Default User Properties
var UP_PLATFORM string = "$platform"
var UP_BROWSER string = "$browser"
var UP_BROWSER_VERSION string = "$browserVersion"
var UP_USER_AGENT string = "$userAgent"
var UP_OS string = "$os"
var UP_OS_VERSION string = "$osVersion"
var UP_SCREEN_WIDTH string = "$screenWidth"
var UP_SCREEN_HEIGHT string = "$screenHeight"
var UP_SCREEN_DENSITY string = "$screenDensity"
var UP_LANGUAGE string = "$language"
var UP_LOCALE string = "$locale"
var UP_DEVICE_ID string = "$deviceId"
var UP_DEVICE_NAME string = "$deviceName"
var UP_DEVICE_ADVERTISING_ID string = "$deviceAdvertisingId"
var UP_DEVICE_BRAND string = "$deviceBrand"
var UP_DEVICE_MODEL string = "$deviceModel"
var UP_DEVICE_TYPE string = "$deviceType"
var UP_DEVICE_FAMILY string = "$deviceFamily"
var UP_DEVICE_MANUFACTURER string = "$deviceManufacturer"
var UP_DEVICE_CARRIER string = "$deviceCarrier"
var UP_DEVICE_ADTRACKING_ENABLED string = "$deviceAdTrackingEnabled"
var UP_NETWORK_BLUETOOTH string = "$networkBluetooth"
var UP_NETWORK_CARRIER string = "$networkCarrier"
var UP_NETWORK_CELLULAR string = "$networkCellular"
var UP_NETWORK_WIFI string = "$networkWifi"
var UP_APP_NAME string = "$appName"
var UP_APP_NAMESPACE string = "$appNamespace"
var UP_APP_VERSION string = "$appVersion"
var UP_APP_BUILD string = "$appBuild"
var UP_COUNTRY string = "$country"
var UP_CITY string = "$city"
var UP_REGION string = "$region"
var UP_CAMPAIGN_NAME string = "$campaignName"
var UP_CAMPAIGN_SOURCE string = "$campaignSource"
var UP_CAMPAIGN_MEDIUM string = "$campaignMedium"
var UP_CAMPAIGN_TERM string = "$campaignTerm"
var UP_CAMPAIGN_CONTENT string = "$campaignContent"
var UP_TIMEZONE string = "$timezone"

var ALLOWED_SDK_DEFAULT_EVENT_PROPERTIES = [...]string{
	EP_INTERNAL_IP,
	EP_LOCATION_LATITUDE,
	EP_LOCATION_LONGITUDE,
	EP_REFERRER,
	EP_PAGE_TITLE,
	EP_RAW_URL,
}

// Event properties that are not visible to user for analysis.
var INTERNAL_EVENT_PROPERTIES = [...]string{
	EP_INTERNAL_IP,
	EP_LOCATION_LATITUDE,
	EP_LOCATION_LONGITUDE,
}

var ALLOWED_SDK_DEFAULT_USER_PROPERTIES = [...]string{
	UP_PLATFORM,
	UP_BROWSER,
	UP_BROWSER_VERSION,
	UP_USER_AGENT,
	UP_OS,
	UP_OS_VERSION,
	UP_SCREEN_WIDTH,
	UP_SCREEN_HEIGHT,
	UP_SCREEN_DENSITY,
	UP_LANGUAGE,
	UP_LOCALE,
	UP_DEVICE_ID,
	UP_DEVICE_NAME,
	UP_DEVICE_ADVERTISING_ID,
	UP_DEVICE_BRAND,
	UP_DEVICE_MODEL,
	UP_DEVICE_TYPE,
	UP_DEVICE_FAMILY,
	UP_DEVICE_MANUFACTURER,
	UP_DEVICE_CARRIER,
	UP_DEVICE_ADTRACKING_ENABLED,
	UP_NETWORK_BLUETOOTH,
	UP_NETWORK_CARRIER,
	UP_NETWORK_CELLULAR,
	UP_NETWORK_WIFI,
	UP_APP_NAME,
	UP_APP_NAMESPACE,
	UP_APP_VERSION,
	UP_APP_BUILD,
	UP_COUNTRY,
	UP_CITY,
	UP_REGION,
	UP_CAMPAIGN_NAME,
	UP_CAMPAIGN_SOURCE,
	UP_CAMPAIGN_MEDIUM,
	UP_CAMPAIGN_TERM,
	UP_CAMPAIGN_CONTENT,
	UP_TIMEZONE,
}

// Event properties that are not visible to user for analysis.
var INTERNAL_USER_PROPERTIES = [...]string{
	UP_DEVICE_ID,
	"_$deviceId", // Here for legacy reason.
}

var VISIBLE_DEFAULT_NUMERIC_EVENT_PROPERTIES = [...]string{EP_OCCURRENCE_COUNT}

const NAME_PREFIX = "$"
const NAME_PREFIX_ESCAPE_CHAR = "_"
const QUERY_PARAM_PROPERTY_PREFIX = "$qp_"

// Platforms
const PLATFORM_WEB = "web"

const (
	PropertyTypeNumerical   = "numerical"
	PropertyTypeCategorical = "categorical"
)

const SamplePropertyValuesLimit = 100

// isValidProperty - Validate property type.
func isPropertyTypeValid(value interface{}) error {
	switch valueType := value.(type) {
	case float64:
	case string:
	case bool:
	default:
		log.WithFields(log.Fields{"value": value,
			"valueType": valueType}).Debug("Invalid type used on property")
		return fmt.Errorf("invalid property type")
	}
	return nil
}

func isSDKUserDefaultProperty(key *string) bool {
	for _, k := range ALLOWED_SDK_DEFAULT_USER_PROPERTIES {
		if k == *key {
			return true
		}
	}
	return false
}

func isSDKEventDefaultProperty(key *string) bool {
	for _, k := range ALLOWED_SDK_DEFAULT_EVENT_PROPERTIES {
		if k == *key {
			return true
		}
	}
	return false
}

func IsInternalEventProperty(key *string) bool {
	for _, k := range INTERNAL_EVENT_PROPERTIES {
		if k == *key {
			return true
		}
	}
	return false
}

func IsInternalUserProperty(key *string) bool {
	for _, k := range INTERNAL_USER_PROPERTIES {
		if k == *key {
			return true
		}
	}
	return false
}

func GetValidatedUserProperties(properties *PropertiesMap) *PropertiesMap {
	validatedProperties := make(PropertiesMap)
	for k, v := range *properties {
		if err := isPropertyTypeValid(v); err == nil {
			if strings.HasPrefix(k, NAME_PREFIX) && !isSDKUserDefaultProperty(&k) {
				validatedProperties[fmt.Sprintf("%s%s", NAME_PREFIX_ESCAPE_CHAR, k)] = v
			} else {
				validatedProperties[k] = v
			}
		}
	}
	return &validatedProperties
}

func GetValidatedEventProperties(properties *PropertiesMap) *PropertiesMap {
	validatedProperties := make(PropertiesMap)
	for k, v := range *properties {
		if err := isPropertyTypeValid(v); err == nil {
			// Escape properties with $ prefix but allow query_params_props with $qp_ prrefix and default properties.
			if strings.HasPrefix(k, NAME_PREFIX) &&
				!strings.HasPrefix(k, QUERY_PARAM_PROPERTY_PREFIX) &&
				!isSDKEventDefaultProperty(&k) {
				validatedProperties[fmt.Sprintf("%s%s", NAME_PREFIX_ESCAPE_CHAR, k)] = v
			} else {
				validatedProperties[k] = v
			}
		}
	}
	return &validatedProperties
}

// ClassifyPropertiesByType - Classifies categorical and numerical properties
// by checking type of values. properties -> map[propertyKey]map[propertyValue]true
func ClassifyPropertiesByType(properties *map[string]map[interface{}]bool) (map[string][]string, error) {
	numProperties := make([]string, 0, 0)
	catProperties := make([]string, 0, 0)

	for propertyKey, v := range *properties {
		isNumericalProperty := true
		for propertyValue := range v {
			switch t := propertyValue.(type) {
			case int, float64:
			case string:
				if !IsNumber(propertyValue.(string)) {
					isNumericalProperty = false
				}
			default:
				return nil, fmt.Errorf("unsupported type %s on property type classification", t)
			}
		}

		if isNumericalProperty {
			numProperties = append(numProperties, propertyKey)
		} else {
			catProperties = append(catProperties, propertyKey)
		}
	}

	propsByType := make(map[string][]string, 0)
	propsByType[PropertyTypeNumerical] = numProperties
	propsByType[PropertyTypeCategorical] = catProperties

	return propsByType, nil
}

// FillPropertyKvsFromPropertiesJson - Fills properties key with limited
// no.of of values propertiesKvs -> map[propertyKey]map[propertyValue]true
func FillPropertyKvsFromPropertiesJson(propertiesJson []byte,
	propertiesKvs *map[string]map[interface{}]bool, valuesLimit int) error {
	var rowProperties map[string]interface{}
	err := json.Unmarshal(propertiesJson, &rowProperties)
	if err != nil {
		return err
	}

	for k, v := range rowProperties {
		if _, ok := (*propertiesKvs)[k]; !ok {
			(*propertiesKvs)[k] = make(map[interface{}]bool, 0)
		}
		if len((*propertiesKvs)[k]) < valuesLimit {
			(*propertiesKvs)[k][v] = true
		}
	}

	return nil
}
