import React, { useState, useEffect } from 'react';
import { formatData, formatVisibleProperties, formatDataInLineChartFormat } from '../../EventsAnalytics/MultipleEventsWIthBreakdown/utils';
import { generateColors } from '../../CoreQuery/FunnelsResultPage/utils';
import BarChart from '../../../components/BarChart';
import MultipleEventsWithBreakdownTable from '../../EventsAnalytics/MultipleEventsWIthBreakdown/MultipleEventsWithBreakdownTable';
import LineChart from '../../../components/LineChart';
// import BreakdownType from '../BreakdownType';

function MultipleEventsWithBreakdown({
  queries, resultState, page, chartType, title, breakdown, unit
}) {
  const [chartsData, setChartsData] = useState([]);
  const [visibleProperties, setVisibleProperties] = useState([]);
  const [hiddenProperties, setHiddenProperties] = useState([]);

  const maxAllowedVisibleProperties = unit.cardSize ? 5 : 3;

  useEffect(() => {
    const appliedColors = generateColors(queries.length);
    const formattedData = formatData(resultState.data, queries, appliedColors);
    setChartsData(formattedData);
    setVisibleProperties([...formattedData.slice(0, maxAllowedVisibleProperties)]);
  }, [resultState.data, queries, maxAllowedVisibleProperties]);

  if (!chartsData.length) {
    return null;
  }

  const mapper = {};
  const reverseMapper = {};

  const visibleLabels = visibleProperties.map(v => `${v.event},${v.label}`);

  visibleLabels.forEach((q, index) => {
    mapper[`${q}`] = `event${index + 1}`;
    reverseMapper[`event${index + 1}`] = q;
  });

  let chartContent = null;

  const lineChartData = formatDataInLineChartFormat(visibleProperties, mapper, hiddenProperties);
  const appliedColors = generateColors(visibleProperties.length);

  if (chartType === 'barchart') {
    chartContent = (
      <div className="flex mt-4">
        <BarChart
          chartData={formatVisibleProperties(visibleProperties, queries)}
          title={title}
          queries={queries}
        />
      </div>
    );
  } else if (chartType === 'table') {
    chartContent = (
      <div className="mt-4">
        <MultipleEventsWithBreakdownTable
          data={chartsData}
          lineChartData={lineChartData}
          queries={queries}
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
    );
  } else {
    chartContent = (
      <div className="flex mt-4">
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
    <div className="total-events">
      {chartContent}
    </div>
  );
}

export default MultipleEventsWithBreakdown;
