import React from 'react';
import moment from 'moment';
import _ from 'lodash';
import {
  calculatePercentage,
  formatDuration,
  getClickableTitleSorter,
  SortResults,
  SortData,
  getDurationInSeconds
} from '../../../utils/dataFormatter';
import { CHART_COLOR_1 } from '../../../constants/color.constants';
import {
  Number as NumFormat,
  Text,
  SVG
} from '../../../components/factorsComponents';
import styles from './index.module.scss';
import {
  parseForDateTimeLabel,
  getBreakdownDisplayName
} from '../EventsAnalytics/eventsAnalytics.helpers';
import {
  GROUPED_MAX_ALLOWED_VISIBLE_PROPERTIES,
  DISPLAY_PROP
} from '../../../utils/constants';
import NonClickableTableHeader from '../../../components/NonClickableTableHeader';
import ControlledComponent from 'Components/ControlledComponent';
import { getCompareGroupsByName } from './GroupedChart/groupedChart.helpers';

const windowSize = {
  w: window.outerWidth,
  h: window.outerHeight,
  iw: window.innerWidth,
  ih: window.innerHeight
};

export const getVisibleData = (data, sorter) => {
  const result = SortResults(data, sorter).slice(
    0,
    GROUPED_MAX_ALLOWED_VISIBLE_PROPERTIES
  );
  return result;
};

function NoBreakdownUsersColumn(d, breakdown, isComparisonApplied) {
  if (breakdown.length) {
    if (d.includes('$no_group')) {
      return 'Overall';
    }
    return d;
  }
  if (isComparisonApplied) {
    return (
      <div className='flex items-center'>
        <Text
          type='title'
          weight='normal'
          color='grey-8'
          extraClass='text-sm mb-0 py-2 px-4 w-1/2'
        >
          All
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
            {`${moment(d.durationObj.from).format('MMM DD')} - ${moment(
              d.durationObj.to
            ).format('MMM DD')}`}
          </Text>
          <Text
            type='title'
            weight='normal'
            color='grey'
            extraClass='text-xs mb-0'
          >{`vs ${moment(d.comparison_duration.from).format(
            'MMM DD'
          )} - ${moment(d.comparison_duration.to).format('MMM DD')}`}</Text>
        </div>
      </div>
    );
  }
  return d;
}

export const grnByIndex = (headersSlice, breakdowns) => {
  const grns = [];
  const clonnedBreakdown = [...breakdowns];
  headersSlice.forEach((h) => {
    const brkIndex = clonnedBreakdown.findIndex((x) => h === x.pr);
    grns.push(clonnedBreakdown[brkIndex]?.grn);
    clonnedBreakdown.splice(brkIndex, 1);
  });
  return grns;
};

