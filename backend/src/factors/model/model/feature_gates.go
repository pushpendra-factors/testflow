package model

import (
	"time"
)

// 0 if feature not available
// 1 if available but not enabled
// 2 if enabled
type FeatureGate struct {
	ProjectID          int64     `json:"project_id"`
	Hubspot            int       `json:"hubspot"`
	Salesforce         int       `json:"salesforce"`
	Leadsquared        int       `json:"leadsquared"`
	GoogleAds          int       `json:"google_ads"`
	Facebook           int       `json:"facebook"`
	Linkedin           int       `json:"linkedin"`
	GoogleOrganic      int       `json:"google_organic"`
	BingAds            int       `json:"bing_ads"`
	Marketo            int       `json:"marketo"`
	Drift              int       `json:"drift"`
	Clearbit           int       `json:"clearbit"`
	SixSignal          int       `json:"six_signal"`
	Dashboard          int       `json:"dashboard"`
	OfflineTouchpoints int       `json:"offline_touchpoints"`
	SavedQueries       int       `json:"saved_queries"`
	ExplainFeature     int       `json:"explain_feature"`
	Filters            int       `json:"filters"`
	ShareableURL       int       `json:"shareable_url"`
	CustomMetrics      int       `json:"custom_metrics"`
	SmartEvents        int       `json:"smart_events"`
	Templates          int       `json:"templates"`
	SmartProperties    int       `json:"smart_properties"`
	ContentGroups      int       `json:"content_groups"`
	DisplayNames       int       `json:"display_names"`
	WeeklyInsights     int       `json:"weekly_insights"`
	Alerts             int       `json:"alerts"`
	Slack              int       `json:"slack"`
	Profiles           int       `json:"profiles"`
	Segment            int       `json:"segment"`
	PathAnalysis       int       `json:"path_analysis"`
	ArchiveEvents      int       `json:"archive_events"`
	BigQueryUpload     int       `json:"big_query_upload"`
	ImportAds          int       `json:"import_ads"`
	Leadgen            int       `json:"leadgen"`
	IntShopify         int       `json:"int_shopify"`
	IntAdwords         int       `json:"int_adwords"`
	IntGoogleOrganic   int       `json:"int_google_organic"`
	IntFacebook        int       `json:"int_facebook"`
	IntLinkedin        int       `json:"int_linkedin"`
	IntSalesforce      int       `json:"int_salesforce"`
	IntHubspot         int       `json:"int_hubspot"`
	IntDelete          int       `json:"int_delete"`
	IntSlack           int       `json:"int_slack"`
	DsAdwords          int       `json:"ds_adwords"`
	DsGoogleOrganic    int       `json:"ds_google_organic"`
	DsHubspot          int       `json:"ds_hubspot"`
	DsFacebook         int       `json:"ds_facebook"`
	DsLinkedin         int       `json:"ds_linkedin"`
	DsMetrics          int       `json:"ds_metrics"`
	UpdatedAt          time.Time `json:"updated_at"`
}

var FeatureStatusTypeAlias = map[int]string{
	0: "unavailable",
	1: "disabled",
	2: "enabled",
}
