export const QUERY_TYPE_FUNNEL = 'funnel';
export const QUERY_TYPE_EVENT = 'events';
export const QUERY_TYPE_ATTRIBUTION = 'attribution';
export const QUERY_TYPE_CAMPAIGN = 'channel_v1';
export const QUERY_TYPE_KPI = 'kpi';
export const QUERY_TYPE_TEMPLATE = 'templates';
export const QUERY_TYPE_WEB = 'web';
export const NAMED_QUERY = 'named_query';
export const QUERY_TYPE_PROFILE = 'profiles';

export const ATTRIBUTION_METHODOLOGY = [
  {
    text: 'First Touch',
    value: 'First_Touch',
  },
  {
    text: 'Last Touch',
    value: 'Last_Touch',
  },
  {
    text: 'First Touch Non-Direct',
    value: 'First_Touch_ND',
  },
  {
    text: 'Last Touch Non-Direct',
    value: 'Last_Touch_ND',
  },
  {
    text: 'Linear Touch',
    value: 'Linear',
  },
  {
    text: 'U Shaped',
    value: 'U_Shaped',
  },
  {
    text: 'Time Decay',
    value: 'Time_Decay',
  },
];

export const CHART_TYPE_HORIZONTAL_BAR_CHART = 'horizontalbarchart';
export const CHART_TYPE_STACKED_AREA = 'stackedareachart';
export const CHART_TYPE_STACKED_BAR = 'stackedbarchart';
export const CHART_TYPE_SPARKLINES = 'sparklines';
export const CHART_TYPE_BARCHART = 'barchart';
export const CHART_TYPE_LINECHART = 'linechart';
export const CHART_TYPE_TABLE = 'table';
export const CHART_TYPE_SCATTER_PLOT = 'scatterplotchart';
export const CHART_TYPE_PIVOT_CHART = 'pivotchart';
export const BARCHART_TICK_LENGTH = 20;
export const UNGROUPED_FUNNEL_TICK_LENGTH = 50;

export const EVENT_BREADCRUMB = {
  [QUERY_TYPE_EVENT]: 'Events',
  [QUERY_TYPE_FUNNEL]: 'Funnel',
  [QUERY_TYPE_ATTRIBUTION]: 'Attribution',
  [QUERY_TYPE_CAMPAIGN]: 'Campaigns',
  [QUERY_TYPE_KPI]: 'KPI',
  [QUERY_TYPE_PROFILE]: 'Profiles',
};

export const valueMapper = {
  $no_group: 'Overall',
};

export const TOTAL_EVENTS_CRITERIA = 'total_events';
export const TOTAL_USERS_CRITERIA = 'total_users';
export const ACTIVE_USERS_CRITERIA = 'active_users';
export const FREQUENCY_CRITERIA = 'frequency';
export const TYPE_EVENTS_OCCURRENCE = 'events_occurrence';
export const TYPE_UNIQUE_USERS = 'unique_users';
export const TYPE_ALL_USERS = 'all_users';

export const EACH_USER_TYPE = 'each';
export const ANY_USER_TYPE = 'any';
export const ALL_USER_TYPE = 'all';

export const constantObj = {
  [EACH_USER_TYPE]: 'each_given_event',
  [ANY_USER_TYPE]: 'any_given_event',
  [ALL_USER_TYPE]: 'all_given_event',
};

export const reverse_user_types = {
  each_given_event: EACH_USER_TYPE,
  any_given_event: ANY_USER_TYPE,
  all_given_event: ALL_USER_TYPE,
};

export const REPORT_SECTION = 'reports';
export const DASHBOARD_MODAL = 'dashboard_modal';
export const DASHBOARD_WIDGET_SECTION = 'dashboardWidget';

export const DASHBOARD_WIDGET_BAR_CHART_HEIGHT = 250;
export const DASHBOARD_WIDGET_AREA_CHART_HEIGHT = 225;
export const DASHBOARD_WIDGET_MULTICOLORED_BAR_CHART_HEIGHT = 200;
export const DASHBOARD_WIDGET_BARLINE_CHART_HEIGHT = 225;
export const DASHBOARD_WIDGET_LINE_CHART_HEIGHT = 225;
export const DASHBOARD_WIDGET_UNGROUPED_FUNNEL_CHART_HEIGHT = 250;
export const DASHBOARD_WIDGET_SCATTERPLOT_CHART_HEIGHT = 275;
export const DASHBOARD_WIDGET_ATTRIBUTION_DUAL_TOUCHPOINT_BAR_CHART_HEIGHT = 225;

