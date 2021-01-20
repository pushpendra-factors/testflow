import React, { useState } from "react";
import { getWebAnalyticsTableData } from "./utils";
import DataTable from "../../../components/DataTable";

function WebsiteAnalyticsTable({ unit, tableData }) {
  const [searchText, setSearchText] = useState("");

  const { columns, data } = getWebAnalyticsTableData(tableData);

  return (
    <DataTable
      isWidgetModal={false}
      tableData={data}
      searchText={searchText}
      setSearchText={setSearchText}
      columns={columns}
      scroll={{ x: 250 }}
    />
  );
}

export default WebsiteAnalyticsTable;
