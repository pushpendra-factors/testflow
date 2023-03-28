import React, { useState, useEffect, useCallback, useMemo } from 'react';
import cx from 'classnames';
import {
  formatData,
  formatDataInSeriesFormat,
  defaultSortProp,
  getVisibleData,
  getVisibleSeriesData
} from '../../CoreQuery/EventsAnalytics/SingleEventSingleBreakdown/utils';
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
import NoDataChart from '../../../components/NoDataChart';
import SingleEventSingleBreakdownHorizontalBarChart from '../../CoreQuery/EventsAnalytics/SingleEventSingleBreakdown/SingleEventSingleBreakdownHorizontalBarChart';
import ColumnChart from 'Components/ColumnChart';
import { CHART_COLOR_1 } from '../../../constants/color.constants';

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
    setSorter((currentSorter) => getNewSorterState(currentSorter, prop));
  }, []);

  const handleDateSorting = useCallback((prop) => {
    setDateSorter((currentSorter) => getNewSorterState(currentSorter, prop));
  }, []);

  useEffect(() => {
    const aggData = formatData(resultState.data);
    const { categories: cats, data: d } = isSeriesChart(chartType)
      ? formatDataInSeriesFormat(
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

  const columnCategories = useMemo(
    () => visibleProperties.map((v) => v.label),
    [visibleProperties]
  );

  const columnSeries = useMemo(() => {
    const series = [
      {
        data: visibleProperties.map((v) => v.value),
        color: CHART_COLOR_1
      }
    ];
    return series;
  }, [visibleProperties]);

  if (!visibleProperties.length) {
    return (
      <div className='flex justify-center items-center w-full h-full pt-4 pb-4'>
        <NoDataChart />
      </div>
    );
  }

  let chartContent = null;

  if (chartType === CHART_TYPE_BARCHART) {
    chartContent = (
      <ColumnChart
        categories={columnCategories}
        multiColored
        height={DASHBOARD_WIDGET_BAR_CHART_HEIGHT}
        cardSize={unit.cardSize}
        chartId={`events${unit.id}`}
        series={columnSeries}
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
        legendsPosition='top'
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
        legendsPosition='top'
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
        legendsPosition='top'
        cardSize={unit.cardSize}
        chartId={`line-${unit.id}`}
      />
    );
  } else if (chartType === CHART_TYPE_HORIZONTAL_BAR_CHART) {
    chartContent = (
      <SingleEventSingleBreakdownHorizontalBarChart
        aggregateData={aggregateData}
        breakdown={resultState.data.meta.query.gbp}
        isDashboardWidget
        cardSize={unit.cardSize}
      />
    );
  }

  return (
    <div
      className={cx('w-full flex-1', {
        'px-2': chartType !== CHART_TYPE_TABLE
      })}
    >
      {chartContent}
    </div>
  );
}

export default SingleEventSingleBreakdown;
