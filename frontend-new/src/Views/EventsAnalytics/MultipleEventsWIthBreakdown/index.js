import React, { useState, useEffect } from 'react';
import { formatData, formatVisibleProperties, formatDataInLineChartFormat } from './utils';
import { generateColors } from '../../CoreQuery/FunnelsResultPage/utils';
import BarChart from '../../../components/BarChart';
import ChartTypeDropdown from '../../../components/ChartTypeDropdown';
import MultipleEventsWithBreakdownTable from './MultipleEventsWithBreakdownTable';
import LineChart from '../../../components/LineChart';
import BreakdownType from '../BreakdownType';

function MultipleEventsWithBreakdown({
  queries, breakdown, resultState, page, breakdownType, handleBreakdownTypeChange
}) {
  const [chartsData, setChartsData] = useState([]);
  const [visibleProperties, setVisibleProperties] = useState([]);
  const [chartType, setChartType] = useState('barchart');
  const [hiddenProperties, setHiddenProperties] = useState([]);

  const maxAllowedVisibleProperties = 5;

  useEffect(() => {
    const appliedColors = generateColors(queries.length);
    const formattedData = formatData(resultState.data, queries, appliedColors);
    setChartsData(formattedData);
    setVisibleProperties([...formattedData.slice(0, maxAllowedVisibleProperties)]);
  }, [resultState.data, queries]);

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

  const lineChartData = formatDataInLineChartFormat(visibleProperties, mapper, hiddenProperties);
  const appliedColors = generateColors(visibleProperties.length);

  if (chartType === 'barchart') {
    chartContent = (
      <div className="flex mt-8">
        <BarChart
          chartData={formatVisibleProperties(visibleProperties, queries)}
          queries={queries}
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
    <div className="total-events w-full">
      <div className="flex items-center justify-between">
        <div className="filters-info w-1/2">

        </div>
        <div className="user-actions w-1/2 flex justify-end">

          {page === 'totalUsers' ? (
            <div className="px-4">
              <BreakdownType
                breakdown={breakdown}
                breakdownType={breakdownType}
                handleBreakdownTypeChange={handleBreakdownTypeChange}
              />
            </div>
          ) : null}

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
    </div>
  );
}

export default MultipleEventsWithBreakdown;
