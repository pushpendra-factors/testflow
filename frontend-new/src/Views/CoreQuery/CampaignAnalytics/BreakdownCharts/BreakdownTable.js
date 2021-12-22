import React, { useState, useCallback, useMemo, useEffect } from 'react';
import {
  getTableColumns,
  getTableData,
  getDateBasedColumns,
  getDateBasedTableData,
  getDefaultSorterState,
} from './utils';
import DataTable from '../../../../components/DataTable';
import {
  CHART_TYPE_BARCHART,
  MAX_ALLOWED_VISIBLE_PROPERTIES,
  DASHBOARD_WIDGET_SECTION,
} from '../../../../utils/constants';
import { getNewSorterState } from '../../../../utils/dataFormatter';

function BreakdownTable({
  chartsData,
  seriesData,
  categories,
  breakdown,
  currentEventIndex,
  chartType,
  arrayMapper,
  isWidgetModal,
  visibleProperties,
  setVisibleProperties,
  section,
  reportTitle = 'CampaignAnalytics',
}) {
  const [sorter, setSorter] = useState(
    getDefaultSorterState(arrayMapper, currentEventIndex)
  );
  const [columns, setColumns] = useState([]);
  const [tableData, setTableData] = useState([]);
  const [dateBasedColumns, setDateBasedColumns] = useState([]);
  const [dateBasedTableData, setDateBasedTableData] = useState([]);
  const [dateSorter, setDateSorter] = useState([]);
  const [searchText, setSearchText] = useState('');

  const handleSorting = useCallback((prop) => {
    setSorter((currentSorter) => {
      if (currentSorter[0].key === prop.key) {
        return [
          {
            ...currentSorter[0],
            order: currentSorter[0].order === 'ascend' ? 'descend' : 'ascend',
          },
        ];
      }
      return [
        {
          ...prop,
          order: 'ascend',
        },
      ];
    });
  }, []);

  const handleDateSorting = useCallback((prop) => {
    setDateSorter((currentSorter) => {
      if (currentSorter[0].key === prop.key) {
        return [
          {
            ...currentSorter[0],
            order: currentSorter[0].order === 'ascend' ? 'descend' : 'ascend',
          },
        ];
      }
      return [
        {
          ...prop,
          order: 'ascend',
        },
      ];
    });
  }, []);

  useEffect(() => {
    setColumns(getTableColumns(arrayMapper, breakdown, sorter, handleSorting));
  }, [arrayMapper, breakdown, sorter, handleSorting]);

  useEffect(() => {
    setTableData(getTableData(chartsData, breakdown, searchText, sorter));
  }, [
    chartsData,
    breakdown,
    arrayMapper,
    currentEventIndex,
    searchText,
    sorter,
  ]);

  useEffect(() => {
    setDateBasedColumns(
      getDateBasedColumns(categories, breakdown, dateSorter, handleDateSorting)
    );
  }, [categories, breakdown, dateSorter, handleDateSorting]);

  useEffect(() => {
    setDateBasedTableData(
      getDateBasedTableData(
        seriesData,
        categories,
        breakdown,
        searchText,
        dateSorter,
        arrayMapper,
        currentEventIndex
      )
    );
  }, [
    seriesData,
    categories,
    breakdown,
    searchText,
    dateSorter,
    arrayMapper,
    currentEventIndex,
  ]);

  const getCSVData = () => {
    const activeTableData =
      chartType === CHART_TYPE_BARCHART ? tableData : dateBasedTableData;
    return {
      fileName: `${reportTitle}.csv`,
      data: activeTableData.map(({ index, ...rest }) => {
        return rest;
      }),
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
        const obj = chartsData.find((d) => d.index === elem.index);
        return obj;
      });
      setVisibleProperties(newVisibleProperties);
    },
    [setVisibleProperties, chartsData]
  );

  const rowSelection = {
    selectedRowKeys,
    onChange: onSelectionChange,
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
      scroll={{ x: 250 }}
      rowSelection={rowSelection}
      getCSVData={getCSVData}
    />
  );
}

export default BreakdownTable;
