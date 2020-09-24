import React, { useState, useCallback } from 'react';
import { getNoGroupingTableData, getColumns } from '../utils';
import DataTable from '../../CoreQuery/FunnelsResultPage/DataTable';

function TotalEventsTable({ data, events, reverseEventsMapper }) {
  const [sorter, setSorter] = useState({});
  const [searchText, setSearchText] = useState('');

  const handleSorting = useCallback((sorter) => {
    setSorter(sorter);
  }, []);

  const columns = getColumns(events, sorter, handleSorting);

  const tableData = getNoGroupingTableData(data, sorter, searchText, reverseEventsMapper);

  return (
        <DataTable
            tableData={tableData}
            searchText={searchText}
            setSearchText={setSearchText}
            columns={columns}
        />
  );
}

export default TotalEventsTable;
