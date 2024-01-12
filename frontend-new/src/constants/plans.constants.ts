export const FEATURES = {
  // analyse section
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
  FEATURE_EXPLAIN: 'explain_feature', // explain is a keyword in memsql.
  FEATURE_FILTERS: 'filters',
  FEATURE_SHAREABLE_URL: 'shareable_url',
  FEATURE_CUSTOM_METRICS: 'custom_metrics',
  FEATURE_SMART_PROPERTIES: 'smart_properties',
  FEATURE_CONTENT_GROUPS: 'content_groups',
  FEATURE_DISPLAY_NAMES: 'display_names',
  FEATURE_WEEKLY_INSIGHTS: 'weekly_insights',
  FEATURE_KPI_ALERTS: 'kpi_alerts',
  FEATURE_EVENT_BASED_ALERTS: 'event_based_alerts',
  FEATURE_REPORT_SHARING: 'report_sharing',
  FEATURE_SLACK: 'slack',
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
  FEATURE_CLICKABLE_ELEMENTS: 'clickable_elements',
  FEATURE_WEB_ANALYTICS_DASHBOARD: 'web_analytics_dashboard',
  FEATURE_G2: 'g2',
  FEATURE_RUDDERSTACK: 'rudderstack',
  CONF_CUSTOM_PROPERTIES: 'conf_custom_properties',
  CONF_CUSTOM_EVENTS: 'conf_custom_events'
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
    description: 'Identify accounts and track their activity for free',
    uniqueFeatures: [
      'Website Intent',
      'Account enrichment',
      'B2B Website Analytics',
      'Account Journey Timelines',
      'Custom Filters & Segments',
      'Real-Time Intent Alerts'
    ],
    isRecommendedPlan: false,
    planIcon: 'Userplus',
    planIconColor: '#40A9FF',
    mtuLimit: 5000,
    accountIdentifiedLimit: 100
  },

  [PLANS.PLAN_BASIC]: {
    name: PLANS.PLAN_BASIC,
    description: 'Automate ABM workflows to co-ordinate better with sales',
    uniqueFeatures: [
      'Everything in Free +',
      'Account Identification Exclusions',
      'Lead Enrichment Automation',
      'Outbound Email Automation',
      'LinkedIn Intent Signals',
      'Ad Performance Analytics',
      'Content Analytics',
      'Custom KPIs & Events',
      'Onboarding support',
      'Ads integration. (LinkedIn, Google, Facebook, Bing)',
      'Webhooks (Zapier, Make etc)'
    ],
    isRecommendedPlan: false,
    planIcon: 'User_friends',
    planIconColor: '#73D13D',
    mtuLimit: 10000,
    accountIdentifiedLimit: 350
  },
  [PLANS.PLAN_GROWTH]: {
    name: PLANS.PLAN_GROWTH,
    description:
      'Qualify leads and track engagement with custom scoring models',
    uniqueFeatures: [
      'Everything in Basic +',
      'G2 Intent Signals',
      'Account + Lead Scoring',
      'LinkedIn View Through Attribution',
      'Basic Single Touch Funnel Attribution',
      'Engaged Channels (Coming soon)',
      'Priority CSM',
      'CRM integration (HubSpot, Salesforce)',
      'Marketing platforms integration (G2, Marketo)'
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
      'Attribute success with Multi-Attribution, AI-Explain and more',
    uniqueFeatures: [
      'Everything in Growth +',
      'Multi-touch attribution',
      'Path analysis',
      'AI-fuelled Explain',
      'Dedicated CSM'
    ],
    isRecommendedPlan: false,
    planIcon: 'Buildings',
    planIconColor: '#FF7A45',
    mtuLimit: 100000,
    accountIdentifiedLimit: 10000
  }
};

export interface PLANS_COFIG_INTERFACE {
  [key: (typeof PLANS)[keyof typeof PLANS]]: PLAN_COFIG;
}

export interface PLAN_COFIG {
  name: (typeof PLANS)[keyof typeof PLANS];
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
