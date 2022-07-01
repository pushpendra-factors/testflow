import React from 'react';
import get from 'lodash/get';
import has from 'lodash/has';
import findIndex from 'lodash/findIndex';
import moment from 'moment';

import { Number as NumFormat } from '../../../../components/factorsComponents';
import {
  SortResults,
  getClickableTitleSorter,
  addQforQuarter
} from '../../../../utils/dataFormatter';
import {
  MAX_ALLOWED_VISIBLE_PROPERTIES,
  DATE_FORMATS,
  DISPLAY_PROP,
  QUERY_TYPE_KPI
} from '../../../../utils/constants';
import { parseForDateTimeLabel } from '../../EventsAnalytics/SingleEventSingleBreakdown/utils';
import {
  getBreakDownGranularities,
  renderHorizontalBarChart,
  getBreakdownDataMapperWithUniqueValues
} from '../../EventsAnalytics/SingleEventMultipleBreakdown/utils';
import { getBreakdownDisplayName } from '../../EventsAnalytics/eventsAnalytics.helpers';
import tableStyles from '../../../../components/DataTable/index.module.scss';
import NonClickableTableHeader from '../../../../components/NonClickableTableHeader';

import { getKpiLabel, getFormattedKpiValue } from '../kpiAnalysis.helpers';
import { BREAKDOWN_TYPES } from '../../constants';

