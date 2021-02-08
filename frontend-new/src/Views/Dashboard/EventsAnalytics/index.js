import React from "react";
import MultipleEventsWithBreakdown from "./MultipleEventsWithBreakdown";
import SingleEventSingleBreakdown from "./SingleEventSingleBreakdown";
import SingleEventMultipleBreakdown from "./SingleEventMultipleBreakdown";
import NoBreakdownCharts from "./NoBreakdownCharts";
import {
  TOTAL_EVENTS_CRITERIA,
  EACH_USER_TYPE,
} from "../../../utils/constants";

function EventsAnalytics({
  breakdown,
  resultState,
  events,
  chartType,
  title,
  unit,
  setwidgetModal,
  durationObj,
  arrayMapper,
  section,
}) {
  let content = null;

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
        setwidgetModal={setwidgetModal}
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
        setwidgetModal={setwidgetModal}
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
        setwidgetModal={setwidgetModal}
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
        setwidgetModal={setwidgetModal}
      />
    );
  }

  return <>{content}</>;
}

export default EventsAnalytics;
