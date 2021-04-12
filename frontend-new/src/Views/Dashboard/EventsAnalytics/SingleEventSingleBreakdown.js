import React, { useState, useEffect } from "react";
import {
  formatData,
  formatDataInLineChartFormat,
  formatDataInStackedAreaFormat,
} from "../../CoreQuery/EventsAnalytics/SingleEventSingleBreakdown/utils";
import BarChart from "../../../components/BarChart";
import SingleEventSingleBreakdownTable from "../../CoreQuery/EventsAnalytics/SingleEventSingleBreakdown/SingleEventSingleBreakdownTable";
import LineChart from "../../../components/LineChart";
import StackedAreaChart from "../../../components/StackedAreaChart";
import { generateColors } from "../../../utils/dataFormatter";
import {
  ACTIVE_USERS_CRITERIA,
  FREQUENCY_CRITERIA,
  CHART_TYPE_TABLE,
  DASHBOARD_WIDGET_BAR_CHART_HEIGHT,
  CHART_TYPE_BARCHART,
  DASHBOARD_WIDGET_LINE_CHART_HEIGHT,
  CHART_TYPE_STACKED_AREA,
  DASHBOARD_WIDGET_AREA_CHART_HEIGHT,
} from "../../../utils/constants";

function SingleEventSingleBreakdown({
  resultState,
  page,
  chartType,
  breakdown,
  queries,
  unit,
  durationObj,
  section,
  setwidgetModal,
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
  }, [resultState.data, maxAllowedVisibleProperties]);

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

  if (chartType === CHART_TYPE_BARCHART) {
    chartContent = (
      <BarChart
        chartData={visibleProperties}
        height={DASHBOARD_WIDGET_BAR_CHART_HEIGHT}
        title={unit.id}
        cardSize={unit.cardSize}
        section={section}
        queries={queries}
      />
    );
  } else if (chartType === CHART_TYPE_TABLE) {
    chartContent = (
      <SingleEventSingleBreakdownTable
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
    );
  } else if (chartType === CHART_TYPE_STACKED_AREA) {
    const { categories, data } = formatDataInStackedAreaFormat(
      resultState.data,
      visibleLabels,
      arrayMapper
    );
    chartContent = (
      <StackedAreaChart
        frequency={durationObj.frequency}
        categories={categories}
        data={data}
        height={DASHBOARD_WIDGET_AREA_CHART_HEIGHT}
        legendsPosition="top"
        cardSize={unit.cardSize}
      />
    );
  } else {
    chartContent = (
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
        cardSize={unit.cardSize}
        section={section}
        height={DASHBOARD_WIDGET_LINE_CHART_HEIGHT}
      />
    );
  }

  return (
    <div className={`w-full px-6 flex flex-1 flex-col justify-center`}>
      {chartContent}
      {tableContent}
    </div>
  );
}

export default SingleEventSingleBreakdown;
