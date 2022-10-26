import React, {
  useState,
  useEffect,
  useCallback,
  forwardRef,
  useImperativeHandle,
  useContext,
  memo,
  useMemo
} from 'react';
import {
  formatData,
  formatDataInStackedAreaFormat,
  defaultSortProp,
  getVisibleData,
  getVisibleSeriesData
} from './utils';
import SingleEventSingleBreakdownTable from './SingleEventSingleBreakdownTable';
import LineChart from '../../../../components/HCLineChart';
import {
  DASHBOARD_MODAL,
  CHART_TYPE_BARCHART,
  CHART_TYPE_STACKED_AREA,
  CHART_TYPE_STACKED_BAR,
  CHART_TYPE_LINECHART
} from '../../../../utils/constants';
import StackedAreaChart from '../../../../components/StackedAreaChart';
import StackedBarChart from '../../../../components/StackedBarChart';
import { getNewSorterState } from '../../../../utils/dataFormatter';
import { CoreQueryContext } from '../../../../contexts/CoreQueryContext';
import SingleEventSingleBreakdownHorizontalBarChart from './SingleEventSingleBreakdownHorizontalBarChart';
import ColumnChart from '../../../../components/ColumnChart/ColumnChart';

const SingleEventSingleBreakdownComponent = forwardRef(
  (
    {
      queries,
      breakdown,
      resultState,
      page,
      chartType,
      durationObj,
      section,
      savedQuerySettings,
      comparisonData
    },
    ref
  ) => {
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
    const [compareCategories, setCompareCategories] = useState([]);
    const [data, setData] = useState([]);

    const handleSorting = useCallback((prop) => {
      setSorter((currentSorter) => getNewSorterState(currentSorter, prop));
    }, []);

    const handleDateSorting = useCallback((prop) => {
      setDateSorter((currentSorter) => getNewSorterState(currentSorter, prop));
    }, []);

    useImperativeHandle(ref, () => ({
      currentSorter: { sorter, dateSorter }
    }));

    useEffect(() => {
      const aggData = formatData(resultState.data, comparisonData.data);
      const {
        categories: cats,
        data: d,
        compareCategories: compareCats
      } = formatDataInStackedAreaFormat(
        resultState.data,
        aggData,
        durationObj.frequency,
        comparisonData.data
      );
      setAggregateData(aggData);
      setCategories(cats);
      setCompareCategories(compareCats);
      setData(d);
    }, [resultState.data, durationObj.frequency, comparisonData.data]);

    useEffect(() => {
      setVisibleSeriesData(getVisibleSeriesData(data, dateSorter));
    }, [data, dateSorter]);

    useEffect(() => {
      setVisibleProperties(getVisibleData(aggregateData, sorter));
    }, [aggregateData, sorter]);

    const columnCategories = useMemo(
      () => visibleProperties.map((v) => v.label),
      [visibleProperties]
    );

    const columnSeries = useMemo(() => {
      const series = [
        {
          data: visibleProperties.map((v) => v.value),
          color: '#4D7DB4'
        }
      ];
      if (comparisonData.data != null) {
        series.unshift({
          data: visibleProperties.map((v) => v.compareValue)
        });
      }
      return series;
    }, [visibleProperties, comparisonData.data]);

    if (!visibleProperties.length) {
      return null;
    }

    let chart = null;

    const table = (
      <div className="mt-12 w-full">
        <SingleEventSingleBreakdownTable
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
          comparisonApplied={comparisonData.data != null}
        />
      </div>
    );

    if (chartType === CHART_TYPE_BARCHART) {
      chart = (
        <div className="w-full">
          <ColumnChart
            comparisonApplied={comparisonData.data != null}
            categories={columnCategories}
            series={columnSeries}
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
            showAllLegends
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
            showAllLegends
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
            showAllLegends
            comparisonApplied={comparisonData.data != null}
            compareCategories={compareCategories}
          />
        </div>
      );
    } else {
      chart = (
        <div className="w-full">
          <SingleEventSingleBreakdownHorizontalBarChart
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

const SingleEventSingleBreakdownMemoized = memo(
  SingleEventSingleBreakdownComponent
);

function SingleEventSingleBreakdown(props) {
  const { renderedCompRef, ...rest } = props;
  const {
    coreQueryState: { savedQuerySettings, comparison_data: comparisonData }
  } = useContext(CoreQueryContext);

  return (
    <SingleEventSingleBreakdownMemoized
      ref={renderedCompRef}
      savedQuerySettings={savedQuerySettings}
      comparisonData={comparisonData}
      {...rest}
    />
  );
}

export default SingleEventSingleBreakdown;