export const formatData = (response, arrayMapper) => {
  if (
    !response ||
    !response.headers ||
    !Array.isArray(response.headers) ||
    !response.headers.length ||
    !response.rows ||
    !Array.isArray(response.rows) ||
    !response.rows.length ||
    !response.meta ||
    !response.meta.metrics ||
    !Array.isArray(response.meta.metrics) ||
    !response.meta.metrics.length
  ) {
    return { groups: [], events: [] };
  }
  console.log('funnels format data');

  const { rows, headers, meta } = response;
  const breakdowns = [...meta.query.gbp];
  const eventsCondition = [...meta.query?.ec];
  const firstEventIdx = headers.findIndex((header) => header === 'step_0');
  const netConversionIndex = headers.findIndex(
    (header) => header === 'conversion_overall'
  );

  const grns = grnByIndex(headers.slice(0, firstEventIdx), breakdowns);

  const eventsData = arrayMapper.map((am, index) => ({
    index: index + 1,
    name: am.mapper,
    data: {}
  }));

  const result = rows.map((row, index) => {
    const breakdownData = {};
    const breakdownVals = row
      .slice(0, firstEventIdx)
      .map((vl) => (DISPLAY_PROP[vl] ? DISPLAY_PROP[vl] : vl));
    for (let i = 0; i < breakdowns.length; i++) {
      const bkd = breakdowns[i];
      const bkdVal = parseForDateTimeLabel(grns[i], breakdownVals[i]);
      breakdownData[`${bkd.pr} - ${bkd.eni}`] =
        bkdVal === '$no_group' ? 'Overall' : bkdVal;
    }
    const name = Object.values(breakdownData).join(', ');
    const nonConvertedName = row.slice(0, firstEventIdx).join(', ');

    const durationMetric = meta.metrics.find(
      (elem) => elem.title === 'MetaStepTimeInfo'
    );
    const durationGrp = durationMetric.rows.find(
      (elem) => elem.slice(0, firstEventIdx).join(', ') === nonConvertedName
    );

    const timeData = {};
    const groupEventData = {};
    let totalDuration = 0;

    arrayMapper.forEach((am, idx) => {
      const eventIdx = headers.findIndex(
        (headers) => headers === `step_${idx}`
      );

      const percent = calculatePercentage(row[eventIdx], row[firstEventIdx]);

      groupEventData[`${am.displayName}-${idx}-percent`] = percent;
      groupEventData[`${am.displayName}-${idx}-count`] = row[eventIdx];
      groupEventData[am.mapper] = percent;

      eventsData[idx].data[name] = row[eventIdx];

      if (idx < arrayMapper.length - 1) {
        const durationIdx = durationMetric.headers.findIndex(
          (elem) => elem === `step_${idx}_${idx + 1}_time`
        );
        timeData[`time[${idx}-${idx + 1}]`] = durationGrp
          ? formatDuration(durationGrp[durationIdx])
          : 'NA';
        totalDuration += durationGrp ? Number(durationGrp[durationIdx]) : 0;
      }
    });

    const value = row[netConversionIndex];
    const result = {
      index,
      name,
      value: `${value}%`,
      ...breakdownData,
      ...groupEventData,
      Conversion: value, // used for sorting, value will be removed soon
      nonConvertedName
    };
    if (eventsCondition !== 'funnel_any_given_event') {
      result['Conversion Time'] = formatDuration(totalDuration);
      result.concat(timeData);
    }
    return result;
  });

  return {
    groups: result,
    events: eventsData
  };
};

const compareSkeleton = (val1, val2) => (
  <div className='flex flex-col'>
    <Text type='title' weight='normal' color='grey-8' extraClass='text-sm mb-0'>
      {val1}
    </Text>
    <Text type='title' weight='normal' color='grey' extraClass='text-xs mb-0'>
      {val2}
    </Text>
  </div>
);

const RenderTotalConversion = (d, breakdown, isComparisonApplied) => {
  if (!isComparisonApplied) {
    return `${d}%`;
  }
  return compareSkeleton(`${d.conversion}%`, `${d.comparison_conversion}%`);
};

const RenderConversionTime = (d, breakdown, isComparisonApplied) => {
  if (!isComparisonApplied) {
    return d;
  }
  return compareSkeleton(d.overallDuration, d.comparisonOverallDuration);
};

export const getBreakdownTitle = (
  breakdown,
  userPropNames,
  eventPropertiesDisplayNames
) => {
  const charArr = ['1', '2', '3', '4', '5', '6'];
  const displayTitle = getBreakdownDisplayName({
    breakdown,
    userPropNames,
    eventPropertiesDisplayNames,
    multipleEvents: true
  });

  if (!breakdown.eni) {
    return <div className='break-all'>{displayTitle}</div>;
  }
  return (
    <div className='break-all'>
      <span>{displayTitle} of </span>
      <span className='inline-block'>
        <span
          style={{ backgroundColor: '#3E516C' }}
          className='text-white w-4 h-4 flex justify-center items-center rounded-full font-semibold leading-5 text-xs'
        >
          {charArr[breakdown.eni - 1]}
        </span>
      </span>
    </div>
  );
};

