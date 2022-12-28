package model

const (
	AdwordsDisplayCategory   = "adwords_metrics"
	GoogleAdsDisplayCategory = "google_ads_metrics"
)

var KpiAdwordsConfig = map[string]interface{}{
	"category":         ChannelCategory,
	"display_category": GoogleAdsDisplayCategory,
}

func GetKPIMetricsForAdwords() []map[string]string {
	allChannelMetrics := GetStaticallyDefinedMetricsForDisplayCategory(AllChannelsDisplayCategory)
	return append(allChannelMetrics, GetStaticallyDefinedMetricsForDisplayCategory(GoogleAdsDisplayCategory)...)
}

// TODO: Move to constants declared in model.
var MapOfCategoryToChannel = map[string]string{
	AllChannelsDisplayCategory:                "all_ads",
	AdwordsDisplayCategory:                    "google_ads",
	BingAdsDisplayCategory:                    "bing_ads",
	GoogleAdsDisplayCategory:                  "google_ads",
	FacebookDisplayCategory:                   "facebook_ads",
	LinkedinDisplayCategory:                   "linkedin_ads",
	GoogleOrganicDisplayCategory:              "search_console",
	LinkedinCompanyEngagementsDisplayCategory: "linkedin_company_engagements",
}
