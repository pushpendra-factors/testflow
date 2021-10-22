import React from 'react';
import NoBreakdownCharts from './NoBreakdownCharts';
import BreakdownCharts from './BreakdownCharts';

function ProfileAnalysis({
  queries,
  resultState,
  chartType,
  section,
  breakdown,
  currMetricsValue = 0,
  unit,
}) {
  if (breakdown.length) {
    return (
      <BreakdownCharts
        queries={queries}
        chartType={chartType}
        data={resultState.data}
        breakdown={breakdown}
        currentEventIndex={currMetricsValue}
        section={section}
        unit={unit}
      />
    );
  } else {
    return (
      <NoBreakdownCharts
        queries={queries}
        chartType={chartType}
        data={resultState.data}
        section={section}
        unit={unit}
      />
    );
  }
}

export default ProfileAnalysis;
