import React from 'react';
import NoBreakdownCharts from '../NoBreakdownCharts';
import BreakdownCharts from '../BreakdownCharts';
import { Spin } from 'antd';

function TotalEvents({ queries, eventsMapper, reverseEventsMapper, breakdown, resultState }) {

  if (resultState.loading) {
    return (
      <div className="flex justify-center items-center w-full h-64">
        <Spin size="large" />
      </div>
    )
  }

  if (resultState.error) {
    return (
      <div className="flex justify-center items-center w-full h-64">
        Something went wrong!
      </div>
    )
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
  } else {
    return (
      <BreakdownCharts
        queries={queries}
        eventsMapper={eventsMapper}
        reverseEventsMapper={reverseEventsMapper}
        breakdown={breakdown}
        resultState={resultState}
      />
    );
  }
}

export default TotalEvents;
