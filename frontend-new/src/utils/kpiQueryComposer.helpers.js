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
    label: kpi.display_category,
    group: kpi.display_category,
    category: kpi.category,
    icon: 'custom_events',
    values: metricsValues
  };
};
