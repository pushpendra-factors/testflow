import React from "react";
import UngroupedChart from "./UngroupedChart";
import GroupedChart from "./GroupedChart";

function Funnels({
  breakdown,
  resultState,
  events,
  chartType,
  title,
  unit,
  setwidgetModal,
  arrayMapper,
  section,
}) {
  try {
    if (!breakdown.length) {
      return (
        <UngroupedChart
          resultState={resultState}
          queries={events}
          title={title}
          chartType={chartType}
          setwidgetModal={setwidgetModal}
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
          title={title}
          setwidgetModal={setwidgetModal}
          arrayMapper={arrayMapper}
          section={section}
        />
      );
    }
  } catch (err) {
    console.log("src/Views/Dashboard/Funnels/index.js", err);
    return null;
  }
}

export default Funnels;
