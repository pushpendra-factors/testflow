import MomentTz from '../MomentTz';

export const groups = {
  Timestamp: (item) =>
    MomentTz(item.timestamp * 1000).format('DD MMMM YYYY, hh:mm:ss '),
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
