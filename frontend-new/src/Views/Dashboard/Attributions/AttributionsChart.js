import React, {
  useState,
  useContext,
  useMemo,
  useCallback,
  useEffect,
} from 'react';
import { useSelector } from 'react-redux';
import {
  defaultSortProp,
  getTableColumns,
  getTableData,
  getSingleTouchPointChartData,
  getDualTouchPointChartData,
} from '../../CoreQuery/AttributionsResult/utils';

import AttributionTable from '../../CoreQuery/AttributionsResult/AttributionTable';

import {
  DASHBOARD_MODAL,
  MAX_ALLOWED_VISIBLE_PROPERTIES,
  GROUPED_MAX_ALLOWED_VISIBLE_PROPERTIES,
  CHART_TYPE_BARCHART,
  CHART_TYPE_SCATTER_PLOT,
  CHART_TYPE_TABLE,
  DASHBOARD_WIDGET_BARLINE_CHART_HEIGHT,
  DASHBOARD_WIDGET_SCATTERPLOT_CHART_HEIGHT,
  DASHBOARD_WIDGET_ATTRIBUTION_DUAL_TOUCHPOINT_BAR_CHART_HEIGHT,
  DASHBOARD_WIDGET_SECTION,
} from '../../../utils/constants';

import OptionsPopover from '../../CoreQuery/AttributionsResult/OptionsPopover';
import { getNewSorterState } from '../../../utils/dataFormatter';
import DualTouchPointChart from '../../CoreQuery/AttributionsResult/DualTouchPointChart';
import SingleTouchPointChart from '../../CoreQuery/AttributionsResult/SingleTouchPointChart';
import AttributionsScatterPlot from '../../CoreQuery/AttributionsResult/AttributionsScatterPlot';
import NoDataChart from '../../../components/NoDataChart';
import { DashboardContext } from '../../../contexts/DashboardContext';

const nodata = (
  <div className='mt-4 flex justify-center items-center w-full h-full'>
    <NoDataChart />
  </div>
);

