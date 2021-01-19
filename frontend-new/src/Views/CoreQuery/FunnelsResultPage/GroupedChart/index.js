import React, { useEffect, useState } from "react";
import {
  generateEventsData,
  generateGroups,
  generateGroupedChartsData,
} from "../utils";
import Chart from "./Chart";
import FunnelsResultTable from "../FunnelsResultTable";

function GroupedChart({
  resultState,
  queries,
  breakdown,
  isWidgetModal,
  arrayMapper,
}) {
  const [groups, setGroups] = useState([]);
  const maxAllowedVisibleProperties = 5;

  useEffect(() => {
    const formattedGroups = generateGroups(
      resultState.data,
      maxAllowedVisibleProperties
    );
    setGroups(formattedGroups);
  }, [resultState.data]);

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

  return (
    <>
      <Chart
        isWidgetModal={isWidgetModal}
        chartData={chartData}
        groups={groups.filter((elem) => elem.is_visible)}
        eventsData={eventsData}
        arrayMapper={arrayMapper}
      />

      <div className="mt-8">
        <FunnelsResultTable
          breakdown={breakdown}
          queries={queries}
          groups={groups}
          setGroups={setGroups}
          chartData={eventsData}
          arrayMapper={arrayMapper}
          maxAllowedVisibleProperties={maxAllowedVisibleProperties}
          isWidgetModal={isWidgetModal}
          arrayMapper={arrayMapper}
        />
      </div>
    </>
  );
}

export default GroupedChart;
