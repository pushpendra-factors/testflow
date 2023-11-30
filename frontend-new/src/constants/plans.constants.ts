export const FEATURES = {
  //Features

  FEATURE_EVENTS: 'events',
  FEATURE_FUNNELS: 'funnels',
  FEATURE_KPIS: 'kpis',
  FEATURE_ATTRIBUTION: 'attribution',
  FEATURE_PROFILES: 'profiles',
  FEATURE_TEMPLATES: 'templates',
  FEATURE_HUBSPOT: 'hubspot',
  FEATURE_SALESFORCE: 'salesforce',
  FEATURE_LEADSQUARED: 'leadsqaured',
  FEATURE_GOOGLE_ADS: 'google_ads',
  FEATURE_FACEBOOK: 'facebook',
  FEATURE_LINKEDIN: 'linkedin',
  FEATURE_GOOGLE_ORGANIC: 'google_organic',
  FEATURE_BING_ADS: 'bing_ads',
  FEATURE_MARKETO: 'marketo',
  FEATURE_DRIFT: 'drift',
  FEATURE_CLEARBIT: 'clearbit',
  FEATURE_SIX_SIGNAL: 'six_signal',
  FEATURE_DASHBOARD: 'dashboard',
  FEATURE_OFFLINE_TOUCHPOINTS: 'offline_touchpoints p',
  FEATURE_SAVED_QUERIES: 'saved_queries',
  FEATURE_EXPLAIN: 'explain_feature',
  FEATURE_FILTERS: 'filters',
  FEATURE_SHAREABLE_URL: 'shareable_url',
  FEATURE_CUSTOM_METRICS: 'custom_metrics',
  FEATURE_SMART_EVENTS: 'smart_events',
  FEATURE_SMART_PROPERTIES: 'smart_properties',
  FEATURE_CONTENT_GROUPS: 'content_groups',
  FEATURE_DISPLAY_NAMES: 'display_names',
  FEATURE_WEEKLY_INSIGHTS: 'weekly_insights',
  FEATURE_KPI_ALERTS: 'kpi_alerts',
  FEATURE_EVENT_BASED_ALERTS: 'event_based_alerts',
  FEATURE_SlACK: 'slack',
  FEATURE_SEGMENT: 'segment',
  FEATURE_PATH_ANALYSIS: 'path_analysis',
  FEATURE_ARCHIVE_EVENTS: 'archive_events',
  FEATURE_BIG_QUERY_UPLOAD: 'big_query_upload',
  FEATURE_IMPORT_ADS: 'import_ads',
  FEATURE_LEADGEN: 'leadgen',
  FEATURE_TEAMS: 'teams',
  FEATURE_SIX_SIGNAL_REPORT: 'six_signal_report',
  FEATURE_ACCOUNT_SCORING: 'account_scoring',
  FEATURE_ENGAGEMENT: 'engagements',
  FEATURE_FACTORS_DEANONYMISATION: 'factors_deanonymisation',
  FEATURE_WEBHOOK: 'webhook',
  FEATURE_ACCOUNT_PROFILES: 'account_profiles',
  FEATURE_PEOPLE_PROFILES: 'people_profiles',
  FEATURE_REPORT_SHARING: 'report_sharing',
  FEATURE_WEB_ANALYTICS_DASHBOARD: 'web_analytics_dashboard',

  //Integrations
  INT_SHOPFIY: 'int_shopify',
  INT_ADWORDS: 'int_adwords',
  INT_GOOGLE_ORGANIC: 'int_google_organic',
  INT_FACEBOOK: 'int_facebook',
  INT_LINKEDIN: 'int_linkedin',
  INT_SALESFORCE: 'int_salesforce',
  INT_HUBSPOT: 'int_hubspot',
  INT_DELETE: 'int_delete',
  INT_SLACK: 'int_slack',
  INT_TEAMS: 'int_teams',
  INT_SEGMENT: 'int_segment',
  INT_RUDDERSTACK: 'int_rudderstack',
  INT_MARKETO: 'int_marketo',
  INT_DRIFT: 'int_drift',
  INT_BING_ADS: 'int_bing_ads',
  INT_CLEARBIT: 'int_clear_bit',
  INT_LEADSQUARED: 'int_leadsquared',
  INT_SIX_SIGNAL: 'int_six_signal',
  INT_FACTORS_DEANONYMISATION: 'int_factors_deanonymistaion',
  INT_G2: 'intg2',

  // DATA SERVICE
  DS_ADWORDS: 'ds_adwords',
  DS_GOOGLE_ORGANIC: 'ds_google_oraganic',
  DS_HUBSPOT: 'ds_hubspot',
  DS_FACEBOOK: 'ds_facebook',
  DS_LINKEDIN: 'ds_linkedin',
  DS_METRICS: 'ds_metrics',

  // CONFIGURATIONS
  CONF_ATTRUBUTION_SETTINGS: 'conf_attribution_settings',
  CONF_CUSTOM_EVENTS: 'conf_custom_events',
  CONF_CUSTOM_PROPERTIES: 'conf_custom_properties',
  CONF_CONTENT_GROUPS: 'conf_content_groups',
  CONF_TOUCHPOINTS: 'conf_touchpoints',
  CONF_CUSTOM_KPIPS: 'conf_custom_kpis',
  CONF_ALERTS: 'conf_alerts'
};

