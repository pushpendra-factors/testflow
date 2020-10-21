import React, { useState, useCallback } from 'react';
import {
  getTableColumns, getTableData
} from './utils';
import DataTable from '../../CoreQuery/FunnelsResultPage/DataTable';

function EventBreakdownTable({
  breakdown, data, visibleProperties, setVisibleProperties, maxAllowedVisibleProperties
}) {
  const [sorter, setSorter] = useState({});
  const [searchText, setSearchText] = useState('');

  //   useEffect(() => {
  //     // reset sorter on change of chart type
  //     setSorter({});
  //   }, [chartType]);

  const handleSorting = useCallback((sorter) => {
    setSorter(sorter);
  }, []);

  // let columns, tableData;

  const tableData = getTableData(data, breakdown, searchText, sorter);
  const columns = getTableColumns(breakdown, sorter, handleSorting);

  const onSelectionChange = (selectedIncices) => {
    if (selectedIncices.length > maxAllowedVisibleProperties) {
      return false;
    }
    if (!selectedIncices.length) {
      return false;
    }
    const newSelectedRows = selectedIncices.map(idx => {
      return data.find(elem => elem.index === idx);
    });
    setVisibleProperties(newSelectedRows);
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

export default EventBreakdownTable;
