import React, { useState, useEffect } from "react";
import { formatData } from "./utils";
import DualAxesChart from "../../../components/DualAxesChart";

function AttributionsChart({ data }) {
  const [chartsData, setChartsData] = useState([]);
  const maxAllowedVisibleProperties = 5;

  useEffect(() => {
    const formattedData = formatData(data);
    console.log(formattedData);
    setChartsData(formattedData.slice(0, maxAllowedVisibleProperties));
  }, [data]);

  if (!chartsData.length) {
    return null;
  }

  return (
    <div className="attribution-results">
      <DualAxesChart chartData={chartsData} />
    </div>
  );
}

export default AttributionsChart;
