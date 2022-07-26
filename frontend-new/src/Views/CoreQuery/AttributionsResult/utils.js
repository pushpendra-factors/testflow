import React from 'react';
import cx from 'classnames';
import moment from 'moment';
import { SVG, Number as NumFormat, Text } from 'factorsComponents';
import {
  SortData,
  formatCount,
  getClickableTitleSorter,
  SortResults
} from '../../../utils/dataFormatter';
import {
  ATTRIBUTION_METHODOLOGY,
  FIRST_METRIC_IN_ATTR_RESPOSE,
  ARR_JOINER,
  ATTRIBUTION_METRICS,
  DISPLAY_PROP
} from '../../../utils/constants';
import styles from './index.module.scss';
import { keys, uniqBy } from 'lodash';
import { ATTRIBUTION_GROUP_ANALYSIS_KEYS } from './attributionsResult.constants';

export const defaultSortProp = () => {
  return [
    {
      order: 'descend',
      key: 'Conversion',
      type: 'numerical',
      subtype: null
    }
  ];
};

export const getDifferentCampaingns = (data) => {
  const { headers } = data.result;
  const campaignIdx = headers.indexOf('Campaign');
  const differentCampaigns = new Set();
  data.result.rows.forEach((row) => {
    differentCampaigns.add(row[campaignIdx]);
  });
  return Array.from(differentCampaigns);
};

const getBarLineChartSeriesKeys = ({
  attrQueries,
  groupAnalysis,
  currentEventIndex = 0,
  headers
}) => {
  if (
    groupAnalysis &&
    groupAnalysis !== ATTRIBUTION_GROUP_ANALYSIS_KEYS.USERS
  ) {
    const query = attrQueries[currentEventIndex];
    const result = [];
    if (query) {
      if (headers.includes(`${query.label} - Conversion Value`)) {
        result.push(`${query.label} - Conversion Value`);
      } else {
        result.push(`${query.label} - Conversion`);
      }
      if (headers.includes(`${query.label} - Return on Cost`)) {
        result.push(`${query.label} - Return on Cost`);
      } else {
        result.push(`${query.label} - Cost Per Conversion`);
      }
    }
    return result;
  }
  return ['Conversion', 'Cost Per Conversion'];
};

const getLegendsLabel = ({ key }) => {
  if (key === 'Conversion') {
    return 'Conversions as Unique users';
  }
  return key;
};

export const getSingleTouchPointChartData = (
  data,
  visibleIndices,
  attr_dimensions,
  content_groups,
  touchPoint,
  isComparisonApplied,
  attrQueries,
  groupAnalysis,
  currentEventIndex = 0
) => {
  const seriesKeys = getBarLineChartSeriesKeys({
    attrQueries,
    groupAnalysis,
    currentEventIndex,
    headers: keys(data[0])
  });
  const listDimensions =
    touchPoint === 'LandingPage'
      ? content_groups.slice()
      : attr_dimensions.slice();
  const enabledDimensions = listDimensions.filter(
    (d) => d.touchPoint === touchPoint && d.enabled
  );
  const slicedTableData = data.filter(
    (d) => visibleIndices.indexOf(d.index) > -1
  );

  const categories = slicedTableData.map((d) => {
    const cat = enabledDimensions.length
      ? enabledDimensions.map((dimension) => {
          return d[dimension.title];
        })
      : [d[touchPoint]];
    return cat.join(', ');
  });

  const series = [
    {
      type: 'column',
      yAxis: 0,
      data: slicedTableData.map((row) =>
        isComparisonApplied
          ? Number(row[seriesKeys[0]].value)
          : Number(row[[seriesKeys[0]]])
      ),
      color: '#4d7db4'
    },
    {
      type: 'line',
      yAxis: 1,
      data: slicedTableData.map((row) =>
        isComparisonApplied
          ? Number(row[seriesKeys[1]].value)
          : Number(row[seriesKeys[1]])
      ),
      color: '#d4787d',
      marker: {
        symbol: 'circle'
      }
    }
  ];
  if (isComparisonApplied) {
    series.push({
      type: 'column',
      yAxis: 0,
      data: slicedTableData.map((row) =>
        Number(row[seriesKeys[0]].compare_value)
      ),
      color: '#4d7db4'
    });
    series.push({
      type: 'line',
      yAxis: 1,
      data: slicedTableData.map((row) =>
        Number(row[seriesKeys[1]].compare_value)
      ),
      color: '#d4787d',
      marker: {
        symbol: 'circle'
      },
      dashStyle: 'dash'
    });
    const temp = series[1];
    series[1] = series[2];
    series[2] = temp;
  }

  const legends = [
    getLegendsLabel({ key: seriesKeys[0] }),
    getLegendsLabel({ key: seriesKeys[1] })
  ];

  return {
    categories,
    series,
    legends
  };
};

