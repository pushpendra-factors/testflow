import React from 'react';
import get from 'lodash/get';
import has from 'lodash/has';
import findIndex from 'lodash/findIndex';
import moment from 'moment';
import {
  getClickableTitleSorter,
  SortResults,
  generateColors,
  addQforQuarter,
  SortData
} from '../../../../utils/dataFormatter';
import { Number as NumFormat } from '../../../../components/factorsComponents';
import {
  parseForDateTimeLabel,
  getBreakdownDisplayName,
  getEventDisplayName
} from '../eventsAnalytics.helpers';
import { labelsObj } from '../../utils';
import {
  DATE_FORMATS,
  MAX_ALLOWED_VISIBLE_PROPERTIES,
  DISPLAY_PROP
} from '../../../../utils/constants';
import HorizontalBarChartCell from './HorizontalBarChartCell';
import tableStyles from '../../../../components/DataTable/index.module.scss';
import NonClickableTableHeader from '../../../../components/NonClickableTableHeader';
import { EVENT_COUNT_KEY } from '../eventsAnalytics.constants';
import { BREAKDOWN_TYPES } from '../../constants';

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

export const getBreakDownGranularities = (breakDownSlice, breakdowns) => {
  const grns = [];
  const brks = [...breakdowns];
  breakDownSlice.forEach((h) => {
    const brkIndex = brks.findIndex((x) => h === (x.pr ? x.pr : x.property));
    grns.push(brks[brkIndex]?.grn);
    brks.splice(brkIndex, 1);
  });
  return grns;
};

export const formatData = (data) => {
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
  const eventNameIndex = headers.findIndex((header) => header === 'event_name');
  const countIndex = headers.findIndex(
    (header) => header === 'count' || header === 'aggregate'
  );

  const headerSlice = headers.slice(eventNameIndex + 1, countIndex);
  const breakdowns = data.meta.query.gbp ? [...data.meta.query.gbp] : [];
  const grns = getBreakDownGranularities(headerSlice, breakdowns);

  const result = rows.map((d, index) => {
    const breakdownVals = d
      .slice(eventNameIndex + 1, countIndex)
      .map((vl) => (DISPLAY_PROP[vl] ? DISPLAY_PROP[vl] : vl));
    const breakdownData = {};
    for (let i = 0; i < breakdowns.length; i++) {
      const bkd = breakdowns[i];
      breakdownData[`${bkd.pr} - ${i}`] = parseForDateTimeLabel(
        grns[i],
        breakdownVals[i]
      );
    }
    const grpLabel = Object.values(breakdownData).join(', ');
    return {
      label: grpLabel,
      value: d[countIndex],
      [EVENT_COUNT_KEY]: d[countIndex], // used for sorting, value key will be removed soon
      index,
      ...breakdownData
    };
  });
  return result;
};

export const getTableColumns = (
  events,
  breakdown,
  currentSorter,
  handleSorting,
  page,
  eventNames,
  userPropNames,
  eventPropNames
) => {
  const breakdownColumns = breakdown.map((e, index) => {
    const displayTitle = getBreakdownDisplayName({
      breakdown: e,
      userPropNames,
      eventPropNames
    });

    return {
      title: getClickableTitleSorter(
        <div className='break-all'>{displayTitle}</div>,
        { key: `${e.property} - ${index}`, type: e.prop_type, subtype: e.grn },
        currentSorter,
        handleSorting
      ),
      dataIndex: `${e.property} - ${index}`,
      fixed: !index ? 'left' : '',
      width: 200,
      render: (d) => {
        if (e.prop_type === 'numerical' && !Number.isNaN(d)) {
          return <NumFormat number={d} />;
        }
        return d;
      }
    };
  });

  const title = getEventDisplayName({ eventNames, event: events[0] });

  const countColumn = {
    title: getClickableTitleSorter(
      <div className='break-all'>
        {title}: {labelsObj[page]}
      </div>,
      { key: EVENT_COUNT_KEY, type: 'numerical', subtype: null },
      currentSorter,
      handleSorting,
      'right'
    ),
    className: 'text-right',
    dataIndex: EVENT_COUNT_KEY,
    width: 200,
    render: (d) => <NumFormat number={d} />
  };

  return [...breakdownColumns, countColumn];
};

