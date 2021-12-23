import React, { useState, useEffect, useMemo } from 'react';
import { formatData, formatDataInHighChartsFormat } from './utils';
import {
  CHART_TYPE_BARCHART,
  CHART_TYPE_LINECHART,
  DASHBOARD_MODAL,
  CHART_TYPE_STACKED_AREA,
  CHART_TYPE_STACKED_BAR,
  MAX_ALLOWED_VISIBLE_PROPERTIES,
} from '../../../../utils/constants';
import LineChart from '../../../../components/HCLineChart';
import BarChart from '../../../../components/BarChart';
import BreakdownTable from './BreakdownTable';
import NoDataChart from '../../../../components/NoDataChart';
import StackedAreaChart from '../../../../components/StackedAreaChart';
import StackedBarChart from '../../../../components/StackedBarChart';
import { generateColors, SortData } from '../../../../utils/dataFormatter';

function BreakdownCharts({
  arrayMapper,
  chartType,
  breakdown,
  data,
  title = 'chart',
  currentEventIndex,
  section,
}) {
  const [visibleProperties, setVisibleProperties] = useState([]);
  const [aggregateData, setAggregateData] = useState([]);
  const [categories, setCategories] = useState([]);
  const [highchartsData, setHighchartsData] = useState([]);

  useEffect(() => {
    const aggData = formatData(data, arrayMapper, breakdown);
    
    //jitesh please check, i've changed it to "data.result ? data.result[0] : data.result_group[0]"
    
    const {
      categories: cat,
      highchartsData: hcd,
    } = formatDataInHighChartsFormat(
      data.result ? data.result[0] : data.result_group[0],
      arrayMapper,
      aggData
    );
    setAggregateData(aggData);
    setCategories(cat);
    setHighchartsData(hcd);
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

  const table = (
    <div className='mt-12 w-full'>
      <BreakdownTable
        chartsData={chartData}
        seriesData={highchartsData}
        categories={categories}
        breakdown={breakdown}
        currentEventIndex={currentEventIndex}
        chartType={chartType}
        arrayMapper={arrayMapper}
        isWidgetModal={section === DASHBOARD_MODAL}
        visibleProperties={visibleProperties}
        setVisibleProperties={setVisibleProperties}
      />
    </div>
  );

  let chart = null;

  console.log(visibleProperties);

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
          data={visibleSeriesData}
          showAllLegends={true}
        />
      </div>
    );
  } else if (chartType === CHART_TYPE_STACKED_BAR) {
    chart = (
      <div className='w-full'>
        <StackedBarChart
          frequency='date'
          categories={categories}
          data={visibleSeriesData}
          showAllLegends={true}
        />
      </div>
    );
  } else if (chartType === CHART_TYPE_LINECHART) {
    chart = (
      <div className='w-full'>
        <LineChart
          frequency='date'
          categories={categories}
          data={visibleSeriesData}
          showAllLegends={true}
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
