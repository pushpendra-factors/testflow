import React from 'react';
import {
  getClickableTitleSorter,
  SortResults,
  getBreakdownDisplayTitle,
  generateColors,
} from '../../../../utils/dataFormatter';
import { Number as NumFormat } from '../../../../components/factorsComponents';
import {
  MAX_ALLOWED_VISIBLE_PROPERTIES,
  ProfileUsersMapper,
} from '../../../../utils/constants';
import {
  getBreakdownDataMapperWithUniqueValues,
  renderHorizontalBarChart,
} from '../../EventsAnalytics/SingleEventMultipleBreakdown/utils';
import tableStyles from '../../../../components/DataTable/index.module.scss';
import { parseForDateTimeLabel } from '../../EventsAnalytics/SingleEventSingleBreakdown/utils';
import { displayName } from 'Components/FaFilterSelect/utils';

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

export const getBreakdownIndices = (headers, breakdown) => {
  const result = breakdown.map((elem) => {
    const str = elem.property;
    return headers.findIndex((elem) => elem === str);
  });
  return result;
};

export const getDateBreakdownIndices = (data, breakdown) => {
  const result = breakdown.map((elem) => {
    const str = elem.name + '_' + elem.property;
    return data.result_group[0].headers.findIndex((elem) => elem === str);
  });
  return result;
};

export const getProfileBreakDownGranularities = (
  breakDownSlice,
  breakdowns
) => {
  const grns = [];
  let brks = [...breakdowns];
  breakDownSlice.forEach((h) => {
    const brkIndex = brks.findIndex((x) => h === x.property);
    grns.push(
      brks[brkIndex].prop_type === 'datetime' && brks[brkIndex].grn
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
    const activeQuery = queries[currentEventIndex];
    const valIndex = headers.findIndex((h) => h === activeQuery);
    const breakdownIndices = getBreakdownIndices(headers, breakdown);
    const breakdownHeaders = headers.slice(breakdownIndices[0]);
    const grns = getProfileBreakDownGranularities(breakdownHeaders, breakdown);

    let result = [];
    rows.forEach((elem, index) => {
      const breakdownVals = {};
      breakdownIndices.forEach((b, bIdx) => {
        if (b > -1) {
          breakdownVals[
            `${breakdown[bIdx].property} - ${bIdx}`
          ] = parseForDateTimeLabel(
            grns[bIdx],
            displayName[elem[b]] ? displayName[elem[b]] : elem[b]
          );
        }
      });
      const color = generateColors(1);
      result.push({
        index,
        label: Object.values(breakdownVals).join(', '),
        value: elem[valIndex],
        color,
        ...breakdownVals,
      });
    });
    return result;
  } catch (err) {
    console.log(err);
    return [];
  }
};

export const getTableColumns = (
  queries,
  breakdown,
  currentEventIndex,
  currentSorter,
  handleSorting,
  eventPropNames,
  userPropNames
) => {
  const breakdownColumns = breakdown.map((e, index) => {
    const displayTitle =
      e.prop_category === 'user'
        ? userPropNames[e.property] || e.property
        : e.prop_category === 'event'
        ? eventPropNames[e.property] || `${e.property}`
        : e.property;

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

  const eventCol = {
    title: getClickableTitleSorter(
      ProfileUsersMapper[queries[currentEventIndex]],
      { key: 'value', type: 'numerical', subtype: null },
      currentSorter,
      handleSorting
    ),
    dataIndex: 'value',
    width: 150,
    render: (d) => {
      return <NumFormat number={d} />;
    },
  };

  return [...breakdownColumns, eventCol];
};

export const getTableData = (data, searchText, currentSorter) => {
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
      order: 'descend',
    },
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
      ),
    };

    const result = [row];
    return result;
  }

  const secondBreakdownKey = `${breakdown[1].property} - 1`;

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

export const getHorizontalBarChartColumns = (
  breakdown,
  userPropNames,
  eventPropNames,
  cardSize = 1
) => {
  const result = breakdown.map((e, index) => {
    const displayTitle = getBreakdownDisplayTitle(
      e,
      userPropNames,
      eventPropNames
    );

    return {
      title: displayTitle,
      dataIndex: `${e.property} - ${index}`,
      width: cardSize !== 1 ? 100 : 200,
      className: tableStyles.horizontalBarTableHeader,
      render: (d) => {
        const obj = {
          children: <div className='h-full p-6 break-all'>{d.value}</div>,
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
