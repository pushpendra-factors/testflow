import React, { useEffect, useState, useCallback } from 'react';
import { getTableColumns, getTableData } from './utils';
import DataTable from '../../CoreQuery/FunnelsResultPage/DataTable';

function MultipleEventsWithBreakdownTable({
  chartType, breakdown, data, visibleProperties, setVisibleProperties, maxAllowedVisibleProperties, page, queries
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

  const columns = getTableColumns(breakdown, sorter, handleSorting, page);
  const tableData = getTableData(data, columns, breakdown, searchText, sorter);

  const onSelectionChange = (_, selectedRows) => {
    if (selectedRows.length > maxAllowedVisibleProperties || !selectedRows.length) {
      return false;
    }
    setVisibleProperties(selectedRows);
  };

  const selectedRowKeys = visibleProperties.map(elem => elem.index);

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

export default MultipleEventsWithBreakdownTable;
