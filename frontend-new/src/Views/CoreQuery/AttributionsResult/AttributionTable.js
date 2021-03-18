import React, { useState, useCallback } from "react";
import moment from 'moment';
import { getCompareTableColumns, getCompareTableData, getTableColumns, getTableData, calcChangePerc } from "./utils";
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

  const calcTotal = (rowTtl, tblItem) => {
    if(rowTtl && tblItem !== NaN) {
      return rowTtl + tblItem;
    } else if (!rowTtl && tblItem === NaN) {
      return rowTtl;
    }else if (!rowTtl && !tblItem) {
      return 0;
    } else {
      return tblItem;
    }
  }

  const constructCompareCSV = (rst, totalRow) => {
      const keys = Object.keys(rst);
      const tbl = {};
      
      keys.forEach((k, ind) => {
        if(ind){
          const firstDateString = {
            from: moment(durationObj.from).toDate().toLocaleDateString(),
            to: moment(durationObj.to).toDate().toLocaleDateString()
          }
          const secondDateString = {
            from: moment(cmprDuration.from).toDate().toLocaleDateString(),
            to: moment(cmprDuration.to).toDate().toLocaleDateString()
          }
          const firstLabel = `${k} (${firstDateString.from} to ${firstDateString.to})`;
          const secondLabel = `${k} (${secondDateString.from} to ${secondDateString.to})`;
          const changeLabel = `${k} % Change`;
          tbl[firstLabel] = rst[k].first;
          tbl[secondLabel] = rst[k].second;
          tbl[changeLabel] = rst[k].change;
          totalRow[firstLabel] = calcTotal(totalRow[firstLabel], Number(rst[k].first));
          totalRow[secondLabel] = calcTotal(totalRow[secondLabel], Number(rst[k].second));
          totalRow[changeLabel] = calcChangePerc(totalRow[firstLabel], totalRow[secondLabel]);
        } else {
          tbl[k] = rst[k];
        }
      })
      return [tbl,totalRow];
  }

  const getCSVData = () => {
    const dt = cmrTableData ? cmrTableData : tableData;
    let dataTotal = {};
    const mappedData = dt.map(({ index, ...rest }) => {
      let results;
      if(cmrTableData) {
        [results, dataTotal] = constructCompareCSV(rest, dataTotal);
      } else {
        results = rest;
      }
      return results;
    });

    dataTotal[touchpoint] = "Total";
    mappedData.push(dataTotal);


    return {
      fileName: `${reportTitle}.csv`,
      data: mappedData
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
