import React from 'react';
import moment from 'moment';
import {
  SortData,
  getTitleWithSorter,
  formatCount,
  SortDataByObject,
} from '../../../utils/dataFormatter';
import {
  ATTRIBUTION_METHODOLOGY,
  FIRST_METRIC_IN_ATTR_RESPOSE,
  ARR_JOINER,
} from '../../../utils/constants';
import styles from './index.module.scss';

import {
  SVG,
  Number as NumFormat,
  Text,
} from '../../../components/factorsComponents';
import { Popover } from 'antd';

export const getDifferentCampaingns = (data) => {
  const { headers } = data.result;
  const campaignIdx = headers.indexOf('Campaign');
  let differentCampaigns = new Set();
  data.result.rows.forEach((row) => {
    differentCampaigns.add(row[campaignIdx]);
  });
  return Array.from(differentCampaigns);
};

export const formatData = (
  data,
  touchPoint,
  event,
  attr_dimensions,
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
      series: [],
    };
  }
  const { headers, rows } = data;
  const enabledDimensions = attr_dimensions.filter(
    (d) => d.touchPoint === touchPoint && d.enabled
  );
  const firstDimensionIdx = headers.findIndex(
    (h) => h === enabledDimensions[0].responseHeader
  );
  const lastDimensionIdx = headers.findIndex(
    (h) => h === enabledDimensions[enabledDimensions.length - 1].responseHeader
  );
  const categories = rows.map((row) => {
    return row.slice(firstDimensionIdx, lastDimensionIdx + 1).join(', ');
  });
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
      color: '#4d7db4',
    },
    {
      type: 'line',
      yAxis: 1,
      data: rows.map((row) => row[costIdx]),
      color: '#d4787d',
      marker: {
        symbol: 'circle',
      },
    },
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
      color: '#4d7db4',
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
        symbol: 'circle',
      },
      dashStyle: 'dash',
    });
    let temp = series[1];
    series[1] = series[2];
    series[2] = temp;
  }
  return {
    categories,
    series,
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
  const str = currMetricsValue ? `Cost Per Conversion` : `${event} - Users`;
  const compareStr = currMetricsValue
    ? `Compare Cost Per Conversion`
    : `Compare - Users`;
  const userIdx = headers.indexOf(str);
  const compareUsersIdx = headers.indexOf(compareStr);
  let rows = data.rows.filter((_, index) => visibleIndices.indexOf(index) > -1);
  rows = SortData(rows, userIdx, 'descend');
  const chartData = rows.map((row) => {
    return {
      name: row[0],
      [attribution_method]: row[userIdx],
      [attribution_method_compare]: row[compareUsersIdx],
    };
  });
  return chartData;
};

