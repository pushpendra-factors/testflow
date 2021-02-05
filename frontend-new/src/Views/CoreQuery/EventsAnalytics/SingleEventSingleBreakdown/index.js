import React, { useState, useEffect } from "react";
import { formatData, formatDataInLineChartFormat } from "./utils";
import BarChart from "../../../../components/BarChart";
import SingleEventSingleBreakdownTable from "./SingleEventSingleBreakdownTable";
import LineChart from "../../../../components/LineChart";
import { generateColors } from "../../../../utils/dataFormatter";
import { ACTIVE_USERS_CRITERIA, FREQUENCY_CRITERIA, DASHBOARD_MODAL } from "../../../../utils/constants";

function SingleEventSingleBreakdown({
  queries,
  breakdown,
  resultState,
  page,
  chartType,
  durationObj,
  title,
  section
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
      <SingleEventSingleBreakdownTable
        isWidgetModal={section === DASHBOARD_MODAL}
        data={chartsData}
        breakdown={breakdown}
        events={queries}
        chartType={chartType}
        page={page}
        setVisibleProperties={setVisibleProperties}
        visibleProperties={visibleProperties}
        maxAllowedVisibleProperties={maxAllowedVisibleProperties}
        lineChartData={lineChartData}
        originalData={resultState.data}
        durationObj={durationObj}
      />
    </div>
  );

  if (chartType === "barchart") {
    chart = <BarChart section={section} title={title} chartData={visibleProperties} />;
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

export default SingleEventSingleBreakdown;
