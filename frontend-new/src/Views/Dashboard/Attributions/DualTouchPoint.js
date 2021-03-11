import React, { useState, useEffect, useCallback } from "react";
import AttributionTable from "../../CoreQuery/AttributionsResult/AttributionTable";
import GroupedBarChart from "../../../components/GroupedBarChart";
import { formatGroupedData } from "../../CoreQuery/AttributionsResult/utils";
import { CHART_TYPE_BARCHART, CHART_TYPE_TABLE } from "../../../utils/constants";

function DualTouchPoint({
  data,
  isWidgetModal,
  event,
  attribution_method,
  attribution_method_compare,
  touchpoint,
  linkedEvents,
  chartType,
  unit,
  resultState,
  setwidgetModal,
  section,
}) {
  const maxAllowedVisibleProperties = unit.cardSize ? 5 : 3;
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
    const campaignIdx = headers.indexOf("Campaign");
    return data.rows
      .filter((_, index) => visibleIndices.indexOf(index) > -1)
      .map((row) => row[campaignIdx]);
  }, [visibleIndices, data]);

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

  if (chartType === CHART_TYPE_BARCHART) {
    chartContent = (
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
        height={225}
        cardSize={unit.cardSize}
      />
    );
  } else {
    chartContent = (
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
    );
  }

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

  return (
    <div
      className={`w-full px-6 flex flex-1 flex-col  justify-center`}
    >
      {chartContent}
      {tableContent}
    </div>
  );
}

export default DualTouchPoint;
