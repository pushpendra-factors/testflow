import React, { useState } from 'react';
import { getWebAnalyticsTableData } from './utils';
import DataTable from '../../../components/DataTable';

function WebsiteAnalyticsTable({
  tableData,
  isWidgetModal = false,
  modalTitle = null
}) {
  const [searchText, setSearchText] = useState('');

  const { columns, data } = getWebAnalyticsTableData(tableData, searchText);

  const getCSVData = () => {
    return {
      fileName: modalTitle,
      data: data.map(({ index, ...rest }) => {
        return rest;
      })
    };
  };

  return (
    <DataTable
      isWidgetModal={isWidgetModal}
      tableData={data}
      searchText={searchText}
      setSearchText={setSearchText}
      columns={columns}
      scroll={{ x: 250 }}
      getCSVData={getCSVData}
    />
  );
}

export default WebsiteAnalyticsTable;
