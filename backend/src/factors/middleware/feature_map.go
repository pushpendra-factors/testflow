package middleware

import ()

type Features []string

// handler name to features associated with it.
var featureMap map[string][]string

// list of features
const (
	HUBSPOT             = "hubspot"
	SALESFORCE          = "salesforce"
	LEADSQUARED         = "leadsqaured"
	GOOGLE_ADS          = "google_ads"
	FACEBOOK            = "facebook"
	LINKEDIN            = "linkedin"
	GOOGLE_ORGANIC      = "google_organic"
	BING_ADS            = "BING_ADS"
	MARKETO             = "marketo"
	DRIFT               = "drift"
	CLEARBIT            = "clearbit"
	SIX_SIGNAL          = "six_signal"
	DASHBOARD           = "dashboard"
	OFFLINE_TOUCHPOINTS = "offline_touchpoints"
	SAVED_QUERIES       = "saved_queries"
	EXPLAIN             = "explain"
	FILTERS             = "filters"
	SHAREABLE_URL       = "shareable_url"
	CUSTOM_METRICS      = "custom_metrics"
	SMART_EVENTS        = "smart_events"
	TEMPLATES           = "templates"
	SMART_PROPERTIES    = "smart_properties"
	CONTENT_GROUPS      = "content_groups"
	DISPLAY_NAMES       = "display_names"
	WEEKLY_INSIGHTS     = "weekly_insights"
	ALERTS              = "alerts"
	SlACK               = "slack"
	PROFILES            = "profiles"
	SEGMENT             = "segment"
	PATH_ANALYSIS       = "path_analysis"
	ARCHIVE_EVENTS      = "archive_events"
	BIG_QUERY_UPLOAD    = "big_query_upload"
	IMPORT_ADS          = "import_ads"
	LEADGEN             = "leadgen"

	// INTEGRAION
	INT_SHOPFIY        = "int_shopify"
	INT_ADWORDS        = "int_adwords"
	INT_GOOGLE_ORGANIC = "int_google_organic"
	INT_FACEBOOK       = "int_facebook"
	INT_LINKEDIN       = "int_linkedin"
	INT_SALESFORCE     = "int_salesforce"
	INT_HUBSPOT        = "int_hubspot"
	INT_DELETE         = "int_delete"
	INT_SLACK          = "int_slack"

	// DATA SERVICE
	DS_ADWORDS        = "ds_adwords"
	DS_GOOGLE_ORGANIC = "ds_google_oraganic"
	DS_HUBSPOT        = "ds_hubspot"
	DS_FACEBOOK       = "ds_facebook"
	DS_LINKEDIN       = "ds_linkedin"
	DS_METRICS        = "ds_metrics"
)

