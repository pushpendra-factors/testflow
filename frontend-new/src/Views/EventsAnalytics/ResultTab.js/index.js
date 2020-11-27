import React, { useEffect, useState } from 'react';
import NoBreakdownCharts from '../NoBreakdownCharts';
import SingleEventSingleBreakdown from '../SingleEventSingleBreakdown';
import { Spin } from 'antd';
import SingleEventMultipleBreakdown from '../SingleEventMultipleBreakdown';
import MultipleEventsWithBreakdown from '../MultipleEventsWIthBreakdown';
import DurationInfo from '../../CoreQuery/DurationInfo';
import ChartTypeDropdown from '../../../components/ChartTypeDropdown';
import BreakdownType from '../BreakdownType';

function ResultTab({
  queries, eventsMapper, reverseEventsMapper, breakdown, resultState, page, index, breakdownType, handleBreakdownTypeChange, durationObj, handleDurationChange, isWidgetModal = false
}) {
  const [chartType, setChartType] = useState('');

  let menuItems;

  if (breakdown.length) {
    menuItems = [
      {
        key: 'barchart',
        onClick: setChartType,
        name: 'Barchart'
      },
      {
        key: 'linechart',
        onClick: setChartType,
        name: 'Line Chart'
      }
    ];
  } else {
    menuItems = [
      {
        key: 'sparklines',
        onClick: setChartType,
        name: 'Sparkline'
      },
      {
        key: 'linechart',
        onClick: setChartType,
        name: 'Line Chart'
      }
    ];
  }

  useEffect(() => {
    if (breakdown.length) {
      setChartType('barchart');
    } else {
      setChartType('sparklines');
    }
  }, [breakdown]);

  if (resultState[index].loading) {
    return (
			<div className="flex justify-center items-center w-full h-64">
				<Spin size="large" />
			</div>
    );
  }

  if (resultState[index].error) {
    return (
			<div className="flex justify-center items-center w-full h-64">
				Something went wrong!
			</div>
    );
  }

  let content = null;
  let breakdownTypeContent = null;

  if (resultState[index].data && resultState[index].data.metrics.rows.length) {
    if (!breakdown.length) {
      if (page === 'totalUsers' && queries.length > 1) {
        breakdownTypeContent = (
					<div className="px-4">
						<BreakdownType
							breakdown={breakdown}
							breakdownType={breakdownType}
							handleBreakdownTypeChange={handleBreakdownTypeChange}
						/>
					</div>
        );
      }
      content = (
				<NoBreakdownCharts
					queries={queries}
					eventsMapper={eventsMapper}
					reverseEventsMapper={reverseEventsMapper}
					resultState={resultState[index]}
					page={page}
					breakdownType={breakdownType}
					handleBreakdownTypeChange={handleBreakdownTypeChange}
					durationObj={durationObj}
					handleDurationChange={handleDurationChange}
					chartType={chartType}
					setChartType={setChartType}
				/>
      );
    }

    if (queries.length === 1 && breakdown.length === 1) {
      content = (
				<SingleEventSingleBreakdown
					queries={queries}
					breakdown={breakdown}
					resultState={resultState[index]}
					page={page}
					durationObj={durationObj}
					handleDurationChange={handleDurationChange}
					chartType={chartType}
					setChartType={setChartType}
				/>
      );
    }

    if (queries.length > 1 && breakdown.length) {
      if (page === 'totalUsers') {
        breakdownTypeContent = (
					<div className="px-4">
						<BreakdownType
							breakdown={breakdown}
							breakdownType={breakdownType}
							handleBreakdownTypeChange={handleBreakdownTypeChange}
						/>
					</div>
        );
      }
      content = (
				<MultipleEventsWithBreakdown
					queries={queries}
					breakdown={breakdown}
					resultState={resultState[index]}
					page={page}
					breakdownType={breakdownType}
					handleBreakdownTypeChange={handleBreakdownTypeChange}
					durationObj={durationObj}
					handleDurationChange={handleDurationChange}
					chartType={chartType}
					setChartType={setChartType}
				/>
      );
    }

    if (queries.length === 1 && breakdown.length > 1) {
      content = (
				<SingleEventMultipleBreakdown
					queries={queries}
					breakdown={breakdown}
					resultState={resultState[index]}
					page={page}
					durationObj={durationObj}
					handleDurationChange={handleDurationChange}
					chartType={chartType}
					setChartType={setChartType}
				/>
      );
    }
  }

  if (resultState[index].data && !resultState[index].data.metrics.rows.length) {
    content = (
			<div className="flex justify-center items-center h-64">
				No Data Found!
			</div>
    );
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
			<div className="flex items-center justify-between">
				{durationContent}
				<div className="flex items-center justify-end">
					{breakdownTypeContent}
					<ChartTypeDropdown
						chartType={chartType}
						menuItems={menuItems}
						onClick={(item) => {
						  setChartType(item.key);
						}}
					/>
				</div>
			</div>
			{content}
		</div >
  );
}

export default ResultTab;
