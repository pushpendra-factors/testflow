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

// Default User Properties - Added from JS SDK
var UP_PLATFORM string = "$platform"
var UP_REFERRER string = "$referrer"
var UP_BROWSER string = "$browser"
var UP_BROWSER_VERSION string = "$browserVersion"
var UP_OS string = "$os"
var UP_OS_VERSION string = "$osVersion"
var UP_SCREEN_WIDTH string = "$screenWidth"
var UP_SCREEN_HEIGHT string = "$screenHeight"

var DEFAULT_USER_PROPERTIES_SDK = [...]string{
	UP_PLATFORM,
	UP_REFERRER,
	UP_BROWSER,
	UP_BROWSER_VERSION,
	UP_OS,
	UP_OS_VERSION,
	UP_SCREEN_WIDTH,
	UP_SCREEN_HEIGHT,
}

// Default User Properties - Added from Backend
var UP_COUNTRY string = "$country"
var UP_CITY string = "$city"

// Default User Property - Added from Backend
var UP_INTERNAL_IP = "$ip"

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
	for _, k := range DEFAULT_USER_PROPERTIES_SDK {
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
			if strings.HasPrefix(k, NAME_PREFIX) && !strings.HasPrefix(k, QUERY_PARAM_PROPERTY_PREFIX) {
				validatedProperties[fmt.Sprintf("%s%s", NAME_PREFIX_ESCAPE_CHAR, k)] = v
			} else {
				validatedProperties[k] = v
			}
		}
	}
	return &validatedProperties
}
