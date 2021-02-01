import React from "react";
import NoBreakdownCharts from "../NoBreakdownCharts";
import SingleEventSingleBreakdown from "../SingleEventSingleBreakdown";
import SingleEventMultipleBreakdown from "../SingleEventMultipleBreakdown";
import MultipleEventsWithBreakdown from "../MultipleEventsWIthBreakdown";

function ResultTab({
  queries,
  breakdown,
  resultState,
  page,
  isWidgetModal = false,
  arrayMapper,
  title = "chart",
  chartType,
  durationObj,
}) {
  let content = null;

  if (resultState.data && !resultState.data.metrics.rows.length) {
    content = (
      <div className="h-64 flex items-center justify-center w-full">
        No Data Found!
      </div>
    );
  }

  if (resultState.data && resultState.data.metrics.rows.length) {
    if (!breakdown.length) {
      content = (
        <NoBreakdownCharts
          queries={queries}
          resultState={resultState}
          page={page}
          chartType={chartType}
          isWidgetModal={isWidgetModal}
          arrayMapper={arrayMapper}
          durationObj={durationObj}
        />
      );
    }

    if (queries.length === 1 && breakdown.length === 1) {
      content = (
        <SingleEventSingleBreakdown
          queries={queries}
          breakdown={breakdown}
          resultState={resultState}
          page={page}
          chartType={chartType}
          isWidgetModal={isWidgetModal}
          title={title}
          durationObj={durationObj}
        />
      );
    }

    if (queries.length > 1 && breakdown.length) {
      content = (
        <MultipleEventsWithBreakdown
          queries={queries}
          breakdown={breakdown}
          resultState={resultState}
          page={page}
          chartType={chartType}
          isWidgetModal={isWidgetModal}
          title={title}
          durationObj={durationObj}
        />
      );
    }

    if (queries.length === 1 && breakdown.length > 1) {
      content = (
        <SingleEventMultipleBreakdown
          queries={queries}
          breakdown={breakdown}
          resultState={resultState}
          page={page}
          chartType={chartType}
          isWidgetModal={isWidgetModal}
          title={title}
          durationObj={durationObj}
        />
      );
    }
  }

  return <>{content}</>;
}

export default ResultTab;
