import React from 'react';
import Header from '../../AppLayout/Header';
import ResultsHeader from '../ResultsHeader';
import EventsInfo from './EventsInfo';
import { Spin } from 'antd';
import ResultantChart from './ResultantChart';
import FiltersInfo from '../FiltersInfo';

function FunnelsResultPage({
  queries, setDrawerVisible, resultState, breakdown, eventsMapper, reverseEventsMapper, requestQuery, setShowResult, querySaved, setQuerySaved, handleDurationChange, durationObj
}) {
  let content = null;

  if (resultState.loading) {
    content = (
      <div className="mt-40 flex justify-center items-center w-full h-64">
        <Spin size="large" />
      </div>
    );
  }

  if (resultState.error) {
    content = (
      <div className="mt-40 flex justify-center items-center w-full h-64">
        Something went wrong!
      </div>
    );
  }

  if (resultState.data) {
    content = (
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
        <EventsInfo setDrawerVisible={setDrawerVisible} queries={queries} />
        <FiltersInfo
          durationObj={durationObj}
          handleDurationChange={handleDurationChange}
          breakdown={breakdown}
        />
      </Header>
      {content}
    </>
  );
}

export default FunnelsResultPage;
