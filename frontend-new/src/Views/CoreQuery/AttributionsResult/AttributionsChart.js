import React, { useState, useEffect } from "react";
import { formatData } from "./utils";
import BarLineChart from "../../../components/BarLineChart";
import AttributionTable from "./AttributionTable";

function AttributionsChart({ data, isWidgetModal, event, attribution_method }) {
  const maxAllowedVisibleProperties = 5;
  const [chartsData, setChartsData] = useState([]);
  const [visibleIndices, setVisibleIndices] = useState(
    Array.from(Array(maxAllowedVisibleProperties).keys())
  );

  useEffect(() => {
    const formattedData = formatData(data, event, visibleIndices);
    setChartsData(formattedData);
  }, [data, event, visibleIndices]);

  if (!chartsData.length) {
    return null;
  }

  return (
    <div className="attribution-results">
      <BarLineChart chartData={chartsData} />
      <div className="mt-8">
        <AttributionTable
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
