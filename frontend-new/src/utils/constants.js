import { GROUP_NAME_DOMAINS } from 'Components/GlobalFilter/FilterWrapper/utils';

export const QUERY_TYPE_FUNNEL = 'funnel';
export const QUERY_TYPE_EVENT = 'events';
export const QUERY_TYPE_ATTRIBUTION = 'attribution';
export const QUERY_TYPE_CAMPAIGN = 'channel_v1';
export const QUERY_TYPE_KPI = 'kpi';
export const QUERY_TYPE_TEMPLATE = 'templates';
export const QUERY_TYPE_WEB = 'web';
export const NAMED_QUERY = 'named_query';
export const QUERY_TYPE_PROFILE = 'profiles';
export const SAVED_QUERY = 'saved_query';
export const FONT_FAMILY =
  "'Inter','Work Sans', sans-serif, 'Helvetica Neue', Arial, 'Noto Sans'";
export const QUERY_TYPE_SEGMENT = 'segments';

export const ATTRIBUTION_METHODOLOGY = [
  {
    text: 'First Touch',
    value: 'First_Touch'
  },
  {
    text: 'Last Touch',
    value: 'Last_Touch'
  },
  {
    text: 'First Touch Non-Direct',
    value: 'First_Touch_ND'
  },
  {
    text: 'Last Touch Non-Direct',
    value: 'Last_Touch_ND'
  },
  {
    text: 'Linear Touch',
    value: 'Linear'
  },
  {
    text: 'U Shaped',
    value: 'U_Shaped'
  },
  {
    text: 'W Shaped',
    value: 'W_Shaped'
  },
  {
    text: 'Influence',
    value: 'Influence'
  },
  {
    text: 'Time Decay',
    value: 'Time_Decay'
  },
  {
    text: 'Last Campaign Touch',
    value: 'Last_Campaign_Touch'
  }
];

export const CHART_TYPE_HORIZONTAL_BAR_CHART = 'horizontalbarchart';
export const CHART_TYPE_STACKED_AREA = 'stackedareachart';
export const CHART_TYPE_STACKED_BAR = 'stackedbarchart';
export const CHART_TYPE_SPARKLINES = 'sparklines';
export const CHART_TYPE_BARCHART = 'barchart';
export const CHART_TYPE_LINECHART = 'linechart';
export const CHART_TYPE_TABLE = 'table';
export const CHART_TYPE_SCATTER_PLOT = 'scatterplotchart';
export const CHART_TYPE_METRIC_CHART = 'metricchart';
export const CHART_TYPE_PIVOT_CHART = 'pivotchart';
export const CHART_TYPE_FUNNEL_CHART = 'funnelchart';
export const BARCHART_TICK_LENGTH = 20;
export const UNGROUPED_FUNNEL_TICK_LENGTH = 50;

export const EVENT_BREADCRUMB = {
  [QUERY_TYPE_EVENT]: 'Events',
  [QUERY_TYPE_FUNNEL]: 'Funnel',
  [QUERY_TYPE_ATTRIBUTION]: 'Attribution',
  [QUERY_TYPE_CAMPAIGN]: 'Campaigns',
  [QUERY_TYPE_KPI]: 'KPI',
  [QUERY_TYPE_PROFILE]: 'Profiles'
};

