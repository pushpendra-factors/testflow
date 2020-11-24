import React, { useEffect, useState } from 'react';
import { generateEventsData, generateGroups, generateGroupedChartsData } from '../../CoreQuery/FunnelsResultPage/utils';
import Chart from '../../CoreQuery/FunnelsResultPage/GroupedChart/Chart';
import FunnelsResultTable from '../../CoreQuery/FunnelsResultPage/FunnelsResultTable';

function GroupedChart({
  resultState, queries, eventsMapper, reverseEventsMapper, title, breakdown, chartType, unit
}) {
  const [groups, setGroups] = useState([]);
  const maxAllowedVisibleProperties = unit.cardSize ? 5 : 3;

  useEffect(() => {
    const formattedGroups = generateGroups(resultState.data, maxAllowedVisibleProperties);
    setGroups(formattedGroups);
  }, [queries, resultState.data, maxAllowedVisibleProperties]);

  if (!groups.length) {
    return null;
  }

  const chartData = generateGroupedChartsData(resultState.data, queries, groups, eventsMapper);
  const eventsData = generateEventsData(resultState.data, queries, eventsMapper);

  let chartContent = null;

  if (chartType === 'barchart') {
    chartContent = (
      <div className="mt-4">
        <Chart
          chartData={chartData}
          groups={groups.filter(elem => elem.is_visible)}
          eventsData={eventsData}
          eventsMapper={eventsMapper}
          reverseEventsMapper={reverseEventsMapper}
          title={title}
        />
      </div>
    );
  } else {
    chartContent = (
      <div className="mt-4">
        <FunnelsResultTable
          breakdown={breakdown}
          queries={queries}
          groups={groups}
          setGroups={setGroups}
          chartData={eventsData}
          eventsMapper={eventsMapper}
          maxAllowedVisibleProperties={maxAllowedVisibleProperties}
        />
      </div>
    );
  }

  return (
    <div className="total-events w-full">
      {chartContent}
    </div>
  );
}

export default GroupedChart;
