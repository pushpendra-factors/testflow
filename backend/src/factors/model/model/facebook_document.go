package model

import (
	U "factors/util"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
)

type FacebookDocument struct {
	ProjectID           uint64          `gorm:"primary_key:true;auto_increment:false" json:"project_id"`
	CustomerAdAccountID string          `gorm:"primary_key:true;auto_increment:false" json:"customer_ad_account_id"`
	Platform            string          `gorm:"primary_key:true;auto_increment:false" json:"platform"`
	TypeAlias           string          `gorm:"-" json:"type_alias"`
	Type                int             `gorm:"primary_key:true;auto_increment:false" json:"type"`
	Timestamp           int64           `gorm:"primary_key:true;auto_increment:false" json:"timestamp"`
	ID                  string          `gorm:"primary_key:true;auto_increment:false" json:"id"`
	CampaignID          string          `json:"-"`
	AdSetID             string          `json:"-"`
	AdID                string          `json:"-"`
	Value               *postgres.Jsonb `json:"value"`
	CreatedAt           time.Time       `json:"created_at"`
	UpdatedAt           time.Time       `json:"updated_at"`
}

// FacebookLastSyncInfo ...
type FacebookLastSyncInfo struct {
	ProjectID           uint64 `json:"project_id"`
	CustomerAdAccountID string `json:"customer_acc_id"`
	DocumentType        int    `json:"-"`
	DocumentTypeAlias   string `json:"type_alias"`
	LastTimestamp       int64  `json:"last_timestamp"`
}

// FacebookLastSyncInfoPayload ...
type FacebookLastSyncInfoPayload struct {
	ProjectId           uint64 `json:"project_id"`
	CustomerAdAccountId string `json:"account_id"`
}

// NOTE: Change KPI metrics in kpi_analytics_common when changed.
var SelectableMetricsForFacebook = []string{
	"video_p50_watched_actions",
	"video_p25_watched_actions",
	"video_30_sec_watched_actions",
	"video_p100_watched_actions",
	"video_p75_watched_actions",
	"cost_per_click",
	"cost_per_link_click",
	"cost_per_thousand_impressions",
	"click_through_rate",
	"link_click_through_rate",
	"link_clicks",
	"frequency",
	"reach",
}

var MapOfFacebookObjectsToPropertiesAndRelated = map[string]map[string]PropertiesAndRelated{
	CAFilterCampaign: {
		"id":                PropertiesAndRelated{TypeOfProperty: U.PropertyTypeCategorical},
		"name":              PropertiesAndRelated{TypeOfProperty: U.PropertyTypeCategorical},
		"daily_budget":      PropertiesAndRelated{TypeOfProperty: U.PropertyTypeCategorical},
		"lifetime_budget":   PropertiesAndRelated{TypeOfProperty: U.PropertyTypeCategorical},
		"configured_status": PropertiesAndRelated{TypeOfProperty: U.PropertyTypeCategorical},
		"effective_status":  PropertiesAndRelated{TypeOfProperty: U.PropertyTypeCategorical},
		"objective":         PropertiesAndRelated{TypeOfProperty: U.PropertyTypeCategorical},
		"bid_strategy":      PropertiesAndRelated{TypeOfProperty: U.PropertyTypeCategorical},
		"buying_type":       PropertiesAndRelated{TypeOfProperty: U.PropertyTypeCategorical},
	},
	CAFilterAdGroup: {
		"id":                PropertiesAndRelated{TypeOfProperty: U.PropertyTypeCategorical},
		"name":              PropertiesAndRelated{TypeOfProperty: U.PropertyTypeCategorical},
		"daily_budget":      PropertiesAndRelated{TypeOfProperty: U.PropertyTypeCategorical},
		"lifetime_budget":   PropertiesAndRelated{TypeOfProperty: U.PropertyTypeCategorical},
		"configured_status": PropertiesAndRelated{TypeOfProperty: U.PropertyTypeCategorical},
		"effective_status":  PropertiesAndRelated{TypeOfProperty: U.PropertyTypeCategorical},
		"objective":         PropertiesAndRelated{TypeOfProperty: U.PropertyTypeCategorical},
		"bid_strategy":      PropertiesAndRelated{TypeOfProperty: U.PropertyTypeCategorical},
	},
	CAFilterAd: {
		"id":                PropertiesAndRelated{TypeOfProperty: U.PropertyTypeCategorical},
		"name":              PropertiesAndRelated{TypeOfProperty: U.PropertyTypeCategorical},
		"configured_status": PropertiesAndRelated{TypeOfProperty: U.PropertyTypeCategorical},
		"effective_status":  PropertiesAndRelated{TypeOfProperty: U.PropertyTypeCategorical},
	},
}

