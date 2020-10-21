import React, { useState, useEffect } from 'react';
import ChartTypeDropdown from '../../../components/ChartTypeDropdown';
import {
  formatData, formatDataInLineChartFormat
} from './utils';
import BarChart from '../../../components/BarChart';
import SingleEventSingleBreakdownTable from './SingleEventSingleBreakdownTable';
import LineChart from '../../../components/LineChart';
import { generateColors } from '../../CoreQuery/FunnelsResultPage/utils';

function SingleEventSingleBreakdown({
  queries, breakdown, resultState, page
}) {
  const [chartsData, setChartsData] = useState([]);
  const [visibleProperties, setVisibleProperties] = useState([]);
  const [chartType, setChartType] = useState('barchart');
  const [hiddenProperties, setHiddenProperties] = useState([]);

  const maxAllowedVisibleProperties = 7;

  useEffect(() => {
    const formattedData = formatData(resultState.data);
    setChartsData(formattedData);
    setVisibleProperties([...formattedData.slice(0, maxAllowedVisibleProperties)]);
  }, [resultState.data]);

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
          isDecimalAllowed = {page === 'activeUsers' || page === 'frequency'}
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
      <div className="mt-8">
        <SingleEventSingleBreakdownTable
          data={chartsData}
          breakdown={breakdown}
          events={queries}
          chartType={chartType}
          page={page}
          setVisibleProperties={setVisibleProperties}
          visibleProperties={visibleProperties}
          maxAllowedVisibleProperties={maxAllowedVisibleProperties}
          lineChartData={lineChartData}
          originalData={resultState.data}
        />
      </div>
    </div>
  );
}

export default SingleEventSingleBreakdown;
