import React from 'react';
import moment from 'moment';
import { labelsObj } from '../../utils';
import {
  getClickableTitleSorter,
  SortResults,
} from '../../../../utils/dataFormatter';
import { Number as NumFormat } from '../../../../components/factorsComponents';
import { parseForDateTimeLabel } from '../SingleEventSingleBreakdown/utils';
import { getBreakDownGranularities } from '../SingleEventMultipleBreakdown/utils';
import {
  DATE_FORMATS,
  MAX_ALLOWED_VISIBLE_PROPERTIES,
} from '../../../../utils/constants';
import { DISPLAY_PROP } from '../../../../utils/constants';

export const defaultSortProp = () => {
  return [
    {
      order: 'descend',
      key: 'Event Count',
      type: 'numerical',
      subtype: null,
    },
  ];
};

export const getVisibleData = (aggregateData, sorter) => {
  const result = SortResults(aggregateData, sorter).slice(
    0,
    MAX_ALLOWED_VISIBLE_PROPERTIES
  );
  return result;
};

export const getVisibleSeriesData = (data, sorter) => {
  const result = SortResults(data, sorter).slice(
    0,
    MAX_ALLOWED_VISIBLE_PROPERTIES
  );
  return result;
};

export const getBreakdownTitle = (breakdown, userPropNames, eventPropNames) => {
  const charArr = ['A', 'B', 'C', 'D', 'E', 'F', 'G', 'H'];
  const displayTitle =
    breakdown.prop_category === 'user'
      ? userPropNames[breakdown.property] || breakdown.property
      : breakdown.prop_category === 'event'
      ? eventPropNames[breakdown.property] || breakdown.property
      : breakdown.property;

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
  console.log('mewb formatData');
  const { headers, rows } = data.metrics;
  const event_indexIndex = headers.findIndex((elem) => elem === 'event_index');
  const countIndex = headers.findIndex(
    (elem) => elem === 'count' || elem === 'aggregate'
  );
  const eventIndex = headers.findIndex((elem) => elem === 'event_name');

  const headerSlice = headers.slice(eventIndex + 1, countIndex);
  const breakdowns = data.meta.query.gbp ? [...data.meta.query.gbp] : [];
  const grns = getBreakDownGranularities(headerSlice, breakdowns);

  const result = rows.map((d, index) => {
    const breakdownVals = d
      .slice(eventIndex + 1, countIndex)
      .map((vl) => (DISPLAY_PROP[vl] ? DISPLAY_PROP[vl] : vl));
    const breakdownData = {};
    for (let i = 0; i < breakdowns.length; i++) {
      const bkd = breakdowns[i];
      breakdownData[`${bkd.pr} - ${i}`] = parseForDateTimeLabel(
        grns[i],
        breakdownVals[i]
      );
    }
    const eventName = eventNames[d[eventIndex]] || d[eventIndex];
    const str = eventName + ',' + Object.values(breakdownData).join(', ');
    const queryIndex = queries.findIndex(
      (_, index) => index === d[event_indexIndex]
    );
    return {
      label: str,
      value: d[countIndex],
      'Event Count': d[countIndex], //used for sorting, value key will be removed soon
      index,
      event: eventName,
      eventIndex: d[event_indexIndex],
      color: colors[queryIndex],
      ...breakdownData,
    };
  });

  return result;
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
  console.log('mewb getTableColumns');
  const result = [];
  result.push({
    title: getClickableTitleSorter(
      'Event',
      { key: 'event', type: 'categorical', subtype: null },
      currentSorter,
      handleSorting
    ),
    dataIndex: 'event',
    width: 200,
    fixed: 'left',
    render: (d) => {
      return eventNames[d] || d;
    },
  });
  breakdown.forEach((b, index) => {
    result.push({
      title: getClickableTitleSorter(
        getBreakdownTitle(b, userPropNames, eventPropNames),
        { key: `${b.property} - ${index}`, type: b.prop_type, subtype: b.grn },
        currentSorter,
        handleSorting
      ),
      dataIndex: `${b.property} - ${index}`,
      width: 200,
    });
  });
  result.push({
    title: getClickableTitleSorter(
      labelsObj[page],
      { key: `Event Count`, type: 'numerical', subtype: null },
      currentSorter,
      handleSorting
    ),
    dataIndex: 'Event Count',
    width: 150,
    render: (d) => {
      return <NumFormat number={d} />;
    },
  });
  return result;
};

