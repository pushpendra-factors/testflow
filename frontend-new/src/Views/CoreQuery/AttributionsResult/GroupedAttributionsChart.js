import React, { useState, useEffect, useCallback } from "react";
import AttributionTable from "./AttributionTable";
import { formatGroupedData } from "./utils";
import GroupedBarChart from "../../../components/GroupedBarChart";

function GroupedAttributionsChart({
  data,
  isWidgetModal,
  event,
  attribution_method,
  attribution_method_compare,
  touchpoint,
  linkedEvents,
  section
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
      attribution_method_compare
    );
    setChartsData(formattedData);
  }, [
    data,
    event,
    visibleIndices,
    attribution_method,
    attribution_method_compare,
  ]);

  const getCategories = useCallback(() => {
    const { headers } = data;
    const campaignIdx = headers.indexOf(touchpoint);
    return data.rows
      .filter((_, index) => visibleIndices.indexOf(index) > -1)
      .map((row) => row[campaignIdx]);
  }, [visibleIndices, data, touchpoint]);

  if (!chartsData.length) {
    return null;
  }

  const getColors = () => {
    return {
      [chartsData[0][0]]: "#4D7DB4",
      [chartsData[1][0]]: "#4CBCBD",
    };
  };

  return (
    <div className="flex items-center justify-center flex-col">
      <GroupedBarChart
        colors={getColors()}
        categories={getCategories()}
        chartData={chartsData}
        visibleIndices={visibleIndices}
        responseRows={data.rows}
        responseHeaders={data.headers}
        method1={attribution_method}
        method2={attribution_method_compare}
        event={event}
        section={section}
      />
      <div className="mt-12 w-full">
        <AttributionTable
          touchpoint={touchpoint}
          linkedEvents={linkedEvents}
          event={event}
          data={data}
          isWidgetModal={isWidgetModal}
          visibleIndices={visibleIndices}
          setVisibleIndices={setVisibleIndices}
          maxAllowedVisibleProperties={maxAllowedVisibleProperties}
          attribution_method={attribution_method}
          attribution_method_compare={attribution_method_compare}
        />
      </div>
    </div>
  );
}

export default GroupedAttributionsChart;
