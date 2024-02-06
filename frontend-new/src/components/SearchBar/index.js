import React, { useCallback, useRef } from 'react';
import { Input } from 'antd';
import { useDispatch } from 'react-redux';
import { TOGGLE_GLOBAL_SEARCH } from 'Reducers/types';
import { SVG } from '../factorsComponents';

function SearchBar({ placeholder, type }) {
  const inputRef = useRef(null);
  const dispatch = useDispatch();
  const handleFocus = useCallback(() => {
    dispatch({ type: TOGGLE_GLOBAL_SEARCH });
  }, [dispatch]);

  return (
    <Input
      data-tour='step-2'
      ref={inputRef}
      size='large'
      placeholder={placeholder}
      prefix={<SVG name='search' size={16} color='#BFBFBF' />}
      className={`fa-global-search--input ${
        type === 1
          ? 'fa-global-search--input-placeholder-lg'
          : 'fa-global-search--input-placeholder-sm'
      }`}
      onFocus={handleFocus}
    />
  );
}

export default SearchBar;