export const PLANS = {
  PLAN_FREE: 'Free',
  PLAN_GROWTH: 'Growth',
  PLAN_BASIC: 'Basic',
  PLAN_PROFESSIONAL: 'Professional',
  PLAN_CUSTOM: 'Custom'
};

//adding for backward compatibility will be removed once we fully move to chargebee
export const PLANS_V0 = {
  PLAN_FREE: 'FREE',
  PLAN_STARTUP: 'STARTUP',
  PLAN_BASIC: 'BASIC',
  PLAN_PROFESSIONAL: 'PROFESSIONAL',
  PLAN_CUSTOM: 'CUSTOM'
};

export const PLANS_COFIG: PLANS_COFIG_INTERFACE = {
  [PLANS.PLAN_FREE]: {
    name: PLANS.PLAN_FREE,
    description: 'Essential marketing tools to engage and convert leads',
    uniqueFeatures: [
      'Account identification',
      'Account enrichment',
      'Core analytics & reporting',
      'Account timelines',
      'Real-time alerts'
    ],
    isRecommendedPlan: false,
    planIcon: 'Userplus',
    planIconColor: '#40A9FF',
    mtuLimit: 5000,
    accountIdentifiedLimit: 100
  },

  [PLANS.PLAN_BASIC]: {
    name: PLANS.PLAN_BASIC,
    description:
      'Essential marketing tools with powerfull analytics and attribution',
    uniqueFeatures: [
      'Everything in Free +',
      'Custom events & KPIs',
      'Custom properties',
      'Content groups',
      'Onboarding support'
    ],
    isRecommendedPlan: false,
    planIcon: 'User_friends',
    planIconColor: '#73D13D',
    mtuLimit: 10000,
    accountIdentifiedLimit: 350
  },
  [PLANS.PLAN_GROWTH]: {
    name: PLANS.PLAN_GROWTH,
    description: 'Essential marketing tools to engage and convert leads',
    uniqueFeatures: [
      'Everything in Basic +',
      'Account & lead scoring',
      'Engaged channels (Coming Soon)',
      'Priority CSM'
    ],
    isRecommendedPlan: true,
    planIcon: 'User',
    planIconColor: '#36CFC9',
    mtuLimit: 50000,
    accountIdentifiedLimit: 5000
  },
  [PLANS.PLAN_PROFESSIONAL]: {
    name: PLANS.PLAN_PROFESSIONAL,
    description:
      'Essential marketing tools with powerfull analytics and attribution',
    uniqueFeatures: [
      'Everything in Growth +',
      'Multi-touch attribution',
      'Path analysis',
      'AI-fuelled Explain',
      'Dedicated CSM'
    ],
    isRecommendedPlan: true,
    planIcon: 'Buildings',
    planIconColor: '#FF7A45',
    mtuLimit: 100000,
    accountIdentifiedLimit: 10000
  }
};

export interface PLANS_COFIG_INTERFACE {
  [key: typeof PLANS[keyof typeof PLANS]]: PLAN_COFIG;
}

export interface PLAN_COFIG {
  name: typeof PLANS[keyof typeof PLANS];
  description: string;
  uniqueFeatures: string[];
  isRecommendedPlan: boolean;
  planIcon: string;
  planIconColor: string;
  mtuLimit: number;
  accountIdentifiedLimit: number;
}

export const ADDITIONAL_ACCOUNTS_ADDON_ID =
  'Additional-500-Accounts-USD-Monthly';
export const ADDITIONAL_ACCOUNTS_ADDON_LIMIT = 500;
