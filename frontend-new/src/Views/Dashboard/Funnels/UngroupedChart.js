import React, { useEffect, useState } from 'react';
import { generateUngroupedChartsData } from '../../CoreQuery/FunnelsResultPage/utils';
import Chart from '../../CoreQuery/FunnelsResultPage/UngroupedChart/Chart';
import FunnelsResultTable from '../../CoreQuery/FunnelsResultPage/FunnelsResultTable';

function UngroupedChart({
  resultState, queries, title, chartType, eventsMapper
}) {
  const [chartData, setChartData] = useState([]);

  useEffect(() => {
    const formattedData = generateUngroupedChartsData(resultState.data, queries);
    setChartData(formattedData);
  }, [queries, resultState.data]);

  if (!chartData.length) {
    return null;
  }

  let chartContent = null;

  if (chartType === 'barchart') {
    chartContent = (
      <div className="mt-4">
        <Chart
          title={title}
          chartData={chartData}
        />
      </div>
    );
  } else {
    chartContent = (
      <div className="mt-4">
        <FunnelsResultTable
          chartData={chartData}
          breakdown={[]}
          queries={queries}
          groups={[]}
          eventsMapper={eventsMapper}
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

export default UngroupedChart;
