import React, { useState, useEffect, useCallback, memo, useMemo } from 'react';
import cx from 'classnames';
import {
  formatData,
  getVisibleData,
  formatDataInSeriesFormat,
  getVisibleSeriesData,
  getDefaultSortProp
} from '../../CoreQuery/KPIAnalysis/BreakdownCharts/utils';
import {
  generateColors,
  getNewSorterState,
  isSeriesChart
} from '../../../utils/dataFormatter';
import NoDataChart from '../../../components/NoDataChart';
import {
  CHART_TYPE_BARCHART,
  CHART_TYPE_HORIZONTAL_BAR_CHART,
  CHART_TYPE_LINECHART,
  CHART_TYPE_STACKED_AREA,
  CHART_TYPE_STACKED_BAR,
  CHART_TYPE_PIVOT_CHART,
  CHART_TYPE_TABLE,
  DASHBOARD_WIDGET_BAR_CHART_HEIGHT,
  DASHBOARD_WIDGET_AREA_CHART_HEIGHT,
  CHART_TYPE_METRIC_CHART,
  MAX_ALLOWED_VISIBLE_PROPERTIES
} from '../../../utils/constants';
import LineChart from '../../../components/HCLineChart';
import BreakdownTable from '../../CoreQuery/KPIAnalysis/BreakdownCharts/BreakdownTable';
import HorizontalBarChartTable from '../../CoreQuery/KPIAnalysis/BreakdownCharts/HorizontalBarChartTable';
import StackedAreaChart from '../../../components/StackedAreaChart';
import StackedBarChart from '../../../components/StackedBarChart';
import ColumnChart from 'Components/ColumnChart';
import { CHART_COLOR_1 } from '../../../constants/color.constants';
import { cardSizeToMetricCount } from 'Constants/charts.constants';
import MetricChart from 'Components/MetricChart/MetricChart';

const colors = generateColors(MAX_ALLOWED_VISIBLE_PROPERTIES);

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
  const [dateSorter, setDateSorter] = useState(
    getDefaultSortProp({ kpis, breakdown })
  );
  const [aggregateData, setAggregateData] = useState([]);
  const [categories, setCategories] = useState([]);
  const [data, setData] = useState([]);
  const [dateWiseTotals, setDateWiseTotals] = useState([]);

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
  
  const unitId = unit.id || unit.inter_id;

  const unitId = unit.id || unit.inter_id;

  useEffect(() => {
    const aggData = formatData(
      responseData,
      kpis,
      breakdown,
      currentEventIndex
    );
    const {
      categories: cats,
      data: d,
      dateWiseTotals: dwt
    } = isSeriesChart(chartType)
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
    setDateWiseTotals(dwt);
  }, [responseData, breakdown, currentEventIndex, kpis, chartType]);

  useEffect(() => {
    setVisibleProperties(getVisibleData(aggregateData, sorter));
  }, [aggregateData, sorter]);

  useEffect(() => {
    setVisibleSeriesData(getVisibleSeriesData(data, dateSorter));
  }, [data, dateSorter]);

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

  if (!aggregateData.length) {
    return (
      <div className='flex justify-center items-center w-full h-full pt-4 pb-4'>
        <NoDataChart />
      </div>
    );
  }

  let chartContent = null;

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
      <ColumnChart
        categories={columnCategories}
        multiColored
        height={DASHBOARD_WIDGET_BAR_CHART_HEIGHT}
        cardSize={unit.cardSize}
        chartId={`kpi${unitId}`}
        series={columnSeries}
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
        legendsPosition='top'
        cardSize={unit.cardSize}
        chartId={`line-${unitId}`}
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
        chartId={`bar-${unitId}`}
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
        chartId={`bar-${unitId}`}
        dateWiseTotals={dateWiseTotals}
      />
    );
  } else if (chartType === CHART_TYPE_METRIC_CHART) {
    chartContent = (
      <div className='flex justify-between w-full col-gap-2 h-full'>
        {aggregateData
          .slice(0, cardSizeToMetricCount[unit.cardSize])
          .map((eachAggregateData, index) => {
            return (
              <MetricChart
                key={eachAggregateData.label}
                headerTitle={eachAggregateData.label}
                value={eachAggregateData.value}
                iconColor={colors[index]}
              />
            );
          })}
      </div>
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
};

export default memo(BreakdownCharts);
