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

export const getEventDisplayName = ({ eventNames, event }) => {
  return get(eventNames, event, event);
};
