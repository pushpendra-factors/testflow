import React, {
  useState, useEffect, useCallback, useRef
} from 'react';
import { Table } from 'antd';
import { generateTableColumns, generateTableData } from '../utils';
import styles from './index.module.scss';
import SearchBar from './SearchBar';

function DataTable({ eventsData, groups, setGroups }) {
  const [tableData, setTableData] = useState([]);
  const [sorter, setSorter] = useState({});
  const [searchText, setSearchText] = useState('');
  const [searchBar, showSearchBar] = useState(false);
  const componentRef = useRef(null);

  const handleSearchTextChange = useCallback((value) => {
    setSearchText(value);
  }, [setSearchText]);

  const handleDocumentClick = useCallback((e) => {
    if (componentRef && !componentRef.current.contains(e.target)) {
      showSearchBar(false);
      handleSearchTextChange('');
    } else {
      showSearchBar(true);
    }
  }, [handleSearchTextChange]);

  useEffect(() => {
    document.addEventListener('mousedown', handleDocumentClick);
    return () => {
      document.removeEventListener('mousedown', handleDocumentClick);
    };
  }, [handleDocumentClick]);

  const handleSorting = useCallback((sorter) => {
    setSorter(sorter);
  }, []);

  useEffect(() => {
    setTableData(generateTableData(eventsData, groups, sorter, searchText));
  }, [eventsData, groups, sorter, searchText]);

  const onSelectionChange = (selectedRowKeys, selectedRows) => {
    const selectedGroups = selectedRows.map(elem => elem.name);
    setGroups(currentData => {
      return currentData.map(elem => {
        return { ...elem, is_visible: selectedGroups.indexOf(elem.name) >= 0 };
      });
    });
  };

  const selectedRowKeys = groups
    .filter(elem => elem.is_visible)
    .map(elem => elem.name)
    .map(elem => {
      return tableData.findIndex(d => d.name === elem);
    });

  const rowSelection = {
    selectedRowKeys,
    onChange: onSelectionChange
  };

  const columns = generateTableColumns(eventsData, sorter, handleSorting);

  return (
    <div ref={componentRef} className="data-table">
      <SearchBar
        searchText={searchText}
        handleSearchTextChange={handleSearchTextChange}
        searchBar={searchBar}
      />
      <Table
        pagination={false}
        bordered={true}
        rowKey="index"
        rowSelection={rowSelection}
        columns={columns}
        dataSource={tableData}
        className={styles.table}
      />
    </div>
  );
}

export default DataTable;