export const getTableColumns = (
  queries,
  currentSorter,
  handleSorting,
  arrayMapper,
  isComparisonApplied,
  resultData,
  userPropNames,
  eventPropertiesDisplayNames,
  tableConfig = {}
) => {
  const showOnlyCount =
    !tableConfig.showDuration && !tableConfig.showPercentage;

  const unsortedBreakdown = _.get(resultData, 'meta.query.gbp', []);
  const eventsCondition = _.get(resultData, 'meta.query.ec', '');
  const isBreakdownApplied = unsortedBreakdown.length > 0;
  const breakdown = SortData(unsortedBreakdown, 'eni', 'ascend');

  const getBreakdownColConfig = (e, index) => ({
    title: getClickableTitleSorter(
      getBreakdownTitle(e, userPropNames, eventPropertiesDisplayNames),
      {
        key: `${e.pr} - ${e.eni}`,
        type: e.pty,
        subtype: e.grn
      },
      currentSorter,
      handleSorting,
      'left',
      'end',
      'pb-3'
    ),
    dataIndex: `${e.pr} - ${e.eni}`,
    width: 200,
    fixed: !index ? 'left' : ''
  });

  const eventBreakdownColumns = isBreakdownApplied
    ? breakdown
        .filter((e) => e.eni)
        .map((e, index) => getBreakdownColConfig(e, index))
    : [];

  const globalBreakdownColumns = isBreakdownApplied
    ? breakdown
        .filter((e) => !e.eni)
        .map((e, index) => ({
          ...getBreakdownColConfig(e),
          fixed: !index && !eventBreakdownColumns.length ? 'left' : ''
        }))
    : [];

  const UserCol = !isBreakdownApplied
    ? [
        {
          title: <NonClickableTableHeader title='Users' />,
          dataIndex: 'Grouping',
          fixed: 'left',
          width: isComparisonApplied ? 300 : 150,
          className: isComparisonApplied ? styles.usersColumn : '',
          render: (d) =>
            NoBreakdownUsersColumn(d, breakdown, isComparisonApplied)
        }
      ]
    : [];

  const conversionRateColumn = {
    title: isBreakdownApplied ? (
      getClickableTitleSorter(
        'Conversion Rate',
        {
          key: 'Conversion',
          type: 'numerical',
          subtype: null
        },
        currentSorter,
        handleSorting,
        'right',
        'end',
        'pb-3'
      )
    ) : (
      <NonClickableTableHeader
        verticalAlignment='end'
        alignment='right'
        title='Conversion Rate'
      />
    ),
    dataIndex: 'Conversion',
    className: 'text-right border-none',
    width: 150,
    render: (d) => RenderTotalConversion(d, breakdown, isComparisonApplied)
  };

  const conversionTimeColumn = {
    title: isBreakdownApplied ? (
      getClickableTitleSorter(
        'Time to Convert',
        {
          key: 'Conversion Time',
          type: 'duration',
          subtype: null
        },
        currentSorter,
        handleSorting,
        'right',
        'end',
        'pb-3'
      )
    ) : (
      <NonClickableTableHeader
        verticalAlignment='end'
        alignment='right'
        title='Time to Convert'
      />
    ),
    dataIndex: 'Conversion Time',
    className: 'text-right has-border',
    width: 150,
    render: (d) => RenderConversionTime(d, breakdown, isComparisonApplied)
  };

  const eventColumns = queries.map((_, index) => {
    const queryColumn = {
      title: arrayMapper[index].displayName,
      className: 'bg-white tableParentHeader'
    };

    const percentCol = [];

    if (tableConfig.showPercentage) {
      percentCol.push({
        width: 200,
        className: 'text-right border-none',
        dataIndex: `${arrayMapper[index].displayName}-${index}-percent`,
        title: isBreakdownApplied ? (
          getClickableTitleSorter(
            <SVG name='percentconversion' />,
            {
              key: `${arrayMapper[index].displayName}-${index}-percent`,
              type: 'numerical',
              subtype: null
            },
            currentSorter,
            handleSorting,
            'right',
            'center',
            '',
            'Conv. from prev. step'
          )
        ) : (
          <NonClickableTableHeader
            titleTooltip='Conv. from prev. step'
            verticalAlignment='end'
            alignment='right'
            title={<SVG name='percentconversion' />}
          />
        ),
        render: (d) =>
          !isComparisonApplied ? (
            <>
              <NumFormat number={d} />%
            </>
          ) : (
            compareSkeleton(`${d.percent}%`, `${d.compare_percent}%`)
          )
      });
    }

    const countLabelText = (
      <Text color='grey-2' type='title' level={7} extraClass='mb-0'>
        count of users
      </Text>
    );

    const countLabelSVG = <SVG name='countconversion' />;

    const countColumnHeader = (
      <>
        <ControlledComponent controller={showOnlyCount}>
          <div className='flex col-gap-1 items-center'>
            {countLabelSVG}
            {countLabelText}
          </div>
        </ControlledComponent>
        <ControlledComponent controller={!showOnlyCount}>
          {countLabelSVG}
        </ControlledComponent>
      </>
    );

    const countCol = {
      width: 200,
      className: `text-right ${
        index < queries.length - 1 ? 'has-border' : 'border-none'
      }`,
      dataIndex: `${arrayMapper[index].displayName}-${index}-count`,
      title: isBreakdownApplied ? (
        getClickableTitleSorter(
          countColumnHeader,
          {
            key: `${arrayMapper[index].displayName}-${index}-count`,
            type: 'numerical',
            subtype: null
          },
          currentSorter,
          handleSorting,
          'right',
          'center',
          '',
          'count'
        )
      ) : (
        <NonClickableTableHeader
          titleTooltip='count'
          verticalAlignment={showOnlyCount ? 'center' : 'end'}
          alignment={'end'}
          title={countColumnHeader}
        />
      ),
      render: (d) =>
        !isComparisonApplied ? (
          <NumFormat shortHand number={d} />
        ) : (
          compareSkeleton(
            <NumFormat shortHand number={d.count} />,
            <NumFormat shortHand number={d.compare_count} />
          )
        )
    };
    const timeCol = [];
    if (index > 0 && tableConfig.showDuration) {
      timeCol.push({
        width: 200,
        className: 'text-right border-none',
        dataIndex: `time[${index - 1}-${index}]`,
        title: isBreakdownApplied ? (
          getClickableTitleSorter(
            <SVG name='timeconversion' />,
            {
              key: `time[${index - 1}-${index}]`,
              type: 'duration',
              subtype: null
            },
            currentSorter,
            handleSorting,
            'right',
            'center',
            '',
            'Duration from prev. step'
          )
        ) : (
          <NonClickableTableHeader
            titleTooltip='Duration from prev. step'
            verticalAlignment='end'
            alignment='right'
            title={<SVG name='timeconversion' />}
          />
        ),
        render: (d) =>
          !isComparisonApplied ? d : compareSkeleton(d.time, d.compare_time)
      });
    }
    queryColumn.children = [...percentCol, ...timeCol, countCol];
    return queryColumn;
  });

  const mergedColumns = [
    ...eventBreakdownColumns,
    ...globalBreakdownColumns,
    ...UserCol,
    conversionRateColumn
  ];
  if (eventsCondition !== 'funnel_any_given_event')
    mergedColumns.push(conversionTimeColumn);
  mergedColumns.push(...eventColumns);
  return mergedColumns;
};

