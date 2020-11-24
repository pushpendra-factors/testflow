import React from 'react';
import GroupedChart from './GroupedChart';
import UngroupedChart from './UngroupedChart';

function ResultantChart({
  queries, setDrawerVisible, resultState, breakdown, eventsMapper, reverseEventsMapper, requestQuery, setShowResult, querySaved, setQuerySaved, modal = false
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
            />
    );
  } else {
    return (
            <GroupedChart
                modal={modal}
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
            />
    );
  }
}

export default ResultantChart;