export const getDualTouchPointChartData = (
  data,
  visibleIndices,
  attr_dimensions,
  content_groups,
  touchpoint,
  attribution_method,
  attribution_method_compare,
  currMetricsValue
) => {
  const listDimensions =
    touchpoint === 'LandingPage'
      ? content_groups.slice()
      : attr_dimensions.slice();
  const enabledDimensions = listDimensions.filter(
    (d) => d.touchPoint === touchpoint && d.enabled
  );
  const slicedTableData = data.filter(
    (d) => visibleIndices.indexOf(d.index) > -1
  );
  const result = slicedTableData.map((d) => {
    const name = enabledDimensions.length
      ? enabledDimensions.map((dimension) => {
          return d[dimension.title];
        })
      : [d[touchpoint]];
    return {
      name: name.join(', '),
      [attribution_method]: !currMetricsValue
        ? d.Conversion
        : d['Cost Per Conversion'],
      [attribution_method_compare]: !currMetricsValue
        ? d.conversion_compare
        : d.cost_compare
    };
  });
  return result;
};

export const formatData = (
  data,
  touchPoint,
  event,
  attr_dimensions,
  content_groups,
  comparison_data
) => {
  if (
    !data ||
    !data.headers ||
    !Array.isArray(data.headers) ||
    !data.headers.length ||
    !data.rows ||
    !Array.isArray(data.rows) ||
    !data.rows.length
  ) {
    return {
      categories: [],
      series: []
    };
  }
  const { headers, rows } = data;
  const touchpointIdx = headers.indexOf(touchPoint);

  const listDimensions =
    touchPoint === 'LandingPage'
      ? content_groups.slice()
      : attr_dimensions.slice();
  const enabledDimensions = listDimensions.filter(
    (d) => d.touchPoint === touchPoint && d.enabled
  );
  let categories;
  if (enabledDimensions.length) {
    const firstDimensionIdx = headers.findIndex(
      (h) => h === enabledDimensions[0].responseHeader
    );
    const lastDimensionIdx = headers.findIndex(
      (h) =>
        h === enabledDimensions[enabledDimensions.length - 1].responseHeader
    );
    categories = rows.map((row) => {
      return row.slice(firstDimensionIdx, lastDimensionIdx + 1).join(', ');
    });
  } else {
    categories = rows.map((row) => {
      return row[touchpointIdx];
    });
  }
  const conversionIdx = headers.findIndex((h) => h === `${event} - Users`);
  const costIdx = headers.findIndex((h) => h === 'Cost Per Conversion');
  const equivalentIndicesMapper = comparison_data
    ? getEquivalentIndicesMapper(data, comparison_data)
    : {};
  const series = [
    {
      type: 'column',
      yAxis: 0,
      data: rows.map((row) => row[conversionIdx]),
      color: '#4d7db4'
    },
    {
      type: 'line',
      yAxis: 1,
      data: rows.map((row) => row[costIdx]),
      color: '#d4787d',
      marker: {
        symbol: 'circle'
      }
    }
  ];
  if (comparison_data) {
    series.push({
      type: 'column',
      yAxis: 0,
      data: rows.map((_, index) => {
        const equivalent_compare_row =
          equivalentIndicesMapper[index] > -1
            ? comparison_data.rows[equivalentIndicesMapper[index]]
            : null;
        return equivalent_compare_row
          ? equivalent_compare_row[conversionIdx]
          : 0;
      }),
      color: '#4d7db4'
    });
    series.push({
      type: 'line',
      yAxis: 1,
      data: rows.map((_, index) => {
        const equivalent_compare_row =
          equivalentIndicesMapper[index] > -1
            ? comparison_data.rows[equivalentIndicesMapper[index]]
            : null;
        return equivalent_compare_row ? equivalent_compare_row[costIdx] : 0;
      }),
      color: '#d4787d',
      marker: {
        symbol: 'circle'
      },
      dashStyle: 'dash'
    });
    const temp = series[1];
    series[1] = series[2];
    series[2] = temp;
  }
  return {
    categories,
    series
  };
};

