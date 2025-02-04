import React from 'react';
import moment from 'moment';
import get from 'lodash/get';
import findIndex from 'lodash/findIndex';
import { labelsObj } from '../../utils';
import {
  addQforQuarter,
  getClickableTitleSorter,
  SortResults
} from '../../../../utils/dataFormatter';
import { Number as NumFormat } from '../../../../components/factorsComponents';
import {
  parseForDateTimeLabel,
  getBreakdownDisplayName
} from '../eventsAnalytics.helpers';
import { getBreakDownGranularities } from '../SingleEventMultipleBreakdown/utils';
import {
  DATE_FORMATS,
  MAX_ALLOWED_VISIBLE_PROPERTIES,
  DISPLAY_PROP
} from '../../../../utils/constants';
import { EVENT_COUNT_KEY } from '../eventsAnalytics.constants';
import { BREAKDOWN_TYPES } from '../../constants';
import { isNumeric } from '../../../../utils/global';
import { Tooltip } from 'antd';
import truncateURL from 'Utils/truncateURL';

export const defaultSortProp = ({ breakdown }) => {
  const dateTimeBreakdownIndex = findIndex(
    breakdown,
    (b) => b.prop_type === BREAKDOWN_TYPES.DATETIME
  );
  if (dateTimeBreakdownIndex > -1) {
    return [
      {
        key: `${breakdown[dateTimeBreakdownIndex].property} - ${dateTimeBreakdownIndex}`,
        type: BREAKDOWN_TYPES.DATETIME,
        subtype: get(breakdown[dateTimeBreakdownIndex], 'grn', null),
        order: 'descend'
      }
    ];
  }
  return [
    {
      order: 'descend',
      key: EVENT_COUNT_KEY,
      type: 'numerical',
      subtype: null
    }
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

export const getBreakdownTitle = (
  breakdown,
  userPropNames,
  eventPropertiesDisplayNames
) => {
  const charArr = ['A', 'B', 'C', 'D', 'E', 'F', 'G', 'H'];
  const displayTitle = getBreakdownDisplayName({
    breakdown,
    userPropNames,
    eventPropertiesDisplayNames,
    multipleEvents: true
  });

  if (!breakdown.eventIndex) {
    return <div className='break-all'>{displayTitle}</div>;
  }
  return (
    <div className='break-all'>
      <span>{displayTitle} of </span>
      <span className='inline-block'>
        <span
          style={{ backgroundColor: '#3E516C' }}
          className='text-white w-4 h-4 flex justify-center items-center rounded-full font-semibold leading-5 text-xs'
        >
          {charArr[breakdown.eventIndex - 1]}
        </span>
      </span>
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
  // eslint-disable-next-line camelcase
  const eventIdxIndex = headers.findIndex((elem) => elem === 'event_index');
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
    const str = `${eventName},${Object.values(breakdownData).join(', ')}`;
    const queryIndex = queries.findIndex((_, i) => i === d[eventIdxIndex]);
    return {
      label: str,
      value: d[countIndex],
      [EVENT_COUNT_KEY]: d[countIndex], // used for sorting, value key will be removed soon
      index,
      event: eventName,
      eventIndex: d[eventIdxIndex],
      color: colors[queryIndex],
      ...breakdownData
    };
  });

  return result;
};

export const formatVisibleProperties = (data, queries) => {
  const vp = data.map((d) => ({
    ...d,
    label: `${d.label}; [${d.eventIndex}]`
  }));
  vp.sort((a, b) => (parseInt(a.value) <= parseInt(b.value) ? 1 : -1));
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
  eventNames,
  userPropNames,
  eventPropertiesDisplayNames,
  projectDomainsList,
  eventGroup
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
    render: (d) => eventNames[d] || d
  });
  breakdown.forEach((b, index) => {
    result.push({
      title: getClickableTitleSorter(
        getBreakdownTitle(b, userPropNames, eventPropertiesDisplayNames),
        { key: `${b.property} - ${index}`, type: b.prop_type, subtype: b.grn },
        currentSorter,
        handleSorting
      ),
      dataIndex: `${b.property} - ${index}`,
      width: 200,
      render: (d) => {
        if (
          b.prop_type === 'numerical' &&
          (typeof d === 'number' || isNumeric(d))
        ) {
          return <NumFormat number={d} />;
        }
        return (
          <Tooltip placement='top' title={d}>
            {truncateURL(d, projectDomainsList)}
          </Tooltip>
        );
      }
    });
  });
  result.push({
    title: getClickableTitleSorter(
      labelsObj[eventGroup] || 'Count',
      { key: EVENT_COUNT_KEY, type: 'numerical', subtype: null },
      currentSorter,
      handleSorting,
      'right'
    ),
    className: 'text-right',
    dataIndex: EVENT_COUNT_KEY,
    width: 150,
    render: (d) => <NumFormat number={d} />
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
  eventPropertiesDisplayNames,
  projectDomainsList
) => {
  console.log('mewb getDateBasedColumns');
  const OverallColumn = {
    title: getClickableTitleSorter(
      'Overall',
      { key: EVENT_COUNT_KEY, type: 'numerical', subtype: null },
      currentSorter,
      handleSorting,
      'right'
    ),
    className: 'text-right',
    dataIndex: EVENT_COUNT_KEY,
    width: 150
  };
  const breakdownColumns = breakdown.map((b, index) => ({
    title: getClickableTitleSorter(
      getBreakdownTitle(b, userPropNames, eventPropertiesDisplayNames),
      { key: `${b.property} - ${index}`, type: b.prop_type, subtype: b.grn },
      currentSorter,
      handleSorting
    ),
    dataIndex: `${b.property} - ${index}`,
    width: 200,
    render: (d) => {
      if (
        b.prop_type === 'numerical' &&
        (typeof d === 'number' || isNumeric(d))
      ) {
        return <NumFormat number={d} />;
      }
      return (
        <Tooltip placement='top' title={d}>
          {truncateURL(d, projectDomainsList)}
        </Tooltip>
      );
    }
  }));
  const format = DATE_FORMATS[frequency] || DATE_FORMATS.date;

  const dateColumns = categories.map((cat) => ({
    title: getClickableTitleSorter(
      addQforQuarter(frequency) + moment(cat).format(format),
      {
        key: addQforQuarter(frequency) + moment(cat).format(format),
        type: 'numerical',
        subtype: null
      },
      currentSorter,
      handleSorting,
      'right'
    ),
    width: 150,
    dataIndex: addQforQuarter(frequency) + moment(cat).format(format),
    render: (d) => <NumFormat number={d} />,
    className: 'text-right'
  }));
  const eventCol = {
    title: getClickableTitleSorter(
      'Event',
      { key: 'event', type: 'categorical', subtype: null },
      currentSorter,
      handleSorting
    ),
    dataIndex: 'event',
    fixed: 'left',
    width: 200
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
      data: []
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
  const dateWiseTotals = Array(differentDates.length).fill(0);
  const initializedDatesData = differentDates.map(() => 0);
  const labelsMapper = {};
  const resultantData = aggregateData.map((d, index) => {
    labelsMapper[d.label] = index;
    return {
      name: d.label,
      data: [...initializedDatesData],
      marker: {
        enabled: false
      },
      ...d
    };
  });

  const headerSlice = data.headers.slice(eventIndex + 1, countIndex);
  const breakdowns = data.meta.query.gbp ? [...data.meta.query.gbp] : [];
  const grns = getBreakDownGranularities(headerSlice, breakdowns);
  const format = DATE_FORMATS[frequency] || DATE_FORMATS.date;

  data.rows.forEach((row) => {
    const eventName = eventNames[row[eventIndex]] || row[eventIndex];
    const breakdownJoin = `${eventName},${row
      .slice(eventIndex + 1, countIndex)
      .map((x, ind) =>
        parseForDateTimeLabel(grns[ind], DISPLAY_PROP[x] ? DISPLAY_PROP[x] : x)
      )
      .join(', ')}`;
    const bIdx = labelsMapper[breakdownJoin];
    const category = row[dateIndex];
    const idx = differentDates.indexOf(category);
    if (resultantData[bIdx]) {
      resultantData[bIdx][
        addQforQuarter(frequency) + moment(category).format(format)
      ] = row[countIndex];
      resultantData[bIdx].data[idx] = row[countIndex];
      dateWiseTotals[idx] += row[countIndex];
    }
  });
  return {
    categories: differentDates,
    data: resultantData,
    dateWiseTotals
  };
};
