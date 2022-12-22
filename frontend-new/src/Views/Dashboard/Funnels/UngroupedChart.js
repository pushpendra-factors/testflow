import React, { useMemo } from 'react';
import { get } from 'lodash';
import cx from 'classnames';
import { generateUngroupedChartsData } from '../../CoreQuery/FunnelsResultPage/utils';
import Chart from '../../CoreQuery/FunnelsResultPage/UngroupedChart/Chart';
import FunnelsResultTable from '../../CoreQuery/FunnelsResultPage/FunnelsResultTable';
import {
  CHART_TYPE_BARCHART,
  CHART_TYPE_METRIC_CHART,
  CHART_TYPE_TABLE,
  DASHBOARD_WIDGET_UNGROUPED_FUNNEL_CHART_HEIGHT
} from '../../../utils/constants';
import MetricChart from 'Components/MetricChart/MetricChart';
import { CONVERSION_RATE_LABEL } from 'Views/CoreQuery/FunnelsResultPage/UngroupedChart/ungroupedChart.constants';

function UngroupedChart({
  resultState,
  queries,
  chartType,
  unit,
  arrayMapper,
  section,
  breakdown
}) {
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
  }
  if (chartType === CHART_TYPE_METRIC_CHART) {
    chartContent = (
      <MetricChart
        headerTitle={CONVERSION_RATE_LABEL}
        valueType={'percentage'}
        value={get(chartData, `${chartData.length - 1}.value`, 0)}
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