export const formatGroupedData = (
  data,
  event,
  visibleIndices,
  attribution_method,
  attribution_method_compare,
  currMetricsValue
) => {
  const { headers } = data;
  const str = currMetricsValue ? 'Cost Per Conversion' : `${event} - Users`;
  const compareStr = currMetricsValue
    ? 'Compare Cost Per Conversion'
    : 'Compare - Users';
  const userIdx = headers.indexOf(str);
  const compareUsersIdx = headers.indexOf(compareStr);
  let rows = data.rows.filter((_, index) => visibleIndices.indexOf(index) > -1);
  rows = SortData(rows, userIdx, 'descend');
  const chartData = rows.map((row) => {
    return {
      name: row[0],
      [attribution_method]: row[userIdx],
      [attribution_method_compare]: row[compareUsersIdx]
    };
  });
  return chartData;
};

const firstColumn = (d, durationObj, cmprDuration) => {
  if (cmprDuration) {
    return (
      <div className="flex items-center">
        <Text
          type="title"
          weight="normal"
          color="grey-8"
          extraClass="text-sm mb-0 py-2 px-4 w-1/2"
        >
          {d}
        </Text>
        <div
          style={{ borderLeft: '1px solid #E7E9ED' }}
          className="flex py-2 flex-col px-4 w-1/2"
        >
          <Text
            type="title"
            weight="normal"
            color="grey-8"
            extraClass="text-sm mb-0"
          >
            {`${moment(durationObj.from).format('MMM DD')} - ${moment(
              durationObj.to
            ).format('MMM DD')}`}
          </Text>
          <Text
            type="title"
            weight="normal"
            color="grey"
            extraClass="text-xs mb-0"
          >{`vs ${moment(cmprDuration.from).format('MMM DD')} - ${moment(
            cmprDuration.to
          ).format('MMM DD')}`}</Text>
        </div>
      </div>
    );
  }
  return d;
};

const renderMetric = (d, comparison_data) => {
  if (!comparison_data) {
    return <NumFormat number={d} />;
  }
  const changePercent = calcChangePerc(d.value, d.compare_value);
  let compareText = null;
  if (isNaN(changePercent) || changePercent === 0) {
    compareText = (
      <>
        <NumFormat number={0} />%
      </>
    );
  } else if (changePercent === 'Infinity') {
    compareText = (
      <>
        <SVG color="#5ACA89" name={'arrowLift'} size={16}></SVG>
        <span>&#8734; %</span>
      </>
    );
  } else {
    compareText = (
      <>
        <SVG
          color={changePercent > 0 ? '#5ACA89' : '#FF0000'}
          name={changePercent > 0 ? 'arrowLift' : 'arrowDown'}
          size={16}
        ></SVG>
        <NumFormat number={Math.abs(changePercent)} />%
      </>
    );
  }
  return (
    <div className="flex gap-x-2 justify-end items-start">
      <div className="flex flex-col items-center">
        <div>
          <NumFormat number={d.value} />
        </div>
        <div className={styles.compareNumber}>
          <NumFormat number={d.compare_value} />
        </div>
      </div>
      <div className={styles.changePercent}>{compareText}</div>
    </div>
  );
};

