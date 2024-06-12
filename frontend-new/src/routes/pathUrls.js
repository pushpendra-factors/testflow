export const PathUrls = {
  Dashboard: '/reports',
  DashboardURL: '/reports/:dashboard_id',
  PreBuildDashboard: '/reports/quick-board',
  PreBuildDashboardReport: '/report/quick-board',
  ComponentsLib: '/components',
  Analyse: '/analyse/:query_type/:query_id',
  Analyse1: '/analyse/:query_type',
  Analyse2: '/analyse',
  Explain: '/explain',
  Alerts: '/alerts',
  Workflows: '/workflows',
  ExplainInsightsV2: '/explainV2/insights',
  ProfilePeople: '/profiles/people',
  ProfilePeopleDetailsURL: '/profiles/people/:id',
  ProfileAccounts: '/',
  ProfileAccountsSegmentsURL: '/accounts/segments/:segment_id',
  ProfileAccountDetailsURL: '/profiles/accounts/:id',
  VisitorIdentificationReport: '/reports/visitor_report',
  PathAnalysis: '/path-analysis',
  PathAnalysisInsights: '/path-analysis/insights',

  // general settings
  SettingsGeneral: '/settings/general',
  SettingsMembers: '/settings/members',
  SettingsPricing: '/settings/pricing',
  SettingsSharing: '/settings/project/sharing',

  // personal settings
  SettingsPersonalUser: '/settings/user',
  SettingsPersonalProjects: '/settings/projects',

  // data management settings
  SettingsIntegration: '/settings/integration',
  SettingsIntegrationURLID: '/settings/integration/:integration_id',
  IntegrationCallbackRedirection: '/callback/integration/:integration_id',
  SettingsTouchpointDefinition: '/settings/touchpoint_definition',
  SettingsCustomDefinition: '/settings/custom_definition',
  SettingsAttribution: '/settings/attribution',
  SettingsAccountScoring: '/settings/account_scoring',
  ConfigurePlans: '/settings/plans',
  ConfigurePlansAdmin: '/settings/plans/admin',

  Settings: '/settings',
  Upgrade: '/upgrade',
  Onboarding: '/onboarding',
  Checklist: '/checklist'
};
