import React, {
  useState,
  useContext,
  useMemo,
  useCallback,
  useEffect,
  forwardRef,
  useImperativeHandle
} from 'react';
import get from 'lodash/get';

import {
  defaultSortProp,
  getTableColumns,
  getTableData,
  getSingleTouchPointChartData,
  getDualTouchPointChartData,
  getResultantMetrics,
  getTableFilterOptions
} from './utils';

import AttributionTable from './AttributionTable';
import {
  DASHBOARD_MODAL,
  MAX_ALLOWED_VISIBLE_PROPERTIES,
  GROUPED_MAX_ALLOWED_VISIBLE_PROPERTIES,
  CHART_TYPE_BARCHART
} from '../../../utils/constants';
import { CoreQueryContext } from '../../../contexts/CoreQueryContext';
import OptionsPopover from './OptionsPopover';
import { useSelector } from 'react-redux';
import { getNewSorterState } from '../../../utils/dataFormatter';
import DualTouchPointChart from './DualTouchPointChart';
import SingleTouchPointChart from './SingleTouchPointChart';
import AttributionsScatterPlot from './AttributionsScatterPlot';
import NoDataChart from '../../../components/NoDataChart';
import { ATTRIBUTION_GROUP_ANALYSIS_KEYS } from './attributionsResult.constants';

const nodata = (
  <div className="flex justify-center items-center w-full h-full pt-4 pb-4">
    <NoDataChart />
  </div>
);

