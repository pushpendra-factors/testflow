import React, { useState } from 'react';
import { formatSingleEventAnalyticsData, formatMultiEventsAnalyticsData, getDataInLineChartFormat } from './utils';
import NoBreakdownTable from './NoBreakdownTable';
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
      <SparkLineChart
        queries={queries}
        chartsData={chartsData}
        parentClass="flex items-center flex-wrap mt-4 justify-center"
        appliedColors={appliedColors}
        eventsMapper={eventsMapper}
        page={page}
        resultState={resultState}
      />
    );
  } else if (chartType === 'linechart') {
    chartContent = (
      <div className="flex mt-8">
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
    <div className="w-full">
      {chartContent}
      <div className="mt-8">
        <NoBreakdownTable
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

export default NoBreakdownCharts;
