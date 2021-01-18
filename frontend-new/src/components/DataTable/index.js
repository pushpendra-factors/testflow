import React, {
  useState, useEffect, useCallback, useRef
} from 'react';
import { Table } from 'antd';
import SearchBar from './SearchBar';
import styles from './index.module.scss';
import { useHistory } from 'react-router-dom';

function DataTable({
  tableData, searchText, setSearchText, columns, className, rowSelection, scroll, isWidgetModal
}) {
  const componentRef = useRef(null);
  const [pageSize, setPageSize] = useState(10);

  const [searchBar, showSearchBar] = useState(false);

  const history = useHistory();

  let isDashboardWidget = !isWidgetModal;

  if (history.location.pathname === '/analyse') {
    isDashboardWidget = false;
  }

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

  const handlePageSizeChange = (...args) => {
    setPageSize(args[1]);
  }

  return (
    <div ref={componentRef} className="data-table">
      {!isDashboardWidget ? (
        <SearchBar
          searchText={searchText}
          handleSearchTextChange={handleSearchTextChange}
          searchBar={searchBar}
        />
      ) : null}
      <Table
        pagination={!isDashboardWidget ? { pageSize, onShowSizeChange: handlePageSizeChange } : false}
        bordered={true}
        rowKey="index"
        rowSelection={!isDashboardWidget ? rowSelection : null}
        columns={columns}
        dataSource={isDashboardWidget ? tableData.slice(0, 5) : tableData}
        className={`${styles.table} ${className}`}
        scroll={scroll}
      />
    </div>
  );
}

export default DataTable;
