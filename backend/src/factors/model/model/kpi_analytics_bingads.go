package model

const (
	BingAdsDisplayCategory = "bingads_metrics"
)

func GetKPIConfigsForBingAds() map[string]interface{} {
	config := map[string]interface{}{
		"category":         ChannelCategory,
		"display_category": BingAdsDisplayCategory,
	}
	allChannelMetrics := GetMetricsForDisplayCategory(AllChannelsDisplayCategory)
	config["metrics"] = append(allChannelMetrics, GetMetricsForDisplayCategory(BingAdsDisplayCategory)...)
	return config
}
