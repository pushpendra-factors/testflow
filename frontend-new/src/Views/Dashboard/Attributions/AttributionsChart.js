import React, {
  useState,
  useContext,
  useCallback,
  useEffect,
  useMemo
} from 'react';
import cx from 'classnames';
import { useSelector } from 'react-redux';
import get from 'lodash/get';
import {
  defaultSortProp,
  getTableColumns,
  getTableData,
  getSingleTouchPointChartData,
  getDualTouchPointChartData,
  getResultantMetrics
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
  DASHBOARD_WIDGET_SECTION
} from '../../../utils/constants';
import { getNewSorterState } from '../../../utils/dataFormatter';
import DualTouchPointChart from '../../CoreQuery/AttributionsResult/DualTouchPointChart';
import SingleTouchPointChart from '../../CoreQuery/AttributionsResult/SingleTouchPointChart';
import AttributionsScatterPlot from '../../CoreQuery/AttributionsResult/AttributionsScatterPlot';
import NoDataChart from '../../../components/NoDataChart';
import { DashboardContext } from '../../../contexts/DashboardContext';
import { ATTRIBUTION_GROUP_ANALYSIS_KEYS } from '../../CoreQuery/AttributionsResult/attributionsResult.constants';

const nodata = (
  <div className='flex justify-center items-center w-full h-full pt-4 pb-4'>
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
  const { attributionMetrics, tableFilters } = useContext(DashboardContext);
  const displayedAttributionMetrics = useMemo(
    () => getResultantMetrics(touchpoint, attributionMetrics),
    [touchpoint, attributionMetrics]
  );

  const { eventNames } = useSelector((state) => state.coreQuery);

  const [aggregateData, setAggregateData] = useState({
    categories: [],
    series: []
  });
  const [dualTouchpointChartData, setDualTouchpointChartData] = useState([]);
  const [searchText, setSearchText] = useState('');
  const [columns, setColumns] = useState([]);
  const [tableData, setTableData] = useState([]);
  const [sorter, setSorter] = useState(
    defaultSortProp(queryOptions, attrQueries, data)
  );
  const [visibleIndices, setVisibleIndices] = useState([]);

  const handleSorting = useCallback((prop) => {
    setSorter((currentSorter) => {
      return getNewSorterState(currentSorter, prop);
    });
  }, []);

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
        displayedAttributionMetrics,
        attr_dimensions,
        content_groups,
        durationObj,
        undefined,
        undefined,
        queryOptions,
        attrQueries,
        data
      )
    );
  }, [
    attr_dimensions,
    content_groups,
    displayedAttributionMetrics,
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
    queryOptions,
    attrQueries
  ]);

  useEffect(() => {
    const { tableData } = getTableData(
      data,
      event,
      searchText,
      sorter,
      attribution_method_compare,
      touchpoint,
      linkedEvents,
      displayedAttributionMetrics,
      attr_dimensions,
      content_groups,
      undefined,
      queryOptions,
      attrQueries,
      tableFilters
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
    displayedAttributionMetrics,
    attribution_method_compare,
    data,
    event,
    linkedEvents,
    searchText,
    sorter,
    touchpoint,
    queryOptions,
    attrQueries,
    tableFilters
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
          currMetricsValue,
          attrQueries,
          get(
            queryOptions,
            'group_analysis',
            ATTRIBUTION_GROUP_ANALYSIS_KEYS.USERS
          )
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
          touchpoint,
          false,
          attrQueries,
          get(
            queryOptions,
            'group_analysis',
            ATTRIBUTION_GROUP_ANALYSIS_KEYS.USERS
          )
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
    queryOptions,
    attrQueries
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
      attributionMetrics={displayedAttributionMetrics}
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

  return (
    <div
      className={cx('w-full flex-1', {
        'px-2 flex justify-center flex-col': chartType !== CHART_TYPE_TABLE
      })}
    >
      {chartContent}
    </div>
  );
};

export default AttributionsChart;
