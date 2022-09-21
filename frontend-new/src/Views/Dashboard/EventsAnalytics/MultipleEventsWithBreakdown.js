import React, {
  useState,
  useEffect,
  useMemo,
  useContext,
  useCallback
} from 'react';
import cx from 'classnames';
import {
  formatData,
  formatDataInStackedAreaFormat,
  defaultSortProp,
  getVisibleData,
  getVisibleSeriesData
} from '../../CoreQuery/EventsAnalytics/MultipleEventsWIthBreakdown/utils';
import BarChart from '../../../components/BarChart';
import MultipleEventsWithBreakdownTable from '../../CoreQuery/EventsAnalytics/MultipleEventsWIthBreakdown/MultipleEventsWithBreakdownTable';
import LineChart from '../../../components/HCLineChart';
import {
  generateColors,
  isSeriesChart,
  getNewSorterState
} from '../../../utils/dataFormatter';
import {
  CHART_TYPE_TABLE,
  CHART_TYPE_BARCHART,
  DASHBOARD_WIDGET_MULTICOLORED_BAR_CHART_HEIGHT,
  CHART_TYPE_STACKED_AREA,
  DASHBOARD_WIDGET_AREA_CHART_HEIGHT,
  CHART_TYPE_STACKED_BAR,
  CHART_TYPE_PIVOT_CHART
} from '../../../utils/constants';
import StackedAreaChart from '../../../components/StackedAreaChart';
import StackedBarChart from '../../../components/StackedBarChart';
import { useSelector } from 'react-redux';
import { DashboardContext } from '../../../contexts/DashboardContext';
import NoDataChart from '../../../components/NoDataChart';
// import BreakdownType from '../BreakdownType';

function MultipleEventsWithBreakdown({
  queries,
  resultState,
  page,
  chartType,
  breakdown,
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
  const { handleEditQuery } = useContext(DashboardContext);
  const { eventNames } = useSelector((state) => state.coreQuery);

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

  const appliedQueries = useMemo(() => {
    return queries.join(';');
  }, [queries]); // its a hack to prevent unwanted rerenders due to queries variable, needs to be optimized

  useEffect(() => {
    const appliedColors = generateColors(appliedQueries.split(';').length);
    const aggData = formatData(
      resultState.data,
      appliedQueries.split(';'),
      appliedColors,
      eventNames
    );
    const { categories: cats, data: d } = isSeriesChart(chartType)
      ? formatDataInStackedAreaFormat(
          resultState.data,
          aggData,
          eventNames,
          durationObj.frequency
        )
      : { categories: [], data: [] };
    setAggregateData(aggData);
    setCategories(cats);
    setData(d);
  }, [
    resultState.data,
    appliedQueries,
    eventNames,
    chartType,
    durationObj.frequency
  ]);

  useEffect(() => {
    setVisibleProperties(getVisibleData(aggregateData, sorter));
  }, [aggregateData, sorter]);

  useEffect(() => {
    setVisibleSeriesData(getVisibleSeriesData(data, dateSorter));
  }, [data, dateSorter]);

  if (!visibleProperties.length) {
    return (
      <div className="flex justify-center items-center w-full h-full pt-4 pb-4">
        <NoDataChart />
      </div>
    );
  }

  let chartContent = null;

  if (chartType === CHART_TYPE_BARCHART) {
    chartContent = (
      <BarChart
        chartData={visibleProperties}
        height={DASHBOARD_WIDGET_MULTICOLORED_BAR_CHART_HEIGHT}
        title={unit.id}
        cardSize={unit.cardSize}
        section={section}
        queries={queries}
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
  } else if (
    chartType === CHART_TYPE_TABLE ||
    chartType === CHART_TYPE_PIVOT_CHART
  ) {
    chartContent = (
      <MultipleEventsWithBreakdownTable
        data={aggregateData}
        seriesData={data}
        queries={queries}
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
  } else {
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
  }

  return (
    <div
      className={cx('w-full flex-1', {
        'p-2': chartType !== CHART_TYPE_TABLE
      })}
    >
      {chartContent}
    </div>
  );
}

export default MultipleEventsWithBreakdown;
