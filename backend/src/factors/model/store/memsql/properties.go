package memsql

import (
	"factors/config"
	U "factors/util"
	"net/http"
	"strings"
)

func (store *MemSQL) GetStandardUserPropertiesBasedOnIntegration(projectID int64) map[string]string {

	finalStandardUserProperties := make(map[string]string)
	for property, propertyDisplayName := range U.STANDARD_USER_PROPERTIES_DISPLAY_NAMES {
		if !strings.HasPrefix(property, U.ENRICHED_PROPERTIES_PREFIX) && !strings.HasPrefix(property, U.SIX_SIGNAL_PROPERTIES_PREFIX) && !strings.HasPrefix(property, U.G2_PROPERTIES_PREFIX) {
			finalStandardUserProperties[property] = propertyDisplayName
		}
	}

	isClearBitIntegrated, statusCode := store.IsClearbitIntegratedByProjectID(projectID)

	if statusCode == http.StatusFound && isClearBitIntegrated {
		for property, propertyDisplayName := range U.STANDARD_USER_PROPERTIES_DISPLAY_NAMES {
			if config.IsCompanyPropsV1Enabled(projectID) {
				if strings.HasPrefix(property, U.ENRICHED_PROPERTIES_PREFIX) || strings.HasPrefix(property, U.SIX_SIGNAL_PROPERTIES_PREFIX) {
					finalStandardUserProperties[property] = propertyDisplayName
				}
			} else {
				if strings.HasPrefix(property, U.CLR_PROPERTIES_PREFIX) {
					finalStandardUserProperties[property] = propertyDisplayName
				}
			}

		}
	}

	isSixSignalIntegrated, statusCode2 := store.IsSixSignalIntegratedByEitherWay(projectID)

	if statusCode2 == http.StatusFound && isSixSignalIntegrated {
		for property, propertyDisplayName := range U.STANDARD_USER_PROPERTIES_DISPLAY_NAMES {
			if strings.HasPrefix(property, U.SIX_SIGNAL_PROPERTIES_PREFIX) {
				if _, exists := finalStandardUserProperties[property]; !exists {
					finalStandardUserProperties[property] = propertyDisplayName
				}
			}
		}
	}

	return finalStandardUserProperties
}
