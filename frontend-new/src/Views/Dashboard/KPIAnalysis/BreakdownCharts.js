import React, {
  useState,
  useEffect,
  useCallback,
  useContext,
  memo
} from 'react';
import {
  formatData,
  getVisibleData,
  formatDataInSeriesFormat,
  getVisibleSeriesData,
  getDefaultSortProp
} from '../../CoreQuery/KPIAnalysis/BreakdownCharts/utils';
import { getNewSorterState, isSeriesChart } from '../../../utils/dataFormatter';
import NoDataChart from '../../../components/NoDataChart';
import {
  CHART_TYPE_BARCHART,
  CHART_TYPE_HORIZONTAL_BAR_CHART,
  CHART_TYPE_LINECHART,
  CHART_TYPE_STACKED_AREA,
  CHART_TYPE_STACKED_BAR,
  CHART_TYPE_PIVOT_CHART
  ,
  CHART_TYPE_TABLE,
  DASHBOARD_WIDGET_BAR_CHART_HEIGHT,
  DASHBOARD_WIDGET_AREA_CHART_HEIGHT
} from '../../../utils/constants';
import LineChart from '../../../components/HCLineChart';
import BarChart from '../../../components/BarChart';
import BreakdownTable from '../../CoreQuery/KPIAnalysis/BreakdownCharts/BreakdownTable';
import HorizontalBarChartTable from '../../CoreQuery/KPIAnalysis/BreakdownCharts/HorizontalBarChartTable';
import StackedAreaChart from '../../../components/StackedAreaChart';
import StackedBarChart from '../../../components/StackedBarChart';
import { DashboardContext } from '../../../contexts/DashboardContext';

const BreakdownCharts = ({
  breakdown,
  kpis,
  responseData,
  chartType,
  section,
  currentEventIndex,
  unit,
  durationObj
}) => {
  const [visibleProperties, setVisibleProperties] = useState([]);
  const [visibleSeriesData, setVisibleSeriesData] = useState([]);
  const [sorter, setSorter] = useState(getDefaultSortProp({ kpis, breakdown }));
  const [dateSorter, setDateSorter] = useState(getDefaultSortProp({ kpis, breakdown }));
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
    const aggData = formatData(responseData, kpis, breakdown, currentEventIndex);
    const { categories: cats, data: d } = isSeriesChart(chartType)
      ? formatDataInSeriesFormat(
        responseData,
        aggData,
        currentEventIndex,
        'date',
        breakdown
      )
      : { categories: [], data: [] };
    setAggregateData(aggData);
    setCategories(cats);
    setData(d);
  }, [responseData, breakdown, currentEventIndex]);

  useEffect(() => {
    setVisibleProperties(getVisibleData(aggregateData, sorter));
  }, [aggregateData, sorter]);

  useEffect(() => {
    setVisibleSeriesData(getVisibleSeriesData(data, dateSorter));
  }, [data, dateSorter]);

  if (!aggregateData.length) {
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

  if (chartType === CHART_TYPE_TABLE || chartType === CHART_TYPE_PIVOT_CHART) {
    chartContent = (
      <BreakdownTable
        kpis={kpis}
        data={aggregateData}
        seriesData={data}
        section={section}
        breakdown={breakdown}
        chartType={chartType}
        setVisibleProperties={setVisibleProperties}
        visibleProperties={visibleProperties}
        // durationObj={durationObj}
        categories={categories}
        sorter={sorter}
        handleSorting={handleSorting}
        dateSorter={dateSorter}
        handleDateSorting={handleDateSorting}
        visibleSeriesData={visibleSeriesData}
        setVisibleSeriesData={setVisibleSeriesData}
      />
    );
  }

  if (chartType === CHART_TYPE_BARCHART) {
    chartContent = (
      <BarChart
        section={section}
        height={DASHBOARD_WIDGET_BAR_CHART_HEIGHT}
        title={unit.id}
        chartData={visibleProperties}
        cardSize={unit.cardSize}
      />
    );
  } else if (chartType === CHART_TYPE_HORIZONTAL_BAR_CHART) {
    chartContent = (
      <HorizontalBarChartTable
        breakdown={breakdown}
        aggregateData={aggregateData}
        cardSize={unit.cardSize}
        isDashboardWidget={true}
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
  } else if (chartType === CHART_TYPE_STACKED_AREA) {
    chartContent = (
      <StackedAreaChart
        frequency={durationObj.frequency}
        categories={categories}
        data={visibleSeriesData}
        height={DASHBOARD_WIDGET_AREA_CHART_HEIGHT}
        legendsPosition="top"
        cardSize={unit.cardSize}
        chartId={`bar-${unit.id}`}
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
  }

  return (
    <div className={'w-full'}>
      {chartContent}
      {tableContent}
    </div>
  );
};

export default memo(BreakdownCharts);
