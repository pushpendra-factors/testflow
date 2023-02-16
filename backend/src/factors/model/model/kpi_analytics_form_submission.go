package model

import U "factors/util"

const (
	FormSubmissionsDisplayCategory = "form_submission"
)

var KPIPropertiesForFormSubmissions = []map[string]string{
	MapOfKPIPropertyNameToData[U.EP_REFERRER_URL][EventEntity],
	MapOfKPIPropertyNameToData[U.EP_PAGE_URL][EventEntity],
	MapOfKPIPropertyNameToData[U.EP_TIMESTAMP][EventEntity],

	MapOfKPIPropertyNameToData[U.UP_OS][EventEntity],
	MapOfKPIPropertyNameToData[U.UP_OS_VERSION][EventEntity],
	MapOfKPIPropertyNameToData[U.UP_BROWSER][EventEntity],
	MapOfKPIPropertyNameToData[U.UP_BROWSER_VERSION][EventEntity],
	MapOfKPIPropertyNameToData[U.UP_COUNTRY][EventEntity],
	MapOfKPIPropertyNameToData[U.UP_REGION][EventEntity],
	MapOfKPIPropertyNameToData[U.UP_CITY][EventEntity],
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
