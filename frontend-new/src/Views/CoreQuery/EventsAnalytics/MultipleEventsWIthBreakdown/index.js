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
import { ACTIVE_USERS_CRITERIA, FREQUENCY_CRITERIA } from "../../../../utils/constants";

function MultipleEventsWithBreakdown({
  queries,
  breakdown,
  resultState,
  page,
  chartType,
  isWidgetModal,
  durationObj,
  title,
  section
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
  const arrayMapper = [];

  const visibleLabels = visibleProperties.map((v) => `${v.event},${v.label}`);

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
    visibleProperties,
    arrayMapper,
    hiddenProperties
  );

  let chart = null;
  const table = (
    <div className="mt-12 w-full">
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
  );
  
  const appliedColors = generateColors(visibleProperties.length);

  if (chartType === "barchart") {
    chart = (
      <BarChart
        chartData={formatVisibleProperties(visibleProperties, queries)}
        queries={queries}
        title={title}
      />
    );
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
        isDecimalAllowed={page === ACTIVE_USERS_CRITERIA || page === FREQUENCY_CRITERIA}
        arrayMapper={arrayMapper}
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

export default MultipleEventsWithBreakdown;