export const BAR_CHART_XAXIS_TICK_LENGTH = {
  0: 10,
  1: 15,
  2: 5,
};

export const BAR_COUNT = {
  0: 5,
  1: 10,
  2: 3,
};

export const BARLINE_COUNT = {
  0: 3,
  1: 5,
  2: 2,
};

export const FUNNELS_COUNT = {
  0: 3,
  1: 10,
  2: 2,
};

export const legend_counts = {
  0: 3,
  1: 6,
  2: 1,
};

export const charts_legend_length = {
  0: 15,
  1: 20,
  2: 10,
};

export const high_charts_default_spacing = [20, 10, 15, 10];
export const high_charts_barLine_default_spacing = [20, 0, 15, 0];
export const high_charts_scatter_plot_default_spacing = [20, 0, 15, 0];

export const presentationObj = {
  pb: CHART_TYPE_BARCHART,
  pl: CHART_TYPE_LINECHART,
  pt: CHART_TYPE_TABLE,
  pc: CHART_TYPE_SPARKLINES,
  pa: CHART_TYPE_STACKED_AREA,
  ps: CHART_TYPE_STACKED_BAR,
  sp: CHART_TYPE_SCATTER_PLOT,
  hb: CHART_TYPE_HORIZONTAL_BAR_CHART,
  pi: CHART_TYPE_PIVOT_CHART,
};

export const apiChartAnnotations = {
  [CHART_TYPE_BARCHART]: 'pb',
  [CHART_TYPE_LINECHART]: 'pl',
  [CHART_TYPE_TABLE]: 'pt',
  [CHART_TYPE_SPARKLINES]: 'pc',
  [CHART_TYPE_STACKED_AREA]: 'pa',
  [CHART_TYPE_STACKED_BAR]: 'ps',
  [CHART_TYPE_SCATTER_PLOT]: 'sp',
  [CHART_TYPE_HORIZONTAL_BAR_CHART]: 'hb',
  [CHART_TYPE_PIVOT_CHART]: 'pi',
};

export const MAX_ALLOWED_VISIBLE_PROPERTIES = 10;
export const GROUPED_MAX_ALLOWED_VISIBLE_PROPERTIES = 5;

export const DASHBOARD_TYPES = {
  WEB: 'web',
  USER_CREATED: 'user_created',
};

export const MARKETING_TOUCHPOINTS = {
  CAMPAIGN: 'Campaign',
  ADGROUP: 'AdGroup',
  SOURCE: 'Source',
  KEYWORD: 'Keyword',
  MATCHTYPE: 'MatchType',
  LANDING_PAGE: 'LandingPage',
};

export const INITIAL_SESSION_ANALYTICS_SEQ = {
  start: 0,
  end: 0,
};

export const ATTRIBUTION_METRICS = [
  {
    title: 'Impressions',
    header: 'Impressions',
    enabled: true,
  },
  {
    title: 'Clicks',
    header: 'Clicks',
    enabled: true,
  },
  {
    title: 'Spend',
    header: 'Spend',
    enabled: true,
  },
  {
    title: 'CTR (%)',
    header: 'CTR(%)',
    enabled: true,
  },
  {
    title: 'Sessions',
    header: 'Sessions OR Website Visitors',
    enabled: true,
  },
  {
    title: 'Users',
    header: 'Users',
    enabled: true,
  },
  {
    title: 'Average CPC',
    header: 'Average CPC',
    enabled: false,
  },
  {
    title: 'CPM',
    header: 'CPM',
    enabled: false,
  },
  {
    title: 'Click Conversion Rate (%)',
    header: 'ConversionRate(%) OR ClickConversionRate(%)',
    enabled: false,
  },
  {
    title: 'Avg Session Time (in sec)',
    header: 'Average Session Time',
    enabled: false,
  },
  {
    title: 'Page Views',
    header: 'PageViews',
    enabled: false,
  },
  {
    title: 'All Cost/Conv',
    header: 'ALL CPC',
    enabled: true,
    isEventMetric: true,
  },
  {
    title: 'All Conv Rate (%)',
    header: 'ALL CR',
    enabled: false,
    isEventMetric: true,
  },
];

