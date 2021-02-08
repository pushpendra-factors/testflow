import React, { useState, useCallback } from "react";
import { getTableColumns, getTableData } from "./utils";
import DataTable from "../../../../components/DataTable";

function EventBreakdownTable({
  breakdown,
  data,
  visibleProperties,
  setVisibleProperties,
  maxAllowedVisibleProperties,
  reportTitle = "Events Analytics",
}) {
  const [sorter, setSorter] = useState({});
  const [searchText, setSearchText] = useState("");

  const handleSorting = useCallback((sorter) => {
    setSorter(sorter);
  }, []);

  const columns = getTableColumns(breakdown, sorter, handleSorting);
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
