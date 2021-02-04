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
import { ACTIVE_USERS_CRITERIA, FREQUENCY_CRITERIA } from "../../../../utils/constants";

function NoBreakdownCharts({
  queries,
  resultState,
  page,
  chartType,
  isWidgetModal,
  durationObj,
  arrayMapper,
  section
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

  let chart = null;

  const lineChartData = getDataInLineChartFormat(
    chartsData,
    queries,
    hiddenEvents,
    arrayMapper
  );

  console.log(page);

  const table = (
    <div className="mt-12 w-full">
      <NoBreakdownTable
        isWidgetModal={isWidgetModal}
        data={chartsData}
        events={queries}
        chartType={chartType}
        setHiddenEvents={setHiddenEvents}
        hiddenEvents={hiddenEvents}
        durationObj={durationObj}
        arrayMapper={arrayMapper}
      />
    </div>
  );

  if (chartType === "sparklines") {
    chart = (
      <SparkLineChart
        frequency={durationObj.frequency}
        queries={queries}
        chartsData={chartsData}
        appliedColors={appliedColors}
        arrayMapper={arrayMapper}
        page={page}
        resultState={resultState}
      />
    );
  } else if (chartType === "linechart") {
    chart = (
      <LineChart
        frequency={durationObj.frequency}
        chartData={lineChartData}
        appliedColors={appliedColors}
        queries={queries}
        setHiddenEvents={setHiddenEvents}
        hiddenEvents={hiddenEvents}
        arrayMapper={arrayMapper}
        isDecimalAllowed={page === ACTIVE_USERS_CRITERIA || page === FREQUENCY_CRITERIA}
        section={section}
      />
    );
  }

  return (
    <div className="flex items-center justify-center flex-col">
      {chart}
      {table}
    </div>
  );
}

export default NoBreakdownCharts;
