import React, {
  useState, useEffect, memo, useCallback
} from 'react';
import moment from 'moment';
import { useSelector } from 'react-redux';
import { find, get } from 'lodash';
import {
  getTableColumns,
  getDataInTableFormat,
  getDateBasedColumns,
  getDateBasedTableData
} from './utils';
import DataTable from '../../../../components/DataTable';
import {
  CHART_TYPE_SPARKLINES,
  DASHBOARD_WIDGET_SECTION,
  DATE_FORMATS,
  METRIC_TYPES
} from '../../../../utils/constants';
import {
  addQforQuarter,
  formatDuration
} from '../../../../utils/dataFormatter';

const NoBreakdownTable = ({
  seriesData,
  categories,
  kpis,
  section,
  chartType,
  sorter,
  handleSorting,
  dateSorter,
  handleDateSorting,
  frequency = 'date'
}) => {
  const { eventNames } = useSelector((state) => state.coreQuery);
  const [searchText, setSearchText] = useState('');
  const [columns, setColumns] = useState([]);
  const [dateBasedColumns, setDateBasedColumns] = useState([]);
  const [tableData, setTableData] = useState([]);
  const [dateBasedTableData, setDateBasedTableData] = useState([]);

  useEffect(() => {
    setColumns(
      getTableColumns(kpis, sorter, handleSorting, eventNames, frequency)
    );
  }, [kpis, sorter, handleSorting, eventNames]);

  useEffect(() => {
    setTableData(getDataInTableFormat(seriesData, categories, kpis, sorter));
  }, [seriesData, categories, kpis, sorter]);

  useEffect(() => {
    setDateBasedColumns(
      getDateBasedColumns(
        kpis,
        categories,
        dateSorter,
        handleDateSorting,
        eventNames,
        frequency
      )
    );
  }, [kpis, categories, dateSorter, handleDateSorting, eventNames, frequency]);

  useEffect(() => {
    setDateBasedTableData(
      getDateBasedTableData(
        seriesData,
        categories,
        searchText,
        dateSorter,
        frequency
      )
    );
  }, [seriesData, searchText, dateSorter]);

  const getCSVData = useCallback(() => {
    const activeTableData =
      chartType !== CHART_TYPE_SPARKLINES ? dateBasedTableData : tableData;
    const format = DATE_FORMATS[frequency] || DATE_FORMATS.date;
    return {
      fileName: 'KPI.csv',
      data: activeTableData.map(({
        index, label, date, ...rest
      }) => {
        if (chartType === CHART_TYPE_SPARKLINES) {
          for (const key in rest) {
            const metricType = get(
              find(kpis, (kpi, index) => kpi.label + ' - ' + index === key),
              'metricType',
              null
            );
            if (metricType === METRIC_TYPES.dateType) {
              rest[key] = formatDuration(rest[key]);
            }
          }
          return {
            ...rest,
            date: addQforQuarter(frequency) + moment(date).format(format)
          };
        }
        const metricType = get(
          find(kpis, (kpi) => kpi.label === rest.event),
          'metricType',
          null
        );
        if (metricType === METRIC_TYPES.dateType) {
          for (const key in rest) {
            if (key !== 'event') {
              rest[key] = formatDuration(rest[key]);
            }
          }
        }
        return { ...rest };
      })
    };
  }, [chartType, frequency, dateBasedTableData, tableData, kpis]);

  return (
    <DataTable
      tableData={
        chartType === CHART_TYPE_SPARKLINES ||
        section === DASHBOARD_WIDGET_SECTION
          ? tableData
          : dateBasedTableData
      }
      searchText={searchText}
      setSearchText={setSearchText}
      columns={
        chartType === CHART_TYPE_SPARKLINES ||
        section === DASHBOARD_WIDGET_SECTION
          ? columns
          : dateBasedColumns
      }
      rowSelection={null}
      scroll={{ x: 250 }}
      getCSVData={getCSVData}
    />
  );
};

export default memo(NoBreakdownTable);