export const valueMapper = {
  $no_group: 'Overall'
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

export const EVENT_QUERY_USER_TYPE = {
  [EACH_USER_TYPE]: 'each_given_event',
  [ANY_USER_TYPE]: 'any_given_event',
  [ALL_USER_TYPE]: 'all_given_event'
};

export const REVERSE_USER_TYPES = {
  each_given_event: EACH_USER_TYPE,
  any_given_event: ANY_USER_TYPE,
  all_given_event: ALL_USER_TYPE
};

export const REPORT_SECTION = 'reports';
export const DASHBOARD_MODAL = 'dashboard_modal';
export const DASHBOARD_WIDGET_SECTION = 'dashboardWidget';

export const DASHBOARD_WIDGET_BAR_CHART_HEIGHT = 250;
export const DASHBOARD_WIDGET_COLUMN_CHART_HEIGHT = 275;
export const DASHBOARD_WIDGET_AREA_CHART_HEIGHT = 225;
export const DASHBOARD_WIDGET_MULTICOLORED_BAR_CHART_HEIGHT = 200;
export const DASHBOARD_WIDGET_BARLINE_CHART_HEIGHT = 225;
export const DASHBOARD_WIDGET_LINE_CHART_HEIGHT = 225;
export const DASHBOARD_WIDGET_UNGROUPED_FUNNEL_CHART_HEIGHT = 250;
export const DASHBOARD_WIDGET_SCATTERPLOT_CHART_HEIGHT = 275;
export const DASHBOARD_WIDGET_ATTRIBUTION_DUAL_TOUCHPOINT_BAR_CHART_HEIGHT = 225;

export const BAR_CHART_XAXIS_TICK_LENGTH = {
  0: 10,
  1: 11,
  2: 5
};

export const BAR_COUNT = {
  0: 5,
  1: 10,
  2: 3
};

export const BARLINE_COUNT = {
  0: 3,
  1: 5,
  2: 2
};

export const FUNNELS_COUNT = {
  0: 3,
  1: 10,
  2: 2
};

export const legend_counts = {
  0: 3,
  1: 6,
  2: 1
};

export const charts_legend_length = {
  0: 15,
  1: 20,
  2: 10
};

export const high_charts_default_spacing = [20, 10, 15, 10];
export const HIGH_CHARTS_BARLINE_DEFAULT_SPACING = [20, 0, 15, 0];
export const HIGH_CHARTS_SCATTER_PLOT_DEFAULT_SPACING = [20, 10, 15, 10];

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
  mc: CHART_TYPE_METRIC_CHART,
  fc: CHART_TYPE_FUNNEL_CHART
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
  [CHART_TYPE_METRIC_CHART]: 'mc',
  [CHART_TYPE_FUNNEL_CHART]: 'fc'
};

export const MAX_ALLOWED_VISIBLE_PROPERTIES = 10;
export const GROUPED_MAX_ALLOWED_VISIBLE_PROPERTIES = 8;

export const DASHBOARD_TYPES = {
  WEB: 'web',
  USER_CREATED: 'user_created'
};

export const MARKETING_TOUCHPOINTS = {
  CAMPAIGN: 'Campaign',
  ADGROUP: 'AdGroup',
  SOURCE: 'Source',
  KEYWORD: 'Keyword',
  MATCHTYPE: 'MatchType',
  LANDING_PAGE: 'LandingPage'
};

export const INITIAL_SESSION_ANALYTICS_SEQ = {
  start: 0,
  end: 0
};

export const ATTRIBUTION_METRICS = [
  {
    title: 'Impressions',
    header: 'Impressions',
    enabled: true,
    valueType: 'numerical'
  },
  {
    title: 'Clicks',
    header: 'Clicks',
    enabled: true,
    valueType: 'numerical'
  },
  {
    title: 'Spend',
    header: 'Spend',
    enabled: true,
    valueType: 'numerical'
  },
  {
    title: 'CTR (%)',
    header: 'CTR(%)',
    enabled: true,
    valueType: 'percentage'
  },
  {
    title: 'Average CPC',
    header: 'Average CPC',
    enabled: false,
    valueType: 'numerical'
  },
  {
    title: 'CPM',
    header: 'CPM',
    enabled: false,
    valueType: 'numerical'
  },
  {
    title: 'Click Conversion Rate (%)',
    header: 'ConversionRate(%) OR ClickConversionRate(%)',
    enabled: false,
    valueType: 'percentage'
  },
  {
    title: 'Cost Per Conversion',
    header: 'Cost Per Conversion',
    enabled: true,
    isEventMetric: true,
    valueType: 'numerical'
  },
  {
    title: 'Conversion Value',
    header: 'CV',
    enabled: true,
    isEventMetric: true
  },
  {
    title: 'Return on Cost',
    header: 'ROC',
    enabled: true,
    isEventMetric: true
  }
  // {
  //   title: 'All Conv Rate (%)',
  //   header: 'ALL CR',
  //   enabled: false,
  //   isEventMetric: true
  // }
];

