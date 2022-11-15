import React from 'react';
import cx from 'classnames';
import MomentTz from 'Components/MomentTz';
import get from 'lodash/get';
import findIndex from 'lodash/findIndex';
import has from 'lodash/has';
import { labelsObj } from '../../utils';
import {
  getClickableTitleSorter,
  SortResults,
  addQforQuarter,
  formatCount
} from '../../../../utils/dataFormatter';
import {
  Number as NumFormat,
  SVG,
  Text
} from '../../../../components/factorsComponents';
import {
  DATE_FORMATS,
  MAX_ALLOWED_VISIBLE_PROPERTIES,
  DISPLAY_PROP
} from '../../../../utils/constants';
import { renderHorizontalBarChart } from '../SingleEventMultipleBreakdown/utils';
import {
  getBreakdownDisplayName,
  parseForDateTimeLabel
} from '../eventsAnalytics.helpers';
import tableStyles from '../../../../components/DataTable/index.module.scss';
import NonClickableTableHeader from '../../../../components/NonClickableTableHeader';
import { EVENT_COUNT_KEY } from '../eventsAnalytics.constants';
import { BREAKDOWN_TYPES } from '../../constants';
import { getDifferentDates } from '../../coreQuery.helpers';
import { isNumeric } from '../../../../utils/global';

