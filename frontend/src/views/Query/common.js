import { createStaticRanges } from 'react-date-range';
import moment from 'moment';
import mt from "moment-timezone";

import {slideUnixTimeWindowToCurrentTime, firstToUpperCase, isNumber} from '../../util';

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

export const NUMERICAL_GROUP_BY_BUCKETED = 'bucketed';
export const NUMERICAL_GROUP_BY_RAW = 'raw';
export const NUMERICAL_GROUP_BY_METHODS = {
  'bucketed': 'with buckets',
  'raw': 'raw values'
};

export const DASHBOARD_TYPE_WEB_ANALYTICS = "Website Analytics";
export const QUERY_CLASS_CHANNEL = "channel";
export const QUERY_CLASS_FUNNEL = "funnel";
export const QUERY_CLASS_ATTRIBUTION = 'attribution';
export const QUERY_CLASS_WEB = 'web';
export const QUERY_CLASS_INSIGHTS = 'insights';
export const QUERY_CLASS_EVENTS = 'events';
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

export const DATE_RANGE_LABEL_CURRENT_MONTH = 'Current Month';
export const DEFAULT_DATE_RANGE_LABEL = 'Current Week';
export const DATE_RANGE_LABEL_LAST_MONTH = 'Last Month';
export const DATE_RANGE_LABEL_LAST_WEEK = 'Last Week';
export const DATE_RANGE_YESTERDAY_LABEL = 'Yesterday';
export const DATE_RANGE_TODAY_LABEL = 'Today';
export const DATE_RANGE_LAST_2_MIN_LABEL = 'Last 2 mins'
export const DATE_RANGE_LAST_30_MIN_LABEL = 'Last 30 mins'

export const LABEL_STYLE = { marginRight: '10px', fontWeight: '600', color: '#777' };

export const DEFAULT_DATE_RANGE = {
  ...(!isTodayTheFirstDayOfWeek() && {
    startDate: moment(getFirstDayOfCurrentWeek()).startOf('day').toDate(),
    endDate: moment(new Date()).subtract(1, 'days').endOf('day').toDate(),
  }),
  ...(isTodayTheFirstDayOfWeek() && {
    startDate: moment(new Date()).startOf('day').toDate(),
    endDate: new Date(),
  }),
  label: DEFAULT_DATE_RANGE_LABEL,
  key: 'selected'
}

function getFirstDayOfCurrentWeek() {
  let d = new Date();
  let first = d.getDate() - d.getDay()
  return new Date(d.setDate(first));
}

function getFirstDayOfLastWeek() {
  let d = new Date();
  let first = d.getDate() - d.getDay() - 7;
  return new Date(d.setDate(first));
}

function getLastDayOfLastWeek() {
  let d = new Date();
  let last = d.getDate() - d.getDay() - 1;
  return new Date(d.setDate(last));
}

function getFirstDayOfLastMonth() {
  let d = new Date();
  return new Date(d.getFullYear(), d.getMonth() - 1, 1);
}

function getLastDayOfLastMonth() {
  let d = new Date();
  return new Date(d.getFullYear(), d.getMonth(), 0);
}

function getFirstDayOfCurrentMonth() {
  let d = new Date();
  return new Date(d.getFullYear(), d.getMonth(), 1);
}

function isTodayTheFirstDayOfMonth() {
  let d = new Date();
  return d.getDate() === 1;
}

