import React from 'react';
import GroupedChart from './GroupedChart';
import UngroupedChart from './UngroupedChart';

function ResultantChart({
  queries, setDrawerVisible, resultState, breakdown, eventsMapper, reverseEventsMapper, requestQuery, setShowResult, querySaved, setQuerySaved, isWidgetModal = false
}) {
  if (!breakdown.length) {
    return (
      <UngroupedChart
        resultState={resultState}
        queries={queries}
        setDrawerVisible={setDrawerVisible}
        eventsMapper={eventsMapper}
        requestQuery={requestQuery}
        setShowResult={setShowResult}
        querySaved={querySaved}
        setQuerySaved={setQuerySaved}
        isWidgetModal={isWidgetModal}
      />
    );
  } else {
    return (
      <GroupedChart
        queries={queries}
        setDrawerVisible={setDrawerVisible}
        resultState={resultState}
        breakdown={breakdown}
        eventsMapper={eventsMapper}
        reverseEventsMapper={reverseEventsMapper}
        requestQuery={requestQuery}
        setShowResult={setShowResult}
        querySaved={querySaved}
        setQuerySaved={setQuerySaved}
        isWidgetModal={isWidgetModal}
      />
    );
  }
}

export default ResultantChart;
