import React, { useState, useCallback, useMemo } from 'react';
import {
  getTableColumns,
  getDataInTableFormat,
  getDateBasedColumns,
  getDateBasedTableData,
} from './utils';
import DataTable from '../../../../components/DataTable';
import {
  CHART_TYPE_BARCHART,
  MAX_ALLOWED_VISIBLE_PROPERTIES,
  DASHBOARD_WIDGET_SECTION,
} from '../../../../utils/constants';

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
  section
}) {
  const [sorter, setSorter] = useState({});
  const [dateSorter, setDateSorter] = useState({});
  const [searchText, setSearchText] = useState('');

  const handleSorting = useCallback((sorter) => {
    setSorter(sorter);
  }, []);

  const handleDateSorting = useCallback((sorter) => {
    setDateSorter(sorter);
  }, []);

  const columns = useMemo(() => {
    return getTableColumns(events, breakdown, sorter, handleSorting, page);
  }, [events, breakdown, sorter, page, handleSorting]);

  const tableData = useMemo(() => {
    return getDataInTableFormat(data, events, breakdown, searchText, sorter);
  }, [data, events, breakdown, searchText, sorter]);

  const dateBasedColumns = useMemo(() => {
    return getDateBasedColumns(
      categories,
      breakdown,
      dateSorter,
      handleDateSorting,
      durationObj.frequency
    );
  }, [
    categories,
    breakdown,
    dateSorter,
    durationObj.frequency,
    handleDateSorting,
  ]);

  const dateBasedTableData = useMemo(() => {
    return getDateBasedTableData(
      seriesData,
      categories,
      breakdown,
      searchText,
      dateSorter,
      durationObj.frequency
    );
  }, [
    seriesData,
    categories,
    breakdown,
    searchText,
    dateSorter,
    durationObj.frequency,
  ]);

  const getCSVData = useCallback(() => {
    const activeTableData =
      chartType === CHART_TYPE_BARCHART ? tableData : dateBasedTableData;
    const activeTableColumns =
      chartType === CHART_TYPE_BARCHART ? columns : dateBasedColumns;
    const csvKeys = activeTableColumns.map((c) => c.dataIndex);
    return {
      fileName: `${reportTitle}.csv`,
      data: activeTableData.map(({ index, ...rest }) => {
        const output = {};
        const existingKeys = [];
        csvKeys.forEach((key) => {
          if (existingKeys.indexOf(key) === -1) {
            output[key] = rest[key];
          } else {
            const index = existingKeys.filter((elem) => elem === key).length;
            output[`${key}-${index}`] = rest[key];
          }
          existingKeys.push(key);
        });
        return output;
      }),
    };
  }, [
    chartType,
    columns,
    dateBasedColumns,
    dateBasedTableData,
    reportTitle,
    tableData,
  ]);

  const selectedRowKeys = useMemo(() => {
    return visibleProperties.map((vp) => vp.index);
  }, [visibleProperties]);

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

  const rowSelection = {
    selectedRowKeys,
    onChange: onSelectionChange,
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

export default SingleEventMultipleBreakdownTable;
