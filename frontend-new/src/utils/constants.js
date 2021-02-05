export const QUERY_TYPE_FUNNEL = "funnel";
export const QUERY_TYPE_EVENT = "events";
export const QUERY_TYPE_ATTRIBUTION = "attribution";
export const QUERY_TYPE_CAMPAIGN = "channel_v1";
export const QUERY_TYPE_TEMPLATE = "templates";
export const QUERY_TYPE_WEB = "web";
export const NAMED_QUERY = "named_query";

export const ATTRIBUTION_METHODOLOGY = [
  {
    text: "First Touch",
    value: "First_Touch",
  },
  {
    text: "Last Touch",
    value: "Last_Touch",
  },
  {
    text: "First Touch Non-Direct",
    value: "First_Touch_ND",
  },
  {
    text: "Last Touch Non-Direct",
    value: "Last_Touch_ND",
  },
  {
    text: "Linear Touch",
    value: "Linear",
  },
];

export const CHART_TYPE_SPARKLINES = "sparklines";
export const CHART_TYPE_BARCHART = "barchart";
export const CHART_TYPE_LINECHART = "linechart";
export const CHART_TYPE_TABLE = "table";
export const BARCHART_TICK_LENGTH = 20;
export const UNGROUPED_FUNNEL_TICK_LENGTH = 50;

export const EVENT_BREADCRUMB = {
  [QUERY_TYPE_EVENT]: "Events",
  [QUERY_TYPE_FUNNEL]: "Funnel",
  [QUERY_TYPE_ATTRIBUTION]: "Attributions",
  [QUERY_TYPE_CAMPAIGN]: "Campaigns",
};

export const valueMapper = {
  $no_group: "Overall",
};

export const TOTAL_EVENTS_CRITERIA = "total_events";
export const TOTAL_USERS_CRITERIA = "total_users";
export const ACTIVE_USERS_CRITERIA = "active_users";
export const FREQUENCY_CRITERIA = "frequency";
export const TYPE_EVENTS_OCCURRENCE = "events_occurrence";
export const TYPE_UNIQUE_USERS = "unique_users";

export const EACH_USER_TYPE = "each";
export const ANY_USER_TYPE = "any";
export const ALL_USER_TYPE = "all";

export const constantObj = {
  [EACH_USER_TYPE]: "each_given_event",
  [ANY_USER_TYPE]: "any_given_event",
  [ALL_USER_TYPE]: "all_given_event",
};

export const reverse_user_types = {
  each_given_event: EACH_USER_TYPE,
  any_given_event: ANY_USER_TYPE,
  all_given_event: ALL_USER_TYPE,
};

export const REPORT_SECTION = "reports";
export const DASHBOARD_MODAL = "dashboard_modal";
export const DASHBOARD_WIDGET_SECTION = "dashboardWidget";

export const linechart_legend_length = 30;

export const legend_counts = {
  0:2,
  1: 5,
}