import React, { useState, useEffect } from "react";
import { formatData } from "../../CoreQuery/EventsAnalytics/EventBreakdown/utils";
import BarChart from "../../../components/BarChart";
import EventBreakdownTable from "../../CoreQuery/EventsAnalytics/EventBreakdown/EventBreakdownTable";
import ChartHeader from "../../../components/SparkLineChart/ChartHeader";
import {
  CHART_TYPE_TABLE,
  CHART_TYPE_BARCHART,
  CHART_TYPE_SPARKLINES,
  DASHBOARD_WIDGET_BAR_CHART_HEIGHT,
} from "../../../utils/constants";
import NoDataChart from "../../../components/NoDataChart";

function EventBreakdownCharts({
  resultState,
  breakdown,
  section,
  chartType,
  setwidgetModal,
  unit,
}) {
  const [chartsData, setChartsData] = useState([]);
  const [visibleProperties, setVisibleProperties] = useState([]);
  const maxAllowedVisibleProperties = 5;

  useEffect(() => {
    const formattedData = formatData(resultState.data);
    setChartsData(formattedData);
    setVisibleProperties([
      ...formattedData.slice(0, maxAllowedVisibleProperties),
    ]);
  }, [resultState.data]);

  if (!chartsData.length) {
    return (
      <div className="mt-4 flex justify-center items-center w-full h-64 ">
        <NoDataChart />
      </div>
    );
  }

  let chartContent = null;

  let tableContent = null;

  if (chartType === CHART_TYPE_BARCHART) {
    chartContent = (
      <BarChart
        height={DASHBOARD_WIDGET_BAR_CHART_HEIGHT}
        title={unit.id}
        section={section}
        chartData={visibleProperties}
        cardSize={unit.cardSize}
      />
    );
  } else if (chartType === CHART_TYPE_SPARKLINES) {
    chartContent = (
      <ChartHeader
        total={resultState.data.rows[0]}
        query={"Count"}
        bgColor="#4D7DB4"
      />
    );
  } else {
    chartContent = (
      <EventBreakdownTable
        data={chartsData}
        breakdown={breakdown}
        setVisibleProperties={setVisibleProperties}
        visibleProperties={visibleProperties}
        maxAllowedVisibleProperties={maxAllowedVisibleProperties}
      />
    );
  }

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
    <div className={`w-full px-6 flex flex-1 flex-col  justify-center`}>
      {chartContent}
      {tableContent}
    </div>
  );
}

export default EventBreakdownCharts;
