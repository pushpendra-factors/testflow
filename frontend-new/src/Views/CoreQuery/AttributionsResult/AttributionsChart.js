import React, {
  useState,
  useContext,
  useMemo,
  useCallback,
  useEffect,
  forwardRef,
  useImperativeHandle,
} from 'react';
import {
  defaultSortProp,
  getTableColumns,
  getTableData,
  getSingleTouchPointChartData,
  getDualTouchPointChartData,
} from './utils';

import AttributionTable from './AttributionTable';
import {
  DASHBOARD_MODAL,
  MAX_ALLOWED_VISIBLE_PROPERTIES,
  GROUPED_MAX_ALLOWED_VISIBLE_PROPERTIES,
  CHART_TYPE_BARCHART,
} from '../../../utils/constants';
import { CoreQueryContext } from '../../../contexts/CoreQueryContext';
import OptionsPopover from './OptionsPopover';
import { useSelector } from 'react-redux';
import { getNewSorterState } from '../../../utils/dataFormatter';
import DualTouchPointChart from './DualTouchPointChart';
import SingleTouchPointChart from './SingleTouchPointChart';
import AttributionsScatterPlot from './AttributionsScatterPlot';
import NoDataChart from '../../../components/NoDataChart';

const nodata = (
  <div className='mt-4 flex justify-center items-center w-full h-full'>
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
      chartType,
    },
    ref
  ) => {
    const {
      coreQueryState: {
        comparison_data,
        comparison_duration,
        savedQuerySettings,
      },
      attributionMetrics,
      setAttributionMetrics,
    } = useContext(CoreQueryContext);

    const { eventNames } = useSelector((state) => state.coreQuery);

    const [aggregateData, setAggregateData] = useState({
      categories: [],
      series: [],
    });
    const [dualTouchpointChartData, setDualTouchpointChartData] = useState([]);
    const [searchText, setSearchText] = useState('');
    const [columns, setColumns] = useState([]);
    const [tableData, setTableData] = useState([]);
    const [sorter, setSorter] = useState(
      savedQuerySettings.sorter || defaultSortProp()
    );
    const [visibleIndices, setVisibleIndices] = useState([]);

    useImperativeHandle(ref, () => {
      return {
        currentSorter: { sorter },
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
          metricsOptionsPopover,
          attr_dimensions,
          durationObj,
          comparison_data.data,
          comparison_duration
        )
      );
    }, [
      attr_dimensions,
      attributionMetrics,
      attribution_method,
      attribution_method_compare,
      event,
      eventNames,
      handleSorting,
      linkedEvents,
      metricsOptionsPopover,
      sorter,
      touchpoint,
      durationObj,
      comparison_data.data,
      comparison_duration,
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
        comparison_data.data
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
      attributionMetrics,
      attribution_method_compare,
      data,
      event,
      linkedEvents,
      searchText,
      sorter,
      touchpoint,
      comparison_data.data,
    ]);

    useEffect(() => {
      if (attribution_method_compare) {
        if (tableData.length && visibleIndices.length) {
          const chartData = getDualTouchPointChartData(
            tableData,
            visibleIndices,
            attr_dimensions,
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
            touchpoint,
            !!comparison_data.data
          );
          setAggregateData(chartData);
        }
      }
    }, [
      tableData,
      visibleIndices,
      attr_dimensions,
      touchpoint,
      comparison_data.data,
      attribution_method,
      attribution_method_compare,
      currMetricsValue,
    ]);

    let chart = null;

    const scatterPlotChart = (
      <AttributionsScatterPlot
        visibleIndices={visibleIndices}
        selectedTouchpoint={touchpoint}
        attr_dimensions={attr_dimensions}
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
      <div className='flex items-center justify-center flex-col'>
        <div className='w-full'>{chart}</div>
        <div className='mt-12 w-full'>
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
            attributionMetrics={attributionMetrics}
            section={section}
            columns={columns}
            tableData={tableData}
            searchText={searchText}
            setSearchText={setSearchText}
          />
        </div>
      </div>
    );
  }
);

export default AttributionsChart;
