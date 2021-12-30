import React from 'react';
import moment from 'moment';
import {
  calculatePercentage,
  formatDuration,
  getClickableTitleSorter,
  SortResults,
  SortData,
  getDurationInSeconds,
} from '../../../utils/dataFormatter';
import {
  Number as NumFormat,
  Text,
} from '../../../components/factorsComponents';
import styles from './index.module.scss';
import { parseForDateTimeLabel } from '../EventsAnalytics/SingleEventSingleBreakdown/utils';
import { GROUPED_MAX_ALLOWED_VISIBLE_PROPERTIES } from '../../../utils/constants';
import DurationCol from './FunnelsResultTable/DurationCol';
import { displayName } from 'Components/FaFilterSelect/utils';

const windowSize = {
  w: window.outerWidth,
  h: window.outerHeight,
  iw: window.innerWidth,
  ih: window.innerHeight,
};

export const getVisibleData = (data, sorter) => {
  const result = SortResults(data, sorter).slice(
    0,
    GROUPED_MAX_ALLOWED_VISIBLE_PROPERTIES
  );
  return result;
};

const NoBreakdownUsersColumn = (d, breakdown, isComparisonApplied) => {
  if (breakdown.length) {
    if (d.includes('$no_group')) {
      return 'Overall';
    } else {
      return d;
    }
  } else {
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
    } else {
      return d;
    }
  }
};

export const grnByIndex = (headersSlice, breakdowns) => {
  let grns = [];
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
  const firstEventIdx = headers.findIndex((header) => header === 'step_0');
  const netConversionIndex = headers.findIndex(
    (header) => header === 'conversion_overall'
  );

  const grns = grnByIndex(headers.slice(0, firstEventIdx), breakdowns);

  const eventsData = arrayMapper.map((am, index) => {
    return {
      index: index + 1,
      name: am.mapper,
      data: {},
    };
  });

  const result = rows.map((row, index) => {
    const breakdownData = {};
    const breakdownVals = row
      .slice(0, firstEventIdx)
      .map((vl) => (displayName[vl] ? displayName[vl] : vl));
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

      groupEventData[`${am.displayName}-${idx}`] = {
        percentage: percent,
        value: row[eventIdx],
      };
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

    return {
      index,
      name,
      value: `${value}%`,
      Conversion: value, //used for sorting, value will be removed soon
      nonConvertedName,
      'Conversion Time': formatDuration(totalDuration),
      ...groupEventData,
      ...breakdownData,
      ...timeData,
    };
  });

  return {
    groups: result,
    events: eventsData,
  };
};

const compareSkeleton = (val1, val2) => {
  return (
    <div className='flex flex-col'>
      <Text
        type='title'
        weight='normal'
        color='grey-8'
        extraClass='text-sm mb-0'
      >
        {val1}
      </Text>
      <Text type='title' weight='normal' color='grey' extraClass='text-xs mb-0'>
        {val2}
      </Text>
    </div>
  );
};

const RenderTotalConversion = (d, breakdown, isComparisonApplied) => {
  if (breakdown.length || !isComparisonApplied) {
    return d + '%';
  } else {
    return compareSkeleton(d.conversion + '%', d.comparsion_conversion + '%');
  }
};

const RenderConversionTime = (d, breakdown, isComparisonApplied) => {
  if (breakdown.length || !isComparisonApplied) {
    return d;
  } else {
    return compareSkeleton(d.overallDuration, d.comparisonOverallDuration);
  }
};

const RenderEventData = (d, breakdown, isComparisonApplied) => {
  if (breakdown.length || !isComparisonApplied) {
    return (
      <>
        <NumFormat number={d.value} /> ({d.percentage}%)
      </>
    );
  } else {
    const val1 = (
      <>
        <NumFormat number={d.value} /> ({d.percentage}%)
      </>
    );
    const val2 = (
      <>
        <NumFormat number={d.compare_count} /> ({d.compare_percent}%)
      </>
    );
    return compareSkeleton(val1, val2);
  }
};

const RenderDurations = (d, breakdown, isComparisonApplied) => {
  if (breakdown.length || !isComparisonApplied) {
    return d;
  } else {
    return compareSkeleton(d.time, d.compare_time);
  }
};

