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
import {
  ACTIVE_USERS_CRITERIA,
  FREQUENCY_CRITERIA,
} from "../../../utils/constants";

function NoBreakdownCharts({
  queries,
  resultState,
  page,
  chartType,
  durationObj,
  arrayMapper,
  unit,
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

  let content = null;

  if (chartType === "sparklines") {
    content = (
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
    );
  } else if (chartType === "table") {
    content = (
      <NoBreakdownTable
        data={chartsData}
        events={queries}
        chartType={chartType}
        setHiddenEvents={setHiddenEvents}
        hiddenEvents={hiddenEvents}
        isWidgetModal={false}
        durationObj={durationObj}
        arrayMapper={arrayMapper}
      />
    );
  } else {
    const lineChartData = getDataInLineChartFormat(
      chartsData,
      queries,
      hiddenEvents,
      arrayMapper
    );
    content = (
      <LineChart
        frequency={durationObj.frequency}
        chartData={lineChartData}
        appliedColors={appliedColors}
        queries={queries}
        setHiddenEvents={setHiddenEvents}
        hiddenEvents={hiddenEvents}
        isDecimalAllowed={
          page === ACTIVE_USERS_CRITERIA || page === FREQUENCY_CRITERIA
        }
        arrayMapper={arrayMapper}
        cardSize={unit.cardSize}
        height={200}
        section={section}
      />
    );
  }

  return <div style={{ boxShadow: "inset 0px 1px 0px rgba(0, 0, 0, 0.1)" }} className="w-full px-6">{content}</div>;
}

export default NoBreakdownCharts;
