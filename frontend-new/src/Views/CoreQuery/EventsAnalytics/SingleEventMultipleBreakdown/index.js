import React, {
  useState,
  useEffect,
  useCallback,
  forwardRef,
  useImperativeHandle,
  useContext
} from 'react';

import BarChart from 'Components/BarChart';
import LineChart from 'Components/HCLineChart';
import StackedAreaChart from 'Components/StackedAreaChart';
import StackedBarChart from 'Components/StackedBarChart';
import PivotTable from 'Components/PivotTable';
import { getNewSorterState } from 'Utils/dataFormatter';
import {
  DASHBOARD_MODAL,
  CHART_TYPE_STACKED_AREA,
  CHART_TYPE_BARCHART,
  CHART_TYPE_STACKED_BAR,
  CHART_TYPE_LINECHART,
  CHART_TYPE_PIVOT_CHART,
  QUERY_TYPE_EVENT
} from 'Utils/constants';

import {
  formatData,
  formatDataInStackedAreaFormat,
  defaultSortProp,
  getVisibleData
} from './utils';
import SingleEventMultipleBreakdownTable from './SingleEventMultipleBreakdownTable';
import SingleEventMultipleBreakdownHorizontalBarChart from './SingleEventMultipleBreakdownHorizontalBarChart';

import { CoreQueryContext } from '../../../../contexts/CoreQueryContext';

const SingleEventMultipleBreakdown = forwardRef(
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
      const aggData = formatData(resultState.data);
      const { categories: cats, data: d } = formatDataInStackedAreaFormat(
        resultState.data,
        aggData,
        durationObj.frequency
      );
      setAggregateData(aggData);
      setCategories(cats);
      setData(d);
    }, [resultState.data, durationObj.frequency]);

    useEffect(() => {
      setVisibleProperties(getVisibleData(aggregateData, sorter));
    }, [aggregateData, sorter]);

    useEffect(() => {
      setVisibleSeriesData(getVisibleData(data, dateSorter));
    }, [data, dateSorter]);

    if (!visibleProperties.length) {
      return null;
    }

    let chart = null;
    const table = (
      <div className="mt-12 w-full">
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
          visibleSeriesData={visibleSeriesData}
          setVisibleSeriesData={setVisibleSeriesData}
        />
      </div>
    );

    if (chartType === CHART_TYPE_BARCHART) {
      chart = (
        <BarChart
          section={section}
          title={title}
          chartData={visibleProperties}
        />
      );
    } else if (chartType === CHART_TYPE_STACKED_AREA) {
      chart = (
        <div className="w-full">
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
        <div className="w-full">
          <StackedBarChart
            frequency={durationObj.frequency}
            categories={categories}
            data={visibleSeriesData}
            showAllLegends={true}
          />
        </div>
      );
    } else if (chartType === CHART_TYPE_LINECHART) {
      chart = (
        <div className="w-full">
          <LineChart
            frequency={durationObj.frequency}
            categories={categories}
            data={visibleSeriesData}
            showAllLegends={true}
          />
        </div>
      );
    } else if (chartType === CHART_TYPE_PIVOT_CHART) {
      chart = (
        <div className="w-full">
          <PivotTable
            data={aggregateData}
            breakdown={breakdown}
            metrics={queries}
            queryType={QUERY_TYPE_EVENT}
          />
        </div>
      );
    } else {
      chart = (
        <div className="w-full">
          <SingleEventMultipleBreakdownHorizontalBarChart
            aggregateData={aggregateData}
            breakdown={resultState.data.meta.query.gbp}
          />
        </div>
      );
    }

    return (
      <div className="flex items-center justify-center flex-col">
        {chart}
        {table}
      </div>
    );
  }
);

export default SingleEventMultipleBreakdown;
