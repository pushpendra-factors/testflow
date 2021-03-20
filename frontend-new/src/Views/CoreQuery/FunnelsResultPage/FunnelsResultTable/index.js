import React, { useCallback, useState } from "react";
import styles from "./index.module.scss";
import { generateTableColumns, generateTableData } from "../utils";
import DataTable from "../../../../components/DataTable";

function FunnelsResultTable({
  chartData,
  breakdown,
  setGroups,
  queries,
  groups,
  maxAllowedVisibleProperties,
  isWidgetModal,
  arrayMapper,
  reportTitle = "FunnelAnalysis",
  durations
}) {
  const [sorter, setSorter] = useState({});
  const [searchText, setSearchText] = useState("");

  const handleSorting = useCallback((sorter) => {
    setSorter(sorter);
  }, []);

  const columns = generateTableColumns(
    breakdown,
    queries,
    sorter,
    handleSorting,
    arrayMapper
  );
  const tableData = generateTableData(
    chartData,
    breakdown,
    queries,
    groups,
    arrayMapper,
    sorter,
    searchText,
    durations
  );

  const getCSVData = () => {
    return {
      fileName: `${reportTitle}.csv`,
      data: tableData.map(({ index, ...rest }) => {
        if (breakdown.length) {
          arrayMapper.forEach((elem) => {
            rest[elem.eventName] = rest[`${elem.mapper}`].count;
            delete rest[`${elem.mapper}`];
          });
        }
        return { ...rest };
      }),
    };
  };

  const onSelectionChange = (selectedRowKeys) => {
    if (
      !selectedRowKeys.length ||
      selectedRowKeys.length > maxAllowedVisibleProperties
    ) {
      return false;
    }
    setGroups((currData) => {
      return currData.map((c) => {
        if (selectedRowKeys.indexOf(c.index) > -1) {
          return { ...c, is_visible: true };
        } else {
          return { ...c, is_visible: false };
        }
      });
    });
  };

  const selectedRowKeys = groups
    .filter((elem) => elem.is_visible)
    .map((elem) => elem.index);

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
      rowSelection={breakdown.length ? rowSelection : null}
      className={styles.funnelResultsTable}
      scroll={{ x: 250 }}
      getCSVData={getCSVData}
    />
  );
}

export default FunnelsResultTable;
