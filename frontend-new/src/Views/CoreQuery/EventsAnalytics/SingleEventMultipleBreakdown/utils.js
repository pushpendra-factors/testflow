import React from 'react';
import moment from 'moment';
import {
  getClickableTitleSorter,
  SortResults,
  getBreakdownDisplayTitle,
  generateColors,
} from '../../../../utils/dataFormatter';
import { Number as NumFormat } from '../../../../components/factorsComponents';
import { parseForDateTimeLabel } from '../SingleEventSingleBreakdown/utils';
import { labelsObj } from '../../utils';
import {
  DATE_FORMATS,
  MAX_ALLOWED_VISIBLE_PROPERTIES,
} from '../../../../utils/constants';
import HorizontalBarChartCell from './HorizontalBarChartCell';
import tableStyles from '../../../../components/DataTable/index.module.scss';

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
  console.log('semb format data');
  const { headers, rows } = data.metrics;
  const eventNameIndex = headers.findIndex((header) => header === 'event_name');
  const countIndex = headers.findIndex(
    (header) => header === 'count' || header === 'aggregate'
  );

  const headerSlice = headers.slice(eventNameIndex + 1, countIndex);
  const breakdowns = data.meta.query.gbp ? [...data.meta.query.gbp] : [];
  const grns = getBreakDownGranularities(headerSlice, breakdowns);

  const result = rows.map((d, index) => {
    const breakdownVals = d.slice(eventNameIndex + 1, countIndex);
    const breakdownData = {};
    for (let i = 0; i < breakdowns.length; i++) {
      const bkd = breakdowns[i];
      breakdownData[`${bkd.pr} - ${i}`] = parseForDateTimeLabel(
        grns[i],
        breakdownVals[i]
      );
    }
    const grpLabel = Object.values(breakdownData).join(',');
    return {
      label: grpLabel,
      value: d[countIndex],
      'Event Count': d[countIndex], //used for sorting, value key will be removed soon
      index,
      ...breakdownData,
    };
  });
  return result;
};

export const getBreakDownGranularities = (breakDownSlice, breakdowns) => {
  const grns = [];
  let brks = [...breakdowns];
  breakDownSlice.forEach((h) => {
    const brkIndex = brks.findIndex((x) => h === (x.pr ? x.pr : x.property));
    grns.push(brks[brkIndex]?.grn);
    brks.splice(brkIndex, 1);
  });
  return grns;
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
  console.log('semb getTableColumns');
  const breakdownColumns = breakdown.map((e, index) => {
    let displayTitle =
      e.prop_category === 'user'
        ? userPropNames[e.property] || e.property
        : e.prop_category === 'event'
        ? eventPropNames[e.property] || `${e.property}`
        : e.property;

    if (e.eventIndex) {
      displayTitle = displayTitle + ' (event)';
    }

    return {
      title: getClickableTitleSorter(
        displayTitle,
        { key: `${e.property} - ${index}`, type: e.prop_type, subtype: e.grn },
        currentSorter,
        handleSorting
      ),
      dataIndex: `${e.property} - ${index}`,
      fixed: !index ? 'left' : '',
      width: 200,
    };
  });

  const e = events[0];
  const title = eventNames[e] || e;

  const countColumn = {
    title: getClickableTitleSorter(
      `${title}: ${labelsObj[page]}`,
      { key: 'Event Count', type: 'numerical', subtype: null },
      currentSorter,
      handleSorting
    ),
    dataIndex: 'Event Count',
    width: 150,
    render: (d) => {
      return <NumFormat number={d} />;
    },
  };

  return [...breakdownColumns, countColumn];
};

