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
  for (const d of data) {
    if (has(d, 'compareIndex')) {
      comparisonData.push(d);
      continue;
    }
    currentData.push(d);
  }
  const result = [];
  const sortedData = SortResults(currentData, sorter).slice(
    0,
    MAX_ALLOWED_VISIBLE_PROPERTIES
  );
  sortedData.forEach((d) => {
    const cmpData = comparisonData.find((cd) => cd.compareIndex === d.index);
    result.push(d, cmpData);
  });
  return result;
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

  for (const i in breakdown) {
    const bkd = breakdown[i].property;
    breakdownData[`${bkd} - ${i}`] = parseForDateTimeLabel(
      grns[i],
      breakdownVals[i]
    );
  }
  const rowLabel = Object.values(breakdownData).join(', ');
  return { rowLabel, breakdownData };
};

const getEquivalentCompareIndex = ({
  data,
  breakdown,
  label,
  grns,
  bkdIndex = 0
}) => {
  for (const i in data) {
    const row = get(data, `${i}`, []);
    const { rowLabel } = getRowLabelAndBreakdownData({
      row,
      breakdown,
      grns,
      bkdIndex
    });
    if (label === rowLabel) {
      return i;
    }
  }
  return -1;
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
  comparison_data
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
    console.log('kpi breakdown format data');
    const { headers, rows } = data[1];

    const headerSlice = headers.slice(0, breakdown.length);
    const grns = getBreakDownGranularities(headerSlice, breakdown);

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

      if (comparison_data != null) {
        const compareDataRows = get(comparison_data, `1.rows`, []);

        const compareIndex = getEquivalentCompareIndex({
          data: compareDataRows,
          breakdown: breakdown,
          label: grpLabel,
          grns
        });

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
      }
      return obj;
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
      render: (d, row) => {
        return (
          <div className="flex flex-col">
            <Text type="title" level={7} color="grey-6">
              {kpi.metricType ? (
                getFormattedKpiValue({ value: d, metricType: kpi.metricType })
              ) : (
                <NumFormat number={d} />
              )}
            </Text>
            {comparisonApplied && (
              <>
                <Text type="title" level={7} color="grey">
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
                <div className="flex col-gap-1 items-center justify-end">
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
                    type="title"
                    color={
                      row[`${kpiLabel} - ${index} - change`] < 0
                        ? 'red'
                        : 'green'
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
        );
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
  breakdown,
  comparison_data
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
      compareCategories: [],
      data: []
    };
  }
  const { headers, rows } = data[dataIndex];
  const dateIndex = headers.findIndex((h) => h === 'datetime');
  const breakdownIndex = dateIndex + 1;
  const differentDates = getDifferentDates(rows, dateIndex);
  const differentComparisonDates =
    comparison_data != null
      ? getDifferentDates(comparison_data[dataIndex].rows, dateIndex)
      : [];
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

  const comparisonData = aggregateData.map((d) => {
    return {
      name: d.label,
      data: [...initializedDatesData],
      marker: {
        enabled: false
      },
      dashStyle: 'dash',
      compareIndex: d.index,
      ...d
    };
  });

  const headerSlice = headers.slice(
    breakdownIndex,
    breakdown.length + breakdownIndex
  );
  const grns = getBreakDownGranularities(headerSlice, breakdown);
  const format = DATE_FORMATS[frequency] || DATE_FORMATS.date;

  const dateAndLabelRowIndexForComparisonData =
    comparison_data != null
      ? comparison_data[dataIndex].rows.reduce((prev, curr, currIndex) => {
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
    }

    if (comparison_data != null) {
      const dateIndex = differentDates.findIndex((dd) => dd === category);
      const compareCategory = differentComparisonDates[dateIndex];
      const compareIndex =
        dateAndLabelRowIndexForComparisonData[
          `${compareCategory}, ${breakdownJoin}`
        ];

      const compareRow =
        compareIndex != null
          ? comparison_data[dataIndex].rows[compareIndex]
          : null;
      if (comparisonData[bIdx] && compareRow != null) {
        const compareKpiVals = compareRow.slice(
          breakdown.length + breakdownIndex
        );
        comparisonData[bIdx][
          addQforQuarter(frequency) + MomentTz(category).format(format)
        ] = compareKpiVals[currentEventIndex];
        comparisonData[bIdx].data[idx] = compareKpiVals[currentEventIndex];
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
      data: resultsWithAtLeastWithOneDataPoint
    };
  }

  return {
    categories: differentDates,
    data:
      comparison_data != null
        ? [...resultantData, ...comparisonData]
        : resultantData,
    compareCategories: differentComparisonDates
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
  console.log('kpi with breakdown getDateBasedColumns');

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
      render: (d, row) => {
        return (
          <div className="flex flex-col">
            <Text type="title" level={7} color="grey-6">
              {kpi.metricType ? (
                getFormattedKpiValue({ value: d, metricType: kpi.metricType })
              ) : (
                <NumFormat number={d} />
              )}
            </Text>
            {comparisonApplied && (
              <>
                <Text type="title" level={7} color="grey">
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
                <div className="flex col-gap-1 items-center justify-end">
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
                    type="title"
                    color={
                      row[`${kpiLabel} - ${index} - change`] < 0
                        ? 'red'
                        : 'green'
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
        );
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
            key:
              addQforQuarter(frequency) +
              MomentTz(compareCategories[catIndex]).format(format) +
              ' - Change',
            type: 'percent',
            subtype: null
          },
          currentSorter,
          handleSorting,
          'right'
        ),
        className: 'text-right',
        width: frequency === 'hour' ? 200 : 150,
        dataIndex:
          addQforQuarter(frequency) +
          MomentTz(compareCategories[catIndex]).format(format) +
          ' - Change',
        render: (d) => {
          const changeIcon = (
            <SVG
              color={d > 0 ? '#5ACA89' : '#FF0000'}
              name={d > 0 ? 'arrowLift' : 'arrowDown'}
              size={16}
            />
          );
          return (
            <div className="flex col-gap-1 items-center justify-end">
              {changeIcon}
              <Text level={7} type="title" color={d < 0 ? 'red' : 'green'}>
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
