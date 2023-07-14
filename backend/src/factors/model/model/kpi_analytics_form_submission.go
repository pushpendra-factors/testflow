package model

import U "factors/util"

const (
	FormSubmissionsDisplayCategory = "form_submission"
)

var KPIPropertiesForFormSubmissions = []map[string]string{
	MapOfKPIPropertyNameToDataWithCategory(U.EP_REFERRER_URL, EventEntity),
	MapOfKPIPropertyNameToDataWithCategory(U.EP_PAGE_URL, EventEntity),
	MapOfKPIPropertyNameToDataWithCategory(U.EP_TIMESTAMP, EventEntity),

	MapOfKPIPropertyNameToDataWithCategory(U.UP_OS, EventEntity),
	MapOfKPIPropertyNameToDataWithCategory(U.UP_OS_VERSION, EventEntity),
	MapOfKPIPropertyNameToDataWithCategory(U.UP_BROWSER, EventEntity),
	MapOfKPIPropertyNameToDataWithCategory(U.UP_BROWSER_VERSION, EventEntity),
	MapOfKPIPropertyNameToDataWithCategory(U.UP_COUNTRY, EventEntity),
	MapOfKPIPropertyNameToDataWithCategory(U.UP_REGION, EventEntity),
	MapOfKPIPropertyNameToDataWithCategory(U.UP_CITY, EventEntity),
}

var KPIConfigForFormSubmissions = map[string]interface{}{
	"category":         EventCategory,
	"display_category": FormSubmissionsDisplayCategory,
}

func ValidateKPIFormSubmissions(kpiQuery KPIQuery) bool {
	return validateKPIQueryMetricsForFormSubmission(kpiQuery.Metrics) ||
		validateKPIQueryFiltersForFormSubmission(kpiQuery.Filters) ||
		validateKPIQueryGroupByForFormSubmission(kpiQuery.GroupBy)
}
func validateKPIQueryMetricsForFormSubmission(kpiQueryMetrics []string) bool {
	return ValidateKPIQueryMetricsForAnyEventType(kpiQueryMetrics, MapOfMetricsToData[FormSubmissionsDisplayCategory])
}

func validateKPIQueryFiltersForFormSubmission(kpiQueryFilters []KPIFilter) bool {
	return ValidateKPIQueryFiltersForAnyEventType(kpiQueryFilters, KPIPropertiesForFormSubmissions)
}

func validateKPIQueryGroupByForFormSubmission(kpiQueryGroupBys []KPIGroupBy) bool {
	return ValidateKPIQueryGroupByForAnyEventType(kpiQueryGroupBys, KPIPropertiesForFormSubmissions)
}
