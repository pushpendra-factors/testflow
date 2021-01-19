import React, { useEffect, useState } from 'react';
import { generateEventsData, generateGroups, generateGroupedChartsData } from '../../CoreQuery/FunnelsResultPage/utils';
import Chart from '../../CoreQuery/FunnelsResultPage/GroupedChart/Chart';
import FunnelsResultTable from '../../CoreQuery/FunnelsResultPage/FunnelsResultTable';

function GroupedChart({
  resultState, queries, arrayMapper, title, breakdown, chartType, unit, setwidgetModal
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

  const chartData = generateGroupedChartsData(resultState.data, queries, groups, arrayMapper);
  const eventsData = generateEventsData(resultState.data, queries, arrayMapper);

  let chartContent = null;

  if (chartType === 'barchart') {
    chartContent = (
      <div className="mt-4">
        <Chart
          chartData={chartData}
          groups={groups.filter(elem => elem.is_visible)}
          eventsData={eventsData}
          title={title}
          arrayMapper={arrayMapper}
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
          arrayMapper={arrayMapper}
          maxAllowedVisibleProperties={maxAllowedVisibleProperties}
        />
      </div>
    );
  }
  
  let tableContent = null;

  if (chartType === 'table') {
    tableContent = (
      <div onClick={() => setwidgetModal({ unit, data: resultState.data })} style={{ color: '#5949BC' }} className="mt-3 font-medium text-base cursor-pointer flex justify-end item-center">Show More &rarr;</div>
    )
  }

  return (
    <div className="total-events w-full">
      {chartContent}
      {tableContent}
    </div>
  );
}

export default GroupedChart;
