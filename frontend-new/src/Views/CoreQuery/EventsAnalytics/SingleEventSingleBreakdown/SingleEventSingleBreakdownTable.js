import React, { useState, useCallback, useMemo } from 'react';
import { useSelector } from 'react-redux';
import DataTable from '../../../../components/DataTable';
import {
  getTableColumns,
  getDataInTableFormat,
  getDateBasedColumns,
  getDateBasedTableData,
} from './utils';
import {
  CHART_TYPE_BARCHART,
  MAX_ALLOWED_VISIBLE_PROPERTIES,
  DASHBOARD_WIDGET_SECTION,
} from '../../../../utils/constants';

function SingleEventSingleBreakdownTable({
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
}) {
  const { eventNames } = useSelector((state) => state.coreQuery);
  
  const appliedBreakdown = useMemo(() => {
    return [breakdown[0].property];
  }, [breakdown]);

  const [sorter, setSorter] = useState({});
  const [dateSorter, setDateSorter] = useState({});
  const [searchText, setSearchText] = useState('');

  const getCSVData = () => {
    const activeTableData =
      chartType === CHART_TYPE_BARCHART ? tableData : dateBasedTableData;
    return {
      fileName: `${reportTitle}.csv`,
      data: activeTableData.map(({ index, ...rest }) => {
        return { ...rest };
      }),
    };
  };

  const handleSorting = useCallback((sorter) => {
    setSorter(sorter);
  }, []);

  const handleDateSorting = useCallback((sorter) => {
    setDateSorter(sorter);
  }, []);

  const columns = useMemo(() => {
    return getTableColumns(
      events,
      appliedBreakdown,
      sorter,
      handleSorting,
      page,
      eventNames
    );
  }, [events, appliedBreakdown, sorter, page, handleSorting, eventNames]);

  const tableData = useMemo(() => {
    return getDataInTableFormat(
      data,
      events,
      appliedBreakdown,
      searchText,
      sorter
    );
  }, [data, events, appliedBreakdown, searchText, sorter]);

  const dateBasedColumns = useMemo(() => {
    return getDateBasedColumns(
      categories,
      appliedBreakdown,
      dateSorter,
      handleDateSorting,
      durationObj.frequency
    );
  }, [
    categories,
    appliedBreakdown,
    dateSorter,
    durationObj.frequency,
    handleDateSorting,
  ]);

  const dateBasedTableData = useMemo(() => {
    return getDateBasedTableData(
      seriesData,
      categories,
      appliedBreakdown,
      searchText,
      dateSorter,
      durationObj.frequency
    );
  }, [
    seriesData,
    categories,
    appliedBreakdown,
    searchText,
    dateSorter,
    durationObj.frequency,
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

export default React.memo(SingleEventSingleBreakdownTable);
