import React, { useState, useEffect, useMemo } from 'react';
import { formatData, formatDataInHighChartsFormat } from './utils';
import {
  CHART_TYPE_BARCHART,
  CHART_TYPE_LINECHART,
  DASHBOARD_MODAL,
  CHART_TYPE_STACKED_AREA,
  CHART_TYPE_STACKED_BAR,
} from '../../../../utils/constants';
import LineChart from '../../../../components/HCLineChart';
import BarChart from '../../../../components/BarChart';
import BreakdownTable from './BreakdownTable';
import NoDataChart from '../../../../components/NoDataChart';
import StackedAreaChart from '../../../../components/StackedAreaChart';
import StackedBarChart from '../../../../components/StackedBarChart';

function BreakdownCharts({
  arrayMapper,
  chartType,
  breakdown,
  data,
  title = 'chart',
  currentEventIndex,
  section,
}) {
  const [chartsData, setChartsData] = useState([]);
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
  }, [data, arrayMapper, currentEventIndex, breakdown]);

  const { categories, highchartsData } = useMemo(() => {
    return formatDataInHighChartsFormat(
      data.result_group[0],
      arrayMapper,
      currentEventIndex,
      visibleProperties
    );
  }, [data.result_group, arrayMapper, currentEventIndex, visibleProperties]);

  if (!chartsData.length) {
    return (
      <div className='mt-4 flex justify-center items-center w-full h-64 '>
        <NoDataChart />
      </div>
    );
  }

  const table = (
    <div className='mt-12 w-full'>
      <BreakdownTable
        currentEventIndex={currentEventIndex}
        chartType={chartType}
        chartsData={chartsData}
        breakdown={breakdown}
        arrayMapper={arrayMapper}
        isWidgetModal={section === DASHBOARD_MODAL}
        responseData={data}
        visibleProperties={visibleProperties}
        maxAllowedVisibleProperties={maxAllowedVisibleProperties}
        setVisibleProperties={setVisibleProperties}
      />
    </div>
  );

  let chart = null;

  if (chartType === CHART_TYPE_BARCHART) {
    chart = (
      <BarChart section={section} title={title} chartData={visibleProperties} />
    );
  } else if (chartType === CHART_TYPE_STACKED_AREA) {
    chart = (
      <div className='w-full'>
        <StackedAreaChart
          frequency='date'
          categories={categories}
          data={highchartsData}
        />
      </div>
    );
  } else if (chartType === CHART_TYPE_STACKED_BAR) {
    chart = (
      <div className='w-full'>
        <StackedBarChart
          frequency='date'
          categories={categories}
          data={highchartsData}
        />
      </div>
    );
  } else if (chartType === CHART_TYPE_LINECHART) {
    chart = (
      <div className='w-full'>
        <LineChart
          frequency='date'
          categories={categories}
          data={highchartsData}
        />
      </div>
    );
  }

  return (
    <div className='flex items-center justify-center flex-col'>
      {chart}
      {table}
    </div>
  );
}

export default BreakdownCharts;
