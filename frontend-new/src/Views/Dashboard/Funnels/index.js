import React from 'react';
import UngroupedChart from './UngroupedChart';
import GroupedChart from './GroupedChart';
import { useSelector } from 'react-redux';

function Funnels({
  breakdown, resultState, events, chartType, title, eventsMapper, reverseEventsMapper
}) {
  const { dashboards_loaded } = useSelector(state => state.dashboard);

  if (!breakdown.length) {
    return (
      <UngroupedChart
        resultState={resultState}
        queries={events}
        eventsMapper={eventsMapper}
        title={title}
        chartType={chartType}
        dashboards_loaded={dashboards_loaded}
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
        dashboards_loaded={dashboards_loaded}
      />
    );
  }
}

export default Funnels;