func initFeatureMap() {
	featureMap = make(map[string][]string)
	// hubspot
	featureMap["HubspotAuthRedirectHandler"] = []string{HUBSPOT}
	featureMap["HubspotCallbackHandler"] = []string{HUBSPOT}

	// salesforce
	featureMap["IntEnableSalesforceHandler"] = []string{SALESFORCE}
	featureMap["SalesforceAuthRedirectHandler"] = []string{SALESFORCE}
	featureMap["SalesforceCallbackHandler"] = []string{SALESFORCE}

	// leadsquared
	featureMap["UpdateLeadSquaredConfigHandler"] = []string{LEADSQUARED}
	featureMap["RemoveLeadSquaredConfigHandler"] = []string{LEADSQUARED}

	// google_ads
	featureMap["IntEnableAdwordsHandler"] = []string{GOOGLE_ADS}

	// delete_integrations
	featureMap["IntDeleteHandler"] = []string{GOOGLE_ADS, FACEBOOK, LINKEDIN, GOOGLE_ORGANIC}

	// facebook
	featureMap["IntFacebookAddAccessTokenHandler"] = []string{FACEBOOK}

	// linkedin
	featureMap["IntLinkedinAuthHandler"] = []string{LINKEDIN}
	featureMap["IntLinkedinAccountHandler"] = []string{LINKEDIN}
	featureMap["IntLinkedinAddAccessTokenHandler"] = []string{LINKEDIN}

	// google_organic
	featureMap["IntEnableGoogleOrganicHandler"] = []string{GOOGLE_ORGANIC}

	// dashboard
	featureMap["GetDashboardsHandler"] = []string{DASHBOARD}
	featureMap["CreateDashboardHandler"] = []string{DASHBOARD}
	featureMap["UpdateDashboardHandler"] = []string{DASHBOARD}
	featureMap["GetDashboardUnitsHandler"] = []string{DASHBOARD}
	featureMap["CreateDashboardUnitHandler"] = []string{DASHBOARD}
	featureMap["UpdateDashboardUnitHandler"] = []string{DASHBOARD}
	featureMap["DeleteDashboardUnitHandler"] = []string{DASHBOARD}
	featureMap["SearchTemplateHandler"] = []string{DASHBOARD}
	featureMap["GetDashboardTemplatesHandler"] = []string{DASHBOARD}
	featureMap["CreateTemplateHandler"] = []string{DASHBOARD}
	featureMap["GenerateDashboardFromTemplateHandler"] = []string{DASHBOARD}
	featureMap["GenerateTemplateFromDashboardHandler"] = []string{DASHBOARD}
	featureMap["CreateDashboardUnitForMultiDashboardsHandler"] = []string{DASHBOARD}
	featureMap["CreateDashboardUnitsForMultipleQueriesHandler"] = []string{DASHBOARD}
	featureMap["DeleteMultiDashboardUnitHandler"] = []string{DASHBOARD}
	featureMap["DeleteDashboardHandler"] = []string{DASHBOARD}

	// Offline Touchpoints

	featureMap["GetOTPRuleHandler"] = []string{OFFLINE_TOUCHPOINTS}
	featureMap["CreateOTPRuleHandler"] = []string{OFFLINE_TOUCHPOINTS}
	featureMap["UpdateOTPRuleHandler"] = []string{OFFLINE_TOUCHPOINTS}
	featureMap["SearchOTPRuleHandler"] = []string{OFFLINE_TOUCHPOINTS}
	featureMap["DeleteOTPRuleHandler"] = []string{OFFLINE_TOUCHPOINTS}

	// Saved Queries

	featureMap["GetQueriesHandler"] = []string{SAVED_QUERIES}
	featureMap["CreateQueryHandler"] = []string{SAVED_QUERIES}
	featureMap["UpdateSavedQueryHandler"] = []string{SAVED_QUERIES}
	featureMap["DeleteSavedQueryHandler"] = []string{SAVED_QUERIES}
	featureMap["SearchQueriesHandler"] = []string{SAVED_QUERIES}

	// explain
	featureMap["GetProjectModelsHandler"] = []string{EXPLAIN}
	featureMap["FactorHandler"] = []string{EXPLAIN}
	featureMap["CreateFactorsTrackedEventsHandler"] = []string{EXPLAIN}
	featureMap["RemoveFactorsTrackedEventsHandler"] = []string{EXPLAIN}
	featureMap["GetAllFactorsTrackedEventsHandler"] = []string{EXPLAIN}
	featureMap["GetAllGroupedFactorsTrackedEventsHandler"] = []string{EXPLAIN}
	featureMap["CreateFactorsTrackedUserPropertyHandler"] = []string{EXPLAIN}
	featureMap["RemoveFactorsTrackedUserPropertyHandler"] = []string{EXPLAIN}
	featureMap["GetAllFactorsTrackedUserPropertiesHandler"] = []string{EXPLAIN}
	featureMap["CreateFactorsGoalsHandler"] = []string{EXPLAIN}
	featureMap["RemoveFactorsGoalsHandler"] = []string{EXPLAIN}
	featureMap["GetAllFactorsGoalsHandler"] = []string{EXPLAIN}
	featureMap["UpdateFactorsGoalsHandler"] = []string{EXPLAIN}
	featureMap["SearchFactorsGoalHandler"] = []string{EXPLAIN}
	featureMap["PostFactorsHandler"] = []string{EXPLAIN}
	featureMap["PostFactorsCompareHandler"] = []string{EXPLAIN}
	featureMap["GetModelMetaData"] = []string{EXPLAIN}
	featureMap["GetFactorsHandler"] = []string{EXPLAIN}
	featureMap["GetFactorsHandlerV2"] = []string{EXPLAIN}
	featureMap["PostFactorsHandlerV2"] = []string{EXPLAIN}
	featureMap["CreateExplainV2EntityHandler"] = []string{EXPLAIN}
	featureMap["DeleteSavedExplainV2EntityHandler"] = []string{EXPLAIN}

	// Filters
	featureMap["GetFiltersHandler"] = []string{FILTERS}
	featureMap["CreateFilterHandler"] = []string{FILTERS}
	featureMap["UpdateFilterHandler"] = []string{FILTERS}
	featureMap["DeleteFilterHandler"] = []string{FILTERS}

	// shareable urls
	featureMap["GetShareableURLsHandler"] = []string{SHAREABLE_URL}
	featureMap["CreateShareableURLHandler"] = []string{SHAREABLE_URL}
	featureMap["DeleteShareableURLHandler"] = []string{SHAREABLE_URL}
	featureMap["RevokeShareableURLHandler"] = []string{SHAREABLE_URL}

	// custom metrics
	featureMap["GetCustomMetricsConfigV1"] = []string{CUSTOM_METRICS}
	featureMap["CreateCustomMetric"] = []string{CUSTOM_METRICS}
	featureMap["GetCustomMetrics"] = []string{CUSTOM_METRICS}
	featureMap["DeleteCustomMetrics"] = []string{CUSTOM_METRICS}
	featureMap["CreateMissingPreBuiltCustomKPI"] = []string{CUSTOM_METRICS}

	// smart events
	featureMap["GetSmartEventFiltersHandler"] = []string{SMART_EVENTS}
	featureMap["CreateSmartEventFilterHandler"] = []string{SMART_EVENTS}
	featureMap["UpdateSmartEventFilterHandler"] = []string{SMART_EVENTS}
	featureMap["DeleteSmartEventFilterHandler"] = []string{SMART_EVENTS}

	// templates
	featureMap["GetTemplateConfigHandler"] = []string{TEMPLATES}
	featureMap["UpdateTemplateConfigHandler"] = []string{TEMPLATES}
	featureMap["ExecuteTemplateQueryHandler"] = []string{TEMPLATES}

	// smart properties
	featureMap["GetSmartPropertyRulesConfigHandler"] = []string{SMART_PROPERTIES}
	featureMap["CreateSmartPropertyRulesHandler"] = []string{SMART_PROPERTIES}
	featureMap["GetSmartPropertyRulesHandler"] = []string{SMART_PROPERTIES}
	featureMap["GetSmartPropertyRuleByRuleIDHandler"] = []string{SMART_PROPERTIES}
	featureMap["UpdateSmartPropertyRulesHandler"] = []string{SMART_PROPERTIES}
	featureMap["DeleteSmartPropertyRulesHandler"] = []string{SMART_PROPERTIES}

	// content groups
	featureMap["CreateContentGroupHandler"] = []string{CONTENT_GROUPS}
	featureMap["GetContentGroupHandler"] = []string{CONTENT_GROUPS}
	featureMap["GetContentGroupByIDHandler"] = []string{CONTENT_GROUPS}
	featureMap["UpdateContentGroupHandler"] = []string{CONTENT_GROUPS}
	featureMap["DeleteContentGroupHandler"] = []string{CONTENT_GROUPS}

	// display names
	featureMap["CreateDisplayNamesHandler"] = []string{DISPLAY_NAMES}
	featureMap["GetAllDistinctEventProperties"] = []string{DISPLAY_NAMES}

	// weekly insights
	featureMap["GetWeeklyInsightsHandler"] = []string{WEEKLY_INSIGHTS}
	featureMap["GetWeeklyInsightsMetadata"] = []string{WEEKLY_INSIGHTS}
	featureMap["PostFeedbackHandler"] = []string{WEEKLY_INSIGHTS}

	// bing ads
	featureMap["CreateBingAdsIntegration"] = []string{BING_ADS}
	featureMap["DisableBingAdsIntegration"] = []string{BING_ADS}
	featureMap["GetBingAdsIntegration"] = []string{BING_ADS}
	featureMap["EnableBingAdsIntegration"] = []string{BING_ADS}

	// marketo
	featureMap["CreateMarketoIntegration"] = []string{MARKETO}
	featureMap["DisableMarketoIntegration"] = []string{MARKETO}
	featureMap["GetMarketoIntegration"] = []string{MARKETO}
	featureMap["EnableMarketoIntegration"] = []string{MARKETO}

	// alerts
	featureMap["CreateAlertHandler"] = []string{ALERTS}
	featureMap["GetAlertsHandler"] = []string{ALERTS}
	featureMap["GetAlertByIDHandler"] = []string{ALERTS}
	featureMap["DeleteAlertHandler"] = []string{ALERTS}
	featureMap["EditAlertHandler"] = []string{ALERTS}
	featureMap["QuerySendNowHandler"] = []string{ALERTS}

	// slack
	featureMap["SlackAuthRedirectHandler"] = []string{SlACK}
	featureMap["GetSlackChannelsListHandler"] = []string{SlACK}
	featureMap["DeleteSlackIntegrationHandler"] = []string{SlACK}
	featureMap["SlackCallbackHandler"] = []string{SlACK}

	// profiles
	featureMap["GetProfileUsersHandler"] = []string{PROFILES}
	featureMap["GetProfileUserDetailsHandler"] = []string{PROFILES}
	featureMap["GetProfileAccountsHandler"] = []string{PROFILES}
	featureMap["GetProfileAccountDetailsHandler"] = []string{PROFILES}

	// segement
	featureMap["CreateSegmentHandler"] = []string{SEGMENT}
	featureMap["GetSegmentsHandler"] = []string{SEGMENT}
	featureMap["GetSegmentByIdHandler"] = []string{SEGMENT}
	featureMap["UpdateSegmentHandler"] = []string{SEGMENT}
	featureMap["DeleteSegmentByIdHandler"] = []string{SEGMENT}

	// path analysis
	featureMap["GetPathAnalysisEntityHandler"] = []string{PATH_ANALYSIS}
	featureMap["CreatePathAnalysisEntityHandler"] = []string{PATH_ANALYSIS}
	featureMap["DeleteSavedPathAnalysisEntityHandler"] = []string{PATH_ANALYSIS}
	featureMap["GetPathAnalysisData"] = []string{PATH_ANALYSIS}

	// INTEGRATION
	featureMap["IntShopifyHandler"] = []string{INT_SHOPFIY}
	featureMap["IntShopifySDKHandler"] = []string{INT_SHOPFIY}
	featureMap["IntEnableAdwordsHandler"] = []string{INT_ADWORDS}
	featureMap["IntEnableGoogleOrganicHandler"] = []string{INT_GOOGLE_ORGANIC}
	featureMap["IntFacebookAddAccessTokenHandler"] = []string{INT_FACEBOOK}
	featureMap["IntLinkedinAuthHandler"] = []string{INT_LINKEDIN}
	featureMap["IntLinkedinAccountHandler"] = []string{INT_LINKEDIN}
	featureMap["IntLinkedinAddAccessTokenHandler"] = []string{INT_LINKEDIN}
	featureMap["IntEnableSalesforceHandler"] = []string{INT_SALESFORCE}
	featureMap["SalesforceAuthRedirectHandler"] = []string{INT_SALESFORCE}
	featureMap["SalesforceCallbackHandler"] = []string{INT_SALESFORCE}
	featureMap["HubspotAuthRedirectHandler"] = []string{INT_HUBSPOT}
	featureMap["HubspotCallbackHandler"] = []string{INT_HUBSPOT}
	featureMap["IntDeleteHandler"] = []string{INT_DELETE}
	featureMap["SlackCallbackHandler"] = []string{INT_SLACK}

	// DATA SERVICE
	featureMap["DataServiceAdwordsAddDocumentHandler"] = []string{DS_ADWORDS}
	featureMap["DataServiceAdwordsAddMultipleDocumentsHandler"] = []string{DS_ADWORDS}
	featureMap["IntAdwordsAddRefreshTokenHandler"] = []string{DS_ADWORDS}
	featureMap["IntAdwordsGetRefreshTokenHandler"] = []string{DS_ADWORDS}
	featureMap["DataServiceAdwordsGetLastSyncForProjectInfoHandler"] = []string{DS_ADWORDS}
	featureMap["DataServiceAdwordsGetLastSyncInfoHandler"] = []string{DS_ADWORDS}
	featureMap["DataServiceGoogleOrganicAddDocumentHandler"] = []string{DS_GOOGLE_ORGANIC}
	featureMap["DataServiceGoogleOrganicAddMultipleDocumentsHandler"] = []string{DS_GOOGLE_ORGANIC}
	featureMap["IntGoogleOrganicAddRefreshTokenHandler"] = []string{DS_GOOGLE_ORGANIC}
	featureMap["IntGoogleOrganicGetRefreshTokenHandler"] = []string{DS_GOOGLE_ORGANIC}
	featureMap["DataServiceGoogleOrganicGetLastSyncForProjectInfoHandler"] = []string{DS_GOOGLE_ORGANIC}
	featureMap["DataServiceGoogleOrganicGetLastSyncInfoHandler"] = []string{DS_GOOGLE_ORGANIC}
	featureMap["DataServiceHubspotAddDocumentHandler"] = []string{DS_HUBSPOT}
	featureMap["DataServiceHubspotAddBatchDocumentHandler"] = []string{DS_HUBSPOT}
	featureMap["DataServiceHubspotGetSyncInfoHandler"] = []string{DS_HUBSPOT}
	featureMap["DataServiceHubspotUpdateSyncInfo"] = []string{DS_HUBSPOT}
	featureMap["DataServiceGetHubspotFormDocumentsHandler"] = []string{DS_HUBSPOT}
	featureMap["DataServiceFacebookGetProjectSettings"] = []string{DS_FACEBOOK}
	featureMap["DataServiceFacebookAddDocumentHandler"] = []string{DS_FACEBOOK}
	featureMap["DataServiceFacebookGetLastSyncInfoHandler"] = []string{DS_FACEBOOK}
	featureMap["DataServiceLinkedinGetLastSyncInfoHandler"] = []string{DS_LINKEDIN}
	featureMap["DataServiceLinkedinAddDocumentHandler"] = []string{DS_LINKEDIN}
	featureMap["DataServiceLinkedinUpdateAccessToken"] = []string{DS_FACEBOOK}
	featureMap["DataServiceRecordMetricHandler"] = []string{DS_METRICS}
	featureMap["DataServiceLinkedinGetProjectSettings"] = []string{DS_LINKEDIN}

}

func GetFeatureMap() map[string][]string {
	initFeatureMap()
	return featureMap
}
