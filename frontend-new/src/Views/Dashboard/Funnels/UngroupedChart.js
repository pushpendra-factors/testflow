import React, { useEffect, useState } from 'react';
import { generateUngroupedChartsData } from '../../CoreQuery/FunnelsResultPage/utils';
import Chart from '../../CoreQuery/FunnelsResultPage/UngroupedChart/Chart';
import FunnelsResultTable from '../../CoreQuery/FunnelsResultPage/FunnelsResultTable';

function UngroupedChart({
  resultState, queries, title, chartType, eventsMapper, setwidgetModal, unit
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
          cardSize={unit.cardSize}
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

export default UngroupedChart;
