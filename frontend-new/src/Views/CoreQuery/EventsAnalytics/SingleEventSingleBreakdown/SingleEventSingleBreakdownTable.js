import React, { useState, useCallback, useEffect } from "react";
import DataTable from "../../../../components/DataTable";
import {
  getTableColumns,
  getDataInTableFormat,
  getDateBasedColumns,
  getDateBasedTableData,
} from "./utils";

function SingleEventSingleBreakdownTable({
  data,
  events,
  breakdown,
  chartType,
  visibleProperties,
  setVisibleProperties,
  maxAllowedVisibleProperties,
  lineChartData,
  originalData,
  page,
  isWidgetModal,
  durationObj,
  reportTitle = "Events Analytics",
}) {
  const appliedBreakdown = [breakdown[0].property];

  const [sorter, setSorter] = useState({});
  const [searchText, setSearchText] = useState("");

  useEffect(() => {
    // reset sorter on change of chart type
    setSorter({});
  }, [chartType]);

  const getCSVData = () => {
    return {
      fileName: `${reportTitle}.csv`,
      data: tableData.map(({ index, ...rest }) => {
        return { ...rest };
      }),
    };
  };

  const handleSorting = useCallback((sorter) => {
    setSorter(sorter);
  }, []);

  let columns;
  let tableData;

  if (chartType === "linechart") {
    columns = getDateBasedColumns(
      lineChartData,
      appliedBreakdown,
      sorter,
      handleSorting,
      durationObj.frequency
    );
    tableData = getDateBasedTableData(
      data.map((elem) => elem.label),
      originalData,
      appliedBreakdown,
      searchText,
      sorter,
      durationObj.frequency
    );
  } else {
    columns = getTableColumns(
      events,
      appliedBreakdown,
      sorter,
      handleSorting,
      page
    );
    tableData = getDataInTableFormat(
      data,
      events,
      appliedBreakdown,
      searchText,
      sorter
    );
  }

  const visibleLabels = visibleProperties.map((elem) => elem.label);

  const selectedRowKeys = [];

  tableData.forEach((elem) => {
    if (visibleLabels.indexOf(elem[appliedBreakdown[0]]) > -1) {
      selectedRowKeys.push(elem.index);
    }
  });

  const onSelectionChange = (_, selectedRows) => {
    if (
      selectedRows.length > maxAllowedVisibleProperties ||
      !selectedRows.length
    ) {
      return false;
    }
    const newVisibleProperties = selectedRows.map((elem) => {
      const obj = data.find((d) => d.label === elem[appliedBreakdown[0]]);
      return obj;
    });
    setVisibleProperties(newVisibleProperties);
  };

  const rowSelection = {
    selectedRowKeys,
    onChange: onSelectionChange,
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
      getCSVData={getCSVData}
    />
  );
}

export default SingleEventSingleBreakdownTable;
