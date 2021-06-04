import React from "react";
import ResultantChart from "./ResultantChart";

function FunnelsResultPage({
  queries,
  resultState,
  breakdown,
  arrayMapper,
  isWidgetModal,
  section,
  durationObj
}) {
  return (
    <ResultantChart
      queries={queries}
      resultState={resultState}
      breakdown={breakdown}
      arrayMapper={arrayMapper}
      isWidgetModal={isWidgetModal}
      section={section}
      durationObj={durationObj}
    />
  );
}

export default FunnelsResultPage;
