import React, {
  useState, useEffect, useCallback, useRef
} from 'react';
import { Table } from 'antd';
import SearchBar from './SearchBar';
import styles from './index.module.scss';
import { useHistory } from 'react-router-dom';

function DataTable({
  tableData, searchText, setSearchText, columns, className, rowSelection, scroll
}) {
  const componentRef = useRef(null);

  const [searchBar, showSearchBar] = useState(false);

  const history = useHistory();

  let isDashboardView = true;

  if (history.location.pathname === '/core-analytics') {
    isDashboardView = false;
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

  return (
    <div ref={componentRef} className="data-table">
      {!isDashboardView && (
        <SearchBar
          searchText={searchText}
          handleSearchTextChange={handleSearchTextChange}
          searchBar={searchBar}
        />
      )}
      <Table
        pagination={{ pageSize: isDashboardView ? 5 : 10 }}
        bordered={true}
        rowKey="index"
        rowSelection={!isDashboardView ? rowSelection : null}
        columns={columns}
        dataSource={tableData}
        className={`${styles.table} ${className}`}
        scroll={scroll}
      />
    </div>
  );
}

export default DataTable;
