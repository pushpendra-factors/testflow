import React, { useState, useCallback, useEffect, memo } from 'react';
import moment from 'moment';
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
  DATE_FORMATS,
} from '../../../../utils/constants';

const NoBreakdownTable = ({
  data,
  seriesData,
  categories,
  kpis,
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
      getTableColumns(kpis, sorter, handleSorting, eventNames, frequency)
    );
  }, [kpis, sorter, handleSorting, eventNames]);

  useEffect(() => {
    setTableData(getDataInTableFormat(seriesData, categories, kpis, sorter));
  }, [seriesData, categories, kpis, sorter]);

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
      chartType !== CHART_TYPE_SPARKLINES
        ? dateBasedTableData
        : tableData;
    const format = DATE_FORMATS[frequency] || DATE_FORMATS['date'];
    return {
      fileName: `KPI.csv`,
      data: activeTableData.map(({ index, label, date, ...rest }) => {
        return chartType === CHART_TYPE_SPARKLINES ? { date: moment(date).format(format), ...rest } : { ...rest };
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
