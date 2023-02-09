package model

import U "factors/util"

const (
	PageViewsDisplayCategory = "page_views"
)

var KPIPropertiesForPageViews = []map[string]string{
	MapOfKPIPropertyNameToData[U.EP_REFERRER_URL][EventEntity],
	MapOfKPIPropertyNameToData[U.EP_PAGE_URL][EventEntity],
	MapOfKPIPropertyNameToData[U.EP_TIMESTAMP][EventEntity],
	MapOfKPIPropertyNameToData[U.EP_PAGE_LOAD_TIME][EventEntity],
	MapOfKPIPropertyNameToData[U.EP_PAGE_SPENT_TIME][EventEntity],
	MapOfKPIPropertyNameToData[U.EP_PAGE_SCROLL_PERCENT][EventEntity],

	MapOfKPIPropertyNameToData[U.UP_OS][EventEntity],
	MapOfKPIPropertyNameToData[U.UP_OS_VERSION][EventEntity],
	MapOfKPIPropertyNameToData[U.UP_BROWSER][EventEntity],
	MapOfKPIPropertyNameToData[U.UP_BROWSER_VERSION][EventEntity],
	MapOfKPIPropertyNameToData[U.UP_COUNTRY][EventEntity],
	MapOfKPIPropertyNameToData[U.UP_REGION][EventEntity],
	MapOfKPIPropertyNameToData[U.UP_CITY][EventEntity],
	MapOfKPIPropertyNameToData[U.UP_DEVICE_TYPE][UserEntity],
	MapOfKPIPropertyNameToData[U.UP_DEVICE_BRAND][UserEntity],
	MapOfKPIPropertyNameToData[U.UP_DEVICE_MODEL][UserEntity],
	MapOfKPIPropertyNameToData[U.UP_DEVICE_NAME][UserEntity],
	MapOfKPIPropertyNameToData[U.UP_PLATFORM][UserEntity],
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
