/* eslint-disable */
import { createStaticRanges } from 'react-date-range';
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

export const DEFAULT_DATE_RANGE = {
  ...(!isTodayTheFirstDayOfWeek() && {
    startDate: MomentTz(getFirstDayOfCurrentWeek()).startOf('day').toDate(),
    endDate: MomentTz(new Date()).subtract(1, 'days').endOf('day').toDate()
  }),
  ...(isTodayTheFirstDayOfWeek() && {
    startDate: MomentTz(new Date()).startOf('day').toDate(),
    endDate: new Date()
  }),
  label: DEFAULT_DATE_RANGE_LABEL,
  key: 'selected'
};

function getFirstDayOfCurrentWeek() {
  const d = new Date();
  const first = d.getDate() - d.getDay();
  return new Date(d.setDate(first));
}

function getFirstDayOfLastWeek() {
  const d = new Date();
  const first = d.getDate() - d.getDay() - 7;
  return new Date(d.setDate(first));
}

function getLastDayOfLastWeek() {
  const d = new Date();
  const last = d.getDate() - d.getDay() - 1;
  return new Date(d.setDate(last));
}

function getFirstDayOfLastMonth() {
  const d = new Date();
  return new Date(d.getFullYear(), d.getMonth() - 1, 1);
}

function getLastDayOfLastMonth() {
  const d = new Date();
  return new Date(d.getFullYear(), d.getMonth(), 0);
}

function getFirstDayOfCurrentMonth() {
  const d = new Date();
  return new Date(d.getFullYear(), d.getMonth(), 1);
}

function isTodayTheFirstDayOfMonth() {
  const d = new Date();
  return d.getDate() === 1;
}

function isTodayTheFirstDayOfWeek() {
  // week starts with Sunday.
  const d = new Date();
  return d.getDay() === 0;
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

export const DEFAULT_TODAY_DATE_RANGES = [
  {
    label: DATE_RANGE_LAST_2_MIN_LABEL,
    range: () => ({
      startDate: MomentTz(new Date()).subtract(60 * 2, 'seconds').toDate(),
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
    label: DATE_RANGE_LAST_30_MIN_LABEL,
    range: () => ({
      startDate: MomentTz(new Date()).subtract(60 * 30, 'seconds').toDate(),
      endDate: new Date()
    }),
    isSelected(range) {
      const definedRange = this.range();
      return (
        MomentTz(range.startDate).isSame(definedRange.startDate, 'seconds') &&
        MomentTz(range.endDate).isSame(definedRange.endDate, 'seconds')
      );
    }
  }
];
export const DEFINED_DATE_RANGES = createStaticRanges(DEFAULT_DATE_RANGES);
export const WEB_ANALYTICS_DEFINED_DATE_RANGES = createStaticRanges([...DEFAULT_TODAY_DATE_RANGES, ...DEFAULT_DATE_RANGES]);

// returns datepicker daterange for stored daterange.
// updates the daterange with currentTime, if ovp true.
// stored = { fr: UNIX_TIMESTAMP, to: UNIX_TIMESTAMP, ovp: true }
// datepicker = [{ startDate: DATE, endDate: DATE, key: 'selected' }]
// export const getDateRangeFromStoredDateRange = (storedRange) => {
//   if (storedRange.ovp) {
//     const newInterval = slideUnixTimeWindowToCurrentTime(storedRange.fr, storedRange.to);
//     storedRange.fr = newInterval.from;
//     storedRange.to = newInterval.to;
//   }

//   return [{
//     startDate: MomentTz.unix(storedRange.fr).toDate(),
//     endDate: MomentTz.unix(storedRange.to).toDate(),
//     key: 'selected'
//   }];
// };

export const readableDateRange = function (range) {
  const defaultRange = DEFAULT_DATE_RANGES.filter((rng) => {
    const rngDates = rng.range();
    return MomentTz(rngDates.startDate).isSame(MomentTz(range.startDate)) && MomentTz(rngDates.endDate).isSame(MomentTz(range.endDate));
  });
  if (defaultRange.length) {
    return defaultRange[0].label;
  }

  return MomentTz(range.startDate).format('MMM DD, YYYY') + ' - ' +
    MomentTz(range.endDate).format('MMM DD, YYYY');
};

export const displayRange = (range) => {
  return MomentTz(range.startDate).format('MMM DD, YYYY') + ' - ' +
    MomentTz(range.endDate).format('MMM DD, YYYY');
}

export const getDateRange = (durationObj) => {
  const ranges = [{ ...DEFAULT_DATE_RANGE }];
  const queryOptionsState = { ...durationObj };

  if (
    queryOptionsState &&
    queryOptionsState.from &&
    queryOptionsState.to
  ) {
    ranges[0].startDate = MomentTz(queryOptionsState.from).toDate();
    ranges[0].endDate = MomentTz(queryOptionsState.to).toDate();
  }

  return ranges;
};
