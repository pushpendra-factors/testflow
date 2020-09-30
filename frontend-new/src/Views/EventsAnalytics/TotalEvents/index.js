import React, { useState } from 'react';
import TotalEventsTable from './TotalEventsTable';
import { getSingleEventAnalyticsData, getDataInLineChartFormat, getMultiEventsAnalyticsData } from '../utils';
import { generateColors } from '../../CoreQuery/FunnelsResultPage/utils';
import ChartTypeDropdown from '../../../components/ChartTypeDropdown';
import SparkLineChart from '../../../components/SparkLineChart';
import LineChart from '../../../components/LineChart';

function TotalEvents({ queries, eventsMapper, reverseEventsMapper }) {
  const [hiddenEvents, setHiddenEvents] = useState([]);
  const appliedColors = generateColors(queries.length);
  // const [chartType, setChartType] = useState('sparklines');
  const [chartType, setChartType] = useState('linechart');

  let chartsData;
  if (queries.length === 1) {
    chartsData = getSingleEventAnalyticsData(queries[0], eventsMapper);
  } else {
    chartsData = getMultiEventsAnalyticsData(queries, eventsMapper);
  }

  if (!chartsData.length) {
    return null;
  }

  const menuItems = [
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
      {chartType === 'sparklines' ? (
        <SparkLineChart
          queries={queries}
          chartsData={chartsData}
          parentClass="flex justify-center items-center flex-wrap mt-8"
          appliedColors={appliedColors}
          eventsMapper={eventsMapper}
        />
      ) : (
          <div className="flex mt-8">
            <LineChart
              chartData={getDataInLineChartFormat(chartsData, queries, eventsMapper, hiddenEvents)}
              appliedColors={appliedColors} queries={queries}
              reverseEventsMapper={reverseEventsMapper}
              eventsMapper={eventsMapper}
              setHiddenEvents={setHiddenEvents}
              hiddenEvents={hiddenEvents}
            />
          </div>
      )}
      <div className="mt-8">
        <TotalEventsTable
          data={chartsData}
          events={queries}
          reverseEventsMapper={reverseEventsMapper}
          chartType={chartType}
          setHiddenEvents={setHiddenEvents}
          hiddenEvents={hiddenEvents}
        />
      </div>
    </div>
  );
}

export default TotalEvents;
