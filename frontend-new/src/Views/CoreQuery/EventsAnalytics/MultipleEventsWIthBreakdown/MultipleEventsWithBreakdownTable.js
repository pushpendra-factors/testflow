import React, { useState, useCallback, useEffect, memo } from 'react';
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
import { useSelector } from 'react-redux';
import { isSeriesChart } from '../../../../utils/dataFormatter';

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
  setVisibleSeriesData,
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
    eventPropNames,
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
    eventPropNames,
  ]);

  useEffect(() => {
    setDateBasedTableData(
      getDateBasedTableData(seriesData, dateSorter, searchText)
    );
  }, [dateSorter, searchText, seriesData]);

  const getCSVData = () => {
    const activeTableData =
      chartType === CHART_TYPE_BARCHART ? tableData : dateBasedTableData;
    return {
      fileName: `${reportTitle}.csv`,
      data: activeTableData.map(
        ({ index, eventIndex, dateWise, color, label, value, ...rest }) => {
          const result = {};
          for (let obj in rest) {
            if (obj.toLowerCase() === 'event') {
              result['Event'] = eventNames[rest[obj]] || rest[obj];
            } else {
              result[obj.split(';')[0]] = rest[obj];
            }
          }
          return result;
        }
      ),
    };
  };

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
      : onSelectionChange,
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
