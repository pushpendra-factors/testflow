import React from 'react';
import ResultantChart from './ResultantChart';

function FunnelsResultPage({
  queries,
  resultState,
  breakdown,
  arrayMapper,
  isWidgetModal,
  section,
  durationObj,
  renderedCompRef,
  chartType,
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
      chartType={chartType}
    />
  );
}

export default FunnelsResultPage;
