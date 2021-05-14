import React, { useState, useCallback, useMemo } from 'react';
import {
  getTableColumns,
  getTableData,
  getDateBasedColumns,
  getDateBasedTableData,
} from './utils';
import DataTable from '../../../../components/DataTable';
import {
  CHART_TYPE_BARCHART,
  MAX_ALLOWED_VISIBLE_PROPERTIES,
  DASHBOARD_WIDGET_SECTION,
} from '../../../../utils/constants';
import { useSelector } from 'react-redux';

function MultipleEventsWithBreakdownTable({
  chartType,
  breakdown,
  data,
  seriesData,
  categories,
  visibleProperties,
  setVisibleProperties,
  page,
  isWidgetModal,
  durationObj,
  reportTitle = 'Events Analytics',
  section,
}) {
  const [sorter, setSorter] = useState({});
  const [dateSorter, setDateSorter] = useState({});
  const [searchText, setSearchText] = useState('');
  const { eventNames } = useSelector((state) => state.coreQuery);

  const handleSorting = useCallback((sorter) => {
    setSorter(sorter);
  }, []);

  const handleDateSorting = useCallback((sorter) => {
    setDateSorter(sorter);
  }, []);

  const columns = useMemo(() => {
    return getTableColumns(breakdown, sorter, handleSorting, page, eventNames);
  }, [breakdown, sorter, handleSorting, page, eventNames]);

  const tableData = useMemo(() => {
    return getTableData(data, breakdown, searchText, sorter);
  }, [data, breakdown, searchText, sorter]);

  const dateBasedColumns = useMemo(() => {
    return getDateBasedColumns(
      categories,
      breakdown,
      dateSorter,
      handleDateSorting,
      durationObj.frequency,
    );
  }, [
    categories,
    breakdown,
    dateSorter,
    handleDateSorting,
    durationObj.frequency,
  ]);

  const dateBasedTableData = useMemo(() => {
    return getDateBasedTableData(
      seriesData,
      categories,
      breakdown,
      dateSorter,
      searchText,
      durationObj.frequency
    );
  }, [
    breakdown,
    categories,
    dateSorter,
    durationObj.frequency,
    searchText,
    seriesData,
  ]);

  const getCSVData = () => {
    const activeTableData =
      chartType === CHART_TYPE_BARCHART ? tableData : dateBasedTableData;
    return {
      fileName: `${reportTitle}.csv`,
      data: activeTableData.map(
        ({ index, eventIndex, dateWise, color, label, value, ...rest }) => {
          const result = {};
          for (let obj in rest) {
            if (obj.toLowerCase() === 'event') {
              result['Event'] = eventNames[rest[obj]] || rest[obj];
            } else {
              result[obj.split(';')[0]] = rest[obj];
            }
          }
          return result;
        }
      ),
    };
  };

  const onSelectionChange = (selectedIncices) => {
    if (selectedIncices.length > MAX_ALLOWED_VISIBLE_PROPERTIES) {
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

  const selectedRowKeys = useMemo(() => {
    return visibleProperties.map((vp) => vp.index);
  }, [visibleProperties]);

  const rowSelection = {
    selectedRowKeys,
    onChange: onSelectionChange,
  };

  return (
    <DataTable
      isWidgetModal={isWidgetModal}
      tableData={
        chartType === CHART_TYPE_BARCHART ||
        section === DASHBOARD_WIDGET_SECTION
          ? tableData
          : dateBasedTableData
      }
      searchText={searchText}
      setSearchText={setSearchText}
      columns={
        chartType === CHART_TYPE_BARCHART ||
        section === DASHBOARD_WIDGET_SECTION
          ? columns
          : dateBasedColumns
      }
      rowSelection={rowSelection}
      scroll={{ x: 250 }}
      getCSVData={getCSVData}
    />
  );
}

export default MultipleEventsWithBreakdownTable;
