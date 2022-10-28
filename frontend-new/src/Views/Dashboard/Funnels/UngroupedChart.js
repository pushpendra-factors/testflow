import React, { useContext, useMemo } from 'react';
import cx from 'classnames';
import { generateUngroupedChartsData } from '../../CoreQuery/FunnelsResultPage/utils';
import Chart from '../../CoreQuery/FunnelsResultPage/UngroupedChart/Chart';
import FunnelsResultTable from '../../CoreQuery/FunnelsResultPage/FunnelsResultTable';
import {
  CHART_TYPE_BARCHART,
  CHART_TYPE_TABLE,
  DASHBOARD_WIDGET_UNGROUPED_FUNNEL_CHART_HEIGHT
} from '../../../utils/constants';
import { DashboardContext } from '../../../contexts/DashboardContext';

function UngroupedChart({
  resultState,
  queries,
  chartType,
  unit,
  arrayMapper,
  section,
  breakdown
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

  return (
    <div
      className={cx('w-full flex-1', {
        'p-2 flex items-center': chartType !== CHART_TYPE_TABLE,
        'overflow-scroll': chartType === CHART_TYPE_TABLE
      })}
    >
      {chartContent}
    </div>
  );
}

export default UngroupedChart;