export const getTableColumns = (
  currentSorter,
  handleSorting,
  attribution_method,
  attribution_method_compare,
  touchpoint,
  linkedEvents,
  event,
  eventNames,
  metrics,
  attr_dimensions,
  content_groups,
  durationObj,
  comparison_data,
  cmprDuration,
  queryOptions,
  attrQueries,
  data
) => {
  if (!data) {
    return [];
  }
  const { headers } = data;

  const getEventColumnConfig = ({ title, key, method, hasBorder = false }) => {
    return {
      title: getClickableTitleSorter(
        <div className="flex flex-col items-start justify-center">
          <div>{title}</div>
          {!!method && (
            <div
              className={cx('w-full text-right', styles.attributionMethodLabel)}
            >
              {ATTRIBUTION_METHODOLOGY.find((m) => m.value === method).text}
            </div>
          )}
        </div>,
        { key, type: 'numerical', subtype: null },
        currentSorter,
        handleSorting,
        'right'
      ),
      className: cx('text-right', { 'border-none': !hasBorder }),
      dataIndex: key,
      width: 200,
      render: (d) => {
        return renderMetric(d, comparison_data);
      }
    };
  };

  const getDimensionsColConfig = (d, index) => {
    return {
      title: getClickableTitleSorter(
        d.title,
        { key: d.title, type: 'categorical', subtype: null },
        currentSorter,
        handleSorting,
        'left',
        'end',
        'pb-3'
      ),
      dataIndex: d.title,
      fixed: !index ? 'left' : '',
      width: comparison_data && !index ? 300 : 200,
      className: cx({ [styles.touchPointCol]: comparison_data && !index }),
      render: (d) =>
        !index
          ? firstColumn(d, durationObj, comparison_data ? cmprDuration : null)
          : d
    };
  };

  const listDimensions =
    touchpoint === 'LandingPage' ? [...content_groups] : [...attr_dimensions];

  const enabledDimensions = listDimensions.filter(
    (d) => d.touchPoint === touchpoint && d.enabled
  );

  let dimensionColumns;

  if (enabledDimensions.length) {
    dimensionColumns = enabledDimensions.map(getDimensionsColConfig);
  } else {
    dimensionColumns = [
      {
        title: getClickableTitleSorter(
          touchpoint === 'ChannelGroup' ? 'Channel' : touchpoint,
          { key: touchpoint, type: 'categorical', subtype: null },
          currentSorter,
          handleSorting,
          'left',
          'end',
          'pb-3'
        ),
        dataIndex: touchpoint,
        fixed: 'left',
        width: 200,
        render: (d) =>
          firstColumn(d, durationObj, comparison_data ? cmprDuration : null)
      }
    ];
  }

  const metricsColumns = metrics
    .filter((metric) => metric.enabled && !metric.isEventMetric)
    .map((metric) => {
      return {
        title: getClickableTitleSorter(
          metric.title,
          { key: metric.title, type: 'numerical', subtype: null },
          currentSorter,
          handleSorting,
          'right',
          'end',
          'pb-3'
        ),
        dataIndex: metric.title,
        width: 180,
        className: 'text-right',
        render: (d) => {
          return renderMetric(d, comparison_data);
        }
      };
    });

  const showCPC = metrics.find((elem) => elem.header === 'ALL CPC')?.enabled;
  const showCR = metrics.find((elem) => elem.header === 'ALL CR')?.enabled;
  const showCV = metrics.find((elem) => elem.header === 'CV')?.enabled;
  const showROC = metrics.find((elem) => elem.header === 'ROC')?.enabled;

  const conversionBorderCondition = !showCPC && !showCR;
  const costBorderCondition = !showCR;
  const eventColumns = [];
  let attrQueryEvents = [];

  if (
    queryOptions.group_analysis &&
    queryOptions.group_analysis !== 'users' &&
    headers.length
  ) {
    attrQueryEvents = attrQueries.map((q) => {
      const lbl = q.label;
      let attrQueryHeaders = headers.filter((h) => h.startsWith(lbl));
      if (!attribution_method_compare) {
        attrQueryHeaders = attrQueryHeaders.filter(
          (hd) => hd.search('(compare)') < 0
        );
      }
      const attrChildren = attrQueryHeaders
        .filter((hd) => {
          const title = hd.split(' - ')[1];
          if (
            title === 'Conversion Value' ||
            title === 'Conversion Value(compare)'
          ) {
            return showCV;
          }
          if (
            title === ('Return on Cost' || title === 'Return on Cost(compare)')
          ) {
            return showROC;
          }

          return hd;
        })
        .map((hd) => {
          let title = hd.split(' - ')[1];
          let attrMetod = attribution_method;
          if (hd.search('UserConversionRate') >= 0) {
            title = title.replace('UserConversionRate', 'Conversion Rate');
          }

          if (hd.search('compare') >= 0) {
            attrMetod = attribution_method_compare;
          }

          return getEventColumnConfig({
            title,
            key: hd,
            method: attrMetod,
            hasBorder: true
          });
        });

      // const attrChildren = [
      //   getEventColumnConfig({
      //     title: 'Conversion',
      //     key: lbl + ' - Conversion',
      //     hasBorder: conversionBorderCondition,
      //   }),
      // ];
      // if (showCPC) {
      //   attrChildren.push(
      //     getEventColumnConfig({
      //       title: 'Cost Per Conversion',
      //       key: lbl + ' - Cost Per Conversion',
      //       hasBorder: costBorderCondition,
      //     })
      //   );
      // }
      // if (showCR) {
      //   attrChildren.push(
      //     getEventColumnConfig({
      //       title: 'Conversion Rate',
      //       key: lbl + ' - UserConversionRate(%)',
      //       hasBorder: true,
      //     })
      //   );
      // }
      return {
        title: eventNames[lbl] || lbl,
        className: 'bg-white tableParentHeader ',
        children: attrChildren
      };
    });
  } else {
    eventColumns.push(
      getEventColumnConfig({
        title: 'Conversion',
        key: 'Conversion',
        method: attribution_method,
        hasBorder: conversionBorderCondition
      })
    );

    if (showCPC) {
      eventColumns.push(
        getEventColumnConfig({
          title: 'Cost Per Conversion',
          key: 'Cost Per Conversion',
          method: attribution_method,
          hasBorder: costBorderCondition
        })
      );
    }
    if (showCR) {
      eventColumns.push(
        getEventColumnConfig({
          title: 'Conversion Rate',
          key: 'Conversion Rate',
          method: attribution_method,
          hasBorder: true
        })
      );
    }

    if (attribution_method_compare) {
      eventColumns.push(
        getEventColumnConfig({
          title: 'Conversion',
          key: 'conversion_compare',
          method: attribution_method_compare,
          hasBorder: conversionBorderCondition
        })
      );
      if (showCPC) {
        eventColumns.push(
          getEventColumnConfig({
            title: 'Cost Per Conversion',
            key: 'cost_compare',
            method: attribution_method_compare,
            hasBorder: costBorderCondition
          })
        );
      }
      if (showCR) {
        eventColumns.push(
          getEventColumnConfig({
            title: 'Conversion Rate',
            key: 'conversion_rate_compare',
            method: attribution_method_compare,
            hasBorder: true
          })
        );
      }
    }
  }

  let linkedEventsColumns = [];
  if (linkedEvents.length) {
    linkedEventsColumns = linkedEvents.map((le) => {
      const linkedEventsChildren = [
        getEventColumnConfig({
          title: 'Conversion',
          key: 'Linked Event - ' + le.label + ' - Users',
          hasBorder: conversionBorderCondition
        })
      ];
      if (showCPC) {
        linkedEventsChildren.push(
          getEventColumnConfig({
            title: 'Cost Per Conversion',
            key: 'Linked Event - ' + le.label + ' - CPC',
            hasBorder: costBorderCondition
          })
        );
      }
      if (showCR) {
        linkedEventsChildren.push(
          getEventColumnConfig({
            title: 'Conversion Rate',
            key: 'Linked Event - ' + le.label + ' - Conversion Rate',
            hasBorder: true
          })
        );
      }
      return {
        title: eventNames[le.label] || le.label,
        className: 'bg-white tableParentHeader ',
        children: linkedEventsChildren
      };
    });
  }

  let tableColumns = [...dimensionColumns, ...metricsColumns];

  if (queryOptions.group_analysis && queryOptions.group_analysis !== 'users') {
    tableColumns = [...tableColumns, ...attrQueryEvents];
  } else {
    tableColumns = [
      ...tableColumns,
      {
        title: eventNames[event] || event,
        className: 'bg-white tableParentHeader ',
        children: eventColumns
      },
      ...linkedEventsColumns
    ];
  }

  return tableColumns;
};

