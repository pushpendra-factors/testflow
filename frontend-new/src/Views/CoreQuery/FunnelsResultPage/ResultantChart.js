import React from 'react';
import GroupedChart from './GroupedChart';
import UngroupedChart from './UngroupedChart';

function ResultantChart({
  queries, resultState, breakdown, isWidgetModal, arrayMapper
}) {
  if (!breakdown.length) {
    return (
      <UngroupedChart
        resultState={resultState}
        queries={queries}
        isWidgetModal={isWidgetModal}
        arrayMapper={arrayMapper}
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
      />
    );
  }
}

export default ResultantChart;
