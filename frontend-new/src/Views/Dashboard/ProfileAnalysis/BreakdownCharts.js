import React, { useState, useEffect, useCallback, useContext } from 'react';
import {
  formatData,
  defaultSortProp,
  getVisibleData,
} from '../../CoreQuery/ProfilesResultPage/BreakdownCharts/utils';
import BarChart from '../../../components/BarChart';
import BreakdownTable from '../../CoreQuery/ProfilesResultPage/BreakdownCharts/BreakdownTable';
import NoDataChart from '../../../components/NoDataChart';
import { getNewSorterState } from '../../../utils/dataFormatter';
import HorizontalBarChartTable from '../../CoreQuery/ProfilesResultPage/BreakdownCharts/HorizontalBarChartTable';
import {
  CHART_TYPE_BARCHART,
  CHART_TYPE_HORIZONTAL_BAR_CHART,
  CHART_TYPE_TABLE,
  DASHBOARD_WIDGET_BAR_CHART_HEIGHT,
} from '../../../utils/constants';
import { DashboardContext } from '../../../contexts/DashboardContext';

const BreakdownCharts = ({
  chartType,
  breakdown,
  data,
  unit,
  currentEventIndex,
  section,
  queries,
}) => {
  const { handleEditQuery } = useContext(DashboardContext);
  const [sorter, setSorter] = useState(defaultSortProp());
  const [visibleProperties, setVisibleProperties] = useState([]);
  const [aggregateData, setAggregateData] = useState([]);

  const handleSorting = useCallback((prop) => {
    setSorter((currentSorter) => {
      return getNewSorterState(currentSorter, prop);
    });
  }, []);

  useEffect(() => {
    const aggData = formatData(data, breakdown, queries, currentEventIndex);
    setAggregateData(aggData);
  }, [data, breakdown, queries, currentEventIndex]);

  useEffect(() => {
    setVisibleProperties(getVisibleData(aggregateData, sorter));
  }, [aggregateData, sorter]);

  if (!aggregateData.length) {
    return (
      <div className='flex justify-center items-center w-full h-full'>
        <NoDataChart />
      </div>
    );
  }

  let chartContent = null;
  let tableContent = null;

  if (chartType === CHART_TYPE_TABLE) {
    tableContent = (
      <div
        onClick={handleEditQuery}
        style={{ color: '#5949BC' }}
        className='mt-3 font-medium text-base cursor-pointer flex justify-end item-center'
      >
        Show More &rarr;
      </div>
    );
  }

  if (chartType === CHART_TYPE_BARCHART) {
    chartContent = (
      <BarChart
        chartData={visibleProperties}
        height={DASHBOARD_WIDGET_BAR_CHART_HEIGHT}
        section={section}
        title={unit.id}
        cardSize={unit.cardSize}
      />
    );
  }

  if (chartType === CHART_TYPE_TABLE) {
    chartContent = (
      <BreakdownTable
        aggregateData={aggregateData}
        sorter={sorter}
        breakdown={breakdown}
        currentEventIndex={currentEventIndex}
        chartType={chartType}
        sorter={sorter}
        handleSorting={handleSorting}
        visibleProperties={visibleProperties}
        isWidgetModal={false}
        setVisibleProperties={setVisibleProperties}
        section={section}
        queries={queries}
      />
    );
  }

  if (chartType === CHART_TYPE_HORIZONTAL_BAR_CHART) {
    chartContent = (
      <HorizontalBarChartTable
        aggregateData={aggregateData}
        breakdown={breakdown}
        cardSize={unit.cardSize}
        isDashboardWidget={true}
      />
    );
  }

  return (
    <div className={`w-full px-6 flex flex-1 flex-col justify-center`}>
      {chartContent}
      {tableContent}
    </div>
  );
};

export default BreakdownCharts;
