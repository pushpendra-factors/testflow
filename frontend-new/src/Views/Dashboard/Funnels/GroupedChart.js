import React, { useEffect, useState, useContext } from "react";
import {
  generateEventsData,
  generateGroups,
  generateGroupedChartsData,
} from "../../CoreQuery/FunnelsResultPage/utils";
import Chart from "../../CoreQuery/FunnelsResultPage/GroupedChart/Chart";
import FunnelsResultTable from "../../CoreQuery/FunnelsResultPage/FunnelsResultTable";
import { DashboardContext } from "../../../contexts/DashboardContext";

function GroupedChart({
  resultState,
  queries,
  arrayMapper,
  breakdown,
  chartType,
  unit,
  section,
}) {
  const [groups, setGroups] = useState([]);
  const { handleEditQuery } = useContext(DashboardContext);
  const maxAllowedVisibleProperties = 5;

  useEffect(() => {
    const formattedGroups = generateGroups(
      resultState.data,
      maxAllowedVisibleProperties
    );
    setGroups(formattedGroups);
  }, [queries, resultState.data, maxAllowedVisibleProperties]);

  if (!groups.length) {
    return null;
  }

  const chartData = generateGroupedChartsData(
    resultState.data,
    queries,
    groups,
    arrayMapper
  );
  const eventsData = generateEventsData(resultState.data, queries, arrayMapper);

  let chartContent = null;

  if (chartType === "barchart") {
    chartContent = (
      <Chart
        chartData={chartData}
        groups={groups.filter((elem) => elem.is_visible)}
        eventsData={eventsData}
        title={unit.id}
        arrayMapper={arrayMapper}
        height={225}
        section={section}
        cardSize={unit.cardSize}
        durations={resultState.data.meta}
      />
    );
  } else {
    chartContent = (
      <FunnelsResultTable
        breakdown={breakdown}
        queries={queries}
        groups={groups}
        setGroups={setGroups}
        chartData={eventsData}
        arrayMapper={arrayMapper}
        maxAllowedVisibleProperties={maxAllowedVisibleProperties}
        durations={resultState.data.meta}
      />
    );
  }

  let tableContent = null;

  if (chartType === "table") {
    tableContent = (
      <div
        onClick={handleEditQuery}
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

export default GroupedChart;