export const getTableData = (
  data,
  queries,
  groups,
  arrayMapper,
  currentSorter,
  searchText,
  durations,
  comparisonChartDurations,
  comparisonChartData,
  durationObj,
  comparison_duration,
  isBreakdownApplied,
  eventsCondition
) => {
  if (!isBreakdownApplied) {
    const queryData = {};
    const overallDuration = getOverAllDuration(durations);
    const comparisonOverallDuration = getOverAllDuration(
      comparisonChartDurations
    );
    queries.forEach((_, index) => {
      const percent = !index
        ? 100
        : calculatePercentage(data[index].netCount, data[index - 1].netCount);
      const compare_percent = comparisonChartData
        ? !index
          ? 100
          : calculatePercentage(
              comparisonChartData[index].netCount,
              comparisonChartData[index - 1].netCount
            )
        : null;
      const count = data[index].netCount;
      const compare_count =
        comparisonChartData && comparisonChartData[index].netCount;
      queryData[`${arrayMapper[index].displayName}-${index}-percent`] =
        comparisonChartData
          ? {
              percent,
              compare_percent
            }
          : percent;
      queryData[`${arrayMapper[index].displayName}-${index}-count`] =
        comparisonChartData
          ? {
              count,
              compare_count
            }
          : count;
      if (index < queries.length - 1) {
        const time = getStepDuration(durations, index, index + 1);
        const compare_time =
          comparisonChartData &&
          getStepDuration(comparisonChartDurations, index, index + 1);
        if (eventsCondition !== 'funnel_any_given_event') {
          queryData[`time[${index}-${index + 1}]`] = comparisonChartData
            ? {
                time,
                compare_time
              }
            : time;
        }
      }
    });
    const conversion = data[data.length - 1].value;
    const comparison_conversion =
      comparisonChartData &&
      comparisonChartData[comparisonChartData.length - 1].value;

    const result = {
      index: 0,
      Grouping: comparisonChartData
        ? { durationObj, comparison_duration }
        : 'All',
      Conversion: comparisonChartData
        ? { conversion, comparison_conversion }
        : conversion
    };

    if (eventsCondition !== 'funnel_any_given_event') {
      const conversionTime = comparisonChartData
        ? { overallDuration, comparisonOverallDuration }
        : overallDuration;
      result['Conversion Time'] = conversionTime;
    }
    return [{ ...result, ...queryData }];
  }

  const isComparisonApplied = comparisonChartData != null;

  const compareGroupsByName = getCompareGroupsByName({
    compareGroups: comparisonChartData
  });

  const appliedGroups = groups.map((group) => {
    const compareGroup = compareGroupsByName[group.name];
    const eventPercentages = arrayMapper.reduce(
      (agg, currentItem, currentIndex) => {
        const prevItem = arrayMapper[currentIndex - 1];
        const percentageValue = !currentIndex
          ? 100
          : calculatePercentage(
              group[`${currentItem.displayName}-${currentIndex}-count`],
              group[`${prevItem.displayName}-${currentIndex - 1}-count`]
            );
        const comparePercentageValue = !currentIndex
          ? 100
          : compareGroup != null
          ? calculatePercentage(
              compareGroup[`${currentItem.displayName}-${currentIndex}-count`],
              compareGroup[`${prevItem.displayName}-${currentIndex - 1}-count`]
            )
          : 0;
        return {
          ...agg,
          // if comparison is applied, we have to pass both count and compare_count, time and compare_time, percent and compare_percent
          ...(isComparisonApplied && {
            [`${currentItem.displayName}-${currentIndex}-count`]: {
              count: group[`${currentItem.displayName}-${currentIndex}-count`],
              compare_count:
                compareGroup != null
                  ? compareGroup[
                      `${currentItem.displayName}-${currentIndex}-count`
                    ]
                  : 0
            },
            ...(currentIndex < queries.length - 1 &&
              eventsCondition !== 'funnel_any_given_event' && {
                [`time[${currentIndex}-${currentIndex + 1}]`]: {
                  time: group[`time[${currentIndex}-${currentIndex + 1}]`],
                  compare_time:
                    compareGroup != null
                      ? compareGroup[
                          `time[${currentIndex}-${currentIndex + 1}]`
                        ]
                      : '0s'
                }
              })
          }),
          [`${currentItem.displayName}-${currentIndex}-percent`]:
            isComparisonApplied
              ? {
                  percent: percentageValue,
                  compare_percent: comparePercentageValue
                }
              : percentageValue
        };
      },
      {}
    );
    const result = {
      ...group,
      ...eventPercentages,
      // if comparison is applied, we have to pass both conversion and comparison_conversion, overallDuration and comparisonOverallDuration
      ...(isComparisonApplied && {
        Conversion: {
          conversion: group.Conversion,
          comparison_conversion:
            compareGroup != null ? compareGroup.Conversion : '0'
        }
      })
    };

    if (eventsCondition !== 'funnel_any_given_event') {
      const conversionTime = {
        overallDuration: group['Conversion Time'],
        comparisonOverallDuration:
          compareGroup != null ? compareGroup['Conversion Time'] : 'N/A'
      };
      result['Conversion Time'] = conversionTime;
    }
    return result;
  });
  const filteredGroups = appliedGroups.filter(
    (elem) => elem.name.toLowerCase().indexOf(searchText.toLowerCase()) > -1
  );

  return SortResults(filteredGroups, currentSorter);
};

