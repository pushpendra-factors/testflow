import React from 'react';
import NoBreakdownCharts from '../NoBreakdownCharts';
import SingleEventSingleBreakdown from '../SingleEventSingleBreakdown';
import { Spin } from 'antd';
import SingleEventMultipleBreakdown from '../SingleEventMultipleBreakdown';

function TotalEvents({
  queries, eventsMapper, reverseEventsMapper, breakdown, resultState
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
      <NoBreakdownCharts
        queries={queries}
        eventsMapper={eventsMapper}
        reverseEventsMapper={reverseEventsMapper}
        resultState={resultState}
      />
    );
  } else if (queries.length === 1 && breakdown.length === 1) {
    return (
      <SingleEventSingleBreakdown
        queries={queries}
        breakdown={breakdown}
      />
    );
  } else if (queries.length === 1) {
    return (
      <SingleEventMultipleBreakdown
        queries={queries}
        breakdown={breakdown}
      />
    );
  }
}

export default TotalEvents;