export const calcChangePerc = (val1, val2) => {
  return formatCount(((val1 - val2) / val2) * 100, 1);
};

export const getEquivalentIndicesMapper = (data, comparison_data) => {
  const { headers, rows } = data;
  const firstMetricIndex = headers.indexOf(FIRST_METRIC_IN_ATTR_RESPOSE);
  const dataStrings = rows.map((row) => {
    return row.slice(0, firstMetricIndex).join(ARR_JOINER);
  });
  const compareDataStrings = comparison_data.rows.map((row) => {
    return row.slice(0, firstMetricIndex).join(ARR_JOINER);
  });
  const equivalentIndicesMapper = {};
  dataStrings.forEach((string, index) => {
    const compareIndex = compareDataStrings.indexOf(string);
    equivalentIndicesMapper[index] = compareIndex;
  });
  return equivalentIndicesMapper;
};

const getHeaderIndexForMetric = (headers, metric) => {
  const result = metric.header
    .split(' OR ')
    .map((ph) => {
      return headers.indexOf(ph);
    })
    .filter((d) => d > -1);
  if (result.length) {
    return result[0];
  }
  return -1;
};

const applyAdvancedFilters = (
  data,
  filters,
  index = 0,
  filteredResults = []
) => {
  if (filters == null || filters.categories == null) {
    return data;
  }

  if (index === filters.categories.length) {
    return filteredResults;
  }

  const currentFilter = filters.categories[index];

  if (index === 0) {
    if (filters.categoryCombinationOperator === 'AND') {
      filteredResults = data;
    } else {
      const filterWithValues = filters.categories.find(
        (category) => category.values.length > 0
      );
      if (!filterWithValues) {
        filteredResults = data;
      }
    }
  }

  if (currentFilter.field === null || currentFilter.values.length === 0) {
    return applyAdvancedFilters(data, filters, index + 1, filteredResults);
  }

  const dataToBeFiltered =
    filters.categoryCombinationOperator === 'AND' ? filteredResults : data;

  const currentFilterResults = dataToBeFiltered.filter((d) => {
    const key = currentFilter.field;
    const fieldValue = d[key];
    if (currentFilter.equalityOperator === 'not-equal') {
      return currentFilter.values.indexOf(fieldValue) === -1;
    }
    if (currentFilter.equalityOperator === 'does-not-contain') {
      return !currentFilter.values.includes(fieldValue);
    }
    if (currentFilter.equalityOperator === 'contains') {
      return currentFilter.values.includes(fieldValue);
    }
    return currentFilter.values.indexOf(fieldValue) > -1;
  });

  if (filters.categoryCombinationOperator === 'OR') {
    filteredResults = uniqBy(
      [...filteredResults, ...currentFilterResults],
      (e) => {
        return e.index;
      }
    );
  }

  return applyAdvancedFilters(
    data,
    filters,
    index + 1,
    filters.categoryCombinationOperator === 'AND'
      ? currentFilterResults
      : filteredResults
  );
};

