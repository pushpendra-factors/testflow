import React, { useState, useEffect, useMemo, useContext } from 'react';
import {
  formatData,
  formatDataInHighChartsFormat,
} from '../../CoreQuery/CampaignAnalytics/BreakdownCharts/utils';
import BarChart from '../../../components/BarChart';
import BreakdownTable from '../../CoreQuery/CampaignAnalytics/BreakdownCharts/BreakdownTable';
import LineChart from '../../../components/HCLineChart';
import {
  CHART_TYPE_BARCHART,
  CHART_TYPE_LINECHART,
  CHART_TYPE_TABLE,
  DASHBOARD_WIDGET_BAR_CHART_HEIGHT,
  CHART_TYPE_STACKED_AREA,
  CHART_TYPE_STACKED_BAR,
  DASHBOARD_WIDGET_AREA_CHART_HEIGHT,
  MAX_ALLOWED_VISIBLE_PROPERTIES,
} from '../../../utils/constants';
import NoDataChart from '../../../components/NoDataChart';
import StackedAreaChart from '../../../components/StackedAreaChart';
import StackedBarChart from '../../../components/StackedBarChart';
import { generateColors, SortData } from '../../../utils/dataFormatter';
import { DashboardContext } from '../../../contexts/DashboardContext';

function BreakdownCharts({
  arrayMapper,
  chartType,
  breakdown,
  data,
  isWidgetModal,
  unit,
  section,
}) {
  const [visibleProperties, setVisibleProperties] = useState([]);
  const { handleEditQuery } = useContext(DashboardContext);
  const currentEventIndex = 0;

  const aggregateData = useMemo(() => {
    return formatData(data, arrayMapper, breakdown);
  }, [data, breakdown, arrayMapper]);

  const chartData = useMemo(() => {
    const colors = generateColors(1);
    const currEventName = arrayMapper.find(
      (elem) => elem.index === currentEventIndex
    ).eventName;
    const result = aggregateData.map((d) => {
      return {
        ...d,
        color: colors[0],
        value: d[currEventName],
      };
    });
    return SortData(result, currEventName, 'descend');
  }, [currentEventIndex, arrayMapper, aggregateData]);

  const { categories, highchartsData } = useMemo(() => {
    if (chartType === CHART_TYPE_BARCHART || chartType === CHART_TYPE_TABLE) {
      return {
        categories: [],
        highchartsData: [],
      };
    }
    return formatDataInHighChartsFormat(
      data.result_group[0],
      arrayMapper,
      aggregateData
    );
  }, [data.result_group, arrayMapper, aggregateData, chartType]);

  const visibleSeriesData = useMemo(() => {
    const colors = generateColors(visibleProperties.length);
    const currEventName = arrayMapper.find(
      (elem) => elem.index === currentEventIndex
    ).eventName;
    return highchartsData
      .filter(
        (elem) =>
          visibleProperties.findIndex((vp) => vp.index === elem.index) > -1
      )
      .map((elem, index) => {
        const color = colors[index];
        return {
          ...elem,
          data: elem[currEventName],
          color,
        };
      });
  }, [highchartsData, visibleProperties, arrayMapper, currentEventIndex]);

  useEffect(() => {
    setVisibleProperties([
      ...chartData.slice(0, MAX_ALLOWED_VISIBLE_PROPERTIES),
    ]);
  }, [chartData]);

  if (!chartData.length) {
    return (
      <div className='mt-4 flex justify-center items-center w-full h-64 '>
        <NoDataChart />
      </div>
    );
  }

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

  let chartContent = null;

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
  } else if (chartType === CHART_TYPE_STACKED_AREA) {
    chartContent = (
      <StackedAreaChart
        frequency='date'
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
        frequency='date'
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
        frequency='date'
        categories={categories}
        data={visibleSeriesData}
        height={DASHBOARD_WIDGET_AREA_CHART_HEIGHT}
        legendsPosition='top'
        cardSize={unit.cardSize}
        chartId={`line-${unit.id}`}
      />
    );
  } else {
    chartContent = (
      <BreakdownTable
        chartsData={chartData}
        seriesData={highchartsData}
        categories={categories}
        breakdown={breakdown}
        currentEventIndex={currentEventIndex}
        chartType={chartType}
        arrayMapper={arrayMapper}
        isWidgetModal={isWidgetModal}
        visibleProperties={visibleProperties}
        setVisibleProperties={setVisibleProperties}
        section={section}
      />
    );
  }

  return (
    <div className={`w-full px-6 flex flex-1 flex-col  justify-center`}>
      {chartContent}
      {tableContent}
    </div>
  );
}

export default BreakdownCharts;
