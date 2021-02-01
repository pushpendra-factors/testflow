import React, { useEffect, useState, useCallback } from "react";
import {
  getTableColumns,
  getTableData,
  getDateBasedColumns,
  getDateBasedTableData,
} from "./utils";
import DataTable from "../../../../components/DataTable";

function MultipleEventsWithBreakdownTable({
  chartType,
  breakdown,
  data,
  visibleProperties,
  setVisibleProperties,
  maxAllowedVisibleProperties,
  page,
  lineChartData,
  isWidgetModal,
  durationObj,
  reportTitle="Events Analytics"
}) {
  const [sorter, setSorter] = useState({});
  const [searchText, setSearchText] = useState("");

  useEffect(() => {
    // reset sorter on change of chart type
    setSorter({});
  }, [chartType]);

  const handleSorting = useCallback((sorter) => {
    setSorter(sorter);
  }, []);

  let columns, tableData;

  const getCSVData = () => {
    return {
      fileName: `${reportTitle}.csv`,
      data: tableData.map(({ index, ...rest }) => {
        // if (breakdown.length) {
        //   arrayMapper.forEach((elem) => {
        //     rest[elem.eventName] = rest[`${elem.mapper}-${elem.index}`];
        //     delete rest[`${elem.mapper}-${elem.index}`];
        //   });
        // }
        return { ...rest };
      }),
    };
  };

  if (chartType === "linechart") {
    columns = getDateBasedColumns(
      lineChartData,
      breakdown,
      sorter,
      handleSorting,
      durationObj.frequency
    );
    tableData = getDateBasedTableData(
      data,
      breakdown,
      sorter,
      searchText,
      durationObj.frequency
    );
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

export default MultipleEventsWithBreakdownTable;
