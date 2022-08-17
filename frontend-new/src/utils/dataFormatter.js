import React from 'react';
import { escapeRegExp, isArray } from 'lodash';
import moment from 'moment';
import { ArrowDownOutlined, ArrowUpOutlined } from '@ant-design/icons';
import { Tooltip } from 'antd';
import {
  QUERY_TYPE_EVENT,
  CHART_TYPE_STACKED_AREA,
  CHART_TYPE_LINECHART,
  CHART_TYPE_STACKED_BAR,
  CHART_TYPE_SPARKLINES,
  PREDEFINED_DATES,
  DATE_FORMATS
} from './constants';
import { Text } from '../components/factorsComponents';

export const visualizationColors = [
  '#4D7DB4',
  '#4C9FC8',
  '#4CBCBD',
  '#86D3A3',
  '#CCC36D',
  '#F9C06E',
  '#E89E7B',
  '#D4787D',
  '#B87B7E',
  '#9982B5'
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
    let val1 = typeof a[key] === 'object' ? a[key]?.value : a[key];
    let val2 = typeof b[key] === 'object' ? b[key]?.value : b[key];

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
    const val1 = moment(a[key].split(' to ')[0], DATE_FORMATS.day).utc().unix();
    const val2 = moment(b[key].split(' to ')[0], DATE_FORMATS.day).utc().unix();
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
  alignment = 'left',
  verticalAlignment = 'center',
  containerClassName,
  titleTooltip = null
) => {
  const sorter = isArray(currentSorter) ? currentSorter : [currentSorter];
  const sorterPropIndex = sorter.findIndex(
    (elem) => elem.key === sorterProp.key
  );

  let titleText;

  if (titleTooltip) {
    titleText = (
      <Tooltip title={titleTooltip}>
        <Text weight="bold" color="grey-2" type="title" extraClass="mb-0">
          {title}
        </Text>
      </Tooltip>
    );
  } else {
    titleText = (
      <Text weight="bold" color="grey-2" type="title" extraClass="mb-0">
        {title}
      </Text>
    );
  }

  const icon = (
    <>
      {sorterPropIndex > -1 && sorter[sorterPropIndex].order === 'descend' ? (
        <ArrowDownOutlined />
      ) : null}
      {sorterPropIndex > -1 && sorter[sorterPropIndex].order === 'ascend' ? (
        <ArrowUpOutlined />
      ) : null}
    </>
  );

  const justifyAlignment =
    alignment === 'left' ? 'justify-start' : 'justify-end';
  const verticalAlignmentClass =
    verticalAlignment === 'start'
      ? 'items-start'
      : verticalAlignment === 'end'
      ? 'items-end'
      : 'items-center';

  return (
    <div
      role="button"
      tabIndex={0}
      onClick={() => handleSorting(sorterProp)}
      className={`flex ${verticalAlignmentClass} ${justifyAlignment} cursor-pointer h-full px-4 py-2 ${containerClassName}`}
    >
      <div className="flex gap-x-1 items-center">
        {alignment === 'left' ? (
          <>
            {titleText}
            {icon}
          </>
        ) : (
          <>
            {icon}
            {titleText}
          </>
        )}
      </div>
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

export const getValidGranularityOptionsFromDaysDiff = ({ daysDiff }) => {
  if (daysDiff > 1) {
    return ['date', 'week', 'month', 'quarter'];
  }
  return ['hour'];
};

export const getValidGranularityOptions = ({ from, to }) => {
  const startDate = moment(from).startOf('day').utc().unix() * 1000;
  const endDate = moment(to).endOf('day').utc().unix() * 1000 + 1000;
  const daysDiff = moment(endDate).diff(startDate, 'days');
  return getValidGranularityOptionsFromDaysDiff({ daysDiff });
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
        message: "Attribution reports don't show data for today"
      };
    }
    if (moment().format('D') === '2') {
      return {
        required: false,
        message: "Attribution reports don't show data for yesterday"
      };
    }
    return {
      required: true,
      message:
        'Attribution reports for "This Month" cover data till the day before yesterday.'
    };
  }
  if (durationObj.dateType === PREDEFINED_DATES.THIS_WEEK) {
    if (moment().format('dddd') === 'Sunday') {
      return {
        required: false,
        message: "Attribution reports don't show data for today"
      };
    }
    if (moment().format('dddd') === 'Monday') {
      return {
        required: false,
        message: "Attribution reports don't show data for yesterday"
      };
    }
    return {
      required: true,
      message:
        'Attribution reports for "This Week" cover data till the day before yesterday.'
    };
  }
  if (durationObj.dateType === PREDEFINED_DATES.TODAY) {
    return {
      required: false,
      message: "Attribution reports don't show data for today"
    };
  }
  if (durationObj.dateType === PREDEFINED_DATES.YESTERDAY) {
    return {
      required: false,
      message: "Attribution reports don't show data for yesterday"
    };
  }
  if (durationObj.dateType === PREDEFINED_DATES.LAST_MONTH) {
    if (moment().format('D') === '1') {
      return {
        required: true,
        message:
          'Attribution reports for "Last Month" cover data till the day before yesterday.'
      };
    }
  }
  if (durationObj.dateType === PREDEFINED_DATES.LAST_WEEK) {
    if (moment().format('dddd') === 'Sunday') {
      return {
        required: true,
        message:
          'Attribution reports for "Last Week" cover data till the day before yesterday.'
      };
    }
  }
  return {
    required: true,
    message: null
  };
};

