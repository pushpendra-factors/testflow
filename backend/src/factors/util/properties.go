package util

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

// Common properties type.
type PropertiesMap map[string]interface{}

/* Properties Constants */

// Event Properties.
var EP_OCCURRENCE_COUNT string = "$occurrenceCount"

// Default User Properties - Added from JS SDK
var UP_REFERRER string = "$referrer"
var UP_BROWSER string = "$browser"
var UP_BROWSER_VERSION string = "$browserVersion"
var UP_OS string = "$os"
var UP_OS_VERSION string = "$osVersion"
var UP_SCREEN_WIDTH string = "$screenWidth"
var UP_SCREEN_HEIGHT string = "$screenHeight"

// Default User Properties - Added from Backend
var UP_COUNTRY string = "$country"

// Default User Property - Added from Backend
var UP_INTERNAL_IP = "$ip"

var DEFAULT_NUMERIC_EVENT_PROPERTIES = [...]string{EP_OCCURRENCE_COUNT}

// IsValidProperty - Validates property key and value.
func IsValidProperty(key string, value interface{}) error {
	// Value - type validation.
	switch valueType := value.(type) {
	case float64:
	case string:
	default:
		log.WithFields(log.Fields{"key": key, "value": value,
			"valueType": valueType}).Debug("Invalid type used on property")
		return fmt.Errorf("invalid property type")
	}
	return nil
}

// FilterValidProperties - Returns only valid properties.
func FilterValidProperties(properties *PropertiesMap) *PropertiesMap {
	validProperties := make(PropertiesMap)
	for k, v := range *properties {
		if err := IsValidProperty(k, v); err == nil {
			validProperties[k] = v
		}
	}
	return &validProperties
}
