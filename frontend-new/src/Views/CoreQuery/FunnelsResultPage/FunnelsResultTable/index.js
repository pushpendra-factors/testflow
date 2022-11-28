/* eslint-disable camelcase */
import React, { useCallback, useState, useMemo, useEffect } from 'react';
import moment from 'moment';
import find from 'lodash/find';
import { getTableColumns, getTableData } from '../utils';
import DataTable from '../../../../components/DataTable';
import { GROUPED_MAX_ALLOWED_VISIBLE_PROPERTIES } from '../../../../utils/constants';
import { getNewSorterState } from '../../../../utils/dataFormatter';
import { useSelector } from 'react-redux';
import { getBreakdownDisplayName } from '../../EventsAnalytics/eventsAnalytics.helpers';

function FunnelsResultTable({
  breakdown,
  visibleProperties,
  setVisibleProperties,
  queries,
  groups,
  arrayMapper,
  reportTitle = 'FunnelAnalysis',
  chartData,
  durations,
  comparisonChartData,
  comparisonChartDurations,
  durationObj,
  comparison_duration,
  resultData,
  sorter,
  setSorter
}) {
  const [columns, setColumns] = useState([]);
  const [tableData, setTableData] = useState([]);
  const [searchText, setSearchText] = useState('');
  const {
    userPropNames,
    eventPropertiesDisplayNames: eventPropertiesDisplayNamesState
  } = useSelector((state) => state.coreQuery);

  const { data: eventPropertiesDisplayNames } =
    eventPropertiesDisplayNamesState;

  const handleSorting = useCallback(
    (prop) => {
      setSorter((currentSorter) => {
        return getNewSorterState(currentSorter, prop);
      });
    },
    [setSorter]
  );

  useEffect(() => {
    setColumns(
      getTableColumns(
        queries,
        sorter,
        handleSorting,
        arrayMapper,
        comparisonChartData,
        resultData,
        userPropNames,
        eventPropertiesDisplayNames
      )
    );
  }, [
    queries,
    sorter,
    handleSorting,
    arrayMapper,
    comparisonChartData,
    resultData,
    userPropNames,
    eventPropertiesDisplayNames
  ]);

  // const columns = ;
  useEffect(() => {
    setTableData(
      getTableData(
        chartData,
        queries,
        groups,
        arrayMapper,
        sorter,
        searchText,
        durations,
        comparisonChartDurations,
        comparisonChartData,
        durationObj,
        comparison_duration,
        resultData
      )
    );
  }, [
    chartData,
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
    resultData
  ]);

  const getCSVData = () => {
    try {
      if (!comparisonChartData) {
        return {
          fileName: `${reportTitle}.csv`,
          data: tableData.map(
            ({ index, value, name, nonConvertedName, ...rest }) => {
              arrayMapper.forEach((elem) => {
                delete rest[`${elem.mapper}`];
              });
              const result = {};
              for (const key in rest) {
                if (key === 'Conversion') {
                  result[key] = rest[key] + '%';
                  continue;
                }
                const isCurrentKeyForBreakdown = find(breakdown, (b) => {
                  return b.property + ' - ' + b.eventIndex === key;
                });
                if (isCurrentKeyForBreakdown) {
                  result[
                    `${getBreakdownDisplayName({
                      breakdown: isCurrentKeyForBreakdown,
                      userPropNames,
                      eventPropertiesDisplayNames
                    })} - ${isCurrentKeyForBreakdown.overAllIndex}`
                  ] = rest[key];
                  continue;
                }
                result[key] = rest[key];
              }
              return result;
            }
          )
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
          rest.Users = 'All';

          rest[`Conversion (${duration_from} - ${duration_to})`] =
            rest.Conversion.conversion;
          rest[
            `Conversion (${compare_duration_from} - ${compare_duration_to})`
          ] = rest.Conversion.comparsion_conversion;

          rest[`Conversion Time (${duration_from} - ${duration_to})`] =
            rest['Conversion Time'].overallDuration;
          rest[
            `Conversion Time (${compare_duration_from} - ${compare_duration_to})`
          ] = rest['Conversion Time'].comparisonOverallDuration;

          delete rest['Conversion Time'];
          delete rest.Conversion;
          delete rest.Grouping;

          arrayMapper.forEach((elem, index) => {
            rest[
              `${elem.displayName}-${index} (${duration_from} - ${duration_to})`
            ] = rest[`${elem.displayName}-${index}-count`].count;
            rest[
              `${elem.displayName}-${index} (${compare_duration_from} - ${compare_duration_to})`
            ] = rest[`${elem.displayName}-${index}-count`].compare_count;

            if (index < arrayMapper.length - 1) {
              rest[
                `time[${index}-${
                  index + 1
                }] (${duration_from} - ${duration_to})`
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
          data
        };
      }
    } catch (err) {
      console.log('err', err);
    }
  };

  const onSelectionChange = useCallback(
    (selectedRowKeys) => {
      if (
        !selectedRowKeys.length ||
        selectedRowKeys.length > GROUPED_MAX_ALLOWED_VISIBLE_PROPERTIES
      ) {
        return false;
      }
      setVisibleProperties(
        groups.filter((g) => selectedRowKeys.indexOf(g.index) > -1)
      );
    },
    [groups, setVisibleProperties]
  );

  const selectedRowKeys = useMemo(() => {
    if (breakdown.length) {
      return visibleProperties.map((elem) => elem.index);
    }
    return null;
  }, [visibleProperties, breakdown]);

  const rowSelection = useMemo(() => {
    return {
      selectedRowKeys,
      onChange: onSelectionChange
    };
  }, [selectedRowKeys, onSelectionChange]);

  return (
    <DataTable
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
