/* eslint-disable camelcase */
import React, { useCallback, useState, useMemo, useEffect } from 'react';
import moment from 'moment';
import _ from 'lodash';
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
  setSorter,
  tableConfig,
  tableConfigPopoverContent,
  isBreakdownApplied = false
}) {
  const eventsCondition = _.get(resultData, 'meta.query.ec', '');
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
        comparisonChartData != null,
        resultData,
        userPropNames,
        eventPropertiesDisplayNames,
        tableConfig
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
    eventPropertiesDisplayNames,
    tableConfig
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
        isBreakdownApplied,
        eventsCondition
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
    comparisonChartDurations,
    comparisonChartData,
    durationObj,
    comparison_duration,
    isBreakdownApplied,
    eventsCondition
  ]);

  const getCSVData = () => {
    try {
      if (!comparisonChartData || isBreakdownApplied) {
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
        tableData.forEach(({ index, ...remaining }) => {
          const rest = {};
          rest.Users = 'All';

          rest[`Conversion (${duration_from} - ${duration_to})`] =
            remaining.Conversion.conversion;
          rest[
            `Conversion (${compare_duration_from} - ${compare_duration_to})`
          ] = remaining.Conversion.comparison_conversion;

          rest[`Conversion Time (${duration_from} - ${duration_to})`] =
            remaining['Conversion Time'].overallDuration;
          rest[
            `Conversion Time (${compare_duration_from} - ${compare_duration_to})`
          ] = remaining['Conversion Time'].comparisonOverallDuration;

          arrayMapper.forEach((elem, index) => {
            rest[
              `${elem.displayName}-${index} (${duration_from} - ${duration_to})`
            ] = remaining[`${elem.displayName}-${index}-count`].count;
            rest[
              `${elem.displayName}-${index} (${compare_duration_from} - ${compare_duration_to})`
            ] = remaining[`${elem.displayName}-${index}-count`].compare_count;

            if (index < arrayMapper.length - 1) {
              rest[
                `time[${index}-${
                  index + 1
                }] (${duration_from} - ${duration_to})`
              ] = remaining[`time[${index}-${index + 1}]`].time;
              rest[
                `time[${index}-${
                  index + 1
                }] (${compare_duration_from} - ${compare_duration_to})`
              ] = remaining[`time[${index}-${index + 1}]`].compare_time;
            }
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
      controlsPopover={tableConfigPopoverContent}
    />
  );
}

export default FunnelsResultTable;
