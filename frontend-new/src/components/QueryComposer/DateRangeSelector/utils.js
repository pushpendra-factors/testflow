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
    startDate: MomentTz(getFirstDayOfCurrentWeek()).startOf('day'),
    endDate: MomentTz(new Date()).subtract(1, 'days').endOf('day')
  }),
  ...(isTodayTheFirstDayOfWeek() && {
    startDate: MomentTz(new Date()).startOf('day'),
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
  const d = MomentTz();
  return d.date() === 1;
}

function isTodayTheFirstDayOfWeek() {
  // week starts with Sunday.
  const d = MomentTz();
  return d.day() === 0;
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
  }
];

export const DEFAULT_TODAY_DATE_RANGES = [
  {
    label: DATE_RANGE_LAST_2_MIN_LABEL,
    range: () => ({
      startDate: MomentTz().subtract(60 * 2, 'seconds'),
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
    label: DATE_RANGE_LAST_30_MIN_LABEL,
    range: () => ({
      startDate: MomentTz().subtract(60 * 30, 'seconds'),
      endDate: MomentTz()
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
//     startDate: MomentTz.unix(storedRange.fr),
//     endDate: MomentTz.unix(storedRange.to),
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
    ranges[0].startDate = MomentTz(queryOptionsState.from);
    ranges[0].endDate = MomentTz(queryOptionsState.to);
  }

  return ranges;
};
