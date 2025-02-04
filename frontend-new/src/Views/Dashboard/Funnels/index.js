import React from "react";
import UngroupedChart from "./UngroupedChart";
import GroupedChart from "./GroupedChart";

function Funnels({
  breakdown,
  resultState,
  events,
  chartType,
  unit,
  arrayMapper,
  section,
}) {
  if (!breakdown.length) {
    return (
      <UngroupedChart
        resultState={resultState}
        breakdown={breakdown}
        queries={events}
        chartType={chartType}
        unit={unit}
        arrayMapper={arrayMapper}
        section={section}
      />
    );
  } else {
    return (
      <GroupedChart
        queries={events}
        resultState={resultState}
        breakdown={breakdown}
        chartType={chartType}
        unit={unit}
        arrayMapper={arrayMapper}
        section={section}
      />
    );
  }
}

export default Funnels;