const AttributionsChart = forwardRef(
  (
    {
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
      queryOptions
    },
    ref
  ) => {
    const {
      coreQueryState: {
        comparison_data,
        comparison_duration,
        savedQuerySettings
      },
      attributionMetrics,
      setAttributionMetrics
    } = useContext(CoreQueryContext);

    const { eventNames, attrQueries } = useSelector((state) => state.coreQuery);

    const [aggregateData, setAggregateData] = useState({
      categories: [],
      series: []
    });
    const [dualTouchpointChartData, setDualTouchpointChartData] = useState([]);
    const [searchText, setSearchText] = useState('');
    const [filters, setFilters] = useState([]);
    const [appliedFilters, setAppliedFilters] = useState({});
    const [filtersVisible, setFiltersVisibility] = useState(false);
    const [columns, setColumns] = useState([]);
    const [tableData, setTableData] = useState([]);
    const [sorter, setSorter] = useState(
      savedQuerySettings.sorter && Array.isArray(savedQuerySettings.sorter)
        ? savedQuerySettings.sorter
        : defaultSortProp()
    );
    const [visibleIndices, setVisibleIndices] = useState([]);

    const displayedAttributionMetrics = useMemo(() => {
      return getResultantMetrics(touchpoint, attributionMetrics);
    }, [touchpoint, attributionMetrics]);

    useImperativeHandle(ref, () => {
      return {
        currentSorter: { sorter }
      };
    });

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
                enabled: !metric.enabled
              };
            }
            return metric;
          });
          const enabledOptions = newState.filter(
            (metric) => metric.enabled && !metric.isEventMetric
          );
          if (!enabledOptions.length) {
            return curMetrics;
          } else {
            return newState;
          }
        });
        if (option.enabled) {
          const isSortedByThisOption = sorter.find(
            (elem) => elem.key === option.title
          );
          if (isSortedByThisOption) {
            setSorter((currentSorter) => {
              return currentSorter.filter((elem) => elem.key !== option.title);
            });
          }
        }
      },
      [setAttributionMetrics, sorter]
    );

    const metricsOptionsPopover = useMemo(() => {
      return (
        <OptionsPopover
          options={displayedAttributionMetrics}
          onChange={handleMetricsVisibilityChange}
        />
      );
    }, [displayedAttributionMetrics, handleMetricsVisibilityChange]);

    const handleApplyFilters = useCallback((filters) => {
      setAppliedFilters(filters);
      setFiltersVisibility(false);
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
          comparison_data.data,
          comparison_duration,
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
      data,
      handleSorting,
      linkedEvents,
      sorter,
      touchpoint,
      durationObj,
      comparison_data.data,
      comparison_duration,
      queryOptions,
      attrQueries
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
        displayedAttributionMetrics,
        attr_dimensions,
        content_groups,
        comparison_data.data,
        queryOptions,
        attrQueries,
        appliedFilters
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
      handleSorting,
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
      comparison_data.data,
      queryOptions,
      attrQueries,
      appliedFilters
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
        return;
      }
      if (tableData.length && visibleIndices.length) {
        const chartData = getSingleTouchPointChartData(
          tableData,
          visibleIndices,
          attr_dimensions,
          content_groups,
          touchpoint,
          !!comparison_data.data,
          attrQueries,
          get(
            queryOptions,
            'group_analysis',
            ATTRIBUTION_GROUP_ANALYSIS_KEYS.USERS
          ),
          currMetricsValue
        );
        setAggregateData(chartData);
      }
    }, [
      tableData,
      visibleIndices,
      attr_dimensions,
      content_groups,
      touchpoint,
      comparison_data.data,
      attribution_method,
      attribution_method_compare,
      currMetricsValue,
      attrQueries,
      queryOptions
    ]);

    useEffect(() => {
      const metricsNotPresentInFilters = attributionMetrics
        .filter((m) => m.enabled)
        .filter((m) => filters.findIndex((f) => f.key === m.title) === -1);

      const metricsNotEnabledButPresentInFilters = attributionMetrics
        .filter((m) => !m.enabled)
        .filter((m) => filters.findIndex((f) => f.key === m.title) > -1);

      if (
        (tableData.length && !filters.length) ||
        metricsNotPresentInFilters.length > 0 ||
        metricsNotEnabledButPresentInFilters.length > 0
      ) {
        const tableFilterOptions = getTableFilterOptions({
          content_groups,
          attr_dimensions,
          touchpoint,
          tableData,
          attributionMetrics
        });
        setFilters(tableFilterOptions);
      }
    }, [
      content_groups,
      attr_dimensions,
      touchpoint,
      tableData,
      attributionMetrics,
      filters
    ]);

    let chart = null;

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
        comparison_duration={comparison_duration}
        comparison_data={comparison_data}
      />
    );

    if (attribution_method_compare) {
      if (!dualTouchpointChartData.length) {
        return nodata;
      }
      if (chartType === CHART_TYPE_BARCHART) {
        chart = (
          <DualTouchPointChart
            attribution_method={attribution_method}
            attribution_method_compare={attribution_method_compare}
            currMetricsValue={currMetricsValue}
            chartsData={dualTouchpointChartData}
            visibleIndices={visibleIndices}
            event={event}
            data={tableData}
            chartType={chartType}
          />
        );
      } else {
        chart = scatterPlotChart;
      }
    } else {
      if (!aggregateData.categories.length) {
        return nodata;
      }
      if (chartType === CHART_TYPE_BARCHART) {
        chart = (
          <SingleTouchPointChart
            aggregateData={aggregateData}
            durationObj={durationObj}
            comparison_duration={comparison_duration}
            comparison_data={comparison_data}
            attribution_method={attribution_method}
            chartType={chartType}
          />
        );
      } else {
        chart = scatterPlotChart;
      }
    }

    return (
      <div className="flex items-center justify-center flex-col">
        <div className="w-full">{chart}</div>
        <div className="mt-12 w-full">
          <AttributionTable
            comparison_data={comparison_data.data}
            durationObj={durationObj}
            cmprDuration={comparison_duration}
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
            metricsOptionsPopover={metricsOptionsPopover}
            filters={filters}
            filtersVisible={filtersVisible}
            appliedFilters={appliedFilters}
            setAppliedFilters={handleApplyFilters}
            setFiltersVisibility={setFiltersVisibility}
          />
        </div>
      </div>
    );
  }
);

export default AttributionsChart;
