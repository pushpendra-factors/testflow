import MomentTz from 'Components/MomentTz';

export const DATE_RANGE_LABEL_CURRENT_MONTH = 'This Month';
export const DEFAULT_DATE_RANGE_LABEL = 'This Week';
export const DATE_RANGE_LABEL_LAST_MONTH = 'Last Month';
export const DATE_RANGE_LABEL_LAST_WEEK = 'Last Week';
export const DATE_RANGE_YESTERDAY_LABEL = 'Yesterday';
export const DATE_RANGE_TODAY_LABEL = 'Today';
export const DATE_RANGE_LABEL_LAST_7_DAYS = 'Last 7 Days';
export const DATE_RANGE_LAST_2_MIN_LABEL = 'Last 2 mins';
export const DATE_RANGE_LAST_30_MIN_LABEL = 'Last 30 mins';

export const getFirstDayOfLastWeek = () => {
    const d = new Date();
    const first = d.getDate() - d.getDay() - 7;
    return new Date(d.setDate(first));
}

export const getLastDayOfLastWeek = () => {
    const d = new Date();
    const last = d.getDate() - d.getDay() - 1;
    return new Date(d.setDate(last));
}

export const getFirstDayOfLastMonth = () => {
    const d = new Date();
    return new Date(d.getFullYear(), d.getMonth() - 1, 1);
}
  
export const getLastDayOfLastMonth = () => {
    const d = new Date();
    return new Date(d.getFullYear(), d.getMonth(), 0);
}

export const isTodayTheFirstDayOfWeek = () => {
    // week starts with Sunday.
    const d = new Date();
    return d.getDay() === 0;
}

export const isTodayTheFirstDayOfMonth = () => {
    const d = new Date();
    return d.getDate() === 1;
}

export const getFirstDayOfCurrentWeek = () => {
    const d = new Date();
    const first = d.getDate() - d.getDay();
    return new Date(d.setDate(first));
}
   
export const getFirstDayOfCurrentMonth = () => {
    const d = new Date();
    return new Date(d.getFullYear(), d.getMonth(), 1);
}

export const getRangeByLabel = (label) => {
    const rangeObj = DEFAULT_DATE_RANGES.filter(rng => rng.label === label);
    let rnge = null;
    if(rangeObj.length >= 1) {
        rnge = rangeObj[0].range();
    }
    return rnge;

}

const DEFAULT_DATE_RANGES = [
    {
      label: DATE_RANGE_TODAY_LABEL,
      range: () => ({
        startDate: MomentTz(new Date()).startOf('day').toDate(),
        endDate: new Date()
      }),
      isSelected(range) {
        const definedRange = this.range();
        return (
          MomentTz(range.startDate).isSame(definedRange.startDate, 'seconds') &&
          MomentTz(range.endDate).isSame(definedRange.endDate, 'seconds')
        );
      }
    },
    {
      label: DATE_RANGE_YESTERDAY_LABEL,
      range: () => ({
        startDate: MomentTz(new Date()).subtract(1, 'days').startOf('day').toDate(),
        endDate: MomentTz(new Date()).subtract(1, 'days').endOf('day').toDate()
      })
    },
    {
      label: DEFAULT_DATE_RANGE_LABEL,
      ...(!isTodayTheFirstDayOfWeek() && {
        range: () => ({
          startDate: MomentTz(getFirstDayOfCurrentWeek()).startOf('day').toDate(),
          endDate: MomentTz(new Date()).subtract(1, 'days').endOf('day').toDate()
        })
      }),
      ...(isTodayTheFirstDayOfWeek() && {
        range: () => ({
          startDate: MomentTz(new Date()).startOf('day').toDate(),
          endDate: new Date()
        })
      })
    },
    {
      label: DATE_RANGE_LABEL_CURRENT_MONTH,
      ...(!isTodayTheFirstDayOfMonth() && {
        range: () => ({
          startDate: MomentTz(getFirstDayOfCurrentMonth()).startOf('day').toDate(),
          endDate: MomentTz(new Date()).subtract(1, 'days').endOf('day').toDate()
        })
      }),
      ...(isTodayTheFirstDayOfMonth() && {
        range: () => ({
          startDate: MomentTz(new Date()).startOf('day').toDate(),
          endDate: new Date()
        })
      })
    },
    {
      label: DATE_RANGE_LABEL_LAST_WEEK,
      range: () => ({
        startDate: MomentTz(getFirstDayOfLastWeek()).startOf('day').toDate(),
        endDate: MomentTz(getLastDayOfLastWeek()).endOf('day').toDate()
      })
    },
    {
      label: DATE_RANGE_LABEL_LAST_MONTH,
      range: () => ({
        startDate: MomentTz(getFirstDayOfLastMonth()).startOf('day').toDate(),
        endDate: MomentTz(getLastDayOfLastMonth()).endOf('day').toDate()
      })
    },
    {
      label: DATE_RANGE_LABEL_LAST_7_DAYS,
      range: () => ({
        startDate: MomentTz(new Date()).subtract(7, 'days').startOf('day').toDate(),
        endDate: MomentTz(new Date())
      })
    }
  ];