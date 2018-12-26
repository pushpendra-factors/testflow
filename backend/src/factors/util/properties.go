package util

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
)

// Common properties type.
type PropertiesMap map[string]interface{}

/* Properties Constants */

// Event Properties.
var EP_OCCURRENCE_COUNT string = "$occurrenceCount"

// Properties that change too often are better part of event properties rather than user property.
// TODO(dineshprabhu): Internal property not to be exposed on Frontend / any http responses.
var EP_INTERNAL_IP = "$ip"
var EP_LOCATION_LATITUDE string = "$locationLat"
var EP_LOCATION_LONGITUDE string = "$locationLng"

// Default User Properties - Added from JS SDK
var UP_PLATFORM string = "$platform"
var UP_REFERRER string = "$referrer"
var UP_BROWSER string = "$browser"
var UP_BROWSER_VERSION string = "$browserVersion"
var UP_OS string = "$os"
var UP_OS_VERSION string = "$osVersion"
var UP_SCREEN_WIDTH string = "$screenWidth"
var UP_SCREEN_HEIGHT string = "$screenHeight"
var UP_LANGUAGE string = "$language"
var UP_DEVICE_BRAND string = "$deviceBrand"
var UP_DEVICE_MODEL string = "$deviceModel"
var UP_DEVICE_TYPE string = "$deviceType"
var UP_DEVICE_FAMILY string = "$deviceFamily"
var UP_DEVICE_MANUFACTURER string = "$deviceManufacturer"
var UP_DEVICE_ID string = "$deviceId"
var UP_DEVICE_CARRIER string = "$deviceCarrier"
var UP_APP_VERSION string = "$appVersion"

// Default User Properties - Added from Backend
var UP_COUNTRY string = "$country"
var UP_CITY string = "$city"
var UP_REGION string = "$region"

var ALLOWED_SDK_DEFAULT_EVENT_PROPERTIES = [...]string{
	EP_INTERNAL_IP,
	EP_LOCATION_LATITUDE,
	EP_LOCATION_LONGITUDE,
}

var ALLOWED_SDK_DEFAULT_USER_PROPERTIES = [...]string{
	UP_PLATFORM,
	UP_REFERRER,
	UP_BROWSER,
	UP_BROWSER_VERSION,
	UP_OS,
	UP_OS_VERSION,
	UP_SCREEN_WIDTH,
	UP_SCREEN_HEIGHT,
	UP_COUNTRY,
	UP_CITY,
	UP_REGION,
	UP_LANGUAGE,
	UP_DEVICE_BRAND,
	UP_DEVICE_MODEL,
	UP_DEVICE_TYPE,
	UP_DEVICE_FAMILY,
	UP_DEVICE_MANUFACTURER,
	UP_DEVICE_ID,
	UP_DEVICE_CARRIER,
	UP_APP_VERSION,
}

var DEFAULT_NUMERIC_EVENT_PROPERTIES = [...]string{EP_OCCURRENCE_COUNT}

const NAME_PREFIX = "$"
const NAME_PREFIX_ESCAPE_CHAR = "_"
const QUERY_PARAM_PROPERTY_PREFIX = "$qp_"

// isValidProperty - Validate property type.
func isPropertyTypeValid(value interface{}) error {
	switch valueType := value.(type) {
	case float64:
	case string:
	default:
		log.WithFields(log.Fields{"value": value,
			"valueType": valueType}).Debug("Invalid type used on property")
		return fmt.Errorf("invalid property type")
	}
	return nil
}

func hasSDKUserDefaultProperty(key *string) bool {
	for _, k := range ALLOWED_SDK_DEFAULT_USER_PROPERTIES {
		if k == *key {
			return true
		}
	}
	return false
}

func hasSDKEventDefaultProperty(key *string) bool {
	for _, k := range ALLOWED_SDK_DEFAULT_EVENT_PROPERTIES {
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
			if strings.HasPrefix(k, NAME_PREFIX) && !hasSDKUserDefaultProperty(&k) {
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
			// Escape properties with $ prefix but allow query_params_props with $qp_ prrefix.
			if strings.HasPrefix(k, NAME_PREFIX) && !strings.HasPrefix(k, QUERY_PARAM_PROPERTY_PREFIX) && !hasSDKEventDefaultProperty(&k) {
				validatedProperties[fmt.Sprintf("%s%s", NAME_PREFIX_ESCAPE_CHAR, k)] = v
			} else {
				validatedProperties[k] = v
			}
		}
	}
	return &validatedProperties
}
