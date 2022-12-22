import React from 'react';
import moment from 'moment';
import FunnelChart from './Chart';
import { Text } from 'Components/factorsComponents';
import TopLegends from 'Components/GroupedBarChart/TopLegends';
import MetricChart from 'Components/MetricChart/MetricChart';
import {
  CHART_TYPE_BARCHART,
  FUNNEL_CHART_MARGIN
} from '../../../../utils/constants';
import { generateColors } from '../../../../utils/dataFormatter';
import { CONVERSION_RATE_LABEL } from './ungroupedChart.constants';
import { get } from 'lodash';

function ChartSection({
  arrayMapper,
  section,
  chartData,
  chartDurations,
  comparisonChartData,
  comparisonChartDurations,
  durationObj,
  comparison_duration,
  chartType
}) {
  if (chartType === CHART_TYPE_BARCHART) {
    if (!comparisonChartData) {
      return (
        <FunnelChart
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
            <FunnelChart
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
            <FunnelChart
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

  return (
    <MetricChart
      value={get(chartData, `${chartData.length - 1}.value`, 0)}
      valueType='percentage'
      headerTitle={CONVERSION_RATE_LABEL}
      showComparison={comparisonChartData != null}
      compareValue={
        comparisonChartData != null
          ? comparisonChartData[comparisonChartData.length - 1].value
          : 0
      }
    />
  );
}

export default ChartSection;
