import React, {
  useState, useEffect, useContext, useCallback
} from 'react';
import {
  formatData,
  formatDataInStackedAreaFormat,
  defaultSortProp,
  getVisibleData
} from '../../CoreQuery/EventsAnalytics/SingleEventMultipleBreakdown/utils';
import BarChart from '../../../components/BarChart';
import LineChart from '../../../components/HCLineChart';
import SingleEventMultipleBreakdownTable from '../../CoreQuery/EventsAnalytics/SingleEventMultipleBreakdown/SingleEventMultipleBreakdownTable';
import { getNewSorterState, isSeriesChart } from '../../../utils/dataFormatter';
import {
  CHART_TYPE_TABLE,
  CHART_TYPE_BARCHART,
  DASHBOARD_WIDGET_BAR_CHART_HEIGHT,
  CHART_TYPE_STACKED_AREA,
  DASHBOARD_WIDGET_AREA_CHART_HEIGHT,
  CHART_TYPE_STACKED_BAR,
  CHART_TYPE_LINECHART,
  CHART_TYPE_HORIZONTAL_BAR_CHART,
  CHART_TYPE_PIVOT_CHART
} from '../../../utils/constants';
import StackedAreaChart from '../../../components/StackedAreaChart';
import StackedBarChart from '../../../components/StackedBarChart';
import { DashboardContext } from '../../../contexts/DashboardContext';
import NoDataChart from '../../../components/NoDataChart';
import SingleEventMultipleBreakdownHorizontalBarChart from '../../CoreQuery/EventsAnalytics/SingleEventMultipleBreakdown/SingleEventMultipleBreakdownHorizontalBarChart';

function SingleEventMultipleBreakdown({
  resultState,
  page,
  chartType,
  breakdown,
  queries,
  unit,
  durationObj,
  section
}) {
  const [visibleProperties, setVisibleProperties] = useState([]);
  const [sorter, setSorter] = useState(defaultSortProp({ breakdown }));
  const [visibleSeriesData, setVisibleSeriesData] = useState([]);
  const [dateSorter, setDateSorter] = useState(defaultSortProp({ breakdown }));
  const [aggregateData, setAggregateData] = useState([]);
  const [categories, setCategories] = useState([]);
  const [data, setData] = useState([]);
  const { handleEditQuery } = useContext(DashboardContext);

  const handleSorting = useCallback((prop) => {
    setSorter((currentSorter) => {
      return getNewSorterState(currentSorter, prop);
    });
  }, []);

  const handleDateSorting = useCallback((prop) => {
    setDateSorter((currentSorter) => {
      return getNewSorterState(currentSorter, prop);
    });
  }, []);

  useEffect(() => {
    const aggData = formatData(resultState.data);
    const { categories: cats, data: d } = isSeriesChart(chartType)
      ? formatDataInStackedAreaFormat(
        resultState.data,
        aggData,
        durationObj.frequency
      )
      : { categories: [], data: [] };
    setAggregateData(aggData);
    setCategories(cats);
    setData(d);
  }, [resultState.data, chartType, durationObj.frequency]);

  useEffect(() => {
    setVisibleProperties(getVisibleData(aggregateData, sorter));
  }, [aggregateData, sorter]);

  useEffect(() => {
    setVisibleSeriesData(getVisibleData(data, dateSorter));
  }, [data, dateSorter]);

  if (!visibleProperties.length) {
    return (
      <div className="flex justify-center items-center w-full h-full pt-4 pb-4">
        <NoDataChart />
      </div>
    );
  }

  let chartContent = null;

  let tableContent = null;

  // if (chartType === CHART_TYPE_TABLE || chartType === CHART_TYPE_PIVOT_CHART) {
  //   tableContent = (
  //     <div
  //       onClick={handleEditQuery}
  //       style={{ color: '#5949BC' }}
  //       className="mt-3 font-medium text-base cursor-pointer flex justify-end item-center"
  //     >
  //       Show More &rarr;
  //     </div>
  //   );
  // }

  if (chartType === CHART_TYPE_BARCHART) {
    chartContent = (
      <div className="flex mt-4">
        <BarChart
          chartData={visibleProperties}
          height={DASHBOARD_WIDGET_BAR_CHART_HEIGHT}
          title={unit.id}
          cardSize={unit.cardSize}
          section={section}
          queries={queries}
        />
      </div>
    );
  } else if (
    chartType === CHART_TYPE_TABLE ||
    chartType === CHART_TYPE_PIVOT_CHART
  ) {
    chartContent = (
      <SingleEventMultipleBreakdownTable
        data={aggregateData}
        seriesData={data}
        breakdown={breakdown}
        events={queries}
        chartType={chartType}
        setVisibleProperties={setVisibleProperties}
        visibleProperties={visibleProperties}
        page={page}
        durationObj={durationObj}
        categories={categories}
        section={section}
        sorter={sorter}
        handleSorting={handleSorting}
        dateSorter={dateSorter}
        handleDateSorting={handleDateSorting}
        visibleSeriesData={visibleSeriesData}
        setVisibleSeriesData={setVisibleSeriesData}
      />
    );
  } else if (chartType === CHART_TYPE_STACKED_AREA) {
    chartContent = (
      <StackedAreaChart
        frequency={durationObj.frequency}
        categories={categories}
        data={visibleSeriesData}
        height={DASHBOARD_WIDGET_AREA_CHART_HEIGHT}
        legendsPosition="top"
        cardSize={unit.cardSize}
        chartId={`area-${unit.id}`}
      />
    );
  } else if (chartType === CHART_TYPE_STACKED_BAR) {
    chartContent = (
      <StackedBarChart
        frequency={durationObj.frequency}
        categories={categories}
        data={visibleSeriesData}
        height={DASHBOARD_WIDGET_AREA_CHART_HEIGHT}
        legendsPosition="top"
        cardSize={unit.cardSize}
        chartId={`bar-${unit.id}`}
      />
    );
  } else if (chartType === CHART_TYPE_LINECHART) {
    chartContent = (
      <LineChart
        frequency={durationObj.frequency}
        categories={categories}
        data={visibleSeriesData}
        height={DASHBOARD_WIDGET_AREA_CHART_HEIGHT}
        legendsPosition="top"
        cardSize={unit.cardSize}
        chartId={`line-${unit.id}`}
      />
    );
  } else if (chartType === CHART_TYPE_HORIZONTAL_BAR_CHART) {
    chartContent = (
      <SingleEventMultipleBreakdownHorizontalBarChart
        aggregateData={aggregateData}
        breakdown={resultState.data.meta.query.gbp}
        isDashboardWidget={true}
        cardSize={unit.cardSize}
      />
    );
  }

  return (
    <div className={'w-full'}>
      {chartContent}
      {tableContent}
    </div>
  );
}

export default SingleEventMultipleBreakdown;
