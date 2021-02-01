import React, { useState, useEffect } from "react";
import { formatData, formatDataInLineChartFormat } from "./utils";
import BarChart from "../../../../components/BarChart";
import LineChart from "../../../../components/LineChart";
import SingleEventMultipleBreakdownTable from "./SingleEventMultipleBreakdownTable";
import { generateColors } from "../../../../utils/dataFormatter";

function SingleEventMultipleBreakdown({
  queries,
  breakdown,
  resultState,
  page,
  chartType,
  isWidgetModal,
  durationObj,
  title,
}) {
  const [chartsData, setChartsData] = useState([]);
  const [visibleProperties, setVisibleProperties] = useState([]);
  const [hiddenProperties, setHiddenProperties] = useState([]);

  const maxAllowedVisibleProperties = 5;

  useEffect(() => {
    const formattedData = formatData(resultState.data);
    setChartsData(formattedData);
    setVisibleProperties([
      ...formattedData.slice(0, maxAllowedVisibleProperties),
    ]);
  }, [resultState.data]);

  if (!chartsData.length) {
    return null;
  }

  const mapper = {};
  const reverseMapper = {};
  const arrayMapper = [];

  const visibleLabels = visibleProperties.map((v) => v.label);

  visibleLabels.forEach((q, index) => {
    mapper[`${q}`] = `event${index + 1}`;
    reverseMapper[`event${index + 1}`] = q;
    arrayMapper.push({
      eventName: q,
      index,
      mapper: `event${index + 1}`,
    });
  });

  const lineChartData = formatDataInLineChartFormat(
    resultState.data,
    visibleProperties,
    mapper,
    hiddenProperties
  );

  const appliedColors = generateColors(visibleProperties.length);

  let chart = null;
  const table = (
    <div className="mt-12 w-full">
      <SingleEventMultipleBreakdownTable
        isWidgetModal={isWidgetModal}
        data={chartsData}
        lineChartData={lineChartData}
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
  );

  if (chartType === "barchart") {
    chart = <BarChart title={title} chartData={visibleProperties} />;
  } else {
    chart = (
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
        arrayMapper={arrayMapper}
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

export default SingleEventMultipleBreakdown;
