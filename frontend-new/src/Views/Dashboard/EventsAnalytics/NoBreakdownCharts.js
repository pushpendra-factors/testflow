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
  CHART_TYPE_TABLE,
} from "../../../utils/constants";

function NoBreakdownCharts({
  queries,
  resultState,
  page,
  chartType,
  durationObj,
  arrayMapper,
  unit,
  section,
  setwidgetModal,
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

  let tableContent = null;

  if (chartType === CHART_TYPE_TABLE) {
    tableContent = (
      <div
        onClick={() => setwidgetModal({ unit, data: resultState.data })}
        style={{ color: "#5949BC" }}
        className="mt-3 font-medium text-base cursor-pointer flex justify-end item-center"
      >
        Show More &rarr;
      </div>
    );
  }

  if (chartType === "sparklines") {
    chartContent = (
      <SparkLineChart
        frequency={durationObj.frequency}
        queries={queries}
        chartsData={chartsData}
        appliedColors={appliedColors}
        arrayMapper={arrayMapper}
        page={page}
        resultState={resultState}
        cardSize={unit.cardSize}
        height={queries.length === 1 && unit.cardSize ? 180 : 100}
      />
    );
  } else if (chartType === "table") {
    chartContent = (
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
    chartContent = (
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
        height={225}
        section={section}
      />
    );
  }

  return (
    <div
      className={`w-full px-6 flex flex-1 flex-col  justify-center`}
    >
      {chartContent}
      {tableContent}
    </div>
  );
}

export default NoBreakdownCharts;
