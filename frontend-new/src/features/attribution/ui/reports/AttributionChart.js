import { get } from 'lodash';
import React, { useCallback, useEffect, useState } from 'react';
import NoDataChart from 'Components/NoDataChart';
import {
  CHART_TYPE_BARCHART,
  CHART_TYPE_SCATTER_PLOT,
  DASHBOARD_MODAL,
  DASHBOARD_WIDGET_ATTRIBUTION_DUAL_TOUCHPOINT_BAR_CHART_HEIGHT,
  DASHBOARD_WIDGET_SCATTERPLOT_CHART_HEIGHT,
  DASHBOARD_WIDGET_SECTION,
  GROUPED_MAX_ALLOWED_VISIBLE_PROPERTIES,
  MAX_ALLOWED_VISIBLE_PROPERTIES
} from 'Utils/constants';
import { getNewSorterState } from 'Utils/dataFormatter';
import {
  defaultSortProp,
  getDualTouchPointChartData,
  getSingleTouchPointChartData,
  getTableColumns,
  getTableData
} from 'Views/CoreQuery/AttributionsResult/utils';
import { ATTRIBUTION_GROUP_ANALYSIS_KEYS } from 'Views/CoreQuery/AttributionsResult/attributionsResult.constants';
import AttributionsScatterPlot from 'Views/CoreQuery/AttributionsResult/AttributionsScatterPlot';
import AttributionTable from 'Views/CoreQuery/AttributionsResult/AttributionTable';
import DualTouchPointChart from 'Views/CoreQuery/AttributionsResult/DualTouchPointChart';
import SingleTouchPointChart from 'Views/CoreQuery/AttributionsResult/SingleTouchPointChart';
import { useSelector } from 'react-redux';

const nodata = (
  <div className='flex justify-center items-center w-full h-full pt-4 pb-4'>
    <NoDataChart />
  </div>
);

function AttributionChart({
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
  queryOptions,
  attributionMetrics
}) {
  const { eventNames } = useSelector((state) => state.coreQuery);
  const [aggregateData, setAggregateData] = useState({
    categories: [],
    series: []
  });
  const [dualTouchpointChartData, setDualTouchpointChartData] = useState([]);
  const [searchText, setSearchText] = useState('');
  const [columns, setColumns] = useState([]);
  const [tableData, setTableData] = useState([]);
  const [sorter, setSorter] = useState(defaultSortProp());
  const [visibleIndices, setVisibleIndices] = useState([]);
  const [filtersVisible, setFiltersVisibility] = useState(false);
  const [tableFilters, setAttributionTableFilters] = useState({});


  const handleApplyFilters = 
    (filters) => {
      setAttributionTableFilters({ attributionTableFilters: filters });
      setFiltersVisibility(false);
    }

  const handleSorting = useCallback((prop) => {
    setSorter((currentSorter) => {
      return getNewSorterState(currentSorter, prop);
    });
  }, []);

  //pass the dependencies of the useEffect

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
        undefined,
        undefined,
        queryOptions,
        attrQueries,
        data
      )
    );
  }, []);

  useEffect(() => {
    const { tableData } = getTableData(
      data,
      event,
      searchText,
      sorter,
      attribution_method_compare,
      touchpoint,
      linkedEvents,
      attributionMetrics,
      attr_dimensions,
      content_groups,
      undefined,
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
  }, []);

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
      attributionMetrics={attributionMetrics}
      filtersVisible={filtersVisible}
      section={section}
      columns={columns}
      tableData={tableData}
      searchText={searchText}
      appliedFilters={tableFilters.appliedFilters}
      setAppliedFilters={handleApplyFilters}
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

  return <div className={'w-full'}>{chartContent}</div>;
}

export default AttributionChart;
