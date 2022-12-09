package memsql

import (
	U "factors/util"
	"net/http"
	"strings"
)

func (store *MemSQL) GetStandardUserPropertiesBasedOnIntegration(projectID int64) map[string]string {

	finalStandardUserProperties := make(map[string]string)
	propertiesAfterClearBitCheck := make(map[string]string)
	propertiesAfterSixSignalCheck := make(map[string]string)
	clearBitKey, statusCode := store.GetClearbitKeyFromProjectSetting(projectID)
	for property, propertyDisplayName := range U.STANDARD_USER_PROPERTIES_DISPLAY_NAMES {
		if strings.HasPrefix(property, U.CLR_PROPERTIES_PREFIX) && (statusCode != http.StatusFound || clearBitKey == "") {
			continue
		} else {
			propertiesAfterClearBitCheck[property] = propertyDisplayName
		}
	}

	sixSignalKey, statusCode2 := store.GetClient6SignalKeyFromProjectSetting(projectID)
	for property, propertyDisplayName := range propertiesAfterClearBitCheck {
		if strings.HasPrefix(property, U.SIX_SIGNAL_PROPERTIES_PREFIX) && (statusCode2 != http.StatusFound || sixSignalKey == "") {
			continue
		} else {
			propertiesAfterSixSignalCheck[property] = propertyDisplayName
		}
	}

	finalStandardUserProperties = propertiesAfterClearBitCheck
	return finalStandardUserProperties
}
