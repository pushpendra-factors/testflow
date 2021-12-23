import React, { useState, useCallback, useEffect } from 'react';
import moment from 'moment';
import {
  CHART_TYPE_SPARKLINES,
  CHART_TYPE_LINECHART,
} from '../../../../utils/constants';
import {
  getTableColumns,
  getTableData,
  getDateBaseTableColumns,
  getDateBasedTableData,
} from './utils';
import DataTable from '../../../../components/DataTable';
import { getNewSorterState } from '../../../../utils/dataFormatter';

function NoBreakdownTable({
  chartsData,
  chartType,
  isWidgetModal,
  frequency,
  reportTitle = 'CampaignAnalytics',
}) {
  const [columns, setColumns] = useState([]);
  const [tableData, setTableData] = useState([]);
  const [dateBasedColumns, setDateBasedColumns] = useState([]);
  const [dateBasedTableData, setDateBasedTableData] = useState([]);
  const [searchText, setSearchText] = useState('');
  const [sorter, setSorter] = useState([
    {
      key: chartsData[0].name,
      type: 'numerical',
      subtype: null,
      order: 'descend',
    },
  ]);
  const [dateSorter, setDateSorter] = useState([]);

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
    setColumns(getTableColumns(chartsData, frequency, sorter, handleSorting));
  }, [chartsData, frequency, sorter, handleSorting]);

  useEffect(() => {
    setTableData(getTableData(chartsData, sorter));
  }, [chartsData, sorter]);

  useEffect(() => {
    setDateBasedColumns(
      getDateBaseTableColumns(
        chartsData,
        frequency,
        dateSorter,
        handleDateSorting
      )
    );
  }, [chartsData, frequency, dateSorter, handleDateSorting]);

  useEffect(() => {
    setDateBasedTableData(
      getDateBasedTableData(chartsData, frequency, dateSorter)
    );
  }, [chartsData, frequency, dateSorter]);

  const getCSVData = () => {
    const result =
      chartType === CHART_TYPE_LINECHART ? dateBasedTableData : tableData;
    return {
      fileName: `${reportTitle}.csv`,
      data: result.map(({ index, date, ...rest }) => {
        if (chartType === CHART_TYPE_SPARKLINES) {
          let format = 'MMM D, YYYY';
          return {
            date: moment(date).format(format),
            ...rest,
          };
        }
        return rest;
      }),
    };
  };

  return (
    <DataTable
      isWidgetModal={isWidgetModal}
      tableData={
        chartType === CHART_TYPE_LINECHART ? dateBasedTableData : tableData
      }
      searchText={searchText}
      setSearchText={setSearchText}
      columns={chartType === CHART_TYPE_LINECHART ? dateBasedColumns : columns}
      scroll={{ x: 250 }}
      getCSVData={getCSVData}
      // rowSelection={rowSelection}
    />
  );
}

export default NoBreakdownTable;
