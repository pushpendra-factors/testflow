package model

import U "factors/util"

const (
	PageViewsDisplayCategory = "page_views"
)

var KPIPropertiesForPageViews = []map[string]string{
	MapOfKPIPropertyNameToDataWithCategory(U.EP_REFERRER_URL,EventEntity),
	MapOfKPIPropertyNameToDataWithCategory(U.EP_PAGE_URL,EventEntity),
	MapOfKPIPropertyNameToDataWithCategory(U.EP_TIMESTAMP,EventEntity),
	MapOfKPIPropertyNameToDataWithCategory(U.EP_PAGE_LOAD_TIME,EventEntity),
	MapOfKPIPropertyNameToDataWithCategory(U.EP_PAGE_SPENT_TIME,EventEntity),
	MapOfKPIPropertyNameToDataWithCategory(U.EP_PAGE_SCROLL_PERCENT,EventEntity),

	MapOfKPIPropertyNameToDataWithCategory(U.UP_OS,EventEntity),
	MapOfKPIPropertyNameToDataWithCategory(U.UP_OS_VERSION,EventEntity),
	MapOfKPIPropertyNameToDataWithCategory(U.UP_BROWSER,EventEntity),
	MapOfKPIPropertyNameToDataWithCategory(U.UP_BROWSER_VERSION,EventEntity),
	MapOfKPIPropertyNameToDataWithCategory(U.UP_COUNTRY,EventEntity),
	MapOfKPIPropertyNameToDataWithCategory(U.UP_REGION,EventEntity),
	MapOfKPIPropertyNameToDataWithCategory(U.UP_CITY,EventEntity),
}

var KPIConfigForPageViews = map[string]interface{}{
	"category":         EventCategory,
	"display_category": PageViewsDisplayCategory,
}

func ValidateKPIPageView(kpiQuery KPIQuery) bool {
	return validateKPIQueryMetricsForPageView(kpiQuery.Metrics) ||
		validateKPIQueryFiltersForPageView(kpiQuery.Filters) ||
		validateKPIQueryGroupByForPageView(kpiQuery.GroupBy)
}
func validateKPIQueryMetricsForPageView(kpiQueryMetrics []string) bool {
	return ValidateKPIQueryMetricsForAnyEventType(kpiQueryMetrics, MapOfMetricsToData[PageViewsDisplayCategory])
}

func validateKPIQueryFiltersForPageView(kpiQueryFilters []KPIFilter) bool {
	return ValidateKPIQueryFiltersForAnyEventType(kpiQueryFilters, KPIPropertiesForPageViews)
}

func validateKPIQueryGroupByForPageView(kpiQueryGroupBys []KPIGroupBy) bool {
	return ValidateKPIQueryGroupByForAnyEventType(kpiQueryGroupBys, KPIPropertiesForPageViews)
}
