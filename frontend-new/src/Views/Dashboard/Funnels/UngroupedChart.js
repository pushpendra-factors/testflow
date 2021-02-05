import React, { useEffect, useState } from "react";
import { generateUngroupedChartsData } from "../../CoreQuery/FunnelsResultPage/utils";
import Chart from "../../CoreQuery/FunnelsResultPage/UngroupedChart/Chart";
import FunnelsResultTable from "../../CoreQuery/FunnelsResultPage/FunnelsResultTable";
import { CHART_TYPE_BARCHART, CHART_TYPE_TABLE } from "../../../utils/constants";

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

  if (chartType === CHART_TYPE_BARCHART) {
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

export default UngroupedChart;
