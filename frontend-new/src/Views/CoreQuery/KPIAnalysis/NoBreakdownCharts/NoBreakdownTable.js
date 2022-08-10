import React, { useState, useEffect, memo, useCallback } from 'react';
import moment from 'moment';
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
  DATE_FORMATS
} from '../../../../utils/constants';
import { addQforQuarter } from '../../../../utils/dataFormatter';
import { getFormattedKpiValue } from '../kpiAnalysis.helpers';

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
  frequency = 'date',
  comparisonApplied = false,
  compareCategories
}) => {
  const [searchText, setSearchText] = useState('');
  const [columns, setColumns] = useState([]);
  const [dateBasedColumns, setDateBasedColumns] = useState([]);
  const [tableData, setTableData] = useState([]);
  const [dateBasedTableData, setDateBasedTableData] = useState([]);

  useEffect(() => {
    setColumns(
      getTableColumns({
        kpis,
        currentSorter: sorter,
        handleSorting,
        frequency,
        comparisonApplied
      })
    );
  }, [kpis, sorter, handleSorting, frequency, comparisonApplied]);

  useEffect(() => {
    setTableData(
      getDataInTableFormat(
        seriesData,
        categories,
        kpis,
        sorter,
        comparisonApplied,
        compareCategories
      )
    );
  }, [
    seriesData,
    categories,
    kpis,
    sorter,
    comparisonApplied,
    compareCategories
  ]);

  useEffect(() => {
    setDateBasedColumns(
      getDateBasedColumns({
        kpis,
        categories,
        currentSorter: dateSorter,
        handleSorting: handleDateSorting,
        frequency,
        comparisonApplied,
        compareCategories
      })
    );
  }, [
    kpis,
    categories,
    dateSorter,
    handleDateSorting,
    frequency,
    compareCategories
  ]);

  useEffect(() => {
    setDateBasedTableData(
      getDateBasedTableData(
        seriesData,
        categories,
        searchText,
        dateSorter,
        frequency,
        comparisonApplied,
        compareCategories
      )
    );
  }, [
    seriesData,
    searchText,
    categories,
    dateSorter,
    frequency,
    comparisonApplied,
    compareCategories
  ]);

  const getCSVData = useCallback(() => {
    const activeTableData =
      chartType !== CHART_TYPE_SPARKLINES ? dateBasedTableData : tableData;
    const format = DATE_FORMATS[frequency] || DATE_FORMATS.date;
    return {
      fileName: 'KPI.csv',
      data: activeTableData.map(({ index, label, date, ...rest }) => {
        if (chartType === CHART_TYPE_SPARKLINES) {
          for (const key in rest) {
            const metricType = get(
              find(kpis, (kpi, index) => kpi.label + ' - ' + index === key),
              'metricType',
              null
            );
            if (metricType) {
              rest[key] = getFormattedKpiValue({
                value: rest[key],
                metricType
              });
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
        if (metricType) {
          for (const key in rest) {
            if (key !== 'event') {
              rest[key] = getFormattedKpiValue({
                value: rest[key],
                metricType
              });
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
