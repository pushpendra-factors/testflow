import React from 'react';
import MultipleEventsWithBreakdown from './MultipleEventsWithBreakdown';
import SingleEventSingleBreakdown from './SingleEventSingleBreakdown';
import SingleEventMultipleBreakdown from './SingleEventMultipleBreakdown';
import NoBreakdownCharts from './NoBreakdownCharts';
import {
  TOTAL_EVENTS_CRITERIA,
  EACH_USER_TYPE,
  ANY_USER_TYPE,
  ALL_USER_TYPE
} from '../../../utils/constants';
import EventBreakdownCharts from './EventBreakdownCharts';

function EventsAnalytics({
  breakdown,
  resultState,
  events,
  chartType,
  unit,
  durationObj,
  arrayMapper,
  section,
  breakdownType
}) {
  let content = null;

  if (breakdownType === EACH_USER_TYPE) {
    if (events.length > 1 && breakdown.length) {
      content = (
        <MultipleEventsWithBreakdown
          queries={events}
          resultState={resultState}
          page={TOTAL_EVENTS_CRITERIA}
          chartType={chartType}
          durationObj={durationObj}
          section={section}
          breakdown={breakdown}
          unit={unit}
        />
      );
    }

    if (events.length === 1 && breakdown.length === 1) {
      content = (
        <SingleEventSingleBreakdown
          queries={events}
          resultState={resultState}
          page={TOTAL_EVENTS_CRITERIA}
          chartType={chartType}
          durationObj={durationObj}
          section={section}
          breakdown={breakdown}
          unit={unit}
        />
      );
    }

    if (events.length === 1 && breakdown.length > 1) {
      content = (
        <SingleEventMultipleBreakdown
          queries={events}
          resultState={resultState}
          page={TOTAL_EVENTS_CRITERIA}
          chartType={chartType}
          durationObj={durationObj}
          section={section}
          breakdown={breakdown}
          unit={unit}
        />
      );
    }

    if (!breakdown.length) {
      content = (
        <NoBreakdownCharts
          queries={events}
          resultState={resultState}
          page={TOTAL_EVENTS_CRITERIA}
          chartType={chartType}
          arrayMapper={arrayMapper}
          durationObj={durationObj}
          section={section}
          unit={unit}
        />
      );
    }
  }

  if (breakdownType === ANY_USER_TYPE || breakdownType === ALL_USER_TYPE) {
    content = (
      <EventBreakdownCharts
        section={section}
        resultState={resultState}
        breakdown={breakdown}
        chartType={chartType}
        unit={unit}
        durationObj={durationObj}
      />
    );
  }

  return <>{content}</>;
}

export default EventsAnalytics;
