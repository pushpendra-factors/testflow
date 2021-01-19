import React from 'react';
import GroupedChart from './GroupedChart';
import UngroupedChart from './UngroupedChart';

function ResultantChart({
  queries, setDrawerVisible, resultState, breakdown, requestQuery, setShowResult, querySaved, setQuerySaved, isWidgetModal = false, arrayMapper
}) {
  if (!breakdown.length) {
    return (
      <UngroupedChart
        resultState={resultState}
        queries={queries}
        setDrawerVisible={setDrawerVisible}
        requestQuery={requestQuery}
        setShowResult={setShowResult}
        querySaved={querySaved}
        setQuerySaved={setQuerySaved}
        isWidgetModal={isWidgetModal}
        arrayMapper={arrayMapper}
      />
    );
  } else {
    return (
      <GroupedChart
        queries={queries}
        setDrawerVisible={setDrawerVisible}
        resultState={resultState}
        breakdown={breakdown}
        requestQuery={requestQuery}
        setShowResult={setShowResult}
        querySaved={querySaved}
        setQuerySaved={setQuerySaved}
        isWidgetModal={isWidgetModal}
        arrayMapper={arrayMapper}
      />
    );
  }
}

export default ResultantChart;
