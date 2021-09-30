package model

const (
	SalesforceDisplayCategory = "salesforce_metrics"
)

var KPIMetricsForSalesforce = []string{
	CountOfContactsCreated, CountOfContactsUpdated,
	CountOfLeadsCreated, CountOfLeadsUpdated,
	CountOfOpportunitiesCreated, CountOfOpportunitiesUpdated,
}

func ValidateKPISalesforce(kpiQuery KPIQuery) bool {
	return validateKPIQueryMetricsForSalesforce(kpiQuery.Metrics)
}
func validateKPIQueryMetricsForSalesforce(kpiQueryMetrics []string) bool {
	return ValidateKPIQueryMetricsForAnyEventType(kpiQueryMetrics, MapOfMetricsToData[SalesforceDisplayCategory])
}
