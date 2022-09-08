package model

const (
	AllChannelsDisplayCategory = "all_channels_metrics"
)

func GetKPIConfigsForAllChannels() map[string]interface{} {
	config := map[string]interface{}{
		"category":         ChannelCategory,
		"display_category": AllChannelsDisplayCategory,
	}
	config["metrics"] = GetStaticallyDefinedMetricsForDisplayCategory(AllChannelsDisplayCategory)
	return config
}

func GetKPIMetricsForAllChannels() []map[string]string {
	return GetStaticallyDefinedMetricsForDisplayCategory(AllChannelsDisplayCategory)
}
