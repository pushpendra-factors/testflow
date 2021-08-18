import React from "react";
import ResultantChart from "./ResultantChart";

function FunnelsResultPage({
  queries,
  resultState,
  breakdown,
  arrayMapper,
  isWidgetModal,
  section,
  durationObj,
  renderedCompRef
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
      renderedCompRef={renderedCompRef}
    />
  );
}

export default FunnelsResultPage;
