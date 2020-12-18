import React, { useState, useEffect, useCallback } from "react";
import AttributionTable from "../../CoreQuery/AttributionsResult/AttributionTable";
import GroupedBarChart from "../../../components/GroupedBarChart";
import { formatGroupedData } from "../../CoreQuery/AttributionsResult/utils";

function DualTouchPoint({
  data,
  isWidgetModal,
  event,
  attribution_method,
  attribution_method_compare,
  touchpoint,
  linkedEvents,
  chartType,
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
    const { headers } = data.result;
    const campaignIdx = headers.indexOf("Campaign");
    return data.result.rows
      .filter((_, index) => visibleIndices.indexOf(index) > -1)
      .map((row) => row[campaignIdx]);
  }, [visibleIndices, data.result]);

  if (!chartsData.length) {
    return null;
  }

  const getColors = () => {
    return {
      [chartsData[0][0]]: "#4D7DB4",
      [chartsData[1][0]]: "#4CBCBD",
    };
  };
  let chartContent = null;

  if (chartType === "barchart") {
    chartContent = (
      <div className="mt-4">
        <GroupedBarChart
          colors={getColors()}
          categories={getCategories()}
          chartData={chartsData}
          visibleIndices={visibleIndices}
          responseRows={data.result.rows}
          responseHeaders={data.result.headers}
          method1={attribution_method}
          method2={attribution_method_compare}
          event={event}
        />
      </div>
    );
  } else {
    chartContent = (
      <div className="mt-4">
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
    );
  }

  let tableContent = null;

  if (chartType === "table") {
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

  return (
    <div className="total-events w-full">
      {chartContent}
      {tableContent}
    </div>
  );
}

export default DualTouchPoint;
