import React, { useState, useCallback, useEffect } from 'react';
import DataTable from '../../CoreQuery/FunnelsResultPage/DataTable';
import {
  getTableColumns, getDataInTableFormat, getDateBasedColumns, getDateBasedTableData, getDateBasedUserTableData
} from './utils';

function SingleEventSingleBreakdownTable({
  data, events, breakdown, chartType, visibleProperties, setVisibleProperties, maxAllowedVisibleProperties, lineChartData, originalData, page
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

  let columns;
  let tableData;

  if (chartType === 'linechart') {
    columns = getDateBasedColumns(lineChartData, breakdown, sorter, handleSorting);
    if (page === 'totalEvents') {
      tableData = getDateBasedTableData(data.map(elem => elem.label), originalData, breakdown, searchText, sorter);
    } else {
      tableData = getDateBasedUserTableData(data.map(elem => elem.label), originalData, breakdown, searchText, sorter);
    }
  } else {
    columns = getTableColumns(events, breakdown, sorter, handleSorting);
    tableData = getDataInTableFormat(data, events, breakdown, searchText, sorter);
  }

  const visibleLabels = visibleProperties.map(elem => elem.label);

  const selectedRowKeys = [];

  tableData.forEach(elem => {
    if (visibleLabels.indexOf(elem[breakdown[0]]) > -1) {
      selectedRowKeys.push(elem.index);
    }
  });

  const onSelectionChange = (_, selectedRows) => {
    if (selectedRows.length > maxAllowedVisibleProperties || !selectedRows.length) {
      return false;
    }
    const newVisibleProperties = selectedRows.map(elem => {
      const obj = data.find(d => d.label === elem[breakdown[0]]);
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
            tableData={tableData}
            searchText={searchText}
            setSearchText={setSearchText}
            columns={columns}
            rowSelection={rowSelection}
        />
  );
}

export default SingleEventSingleBreakdownTable;
