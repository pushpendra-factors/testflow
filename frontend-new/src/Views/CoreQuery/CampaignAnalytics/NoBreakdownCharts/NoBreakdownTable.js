import React, { useState, useCallback } from "react";
import { CHART_TYPE_SPARKLINES } from "../../../../utils/constants";
import {
  getTableColumns,
  getTableData,
  getDateBaseTableColumns,
  getDateBasedTableData
} from "./utils";
import DataTable from "../../../../components/DataTable";

function NoBreakdownTable({
  chartsData,
  chartType,
  isWidgetModal,
  frequency,
}) {
  let columns, data;
  const [sorter, setSorter] = useState({});
  const [searchText, setSearchText] = useState("");

  const handleSorting = useCallback((sorter) => {
    setSorter(sorter);
  }, []);

  if (chartType === CHART_TYPE_SPARKLINES) {
    columns = getTableColumns(chartsData, sorter, handleSorting);
    data = getTableData(chartsData, frequency, sorter);
  } else {
    columns = getDateBaseTableColumns(
      chartsData,
      frequency,
      sorter,
      handleSorting
    );
    data = getDateBasedTableData(chartsData, frequency, sorter);
  }

  return (
    <DataTable
      isWidgetModal={isWidgetModal}
      tableData={data}
      searchText={searchText}
      setSearchText={setSearchText}
      columns={columns}
      scroll={{ x: 250 }}
      // rowSelection={rowSelection}
    />
  );
}

export default NoBreakdownTable;
