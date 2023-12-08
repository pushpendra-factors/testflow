package model

import (
	U "factors/util"
)

const (
	WebsiteSessionDisplayCategory = "website_session"
)

var KPIPropertiesForWebsiteSessions = []map[string]string{
	MapOfKPIPropertyNameToDataWithCategory(U.EP_SOURCE, EventEntity, true),
	MapOfKPIPropertyNameToDataWithCategory(U.EP_MEDIUM, EventEntity, true),
	MapOfKPIPropertyNameToDataWithCategory(U.EP_CAMPAIGN, EventEntity, true),
	MapOfKPIPropertyNameToDataWithCategory(U.EP_ADGROUP, EventEntity, true),
	MapOfKPIPropertyNameToDataWithCategory(U.EP_KEYWORD, EventEntity, true),
	MapOfKPIPropertyNameToDataWithCategory(U.EP_CONTENT, EventEntity, true),
	MapOfKPIPropertyNameToDataWithCategory(U.EP_CHANNEL, EventEntity, true),
	MapOfKPIPropertyNameToDataWithCategory(U.UP_INITIAL_REFERRER_URL, EventEntity, true),
	MapOfKPIPropertyNameToDataWithCategory(U.SP_INITIAL_PAGE_URL, EventEntity, true),
	MapOfKPIPropertyNameToDataWithCategory(U.SP_LATEST_PAGE_URL, EventEntity, true),
	MapOfKPIPropertyNameToDataWithCategory(U.UP_OS, EventEntity, true),
	MapOfKPIPropertyNameToDataWithCategory(U.UP_OS_VERSION, EventEntity, true),
	MapOfKPIPropertyNameToDataWithCategory(U.UP_BROWSER, EventEntity, true),
	MapOfKPIPropertyNameToDataWithCategory(U.UP_BROWSER_VERSION, EventEntity, true),
	MapOfKPIPropertyNameToDataWithCategory(U.UP_COUNTRY, EventEntity, true),
	MapOfKPIPropertyNameToDataWithCategory(U.UP_CITY, EventEntity, true),
	MapOfKPIPropertyNameToDataWithCategory(U.EP_TIMESTAMP, EventEntity, true),
	MapOfKPIPropertyNameToDataWithCategory(U.SP_SPENT_TIME, EventEntity, true),
	MapOfKPIPropertyNameToDataWithCategory(U.SP_INITIAL_PAGE_LOAD_TIME, EventEntity, true),
	MapOfKPIPropertyNameToDataWithCategory(U.SP_INITIAL_PAGE_SCROLL_PERCENT, EventEntity, true),
}

var KPIConfigForWebsiteSessions = map[string]interface{}{
	"category":         EventCategory,
	"display_category": WebsiteSessionDisplayCategory,
}

func ValidateKPIQueryMetricsForWebsiteSession(kpiQueryMetrics []string) bool {
	return ValidateKPIQueryMetricsForAnyEventType(kpiQueryMetrics, MapOfMetricsToData[WebsiteSessionDisplayCategory])
}
