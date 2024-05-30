package model

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
	FEATURE_SEGMENTKPI_OVERVIEW     = "segment_kpi"
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
	FEATURE_FACTORS_DEANONYMISATION = "factors_deanonymisation"
	FEATURE_WEBHOOK                 = "webhook"
	FEATURE_ACCOUNT_PROFILES        = "account_profiles"
	FEATURE_PEOPLE_PROFILES         = "people_profiles"
	FEATURE_CLICKABLE_ELEMENTS      = "clickable_elements"
	FEATURE_WEB_ANALYTICS_DASHBOARD = "web_analytics_dashboard"
	FEATURE_G2                      = "g2"
	FEATURE_RUDDERSTACK             = "rudderstack"
	FEATURE_WORKFLOWS               = "workflows"
	CONF_CUSTOM_PROPERTIES          = "conf_custom_properties"
	CONF_CUSTOM_EVENTS              = "conf_custom_events"
	FEATURE_LINKEDIN_FREQ_CAPPING   = "linkedin_frequency_capping"
)

func GetAllAvailableFeatures() []string {
	constArray := []string{
		FEATURE_EVENTS, FEATURE_FUNNELS, FEATURE_KPIS, FEATURE_ATTRIBUTION, FEATURE_PROFILES, FEATURE_TEMPLATES,
		FEATURE_HUBSPOT, FEATURE_SALESFORCE, FEATURE_LEADSQUARED, FEATURE_GOOGLE_ADS, FEATURE_FACEBOOK, FEATURE_LINKEDIN,
		FEATURE_GOOGLE_ORGANIC, FEATURE_BING_ADS, FEATURE_MARKETO, FEATURE_DRIFT, FEATURE_CLEARBIT, FEATURE_SIX_SIGNAL,
		FEATURE_DASHBOARD, FEATURE_OFFLINE_TOUCHPOINTS, FEATURE_SAVED_QUERIES, FEATURE_EXPLAIN, FEATURE_FILTERS,
		FEATURE_SHAREABLE_URL, FEATURE_CUSTOM_METRICS, FEATURE_SMART_PROPERTIES,
		FEATURE_CONTENT_GROUPS, FEATURE_DISPLAY_NAMES, FEATURE_WEEKLY_INSIGHTS, FEATURE_KPI_ALERTS,
		FEATURE_EVENT_BASED_ALERTS, FEATURE_REPORT_SHARING, FEATURE_SLACK, FEATURE_SEGMENT, FEATURE_PATH_ANALYSIS,
		FEATURE_ARCHIVE_EVENTS, FEATURE_BIG_QUERY_UPLOAD, FEATURE_IMPORT_ADS, FEATURE_LEADGEN, FEATURE_TEAMS,
		FEATURE_SIX_SIGNAL_REPORT, FEATURE_ACCOUNT_SCORING, FEATURE_FACTORS_DEANONYMISATION,
		FEATURE_WEBHOOK, FEATURE_ACCOUNT_PROFILES, FEATURE_PEOPLE_PROFILES, FEATURE_CLICKABLE_ELEMENTS, FEATURE_G2,
		FEATURE_WEB_ANALYTICS_DASHBOARD, FEATURE_RUDDERSTACK, FEATURE_WORKFLOWS, CONF_CUSTOM_EVENTS, CONF_CUSTOM_PROPERTIES,
		FEATURE_SEGMENTKPI_OVERVIEW,
	}
	return constArray
}
