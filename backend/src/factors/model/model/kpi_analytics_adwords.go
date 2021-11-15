package model

const (
	AdwordsDisplayCategory = "adwords_metrics"
)

func GetKPIConfigsForAdwords() map[string]interface{} {
	config := map[string]interface{}{
		"category":         ChannelCategory,
		"display_category": AdwordsDisplayCategory,
	}
	allChannelMetrics := GetMetricsForDisplayCategory(AllChannelsDisplayCategory)
	config["metrics"] = append(allChannelMetrics, GetMetricsForDisplayCategory(AdwordsDisplayCategory)...)
	return config
}

// TODO: Move to constants declared in model.
var MapOfCategoryToChannel = map[string]string{
	AllChannelsDisplayCategory:   "all_ads",
	AdwordsDisplayCategory:       "google_ads",
	FacebookDisplayCategory:      "facebook_ads",
	LinkedinDisplayCategory:      "linkedin_ads",
	GoogleOrganicDisplayCategory: "search_console",
}
