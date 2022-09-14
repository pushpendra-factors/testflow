package model

const (
	GoogleOrganicDisplayCategory = "google_organic_metrics"
)

var KpiGoogleOrganicConfig = map[string]interface{}{
	"category":         ChannelCategory,
	"display_category": GoogleOrganicDisplayCategory,
}

func GetKPIMetricsForGoogleOrganic() []map[string]string {
	return GetStaticallyDefinedMetricsForDisplayCategory(GoogleOrganicDisplayCategory)
}
