import React, { useMemo, useContext } from 'react';
import { generateUngroupedChartsData } from '../utils';
import FunnelsResultTable from '../FunnelsResultTable';
import { DASHBOARD_MODAL } from '../../../../utils/constants';
import { CoreQueryContext } from '../../../../contexts/CoreQueryContext';
import ChartSection from './ChartSection';

function UngroupedChart({
  resultState,
  queries,
  section,
  arrayMapper,
  durationObj,
}) {
  const {
    coreQueryState: { comparison_data, comparison_duration },
  } = useContext(CoreQueryContext);

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
          breakdown={[]}
          queries={queries}
          groups={[]}
          arrayMapper={arrayMapper}
          chartData={chartData}
          comparisonChartData={comparisonChartData}
          durations={resultState.data.meta}
          comparisonChartDurations={
            comparisonChartData ? comparison_data.data.meta : null
          }
          durationObj={durationObj}
          comparison_duration={comparison_duration}
        />
      </div>
    </div>
  );
}

export default UngroupedChart;
