import React from 'react';
import GroupedChart from './GroupedChart';
// import FiltersInfo from './FiltersInfo';
import UngroupedChart from './UngroupedChart';
import { Spin } from 'antd';

function FunnelsResultPage({
  queries, setDrawerVisible, resultState, breakdown, eventsMapper, reverseEventsMapper, requestQuery, setShowResult, querySaved, setQuerySaved
}) {
  if (resultState.loading) {
    return (
      <div className="flex justify-center items-center w-full h-64">
        <Spin size="large" />
      </div>
    );
  }

  if (resultState.error) {
    return (
      <div className="flex justify-center items-center w-full h-64">
        Something went wrong!
      </div>
    );
  }

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

export default FunnelsResultPage;
