import {
  QUERY_TYPE_CAMPAIGN,
  QUERY_TYPE_EVENT,
  CHART_TYPE_STACKED_AREA,
  CHART_TYPE_LINECHART,
  CHART_TYPE_STACKED_BAR,
  CHART_TYPE_SPARKLINES,
  QUERY_TYPE_ATTRIBUTION,
  CHART_TYPE_BARCHART,
  CHART_TYPE_SCATTER_PLOT,
  QUERY_TYPE_FUNNEL,
  CHART_TYPE_HORIZONTAL_BAR_CHART,
  QUERY_TYPE_KPI,
  QUERY_TYPE_PROFILE,
  CHART_TYPE_PIVOT_CHART
} from './constants';

export const isPivotSupported = ({ queryType }) => {
  return (
    queryType === QUERY_TYPE_KPI ||
    queryType === QUERY_TYPE_EVENT ||
    queryType === QUERY_TYPE_PROFILE
  );
};

export const getChartTypeMenuItems = (queryType, breakdownLength, events) => {
  let menuItems = [];
  if (queryType === QUERY_TYPE_EVENT || queryType === QUERY_TYPE_CAMPAIGN) {
    if (breakdownLength) {
      menuItems = [
        {
          key: CHART_TYPE_BARCHART,
          name: 'Columns'
        },
        {
          key: CHART_TYPE_LINECHART,
          name: 'Line Chart'
        },
        {
          key: CHART_TYPE_STACKED_AREA,
          name: 'Stacked Area'
        },
        {
          key: CHART_TYPE_STACKED_BAR,
          name: 'Stacked Column'
        }
      ];
      if (
        queryType === QUERY_TYPE_EVENT &&
        events.length === 1 &&
        breakdownLength <= 3
      ) {
        // this chart type is only supported when there is atmost one event and there is atleast 1 breakdown and atmost 3 breakdowns
        menuItems.push({
          key: CHART_TYPE_HORIZONTAL_BAR_CHART,
          name: 'Bars'
        });
      }
      if (queryType === QUERY_TYPE_EVENT && breakdownLength > 1) {
        menuItems.push({
          key: CHART_TYPE_PIVOT_CHART,
          name: 'Pivot Chart'
        });
      }
    } else {
      menuItems = [
        {
          key: CHART_TYPE_SPARKLINES,
          name: 'Sparkline'
        },
        {
          key: CHART_TYPE_LINECHART,
          name: 'Line Chart'
        }
      ];
    }
  }
  if (queryType === QUERY_TYPE_ATTRIBUTION) {
    menuItems = [
      {
        key: CHART_TYPE_BARCHART,
        name: 'Barchart'
      },
      {
        key: CHART_TYPE_SCATTER_PLOT,
        name: 'Scatter Plot'
      }
    ];
  }
  if (queryType === QUERY_TYPE_FUNNEL && breakdownLength) {
    menuItems = [
      {
        key: CHART_TYPE_BARCHART,
        name: 'Barchart'
      },
      {
        key: CHART_TYPE_SCATTER_PLOT,
        name: 'Scatter Plot'
      }
    ];
  }
  if (queryType === QUERY_TYPE_KPI && !breakdownLength) {
    menuItems = [
      {
        key: CHART_TYPE_SPARKLINES,
        name: 'Sparkline'
      },
      {
        key: CHART_TYPE_LINECHART,
        name: 'Line Chart'
      }
    ];
  }

  if (queryType === QUERY_TYPE_KPI && breakdownLength) {
    menuItems = [
      {
        key: CHART_TYPE_BARCHART,
        name: 'Columns'
      },
      {
        key: CHART_TYPE_LINECHART,
        name: 'Line Chart'
      },
      {
        key: CHART_TYPE_STACKED_AREA,
        name: 'Stacked Area'
      },
      {
        key: CHART_TYPE_STACKED_BAR,
        name: 'Stacked Column'
      }
    ];
    if (breakdownLength <= 3) {
      menuItems.push({
        key: CHART_TYPE_HORIZONTAL_BAR_CHART,
        name: 'Bars'
      });
    }
    if (breakdownLength > 1) {
      menuItems.push({
        key: CHART_TYPE_PIVOT_CHART,
        name: 'Pivot Chart'
      });
    }
  }

  if (queryType === QUERY_TYPE_PROFILE && breakdownLength) {
    menuItems = [
      {
        key: CHART_TYPE_BARCHART,
        name: 'Columns'
      }
    ];
    if (breakdownLength <= 3) {
      menuItems.push({
        key: CHART_TYPE_HORIZONTAL_BAR_CHART,
        name: 'Bars'
      });
    }
    if (breakdownLength > 1) {
      menuItems.push({
        key: CHART_TYPE_PIVOT_CHART,
        name: 'Pivot Chart'
      });
    }
  }
  return menuItems;
};

export const getDateFormatForTimeSeriesChart = ({ frequency }) => {
  return frequency === 'hour'
    ? 'h A, MMM D, YYYY'
    : frequency === 'date' || frequency === 'week'
      ? 'MMM D, YYYY'
      : frequency === 'month'
        ? 'MMM YYYY'
        : 'Q, YYYY';
};
