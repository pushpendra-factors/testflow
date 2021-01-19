import React, { useState, useEffect } from "react";
import { formatData } from "./utils";
import BarLineChart from "../../../components/BarLineChart";
import AttributionTable from "./AttributionTable";

function AttributionsChart({
  data,
  isWidgetModal,
  event,
  attribution_method,
  touchpoint,
  linkedEvents,
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

  return (
    <div className="attribution-results">
      <BarLineChart
        responseRows={data.rows}
        responseHeaders={data.headers}
        chartData={chartsData}
        visibleIndices={visibleIndices}
      />
      <div className="mt-8">
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
    </div>
  );
}

export default AttributionsChart;
