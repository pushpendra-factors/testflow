import React, { useEffect, useState } from "react";
import {
  generateEventsData,
  generateGroups,
  generateGroupedChartsData,
} from "../utils";
import Chart from "./Chart";
import FunnelsResultTable from "../FunnelsResultTable";
import { DASHBOARD_MODAL } from "../../../../utils/constants";

function GroupedChart({
  resultState,
  queries,
  breakdown,
  isWidgetModal,
  arrayMapper,
  section
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
    <div className="flex items-center justify-center flex-col">
      <Chart
        isWidgetModal={isWidgetModal}
        chartData={chartData}
        groups={groups.filter((elem) => elem.is_visible)}
        eventsData={eventsData}
        arrayMapper={arrayMapper}
        section={section}
      />

      <div className="mt-12 w-full">
        <FunnelsResultTable
          breakdown={breakdown}
          queries={queries}
          groups={groups}
          setGroups={setGroups}
          chartData={eventsData}
          arrayMapper={arrayMapper}
          maxAllowedVisibleProperties={maxAllowedVisibleProperties}
          isWidgetModal={section === DASHBOARD_MODAL}
        />
      </div>
    </div>
  );
}

export default GroupedChart;
