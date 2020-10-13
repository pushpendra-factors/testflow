import React, { useState, useEffect } from 'react';
import {
  formatData, formatDataInLineChartFormat, formatUserData, formatUserDataInLineChartFormat
} from './utils';
import BarChart from '../../../components/BarChart';
import LineChart from '../../../components/LineChart';
import ChartTypeDropdown from '../../../components/ChartTypeDropdown';
import SingleEventMultipleBreakdownTable from './SingleEventMultipleBreakdownTable';
import { generateColors } from '../../CoreQuery/FunnelsResultPage/utils';

function SingleEventMultipleBreakdown({
  queries, breakdown, resultState, page
}) {
  const [chartsData, setChartsData] = useState([]);
  const [visibleProperties, setVisibleProperties] = useState([]);
  const [chartType, setChartType] = useState('barchart');
  const [hiddenProperties, setHiddenProperties] = useState([]);

  const maxAllowedVisibleProperties = 5;

  useEffect(() => {
    let formattedData;
    if (page === 'totalEvents') {
      formattedData = formatData(resultState.data.result_group[0]);
    } else {
      formattedData = formatUserData(resultState.data.result_group[0]);
    }

    setChartsData(formattedData);
    setVisibleProperties([...formattedData.slice(0, maxAllowedVisibleProperties)]);
  }, [resultState.data.result_group]);

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

  let lineChartData;

  if (page === 'totalEvents') {
    lineChartData = formatDataInLineChartFormat(resultState.data.result_group[0], visibleProperties, mapper, hiddenProperties);
  } else {
    lineChartData = formatUserDataInLineChartFormat(resultState.data.result_group[0], visibleProperties, mapper, hiddenProperties);
  }

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
        <SingleEventMultipleBreakdownTable
          data={chartsData}
          lineChartData={lineChartData}
          breakdown={breakdown}
          events={queries}
          chartType={chartType}
          setVisibleProperties={setVisibleProperties}
          visibleProperties={visibleProperties}
          maxAllowedVisibleProperties={maxAllowedVisibleProperties}
          originalData={resultState.data.result_group[0]}
          page={page}
        />
      </div>
    </div>
  );
}

export default SingleEventMultipleBreakdown;
