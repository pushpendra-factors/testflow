import React, { useEffect, useState } from "react";
import { generateUngroupedChartsData } from "../../CoreQuery/FunnelsResultPage/utils";
import Chart from "../../CoreQuery/FunnelsResultPage/UngroupedChart/Chart";
import FunnelsResultTable from "../../CoreQuery/FunnelsResultPage/FunnelsResultTable";

function UngroupedChart({
  resultState,
  queries,
  title,
  chartType,
  setwidgetModal,
  unit,
  arrayMapper,
}) {
  const [chartData, setChartData] = useState([]);

  useEffect(() => {
    const formattedData = generateUngroupedChartsData(
      resultState.data,
      arrayMapper
    );
    setChartData(formattedData);
  }, [arrayMapper, resultState.data]);

  if (!chartData.length) {
    return null;
  }

  let chartContent = null;

  if (chartType === "barchart") {
    chartContent = (
      <Chart
        title={title}
        chartData={chartData}
        cardSize={unit.cardSize}
        arrayMapper={arrayMapper}
        height={275}
      />
    );
  } else {
    chartContent = (
      <FunnelsResultTable
        chartData={chartData}
        breakdown={[]}
        queries={queries}
        groups={[]}
        arrayMapper={arrayMapper}
      />
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
    <div
      style={{
        boxShadow:
          chartType === "barchart"
            ? "inset 0px 1px 0px rgba(0, 0, 0, 0.1)"
            : "",
      }}
      className="w-full px-6 mt-4"
    >
      {chartContent}
      {tableContent}
    </div>
  );
}

export default UngroupedChart;
