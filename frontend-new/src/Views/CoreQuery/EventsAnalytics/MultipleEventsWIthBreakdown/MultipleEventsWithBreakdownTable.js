import React, {
  useState, useCallback, useEffect, memo
} from 'react';
import find from 'lodash/find';
import {
  getTableColumns,
  getTableData,
  getDateBasedColumns,
  getDateBasedTableData
} from './utils';
import DataTable from '../../../../components/DataTable';
import {
  CHART_TYPE_BARCHART,
  MAX_ALLOWED_VISIBLE_PROPERTIES,
  DASHBOARD_WIDGET_SECTION
} from '../../../../utils/constants';
import { useSelector } from 'react-redux';
import { isSeriesChart } from '../../../../utils/dataFormatter';
import { getBreakdownDisplayName } from '../eventsAnalytics.helpers';

function MultipleEventsWithBreakdownTable({
  chartType,
  breakdown,
  data,
  seriesData,
  categories,
  visibleProperties,
  setVisibleProperties,
  page,
  isWidgetModal,
  durationObj,
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
  const { eventNames, userPropNames, eventPropNames } = useSelector(
    (state) => state.coreQuery
  );

  const [columns, setColumns] = useState([]);
  const [dateBasedColumns, setDateBasedColumns] = useState([]);
  const [tableData, setTableData] = useState([]);
  const [dateBasedTableData, setDateBasedTableData] = useState([]);

  useEffect(() => {
    setColumns(
      getTableColumns(
        breakdown,
        sorter,
        handleSorting,
        page,
        eventNames,
        userPropNames,
        eventPropNames
      )
    );
  }, [
    breakdown,
    sorter,
    handleSorting,
    page,
    eventNames,
    userPropNames,
    eventPropNames
  ]);

  useEffect(() => {
    setTableData(getTableData(data, searchText, sorter));
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
        eventPropNames
      )
    );
  }, [
    categories,
    breakdown,
    dateSorter,
    handleDateSorting,
    durationObj.frequency,
    userPropNames,
    eventPropNames
  ]);

  useEffect(() => {
    setDateBasedTableData(
      getDateBasedTableData(seriesData, dateSorter, searchText)
    );
  }, [dateSorter, searchText, seriesData]);

  const getCSVData = useCallback(() => {
    const activeTableData = isSeriesChart(chartType)
      ? dateBasedTableData
      : tableData;
    return {
      fileName: `${reportTitle}.csv`,
      data: activeTableData.map(
        ({
          index,
          eventIndex,
          dateWise,
          color,
          label,
          value,
          name,
          data,
          marker,
          ...rest
        }) => {
          const result = {};
          for (const key in rest) {
            if (key.toLowerCase() === 'event') {
              result.Event = rest[key];
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
                  eventPropNames
                })} - ${key.split(' - ')[1]}`
              ] = rest[key];
              continue;
            }
            result[key.split(';')[0]] = rest[key];
          }
          return result;
        }
      )
    };
  }, [
    tableData,
    dateBasedTableData,
    reportTitle,
    eventNames,
    chartType,
    breakdown,
    userPropNames,
    eventPropNames
  ]);

  const onSelectionChange = (selectedIncices) => {
    if (selectedIncices.length > MAX_ALLOWED_VISIBLE_PROPERTIES) {
      return false;
    }
    if (!selectedIncices.length) {
      return false;
    }
    const newSelectedRows = selectedIncices.map((idx) => {
      return data.find((elem) => elem.index === idx);
    });
    setVisibleProperties(newSelectedRows);
  };

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

  const selectedRowKeys = useCallback((rows) => {
    return rows.map((vp) => vp.index);
  }, []);

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
        chartType === CHART_TYPE_BARCHART ||
        section === DASHBOARD_WIDGET_SECTION
          ? tableData
          : dateBasedTableData
      }
      searchText={searchText}
      setSearchText={setSearchText}
      columns={
        chartType === CHART_TYPE_BARCHART ||
        section === DASHBOARD_WIDGET_SECTION
          ? columns
          : dateBasedColumns
      }
      rowSelection={rowSelection}
      scroll={{ x: 250 }}
      getCSVData={getCSVData}
    />
  );
}

export default memo(MultipleEventsWithBreakdownTable);