function isTodayTheFirstDayOfWeek() {
  // week starts with Sunday.
  let d = new Date();
  return d.getDay() === 0;
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
    ...(!isTodayTheFirstDayOfWeek() && {
      range: () => ({
      startDate: moment(getFirstDayOfCurrentWeek()).startOf('day').toDate(),
      endDate: moment(new Date()).subtract(1, 'days').endOf('day').toDate(),
    })}),
    ...(isTodayTheFirstDayOfWeek() && {
      range: () => ({
        startDate: moment(new Date()).startOf('day').toDate(),
        endDate: new Date(),
      })})
  },
  {
    label: DATE_RANGE_LABEL_CURRENT_MONTH,
    ...(!isTodayTheFirstDayOfMonth() && {
      range: () => ({
        startDate: moment(getFirstDayOfCurrentMonth()).startOf('day').toDate(),
        endDate: moment(new Date()).subtract(1, 'days').endOf('day').toDate(),
      })}),
    ...(isTodayTheFirstDayOfMonth() && {
      range: () => ({
        startDate: moment(new Date()).startOf('day').toDate(),
        endDate: new Date(),
      })}),
  },
  {
    label: DATE_RANGE_LABEL_LAST_WEEK,
    range: () => ({
      startDate: moment(getFirstDayOfLastWeek()).startOf('day').toDate(),
      endDate: moment(getLastDayOfLastWeek()).endOf('day').toDate(),
    }),
  },
  {
    label: DATE_RANGE_LABEL_LAST_MONTH,
    range: () => ({
      startDate: moment(getFirstDayOfLastMonth()).startOf('day').toDate(),
      endDate: moment(getLastDayOfLastMonth()).endOf('day').toDate(),
    }),
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
export const jsonToCSV = (result, selectedPresentation, queryName) => {
  const csvRows = [];
  let newJSON = {...result}
  if (selectedPresentation === PRESENTATION_LINE) {
    newJSON = convertLineJSON(newJSON)
  }
  let headers = [...newJSON.headers]
  csvRows.push(headers.join(','));
  let rows = [...newJSON.rows]
  let jsonRows = rows.map((row)=> {
    let newRow = [...row]
    let values = newRow.map((val)=>{
      const escaped = (''+val).replace(/"/g,'\\"');
      return `"${escaped}"`
    })
    return values.join(',');
  });
  jsonRows = jsonRows.join('\n')
  const csv = csvRows+ "\n"+ jsonRows
  return downloadCSV(csv, queryName);
}

export const downloadCSV = function(data, queryName) {
  const blob = new Blob([data], {type: 'text/csv'})
  const url = window.URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.setAttribute('hidden','')
  a.setAttribute('href', url)
  a.setAttribute('download', queryName+new Date()+'.csv')
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
}

export const convertLineJSON = function(data) {

  let convertedData = {}
  let datetimeKey = 0;
  let headers = [...data.headers]
  headers[0]= "event_name"
  for(var i = 1; i<=data.meta.query.gbp.length; i++) {
    headers[i] = data.meta.query.gbp[i-1].pr
  }
  for (var i=0; i<headers.length; i++) {
    if(headers[i] == "datetime") {
      datetimeKey = i;
      if(data.meta.query.gbt == "date"){
        headers[i] = "date(UTC)"
      } else {
        headers.splice(i,1,"date(UTC)", "hour")
      }

    }
  }
  convertedData.headers = headers
  let rows = [...data.rows]
  convertedData.rows = rows.map((row)=> {
    let dateTime= row[datetimeKey].split("T")
    if(data.meta.query.gbt == "date"){
      row[datetimeKey] = dateTime[0]
    }
    else {
      let time = (dateTime[1].split("+"))[0]
      row.splice(datetimeKey, 1, dateTime[0],time)
    }
    return row
  })

  return convertedData
}

export const getEventsWithProperties = function(events) {

  let ewps = [];
  for (let ei = 0; ei < events.length; ei++) {
    let event = events[ei];
    if (event.name === "")
      continue;
    let ewp = getEventWithProperties(event);
    ewps.push(ewp)
  }
  return ewps;
}

export const getEventWithProperties = function(event) {

  let ewp = {};
  if (event.name === "")
    return ewp;
  ewp.na = event.name;
  ewp.pr = [];

  for (let pi = 0; pi < event.properties.length; pi++) {
    let property = event.properties[pi];
    let cProperty = {}

    if (property.entity !== '' && property.name !== '' &&
      property.operator !== '' && property.value !== '' &&
      property.valueType !== '') {

      if (property.valueType === 'numerical' && !isNumber(property.value))
        continue;

      cProperty.en = property.entity;
      cProperty.pr = property.name;
      cProperty.op = property.op;
      cProperty.va = property.value;
      cProperty.ty = property.valueType;
      cProperty.lop = property.logicalOp;

      // update datetime with current time window if ovp is true.
      if (property.valueType === PROPERTY_VALUE_TYPE_DATE_TIME) {
        let dateRange = JSON.parse(cProperty.va);
        if (dateRange.ovp) {
          let newRange = slideUnixTimeWindowToCurrentTime(dateRange.fr, dateRange.to);
          dateRange.fr = newRange.from;
          dateRange.to = newRange.to;
          cProperty.va = JSON.stringify(dateRange);
        }
      }
      ewp.pr.push(cProperty);
    }
  }
  return ewp
}
