import React, { useState, useEffect } from "react";
import { formatData } from "../../CoreQuery/AttributionsResult/utils";
import AttributionTable from "../../CoreQuery/AttributionsResult/AttributionTable";
import BarLineChart from "../../../components/BarLineChart";

function SingleTouchPoint({
  data,
  isWidgetModal,
  event,
  attribution_method,
  touchpoint,
  linkedEvents,
  setwidgetModal,
  chartType
}) {
  const maxAllowedVisibleProperties = 5;
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

  if (chartType === "barchart") {
    chartContent = (
      <div className="mt-4">
        <BarLineChart chartData={chartsData} />
      </div>
    );
  } else {
    chartContent = (
      <div className="mt-4">
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

export default SingleTouchPoint;
