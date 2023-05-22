import React from 'react';
import { Input } from 'antd';
import styles from './index.module.scss';
import { SVG } from 'Components/factorsComponents';

const SidebarSearch = ({ searchText, setSearchText }) => {
  const handleSearchTextChange = (e) => {
    setSearchText(e.target.value);
  };

  return (
    <Input
      className={styles['sidebar-search-input']}
      value={searchText}
      onChange={handleSearchTextChange}
      placeholder='Search board'
      prefix={<SVG name={'search'} size={16} color={'#BFBFBF'} />}
    />
  );
};

export default SidebarSearch;
