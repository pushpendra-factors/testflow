import React from 'react';
import moment from 'moment';
import {
  QUERY_TYPE_CAMPAIGN,
  QUERY_TYPE_EVENT,
  CHART_TYPE_STACKED_AREA,
  CHART_TYPE_LINECHART,
  CHART_TYPE_STACKED_BAR,
  CHART_TYPE_SPARKLINES,
  PREDEFINED_DATES,
  DATE_FORMATS,
  QUERY_TYPE_ATTRIBUTION,
  CHART_TYPE_BARCHART,
  CHART_TYPE_SCATTER_PLOT,
  QUERY_TYPE_FUNNEL,
  CHART_TYPE_HORIZONTAL_BAR_CHART,
} from './constants';
import { ArrowDownOutlined, ArrowUpOutlined } from '@ant-design/icons';

const visualizationColors = [
  '#4D7DB4',
  '#4C9FC8',
  '#4CBCBD',
  '#86D3A3',
  '#CCC36D',
  '#F9C06E',
  '#E89E7B',
  '#D4787D',
  '#B87B7E',
  '#9982B5',
];

export const numberWithCommas = (x) => {
  return x.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ',');
};

export const calculatePercentage = (numerator, denominator, precision = 1) => {
  if (!denominator) {
    return 0;
  }
  const result = (numerator / denominator) * 100;
  return result % 1 !== 0 ? result.toFixed(precision) : result;
};

export const getDurationInSeconds = (duration) => {
  if (duration.indexOf(' ') === -1) {
    return duration.split('s')[0];
  } else {
    if (duration.indexOf('d') > -1) {
      const dayStr = duration.split(' ')[0];
      const hourStr = duration.split(' ')[1];
      const days = Number(dayStr.split('d')[0]);
      const hours = Number(hourStr.split('h')[0]);
      return days * 86400 + hours * 3600;
    }
    if (duration.indexOf('h') > -1) {
      const hourStr = duration.split(' ')[0];
      const minsStr = duration.split(' ')[1];
      const hours = Number(hourStr.split('h')[0]);
      const minutes = Number(minsStr.split('m')[0]);
      return hours * 3600 + minutes * 60;
    }
    if (duration.indexOf('m') > -1) {
      const minsStr = duration.split(' ')[0];
      const secondStr = duration.split(' ')[1];
      const mins = Number(minsStr.split('m')[0]);
      const seconds = Number(secondStr.split('s')[0]);
      return mins * 60 + seconds;
    }
  }
  return 0;
};

export const SortDataByDuration = (arr, key, order) => {
  const result = [...arr];
  result.sort((a, b) => {
    const val1 = getDurationInSeconds(a[key]);
    const val2 = getDurationInSeconds(b[key]);
    if (order === 'ascend') {
      return Number(val1) >= Number(val2) ? 1 : -1;
    }
    if (order === 'descend') {
      return Number(val1) <= Number(val2) ? 1 : -1;
    }
    return 0;
  });
  return result;
};

export const SortDataByDate = (arr, key, order, format = null) => {
  const result = [...arr];
  result.sort((a, b) => {
    const val1 = format
      ? moment(a[key], format).utc().unix()
      : moment(a[key]).utc().unix();
    const val2 = format
      ? moment(b[key], format).utc().unix()
      : moment(b[key]).utc().unix();
    if (order === 'ascend') {
      return val1 >= val2 ? 1 : -1;
    }
    if (order === 'descend') {
      return val1 <= val2 ? 1 : -1;
    }
    return 0;
  });
  return result;
};

export const SortData = (arr, key, order) => {
  if (key === 'date') {
    return SortDataByDate(arr, key, order);
  }
  const result = [...arr];
  result.sort((a, b) => {
    // type of a[key] can be an object when the comparison is applied
    let val1 = typeof a[key] === 'object' ? a[key].value : a[key];
    let val2 = typeof b[key] === 'object' ? b[key].value : b[key];

    if (isNaN(val1)) {
      val1 = 0;
    }

    if (isNaN(val2)) {
      val2 = 0;
    }

    if (order === 'ascend') {
      return Number(val1) >= Number(val2) ? 1 : -1;
    }
    if (order === 'descend') {
      return Number(val1) <= Number(val2) ? 1 : -1;
    }
    return 0;
  });
  return result;
};

export const SortDataByAlphabets = (arr, key, order) => {
  const result = [...arr];
  result.sort((a, b) => {
    if (order === 'ascend') {
      return a[key] >= b[key] ? 1 : -1;
    }
    if (order === 'descend') {
      return a[key] <= b[key] ? 1 : -1;
    }
    return 0;
  });
  return result;
};

export const SortWeekFormattedData = (arr, key, order) => {
  const result = [...arr];
  result.sort((a, b) => {
    const val1 = moment(a[key].split(' to ')[0], DATE_FORMATS['day'])
      .utc()
      .unix();
    const val2 = moment(b[key].split(' to ')[0], DATE_FORMATS['day'])
      .utc()
      .unix();
    if (order === 'ascend') {
      return val1 >= val2 ? 1 : -1;
    }
    if (order === 'descend') {
      return val1 <= val2 ? 1 : -1;
    }
    return 0;
  });
  return result;
};