export const getDataInTableFormat = (data, searchText, currentSorter) => {
  const filteredData = data.filter((elem) =>
    elem.label.toLowerCase().includes(searchText.toLowerCase())
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
  const breakdownColumns = breakdown.map((e, index) => {
    const displayTitle = getBreakdownDisplayName({
      breakdown: e,
      userPropNames,
      eventPropNames
    });

    return {
      title: getClickableTitleSorter(
        <div className='break-all'>{displayTitle}</div>,
        { key: `${e.property} - ${index}`, type: e.prop_type, subtype: e.grn },
        currentSorter,
        handleSorting
      ),
      dataIndex: `${e.property} - ${index}`,
      fixed: !index ? 'left' : '',
      width: 200,
      render: (d) => {
        if (e.prop_type === 'numerical' && !Number.isNaN(d)) {
          return <NumFormat number={d} />;
        }
        return d;
      }
    };
  });

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
    className: 'text-right',
    width: 150,
    dataIndex: addQforQuarter(frequency) + moment(cat).format(format),
    render: (d) => <NumFormat number={d} />
  }));
  return [...breakdownColumns, ...dateColumns, OverallColumn];
};

export const getDateBasedTableData = (
  seriesData,
  searchText,
  currentSorter
) => {
  const result = seriesData.filter((sd) =>
    sd.name.toLowerCase().includes(searchText.toLowerCase())
  );

  return SortResults(result, currentSorter);
};

export const formatDataInStackedAreaFormat = (
  data,
  aggregateData,
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
  const dateIndex = data.headers.findIndex((h) => h === 'datetime');
  const countIndex = data.headers.findIndex(
    (h) => h === 'count' || h === 'aggregate'
  );
  const eventIndex = data.headers.findIndex((h) => h === 'event_name');
  const breakdownIndex = eventIndex + 1;
  let differentDates = new Set();
  data.rows.forEach((row) => {
    differentDates.add(row[dateIndex]);
  });
  differentDates = Array.from(differentDates);
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

  const headerSlice = data.headers.slice(breakdownIndex, countIndex);
  const breakdowns = data.meta.query.gbp ? [...data.meta.query.gbp] : [];
  const grns = getBreakDownGranularities(headerSlice, breakdowns);
  const format = DATE_FORMATS[frequency] || DATE_FORMATS.date;

  data.rows.forEach((row) => {
    const breakdownJoin = row
      .slice(breakdownIndex, countIndex)
      .map((x, ind) =>
        parseForDateTimeLabel(grns[ind], DISPLAY_PROP[x] ? DISPLAY_PROP[x] : x)
      )
      .join(', ');
    const bIdx = labelsMapper[breakdownJoin];
    const category = row[dateIndex];
    const idx = differentDates.indexOf(category);
    if (resultantData[bIdx]) {
      resultantData[bIdx][
        addQforQuarter(frequency) + moment(category).format(format)
      ] = row[countIndex];
      resultantData[bIdx].data[idx] = row[countIndex];
    }
  });
  return {
    categories: differentDates,
    data: resultantData
  };
};

export const renderHorizontalBarChart = (
  data,
  key,
  cardSize = 1,
  isDashboardWidget = false,
  multipleBreakdowns = true
) => {
  const series = [
    {
      data: []
    }
  ];
  const colors = generateColors(10);
  const sortedData = SortData(data, 'value', 'descend');
  const categories = sortedData.map((elem, index) => {
    series[0].data.push({
      y: elem.value,
      color: colors[index % 10],
      metricType: get(elem, 'metricType', null)
    });
    return elem[key];
  });

  if (isDashboardWidget) {
    series[0].data = series[0].data.slice(0, 3);
  }

  return (
    <HorizontalBarChartCell
      series={series}
      categories={categories}
      cardSize={cardSize}
      isDashboardWidget={isDashboardWidget}
      width={isDashboardWidget || !multipleBreakdowns ? null : 600}
    />
  );
};