const AttributionsChart = ({
  data,
  event,
  attribution_method,
  attribution_method_compare = null,
  currMetricsValue,
  touchpoint,
  linkedEvents,
  section,
  durationObj,
  attr_dimensions,
  content_groups,
  chartType,
  cardSize,
  unitId,
  attrQueries,
  queryOptions
}) => {
  const {
    attributionMetrics,
    setAttributionMetrics,
    handleEditQuery,
  } = useContext(DashboardContext);

  const { eventNames } = useSelector((state) => state.coreQuery);

  const [aggregateData, setAggregateData] = useState({
    categories: [],
    series: [],
  });
  const [dualTouchpointChartData, setDualTouchpointChartData] = useState([]);
  const [searchText, setSearchText] = useState('');
  const [columns, setColumns] = useState([]);
  const [tableData, setTableData] = useState([]);
  const [sorter, setSorter] = useState(defaultSortProp());
  const [visibleIndices, setVisibleIndices] = useState([]);

  const handleSorting = useCallback((prop) => {
    setSorter((currentSorter) => {
      return getNewSorterState(currentSorter, prop);
    });
  }, []);

  const handleMetricsVisibilityChange = useCallback(
    (option) => {
      setAttributionMetrics((curMetrics) => {
        const newState = curMetrics.map((metric) => {
          if (metric.header === option.header) {
            return {
              ...metric,
              enabled: !metric.enabled,
            };
          }
          return metric;
        });
        const enabledOptions = newState.filter((metric) => metric.enabled);
        if (!enabledOptions.length) {
          return curMetrics;
        } else {
          return newState;
        }
      });
    },
    [setAttributionMetrics]
  );

  const metricsOptionsPopover = useMemo(() => {
    return (
      <OptionsPopover
        options={attributionMetrics}
        onChange={handleMetricsVisibilityChange}
      />
    );
  }, [attributionMetrics, handleMetricsVisibilityChange]);

  useEffect(() => {
    setColumns(
      getTableColumns(
        sorter,
        handleSorting,
        attribution_method,
        attribution_method_compare,
        touchpoint,
        linkedEvents,
        event,
        eventNames,
        attributionMetrics,
        attr_dimensions,
        content_groups,
        durationObj,
        undefined, undefined, 
        queryOptions,
        attrQueries,
        data
      )
    );
  }, [
    attr_dimensions,
    content_groups,
    attributionMetrics,
    attribution_method,
    attribution_method_compare,
    event,
    eventNames,
    handleSorting,
    data,
    linkedEvents,
    sorter,
    touchpoint,
    durationObj,
  ]);

  useEffect(() => {
    const tableData = getTableData(
      data,
      event,
      searchText,
      sorter,
      attribution_method_compare,
      touchpoint,
      linkedEvents,
      attributionMetrics,
      attr_dimensions,
      content_groups, undefined, 
      queryOptions,
      attrQueries
    );
    setTableData(tableData);
    setVisibleIndices(
      tableData
        .slice(
          0,
          attribution_method_compare
            ? GROUPED_MAX_ALLOWED_VISIBLE_PROPERTIES
            : MAX_ALLOWED_VISIBLE_PROPERTIES
        )
        .map((v) => v.index)
    );
  }, [
    attr_dimensions,
    content_groups,
    attributionMetrics,
    attribution_method_compare,
    data,
    event,
    linkedEvents,
    searchText,
    sorter,
    touchpoint,
  ]);

  useEffect(() => {
    if (attribution_method_compare) {
      if (tableData.length && visibleIndices.length) {
        const chartData = getDualTouchPointChartData(
          tableData,
          visibleIndices,
          attr_dimensions,
          content_groups,
          touchpoint,
          attribution_method,
          attribution_method_compare,
          currMetricsValue
        );
        setDualTouchpointChartData(chartData);
      }
    } else {
      if (tableData.length && visibleIndices.length) {
        const chartData = getSingleTouchPointChartData(
          tableData,
          visibleIndices,
          attr_dimensions,
          content_groups,
          touchpoint
        );
        setAggregateData(chartData);
      }
    }
  }, [
    tableData,
    visibleIndices,
    attr_dimensions,
    content_groups,
    touchpoint,
    attribution_method,
    attribution_method_compare,
    currMetricsValue,
  ]);

  let chartContent = null;

  const scatterPlotChart = (
    <AttributionsScatterPlot
      visibleIndices={visibleIndices}
      selectedTouchpoint={touchpoint}
      attr_dimensions={attr_dimensions}
      content_groups={content_groups}
      data={tableData}
      attribution_method={attribution_method}
      attribution_method_compare={attribution_method_compare}
      section={section}
      linkedEvents={linkedEvents}
      durationObj={durationObj}
      cardSize={cardSize}
      height={DASHBOARD_WIDGET_SCATTERPLOT_CHART_HEIGHT}
      chartId={`scatterPlot-${unitId}`}
    />
  );

  const table = (
    <AttributionTable
      durationObj={durationObj}
      isWidgetModal={section === DASHBOARD_MODAL}
      visibleIndices={visibleIndices}
      setVisibleIndices={setVisibleIndices}
      maxAllowedVisibleProperties={
        attribution_method_compare
          ? GROUPED_MAX_ALLOWED_VISIBLE_PROPERTIES
          : MAX_ALLOWED_VISIBLE_PROPERTIES
      }
      attributionMetrics={attributionMetrics}
      section={section}
      columns={columns}
      tableData={tableData}
      searchText={searchText}
      setSearchText={setSearchText}
    />
  );

  if (attribution_method_compare) {
    if (!dualTouchpointChartData.length) {
      return nodata;
    }
    if (chartType === CHART_TYPE_BARCHART) {
      chartContent = (
        <DualTouchPointChart
          attribution_method={attribution_method}
          attribution_method_compare={attribution_method_compare}
          currMetricsValue={currMetricsValue}
          chartsData={dualTouchpointChartData}
          visibleIndices={visibleIndices}
          event={event}
          data={tableData}
          chartType={chartType}
          height={DASHBOARD_WIDGET_ATTRIBUTION_DUAL_TOUCHPOINT_BAR_CHART_HEIGHT}
          section={DASHBOARD_WIDGET_SECTION}
          cardSize={cardSize}
          chartId={`groupedBarChart-${unitId}`}
        />
      );
    } else if (chartType === CHART_TYPE_SCATTER_PLOT) {
      chartContent = scatterPlotChart;
    } else {
      chartContent = table;
    }
  } else {
    if (!aggregateData.categories.length) {
      return nodata;
    }
    if (chartType === CHART_TYPE_BARCHART) {
      chartContent = (
        <SingleTouchPointChart
          aggregateData={aggregateData}
          durationObj={durationObj}
          attribution_method={attribution_method}
          chartType={chartType}
          height={DASHBOARD_WIDGET_BARLINE_CHART_HEIGHT}
          cardSize={cardSize}
          legendsPosition='top'
          chartId={`barLineChart-${unitId}`}
        />
      );
    } else if (chartType === CHART_TYPE_SCATTER_PLOT) {
      chartContent = scatterPlotChart;
    } else {
      chartContent = table;
    }
  }

  let tableContent = null;

  if (chartType === CHART_TYPE_TABLE) {
    tableContent = (
      <div
        onClick={handleEditQuery}
        style={{ color: '#5949BC' }}
        className='mt-3 font-medium text-base cursor-pointer flex justify-end item-center'
      >
        Show More &rarr;
      </div>
    );
  }

  return (
    <div className={`w-full px-6 flex flex-1 flex-col  justify-center`}>
      {chartContent}
      {tableContent}
    </div>
  );
};

export default AttributionsChart;
