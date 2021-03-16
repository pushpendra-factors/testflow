import React, { useState, useEffect } from "react";
import { formatData } from "./utils";
import BarLineChart from "../../../components/BarLineChart";
import AttributionTable from "./AttributionTable";
import { DASHBOARD_MODAL } from "../../../utils/constants";

function AttributionsChart({
  data,
  event,
  attribution_method,
  touchpoint,
  linkedEvents,
  section,
  data2,
  durationObj,
  cmprDuration
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
    <div className="flex items-center justify-center flex-col">
      <BarLineChart
        responseRows={data.rows}
        responseHeaders={data.headers}
        chartData={chartsData}
        visibleIndices={visibleIndices}
        section={section}
      />
      <div className="mt-12 w-full">
        <AttributionTable
          linkedEvents={linkedEvents}
          touchpoint={touchpoint}
          event={event}
          data={data}
          data2={data2}
          durationObj={durationObj}
          cmprDuration={cmprDuration}
          isWidgetModal={section === DASHBOARD_MODAL}
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
