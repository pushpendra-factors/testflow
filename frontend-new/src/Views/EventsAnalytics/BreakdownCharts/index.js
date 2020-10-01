import React, { useState } from 'react';
import ChartTypeDropdown from '../../../components/ChartTypeDropdown';
// import TotalEventsTable from '../TotalEvents/TotalEventsTable';
import { singleEventSinglePropertyDateTimeResponse } from '../SampleResponse';
import { formatSingleEventSinglePropertyData } from '../utils';
import BarChart from '../../../components/BarChart';

function BreakdownCharts({
  queries, eventsMapper, reverseEventsMapper, breakdown
}) {
  console.log(queries, eventsMapper, reverseEventsMapper);
  // const [hiddenEvents, setHiddenEvents] = useState([]);
  const [chartType, setChartType] = useState('barchart');

  let chartsData = [];
  if (breakdown.length === 1) {
    chartsData = formatSingleEventSinglePropertyData(singleEventSinglePropertyDateTimeResponse);
  }

  if (!chartsData.length) {
    return null;
  }

  const menuItems = [
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

  let chartContent = null;

  if (chartType === 'barchart') {
    chartContent = (
      <div className="flex mt-8">
        <BarChart
          chartData={[...chartsData.slice(0, 7)]}
        />
      </div>
    );
  }

  return (
    <div className="total-events">
      <div className="flex items-center justify-between">
        <div className="filters-info">

        </div>
        <div className="user-actions">
          <ChartTypeDropdown
            chartType={chartType}
            menuItems={menuItems}
            onClick={(item) => {
              setChartType(item.key);
            }}
          />
        </div>
      </div>
      {chartContent}
      {/* <div className="mt-8">
        <TotalEventsTable
          data={chartsData}
          events={queries}
          reverseEventsMapper={reverseEventsMapper}
          chartType={chartType}
          setHiddenEvents={setHiddenEvents}
          hiddenEvents={hiddenEvents}
        />
      </div> */}
    </div>
  );
}

export default BreakdownCharts;
