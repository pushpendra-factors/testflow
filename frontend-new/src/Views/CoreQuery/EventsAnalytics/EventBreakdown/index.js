import React from 'react';
import { Spin } from 'antd';
import EventBreakdownCharts from './EventBreakdownCharts';
import BreakdownType from '../BreakdownType';
import DurationInfo from '../../../../components/DurationInfo';
import NoDataChart from 'Components/NoDataChart';

function EventBreakdown({
  breakdown, data, breakdownType, handleBreakdownTypeChange, durationObj, handleDurationChange, isWidgetModal = false
}) {
  if (data.loading) {
    return (
      <div className="flex justify-center items-center w-full h-64">
        <Spin size="large" />
      </div>
    );
  }

  if (data.error) {
    return (
      <div className="flex justify-center items-center w-full h-full pt-4 pb-4">
        <NoDataChart />
      </div>
    );
  }

  if (!data[breakdownType]) {
    return null;
  }

  let durationContent = (
		<div></div>
  );

  if (!isWidgetModal) {
    durationContent = (
			<div className="flex items-center filters-info">
				<div className="mr-1">Data from </div>
				<DurationInfo
					durationObj={durationObj}
					handleDurationChange={handleDurationChange}
				/>
				{breakdown.length ? (
					<div className="ml-1">shown as top 5 groups</div>
				) : null}
			</div>
    );
  }

  return (
    <div className="total-events w-full">
      <div className="flex items-center justify-end">
        {/* {durationContent} */}
        <div className="flex justify-end">
          <div className="px-4">
            <BreakdownType
              breakdown={breakdown}
              breakdownType={breakdownType}
              handleBreakdownTypeChange={handleBreakdownTypeChange}
            />
          </div>
        </div>
      </div>
      <EventBreakdownCharts
        data={data[breakdownType]}
        breakdown={breakdown}
        breakdownType={breakdownType}
        handleBreakdownTypeChange={handleBreakdownTypeChange}
      />
    </div>
  );
}

export default EventBreakdown;
