import React, { useCallback, useState, useMemo } from 'react';
import moment from 'moment';
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
    if (!comparisonChartData) {
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
    } else {
      const data = [];
      const duration_from = moment(durationObj.from).format('MMM DD');
      const duration_to = moment(durationObj.to).format('MMM DD');
      const compare_duration_from = moment(comparison_duration.from).format(
        'MMM DD'
      );
      const compare_duration_to = moment(comparison_duration.to).format(
        'MMM DD'
      );
      tableData.forEach(({ index, ...rest }) => {
        rest['Users'] = 'All';

        rest[`Conversion (${duration_from} - ${duration_to})`] =
          rest[`Conversion`].conversion;
        rest[`Conversion (${compare_duration_from} - ${compare_duration_to})`] =
          rest[`Conversion`].comparsion_conversion;

        rest[`Converstion Time (${duration_from} - ${duration_to})`] =
          rest[`Converstion Time`].overallDuration;
        rest[
          `Converstion Time (${compare_duration_from} - ${compare_duration_to})`
        ] = rest[`Converstion Time`].comparisonOverallDuration;

        delete rest[`Converstion Time`];
        delete rest[`Conversion`];
        delete rest['Grouping'];

        arrayMapper.forEach((elem, index) => {
          rest[
            `${elem.displayName}-${index} (${duration_from} - ${duration_to})`
          ] = rest[`${elem.mapper}`].count;
          rest[
            `${elem.displayName}-${index} (${compare_duration_from} - ${compare_duration_to})`
          ] = rest[`${elem.mapper}`].compare_count;

          if (index < arrayMapper.length - 1) {
            rest[
              `time[${index}-${index + 1}] (${duration_from} - ${duration_to})`
            ] = rest[`time[${index}-${index + 1}]`].time;
            rest[
              `time[${index}-${
                index + 1
              }] (${compare_duration_from} - ${compare_duration_to})`
            ] = rest[`time[${index}-${index + 1}]`].compare_time;
            delete rest[`time[${index}-${index + 1}]`];
          }

          delete rest[`${elem.mapper}`];
        });
        data.push(rest);
      });
      return {
        fileName: `${reportTitle}.csv`,
        data,
      };
    }
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
