import React from "react";
import ResultantChart from "./ResultantChart";

function FunnelsResultPage({
  queries,
  setDrawerVisible,
  resultState,
  breakdown,
  requestQuery,
  setShowResult,
  querySaved,
  setQuerySaved,
  arrayMapper,
}) {
  return (
    <ResultantChart
      queries={queries}
      setDrawerVisible={setDrawerVisible}
      resultState={resultState}
      breakdown={breakdown}
      requestQuery={requestQuery}
      setShowResult={setShowResult}
      querySaved={querySaved}
      setQuerySaved={setQuerySaved}
      arrayMapper={arrayMapper}
    />
  );
}

export default FunnelsResultPage;
