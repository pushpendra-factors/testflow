import React, { useState, useEffect, useCallback } from 'react';
import {
  getTableColumns, getDataInTableFormat, getDateBasedColumns, getDateBasedTableData
} from './utils';
import DataTable from '../../../components/DataTable';

function SingleEventMultipleBreakdownTable({
  originalData, chartType, breakdown, data, visibleProperties, setVisibleProperties, maxAllowedVisibleProperties, lineChartData, page, events, isWidgetModal, durationObj
}) {
  const [sorter, setSorter] = useState({});
  const [searchText, setSearchText] = useState('');

  useEffect(() => {
    // reset sorter on change of chart type
    setSorter({});
  }, [chartType]);

  const handleSorting = useCallback((sorter) => {
    setSorter(sorter);
  }, []);

  const nonDatecolumns = getTableColumns(events, breakdown, sorter, handleSorting, page);

  let columns;
  let tableData = [];

  if (chartType === 'linechart') {
    tableData = getDateBasedTableData(data.map(elem => elem.label), originalData, nonDatecolumns, searchText, sorter, durationObj.frequency);
    columns = getDateBasedColumns(lineChartData, breakdown, sorter, handleSorting, durationObj.frequency);
  } else {
    tableData = getDataInTableFormat(data, nonDatecolumns, searchText, sorter);
    columns = nonDatecolumns;
  }

  const visibleLabels = visibleProperties.map(elem => elem.label);

  const selectedRowKeys = [];

  tableData.forEach(elem => {
    const variableColumns = nonDatecolumns.slice(0, nonDatecolumns.length - 1);
    const val = variableColumns.map(v => {
      return elem[v.title];
    });
    if (visibleLabels.indexOf(val.join(',')) > -1) {
      selectedRowKeys.push(elem.index);
    }
  });

  const onSelectionChange = (_, selectedRows) => {
    if (selectedRows.length > maxAllowedVisibleProperties || !selectedRows.length) {
      return false;
    }
    const newVisibleProperties = selectedRows.map(elem => {
      const variableColumns = nonDatecolumns.slice(0, nonDatecolumns.length - 1);
      const val = variableColumns.map(v => {
        return elem[v.title];
      });
      const obj = data.find(d => d.label === val.join(','));
      return obj;
    });
    setVisibleProperties(newVisibleProperties);
  };

  const rowSelection = {
    selectedRowKeys,
    onChange: onSelectionChange
  };

  return (
    <DataTable
      isWidgetModal={isWidgetModal}
      tableData={tableData}
      searchText={searchText}
      setSearchText={setSearchText}
      columns={columns}
      rowSelection={rowSelection}
      scroll={{ x: 250 }}
    />
  );
}

export default SingleEventMultipleBreakdownTable;
