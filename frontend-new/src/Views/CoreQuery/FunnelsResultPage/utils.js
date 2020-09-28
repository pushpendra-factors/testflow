/* eslint-disable */
import React from 'react';
import tableStyles from './FunnelsResultTable/index.module.scss';
import { funnelsDataWithoutBreakdown } from '../../EventsAnalytics/SampleResponse';

const windowSize = {
  w: window.outerWidth,
  h: window.outerHeight,
  iw: window.innerWidth,
  ih: window.innerHeight
};

const visualizationColors = ['#4D7DB4', '#4C9FC8', '#4CBCBD', '#86D3A3', '#CCC36D', '#F9C06E', '#E89E7B', '#D4787D', '#B87B7E', '#9982B5'];

export const generateGroupedChartsData = (data, groups) => {
  const displayedData = data.filter(elem => elem.display);
  const result = displayedData.map(elem => {
    const values = [];
    for (const key in elem.data) {
      const group = groups.find(g => g.name === key);
      if (group.is_visible) {
        values.push(calculatePercentage(elem.data[key], data[0].data[key]));
      }
    }
    return [elem.name, ...values];
  });
  return result;
};

export const generateColors = (requiredCumberOfColors) => {
  const adder = Math.floor(visualizationColors.length / requiredCumberOfColors);
  const colors = [];
  for (let i = 0; i < requiredCumberOfColors; i++) {
    colors.push(visualizationColors[(i * adder) % 10]);
  }
  return colors;
};

export const generateGroups = (data) => {
  const cat_names = Object.keys(data[0].data);
  const result = cat_names.map(elem => {
    return {
      name: elem,
      conversion_rate: calculatePercentage(data[data.length - 1].data[elem], data[0].data[elem]) + '%',
      is_visible: true
    };
  });
  return result;
};

export const generateTableColumns = (data, currentSorter, handleSorting) => {
  const result = [
    {
      title: 'Grouping',
      dataIndex: 'name',
      className: tableStyles.groupColumn
    },
    {
      title: 'Conversion',
      dataIndex: 'conversion',
      className: tableStyles.conversionColumn
    }
  ];
  const eventColumns = data.map((elem, index) => {
    return {
      title: getTitleWithSorter(elem.name, elem.name, currentSorter, handleSorting),
      dataIndex: elem.name,
      className: index === data.length - 1 ? tableStyles.lastColumn : ''
    };
  });
  return [...result, ...eventColumns];
};

export const generateTableData = (data, groups, currentSorter, searchText) => {
  const appliedGroups = groups.map(elem => elem.name).filter(elem => elem.toLowerCase().indexOf(searchText.toLowerCase()) > -1);
  const result = appliedGroups.map((group, index) => {
    const eventsData = {};
    data.forEach(d => {
      eventsData[d.name] = d.data[group] + ' (' + calculatePercentage(d.data[group], data[0].data[group]) + '%)';
    });
    return {
      index,
      name: group,
      conversion: calculatePercentage(data[data.length - 1].data[group], data[0].data[group]) + '%',
      ...eventsData
    };
  });

  result.sort((a, b) => {
    if (currentSorter.order === 'ascend') {
      return parseInt(a[currentSorter.key].split(' ')[0]) >= parseInt(b[currentSorter.key].split(' ')[0]) ? 1 : -1;
    }
    if (currentSorter.order === 'descend') {
      return parseInt(a[currentSorter.key].split(' ')[0]) <= parseInt(b[currentSorter.key].split(' ')[0]) ? 1 : -1;
    }
    return 0;
  });

  return result;
};

const groupedDummyData = [
  {
    index: 1,
    display: true,
    data: {
      Chennai: 20000,
      Mumbai: 20000,
      'New Delhi': 20000,
      Amritsar: 20000,
      Jalandhar: 20000,
      Kolkatta: 20000
    }
  },
  {
    index: 2,
    display: true,
    data: {
      Chennai: 8000,
      Mumbai: 8000,
      'New Delhi': 12000,
      Amritsar: 10000,
      Jalandhar: 12000,
      Kolkatta: 6000
    }
  },
  {
    index: 3,
    display: true,
    data: {
      Chennai: 6000,
      Mumbai: 6000,
      'New Delhi': 6000,
      Amritsar: 8000,
      Jalandhar: 8000,
      Kolkatta: 5000
    }
  },
  {
    index: 4,
    display: true,
    data: {
      Chennai: 2000,
      Mumbai: 3000,
      'New Delhi': 3000,
      Amritsar: 6000,
      Jalandhar: 4000,
      Kolkatta: 4000
    }
  },
  {
    index: 5,
    display: true,
    data: {
      Chennai: 1000,
      Mumbai: 2000,
      'New Delhi': 1600,
      Amritsar: 4000,
      Jalandhar: 1050,
      Kolkatta: 3000
    }
  },
  {
    index: 6,
    display: true,
    data: {
      Chennai: 600,
      Mumbai: 1600,
      'New Delhi': 1200,
      Amritsar: 3600,
      Jalandhar: 300,
      Kolkatta: 2000
    }
  },
  {
    index: 7,
    display: true,
    data: {
      Chennai: 300,
      Mumbai: 800,
      'New Delhi': 600,
      Amritsar: 1800,
      Jalandhar: 100,
      Kolkatta: 1200
    }
  }
];

