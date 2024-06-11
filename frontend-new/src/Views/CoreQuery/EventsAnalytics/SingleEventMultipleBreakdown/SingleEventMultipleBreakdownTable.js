import React, { useState, useCallback, useEffect, useContext } from 'react';
import { useSelector } from 'react-redux';
import find from 'lodash/find';
import {
  getTableColumns,
  getDataInTableFormat,
  getDateBasedColumns,
  getDateBasedTableData,
  formatData,
  formatDataInStackedAreaFormat
} from './utils';
import DataTable from '../../../../components/DataTable';
import {
  MAX_ALLOWED_VISIBLE_PROPERTIES,
  DASHBOARD_WIDGET_SECTION,
  CHART_TYPE_HORIZONTAL_BAR_CHART
} from '../../../../utils/constants';
import { isSeriesChart } from '../../../../utils/dataFormatter';
import { EVENT_COUNT_KEY } from '../eventsAnalytics.constants';
import {
  getEventDisplayName,
  getBreakdownDisplayName
} from '../eventsAnalytics.helpers';
import { CoreQueryContext } from 'Context/CoreQueryContext';
import {
  fetchDataCSVDownload,
  getEventsCSVData,
  getQuery
} from 'Views/CoreQuery/utils';

function SingleEventMultipleBreakdownTable({
  data,
  seriesData,
  events,
  breakdown,
  chartType,
  visibleProperties,
  setVisibleProperties,
  isWidgetModal,
  durationObj,
  categories,
  reportTitle = 'Events Analytics',
  section,
  sorter,
  handleSorting,
  dateSorter,
  handleDateSorting,
  visibleSeriesData,
  setVisibleSeriesData,
  eventGroup,
  resultState
}) {
  const [searchText, setSearchText] = useState('');
  const {
    eventNames,
    userPropNames,
    eventPropertiesDisplayNames: eventPropertiesDisplayNamesState
  } = useSelector((state) => state.coreQuery);
  const { projectDomainsList } = useSelector((state) => state.global);

  const { data: eventPropertiesDisplayNames } =
    eventPropertiesDisplayNamesState;
  const [columns, setColumns] = useState([]);
  const [dateBasedColumns, setDateBasedColumns] = useState([]);
  const [tableData, setTableData] = useState([]);
  const [dateBasedTableData, setDateBasedTableData] = useState([]);
  const coreQueryContext = useContext(CoreQueryContext);
  const { show_criteria: result_criteria, performance_criteria: user_type } =
    useSelector((state) => state.analyticsQuery);
  const { active_project } = useSelector((state) => state.global);
  const tableDataSelector = useCallback(
    (data) =>
      isSeriesChart(chartType)
        ? getDateBasedTableData(data, searchText, dateSorter)
        : getDataInTableFormat(data, searchText, sorter),
    []
  );
  const getCSVDataCallback = useCallback(
    (data) =>
      data.map(({ index, label, value, name, data, marker, ...rest }) => {
        const result = {};
        for (const key in rest) {
          if (key === EVENT_COUNT_KEY) {
            result[getEventDisplayName({ eventNames, event: events[0] })] =
              rest[EVENT_COUNT_KEY];
            continue;
          }
          if (key === events[0]) {
            result[getEventDisplayName({ eventNames, event: events[0] })] =
              rest[events[0]];
            continue;
          }
          const isCurrentKeyForBreakdown = find(
            breakdown,
            (b, index) => `${b.property} - ${index}` === key
          );
          if (isCurrentKeyForBreakdown) {
            result[
              `${getBreakdownDisplayName({
                breakdown: isCurrentKeyForBreakdown,
                userPropNames,
                eventPropertiesDisplayNames
              })} - ${key.split(' - ')[1]}`
            ] = rest[key];
            continue;
          }
          result[key] = rest[key];
        }
        return result;
      }),
    [eventNames, events, chartType]
  );
  const fetchData = async () => {
    const results = await fetchDataCSVDownload(coreQueryContext, {
      projectID: active_project?.id,
      step2Properties: {
        result_criteria,
        user_type,
        shouldStack: !isSeriesChart(chartType),
        durationObj,
        resultState
      },
      formatData,
      formatDataBasedOnChart: formatDataInStackedAreaFormat,
      tableDataSelector,
      getCSVDataCallback,
      EventBreakDownType: 'semb',
      formatDataParams: {},
      formatDataBasedOnChartParams: {}
    });
    return results;
  };
  const getCSVData = async () => {
    const results = await fetchData();
    return {
      fileName: reportTitle,
      data: results
    };
  };
  useEffect(() => {
    setColumns(
      getTableColumns(
        events,
        breakdown,
        sorter,
        handleSorting,
        eventNames,
        userPropNames,
        eventPropertiesDisplayNames,
        projectDomainsList,
        eventGroup
      )
    );
  }, [
    events,
    breakdown,
    sorter,
    handleSorting,
    eventNames,
    userPropNames,
    eventPropertiesDisplayNames,
    projectDomainsList,
    eventGroup
  ]);

  useEffect(() => {
    setTableData(getDataInTableFormat(data, searchText, sorter));
  }, [data, searchText, sorter]);

  useEffect(() => {
    setDateBasedColumns(
      getDateBasedColumns(
        categories,
        breakdown,
        dateSorter,
        handleDateSorting,
        durationObj.frequency,
        userPropNames,
        eventPropertiesDisplayNames,
        projectDomainsList
      )
    );
  }, [
    categories,
    breakdown,
    dateSorter,
    durationObj.frequency,
    handleDateSorting,
    userPropNames,
    eventPropertiesDisplayNames,
    projectDomainsList
  ]);

  useEffect(() => {
    setDateBasedTableData(
      getDateBasedTableData(seriesData, searchText, dateSorter)
    );
  }, [seriesData, searchText, dateSorter]);

  const selectedRowKeys = useCallback((rows) => rows.map((vp) => vp.index), []);

  const onSelectionChange = useCallback(
    (_, selectedRows) => {
      if (
        selectedRows.length > MAX_ALLOWED_VISIBLE_PROPERTIES ||
        !selectedRows.length
      ) {
        return false;
      }
      const newVisibleProperties = selectedRows.map((elem) => {
        const obj = data.find((d) => d.index === elem.index);
        return obj;
      });
      setVisibleProperties(newVisibleProperties);
    },
    [setVisibleProperties, data]
  );

  const onDateWiseSelectionChange = useCallback(
    (_, selectedRows) => {
      if (
        selectedRows.length > MAX_ALLOWED_VISIBLE_PROPERTIES ||
        !selectedRows.length
      ) {
        return false;
      }
      const newVisibleSeriesData = selectedRows.map((elem) => {
        const obj = seriesData.find((d) => d.index === elem.index);
        return obj;
      });
      setVisibleSeriesData(newVisibleSeriesData);
    },
    [setVisibleSeriesData, seriesData]
  );

  const rowSelection = {
    selectedRowKeys: isSeriesChart(chartType)
      ? selectedRowKeys(visibleSeriesData)
      : selectedRowKeys(visibleProperties),
    onChange: isSeriesChart(chartType)
      ? onDateWiseSelectionChange
      : onSelectionChange
  };

  return (
    <DataTable
      isWidgetModal={isWidgetModal}
      tableData={
        !isSeriesChart(chartType) || section === DASHBOARD_WIDGET_SECTION
          ? tableData
          : dateBasedTableData
      }
      searchText={searchText}
      setSearchText={setSearchText}
      columns={
        !isSeriesChart(chartType) || section === DASHBOARD_WIDGET_SECTION
          ? columns
          : dateBasedColumns
      }
      rowSelection={
        chartType !== CHART_TYPE_HORIZONTAL_BAR_CHART ? rowSelection : null
      }
      scroll={{ x: 250 }}
      getCSVData={getCSVData}
    />
  );
}

export default React.memo(SingleEventMultipleBreakdownTable);
