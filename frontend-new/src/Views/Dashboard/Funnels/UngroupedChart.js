import React, { useEffect, useState } from 'react';
import { generateUngroupedChartsData } from '../../CoreQuery/FunnelsResultPage/utils';
import Chart from '../../CoreQuery/FunnelsResultPage/UngroupedChart/Chart';
// import FunnelsResultTable from '../FunnelsResultTable';

function UngroupedChart({
  resultState, queries, title
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

            <div>

                <Chart
                    title={title}
                    chartData={chartData}
                />

                {/* <div className="mt-8">
          <FunnelsResultTable
            chartData={chartData}
            breakdown={[]}
            queries={queries}
            groups={[]}
            eventsMapper={eventsMapper}
          />
        </div> */}
            </div>
    </>
  );
}

export default UngroupedChart;
