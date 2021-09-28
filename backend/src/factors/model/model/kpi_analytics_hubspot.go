package model

const (
	HubspotDisplayCategory = "hubspot_metrics"
)

func ValidateKPIHubspot(kpiQuery KPIQuery) bool {
	return validateKPIQueryMetricsForHubspot(kpiQuery.Metrics)
}

func validateKPIQueryMetricsForHubspot(kpiQueryMetrics []string) bool {
	return ValidateKPIQueryMetricsForAnyEventType(kpiQueryMetrics, MapOfMetricsToData[HubspotDisplayCategory])
}
