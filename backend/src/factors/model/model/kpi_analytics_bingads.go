package model

const (
	BingAdsDisplayCategory = "bingads_metrics"
)

var KpiBingAdsConfig = map[string]interface{}{
	"category":         ChannelCategory,
	"display_category": BingAdsDisplayCategory,
}

func GetKPIMetricsForBingAds() []map[string]string {
	allChannelMetrics := GetStaticallyDefinedMetricsForDisplayCategory(AllChannelsDisplayCategory)
	return append(allChannelMetrics, GetStaticallyDefinedMetricsForDisplayCategory(BingAdsDisplayCategory)...)
}
