import React, { useEffect, useState } from 'react';
// import styles from './index.module.scss';
import SpikeChart from './SpikeChart';
import TotalEventsTable from './TotalEventsTable';
import { getSpikeChartData } from '../utils';
import EventHeader from '../EventHeader';

function TotalEvents({ queries }) {
  const [chartData, setChartData] = useState([]);

  useEffect(() => {
    setChartData(getSpikeChartData([queries[0]]));
  }, [queries]);

  let total = 0;

  chartData.forEach(elem => {
    total += elem[queries[0]];
  });

  if (!chartData.length) {
    return null;
  }

  return (
        <div className="total-events">
            <div className="flex justify-center items-center mt-4">
                <div className="w-1/4">
                    <EventHeader bgColor="#4D7DB4" query={queries[0]} total={total} />
                </div>
                <div className="w-3/4">
                    <SpikeChart event={queries[0]} page="totalEvents" chartData={chartData} chartColor="#4D7DB4" />
                </div>
            </div>
            <div className="mt-8">
                <TotalEventsTable
                    data={chartData}
                    events={[queries[0]]}
                />
            </div>
        </div>
  );
}

export default TotalEvents;
