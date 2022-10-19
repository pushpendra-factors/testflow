import {
  QUERY_TYPE_EVENT,
  QUERY_TYPE_FUNNEL,
  QUERY_TYPE_ATTRIBUTION,
  QUERY_TYPE_CAMPAIGN,
  QUERY_TYPE_KPI,
  QUERY_TYPE_PROFILE,
  CHART_TYPE_BARCHART,
  CHART_TYPE_HORIZONTAL_BAR_CHART,
  CHART_TYPE_SPARKLINES,
  CHART_TYPE_TABLE,
  CHART_TYPE_PIVOT_CHART,
  CHART_TYPE_LINECHART
} from '../../../utils/constants';

export const addShadowToHeader = () => {
  const scrollTop =
    window.pageYOffset !== undefined
      ? window.pageYOffset
      : (document.documentElement || document.body.parentNode || document.body)
          .scrollTop;
  if (scrollTop > 0) {
    document.getElementById('app-header').style.filter =
      'drop-shadow(0px 2px 0px rgba(200, 200, 200, 0.25))';
  } else {
    document.getElementById('app-header').style.filter = 'none';
  }
};

export const getChartType = ({
  queryType,
  breakdown,
  chartTypes,
  attributionModels,
  campaignGroupBy
}) => {
  if (queryType === QUERY_TYPE_FUNNEL) {
    const key = breakdown.length ? 'breakdown' : 'no_breakdown';
    return chartTypes[queryType][key] === CHART_TYPE_TABLE
      ? CHART_TYPE_BARCHART
      : chartTypes[queryType][key];
  }

  if (
    queryType === QUERY_TYPE_EVENT ||
    queryType === QUERY_TYPE_PROFILE ||
    queryType === QUERY_TYPE_KPI
  ) {
    const key = breakdown.length ? 'breakdown' : 'no_breakdown';
    if (
      breakdown.length &&
      breakdown.length > 3 &&
      chartTypes[queryType][key] === CHART_TYPE_HORIZONTAL_BAR_CHART
    ) {
      // horizontal bar charts are not supported for more than 3 breakdowns
      return CHART_TYPE_BARCHART;
    }
    if (
      breakdown.length &&
      breakdown.length === 1 &&
      chartTypes[queryType][key] === CHART_TYPE_PIVOT_CHART
    ) {
      // pivot charts are not supported for 1 breakdown
      return CHART_TYPE_BARCHART;
    }
    if (chartTypes[queryType][key] === CHART_TYPE_TABLE) {
      if (breakdown.length) {
        return CHART_TYPE_BARCHART;
      }
      return CHART_TYPE_SPARKLINES;
    }
    return chartTypes[queryType][key];
  }

  if (queryType === QUERY_TYPE_CAMPAIGN) {
    const key = campaignGroupBy.length ? 'breakdown' : 'no_breakdown';
    if (campaignGroupBy.length >= 1) {
      return chartTypes[queryType][key] === CHART_TYPE_TABLE
        ? CHART_TYPE_BARCHART
        : chartTypes[queryType][key];
    }
    return chartTypes[queryType][key] === CHART_TYPE_TABLE
      ? CHART_TYPE_SPARKLINES
      : chartTypes[queryType][key];
  }

  if (queryType === QUERY_TYPE_ATTRIBUTION) {
    const key =
      attributionModels?.length === 1
        ? 'single_touch_point'
        : 'dual_touch_point';
    return chartTypes[queryType][key] === CHART_TYPE_TABLE
      ? CHART_TYPE_BARCHART
      : chartTypes[queryType][key];
  }
  return CHART_TYPE_LINECHART;
};

export const getChartChangedKey = ({
  queryType,
  breakdown,
  campaignGroupBy,
  attributionModels
}) => {
  if (
    queryType === QUERY_TYPE_EVENT ||
    queryType === QUERY_TYPE_FUNNEL ||
    queryType === QUERY_TYPE_PROFILE ||
    queryType === QUERY_TYPE_KPI
  ) {
    return breakdown.length ? 'breakdown' : 'no_breakdown';
  }
  if (queryType === QUERY_TYPE_CAMPAIGN) {
    return campaignGroupBy.length ? 'breakdown' : 'no_breakdown';
  }
  if (queryType === QUERY_TYPE_ATTRIBUTION) {
    return attributionModels.length > 1
      ? 'dual_touch_point'
      : 'single_touch_point';
  }
  return 'no_breakdown';
};
