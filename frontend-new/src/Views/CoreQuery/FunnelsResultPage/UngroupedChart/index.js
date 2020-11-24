import React, { useEffect, useState } from 'react';
import { generateUngroupedChartsData } from '../utils';

import Chart from './Chart';
import FunnelsResultTable from '../FunnelsResultTable';

function UngroupedChart({
  resultState, queries, eventsMapper
}) {
  const [chartData, setChartData] = useState([]);

  useEffect(() => {
    const formattedData = generateUngroupedChartsData(resultState.data, queries);
    setChartData(formattedData);
  }, [queries, resultState.data]);

  if (!chartData.length) {
    return null;
  }

  return (
    <>
      <Chart
        chartData={chartData}
      />

      <div className="mt-8">
        <FunnelsResultTable
          chartData={chartData}
          breakdown={[]}
          queries={queries}
          groups={[]}
          eventsMapper={eventsMapper}
        />
      </div>
    </>
  );
}

export default UngroupedChart;