export const KEY_TOUCH_POINT_DIMENSIONS = [
  {
    title: 'Ads Platform',
    header: 'channel_name',
    responseHeader: 'ChannelName',
    enabled: false,
    type: 'key',
    touchPoint: MARKETING_TOUCHPOINTS.CAMPAIGN,
    defaultValue: false
  },
  {
    title: 'Campaign Name',
    header: 'campaign_name',
    responseHeader: MARKETING_TOUCHPOINTS.CAMPAIGN,
    enabled: true,
    type: 'key',
    touchPoint: MARKETING_TOUCHPOINTS.CAMPAIGN,
    defaultValue: true
  },
  {
    title: 'Ads Platform',
    header: 'channel_name',
    responseHeader: 'ChannelName',
    enabled: false,
    type: 'key',
    touchPoint: MARKETING_TOUCHPOINTS.ADGROUP,
    defaultValue: false
  },
  {
    title: 'Campaign Name',
    header: 'campaign_name',
    responseHeader: MARKETING_TOUCHPOINTS.CAMPAIGN,
    enabled: true,
    type: 'key',
    touchPoint: MARKETING_TOUCHPOINTS.ADGROUP,
    defaultValue: true
  },
  {
    title: 'AdGroup Name',
    header: 'adgroup_name',
    responseHeader: MARKETING_TOUCHPOINTS.ADGROUP,
    enabled: true,
    type: 'key',
    touchPoint: MARKETING_TOUCHPOINTS.ADGROUP,
    defaultValue: true
  },
  {
    title: 'Ads Platform',
    header: 'channel_name',
    responseHeader: 'ChannelName',
    enabled: false,
    type: 'key',
    touchPoint: MARKETING_TOUCHPOINTS.KEYWORD,
    defaultValue: false
  },
  {
    title: 'Campaign Name',
    header: 'campaign_name',
    responseHeader: MARKETING_TOUCHPOINTS.CAMPAIGN,
    enabled: true,
    type: 'key',
    touchPoint: MARKETING_TOUCHPOINTS.KEYWORD,
    defaultValue: true
  },
  {
    title: 'AdGroup Name',
    header: 'adgroup_name',
    responseHeader: MARKETING_TOUCHPOINTS.ADGROUP,
    enabled: true,
    type: 'key',
    touchPoint: MARKETING_TOUCHPOINTS.KEYWORD,
    defaultValue: true
  },
  {
    title: 'Keyword Match Type',
    header: 'keyword_match_type',
    responseHeader: MARKETING_TOUCHPOINTS.MATCHTYPE,
    enabled: true,
    type: 'key',
    touchPoint: MARKETING_TOUCHPOINTS.KEYWORD,
    defaultValue: true
  },
  {
    title: 'Keyword',
    header: 'keyword',
    responseHeader: MARKETING_TOUCHPOINTS.KEYWORD,
    enabled: true,
    type: 'key',
    touchPoint: MARKETING_TOUCHPOINTS.KEYWORD,
    defaultValue: true
  },
  {
    title: 'Landing Page URL',
    header: 'landing_page_url',
    responseHeader: MARKETING_TOUCHPOINTS.LANDING_PAGE,
    enabled: true,
    type: 'key',
    touchPoint: MARKETING_TOUCHPOINTS.LANDING_PAGE,
    defaultValue: true
  }
];

export const KEY_CONTENT_GROUPS = [
  {
    title: 'Landing Page URL',
    header: 'landing_page_url',
    responseHeader: MARKETING_TOUCHPOINTS.LANDING_PAGE,
    enabled: true,
    type: 'key',
    touchPoint: MARKETING_TOUCHPOINTS.LANDING_PAGE,
    defaultValue: true
  }
];

export const MARKETING_TOUCHPOINTS_ALIAS = {
  campaign: MARKETING_TOUCHPOINTS.CAMPAIGN,
  ad_group: MARKETING_TOUCHPOINTS.ADGROUP
};

export const FUNNEL_CHART_MARGIN = {
  top: 20,
  right: 0,
  bottom: 30,
  left: 40
};

export const DateBreakdowns = [
  {
    title: 'Hourly Trend',
    key: 'hour',
    disabled: false
  },
  {
    title: 'Daily Trend',
    key: 'date',
    disabled: false
  },
  {
    title: 'Weekly Trend',
    key: 'week',
    disabled: false
  },
  {
    title: 'Monthly Trend',
    key: 'month',
    disabled: false
  },
  {
    title: 'Quarterly Trend',
    key: 'quarter',
    disabled: false
  }
];

export const DefaultChartTypes = {
  [QUERY_TYPE_EVENT]: {
    no_breakdown: CHART_TYPE_SPARKLINES,
    breakdown: CHART_TYPE_BARCHART
  },
  [QUERY_TYPE_CAMPAIGN]: {
    no_breakdown: CHART_TYPE_SPARKLINES,
    breakdown: CHART_TYPE_BARCHART
  },
  [QUERY_TYPE_KPI]: {
    no_breakdown: CHART_TYPE_SPARKLINES,
    breakdown: CHART_TYPE_BARCHART
  },
  [QUERY_TYPE_ATTRIBUTION]: {
    single_touch_point: CHART_TYPE_BARCHART,
    dual_touch_point: CHART_TYPE_BARCHART
  },
  [QUERY_TYPE_FUNNEL]: {
    breakdown: CHART_TYPE_FUNNEL_CHART,
    no_breakdown: CHART_TYPE_FUNNEL_CHART
  },
  [QUERY_TYPE_PROFILE]: {
    no_breakdown: CHART_TYPE_HORIZONTAL_BAR_CHART,
    breakdown: CHART_TYPE_BARCHART
  }
};