export const generateUngroupedChartsData = (response, arrayMapper) => {
  if (!response) {
    return [];
  }

  console.log('funnels generateUngroupedChartsData');

  const netCounts = response.rows[0].filter((elem) => typeof elem === 'number');
  const result = [];
  let index = 0;

  while (index < arrayMapper.length) {
    if (index === 0) {
      result.push({
        event: arrayMapper[index].mapper,
        netCount: netCounts[index],
        value: 100
      });
    } else {
      result.push({
        event: arrayMapper[index].mapper,
        netCount: netCounts[index],
        value: calculatePercentage(netCounts[index], netCounts[0])
      });
    }
    index++;
  }
  return result;
};

export const checkForWindowSizeChange = (callback) => {
  if (
    window.outerWidth !== windowSize.w ||
    window.outerHeight !== windowSize.h
  ) {
    setTimeout(() => {
      windowSize.w = window.outerWidth; // update object with current window properties
      windowSize.h = window.outerHeight;
      windowSize.iw = window.innerWidth;
      windowSize.ih = window.innerHeight;
    }, 0);
    callback();
  } else if (
    window.innerWidth + window.innerWidth * 0.05 < windowSize.iw ||
    window.innerWidth - window.innerWidth * 0.05 > windowSize.iw
  ) {
    setTimeout(() => {
      windowSize.iw = window.innerWidth;
    }, 0);
    callback();
  }
};

