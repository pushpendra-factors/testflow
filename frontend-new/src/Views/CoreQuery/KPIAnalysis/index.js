import React from 'react';
import NoBreakdownCharts from './NoBreakdownCharts';
import BreakdownCharts from './BreakdownCharts';

function KPIAnalysis({
  kpis,
  resultState,
  chartType,
  section,
  breakdown,
  currMetricsValue,
  renderedCompRef,
  durationObj,
  secondAxisKpiIndices = []
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
        renderedCompRef={renderedCompRef}
        durationObj={durationObj}
      />
    );
  }
  return (
    <NoBreakdownCharts
      kpis={kpis}
      chartType={chartType}
      responseData={resultState.data}
      section={section}
      ref={renderedCompRef}
      durationObj={durationObj}
      secondAxisKpiIndices={secondAxisKpiIndices}
      currentEventIndex={currMetricsValue}
    />
  );
}

export default KPIAnalysis;