export const getNewSorterState = (currentSorter, newSortProp) => {
  const newSortPropIndex = currentSorter.findIndex(
    (elem) => elem.key === newSortProp.key
  );

  // if user is sorting by a numerical column which is not already in use for sorting then we will just sort by this column
  if (newSortPropIndex === -1 && newSortProp.type === 'numerical') {
    return [
      {
        ...newSortProp,
        order: 'descend'
      }
    ];
  }

  if (currentSorter.length === 3) {
    // we already have three levels of sorting and user has applied sorting on fourth column then we will reset the sorting state and only kepp the newly selected one
    if (newSortPropIndex === -1) {
      return [
        {
          ...newSortProp,
          order: 'descend'
        }
      ];
    } else {
      // we are editing existing level of sorting here
      if (currentSorter[newSortPropIndex].order === 'ascend') {
        return [
          ...currentSorter.slice(0, newSortPropIndex),
          ...currentSorter.slice(newSortPropIndex + 1)
        ];
      } else {
        return [
          ...currentSorter.slice(0, newSortPropIndex),
          { ...newSortProp, order: 'ascend' },
          ...currentSorter.slice(newSortPropIndex + 1)
        ];
      }
    }
  } else {
    if (newSortPropIndex === -1) {
      // we are inserting new level of sorting here
      return [
        {
          ...newSortProp,
          order: 'descend'
        },
        ...currentSorter
      ];
    } else {
      // we are editing existing level of sorting here
      if (currentSorter[newSortPropIndex].order === 'ascend') {
        return [
          ...currentSorter.slice(0, newSortPropIndex),
          ...currentSorter.slice(newSortPropIndex + 1)
        ];
      } else {
        return [
          ...currentSorter.slice(0, newSortPropIndex),
          { ...newSortProp, order: 'ascend' },
          ...currentSorter.slice(newSortPropIndex + 1)
        ];
      }
    }
  }
};

