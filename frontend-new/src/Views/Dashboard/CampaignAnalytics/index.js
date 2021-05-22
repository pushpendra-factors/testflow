import React from "react";
import BreakdownCharts from "./BreakdownCharts";
import NonBreakdownCharts from "./NonBreakdownCharts";

function CampaignAnalytics({
  campaignState,
  chartType,
  arrayMapper,
  resultState,
  unit,
  section
}) {
  const { group_by: breakdown } = campaignState;
  if (breakdown.length) {
    return (
      <BreakdownCharts
        arrayMapper={arrayMapper}
        chartType={chartType}
        data={resultState.data}
        breakdown={breakdown}
        isWidgetModal={false}
        unit={unit}
        section={section}
      />
    );
  } else {
    return (
      <NonBreakdownCharts
        arrayMapper={arrayMapper}
        chartType={chartType}
        data={resultState.data}
        isWidgetModal={false}
        unit={unit}
        section={section}
      />
    );
  }
}

export default CampaignAnalytics;
