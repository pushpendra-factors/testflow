import React, { useState, useCallback, useContext } from 'react';
import MomentTz from 'Components/MomentTz';
import DataTable from '../../../../components/DataTable';
import {
  getTableData,
  getColumns,
  getDateBasedColumns,
  getDateBasedTableData,
  formatData,
  getDataInLineChartFormat
} from './utils';
import {
  CHART_TYPE_SPARKLINES,
  DASHBOARD_WIDGET_SECTION
} from '../../../../utils/constants';
import { useSelector } from 'react-redux';
import {
  addQforQuarter,
  getNewSorterState
} from '../../../../utils/dataFormatter';
import { getEventDisplayName } from '../eventsAnalytics.helpers';
import {
  fetchDataCSVDownload,
  getEventsCSVData,
  getQuery
} from 'Views/CoreQuery/utils';
import { CoreQueryContext } from 'Context/CoreQueryContext';

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
  section,
  comparisonApplied = false,
  resultState
}) {
  const [searchText, setSearchText] = useState('');
  const { eventNames } = useSelector((state) => state.coreQuery);
  const coreQueryContext = useContext(CoreQueryContext);
  const { show_criteria: result_criteria, performance_criteria: user_type } =
    useSelector((state) => state.analyticsQuery);
  const { active_project } = useSelector((state) => state.global);
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
      eventNames,
      comparisonApplied
    );
    tableData = getTableData({ data, currentSorter: sorter });
  } else {
    columns = getDateBasedColumns(
      data,
      dateSorter,
      handleDateSorting,
      durationObj.frequency,
      eventNames,
      comparisonApplied
    );
    tableData = getDateBasedTableData(
      data,
      dateSorter,
      searchText,
      arrayMapper,
      durationObj.frequency,
      responseData.metrics,
      comparisonApplied
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

  const tableDataSelector = (data) =>
    chartType === CHART_TYPE_SPARKLINES
      ? getTableData({ data, currentSorter: sorter })
      : getDateBasedTableData(
          data,
          dateSorter,
          searchText,
          arrayMapper,
          durationObj.frequency,
          responseData.metrics,
          comparisonApplied
        );
  const fetchData = async () => {
    const results = await fetchDataCSVDownload(coreQueryContext, {
      projectID: active_project?.id,
      step2Properties: {
        result_criteria,
        user_type,
        shouldStack:
          chartType === CHART_TYPE_SPARKLINES ||
          section === DASHBOARD_WIDGET_SECTION,
        durationObj,
        resultState
      },
      formatData,
      formatDataBasedOnChart: (a) => ({ data: a }),
      tableDataSelector,
      getCSVDataCallback,
      EventBreakDownType: 'nob',
      formatDataParams: { arrayMapper, eventNames },
      formatDataBasedOnChartParams: { sorter }
    });
    return results;
  };

  const getCSVDataCallback = (d) =>
    d.map(({ index, date, event, ...rest }) => {
      if (chartType === CHART_TYPE_SPARKLINES) {
        let format = 'MMM D, YYYY';
        if (durationObj.frequency === 'hour') {
          format = 'h A, MMM D';
        }
        const eventsData = {};
        for (const key in rest) {
          const mapper = arrayMapper.find((elem) => elem.mapper === key);
          if (mapper) {
            const displayName = getEventDisplayName({
              event: mapper.eventName,
              eventNames
            });

            eventsData[`${displayName} - ${mapper.index}`] = rest[key];
            if (`${key} - compareValue` in rest) {
              eventsData[`${displayName} - ${mapper.index} compareValue`] =
                rest[`${key} - compareValue`];
            }
            if (`${key} - change` in rest) {
              eventsData[`${displayName} - ${mapper.index} change`] =
                rest[`${key} - change`];
            }
          }
        }
        return {
          date:
            addQforQuarter(durationObj.frequency) +
            MomentTz(date).format(format),
          ...eventsData
        };
      } else {
        return {
          Event: getEventDisplayName({ eventNames, event }),
          ...rest
        };
      }
    });
  const getCSVData = async () => {
    let csvData = [];
    csvData = await fetchData();
    return {
      fileName: reportTitle,
      data: csvData
    };
  };
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
