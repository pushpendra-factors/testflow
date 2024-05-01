import MomentTz from 'Components/MomentTz';

export const DEFAULT_DATE_RANGE = {
  startDate:
    MomentTz().format('DD') === '01'
      ? MomentTz().subtract(1, 'day').startOf('month')
      : MomentTz().startOf('month'),
  endDate: MomentTz().subtract(1, 'day').endOf('day'),
  dateString: MomentTz().format('DD') === '01' ? 'Last Month' : 'This Month',
  dateType: MomentTz().format('DD') === '01' ? 'last_month' : 'this_month'
};
