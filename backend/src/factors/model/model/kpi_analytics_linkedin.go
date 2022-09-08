package model

const (
	LinkedinDisplayCategory = "linkedin_metrics"
)

var KpiLinkedinConfig = map[string]interface{}{
	"category":         ChannelCategory,
	"display_category": LinkedinDisplayCategory,
}

func GetKPIMetricsForLinkedin() []map[string]string {
	return GetStaticallyDefinedMetricsForDisplayCategory(AllChannelsDisplayCategory)
}
