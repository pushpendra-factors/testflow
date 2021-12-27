import React from 'react';
import moment from 'moment';
import { Number as NumFormat } from '../../../../components/factorsComponents';
import {
  SortResults,
  getClickableTitleSorter,
} from '../../../../utils/dataFormatter';
import {
  MAX_ALLOWED_VISIBLE_PROPERTIES,
  DATE_FORMATS,
} from '../../../../utils/constants';
import { parseForDateTimeLabel } from '../../EventsAnalytics/SingleEventSingleBreakdown/utils';
import {
  getBreakDownGranularities,
  renderHorizontalBarChart,
  getBreakdownDataMapperWithUniqueValues,
} from '../../EventsAnalytics/SingleEventMultipleBreakdown/utils';
import tableStyles from '../../../../components/DataTable/index.module.scss';

export const defaultSortProp = () => {
  return [
    {
      order: 'descend',
      key: 'value',
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

export const formatData = (data, breakdown, currentEventIndex) => {
  try {
    const dataIndex = currentEventIndex * 2;
    if (
      !data ||
      !Array.isArray(data) ||
      !data.length ||
      !data[dataIndex] ||
      !data[dataIndex].headers ||
      !Array.isArray(data[dataIndex].headers) ||
      !data[dataIndex].headers.length ||
      !data[dataIndex].rows ||
      !Array.isArray(data[dataIndex].rows) ||
      !data[dataIndex].rows.length
    ) {
      return [];
    }
    console.log('kpi breakdown format data');
    const { headers, rows } = data[dataIndex];
    const countIndex = headers.findIndex((header) => header === 'aggregate');

    const headerSlice = headers.slice(0, countIndex);
    const grns = getBreakDownGranularities(headerSlice, breakdown);

    const result = rows.map((d, index) => {
      const breakdownVals = d.slice(0, countIndex);
      const breakdownData = {};
      for (let i = 0; i < breakdown.length; i++) {
        const bkd = breakdown[i].property;
        breakdownData[`${bkd} - ${i}`] = parseForDateTimeLabel(
          grns[i],
          breakdownVals[i]
        );
      }
      const grpLabel = Object.values(breakdownData).join(',');
      return {
        label: grpLabel,
        value: d[countIndex],
        index,
        ...breakdownData,
      };
    });
    return result;
  } catch (err) {
    console.log(err);
    return [];
  }
};

export const getTableColumns = (
  breakdown,
  currentSorter,
  handleSorting,
  eventNames,
  userPropNames,
  eventPropNames
) => {
  const breakdownColumns = breakdown.map((e, index) => {
    return {
      title: getClickableTitleSorter(
        e.property,
        { key: `${e.property} - ${index}`, type: e.prop_type, subtype: e.grn },
        currentSorter,
        handleSorting
      ),
      dataIndex: `${e.property} - ${index}`,
      fixed: !index ? 'left' : '',
      width: 200,
    };
  });
  const valueCol = {
    title: getClickableTitleSorter(
      'Value',
      { key: `value`, type: 'numerical', subtype: null },
      currentSorter,
      handleSorting
    ),
    dataIndex: `value`,
    width: 200,
    render: (d) => {
      return d ? <NumFormat number={d} /> : 0;
    },
  };
  return [...breakdownColumns, valueCol];
};

export const getDataInTableFormat = (data, searchText, currentSorter) => {
  console.log('kpi breakdown getDataInTableFormat');
  const filteredData = data.filter((elem) =>
    elem.label.toLowerCase().includes(searchText.toLowerCase())
  );
  return SortResults(filteredData, currentSorter);
};

export const getHorizontalBarChartColumns = (
  breakdown,
  userPropNames,
  eventPropNames,
  cardSize = 1
) => {
  console.log('kpi with breakdown getHorizontalBarChartColumns');
  const result = breakdown.map((e, index) => {
    // const displayTitle = getBreakdownDisplayTitle(
    //   e,
    //   userPropNames,
    //   eventPropNames
    // );

    const displayTitle = e.property;

    return {
      title: displayTitle,
      dataIndex: `${e.property} - ${index}`,
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

export const getDataInHorizontalBarChartFormat = (
  aggregateData,
  breakdown,
  cardSize = 1,
  isDashboardWidget = false
) => {
  console.log('kpi with breakdown getDataInHorizontalBarChartFormat');
  const sortedData = SortResults(aggregateData, {
    key: 'value',
    order: 'descend',
  });

  const firstBreakdownKey = `${breakdown[0].property} - 0`;

  if (breakdown.length === 1) {
    const row = {};

    row.index = 0;

    row[firstBreakdownKey] = {
      value: renderHorizontalBarChart(
        sortedData,
        firstBreakdownKey,
        cardSize,
        isDashboardWidget,
        false
      ),
    };

    const result = [row];
    return result;
  }

  const {
    values: uniqueFirstBreakdownValues,
    breakdownMapper: firstBreakdownMapper,
  } = getBreakdownDataMapperWithUniqueValues(sortedData, firstBreakdownKey);

  const secondBreakdownKey = `${breakdown[1].property} - 1`;

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
  }

  if (breakdown.length === 3) {
    const thirdBreakdownKey = `${breakdown[2].property} - 2`;
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

const getDifferentDates = (dataRows, dateIndex) => {
  const differentDates = new Set();
  dataRows.forEach((row) => {
    differentDates.add(row[dateIndex]);
  });
  return Array.from(differentDates);
};

export const formatDataInSeriesFormat = (
  data,
  aggregateData,
  currentEventIndex,
  frequency,
  breakdown
) => {
  // console.log('kpi with breakdown formatDataInSeriesFormat');
  const dataIndex = currentEventIndex * 2 + 1;
  // console.log('dataIndex', dataIndex);
  if (
    !aggregateData.length ||
    !data[dataIndex] ||
    !data[dataIndex].headers ||
    !Array.isArray(data[dataIndex].headers) ||
    !data[dataIndex].headers.length ||
    !data[dataIndex].rows ||
    !Array.isArray(data[dataIndex].rows) ||
    !data[dataIndex].rows.length
  ) {
    return {
      categories: [],
      data: [],
    };
  }
  const { headers, rows } = data[dataIndex];
  const dateIndex = headers.findIndex((h) => h === 'datetime');
  const countIndex = headers.findIndex(
    (h) => h === 'count' || h === 'aggregate'
  );
  const breakdownIndex = dateIndex + 1;
  const differentDates = getDifferentDates(rows, dateIndex);
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
  const headerSlice = headers.slice(breakdownIndex, countIndex);
  const grns = getBreakDownGranularities(headerSlice, breakdown);
  const format = DATE_FORMATS[frequency] || DATE_FORMATS['date'];

  rows.forEach((row) => {
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

export const getDateBasedColumns = (
  categories,
  breakdown,
  currentSorter,
  handleSorting,
  frequency,
  userPropNames,
  eventPropNames
) => {
  console.log('kpi with breakdown getDateBasedColumns');
  const OverallColumn = {
    title: getClickableTitleSorter(
      'Overall',
      { key: `value`, type: 'numerical', subtype: null },
      currentSorter,
      handleSorting
    ),
    dataIndex: `value`,
    width: 150,
  };
  const breakdownColumns = breakdown.map((e, index) => {
    return {
      title: getClickableTitleSorter(
        e.property,
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
        return d ? <NumFormat number={d} /> : 0;
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
  console.log('kpi with breakdown getDateBasedTableData');
  const result = seriesData.filter((sd) =>
    sd.name.toLowerCase().includes(searchText.toLowerCase())
  );

  return SortResults(result, currentSorter);
};
