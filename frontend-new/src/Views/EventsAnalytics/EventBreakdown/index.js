import React from 'react';
import { Spin } from 'antd';
import EventBreakdownCharts from './EventBreakdownCharts';

function EventBreakdown({
  breakdown, data, breakdownType, handleBreakdownTypeChange
}) {
  console.log(breakdown);
  if (data.loading) {
    return (
			<div className="flex justify-center items-center w-full h-64">
				<Spin size="large" />
			</div>
    );
  }

  if (data.error) {
    return (
			<div className="flex justify-center items-center w-full h-64">
				Something went wrong!
			</div>
    );
  }

  if (!data[breakdownType]) {
    return null;
  }

  return (
		<EventBreakdownCharts
			data={data[breakdownType]}
			breakdown={breakdown}
			breakdownType={breakdownType}
			handleBreakdownTypeChange={handleBreakdownTypeChange}
		/>
  );
}

export default EventBreakdown;
