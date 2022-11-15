import startCase from 'lodash/startCase';
import { formatDuration, formatCount } from '../../../utils/dataFormatter';
import { METRIC_TYPES } from '../../../utils/constants';

export const getKpiLabel = (kpi) => {
  const label = kpi.alias || kpi.label;
  if (kpi.category !== 'channels' && kpi.category !== 'custom_channels') {
    return label;
  }
  const labelWithGroupName = `${startCase(kpi.group)} ${label}`;
  if (labelWithGroupName.includes(' Metrics')) {
    return labelWithGroupName.replace(' Metrics', '');
  }
  return labelWithGroupName;
};

export const getFormattedKpiValue = ({ value, metricType }) => {
  if (metricType === METRIC_TYPES.dateType) {
    return formatDuration(value);
  }
  if (metricType === METRIC_TYPES.percentType) {
    return `${formatCount(value, 1)}%`;
  }
  return value;
};
