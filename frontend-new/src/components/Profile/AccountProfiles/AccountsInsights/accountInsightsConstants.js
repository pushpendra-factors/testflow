import MomentTz from 'Components/MomentTz';

export const DEFAULT_DATE_RANGE = {
  startDate: MomentTz().subtract(1, 'month').startOf('month'),
  endDate: MomentTz().subtract(1, 'month').endOf('month'),
  dateString: 'Last Month',
  dateType: 'last_month'
};
