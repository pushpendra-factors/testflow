import React from 'react';
import { Input } from 'antd';
import styles from './index.module.scss';
import { SVG } from '../factorsComponents';

function SearchBar() {
  return (
    <Input
      size="large"
      placeholder="Lookup factors.ai"
      prefix={(
        <SVG name={'search'} size={24} color={'black'}/>
      )}
      className={styles.searchBarBox}
    />
  );
}

export default SearchBar;
