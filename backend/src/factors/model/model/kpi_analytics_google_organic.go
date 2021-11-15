package model

const (
	GoogleOrganicDisplayCategory = "google_organic_metrics"
)

func GetKPIConfigsForGoogleOrganic() map[string]interface{} {
	return map[string]interface{}{
		"category":         ChannelCategory,
		"display_category": GoogleOrganicDisplayCategory,
		"metrics":          SelectableMetricsForGoogleOrganic,
	}
}
