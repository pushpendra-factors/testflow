import React, { useState, useEffect } from "react";
import {
  formatData,
  formatVisibleProperties,
  formatDataInLineChartFormat,
} from "../../CoreQuery/EventsAnalytics/MultipleEventsWIthBreakdown/utils";
import BarChart from "../../../components/BarChart";
import MultipleEventsWithBreakdownTable from "../../CoreQuery/EventsAnalytics/MultipleEventsWIthBreakdown/MultipleEventsWithBreakdownTable";
import LineChart from "../../../components/LineChart";
import { generateColors } from "../../../utils/dataFormatter";
import {
  ACTIVE_USERS_CRITERIA,
  FREQUENCY_CRITERIA,
  CHART_TYPE_TABLE,
  CHART_TYPE_BARCHART,
  DASHBOARD_WIDGET_LINE_CHART_HEIGHT,
  DASHBOARD_WIDGET_MULTICOLORED_BAR_CHART_HEIGHT,
} from "../../../utils/constants";
// import BreakdownType from '../BreakdownType';

function MultipleEventsWithBreakdown({
  queries,
  resultState,
  page,
  chartType,
  breakdown,
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
    const appliedColors = generateColors(queries.length);
    const formattedData = formatData(resultState.data, queries, appliedColors);
    setChartsData(formattedData);
    setVisibleProperties([
      ...formattedData.slice(0, maxAllowedVisibleProperties),
    ]);
  }, [resultState.data, queries, maxAllowedVisibleProperties]);

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

  const lineChartData = formatDataInLineChartFormat(
    visibleProperties,
    arrayMapper,
    hiddenProperties
  );
  const appliedColors = generateColors(visibleProperties.length);

  if (chartType === CHART_TYPE_BARCHART) {
    chartContent = (
      <BarChart
        chartData={formatVisibleProperties(visibleProperties, queries)}
        height={DASHBOARD_WIDGET_MULTICOLORED_BAR_CHART_HEIGHT}
        title={unit.id}
        cardSize={unit.cardSize}
        section={section}
        queries={queries}
      />
    );
  } else if (chartType === CHART_TYPE_TABLE) {
    chartContent = (
      <MultipleEventsWithBreakdownTable
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
    <div className={`w-full px-6 flex flex-1 flex-col  justify-center`}>
      {chartContent}
      {tableContent}
    </div>
  );
}

export default MultipleEventsWithBreakdown;
