import MomentTz from 'Components/MomentTz';

export const DATE_RANGE_LABEL_CURRENT_QUARTER = 'This Quarter';
export const DATE_RANGE_LABEL_LAST_QUARTER = 'Last Quarter';
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
    const d = MomentTz();
    return d.day() === 0;
}

export const isTodayTheFirstDayOfMonth = () => {
    const d = MomentTz();
    return d.date() === 1;
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

const QuarterMap = (month, lastXNo = 0) => {
  let rng;
  let mnth = month;
  let year = new Date().getFullYear();
  if(lastXNo) {
    month-3 >= 0? mnth = month-3 : (mnth = 11 , year = year-1);
  }
  
  if(mnth<=2) {
    rng = {
        startDate: MomentTz().set({'month': 0, 'date': 1, 'year': year}).startOf('day'),
        endDate: MomentTz().set({'month': 2, 'year': year}).endOf('month').endOf('day'),
        dateStr: `${year}, Q1`
      }
    } else if(mnth<=5) {
      rng = {
        startDate: MomentTz().set({ 'date': 1, 'month': 3, 'year': year}).startOf('day'),
        endDate: MomentTz().set({'month': 5, 'year': year}).endOf('month').endOf('day'),
        dateStr: `${year}, Q2`
      }
    }
    else if(mnth<=7) {
      rng = {
        startDate: MomentTz().set({ 'date': 1, 'month': 6, 'year': year}).startOf('day'),
        endDate: MomentTz().set({'month': 8, 'year': year}).endOf('month').endOf('day'),
        dateStr: `${year}, Q3`
      }
    } else if (mnth<=11) {
      rng = {
        startDate: MomentTz().set({'month': 9, 'date': 1, 'year': year}).startOf('day'),
        endDate: MomentTz().set({'month': 11, 'year': year}).endOf('month').endOf('day'),
        dateStr: `${year}, Q4`
      }
    }
  return rng;
}

const DEFAULT_DATE_RANGES = [
    {
      label: DATE_RANGE_TODAY_LABEL,
      range: () => ({
        startDate: MomentTz().startOf('day'),
        endDate: MomentTz()
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
        startDate: MomentTz().subtract(1, 'days').startOf('day'),
        endDate: MomentTz().subtract(1, 'days').endOf('day')
      })
    },
    {
      label: DEFAULT_DATE_RANGE_LABEL,
      ...(!isTodayTheFirstDayOfWeek() && {
        range: () => ({
          startDate: MomentTz().startOf('week'),
          endDate: MomentTz().subtract(1, 'days').endOf('day')
        })
      }),
      ...(isTodayTheFirstDayOfWeek() && {
        range: () => ({
          startDate: MomentTz().startOf('day'),
          endDate: MomentTz()
        })
      })
    },
    {
      label: DATE_RANGE_LABEL_CURRENT_MONTH,
      ...(!isTodayTheFirstDayOfMonth() && {
        range: () => ({ 
          startDate: MomentTz().startOf('month'),
          endDate: MomentTz().subtract(1,'days').endOf('day')
        })
      }),
      ...(isTodayTheFirstDayOfMonth() && {
        range: () => ({ 
          startDate: MomentTz().startOf('day'),
          endDate: MomentTz()
        })
      })
    },
    {
      label: DATE_RANGE_LABEL_LAST_WEEK,
      range: () => ({ 
        startDate: MomentTz().startOf('week'),
        endDate: MomentTz().endOf('week')
      })
    },
    {
      label: DATE_RANGE_LABEL_LAST_MONTH,
      range: () => ({ 
        startDate: MomentTz().subtract(1,'months').startOf('month'),
        endDate: MomentTz().subtract(1,'months').endOf('month'),
      })
    },
    {
      label: DATE_RANGE_LABEL_LAST_7_DAYS,
      range: () => ({
        startDate: MomentTz().subtract(7, 'days').startOf('day'),
        endDate: MomentTz()
      })
    },
    {
      label: DATE_RANGE_LABEL_CURRENT_QUARTER,
      range: () => {
        const currMonthNumber = MomentTz().month();
        return QuarterMap(currMonthNumber);
      }
    },
    {
      label: DATE_RANGE_LABEL_LAST_QUARTER,
      range: () => {
        const currMonthNumber = MomentTz().month();
        return QuarterMap(currMonthNumber, 1);
      }
    }
  ]; 