import React, { useState, useCallback, useMemo, useEffect } from 'react';
import { getTableColumns, getTableData } from './utils';
import DataTable from '../../../../components/DataTable';
import {
  MAX_ALLOWED_VISIBLE_PROPERTIES,
  DASHBOARD_WIDGET_SECTION,
  CHART_TYPE_HORIZONTAL_BAR_CHART,
} from '../../../../utils/constants';
import { useSelector } from 'react-redux';

function BreakdownTable({
  aggregateData,
  queries,
  breakdown,
  currentEventIndex,
  chartType,
  isWidgetModal,
  visibleProperties,
  setVisibleProperties,
  section,
  reportTitle = 'Profile Analytics',
  handleSorting,
  sorter,
}) {
  const [searchText, setSearchText] = useState('');
  const [columns, setColumns] = useState([]);
  const [tableData, setTableData] = useState([]);

  const { userPropNames, eventPropNames } = useSelector(
    (state) => state.coreQuery
  );

  useEffect(() => {
    setColumns(
      getTableColumns(
        queries,
        breakdown,
        currentEventIndex,
        sorter,
        handleSorting,
        eventPropNames,
        userPropNames
      )
    );
  }, [
    queries,
    breakdown,
    currentEventIndex,
    sorter,
    handleSorting,
    eventPropNames,
    userPropNames,
  ]);

  useEffect(() => {
    setTableData(getTableData(aggregateData, searchText, sorter));
  }, [aggregateData, breakdown, searchText, sorter]);

  const getCSVData = () => {
    return {
      fileName: `${reportTitle}.csv`,
      data: tableData.map(({ index, color, label, ...rest }) => {
        return rest;
      }),
    };
  };

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
        const obj = aggregateData.find((d) => d.index === elem.index);
        return obj;
      });
      setVisibleProperties(newVisibleProperties);
    },
    [setVisibleProperties, aggregateData]
  );

  const rowSelection =
    chartType !== CHART_TYPE_HORIZONTAL_BAR_CHART
      ? {
          selectedRowKeys,
          onChange: onSelectionChange,
        }
      : null;

  return (
    <DataTable
      isWidgetModal={isWidgetModal}
      tableData={tableData}
      searchText={searchText}
      setSearchText={setSearchText}
      columns={columns}
      scroll={{ x: 250 }}
      rowSelection={rowSelection}
      getCSVData={getCSVData}
    />
  );
}

export default BreakdownTable;
