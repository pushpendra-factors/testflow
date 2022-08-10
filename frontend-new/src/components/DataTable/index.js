import React, { useState, useEffect, useCallback, useRef } from 'react';
import { Table } from 'antd';
import cx from 'classnames';
import SearchBar from './SearchBar';
import styles from './index.module.scss';
import { useHistory } from 'react-router-dom';
import useToggle from '../../hooks/useToggle';
import ControlledComponent from '../ControlledComponent/ControlledComponent';

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
  controlsPopover,
  filtersVisible,
  setFiltersVisibility,
  filters,
  appliedFilters,
  setAppliedFilters
}) {
  const componentRef = useRef(null);
  const downloadBtnRef = useRef(null);
  const [pageSize, setPageSize] = useState(defaultPageSize);
  const [searchBar, toggleSearchBar] = useToggle(false);
  const history = useHistory();
  let isDashboardWidget = !isWidgetModal;
  if (history.location.pathname.includes('/analyse')) {
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
      if (
        componentRef &&
        searchBar &&
        !componentRef.current.contains(e.target)
      ) {
        toggleSearchBar();
        handleSearchTextChange('');
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
    <div ref={componentRef} className="data-table">
      <ControlledComponent controller={!isDashboardWidget && renderSearch}>
        <SearchBar
          searchText={searchText}
          handleSearchTextChange={handleSearchTextChange}
          searchBar={searchBar}
          getCSVData={getCSVData}
          toggleSearchBar={toggleSearchBar}
          controlsPopover={controlsPopover}
          filters={filters}
          appliedFilters={appliedFilters}
          setAppliedFilters={setAppliedFilters}
          filtersVisible={filtersVisible}
          setFiltersVisibility={setFiltersVisibility}
        />
      </ControlledComponent>
      <Table
        pagination={
          !isDashboardWidget && isPaginationEnabled
            ? {
                pageSize,
                onShowSizeChange: handlePageSizeChange,
                showSizeChanger: tableData.length > 10
              }
            : false
        }
        bordered={true}
        rowKey="index"
        rowSelection={!isDashboardWidget ? rowSelection : null}
        columns={columns}
        dataSource={isDashboardWidget ? tableData.slice(0, 3) : tableData}
        className={cx(styles.table, className, {
          [styles.dashboardTable]: isDashboardWidget
        })}
        scroll={scroll}
      />
    </div>
  );
}
export default DataTable;