export const QUERY_TYPE_TEXT = {
  [QUERY_TYPE_EVENT]: 'Events',
  [QUERY_TYPE_FUNNEL]: 'Funnels',
  [QUERY_TYPE_CAMPAIGN]: 'Campaigns',
  [QUERY_TYPE_ATTRIBUTION]: 'Attributions',
  [QUERY_TYPE_KPI]: 'KPI',
  [QUERY_TYPE_PROFILE]: 'Profiles'
};

export const FIRST_METRIC_IN_ATTR_RESPONSE = 'Impressions';

export const ARR_JOINER = ';;;';

export const PREDEFINED_DATES = {
  THIS_WEEK: 'this_week',
  LAST_WEEK: 'last_week',
  THIS_MONTH: 'this_month',
  LAST_MONTH: 'last_month',
  TODAY: 'today',
  YESTERDAY: 'yesterday'
};

export const getTimeZoneNameFromCity = (name) =>
  TimeZoneOffsetValueArr.find((item) => item?.city === name);

export const DATE_FORMATS = {
  quarter: 'Q-YYYY',
  month: 'MMM-YYYY',
  date: 'D-MMM-YYYY',
  day: 'D-MMM-YYYY',
  hour: 'D-MMM-YYYY H [h]'
};

export const ProfileMapper = {
  'Website Visitors': 'web',
  'Hubspot Contacts': 'hubspot',
  'Salesforce Users': 'salesforce',
  'All Opportunities': 'salesforce',
  'All Deals': 'hubspot',
  'All Accounts': 'salesforce',
  'All Companies': 'hubspot',
  'Marketo Person': 'marketo',
  'LeadSquared Person': 'leadsquared',
  'All Domains': '6signal',
  'All Linkedin Engagements': 'linkedin_company',
  'All G2 Engagements': 'g2'
};

export const ReverseProfileMapper = {
  web: { users: 'Website Visitors' },
  hubspot: {
    users: 'Hubspot Contacts',
    $hubspot_deal: 'All Deals',
    $hubspot_company: 'All Companies'
  },
  salesforce: {
    users: 'Salesforce Users',
    $salesforce_opportunity: 'All Opportunities',
    $salesforce_account: 'All Accounts'
  },
  marketo: { users: 'Marketo Person' },
  leadsquared: { users: 'LeadSquared Person' },
  '6signal': { $6signal: 'All Domains' },
  linkedin_company: { $linkedin_company: 'All Linkedin Engagements' },
  g2: { $g2: 'All G2 Engagements' }
};

export const profileOptions = {
  users: [
    ['Website Visitors'],
    ['Hubspot Contacts'],
    ['Salesforce Users'],
    ['Marketo Person'],
    ['LeadSquared Person']
  ],
  $salesforce_opportunity: [['All Opportunities']],
  $hubspot_deal: [['All Deals']],
  $salesforce_account: [['All Accounts']],
  $hubspot_company: [['All Companies']],
  $6signal: [['All Domains']],
  $linkedin_company: [['All Linkedin Engagements']],
  $g2: [['All G2 Engagements']]
};

export const DISPLAY_PROP = { $none: '(Not Set)' };
export const REV_DISPLAY_PROP = { '(Not Set)': '$none' };

export const METRIC_TYPES = {
  dateType: 'date_type',
  percentType: 'percentage_type'
};

export const QUERY_OPTIONS_DEFAULT_VALUE = {
  group_analysis: GROUP_NAME_DOMAINS,
  groupBy: [
    {
      prop_category: '', // user / event
      property: '', // user/eventproperty
      prop_type: '', // categorical  /numberical
      eventValue: '', // event name (funnel only)
      eventName: '', // eventName $present for global user breakdown
      eventIndex: 0
    }
  ],
  globalFilters: [],
  event_analysis_seq: '',
  session_analytics_seq: {},
  date_range: {},
  events_condition: 'any_given_event'
};

export const DealOrOppurtunity = 'Deal / Opportunity';
export const CompanyOrAccount = 'Company / Account';

export const AttributionGroupOptions = [DealOrOppurtunity, CompanyOrAccount];

export const FunnelEventsConditionMap = {
  any_given_event: 'This Order',
  funnel_any_given_event: 'Any Order'
};

export const RevFunnelEventsConditionMap = {
  'This Order': 'any_given_event',
  'Any Order': 'funnel_any_given_event'
};

