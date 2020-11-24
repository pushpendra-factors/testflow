import React from 'react';
import UngroupedChart from './UngroupedChart';
import GroupedChart from './GroupedChart';
import { useSelector } from 'react-redux';

function Funnels({
  breakdown, resultState, events, chartType, title, eventsMapper, reverseEventsMapper, unit
}) {
  const { dashboardsLoaded } = useSelector(state => state.dashboard);

  if (!breakdown.length) {
    return (
      <UngroupedChart
        resultState={resultState}
        queries={events}
        eventsMapper={eventsMapper}
        title={title}
        chartType={chartType}
        dashboardsLoaded={dashboardsLoaded}
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
        dashboardsLoaded={dashboardsLoaded}
        unit={unit}
      />
    );
  }
}

export default Funnels;
