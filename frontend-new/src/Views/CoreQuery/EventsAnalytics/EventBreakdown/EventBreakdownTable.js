import React, { useState, useCallback, useContext } from 'react';
import { formatData, getTableColumns, getTableData } from './utils';
import DataTable from '../../../../components/DataTable';
import { useSelector } from 'react-redux';
import { getNewSorterState } from '../../../../utils/dataFormatter';
import { MAX_ALLOWED_VISIBLE_PROPERTIES } from '../../../../utils/constants';
import { CoreQueryContext } from 'Context/CoreQueryContext';
import {
  fetchDataCSVDownload,
  getEventsCSVData,
  getQuery
} from 'Views/CoreQuery/utils';

function EventBreakdownTable({
  breakdown,
  data,
  visibleProperties,
  setVisibleProperties,
  reportTitle = 'Events Analytics',
  sorter,
  setSorter,
  durationObj,
  resultState
}) {
  const {
    userPropNames,
    eventPropertiesDisplayNames: eventPropertiesDisplayNamesState
  } = useSelector((state) => state.coreQuery);
  const { data: eventPropertiesDisplayNames } =
    eventPropertiesDisplayNamesState;
  const coreQueryContext = useContext(CoreQueryContext);
  const { show_criteria: result_criteria, performance_criteria: user_type } =
    useSelector((state) => state.analyticsQuery);
  const { active_project } = useSelector((state) => state.global);

  const [searchText, setSearchText] = useState('');

  const handleSorting = useCallback(
    (prop) => {
      setSorter((currentSorter) => {
        return getNewSorterState(currentSorter, prop);
      });
    },
    [setSorter]
  );

  const columns = getTableColumns(
    breakdown,
    sorter,
    handleSorting,
    userPropNames,
    eventPropertiesDisplayNames
  );
  const tableDataSelector = (data) =>
    getTableData(data, breakdown, searchText, sorter);
  const getCSVDataCallback = (d) =>
    d.map(({ index, ...rest }) => ({ ...rest }));
  const fetchData = async () => {
    const results = await fetchDataCSVDownload(coreQueryContext, {
      projectID: active_project?.id,
      step2Properties: {
        result_criteria,
        user_type,
        shouldStack: true,
        durationObj,
        resultState
      },
      formatData,
      formatDataBasedOnChart: (a) => a,
      tableDataSelector,
      getCSVDataCallback,
      EventBreakDownType: 'eb',
      formatDataParams: {},
      formatDataBasedOnChartParams: {}
    });
    return results;
  };
  const tableData = getTableData(data, breakdown, searchText, sorter);

  const getCSVData = async () => ({
    fileName: reportTitle,
    data: await fetchData()
  });

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

  const selectedRowKeys = visibleProperties.map((elem) => elem.index);

  const rowSelection = {
    selectedRowKeys,
    onChange: onSelectionChange
  };

  return (
    <DataTable
      tableData={tableData}
      searchText={searchText}
      setSearchText={setSearchText}
      columns={columns}
      rowSelection={rowSelection}
      getCSVData={getCSVData}
    />
  );
}

export default EventBreakdownTable;
