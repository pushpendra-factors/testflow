import React, { useState, useEffect } from 'react';
import { formatData } from './utils';
import BarChart from '../../../components/BarChart';
import EventBreakdownTable from './EventBreakdownTable';
import BreakdownType from '../BreakdownType';
import ChartHeader from '../../../components/SparkLineChart/ChartHeader';

function EventBreakdownCharts({
  data, breakdownType, handleBreakdownTypeChange, breakdown
}) {
  console.log(data);
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

  let content = null;

  if (breakdown.length) {
    content = (
            <div className="flex mt-8">
                <BarChart
                    chartData={visibleProperties}
                />
            </div>
    );
  } else {
    content = (
            <div className="flex mt-8 justify-center">
                <ChartHeader total={data.rows[0]} query={'Count'} bgColor="#4D7DB4" />
            </div>
    );
  }

  return (
        <div className="total-events">
            <div className="flex items-center justify-between">
                <div className="filters-info w-1/2">

                </div>
                <div className="user-actions w-1/2 flex justify-end">
                    <div className="px-4">
                        <BreakdownType
                            breakdown={breakdown}
                            breakdownType={breakdownType}
                            handleBreakdownTypeChange={handleBreakdownTypeChange}
                        />
                    </div>
                </div>
            </div>
            {content}
            <div className="mt-8">
                <EventBreakdownTable
                    data={chartsData}
                    breakdown={breakdown}
                    setVisibleProperties={setVisibleProperties}
                    visibleProperties={visibleProperties}
                    maxAllowedVisibleProperties={maxAllowedVisibleProperties}
                />
            </div>
        </div>
  );
}

export default EventBreakdownCharts;
