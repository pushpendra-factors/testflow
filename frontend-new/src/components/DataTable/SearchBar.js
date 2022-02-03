import React, { useCallback } from 'react';
import styles from './index.module.scss';
import { Input, Button, Popover } from 'antd';
import { SVG } from '../factorsComponents';
import { CSVLink } from 'react-csv';

function SearchBar({
  searchText,
  handleSearchTextChange,
  searchBar,
  getCSVData,
  toggleSearchBar,
  controlsPopover
}) {
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
  }

  const handleSearchInputChange = ({ target: { value } }) => {
    handleSearchTextChange(value)
  }

  const downloadBtn = (
    <Button
      size={'large'}
      icon={<SVG name={'download'} size={20} color={'grey'} />}
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
      ></CSVLink>
    </Button>
  );

  const searchBtn = (
    <Button
      size={'large'}
      onClick={toggleSearchBar}
      icon={<SVG name={'search'} size={20} color={'grey'} />}
      type='text'
    />
  );

  const closeBtn = (
    <Button
      size={'large'}
      onClick={handleSearchBarClose}
      icon={<SVG name={'close'} size={20} color={'grey'} />}
      type='text'
    />
  );

  const controlsBtn = (
    <Popover
      placement='bottomLeft'
      trigger='click'
      content={controlsPopover}
    >
      <Button
        size={'large'}
        icon={<SVG name={'controls'} />}
        type='text'
      />
    </Popover>
  );

  return (
    <div className={`flex items-center px-4 ${styles.searchBar}`}>
      <div className='flex justify-between w-full'>
        {!searchBar ? (
          <div className={'flex items-center cursor-pointer'}>
            <div className={styles.breakupHeading}>Break-up</div>
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
              prefix={<SVG name={'search'} size={20} color={'grey'} />}
            />
          )}
        <div className='flex items-center'>
          {searchBar ? (
            <div className='flex items-center'>{closeBtn}</div>
          ) : (
              <div className='flex items-center'>{searchBtn}</div>
            )}
          {!!controlsPopover && <div className="flex items-center">{controlsBtn}</div>}
          <div className='flex items-center'>{downloadBtn}</div>
        </div>
      </div>
    </div>
  );
}

export default SearchBar;
