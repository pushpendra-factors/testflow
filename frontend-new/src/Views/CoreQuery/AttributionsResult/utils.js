import React from 'react';
import {
  SortData,
  getTitleWithSorter,
  formatCount,
} from '../../../utils/dataFormatter';
import {
  ATTRIBUTION_METHODOLOGY,
  MARKETING_TOUCHPOINTS,
} from '../../../utils/constants';
import styles from './index.module.scss';

import {
  SVG,
  Number as NumFormat,
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

export const formatData = (data, event, visibleIndices, touchpoint) => {
  const { headers } = data;
  const touchpointIdx = headers.indexOf(touchpoint);
  const costIdx = headers.indexOf('Cost Per Conversion');
  const userIdx = headers.indexOf(`${event} - Users`);
  const rows = data.rows.filter(
    (_, index) => visibleIndices.indexOf(index) > -1
  );
  const result = rows.map((row) => {
    return [row[touchpointIdx], row[costIdx], row[userIdx]];
  });
  return SortData(result, 2, 'descend');
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

const renderComparCell = (obj) => {
  let changeMetric = null;
  if (obj.change) {
    if (obj.change > 0 || obj.change < 0) {
      const change = Math.abs(obj.change);
      changeMetric = (
        <div className={`${styles.cmprCell__change}`}>
          <SVG
            name={obj.change > 0 ? `arrowLift` : `arrowDown`}
            size={16}
          ></SVG>
          <span>
            {obj.change === 'Infinity' ? <>&#8734;</> : <>{change} &#37;</>}
          </span>
        </div>
      );
    }
  }

  return (
    <div className={styles.cmprCell}>
      <span className={styles.cmprCell__first}>
        <NumFormat number={obj.first} />
      </span>
      <span className={styles.cmprCell__second}>
        <NumFormat number={obj.second} />
      </span>
      {changeMetric}
    </div>
  );
};

export const getCompareTableColumns = (
  currentSorter,
  handleSorting,
  attribution_method,
  attribution_method_compare,
  touchpoint,
  linkedEvents,
  event,
  eventNames,
  metrics,
  OptionsPopover
) => {
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
        width: 150,
        className: 'align-bottom',
        render: renderComparCell,
      };
    });
  const result = [
    {
      title: touchpoint,
      dataIndex: touchpoint,
      fixed: 'left',
      width: 150,
      className: 'align-bottom',
    },
    ...metricsColumns,
    {
      title: eventNames[event] || event,
      className: 'tableParentHeader',
      children: [
        {
          title: getTitleWithSorter(
            <div className='flex flex-col items-start justify-ceneter'>
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
          width: 150,
          render: renderComparCell,
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
          width: 150,
          render: renderComparCell,
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
      render: renderComparCell,
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
        currentSorter,
        handleSorting
      ),
      dataIndex: 'cost_compare',
      width: 150,
      render: renderComparCell,
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
            render: renderComparCell,
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
            render: renderComparCell,
          },
        ],
      };
    });
  }
  let extraCols = [];
  if (touchpoint === MARKETING_TOUCHPOINTS.ADGROUP) {
    extraCols = [
      {
        title: MARKETING_TOUCHPOINTS.CAMPAIGN,
        dataIndex: MARKETING_TOUCHPOINTS.CAMPAIGN,
        fixed: 'left',
        width: 150,
        className: 'align-bottom',
      },
    ];
  }
  if (touchpoint === MARKETING_TOUCHPOINTS.KEYWORD) {
    extraCols = [
      {
        title: MARKETING_TOUCHPOINTS.CAMPAIGN,
        dataIndex: MARKETING_TOUCHPOINTS.CAMPAIGN,
        fixed: 'left',
        width: 150,
        className: 'align-bottom',
      },
      {
        title: MARKETING_TOUCHPOINTS.ADGROUP,
        dataIndex: MARKETING_TOUCHPOINTS.ADGROUP,
        fixed: 'left',
        width: 150,
        className: 'align-bottom',
      },
      {
        title: MARKETING_TOUCHPOINTS.MATCHTYPE,
        dataIndex: MARKETING_TOUCHPOINTS.MATCHTYPE,
        fixed: 'left',
        width: 150,
        className: 'align-bottom',
      },
    ];
  }
  return [...extraCols, ...result, ...linkedEventsColumns];
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
  OptionsPopover
) => {
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
        width: 150,
        className: 'align-bottom',
        render: (d) => {
          return <NumFormat number={d} />;
        },
      };
    });
  const result = [
    {
      title: touchpoint,
      dataIndex: touchpoint,
      fixed: 'left',
      width: 150,
      className: 'align-bottom',
    },
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
          width: 150,
          render: (d) => {
            return <NumFormat number={d} />;
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
          width: 150,
          render: (d) => {
            return <NumFormat number={d} />;
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
  let extraCols = [];
  if (touchpoint === MARKETING_TOUCHPOINTS.ADGROUP) {
    extraCols = [
      {
        title: MARKETING_TOUCHPOINTS.CAMPAIGN,
        dataIndex: MARKETING_TOUCHPOINTS.CAMPAIGN,
        fixed: 'left',
        width: 150,
        className: 'align-bottom',
      },
    ];
  }
  if (touchpoint === MARKETING_TOUCHPOINTS.KEYWORD) {
    extraCols = [
      {
        title: MARKETING_TOUCHPOINTS.CAMPAIGN,
        dataIndex: MARKETING_TOUCHPOINTS.CAMPAIGN,
        fixed: 'left',
        width: 150,
        className: 'align-bottom',
      },
      {
        title: MARKETING_TOUCHPOINTS.ADGROUP,
        dataIndex: MARKETING_TOUCHPOINTS.ADGROUP,
        fixed: 'left',
        width: 150,
        className: 'align-bottom',
      },
      {
        title: MARKETING_TOUCHPOINTS.MATCHTYPE,
        dataIndex: MARKETING_TOUCHPOINTS.MATCHTYPE,
        fixed: 'left',
        width: 150,
        className: 'align-bottom',
      },
    ];
  }
  return [...extraCols, ...result, ...linkedEventsColumns];
};

const constrComparisionCellData = (row, row2, index) => {
  return {
    first: formatCount(row[index], 1),
    second: row2 ? formatCount(row2[index], 1) : NaN,
    change: row2 ? calcChangePerc(row[index], row2[index]) : NaN,
  };
};

export const getCompareTableData = (
  data,
  data2,
  event,
  searchText,
  currentSorter,
  attribution_method_compare,
  touchpoint,
  linkedEvents,
  metrics
) => {
  const { headers } = data;
  const touchpointIdx = headers.indexOf(touchpoint);
  const costIdx = headers.indexOf('Cost Per Conversion');
  const userIdx = headers.indexOf(`${event} - Users`);
  const compareUsersIdx = headers.indexOf(`Compare - Users`);
  const compareCostIdx = headers.indexOf(`Compare Cost Per Conversion`);
  const data2Rows = data2.rows;
  const result = data.rows
    .map((row, index) => {
      const row2 = data2Rows.filter(
        (r) => r[touchpointIdx] === row[touchpointIdx]
      )[0];
      const metricsData = {};
      const enabledMetrics = metrics.filter((metric) => metric.enabled);
      enabledMetrics.forEach((metric) => {
        const index = headers.indexOf(metric.header);
        metricsData[metric.title] = constrComparisionCellData(row, row2, index);
      });
      let resultantRow = {
        index,
        [touchpoint]: row[touchpointIdx],
        ...metricsData,
        conversion: constrComparisionCellData(row, row2, userIdx),
        cost: constrComparisionCellData(row, row2, costIdx),
      };
      if (touchpoint === MARKETING_TOUCHPOINTS.ADGROUP) {
        const campaignIdx = headers.indexOf(MARKETING_TOUCHPOINTS.CAMPAIGN);
        resultantRow = {
          [MARKETING_TOUCHPOINTS.CAMPAIGN]: row[campaignIdx],
          ...resultantRow,
        };
      }
      if (touchpoint === MARKETING_TOUCHPOINTS.KEYWORD) {
        const campaignIdx = headers.indexOf(MARKETING_TOUCHPOINTS.CAMPAIGN);
        const adGroupIdx = headers.indexOf(MARKETING_TOUCHPOINTS.ADGROUP);
        const matchTypeIdx = headers.indexOf(MARKETING_TOUCHPOINTS.MATCHTYPE);
        resultantRow = {
          [MARKETING_TOUCHPOINTS.CAMPAIGN]: row[campaignIdx],
          [MARKETING_TOUCHPOINTS.ADGROUP]: row[adGroupIdx],
          [MARKETING_TOUCHPOINTS.MATCHTYPE]: row[matchTypeIdx],
          ...resultantRow,
        };
      }
      if (linkedEvents.length) {
        linkedEvents.forEach((le) => {
          const eventUsersIdx = headers.indexOf(`${le.label} - Users`);
          const eventCPCIdx = headers.indexOf(`${le.label} - CPC`);
          resultantRow[`${le.label} - Users`] = constrComparisionCellData(
            row,
            row2,
            eventUsersIdx
          );
          resultantRow[`${le.label} - CPC`] = constrComparisionCellData(
            row,
            row2,
            eventCPCIdx
          );
        });
      }
      if (attribution_method_compare) {
        resultantRow['conversion_compare'] = constrComparisionCellData(
          row,
          row2,
          [compareUsersIdx]
        );
        resultantRow['cost_compare'] = constrComparisionCellData(row, row2, [
          compareCostIdx,
        ]);
      }
      return resultantRow;
    })
    .filter(
      (row) =>
        row[touchpoint].toLowerCase().indexOf(searchText.toLowerCase()) > -1
    );

  if (!currentSorter.key) {
    result.sort((a, b) => {
      return parseFloat(a['conversion'].first) <=
        parseFloat(b['conversion'].first)
        ? 1
        : -1;
    });
  } else {
    result.sort((a, b) => {
      if (currentSorter.order === 'ascend') {
        return parseFloat(a[currentSorter.key].first) >=
          parseFloat(b[currentSorter.key].first)
          ? 1
          : -1;
      }
      if (currentSorter.order === 'descend') {
        return parseFloat(a[currentSorter.key].first) <=
          parseFloat(b[currentSorter.key].first)
          ? 1
          : -1;
      }
      return 0;
    });
  }
  return result;
};

export const calcChangePerc = (val1, val2) => {
  return formatCount(((val1 - val2) / val2) * 100, 1);
};

export const getTableData = (
  data,
  event,
  searchText,
  currentSorter,
  attribution_method_compare,
  touchpoint,
  linkedEvents,
  metrics
) => {
  const { headers } = data;
  const touchpointIdx = headers.indexOf(touchpoint);
  const costIdx = headers.indexOf('Cost Per Conversion');
  const userIdx = headers.indexOf(`${event} - Users`);
  const compareUsersIdx = headers.indexOf(`Compare - Users`);
  const compareCostIdx = headers.indexOf(`Compare Cost Per Conversion`);
  const result = data.rows
    .map((row, index) => {
      const metricsData = {};
      const enabledMetrics = metrics.filter((metric) => metric.enabled);
      enabledMetrics.forEach((metric) => {
        const index = headers.indexOf(metric.header);
        metricsData[metric.title] = row[index];
      });
      let resultantRow = {
        index,
        [touchpoint]: row[touchpointIdx],
        ...metricsData,
        conversion: formatCount(row[userIdx], 1),
        cost: formatCount(row[costIdx], 1),
      };
      if (touchpoint === MARKETING_TOUCHPOINTS.ADGROUP) {
        const campaignIdx = headers.indexOf(MARKETING_TOUCHPOINTS.CAMPAIGN);
        resultantRow = {
          [MARKETING_TOUCHPOINTS.CAMPAIGN]: row[campaignIdx],
          ...resultantRow,
        };
      }
      if (touchpoint === MARKETING_TOUCHPOINTS.KEYWORD) {
        const campaignIdx = headers.indexOf(MARKETING_TOUCHPOINTS.CAMPAIGN);
        const adGroupIdx = headers.indexOf(MARKETING_TOUCHPOINTS.ADGROUP);
        const matchTypeIdx = headers.indexOf(MARKETING_TOUCHPOINTS.MATCHTYPE);
        resultantRow = {
          [MARKETING_TOUCHPOINTS.CAMPAIGN]: row[campaignIdx],
          [MARKETING_TOUCHPOINTS.ADGROUP]: row[adGroupIdx],
          [MARKETING_TOUCHPOINTS.MATCHTYPE]: row[matchTypeIdx],
          ...resultantRow,
        };
      }
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
    .filter(
      (row) =>
        row[touchpoint].toLowerCase().indexOf(searchText.toLowerCase()) > -1
    );

  if (!currentSorter.key) {
    return SortData(result, 'conversion', 'descend');
  }
  return SortData(result, currentSorter.key, currentSorter.order);
};
