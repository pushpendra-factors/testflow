import React from 'react';
import moment from 'moment';
import {
  calculatePercentage,
  getTitleWithSorter,
  formatDuration,
} from '../../../utils/dataFormatter';
import {
  SVG,
  Number as NumFormat,
  Text,
} from '../../../components/factorsComponents';
import styles from './index.module.scss';
import { parseForDateTimeLabel } from '../EventsAnalytics/SingleEventSingleBreakdown/utils';

const windowSize = {
  w: window.outerWidth,
  h: window.outerHeight,
  iw: window.innerWidth,
  ih: window.innerHeight,
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

export const generateGroupedChartsData = (
  response,
  queries,
  groups,
  arrayMapper,
  grn
) => {
  if (!response) {
    return [];
  }
  const result = groups
    .filter((g) => g.is_visible)
    .map((g) => {
      return { name: g.name };
    });
  const firstEventIdx = response.headers.findIndex((elem) => elem === 'step_0');
  let breakdowns = [...response.meta.query.gbp];
  let grns = grnByIndex(response.headers.slice(0, firstEventIdx), breakdowns);
  response.rows.forEach((row) => {
    const breakdownName = row
      .slice(0, firstEventIdx)
      .map((x, ind) => parseForDateTimeLabel(grns[ind], x))
      .join(',');

    const obj = result.find((r) => r.name === breakdownName);
    if (obj) {
      const netCounts = row.filter((val) => typeof val === 'number');
      queries.forEach((_, idx) => {
        const eventIdx = response.headers.findIndex(
          (elem) => elem === `step_${idx}`
        );
        obj[arrayMapper[idx].mapper] = calculatePercentage(
          row[eventIdx],
          netCounts[0]
        );
      });
    }
  });
  return result;
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

export const generateGroups = (response, maxAllowedVisibleProperties, grn) => {
  if (!response) {
    return [];
  }
  let breakdowns = [...response.meta.query.gbp];
  const firstEventIdx = response.headers.findIndex((elem) => elem === 'step_0');
  let grns = grnByIndex(response.headers.slice(0, firstEventIdx), breakdowns);
  const result = response.rows.map((elem, index) => {
    const row = elem.map((item) => {
      return item;
    });
    const netCounts = row.filter((row) => typeof row === 'number');
    const name = row
      .slice(0, firstEventIdx)
      ?.map((label, ind) => {
        return parseForDateTimeLabel(grns[ind], label);
      })
      ?.join(',');
    return {
      index,
      name,
      value:
        calculatePercentage(netCounts[netCounts.length - 1], netCounts[0]) +
        '%',
      is_visible: index < maxAllowedVisibleProperties ? true : false,
    };
  });
  return result;
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
    return d;
  } else {
    return compareSkeleton(d.conversion, d.comparsion_conversion);
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
        <NumFormat number={d.count} /> ({d.percentage}%)
      </>
    );
  } else {
    const val1 = (
      <>
        <NumFormat number={d.count} /> ({d.percentage}%)
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

export const generateTableColumns = (
  breakdown,
  queries,
  currentSorter,
  handleSorting,
  arrayMapper,
  isComparisonApplied
) => {
  const result = [
    {
      title: breakdown.length ? 'Grouping' : 'Users',
      dataIndex: 'Grouping',
      fixed: 'left',
      width: isComparisonApplied ? 300 : 100,
      className: isComparisonApplied ? styles.usersColumn : '',
      render: (d) => NoBreakdownUsersColumn(d, breakdown, isComparisonApplied),
    },
    {
      title: 'Total Conversion',
      dataIndex: 'Conversion',
      width: 100,
      render: (d) => RenderTotalConversion(d, breakdown, isComparisonApplied),
    },
    {
      title: 'Conversion Time',
      dataIndex: 'Converstion Time',
      width: 100,
      render: (d) => RenderConversionTime(d, breakdown, isComparisonApplied),
    },
  ];
  const eventColumns = [];
  queries.forEach((elem, index) => {
    eventColumns.push({
      title: breakdown.length
        ? getTitleWithSorter(
            arrayMapper[index].displayName,
            arrayMapper[index].mapper,
            currentSorter,
            handleSorting
          )
        : arrayMapper[index].displayName,
      dataIndex: arrayMapper[index].mapper,
      width: 200,
      render: (d) => RenderEventData(d, breakdown, isComparisonApplied),
    });
    if (index < queries.length - 1) {
      eventColumns.push({
        title: (
          <div className='flex items-center justify-between'>
            <div className='text-base' style={{ color: '#8692A3' }}>
              &mdash;
            </div>
            <SVG name='clock' />
            <div
              className='text-base'
              style={{ color: '#8692A3', marginTop: '2px' }}
            >
              &rarr;
            </div>
          </div>
        ),
        dataIndex: `time[${index}-${index + 1}]`,
        width: 75,
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
  if (breakdown.length) {
    return [...result, ...eventColumns];
  } else {
    return [blankCol, ...result, ...eventColumns];
  }
};

export const generateTableData = (
  data,
  breakdown,
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
  if (!breakdown.length) {
    const queryData = {};
    const overallDuration = getOverAllDuration(durations);
    const comparisonOverallDuration = getOverAllDuration(
      comparisonChartDurations
    );
    queries.forEach((q, index) => {
      queryData[arrayMapper[index].mapper] = {
        percentage: data[index].value,
        count: data[index].netCount,
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
    const conversion = data[data.length - 1].value + '%';
    const comparsion_conversion =
      comparisonChartData &&
      comparisonChartData[comparisonChartData.length - 1].value + '%';
    return [
      {
        index: 0,
        Grouping: comparisonChartData
          ? { durationObj, comparison_duration }
          : 'All',
        Conversion: comparisonChartData
          ? { conversion, comparsion_conversion }
          : conversion,
        'Converstion Time': comparisonChartData
          ? { overallDuration, comparisonOverallDuration }
          : overallDuration,
        ...queryData,
      },
    ];
  } else {
    const appliedGroups = groups
      .map((elem) => elem.name)
      .filter(
        (elem) => elem.toLowerCase().indexOf(searchText.toLowerCase()) > -1
      );
    const durationMetric = durations.metrics.find(
      (elem) => elem.title === 'MetaStepTimeInfo'
    );
    const firstEventIdx = durationMetric.headers.findIndex(
      (elem) => elem === 'step_0_1_time'
    );

    let breakdowns = [...resultData?.meta?.query?.gbp];
    let grns = grnByIndex(
      resultData?.headers?.slice(0, firstEventIdx),
      breakdowns
    );

    const result = appliedGroups.map((grp, index) => {
      const group = grp;
      const durationGrp = durationMetric.rows.find(
        (elem) =>
          elem
            .slice(0, firstEventIdx)
            .map((x, ind) => parseForDateTimeLabel(grns[ind], x))
            .join(',') === grp
      );
      const eventsData = {};
      let totalDuration = 0;
      data.forEach((d, idx) => {
        eventsData[`${d.name}`] = {
          percentage: calculatePercentage(d.data[group], data[0].data[group]),
          count: d.data[group],
        };
        if (idx < data.length - 1) {
          const durationIdx = durationMetric.headers.findIndex(
            (elem) => elem === `step_${idx}_${idx + 1}_time`
          );
          eventsData[`time[${idx}-${idx + 1}]`] = durationGrp
            ? formatDuration(durationGrp[durationIdx])
            : 'NA';
          totalDuration += durationGrp ? Number(durationGrp[durationIdx]) : 0;
        }
      });
      // const groupLabel = group.split(',')?.map((lbl) => );
      return {
        index,
        Grouping: group,
        'Converstion Time': formatDuration(totalDuration),
        Conversion:
          calculatePercentage(
            data[data.length - 1].data[group],
            data[0].data[group]
          ) + '%',
        ...eventsData,
      };
    });
    if (currentSorter.key) {
      const sortKey = currentSorter.key;
      const { order } = currentSorter;
      result.sort((a, b) => {
        if (order === 'ascend') {
          return parseFloat(a[sortKey].count) >= parseFloat(b[sortKey].count)
            ? 1
            : -1;
        }
        if (order === 'descend') {
          return parseFloat(a[sortKey].count) <= parseFloat(b[sortKey].count)
            ? 1
            : -1;
        }
        return 0;
      });
    }
    return result;
  }
};

export const generateUngroupedChartsData = (response, arrayMapper) => {
  if (!response) {
    return [];
  }

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

export const generateEventsData = (response, queries, arrayMapper) => {
  if (!response) {
    return [];
  }
  const firstEventIdx = response.headers.findIndex((elem) => elem === 'step_0');
  const breakdowns = [...response.meta.query.gbp];
  const grns = grnByIndex(response.headers.slice(0, firstEventIdx), breakdowns);
  const result = queries.map((q, idx) => {
    const data = {};
    response.rows.forEach((r) => {
      const name = r
        .slice(0, firstEventIdx)
        ?.map((label, ind) => {
          return parseForDateTimeLabel(grns[ind], label);
        })
        ?.join(',');
      const netCounts = r.filter((elem) => typeof elem === 'number');
      data[name] = netCounts[idx];
    });
    return {
      index: idx + 1,
      data,
      name: arrayMapper[idx].mapper,
    };
  });
  return result;
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
