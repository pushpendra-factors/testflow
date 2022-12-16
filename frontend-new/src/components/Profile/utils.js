import MomentTz from '../MomentTz';
import { operatorMap } from '../../Views/CoreQuery/utils';
import { formatDurationIntoString } from 'Utils/dataFormatter';

export const granularityOptions = [
  'Timestamp',
  'Hourly',
  'Daily',
  'Weekly',
  'Monthly'
];

export const groups = {
  Timestamp: (item) =>
    MomentTz(item.timestamp * 1000).format('DD MMM YYYY, hh:mm:ss '),
  Hourly: (item) =>
    `${MomentTz(item.timestamp * 1000)
      .startOf('hour')
      .format('hh A')} - ${MomentTz(item.timestamp * 1000)
      .add(1, 'hour')
      .startOf('hour')
      .format('hh A')} ${MomentTz(item.timestamp * 1000)
      .startOf('hour')
      .format('DD MMM YYYY')}`,
  Daily: (item) =>
    MomentTz(item.timestamp * 1000)
      .startOf('day')
      .format('DD MMM YYYY'),
  Weekly: (item) =>
    `${MomentTz(item.timestamp * 1000)
      .endOf('week')
      .format('DD MMM YYYY')} - ${MomentTz(item.timestamp * 1000)
      .startOf('week')
      .format('DD MMM YYYY')}`,
  Monthly: (item) =>
    MomentTz(item.timestamp * 1000)
      .startOf('month')
      .format('MMM YYYY')
};

export const hoverEvents = [
  '$session',
  '$form_submitted',
  '$offline_touch_point',
  '$sf_campaign_member_created',
  '$sf_campaign_member_updated',
  '$hubspot_form_submission',
  '$hubspot_engagement_email',
  '$hubspot_engagement_meeting_created',
  '$hubspot_engagement_meeting_updated',
  '$hubspot_engagement_call_created',
  '$hubspot_engagement_call_updated'
];

export const TimelineHoverPropDisplayNames = {
  $timestamp: 'Date and Time',
  '$hubspot_form_submission_form-type': 'Form Type',
  $hubspot_form_submission_title: 'Form Title',
  '$hubspot_form_submission_form-id': 'Form ID',
  '$hubspot_form_submission_conversion-id': 'Conversion ID',
  $hubspot_form_submission_email: 'Email',
  '$hubspot_form_submission_page-url-no-qp': 'Page URL',
  '$hubspot_form_submission_page-title': 'Page Title',
  $hubspot_form_submission_timestamp: 'Form Submit Timestamp'
};

export const formatFiltersForPayload = (filters = []) => {
  const filterProps = [];
  filters.forEach((fil) => {
    if (Array.isArray(fil.values)) {
      fil.values.forEach((val, index) => {
        filterProps.push({
          en: 'user_g',
          lop: !index ? 'AND' : 'OR',
          op: operatorMap[fil.operator],
          pr: fil.props[0],
          ty: fil.props[1],
          va: fil.props[1] === 'datetime' ? val : val
        });
      });
    } else {
      filterProps.push({
        en: 'user_g',
        lop: 'AND',
        op: operatorMap[fil.operator],
        pr: fil.props[0],
        ty: fil.props[1],
        va: fil.props[1] === 'datetime' ? fil.values : fil.values
      });
    }
  });
  return filterProps;
};

export const eventsFormattedForGranularity = (
  events,
  granularity,
  collapse = true
) => {
  const output = events.reduce((result, item) => {
    const byTimestamp = (result[groups[granularity](item)] =
      result[groups[granularity](item)] || {});
    const byUser = (byTimestamp[item.user] = byTimestamp[item.user] || {
      events: [],
      collapsed: collapse
    });
    byUser.events.push(item);
    return result;
  }, {});
  return output;
};

export const toggleCellCollapse = (
  formattedData,
  timestamp,
  username,
  collapseState
) => {
  const data = { ...formattedData };
  data[timestamp][username].collapsed = collapseState;
  return data;
};

const isValidHttpUrl = (string) => {
  let url;
  try {
    url = new URL(string);
  } catch (_) {
    return false;
  }
  return url.protocol === 'http:' || url.protocol === 'https:';
};

export const getHost = (urlstr) => {
  const uri = isValidHttpUrl(urlstr) ? new URL(urlstr).hostname : urlstr;
  return uri;
};

export const getUniqueItemsByKeyAndSearchTerm = (activities, searchTerm) =>
  activities?.filter(
    (value, index, self) =>
      index === self.findIndex((t) => t.display_name === value.display_name) &&
      value.display_name.toLowerCase().includes(searchTerm.toLowerCase())
  );


export const propValueFormat = (key, value) => {
    if (
      key.includes('timestamp') ||
      key.includes('starttime') ||
      key.includes('endtime')
    ) {
      return MomentTz(value * 1000).format('DD MMMM YYYY, hh:mm A');
    }
    if (key.includes('_time')) {
      return formatDurationIntoString(value);
    }
    if (key.includes('durationmilliseconds')) {
      return formatDurationIntoString(parseInt(value / 1000));
    }
    return value;
  };