export const getTableData = (
  data,
  event,
  searchText,
  currentSorter,
  attribution_method_compare,
  touchpoint,
  linkedEvents,
  metrics,
  attr_dimensions,
  content_groups,
  comparison_data,
  queryOptions,
  attrQueries,
  appliedFilters
) => {
  const { headers } = data;
  const costIdx = headers.indexOf('Cost Per Conversion');
  const userIdx = headers.indexOf(`${event} - Users`);
  const conversionRateIdx = headers.indexOf('UserConversionRate(%)');
  const compareUsersIdx = headers.indexOf('Compare - Users');
  const compareCostIdx = headers.indexOf('Compare Cost Per Conversion');
  const compareConvRateIdx = headers.indexOf('Compare UserConversionRate(%)');

  const listDimensions =
    touchpoint === 'LandingPage'
      ? content_groups.slice()
      : attr_dimensions.slice();
  const enabledDimensions = listDimensions.filter(
    (d) => d.touchPoint === touchpoint && d.enabled
  );
  const equivalentIndicesMapper = comparison_data
    ? getEquivalentIndicesMapper(data, comparison_data)
    : {};
  const result = data.rows
    .map((row, index) => {
      const metricsData = {};
      const enabledMetrics = metrics.filter((metric) => !metric.isEventMetric);
      const equivalent_compare_row =
        comparison_data && equivalentIndicesMapper[index] > -1
          ? comparison_data.rows[equivalentIndicesMapper[index]]
          : null;
      enabledMetrics.forEach((metric) => {
        const metricIndex = getHeaderIndexForMetric(headers, metric);
        if (comparison_data) {
          metricsData[metric.title] = {
            value: row[metricIndex],
            compare_value: equivalent_compare_row
              ? equivalent_compare_row[metricIndex]
              : 0
          };
        } else {
          metricsData[metric.title] = row[metricIndex];
        }
      });

      const dimensionsData = {};
      if (enabledDimensions.length) {
        enabledDimensions.forEach((dimension) => {
          const index = headers.indexOf(dimension.responseHeader);
          dimensionsData[dimension.title] =
            index > -1
              ? DISPLAY_PROP[row[index]]
                ? DISPLAY_PROP[row[index]]
                : row[index]
              : '';
        });
      } else {
        const touchpointIdx = headers.indexOf(touchpoint);
        dimensionsData[touchpoint] = DISPLAY_PROP[row[touchpointIdx]]
          ? DISPLAY_PROP[row[touchpointIdx]]
          : row[touchpointIdx];
      }

      let resultantRow = {
        index,
        category: Object.values(dimensionsData).join(', '),
        ...dimensionsData,
        ...metricsData
      };

      if (
        queryOptions.group_analysis &&
        queryOptions.group_analysis !== 'users' &&
        attrQueries.length &&
        headers.length
      ) {
        attrQueries.forEach((q) => {
          const lbl = q.label;
          headers.forEach((head, i) => {
            if (head.startsWith(`${lbl} - `)) {
              resultantRow[head] = !comparison_data
                ? formatCount(row[i], 1)
                : {
                    value: formatCount(row[i]),
                    compare_value: equivalent_compare_row
                      ? formatCount(equivalent_compare_row[i], 1)
                      : 0
                  };
            }
          });
        });
      } else {
        resultantRow = {
          index,
          category: Object.values(dimensionsData).join(', '),
          ...dimensionsData,
          ...metricsData,
          Conversion: !comparison_data
            ? formatCount(row[userIdx], 1)
            : {
                value: formatCount(row[userIdx], 1),
                compare_value: equivalent_compare_row
                  ? equivalent_compare_row[userIdx]
                  : 0
              },
          'Cost Per Conversion': !comparison_data
            ? formatCount(row[costIdx], 1)
            : {
                value: formatCount(row[costIdx], 1),
                compare_value: equivalent_compare_row
                  ? formatCount(equivalent_compare_row[costIdx], 1)
                  : 0
              },
          'Conversion Rate': !comparison_data
            ? formatCount(row[conversionRateIdx], 1)
            : {
                value: formatCount(row[conversionRateIdx], 1),
                compare_value: equivalent_compare_row
                  ? formatCount(equivalent_compare_row[conversionRateIdx], 1)
                  : 0
              }
        };
      }
      if (linkedEvents.length) {
        linkedEvents.forEach((le) => {
          const eventUsersIdx = headers.indexOf(`${le.label} - Users`);
          const eventCPCIdx = headers.indexOf(`${le.label} - CPC`);
          const eventConvRateIdx = headers.indexOf(
            `${le.label} - UserConversionRate(%)`
          );
          resultantRow[`Linked Event - ${le.label} - Users`] = !comparison_data
            ? formatCount(row[eventUsersIdx], 1)
            : {
                value: formatCount(row[eventUsersIdx], 1),
                compare_value: equivalent_compare_row
                  ? formatCount(equivalent_compare_row[eventUsersIdx], 1)
                  : 0
              };
          resultantRow[`Linked Event - ${le.label} - CPC`] = !comparison_data
            ? formatCount(row[eventCPCIdx], 1)
            : {
                value: formatCount(row[eventCPCIdx], 1),
                compare_value: equivalent_compare_row
                  ? formatCount(equivalent_compare_row[eventCPCIdx], 1)
                  : 0
              };
          resultantRow[`Linked Event - ${le.label} - Conversion Rate`] =
            !comparison_data
              ? formatCount(row[eventConvRateIdx], 1)
              : {
                  value: formatCount(row[eventConvRateIdx], 1),
                  compare_value: equivalent_compare_row
                    ? formatCount(equivalent_compare_row[eventConvRateIdx], 1)
                    : 0
                };
        });
      }
      if (attribution_method_compare) {
        resultantRow.conversion_compare = row[compareUsersIdx];
        resultantRow.cost_compare = formatCount(row[compareCostIdx], 1);
        resultantRow.conversion_rate_compare = formatCount(
          row[compareConvRateIdx],
          1
        );
      }
      return resultantRow;
    })
    .filter((row) => {
      if (enabledDimensions.length) {
        const filteredRows = enabledDimensions.filter((dimension) =>
          row[dimension.title]?.toLowerCase().includes(searchText.toLowerCase())
        );
        return filteredRows.length > 0;
      } else {
        return row[touchpoint].toLowerCase().includes(searchText.toLowerCase());
      }
    });
  const filteredResults = applyAdvancedFilters(result, appliedFilters);
  return SortResults(filteredResults, currentSorter);
};