export const KEY_TOUCH_POINT_DIMENSIONS = [
  {
    title: 'Ads Platform',
    header: 'channel_name',
    responseHeader: 'ChannelName',
    enabled: false,
    type: 'key',
    touchPoint: MARKETING_TOUCHPOINTS.CAMPAIGN,
    defaultValue: false,
  },
  {
    title: 'Campaign Name',
    header: 'campaign_name',
    responseHeader: MARKETING_TOUCHPOINTS.CAMPAIGN,
    enabled: true,
    type: 'key',
    touchPoint: MARKETING_TOUCHPOINTS.CAMPAIGN,
    defaultValue: true,
  },
  {
    title: 'Ads Platform',
    header: 'channel_name',
    responseHeader: 'ChannelName',
    enabled: false,
    type: 'key',
    touchPoint: MARKETING_TOUCHPOINTS.ADGROUP,
    defaultValue: false,
  },
  {
    title: 'Campaign Name',
    header: 'campaign_name',
    responseHeader: MARKETING_TOUCHPOINTS.CAMPAIGN,
    enabled: true,
    type: 'key',
    touchPoint: MARKETING_TOUCHPOINTS.ADGROUP,
    defaultValue: true,
  },
  {
    title: 'AdGroup Name',
    header: 'adgroup_name',
    responseHeader: MARKETING_TOUCHPOINTS.ADGROUP,
    enabled: true,
    type: 'key',
    touchPoint: MARKETING_TOUCHPOINTS.ADGROUP,
    defaultValue: true,
  },
  {
    title: 'Ads Platform',
    header: 'channel_name',
    responseHeader: 'ChannelName',
    enabled: false,
    type: 'key',
    touchPoint: MARKETING_TOUCHPOINTS.KEYWORD,
    defaultValue: false,
  },
  {
    title: 'Campaign Name',
    header: 'campaign_name',
    responseHeader: MARKETING_TOUCHPOINTS.CAMPAIGN,
    enabled: true,
    type: 'key',
    touchPoint: MARKETING_TOUCHPOINTS.KEYWORD,
    defaultValue: true,
  },
  {
    title: 'AdGroup Name',
    header: 'adgroup_name',
    responseHeader: MARKETING_TOUCHPOINTS.ADGROUP,
    enabled: true,
    type: 'key',
    touchPoint: MARKETING_TOUCHPOINTS.KEYWORD,
    defaultValue: true,
  },
  {
    title: 'Keyword Match Type',
    header: 'keyword_match_type',
    responseHeader: MARKETING_TOUCHPOINTS.MATCHTYPE,
    enabled: true,
    type: 'key',
    touchPoint: MARKETING_TOUCHPOINTS.KEYWORD,
    defaultValue: true,
  },
  {
    title: 'Keyword',
    header: 'keyword',
    responseHeader: MARKETING_TOUCHPOINTS.KEYWORD,
    enabled: true,
    type: 'key',
    touchPoint: MARKETING_TOUCHPOINTS.KEYWORD,
    defaultValue: true,
  },
  {
    title: 'Landing Page URL',
    header: 'landing_page_url',
    responseHeader: MARKETING_TOUCHPOINTS.LANDING_PAGE,
    enabled: true,
    type: 'key',
    touchPoint: MARKETING_TOUCHPOINTS.LANDING_PAGE,
    defaultValue: true,
  },
];

export const KEY_CONTENT_GROUPS = [
  {
    title: 'Landing Page URL',
    header: 'landing_page_url',
    responseHeader: MARKETING_TOUCHPOINTS.LANDING_PAGE,
    enabled: true,
    type: 'key',
    touchPoint: MARKETING_TOUCHPOINTS.LANDING_PAGE,
    defaultValue: true,
  },
];

export const MARKETING_TOUCHPOINTS_ALIAS = {
  campaign: MARKETING_TOUCHPOINTS.CAMPAIGN,
  ad_group: MARKETING_TOUCHPOINTS.ADGROUP,
};

export const FUNNEL_CHART_MARGIN = {
  top: 20,
  right: 0,
  bottom: 30,
  left: 40,
};

