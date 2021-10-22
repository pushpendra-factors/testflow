import React from 'react';
import NoBreakdownCharts from './NoBreakdownCharts';
import BreakdownCharts from './BreakdownCharts';

function ProfilesResultPage({
  queries,
  resultState,
  chartType,
  section,
  breakdown,
  currMetricsValue,
  renderedCompRef,
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
        ref={renderedCompRef}
      />
    );
  } else {
    return (
      <NoBreakdownCharts
        queries={queries}
        chartType={chartType}
        data={resultState.data}
        section={section}
        ref={renderedCompRef}
      />
    );
  }
}

export default ProfilesResultPage;
