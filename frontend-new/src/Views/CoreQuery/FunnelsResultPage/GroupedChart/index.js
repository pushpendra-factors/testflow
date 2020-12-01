import React, { useEffect, useState } from 'react';
import { generateEventsData, generateGroups, generateGroupedChartsData } from '../utils';
import Chart from './Chart';
import FunnelsResultTable from '../FunnelsResultTable';

function GroupedChart({
  resultState, queries, breakdown, eventsMapper, reverseEventsMapper, isWidgetModal
}) {
  const [groups, setGroups] = useState([]);
  const maxAllowedVisibleProperties = 5;

  useEffect(() => {
    const formattedGroups = generateGroups(resultState.data, maxAllowedVisibleProperties);
    setGroups(formattedGroups);
  }, [queries, resultState.data]);

  if (!groups.length) {
    return null;
  }

  const chartData = generateGroupedChartsData(resultState.data, queries, groups, eventsMapper);
  const eventsData = generateEventsData(resultState.data, queries, eventsMapper);

  return (
    <>

      <Chart
        isWidgetModal={isWidgetModal}
        chartData={chartData}
        groups={groups.filter(elem => elem.is_visible)}
        eventsData={eventsData}
        eventsMapper={eventsMapper}
        reverseEventsMapper={reverseEventsMapper}
      />

      <div className="mt-8">
        <FunnelsResultTable
          breakdown={breakdown}
          queries={queries}
          groups={groups}
          setGroups={setGroups}
          chartData={eventsData}
          eventsMapper={eventsMapper}
          maxAllowedVisibleProperties={maxAllowedVisibleProperties}
          isWidgetModal={isWidgetModal}
        />
      </div>
    </>
  );
}

export default GroupedChart;
