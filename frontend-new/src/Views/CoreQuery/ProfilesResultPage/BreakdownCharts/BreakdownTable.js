import React, {
  useState, useCallback, useMemo, useEffect
} from 'react';
import find from 'lodash/find';
import { getTableColumns, getTableData } from './utils';
import DataTable from '../../../../components/DataTable';
import {
  MAX_ALLOWED_VISIBLE_PROPERTIES,
  CHART_TYPE_HORIZONTAL_BAR_CHART
} from '../../../../utils/constants';
import { useSelector } from 'react-redux';
import { getBreakdownDisplayName } from '../../EventsAnalytics/eventsAnalytics.helpers';

function BreakdownTable({
  aggregateData,
  queries,
  breakdown,
  groupAnalysis,
  currentEventIndex,
  chartType,
  isWidgetModal,
  visibleProperties,
  setVisibleProperties,
  reportTitle = 'Profile Analytics',
  handleSorting,
  sorter
}) {
  const [searchText, setSearchText] = useState('');
  const [columns, setColumns] = useState([]);
  const [tableData, setTableData] = useState([]);

  const { userPropNames, eventPropNames } = useSelector(
    (state) => state.coreQuery
  );

  useEffect(() => {
    setColumns(
      getTableColumns(
        queries,
        breakdown,
        groupAnalysis,
        currentEventIndex,
        sorter,
        handleSorting,
        eventPropNames,
        userPropNames
      )
    );
  }, [
    queries,
    breakdown,
    currentEventIndex,
    sorter,
    handleSorting,
    eventPropNames,
    userPropNames
  ]);

  useEffect(() => {
    setTableData(
      getTableData(
        aggregateData,
        searchText,
        sorter,
        queries,
        currentEventIndex,
        groupAnalysis
      )
    );
  }, [
    aggregateData,
    searchText,
    sorter,
    queries,
    currentEventIndex,
    groupAnalysis
  ]);

  const getCSVData = () => {
    return {
      fileName: `${reportTitle}.csv`,
      data: tableData.map(({
        index, color, label, ...rest
      }) => {
        const result = {};
        for (const key in rest) {
          const isCurrentKeyForBreakdown = find(
            breakdown,
            (b, index) => b.property + ' - ' + index === key
          );
          if (isCurrentKeyForBreakdown) {
            result[
              `${getBreakdownDisplayName({
                breakdown: isCurrentKeyForBreakdown,
                userPropNames,
                eventPropNames
              })} - ${key.split(' - ')[1]}`
            ] = rest[key];
            continue;
          }
          result[key] = rest[key];
        }
        return result;
      })
    };
  };

  const selectedRowKeys = useMemo(() => {
    return visibleProperties.map((vp) => vp.index);
  }, [visibleProperties]);

  const onSelectionChange = useCallback(
    (_, selectedRows) => {
      if (
        selectedRows.length > MAX_ALLOWED_VISIBLE_PROPERTIES ||
        !selectedRows.length
      ) {
        return false;
      }
      const newVisibleProperties = selectedRows.map((elem) => {
        const obj = aggregateData.find((d) => d.index === elem.index);
        return obj;
      });
      setVisibleProperties(newVisibleProperties);
    },
    [setVisibleProperties, aggregateData]
  );

  const rowSelection =
    chartType !== CHART_TYPE_HORIZONTAL_BAR_CHART
      ? {
        selectedRowKeys,
        onChange: onSelectionChange
      }
      : null;

  return (
    <DataTable
      isWidgetModal={isWidgetModal}
      tableData={tableData}
      searchText={searchText}
      setSearchText={setSearchText}
      columns={columns}
      scroll={{ x: 250 }}
      rowSelection={rowSelection}
      getCSVData={getCSVData}
    />
  );
}

export default BreakdownTable;
