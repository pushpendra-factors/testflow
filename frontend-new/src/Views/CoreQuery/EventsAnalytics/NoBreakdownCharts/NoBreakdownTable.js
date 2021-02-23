import React, { useState, useCallback } from "react";
import DataTable from "../../../../components/DataTable";
import {
  getNoGroupingTableData,
  getColumns,
  getDateBasedColumns,
  getNoGroupingTablularDatesBasedData,
} from "./utils";
import {
  CHART_TYPE_SPARKLINES,
} from "../../../../utils/constants";

function NoBreakdownTable({
  data,
  events,
  chartType,
  setHiddenEvents,
  hiddenEvents,
  isWidgetModal,
  durationObj,
  arrayMapper,
  reportTitle = "Events Analytics",
}) {
  const [sorter, setSorter] = useState({});
  const [searchText, setSearchText] = useState("");

  const handleSorting = useCallback((sorter) => {
    setSorter(sorter);
  }, []);

  let columns;
  let tableData;
  let rowSelection = null;
  let onSelectionChange;
  let selectedRowKeys;

  if (chartType === CHART_TYPE_SPARKLINES) {
    columns = getColumns(events, arrayMapper, sorter, handleSorting);
    tableData = getNoGroupingTableData(
      data,
      sorter,
      durationObj.frequency
    );
  } else {
    columns = getDateBasedColumns(
      data,
      sorter,
      handleSorting,
      durationObj.frequency
    );
    tableData = getNoGroupingTablularDatesBasedData(
      data,
      sorter,
      searchText,
      arrayMapper,
      durationObj.frequency
    );

    onSelectionChange = (_, selectedRows) => {
      const skippedEvents = events.filter(
        (event) => selectedRows.findIndex((r) => r.event === event) === -1
      );
      if (skippedEvents.length === events.length) {
        return false;
      }
      setHiddenEvents(skippedEvents);
    };

    selectedRowKeys = [];

    events.forEach((event, index) => {
      if (hiddenEvents.indexOf(event) === -1) {
        selectedRowKeys.push(index);
      }
    });

    rowSelection = {
      selectedRowKeys,
      onChange: onSelectionChange,
    };
  }

  const getCSVData = () => {
    return {
      fileName: `${reportTitle}.csv`,
      data: tableData.map(({ index, ...rest }) => {
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
      rowSelection={rowSelection}
      getCSVData={getCSVData}
    />
  );
}

export default NoBreakdownTable;
