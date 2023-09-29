package model

import U "factors/util"

const (
	PageViewsDisplayCategory = "page_views"
)

var KPIPropertiesForPageViews = []map[string]string{
	MapOfKPIPropertyNameToDataWithCategory(U.EP_REFERRER_URL, EventEntity, false),
	MapOfKPIPropertyNameToDataWithCategory(U.EP_PAGE_URL, EventEntity, false),
	MapOfKPIPropertyNameToDataWithCategory(U.EP_TIMESTAMP, EventEntity, false),
	MapOfKPIPropertyNameToDataWithCategory(U.EP_PAGE_LOAD_TIME, EventEntity, false),
	MapOfKPIPropertyNameToDataWithCategory(U.EP_PAGE_SPENT_TIME, EventEntity, false),
	MapOfKPIPropertyNameToDataWithCategory(U.EP_PAGE_SCROLL_PERCENT, EventEntity, false),

	MapOfKPIPropertyNameToDataWithCategory(U.UP_OS, EventEntity, false),
	MapOfKPIPropertyNameToDataWithCategory(U.UP_OS_VERSION, EventEntity, false),
	MapOfKPIPropertyNameToDataWithCategory(U.UP_BROWSER, EventEntity, false),
	MapOfKPIPropertyNameToDataWithCategory(U.UP_BROWSER_VERSION, EventEntity, false),
	MapOfKPIPropertyNameToDataWithCategory(U.UP_COUNTRY, EventEntity, false),
	MapOfKPIPropertyNameToDataWithCategory(U.UP_REGION, EventEntity, false),
	MapOfKPIPropertyNameToDataWithCategory(U.UP_CITY, EventEntity, false),
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
