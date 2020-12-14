import React, { useState, useEffect } from "react";
import { formatData } from "./utils";
import BarLineChart from "../../../components/BarLineChart";

function AttributionsChart({ data }) {
  const [chartsData, setChartsData] = useState([]);
  const maxAllowedVisibleProperties = 5;

  useEffect(() => {
    const formattedData = formatData(data);
    setChartsData(formattedData.slice(0, maxAllowedVisibleProperties));
  }, [data]);

  if (!chartsData.length) {
    return null;
  }

  return (
    <div className="attribution-results">
      <BarLineChart chartData={chartsData} />
    </div>
  );
}

export default AttributionsChart;
