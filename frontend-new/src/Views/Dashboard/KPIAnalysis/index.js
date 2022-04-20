import React from 'react';
import NoBreakdownCharts from './NoBreakdownCharts';
import BreakdownCharts from './BreakdownCharts';

function KPIAnalysis({
  kpis,
  resultState,
  chartType,
  section,
  breakdown,
  currMetricsValue = 0,
  unit,
  arrayMapper
}) {
  if (breakdown.length) {
    return (
      <BreakdownCharts
        kpis={kpis}
        chartType={chartType}
        responseData={resultState.data}
        breakdown={breakdown}
        currentEventIndex={currMetricsValue}
        section={section}
        unit={unit}
      />
    );
  } else {
    return (
      <NoBreakdownCharts
        kpis={kpis}
        chartType={chartType}
        responseData={resultState.data}
        section={section}
        unit={unit}
        arrayMapper={arrayMapper}
      />
    );
    return null;
  }
}

export default KPIAnalysis;
