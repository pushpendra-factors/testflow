import _ from 'lodash';
import { QUERY_TYPE_KPI } from '../../../utils/constants';

export const getBreakdownDisplayName = ({
  breakdown,
  userPropNames,
  eventPropNames,
  queryType
}) => {
  if (queryType === QUERY_TYPE_KPI) {
    return _.startCase(breakdown.property);
  }
  const property = breakdown.pr || breakdown.property;
  const propCategory = breakdown.en || breakdown.prop_category;
  const displayTitle =
    propCategory === 'user'
      ? _.get(userPropNames, property, property)
      : propCategory === 'event'
        ? _.get(eventPropNames, property, property)
        : property;

  if (breakdown.eventIndex) {
    return displayTitle + ' (event)';
  }
  return displayTitle;
};

export const getEventDisplayName = ({ eventNames, event }) => {
  return _.get(eventNames, event, event);
};