const firstColumn = (d, durationObj, cmprDuration) => {
  if (cmprDuration) {
    return (
      <div className='flex items-center'>
        <Text
          type='title'
          weight='normal'
          color='grey-8'
          extraClass='text-sm mb-0 py-2 px-4 w-1/2'
        >
          {d}
        </Text>
        <div
          style={{ borderLeft: '1px solid #E7E9ED' }}
          className='flex py-2 flex-col px-4 w-1/2'
        >
          <Text
            type='title'
            weight='normal'
            color='grey-8'
            extraClass='text-sm mb-0'
          >
            {`${moment(durationObj.from).format('MMM DD')} - ${moment(
              durationObj.to
            ).format('MMM DD')}`}
          </Text>
          <Text
            type='title'
            weight='normal'
            color='grey'
            extraClass='text-xs mb-0'
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
  let changePercent = calcChangePerc(d.value, d.compare_value);
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
        <SVG color='#5ACA89' name={`arrowLift`} size={16}></SVG>
        <span>&#8734; %</span>
      </>
    );
  } else {
    compareText = (
      <>
        <SVG
          color={changePercent > 0 ? '#5ACA89' : '#FF0000'}
          name={changePercent > 0 ? `arrowLift` : `arrowDown`}
          size={16}
        ></SVG>
        <NumFormat number={Math.abs(changePercent)} />%
      </>
    );
  }
  return (
    <div className='flex flex-col justify-center items-start'>
      <div className='flex items-center'>
        <div className='mr-2'>
          <NumFormat number={d.value} />
        </div>
        <div className={styles.changePercent}>{compareText}</div>
      </div>
      <div className={styles.compareNumber}>
        <NumFormat number={d.compare_value} />
      </div>
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
  OptionsPopover,
  attr_dimensions,
  durationObj,
  comparison_data,
  cmprDuration
) => {
  const enabledDimensions = attr_dimensions.filter(
    (d) => d.touchPoint === touchpoint && d.enabled
  );
  let dimensionColumns;
  if (enabledDimensions.length) {
    dimensionColumns = enabledDimensions.map((d, index) => {
      if (!index) {
        return {
          title: d.title,
          dataIndex: d.title,
          fixed: 'left',
          width: comparison_data ? 300 : 200,
          className: `align-bottom ${
            comparison_data ? styles.touchPointCol : ''
          }`,
          render: (d) =>
            firstColumn(d, durationObj, comparison_data ? cmprDuration : null),
        };
      } else {
        return {
          title: d.title,
          dataIndex: d.title,
          width: 200,
          className: 'align-bottom',
        };
      }
    });
  } else {
    dimensionColumns = [
      {
        title: touchpoint,
        dataIndex: touchpoint,
        fixed: 'left',
        width: 200,
        className: 'align-bottom',
      },
    ];
  }
  const metricsColumns = metrics
    .filter((metric) => metric.enabled)
    .map((metric, index, arr) => {
      return {
        title:
          index !== arr.length - 1 ? (
            getTitleWithSorter(
              metric.title,
              metric.title,
              currentSorter,
              handleSorting
            )
          ) : (
            <div className='flex flex-col'>
              <div className='mb-6 flex justify-end '>
                <Popover
                  placement='bottomLeft'
                  trigger='click'
                  content={OptionsPopover}
                >
                  <span
                    className={`text-2xl font-normal inline-flex items-center justify-center cursor-pointer absolute bg-white top-4 -right-4 z-10 shadow ${styles.metricOptionsToggler}`}
                  >
                    +
                  </span>
                </Popover>
              </div>
              {getTitleWithSorter(
                metric.title,
                metric.title,
                currentSorter,
                handleSorting
              )}
            </div>
          ),
        dataIndex: metric.title,
        width: 200,
        className: 'align-bottom',
        render: (d) => {
          return renderMetric(d, comparison_data);
        },
      };
    });
  const result = [
    ...dimensionColumns,
    ...metricsColumns,
    {
      title: eventNames[event] || event,
      className: 'tableParentHeader',
      children: [
        {
          title: getTitleWithSorter(
            <div className='flex flex-col items-start justify-center'>
              <div>Conversion</div>
              <div style={{ fontSize: '10px', color: '#8692A3' }}>
                {
                  ATTRIBUTION_METHODOLOGY.find(
                    (m) => m.value === attribution_method
                  ).text
                }
              </div>
            </div>,
            'conversion',
            currentSorter,
            handleSorting
          ),
          dataIndex: 'conversion',
          width: 200,
          render: (d) => {
            return renderMetric(d, comparison_data);
          },
        },
        {
          title: getTitleWithSorter(
            <div className='flex flex-col items-start justify-ceneter'>
              <div>Cost per Conversion</div>
              <div style={{ fontSize: '10px', color: '#8692A3' }}>
                {
                  ATTRIBUTION_METHODOLOGY.find(
                    (m) => m.value === attribution_method
                  ).text
                }
              </div>
            </div>,
            'cost',
            currentSorter,
            handleSorting
          ),
          dataIndex: 'cost',
          width: 200,
          render: (d) => {
            return renderMetric(d, comparison_data);
          },
        },
      ],
    },
  ];
  if (attribution_method_compare) {
    result[result.length - 1].children.push({
      title: getTitleWithSorter(
        <div className='flex flex-col items-start justify-ceneter'>
          <div>Conversion</div>
          <div style={{ fontSize: '10px', color: '#8692A3' }}>
            {
              ATTRIBUTION_METHODOLOGY.find(
                (m) => m.value === attribution_method_compare
              ).text
            }
          </div>
        </div>,
        'conversion_compare',
        currentSorter,
        handleSorting
      ),
      dataIndex: 'conversion_compare',
      width: 150,
      render: (d) => {
        return <NumFormat number={d} />;
      },
    });
    result[result.length - 1].children.push({
      title: getTitleWithSorter(
        <div className='flex flex-col items-start justify-ceneter'>
          <div>Cost per Conversion</div>
          <div style={{ fontSize: '10px', color: '#8692A3' }}>
            {
              ATTRIBUTION_METHODOLOGY.find(
                (m) => m.value === attribution_method_compare
              ).text
            }
          </div>
        </div>,
        'cost_compare',
        currentSorter,
        handleSorting
      ),
      dataIndex: 'cost_compare',
      width: 150,
      render: (d) => {
        return <NumFormat number={d} />;
      },
    });
  }
  let linkedEventsColumns = [];
  if (linkedEvents.length) {
    linkedEventsColumns = linkedEvents.map((le) => {
      return {
        title: eventNames[le.label] || le.label,
        className: 'tableParentHeader',
        children: [
          {
            title: getTitleWithSorter(
              <div className='flex flex-col items-start justify-ceneter'>
                <div>Users</div>
              </div>,
              le.label + ' - Users',
              currentSorter,
              handleSorting
            ),
            dataIndex: le.label + ' - Users',
            width: 150,
            render: (d) => {
              return <NumFormat number={d} />;
            },
          },
          {
            title: getTitleWithSorter(
              <div className='flex flex-col items-start justify-ceneter'>
                <div>Cost per Conversion</div>
              </div>,
              le.label + ' - CPC',
              currentSorter,
              handleSorting
            ),
            dataIndex: le.label + ' - CPC',
            width: 150,
            render: (d) => {
              return <NumFormat number={d} />;
            },
          },
        ],
      };
    });
  }
  return [...result, ...linkedEventsColumns];
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
  comparison_data
) => {
  const { headers } = data;
  const costIdx = headers.indexOf('Cost Per Conversion');
  const userIdx = headers.indexOf(`${event} - Users`);
  const compareUsersIdx = headers.indexOf(`Compare - Users`);
  const compareCostIdx = headers.indexOf(`Compare Cost Per Conversion`);
  const enabledDimensions = attr_dimensions.filter(
    (d) => d.touchPoint === touchpoint && d.enabled
  );
  const equivalentIndicesMapper = comparison_data
    ? getEquivalentIndicesMapper(data, comparison_data)
    : {};
  const result = data.rows
    .map((row, index) => {
      const metricsData = {};
      const enabledMetrics = metrics.filter((metric) => metric.enabled);
      const equivalent_compare_row =
        comparison_data && equivalentIndicesMapper[index] > -1
          ? comparison_data.rows[equivalentIndicesMapper[index]]
          : null;
      enabledMetrics.forEach((metric) => {
        const metricIndex = headers.indexOf(metric.header);
        if (comparison_data) {
          metricsData[metric.title] = {
            value: row[metricIndex],
            compare_value: equivalent_compare_row
              ? equivalent_compare_row[metricIndex]
              : 0,
          };
        } else {
          metricsData[metric.title] = row[metricIndex];
        }
      });

      const dimensionsData = {};
      if (enabledDimensions.length) {
        enabledDimensions.forEach((dimension) => {
          const index = headers.indexOf(dimension.responseHeader);
          dimensionsData[dimension.title] = row[index];
        });
      } else {
        const touchpointIdx = headers.indexOf(touchpoint);
        dimensionsData[touchpoint] = row[touchpointIdx];
      }

      let resultantRow = {
        index,
        ...dimensionsData,
        ...metricsData,
        conversion: !comparison_data
          ? formatCount(row[userIdx], 1)
          : {
              value: formatCount(row[userIdx], 1),
              compare_value: equivalent_compare_row
                ? equivalent_compare_row[userIdx]
                : 0,
            },
        cost: !comparison_data
          ? formatCount(row[costIdx], 1)
          : {
              value: formatCount(row[costIdx], 1),
              compare_value: equivalent_compare_row
                ? formatCount(equivalent_compare_row[costIdx], 1)
                : 0,
            },
      };
      if (linkedEvents.length) {
        linkedEvents.forEach((le) => {
          const eventUsersIdx = headers.indexOf(`${le.label} - Users`);
          const eventCPCIdx = headers.indexOf(`${le.label} - CPC`);
          resultantRow[`${le.label} - Users`] = formatCount(
            row[eventUsersIdx],
            0
          );
          resultantRow[`${le.label} - CPC`] = formatCount(row[eventCPCIdx], 0);
        });
      }
      if (attribution_method_compare) {
        resultantRow['conversion_compare'] = row[compareUsersIdx];
        resultantRow['cost_compare'] = formatCount(row[compareCostIdx], 0);
      }
      return resultantRow;
    })
    .filter((row) => {
      if (enabledDimensions.length) {
        const filteredRows = enabledDimensions.filter((dimension) =>
          row[dimension.title].toLowerCase().includes(searchText.toLowerCase())
        );
        return filteredRows.length > 0;
      } else {
        return row[touchpoint].toLowerCase().includes(searchText.toLowerCase());
      }
    });

  if (comparison_data) {
    if (!currentSorter.key) {
      return SortDataByObject(result, 'conversion', 'value', 'descend');
    }
    return SortDataByObject(
      result,
      currentSorter.key,
      'value',
      currentSorter.order
    );
  }

  if (!currentSorter.key) {
    return SortData(result, 'conversion', 'descend');
  }
  return SortData(result, currentSorter.key, currentSorter.order);
};
