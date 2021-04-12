import React from 'react';
import moment from 'moment';
import { QUERY_TYPE_CAMPAIGN, QUERY_TYPE_EVENT } from './constants';

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

const SortDataByDate = (arr, key, order) => {
  const result = [...arr];
  result.sort((a, b) => {
    if (order === 'ascend') {
      return moment(a[key]).utc().unix() >= moment(b[key]).utc().unix()
        ? 1
        : -1;
    }
    if (order === 'descend') {
      return moment(a[key]).utc().unix() <= moment(b[key]).utc().unix()
        ? 1
        : -1;
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
    if (order === 'ascend') {
      return parseFloat(a[key]) >= parseFloat(b[key]) ? 1 : -1;
    }
    if (order === 'descend') {
      return parseFloat(a[key]) <= parseFloat(b[key]) ? 1 : -1;
    }
    return 0;
  });
  return result;
};

export const getTitleWithSorter = (
  title,
  key,
  currentSorter,
  handleSorting
) => {
  return (
    <div className='flex items-center justify-between'>
      <div className='mr-2 break-all'>{title}</div>
      <div className='flex flex-col items-center'>
        {currentSorter.key === key && currentSorter.order === 'ascend' ? (
          <div
            onClick={() => handleSorting({})}
            style={{ marginBottom: '1px' }}
          >
            <svg
              width='10'
              height='6'
              viewBox='0 0 10 6'
              fill='none'
              xmlns='http://www.w3.org/2000/svg'
            >
              <path
                d='M5.35117 1.3516L8.39362 4.36853C8.61554 4.58859 8.45971 4.96706 8.14718 4.96706H5.00002H1.83533C1.52237 4.96706 1.36673 4.58769 1.58953 4.36789L4.64797 1.3507C4.84302 1.15827 5.15661 1.15867 5.35117 1.3516Z'
                fill='#0E2647'
                stroke='#0E2647'
              />
            </svg>
          </div>
        ) : (
          <div
            onClick={() => handleSorting({ key, order: 'ascend' })}
            style={{ marginBottom: '1px' }}
          >
            <svg
              width='9'
              height='6'
              viewBox='0 0 9 6'
              fill='none'
              xmlns='http://www.w3.org/2000/svg'
            >
              <path
                d='M5.35117 1.3516L8.39362 4.36853C8.61554 4.58859 8.45971 4.96706 8.14718 4.96706H5.00002H1.83533C1.52237 4.96706 1.36673 4.58769 1.58953 4.36789L4.64797 1.3507C4.84302 1.15827 5.15661 1.15867 5.35117 1.3516Z'
                stroke='#0E2647'
              />
            </svg>
          </div>
        )}
        {currentSorter.key === key && currentSorter.order === 'descend' ? (
          <div onClick={() => handleSorting({})} style={{ marginTop: '1px' }}>
            <svg
              width='11'
              height='7'
              viewBox='0 0 11 7'
              fill='none'
              xmlns='http://www.w3.org/2000/svg'
            >
              <path
                d='M6.35165 5.65415L10.3949 1.63808C10.6165 1.41794 10.4606 1.03976 10.1482 1.03976H6.00002H1.83368C1.52095 1.03976 1.36521 1.41866 1.58754 1.63859L5.64766 5.65488C5.84276 5.84787 6.15695 5.84755 6.35165 5.65415Z'
                fill='#0E2647'
                stroke='#0E2647'
              />
            </svg>
          </div>
        ) : (
          <div
            onClick={() => handleSorting({ key, order: 'descend' })}
            style={{ marginTop: '1px' }}
          >
            <svg
              width='9'
              height='6'
              viewBox='0 0 9 6'
              fill='none'
              xmlns='http://www.w3.org/2000/svg'
            >
              <path
                d='M4.64721 4.65009L1.60105 1.59899C1.38084 1.37842 1.53706 1.0017 1.84874 1.0017L4.99996 1.0017L8.1721 1.00172C8.4843 1.00172 8.64028 1.3795 8.41903 1.59976L5.3538 4.65118C5.1583 4.8458 4.84211 4.84532 4.64721 4.65009Z'
                stroke='#0E2647'
              />
            </svg>
          </div>
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

export const getChartTypeMenuItems = (queryType, hasBreakdown) => {
  let menuItems = [];
  if (queryType === QUERY_TYPE_EVENT) {
    if (hasBreakdown) {
      menuItems = [
        {
          key: 'barchart',
          name: 'Barchart',
        },
        {
          key: 'linechart',
          name: 'Line Chart',
        },
        {
          key: 'stackedareachart',
          name: 'Stacked Area Chart',
        },
      ];
    } else {
      menuItems = [
        {
          key: 'sparklines',
          name: 'Sparkline',
        },
        {
          key: 'linechart',
          name: 'Line Chart',
        },
      ];
    }
  }
  if (queryType === QUERY_TYPE_CAMPAIGN) {
    if (hasBreakdown) {
      menuItems = [
        {
          key: 'barchart',
          name: 'Barchart',
        },
        {
          key: 'linechart',
          name: 'Line Chart',
        },
        {
          key: "stackedareachart",
          name: "Stacked Area Chart",
        },
      ];
    } else {
      menuItems = [
        {
          key: 'sparklines',
          name: 'Sparkline',
        },
        {
          key: 'linechart',
          name: 'Line Chart',
        },
      ];
    }
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
