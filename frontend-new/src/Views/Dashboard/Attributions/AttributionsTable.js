import React, { useState, useCallback, useMemo, useEffect } from 'react';
import moment from 'moment';
import {
  getTableColumns,
  getTableData,
  calcChangePerc,
  defaultSortProp,
} from '../../CoreQuery/AttributionsResult/utils';
import DataTable from '../../../components/DataTable';
import { useSelector } from 'react-redux';
import { DASHBOARD_WIDGET_SECTION } from '../../../utils/constants';
import { getNewSorterState } from '../../../utils/dataFormatter';
import OptionsPopover from '../../CoreQuery/AttributionsResult/OptionsPopover';

function AttributionsTable({
  data,
  comparison_data,
  isWidgetModal,
  event,
  setVisibleIndices,
  visibleIndices,
  maxAllowedVisibleProperties,
  attribution_method,
  attribution_method_compare,
  touchpoint,
  linkedEvents,
  reportTitle = 'Attributions',
  durationObj,
  cmprDuration,
  attributionMetrics,
  setAttributionMetrics,
  section = null,
  attr_dimensions,
}) {
  const [searchText, setSearchText] = useState('');
  const [sorter, setSorter] = useState(defaultSortProp());
  const [columns, setColumns] = useState([]);
  const [tableData, setTableData] = useState([]);
  const { eventNames } = useSelector((state) => state.coreQuery);

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
        comparison_data,
        cmprDuration
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
    comparison_data,
    cmprDuration,
  ]);

  useEffect(() => {
    setTableData(
      getTableData(
        data,
        event,
        searchText,
        sorter,
        attribution_method_compare,
        touchpoint,
        linkedEvents,
        attributionMetrics,
        attr_dimensions,
        comparison_data
      )
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
    comparison_data,
  ]);

  const getCSVData = () => {
    const dt = tableData;
    const enabledAttributionMetricKeys = attributionMetrics
      .filter((d) => d.enabled)
      .map((d) => d.title);
    const mappedData = dt.map(({ index, ...rest }) => {
      if (!comparison_data) {
        return rest;
      }
      const fromDate = moment(durationObj.from).format('MMM DD');
      const toDate = moment(durationObj.to).format('MMM DD');
      const compareFromDate = moment(cmprDuration.from).format('MMM DD');
      const compareToDate = moment(cmprDuration.to).format('MMM DD');
      const result = {};
      Object.keys(rest).forEach((key) => {
        if (
          !enabledAttributionMetricKeys.includes(key) &&
          key !== 'Conversion' &&
          key !== 'Cost per Conversion' &&
          key !== 'Conversion Rate' &&
          !key.includes('Linked Event')
        ) {
          result[key] = rest[key];
        } else {
          const changePercent = calcChangePerc(
            rest[key].value,
            rest[key].compare_value
          );
          result[`${key} (${fromDate} - ${toDate})`] = rest[key].value;
          result[`${key} (${compareFromDate} - ${compareToDate})`] =
            rest[key].compare_value;
          result[`${key} change`] = isNaN(changePercent)
            ? '0%'
            : changePercent === 'Infinity' || changePercent === '-Infinity'
            ? 'Infinity'
            : changePercent + '%';
        }
      });
      return result;
    });

    return {
      fileName: `${reportTitle}.csv`,
      data: mappedData,
    };
  };

  const onSelectionChange = (selectedIncices) => {
    if (selectedIncices.length > maxAllowedVisibleProperties) {
      return false;
    }
    if (!selectedIncices.length) {
      return false;
    }
    selectedIncices.sort();
    setVisibleIndices(selectedIncices);
  };

  const rowSelection = {
    selectedRowKeys: visibleIndices,
    onChange: onSelectionChange,
  };

  return (
    <DataTable
      isWidgetModal={isWidgetModal}
      tableData={tableData}
      searchText={searchText}
      setSearchText={setSearchText}
      columns={columns}
      rowSelection={rowSelection}
      scroll={{ x: 250 }}
      getCSVData={getCSVData}
      ignoreDocumentClick={section === DASHBOARD_WIDGET_SECTION}
    />
  );
}

export default AttributionsTable;
