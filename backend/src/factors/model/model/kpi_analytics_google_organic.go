package model

const (
	GoogleOrganicDisplayCategory = "google_organic_metrics"
)

func GetKPIConfigsForGoogleOrganic() map[string]interface{} {
	config := map[string]interface{}{
		"category":         ChannelCategory,
		"display_category": GoogleOrganicDisplayCategory,
	}
	config["metrics"] = GetMetricsForDisplayCategory(GoogleOrganicDisplayCategory)
	return config
}
