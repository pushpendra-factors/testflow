import React from 'react';
import NoBreakdownCharts from '../NoBreakdownCharts';
import BreakdownCharts from '../BreakdownCharts';

function TotalEvents({
  queries, eventsMapper, reverseEventsMapper, breakdown
}) {
  if (!breakdown.length) {
    return (
      <NoBreakdownCharts
        queries={queries}
        eventsMapper={eventsMapper}
        reverseEventsMapper={reverseEventsMapper}
      />
    );
  } else {
    return (
      <BreakdownCharts
        queries={queries}
        eventsMapper={eventsMapper}
        reverseEventsMapper={reverseEventsMapper}
        breakdown={breakdown}
      />
    );
  }
}

export default TotalEvents;
