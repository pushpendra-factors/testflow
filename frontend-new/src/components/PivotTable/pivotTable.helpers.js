import _ from 'lodash';
import { EMPTY_ARRAY } from 'Utils/global';
import {
  QUERY_TYPE_KPI,
  QUERY_TYPE_EVENT,
  QUERY_TYPE_PROFILE
} from 'Utils/constants';

import { getKpiLabel } from '../../Views/CoreQuery/KPIAnalysis/kpiAnalysis.helpers';
import { EVENT_COUNT_KEY } from '../../Views/CoreQuery/EventsAnalytics/eventsAnalytics.constants';
import {
  getBreakdownDisplayName,
  getEventDisplayName
} from '../../Views/CoreQuery/EventsAnalytics/eventsAnalytics.helpers';
import { getProfileQueryDisplayName } from '../../Views/CoreQuery/ProfilesResultPage/BreakdownCharts/utils';

export const getMetricLabel = ({ metric, queryType, eventNames }) => {
  if (queryType === QUERY_TYPE_KPI) {
    return getKpiLabel(metric);
  }
  if (queryType === QUERY_TYPE_EVENT) {
    return getEventDisplayName({ event: metric, eventNames });
  }
  if (queryType === QUERY_TYPE_PROFILE) {
    return getProfileQueryDisplayName({
      query: metric,
      groupAnalysis: 'users'
    });
  }
  return metric;
};

export const getMetricValue = ({
  metric,
  index,
  dataObject,
  queryType,
  eventNames
}) => {
  if (queryType === QUERY_TYPE_KPI) {
    return dataObject[
      `${getMetricLabel({ metric, queryType, eventNames })} - ${index}`
    ];
  }
  if (queryType === QUERY_TYPE_EVENT) {
    return dataObject[EVENT_COUNT_KEY];
    // if (metricsLength === 1) {
    // }
    // return dataObject[
    //   `${getMetricLabel({ metric, queryType, eventNames })} - ${index}`
    // ];
  }
  if (queryType === QUERY_TYPE_PROFILE) {
    return dataObject.value;
  }
  return 0;
};

export const formatPivotData = ({
  data,
  breakdown,
  metrics,
  queryType,
  eventNames,
  userPropNames,
  eventPropNames
}) => {
  try {
    const breakdownAttributes = breakdown.map((b) =>
      getBreakdownDisplayName({
        breakdown: b,
        userPropNames,
        eventPropNames,
        queryType
      })
    );
    const metricAttributes = metrics.map((metric) =>
      getMetricLabel({ metric, queryType, eventNames })
    );
    const attributesRow = breakdownAttributes.concat(metricAttributes);
    const values = data.map((d) => {
      const breakdownVals = breakdown.map((b, index) => {
        return d[`${b.property} - ${index}`];
      });
      const metricVals = metrics.map((metric, index) => {
        return getMetricValue({
          metric,
          index,
          dataObject: d,
          queryType,
          eventNames
        });
      });
      const current = breakdownVals.concat(metricVals);
      return current;
    });
    return [breakdownAttributes, attributesRow, values];
  } catch (err) {
    console.log('formatPivotData -> err', err);
    return EMPTY_ARRAY;
  }
};

export const getValueOptions = ({ metrics, queryType, eventNames }) => {
  return _.map(metrics, (metric) =>
    getMetricLabel({ metric, queryType, eventNames })
  );
};

export const getColumnOptions = ({
  breakdown,
  userPropNames,
  eventPropNames,
  queryType
}) => {
  return _.map(breakdown, (b) =>
    getBreakdownDisplayName({
      breakdown: b,
      userPropNames,
      eventPropNames,
      queryType
    })
  );
};

export const getRowOptions = ({
  selectedRows,
  metrics,
  breakdown,
  queryType,
  eventNames,
  userPropNames,
  eventPropNames
}) => {
  const valueOptions = getValueOptions({ metrics, queryType, eventNames });
  const columnOptions = getColumnOptions({
    breakdown,
    userPropNames,
    eventPropNames,
    queryType
  });
  const allOptions = _.concat(valueOptions, columnOptions);
  return _.filter(allOptions, (option) => !selectedRows.includes(option));
};

export const SortRowOptions = ({
  data,
  metrics,
  breakdown,
  queryType,
  eventNames,
  userPropNames,
  eventPropNames
}) => {
  const breakdownOptions = getColumnOptions({
    breakdown,
    userPropNames,
    eventPropNames,
    queryType
  });
  const metricsOptions = getValueOptions({ metrics, queryType, eventNames });
  const breakdownsSelected = data.filter((d) => breakdownOptions.includes(d));
  const metricsSelected = data.filter((d) => metricsOptions.includes(d));
  return [...breakdownsSelected, ...metricsSelected];
};

export const getFunctionOptions = () => {
  return [
    'Integer Sum',
    'Sum',
    'Count',
    'Average',
    'Median',
    'Sum as Fraction of Rows',
    'Sum as Fraction of Columns',
    'Sum as Fraction of Total'
  ];
};
