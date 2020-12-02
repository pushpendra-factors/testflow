import React from 'react';
import MultipleEventsWithBreakdown from './MultipleEventsWithBreakdown';
import SingleEventSingleBreakdown from './SingleEventSingleBreakdown';
import SingleEventMultipleBreakdown from './SingleEventMultipleBreakdown';
import NoBreakdownCharts from './NoBreakdownCharts';

function EventsAnalytics({
  breakdown, resultState, events, chartType, title, eventsMapper, reverseEventsMapper, unit, setwidgetModal
}) {
  let content = null;

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
        unit={unit}
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
        unit={unit}
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
        unit={unit}
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

  let tableContent = null;

  if (chartType === 'table') {
    tableContent = (
      <div onClick={() => setwidgetModal({ unit, data: resultState.data })} style={{ color: '#5949BC' }} className="mt-3 font-medium text-base cursor-pointer flex justify-end item-center">Show More &rarr;</div>
    )
  }

  return (
    <div className="card-content">
      {content}
      {tableContent}
    </div>
  );
}

export default EventsAnalytics;
