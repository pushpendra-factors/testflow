import React from 'react';
import get from 'lodash/get';
import has from 'lodash/has';
import cx from 'classnames';
import findIndex from 'lodash/findIndex';
import MomentTz from 'Components/MomentTz';

import {
  Number as NumFormat,
  SVG,
  Text
} from '../../../../components/factorsComponents';
import {
  SortResults,
  getClickableTitleSorter,
  addQforQuarter,
  formatCount
} from '../../../../utils/dataFormatter';
import {
  MAX_ALLOWED_VISIBLE_PROPERTIES,
  DATE_FORMATS,
  DISPLAY_PROP,
  QUERY_TYPE_KPI
} from '../../../../utils/constants';
import {
  parseForDateTimeLabel,
  getBreakdownDisplayName
} from '../../EventsAnalytics/eventsAnalytics.helpers';
import {
  getBreakDownGranularities,
  renderHorizontalBarChart,
  getBreakdownDataMapperWithUniqueValues
} from '../../EventsAnalytics/SingleEventMultipleBreakdown/utils';
import tableStyles from '../../../../components/DataTable/index.module.scss';
import NonClickableTableHeader from '../../../../components/NonClickableTableHeader';

import { getKpiLabel, getFormattedKpiValue } from '../kpiAnalysis.helpers';
import { BREAKDOWN_TYPES } from '../../constants';
import { getDifferentDates } from '../../coreQuery.helpers';
import { isNumeric } from 'Utils/global';

export const getDefaultSortProp = ({ kpis, breakdown }) => {
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
  const result = [];
  const sortedData = SortResults(currentData, sorter).slice(
    0,
    MAX_ALLOWED_VISIBLE_PROPERTIES
  );
  sortedData.forEach((d) => {
    const cmpData = comparisonData.find((cd) => cd.compareIndex === d.index);
    result.push(cmpData);
  });
  return [...sortedData, ...result];
};

const getDefaultCompareData = ({ kpis }) => {
  const kpisData = {};

  for (let j = 0; j < kpis.length; j++) {
    const dataKey = `${getKpiLabel(kpis[j])} - ${j} - compareValue`;
    kpisData[dataKey] = 0;
    kpisData[`${getKpiLabel(kpis[j])} - ${j} - change`] = 0;
  }
  return kpisData;
};

const getRowLabelAndBreakdownData = ({
  row,
  breakdown,
  grns,
  bkdIndex = 0
}) => {
  const breakdownVals = row
    .slice(bkdIndex, bkdIndex + breakdown.length)
    .map((vl) => (DISPLAY_PROP[vl] ? DISPLAY_PROP[vl] : vl));

  const breakdownData = {};

  for (let i = 0; i < breakdown.length; i++) {
    const bkd = breakdown[i].property;
    breakdownData[`${bkd} - ${i}`] = parseForDateTimeLabel(
      grns[i],
      breakdownVals[i]
    );
  }
  const rowLabel = Object.values(breakdownData).join(', ');
  return { rowLabel, breakdownData };
};

const getKpiValues = ({
  row,
  kpis,
  breakdown,
  originalValues,
  isCompare = false
}) => {
  const kpiVals = row.slice(breakdown.length);
  const kpisData = {};

  for (let j = 0; j < kpis.length; j++) {
    const dataKey = `${getKpiLabel(kpis[j])} - ${j}${
      isCompare ? ` - compareValue` : ''
    }`;
    kpisData[dataKey] = kpiVals[j];

    if (isCompare) {
      const newVal = originalValues[j];
      const oldVal = kpiVals[j];
      kpisData[`${getKpiLabel(kpis[j])} - ${j} - change`] =
        ((newVal - oldVal) / oldVal) * 100;
    }
  }
  return { kpisData, kpiVals };
};

export const formatData = (
  data,
  kpis,
  breakdown,
  currentEventIndex,
  comparisonData
) => {
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
    const { headers, rows } = data[1];

    const headerSlice = headers.slice(0, breakdown.length);
    const grns = getBreakDownGranularities(headerSlice, breakdown);

    const comparisonDataLabelIndex = get(comparisonData, `1.rows`, []).reduce(
      (prev, curr, currIndex) => {
        const { rowLabel: compareRowLabel } = getRowLabelAndBreakdownData({
          row: curr,
          breakdown,
          grns
        });
        return {
          ...prev,
          [compareRowLabel]: currIndex
        };
      },
      {}
    );

    const result = rows.map((d, index) => {
      const { rowLabel: grpLabel, breakdownData } = getRowLabelAndBreakdownData(
        {
          row: d,
          breakdown,
          grns
        }
      );

      const { kpisData, kpiVals } = getKpiValues({ row: d, kpis, breakdown });

      const obj = {
        label: grpLabel,
        value: kpiVals[currentEventIndex],
        metricType: get(kpis[currentEventIndex], 'metricType', null),
        index,
        ...breakdownData,
        ...kpisData
      };

      if (comparisonData != null) {
        const compareDataRows = get(comparisonData, `1.rows`, []);

        const compareIndex =
          comparisonDataLabelIndex[grpLabel] != null
            ? comparisonDataLabelIndex[grpLabel]
            : -1;

        if (compareIndex > -1) {
          const compareRow = compareDataRows[compareIndex];
          const { kpiVals: compareKpiVals, kpisData: compareKpisData } =
            getKpiValues({
              row: compareRow,
              breakdown,
              kpis,
              isCompare: true,
              originalValues: kpiVals
            });
          return {
            ...obj,
            compareValue: compareKpiVals[currentEventIndex],
            ...compareKpisData
          };
        }
        return {
          ...obj,
          compareValue: 0,
          ...getDefaultCompareData({
            kpis
          })
        };
      }
      return obj;
    });
    return result;
  } catch (err) {
    return [];
  }
};

