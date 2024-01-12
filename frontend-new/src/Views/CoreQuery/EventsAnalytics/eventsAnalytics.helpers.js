import { startCase, get } from 'lodash';
import moment from 'moment';
import { DATE_FORMATS, QUERY_TYPE_KPI } from '../../../utils/constants';

const capitalizePropertyName = (property) => {
  if (property.includes('$')) {
    return startCase(property.slice(1));
  }
  return property;
};

export const getBreakdownDisplayName = ({
  breakdown,
  userPropNames,
  eventPropertiesDisplayNames,
  queryType,
  multipleEvents
}) => {
  if (queryType === QUERY_TYPE_KPI) {
    return breakdown.display_name
      ? breakdown.display_name
      : startCase(breakdown.property);
  }
  const property = breakdown.pr || breakdown.property;
  const propCategory = breakdown.en || breakdown.prop_category;
  const sessionEventPropertiesDisplayNames = get(
    eventPropertiesDisplayNames,
    '$session',
    {}
  );
  const globalEventPropertiesDisplayNames = get(
    eventPropertiesDisplayNames,
    'global',
    {}
  );

  const displayTitle =
    propCategory === 'user'
      ? get(userPropNames, property, capitalizePropertyName(property))
      : sessionEventPropertiesDisplayNames[property] ||
        globalEventPropertiesDisplayNames[property] ||
        capitalizePropertyName(property);

  if (breakdown.eventIndex && !multipleEvents) {
    return `${displayTitle} (event)`;
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
      return `Facebook ${kpi.label}`;
    }
    if (kpiGroup === 'google_organic_metrics') {
      return `Google ${kpi.label}`;
    }
    if (kpiGroup === 'all_channels_metrics') {
      return `All Channels ${kpi.label}`;
    }
    if (kpiGroup === 'bingads_metrics') {
      return `Bing Ads ${kpi.label}`;
    }
    if (kpiGroup === 'linkedin_metrics') {
      return `Linkedin ${kpi.label}`;
    }
  }
  return get(eventNames, event, event);
};

const getWeekFormat = (m) => {
  const startDate = m.format('D-MMM-YYYY');
  const endDate = m.endOf('week').format('D-MMM-YYYY');
  return `${startDate} to ${endDate}`;
};

export const parseForDateTimeLabel = (grn, label) => {
  let labelValue = label;
  if (grn && moment(label).isValid()) {
    let dateLabel;
    try {
      const newDatr = new Date(label);
      dateLabel = moment(newDatr);
    } catch (e) {
      return label;
    }

    if (
      grn === 'date' ||
      grn === 'day' ||
      grn === 'month' ||
      grn === 'hour' ||
      grn === 'quarter'
    ) {
      labelValue = dateLabel.format(DATE_FORMATS[grn]);
    } else if (grn === 'week') {
      labelValue = getWeekFormat(dateLabel);
    }
  }

  return labelValue;
};
