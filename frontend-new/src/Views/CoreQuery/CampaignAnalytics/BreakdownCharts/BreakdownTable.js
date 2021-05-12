import React, { useState, useCallback, useMemo } from 'react';
import {
  getTableColumns,
  getTableData,
  getDateBasedColumns,
  getDateBasedTableData,
} from './utils';
import DataTable from '../../../../components/DataTable';
import {
  CHART_TYPE_BARCHART,
  MAX_ALLOWED_VISIBLE_PROPERTIES,
  DASHBOARD_WIDGET_SECTION,
} from '../../../../utils/constants';

function BreakdownTable({
  chartsData,
  seriesData,
  categories,
  breakdown,
  currentEventIndex,
  chartType,
  arrayMapper,
  isWidgetModal,
  visibleProperties,
  setVisibleProperties,
  section,
  reportTitle = 'CampaignAnalytics',
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

  const getCSVData = () => {
    const activeTableData =
      chartType === CHART_TYPE_BARCHART ? tableData : dateBasedTableData;
    return {
      fileName: `${reportTitle}.csv`,
      data: activeTableData.map(({ index, ...rest }) => {
        return rest;
      }),
    };
  };

  const columns = useMemo(() => {
    return getTableColumns(arrayMapper, breakdown, sorter, handleSorting);
  }, [arrayMapper, breakdown, sorter, handleSorting]);

  const tableData = useMemo(() => {
    return getTableData(
      chartsData,
      breakdown,
      arrayMapper,
      currentEventIndex,
      searchText,
      sorter
    );
  }, [
    chartsData,
    breakdown,
    arrayMapper,
    currentEventIndex,
    searchText,
    sorter,
  ]);

  const dateBasedColumns = useMemo(() => {
    return getDateBasedColumns(
      categories,
      breakdown,
      dateSorter,
      handleDateSorting
    );
  }, [categories, breakdown, dateSorter, handleDateSorting]);

  const dateBasedTableData = useMemo(() => {
    return getDateBasedTableData(
      seriesData,
      categories,
      breakdown,
      searchText,
      dateSorter,
      arrayMapper,
      currentEventIndex
    );
  }, [
    seriesData,
    categories,
    breakdown,
    searchText,
    dateSorter,
    arrayMapper,
    currentEventIndex,
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
        const obj = chartsData.find((d) => d.index === elem.index);
        return obj;
      });
      setVisibleProperties(newVisibleProperties);
    },
    [setVisibleProperties, chartsData]
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
      scroll={{ x: 250 }}
      rowSelection={rowSelection}
      getCSVData={getCSVData}
    />
  );
}

export default BreakdownTable;
