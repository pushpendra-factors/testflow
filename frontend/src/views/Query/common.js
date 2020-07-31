import { createStaticRanges } from 'react-date-range';
import moment from 'moment';
import mt from "moment-timezone";

import { slideUnixTimeWindowToCurrentTime, firstToUpperCase, getTimezoneString } from '../../util';

export const QUERY_TYPE_UNIQUE_USERS = "unique_users";
export const QUERY_TYPE_EVENTS_OCCURRENCE = "events_occurrence";

export const PROPERTY_KEY_JOIN_TIME = "$joinTime";
export const PROPERTY_VALUE_NONE = "$none";

export const PROPERTY_TYPE_EVENT = 'event';
export const PROPERTY_TYPE_USER = 'user';
export const PROPERTY_TYPE_OPTS = {
  'event': 'event property',
  'user': 'user property'
};
export const PROPERTY_LOGICAL_OP_OPTS = {
  'AND': 'and',
  'OR': 'or',
}; 

export const DASHBOARD_TYPE_WEB_ANALYTICS = "Website Analytics";
export const QUERY_CLASS_CHANNEL = "channel";
export const QUERY_CLASS_FUNNEL = "funnel";
export const QUERY_CLASS_ATTRIBUTION = 'attribution';
export const QUERY_CLASS_WEB = 'web';
export const PROPERTY_VALUE_TYPE_DATE_TIME = 'datetime';

export const USER_PREF_PROPERTY_TYPE_OPTS = {
  // user property preferred on top/default.
  'user': 'user property', 
  'event': 'event property'
};

export const HEADER_COUNT = "count";
export const HEADER_DATE = "datetime";

export const PRESENTATION_TABLE = 'pt';
export const PRESENTATION_LINE =  'pl';
export const PRESENTATION_BAR = 'pb';
// alias for single count table view.
export const PRESENTATION_CARD = 'pc';
export const PRESENTATION_FUNNEL = 'pf';


export const DEFAULT_DATE_RANGE_LABEL = 'Last 7 days';
export const DATE_RANGE_TODAY_LABEL = 'Today';
export const DATE_RANGE_YESTERDAY_LABEL = 'Yesterday';
export const DATE_RANGE_LAST_30_DAYS_LABEL = 'Last 30 days';
export const DATE_RANGE_LAST_2_MIN_LABEL = 'Last 2 mins'
export const DATE_RANGE_LAST_30_MIN_LABEL = 'Last 30 mins'

export const LABEL_STYLE = { marginRight: '10px', fontWeight: '600', color: '#777' };

export const DEFAULT_DATE_RANGE = {
  startDate: moment(new Date()).subtract(7, 'days').startOf('day').toDate(),
  endDate: moment(new Date()).subtract(1, 'days').endOf('day').toDate(),
  label: DEFAULT_DATE_RANGE_LABEL,
  key: 'selected'
}
const DEFAULT_DATE_RANGES = [
  {
    label: DATE_RANGE_TODAY_LABEL,
    range: () => ({
      startDate: moment(new Date()).startOf('day').toDate(),
      endDate: new Date(),
    }),
    isSelected(range) {
      const definedRange = this.range();
      return (
        moment(range.startDate).isSame(definedRange.startDate,"seconds") && 
        moment(range.endDate).isSame(definedRange.endDate,"seconds")
      );
    }
  },
  {
    label: DATE_RANGE_YESTERDAY_LABEL,
    range: () => ({
      startDate: moment(new Date()).subtract(1, 'days').startOf('day').toDate(),
      endDate: moment(new Date()).subtract(1, 'days').endOf('day').toDate(),
    }),
  },
  {
    label: DEFAULT_DATE_RANGE_LABEL,
    range: () => ({
      startDate: moment(new Date()).subtract(7, 'days').startOf('day').toDate(),
      endDate: moment(new Date()).subtract(1, 'days').endOf('day').toDate(),
    }),
  },
  {
    label: DATE_RANGE_LAST_30_DAYS_LABEL,
    range: () => ({
      startDate: moment(new Date()).subtract(30, 'days').startOf('day').toDate(),
      endDate: moment(new Date()).subtract(1, 'days').endOf('day').toDate(),
    })
  },
];

