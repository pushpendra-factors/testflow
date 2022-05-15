import React, {
  useState,
  useEffect,
  useCallback,
  forwardRef,
  useImperativeHandle,
  useContext,
  memo
} from 'react';
import {
  formatData,
  getVisibleData,
  formatDataInSeriesFormat,
  getVisibleSeriesData,
  getDefaultSortProp
} from './utils';
import { getNewSorterState } from '../../../../utils/dataFormatter';
import { CoreQueryContext } from '../../../../contexts/CoreQueryContext';
import NoDataChart from '../../../../components/NoDataChart';
import {
  CHART_TYPE_BARCHART,
  CHART_TYPE_HORIZONTAL_BAR_CHART,
  CHART_TYPE_LINECHART,
  CHART_TYPE_STACKED_AREA,
  CHART_TYPE_STACKED_BAR,
  CHART_TYPE_PIVOT_CHART
} from '../../../../utils/constants';
import LineChart from '../../../../components/HCLineChart';
import BarChart from '../../../../components/BarChart';
import BreakdownTable from './BreakdownTable';
import HorizontalBarChartTable from './HorizontalBarChartTable';
import StackedAreaChart from '../../../../components/StackedAreaChart';
import StackedBarChart from '../../../../components/StackedBarChart';
import PivotTable from '../../../../components/PivotTable';

const BreakdownCharts = forwardRef(
  (
    {
      kpis,
      breakdown,
      responseData,
      chartType,
      durationObj,
      title = 'Kpi',
      section,
      currentEventIndex
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
        : getDefaultSortProp(kpis)
    );
    const [dateSorter, setDateSorter] = useState(
      savedQuerySettings.dateSorter &&
        Array.isArray(savedQuerySettings.dateSorter)
        ? savedQuerySettings.dateSorter
        : getDefaultSortProp(kpis)
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
      const aggData = formatData(
        responseData,
        kpis,
        breakdown,
        currentEventIndex
      );
      const { categories: cats, data: d } = formatDataInSeriesFormat(
        responseData,
        aggData,
        currentEventIndex,
        durationObj.frequency,
        breakdown
      );
      setAggregateData(aggData);
      setCategories(cats);
      setData(d);
    }, [responseData, breakdown, currentEventIndex, kpis, durationObj]);

    useEffect(() => {
      setVisibleProperties(getVisibleData(aggregateData, sorter));
    }, [aggregateData, sorter]);

    useEffect(() => {
      setVisibleSeriesData(getVisibleSeriesData(data, dateSorter));
    }, [data, dateSorter]);

    if (!aggregateData.length) {
      return (
        <div className="mt-4 flex justify-center items-center w-full h-64 ">
          <NoDataChart />
        </div>
      );
    }

    let chart = null;
    const table = (
      <div className="mt-12 w-full">
        <BreakdownTable
          kpis={kpis}
          data={aggregateData}
          seriesData={data}
          section={section}
          breakdown={breakdown}
          chartType={chartType}
          setVisibleProperties={setVisibleProperties}
          visibleProperties={visibleProperties}
          frequency={durationObj.frequency}
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
    } else if (chartType === CHART_TYPE_HORIZONTAL_BAR_CHART) {
      chart = (
        <div className="w-full">
          <HorizontalBarChartTable
            breakdown={breakdown}
            aggregateData={aggregateData}
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
    } else if (chartType === CHART_TYPE_PIVOT_CHART) {
      chart = (
        <div className="w-full">
          <PivotTable
            data={aggregateData}
            breakdown={breakdown}
            metrics={kpis}
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

export default memo(BreakdownCharts);
