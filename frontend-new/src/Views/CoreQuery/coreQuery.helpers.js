import { get } from 'lodash';
import {
  QUERY_TYPE_EVENT,
  QUERY_TYPE_FUNNEL,
  QUERY_TYPE_KPI,
  QUERY_TYPE_CAMPAIGN,
  QUERY_TYPE_PROFILE,
  QUERY_TYPE_ATTRIBUTION
} from '../../utils/constants';
import { DEFAULT_PIVOT_CONFIG } from './constants';
import { isPivotSupported } from '../../utils/chart.helpers';
import { parseForDateTimeLabel } from './EventsAnalytics/eventsAnalytics.helpers';
import { formatCount } from 'Utils/dataFormatter';
import { isNumeric } from 'Utils/global';

export const IconAndTextSwitchQueryType = (queryType) => {
  switch (queryType) {
    case QUERY_TYPE_EVENT:
      return {
        text: 'Analyse Events',
        icon: 'events_cq'
      };
    case QUERY_TYPE_FUNNEL:
      return {
        text: 'Find event funnel for',
        icon: 'funnels_cq'
      };
    case QUERY_TYPE_CAMPAIGN:
      return {
        text: 'Campaign Analytics',
        icon: 'campaigns_cq'
      };
    case QUERY_TYPE_ATTRIBUTION:
      return {
        text: 'Attributions',
        icon: 'attributions_cq'
      };
    case QUERY_TYPE_KPI:
      return {
        text: 'KPI',
        icon: 'attributions_cq'
      };
    case QUERY_TYPE_PROFILE:
      return {
        text: 'Profile Analysis',
        icon: 'profiles_cq'
      };
    default:
      return {
        text: 'Templates',
        icon: 'templates_cq'
      };
  }
};

export const getSavedPivotConfig = ({ queryType, selectedReport }) => {
  if (!isPivotSupported({ queryType })) {
    return { ...DEFAULT_PIVOT_CONFIG };
  }
  const savedPivotConfig = get(selectedReport, 'settings.pivotConfig', null);
  return savedPivotConfig
    ? JSON.parse(savedPivotConfig)
    : { ...DEFAULT_PIVOT_CONFIG };
};

export const getDifferentDates = ({ rows, dateIndex }) => {
  const differentDates = new Set();
  rows.forEach((row) => {
    differentDates.add(row[dateIndex]);
  });
  return Array.from(differentDates);
};

export const formatBreakdownLabel = ({ grn, propType, label }) => {
  if (grn) {
    return parseForDateTimeLabel(grn, label);
  }
  if (propType === 'numerical') {
    if (typeof label === 'number') {
      return formatCount(label);
    }
    if (isNumeric(label)) {
      return formatCount(Number(label)).toString();
    }
  }
  return label;
};