export const getScatterPlotChartData = (
  selectedTouchPoint,
  attr_dimensions,
  content_groups,
  data,
  visibleIndices,
  xAxisMetric,
  yAxisMetric,
  isComparisonApplied
) => {
  const listDimensions =
    selectedTouchPoint === 'LandingPage'
      ? content_groups.slice()
      : attr_dimensions.slice();
  const enabledDimensions = listDimensions.filter(
    (d) => d.touchPoint === selectedTouchPoint && d.enabled
  );
  const visibleData = data.filter((d) => visibleIndices.indexOf(d.index) > -1);
  const categories = [];
  const comparisonPlotData = [];
  const plotData = visibleData.map((d) => {
    const category = [];
    if (enabledDimensions.length) {
      for (const dimension of enabledDimensions) {
        category.push(d[dimension.title]);
      }
    } else {
      category.push(d[selectedTouchPoint]);
    }

    categories.push(category.join(', '));
    if (isComparisonApplied) {
      comparisonPlotData.push([
        Number(d[xAxisMetric].compare_value),
        Number(d[yAxisMetric].compare_value)
      ]);
      return [Number(d[xAxisMetric].value), Number(d[yAxisMetric].value)];
    }
    return [Number(d[xAxisMetric]), Number(d[yAxisMetric])];
  });

  const finalResult = {
    series: [
      {
        color: '#4D7DB4',
        data: plotData
      }
    ],
    categories
  };

  if (isComparisonApplied) {
    finalResult.series.push({
      color: '#d4787d',
      data: comparisonPlotData
    });
  }

  return finalResult;
};

