import React, { useState, useEffect, useMemo } from 'react';
import {
  formatData,
  formatVisibleProperties,
  formatDataInStackedAreaFormat,
} from './utils';
import BarChart from '../../../../components/BarChart';
import MultipleEventsWithBreakdownTable from './MultipleEventsWithBreakdownTable';
import LineChart from '../../../../components/HCLineChart';
import { generateColors } from '../../../../utils/dataFormatter';
import {
  DASHBOARD_MODAL,
  CHART_TYPE_BARCHART,
  CHART_TYPE_STACKED_AREA,
  CHART_TYPE_STACKED_BAR,
  MAX_ALLOWED_VISIBLE_PROPERTIES,
} from '../../../../utils/constants';
import StackedAreaChart from '../../../../components/StackedAreaChart';
import StackedBarChart from '../../../../components/StackedBarChart';
import { useSelector } from 'react-redux';

function MultipleEventsWithBreakdown({
  queries,
  breakdown,
  resultState,
  page,
  chartType,
  durationObj,
  title,
  section,
}) {
  const [visibleProperties, setVisibleProperties] = useState([]);
  const { eventNames } = useSelector((state) => state.coreQuery);

  const aggregateData = useMemo(() => {
    const appliedColors = generateColors(queries.length);
    return formatData(resultState.data, queries, appliedColors, eventNames);
  }, [resultState.data, queries, eventNames]);
  
  const { categories, data } = useMemo(() => {
    return formatDataInStackedAreaFormat(resultState.data, aggregateData, eventNames);
  }, [resultState.data, aggregateData, eventNames]);

  const visibleSeriesData = useMemo(() => {
    const colors = generateColors(visibleProperties.length);
    return data
      .filter(
        (elem) =>
          visibleProperties.findIndex((vp) => vp.index === elem.index) > -1
      )
      .map((elem, index) => {
        const color = colors[index];
        return {
          ...elem,
          color,
        };
      });
  }, [data, visibleProperties]);

  useEffect(() => {
    setVisibleProperties([
      ...aggregateData.slice(0, MAX_ALLOWED_VISIBLE_PROPERTIES),
    ]);
  }, [aggregateData]);

  if (!visibleProperties.length) {
    return null;
  }

  let chart = null;

  const table = (
    <div className='mt-12 w-full'>
      <MultipleEventsWithBreakdownTable
        isWidgetModal={section === DASHBOARD_MODAL}
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
      />
    </div>
  );

  if (chartType === CHART_TYPE_BARCHART) {
    chart = (
      <BarChart
        section={section}
        chartData={formatVisibleProperties(visibleProperties, queries)}
        queries={queries}
        title={title}
      />
    );
  } else if (chartType === CHART_TYPE_STACKED_AREA) {
    chart = (
      <div className='w-full'>
        <StackedAreaChart
          frequency={durationObj.frequency}
          categories={categories}
          data={visibleSeriesData}
        />
      </div>
    );
  } else if (chartType === CHART_TYPE_STACKED_BAR) {
    chart = (
      <div className='w-full'>
        <StackedBarChart
          frequency={durationObj.frequency}
          categories={categories}
          data={visibleSeriesData}
        />
      </div>
    );
  } else {
    chart = (
      <div className='w-full'>
        <LineChart
          frequency={durationObj.frequency}
          categories={categories}
          data={visibleSeriesData}
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

export default MultipleEventsWithBreakdown;
