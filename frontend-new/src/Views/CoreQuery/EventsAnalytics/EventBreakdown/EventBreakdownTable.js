import React, { useState, useCallback } from 'react';
import { getTableColumns, getTableData } from './utils';
import DataTable from '../../../../components/DataTable';
import { useSelector } from 'react-redux';
import { getNewSorterState } from '../../../../utils/dataFormatter';

function EventBreakdownTable({
  breakdown,
  data,
  visibleProperties,
  setVisibleProperties,
  maxAllowedVisibleProperties,
  reportTitle = 'Events Analytics',
}) {
  const { userPropNames, eventPropNames } = useSelector(
    (state) => state.coreQuery
  );
  const [sorter, setSorter] = useState({
    key: 'User Count',
    type: 'numerical',
    subtype: null,
    order: 'descend',
  });
  const [searchText, setSearchText] = useState('');

  const handleSorting = useCallback((prop) => {
    setSorter((currentSorter) => {
      return getNewSorterState(currentSorter, prop);
    });
  }, []);

  const columns = getTableColumns(
    breakdown,
    sorter,
    handleSorting,
    userPropNames,
    eventPropNames
  );
  const tableData = getTableData(data, breakdown, searchText, sorter);

  const getCSVData = () => {
    return {
      fileName: `${reportTitle}.csv`,
      data: tableData.map(({ index, ...rest }) => {
        return { ...rest };
      }),
    };
  };

  const onSelectionChange = (selectedIncices) => {
    if (selectedIncices.length > maxAllowedVisibleProperties) {
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

  const selectedRowKeys = visibleProperties.map((elem) => elem.index);

  const rowSelection = {
    selectedRowKeys,
    onChange: onSelectionChange,
  };

  return (
    <DataTable
      tableData={tableData}
      searchText={searchText}
      setSearchText={setSearchText}
      columns={columns}
      rowSelection={rowSelection}
      getCSVData={getCSVData}
    />
  );
}

export default EventBreakdownTable;
