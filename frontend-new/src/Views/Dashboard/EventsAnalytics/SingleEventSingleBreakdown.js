import React, {
  useState, useEffect, useContext, useCallback
} from 'react';
import {
  formatData,
  formatDataInStackedAreaFormat,
  defaultSortProp,
  getVisibleData,
  getVisibleSeriesData
} from '../../CoreQuery/EventsAnalytics/SingleEventSingleBreakdown/utils';
import BarChart from '../../../components/BarChart';
import SingleEventSingleBreakdownTable from '../../CoreQuery/EventsAnalytics/SingleEventSingleBreakdown/SingleEventSingleBreakdownTable';
import LineChart from '../../../components/HCLineChart';
import StackedAreaChart from '../../../components/StackedAreaChart';
import { getNewSorterState, isSeriesChart } from '../../../utils/dataFormatter';
import {
  CHART_TYPE_TABLE,
  DASHBOARD_WIDGET_BAR_CHART_HEIGHT,
  CHART_TYPE_BARCHART,
  CHART_TYPE_STACKED_AREA,
  DASHBOARD_WIDGET_AREA_CHART_HEIGHT,
  CHART_TYPE_STACKED_BAR,
  CHART_TYPE_LINECHART,
  CHART_TYPE_HORIZONTAL_BAR_CHART
} from '../../../utils/constants';
import StackedBarChart from '../../../components/StackedBarChart';
import { DashboardContext } from '../../../contexts/DashboardContext';
import NoDataChart from '../../../components/NoDataChart';
import SingleEventSingleBreakdownHorizontalBarChart from '../../CoreQuery/EventsAnalytics/SingleEventSingleBreakdown/SingleEventSingleBreakdownHorizontalBarChart';

function SingleEventSingleBreakdown({
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
  const [visibleSeriesData, setVisibleSeriesData] = useState([]);
  const [sorter, setSorter] = useState(defaultSortProp({ breakdown }));
  const [dateSorter, setDateSorter] = useState(defaultSortProp({ breakdown }));
  const [aggregateData, setAggregateData] = useState([]);
  const [categories, setCategories] = useState([]);
  const [data, setData] = useState([]);

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

  const { handleEditQuery } = useContext(DashboardContext);

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
  }, [resultState.data, durationObj.frequency, chartType]);

  useEffect(() => {
    setVisibleSeriesData(getVisibleSeriesData(data, dateSorter));
  }, [data, dateSorter]);

  useEffect(() => {
    setVisibleProperties(getVisibleData(aggregateData, sorter));
  }, [aggregateData, sorter]);

  if (!visibleProperties.length) {
    return (
      <div className="flex justify-center items-center w-full h-full">
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
        className="mt-3 font-medium text-base cursor-pointer flex justify-end item-center"
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
        title={unit.id}
        cardSize={unit.cardSize}
        section={section}
        queries={queries}
      />
    );
  } else if (chartType === CHART_TYPE_TABLE) {
    chartContent = (
      <SingleEventSingleBreakdownTable
        data={aggregateData}
        seriesData={data}
        breakdown={breakdown}
        events={queries}
        chartType={chartType}
        page={page}
        setVisibleProperties={setVisibleProperties}
        visibleProperties={visibleProperties}
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
      <SingleEventSingleBreakdownHorizontalBarChart
        aggregateData={aggregateData}
        breakdown={resultState.data.meta.query.gbp}
        isDashboardWidget={true}
        cardSize={unit.cardSize}
      />
    );
  }

  return (
    <div className={'w-full px-6 flex flex-1 flex-col justify-center'}>
      {chartContent}
      {tableContent}
    </div>
  );
}

export default SingleEventSingleBreakdown;
