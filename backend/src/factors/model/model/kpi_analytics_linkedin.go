package model

const (
	LinkedinDisplayCategory = "linkedin_metrics"
)

func GetKPIConfigsForLinkedin() map[string]interface{} {
	config := map[string]interface{}{
		"category":         ChannelCategory,
		"display_category": LinkedinDisplayCategory,
	}
	config["metrics"] = GetMetricsForDisplayCategory(AllChannelsDisplayCategory)
	return config
}
