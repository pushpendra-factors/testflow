import React, { useState, useEffect } from "react";
import { formatData } from "./utils";
import BarChart from "../../../../components/BarChart";
import EventBreakdownTable from "./EventBreakdownTable";
import ChartHeader from "../../../../components/SparkLineChart/ChartHeader";

function EventBreakdownCharts({ data, breakdown }) {
  const [chartsData, setChartsData] = useState([]);
  const [visibleProperties, setVisibleProperties] = useState([]);
  const maxAllowedVisibleProperties = 5;

  useEffect(() => {
    const formattedData = formatData(data);
    setChartsData(formattedData);
    setVisibleProperties([
      ...formattedData.slice(0, maxAllowedVisibleProperties),
    ]);
  }, [data]);

  if (!chartsData.length) {
    return null;
  }

  let chart = null;

  const table = (
    <div className="mt-12 w-full">
      <EventBreakdownTable
        data={chartsData}
        breakdown={breakdown}
        setVisibleProperties={setVisibleProperties}
        visibleProperties={visibleProperties}
        maxAllowedVisibleProperties={maxAllowedVisibleProperties}
      />
    </div>
  );

  if (breakdown.length) {
    chart = <BarChart chartData={visibleProperties} />;
  } else {
    chart = (
      <ChartHeader total={data.rows[0]} query={"Count"} bgColor="#4D7DB4" />
    );
  }

  return (
    <div className="flex items-center justify-center flex-col">
      {chart}
      {table}
    </div>
  );
}

export default EventBreakdownCharts;
