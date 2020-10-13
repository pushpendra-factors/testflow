import React from 'react';
import NoBreakdownCharts from '../NoBreakdownCharts';
import SingleEventSingleBreakdown from '../SingleEventSingleBreakdown';
import { Spin } from 'antd';
import SingleEventMultipleBreakdown from '../SingleEventMultipleBreakdown';

function TotalEvents({
  queries, eventsMapper, reverseEventsMapper, breakdown, resultState, page, index
}) {
  if (resultState[index].loading) {
    return (
      <div className="flex justify-center items-center w-full h-64">
        <Spin size="large" />
      </div>
    );
  }

  if (resultState[index].error) {
    return (
      <div className="flex justify-center items-center w-full h-64">
        Something went wrong!
      </div>
    );
  }

  if (!resultState[index].data) {
    return null;
  }

  if (!breakdown.length) {
    return (
      <NoBreakdownCharts
        queries={queries}
        eventsMapper={eventsMapper}
        reverseEventsMapper={reverseEventsMapper}
        resultState={resultState[index]}
        page={page}
      />
    );
  } else if (queries.length === 1 && breakdown.length === 1) {
    return (
      <SingleEventSingleBreakdown
        queries={queries}
        breakdown={breakdown}
        resultState={resultState[index]}
        page={page}
      />
    );
  } else if (queries.length === 1) {
    return (
      <SingleEventMultipleBreakdown
        queries={queries}
        breakdown={breakdown}
        resultState={resultState[index]}
        page={page}
      />
    );
  }
}

export default TotalEvents;
