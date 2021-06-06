import React, { useCallback, useState, useMemo } from 'react';
import { generateTableColumns, generateTableData } from '../utils';
import DataTable from '../../../../components/DataTable';

function FunnelsResultTable({
  breakdown,
  setGroups,
  queries,
  groups,
  maxAllowedVisibleProperties,
  isWidgetModal,
  arrayMapper,
  reportTitle = 'FunnelAnalysis',
  chartData,
  durations,
  comparisonChartData,
  comparisonChartDurations,
  durationObj,
  comparison_duration,
}) {
  const [sorter, setSorter] = useState({});
  const [searchText, setSearchText] = useState('');

  const handleSorting = useCallback((sorter) => {
    setSorter(sorter);
  }, []);

  const columns = useMemo(() => {
    return generateTableColumns(
      breakdown,
      queries,
      sorter,
      handleSorting,
      arrayMapper,
      comparisonChartData
    );
  }, [
    breakdown,
    queries,
    sorter,
    handleSorting,
    arrayMapper,
    comparisonChartData,
  ]);

  // const columns = ;
  const tableData = useMemo(() => {
    return generateTableData(
      chartData,
      breakdown,
      queries,
      groups,
      arrayMapper,
      sorter,
      searchText,
      durations,
      comparisonChartDurations,
      comparisonChartData,
      durationObj,
      comparison_duration
    );
  }, [
    chartData,
    breakdown,
    queries,
    groups,
    arrayMapper,
    sorter,
    searchText,
    durations,
    comparisonChartData,
    comparisonChartDurations,
    comparison_duration,
    durationObj,
  ]);

  const getCSVData = () => {
    return {
      fileName: `${reportTitle}.csv`,
      data: tableData.map(({ index, ...rest }) => {
        arrayMapper.forEach((elem, index) => {
          rest[`${elem.displayName}-${index}`] = rest[`${elem.mapper}`].count;
          delete rest[`${elem.mapper}`];
        });
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
      scroll={{ x: 250 }}
      getCSVData={getCSVData}
    />
  );
}

export default FunnelsResultTable;
