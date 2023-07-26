package model

import ()

// static list of all features, components, integrations and configurations.
// add to the list in case of new features.

const (
	// analyse section
	FEATURE_EVENTS      = "events"
	FEATURE_FUNNELS     = "funnels"
	FEATURE_KPIS        = "kpis"
	FEATURE_ATTRIBUTION = "attribution"
	FEATURE_PROFILES    = "profiles"
	FEATURE_TEMPLATES   = "templates"

	FEATURE_HUBSPOT                 = "hubspot"
	FEATURE_SALESFORCE              = "salesforce"
	FEATURE_LEADSQUARED             = "leadsqaured"
	FEATURE_GOOGLE_ADS              = "google_ads"
	FEATURE_FACEBOOK                = "facebook"
	FEATURE_LINKEDIN                = "linkedin"
	FEATURE_GOOGLE_ORGANIC          = "google_organic"
	FEATURE_BING_ADS                = "bing_ads"
	FEATURE_MARKETO                 = "marketo"
	FEATURE_DRIFT                   = "drift"
	FEATURE_CLEARBIT                = "clearbit"
	FEATURE_SIX_SIGNAL              = "six_signal"
	FEATURE_DASHBOARD               = "dashboard"
	FEATURE_OFFLINE_TOUCHPOINTS     = "offline_touchpoints p"
	FEATURE_SAVED_QUERIES           = "saved_queries"
	FEATURE_EXPLAIN                 = "explain_feature" // explain is a keyword in memsql.
	FEATURE_FILTERS                 = "filters"
	FEATURE_SHAREABLE_URL           = "shareable_url"
	FEATURE_CUSTOM_METRICS          = "custom_metrics"
	FEATURE_SMART_EVENTS            = "smart_events"
	FEATURE_SMART_PROPERTIES        = "smart_properties"
	FEATURE_CONTENT_GROUPS          = "content_groups"
	FEATURE_DISPLAY_NAMES           = "display_names"
	FEATURE_WEEKLY_INSIGHTS         = "weekly_insights"
	FEATURE_KPI_ALERTS              = "kpi_alerts"
	FEATURE_EVENT_BASED_ALERTS      = "event_based_alerts"
	FEATURE_REPORT_SHARING          = "report_sharing"
	FEATURE_SLACK                   = "slack"
	FEATURE_SEGMENT                 = "segment"
	FEATURE_PATH_ANALYSIS           = "path_analysis"
	FEATURE_ARCHIVE_EVENTS          = "archive_events"
	FEATURE_BIG_QUERY_UPLOAD        = "big_query_upload"
	FEATURE_IMPORT_ADS              = "import_ads"
	FEATURE_LEADGEN                 = "leadgen"
	FEATURE_TEAMS                   = "teams"
	FEATURE_SIX_SIGNAL_REPORT       = "six_signal_report"
	FEATURE_ACCOUNT_SCORING         = "account_scoring"
	FEATURE_ENGAGEMENT              = "engagements"
	FEATURE_FACTORS_DEANONYMISATION = "factors_deanonymisation"
	FEATURE_WEBHOOK                 = "webhook"
	FEATURE_ACCOUNT_PROFILES        = "account_profiles"
	FEATURE_PEOPLE_PROFILES         = "people_profiles"
	FEATURE_CLICKABLE_ELEMENTS      = "clickable_elements"
	FEATURE_G2

	// INTEGRATION
	INT_SHOPFIY                 = "int_shopify"
	INT_ADWORDS                 = "int_adwords"
	INT_GOOGLE_ORGANIC          = "int_google_organic"
	INT_FACEBOOK                = "int_facebook"
	INT_LINKEDIN                = "int_linkedin"
	INT_SALESFORCE              = "int_salesforce"
	INT_HUBSPOT                 = "int_hubspot"
	INT_DELETE                  = "int_delete"
	INT_SLACK                   = "int_slack"
	INT_TEAMS                   = "int_teams"
	INT_SEGMENT                 = "int_segment"
	INT_RUDDERSTACK             = "int_rudderstack"
	INT_MARKETO                 = "int_marketo"
	INT_DRIFT                   = "int_drift"
	INT_BING_ADS                = "int_bing_ads"
	INT_CLEARBIT                = "int_clear_bit"
	INT_LEADSQUARED             = "int_leadsquared"
	INT_SIX_SIGNAL              = "int_six_signal"
	INT_FACTORS_DEANONYMISATION = "int_factors_deanonymistaion"
	INT_G2                      = "intg2"

	// DATA SERVICE
	DS_ADWORDS        = "ds_adwords"
	DS_GOOGLE_ORGANIC = "ds_google_oraganic"
	DS_HUBSPOT        = "ds_hubspot"
	DS_FACEBOOK       = "ds_facebook"
	DS_LINKEDIN       = "ds_linkedin"
	DS_METRICS        = "ds_metrics"

	// CONFIGURATIONS
	CONF_ATTRUBUTION_SETTINGS = "conf_attribution_settings"
	CONF_CUSTOM_EVENTS        = "conf_custom_events"
	CONF_CUSTOM_PROPERTIES    = "conf_custom_properties"
	CONF_CONTENT_GROUPS       = "conf_content_groups"
	CONF_TOUCHPOINTS          = "conf_touchpoints"
	CONF_CUSTOM_KPIS          = "conf_custom_kpis"
	CONF_ALERTS               = "conf_alerts"
)

