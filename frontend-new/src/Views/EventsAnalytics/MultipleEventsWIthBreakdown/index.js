import React, { useState, useEffect } from 'react';
import { formatData, formatUserData, formatVisibleProperties } from './utils';
import { generateColors } from '../../CoreQuery/FunnelsResultPage/utils';
import BarChart from '../../../components/BarChart';
import ChartTypeDropdown from '../../../components/ChartTypeDropdown';
import MultipleEventsWithBreakdownTable from './MultipleEventsWithBreakdownTable';

function MultipleEventsWithBreakdown({
  queries, breakdown, resultState, page
}) {
  const [chartsData, setChartsData] = useState([]);
  const [visibleProperties, setVisibleProperties] = useState([]);
  const [chartType, setChartType] = useState('barchart');
  const [hiddenProperties, setHiddenProperties] = useState([]);

  const maxAllowedVisibleProperties = 5;

  useEffect(() => {
    let formattedData;
    const appliedColors = generateColors(queries.length);
    if (page === 'totalEvents') {
      formattedData = formatData(resultState.data, queries, appliedColors);
    } else {
      formattedData = formatUserData(resultState.data, queries, appliedColors);
    }

    setChartsData(formattedData);
    setVisibleProperties([...formattedData.slice(0, maxAllowedVisibleProperties)]);
  }, [resultState.data, queries]);

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
                    chartData={formatVisibleProperties(visibleProperties, queries)}
                    queries={queries}
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
                <MultipleEventsWithBreakdownTable
                    data={chartsData}
                    // lineChartData={lineChartData}
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
