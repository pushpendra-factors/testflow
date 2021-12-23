import React from 'react';
import NoBreakdownCharts from './NoBreakdownCharts';
import BreakdownCharts from './BreakdownCharts';

function KPIAnalysis({
  queries,
  resultState,
  chartType,
  section,
  breakdown,
  currMetricsValue,
  renderedCompRef,
  durationObj,
}) {
  if (breakdown.length) {
    return (
      <BreakdownCharts
        queries={queries}
        chartType={chartType}
        responseData={resultState.data}
        breakdown={breakdown}
        currentEventIndex={currMetricsValue}
        section={section}
        ref={renderedCompRef}
        durationObj={durationObj}
      />
    );
  } else {
    return (
      <NoBreakdownCharts
        queries={queries}
        chartType={chartType}
        responseData={resultState.data}
        section={section}
        ref={renderedCompRef}
        durationObj={durationObj}
      />
    );
  }
}

export default KPIAnalysis;
