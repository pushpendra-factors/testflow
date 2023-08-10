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
import has from 'lodash/has';
import {
  formatData,
  formatDataInSeriesFormat,
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
  CHART_TYPE_LINECHART,
  CHART_TYPE_METRIC_CHART,
  MAX_ALLOWED_VISIBLE_PROPERTIES
} from '../../../../utils/constants';
import StackedAreaChart from '../../../../components/StackedAreaChart';
import StackedBarChart from '../../../../components/StackedBarChart';
import {
  generateColors,
  getNewSorterState
} from '../../../../utils/dataFormatter';
import { CHART_COLOR_1 } from '../../../../constants/color.constants';
import { CoreQueryContext } from '../../../../contexts/CoreQueryContext';
import SingleEventSingleBreakdownHorizontalBarChart from './SingleEventSingleBreakdownHorizontalBarChart';
import ColumnChart from '../../../../components/ColumnChart/ColumnChart';
import MetricChart from 'Components/MetricChart/MetricChart';

const legendsProps = {
  position: 'bottom',
  showAll: true
};

const colors = generateColors(MAX_ALLOWED_VISIBLE_PROPERTIES);
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
    const [dateWiseTotals, setDateWiseTotals] = useState([]);
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
        compareCategories: compareCats,
        dateWiseTotals: dwt
      } = formatDataInSeriesFormat(
        resultState.data,
        aggData,
        durationObj.frequency,
        comparisonData.data
      );
      setAggregateData(aggData);
      setCategories(cats);
      setCompareCategories(compareCats);
      setData(d);
      setDateWiseTotals(dwt);
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
          color: CHART_COLOR_1
        }
      ];
      if (comparisonData.data != null) {
        series.unshift({
          data: visibleProperties.map((v) => v.compareValue)
        });
      }
      return series;
    }, [visibleProperties, comparisonData.data]);

    const visibleSeriesDataWithoutComparisonData = useMemo(
      () => visibleSeriesData.filter((sd) => !has(sd, 'compareIndex')),
      [visibleSeriesData]
    );

    if (!visibleProperties.length) {
      return null;
    }

    let chart = null;

    const table = (
      <div className='mt-12 w-full'>
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
          compareCategories={compareCategories}
          frequency={durationObj.frequency}
        />
      </div>
    );

    if (chartType === CHART_TYPE_BARCHART) {
      chart = (
        <div className='w-full'>
          <ColumnChart
            comparisonApplied={comparisonData.data != null}
            categories={columnCategories}
            series={columnSeries}
            multiColored
            legendsProps={legendsProps}
          />
        </div>
      );
    } else if (chartType === CHART_TYPE_STACKED_AREA) {
      chart = (
        <div className='w-full'>
          <StackedAreaChart
            frequency={durationObj.frequency}
            categories={categories}
            data={visibleSeriesDataWithoutComparisonData}
            showAllLegends
          />
        </div>
      );
    } else if (chartType === CHART_TYPE_STACKED_BAR) {
      chart = (
        <div className='w-full'>
          <StackedBarChart
            frequency={durationObj.frequency}
            categories={categories}
            showAllLegends
            data={visibleSeriesDataWithoutComparisonData}
            dateWiseTotals={dateWiseTotals}
          />
        </div>
      );
    } else if (chartType === CHART_TYPE_LINECHART) {
      chart = (
        <div className='w-full'>
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
    } else if (chartType === CHART_TYPE_METRIC_CHART) {
      chart = (
        <div className='grid grid-cols-3 w-full col-gap-2 row-gap-12'>
          {visibleSeriesData &&
            visibleSeriesData.map((eachSeriesData, eachIndex) => {
              return (
                <MetricChart
                  key={eachSeriesData.name}
                  headerTitle={eachSeriesData.name}
                  value={eachSeriesData.value}
                  iconColor={colors[eachIndex]}
                  compareValue={eachSeriesData.compareValue}
                  showComparison={comparisonData.data != null}
                />
              );
            })}
        </div>
      );
    } else {
      chart = (
        <div className='w-full'>
          <SingleEventSingleBreakdownHorizontalBarChart
            aggregateData={aggregateData}
            breakdown={resultState.data.meta.query.gbp}
            comparisonApplied={comparisonData.data != null}
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
