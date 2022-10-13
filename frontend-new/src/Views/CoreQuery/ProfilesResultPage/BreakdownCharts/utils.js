import React from 'react';
import { findIndex, get, has } from 'lodash';
import {
  getClickableTitleSorter,
  SortResults,
  generateColors
} from '../../../../utils/dataFormatter';
import { Number as NumFormat } from '../../../../components/factorsComponents';
import {
  MAX_ALLOWED_VISIBLE_PROPERTIES,
  ReverseProfileMapper,
  DISPLAY_PROP,
  QUERY_TYPE_PROFILE
} from '../../../../utils/constants';
import {
  getBreakdownDataMapperWithUniqueValues,
  renderHorizontalBarChart
} from '../../EventsAnalytics/SingleEventMultipleBreakdown/utils';
import { getBreakdownDisplayName , parseForDateTimeLabel } from '../../EventsAnalytics/eventsAnalytics.helpers';
import tableStyles from '../../../../components/DataTable/index.module.scss';
import NonClickableTableHeader from '../../../../components/NonClickableTableHeader';
import { BREAKDOWN_TYPES } from '../../constants';

export const defaultSortProp = ({ breakdown }) => {
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
  return [
    {
      order: 'descend',
      key: 'value',
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

export const getBreakdownIndices = (headers, breakdown) => {
  const result = breakdown.map((elem) => {
    const str = elem.property;
    const strIndex = headers.findIndex((elem) => elem === str);
    return strIndex;
  });
  return result;
};

export const getDateBreakdownIndices = (data, breakdown) => {
  const result = breakdown.map((elem) => {
    const str = `${elem.name  }_${  elem.property}`;
    const strIndex = data.result_group[0].headers.findIndex(
      (elem) => elem === str
    );
    return strIndex;
  });
  return result;
};

export const getProfileBreakDownGranularities = (
  breakDownSlice,
  breakdowns
) => {
  const grns = [];
  const brks = [...breakdowns];
  breakDownSlice.forEach((h) => {
    const brkIndex = brks.findIndex((x) => h === x.property);
    grns.push(
      brks[brkIndex]?.prop_type === 'datetime' && brks[brkIndex]?.grn
        ? brks[brkIndex].grn
        : undefined
    );
    brks.splice(brkIndex, 1);
  });
  return grns;
};

export const formatData = (data, breakdown, queries, currentEventIndex) => {
  if (
    !data ||
    !data.result_group ||
    !Array.isArray(data.result_group) ||
    !data.result_group[currentEventIndex] ||
    !data.result_group[currentEventIndex].headers ||
    !Array.isArray(data.result_group[currentEventIndex].headers) ||
    !data.result_group[currentEventIndex].headers.length ||
    !data.result_group[currentEventIndex].rows ||
    !Array.isArray(data.result_group[currentEventIndex].rows) ||
    !data.result_group[currentEventIndex].rows.length
  ) {
    return [];
  }
  try {
    const { headers, rows } = data.result_group[currentEventIndex];
    // const activeQuery = queries[currentEventIndex];
    const valIndex = headers.findIndex(
      (elem) => elem === 'count' || elem === 'aggregate' || elem === 'all_users'
    );
    const breakdownIndices = getBreakdownIndices(headers, breakdown);
    const breakdownHeaders = headers.slice(breakdownIndices[0]);
    const grns = getProfileBreakDownGranularities(breakdownHeaders, breakdown);

    const result = [];
    rows.forEach((elem, index) => {
      const breakdownVals = {};
      breakdownIndices.forEach((b, bIdx) => {
        if (b > -1) {
          breakdownVals[`${breakdown[bIdx].property} - ${bIdx}`] =
            parseForDateTimeLabel(
              grns[bIdx],
              DISPLAY_PROP[elem[b]] ? DISPLAY_PROP[elem[b]] : elem[b]
            );
        }
      });
      const color = generateColors(1);
      result.push({
        index,
        label: Object.values(breakdownVals).join(),
        color,
        ...breakdownVals,
        value: elem[valIndex]
      });
    });
    return result;
  } catch (err) {
    console.log(err);
    return [];
  }
};

export const getProfileQueryDisplayName = ({ query, groupAnalysis }) => get(ReverseProfileMapper, `${query}.${groupAnalysis}`, query);

export const getTableColumns = (
  queries,
  breakdown,
  groupAnalysis,
  currentEventIndex,
  currentSorter,
  handleSorting,
  eventPropNames,
  userPropNames
) => {
  const breakdownColumns = breakdown.map((e, index) => {
    const displayTitle = getBreakdownDisplayName({
      breakdown: e,
      userPropNames,
      eventPropNames,
      queryType: QUERY_TYPE_PROFILE
    });

    return {
      title: getClickableTitleSorter(
        <div className="break-all">{displayTitle}</div>,
        { key: `${e.property} - ${index}`, type: e.prop_type, subtype: e.grn },
        currentSorter,
        handleSorting
      ),
      dataIndex: `${e.property} - ${index}`,
      fixed: !index ? 'left' : '',
      width: 200
    };
  });

  const queryDisplayName = getProfileQueryDisplayName({
    query: queries[currentEventIndex],
    groupAnalysis
  });

  const eventCol = {
    title: getClickableTitleSorter(
      queryDisplayName,
      { key: 'value', type: 'numerical', subtype: null },
      currentSorter,
      handleSorting,
      'right'
    ),
    className: 'text-right',
    dataIndex: 'value',
    width: 150,
    render: (d) => <NumFormat number={d} />
  };

  return [...breakdownColumns, eventCol];
};

export const getTableData = (
  data,
  searchText,
  currentSorter
  // queries,
  // currentEventIndex,
  // groupAnalysis
) => {
  const filteredData = data.filter((elem) =>
    elem.label.toLowerCase().includes(searchText.toLowerCase())
  );
  return SortResults(filteredData, currentSorter);
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

  const secondBreakdownKey = `${breakdown[1].property} - 1`;

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
  } if (breakdown.length === 3) {
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
      dataIndex: `${e.property} - ${index}`,
      width: cardSize !== 1 ? 100 : 200,
      className: tableStyles.horizontalBarTableHeader,
      render: (d) => {
        const obj = {
          children: <div className="h-full p-6 break-all">{d.value}</div>,
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