export const getBreakdownTitle = (breakdown, userPropNames, eventPropNames) => {
  const charArr = ['1', '2', '3', '4', '5', '6'];
  const displayTitle =
    breakdown.en === 'user'
      ? userPropNames[breakdown.pr] || breakdown.pr
      : breakdown.en === 'event'
      ? eventPropNames[breakdown.pr] || breakdown.pr
      : breakdown.pr;

  if (!breakdown.eni) {
    return displayTitle;
  }
  return (
    <div className='flex items-center'>
      <div className='mr-1'>{displayTitle} of </div>
      <div
        style={{ backgroundColor: '#3E516C' }}
        className='text-white w-4 h-4 flex justify-center items-center rounded-full font-semibold leading-5 text-xs'
      >
        {charArr[breakdown.eni - 1]}
      </div>
    </div>
  );
};

export const generateTableColumns = (
  queries,
  currentSorter,
  handleSorting,
  arrayMapper,
  isComparisonApplied,
  resultData,
  userPropNames,
  eventPropNames
) => {
  console.log('funnels generateTableColumns');
  let breakdown = resultData?.meta?.query?.gbp;

  const isBreakdownApplied =
    !!breakdown && Array.isArray(breakdown) && breakdown.length > 0;

  if (isBreakdownApplied) {
    breakdown = SortData(breakdown, 'eni', 'ascend');
  }
  const eventBreakdownColumns = isBreakdownApplied
    ? breakdown
        .filter((e) => e.eni)
        .map((e, index) => {
          return {
            title: getClickableTitleSorter(
              getBreakdownTitle(e, userPropNames, eventPropNames),
              {
                key: `${e.pr} - ${e.eni}`,
                type: e.pty,
                subtype: e.grn,
              },
              currentSorter,
              handleSorting
            ),
            dataIndex: `${e.pr} - ${e.eni}`,
            fixed: !index ? 'left' : '',
            width: 200,
          };
        })
    : [];

  const globalBreakdownColumns = isBreakdownApplied
    ? breakdown
        .filter((e) => !e.eni)
        .map((e, index) => {
          return {
            title: getClickableTitleSorter(
              getBreakdownTitle(e, userPropNames, eventPropNames),
              {
                key: `${e.pr} - ${e.eni}`,
                type: e.pty,
                subtype: e.grn,
              },
              currentSorter,
              handleSorting
            ),
            dataIndex: `${e.pr} - ${e.eni}`,
            fixed: !index && !eventBreakdownColumns.length ? 'left' : '',
            width: 200,
          };
        })
    : [];

  const UserCol = !isBreakdownApplied
    ? [
        {
          title: 'Users',
          dataIndex: 'Grouping',
          fixed: 'left',
          width: isComparisonApplied ? 300 : 100,
          className: isComparisonApplied ? styles.usersColumn : '',
          render: (d) =>
            NoBreakdownUsersColumn(d, breakdown, isComparisonApplied),
        },
      ]
    : [];
  const result = [
    ...eventBreakdownColumns,
    ...globalBreakdownColumns,
    ...UserCol,
    {
      title: isBreakdownApplied
        ? getClickableTitleSorter(
            'Total Conversion',
            {
              key: `Conversion`,
              type: 'numerical',
              subtype: null,
            },
            currentSorter,
            handleSorting
          )
        : 'Total Conversion',
      dataIndex: 'Conversion',
      width: 150,
      render: (d) => RenderTotalConversion(d, breakdown, isComparisonApplied),
    },
    {
      title: isBreakdownApplied
        ? getClickableTitleSorter(
            'Conversion Time',
            {
              key: `Conversion Time`,
              type: 'duration',
              subtype: null,
            },
            currentSorter,
            handleSorting
          )
        : 'Conversion Time',
      dataIndex: 'Conversion Time',
      width: 150,
      render: (d) => RenderConversionTime(d, breakdown, isComparisonApplied),
    },
  ];
  const eventColumns = [];

  const clockCol = <DurationCol />;

  queries.forEach((q, index) => {
    eventColumns.push({
      title: isBreakdownApplied
        ? getClickableTitleSorter(
            arrayMapper[index].displayName,
            {
              key: `${arrayMapper[index].displayName}-${index}`,
              type: 'numerical',
              subtype: null,
            },
            currentSorter,
            handleSorting
          )
        : arrayMapper[index].displayName,
      dataIndex: `${arrayMapper[index].displayName}-${index}`,
      width: 200,
      render: (d) => RenderEventData(d, breakdown, isComparisonApplied),
    });
    if (index < queries.length - 1) {
      eventColumns.push({
        title: isBreakdownApplied
          ? getClickableTitleSorter(
              clockCol,
              {
                key: `time[${index}-${index + 1}]`,
                type: 'duration',
                subtype: null,
              },
              currentSorter,
              handleSorting
            )
          : clockCol,
        dataIndex: `time[${index}-${index + 1}]`,
        width: isBreakdownApplied ? 90 : 75,
        render: (d) => RenderDurations(d, breakdown, isComparisonApplied),
      });
    }
  });

  const blankCol = {
    title: '',
    dataIndex: '',
    width: 37,
    fixed: 'left',
  };
  if (isBreakdownApplied) {
    return [...result, ...eventColumns];
  } else {
    return [blankCol, ...result, ...eventColumns];
  }
};

