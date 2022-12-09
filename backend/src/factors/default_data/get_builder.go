package default_data

const (
	SalesforceIntegrationName  = "salesforce"
	HubspotIntegrationName     = "hubspot"
	MarketoIntegrationName     = "marketo"
	LeadSquaredIntegrationName = "leadsquared"
)

var DefaultDataIntegrations = []string{SalesforceIntegrationName, HubspotIntegrationName, MarketoIntegrationName, LeadSquaredIntegrationName}

type BuildDefaultForCustomKPI interface {
	Build(projectID int64) int
}

// if we require a different kind of Query object to be fetched, a new factory can help.
func GetDefaultDataCustomKPIFactory(integration string) BuildDefaultForCustomKPI {

	if integration == SalesforceIntegrationName {
		return BuildDefaultSalesforceCustomKPI{}
	} else if integration == HubspotIntegrationName {
		return BuildDefaultHubspotCustomKPI{}
	} else if integration == MarketoIntegrationName {
		return BuildDefaultMarketoCustomKPI{}
		// } else if integration == leadSquaredIntegrationName {
	} else {
		return BuildDefaultLeadSquaredCustomKPI{}
	}
}
