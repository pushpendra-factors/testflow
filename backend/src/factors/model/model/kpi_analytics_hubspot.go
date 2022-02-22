package model

const (
	HubspotContactsDisplayCategory  = "hubspot_contacts"
	HubspotCompaniesDisplayCategory = "hubspot_companies"
	HubspotDealsDisplayCategory     = "hubspot_deals"
)

var DisplayCategoriesForHubspot = []string{HubspotContactsDisplayCategory, HubspotCompaniesDisplayCategory}

func ValidateKPIHubspotContacts(kpiQuery KPIQuery) bool {
	return ValidateKPIQueryMetricsForAnyEventType(kpiQuery.Metrics, MapOfMetricsToData[HubspotContactsDisplayCategory])
}

func ValidateKPIHubspotCompanies(kpiQuery KPIQuery) bool {
	return ValidateKPIQueryMetricsForAnyEventType(kpiQuery.Metrics, MapOfMetricsToData[HubspotCompaniesDisplayCategory])
}

func ValidateKPIHubspotDeals(kpiQuery KPIQuery) bool {
	return ValidateKPIQueryMetricsForAnyEventType(kpiQuery.Metrics, MapOfMetricsToData[HubspotDealsDisplayCategory])
}
