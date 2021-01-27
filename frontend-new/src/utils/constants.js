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

export const EVENT_BREADCRUMB = {
  [QUERY_TYPE_EVENT]: "Events",
  [QUERY_TYPE_FUNNEL]: "Funnel",
  [QUERY_TYPE_ATTRIBUTION]: "Attributions",
  [QUERY_TYPE_CAMPAIGN]: "Campaigns",
};

export const valueMapper = {
  $no_group: "Overall",
};
