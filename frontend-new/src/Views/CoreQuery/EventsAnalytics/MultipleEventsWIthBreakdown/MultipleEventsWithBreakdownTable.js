import React, { useEffect, useState, useCallback } from 'react';
import {
  getTableColumns, getTableData, getDateBasedColumns, getDateBasedTableData
} from './utils';
import DataTable from '../../../../components/DataTable';

function MultipleEventsWithBreakdownTable({
  chartType, breakdown, data, visibleProperties, setVisibleProperties, maxAllowedVisibleProperties, page, lineChartData, isWidgetModal, durationObj
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

  let columns, tableData;

  if (chartType === 'linechart') {
    columns = getDateBasedColumns(lineChartData, breakdown, sorter, handleSorting, durationObj.frequency);
    tableData = getDateBasedTableData(data, breakdown, sorter, searchText, durationObj.frequency);
  } else {
    tableData = getTableData(data, breakdown, searchText, sorter);
    columns = getTableColumns(breakdown, sorter, handleSorting, page);
  }

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

export default MultipleEventsWithBreakdownTable;