export const defaultSortProp = ({ breakdown }) => {
  const dateTimeBreakdownIndex = findIndex(
    breakdown,
    (b) => b.prop_type === BREAKDOWN_TYPES.DATETIME
  );
  if (dateTimeBreakdownIndex > -1) {
    return [
      {
        key: `${breakdown[dateTimeBreakdownIndex].property}`,
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
  const comparisonApplied = data.some((d) => has(d, 'compareIndex'));
  if (!comparisonApplied) {
    const result = SortResults(data, sorter).slice(
      0,
      MAX_ALLOWED_VISIBLE_PROPERTIES
    );
    return result;
  }
  const currentData = [];
  const comparisonData = [];
  data.forEach((d) => {
    if (has(d, 'compareIndex')) {
      comparisonData.push(d);
      return;
    }
    currentData.push(d);
  });
  const sortedData = SortResults(currentData, sorter).slice(
    0,
    MAX_ALLOWED_VISIBLE_PROPERTIES
  );
  const compareLines = [];
  sortedData.forEach((d) => {
    const cmpData = comparisonData.find((cd) => cd.compareIndex === d.index);
    compareLines.push(cmpData);
  });
  return [...sortedData, ...compareLines];
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
  const breakdownColumns = breakdown.map((e) => {
    const displayTitle =
      e.prop_category === 'user'
        ? userPropNames[e.property] || e.property
        : e.prop_category === 'event'
        ? eventPropNames[e.property] || e.property
        : e.property;
    return {
      title: getClickableTitleSorter(
        displayTitle,
        { key: e.property, type: e.prop_type, subtype: e.grn },
        currentSorter,
        handleSorting
      ),
      dataIndex: e.property,
      width: '50%',
      fixed: 'left',
      render: (d) => {
        if (
          e.prop_type === 'numerical' &&
          (typeof d === 'number' || isNumeric(d))
        ) {
          return <NumFormat number={d} />;
        }
        return d;
      }
    };
  });

  const e = events[0];

  const title = eventNames[e] || e;

  const countColumn = {
    title: getClickableTitleSorter(
      `${title}: ${labelsObj[page]}`,
      { key: EVENT_COUNT_KEY, type: 'numerical', subtype: null },
      currentSorter,
      handleSorting,
      'right'
    ),
    className: 'text-right',
    dataIndex: EVENT_COUNT_KEY,
    render: (d, row) => (
      <div className='flex flex-col'>
        <Text type='title' level={7} color='grey-6'>
          <NumFormat number={d} />
        </Text>
        {row.compareValue != null && (
          <>
            <Text type='title' level={7} color='grey'>
              <NumFormat number={row.compareValue} />
            </Text>
            <div className='flex col-gap-1 items-center justify-end'>
              <SVG
                color={row.change > 0 ? '#5ACA89' : '#FF0000'}
                name={row.change > 0 ? 'arrowLift' : 'arrowDown'}
                size={16}
              />
              <Text
                level={7}
                type='title'
                color={row.change < 0 ? 'red' : 'green'}
              >
                <NumFormat number={Math.abs(row.change)} />%
              </Text>
            </div>
          </>
        )}
      </div>
    )
  };
  return [...breakdownColumns, countColumn];
};

export const getDataInTableFormat = (data, searchText, currentSorter) => {
  const filteredData = data.filter(
    (d) => d.label.toLowerCase().indexOf(searchText.toLowerCase()) > -1
  );
  return SortResults(filteredData, currentSorter);
};

export const formatData = (data, comparisonData) => {
  if (
    !data ||
    !data.metrics ||
    !data.metrics.rows ||
    !data.metrics.rows.length
  ) {
    return [];
  }

  const breakdownIndex = 2;
  const valueIndex = 3;

  const breakdowns = data?.meta?.query?.gbp;
  const grn = data.meta?.query?.gbp[0]?.grn;
  const comparisonRows = get(comparisonData, 'metrics.rows', []);
  const compareDataLabelIndexMapper = comparisonRows.reduce(
    (prev, curr, currIndex) => {
      const labelVal = parseForDateTimeLabel(grn, curr[breakdownIndex]);
      return {
        ...prev,
        [labelVal]: currIndex
      };
    },
    {}
  );

  const result = data.metrics.rows.map((elem, index) => {
    const labelVal = parseForDateTimeLabel(grn, elem[breakdownIndex]);
    const displayLabel = DISPLAY_PROP[labelVal] || labelVal;
    const obj = {
      label: displayLabel,
      value: elem[valueIndex],
      [breakdowns[0].pr]: displayLabel,
      [EVENT_COUNT_KEY]: elem[valueIndex], // used for sorting, value key will be removed soon
      index
    };
    if (comparisonData != null) {
      const equivalentComparisonDataIndex =
        compareDataLabelIndexMapper[labelVal] > -1
          ? compareDataLabelIndexMapper[labelVal]
          : -1;
      obj.compareValue =
        equivalentComparisonDataIndex > -1
          ? comparisonData.metrics.rows[equivalentComparisonDataIndex][
              valueIndex
            ]
          : 0;

      const newVal = obj.value;
      const oldVal = obj.compareValue;

      obj.change =
        equivalentComparisonDataIndex === -1
          ? 0
          : ((newVal - oldVal) / oldVal) * 100;
    }
    return obj;
  });
  return result;
};

export const getDateBasedColumns = (
  categories,
  breakdown,
  currentSorter,
  handleSorting,
  frequency,
  userPropNames,
  eventPropNames,
  comparisonApplied,
  compareCategories
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
  const breakdownColumns = breakdown.map((e) => {
    const displayTitle =
      e.prop_category === 'user'
        ? userPropNames[e.property] || e.property
        : e.prop_category === 'event'
        ? eventPropNames[e.property] || e.property
        : e.property;

    return {
      title: getClickableTitleSorter(
        displayTitle,
        { key: e.property, type: e.prop_type, subtype: e.grn },
        currentSorter,
        handleSorting
      ),
      dataIndex: e.property,
      width: 200,
      fixed: 'left',
      render: (d) => {
        if (
          e.prop_type === 'numerical' &&
          (typeof d === 'number' || isNumeric(d))
        ) {
          return <NumFormat number={d} />;
        }
        return d;
      }
    };
  });

  const format = DATE_FORMATS[frequency] || DATE_FORMATS.date;
  const dateColumns = [];
  categories.forEach((cat, catIndex) => {
    dateColumns.push({
      title: getClickableTitleSorter(
        addQforQuarter(frequency) + MomentTz(cat).format(format),
        {
          key: addQforQuarter(frequency) + MomentTz(cat).format(format),
          type: 'numerical',
          subtype: null
        },
        currentSorter,
        handleSorting,
        'right'
      ),
      width: 150,
      className: cx('text-right', { 'border-none': comparisonApplied }),
      dataIndex: addQforQuarter(frequency) + MomentTz(cat).format(format),
      render: (d) => <NumFormat number={d} />
    });

    if (comparisonApplied) {
      dateColumns.push({
        title: getClickableTitleSorter(
          addQforQuarter(frequency) +
            MomentTz(compareCategories[catIndex]).format(format),
          {
            key:
              addQforQuarter(frequency) +
              MomentTz(compareCategories[catIndex]).format(format),
            type: 'numerical',
            subtype: null
          },
          currentSorter,
          handleSorting,
          'right'
        ),
        className: 'text-right border-none',
        width: frequency === 'hour' ? 200 : 150,
        dataIndex:
          addQforQuarter(frequency) +
          MomentTz(compareCategories[catIndex]).format(format),
        render: (d) => <NumFormat number={d} />
      });
      dateColumns.push({
        title: getClickableTitleSorter(
          'Change',
          {
            key: `${
              addQforQuarter(frequency) +
              MomentTz(compareCategories[catIndex]).format(format)
            } - Change`,
            type: 'percent',
            subtype: null
          },
          currentSorter,
          handleSorting,
          'right'
        ),
        className: 'text-right',
        width: frequency === 'hour' ? 200 : 150,
        dataIndex: `${
          addQforQuarter(frequency) +
          MomentTz(compareCategories[catIndex]).format(format)
        } - Change`,
        render: (d) => {
          const changeIcon = (
            <SVG
              color={d >= 0 ? '#5ACA89' : '#FF0000'}
              name={d >= 0 ? 'arrowLift' : 'arrowDown'}
              size={16}
            />
          );
          return (
            <div className='flex col-gap-1 items-center justify-end'>
              {changeIcon}
              <Text level={7} type='title' color={d < 0 ? 'red' : 'green'}>
                <NumFormat number={Math.abs(d)} />%
              </Text>
            </div>
          );
        }
      });
    }
  });
  return [...breakdownColumns, ...dateColumns, OverallColumn];
};

export const getDateBasedTableData = (
  seriesData,
  searchText,
  currentSorter,
  categories,
  comparisonApplied,
  compareCategories,
  frequency
) => {
  const format = DATE_FORMATS[frequency] || DATE_FORMATS.date;
  const result = seriesData
    .filter((s) => !has(s, 'compareIndex'))
    .map((sd, index) => {
      const obj = {
        index,
        Overall: sd.total,
        ...sd
      };
      const compareRow = seriesData.find(
        (elem) => elem.compareIndex === sd.index
      );
      const dateData = {};
      categories.forEach((cat, catIndex) => {
        dateData[addQforQuarter(frequency) + MomentTz(cat).format(format)] =
          sd.data[catIndex];
        if (comparisonApplied && compareRow != null) {
          const val1 = sd.data[catIndex];
          const val2 = compareRow.data[catIndex];
          dateData[
            addQforQuarter(frequency) +
              MomentTz(compareCategories[catIndex]).format(format)
          ] = val2;
          dateData[
            `${
              addQforQuarter(frequency) +
              MomentTz(compareCategories[catIndex]).format(format)
            } - Change`
          ] = formatCount(((val1 - val2) / val2) * 100);
        }
      });
      return {
        ...obj,
        ...dateData
      };
    });
  const filteredResult = result.filter((sd) =>
    sd.name.toLowerCase().includes(searchText.toLowerCase())
  );

  return SortResults(filteredResult, currentSorter);
};

export const formatDataInSeriesFormat = (
  data,
  aggregateData,
  frequency,
  comparisonData
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
      compareCategories: []
    };
  }
  const dateIndex = data.headers.findIndex((h) => h === 'datetime');
  const countIndex = data.headers.findIndex(
    (h) => h === 'count' || h === 'aggregate'
  );
  const eventIndex = data.headers.findIndex((h) => h === 'event_name');
  const breakdownIndex = eventIndex + 1;
  const differentDates = getDifferentDates({ rows: data.rows, dateIndex });

  const comparisonDataRows = get(comparisonData, `rows`, []);

  const differentComparisonDates = getDifferentDates({
    rows: comparisonDataRows,
    dateIndex
  });

  const initializedDatesData = differentDates.map(() => 0);
  const labelsMapper = {};
  const resultantData = aggregateData.map((d, index) => {
    labelsMapper[d.label] = index;
    return {
      name: d.label,
      data: [...initializedDatesData],
      index: d.index,
      marker: {
        enabled: false
      },
      ...d
    };
  });

  const resultantComparisonData = aggregateData.map((d) => ({
    name: d.label,
    data: [...initializedDatesData],
    marker: {
      enabled: false
    },
    dashStyle: 'dash',
    compareIndex: d.index,
    ...d
  }));

  const grn = data.meta?.query?.gbp[0]?.grn;

  const dateAndLabelRowIndexForComparisonData = comparisonDataRows.reduce(
    (prev, curr, currIndex) => {
      const date = curr[dateIndex];
      const labelVal = parseForDateTimeLabel(grn, curr[3]);
      return {
        ...prev,
        [`${date}, ${labelVal}`]: currIndex
      };
    },
    {}
  );
  const format = DATE_FORMATS[frequency] || DATE_FORMATS.date;

  data.rows.forEach((row) => {
    const breakdownJoin = row
      .slice(breakdownIndex, countIndex)
      .map((x) =>
        parseForDateTimeLabel(
          data.meta?.query?.gbp[0]?.grn,
          DISPLAY_PROP[x] ? DISPLAY_PROP[x] : x
        )
      )
      .join(', ');
    const bIdx = labelsMapper[breakdownJoin];
    const category = row[dateIndex];
    const idx = differentDates.indexOf(category);
    if (resultantData[bIdx]) {
      resultantData[bIdx][
        addQforQuarter(frequency) + MomentTz(category).format(format)
      ] = row[countIndex];
      resultantData[bIdx].data[idx] = row[countIndex];
    }

    if (comparisonData != null) {
      const currentDateIndex = differentDates.findIndex(
        (dd) => dd === category
      );
      const compareCategory = differentComparisonDates[currentDateIndex];
      const compareIndex =
        dateAndLabelRowIndexForComparisonData[
          `${compareCategory}, ${breakdownJoin}`
        ];
      const compareRow =
        compareIndex != null ? comparisonData.rows[compareIndex] : null;
      if (resultantComparisonData[bIdx]) {
        resultantComparisonData[bIdx][
          addQforQuarter(frequency) + MomentTz(category).format(format)
        ] = compareRow != null ? compareRow[countIndex] : 0;
        resultantComparisonData[bIdx].data[idx] =
          compareRow != null ? compareRow[countIndex] : 0;
      }
    }
  });

  return {
    categories: differentDates,
    data:
      comparisonData != null
        ? [...resultantData, ...resultantComparisonData]
        : resultantData,
    compareCategories: differentComparisonDates
  };
};

export const getDataInHorizontalBarChartFormat = (
  aggregateData,
  breakdown,
  cardSize = 1,
  isDashboardWidget = false,
  comparisonApplied = false
) => {
  const sortedData = SortResults(aggregateData, [
    {
      key: 'value',
      order: 'descend'
    }
  ]);

  const firstBreakdownKey = breakdown[0].pr;
  const row = {};

  row.index = 0;

  row[firstBreakdownKey] = renderHorizontalBarChart(
    sortedData,
    firstBreakdownKey,
    cardSize,
    isDashboardWidget,
    false,
    comparisonApplied
  );

  const result = [row];
  return result;
};

export const getHorizontalBarChartColumns = (
  breakdown,
  userPropNames,
  eventPropNames
) => {
  const result = breakdown.map((e) => {
    const displayTitle = getBreakdownDisplayName({
      breakdown: e,
      userPropNames,
      eventPropNames
    });

    return {
      title: <NonClickableTableHeader title={displayTitle} />,
      dataIndex: e.pr,
      className: tableStyles.horizontalBarTableHeader,
      render: (d) => {
        const obj = {
          children: <div className='h-full p-6'>{d}</div>
        };
        return obj;
      }
    };
  });
  return result;
};
