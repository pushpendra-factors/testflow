import React, { useState, useCallback } from "react";
import { CHART_TYPE_BARCHART } from "../../../../utils/constants";
import { getTableColumns, getTableData } from "./utils";
import DataTable from "../../../../components/DataTable";

function BreakdownTable({
  chartsData,
  breakdown,
  currentEventIndex,
  chartType,
  arrayMapper,
  isWidgetModal,
  responseData,
  frequency,
  visibleProperties,
  maxAllowedVisibleProperties,
  setVisibleProperties,
  reportTitle = "CampaignAnalytics",
}) {
  let columns, data;
  const [sorter, setSorter] = useState({});
  const [searchText, setSearchText] = useState("");

  const handleSorting = useCallback((sorter) => {
    setSorter(sorter);
  }, []);

  const getCSVData = () => {
    return {
      fileName: `${reportTitle}.csv`,
      data: data.map(({ index, ...rest }) => {
        return rest;
      }),
    };
  };

  if (true) {
    columns = getTableColumns(
      responseData,
      breakdown,
      arrayMapper,
      sorter,
      handleSorting
    );
    data = getTableData(
      responseData,
      breakdown,
      currentEventIndex,
      arrayMapper,
      sorter,
      searchText
    );
  } else {
    columns = [];
    data = [];
  }

  const onSelectionChange = (visibleIndices, b) => {
    if (visibleIndices.length > maxAllowedVisibleProperties) {
      return false;
    }
    if (!visibleIndices.length) {
      return false;
    }
    const newVisibleProperties = chartsData.filter(
      (elem) => visibleIndices.indexOf(elem.index) > -1
    );
    setVisibleProperties(newVisibleProperties);
  };

  const rowSelection = {
    selectedRowKeys: visibleProperties.map((v) => v.index),
    onChange: onSelectionChange,
  };

  return (
    <DataTable
      isWidgetModal={isWidgetModal}
      tableData={data}
      searchText={searchText}
      setSearchText={setSearchText}
      columns={columns}
      scroll={{ x: 250 }}
      rowSelection={rowSelection}
      getCSVData={getCSVData}
    />
  );
}

export default BreakdownTable;