// To change the below when facebook objectAndPropertyToValueInFacebookReportsMapping or any other changes.
var ObjectToValueInFacebookJobsMapping = map[string]string{
	"campaign:daily_budget":      "campaign_daily_budget",
	"campaign:lifetime_budget":   "campaign_lifetime_budget",
	"campaign:configured_status": "campaign_configured_status",
	"campaign:effective_status":  "campaign_effective_status",
	"campaign:objective":         "campaign_objective",
	"campaign:buying_type":       "campaign_buying_type",
	"campaign:bid_strategy":      "campaign_bid_strategy",
	"ad_set:daily_budget":        "ad_set_daily_budget",
	"ad_set:lifetime_budget":     "ad_set_lifetime_budget",
	"ad_set:configured_status":   "ad_set_configured_status",
	"ad_set:effective_status":    "ad_set_effective_status",
	"ad_set:objective":           "ad_set_objective",
	"ad_set:bid_strategy":        "ad_set_bid_strategy",
	"campaign:name":              "campaign_name",
	"ad_set:name":                "adset_name",
	"campaign:id":                "campaign_id",
	"ad_set:id":                  "ad_set_id",
	"ad:id":                      "ad_id",
	"ad:name":                    "adset_name",
	"ad:configured_status":       "ad_configured_status",
	"ad:effective_status":        "ad_effective_status",
}
var ObjectAndKeyInFacebookToPropertyMapping = map[string]string{
	"campaign:name": "campaign_name",
	"ad_group:name": "adset_name",
}

var FacebookExternalRepresentationToInternalRepresentation = map[string]string{
	"name":                          "name",
	"id":                            "id",
	"configured_status":             "configured_status",
	"effective_status":              "effective_status",
	"daily_budget":                  "daily_budget",
	"lifetime_budget":               "lifetime_budget",
	"objective":                     "objective",
	"bid_strategy":                  "bid_strategy",
	"buying_type":                   "buying_type",
	"impressions":                   "impressions",
	"clicks":                        "clicks",
	"link_clicks":                   "link_clicks",
	"spend":                         "spend",
	"video_p50_watched_actions":     "video_p50_watched_actions",
	"video_p25_watched_actions":     "video_p25_watched_actions",
	"video_30_sec_watched_actions":  "video_30_sec_watched_actions",
	"video_p100_watched_actions":    "video_p100_watched_actions",
	"video_p75_watched_actions":     "video_p75_watched_actions",
	"cost_per_click":                "cost_per_click",
	"cost_per_link_click":           "cost_per_link_click",
	"cost_per_thousand_impressions": "cost_per_thousand_impressions",
	"click_through_rate":            "click_through_rate",
	"link_click_through_rate":       "link_click_through_rate",
	"frequency":                     "frequency",
	"reach":                         "reach",
	"campaign":                      "campaign",
	"ad_group":                      "ad_set",
	"ad":                            "ad",
	"channel":                       "channel",
}

var FacebookInternalRepresentationToExternalRepresentation = map[string]string{
	"impressions":                   "impressions",
	"clicks":                        "clicks",
	"link_clicks":                   "link_clicks",
	"spend":                         "spend",
	"video_p50_watched_actions":     "video_p50_watched_actions",
	"video_p25_watched_actions":     "video_p25_watched_actions",
	"video_30_sec_watched_actions":  "video_30_sec_watched_actions",
	"video_p100_watched_actions":    "video_p100_watched_actions",
	"video_p75_watched_actions":     "video_p75_watched_actions",
	"cost_per_click":                "cost_per_click",
	"cost_per_link_click":           "cost_per_link_click",
	"cost_per_thousand_impressions": "cost_per_thousand_impressions",
	"click_through_rate":            "click_through_rate",
	"link_click_through_rate":       "link_click_through_rate",
	"frequency":                     "frequency",
	"reach":                         "reach",
	"campaign:name":                 "campaign_name",
	"campaign:id":                   "campaign_id",
	"campaign:daily_budget":         "campaign_daily_budget",
	"campaign:lifetime_budget":      "campaign_lifetime_budget",
	"campaign:configured_status":    "campaign_configured_status",
	"campaign:effective_status":     "campaign_effective_status",
	"campaign:objective":            "campaign_objective",
	"campaign:buying_type":          "campaign_buying_type",
	"campaign:bid_strategy":         "campaign_bid_strategy",
	"ad_set:id":                     "ad_group_id",
	"ad_set:name":                   "ad_group_name",
	"ad_set:daily_budget":           "ad_group_daily_budget",
	"ad_set:lifetime_budget":        "ad_group_lifetime_budget",
	"ad_set:configured_status":      "ad_group_configured_status",
	"ad_set:effective_status":       "ad_group_effective_status",
	"ad_set:objective":              "ad_group_objective",
	"ad_set:bid_strategy":           "ad_group_bid_strategy",
	"ad:id":                         "ad_id",
	"ad:name":                       "ad_name",
	"ad:configured_status":          "ad_configured_status",
	"ad:effective_status":           "ad_effective_status",
	"channel:name":                  "channel_name",
}
var FacebookObjectMapForSmartProperty = map[string]string{
	"campaign": "campaign",
	"ad_set":   "ad_group",
}

var ObjectsForFacebook = []string{CAFilterCampaign, CAFilterAdGroup, CAFilterAd}

const (
	FacebookSpecificError = "Failed in facebook with the following error."
	CAFilterCampaign      = "campaign"
	CAFilterAdGroup       = "ad_group"
	CAFilterAd            = "ad"
)
