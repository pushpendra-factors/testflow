import React, { useEffect, useState } from 'react';
import { generateEventsData, generateGroups, generateGroupedChartsData } from '../../CoreQuery/FunnelsResultPage/utils';
import Chart from '../../CoreQuery/FunnelsResultPage/GroupedChart/Chart';
// import FunnelsResultTable from '../FunnelsResultTable';

function GroupedChart({
  resultState, queries, eventsMapper, reverseEventsMapper, title
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

            <div>

                <Chart
                    chartData={chartData}
                    groups={groups.filter(elem => elem.is_visible)}
                    eventsData={eventsData}
                    eventsMapper={eventsMapper}
                    reverseEventsMapper={reverseEventsMapper}
                    title={title}
                />

                {/* <div className="mt-8">
          <FunnelsResultTable
            breakdown={breakdown}
            queries={queries}
            groups={groups}
            setGroups={setGroups}
            chartData={eventsData}
            eventsMapper={eventsMapper}
            maxAllowedVisibleProperties={maxAllowedVisibleProperties}
          />
        </div> */}
            </div>
    </>
  );
}

export default GroupedChart;
