import React, { useEffect, useState } from 'react';
import { generateUngroupedChartsData } from '../utils';

import Chart from './Chart';
import FunnelsResultTable from '../FunnelsResultTable';

function UngroupedChart({
  resultState, queries, isWidgetModal, arrayMapper
}) {
  const [chartData, setChartData] = useState([]);

  useEffect(() => {
    const formattedData = generateUngroupedChartsData(resultState.data, arrayMapper);
    setChartData(formattedData);
  }, [arrayMapper, resultState.data]);

  if (!chartData.length) {
    return null;
  }

  return (
    <>
      <Chart
        chartData={chartData}
        arrayMapper={arrayMapper}
      />

      <div className="mt-8">
        <FunnelsResultTable
          isWidgetModal={isWidgetModal}
          chartData={chartData}
          breakdown={[]}
          queries={queries}
          groups={[]}
          arrayMapper={arrayMapper}
        />
      </div>
    </>
  );
}

export default UngroupedChart;
