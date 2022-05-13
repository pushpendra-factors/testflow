import React, {
  useState, useCallback, useEffect, memo
} from 'react';
import { useSelector } from 'react-redux';
import {
  getTableColumns,
  getDataInTableFormat,
  getDateBasedColumns,
  getDateBasedTableData
} from './utils';
import DataTable from '../../../../components/DataTable';
import {
  MAX_ALLOWED_VISIBLE_PROPERTIES,
  CHART_TYPE_HORIZONTAL_BAR_CHART,
  DASHBOARD_WIDGET_SECTION
} from '../../../../utils/constants';
import { isSeriesChart } from '../../../../utils/dataFormatter';

const BreakdownTable = ({
  data,
  kpis,
  seriesData,
  categories,
  section,
  breakdown,
  chartType,
  setVisibleProperties,
  visibleProperties,
  sorter,
  handleSorting,
  dateSorter,
  handleDateSorting,
  visibleSeriesData,
  setVisibleSeriesData,
  frequency = 'date'
}) => {
  const [searchText, setSearchText] = useState('');
  const { userPropNames, eventPropNames } = useSelector(
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
        kpis,
        sorter,
        handleSorting,
        userPropNames,
        eventPropNames
      )
    );
  }, [breakdown, sorter, handleSorting, kpis, userPropNames, eventPropNames]);

  useEffect(() => {
    setTableData(getDataInTableFormat(data, searchText, sorter));
  }, [data, searchText, sorter]);

  useEffect(() => {
    setDateBasedColumns(
      getDateBasedColumns(
        categories,
        breakdown,
        kpis,
        dateSorter,
        handleDateSorting,
        frequency,
        userPropNames,
        eventPropNames
      )
    );
  }, [
    categories,
    breakdown,
    kpis,
    dateSorter,
    handleDateSorting,
    userPropNames,
    eventPropNames
  ]);

  useEffect(() => {
    setDateBasedTableData(
      getDateBasedTableData(seriesData, searchText, dateSorter)
    );
  }, [seriesData, searchText, dateSorter]);

  const getCSVData = () => {
    const activeTableData =
      chartType === isSeriesChart(chartType) ? dateBasedTableData : tableData;
    return {
      fileName: 'KPI.csv',
      data: activeTableData.map(({ index, label, ...rest }) => {
        return { ...rest };
      })
    };
  };

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
};

export default memo(BreakdownTable);
