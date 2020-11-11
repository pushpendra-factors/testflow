import React from 'react';
import MultipleEventsWithBreakdown from './MultipleEventsWithBreakdown';
import SingleEventSingleBreakdown from './SingleEventSingleBreakdown';
import SingleEventMultipleBreakdown from './SingleEventMultipleBreakdown';
import NoBreakdownCharts from './NoBreakdownCharts';
import { useSelector } from 'react-redux';

function EventsAnalytics({
  breakdown, resultState, events, chartType, title, eventsMapper, reverseEventsMapper
}) {
  let content = null;

  const { dashboards_loaded } = useSelector(state => state.dashboard)

  if (events.length > 1 && breakdown.length) {
    content = (
            <MultipleEventsWithBreakdown
                breakdownType="each"
                queries={events}
                breakdown={breakdown}
                resultState={resultState}
                page="totalEvents"
                chartType={chartType}
                title={title}
                dashboards_loaded={dashboards_loaded}
            />
    );
  }

  if (events.length === 1 && breakdown.length === 1) {
    content = (
            <SingleEventSingleBreakdown
                breakdownType="each"
                queries={events}
                breakdown={breakdown}
                resultState={resultState}
                page="totalEvents"
                chartType={chartType}
                title={title}
                dashboards_loaded={dashboards_loaded}
            />
    );
  }

  if (events.length === 1 && breakdown.length > 1) {
    content = (
            <SingleEventMultipleBreakdown
                breakdownType="each"
                queries={events}
                breakdown={breakdown}
                resultState={resultState}
                page="totalEvents"
                chartType={chartType}
                title={title}
            />
    );
  }

  if (!breakdown.length) {
    content = (
            <NoBreakdownCharts
                queries={events}
                eventsMapper={eventsMapper}
                reverseEventsMapper={reverseEventsMapper}
                resultState={resultState}
                page="totalEvents"
                chartType={chartType}
                title={title}
            />
    );
  }

  return (
        <div className="card-content">
            {content}
        </div>
  );
}

export default EventsAnalytics;