export const generateTableData = (
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
  resultData
) => {
  console.log('funnels generateTableData');
  const breakdown = resultData?.meta?.query?.gbp;
  const isBreakdownApplied =
    !!breakdown && Array.isArray(breakdown) && breakdown.length > 0;
  if (!isBreakdownApplied) {
    const queryData = {};
    const overallDuration = getOverAllDuration(durations);
    const comparisonOverallDuration = getOverAllDuration(
      comparisonChartDurations
    );
    queries.forEach((_, index) => {
      queryData[`${arrayMapper[index].displayName}-${index}`] = {
        percentage: data[index].value,
        value: data[index].netCount,
        compare_percent:
          comparisonChartData && comparisonChartData[index].value,
        compare_count:
          comparisonChartData && comparisonChartData[index].netCount,
      };
      if (index < queries.length - 1) {
        const time = getStepDuration(durations, index, index + 1);
        const compare_time =
          comparisonChartData &&
          getStepDuration(comparisonChartDurations, index, index + 1);
        queryData[`time[${index}-${index + 1}]`] = comparisonChartData
          ? {
              time,
              compare_time,
            }
          : time;
      }
    });
    const conversion = data[data.length - 1].value;
    const comparsion_conversion =
      comparisonChartData &&
      comparisonChartData[comparisonChartData.length - 1].value;
    return [
      {
        index: 0,
        Grouping: comparisonChartData
          ? { durationObj, comparison_duration }
          : 'All',
        Conversion: comparisonChartData
          ? { conversion, comparsion_conversion }
          : conversion,
        'Conversion Time': comparisonChartData
          ? { overallDuration, comparisonOverallDuration }
          : overallDuration,
        ...queryData,
      },
    ];
  } else {
    const appliedGroups = groups.filter(
      (elem) => elem.name.toLowerCase().indexOf(searchText.toLowerCase()) > -1
    );
    return SortResults(appliedGroups, currentSorter);
  }
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
        value: 100,
      });
    } else {
      result.push({
        event: arrayMapper[index].mapper,
        netCount: netCounts[index],
        value: calculatePercentage(netCounts[index], netCounts[0]),
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
  }

  // if the window doesn't resize but the content inside does by + or - 5%
  else if (
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
        color: '#4D7DB4',
        data: plotData,
      },
    ],
    categories,
  };
};

export const getAxisMetricOptions = (arrayMapper) => {
  console.log('funnels getAxisMetricOptions');
  const result = [
    {
      title: 'Conversion',
      value: 'Conversion',
    },
    {
      title: 'Conversion Time (in seconds)',
      value: 'Conversion Time',
    },
  ];
  for (let i = 0; i < arrayMapper.length; i++) {
    result.push({
      title: `${
        arrayMapper[i].displayName || arrayMapper[i].displayName
      } (Event ${i + 1})`,
      value: `${arrayMapper[i].displayName || arrayMapper[i].displayName}-${i}`,
    });
    if (i < arrayMapper.length - 1) {
      result.push({
        title: `Conversion Time from Event ${i + 1} to Event ${i + 2}`,
        value: `time[${i}-${i + 1}]`,
      });
    }
  }
  return result;
};
