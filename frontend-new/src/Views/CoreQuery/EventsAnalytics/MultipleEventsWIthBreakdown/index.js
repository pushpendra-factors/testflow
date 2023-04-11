import React, {
  useState,
  useEffect,
  useCallback,
  forwardRef,
  useContext,
  useImperativeHandle
} from 'react';
import { useSelector } from 'react-redux';

import BarChart from 'Components/BarChart';
import LineChart from 'Components/HCLineChart';
import StackedAreaChart from 'Components/StackedAreaChart';
import StackedBarChart from 'Components/StackedBarChart';
import PivotTable from 'Components/PivotTable';

import { generateColors, getNewSorterState } from 'Utils/dataFormatter';
import {
  DASHBOARD_MODAL,
  CHART_TYPE_BARCHART,
  CHART_TYPE_STACKED_AREA,
  CHART_TYPE_STACKED_BAR,
  CHART_TYPE_PIVOT_CHART,
  QUERY_TYPE_EVENT,
  CHART_TYPE_METRIC_CHART
} from 'Utils/constants';

import {
  formatData,
  formatDataInStackedAreaFormat,
  defaultSortProp,
  getVisibleData,
  getVisibleSeriesData
} from './utils';
import MultipleEventsWithBreakdownTable from './MultipleEventsWithBreakdownTable';

import { CoreQueryContext } from '../../../../contexts/CoreQueryContext';
import MetricChart from 'Components/MetricChart/MetricChart';

const MultipleEventsWithBreakdown = forwardRef(
  (
    {
      queries,
      breakdown,
      resultState,
      page,
      chartType,
      durationObj,
      title,
      section
    },
    ref
  ) => {
    const {
      coreQueryState: { savedQuerySettings }
    } = useContext(CoreQueryContext);
    const [visibleProperties, setVisibleProperties] = useState([]);
    const [visibleSeriesData, setVisibleSeriesData] = useState([]);
    const { eventNames } = useSelector((state) => state.coreQuery);
    const [sorter, setSorter] = useState(
      savedQuerySettings.sorter && Array.isArray(savedQuerySettings.sorter)
        ? savedQuerySettings.sorter
        : defaultSortProp({ breakdown })
    );
    const [dateSorter, setDateSorter] = useState(
      savedQuerySettings.dateSorter &&
        Array.isArray(savedQuerySettings.dateSorter)
        ? savedQuerySettings.dateSorter
        : defaultSortProp({ breakdown })
    );
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

    useImperativeHandle(ref, () => {
      return {
        currentSorter: { sorter, dateSorter }
      };
    });

    useEffect(() => {
      const appliedColors = generateColors(queries.length);
      const aggData = formatData(
        resultState.data,
        queries,
        appliedColors,
        eventNames
      );
      const { categories: cats, data: d } = formatDataInStackedAreaFormat(
        resultState.data,
        aggData,
        eventNames,
        durationObj.frequency
      );
      setAggregateData(aggData);
      setCategories(cats);
      setData(d);
    }, [resultState.data, queries, eventNames, durationObj.frequency]);

    useEffect(() => {
      setVisibleProperties(getVisibleData(aggregateData, sorter));
    }, [aggregateData, sorter]);

    useEffect(() => {
      setVisibleSeriesData(getVisibleSeriesData(data, dateSorter));
    }, [data, dateSorter]);

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
          sorter={sorter}
          handleSorting={handleSorting}
          dateSorter={dateSorter}
          handleDateSorting={handleDateSorting}
          visibleSeriesData={visibleSeriesData}
          setVisibleSeriesData={setVisibleSeriesData}
        />
      </div>
    );

    if (chartType === CHART_TYPE_BARCHART) {
      chart = (
        <BarChart
          section={section}
          chartData={visibleProperties}
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
            showAllLegends={true}
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
            showAllLegends={true}
          />
        </div>
      );
    } else if (chartType === CHART_TYPE_PIVOT_CHART) {
      chart = (
        <div className='w-full'>
          <PivotTable
            data={aggregateData}
            breakdown={breakdown}
            metrics={queries}
            queryType={QUERY_TYPE_EVENT}
          />
        </div>
      );
    } else if (chartType === CHART_TYPE_METRIC_CHART) {
      console.log(visibleSeriesData);
      chart = (
        <div className='grid grid-cols-3 w-full col-gap-2 row-gap-12'>
          {visibleSeriesData &&
            visibleSeriesData.map((eachSeriesData) => {
              return (
                <MetricChart
                  key={eachSeriesData.name}
                  headerTitle={eachSeriesData.name}
                  value={eachSeriesData.value}
                />
              );
            })}
        </div>
      );
    } else {
      chart = (
        <div className='w-full'>
          <LineChart
            frequency={durationObj.frequency}
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
);

export default MultipleEventsWithBreakdown;
