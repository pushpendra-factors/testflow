import React from 'react';
import { Input, Button, Popover, Tooltip } from 'antd';
import { CSVLink } from 'react-csv';
import { SVG } from '../factorsComponents';
import DataTableFilters from '../DataTableFilters/DataTableFilters';
import ControlledComponent from '../ControlledComponent/ControlledComponent';
import styles from './index.module.scss';
import { TOOLTIP_CONSTANTS } from '../../constants/tooltips.constans';
import useAutoFocus from 'hooks/useAutoFocus';

function SearchBar({
  searchText,
  handleSearchTextChange,
  searchBar,
  getCSVData,
  toggleSearchBar,
  controlsPopover,
  filtersVisible,
  setFiltersVisibility,
  filters,
  appliedFilters,
  setAppliedFilters,
  breakupHeading = 'Break-up'
}) {
  const inputComponentRef = useAutoFocus(searchBar);
  let csvData = { data: [], fileName: 'data' };

  if (getCSVData) {
    csvData = getCSVData();
  }

  const handleDownloadBtnClick = () => {
    document.getElementById('csvLink').click();
  };

  const handleSearchBarClose = () => {
    toggleSearchBar();
    handleSearchTextChange('');
  };

  const handleSearchInputChange = ({ target: { value } }) => {
    handleSearchTextChange(value);
  };

  const downloadBtn = (
    <Tooltip title='Export to CSV' color={TOOLTIP_CONSTANTS.DARK}>
      <Button
        size='large'
        icon={<SVG name='download' size={20} color='grey' />}
        type='text'
        onClick={handleDownloadBtnClick}
      >
        <CSVLink
          id='csvLink'
          style={{ color: '#0E2647' }}
          onClick={() => {
            if (!csvData.data.length) return false;
          }}
          filename={csvData.fileName}
          data={csvData.data}
        />
      </Button>
    </Tooltip>
  );

  const searchBtn = (
    <Tooltip title='Search' color={TOOLTIP_CONSTANTS.DARK}>
      <Button
        size='large'
        onClick={toggleSearchBar}
        icon={<SVG name='search' size={20} color='grey' />}
        type='text'
      />
    </Tooltip>
  );

  const closeBtn = (
    <Tooltip title='Close' color={TOOLTIP_CONSTANTS.DARK}>
      <Button
        size='large'
        onClick={handleSearchBarClose}
        icon={<SVG name='close' size={20} color='grey' />}
        type='text'
      />
    </Tooltip>
  );

  const controlsBtn = (
    <Popover placement='bottomLeft' trigger='click' content={controlsPopover}>
      <Tooltip title='Edit Table Headers'>
        <Button size='large' icon={<SVG name='controls' />} type='text' />
      </Tooltip>
    </Popover>
  );

  const filtersContent = () => (
    <DataTableFilters
      key={filtersVisible}
      filters={filters}
      appliedFilters={appliedFilters}
      setAppliedFilters={setAppliedFilters}
      setFiltersVisibility={setFiltersVisibility}
    />
  );

  const filtersBtn = (
    <Popover
      onVisibleChange={setFiltersVisibility}
      overlayClassName={styles['filter-overlay']}
      placement='bottomRight'
      trigger='click'
      content={filtersContent}
      visible={filtersVisible}
    >
      <Button
        onClick={setFiltersVisibility?.bind(null, true)}
        size='large'
        icon={<SVG name='filter' />}
        type='text'
      />
    </Popover>
  );

  return (
    <div className={`flex items-center px-4 ${styles.searchBar}`}>
      <div className='flex justify-between w-full'>
        {!searchBar ? (
          <div className='flex items-center cursor-pointer'>
            <div className={styles.breakupHeading}>{breakupHeading}</div>
          </div>
        ) : (
          <Input
            onChange={handleSearchInputChange}
            value={searchText}
            className={`${styles.inputSearchBar} ${
              !searchText.length
                ? styles.inputPlaceHolderFont
                : styles.inputTextFont
            }`}
            size='large'
            placeholder='Search'
            prefix={<SVG name='search' size={20} color='grey' />}
            ref={inputComponentRef}
          />
        )}
        <div className='flex items-center'>
          {searchBar ? (
            <div className='flex items-center'>{closeBtn}</div>
          ) : (
            <div className='flex items-center'>{searchBtn}</div>
          )}

          <ControlledComponent controller={!!filters}>
            <div className='flex items-center'>{filtersBtn}</div>
          </ControlledComponent>

          <ControlledComponent controller={!!controlsPopover}>
            <div className='flex items-center'>{controlsBtn}</div>
          </ControlledComponent>

          <div className='flex items-center'>{downloadBtn}</div>
        </div>
      </div>
    </div>
  );
}

export default SearchBar;
