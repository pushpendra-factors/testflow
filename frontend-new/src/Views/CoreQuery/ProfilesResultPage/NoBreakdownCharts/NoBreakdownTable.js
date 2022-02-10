import React, { useState, useEffect } from 'react';
import { getTableColumns, getTableData } from './utils';
import DataTable from '../../../../components/DataTable';

function NoBreakdownTable({
  data,
  queries,
  groupAnalysis,
  reportTitle = 'Profile Analytics',
  handleSorting,
  sorter,
  isWidgetModal,
}) {
  const [searchText, setSearchText] = useState('');
  const [columns, setColumns] = useState([]);
  const [tableData, setTableData] = useState([]);

  useEffect(() => {
    setColumns(getTableColumns(sorter, handleSorting, groupAnalysis));
  }, [sorter, handleSorting]);

  useEffect(() => {
    setTableData(
      getTableData(data, queries, groupAnalysis, sorter, searchText)
    );
  }, [data, queries, groupAnalysis, searchText, sorter]);

  const getCSVData = () => {
    return {
      fileName: `${reportTitle}.csv`,
      data: tableData.map(({ index, color, ...rest }) => {
        return rest;
      }),
    };
  };

  return (
    <DataTable
      isWidgetModal={isWidgetModal}
      tableData={tableData}
      searchText={searchText}
      setSearchText={setSearchText}
      columns={columns}
      scroll={{ x: 250 }}
      rowSelection={null}
      getCSVData={getCSVData}
    />
  );
}

export default NoBreakdownTable;
