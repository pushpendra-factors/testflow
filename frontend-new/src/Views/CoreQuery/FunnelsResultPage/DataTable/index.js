import React, {
  useState, useEffect, useCallback, useRef
} from 'react';
import { Table } from 'antd';
import SearchBar from './SearchBar';
import styles from './index.module.scss';

function DataTable({
  tableData, searchText, setSearchText, columns, className, rowSelection
}) {
  const componentRef = useRef(null);

  const [searchBar, showSearchBar] = useState(false);

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
                className={`${styles.table} ${className}`}
            />
        </div>
  );
}

export default DataTable;
