import React, { useState, useCallback, useEffect } from 'react';
import { useSelector } from 'react-redux';
import find from 'lodash/find';
import {
  getTableColumns,
  getDataInTableFormat,
  getDateBasedColumns,
  getDateBasedTableData
} from './utils';
import DataTable from '../../../../components/DataTable';
import {
  MAX_ALLOWED_VISIBLE_PROPERTIES,
  DASHBOARD_WIDGET_SECTION,
  CHART_TYPE_HORIZONTAL_BAR_CHART
} from '../../../../utils/constants';
import { isSeriesChart } from '../../../../utils/dataFormatter';
import { EVENT_COUNT_KEY } from '../eventsAnalytics.constants';
import {
  getEventDisplayName,
  getBreakdownDisplayName
} from '../eventsAnalytics.helpers';

function SingleEventMultipleBreakdownTable({
  data,
  seriesData,
  events,
  breakdown,
  chartType,
  visibleProperties,
  setVisibleProperties,
  page,
  isWidgetModal,
  durationObj,
  categories,
  reportTitle = 'Events Analytics',
  section,
  sorter,
  handleSorting,
  dateSorter,
  handleDateSorting,
  visibleSeriesData,
  setVisibleSeriesData
}) {
  const [searchText, setSearchText] = useState('');
  const {
    eventNames,
    userPropNames,
    eventPropertiesDisplayNames: eventPropertiesDisplayNamesState
  } = useSelector((state) => state.coreQuery);
  const { data: eventPropertiesDisplayNames } =
    eventPropertiesDisplayNamesState;
  const [columns, setColumns] = useState([]);
  const [dateBasedColumns, setDateBasedColumns] = useState([]);
  const [tableData, setTableData] = useState([]);
  const [dateBasedTableData, setDateBasedTableData] = useState([]);

  useEffect(() => {
    setColumns(
      getTableColumns(
        events,
        breakdown,
        sorter,
        handleSorting,
        page,
        eventNames,
        userPropNames,
        eventPropertiesDisplayNames
      )
    );
  }, [
    events,
    breakdown,
    sorter,
    page,
    handleSorting,
    eventNames,
    userPropNames,
    eventPropertiesDisplayNames
  ]);

  useEffect(() => {
    setTableData(getDataInTableFormat(data, searchText, sorter));
  }, [data, searchText, sorter]);

  useEffect(() => {
    setDateBasedColumns(
      getDateBasedColumns(
        categories,
        breakdown,
        dateSorter,
        handleDateSorting,
        durationObj.frequency,
        userPropNames,
        eventPropertiesDisplayNames
      )
    );
  }, [
    categories,
    breakdown,
    dateSorter,
    durationObj.frequency,
    handleDateSorting,
    userPropNames,
    eventPropertiesDisplayNames
  ]);

  useEffect(() => {
    setDateBasedTableData(
      getDateBasedTableData(seriesData, searchText, dateSorter)
    );
  }, [seriesData, searchText, dateSorter]);

  const getCSVData = useCallback(() => {
    const activeTableData = isSeriesChart(chartType)
      ? dateBasedTableData
      : tableData;
    return {
      fileName: `${reportTitle}.csv`,
      data: activeTableData.map(
        ({ index, label, value, name, data, marker, ...rest }) => {
          const result = {};
          for (const key in rest) {
            if (key === EVENT_COUNT_KEY) {
              result[getEventDisplayName({ eventNames, event: events[0] })] =
                rest[EVENT_COUNT_KEY];
              continue;
            }
            if (key === events[0]) {
              result[getEventDisplayName({ eventNames, event: events[0] })] =
                rest[events[0]];
              continue;
            }
            const isCurrentKeyForBreakdown = find(
              breakdown,
              (b, index) => b.property + ' - ' + index === key
            );
            if (isCurrentKeyForBreakdown) {
              result[
                `${getBreakdownDisplayName({
                  breakdown: isCurrentKeyForBreakdown,
                  userPropNames,
                  eventPropertiesDisplayNames
                })} - ${key.split(' - ')[1]}`
              ] = rest[key];
              continue;
            }
            result[key] = rest[key];
          }
          return result;
        }
      )
    };
  }, [
    dateBasedTableData,
    tableData,
    reportTitle,
    eventNames,
    events,
    breakdown,
    chartType,
    userPropNames,
    eventPropertiesDisplayNames
  ]);

  const selectedRowKeys = useCallback((rows) => {
    return rows.map((vp) => vp.index);
  }, []);

  const onSelectionChange = useCallback(
    (_, selectedRows) => {
      if (
        selectedRows.length > MAX_ALLOWED_VISIBLE_PROPERTIES ||
        !selectedRows.length
      ) {
        return false;
      }
      const newVisibleProperties = selectedRows.map((elem) => {
        const obj = data.find((d) => d.index === elem.index);
        return obj;
      });
      setVisibleProperties(newVisibleProperties);
    },
    [setVisibleProperties, data]
  );

  const onDateWiseSelectionChange = useCallback(
    (_, selectedRows) => {
      if (
        selectedRows.length > MAX_ALLOWED_VISIBLE_PROPERTIES ||
        !selectedRows.length
      ) {
        return false;
      }
      const newVisibleSeriesData = selectedRows.map((elem) => {
        const obj = seriesData.find((d) => d.index === elem.index);
        return obj;
      });
      setVisibleSeriesData(newVisibleSeriesData);
    },
    [setVisibleSeriesData, seriesData]
  );

  const rowSelection = {
    selectedRowKeys: isSeriesChart(chartType)
      ? selectedRowKeys(visibleSeriesData)
      : selectedRowKeys(visibleProperties),
    onChange: isSeriesChart(chartType)
      ? onDateWiseSelectionChange
      : onSelectionChange
  };

  return (
    <DataTable
      isWidgetModal={isWidgetModal}
      tableData={
        !isSeriesChart(chartType) || section === DASHBOARD_WIDGET_SECTION
          ? tableData
          : dateBasedTableData
      }
      searchText={searchText}
      setSearchText={setSearchText}
      columns={
        !isSeriesChart(chartType) || section === DASHBOARD_WIDGET_SECTION
          ? columns
          : dateBasedColumns
      }
      rowSelection={
        chartType !== CHART_TYPE_HORIZONTAL_BAR_CHART ? rowSelection : null
      }
      scroll={{ x: 250 }}
      getCSVData={getCSVData}
    />
  );
}

export default React.memo(SingleEventMultipleBreakdownTable);
