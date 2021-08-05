import React, { useState, useEffect, useMemo, useCallback } from 'react';
import {
  formatData,
  formatDataInStackedAreaFormat,
  defaultSortProp,
} from './utils';
import BarChart from '../../../../components/BarChart';
import LineChart from '../../../../components/HCLineChart';
import SingleEventMultipleBreakdownTable from './SingleEventMultipleBreakdownTable';
import {
  generateColors,
  getNewSorterState,
} from '../../../../utils/dataFormatter';
import {
  DASHBOARD_MODAL,
  CHART_TYPE_STACKED_AREA,
  CHART_TYPE_BARCHART,
  CHART_TYPE_STACKED_BAR,
  MAX_ALLOWED_VISIBLE_PROPERTIES,
} from '../../../../utils/constants';
import StackedAreaChart from '../../../../components/StackedAreaChart';
import StackedBarChart from '../../../../components/StackedBarChart';

function SingleEventMultipleBreakdown({
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
  const [sorter, setSorter] = useState(defaultSortProp());
  const [dateSorter, setDateSorter] = useState({});
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

  useEffect(() => {
    const aggData = formatData(resultState.data);
    const { categories: cats, data: d } = formatDataInStackedAreaFormat(
      resultState.data,
      aggData
    );
    setAggregateData(aggData);
    setCategories(cats);
    setData(d);
  }, [resultState.data]);

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
      <SingleEventMultipleBreakdownTable
        isWidgetModal={section === DASHBOARD_MODAL}
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
        sorter={sorter}
        handleSorting={handleSorting}
        dateSorter={dateSorter}
        handleDateSorting={handleDateSorting}
      />
    </div>
  );

  if (chartType === CHART_TYPE_BARCHART) {
    chart = (
      <BarChart section={section} title={title} chartData={visibleProperties} />
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

export default SingleEventMultipleBreakdown;
