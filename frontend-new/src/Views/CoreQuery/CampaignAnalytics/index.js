import React from "react";
import NoBreakdownCharts from "./NoBreakdownCharts";
import BreakdownCharts from "./BreakdownCharts";

function CampaignAnalytics({
  resultState,
  arrayMapper,
  campaignState,
  chartType,
  currMetricsValue
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
      />
    );
  } else {
    content = (
      <NoBreakdownCharts
        arrayMapper={arrayMapper}
        chartType={chartType}
        data={resultState.data}
      />
    );
  }

  return <>{content}</>;
}

export default CampaignAnalytics;
