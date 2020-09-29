import React, { useState, useCallback } from 'react';
import { getNoGroupingTableData, getColumns, getDateBasedColumns, getNoGroupingTablularDatesBasedData } from '../utils';
import DataTable from '../../CoreQuery/FunnelsResultPage/DataTable';

function TotalEventsTable({ data, events, reverseEventsMapper, chartType, setHiddenEvents, hiddenEvents }) {
  const [sorter, setSorter] = useState({});
  const [searchText, setSearchText] = useState('');

  const handleSorting = useCallback((sorter) => {
    setSorter(sorter);
  }, []);

  let columns, tableData, rowSelection = null, onSelectionChange, selectedRowKeys;

  if (chartType === 'sparklines') {
    columns = getColumns(events, sorter, handleSorting);
    tableData = getNoGroupingTableData(data, sorter, searchText, reverseEventsMapper);
  } else {
    columns = getDateBasedColumns(data, sorter, handleSorting);
    tableData = getNoGroupingTablularDatesBasedData(data, sorter, searchText, reverseEventsMapper);

    onSelectionChange = (selectedRowKeys, selectedRows) => {
      const skippedEvents = events.filter(event => selectedRows.findIndex(r => r.event === event) === -1);
      if (skippedEvents.length === events.length) {
        return false;
      }
      setHiddenEvents(skippedEvents);
    };

    selectedRowKeys = [];

    events.forEach((event, index) => {
      if (hiddenEvents.indexOf(event) === -1) {
        selectedRowKeys.push(index);
      }
    })

    rowSelection = {
      selectedRowKeys,
      onChange: onSelectionChange
    };
  }



  return (
    <DataTable
      tableData={tableData}
      searchText={searchText}
      setSearchText={setSearchText}
      columns={columns}
      scroll={{ x: 250 }}
      rowSelection={rowSelection}
    />
  );
}

export default TotalEventsTable;