export const getBreakdownDataMapperWithUniqueValues = (data, key) => {
  let values = new Set();
  const breakdownMapper = {};
  data.forEach((d) => {
    const bValue = d[key];
    if (breakdownMapper[bValue]) {
      breakdownMapper[bValue].push(d);
    } else {
      breakdownMapper[bValue] = [d];
    }
    values.add(d[key]);
  });
  values = [...values];
  return {
    values,
    breakdownMapper
  };
};

export const getDataInHorizontalBarChartFormat = (
  aggregateData,
  breakdown,
  cardSize = 1,
  isDashboardWidget = false
) => {
  const sortedData = SortResults(aggregateData, [
    {
      key: 'value',
      order: 'descend'
    }
  ]);

  const firstBreakdownKey = `${breakdown[0].pr} - 0`;
  const secondBreakdownKey = `${breakdown[1].pr} - 1`;

  const {
    values: uniqueFirstBreakdownValues,
    breakdownMapper: firstBreakdownMapper
  } = getBreakdownDataMapperWithUniqueValues(sortedData, firstBreakdownKey);

  if (breakdown.length === 2) {
    const result = uniqueFirstBreakdownValues.map((bValue) => {
      const row = {};
      row.index = bValue;
      row[firstBreakdownKey] = { value: bValue };
      row[secondBreakdownKey] = {
        value: renderHorizontalBarChart(
          firstBreakdownMapper[bValue],
          secondBreakdownKey,
          cardSize,
          isDashboardWidget
        )
      };
      return row;
    });
    if (isDashboardWidget && result.length) {
      return [result[0]];
    }
    return result;
  }
  if (breakdown.length === 3) {
    const thirdBreakdownKey = `${breakdown[2].pr} - 2`;
    const result = [];
    uniqueFirstBreakdownValues.forEach((bValue) => {
      const {
        values: uniqueSecondBreakdownValues,
        breakdownMapper: secondBreakdownMapper
      } = getBreakdownDataMapperWithUniqueValues(
        firstBreakdownMapper[bValue],
        secondBreakdownKey
      );

      uniqueSecondBreakdownValues.forEach((sbValue, sbIndex) => {
        const row = {};
        row.index = bValue + firstBreakdownKey + sbValue + secondBreakdownKey;
        row[firstBreakdownKey] = {
          value: bValue,
          rowSpan: !sbIndex ? uniqueSecondBreakdownValues.length : 0
        };
        row[secondBreakdownKey] = { value: sbValue };
        row[thirdBreakdownKey] = {
          value: renderHorizontalBarChart(
            secondBreakdownMapper[sbValue],
            thirdBreakdownKey,
            cardSize,
            isDashboardWidget
          )
        };
        result.push(row);
      });
    });
    if (isDashboardWidget && result.length) {
      return [result[0]];
    }
    return result;
  }
  return null;
};

export const getHorizontalBarChartColumns = (
  breakdown,
  userPropNames,
  eventPropNames,
  cardSize = 1
) => {
  const result = breakdown.map((e, index) => {
    const displayTitle = getBreakdownDisplayName({
      breakdown: e,
      userPropNames,
      eventPropNames
    });

    return {
      title: <NonClickableTableHeader title={displayTitle} />,
      dataIndex: `${e.pr} - ${index}`,
      width: cardSize !== 1 ? 100 : 200,
      className: tableStyles.horizontalBarTableHeader,
      render: (d) => {
        const obj = {
          children: <div className='h-full p-6'>{d.value}</div>,
          props: has(d, 'rowSpan') ? { rowSpan: d.rowSpan } : {}
        };
        return obj;
      }
    };
  });
  if (cardSize !== 1) {
    if (cardSize === 0) {
      return result.slice(result.length - 2);
    }
    if (cardSize === 2) {
      return result.slice(result.length - 1);
    }
  }
  return result;
};
