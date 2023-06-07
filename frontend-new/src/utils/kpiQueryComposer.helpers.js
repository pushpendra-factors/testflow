import { get, map } from 'lodash';
import { EMPTY_ARRAY } from './global';
import getGroupIcon from './getGroupIcon';
export const getNormalizedKpi = ({ kpi }) => {
  const metrics = get(kpi, 'metrics', EMPTY_ARRAY);
  const metricsValues = map(metrics, (metric) => {
    return [
      get(metric, 'display_name', metric),
      get(metric, 'name', metric),
      get(metric, 'type', '')
    ];
  });
  return {
    label: get(kpi, 'display_category'),
    group: get(kpi, 'display_category'),
    category: get(kpi, 'category'),
    icon: getGroupIcon(get(kpi, 'display_category')),
    values: metricsValues
  };
};

export const getNormalizedKpiWithConfigs = ({ kpi }) => {
  const metrics = get(kpi, 'metrics', EMPTY_ARRAY);
  const metricsValues = map(metrics, (metric) => {
    return [
      get(metric, 'display_name', metric),
      get(metric, 'name', metric),
      get(metric, 'type', ''),
      get(metric, 'kpi_query_type', metric),
    ];
  });
  return {
    label: get(kpi, 'display_category'),
    group: get(kpi, 'display_category'),
    category: get(kpi, 'category'),
    icon: getGroupIcon(get(kpi, 'display_category')),
    values: metricsValues
  };
};

export const areKpisInSameGroup = ({ kpis }) => {
  return kpis.every((_, index) => {
    if (kpis[0].group == 'others') {
      return false;
    } else return kpis[0].group === kpis[index].group;
  });
};
