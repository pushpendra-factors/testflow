package model

const (
	SalesforceUsersDisplayCategory         = "salesforce_users"
	SalesforceAccountsDisplayCategory      = "salesforce_accounts"
	SalesforceOpportunitiesDisplayCategory = "salesforce_opportunities"
)

var DisplayCategoriesForSalesforce = []string{SalesforceUsersDisplayCategory, SalesforceAccountsDisplayCategory, SalesforceOpportunitiesDisplayCategory}

func ValidateKPISalesforceUsers(kpiQuery KPIQuery) bool {
	return ValidateKPIQueryMetricsForAnyEventType(kpiQuery.Metrics, MapOfMetricsToData[SalesforceUsersDisplayCategory])
}

func ValidateKPISalesforceAccounts(kpiQuery KPIQuery) bool {
	return ValidateKPIQueryMetricsForAnyEventType(kpiQuery.Metrics, MapOfMetricsToData[SalesforceAccountsDisplayCategory])
}

func ValidateKPISalesforceOpportunities(kpiQuery KPIQuery) bool {
	return ValidateKPIQueryMetricsForAnyEventType(kpiQuery.Metrics, MapOfMetricsToData[SalesforceOpportunitiesDisplayCategory])
}
