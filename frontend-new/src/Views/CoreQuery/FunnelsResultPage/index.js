import React from "react";
import ResultantChart from "./ResultantChart";

function FunnelsResultPage({
  queries,
  resultState,
  breakdown,
  arrayMapper,
  isWidgetModal
}) {
  return (
    <ResultantChart
      queries={queries}
      resultState={resultState}
      breakdown={breakdown}
      arrayMapper={arrayMapper}
      isWidgetModal={isWidgetModal}
    />
  );
}

export default FunnelsResultPage;
