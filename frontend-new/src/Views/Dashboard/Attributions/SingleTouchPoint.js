import React, { useState, useEffect } from "react";
import { formatData } from "../../CoreQuery/AttributionsResult/utils";
import AttributionTable from "../../CoreQuery/AttributionsResult/AttributionTable";
import BarLineChart from "../../../components/BarLineChart";
import {
  CHART_TYPE_BARCHART,
  CHART_TYPE_TABLE,
} from "../../../utils/constants";

function SingleTouchPoint({
  data,
  isWidgetModal,
  event,
  attribution_method,
  touchpoint,
  linkedEvents,
  setwidgetModal,
  chartType,
  title,
  resultState,
  unit,
  section,
}) {
  const maxAllowedVisibleProperties = unit.cardSize ? 5 : 3;
  const [chartsData, setChartsData] = useState([]);
  const [visibleIndices, setVisibleIndices] = useState(
    Array.from(Array(maxAllowedVisibleProperties).keys())
  );

  useEffect(() => {
    const formattedData = formatData(data, event, visibleIndices, touchpoint);
    setChartsData(formattedData);
  }, [data, event, visibleIndices, touchpoint]);

  if (!chartsData.length) {
    return null;
  }

  let chartContent = null;
  
  if (chartType === CHART_TYPE_BARCHART) {
    chartContent = (
      <BarLineChart
        responseRows={data.rows}
        responseHeaders={data.headers}
        visibleIndices={visibleIndices}
        title={title}
        chartData={chartsData}
        section={section}
        height={225}
      />
    );
  } else {
    chartContent = (
      <AttributionTable
        linkedEvents={linkedEvents}
        touchpoint={touchpoint}
        event={event}
        data={data}
        isWidgetModal={isWidgetModal}
        visibleIndices={visibleIndices}
        setVisibleIndices={setVisibleIndices}
        maxAllowedVisibleProperties={maxAllowedVisibleProperties}
        attribution_method={attribution_method}
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
      style={{
        boxShadow:
          chartType === CHART_TYPE_BARCHART
            ? "inset 0px 1px 0px rgba(0, 0, 0, 0.1)"
            : "",
      }}
      className="w-full px-6"
    >
      {chartContent}
      {tableContent}
    </div>
  );
}

export default SingleTouchPoint;
