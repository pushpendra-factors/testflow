import React, { useState } from "react";
import {
  formatSingleEventAnalyticsData,
  formatMultiEventsAnalyticsData,
  getDataInLineChartFormat,
} from "../../CoreQuery/EventsAnalytics/NoBreakdownCharts/utils";
import NoBreakdownTable from "../../CoreQuery/EventsAnalytics/NoBreakdownCharts/NoBreakdownTable";
import SparkLineChart from "../../../components/SparkLineChart";
import LineChart from "../../../components/LineChart";
import { generateColors } from "../../../utils/dataFormatter";

function NoBreakdownCharts({
  queries,
  eventsMapper,
  reverseEventsMapper,
  resultState,
  page,
  chartType,
  durationObj,
  arrayMapper,
  unit,
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

  if (chartType === "sparklines") {
    chartContent = (
      <div className="mt-4">
        <SparkLineChart
          frequency={durationObj.frequency}
          queries={queries}
          chartsData={chartsData}
          parentClass={`flex items-center flex-wrap justify-center ${
            !unit.cardSize ? "mt-8 flex-col" : "mt-4"
          }`}
          appliedColors={appliedColors}
          arrayMapper={arrayMapper}
          page={page}
          resultState={resultState}
        />
      </div>
    );
  } else if (chartType === "table") {
    chartContent = (
      <div className="mt-4">
        <NoBreakdownTable
          data={chartsData}
          events={queries}
          reverseEventsMapper={reverseEventsMapper}
          chartType={chartType}
          setHiddenEvents={setHiddenEvents}
          hiddenEvents={hiddenEvents}
          durationObj={durationObj}
        />
      </div>
    );
  } else {
    const lineChartData = getDataInLineChartFormat(
      chartsData,
      queries,
      eventsMapper,
      hiddenEvents,
      arrayMapper
    );
    chartContent = (
      <div className="flex mt-4">
        <LineChart
          frequency={durationObj.frequency}
          chartData={lineChartData}
          appliedColors={appliedColors}
          queries={queries}
          reverseEventsMapper={reverseEventsMapper}
          eventsMapper={eventsMapper}
          setHiddenEvents={setHiddenEvents}
          hiddenEvents={hiddenEvents}
          isDecimalAllowed={page === "activeUsers" || page === "frequency"}
          arrayMapper={arrayMapper}
          cardSize={unit.cardSize}
        />
      </div>
    );
  }

  return <div className="total-events w-full">{chartContent}</div>;
}

export default NoBreakdownCharts;
