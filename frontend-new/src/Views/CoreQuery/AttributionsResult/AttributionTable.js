import React, { useState, useCallback } from "react";
import moment from 'moment';
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
  durationObj,
  cmprDuration
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

  const constructCompareCSV = (rst) => {
      const keys = Object.keys(rst);
      const tbl = {};
      keys.forEach((k, ind) => {
        if(ind){
          const firstDateString = {
            from: moment(durationObj.from).toDate().toLocaleDateString(),
            to: moment(durationObj.to).toDate().toLocaleDateString()
          }
          const secondDateString = {
            from: moment(durationObj.from).toDate().toLocaleDateString(),
            to: moment(durationObj.to).toDate().toLocaleDateString()
          }
          
          tbl[`${k} (${firstDateString.from} to ${firstDateString.to})`] = rst[k].first;
          tbl[`${k} (${secondDateString.from} to ${secondDateString.to})`] = rst[k].second;
          tbl[`${k} Change`] = rst[k].change;
        } else {
          tbl[k] = rst[k];
        }
      })
      return tbl;
  }

  const getCSVData = () => {
    const dt = cmrTableData ? cmrTableData : tableData;
    return {
      fileName: `${reportTitle}.csv`,
      data: dt.map(({ index, ...rest }) => {
        let results;
        if(cmrTableData) {
          results = constructCompareCSV(rest);
        } else {
          results = rest;
        }
        return results;
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
