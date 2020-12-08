import React, { useState, useEffect } from "react";
// import ChartTypeDropdown from '../../../components/ChartTypeDropdown';
import {
  formatData,
  formatDataInLineChartFormat,
} from "../../CoreQuery/EventsAnalytics/SingleEventSingleBreakdown/utils";
import BarChart from "../../../components/BarChart";
import SingleEventSingleBreakdownTable from "../../CoreQuery/EventsAnalytics/SingleEventSingleBreakdown/SingleEventSingleBreakdownTable";
import LineChart from "../../../components/LineChart";
import { generateColors } from "../../../utils/dataFormatter";

function SingleEventSingleBreakdown({
  resultState,
  page,
  chartType,
  title,
  breakdown,
  queries,
  unit,
  durationObj,
}) {
  const [chartsData, setChartsData] = useState([]);
  const [visibleProperties, setVisibleProperties] = useState([]);
  const [hiddenProperties, setHiddenProperties] = useState([]);

  const maxAllowedVisibleProperties = unit.cardSize ? 5 : 3;

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

  if (chartType === "barchart") {
    chartContent = (
      <div className="flex mt-4">
        <BarChart title={title} chartData={visibleProperties} />
      </div>
    );
  } else if (chartType === "table") {
    chartContent = (
      <div className="mt-4">
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
      </div>
    );
  } else {
    chartContent = (
      <div className="flex mt-4">
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
      </div>
    );
  }

  return <div className="total-events w-full">{chartContent}</div>;
}

export default SingleEventSingleBreakdown;
