import MomentTz from '../MomentTz';
import { operatorMap } from '../../Views/CoreQuery/utils';

export const granularityOptions = [
  'Timestamp',
  'Hourly',
  'Daily',
  'Weekly',
  'Monthly',
];

export const groups = {
  Timestamp: (item) =>
    MomentTz(item.timestamp * 1000).format('DD MMM YYYY, hh:mm:ss '),
  Hourly: (item) =>
    MomentTz(item.timestamp * 1000)
      .startOf('hour')
      .format('hh A') +
    ' - ' +
    MomentTz(item.timestamp * 1000)
      .add(1, 'hour')
      .startOf('hour')
      .format('hh A') +
    ' ' +
    MomentTz(item.timestamp * 1000)
      .startOf('hour')
      .format('DD MMM YYYY'),
  Daily: (item) =>
    MomentTz(item.timestamp * 1000)
      .startOf('day')
      .format('DD MMM YYYY'),
  Weekly: (item) =>
    MomentTz(item.timestamp * 1000)
      .endOf('week')
      .format('DD MMM YYYY') +
    ' - ' +
    MomentTz(item.timestamp * 1000)
      .startOf('week')
      .format('DD MMM YYYY'),
  Monthly: (item) =>
    MomentTz(item.timestamp * 1000)
      .startOf('month')
      .format('MMM YYYY'),
};

export const hoverEvents = [
  'Website Session',
  'Page View',
  'Form Button Click',
  'Campaign Member Created',
  'Campaign Member Updated',
  'Offline Touchpoint',
];

export const getLoopLength = (allEvents) => {
  let maxLength = -1;
  Object.entries(allEvents).forEach(([user, events]) => {
    if (maxLength < events.length) maxLength = events.length;
  });
  return maxLength;
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
          va: fil.props[1] === 'datetime' ? val : val,
        });
      });
    } else {
      filterProps.push({
        en: 'user_g',
        lop: 'AND',
        op: operatorMap[fil.operator],
        pr: fil.props[0],
        ty: fil.props[1],
        va: fil.props[1] === 'datetime' ? fil.values : fil.values,
      });
    }
  });
  return filterProps;
};

export const eventsFormattedForGranularity = (
  events,
  granularity,
  collapse
) => {
  const data = _.groupBy(events, groups[granularity]);
  let retData = {};
  Object.entries(data).forEach(([timestamp, events]) => {
    const groupByUser = _.groupBy(events, (item) => item.user);
    const obj = new Object();
    obj[timestamp] = groupByUser;
    retData = { ...retData, ...obj };
  });
  return retData;
};
