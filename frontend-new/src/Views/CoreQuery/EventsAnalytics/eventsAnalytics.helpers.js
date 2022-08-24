import { startCase, get } from 'lodash';
import { QUERY_TYPE_KPI } from '../../../utils/constants';

export const getBreakdownDisplayName = ({
  breakdown,
  userPropNames,
  eventPropNames,
  queryType,
  multipleEvents
}) => {
  if (queryType === QUERY_TYPE_KPI) {
    return startCase(breakdown.property);
  }
  const property = breakdown.pr || breakdown.property;
  const propCategory = breakdown.en || breakdown.prop_category;
  const displayTitle =
    propCategory === 'user'
      ? get(userPropNames, property, property)
      : propCategory === 'event'
      ? get(eventPropNames, property, property)
      : property;

  if (breakdown.eventIndex && !multipleEvents) {
    return displayTitle + ' (event)';
  }
  return displayTitle;
};

export const getEventDisplayName = ({ eventNames, event, queryType, kpi }) => {
  if (
    queryType === QUERY_TYPE_KPI &&
    kpi != null &&
    get(kpi, 'category', null) === 'channels'
  ) {
    const kpiGroup = get(kpi, 'group', null);
    if (kpiGroup === 'facebook_metrics') {
      return 'Facebook ' + kpi.label;
    }
    if (kpiGroup === 'google_organic_metrics') {
      return 'Google ' + kpi.label;
    }
    if (kpiGroup === 'all_channels_metrics') {
      return 'All Channels ' + kpi.label;
    }
    if (kpiGroup === 'bingads_metrics') {
      return 'Bing Ads ' + kpi.label;
    }
    if (kpiGroup === 'linkedin_metrics') {
      return 'Linkedin ' + kpi.label;
    }
  }
  return get(eventNames, event, event);
};
