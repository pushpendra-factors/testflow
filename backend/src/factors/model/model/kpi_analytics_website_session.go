package model

import (
	U "factors/util"
)

const (
	WebsiteSessionDisplayCategory = "website_session"
)

var KPIPropertiesForWebsiteSessions = []map[string]string{
	MapOfKPIPropertyNameToData[U.EP_SOURCE][EventEntity],
	MapOfKPIPropertyNameToData[U.EP_MEDIUM][EventEntity],
	MapOfKPIPropertyNameToData[U.EP_CAMPAIGN][EventEntity],
	MapOfKPIPropertyNameToData[U.EP_ADGROUP][EventEntity],
	MapOfKPIPropertyNameToData[U.EP_KEYWORD][EventEntity],
	MapOfKPIPropertyNameToData[U.EP_CONTENT][EventEntity],
	MapOfKPIPropertyNameToData[U.EP_CHANNEL][EventEntity],
	MapOfKPIPropertyNameToData[U.UP_INITIAL_REFERRER_URL][EventEntity],
	MapOfKPIPropertyNameToData[U.SP_INITIAL_PAGE_URL][EventEntity],
	MapOfKPIPropertyNameToData[U.SP_LATEST_PAGE_URL][EventEntity],
	MapOfKPIPropertyNameToData[U.UP_OS][EventEntity],
	MapOfKPIPropertyNameToData[U.UP_OS_VERSION][EventEntity],
	MapOfKPIPropertyNameToData[U.UP_BROWSER][EventEntity],
	MapOfKPIPropertyNameToData[U.UP_BROWSER_VERSION][EventEntity],
	MapOfKPIPropertyNameToData[U.UP_COUNTRY][EventEntity],
	MapOfKPIPropertyNameToData[U.UP_REGION][EventEntity],
	MapOfKPIPropertyNameToData[U.UP_CITY][EventEntity],
	MapOfKPIPropertyNameToData[U.EP_TIMESTAMP][EventEntity],
	MapOfKPIPropertyNameToData[U.SP_SPENT_TIME][EventEntity],
	MapOfKPIPropertyNameToData[U.SP_INITIAL_PAGE_LOAD_TIME][EventEntity],

	MapOfKPIPropertyNameToData[U.UP_DEVICE_TYPE][UserEntity],
	MapOfKPIPropertyNameToData[U.UP_DEVICE_BRAND][UserEntity],
	MapOfKPIPropertyNameToData[U.UP_DEVICE_MODEL][UserEntity],
	MapOfKPIPropertyNameToData[U.UP_DEVICE_NAME][UserEntity],
	MapOfKPIPropertyNameToData[U.UP_PLATFORM][UserEntity],
}

var KPIConfigForWebsiteSessions = map[string]interface{}{
	"category":         EventCategory,
	"display_category": WebsiteSessionDisplayCategory,
}

func ValidateKPIQueryMetricsForWebsiteSession(kpiQueryMetrics []string) bool {
	return ValidateKPIQueryMetricsForAnyEventType(kpiQueryMetrics, MapOfMetricsToData[WebsiteSessionDisplayCategory])
}
