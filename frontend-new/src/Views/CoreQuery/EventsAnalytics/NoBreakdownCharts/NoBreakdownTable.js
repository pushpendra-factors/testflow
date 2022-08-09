import React, { useState, useCallback } from 'react';
import moment from 'moment';
import DataTable from '../../../../components/DataTable';
import {
  getNoGroupingTableData,
  getColumns,
  getDateBasedColumns,
  getNoGroupingTablularDatesBasedData
} from './utils';
import { CHART_TYPE_SPARKLINES, DASHBOARD_WIDGET_SECTION } from '../../../../utils/constants';
import { useSelector } from 'react-redux';
import {
  addQforQuarter,
  getNewSorterState
} from '../../../../utils/dataFormatter';
import { getEventDisplayName } from '../eventsAnalytics.helpers';

function NoBreakdownTable({
  data,
  events,
  chartType,
  setHiddenEvents,
  hiddenEvents,
  isWidgetModal,
  durationObj,
  arrayMapper,
  reportTitle = 'Events Analytics',
  sorter,
  setSorter,
  dateSorter,
  setDateSorter,
  responseData,
  section
}) {
  const [searchText, setSearchText] = useState('');
  const { eventNames } = useSelector((state) => state.coreQuery);

  const handleSorting = useCallback(
    (prop) => {
      setSorter((currentSorter) => {
        return getNewSorterState(currentSorter, prop);
      });
    },
    [setSorter]
  );

  const handleDateSorting = useCallback(
    (prop) => {
      setDateSorter((currentSorter) => {
        return getNewSorterState(currentSorter, prop);
      });
    },
    [setDateSorter]
  );

  let columns;
  let tableData;
  let rowSelection = null;
  let onSelectionChange;
  let selectedRowKeys;

  if (
    chartType === CHART_TYPE_SPARKLINES ||
    section === DASHBOARD_WIDGET_SECTION
  ) {
    columns = getColumns(
      events,
      arrayMapper,
      durationObj.frequency,
      sorter,
      handleSorting,
      eventNames
    );
    tableData = getNoGroupingTableData(data, arrayMapper, sorter);
  } else {
    columns = getDateBasedColumns(
      data,
      dateSorter,
      handleDateSorting,
      durationObj.frequency,
      eventNames
    );
    tableData = getNoGroupingTablularDatesBasedData(
      data,
      dateSorter,
      searchText,
      arrayMapper,
      durationObj.frequency,
      responseData.metrics
    );

    onSelectionChange = (_, selectedRows) => {
      const skippedEvents = events.filter(
        (event) => selectedRows.findIndex((r) => r.event === event) === -1
      );
      if (skippedEvents.length === events.length) {
        return false;
      }
      setHiddenEvents(skippedEvents);
    };

    selectedRowKeys = [];

    events.forEach((event, index) => {
      if (hiddenEvents.indexOf(event) === -1) {
        selectedRowKeys.push(index);
      }
    });

    rowSelection = {
      selectedRowKeys,
      onChange: onSelectionChange
    };
  }

  const getCSVData = useCallback(() => {
    return {
      fileName: `${reportTitle}.csv`,
      data: tableData.map(({ index, date, event, ...rest }) => {
        if (chartType === CHART_TYPE_SPARKLINES) {
          let format = 'MMM D, YYYY';
          if (durationObj.frequency === 'hour') {
            format = 'h A, MMM D';
          }
          const eventsData = {};
          for (const key in rest) {
            const mapper = arrayMapper.find((elem) => elem.mapper === key);
            if (mapper) {
              eventsData[
                `${getEventDisplayName({
                  event: mapper.eventName,
                  eventNames
                })} - ${mapper.index}`
              ] = rest[key];
            }
          }
          return {
            date:
              addQforQuarter(durationObj.frequency) +
              moment(date).format(format),
            ...eventsData
          };
        } else {
          return {
            Event: getEventDisplayName({ eventNames, event }),
            ...rest
          };
        }
      })
    };
  }, [tableData, chartType, arrayMapper, eventNames, durationObj, reportTitle]);

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

export default NoBreakdownTable;
