package model

const (
	FacebookDisplayCategory = "facebook_metrics"
)

var KpiFacebookConfig = map[string]interface{}{
	"category":         ChannelCategory,
	"display_category": FacebookDisplayCategory,
}

func GetKPIMetricsForFacebook() []map[string]string {
	allChannelMetrics := GetStaticallyDefinedMetricsForDisplayCategory(AllChannelsDisplayCategory)
	return append(allChannelMetrics, GetStaticallyDefinedMetricsForDisplayCategory(FacebookDisplayCategory)...)
}