export const generateDummyData = (labels) => {
  const result = labels.map((elem, index) => {
    return { ...groupedDummyData[index], name: elem };
  });
  return result;
};

export const generateUngroupedChartsData = (events) => {

  const response = funnelsDataWithoutBreakdown;

  let i = 0, result = [];

  events.forEach((event, index) => {
    if (index === 0) {
      result.push({
        event,
        netCount: response.rows[0][i],
        value: 100
      })
      i++;
    } else {
      result.push({
        event,
        netCount: response.rows[0][i],
        value: response.rows[0][i + index]
      });
      i = i + 2;
    }
  });

  return result;
};

export const checkForWindowSizeChange = (callback) => {
  if (window.outerWidth !== windowSize.w || window.outerHeight !== windowSize.h) {
    setTimeout(() => {
      windowSize.w = window.outerWidth; // update object with current window properties
      windowSize.h = window.outerHeight;
      windowSize.iw = window.innerWidth;
      windowSize.ih = window.innerHeight;
    }, 0);
    callback();
  }

  // if the window doesn't resize but the content inside does by + or - 5%
  else if (window.innerWidth + window.innerWidth * 0.05 < windowSize.iw ||
    window.innerWidth - window.innerWidth * 0.05 > windowSize.iw) {
    setTimeout(() => {
      windowSize.iw = window.innerWidth;
    }, 0);
    callback();
  }
};

export const calculatePercentage = (numerator, denominator, precision = 1) => {
  const result = ((numerator / denominator) * 100);
  return result % 1 !== 0 ? result.toFixed(precision) : result;
};

export const getTitleWithSorter = (title, key, currentSorter, handleSorting) => {
  return (
    <div className="flex items-center justify-between">
      <div className="mr-2">{title}</div>
      <div className="flex flex-col items-center">
        {currentSorter.key === key && currentSorter.order === 'ascend' ? (
          <div onClick={() => handleSorting({})} style={{ marginBottom: '1px' }}>
            <svg width="10" height="6" viewBox="0 0 10 6" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path d="M5.35117 1.3516L8.39362 4.36853C8.61554 4.58859 8.45971 4.96706 8.14718 4.96706H5.00002H1.83533C1.52237 4.96706 1.36673 4.58769 1.58953 4.36789L4.64797 1.3507C4.84302 1.15827 5.15661 1.15867 5.35117 1.3516Z" fill="#0E2647" stroke="#0E2647" />
            </svg>
          </div>
        ) : (
            <div onClick={() => handleSorting({ key, order: 'ascend' })} style={{ marginBottom: '1px' }}>
              <svg width="9" height="6" viewBox="0 0 9 6" fill="none" xmlns="http://www.w3.org/2000/svg">
                <path d="M5.35117 1.3516L8.39362 4.36853C8.61554 4.58859 8.45971 4.96706 8.14718 4.96706H5.00002H1.83533C1.52237 4.96706 1.36673 4.58769 1.58953 4.36789L4.64797 1.3507C4.84302 1.15827 5.15661 1.15867 5.35117 1.3516Z" stroke="#0E2647" />
              </svg>
            </div>
          )}
        {currentSorter.key === key && currentSorter.order === 'descend' ? (
          <div onClick={() => handleSorting({})} style={{ marginTop: '1px' }}>
            <svg width="11" height="7" viewBox="0 0 11 7" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path d="M6.35165 5.65415L10.3949 1.63808C10.6165 1.41794 10.4606 1.03976 10.1482 1.03976H6.00002H1.83368C1.52095 1.03976 1.36521 1.41866 1.58754 1.63859L5.64766 5.65488C5.84276 5.84787 6.15695 5.84755 6.35165 5.65415Z" fill="#0E2647" stroke="#0E2647" />
            </svg>
          </div>
        ) : (
            <div onClick={() => handleSorting({ key, order: 'descend' })} style={{ marginTop: '1px' }}>
              <svg width="9" height="6" viewBox="0 0 9 6" fill="none" xmlns="http://www.w3.org/2000/svg">
                <path d="M4.64721 4.65009L1.60105 1.59899C1.38084 1.37842 1.53706 1.0017 1.84874 1.0017L4.99996 1.0017L8.1721 1.00172C8.4843 1.00172 8.64028 1.3795 8.41903 1.59976L5.3538 4.65118C5.1583 4.8458 4.84211 4.84532 4.64721 4.65009Z" stroke="#0E2647" />
              </svg>

            </div>
          )}
      </div>
    </div>
  );
};