export const getClickableTitleSorter = (
  title,
  sorterProp,
  currentSorter,
  handleSorting,
  alignmentClass = 'items-center'
) => {
  return (
    <div
      onClick={() => handleSorting(sorterProp)}
      className={`flex ${alignmentClass} justify-between cursor-pointer h-full`}
    >
      <div className='mr-2 break-all'>{title}</div>
      {currentSorter.key === sorterProp.key &&
      currentSorter.order === 'descend' ? (
        <ArrowDownOutlined />
      ) : null}
      {currentSorter.key === sorterProp.key &&
      currentSorter.order === 'ascend' ? (
        <ArrowUpOutlined />
      ) : null}
    </div>
  );
};

export const generateColors = (requiredCumberOfColors) => {
  const adder = Math.floor(visualizationColors.length / requiredCumberOfColors);
  const colors = [];
  for (let i = 0; i < requiredCumberOfColors; i++) {
    colors.push(visualizationColors[(i * adder) % 10]);
  }
  return colors;
};

export const formatCount = (count, precision) => {
  try {
    return count % 1 !== 0 ? count.toFixed(precision) : count;
  } catch (err) {
    return count;
  }
};

export const getChartTypeMenuItems = (queryType, breakdownLength, events) => {
  let menuItems = [];
  if (queryType === QUERY_TYPE_EVENT || queryType === QUERY_TYPE_CAMPAIGN) {
    if (breakdownLength) {
      menuItems = [
        {
          key: CHART_TYPE_BARCHART,
          name: 'Columns',
        },
        {
          key: CHART_TYPE_LINECHART,
          name: 'Line Chart',
        },
        {
          key: CHART_TYPE_STACKED_AREA,
          name: 'Stacked Area',
        },
        {
          key: CHART_TYPE_STACKED_BAR,
          name: 'Stacked Column',
        },
      ];
      if (
        events.length === 1 &&
        queryType === QUERY_TYPE_EVENT &&
        breakdownLength <= 3
      ) {
        // this chart type is only supported when there is atmost one event and there is atleast 1 breakdown and atmost 3 breakdowns
        menuItems.push({
          key: CHART_TYPE_HORIZONTAL_BAR_CHART,
          name: 'Bars',
        });
      }
    } else {
      menuItems = [
        {
          key: CHART_TYPE_SPARKLINES,
          name: 'Sparkline',
        },
        {
          key: CHART_TYPE_LINECHART,
          name: 'Line Chart',
        },
      ];
    }
  }
  if (queryType === QUERY_TYPE_ATTRIBUTION) {
    menuItems = [
      {
        key: CHART_TYPE_BARCHART,
        name: 'Barchart',
      },
      {
        key: CHART_TYPE_SCATTER_PLOT,
        name: 'Scatter Plot',
      },
    ];
  }
  if (queryType === QUERY_TYPE_FUNNEL && breakdownLength) {
    menuItems = [
      {
        key: CHART_TYPE_BARCHART,
        name: 'Barchart',
      },
      {
        key: CHART_TYPE_SCATTER_PLOT,
        name: 'Scatter Plot',
      },
    ];
  }
  return menuItems;
};

export const formatDuration = (seconds) => {
  seconds = Number(seconds);
  if (seconds < 60) {
    return Math.floor(seconds) + 's';
  }
  if (seconds < 3600) {
    const minutes = Math.floor(seconds / 60);
    const remains = Math.floor(seconds % 60);
    return `${minutes}m ${remains}s`;
  }
  if (seconds < 86400) {
    const hours = Math.floor(seconds / 3600);
    const remains = seconds % 3600;
    const minutes = Math.floor(remains / 60);
    return `${hours}h ${minutes}m`;
  }
  const days = Math.floor(seconds / 86400);
  const remains = seconds % 86400;
  const hours = Math.floor(remains / 3600);
  return `${days}d ${hours}h`;
};

export const getErrorMessage = (err) => {
  if (err && typeof err.data === 'string') {
    return err.data;
  }

  if (err && err.data && err.data.error && typeof err.data.error === 'string') {
    return err.data.error;
  }

  return 'Something went wrong!';
};

export const setItemToLocalStorage = (key, payload) => {
  localStorage.setItem(key, payload);
};

export const getItemFromLocalStorage = (key) => {
  return localStorage.getItem(key);
};

export const removeItemFromLocalStorage = (key) => {
  localStorage.removeItem(key);
};

export const clearLocalStorage = (key) => {
  localStorage.clear();
};

export const getValidGranularityOptions = ({ from, to }, queryType) => {
  const startDate = moment(from).startOf('day').utc().unix() * 1000;
  const endDate = moment(to).endOf('day').utc().unix() * 1000 + 1000;
  const daysDiff = moment(endDate).diff(startDate, 'days');
  //whatever will be returned, 0th element will be treated as default
  if (daysDiff > 93) {
    return ['date', 'week', 'month', 'quarter'];
  }
  if (daysDiff > 31) {
    return ['date', 'week', 'month'];
  }
  if (daysDiff > 7) {
    return ['date', 'week'];
  }
  if (daysDiff > 1) {
    return ['date'];
  }
  //hourly data is not supported for campaigns
  if (queryType === QUERY_TYPE_CAMPAIGN) {
    return ['date'];
  }
  return ['hour'];
};

