import React, { useState } from 'react';
import TotalEventsTable from './TotalEventsTable';
import { getSingleEventAnalyticsData, getDataInLineChartFormat, getMultiEventsAnalyticsData } from '../utils';
import { generateColors } from '../../CoreQuery/FunnelsResultPage/utils';
import ChartTypeDropdown from '../../../components/ChartTypeDropdown';
import SparkLineChart from '../../../components/SparkLineChart';
import LineChart from '../../../components/LineChart';


function TotalEvents({ queries }) {
  const appliedColors = generateColors(queries.length);
  const [chartType, setChartType] = useState('linechart');

  const eventsMapper = {};
  const reverseEventsMapper = {};
  queries.forEach((q, index) => {
    eventsMapper[`${q}`] = `event${index}`;
    reverseEventsMapper[`event${index}`] = q;
  })

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
      name: 'Sparkline',
    },
    {
      key: 'linechart',
      onClick: setChartType,
      name: 'Line Chart',
    }
  ]

  const sparkLinesJsx = (
    <SparkLineChart
      queries={queries}
      chartsData={chartsData}
      parentClass="flex justify-center items-center flex-wrap mt-8"
      appliedColors={appliedColors}
      eventsMapper={eventsMapper}
    />
  )

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
        <div>{sparkLinesJsx}</div>
      ) : (
          <div className="flex mt-8">
            <LineChart chartData={getDataInLineChartFormat(chartsData, queries, eventsMapper)} appliedColors={appliedColors} queries={queries} />
          </div>
        )}
      <div className="mt-8">
        <TotalEventsTable
          data={chartsData}
          events={queries}
          reverseEventsMapper={reverseEventsMapper}
        />
      </div>
    </div>
  );
}

export default TotalEvents;