export const getTableColumns = (
  breakdown,
  kpis,
  currentSorter,
  handleSorting,
  comparisonApplied
) => {
  const breakdownColumns = breakdown.map((e, index) => {
    const displayTitle = getBreakdownDisplayName({
      breakdown: e,
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
      width: 200,
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
      render: (d, row) => (
        <div className='flex flex-col'>
          <Text type='title' level={7} color='grey-6'>
            {kpi.metricType ? (
              getFormattedKpiValue({ value: d, metricType: kpi.metricType })
            ) : (
              <NumFormat number={d} />
            )}
          </Text>
          {comparisonApplied && (
            <>
              <Text type='title' level={7} color='grey'>
                {kpi.metricType ? (
                  getFormattedKpiValue({
                    value: row[`${kpiLabel} - ${index} - compareValue`],
                    metricType: kpi.metricType
                  })
                ) : (
                  <NumFormat
                    number={row[`${kpiLabel} - ${index} - compareValue`]}
                  />
                )}
              </Text>
              <div className='flex col-gap-1 items-center justify-end'>
                <SVG
                  color={
                    row[`${kpiLabel} - ${index} - change`] >= 0
                      ? '#5ACA89'
                      : '#FF0000'
                  }
                  name={
                    row[`${kpiLabel} - ${index} - change`] >= 0
                      ? 'arrowLift'
                      : 'arrowDown'
                  }
                  size={16}
                />
                <Text
                  level={7}
                  type='title'
                  color={
                    row[`${kpiLabel} - ${index} - change`] < 0 ? 'red' : 'green'
                  }
                >
                  <NumFormat
                    number={Math.abs(row[`${kpiLabel} - ${index} - change`])}
                  />
                  %
                </Text>
              </div>
            </>
          )}
        </div>
      )
    };
  });
  return [...breakdownColumns, ...kpiColumns];
};

export const getDataInTableFormat = (data, searchText, currentSorter) => {
  const filteredData = data.filter((elem) =>
    elem.label.toLowerCase().includes(searchText.toLowerCase())
  );
  return SortResults(filteredData, currentSorter);
};

export const getHorizontalBarChartColumns = (
  breakdown,
  userPropNames,
  eventPropertiesDisplayNames,
  cardSize = 1
) => {
  const result = breakdown.map((e, index) => {
    const displayTitle = getBreakdownDisplayName({
      breakdown: e,
      userPropNames,
      eventPropertiesDisplayNames,
      queryType: QUERY_TYPE_KPI
    });

    return {
      title: <NonClickableTableHeader title={displayTitle} />,
      dataIndex: `${e.property} - ${index}`,
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

export const getDataInHorizontalBarChartFormat = (
  aggregateData,
  breakdown,
  cardSize = 1,
  isDashboardWidget = false,
  comparisonApplied = false
) => {
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
        false,
        comparisonApplied
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
          isDashboardWidget,
          true,
          comparisonApplied
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
            isDashboardWidget,
            true,
            comparisonApplied
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

// const getDifferentDates = (dataRows, dateIndex) => {
//   const differentDates = new Set();
//   dataRows.forEach((row) => {
//     differentDates.add(row[dateIndex]);
//   });
//   return Array.from(differentDates);
// };

export const formatDataInSeriesFormat = (
  data,
  aggregateData,
  currentEventIndex,
  frequency,
  breakdown,
  comparisonData
) => {
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
      compareCategories: [],
      data: []
    };
  }
  const { headers, rows } = data[dataIndex];
  const dateIndex = headers.findIndex((h) => h === 'datetime');
  const breakdownIndex = dateIndex + 1;
  const differentDates = getDifferentDates({ rows, dateIndex });
  const dateWiseTotals = Array(differentDates.length).fill(0);
  const differentComparisonDates =
    comparisonData != null
      ? getDifferentDates({ rows: comparisonData[dataIndex].rows, dateIndex })
      : [];
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

  const headerSlice = headers.slice(
    breakdownIndex,
    breakdown.length + breakdownIndex
  );
  const grns = getBreakDownGranularities(headerSlice, breakdown);
  const format = DATE_FORMATS[frequency] || DATE_FORMATS.date;

  const dateAndLabelRowIndexForComparisonData =
    comparisonData != null
      ? comparisonData[dataIndex].rows.reduce((prev, curr, currIndex) => {
          const date = curr[dateIndex];
          const { rowLabel } = getRowLabelAndBreakdownData({
            row: curr,
            breakdown,
            grns,
            bkdIndex: dateIndex + 1
          });
          return {
            ...prev,
            [`${date}, ${rowLabel}`]: currIndex
          };
        }, {})
      : {};

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
        addQforQuarter(frequency) + MomentTz(category).format(format)
      ] = kpiVals[currentEventIndex];
      resultantData[bIdx].data[idx] = kpiVals[currentEventIndex];
      dateWiseTotals[idx] += kpiVals[currentEventIndex];
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
        compareIndex != null
          ? comparisonData[dataIndex].rows[compareIndex]
          : null;
      if (resultantComparisonData[bIdx] && compareRow != null) {
        const compareKpiVals = compareRow.slice(
          breakdown.length + breakdownIndex
        );
        resultantComparisonData[bIdx][
          addQforQuarter(frequency) + MomentTz(category).format(format)
        ] = compareKpiVals[currentEventIndex];
        resultantComparisonData[bIdx].data[idx] =
          compareKpiVals[currentEventIndex];
      }
    }
  });

  if (resultantData.length > 1000) {
    const resultsWithAtLeastWithOneDataPoint = resultantData.filter((d) =>
      d.data.some((item) => item !== 0)
    );
    return {
      categories: differentDates,
      compareCategories: differentComparisonDates,
      data: resultsWithAtLeastWithOneDataPoint,
      dateWiseTotals
    };
  }

  return {
    categories: differentDates,
    data:
      comparisonData != null
        ? [...resultantData, ...resultantComparisonData]
        : resultantData,
    compareCategories: differentComparisonDates,
    dateWiseTotals
  };
};