export const getDefaultSortProp = ({ kpis, breakdown }) => {
  const dateTimeBreakdownIndex = findIndex(breakdown, b => b.prop_type === BREAKDOWN_TYPES.DATETIME);
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
  if (Array.isArray(kpis) && kpis.length) {
    return [
      {
        key: `${getKpiLabel(kpis[0])} - 0`,
        type: 'numerical',
        subtype: null,
        order: 'descend'
      }
    ];
  }
  return [];
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

export const formatData = (data, kpis, breakdown, currentEventIndex) => {
  try {
    if (
      !data ||
      !Array.isArray(data) ||
      !data.length ||
      !data[1].headers ||
      !Array.isArray(data[1].headers) ||
      !data[1].headers.length ||
      !data[1].rows ||
      !Array.isArray(data[1].rows) ||
      !data[1].rows.length
    ) {
      return [];
    }
    console.log('kpi breakdown format data');
    const { headers, rows } = data[1];

    const headerSlice = headers.slice(0, breakdown.length);
    const grns = getBreakDownGranularities(headerSlice, breakdown);

    const result = rows.map((d, index) => {
      const breakdownVals = d
        .slice(0, breakdown.length)
        .map((vl) => (DISPLAY_PROP[vl] ? DISPLAY_PROP[vl] : vl));
      const breakdownData = {};
      for (let i = 0; i < breakdown.length; i++) {
        const bkd = breakdown[i].property;
        breakdownData[`${bkd} - ${i}`] = parseForDateTimeLabel(
          grns[i],
          breakdownVals[i]
        );
      }
      const kpiVals = d.slice(breakdown.length);
      const kpisData = {};
      for (let j = 0; j < kpis.length; j++) {
        kpisData[`${getKpiLabel(kpis[j])} - ${j}`] = kpiVals[j];
      }
      const grpLabel = Object.values(breakdownData).join(', ');
      return {
        label: grpLabel,
        value: kpiVals[currentEventIndex],
        metricType: get(kpis[currentEventIndex], 'metricType', null),
        index,
        ...breakdownData,
        ...kpisData
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
  kpis,
  currentSorter,
  handleSorting,
  userPropNames,
  eventPropNames
) => {
  const breakdownColumns = breakdown.map((e, index) => {
    const displayTitle = getBreakdownDisplayName({
      breakdown: e,
      userPropNames,
      eventPropNames,
      queryType: QUERY_TYPE_KPI
    });
    return {
      title: getClickableTitleSorter(
        displayTitle,
        { key: `${e.property} - ${index}`, type: e.prop_type, subtype: e.grn },
        currentSorter,
        handleSorting
      ),
      dataIndex: `${e.property} - ${index}`,
      fixed: !index ? 'left' : '',
      width: 200
    };
  });

  const kpiColumns = kpis.map((kpi, index) => {
    const kpiLabel = getKpiLabel(kpi);
    return {
      title: getClickableTitleSorter(
        kpiLabel,
        {
          key: `${kpiLabel} - ${index}`,
          type: 'numerical',
          subtype: null
        },
        currentSorter,
        handleSorting,
        'right'
      ),
      className: 'text-right',
      dataIndex: `${kpiLabel} - ${index}`,
      width: 300,
      render: (d) => {
        if (kpi.metricType) {
          return getFormattedKpiValue({ value: d, metricType: kpi.metricType });
        }
        return d ? <NumFormat number={d} /> : 0;
      }
    };
  });
  return [...breakdownColumns, ...kpiColumns];
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
    const displayTitle = getBreakdownDisplayName({
      breakdown: e,
      userPropNames,
      eventPropNames,
      queryType: QUERY_TYPE_KPI
    });

    return {
      title: <NonClickableTableHeader title={displayTitle} />,
      dataIndex: `${e.property} - ${index}`,
      width: cardSize !== 1 ? 100 : 200,
      className: tableStyles.horizontalBarTableHeader,
      render: (d) => {
        const obj = {
          children: <div className="h-full p-6">{d.value}</div>,
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

export const getDataInHorizontalBarChartFormat = (
  aggregateData,
  breakdown,
  cardSize = 1,
  isDashboardWidget = false
) => {
  console.log('kpi with breakdown getDataInHorizontalBarChartFormat');
  const sortedData = SortResults(aggregateData, {
    key: 'value',
    order: 'descend'
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
      )
    };

    const result = [row];
    return result;
  }

  const {
    values: uniqueFirstBreakdownValues,
    breakdownMapper: firstBreakdownMapper
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
    const thirdBreakdownKey = `${breakdown[2].property} - 2`;
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
  console.log('kpi with breakdown formatDataInSeriesFormat');
  const dataIndex = 0;
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
      data: []
    };
  }
  const { headers, rows } = data[dataIndex];
  const dateIndex = headers.findIndex((h) => h === 'datetime');
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
        enabled: false
      },
      ...d
    };
  });
  const headerSlice = headers.slice(
    breakdownIndex,
    breakdown.length + breakdownIndex
  );
  const grns = getBreakDownGranularities(headerSlice, breakdown);
  const format = DATE_FORMATS[frequency] || DATE_FORMATS.date;

  rows.forEach((row) => {
    const kpiVals = row.slice(breakdown.length + breakdownIndex);
    const breakdownJoin = row
      .slice(breakdownIndex, breakdown.length + breakdownIndex)
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
      ] = kpiVals[currentEventIndex];
      resultantData[bIdx].data[idx] = kpiVals[currentEventIndex];
    }
  });
  return {
    categories: differentDates,
    data: resultantData
  };
};

export const getDateBasedColumns = (
  categories,
  breakdown,
  kpis,
  currentSorter,
  handleSorting,
  frequency,
  userPropNames,
  eventPropNames
) => {
  console.log('kpi with breakdown getDateBasedColumns');

  const breakdownColumns = breakdown.map((e, index) => {
    const displayTitle = getBreakdownDisplayName({
      breakdown: e,
      userPropNames,
      eventPropNames,
      queryType: QUERY_TYPE_KPI
    });
    return {
      title: getClickableTitleSorter(
        displayTitle,
        { key: `${e.property} - ${index}`, type: e.prop_type, subtype: e.grn },
        currentSorter,
        handleSorting
      ),
      dataIndex: `${e.property} - ${index}`,
      fixed: !index ? 'left' : '',
      width: 200
    };
  });

  const kpiColumns = kpis.map((kpi, index) => {
    const kpiLabel = getKpiLabel(kpi);
    return {
      title: getClickableTitleSorter(
        kpiLabel,
        {
          key: `${kpiLabel} - ${index}`,
          type: 'numerical',
          subtype: null
        },
        currentSorter,
        handleSorting,
        'right'
      ),
      className: 'text-right',
      dataIndex: `${kpiLabel} - ${index}`,
      width: 300,
      render: (d) => {
        if (kpi.metricType) {
          return getFormattedKpiValue({ value: d, metricType: kpi.metricType });
        }
        return d ? <NumFormat number={d} /> : 0;
      }
    };
  });

  const format = DATE_FORMATS[frequency] || DATE_FORMATS.date;

  const dateColumns = categories.map((cat) => {
    return {
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
      render: (d, rowDetails) => {
        const metricType = get(rowDetails, 'metricType', null);
        return d ? (
          metricType ? (
            getFormattedKpiValue({ value: d, metricType })
          ) : (
            <NumFormat number={d} />
          )
        ) : (
          0
        );
      }
    };
  });
  return [...breakdownColumns, ...kpiColumns, ...dateColumns];
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
