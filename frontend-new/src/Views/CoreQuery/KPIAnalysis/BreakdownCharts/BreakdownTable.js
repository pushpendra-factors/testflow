import React, { useState, useCallback, useEffect, memo } from 'react';
import { useSelector } from 'react-redux';
import { find } from 'lodash';
import {
  getTableColumns,
  getDataInTableFormat,
  getDateBasedColumns,
  getDateBasedTableData
} from './utils';
import DataTable from '../../../../components/DataTable';
import {
  MAX_ALLOWED_VISIBLE_PROPERTIES,
  CHART_TYPE_HORIZONTAL_BAR_CHART,
  DASHBOARD_WIDGET_SECTION,
  QUERY_TYPE_KPI
} from '../../../../utils/constants';
import { isSeriesChart } from '../../../../utils/dataFormatter';
import { getFormattedKpiValue } from '../kpiAnalysis.helpers';
import { getBreakdownDisplayName } from '../../EventsAnalytics/eventsAnalytics.helpers';

const BreakdownTable = ({
  data,
  kpis,
  seriesData,
  categories,
  section,
  breakdown,
  chartType,
  setVisibleProperties,
  visibleProperties,
  sorter,
  handleSorting,
  dateSorter,
  handleDateSorting,
  visibleSeriesData,
  setVisibleSeriesData,
  frequency = 'date',
  comparisonApplied,
  compareCategories
}) => {
  const [searchText, setSearchText] = useState('');
  const [columns, setColumns] = useState([]);
  const [dateBasedColumns, setDateBasedColumns] = useState([]);
  const [tableData, setTableData] = useState([]);
  const [dateBasedTableData, setDateBasedTableData] = useState([]);
  const { projectDomainsList } = useSelector((state) => state.global);

  useEffect(() => {
    setColumns(
      getTableColumns(
        breakdown,
        kpis,
        sorter,
        handleSorting,
        comparisonApplied,
        projectDomainsList
      )
    );
  }, [
    breakdown,
    sorter,
    handleSorting,
    kpis,
    comparisonApplied,
    projectDomainsList
  ]);

  useEffect(() => {
    setTableData(getDataInTableFormat(data, searchText, sorter));
  }, [data, searchText, sorter]);

  useEffect(() => {
    setDateBasedColumns(
      getDateBasedColumns(
        categories,
        breakdown,
        kpis,
        dateSorter,
        handleDateSorting,
        frequency,
        comparisonApplied,
        compareCategories,
        projectDomainsList
      )
    );
  }, [
    categories,
    breakdown,
    kpis,
    dateSorter,
    handleDateSorting,
    frequency,
    comparisonApplied,
    compareCategories,
    projectDomainsList
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
    if (!isSeriesChart(chartType)) {
      return {
        fileName: 'KPI.csv',
        data: tableData.map(({ index, label, value, metricType, ...rest }) => {
          const result = {};
          for (const key in rest) {
            const isCurrentKeyKpi = find(
              kpis,
              (kpi, index) => kpi.label + ' - ' + index === key
            );
            if (isCurrentKeyKpi && isCurrentKeyKpi.metricType) {
              result[key] = getFormattedKpiValue({
                value: rest[key],
                metricType: isCurrentKeyKpi.metricType
              });
              continue;
            }
            const isCurrentKeyForBreakdown = find(
              breakdown,
              (b, index) => b.property + ' - ' + index === key
            );
            if (isCurrentKeyForBreakdown) {
              result[
                `${getBreakdownDisplayName({
                  breakdown: isCurrentKeyForBreakdown,
                  queryType: QUERY_TYPE_KPI
                })} - ${key.split(' - ')[1]}`
              ] = rest[key];
              continue;
            }
            result[key] = rest[key];
          }
          return result;
        })
      };
    }
    return {
      fileName: 'KPI.csv',
      data: dateBasedTableData.map(
        ({
          index,
          label,
          value,
          metricType,
          data,
          marker,
          name,
          event,
          Overall,
          ...rest
        }) => {
          const result = {};
          const keys = Object.keys(rest);
          for (let i = 0; i < keys.length; i++) {
            const key = keys[i];
            const isCurrentKeyKpi = find(
              kpis,
              (kpi, kpiIdx) => `${kpi.label} - ${kpiIdx}` === key
            );
            if (isCurrentKeyKpi && isCurrentKeyKpi.metricType) {
              result[key] = getFormattedKpiValue({
                value: rest[key],
                metricType: isCurrentKeyKpi.metricType
              });
            } else {
              const isCurrentKeyForBreakdown = find(
                breakdown,
                (b, bIdx) => `${b.property} - ${bIdx}` === key
              );
              if (isCurrentKeyForBreakdown) {
                result[
                  `${getBreakdownDisplayName({
                    breakdown: isCurrentKeyForBreakdown,
                    queryType: QUERY_TYPE_KPI
                  })} - ${key.split(' - ')[1]}`
                ] = rest[key];
              } else {
                result[key] = rest[key];
              }
            }
          }
          return result;
        }
      )
    };
  }, [dateBasedTableData, chartType, tableData]);

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
      return false;
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
      return false;
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
      rowClassName={(record, index) => `multi-colored-checkbox-${index}`}
    />
  );
};

export default memo(BreakdownTable);
