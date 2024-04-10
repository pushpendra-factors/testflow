import MomentTz from 'Components/MomentTz';

export const DEFAULT_DATE_RANGE = {
  startDate: MomentTz().startOf('month'),
  endDate: MomentTz(),
  dateString: 'This Month',
  dateType: 'this_month'
};
