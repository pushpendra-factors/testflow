import React from 'react';
import NoBreakdownCharts from '../NoBreakdownCharts';
import BreakdownCharts from '../BreakdownCharts';

function TotalEvents({ queries, eventsMapper, reverseEventsMapper, breakdown, resultState }) {

  if(resultState.loading) {
    return 'Loading....';
  }

  if(resultState.error) {
    return 'Something went wrong!';
  }

  if (!breakdown.length) {
    return (
      <NoBreakdownCharts
        queries={queries}
        eventsMapper={eventsMapper}
        reverseEventsMapper={reverseEventsMapper}
        resultState={resultState}
      />
    )
  } else {
    return (
      <BreakdownCharts
        queries={queries}
        eventsMapper={eventsMapper}
        reverseEventsMapper={reverseEventsMapper}
        breakdown={breakdown}
        resultState={resultState}
      />
    )
  }
}

export default TotalEvents;
