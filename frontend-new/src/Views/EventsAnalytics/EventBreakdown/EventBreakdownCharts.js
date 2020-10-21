import React, { useState, useEffect } from 'react';
import { Radio } from 'antd';
import { formatData } from './utils';
import BarChart from '../../../components/BarChart';
// import EventBreakdownTable from './EventBreakdownTable';

function EventBreakdownCharts({
  data, breakdownType, handleBreakdownTypeChange
}) {
  const [chartsData, setChartsData] = useState([]);
  const [visibleProperties, setVisibleProperties] = useState([]);
  const maxAllowedVisibleProperties = 5;

  useEffect(() => {
    const formattedData = formatData(data);
    setChartsData(formattedData);
    setVisibleProperties([...formattedData.slice(0, maxAllowedVisibleProperties)]);
  }, [data]);

  if (!chartsData.length) {
    return null;
  }

  return (
        <div className="total-events">
            <div className="flex items-center justify-between">
                <div className="filters-info w-1/2">

                </div>
                <div className="user-actions w-1/2 flex justify-end">
                    <div className="px-4">
                        <Radio.Group value={breakdownType} onChange={handleBreakdownTypeChange}>
                            <Radio.Button value="each">Each Event</Radio.Button>
                            <Radio.Button disabled value="any">Any Event</Radio.Button>
                            <Radio.Button value="all">All Events</Radio.Button>
                        </Radio.Group>
                    </div>
                </div>
            </div>
            <div className="flex mt-8">
                <BarChart
                    chartData={visibleProperties}
                />
            </div>
            {/* {chartContent} */}
            {/* <div className="mt-8">
                <EventBreakdownTable
                    data={chartsData}
                    // lineChartData={lineChartData}
                    // queries={queries}
                    breakdown={breakdown}
                    // events={queries}
                    // chartType={chartType}
                    setVisibleProperties={setVisibleProperties}
                    visibleProperties={visibleProperties}
                    maxAllowedVisibleProperties={maxAllowedVisibleProperties}
                    // originalData={resultState.data}
                    // page={page}
                />
            </div> */}
        </div>
  );
}

export default EventBreakdownCharts;
