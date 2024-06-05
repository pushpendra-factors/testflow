import React, {
  useState,
  useCallback,
  useEffect,
  memo,
  useContext
} from 'react';
import find from 'lodash/find';
import {
  getTableColumns,
  getTableData,
  getDateBasedColumns,
  getDateBasedTableData,
  formatData,
  formatDataInStackedAreaFormat
} from './utils';
import DataTable from '../../../../components/DataTable';
import {
  CHART_TYPE_BARCHART,
  MAX_ALLOWED_VISIBLE_PROPERTIES,
  DASHBOARD_WIDGET_SECTION
} from '../../../../utils/constants';
import { useSelector } from 'react-redux';
import { isSeriesChart } from '../../../../utils/dataFormatter';
import { getBreakdownDisplayName } from '../eventsAnalytics.helpers';
import { CoreQueryContext } from 'Context/CoreQueryContext';
import {
  fetchDataCSVDownload,
  getEventsCSVData,
  getQuery
} from 'Views/CoreQuery/utils';

function MultipleEventsWithBreakdownTable({
  chartType,
  breakdown,
  data,
  seriesData,
  categories,
  visibleProperties,
  setVisibleProperties,
  isWidgetModal,
  durationObj,
  reportTitle = 'Events Analytics',
  section,
  sorter,
  handleSorting,
  dateSorter,
  handleDateSorting,
  visibleSeriesData,
  setVisibleSeriesData,
  eventGroup,
  resultState,
  queries
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

  const tableDataSelector = (data) =>
    isSeriesChart(chartType)
      ? getDateBasedTableData(data, dateSorter, searchText)
      : getTableData(data, searchText, sorter);
  const getCSVDataCallback = (d) =>
    d.map(
      ({
        index,
        eventIndex,
        dateWise,
        color,
        label,
        value,
        name,
        data,
        marker,
        ...rest
      }) => {
        const result = {};
        for (const key in rest) {
          if (key.toLowerCase() === 'event') {
            result.Event = rest[key];
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
          result[key.split(';')[0]] = rest[key];
        }
        return result;
      }
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
      EventBreakDownType: 'meb',
      formatDataParams: { eventNames },
      formatDataBasedOnChartParams: { eventNames, queries }
    });
    return results;
    // const q = x?.resultState?.data?.meta?.query;

    // const q1 = [q];
    // if (q) {
    //   q1.push({ ...q, gbt: '' });
    // }

    // const query = q1;
    // const result = await getEventsCSVData(
    //   active_project?.id,
    //   query,
    //   {
    //     result_criteria,
    //     user_type,
    //     shouldStack: !isSeriesChart(chartType)
    //   },
    //   formatData,
    //   formatDataInStackedAreaFormat,
    //   tableDataSelector,
    //   getCSVDataCallback,
    //   x.coreQueryState,
    //   { durationObj, resultState },
    //   'meb',
    //   { eventNames },
    //   { eventNames, queries }
    // );

    // return result;
  };
  useEffect(() => {
    setColumns(
      getTableColumns(
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
    setTableData(getTableData(data, searchText, sorter));
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
    handleDateSorting,
    durationObj.frequency,
    userPropNames,
    eventPropertiesDisplayNames,
    projectDomainsList
  ]);

  useEffect(() => {
    setDateBasedTableData(
      getDateBasedTableData(seriesData, dateSorter, searchText)
    );
  }, [dateSorter, searchText, seriesData]);

  const getCSVData = async () => ({
    fileName: reportTitle,
    data: await fetchData()
  });
  // tableData,
  // dateBasedTableData,
  // reportTitle,
  // eventNames,
  // chartType,
  // breakdown,
  // userPropNames,
  // eventPropertiesDisplayNames

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

  const selectedRowKeys = useCallback((rows) => {
    return rows.map((vp) => vp.index);
  }, []);

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

export default memo(MultipleEventsWithBreakdownTable);
