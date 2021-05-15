import React from 'react';
import moment from 'moment';
import { labelsObj } from '../../utils';
import { SortData, getTitleWithSorter } from '../../../../utils/dataFormatter';
import { Number as NumFormat } from '../../../../components/factorsComponents';

export const getBreakdownTitle = (breakdown, userPropNames, eventPropNames) => {
  const charArr = ['A', 'B', 'C', 'D', 'E', 'F', 'G', 'H'];
  let displayTitle = breakdown.property;
  if(breakdown.prop_category === 'user') {
    displayTitle = userPropNames[breakdown.property]? userPropNames[breakdown.property] : breakdown.property;
  }

  if(breakdown.prop_category === 'event') {
    displayTitle = eventPropNames[breakdown.property]? eventPropNames[breakdown.property] : breakdown.property;
  }
  
  if (!breakdown.eventIndex) {
    return displayTitle;
  }
  return (
    <div className='flex items-center'>
      <div className='mr-1'>{displayTitle} of </div>
      <div
        style={{ backgroundColor: '#3E516C' }}
        className='text-white w-4 h-4 flex justify-center items-center rounded-full font-semibold leading-5 text-xs'
      >
        {charArr[breakdown.eventIndex - 1]}
      </div>
    </div>
  );
};

export const formatData = (data, queries, colors, eventNames) => {
  if (
    !data ||
    !data.metrics ||
    !data.metrics.headers ||
    !data.metrics.headers.length ||
    !data.metrics.rows ||
    !data.metrics.rows.length
  ) {
    return [];
  }
  const { headers, rows } = data.metrics;
  const event_indexIndex = headers.findIndex((elem) => elem === 'event_index');
  const countIndex = headers.findIndex((elem) => elem === 'count');
  const eventIndex = headers.findIndex((elem) => elem === 'event_name');

  const result = rows.map((d, index) => {
    const eventName = eventNames[d[eventIndex]] || d[eventIndex];
    const str = eventName + ',' + d.slice(eventIndex + 1, countIndex).join(',');
    const queryIndex = queries.findIndex(
      (_, index) => index === d[event_indexIndex]
    );
    return {
      label: str,
      value: d[countIndex],
      index,
      event: d[eventIndex],
      eventIndex: d[event_indexIndex],
      color: colors[queryIndex],
    };
  });

  const sortedData = SortData(result, 'value', 'descend');
  const maxIndices = [];
  queries.forEach((_, qIdx) => {
    const idx = sortedData.findIndex((elem) => elem.eventIndex === qIdx);
    if (idx > -1) {
      maxIndices.push(idx);
    }
  });
  const finalResult = maxIndices.map((m) => {
    return sortedData[m];
  });
  sortedData.forEach((sd, idx) => {
    if (maxIndices.indexOf(idx) === -1) {
      finalResult.push(sd);
    }
  });
  return finalResult;
};

export const formatVisibleProperties = (data, queries) => {
  const vp = data.map((d) => {
    return { ...d, label: `${d.label}; [${d.eventIndex}]` };
  });
  vp.sort((a, b) => {
    return parseInt(a.value) <= parseInt(b.value) ? 1 : -1;
  });
  vp.sort((a, b) => {
    const idx1 = queries.findIndex((_, index) => index === a.eventIndex);
    const idx2 = queries.findIndex((_, index) => index === b.eventIndex);
    return idx1 >= idx2 ? 1 : -1;
  });
  return vp;
};

export const getTableColumns = (
  breakdown,
  currentSorter,
  handleSorting,
  page,
  eventNames,
  userPropNames,
  eventPropNames
) => {
  const result = [];
  result.push({
    title: 'Event',
    dataIndex: 'event',
    render: (d) => {
      return eventNames[d] || d;
    },
  });
  breakdown.forEach((b, index) => {
    result.push({
      title: getBreakdownTitle(b, userPropNames, eventPropNames),
      dataIndex: b.property + ';' + index,
    });
  });
  result.push({
    title: getTitleWithSorter(
      labelsObj[page],
      'Event Count',
      currentSorter,
      handleSorting
    ),
    dataIndex: 'Event Count',
    render: (d) => {
      return <NumFormat number={d} />;
    },
  });
  return result;
};

