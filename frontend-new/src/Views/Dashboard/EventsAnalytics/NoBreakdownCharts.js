import React, { useState } from 'react';
import { formatSingleEventAnalyticsData, formatMultiEventsAnalyticsData, getDataInLineChartFormat } from '../../EventsAnalytics/NoBreakdownCharts/utils';
import NoBreakdownTable from '../../EventsAnalytics/NoBreakdownCharts/NoBreakdownTable';
import SparkLineChart from '../../../components/SparkLineChart';
import LineChart from '../../../components/LineChart';
import { generateColors } from '../../CoreQuery/FunnelsResultPage/utils';

function NoBreakdownCharts({
  queries, eventsMapper, reverseEventsMapper, resultState, page, chartType
}) {
  const [hiddenEvents, setHiddenEvents] = useState([]);
  const appliedColors = generateColors(queries.length);

  let chartsData = [];
  if (queries.length === 1) {
    chartsData = formatSingleEventAnalyticsData(resultState.data, queries[0], eventsMapper);
  } else {
    chartsData = formatMultiEventsAnalyticsData(resultState.data, queries, eventsMapper);
  }

  if (!chartsData.length) {
    return null;
  }

  let chartContent = null;

  if (chartType === 'sparklines') {
    chartContent = (
      <div className="mt-4">
        <SparkLineChart
          queries={queries}
          chartsData={chartsData}
          parentClass="flex items-center flex-wrap mt-4 justify-center"
          appliedColors={appliedColors}
          eventsMapper={eventsMapper}
          page={page}
          resultState={resultState}
        />
      </div>
    );
  } else if (chartType === 'table') {
    chartContent = (
      <div className="mt-4">
        <NoBreakdownTable
          data={chartsData}
          events={queries}
          reverseEventsMapper={reverseEventsMapper}
          chartType={chartType}
          setHiddenEvents={setHiddenEvents}
          hiddenEvents={hiddenEvents}
        />
      </div>
    );
  } else {
    chartContent = (
      <div className="flex mt-4">
        <LineChart
          chartData={getDataInLineChartFormat(chartsData, queries, eventsMapper, hiddenEvents)}
          appliedColors={appliedColors}
          queries={queries}
          reverseEventsMapper={reverseEventsMapper}
          eventsMapper={eventsMapper}
          setHiddenEvents={setHiddenEvents}
          hiddenEvents={hiddenEvents}
          isDecimalAllowed={page === 'activeUsers' || page === 'frequency'}
        />
      </div>
    );
  }

  return (
    <div className="total-events w-full">
      {chartContent}
    </div>
  );
}

export default NoBreakdownCharts;