func GetAllAvailableFeatures() []string {
	constArray := []string{
		FEATURE_EVENTS, FEATURE_FUNNELS, FEATURE_KPIS, FEATURE_ATTRIBUTION, FEATURE_PROFILES, FEATURE_TEMPLATES,
		FEATURE_HUBSPOT, FEATURE_SALESFORCE, FEATURE_LEADSQUARED, FEATURE_GOOGLE_ADS, FEATURE_FACEBOOK, FEATURE_LINKEDIN,
		FEATURE_GOOGLE_ORGANIC, FEATURE_BING_ADS, FEATURE_MARKETO, FEATURE_DRIFT, FEATURE_CLEARBIT, FEATURE_SIX_SIGNAL,
		FEATURE_DASHBOARD, FEATURE_OFFLINE_TOUCHPOINTS, FEATURE_SAVED_QUERIES, FEATURE_EXPLAIN, FEATURE_FILTERS,
		FEATURE_SHAREABLE_URL, FEATURE_CUSTOM_METRICS, FEATURE_SMART_EVENTS, FEATURE_SMART_PROPERTIES,
		FEATURE_CONTENT_GROUPS, FEATURE_DISPLAY_NAMES, FEATURE_WEEKLY_INSIGHTS, FEATURE_KPI_ALERTS,
		FEATURE_EVENT_BASED_ALERTS, FEATURE_REPORT_SHARING, FEATURE_SLACK, FEATURE_SEGMENT, FEATURE_PATH_ANALYSIS,
		FEATURE_ARCHIVE_EVENTS, FEATURE_BIG_QUERY_UPLOAD, FEATURE_IMPORT_ADS, FEATURE_LEADGEN, FEATURE_TEAMS,
		FEATURE_SIX_SIGNAL_REPORT, FEATURE_ACCOUNT_SCORING, FEATURE_ENGAGEMENT, FEATURE_FACTORS_DEANONYMISATION,
		FEATURE_WEBHOOK, FEATURE_ACCOUNT_PROFILES, FEATURE_PEOPLE_PROFILES, FEATURE_CLICKABLE_ELEMENTS, FEATURE_G2,
		INT_SHOPFIY, INT_ADWORDS, INT_GOOGLE_ORGANIC, INT_FACEBOOK, INT_LINKEDIN, INT_SALESFORCE, INT_HUBSPOT,
		INT_DELETE, INT_SLACK, INT_TEAMS, INT_SEGMENT, INT_RUDDERSTACK, INT_MARKETO, INT_DRIFT, INT_BING_ADS,
		INT_CLEARBIT, INT_LEADSQUARED, INT_SIX_SIGNAL, INT_FACTORS_DEANONYMISATION, INT_G2,
		DS_ADWORDS, DS_GOOGLE_ORGANIC, DS_HUBSPOT, DS_FACEBOOK, DS_LINKEDIN, DS_METRICS,
		CONF_ATTRUBUTION_SETTINGS, CONF_CUSTOM_EVENTS, CONF_CUSTOM_PROPERTIES, CONF_CONTENT_GROUPS,
		CONF_TOUCHPOINTS, CONF_CUSTOM_KPIS, CONF_ALERTS,
	}
	return constArray
}
