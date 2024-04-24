import MomentTz from 'Components/MomentTz';

export const DEFAULT_DATE_RANGE = {
  startDate: MomentTz().startOf('month'),
  endDate: MomentTz().subtract(1, 'day').endOf('day'),
  dateString: 'This Month',
  dateType: 'this_month'
};
