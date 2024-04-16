import React from 'react';
import { Input } from 'antd';
import { SVG } from 'Components/factorsComponents';
import styles from './index.module.scss';

const SidebarSearch = ({
  searchText,
  setSearchText,
  onFocusSearch,
  placeholder
}) => {
  const handleSearchTextChange = (e) => {
    setSearchText(e.target.value);
  };

  return (
    <Input
      className={styles['sidebar-search-input']}
      value={searchText}
      onChange={handleSearchTextChange}
      placeholder={placeholder}
      prefix={<SVG name='search' size={16} color='#BFBFBF' />}
      onFocus={onFocusSearch}
    />
  );
};

export default SidebarSearch;
