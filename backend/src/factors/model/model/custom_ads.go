package model

import U "factors/util"

var ObjectsForCustomAds = []string{FilterCampaign, FilterAdGroup, FilterKeyword}

var CustomAdsIntegration = "custom_ads"

var MapOfCustomAdsObjectsToPropertiesAndRelated = map[string]map[string]PropertiesAndRelated{
	FilterCampaign: {
		"name":   PropertiesAndRelated{TypeOfProperty: U.PropertyTypeCategorical},
		"id":     PropertiesAndRelated{TypeOfProperty: U.PropertyTypeCategorical},
		"status": PropertiesAndRelated{TypeOfProperty: U.PropertyTypeCategorical},
		"type":   PropertiesAndRelated{TypeOfProperty: U.PropertyTypeCategorical},
	},
	FilterAdGroup: {
		"name":   PropertiesAndRelated{TypeOfProperty: U.PropertyTypeCategorical},
		"id":     PropertiesAndRelated{TypeOfProperty: U.PropertyTypeCategorical},
		"status": PropertiesAndRelated{TypeOfProperty: U.PropertyTypeCategorical},
		"type":   PropertiesAndRelated{TypeOfProperty: U.PropertyTypeCategorical},
	},
	FilterKeyword: {
		"name":       PropertiesAndRelated{TypeOfProperty: U.PropertyTypeCategorical},
		"id":         PropertiesAndRelated{TypeOfProperty: U.PropertyTypeCategorical},
		"status":     PropertiesAndRelated{TypeOfProperty: U.PropertyTypeCategorical},
		"match_type": PropertiesAndRelated{TypeOfProperty: U.PropertyTypeCategorical},
	},
}

const (
	CustomAdsSpecificError = "Failed in custom ads with the error."
)

var CustomadsDocumentTypeAlias = map[string]int{
	"campaigns":               1,
	"ad_groups":               2,
	"keyword":                 3,
	CampaignPerformanceReport: 4,
	AdGroupPerformanceReport:  5,
	KeywordPerformanceReport:  6,
	"account":                 7,
}

var CustomAdsObjectInternalRepresentationToExternalRepresentation = map[string]string{
	FilterCampaign: "campaigns",
	FilterAdGroup:  "ad_groups",
	FilterKeyword:  "keyword",
	"channel":      "channel",
}

var CustomAdsInternalRepresentationToExternalRepresentation = map[string]string{
	"campaigns.id":       "id",
	"campaigns.status":   "status",
	"campaigns.name":     "name",
	"campaigns.type":     "type",
	"ad_groups.id":       "id",
	"ad_groups.status":   "status",
	"ad_groups.name":     "name",
	"ad_groups.type":     "type",
	"keyword.id":         "id",
	"keyword.name":       "name",
	"keyword.status":     "status",
	"keyword.match_type": "match_type",
	"impressions":        "impressions",
	"clicks":             "clicks",
	"spend":              "spend",
	"channel.name":       "channel_name",
}

var CustomAdsInternalRepresentationToExternalRepresentationForReports = map[string]string{
	"campaigns.id":       "campaign_id",
	"campaigns.status":   "campaign_status",
	"campaigns.name":     "campaign_name",
	"campaigns.type":     "campaign_type",
	"ad_groups.id":       "ad_group_id",
	"ad_groups.status":   "ad_group_status",
	"ad_groups.name":     "ad_group_name",
	"ad_groups.type":     "ad_group_type",
	"keyword.id":         "keyword_id",
	"keyword.name":       "keyword_name",
	"keyword.status":     "keyword_status",
	"keyword.match_type": "keyword_match_type",
	"impressions":        "impressions",
	"clicks":             "clicks",
	"spend":              "spend",
	"channel.name":       "channel_name",
}

var CustomAdsObjectToPerfomanceReportRepresentation = map[string]string{
	"campaigns": CampaignPerformanceReport,
	"ad_groups": AdGroupPerformanceReport,
	"keyword":   KeywordPerformanceReport,
}

var CustomAdsObjectMapForSmartProperty = map[string]string{
	"campaigns": FilterCampaign,
	"ad_groups": FilterAdGroup,
}
