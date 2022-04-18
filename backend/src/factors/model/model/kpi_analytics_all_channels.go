package model

const (
	AllChannelsDisplayCategory = "all_channels_metrics"
)

func GetKPIConfigsForAllChannels() map[string]interface{} {
	config := map[string]interface{}{
		"category":         ChannelCategory,
		"display_category": AllChannelsDisplayCategory,
	}
	config["metrics"] = GetMetricsForDisplayCategory(AllChannelsDisplayCategory)
	return config
}