export const getDateBasedColumns = (
  categories,
  breakdown,
  kpis,
  currentSorter,
  handleSorting,
  frequency,
  comparisonApplied,
  compareCategories
) => {
  const breakdownColumns = breakdown.map((e, index) => {
    const displayTitle = getBreakdownDisplayName({
      breakdown: e,
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
      width: 200,
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
      render: (d, row) => (
        <div className='flex flex-col'>
          <Text type='title' level={7} color='grey-6'>
            {kpi.metricType ? (
              getFormattedKpiValue({ value: d, metricType: kpi.metricType })
            ) : (
              <NumFormat number={d} />
            )}
          </Text>
          {comparisonApplied && (
            <>
              <Text type='title' level={7} color='grey'>
                {kpi.metricType ? (
                  getFormattedKpiValue({
                    value: row[`${kpiLabel} - ${index} - compareValue`],
                    metricType: kpi.metricType
                  })
                ) : (
                  <NumFormat
                    number={row[`${kpiLabel} - ${index} - compareValue`]}
                  />
                )}
              </Text>
              <div className='flex col-gap-1 items-center justify-end'>
                <SVG
                  color={
                    row[`${kpiLabel} - ${index} - change`] > 0
                      ? '#5ACA89'
                      : '#FF0000'
                  }
                  name={
                    row[`${kpiLabel} - ${index} - change`] > 0
                      ? 'arrowLift'
                      : 'arrowDown'
                  }
                  size={16}
                />
                <Text
                  level={7}
                  type='title'
                  color={
                    row[`${kpiLabel} - ${index} - change`] < 0 ? 'red' : 'green'
                  }
                >
                  <NumFormat
                    number={Math.abs(row[`${kpiLabel} - ${index} - change`])}
                  />
                  %
                </Text>
              </div>
            </>
          )}
        </div>
      )
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
      className: cx('text-right', { 'border-none': comparisonApplied }),
      width: frequency === 'hour' ? 200 : 150,
      dataIndex: addQforQuarter(frequency) + MomentTz(cat).format(format),
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
        render: (d, row) => {
          const metricType = get(row, 'metricType', null);
          return metricType ? (
            getFormattedKpiValue({ value: d, metricType })
          ) : (
            <NumFormat number={d} />
          );
        }
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
              color={d > 0 ? '#5ACA89' : '#FF0000'}
              name={d > 0 ? 'arrowLift' : 'arrowDown'}
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
  return [...breakdownColumns, ...kpiColumns, ...dateColumns];
};

export const getDateBasedTableData = (
  seriesData,
  categories,
  searchText,
  currentSorter,
  frequency,
  comparisonApplied,
  compareCategories
) => {
  const format = DATE_FORMATS[frequency] || DATE_FORMATS.date;
  const result = seriesData
    .filter((s) => !has(s, 'compareIndex'))
    .map((sd, index) => {
      const obj = {
        index,
        event: sd.name,
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
  const filteredResult = result.filter(
    (elem) => elem.event.toLowerCase().indexOf(searchText.toLowerCase()) > -1
  );
  return SortResults(filteredResult, currentSorter);
};
