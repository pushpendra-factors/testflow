import React, { useState } from "react";
import {
  formatSingleEventAnalyticsData,
  formatMultiEventsAnalyticsData,
  getDataInLineChartFormat,
} from "./utils";
import NoBreakdownTable from "./NoBreakdownTable";
import SparkLineChart from "../../../../components/SparkLineChart";
import LineChart from "../../../../components/LineChart";
import { generateColors } from "../../../../utils/dataFormatter";

function NoBreakdownCharts({
  queries,
  eventsMapper,
  reverseEventsMapper,
  resultState,
  page,
  chartType,
  isWidgetModal,
  durationObj,
  arrayMapper,
}) {
  const [hiddenEvents, setHiddenEvents] = useState([]);
  const appliedColors = generateColors(queries.length);

  let chartsData = [];
  if (queries.length === 1) {
    chartsData = formatSingleEventAnalyticsData(resultState.data, arrayMapper);
  } else {
    chartsData = formatMultiEventsAnalyticsData(resultState.data, arrayMapper);
  }

  if (!chartsData.length) {
    return null;
  }

  let chartContent = null;

  const lineChartData = getDataInLineChartFormat(
    chartsData,
    queries,
    eventsMapper,
    hiddenEvents,
    arrayMapper
  );

  if (chartType === "sparklines") {
    chartContent = (
      <SparkLineChart
        frequency={durationObj.frequency}
        queries={queries}
        chartsData={chartsData}
        parentClass="flex items-center flex-wrap mt-4 justify-center"
        appliedColors={appliedColors}
        eventsMapper={eventsMapper}
        page={page}
        resultState={resultState}
      />
    );
  } else if (chartType === "linechart") {
    chartContent = (
      <div className="flex mt-8">
        <LineChart
          frequency={durationObj.frequency}
          chartData={lineChartData}
          appliedColors={appliedColors}
          queries={queries}
          reverseEventsMapper={reverseEventsMapper}
          eventsMapper={eventsMapper}
          setHiddenEvents={setHiddenEvents}
          hiddenEvents={hiddenEvents}
          arrayMapper={arrayMapper}
          isDecimalAllowed={page === "activeUsers" || page === "frequency"}
        />
      </div>
    );
  }

  return (
    <div className="w-full">
      {chartContent}
      <div className="mt-8">
        <NoBreakdownTable
          isWidgetModal={isWidgetModal}
          data={chartsData}
          events={queries}
          reverseEventsMapper={reverseEventsMapper}
          chartType={chartType}
          setHiddenEvents={setHiddenEvents}
          hiddenEvents={hiddenEvents}
          durationObj={durationObj}
        />
      </div>
    </div>
  );
}

export default NoBreakdownCharts;