export const getOverAllDuration = (durationsObj) => {
  if (durationsObj && durationsObj.metrics) {
    const durationMetric = durationsObj.metrics.find(
      (d) => d.title === 'MetaStepTimeInfo'
    );
    if (durationMetric && durationMetric.rows && durationMetric.rows.length) {
      try {
        let total = 0;
        durationMetric.rows[0].forEach((r) => {
          total += Number(r);
        });
        return formatDuration(total);
      } catch (err) {
        return 'NA';
      }
    }
  }
  return 'NA';
};

export const getStepDuration = (durationsObj, index1, index2) => {
  let durationVal = 'NA';
  if (durationsObj && durationsObj.metrics) {
    const durationMetric = durationsObj.metrics.find(
      (d) => d.title === 'MetaStepTimeInfo'
    );
    if (
      durationMetric &&
      durationMetric.headers &&
      durationMetric.headers.length
    ) {
      try {
        const stepIndex = durationMetric.headers.findIndex(
          (elem) => elem === `step_${index1}_${index2}_time`
        );
        if (stepIndex > -1) {
          durationVal = formatDuration(durationMetric.rows[0][stepIndex]);
        }
      } catch (err) {
        console.log(err);
      }
    }
  }
  return durationVal;
};

const getConvertedValuesForScatterPlot = (metric, originalValue) => {
  if (metric === 'Conversion Time' || metric.includes('time')) {
    return Number(getDurationInSeconds(originalValue));
  }
  return originalValue.value || originalValue.value === 0
    ? Number(originalValue.value)
    : Number(originalValue);
};

export const getScatterPlotChartData = (
  visibleData,
  xAxisMetric,
  yAxisMetric
) => {
  console.log('funnels getScatterPlotChartData');
  const categories = [];
  const plotData = visibleData.map((d) => {
    categories.push(d.name);
    const xValue = getConvertedValuesForScatterPlot(
      xAxisMetric,
      d[xAxisMetric]
    );
    const yValue = getConvertedValuesForScatterPlot(
      yAxisMetric,
      d[yAxisMetric]
    );
    return [xValue, yValue];
  });
  return {
    series: [
      {
        color: CHART_COLOR_1,
        data: plotData
      }
    ],
    categories
  };
};

export const getAxisMetricOptions = (arrayMapper) => {
  console.log('funnels getAxisMetricOptions');
  const result = [
    {
      title: 'Conversion',
      value: 'Conversion'
    },
    {
      title: 'Conversion Time (in seconds)',
      value: 'Conversion Time'
    }
  ];
  for (let i = 0; i < arrayMapper.length; i++) {
    result.push({
      title: `${
        arrayMapper[i].displayName || arrayMapper[i].displayName
      } (Event ${i + 1})`,
      value: `${
        arrayMapper[i].displayName || arrayMapper[i].displayName
      }-${i}-count`
    });
    if (i < arrayMapper.length - 1) {
      result.push({
        title: `Conversion Time from Event ${i + 1} to Event ${i + 2}`,
        value: `time[${i}-${i + 1}]`
      });
    }
  }
  return result;
};
