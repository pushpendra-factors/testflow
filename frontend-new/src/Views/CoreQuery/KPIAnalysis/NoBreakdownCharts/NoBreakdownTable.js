import React, { useState, useCallback, useEffect, memo } from 'react';
import { useSelector } from 'react-redux';
import {
  getTableColumns,
  getDataInTableFormat,
  getDateBasedColumns,
  getDateBasedTableData,
} from './utils';
import DataTable from '../../../../components/DataTable';
import {
  CHART_TYPE_SPARKLINES,
  DASHBOARD_WIDGET_SECTION,
} from '../../../../utils/constants';

const NoBreakdownTable = ({
  data,
  seriesData,
  categories,
  queries,
  section,
  chartType,
  sorter,
  handleSorting,
  dateSorter,
  handleDateSorting,
  frequency = 'date',
}) => {
  const { eventNames } = useSelector((state) => state.coreQuery);
  const [searchText, setSearchText] = useState('');
  const [columns, setColumns] = useState([]);
  const [dateBasedColumns, setDateBasedColumns] = useState([]);
  const [tableData, setTableData] = useState([]);
  const [dateBasedTableData, setDateBasedTableData] = useState([]);

  useEffect(() => {
    setColumns(
      getTableColumns(queries, sorter, handleSorting, eventNames, frequency)
    );
  }, [queries, sorter, handleSorting, eventNames]);

  useEffect(() => {
    setTableData(getDataInTableFormat(seriesData, categories, queries, sorter));
  }, [seriesData, categories, queries, sorter]);

  useEffect(() => {
    setDateBasedColumns(
      getDateBasedColumns(
        categories,
        dateSorter,
        handleDateSorting,
        eventNames,
        frequency
      )
    );
  }, [categories, dateSorter, handleDateSorting]);

  useEffect(() => {
    setDateBasedTableData(
      getDateBasedTableData(
        seriesData,
        categories,
        searchText,
        dateSorter,
        frequency
      )
    );
  }, [seriesData, searchText, dateSorter]);

  const getCSVData = () => {
    const activeTableData =
      (chartType === chartType) === CHART_TYPE_SPARKLINES
        ? dateBasedTableData
        : tableData;
    return {
      fileName: `KPI.csv`,
      data: activeTableData.map(({ index, label, ...rest }) => {
        return { ...rest };
      }),
    };
  };

  return (
    <DataTable
      tableData={
        chartType === CHART_TYPE_SPARKLINES ||
        section === DASHBOARD_WIDGET_SECTION
          ? tableData
          : dateBasedTableData
      }
      searchText={searchText}
      setSearchText={setSearchText}
      columns={
        chartType === CHART_TYPE_SPARKLINES ||
        section === DASHBOARD_WIDGET_SECTION
          ? columns
          : dateBasedColumns
      }
      rowSelection={null}
      scroll={{ x: 250 }}
      getCSVData={getCSVData}
    />
  );
};

export default memo(NoBreakdownTable);
