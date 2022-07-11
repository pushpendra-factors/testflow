import { get, map } from 'lodash';
import { EMPTY_ARRAY } from './global';

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
    icon: 'custom_events',
    values: metricsValues
  };
};

export const areKpisInSameGroup = ({ kpis }) => {
  return kpis.every(
    (_, index) => kpis[0].group === kpis[index].group
  );
};
