import React, { useState, useEffect, useCallback, useRef } from 'react';
import { Table } from 'antd';
import SearchBar from './SearchBar';
import styles from './index.module.scss';
import { useHistory } from 'react-router-dom';

function DataTable({
  tableData,
  searchText,
  setSearchText,
  columns,
  className,
  rowSelection,
  scroll,
  isWidgetModal,
  getCSVData,
  ignoreDocumentClick,
  renderSearch = true,
  isPaginationEnabled = true,
  defaultPageSize = 10,
}) {
  const componentRef = useRef(null);
  const downloadBtnRef = useRef(null);
  const [pageSize, setPageSize] = useState(defaultPageSize);

  const [searchBar, showSearchBar] = useState(false);

  const history = useHistory();

  let isDashboardWidget = !isWidgetModal;

  if (history.location.pathname === '/analyse') {
    isDashboardWidget = false;
  }

  const handleSearchTextChange = useCallback(
    (value) => {
      setSearchText(value);
    },
    [setSearchText]
  );

  const handleDocumentClick = useCallback(
    (e) => {
      if (ignoreDocumentClick) {
        return false;
      }
      if (componentRef && !componentRef.current.contains(e.target)) {
        showSearchBar(false);
        handleSearchTextChange('');
      } else {
        if (
          !searchBar &&
          downloadBtnRef &&
          downloadBtnRef.current?.contains(e.target)
        ) {
          document.getElementById('csvLink').click();
        }
        showSearchBar(true);
      }
    },
    [handleSearchTextChange, searchBar]
  );

  useEffect(() => {
    document.addEventListener('mousedown', handleDocumentClick);
    return () => {
      document.removeEventListener('mousedown', handleDocumentClick);
    };
  }, [handleDocumentClick]);

  const handlePageSizeChange = (...args) => {
    setPageSize(args[1]);
  };

  return (
    <div ref={componentRef} className='data-table'>
      {!isDashboardWidget && renderSearch ? (
        <SearchBar
          searchText={searchText}
          handleSearchTextChange={handleSearchTextChange}
          searchBar={searchBar}
          getCSVData={getCSVData}
          downloadBtnRef={downloadBtnRef}
        />
      ) : null}
      <Table
        pagination={
          !isDashboardWidget && isPaginationEnabled
            ? {
                pageSize,
                onShowSizeChange: handlePageSizeChange,
                showSizeChanger: tableData.length > 10,
              }
            : false
        }
        bordered={true}
        rowKey='index'
        rowSelection={!isDashboardWidget ? rowSelection : null}
        columns={columns}
        dataSource={isDashboardWidget ? tableData.slice(0, 3) : tableData}
        className={`${styles.table} ${className}`}
        scroll={scroll}
      />
    </div>
  );
}

export default DataTable;
