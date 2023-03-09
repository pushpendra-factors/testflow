import { ConfigProvider } from 'antd';
import DataTable from 'Components/DataTable';
import useAgentInfo from 'hooks/useAgentInfo';
import React, { useEffect, useState, useCallback } from 'react';
import { useHistory } from 'react-router-dom';
import { KEY_LABELS, PAGE_COUNT_KEY } from '../../../const';
import { ReportTableProps, StringObject } from '../../../types';
import {
  getDefaultTableColumns,
  getTableColumuns,
  getTableData
} from '../../../utils';
import EmptyDataState from './EmptyDataState';

const ReportTable = ({
  data,
  selectedChannel,
  selectedCampaigns,
  isSixSignalActivated
}: ReportTableProps) => {
  const [searchText, setSearchText] = useState<string>('');
  const [tableData, setTableData] = useState<StringObject[]>([]);
  const [columns, setColumns] = useState(getDefaultTableColumns());
  // const [visibleHeaders,setVisibleHeaders]= useState(data.headers);
  const [sorter, setSorter] = useState([
    {
      key: PAGE_COUNT_KEY,
      type: 'numerical',
      subtype: null,
      order: 'descend'
    }
  ]);
  const { isLoggedIn } = useAgentInfo();
  const history = useHistory();
  const handleSorting = useCallback((prop) => {
    setSorter((currentSorter) => {
      if (currentSorter[0].key === prop.key) {
        return [
          {
            ...currentSorter[0],
            order: currentSorter[0].order === 'ascend' ? 'descend' : 'ascend'
          }
        ];
      }
      return [
        {
          ...prop,
          order: 'ascend'
        }
      ];
    });
  }, []);
  const getCSVData = () => {
    const columnsArray = columns?.map(
      (column) => KEY_LABELS?.[column?.dataIndex] || column
    );
    const rowsArray = tableData.map((row, i) => {
      return columns.map((column) => {
        return row[column.dataIndex] || '';
      });
    });
    return {
      fileName: `SixSignalReport.csv`,
      data: [columnsArray, ...rowsArray]
    };
  };

  useEffect(() => {
    if (data && data?.headers && data?.rows) {
      let dataSource: StringObject[] = getTableData(
        data,
        searchText,
        selectedChannel,
        selectedCampaigns,
        sorter
      );
      if (dataSource) setTableData(dataSource);
      const tColumns = getTableColumuns(data, sorter, handleSorting);

      if (tColumns) setColumns(tColumns);
    }
  }, [
    data,
    selectedChannel,
    searchText,
    handleSorting,
    sorter,
    selectedCampaigns
  ]);

  const NoDataState = (
    <EmptyDataState
      title='New deals are just around the corner'
      subtitle="Looks like there isn't much here yet"
      icon={{ name: 'EmptyDataBox', height: 111, width: 131, color: '' }}
    />
  );

  const NoIntegrationState = (
    <EmptyDataState
      title='Get started by integrating with 6signal'
      subtitle='Use your own API key, or use ours to get going immediately '
      icon={{ name: 'UserSearch', height: 57, width: 71, color: '' }}
      action={{
        name: 'Setup 6signal',
        handleClick: () => {
          history.push('/settings/integration');
        }
      }}
    />
  );

  return (
    <div>
      <ConfigProvider
        renderEmpty={() => {
          if (isLoggedIn) {
            return isSixSignalActivated ? NoDataState : NoIntegrationState;
          }
          return NoDataState;
        }}
      >
        {/* @ts-ignore */}
        <DataTable
          tableData={tableData}
          searchText={searchText}
          setSearchText={setSearchText}
          columns={columns}
          getCSVData={getCSVData}
          renderSearch
          isPaginationEnabled
          isWidgetModal={true}
          breakupHeading={'Top accounts'}
        />
      </ConfigProvider>
    </div>
  );
};

export default ReportTable;