export const LOCAL_STORAGE_ITEMS = {
  DASHBOARD_DURATION: 'dashboard_duration_v1',
};

export const DateBreakdowns = [
  {
    title: 'Hourly Trend',
    key: 'hour',
    disabled: false,
  },
  {
    title: 'Daily Trend',
    key: 'date',
    disabled: false,
  },
  {
    title: 'Weekly Trend',
    key: 'week',
    disabled: false,
  },
  {
    title: 'Monthly Trend',
    key: 'month',
    disabled: false,
  },
  {
    title: 'Quarterly Trend',
    key: 'quarter',
    disabled: false,
  },
];

export const DefaultChartTypes = {
  [QUERY_TYPE_EVENT]: {
    no_breakdown: CHART_TYPE_SPARKLINES,
    breakdown: CHART_TYPE_BARCHART,
  },
  [QUERY_TYPE_CAMPAIGN]: {
    no_breakdown: CHART_TYPE_SPARKLINES,
    breakdown: CHART_TYPE_BARCHART,
  },
  [QUERY_TYPE_KPI]: {
    no_breakdown: CHART_TYPE_SPARKLINES,
    breakdown: CHART_TYPE_BARCHART,
  },
  [QUERY_TYPE_ATTRIBUTION]: {
    single_touch_point: CHART_TYPE_BARCHART,
    dual_touch_point: CHART_TYPE_BARCHART,
  },
  [QUERY_TYPE_FUNNEL]: {
    breakdown: CHART_TYPE_BARCHART,
    no_breakdown: CHART_TYPE_BARCHART,
  },
  [QUERY_TYPE_PROFILE]: {
    no_breakdown: CHART_TYPE_HORIZONTAL_BAR_CHART,
    breakdown: CHART_TYPE_BARCHART,
  },
};

export const QUERY_TYPE_TEXT = {
  [QUERY_TYPE_EVENT]: 'Events',
  [QUERY_TYPE_FUNNEL]: 'Funnels',
  [QUERY_TYPE_CAMPAIGN]: 'Campaigns',
  [QUERY_TYPE_ATTRIBUTION]: 'Attributions',
  [QUERY_TYPE_KPI]: 'KPI',
  [QUERY_TYPE_PROFILE]: 'Profiles',
};

export const FIRST_METRIC_IN_ATTR_RESPOSE = 'Impressions';

export const ARR_JOINER = ';;;';

export const PREDEFINED_DATES = {
  THIS_WEEK: 'this_week',
  LAST_WEEK: 'last_week',
  THIS_MONTH: 'this_month',
  LAST_MONTH: 'last_month',
  TODAY: 'today',
  YESTERDAY: 'yesterday',
};

export const TimeZoneOffsetValues = {
  IST: { offset: '+05:30', city: 'Asia/Kolkata' },
  PT: { offset: '−08:00', city: 'America/Vancouver' },
  CT: { offset: '−06:00', city: 'America/Costa_Rica' },
  ET: { offset: '−05:00', city: 'America/Chicago' },
  GMT: { offset: '+00:00', city: 'UTC' },
  AEST: { offset: '+10:00', city: 'Australia/Sydney' },
};

export const DATE_FORMATS = {
  quarter: 'MMM-YYYY',
  month: 'MMM-YYYY',
  date: 'D-MMM-YYYY',
  day: 'D-MMM-YYYY',
  hour: 'D-MMM-YYYY H [h]',
};

export const ProfileMapper = {
  'Website Visitors': 'web',
  'Hubspot Contacts': 'hubspot',
  'Salesforce Users': 'salesforce',
  'All Opportunities': 'salesforce',
  'All Deals': 'hubspot',
  'All Accounts': 'salesforce',
  'All Companies': 'hubspot',
};

export const ReverseProfileMapper = {
  web: { users: 'Website Visitors' },
  hubspot: {
    users: 'Hubspot Contacts',
    $hubspot_deal: 'All Deals',
    $hubspot_company: 'All Companies',
  },
  salesforce: {
    users: 'Salesforce Users',
    $salesforce_opportunity: 'All Opportunities',
    $salesforce_account: 'All Accounts',
  },
};

export const DISPLAY_PROP = { $none: '(Not Set)' };
export const REV_DISPLAY_PROP = { '(Not Set)': '$none' };
