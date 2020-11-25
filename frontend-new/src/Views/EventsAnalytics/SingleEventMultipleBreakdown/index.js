import React, { useState, useEffect } from 'react';
import {
  formatData, formatDataInLineChartFormat
} from './utils';
import BarChart from '../../../components/BarChart';
import LineChart from '../../../components/LineChart';
import SingleEventMultipleBreakdownTable from './SingleEventMultipleBreakdownTable';
import { generateColors } from '../../CoreQuery/FunnelsResultPage/utils';

function SingleEventMultipleBreakdown({
  queries, breakdown, resultState, page, chartType
}) {
  const [chartsData, setChartsData] = useState([]);
  const [visibleProperties, setVisibleProperties] = useState([]);
  const [hiddenProperties, setHiddenProperties] = useState([]);

  const maxAllowedVisibleProperties = 5;

  useEffect(() => {
    const formattedData = formatData(resultState.data);
    setChartsData(formattedData);
    setVisibleProperties([...formattedData.slice(0, maxAllowedVisibleProperties)]);
  }, [resultState.data]);

  if (!chartsData.length) {
    return null;
  }

  const mapper = {};
  const reverseMapper = {};

  const visibleLabels = visibleProperties.map(v => v.label);

  visibleLabels.forEach((q, index) => {
    mapper[`${q}`] = `event${index + 1}`;
    reverseMapper[`event${index + 1}`] = q;
  });

  const lineChartData = formatDataInLineChartFormat(resultState.data, visibleProperties, mapper, hiddenProperties);

  const appliedColors = generateColors(visibleProperties.length);

  let chartContent = null;

  if (chartType === 'barchart') {
    chartContent = (
      <div className="flex mt-8">
        <BarChart
          chartData={visibleProperties}
        />
      </div>
    );
  } else {
    chartContent = (
      <div className="flex mt-8">
        <LineChart
          chartData={lineChartData}
          appliedColors={appliedColors}
          queries={visibleLabels}
          reverseEventsMapper={reverseMapper}
          eventsMapper={mapper}
          setHiddenEvents={setHiddenProperties}
          hiddenEvents={hiddenProperties}
          isDecimalAllowed={page === 'activeUsers' || page === 'frequency'}
        />
      </div>
    );
  }

  return (
    <div className="w-full">
      {chartContent}
      <div className="mt-8">
        <SingleEventMultipleBreakdownTable
          data={chartsData}
          lineChartData={lineChartData}
          breakdown={breakdown}
          events={queries}
          chartType={chartType}
          setVisibleProperties={setVisibleProperties}
          visibleProperties={visibleProperties}
          maxAllowedVisibleProperties={maxAllowedVisibleProperties}
          originalData={resultState.data}
          page={page}
        />
      </div>
    </div>
  );
}

export default SingleEventMultipleBreakdown;