export const getTableData = (data, breakdown, searchText, currentSorter) => {
  const filteredData = data.filter(
    (elem) =>
      elem.label.toLowerCase().includes(searchText.toLowerCase()) ||
      elem.event.toLowerCase().includes(searchText.toLowerCase())
  );
  const result = [];
  filteredData.forEach((d) => {
    const breakdownValues = {};
    breakdown.forEach((b, index) => {
      breakdownValues[b.property + ';' + index] = d.label.split(',')[index + 1];
    });
    result.push({
      ...d,
      'Event Count': d.value,
      ...breakdownValues,
    });
  });
  return SortData(result, currentSorter.key, currentSorter.order);
};

export const getDateBasedColumns = (
  categories,
  breakdown,
  currentSorter,
  handleSorting,
  frequency,
  userPropNames, eventPropNames
) => {
  const breakdownColumns = breakdown.map((elem, index) => {
    return {
      title: getBreakdownTitle(elem, userPropNames, eventPropNames),
      dataIndex: elem.property + ';' + index,
      width: 200,
    };
  });
  let format = 'MMM D';
  if (frequency === 'hour') {
    format = 'h A, MMM D';
  }
  const dateColumns = categories.map((cat) => {
    return {
      title: getTitleWithSorter(
        moment(cat).format(format),
        moment(cat).format(format),
        currentSorter,
        handleSorting
      ),
      width: 100,
      dataIndex: moment(cat).format(format),
      render: (d) => {
        return <NumFormat number={d} />;
      },
    };
  });
  const eventCol = {
    title: 'Event',
    dataIndex: 'Event',
    fixed: 'left',
    width: 200,
  };
  return [eventCol, ...breakdownColumns, ...dateColumns];
};

export const getDateBasedTableData = (
  seriesData,
  categories,
  breakdown,
  currentSorter,
  searchText,
  frequency
) => {
  let format = 'MMM D';
  if (frequency === 'hour') {
    format = 'h A, MMM D';
  }
  const result = seriesData
    .filter((sd) => sd.name.toLowerCase().includes(searchText.toLowerCase()))
    .map((sd) => {
      const dateWiseData = {};
      categories.forEach((cat, index) => {
        dateWiseData[moment(cat).format(format)] = sd.data[index];
      });
      const splittedLabel = sd.name.split(',');
      const breakdownData = {};
      breakdown.forEach((b, index) => {
        breakdownData[b.property + ';' + index] = splittedLabel[index + 1];
      });
      return {
        index: sd.index,
        Event: sd.name.split(',')[0],
        ...breakdownData,
        ...dateWiseData,
      };
    });
  return SortData(result, currentSorter.key, currentSorter.order);
};

export const formatDataInStackedAreaFormat = (
  data,
  aggregateData,
  eventNames
) => {
  if (
    !data.headers ||
    !data.headers.length ||
    !data.rows ||
    !data.rows.length ||
    !aggregateData.length
  ) {
    return {
      categories: [],
      data: [],
    };
  }
  const dateIndex = data.headers.findIndex((h) => h === 'datetime');
  const countIndex = data.headers.findIndex((h) => h === 'count');
  const eventIndex = data.headers.findIndex((h) => h === 'event_name');
  let differentDates = new Set();
  data.rows.forEach((row) => {
    differentDates.add(row[dateIndex]);
  });
  differentDates = Array.from(differentDates);
  const initializedDatesData = differentDates.map(() => {
    return 0;
  });
  const labelsMapper = {};
  const resultantData = aggregateData.map((d, index) => {
    labelsMapper[d.label] = index;
    return {
      name: d.label,
      data: [...initializedDatesData],
      index: d.index,
      marker: {
        enabled: false,
      },
    };
  });
  data.rows.forEach((row) => {
    const eventName = eventNames[row[eventIndex]] || row[eventIndex];
    const breakdownJoin =
      eventName + ',' + row.slice(eventIndex + 1, countIndex).join(',');
    const bIdx = labelsMapper[breakdownJoin];
    const idx = differentDates.indexOf(row[dateIndex]);
    if (resultantData[bIdx]) {
      resultantData[bIdx].data[idx] = row[countIndex];
    }
  });
  return {
    categories: differentDates,
    data: resultantData,
  };
};
