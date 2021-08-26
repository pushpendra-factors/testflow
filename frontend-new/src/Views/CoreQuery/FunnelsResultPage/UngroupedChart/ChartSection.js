import React from 'react';
import moment from 'moment';
import Chart from './Chart';
import { FUNNEL_CHART_MARGIN } from '../../../../utils/constants';
import { Text } from '../../../../components/factorsComponents';
import TopLegends from '../../../../components/GroupedBarChart/TopLegends';
import { generateColors } from '../../../../utils/dataFormatter';

function ChartSection({
  arrayMapper,
  section,
  chartData,
  chartDurations,
  comparisonChartData,
  comparisonChartDurations,
  durationObj,
  comparison_duration,
}) {
  if (!comparisonChartData) {
    return (
      <Chart
        chartData={chartData}
        arrayMapper={arrayMapper}
        section={section}
        durations={chartDurations}
      />
    );
  }

  const colors = generateColors(arrayMapper.length);

  return (
    <>
      <div className='flex w-full'>
        <div className='w-1/2 flex flex-col'>
          <Chart
            chartData={chartData}
            arrayMapper={arrayMapper}
            section={section}
            durations={chartDurations}
            title='chart'
            showXAxisLabels={false}
            margin={{ ...FUNNEL_CHART_MARGIN, right: 0, bottom: 10 }}
            showStripes={true}
          />
          <Text
            type='title'
            weight='normal'
            color='grey-2'
            lineHeight='medium'
            extraClass='text-xs mb-0 flex item-center justify-center'
          >
            {`${moment(durationObj.from).format('MMM DD, YYYY')} - ${moment(
              durationObj.to
            ).format('MMM DD, YYYY')}`}
          </Text>
        </div>
        <div
          style={{ borderLeft: '2px solid #E7E9ED' }}
          className='w-1/2 flex flex-col'
        >
          <Chart
            chartData={comparisonChartData}
            arrayMapper={arrayMapper}
            section={section}
            durations={comparisonChartDurations}
            title='compareChart'
            showXAxisLabels={false}
            showYAxisLabels={false}
            margin={{ ...FUNNEL_CHART_MARGIN, left: 0, bottom: 10 }}
          />
          <Text
            type='title'
            weight='normal'
            color='grey-2'
            lineHeight='medium'
            extraClass='text-xs mb-0 flex item-center justify-center'
          >
            {`${moment(comparison_duration.from).format(
              'MMM DD, YYYY'
            )} - ${moment(comparison_duration.to).format('MMM DD, YYYY')}`}
          </Text>
        </div>
      </div>
      <div className='mt-4'>
        <TopLegends
          cardSize={1}
          legends={arrayMapper.map((d) => d.eventName)}
          colors={colors.map((c) => c)}
        />
      </div>
    </>
  );
}

export default ChartSection;