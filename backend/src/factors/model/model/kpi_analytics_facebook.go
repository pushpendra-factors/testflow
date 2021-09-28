package model

const (
	FacebookDisplayCategory = "facebook_metrics"
)

func GetKPIConfigsForFacebook() map[string]interface{} {
	config := map[string]interface{}{
		"category":         ChannelCategory,
		"display_category": FacebookDisplayCategory,
		"metrics":          SelectableMetricsForFacebook,
		"properties":       tranformChannelConfigStructToKPISpecificConfig(MapOfFacebookObjectsToPropertiesAndRelated),
	}
	allChannelMetrics := GetMetricsForDisplayCategory(AllChannelsDisplayCategory)
	config["metrics"] = append(allChannelMetrics, GetMetricsForDisplayCategory(FacebookDisplayCategory)...)
	return config
}