export const getDataInTableFormat = (data, searchText, currentSorter) => {
  console.log('semb getDataInTableFormat');
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
  console.log('semb getDateBasedColumns');
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
  const breakdownColumns = breakdown.map((e, index) => {
    let displayTitle =
      e.prop_category === 'user'
        ? userPropNames[e.property] || e.property
        : e.prop_category === 'event'
        ? eventPropNames[e.property] || e.property
        : e.property;

    if (e.eventIndex) {
      displayTitle = displayTitle + ' (event)';
    }

    return {
      title: getClickableTitleSorter(
        displayTitle,
        { key: `${e.property} - ${index}`, type: e.prop_type, subtype: e.grn },
        currentSorter,
        handleSorting
      ),
      dataIndex: `${e.property} - ${index}`,
      fixed: !index ? 'left' : '',
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
  return [...breakdownColumns, ...dateColumns, OverallColumn];
};

export const getDateBasedTableData = (
  seriesData,
  searchText,
  currentSorter
) => {
  console.log('semb getDateBasedTableData');
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
      data: [],
    };
  }
  console.log('semb formatDataInStackedAreaFormat');
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

  const headerSlice = data.headers.slice(breakdownIndex, countIndex);
  let breakdowns = data.meta.query.gbp ? [...data.meta.query.gbp] : [];
  let grns = getBreakDownGranularities(headerSlice, breakdowns);
  const format = DATE_FORMATS[frequency] || DATE_FORMATS['date'];

  data.rows.forEach((row) => {
    const breakdownJoin = row
      .slice(breakdownIndex, countIndex)
      .map((x, ind) => parseForDateTimeLabel(grns[ind], x))
      .join(',');
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

export const renderHorizontalBarChart = (
  data,
  key,
  cardSize = 1,
  isDashboardWidget = false,
  multipleBreakdowns = true
) => {
  const series = [
    {
      data: [],
    },
  ];
  const colors = generateColors(10);
  const categories = data.map((elem, index) => {
    series[0].data.push({
      y: elem.value,
      color: colors[index % 10],
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
    breakdownMapper,
  };
};

export const getDataInHorizontalBarChartFormat = (
  aggregateData,
  breakdown,
  cardSize = 1,
  isDashboardWidget = false
) => {
  console.log('semb getDataInHorizontalBarChartFormat');
  const sortedData = SortResults(aggregateData, [
    {
      key: 'value',
      order: 'descend',
    },
  ]);

  const firstBreakdownKey = `${breakdown[0].pr} - 0`;
  const secondBreakdownKey = `${breakdown[1].pr} - 1`;

  const {
    values: uniqueFirstBreakdownValues,
    breakdownMapper: firstBreakdownMapper,
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
        ),
      };
      return row;
    });
    if (isDashboardWidget && result.length) {
      return [result[0]];
    }
    return result;
  } else if (breakdown.length === 3) {
    const thirdBreakdownKey = `${breakdown[2].pr} - 2`;
    const result = [];
    uniqueFirstBreakdownValues.forEach((bValue) => {
      const {
        values: uniqueSecondBreakdownValues,
        breakdownMapper: secondBreakdownMapper,
      } = getBreakdownDataMapperWithUniqueValues(
        firstBreakdownMapper[bValue],
        secondBreakdownKey
      );

      uniqueSecondBreakdownValues.forEach((sbValue, sbIndex) => {
        const row = {};
        row.index = bValue + firstBreakdownKey + sbValue + secondBreakdownKey;
        row[firstBreakdownKey] = {
          value: bValue,
          rowSpan: !sbIndex ? uniqueSecondBreakdownValues.length : 0,
        };
        row[secondBreakdownKey] = { value: sbValue };
        row[thirdBreakdownKey] = {
          value: renderHorizontalBarChart(
            secondBreakdownMapper[sbValue],
            thirdBreakdownKey,
            cardSize,
            isDashboardWidget
          ),
        };
        result.push(row);
      });
    });
    if (isDashboardWidget && result.length) {
      return [result[0]];
    }
    return result;
  }
};

export const getHorizontalBarChartColumns = (
  breakdown,
  userPropNames,
  eventPropNames,
  cardSize = 1
) => {
  console.log('semb getHorizontalBarChartColumns');
  const result = breakdown.map((e, index) => {
    const displayTitle = getBreakdownDisplayTitle(
      e,
      userPropNames,
      eventPropNames
    );

    return {
      title: displayTitle,
      dataIndex: `${e.pr} - ${index}`,
      width: cardSize !== 1 ? 100 : 200,
      className: tableStyles.horizontalBarTableHeader,
      render: (d) => {
        const obj = {
          children: <div className='h-full p-6'>{d.value}</div>,
          props: d.hasOwnProperty('rowSpan') ? { rowSpan: d.rowSpan } : {},
        };
        return obj;
      },
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
