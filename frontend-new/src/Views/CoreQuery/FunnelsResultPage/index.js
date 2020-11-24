import React from 'react';
// import GroupedChart from './GroupedChart';
// import FiltersInfo from './FiltersInfo';
// import UngroupedChart from './UngroupedChart';
import Header from '../../AppLayout/Header';
import ResultsHeader from '../ResultsHeader';
import EventsInfo from './EventsInfo';
import { Spin } from 'antd';
import ResultantChart from './ResultantChart';

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

  return (
    <>
      <Header>
        <ResultsHeader
          setShowResult={setShowResult}
          requestQuery={requestQuery}
          querySaved={querySaved}
          setQuerySaved={setQuerySaved}
          queryType="funnel"
        />
        <div className="py-4">
          <EventsInfo setDrawerVisible={setDrawerVisible} queries={queries} />
        </div>
      </Header>
      <div className="mt-40 mb-8 fa-container">
        <ResultantChart
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
      </div>
    </>
  );
}

export default FunnelsResultPage;