export const getAxisMetricOptions = (
  selectedTouchPoint,
  linkedEvents,
  attribution_method,
  attribution_method_compare,
  eventNames
) => {
  const result = getResultantMetrics(selectedTouchPoint, ATTRIBUTION_METRICS)
    .filter((metric) => !metric.isEventMetric)
    .map((metric) => {
      return {
        title: metric.title,
        value: metric.title
      };
    });

  result.push({
    title: attribution_method_compare
      ? `Conversion - ${
          ATTRIBUTION_METHODOLOGY.find((m) => m.value === attribution_method)
            .text
        }`
      : 'Conversion',
    value: 'Conversion'
  });

  result.push({
    title: attribution_method_compare
      ? `Cost Per Conversion - ${
          ATTRIBUTION_METHODOLOGY.find((m) => m.value === attribution_method)
            .text
        }`
      : 'Cost Per Conversion',
    value: 'Cost Per Conversion'
  });

  result.push({
    title: attribution_method_compare
      ? `Conversion Rate - ${
          ATTRIBUTION_METHODOLOGY.find((m) => m.value === attribution_method)
            .text
        }`
      : 'Conversion Rate',
    value: 'Conversion Rate'
  });

  if (attribution_method_compare) {
    result.push({
      title: `Conversion - ${
        ATTRIBUTION_METHODOLOGY.find(
          (m) => m.value === attribution_method_compare
        ).text
      }`,
      value: 'conversion_compare'
    });

    result.push({
      title: `Cost Per Conversion - ${
        ATTRIBUTION_METHODOLOGY.find(
          (m) => m.value === attribution_method_compare
        ).text
      }`,
      value: 'cost_compare'
    });

    result.push({
      title: `Conversion Rate - ${
        ATTRIBUTION_METHODOLOGY.find(
          (m) => m.value === attribution_method_compare
        ).text
      }`,
      value: 'conversion_rate_compare'
    });
  }

  linkedEvents.map((le) => {
    result.push({
      title: `Conversion - ${eventNames[le.label] || le.label}`,
      value: `Linked Event - ${le.label} - Users`
    });

    result.push({
      title: `Cost Per Conversion - ${eventNames[le.label] || le.label}`,
      value: `Linked Event - ${le.label} - CPC`
    });

    result.push({
      title: `Conversion Rate - ${eventNames[le.label] || le.label}`,
      value: `Linked Event - ${le.label} - Conversion Rate`
    });
  });

  return result;
};

export const listAttributionDimensions = (
  touchpoint,
  attr_dimensions,
  content_groups
) => {
  return touchpoint === 'LandingPage'
    ? content_groups.slice()
    : attr_dimensions.slice();
};

export const getResultantMetrics = (touchpoint, attribution_metrics) => {
  return touchpoint === 'LandingPage'
    ? attribution_metrics.filter(
        (metrics) =>
          metrics.header.includes('Sessions') ||
          metrics.header.includes('Users') ||
          metrics.header.includes('Average Session Time') ||
          metrics.header.includes('PageViews') ||
          metrics.header.includes('ALL CR')
      )
    : attribution_metrics;
};
