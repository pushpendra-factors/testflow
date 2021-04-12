import React, { useState, useEffect } from "react";
import {
  formatData,
  formatVisibleProperties,
  formatDataInLineChartFormat,
  formatDataInStackedAreaFormat,
} from "./utils";
import BarChart from "../../../../components/BarChart";
import MultipleEventsWithBreakdownTable from "./MultipleEventsWithBreakdownTable";
import LineChart from "../../../../components/LineChart";
import { generateColors } from "../../../../utils/dataFormatter";
import {
  ACTIVE_USERS_CRITERIA,
  FREQUENCY_CRITERIA,
  DASHBOARD_MODAL,
  CHART_TYPE_BARCHART,
  CHART_TYPE_STACKED_AREA,
} from "../../../../utils/constants";
import StackedAreaChart from "../../../../components/StackedAreaChart";

function MultipleEventsWithBreakdown({
  queries,
  breakdown,
  resultState,
  page,
  chartType,
  durationObj,
  title,
  section,
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
        isWidgetModal={section === DASHBOARD_MODAL}
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
  if (chartType === CHART_TYPE_BARCHART) {
    chart = (
      <BarChart
        section={section}
        chartData={formatVisibleProperties(visibleProperties, queries)}
        queries={queries}
        title={title}
      />
    );
  } else if (chartType === CHART_TYPE_STACKED_AREA) {
    const { categories, data } = formatDataInStackedAreaFormat(
      resultState.data,
      visibleLabels,
      arrayMapper
    );
    chart = (
      <div className="w-full">
        <StackedAreaChart
          frequency={durationObj.frequency}
          categories={categories}
          data={data}
        />
      </div>
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
        isDecimalAllowed={
          page === ACTIVE_USERS_CRITERIA || page === FREQUENCY_CRITERIA
        }
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