export const getTableData = (data, searchText, currentSorter) => {
  console.log('mewb getTableData');
  const filteredData = data.filter(
    (elem) =>
      elem.label.toLowerCase().includes(searchText.toLowerCase()) ||
      elem.event.toLowerCase().includes(searchText.toLowerCase())
  );
  return SortResults(filteredData, currentSorter);
};

export const getDateBasedColumns = (
  categories,
  breakdown,
  currentSorter,
  handleSorting,
  frequency,
  userPropNames,
  eventPropNames
) => {
  console.log('mewb getDateBasedColumns');
  const OverallColumn = {
    title: getClickableTitleSorter(
      'Overall',
      { key: `Event Count`, type: 'numerical', subtype: null },
      currentSorter,
      handleSorting
    ),
    dataIndex: `Event Count`,
    width: 150,
  };
  const breakdownColumns = breakdown.map((b, index) => {
    return {
      title: getClickableTitleSorter(
        getBreakdownTitle(b, userPropNames, eventPropNames),
        { key: `${b.property} - ${index}`, type: b.prop_type, subtype: b.grn },
        currentSorter,
        handleSorting
      ),
      dataIndex: `${b.property} - ${index}`,
      width: 200,
    };
  });
  const format = DATE_FORMATS[frequency] || DATE_FORMATS['date'];

  const dateColumns = categories.map((cat) => {
    return {
      title: getClickableTitleSorter(
        moment(cat).format(format),
        { key: moment(cat).format(format), type: 'numerical', subtype: null },
        currentSorter,
        handleSorting
      ),
      width: 150,
      dataIndex: moment(cat).format(format),
      render: (d) => {
        return <NumFormat number={d} />;
      },
    };
  });
  const eventCol = {
    title: getClickableTitleSorter(
      'Event',
      { key: 'event', type: 'categorical', subtype: null },
      currentSorter,
      handleSorting
    ),
    dataIndex: 'event',
    fixed: 'left',
    width: 200,
  };
  return [eventCol, ...breakdownColumns, ...dateColumns, OverallColumn];
};

export const getDateBasedTableData = (
  seriesData,
  currentSorter,
  searchText
) => {
  console.log('mewb getDateBasedTableData');
  const result = seriesData.filter((sd) =>
    sd.name.toLowerCase().includes(searchText.toLowerCase())
  );
  return SortResults(result, currentSorter);
};

export const formatDataInStackedAreaFormat = (
  data,
  aggregateData,
  eventNames,
  frequency
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
  console.log('mewb formatDataInStackedAreaFormat');
  const dateIndex = data.headers.findIndex((h) => h === 'datetime');
  const countIndex = data.headers.findIndex(
    (h) => h === 'count' || h === 'aggregate'
  );
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
      marker: {
        enabled: false,
      },
      ...d,
    };
  });

  const headerSlice = data.headers.slice(eventIndex + 1, countIndex);
  const breakdowns = data.meta.query.gbp ? [...data.meta.query.gbp] : [];
  const grns = getBreakDownGranularities(headerSlice, breakdowns);
  const format = DATE_FORMATS[frequency] || DATE_FORMATS['date'];

  data.rows.forEach((row) => {
    const eventName = eventNames[row[eventIndex]] || row[eventIndex];
    const breakdownJoin =
      eventName +
      ',' +
      row
        .slice(eventIndex + 1, countIndex)
        .map((x, ind) =>
          parseForDateTimeLabel(
            grns[ind],
            DISPLAY_PROP[x] ? DISPLAY_PROP[x] : x
          )
        )
        .join(', ');
    const bIdx = labelsMapper[breakdownJoin];
    const category = row[dateIndex];
    const idx = differentDates.indexOf(category);
    if (resultantData[bIdx]) {
      resultantData[bIdx][moment(category).format(format)] = row[countIndex];
      resultantData[bIdx].data[idx] = row[countIndex];
    }
  });
  return {
    categories: differentDates,
    data: resultantData,
  };
};
