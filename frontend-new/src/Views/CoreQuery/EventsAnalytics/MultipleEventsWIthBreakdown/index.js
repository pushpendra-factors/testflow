import React, { useState, useEffect } from "react";
import {
  formatData,
  formatVisibleProperties,
  formatDataInLineChartFormat,
} from "./utils";
import BarChart from "../../../../components/BarChart";
import MultipleEventsWithBreakdownTable from "./MultipleEventsWithBreakdownTable";
import LineChart from "../../../../components/LineChart";
import { generateColors } from "../../../../utils/dataFormatter";

function MultipleEventsWithBreakdown({
  queries,
  breakdown,
  resultState,
  page,
  chartType,
  isWidgetModal,
  durationObj,
}) {
  const [chartsData, setChartsData] = useState([]);
  const [visibleProperties, setVisibleProperties] = useState([]);
  const [hiddenProperties, setHiddenProperties] = useState([]);

  const maxAllowedVisibleProperties = 5;

  useEffect(() => {
    const appliedColors = generateColors(queries.length);
    const formattedData = formatData(resultState.data, queries, appliedColors);
    setChartsData(formattedData);
    setVisibleProperties([
      ...formattedData.slice(0, maxAllowedVisibleProperties),
    ]);
  }, [resultState.data, queries]);

  if (!chartsData.length) {
    return null;
  }

  const mapper = {};
  const reverseMapper = {};

  const visibleLabels = visibleProperties.map((v) => `${v.event},${v.label}`);

  visibleLabels.forEach((q, index) => {
    mapper[`${q}`] = `event${index + 1}`;
    reverseMapper[`event${index + 1}`] = q;
  });

  let chartContent = null;

  const lineChartData = formatDataInLineChartFormat(
    visibleProperties,
    mapper,
    hiddenProperties,
    durationObj.frequency
  );
  const appliedColors = generateColors(visibleProperties.length);

  if (chartType === "barchart") {
    chartContent = (
      <div className="flex mt-8">
        <BarChart
          chartData={formatVisibleProperties(visibleProperties, queries)}
          queries={queries}
        />
      </div>
    );
  } else {
    chartContent = (
      <div className="flex mt-8">
        <LineChart
          frequency={durationObj.frequency}
          chartData={lineChartData}
          appliedColors={appliedColors}
          queries={visibleLabels}
          reverseEventsMapper={reverseMapper}
          eventsMapper={mapper}
          setHiddenEvents={setHiddenProperties}
          hiddenEvents={hiddenProperties}
          isDecimalAllowed={page === "activeUsers" || page === "frequency"}
        />
      </div>
    );
  }

  return (
    <div className="total-events w-full">
      {chartContent}
      <div className="mt-8">
        <MultipleEventsWithBreakdownTable
          isWidgetModal={isWidgetModal}
          data={chartsData}
          lineChartData={lineChartData}
          queries={queries}
          breakdown={breakdown}
          events={queries}
          chartType={chartType}
          setVisibleProperties={setVisibleProperties}
          visibleProperties={visibleProperties}
          maxAllowedVisibleProperties={maxAllowedVisibleProperties}
          originalData={resultState.data}
          page={page}
          durationObj={durationObj}
        />
      </div>
    </div>
  );
}

export default MultipleEventsWithBreakdown;
