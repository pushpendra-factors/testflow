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
import { ACTIVE_USERS_CRITERIA, FREQUENCY_CRITERIA } from "../../../utils/constants";
// import BreakdownType from '../BreakdownType';

function MultipleEventsWithBreakdown({
  queries,
  resultState,
  page,
  chartType,
  title,
  breakdown,
  unit,
  durationObj,
  section
}) {
  const [chartsData, setChartsData] = useState([]);
  const [visibleProperties, setVisibleProperties] = useState([]);
  const [hiddenProperties, setHiddenProperties] = useState([]);

  const maxAllowedVisibleProperties = unit.cardSize ? 5 : 3;

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

  const lineChartData = formatDataInLineChartFormat(
    visibleProperties,
    arrayMapper,
    hiddenProperties
  );
  const appliedColors = generateColors(visibleProperties.length);

  if (chartType === "barchart") {
    chartContent = (
      <div className="flex mt-4">
        <BarChart
          chartData={formatVisibleProperties(visibleProperties, queries)}
          title={title}
          queries={queries}
        />
      </div>
    );
  } else if (chartType === "table") {
    chartContent = (
      <div className="mt-4">
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
          isDecimalAllowed={page === ACTIVE_USERS_CRITERIA || page === FREQUENCY_CRITERIA}
          arrayMapper={arrayMapper}
          cardSize={unit.cardSize}
          section={section}
          height={200}
        />
      </div>
    );
  }

  return <div style={{ boxShadow: "inset 0px 1px 0px rgba(0, 0, 0, 0.1)" }} className="w-full px-6">{chartContent}</div>;
}

export default MultipleEventsWithBreakdown;
