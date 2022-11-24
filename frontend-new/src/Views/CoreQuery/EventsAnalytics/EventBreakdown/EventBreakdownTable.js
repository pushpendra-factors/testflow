import React, { useState, useCallback } from 'react';
import { getTableColumns, getTableData } from './utils';
import DataTable from '../../../../components/DataTable';
import { useSelector } from 'react-redux';
import { getNewSorterState } from '../../../../utils/dataFormatter';
import { MAX_ALLOWED_VISIBLE_PROPERTIES } from '../../../../utils/constants';

function EventBreakdownTable({
  breakdown,
  data,
  visibleProperties,
  setVisibleProperties,
  reportTitle = 'Events Analytics',
  sorter,
  setSorter
}) {
  const {
    userPropNames,
    eventPropertiesDisplayNames: eventPropertiesDisplayNamesState
  } = useSelector((state) => state.coreQuery);
  const { data: eventPropertiesDisplayNames } =
    eventPropertiesDisplayNamesState;

  const [searchText, setSearchText] = useState('');

  const handleSorting = useCallback(
    (prop) => {
      setSorter((currentSorter) => {
        return getNewSorterState(currentSorter, prop);
      });
    },
    [setSorter]
  );

  const columns = getTableColumns(
    breakdown,
    sorter,
    handleSorting,
    userPropNames,
    eventPropertiesDisplayNames
  );
  const tableData = getTableData(data, breakdown, searchText, sorter);

  const getCSVData = () => {
    return {
      fileName: `${reportTitle}.csv`,
      data: tableData.map(({ index, ...rest }) => {
        return { ...rest };
      })
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

  const selectedRowKeys = visibleProperties.map((elem) => elem.index);

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
      getCSVData={getCSVData}
    />
  );
}

export default EventBreakdownTable;
