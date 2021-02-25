import React, { useState, useCallback } from "react";
import { getCompareTableColumns, getCompareTableData, getTableColumns, getTableData } from "./utils";
import DataTable from "../../../components/DataTable";

function AttributionTable({
  data,
  data2,
  isWidgetModal,
  event,
  setVisibleIndices,
  visibleIndices,
  maxAllowedVisibleProperties,
  attribution_method,
  attribution_method_compare,
  touchpoint,
  linkedEvents,
  reportTitle = "Attributions",
}) {
  const [searchText, setSearchText] = useState("");
  const [sorter, setSorter] = useState({});
  const handleSorting = useCallback((sorter) => {
    setSorter(sorter);
  }, []);
  const columns = getTableColumns(
    sorter,
    handleSorting,
    attribution_method,
    attribution_method_compare,
    touchpoint,
    linkedEvents,
    event
  );

  const cmprColums = data2? getCompareTableColumns(sorter,
    handleSorting,
    attribution_method,
    attribution_method_compare,
    touchpoint,
    linkedEvents,
    event) : null;

  const tableData = getTableData(data, event, searchText, sorter, attribution_method_compare, touchpoint, linkedEvents);

  const cmrTableData = data2 ? 
    getCompareTableData(data, data2, event, searchText, sorter, attribution_method_compare, touchpoint, linkedEvents) 
    : null;

  const getCSVData = () => {
    return {
      fileName: `${reportTitle}.csv`,
      data: tableData.map(({ index, ...rest }) => {
        return rest;
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
    selectedIncices.sort();
    setVisibleIndices(selectedIncices);
  };

  const rowSelection = {
    selectedRowKeys: visibleIndices,
    onChange: onSelectionChange,
  };

  return (
    <DataTable
      isWidgetModal={isWidgetModal}
      tableData={cmrTableData? cmrTableData: tableData}
      searchText={searchText}
      setSearchText={setSearchText}
      columns={cmprColums? cmprColums : columns}
      rowSelection={rowSelection}
      scroll={{ x: 250 }}
      getCSVData={getCSVData}
    />
  );
}

export default AttributionTable;
