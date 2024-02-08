import React, { useCallback, useRef } from 'react';
import { Input } from 'antd';
import { useDispatch } from 'react-redux';
import { TOGGLE_GLOBAL_SEARCH } from 'Reducers/types';
import { SVG } from '../factorsComponents';

function SearchBar({ placeholder }) {
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
      className={`fa-global-search--input`}
      onFocus={handleFocus}
    />
  );
}

export default SearchBar;