export const isSeriesChart = (chartType) => {
  return (
    chartType === CHART_TYPE_STACKED_AREA ||
    chartType === CHART_TYPE_LINECHART ||
    chartType === CHART_TYPE_STACKED_BAR ||
    chartType === CHART_TYPE_SPARKLINES
  );
};

export const getQueryType = (query) => {
  const cl = query.cl
    ? query.cl
    : Array.isArray(query.query_group) && query.query_group.length
    ? query.query_group[0].cl
    : QUERY_TYPE_EVENT;
  return cl;
};

export const renderBigLengthTicks = (text, allowdLength) => {
  if (text.length > allowdLength) {
    return text.slice(0, allowdLength) + '...';
  }
  return text;
};

export const shouldDataFetch = (durationObj) => {
  if (durationObj.dateType === PREDEFINED_DATES.THIS_MONTH) {
    if (moment().format('D') === '1') {
      return {
        required: false,
        message: `Attribution reports don't show data for today`,
      };
    }
    if (moment().format('D') === '2') {
      return {
        required: false,
        message: `Attribution reports don't show data for yesterday`,
      };
    }
    return {
      required: true,
      message:
        'Attribution reports for "This Month" cover data till the day before yesterday.',
    };
  }
  if (durationObj.dateType === PREDEFINED_DATES.THIS_WEEK) {
    if (moment().format('dddd') === 'Sunday') {
      return {
        required: false,
        message: `Attribution reports don't show data for today`,
      };
    }
    if (moment().format('dddd') === 'Monday') {
      return {
        required: false,
        message: `Attribution reports don't show data for yesterday`,
      };
    }
    return {
      required: true,
      message: `Attribution reports for "This Week" cover data till the day before yesterday.`,
    };
  }
  if (durationObj.dateType === PREDEFINED_DATES.TODAY) {
    return {
      required: false,
      message: `Attribution reports don't show data for today`,
    };
  }
  if (durationObj.dateType === PREDEFINED_DATES.YESTERDAY) {
    return {
      required: false,
      message: `Attribution reports don't show data for yesterday`,
    };
  }
  if (durationObj.dateType === PREDEFINED_DATES.LAST_MONTH) {
    if (moment().format('D') === '1') {
      return {
        required: true,
        message: `Attribution reports for "Last Month" cover data till the day before yesterday.`,
      };
    }
  }
  if (durationObj.dateType === PREDEFINED_DATES.LAST_WEEK) {
    if (moment().format('dddd') === 'Sunday') {
      return {
        required: true,
        message: `Attribution reports for "Last Week" cover data till the day before yesterday.`,
      };
    }
  }
  return {
    required: true,
    message: null,
  };
};

export const getNewSorterState = (currentSorter, newSortProp) => {
  if (currentSorter.key === newSortProp.key) {
    return {
      ...currentSorter,
      order: currentSorter.order === 'ascend' ? 'descend' : 'ascend',
    };
  }
  return {
    ...newSortProp,
    order: 'ascend',
  };
};

export const SortResults = (result, currentSorter) => {
  if (currentSorter.key) {
    if (currentSorter.type === 'datetime') {
      if (
        currentSorter.subtype === 'day' ||
        currentSorter.subtype === 'date' ||
        currentSorter.subtype === 'month' ||
        currentSorter.subtype === 'hour'
      ) {
        return SortDataByDate(
          result,
          currentSorter.key,
          currentSorter.order,
          DATE_FORMATS[currentSorter.subtype]
        );
      }
      if (currentSorter.subtype === 'week') {
        return SortWeekFormattedData(
          result,
          currentSorter.key,
          currentSorter.order
        );
      }
    } else if (currentSorter.type === 'categorical') {
      return SortDataByAlphabets(
        result,
        currentSorter.key,
        currentSorter.order
      );
    } else if (currentSorter.type === 'duration') {
      return SortDataByDuration(result, currentSorter.key, currentSorter.order);
    } else {
      return SortData(result, currentSorter.key, currentSorter.order);
    }
  }
  return result;
};

export const getBreakdownDisplayTitle = (
  breakdown,
  userPropNames,
  eventPropNames
) => {
  let displayTitle =
    breakdown.en === 'user'
      ? userPropNames[breakdown.pr] || breakdown.pr
      : breakdown.en === 'event'
      ? eventPropNames[breakdown.pr] || breakdown.pr
      : breakdown.pr;

  if (breakdown.eventIndex) {
    displayTitle = displayTitle + ' (event)';
  }
  return displayTitle;
};

export const Wait = (duration) => {
  return new Promise((resolve) => {
    setTimeout(() => {
      resolve();
    }, duration);
  });
};