export const DEFAULT_TODAY_DATE_RANGES = [
  {
    label: DATE_RANGE_LAST_2_MIN_LABEL,
    range: () => ({
      startDate: moment(new Date()).subtract(60*2,'seconds').toDate(),
      endDate: new Date(),
    }),
    isSelected(range) {
      const definedRange = this.range();
      return (
        moment(range.startDate).isSame(definedRange.startDate,"seconds") && 
        moment(range.endDate).isSame(definedRange.endDate,"seconds")
      );
    }  
  },
  {
    label:DATE_RANGE_LAST_30_MIN_LABEL,
    range: () => ({
      startDate: moment(new Date()).subtract(60*30, 'seconds').toDate(),
      endDate: new Date(),
    }),
    isSelected(range) {
      const definedRange = this.range();
      return (
        moment(range.startDate).isSame(definedRange.startDate,"seconds") && 
        moment(range.endDate).isSame(definedRange.endDate,"seconds")
      );
    } 
  }
]
export const DEFINED_DATE_RANGES = createStaticRanges(DEFAULT_DATE_RANGES);
export const WEB_ANALYTICS_DEFINED_DATE_RANGES = createStaticRanges([...DEFAULT_TODAY_DATE_RANGES,...DEFAULT_DATE_RANGES])


// returns datepicker daterange for stored daterange.
// updates the daterange with currentTime, if ovp true.
// stored = { fr: UNIX_TIMESTAMP, to: UNIX_TIMESTAMP, ovp: true }
// datepicker = [{ startDate: DATE, endDate: DATE, key: 'selected' }]
export const getDateRangeFromStoredDateRange = (storedRange) => {
  if (storedRange.ovp) {
    let newInterval = slideUnixTimeWindowToCurrentTime(storedRange.fr, storedRange.to);
    storedRange.fr = newInterval.from;
    storedRange.to = newInterval.to;
  }

  return [{ 
    startDate: moment.unix(storedRange.fr).toDate(), 
    endDate: moment.unix(storedRange.to).toDate(),
    key: "selected",
  }];
}

export const getYAxesStr = function(queryType, aggr="count") {
  let dAggr = firstToUpperCase(aggr);
  if (!queryType || queryType == "") return dAggr;
  let entity = queryType == QUERY_TYPE_UNIQUE_USERS ? "Users" : "Events";
  return dAggr+" of "+entity;
}

export const isGroupByHourWindow = function(from, to) {
  let windowInSecs = to - from;
  return windowInSecs <= 86400;
}

export const getGroupByTimestampType = function(from, to) {
  // group by hour if window is <= 24hrs.
  return isGroupByHourWindow(from, to) ? 'hour' : 'date';
}

export const readableDateRange = function(range) {
  // Use label for default date range.
  if(range.startDate ==  DEFAULT_DATE_RANGE.startDate 
    && range.endDate == DEFAULT_DATE_RANGE.endDate)
    return DEFAULT_DATE_RANGE.label;

  return moment(range.startDate).format('MMM DD, YYYY') + " - " +
    moment(range.endDate).format('MMM DD, YYYY');
}

export const sameDay = function(d1, d2) {
  return d1.getFullYear() == d2.getFullYear() && d1.getMonth() == d2.getMonth() && d1.getDate() == d2.getDate();
}

export const getQueryPeriod = function(selectedRange, timezone)  {
  if (!selectedRange) {
    console.error("Invalid selected date range. Failed to get query period.");
    return
  }

  if (!timezone){
    timezone=moment.tz.guess();
    console.error("no timezone provided, default to ", timezone)
  }

  let isValidTimezone = mt.tz(timezone);
  if(!isValidTimezone){
    console.error("Invalid timezone: ", timezone,", default to browser timezone: ", timezone )
  }
  
  let isTzEndDateToday = mt(selectedRange.endDate).tz(timezone).isSame(mt().tz(timezone), 'day');
  let from =  overwriteTimezone(selectedRange.startDate, timezone).unix();
  let to = overwriteTimezone(selectedRange.endDate, timezone).unix();

  if (selectedRange.label){
    let slideToCurrentTime = DEFAULT_TODAY_DATE_RANGES.find(range => range.label === selectedRange.label);
    if (slideToCurrentTime){
      let timediff = to-from;
      to = moment(new Date()).unix();
      from = to-timediff;
      return { from: from, to: to };
    }
  }

  // Adjust the duration window respective to current time.
  if (isTzEndDateToday) {
    let newRange = slideUnixTimeWindowToCurrentTime(from, to)
    from = newRange.from;
    to = newRange.to;
  } else {
    //moves timestamp to end of the day
    to = moment.unix(to).tz(timezone).endOf("Day").unix()
  }

  // in utc.
  return { from: from, to: to }
}

