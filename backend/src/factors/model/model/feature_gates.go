package model

import (
	"time"
)

// 0 if feature not available
// 1 if available but not enabled
// 2 if enabled
type FeatureGate struct {
	ProjectID          int64     `json:"project_id"`
	Hubspot            int       `gorm:"default:2" json:"hubspot"`
	Salesforce         int       `gorm:"default:2" json:"salesforce"`
	Leadsquared        int       `gorm:"default:2" json:"leadsquared"`
	GoogleAds          int       `gorm:"default:2" json:"google_ads"`
	Facebook           int       `gorm:"default:2" json:"facebook"`
	Linkedin           int       `gorm:"default:2" json:"linkedin"`
	GoogleOrganic      int       `gorm:"default:2" json:"google_organic"`
	BingAds            int       `gorm:"default:2" json:"bing_ads"`
	Marketo            int       `gorm:"default:2" json:"marketo"`
	Drift              int       `gorm:"default:2" json:"drift"`
	Clearbit           int       `gorm:"default:2" json:"clearbit"`
	SixSignal          int       `gorm:"default:1" json:"six_signal"`
	Dashboard          int       `gorm:"default:2" json:"dashboard"`
	OfflineTouchpoints int       `gorm:"default:2" json:"offline_touchpoints"`
	SavedQueries       int       `gorm:"default:2" json:"saved_queries"`
	Explain            int       `gorm:"default:1" json:"explain"`
	Filters            int       `gorm:"default:2" json:"filters"`
	ShareableURL       int       `gorm:"default:2" json:"shareable_url"`
	CustomMetrics      int       `gorm:"default:2" json:"custom_metrics"`
	SmartEvents        int       `gorm:"default:2" json:"smart_events"`
	Templates          int       `gorm:"default:2" json:"templates"`
	SmartProperties    int       `gorm:"default:2" json:"smart_properties"`
	ContentGroups      int       `gorm:"default:2" json:"content_groups"`
	DisplayNames       int       `gorm:"default:2" json:"display_names"`
	WeeklyInsights     int       `gorm:"default:1" json:"weekly_insights"`
	Alerts             int       `gorm:"default:2" json:"alerts"`
	Slack              int       `gorm:"default:2" json:"slack"`
	Profiles           int       `gorm:"default:2" json:"profiles"`
	Segment            int       `gorm:"default:2" json:"segment"`
	PathAnalysis       int       `gorm:"default:1" json:"path_analysis"`
	ArchiveEvents      int       `gorm:"default:1" json:"archive_events"`
	BigQueryUpload     int       `gorm:"default:1" json:"big_query_upload"`
	ImportAds          int       `gorm:"default:2" json:"import_ads"`
	Leadgen            int       `gorm:"default:2" json:"leadgen"`
	IntShopify         int       `gorm:"default:2" json:"int_shopify"`
	IntAdwords         int       `gorm:"default:2" json:"int_adwords"`
	IntGoogleOrganic   int       `gorm:"default:2" json:"int_google_organic"`
	IntFacebook        int       `gorm:"default:2" json:"int_facebook"`
	IntLinkedin        int       `gorm:"default:2" json:"int_linkedin"`
	IntSalesforce      int       `gorm:"default:2" json:"int_salesforce"`
	IntHubspot         int       `gorm:"default:2" json:"int_hubspot"`
	IntDelete          int       `gorm:"default:2" json:"int_delete"`
	IntSlack           int       `gorm:"default:2" json:"int_slack"`
	DsAdwords          int       `gorm:"default:2" json:"ds_adwords"`
	DsGoogleOrganic    int       `gorm:"default:2" json:"ds_google_organic"`
	DsHubspot          int       `gorm:"default:2" json:"ds_hubspot"`
	DsFacebook         int       `gorm:"default:2" json:"ds_facebook"`
	DsLinkedin         int       `gorm:"default:2" json:"ds_linkedin"`
	DsMetrics          int       `gorm:"default:2" json:"ds_metrics"`
	UpdatedAt          time.Time `json:"updated_at"`
}

var FeatureStatusTypeAlias = map[int]string{
	0: "unavailable",
	1: "disabled",
	2: "enabled",
}
