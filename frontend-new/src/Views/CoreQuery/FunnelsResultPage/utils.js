/* eslint-disable */
import React from 'react';
import tableStyles from './FunnelsResultTable/index.module.scss';
import { SortData } from '../utils';

const windowSize = {
  w: window.outerWidth,
  h: window.outerHeight,
  iw: window.innerWidth,
  ih: window.innerHeight
};

const visualizationColors = ['#4D7DB4', '#4C9FC8', '#4CBCBD', '#86D3A3', '#CCC36D', '#F9C06E', '#E89E7B', '#D4787D', '#B87B7E', '#9982B5'];

export const generateGroupedChartsData = (response, queries, groups, eventsMapper) => {
  if (!response) {
    return [];
  }
  const result = queries.map(elem => {
    return [eventsMapper[elem]];
  });
  const firstEventIdx = response.headers.findIndex(elem => elem === 'step_0');
  response.rows.forEach((elem) => {
    const breakdownName = elem.slice(0, firstEventIdx).join(",");
    const isVisible = groups.filter(g => g.name === breakdownName && g.is_visible).length
    if (isVisible) {
      const netCounts = elem.filter(elem => typeof elem === 'number');
      netCounts.forEach((n, idx) => {
        result[idx].push(calculatePercentage(n, netCounts[0]));
      })
    }
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

export const generateGroups = (response, maxAllowedVisibleProperties) => {
  if (!response) {
    return [];
  }
  const firstEventIdx = response.headers.findIndex(elem => elem === 'step_0');
  const result = response.rows.map((elem, index) => {
    const netCounts = elem.filter(elem => typeof elem === 'number');
    return {
      index,
      name: elem.slice(0, firstEventIdx).join(","),
      conversion_rate: calculatePercentage(netCounts[netCounts.length - 1], netCounts[0]) + '%',
      is_visible: index < maxAllowedVisibleProperties ? true : false,
    };
  });
  return result;
};

export const generateTableColumns = (breakdown, queries, eventsMapper, currentSorter, handleSorting) => {
  const result = [
    {
      title: breakdown.length ? 'Grouping' : 'Users',
      dataIndex: 'name',
      className: tableStyles.groupColumn
    },
    {
      title: 'Conversion',
      dataIndex: 'conversion',
      className: tableStyles.conversionColumn
    }
  ];
  const eventColumns = queries.map((elem, index) => {
    return {
      title: getTitleWithSorter(elem, elem, currentSorter, handleSorting),
      dataIndex: breakdown.length ? eventsMapper[elem] : elem,
      className: index === queries.length - 1 ? tableStyles.lastColumn : ''
    };
  });

  const blankCol = {
    title: '',
    dataIndex: '',
    width: 37
  };
  if (breakdown.length) {
    return [...result, ...eventColumns];
  } else {
    return [blankCol, ...result, ...eventColumns];
  }

};

export const generateTableData = (data, breakdown, queries, groups, eventsMapper, currentSorter, searchText) => {
  if (!breakdown.length) {
    const queryData = {};
    queries.forEach((q, index) => {
      queryData[q] = `${data[index].netCount} (${data[index].value}%)`;
    })
    return [
      {
        index: 0,
        ...queryData,
        name: 'All',
        conversion: data[data.length - 1].value + '%'
      }
    ]
  } else {
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

    return SortData(result, eventsMapper[currentSorter.key], currentSorter.order);
  }
};

export const generateUngroupedChartsData = (response, events) => {
  if (!response) {
    return [];
  }

  const netCounts = response.rows[0].filter(elem => typeof elem === 'number');
  const result = [];
  let index = 0;

  while (index < events.length) {
    if (index === 0) {
      result.push({
        event: events[index],
        netCount: netCounts[index],
        value: 100
      })
    } else {
      result.push({
        event: events[index],
        netCount: netCounts[index],
        value: calculatePercentage(netCounts[index], netCounts[0])
      })
    }
    index++;
  }
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


export const generateEventsData = (response, queries, eventsMapper) => {
  if (!response) {
    return [];
  }
  const firstEventIdx = response.headers.findIndex(elem => elem === 'step_0');
  const result = queries.map((q, idx) => {
    const data = {};
    response.rows.forEach(r => {
      const name = r.slice(0, firstEventIdx).join(",");
      const netCounts = r.filter(elem => typeof elem === 'number');
      data[name] = netCounts[idx];
    });
    return {
      index: idx + 1,
      data,
      name: eventsMapper[q]
    }
  });
  return result;
}