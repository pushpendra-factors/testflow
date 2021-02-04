import React from "react";
import ResultantChart from "./ResultantChart";

function FunnelsResultPage({
  queries,
  resultState,
  breakdown,
  arrayMapper,
  isWidgetModal,
  section
}) {
  return (
    <ResultantChart
      queries={queries}
      resultState={resultState}
      breakdown={breakdown}
      arrayMapper={arrayMapper}
      isWidgetModal={isWidgetModal}
      section={section}
    />
  );
}

export default FunnelsResultPage;
