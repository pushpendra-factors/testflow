import React, { useState, useCallback, useEffect, useContext } from 'react';
import { useSelector } from 'react-redux';
import DataTable from '../../../../components/DataTable';
import {
  getTableColumns,
  getDataInTableFormat,
  getDateBasedColumns,
  getDateBasedTableData,
  formatDataInSeriesFormat,
  formatData
} from './utils';
import {
  CHART_TYPE_BARCHART,
  MAX_ALLOWED_VISIBLE_PROPERTIES,
  DASHBOARD_WIDGET_SECTION,
  CHART_TYPE_HORIZONTAL_BAR_CHART
} from '../../../../utils/constants';
import { isSeriesChart } from '../../../../utils/dataFormatter';
import { EVENT_COUNT_KEY } from '../eventsAnalytics.constants';
import { getEventDisplayName } from '../eventsAnalytics.helpers';
import { CoreQueryContext } from 'Context/CoreQueryContext';
import {
  fetchDataCSVDownload,
  getEventsCSVData,
  getQuery
} from 'Views/CoreQuery/utils';

function SingleEventSingleBreakdownTable({
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
  comparisonApplied,
  compareCategories,
  frequency,
  eventGroup,
  resultState
}) {
  const [searchText, setSearchText] = useState('');
  const {
    eventNames,
    userPropNames,
    eventPropertiesDisplayNames: eventPropertiesDisplayNamesState
  } = useSelector((state) => state.coreQuery);
  const { data: eventPropertiesDisplayNames } =
    eventPropertiesDisplayNamesState;
  const { projectDomainsList } = useSelector((state) => state.global);
  const [columns, setColumns] = useState([]);
  const [dateBasedColumns, setDateBasedColumns] = useState([]);
  const [tableData, setTableData] = useState([]);
  const [dateBasedTableData, setDateBasedTableData] = useState([]);
  const { show_criteria: result_criteria, performance_criteria: user_type } =
    useSelector((state) => state.analyticsQuery);
  const { active_project } = useSelector((state) => state.global);
  const coreQueryContext = useContext(CoreQueryContext);
  const activeTableData =
    chartType === CHART_TYPE_BARCHART ? tableData : dateBasedTableData;
  const tableDataSelector = useCallback(
    (data) =>
      chartType === CHART_TYPE_BARCHART
        ? getDataInTableFormat(data, searchText, sorter)
        : getDateBasedTableData(
            data,
            searchText,
            dateSorter,
            categories,
            comparisonApplied,
            compareCategories,
            frequency
          ),
    [seriesData, searchText, dateSorter, chartType]
  );

  const fetchData = async () => {
    const results = await fetchDataCSVDownload(coreQueryContext, {
      projectID: active_project?.id,
      step2Properties: {
        result_criteria,
        user_type,
        shouldStack: chartType === CHART_TYPE_BARCHART,
        durationObj,
        resultState
      },
      formatData,
      formatDataBasedOnChart: formatDataInSeriesFormat,
      tableDataSelector,
      getCSVDataCallback,
      EventBreakDownType: 'sesb',
      formatDataParams: {},
      formatDataBasedOnChartParams: {}
    });
    return results;
  };

  const getCSVDataCallback = useCallback(
    (data) =>
      data.map(({ index, label, value, name, marker, ...rest }) => {
        const result = {};
        Object.keys(rest).forEach((key) => {
          if (key === 'data') {
            return;
          }
          if (key === EVENT_COUNT_KEY) {
            result[getEventDisplayName({ eventNames, event: events[0] })] =
              rest[EVENT_COUNT_KEY];
            return;
          }
          if (key === events[0]) {
            result[getEventDisplayName({ eventNames, event: events[0] })] =
              rest[events[0]];
            return;
          }
          result[key] = rest[key];
        });
        return result;
      }),

    [chartType, eventNames, events]
  );
  async function getCSVData() {
    try {
      let csvData = [];
      csvData = await fetchData();
      return {
        fileName: reportTitle,
        data: csvData
      };
    } catch (error) {
      console.error(error);
      return [];
    }
  }

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
        comparisonApplied,
        compareCategories,
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
    comparisonApplied,
    compareCategories,
    projectDomainsList
  ]);

  useEffect(() => {
    setDateBasedTableData(
      getDateBasedTableData(
        seriesData,
        searchText,
        dateSorter,
        categories,
        comparisonApplied,
        compareCategories,
        frequency
      )
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
      return null;
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
      return null;
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

export default React.memo(SingleEventSingleBreakdownTable);
