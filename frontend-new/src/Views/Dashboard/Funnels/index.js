import React from 'react';
import UngroupedChart from './UngroupedChart';
import GroupedChart from './GroupedChart';

function Funnels({
  breakdown, resultState, events, chartType, title, eventsMapper, reverseEventsMapper
}) {
  if (!breakdown.length) {
    return (
            <UngroupedChart
                resultState={resultState}
                queries={events}
                eventsMapper={eventsMapper}
                title={title}
                chartType={chartType}
            />
    );
  } else {
    return (
            <GroupedChart
                queries={events}
                resultState={resultState}
                breakdown={breakdown}
                eventsMapper={eventsMapper}
                reverseEventsMapper={reverseEventsMapper}
                chartType={chartType}
            />
    );
  }
}

export default Funnels;
