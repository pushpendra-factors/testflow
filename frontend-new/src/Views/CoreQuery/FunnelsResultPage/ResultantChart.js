import React from 'react';
import GroupedChart from './GroupedChart';
import UngroupedChart from './UngroupedChart';

function ResultantChart({
  queries,
  resultState,
  breakdown,
  isWidgetModal,
  arrayMapper,
  section,
  durationObj,
  chartType,
  renderedCompRef,
}) {
  if (!breakdown.length) {
    return (
      <UngroupedChart
        resultState={resultState}
        queries={queries}
        isWidgetModal={isWidgetModal}
        arrayMapper={arrayMapper}
        section={section}
        durationObj={durationObj}
        ref={renderedCompRef}
      />
    );
  } else {
    return (
      <GroupedChart
        queries={queries}
        resultState={resultState}
        breakdown={breakdown}
        isWidgetModal={isWidgetModal}
        arrayMapper={arrayMapper}
        section={section}
        ref={renderedCompRef}
        chartType={chartType}
      />
    );
  }
}

export default ResultantChart;