//overwrites only the timezone on a given date
export const overwriteTimezone=(date, timezone)=>{
  let dateStr=moment(date).format("YYYY-MM-DD HH:mm:ss")
  return mt.tz(dateStr, timezone)
}

export const convertFunnelResultForTable = function(result) {
  let headers = result.headers;
  let rows = result.rows;
  let query = result.meta.query;

  // convert headers to readable.
  for(let i=0; i<headers.length; i++) {
    let newHeader = '';

    if (headers[i].indexOf('step_') == 0) {
      let headerSplit = headers[i].split('_');
      if (headerSplit.length < 2) continue;
      let index = parseInt(headerSplit[1]);

      newHeader = query.ewp[index].na;
      if (index > 0) {
        let prevIndex = index-1;
        if (query.ewp[prevIndex])
          newHeader = query.ewp[prevIndex].na + " to " + query.ewp[index].na;
        else
          console.error("No event name available for index ", prevIndex);
      }
    }

    if (headers[i].indexOf('conversion_') == 0) {
      if (headers[i] == 'conversion_overall') {
        newHeader = 'Total Conversion Rate'
      } else {
        let conversionSplit = headers[i].split('_');
        if (conversionSplit.length < 5) continue;

        let stepXIndex = parseInt(conversionSplit[2]),
        stepYIndex = parseInt(conversionSplit[4]);
        newHeader = query.ewp[stepXIndex].na + " to " + query.ewp[stepYIndex].na + " conversion rate";
      }
    }

    if (newHeader != '') headers[i] = newHeader;
  }

  // replace $no_group with $overall.
  for (let i=0; i<rows.length; i++) {
    for (let j=0; j<rows[i].length; j++) {
      if (rows[i][j] == '$no_group') {
        rows[i][j] = '$overall';
      }
    }
  }

  return result
}

export const convertSecondsToHMSAgo = function(timeInSeconds) {
  let hours = Math.floor(timeInSeconds / 3600);
  let minutes = Math.floor((timeInSeconds % 3600) / 60);

  if (timeInSeconds < 60) {
    return "Just Now";
  } else if (hours == 0) {
    return minutes + "m ago";
  }
  return hours + "h ago";
}

export const getPresetLabelForDateRange = function(range) {
  for (let i = 0; i < WEB_ANALYTICS_DEFINED_DATE_RANGES.length; i++) {
    let preset = WEB_ANALYTICS_DEFINED_DATE_RANGES[i]
    let presetRange = preset.range()
    if (areSameDateRanges(range, presetRange)) {
      return preset.label
    }
  }
  return null
}

// Set's start and end date for the passed dateRange object if label matches any preset.
// Else dateRange object remains unchanged.
export const setDateRangeForPresetLabel = function(dateRangeWithLabel) {
  for (let i = 0; i < DEFINED_DATE_RANGES.length; i++) {
    let preset = DEFINED_DATE_RANGES[i]
    let presetRange = preset.range()
    
    if (preset.label == dateRangeWithLabel.label) {
      if ((dateRangeWithLabel.label == DATE_RANGE_TODAY_LABEL && moment(dateRangeWithLabel.startDate).unix() != moment(presetRange.startDate).unix()) ||
        dateRangeWithLabel.label != DATE_RANGE_TODAY_LABEL && !areSameDateRanges(dateRangeWithLabel, presetRange)) {
        dateRangeWithLabel.startDate = presetRange.startDate
        dateRangeWithLabel.endDate = presetRange.endDate
        return true
      }
    }
  }
  return false
}

export const areSameDateRanges = function(dateRange1, dateRange2) {
  return moment(dateRange1.startDate).unix() == moment(dateRange2.startDate).unix() &&
    moment(dateRange1.endDate).unix() == moment(dateRange2.endDate).unix()
}