export const SortByKey = (result, currentSorter) => {
  if (currentSorter.key) {
    if (currentSorter.type === 'datetime') {
      if (
        currentSorter.subtype === 'day' ||
        currentSorter.subtype === 'date' ||
        currentSorter.subtype === 'month' ||
        currentSorter.subtype === 'hour' ||
        currentSorter.subtype === 'quarter'
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

export const SortResults = (result, sortSelections) => {
  if (!Array.isArray(sortSelections) || !sortSelections.length) {
    return result;
  }
  const firstSortedResult = SortByKey(result, sortSelections[0]);
  if (sortSelections.length === 1) {
    return firstSortedResult;
  }

  const key1 = sortSelections[0].key;
  let secondSortedResult = [];
  let i = 0;
  let j;

  while (i < firstSortedResult.length) {
    const key1Value = firstSortedResult[i][key1];
    const elemsWithSameValueForKey1 = [firstSortedResult[i]];
    j = i + 1;
    while (j < firstSortedResult.length) {
      if (firstSortedResult[j][key1] !== key1Value) {
        break;
      }
      elemsWithSameValueForKey1.push(firstSortedResult[j]);
      j++;
    }
    if (elemsWithSameValueForKey1.length === 1) {
      secondSortedResult.push(elemsWithSameValueForKey1[0]);
    } else {
      secondSortedResult = secondSortedResult.concat(
        SortByKey(elemsWithSameValueForKey1, sortSelections[1])
      );
    }
    i = j;
  }

  if (sortSelections.length === 2) {
    return secondSortedResult;
  }

  const key2 = sortSelections[1].key;
  let thirdSortedResult = [];
  i = 0;

  while (i < secondSortedResult.length) {
    const key1Value = secondSortedResult[i][key1];
    const key2Value = secondSortedResult[i][key2];
    const elemsWithSameValueForKey1AndKey2 = [secondSortedResult[i]];
    j = i + 1;
    while (j < secondSortedResult.length) {
      if (
        secondSortedResult[j][key1] !== key1Value ||
        secondSortedResult[j][key2] !== key2Value
      ) {
        break;
      }
      elemsWithSameValueForKey1AndKey2.push(secondSortedResult[j]);
      j++;
    }
    if (elemsWithSameValueForKey1AndKey2.length === 1) {
      thirdSortedResult.push(elemsWithSameValueForKey1AndKey2[0]);
    } else {
      thirdSortedResult = thirdSortedResult.concat(
        SortByKey(elemsWithSameValueForKey1AndKey2, sortSelections[2])
      );
    }
    i = j;
  }

  return thirdSortedResult;
};

export const formatFilterDate = (selectedDates) => {
  const parsedVal = JSON.parse(selectedDates);
  const fromDateKey = parsedVal.fr ? 'fr' : 'from';
  const toDateKey = 'to';
  const fromDate = parsedVal[fromDateKey];
  const toDate = parsedVal[toDateKey];
  const convertedKeys = {};
  if (fromDate) {
    const fr = isDateInMilliSeconds(fromDate)
      ? moment(fromDate).utc().unix()
      : fromDate;
    convertedKeys[fromDateKey] = fr;
  }
  if (toDate) {
    const to = isDateInMilliSeconds(toDate)
      ? moment(toDate).utc().unix()
      : toDate;
    convertedKeys[toDateKey] = to;
  }

  const convertedVal = {
    ...parsedVal,
    ...convertedKeys
  };
  return JSON.stringify(convertedVal);
};

export function isDateInMilliSeconds(date) {
  return date?.toString().length === 13;
}

export const Wait = (duration) => {
  return new Promise((resolve) => {
    setTimeout(() => {
      resolve();
    }, duration);
  });
};

export const toLetters = (num) => {
  const charArr = ['A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J'];
  return charArr[num];
};

export const PropTextFormat = (prop = 'users') => {
  const formatText = prop.replace('$', '').split('_');
  formatText.forEach((word, i) => {
    formatText[i] = formatText[i][0].toUpperCase() + formatText[i].substr(1);
  });
  return formatText.join(' ');
};

export const HighlightSearchText = ({ text = '', highlight = '' }) => {
  if (!highlight.trim()) {
    return <span>{text}</span>;
  }
  const regex = new RegExp(`(${escapeRegExp(highlight)})`, 'gi');
  const parts = text.split(regex);
  return (
    <span className={'truncate'}>
      {parts.map((part, i) =>
        regex.test(part) ? <b key={i}>{part}</b> : <span key={i}>{part}</span>
      )}
    </span>
  );
};

export const addQforQuarter = (freq) => {
  return freq === 'quarter' ? 'Q' : '';
};

export const formatDurationIntoString = (seconds) => {
  let returnString = '',
    i = 0,
    stringLength = 0;
  if (seconds > 0) {
    let timeUnits = [
      [Math.floor(seconds / 31536000), 'years'],
      [Math.floor((seconds % 31536000) / 2592000), 'months'],
      [Math.floor(((seconds % 31536000) % 2592000) / 604800), 'weeks'],
      [Math.floor((((seconds % 31536000) % 2592000) % 604800) / 86400), 'days'],
      [Math.floor(((seconds % 31536000) % 86400) / 3600), 'hours'],
      [Math.floor((((seconds % 31536000) % 86400) % 3600) / 60), 'minutes'],
      [(((seconds % 31536000) % 86400) % 3600) % 60, 'seconds']
    ];
    while (i < timeUnits.length && stringLength < 4) {
      if (timeUnits[i][0] === 0) {
        i++;
        continue;
      }
      returnString +=
        ' ' +
        parseInt(timeUnits[i][0]) +
        ' ' +
        (timeUnits[i][0] === 1
          ? timeUnits[i][1].substr(0, timeUnits[i][1].length - 1)
          : timeUnits[i][1]);
      i++;
      stringLength = returnString.split(' ').length;
    }
  } else return '0 seconds';
  return returnString.trim();
};

export const abbreviateNumber = (n) => {
  if (n < 1e3) return n;
  if (n >= 1e3 && n < 1e6) return +(n / 1e3).toFixed(1) + 'K';
  if (n >= 1e6 && n < 1e9) return +(n / 1e6).toFixed(1) + 'M';
  if (n >= 1e9 && n < 1e12) return +(n / 1e9).toFixed(1) + 'B';
  if (n >= 1e12) return +(n / 1e12).toFixed(1) + 'T';
};
