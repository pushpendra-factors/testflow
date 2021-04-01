import React, { useState, useCallback } from "react";
import moment from "moment";
import { CHART_TYPE_SPARKLINES } from "../../../../utils/constants";
import {
  getTableColumns,
  getTableData,
  getDateBaseTableColumns,
  getDateBasedTableData,
} from "./utils";
import DataTable from "../../../../components/DataTable";

function NoBreakdownTable({
  chartsData,
  chartType,
  isWidgetModal,
  frequency,
  reportTitle = "CampaignAnalytics",
}) {
  let columns = [],
    data = [];
  const [sorter, setSorter] = useState({});
  const [searchText, setSearchText] = useState("");

  const handleSorting = useCallback((sorter) => {
    setSorter(sorter);
  }, []);

  const getCSVData = () => {
    return {
      fileName: `${reportTitle}.csv`,
      data: data.map(({ index, date, ...rest }) => {
        if (chartType === CHART_TYPE_SPARKLINES) {
          let format = "MMM D, YYYY";
          return {
            date: moment(date).format(format),
            ...rest,
          };
        }
        return rest;
      }),
    };
  };

  if (chartType === CHART_TYPE_SPARKLINES) {
    columns = getTableColumns(chartsData, frequency, sorter, handleSorting);
    data = getTableData(chartsData, sorter);
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
      getCSVData={getCSVData}
      // rowSelection={rowSelection}
    />
  );
}

export default NoBreakdownTable;
