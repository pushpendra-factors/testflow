import React, { useState, useEffect, useMemo } from 'react';
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
} from '../../../utils/constants';
import NoDataChart from '../../../components/NoDataChart';
import StackedAreaChart from '../../../components/StackedAreaChart';
import StackedBarChart from '../../../components/StackedBarChart';

function BreakdownCharts({
  arrayMapper,
  chartType,
  breakdown,
  data,
  isWidgetModal,
  setwidgetModal,
  unit,
  section,
}) {
  const [chartsData, setChartsData] = useState([]);
  const currentEventIndex = 0;
  const [visibleProperties, setVisibleProperties] = useState([]);
  const maxAllowedVisibleProperties = 5;

  useEffect(() => {
    const formattedData = formatData(
      data,
      arrayMapper,
      breakdown,
      currentEventIndex
    );
    setVisibleProperties(formattedData.slice(0, maxAllowedVisibleProperties));
    setChartsData(formattedData);
  }, [
    data,
    arrayMapper,
    currentEventIndex,
    breakdown,
    maxAllowedVisibleProperties,
  ]);

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
      currentEventIndex,
      visibleProperties
    );
  }, [
    data.result_group,
    arrayMapper,
    currentEventIndex,
    visibleProperties,
    chartType,
  ]);

  if (!chartsData.length) {
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
        onClick={() => setwidgetModal({ unit, data })}
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
        data={highchartsData}
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
        data={highchartsData}
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
        data={highchartsData}
        height={DASHBOARD_WIDGET_AREA_CHART_HEIGHT}
        legendsPosition='top'
        cardSize={unit.cardSize}
        chartId={`line-${unit.id}`}
      />
    );
  } else {
    chartContent = (
      <BreakdownTable
        currentEventIndex={currentEventIndex}
        chartType={chartType}
        chartsData={chartsData}
        breakdown={breakdown}
        arrayMapper={arrayMapper}
        isWidgetModal={isWidgetModal}
        responseData={data}
        visibleProperties={visibleProperties}
        maxAllowedVisibleProperties={maxAllowedVisibleProperties}
        setVisibleProperties={setVisibleProperties}
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
