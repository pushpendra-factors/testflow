import React from 'react';
import NoBreakdownCharts from './NoBreakdownCharts';
import BreakdownCharts from './BreakdownCharts';

function CampaignAnalytics({
  resultState,
  arrayMapper,
  campaignState,
  chartType,
  currMetricsValue,
  section,
  durationObj,
}) {
  const { group_by: breakdown } = campaignState;

  let content = null;

  if (breakdown.length) {
    content = (
      <BreakdownCharts
        arrayMapper={arrayMapper}
        chartType={chartType}
        data={resultState.data}
        breakdown={breakdown}
        currentEventIndex={currMetricsValue}
        section={section}
      />
    );
  } else {
    content = (
      <NoBreakdownCharts
        arrayMapper={arrayMapper}
        chartType={chartType}
        data={resultState.data}
        section={section}
        durationObj={durationObj}
      />
    );
  }

  return <>{content}</>;
}

export default CampaignAnalytics;
