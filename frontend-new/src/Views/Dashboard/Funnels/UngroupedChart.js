import React, { useContext, useMemo } from 'react';
import { generateUngroupedChartsData } from '../../CoreQuery/FunnelsResultPage/utils';
import Chart from '../../CoreQuery/FunnelsResultPage/UngroupedChart/Chart';
import FunnelsResultTable from '../../CoreQuery/FunnelsResultPage/FunnelsResultTable';
import {
  CHART_TYPE_BARCHART,
  CHART_TYPE_TABLE,
  DASHBOARD_WIDGET_UNGROUPED_FUNNEL_CHART_HEIGHT,
} from '../../../utils/constants';
import { DashboardContext } from '../../../contexts/DashboardContext';

function UngroupedChart({
  resultState,
  queries,
  chartType,
  unit,
  arrayMapper,
  section,
  breakdown,
}) {
  const { handleEditQuery } = useContext(DashboardContext);
  const groups = useMemo(() => {
    return [];
  }, []);

  const chartData = useMemo(() => {
    return generateUngroupedChartsData(resultState.data, arrayMapper);
  }, [resultState.data, arrayMapper]);

  if (!chartData.length) {
    return null;
  }

  let chartContent = null;

  if (chartType === CHART_TYPE_BARCHART) {
    chartContent = (
      <Chart
        title={unit.id}
        chartData={chartData}
        cardSize={unit.cardSize}
        arrayMapper={arrayMapper}
        height={DASHBOARD_WIDGET_UNGROUPED_FUNNEL_CHART_HEIGHT}
        section={section}
        durations={resultState.data.meta}
      />
    );
  } else {
    chartContent = (
      <FunnelsResultTable
        chartData={chartData}
        breakdown={breakdown}
        queries={queries}
        groups={groups}
        arrayMapper={arrayMapper}
        durations={resultState.data.meta}
        resultData={resultState.data}
      />
    );
  }

  let tableContent = null;

  // if (chartType === CHART_TYPE_TABLE) {
  //   tableContent = (
  //     <div
  //       onClick={handleEditQuery}
  //       style={{ color: '#5949BC' }}
  //       className='mt-3 font-medium text-base cursor-pointer flex justify-end item-center'
  //     >
  //       Show More &rarr;
  //     </div>
  //   );
  // }

  return (
    <div className={`w-full flex-1`}>
      {chartContent}
      {tableContent}
    </div>
  );
}

export default UngroupedChart;
