package model

import (
	U "factors/util"
)

const (
	WebsiteSessionDisplayCategory = "website_session"
)

var KPIPropertiesForWebsiteSessions = []map[string]string{
	MapOfKPIPropertyNameToDataWithCategory(U.EP_SOURCE,EventEntity),
	MapOfKPIPropertyNameToDataWithCategory(U.EP_MEDIUM,EventEntity),
	MapOfKPIPropertyNameToDataWithCategory(U.EP_CAMPAIGN,EventEntity),
	MapOfKPIPropertyNameToDataWithCategory(U.EP_ADGROUP,EventEntity),
	MapOfKPIPropertyNameToDataWithCategory(U.EP_KEYWORD,EventEntity),
	MapOfKPIPropertyNameToDataWithCategory(U.EP_CONTENT,EventEntity),
	MapOfKPIPropertyNameToDataWithCategory(U.EP_CHANNEL,EventEntity),
	MapOfKPIPropertyNameToDataWithCategory(U.UP_INITIAL_REFERRER_URL,EventEntity),
	MapOfKPIPropertyNameToDataWithCategory(U.SP_INITIAL_PAGE_URL,EventEntity),
	MapOfKPIPropertyNameToDataWithCategory(U.SP_LATEST_PAGE_URL,EventEntity),
	MapOfKPIPropertyNameToDataWithCategory(U.UP_OS,EventEntity),
	MapOfKPIPropertyNameToDataWithCategory(U.UP_OS_VERSION,EventEntity),
	MapOfKPIPropertyNameToDataWithCategory(U.UP_BROWSER,EventEntity),
	MapOfKPIPropertyNameToDataWithCategory(U.UP_BROWSER_VERSION,EventEntity),
	MapOfKPIPropertyNameToDataWithCategory(U.UP_COUNTRY,EventEntity),
	MapOfKPIPropertyNameToDataWithCategory(U.UP_REGION,EventEntity),
	MapOfKPIPropertyNameToDataWithCategory(U.UP_CITY,EventEntity),
	MapOfKPIPropertyNameToDataWithCategory(U.EP_TIMESTAMP,EventEntity),
	MapOfKPIPropertyNameToDataWithCategory(U.SP_SPENT_TIME,EventEntity),
	MapOfKPIPropertyNameToDataWithCategory(U.SP_INITIAL_PAGE_LOAD_TIME,EventEntity),
	MapOfKPIPropertyNameToDataWithCategory(U.SP_INITIAL_PAGE_SCROLL_PERCENT,EventEntity),
}

var KPIConfigForWebsiteSessions = map[string]interface{}{
	"category":         EventCategory,
	"display_category": WebsiteSessionDisplayCategory,
}

func ValidateKPIQueryMetricsForWebsiteSession(kpiQueryMetrics []string) bool {
	return ValidateKPIQueryMetricsForAnyEventType(kpiQueryMetrics, MapOfMetricsToData[WebsiteSessionDisplayCategory])
}
