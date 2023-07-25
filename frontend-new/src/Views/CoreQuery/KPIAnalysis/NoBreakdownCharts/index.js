import React, {
  useState,
  useEffect,
  useCallback,
  forwardRef,
  useImperativeHandle,
  useContext,
  memo
} from 'react';
import { CoreQueryContext } from '../../../../contexts/CoreQueryContext';
import {
  getDefaultDateSortProp,
  getDefaultSortProp,
  formatData,
  formatDataInSeriesFormat
} from './utils';
import NoDataChart from '../../../../components/NoDataChart';
import {
  generateColors,
  getNewSorterState
} from '../../../../utils/dataFormatter';
import {
  CHART_TYPE_SPARKLINES,
  CHART_TYPE_LINECHART,
  CHART_TYPE_METRIC_CHART,
  MAX_ALLOWED_VISIBLE_PROPERTIES
} from '../../../../utils/constants';
import LineChart from '../../../../components/HCLineChart';
import NoBreakdownTable from './NoBreakdownTable';
import SparkChartWithCount from '../../../../components/SparkChartWithCount/SparkChartWithCount';
import { getKpiLabel } from '../kpiAnalysis.helpers';
import MetricChart from 'Components/MetricChart/MetricChart';

const colors = generateColors(MAX_ALLOWED_VISIBLE_PROPERTIES);
const NoBreakdownChartsComponent = forwardRef(
  (
    {
      kpis,
      responseData,
      chartType,
      durationObj,
      section,
      savedQuerySettings,
      comparisonData,
      secondAxisKpiIndices = []
    },
    ref
  ) => {
    const comparisonApplied = !!comparisonData.data;

    const [sorter, setSorter] = useState(
      savedQuerySettings.sorter && Array.isArray(savedQuerySettings.sorter)
        ? savedQuerySettings.sorter
        : getDefaultSortProp(kpis)
    );
    const [dateSorter, setDateSorter] = useState(
      savedQuerySettings.dateSorter &&
        Array.isArray(savedQuerySettings.dateSorter)
        ? savedQuerySettings.dateSorter
        : getDefaultDateSortProp()
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
      const aggData = formatData(responseData, kpis, comparisonData.data);
      const {
        categories: cats,
        data: d,
        compareCategories: compareCats
      } = formatDataInSeriesFormat(aggData, !!comparisonData.data);
      setAggregateData(aggData);
      setCategories(cats);
      setCompareCategories(compareCats);
      setData(d);
    }, [responseData, kpis, comparisonData.data, secondAxisKpiIndices]);

    if (!aggregateData.length) {
      return (
        <div className='flex justify-center items-center w-full h-full pt-4 pb-4'>
          <NoDataChart />
        </div>
      );
    }

    let chart = null;
    const table = (
      <div className='mt-12 w-full'>
        <NoBreakdownTable
          seriesData={data}
          section={section}
          chartType={chartType}
          frequency={durationObj.frequency}
          categories={categories}
          sorter={sorter}
          handleSorting={handleSorting}
          dateSorter={dateSorter}
          handleDateSorting={handleDateSorting}
          kpis={kpis}
          comparisonApplied={comparisonApplied}
          compareCategories={compareCategories}
        />
      </div>
    );

    if (chartType === CHART_TYPE_SPARKLINES) {
      if (aggregateData.length === 1) {
        chart = (
          <div className='flex items-center justify-center w-full'>
            <SparkChartWithCount
              total={aggregateData[0].total}
              event={aggregateData[0].name}
              frequency={durationObj.frequency}
              metricType={aggregateData[0].metricType}
              chartData={aggregateData[0].dataOverTime}
              compareTotal={aggregateData[0].compareTotal}
              comparisonApplied={
                aggregateData[0].compareTotal != null &&
                aggregateData[0].compareTotal > 0
              }
              headerTitle={getKpiLabel(kpis[0])}
            />
          </div>
        );
      }

      if (aggregateData.length > 1) {
        const appliedColors = generateColors(aggregateData.length);
        const kpisWithData = kpis.filter(
          (_, index) => aggregateData[index].total
        );
        chart = (
          <div className='flex items-center flex-wrap justify-center w-full'>
            {aggregateData
              .filter((d) => d.total)
              .map((chartData, index) => (
                <div
                  style={{ minWidth: '300px' }}
                  key={chartData.index}
                  className='w-1/3 mt-4 px-4'
                >
                  <SparkChartWithCount
                    total={chartData.total}
                    event={chartData.name}
                    frequency={durationObj.frequency}
                    metricType={chartData.metricType}
                    chartData={chartData.dataOverTime}
                    chartColor={appliedColors[index]}
                    alignment='vertical'
                    compareTotal={chartData.compareTotal}
                    comparisonApplied={
                      chartData.compareTotal != null &&
                      chartData.compareTotal > 0
                    }
                    headerTitle={getKpiLabel(kpisWithData[index])}
                  />
                </div>
              ))}
          </div>
        );
      }
    } else if (chartType === CHART_TYPE_LINECHART) {
      chart = (
        <div className='w-full'>
          <LineChart
            frequency={durationObj.frequency}
            categories={categories}
            data={data}
            showAllLegends
            comparisonApplied={comparisonApplied}
            compareCategories={compareCategories}
            secondaryYAxisIndices={secondAxisKpiIndices}
          />
        </div>
      );
    } else if (chartType === CHART_TYPE_METRIC_CHART) {
      chart = (
        <div className='grid grid-cols-3 w-full col-gap-2 row-gap-12'>
          {aggregateData &&
            aggregateData.map((eachAggregateData, eachIndex) => {
              return (
                <MetricChart
                  key={eachAggregateData.name}
                  headerTitle={eachAggregateData.name}
                  value={eachAggregateData.total}
                  iconColor={colors[eachIndex]}
                  compareValue={eachAggregateData.compareTotal}
                  showComparison={comparisonData.data != null}
                  valueType={
                    eachAggregateData.metricType === 'percentage_type'
                      ? 'percentage'
                      : 'numerical'
                  }
                />
              );
            })}
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

const NoBreakdownChartsMemoized = memo(NoBreakdownChartsComponent);

function NoBreakdownCharts(props) {
  const {
    coreQueryState: {
      savedQuerySettings,
      comparison_data: comparisonData
      // comparison_duration: comparisonDuration
    }
  } = useContext(CoreQueryContext);

  return (
    <NoBreakdownChartsMemoized
      savedQuerySettings={savedQuerySettings}
      comparisonData={comparisonData}
      // comparisonDuration={comparisonDuration}
      {...props}
    />
  );
}

export default NoBreakdownCharts;
