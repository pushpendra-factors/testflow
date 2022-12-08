import React, { useMemo, useContext, memo } from 'react';
import { generateUngroupedChartsData } from '../utils';
import FunnelsResultTable from '../FunnelsResultTable';
import { DASHBOARD_MODAL } from '../../../../utils/constants';
import { CoreQueryContext } from '../../../../contexts/CoreQueryContext';
import ChartSection from './ChartSection';
import { EMPTY_ARRAY } from 'Utils/global';

function UngroupedChartComponent({
  resultState,
  queries,
  section,
  arrayMapper,
  durationObj,
  comparison_data,
  comparison_duration,
  tableConfig,
  tableConfigPopoverContent
}) {
  const chartData = useMemo(() => {
    return generateUngroupedChartsData(resultState.data, arrayMapper);
  }, [resultState.data, arrayMapper]);

  const comparisonChartData = useMemo(() => {
    return comparison_data.data
      ? generateUngroupedChartsData(comparison_data.data, arrayMapper)
      : null;
  }, [comparison_data.data, arrayMapper]);

  if (!chartData.length) {
    return null;
  }

  return (
    <div className='flex items-center justify-center flex-col'>
      <ChartSection
        arrayMapper={arrayMapper}
        section={section}
        chartData={chartData}
        chartDurations={resultState.data.meta}
        comparisonChartData={comparisonChartData}
        comparisonChartDurations={
          comparisonChartData ? comparison_data.data.meta : null
        }
        durationObj={durationObj}
        comparison_duration={comparison_duration}
      />

      <div className='mt-12 w-full'>
        <FunnelsResultTable
          isWidgetModal={section === DASHBOARD_MODAL}
          breakdown={EMPTY_ARRAY}
          queries={queries}
          groups={EMPTY_ARRAY}
          arrayMapper={arrayMapper}
          chartData={chartData}
          comparisonChartData={comparisonChartData}
          durations={resultState.data.meta}
          comparisonChartDurations={
            comparisonChartData ? comparison_data.data.meta : null
          }
          durationObj={durationObj}
          comparison_duration={comparison_duration}
          resultData={resultState.data}
          tableConfig={tableConfig}
          tableConfigPopoverContent={tableConfigPopoverContent}
        />
      </div>
    </div>
  );
}

const UngroupedChartMemoized = memo(UngroupedChartComponent);

function UngroupedChart(props) {
  const {
    coreQueryState: { comparison_data, comparison_duration }
  } = useContext(CoreQueryContext);

  return (
    <UngroupedChartMemoized
      comparison_data={comparison_data}
      comparison_duration={comparison_duration}
      {...props}
    />
  );
}

export default memo(UngroupedChart);
