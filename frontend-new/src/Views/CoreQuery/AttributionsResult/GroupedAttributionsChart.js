import React, { useState, useEffect } from "react";
import AttributionTable from "./AttributionTable";
import { formatGroupedData } from "./utils";
import GroupedBarChart from "../../../components/GroupedBarChart";
import { DASHBOARD_MODAL } from "../../../utils/constants";

function GroupedAttributionsChart({
  data,
  isWidgetModal,
  event,
  attribution_method,
  attribution_method_compare,
  touchpoint,
  linkedEvents,
  section,
  data2,
  currMetricsValue,
  durationObj,
  cmprDuration
}) {
  const maxAllowedVisibleProperties = 5;
  const [chartsData, setChartsData] = useState([]);
  const [visibleIndices, setVisibleIndices] = useState(
    Array.from(Array(maxAllowedVisibleProperties).keys())
  );

  useEffect(() => {
    const formattedData = formatGroupedData(
      data,
      event,
      visibleIndices,
      attribution_method,
      attribution_method_compare,
      currMetricsValue
    );
    setChartsData(formattedData);
  }, [
    data,
    event,
    visibleIndices,
    attribution_method,
    attribution_method_compare,
    currMetricsValue,
  ]);

  if (!chartsData.length) {
    return null;
  }

  const allValues = [];

  chartsData.forEach((cd) => {
    allValues.push(cd[attribution_method]);
    allValues.push(allValues.push(cd[attribution_method_compare]));
  });

  const getColors = () => {
    return ["#4D7DB4", "#4CBCBD"];
  };

  let legends, tooltipTitle;
  if (currMetricsValue) {
    legends = [
      `Cost Per Conversion (${attribution_method})`,
      `Cost Per Conversion (${attribution_method_compare})`,
    ];
    tooltipTitle = "Cost Per Conversion";
  } else {
    legends = [
      `Conversions as Unique users (${attribution_method})`,
      `Conversions as Unique users (${attribution_method_compare})`,
    ];
    tooltipTitle = "Conversions";
  }

  return (
    <div className="flex items-center justify-center flex-col">
      <GroupedBarChart
        colors={getColors()}
        chartData={chartsData}
        visibleIndices={visibleIndices}
        responseRows={data.rows}
        responseHeaders={data.headers}
        method1={attribution_method}
        method2={attribution_method_compare}
        event={event}
        section={section}
        allValues={allValues}
        legends={legends}
        tooltipTitle={tooltipTitle}
      />
      <div className="mt-12 w-full">
        <AttributionTable
          touchpoint={touchpoint}
          linkedEvents={linkedEvents}
          event={event}
          data={data}
          data2={data2}
          isWidgetModal={section === DASHBOARD_MODAL}
          visibleIndices={visibleIndices}
          setVisibleIndices={setVisibleIndices}
          maxAllowedVisibleProperties={maxAllowedVisibleProperties}
          attribution_method={attribution_method}
          attribution_method_compare={attribution_method_compare}
          durationObj={durationObj}
          cmprDuration={cmprDuration}
        />
      </div>
    </div>
  );
}

export default GroupedAttributionsChart;
