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
import ColumnChart from '../../../../components/ColumnChart/ColumnChart';
import { has } from 'lodash';

const BreakdownChartsComponent = forwardRef(
  (
    {
      kpis,
      breakdown,
      responseData,
      chartType,
      durationObj,
      section,
      currentEventIndex,
      savedQuerySettings,
      comparison_data
    },
    ref
  ) => {
    const [visibleProperties, setVisibleProperties] = useState([]);
    const [visibleSeriesData, setVisibleSeriesData] = useState([]);
    const [sorter, setSorter] = useState(
      savedQuerySettings.sorter && Array.isArray(savedQuerySettings.sorter)
        ? savedQuerySettings.sorter
        : getDefaultSortProp({ kpis, breakdown })
    );
    const [dateSorter, setDateSorter] = useState(
      savedQuerySettings.dateSorter &&
        Array.isArray(savedQuerySettings.dateSorter)
        ? savedQuerySettings.dateSorter
        : getDefaultSortProp({ kpis, breakdown })
    );
    const [aggregateData, setAggregateData] = useState([]);
    const [categories, setCategories] = useState([]);
    const [compareCategories, setCompareCategories] = useState([]);
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
        currentEventIndex,
        comparison_data.data
      );
      const {
        categories: cats,
        data: d,
        compareCategories: compCategories
      } = formatDataInSeriesFormat(
        responseData,
        aggData,
        currentEventIndex,
        durationObj.frequency,
        breakdown,
        comparison_data.data
      );
      setAggregateData(aggData);
      setCategories(cats);
      setCompareCategories(compCategories);
      setData(d);
    }, [
      responseData,
      breakdown,
      currentEventIndex,
      kpis,
      durationObj,
      comparison_data.data
    ]);

    useEffect(() => {
      setVisibleProperties(getVisibleData(aggregateData, sorter));
    }, [aggregateData, sorter]);

    useEffect(() => {
      setVisibleSeriesData(getVisibleSeriesData(data, dateSorter));
    }, [data, dateSorter]);

    const visibleSeriesDataWithoutComparisonData = useMemo(() => {
      return visibleSeriesData.filter((sd) => !has(sd, 'compareIndex'));
    }, [visibleSeriesData]);

    const columnCategories = useMemo(() => {
      return visibleProperties.map((v) => v.label);
    }, [visibleProperties]);

    const columnSeries = useMemo(() => {
      const series = [
        {
          data: visibleProperties.map((v) => v.value),
          color: '#4D7DB4'
        }
      ];
      if (comparison_data.data != null) {
        series.unshift({
          data: visibleProperties.map((v) => v.compareValue)
        });
      }
      return series;
    }, [visibleProperties, comparison_data.data]);

    if (!aggregateData.length) {
      return (
        <div className="flex justify-center items-center w-full h-full pt-4 pb-4">
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
          comparisonApplied={comparison_data.data != null}
          compareCategories={compareCategories}
        />
      </div>
    );

    if (chartType === CHART_TYPE_BARCHART) {
      chart = (
        <div className="w-full">
          <ColumnChart
            comparisonApplied={comparison_data.data != null}
            categories={columnCategories}
            series={columnSeries}
          />
        </div>
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
            comparisonApplied={comparison_data.data != null}
            compareCategories={compareCategories}
          />
        </div>
      );
    } else if (chartType === CHART_TYPE_STACKED_AREA) {
      chart = (
        <div className="w-full">
          <StackedAreaChart
            frequency={durationObj.frequency}
            categories={categories}
            data={visibleSeriesDataWithoutComparisonData}
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
            data={visibleSeriesDataWithoutComparisonData}
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

const BreakdownChartsMemoized = memo(BreakdownChartsComponent);

const BreakdownCharts = (props) => {
  const { renderedCompRef, ...rest } = props;
  const {
    coreQueryState: { savedQuerySettings, comparison_data }
  } = useContext(CoreQueryContext);

  return (
    <BreakdownChartsMemoized
      ref={renderedCompRef}
      savedQuerySettings={savedQuerySettings}
      comparison_data={comparison_data}
      {...rest}
    />
  );
};

export default memo(BreakdownCharts);