export const OPERATORS = {
  equalTo: 'equals',
  notEqualTo: 'not equals',
  contain: 'contains',
  doesNotContain: 'does not contain',
  lesserThan: '<',
  lesserThanOrEqual: '<=',
  greaterThan: '>',
  greaterThanOrEqual: '>=',
  between: 'between',
  notBetween: 'not between',
  inThePrevious: 'in the previous',
  notInThePrevious: 'not in the previous',
  inTheLast: 'in the last',
  notInTheLast: 'not in the last',
  inTheCurrent: 'in the current',
  notInTheCurrent: 'not in the current',
  before: 'before',
  since: 'since',
  isKnown: 'is known',
  isUnknown: 'is unknown',
  inList: 'is in a list',
  notInList: 'not in a list'
};

//always make sure "city": "UTC" and "city": "Europe/Berlin" is defined for backward compatibility.

export const TimeZoneOffsetValueArr = [
  {
    "abbr": "UTC",
    "name": "UTC",
    "offset": "+00:00",
    "text": "(UTC) Coordinated Universal Time",
    "city": "UTC"
  },
  {
    "abbr": "DST",
    "name": "Dateline Standard Time",
    "offset": "-12:00",
    "text": "(UTC-12:00) International Date Line West",
    "city": "Etc/GMT+12"
  },
  {
    "abbr": "U",
    "name": "UTC-11",
    "offset": "-11:00",
    "text": "(UTC-11:00) Coordinated Universal Time-11",
    "city": "Etc/GMT+11"
  },
  {
    "abbr": "HST",
    "name": "Hawaiian Standard Time",
    "offset": "-10:00",
    "text": "(UTC-10:00) Hawaii",
    "city": "Pacific/Tahiti"
  },
  {
    "abbr": "AKDT",
    "name": "Alaskan Standard Time",
    "offset": "-09:00",
    "text": "(UTC-09:00) Alaska",
    "city": "America/Juneau"
  },
  {
    "abbr": "PST",
    "name": "Pacific Standard Time",
    "offset": "-08:00",
    "text": "(UTC-08:00) Pacific Standard Time (US & Canada)",
    "city": "America/Vancouver"
  },
  {
    "abbr": "UMST",
    "name": "US Mountain Standard Time",
    "offset": "-07:00",
    "text": "(UTC-07:00) Arizona",
    "city": "America/Creston"
  },
  {
    "abbr": "CST",
    "name": "Central Standard Time",
    "offset": "-06:00",
    "text": "(UTC-06:00) Central Time (US & Canada)",
    "city": "America/Chicago"
  },
  {
    "abbr": "CAST",
    "name": "Central America Standard Time",
    "offset": "-06:00",
    "text": "(UTC-06:00) Central America",
    "city": "America/Costa_Rica"
  },
  {
    "abbr": "SPST",
    "name": "SA Pacific Standard Time",
    "offset": "-05:00",
    "text": "(UTC-05:00) Bogota, Lima, Quito",
    "city": "America/Bogota"
  },
  {
    "abbr": "EST",
    "name": "Eastern Standard Time",
    "offset": "-05:00",
    "text": "(UTC-05:00) Eastern Time (US & Canada)",
    "city": "America/New_York"
  },
  {
    "abbr": "VST",
    "name": "Venezuela Standard Time",
    "offset": "-04:00",
    "text": "(UTC-04:00) Caracas",
    "city": "America/Caracas"
  },
  {
    "abbr": "PYT",
    "name": "Paraguay Standard Time",
    "offset": "-04:00",
    "text": "(UTC-04:00) Asuncion",
    "city": "America/Asuncion"
  },
  {
    "abbr": "CBST",
    "name": "Central Brazilian Standard Time",
    "offset": "-04:00",
    "text": "(UTC-04:00) Cuiaba",
    "city": "America/Cuiaba"
  },
  {
    "abbr": "SWST",
    "name": "SA Western Standard Time",
    "offset": "-04:00",
    "text": "(UTC-04:00) Georgetown, La Paz, Manaus, San Juan",
    "city": "America/La_Paz"
  },
  {
    "abbr": "PSST",
    "name": "Pacific SA Standard Time",
    "offset": "-04:00",
    "text": "(UTC-04:00) Santiago",
    "city": "America/Santiago"
  },
  {
    "abbr": "ESAST",
    "name": "E. South America Standard Time",
    "offset": "-03:00",
    "text": "(UTC-03:00) Brasilia",
    "city": "America/Sao_Paulo"
  },
  {
    "abbr": "AST",
    "name": "Argentina Standard Time",
    "offset": "-03:00",
    "text": "(UTC-03:00) Buenos Aires",
    "city": "America/Argentina/Buenos_Aires"
  },
  {
    "abbr": "SEST",
    "name": "SA Eastern Standard Time",
    "offset": "-03:00",
    "text": "(UTC-03:00) Cayenne, Fortaleza",
    "city": "America/Cayenne"
  },
  {
    "abbr": "MST",
    "name": "Montevideo Standard Time",
    "offset": "-03:00",
    "text": "(UTC-03:00) Montevideo",
    "city": "America/Montevideo"
  },
  {
    "abbr": "BST",
    "name": "Bahia Standard Time",
    "offset": "-03:00",
    "text": "(UTC-03:00) Salvador",
    "city": "America/Bahia"
  },
  {
    "abbr": "U",
    "name": "UTC-02",
    "offset": "-02:00",
    "text": "(UTC-02:00) Coordinated Universal Time-02",
    "city": "Etc/GMT+2"
  },
  {
    "abbr": "CVST",
    "name": "Cape Verde Standard Time",
    "offset": "-01:00",
    "text": "(UTC-01:00) Cape Verde Is.",
    "city": "Atlantic/Cape_Verde"
  },
  {
    "abbr": "UTC",
    "name": "UTC",
    "offset": "+00:00",
    "text": "(UTC) Coordinated Universal Time",
    "city": "Etc/GMT"
  },
  {
    "abbr": "GMT",
    "name": "GMT Standard Time",
    "offset": "+00:00",
    "text": "(UTC) Edinburgh, London",
    "city": "Europe/London"
  },
  {
    "abbr": "GST",
    "name": "Greenwich Standard Time",
    "offset": "+00:00",
    "text": "(UTC) Monrovia, Reykjavik",
    "city": "Africa/Monrovia"
  },
  {
    "abbr": "CEST",
    "name": "Central European Standard Time",
    "offset": "+01:00",
    "text" : "(UTC+01:00) Central European Standard Time",
    "city": "Europe/Budapest"
  },
  {
    "abbr": "CEST",
    "name": "Central European Standard Time",
    "offset": "+01:00",
    "text" : "(UTC+01:00) Central European Standard Time",
    "city": "Europe/Berlin"
  },
  {
    "abbr": "WCAST",
    "name": "W. Central Africa Standard Time",
    "offset": "+01:00",
    "text": "(UTC+01:00) West Central Africa",
    "city": "Africa/Algiers"
  },
  {
    "abbr": "EET",
    "name": "Eastern European Time",
    "offset": "+02:00",
    "text": "(UTC+02:00) Eastern European Time",
    "city": "Europe/Athens"
  },
  {
    "abbr": "NST",
    "name": "Namibia Standard Time",
    "offset": "+02:00",
    "text": "(UTC+02:00) Windhoek",
    "city": "Africa/Windhoek"
  },
  {
    "abbr": "EST",
    "name": "Egypt Standard Time",
    "offset": "+02:00",
    "text": "(UTC+02:00) Cairo",
    "city": "Africa/Cairo"
  },
  {
    "abbr": "SAST",
    "name": "South Africa Standard Time",
    "offset": "+02:00",
    "text": "(UTC+02:00) Harare, Pretoria",
    "city": "Africa/Harare"
  },
  {
    "abbr": "LST",
    "name": "Libya Standard Time",
    "offset": "+02:00",
    "text": "(UTC+02:00) Tripoli",
    "city": "Africa/Tripoli"
  },
  {
    "abbr": "KST",
    "name": "Kaliningrad Standard Time",
    "offset": "+02:00",
    "text": "(UTC+02:00) Kaliningrad",
    "city": "Europe/Kaliningrad"
  },
  {
    "abbr": "TDT",
    "name": "Turkey Standard Time",
    "offset": "+03:00",
    "text": "(UTC+03:00) Istanbul",
    "city": "Europe/Istanbul"
  },
  {
    "abbr": "AST",
    "name": "Arabia Standard Time",
    "offset": "+03:00",
    "text": "(UTC+03:00) Kuwait, Riyadh",
    "city": "Asia/Kuwait"
  },
  {
    "abbr": "EAST",
    "name": "E. Africa Standard Time",
    "offset": "+03:00",
    "text": "(UTC+03:00) Nairobi",
    "city": "Africa/Nairobi"
  },
  {
    "abbr": "MSK",
    "name": "Moscow Standard Time",
    "offset": "+03:00",
    "text": "(UTC+03:00) Moscow, St. Petersburg, Volgograd, Minsk",
    "city": "Europe/Moscow"
  },
  {
    "abbr": "SAMT",
    "name": "Samara Time",
    "offset": "+04:00",
    "text": "(UTC+04:00) Samara, Ulyanovsk, Saratov",
    "city": "Europe/Samara"
  },
  {
    "abbr": "GST",
    "name": "Gulf Standard Time",
    "offset": "+04:00",
    "text": "(UTC+04:00) Abu Dhabi, Muscat",
    "city": "Asia/Dubai"
  },
  {
    "abbr": "MST",
    "name": "Mauritius Standard Time",
    "offset": "+04:00",
    "text": "(UTC+04:00) Port Louis",
    "city": "Indian/Mauritius"
  },
  {
    "abbr": "GET",
    "name": "Georgian Standard Time",
    "offset": "+04:00",
    "text": "(UTC+04:00) Tbilisi",
    "city": "Asia/Tbilisi"
  },
  {
    "abbr": "CST",
    "name": "Caucasus Standard Time",
    "offset": "+04:00",
    "text": "(UTC+04:00) Yerevan",
    "city": "Asia/Yerevan"
  },
  {
    "abbr": "AST",
    "name": "Afghanistan Standard Time",
    "offset": "+04:30",
    "text": "(UTC+04:30) Kabul",
    "city": "Asia/Kabul"
  },
  {
    "abbr": "WAST",
    "name": "West Asia Standard Time",
    "offset": "+05:00",
    "text": "(UTC+05:00) Ashgabat, Tashkent",
    "city": "Asia/Ashgabat"
  },
  {
    "abbr": "YEKT",
    "name": "Yekaterinburg Time",
    "offset": "+05:00",
    "text": "(UTC+05:00) Yekaterinburg",
    "city": "Asia/Yekaterinburg"
  },
  {
    "abbr": "PKT",
    "name": "Pakistan Standard Time",
    "offset": "+05:00",
    "text": "(UTC+05:00) Islamabad, Karachi",
    "city": "Asia/Karachi"
  },
  {
    "abbr": "IST",
    "name": "India Standard Time",
    "offset": "+05:30",
    "text": "(UTC+05:30) Chennai, Kolkata, Mumbai, New Delhi",
    "city": "Asia/Kolkata"
  },
  {
    "abbr": "SLST",
    "name": "Sri Lanka Standard Time",
    "offset": "+05:30",
    "text": "(UTC+05:30) Sri Jayawardenepura",
    "city": "Asia/Colombo"
  },
  {
    "abbr": "NST",
    "name": "Nepal Standard Time",
    "offset": "+05:45",
    "text": "(UTC+05:45) Kathmandu",
    "city": "Asia/Kathmandu"
  },
  {
    "abbr": "CAST",
    "name": "Central Asia Standard Time",
    "offset": "+06:00",
    "text": "(UTC+06:00) Nur-Sultan (Astana)",
    "city": "Asia/Almaty"
  },
  {
    "abbr": "OMST",
    "name": "Omsk Standard Time",
    "offset": "+06:00",
    "text": "(UTC+06:00) Omsk",
    "city": "Asia/Omsk"
  },
  {
    "abbr": "BST",
    "name": "Bangladesh Standard Time",
    "offset": "+06:00",
    "text": "(UTC+06:00) Dhaka",
    "city": "Asia/Dhaka"
  },
  {
    "abbr": "MST",
    "name": "Myanmar Standard Time",
    "offset": "+06:30",
    "text": "(UTC+06:30) Yangon (Rangoon)",
    "city": "Asia/Rangoon"
  },
  {
    "abbr": "WIB",
    "name": "Western Indonesia Time",
    "offset": "+07:00",
    "text": "(UTC+07:00) Western Indonesia Time",
    "city": "Asia/Jakarta"
  },
  {
    "abbr": "SAST",
    "name": "SE Asia Standard Time",
    "offset": "+07:00",
    "text": "(UTC+07:00) Bangkok, Hanoi, Jakarta",
    "city": "Asia/Bangkok"
  },
  {
    "abbr": "KRAT",
    "name": "Krasnoyarsk Time",
    "offset": "+07:00",
    "text": "(UTC+07:00) Krasnoyarsk, Novosibirsk",
    "city": "Asia/Krasnoyarsk"
  },
  {
    "abbr": "CST",
    "name": "China Standard Time",
    "offset": "+08:00",
    "text": "(UTC+08:00) Beijing, Chongqing, Hong Kong, Urumqi",
    "city": "Asia/Hong_Kong"
  },
  {
    "abbr": "SGT",
    "name": "Singapore Standard Time",
    "offset": "+08:00",
    "text": "(UTC+08:00) Kuala Lumpur, Singapore",
    "city": "Asia/Singapore"
  },
  {
    "abbr": "WAST",
    "name": "W. Australia Standard Time",
    "offset": "+08:00",
    "text": "(UTC+08:00) Perth",
    "city": "Australia/Perth"
  },
  {
    "abbr": "TST",
    "name": "Taipei Standard Time",
    "offset": "+08:00",
    "text": "(UTC+08:00) Taipei",
    "city": "Asia/Taipei"
  },
  {
    "abbr": "UST",
    "name": "Ulaanbaatar Standard Time",
    "offset": "+08:00",
    "text": "(UTC+08:00) Ulaanbaatar",
    "city": "Asia/Ulaanbaatar"
  },
  {
    "abbr": "IRKT",
    "name": "Irkutsk Time",
    "offset": "+08:00",
    "text": "(UTC+08:00) Irkutsk",
    "city": "Asia/Irkutsk"
  },
  {
    "abbr": "JST",
    "name": "Japan Standard Time",
    "offset": "+09:00",
    "text": "(UTC+09:00) Osaka, Sapporo, Tokyo",
    "city": "Asia/Tokyo"
  },
  {
    "abbr": "KST",
    "name": "Korea Standard Time",
    "offset": "+09:00",
    "text": "(UTC+09:00) Seoul",
    "city": "Asia/Seoul"
  },
  {
    "abbr": "YAKT",
    "name": "Yakutsk Time",
    "offset": "+09:00",
    "text": "(UTC+09:00) Yakutsk",
    "city": "Asia/Yakutsk"
  },
  {
    "abbr": "CAST",
    "name": "Cen. Australia Standard Time",
    "offset": "+09:30",
    "text": "(UTC+09:30) Adelaide",
    "city": "Australia/Adelaide"
  },
  {
    "abbr": "ACST",
    "name": "AUS Central Standard Time",
    "offset": "+09:30",
    "text": "(UTC+09:30) Darwin",
    "city": "Australia/Darwin"
  },
  {
    "abbr": "EAST",
    "name": "E. Australia Standard Time",
    "offset": "+10:00",
    "text": "(UTC+10:00) Brisbane",
    "city": "Australia/Brisbane"
  },
  {
    "abbr": "AEST",
    "name": "AUS Eastern Standard Time",
    "offset": "+10:00",
    "text": "(UTC+10:00) Canberra, Melbourne, Sydney",
    "city": "Australia/Sydney"
  },
  {
    "abbr": "WPST",
    "name": "West Pacific Standard Time",
    "offset": "+10:00",
    "text": "(UTC+10:00) Guam, Port Moresby",
    "city": "Pacific/Guam"
  },
  {
    "abbr": "TST",
    "name": "Tasmania Standard Time",
    "offset": "+10:00",
    "text": "(UTC+10:00) Hobart",
    "city": "Australia/Hobart"
  },
  {
    "abbr": "VLAT",
    "name": "Vladivostok Time",
    "offset": "+10:00",
    "text": "(UTC+10:00) Vladivostok",
    "city": "Asia/Vladivostok"
  },
  {
    "abbr": "CPST",
    "name": "Central Pacific Standard Time",
    "offset": "+11:00",
    "text": "(UTC+11:00) Solomon Is., New Caledonia",
    "city": "Pacific/Guadalcanal"
  },
  {
    "abbr": "MAGT",
    "name": "Magadan Time",
    "offset": "+11:00",
    "text": "(UTC+11:00) Magadan, Srednekolymsk",
    "city": "Asia/Magadan"
  },
  {
    "abbr": "NZST",
    "name": "New Zealand Standard Time",
    "offset": "+12:00",
    "text": "(UTC+12:00) Auckland, Wellington",
    "city": "Pacific/Auckland"
  },
  {
    "abbr": "U",
    "name": "UTC+12",
    "offset": "+12:00",
    "text": "(UTC+12:00) Coordinated Universal Time+12",
    "city": "Etc/GMT-12"
  },
  {
    "abbr": "FST",
    "name": "Fiji Standard Time",
    "offset": "+12:00",
    "text": "(UTC+12:00) Fiji",
    "city": "Pacific/Fiji"
  },
  {
    "abbr": "TST",
    "name": "Tonga Standard Time",
    "offset": "+13:00",
    "text": "(UTC+13:00) Nuku'alofa",
    "city": "Pacific/Tongatapu"
  },
  {
    "abbr": "SST",
    "name": "Samoa Standard Time",
    "offset": "+13:00",
    "text": "(UTC+13:00) Samoa",
    "city": "Pacific/Apia"
  }
];

export const customerSupportLink =
  'https://factors.schedulehero.io/campaign/global-round-robin-ssos';
